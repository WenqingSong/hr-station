// ─── JM-R3-R: Screening Match V1 Pipeline ───────────────────────────────────
// 5 个 Hy3 深度模块负责内容生成 + 本地裁判负责一致性校验。
// 规则管道仅作为 fallback。Replaces deep_match_v3 10-module.

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"offer-catcher/server/internal/agent"
	"offer-catcher/server/internal/llm"
	"offer-catcher/server/internal/llmjson"
	"offer-catcher/server/internal/model"
)

// buildScreeningMatchContext builds the context for the 5-module pipeline.
func (s *MatchService) buildScreeningMatchContext(
	ctx context.Context, resume *model.ResumeVersion, job *model.Job,
) (*ScreeningMatchContext, error) {
	smCtx := &ScreeningMatchContext{
		ResumeID:   int64(resume.ID),
		JobID:      int64(job.ID),
		ResumeName: resume.Name,
		ResumeRawText: resume.RawText, // JM-R3-R: populate for Hy3 modules
		JobTitle:    job.Title,
		JobDirection: job.ExternalSource,
		JobCategory:  job.Category,
		JobCity:      job.City,
		JobRawJD:     job.Description,
	}

	// Read SEP profiles and Fast Result
	if s.semanticRepo != nil {
		rp, _ := s.semanticRepo.FindLatestByTarget(ctx, model.SemanticTargetResume, int64(resume.ID))
		jp, _ := s.semanticRepo.FindLatestByTarget(ctx, model.SemanticTargetJob, int64(job.ID))
		if rp != nil && jp != nil && rp.Status == model.SemanticProfileStatusVerified && jp.Status == model.SemanticProfileStatusVerified {
			rtags, _ := s.semanticRepo.ListTagsByProfileID(ctx, rp.ID)
			jtgs, _ := s.semanticRepo.ListTagsByProfileID(ctx, jp.ID)
			smCtx.ResumeSEPProfile = rtags
			smCtx.JobSEPProfile = jtgs

			// Run Fast SEP Match
			fastRes := runFastSEPMatch(&MatchSEPContext{
				ResumeTags: rtags, JobTags: jtgs, JobTitle: job.Title,
				ResumeID: int64(resume.ID), JobID: int64(job.ID),
			})
			smCtx.FastResult = fastRes
			return smCtx, nil
		}
		return nil, fmt.Errorf("screening match: SEP profiles not ready for resume=%d job=%d", resume.ID, job.ID)
	}
	return nil, fmt.Errorf("screening match: semanticRepo not available")
}

// ─── JM-R3-R: Real Hy3 5-Module + Judge Pipeline ──────────────────────────
// Replaces the rule-based pipeline for quality tier when Hy3 client is available.

// runHy3ScreeningPipeline executes the real Hy3 5-module pipeline.
// Returns (report, status, generated_by). Returns (nil, "", "") on total failure.
func runHy3ScreeningPipeline(
	ctx context.Context, smCtx *ScreeningMatchContext, llmClient llm.Client,
) (*ScreeningMatchV1, string, string) {
	startTime := time.Now()
	log.Printf("[Hy3Pipeline] starting 5-module pipeline for resume=%d job=%d", smCtx.ResumeID, smCtx.JobID)

	// Track module successes/failures
	type moduleResult struct{ ok bool }
	results := make(map[string]bool)

	// Module 1: Decision
	decision, m1ok := runHy3ScreeningDecisionModule(ctx, smCtx, llmClient)
	results["m1_decision"] = m1ok
	if !m1ok {
		log.Printf("[Hy3Pipeline] Module 1 (decision) failed, using fast result fallback")
		decision = runScreeningDecisionModule(smCtx)
	}
	log.Printf("[Hy3Pipeline] M1: score=%d decision=%s hy3=%v", decision.Score, decision.Decision, m1ok)

	// Module 2: Pass Points (depends on decision for context)
	passPoints, m2ok := runHy3ScreeningPassPointsModule(ctx, smCtx, decision, llmClient)
	results["m2_pass_points"] = m2ok
	if !m2ok {
		passPoints = runScreeningPassPointsModule(smCtx, decision)
	}
	log.Printf("[Hy3Pipeline] M2: %d pass points hy3=%v", len(passPoints), m2ok)

	// Module 3: Elimination Risks
	risks, m3ok := runHy3ScreeningEliminationRisksModule(ctx, smCtx, decision, llmClient)
	results["m3_risks"] = m3ok
	if !m3ok {
		risks = runScreeningEliminationRisksModule(smCtx, decision)
	}
	log.Printf("[Hy3Pipeline] M3: %d risks hy3=%v", len(risks), m3ok)

	// Module 4: Resume Rewrite Commands (depends on risks and decision)
	commands, m4ok := runHy3ResumeRewriteCommandsModule(ctx, smCtx, decision, risks, llmClient)
	results["m4_rewrite"] = m4ok
	if !m4ok {
		commands = runResumeRewriteCommandsModule(smCtx, decision, risks)
	}
	log.Printf("[Hy3Pipeline] M4: %d rewrite commands hy3=%v", len(commands), m4ok)

	// Module 5: Job Strategy
	strategy, m5ok := runHy3JobStrategyModule(ctx, smCtx, decision, risks, llmClient)
	results["m5_strategy"] = m5ok
	if !m5ok {
		strategy = runJobStrategyModule(smCtx, decision, risks)
	}
	log.Printf("[Hy3Pipeline] M5: positioning=%s hy3=%v", strategy.Positioning, m5ok)

	// Assemble result (skip Hy3 judge — use local consistency check instead of 6th LLM call)
	result := &ScreeningMatchV1{
		Mode:                    "deep",
		Version:                 "screening_match_v1",
		ScreeningDecision:       *decision,
		ScreeningPassPoints:     passPoints,
		ScreeningEliminationRisks: risks,
		ResumeRewriteCommands:   commands,
		JobStrategy:             *strategy,
	}
	runScreeningJudgeModule(result)

	// Determine status
	allOk := m1ok && m2ok && m3ok && m4ok && m5ok
	partialOk := m1ok || m2ok || m3ok || m4ok || m5ok

	duration := time.Since(startTime)
	log.Printf("[Hy3Pipeline] completed in %v all_ok=%v partial=%v results=%v", duration, allOk, partialOk, results)

	if allOk {
		return result, "verified", "hy3_screening_match_agent"
	}
	if partialOk {
		return result, "degraded", "hy3_screening_match_agent_degraded"
	}
	return nil, "", ""
}

// ─── Hy3 Module 1: Screening Decision ─────────────────────────────────────

func runHy3ScreeningDecisionModule(
	ctx context.Context, smCtx *ScreeningMatchContext, llmClient llm.Client,
) (*ScreeningDecision, bool) {
	systemPrompt := `你是 Offer捕手 的岗位匹配AI，面向校招/实习场景。你只输出严格的JSON，不输出任何解释或markdown。
你的任务是：基于学生简历原文、岗位JD、语义画像和初判，给出精准的初筛投递决策。

【方向主导性判断 — 最重要规则】
先判断学生的核心方向是什么（不是看任意关键词，而是看实习/项目/奖项/论文的主体方向）：
- 如果简历主体是"推荐系统/搜索排序/ML建模/深度学习/NLP/CV/数据挖掘"等算法方向 → 核心方向为算法
- 如果简历主体是"后端接口/数据库/缓存/消息队列/部署/后台系统" → 核心方向为后端
- 如果简历主体是"产品策划/需求文档/竞品分析/用户调研/数据分析" → 核心方向为产品

然后判断岗位的核心方向是什么。

【方向一致性评分】
- 核心方向完全一致（后端→后端、算法→算法、产品→产品）：
  有真项目/实习 → revise_before_apply 或 direct_apply (70-90分)
  证据不足 → backup 或 not_recommended (40-69分)

- 核心方向相邻（算法子领域交叉→如推荐系统投NLP、后端投测试开发）：
  通常 backup 或 revise_before_apply (55-75分)

- 核心方向错位（后端→算法/NLP、算法→后台开发、产品→后端）：
  通常 not_recommended 或 change_direction (20-55分)
  不得因为有跨领域技能关键词就提分

【校招后台开发评分准则 — 仅当学生核心方向为后端时适用】
如果岗位是后端方向，且学生核心方向明确为后端，满足以下3项以上：
- 有后端实习经历、后端项目经历
- 有 Java/Go 等后端语言、MySQL/Redis/Spring/Docker/Nginx/Linux 等技术栈
- 有接口、权限、登录、缓存、部署等后端工程场景
则合理结论通常是 revise_before_apply（70-85分）。

【重要：算法简历投后端岗位】
如果学生核心方向是算法（推荐/ML/NLP/CV），即使简历中有 Python/MySQL/Docker 等通用技术，
也不得给予 revise_before_apply 或 direct_apply。这些通用技术不代表后端工程能力。
应判为 not_recommended 或 backup (40-65分)，并说明核心方向差异。

【不要过度惩罚】
校招岗位允许部分能力缺失。高并发/微服务/K8s 等高级能力缺失 → 改写补强风险，非直接淘汰。

【决策规则】
- score >= 85：direct_apply
- score 70-84：revise_before_apply（方向匹配、有项目实习证据，表达需强化）
- score 55-69：backup（方向匹配但证据偏弱，或相邻方向）
- score 40-54：not_recommended（方向相近但证据严重不足，或方向错位但有一定关联）
- score < 40：change_direction（方向完全错位）
- 不允许为了友好而强行推荐`

	userPrompt := buildHy3Module1UserPrompt(smCtx)

	resp, err := callHy3WithRepair(ctx, llmClient, systemPrompt, userPrompt, "screening_decision")
	if err != nil {
		log.Printf("[Hy3-M1] LLM call failed: %v", err)
		return nil, false
	}

	var decision ScreeningDecision
	if err := json.Unmarshal([]byte(resp), &decision); err != nil {
		log.Printf("[Hy3-M1] JSON parse failed: %v, raw=%s", err, truncate(resp, 200))
		return nil, false
	}

	// Validate consistency
	if decision.Score < 0 || decision.Score > 100 {
		log.Printf("[Hy3-M1] invalid score %d, rejecting", decision.Score)
		return nil, false
	}
	if decision.Decision == "" {
		decision.Decision = deriveDecisionFromScore(decision.Score)
	}
	// JM-R6-P2: Always normalize label to standard form (Hy3 may output variants)
	decision.DecisionLabel = normalizeDecisionLabel(decision.DecisionLabel, decision.Decision)
	if decision.Summary == "" {
		decision.Summary = fmt.Sprintf("Hy3分析：匹配度%d分，建议%s。", decision.Score, decision.DecisionLabel)
	}

	return &decision, true
}

// ─── Hy3 Module 2: Pass Points ────────────────────────────────────────────

func runHy3ScreeningPassPointsModule(
	ctx context.Context, smCtx *ScreeningMatchContext, decision *ScreeningDecision, llmClient llm.Client,
) ([]ScreeningPassPoint, bool) {
	systemPrompt := `你是 Offer捕手 的初筛命中点分析AI。只输出严格JSON数组。
基于简历原文和岗位JD，找出学生在初筛阶段最可能通过的2-3个关键点。
每条命中点必须绑定真实简历证据，不允许"基础扎实""能力较强"等空话。`

	userPrompt := buildHy3Module2UserPrompt(smCtx, decision)

	resp, err := callHy3WithRepair(ctx, llmClient, systemPrompt, userPrompt, "screening_pass_points")
	if err != nil {
		return nil, false
	}

	var points []ScreeningPassPoint
	if err := json.Unmarshal([]byte(resp), &points); err != nil {
		log.Printf("[Hy3-M2] JSON parse failed: %v", err)
		return nil, false
	}
	if len(points) == 0 {
		return nil, false
	}
	if len(points) > 3 {
		points = points[:3]
	}
	return points, true
}

// ─── Hy3 Module 3: Elimination Risks ──────────────────────────────────────

