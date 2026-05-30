# Prompt 模板

> 最后更新：2026-05-30（阶段 4-3.5.B 战略差异化与杀手锏确认）

## 当前状态

阶段 4-1.4.2 已完成，匹配分析 Prompt 首轮输出验收结论：**partial_pass**。

Prompt 将作为独立产品资产管理：
- 目录 `server/prompts/`（替代原 `server/internal/agent/prompts/` 规划）
- ✅ `match_analysis_v1.md` 已登记 — Match Analysis Agent Prompt Template（4-1.4.1A 人工定稿，4-1.4.1B CodeBuddy 登记）
  - 模板包含 HR 初筛 + 技术面试官双视角
  - 模板包含 Offer鹅展示语气、证据链、P0/P1/P2 优化建议、面试追问预测和真实性提醒
  - 模板要求严格 JSON 输出，包含完整 JSON 结构定义和枚举约束
  - 当前尚未接入 Go 后端和 Hunyuan
  - **4-1.4.2 首轮输出验收结论：partial_pass**
    - JSON 输出合法，17 个顶层字段完整，枚举值全部合法
    - 证据链真实引用 JD 和简历，Offer鹅语气亲和专业，分数克制（72 分）
    - 发现问题：rewrite_example 中少量具体数字超出简历事实，证据链缺少 missing 判断
    - 问题不阻塞进入 4-1.4.3A JSON Schema 定稿
    - 验收报告见 `docs/05-Agent设计/匹配分析Prompt首轮输出验收报告.md`
  - **4-1.4.3A/B 已完成：匹配分析 JSON Schema 定稿与登记**
    - Schema 路径：`server/prompts/schema/match_analysis_schema.json`
    - Prompt 与 Schema 当前关系：
      - Prompt 定义任务、角色、分析原则和输出结构草案
      - Schema 固定输出字段、类型、枚举和数组长度
    - Schema 已吸收 4-1.4.2 首轮验收问题（rewrite_example 禁止编造数字、evidence_chain 支持 missing）
    - 当前仍未接入 Go 后端和 Hunyuan
- 4-1.4.4 开始创建推荐 Prompt 文件 `recommendation_v1.md`
  - **4-1.4.4A 已完成：人工定稿 recommendation_v1.md**
  - **4-1.4.4B 已完成：CodeBuddy 登记 recommendation_v1.md**
  - recommendation_v1.md 是 RecommendationAgent 的 Agent Prompt Template
  - 该模板由人工定稿，CodeBuddy 只登记和检查，不重写核心 Prompt
  - 模板面向"不知道选择哪个岗位"的学生求职场景
  - 模板采用四档推荐：优先投递 / 建议尝试 / 先了解再决定 / 暂不建议
  - 模板包含 Offer鹅展示语气、推荐理由、证据链、风险点、查询建议、下一步行动和真实性提醒
  - **4-1.4.5 首轮输出验收结论：partial_pass**
    - JSON 输出合法，7 个顶层字段完整，4 个推荐项各含 19 个必要字段
    - 证据链真实引用 JD 和简历，Offer鹅语气亲和专业，分数克制（88/42/35/33）
    - 发现问题：四档推荐覆盖不完整（仅使用 priority_apply 和 not_recommended 两档），岗位 13 分类偏保守
    - 问题不阻塞进入 4-1.4.6A 推荐 Schema 定稿
    - 验收报告见 `docs/05-Agent设计/推荐Prompt首轮输出验收报告.md`
  - **4-1.4.6A/B 已完成：推荐 JSON Schema 定稿与登记**
    - Schema 路径：`server/prompts/schema/recommendation_schema.json`
    - Prompt 与 Schema 当前关系：
      - Prompt 定义任务、角色、推荐原则、四档推荐和输出结构草案
      - Schema 固定输出字段、类型、枚举、数组长度和推荐等级一致性约束
    - 当前仍未接入 Go 后端和 Hunyuan
