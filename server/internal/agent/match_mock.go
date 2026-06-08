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

// AnalyzeFast generates a simplified fast-tier match analysis.
func (a *MatchMockAgent) AnalyzeFast(ctx context.Context, resume *model.ResumeVersion, job *model.Job) (*model.MatchAnalysisResult, error) {
	cp := a.getProfile(ctx, resume)
	score := a.calcScore(cp, resume, job)
	level := matchLevel(score)
	fitPoints := buildFitPointsV2(cp, resume, job, 2)
	risks := buildMatchRiskPoints(cp, job)
	improvements := buildMatchImprovements(cp, job)
	summary := buildMatchSummary(cp, resume, job, score)

	return &model.MatchAnalysisResult{
		JobID: job.ID, JobTitle: job.Title, ResumeID: resume.ID,
		MatchScore: score, MatchLevel: level, ScreeningResult: level,
		Summary: summary,
		MatchHighlights:         fitPoints,
		RiskPoints:              risks,
		OptimizationSuggestions: toMatchSuggestions(improvements),
		ProfileType:             "fast",
		ProfileDominantProfile:  cp.dominant,
		GeneratedBy:             "mock_agent",
		AgentStatus:             model.AgentStatusMockSuccess,
	}, nil
}

// AnalyzeQuality generates a full quality-tier deep match analysis.
func (a *MatchMockAgent) AnalyzeQuality(ctx context.Context, resume *model.ResumeVersion, job *model.Job) (*model.MatchAnalysisResult, error) {
	cp := a.getProfile(ctx, resume)

	// Feature Cards
	resumeFC := buildMatchResumeFC(resume, cp.dominant)
	jobFC := buildMatchJobFC(job)

	// Hard Guard
	guard := applyMatchHardGuard(resumeFC, jobFC)

	// Score with guard cap
	score := a.calcScore(cp, resume, job)
	if guard.MaxScore < score {
		score = guard.MaxScore
	}
	level := matchLevel(score)
	summary := buildMatchSummary(cp, resume, job, score)

	// Deep fields
	evidenceChain := buildMatchEvidenceChain(cp, resume, job, 6)
	gapPoints := buildMatchGapPointsV2(cp, resumeFC, jobFC)
	fitPoints := buildFitPointsV2(cp, resume, job, 5)
	risks := buildMatchRiskPoints(cp, job)
	improvements := buildMatchImprovements(cp, job)

	topFixItems := buildMatchTopFixItems(cp, resumeFC, jobFC)
	resumeEvidence := buildMatchResumeEvidence(resume)
	jobRequirements := buildJobRequirements(job)

	applyReadiness := buildApplyReadiness(score, guard, resumeFC)
	optimizationStopSignal := buildOptimizationStopSignal(score, gapPoints)
	nextSteps := buildMatchNextSteps(score, resumeFC, jobFC, gapPoints)
	interviewTalkingPoints := buildTalkingPoints(cp, resume, job)
	deliveryStrategy := buildDeliveryStrategyV2(score, job, resumeFC, guard)

	return &model.MatchAnalysisResult{
		JobID: job.ID, JobTitle: job.Title, ResumeID: resume.ID,
		MatchScore: score, MatchLevel: level, ScreeningResult: level,
		Summary:                 summary,
		EvidenceChain:           evidenceChain,
		MatchHighlights:         fitPoints,
		RiskPoints:              risks,
		OptimizationSuggestions: toMatchSuggestions(improvements),
		NextSteps:               nextSteps,

		GapPoints:               gapPoints,
		FitPoints:               fitPoints,
		ResumeEvidence:          resumeEvidence,
		JobRequirements:         jobRequirements,
		InterviewTalkingPoints:  interviewTalkingPoints,
		DeliveryStrategy:        deliveryStrategy,
		InterviewQuestions:      interviewTalkingPoints,
		TopFixItems:             topFixItems,
		ApplyReadiness:          applyReadiness,
		OptimizationStopSignal:  optimizationStopSignal,

		ProfileType:             "quality",
		ProfileDominantProfile:  cp.dominant,
		GeneratedBy:             "mock_quality",
		AgentStatus:             model.AgentStatusMockSuccess,
	}, nil
}

// Deprecated: Use AnalyzeFast or AnalyzeQuality instead.
func (a *MatchMockAgent) Analyze(ctx context.Context, resume *model.ResumeVersion, job *model.Job) (*model.MatchAnalysisResult, error) {
	return a.AnalyzeFast(ctx, resume, job)
}

// ─── Feature Cards ────────────────────────────────────────────────────────────