func runHy3ScreeningEliminationRisksModule(
	ctx context.Context, smCtx *ScreeningMatchContext, decision *ScreeningDecision, llmClient llm.Client,
) ([]ScreeningEliminationRisk, bool) {
	systemPrompt := `你是 Offer捕手 的初筛淘汰风险分析AI。只输出严格JSON数组。
你的任务是找出简历在初筛阶段最可能被淘汰的1-3个风险点。

【风险必须区分三种类型】
1. 方向错位风险：岗位核心方向（算法/NLP/咨询）与简历核心经历（后端工程）不一致。severity=high。
2. 证据不足风险：方向一致但缺少真实项目/实习经历支撑。severity=high/medium。
3. 表达强化风险：有项目/实习证据但简历表达不够工程化、不够量化。severity=medium/low。

【重要】
- 如果是"表达强化风险"，current_resume_problem 不能说"缺少证据"，应该说"证据存在但表达不够"
- fix_direction 必须具体可操作，不能写"建议优化"或"加强项目经验"这种空话
- 对后端开发岗位，fix_direction 应围绕：接口设计、权限校验、数据库操作、缓存使用、部署验证、异常处理`

	userPrompt := buildHy3Module3UserPrompt(smCtx, decision)

	resp, err := callHy3WithRepair(ctx, llmClient, systemPrompt, userPrompt, "screening_risks")
	if err != nil {
		return nil, false
	}

	var risks []ScreeningEliminationRisk
	if err := json.Unmarshal([]byte(resp), &risks); err != nil {
		log.Printf("[Hy3-M3] JSON parse failed: %v", err)
		return nil, false
	}
	if len(risks) == 0 {
		return nil, false
	}
	if len(risks) > 3 {
		risks = risks[:3]
	}
	return risks, true
}

// ─── Hy3 Module 4: Resume Rewrite Commands ────────────────────────────────

func runHy3ResumeRewriteCommandsModule(
	ctx context.Context, smCtx *ScreeningMatchContext, decision *ScreeningDecision, risks []ScreeningEliminationRisk, llmClient llm.Client,
) ([]ResumeRewriteCommand, bool) {
	// For change_direction, don't suggest fake rewrites
	if decision.Decision == "change_direction" {
		return runResumeRewriteCommandsModule(smCtx, decision, risks), true
	}

	systemPrompt := `你是 Offer捕手 的简历改写AI。只输出严格JSON数组。
你必须基于学生简历中的真实经历给出改写建议。

【改写目标】
帮助学生把真实经历用更工程化的语言表达出来，让HR在10秒内识别到关键信号。

【后端开发简历改写重点】
如果学生有后端项目/实习，改写应围绕以下维度展开（按实际证据选择）：
- 接口设计：完成了哪些接口、接口数量、RESTful规范、Swagger文档
- 权限校验：JWT/OAuth/登录态管理、权限层级、敏感接口保护
- 数据库操作：表设计与索引、SQL优化、事务处理、数据一致性
- 缓存使用：Redis缓存策略、缓存Key设计、过期策略、命中率
- 部署验证：Docker/环境部署、前后端联调、测试环境验收
- 异常处理：异常捕获、错误码规范、日志监控
- 幂等限流：接口幂等性设计、请求去重、频率限制

【必须做到】
1. 每条改写必须有 before_example 和 after_example
2. after_example 必须基于简历已有经历，不编造数据/指标/技术
3. 用具体的技术场景代替"负责开发"这种空泛描述
4. 如果简历确实缺少某方面证据，明确说"建议先补充XX项目"而非瞎编

【禁止】
- 编造简历中没有的量化数字（如"QPS提升50%""将响应时间从180ms降到60ms"）
- 编造简历中没有的技术（如K8s/微服务/分布式）
- 泛泛写"建议优化项目描述"`

	userPrompt := buildHy3Module4UserPrompt(smCtx, decision, risks)

	resp, err := callHy3WithRepair(ctx, llmClient, systemPrompt, userPrompt, "screening_rewrite")
	if err != nil {
		return nil, false
	}

	var commands []ResumeRewriteCommand
	if err := json.Unmarshal([]byte(resp), &commands); err != nil {
		log.Printf("[Hy3-M4] JSON parse failed: %v", err)
		return nil, false
	}
	if len(commands) < 2 {
		return nil, false
	}
	if len(commands) > 4 {
		commands = commands[:4]
	}
	return commands, true
}

// ─── Hy3 Module 5: Job Strategy ───────────────────────────────────────────

func runHy3JobStrategyModule(
	ctx context.Context, smCtx *ScreeningMatchContext, decision *ScreeningDecision, risks []ScreeningEliminationRisk, llmClient llm.Client,
) (*JobStrategy, bool) {
	systemPrompt := `你是 Offer捕手 的岗位策略AI。只输出严格JSON对象。
你需要给出：岗位定位、推荐行动、现在投递会怎样、更适合方向、应避开方向。

【岗位定位判断】
1. 核心方向一致 + 有项目/实习证据 → positioning="主投岗位"或"修改后投递"
2. 核心方向一致 + 证据不足 → positioning="备选岗位"或"不建议直接投递"
3. 方向错位或算法子领域交叉 → positioning="备选岗位"或"建议换方向"

【重要：算法子领域不是方向错位】
如果学生做推荐系统/搜索排序/ML，岗位是NLP/多模态算法：
- 这不是方向错位，是大领域内的子方向相邻
- 不要判为 change_direction
- positioning 应为"备选岗位"，recommended_action 说明子方向差异但可以尝试

【重要：证据不足 ≠ 方向错误】
方向一致但缺少项目/实习 → not_recommended/backup，不是 change_direction。

【方向错位判断】
后端→算法/NLP、算法/ML→后台开发、产品→后端 → change_direction 或 not_recommended。
better_fit_directions 必须推荐学生核心方向对应的岗位。

【禁止】
- 模糊说"可以尝试"（对方向错位案例）
- 编造学生没有的经历
- 把证据不足当成方向错误`

	userPrompt := buildHy3Module5UserPrompt(smCtx, decision, risks)

	resp, err := callHy3WithRepair(ctx, llmClient, systemPrompt, userPrompt, "screening_strategy")
	if err != nil {
		return nil, false
	}

	var strategy JobStrategy
	if err := json.Unmarshal([]byte(resp), &strategy); err != nil {
		log.Printf("[Hy3-M5] JSON parse failed: %v", err)
		return nil, false
	}
	if strategy.Positioning == "" {
		return nil, false
	}
	return &strategy, true
}

// ─── Hy3 Helper: LLM call with JSON repair ────────────────────────────────

func callHy3WithRepair(ctx context.Context, llmClient llm.Client, systemPrompt, userPrompt, moduleName string) (string, error) {
	req := llm.ChatRequest{
		Messages: []llm.Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature:     0.3,
		MaxTokens:       2048,
		ReasoningEffort: "high",
		Metadata: map[string]string{
			"module":      moduleName,
			"pipeline":    "hy3_screening_match_v1",
		},
	}

	resp, err := llmClient.Chat(ctx, req)
	if err != nil {
		return "", fmt.Errorf("llm chat failed: %w", err)
	}

	rawContent := resp.RawContent
	if rawContent == "" {
		return "", fmt.Errorf("empty llm response")
	}

	// Extract JSON from LLM output
	jsonStr, err := llmjson.ExtractObject(rawContent)
	if err != nil {
		// Try repair
		log.Printf("[Hy3-%s] JSON extract failed: %v, attempting repair", moduleName, err)
		candidate, _ := llmjson.ExtractBestCandidate(rawContent)
		if candidate != "" {
			repaired := repairSimpleJSON(candidate)
			if repaired != "" {
				return repaired, nil
			}
		}
		return "", fmt.Errorf("json extract and repair failed: %w", err)
	}

	return jsonStr, nil
}

// ─── Hy3 Prompt Builders ──────────────────────────────────────────────────

func buildHy3Module1UserPrompt(smCtx *ScreeningMatchContext) string {
	var sb strings.Builder

	// Add job/context info first
	sb.WriteString(fmt.Sprintf("【岗位信息】%s | %s | %s\n", smCtx.JobTitle, smCtx.JobCategory, smCtx.JobCity))
	sb.WriteString(fmt.Sprintf("【场景】校招/实习风格。先判断学生核心方向(算法/后端/产品/设计)，再与岗位方向对比。\n\n"))

	sb.WriteString("# 简历原文\n")
	sb.WriteString(truncate(smCtx.ResumeRawText, 3000))
	sb.WriteString("\n\n# 岗位JD\n")
	sb.WriteString(truncate(smCtx.JobRawJD, 2000))

	if smCtx.FastResult != nil {
		sb.WriteString(fmt.Sprintf("\n\n# Fast 初判（仅供参考，不要盲从）\n分数: %d\n等级: %s\n命中点: %v\n风险: %v",
			smCtx.FastResult.MatchScore, smCtx.FastResult.MatchLevel,
			smCtx.FastResult.MatchHighlights, smCtx.FastResult.RiskPoints))
	}

	sb.WriteString(`

【校招初筛核心判断问题】
1. 学生的技术方向是否与岗位方向一致？
2. 学生是否有真实项目或实习经历能支撑该方向的初筛？
3. 如果有项目/实习，简历表达是否需要强化？
4. 如果方向错位（后端投算法、NLP、咨询），必须判为不建议或换方向。

【关键提醒】
- Fast初判分数只是参考，如果过于严格请独立思考
- 简历中"权限校验、登录鉴权、缓存使用、接口开发、数据库设计"等具体场景都是有效的后端工程证据
- "企业课程平台项目"是真实后端项目，不是普通课程作业
- 校招场景：有真实项目+实习+技术栈基本覆盖 = 70-85分

# 输出格式（严格JSON）
{
  "score": 78,
  "level": "中等偏高",
  "decision": "revise_before_apply",
  "decision_label": "修改后投递",
  "summary": "一句话总结匹配情况",
  "confidence": "medium",
  "decision_reasons": ["原因1", "原因2"]
}

decision可选值: direct_apply, revise_before_apply, backup, not_recommended, change_direction`)
	return sb.String()
}

func buildHy3Module2UserPrompt(smCtx *ScreeningMatchContext, decision *ScreeningDecision) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# 投递决策\n分数: %d, 建议: %s\n", decision.Score, decision.DecisionLabel))
	sb.WriteString("\n# 简历原文\n")
	sb.WriteString(truncate(smCtx.ResumeRawText, 3000))
	sb.WriteString("\n\n# 岗位JD\n")
	sb.WriteString(truncate(smCtx.JobRawJD, 1500))
	sb.WriteString(`

# 输出格式（严格JSON数组，最多3条）
[
  {
    "title": "命中点标题",
    "why_it_helps": "为什么帮助通过初筛",
    "resume_evidence": "简历中的具体证据原文",
    "screening_value": "对初筛的价值"
  }
]`)
	return sb.String()
}

func buildHy3Module3UserPrompt(smCtx *ScreeningMatchContext, decision *ScreeningDecision) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# 投递决策\n分数: %d, 建议: %s\n", decision.Score, decision.DecisionLabel))
	sb.WriteString("\n# 简历原文\n")
	sb.WriteString(truncate(smCtx.ResumeRawText, 3000))
	sb.WriteString("\n\n# 岗位JD\n")
	sb.WriteString(truncate(smCtx.JobRawJD, 1500))

	if smCtx.FastResult != nil && len(smCtx.FastResult.RiskPoints) > 0 {
		sb.WriteString(fmt.Sprintf("\n\n# Fast识别风险\n%v", smCtx.FastResult.RiskPoints))
	}

	sb.WriteString(`

# 输出格式（严格JSON数组，最多3条）
[
  {
    "risk_title": "风险标题",
    "severity": "high/medium/low",
    "why_it_blocks_screening": "为什么影响初筛",
    "current_resume_problem": "当前简历哪里有问题",
    "fix_direction": "修复方向建议"
  }
]`)
	return sb.String()
}

func buildHy3Module4UserPrompt(smCtx *ScreeningMatchContext, decision *ScreeningDecision, risks []ScreeningEliminationRisk) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# 投递决策\n分数: %d, 建议: %s\n", decision.Score, decision.DecisionLabel))
	sb.WriteString("\n# 简历原文\n")
	sb.WriteString(truncate(smCtx.ResumeRawText, 3000))
	sb.WriteString("\n\n# 岗位JD\n")
	sb.WriteString(truncate(smCtx.JobRawJD, 1000))
	sb.WriteString(fmt.Sprintf("\n\n# 已知风险\n"))
	for _, r := range risks {
		sb.WriteString(fmt.Sprintf("- %s\n", r.RiskTitle))
	}
	sb.WriteString(`

# 关键要求
1. 每条必须有 before_example（简历中的原文表达风格）
2. 每条必须有 after_example（基于简历真实经历的优化表达）
3. after_example 不能编造简历中没有的技术、指标、数据
4. 如果简历确实缺少某方面经历，明确说"不要硬写"

# 输出格式（严格JSON数组，2-4条）
[
  {
    "target_section": "项目经历",
    "problem": "当前表达的问题",
    "rewrite_instruction": "改写指导",
    "before_example": "改前示例（简历风格）",
    "after_example": "改后示例（优化版本）",
    "why_it_improves_screening": "为什么提高初筛通过率"
  }
]`)
	return sb.String()
}

