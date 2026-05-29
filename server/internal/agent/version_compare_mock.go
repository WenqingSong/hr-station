package agent

import (
	"encoding/json"
	"fmt"
	"strings"

	"offer-catcher/server/internal/model"
)

// VersionCompareMockAgent generates mock version comparison using simple rule-based analysis.
// It does NOT call any real LLM or read LLM_API_KEY.
type VersionCompareMockAgent struct{}

// NewVersionCompareMockAgent creates a new VersionCompareMockAgent.
func NewVersionCompareMockAgent() *VersionCompareMockAgent {
	return &VersionCompareMockAgent{}
}

// Compare generates a mock version comparison based on two resume versions.
func (a *VersionCompareMockAgent) Compare(from *model.ResumeVersion, to *model.ResumeVersion) (*model.VersionCompareAnalysis, error) {
	scoreBefore := calculateResumeScore(from)
	scoreAfter := calculateResumeScore(to)
	scoreChange := scoreAfter - scoreBefore

	fieldChanges := generateFieldChanges(from, to)
	riskChanges := generateRiskChanges(from, to)
	suggestionCompletion := generateSuggestionCompletion(from, to)
	keyImprovements := generateKeyImprovements(from, to)
	remainingRisks := generateRemainingRisks(to)
	summary := generateSummary(from, to, scoreBefore, scoreAfter)

	notice := "版本对比仅用于呈现表达变化，请只补充真实经历，不要虚构项目或能力。"

	return &model.VersionCompareAnalysis{
		FromVersionNo:        from.VersionNo,
		ToVersionNo:          to.VersionNo,
		ScoreBefore:          scoreBefore,
		ScoreAfter:           scoreAfter,
		ScoreChange:          scoreChange,
		FieldChanges:         fieldChanges,
		RiskChanges:          riskChanges,
		SuggestionCompletion: suggestionCompletion,
		KeyImprovements:      keyImprovements,
		RemainingRisks:       remainingRisks,
		Summary:              summary,
		AuthenticityNotice:   notice,
	}, nil
}

// calculateResumeScore computes a mock score (0-100) for a resume version.
// This is a simple heuristic based on completeness and detail level.
func calculateResumeScore(resume *model.ResumeVersion) int {
	score := 40 // base score

	// Target role specified
	if strings.TrimSpace(resume.TargetRole) != "" {
		score += 8
	}

	// Target city specified
	if strings.TrimSpace(resume.TargetCity) != "" {
		score += 5
	}

	// Skills count
	skills := parseJSONStringArray(resume.SkillsJSON)
	score += min(len(skills), 6) * 3 // max 18

	// Key backend skills bonus
	backendKeywords := []string{"go", "mysql", "redis", "docker", "kubernetes", "云", "分布式", "微服务"}
	skillSet := toLowerSet(skills)
	for _, kw := range backendKeywords {
		if skillSet[kw] {
			score += 2
		}
	}

	// Project detail
	if len(resume.ProjectJSON) > 0 {
		var projects []map[string]interface{}
		if json.Unmarshal(resume.ProjectJSON, &projects) == nil {
			for _, p := range projects {
				if desc, ok := p["description"].(string); ok && len(desc) > 20 {
					score += 3
				}
			}
		}
	}

	// Experience detail
	if len(resume.ExperienceJSON) > 0 {
		var experiences []map[string]interface{}
		if json.Unmarshal(resume.ExperienceJSON, &experiences) == nil {
			for _, e := range experiences {
				if desc, ok := e["description"].(string); ok && len(desc) > 10 {
					score += 3
				}
			}
		}
	}

	// Self evaluation
	if len(strings.TrimSpace(resume.SelfEvaluation)) > 10 {
		score += 4
	}

	// Raw text length bonus
	if len(resume.RawText) > 100 {
		score += 3
	}

	// Cap at 100
	if score > 100 {
		score = 100
	}

	return score
}

