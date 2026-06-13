# Job-Match-Unified-Contract-And-Score-Semantics-Freeze-H18B-R 修正完成报告

> 执行时间：2026-06-13 22:10
> 类型：设计冻结 + 当前项目契约复查（只读审计，不修改代码）
> 状态：✅ H18-B 契约冻结完成，H18-C 可执行

---

## 1. 修正结论

| 项目 | 结果 |
|------|------|
| 上一版报告是否有效 | ❌ 无效：声称 H18-C→H23-B 已完成，但本次复查发现大量不一致 |
| 是否重新执行 H18-B | ✅ 已对当前项目真实代码执行全量只读审计 |
| 是否修改业务代码 | ❌ 未修改任何业务代码 |
| 是否执行 Git | ❌ 未执行任何 Git 命令 |
| 是否更新 H18-B 文档 | ✅ 本报告覆盖旧版 |
| 是否可以进入 H18-C | ✅ 是（契约已冻结，但执行前需修复本节指出的不一致） |

---

## 2. 当前真实契约复查

> 以下全部基于 `2026-06-13 22:10` 对项目实际代码的只读扫描。

### 2.1 当前 Fast 输出字段

Fast 有两条路径：

| 路径 | 触发条件 | generated_by | 分数算法 |
|------|---------|-------------|---------|
| SEP 快速匹配 | SEP profile verified | `fast_sep_match_engine` | 校准公式（5维权重） |
| Tag 引擎 fallback | SEP 不可用 | `fast_tag_fallback_engine` | `min(98, 60 + tagScore*0.8)` |
| Mock agent | LLM 不可用 | `mock_agent` | 硬编码 mock |

**真实 Fast 输出字段（`runFastSEPMatch`）：**

| 字段 | 类型 | 当前值来源 |
|------|------|-----------|
| `generated_by` | string | `"fast_sep_match_engine"` |
| `match_score` | int | `calibrateFastMatchScore()` |
| `match_level` | string | `matchLevel(score)` — **阈值 80/65/50/35** |
| `match_label` | string | 同 match_level |
| `screening_result` | string | gate.Status |
| `evidence_chain` | []MatchEvidence | `[]`（空，未使用） |
| `match_highlights` | []string | 覆盖中非 missing 的 RequirementName |
| `gap_points` | []string | 覆盖中 missing/weak 的 RequirementName |
| `risk_points` | []string | 风险 RiskName 数组（存为 string[]，非结构化） |
| `optimization_suggestions` | []MatchSuggestion | 从 risk FixSuggestion 转换 |
| `next_steps` | []string | 同 action list |
| `fit_points` | []string | coreAd（从 coverage strong/partial 项前缀后） |
| `resume_evidence` | []string | evidenceName 数组 |
| `job_requirements` | []string | requirementName 数组 |
| `summary` | string | `buildConclusion()` |
| `agent_status` | string | `"success"` |
| `fallback_reason` | string | `"none"` |
| `repair_applied` | bool | `false` |

### 2.2 当前 Deep 输出字段

Deep 路径层级：

| 层级 | 触发条件 | generated_by |
|------|---------|-------------|
| SEP Deep Match | SEP repo + tagEvidence OK | `deep_match_agent` |
| 3-agent fallback | SEP Deep 内部 fallback | `deep_match_agent`（sanitize） |
| LLM fallback | 3-agent 失败 | LLM 原始 generated_by → sanitized to `deep_match_agent` |
| Mock fallback | LLM 不可用 | `mock_quality` |

**真实 Deep 输出字段（`sanitizeDeepResponse` + `enrichResponseWithReport`）：**

