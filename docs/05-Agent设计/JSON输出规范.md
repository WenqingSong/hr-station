# JSON 输出规范

> 最后更新：2026-05-30（阶段 4-3.5.B 战略差异化与杀手锏确认）

## 1. 为什么必须强制 JSON

1. **结构化消费**：前端和后端必须以结构化方式消费 Agent 输出，不能解析自由文本
2. **字段校验**：只有 JSON 才能做字段完整性、分数范围、真实性校验
3. **稳定替换**：阶段 3 Mock Agent 已定义 JSON 结构，真实 LLM 必须输出相同结构
4. **兜底判断**：只有 JSON 才能判断输出是否合规，决定是否回退 Mock

## 2. 通用输出要求

所有 Agent 的 LLM 输出必须满足：

1. **只输出 JSON**：不要 Markdown、不要解释性前后缀、不要代码块包裹
2. **合法 JSON**：必须是可被 `json.Unmarshal` 解析的合法 JSON
3. **分数为 0-100 整数**：所有分数字段必须是 0-100 之间的整数
4. **必填字段不可缺**：每个 Agent 有明确的必填字段列表
5. **不虚构经历**：不允许输出用户简历中不存在的经历或能力
6. **不自动生成简历**：不允许输出优化版简历内容
7. **包含真实性提示**：必须包含 `authenticity_notice` 字段

## 3. 推荐 Agent JSON 结构

```json
{
  "recommended_jobs": [
    {
      "job_id": 1,
      "title": "Go 后端开发工程师",
      "department": "技术部",
      "city": "深圳",
      "category": "后端开发",
      "job_type": "校招",
      "priority": "strong_recommend",
      "recommendation_score": 85,
      "recommendation_reason": "...",
      "match_highlights": ["...", "..."],
      "risk_points": ["..."],
      "suggested_action": "..."
    }
  ],
  "summary": "..."
}
```

### 必填字段

| 字段 | 类型 | 约束 |
|------|------|------|
| `recommended_jobs` | array | 至少 1 个元素 |
| `recommended_jobs[].job_id` | int | 必须与输入岗位 ID 一致 |
| `recommended_jobs[].priority` | string | 只允许 `strong_recommend` / `recommend` / `try` |
| `recommended_jobs[].recommendation_score` | int | 0-100 |
| `recommended_jobs[].recommendation_reason` | string | 非空 |
| `recommended_jobs[].match_highlights` | array | 至少 1 个元素 |
| `recommended_jobs[].risk_points` | array | 可为空 |
| `recommended_jobs[].suggested_action` | string | 非空 |
| `summary` | string | 非空 |

## 4. 匹配分析 Agent JSON 结构

```json
{
  "job_id": 1,
  "job_title": "Go 后端开发工程师",
  "resume_version_id": 2,
  "resume_version_no": 2,
  "match_score": 78,
  "screening_result": "likely",
  "evidence_chain": [
    {
      "requirement": "岗位要求 Go 相关经验",
      "resume_point": "技能列表中包含Go",
      "judgement": "strong_match"
    }
  ],
  "match_highlights": ["...", "..."],
  "risk_points": ["..."],
  "optimization_suggestions": [
    {
      "priority": "P0",
      "title": "...",
      "reason": "...",
      "action": "..."
    }
  ],
  "next_steps": ["...", "..."],
  "authenticity_notice": "以下建议仅用于表达优化，请只补充真实经历，不要虚构项目或能力。"
}
```

### 必填字段