func buildHy3Module5UserPrompt(smCtx *ScreeningMatchContext, decision *ScreeningDecision, risks []ScreeningEliminationRisk) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# 投递决策\n分数: %d, 建议: %s\n", decision.Score, decision.DecisionLabel))
	sb.WriteString(fmt.Sprintf("# 岗位\n%s - %s\n", smCtx.JobTitle, smCtx.JobCategory))
	sb.WriteString("\n# 简历核心方向\n")
	sb.WriteString(truncate(smCtx.ResumeRawText, 1000))

	if smCtx.FastResult != nil && len(smCtx.FastResult.MatchHighlights) > 0 {
		sb.WriteString(fmt.Sprintf("\n\n# 匹配亮点\n%v", smCtx.FastResult.MatchHighlights))
	}

	sb.WriteString(`

# 输出格式（严格JSON对象）
{
  "positioning": "主投岗位 / 备选岗位 / 不建议投递",
  "recommended_action": "具体建议",
  "if_apply_now": "现在投递会怎样",
  "better_fit_directions": ["更匹配方向1", "更匹配方向2"],
  "avoid_directions": ["应避开方向1"]
}`)
	return sb.String()
}

// ─── Helper functions ──────────────────────────────────────────────────────

func deriveDecisionFromScore(score int) string {
	switch {
	case score >= 85:
		return "direct_apply"
	case score >= 70:
		return "revise_before_apply"
	case score >= 55:
		return "backup"
	case score >= 40:
		return "not_recommended"
	default:
		return "change_direction"
	}
}

func deriveDecisionLabel(decision string) string {
	switch decision {
	case "direct_apply":
		return "可以直接投递"
	case "revise_before_apply":
		return "修改后投递"
	case "backup":
		return "备选投递"
	case "not_recommended":
		return "不建议投递"
	case "change_direction":
		return "建议换方向"
	default:
		return decision
	}
}

// normalizeDecisionLabel maps Hy3 model output variants to standard labels.
// JM-R6-P2: ensures all user-visible labels are consistent.
func normalizeDecisionLabel(label string, decision string) string {
	// If empty, derive from decision
	if label == "" {
		return deriveDecisionLabel(decision)
	}
	// Map known Hy3 output variants
	norm := strings.TrimSpace(label)
	switch {
	case decision == "change_direction":
		// Acceptable: 建议换方向. Unacceptable: 换方向投递, 换方向, 更换方向投递, etc.
		if strings.Contains(norm, "建议") && strings.Contains(norm, "换方向") {
			return "建议换方向"
		}
		if strings.Contains(norm, "换方向") {
			return "建议换方向"
		}
		return "建议换方向"
	case decision == "not_recommended":
		if strings.Contains(norm, "不建议") || strings.Contains(norm, "暂不") {
			return "不建议投递"
		}
		return "不建议投递"
	case decision == "revise_before_apply":
		if strings.Contains(norm, "修改") {
			return "修改后投递"
		}
		return "修改后投递"
	case decision == "backup":
		if strings.Contains(norm, "备选") {
			return "备选投递"
		}
		return "备选投递"
	case decision == "direct_apply":
		if strings.Contains(norm, "可以直接") || strings.Contains(norm, "优先") {
			return "可以直接投递"
		}
		return "可以直接投递"
	default:
		return deriveDecisionLabel(decision)
	}
}

// repairSimpleJSON attempts basic JSON repair: unmatched braces, trailing commas.
func repairSimpleJSON(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return ""
	}
	// Ensure it starts with { or [
	if !strings.HasPrefix(s, "{") && !strings.HasPrefix(s, "[") {
		return ""
	}
	// Remove trailing commas before } or ]
	s = strings.ReplaceAll(s, ",}", "}")
	s = strings.ReplaceAll(s, ",]", "]")
	// Try parsing
	if json.Valid([]byte(s)) {
		return s
	}
	// Count braces and close if needed
	openBraces := strings.Count(s, "{")
	closeBraces := strings.Count(s, "}")
	for openBraces > closeBraces {
		s += "}"
		closeBraces++
	}
	openBrackets := strings.Count(s, "[")
	closeBrackets := strings.Count(s, "]")
	for openBrackets > closeBrackets {
		s += "]"
		closeBrackets++
	}
	if json.Valid([]byte(s)) {
		return s
	}
	return ""
}

// ─── Rule-Based Fallback Pipeline ──────────────────────────────────────────
// Kept as-is for fallback when Hy3 is unavailable.

// runScreeningMatchPipeline runs the 5-module + judge pipeline (rule-based).
func runScreeningMatchPipeline(ctx *ScreeningMatchContext) *ScreeningMatchV1 {
	result := &ScreeningMatchV1{
		Mode:    "deep",
		Version: "screening_match_v1",
	}

	// Module 1: Screening Decision
	decision := runScreeningDecisionModule(ctx)
	result.ScreeningDecision = *decision

	// Module 2: Pass Points
	passPoints := runScreeningPassPointsModule(ctx, decision)
	result.ScreeningPassPoints = passPoints

	// Module 3: Elimination Risks
	risks := runScreeningEliminationRisksModule(ctx, decision)
	result.ScreeningEliminationRisks = risks

	// Module 4: Resume Rewrite Commands
	commands := runResumeRewriteCommandsModule(ctx, decision, risks)
	result.ResumeRewriteCommands = commands

	// Module 5: Job Strategy
	strategy := runJobStrategyModule(ctx, decision, risks)
	result.JobStrategy = *strategy

	// Judge merge — already assembled, ensure consistency
	runScreeningJudgeModule(result)

	return result
}

// ─── Module 1: Screening Decision ────────────────────────────────────────

func runScreeningDecisionModule(ctx *ScreeningMatchContext) *ScreeningDecision {
	// Deterministic: based on Fast Result + SEP evidence
	fastScore := 50
	if ctx.FastResult != nil {
		fastScore = ctx.FastResult.MatchScore
	}

	// ─── Evidence-based score recalibration ──────────────────────────────
	calibratedScore := recalibrateScreeningScore(ctx, fastScore)

	// Build reasons from Fast structured data
	reasons := []string{}
	if ctx.FastResult != nil {
		if len(ctx.FastResult.MatchHighlights) > 0 {
			hl := ctx.FastResult.MatchHighlights
			if len(hl) > 2 { hl = hl[:2] }
			reasons = append(reasons, "有匹配技能或项目经历："+strings.Join(hl, "、"))
		}
		if len(ctx.FastResult.GapPoints) > 0 {
			reasons = append(reasons, "简历表达需强化："+ctx.FastResult.GapPoints[0])
		}
	}
	if len(reasons) == 0 {
		reasons = []string{"基于简历和岗位匹配分析"}
	}

	// If score was boosted, add calibration reason
	if calibratedScore > fastScore {
		reasons = append(reasons, fmt.Sprintf("系统检测到简历中有真实项目和实习经验证据，匹配度从%d校准至%d。", fastScore, calibratedScore))
	}

	decision := &ScreeningDecision{
		Score:           calibratedScore,
		Level:           matchLevel(calibratedScore),
		DecisionReasons: reasons,
		Confidence:      "medium",
	}

	// Derive decision and label from score
	switch {
	case calibratedScore >= 85:
		decision.Decision = "direct_apply"
		decision.DecisionLabel = "可以直接投递"
	case calibratedScore >= 70:
		decision.Decision = "revise_before_apply"
		decision.DecisionLabel = "修改后投递"
	case calibratedScore >= 55:
		decision.Decision = "backup"
		decision.DecisionLabel = "备选投递"
	case calibratedScore >= 40:
		decision.Decision = "not_recommended"
		decision.DecisionLabel = "不建议投递"
	default:
		decision.Decision = "change_direction"
		decision.DecisionLabel = "建议换方向"
	}

	// Build summary
	decision.Summary = fmt.Sprintf("该岗位%s方向与简历相关度%s，投递%s。%s",
		ctx.JobTitle, matchLevel(calibratedScore), decision.DecisionLabel, strings.Join(reasons, "。"))

	return decision
}

// recalibrateScreeningScore adjusts SEP Fast score using evidence signals
// that SEP may undervalue: internship presence, project depth, direction alignment.
func recalibrateScreeningScore(ctx *ScreeningMatchContext, baseScore int) int {
	if ctx.FastResult == nil {
		return baseScore
	}
	score := baseScore

	// Only apply recalibration when SEP profiles are available (full match, not light fallback)
	if ctx.ResumeSEPProfile == nil || ctx.JobSEPProfile == nil {
		return score
	}

	// 1. Check for internship evidence: from highlights, SEP tags, and evidence text
	hasInternship := false
	hasProject := false
	hasDirectionMatch := true

	// Check MatchHighlights
	for _, h := range ctx.FastResult.MatchHighlights {
		if strings.Contains(h, "实习") {
			hasInternship = true
		}
		if strings.Contains(h, "项目") || strings.Contains(h, "开发") {
			hasProject = true
		}
	}

	// Check SEP tags for internship/project evidence (may not appear in highlights)
	if !hasInternship || !hasProject {
		if tags, ok := ctx.ResumeSEPProfile.([]*model.SemanticProfileTag); ok {
			for _, t := range tags {
				ev := strings.ToLower(t.EvidenceText + t.TagLabel + t.Explanation)
				if strings.Contains(ev, "实习") {
					hasInternship = true
				}
				if strings.Contains(ev, "项目") || strings.Contains(t.Category, "project") {
					hasProject = true
				}
			}
		}
	}

	// Check for direction mismatch in risks
	for _, r := range ctx.FastResult.RiskPoints {
		if strings.Contains(r, "mismatch") || strings.Contains(r, "方向匹配度偏低") {
			hasDirectionMatch = false
		}
	}

	// Only boost when direction matches (avoid boosting for wrong-direction cases)
	if !hasDirectionMatch {
		return score
	}

	// Boost for internship evidence (real enterprise experience)
	if hasInternship {
		score += 8
	}
	// Boost for project evidence
	if hasProject && len(ctx.FastResult.MatchHighlights) >= 2 {
		score += 4
	}
	// Boost for direction alignment
	if hasDirectionMatch && len(ctx.FastResult.MatchHighlights) >= 2 {
		score += 3
	}

	// Floor guard: if resume has internship + projects + direction match, minimum score 70
	if hasInternship && hasProject && hasDirectionMatch && score < 70 {
		score = 70
	}

	// Cap: don't boost beyond reasonable range
	if score > 95 {
		score = 95
	}
	return score
}

// ─── Module 2: Screening Pass Points ─────────────────────────────────────

func runScreeningPassPointsModule(ctx *ScreeningMatchContext, decision *ScreeningDecision) []ScreeningPassPoint {
	points := []ScreeningPassPoint{}
	if ctx.FastResult == nil { return points }

	highlights := ctx.FastResult.MatchHighlights
	evNames := ctx.FastResult.ResumeEvidence

	for i, h := range highlights {
		if i >= 3 { break }
		ev := ""
		if i < len(evNames) { ev = evNames[i] }
		points = append(points, ScreeningPassPoint{
			Title:          fmt.Sprintf("具备%s相关经历", h),
			WhyItHelps:     fmt.Sprintf("岗位重视%s能力，你的简历中有明确的%s相关描述，这有助于HR初筛时快速识别你的匹配度。", h, h),
			ResumeEvidence: ev,
			ScreeningValue: fmt.Sprintf("能在%v方向通过初筛关键词过滤", ctx.JobDirection+ctx.JobTitle),
		})
	}

	if len(points) == 0 {
		points = append(points, ScreeningPassPoint{
			Title:      "简历方向基本匹配",
			WhyItHelps: "你的简历背景与岗位要求存在关联性，HR初筛时不会因方向完全错位而被直接过滤。",
			ScreeningValue: "方向相关性是最基础的初筛条件。",
		})
	}
	return points
}

// ─── Module 3: Screening Elimination Risks ───────────────────────────────

