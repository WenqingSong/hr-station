# Agent 输出 Quality Gate 设计

> 创建时间：2026-05-30
> 所属阶段：4-3.5.E
> 依据文档：战略差异化与杀手锏确认、投递边界与优化停止条件设计、JD证据链匹配与START简历优化规则、阶段4-3方案、岗位匹配Agent、岗位推荐Agent、Agent总览、Prompt模板、JSON输出规范、失败兜底与Mock回退策略

## 1. 背景

当前项目已完成三个杀手锏中的两个：

- 杀手锏一：投递边界与优化停止条件（4-3.5.C）
- 杀手锏二：JD 证据链匹配（4-3.5.D）

本项目定位为"面向学生的投递前决策与简历初筛优化系统"，Agent 输出直接影响学生的求职决策。但 LLM 生成内容不一定可靠，尤其在求职场景中可能出现：

1. 泛泛建议，缺乏针对性
2. 缺少 JD 证据链支撑
3. 编造经历、岗位、指标
4. 缺少投递边界判断
5. 缺少用户核实提醒
6. 无限建议用户继续优化
7. 与 apply_readiness 矛盾
8. 对用户做不恰当评价

因此需要 Quality Gate：在 Agent 输出返回给用户之前，对其内容质量进行门禁检查。

## 2. 核心问题：结构合法不等于内容可靠

JSON Schema 只能保证输出结构合法，不能保证内容有业务价值。

即使模型输出是合法 JSON，也可能存在：

- 泛泛建议（"建议突出项目经历"，但不对应哪个 JD 要求）
- 缺少 JD 证据链（只给结论不给依据）
- 编造经历（替用户编造项目、指标、实习）
- 编造岗位（推荐候选集外的岗位）
- 缺少投递边界（不给 apply_readiness）
- 缺少优化停止条件（不给 optimization_stop_signal）
- 缺少用户核实提醒（没有 verification_notice）
- 无限建议用户继续优化（与 apply_readiness=ready_to_apply 矛盾）
- 对用户做不恰当评价（"你不适合互联网"）

因此需要在 JSON Schema 校验之上，增加内容质量层面的检查。

## 3. 设计目标

1. 确保 Agent 输出不仅有合法结构，还有可靠内容
2. 确保每条建议都有 JD 依据和简历证据
3. 确保不编造用户没有的经历、指标、岗位
4. 确保给出投递边界和优化停止条件
5. 确保提醒用户核实 AI 建议
6. 控制成本和延迟，不做无限自我反省
7. 保持 Demo 稳定性
8. 不替代 JSON Repair / Schema Validate 的职责

## 4. 基本原则

### 4.1 Quality Gate 解决的是"内容质量"问题

Quality Gate 检查的是 Agent 输出的内容是否符合业务标准，而非结构是否合法。

| 层级 | 负责方 | 检查内容 |
|------|--------|---------|
| 结构合法 | JSON Schema | 字段存在、类型正确、枚举合法 |
| 格式修复 | JSON Repair | 修复 Markdown 包裹、语法错误、字段缺失 |
| 内容质量 | Quality Gate | 是否基于 JD、是否基于证据、是否编造、是否泛化、是否给出边界 |

### 4.2 Quality Gate 不是无限自我反省

Quality Gate 不是让模型无限反思、无限重写。

- 最多允许一次 revise
- 如果 revise 后仍不合格，系统应返回保守结果或进入 Mock fallback

原因：

1. 控制成本：每次 LLM 调用都有 Token 开销
2. 控制延迟：用户不能等太久
3. 避免模型自我循环：模型可能越改越差
4. 保证 Demo 稳定：宁可返回保守结果，也不能返回坏结果
5. 保持用户体验可控：快速返回比完美输出更重要

### 4.3 Quality Gate 是内容层门禁，不替代 JSON Repair

JSON Repair 负责修复格式和结构问题。Quality Gate 负责检查内容是否符合业务标准。

示例：

```
JSON 不合法 → JSON Repair
JSON 合法但建议泛泛、无证据、编造内容 → Quality Gate
Quality Gate 不通过 → revise once 或 conservative fallback
```

### 4.4 Quality Gate 必须围绕课题主旨

Quality Gate 的目标不是做职业规划，也不是评价用户人格。

目标仍然是：

1. 帮助学生有效匹配合适岗位
2. 提升对心仪岗位的简历初筛命中率

因此检查重点应围绕：

- JD 是否被使用
- 简历证据是否被使用
- 是否说明初筛风险
- 是否给出可执行优化动作
- 是否告诉用户是否可以投
- 是否提醒用户核实 AI 建议