| 字段 | 类型 | 约束 |
|------|------|------|
| `match_score` | int | 0-100 |
| `screening_result` | string | 只允许 `highly_likely` / `likely` / `possible` / `risky` |
| `evidence_chain` | array | 至少 1 个元素 |
| `evidence_chain[].requirement` | string | 非空 |
| `evidence_chain[].resume_point` | string | 非空 |
| `evidence_chain[].judgement` | string | 只允许 `strong_match` / `partial_match` / `missing` |
| `match_highlights` | array | 至少 1 个元素 |
| `risk_points` | array | 可为空 |
| `optimization_suggestions` | array | 至少 1 个元素 |
| `optimization_suggestions[].priority` | string | 只允许 `P0` / `P1` |
| `optimization_suggestions[].title` | string | 非空 |
| `optimization_suggestions[].reason` | string | 非空 |
| `optimization_suggestions[].action` | string | 非空 |
| `next_steps` | array | 至少 1 个元素 |
| `authenticity_notice` | string | 必须包含"真实经历"相关提示 |

## 5. 版本对比 Agent JSON 结构

```json
{
  "from_version_no": 1,
  "to_version_no": 2,
  "score_before": 65,
  "score_after": 78,
  "score_change": 13,
  "field_changes": [
    {
      "field": "skills",
      "before": "...",
      "after": "...",
      "change_type": "modified",
      "impact": "..."
    }
  ],
  "risk_changes": [
    {
      "risk": "...",
      "before": "...",
      "after": "...",
      "trend": "reduced"
    }
  ],
  "suggestion_completion": [
    {
      "suggestion": "...",
      "status": "partially_completed",
      "evidence": "..."
    }
  ],
  "key_improvements": ["...", "..."],
  "remaining_risks": ["..."],
  "summary": "...",
  "authenticity_notice": "版本对比仅用于呈现表达变化，请只补充真实经历，不要虚构项目或能力。"
}
```

### 必填字段

| 字段 | 类型 | 约束 |
|------|------|------|
| `score_before` | int | 0-100 |
| `score_after` | int | 0-100 |
| `score_change` | int | = score_after - score_before |
| `field_changes` | array | 至少 1 个元素 |
| `field_changes[].field` | string | 非空 |
| `field_changes[].change_type` | string | 只允许 `added` / `modified` / `removed` / `unchanged` |
| `risk_changes` | array | 可为空 |
| `risk_changes[].trend` | string | 只允许 `reduced` / `unchanged` / `stable` / `increased` |
| `suggestion_completion` | array | 可为空 |
| `suggestion_completion[].status` | string | 只允许 `completed` / `partially_completed` / `not_started` |
| `key_improvements` | array | 至少 1 个元素 |
| `remaining_risks` | array | 可为空 |
| `summary` | string | 非空 |
| `authenticity_notice` | string | 必须包含"真实经历"相关提示 |

## 6. 字段校验规则

### 6.1 校验流程

```
LLM 输出文本
  → 提取 JSON（去除可能的 Markdown 包裹）
  → json.Unmarshal
  → 必填字段检查
  → 枚举值检查
  → 分数范围检查
  → 真实性检查
  → 通过 / 失败
```

### 6.2 校验失败处理

| 校验类型 | 失败处理 |
|---------|---------|
| JSON 解析失败 | 回退 Mock |
| 缺必填字段 | 回退 Mock |
| 枚举值不在允许范围 | 忽略该条目或回退 Mock |
| 分数超范围 | 截断到 0-100 |

## 7. 分数范围规则

1. 所有分数字段必须为整数
2. 分数范围为 0-100
3. 小于 0 的分数截断为 0
4. 大于 100 的分数截断为 100
5. 非整数分数截断为整数
6. 版本对比的 `score_change` 必须等于 `score_after - score_before`

## 8. 禁止输出内容

以下内容如果出现在 LLM 输出中，该输出视为无效，触发回退 Mock：

1. 自动生成优化版简历（如"优化后的简历为：..."）
2. 虚构不存在的经历（如"你在XX公司实习期间..."但简历中无此经历）
3. 与用户简历无关的内容
4. Markdown 格式（如 `# 标题`、`**加粗**` 等）
5. 解释性前后缀（如"以下是分析结果："、"希望对你有帮助！"）
6. 代码块包裹（如 ` ```json ... ``` `）

## 9. 解析失败处理

