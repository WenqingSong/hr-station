package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"offer-catcher/server/internal/model"
)

// VersionComparer defines the interface for version comparison agents.
// Mock and LLM implementations both satisfy this interface,
// allowing the service layer to dispatch uniformly.
type VersionComparer interface {
	Compare(from *model.ResumeVersion, to *model.ResumeVersion) (*model.VersionCompareAnalysis, error)
}

// VersionCompareMockAgent generates mock version comparison using simple rule-based analysis.
// It does NOT call any real LLM or read LLM_API_KEY.
type VersionCompareMockAgent struct {
	// ProfileProvider optionally provides profile dominant info.
	ProfileProvider interface {
		GetProfileStatus(ctx context.Context, userID, resumeID uint64) (*model.ResumeProfileStatusResponse, error)
	}
}

// SetProfileProvider injects a profile provider for profile-aware comparison.
func (a *VersionCompareMockAgent) SetProfileProvider(p interface {
	GetProfileStatus(ctx context.Context, userID, resumeID uint64) (*model.ResumeProfileStatusResponse, error)
}) {
	a.ProfileProvider = p
}

// Compile-time check: VersionCompareMockAgent satisfies VersionComparer.
var _ VersionComparer = (*VersionCompareMockAgent)(nil)

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
		FromResumeID:        from.ID,
		ToResumeID:          to.ID,
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
			ModuleName: "projects", // K8-5-26-E: job_intention removed
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
			ModuleName: "skills",
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
			ModuleName: "projects",
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
			ModuleName: "extra_info",
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
			ModuleName: "overall",
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
		parts = append(parts, fmt.Sprintf("V%d 相比 V%d 在简历完整度上有所提升", to.ID, from.ID))
	} else if scoreAfter == scoreBefore {
		parts = append(parts, fmt.Sprintf("V%d 与 V%d 在简历完整度上基本一致", to.ID, from.ID))
	} else {
		parts = append(parts, fmt.Sprintf("V%d 相比 V%d 在简历完整度上有所下降", to.ID, from.ID))
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

// ── K9-4-C: V2 Compare (profile-aware, Fast/Quality) ──────────────────────

// CompareV2 generates a profile-aware resume comparison for the V2 API.
// qualityMode=true → full merge_suggestions + rewrite_examples + next_version_suggestions.
func (a *VersionCompareMockAgent) CompareV2(from, to *model.ResumeVersion, mode string, targetDirection string, jobID uint64, qualityMode bool, fromProfile, toProfile string) *model.CompareResponseV2 {
	now := time.Now().Format("2006-01-02T15:04:05Z07:00")
	genBy := "mock_agent"
	tier := "fast"
	if qualityMode {
		genBy = "mock_quality"
		tier = "quality"
	}

	fromProf := fromProfile
	if fromProf == "" {
		fromProf = detectProfileFromResume(from)
	}
	toProf := toProfile
	if toProf == "" {
		toProf = detectProfileFromResume(to)
	}

	score := calcCompareScore(from, to, fromProf, toProf, mode, targetDirection, jobID)
	level, winner, winnerLabel := judgeWinner(fromProf, toProf, score, mode, targetDirection)

	resp := &model.CompareResponseV2{
		CompareID:        0, // set by service after DB save
		FromResumeID:     from.ID,
		ToResumeID:       to.ID,
		ComparisonMode:   mode,
		TargetDirection:  targetDirection,
		GenerationTier:   tier,
		GeneratedBy:      genBy,
		Status:           "ready",
		FromProfileType:            tier,
		FromProfileDominantProfile: fromProf,
		ToProfileType:              tier,
		ToProfileDominantProfile:   toProf,

		CompareScore:       score,
		CompareLevel:       level,
		WinnerResumeID:     winner,
		WinnerLabel:        winnerLabel,
		OneSentenceSummary: genOneSentence(fromProf, toProf, winner),

		KeyDifferences:    genKeyDiffs(from, to, fromProf, toProf),
		FromStrengths:     genProfileStrengths(fromProf, 2),
		ToStrengths:       genProfileStrengths(toProf, 2),
		FromRisks:         genProfileRisks(fromProf, 1),
		ToRisks:           genProfileRisks(toProf, 1),
		DeliveryAdvice:    genDeliveryAdvice(winner, toProf),
		CreatedAt:         now,

		// Default empty arrays (avoid JSON null)
		FromResumeSummary:      "",
		ToResumeSummary:        "",
		KeepFromItems:          []string{},
		KeepToItems:            []string{},
		MergeSuggestions:       []string{},
		RewriteExamples:        []model.RewriteCompareExample{},
		NextVersionSuggestions: []string{},
	}

	if qualityMode {
		resp.FromResumeSummary = genResumeSummary(from, fromProf)
		resp.ToResumeSummary = genResumeSummary(to, toProf)
		resp.FromStrengths = genProfileStrengths(fromProf, 4)
		resp.ToStrengths = genProfileStrengths(toProf, 4)
		resp.FromRisks = genProfileRisks(fromProf, 3)
		resp.ToRisks = genProfileRisks(toProf, 3)
		resp.KeyDifferences = genKeyDiffs(from, to, fromProf, toProf)
		resp.KeepFromItems = genKeepItems(from, fromProf, true)
		resp.KeepToItems = genKeepItems(to, toProf, false)
		resp.MergeSuggestions = genMergeSuggestions(fromProf, toProf)
		resp.RewriteExamples = genRewriteExamples(from, to, fromProf, toProf)
		resp.NextVersionSuggestions = genNextVerSuggestions(fromProf, toProf, mode, targetDirection)
		resp.DeliveryAdvice = genDeliveryAdviceFull(winner, toProf, mode, targetDirection)
	}

	// Ensure minimums for Quality
	if qualityMode {
		if len(resp.RewriteExamples) < 2 {
			resp.RewriteExamples = genFallbackRewriteExamples(fromProf, toProf)
		}
		if len(resp.MergeSuggestions) < 2 {
			resp.MergeSuggestions = genFallbackMerge(fromProf, toProf)
		}
		if len(resp.NextVersionSuggestions) < 2 {
			resp.NextVersionSuggestions = genFallbackNextVer(toProf)
		}
	}

	return resp
}

func detectProfileFromResume(r *model.ResumeVersion) string {
	text := r.RawText
	if text == "" { text = r.ProfessionalSkills + " " + r.AIApplication }
	lower := strings.ToLower(text)
	a := countKW(lower, "算法", "pytorch", "tensorflow", "深度学习", "推荐系统", "nlp", "多模态")
	be := countKW(lower, "go ", "java", "后端", "mysql", "redis", "docker", "后台", "接口")
	prod := countKW(lower, "产品", "需求分析", "用户研究", "prd", "运营")
	biz := countKW(lower, "商务", "采购", "供应链", "解决方案", "合同")
	if a > be && a > prod { return "algorithm_research" }
	if be > a && be > prod { return "backend_engineering" }
	if prod > a && prod > be { return "product_operation" }
	if biz > 0 && biz > a && biz > be { return "business_solution" }
	return "career_exploration"
}

func countKW(text string, keywords ...string) int {
	c := 0
	for _, kw := range keywords {
		if strings.Contains(text, kw) { c++ }
	}
	return c
}

func calcCompareScore(from, to *model.ResumeVersion, fromProf, toProf string, mode, targetDir string, jobID uint64) int {
	score := 70 // base
	// Same direction → closer scores
	if fromProf == toProf {
		score += 8
	} else {
		score += 3
	}
	// target_direction mode: profile matches target
	if mode == "target_direction" && targetDir != "" {
		if toProf == targetDir { score += 10 }
		if fromProf == targetDir { score += 5 }
	}
	// Content depth bonus on "to" resume
	if len(to.RawText) > len(from.RawText) { score += 5 }
	if len(to.ProfessionalSkills) > len(from.ProfessionalSkills) { score += 3 }
	if score > 95 { score = 95 }
	if score < 60 { score = 60 }
	return score
}

func judgeWinner(fromProf, toProf string, score int, mode, targetDir string) (level string, winner uint64, label string) {
	winnerID := uint64(2)
	lvl := "目标版本更适合当前场景"
	lbl := "新版更适合作为当前主投递版本"

	if mode == "target_direction" && targetDir != "" {
		if toProf == targetDir && fromProf != targetDir {
			return "目标版本明显更适合", 2, "目标版本的方向更匹配目标方向"
		}
		if fromProf == targetDir && toProf != targetDir {
			return "原版本方向更匹配", 1, "原版本的方向更匹配目标方向"
		}
	}

	if fromProf == toProf {
		if score > 75 {
			return "两版方向一致，新版细节更优", winnerID, "两版方向一致，新版在项目深度和表达细节上更优"
		}
		return "两版方向一致，差异较小", winnerID, "两版内容差异较小，均可使用"
	}

	lbl = fmt.Sprintf("新版（%s方向）比旧版（%s方向）更适合投递", cmpProfileLabel(toProf), cmpProfileLabel(fromProf))
	return lvl, winnerID, lbl
}

var profileStrengthMap = map[string][]string{
	"algorithm_research":  {"算法与模型经验丰富", "实验评估能力较强", "研究方向明确", "机器学习/深度学习基础扎实"},
	"backend_engineering": {"后端开发技术栈完整", "工程化能力强", "具备系统设计思维", "有实际服务端项目经验"},
	"product_operation":   {"需求分析与产品思维", "用户研究/竞品分析能力", "项目推动与跨部门协作", "数据驱动决策意识"},
	"business_solution":   {"商务沟通与方案表达", "客户需求理解能力强", "行业分析与咨询思维", "文档能力与协调推进"},
	"career_exploration":  {"多方向探索有基础信号", "学习能力较强", "有进一步聚焦潜力"},
}

func genProfileStrengths(profile string, limit int) []string {
	if s, ok := profileStrengthMap[profile]; ok {
		if len(s) > limit { return s[:limit] }
		return s
	}
	return []string{"具备一定基础", "有成长潜力"}
}

var profileRiskMap = map[string][]string{
	"algorithm_research":  {"缺少工程化落地经验", "后端系统设计经验不足", "缺少大规模服务部署经历"},
	"backend_engineering": {"缺少算法/数据方向深度", "缺少研究型项目经历", "高并发场景经验可补充"},
	"product_operation":   {"缺少技术背景深度", "缺少量化运营指标", "产品上线案例可补充"},
	"business_solution":   {"缺少具体行业深耕经验", "缺少技术方案落地案例", "独立项目闭环经验可补充"},
	"career_exploration":  {"方向不够聚焦", "缺少深度项目证据", "建议先明确职业方向"},
}

func genProfileRisks(profile string, limit int) []string {
	if s, ok := profileRiskMap[profile]; ok {
		if len(s) > limit { return s[:limit] }
		return s
	}
	return []string{"建议进一步丰富简历内容"}
}

func genOneSentence(fromProf, toProf string, winner uint64) string {
	if fromProf == toProf {
		return "两版简历方向一致，建议使用目标版本继续打磨。"
	}
	if winner == 2 {
		return fmt.Sprintf("目标版本（%s）相比原版本（%s）更适合投递。", cmpProfileLabel(toProf), cmpProfileLabel(fromProf))
	}
	return fmt.Sprintf("原版本（%s）相比目标版本（%s）更适合投递。", cmpProfileLabel(fromProf), cmpProfileLabel(toProf))
}

func cmpProfileLabel(p string) string {
	switch p {
	case "algorithm_research": return "算法研究方向"
	case "backend_engineering": return "后端工程方向"
	case "product_operation": return "产品运营方向"
	case "business_solution": return "商务解决方案方向"
	default: return "职业探索方向"
	}
}

func genKeyDiffs(from, to *model.ResumeVersion, fromProf, toProf string) []string {
	diffs := []string{}
	if fromProf != toProf {
		diffs = append(diffs, fmt.Sprintf("求职方向从「%s」转向「%s」，新版投递信号更集中", cmpProfileLabel(fromProf), cmpProfileLabel(toProf)))
	}
	if len(to.ProfessionalSkills) > len(from.ProfessionalSkills) {
		diffs = append(diffs, "新版技能描述从泛覆盖升级为具体技术栈，能更直接匹配岗位初筛关键词")
	}
	if len(to.RawText) > len(from.RawText) {
		diffs = append(diffs, "新版内容更充实，从学习型描述转向实践型描述，更能证明候选人实际动手能力")
	}
	if len(to.SelfEvaluation) > len(from.SelfEvaluation) {
		diffs = append(diffs, "新版自我评价更具体，从空洞描述转为可验证的能力陈述")
	}
	// Add content direction change insight
	if toProf == "backend_engineering" && fromProf != toProf {
		diffs = append(diffs, "旧版偏探索型表达适合说明成长路径，但新版的后端项目经历才是投递后端岗位的核心证据")
	}
	if len(diffs) < 3 {
		if len(diffs) == 0 {
			diffs = append(diffs, "两版简历表达差异较小，建议从项目深度和技能细节入手提升区分度")
		}
		diffs = append(diffs, "建议新版增加项目量化结果（接口数量、数据表设计、测试通过率等）以增强说服力")
		diffs = append(diffs, "建议统一技术栈描述风格，确保每个关键词都绑定具体使用场景")
	}
	// Cap at 5
	if len(diffs) > 5 {
		diffs = diffs[:5]
	}
	return diffs
}

func genDeliveryAdvice(winner uint64, toProf string) string {
	if winner == 2 {
		return fmt.Sprintf("建议以新版（%s方向）作为主投递版本，适合投递该方向实习岗位。投递前应补充每个项目的接口数量、数据库设计和实际结果，避免只展示技术名词。", cmpProfileLabel(toProf))
	}
	return "建议保留旧版作为基础版本，参考新版中更丰富的技能和项目描述进行优化后再投递。"
}

func genDeliveryAdviceFull(winner uint64, toProf, mode, targetDir string) string {
	advice := genDeliveryAdvice(winner, toProf)
	if mode == "target_direction" && targetDir != "" {
		advice += " 本对比基于目标方向「" + targetDir + "」，建议针对性修改后投递。"
	}
	return advice
}

func genResumeSummary(r *model.ResumeVersion, prof string) string {
	skills := r.ProfessionalSkills
	if len(skills) > 80 { skills = skills[:80] + "…" }
	return fmt.Sprintf("方向：%s。技能侧重：%s。简历内容总量约 %d 字。", cmpProfileLabel(prof), skills, len(r.RawText))
}

func genKeepItems(r *model.ResumeVersion, prof string, isFrom bool) []string {
	versionLabel := "新版"
	if isFrom {
		versionLabel = "旧版"
	}
	items := []string{}
	if len(r.ProfessionalSkills) > 20 {
		items = append(items, fmt.Sprintf("保留%s的核心技术栈描述（%s），用于支撑岗位匹配筛选", versionLabel, truncateStr(r.ProfessionalSkills, 50)))
	}
	if len(r.SelfEvaluation) > 20 {
		items = append(items, fmt.Sprintf("保留%s的自我评价中具体的项目经验部分，压缩为1-2句话", versionLabel))
	}
	if len(r.AIApplication) > 10 {
		items = append(items, fmt.Sprintf("保留%s的AI应用经验，放在技能信息中作为加分项", versionLabel))
	}
	if len(items) < 2 {
		if isFrom {
			items = append(items, "保留旧版中体现学习路径和成长轨迹的内容，对于体现候选人发展过程有参考价值")
			items = append(items, "保留旧版的基础课程信息，压缩为教育经历中的一行补充")
		} else {
			items = append(items, "保留新版的项目经历，这是支撑目标岗位投递的核心证据")
			items = append(items, "保留新版中更聚焦的求职方向表达，避免简历同时传递多个岗位信号")
		}
	}
	// Cap at 4
	if len(items) > 4 {
		items = items[:4]
	}
	return items
}

func genMergeSuggestions(fromProf, toProf string) []string {
	suggestions := []string{
		"以新版作为主体，旧版的学习基础信息压缩为教育经历补充，不要单独占一个大模块",
		"删除或弱化旧版中与目标岗位无关的泛方向表达，避免投递信号混乱",
		"把旧版的学习经历转化为新版项目经历的背景句，说明候选人从学习到实践的成长过程",
	}
	if fromProf != toProf {
		suggestions = append(suggestions, fmt.Sprintf("注意统一方向表达：旧版偏向「%s」，新版偏向「%s」，建议聚焦新版方向", cmpProfileLabel(fromProf), cmpProfileLabel(toProf)))
	}
	return suggestions
}

func genRewriteExamples(from, to *model.ResumeVersion, fromProf, toProf string) []model.RewriteCompareExample {
	examples := []model.RewriteCompareExample{
		{
			Section:       "技能描述",
			FromVersion:   truncateStr(from.ProfessionalSkills, 80),
			ToVersion:     truncateStr(to.ProfessionalSkills, 80),
			SuggestedNext: truncateStr(to.ProfessionalSkills, 80) + "（建议每个技能关键词绑定到一个具体项目，说明在项目中如何使用）",
			Reason:        "新版技能描述更丰富，但如果直接将列表放上去会像堆砌。建议每个关键词都配合项目经历中的实际使用场景，这样更容易通过技术初筛。",
		},
	}
	if fromProf != toProf {
		examples = append(examples, model.RewriteCompareExample{
			Section:       "求职方向",
			FromVersion:   fmt.Sprintf("侧重 %s 方向表达", cmpProfileLabel(fromProf)),
			ToVersion:     fmt.Sprintf("聚焦 %s 方向表达", cmpProfileLabel(toProf)),
			SuggestedNext: fmt.Sprintf("求职方向：%s实习生。%s", cmpProfileLabel(toProf), "专注于该方向的技能深度和项目实践。"),
			Reason:        fmt.Sprintf("旧版方向偏泛，投递时会让HR无法快速判断候选人匹配度。建议统一为「%s」方向，让简历的投递信号在一眼内被识别。", cmpProfileLabel(toProf)),
		})
	}
	// Add project rewrite example if raw text differs significantly
	if len(to.RawText) > len(from.RawText)+100 {
		examples = append(examples, model.RewriteCompareExample{
			Section:       "项目经历",
			FromVersion:   "偏向课程学习的描述风格",
			ToVersion:     "包含具体项目和技术的描述风格",
			SuggestedNext: "下一个版本可以用「背景 / 职责 / 技术栈 / 结果」四段式重写项目经历，每个项目补充一个量化结果",
			Reason:        "从学习型描述升级为实践型描述后，还需要补充量化数据才能让简历更有说服力。",
		})
	}
	return examples
}

func genNextVerSuggestions(fromProf, toProf, mode, targetDir string) []string {
	suggestions := []string{
		fmt.Sprintf("把新版作为主版本，旧版仅保留学习背景描述，不再保留大段泛方向表达"),
		fmt.Sprintf("每个项目补充一个量化结果：接口数量、数据表数量、测试通过率或部署结果"),
		fmt.Sprintf("将技能关键词绑定到项目经历中，避免技能列表和项目内容脱节"),
	}
	if mode == "target_direction" {
		suggestions = append(suggestions, fmt.Sprintf("围绕「%s」方向调整项目描述顺序，把与该方向最相关的项目放在首位", targetDir))
	}
	if fromProf != toProf {
		suggestions = append(suggestions, fmt.Sprintf("统一全篇求职信号为「%s」方向，删除与目标岗位无关的探索描述", cmpProfileLabel(toProf)))
	}
	return suggestions
}

func genFallbackRewriteExamples(fromProf, toProf string) []model.RewriteCompareExample {
	return []model.RewriteCompareExample{
		{Section: "技能", FromVersion: "原版本技能描述", ToVersion: "目标版本技能描述更丰富", SuggestedNext: "建议整合目标版本的技能关键词", Reason: "提升简历关键词匹配度"},
		{Section: "表达", FromVersion: "原版本表述偏概括", ToVersion: "目标版本表述更具体", SuggestedNext: "建议统一为具体、量化的表达方式", Reason: "具体表达更能通过ATS筛选"},
	}
}

func genFallbackMerge(fromProf, toProf string) []string {
	return []string{
		"将目标版本的优势模块合并到原版本框架中",
		"统一技能和方向表达，避免信号混乱",
	}
}

func genFallbackNextVer(prof string) []string {
	return []string{
		fmt.Sprintf("聚焦 %s 方向，强化核心技能表达", cmpProfileLabel(prof)),
		"补充量化成果和项目细节",
	}
}

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen { return s }
	return s[:maxLen] + "…"
}

