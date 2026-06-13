// ─── H18-D: Fast Match SEP Engine ─────────────────────────────────────────
// Evidence-Grounded Fast Match: Direction Fit + Hard Gate + Requirement Coverage
// + Evidence Mapping + Risk Detection + Score Calibration

package service

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"offer-catcher/server/internal/model"
)

// ─── H18-D Match Engine Types ────────────────────────────────────────────

type MatchSEPContext struct {
	ResumeProfile *model.SemanticProfile
	JobProfile    *model.SemanticProfile
	ResumeTags    []*model.SemanticProfileTag
	JobTags       []*model.SemanticProfileTag
	JobTitle      string
	ResumeID      int64
	JobID         int64
}

type DirectionFitResult struct {
	Level     string   `json:"level"` // strong/adjacent/transferable/weak/mismatch
	Score     int      `json:"score"`
	Reason    string   `json:"reason"`
	Evidence  []string `json:"evidence"`
	ScoreCap  int      `json:"score_cap"`
}

type RequirementCoverageItem struct {
	RequirementName string   `json:"requirement_name"`
	RequirementType string   `json:"requirement_type"`
	CoverageStatus  string   `json:"coverage_status"`
	CoverageScore   int      `json:"coverage_score"`
	MatchedEvidence []string `json:"matched_evidence"`
	Gap             string   `json:"gap"`
	Suggestion      string   `json:"suggestion"`
}

type ResumeEvidenceItem struct {
	EvidenceName        string `json:"evidence_name"`
	EvidenceLevel       int    `json:"evidence_level"`
	SourceType          string `json:"source_type"`
	SourceTitle         string `json:"source_title"`
	EvidenceText        string `json:"evidence_text"`
	SupportsRequirement string `json:"supports_requirement"`
	Confidence          string `json:"confidence"`
}

type MatchRiskPoint struct {
	RiskName      string `json:"risk_name"`
	RiskLevel     string `json:"risk_level"`
	WhyItMatters  string `json:"why_it_matters"`
	EvidenceGap   string `json:"evidence_gap"`
	FixSuggestion string `json:"fix_suggestion"`
}

type ScoreBreakdownResult struct {
	DirectionFit       int `json:"direction_fit"`
	MustHaveCoverage   int `json:"must_have_coverage"`
	EvidenceStrength   int `json:"evidence_strength"`
	ExperienceRelevance int `json:"experience_relevance"`
	RiskControl        int `json:"risk_control"`
}

type HardGateResult struct {
	Status   string   `json:"status"` // passed/weak_pass/warning/blocked
	ScoreCap int      `json:"score_cap"`
	Warnings []string `json:"warnings"`
}

// ─── Normalization ────────────────────────────────────────────────────────

var termNormalizationMap = map[string]string{
	// Programming languages
	"go": "go", "golang": "go", "go语言": "go",
	"java": "java", "java后端": "java", "java开发": "java",
	"python": "python", "python编程": "python",
	"c++": "cpp", "cpp": "cpp",
	"javascript": "javascript", "js": "javascript", "typescript": "typescript", "ts": "typescript",
	"rust": "rust", "kotlin": "kotlin",
	// Data stores
	"mysql": "mysql", "sql": "mysql", "关系型数据库": "mysql", "数据库": "mysql",
	"redis": "redis", "缓存": "redis",
	"mongodb": "mongodb", "elasticsearch": "elasticsearch", "es": "elasticsearch",
	"postgresql": "postgresql", "postgres": "postgresql",
	// Infrastructure
	"docker": "docker", "容器化": "docker", "容器": "docker",
	"kubernetes": "kubernetes", "k8s": "kubernetes",
	"linux": "linux",
	"nginx": "nginx",
	"git": "git",
	// Directions
	"后端": "backend", "后台": "backend", "服务端": "backend", "后端开发": "backend", "后台开发": "backend",
	"算法": "algorithm", "机器学习": "algorithm", "深度学习": "algorithm", "ai": "algorithm",
	"nlp": "nlp", "自然语言处理": "nlp",
	"cv": "cv", "计算机视觉": "cv",
	"前端": "frontend", "web前端": "frontend",
	"数据": "data", "数据分析": "data", "数据工程": "data",
	"产品": "product", "产品策划": "product", "产品运营": "product",
	"运营": "operation", "用户运营": "operation",
	"测试": "testing", "测试开发": "testing",
	"设计": "design",
	"安全": "security",
	"云计算": "cloud", "云平台": "cloud",
	"大数据": "bigdata",
	// Engineering practices
	"微服务": "microservice",
	"分布式": "distributed",
	"高并发": "high_concurrency", "并发": "high_concurrency",
	"限流": "rate_limit", "熔断": "circuit_breaker",
	"单元测试": "unit_test",
	"ci/cd": "cicd", "持续集成": "cicd", "devops": "cicd",
}

var normalizeWordRe = regexp.MustCompile(`[\s\-_/、，。；：,\.]+`)