// generateFieldChanges creates field change entries between two versions.
func generateFieldChanges(from *model.ResumeVersion, to *model.ResumeVersion) []model.FieldChange {
	var changes []model.FieldChange

	// Target role change
	fromRole := strings.TrimSpace(from.TargetRole)
	toRole := strings.TrimSpace(to.TargetRole)
	if fromRole != toRole {
		changeType := "added"
		if fromRole != "" && toRole != "" {
			changeType = "modified"
		} else if toRole == "" {
			changeType = "removed"
		}
		before := formatEmpty(fromRole)
		after := formatEmpty(toRole)
		impact := "目标岗位表达"
		if toRole != "" && strings.Contains(strings.ToLower(toRole), "后端") {
			impact = "目标岗位更聚焦后端开发方向"
		}
		changes = append(changes, model.FieldChange{
			Field:      "target_role",
			Before:     before,
			After:      after,
			ChangeType: changeType,
			Impact:     impact,
		})
	}

	// Skills change
	fromSkills := parseJSONStringArray(from.SkillsJSON)
	toSkills := parseJSONStringArray(to.SkillsJSON)
	fromSkillSet := toLowerSet(fromSkills)
	toSkillSet := toLowerSet(toSkills)

	var addedSkills, removedSkills []string
	for _, s := range toSkills {
		if !fromSkillSet[strings.ToLower(s)] {
			addedSkills = append(addedSkills, s)
		}
	}
	for _, s := range fromSkills {
		if !toSkillSet[strings.ToLower(s)] {
			removedSkills = append(removedSkills, s)
		}
	}

	if len(addedSkills) > 0 || len(removedSkills) > 0 {
		before := "基础技能列表"
		if len(fromSkills) > 0 {
			before = fmt.Sprintf("%d 项技能（%s）", len(fromSkills), joinLimited(fromSkills, 3))
		}
		after := "基础技能列表"
		if len(toSkills) > 0 {
			after = fmt.Sprintf("%d 项技能（%s）", len(toSkills), joinLimited(toSkills, 3))
		}

		changeType := "modified"
		if len(fromSkills) == 0 && len(toSkills) > 0 {
			changeType = "added"
		}

		impact := "技能关键词更贴合后端开发方向"
		if len(addedSkills) > 0 {
			impact = fmt.Sprintf("新增了 %s 等后端核心技能", joinLimited(addedSkills, 3))
		}

		changes = append(changes, model.FieldChange{
			Field:      "skills",
			Before:     before,
			After:      after,
			ChangeType: changeType,
			Impact:     impact,
		})
	}

	// Project change
	fromProjDesc := getProjectSummary(from)
	toProjDesc := getProjectSummary(to)
	if fromProjDesc != toProjDesc {
		changeType := "modified"
		if fromProjDesc == "无项目经历" && toProjDesc != "无项目经历" {
			changeType = "added"
		}

		impact := "项目经历与后端岗位要求的关联更清晰"
		changes = append(changes, model.FieldChange{
			Field:      "project",
			Before:     fromProjDesc,
			After:      toProjDesc,
			ChangeType: changeType,
			Impact:     impact,
		})
	}

	// Self evaluation change
	fromEval := strings.TrimSpace(from.SelfEvaluation)
	toEval := strings.TrimSpace(to.SelfEvaluation)
	if fromEval != toEval {
		changeType := "modified"
		if fromEval == "" && toEval != "" {
			changeType = "added"
		}

		impact := "自我评价更具体"
		if toEval != "" && len(toEval) > len(fromEval) {
			impact = "自我评价更充实，突出后端开发能力"
		}

		changes = append(changes, model.FieldChange{
			Field:      "self_evaluation",
			Before:     formatEmpty(fromEval),
			After:      formatEmpty(toEval),
			ChangeType: changeType,
			Impact:     impact,
		})
	}

	// Ensure at least one change entry
	if len(changes) == 0 {
		changes = append(changes, model.FieldChange{
			Field:      "overall",
			Before:     "简历内容无显著差异",
			After:      "简历内容无显著差异",
			ChangeType: "unchanged",
			Impact:     "两个版本之间未检测到明显变化",
		})
	}

	return changes
}

