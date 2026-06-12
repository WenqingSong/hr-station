// ─── Deep Cascade Prompt Builders ──────────────────────────────────────────

package service

import (
	"fmt"
	"strings"

	"offer-catcher/server/internal/model"
)

// ─── Layer 1: Candidate Discovery Prompt ──────────────────────────────────

func buildLayer1DiscoveryPrompt(resumeID int64, resumeSummary string, resumeTags []*model.SemanticProfileTag, jobSEP map[int64]*gateJobSep) (string, string) {
	sys := `你是校园招聘岗位推荐专家（候选岗位发现阶段）。
你的任务是从给定的岗位池中筛选 10-15 个值得深入分析的候选岗位。
你必须只从给定岗位池中选择岗位，不允许编造岗位或 job_id。
筛选标准：结合候选人的真实经历、岗位方向、岗位要求、风险门槛和投递价值。
输出严格 JSON 格式，不允许 markdown 代码块包裹。`

	// Build resume profile
	resumeProfile := fmt.Sprintf("候选人画像：%s\n\n简历语义标签：\n", resumeSummary)
	for _, t := range resumeTags {
		resumeProfile += fmt.Sprintf("  - [%s] %s (置信度:%.0f%%) %s\n", t.Category, t.TagLabel, t.Confidence*100, truncateStr(t.EvidenceText, 80))
	}

	// Build job pool
	jobPool := fmt.Sprintf("岗位池（共 %d 个岗位）：\n", len(jobSEP))
	for _, j := range jobSEP {
		tags := topTagLabels(j.tags, 4)
		jobPool += fmt.Sprintf("  job_id=%d  %s  [%s/%s]  标签: %s\n", j.profileResult.TargetID, j.title, j.direction, j.category, strings.Join(tags, " / "))
	}

	// Output format
	outputFmt := `请输出 JSON：
{
  "candidates": [
    {
      "job_id": 10001,
      "job_title": "岗位名",
      "candidate_reason": "为什么值得深入分析（完整中文句）",
      "matched_signals": ["匹配信号1", "匹配信号2"],
      "risk_signals": ["风险信号1"],
      "initial_score": 85
    }
  ],
  "summary": "一句话总结筛选结果"
}
要求：candidates 数量 10-15，job_id 必须来自岗位池，candidate_reason 和 matched_signals 必须非空。`

	return sys, fmt.Sprintf("%s\n\n%s\n\n%s", resumeProfile, jobPool, outputFmt)
}

// ─── Layer 2: Semantic Rerank Prompt ──────────────────────────────────────

func buildLayer2RerankPrompt(resumeID int64, resumeSummary string, resumeTags []*model.SemanticProfileTag, l1 *DeepCandidateDiscoveryOutput, jobSEP map[int64]*gateJobSep) (string, string) {
	sys := `你是校园招聘深度推荐排序专家（语义重排序阶段）。
请根据候选人的简历证据和岗位真实要求，从候选岗位中选出最值得投递的 5 个岗位并排序。
排序不能只看技能重合，还要考虑岗位方向一致性、技术能力匹配、工程能力证据、项目/实习经历支撑、风险可控程度和投递价值。
输出严格 JSON 格式。`

	profile := fmt.Sprintf("候选人：%s\n", resumeSummary)
	profile += "标签："
	for _, t := range resumeTags {
		profile += fmt.Sprintf(" [%s]%s", t.Category, t.TagLabel)
	}
	profile += "\n"

	candidates := "候选岗位：\n"
	for _, c := range l1.Candidates {
		j, ok := jobSEP[c.JobID]
		tags := ""
		if ok {
			tags = strings.Join(topTagLabels(j.tags, 3), ", ")
		}
		candidates += fmt.Sprintf("  %d. %s [%s/%s] 初始分:%d 理由:%s 标签:%s\n", c.JobID, c.JobTitle, j.direction, j.category, c.InitialScore, truncateStr(c.CandidateReason, 80), tags)
	}

	outputFmt := `请输出 JSON：
{
  "top_jobs": [{"job_id":10001,"rank":1,"score":88,"fit_level":"优先投递","why_this_rank":"...","core_evidence":["证据1","证据2"],"main_risks":["风险1"],"synthesis_focus":["分析重点1","分析重点2"]}],
  "excluded_jobs": [{"job_id":99999,"reason":"排除原因"}],
  "ranking_rationale": "排序总结"
}
要求：top_jobs 必须正好 5 个，rank 1-5 不重复，fit_level 使用 优先投递/重点关注/可以尝试/谨慎尝试。`

	return sys, fmt.Sprintf("%s\n%s\n%s", profile, candidates, outputFmt)
}

