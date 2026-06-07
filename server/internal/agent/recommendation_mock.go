package agent

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"offer-catcher/server/internal/model"
)

type RecommendationMockAgent struct{}

func NewRecommendationMockAgent() *RecommendationMockAgent { return &RecommendationMockAgent{} }

// ══════════════════════════════════════════════════════════════════════════
// Tag System
// ══════════════════════════════════════════════════════════════════════════

type tagSet map[string]int // tag → confidence (0-3)

func (ts tagSet) get(t string) int { return ts[t] }
func (ts tagSet) set(t string, v int) {
	if v > ts[t] {
		ts[t] = v
	}
}
func (ts tagSet) has(t string, min int) bool { return ts[t] >= min }

// ══════════════════════════════════════════════════════════════════════════
// Candidate Tags
// ══════════════════════════════════════════════════════════════════════════

type candidateTagProfile struct {
	tags           tagSet
	dominantProfile string
	primaryTags    []string
	allText        string
	eduText        string
	projText       string
	internText     string
	skillsText     string
	schoolLevel    int
	degreeLevel    int
}

func extractCandidateTags(resume *model.ResumeVersion) candidateTagProfile {
	cp := candidateTagProfile{tags: make(tagSet)}
	cp.skillsText = strings.ToLower(resume.ProfessionalSkills + " " + resume.AIApplication + " " + resume.AdditionalInfo)

	// ── Education signals (high weight) ──
	for _, edu := range resume.Educations {
		txt := strings.ToLower(edu.School + " " + edu.Major + " " + edu.Department + " " +
			edu.ResearchDirection + " " + edu.CoreCourses + " " + edu.PaperSummary + " " +
			edu.Advisor + " " + edu.Lab)
		cp.eduText += txt + " "
		cp.tags.set("education_evidence", 3)
		detectSchoolLevel(&cp, edu)
		detectResearchDirection(&cp, edu)
		detectDegreeLevel(&cp, edu)
	}

	// ── Project signals (high weight) ──
	for _, proj := range resume.Projects {
		txt := strings.ToLower(proj.Name + " " + proj.Role + " " + proj.Description)
		cp.projText += txt + " "
		cp.tags.set("project_evidence", 3)
	}
	detectModelingSignals(&cp, cp.projText)
	detectExperimentSignals(&cp, cp.projText)

	// ── Internship signals (high weight) ──
	for _, intern := range resume.Internships {
		txt := strings.ToLower(intern.Company + " " + intern.Position + " " + intern.Description)
		cp.internText += txt + " "
		cp.tags.set("internship_evidence", 3)
	}
	detectModelingSignals(&cp, cp.internText)
	detectExperimentSignals(&cp, cp.internText)

	// ── Paper/research signals ──
	for _, edu := range resume.Educations {
		if edu.PaperSummary != "" {
			cp.tags.set("research_evidence", 3)
			cp.tags.set("research_ability", maxOf(cp.tags.get("research_ability"), 2))
		}
	}

	// ── Skill signals (low weight - only skills list) ──
	detectCapabilityFromSkills(&cp, cp.skillsText)
	detectDomainFromSkills(&cp, cp.skillsText)

	// ── Full text ──
	cp.allText = strings.ToLower(cp.eduText + " " + cp.projText + " " + cp.internText + " " +
		cp.skillsText + " " + strings.ToLower(resume.RawText))

	// Direction tags from all text
	detectAlgorithmDirection(&cp)
	detectBackendDirection(&cp)

	// ── Dominant profile ──
	cp.dominantProfile = determineDominantProfile(cp.tags)
	cp.primaryTags = topTags(cp.tags, 5)

	return cp
}

func detectSchoolLevel(cp *candidateTagProfile, edu model.ResumeEducation) {
	s := strings.ToLower(edu.School)
	if matchAny(s, "清华", "北大", "浙江大学", "上海交通", "哈工大", "中国科大") {
		cp.schoolLevel = 2
		cp.tags.set("education_evidence", 3)
	} else if strings.Contains(s, "大学") && cp.schoolLevel < 1 {
		cp.schoolLevel = 1
	}
}

func detectDegreeLevel(cp *candidateTagProfile, edu model.ResumeEducation) {
	d := strings.ToLower(edu.Degree)
	if strings.Contains(d, "博士") {
		cp.degreeLevel = 3
	} else if strings.Contains(d, "硕士") && cp.degreeLevel < 2 {
		cp.degreeLevel = 2
	} else if strings.Contains(d, "本科") && cp.degreeLevel < 1 {
		cp.degreeLevel = 1
	}
}

func detectResearchDirection(cp *candidateTagProfile, edu model.ResumeEducation) {
	rd := strings.ToLower(edu.ResearchDirection)
	if rd == "" {
		return
	}
	// High-confidence research signals
	if matchAny(rd, "推荐系统", "搜索推荐", "推荐", "information retrieval") {
		cp.tags.set("recommendation_system", 3)
		cp.tags.set("research_ability", 3)
		cp.tags.set("research_evidence", 3)
	}
	if matchAny(rd, "用户行为建模", "用户行为") {
		cp.tags.set("modeling_ability", clampMin(cp.tags.get("modeling_ability")+1, 2))
	}
	if matchAny(rd, "表示学习", "embedding", "向量检索") {
		cp.tags.set("modeling_ability", clampMin(cp.tags.get("modeling_ability")+1, 2))
	}
	if matchAny(rd, "nlp", "自然语言", "语言处理") {
		cp.tags.set("nlp", 3)
		cp.tags.set("research_ability", 2)
	}
	if matchAny(rd, "多模态", "multimodal", "计算机视觉", "cv", "视觉") {
		cp.tags.set("multimodal", clampMin(cp.tags.get("multimodal")+1, 2))
		cp.tags.set("research_ability", clampMin(cp.tags.get("research_ability")+1, 2))
	}
	if matchAny(rd, "数据挖掘", "data mining", "大数据", "spark", "hive", "数据分析") {
		cp.tags.set("data", 2)
		cp.tags.set("data_analysis_ability", 2)
	}
}

func detectModelingSignals(cp *candidateTagProfile, text string) {
	signals := []string{"pytorch", "tensorflow", "深度学习", "deep learning", "ctr", "排序模型",
		"embedding", "attention", "transformer", "模型训练", "模型优化", "neural network"}
	for _, s := range signals {
		if strings.Contains(text, s) {
			cp.tags.set("modeling_ability", clampMin(cp.tags.get("modeling_ability")+1, 2))
			cp.tags.set("algorithm", clampMin(cp.tags.get("algorithm")+1, 2))
		}
	}
}

func detectExperimentSignals(cp *candidateTagProfile, text string) {
	signals := []string{"auc", "ndcg", "recall", "precision", "消融实验", "实验复现",
		"实验评估", "指标对比", "离线评估", "ablation", "benchmark", "实验平台"}
	for _, s := range signals {
		if strings.Contains(text, s) {
			cp.tags.set("experiment_ability", clampMin(cp.tags.get("experiment_ability")+1, 2))
			cp.tags.set("research_ability", clampMin(cp.tags.get("research_ability")+1, 2))
		}
	}
}

func detectAlgorithmDirection(cp *candidateTagProfile) {
	kws := []string{"算法", "机器学习", "深度学习", "推荐系统", "搜索推荐", "ctr", "排序模型",
		"nlp", "自然语言处理", "多模态", "计算机视觉", "cv", "数据挖掘",
		"表示学习", "用户行为建模", "pytorch", "tensorflow", "faiss",
		"spark", "hive", "auc", "ndcg", "recall", "embedding", "attention",
		"特征工程", "数据科学", "大数据", "向量检索", "模型训练", "模型部署"}
	for _, kw := range kws {
		if strings.Contains(cp.allText, kw) {
			cp.tags.set("algorithm", clampMin(cp.tags.get("algorithm")+1, 2))
		}
	}
	if cp.tags.has("research_ability", 2) {
		cp.tags.set("algorithm", clampMin(cp.tags.get("algorithm")+1, 2))
	}
	if cp.tags.has("modeling_ability", 2) {
		cp.tags.set("algorithm", clampMin(cp.tags.get("algorithm")+1, 2))
	}
}

