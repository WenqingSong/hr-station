// ─── Fast SEP-based Recommendation Ranker ──────────────────────────────────────
//
// 职责：基于 SEP semantic_profile_tags 进行语义匹配、打分、排序。
// 不调用 LLM，纯规则匹配。
// 需要 normalize 层处理 LLM 生成的 tag_code 不完全一致的问题。

package service

import (
	"fmt"
	"math"
	"strings"

	"offer-catcher/server/internal/model"
)

// ─── Tag Normalization ──────────────────────────────────────────────────────

// normalizeSemanticTagKey converts an LLM-generated tag_code to a normalized key for matching.
func normalizeSemanticTagKey(tag *model.SemanticProfileTag) string {
	code := strings.ToLower(strings.TrimSpace(tag.TagCode))
	label := strings.ToLower(strings.TrimSpace(tag.TagLabel))

	// synonym mappings — map LLM-produced codes to canonical match keys
	syn := map[string]string{
		"go": "", "golang": "", "go语言": "", "go开发": "", "go后": "", "go_backend": "",
		"java": "", "java语言": "", "java开发": "", "java后": "", "java_backend": "",
		"python": "", "python开发": "", "python语言": "", "python_": "",
		"c": "", "c语言": "", "c开发": "", "c/c": "", "c_c": "",
		"mysql": "", "sql": "", "postgresql": "", "postgres": "", "database": "", "数据库": "",
		"redis": "", "redis缓": "", "redis_": "", "缓": "", "cache": "",
		"后端": "", "后台": "", "服务端": "", "backend": "", "后端开": "", "server": "",
		"前端": "", "frontend": "", "front": "", "react": "", "vue": "",
		"docker": "", "k8s": "", "kubernetes": "", "容器": "", "部署": "", "deploy": "", "nginx": "",
		"算法": "", "algorithm": "", "推荐算法": "",
		"ml": "", "深度学习": "", "pytorch": "", "tensorflow": "", "ai": "", "machine": "",
		"测试": "", "test": "", "测试开": "", "test_dev": "",
		"云": "", "cloud": "", "云原生": "", "cloud_native": "",
		"linux": "", "linux系统": "", "linux_": "",
		"网络": "", "network": "", "tcp": "", "udp": "", "http": "",
		"鉴权": "", "权限": "", "auth": "", "permission": "", "rbac": "",
		"幂等": "", "idempotent": "", "idempotency": "",
		"限流": "", "rate_limit": "", "rate-limit": "",
		"产品": "", "product": "", "运营": "", "operation": "",
		"数据": "", "data": "", "数据分析": "", "data_analysis": "",
		"实习": "", "internship": "", "实习经历": "",
		"项目": "", "project": "", "项目经验": "", "project_exp": "",
		"论文": "", "paper": "", "学术": "", "research": "",
		"学历": "", "degree": "", "本科": "", "bachelor": "", "硕士": "", "master": "",
		"cpp": "", "c++": "", "cplusplus": "",
	}

	// Build lookup from combined code + label
	keys := []string{code, label}
	if syn[code] == "" {
		candidate := code
		// Try partial matches
		for pattern, _ := range syn {
			if strings.Contains(code, pattern) || strings.Contains(label, pattern) {
				candidate = pattern
				break
			}
		}
		if candidate != code {
			keys = append(keys, candidate)
		}
	}

	// Check if any key maps to a canonical tag
	for _, k := range keys {
		if _, ok := syn[k]; ok {
			return k
		}
	}

	// Fallback: use the category + first 3 words of code
	words := strings.Split(strings.ReplaceAll(strings.ReplaceAll(code, "_", " "), "-", " "), " ")
	if len(words) > 3 {
		words = words[:3]
	}
	return tag.Category + ":" + strings.Join(words, "_")
}