- 4-1.4.7 开始创建简历优化 Prompt 文件 `resume_optimization_v1.md`
  - **4-1.4.7A 已完成：人工定稿 resume_optimization_v1.md**
  - **4-1.4.7B 已完成：CodeBuddy 登记 resume_optimization_v1.md**
  - resume_optimization_v1.md 是 ResumeOptimizationAgent 的 Agent Prompt Template
  - 该模板由人工定稿，CodeBuddy 只登记和检查，不重写核心 Prompt
  - 模板面向"目标岗位导向的简历优化"场景
  - 模板支持简历 + 目标岗位 JD + 可选匹配分析结果输入
  - 模板输出 P0/P1/P2 优化建议、关键词对齐、项目优化、技能栏优化、摘要优化、缺失证据补足计划、改写示例和真实性提醒
  - 当前尚未接入 Go 后端和 Hunyuan
  - **4-1.4.8 首轮输出验收结论：partial_pass**
    - JSON 输出合法，15 个顶层字段完整，枚举值全部合法
    - P0/P1/P2 优化建议具体且围绕目标岗位（p0=2, p1=2, p2=1）
    - keyword_alignment 正确区分 present/partial/missing（7 项）
    - rewrite_examples 全部 requires_user_verification=true，保守但正确
    - Offer鹅语气亲和专业，真实性提醒完整
    - 发现轻微问题：rewrite_examples 全部需要用户确认，无直接可用示例
    - 问题不阻塞进入 4-1.4.9A 简历优化 Schema 定稿
    - 验收报告见 `docs/05-Agent设计/简历优化Prompt首轮输出验收报告.md`
  - **4-1.4.9A/B 已完成：简历优化 JSON Schema 定稿与登记**
    - Schema 路径：`server/prompts/schema/resume_optimization_schema.json`
    - Prompt 与 Schema 当前关系：
      - Prompt 定义任务、角色、简历优化原则、P0/P1/P2 优先级、改写约束和输出结构草案
      - Schema 固定输出字段、类型、枚举、数组长度和真实性约束
    - Schema 已吸收 4-1.4.8 首轮验收问题（允许 rewrite_examples 全部 can_use_directly=false、涉及数字/指标等时必须 requires_user_verification=true、不得编造具体数字/指标/系统规模/职责或成果、p2 允许 0-4 项）
    - 当前仍未接入 Go 后端和 Hunyuan
- JSON 修复 Prompt `json_repair_v1.md`
  - **4-1.4.10A/B 已完成：JSON 修复 Agent 指令模板定稿与登记**
  - json_repair_v1.md 是 JSONRepairAgent 的 Agent Prompt Template
  - 该模板由人工定稿，CodeBuddy 只登记和检查，不重写核心 Prompt
  - 模板面向"LLM JSON 输出结构修复"场景
  - 模板支持 match_analysis / recommendation / resume_optimization 三类业务输出修复
  - 模板输入 TARGET_OUTPUT_TYPE / TARGET_SCHEMA_JSON / BROKEN_MODEL_OUTPUT / VALIDATION_ERRORS / REPAIR_CONTEXT
  - 模板只修复结构，不重新做业务分析，不改变业务结果来源
  - 当前尚未接入 Go 后端和 Hunyuan
  - **4-1.4.11 已完成：失败兜底与 Mock 回退策略**
    - JSONRepairAgent 已纳入失败兜底链路
    - json_repair_v1.md 只负责结构修复，不重新做业务分析
    - repair 成功后业务 generated_by 仍保持 llm_agent
    - repair 失败后进入 Mock Agent fallback
    - 策略文档见 `docs/05-Agent设计/失败兜底与Mock回退策略.md`
  - **4-1.4.12 已完成：阶段 4-1.4 总验收**
    - 四个 Prompt 资产均已登记：match_analysis_v1.md、recommendation_v1.md、resume_optimization_v1.md、json_repair_v1.md
    - 三个 Schema 资产均已登记且 JSON 合法
    - 三个首轮输出验收均为 partial_pass，问题已被 Schema 吸收
    - 失败兜底与 Mock 回退策略已通过验收
    - 验收报告见 `docs/05-Agent设计/阶段4-1.4总验收报告.md`
  - **4-1.4.13 已完成：阶段 4-1.4 汇总**
    - 阶段 4-1.4 已正式闭环
    - 四个 Prompt 资产均已定稿并登记：match_analysis_v1.md、recommendation_v1.md、resume_optimization_v1.md、json_repair_v1.md
    - 当前均未接入 Go 后端
    - 下一阶段需要实现 Prompt Loader
    - 汇总文档见 `docs/00-项目总览/阶段4-1.4Prompt与Agent真实接入汇总.md`
- JSON Schema 文件将存放在 `server/prompts/schema/` 目录
- 方案文档见 `docs/05-Agent设计/阶段4-1.4Prompt与Agent真实接入方案.md`

## 阶段 4-2 更新

4-2.0 已设计 Prompt Loader，后续 Prompt Loader 将从 `server/prompts/` 加载以下 Prompt：

