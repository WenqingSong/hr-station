# Prompt / Schema 战略一致性核验与改动评审

> 创建时间：2026-05-30
> 所属阶段：4-3.7.C
> 依据文档：战略差异化与杀手锏确认、投递边界与优化停止条件设计、JD证据链匹配与START简历优化规则、Agent输出QualityGate设计、学生求职极端情况处理边界、灵活在线简历Profile与纯文字Agent上下文设计、阶段4-3.5战略转型汇总与工程评审

## 1. 背景

项目定位已从"岗位推荐工具"升级为"面向学生的投递前决策与简历初筛优化系统"。

4-3.5.B ~ 4-3.5.G 完成了三个杀手锏和配套设计的战略文档，但 **Prompt 和 Schema 尚未同步更新**。

本阶段目标是核验当前 Prompt / Schema 与战略设计之间的一致性，判断是否可以在 4-3.8 之前不做修改直接进入集成验收，还是需要插入 4-3.7.C.1 最小修复阶段。

**本阶段为 read-only + docs，不修改 Prompt / Schema / 代码。**

## 2. 核验范围

| 维度 | 来源文档 | 对象 |
|------|----------|------|
| 投递边界 | 4-3.5.C 投递边界与优化停止条件设计 | Prompt / Schema 是否包含 apply_readiness |
| 优化停止条件 | 4-3.5.C 投递边界与优化停止条件设计 | Prompt / Schema 是否包含 optimization_stop_signal |
| JD 证据链 | 4-3.5.D JD证据链匹配与START简历优化规则 | Prompt / Schema 是否包含 evidence_chain |
| 示例参考与用户核实 | 4-3.5.D JD证据链匹配与START简历优化规则 | Prompt / Schema 是否支持 rewrite_example + verification |
| START / STAR-T | 4-3.5.D JD证据链匹配与START简历优化规则 | Prompt 是否要求 START 表达法 |
| Quality Gate | 4-3.5.E Agent输出QualityGate设计 | Schema 是否包含 quality_gate_result |
| 极端情况边界 | 4-3.5.F 学生求职极端情况处理边界 | Prompt / Schema 是否包含 candidate_case_type |
| ResumeContext 输入 | 4-3.5.G 灵活在线简历Profile与纯文字Agent上下文设计 | Prompt 输入是否使用 ResumeContextText / ResumeContextPayload |
| Fallback / generated_by | 4-3 Agent 接入 Pipeline | Schema 是否正确区分 llm_agent / mock_agent / llm_agent_fallback |
| 不替用户做人生决策 | 4-3.5.B 战略差异化 | Prompt 是否避免职业规划式判断 |

## 3. 当前 Prompt / Schema 文件清单

| 文件 | 类型 | 行数 |
|------|------|------|
| `server/prompts/match_analysis_v1.md` | Prompt | 1041 |
| `server/prompts/recommendation_v1.md` | Prompt | 988 |
| `server/prompts/resume_optimization_v1.md` | Prompt | 1073 |
| `server/prompts/json_repair_v1.md` | Prompt | — |
| `server/prompts/schema/match_analysis_schema.json` | Schema | 541 |
| `server/prompts/schema/recommendation_schema.json` | Schema | 541 |
| `server/prompts/schema/resume_optimization_schema.json` | Schema | 618 |

## 4. 战略一致性总览

| 战略维度 | MatchAnalysis | Recommendation | ResumeOptimization | 整体 |
|----------|:---:|:---:|:---:|:---:|
| 投递边界 (apply_readiness) | ✅ | ✅ | ✅ | 已支持（4-3.7.C.1） |
| 优化停止条件 (optimization_stop_signal) | ✅ | ✅ | ✅ | 已支持（4-3.7.C.1） |
| JD 证据链 (evidence_chain) | ✅ | ✅ | ❌ (间接) | 部分支持 |
| 示例参考与用户核实 | ✅ | ⚠️ | ✅ | 部分支持 |
| START / STAR-T 表达法 | ❌ | ❌ | ❌ | 不支持 |
| Quality Gate | ❌ | ❌ | ❌ | 不支持 |
| 极端情况边界 | ❌ | ❌ | ❌ | 不支持 |
| ResumeContext 输入 | ❌ | ❌ | ❌ | 不支持 |
| Fallback / generated_by | ✅ | ✅ | ✅ | 已支持 |
| 不替用户做人生决策 | ✅ | ✅ | ✅ | 已支持 |