func buildMatchResumeFC(resume *model.ResumeVersion, dominant string) *model.MatchResumeFeatureCard {
	text := strings.ToLower(resume.RawText)
	if text == "" {
		var sb strings.Builder
		sb.WriteString(strings.ToLower(resume.Name + " "))
		sb.WriteString(strings.ToLower(resume.ProfessionalSkills + " "))
		sb.WriteString(strings.ToLower(resume.AIApplication + " "))
		sb.WriteString(strings.ToLower(resume.AdditionalInfo + " "))
		text = sb.String()
	}
	// Also check structured fields
	eduText := strings.ToLower(educationsText(resume))
	projText := strings.ToLower(projectsText(resume))
	internText := strings.ToLower(internshipsText(resume))
	fullText := text + " " + eduText + " " + projText + " " + internText

	card := &model.MatchResumeFeatureCard{
		ResumeID:         resume.ID,
		PrimaryDirection: dominant,
	}

	// Signal detection
	hasAlgo := strings.Contains(fullText, "pytorch") || strings.Contains(fullText, "tensorflow") || strings.Contains(fullText, "深度学习") || strings.Contains(fullText, "算法") || strings.Contains(fullText, "推荐系统") || strings.Contains(fullText, "排序模型") || strings.Contains(fullText, "ctr") || strings.Contains(fullText, "ndcg") || strings.Contains(fullText, "auc") || strings.Contains(fullText, "nlp") || strings.Contains(fullText, "多模态") || strings.Contains(fullText, "神经网络")
	hasBackend := strings.Contains(fullText, "后端") || strings.Contains(fullText, "后台开发") || strings.Contains(fullText, "mysql") || strings.Contains(fullText, "redis") || strings.Contains(fullText, "docker") || strings.Contains(fullText, "接口开发") || strings.Contains(fullText, "gin") || strings.Contains(fullText, "goframe") || strings.Contains(fullText, "spring") || strings.Contains(fullText, "部署") || strings.Contains(fullText, "数据库") || strings.Contains(fullText, "系统设计")
	hasProduct := strings.Contains(fullText, "产品经理") || strings.Contains(fullText, "需求分析") || strings.Contains(fullText, "prd") || strings.Contains(fullText, "原型设计") || strings.Contains(fullText, "用户研究") || strings.Contains(fullText, "axure") || strings.Contains(fullText, "运营") || strings.Contains(fullText, "活动策划")
	hasPublicSvc := strings.Contains(fullText, "公益") || strings.Contains(fullText, "公共服务") || strings.Contains(fullText, "社会调研") || strings.Contains(fullText, "组织支持") || strings.Contains(fullText, "教育公益") || strings.Contains(fullText, "社区服务") || strings.Contains(fullText, "问卷调查") || strings.Contains(fullText, "访谈")
	hasBiz := strings.Contains(fullText, "商务分析") || strings.Contains(fullText, "采购管理") || strings.Contains(fullText, "供应链") || strings.Contains(fullText, "客户需求") || strings.Contains(fullText, "合同") || strings.Contains(fullText, "方案") || strings.Contains(fullText, "行业研究")

	// Fallback to dominant direction when text detection fails
	if !hasAlgo && dominant == "algorithm_research" {
		hasAlgo = true
	}
	if !hasBackend && dominant == "backend_engineering" {
		hasBackend = true
	}
	if !hasProduct && dominant == "product_operation" {
		hasProduct = true
	}
	if !hasPublicSvc && (dominant == "public_service_research" || dominant == "business_solution") {
		hasBiz = true
	}

	card.IsTechnical = hasAlgo || hasBackend
	card.IsAlgorithm = hasAlgo
	card.IsBackend = hasBackend
	card.IsProduct = hasProduct
	card.IsPublicService = hasPublicSvc && !hasAlgo && !hasBackend
	card.IsBusinessSolution = hasBiz || hasPublicSvc
	card.IsExploration = !hasAlgo && !hasBackend && !hasProduct && !hasPublicSvc && !hasBiz

	// Detailed capabilities
	if hasAlgo {
		card.StrongCapabilities = append(card.StrongCapabilities, "算法建模", "深度学习", "实验评估")
		card.EvidenceSummary = append(card.EvidenceSummary, "算法项目经历", "模型训练与评估")
	}
	if hasBackend {
		card.StrongCapabilities = append(card.StrongCapabilities, "后端开发", "数据库设计", "接口开发")
		card.EvidenceSummary = append(card.EvidenceSummary, "后端项目实践", "工程部署经验")
	}
	if hasProduct {
		card.StrongCapabilities = append(card.StrongCapabilities, "需求分析", "产品设计", "项目推进")
		card.EvidenceSummary = append(card.EvidenceSummary, "产品策划经历", "用户需求理解")
	}
	if hasPublicSvc {
		card.StrongCapabilities = append(card.StrongCapabilities, "社会调研", "报告写作", "项目协调")
		card.EvidenceSummary = append(card.EvidenceSummary, "调研分析经验", "公共服务项目")
	}
	if hasBiz {
		card.StrongCapabilities = append(card.StrongCapabilities, "商务沟通", "方案表达", "客户需求")
		card.EvidenceSummary = append(card.EvidenceSummary, "商业分析经验", "方案策划能力")
	}

	// Weak capabilities
	if !hasAlgo {
		card.WeakCapabilities = append(card.WeakCapabilities, "算法建模")
	}
	if !hasBackend {
		card.WeakCapabilities = append(card.WeakCapabilities, "后端开发")
	}

	// Project signals
	for _, p := range resume.Projects {
		if len(card.ProjectSignals) >= 3 { break }
		if p.Name != "" {
			card.ProjectSignals = append(card.ProjectSignals, p.Name)
		}
	}
	// Internship signals
	for _, i := range resume.Internships {
		if len(card.InternshipSignals) >= 3 { break }
		if i.Company != "" {
			card.InternshipSignals = append(card.InternshipSignals, i.Company)
		}
	}

	// Negative preferences for exploration/public-service
	if card.IsExploration || (card.IsPublicService && !card.IsTechnical) {
		card.NegativePreferences = append(card.NegativePreferences, "不适合快节奏技术密集型岗位")
	}

	if len(card.StrongCapabilities) == 0 {
		card.StrongCapabilities = []string{"基础学习能力"}
	}

	return card
}