// matchesJobTag checks if a resume tag could match a job tag semantically.
func matchesJobTag(resumeTag, jobTag *model.SemanticProfileTag) bool {
	// Must be same category
	if resumeTag.Category != jobTag.Category {
		return false
	}
	// Normalize and compare
	rNorm := normalizeSemanticTagKey(resumeTag)
	jNorm := normalizeSemanticTagKey(jobTag)
	if strings.Contains(rNorm, jNorm) || strings.Contains(jNorm, rNorm) {
		return true
	}
	// Also try direct code/label matching
	rCode := strings.ToLower(resumeTag.TagCode)
	jCode := strings.ToLower(jobTag.TagCode)
	if rCode == jCode || strings.Contains(rCode, jCode) || strings.Contains(jCode, rCode) {
		return true
	}
	return false
}

// ─── M2B: Career Direction Gate ───────────────────────────────────────────────

// careerDirection represents the normalized career direction of a resume or job.
type careerDirection string

const (
	dirBackend    careerDirection = "backend"
	dirFullstack  careerDirection = "fullstack"
	dirFrontend   careerDirection = "frontend"
	dirClient     careerDirection = "client"
	dirAlgorithm  careerDirection = "algorithm"
	dirData       careerDirection = "data"
	dirTest       careerDirection = "test"
	dirAIApp      careerDirection = "ai_app"
	dirOperation  careerDirection = "operation"
	dirProduct    careerDirection = "product"
	dirMarket     careerDirection = "market"
	dirUnknown    careerDirection = "unknown"
)

// inferResumeCareerDirection determines the primary career direction from resume SEP tags.
func inferResumeCareerDirection(tags []*model.SemanticProfileTag) careerDirection {
	scores := map[careerDirection]int{}
	for _, t := range tags {
		dir := inferDirectionFromText(t.TagLabel + " " + t.TagCode)
		if dir != dirUnknown {
			scores[dir] += int(t.Weight + 1)
		}
	}
	// Also infer from evidence text for stronger signal
	for _, t := range tags {
		if strings.Contains(t.EvidenceText, "后端") || strings.Contains(t.EvidenceText, "Java") || strings.Contains(t.EvidenceText, "Spring") {
			scores[dirBackend] += 3
		}
	}
	best := dirUnknown
	bestScore := 0
	for d, s := range scores {
		if s > bestScore { best, bestScore = d, s }
	}
	return best
}

// inferJobCareerDirection determines the career direction from job metadata + SEP tags.
func inferJobCareerDirection(title, direction, category string, tags []*model.SemanticProfileTag) careerDirection {
	text := strings.ToLower(title + " " + direction + " " + category)
	// Title-based heuristics first (strongest signal)
	if strings.Contains(text, "后端") || strings.Contains(text, "后台") || strings.Contains(text, "服务端") {
		return dirBackend
	}
	if strings.Contains(text, "全栈") {
		return dirFullstack
	}
	if strings.Contains(text, "前端") || strings.Contains(text, "web前端") || strings.Contains(text, "web开发") {
		return dirFrontend
	}
	if strings.Contains(text, "客户端") || strings.Contains(text, "pc客户端") || strings.Contains(text, "游戏引擎") {
		return dirClient
	}
	if strings.Contains(text, "算法") || strings.Contains(text, "多媒体处理") || strings.Contains(text, "nlp") || strings.Contains(text, "机器学习") || strings.Contains(text, "深度学习") || strings.Contains(text, "计算机视觉") || strings.Contains(text, "推荐算法") || strings.Contains(text, "大模型") {
		return dirAlgorithm
	}
	if strings.Contains(text, "数据工程") || strings.Contains(text, "大数据") || strings.Contains(text, "数据仓库") || strings.Contains(text, "etl") {
		return dirData
	}
	if strings.Contains(text, "测试") || strings.Contains(text, "质量") {
		return dirTest
	}
	if strings.Contains(text, "ai应用") || strings.Contains(text, "aigc") || strings.Contains(text, "agent应用") {
		return dirAIApp
	}
	if strings.Contains(text, "运营开发") || strings.Contains(text, "devops") || strings.Contains(text, "运维") {
		return dirOperation
	}
	if strings.Contains(text, "市场") || strings.Contains(text, "产品") || strings.Contains(text, "商务") || strings.Contains(text, "采购") {
		if strings.Contains(text, "开发") { return dirBackend }
		return dirMarket
	}
	if strings.Contains(text, "硬件") || strings.Contains(text, "嵌入式") {
		return dirClient
	}
	// Fall back to SEP role_direction tags
	for _, t := range tags {
		if t.Category == model.SemanticTagRoleDirection {
			dir := inferDirectionFromText(t.TagLabel + " " + t.TagCode)
			if dir != dirUnknown { return dir }
		}
	}
	return dirUnknown
}