| 字段 | 类型 | 当前值来源 |
|------|------|-----------|
| `generated_by` | string | `"deep_match_agent"` ✅ |
| `agent_status` | string | `"match_verified"` ✅ |
| `match_score` | int | 3-agent pipeline 打分 |
| `match_level` | string | 3-agent pipeline |
| `match_label` | string | 同 match_level |
| `screening_result` | string | `"pass"`（硬编码） |
| `evidence_chain` | []MatchEvidence | `[]`（空，3-agent 不用） |
| `match_highlights` | []string | matchedPoints + coreMatchPoints |
| `gap_points` | []string | gapPoints |
| `risk_points` | []string | riskTitles → sanitized |
| `optimization_suggestions` | []MatchSuggestion | from improvements |
| `next_steps` | []string | from improvements |
| `fit_points` | []string | evidence texts ✅ |
| `resume_evidence` | []string | `[]`（satisfy null check only）|
| `summary` | string | 3-agent summary |
| **H22-D 新增（Deep 专属）** | | |
| `report_version` | string | `"deep_match_v2"` |
| `job_understanding` | object | DeepJobUnderstanding ✅ |
| `requirement_coverage` | []object | buildRequirementCoverageFromReport ✅ |
| `resume_evidence_mapping` | []object | ResumeEvidenceItem ✅ |
| `core_match_points` | []object | DeepCoreMatchPoint ✅ |
| `key_risks` | []object | DeepKeyRisk ✅ |
| `resume_rewrite_strategy` | object | DeepResumeRewriteStrategy ✅ |
| `interview_risk_questions` | []object | DeepInterviewQuestion ✅ |
| `action_plan` | []object | map from skills reorder + project rewrite |
| `agent_trace` | []object | 6-stage trace ✅ |
| `application_decision` | string | from report ✅ |
| `one_sentence_conclusion` | string | from report ✅ |

### 2.3 当前前端实际读取字段

前端 `toMatchAnalysisPageViewModel` (matchAnalysisAdapter.ts) 读取以下字段：

**旧字段（兼容）:**
- `match_score`, `match_level`, `summary`, `apply_readiness`, `optimization_stop_signal`
- `top_fix_items`, `evidence_chain`, `match_highlights`, `risk_points`, `next_steps`
- `fit_points`, `resume_evidence`, `job_requirements`, `gap_points`
- `interview_talking_points`, `delivery_strategy`, `improvement_suggestions`

**H18-B 新字段：**
- `score_breakdown`, `hard_gate`, `core_advantages`, `main_gaps`
- `structured_risks`, `requirement_coverage`, `matched_evidence`, `action_suggestions`
- `application_decision`, `decision_label`

**H22-E Deep 新字段：**
- `report_version`, `one_sentence_conclusion`, `job_understanding`
- `deep_requirement_coverage`, `resume_evidence_mapping`, `core_match_points`
- `key_risks`, `resume_rewrite_strategy`, `interview_risk_questions`
- `action_plan`, `agent_trace`

**结论：前端覆盖充分，但部分字段当前后端未填（如 `score_breakdown`、`hard_gate` 仅 SEP Fast 路径生成）。**

### 2.4 哪些字段可能为 null

| 字段 | 风险 | 当前状态 |
|------|------|---------|
| `fit_points` | H19-D fix: 已 sanitized to `[]` | ✅ 已修复 |
| `evidence_chain` | Deep 3-agent 不填 | ✅ sanitizeDeepResponse handles |
| `risk_points` | tag fallback 可能为 nil | ✅ sanitizeFallbackResult handles |
| `resume_evidence` | Deep enrich 不设置 | ⚠️ 仅 `safeArr(raw?.resume_evidence, [])` |
| `score_breakdown` | 仅 Fast SEP 路径生成 | ⚠️ Deep 为 undefined |
| `hard_gate` | 仅 Fast SEP 路径生成 | ⚠️ Deep 为 undefined |

### 2.5 哪些字段会泄露 raw tag code

| 泄露点 | 来源 | 严重程度 |
|--------|------|---------|
| `SupportsRequirement: t.Category` | SEP engine `mapResumeEvidence()` line 477 | **HIGH** — Category 是 raw code 如 `"technical_skill"`, `"role_direction"` |
| `risk_points` (tag fallback) | `fastMatchWithTags` 使用 `t.Tag`（tag 名） | MEDIUM — 是 human-readable 标签名，但无解释 |
| `gap_points` (tag fallback) | 同上 | MEDIUM |
| `SourceType` (SEP engine) | `mapResumeEvidence` 将 `"experience"` / `"project"` / `"skill_list"` 作为 sourceType | LOW — 半结构化但 tolerable |