// generateRiskChanges creates risk change entries between two versions.
func generateRiskChanges(from *model.ResumeVersion, to *model.ResumeVersion) []model.RiskChange {
	var risks []model.RiskChange

	fromSkills := parseJSONStringArray(from.SkillsJSON)
	toSkills := parseJSONStringArray(to.SkillsJSON)
	fromSkillSet := toLowerSet(fromSkills)
	toSkillSet := toLowerSet(toSkills)

	// Backend experience evidence
	fromHasBackendExp := hasBackendEvidence(from)
	toHasBackendExp := hasBackendEvidence(to)

	if !fromHasBackendExp && toHasBackendExp {
		risks = append(risks, model.RiskChange{
			Risk:   "后端项目证据不足",
			Before: "较明显",
			After:  "有所降低",
			Trend:  "reduced",
		})
	} else if fromHasBackendExp && toHasBackendExp {
		risks = append(risks, model.RiskChange{
			Risk:   "后端项目证据不足",
			Before: "有所降低",
			After:  "有所降低",
			Trend:  "stable",
		})
	} else {
		risks = append(risks, model.RiskChange{
			Risk:   "后端项目证据不足",
			Before: "较明显",
			After:  "较明显",
			Trend:  "unchanged",
		})
	}

	// Missing key skills risk
	backendKeywords := []string{"redis", "kubernetes", "docker", "分布式", "微服务"}
	fromMissing := countMissing(fromSkillSet, backendKeywords)
	toMissing := countMissing(toSkillSet, backendKeywords)

	if fromMissing > toMissing {
		risks = append(risks, model.RiskChange{
			Risk:   "核心后端技术覆盖不全",
			Before: "较明显",
			After:  "有所降低",
			Trend:  "reduced",
		})
	} else if toMissing > 0 {
		risks = append(risks, model.RiskChange{
			Risk:   "核心后端技术覆盖不全",
			Before: "较明显",
			After:  "较明显",
			Trend:  "unchanged",
		})
	}

	// Quantification risk
	fromHasQuant := hasQuantifiableData(from)
	toHasQuant := hasQuantifiableData(to)
	if !fromHasQuant && toHasQuant {
		risks = append(risks, model.RiskChange{
			Risk:   "量化指标不足",
			Before: "较明显",
			After:  "有所降低",
			Trend:  "reduced",
		})
	} else if !toHasQuant {
		risks = append(risks, model.RiskChange{
			Risk:   "量化指标不足",
			Before: "较明显",
			After:  "较明显",
			Trend:  "unchanged",
		})
	}

	if len(risks) == 0 {
		risks = append(risks, model.RiskChange{
			Risk:   "整体风险",
			Before: "中等",
			After:  "较低",
			Trend:  "reduced",
		})
	}

	return risks
}

