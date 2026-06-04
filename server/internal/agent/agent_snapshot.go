package agent

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"offer-catcher/server/internal/model"
)

// normalizeResumeToSnapshot is the agent-local copy of NormalizeResumeToSnapshotV2.
// It exists here to avoid circular imports (agent → service).
// The implementation mirrors service.NormalizeResumeToSnapshotV2 exactly.
func normalizeResumeToSnapshot(version *model.ResumeVersion) (model.ResumeSnapshotV2, error) {
	snapshot := model.ResumeSnapshotV2{
		Meta: model.ResumeSnapshotMeta{
			SchemaVersion:   2,
			ResumeVersionNo: version.VersionNo,
			ResumeName:      version.Name,
			TargetDirection: version.TargetRole,
		},
	}

	isV2 := version.SchemaVersion >= 2

	if isV2 {
		if err := normalizeV2Snapshot(&snapshot, version); err != nil {
			return snapshot, err
		}
	} else {
		if err := normalizeV1Snapshot(&snapshot, version); err != nil {
			return snapshot, err
		}
	}

	return snapshot, nil
}

// snapshotToPromptMap converts ResumeSnapshotV2 to a map for LLM prompt injection.
func snapshotToPromptMap(snapshot model.ResumeSnapshotV2) map[string]any {
	emptyStr := func(s string) string {
		if s == "" { return "未填写" }
		return s
	}
	emptySlice := func(s []string) []string {
		if len(s) == 0 { return []string{"未填写"} }
		return s
	}

	return map[string]any{
		"resume_name":       emptyStr(snapshot.Meta.ResumeName),
		"resume_version_no": snapshot.Meta.ResumeVersionNo,
		"target_direction":  emptyStr(snapshot.Meta.TargetDirection),
		"profile": map[string]any{
			"name": emptyStr(snapshot.Profile.Name), "school": emptyStr(snapshot.Profile.School),
			"degree": emptyStr(snapshot.Profile.Degree), "major": emptyStr(snapshot.Profile.Major),
			"graduation_year": emptyStr(snapshot.Profile.GraduationYear), "current_city": emptyStr(snapshot.Profile.CurrentCity),
		},
		"job_intention": map[string]any{
			"target_roles": emptySlice(snapshot.JobIntention.TargetRoles),
			"job_type": emptyStr(snapshot.JobIntention.JobType),
			"expected_cities": emptySlice(snapshot.JobIntention.ExpectedCities),
			"accept_other_cities": snapshot.JobIntention.AcceptOtherCities,
			"earliest_start_date": emptyStr(snapshot.JobIntention.EarliestStartDate),
			"internship_duration": emptyStr(snapshot.JobIntention.InternshipDuration),
			"weekly_available_days": snapshot.JobIntention.WeeklyAvailableDays,
			"target_industries": emptySlice(snapshot.JobIntention.TargetIndustries),
			"target_companies": emptySlice(snapshot.JobIntention.TargetCompanies),
		},
		"educations":   snapshot.Educations,
		"internships":  snapshot.Internships,
		"projects":     snapshot.Projects,
		"skills":       snapshot.Skills,
		"certificates": snapshot.Certificates,
		"awards":       snapshot.Awards,
		"portfolio":    snapshot.Portfolio,
		"extra_info":   snapshot.ExtraInfo,
	}
}

// buildResumeSnapshotJSON converts a ResumeVersion into RESUME_SNAPSHOT_JSON string.
func buildResumeSnapshotJSON(resume *model.ResumeVersion) (string, error) {
	snapshot, err := normalizeResumeToSnapshot(resume)
	if err != nil {
		log.Printf("[agent] WARNING: snapshot normalize failed for resume %d: %v", resume.ID, err)
		return "{}", fmt.Errorf("snapshot normalize: %w", err)
	}
	data, err := json.Marshal(snapshotToPromptMap(snapshot))
	if err != nil {
		return "{}", fmt.Errorf("snapshot marshal: %w", err)
	}
	return string(data), nil
}