// inferDirectionFromText normalizes text keywords to a career direction.
func inferDirectionFromText(text string) careerDirection {
	t := strings.ToLower(text)
	if strings.Contains(t, "后端") || strings.Contains(t, "后台") || strings.Contains(t, "服务端") || strings.Contains(t, "java后") || strings.Contains(t, "go后") || strings.Contains(t, "java_backend") || strings.Contains(t, "go_backend") || strings.Contains(t, "spring") {
		return dirBackend
	}
	if strings.Contains(t, "全栈") || strings.Contains(t, "fullstack") || strings.Contains(t, "full_stack") {
		return dirFullstack
	}
	if strings.Contains(t, "前端") || strings.Contains(t, "frontend") || strings.Contains(t, "web前") || strings.Contains(t, "vue") || strings.Contains(t, "react") {
		return dirFrontend
	}
	if strings.Contains(t, "客户端") || strings.Contains(t, "pc客户") || strings.Contains(t, "游戏引擎") || strings.Contains(t, "unity") || strings.Contains(t, "cocos") {
		return dirClient
	}
	if strings.Contains(t, "算法") || strings.Contains(t, "机器学习") || strings.Contains(t, "深度学习") || strings.Contains(t, "nlp") || strings.Contains(t, "计算机视觉") || strings.Contains(t, "大模型") || strings.Contains(t, "algorithm") || strings.Contains(t, "ml") || strings.Contains(t, "pytorch") || strings.Contains(t, "tensorflow") {
		return dirAlgorithm
	}
	if strings.Contains(t, "数据工程") || strings.Contains(t, "大数据") || strings.Contains(t, "数据仓库") || strings.Contains(t, "etl") || strings.Contains(t, "spark") || strings.Contains(t, "flink") || strings.Contains(t, "data_eng") {
		return dirData
	}
	if strings.Contains(t, "测试") || strings.Contains(t, "qa") || strings.Contains(t, "test_dev") || strings.Contains(t, "自动化测试") {
		return dirTest
	}
	if strings.Contains(t, "ai应用") || strings.Contains(t, "aigc") || strings.Contains(t, "agent") || strings.Contains(t, "ai_app") {
		return dirAIApp
	}
	if strings.Contains(t, "运营") || strings.Contains(t, "devops") || strings.Contains(t, "运维") || strings.Contains(t, "operation") || strings.Contains(t, "sre") {
		return dirOperation
	}
	if strings.Contains(t, "市场") || strings.Contains(t, "产品") || strings.Contains(t, "商务") || strings.Contains(t, "采购") {
		if strings.Contains(t, "开发") || strings.Contains(t, "技术") { return dirBackend }
		return dirMarket
	}
	return dirUnknown
}

