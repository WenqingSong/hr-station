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
	sys := `你是校园招聘投递前岗位适配判断专家（逐岗位深度分析阶段）。
你不是在生成通用推荐文案，而是在做"这个岗位值不值得投、怎么投、风险多大"的专业判断。

必须遵守的规则：
1. 只基于简历事实和岗位要求判断，不虚构经历。
2. 不把岗位要求写成候选人已具备能力。
3. 每个字段承担不同作用，不允许复制粘贴同一段话到多个字段。
4. one_sentence_conclusion 不超过 45 字，是一句有判断的结论。
5. job_specific_reason 必须包含"岗位要求-简历证据"对应关系，比 one_sentence_conclusion 更具体，不能相同。
6. resume_evidence 每条必须是候选人简历中的真实事实，不能用岗位要求代替。
7. risk_points 必须具体说明"为什么会成为风险"，禁止"经验不足""需提升"等空话。
8. job_specific_risks 是比 risk_points 更细的岗位级风险；如果没有额外风险，可为空数组但不能和 risk_points 完全相同。
9. application_strategy 必须告诉用户"如何投递该岗位"（80-150字），不能只说"建议投递"。
10. suggested_action 是"马上做什么"的短句（≤25字），如"优先投递""补充项目描述后投递"。
11. application_strategy 和 suggested_action 不能相同。
12. resume_improvement_hint 必须具体到简历模块或表达方式（80-150字），不能和 application_strategy 相同。
13. why_ranked_here 必须说明相对于其他 Top5 岗位的排序原因。
14. candidate_core_strength 是该岗位下候选人最核心的竞争力（≤40字）。
15. fit_confidence_reason 说明为什么给这个分数（≤60字）。
16. display_tags 必须是 3-5 个中文短语，禁止 raw tag code（如 redis/golang/degree_bachelor）。
17. 所有输出为中文。

禁止输出以下内容：
- "该岗位与当前简历存在一定关联"
- "建议进一步确认匹配度"
- "候选人具备岗位所需能力"（除非简历证据明确支持）
- "需要提升综合能力"
- "适合投递相关岗位"
- "该岗位可考虑投递"
- "建议进一步补充"（必须说明补充什么）
- 任何 raw tag code（redis/golang/mysql/docker_k8s/skill_0 等）
- "标签匹配""缺失标签"

输出严格 JSON 格式，不允许 markdown 代码块包裹。`

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
  "one_sentence_conclusion": "≤45字，明确是否值得投递",
  "job_specific_reason": "100-200字，包含岗位要求与简历证据对应关系，比 conclusion 更具体",
  "why_recommended": ["3-5条完整中文句，结合岗位要求和简历证据"],
  "resume_evidence": ["2-4条简历真实事实"],
  "job_requirement_mapping": ["2-4条，格式：岗位要求：xxx；简历证据：xxx；判断：强匹配/部分匹配/待补充"],
  "risk_points": ["1-3条具体风险，说明为什么是风险"],
  "job_specific_risks": ["0-2条比risk_points更细的风险，不能完全相同；如无额外风险可为空数组"],
  "application_strategy": "80-150字，如何投递该岗位",
  "suggested_action": "≤25字马上做什么",
  "resume_improvement_hint": "80-150字简历具体补哪块",
  "display_tags": ["3-5个中文短语"],
  "why_ranked_here": "40-120字相对Top5其他岗位的排序原因",
  "candidate_core_strength": "≤40字核心竞争",
  "fit_confidence_reason": "≤60字为什么给这个分数"
}
要求：所有字符串字段非空（job_specific_risks可为空数组）；one_sentence_conclusion≠job_specific_reason；application_strategy≠suggested_action≠resume_improvement_hint。`

	return sys, fmt.Sprintf("%s\n\n%s\n\n%s", resume, jd, outputFmt)
}

// ─── Layer 4: Cross-Job Calibration Prompt ─────────────────────────────────

func buildLayer4CalibrationPrompt(resumeID int64, resumeSummary string, l2 *DeepSemanticRerankOutput, l3Results []*DeepPerJobSynthesisOutput) (string, string) {
	sys := `你是推荐结果质量校准专家，不是总结器。

