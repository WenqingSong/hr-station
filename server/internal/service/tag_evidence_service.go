// ─── K13-Fix-A: Tag Evidence Service ───────────────────────────────────────
// Provides unified tag extraction and evidence generation for resumes and jobs.
// Serves both Fast (rule) and Deep Agent 1 (LLM-enhanced) paths.

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sort"
	"strings"

	"offer-catcher/server/internal/model"
	"offer-catcher/server/internal/repository"
)

// ─── TagEvidenceService ──────────────────────────────────────────────────

// TagEvidenceService is the unified Tag + Evidence generation service.
// It uses rule-based extraction as primary path, with optional LLM enhancement.
type TagEvidenceService struct {
	resumeRepo    *repository.ResumeRepository
	profileRepo   *repository.ResumeProfileRepository
	jobProfileRepo *repository.JobProfileRepository
	jobRepo       interface{ FindByID(context.Context, uint64) (*model.Job, error) }
}

// NewTagEvidenceService creates a new TagEvidenceService.
func NewTagEvidenceService(
	resumeRepo *repository.ResumeRepository,
	profileRepo *repository.ResumeProfileRepository,
	jobProfileRepo *repository.JobProfileRepository,
) *TagEvidenceService {
	return &TagEvidenceService{
		resumeRepo:     resumeRepo,
		profileRepo:    profileRepo,
		jobProfileRepo: jobProfileRepo,
	}
}

// SetJobRepo injects a job repository for job profile generation.
func (s *TagEvidenceService) SetJobRepo(jobRepo interface{ FindByID(context.Context, uint64) (*model.Job, error) }) {
	s.jobRepo = jobRepo
}

// ─── Resume Tag + Evidence ───────────────────────────────────────────────

// EnsureResumeTagProfile generates or loads a resume tag profile with evidence.
// Returns the unified ResumeTagProfile structure.
func (s *TagEvidenceService) EnsureResumeTagProfile(ctx context.Context, resumeID, userID uint64) (*model.ResumeTagProfile, error) {
	// 1. Load resume with structured data
	resume, err := s.resumeRepo.GetByID(ctx, userID, resumeID)
	if err != nil {
		return nil, fmt.Errorf("ensure resume tag: %w", err)
	}

	// 2. Extract tags (reuse existing extraction logic)
	text := resume.RawText
	if text == "" {
		text = buildTextFromResume(resume)
	}
	tags := extractTags(text)
	dominant := dominantProfileV3(resume)
	primaryTags := buildPrimaryTags(tags, dominant)
	riskTags := extractRiskTags(tags, dominant)

	// 3. Build skill/project/internship tags from structured data
	skillTags := extractSkillTags(resume)
	projectTags := extractProjectTags(resume)
	internshipTags := extractInternshipTags(resume)
	educationTags := extractEducationTags(resume)
	intentTags := extractIntentTags(resume)

	// 4. K13: Build real evidence (NOT "{}")
	evidence := buildResumeEvidence(resume, text)

	// 5. Persist to resume_profiles (reuse existing table)
	now := true // use NOW() in DB
	_ = now
	profile := &model.ResumeProfile{
		ResumeID:        resumeID,
		UserID:          userID,
		ProfileType:     model.ProfileTypeFast,
		ProfileVersion:  model.ProfileVersionV1,
		Status:          model.ProfileStatusReady,
		DominantProfile: dominant,
		PrimaryTagsJSON: marshalTagsToJSON(convertToProfileTags(primaryTags)),
		CandidateTagsJSON: marshalTagsToJSON(convertToProfileTags(tags)),
		RiskTagsJSON:    marshalTagsToJSON(convertToProfileTags(riskTags)),
		EvidenceJSON:    mustMarshalJSON(evidence),
		ProfileSummary:  buildProfileSummary(dominant, tags),
		Source:          "rule",
		Generator:       "tag_agent_v1",
	}
	if err := s.profileRepo.UpsertProfile(ctx, profile); err != nil {
		log.Printf("[TagEvidence] failed to persist resume profile resume=%d: %v", resumeID, err)
	}

	// 6. Return unified structure
	return &model.ResumeTagProfile{
		ResumeID:        resumeID,
		DominantProfile: dominant,
		ProfileSummary:  profile.ProfileSummary,
		PrimaryTags:     convertTagEntries(primaryTags),
		CandidateTags:   convertTagEntries(tags),
		SkillTags:       skillTags,
		ProjectTags:     projectTags,
		InternshipTags:  internshipTags,
		EducationTags:   educationTags,
		RiskTags:        convertTagEntries(riskTags),
		IntentTags:      intentTags,
		Evidence:        evidence,
		Source:          "rule",
		Version:         model.TagEvidenceVersionV1,
	}, nil
}