func buildMatchJobFC(job *model.Job) *model.MatchJobFeatureCard {
	title := strings.ToLower(job.Title)
	desc := strings.ToLower(job.Description + " " + job.Requirements)
	card := &model.MatchJobFeatureCard{
		JobID:      job.ID,
		Title:      job.Title,
		Direction:  job.Direction,
		Category:  job.Category,
		Department: job.Department,
	}

	card.IsAlgorithm = strings.Contains(title, "算法") || strings.Contains(title, "nlp") || strings.Contains(title, "自然语言") || strings.Contains(title, "推荐") || strings.Contains(title, "数据科学") || strings.Contains(title, "机器学习") || strings.Contains(title, "ai")
	card.IsComputerVision = strings.Contains(title, "计算机视觉") || strings.Contains(title, "视觉") || strings.Contains(title, "图像") || strings.Contains(title, "cv")
	card.IsBackend = strings.Contains(title, "后台") || strings.Contains(title, "后端") || strings.Contains(title, "全栈") || strings.Contains(title, "平台") || strings.Contains(desc, "后端") || strings.Contains(title, "软件开发")
	card.IsProduct = strings.Contains(title, "产品") || strings.Contains(title, "用户研究") || strings.Contains(title, "运营")
	card.IsUserResearch = strings.Contains(title, "用户研究") || strings.Contains(title, "调研") || strings.Contains(title, "行业研究")
	card.IsProjectManagement = strings.Contains(title, "项目管理") || strings.Contains(title, "项目") || strings.Contains(title, "组织")
	card.IsBusinessSolution = strings.Contains(title, "解决方案") || strings.Contains(title, "商务") || strings.Contains(title, "方案") || strings.Contains(title, "行业咨询") || strings.Contains(title, "项目管理")
	card.IsPublicServiceLike = strings.Contains(title, "公共") || strings.Contains(title, "公益") || strings.Contains(title, "教育") || strings.Contains(title, "内容研究") || strings.Contains(title, "社会")
	card.IsTechnical = card.IsAlgorithm || card.IsBackend || strings.Contains(title, "数据") || strings.Contains(title, "工程") || strings.Contains(title, "系统") || strings.Contains(title, "技术")

	// Core requirements from job text
	if card.IsBackend {
		card.CoreRequirements = append(card.CoreRequirements, "后端开发与系统设计能力")
		card.HardRequirements = append(card.HardRequirements, "Go/Python/Java 至少一门后端语言")
	}
	if card.IsAlgorithm {
		card.CoreRequirements = append(card.CoreRequirements, "算法模型设计与训练能力")
		card.HardRequirements = append(card.HardRequirements, "深度学习框架使用经验")
	}
	if card.IsProduct {
		card.CoreRequirements = append(card.CoreRequirements, "产品需求分析与设计能力")
	}
	if card.IsUserResearch {
		card.CoreRequirements = append(card.CoreRequirements, "用户研究方法与调研能力")
	}
	if card.IsBusinessSolution {
		card.CoreRequirements = append(card.CoreRequirements, "方案策划与客户沟通能力")
	}

	// Keyword requirements from job description
	if strings.Contains(desc, "数据库") || strings.Contains(desc, "mysql") {
		card.KeywordRequirements = append(card.KeywordRequirements, "数据库使用经验")
	}
	if strings.Contains(desc, "redis") || strings.Contains(desc, "缓存") {
		card.KeywordRequirements = append(card.KeywordRequirements, "缓存系统经验")
	}
	if strings.Contains(desc, "docker") || strings.Contains(desc, "部署") || strings.Contains(desc, "kubernetes") {
		card.KeywordRequirements = append(card.KeywordRequirements, "部署与运维经验")
	}
	if strings.Contains(desc, "pytorch") || strings.Contains(desc, "tensorflow") {
		card.KeywordRequirements = append(card.KeywordRequirements, "深度学习框架")
	}
	if strings.Contains(desc, "nlp") || strings.Contains(desc, "文本") {
		card.KeywordRequirements = append(card.KeywordRequirements, "NLP相关经验")
	}

	if len(card.CoreRequirements) == 0 {
		card.CoreRequirements = append(card.CoreRequirements, "与岗位方向相关的工作能力")
	}

	return card
}

// ─── Hard Guard ──────────────────────────────────────────────────────────────

func applyMatchHardGuard(resume *model.MatchResumeFeatureCard, job *model.MatchJobFeatureCard) *model.MatchHardGuardResult {
	guard := &model.MatchHardGuardResult{
		Result:   "pass",
		MaxScore: 95,
	}

	// Rule 1: non-technical resume + technical job
	if !resume.IsTechnical && job.IsTechnical {
		guard.MaxScore = minInt(guard.MaxScore, 65)
		guard.Result = "downgrade"
		guard.Reasons = append(guard.Reasons, "非技术简历与技术岗方向不匹配")
		guard.RiskPoints = append(guard.RiskPoints, "缺少技术项目/工程证据，投递竞争力受限")
	}

	// Rule 2: no algorithm evidence + algorithm job
	if !resume.IsAlgorithm && job.IsAlgorithm {
		guard.MaxScore = minInt(guard.MaxScore, 60)
		guard.Result = "downgrade"
		guard.Reasons = append(guard.Reasons, "简历缺少算法/模型/实验/论文证据")
		guard.RiskPoints = append(guard.RiskPoints, "算法岗位需要较强的建模和实验评估经历")
	}

	// Rule 3: no backend evidence + backend job
	if !resume.IsBackend && job.IsBackend {
		guard.MaxScore = minInt(guard.MaxScore, 65)
		guard.Result = "downgrade"
		guard.Reasons = append(guard.Reasons, "简历缺少后端工程/数据库/接口/系统设计证据")
		guard.RiskPoints = append(guard.RiskPoints, "后端岗位需要接口开发、数据库使用和工程部署经验")
	}

	// Rule 4: product resume + product/user research/project management job → pass
	if resume.IsProduct && (job.IsProduct || job.IsUserResearch || job.IsProjectManagement) {
		// Boost: allow high scores
	}

	// Rule 5: public service + business solution → weak_related
	if resume.IsPublicService && (job.IsBusinessSolution || job.IsProjectManagement) {
		guard.MaxScore = minInt(guard.MaxScore, 80)
		guard.Result = "weak_related"
		guard.Reasons = append(guard.Reasons, "公共服务背景与解决方案岗位可迁移，但非直接匹配")
		guard.RiskPoints = append(guard.RiskPoints, "建议突出调研、需求理解、材料输出、项目支持等迁移能力")
	}

	// Rule 6: exploration type resume → conservative
	if resume.IsExploration {
		guard.MaxScore = minInt(guard.MaxScore, 78)
		guard.Reasons = append(guard.Reasons, "简历方向尚需聚焦，证据不够集中")
		guard.RiskPoints = append(guard.RiskPoints, "建议先明确职业方向，积累2-3个深度项目经历后再重点投递")
	}

	// Rule 7: public service + algorithm/CV → hard mismatch
	if resume.IsPublicService && (job.IsAlgorithm || job.IsComputerVision) {
		guard.MaxScore = minInt(guard.MaxScore, 50)
		guard.Result = "mismatch"
		guard.Reasons = append(guard.Reasons, "公共服务背景与算法/CV技术岗严重不匹配")
		guard.RiskPoints = append(guard.RiskPoints, "建议关注调研分析、用户研究、项目支持、公共服务类岗位")
	}

	if guard.Result == "pass" {
		guard.Result = "pass"
	}

	return guard
}

