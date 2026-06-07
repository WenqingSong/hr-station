package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"offer-catcher/server/internal/model"
)

// ProfileProvider is the minimal interface needed by MatchMockAgent for profile access.
type ProfileProvider interface {
	EnsureProfile(ctx context.Context, userID, resumeID uint64, profileType string) (*model.ResumeProfile, string, error)
}

type MatchMockAgent struct {
	profileProvider ProfileProvider
}

func NewMatchMockAgent() *MatchMockAgent {
	return &MatchMockAgent{}
}

func (a *MatchMockAgent) SetProfileProvider(pp ProfileProvider) { a.profileProvider = pp }

func (a *MatchMockAgent) Analyze(ctx context.Context, resume *model.ResumeVersion, job *model.Job) (*model.MatchAnalysisResult, error) {
	cp := a.getProfile(ctx, resume)
	score := a.calcScore(cp, resume, job)
	level := matchLevel(score)
	fitPoints := buildFitPointsV2(cp, resume, job, 3)
	risks := buildMatchRiskPoints(cp, job)
	improvements := buildMatchImprovements(cp, job)

	return &model.MatchAnalysisResult{
		JobID:         job.ID, JobTitle: job.Title, ResumeID: resume.ID,
		MatchScore: score, MatchLevel: level, ScreeningResult: level,
		MatchHighlights: fitPoints, RiskPoints: risks,
		OptimizationSuggestions: toMatchSuggestions(improvements),
		ProfileType:            "fast",
		ProfileDominantProfile: cp.dominant,
		GeneratedBy: "mock_agent",
	}, nil
}

type matchProfile struct {
	dominant string
	tags     tagSet
}

func (a *MatchMockAgent) getProfile(ctx context.Context, resume *model.ResumeVersion) matchProfile {
	cp := matchProfile{tags: make(tagSet)}
	if a.profileProvider != nil && resume.ID > 0 {
		p, st, err := a.profileProvider.EnsureProfile(ctx, resume.UserID, resume.ID, model.ProfileTypeFast)
		if err == nil && st == model.ProfileStatusReady && p != nil && p.DominantProfile != "" {
			cp.dominant = p.DominantProfile
			if p.PrimaryTagsJSON != "" {
				var tags []model.ProfileTag
				if json.Unmarshal([]byte(p.PrimaryTagsJSON), &tags) == nil {
					for _, t := range tags {
						cp.tags.set(t.Tag, t.Confidence)
					}
				}
			}
			return cp // K9-2-C-Fix: return immediately, don't fall through to fallback
		}
	}
	// Fallback: extract from resume
	cp2 := extractCandidateTags(resume)
	cp.dominant = cp2.dominantProfile
	cp.tags = cp2.tags
	return cp
}