func detectBackendDirection(cp *candidateTagProfile) {
	kws := []string{"go", "java", "c++", "微服务", "mysql", "redis", "kafka",
		"docker", "nginx", "接口", "权限", "分布式", "后端", "后台", "api", "服务端"}
	for _, kw := range kws {
		if strings.Contains(cp.allText, kw) {
			cp.tags.set("backend", clampMin(cp.tags.get("backend")+1, 1))
			cp.tags.set("engineering_ability", clampMin(cp.tags.get("engineering_ability")+1, 1))
		}
	}
	// Boost if backend skills appear in projects
	if strings.Contains(cp.projText, "go") || strings.Contains(cp.projText, "mysql") || strings.Contains(cp.projText, "redis") {
		cp.tags.set("backend_service", clampMin(cp.tags.get("backend_service")+1, 2))
	}
}

func detectCapabilityFromSkills(cp *candidateTagProfile, text string) {
	engKws := []string{"go", "java", "c++", "python", "mysql", "redis", "docker", "nginx", "kafka", "linux", "git"}
	for _, kw := range engKws {
		if strings.Contains(text, kw) {
			cp.tags.set("engineering_ability", clampMin(cp.tags.get("engineering_ability")+1, 1))
		}
	}
	dataKws := []string{"sql", "spark", "hive", "pandas", "numpy", "数据分析", "数据清洗"}
	for _, kw := range dataKws {
		if strings.Contains(text, kw) {
			cp.tags.set("data_analysis_ability", clampMin(cp.tags.get("data_analysis_ability")+1, 2))
		}
	}
}

func detectDomainFromSkills(cp *candidateTagProfile, text string) {
	domains := map[string]string{
		"pytorch": "machine_learning_platform", "tensorflow": "machine_learning_platform",
		"faiss": "machine_learning_platform", "spark": "high_performance_computing",
		"hive": "data_mining", "推荐系统": "recommendation_system",
	}
	for kw, tag := range domains {
		if strings.Contains(text, kw) {
			cp.tags.set(tag, 2)
		}
	}
}

// ══════════════════════════════════════════════════════════════════════════
// Dominant Profile
// ══════════════════════════════════════════════════════════════════════════

func determineDominantProfile(tags tagSet) string {
	algoRes := tags.has("algorithm", 2) &&
		(tags.has("research_ability", 2) || tags.has("experiment_ability", 2) || tags.has("modeling_ability", 2)) &&
		(tags.has("recommendation_system", 2) || tags.has("nlp", 2) || tags.has("multimodal", 2) ||
			tags.has("data_mining", 2) || tags.has("machine_learning_platform", 2))
	if algoRes {
		return "algorithm_research"
	}

	beEng := tags.has("backend", 2) && tags.has("engineering_ability", 2) && tags.has("backend_service", 2)
	if beEng {
		return "backend_engineering"
	}

	return "career_exploration"
}

// ══════════════════════════════════════════════════════════════════════════
// Job Tags
// ══════════════════════════════════════════════════════════════════════════

type jobTagProfile struct {
	job  *model.Job
	tags tagSet
}

func extractJobTags(job *model.Job) jobTagProfile {
	jp := jobTagProfile{job: job, tags: make(tagSet)}
	t := strings.ToLower(job.Title + " " + job.Category + " " + job.Department + " " +
		job.Description + " " + job.Requirements)

	// Direction detection (order matters - algorithm before backend)
	if matchAny(t, "算法", "数据科学", "nlp", "自然语言", "多模态", "计算机视觉",
		"推荐", "搜索", "数据挖掘", "ai", "模型", "深度学习", "机器学习") && !matchAny(t, "解决方案") {
		jp.tags.set("algorithm_core", 3)
		if matchAny(t, "nlp", "自然语言") {
			jp.tags.set("nlp", 3)
			jp.tags.set("requires_modeling_ability", 3)
			jp.tags.set("requires_experiment_ability", 2)
			jp.tags.set("requires_research_ability", 2)
		}
		if matchAny(t, "多模态", "视觉", "cv", "多媒体") && !matchAny(t, "多媒体处理方向", "多媒体") {
			jp.tags.set("multimodal", 3)
			jp.tags.set("requires_modeling_ability", 3)
			jp.tags.set("requires_experiment_ability", 2)
		}
	} else if matchAny(t, "技术研究", "高性能计算", "hpc") {
		jp.tags.set("technical_research", 3)
		jp.tags.set("high_performance_computing", 3)
		jp.tags.set("requires_research_ability", 3)
		jp.tags.set("requires_engineering_ability", 2)
	} else if matchAny(t, "后台", "后端", "服务端", "软件开发", "pc客户端", "客户端") {
		jp.tags.set("backend_service", 3)
		jp.tags.set("requires_engineering_ability", 3)
		jp.tags.set("requires_modeling_ability", 0)
		jp.tags.set("requires_research_ability", 0)
		if matchAny(t, "云平台", "云服务", "云开发", "云", "paas", "容器") {
			jp.tags.set("cloud_infrastructure", 2)
		}
	} else if matchAny(t, "产品经理", "产品策划", "产品运营", "用户研究", "运营", "增长", "项目管理") {
		jp.tags.set("product_operation", 3)
		jp.tags.set("requires_product_sense", 3)
		jp.tags.set("requires_business_understanding", 2)
		jp.tags.set("requires_communication", 2)
		jp.tags.set("requires_modeling_ability", 0)
		jp.tags.set("requires_research_ability", 0)
	} else if matchAny(t, "解决方案", "行业咨询", "采购", "商务") {
		jp.tags.set("solution_consulting", 3)
		jp.tags.set("requires_business_understanding", 3)
		jp.tags.set("requires_communication", 3)
		jp.tags.set("requires_engineering_ability", 1)
	}

	return jp
}

// ══════════════════════════════════════════════════════════════════════════
// Job Domain Detection (K9-1-T-C)
// ══════════════════════════════════════════════════════════════════════════

type jobDomain string

const (
	domainAlgorithmMultimodal  jobDomain = "algorithm_multimodal"
	domainAlgorithmNLP         jobDomain = "algorithm_nlp"
	domainAlgorithmDataScience jobDomain = "algorithm_data_science"
	domainAlgorithmCV          jobDomain = "algorithm_cv"
	domainAlgorithmGraphics    jobDomain = "algorithm_graphics"
	domainAlgorithmGeneral     jobDomain = "algorithm_general"
	domainAIApplication        jobDomain = "ai_application"
	domainBackend              jobDomain = "backend"
	domainClient               jobDomain = "client"
	domainFullstack            jobDomain = "fullstack"
	domainDataEngineering      jobDomain = "data_engineering"
	domainProduct              jobDomain = "product"
	domainOperation            jobDomain = "operation"
	domainBusinessSolution     jobDomain = "business_solution"
	domainBusinessFunction     jobDomain = "business_function"
	domainUnknown              jobDomain = "unknown"
)

func detectJobDomain(jp jobTagProfile) jobDomain {
	t := strings.ToLower(jp.job.Title + " " + jp.job.Category + " " + jp.job.Department + " " + jp.job.Description)

	// Algorithm sub-domains
	if jp.tags.has("algorithm_core", 2) {
		if matchAny(t, "多模态", "multimodal") && !matchAny(t, "多媒体处理") {
			return domainAlgorithmMultimodal
		}
		if matchAny(t, "图形", "cg", "动画", "三维", "3d", "图形学") {
			return domainAlgorithmGraphics
		}
		if matchAny(t, "nlp", "自然语言", "文本", "语义", "对话", "翻译", "语音") {
			return domainAlgorithmNLP
		}
		if matchAny(t, "计算机视觉", "cv", "图像", "视觉", "目标检测", "分割") {
			if jp.tags.has("multimodal", 2) {
				return domainAlgorithmMultimodal
			}
			return domainAlgorithmCV
		}
		if matchAny(t, "数据科学", "数据挖掘", "数据分析", "数据工程") {
			return domainAlgorithmDataScience
		}
		if matchAny(t, "ai应用", "ai平台", "智能应用", "算法应用", "ai工程") {
			return domainAIApplication
		}
		return domainAlgorithmGeneral
	}

	// Backend sub-domains
	if jp.tags.has("backend_service", 2) {
		if matchAny(t, "客户端", "pc客户端", "前端") {
			return domainClient
		}
		if matchAny(t, "全栈", "fullstack") {
			return domainFullstack
		}
		if matchAny(t, "数据工程", "数据平台", "etl", "数据仓库") {
			return domainDataEngineering
		}
		return domainBackend
	}

	// Product sub-domains
	if jp.tags.has("product_operation", 2) {
		if matchAny(t, "运营", "增长", "活动", "用户运营") {
			return domainOperation
		}
		return domainProduct
	}

	// Solution sub-domains
	if jp.tags.has("solution_consulting", 2) {
		if matchAny(t, "秘书") {
			return domainBusinessFunction
		}
		return domainBusinessSolution
	}

	return domainUnknown
}