// ── Original helpers (restored) ────────────────────────────────────────────

func countMissing(skillSet map[string]bool, keywords []string) int {
	count := 0
	for _, kw := range keywords {
		if !skillSet[kw] { count++ }
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
				if desc, ok := p["description"].(string); ok { texts = append(texts, desc) }
			}
		}
	}
	if len(resume.ExperienceJSON) > 0 {
		var experiences []map[string]interface{}
		if json.Unmarshal(resume.ExperienceJSON, &experiences) == nil {
			for _, e := range experiences {
				if desc, ok := e["description"].(string); ok { texts = append(texts, desc) }
			}
		}
	}
	for _, text := range texts {
		textLower := strings.ToLower(text)
		for _, kw := range quantKeywords {
			if strings.Contains(textLower, kw) { return true }
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
				if desc, ok := p["description"].(string); ok { texts = append(texts, desc) }
			}
		}
	}
	if len(resume.ExperienceJSON) > 0 {
		var experiences []map[string]interface{}
		if json.Unmarshal(resume.ExperienceJSON, &experiences) == nil {
			for _, e := range experiences {
				if desc, ok := e["description"].(string); ok { texts = append(texts, desc) }
			}
		}
	}
	for _, text := range texts {
		textLower := strings.ToLower(text)
		for _, kw := range keywords {
			if strings.Contains(textLower, kw) { return true }
		}
	}
	return false
}