## 5. Quality Gate 总体流程

```
Business Agent 生成结果
→ JSON parse
→ Schema validate
→ JSON Repair once（如需要）
→ Quality Gate 检查
  → 通过（pass）：返回结果
  → 可修复（revise_required）：revise once → 再检查
    → 通过：返回 revise 后结果
    → 仍不通过：conservative fallback 或 Mock fallback
  → 严重问题（fallback_required）：conservative fallback 或 Mock fallback
  → 不确定性高（manual_review_suggested）：返回保守建议 + 用户核实提示
```

注意：

- 本阶段只做设计，不改当前 Pipeline
- 是否把 Quality Gate 接入现有 Pipeline，放到 4-3.5.H 或后续工程评审决定
- Quality Gate 可以作为 Pipeline 后处理阶段，也可以作为独立 Agent
- 但本阶段不实现

## 6. Quality Gate 与 JSON Repair / Schema Validate 的区别

| 维度 | JSON Repair | Schema Validate | Quality Gate |
|------|-------------|-----------------|-------------|
| 负责什么 | 修复格式和结构 | 校验字段和类型 | 检查内容质量和业务边界 |
| 输入 | 不合法 JSON | 合法 JSON | 合法 JSON（已通过 Schema） |
| 输出 | 修复后 JSON | 通过/失败 | pass / revise_required / fallback_required / manual_review_suggested |
| 能否改变业务内容 | 不能 | 不能 | 指出问题，但不直接改内容 |
| 重试次数 | 最多 1 次 | 不重试 | 最多 revise 1 次 |
| 失败后 | Mock fallback | JSON Repair | conservative fallback 或 Mock fallback |

## 7. quality_gate_result 枚举设计

### 7.1 pass

**定义**：输出符合业务边界，可以返回给用户。

**触发条件**：

1. 有 JD 证据链（至少部分建议绑定了 JD 要求）
2. 有简历证据（建议基于用户已有内容）
3. 没有编造经历
4. 有投递边界（apply_readiness 存在且合理）
5. 有优化停止条件（optimization_stop_signal 存在且与 apply_readiness 一致）
6. 简历优化使用"示例参考"而非"建议改为"
7. 有核实提醒或自查问题
8. 输出符合 Schema

**后续动作**：直接返回。

**是否返回给用户**：是。

**示例**：

```
匹配分析输出了 5 条 JD 证据链，每条绑定了 JD 要求和简历证据，apply_readiness=quick_fix_before_apply，
optimization_stop_signal=fix_top_3_then_apply，有 verification_notice。
→ pass
```

---

### 7.2 revise_required

**定义**：输出总体可修复，但存在明显缺漏，需要重试一次。

**触发条件**：

1. 缺少 JD 证据链（建议没有绑定 JD 要求）
2. 建议过于泛泛（如"提升自己""多学习"等无操作性建议）
3. 缺少 self_check_question
4. 缺少 verification_notice
5. 缺少 optimization_stop_signal
6. 示例参考表达过强（使用"建议改为"而非"示例参考"）
7. 少量字段不完整但可修复

**后续动作**：将质量问题传回生成 Agent，要求 revise once。

**是否返回给用户**：revise 后若通过则返回，否则返回保守结果。

**示例**：

```
匹配分析输出了匹配分数和风险点，但没有 JD 证据链，建议都是泛泛的"提升自己"。
→ revise_required，补充 JD 证据链和具体建议。
```

---

### 7.3 fallback_required

**定义**：输出存在严重业务风险，不应直接返回。

**触发条件**：

1. 编造经历（AI 替用户编造了项目、实习、指标）
2. 编造岗位（推荐了候选集外的岗位）
3. 编造指标（编造了 QPS、DAU、提升百分比等具体数字）
4. 鼓励造假（建议用户写不真实的经历）
5. 给出录用承诺（"你一定能通过初筛"）
6. 输出与 JD 完全无关
7. 出现敏感或不当判断（人格评价、价值观判断、歧视性表述）
8. 与 apply_readiness / optimization_stop_signal 严重矛盾（如 ready_to_apply + needs_more_evidence）

**后续动作**：进入保守 fallback 或 Mock fallback。

**是否返回给用户**：不直接返回，返回保守结果或 Mock 结果。

**示例**：

```
简历优化建议中包含了"建议改为：负责高并发系统优化，将接口性能提升 80%"，
但用户简历中没有高并发和性能数据。
→ fallback_required，进入 conservative fallback。
```