| PromptName | 文件路径 |
|-----------|---------|
| match_analysis | server/prompts/match_analysis_v1.md |
| recommendation | server/prompts/recommendation_v1.md |
| resume_optimization | server/prompts/resume_optimization_v1.md |
| json_repair | server/prompts/json_repair_v1.md |

Prompt Loader 需支持：按 PromptName 加载、校验文件存在、替换占位符、保留原始模板、不把 Prompt 写死在 Go 代码、支持后续版本化。

4-2.7 已将 json_repair_v1.md 接入 JSONRepairAgent 调用链路。该 Prompt 仍保持人工定稿内容不变，JSONRepairAgent 通过 Prompt Loader 加载并渲染 5 个占位符（TARGET_OUTPUT_TYPE / TARGET_SCHEMA_JSON / BROKEN_MODEL_OUTPUT / VALIDATION_ERRORS / REPAIR_CONTEXT）。

4-2.2 已完成 Prompt Loader，支持逻辑名称：

- match_analysis
- recommendation
- resume_optimization
- json_repair

Prompt Loader 从 server/prompts/ 读取 Prompt 文件，不再要求后续业务代码硬编码 Prompt 内容。

4-2.2 才开始实现 Prompt Loader。

后续 Prompt 将以腾讯真实岗位 JD 为输入：
- 推荐 Prompt 和匹配 Prompt 需要分别设计
- JSON 输出 Schema 需要固定
- 失败兜底和 Mock 回退需要保留
- data_quality 分层已建立，Prompt 可针对不同质量等级差异化处理

## 模板管理规范

所有 Prompt 模板统一管理，便于版本控制和调优。

### 文件结构

```
server/internal/agent/prompts/
├── resume_extract.md     # 简历提取 Prompt（阶段 4）
├── job_recommend.md      # 岗位推荐 Prompt（阶段 4）
├── job_match.md          # 岗位匹配 Prompt（阶段 4）
└── version_compare.md    # 版本对比 Prompt（阶段 4）
```

### 模板格式

使用 Go template 语法，变量用 `{{.Variable}}` 标记。

## 设计原则

1. **不虚构**：Prompt 中明确要求不编造不存在的经历
2. **有证据**：所有结论必须基于简历中的真实内容
3. **可追溯**：每条建议都要指出对应的简历原文
4. **有建设性**：不仅指出问题，还给出改进方向

## 阶段 4 Prompt 必须包含的约束

1. **只输出 JSON**：必须要求模型只输出 JSON，不要 Markdown、不要解释性前后缀、不要代码块
2. **禁止事项**：不虚构经历、不编造不存在的项目或能力、不生成优化版简历
3. **真实性约束**：所有分析必须基于用户简历中实际存在的内容
4. **分数范围**：分数必须是 0-100 整数
5. **必填字段**：明确列出输出 JSON 中必须包含的字段
6. **JSON 结构示例**：在 Prompt 中提供完整的 JSON 输出结构示例

### Prompt 通用结构

```
你是一个简历分析助手。

## 约束
1. 只输出 JSON，不要输出任何其他内容
2. 不要输出 Markdown 格式
3. 不要输出解释性前后缀
4. 不要输出代码块包裹
5. 不要虚构用户简历中不存在的经历
6. 不要自动生成优化版简历
7. 所有分析必须基于用户简历中实际存在的内容

## 输出格式
{JSON 结构示例}

## 输入数据
{业务数据}

## 分析要求
{具体分析要求}
```

## 各 Agent Prompt 方向

### 简历提取 Prompt

- 输入：结构化简历数据
- 输出：关键技能标签、求职方向推断、表达风险列表
- 阶段 4 暂不独立实现

### 岗位推荐 Prompt

- 输入：简历提取结果 + 岗位列表
- 输出：推荐岗位列表 + 推荐理由 + 推荐等级
- 4-5 实现

### 岗位匹配 Prompt

- 输入：简历数据 + 目标岗位 JD
- 输出：匹配分数 + 证据链 + 风险点 + 优化建议
- **核心 Prompt**，4-4 重点实现
- 必须包含 JD-简历证据链构建要求

### 版本对比 Prompt

- 输入：旧版本 + 新版本简历
- 输出：差异分析 + 改进评价 + 仍需改进
- 4-6 实现

## 后续待补充

- 各 Agent 的真实 Prompt 模板内容（4-4 / 4-5 / 6 实现时编写）
- 模板版本管理方案
- Prompt 调优记录