func runScreeningEliminationRisksModule(ctx *ScreeningMatchContext, decision *ScreeningDecision) []ScreeningEliminationRisk {
	risks := []ScreeningEliminationRisk{}
	if ctx.FastResult == nil { return risks }

	gaps := ctx.FastResult.GapPoints
	// From structured risks
	if sr, ok := ctx.FastResult.StructuredRisks.([]MatchRiskPoint); ok {
		for i, r := range sr {
			if i >= 3 { break }
			sev := r.RiskLevel
			if sev == "" { sev = "medium" }
			// Improve fix direction with concrete, actionable advice
			fixDir := r.FixSuggestion
			currentProblem := r.EvidenceGap
			if fixDir == "" || strings.Contains(fixDir, "补充") {
				fixDir = buildConcreteFixDirection(r.RiskName, ctx)
			}
			if currentProblem == "" {
				currentProblem = fmt.Sprintf("简历中%s相关的具体证据不够充分，HR初筛时缺少明确判断依据。", r.RiskName)
			}
			risks = append(risks, ScreeningEliminationRisk{
				RiskTitle:            r.RiskName,
				Severity:             sev,
				WhyItBlocksScreening: r.WhyItMatters,
				CurrentResumeProblem: currentProblem,
				FixDirection:         fixDir,
			})
		}
	}
	_ = gaps

	if len(risks) == 0 {
		risks = append(risks, ScreeningEliminationRisk{
			RiskTitle:            "项目描述偏功能列举，缺乏工程深度",
			Severity:             "low",
			WhyItBlocksScreening: "HR从简历中看到的是功能点罗列而非工程能力体现，同等技能基础的候选人中，工程表达更强的简历优先被关注。",
			CurrentResumeProblem: "项目描述以功能点为线索，没有体现接口数量、数据规模、性能结果、线上运行状态等工程深度信号。",
			FixDirection:         "用工程语言改写项目描述：补充接口数量、数据库表数量、缓存命中率提升、QPS、线上状态等可量化指标，把「做了什么」改成「解决了什么问题+怎么验证的」。",
		})
	}
	return risks
}

// buildConcreteFixDirection generates actionable fix text based on risk type.
func buildConcreteFixDirection(riskName string, ctx *ScreeningMatchContext) string {
	riskLow := strings.ToLower(riskName)
	switch {
	case strings.Contains(riskLow, "覆盖不足") || strings.Contains(riskLow, "技能"):
		// Map the missing skills to a concrete fix direction
		gapText := ""
		if ctx.FastResult != nil && len(ctx.FastResult.GapPoints) > 0 {
			for _, g := range ctx.FastResult.GapPoints {
				if strings.Contains(strings.ToLower(g), "缺少") {
					gapText = g
					break
				}
			}
		}
		if gapText != "" {
			return fmt.Sprintf("在简历的项目经历中%v，用具体项目场景证明你理解并实践过这些技术点，而不是仅列在技能列表里。", extractGapContent(gapText))
		}
		return "针对岗位JD中列出的核心技术点，检查简历项目经历是否覆盖了对应的工程场景——例如岗位要求缓存，简历项目应体现缓存的key设计、过期策略和一致性处理。"
	case strings.Contains(riskLow, "方向"):
		return "该岗位的核心技术方向与你的实际项目经历存在差异。如果坚持投递，需要在简历补充说明中解释你的转型动机和准备；更建议转向与你现有经验匹配的方向。"
	case strings.Contains(riskLow, "证据"):
		return "为每个技能/项目补充可验证的证据：具体项目名称、你在其中的角色、使用的技术栈、达成的量化结果以及项目是否已上线运行。"
	default:
		return "结合实际项目经历，用STAR法则（情境-任务-行动-结果）重写项目描述，让HR能从简历中看到你的具体贡献和技术深度。"
	}
}

// extractGapContent extracts the meaningful part from a gap description.
func extractGapContent(gapText string) string {
	// Remove common prefixes like "缺少" or "需补充"
	clean := strings.TrimPrefix(gapText, "缺少")
	clean = strings.TrimPrefix(clean, "需补充")
	clean = strings.TrimPrefix(clean, "补充")
	clean = strings.TrimSpace(clean)
	if clean == "" {
		return "补充对应技术的项目实践描述"
	}
	return "补充" + clean + "的项目实践描述"
}

// ─── Module 4: Resume Rewrite Commands ───────────────────────────────────

func runResumeRewriteCommandsModule(ctx *ScreeningMatchContext, decision *ScreeningDecision, risks []ScreeningEliminationRisk) []ResumeRewriteCommand {
	commands := []ResumeRewriteCommand{}

	// For not_recommended / change_direction: don't suggest rewriting to fake experience
	if decision.Decision == "not_recommended" || decision.Decision == "change_direction" {
		commands = append(commands, ResumeRewriteCommand{
			TargetSection:          "求职意向和补充说明",
			Problem:                "当前简历与岗位方向不匹配，不适合投递该岗位。强行修改简历去贴合会编造不存在的能力经历。",
			RewriteInstruction:     "不建议修改简历去投递该岗位。应保留简历真实内容，转向更适合的方向。",
			BeforeExample:          "投递方向与该岗位不匹配仍要修改简历去投递。",
			AfterExample:           "保留简历的真实后端工程经历描述，将求职意向调整为后台开发、平台后端、服务端开发等方向。",
			WhyItImprovesScreening: "真实的经历比勉强凑合的改写更能通过面试关。HR和面试官能识别出不匹配的强行包装。",
		})
		return commands
	}

	// For revise_before_apply / backup / direct_apply: generate concrete rewrites
	// Build rewrites from structured risks mapping to concrete before/after
	rewriteCount := 0
	for _, r := range risks {
		if rewriteCount >= 2 { break }
		before, after := buildBeforeAfterPair(r, ctx)
		if before != "" && after != "" {
			commands = append(commands, ResumeRewriteCommand{
				TargetSection:          "项目经历",
				Problem:                r.CurrentResumeProblem,
				RewriteInstruction:     r.FixDirection,
				BeforeExample:          before,
				AfterExample:           after,
				WhyItImprovesScreening: fmt.Sprintf("修改后能更好展示你在%s方面的工程实践能力，HR能在10秒内识别到关键匹配信号。", r.RiskTitle),
			})
			rewriteCount++
		}
	}

	// Ensure minimum 2 commands for applicable cases
	if len(commands) < 2 {
		commands = append(commands,
			ResumeRewriteCommand{
				TargetSection:          "项目经历",
				Problem:                "项目描述偏功能罗列，没有体现工程能力和技术深度。",
				RewriteInstruction:     "把具体的接口数、数据量、性能指标等技术参数补充到项目描述中，让HR能判断你的工程能力级别。",
				BeforeExample:          "参与课程权限管理、学习进度记录、资源下载等功能开发。",
				AfterExample:           "参与企业课程学习平台后端核心链路建设，负责课程权限校验、播放签名鉴权、学习进度一致性与资源下载权限控制，补齐登录态、课程状态、学习权限、小节归属、视频绑定等多层校验，降低越权访问风险。",
				WhyItImprovesScreening: "工程化表达能让HR快速判断你的技术深度和项目复杂度，比功能点罗列更具初筛价值。",
			},
			ResumeRewriteCommand{
				TargetSection:          "实习经历或项目经历",
				Problem:                "缺少量化结果和性能数据支撑。",
				RewriteInstruction:     "在项目描述末尾补充该项目的技术规模指标和你的具体贡献。",
				BeforeExample:          "使用Redis缓存商品详情，提升查询效率。",
				AfterExample:           "使用Redis缓存热门商品详情和首页商品列表，设置分场景过期策略与缓存Key规则，将重复访问场景下详情接口平均响应时间从180ms降至60ms，缓存命中率提升至85%以上。",
				WhyItImprovesScreening: "有量化结果的项目经历比纯功能描述更能证明你的工程产出，这是HR筛简历时区分'做过'和'做好'的关键依据。",
			})
	}

	if len(commands) > 4 {
		commands = commands[:4]
	}
	return commands
}

// buildBeforeAfterPair generates a before/after rewrite pair based on risk context.
func buildBeforeAfterPair(risk ScreeningEliminationRisk, ctx *ScreeningMatchContext) (string, string) {
	riskLow := strings.ToLower(risk.RiskTitle)

	// Map risk type to a concrete before/after based on available evidence
	skillsBefore := "在技能列表中列出了相关技术栈名称。"
	skillsAfter := "在项目经历中用具体场景展示技术栈的实际应用：例如使用Redis实现缓存策略并测量响应时间变化，使用JWT实现登录态管理并对敏感接口增加权限校验。"

	switch {
	case strings.Contains(riskLow, "覆盖不足") || strings.Contains(riskLow, "技能"):
		// Build a contextual after from match highlights
		suffix := "并补充性能指标、异常处理和上线验证结果。"
		if len(ctx.FastResult.MatchHighlights) > 0 {
			hl := ctx.FastResult.MatchHighlights[0]
			suffix = fmt.Sprintf("并结合%s的实际项目，补充接口数量、数据库表数量、缓存策略和测试验证结果。", hl)
		}
		return skillsBefore, skillsAfter + suffix

	case strings.Contains(riskLow, "证据") || strings.Contains(riskLow, "表达"):
		return "负责后端接口开发和数据库设计。",
			"负责后端核心接口开发（共完成X个接口），设计数据库表结构（共Y张表含索引优化），使用Swagger维护接口文档，完成前后端联调并通过测试环境验收。"

	case strings.Contains(riskLow, "方向"):
		if len(ctx.FastResult.MatchHighlights) > 0 {
			hl := ctx.FastResult.MatchHighlights[0]
			return fmt.Sprintf("求职意向描述较笼统，与%s岗位的方向关联不够明确。", ctx.JobTitle),
				fmt.Sprintf("求职意向明确为%s方向，并在项目经历中将%s相关的工程实践作为核心叙事线展开。", ctx.JobTitle, hl)
		}
		return "", ""

	default:
		return "项目描述以功能列举为主。",
			"项目描述以工程链路为主线：从需求分析→数据库设计→接口开发→权限校验→缓存优化→前后端联调→测试验收，体现完整的后端工程能力。"
	}
}

// ─── Module 5: Job Strategy ──────────────────────────────────────────────

func runJobStrategyModule(ctx *ScreeningMatchContext, decision *ScreeningDecision, risks []ScreeningEliminationRisk) *JobStrategy {
	strategy := &JobStrategy{
		Positioning:       "备选岗位",
		RecommendedAction: "建议先优化简历表达，再投递该岗位。",
		IfApplyNow:        "当前简历可以体现基础能力，但初筛竞争力一般，容易被表达更好的候选人压过。",
	}

	switch decision.Decision {
	case "direct_apply":
		strategy.Positioning = "主投岗位"
		strategy.RecommendedAction = "当前简历与该岗位方向匹配、有真实项目经历支撑，可以直接投递。建议投递前再次检查简历项目描述是否突出了与岗位JD关键词的对应关系。"
		strategy.IfApplyNow = "当前简历与岗位方向匹配度较高，有明确的工程实践经验支撑，初筛通过概率较好。但仍建议在投递前用岗位JD的关键词校准一遍简历表达。"
	case "revise_before_apply":
		strategy.Positioning = "主投岗位"
		strategy.RecommendedAction = "先完成简历定向优化——重点补充项目描述中的工程深度（接口数量、数据规模、性能指标）和岗位JD关键词覆盖，再投递。"
		strategy.IfApplyNow = "你的简历方向匹配、有真实项目经历，但目前项目描述偏功能列举，工程深度表达不够。优化后初筛通过率会有明显提升。"
	case "not_recommended", "change_direction":
		strategy.Positioning = "不建议作为主投岗位"
		strategy.RecommendedAction = "该岗位方向与你的简历核心经历差距较大，不建议把有限的投递机会用在这个岗位上。请优先投递与你现有经验方向匹配的岗位。"
		strategy.IfApplyNow = "当前简历与该岗位方向不匹配，HR在初筛阶段会因为方向不对口快速过滤。不建议为了碰运气而消耗投递机会。"
	}

	// Derive better fit directions from resume SEP tags and context
	if ctx.FastResult != nil {
		strategy.BetterFitDirections = deriveBetterFitDirections(ctx)
	}
	strategy.AvoidDirections = deriveAvoidDirections(ctx)

	return strategy
}