// ─── Inline V1/V2 normalizers (mirror from service/resume_snapshot_normalizer.go) ───

func normalizeV1Snapshot(snapshot *model.ResumeSnapshotV2, v *model.ResumeVersion) error {
	snapshot.Meta.TargetDirection = v.TargetRole
	if v.Name != "" {
		snapshot.Meta.ResumeName = v.Name
	}
	snapshot.Profile = model.ResumeSnapshotProfile{CurrentCity: v.Location}

	snapshot.JobIntention = model.ResumeSnapshotJobIntention{
		TargetRoles: make([]string, 0), ExpectedCities: make([]string, 0),
	}
	if v.TargetRole != "" {
		snapshot.JobIntention.TargetRoles = append(snapshot.JobIntention.TargetRoles, v.TargetRole)
	}
	if v.TargetCity != "" {
		snapshot.JobIntention.ExpectedCities = append(snapshot.JobIntention.ExpectedCities, v.TargetCity)
	}

	// educations
	edus, _ := decodeJSONArray[map[string]any](bytesToStr(v.EducationJSON))
	for _, e := range edus {
		edu := model.ResumeSnapshotEducation{
			School: s(e, "school"), Degree: s(e, "degree"), Major: s(e, "major"),
			Department: s(e, "department"), StartDate: s(e, "start"), EndDate: s(e, "end"),
			GPA: s(e, "gpa"), GPAScale: s(e, "gpa_scale"), Rank: s(e, "ranking"),
			ResearchDirection: s(e, "research"), Lab: s(e, "lab"), PaperSummary: s(e, "papers"),
		}
		if courses := toStrSlice(e["courses"]); len(courses) > 0 {
			edu.CoreCourses = courses
		}
		snapshot.Educations = append(snapshot.Educations, edu)
	}
	if len(snapshot.Educations) > 0 {
		first := snapshot.Educations[0]
		if snapshot.Profile.School == "" { snapshot.Profile.School = first.School }
		if snapshot.Profile.Degree == "" { snapshot.Profile.Degree = first.Degree }
		if snapshot.Profile.Major == "" { snapshot.Profile.Major = first.Major }
	}

	// internships
	exps, _ := decodeJSONArray[map[string]any](bytesToStr(v.ExperienceJSON))
	for _, exp := range exps {
		intern := model.ResumeSnapshotInternship{
			Company: s(exp, "company"), Role: s(exp, "role"),
			StartDate: s(exp, "start"), EndDate: s(exp, "end"),
			WorkContent: s(exp, "description"), Metrics: s(exp, "metrics"),
		}
		if ts := toStrSlice(exp["tech_stack"]); len(ts) > 0 { intern.TechStack = ts }
		if ach := toStrSlice(exp["achievements"]); len(ach) > 0 { intern.Achievements = ach }
		snapshot.Internships = append(snapshot.Internships, intern)
	}

	// projects
	projs, _ := decodeJSONArray[map[string]any](bytesToStr(v.ProjectJSON))
	for _, p := range projs {
		proj := model.ResumeSnapshotProject{
			Name: s(p, "name"), Role: s(p, "role"), Description: s(p, "description"),
			Responsibilities: s(p, "responsibilities"), Link: s(p, "link"),
			Type: s(p, "type"), StartDate: s(p, "start"), EndDate: s(p, "end"),
		}
		if ts := toStrSlice(p["tech_stack"]); len(ts) > 0 { proj.TechStack = ts }
		if ach := toStrSlice(p["achievements"]); len(ach) > 0 { proj.Achievements = ach }
		snapshot.Projects = append(snapshot.Projects, proj)
	}

	// skills
	raw := bytesToStr(v.SkillsJSON)
	skillStrs, _ := decodeJSONArray[string](raw)
	snapshot.Skills = model.ResumeSnapshotSkills{OtherKeywords: skillStrs}
	snapshot.ExtraInfo = model.ResumeSnapshotExtraInfo{SelfEvaluation: v.SelfEvaluation}
	snapshot.Certificates = make([]model.ResumeSnapshotCertificate, 0)
	snapshot.Awards = make([]model.ResumeSnapshotAward, 0)
	snapshot.Portfolio = model.ResumeSnapshotPortfolio{}
	return nil
}

