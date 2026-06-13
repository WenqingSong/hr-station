# Deep-Match-Schema-And-Prompt-Contract-Freeze-H19A 完成报告

> 执行时间：2026-06-13 23:16
> 类型：Deep Match Schema/Prompt/Agent 流程契约冻结（只读审计，不修改代码）
> 状态：✅ 冻结完成，H19-B 可执行

---

## 1. 总体结论

| 项目 | 结果 |
|------|------|
| 是否完成 H19-A | ✅ 冻结完成 |
| 是否修改业务代码 | ❌ 未修改 |
| 是否修改前端 | ❌ 未修改 |
| 是否修改数据库 | ❌ 未修改 |
| 是否执行 Git | ❌ 未执行 |
| 当前 Deep 是否仍依赖 TagEvidenceService | ✅ **是** — 3-Agent pipeline 使用 TagEvidenceService |
| 当前 Deep 是否读取 SEP | ⚠️ **部分** — DeepMatchReport 读取 SEP，但 3-Agent 不读 |
| 当前 Deep 是否读取 Fast 初判 | ❌ **否** — FastResult 字段声明但从未赋值 |
| 是否冻结 deep_match_v3 Schema | ✅ 冻结 |
| 是否冻结 score_alignment | ✅ 冻结 |
| 是否冻结 Prompt 约束 | ✅ 冻结 |
| 是否可以进入 H19-B | ✅ 可以 |

---

## 2. 当前 Deep 链路审计

| 项目 | 当前状态 | 问题 | H19-A 冻结目标 |
|------|---------|------|---------------|
| Deep 入口 | `Analyze()` → `runDeepMatchSEP()` | 降级链复杂（4层 fallback） | 简化：SEP Gate → 8-Agent → report |
| SEP 使用 | DeepMatchReport 读取 SEP，3-Agent 不读 | **数据源分裂**：SEP 与 TagEvidenceService 并行独立 | 统一：所有 Agent 读 SEP Profile |
| Fast 初判使用 | `FastResult` 字段未赋值 | DeepMatchContext.FastResult 预留但从未使用 | H19-B 必须传入 Fast Result |
| TagEvidenceService | 3-Agent 强依赖 `EnsureResumeTagProfile/EnsureJobTagProfile` | Deep 主路径仍依赖旧标签系统 | H19-B 必须移除 |
| LLM fallback | `analyzeWithLLM()` 单 LLM 管道 | generated_by=llm_agent 被 sanitizeDeepResponse 覆盖为 deep_match_agent，丢失可观测性 | H19-B 保留 fallback 但保留原始 generated_by |
| Mock fallback | `analyzeWithMock()` → mock_quality | 伪造 match_analysis 输出 | H19-B 保留但显式标记为 mock |
| 输出 Schema | mix of 3-agent strings + SEP structured items | ScoreBreakdown 未填充, DecisionLabel 为空, StructuredRisks 未填充 | deep_match_v3 完整 Schema |
| 分数字义 | 3-Agent 使用旧 matchLevel 标签 | 可能使用 "建议投递" 等旧标签 | 统一 90/80/70/60 |
| 空字段风险 | ScoreBreakdown/DecisionLabel/HardGate/StructuredRisks 未填充 | 前端读取时可能遗漏 | H19-B 必须全部填充 |

---

## 3. Deep API Baseline

> 当前运行旧二进制，Deep 路径使用 H18-D 之前的代码。以下基于代码路径审计 + frontend 实际渲染字段。

