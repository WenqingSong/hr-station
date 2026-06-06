package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"offer-catcher/server/internal/model"
)

// MatchMockAgent generates mock match analysis using simple rule-based scoring.
// It does NOT call any real LLM or read LLM_API_KEY.
type MatchMockAgent struct{}

// NewMatchMockAgent creates a new MatchMockAgent.
func NewMatchMockAgent() *MatchMockAgent {
	return &MatchMockAgent{}
}

// Analyze generates a mock match analysis based on the resume and target job.
func (a *MatchMockAgent) Analyze(_ context.Context, resume *model.ResumeVersion, job *model.Job) (*model.MatchAnalysisResult, error) {
	score := calculateMatchScore(resume, job)
	screeningResult := getScreeningResult(score)
	evidenceChain := generateEvidenceChain(resume, job)
	highlights := generateMatchHighlights(resume, job)
	risks := generateRiskPoints(resume, job)
	suggestions := generateOptimizationSuggestions(resume, job)
	nextSteps := generateNextSteps(suggestions)
	notice := "以下建议仅用于表达优化，请只补充真实经历，不要虚构项目或能力。"

	// For summary_only jobs, add degradation notice
	if job.DataQuality == "summary_only" {
		notice = "当前岗位详情信息不足（data_quality=summary_only），匹配分析结果可能不够完整。建议查看岗位完整 JD 后再进行深度分析。" + notice
		risks = append([]string{"岗位详情信息不足，匹配分析可能不够准确"}, risks...)
	}

	// Determine apply_readiness and optimization_stop_signal
	applyReadiness, stopSignal := determineMatchApplyReadiness(score, resume, job)
	topFixItems := generateMatchTopFixItems(resume, job)

	return &model.MatchAnalysisResult{
		JobID:                   job.ID,
		JobTitle:                job.Title,
		ResumeID:         resume.ID,
		MatchScore:              score,
		ScreeningResult:         screeningResult,
		EvidenceChain:           evidenceChain,
		MatchHighlights:         highlights,
		RiskPoints:              risks,
		OptimizationSuggestions: suggestions,
		NextSteps:               nextSteps,
		AuthenticityNotice:      notice,
		ExternalSource:          job.ExternalSource,
		ExternalID:              job.ExternalID,
		DataQuality:             job.DataQuality,
		SourceURL:               job.SourceURL,
		IsFeatured:              job.IsFeatured,
		ApplyReadiness:          applyReadiness,
		OptimizationStopSignal:  stopSignal,
		TopFixItems:             topFixItems,
	}, nil
}

// calculateMatchScore computes a match score (0-100) based on simple rules.
func calculateMatchScore(resume *model.ResumeVersion, job *model.Job) int {
	score := 0

	targetCity := strings.ToLower(strings.TrimSpace(resume.TargetCity))
	targetRole := strings.ToLower(strings.TrimSpace(resume.TargetRole))
	jobCity := strings.ToLower(job.City)
	jobCategory := strings.ToLower(job.Category)
	jobTitle := strings.ToLower(job.Title)

	// City match (+15)
	if targetCity != "" && strings.Contains(jobCity, targetCity) {
		score += 15
	}

	// Role/category match (+25)
	if targetRole != "" {
		if containsAny(targetRole, jobCategory, jobTitle) {
			score += 25
		}
		roleKeywords := strings.Fields(targetRole)
		for _, kw := range roleKeywords {
			if len(kw) >= 2 && (strings.Contains(jobCategory, kw) || strings.Contains(jobTitle, kw)) {
				score += 8
				break
			}
		}
	}

	// Skill overlap (+5 per matched skill, max 35)
	resumeSkills := parseJSONStringArray(resume.SkillsJSON)
	jobSkills := parseJSONStringArray(job.SkillKeywordsJSON)

	// If skill_keywords is empty, extract from description/requirements
	if len(jobSkills) == 0 && (job.Description != "" || job.Requirements != "") {
		jobSkills = extractKeywordsFromJD(job.Description, job.Requirements)
	}

	resumeSkillSet := toLowerSet(resumeSkills)
	jobSkillSet := toLowerSet(jobSkills)
	overlap := 0
	for skill := range resumeSkillSet {
		if jobSkillSet[skill] {
			overlap++
		}
	}
	score += overlap * 5
	if overlap > 5 {
		score += 5 // bonus for high overlap
	}

	// Key technology bonus (+3 each, max 15)
	keywords := []string{"go", "mysql", "redis", "docker", "kubernetes", "云", "分布式", "微服务", "grpc", "etcd"}
	bonus := 0
	for _, kw := range keywords {
		if resumeSkillSet[kw] && jobSkillSet[kw] {
			bonus += 3
			if bonus >= 15 {
				break
			}
		}
	}
	score += bonus

	// Experience/project mention bonus (+5 each, max 10)
	if len(resume.ExperienceJSON) > 0 {
		var experiences []map[string]interface{}
		if json.Unmarshal(resume.ExperienceJSON, &experiences) == nil && len(experiences) > 0 {
			score += 5
		}
	}
	if len(resume.ProjectJSON) > 0 {
		var projects []map[string]interface{}
		if json.Unmarshal(resume.ProjectJSON, &projects) == nil && len(projects) > 0 {
			score += 5
		}
	}

	// data_quality=full_jd bonus (+5)
	if job.DataQuality == "full_jd" {
		score += 5
	}

	// summary_only penalty (-10)
	if job.DataQuality == "summary_only" {
		score -= 10
	}

	// Cap at 100
	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}

	return score
}