func normalizeV2Snapshot(snapshot *model.ResumeSnapshotV2, v *model.ResumeVersion) error {
	snapshot.Meta.TargetDirection = v.TargetDirection
	if v.Name != "" { snapshot.Meta.ResumeName = v.Name }

	// profile
	prof, _ := decodeJSONObject[map[string]any](bytesToStr(v.ProfileJSON), map[string]any{})
	snapshot.Profile = model.ResumeSnapshotProfile{
		Name: s(prof, "name"), School: s(prof, "school"), Degree: s(prof, "degree"),
		Major: s(prof, "major"), GraduationYear: s(prof, "graduation_year"), CurrentCity: s(prof, "current_city"),
	}
	if snapshot.Profile.School == "" { snapshot.Profile.School = v.School }
	if snapshot.Profile.Degree == "" { snapshot.Profile.Degree = v.Degree }
	if snapshot.Profile.Major == "" { snapshot.Profile.Major = v.Major }
	if snapshot.Profile.GraduationYear == "" { snapshot.Profile.GraduationYear = v.GraduationYear }
	if snapshot.Profile.CurrentCity == "" { snapshot.Profile.CurrentCity = v.Location }

	// jobIntention
	ji, _ := decodeJSONObject[map[string]any](bytesToStr(v.JobIntentionJSON), map[string]any{})
	snapshot.JobIntention = model.ResumeSnapshotJobIntention{
		TargetRoles: toStrSlice(ji["target_roles"]), JobType: s(ji, "job_type"),
		ExpectedCities: toStrSlice(ji["expected_cities"]), AcceptOtherCities: toBool(ji["accept_other_cities"]),
		EarliestStartDate: s(ji, "earliest_start_date"), InternshipDuration: s(ji, "internship_duration"),
		TargetIndustries: toStrSlice(ji["target_industries"]), TargetCompanies: toStrSlice(ji["target_companies"]),
	}
	if wad, ok := ji["weekly_available_days"]; ok { snapshot.JobIntention.WeeklyAvailableDays = toInt(wad) }
	if snapshot.JobIntention.JobType == "" { snapshot.JobIntention.JobType = v.JobType }
	if len(snapshot.JobIntention.ExpectedCities) == 0 && v.ExpectedCities != "" {
		for _, p := range strings.Split(v.ExpectedCities, ",") {
			p = strings.TrimSpace(p)
			if p != "" { snapshot.JobIntention.ExpectedCities = append(snapshot.JobIntention.ExpectedCities, p) }
		}
	}
	if len(snapshot.JobIntention.TargetRoles) == 0 && v.TargetRole != "" {
		snapshot.JobIntention.TargetRoles = append(snapshot.JobIntention.TargetRoles, v.TargetRole)
	}

	// educations
	edus, _ := decodeJSONArray[map[string]any](bytesToStr(v.EducationJSON))
	for _, e := range edus {
		edu := model.ResumeSnapshotEducation{
			School: s(e, "school"), Degree: s(e, "degree"), Major: s(e, "major"),
			Department: s(e, "department"), StartDate: s(e, "start_date"), EndDate: s(e, "end_date"),
			GPA: s(e, "gpa"), GPAScale: s(e, "gpa_scale"), Rank: s(e, "ranking"),
			ResearchDirection: s(e, "research_direction"), Lab: s(e, "advisor"), PaperSummary: s(e, "papers"),
		}
		if courses := toStrSlice(e["core_courses"]); len(courses) > 0 {
			edu.CoreCourses = courses
		} else if courses := toStrSlice(e["courses"]); len(courses) > 0 {
			edu.CoreCourses = courses
		}
		if edu.StartDate == "" { edu.StartDate = s(e, "start") }
		if edu.EndDate == "" { edu.EndDate = s(e, "end") }
		snapshot.Educations = append(snapshot.Educations, edu)
	}

	// internships
	interns, _ := decodeJSONArray[map[string]any](bytesToStr(v.InternshipJSON))
	if len(interns) == 0 {
		interns, _ = decodeJSONArray[map[string]any](bytesToStr(v.ExperienceJSON))
	}
	for _, exp := range interns {
		intern := model.ResumeSnapshotInternship{
			Company: s(exp, "company"), Role: s(exp, "role"),
			StartDate: s(exp, "start_date"), EndDate: s(exp, "end_date"),
			WorkContent: s(exp, "content"), Metrics: s(exp, "metrics"),
		}
		if intern.WorkContent == "" { intern.WorkContent = s(exp, "work_content") }
		if intern.StartDate == "" { intern.StartDate = s(exp, "start") }
		if intern.EndDate == "" { intern.EndDate = s(exp, "end") }
		if ts := toStrSlice(exp["tech_stack"]); len(ts) > 0 { intern.TechStack = ts }
		if ach := toStrSlice(exp["achievements"]); len(ach) > 0 { intern.Achievements = ach }
		snapshot.Internships = append(snapshot.Internships, intern)
	}

	// projects
	projs, _ := decodeJSONArray[map[string]any](bytesToStr(v.ProjectJSON))
	for _, p := range projs {
		proj := model.ResumeSnapshotProject{
			Name: s(p, "name"), Role: s(p, "role"), Type: s(p, "type"),
			StartDate: s(p, "start_date"), EndDate: s(p, "end_date"),
			Description: s(p, "description"), Responsibilities: s(p, "responsibility"), Link: s(p, "link"),
		}
		if proj.Responsibilities == "" { proj.Responsibilities = s(p, "responsibilities") }
		if proj.StartDate == "" { proj.StartDate = s(p, "start") }
		if proj.EndDate == "" { proj.EndDate = s(p, "end") }
		if ts := toStrSlice(p["tech_stack"]); len(ts) > 0 { proj.TechStack = ts }
		if ach := toStrSlice(p["achievements"]); len(ach) > 0 { proj.Achievements = ach }
		snapshot.Projects = append(snapshot.Projects, proj)
	}

	// skills (K8-5-26-B: final 9-module — new keys first, old fallback)
	sk, _ := decodeJSONObject[map[string]any](bytesToStr(v.SkillsJSON), map[string]any{})
	profSkills := toStrSlice(sk["development_languages"])
	aiApp := s(sk, "ai_application")
	oldLangs := toStrSlice(sk["languages"])
	oldAI := toStrSlice(sk["ai_scenarios"])
	snapshot.Skills = model.ResumeSnapshotSkills{
		ProfessionalSkills: profSkills,
		AIApplication:      aiApp,
		ProgrammingLanguages:   orSliceStr(oldLangs, profSkills),
		BackendFrameworks:      toStrSlice(sk["frameworks"]),
		FrontendTechnologies:   toStrSlice(sk["frontend"]),
		Databases:              toStrSlice(sk["databases"]),
		Middleware:             toStrSlice(sk["middleware"]),
		CloudPlatforms:         toStrSlice(sk["cloud"]),
		DevTools:               toStrSlice(sk["tools"]),
		AITools:                toStrSlice(sk["ai_skills"]),
		AIApplicationScenarios: orSliceStr(oldAI, splitStr1(aiApp)),
		OtherKeywords:          toStrSlice(sk["other_keywords"]),
	}

	// certs
	certs, _ := decodeJSONArray[map[string]any](bytesToStr(v.CertificatesJSON))
	for _, c := range certs {
		cert := model.ResumeSnapshotCertificate{
			Name: s(c, "name"), Type: s(c, "type"), Level: s(c, "level"), Score: s(c, "score"),
		}
		cert.IssuedAt = s(c, "date")
		if cert.IssuedAt == "" { cert.IssuedAt = s(c, "issued_at") }
		snapshot.Certificates = append(snapshot.Certificates, cert)
	}

	// awards
	awards, _ := decodeJSONArray[map[string]any](bytesToStr(v.AwardsJSON))
	for _, a := range awards {
		awd := model.ResumeSnapshotAward{
			Name: s(a, "name"), Type: s(a, "type"), Level: s(a, "level"), Description: s(a, "description"),
		}
		awd.AwardedAt = s(a, "date")
		if awd.AwardedAt == "" { awd.AwardedAt = s(a, "awarded_at") }
		snapshot.Awards = append(snapshot.Awards, awd)
	}

	// portfolio (K8-5-26-B: new works + homepage, old fallback)
	port, _ := decodeJSONObject[map[string]any](bytesToStr(v.PortfolioJSON), map[string]any{})
	hp := s(port, "homepage")
	snapshot.Portfolio = model.ResumeSnapshotPortfolio{
		Works:    parseAgentPortfolioWorks(port),
		Homepage: hp,
		Github: s(port, "github"), Gitee: s(port, "gitee"), CNB: s(port, "cnb"),
		Blog: s(port, "blog"), DemoLinks: toStrSlice(port["demo_links"]), PortfolioLinks: toStrSlice(port["other_links"]),
	}
	if snapshot.Portfolio.Homepage == "" {
		snapshot.Portfolio.Homepage = s(port, "blog")
	}

	// extraInfo
	extra, _ := decodeJSONObject[map[string]any](bytesToStr(v.ExtraJSON), map[string]any{})
	snapshot.ExtraInfo = model.ResumeSnapshotExtraInfo{
		SelfEvaluation: s(extra, "self_evaluation"), AdditionalInfo: s(extra, "additional_info"),
	}
	if snapshot.ExtraInfo.SelfEvaluation == "" { snapshot.ExtraInfo.SelfEvaluation = v.SelfEvaluation }
	return nil
}