---

### 7.4 manual_review_suggested

**定义**：输出涉及不确定性较高的复杂情况，建议用户自行核实或人工判断。

**触发条件**：

1. 用户简历信息过少，无法形成可靠证据链
2. 岗位要求高度专业，超出系统能力范围
3. 海外、地点、学历、实习资格等条件不明确
4. 系统无法稳定判断 apply_readiness

**后续动作**：返回保守建议，并明确提示用户核实。

**是否返回给用户**：是，但附带明显的"请自行核实"提示。

**示例**：

```
用户简历只有自我评价，没有任何项目、实习或技能信息。
→ manual_review_suggested，返回保守结果 + "当前信息不足，系统无法稳定判断"。
```

## 8. quality_gate_checks 检查维度

### 8.1 jd_grounding_check

**检查内容**：是否引用或使用了目标岗位 JD。是否存在脱离 JD 的泛泛建议。

**失败示例**：

```
建议你突出项目经历。
```

**通过示例**：

```
岗位要求包含 Go 后端、数据库、缓存，因此建议你在项目经历中明确 Go 接口开发、MySQL 表设计和 Redis 使用场景。
```

---

### 8.2 resume_evidence_check

**检查内容**：是否基于用户简历已有内容。是否编造用户经历。

**失败示例**：

```
假设用户有高并发项目经验并直接写入示例。
```

**通过示例**：

```
简历已提到 GoFrame 项目，因此可以围绕该项目补充职责边界和技术动作。
```

---

### 8.3 fabrication_check

**检查内容**：是否编造经历、岗位、职责、指标、证书、学历、论文、获奖。

**一旦失败**：应直接 fallback_required。

---

### 8.4 apply_boundary_check

**检查内容**：是否给出 apply_readiness。是否说明什么时候可以投、什么时候先改、什么时候不建议投、什么时候信息不足。

**失败示例**：

```
只给出 10 条优化建议，不说明是否可以投。
```

---

### 8.5 optimization_stop_check

**检查内容**：是否给出 optimization_stop_signal。是否避免无限优化。

**失败示例**：

```
不断建议用户继续学习、继续补项目、继续提升，但没有停止点。
```

---

### 8.6 example_reference_check

**检查内容**：简历优化是否使用"示例参考"而不是"建议改为"。是否提醒用户不要直接照抄未核实内容。

---

### 8.7 verification_check

**检查内容**：是否包含 verification_notice / self_check_question / learn_more_query。

**最低要求**：关键优化建议中应包含核实提醒或自查问题。

---

### 8.8 safety_tone_check

**检查内容**：是否存在羞辱用户、人格判断、价值观判断、敏感属性判断、录用承诺、过度保证。

**失败示例**：

```
你不适合互联网。
你不符合腾讯价值观。
你一定能通过初筛。
```

**通过表达**：

```
当前简历对该岗位关键要求的覆盖不足，暂不建议直接投递该岗位。
```

---

### 8.9 consistency_check

**检查内容**：apply_readiness 与 optimization_stop_signal 是否一致。

**禁止出现**：

```
apply_readiness=ready_to_apply + optimization_stop_signal=needs_more_evidence
apply_readiness=ready_to_apply + optimization_stop_signal=target_role_mismatch
apply_readiness=quick_fix_before_apply + optimization_stop_signal=stop_and_apply
apply_readiness=not_ready_for_this_role + optimization_stop_signal=stop_and_apply
```

---

### 8.10 schema_display_check

**检查内容**：输出是否适合前端展示。是否字段过长、结构混乱、难以渲染。

---

## 9. Quality Gate 输出结构

```json
{
  "quality_gate_result": "pass / revise_required / fallback_required / manual_review_suggested",
  "quality_score": 0.0,
  "failed_checks": [
    {
      "check_name": "jd_grounding_check",
      "severity": "high",
      "issue": "缺少 JD 证据链",
      "revision_instruction": "请补充每条建议对应的 JD 要求和简历证据。"
    }
  ],
  "revision_allowed": true,
  "revision_count": 0,
  "final_action": "return / revise_once / conservative_fallback / mock_fallback"
}
```

说明：

- `quality_score` 只用于内部判断，不一定展示给用户
- `failed_checks` 不一定展示给用户
- 前端可展示更自然的提示，而不是完整质检日志
- `quality_gate_result` 的四种值与第 7 节枚举对应
- `revision_count` 记录已执行的 revise 次数（0 或 1）
- `revision_allowed` 在 `revision_count < 1` 且 `quality_gate_result=revise_required` 时为 true