1. 尝试直接解析 JSON
2. 如果失败，尝试去除 Markdown 代码块标记后再解析
3. 如果仍失败，记录原始输出到日志（便于调试 Prompt）
4. 回退 Mock Agent
5. 响应中 `generated_by` 字段标记为 `llm_agent_fallback`

## 10. 回退 Mock 策略

| 条件 | 行为 |
|------|------|
| `ENABLE_LLM_AGENT=false` | 始终使用 Mock Agent |
| `ENABLE_LLM_AGENT=true` 且 LLM 成功 | 使用 LLM 结果，`generated_by="llm_agent"` |
| `ENABLE_LLM_AGENT=true` 且 LLM 失败 且 `ENABLE_LLM_FALLBACK=true` | 回退 Mock，`generated_by="llm_agent_fallback"` |
| `ENABLE_LLM_AGENT=true` 且 LLM 失败 且 `ENABLE_LLM_FALLBACK=false` | 返回错误 |

## 阶段 4-2 更新

4-2.0 已设计 Schema Loader、JSON parse、Schema validate 和 JSON repair once 链路，详细方案见 `docs/05-Agent设计/阶段4-2LLMClient与Hunyuan接入方案.md`。

后续 4-2 将实现以下基础设施能力：

1. **Schema Loader**：从 `server/prompts/schema/` 加载 JSON Schema 文件
2. **JSON parse**：从模型 raw output 提取合法 JSON（去除 Markdown code fence、提取 JSON object）
3. **Schema validate**：推荐使用 `github.com/santhosh-tekuri/jsonschema/v5` 校验（支持 draft 2020-12）
4. **JSON repair once**：JSONRepairAgent 调用一次修复，repair 不改变 generated_by
5. **Mock fallback**：repair 失败后进入 Mock fallback

4-2.1 已新增 `LLM_SCHEMA_DIR=server/prompts/schema`、`LLM_REPAIR_ENABLE=true`、`LLM_REPAIR_MAX_ATTEMPTS=1` 配置，后续 Schema Loader 和 JSON repair 链路将基于这些配置实现。

4-2.3 已完成 JSON parse + Schema validate 工具。

模型输出不允许直接返回，必须经过：
raw output → extract JSON object → json.Unmarshal → Schema validate
通过后才允许进入后续业务返回链路。
失败后后续 4-2.7 进入 JSONRepairAgent。

4-2.4 已完成 LLM Client 抽象。`ChatResponse.RawContent` 后续必须经过 4-2.3 的 JSON extract / parse / Schema validate 工具处理，不能直接返回业务接口。

4-2.5 已完成 HunyuanClient。HunyuanClient 返回的 ChatResponse.RawContent 仍不得直接返回业务接口，后续必须进入 JSON extract / parse / Schema validate 链路。

4-2.7 已完成 JSON repair once 工程能力。模型输出处理链路变为：
raw output → extract / parse → schema validate → repair once → repair output extract / parse → schema validate
repair 仍失败时交由后续 fallback decision。

4-2.8 已完成统一 LLM Agent Pipeline。完整执行链路：
Prompt Render → LLM Call → JSON extract/parse → Schema validate → repair once → fallback decision
Pipeline 返回 RunResult，包含 AgentStatus（success / repaired / fallback_required）和完整诊断信息。

新增：
- `server/internal/agent/pipeline/` 包：Pipeline Runner
- AgentStatus 三种状态：success / repaired / fallback_required
- 20+ Pipeline 错误码
- 18 项单元测试全部通过

4-2.6 smoke test 会尝试对 Hunyuan raw output 执行 llmjson ExtractAndParse。json_parse_success 用于观察模型是否遵守严格 JSON 输出，但 smoke test 的核心仍是连通性。JSON 合法性失败可以记录为 warning，后续 4-2.7 / 4-2.8 处理。

新增：
- `server/internal/llm/hunyuan_client.go`：HunyuanClient 实现 llm.Client 接口
- `server/internal/llm/hunyuan_types.go`：OpenAI-compatible 内部请求/响应类型