func normalizeTerm(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = normalizeWordRe.ReplaceAllString(s, "")
	if mapped, ok := termNormalizationMap[s]; ok {
		return mapped
	}
	for k, v := range termNormalizationMap {
		if strings.Contains(s, k) {
			return v
		}
	}
	return s
}

func normalizeTermSet(tags []*model.SemanticProfileTag) map[string]string {
	out := make(map[string]string)
	for _, t := range tags {
		norm := normalizeTerm(t.TagLabel)
		out[norm] = t.TagLabel
	}
	return out
}

func termMatch(resumeSet, jobSet map[string]string) (matched, missing []string) {
	for n, jl := range jobSet {
		if rl, ok := resumeSet[n]; ok {
			matched = append(matched, rl)
		} else {
			missing = append(missing, jl)
		}
	}
	return
}

// ─── Direction Fit Analyzer ───────────────────────────────────────────────

var directionMapping = map[string][]string{
	"backend":    {"backend", "java", "go", "cpp", "python", "microservice", "distributed", "mysql", "redis", "docker", "linux", "cloud", "cicd", "high_concurrency"},
	"algorithm":  {"algorithm", "nlp", "cv", "ai", "python", "tensorflow", "pytorch", "data"},
	"frontend":   {"frontend", "javascript", "typescript", "react", "vue", "css", "html"},
	"data":       {"data", "bigdata", "mysql", "python", "spark", "hadoop"},
	"product":    {"product", "operation"},
	"testing":    {"testing", "python", "java", "cicd", "sql"},
	"cloud":      {"cloud", "docker", "kubernetes", "linux", "devops"},
	"security":   {"security", "linux", "network"},
}

var adjacentGroups = map[string][]string{
	"backend":  {"cloud", "data", "testing"},
	"algorithm": {"data", "backend"},
	"cloud":    {"backend", "data", "security"},
	"data":     {"algorithm", "backend", "cloud"},
	"testing":  {"backend", "frontend"},
}

func detectDirection(tags []*model.SemanticProfileTag, title string) string {
	set := normalizeTermSet(tags)
	scores := make(map[string]int)
	for dir, keywords := range directionMapping {
		for _, kw := range keywords {
			if _, ok := set[kw]; ok {
				scores[dir]++
			}
		}
	}
	// Heuristic: check job title for direction hints
	titleLow := strings.ToLower(title)
	if strings.Contains(titleLow, "后端") || strings.Contains(titleLow, "后台") || strings.Contains(titleLow, "服务端") {
		scores["backend"] += 3
	}
	if strings.Contains(titleLow, "算法") || strings.Contains(titleLow, "机器学习") || strings.Contains(titleLow, "深度学习") || strings.Contains(titleLow, "nlp") || strings.Contains(titleLow, "自然语言") {
		scores["algorithm"] += 3
	}
	if strings.Contains(titleLow, "前端") || strings.Contains(titleLow, "web") || strings.Contains(titleLow, "h5") {
		scores["frontend"] += 3
	}
	if strings.Contains(titleLow, "数据") || strings.Contains(titleLow, "大数据") {
		scores["data"] += 3
	}
	if strings.Contains(titleLow, "产品") {
		scores["product"] += 3
	}
	if strings.Contains(titleLow, "测试") {
		scores["testing"] += 3
	}
	if strings.Contains(titleLow, "运营") {
		scores["operation"] += 3
	}

	best := ""
	bestScore := 0
	for d, s := range scores {
		if s > bestScore {
			bestScore = s
			best = d
		}
	}
	if bestScore == 0 {
		return ""
	}
	return best
}

func analyzeDirectionFit(ctx *MatchSEPContext) DirectionFitResult {
	resumeDir := detectDirection(ctx.ResumeTags, "")
	jobDir := detectDirection(ctx.JobTags, ctx.JobTitle)

	result := DirectionFitResult{Level: "weak", Score: 40, ScoreCap: 68}

	if resumeDir == "" && jobDir == "" {
		result.Level = "weak"
		result.Reason = "无法从 SE profile 中提取明确的岗位方向"
		result.ScoreCap = 68
		return result
	}
	if resumeDir == "" {
		result.Level = "transferable"
		result.Reason = "简历方向不明确，但存在可迁移的技术能力"
		result.ScoreCap = 78
		result.Score = 55
		return result
	}

	if resumeDir == jobDir {
		result.Level = "strong"
		result.Score = 95
		result.Reason = fmt.Sprintf("简历方向(%s)与岗位方向完全一致", resumeDir)
		result.ScoreCap = 100
		return result
	}

	// Check adjacent
	if adjacents, ok := adjacentGroups[resumeDir]; ok {
		for _, a := range adjacents {
			if a == jobDir {
				result.Level = "adjacent"
				result.Score = 80
				result.Reason = fmt.Sprintf("简历方向(%s)与岗位方向(%s)相邻", resumeDir, jobDir)
				result.ScoreCap = 88
				return result
			}
		}
	}

	// Check transferable — shared keywords in different domains
	rSet := normalizeTermSet(ctx.ResumeTags)
	jSet := normalizeTermSet(ctx.JobTags)
	crossHits := 0
	for n := range jSet {
		if _, ok := rSet[n]; ok {
			crossHits++
		}
	}
	if crossHits >= 3 {
		result.Level = "transferable"
		result.Score = 68
		result.Reason = "存在跨方向的技能重合，能力可迁移"
		result.ScoreCap = 78
		return result
	}

	// Complete mismatch
	if resumeDir != jobDir && crossHits < 2 {
		result.Level = "mismatch"
		result.Score = 20
		result.Reason = fmt.Sprintf("简历方向(%s)与岗位方向(%s)明显不一致", resumeDir, jobDir)
		result.ScoreCap = 60
		return result
	}

	result.Level = "weak"
	result.Score = 45
	result.Reason = "技能重合较少，方向匹配度偏低"
	result.ScoreCap = 68
	return result
}