// getScreeningResult returns the screening result based on the match score.
func getScreeningResult(score int) string {
	if score >= 85 {
		return "highly_likely"
	}
	if score >= 70 {
		return "likely"
	}
	if score >= 55 {
		return "possible"
	}
	return "risky"
}

// generateEvidenceChain creates JD-resume evidence chain entries.
func generateEvidenceChain(resume *model.ResumeVersion, job *model.Job) []model.MatchEvidence {
	var chain []model.MatchEvidence

	resumeSkills := parseJSONStringArray(resume.SkillsJSON)
	jobSkills := parseJSONStringArray(job.SkillKeywordsJSON)

	// If skill_keywords_json is empty, extract from description/requirements
	if len(jobSkills) == 0 && (job.Description != "" || job.Requirements != "") {
		jobSkills = extractKeywordsFromJD(job.Description, job.Requirements)
	}

	resumeSkillSet := toLowerSet(resumeSkills)

	// Check each job skill requirement against resume
	for _, req := range jobSkills {
		if resumeSkillSet[strings.ToLower(req)] {
			chain = append(chain, model.MatchEvidence{
				Requirement:  fmt.Sprintf("岗位要求 %s 相关经验", req),
				ResumePoint:  getResumeEvidenceForSkill(resume, req),
				Judgement:    "strong_match",
				ResumeModule: "skills",
			})
		} else {
			chain = append(chain, model.MatchEvidence{
				Requirement:  fmt.Sprintf("岗位要求 %s 相关经验", req),
				ResumePoint:  getResumeEvidenceForSkill(resume, req),
				Judgement:    "missing",
				ResumeModule: "skills",
			})
		}
	}

	// If no skills were extracted, generate evidence from description/requirements directly
	if len(chain) == 0 && job.Requirements != "" {
		// Check for key concepts in requirements text
		reqLower := strings.ToLower(job.Requirements)
		keyConcepts := []string{"本科", "硕士", "计算机", "编程", "开发", "算法", "数据结构", "沟通", "团队"}
		for _, concept := range keyConcepts {
			if strings.Contains(reqLower, concept) {
				judgement := "missing"
				resumePoint := "简历中未体现相关内容"
				if concept == "计算机" || concept == "编程" || concept == "开发" {
					if resume.TargetRole != "" {
						judgement = "partial_match"
						resumePoint = fmt.Sprintf("目标岗位为%s，可能与要求相关", resume.TargetRole)
					}
				}
				if concept == "沟通" || concept == "团队" {
					if len(resume.ExperienceJSON) > 0 {
						judgement = "partial_match"
						resumePoint = "简历中包含实习或项目经历，可能涉及团队协作"
					}
				}
				chain = append(chain, model.MatchEvidence{
					Requirement: fmt.Sprintf("岗位要求具备%s能力", concept),
					ResumePoint: resumePoint,
					Judgement:   judgement,
				})
			}
		}
	}

	// Check city match
	targetCity := strings.TrimSpace(resume.TargetCity)
	if targetCity != "" {
		judgement := "missing"
		resumePoint := "简历中未体现该城市的工作经历"
		if strings.Contains(strings.ToLower(job.City), strings.ToLower(targetCity)) {
			judgement = "strong_match"
			resumePoint = fmt.Sprintf("目标城市为%s，与岗位所在城市一致", targetCity)
		}
		chain = append(chain, model.MatchEvidence{
			Requirement: fmt.Sprintf("工作地点为%s", job.City),
			ResumePoint: resumePoint,
			Judgement:   judgement,
		})
	}

	// Check role/direction match
	targetRole := strings.TrimSpace(resume.TargetRole)
	if targetRole != "" {
		jobText := strings.ToLower(job.Category + " " + job.Title)
		if containsAny(strings.ToLower(targetRole), jobText) {
			chain = append(chain, model.MatchEvidence{
				Requirement:  fmt.Sprintf("岗位方向为%s", job.Category),
				ResumePoint:  fmt.Sprintf("目标岗位为%s，方向一致", targetRole),
				Judgement:    "strong_match",
				ResumeModule: "projects",
			})
		} else {
			chain = append(chain, model.MatchEvidence{
				Requirement:  fmt.Sprintf("岗位方向为%s", job.Category),
				ResumePoint:  fmt.Sprintf("目标岗位为%s，方向不完全一致", targetRole),
				Judgement:    "partial_match",
				ResumeModule: "projects",
			})
		}
	}

	// Ensure at least 3 evidence entries
	if len(chain) < 3 {
		chain = append(chain, model.MatchEvidence{
			Requirement:  "具备相关项目或实习经历",
			ResumePoint:  getExperienceSummary(resume),
			Judgement:    getExperienceJudgement(resume),
			ResumeModule: "internships",
		})
	}

	return chain
}