## 5. MatchAnalysisAgent 评审

### 5.1 已支持的战略能力

| 能力 | 字段 | 评价 |
|------|------|------|
| JD 证据链 | `evidence_chain` (jd_requirement → resume_evidence → judgement → explanation) | ✅ 完整 |
| JD 关键要求分析 | `jd_key_requirements` (requirement, importance, source, reason) | ✅ 完整 |
| 简历证据汇总 | `resume_evidence_summary` (evidence, related_requirement, strength_level) | ✅ 完整 |
| 匹配风险 | `risk_points` (risk, severity, reason, mitigation) | ✅ 完整 |
| 缺口分析 | `gap_points` (gap, impact, how_to_improve) | ✅ 完整 |
| 简历优化建议 | `resume_optimization` (P0/P1/P2, rewrite_example, action) | ✅ 完整 |
| 改写示例真实性 | `rewrite_example` 必须基于真实经历，条件式表达 | ✅ 完整 |
| 真实性提示 | `authenticity_notice` | ✅ 完整 |
| 分析依据 | `analysis_basis` (job_description_used, data_quality) | ✅ 完整 |
| 不承诺通过初筛 | Prompt 明确要求 | ✅ 完整 |
| 不虚构经历 | Prompt 明确要求 | ✅ 完整 |
| 不做人生决策 | Prompt 明确要求 | ✅ 完整 |
| generated_by | Schema 枚举 `llm_agent` | ✅ 完整 |
| 投递建议 | `match_level` + `screening_result` 间接表达 | ⚠️ 部分支持 |

### 5.2 未支持的战略能力

| 缺口 | 严重度 | 说明 |
|------|--------|------|
| `apply_readiness` 字段缺失 | 🔴 高 | 战略杀手锏一的核心字段，当前 match_level + screening_result 可以部分代替，但语义不同 |
| `optimization_stop_signal` 字段缺失 | 🔴 高 | 战略杀手锏一的核心字段，当前无等价表达 |
| `candidate_case_type` 字段缺失 | 🟡 中 | 极端情况识别能力，当前无法表达 |
| `case_risk_level` 字段缺失 | 🟡 中 | 极端情况风险等级 |
| `case_handling_strategy` 字段缺失 | 🟡 中 | 极端情况处理策略 |
| `quality_gate_result` 字段缺失 | 🟡 中 | Quality Gate 输出，当前 Schema 不包含 |
| `verification_note` 字段缺失 | 🟢 低 | resume_optimization 中的 rewrite_example 没有 verification_note |
| START / STAR-T 表达规则 | 🟡 中 | Prompt 未提及 START 表达法 |
| ResumeContextText 输入 | 🟡 中 | 当前使用 RESUME_CONTEXT_JSON 旧格式 |

### 5.3 等价替代评估

| 战略字段 | 当前等价 | 可替代程度 |
|----------|----------|-----------|
| `apply_readiness` | `match_level` + `screening_result` | ⚠️ 部分可替代：strong_match ≈ ready_to_apply, not_recommended ≈ not_ready_for_this_role，但缺少 quick_fix_before_apply 和 insufficient_information 语义 |
| `optimization_stop_signal` | 无 | ❌ 不可替代 |
| `candidate_case_type` | 无 | ❌ 不可替代 |

### 5.4 结论

**需要最小修复**：`apply_readiness` 和 `optimization_stop_signal` 缺失会明显影响 4-3.8 验收中"投递边界"这一核心战略价值的验证。但 MatchAnalysis 其余能力已非常完整，修复范围可控。

---

## 6. RecommendationAgent 评审

### 6.1 已支持的战略能力

| 能力 | 字段 | 评价 |
|------|------|------|
| 岗位证据链 | `evidence` (job_requirement → resume_evidence → judgement → explanation) | ✅ 完整 |
| 推荐级别 | `recommend_level` + `recommend_label` (四级别枚举) | ✅ 完整 |
| 风险点 | `risk_points` (risk, severity, mitigation) | ✅ 完整 |
| 查询建议 | `query_suggestions` (query, why_search) | ✅ 完整 |
| 下一步行动 | `next_action` (action, priority, expected_effect) | ✅ 完整 |
| 真实性提示 | `authenticity_notice` | ✅ 完整 |
| 分析依据 | `analysis_basis` | ✅ 完整 |
| 不编造岗位 | Prompt 明确要求 | ✅ 完整 |
| 不虚构经历 | Prompt 明确要求 | ✅ 完整 |
| generated_by | Schema 枚举 `llm_agent` | ✅ 完整 |

