# 失败兜底与 Mock 回退策略

> 创建时间：2026-05-29
> 所属阶段：4-1.4.11
> 依据文档：阶段4-1.4方案、Agent总览、Prompt模板、JSON输出规范、LLMClient设计、Hunyuan接入准备、json_repair_v1.md、match_analysis_v1.md、recommendation_v1.md、resume_optimization_v1.md
> 最后更新：2026-05-30（阶段 4-3.5.B 战略差异化与杀手锏确认）

## 1. 设计目标

目标不是让 LLM 永远成功，而是让系统在 LLM 不稳定时仍然可演示、可返回、可解释、可追踪。

核心原则：

1. **Demo 稳定性优先**：用户始终能拿到业务结果，不会因为模型波动导致接口崩溃或前端无结果
2. **不返回坏 JSON**：任何不合 Schema 或非法 JSON 都不会直接返回给前端
3. **不把底层错误暴露给用户**：用户不会看到"模型报错""JSON 解析失败""Schema 校验失败"等技术错误
4. **不因为 Hunyuan 波动导致接口失败**：模型不可用时自动回退 Mock Agent
5. **不让 JSON repair 变成二次业务生成**：JSONRepairAgent 只修结构，不重做业务分析
6. **失败后可回退 Mock Agent**：Mock Agent 是最终兜底，保证接口结构稳定

## 2. 适用范围

本策略覆盖以下 Agent：

| Agent | 用途 | Mock 兜底 |
|-------|------|----------|
| MatchAnalysisAgent | 岗位匹配分析 | match_mock.go |
| RecommendationAgent | 岗位推荐 | recommendation_mock.go |
| ResumeOptimizationAgent | 简历优化 | 需后续实现确定性 Mock |
| JSONRepairAgent | JSON 结构修复 | 不直接面向用户，不直接作为业务结果来源 |

说明：JSONRepairAgent 只服务于结构修复，不直接面向用户，不直接作为业务结果来源。

## 3. 当前前提

- 当前阶段仍未接 Hunyuan
- 当前阶段仍未实现 LLM Client
- 当前阶段仍未替换 Mock Agent
- 当前阶段只做策略设计
- 真实工程实现放到后续 4-2 / 4-3 阶段

## 4. 总体链路

推荐链路：

```
业务请求
→ 组装业务上下文
→ 选择业务 Prompt
→ 选择对应 Output Schema
→ 判断 LLM 是否启用
  → 如果 LLM 未启用：直接 Mock Agent
→ 调用 LLM
→ 提取 raw output
→ JSON parse
  → 如果 parse 失败：调用 JSONRepairAgent 一次
    → repair output JSON parse
    → repair output Schema validate
    → 如果通过：返回 repaired LLM 结果
    → 如果仍失败：回退 Mock Agent
→ Schema validate
  → 如果 validate 失败：调用 JSONRepairAgent 一次
    → repair output JSON parse
    → repair output Schema validate
    → 如果通过：返回 repaired LLM 结果
    → 如果仍失败：回退 Mock Agent
→ 如果通过：返回 LLM 结果
→ 返回稳定结果
```

核心约束：

- **JSON repair 只尝试一次**
- **Repair 失败后不继续递归修复**
- **不允许无限重试**
- **LLM 调用本身不重试（JSON parse/validate 失败才进入 repair）**

## 5. 失败场景分类

