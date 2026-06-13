// ─── H19-B: Deep Match Prompt ────────────────────────────────────────────

package service

import (
	"fmt"
	"strings"

	"offer-catcher/server/internal/model"
)

const deepMatchSystemPrompt = `你是岗位匹配深度分析智能体（Deep Match Agent），不是简历夸奖助手。

你必须基于以下输入，输出结构化 DeepMatchReport JSON：
1. 简历原文
2. 岗位 JD 原文
3. Resume SEP 语义画像（标签+证据+方向）
4. Job SEP 语义画像（标签+要求+方向）
5. Fast Match 快速初判结果

核心规则：
- 所有优势必须有证据支撑，不允许编造。
- 所有风险必须解释影响和修复方式。
- 所有数组没有内容时输出 []，禁止 null。
- 所有对象没有内容时输出空字段结构，禁止 null。
- 不允许输出泛泛建议（如"继续提升能力"）。
- 不允许输出 raw tag code。
- 严格输出 JSON，不要 Markdown，不要代码块，不要解释。

你需要按 8 层 Agent 逻辑完成分析：
L1 JD 深度解构 → L2 简历证据审计 → L3 要求对齐 → L4 风险诊断 → L5 投递策略 → L6 简历改写 → L7 面试追问 → L8 全局校准

输出 JSON Schema：
{
  "generation_tier": "quality",
  "generated_by": "deep_match_agent",
  "agent_status": "match_verified",
  "match_score": int(0-100),
  "match_level": string(高度匹配|较高匹配|中等匹配|弱匹配|不推荐),
  "application_decision": string(priority_apply|apply_directly|revise_before_apply|backup_apply|not_recommended),
  "decision_label": string(优先投递|可以直接投递|修改后投递|作为备选投递|暂不建议投递),
  "one_sentence_conclusion": string(一句话说明匹配情况),
  "score_breakdown": { "direction_fit": 0, "must_have_coverage": 0, "evidence_strength": 0, "experience_relevance": 0, "risk_control": 0 },
  "job_understanding": {
    "job_positioning": string(这个岗位真正在招什么样的人),
    "core_responsibilities": [string],
    "must_have_requirements": [string],
    "important_requirements": [string],
    "bonus_requirements": [string],
    "hidden_expectations": [string]
  },
  "requirement_coverage": [{"requirement_name":"","requirement_type":"must_have|should_have|nice_to_have","coverage_status":"strong|partial|weak|missing","coverage_score":0,"matched_evidence":[],"gap":"","suggestion":""}],
  "resume_evidence": [{"evidence_name":"","evidence_level":0,"source_type":"","source_title":"","evidence_text":"","supports_requirement":"","confidence":"high|medium|low"}],
  "core_match_points": [{"point_name":"","match_strength":"strong|partial|transferable|weak","why_it_matches":"","supporting_evidence":[],"job_requirement":""}],
  "key_risks": [{"risk_name":"","risk_level":"high|medium|low","why_it_matters":"","evidence_gap":"","fix_suggestion":"","interview_impact":""}],
  "resume_rewrite_strategy": {"skills_reorder":[],"project_rewrite_points":[],"internship_rewrite_points":[],"keywords_to_add":[],"content_to_weaken":[]},
  "application_strategy": {"best_timing":"","delivery_strategy":"","competing_positioning":""},
  "interview_risk_questions": [{"question":"","why_asked":"","answer_direction":"","related_risk":""}],
  "agent_trace": [{"stage":"","agent":"","status":"done","summary":""}]
}`

func buildDeepMatchPrompt(ctx *DeepMatchContext, fastResult *model.MatchAnalysisResult) (string, string) {
	_ = fastResult
	var sb strings.Builder

	sb.WriteString("【简历原文】\n")
	if ctx.ResumeRaw != "" {
		sb.WriteString(truncDeepStr(ctx.ResumeRaw, 3000))
	} else {
		sb.WriteString("(未提供)")
	}

	sb.WriteString("\n\n【岗位 JD 原文】\n")
	if ctx.JobRaw != "" {
		sb.WriteString(truncDeepStr(ctx.JobRaw, 4000))
	} else {
		sb.WriteString("(未提供)")
	}

	sb.WriteString("\n\n【Resume SEP 语义画像】\n")
	if ctx.ResumeSEP != nil && len(ctx.ResumeSEP.ResumeTags) > 0 {
		for _, t := range ctx.ResumeSEP.ResumeTags {
			sb.WriteString(fmt.Sprintf("- %s [%s] confidence=%.2f evidence=%s\n", t.TagLabel, t.Category, t.Confidence, truncDeepStr(t.EvidenceText, 200)))
		}
	} else {
		sb.WriteString("(SEP 画像不可用)")
	}

	sb.WriteString("\n【Job SEP 语义画像】\n")
	if ctx.JobSEP != nil && len(ctx.JobSEP.JobTags) > 0 {
		sb.WriteString(fmt.Sprintf("岗位名称: %s\n", ctx.JobSEP.JobTitle))
		for _, t := range ctx.JobSEP.JobTags {
			sb.WriteString(fmt.Sprintf("- %s [%s] confidence=%.2f\n", t.TagLabel, t.Category, t.Confidence))
		}
	} else {
		sb.WriteString("(SEP 画像不可用)")
	}

	sb.WriteString("\n【Fast Match 初判结果】\n")
	if fastResult != nil {
		sb.WriteString(fmt.Sprintf("score=%d level=%s\n", fastResult.MatchScore, fastResult.MatchLevel))
		sb.WriteString(fmt.Sprintf("highlights=%v\n", fastResult.MatchHighlights))
		sb.WriteString(fmt.Sprintf("gaps=%v\n", fastResult.GapPoints))
	} else {
		sb.WriteString("(无 Fast 结果)")
	}

	sb.WriteString("\n\n请基于以上输入，输出完整 DeepMatchReport JSON。")

	return deepMatchSystemPrompt, sb.String()
}

func truncDeepStr(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}