func (a *MatchMockAgent) calcScore(cp matchProfile, resume *model.ResumeVersion, job *model.Job) int {
	jp := extractJobTags(job)
	dirF := directionFit(cp.dominant, jp.tags)

	// K9-2-C-Fix: override directionFit when extractJobTags assigns wrong tag
	// (e.g. "软件开发-后台开发方向" gets algorithm_core due to if-else chain, not backend_service)
	jobText := strings.ToLower(job.Title + " " + job.Category + " " + job.Department)
	if cp.dominant == "backend_engineering" {
		if strings.Contains(jobText, "后台") || strings.Contains(jobText, "后端") ||
			strings.Contains(jobText, "服务端") || strings.Contains(jobText, "全栈") {
			dirF = "strong"
		}
	}
	if cp.dominant == "algorithm_research" {
		if strings.Contains(jobText, "算法") || strings.Contains(jobText, "nlp") ||
			strings.Contains(jobText, "数据科学") || strings.Contains(jobText, "多模态") ||
			strings.Contains(jobText, "机器学习") || strings.Contains(jobText, "ai") {
			dirF = "strong"
		}
	}

	tmScore := tagMatrixScore(cp.tags, jp.tags)

	score := 50
	switch dirF {
	case "strong":
		score += 25
	case "medium":
		score += 15
	case "weak":
		score += 5
	}

	tmCapped := tmScore
	if tmCapped > 30 {
		tmCapped = 30
	}
	score += tmCapped

	if cp.tags.has("project_evidence", 2) {
		score += 4
	}
	if cp.tags.has("internship_evidence", 2) {
		score += 3
	}

	// K9-2-C-Fix: backend profile gets unconditional direction boost
	if cp.dominant == "backend_engineering" && dirF == "strong" {
		score += 10 // backend profile + strong direction = enough evidence
	}
	// K9-2-C-Fix: algorithm profile also gets boost for algo jobs  
	if cp.dominant == "algorithm_research" && dirF == "strong" {
		score += 5 // algorithm profile + strong direction
	}

	// conflict cap — with K9-2-C-Fix exception for actual domain matches
	isAlgo := jp.tags.has("algorithm_core", 2) || jp.tags.has("nlp", 2) || jp.tags.has("multimodal", 2)
	isBackend := jp.tags.has("backend_service", 2)
	// K9-2-C-Fix: when job title is actually backend but extractJobTags misidentifies as algo, skip cap
	isActuallyBackendJob := strings.Contains(jobText, "后台") || strings.Contains(jobText, "后端") || strings.Contains(jobText, "服务端")
	switch cp.dominant {
	case "algorithm_research":
		if isBackend && score > 74 {
			score = 74
		}
	case "backend_engineering":
		if isAlgo && !isActuallyBackendJob && score > 74 {
			score = 74
		}
		if isAlgo && isActuallyBackendJob && score > 89 {
			score = 89 // allow higher for actually backend jobs
		}
	case "product_operation":
		if isAlgo || isBackend {
			if score > 69 {
				score = 69
			}
		}
	case "business_solution":
		if isAlgo || isBackend {
			if score > 69 {
				score = 69
			}
		}
	}

	if score > 95 {
		score = 95
	}
	if score < 35 {
		score = 35
	}
	return score
}

func matchLevel(score int) string {
	switch {
	case score >= 90:
		return "强匹配，优先投递"
	case score >= 80:
		return "高匹配，建议重点投递"
	case score >= 70:
		return "可尝试，需要补强"
	case score >= 60:
		return "备选参考，不建议优先"
	default:
		return "不建议投递"
	}
}

func buildMatchSummary(cp matchProfile, resume *model.ResumeVersion, job *model.Job, score int) string {
	domain := domainLabel(detectJobDomain(extractJobTags(job)))
	projCount := len(resume.Projects)
	internCount := len(resume.Internships)
	if score >= 80 {
		return fmt.Sprintf("这份简历与%s岗位在技能、项目经验上匹配度较高（%d分），建议重点投递。共%d个项目、%d段实习经验可支撑投递。", domain, score, projCount, internCount)
	}
	if score >= 70 {
		return fmt.Sprintf("这份简历与%s岗位存在一定匹配，部分技能可迁移，但需在投递前针对性补强关键证据。", domain)
	}
	return fmt.Sprintf("这份简历与该岗位目前匹配度偏低，建议优先提升相关技能或寻找更匹配的方向。")
}