// generateSuggestionCompletion creates suggestion completion entries.
func generateSuggestionCompletion(from *model.ResumeVersion, to *model.ResumeVersion) []model.SuggestionCompletion {
	var completions []model.SuggestionCompletion

	fromSkills := parseJSONStringArray(from.SkillsJSON)
	toSkills := parseJSONStringArray(to.SkillsJSON)
	fromSkillSet := toLowerSet(fromSkills)
	toSkillSet := toLowerSet(toSkills)

	// Suggestion: Add backend skills
	backendKws := []string{"go", "mysql", "docker", "redis", "kubernetes"}
	fromHasBackend := false
	toHasBackend := false
	for _, kw := range backendKws {
		if fromSkillSet[kw] {
			fromHasBackend = true
		}
		if toSkillSet[kw] {
			toHasBackend = true
		}
	}

	if !fromHasBackend && toHasBackend {
		completions = append(completions, model.SuggestionCompletion{
			Suggestion: "补充项目中的后端职责和技术栈",
			Status:     "completed",
			Evidence:   "新版本增加了 Go、MySQL、Docker 等关键词和项目职责描述",
		})
	} else if fromHasBackend && toHasBackend {
		completions = append(completions, model.SuggestionCompletion{
			Suggestion: "补充项目中的后端职责和技术栈",
			Status:     "partially_completed",
			Evidence:   "后端技术栈已覆盖，但仍可进一步细化项目职责",
		})
	} else {
		completions = append(completions, model.SuggestionCompletion{
			Suggestion: "补充项目中的后端职责和技术栈",
			Status:     "not_started",
			Evidence:   "两个版本均未体现后端技术栈",
		})
	}

	// Suggestion: Quantify project results
	fromProjDetail := getProjectDetailLevel(from)
	toProjDetail := getProjectDetailLevel(to)

	if toProjDetail > fromProjDetail {
		completions = append(completions, model.SuggestionCompletion{
			Suggestion: "在项目描述中增加量化指标",
			Status:     "partially_completed",
			Evidence:   "新版本项目描述有所充实，但量化指标仍可进一步补充",
		})
	} else {
		completions = append(completions, model.SuggestionCompletion{
			Suggestion: "在项目描述中增加量化指标",
			Status:     "not_started",
			Evidence:   "项目描述中尚未补充量化指标",
		})
	}

	// Suggestion: Add caching/concurrency experience
	fromHasCache := hasKeywordInResume(from, []string{"redis", "缓存", "并发", "限流"})
	toHasCache := hasKeywordInResume(to, []string{"redis", "缓存", "并发", "限流"})

	if !fromHasCache && toHasCache {
		completions = append(completions, model.SuggestionCompletion{
			Suggestion: "补充缓存与并发场景表达",
			Status:     "completed",
			Evidence:   "新版本中补充了 Redis、缓存或并发相关内容",
		})
	} else if fromHasCache && toHasCache {
		completions = append(completions, model.SuggestionCompletion{
			Suggestion: "补充缓存与并发场景表达",
			Status:     "partially_completed",
			Evidence:   "已有缓存/并发相关内容，可进一步补充具体场景",
		})
	} else {
		completions = append(completions, model.SuggestionCompletion{
			Suggestion: "补充缓存与并发场景表达",
			Status:     "not_started",
			Evidence:   "两个版本均未体现缓存或并发经验",
		})
	}

	return completions
}

// generateKeyImprovements creates key improvement strings.
func generateKeyImprovements(from *model.ResumeVersion, to *model.ResumeVersion) []string {
	var improvements []string

	fromSkills := parseJSONStringArray(from.SkillsJSON)
	toSkills := parseJSONStringArray(to.SkillsJSON)
	fromSkillSet := toLowerSet(fromSkills)

	// Check for new skills
	var addedSkills []string
	for _, s := range toSkills {
		if !fromSkillSet[strings.ToLower(s)] {
			addedSkills = append(addedSkills, s)
		}
	}
	if len(addedSkills) > 0 {
		improvements = append(improvements, fmt.Sprintf("技能关键词更贴合后端开发方向（新增 %s）", joinLimited(addedSkills, 3)))
	}

	// Check for role improvement
	fromRole := strings.TrimSpace(from.TargetRole)
	toRole := strings.TrimSpace(to.TargetRole)
	if fromRole != toRole && toRole != "" {
		improvements = append(improvements, "目标岗位表达更聚焦")
	}

	// Check for project improvement
	fromProjDetail := getProjectDetailLevel(from)
	toProjDetail := getProjectDetailLevel(to)
	if toProjDetail > fromProjDetail {
		improvements = append(improvements, "项目经历更适合用于后续岗位匹配分析")
	}

	// Check for self evaluation improvement
	if len(strings.TrimSpace(to.SelfEvaluation)) > len(strings.TrimSpace(from.SelfEvaluation)) {
		improvements = append(improvements, "自我评价更具体充实")
	}

	if len(improvements) == 0 {
		improvements = append(improvements, "简历版本间未检测到显著改进")
	}

	return improvements
}

