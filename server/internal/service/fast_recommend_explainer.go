// ─── Fast Recommend Explainer ────────────────────────────────────────────────
//
// Fast 推荐的可解释性模块。
// 职责：将 raw tag code 转换为用户可读的中文解释。
// 不调用 LLM，纯规则映射。
// 不影响 Fast Ranker 的排序结果。

package service

import (
	"fmt"
	"math"
	"strings"

	"offer-catcher/server/internal/model"
)

// ─── Tag Label Mapping ──────────────────────────────────────────────────────

var fastTagLabels = map[string]string{
	// 方向标签
	"research":  "研究方向相关",
	"algorithm": "算法能力",
	"backend":   "后端开发",
	"frontend":  "前端开发",
	"data":      "数据分析",
	"product":   "产品方向",
	"ai_infra":  "AI 基础设施",
	// 技能标签
	"redis":      "Redis 工程能力",
	"golang":     "Go 语言开发",
	"python":     "Python 开发",
	"java":       "Java 开发",
	"cpp":        "C/C++ 开发",
	"mysql":      "MySQL 数据库",
	"linux":      "Linux 系统",
	"docker_k8s": "容器化与编排",
	"pytorch_tf": "深度学习框架",
	// 要求标签
	"degree_bachelor": "本科学历要求",
	"degree_master":   "硕士学历要求",
	"internship_exp":  "实习经历",
	"project_exp":     "项目经验",
	"paper":           "论文或学术成果",
}

// ─── Tag → Highlight Sentences ──────────────────────────────────────────────

var fastHighlightSentences = map[string]string{
	"research":  "岗位具备研究/算法方向特征，适合有算法探索经历的候选人尝试。",
	"algorithm": "算法基础与岗位能力要求存在一定重合。",
	"backend":   "岗位方向与后端开发背景较匹配。",
	"frontend":  "前端开发技能与岗位技术栈存在关联。",
	"data":      "数据分析能力与岗位职责有一定契合度。",
	"redis":     "简历与岗位在 Redis / 缓存相关能力上存在匹配。",
	"golang":    "Go 语言开发经历与岗位技术栈有较强关联。",
	"python":    "Python 开发经验与岗位需求存在交集。",
	"java":      "Java 开发背景与岗位后端要求相关联。",
	"cpp":       "C/C++ 工程能力与岗位底层技术需求有一定匹配。",
	"mysql":     "数据库能力与岗位后端工程要求相关。",
	"linux":     "Linux 系统经验有助于支撑后端开发环境要求。",
	"docker_k8s": "容器化与部署经验有助于支撑工程落地能力。",
	"pytorch_tf": "深度学习框架经验与岗位 AI/算法要求相匹配。",
}

// ─── Risk Reminder Messages ───────────────────────────────────────────────

var fastRiskMessages = map[string]string{
	"degree_bachelor": "岗位可能要求本科及以上学历，请确认简历中的学历信息清晰完整。",
	"degree_master":   "岗位可能更偏好研究生背景，如投递需突出项目和工程能力补足学历差异。",
	"internship_exp":  "岗位可能更看重实习经历，建议在简历中突出真实实习或项目落地证据。",
	"project_exp":     "岗位关注项目经验，建议补充项目目标、技术方案和结果数据。",
	"paper":           "岗位可能关注论文或研究成果，若没有相关成果，建议强化实验、算法或工程实践证明。",
}

// ─── Explainer Entry ────────────────────────────────────────────────────────

// buildFastRecommendReasonExplainer replaces buildFastRecommendReason.
// Generates a user-readable one-sentence recommendation reason.
func buildFastRecommendReasonExplainer(c model.JobCandidate) string {
	if len(c.MatchedTags) > 0 {
		labels := tagCodesToLabels(c.MatchedTags, 3)
		if len(labels) > 0 {
			return fmt.Sprintf("标签匹配：%s", strings.Join(labels, "、"))
		}
	}
	return "该岗位可能与你的简历方向存在潜在匹配"
}

// buildFastHighlights converts matched tag codes into readable highlight sentences.
func buildFastHighlights(c model.JobCandidate) []string {
	return buildHumanReadableList(c.MatchedTags, fastHighlightSentences, 3, "该岗位与当前简历在基础背景上存在一定关联，建议进一步做岗位匹配分析确认。")
}

// buildFastRiskReminders converts missing tag codes into risk reminder sentences.
func buildFastRiskReminders(c model.JobCandidate) []string {
	return buildHumanReadableList(c.MissingTags, fastRiskMessages, 2, "当前未发现明显硬性风险，仍建议结合岗位详情进一步确认。")
}

// buildFastOneSentenceReason generates a user-readable one-sentence reason based on score.
func buildFastOneSentenceReason(c model.JobCandidate) string {
	scaledScore := int(math.Min(95, math.Max(70, 70 + c.TagScore*0.5)))
	matchedLabels := tagCodesToLabels(c.MatchedTags, 2)
	coreMatch := strings.Join(matchedLabels, "、")
	if coreMatch == "" {
		coreMatch = "技能背景"
	}

	switch {
	case scaledScore >= 85:
		return fmt.Sprintf("你的简历与该岗位在 %s 上重合度较高，适合作为优先投递岗位。", coreMatch)
	case scaledScore >= 70:
		return fmt.Sprintf("你的简历与该岗位在 %s 上存在较明确的能力交集，适合作为重点关注岗位。", coreMatch)
	default:
		return fmt.Sprintf("该岗位与当前简历在 %s 上存在一定关联，建议先补强相关经历后再投递。", coreMatch)
	}
}

// buildFastSuggestedAction generates a reasonable next-step suggestion based on score.
func buildFastSuggestedAction(c model.JobCandidate) string {
	scaledScore := int(math.Min(95, math.Max(70, 70 + c.TagScore*0.5)))
	switch {
	case scaledScore >= 85:
		return "建议进入岗位详情确认要求，再进行岗位匹配分析，判断是否适合重点投递。"
	case scaledScore >= 70:
		return "建议先做岗位匹配分析，确认风险点后再决定是否投递。"
	default:
		return "建议优先补充相关项目或实习证据，再考虑投递该岗位。"
	}
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

// tagCodesToLabels converts raw tag codes to Chinese display labels.
func tagCodesToLabels(codes []string, max int) []string {
	var result []string
	for _, c := range codes {
		if len(result) >= max {
			break
		}
		if label, ok := fastTagLabels[c]; ok {
			result = append(result, label)
		} else {
			// Unknown code: skip it rather than leaking raw code
			continue
		}
	}
	return result
}

// buildHumanReadableList converts a list of raw tag codes into readable sentences.
// Uses the provided sentenceMap. Falls back to defaultMsg if no match.
func buildHumanReadableList(codes []string, sentenceMap map[string]string, max int, defaultMsg string) []string {
	var result []string
	for _, code := range codes {
		if len(result) >= max {
			break
		}
		if sentence, ok := sentenceMap[code]; ok {
			result = append(result, sentence)
		}
	}
	if len(result) == 0 {
		return []string{defaultMsg}
	}
	return result
}