// ─── Hard Gate Checker ────────────────────────────────────────────────────

func checkHardGate(direction DirectionFitResult, resumeTags, jobTags []*model.SemanticProfileTag) HardGateResult {
	gate := HardGateResult{Status: "passed", ScoreCap: 100}

	// ─── Cap 1: Direction mismatch ────────────────────────────────────
	if direction.Level == "mismatch" {
		gate.Status = "blocked"
		gate.ScoreCap = 60
		gate.Warnings = append(gate.Warnings, "简历与岗位方向明显不一致，匹配分数上限60")
		return gate
	}
	if direction.Level == "weak" {
		gate.Status = "warning"
		gate.ScoreCap = 65
		gate.Warnings = append(gate.Warnings, "方向匹配偏弱，建议确认岗位适配性，分数上限65")
	}

	// ─── Cap 2: Missing core must-have skills ─────────────────────────
	missedCore := 0
	jSet := normalizeTermSet(jobTags)
	rSet := normalizeTermSet(resumeTags)
	for n := range jSet {
		if _, ok := rSet[n]; !ok {
			if confidenceOfTag(jobTags, n) >= 0.85 {
				missedCore++
			}
		}
	}
	if missedCore >= 3 {
		if gate.Status == "passed" { gate.Status = "warning" }
		if gate.ScoreCap > 75 {
			gate.ScoreCap = 75
		}
		gate.Warnings = append(gate.Warnings, fmt.Sprintf("缺少%d个核心要求标签，分数上限75", missedCore))
	}

	// ─── Cap 3/4/5: Evidence strength ─────────────────────────────────
	hasProjectEvidence := false
	hasInternshipEvidence := false
	hasAnyEvidence := false
	hasStrongExpression := false
	for _, t := range resumeTags {
		if t.EvidenceText != "" {
			hasAnyEvidence = true
			evLow := strings.ToLower(t.EvidenceText)
			// "实习" and "项目" are both project-level evidence
			if strings.Contains(evLow, "项目") || strings.Contains(evLow, "实习") ||
				strings.Contains(strings.ToLower(t.Category), "experience") {
				hasProjectEvidence = true
			}
			if strings.Contains(evLow, "实习") || strings.Contains(evLow, "公司") || strings.Contains(evLow, "上线") {
				hasInternshipEvidence = true
			}
			if t.Confidence >= 0.85 && len(t.EvidenceText) > 15 {
				hasStrongExpression = true
			}
		}
	}
	if !hasAnyEvidence {
		if gate.Status == "passed" { gate.Status = "warning" }
		if gate.ScoreCap > 72 {
			gate.ScoreCap = 72
		}
		gate.Warnings = append(gate.Warnings, "仅技能列表命中，无项目/实习证据，分数上限72")
	}
	if !hasProjectEvidence && hasAnyEvidence {
		if gate.ScoreCap > 70 {
			gate.ScoreCap = 70
		}
		gate.Warnings = append(gate.Warnings, "无相关项目证据，分数上限70")
	}

	// ─── Cap 5: No internship but job may depend on it ─────────────────
	// Heuristic: if job description contains intern/experience keywords
	jobTitleLow := ""  // inferred from context; use direction
	_ = jobTitleLow
	if !hasInternshipEvidence && hasProjectEvidence {
		// Not blocking, just logging. Full internship check needs JD text.
	}

	// ─── Cap 6: Evidence exists but weak expression ────────────────────
	if hasAnyEvidence && !hasStrongExpression {
		if gate.ScoreCap > 85 {
			gate.ScoreCap = 85
		}
		gate.Warnings = append(gate.Warnings, "证据表达偏弱，建议补充STAR法则和量化结果，分数上限85")
	}

	// ─── Cap 8: Strong match allow 90+ ─────────────────────────────────
	if direction.Level == "strong" && hasInternshipEvidence && gate.ScoreCap >= 90 {
		gate.ScoreCap = 95
		gate.Warnings = append(gate.Warnings, "方向强一致且有实习证据，允许90+高分")
	}

	return gate
}