// generateRemainingRisks creates remaining risk strings.
func generateRemainingRisks(to *model.ResumeVersion) []string {
	var risks []string

	skills := parseJSONStringArray(to.SkillsJSON)
	skillSet := toLowerSet(skills)

	// Check for missing key backend skills
	missingKws := []string{"redis", "kubernetes", "分布式", "微服务"}
	var missing []string
	for _, kw := range missingKws {
		if !skillSet[kw] {
			missing = append(missing, kw)
		}
	}
	if len(missing) > 0 {
		risks = append(risks, fmt.Sprintf("%s 相关证据仍偏少", joinLimited(missing, 2)))
	}

	// Check for quantifiable data
	if !hasQuantifiableData(to) {
		risks = append(risks, "量化指标仍可进一步补充")
	}

	// Check for project detail
	if getProjectDetailLevel(to) < 2 {
		risks = append(risks, "项目经历仍可进一步充实细节")
	}

	if len(risks) == 0 {
		risks = append(risks, "当前简历已较完善，后续可根据目标岗位微调")
	}

	return risks
}

// generateSummary creates a comparison summary.
func generateSummary(from *model.ResumeVersion, to *model.ResumeVersion, scoreBefore, scoreAfter int) string {
	parts := []string{}

	if scoreAfter > scoreBefore {
		parts = append(parts, fmt.Sprintf("V%d 相比 V%d 在简历完整度上有所提升", to.VersionNo, from.VersionNo))
	} else if scoreAfter == scoreBefore {
		parts = append(parts, fmt.Sprintf("V%d 与 V%d 在简历完整度上基本一致", to.VersionNo, from.VersionNo))
	} else {
		parts = append(parts, fmt.Sprintf("V%d 相比 V%d 在简历完整度上有所下降", to.VersionNo, from.VersionNo))
	}

	fromSkills := parseJSONStringArray(from.SkillsJSON)
	toSkills := parseJSONStringArray(to.SkillsJSON)
	if len(toSkills) > len(fromSkills) {
		parts = append(parts, "技能关键词更丰富")
	}

	fromRole := strings.TrimSpace(from.TargetRole)
	toRole := strings.TrimSpace(to.TargetRole)
	if fromRole != toRole && toRole != "" {
		parts = append(parts, "目标岗位更聚焦")
	}

	if len(parts) > 1 {
		return strings.Join(parts, "，") + "，适合继续用于后端开发岗位匹配分析。"
	}

	return parts[0] + "。"
}

// Helper functions

func formatEmpty(s string) string {
	if strings.TrimSpace(s) == "" {
		return "未填写"
	}
	return s
}

func joinLimited(items []string, limit int) string {
	if len(items) <= limit {
		return strings.Join(items, "、")
	}
	return strings.Join(items[:limit], "、") + " 等"
}

func getProjectSummary(resume *model.ResumeVersion) string {
	if len(resume.ProjectJSON) == 0 {
		return "无项目经历"
	}
	var projects []map[string]interface{}
	if json.Unmarshal(resume.ProjectJSON, &projects) != nil || len(projects) == 0 {
		return "无项目经历"
	}

	totalLen := 0
	for _, p := range projects {
		if desc, ok := p["description"].(string); ok {
			totalLen += len(desc)
		}
	}

	if totalLen > 100 {
		return fmt.Sprintf("项目描述较充实（%d 个项目，描述总长 %d 字）", len(projects), totalLen)
	} else if totalLen > 30 {
		return fmt.Sprintf("项目描述偏概括（%d 个项目，描述总长 %d 字）", len(projects), totalLen)
	}
	return fmt.Sprintf("项目描述较简略（%d 个项目）", len(projects))
}