新增：
- `server/internal/llm/` 包：Client 接口 + ChatRequest / ChatResponse + Config + ErrorCode + NoopClient
- 业务 Agent 后续只依赖 llm.Client，不直接依赖 Hunyuan

新增工具：
- `server/internal/llmjson/` 包：ExtractObject / ParseObject / ExtractAndParse，支持去除 Markdown code fence、提取 JSON object、错误码分类
- `server/internal/schema/validator.go`：Validator.Validate / Validator.ValidateJSON，使用 santhosh-tekuri/jsonschema/v5 实现完整 JSON Schema 校验

4-2.2 已完成 Schema Loader，支持逻辑名称：

- match_analysis
- recommendation
- resume_optimization

Schema Loader 当前只负责加载 JSON Schema 文件和检查 JSON 合法性，完整 Schema validate 放到 4-2.3。

4-2.2 实现 Schema Loader，4-2.3 实现 JSON parse + Schema validate，4-2.7 实现 JSON repair 调用链路。

## 阶段 4-1.4 更新

4-1.4 将固定推荐 / 匹配 / 简历优化 JSON Schema：

- 所有模型输出必须可解析为合法 JSON
- 失败时回退 Mock Agent（generated_by=llm_agent_fallback）
- JSON Schema 文件将存放在 `server/prompts/schema/` 目录
- 匹配分析 Schema 优先固定（4-1.4.3）
- Schema 与 Prompt 配套，每个 Prompt 对应一个 JSON Schema
- `match_analysis_v1.md` 已包含匹配分析 JSON 输出结构草案（4-1.4.1B 登记）
- 正式 JSON Schema 将在 4-1.4.3 固定
- 当前 4-1.4.1B 只登记 Prompt 中的输出结构，不创建 schema 文件
- **4-1.4.2 首轮输出验收已完成，结论：partial_pass**
  - 模型输出 JSON 合法，17 个顶层字段完整
  - 验收报告见 `docs/05-Agent设计/匹配分析Prompt首轮输出验收报告.md`
  - 验收结果将作为 match_analysis_schema.json 的依据
- **4-1.4.3A 已完成：人工定稿 match_analysis_schema.json**
- **4-1.4.3B 已完成：CodeBuddy 登记匹配分析 JSON Schema**
- Schema 路径：`server/prompts/schema/match_analysis_schema.json`
- 该 Schema 约束 MatchAnalysisAgent 的 LLM 输出
- 当前 Schema 包含 17 个顶层 required 字段
- 当前 Schema 禁止 additionalProperties
- 当前 Schema 已吸收 4-1.4.2 首轮验收问题：
  - rewrite_example 禁止编造具体数字，建议条件式表达
  - evidence_chain 支持 missing 判断
  - 后续 LLM_MAX_OUTPUT_TOKENS 建议 ≥8192
- **4-1.4.7A/B 已完成：简历优化 Agent 指令模板定稿与登记**
  - resume_optimization_v1.md 已包含简历优化输出结构草案
  - 正式简历优化 JSON Schema 将在 4-1.4.9A/B 固定
  - 当前 4-1.4.7B 只登记 Prompt 中的输出结构，不创建简历优化 schema 文件
- **4-1.4.8 已完成：简历优化 Prompt 首轮输出验收（partial_pass）**
  - 模型输出 JSON 合法，15 个顶层字段完整
  - 验收结果将作为 resume_optimization_schema.json 的依据
  - 验收报告见 `docs/05-Agent设计/简历优化Prompt首轮输出验收报告.md`