| 项目 | 当前实际状态 |
|------|-------------|
| Fast score | N/A (当前二进制未重启) |
| Deep score | N/A |
| generated_by | 代码：`deep_match_agent`（sanitizeDeepResponse 强制覆盖） |
| agent_status | 代码：`match_verified`（sanitizeDeepResponse 统一设置） |
| report_version | 代码：`deep_match_v2`（enrichResponseWithReport 固定值） |
| job_understanding | ✅ SEP Profile 填充，6 字段 |
| requirement_coverage | ✅ 规则引擎填充（关键词匹配，最多 8 条） |
| resume_evidence_mapping | ✅ SEP Resume Tags 填充 |
| key_risks | ⚠️ 从 3-Agent risk_points string[] 转换（仅 RiskName，无解释） |
| resume_rewrite_strategy | ⚠️ 从 gap_points string[] 填充（仅 skills_reorder） |
| interview_risk_questions | ✅ 规则引擎填充（最多 5 条） |
| agent_trace | ✅ 硬编码 6 阶段 |
| score_breakdown | ❌ **空** — DeepMatchReport.ScoreBreakdown 零值结构 |
| application_decision | ❌ **空** — buildDeepMatchReportFromSEP 未赋值 |
| decision_label | ❌ **空** |
| structured_risks | ❌ **未填充** |
| hard_gate | ❌ **未填充** |
| core_advantages/main_gaps/action_suggestions | ❌ **未填充** |

**前端实际渲染的 Deep 卡片（10 张）：**

| Card | 使用字段 | 后端来源 |
|------|---------|---------|
| 结论卡 | matchScore, summary | 3-Agent 输出 |
| JD 要求与差距 | gapPoints | 3-Agent GapPoints |
| 简历证据映射 | resumeEvidenceMapping | SEP |
| 核心匹配点 | fitPoints / matchHighlights | 3-Agent |
| 主要短板 | gapPoints | 3-Agent |
| 风险诊断 | riskPoints (string[]) | 3-Agent sanitized |
| 投递策略 | deliveryStrategy | 3-Agent |
| 面试表达要点 | interviewTalkingPoints | 可能为空 |
| 岗位要求覆盖 | requirementCoverage | SEP 规则引擎 |
| 面试追问预测 | interviewRiskQuestions | SEP 规则引擎 |

**未渲染但有结构的字段：** reportVersion, oneSentenceConclusion, jobUnderstanding, coreMatchPoints(结构化), keyRisks(结构化), resumeRewriteStrategy, actionPlan, agentTrace

---

## 4. 冻结 Deep 输入上下文（13 项）

| # | 输入项 | 类型 | 优先级 | 说明 |
|---|--------|------|--------|------|
| 1 | resume_id | int | must | 简历 ID |
| 2 | job_id | int | must | 岗位 ID |
| 3 | resume_raw_text | string | must | 简历原始文本 |
| 4 | job_raw_text | string | must | JD 全文 |
| 5 | resume_sep_profile | SemanticProfile | **首选** | 简历语义画像 |
| 6 | job_sep_profile | SemanticProfile | **首选** | 岗位语义画像 |
| 7 | fast_match_result | MatchAnalysisResult | **首选** | Fast 初判结果（含全部结构化字段） |
| 8 | fast_score_breakdown | ScoreBreakdownResult | high | Fast 五维评分 |
| 9 | fast_hard_gate | HardGateResult | high | Fast 质量门槛 |
| 10 | fast_requirement_coverage | []RequirementCoverageItem | high | Fast 要求覆盖 |
| 11 | fast_resume_evidence | []ResumeEvidenceItem | high | Fast 简历证据 |
| 12 | fast_structured_risks | []MatchRiskPoint | high | Fast 结构化风险 |
| 13 | user_intent | string | optional | 用户自定义意图 |

**输入优先级规则：**
```
SEP Profile > Fast Match Evidence > 原始简历/JD > Legacy Tag Profile
```

**Legacy Tag Profile 仅允许作为兼容兜底，不允许作为 Deep 主判断依据。**

---

## 5. 冻结 Deep Agent 分层流程（8 层）