func getProjectDetailLevel(resume *model.ResumeVersion) int {
	if len(resume.ProjectJSON) == 0 {
		return 0
	}
	var projects []map[string]interface{}
	if json.Unmarshal(resume.ProjectJSON, &projects) != nil {
		return 0
	}

	level := 0
	for _, p := range projects {
		if desc, ok := p["description"].(string); ok {
			if len(desc) > 50 {
				level += 2
			} else if len(desc) > 10 {
				level += 1
			}
		}
	}
	return level
}

func hasBackendEvidence(resume *model.ResumeVersion) bool {
	skills := parseJSONStringArray(resume.SkillsJSON)
	skillSet := toLowerSet(skills)
	backendKws := []string{"go", "mysql", "docker", "redis", "kubernetes"}
	for _, kw := range backendKws {
		if skillSet[kw] {
			return true
		}
	}

	// Check project descriptions
	if len(resume.ProjectJSON) > 0 {
		var projects []map[string]interface{}
		if json.Unmarshal(resume.ProjectJSON, &projects) == nil {
			for _, p := range projects {
				if desc, ok := p["description"].(string); ok {
					descLower := strings.ToLower(desc)
					if strings.Contains(descLower, "接口") || strings.Contains(descLower, "数据库") ||
						strings.Contains(descLower, "后端") || strings.Contains(descLower, "服务") {
						return true
					}
				}
			}
		}
	}

	return false
}

func countMissing(skillSet map[string]bool, keywords []string) int {
	count := 0
	for _, kw := range keywords {
		if !skillSet[kw] {
			count++
		}
	}
	return count
}

func hasQuantifiableData(resume *model.ResumeVersion) bool {
	quantKeywords := []string{"qps", "tps", "响应", "性能", "并发", "接口数", "表设计", "优化", "提升", "降低", "%"}
	texts := []string{resume.SelfEvaluation, resume.RawText}

	if len(resume.ProjectJSON) > 0 {
		var projects []map[string]interface{}
		if json.Unmarshal(resume.ProjectJSON, &projects) == nil {
			for _, p := range projects {
				if desc, ok := p["description"].(string); ok {
					texts = append(texts, desc)
				}
			}
		}
	}

	if len(resume.ExperienceJSON) > 0 {
		var experiences []map[string]interface{}
		if json.Unmarshal(resume.ExperienceJSON, &experiences) == nil {
			for _, e := range experiences {
				if desc, ok := e["description"].(string); ok {
					texts = append(texts, desc)
				}
			}
		}
	}

	for _, text := range texts {
		textLower := strings.ToLower(text)
		for _, kw := range quantKeywords {
			if strings.Contains(textLower, kw) {
				return true
			}
		}
	}

	return false
}

func hasKeywordInResume(resume *model.ResumeVersion, keywords []string) bool {
	texts := []string{resume.SelfEvaluation, resume.RawText}
	skills := parseJSONStringArray(resume.SkillsJSON)
	texts = append(texts, skills...)

	if len(resume.ProjectJSON) > 0 {
		var projects []map[string]interface{}
		if json.Unmarshal(resume.ProjectJSON, &projects) == nil {
			for _, p := range projects {
				if desc, ok := p["description"].(string); ok {
					texts = append(texts, desc)
				}
			}
		}
	}

	if len(resume.ExperienceJSON) > 0 {
		var experiences []map[string]interface{}
		if json.Unmarshal(resume.ExperienceJSON, &experiences) == nil {
			for _, e := range experiences {
				if desc, ok := e["description"].(string); ok {
					texts = append(texts, desc)
				}
			}
		}
	}

	for _, text := range texts {
		textLower := strings.ToLower(text)
		for _, kw := range keywords {
			if strings.Contains(textLower, kw) {
				return true
			}
		}
	}

	return false
}