## 关联文档

- [Agent总览.md](./Agent总览.md)
- [阶段4Agent与模型接入方案.md](./阶段4Agent与模型接入方案.md)
- [LLMClient设计.md](./LLMClient设计.md)
- [JSON输出规范.md](./JSON输出规范.md)
- [模型接入.md](./模型接入.md)
- [../02-后端设计/阶段3核心业务后端Mock方案.md](../02-后端设计/阶段3核心业务后端Mock方案.md)

## 阶段 4-3.0 更新

4-3.0 已完成 PromptValues 组装策略设计（docs only）。关键策略：

### 各 Agent 的 PromptValues 组装规则

|| Agent | PromptName | 占位符 | 组装方 |
||-------|-----------|--------|-------|
|| MatchAnalysisAgent | match_analysis | `RESUME_CONTEXT_JSON` / `JOB_CONTEXT_JSON` / `BUSINESS_CONTEXT` | MatchService |
|| RecommendationAgent | recommendation | `RESUME_CONTEXT_JSON` / `JOB_CANDIDATES_JSON` / `BUSINESS_CONTEXT` | RecommendationService |
|| ResumeOptimizationAgent | resume_optimization | `RESUME_CONTEXT_JSON` / `TARGET_JOB_CONTEXT_JSON` / `MATCH_ANALYSIS_JSON` / `BUSINESS_CONTEXT` | 待实现 |

### PromptValues 组装原则

1. **Service 层负责组装**：从数据库读取业务数据，序列化为 JSON 后填入占位符
2. **BUSINESS_CONTEXT 统一格式**：包含 user_id、resume_version_id、job_id 等业务上下文
3. **序列化失败触发 Mock fallback**：占位符组装异常不应导致接口崩溃
4. **不把完整简历/JD 写入日志**：PromptValues 仅传递给 Pipeline，不记录完整内容

方案文档见 `docs/05-Agent设计/阶段4-3业务Agent接入LLM方案.md`

## 阶段 4-3.4 更新

4-3.4 已将 recommendation_v1.md 用于 RecommendationAgent Pipeline 调用。Prompt 文件本身未修改。PromptValues 组装包含 RESUME_CONTEXT_JSON / JOB_CANDIDATES_JSON / BUSINESS_CONTEXT 三个占位符。

## 阶段 4-3.5.B 更新

4-3.5.B 已完成战略差异化确认。后续 Prompt 优化需强化：

1. **JD 证据链**：每条建议都要绑定 JD 要求和简历证据
2. **投递边界**：输出 apply_readiness 判断
3. **优化停止条件**：输出 optimization_stop_signal
4. **示例参考**：改写建议使用"示例参考"而非"建议改为"
5. **用户核实提醒**：每条建议附带 verification_notice
6. **学习关键词**：附带 learn_more_query
7. **自查问题**：附带 self_check_question

本阶段不修改 Prompt 文件，只登记优化方向。

详见 `docs/05-Agent设计/战略差异化与杀手锏确认.md`

## 阶段 4-3.5.C 更新

4-3.5.C 已完成投递边界与优化停止条件设计。后续 Prompt 优化需加入：

1. **apply_readiness**：投递准备度判断，匹配分析和推荐结果中输出
2. **optimization_stop_signal**：优化停止信号，避免无限优化
3. **Top 3 快速优化项**：apply_readiness=quick_fix_before_apply 时最多给 3 个快速优化项
4. **长期提升项拆分**：long_term_improvement 与 top_fix_items 分开展示
5. **estimated_effort**：每个优化项的预计耗时（within_10_min / within_30_min / within_60_min / long_term）

本阶段不修改 Prompt 文件。

## 阶段 4-3.5.G 灵活在线简历 Profile 与纯文字 Agent 上下文设计

4-3.5.G 已完成灵活在线简历 Profile 与纯文字 Agent 上下文设计。后续 Prompt 需适配新的输入格式：

1. **占位符迁移**：`{{RESUME_CONTEXT_JSON}}` → `{{RESUME_CONTEXT_TEXT}}` + `{{PROFILE_SUMMARY_JSON}}` + `{{PRIVACY_FLAGS_JSON}}`
2. **简历输入说明**：从"以下是用户的简历 JSON"改为"以下是用户简历的结构化纯文字上下文"
3. **隐私约束**：增加"简历上下文中不包含姓名、手机号、邮箱等隐私字段，不要在输出中推测或编造这些信息"
4. **概况辅助**：增加"profile_summary 提供了简历概况信号，可直接用于快速判断"
5. **Section 引用**：增加"引用简历证据时，尽量引用具体的 Section 标题和内容"
6. **简历优化额外占位符**：增加 `{{CURRENT_TARGET_SECTION}}` 表示当前优化的目标 Section