func domainLabel(d jobDomain) string {
	switch d {
	case domainAlgorithmMultimodal:
		return "多模态方向"
	case domainAlgorithmNLP:
		return "NLP方向"
	case domainAlgorithmDataScience:
		return "数据科学方向"
	case domainAlgorithmCV:
		return "计算机视觉方向"
	case domainAlgorithmGraphics:
		return "图形学方向"
	case domainAlgorithmGeneral:
		return "算法方向"
	case domainAIApplication:
		return "AI应用开发"
	case domainBackend:
		return "后台开发"
	case domainClient:
		return "客户端开发"
	case domainFullstack:
		return "全栈开发"
	case domainDataEngineering:
		return "数据工程"
	case domainProduct:
		return "产品策划"
	case domainOperation:
		return "运营方向"
	case domainBusinessSolution:
		return "解决方案"
	case domainBusinessFunction:
		return "职能方向"
	default:
		return "综合方向"
	}
}

// precisionBonus computes the direction precision score for a given job.
func precisionBonus(cp *candidateTagProfile, sj scoredJobV2) int {
	d := detectJobDomain(sj.jp)

	// Exact domain tag matches yield highest precision
	switch d {
	case domainAlgorithmNLP:
		if cp.tags.has("nlp", 2) {
			return 8
		}
	case domainAlgorithmMultimodal, domainAlgorithmCV:
		if cp.tags.has("multimodal", 2) {
			return 8
		}
	case domainAlgorithmGeneral:
		if cp.tags.has("recommendation_system", 2) {
			return 7
		}
	}

	// Cross-domain bonuses
	switch cp.dominantProfile {
	case "algorithm_research":
		switch d {
		case domainAlgorithmMultimodal:
			if cp.tags.has("recommendation_system", 2) {
				return 7 // recommendation research is closest proxy for multimodal
			}
			return 5
		case domainAlgorithmNLP:
			return 8
		case domainAlgorithmDataScience:
			if cp.tags.has("data_analysis_ability", 2) {
				return 6
			}
			return 3
		case domainAlgorithmCV:
			return 5
		case domainAlgorithmGeneral:
			return 4
		case domainAIApplication:
			if cp.tags.has("engineering_ability", 2) {
				return 5
			}
			return 3
		case domainAlgorithmGraphics:
			return 1
		}
	case "backend_engineering":
		switch d {
		case domainBackend:
			return 6
		case domainFullstack:
			return 4
		case domainDataEngineering:
			return 3
		case domainAIApplication:
			return 2
		case domainClient:
			return 2
		}
	case "product_operation":
		switch d {
		case domainProduct:
			return 6
		case domainOperation:
			return 5
		}
	case "business_solution":
		switch d {
		case domainBusinessSolution:
			return 6
		case domainBusinessFunction:
			return 4
		}
	}

	return 0
}

// ══════════════════════════════════════════════════════════════════════════
// Tag Matrix Scoring
// ══════════════════════════════════════════════════════════════════════════

func tagMatrixScore(cTags tagSet, jTags tagSet) int {
	score := 0
		pairs := []struct {
		cTag, jTag string
		weight     int
	}{
		// K9-1-M: algorithm-domain pairs
		{"recommendation_system", "recommendation_system", 18},
		{"search_algorithm", "search_algorithm", 18},
		{"nlp", "nlp", 18},
		{"multimodal", "multimodal", 16},
		{"modeling_ability", "requires_modeling_ability", 14},
		{"experiment_ability", "requires_experiment_ability", 12},
		{"research_ability", "technical_research", 14},
		{"research_ability", "requires_research_ability", 12},
		{"data_analysis_ability", "data_science", 12},
		// K9-1-M: backend-domain pairs (was too weak)
		{"backend", "backend_service", 14},
		{"engineering_ability", "backend_service", 12},
		{"engineering_ability", "requires_engineering_ability", 14},
		{"backend", "cloud_infrastructure", 10},
		{"backend_service", "cloud_infrastructure", 10},
		// K9-1-M: product/business pairs (new)
		{"product", "product_operation", 14},
		{"product_sense", "requires_product_sense", 14},
		{"business_understanding", "requires_business_understanding", 12},
		{"communication_ability", "requires_communication", 10},
		{"business_understanding", "solution_consulting", 10},
		// Medium
		{"algorithm", "algorithm_core", 12},
		{"data", "data_science", 10},
		{"engineering_ability", "technical_research", 6},
		{"modeling_ability", "machine_learning_platform", 8},
		{"business_understanding", "business_function", 10},
		{"business", "business_function", 12},
		// Weak
		{"engineering_ability", "client_development", 4},
		{"algorithm", "backend_service", 2},
	}
	for _, p := range pairs {
		cv := cTags.get(p.cTag)
		jv := jTags.get(p.jTag)
		if cv > 0 && jv > 0 {
			mult := float64(cv+jv) / 6.0
			if cv >= 2 && jv >= 2 {
				mult = float64(cv+jv) / 5.0
			}
			score += int(float64(p.weight) * mult)
		}
	}
	return score
}

// ══════════════════════════════════════════════════════════════════════════
// Direction Fit / Tag Fit / Risk Level
// ══════════════════════════════════════════════════════════════════════════

func directionFit(profile string, jTags tagSet) string {
	// K9-1-M: profile-aware direction matching
	isAlgo := jTags.has("algorithm_core", 2) || jTags.has("data_science", 2) || jTags.has("technical_research", 2) ||
		jTags.has("nlp", 2) || jTags.has("multimodal", 2) || jTags.has("recommendation_system", 2) ||
		jTags.has("machine_learning_platform", 2)
	isBackend := jTags.has("backend_service", 2) || jTags.has("cloud_infrastructure", 2)
	isProduct := jTags.has("product_operation", 2) || jTags.has("requires_product_sense", 2)
	isSolution := jTags.has("solution_consulting", 2) || jTags.has("business_function", 2) || jTags.has("supply_chain_business", 2)

	// algorithm_research
	if profile == "algorithm_research" {
		if isAlgo { return "strong" }
		if isBackend { return "weak" }
		if isSolution { return "weak" }
		return "medium"
	}
	// backend_engineering
	if profile == "backend_engineering" {
		if isBackend { return "strong" }
		if isAlgo { return "weak" }
		if isSolution { return "weak" }
		return "medium"
	}
	// product_operation
	if profile == "product_operation" {
		if isProduct { return "strong" }
		if isSolution { return "medium" }
		if isAlgo || isBackend { return "weak" }
		return "medium"
	}
	// business_solution
	if profile == "business_solution" {
		if isSolution { return "strong" }
		if isProduct { return "medium" }
		if isAlgo || isBackend { return "weak" }
		return "medium"
	}
	// career_exploration / default
	if profile == "career_exploration" {
		if isAlgo || isBackend { return "weak" }
		if isProduct || isSolution { return "medium" }
		return "medium"
	}
	return "medium"
}

func tagFit(tagScore int) string {
	if tagScore >= 35 {
		return "strong"
	}
	if tagScore >= 20 {
		return "medium"
	}
	return "weak"
}

func riskLevel(dirFit string, tFit string, tagScore int) string {
	if dirFit == "weak" && tFit == "weak" {
		return "high"
	}
	if dirFit == "weak" || tFit == "weak" {
		return "medium"
	}
	if tagScore <= 10 {
		return "medium"
	}
	return "low"
}

// ══════════════════════════════════════════════════════════════════════════
// Label
// ══════════════════════════════════════════════════════════════════════════

func recommendationLabelV2(score int, dirFit, tFit, risk string, profile string, jTags tagSet) (string, string) {
	return recommendationLabelV3(score, dirFit, tFit, risk, profile, jTags, 0)
}