## 10. MatchAnalysisAgent 质检重点

### 10.1 重点检查项

1. 是否有 JD 证据链
2. 是否有 apply_readiness
3. 是否有 optimization_stop_signal
4. 是否说明筛选风险
5. 是否避免泛泛匹配
6. 是否没有编造简历证据

### 10.2 典型不合格输出

```
你很适合这个岗位，因为你的背景不错。
```

问题：没有 JD 证据链、没有 apply_readiness、没有具体证据、泛泛评价。

### 10.3 合格方向

```
岗位要求 Go 后端、数据库、缓存；简历中有 GoFrame、MySQL、Redis 项目证据；
但缺少结果表达，因此建议 quick_fix_before_apply。
```

## 11. RecommendationAgent 质检重点

### 11.1 重点检查项

1. 推荐岗位是否来自候选岗位集
2. 是否没有编造 job_id / title / department
3. 是否解释为什么先推荐这个岗位
4. 是否有推荐优先级
5. 是否有 recommendation_confidence
6. 是否指出是否需要先优化简历再投

### 11.2 一旦推荐了候选集外岗位

应 fallback_required。

## 12. ResumeOptimizationAgent 质检重点

### 12.1 重点检查项

1. 是否使用"示例参考"
2. 是否包含 verification_notice
3. 是否包含 self_check_question
4. 是否包含 learn_more_query
5. 是否基于 JD 和简历证据
6. 是否没有编造经历
7. 是否区分 top_fix_items 与 long_term_improvement
8. 是否有 optimization_stop_signal

### 12.2 典型不合格输出

```
建议改为：负责高并发系统优化，将接口性能提升 80%。
```

如果用户没有提供高并发和性能数据，这种输出必须判定不合格。

原因：

1. 使用"建议改为"而非"示例参考"
2. 编造了具体指标"80%"
3. 编造了用户没有的经历"高并发系统优化"

## 13. revise once 设计

### 13.1 核心约束

Quality Gate 不通过时，只允许 revise once。

### 13.2 revise 输入

revise 输入应包括：

1. 原始 Agent 输出
2. failed_checks（每项含 check_name、severity、issue、revision_instruction）
3. 原始 JD
4. 原始简历上下文
5. 原始 Schema 要求

### 13.3 revise 要求

1. 不得新增用户没有提供的事实
2. 必须补充 JD 证据链
3. 必须修正泛泛建议
4. 必须补充核实提醒
5. 必须保持 JSON Schema 合法
6. 必须保持 generated_by=llm_agent
7. 不得把 generated_by 改成 quality_gate_agent

### 13.4 revise 后的再检查

revise 后的输出需要再次通过 Quality Gate 检查：

- 如果通过：返回 revise 后结果，`revision_count=1`
- 如果仍不通过：进入 conservative fallback 或 Mock fallback

## 14. conservative_fallback / mock_fallback 策略

### 14.1 conservative_fallback

**适合场景**：LLM 输出有一定价值，但缺少足够证据，不适合给强结论。

**输出风格**：

```
当前信息不足，系统无法稳定给出强匹配判断。建议补充项目经历、个人职责和结果表达后再进行分析。
```

### 14.2 mock_fallback

**适合场景**：LLM 输出存在严重问题，或者无法产生结构化可用结果。

**使用现有 Mock Agent**：

- `match_mock.go`
- `recommendation_mock.go`
- 简历优化 Mock（待实现）

### 14.3 fallback 不是失败

必须说明：fallback 不是失败，而是 Demo 稳定性保障。

当 Quality Gate 判定输出不可用时，fallback 到保守结果或 Mock 结果是正确决策，不是系统故障。

## 15. 与现有 Pipeline 的关系

### 15.1 当前 Pipeline

```
Prompt → LLM → JSON parse → Schema validate → repair once → fallback_required
```

### 15.2 后续可扩展

```
Prompt → LLM → JSON parse → Schema validate → repair once → Quality Gate → revise once → fallback decision
```

### 15.3 本阶段只设计，不实现

本阶段只做 Quality Gate 规则设计，不改当前 Pipeline 代码。

### 15.4 需要评估的问题

1. 是否将 Quality Gate 作为 Pipeline 内置阶段
2. 是否作为业务 Agent 后处理阶段
3. 是否使用同一个 LLM 进行质检
4. 是否设计纯规则质检
5. 是否先做轻量规则版本，再做 LLM Critic 版本

### 15.5 推荐初步方案