本阶段不修改 Prompt 文件，只做设计登记。

详见 `docs/05-Agent设计/灵活在线简历Profile与纯文字Agent上下文设计.md`

详见 `docs/05-Agent设计/投递边界与优化停止条件设计.md`

## 阶段 4-3.5.D 更新

4-3.5.D 已完成 JD 证据链匹配与 START 简历优化规则设计。后续 Prompt 优化需加入：

1. **jd_evidence_chain**：JD 要求 → 简历证据 → 证据强度 → 匹配判断 → 差距风险 → 优化动作
2. **evidence_strength**：strong / medium / weak / missing
3. **match_judgement**：matched / partially_matched / missing / risky
4. **screening_risk**：keyword_missing / responsibility_unclear / result_missing 等九类风险
5. **START / STAR-T 技术岗表达法**：检查简历经历是否缺少 Situation / Task / Action / Result / Tech
6. **示例参考**：使用"示例参考"替代"建议改为"，每条附带核实提醒
7. **self_check_question / learn_more_query / verification_notice**：建议可核实机制

本阶段不修改 Prompt 文件。

## 阶段 4-3.5.G 灵活在线简历 Profile 与纯文字 Agent 上下文设计

4-3.5.G 已完成灵活在线简历 Profile 与纯文字 Agent 上下文设计。后续 Prompt 需适配新的输入格式：

1. **占位符迁移**：`{{RESUME_CONTEXT_JSON}}` → `{{RESUME_CONTEXT_TEXT}}` + `{{PROFILE_SUMMARY_JSON}}` + `{{PRIVACY_FLAGS_JSON}}`
2. **简历输入说明**：从"以下是用户的简历 JSON"改为"以下是用户简历的结构化纯文字上下文"
3. **隐私约束**：增加"简历上下文中不包含姓名、手机号、邮箱等隐私字段，不要在输出中推测或编造这些信息"
4. **概况辅助**：增加"profile_summary 提供了简历概况信号，可直接用于快速判断"
5. **Section 引用**：增加"引用简历证据时，尽量引用具体的 Section 标题和内容"
6. **简历优化额外占位符**：增加 `{{CURRENT_TARGET_SECTION}}` 表示当前优化的目标 Section

本阶段不修改 Prompt 文件，只做设计登记。

详见 `docs/05-Agent设计/灵活在线简历Profile与纯文字Agent上下文设计.md`

详见 `docs/05-Agent设计/JD证据链匹配与START简历优化规则.md`

## 阶段 4-3.5.H 战略转型阶段汇总与工程评审

4-3.5.H 工程评审结论：暂不立即修改 Prompt 文件。
后续可单独评估是否加入 apply_readiness、jd_evidence_chain、quality_gate_result、candidate_case_type、ResumeContextText 等规则。

详见 `docs/05-Agent设计/阶段4-3.5战略转型汇总与工程评审.md`

## 阶段 4-3.5.E 更新

4-3.5.E 已完成 Agent 输出 Quality Gate 设计。后续 Prompt 优化需配合 Quality Gate 检查项：

1. **quality_gate_result**：后续 Schema 可评估增加 Quality Gate 检查结果字段
2. **revise once**：当 Quality Gate 判定 revise_required 时，需要将 failed_checks 反馈给生成 Agent
3. **Prompt 需强化**：确保输出包含 JD 证据链、apply_readiness、optimization_stop_signal、verification_notice 等 Quality Gate 检查所需字段
4. **保守输出**：Prompt 应引导模型在不确定时输出保守结果，避免触发 fallback_required

本阶段不修改 Prompt 文件。

## 阶段 4-3.5.G 灵活在线简历 Profile 与纯文字 Agent 上下文设计

4-3.5.G 已完成灵活在线简历 Profile 与纯文字 Agent 上下文设计。后续 Prompt 需适配新的输入格式：