// ─── Layer 3: Per-Job Deep Synthesis Prompt ───────────────────────────────

func buildLayer3SynthesisPrompt(resumeID int64, resumeSummary string, resumeTags []*model.SemanticProfileTag, job DeepRerankedJob, jsep *gateJobSep) (string, string) {
	sys := `你是校园招聘求职顾问和岗位匹配专家（岗位级深度分析阶段）。
你要针对一个具体岗位，判断这名学生为什么适合、风险在哪里、如何投递、简历应该怎么强化。
必须基于简历和 JD 证据，不能虚构经历，不能把岗位要求当成候选人已具备能力。
输出严格 JSON 格式。`

	resume := fmt.Sprintf("## 候选人简历\n%s\n\n简历语义标签：\n", resumeSummary)
	for _, t := range resumeTags {
		resume += fmt.Sprintf("- [%s] %s (%.0f%%) %s\n", t.Category, t.TagLabel, t.Confidence*100, truncateStr(t.EvidenceText, 100))
	}

	jd := fmt.Sprintf("## 目标岗位\njob_id=%d  %s\n方向:%s  类别:%s\n排序理由:%s\n分析重点:%s\n\n岗位描述:%s\n要求:%s",
		jsep.profileResult.TargetID, jsep.title, jsep.direction, jsep.category,
		job.WhyThisRank, strings.Join(job.SynthesisFocus, "；"),
		truncateStr(jsep.description, 2000), truncateStr(jsep.requirements, 1500))

	jd += "\n\n岗位语义标签：\n"
	for _, t := range jsep.tags {
		jd += fmt.Sprintf("- [%s] %s %s\n", t.Category, t.TagLabel, truncateStr(t.EvidenceText, 80))
	}

	outputFmt := `请输出 JSON：
{
  "job_id": 10001,
  "final_score": 85,
  "fit_level": "优先投递",
  "one_sentence_conclusion": "30-90字，明确是否值得投递",
  "why_recommended": ["3-5条，完整中文句，结合岗位要求和简历证据"],
  "resume_evidence": ["2-4条，来自简历事实"],
  "job_requirement_mapping": ["2-4条，岗位要求与简历能力对应"],
  "risk_points": ["1-3条，具体谨慎，不绝对化"],
  "application_strategy": "80-180字投递策略",
  "resume_improvement_hint": "80-180字简历优化建议",
  "display_tags": ["3-5个中文短语"],
  "why_ranked_here": "40-120字说明该岗位在Top5中的位置"
}
要求：所有字段非空，中文完整句，不虚构证据。`

	return sys, fmt.Sprintf("%s\n\n%s\n\n%s", resume, jd, outputFmt)
}

// ─── Layer 4: Cross-Job Calibration Prompt ─────────────────────────────────

func buildLayer4CalibrationPrompt(resumeID int64, resumeSummary string, l2 *DeepSemanticRerankOutput, l3Results []*DeepPerJobSynthesisOutput) (string, string) {
	sys := `你是推荐结果质量校准专家。
检查 5 个岗位推荐结果是否排序合理、分数一致、风险表述一致，是否存在过度承诺或证据不足。
只能在已有 5 个岗位内调整排序和分数，不能新增岗位，不能编造经历。
输出严格 JSON 格式。`

	profile := fmt.Sprintf("候选人：%s\n", resumeSummary)

	jobs := "5个岗位推荐结果：\n"
	for _, l3 := range l3Results {
		jobs += fmt.Sprintf("---\njob_id=%d fit=%s score=%d\n结论:%s\n推荐理由:%s\n风险:%s\n投递策略:%s\n优化建议:%s\n",
			l3.JobID, l3.FitLevel, l3.FinalScore,
			l3.OneSentenceConclusion,
			strings.Join(l3.WhyRecommended, "；"),
			strings.Join(l3.RiskPoints, "；"),
			l3.ApplicationStrategy,
			l3.ResumeImprovementHint,
		)
	}

	outputFmt := `请输出 JSON：
{
  "final_order": [
    {"job_id":10001,"final_rank":1,"calibrated_score":88,"calibration_reason":"调整原因"},
    ...
  ],
  "global_summary": "80-200字全局推荐总结",
  "quality_status": "verified",
  "quality_warnings": []
}
要求：final_order 必须正好 5 个，quality_status 使用 verified/repaired/failed。`

	return sys, fmt.Sprintf("%s\n%s\n%s", profile, jobs, outputFmt)
}

// ─── Helpers ──────────────────────────────────────────────────────────────

func topTagLabels(tags []*model.SemanticProfileTag, n int) []string {
	var r []string
	for _, t := range tags {
		if len(r) >= n { break }
		r = append(r, t.TagLabel)
	}
	return r
}