- **4-1.4.9A/B 已完成：简历优化 JSON Schema 定稿与登记**
  - Schema 路径：`server/prompts/schema/resume_optimization_schema.json`
  - 该 Schema 约束 ResumeOptimizationAgent 的 LLM 输出
  - 当前 Schema 包含 15 个顶层 required 字段
  - 当前 Schema 禁止 additionalProperties
  - 当前 Schema 已固定：
    - P0 / P1 / P2 优先级建议
    - keyword_alignment
    - project_optimization
    - skill_section_optimization
    - summary_optimization
    - missing_evidence_plan
    - rewrite_examples
    - next_actions
    - authenticity_notice
  - 当前 Schema 已吸收 4-1.4.8 首轮验收问题：
    - 允许 rewrite_examples 全部 can_use_directly=false
    - 涉及数字/指标/规模/职责/成果时，输入无明确事实则必须 requires_user_verification=true
    - 不得编造具体数字、指标、系统规模、职责或成果
    - p2 允许 0-4 项
- **4-1.4.10A/B 已完成：JSON 修复 Agent 指令模板定稿与登记**
  - json_repair_v1.md 已登记为 JSON 修复 Prompt
  - 用途：当业务 Agent 输出不合法 JSON 或不符合 Schema 时，尝试进行一次保守修复
  - 修复范围：
    - Markdown / code fence 包裹
    - JSON 语法错误
    - 字段缺失
    - 字段类型错误
    - 枚举不合法
    - 额外字段
    - 数组/对象结构错误
  - 边界：
    - 不重新做业务分析
    - 不编造岗位、经历、项目指标、系统规模、职责范围
    - generated_by 仍保持 llm_agent
    - 修复失败后后端进入 Mock fallback
- **4-1.4.11 已完成：失败兜底与 Mock 回退策略**
  - 新增失败兜底链路：
    1. 业务 Agent 输出
    2. JSON parse
    3. Schema validate
    4. 如果失败，调用 JSONRepairAgent 一次
    5. repair 后重新 parse + validate
    6. 仍失败则 Mock fallback
  - 明确 JSONRepairAgent 不改变 generated_by，不创造业务事实
  - generated_by 只有 llm_agent 和 mock_agent 两种值
  - 不使用 llm_agent_fallback 作为 generated_by 值
  - fallback 状态由外层 agent_status / fallback_reason 表达
  - 策略文档见 `docs/05-Agent设计/失败兜底与Mock回退策略.md`
- **4-1.4.12 已完成：阶段 4-1.4 总验收**
  - 三个业务 Schema 均已登记且 JSON 合法：match_analysis_schema.json、recommendation_schema.json、resume_optimization_schema.json
  - 三个首轮输出验收均为 partial_pass，问题已被 Schema 吸收
  - 失败兜底链路已通过总验收
  - 验收报告见 `docs/05-Agent设计/阶段4-1.4总验收报告.md`
- **4-1.4.13 已完成：阶段 4-1.4 汇总**
  - 阶段 4-1.4 已正式闭环
  - 三个业务 Schema 均已定稿并登记：match_analysis_schema.json、recommendation_schema.json、resume_optimization_schema.json
  - 下一阶段需要实现 Schema Loader、JSON parse、Schema validate、JSON repair once
  - 汇总文档见 `docs/00-项目总览/阶段4-1.4Prompt与Agent真实接入汇总.md`
- 方案文档见 `docs/05-Agent设计/阶段4-1.4Prompt与Agent真实接入方案.md`
- **4-1.4.4A/B 已完成：岗位推荐 Agent 指令模板定稿与登记**
  - recommendation_v1.md 已包含推荐输出结构草案
  - **4-1.4.5 已完成：推荐 Prompt 首轮输出验收（partial_pass）**
    - 模型输出 JSON 合法，字段完整，证据链真实，无编造内容
    - 验收报告见 `docs/05-Agent设计/推荐Prompt首轮输出验收报告.md`
    - 验收结果将作为 recommendation_schema.json 的依据
  - **4-1.4.6A 已完成：人工定稿 recommendation_schema.json**
  - **4-1.4.6B 已完成：CodeBuddy 登记推荐 JSON Schema**
  - Schema 路径：`server/prompts/schema/recommendation_schema.json`
  - 该 Schema 约束 RecommendationAgent 的 LLM 输出
  - 当前 Schema 包含 7 个顶层 required 字段
  - 当前 Schema 禁止 additionalProperties
  - 当前 Schema 已固定四档推荐：
    - priority_apply / 优先投递
    - recommended / 建议尝试
    - explore_first / 先了解再决定
    - not_recommended / 暂不建议
  - 当前 Schema 已吸收 4-1.4.5 首轮验收问题：
    - 固定四档推荐，但不强制一次输出覆盖四档
    - recommend_level 与 recommend_label 一致（allOf if/then）
    - match_score 与 recommend_level 区间一致（allOf if/then）
    - 后续仍需补中匹配样本验证 recommended / explore_first