// directionCompatibility returns level, bonus points, and penalty for direction matching.
func directionCompatibility(resumeDir, jobDir careerDirection) (level string, bonus int, penalty int) {
	if resumeDir == dirUnknown || jobDir == dirUnknown { return "unknown", 0, 0 }
	if resumeDir == jobDir { return "strong", 40, 0 }

	// Define proximity groups
	strongMatch := map[careerDirection][]careerDirection{
		dirBackend:   {dirFullstack, dirTest, dirAIApp, dirOperation},
		dirFullstack: {dirBackend, dirFrontend},
		dirFrontend:  {dirFullstack},
		dirTest:      {dirBackend, dirOperation},
		dirAIApp:     {dirBackend, dirAlgorithm},
		dirOperation: {dirBackend, dirTest},
		dirAlgorithm: {dirData, dirAIApp},
		dirData:      {dirAlgorithm, dirBackend},
		dirClient:    {},
		dirProduct:   {},
		dirMarket:    {},
	}
	relatedMatch := map[careerDirection][]careerDirection{
		dirBackend:   {dirData, dirFrontend, dirAlgorithm},
		dirFullstack: {dirTest, dirAIApp},
	}
	mismatch := map[careerDirection][]careerDirection{
		dirBackend: {dirClient, dirMarket, dirProduct, dirAlgorithm},
		dirAlgorithm: {dirBackend, dirClient, dirMarket, dirProduct},
	}

	// Check strong
	for _, d := range strongMatch[resumeDir] {
		if d == jobDir { return "strong", 40, 0 }
	}
	// Check related
	for _, d := range relatedMatch[resumeDir] {
		if d == jobDir { return "related", 22, 0 }
	}
	// Check mismatch
	for _, d := range mismatch[resumeDir] {
		if d == jobDir { return "mismatch", 0, 40 }
	}
	// Default: weak
	return "weak", 0, 10
}

// ─── SEP Candidate & Scoring ──────────────────────────────────────────────────

// SEPMatchScore is the result of matching a resume against a single job via SEP tags.
type SEPMatchScore struct {
	JobID          int64
	Score          int
	MatchedTags    []string // tag_label or explanation
	RiskTags       []string // job tags not matched by resume
	EvidenceTexts  []string // evidence_text from matched resume tags
	DisplayTags    []string // short Chinese labels for display
	RankReason     string
	RoleMatch      int
	TechMatch      int
	EngMatch       int
	ResumeDirection careerDirection
	JobDirection    careerDirection
	DirectionLevel  string
}