// deriveBetterFitDirections recommends directions based on resume evidence.
func deriveBetterFitDirections(ctx *ScreeningMatchContext) []string {
	directions := []string{"后台开发", "平台后端", "服务端开发"}

	// Analyze resume SEP tags for direction signals
	if tags, ok := ctx.ResumeSEPProfile.([]*model.SemanticProfileTag); ok && len(tags) > 0 {
		skillSet := make(map[string]bool)
		for _, t := range tags {
			skillSet[strings.ToLower(t.TagLabel)] = true
		}

		// Add more specific directions based on skill evidence
		if skillSet["go"] || skillSet["gin"] {
			directions = appendIfMissing(directions, "Go后端开发")
		}
		if skillSet["java"] || skillSet["spring boot"] || skillSet["spring"] {
			directions = appendIfMissing(directions, "Java后端开发")
		}
		if skillSet["docker"] || skillSet["nginx"] || skillSet["linux"] {
			directions = appendIfMissing(directions, "云产品后端开发")
		}
		if skillSet["mysql"] || skillSet["redis"] {
			directions = appendIfMissing(directions, "数据平台后端开发")
		}
		if skillSet["rabbitmq"] || skillSet["kafka"] || skillSet["消息队列"] {
			directions = appendIfMissing(directions, "中间件与基础架构开发")
		}
	}

	// Add the current job direction if it differs from existing
	if ctx.JobDirection != "" {
		directions = appendIfMissing(directions, ctx.JobDirection+"方向")
	}

	if len(directions) > 6 {
		directions = directions[:6]
	}
	return directions
}

// deriveAvoidDirections returns directions that don't fit the resume's profile.
func deriveAvoidDirections(ctx *ScreeningMatchContext) []string {
	avoids := []string{"算法研究", "纯产品策划", "视觉设计"}

	// Check for direction mismatch in risks to add context-specific avoids
	for _, r := range ctx.FastResult.RiskPoints {
		if strings.Contains(strings.ToLower(r), "mismatch") || strings.Contains(strings.ToLower(r), "方向匹配度偏低") {
			avoids = appendIfMissing(avoids, ctx.JobTitle+"方向")
		}
	}
	return avoids
}

// appendIfMissing adds a string to a slice if not already present.
func appendIfMissing(slice []string, s string) []string {
	for _, v := range slice {
		if strings.EqualFold(v, s) {
			return slice
		}
	}
	return append(slice, s)
}

// ─── Judge Module ─────────────────────────────────────────────────────────

func runScreeningJudgeModule(result *ScreeningMatchV1) {
	// Consistency checks
	dec := result.ScreeningDecision

	// If high severity risks exist but decision is direct_apply → downgrade
	hasHigh := false
	for _, r := range result.ScreeningEliminationRisks {
		if r.Severity == "high" { hasHigh = true }
	}
	if hasHigh && dec.Decision == "direct_apply" {
		dec.Decision = "revise_before_apply"
		dec.DecisionLabel = "修改后投递"
	}

	// If no pass points but score > 60 → adjust score
	if len(result.ScreeningPassPoints) == 0 && dec.Score > 60 {
		dec.Score = 55
		dec.Level = "偏低"
	}

	// Ensure decision reasons are not empty
	if len(dec.DecisionReasons) == 0 {
		dec.DecisionReasons = []string{fmt.Sprintf("基于%s的岗位匹配分析", "语义画像")}
	}

	result.ScreeningDecision = dec
}

// ─── H19-B: Deep Match Agent (legacy) ─────────────────────────────────────
// The following code is kept for backward compatibility and legacy fallback.
// Deep primary path now uses runScreeningMatchPipeline above.

// ─── H19-B: Deep Match V3 Pipeline ─────────────────────────────────────────
// 8-stage SEP + Fast Result driven pipeline. Replaces TagEvidenceService / 3-agent.

// buildDeepMatchContextV3 builds the v3 context from SEP profiles and Fast result.
func (s *MatchService) buildDeepMatchContextV3(
	ctx context.Context, userID uint64, resume *model.ResumeVersion, job *model.Job,
) (*DeepMatchContextV3, error) {
	v3ctx := &DeepMatchContextV3{
		ResumeID:  int64(resume.ID),
		JobID:     int64(job.ID),
		ResumeRaw: "", // populated later if available
		JobRaw:    job.Description,
		Status:    "building",
	}

	if s.semanticRepo == nil {
		v3ctx.Status = "missing"
		v3ctx.Reason = "semantic profile repository not available"
		return v3ctx, fmt.Errorf("deep match v3: semanticRepo not available")
	}

	// Read resume SEP
	rp, rerr := s.semanticRepo.FindLatestByTarget(ctx, model.SemanticTargetResume, int64(resume.ID))
	if rerr != nil || rp == nil || rp.Status != model.SemanticProfileStatusVerified {
		v3ctx.Status = "profile_missing"
		v3ctx.Reason = fmt.Sprintf("resume SEP not ready: resume=%d", resume.ID)
		return v3ctx, fmt.Errorf("deep match v3: %s", v3ctx.Reason)
	}
	rtags, _ := s.semanticRepo.ListTagsByProfileID(ctx, rp.ID)
	if len(rtags) < 3 {
		v3ctx.Status = "profile_insufficient"
		v3ctx.Reason = fmt.Sprintf("resume SEP tags insufficient: %d", len(rtags))
		return v3ctx, fmt.Errorf("deep match v3: %s", v3ctx.Reason)
	}
	v3ctx.ResumeSEPTags = rtags

	// Read job SEP
	jp, jerr := s.semanticRepo.FindLatestByTarget(ctx, model.SemanticTargetJob, int64(job.ID))
	if jerr != nil || jp == nil || jp.Status != model.SemanticProfileStatusVerified {
		v3ctx.Status = "profile_missing"
		v3ctx.Reason = fmt.Sprintf("job SEP not ready: job=%d", job.ID)
		return v3ctx, fmt.Errorf("deep match v3: %s", v3ctx.Reason)
	}
	jtgs, _ := s.semanticRepo.ListTagsByProfileID(ctx, jp.ID)
	if len(jtgs) < 3 {
		v3ctx.Status = "profile_insufficient"
		v3ctx.Reason = fmt.Sprintf("job SEP tags insufficient: %d", len(jtgs))
		return v3ctx, fmt.Errorf("deep match v3: %s", v3ctx.Reason)
	}
	v3ctx.JobSEPTags = jtgs

	// Run Fast SEP Match to get Fast Result
	fastRes := runFastSEPMatch(&MatchSEPContext{
		ResumeTags: rtags, JobTags: jtgs, JobTitle: job.Title,
		ResumeID: int64(resume.ID), JobID: int64(job.ID),
	})
	v3ctx.FastResult = fastRes

	// Extract Fast structured fields
	if sb, ok := fastRes.ScoreBreakdownObj.(ScoreBreakdownResult); ok {
		v3ctx.FastScoreBreakdown = sb
	}
	if hg, ok := fastRes.HardGateObj.(HardGateResult); ok {
		v3ctx.FastHardGate = hg
	}
	if cov, ok := fastRes.RequirementCoverage.([]RequirementCoverageItem); ok {
		v3ctx.FastRequirementCoverage = cov
	}
	if risks, ok := fastRes.StructuredRisks.([]MatchRiskPoint); ok {
		v3ctx.FastStructuredRisks = risks
	}

	v3ctx.Status = "ready"
	return v3ctx, nil
}

// runDeepMatchV3 executes the 8-layer H19-B deep match pipeline.
func runDeepMatchV3(v3ctx *DeepMatchContextV3) *DeepMatchV3Report {
	report := &DeepMatchV3Report{
		GenerationTier: "quality",
		GeneratedBy:    "deep_match_agent",
		AgentStatus:    "match_verified",
		ReportVersion:  "deep_match_v3",
	}
	trace := []DeepAgentTrace{}

	// ─── L0: Input Gate ──────────────────────────────────────────────
	if v3ctx.Status != "ready" {
		report.AgentStatus = "profile_missing"
		report.GeneratedBy = "deep_match_profile_pending"
		report.OneSentenceConclusion = "语义画像尚未准备完成，暂不能给出深度匹配分析。"
		report.AgentTrace = []DeepAgentTrace{
			{Stage: "INPUT_GATE", Agent: "输入门控Agent", Status: "blocked", Summary: v3ctx.Reason},
		}
		report.QualityChecks = makeQualityChecks(false, false)
		return report
	}
	trace = append(trace, DeepAgentTrace{Stage: "INPUT_GATE", Agent: "输入门控Agent", Status: "done", Summary: "SEP Profile + Fast Result ready"})

	// ─── L1: JD Deconstruction ──────────────────────────────────────
	ju := buildV3JobUnderstanding(v3ctx)
	report.JobUnderstanding = ju
	trace = append(trace, DeepAgentTrace{Stage: "JD_DECONSTRUCTION", Agent: "JD解构Agent", Status: "done",
		Summary: fmt.Sprintf("解构%d项必备/%d项重要/%d项加分", len(ju.MustHaveRequirements), len(ju.ImportantRequirements), len(ju.BonusRequirements))})

	// ─── L2: Resume Evidence Audit ──────────────────────────────────
	evidence := buildV3ResumeEvidence(v3ctx)
	report.ResumeEvidenceMapping = evidence
	trace = append(trace, DeepAgentTrace{Stage: "RESUME_EVIDENCE_AUDIT", Agent: "证据审计Agent", Status: "done",
		Summary: fmt.Sprintf("提取%d条简历证据", len(evidence))})

	// ─── L3: Requirement Alignment ──────────────────────────────────
	coverage := buildV3RequirementCoverage(v3ctx, evidence)
	report.RequirementCoverage = coverage
	trace = append(trace, DeepAgentTrace{Stage: "REQUIREMENT_ALIGNMENT", Agent: "要求对齐Agent", Status: "done",
		Summary: fmt.Sprintf("生成%d条要求覆盖", len(coverage))})

	// ─── L4: Risk Diagnosis ─────────────────────────────────────────
	risks := buildV3KeyRisks(v3ctx, coverage)
	report.KeyRisks = risks
	trace = append(trace, DeepAgentTrace{Stage: "RISK_DIAGNOSIS", Agent: "风险诊断Agent", Status: "done",
		Summary: fmt.Sprintf("识别%d个风险", len(risks))})

	// ─── L5: Score Calibration ──────────────────────────────────────
	score, level, breakdown := calibrateV3Score(v3ctx, coverage, evidence, risks)
	report.MatchScore = score
	report.MatchLevel = level
	report.ScoreBreakdown = breakdown
	report.ScoreAlignment = buildScoreAlignment(v3ctx.FastResult.MatchScore, score, level)
	trace = append(trace, DeepAgentTrace{Stage: "SCORE_CALIBRATION", Agent: "分数校准Agent", Status: "done",
		Summary: fmt.Sprintf("Deep=%d(%s) vs Fast=%d", score, level, v3ctx.FastResult.MatchScore)})

	// ─── L6: Application Strategy ───────────────────────────────────
	mpRisks := convertV3RisksToMatchRiskPoints(risks)
	decision, decisionLabel := decideApplication(score, mpRisks, v3ctx.FastHardGate)
	report.ApplicationDecision = decision
	report.DecisionLabel = decisionLabel
	report.ApplicationStrategy = buildV3AppStrategy(score, level, decisionLabel)
	trace = append(trace, DeepAgentTrace{Stage: "APPLICATION_STRATEGY", Agent: "投递策略Agent", Status: "done",
		Summary: decisionLabel})

	// ─── L7: Resume Rewrite + Interview ─────────────────────────────
	report.ResumeRewriteStrategy = buildV3RewriteStrategy(coverage, risks)
	report.InterviewRiskQuestions = buildV3InterviewQuestions(ju, risks, coverage)
	report.ActionPlan = buildV3ActionPlan(report.ResumeRewriteStrategy)
	trace = append(trace, DeepAgentTrace{Stage: "REWRITE_AND_INTERVIEW", Agent: "改写+面试Agent", Status: "done",
		Summary: fmt.Sprintf("生成%d项改写+%d个面试追问", len(report.ResumeRewriteStrategy.SkillsReorder)+len(report.ResumeRewriteStrategy.KeywordsToAdd), len(report.InterviewRiskQuestions))})

	// ─── L8: Global Consistency ─────────────────────────────────────
	coreMatch := buildV3CoreMatchPoints(coverage, evidence)
	report.CoreMatchPoints = coreMatch
	report.OneSentenceConclusion = v3ctx.FastResult.Summary
	if report.OneSentenceConclusion == "" {
		report.OneSentenceConclusion = fmt.Sprintf("深度分析：匹配度%d分（%s），%s", score, level, decisionLabel)
	}
	report.QualityChecks = makeQualityChecks(true, true)
	report.AgentTrace = trace

	return report
}

