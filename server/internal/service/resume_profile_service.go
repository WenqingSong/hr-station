package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"offer-catcher/server/internal/agent"
	"offer-catcher/server/internal/model"
	"offer-catcher/server/internal/repository"
)

// ResumeProfileService manages resume profile generation and status.
type ResumeProfileService struct {
	resumeRepo   *repository.ResumeRepository
	profileRepo  *repository.ResumeProfileRepository
	profileAgent *agent.ProfileAgent
}

// NewResumeProfileService creates a new ResumeProfileService.
func NewResumeProfileService(resumeRepo *repository.ResumeRepository, profileRepo *repository.ResumeProfileRepository) *ResumeProfileService {
	return &ResumeProfileService{resumeRepo: resumeRepo, profileRepo: profileRepo}
}

// SetProfileAgent injects the LLM profile agent for quality generation.
func (s *ResumeProfileService) SetProfileAgent(pa *agent.ProfileAgent) {
	s.profileAgent = pa
}

// EnsureProfile ensures a profile exists for the given resume and type.
// Returns the profile, a status string, and error.
func (s *ResumeProfileService) EnsureProfile(ctx context.Context, userID, resumeID uint64, profileType string) (*model.ResumeProfile, string, error) {
	resume, err := s.resumeRepo.GetByID(ctx, userID, resumeID)
	if err != nil {
		return nil, "", fmt.Errorf("ensure profile: %w", err)
	}

	profile, err := s.profileRepo.FindByResumeIDAndType(ctx, resumeID, profileType)
	if err != nil {
		return nil, "", fmt.Errorf("ensure profile: %w", err)
	}

	if profile == nil {
		if profileType == model.ProfileTypeFast {
			p, err := s.generateFast(ctx, userID, resumeID, resume)
			if err != nil {
				return nil, model.ProfileStatusFailed, err
			}
			return p, model.ProfileStatusReady, nil
		}
		_, _, _ = s.profileRepo.CreatePendingIfMissing(ctx, userID, resumeID, profileType)
		go s.generateQualityAsync(context.Background(), userID, resumeID)
		return nil, model.ProfileStatusPending, nil
	}

	switch profile.Status {
	case model.ProfileStatusPending:
		return profile, model.ProfileStatusPending, nil

	case model.ProfileStatusReady:
		// Check if stale: profile must have recorded resume_updated_at that matches current
		if profile.ResumeUpdatedAt != nil && !resume.UpdatedAt.IsZero() &&
			profile.ResumeUpdatedAt.Equal(resume.UpdatedAt) {
			return profile, model.ProfileStatusReady, nil
		}
		// K9-1-L: nil resume_updated_at means old profile → regenerate; mismatch means stale
		_ = s.profileRepo.MarkStaleByResumeID(ctx, resumeID)
		if profileType == model.ProfileTypeFast {
			p, err := s.generateFast(ctx, userID, resumeID, resume)
			if err != nil {
				return nil, model.ProfileStatusFailed, err
			}
			return p, model.ProfileStatusReady, nil
		}
		go s.generateQualityAsync(context.Background(), userID, resumeID)
		return nil, model.ProfileStatusPending, nil

	case model.ProfileStatusFailed:
		return profile, model.ProfileStatusFailed, nil

	case model.ProfileStatusStale:
		if profileType == model.ProfileTypeFast {
			p, err := s.generateFast(ctx, userID, resumeID, resume)
			if err != nil {
				return nil, model.ProfileStatusFailed, err
			}
			return p, model.ProfileStatusReady, nil
		}
		go s.generateQualityAsync(context.Background(), userID, resumeID)
		return nil, model.ProfileStatusPending, nil
	}

	return profile, profile.Status, nil
}