// GetResumeTagProfile loads a cached resume tag profile.
func (s *TagEvidenceService) GetResumeTagProfile(ctx context.Context, resumeID, userID uint64) (*model.ResumeTagProfile, error) {
	profile, err := s.profileRepo.FindByResumeIDAndType(ctx, resumeID, model.ProfileTypeFast)
	if err != nil || profile == nil {
		return s.EnsureResumeTagProfile(ctx, resumeID, userID)
	}
	// Regenerate if evidence is empty
	if profile.EvidenceJSON == "" || profile.EvidenceJSON == "{}" || profile.EvidenceJSON == "null" {
		return s.EnsureResumeTagProfile(ctx, resumeID, userID)
	}
	return s.EnsureResumeTagProfile(ctx, resumeID, userID)
}

// ─── Job Tag + Evidence ──────────────────────────────────────────────────

// EnsureJobTagProfile generates or loads a job tag profile with evidence.
func (s *TagEvidenceService) EnsureJobTagProfile(ctx context.Context, jobID uint64) (*model.JobTagProfile, error) {
	// 1. Check existing
	existing, err := s.jobProfileRepo.FindByJobID(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("ensure job tag: %w", err)
	}
	if existing != nil && existing.EvidenceJSON != "" && existing.EvidenceJSON != "{}" {
		// Return cached
		return s.jobProfileToTagProfile(existing), nil
	}

	// 2. Load job from DB
	if s.jobRepo == nil {
		return nil, fmt.Errorf("job repo not wired for tag evidence service")
	}
	job, err := s.jobRepo.FindByID(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("ensure job tag: load job: %w", err)
	}

	// 3. Extract tags
	jdText := job.Description + " " + job.Requirements
	jTags := extractJobTags(jdText, job.Direction, job.Category)
	primaryTags := jobPrimaryTags(jTags)
	skillTags := jobSkillTags(jTags)
	reqTags := jobReqTags(jTags)
	respTags := jobRespTags(jTags)
	riskTags := jobRiskTags(jTags)

	// 4. Build evidence
	evidence := buildJobEvidence(job, jdText)

	// 5. Persist
	primaryJSON := mustMarshalJSON(convertToProfileTags(primaryTags))
	skillJSON := mustMarshalJSON(convertToProfileTags(skillTags))
	reqJSON := mustMarshalJSON(convertToProfileTags(reqTags))
	respJSON := mustMarshalJSON(convertToProfileTags(respTags))
	riskJSON := mustMarshalJSON(convertToProfileTags(riskTags))
	evidenceJSON := mustMarshalJSON(evidence)

	jp := &model.JobProfile{
		JobID:                  jobID,
		Direction:              job.Direction,
		Category:               job.Category,
		PrimaryTagsJSON:        primaryJSON,
		SkillTagsJSON:          skillJSON,
		RequirementTagsJSON:    reqJSON,
		ResponsibilityTagsJSON: respJSON,
		RiskTagsJSON:           riskJSON,
		EvidenceJSON:           evidenceJSON,
		Source:                 "rule",
		Version:                model.TagEvidenceVersionV1,
	}
	if err := s.jobProfileRepo.Upsert(ctx, jp); err != nil {
		log.Printf("[TagEvidence] failed to persist job profile job=%d: %v", jobID, err)
	}

	return &model.JobTagProfile{
		JobID:              jobID,
		Direction:          job.Direction,
		Category:           job.Category,
		PrimaryTags:        convertTagEntries(primaryTags),
		SkillTags:          convertTagEntries(skillTags),
		RequirementTags:    convertTagEntries(reqTags),
		ResponsibilityTags: convertTagEntries(respTags),
		RiskTags:           convertTagEntries(riskTags),
		Evidence:           evidence,
		Source:             "rule",
		Version:            model.TagEvidenceVersionV1,
	}, nil
}

