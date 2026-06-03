package main

import (
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"
)

// ─── JSON 类型（复用 import_jobs 结构）────────────────────────────────────────────

type JobRecord struct {
	Source             SourceInfo `json:"source"`
	Job                JobInfo    `json:"job"`
	Organizations      []OrgInfo  `json:"organizations"`
	Locations          []string   `json:"locations"`
	InterviewLocations []string   `json:"interview_locations"`
	Description        DescInfo   `json:"description"`
	Requirements       ReqInfo    `json:"requirements"`
	Skills             []string   `json:"skills"`
	Facets             FacetsInfo `json:"facets"`
}

type SourceInfo struct {
	SourceType  string `json:"source_type"`
	SourceName  string `json:"source_name"`
	SourceURL   string `json:"source_url"`
	DataQuality string `json:"data_quality"`
	CollectedAt string `json:"collected_at"`
}

type JobInfo struct {
	ID              string `json:"id"`
	Title           string `json:"title"`
	JobType         string `json:"job_type"`
	ApplicationType string `json:"application_type"`
	Direction       string `json:"direction"`
	Category        string `json:"category"`
	Status          string `json:"status"`
	Slogan          string `json:"slogan"`
}

type OrgInfo struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type DescInfo struct {
	Summary         string   `json:"summary"`
	Responsibilities []string `json:"responsibilities"`
}

type ReqInfo struct {
	MustHave    []string `json:"must_have"`
	NiceToHave  []string `json:"nice_to_have"`
	BonusPoints []string `json:"bonus_points"`
	Notes       []string `json:"notes,omitempty"`
}

type FacetsInfo struct {
	ApplicationProject string   `json:"application_project"`
	ApplicationSubtype string   `json:"application_subtype"`
	JobCategory        string   `json:"job_category"`
	JobType            string   `json:"job_type"`
	Direction          string   `json:"direction"`
	Cities             []string `json:"cities"`
	BusinessGroups     []string `json:"business_groups"`
}

// ─── 筛选维表辅助类型 ──────────────────────────────────────────────────────────

type FilterGroup struct {
	Code              string
	Name              string
	DisplayOrder      int
	DefaultExpanded   bool
	MaxVisibleOptions int
}

type FilterOption struct {
	GroupCode       string
	Code            string
	Label           string
	Value           string
	DisplayOrder    int
	ParentGroupCode string // 父选项所属的 group code，空表示无父级
	ParentValue     string // 父选项的 value，空表示无父级
}

// ─── SQL 转义工具 ──────────────────────────────────────────────────────────────

