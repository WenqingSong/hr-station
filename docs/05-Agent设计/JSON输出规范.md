# JSON 输出规范

> 最后更新：2026-05-29（阶段 4-1.4.4B 登记岗位推荐 Agent 指令模板）

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