| 层级 | Agent | 职责 | 输入 | 输出 |
|------|-------|------|------|------|
| **L0** | Input Gate | 检查 resume/job/SEP/Fast 齐全；若缺 SEP → profile_pending | 13 项输入 | gate_result |
| **L1** | JD Deconstruction | 解构岗位定位、必备要求、重要要求、加分项、隐藏期望 | SEP Job Profile + JD raw | job_understanding |
| **L2** | Resume Evidence Audit | 审计简历证据，标注 evidence_level，禁止无证据硬夸 | SEP Resume Profile + Resume raw | resume_evidence_mapping |
| **L3** | Requirement Alignment | 逐条对齐 JD 要求与简历证据 | L1 + L2 输出 + Fast coverage | requirement_coverage |
| **L4** | Risk Diagnosis | 输出 high/medium/low 风险（必须 why_it_matters/evidence_gap/fix_suggestion） | L3 输出 + Fast risks | key_risks |
| **L5** | Score Calibration | 5-6 维评分 + 统一分数字义 + 对比 Fast 分数 | L1-L4 输出 + Fast score_breakdown | score_breakdown + score_alignment |
| **L6** | Application Strategy | 投递建议 + 竞争定位 + 投递策略 | L5 输出 + Fast hard_gate | application_decision + application_strategy |
| **L7** | Resume Rewrite + Interview | 简历改写策略 + 面试追问预测 | L3-L4 输出 | resume_rewrite_strategy + interview_risk_questions |
| **L8** | Global Consistency | 检查分数/风险/证据/建议是否矛盾；空字段检测 | 全部输出 | quality_checks |

**Agent Trace 必须记录至少 8 阶段（可合并相邻层），不能伪装实时进度。**

---

## 6. 冻结 Deep 分数字义

### 6.1 等级映射（与 Fast 一致）

| 分数 | 等级 | 投递含义 |
|------|------|---------|
| 90-100 | 高度匹配 | 方向强一致 + 必备覆盖充分 + 证据充实 + 风险可控 |
| 80-89 | 较高匹配 | 方向一致/相邻 + 主要覆盖较好 + 少量短板 |
| 70-79 | 中等匹配 | 有相关性但核心或证据不足 |
| 60-69 | 弱匹配 | 方向或核心能力明显短板 |
| 0-59 | 不建议投递 | 明显不匹配 |

### 6.2 Score Breakdown（6 维，比 Fast 多 jd_competitiveness）

| 维度 | 权重 | 说明 |
|------|------|------|
| direction_fit | 20% | 方向匹配度 |
| must_have_coverage | 20% | 必备要求覆盖 |
| evidence_strength | 20% | 证据强度 |
| experience_relevance | 15% | 经历相关性 |
| risk_control | 15% | 风险可控度 |
| jd_competitiveness | 10% | JD 竞争激烈度估计 |

### 6.3 Score Cap（与 Fast 8 条完全一致）

| # | 条件 | 最高分 |
|---|------|--------|
| 1 | 方向明显不一致 | 65 |
| 2 | 缺少核心必备技能 | 75 |
| 3 | 仅技能列表无项目/实习 | 72 |
| 4 | 无相关项目证据 | 70 |
| 5 | 无实习但岗位强依赖实习 | 78 |
| 6 | 证据充分但表达弱 | 85 |
| 7 | gate blocked | 60 |
| 8 | 强一致+强实习+低风险 | 90+ allowed |

### 6.4 Fast/Deep 分数对齐

```
Deep 可以比 Fast 更低（发现隐藏要求/竞争风险/表达风险）
但不能更低却无解释
```

**score_alignment 结构：**

```json
{
  "fast_score": 76,
  "deep_score": 70,
  "score_delta": -6,
  "is_consistent": true,
  "delta_reason": "Deep 补充考虑了岗位隐藏期望和竞争风险，因此比 Fast 低 6 分。"
}
```

**差距 > 15 必须 warning：** `is_consistent = false`，`delta_reason` 具体解释。

---

## 7. 冻结 Deep 输出 Schema（deep_match_v3）

### 7.1 顶层字段