func sqlStr(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

func jsonArr(ss []string) string {
	if len(ss) == 0 {
		return sqlStr("[]")
	}
	var parts []string
	for _, s := range ss {
		parts = append(parts, sqlStr(s))
	}
	return "JSON_ARRAY(" + strings.Join(parts, ",") + ")"
}

func jsonOrgs(orgs []OrgInfo) string {
	if len(orgs) == 0 {
		return sqlStr("[]")
	}
	var parts []string
	for _, o := range orgs {
		parts = append(parts, sqlStr(`{"code":`+sqlStr(o.Code)+`,"name":`+sqlStr(o.Name)+`}`))
	}
	return "CONCAT('['," + strings.Join(parts, ",'\\n',") + ",'\\n]')"
}

// ─── 主函数 ────────────────────────────────────────────────────────────────────

func main() {
	inputPath := flag.String("input", "", "Path to tencent-campus-jobs.json")
	outputPath := flag.String("output", "", "Path for output SQL seed")
	flag.Parse()

	if *inputPath == "" || *outputPath == "" {
		log.Fatal("--input and --output are required")
	}

	// 1. 读取并解析 JSON
	data, err := os.ReadFile(*inputPath)
	if err != nil {
		log.Fatalf("read input: %v", err)
	}
	var jobs []JobRecord
	if err := json.Unmarshal(data, &jobs); err != nil {
		log.Fatalf("parse JSON: %v", err)
	}

	// 2. 校验
	if len(jobs) < 40 {
		log.Fatalf("expected >= 40 jobs, got %d", len(jobs))
	}
	seenIDs := map[string]bool{}
	for i, j := range jobs {
		if j.Job.ID == "" {
			log.Fatalf("job #%d: missing id", i)
		}
		if seenIDs[j.Job.ID] {
			log.Fatalf("job #%d: duplicate id=%s", i, j.Job.ID)
		}
		seenIDs[j.Job.ID] = true
		if j.Job.Title == "" {
			log.Fatalf("job %s: missing title", j.Job.ID)
		}
		if len(j.Description.Responsibilities) == 0 {
			log.Fatalf("job %s: missing responsibilities", j.Job.ID)
		}
	}
	fmt.Printf("validation pass: %d jobs\n", len(jobs))

	// 3. 构建筛选维表数据
	groups := []FilterGroup{
		{Code: "application_project", Name: "应聘项目", DisplayOrder: 1, DefaultExpanded: true, MaxVisibleOptions: 8},
		{Code: "job_category", Name: "岗位类别", DisplayOrder: 2, DefaultExpanded: true, MaxVisibleOptions: 8},
		{Code: "city", Name: "工作城市", DisplayOrder: 3, DefaultExpanded: true, MaxVisibleOptions: 8},
		{Code: "interview_city", Name: "面试地点", DisplayOrder: 4, DefaultExpanded: false, MaxVisibleOptions: 6},
		{Code: "business_group", Name: "事业群", DisplayOrder: 5, DefaultExpanded: false, MaxVisibleOptions: 6},
		{Code: "skill", Name: "技能关键词", DisplayOrder: 6, DefaultExpanded: false, MaxVisibleOptions: 10},
	}

	// 岗位类别父子层级: parent -> children
	catParents := map[string][]string{}
	for _, j := range jobs {
		catParents[j.Job.JobType] = append(catParents[j.Job.JobType], j.Job.Category)
	}
	// 去重并排序
	parentOrder := []string{"技术", "产品", "设计", "市场", "职能"}
	catParentIdx := map[string]int{}
	for i, p := range parentOrder {
		catParentIdx[p] = i
	}
	var categoryOptions []FilterOption
	parentOptionID := 1
	catDisplayOrder := 10
	for _, pt := range parentOrder {
		cats := catParents[pt]
		if len(cats) == 0 {
			continue
		}
		uniq := map[string]bool{}
		for _, c := range cats {
			uniq[c] = true
		}
		sorted := make([]string, 0, len(uniq))
		for c := range uniq {
			sorted = append(sorted, c)
		}
		sort.Strings(sorted)

		parentCode := pt
		parentLabel := pt
		categoryOptions = append(categoryOptions, FilterOption{
			GroupCode: "job_category", Code: parentCode, Label: parentLabel, Value: parentLabel,
			DisplayOrder: catDisplayOrder,
		})
		catDisplayOrder += 10
		parentOptionID++

		for _, cat := range sorted {
			categoryOptions = append(categoryOptions, FilterOption{
				GroupCode: "job_category", Code: cat, Label: cat, Value: cat,
				DisplayOrder: catDisplayOrder, ParentValue: parentLabel,
			})
			catDisplayOrder++
		}
	}

	// 城市选项: 按频次降序
	cityCount := map[string]int{}
	for _, j := range jobs {
		for _, c := range j.Locations {
			cityCount[c]++
		}
	}
	type kv struct{ k string; v int }
	var cityList []kv
	for k, v := range cityCount {
		cityList = append(cityList, kv{k, v})
	}
	sort.Slice(cityList, func(i, j int) bool { return cityList[i].v > cityList[j].v })
	var cityOptions []FilterOption
	for i, c := range cityList {
		cityOptions = append(cityOptions, FilterOption{
			GroupCode: "city", Code: c.k, Label: c.k, Value: c.k, DisplayOrder: i * 10,
		})
	}

	// 面试城市
	ivCityCount := map[string]int{}
	for _, j := range jobs {
		for _, c := range j.InterviewLocations {
			ivCityCount[c]++
		}
	}
	var ivCityList []kv
	for k, v := range ivCityCount {
		ivCityList = append(ivCityList, kv{k, v})
	}
	sort.Slice(ivCityList, func(i, j int) bool { return ivCityList[i].v > ivCityList[j].v })
	var ivOptions []FilterOption
	for i, c := range ivCityList {
		ivOptions = append(ivOptions, FilterOption{
			GroupCode: "interview_city", Code: c.k, Label: c.k, Value: c.k, DisplayOrder: i * 10,
		})
	}

	// 事业群选项
	bgSet := map[string]OrgInfo{}
	for _, j := range jobs {
		for _, o := range j.Organizations {
			bgSet[o.Code] = o
		}
	}
	var bgList []OrgInfo
	for _, o := range bgSet {
		bgList = append(bgList, o)
	}
	sort.Slice(bgList, func(i, j int) bool { return bgList[i].Code < bgList[j].Code })
	var bgOptions []FilterOption
	for i, o := range bgList {
		bgOptions = append(bgOptions, FilterOption{
			GroupCode: "business_group", Code: o.Code, Label: o.Name, Value: o.Code, DisplayOrder: i * 10,
		})
	}

	// 项目选项
	projSet := map[string]bool{}
	for _, j := range jobs {
		if j.Facets.ApplicationProject != "" {
			projSet[j.Facets.ApplicationProject] = true
		}
	}
	var projList []string
	for p := range projSet {
		projList = append(projList, p)
	}
	sort.Strings(projList)
	var projOptions []FilterOption
	for i, p := range projList {
		projOptions = append(projOptions, FilterOption{
			GroupCode: "application_project", Code: p, Label: p, Value: p, DisplayOrder: i * 10,
		})
	}

	// 技能选项: 按频次
	skillCount := map[string]int{}
	for _, j := range jobs {
		for _, s := range j.Skills {
			skillCount[s]++
		}
	}
	var skillList []kv
	for k, v := range skillCount {
		skillList = append(skillList, kv{k, v})
	}
	sort.Slice(skillList, func(i, j int) bool { return skillList[i].v > skillList[j].v })
	var skillOptions []FilterOption
	for i, s := range skillList {
		skillOptions = append(skillOptions, FilterOption{
			GroupCode: "skill", Code: s.k, Label: s.k, Value: s.k, DisplayOrder: i * 10,
		})
	}

	allOptions := append(categoryOptions, cityOptions...)
	allOptions = append(allOptions, ivOptions...)
	allOptions = append(allOptions, bgOptions...)
	allOptions = append(allOptions, projOptions...)
	allOptions = append(allOptions, skillOptions...)

	// Option ID 分配
	optionIDs := map[string]int{} // key: "groupCode::optionCode"
	for i := range allOptions {
		key := allOptions[i].GroupCode + "::" + allOptions[i].Code
		optionIDs[key] = i + 1
	}
	// Parent ID lookups
	parentIDMap := map[string]int{} // key: "parentGroupCode::parentValue"
	for i := range categoryOptions {
		if categoryOptions[i].ParentValue == "" {
			key := categoryOptions[i].GroupCode + "::" + categoryOptions[i].Value
			parentIDMap[key] = i + 1
		}
	}

	// 4. 生成 SQL
	seedVersion := "2026-06-03-001"
	now := time.Now().Format("2006-01-02T15:04:05Z")
	hash := sha256.Sum256(data)
	checksum := fmt.Sprintf("%x", hash)[:16]

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`-- ============================================================
-- 腾讯校园招聘岗位 Seed 数据
-- 自动生成，不要手工编辑！
-- 生成时间: %s
-- seed_version: %s
-- checksum: %s
-- 数据源: resources/jobs/tencent-campus-jobs.json
-- 岗位数量: %d
-- 重新生成: cd server && go run ./cmd/generate_jobs_seed --input ../resources/jobs/tencent-campus-jobs.json --output ./sql/seed/tencent_jobs.sql
-- ============================================================

-- K8-4-3-R2: 确保导入时使用 utf8mb4 编码
SET NAMES utf8mb4;

START TRANSACTION;

-- ── 1. 清理旧 curated / mock 数据 ──
DELETE FROM match_analyses WHERE job_id IN (SELECT id FROM jobs WHERE external_source IN ('manual_curated','local_mock'));
DELETE FROM job_filter_option_relations WHERE job_id IN (SELECT id FROM jobs WHERE external_source IN ('manual_curated','local_mock'));
DELETE FROM job_skills WHERE job_id IN (SELECT id FROM jobs WHERE external_source IN ('manual_curated','local_mock'));
DELETE FROM job_locations WHERE job_id IN (SELECT id FROM jobs WHERE external_source IN ('manual_curated','local_mock'));
DELETE FROM job_business_groups WHERE job_id IN (SELECT id FROM jobs WHERE external_source IN ('manual_curated','local_mock'));
DELETE FROM jobs WHERE external_source IN ('manual_curated','local_mock');
DELETE FROM job_filter_options;
DELETE FROM job_filter_groups;
DELETE FROM seed_versions WHERE seed_key = 'tencent_jobs';

`, now, seedVersion, checksum, len(jobs)))

	// ── 2. 写入筛选分组 ──
	sb.WriteString("-- ── 2. 筛选分组 ──\n")
	sb.WriteString("INSERT INTO job_filter_groups (id, code, name, display_order, default_expanded, visible, max_visible_options) VALUES\n")
	for i, g := range groups {
		exp := 0
		if g.DefaultExpanded {
			exp = 1
		}
		comma := ","
		if i == len(groups)-1 {
			comma = ";"
		}
		sb.WriteString(fmt.Sprintf("  (%d, %s, %s, %d, %d, 1, %d)%s\n",
			i+1, sqlStr(g.Code), sqlStr(g.Name), g.DisplayOrder, exp, g.MaxVisibleOptions, comma))
	}
	sb.WriteString("\n")

	// ── 3. 写入筛选选项 ──
	sb.WriteString("-- ── 3. 筛选选项 ──\n")
	sb.WriteString("INSERT INTO job_filter_options (id, group_id, code, label, value, display_order, visible, parent_option_id) VALUES\n")
	for i, opt := range allOptions {
		// Find group_id
		groupID := 0
		for gi, g := range groups {
			if g.Code == opt.GroupCode {
				groupID = gi + 1
				break
			}
		}
		// Find parent_option_id
		parentID := "NULL"
		if opt.ParentValue != "" {
			if pid, ok := parentIDMap[opt.GroupCode+"::"+opt.ParentValue]; ok {
				parentID = fmt.Sprintf("%d", pid)
			}
		}
		comma := ","
		if i == len(allOptions)-1 {
			comma = ";"
		}
		sb.WriteString(fmt.Sprintf("  (%d, %d, %s, %s, %s, %d, 1, %s)%s\n",
			optionIDs[opt.GroupCode+"::"+opt.Code], groupID, sqlStr(opt.Code), sqlStr(opt.Label), sqlStr(opt.Value), opt.DisplayOrder, parentID, comma))
	}
	sb.WriteString("\n")

	// ── 4. 写入岗位主表 ──
	sb.WriteString("-- ── 4. 岗位主表 (共 46 条) ──\n")
	sb.WriteString("INSERT INTO jobs (id, title, department, city, direction, category, job_type, application_type, application_project, application_subtype, slogan, summary, description, requirements, responsibilities_json, must_have_json, nice_to_have_json, bonus_points_json, notes_json, skill_keywords_json, skills_json, organizations_json, locations_json, interview_locations_json, status, external_source, external_id, source_url, source_name, raw_json, data_quality, is_featured) VALUES\n")
	for i, j := range jobs {
		jobID := 10001 + i
		dept := ""
		if len(j.Organizations) > 0 {
			dept = j.Organizations[0].Name
		}
		city := ""
		if len(j.Locations) > 0 {
			city = j.Locations[0]
		}
		summary := j.Description.Summary
		if summary == "" {
			summary = j.Job.Title
		}

		reqText := buildReqText(j.Requirements)
		respJSON := toJSONStr(j.Description.Responsibilities)
		mustJSON := toJSONStr(j.Requirements.MustHave)
		niceJSON := toJSONStr(j.Requirements.NiceToHave)
		bonusJSON := toJSONStr(j.Requirements.BonusPoints)
		notesJSON := toJSONStr(j.Requirements.Notes)
		skillsJSON := toJSONStr(j.Skills)
		orgsJSON := toJSONStr(j.Organizations)
		locsJSON := toJSONStr(j.Locations)
		ivJSON := toJSONStr(j.InterviewLocations)
		rawJSON := toJSONStr(j)
		slogan := j.Job.Slogan
		if len(slogan) > 200 {
			slogan = slogan[:200]
		}

		comma := ","
		if i == len(jobs)-1 {
			comma = ";"
		}
		sb.WriteString(fmt.Sprintf("  (%d, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, 0)%s\n",
			jobID,
			sqlStr(j.Job.Title), sqlStr(dept), sqlStr(city), sqlStr(j.Job.Direction), sqlStr(j.Job.Category),
			sqlStr(j.Job.JobType), sqlStr(j.Job.ApplicationType),
			sqlStr(j.Facets.ApplicationProject), sqlStr(j.Facets.ApplicationSubtype),
			sqlStr(slogan), sqlStr(summary), sqlStr(summary), sqlStr(reqText),
			respJSON, mustJSON, niceJSON, bonusJSON, notesJSON,
			skillsJSON, skillsJSON, orgsJSON, locsJSON, ivJSON,
			sqlStr(j.Job.Status), sqlStr("manual_curated"), sqlStr(j.Job.ID),
			sqlStr(j.Source.SourceURL), sqlStr(j.Source.SourceName),
			rawJSON, sqlStr("manual_verified"), comma))
	}
	sb.WriteString("\n")

	// ── 5. 关联表 ──
	fmt.Fprintln(&sb, "-- ── 5. 岗位-地点 ──")
	fmt.Fprintln(&sb, "INSERT INTO job_locations (job_id, city) VALUES")
	var locRows []string
	for i, j := range jobs {
		jobID := 10001 + i
		for _, c := range j.Locations {
			locRows = append(locRows, fmt.Sprintf("  (%d, %s)", jobID, sqlStr(c)))
		}
	}
	fmt.Fprintln(&sb, strings.Join(locRows, ",\n"))
	fmt.Fprint(&sb, "ON DUPLICATE KEY UPDATE city = VALUES(city);\n\n")

	fmt.Fprintln(&sb, "-- ── 6. 岗位-事业群 ──")
	fmt.Fprintln(&sb, "INSERT INTO job_business_groups (job_id, code, name) VALUES")
	var bgRows []string
	for i, j := range jobs {
		jobID := 10001 + i
		for _, o := range j.Organizations {
			bgRows = append(bgRows, fmt.Sprintf("  (%d, %s, %s)", jobID, sqlStr(o.Code), sqlStr(o.Name)))
		}
	}
	fmt.Fprintln(&sb, strings.Join(bgRows, ",\n"))
	fmt.Fprint(&sb, "ON DUPLICATE KEY UPDATE name = VALUES(name);\n\n")

	fmt.Fprintln(&sb, "-- ── 7. 岗位-技能 ──")
	fmt.Fprintln(&sb, "INSERT INTO job_skills (job_id, skill) VALUES")
	var skRows []string
	for i, j := range jobs {
		jobID := 10001 + i
		for _, s := range j.Skills {
			skRows = append(skRows, fmt.Sprintf("  (%d, %s)", jobID, sqlStr(s)))
		}
	}
	fmt.Fprintln(&sb, strings.Join(skRows, ",\n"))
	fmt.Fprint(&sb, "ON DUPLICATE KEY UPDATE skill = VALUES(skill);\n\n")

	// ── 8. 岗位-筛选项关联 ──
	fmt.Fprintln(&sb, "-- ── 8. 岗位-筛选项关联 ──")
	fmt.Fprintln(&sb, "INSERT INTO job_filter_option_relations (job_id, option_id) VALUES")
	var relRows []string
	for i, j := range jobs {
		jobID := 10001 + i
		// 应聘项目
		if j.Facets.ApplicationProject != "" {
			key := "application_project::" + j.Facets.ApplicationProject
			if oid, ok := optionIDs[key]; ok {
				relRows = append(relRows, fmt.Sprintf("  (%d, %d)", jobID, oid))
			}
		}
		// 岗位类别: 父级 + 细类
		parentKey := "job_category::" + j.Job.JobType
		if oid, ok := optionIDs[parentKey]; ok {
			relRows = append(relRows, fmt.Sprintf("  (%d, %d)", jobID, oid))
		}
		childKey := "job_category::" + j.Job.Category
		if oid, ok := optionIDs[childKey]; ok {
			relRows = append(relRows, fmt.Sprintf("  (%d, %d)", jobID, oid))
		}
		// 城市
		for _, c := range j.Locations {
			key := "city::" + c
			if oid, ok := optionIDs[key]; ok {
				relRows = append(relRows, fmt.Sprintf("  (%d, %d)", jobID, oid))
			}
		}
		// 面试城市
		for _, c := range j.InterviewLocations {
			key := "interview_city::" + c
			if oid, ok := optionIDs[key]; ok {
				relRows = append(relRows, fmt.Sprintf("  (%d, %d)", jobID, oid))
			}
		}
		// 事业群
		for _, o := range j.Organizations {
			key := "business_group::" + o.Code
			if oid, ok := optionIDs[key]; ok {
				relRows = append(relRows, fmt.Sprintf("  (%d, %d)", jobID, oid))
			}
		}
		// 技能
		for _, s := range j.Skills {
			key := "skill::" + s
			if oid, ok := optionIDs[key]; ok {
				relRows = append(relRows, fmt.Sprintf("  (%d, %d)", jobID, oid))
			}
		}
	}
	fmt.Fprintln(&sb, strings.Join(relRows, ",\n"))
	fmt.Fprint(&sb, "ON DUPLICATE KEY UPDATE option_id = VALUES(option_id);\n\n")

	// ── 9. seed version ──
	fmt.Fprintln(&sb, "-- ── 9. seed 版本记录 ──")
	fmt.Fprintf(&sb, "INSERT INTO seed_versions (seed_key, seed_version, applied_at, checksum) VALUES (%s, %s, NOW(), %s) ON DUPLICATE KEY UPDATE seed_version = VALUES(seed_version), applied_at = VALUES(applied_at), checksum = VALUES(checksum);\n\n",
		sqlStr("tencent_jobs"), sqlStr(seedVersion), sqlStr(checksum))

	fmt.Fprintln(&sb, "COMMIT;")
	fmt.Fprintln(&sb, "-- EOF")

	// 5. 写入文件
	if err := os.MkdirAll(extractDir(*outputPath), 0755); err != nil {
		log.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(*outputPath, []byte(sb.String()), 0644); err != nil {
		log.Fatalf("write output: %v", err)
	}

	fmt.Printf("written: %s\n", *outputPath)
	fmt.Printf("seed_version: %s\n", seedVersion)
	fmt.Printf("checksum: %s\n", checksum)
	fmt.Printf("bytes: %d\n", len(sb.String()))
	fmt.Printf("estimated option count: %d\n", len(allOptions))
	fmt.Printf("estimated relation count: %d\n", len(relRows))
}

func buildReqText(r ReqInfo) string {
	var p []string
	if len(r.MustHave) > 0 {
		p = append(p, "【必备要求】\n"+strings.Join(r.MustHave, "\n"))
	}
	if len(r.NiceToHave) > 0 {
		p = append(p, "【了解项】\n"+strings.Join(r.NiceToHave, "\n"))
	}
	if len(r.BonusPoints) > 0 {
		p = append(p, "【加分项】\n"+strings.Join(r.BonusPoints, "\n"))
	}
	return strings.Join(p, "\n\n")
}

func toJSONStr(v interface{}) string {
	b, _ := json.Marshal(v)
	return sqlStr(string(b))
}

func extractDir(path string) string {
	idx := strings.LastIndex(path, "/")
	if idx < 0 {
		return "."
	}
	return path[:idx]
}