// extractKeywordsFromJD extracts common tech keywords from job description and requirements text.
func extractKeywordsFromJD(description, requirements string) []string {
	commonKeywords := []string{
		"go", "python", "java", "c++", "javascript", "typescript",
		"mysql", "redis", "mongodb", "postgresql", "elasticsearch",
		"docker", "kubernetes", "微服务", "分布式", "高并发", "高可用",
		"linux", "git", "grpc", "restful", "api",
		"算法", "数据结构", "机器学习", "深度学习",
		"hadoop", "spark", "flink", "kafka",
		"云计算", "网络", "安全", "数据库", "缓存", "消息队列",
	}

	text := strings.ToLower(description + " " + requirements)
	var found []string
	for _, kw := range commonKeywords {
		if strings.Contains(text, kw) {
			found = append(found, kw)
		}
	}
	return found
}

// getResumeEvidenceForSkill returns evidence from the resume for a specific skill.
func getResumeEvidenceForSkill(resume *model.ResumeVersion, skill string) string {
	skillLower := strings.ToLower(skill)

	// Check in skills list
	resumeSkills := parseJSONStringArray(resume.SkillsJSON)
	for _, s := range resumeSkills {
		if strings.ToLower(s) == skillLower {
			return fmt.Sprintf("技能列表中包含%s", s)
		}
	}

	// Check in raw text
	if resume.RawText != "" && strings.Contains(strings.ToLower(resume.RawText), skillLower) {
		return fmt.Sprintf("简历内容中提及了%s相关内容", skill)
	}

	// Check in project JSON
	if len(resume.ProjectJSON) > 0 {
		var projects []map[string]interface{}
		if json.Unmarshal(resume.ProjectJSON, &projects) == nil {
			for _, p := range projects {
				if desc, ok := p["description"].(string); ok {
					if strings.Contains(strings.ToLower(desc), skillLower) {
						return fmt.Sprintf("项目经历中涉及%s", skill)
					}
				}
			}
		}
	}

	// Check in experience JSON
	if len(resume.ExperienceJSON) > 0 {
		var experiences []map[string]interface{}
		if json.Unmarshal(resume.ExperienceJSON, &experiences) == nil {
			for _, e := range experiences {
				if desc, ok := e["description"].(string); ok {
					if strings.Contains(strings.ToLower(desc), skillLower) {
						return fmt.Sprintf("实习/工作经历中涉及%s", skill)
					}
				}
			}
		}
	}

	return "简历中未体现该技能相关经验"
}