// GetJobTagProfile loads a cached job tag profile.
func (s *TagEvidenceService) GetJobTagProfile(ctx context.Context, jobID uint64) (*model.JobTagProfile, error) {
	return s.EnsureJobTagProfile(ctx, jobID)
}

// EnsureAllOpenJobTagProfiles generates profiles for all jobs missing profiles.
func (s *TagEvidenceService) EnsureAllOpenJobTagProfiles(ctx context.Context, limit int) error {
	ids, err := s.jobProfileRepo.ListPendingJobIDs(ctx, limit)
	if err != nil {
		return fmt.Errorf("ensure all job tags: %w", err)
	}
	log.Printf("[TagEvidence] ensuring %d pending job profiles", len(ids))
	for _, jid := range ids {
		if _, err := s.EnsureJobTagProfile(ctx, jid); err != nil {
			log.Printf("[TagEvidence] failed to generate job profile job=%d: %v", jid, err)
		}
	}
	return nil
}

// ─── Recall ──────────────────────────────────────────────────────────────

// GetAllJobTagProfiles loads all cached job tag profiles (K13-Fix-C1).
func (s *TagEvidenceService) GetAllJobTagProfiles(ctx context.Context) ([]*model.JobTagProfile, error) {
	allJps, err := s.jobProfileRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all job profiles: %w", err)
	}
	result := make([]*model.JobTagProfile, 0, len(allJps))
	for _, jp := range allJps {
		result = append(result, s.jobProfileToTagProfile(jp))
	}
	return result, nil
}

// RecallJobsByTags returns candidate jobs ranked by tag match score.
func (s *TagEvidenceService) RecallJobsByTags(ctx context.Context, resumeID, userID uint64, limit int) ([]model.JobCandidate, error) {
	if limit <= 0 {
		limit = 20
	}

	// 1. Ensure resume profile
	rp, err := s.EnsureResumeTagProfile(ctx, resumeID, userID)
	if err != nil {
		return nil, fmt.Errorf("recall: %w", err)
	}

	// 2. Load all job profiles
	allJps, err := s.jobProfileRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("recall: load jobs: %w", err)
	}

	// 3. Score each job against resume tags
	var candidates []scoredCandidate

	for _, jp := range allJps {
		jpParsed := s.jobProfileToTagProfile(jp)
		matched, missing, score := computeTagMatch(rp, jpParsed)
		evRefs := collectEvidenceRefs(rp, matched)

		candidates = append(candidates, scoredCandidate{
			candidate: model.JobCandidate{
				JobID:        jpParsed.JobID,
				Direction:    jpParsed.Direction,
				Category:     jpParsed.Category,
				TagScore:     math.Round(score*100) / 100,
				MatchedTags:  matched,
				MissingTags:  missing,
				EvidenceRefs: evRefs,
			},
			score: score,
		})
	}

	// 4. Sort desc and limit
	sortCandidates(candidates)
	if limit > len(candidates) {
		limit = len(candidates)
	}
	result := make([]model.JobCandidate, limit)
	for i := 0; i < limit; i++ {
		result[i] = candidates[i].candidate
	}
	return result, nil
}

// ─── Evidence Builders ───────────────────────────────────────────────────