// ─── Quality Deep Builders ───────────────────────────────────────────────────

func buildMatchEvidenceChain(cp matchProfile, resume *model.ResumeVersion, job *model.Job, maxN int) []model.MatchEvidence {
	var chain []model.MatchEvidence
	projText := strings.ToLower(projectsText(resume))
	internText := strings.ToLower(internshipsText(resume))
	skillsText := strings.ToLower(resume.ProfessionalSkills)
	eduText := strings.ToLower(educationsText(resume))
	jdText := strings.ToLower(job.Description + " " + job.Requirements)
	titleText := strings.ToLower(job.Title)

	// Evidence 1: Backend engineering
	if strings.Contains(jdText, "后端") || strings.Contains(jdText, "后台") || strings.Contains(titleText, "软件开发") {
		if strings.Contains(projText, "go") || strings.Contains(projText, "mysql") || strings.Contains(projText, "redis") {
			chain = append(chain, model.MatchEvidence{
				Requirement: "后端工程实践能力", ResumePoint: pickBackendEvidence(projText, internText),
				Judgement: "匹配", ResumeModule: "项目经历",
			})
		} else if cp.dominant == "backend_engineering" {
			chain = append(chain, model.MatchEvidence{
				Requirement: "后端工程实践能力", ResumePoint: "包含Go/数据库相关项目经验",
				Judgement: "部分匹配", ResumeModule: "项目经历",
			})
		}
	}

	// Evidence 2: Database experience
	if strings.Contains(jdText, "数据库") || strings.Contains(jdText, "mysql") || strings.Contains(jdText, "redis") || strings.Contains(projText, "mysql") {
		chain = append(chain, model.MatchEvidence{
			Requirement: "数据库与缓存使用经验", ResumePoint: pickDBEvidence(projText, skillsText),
			Judgement: "匹配", ResumeModule: "专业技能",
		})
	}

	// Evidence 3: Algorithm/AI
	if strings.Contains(jdText, "算法") || strings.Contains(jdText, "模型") || strings.Contains(jdText, "nlp") || strings.Contains(jdText, "机器学习") {
		if cp.dominant == "algorithm_research" {
			chain = append(chain, model.MatchEvidence{
				Requirement: "算法模型设计与训练能力", ResumePoint: pickAlgoEvidence(projText, skillsText),
				Judgement: "匹配", ResumeModule: "项目经历",
			})
		} else if !cp.tags.has("modeling_ability", 2) {
			chain = append(chain, model.MatchEvidence{
				Requirement: "算法模型设计与训练能力", ResumePoint: "简历中缺少算法/模型训练相关项目",
				Judgement: "缺口", ResumeModule: "整体评估",
			})
		}
	}

	// Evidence 4: Project/Internship presence
	projCount := len(resume.Projects)
	internCount := len(resume.Internships)
	if projCount > 0 || internCount > 0 {
		ev := fmt.Sprintf("共 %d 个项目经历、%d 段实习经历", projCount, internCount)
		chain = append(chain, model.MatchEvidence{
			Requirement: "项目实践经验", ResumePoint: ev,
			Judgement: "匹配", ResumeModule: "整体评估",
		})
	}

	// Evidence 5: Skill alignment with job keywords
	jobKeywords := extractJobKeywords(job)
	foundSkills := matchKeywords(skillsText+" "+projText, jobKeywords)
	if len(foundSkills) > 0 {
		chain = append(chain, model.MatchEvidence{
			Requirement: "技能关键词匹配", ResumePoint: fmt.Sprintf("简历中匹配到 %d 项岗位关键词（%s）", len(foundSkills), strings.Join(foundSkills[:min(3,len(foundSkills))], "、")),
			Judgement: "匹配", ResumeModule: "技能",
		})
	}

	// Evidence 6: Education level match
	hasHigherEdu := strings.Contains(eduText, "硕士") || strings.Contains(eduText, "博士") || strings.Contains(eduText, "研究生")
	if hasHigherEdu {
		chain = append(chain, model.MatchEvidence{
			Requirement: "学历背景", ResumePoint: "具备硕士及以上学历",
			Judgement: "匹配", ResumeModule: "教育经历",
		})
	} else {
		chain = append(chain, model.MatchEvidence{
			Requirement: "学历背景", ResumePoint: "本科学历",
			Judgement: "部分匹配", ResumeModule: "教育经历",
		})
	}

	if len(chain) == 0 {
		chain = append(chain, model.MatchEvidence{
			Requirement: "岗位核心能力", ResumePoint: "请查看简历详情了解具体项目与技能",
			Judgement: "待评估", ResumeModule: "整体评估",
		})
	}
	if len(chain) > maxN {
		chain = chain[:maxN]
	}
	// K9-2-Regression-F-B: populate Risk/FixAction per judgement
	for i := range chain {
		switch chain[i].Judgement {
		case "匹配":
			chain[i].Risk = "当前证据较充分，但仍建议补充量化结果或更具体的场景细节"
			chain[i].FixAction = "在对应项目经历中补充该能力的使用场景、结果指标或业务影响"
		case "部分匹配":
			chain[i].Risk = "已有相关经历，但与岗位要求的直接证据仍不够完整"
			chain[i].FixAction = "在项目经历或技能字段中补充该要求对应的工具使用、问题规模或量化结果"
		case "缺口":
			chain[i].Risk = "当前简历缺少该岗位要求的直接证据，可能影响初筛判断"
			chain[i].FixAction = "补充学习、项目或实践经历中与该要求相关的可验证证据"
		default:
			chain[i].Risk = "需要进一步确认简历中是否有对应证据"
			chain[i].FixAction = "在简历对应模块中补充与该要求直接相关的项目证据、结果指标或技能使用场景"
		}
	}
	return chain
}