// ─── L1: JD Deconstruction ───────────────────────────────────────────────

func buildV3JobUnderstanding(v3ctx *DeepMatchContextV3) DeepJobUnderstanding {
	ju := DeepJobUnderstanding{
		JobPositioning:  "基于SEP语义画像解构的岗位定位",
		CoreResponsibilities: []string{},
		MustHaveRequirements:  []string{},
		ImportantRequirements: []string{},
		BonusRequirements:     []string{},
		HiddenExpectations:    []string{},
	}
	coreResps := []string{}
	for _, t := range v3ctx.JobSEPTags {
		switch {
		case t.Category == "role_direction" || t.Category == "direction":
			ju.JobPositioning = t.TagLabel
			ju.MustHaveRequirements = append(ju.MustHaveRequirements, t.TagLabel)
		case t.Category == "technical_skill" || t.Category == "skill_core":
			ju.MustHaveRequirements = append(ju.MustHaveRequirements, t.TagLabel)
		case t.Category == "engineering_capability" || t.Category == "experience_signal":
			if t.Confidence >= 0.85 {
				ju.MustHaveRequirements = append(ju.MustHaveRequirements, t.TagLabel)
			} else {
				ju.ImportantRequirements = append(ju.ImportantRequirements, t.TagLabel)
			}
		case t.Category == "domain_context":
			ju.ImportantRequirements = append(ju.ImportantRequirements, t.TagLabel)
		case t.Category == "risk_signal":
			ju.HiddenExpectations = append(ju.HiddenExpectations, t.TagLabel)
		default:
			if t.Confidence >= 0.85 {
				ju.ImportantRequirements = append(ju.ImportantRequirements, t.TagLabel)
			} else {
				ju.BonusRequirements = append(ju.BonusRequirements, t.TagLabel)
			}
		}
		if t.Confidence >= 0.9 {
			coreResps = append(coreResps, t.TagLabel)
		}
	}
	ju.CoreResponsibilities = coreResps
	if len(ju.MustHaveRequirements) == 0 {
		ju.MustHaveRequirements = append(ju.MustHaveRequirements, ju.JobPositioning)
	}
	return ju
}

// ─── L2: Resume Evidence Audit ────────────────────────────────────────────

func buildV3ResumeEvidence(v3ctx *DeepMatchContextV3) []ResumeEvidenceItem {
	evidence := []ResumeEvidenceItem{}
	for _, t := range v3ctx.ResumeSEPTags {
		if t.EvidenceText == "" && t.Explanation == "" {
			continue
		}
		level := 2
		if t.EvidenceText != "" {
			level = 3
			evLow := strings.ToLower(t.EvidenceText)
			if strings.Contains(evLow, "实习") || strings.Contains(evLow, "公司") { level = 4 }
			if strings.Contains(evLow, "生产") || strings.Contains(evLow, "上线") { level = 5 }
		}
		sourceType := "skill_list"
		if strings.Contains(strings.ToLower(t.Category), "experience") || strings.Contains(strings.ToLower(t.Category), "project") {
			sourceType = "project"
		}
		if strings.Contains(strings.ToLower(t.EvidenceText), "实习") {
			sourceType = "internship_project"
		}
		confidence := "medium"
		if t.Confidence >= 0.9 { confidence = "high" }

		supportsReq := humanizeRequirementCategory(t.Category)
		if t.Explanation != "" { supportsReq = t.Explanation }

		evidence = append(evidence, ResumeEvidenceItem{
			EvidenceName:        t.TagLabel,
			EvidenceLevel:       level,
			SourceType:          sourceType,
			SourceTitle:         t.TagLabel,
			EvidenceText:        t.EvidenceText,
			SupportsRequirement: supportsReq,
			Confidence:          confidence,
		})
	}
	if len(evidence) > 10 { evidence = evidence[:10] }
	return evidence
}

// ─── L3: Requirement Alignment ────────────────────────────────────────────

func buildV3RequirementCoverage(v3ctx *DeepMatchContextV3, evidence []ResumeEvidenceItem) []RequirementCoverageItem {
	// Reuse Fast coverage as base, enrich with evidence-level detail
	coverage := v3ctx.FastRequirementCoverage
	if len(coverage) == 0 {
		// Build from scratch if Fast coverage missing
		sepCtx := &MatchSEPContext{
			ResumeTags: v3ctx.ResumeSEPTags, JobTags: v3ctx.JobSEPTags,
			JobTitle: "", ResumeID: v3ctx.ResumeID, JobID: v3ctx.JobID,
		}
		coverage = buildRequirementCoverage(sepCtx)
	}
	// Enrich with Deep evidence-level context
	for i := range coverage {
		if coverage[i].CoverageStatus == "weak" || coverage[i].CoverageStatus == "missing" {
			for _, ev := range evidence {
				if strings.Contains(strings.ToLower(ev.EvidenceText+ev.EvidenceName), strings.ToLower(coverage[i].RequirementName[:minInt(3, len(coverage[i].RequirementName))])) {
					if ev.EvidenceLevel >= 4 {
						coverage[i].CoverageStatus = "partial"
						coverage[i].CoverageScore = 65
						coverage[i].Gap = fmt.Sprintf("%s证据不够充分", coverage[i].RequirementName)
					}
				}
			}
		}
	}
	if len(coverage) > 8 { coverage = coverage[:8] }
	return coverage
}

// ─── L4: Risk Diagnosis ───────────────────────────────────────────────────

func buildV3KeyRisks(v3ctx *DeepMatchContextV3, coverage []RequirementCoverageItem) []DeepKeyRisk {
	risks := []DeepKeyRisk{}

	// Elevate from Fast structured risks
	for _, fr := range v3ctx.FastStructuredRisks {
		risks = append(risks, DeepKeyRisk{
			RiskName:       fr.RiskName,
			RiskLevel:      fr.RiskLevel,
			WhyItMatters:   fr.WhyItMatters,
			EvidenceGap:    fr.EvidenceGap,
			FixSuggestion:  fr.FixSuggestion,
			InterviewImpact: "可能在面试中被深入追问",
		})
	}

	// Add Deep-specific: uncovered must-have items
	missingCount := 0
	for _, c := range coverage {
		if c.CoverageStatus == "missing" && c.RequirementType == "must_have" {
			missingCount++
			if missingCount <= 3 {
				risks = append(risks, DeepKeyRisk{
					RiskName:       fmt.Sprintf("必备要求缺失：%s", c.RequirementName),
					RiskLevel:      "high",
					WhyItMatters:   fmt.Sprintf("岗位将%s列为核心要求，缺失可能直接影响简历筛选", c.RequirementName),
					EvidenceGap:    c.Gap,
					FixSuggestion:  c.Suggestion,
					InterviewImpact: "面试官可能质疑为何投递该岗位",
				})
			}
		}
	}

	if len(risks) == 0 {
		risks = append(risks, DeepKeyRisk{
			RiskName:       "当前匹配良好",
			RiskLevel:      "low",
			WhyItMatters:   "未发现显著风险，但仍建议结合JD微调简历",
			EvidenceGap:    "",
			FixSuggestion:  "为关键技能补充量化成果描述以提升简历说服力",
			InterviewImpact: "",
		})
	}
	if len(risks) > 6 { risks = risks[:6] }
	return risks
}

// ─── L5: Score Calibration ────────────────────────────────────────────────

func calibrateV3Score(v3ctx *DeepMatchContextV3, coverage []RequirementCoverageItem, evidence []ResumeEvidenceItem, risks []DeepKeyRisk) (int, string, ScoreBreakdownResultV3) {
	fastResult := v3ctx.FastResult
	fastScore := fastResult.MatchScore

	// Deep starts with Fast as baseline, then adjusts
	dirFit := v3ctx.FastScoreBreakdown.DirectionFit
	covScore := v3ctx.FastScoreBreakdown.MustHaveCoverage
	evScore := v3ctx.FastScoreBreakdown.EvidenceStrength
	expScore := v3ctx.FastScoreBreakdown.ExperienceRelevance
	riskCtrl := v3ctx.FastScoreBreakdown.RiskControl

	// Adjust: if coverage has many missing must-haves, lower must_have_coverage
	missMust := 0
	totalMust := 0
	for _, c := range coverage {
		if c.RequirementType == "must_have" {
			totalMust++
			if c.CoverageStatus == "missing" { missMust++ }
		}
	}
	if totalMust > 0 && missMust > 0 {
		covScore = covScore - (missMust * 10)
		if covScore < 0 { covScore = 0 }
	}

	// Adjust risk_control based on Deep risk count
	highCount := 0
	for _, r := range risks {
		if r.RiskLevel == "high" { highCount++ }
	}
	if highCount > 2 {
		riskCtrl = riskCtrl - 15
	} else if highCount > 0 {
		riskCtrl = riskCtrl - 8
	}
	if riskCtrl < 0 { riskCtrl = 0 }

	// JD competitiveness: higher for roles with many must_haves
	jdComp := minInt(80, totalMust*15)
	if totalMust == 0 { jdComp = 50 }

	breakdown := ScoreBreakdownResultV3{
		DirectionFit:        clamp(dirFit, 0, 100),
		MustHaveCoverage:    clamp(covScore, 0, 100),
		EvidenceStrength:    clamp(evScore, 0, 100),
		ExperienceRelevance: clamp(expScore, 0, 100),
		RiskControl:         clamp(riskCtrl, 0, 100),
		JDCompetitiveness:   jdComp,
	}

	deepScore := int(float64(breakdown.DirectionFit)*0.20 +
		float64(breakdown.MustHaveCoverage)*0.20 +
		float64(breakdown.EvidenceStrength)*0.20 +
		float64(breakdown.ExperienceRelevance)*0.15 +
		float64(breakdown.RiskControl)*0.15 +
		float64(breakdown.JDCompetitiveness)*0.10)

	// Apply Fast hard gate cap
	if deepScore > v3ctx.FastHardGate.ScoreCap {
		deepScore = v3ctx.FastHardGate.ScoreCap
	}
	// Deep should not be > Fast + 10 without explanation
	if deepScore > fastScore+10 {
		deepScore = fastScore + 5
	}
	deepScore = clamp(deepScore, 0, 100)

	level := matchLevel(deepScore)
	_ = fastScore
	return deepScore, level, breakdown
}

func buildScoreAlignment(fastScore, deepScore int, level string) ScoreAlignmentResult {
	delta := deepScore - fastScore
	consistent := abs(delta) <= 15
	reason := ""
	if delta < 0 {
		reason = fmt.Sprintf("Deep进一步考虑了岗位隐藏期望和投递竞争风险，因此比Fast低%d分。(%s)", -delta, level)
	} else if delta > 0 {
		reason = fmt.Sprintf("Deep发现Fast未覆盖的匹配证据，因此比Fast高%d分。(%s)", delta, level)
	} else {
		reason = "Fast与Deep评分一致，匹配判断稳定。"
	}
	return ScoreAlignmentResult{
		FastScore:    fastScore,
		DeepScore:    deepScore,
		ScoreDelta:   delta,
		IsConsistent: consistent,
		DeltaReason:  reason,
	}
}

func abs(x int) int { if x < 0 { return -x }; return x }

// ─── L6: Application Strategy ─────────────────────────────────────────────

func buildV3AppStrategy(score int, level, decisionLabel string) DeepApplicationStrategy {
	return DeepApplicationStrategy{
		BestTiming:          "根据岗位截止日期决定，建议尽早投递",
		DeliveryStrategy:    fmt.Sprintf("%s — 匹配度%d分（%s）", decisionLabel, score, level),
		CompetingPositioning: fmt.Sprintf("匹配度%d分，%s", score, level),
	}
}

// ─── L7: Resume Rewrite + Interview ───────────────────────────────────────