func buildResumeEvidence(resume *model.Resume, rawText string) model.ResumeEvidence {
	ev := model.ResumeEvidence{
		ProjectCount:    len(resume.Projects),
		InternshipCount: len(resume.Internships),
		SkillCount:      0,
	}

	// Skill evidence
	if resume.ProfessionalSkills != "" {
		skills := strings.Split(resume.ProfessionalSkills, ",")
		ev.SkillCount = len(skills)
		for i, sk := range skills {
			sk = strings.TrimSpace(sk)
			if sk == "" {
				continue
			}
			if len(ev.SkillEvidence) >= 8 {
				break
			}
			ev.SkillEvidence = append(ev.SkillEvidence, model.EvidenceItem{
				ID:       fmt.Sprintf("skill_%d", i),
				Type:     "skill",
				Text:     truncateEvidence(sk, 200),
				Source:   "professional_skills",
				Strength: 0.8,
			})
		}
	}

	// Project evidence
	for i, proj := range resume.Projects {
		if i >= 5 {
			break
		}
		summary := proj.Name
		if proj.Description != "" {
			summary += ": " + proj.Description
		}
		ev.ProjectEvidence = append(ev.ProjectEvidence, model.EvidenceItem{
			ID:       fmt.Sprintf("proj_%d", i),
			Type:     "project",
			Text:     truncateEvidence(summary, 300),
			Source:   "projects",
			Strength: 0.7,
		})
	}

	// Internship evidence
	for i, intern := range resume.Internships {
		if i >= 5 {
			break
		}
		summary := intern.Company
		if intern.Position != "" {
			summary += " - " + intern.Position
		}
		if intern.Description != "" {
			summary += ": " + intern.Description
		}
		ev.InternshipEvidence = append(ev.InternshipEvidence, model.EvidenceItem{
			ID:       fmt.Sprintf("intern_%d", i),
			Type:     "internship",
			Text:     truncateEvidence(summary, 300),
			Source:   "internships",
			Strength: 0.9,
		})
	}

	// Education evidence
	for i, edu := range resume.Educations {
		if i >= 3 {
			break
		}
		summary := fmt.Sprintf("%s %s %s", edu.Degree, edu.Major, edu.School)
		ev.EducationEvidence = append(ev.EducationEvidence, model.EvidenceItem{
			ID:       fmt.Sprintf("edu_%d", i),
			Type:     "education",
			Text:     truncateEvidence(summary, 200),
			Source:   "educations",
			Strength: 0.6,
		})
	}

	// Certificate evidence
	for i, cert := range resume.Certificates {
		if i >= 3 {
			break
		}
		ev.CertificateEvidence = append(ev.CertificateEvidence, model.EvidenceItem{
			ID:       fmt.Sprintf("cert_%d", i),
			Type:     "cert",
			Text:     truncateEvidence(cert.Name, 200),
			Source:   "certificates",
			Strength: 0.5,
		})
	}

	return ev
}

func buildJobEvidence(job *model.Job, jdText string) model.JobEvidence {
	ev := model.JobEvidence{}

	// Responsibility evidence (from description)
	if job.Description != "" {
		sentences := splitIntoSentences(job.Description)
		for i, s := range sentences {
			if i >= 5 {
				break
			}
			ev.ResponsibilityEvidence = append(ev.ResponsibilityEvidence, model.EvidenceItem{
				ID:       fmt.Sprintf("resp_%d", i),
				Type:     "jd_resp",
				Text:     truncateEvidence(s, 300),
				Source:   "description",
				Strength: 0.7,
			})
		}
	}

	// Requirement evidence
	if job.Requirements != "" {
		sentences := splitIntoSentences(job.Requirements)
		for i, s := range sentences {
			if i >= 5 {
				break
			}
			ev.RequirementEvidence = append(ev.RequirementEvidence, model.EvidenceItem{
				ID:       fmt.Sprintf("req_%d", i),
				Type:     "jd_req",
				Text:     truncateEvidence(s, 300),
				Source:   "requirements",
				Strength: 0.9,
			})
		}
	}

	// Keyword evidence (from skill_keywords_json)
	var skillKeywords []string
	if len(job.SkillKeywordsJSON) > 0 {
		_ = json.Unmarshal(job.SkillKeywordsJSON, &skillKeywords)
	}
	for i, kw := range skillKeywords {
		if i >= 10 {
			break
		}
		ev.KeywordEvidence = append(ev.KeywordEvidence, model.EvidenceItem{
			ID:       fmt.Sprintf("kw_%d", i),
			Type:     "jd_kw",
			Text:     truncateEvidence(kw, 100),
			Source:   "skill_keywords",
			Strength: 0.6,
		})
	}

	return ev
}

// ─── Tag Extraction: Jobs ────────────────────────────────────────────────