func extractJobKeywords(job *model.Job) []string {
	t := strings.ToLower(job.Description + " " + job.Requirements + " " + job.Title)
	var kws []string
	for _, kw := range []string{"go", "python", "java", "mysql", "redis", "docker", "kubernetes", "linux", "pytorch", "tensorflow", "gin", "spring", "api", "接口", "部署"} {
		if strings.Contains(t, kw) { kws = append(kws, kw) }
	}
	return kws
}

func pickBackendEvidence(projText, internText string) string {
	if strings.Contains(projText, "goframe") || strings.Contains(projText, "gin") {
		return "包含Go后端框架（GoFrame/Gin）项目开发经验"
	}
	if strings.Contains(projText, "mysql") && strings.Contains(projText, "redis") {
		return "包含MySQL+Redis数据库项目实践"
	}
	if strings.Contains(internText, "后端") || strings.Contains(internText, "后台") {
		return "包含后端开发实习经历"
	}
	return "包含软件开发相关项目经验"
}

func pickDBEvidence(projText, skillsText string) string {
	if strings.Contains(projText+" "+skillsText, "mysql") && strings.Contains(projText+" "+skillsText, "redis") {
		return "MySQL + Redis 使用经验"
	}
	if strings.Contains(projText+" "+skillsText, "mysql") {
		return "MySQL 数据库使用经验"
	}
	return "数据库相关技能"
}

func pickAlgoEvidence(projText, skillsText string) string {
	if strings.Contains(projText+" "+skillsText, "pytorch") || strings.Contains(projText+" "+skillsText, "tensorflow") {
		return "深度学习框架（PyTorch/TensorFlow）使用经验"
	}
	if strings.Contains(projText, "nlp") || strings.Contains(projText, "自然语言") {
		return "NLP相关项目经历"
	}
	if strings.Contains(projText, "推荐") || strings.Contains(projText, "ctr") {
		return "推荐系统相关项目经历"
	}
	return "算法方向学习与项目经验"
}

func buildMatchGapPointsV2(cp matchProfile, resumeFC *model.MatchResumeFeatureCard, jobFC *model.MatchJobFeatureCard) []string {
	var gaps []string

	if jobFC.IsBackend && !resumeFC.IsBackend {
		gaps = append(gaps, "缺少后端开发项目经历，建议补充Go/Java+数据库+接口开发类项目")
	}
	if jobFC.IsAlgorithm && !resumeFC.IsAlgorithm {
		gaps = append(gaps, "缺少算法/模型训练相关项目，建议补充深度学习或NLP/CV方向项目")
	}
	if jobFC.IsComputerVision && (len(resumeFC.EvidenceSummary) == 0 || !strings.Contains(strings.Join(resumeFC.EvidenceSummary, " "), "视觉")) {
		gaps = append(gaps, "缺少计算机视觉相关项目或论文经历")
	}
	if (jobFC.IsProduct || jobFC.IsUserResearch) && !resumeFC.IsProduct {
		gaps = append(gaps, "缺少产品/用户研究相关经历，建议补充需求分析或用户调研项目")
	}
	if resumeFC.IsExploration {
		gaps = append(gaps, "简历方向尚不聚焦，建议先明确1-2个核心方向，积累深度项目经历")
	}
	if resumeFC.IsPublicService && jobFC.IsTechnical {
		gaps = append(gaps, "公共服务背景与目标技术岗位存在方向差异，部分软技能可迁移但需要补足硬技能")
	}

	// Always add a skill-gap check
	if len(resumeFC.WeakCapabilities) > 0 {
		caps := resumeFC.WeakCapabilities[:min(2, len(resumeFC.WeakCapabilities))]
		gaps = append(gaps, fmt.Sprintf("当前简历在%s方面缺少证据，建议针对性补充相关项目或学习经历", strings.Join(caps, "、")))
	}

	// Always add a keyword gap
	if len(jobFC.KeywordRequirements) > 0 {
		kws := jobFC.KeywordRequirements[:min(2, len(jobFC.KeywordRequirements))]
		gaps = append(gaps, fmt.Sprintf("简历中未明确体现%s相关经验，建议在技能栏或项目描述中补充", strings.Join(kws, "、")))
	}

	if len(gaps) < 2 {
		gaps = append(gaps, "建议针对目标岗位的核心要求，补充更具体的项目证据和量化成果")
	}
	return gaps
}