func confidenceOfTag(tags []*model.SemanticProfileTag, normKey string) float64 {
	for _, t := range tags {
		if normalizeTerm(t.TagLabel) == normKey {
			return t.Confidence
		}
	}
	return 0
}

// ─── Requirement Coverage Engine ──────────────────────────────────────────

func buildRequirementCoverage(ctx *MatchSEPContext) []RequirementCoverageItem {
	result := []RequirementCoverageItem{}
	resumeSet := normalizeTermSet(ctx.ResumeTags)

	// Pre-build evidence level map for status determination
	evLevels := map[string]int{}
	for _, t := range ctx.ResumeTags {
		norm := normalizeTerm(t.TagLabel)
		lvl := 2
		if t.EvidenceText != "" {
			lvl = 3
			evLow := strings.ToLower(t.EvidenceText)
			if strings.Contains(evLow, "实习") || strings.Contains(evLow, "公司") { lvl = 4 }
			if strings.Contains(evLow, "生产") || strings.Contains(evLow, "上线") { lvl = 5 }
		}
		if evLevels[norm] < lvl { evLevels[norm] = lvl }
	}

	for _, jt := range ctx.JobTags {
		reqType := classifyRequirementType(jt.Category, jt.TagLabel)
		normKey := normalizeTerm(jt.TagLabel)
		resumeLabel, matched := resumeSet[normKey]
		evLvl := evLevels[normKey]

		status := "missing"
		score := 0
		gap := jt.TagLabel
		suggestion := fmt.Sprintf("补充%s相关经历或技能描述", jt.TagLabel)
		evidence := []string{}

		if matched {
			if evLvl >= 4 {
				// Level 4+ evidence → strong
				status = "strong"
				score = 85
				if evLvl >= 5 {
					score = 95
				}
				evidence = append(evidence, resumeLabel)
				gap = ""
				suggestion = ""
			} else if evLvl >= 3 {
				// Level 3: evidence exists (text non-empty) → partial
				status = "partial"
				score = 65
				evidence = append(evidence, resumeLabel)
				gap = fmt.Sprintf("%s证据不够充分", jt.TagLabel)
				suggestion = fmt.Sprintf("强化%s相关经历的细节表达", jt.TagLabel)
			} else if evLvl >= 2 {
				// Only skill list → weak
				status = "weak"
				score = 40
				gap = fmt.Sprintf("%s仅出现在技能列表中，缺乏项目证据", jt.TagLabel)
				suggestion = fmt.Sprintf("为%s补充项目经历或实习经验", jt.TagLabel)
			} else {
				status = "weak"
				score = 30
				evidence = append(evidence, resumeLabel)
				suggestion = fmt.Sprintf("进一步加强%s的能力表达", jt.TagLabel)
			}
		} else {
			// Not directly matched — check for adjacent/transferable skills
			transferableFound := false
			normJobTag := normalizeTerm(jt.TagLabel)
			for normResumeTag := range resumeSet {
				if hasCommonPrefix(normJobTag, normResumeTag, 3) || hasCommonPrefix(normResumeTag, normJobTag, 3) {
					status = "transferable"
					score = 35
					transferableFound = true
					gap = fmt.Sprintf("无%s直接证据，但有相邻%s能力", jt.TagLabel, resumeSet[normResumeTag])
					suggestion = fmt.Sprintf("强调%s与%s的关联性，展示可迁移能力", resumeSet[normResumeTag], jt.TagLabel)
					break
				}
			}
			if !transferableFound {
				// Check if truly unknown (resume data insufficient)
				if jt.Confidence < 0.6 {
					status = "unknown"
					score = 20
					gap = "简历信息不足以判断此要求"
					suggestion = fmt.Sprintf("建议在简历中明确提及%s相关经验", jt.TagLabel)
				} else {
					status = "missing"
					score = 0
				}
			}
		}

		result = append(result, RequirementCoverageItem{
			RequirementName: jt.TagLabel,
			RequirementType: reqType,
			CoverageStatus:  status,
			CoverageScore:   score,
			MatchedEvidence: evidence,
			Gap:             gap,
			Suggestion:      suggestion,
		})
	}

	if len(result) > 8 {
		result = result[:8]
	}
	return result
}

func hasCommonPrefix(a, b string, minLen int) bool {
	if len(a) < minLen || len(b) < minLen { return false }
	return a[:minLen] == b[:minLen]
}

func classifyRequirementType(category, label string) string {
	cat := strings.ToLower(category)
	lab := strings.ToLower(label)
	if strings.Contains(cat, "technical_skill") || strings.Contains(lab, "编程") || strings.Contains(lab, "语言") {
		return "must_have"
	}
	if strings.Contains(cat, "role_direction") || strings.Contains(cat, "domain") {
		return "must_have"
	}
	if strings.Contains(cat, "engineering") || strings.Contains(cat, "capability") {
		return "should_have"
	}
	if strings.Contains(cat, "experience") {
		return "should_have"
	}
	if strings.Contains(cat, "tool") || strings.Contains(cat, "certification") || strings.Contains(cat, "soft") {
		return "nice_to_have"
	}
	return "should_have"
}