### 6.2 未支持的战略能力

| 缺口 | 严重度 | 说明 |
|------|--------|------|
| `apply_readiness` per job 缺失 | 🔴 高 | 每个岗位应独立判断投递准备度 |
| `optimization_stop_signal` 缺失 | 🔴 高 | 整体推荐结果缺少"什么时候可以停止优化"信号 |
| `candidate_case_type` 缺失 | 🟡 中 | 极端情况识别 |
| `quality_gate_result` 缺失 | 🟡 中 | Quality Gate 输出 |
| START / STAR-T 规则 | 🟡 中 | 未提及 |
| ResumeContextText 输入 | 🟡 中 | 旧格式 |
| `verification_note` | 🟢 低 | evidence 中无 verification_note |
| 避免同质化岗位推荐 | ⚠️ 部分 | Prompt 已要求"不能把所有岗位都标成优先投递"，但未显式要求差异化理由 |

### 6.3 结论

**需要最小修复**：与 MatchAnalysis 相同，缺少 `apply_readiness` 和 `optimization_stop_signal`。其余能力完整。

---

## 7. ResumeOptimizationAgent 评审

### 7.1 已支持的战略能力

| 能力 | 字段 | 评价 |
|------|------|------|
| P0/P1/P2 优先级 | `priority_suggestions` (p0/p1/p2) | ✅ 完整 |
| JD 关键词对齐 | `keyword_alignment` (keyword, status, resume_evidence, suggested_action) | ✅ 完整 |
| 项目优化 | `project_optimization` (project_name, current_problem, targeted_improvement, jd_relevance, rewrite_direction, interview_risk) | ✅ 完整 |
| 技能栏优化 | `skill_section_optimization` (skill, current_status, suggested_positioning, evidence_needed) | ✅ 完整 |
| 摘要优化 | `summary_optimization` | ✅ 完整 |
| 缺失证据计划 | `missing_evidence_plan` (should_not_fake = true) | ✅ 完整 |
| 改写示例 + 用户核实 | `rewrite_examples` (can_use_directly, requires_user_verification, verification_note) | ✅ **完整** — 三个 Agent 中唯一完整支持 |
| 真实性约束 | `authenticity_rule` per suggestion + `authenticity_notice` | ✅ 完整 |
| 初筛关注点 | `screening_focus` | ✅ 完整 |
| 不虚构经历 | Prompt 明确要求 | ✅ 完整 |
| 不承诺通过初筛 | Prompt 明确要求 | ✅ 完整 |
| generated_by | Schema 枚举 `llm_agent` | ✅ 完整 |
| 接受 MatchAnalysis 结果 | `MATCH_ANALYSIS_JSON` 可选输入 | ✅ 完整 |

### 7.2 未支持的战略能力

| 缺口 | 严重度 | 说明 |
|------|--------|------|
| `optimization_stop_signal` 缺失 | 🔴 高 | 简历优化场景最需要"什么时候停止"信号 |
| `apply_readiness` 缺失 | 🔴 高 | 简历优化后应表达投递准备度 |
| Pipeline 未接入 | 🔴 高 | 当前仅 Mock Agent，4-3.7 已补业务链路但未接入 Pipeline |
| `quality_gate_result` 缺失 | 🟡 中 | Quality Gate 输出 |
| `candidate_case_type` 缺失 | 🟡 中 | 极端情况识别 |
| START / STAR-T 规则 | 🟡 中 | 未提及 |
| ResumeContextText 输入 | 🟡 中 | 旧格式 |
| quick_fix vs long_term 区分 | 🟢 低 | P0/P1/P2 间接表达，但无显式 quick_fix / long_term 标签 |

### 7.3 结论

**需要最小修复**：`apply_readiness` 和 `optimization_stop_signal` 缺失影响最大。Pipeline 未接入已在 4-3.7 记录，4-3.8 处理。