func extractJobTags(jdText, direction, category string) []tagEntry {
	lower := strings.ToLower(jdText)
	var tags []tagEntry

	// Direction tags
	addTag(&tags, "algorithm", "direction", countContains(lower, "算法", "机器学习", "深度学习", "推荐", "模型", "nlp", "自然语言处理", "数据挖掘", "搜索"), "JD中包含算法相关要求")
	addTag(&tags, "backend", "direction", countContains(lower, "后端", "后台", "服务端", "api", "微服务", "分布式", "高并发")/2, "JD中包含后端开发相关要求")
	addTag(&tags, "frontend", "direction", countContains(lower, "前端", "h5", "小程序", "react", "vue", "javascript")/2, "JD中包含前端相关要求")
	addTag(&tags, "data", "direction", countContains(lower, "数据", "大数据", "spark", "hive", "etl", "数据仓库", "数据分析"), "JD中包含数据相关要求")
	addTag(&tags, "product", "direction", countContains(lower, "产品", "需求分析", "用户体验"), "JD中包含产品相关要求")
	addTag(&tags, "research", "direction", countContains(lower, "研究", "论文", "科研", "发表", "专利"), "JD中包含研究相关要求")
	addTag(&tags, "ai_infra", "direction", countContains(lower, "gpu", "训练", "推理", "部署", "模型优化", "tensorrt"), "JD中包含AI基础设施相关要求")

	// Skill tags
	addTag(&tags, "golang", "skill", countContains(lower, "go ", "golang", "go语言")*2, "JD明确要求Go语言")
	addTag(&tags, "python", "skill", countContains(lower, "python")*2, "JD明确要求Python")
	addTag(&tags, "java", "skill", countContains(lower, "java ") * 2, "JD明确要求Java")
	addTag(&tags, "cpp", "skill", countContains(lower, "c++", "c/c++")*2, "JD明确要求C++")
	addTag(&tags, "mysql", "skill", countContains(lower, "mysql", "数据库")*2, "JD要求数据库能力")
	addTag(&tags, "redis", "skill", countContains(lower, "redis", "缓存")*2, "JD要求缓存能力")
	addTag(&tags, "linux", "skill", countContains(lower, "linux", "操作系统"), "JD要求Linux")
	addTag(&tags, "docker_k8s", "skill", countContains(lower, "docker", "k8s", "kubernetes", "容器")*2, "JD要求容器化能力")
	addTag(&tags, "pytorch_tf", "skill", countContains(lower, "pytorch", "tensorflow")*2, "JD要求深度学习框架")

	// Requirement tags
	addTag(&tags, "degree_bachelor", "requirement", map[bool]int{true: 2}[strings.Contains(lower, "本科") || strings.Contains(lower, "学士")], "要求本科学历")
	addTag(&tags, "degree_master", "requirement", map[bool]int{true: 2}[strings.Contains(lower, "硕士") || strings.Contains(lower, "研究生")], "要求硕士学历")
	addTag(&tags, "internship_exp", "requirement", map[bool]int{true: 2}[strings.Contains(lower, "实习")], "要求实习经验")
	addTag(&tags, "project_exp", "requirement", map[bool]int{true: 2}[strings.Contains(lower, "项目经验") || strings.Contains(lower, "项目经历")], "要求项目经验")
	addTag(&tags, "paper", "requirement", map[bool]int{true: 2}[strings.Contains(lower, "论文") || strings.Contains(lower, "发表")], "要求论文")

	return tags
}

// ─── Job Tag Builders ────────────────────────────────────────────────────

func jobPrimaryTags(tags []tagEntry) []tagEntry {
	var primary []tagEntry
	for _, t := range tags {
		if t.Confidence >= 2 && len(primary) < 6 {
			primary = append(primary, t)
		}
	}
	return primary
}
func jobSkillTags(tags []tagEntry) []tagEntry   { return filterByGroup(tags, "skill") }
func jobReqTags(tags []tagEntry) []tagEntry     { return filterByGroup(tags, "requirement") }
func jobRespTags(tags []tagEntry) []tagEntry    { return filterByGroup(tags, "responsibility") }
func jobRiskTags(tags []tagEntry) []tagEntry {
	var risks []tagEntry
	for _, t := range tags {
		if t.Confidence <= 1 && len(risks) < 4 {
			t.Group = "risk"
			risks = append(risks, t)
		}
	}
	return risks
}

func filterByGroup(tags []tagEntry, group string) []tagEntry {
	var result []tagEntry
	for _, t := range tags {
		if t.Group == group && len(result) < 8 {
			result = append(result, t)
		}
	}
	return result
}

// ─── Tag Match Scoring ───────────────────────────────────────────────────