// ─── Evidence Mapper ──────────────────────────────────────────────────────

// humanizeRequirementCategory maps raw SEP tag categories to user-facing Chinese labels.
func humanizeRequirementCategory(category string) string {
	cat := strings.ToLower(strings.TrimSpace(category))
	switch {
	case strings.Contains(cat, "technical_skill") || strings.Contains(cat, "tech_skill") || strings.Contains(cat, "skill"):
		return "技术能力要求"
	case strings.Contains(cat, "role_direction") || strings.Contains(cat, "direction"):
		return "岗位方向要求"
	case strings.Contains(cat, "project_exp") || strings.Contains(cat, "project_experience"):
		return "项目经验要求"
	case strings.Contains(cat, "experience_signal") || strings.Contains(cat, "internship"):
		return "实践经验要求"
	case strings.Contains(cat, "education") || strings.Contains(cat, "degree"):
		return "学历与专业要求"
	case strings.Contains(cat, "tool") || strings.Contains(cat, "devops") || strings.Contains(cat, "engineering"):
		return "工程工具能力"
	case strings.Contains(cat, "domain") || strings.Contains(cat, "context"):
		return "业务领域能力"
	case strings.Contains(cat, "risk_signal"):
		return "风险关注项"
	case strings.Contains(cat, "preference"):
		return "偏好匹配要求"
	default:
		return "岗位要求"
	}
}

func mapResumeEvidence(ctx *MatchSEPContext) []ResumeEvidenceItem {
	result := []ResumeEvidenceItem{}
	for _, t := range ctx.ResumeTags {
		level := 2 // default: skill list level
		if t.EvidenceText != "" {
			level = 3
			evLow := strings.ToLower(t.EvidenceText)
			if strings.Contains(evLow, "项目") { level = 3 }
			if strings.Contains(evLow, "实习") { level = 4 }
			if strings.Contains(evLow, "生产") || strings.Contains(evLow, "上线") || strings.Contains(evLow, "部署") {
				level = 5
			}
		}
		sourceType := "skill_list"
		if strings.Contains(strings.ToLower(t.Category), "experience") {
			sourceType = "project"
		}
		if strings.Contains(strings.ToLower(t.Category), "project") {
			sourceType = "project"
		}
		if strings.Contains(strings.ToLower(t.EvidenceText), "实习") || strings.Contains(strings.ToLower(t.EvidenceText), "公司") {
			sourceType = "internship"
		}

		confidence := "medium"
		if t.Confidence >= 0.9 { confidence = "high" }
		if t.Confidence < 0.8 { confidence = "low" }

		// H18-D: Humanize SupportsRequirement, never leak raw Category codes
		supportsReq := humanizeRequirementCategory(t.Category)
		if t.Explanation != "" {
			supportsReq = t.Explanation
		}

		result = append(result, ResumeEvidenceItem{
			EvidenceName:  t.TagLabel,
			EvidenceLevel: level,
			SourceType:    sourceType,
			SourceTitle:   t.TagLabel,
			EvidenceText:  t.EvidenceText,
			SupportsRequirement: supportsReq,
			Confidence:    confidence,
		})
	}
	if len(result) > 8 {
		result = result[:8]
	}
	return result
}

// ─── Risk Detector ────────────────────────────────────────────────────────

func detectMatchRisks(ctx *MatchSEPContext, direction DirectionFitResult) []MatchRiskPoint {
	risks := []MatchRiskPoint{}

	if direction.Level == "mismatch" || direction.Level == "weak" {
		risks = append(risks, MatchRiskPoint{
			RiskName:      fmt.Sprintf("岗位方向匹配度偏低(%s)", direction.Level),
			RiskLevel:     "high",
			WhyItMatters:  "方向不一致会显著影响HR简历筛选通过率",
			EvidenceGap:   direction.Reason,
			FixSuggestion: "建议优先投递方向更匹配的岗位，或针对该方向补充具体项目经历",
		})
	}

	// Check for missing hard skills
	missingHard := 0
	resumeSet := normalizeTermSet(ctx.ResumeTags)
	for _, jt := range ctx.JobTags {
		if _, ok := resumeSet[normalizeTerm(jt.TagLabel)]; !ok && jt.Confidence >= 0.85 {
			missingHard++
		}
	}
	if missingHard >= 3 {
		risks = append(risks, MatchRiskPoint{
			RiskName:      fmt.Sprintf("核心技能覆盖不足(缺少%d项)", missingHard),
			RiskLevel:     "high",
			WhyItMatters:  "岗位核心技能要求未在简历中体现，可能被系统筛选过滤",
			EvidenceGap:   "SEP 标签覆盖分析发现多个高置信度岗位标签无对应简历标签",
			FixSuggestion: "在简历中补充相关技术栈描述，或补充项目经历体现对应技能",
		})
	}

	// Evidence strength
	hasStrongEvidence := false
	for _, t := range ctx.ResumeTags {
		if t.EvidenceText != "" && t.Confidence >= 0.85 {
			hasStrongEvidence = true
			break
		}
	}
	if !hasStrongEvidence && len(ctx.ResumeTags) > 0 {
		risks = append(risks, MatchRiskPoint{
			RiskName:      "简历证据强度偏弱",
			RiskLevel:     "medium",
			WhyItMatters:  "仅有标签命中但无具体证据支撑，容易在面试中暴露",
			EvidenceGap:   "SEP 标签中几乎无 evidence_text 或项目/实习类证据",
			FixSuggestion: "为现有技能补充具体的项目实例，使用STAR法则描述",
		})
	}

	if len(risks) == 0 {
		risks = append(risks, MatchRiskPoint{
			RiskName:      "基础能力覆盖",
			RiskLevel:     "low",
			WhyItMatters:  "当前匹配未发现明显方向性或技能性风险",
			EvidenceGap:   "",
			FixSuggestion: "可为关键技能补充量化成果描述以提升简历说服力",
		})
	}
	if len(risks) > 6 {
		risks = risks[:6]
	}
	return risks
}