// GenerateFastProfile generates a fast rule-based profile.
func (s *ResumeProfileService) GenerateFastProfile(ctx context.Context, userID, resumeID uint64) (*model.ResumeProfile, error) {
	resume, err := s.resumeRepo.GetByID(ctx, userID, resumeID)
	if err != nil {
		return nil, fmt.Errorf("generate fast: %w", err)
	}
	return s.generateFast(ctx, userID, resumeID, (*model.ResumeVersion)(resume))
}

func (s *ResumeProfileService) generateFast(ctx context.Context, userID, resumeID uint64, resume *model.Resume) (*model.ResumeProfile, error) {
	// Create pending
	_, created, err := s.profileRepo.CreatePendingIfMissing(ctx, userID, resumeID, model.ProfileTypeFast)
	if err != nil {
		return nil, fmt.Errorf("generate fast: %w", err)
	}

	// Extract tags
	text := resume.RawText
	if text == "" {
		text = buildTextFromResume(resume)
	}
	tags := extractTags(text)

	dominant := dominantProfileV3(resume)
	primaryTags := buildPrimaryTags(tags, dominant)
	profileSummary := buildProfileSummary(dominant, tags)
	riskTags := extractRiskTags(tags, dominant)

	now := time.Now()
	primaryJSON, _ := json.Marshal(primaryTags)
	if string(primaryJSON) == "null" {
		primaryJSON = []byte("[]")
	}
	candJSON, _ := json.Marshal(tags)
	if string(candJSON) == "null" {
		candJSON = []byte("[]")
	}
	riskJSON, _ := json.Marshal(riskTags)
	if string(riskJSON) == "null" {
		riskJSON = []byte("[]")
	}
	evidenceJSON := []byte("{}")

	var resumeUpdAt *time.Time
	if !resume.UpdatedAt.IsZero() {
		resumeUpdAt = &resume.UpdatedAt
	}

	profile := &model.ResumeProfile{
		ResumeID:          resumeID,
		UserID:            userID,
		ProfileType:       model.ProfileTypeFast,
		ProfileVersion:    model.ProfileVersionV1,
		Status:            model.ProfileStatusReady,
		DominantProfile:   dominant,
		PrimaryTagsJSON:   string(primaryJSON),
		CandidateTagsJSON: string(candJSON),
		RiskTagsJSON:      string(riskJSON),
		EvidenceJSON:      string(evidenceJSON),
		ProfileSummary:    profileSummary,
		Source:            "rule",
		Generator:         "rule_fast_profile_v1",
		ResumeUpdatedAt:   resumeUpdAt,
		GeneratedAt:       &now,
	}

	if err := s.profileRepo.UpdateReady(ctx, profile); err != nil {
		_ = s.profileRepo.UpdateFailed(ctx, resumeID, model.ProfileTypeFast, err.Error())
		return nil, fmt.Errorf("generate fast: %w", err)
	}

	if created {
		profile.Status = model.ProfileStatusReady
	}
	return profile, nil
}

// GenerateQualityProfileAsync triggers async quality profile generation.
func (s *ResumeProfileService) GenerateQualityProfileAsync(ctx context.Context, userID, resumeID uint64) {
	go s.generateQualityAsync(context.Background(), userID, resumeID)
}