| 字段 | 类型 | 必填 | 默认值 | H19-B 变更 |
|------|------|------|--------|-----------|
| generation_tier | string | ✅ | `"quality"` | 无 |
| generated_by | string | ✅ | `"deep_match_agent"` | 保留 LLM fallback 的原始标签 |
| agent_status | string | ✅ | `"match_verified"` | 区分 verified/repaired/fallback |
| report_version | string | ✅ | `"deep_match_v3"` | **v2→v3** |
| match_score | int | ✅ | 0 | 统一 0-100 |
| match_level | string | ✅ | `""` | 90/80/70/60 |
| application_decision | string | ✅ | `""` | **新增** |
| decision_label | string | ✅ | `""` | **新增** |
| one_sentence_conclusion | string | ✅ | `""` | 具体结论 |
| score_breakdown | object | ✅ | `{}` | **新增填充** 6 维 |
| score_alignment | object | ✅ | `{}` | **新增** |
| job_understanding | object | ✅ | `{}` | 修复字段名 |
| requirement_coverage | []object | ✅ | `[]` | 6 状态 |
| resume_evidence_mapping | []object | ✅ | `[]` | 修复 source_type |
| core_match_points | []object | ✅ | `[]` | 结构化 |
| key_risks | []object | ✅ | `[]` | 6 字段 |
| resume_rewrite_strategy | object | ✅ | `{}` | 5 数组 |
| interview_risk_questions | []object | ✅ | `[]` | 4 字段 |
| application_strategy | object | ✅ | `{}` | 3 字段 |
| action_plan | []object | ✅ | `[]` | 结构化 |
| agent_trace | []object | ✅ | `[]` | 8 阶段 |
| quality_checks | object | ✅ | `{}` | **新增** |

### 7.2 key_risks 子结构

```json
{
  "risk_name": "string",
  "risk_level": "high|medium|low",
  "why_it_matters": "string",
  "evidence_gap": "string",
  "fix_suggestion": "string",
  "interview_impact": "string"
}
```

### 7.3 resume_evidence_mapping 子结构

```json
{
  "evidence_name": "string",
  "evidence_level": 0-5,
  "source_type": "internship_project|project|skill_list",
  "source_title": "string",
  "evidence_text": "string",
  "supports_requirement": "string",
  "confidence": "high|medium|low"
}
```

### 7.4 score_alignment 子结构

```json
{
  "fast_score": 76,
  "deep_score": 70,
  "score_delta": -6,
  "is_consistent": true,
  "delta_reason": "string"
}
```

### 7.5 quality_checks 子结构

```json
{
  "has_jd_understanding": true,
  "has_resume_evidence": true,
  "has_requirement_coverage": true,
  "has_structured_risks": true,
  "has_rewrite_strategy": true,
  "has_interview_questions": true,
  "has_score_alignment": true,
  "no_null_arrays": true,
  "no_raw_tag_code": true
}
```

---

## 8. 冻结 Prompt 约束

### 必须遵守：

```
1. 你是 evidence-grounded 岗位匹配顾问。
2. 目标不是夸学生，而是判断岗位是否值得投递，以及如何提高初筛命中率。
3. 必须基于简历证据和 JD 要求分析。
4. 如果简历没有证据，不允许编造。
5. 如果能力只是技能列表出现，不能当成强证据。
6. 必须区分 must_have / should_have / nice_to_have / hidden_expectation。
7. 必须输出结构化 JSON（非 Markdown 长文）。
8. 所有数组必须返回 []，不能返回 null。
9. 所有用户可见字段必须是中文。
10. 禁止输出 raw tag code（如 technical_skill, role_direction）。
11. 禁止泛泛建议（如"继续提升技术能力"）。
12. 每个风险必须说明 why_it_matters、evidence_gap、fix_suggestion。
13. 如果 Deep 分数与 Fast 分数差距超过 15，必须解释原因。
```

### 禁止 Prompt 输出：

```
1. "你很适合该岗位" 但无证据
2. "建议提升项目经验" 但不说怎么改
3. "风险：经验不足" 但不解释影响
4. fit_points=null
5. evidence_chain=null
6. raw tag code
7. Markdown 长文而非 JSON
```

---

## 9. 冻结 Deep 前端展示模块（10 个）