---

## 8. 投递边界支持情况

| Agent | apply_readiness | 等价替代 | 结论 |
|-------|----------------|----------|------|
| MatchAnalysis | ✅ 已添加 | — | **已支持** |
| Recommendation | ✅ 已添加（per job） | — | **已支持** |
| ResumeOptimization | ✅ 已添加 | — | **已支持** |

**判断**：`apply_readiness` 缺失是战略一致性最大缺口，但 match_level / recommend_level 在语义上可以覆盖 4 种投递准备度中的 3 种（ready / not_ready / insufficient_info），只有 `quick_fix_before_apply` 缺少显式表达。

## 9. 优化停止条件支持情况

| Agent | optimization_stop_signal | 等价替代 | 结论 |
|-------|------------------------|----------|------|
| MatchAnalysis | ✅ 已添加 | — | **已支持** |
| Recommendation | ✅ 已添加（per job） | — | **已支持** |
| ResumeOptimization | ✅ 已添加 | — | **已支持** |

**判断**：`optimization_stop_signal` 是三个 Agent 统一缺失的字段，且无等价替代。这是"投递前决策系统"最核心的差异化能力。

## 10. JD 证据链支持情况

| Agent | evidence_chain | 评价 |
|-------|---------------|------|
| MatchAnalysis | ✅ `evidence_chain` (jd_requirement → resume_evidence → judgement → explanation) | **已支持** |
| Recommendation | ✅ `evidence` (job_requirement → resume_evidence → judgement → explanation) | **已支持** |
| ResumeOptimization | ⚠️ 间接：通过 `keyword_alignment` + `priority_suggestions.related_jd_requirement` | **部分支持** |

**判断**：MatchAnalysis 和 Recommendation 的证据链已完整。ResumeOptimization 通过输入 MatchAnalysis 的结果间接获取证据链。

## 11. 示例参考与用户核实支持情况

| Agent | rewrite_example | verification_note | can_use_directly | requires_user_verification | 结论 |
|-------|----------------|-------------------|-----------------|---------------------------|------|
| MatchAnalysis | ✅ | ❌ | ❌ | ❌ | 部分支持 |
| Recommendation | ❌ | ❌ | ❌ | ❌ | 不支持 |
| ResumeOptimization | ✅ | ✅ | ✅ | ✅ | **已支持** |

**判断**：ResumeOptimization 的用户核实机制是三个 Agent 中最完整的（`can_use_directly` / `requires_user_verification` / `verification_note`）。MatchAnalysis 有 rewrite_example 但缺少 verification 机制。Recommendation 无改写场景，不需要 verification。

## 12. START / STAR-T 支持情况

三个 Agent 的 Prompt 均未提及 START / STAR-T 表达法。

4-3.5.D 文档定义了 START 表达法（Situation / Task / Action / Result / Technology），要求用于简历经历表达完整性检查。

**判断**：**不支持**。但 START 表达法更偏向前端展示和 Prompt 引导层面的增强，不影响 4-3.8 的 LLM 输出结构验收。

## 13. Quality Gate 支持情况

4-3.5.E 定义了 Quality Gate 输出字段：

- `quality_gate_result`: pass / revise / conservative_fallback / mock_fallback
- `self_check_items`: 十项检查维度
- `revise_once` 约束
- `anti_hallucination` 检查

当前三个 Schema 均不包含 `quality_gate_result` 字段。

**判断**：**不支持**。但 Quality Gate 是 Pipeline 层面（代码层）的检查，不一定要嵌入 LLM 输出 Schema。可以在 Pipeline 代码中实现 Quality Gate 检查逻辑，而不修改 Schema。这不阻塞 4-3.8。

## 14. 极端情况边界支持情况

4-3.5.F 定义了：

- `candidate_case_type`: 十种枚举
- `case_risk_level`: 四级枚举
- `case_handling_strategy`: 七种策略

当前三个 Schema 均不包含这些字段。

**判断**：**不支持**。但极端情况处理更多是 Prompt 引导 + Mock fallback 策略，不需要 Schema 强制字段。可以在 Prompt 中增加极端情况处理指引，而不修改 Schema 结构。不阻塞 4-3.8。

## 15. ResumeContextText 输入一致性