func (s *ResumeProfileService) generateQualityAsync(ctx context.Context, userID, resumeID uint64) {
	_, created, err := s.profileRepo.CreatePendingIfMissing(ctx, userID, resumeID, model.ProfileTypeQuality)
	if err != nil {
		log.Printf("[ResumeProfile] quality create pending failed resume=%d: %v", resumeID, err)
		return
	}
	if !created {
		log.Printf("[ResumeProfile] quality already exists resume=%d, skip", resumeID)
		return
	}

	resume, err := s.resumeRepo.GetByID(ctx, userID, resumeID)
	if err != nil {
		_ = s.profileRepo.UpdateFailed(ctx, resumeID, model.ProfileTypeQuality, err.Error())
		return
	}

	// K9-1-H5: try LLM extraction first, fallback to rule
	var dominant string
	var primaryTags []tagEntry
	var tags []tagEntry
	var riskTags []tagEntry
	var summary string
	var generator string
	var source string
	usedLLM := false

	if s.profileAgent != nil {
		llmResult, llmErr := s.profileAgent.ExtractCandidateProfile(ctx, resume)
		if llmErr == nil && llmResult != nil {
			dominant = llmResult.DominantProfile
			summary = llmResult.ProfileSummary
			primaryTags = convertLLMTags(llmResult.PrimaryTags)
			tags = convertLLMTags(llmResult.CandidateTags)
			riskTags = convertLLMTags(llmResult.RiskTags)
			generator = "llm_candidate_profile_v1"
			source = "agent"
			usedLLM = true
			log.Printf("[ResumeProfile] quality LLM success resume=%d dominant=%s", resumeID, dominant)
		} else {
			log.Printf("[ResumeProfile] quality LLM failed resume=%d: %v, fallback to rule", resumeID, llmErr)
		}
	}

	if !usedLLM {
		text := resume.RawText
		if text == "" {
			text = buildTextFromResume(resume)
		}
		tags = extractTags(text)
		dominant = dominantProfileV3(resume)
		primaryTags = buildPrimaryTags(tags, dominant)
		summary = buildProfileSummary(dominant, tags)
		riskTags = extractRiskTags(tags, dominant)
		generator = "rule_quality_fallback_v1"
		source = "rule"
	}

	now := time.Now()
	primaryJSON, _ := json.Marshal(primaryTags)
	if string(primaryJSON) == "null" { primaryJSON = []byte("[]") }
	candJSON, _ := json.Marshal(tags)
	if string(candJSON) == "null" { candJSON = []byte("[]") }
	riskJSON, _ := json.Marshal(riskTags)
	if string(riskJSON) == "null" { riskJSON = []byte("[]") }

	var qResumeUpdAt *time.Time
	if !resume.UpdatedAt.IsZero() {
		qResumeUpdAt = &resume.UpdatedAt
	}

	profile := &model.ResumeProfile{
		ResumeID:          resumeID,
		UserID:            userID,
		ProfileType:       model.ProfileTypeQuality,
		ProfileVersion:    model.ProfileVersionV1,
		Status:            model.ProfileStatusReady,
		DominantProfile:   dominant,
		PrimaryTagsJSON:   string(primaryJSON),
		CandidateTagsJSON: string(candJSON),
		RiskTagsJSON:      string(riskJSON),
		EvidenceJSON:      "{}",
		ProfileSummary:    summary,
		Source:            source,
		Generator:         generator,
		ResumeUpdatedAt:   qResumeUpdAt,
		GeneratedAt:       &now,
	}

	if err := s.profileRepo.UpdateReady(ctx, profile); err != nil {
		_ = s.profileRepo.UpdateFailed(ctx, resumeID, model.ProfileTypeQuality, err.Error())
	}
	log.Printf("[ResumeProfile] quality done resume=%d dominant=%s", resumeID, dominant)
}

// BackfillMissingProfiles generates fast profiles for all resumes without one.
func (s *ResumeProfileService) BackfillMissingProfiles(ctx context.Context, userID uint64) error {
	missing, err := s.profileRepo.ListMissingFastProfiles(ctx, userID)
	if err != nil {
		return fmt.Errorf("backfill: %w", err)
	}
	log.Printf("[ResumeProfile] backfill user=%d missing=%d", userID, len(missing))

	for _, rid := range missing {
		_, _, err := s.EnsureProfile(ctx, userID, rid, model.ProfileTypeFast)
		if err != nil {
			log.Printf("[ResumeProfile] backfill failed resume=%d: %v", rid, err)
			continue
		}
		log.Printf("[ResumeProfile] backfill done resume=%d", rid)
	}
	return nil
}

// MarkStaleOnResumeUpdate marks all profiles for a resume as stale.
func (s *ResumeProfileService) MarkStaleOnResumeUpdate(ctx context.Context, resumeID uint64) error {
	return s.profileRepo.MarkStaleByResumeID(ctx, resumeID)
}