func buildMatchTopFixItems(cp matchProfile, resumeFC *model.MatchResumeFeatureCard, jobFC *model.MatchJobFeatureCard) []model.TopFixItem {
	var items []model.TopFixItem

	if jobFC.IsBackend && resumeFC.IsBackend {
		items = append(items, model.TopFixItem{
			Issue: "项目经历中缺少后端技术栈的具体实现细节", WhyItMatters: "后端岗位初筛关注接口开发、数据库设计和部署能力的证据",
			Action: "在项目经历中补充使用Go/MySQL/Redis完成接口开发、权限校验或部署联调的具体结果",
			ExampleReference: "例如：使用GoFrame实现RESTful接口，完成JWT鉴权和Redis缓存，支撑QPS从300提升至2000",
			SelfCheckQuestion: "你的项目描述中是否包含了具体的后端技术栈名称和量化结果？", EstimatedEffort: "15-30分钟", ResumeModule: "项目经历",
		})
	}
	if jobFC.IsBackend && !resumeFC.IsBackend {
		items = append(items, model.TopFixItem{
			Issue: "简历缺少后端开发相关项目经验", WhyItMatters: "后端岗位是技术栈驱动的，需要展示实际的工程实践能力",
			Action: "补充一个包含Go/Java+MySQL+Redis+API设计的项目，展示从需求到部署的完整流程",
			ExampleReference: "例如：开发一个包含用户认证、权限管理、数据CRUD和缓存策略的后台管理系统",
			SelfCheckQuestion: "是否有至少一个项目展示了完整的后端工程实践？", EstimatedEffort: "2-4周", ResumeModule: "项目经历",
		})
	}
	if jobFC.IsAlgorithm {
		items = append(items, model.TopFixItem{
			Issue: "需要展示算法模型训练和评估的具体证据", WhyItMatters: "算法岗位看重模型设计、训练过程和评估指标的专业性",
			Action: "补充具体的模型结构描述、训练数据规模、评估指标(如NDCG/AUC)和实验对比结果",
			ExampleReference: "例如：基于Transformer的点击率预估模型，在百万级数据集上训练，NDCG@10提升12%",
			SelfCheckQuestion: "你的项目描述中是否包含模型结构、数据集规模和评估指标？", EstimatedEffort: "20-40分钟", ResumeModule: "项目经历",
		})
	}
	if (jobFC.IsProduct || jobFC.IsUserResearch) && !resumeFC.IsProduct {
		items = append(items, model.TopFixItem{
			Issue: "缺少产品/用户研究相关经历", WhyItMatters: "产品和用户研究岗位需要展示需求分析、用户洞察和项目推进能力",
			Action: "补充产品分析报告、用户调研报告或项目策划文档作为作品集",
			ExampleReference: "例如：对某产品进行用户调研，撰写需求分析报告并提出3条产品优化建议",
			SelfCheckQuestion: "是否有产品分析或用户调研的经历可以写入简历？", EstimatedEffort: "1-2周", ResumeModule: "附加信息",
		})
	}

	// Always add fix items to reach minimum 3
	fixItem2 := model.TopFixItem{
		Issue: "简历缺少量化成果和具体技术细节", WhyItMatters: "HR和面试官通过具体数字和细节判断候选人的真实水平",
		Action: "在每个项目经历中补充技术栈、量化成果（如QPS/数据量/提升百分比）和你的具体贡献",
		ExampleReference: "用STAR法则（情境-任务-行动-结果）描述每个项目",
		SelfCheckQuestion: "每个项目经历是否包含技术栈名称+量化结果+你的具体职责？", EstimatedEffort: "30-60分钟", ResumeModule: "项目经历",
	}
	fixItem3 := model.TopFixItem{
		Issue: "技能栏缺少与目标岗位匹配的关键词", WhyItMatters: "JD关键词出现在简历中会显著提高初筛通过率",
		Action: fmt.Sprintf("在技能栏和项目描述中融入目标岗位的关键词（如%s）", strings.Join(jobFC.KeywordRequirements[:minInt(3,len(jobFC.KeywordRequirements))], "、")),
		ExampleReference: "在专业技能区域按优先级排列与岗位相关的技能", SelfCheckQuestion: "简历的技能描述中是否包含了JD中列出的技术要求？",
		EstimatedEffort: "10-20分钟", ResumeModule: "技能",
	}

	if len(items) < 3 {
		items = append(items, fixItem2)
	}
	if len(items) < 3 {
		items = append(items, fixItem3)
	}

	if len(items) > 5 {
		items = items[:5]
	}
	return items
}

func buildApplyReadiness(score int, guard *model.MatchHardGuardResult, resumeFC *model.MatchResumeFeatureCard) string {
	if guard.Result == "mismatch" {
		return "not_recommended: 方向严重不匹配，建议选择更对口的岗位"
	}
	if resumeFC.IsExploration {
		return "prepare_first: 建议先明确职业方向并积累深度项目经验后再重点投递"
	}
	if score >= 85 {
		return "ready: 整体匹配度较高，可直接投递，建议投递前微调简历突出关键优势"
	}
	if score >= 70 {
		return "prepare_first: 建议先修复Top 3简历问题，并补充与岗位更匹配的项目证据后再投递"
	}
	return "cautious_try: 可尝试投递，但建议优先补强关键短板以提高通过率"
}