// ─── Generic JSON helpers (agent-local) ─────────────────────────────────

func decodeJSONObject[T any](raw string, fallback T) (T, error) {
	if raw == "" || raw == "null" { return fallback, nil }
	var result T
	if err := json.Unmarshal([]byte(raw), &result); err != nil { return fallback, fmt.Errorf("invalid JSON: %w", err) }
	return result, nil
}

func decodeJSONArray[T any](raw string) ([]T, error) {
	if raw == "" || raw == "null" { return []T{}, nil }
	var result []T
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		var single T
		if err2 := json.Unmarshal([]byte(raw), &single); err2 == nil { return []T{single}, nil }
		return nil, fmt.Errorf("invalid JSON array: %w", err)
	}
	return result, nil
}

func bytesToStr(b []byte) string { return string(b) }

func s(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		if str, ok := v.(string); ok { return str }
	}
	return ""
}

func toStrSlice(v any) []string {
	if v == nil { return nil }
	arr, ok := v.([]any)
	if !ok { return nil }
	result := make([]string, 0, len(arr))
	for _, item := range arr {
		if str, ok := item.(string); ok { result = append(result, str) }
	}
	return result
}

func toBool(v any) bool {
	if v == nil { return false }
	b, ok := v.(bool)
	return ok && b
}

func toInt(v any) int {
	if v == nil { return 0 }
	switch n := v.(type) {
	case float64: return int(n)
	case int: return n
	case int64: return int(n)
	case json.Number:
		if i, err := n.Int64(); err == nil { return int(i) }
	}
	return 0
}

// K8-5-26-B helpers

func orSliceStr(a, b []string) []string {
	if len(a) > 0 { return a }
	return b
}

func splitStr1(s string) []string {
	if s == "" { return nil }
	return []string{s}
}

func parseAgentPortfolioWorks(port map[string]any) []model.ResumeSnapshotPortfolioWork {
	raw, ok := port["works"]
	if !ok { return nil }
	arr, ok := raw.([]any)
	if !ok { return nil }
	works := make([]model.ResumeSnapshotPortfolioWork, 0, len(arr))
	for _, w := range arr {
		wm, ok := w.(map[string]any)
		if !ok { continue }
		works = append(works, model.ResumeSnapshotPortfolioWork{
			Name: s(wm, "name"), Link: s(wm, "link"), Password: s(wm, "password"),
		})
	}
	return works
}