1. **占位符迁移**：`{{RESUME_CONTEXT_JSON}}` → `{{RESUME_CONTEXT_TEXT}}` + `{{PROFILE_SUMMARY_JSON}}` + `{{PRIVACY_FLAGS_JSON}}`
2. **简历输入说明**：从"以下是用户的简历 JSON"改为"以下是用户简历的结构化纯文字上下文"
3. **隐私约束**：增加"简历上下文中不包含姓名、手机号、邮箱等隐私字段，不要在输出中推测或编造这些信息"
4. **概况辅助**：增加"profile_summary 提供了简历概况信号，可直接用于快速判断"
5. **Section 引用**：增加"引用简历证据时，尽量引用具体的 Section 标题和内容"
6. **简历优化额外占位符**：增加 `{{CURRENT_TARGET_SECTION}}` 表示当前优化的目标 Section

本阶段不修改 Prompt 文件，只做设计登记。

详见 `docs/05-Agent设计/灵活在线简历Profile与纯文字Agent上下文设计.md`

详见 `docs/05-Agent设计/Agent输出QualityGate设计.md`

## 阶段 4-3.5.F 更新

4-3.5.F 已完成学生求职极端情况处理边界设计。后续 Prompt 优化需加入极端情况判断规则：

1. **candidate_case_type 识别**：Prompt 应引导模型识别候选人情况类型（信息不足 / 弱简历 / 岗位不匹配 / 海外风险 / 优秀候选人 / 求职目标不清晰 / 真实性风险 / 过度优化 / 造假请求）
2. **禁止表达规则**：禁止"你不适合互联网""你的能力太差""你价值观不符"等表达
3. **推荐表达规则**：使用"当前简历对该岗位关键要求的覆盖不足""当前信息不足，无法稳定判断"等客观表达
4. **处理策略映射**：不同极端情况对应不同处理策略（clarify_before_recommendation / quick_fix_then_apply / recommend_alternative_jobs / conservative_advice / refuse_fabrication）
5. **case_risk_level**：高风险（high/critical）情况应触发保守输出或 fallback

本阶段不修改 Prompt 文件。

## 阶段 4-3.5.G 灵活在线简历 Profile 与纯文字 Agent 上下文设计

4-3.5.G 已完成灵活在线简历 Profile 与纯文字 Agent 上下文设计。后续 Prompt 需适配新的输入格式：

1. **占位符迁移**：`{{RESUME_CONTEXT_JSON}}` → `{{RESUME_CONTEXT_TEXT}}` + `{{PROFILE_SUMMARY_JSON}}` + `{{PRIVACY_FLAGS_JSON}}`
2. **简历输入说明**：从"以下是用户的简历 JSON"改为"以下是用户简历的结构化纯文字上下文"
3. **隐私约束**：增加"简历上下文中不包含姓名、手机号、邮箱等隐私字段，不要在输出中推测或编造这些信息"
4. **概况辅助**：增加"profile_summary 提供了简历概况信号，可直接用于快速判断"
5. **Section 引用**：增加"引用简历证据时，尽量引用具体的 Section 标题和内容"
6. **简历优化额外占位符**：增加 `{{CURRENT_TARGET_SECTION}}` 表示当前优化的目标 Section

本阶段不修改 Prompt 文件，只做设计登记。

详见 `docs/05-Agent设计/灵活在线简历Profile与纯文字Agent上下文设计.md`

## 阶段 4-3.5.G 灵活在线简历 Profile 与纯文字 Agent 上下文设计

4-3.5.G 已完成灵活在线简历 Profile 与纯文字 Agent 上下文设计。后续 Prompt 需适配新的输入格式：

1. **占位符迁移**：`{{RESUME_CONTEXT_JSON}}` → `{{RESUME_CONTEXT_TEXT}}` + `{{PROFILE_SUMMARY_JSON}}` + `{{PRIVACY_FLAGS_JSON}}`
2. **简历输入说明**：从"以下是用户的简历 JSON"改为"以下是用户简历的结构化纯文字上下文"
3. **隐私约束**：增加"简历上下文中不包含姓名、手机号、邮箱等隐私字段，不要在输出中推测或编造这些信息"
4. **概况辅助**：增加"profile_summary 提供了简历概况信号，可直接用于快速判断"
5. **Section 引用**：增加"引用简历证据时，尽量引用具体的 Section 标题和内容"
6. **简历优化额外占位符**：增加 `{{CURRENT_TARGET_SECTION}}` 表示当前优化的目标 Section

本阶段不修改 Prompt 文件，只做设计登记。

详见 `docs/05-Agent设计/灵活在线简历Profile与纯文字Agent上下文设计.md`

详见 `docs/05-Agent设计/学生求职极端情况处理边界.md`