## 关联文档

- [阶段4Agent与模型接入方案.md](./阶段4Agent与模型接入方案.md)
- [LLMClient设计.md](./LLMClient设计.md)
- [Agent总览.md](./Agent总览.md)
- [Prompt模板.md](./Prompt模板.md)

## 阶段 4-3.0 更新

4-3.0 已设计 RunResult → Go Model 字段映射策略（docs only）。关键策略：

### RunResult.Parsed → Go Model 映射

Pipeline RunResult.Parsed 类型为 `map[string]any`，业务 Agent 需将其映射为 Go typed struct：

|| Agent | Parsed → Go Model | 映射方式 |
||-------|-------------------|---------|
|| MatchAnalysisAgent | map[string]any → model.MatchAnalysisResult | json re-marshal + unmarshal |
|| RecommendationAgent | map[string]any → []model.RecommendedJob | json re-marshal + unmarshal |
|| ResumeOptimizationAgent | map[string]any → 待定义 | 待实现 |

### 映射失败处理

- 映射失败（字段类型不匹配、缺失必要字段等）触发 Mock fallback
- 映射失败不应导致接口崩溃
- 映射失败记录日志，包含 error_code 和 fallback_reason

### generated_by / agent_status / repair_applied 字段策略

|| RunResult 字段 | 业务处理 | API 响应 |
||---------------|---------|---------|
|| generated_by | 立即使用（llm_agent / mock_agent） | 返回给前端 |
|| agent_status | 4-3 内部记录 | 4-4 统一收口到 API meta |
|| repair_applied | 4-3 内部记录 | 4-4 统一收口到 API meta |
|| fallback_reason | 4-3 内部记录 | 4-4 统一收口到 API meta |
|| error_code | 4-3 内部记录 | 4-4 统一收口到 API meta |

方案文档见 `docs/05-Agent设计/阶段4-3业务Agent接入LLM方案.md`

## 阶段 4-3.1 更新

4-3.1 已执行真实 LLM Smoke 前置验收，但真实 smoke 未执行（本地 LLM 配置不完整）。

因此未能观察真实 Hunyuan smoke 输出的 JSON parse 结果。

后续补执行真实 smoke 时需重点关注：
- json_parse_success 是否为 true
- 如果 json_parse_success=false，后续业务接入必须重点依赖 Schema validate / JSON repair / Mock fallback
- 模型最小 Prompt 下是否严格输出 JSON

验收报告见 `docs/05-Agent设计/真实LLMSmoke前置验收报告.md`。

## 阶段 4-3.1.B 更新

4-3.1.B 再次执行真实 LLM Smoke 核验，但真实 smoke 仍未执行（本地 LLM 配置仍不完整，与前次 4-3.1 一致）。

因此仍未能观察真实 Hunyuan smoke 输出的 JSON parse 结果。

后续补执行真实 smoke 时需重点关注：
- json_parse_success 是否为 true
- 如果 json_parse_success=false，后续业务接入必须重点依赖 Schema validate / JSON repair / Mock fallback
- 模型最小 Prompt 下是否严格输出 JSON

## 阶段 4-3.1 重新执行更新

4-3.1 重新执行仍未能观察真实 Hunyuan smoke 输出的 JSON parse 结果。