func computeTagMatch(rp *model.ResumeTagProfile, jp *model.JobTagProfile) (matched, missing []string, score float64) {
	resumeTagNames := make(map[string]bool)
	for _, t := range rp.PrimaryTags {
		resumeTagNames[t.Tag] = true
	}
	for _, t := range rp.SkillTags {
		resumeTagNames[t.Tag] = true
	}
	for _, t := range rp.ProjectTags {
		resumeTagNames[t.Tag] = true
	}
	for _, t := range rp.InternshipTags {
		resumeTagNames[t.Tag] = true
	}

	jobRequireTags := make(map[string]bool)
	for _, t := range jp.PrimaryTags {
		jobRequireTags[t.Tag] = true
	}
	for _, t := range jp.SkillTags {
		jobRequireTags[t.Tag] = true
	}
	for _, t := range jp.RequirementTags {
		jobRequireTags[t.Tag] = true
	}

	totalJobTags := len(jobRequireTags)
	if totalJobTags == 0 {
		return nil, nil, 0
	}

	matchedCount := 0
	for tag := range jobRequireTags {
		if resumeTagNames[tag] {
			matched = append(matched, tag)
			matchedCount++
		} else {
			missing = append(missing, tag)
		}
	}

	score = float64(matchedCount) / float64(totalJobTags) * 100
	// Direction bonus
	if rp.DominantProfile == model.DominantAlgorithmResearch && jp.Direction == "技术研究" {
		score += 10
	}
	if rp.DominantProfile == model.DominantBackendEngineering && jp.Direction == "技术" {
		score += 10
	}

	if score > 100 {
		score = 100
	}
	return
}

func collectEvidenceRefs(rp *model.ResumeTagProfile, matchedTags []string) []string {
	refs := make([]string, 0)
	tagSet := make(map[string]bool)
	for _, t := range matchedTags {
		tagSet[t] = true
	}

	for _, ev := range rp.Evidence.SkillEvidence {
		refs = append(refs, ev.ID)
		if len(refs) >= 3 {
			return refs
		}
	}
	for _, ev := range rp.Evidence.ProjectEvidence {
		refs = append(refs, ev.ID)
		if len(refs) >= 3 {
			return refs
		}
	}
	for _, ev := range rp.Evidence.InternshipEvidence {
		refs = append(refs, ev.ID)
		if len(refs) >= 3 {
			return refs
		}
	}
	return refs
}

// ─── Structured Tag Extraction ───────────────────────────────────────────

func extractSkillTags(resume *model.Resume) []model.TagItem {
	if resume.ProfessionalSkills == "" {
		return nil
	}
	skills := strings.Split(resume.ProfessionalSkills, ",")
	var tags []model.TagItem
	for _, sk := range skills {
		sk = strings.TrimSpace(sk)
		if sk == "" {
			continue
		}
		tags = append(tags, model.TagItem{
			Tag:        strings.ToLower(sk),
			Group:      model.TagGroupSkill,
			Weight:     0.6,
			Confidence: 2,
		})
	}
	return tags
}

func extractProjectTags(resume *model.Resume) []model.TagItem {
	var tags []model.TagItem
	for _, proj := range resume.Projects {
		keywords := extractKeywords(proj.Name + " " + proj.Description)
		for _, kw := range keywords {
			if len(tags) >= 8 {
				break
			}
			tags = append(tags, model.TagItem{
				Tag:        kw,
				Group:      model.TagGroupDomain,
				Weight:     0.5,
				Confidence: 2,
			})
		}
	}
	return tags
}

func extractInternshipTags(resume *model.Resume) []model.TagItem {
	var tags []model.TagItem
	for _, intern := range resume.Internships {
		keywords := extractKeywords(intern.Position + " " + intern.Company + " " + intern.Description)
		for _, kw := range keywords {
			if len(tags) >= 6 {
				break
			}
			tags = append(tags, model.TagItem{
				Tag:        kw,
				Group:      model.TagGroupCapability,
				Weight:     0.7,
				Confidence: 2,
			})
		}
	}
	return tags
}

func extractEducationTags(resume *model.Resume) []model.TagItem {
	var tags []model.TagItem
	for _, edu := range resume.Educations {
		if edu.Degree != "" {
			tags = append(tags, model.TagItem{
				Tag:        strings.ToLower(edu.Degree),
				Group:      model.TagGroupSkill,
				Weight:     0.4,
				Confidence: 2,
			})
		}
		if edu.Major != "" {
			tags = append(tags, model.TagItem{
				Tag:        strings.ToLower(edu.Major),
				Group:      model.TagGroupDomain,
				Weight:     0.4,
				Confidence: 1,
			})
		}
	}
	return tags
}