// matchResumeJobSEP scores one resume against one job using their SEP tags + career direction gate.
func matchResumeJobSEP(resumeTags, jobTags []*model.SemanticProfileTag, jobTitle, jobDirection, jobCategory string, resumeDir careerDirection) *SEPMatchScore {
	s := &SEPMatchScore{ResumeDirection: resumeDir}
	matchedResumeTags := make([]int, len(resumeTags))

	for _, jt := range jobTags {
		for ri, rt := range resumeTags {
			if matchesJobTag(rt, jt) {
				matchedResumeTags[ri] = 1
				break
			}
		}
	}

	// Collect matched resume tags
	for ri := range resumeTags {
		if matchedResumeTags[ri] == 1 {
			rt := resumeTags[ri]
			label := rt.TagLabel
			if label == "" {
				label = rt.TagCode
			}
			s.MatchedTags = append(s.MatchedTags, label)
			ev := rt.EvidenceText
			if ev == "" {
				ev = rt.Explanation
			}
			if ev != "" {
				s.EvidenceTexts = append(s.EvidenceTexts, truncateStr(ev, 120))
			}
		}
	}

	// Collect risk tags
	for _, jt := range jobTags {
		if jt.Category == model.SemanticTagRiskSignal {
			found := false
			for ri := range resumeTags {
				if matchesJobTag(resumeTags[ri], jt) { found = true; break }
			}
			if !found { s.RiskTags = append(s.RiskTags, jt.TagLabel) }
		}
	}

	// ── M2B Career Direction Gate ──
	jobDir := inferJobCareerDirection(jobTitle, jobDirection, jobCategory, jobTags)
	s.JobDirection = jobDir
	level, _, _ := directionCompatibility(resumeDir, jobDir)
	s.DirectionLevel = level

	// M2B: Scoring with direction gate as PRIMARY factor
	base := 25
	switch level {
	case "strong":
		base += 40
	case "related":
		base += 22
	case "weak":
		base -= 10
	case "mismatch":
		base -= 40
	}

	// technical_skill (secondary)
	techMatch := 0
	for _, jt := range jobTags {
		if jt.Category == model.SemanticTagTechnicalSkill {
			for ri := range resumeTags {
				if matchedResumeTags[ri] == 1 && resumeTags[ri].Category == model.SemanticTagTechnicalSkill {
					techMatch++; break
				}
			}
		}
	}
	s.TechMatch = techMatch
	base += int(math.Min(float64(techMatch*3), 15))

	// engineering_capability
	engMatch := 0
	for _, jt := range jobTags {
		if jt.Category == model.SemanticTagEngineeringCapability {
			for ri := range resumeTags {
				if matchedResumeTags[ri] == 1 && resumeTags[ri].Category == model.SemanticTagEngineeringCapability {
					engMatch++; break
				}
			}
		}
	}
	s.EngMatch = engMatch
	base += int(math.Min(float64(engMatch*3), 12))

	// experience_signal
	expBonus := 0
	for _, jt := range jobTags {
		if jt.Category == model.SemanticTagExperienceSignal {
			for ri := range resumeTags {
				if matchedResumeTags[ri] == 1 && resumeTags[ri].Category == model.SemanticTagExperienceSignal {
					expBonus += 3; break
				}
			}
		}
	}
	base += int(math.Min(float64(expBonus), 8))

	// domain_context
	domainBonus := 0
	for _, jt := range jobTags {
		if jt.Category == model.SemanticTagDomainContext {
			for ri := range resumeTags {
				if matchedResumeTags[ri] == 1 && resumeTags[ri].Category == model.SemanticTagDomainContext {
					domainBonus += 2; break
				}
			}
		}
	}
	base += int(math.Min(float64(domainBonus), 5))

	// risk penalty
	riskPenalty := len(s.RiskTags) * 4
	if riskPenalty > 12 { riskPenalty = 12 }
	base -= riskPenalty

	// Clamp
	if base < 0 { base = 0 }
	if base > 100 { base = 100 }
	s.Score = base

	// M2: DisplayTags
	seen := map[string]bool{}
	for _, m := range s.MatchedTags {
		label := m
		if len([]rune(label)) > 8 {
			if short, ok := fastTagLabels[strings.ToLower(label)]; ok && len([]rune(short)) <= 8 {
				label = short
			}
		}
		if !seen[label] { seen[label] = true; s.DisplayTags = append(s.DisplayTags, label) }
	}
	if len(s.DisplayTags) > 4 { s.DisplayTags = s.DisplayTags[:4] }

	// Rank reason
	if len(s.MatchedTags) > 0 {
		top := s.MatchedTags
		if len(top) > 2 { top = top[:2] }
		s.RankReason = fmt.Sprintf("方向%s · 在 %s 上与简历存在语义匹配", level, strings.Join(top, "、"))
	} else {
		s.RankReason = "该岗位与简历在基础背景上存在一定关联"
	}

	return s
}

// rankSEPCandidates scores and sorts all job candidates against a resume using SEP tags.
func rankSEPCandidates(resumeTags []*model.SemanticProfileTag, candidates []*SEPRecommendCandidate) []*SEPMatchScore {
	resumeDir := inferResumeCareerDirection(resumeTags)
	var scores []*SEPMatchScore
	for _, c := range candidates {
		s := matchResumeJobSEP(resumeTags, c.JobTags, c.JobTitle, c.Direction, c.Category, resumeDir)
		s.JobID = c.JobID
		scores = append(scores, s)
	}
	// Sort by score desc, tie-break: RoleMatch > TechMatch > EngMatch
	for i := 0; i < len(scores)-1; i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[j].Score > scores[i].Score ||
				(scores[j].Score == scores[i].Score && scores[j].RoleMatch > scores[i].RoleMatch) ||
				(scores[j].Score == scores[i].Score && scores[j].RoleMatch == scores[i].RoleMatch && scores[j].TechMatch > scores[i].TechMatch) ||
				(scores[j].Score == scores[i].Score && scores[j].RoleMatch == scores[i].RoleMatch && scores[j].TechMatch == scores[i].TechMatch && scores[j].EngMatch > scores[i].EngMatch) {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}
	return scores
}

// ─── SEP Candidate Types ─────────────────────────────────────────────────────

// SEPRecommendCandidate is a job candidate for SEP-based ranking.
type SEPRecommendCandidate struct {
	JobID     int64
	JobTitle  string
	Direction string
	Category  string
	JobTags   []*model.SemanticProfileTag
}