你的核心任务：
1. 检查 Top5 排序是否合理（分数是否和推荐理由一致）。
2. 检查每个岗位的风险是否被低估。
3. 检查是否有把岗位要求误写成候选人能力的情况。
4. 给出全局投递策略。
5. 给出候选人整体定位。

必须遵守：
- 只能在已有 5 个岗位内调整排序和分数，不能新增或删除岗位。
- global_summary 必须包含：候选人最适合方向、Top1 为什么排第一、Top5 分布特点、最大共性风险、推荐策略。控制在 150 字以内。
- candidate_positioning 是候选人的全局定位（≤60字）。
- top_recommendation_reason 是 Top1 排第一的核心原因（≤80字）。
- portfolio_risk_summary 是 Top5 共同风险总结（≤100字）。
- delivery_strategy 是整体投递策略（≤120字）。
- quality_notes 是内部校准说明。
- 所有输出为中文，禁止 raw tag code。

global_summary 不允许：
- 泛泛说"候选人与多个岗位匹配"
- 只重复 Top5 岗位名称
- 超过 200 字
- 像系统日志

合格示例：
"候选人当前最适合后台开发方向，其中「软件开发-后台开发方向」排序第一，原因是简历中的后端接口、MySQL/Redis、项目联调和部署经验与岗位核心要求重合度最高。Top5 整体偏向工程开发类岗位；主要短板是大规模系统、高并发和云原生生产经验表达不足，建议投递前补强项目指标和技术边界描述。"

输出严格 JSON 格式，不允许 markdown 代码块包裹。`

	profile := fmt.Sprintf("候选人：%s\n", resumeSummary)

	jobs := "5个岗位推荐结果：\n"
	for _, l3 := range l3Results {
		jobs += fmt.Sprintf("---\njob_id=%d fit=%s score=%d\n结论:%s\n岗位详细理由:%s\n推荐理由:%s\n简历证据:%s\n风险:%s\n投递策略:%s\n优化建议:%s\n",
			l3.JobID, l3.FitLevel, l3.FinalScore,
			l3.OneSentenceConclusion,
			l3.JobSpecificReason,
			strings.Join(l3.WhyRecommended, "；"),
			strings.Join(l3.ResumeEvidence, "；"),
			strings.Join(l3.RiskPoints, "；"),
			l3.ApplicationStrategy,
			l3.ResumeImprovementHint,
		)
	}

	outputFmt := `请输出 JSON：
{
  "final_order": [
    {"job_id":10001,"final_rank":1,"calibrated_score":88,"calibration_reason":"排名调整原因"},
    {"job_id":10002,"final_rank":2,"calibrated_score":82,"calibration_reason":"排名调整原因"},
    {"job_id":10003,"final_rank":3,"calibrated_score":80,"calibration_reason":"排名调整原因"},
    {"job_id":10004,"final_rank":4,"calibrated_score":75,"calibration_reason":"排名调整原因"},
    {"job_id":10005,"final_rank":5,"calibrated_score":72,"calibration_reason":"排名调整原因"}
  ],
  "global_summary": "120-180字全局推荐总结（含方向、Top1原因、分布、共性风险和策略）",
  "candidate_positioning": "≤60字候选人全局定位",
  "top_recommendation_reason": "≤80字Top1核心原因",
  "portfolio_risk_summary": "≤100字Top5共同风险",
  "delivery_strategy": "≤120字整体投递策略",
  "quality_status": "verified",
  "quality_warnings": [],
  "quality_notes": ["校准说明1","校准说明2"]
}
要求：final_order 必须正好 5 个；quality_status 使用 verified/repaired/failed；所有文本字段非空。`

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