func recommendationLabelV3(score int, dirFit, tFit, risk string, profile string, jTags tagSet, precision int) (string, string) {
	isAlgo := jTags.has("algorithm_core", 2) || jTags.has("data_science", 2) || jTags.has("nlp", 2) || jTags.has("multimodal", 2)
	isBackend := jTags.has("backend_service", 2)
	isSolution := jTags.has("solution_consulting", 2)

	// K9-1-M: hard guard — algorithm_research + backend cannot be strong/high
	if profile == "algorithm_research" && isBackend {
		if score >= 70 { return "可尝试，需要补强", "try" }
		return "备选参考，不建议优先", "backup"
	}
	if profile == "algorithm_research" && isSolution {
		return "备选参考，不建议优先", "backup"
	}

	// K9-1-M: hard guard — backend_engineering + pure algorithm cannot be strong/high
	if profile == "backend_engineering" && isAlgo {
		if score >= 75 { return "可尝试，需要补强", "try" }
		return "备选参考，不建议优先", "backup"
	}

	// K9-1-M: hard guard — product_operation + algorithm/backend cannot be strong/high
	if profile == "product_operation" && (isAlgo || isBackend) {
		if score >= 70 { return "可尝试，需要补强", "try" }
		return "备选参考，不建议优先", "backup"
	}

	// K9-1-M: hard guard — business_solution + algorithm/backend capped at backup
	if profile == "business_solution" && (isAlgo || isBackend) {
		return "备选参考，不建议优先", "backup"
	}

	// K9-1-M: career_exploration — very conservative
	if profile == "career_exploration" {
		if isAlgo || isBackend { return "备选参考，不建议优先", "backup" }
		if score >= 80 { return "高匹配，重点关注", "high" }
		if score >= 70 { return "可尝试，需要补强", "try" }
		return "备选参考，不建议优先", "backup"
	}

	// K9-1-T-C: precision-aware strong
	// strong requires precision >= 5
	if score >= 90 && dirFit == "strong" && tFit == "strong" && risk != "high" && precision >= 5 {
		return "强匹配，优先投递", "strong"
	}
	// K9-1-T-C: risk hard downgrade — medium risk cannot be strong
	if score >= 88 && dirFit == "strong" && tFit == "strong" && risk == "low" && precision >= 3 {
		// Close to strong but either precision < 5 or score < 90 → high
	} else if score >= 90 && risk == "medium" {
		// Downgrade: medium risk cannot be strong even with high score
	}
	if score >= 82 && dirFit != "weak" && tFit != "weak" && risk != "high" {
		return "高匹配，重点关注", "high"
	}
	if score >= 72 && dirFit != "conflict" {
		return "可尝试，需要补强", "try"
	}
	if score >= 60 {
		return "备选参考，不建议优先", "backup"
	}
	return "不推荐", "not_recommended"
}

// ══════════════════════════════════════════════════════════════════════════
// Scoring Engine
// ══════════════════════════════════════════════════════════════════════════

type scoredJobV2 struct {
	job       *model.Job
	jp        jobTagProfile
	score     int
	dirFit    string
	tagFit    string
	riskLevel string
	precision int // K9-1-T-C: direction precision bonus
}

func scoreAllJobs(cp *candidateTagProfile, jobs []*model.Job, qualityMode bool) []scoredJobV2 {
	var scored []scoredJobV2
	for _, job := range jobs {
		jp := extractJobTags(job)
		tmScore := tagMatrixScore(cp.tags, jp.tags)
		dirF := directionFit(cp.dominantProfile, jp.tags)
		tF := tagFit(tmScore)
		risk := riskLevel(dirF, tF, tmScore)

		// K9-1-T-C: base 45 + capped matrix ≤34 for score headroom
		tmCapped := tmScore
		if tmCapped > 34 {
			tmCapped = 34
		}
		score := 45 + tmCapped

		// K9-1-T-C: minimal evidence bonuses (precision drives differentiation)
		if cp.tags.has("project_evidence", 2) {
			score += 2
		}
		if cp.tags.has("internship_evidence", 2) {
			score += 1
		}
		if cp.tags.has("research_evidence", 2) {
			score += 1
		}
		if cp.schoolLevel >= 2 {
			score += 1
		}

		if dirF == "weak" && score > 78 {
			score = 78
		}

		// Build temp scoredJob for precision calculation
		tmp := scoredJobV2{job: job, jp: jp, score: score, dirFit: dirF, tagFit: tF, riskLevel: risk}
		pbonus := precisionBonus(cp, tmp)

		// K9-1-T-C: risk penalty — penalize when specific domain evidence is missing
		d := detectJobDomain(jp)
		riskPenalty := 0
		if cp.dominantProfile == "algorithm_research" {
			switch d {
			case domainAlgorithmMultimodal:
				if !cp.tags.has("multimodal", 2) {
					riskPenalty += 3
				}
			case domainAlgorithmNLP:
				if !cp.tags.has("nlp", 2) {
					riskPenalty += 3
				}
			case domainAlgorithmCV:
				if !cp.tags.has("multimodal", 2) && !cp.tags.has("computer_vision", 1) {
					riskPenalty += 2
				}
			case domainAlgorithmGraphics:
				riskPenalty += 3 // low relevance for non-Graphics candidates
			case domainAlgorithmDataScience:
				if !cp.tags.has("data_analysis_ability", 2) && !cp.tags.has("data", 1) {
					riskPenalty += 2
				}
			case domainAIApplication:
				if !cp.tags.has("engineering_ability", 2) && !cp.tags.has("backend_service", 1) {
					riskPenalty += 2
				}
			}
		}

		score = score + pbonus - riskPenalty

		// K9-1-M: profile-aware score caps
		isAlgoJob := jp.tags.has("algorithm_core", 2) || jp.tags.has("data_science", 2)
		isBackendJob := jp.tags.has("backend_service", 2)

		switch cp.dominantProfile {
		case "algorithm_research":
			if isBackendJob && score > 74 {
				score = 74
			}
		case "backend_engineering":
			if isAlgoJob && score > 74 {
				score = 74
			}
		case "product_operation":
			if isAlgoJob || isBackendJob {
				if score > 69 {
					score = 69
				}
			}
		case "business_solution":
			if isAlgoJob || isBackendJob {
				if score > 69 {
					score = 69
				}
			}
		case "career_exploration":
			if isAlgoJob || isBackendJob {
				if score > 74 {
					score = 74
				}
			}
		}

		// K9-1-T-C: tiered score caps
		if dirF == "strong" && pbonus < 3 {
			if score > 90 {
				score = 90
			}
		}
		if risk == "medium" && score > 88 {
			score = 88
		}

		if score > 95 {
			score = 95
		}
		if score < 55 {
			score = 55
		}

		scored = append(scored, scoredJobV2{job: job, jp: jp, score: score, dirFit: dirF, tagFit: tF, riskLevel: risk, precision: pbonus})
	}

	// K9-1-T-C: sort by score, then by dominance profile preference
	sort.Slice(scored, func(i, j int) bool {
		if scored[i].score != scored[j].score {
			return scored[i].score > scored[j].score
		}
		aiAlgo := scored[i].jp.tags.has("algorithm_core", 2)
		ajAlgo := scored[j].jp.tags.has("algorithm_core", 2)
		aiBack := scored[i].jp.tags.has("backend_service", 2)
		ajBack := scored[j].jp.tags.has("backend_service", 2)
		switch cp.dominantProfile {
		case "algorithm_research":
			if aiAlgo && !ajAlgo {
				return true
			}
			if !aiAlgo && ajAlgo {
				return false
			}
		case "backend_engineering":
			if aiBack && !ajBack {
				return true
			}
			if !aiBack && ajBack {
				return false
			}
		}
		// Fallback: prefer higher precision
		return scored[i].precision > scored[j].precision
	})

	return scored
}

// ══════════════════════════════════════════════════════════════════════════
// Judge
// ══════════════════════════════════════════════════════════════════════════