| 错误码 | 触发条件 | 是否进入 JSON repair | 是否直接 Mock fallback | 是否记录日志 | 是否暴露给用户 |
|--------|---------|--------------------|-----------------------|------------|---------------|
| LLM_DISABLED | ENABLE_LLM_AGENT=false | ❌ | ✅ 直接 Mock | ✅ | ❌ |
| LLM_CONFIG_MISSING | LLM_BASE_URL / LLM_MODEL 为空 | ❌ | ✅ 直接 Mock | ✅ | ❌ |
| LLM_API_KEY_MISSING | LLM_API_KEY 为空 | ❌ | ✅ 直接 Mock | ✅ | ❌ |
| LLM_TIMEOUT | 请求超过 LLM_TIMEOUT_SECONDS | ❌ | ✅ 直接 Mock | ✅ | ❌ |
| LLM_HTTP_ERROR | HTTP 5xx / 网络错误 | ❌ | ✅ 直接 Mock | ✅ | ❌ |
| LLM_RATE_LIMIT | HTTP 429 / 配额超限 | ❌ | ✅ 直接 Mock | ✅ | ❌ |
| LLM_EMPTY_RESPONSE | LLM 输出为空字符串 | ❌ | ✅ 直接 Mock | ✅ | ❌ |
| LLM_CONTEXT_TOO_LONG | prompt_tokens 超出模型上限 | ❌ | ✅ 直接 Mock | ✅ | ❌ |
| LLM_OUTPUT_TRUNCATED | finish_reason=length，输出被截断 | ✅ 尝试 repair | 看 repair 结果 | ✅ | ❌ |
| JSON_PARSE_FAILED | json.Unmarshal 失败 | ✅ | 看 repair 结果 | ✅ | ❌ |
| JSON_MARKDOWN_WRAPPED | 输出被 ```json 包裹 | ✅ | 看 repair 结果 | ✅ | ❌ |
| SCHEMA_VALIDATION_FAILED | Schema 校验失败 | ✅ | 看 repair 结果 | ✅ | ❌ |
| SCHEMA_REQUIRED_FIELD_MISSING | 必填字段缺失 | ✅ | 看 repair 结果 | ✅ | ❌ |
| SCHEMA_ENUM_INVALID | 枚举值不合法 | ✅ | 看 repair 结果 | ✅ | ❌ |
| SCHEMA_TYPE_INVALID | 字段类型错误 | ✅ | 看 repair 结果 | ✅ | ❌ |
| SCHEMA_ADDITIONAL_PROPERTY | 存在 Schema 不允许的额外字段 | ✅ | 看 repair 结果 | ✅ | ❌ |
| JSON_REPAIR_FAILED | JSONRepairAgent 输出仍无法 parse | ❌ | ✅ 直接 Mock | ✅ | ❌ |
| JSON_REPAIR_SCHEMA_FAILED | JSONRepairAgent 输出不符合 Schema | ❌ | ✅ 直接 Mock | ✅ | ❌ |
| BUSINESS_SAFETY_RISK | 输出包含虚构经历、建议造假等 | ❌ | ✅ Mock 或安全降级 | ✅ | ❌ |
| UNKNOWN_ERROR | 未知异常 | ❌ | ✅ 直接 Mock | ✅ | ❌ |

## 6. 兜底决策表

| 失败阶段 | 错误码 | 触发条件 | 是否调用 JSONRepairAgent | 是否回退 Mock | 用户侧表现 | 日志记录 |
|---------|--------|---------|------------------------|-------------|-----------|---------|
| LLM 配置 | LLM_DISABLED | LLM 未启用 | ❌ | ✅ | 正常业务结果（Mock） | llm_enabled=false |
| LLM 配置 | LLM_CONFIG_MISSING | 配置缺失 | ❌ | ✅ | 正常业务结果（Mock） | 缺失配置项 |
| LLM 配置 | LLM_API_KEY_MISSING | API Key 缺失 | ❌ | ✅ | 正常业务结果（Mock） | api_key_missing=true |
| LLM 调用 | LLM_TIMEOUT | 超时 | ❌ | ✅ | 正常业务结果（Mock） | timeout=true |
| LLM 调用 | LLM_HTTP_ERROR | HTTP 错误 | ❌ | ✅ | 正常业务结果（Mock） | http_status=5xx |
| LLM 调用 | LLM_RATE_LIMIT | 限流 | ❌ | ✅ | 正常业务结果（Mock） | rate_limited=true |
| LLM 调用 | LLM_EMPTY_RESPONSE | 空响应 | ❌ | ✅ | 正常业务结果（Mock） | empty_response=true |
| LLM 调用 | LLM_CONTEXT_TOO_LONG | 上下文过长 | ❌ | ✅ | 正常业务结果（Mock） | context_too_long=true |
| LLM 调用 | LLM_OUTPUT_TRUNCATED | 输出截断 | ✅ | 看结果 | 正常业务结果（repair 或 Mock） | truncated=true |
| JSON 解析 | JSON_PARSE_FAILED | 解析失败 | ✅ | 看结果 | 正常业务结果（repair 或 Mock） | parse_failed=true |
| JSON 解析 | JSON_MARKDOWN_WRAPPED | Markdown 包裹 | ✅ | 看结果 | 正常业务结果（repair 或 Mock） | markdown_wrapped=true |
| Schema 校验 | SCHEMA_VALIDATION_FAILED | 校验失败 | ✅ | 看结果 | 正常业务结果（repair 或 Mock） | validation_errors=... |
| Schema 校验 | SCHEMA_REQUIRED_FIELD_MISSING | 必填字段缺失 | ✅ | 看结果 | 正常业务结果（repair 或 Mock） | missing_fields=... |
| Schema 校验 | SCHEMA_ENUM_INVALID | 枚举不合法 | ✅ | 看结果 | 正常业务结果（repair 或 Mock） | invalid_enums=... |
| Schema 校验 | SCHEMA_TYPE_INVALID | 类型错误 | ✅ | 看结果 | 正常业务结果（repair 或 Mock） | type_errors=... |
| Schema 校验 | SCHEMA_ADDITIONAL_PROPERTY | 额外字段 | ✅ | 看结果 | 正常业务结果（repair 或 Mock） | additional_props=... |
| JSON Repair | JSON_REPAIR_FAILED | repair 输出 parse 失败 | ❌ | ✅ | 正常业务结果（Mock） | repair_failed=true |
| JSON Repair | JSON_REPAIR_SCHEMA_FAILED | repair 输出 Schema 失败 | ❌ | ✅ | 正常业务结果（Mock） | repair_schema_failed=true |
| 业务安全 | BUSINESS_SAFETY_RISK | 虚构/造假内容 | ❌ | ✅ | 正常业务结果（Mock） | safety_risk=true |
| 未知 | UNKNOWN_ERROR | 未知异常 | ❌ | ✅ | 正常业务结果（Mock） | unknown_error=true |

核心决策总结：

- **LLM 未启用 / 配置缺失 / API Key 缺失**：直接 Mock
- **LLM timeout / HTTP error / rate limit**：直接 Mock
- **LLM 输出为空**：直接 Mock
- **LLM 输出被 Markdown 包裹**：先 repair
- **JSON parse failed**：先 repair 一次
- **Schema validation failed**：先 repair 一次
- **Repair parse failed**：Mock
- **Repair schema failed**：Mock
- **业务安全风险**：Mock 或安全降级
- **未知异常**：Mock

## 7. JSON Repair 策略

### 7.1 定位

JSONRepairAgent 只修复结构，不重新做业务分析。

### 7.2 核心规则

- JSONRepairAgent 不改变业务结果来源
- JSONRepairAgent 只执行一次
- JSONRepairAgent 输入包括：
  - `TARGET_OUTPUT_TYPE`：match_analysis / recommendation / resume_optimization
  - `TARGET_SCHEMA_JSON`：对应业务 Schema
  - `BROKEN_MODEL_OUTPUT`：原始异常输出
  - `VALIDATION_ERRORS`：解析/校验错误信息
  - `REPAIR_CONTEXT`：可选上下文（resume_id、job_id 等）
- JSONRepairAgent 输出仍必须符合对应业务 Schema

### 7.3 repair 成功

如果 repair 成功：

- 业务结果 `generated_by` 仍为 `llm_agent`
- 系统日志记录 `repair_applied=true`
- 可在内部 meta 中记录 `repaired=true`
- 不改变业务结果的来源标记

### 7.4 repair 失败

如果 repair 失败：

- 不再继续 repair
- 不允许递归修复
- 进入 Mock fallback
- 系统日志记录 `repair_failed=true`

### 7.5 repair 不做的事

- 不重新做业务分析
- 不编造岗位、经历、项目指标、系统规模、职责范围
- 不把 `generated_by` 改成 `json_repair_agent`
- 不创造原始输出和上下文中都没有的业务事实

## 8. Mock Agent 回退策略

### 8.1 Mock Agent 定位

Mock Agent 是最终兜底，不是低质量替代。

### 8.2 Mock fallback 基本原则

- Mock fallback 需要保证接口结构稳定
- Mock fallback 结果应尽量基于当前本地数据、已有规则和现有 Mock Agent
- Mock fallback 不请求外部接口
- Mock fallback 不调用 Hunyuan

### 8.3 各 Agent 的 Mock fallback

#### MatchAnalysisAgent fallback

- 使用现有 `match_mock.go` 逻辑
- 输入：简历版本 + 目标岗位 JD
- 输出：匹配分数 + 证据链 + 风险点 + P0/P1 建议 + 真实性提示
- 回退后 `generated_by = "mock_agent"`

#### RecommendationAgent fallback

- 使用现有 `recommendation_mock.go` 逻辑
- 输入：简历版本 + 候选岗位列表
- 输出：推荐岗位列表 + 推荐理由 + 风险点 + 优先级
- 回退后 `generated_by = "mock_agent"`

#### ResumeOptimizationAgent fallback

- 当前项目尚未实现 resume_optimization_mock.go
- 后续实现时需要设计确定性 Mock 输出
- 当前阶段仅策略预留
- 回退后 `generated_by = "mock_agent"`

## 9. generated_by 与 fallback 标记策略

### 9.1 generated_by 约定

| 场景 | generated_by 值 | 说明 |
|------|----------------|------|
| LLM 输出合法且通过 Schema 校验 | `llm_agent` | 正常 LLM 结果 |
| LLM 输出经 JSONRepairAgent 修复后通过 | `llm_agent` | repair 不改变业务来源标记 |
| Mock Agent 正常兜底（LLM 未启用或失败） | `mock_agent` | Mock 结果 |

### 9.2 不使用的值

以下值不应作为业务 Schema 内 `generated_by`：

- ❌ `llm_agent_fallback`：不建议使用。fallback 状态应由外层 `agent_status` / `fallback_reason` 表达
- ❌ `json_repair_agent`：JSONRepairAgent 不改变业务结果来源
- ❌ `repair_agent`：同上

### 9.3 历史文档处理

已有文档中出现的 `llm_agent_fallback`：

- 不建议作为业务 Schema 内 `generated_by`
- fallback 状态应由外层 `agent_status` / `fallback_reason` 表达
- Mock fallback 的业务结果仍用 `mock_agent`

### 9.4 后续 API meta 层建议

后续 API 响应中建议增加 meta 字段：

```json
{
  "agent_status": "success | repaired | fallback",
  "fallback_reason": "",
  "error_code": "",
  "repair_applied": false,
  "llm_enabled": true
}
```

注意：本阶段只写策略，不修改接口结构。如果当前接口模型暂时没有 meta 字段，只在文档中写"后续实现建议"。

## 10. error_code 设计

| error_code | 含义 | 触发阶段 | 处理方式 | 是否用户可见 |
|-----------|------|---------|---------|------------|
| LLM_DISABLED | LLM 未启用 | 配置检查 | 直接 Mock | ❌ |
| LLM_CONFIG_MISSING | LLM 配置缺失 | 配置检查 | 直接 Mock | ❌ |
| LLM_TIMEOUT | LLM 请求超时 | LLM 调用 | 直接 Mock | ❌ |
| LLM_HTTP_ERROR | LLM HTTP 错误 | LLM 调用 | 直接 Mock | ❌ |
| LLM_RATE_LIMIT | LLM 限流 | LLM 调用 | 直接 Mock | ❌ |
| LLM_EMPTY_RESPONSE | LLM 空响应 | LLM 调用 | 直接 Mock | ❌ |
| LLM_OUTPUT_TRUNCATED | LLM 输出截断 | LLM 调用 | 尝试 repair | ❌ |
| JSON_PARSE_FAILED | JSON 解析失败 | JSON 解析 | 尝试 repair | ❌ |
| SCHEMA_VALIDATION_FAILED | Schema 校验失败 | Schema 校验 | 尝试 repair | ❌ |
| JSON_REPAIR_FAILED | JSON repair 输出仍无法 parse | JSON repair | 直接 Mock | ❌ |
| MOCK_FALLBACK_USED | Mock fallback 已使用 | 兜底 | 正常返回 | ❌ |
| UNKNOWN_ERROR | 未知异常 | 任意 | 直接 Mock | ❌ |

用户可见策略：

- 默认不直接暴露技术错误码给普通用户
- 可以在调试模式或日志中展示
- Demo 用户侧只看到正常业务结果

## 11. 日志记录策略

### 11.1 日志字段

| 字段 | 说明 | 是否记录 |
|------|------|---------|
| request_id | 请求唯一 ID | ✅ |
| agent_name | 业务 Agent 名称 | ✅ |
| target_output_type | 目标输出类型 | ✅ |
| llm_enabled | LLM 是否启用 | ✅ |
| llm_model | 使用的模型名称 | ✅ |
| prompt_version | Prompt 版本 | ✅ |
| schema_version | Schema 版本 | ✅ |
| call_started_at | 调用开始时间 | ✅ |
| call_duration_ms | 调用耗时（毫秒） | ✅ |
| raw_output_length | 原始输出长度 | ✅ |
| json_parse_success | JSON 解析是否成功 | ✅ |
| schema_validate_success | Schema 校验是否成功 | ✅ |
| repair_attempted | 是否尝试 repair | ✅ |
| repair_success | repair 是否成功 | ✅ |
| fallback_used | 是否使用 fallback | ✅ |
| fallback_reason | fallback 原因 | ✅ |
| error_code | 错误码 | ✅ |

### 11.2 安全约束

- ❌ 不要记录 API Key
- ❌ 不要记录完整简历
- ❌ 不要记录完整 JD
- ❌ 不要记录完整模型输出
- ✅ 可以记录长度、摘要、错误类型

## 12. 用户侧展示策略

### 12.1 核心原则

- 用户不应看到"模型报错""JSON 解析失败""Schema 校验失败"等底层技术错误
- 用户应看到稳定的业务结果
- 如果是 fallback，可以在内部调试信息中标记，但 Demo 正常展示

### 12.2 展示策略

| 用户类型 | 展示内容 |
|---------|---------|
| 普通用户 | 不展示 fallback 细节，只展示正常业务结果 |
| 开发调试模式 | 展示 agent_status / fallback_reason / error_code |

### 12.3 前端展示建议

- 前端不根据 `generated_by` 展示不同 UI
- 前端始终按照统一业务结构渲染
- 调试模式下可在页面底部展示 agent_status 信息

## 13. 三类业务 Agent 的回退边界

### 13.1 MatchAnalysisAgent

```
优先使用 LLM + match_analysis_schema.json
→ 失败后 JSON repair 一次
→ 仍失败后使用 match_mock.go
```

回退后：
- `generated_by = "mock_agent"`
- 日志记录 `fallback_reason` 和 `error_code`
- 用户看到稳定匹配分析结果

### 13.2 RecommendationAgent

```
优先使用 LLM + recommendation_schema.json
→ 失败后 JSON repair 一次
→ 仍失败后使用 recommendation_mock.go
```

回退后：
- `generated_by = "mock_agent"`
- 日志记录 `fallback_reason` 和 `error_code`
- 用户看到稳定推荐结果

### 13.3 ResumeOptimizationAgent

```
优先使用 LLM + resume_optimization_schema.json
→ 失败后 JSON repair 一次
→ 仍失败后使用确定性 Mock 输出
```

当前项目可能尚未实现该 Mock，后续接入时需要补齐。

回退后：
- `generated_by = "mock_agent"`
- 日志记录 `fallback_reason` 和 `error_code`
- 用户看到稳定简历优化结果

## 14. JSONRepairAgent 边界

- JSONRepairAgent 不直接面向用户
- JSONRepairAgent 不生成业务结论
- JSONRepairAgent 不修改岗位匹配判断
- JSONRepairAgent 不重新推荐岗位
- JSONRepairAgent 不重新优化简历
- JSONRepairAgent 不改变 `generated_by`
- JSONRepairAgent 只在 JSON parse / Schema validate 失败后触发
- JSONRepairAgent 只执行一次，不递归修复
- JSONRepairAgent 不编造岗位、经历、项目指标、系统规模、职责范围

## 15. 后续实现建议

后续落地步骤：

1. 4-2 实现 LLM Client 抽象
2. 4-2 实现 Hunyuan Client
3. 加载 Prompt 文件（server/prompts/*.md）
4. 加载 Schema 文件（server/prompts/schema/*.json）
5. 业务 Agent 调用 LLM
6. JSON parse
7. Schema validate
8. JSON repair 一次（调用 JSONRepairAgent）
9. Mock fallback（调用对应 Mock Agent）
10. 记录日志和 agent_status

## 16. 验收标准

未来实现时的验收标准：

1. LLM 成功时返回 `generated_by = "llm_agent"`
2. LLM 返回 Markdown 包裹 JSON 时可修复
3. LLM 返回缺字段 JSON 时可修复或 fallback
4. JSON repair 成功时 `repair_applied = true`
5. JSON repair 失败时 `fallback_used = true`
6. Mock fallback 不崩溃
7. 用户侧始终有稳定业务结果
8. 日志不泄露 API Key、完整简历、完整 JD
9. `generated_by` 只有 `llm_agent` 和 `mock_agent` 两种值
10. JSON repair 不改变业务结果来源标记
11. LLM 配置缺失时直接 Mock，不报错
12. LLM 超时/HTTP 错误时直接 Mock，不报错

## 阶段 4-2 更新

4-2.0 已确定 fallback 策略在 LLM 基础设施层的落地边界，详细方案见 `docs/05-Agent设计/阶段4-2LLMClient与Hunyuan接入方案.md`。

关键落地边界：

1. **4-2 将先实现通用 fallback decision**：不直接调用具体业务 Mock Agent，只返回"应使用 Mock"
2. **4-3 再把具体业务 Mock Agent 接入业务链路**：MatchAnalysisAgent / RecommendationAgent / ResumeOptimizationAgent
3. **LLMRunner 统一调用链路**：Chat → JSON parse → Schema validate → JSON repair once → fallback decision
4. **新增错误码**：PROMPT_LOAD_FAILED / SCHEMA_LOAD_FAILED
5. **错误码触发阶段和处理方式已在方案中明确**

4-2.7 已落地 JSONRepairAgent 调用链路，4-2.8 实现统一 LLM Agent Pipeline。

新增：
- `server/internal/agent/pipeline/` 包：Pipeline Runner（Prompt Render → LLM Call → JSON extract/parse → Schema validate → repair once → fallback decision）
- AgentStatus 三种状态：success / repaired / fallback_required
- RunRequest / RunResult 完整结构
- 20+ Pipeline 错误码 + 13 个 FallbackReason 常量
- safeErr / redactSensitive API Key 脱敏
- 18 项单元测试全部通过
- Pipeline 只返回 fallback_required 决策，不调用具体 Mock Agent（4-3 实现）

## 阶段 4-2.5 更新

4-2.5 已实现 HunyuanClient 的基础错误码映射。LLM_TIMEOUT / LLM_HTTP_ERROR / LLM_RATE_LIMIT / LLM_EMPTY_RESPONSE 后续将进入 fallback decision。

HunyuanClient 错误映射：
- HTTP 429 → LLM_RATE_LIMIT → 直接 Mock fallback
- HTTP 5xx → LLM_HTTP_ERROR → 直接 Mock fallback
- HTTP 401/403 → LLM_API_KEY_MISSING → 直接 Mock fallback
- 空 body / 空 choices / 空 content → LLM_EMPTY_RESPONSE → 直接 Mock fallback
- context deadline / cancel → LLM_TIMEOUT → 直接 Mock fallback
- finish_reason=length → 不视为硬错误，返回 ChatResponse，Pipeline 判断是否 repair
- 非 JSON 响应 → LLM_HTTP_ERROR → 直接 Mock fallback

新增：
- `server/internal/llm/hunyuan_client.go`：HunyuanClient 错误码映射实现
- `server/internal/llm/hunyuan_client_test.go`：错误分类测试（429/500/401/空响应/非JSON/超时）

4-2.7 已落地 JSONRepairAgent 调用链路。JSON repair 只执行一次，repair 失败后不继续递归修复。Mock fallback 仍放到 4-2.8 / 4-3。

新增：
- `server/internal/agent/repair/` 包：JSONRepairAgent repair once 调用链路
- 7 个 repair 错误码：JSON_REPAIR_FAILED / JSON_REPAIR_PARSE_FAILED / JSON_REPAIR_SCHEMA_FAILED / JSON_REPAIR_LLM_FAILED / JSON_REPAIR_PROMPT_LOAD_FAILED / JSON_REPAIR_SCHEMA_LOAD_FAILED / JSON_REPAIR_TARGET_INVALID
- 14 项单元测试全部通过

## 阶段 4-2.6 更新

4-2.6 只验证模型连通性，不进入业务 fallback。真实 fallback decision 仍放到 4-2.7 / 4-2.8。

新增：
- `server/cmd/llm-smoke/main.go`：LLM smoke test Go 命令（使用最小 Prompt，不接业务链路）
- `scripts/dev/llm_smoke.sh`：LLM smoke test 脚本
- Makefile `llm-help` / `llm-check` / `llm-smoke` target

## 阶段 4-2.4 更新

4-2.4 已定义 LLM ErrorCode 和 LLMError，后续 fallback decision 将根据错误码判断是否直接 Mock fallback 或先尝试 JSON repair。

新增错误码与 fallback 判断：
- `LLM_DISABLED` / `LLM_CONFIG_MISSING` / `LLM_API_KEY_MISSING` / `LLM_TIMEOUT` / `LLM_HTTP_ERROR` / `LLM_RATE_LIMIT` / `LLM_EMPTY_RESPONSE` / `UNKNOWN_ERROR`：直接 Mock fallback
- `LLM_OUTPUT_TRUNCATED`：尝试 JSON repair
- `IsDirectMockFallback(err)` 函数：判断错误是否应直接 Mock fallback，不需要进入 repair

新增：
- `server/internal/llm/` 包：LLMError（结构化错误 + Unwrap）+ ErrorCode（9 个错误码）+ IsErrorCode + IsDirectMockFallback

## 阶段 4-2.3 更新

4-2.3 已落地 JSON parse / Schema validate 工具。parse 或 validate 失败后，后续将进入 JSON repair once，再失败则 Mock fallback。

新增工具：
- `server/internal/llmjson/`：ExtractObject / ParseObject / ExtractAndParse，支持 4 种错误码
- `server/internal/schema/validator.go`：Validator.Validate / ValidateJSON，返回结构化 ValidationResult

## 阶段 4-2.1 更新

4-2.1 已新增 `LLM_FALLBACK_ENABLE=true` 配置，当 LLM 关闭、配置缺失、调用失败、repair 失败时，后续统一进入 Mock fallback。

详细配置见 `docs/06-工程部署/环境变量.md`。

## 17. 当前不做事项

- 本阶段不写 Go 代码
- 本阶段不实现 LLM Client
- 本阶段不调用 Hunyuan
- 本阶段不修改 Prompt
- 本阶段不修改 Schema
- 本阶段不修改接口结构
- 本阶段不修改前端
- 本阶段不做真实 fallback 验收

## 18. 阶段 4-1.4 总验收记录

4-1.4.12 阶段总验收已确认：

- 4 个 Prompt 文件全部存在
- 3 个 Schema 文件全部存在且 JSON 合法
- 3 个首轮输出验收均为 partial_pass，问题已被 Schema 吸收
- 失败兜底与 Mock 回退策略已完整
- 文档已同步
- 工程边界未被破坏

验收报告见 `docs/05-Agent设计/阶段4-1.4总验收报告.md`

## 19. 阶段 4-1.4 汇总记录

4-1.4.13 阶段汇总已确认：

- 阶段 4-1.4 已正式闭环
- 失败兜底策略已纳入阶段闭环
- 后续 4-2 实现时应按本文档落地 JSON repair once + Mock fallback
- 汇总文档见 `docs/00-项目总览/阶段4-1.4Prompt与Agent真实接入汇总.md`

## 阶段 4-3.0 更新

4-3.0 已完成 fallback 业务落地方案设计（docs only）。关键落地决策：

### fallback_required 与 Mock Agent 的结合

1. **Pipeline 只返回 fallback_required 决策**：不调用具体 Mock Agent
2. **业务 Agent Service 层处理 fallback**：检查 RunResult.FallbackRequired，若为 true 则调用对应 Mock Agent
3. **MatchAnalysisAgent fallback**：调用 `s.mockAgent.Analyze(resume, job)` 获取 model.MatchAnalysisResult
4. **RecommendationAgent fallback**：调用 `s.mockAgent.Recommend(resume, jobs)` 获取 []model.RecommendedJob
5. **ResumeOptimizationAgent fallback**：当前无 Mock Agent，4-3.6 评估是否需要补齐

### generated_by 设置规则（4-3 落地版）

```
if !runResult.FallbackRequired {
    result.GeneratedBy = "llm_agent"  // success or repaired
} else {
    mockResult := s.mockAgent.Xxx(...)
    mockResult.GeneratedBy = "mock_agent"
    return mockResult
}
```

### 字段映射失败与 Mock fallback

- RunResult.Parsed → Go Model 映射失败时，也应触发 Mock fallback
- 映射失败与 LLM 错误统一走 fallback_required 逻辑
- 日志记录映射失败的 error_code 和 fallback_reason

方案文档见 `docs/05-Agent设计/阶段4-3业务Agent接入LLM方案.md`

## 阶段 4-3.1 更新

4-3.1 已完成真实 LLM smoke 前置验收，但真实 smoke 未执行（本地 LLM 配置不完整）。

结论：
- 真实 LLM 连通性未验证
- 后续业务 Agent 接入时仍必须保留 Mock fallback
- Mock fallback 当前仍为接口稳定性的最终保障
- 暂不建议进入 4-3.2

验收报告见 `docs/05-Agent设计/真实LLMSmoke前置验收报告.md`。

## 阶段 4-3.1.B 更新

4-3.1.B 再次完成真实 LLM smoke 二次核验，但真实 smoke 仍未执行（本地 LLM 配置仍不完整，与前次 4-3.1 一致）。

结论：
- 真实 LLM 连通性仍未验证
- 两次核验均因配置缺失无法执行
- 后续业务 Agent 接入时仍必须保留 Mock fallback
- Mock fallback 当前仍为接口稳定性的最终保障
- 暂不建议进入 4-3.2

## 阶段 4-3.1 重新执行更新

4-3.1 重新执行仍未能验证真实 LLM 连通性。

原因：LLM_ENABLE=false，LLM_BASE_URL 未配置。`.env` 已存在（4-3.1.C 自动创建）但只有模板值。

三次核验均因配置不完整无法执行。后续业务 Agent 接入时仍必须保留 Mock fallback。

## 阶段 4-3.1 通过结果登记更新

4-3.1 已通过真实 LLM Smoke 前置验收。Hunyuan / TokenHub 最小调用可用，`json_parse_success=true`。

即使连通性通过，后续业务 Agent 接入仍必须保留 Mock fallback，避免业务 Prompt 输出异常导致 Demo 不稳定。

4-3.3 已在匹配分析接口层验收 Mock fallback。fallback_required 不作为接口失败，而是触发 match_mock.go 稳定兜底。6 大验收场景全部通过：LLM success / repaired / fallback_required / LLM_ENABLE=false / 配置异常 / 请求参数异常。

4-3.4 已在 RecommendationAgent 中落地 Mock fallback：Pipeline fallback_required 时回退 recommendation_mock.go。RecFallbackError 标识推荐链路的 fallback 触发原因。推荐结果中无效 job_id 被过滤后也可能触发 fallback。

## 阶段 4-3.5.B 更新

4-3.5.B 已完成战略差异化确认。后续 Quality Gate 可作为内容质量层面的兜底策略：

1. 如果 Agent 输出泛泛、无证据、编造经历或缺少投递边界，可重试一次或返回保守结果
2. Quality Gate 检查项：是否基于真实 JD、是否基于简历证据、是否编造经历、是否编造岗位、是否给出投递边界、是否给出优化停止条件、是否使用示例参考、是否提醒用户核实、是否包含自查问题、不评价敏感属性、符合 Schema、前端可展示
3. 最多重试一次，仍不合格则输出保守结果或 fallback

本阶段只做设计确认，不实现。

详见 `docs/05-Agent设计/战略差异化与杀手锏确认.md`

## 阶段 4-3.5.C 更新

4-3.5.C 已完成投递边界与优化停止条件设计。后续 Quality Gate 可检查 Agent 输出是否缺少投递边界或优化停止条件：

1. 如果输出只给泛泛优化建议但没有 apply_readiness，可触发重试或保守结果
2. 如果 apply_readiness=quick_fix_before_apply 但没有 top_fix_items，可触发补充
3. 如果 apply_readiness 与 optimization_stop_signal 矛盾（如 ready_to_apply + needs_more_evidence），可触发修正

本阶段只做设计确认，不实现。

详见 `docs/05-Agent设计/投递边界与优化停止条件设计.md`

## 阶段 4-3.5.D 更新

4-3.5.D 已完成 JD 证据链匹配与 START 简历优化规则设计。后续 Quality Gate 可检查 Agent 输出是否缺少 JD 证据链、是否泛泛建议、是否编造经历、是否缺少核实提醒：

1. 如果输出建议没有绑定 JD 要求，可触发重试或保守结果
2. 如果输出使用"建议改为"而非"示例参考"，可触发修正
3. 如果输出编造了用户没有的指标或经历，可触发 fallback
4. 如果 evidence_strength=missing 但输出了强匹配结论，可触发修正

本阶段只做设计确认，不实现。

详见 `docs/05-Agent设计/JD证据链匹配与START简历优化规则.md`

## 阶段 4-3.5.E 更新

4-3.5.E 已完成 Agent 输出 Quality Gate 设计。Quality Gate 是内容层面的兜底策略：

1. **Quality Gate 在 JSON Repair 之上运行**：JSON Repair 修结构，Quality Gate 检内容
2. **revise once**：Quality Gate 判定 revise_required 时，最多重试一次，仍不合格则 fallback
3. **conservative_fallback**：LLM 输出有一定价值但缺少足够证据时，返回保守结果
4. **mock_fallback**：LLM 输出存在严重问题（编造经历、编造岗位、不当承诺等）时，使用 Mock Agent
5. **fallback 不是失败**：Quality Gate 触发 fallback 是正确决策，不是系统故障
6. **fabrication_check 一旦失败**：应直接 fallback_required，不尝试 revise

后续 Pipeline 可在 JSON Repair 之后接入 Quality Gate：

```
Prompt → LLM → JSON parse → Schema validate → repair once → Quality Gate → revise once → fallback decision
```

本阶段只做设计，不实现。

详见 `docs/05-Agent设计/Agent输出QualityGate设计.md`

## 阶段 4-3.5.F 更新

4-3.5.F 已完成学生求职极端情况处理边界设计。极端情况与 fallback 策略的结合：

1. **unsafe_user_request**：用户要求编造经历或夸大成果，必须拒绝并返回替代建议，不 fallback 到 Mock，而是直接返回拒绝响应
2. **authenticity_risk**：如果 Agent 忽略真实性风险继续输出编造内容，Quality Gate fabrication_check 应直接 fallback_required
3. **insufficient_information**：如果 Agent 忽略信息不足强行输出 ready_to_apply，Quality Gate apply_boundary_check 应判定不合格
4. **target_role_mismatch**：如果 Agent 把不匹配包装成高匹配，Quality Gate jd_grounding_check 应判定不合格
5. **over_optimization_risk**：如果 apply_readiness=ready_to_apply 但 Agent 继续输出大量优化项，Quality Gate consistency_check 应判定不合格
6. **extreme_case_check**：后续 Quality Gate 可增加极端情况检查，如果 candidate_case_type != normal_case 但 quality_gate_result=pass，应确认 Agent 是否正确处理了极端情况

本阶段只做设计，不实现。

详见 `docs/05-Agent设计/学生求职极端情况处理边界.md`

## 阶段 4-3.5.G 更新

4-3.5.G 已完成灵活在线简历 Profile 与纯文字 Agent 上下文设计。ResumeContextText 信息完整性影响 fallback 决策：

1. **信息不足时的 fallback**：如果 ResumeContextText 缺少必填字段（target_roles / expected_locations / education_summary / 能力证明 Section），Agent 应输出 `apply_readiness = insufficient_information`，而不是编造内容
2. **profile_summary.data_quality = insufficient** 时，Agent 输出质量更不可控，应降低 quality_gate_result 通过阈值
3. **隐私字段异常**：如果 Agent 输出中出现了姓名/手机/邮箱等应被排除的字段，Quality Gate 应判定为 fallback_required
4. **Section 内容缺失**：如果 profile_summary 显示无项目经历但 Agent 输出中引用了项目经历，应判定为编造，触发 fallback

详见 `docs/05-Agent设计/灵活在线简历Profile与纯文字Agent上下文设计.md`

详见 `docs/05-Agent设计/学生求职极端情况处理边界.md`

## 阶段 4-3.5.H 战略转型阶段汇总与工程评审

4-3.5.H 工程评审结论：暂不立即扩展 fallback 工程逻辑。
后续如实现 Quality Gate，可再评估 conservative_fallback 和 mock_fallback 的内容质量分支。

详见 `docs/05-Agent设计/阶段4-3.5战略转型汇总与工程评审.md`