原因：LLM_ENABLE=false，LLM_BASE_URL 未配置。`.env` 已存在但只有模板值。

三次核验均因配置不完整无法执行真实 smoke。

## 阶段 4-3.1 通过结果登记更新

4-3.1 已观察真实 Hunyuan smoke 输出。

本次最小 Prompt 返回 JSON 可被成功解析，`json_parse_success=true`。`raw_output_preview={"ok":true,"message":"pong"}`。

后续业务输出仍必须经过 JSON extract / parse / Schema validate / repair once / fallback_required。业务 Prompt 输出复杂度远高于最小连通性测试，JSON 遵循性仍需重点关注。

4-3.3 已验证匹配分析接口在 JSON parse / Schema validate / repair once / fallback_required 下的稳定性。6 大验收场景全部通过，修复了 LLM 成功路径快照保存缺失的问题。

4-3.4 推荐链路已开始执行 JSON parse / Schema validate / repair once / fallback_required 的统一处理。推荐输出还必须校验岗位来自候选集（job_id 白名单过滤），LLM 返回不存在 job_id 的推荐项被丢弃，全部丢弃后触发 fallback。

## 阶段 4-3.5.B 更新

4-3.5.B 已完成战略差异化确认。后续 Schema 可评估增加以下字段：

1. `apply_readiness`：投递准备度（ready_to_apply / quick_fix_before_apply / not_ready_for_this_role / insufficient_information）
2. `optimization_stop_signal`：优化停止信号（stop_and_apply / fix_top_3_then_apply / needs_more_evidence / target_role_mismatch）
3. `self_check_question`：帮助用户确认建议是否适用的自查问题
4. `learn_more_query`：用户可进一步搜索学习的方向或关键词
5. `verification_notice`：提醒用户 AI 建议仅供参考，需自行核实
6. `quality_gate_result`：Quality Gate 检查结果

本阶段不修改 Schema，仅登记方向。

详见 `docs/05-Agent设计/战略差异化与杀手锏确认.md`

## 阶段 4-3.5.C 更新

4-3.5.C 已完成投递边界与优化停止条件设计。后续 Schema 可评估增加以下字段：

1. `apply_readiness`：投递准备度枚举（ready_to_apply / quick_fix_before_apply / not_ready_for_this_role / insufficient_information）
2. `optimization_stop_signal`：优化停止信号枚举（stop_and_apply / fix_top_3_then_apply / needs_more_evidence / target_role_mismatch）
3. `top_fix_items`：Top 3 快速优化项数组，每项含 issue / why_it_matters / action / example_reference / self_check_question / estimated_effort
4. `long_term_improvement`：长期提升项数组
5. `estimated_effort`：预计耗时枚举（within_10_min / within_30_min / within_60_min / long_term）
6. `screening_risks`：初筛风险
7. `recommendation_confidence`：推荐置信度

本阶段不修改 Schema。

详见 `docs/05-Agent设计/投递边界与优化停止条件设计.md`

## 阶段 4-3.5.D 更新

4-3.5.D 已完成 JD 证据链匹配与 START 简历优化规则设计。后续 Schema 可评估增加以下字段：

1. `jd_evidence_chain`：JD 证据链数组，每项含 jd_requirement / jd_requirement_source / resume_evidence / evidence_strength / match_judgement / gap_analysis / screening_risk / action_suggestion / example_reference / self_check_question / learn_more_query
2. `evidence_strength`：证据强度枚举（strong / medium / weak / missing）
3. `match_judgement`：匹配判断枚举（matched / partially_matched / missing / risky）
4. `screening_risk`：初筛风险类型（keyword_missing / responsibility_unclear / result_missing / tech_stack_not_supported / scope_unclear / experience_missing / role_mismatch / authenticity_risk / location_or_preference_risk）
5. `start_missing_parts`：START 缺失要素数组（Situation / Task / Action / Result / Tech）