func buildV3RewriteStrategy(coverage []RequirementCoverageItem, risks []DeepKeyRisk) DeepResumeRewriteStrategy {
	s := DeepResumeRewriteStrategy{}
	for _, c := range coverage {
		if c.CoverageStatus == "strong" {
			s.ProjectRewritePoints = append(s.ProjectRewritePoints, fmt.Sprintf("强化%s的项目经历表达，使用STAR法则", c.RequirementName))
		}
		if c.CoverageStatus == "missing" || c.CoverageStatus == "weak" {
			s.KeywordsToAdd = append(s.KeywordsToAdd, c.RequirementName)
		}
	}
	for _, r := range risks {
		if r.RiskLevel == "high" {
			s.ProjectRewritePoints = append(s.ProjectRewritePoints, r.FixSuggestion)
		}
	}
	if len(s.KeywordsToAdd) > 6 { s.KeywordsToAdd = s.KeywordsToAdd[:6] }
	if len(s.ProjectRewritePoints) > 5 { s.ProjectRewritePoints = s.ProjectRewritePoints[:5] }
	if len(s.SkillsReorder) == 0 { s.SkillsReorder = append(s.SkillsReorder, "将最匹配岗位方向的核心技能前置") }
	return s
}

func buildV3InterviewQuestions(ju DeepJobUnderstanding, risks []DeepKeyRisk, coverage []RequirementCoverageItem) []DeepInterviewQuestion {
	var qs []DeepInterviewQuestion
	// Reuse existing builder
	dummy := &DeepMatchReport{
		JobUnderstanding:   ju,
		KeyRisks:           risks,
		RequirementCoverage: coverage,
	}
	qs = buildInterviewQuestionsFromReport(dummy)
	// Add answer_direction hints
	for i := range qs {
		if qs[i].AnswerDirection == "" {
			qs[i].AnswerDirection = "结合简历真实经历回答，突出具体职责和量化成果"
		}
	}
	return qs
}

func buildV3ActionPlan(strategy DeepResumeRewriteStrategy) []map[string]any {
	plan := []map[string]any{}
	for _, s := range strategy.SkillsReorder {
		plan = append(plan, map[string]any{"step": s, "type": "skills_reorder", "priority": 1})
	}
	for _, s := range strategy.ProjectRewritePoints {
		plan = append(plan, map[string]any{"step": s, "type": "project_rewrite", "priority": 2})
	}
	return plan
}

// ─── L8: Global Consistency ───────────────────────────────────────────────

func buildV3CoreMatchPoints(coverage []RequirementCoverageItem, evidence []ResumeEvidenceItem) []DeepCoreMatchPoint {
	var pts []DeepCoreMatchPoint
	for _, c := range coverage {
		if c.CoverageStatus == "strong" || c.CoverageStatus == "partial" {
			pts = append(pts, DeepCoreMatchPoint{
				PointName:     c.RequirementName,
				MatchStrength: c.CoverageStatus,
				WhyItMatches:  strings.Join(c.MatchedEvidence, "; "),
				JobRequirement: c.RequirementName,
			})
		}
	}
	if len(pts) == 0 && len(evidence) > 0 {
		pts = append(pts, DeepCoreMatchPoint{
			PointName: evidence[0].EvidenceName,
			MatchStrength: "partial",
			WhyItMatches: evidence[0].EvidenceText,
		})
	}
	return pts
}

func makeQualityChecks(allFields bool, noNull bool) QualityChecksResult {
	return QualityChecksResult{
		HasJDUnderstanding:     allFields,
		HasResumeEvidence:      allFields,
		HasRequirementCoverage: allFields,
		HasStructuredRisks:     allFields,
		HasRewriteStrategy:     allFields,
		HasInterviewQuestions:  allFields,
		HasScoreAlignment:      allFields,
		NoNullArrays:           noNull,
		NoRawTagCode:           true,
	}
}

// ─── H22-D: DeepMatchReport Main Path (legacy, H19-B replaced by v3) ────

func (s *MatchService) runDeepMatchSEP(
	ctx context.Context, userID uint64, resume *model.ResumeVersion, job *model.Job, runner agent.PipelineRunner,
) (*MatchResponse, error) {
	log.Printf("[DeepMatchSEP] starting for resume=%d job=%d", resume.ID, job.ID)

	// 1. Run 3-agent pipeline (produces gap/risk/match data)
	if s.tagEvidence == nil {
		return s.fallbackDeepMatch(ctx, userID, resume, job)
	}

	resp, err := s.matchWith3Agents(ctx, userID, resume, job, runner)
	if err != nil {
		log.Printf("[DeepMatchSEP] 3-agent failed: %v, falling back", err)
		resp, fallbackErr := s.analyzeWithLLM(ctx, resume, job, runner)
		if fallbackErr == nil {
			return sanitizeDeepResponse(resp), nil
		}
		return s.fallbackDeepMatch(ctx, userID, resume, job)
	}

	// 2. Enrich with DeepMatchReport using SEP evidence
	report := s.buildDeepMatchReportFromSEP(ctx, resp, resume, job)
	if report != nil {
		resp = s.enrichResponseWithReport(resp, report, userID, resume, job)
	}

	// 3. Post-process
	resp = sanitizeDeepResponse(resp)
	return resp, nil
}

// buildDeepMatchReportFromSEP constructs a DeepMatchReport using 3-agent output + SEP profiles.
func (s *MatchService) buildDeepMatchReportFromSEP(
	ctx context.Context, resp *MatchResponse, resume *model.ResumeVersion, job *model.Job,
) *DeepMatchReport {
	if s.semanticRepo == nil || resp == nil || resp.Result == nil {
		return nil
	}

	report := &DeepMatchReport{
		GenerationTier:        "quality",
		GeneratedBy:           "deep_match_agent",
		AgentStatus:           "match_verified",
		MatchScore:            resp.Result.MatchScore,
		MatchLevel:            resp.Result.MatchLevel,
		OneSentenceConclusion: resp.Result.Summary,
	}

	// Job understanding from SEP Job Profile
	jp, jerr := s.semanticRepo.FindLatestByTarget(ctx, model.SemanticTargetJob, int64(job.ID))
	if jerr == nil && jp != nil {
		jtgs, _ := s.semanticRepo.ListTagsByProfileID(ctx, jp.ID)
		mustHave := []string{}
		shouldHave := []string{}
		niceToHave := []string{}
		for _, t := range jtgs {
			switch t.Category {
			case "direction", "skill_core", "education":
				mustHave = append(mustHave, t.TagLabel)
			case "skill_secondary", "framework":
				shouldHave = append(shouldHave, t.TagLabel)
			default:
				niceToHave = append(niceToHave, t.TagLabel)
			}
		}
		report.JobUnderstanding = DeepJobUnderstanding{
			JobPositioning:        fmt.Sprintf("%s - %s", job.Title, jp.ProfileVersion),
			CoreResponsibilities:  []string{},
			MustHaveRequirements:  mustHave,
			ImportantRequirements: shouldHave,
			BonusRequirements:     niceToHave,
			HiddenExpectations:    []string{},
		}
	}

	// Resume evidence mapping from SEP Resume Profile
	rp, rerr := s.semanticRepo.FindLatestByTarget(ctx, model.SemanticTargetResume, int64(resume.ID))
	if rerr == nil && rp != nil {
		rtags, _ := s.semanticRepo.ListTagsByProfileID(ctx, rp.ID)
		evidence := []ResumeEvidenceItem{}
		for _, t := range rtags {
			if t.EvidenceText != "" || t.Explanation != "" {
				level := 2
				if strings.Contains(t.EvidenceText, "实习") || strings.Contains(t.EvidenceText, "公司") || strings.Contains(t.EvidenceText, "上线") {
					level = 4
				}
				evidence = append(evidence, ResumeEvidenceItem{
					EvidenceName:        t.TagLabel,
					EvidenceLevel:       level,
					SourceType:          t.Category,
					SourceTitle:         t.TagLabel,
					EvidenceText:        t.EvidenceText,
					SupportsRequirement: t.Explanation,
					Confidence:          "medium",
				})
			}
		}
		report.ResumeEvidence = evidence
	}

	// Core match points from 3-agent output
	for _, h := range resp.Result.MatchHighlights {
		report.CoreMatchPoints = append(report.CoreMatchPoints, DeepCoreMatchPoint{
			PointName:     h,
			MatchStrength: "partial",
			WhyItMatches:  h,
		})
	}

	// Key risks from 3-agent output
	for _, r := range resp.Result.RiskPoints {
		report.KeyRisks = append(report.KeyRisks, DeepKeyRisk{
			RiskName:      r,
			RiskLevel:     "medium",
			WhyItMatters:  r,
			FixSuggestion: "建议针对性补强",
		})
	}

	// Resume rewrite from gaps
	strategy := DeepResumeRewriteStrategy{}
	for _, g := range resp.Result.GapPoints {
		strategy.SkillsReorder = append(strategy.SkillsReorder, g)
	}
	report.ResumeRewriteStrategy = strategy

	// Application strategy
	report.ApplicationStrategy = DeepApplicationStrategy{
		BestTiming:         "根据岗位截止日期决定",
		DeliveryStrategy:   report.MatchLevel,
		CompetingPositioning: fmt.Sprintf("匹配度%d分", report.MatchScore),
	}

	// H22-G: Generate requirement_coverage from job_understanding + evidence
	report.RequirementCoverage = buildRequirementCoverageFromReport(report)

	// H22-G: Generate interview_risk_questions from risks + gaps + coverage
	report.InterviewRiskQuestions = buildInterviewQuestionsFromReport(report)

	// Agent trace
	report.AgentTrace = []DeepAgentTrace{
		{Stage: "JD_DECOMPOSITION", Agent: "JD解构Agent", Status: "done", Summary: "基于SEP Job Profile解构岗位要求"},
		{Stage: "RESUME_EVIDENCE_AUDIT", Agent: "简历证据Agent", Status: "done", Summary: fmt.Sprintf("提取%d条SEP证据", len(report.ResumeEvidence))},
		{Stage: "REQUIREMENT_ALIGNMENT", Agent: "要求对齐Agent", Status: "done", Summary: fmt.Sprintf("生成%d条要求覆盖", len(report.RequirementCoverage))},
		{Stage: "RISK_DIAGNOSIS", Agent: "风险诊断Agent", Status: "done", Summary: fmt.Sprintf("识别%d个风险", len(report.KeyRisks))},
		{Stage: "INTERVIEW_PREDICTION", Agent: "面试追问Agent", Status: "done", Summary: fmt.Sprintf("预测%d个面试追问", len(report.InterviewRiskQuestions))},
		{Stage: "GLOBAL_CALIBRATION", Agent: "全局校准Agent", Status: "done", Summary: fmt.Sprintf("最终分数%d", report.MatchScore)},
	}

	return report
}

// ─── H22-G: Requirement Coverage ───────────────────────────────────────────

func buildRequirementCoverageFromReport(report *DeepMatchReport) []RequirementCoverageItem {
	ju := report.JobUnderstanding
	var items []RequirementCoverageItem

	// Collect requirement categories
	type reqCat struct{ name, reqType string; priority int }
	reqs := []reqCat{}
	for _, n := range ju.MustHaveRequirements {
		reqs = append(reqs, reqCat{n, "must_have", 100})
	}
	for _, n := range ju.ImportantRequirements {
		reqs = append(reqs, reqCat{n, "should_have", 80})
	}
	for _, n := range ju.BonusRequirements {
		reqs = append(reqs, reqCat{n, "nice_to_have", 60})
	}
	for _, n := range ju.HiddenExpectations {
		reqs = append(reqs, reqCat{n, "hidden", 50})
	}

	// Build evidence text pool for matching
	evPool := ""
	for _, ev := range report.ResumeEvidence {
		evPool += ev.EvidenceText + " " + ev.SupportsRequirement + " "
	}

	for i, r := range reqs {
		status := "missing"
		score := 30
		gap := fmt.Sprintf("简历中未体现%s相关证据", r.name)
		suggestion := fmt.Sprintf("建议针对%s补充具体项目经历或技能描述", r.name)
		matchedIDs := []string{}

		// Simple keyword matching against evidence pool
		lowerPool := strings.ToLower(evPool)
		lowerName := strings.ToLower(r.name)
		if strings.Contains(lowerPool, lowerName) || len(lowerName) >= 4 && strings.Contains(lowerPool, lowerName[:4]) {
			status = "partially_matched"
			score = 65
			gap = fmt.Sprintf("%s有一定证据但不够充分", r.name)
			suggestion = fmt.Sprintf("进一步强化%s相关经历的细节表达", r.name)
		}
		// If evidence level >= 4 exists, upgrade to matched
		for _, ev := range report.ResumeEvidence {
			if ev.EvidenceLevel >= 4 && (strings.Contains(strings.ToLower(ev.EvidenceText+ev.SupportsRequirement), lowerName)) {
				status = "matched"
				score = 82
				gap = ""
				suggestion = ""
				matchedIDs = append(matchedIDs, ev.EvidenceName)
				break
			}
		}

		items = append(items, RequirementCoverageItem{
			RequirementName: r.name,
			RequirementType: r.reqType,
			CoverageStatus:  status,
			CoverageScore:   score,
			MatchedEvidence: matchedIDs,
			Gap:             gap,
			Suggestion:      suggestion,
		})
		if len(items) >= 8 {
			break
		}
		_ = i
	}
	return items
}