// ─── Score Calibrator ─────────────────────────────────────────────────────

func calibrateFastMatchScore(
	direction DirectionFitResult,
	gate HardGateResult,
	coverage []RequirementCoverageItem,
	evidence []ResumeEvidenceItem,
	risks []MatchRiskPoint,
) (int, string, ScoreBreakdownResult) {
	// ─── Direction fit (25%) ──────────────────────────────────────────
	dirScore := clamp(direction.Score, 0, 100)

	// ─── Must-have coverage (30%) ──────────────────────────────────────
	mustCount := 0
	mustTotal := 0
	for _, c := range coverage {
		if c.RequirementType == "must_have" || c.RequirementType == "should_have" {
			mustCount++
			mustTotal += c.CoverageScore
		}
	}
	avgCoverage := 0
	if mustCount > 0 {
		avgCoverage = mustTotal / mustCount
	}
	covScore := clamp(avgCoverage, 0, 100)

	// ─── Evidence strength (25%) ───────────────────────────────────────
	avgEvLevel := 0.0
	evCount := 0
	for _, e := range evidence {
		if e.EvidenceLevel > 0 {
			avgEvLevel += float64(e.EvidenceLevel)
			evCount++
		}
	}
	if evCount > 0 {
		avgEvLevel = avgEvLevel / float64(evCount)
	}
	// Map 0-5 evidence level to 0-100
	evScore := clamp(int(avgEvLevel*20), 0, 100)

	// ─── Experience relevance (10%) ────────────────────────────────────
	expScore := 50
	hasProject := false
	hasInternship := false
	for _, e := range evidence {
		if e.EvidenceLevel >= 4 {
			if e.SourceType == "internship" || strings.Contains(strings.ToLower(e.EvidenceText), "实习") {
				hasInternship = true
			}
			hasProject = true
		}
	}
	if hasInternship {
		expScore = 90
	} else if hasProject {
		expScore = 75
	} else if evCount > 0 {
		expScore = 60
	}
	expScore = clamp(expScore, 0, 100)

	// ─── Risk control (10%) ────────────────────────────────────────────
	highCount := 0
	medCount := 0
	lowCount := 0
	for _, r := range risks {
		switch r.RiskLevel {
		case "high":
			highCount++
		case "medium":
			medCount++
		case "low":
			lowCount++
		}
	}
	riskScore := 90
	if highCount >= 2 {
		riskScore = 40
	} else if highCount == 1 {
		riskScore = 55
	} else if medCount >= 2 {
		riskScore = 65
	} else if medCount == 1 {
		riskScore = 75
	} else if lowCount > 0 {
		riskScore = 82
	}
	riskScore = clamp(riskScore, 0, 100)

	// ─── Weighted composite ─────────────────────────────────────────────
	rawScore := int(float64(dirScore)*0.25 +
		float64(covScore)*0.30 +
		float64(evScore)*0.25 +
		float64(expScore)*0.10 +
		float64(riskScore)*0.10)

	breakdown := ScoreBreakdownResult{
		DirectionFit:        dirScore,
		MustHaveCoverage:    covScore,
		EvidenceStrength:    evScore,
		ExperienceRelevance: expScore,
		RiskControl:         riskScore,
	}

	// Apply hard gate cap
	finalScore := rawScore
	if finalScore > gate.ScoreCap {
		finalScore = gate.ScoreCap
	}
	finalScore = clamp(finalScore, 0, 100)

	level := matchLevel(finalScore)

	return finalScore, level, breakdown
}

func clamp(v, lo, hi int) int {
	if v < lo { return lo }
	if v > hi { return hi }
	return v
}

// ─── Score Level ─────────────────────────────────────────────────────────