本阶段不修改 Schema。

详见 `docs/05-Agent设计/JD证据链匹配与START简历优化规则.md`

## 阶段 4-3.5.E 更新

4-3.5.E 已完成 Agent 输出 Quality Gate 设计。后续 Schema 可评估增加以下字段：

1. `quality_gate_result`：Quality Gate 检查结果（pass / revise_required / fallback_required / manual_review_suggested）
2. `quality_score`：内容质量评分（内部使用）
3. `failed_checks`：未通过的检查项列表
4. `revision_allowed`：是否允许 revise
5. `revision_count`：已执行 revise 次数
6. `final_action`：最终动作（return / revise_once / conservative_fallback / mock_fallback）
7. `quality_gate_summary`：质检摘要（可选，前端展示用）

Quality Gate 是内容层检查，在 JSON Repair 和 Schema Validate 之上运行。结构合法不等于内容可靠。

本阶段不修改 Schema。

详见 `docs/05-Agent设计/Agent输出QualityGate设计.md`

## 阶段 4-3.5.F 更新

4-3.5.F 已完成学生求职极端情况处理边界设计。后续 Schema 可评估增加以下字段：

1. `candidate_case_type`：候选人情况类型枚举（normal_case / insufficient_information / weak_resume / target_role_mismatch / overseas_or_location_risk / high_potential_research_or_deep_tech / unclear_job_intention / authenticity_risk / over_optimization_risk / unsafe_user_request）
2. `case_risk_level`：情况风险等级（low / medium / high / critical）
3. `case_handling_strategy`：处理策略（normal_response / clarify_before_recommendation / quick_fix_then_apply / recommend_alternative_jobs / conservative_advice / refuse_fabrication / manual_review_suggested）
4. `recommendation_confidence`：推荐置信度（极端情况下应降低）
5. `authenticity_risk`：真实性风险标记
6. `over_optimization_risk`：过度优化风险标记
7. `location_or_preference_risk`：地点或偏好风险标记
8. `unsafe_user_request_detected`：检测到不安全用户请求

本阶段不修改 Schema。

详见 `docs/05-Agent设计/学生求职极端情况处理边界.md`

## 阶段 4-3.5.G 更新

4-3.5.G 已完成灵活在线简历 Profile 与纯文字 Agent 上下文设计。后续 Schema 可评估增加以下字段：

1. `resume_context_text`：输入的简历上下文摘要（用于审计和调试）
2. `profile_summary`：输入的简历概况信号（target_roles / expected_locations / section_count / has_project 等）
3. `section_count`：简历 Section 数量
4. `section_types`：简历包含的 Section 类型列表
5. `privacy_flags`：隐私标记（contact_info_provided / email_provided / phone_provided / name_excluded）
6. `data_quality`：数据质量评估（sufficient / minimal / insufficient）

本阶段不修改 Schema 文件，只做设计登记。

详见 `docs/05-Agent设计/灵活在线简历Profile与纯文字Agent上下文设计.md`

## 阶段 4-3.5.H 战略转型阶段汇总与工程评审

4-3.5.H 工程评审结论：暂不立即修改 Schema 文件。
后续可单独评估是否增加投递边界、证据链、Quality Gate、极端情况和 ResumeContext 相关字段。

详见 `docs/05-Agent设计/阶段4-3.5战略转型汇总与工程评审.md`

## 阶段 4-3.6 ResumeOptimizationAgent 现状核对

4-3.6 已核对 resume_optimization_schema.json 是否存在及是否已接入工程：
- `server/prompts/schema/resume_optimization_schema.json` ✅ 存在，15 个 required 字段
- Schema Loader 支持 `resume_optimization` 逻辑名称 ✅
- Pipeline 和 Repair 支持 `TargetResumeOptimization` ✅
- 但未被业务代码消费 ❌

本阶段不修改 Schema 文件。

详见 `docs/05-Agent设计/ResumeOptimizationAgent现状核对与接入方案.md`