func extractIntentTags(resume *model.Resume) []model.TagItem {
	var tags []model.TagItem
	if resume.TargetDirection != "" {
		tags = append(tags, model.TagItem{
			Tag:        strings.ToLower(resume.TargetDirection),
			Group:      model.TagGroupDirection,
			Weight:     0.5,
			Confidence: 1,
		})
	}
	return tags
}

// ─── Helpers ─────────────────────────────────────────────────────────────

func extractKeywords(text string) []string {
	lower := strings.ToLower(text)
	var keywords []string
	kwMap := map[string]bool{
		"算法": true, "后端": true, "数据": true, "推荐": true,
		"go": true, "java": true, "python": true, "mysql": true, "redis": true,
		"微服务": true, "docker": true, "k8s": true, "api": true,
		"nlp": true, "多模态": true, "搜索": true, "排序": true,
	}
	for kw := range kwMap {
		if strings.Contains(lower, kw) && len(keywords) < 6 {
			keywords = append(keywords, kw)
		}
	}
	return keywords
}

func convertTagEntries(entries []tagEntry) []model.TagItem {
	result := make([]model.TagItem, len(entries))
	for i, e := range entries {
		result[i] = model.TagItem{
			Tag:        e.Tag,
			Group:      e.Group,
			Weight:     float64(e.Confidence) / 3.0,
			Confidence: e.Confidence,
		}
	}
	return result
}

func convertToProfileTags(entries []tagEntry) []model.ProfileTag {
	result := make([]model.ProfileTag, len(entries))
	for i, e := range entries {
		result[i] = model.ProfileTag{
			Tag:        e.Tag,
			Group:      e.Group,
			Confidence: e.Confidence,
			Evidence:   e.Evidence,
		}
	}
	return result
}

func marshalTagsToJSON(tags []model.ProfileTag) string {
	b, _ := json.Marshal(tags)
	if string(b) == "null" {
		return "[]"
	}
	return string(b)
}

func mustMarshalJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	if string(b) == "null" {
		return "{}"
	}
	return string(b)
}

func truncateEvidence(text string, maxLen int) string {
	runes := []rune(strings.TrimSpace(text))
	if len(runes) <= maxLen {
		return string(runes)
	}
	return string(runes[:maxLen]) + "..."
}

func splitIntoSentences(text string) []string {
	text = strings.TrimSpace(text)
	var sentences []string
	// Split by Chinese/English sentence terminators
	sep := func(r rune) bool {
		return r == '。' || r == '！' || r == '？' || r == '；' || r == '.' || r == '!' || r == '?' || r == '\n'
	}
	parts := strings.FieldsFunc(text, sep)
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if len([]rune(p)) > 3 && len(sentences) < 8 {
			sentences = append(sentences, p)
		}
	}
	return sentences
}

type scoredCandidate struct {
	candidate model.JobCandidate
	score     float64
}

func sortCandidates(candidates []scoredCandidate) {
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})
}

func (s *TagEvidenceService) jobProfileToTagProfile(jp *model.JobProfile) *model.JobTagProfile {
	var primary, skill, req, resp, risk []model.TagItem
	_ = json.Unmarshal([]byte(jp.PrimaryTagsJSON), &primary)
	_ = json.Unmarshal([]byte(jp.SkillTagsJSON), &skill)
	_ = json.Unmarshal([]byte(jp.RequirementTagsJSON), &req)
	_ = json.Unmarshal([]byte(jp.ResponsibilityTagsJSON), &resp)
	_ = json.Unmarshal([]byte(jp.RiskTagsJSON), &risk)
	var evidence model.JobEvidence
	_ = json.Unmarshal([]byte(jp.EvidenceJSON), &evidence)
	return &model.JobTagProfile{
		JobID:              jp.JobID,
		Direction:          jp.Direction,
		Category:           jp.Category,
		PrimaryTags:        safeTagItems(primary),
		SkillTags:          safeTagItems(skill),
		RequirementTags:    safeTagItems(req),
		ResponsibilityTags: safeTagItems(resp),
		RiskTags:           safeTagItems(risk),
		Evidence:           evidence,
		Source:             jp.Source,
		Version:            jp.Version,
	}
}

func safeTagItems(items []model.TagItem) []model.TagItem {
	if items == nil {
		return []model.TagItem{}
	}
	return items
}

// Stub until job repo is wired
var _ = context.Background