func buildOptimizationStopSignal(score int, gapPoints []string) string {
	if score >= 85 && len(gapPoints) <= 1 {
		return "当前可先投递，建议在项目经历中补充量化结果和岗位关键词"
	}
	if score >= 70 {
		return "建议先补充简历中的关键缺口证据再投递，当前状态下通过率可能偏低"
	}
	return "建议先针对性补强核心技能和项目经历，不建议在修复前大量投递"
}

func buildMatchNextSteps(score int, resumeFC *model.MatchResumeFeatureCard, jobFC *model.MatchJobFeatureCard, gapPoints []string) []string {
	steps := []string{
		"仔细阅读岗位JD，确认岗位核心要求和加分项",
	}
	if len(gapPoints) > 0 {
		steps = append(steps, "针对检测到的缺口补强简历中对应的项目证据和技能描述")
	}
	if score < 80 {
		steps = append(steps, "进入简历优化页，生成面向该岗位的详细优化建议")
	}
	steps = append(steps, "在投递前确认你的简历突出了与岗位最匹配的经验和能力")
	if score >= 80 {
		steps = append(steps, "准备好面试中重点介绍与岗位匹配的项目经历和技术方案")
	}
	return steps
}

func buildDeliveryStrategyV2(score int, job *model.Job, resumeFC *model.MatchResumeFeatureCard, guard *model.MatchHardGuardResult) string {
	var sb strings.Builder
	sb.WriteString("【投递建议】")

	if score >= 85 {
		sb.WriteString(fmt.Sprintf("与当前%s岗位的匹配度较高，建议在项目经历中补充与岗位要求直接相关的量化结果和技能使用场景后投递。", job.Category))
	} else if score >= 70 {
		sb.WriteString(fmt.Sprintf("可尝试投递当前岗位，但建议先围绕岗位核心要求补强相关项目证据。投递时强调可迁移能力和学习潜力。"))
	} else {
		if guard.Result == "mismatch" {
			sb.WriteString("不建议当前投递此岗位。方向差异较大，建议关注与自身背景更匹配的方向。")
		} else {
			sb.WriteString(fmt.Sprintf("建议先围绕%s岗位的核心要求补强相关技能和项目经历后再投递。", job.Category))
		}
	}

	if len(guard.Reasons) > 0 {
		sb.WriteString(fmt.Sprintf("【风险提示】%s。", strings.Join(guard.Reasons, "；")))
	}

	if resumeFC.IsExploration {
		sb.WriteString("【特别建议】你的简历方向尚在探索阶段，建议先明确方向并积累对应项目经验后再重点投递。")
	}
	return sb.String()
}

// ─── Profile / Score / Helpers (unchanged from existing) ─────────────────────

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
			return cp
		}
	}
	cp2 := extractCandidateTags(resume)
	cp.dominant = cp2.dominantProfile
	cp.tags = cp2.tags
	return cp
}

func (a *MatchMockAgent) calcScore(cp matchProfile, resume *model.ResumeVersion, job *model.Job) int {
	jp := extractJobTags(job)
	dirF := directionFit(cp.dominant, jp.tags)

	jobText := strings.ToLower(job.Title + " " + job.Category + " " + job.Department)
	if cp.dominant == "backend_engineering" {
		if strings.Contains(jobText, "后台") || strings.Contains(jobText, "后端") || strings.Contains(jobText, "服务端") || strings.Contains(jobText, "全栈") {
			dirF = "strong"
		}
	}
	if cp.dominant == "algorithm_research" {
		if strings.Contains(jobText, "算法") || strings.Contains(jobText, "nlp") || strings.Contains(jobText, "数据科学") || strings.Contains(jobText, "多模态") || strings.Contains(jobText, "机器学习") || strings.Contains(jobText, "ai") {
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

	if cp.dominant == "backend_engineering" && dirF == "strong" {
		score += 10
	}
	if cp.dominant == "algorithm_research" && dirF == "strong" {
		score += 5
	}

	isAlgo := jp.tags.has("algorithm_core", 2) || jp.tags.has("nlp", 2) || jp.tags.has("multimodal", 2)
	isBackend := jp.tags.has("backend_service", 2)
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
			score = 89
		}
	case "product_operation":
		if isAlgo || isBackend {
			if score > 69 { score = 69 }
		}
	case "business_solution":
		if isAlgo || isBackend {
			if score > 69 { score = 69 }
		}
	}

	if score > 95 { score = 95 }
	if score < 35 { score = 35 }
	return score
}