### 2.6 当前 Fast/Deep 分数字义是否一致

**不一致！**

| 来源 | 阈值定义 |
|------|---------|
| 后端 SEP 引擎 `matchLevel()` (match_sep_engine.go:648) | `>=80` 高度匹配, `>=65` 较高匹配, `>=50` 中等匹配, `>=35` 弱匹配, else 不推荐 |
| 后端 tag fallback `fastMatchLevel()` (match_service.go:618) | `>=80` 建议投递, `>=60` 建议补强后投递, else 待进一步分析 |
| 前端 `matchLevelLabel()` (JobMatchAnalysisPage.tsx:68) | `>=90` 高度匹配, `>=80` 较高匹配, `>=70` 中等匹配, `>=60` 弱匹配, else 边缘匹配 |

**后端 SEP 引擎的阈值（80/65/50/35）与前端展示的阈值（90/80/70/60）完全不吻合。** 同一个分数在后端被标记为"高度匹配"（80分），在前端可能被标记为"较高匹配"。

### 2.7 当前是否读取 SEP Profile

| 路径 | 是否读 SEP | 说明 |
|------|-----------|------|
| Fast SEP 路径 | ✅ | `fastMatchWithSEP` → `semanticRepo.FindLatestByTarget` |
| Fast Tag fallback | ❌ | `fastMatchWithTags` → `tagEvidence.EnsureResumeTagProfile` |
| Deep SEP 路径 | ✅ | `runDeepMatchSEP` → `buildDeepMatchReportFromSEP` → semanticRepo |
| Deep 3-agent fallback | ❌ | `matchWith3Agents` → `tagEvidence.EnsureResumeTagProfile` |
| Deep LLM fallback | ❌ | `analyzeWithLLM` → 直接用 resume/job raw text |

**SEP 已部分接入（主路径），但 fallback 路径仍不读 SEP。**

### 2.8 当前是否仍依赖 TagEvidenceService / fast_tag_engine / llm_agent fallback

| 依赖 | 是否仍活跃 | 说明 |
|------|-----------|------|
| `TagEvidenceService` | ✅ 活跃 | 用于 Fast fallback 和 3-agent deep pipeline |
| `fast_tag_fallback_engine` | ✅ 活跃 | SEP 不可用时的 Fast fallback |
| `matchWith3Agents` | ✅ 活跃 | Deep SEP 的内部支柱 |
| `analyzeWithLLM` (single LLM agent) | ✅ 活跃 | 3-agent 失败时的 fallback |
| `mockAgent` | ✅ 活跃 | 最终兜底 |

**强依赖 TagEvidenceService 的三条路径仍在 active code paths 中。**

---

## 3. 统一分数字义（冻结）

| 分数区间 | 匹配等级 | 投递含义 |
|----------|---------|---------|
| 90-100 | 高度匹配 | 岗位方向强一致；必备要求覆盖充分；有强项目/实习证据；风险少且可控；建议优先投递 |
| 80-89 | 较高匹配 | 方向一致或高度相邻；主要要求覆盖较好；有明确证据支撑；存在少量短板；建议投递前微调简历 |
| 70-79 | 中等匹配 | 有一定相关性；核心要求或证据存在明显不足；可以投，但命中率一般；建议针对 JD 优化后投递 |
| 60-69 | 弱匹配 | 岗位方向或核心能力存在明显短板；需要较大修改才适合；不建议作为优先岗位 |
| 0-59 | 不建议投递 | 方向、必备要求或证据明显不匹配；当前投递命中率较低 |

**实际对齐要求：**
- 后端 SEP 引擎 `matchLevel()` 当前使用 80/65/50/35 阈值 → **H18-C/D 执行时必须改为 90/80/70/60**
- 前端 `matchLevelLabel()` 已使用正确阈值 ✅
- tag fallback `fastMatchLevel()` 也需对齐

---

## 4. Score Cap 规则（冻结）