// matchLevel maps a 0-100 score to the unified 5-level semantic per H18-B contract.
func matchLevel(score int) string {
	switch {
	case score >= 90:
		return "高度匹配"
	case score >= 80:
		return "较高匹配"
	case score >= 70:
		return "中等匹配"
	case score >= 60:
		return "弱匹配"
	default:
		return "不建议投递"
	}
}

// ─── Application Decision ─────────────────────────────────────────────────

func decideApplication(score int, risks []MatchRiskPoint, gate HardGateResult) (string, string) {
	highCount := 0
	for _, r := range risks {
		if r.RiskLevel == "high" { highCount++ }
	}
	if gate.Status == "blocked" {
		return "not_recommended", "暂不建议投递"
	}
	if score >= 90 && highCount == 0 {
		return "priority_apply", "优先投递"
	}
	if score >= 80 && highCount == 0 {
		return "apply_directly", "可以直接投递"
	}
	if score >= 70 || (score >= 80 && highCount >= 1) {
		return "revise_before_apply", "修改后投递"
	}
	if score >= 60 {
		return "backup_apply", "作为备选投递"
	}
	return "not_recommended", "暂不建议投递"
}

// ─── One-Sentence Conclusion ──────────────────────────────────────────────

func buildConclusion(score int, level string, direction DirectionFitResult, gate HardGateResult, risks []MatchRiskPoint) string {
	base := ""
	switch direction.Level {
	case "strong":
		base = fmt.Sprintf("简历方向与岗位高度一致")
	case "adjacent":
		base = fmt.Sprintf("简历方向与岗位相邻")
	case "transferable":
		base = fmt.Sprintf("存在跨方向可迁移能力")
	case "mismatch":
		base = fmt.Sprintf("简历方向与岗位差异明显")
	default:
		base = "方向匹配偏弱"
	}

	detail := ""
	if gate.Status == "warning" || gate.Status == "blocked" {
		detail = "，但简历证据不足导致分数受限"
	}
	if score >= 90 {
		return fmt.Sprintf("%s%s，匹配度%d分（%s），建议优先投递", base, detail, score, level)
	}
	if score >= 80 {
		return fmt.Sprintf("%s%s，匹配度%d分（%s），建议投递", base, detail, score, level)
	}
	if score >= 70 {
		return fmt.Sprintf("%s%s，匹配度%d分（%s），建议微调后投递", base, detail, score, level)
	}
	return fmt.Sprintf("%s%s，匹配度%d分（%s），建议针对性补强后投递", base, detail, score, level)
}

// ─── Result Assembler ─────────────────────────────────────────────────────

func assembleFastMatchResult(
	ctx *MatchSEPContext,
	direction DirectionFitResult,
	gate HardGateResult,
	coverage []RequirementCoverageItem,
	evidence []ResumeEvidenceItem,
	risks []MatchRiskPoint,
	score int,
	level string,
	breakdown ScoreBreakdownResult,
	decision, decisionLabel string,
) *model.MatchAnalysisResult {
	conclusion := buildConclusion(score, level, direction, gate, risks)

	// ─── Core advantages: strong/partial coverage items, evidence-backed ──
	coreAd := []string{}
	for _, c := range coverage {
		if c.CoverageStatus == "strong" {
			coreAd = append(coreAd, fmt.Sprintf("✅ %s：有明确项目经验支撑", c.RequirementName))
		} else if c.CoverageStatus == "partial" {
			coreAd = append(coreAd, fmt.Sprintf("🔹 %s：有相关经历", c.RequirementName))
		}
	}
	if len(coreAd) == 0 {
		coreAd = append(coreAd, "简历方向与岗位基本匹配")
	}
	if len(coreAd) > 4 {
		coreAd = coreAd[:4]
	}

	// ─── Main gaps: missing/weak/transferable, specific and actionable ────
	mainGaps := []string{}
	for _, c := range coverage {
		if c.CoverageStatus == "missing" {
			mainGaps = append(mainGaps, fmt.Sprintf("缺少%s：%s", c.RequirementName, c.Suggestion))
		} else if c.CoverageStatus == "weak" || c.CoverageStatus == "transferable" {
			mainGaps = append(mainGaps, fmt.Sprintf("%s不足：%s", c.RequirementName, c.Gap))
		}
	}
	if len(mainGaps) == 0 {
		mainGaps = append(mainGaps, "当前未发现明显技能短板")
	}
	if len(mainGaps) > 5 {
		mainGaps = mainGaps[:5]
	}

	// ─── Action suggestions: from risk fix_suggestion + coverage suggestion ─
	actionSugs := []string{}
	for _, r := range risks {
		if r.FixSuggestion != "" {
			actionSugs = append(actionSugs, r.FixSuggestion)
		}
	}
	if len(actionSugs) == 0 {
		actionSugs = append(actionSugs, "针对缺失标签补充对应经历或技能描述")
	}
	if len(actionSugs) > 5 {
		actionSugs = actionSugs[:5]
	}

	// ─── Legacy compat fills ────────────────────────────────────────────────
	screeningResult := "pass"
	if gate.Status == "blocked" { screeningResult = "warning" }
	if gate.Status == "warning" { screeningResult = "weak_pass" }

	// Legacy risk_points as string[] (backward compat)
	riskStrs := []string{}
	for _, r := range risks {
		riskStrs = append(riskStrs, r.RiskName)
	}

	// Legacy optimization suggestions
	opts := []model.MatchSuggestion{}
	for i, a := range actionSugs {
		p := "P1"
		if i == 0 { p = "P0" }
		opts = append(opts, model.MatchSuggestion{
			Priority: p,
			Title:    a[:minInt(30, len(a))],
			Reason:   "",
			Action:   a,
		})
	}

	// Legacy match highlights
	matchHighlights := []string{}
	for _, c := range coverage {
		if c.CoverageStatus != "missing" {
			matchHighlights = append(matchHighlights, c.RequirementName)
		}
	}

	return &model.MatchAnalysisResult{
		MatchScore:              score,
		MatchLevel:              level,
		MatchLabel:              level,
		ScreeningResult:         screeningResult,
		EvidenceChain:           []model.MatchEvidence{},
		MatchHighlights:         matchHighlights,
		GapPoints:               mainGaps,
		RiskPoints:              riskStrs,
		OptimizationSuggestions: opts,
		NextSteps:               actionSugs,
		GeneratedBy:             "fast_sep_match_engine",
		AgentStatus:             "success",
		FallbackReason:          "none",
		RepairApplied:           false,
		DataQuality:             "sep",
		Summary:                 conclusion,
		FitPoints:               append([]string{}, coreAd...),
		ResumeEvidence:          collectEvidenceNames(evidence),
		JobRequirements:         requirementNames(coverage),

		// H18-D: New structured fields
		StructuredRisks:   risks,
		ScoreBreakdownObj: breakdown,
		HardGateObj:       gate,
		CoreAdvantages:    coreAd,
		MainGaps:          mainGaps,
		ActionSuggestions: actionSugs,
		DecisionLabel:     decisionLabel,
		ApplicationDecision: decision,
	}
}