func judgeRecommendationV2(cp *candidateTagProfile, ranked []scoredJobV2) model.RecommendationJudgeResult {
	var issues []string

	if len(ranked) == 0 {
		return model.RecommendationJudgeResult{Pass: false, Severity: "major", Issues: []string{"empty"}, FixStrategy: "fallback"}
	}

	// Check top-1 direction for algorithm_research
	if cp.dominantProfile == "algorithm_research" {
		if ranked[0].jp.tags.has("backend_service", 2) || ranked[0].jp.tags.has("solution_consulting", 2) {
			issues = append(issues, "top-1 is backend/business despite algorithm_research profile")
		}
		algoCount := 0
		for i := 0; i < min(3, len(ranked)); i++ {
			if ranked[i].jp.tags.has("algorithm_core", 2) || ranked[i].jp.tags.has("data_science", 2) ||
				ranked[i].jp.tags.has("nlp", 2) || ranked[i].jp.tags.has("multimodal", 2) ||
				ranked[i].jp.tags.has("technical_research", 2) {
				algoCount++
			}
		}
		if algoCount < 2 {
			issues = append(issues, fmt.Sprintf("top-3 only %d algorithm/research jobs", algoCount))
		}
	}

	// Check score homogeneity
	allSame := true
	for i := 1; i < len(ranked); i++ {
		if ranked[i].score != ranked[0].score {
			allSame = false
			break
		}
	}
	if allSame {
		issues = append(issues, "all scores identical")
	}

	if len(issues) > 0 {
		return model.RecommendationJudgeResult{
			Pass: false, Severity: "major", Issues: issues, FixStrategy: "rerank",
		}
	}
	return model.RecommendationJudgeResult{Pass: true, Severity: "pass", FixStrategy: "none"}
}

// ══════════════════════════════════════════════════════════════════════════
// Content Generators
// ══════════════════════════════════════════════════════════════════════════

func buildReasonV2(cp *candidateTagProfile, sj scoredJobV2) string {
	dp := cp.dominantProfile
	if sj.dirFit == "strong" {
		return fmt.Sprintf("你的%s画像与该岗位高度匹配，技能积累和项目经验可直接复用。", profileLabel(dp))
	}
	if sj.dirFit == "medium" {
		return fmt.Sprintf("你的%s背景与%s岗位有一定关联，技能可部分迁移。", profileLabel(dp), sj.job.Category)
	}
	if dp == "algorithm_research" && sj.jp.tags.has("backend_service", 2) {
		return fmt.Sprintf("你具备工程基础，但当前简历主线更偏算法研究，该后台开发岗位适合作为备选方向了解。")
	}
	return fmt.Sprintf("该岗位可作为求职参考，建议查看岗位要求进一步评估匹配度。")
}

func buildHighlightsV2(cp *candidateTagProfile, sj scoredJobV2) []string {
	var h []string
	if sj.dirFit == "strong" {
		h = append(h, "岗位方向与你的核心技能方向高度一致")
	} else {
		h = append(h, "岗位方向与简历背景存在部分交集")
	}
	if sj.tagFit == "strong" {
		h = append(h, "标签匹配度较高")
	}
	if cp.tags.has("research_ability", 2) && sj.jp.tags.has("algorithm_core", 2) {
		h = append(h, "研究能力与该岗位要求匹配")
	}
	if cp.tags.has("experiment_ability", 2) {
		h = append(h, "具备实验复现和模型评估经验")
	}
	if cp.degreeLevel >= 2 {
		h = append(h, fmt.Sprintf("学历背景（%s）有竞争力", degreeLabel(cp.degreeLevel)))
	}
	if len(h) < 2 {
		h = append(h, "建议进一步查看岗位描述以确认匹配度")
	}
	return h[:min(3, len(h))]
}

func buildRisksV2(cp *candidateTagProfile, sj scoredJobV2) []string {
	var r []string
	if sj.dirFit == "weak" && cp.dominantProfile == "algorithm_research" && sj.jp.tags.has("backend_service", 2) {
		r = append(r, "岗位偏后台开发方向，与你的算法研究主线差距较大")
	} else if sj.dirFit == "weak" {
		r = append(r, fmt.Sprintf("岗位核心方向与你的主要技能方向不完全一致"))
	}
	if cp.dominantProfile == "algorithm_research" && sj.jp.tags.has("backend_service", 2) {
		r = append(r, "简历中缺少完整后端业务系统项目和生产环境性能优化证据")
	}
	if len(r) == 0 {
		r = append(r, "建议补充与该岗位要求更匹配的项目细节")
	}
	return r
}

func buildActionV2(cp *candidateTagProfile, sj scoredJobV2) string {
	if sj.score >= 80 && sj.dirFit == "strong" {
		return fmt.Sprintf("建议优先查看岗位详情并进行匹配分析。%s方向是你的强项。", sj.job.Category)
	}
	if sj.dirFit == "weak" && cp.dominantProfile == "algorithm_research" {
		return fmt.Sprintf("该岗位属于%s方向，与你的主业不同。建议优先寻找算法/数据科学方向岗位。", sj.job.Category)
	}
	if sj.score >= 70 {
		return fmt.Sprintf("建议关注该岗位并针对性补充相关技能储备。")
	}
	return "建议先查看岗位详情，了解具体要求后再判断是否适合投递。"
}

func buildEvidenceV2(cp *candidateTagProfile, sj scoredJobV2) []model.RecommendationEvidence {
	var ev []model.RecommendationEvidence
	if cp.tags.has("modeling_ability", 2) && sj.jp.tags.has("requires_modeling_ability", 2) {
		ev = append(ev, model.RecommendationEvidence{ResumeSignal: "建模能力", JobRequirement: "岗位要求建模能力", Strength: "strong", Explanation: "简历中有PyTorch和深度学习项目经验"})
	}
	if cp.tags.has("experiment_ability", 2) && sj.jp.tags.has("requires_experiment_ability", 2) {
		ev = append(ev, model.RecommendationEvidence{ResumeSignal: "实验能力", JobRequirement: "岗位要求实验评估能力", Strength: "strong", Explanation: "简历中有AUC/NDCG/消融实验等实验经验"})
	}
	if cp.tags.has("research_ability", 2) {
		ev = append(ev, model.RecommendationEvidence{ResumeSignal: "研究能力", JobRequirement: "岗位要求研究能力", Strength: "medium", Explanation: "简历中有研究方向和论文成果"})
	}
	if cp.degreeLevel >= 2 {
		ev = append(ev, model.RecommendationEvidence{ResumeSignal: fmt.Sprintf("学历：%s", degreeLabel(cp.degreeLevel)), JobRequirement: "岗位要求较强理论基础", Strength: "medium", Explanation: "研究生学历表明具备深入研究能力"})
	}
	return ev
}

// ══════════════════════════════════════════════════════════════════════════
// K9-1-T-C: V3 Content Generators (per-job differentiated)
// ══════════════════════════════════════════════════════════════════════════

func buildWhyRankedHereV3(rank int, cp *candidateTagProfile, sj scoredJobV2, allRanked []scoredJobV2) string {
	domain := domainLabel(detectJobDomain(sj.jp))

	if rank == 1 {
		total := len(allRanked)
		if sj.precision >= 6 {
			return fmt.Sprintf("该岗位在%s与你的简历证据匹配度最高，在%d个候选岗位中排名第1。", domain, total)
		}
		return fmt.Sprintf("该岗位在%s与你的核心方向高度匹配，综合评分排名第1。", domain)
	}

	prev := allRanked[rank-2]
	prevDomain := domainLabel(detectJobDomain(prev.jp))
	scoreGap := prev.score - sj.score

	if sj.precision < prev.precision && scoreGap > 0 {
		return fmt.Sprintf("该岗位%s与你的方向关联度略低于Top%d（%s），因细分证据差异排在第%d。", domain, rank-1, prevDomain, rank)
	}
	if scoreGap > 3 {
		return fmt.Sprintf("该岗位%s在方向匹配强度上与Top%d存在%d分差距，排在第%d。", domain, rank-1, scoreGap, rank)
	}
	if sj.dirFit == "weak" {
		return fmt.Sprintf("该岗位方向与你的主要技能方向存在偏差，作为备选排在第%d。", rank)
	}
	return fmt.Sprintf("该岗位%s与你的背景有一定关联，但匹配证据强度低于前列岗位，排在第%d。", domain, rank)
}