4-3.5.G 定义了：

- `ResumeContextText`: 纯文字简历上下文
- `ResumeContextPayload`: ResumeContextText + profile_summary + privacy_flags
- 隐私字段处理规则

当前三个 Prompt 均使用 `{{RESUME_CONTEXT_JSON}}` 占位符，未使用 `ResumeContextText` / `ResumeContextPayload`。

**判断**：**不支持**。但这是输入层改造，影响 Pipeline 代码中的 PromptValues 构建逻辑，不影响 Schema 输出结构。可以在 Pipeline 代码中做格式转换，不阻塞 4-3.8。

## 16. 当前缺口汇总

| 缺口 | 严重度 | 影响范围 | 是否阻塞 4-3.8 | 建议处理阶段 |
|------|--------|----------|----------------|-------------|
| `apply_readiness` 字段缺失 | 🔴 高 | 三个 Agent | ⚠️ 影响"投递边界"验收 | C.1 最小修复 |
| `optimization_stop_signal` 字段缺失 | 🔴 高 | 三个 Agent | ⚠️ 影响"优化停止条件"验收 | C.1 最小修复 |
| `verification_note` (Match/Rec) 缺失 | 🟢 低 | Match/Rec | ❌ 不阻塞 | 后续增强 |
| START / STAR-T 规则 | 🟡 中 | Prompt 层面 | ❌ 不阻塞 Schema 验收 | 后续增强 |
| Quality Gate 字段 | 🟡 中 | Pipeline 代码层面 | ❌ 不阻塞 Schema 验收 | 4-3.8 Pipeline 实现 |
| 极端情况字段 | 🟡 中 | Prompt + Mock 层面 | ❌ 不阻塞 Schema 验收 | 后续增强 |
| ResumeContextText 输入 | 🟡 中 | Pipeline 代码层面 | ❌ 不阻塞 Schema 验收 | 后续增强 |
| ResumeOptimization Pipeline 未接入 | 🔴 高 | 业务代码层面 | ⚠️ 不阻塞 Match/Rec 验收 | 4-3.8 |

## 17. 是否需要 4-3.7.C.1

**需要插入 4-3.7.C.1 Prompt / Schema 最小一致性修复。**

理由：

1. `apply_readiness` 是战略杀手锏一"投递边界"的核心字段，三个 Agent 全部缺失
2. `optimization_stop_signal` 是"优化停止条件"的核心字段，三个 Agent 全部缺失
3. 缺少这两个字段，4-3.8 验收时无法验证"投递前决策系统"这一核心定位
4. 修复范围可控：仅需要在三个 Schema 中新增 2 个字段，在三个 Prompt 中新增对应指引
5. 其余缺口（Quality Gate / 极端情况 / ResumeContextText / START）不阻塞 4-3.8，可在后续阶段增强

### C.1 最小修复范围

| 修改 | Agent | 范围 |
|------|-------|------|
| 新增 `apply_readiness` 字段 | Match / Rec / Resume | Schema + Prompt |
| 新增 `optimization_stop_signal` 字段 | Match / Rec / Resume | Schema + Prompt |
| 新增 `quick_fix_items` 字段 | ResumeOptimization | Schema + Prompt（区分 quick fix 和 long term） |

**不修改**：Quality Gate / 极端情况 / ResumeContextText / START / 代码 / 前端

## 18. 是否可以进入 4-3.8

**完成 C.1 最小修复后可以进入 4-3.8。**

当前状态评估：

| Agent | 当前可否进入 4-3.8 | C.1 后可否进入 |
|-------|-------------------|---------------|
| MatchAnalysis | ⚠️ 缺投递边界/停止条件，其余完整 | ✅ |
| Recommendation | ⚠️ 缺投递边界/停止条件，其余完整 | ✅ |
| ResumeOptimization | ⚠️ 缺投递边界/停止条件 + Pipeline 未接入 | ✅ (C.1 后进入 4-3.8) |

## 19. 下一步建议

1. **进入 4-3.7.C.1**：Prompt / Schema 最小一致性修复
   - 三个 Schema 新增 `apply_readiness` + `optimization_stop_signal`
   - 三个 Prompt 新增投递边界和优化停止条件指引
   - ResumeOptimization Schema 新增 `quick_fix_items`