// GetProfileStatus returns profile status for an API response.
func (s *ResumeProfileService) GetProfileStatus(ctx context.Context, userID, resumeID uint64) (*model.ResumeProfileStatusResponse, error) {
	resume, err := s.resumeRepo.FindByUserIDAndID(ctx, userID, resumeID)
	if err != nil {
		return nil, fmt.Errorf("get profile status: %w", err)
	}

	fast, _ := s.profileRepo.FindByResumeIDAndType(ctx, resumeID, model.ProfileTypeFast)
	quality, _ := s.profileRepo.FindByResumeIDAndType(ctx, resumeID, model.ProfileTypeQuality)

	resp := &model.ResumeProfileStatusResponse{
		ResumeID:      resumeID,
		FastStatus:    "missing",
		QualityStatus: "missing",
		UpdatedAt:     resume.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if fast != nil {
		resp.FastStatus = fast.Status
		if fast.DominantProfile != "" {
			resp.DominantProfile = fast.DominantProfile
		}
		if fast.PrimaryTagsJSON != "" && fast.PrimaryTagsJSON != "null" {
			var tags []model.ProfileTag
			if json.Unmarshal([]byte(fast.PrimaryTagsJSON), &tags) == nil {
				resp.PrimaryTags = tags
			}
		}
	}
	if quality != nil {
		resp.QualityStatus = quality.Status
		if resp.DominantProfile == "" && quality.DominantProfile != "" {
			resp.DominantProfile = quality.DominantProfile
		}
	}
	return resp, nil
}

// RegenerateProfile triggers profile regeneration for a resume.
func (s *ResumeProfileService) RegenerateProfile(ctx context.Context, userID, resumeID uint64, profileType string) error {
	if profileType != model.ProfileTypeFast && profileType != model.ProfileTypeQuality {
		return fmt.Errorf("invalid profile_type: %s", profileType)
	}

	_, err := s.resumeRepo.FindByUserIDAndID(ctx, userID, resumeID)
	if err != nil {
		return fmt.Errorf("regenerate: %w", err)
	}

	if profileType == model.ProfileTypeFast {
		_, err := s.GenerateFastProfile(ctx, userID, resumeID)
		return err
	}

	s.GenerateQualityProfileAsync(ctx, userID, resumeID)
	return nil
}

// ── tag extraction helpers ─────────────────────────────────────────────────

type tagEntry struct {
	Tag        string `json:"tag"`
	Group      string `json:"group"`
	Confidence int    `json:"confidence"`
	Evidence   string `json:"evidence,omitempty"`
}

func extractTags(text string) []tagEntry {
	lower := strings.ToLower(text)
	var tags []tagEntry

	// Direction tags
	addTag(&tags, "algorithm", "direction", countContains(lower,
		"算法", "机器学习", "深度学习", "推荐系统", "搜索推荐", "ctr", "排序模型", "nlp", "自然语言处理",
		"多模态", "计算机视觉", "数据挖掘", "表示学习", "用户行为建模", "pytorch", "tensorflow",
		"faiss", "auc", "ndcg", "recall", "embedding", "attention"), "技能和项目涉及算法与机器学习")

	addTag(&tags, "backend", "direction", countContains(lower,
		"go ", "java", "后端", "后台", "服务端", "微服务", "mysql", "redis", "kafka",
		"docker", "nginx", "接口", "权限", "分布式", "缓存", "api")/2, "技能列表中出现工程关键词")

	addTag(&tags, "research", "direction", countContains(lower,
		"研究", "论文", "实验室", "科研", "学术"), "包含研究方向或论文相关经历")

	addTag(&tags, "data", "direction", countContains(lower,
		"spark", "hive", "sql", "数据分析", "数据处理", "用户行为", "指标"), "包含数据处理和分析技能")

	// Capability tags
	addTag(&tags, "research_ability", "capability", countContains(lower,
		"研究", "论文", "实验复现", "消融", "学术", "实验室", "科研"), "研究方向与论文经验")

	addTag(&tags, "experiment_ability", "capability", countContains(lower,
		"实验", "消融", "指标", "auc", "ndcg", "recall", "准确率", "模型评估", "训练", "复现"), "包含模型实验和指标评估经历")

	addTag(&tags, "modeling_ability", "capability", countContains(lower,
		"pytorch", "tensorflow", "深度学习", "ctr", "排序模型", "embedding", "attention",
		"nlp", "多模态", "表示学习", "用户行为建模")/2, "具备深度学习建模和模型训练能力")

	addTag(&tags, "engineering_ability", "capability", min1(countContains(lower,
		"go", "java", "mysql", "redis", "docker", "nginx", "接口", "部署", "git", "linux")/3), "具备基础工程和工具使用能力")

	addTag(&tags, "data_analysis_ability", "capability", countContains(lower,
		"spark", "hive", "sql", "数据分析", "指标", "数据处理"), "具备数据分析和处理能力")

	// Domain tags
	addTag(&tags, "recommendation_system", "domain", countContains(lower,
		"推荐系统", "搜索推荐", "ctr", "点击率", "排序模型", "用户行为建模"), "研究方向与项目涉及推荐系统")

	addTag(&tags, "nlp", "domain", countContains(lower,
		"nlp", "自然语言处理", "文本", "语言模型"), "涉及自然语言处理相关技术")

	addTag(&tags, "multimodal", "domain", countContains(lower,
		"多模态", "视觉", "cv", "计算机视觉", "图像"), "涉及多模态或视觉相关技术")

	addTag(&tags, "data_mining", "domain", countContains(lower,
		"数据挖掘", "用户行为", "spark", "hive", "数据分析"), "涉及数据挖掘和分析")

	addTag(&tags, "backend_service", "domain", countContains(lower,
		"go", "java", "后端", "后台", "服务端", "微服务", "接口")/2, "具备后端服务开发基础")

	addTag(&tags, "business_understanding", "capability", countContains(lower,
		"业务", "产品", "商务", "供应链", "采购"), "涉及业务相关经验")

	return tags
}

func convertLLMTags(llmTags []agent.ProfileTagOutput) []tagEntry {
	result := make([]tagEntry, len(llmTags))
	for i, t := range llmTags {
		result[i] = tagEntry{Tag: t.Tag, Group: t.Group, Confidence: t.Confidence, Evidence: t.Evidence}
	}
	return result
}

func addTag(tags *[]tagEntry, tag, group string, conf int, evidence string) {
	if conf > 3 {
		conf = 3
	}
	if conf < 1 {
		conf = 1
	}
	if conf <= 0 {
		return
	}
	*tags = append(*tags, tagEntry{Tag: tag, Group: group, Confidence: conf, Evidence: evidence})
}

func countContains(text string, keywords ...string) int {
	count := 0
	for _, kw := range keywords {
		if strings.Contains(text, kw) {
			count++
		}
	}
	return count
}

func min1(v int) int {
	if v > 1 {
		return 1
	}
	return v
}

func dominantProfile(tags []tagEntry) string {
	tagMap := make(map[string]int)
	for _, t := range tags {
		tagMap[t.Tag] = t.Confidence
	}

	algo := tagMap["algorithm"] + tagMap["research_ability"] + tagMap["experiment_ability"] + tagMap["modeling_ability"]
	be := tagMap["backend"] + tagMap["engineering_ability"] + tagMap["backend_service"]
	prod := tagMap["product_sense"] + tagMap["business_understanding"]

	algoCount := 0
	if tagMap["algorithm"] >= 1 { algoCount++ }
	if tagMap["research_ability"] >= 1 { algoCount++ }
	if tagMap["experiment_ability"] >= 1 { algoCount++ }
	if tagMap["modeling_ability"] >= 1 { algoCount++ }
	hasDomain := tagMap["recommendation_system"] >= 1 || tagMap["nlp"] >= 1 || tagMap["multimodal"] >= 1 || tagMap["data_mining"] >= 1

	if algo > be || (algoCount >= 2) || (algoCount >= 1 && hasDomain) {
		return "algorithm_research"
	}

	if be >= algo && tagMap["backend"] >= 1 && tagMap["engineering_ability"] >= 1 && algo < 2 {
		return "backend_engineering"
	}
	if prod >= algo && prod >= be && prod >= 4 {
		return "product_operation"
	}

	return "career_exploration"
}

func buildPrimaryTags(tags []tagEntry, dominant string) []tagEntry {
	var primary []tagEntry
	seen := map[string]bool{}

	prefer := []string{
		"recommendation_system", "research_ability", "experiment_ability", "modeling_ability",
		"algorithm", "data_analysis_ability", "nlp", "multimodal",
	}

	if dominant != "algorithm_research" {
		prefer = []string{"backend", "backend_service", "engineering_ability", "algorithm", "data_analysis_ability", "product_sense", "business_understanding"}
	}

	for _, t := range tags {
		if t.Confidence >= 1 && len(primary) < 8 {
			if !seen[t.Tag] {
				primary = append(primary, t)
				seen[t.Tag] = true
			}
		}
	}

	// Ensure preferred tags appear early
	var ordered []tagEntry
	for _, pref := range prefer {
		for _, p := range primary {
			if p.Tag == pref {
				ordered = append(ordered, p)
			}
		}
	}
	for _, p := range primary {
		found := false
		for _, o := range ordered {
			if o.Tag == p.Tag {
				found = true
				break
			}
		}
		if !found {
			ordered = append(ordered, p)
		}
	}
	if len(ordered) > 6 {
		ordered = ordered[:6]
	}
	return ordered
}

func extractRiskTags(tags []tagEntry, dominant string) []tagEntry {
	var risks []tagEntry
	tagMap := make(map[string]int)
	for _, t := range tags {
		tagMap[t.Tag] = t.Confidence
	}

	if dominant == "algorithm_research" {
		if tagMap["backend_service"] >= 2 {
			risks = append(risks, tagEntry{Tag: "over_general_skills", Group: "risk", Confidence: 2, Evidence: "工程技能可能冲淡算法主方向"})
		}
		if tagMap["engineering_ability"] < 2 {
			risks = append(risks, tagEntry{Tag: "lack_engineering_depth", Group: "risk", Confidence: 1, Evidence: "缺乏完整后端系统项目经验"})
		}
	} else {
		if tagMap["recommendation_system"] == 0 && tagMap["nlp"] == 0 && tagMap["multimodal"] == 0 {
			risks = append(risks, tagEntry{Tag: "lack_algorithm_project", Group: "risk", Confidence: 2, Evidence: "缺少算法方向深度项目证据"})
		}
	}
	if tagMap["research_ability"] < 1 {
		risks = append(risks, tagEntry{Tag: "lack_research_depth", Group: "risk", Confidence: 2, Evidence: "缺少科研论文或实验方向"})
	}
	return risks
}

func buildProfileSummary(dominant string, tags []tagEntry) string {
	tagMap := make(map[string]int)
	for _, t := range tags {
		tagMap[t.Tag] = t.Confidence
	}

	switch dominant {
	case "algorithm_research":
		parts := []string{}
		if tagMap["recommendation_system"] >= 2 {
			parts = append(parts, "推荐系统")
		}
		if tagMap["nlp"] >= 2 {
			parts = append(parts, "NLP")
		}
		if tagMap["multimodal"] >= 2 {
			parts = append(parts, "多模态")
		}
		if tagMap["experiment_ability"] >= 2 {
			parts = append(parts, "实验与评估能力")
		}
		if len(parts) > 0 {
			return fmt.Sprintf("算法研究型画像，核心方向：%s。", strings.Join(parts, "、"))
		}
		return "算法研究型画像，具备机器学习和深度学习基础。"
	case "backend_engineering":
		return "后端工程型画像，具备服务端开发和基础架构能力。"
	case "career_exploration":
		return "职业探索型画像，多方向有基础信号，尚无清晰主方向。"
	default:
		return "通用画像。"
	}
}

func buildTextFromResume(resume *model.Resume) string {
	var b strings.Builder
	b.WriteString(resume.Name)
	b.WriteString(" ")
	b.WriteString(resume.ProfessionalSkills)
	b.WriteString(" ")
	b.WriteString(resume.AIApplication)
	for _, e := range resume.Educations {
		b.WriteString(" ")
		b.WriteString(e.School)
		b.WriteString(" ")
		b.WriteString(e.Major)
		b.WriteString(" ")
		b.WriteString(e.ResearchDirection)
		b.WriteString(" ")
		b.WriteString(e.CoreCourses)
	}
	for _, p := range resume.Projects {
		b.WriteString(" ")
		b.WriteString(p.Name)
		b.WriteString(" ")
		b.WriteString(p.Description)
	}
	for _, v := range resume.Internships {
		b.WriteString(" ")
		b.WriteString(v.Position)
		b.WriteString(" ")
		b.WriteString(v.Description)
	}
	b.WriteString(" ")
	b.WriteString(resume.AdditionalInfo)
	return b.String()
}


func dominantProfileV3(resume *model.Resume) string {
    eduT := strings.ToLower(eduText(resume))
    projT := strings.ToLower(projText(resume))
    internT := strings.ToLower(internText(resume))
    skillsT := strings.ToLower(resume.ProfessionalSkills + " " + resume.AIApplication)
    allT := strings.ToLower(resume.RawText)
    if allT == "" { allT = strings.ToLower(buildTextFromResume(resume)) }
    a := cm(eduT+projT+internT, "推荐系统","搜索推荐","ctr","点击率预估","排序模型","消融实验","表示学习","自然语言处理","nlp","pytorch","ndcg","recall","auc","深度学习","数据挖掘")
    be := cm(projT+internT+skillsT, "gin","spring boot","jwt","接口开发","后端开发","服务端","权限校验","缓存","限流","部署","前后端联调","数据库表结构","登录鉴权","rabbitmq","swagger")
    prod := cm(allT, "产品需求分析","用户访谈","竞品分析","prd","原型设计","axure","figma","墨刀","用户体验","业务流程","产品经理","产品运营")
    biz := cm(allT, "商务","采购","供应链","解决方案","合同","供应商","社会调研","公共政策")
    if a>=2 && a>be && a>prod { return "algorithm_research" }
	log.Printf("[Profile] V3 scores a=%d be=%d prod=%d biz=%d resume=%d", a, be, prod, biz, resume.ID)
    if be>=3 && be>a && be>prod { return "backend_engineering" }
    if prod>=3 && prod>=a && prod>=be { return "product_operation" }
    if biz>=2 && biz>a && biz>be { return "business_solution" }
    return "career_exploration"
}

func cm(text string, phrases ...string) int { c:=0; for _,p:=range phrases { if strings.Contains(text, p) {c++} }; return c }

func eduText(r *model.Resume) string { var p []string; for _,e:=range r.Educations { p=append(p,e.ResearchDirection+" "+e.PaperSummary+" "+e.CoreCourses) }; return strings.Join(p," ") }
func projText(r *model.Resume) string { var p []string; for _,x:=range r.Projects { p=append(p,x.Name+" "+x.Role+" "+x.Description) }; return strings.Join(p," ") }
func internText(r *model.Resume) string { var p []string; for _,x:=range r.Internships { p=append(p,x.Company+" "+x.Position+" "+x.Description) }; return strings.Join(p," ") }