// getExperienceSummary returns a summary of the resume's experience.
func getExperienceSummary(resume *model.ResumeVersion) string {
	if len(resume.ExperienceJSON) > 0 {
		var experiences []map[string]interface{}
		if json.Unmarshal(resume.ExperienceJSON, &experiences) == nil && len(experiences) > 0 {
			return fmt.Sprintf("简历中包含%d段实习/工作经历", len(experiences))
		}
	}
	if len(resume.ProjectJSON) > 0 {
		var projects []map[string]interface{}
		if json.Unmarshal(resume.ProjectJSON, &projects) == nil && len(projects) > 0 {
			return fmt.Sprintf("简历中包含%d段项目经历", len(projects))
		}
	}
	return "简历中未体现相关项目或实习经历"
}

// getExperienceJudgement returns the judgement for the experience evidence.
func getExperienceJudgement(resume *model.ResumeVersion) string {
	if len(resume.ExperienceJSON) > 0 {
		var experiences []map[string]interface{}
		if json.Unmarshal(resume.ExperienceJSON, &experiences) == nil && len(experiences) > 0 {
			return "partial_match"
		}
	}
	return "missing"
}

// generateMatchHighlights creates match highlight strings.
func generateMatchHighlights(resume *model.ResumeVersion, job *model.Job) []string {
	var highlights []string

	// City match
	if strings.TrimSpace(resume.TargetCity) != "" && job.City == strings.TrimSpace(resume.TargetCity) {
		highlights = append(highlights, fmt.Sprintf("目标城市与岗位所在城市一致（%s）", job.City))
	}

	// Role match
	targetRole := strings.TrimSpace(resume.TargetRole)
	if targetRole != "" {
		jobText := strings.ToLower(job.Category + " " + job.Title)
		if containsAny(strings.ToLower(targetRole), jobText) {
			highlights = append(highlights, fmt.Sprintf("目标岗位与岗位方向高度一致（%s）", job.Category))
		}
	}

	// Skill overlap
	resumeSkills := parseJSONStringArray(resume.SkillsJSON)
	jobSkills := parseJSONStringArray(job.SkillKeywordsJSON)
	jobSkillSet := toLowerSet(jobSkills)
	var matchedSkills []string
	for _, s := range resumeSkills {
		if jobSkillSet[strings.ToLower(s)] {
			matchedSkills = append(matchedSkills, s)
		}
	}
	if len(matchedSkills) > 0 {
		highlights = append(highlights, fmt.Sprintf("技能关键词与岗位要求存在交集（%s）", strings.Join(matchedSkills[:min(3, len(matchedSkills))], "、")))
	}

	// Experience mention
	if len(resume.ExperienceJSON) > 0 {
		var experiences []map[string]interface{}
		if json.Unmarshal(resume.ExperienceJSON, &experiences) == nil && len(experiences) > 0 {
			highlights = append(highlights, "具备相关实习或工作经验")
		}
	}

	if len(highlights) == 0 {
		highlights = append(highlights, "岗位方向与求职目标部分相关")
	}

	return highlights
}

// generateRiskPoints creates risk point strings.
func generateRiskPoints(resume *model.ResumeVersion, job *model.Job) []string {
	var risks []string

	resumeSkills := parseJSONStringArray(resume.SkillsJSON)
	jobSkills := parseJSONStringArray(job.SkillKeywordsJSON)
	resumeSkillSet := toLowerSet(resumeSkills)

	// Find missing skills
	var missing []string
	for _, s := range jobSkills {
		if !resumeSkillSet[strings.ToLower(s)] {
			missing = append(missing, s)
		}
	}
	if len(missing) > 0 {
		top := missing[:min(2, len(missing))]
		risks = append(risks, fmt.Sprintf("%s 经验表达仍可进一步强化", strings.Join(top, " 或 ")))
	}

	// City mismatch
	if strings.TrimSpace(resume.TargetCity) != "" && job.City != strings.TrimSpace(resume.TargetCity) {
		risks = append(risks, fmt.Sprintf("工作地点为%s，与目标城市不一致", job.City))
	}

	// Direction mismatch
	targetRole := strings.TrimSpace(resume.TargetRole)
	if targetRole != "" {
		jobText := strings.ToLower(job.Category + " " + job.Title)
		if !containsAny(strings.ToLower(targetRole), jobText) {
			risks = append(risks, fmt.Sprintf("目标岗位与%s方向不完全一致", job.Category))
		}
	}

	if len(risks) == 0 {
		risks = append(risks, "项目经历偏校园级别，企业级项目深度可进一步提升")
	}

	return risks
}