func buildResumeEvidenceV3(cp *candidateTagProfile, sj scoredJobV2) []string {
	var ev []string
	d := detectJobDomain(sj.jp)

	// Extract project names from candidate profile
	projNames := extractProjectNames(cp)
	internPositions := extractInternPositions(cp)
	eduDirection := extractEduDirection(cp)

	switch cp.dominantProfile {
	case "algorithm_research":
		// Add research direction evidence if available
		if eduDirection != "" {
			ev = append(ev, fmt.Sprintf("%s方向研究", eduDirection))
		}
		// Add project evidence
		if len(projNames) > 0 {
			ev = append(ev, fmt.Sprintf("%s项目经验", projNames[0]))
			if len(projNames) > 1 {
				ev = append(ev, fmt.Sprintf("%s项目经验", projNames[1]))
			}
		}
		if len(internPositions) > 0 && len(ev) < 3 {
			ev = append(ev, fmt.Sprintf("%s实习经历", internPositions[0]))
		}
		// Domain-specific tags
		if cp.tags.has("modeling_ability", 2) && len(ev) < 3 {
			ev = append(ev, "深度学习模型训练经验")
		}
		if cp.tags.has("experiment_ability", 2) && len(ev) < 3 {
			ev = append(ev, "实验评估与消融分析能力")
		}

	case "backend_engineering":
		if len(projNames) > 0 {
			ev = append(ev, fmt.Sprintf("%s项目经验", projNames[0]))
		}
		if len(internPositions) > 0 {
			ev = append(ev, fmt.Sprintf("%s实习经历", internPositions[0]))
		}
		if cp.tags.has("engineering_ability", 2) && len(ev) < 3 {
			ev = append(ev, "工程开发与系统设计能力")
		}

	case "product_operation":
		if len(internPositions) > 0 {
			ev = append(ev, fmt.Sprintf("%s实习经历", internPositions[0]))
		}
		if len(projNames) > 0 && len(ev) < 3 {
			ev = append(ev, fmt.Sprintf("%s项目经验", projNames[0]))
		}
		if cp.tags.has("product_sense", 2) && len(ev) < 3 {
			ev = append(ev, "产品分析与用户研究能力")
		}

	case "business_solution":
		if len(internPositions) > 0 {
			ev = append(ev, fmt.Sprintf("%s实习经历", internPositions[0]))
		}
		if cp.tags.has("business_understanding", 2) && len(ev) < 3 {
			ev = append(ev, "商务沟通与需求理解能力")
		}
	}

	// Fallback: use candidate tags for evidence
	if len(ev) == 0 {
		ev = skillBasedEvidence(cp, d)
	}
	if len(ev) == 0 {
		ev = append(ev, "简历背景与岗位方向具有相关性")
	}

	return ev[:min(3, len(ev))]
}

func extractProjectNames(cp *candidateTagProfile) []string {
	var names []string
	for _, line := range splitNonEmpty(strings.Split(cp.projText, " ")) {
		if len(line) > 2 && !isCommonWord(line) {
			names = append(names, line)
			if len(names) >= 2 {
				break
			}
		}
	}
	return names
}

func extractInternPositions(cp *candidateTagProfile) []string {
	var pos []string
	for _, line := range splitNonEmpty(strings.Split(cp.internText, " ")) {
		if len(line) > 2 && !isCommonWord(line) {
			pos = append(pos, line)
			if len(pos) >= 1 {
				break
			}
		}
	}
	return pos
}

func extractEduDirection(cp *candidateTagProfile) string {
	if cp.tags.has("recommendation_system", 2) {
		return "推荐系统"
	}
	if cp.tags.has("nlp", 2) {
		return "自然语言处理"
	}
	if cp.tags.has("multimodal", 2) {
		return "多模态"
	}
	return ""
}

func skillBasedEvidence(cp *candidateTagProfile, d jobDomain) []string {
	var ev []string
	switch d {
	case domainAlgorithmMultimodal, domainAlgorithmCV:
		if cp.tags.has("modeling_ability", 2) {
			ev = append(ev, "深度学习建模基础")
		}
	case domainAlgorithmNLP:
		if cp.tags.has("modeling_ability", 2) {
			ev = append(ev, "序列建模与实验能力")
		}
	case domainBackend:
		if cp.tags.has("engineering_ability", 2) {
			ev = append(ev, "工程化开发能力")
		}
	}
	return ev
}