func collectEvidenceNames(ev []ResumeEvidenceItem) []string {
	out := []string{}
	for _, e := range ev {
		out = append(out, e.EvidenceName)
	}
	return out
}

func requirementNames(coverage []RequirementCoverageItem) []string {
	out := []string{}
	for _, c := range coverage {
		out = append(out, c.RequirementName)
	}
	return out
}

func minInt(a, b int) int { if a < b { return a }; return b }
func ensureNonNil(s []string) []string { if s == nil { return []string{} }; return s }

// ─── Main SEP Fast Match Pipeline ─────────────────────────────────────────

func runFastSEPMatch(ctx *MatchSEPContext) *model.MatchAnalysisResult {
	log.Printf("[FastSEPEngine] starting for resume=%d job=%d", ctx.ResumeID, ctx.JobID)

	// 1. Direction Fit
	direction := analyzeDirectionFit(ctx)
	log.Printf("[FastSEPEngine] direction: level=%s score=%d cap=%d", direction.Level, direction.Score, direction.ScoreCap)

	// 2. Hard Gate
	gate := checkHardGate(direction, ctx.ResumeTags, ctx.JobTags)
	log.Printf("[FastSEPEngine] gate: status=%s cap=%d warns=%v", gate.Status, gate.ScoreCap, gate.Warnings)

	// 3. Requirement Coverage
	coverage := buildRequirementCoverage(ctx)
	log.Printf("[FastSEPEngine] coverage: %d items", len(coverage))

	// 4. Evidence Mapping
	evidence := mapResumeEvidence(ctx)
	log.Printf("[FastSEPEngine] evidence: %d items", len(evidence))

	// 5. Risk Detection
	risks := detectMatchRisks(ctx, direction)
	log.Printf("[FastSEPEngine] risks: %d items", len(risks))

	// 6. Score Calibration
	score, level, breakdown := calibrateFastMatchScore(direction, gate, coverage, evidence, risks)
	log.Printf("[FastSEPEngine] score: raw=%d (level=%s) breakdown=%+v", score, level, breakdown)

	// 7. Application Decision
	decision, decisionLabel := decideApplication(score, risks, gate)
	log.Printf("[FastSEPEngine] decision: %s / %s", decision, decisionLabel)

	// 8. Assemble
	result := assembleFastMatchResult(ctx, direction, gate, coverage, evidence, risks, score, level, breakdown, decision, decisionLabel)
	_ = breakdown // used for logging only for now

	return result
}

// ─── computeTagMatch (shared with fastMatchWithTags) ──────────────────────

func computeTagMatchSealed(rp, jp interface{}) (matched, missing []string, tagScore float64) {
	// stub — kept for backward compat, SEP engine uses normalizeTermSet
	return nil, nil, 0
}