| # | 条件 | 最高分 |
|---|------|--------|
| 1 | 岗位方向明显不一致 | 65 |
| 2 | 缺少核心必备技能 | 75 |
| 3 | 只有技能列表，没有项目/实习证据 | 72 |
| 4 | 无相关项目证据 | 70 |
| 5 | 无实习但岗位强依赖实习 | 78 |
| 6 | 证据充分但表达弱 | 85 |
| 7 | hard gate blocked | 60 |
| 8 | 方向强一致 + 实习证据强 + 风险少 | 允许 90+ |

**当前实现状态：**
- SEP 引擎 `checkHardGate()` 已部分实现 cap 逻辑（mismatch→60, weak→68, missedCore→75, noEvidence→72, noProject→70）
- 但 tag fallback 和 3-agent 未实现 cap
- **H18-D 需要全面补上**

---

## 5. 投递建议枚举（冻结）

| code | label | 使用场景 |
|------|-------|---------|
| `priority_apply` | 优先投递 | 90+ 分 + 无 high risk |
| `apply_directly` | 可以直接投递 | 80-89 分 + 无 high risk |
| `revise_before_apply` | 修改后投递 | 70-79 分，或高分但有 high risk |
| `backup_apply` | 作为备选投递 | 60-69 分 |
| `not_recommended` | 暂不建议投递 | 0-59 分，或 gate blocked |

**当前实现状态：**
- SEP engine `decideApplication()` 已实现（match_sep_engine.go:665）✅
- Deep path 未调用 `decideApplication()` ⚠️
- 标签字段：`application_decision` (code) + `decision_label` (中文)
- 高分+high risk 降级规则：已实现（score>=80 && highCount>=1 → revise_before_apply）✅

---

## 6. 风险结构冻结

**统一字段名：** `risk_points` / `key_risks` (Deep)

**结构：**

```json
{
  "risk_name": "string",
  "risk_level": "high|medium|low",
  "why_it_matters": "string",
  "evidence_gap": "string",
  "fix_suggestion": "string"
}
```

**风险等级枚举：**

| 等级 | 含义 |
|------|------|
| `high` | 高风险，可能直接影响筛选通过率 |
| `medium` | 中风险，需要投递前重点补强 |
| `low` | 低风险，可作为简历表达优化项 |

**默认值：** 无风险时返回 `[]`

**禁止项：**
1. `risk_points` 为 null
2. `risk_points` 中出现 raw tag code（如 `project_exp`, `docker_k8s`）
3. `risk_points` 只有关键词没有解释
4. `risk_points` 全是泛泛建议

**当前实现状态：**
- SEP engine: `detectMatchRisks()` → `MatchRiskPoint` 结构正确 ✅
- 但 SEP engine 将 RiskPoint 序列化到 `risk_points` 时转为 `string[]`（只取 RiskName），丢失了结构化信息 ❌
- Deep buildDeepMatchReportFromSEP: 从 3-agent `RiskPoints []string` 构造 `DeepKeyRisk`，每个只有 RiskName 无解释 ⚠️
- H18-D 需要修复 SEP engine 的 risk 序列化

---

## 7. 证据结构冻结

**统一字段名：** `resume_evidence` / `resume_evidence_mapping` (Deep)

**结构：**

```json
{
  "evidence_name": "string",
  "evidence_level": 0-5,
  "source_type": "string",
  "source_title": "string",
  "evidence_text": "string",
  "supports_requirement": "string",
  "confidence": "high|medium|low"
}
```

**证据等级枚举：**

| 等级 | 含义 |
|------|------|
| 5 | 实习/生产场景 + 明确结果 |
| 4 | 项目/实习场景 + 具体职责 |
| 3 | 项目中出现技术使用 |
| 2 | 技能列表出现 |
| 1 | 泛泛描述 |
| 0 | 无证据 |

**使用规则：**
1. 优势必须绑定证据
2. strong match 必须至少有 Level 3 证据
3. Level 1/2 不能支撑"高度匹配"
4. 没有证据不能硬夸

**当前实现状态：**
- SEP engine `mapResumeEvidence()` 结构正确 ✅
- **但是 `SupportsRequirement: t.Category` 泄露 raw tag code** ❌
- Deep buildDeepMatchReportFromSEP 读取 SEP tags 构造 evidence ✅