func buildFitPointsV2(cp matchProfile, resume *model.ResumeVersion, job *model.Job, maxN int) []string {
	var pts []string
	t := strings.ToLower(job.Title + " " + job.Category)
	projText := strings.ToLower(projectsText(resume))
	internText := strings.ToLower(internshipsText(resume))

	if cp.tags.has("modeling_ability", 2) && (cp.dominant == "algorithm_research") {
		pts = append(pts, "具备深度学习建模和模型训练经验，可支撑算法岗位要求")
	}
	if cp.tags.has("experiment_ability", 2) {
		pts = append(pts, "具备实验复现和指标评估经验")
	}
	if cp.tags.has("engineering_ability", 2) {
		pts = append(pts, "具备工程开发和系统设计能力")
	}
	if cp.tags.has("backend_service", 2) && strings.Contains(strings.ToLower(job.Title+" "+job.Category), "后端") {
		pts = append(pts, "已具备后端开发和数据库使用经验，与岗位工程要求匹配")
	}
	if cp.dominant == "backend_engineering" {
		backendSkills := []string{"go", "mysql", "redis", "docker", "gin", "接口", "api", "权限", "部署"}
		found := make(map[string]bool)
		for _, kw := range backendSkills {
			if strings.Contains(projText, kw) {
				found[kw] = true
			}
		}
		if len(found) >= 2 {
			pts = append(pts, fmt.Sprintf("项目经历覆盖%d类后端技术栈（接口、数据库、部署等），与岗位要求吻合", len(found)))
		} else if strings.Contains(projText, "go") || strings.Contains(projText, "mysql") {
			pts = append(pts, "项目经历包含Go和数据库实践，具备后端开发基础")
		}
	}
	if len(internText) > 10 {
		pts = append(pts, fmt.Sprintf("具有%d段实习经历，具备真实工作场景经验", len(resume.Internships)))
	}
	if len(pts) == 0 {
		pts = append(pts, fmt.Sprintf("简历背景与%s岗位方向具有一定关联性", t))
	}
	if len(pts) > maxN {
		pts = pts[:maxN]
	}
	return pts
}

func buildMatchResumeEvidence(resume *model.ResumeVersion) []string {
	var ev []string
	projText := strings.ToLower(projectsText(resume))
	for _, p := range resume.Projects {
		if len(ev) >= 3 {
			break
		}
		if p.Name != "" {
			ev = append(ev, p.Name+"项目经验")
		}
	}
	for _, i := range resume.Internships {
		if len(ev) >= 3 {
			break
		}
		if i.Position != "" {
			ev = append(ev, i.Position+"实习经历")
		}
	}
	// Add tech detail if skills detected
	backendSkills := []string{"go", "mysql", "redis", "docker", "linux", "nginx"}
	for _, kw := range backendSkills {
		if len(ev) >= 3 {
			break
		}
		if strings.Contains(projText, kw) && !containsName(ev, kw) {
			if kw == "go" {
				ev = append(ev, "Go 开发语言经验")
			} else {
				ev = append(ev, strings.ToUpper(kw)+" 使用经验")
			}
		}
	}
	if len(ev) == 0 {
		ev = append(ev, "简历包含相关学习与项目经历")
	}
	return ev
}

func containsName(ev []string, kw string) bool {
	kl := strings.ToLower(kw)
	for _, e := range ev {
		if strings.Contains(strings.ToLower(e), kl) {
			return true
		}
	}
	return false
}

func buildJobRequirements(job *model.Job) []string {
	var reqs []string
	t := strings.ToLower(job.Title + " " + job.Category + " " + job.Requirements)
	if strings.Contains(t, "后端") {
		reqs = append(reqs, "后端接口开发与系统设计能力")
	}
	if strings.Contains(t, "数据库") || strings.Contains(t, "mysql") || strings.Contains(t, "redis") {
		reqs = append(reqs, "数据库与缓存使用经验")
	}
	if strings.Contains(t, "算法") || strings.Contains(t, "模型") {
		reqs = append(reqs, "算法模型训练与实验评估能力")
	}
	if strings.Contains(t, "产品") || strings.Contains(t, "运营") {
		reqs = append(reqs, "需求分析与用户研究能力")
	}
	if len(reqs) == 0 {
		reqs = append(reqs, "与岗位方向相关的工作能力与经验")
	}
	return reqs
}