// generateOptimizationSuggestions creates P0/P1 optimization suggestions.
func generateOptimizationSuggestions(resume *model.ResumeVersion, job *model.Job) []model.MatchSuggestion {
	var suggestions []model.MatchSuggestion

	resumeSkills := parseJSONStringArray(resume.SkillsJSON)
	jobSkills := parseJSONStringArray(job.SkillKeywordsJSON)

	// If skill_keywords is empty, extract from description/requirements
	if len(jobSkills) == 0 && (job.Description != "" || job.Requirements != "") {
		jobSkills = extractKeywordsFromJD(job.Description, job.Requirements)
	}

	resumeSkillSet := toLowerSet(resumeSkills)

	// P0: Missing critical skills
	var missingCritical []string
	for _, s := range jobSkills {
		if !resumeSkillSet[strings.ToLower(s)] {
			missingCritical = append(missingCritical, s)
		}
	}
	if len(missingCritical) > 0 {
		topMissing := missingCritical[:min(2, len(missingCritical))]
		suggestions = append(suggestions, model.MatchSuggestion{
			Priority: "P0",
			Title:    fmt.Sprintf("强化%s相关经验", strings.Join(topMissing, "、")),
			Reason:   fmt.Sprintf("岗位要求中明确提及%s，当前简历中相关证据不足", strings.Join(topMissing, "、")),
			Action:   fmt.Sprintf("如果真实参与过相关内容，可在项目经历中补充%s的使用场景和具体成果", strings.Join(topMissing, "、")),
		})
	}

	// P0: Missing teamwork/communication
	if job.Requirements != "" && (strings.Contains(strings.ToLower(job.Requirements), "沟通") || strings.Contains(strings.ToLower(job.Requirements), "团队")) {
		suggestions = append(suggestions, model.MatchSuggestion{
			Priority: "P0",
			Title:    "补充团队协作和沟通能力的具体实例",
			Reason:   "岗位明确要求沟通能力和团队协作精神",
			Action:   "在实习或项目经历中补充跨团队协作、技术分享、代码评审等具体实例",
		})
	}

	// P1: Quantify project results
	if len(resume.ProjectJSON) > 0 {
		suggestions = append(suggestions, model.MatchSuggestion{
			Priority: "P1",
			Title:    "在项目描述中增加更多量化指标",
			Reason:   "当前部分项目描述缺乏具体技术指标，量化数据能提升说服力",
			Action:   "补充接口数量、性能指标（QPS、响应时间）、数据库表设计职责等量化成果",
		})
	}

	// P1: City/region if mismatch
	if strings.TrimSpace(resume.TargetCity) != "" && job.City != strings.TrimSpace(resume.TargetCity) {
		suggestions = append(suggestions, model.MatchSuggestion{
			Priority: "P1",
			Title:    fmt.Sprintf("确认是否接受%s工作地点", job.City),
			Reason:   fmt.Sprintf("目标城市为%s，岗位位于%s", resume.TargetCity, job.City),
			Action:   "如果可接受异地，可在自我评价中补充对工作地点的灵活性说明",
		})
	}

	// Ensure at least P0 and P1
	if len(suggestions) == 0 {
		suggestions = append(suggestions, model.MatchSuggestion{
			Priority: "P0",
			Title:    "强化项目经历中的后端职责",
			Reason:   "岗位强调后端接口设计和数据库能力，当前简历已有基础但证据仍可量化",
			Action:   "在项目经历中补充接口数量、性能指标或数据库表设计职责",
		})
	}
	if len(suggestions) == 1 {
		suggestions = append(suggestions, model.MatchSuggestion{
			Priority: "P1",
			Title:    "补充缓存与并发场景表达",
			Reason:   "岗位可能涉及 Redis、并发或高可用场景",
			Action:   "如果真实参与过相关内容，可补充缓存穿透、限流、幂等或并发处理经验",
		})
	}

	return suggestions
}