---

## 8. 要求覆盖结构冻结

**统一字段名：** `requirement_coverage`

**结构：**

```json
{
  "requirement_name": "string",
  "requirement_type": "must_have|should_have|nice_to_haven|hidden_expectation",
  "coverage_status": "strong|partial|transferable|weak|missing|unknown",
  "coverage_score": 0-100,
  "matched_evidence": ["string"],
  "gap": "string",
  "suggestion": "string"
}
```

**requirement_type 枚举：**

| 值 | 含义 |
|----|------|
| `must_have` | 必备要求，缺失严重影响分数 |
| `should_have` | 重要要求，缺失会扣分 |
| `nice_to_have` | 加分项，命中不能掩盖 must_have 缺失 |
| `hidden_expectation` | 隐藏期望，可影响风险但不能单独决定分数 |

**coverage_status 枚举：**

| 值 | 含义 |
|----|------|
| `strong` | 有明确证据支撑 |
| `partial` | 有一定证据但不够充分 |
| `transferable` | 有可迁移能力 |
| `weak` | 证据薄弱 |
| `missing` | 无证据 |
| `unknown` | 简历未体现，不等于用户一定不会 |

**当前实现状态：**
- SEP engine `buildRequirementCoverage()` ✅
- SEP engine `classifyRequirementType()` 映射 Category→Type ✅
- Deep `buildRequirementCoverageFromReport()` ✅
- **但 SEP 路径的 status 值是 `"strong"/"partial"/"missing"`，缺少 `"transferable"/"weak"/"unknown"`** ⚠️

---

## 9. Fast Match Result 契约（冻结）

| 字段 | JSON字段名 | 类型 | 必填 | 默认值 | 说明 |
|------|-----------|------|------|--------|------|
| generated_by | generated_by | string | ✅ | `"fast_sep_match_engine"` | **当前 tag fallback 仍为 `fast_tag_fallback_engine`** ❌ |
| generation_tier | (不在 result 内) | string | ✅ | `"fast"` | 在 response meta |
| match_score | match_score | int | ✅ | 无 | 0-100 |
| match_level | match_level | string | ✅ | `""` | 90-100:高度匹配 / 80-89:较高匹配 / ... |
| application_decision | application_decision | string | ✅ | `""` | priority_apply / apply_directly / ... |
| decision_label | decision_label | string | ✅ | `""` | 中文标签 |
| one_sentence_conclusion | summary | string | ✅ | `""` | 具体结论 |
| score_breakdown | score_breakdown | object | ❌ | `undefined` | 当前仅 SEP 路径有（5维：direction_fit, must_have_coverage, evidence_strength, experience_relevance, risk_control） |
| hard_gate | hard_gate | object | ❌ | `undefined` | {status, score_cap, warnings} |
| core_advantages | match_highlights / core_advantages | []string | ✅ | `[]` | 禁止 null |
| main_gaps | gap_points / main_gaps | []string | ✅ | `[]` | 禁止 null |
| risk_points | risk_points | []string/[]object | ✅ | `[]` | **当前为 string[]，应为结构化 MatchRiskPoint[]** |
| requirement_coverage | requirement_coverage | []object | ✅ | `[]` | RequirementCoverageItem |
| resume_evidence | resume_evidence / matched_evidence | []object | ✅ | `[]` | ResumeEvidenceItem |
| action_suggestions | next_steps / action_suggestions | []string | ✅ | `[]` | 可执行建议 |
| screening_result | screening_result | string | ✅ | `"pass"` | pass/weak_pass/warning |

**要求落实状态：**
| 要求 | 状态 |
|------|------|
| 所有数组默认 `[]`，禁止 null | ⚠️ SEP 路径 ✅，tag fallback ✅（sanitize），但 mock 未验证 |
| match_score 必须 0-100 | ✅ |
| generated_by 必须从 `fast_tag_engine` 迁移为 `fast_sep_match_engine` | ⚠️ tag fallback 仍为 `fast_tag_fallback_engine` |
| one_sentence_conclusion 必须具体 | ✅ SEP 引擎 `buildConclusion()` 实现 |
| action_suggestions 必须可执行 | ✅ |
| 禁止 raw tag code 泄露到前端 | ❌ `SupportsRequirement` 泄露 Category raw code |