func buildGapPoints(cp matchProfile, job *model.Job) []string {
	var gaps []string
	jp := extractJobTags(job)
	d := detectJobDomain(jp)

	switch cp.dominant {
	case "algorithm_research":
		switch d {
		case domainAlgorithmNLP:
			if !cp.tags.has("nlp", 2) {
				gaps = append(gaps, "缺少NLP专项项目或论文经历")
			}
		case domainAlgorithmMultimodal:
			if !cp.tags.has("multimodal", 2) {
				gaps = append(gaps, "缺少多模态或计算机视觉项目经历")
			}
		case domainAIApplication:
			if !cp.tags.has("engineering_ability", 2) {
				gaps = append(gaps, "缺少AI工程化部署和线上服务经验")
			}
		}
	case "backend_engineering":
		if d == domainClient {
			gaps = append(gaps, "缺少客户端开发经验")
		}
	}
	if len(gaps) == 0 {
		gaps = append(gaps, "建议查看岗位详细要求，针对性补充相关项目经历")
	}
	return gaps
}

func buildMatchRiskPoints(cp matchProfile, job *model.Job) []string {
	var risks []string
	jp := extractJobTags(job)
	dirF := directionFit(cp.dominant, jp.tags)
	if dirF == "weak" {
		risks = append(risks, "岗位方向与你的主要技能方向不完全一致，投递竞争力可能受限")
	}
	if dirF == "strong" && cp.dominant == "algorithm_research" && !cp.tags.has("research_ability", 2) {
		risks = append(risks, "算法岗位通常偏好较强的研究经历，当前简历研究证据不足")
	}
	if len(risks) == 0 {
		risks = append(risks, "建议补充与岗位更匹配的项目细节，提高初筛通过率")
	}
	return risks
}

func buildMatchImprovements(cp matchProfile, job *model.Job) []string {
	var imps []string
	jp := extractJobTags(job)
	d := detectJobDomain(jp)

	switch cp.dominant {
	case "algorithm_research":
		switch d {
		case domainAlgorithmNLP:
			imps = append(imps, "建议补充NLP方向项目（文本分类/情感分析/问答系统），展示Transformer使用经验")
		case domainAlgorithmMultimodal:
			imps = append(imps, "建议补充图文检索或视频理解项目，说明数据集、模型结构和评估指标")
		case domainAIApplication:
			imps = append(imps, "建议补充AI服务化项目：模型接口封装、Docker部署、性能优化")
		default:
			imps = append(imps, "建议继续深化算法方向项目，补充模型实验与评估证据")
		}
	case "backend_engineering":
		imps = append(imps, "建议在项目经历中补充接口QPS、缓存策略、权限模型和部署架构")
	case "product_operation":
		imps = append(imps, "建议补充产品数据分析、用户调研或A/B测试相关经验")
	case "business_solution":
		imps = append(imps, "建议补充行业分析、客户需求沟通或方案策划相关经验")
	default:
		imps = append(imps, "建议针对目标岗位补强相关技能和项目经历")
	}
	return imps
}

func buildTalkingPoints(cp matchProfile, resume *model.ResumeVersion, job *model.Job) []string {
	var pts []string
	for _, p := range resume.Projects {
		if len(pts) >= 2 {
			break
		}
		if p.Name != "" {
			pts = append(pts, fmt.Sprintf("可以重点讲%s项目中你的技术方案和具体成果", p.Name))
		}
	}
	if len(pts) == 0 {
		pts = append(pts, "可以重点介绍你的项目经历和解决的实际问题")
	}
	return pts
}

func buildDeliveryStrategy(cp matchProfile, resume *model.ResumeVersion, job *model.Job, score int) string {
	category := job.Category
	if score >= 80 {
		return fmt.Sprintf("建议优先投递%s类岗位，投递前把核心技术栈和项目成果放在简历前半部分。", category)
	}
	return fmt.Sprintf("建议先补强与%s岗位相关的技能和项目经历，再投递以提高成功率。", category)
}

func toMatchSuggestions(strs []string) []model.MatchSuggestion {
	var s []model.MatchSuggestion
	for i, str := range strs {
		prio := "P1"
		if i == 0 {
			prio = "P0"
		}
		s = append(s, model.MatchSuggestion{Priority: prio, Title: str, Action: str})
	}
	return s
}