2. **C.1 完成后进入 4-3.8**：ResumeOptimizationAgent 接入 Pipeline
3. **后续增强项**（不阻塞主线）：
   - Match/Rec Schema 增加 `verification_note`
   - Prompt 增加 START / STAR-T 表达规则
   - Pipeline 实现 Quality Gate 检查逻辑
   - Prompt 增加极端情况处理指引
   - Pipeline 切换到 ResumeContextText 输入

## 20. 4-3.7.C.1 完成记录

4-3.7.C.1 已完成 Prompt / Schema 最小一致性修复。

### 修改文件

| 文件 | 修改内容 |
|------|---------|
| `server/prompts/schema/match_analysis_schema.json` | 新增 `apply_readiness` / `optimization_stop_signal` / `top_fix_items` 三个 required 字段及 `$defs/top_fix_item` 定义 |
| `server/prompts/schema/recommendation_schema.json` | 在 `recommendation_item` 内新增 `apply_readiness` / `optimization_stop_signal` / `top_fix_items` 及 `$defs/top_fix_item` 定义 |
| `server/prompts/schema/resume_optimization_schema.json` | 新增 `apply_readiness` / `optimization_stop_signal` / `top_fix_items` 三个 required 字段及 `$defs/top_fix_item` 定义 |
| `server/prompts/match_analysis_v1.md` | 新增 7.5 apply_readiness / 7.6 optimization_stop_signal / 7.7 top_fix_items 指引；更新 Output JSON Structure / Enum Constraints / 数组长度要求 |
| `server/prompts/recommendation_v1.md` | 新增 Section 9 Apply Readiness and Stop Signal（per job）；更新 Output JSON Structure / Enum Constraints |
| `server/prompts/resume_optimization_v1.md` | 新增 Section 6 Apply Readiness and Stop Signal；更新 Output JSON Structure / Enum Constraints / 数组长度要求 |
| `server/internal/schema/validator_test.go` | 测试数据新增三个字段 |
| `server/internal/agent/repair/json_repair_agent_test.go` | 测试数据新增三个字段 |
| `server/internal/agent/pipeline/pipeline_test.go` | 测试数据新增三个字段 |

### 验证结果

- 三个 JSON Schema 文件 `python3 -m json.tool` 验证通过
- `go test ./...` 全部通过
- grep 验证 `apply_readiness` / `optimization_stop_signal` / `top_fix_items` 在 6 个文件（3 schema + 3 prompt）中均已覆盖

## 21. 4-3.7-fix.2 完成记录

4-3.7-fix.2 已完成 Go 契约最小一致性补齐。

### 修改文件

| 文件 | 修改内容 |
|------|---------|
| `server/internal/model/match.go` | 新增 `TopFixItem` 结构体；`MatchAnalysisResult` 新增 `ApplyReadiness` / `OptimizationStopSignal` / `TopFixItems` |
| `server/internal/model/recommendation.go` | `RecommendedJob` 新增 `ApplyReadiness` / `OptimizationStopSignal` / `TopFixItems` |
| `server/internal/agent/match_agent.go` | `mapParsedToMatchResult` 新增三字段映射；新增 `mapTopFixItems` 函数 |
| `server/internal/agent/recommendation_agent.go` | `mapParsedToRecommendedJobs` 新增三字段映射；新增 `mapRecTopFixItems` 函数 |
| `server/internal/agent/match_mock.go` | 新增 `determineMatchApplyReadiness` / `generateMatchTopFixItems`；Mock 结果包含新字段 |
| `server/internal/agent/recommendation_mock.go` | 新增 `determineRecApplyReadiness` / `generateRecTopFixItems`；Mock 结果包含新字段 |
| `server/internal/agent/match_agent_test.go` | `minimalLLMOutput` 新增三字段 fixture |
| `server/internal/agent/recommendation_agent_test.go` | `minimalRecParsedOutput` 及其他测试 fixture 新增三字段 |

### 验证结果

- `go build ./...` 编译通过
- `go test ./...` 全部通过
- `make llm-check` 通过（未调用真实 LLM）
- Prompt / Schema / Go model / Agent mapping / mock / fallback / fixture 已保持一致
- fix.2 完成后 C.1 判定升级为 pass，可进入 fix.3 最终闸门验收