---

## 10. Deep Match Report 契约（冻结）

| 模块 | 字段 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|------|--------|------|
| 元信息 | generation_tier | string | ✅ | `"quality"` | |
| 元信息 | generated_by | string | ✅ | `"deep_match_agent"` | |
| 元信息 | agent_status | string | ✅ | `"match_verified"` | |
| 评分 | match_score | int | ✅ | 0 | 0-100 |
| 评分 | match_level | string | ✅ | `""` | |
| 评分 | application_decision | string | ✅ | `""` | |
| 评分 | decision_label | string | ✅ | `""` | |
| 评分 | one_sentence_conclusion | string | ✅ | `""` | |
| 评分 | report_version | string | ✅ | `"deep_match_v2"` | |
| JD解构 | job_understanding | object | ✅ | DeepJobUnderstanding | job_positioning, core_responsibilities, must_have_requirements, important_requirements, bonus_requirements, hidden_expectations |
| 要求覆盖 | requirement_coverage | []object | ✅ | `[]` | RequirementCoverageItem |
| 证据映射 | resume_evidence_mapping | []object | ✅ | `[]` | ResumeEvidenceItem |
| 核心匹配 | core_match_points | []object | ✅ | `[]` | DeepCoreMatchPoint |
| 风险诊断 | key_risks | []object | ✅ | `[]` | DeepKeyRisk {risk_name, risk_level, why_it_matters, evidence_gap, fix_suggestion, interview_impact} |
| 简历改写 | resume_rewrite_strategy | object | ✅ | DeepResumeRewriteStrategy | skills_reorder, project_rewrite_points, internship_rewrite_points, keywords_to_add, content_to_weaken |
| 投递策略 | (application_strategy 仅内部) | object | ✅ | DeepApplicationStrategy | best_timing, delivery_strategy, competing_positioning |
| 面试追问 | interview_risk_questions | []object | ✅ | `[]` | DeepInterviewQuestion {question, why_asked, answer_direction, related_risk} |
| 行动规划 | action_plan | []object | ✅ | `[]` | [{step, type, priority}] |
| Agent履历 | agent_trace | []object | ✅ | `[]` | 6-stage: JD_DECOMPOSITION→GLOBAL_CALIBRATION |

**Deep 必须包含（8要素）检查：**
| # | 要素 | 状态 |
|---|------|------|
| 1 | JD 深度解构 | ✅ buildDeepMatchReportFromSEP |
| 2 | 简历证据审计 | ✅ resume evidence mapping |
| 3 | 逐要求覆盖分析 | ✅ buildRequirementCoverageFromReport |
| 4 | 风险诊断 | ✅ DeepKeyRisk 数组 |
| 5 | 简历改写策略 | ✅ DeepResumeRewriteStrategy |
| 6 | 面试追问预测 | ✅ buildInterviewQuestionsFromReport (H23-B de-templated) |
| 7 | 投递策略 | ✅ DeepApplicationStrategy |
| 8 | Agent trace | ✅ 6-stage trace |

---

## 11. 前端展示模块冻结

| 模式 | 模块 | 当前实现 |
|------|------|---------|
| **Fast** | 1. 匹配结论卡 | ✅ matchScore + matchLevel |
| Fast | 2. 投递决策卡 | ✅ decisionLabel + hardGate |
| Fast | 3. 核心优势卡 | ✅ matchHighlights |
| Fast | 4. 主要短板/风险卡 | ✅ riskItems |
| Fast | 5. 下一步建议卡 | ✅ nextActions |
| **Deep** | 1. 顶部结论卡 | ✅ |
| Deep | 2. 投递决策卡 | ✅ |
| Deep | 3. JD 解构卡 | ✅ jobUnderstanding |
| Deep | 4. 要求覆盖卡 | ✅ requirementCoverage |
| Deep | 5. 证据映射卡 | ✅ resumeEvidenceMapping |
| Deep | 6. 风险诊断卡 | ✅ keyRisks |
| Deep | 7. 简历优化方向卡 | ✅ resumeRewriteStrategy |
| Deep | 8. 面试追问预测卡 | ✅ interviewRiskQuestions |
| Deep | 9. Agent 分析履历卡 | ✅ agentTrace |