func matchLevel(score int) string {
	switch {
	case score >= 90: return "强匹配，优先投递"
	case score >= 80: return "高匹配，建议重点投递"
	case score >= 70: return "可尝试，需要补强"
	case score >= 60: return "备选参考，不建议优先"
	default: return "不建议投递"
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

// ─── Existing Helper Functions (unchanged) ────────────────────────────────────

func buildFitPointsV2(cp matchProfile, resume *model.ResumeVersion, job *model.Job, maxN int) []string {
	var pts []string
	t := strings.ToLower(job.Title + " " + job.Category)
	projText := strings.ToLower(projectsText(resume))
	internText := strings.ToLower(internshipsText(resume))

	if cp.tags.has("modeling_ability", 2) && cp.dominant == "algorithm_research" {
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
			if strings.Contains(projText, kw) { found[kw] = true }
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
	skillsText := strings.ToLower(resume.ProfessionalSkills)

	for _, p := range resume.Projects {
		if len(ev) >= 5 { break }
		if p.Name != "" {
			ev = append(ev, p.Name+"项目经验")
		}
	}
	for _, i := range resume.Internships {
		if len(ev) >= 5 { break }
		if i.Position != "" {
			ev = append(ev, i.Position+"实习经历")
		}
	}
	for _, e := range resume.Educations {
		if len(ev) >= 5 { break }
		if e.Degree != "" {
			ev = append(ev, fmt.Sprintf("%s学历（%s %s）", e.Degree, e.School, e.Major))
		}
	}
	allSkills := []string{"go", "python", "java", "mysql", "redis", "docker", "linux", "pytorch", "tensorflow"}
	for _, kw := range allSkills {
		if len(ev) >= 5 { break }
		if strings.Contains(projText+" "+skillsText, kw) && !containsName(ev, kw) {
			ev = append(ev, strings.ToUpper(kw)+" 使用经验")
		}
	}
	if len(ev) < 2 {
		ev = append(ev, "简历包含相关学习与项目经历")
	}
	// Ensure minimum 2: add skills summary from ProfessionalSkills
	if len(ev) < 2 && resume.ProfessionalSkills != "" {
		skills := resume.ProfessionalSkills
		if len(skills) > 40 { skills = skills[:40] }
		ev = append(ev, fmt.Sprintf("专业技能：%s等", skills))
	}
	if len(ev) < 2 {
		ev = append(ev, "可通过简历优化补充更多项目证据")
	}
	for _, c := range resume.Certificates {
		if len(ev) >= 5 { break }
		if c.Name != "" {
			ev = append(ev, c.Name+"证书")
		}
	}
	return ev
}

func containsName(ev []string, kw string) bool {
	kl := strings.ToLower(kw)
	for _, e := range ev {
		if strings.Contains(strings.ToLower(e), kl) { return true }
	}
	return false
}

func buildJobRequirements(job *model.Job) []string {
	var reqs []string
	t := strings.ToLower(job.Title + " " + job.Category + " " + job.Requirements + " " + job.Description)
	if strings.Contains(t, "后端") || strings.Contains(t, "后台") { reqs = append(reqs, "后端接口开发与系统设计能力") }
	if strings.Contains(t, "数据库") || strings.Contains(t, "mysql") || strings.Contains(t, "redis") { reqs = append(reqs, "数据库与缓存使用经验") }
	if strings.Contains(t, "算法") || strings.Contains(t, "模型") { reqs = append(reqs, "算法模型训练与实验评估能力") }
	if strings.Contains(t, "产品") || strings.Contains(t, "运营") { reqs = append(reqs, "需求分析与用户研究能力") }
	if strings.Contains(t, "docker") || strings.Contains(t, "部署") || strings.Contains(t, "kubernetes") { reqs = append(reqs, "容器部署与运维经验") }
	if strings.Contains(t, "go") || strings.Contains(t, "python") || strings.Contains(t, "java") || strings.Contains(t, "c++") { reqs = append(reqs, "至少掌握一门开发语言") }
	if strings.Contains(t, "linux") { reqs = append(reqs, "Linux系统操作能力") }
	if strings.Contains(t, "团队") || strings.Contains(t, "协作") || strings.Contains(t, "沟通") { reqs = append(reqs, "团队协作与沟通能力") }
	if strings.Contains(t, "项目") || strings.Contains(t, "管理") { reqs = append(reqs, "项目管理与推进能力") }
	if len(reqs) < 3 { reqs = append(reqs, "与岗位方向相关的学习与工作能力") }
	if len(reqs) < 3 { reqs = append(reqs, "快速学习和适应新技术的能力") }
	return reqs
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
		case domainAlgorithmNLP: imps = append(imps, "建议补充NLP方向项目（文本分类/情感分析/问答系统），展示Transformer使用经验")
		case domainAlgorithmMultimodal: imps = append(imps, "建议补充图文检索或视频理解项目，说明数据集、模型结构和评估指标")
		case domainAIApplication: imps = append(imps, "建议补充AI服务化项目：模型接口封装、Docker部署、性能优化")
		default: imps = append(imps, "建议继续深化算法方向项目，补充模型实验与评估证据")
		}
	case "backend_engineering": imps = append(imps, "建议在项目经历中补充接口QPS、缓存策略、权限模型和部署架构")
	case "product_operation": imps = append(imps, "建议补充产品数据分析、用户调研或A/B测试相关经验")
	case "business_solution": imps = append(imps, "建议补充行业分析、客户需求沟通或方案策划相关经验")
	default: imps = append(imps, "建议针对目标岗位补强相关技能和项目经历")
	}
	return imps
}

func buildTalkingPoints(cp matchProfile, resume *model.ResumeVersion, job *model.Job) []string {
	var pts []string
	for _, p := range resume.Projects {
		if len(pts) >= 3 { break }
		if p.Name != "" {
			pts = append(pts, fmt.Sprintf("介绍你在%s项目中承担的角色、技术选型和核心成果", p.Name))
		}
	}
	for _, i := range resume.Internships {
		if len(pts) >= 3 { break }
		if i.Company != "" {
			pts = append(pts, fmt.Sprintf("分享你在%s实习期间解决的关键问题和业务成果", i.Company))
		}
	}
	if len(pts) < 3 {
		pts = append(pts, fmt.Sprintf("准备1-2个深度技术案例，说明从问题分析到方案落地的完整过程，突出对%s的实践理解", job.Direction))
	}
	if len(pts) < 3 {
		pts = append(pts, "提前梳理目标岗位的JD关键词，在面试中主动展示匹配的实践经验")
	}
	if len(pts) < 3 {
		pts = append(pts, "说明你对目标行业或业务场景的理解，展示业务敏感度和学习能力")
	}
	return pts
}

func toMatchSuggestions(strs []string) []model.MatchSuggestion {
	var s []model.MatchSuggestion
	for i, str := range strs {
		prio := "P1"
		if i == 0 { prio = "P0" }
		s = append(s, model.MatchSuggestion{Priority: prio, Title: str, Action: str})
	}
	return s
}
