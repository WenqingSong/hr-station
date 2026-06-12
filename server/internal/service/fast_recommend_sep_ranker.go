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

// ─── SEP Candidate & Scoring ──────────────────────────────────────────────────

// SEPMatchScore is the result of matching a resume against a single job via SEP tags.
type SEPMatchScore struct {
	JobID         int64
	Score         int
	MatchedTags   []string // tag_label or explanation
	RiskTags      []string // job tags not matched by resume
	EvidenceTexts []string // evidence_text from matched resume tags
	DisplayTags   []string // short Chinese labels for display
	RankReason    string
	RoleMatch     int
	TechMatch     int
	EngMatch      int
}

// matchResumeJobSEP scores one resume against one job using their SEP tags.
func matchResumeJobSEP(resumeTags, jobTags []*model.SemanticProfileTag) *SEPMatchScore {
	s := &SEPMatchScore{}
	matchedResumeTags := make([]int, len(resumeTags)) // indices of matched resume tags

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

	// Collect job risk tags (job tags with category=risk_signal not matched)
	for _, jt := range jobTags {
		if jt.Category == model.SemanticTagRiskSignal {
			found := false
			for ri := range resumeTags {
				if matchesJobTag(resumeTags[ri], jt) {
					found = true
					break
				}
			}
			if !found {
				s.RiskTags = append(s.RiskTags, jt.TagLabel)
			}
		}
	}

	// Scoring by category
	base := 50
	// role_direction
	for _, jt := range jobTags {
		if jt.Category == model.SemanticTagRoleDirection {
			for ri := range resumeTags {
				if matchedResumeTags[ri] == 1 && resumeTags[ri].Category == model.SemanticTagRoleDirection {
					s.RoleMatch++
					break
				}
			}
		}
	}
	base += int(math.Min(float64(s.RoleMatch*18), 25))

	// technical_skill
	for _, jt := range jobTags {
		if jt.Category == model.SemanticTagTechnicalSkill {
			for ri := range resumeTags {
				if matchedResumeTags[ri] == 1 && resumeTags[ri].Category == model.SemanticTagTechnicalSkill {
					s.TechMatch++
					break
				}
			}
		}
	}
	base += int(math.Min(float64(s.TechMatch*5), 20))

	// engineering_capability
	for _, jt := range jobTags {
		if jt.Category == model.SemanticTagEngineeringCapability {
			for ri := range resumeTags {
				if matchedResumeTags[ri] == 1 && resumeTags[ri].Category == model.SemanticTagEngineeringCapability {
					s.EngMatch++
					break
				}
			}
		}
	}
	base += int(math.Min(float64(s.EngMatch*6), 18))

	// domain_context
	for _, jt := range jobTags {
		if jt.Category == model.SemanticTagDomainContext {
			for ri := range resumeTags {
				if matchedResumeTags[ri] == 1 && resumeTags[ri].Category == model.SemanticTagDomainContext {
					base += 3
					break
				}
			}
		}
	}

	// experience_signal
	for _, jt := range jobTags {
		if jt.Category == model.SemanticTagExperienceSignal {
			for ri := range resumeTags {
				if matchedResumeTags[ri] == 1 && resumeTags[ri].Category == model.SemanticTagExperienceSignal {
					base += 4
					break
				}
			}
		}
	}

	// Risk penalty
	riskPenalty := len(s.RiskTags) * 4
	if riskPenalty > 12 {
		riskPenalty = 12
	}
	base -= riskPenalty

	// Clamp
	if base < 0 { base = 0 }
	if base > 100 { base = 100 }
	s.Score = base

	// DisplayTags: 2-4 short Chinese labels from matched
	seen := map[string]bool{}
	for _, m := range s.MatchedTags {
		short := m
		if len(short) > 10 {
			short = short[:10]
		}
		if !seen[short] {
			seen[short] = true
			s.DisplayTags = append(s.DisplayTags, short)
		}
	}
	if len(s.DisplayTags) > 4 {
		s.DisplayTags = s.DisplayTags[:4]
	}

	// Rank reason
	if len(s.MatchedTags) > 0 {
		top := s.MatchedTags
		if len(top) > 2 { top = top[:2] }
		s.RankReason = fmt.Sprintf("该岗位在 %s 上与简历存在语义匹配", strings.Join(top, "、"))
	} else {
		s.RankReason = "该岗位与简历在基础背景上存在一定关联"
	}

	return s
}

// rankSEPCandidates scores and sorts all job candidates against a resume using SEP tags.
func rankSEPCandidates(resumeTags []*model.SemanticProfileTag, candidates []*SEPRecommendCandidate) []*SEPMatchScore {
	var scores []*SEPMatchScore
	for _, c := range candidates {
		s := matchResumeJobSEP(resumeTags, c.JobTags)
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