// generateNextSteps creates next step suggestions.
func generateNextSteps(suggestions []model.MatchSuggestion) []string {
	var steps []string

	hasP0 := false
	for _, s := range suggestions {
		if s.Priority == "P0" {
			steps = append(steps, fmt.Sprintf("根据 P0 建议：%s", s.Action))
			hasP0 = true
			break
		}
	}

	if hasP0 {
		steps = append(steps, "优化简历后添加新版本")
	}

	steps = append(steps, "新版本创建后可再次进行匹配分析，查看匹配度变化")
	steps = append(steps, "确认岗位详情中的核心技术要求，逐条对照简历内容")

	return steps
}

// determineMatchApplyReadiness determines apply_readiness and optimization_stop_signal
// for mock match analysis based on score and data quality.
func determineMatchApplyReadiness(score int, resume *model.ResumeVersion, job *model.Job) (string, string) {
	if job.DataQuality == "summary_only" {
		return "insufficient_information", "needs_more_evidence"
	}
	if score >= 85 {
		return "ready_to_apply", "stop_and_apply"
	}
	if score >= 60 {
		return "quick_fix_before_apply", "fix_top_3_then_apply"
	}
	return "not_ready_for_this_role", "needs_more_evidence"
}

// generateMatchTopFixItems generates up to 3 quick fix items for mock match analysis.
func generateMatchTopFixItems(resume *model.ResumeVersion, job *model.Job) []model.TopFixItem {
	var items []model.TopFixItem

	resumeSkills := parseJSONStringArray(resume.SkillsJSON)
	jobSkills := parseJSONStringArray(job.SkillKeywordsJSON)
	resumeSkillSet := toLowerSet(resumeSkills)

	// Find missing skills
	var missing []string
	for _, s := range jobSkills {
		if !resumeSkillSet[strings.ToLower(s)] {
			missing = append(missing, s)
		}
	}

	if len(missing) > 0 {
		top := missing[:min(2, len(missing))]
		items = append(items, model.TopFixItem{
			Issue:            fmt.Sprintf("简历中未体现 %s 相关经验", strings.Join(top, "、")),
			WhyItMatters:     "岗位 JD 明确提及这些技术，是初筛关键词。",
			Action:           fmt.Sprintf("如果真实使用过 %s，在项目或实习描述中补充使用场景。", strings.Join(top, "、")),
			ExampleReference: fmt.Sprintf("示例参考：使用 %s 完成数据持久化，通过连接池优化查询性能。", strings.Join(top, "、")),
			SelfCheckQuestion: "请确认你确实在项目中使用过这些技术，而不是仅了解概念。",
			EstimatedEffort:  "within_30_min",
			ResumeModule:     "skills",
		})
	}

	// Check project experience
	if len(resume.ProjectJSON) == 0 {
		items = append(items, model.TopFixItem{
			Issue:            "缺少项目经历描述",
			WhyItMatters:     "岗位 JD 要求开发经验，当前简历没有项目经历会影响初筛通过率。",
			Action:           "补充你参与过的课程项目、个人项目或实验室项目。",
			ExampleReference: "示例参考：基于 Gin 框架实现博客系统，包含用户注册登录、文章 CRUD 和评论功能。",
			SelfCheckQuestion: "请确认项目确实是你本人参与完成的。",
			EstimatedEffort:  "within_60_min",
			ResumeModule:     "projects",
		})
	} else {
		items = append(items, model.TopFixItem{
			Issue:            "项目职责描述不够清晰",
			WhyItMatters:     "岗位 JD 强调后台接口开发和工程实践，当前表达难以证明个人贡献。",
			Action:           "补充你负责的模块、使用的技术栈和最终产出。",
			ExampleReference: "示例参考：基于 GoFrame 完成后台管理系统用户、角色、权限相关接口开发，并与前端完成联调。",
			SelfCheckQuestion: "请确认你是否确实负责过这些模块，而不是团队整体成果。",
			EstimatedEffort:  "within_30_min",
			ResumeModule:     "projects",
		})
	}

	// Cap at 3 items
	if len(items) > 3 {
		items = items[:3]
	}

	return items
}