先做轻量规则 Quality Gate，避免引入过多 LLM 调用成本。后续如时间允许，再接入 LLM Critic。

轻量规则版本可以包含：

1. 关键词匹配：检查是否包含 JD 关键词
2. 字段存在性检查：检查 apply_readiness、optimization_stop_signal、verification_notice 等是否存在
3. 一致性检查：检查 apply_readiness 与 optimization_stop_signal 是否矛盾
4. 禁止词检查：检查是否包含"建议改为"、"你一定能"等不当表述
5. 编造检测：检查示例参考是否包含用户简历中没有的具体数字或指标

## 16. 对 Prompt / Schema 的后续影响

### 16.1 后续可评估新增字段

| 字段 | 说明 |
|------|------|
| `quality_gate_result` | Quality Gate 检查结果（pass / revise_required / fallback_required / manual_review_suggested） |
| `quality_score` | 内容质量评分（内部使用） |
| `failed_checks` | 未通过的检查项列表 |
| `revision_allowed` | 是否允许 revise |
| `revision_count` | 已执行 revise 次数 |
| `final_action` | 最终动作（return / revise_once / conservative_fallback / mock_fallback） |
| `quality_gate_summary` | 质检摘要（可选，前端展示用） |

### 16.2 重要说明

本阶段不修改 Prompt 文件和 Schema 文件。仅登记后续优化方向。

是否真正修改 Prompt / Schema / Pipeline，放到 4-3.5.H 或后续工程评审阶段决定。

## 17. 对前端展示的影响

### 17.1 可以展示

1. 建议已基于 JD 和简历证据生成
2. AI 建议仅供参考，请基于真实经历核实
3. 当前建议存在信息不足
4. 当前建议为保守结果
5. 当前已触发 fallback，但仍返回可用结果

### 17.2 不建议展示

1. 完整 failed_checks
2. 完整 LLM 质检日志
3. 复杂的内部质量评分
4. 模型原始输出

### 17.3 重要说明

前端展示改造属于阶段 5/6 范围，本阶段不修改前端代码。

## 18. 当前不做事项

| 不做 | 原因 |
|------|------|
| 不修改 Prompt 文件 | 本阶段只做设计登记 |
| 不修改 Schema 文件 | 本阶段只做设计登记 |
| 不修改 Go 代码 | 本阶段只做 docs only |
| 不修改前端代码 | 本阶段只做 docs only |
| 不实现 Quality Gate | 只做设计，4-3.5.H 再评估是否实施 |
| 不实现 quality_gate_result 字段 | 只做枚举定义和流程设计 |
| 不接入 Pipeline | 当前 Pipeline 不含 Quality Gate 阶段 |
| 不调用真实 Hunyuan | 本阶段不涉及 LLM 调用 |
| 不进入 4-3.6 | 本阶段是设计，不是代码开发 |

## 19. 是否需要立即改工程

**本阶段暂不立即改工程代码。**

先完成 Agent 输出 Quality Gate 规则设计。

后续是否修改 Prompt / Schema / Pipeline，放到 4-3.5.H 统一评审。

Quality Gate 的实现路径建议：

1. 先做轻量规则版本（纯代码检查，不调 LLM）
2. 验证规则有效性
3. 再考虑是否接入 LLM Critic

## 20. 与学生求职极端情况的关系

4-3.5.F 已完成学生求职极端情况处理边界设计。Quality Gate 应检查 Agent 是否正确识别极端情况：

| 极端情况 | Quality Gate 检查项 | 不合格判定 |
|---------|-------------------|-----------|
| insufficient_information | 被强行判断为 ready_to_apply | apply_boundary_check 失败 |
| authenticity_risk | 被忽略，输出编造内容 | fabrication_check 失败 |
| target_role_mismatch | 被包装成高匹配 | jd_grounding_check 失败 |
| unsafe_user_request | 被满足，输出虚构表达 | fabrication_check 失败 |
| over_optimization_risk | 继续给大量优化项 | consistency_check 失败 |
| overseas_or_location_risk | 未提及地点冲突 | safety_tone_check 需关注 |

后续 Quality Gate 可增加 extreme_case_check：如果 candidate_case_type != normal_case 但 quality_gate_result=pass，应确认 Agent 是否正确处理了极端情况。详见 `docs/05-Agent设计/学生求职极端情况处理边界.md`

## 21. 下一步建议

**下一步进入 4-3.5.G：灵活在线简历 Profile 与纯文字 Agent 上下文设计。**

不进入 4-3.6。