| # | 模块 | 主要字段 | 说明 |
|---|------|---------|------|
| 1 | 顶部 AI 投递决策卡 | matchScore, matchLevel, decisionLabel, one_sentence_conclusion, generated_by | 分数 + 等级 + 建议 + 一句话结论 |
| 2 | Fast/Deep 分数对齐卡 | score_alignment | 对比 Fast 分数、解释差距 |
| 3 | JD 深度解构卡 | job_understanding | 定位 + 必备/重要/加分/隐藏要求 |
| 4 | 要求覆盖分析卡 | requirement_coverage | 逐要求覆盖状态 + 分数 |
| 5 | 简历证据映射卡 | resume_evidence_mapping | evidence_level + source + text |
| 6 | 核心匹配点卡 | core_match_points | 匹配点 + match_strength + 证据 |
| 7 | 风险诊断卡 | key_risks | risk_level badge + why_it_matters + fix_suggestion |
| 8 | 简历改写策略卡 | resume_rewrite_strategy | 技能重排 + 项目改写 + 关键词 + 弱化 |
| 9 | 面试追问预测卡 | interview_risk_questions | question + why_asked + answer_direction |
| 10 | Agent 分析履历卡 | agent_trace | 8 阶段 + status + summary |

**前端展示原则：** 不展示 raw JSON / null / raw tag code；分数差异必须解释；风险必须可读；改写策略必须可执行。

---

## 10. 冻结 H19-B 实现边界

| 允许修改 | 禁止修改 |
|---------|---------|
| `server/internal/service/match_service.go` | 前端页面 |
| `server/internal/service/match_deep_agent.go` | 数据库结构 |
| `server/internal/service/match_deep_schema.go` | Fast Match Engine |
| `server/internal/agent` 与 match deep 相关文件 | 岗位推荐 / 简历优化 / 简历对比 |
| `server/internal/model/match.go` | Deep Match Prompt（当前已冻结） |
| 必要的 prompt/schema 文件 | `.env` / `docker-compose.yml` |

**H19-B 最低目标：**
1. Deep 主路径读取 SEP + Fast Match Result（移除 TagEvidenceService 依赖）
2. Deep 输出 deep_match_v3 Schema（全部 25 个顶层字段）
3. Deep 使用统一分数字义（90/80/70/60）
4. Deep 输出 score_alignment（Fast/Deep 对比）
5. Deep 输出 quality_checks
6. Legacy fallback 显式标记为 legacy，不伪装为 verified

---

## 11. 构建检查

| 项目 | 结果 |
|------|------|
| npm build | ✅ 通过 (2.38s) |
| go build ./... | ✅ 通过 |
| go vet ./... | ✅ 通过 |
| go test ./... | ✅ 通过 |

---

## 12. docs

- **新增：** `docs/09-开发日志/2026-06-13-Deep-Match-Schema-And-Prompt-Contract-Freeze-H19A.md`
- **更新：** `docs/变更记录.md`
- **更新：** `docs/下一步任务.md`

---

## 13. 给用户的结论

1. 当前 Deep Match 存在**数据源分裂**：3-Agent pipeline 使用 TagEvidenceService（旧标签），DeepMatchReport 使用 SEP 语义画像（新标签），两条数据源独立运行。
2. Deep 当前**不读取 Fast 初判结果** — FastResult 字段已声明但从未赋值，导致 Deep 无法对比或解释 Fast/Deep 分数差异。
3. 当前输出**缺失多个关键字段**：ScoreBreakdown、DecisionLabel、StructuredRisks、HardGate 全部为空。
4. 4 层降级链（3-agent → single LLM → fast fallback → mock）增加了复杂性但未提升质量。
5. H19-A 冻结了 **8 层 Agent 流程**（L0-L8）、**13 项输入上下文**（SEP Profile + Fast Result 为首选）、**deep_match_v3 完整 Schema**（25 个顶层字段含 score_alignment + quality_checks）。
6. H19-A 冻结了 **Prompt 约束**（13 条必须 + 7 条禁止）和 **10 个前端展示模块**。
7. H19-A 冻结了**统一分数字义**：6 维 score_breakdown + 8 条 score cap + Fast/Deep 分数对齐契约。
8. H19-B 必须实现：移除 TagEvidenceService 依赖、读取 Fast Result 作为输入、填充全部 deep_match_v3 字段、输出 score_alignment、输出 quality_checks。
9. H19-B 不能直接复用当前 3-agent pipeline — 需要重建 Agent 流程为 SEP-first 的 8 层顾问架构。
10. 不进入 H19-B。