func splitNonEmpty(parts []string) []string {
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func isCommonWord(s string) bool {
	common := []string{"the", "a", "an", "and", "or", "of", "in", "to", "is", "with", "for", "on", "as", "at", "by", "de", "la", "le", "el", "en"}
	for _, w := range common {
		if strings.EqualFold(s, w) {
			return true
		}
	}
	return false
}

func buildJobSpecificReasonV3(cp *candidateTagProfile, sj scoredJobV2) string {
	d := detectJobDomain(sj.jp)
	domain := domainLabel(d)
	category := sj.job.Category
	if category == "" {
		category = domain
	}

	switch cp.dominantProfile {
	case "algorithm_research":
		switch d {
		case domainAlgorithmMultimodal:
			return fmt.Sprintf("%s岗位需要模型实验和跨模态理解能力，你的推荐系统实验和深度学习基础具备迁移价值，但简历中还缺少图文或视频理解项目。", domain)
		case domainAlgorithmNLP:
			return fmt.Sprintf("%s岗位与推荐系统在模型训练和数据实验上有共通方法论，但你目前缺少NLP专项项目（如文本分类、命名实体识别、预训练语言模型微调）。", domain)
		case domainAlgorithmDataScience:
			return fmt.Sprintf("%s岗位需要较强的数据工程和业务分析能力，你的简历在模型实验方面扎实，但在ETL、数据管道和业务指标分析方面的证据不足。", domain)
		case domainAlgorithmCV:
			return fmt.Sprintf("%s岗位需要对视觉任务有深入理解，你的深度学习基础可部分迁移，但缺少计算机视觉专项项目经验。", domain)
		case domainAlgorithmGraphics:
			return fmt.Sprintf("%s岗位需要图形学和渲染知识，你的深度学习和实验能力可迁移到图形/视觉方向，但简历中缺少相关证据。", domain)
		case domainAIApplication:
			return fmt.Sprintf("%s岗位需要将算法落地为工程应用，你的算法基础和工程能力有迁移价值，但缺少AI服务化和工程化部署经验。", domain)
		case domainAlgorithmGeneral:
			return fmt.Sprintf("%s岗位与你的算法研究方向有较强关联，推荐系统实验和模型评估经验可以迁移。", domain)
		default:
			return fmt.Sprintf("你的算法研究背景与%s岗位存在一定关联，建议查看岗位要求进一步评估。", category)
		}

	case "backend_engineering":
		switch d {
		case domainBackend:
			return fmt.Sprintf("%s岗位与你的后端开发能力高度匹配，接口设计、数据库和分布式系统经验可直接复用。", domain)
		case domainFullstack:
			return fmt.Sprintf("%s岗位需要前后端综合能力，你的后端工程基础有迁移价值，但前端或全栈项目经验可能需要补充。", domain)
		case domainDataEngineering:
			return fmt.Sprintf("%s岗位需要数据平台和管道建设能力，你的工程基础可迁移，但大数据平台（Spark/Hive）经验可能需要补强。", domain)
		case domainAIApplication:
			return fmt.Sprintf("%s岗位需要工程化落地AI能力，你的后端基础可支撑AI服务化，但模型调用和推理优化经验需要补充。", domain)
		case domainClient:
			return fmt.Sprintf("%s岗位侧重客户端开发，与你的后端主线有差异，需要评估技术栈匹配度。", domain)
		default:
			return fmt.Sprintf("你的后台工程背景与%s岗位有一定关联，建议查看技术要求进一步评估。", category)
		}

	case "product_operation":
		switch d {
		case domainProduct:
			return fmt.Sprintf("%s岗位需要需求分析和产品设计能力，与你的产品背景匹配。", domain)
		case domainOperation:
			return fmt.Sprintf("%s岗位需要用户运营和数据驱动增长能力，与你的运营经验有一定关联。", domain)
		default:
			return fmt.Sprintf("你的产品背景与%s岗位存在部分交集，建议查看岗位要求。", category)
		}

	case "business_solution":
		switch d {
		case domainBusinessSolution:
			return fmt.Sprintf("%s岗位需要客户需求和方案设计能力，与你的商务分析背景匹配。", domain)
		case domainBusinessFunction:
			return fmt.Sprintf("%s岗位需要综合协调和文档处理能力，与你的职能经验有契合点。", domain)
		default:
			return fmt.Sprintf("你的商务背景与%s岗位有一定关联。", category)
		}

	case "career_exploration":
		if sj.dirFit == "weak" {
			return fmt.Sprintf("%s岗位方向与你的简历背景存在较大差距，建议优先探索与你背景更匹配的方向。", category)
		}
		return fmt.Sprintf("你的简历背景与%s岗位存在部分交集，可以作为职业探索方向了解。", category)
	}

	return fmt.Sprintf("你的背景与该%s岗位存在部分交集，建议查看详细要求进一步评估。", category)
}

func buildJobSpecificRisksV3(cp *candidateTagProfile, sj scoredJobV2) []string {
	var risks []string
	d := detectJobDomain(sj.jp)

	switch cp.dominantProfile {
	case "algorithm_research":
		switch d {
		case domainAlgorithmMultimodal:
			risks = append(risks, "缺少明确多模态项目经历")
			risks = append(risks, "图像/视频理解相关证据不足")
		case domainAlgorithmNLP:
			risks = append(risks, "缺少NLP专项项目或论文")
			risks = append(risks, "未涉及Transformer/预训练语言模型应用")
		case domainAlgorithmDataScience:
			risks = append(risks, "缺少数据工程和pipeline项目")
			risks = append(risks, "缺少业务分析和A/B测试经验")
		case domainAlgorithmCV:
			risks = append(risks, "缺少计算机视觉专项项目经历")
			risks = append(risks, "图像分类/目标检测/分割等任务经验不足")
		case domainAlgorithmGraphics:
			risks = append(risks, "缺少渲染、图形学相关经历")
			risks = append(risks, "OpenGL/Unity/3D建模等技能证据不足")
		case domainAIApplication:
			risks = append(risks, "后端工程化、接口服务、线上部署证据不足")
			risks = append(risks, "缺少AI模型服务化和推理优化经验")
		default:
			if sj.dirFit == "weak" {
				risks = append(risks, fmt.Sprintf("岗位核心方向与你的主要技能方向不完全一致"))
			}
		}

	case "backend_engineering":
		switch d {
		case domainClient:
			risks = append(risks, "缺少客户端开发经验证据")
			risks = append(risks, "C++/桌面框架/图形界面项目经历不足")
		case domainDataEngineering:
			risks = append(risks, "缺少大数据平台（Spark/Hive/Hadoop）实操经验")
		case domainAIApplication:
			risks = append(risks, "缺少AI模型调用和推理部署经验")
		}
		if sj.dirFit == "weak" {
			risks = append(risks, "岗位方向与你的后端主线存在偏差")
		}

	case "product_operation":
		if sj.dirFit == "weak" {
			risks = append(risks, "该岗位偏技术方向，与你的产品主线差距较大")
		}

	case "business_solution":
		if sj.dirFit == "weak" {
			risks = append(risks, "该岗位方向与你的商务背景存在差距")
		}
	}

	if len(risks) == 0 {
		if sj.dirFit == "weak" {
			risks = append(risks, "岗位方向与你的背景存在较大差异")
		}
	}
	if len(risks) == 0 {
		risks = append(risks, "建议查看岗位详细要求以确认匹配度")
	}
	return risks
}

func buildSuggestedImprovementV3(cp *candidateTagProfile, sj scoredJobV2) string {
	d := detectJobDomain(sj.jp)

	switch cp.dominantProfile {
	case "algorithm_research":
		switch d {
		case domainAlgorithmMultimodal:
			return "建议在简历中补充图文检索、视频理解或多模态模型实验经历，并说明数据集、模型结构和评估指标。"
		case domainAlgorithmNLP:
			return "建议增加至少1个NLP方向项目（如文本分类/情感分析/问答系统），展示对Transformer架构的实际使用经验。"
		case domainAlgorithmDataScience:
			return "建议增加数据清洗/ETL项目，或展示在业务场景中使用数据驱动决策的经历。"
		case domainAlgorithmCV:
			return "建议补充图像分类或目标检测项目，展示OpenCV/PyTorch视觉模型训练经验。"
		case domainAlgorithmGraphics:
			return "建议补充图形学或视觉相关项目，如3D渲染、图像处理或视觉特效。"
		case domainAIApplication:
			return "建议补充AI服务化项目：模型接口封装、Docker部署、性能优化和监控。"
		default:
			return "建议继续深化算法方向项目，补充更多模型实验与评估证据。"
		}

	case "backend_engineering":
		switch d {
		case domainClient:
			return "如对该方向有兴趣，建议补充C++客户端或跨平台开发项目经验。"
		case domainDataEngineering:
			return "建议补充Spark/Hive数据处理项目，展示大规模数据管道建设能力。"
		case domainAIApplication:
			return "建议补充AI模型部署和在线推理服务相关项目经验。"
		default:
			return "建议继续强化后端工程项目，补充分布式系统或高并发实践经验。"
		}

	case "product_operation":
		return "建议补充产品数据分析、用户调研或A/B测试相关经验。"

	case "business_solution":
		return "建议补充行业分析、客户需求沟通或方案策划相关经验。"

	case "career_exploration":
		return "建议先明确职业方向，再针对目标岗位补充相关项目或实习经历。"
	}

	return "建议查看岗位详细要求，针对性地补充相关技能和项目经历。"
}

// ══════════════════════════════════════════════════════════════════════════
// Label Calibration (K9-1-T-C)
// ══════════════════════════════════════════════════════════════════════════

type labeledJob struct {
	job             *model.RecommendedJob
	priority        string
	label           string
	rank            int
	precision       int
	specificRiskLen int
}

func calibrateLabels(lbls []labeledJob) {
	strongCount := 0
	for i := range lbls {
		lj := &lbls[i]

		if lj.priority == "strong" {
			// Rule 1: strong cap at 2
			if strongCount >= 2 {
				lj.priority = "high"
				lj.label = "高匹配，重点关注"
				continue
			}
			// Rule 2: rank > 3 cannot be strong
			if lj.rank > 3 {
				lj.priority = "high"
				lj.label = "高匹配，重点关注"
				continue
			}
			// Rule 3: precision < 5 cannot be strong
			if lj.precision < 5 {
				lj.priority = "high"
				lj.label = "高匹配，重点关注"
				continue
			}
			// Rule 4: 2+ specific risks cannot be strong
			if lj.specificRiskLen >= 2 {
				lj.priority = "high"
				lj.label = "高匹配，重点关注"
				continue
			}
			strongCount++
		}
	}
}

// ══════════════════════════════════════════════════════════════════════════
// Build
// ══════════════════════════════════════════════════════════════════════════

func buildResultsV2(cp *candidateTagProfile, ranked []scoredJobV2, limit int) []model.RecommendedJob {
	var results []model.RecommendedJob
	var lbls []labeledJob

	for i := 0; i < min(limit, len(ranked)); i++ {
		sj := ranked[i]
		rank := i + 1

		// K9-1-T-C: use V3 label with precision awareness
		label, priority := recommendationLabelV3(sj.score, sj.dirFit, sj.tagFit, sj.riskLevel, cp.dominantProfile, sj.jp.tags, sj.precision)

		// Keep V2 builders for existing fields (compatibility)
		reason := buildReasonV2(cp, sj)
		highlights := buildHighlightsV2(cp, sj)
		// Use V3 risks for existing RiskPoints too, so card displays differentiate
		risks := buildJobSpecificRisksV3(cp, sj)
		action := buildActionV2(cp, sj)
		ev := buildEvidenceV2(cp, sj)

		// K9-1-T-C: V3 differentiated builders for new fields
		whyRanked := buildWhyRankedHereV3(rank, cp, sj, ranked[:min(limit, len(ranked))])
		resumeEv := buildResumeEvidenceV3(cp, sj)
		jobReason := buildJobSpecificReasonV3(cp, sj)
		jobRisks := buildJobSpecificRisksV3(cp, sj)
		improvement := buildSuggestedImprovementV3(cp, sj)

		results = append(results, model.RecommendedJob{
			JobID:                  sj.job.ID,
			Title:                  sj.job.Title,
			Department:             sj.job.Department,
			City:                   sj.job.City,
			Category:               sj.job.Category,
			JobType:                sj.job.JobType,
			Priority:               priority,
			RecommendationScore:    sj.score,
			RecommendationLabel:    label,
			RecommendationPriority: priority,
			RecommendationReason:   reason,
			MatchHighlights:        highlights,
			RiskPoints:             risks,
			SuggestedAction:        action,
			ExternalSource:         sj.job.ExternalSource,
			ExternalID:             sj.job.ExternalID,
			DataQuality:            sj.job.DataQuality,
			SourceURL:              sj.job.SourceURL,
			IsFeatured:             sj.job.IsFeatured,
			ApplyReadiness:         getReadinessV2(sj.score),
			OptimizationStopSignal: getStopSignalV2(sj.score),
			MatchEvidence:          ev,
			// K9-1-T-C: new per-job explanation fields
			WhyRankedHere:        whyRanked,
			ResumeEvidence:       resumeEv,
			JobSpecificReason:    jobReason,
			JobSpecificRisks:     jobRisks,
			SuggestedImprovement: improvement,
		})
		lbls = append(lbls, labeledJob{
			job:             &results[i],
			priority:        priority,
			label:           label,
			rank:            rank,
			precision:       sj.precision,
			specificRiskLen: len(jobRisks),
		})
	}

	// K9-1-T-C: post-process label calibration (strong ≤ 2, etc.)
	calibrateLabels(lbls)
	for i := range lbls {
		results[i].RecommendationLabel = lbls[i].label
		results[i].RecommendationPriority = lbls[i].priority
	}

	return results
}

func getReadinessV2(score int) string {
	if score >= 80 {
		return "ready_to_apply"
	}
	if score >= 60 {
		return "quick_fix_before_apply"
	}
	return "not_ready_for_this_role"
}
func getStopSignalV2(score int) string {
	if score >= 80 {
		return "stop_and_apply"
	}
	if score >= 65 {
		return "fix_top_3_then_apply"
	}
	return "needs_more_evidence"
}

// ══════════════════════════════════════════════════════════════════════════
// Entry Points
// ══════════════════════════════════════════════════════════════════════════

func (a *RecommendationMockAgent) Recommend(_ context.Context, resume *model.ResumeVersion, jobs []*model.Job) ([]model.RecommendedJob, error) {
	cp := extractCandidateTags(resume)
	ranked := scoreAllJobs(&cp, jobs, false)
	return buildResultsV2(&cp, ranked, 5), nil
}

func (a *RecommendationMockAgent) RecommendQuality(_ context.Context, resume *model.ResumeVersion, jobs []*model.Job) ([]model.RecommendedJob, error) {
	cp := extractCandidateTags(resume)
	ranked := scoreAllJobs(&cp, jobs, true)
	judge := judgeRecommendationV2(&cp, ranked)
	if !judge.Pass && judge.Severity == "major" && judge.FixStrategy == "rerank" {
		ranked = scoreAllJobs(&cp, jobs, true)
	}
	return buildResultsV2(&cp, ranked, 5), nil
}

// K9-1-H4-C: recommend using stored profile tags (consumes resume_profiles)
func (a *RecommendationMockAgent) RecommendWithStoredProfile(resume *model.ResumeVersion, jobs []*model.Job, dominantProfile string, candidateTags, riskTags []StoredTag) ([]model.RecommendedJob, error) {
	cp := buildProfileFromStored(resume, dominantProfile, candidateTags, riskTags)
	ranked := scoreAllJobs(&cp, jobs, false)
	return buildResultsV2(&cp, ranked, 5), nil
}

func (a *RecommendationMockAgent) RecommendQualityWithStoredProfile(resume *model.ResumeVersion, jobs []*model.Job, dominantProfile string, candidateTags, riskTags []StoredTag) ([]model.RecommendedJob, error) {
	cp := buildProfileFromStored(resume, dominantProfile, candidateTags, riskTags)
	ranked := scoreAllJobs(&cp, jobs, true)
	judge := judgeRecommendationV2(&cp, ranked)
	if !judge.Pass && judge.Severity == "major" && judge.FixStrategy == "rerank" {
		ranked = scoreAllJobs(&cp, jobs, true)
	}
	return buildResultsV2(&cp, ranked, 5), nil
}

// StoredTag is the flattened tag format from resume_profiles JSON.
type StoredTag struct {
	Tag        string `json:"tag"`
	Group      string `json:"group,omitempty"`
	Confidence int    `json:"confidence"`
	Evidence   string `json:"evidence,omitempty"`
}

func buildProfileFromStored(resume *model.ResumeVersion, dominant string, candTags, riskTags []StoredTag) candidateTagProfile {
	cp := candidateTagProfile{
		tags:            make(tagSet),
		dominantProfile: dominant,
		allText:         strings.ToLower(resume.RawText),
	}
	if cp.allText == "" {
		cp.allText = strings.ToLower(resume.ProfessionalSkills + " " + resume.AdditionalInfo)
	}
	for _, t := range candTags {
		if t.Confidence < 1 {
			t.Confidence = 1
		}
		cp.tags.set(t.Tag, t.Confidence)
	}
	for _, t := range riskTags {
		cp.tags.set(t.Tag, t.Confidence)
	}
	// Fill required text fields
	cp.skillsText = strings.ToLower(resume.ProfessionalSkills)
	cp.eduText = strings.ToLower(educationsText(resume))
	cp.projText = strings.ToLower(projectsText(resume))
	cp.internText = strings.ToLower(internshipsText(resume))
	cp.schoolLevel = 1
	cp.degreeLevel = 1
	// K9-1-M: Override tags based on profile-determined dominant
	if dominant == "algorithm_research" {
		cp.tags.set("algorithm", maxOf(cp.tags.get("algorithm"), 2))
		cp.tags.set("research_ability", maxOf(cp.tags.get("research_ability"), 2))
	}
	if dominant == "backend_engineering" {
		cp.tags.set("backend", maxOf(cp.tags.get("backend"), 2))
		cp.tags.set("engineering_ability", maxOf(cp.tags.get("engineering_ability"), 2))
	}
	if dominant == "product_operation" {
		cp.tags.set("product", maxOf(cp.tags.get("product"), 2))
		cp.tags.set("product_sense", maxOf(cp.tags.get("product_sense"), 2))
		cp.tags.set("business_understanding", maxOf(cp.tags.get("business_understanding"), 2))
	}
	if dominant == "business_solution" {
		cp.tags.set("business", maxOf(cp.tags.get("business"), 2))
		cp.tags.set("business_understanding", maxOf(cp.tags.get("business_understanding"), 2))
		cp.tags.set("communication_ability", maxOf(cp.tags.get("communication_ability"), 2))
	}
	return cp
}

func educationsText(r *model.ResumeVersion) string {
	var parts []string
	for _, e := range r.Educations {
		parts = append(parts, e.School+" "+e.Degree+" "+e.Major+" "+e.ResearchDirection)
	}
	return strings.Join(parts, " ")
}

func projectsText(r *model.ResumeVersion) string {
	var parts []string
	for _, p := range r.Projects {
		parts = append(parts, p.Name+" "+p.Description)
	}
	return strings.Join(parts, " ")
}

func internshipsText(r *model.ResumeVersion) string {
	var parts []string
	for _, i := range r.Internships {
		parts = append(parts, i.Company+" "+i.Position+" "+i.Description)
	}
	return strings.Join(parts, " ")
}

// ══════════════════════════════════════════════════════════════════════════
// Helpers
// ══════════════════════════════════════════════════════════════════════════

func clampMin(v, target int) int {
	if v < target {
		return target
	}
	return v
}
func maxOf(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func topTags(ts tagSet, n int) []string {
	type kv struct{ k string; v int }
	var kvs []kv
	for k, v := range ts {
		kvs = append(kvs, kv{k, v})
	}
	sort.Slice(kvs, func(i, j int) bool { return kvs[i].v > kvs[j].v })
	var result []string
	for i := 0; i < min(n, len(kvs)); i++ {
		result = append(result, kvs[i].k)
	}
	return result
}
func profileLabel(dp string) string {
	switch dp {
	case "algorithm_research":
		return "算法研究"
	case "backend_engineering":
		return "后台工程"
	default:
		return "综合"
	}
}
func degreeLabel(l int) string {
	switch l {
	case 3:
		return "博士"
	case 2:
		return "硕士"
	case 1:
		return "本科"
	default:
		return "高等教育"
	}
}