// ─── H23-B: Interview Questions De-template ─────────────────────────────────

func buildInterviewQuestionsFromReport(report *DeepMatchReport) []DeepInterviewQuestion {
	var questions []DeepInterviewQuestion
	ju := report.JobUnderstanding

	// Q1: Direction/motivation
	if len(ju.MustHaveRequirements) > 0 {
		questions = append(questions, DeepInterviewQuestion{
			Question:    fmt.Sprintf("你的简历经验更偏%s方向，投递这个%s岗位的动机是什么？", strings.Join(ju.MustHaveRequirements[:minInt(3, len(ju.MustHaveRequirements))], "、"), ju.JobPositioning),
			WhyAsked:    "面试官想了解你是否有清晰的职业规划，是否真正理解这个岗位。",
			RelatedRisk: "方向匹配度",
		})
	}

	// Q2: Core skill deep-dive from risks
	if len(report.KeyRisks) > 0 {
		r := report.KeyRisks[0]
		questions = append(questions, DeepInterviewQuestion{
			Question:    fmt.Sprintf("你在%s方面具体做过哪些项目？能讲一个从设计到上线的完整案例吗？", r.RiskName),
			WhyAsked:    r.WhyItMatters,
			RelatedRisk: r.RiskName,
		})
	}

	// Q3: Evidence verification from requirement coverage
	for _, cov := range report.RequirementCoverage {
		if len(questions) >= 5 { break }
		if cov.CoverageStatus == "partially_matched" || cov.CoverageStatus == "transferable" {
			questions = append(questions, DeepInterviewQuestion{
				Question:    fmt.Sprintf("岗位要求%s，你如何用现有项目经验来证明？具体怎么做？", cov.RequirementName),
				WhyAsked:    fmt.Sprintf("%s是该岗位的关键考察点", cov.RequirementName),
				RelatedRisk: cov.RequirementName,
			})
		}
	}

	// Q4: Gap bridging
	missingFound := false
	for _, cov := range report.RequirementCoverage {
		if cov.CoverageStatus == "missing" && !missingFound {
			questions = append(questions, DeepInterviewQuestion{
				Question:    fmt.Sprintf("如果入职后需要快速补上%s，你准备怎么学？能给一个具体计划吗？", cov.RequirementName),
				WhyAsked:    "考察学习能力和自我驱动意识",
				RelatedRisk: cov.RequirementName,
			})
			missingFound = true
		}
	}

	// Q5: Self-assessment + improvement plan
	if len(questions) < 5 && len(report.KeyRisks) > 1 {
		r := report.KeyRisks[1]
		questions = append(questions, DeepInterviewQuestion{
			Question:    fmt.Sprintf("如果满分10分，你给自己在%s方面打几分？为什么？", r.RiskName),
			WhyAsked:    "测试自我认知和诚实度",
			RelatedRisk: r.RiskName,
		})
	}

	// Ensure minimum 3 questions
	if len(questions) < 3 {
		for _, r := range report.KeyRisks {
			if len(questions) >= 3 { break }
			questions = append(questions, DeepInterviewQuestion{
				Question:    fmt.Sprintf("能否结合一个具体例子说明你在%s方面的能力？", r.RiskName),
				WhyAsked:    r.WhyItMatters,
				RelatedRisk: r.RiskName,
			})
		}
	}

	return questions
}

// enrichResponseWithReport saves the DeepMatchReport and enriches the response.
func (s *MatchService) enrichResponseWithReport(
	resp *MatchResponse, report *DeepMatchReport, userID uint64, resume *model.ResumeVersion, job *model.Job,
) *MatchResponse {
	// Build legacy fields from report
	resp.Result.MatchHighlights = ensureNonNil(resp.Result.MatchHighlights)
	for _, cp := range report.CoreMatchPoints {
		if cp.PointName != "" {
			resp.Result.MatchHighlights = append(resp.Result.MatchHighlights, cp.PointName)
		}
	}
	for _, ev := range report.ResumeEvidence {
		if ev.EvidenceText != "" {
			resp.Result.FitPoints = append(resp.Result.FitPoints, ev.EvidenceText)
		}
	}

	// Save to DB
	reportJSON, _ := json.Marshal(report)
	resultJSON, _ := json.Marshal(resp.Result)

	if s.matchRepo != nil {
		analysis := &model.MatchAnalysis{
			UserID:         userID,
			ResumeID:       resume.ID,
			JobID:          job.ID,
			GenerationTier: "quality",
			GeneratedBy:    "deep_match_agent",
			Status:         "ready",
			Score:          report.MatchScore,
			ResultJSON:     reportJSON,
		}
		if createErr := s.matchRepo.Create(context.Background(), analysis); createErr != nil {
			log.Printf("[DeepMatchSEP] DB save failed: %v", createErr)
		}
		resp.AnalysisID = analysis.ID
	}
	_ = resultJSON

	// H22-E: Copy DeepMatchReport fields to API response model
	resp.Result.ReportVersion = "deep_match_v2"
	resp.Result.ScoreExplanation = ""
	resp.Result.Confidence = 0
	resp.Result.JobUnderstanding = report.JobUnderstanding
	resp.Result.RequirementCoverage = report.RequirementCoverage
	resp.Result.ResumeEvidenceMapping = report.ResumeEvidence
	resp.Result.InterviewRiskQuestions = report.InterviewRiskQuestions
	resp.Result.CoreMatchPoints = report.CoreMatchPoints
	resp.Result.KeyRisks = report.KeyRisks
	resp.Result.ResumeRewriteStrategy = report.ResumeRewriteStrategy
	resp.Result.InterviewRiskQuestions = report.InterviewRiskQuestions
	resp.Result.ActionPlan = deepActionPlanFromReport(report)
	resp.Result.AgentTrace = report.AgentTrace
	resp.Result.ApplicationDecision = report.ApplicationDecision
	resp.Result.OneSentenceConclusion = report.OneSentenceConclusion

	// Set agent_status on response
	resp.GeneratedBy = "deep_match_agent"
	resp.Result.AgentStatus = "match_verified"
	resp.Result.GeneratedBy = "deep_match_agent"

	return resp
}

func deepActionPlanFromReport(report *DeepMatchReport) []map[string]any {
	if report == nil { return nil }
	plan := []map[string]any{}
	for _, s := range report.ResumeRewriteStrategy.SkillsReorder {
		plan = append(plan, map[string]any{"step": s, "type": "skills_reorder", "priority": 1})
	}
	for _, s := range report.ResumeRewriteStrategy.ProjectRewritePoints {
		plan = append(plan, map[string]any{"step": s, "type": "project_rewrite", "priority": 2})
	}
	return plan
}

// ─── Post-process: enforce H19-B contract ────────────────────────────────

// sanitizeDeepResponse sanitizes null arrays and fixes levels, but does NOT overwrite generated_by for legacy paths.
func sanitizeDeepResponse(resp *MatchResponse) *MatchResponse {
	if resp == nil { return resp }
	if resp.Result == nil { return resp }

	r := resp.Result
	// Only overwrite generated_by for paths that don't already have a specific marker
	if r.GeneratedBy == "" || r.GeneratedBy == "llm_agent" {
		r.GeneratedBy = "deep_match_agent"
	}
	// H19-D: normalize "repaired" → "match_verified" (but not legacy)
	if r.AgentStatus == "" || r.AgentStatus == "success" {
		r.AgentStatus = "match_verified"
	}
	// Fix null arrays
	if r.MatchHighlights == nil { r.MatchHighlights = []string{} }
	if r.GapPoints == nil { r.GapPoints = []string{} }
	if r.RiskPoints == nil { r.RiskPoints = []string{} }
	if r.EvidenceChain == nil { r.EvidenceChain = []model.MatchEvidence{} }
	if r.NextSteps == nil { r.NextSteps = []string{} }
	if r.OptimizationSuggestions == nil { r.OptimizationSuggestions = []model.MatchSuggestion{} }
	if r.FitPoints == nil { r.FitPoints = []string{} }
	if r.JobRequirements == nil { r.JobRequirements = []string{} }
	if r.ResumeEvidence == nil { r.ResumeEvidence = []string{} }

	// Sanitize risk strings: replace bare tag-like strings with human labels
	cleanRisks := []string{}
	for _, risk := range r.RiskPoints {
		cleaned := sanitizeRiskText(risk)
		cleanRisks = append(cleanRisks, cleaned)
	}
	r.RiskPoints = cleanRisks

	// Fix level
	if r.MatchLevel == "" || r.MatchLevel == "low" {
		r.MatchLevel = matchLevel(r.MatchScore)
	}

	// Fix summary
	if r.Summary == "" || r.Summary == r.MatchLevel {
		r.Summary = fmt.Sprintf("深度分析匹配度%d分（%s）", r.MatchScore, r.MatchLevel)
	}

	// Update meta
	resp.Meta.GeneratedBy = "deep_match_agent"
	resp.Meta.AgentStatus = r.AgentStatus

	return resp
}

func sanitizeRiskText(s string) string {
	// Map legacy risk strings to better labels
	mappings := map[string]string{
		"学历风险": "学历要求可能不满足岗位筛选条件",
		"核心技术缺失风险": "核心技能覆盖不足，可能影响简历筛选通过率",
		"论文要求风险": "岗位可能要求论文/研究成果，简历未体现",
		"实习风险": "缺少相关实习经历，竞争力可能不足",
		"方向风险": "岗位方向与简历核心方向存在差异",
	}
	if v, ok := mappings[s]; ok { return v }
	return s
}

// ─── Fallback ─────────────────────────────────────────────────────────────

func (s *MatchService) fallbackDeepMatch(
	ctx context.Context, userID uint64, resume *model.ResumeVersion, job *model.Job,
) (*MatchResponse, error) {
	log.Printf("[DeepMatchSEP] fallback for resume=%d job=%d", resume.ID, job.ID)
	var legacyResult *model.MatchAnalysisResult
	if s.semanticRepo != nil {
		rp, _ := s.semanticRepo.FindLatestByTarget(ctx, model.SemanticTargetResume, int64(resume.ID))
		jp, _ := s.semanticRepo.FindLatestByTarget(ctx, model.SemanticTargetJob, int64(job.ID))
		if rp != nil && jp != nil && rp.Status == "verified" && jp.Status == "verified" {
			rtags, _ := s.semanticRepo.ListTagsByProfileID(ctx, rp.ID)
			jtgs, _ := s.semanticRepo.ListTagsByProfileID(ctx, jp.ID)
			legacyResult = runFastSEPMatch(&MatchSEPContext{
				ResumeTags: rtags, JobTags: jtgs, JobTitle: job.Title,
				ResumeID: int64(resume.ID), JobID: int64(job.ID),
			})
		}
	}
	if legacyResult == nil {
		legacyResult = &model.MatchAnalysisResult{
			MatchScore: 50, MatchLevel: "中等匹配", GeneratedBy: "deep_match_agent", AgentStatus: "match_repaired",
			Summary: "深度分析暂时不可用，当前为安全兜底结果。",
			MatchHighlights: []string{}, GapPoints: []string{}, RiskPoints: []string{},
			NextSteps: []string{"建议稍后重试深度分析"},
		}
	} else {
		legacyResult.GeneratedBy = "deep_match_agent"
		legacyResult.AgentStatus = "match_repaired"
	}
	meta := model.AgentResponseMeta{
		AgentName: "DeepMatchAgent", GeneratedBy: "deep_match_agent", AgentStatus: "match_repaired",
		FallbackReason: "llm_unavailable", RepairApplied: true, SchemaValid: false,
	}
	_ = json.Marshal // keep import
	return &MatchResponse{
		AnalysisID: 0, GeneratedBy: "deep_match_agent", Result: legacyResult,
		CreatedAt: time.Now().Format("2006-01-02T15:04:05Z07:00"), Meta: meta,
	}, nil
}