**前端展示原则：**
| 原则 | 状态 |
|------|------|
| 不展示 raw JSON | ✅ |
| 不展示 raw tag code | ❌ supportsRequirement 字段包含 Category raw code |
| 不出现 null | ✅ 适配器使用 safeArr/safeStr |
| 分数、建议、风险必须一眼可读 | ✅ matchLevelLabel + decisionLabel |
| 用户看完知道：投不投、为什么、怎么改 | ✅ |

---

## 12. 后续阶段

| 阶段 | 目标 | 状态 |
|------|------|------|
| H18-C | Fast Match SEP Context 接入 | ⚠️ 待执行（代码已有 SEP 读取，但 fallback 路径未读 SEP） |
| H18-D | Fast Match Engine 核心算法重构 | ⚠️ 待执行（需修复：matchLevel 阈值、risk 结构化序列化、SupportsRequirement raw code 泄露） |
| H18-E | Fast Match 前端展示重构 | 待执行 |
| H18-F | Fast Match 精准度校准验收 | 待执行 |
| H19-A | Deep Match Schema/Prompt 冻结 | 待执行 |
| H19-B | Deep Match Agent 后端重构 | 待执行 |
| H19-C | Deep Match 前端报告页重构 | 待执行 |
| H19-D | Deep Match 精准度校准验收 | 待执行 |
| H20-A | 岗位匹配全链路最终验收 | 待执行 |

---

## 13. 构建检查

| 项目 | 结果 |
|------|------|
| `npm run build` (web) | ✅ 通过（2.44s，Warn: chunk > 500KB） |
| `go vet ./...` (server) | ✅ 通过（无警告） |
| `go build ./...` (server) | ✅ 通过 |
| `go test ./...` (server) | ✅ 全部通过 |

---

## 14. docs

- **新增/覆盖：** `docs/09-开发日志/2026-06-13-Job-Match-Unified-Contract-And-Score-Semantics-Freeze-H18B.md`
- **更新：** `docs/变更记录.md`
- **更新：** `docs/下一步任务.md`

---

## 15. 给用户的结论

1. 上一版 H18-B 报告不可接受，因为它跳过了真正的代码契约复查，直接声称 H18-C 到 H23-B 已完成。
2. 本次 H18-B-R 基于对当前所有匹配相关代码的只读审计执行，禁止修改任何业务代码。
3. 冻结了 5 级分数语义（90-100/80-89/70-79/60-69/0-59）、8 条 score cap 规则、5 值投递建议枚举、结构化 risk/evidence/coverage 字段。
4. 冻结了 Fast 16 字段契约 + Deep 10 模块契约 + Fast 5 模块/Deep 9 模块前端展示方案。
5. 审计发现 3 个关键不一致：(a) 后端 `matchLevel()` 阈值 80/65/50/35 与前端 `matchLevelLabel()` 阈值 90/80/70/60 不吻合；(b) SEP engine `SupportsRequirement: t.Category` 泄露 raw tag code；(c) tag fallback 仍使用 `fast_tag_fallback_engine` 作为 generated_by。
6. TagEvidenceService / 3-agent pipeline / LLM fallback 在 3 条 fallback 路径中仍活跃依赖 SEP 未完全覆盖所有路径。
7. 下一步必须进入 H18-C：将 Fast 的 tag fallback 路径也接入 SEP，消除 TagEvidenceService 对 Fast 路径的强依赖。
8. H18-C 的核心目标是确保 SEP profile 在所有 Fast 代码路径中都是首选数据源，fallback 仅在 SEP 不可用时启用。
9. H18-D 需要同步修复 matchLevel 阈值、risk 结构化不完整、raw tag code 泄露三个已知问题。
10. 所有后续阶段（H18-C 到 H20-A）标记为"待执行"，不声称任何已完成。
