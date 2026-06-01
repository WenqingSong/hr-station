# 真实 LLM 验证与上线准入现状核验

> 阶段：11-0 | 状态：pass | 日期：2026-06-01

---

## 一、本阶段性质

**只做 read-only 核验，不改代码。** 查清 LLM 基础设施、三类 Agent 链路、Key 安全边界，为进入 11-1 提供判断依据。

---

## 二、LLM Client 结构

### 接口定义
`server/internal/llm/client.go`：`Client` 接口，单一 `Chat(ctx, ChatRequest) -> ChatResponse`。

### Provider 类型
| Provider | 文件 | 说明 |
|------|------|------|
| `hunyuan` | `hunyuan_client.go` | 腾讯混元，OpenAI-compatible API |
| `noop` | `noop_client.go` | 测试用，不访问外部网络 |

### 配置
`server/internal/llm/config.go`：
- `Enable` / `Provider` / `BaseURL` / `APIKey` / `Model` / `Temperature` / `MaxOutputTokens`
- `TimeoutSeconds` (默认 30s)
- `RepairEnable` / `RepairMaxAttempts` (最多 1 次)
- `FallbackEnable` (默认 true)
- `PromptDir` / `SchemaDir`
- APIKey: `json:"-"` tag — 不会序列化
- `Validate()`: Enable=true 时强制要求 Provider/BaseURL/APIKey/Model

### 错误类型
`server/internal/llm/errors.go`：`LLMError` 包含 `Code` 枚举（LLM_DISABLED、LLM_TIMEOUT、LLM_API_KEY_MISSING 等）

---

## 三、Provider / Env 配置现状

| 变量 | 默认值 | 来源 | 说明 |
|------|------|------|------|
| `LLM_ENABLE` | `false` | `.env` | 默认禁用 LLM |
| `LLM_PROVIDER` | `hunyuan` | `.env` | Provider 选择 |
| `LLM_BASE_URL` | `""` | `.env` | Hunyuan/TokenHub 地址 |
| `LLM_API_KEY` | `""` | `.env` | ⚠️ 必须为真实 Key 才启用 LLM |
| `LLM_MODEL` | `""` | `.env` | hy3-preview 等 |
| `LLM_TEMPERATURE` | `0.2` | `.env` | |
| `LLM_MAX_OUTPUT_TOKENS` | `8192` | `.env` | |
| `LLM_TIMEOUT_SECONDS` | `30` | `.env` | |
| `LLM_REPAIR_ENABLE` | `true` | `.env` | JSON 修复开关 |
| `LLM_FALLBACK_ENABLE` | `true` | `.env` | Mock fallback 开关 |
| `LLM_PROMPT_DIR` | `server/prompts` | `.env` | |
| `LLM_SCHEMA_DIR` | `server/prompts/schema` | `.env` | |

`.env.example` 已声明全部变量，`LLM_API_KEY=sk-your-real-api-key` 为安全占位符。

---

## 四、API Key 安全核验

| 检查项 | 结果 |
|------|:--:|
| `.env` 已被 gitignore | ✅ |
| `.env.example` Key 为占位符 | ✅ `sk-your-real-api-key` |
| config.APIKey `json:"-"` 不序列化 | ✅ |
| pipeline `safeErr()` 遮蔽 `key=` / `sk-` / `api_key=` | ✅ |
| llm-smoke `maskKey()` 遮蔽输出 | ✅ |
| docs 中无真实 Key | ✅ |
| 三页面前端无 Authorization/Key | ✅ |

---

## 五、Prompt / Schema 现状

| 文件 | 用途 |
|------|------|
| `server/prompts/recommendation_v1.md` | 岗位推荐 Prompt |
| `server/prompts/match_analysis_v1.md` | 岗位匹配分析 Prompt |
| `server/prompts/resume_optimization_v1.md` | 简历优化 Prompt |
| `server/prompts/json_repair_v1.md` | JSON 修复 Prompt |
| `server/prompts/schema/recommendation_schema.json` | 推荐输出 Schema |
| `server/prompts/schema/match_analysis_schema.json` | 匹配分析输出 Schema |
| `server/prompts/schema/resume_optimization_schema.json` | 优化输出 Schema |

`prompt.Loader` 负责模板渲染 + Schema 文件加载。`schema.Validator` 基于 JSON Schema 做输出校验。

---

## 六、JSON Parse / Schema Validate 现状

| 环节 | 文件 | 说明 |
|------|------|------|
| JSON 提取+解析 | `llmjson/extractor.go` + `parser.go` | 从 LLM 原始输出中提取 JSON |
| Schema 校验 | `schema/validator.go` | 基于 JSON Schema 校验 |
| JSON Repair | `agent/repair/json_repair_agent.go` | 失败时调 LLM 修复一次 |

Pipeline 流程：
```
Prompt Render → LLM Call → JSON Extract/Parse → Schema Validate
  → 失败 → JSONRepairAgent repair once
  → 修复输出再 Parse + Validate
  → 仍失败 → fallback_required
```

---

## 七、Timeout / Fallback 现状

| 机制 | 说明 |
|------|------|
| LLM 超时 | `TimeoutSeconds: 30`，HunyuanClient 内 `context.WithTimeout` |
| LLM 禁用 | `LLM_ENABLE=false` → 直接 Mock Agent |
| LLM 失败 | LLM error → `handleLLMError()` → `fallback_required` → Mock Agent |
| JSON Parse 失败 | → JSONRepairAgent repair once |
| Repair 失败 | → `fallback_required` → Mock Agent |
| Mock 兜底 | 三类 Service 均有 `recommendWithMock()` / `analyzeWithMock()` / `optimizeWithMock()` |

---

## 八、三类 Agent 链路

### 推荐 Agent
| 层级 | 文件 | 说明 |
|------|------|------|
| Handler | `http/recommendation_handler.go` | POST /api/jobs/recommend |
| Service | `service/recommendation_service.go` | LLM→Mock fallback |
| LLM Agent | `agent/recommendation_agent.go` | 调用 Pipeline Runner |
| Mock Agent | `agent/recommendation_mock.go` | 规则打分兜底 |
| Pipeline | `pipeline/pipeline.go` | Prompt→LLM→Parse→Validate→Repair |
| Prompt | `prompts/recommendation_v1.md` | |
| Schema | `prompts/schema/recommendation_schema.json` | |

### 匹配 Agent
| 层级 | 文件 | 说明 |
|------|------|------|
| Handler | `http/match_handler.go` | POST /api/jobs/{id}/match |
| Service | `service/match_service.go` | LLM→Mock fallback |
| LLM Agent | `agent/match_agent.go` | 调用 Pipeline Runner |
| Mock Agent | `agent/match_mock.go` | 规则打分兜底 |
| Pipeline | `pipeline/pipeline.go` | Prompt→LLM→Parse→Validate→Repair |
| Prompt | `prompts/match_analysis_v1.md` | |
| Schema | `prompts/schema/match_analysis_schema.json` | |

### 优化 Agent
| 层级 | 文件 | 说明 |
|------|------|------|
| Handler | `http/resume_optimization_handler.go` | POST /api/resumes/optimize |
| Service | `service/resume_optimization_service.go` | LLM→Mock fallback |
| LLM Agent | `agent/resume_optimization_agent.go` | 调用 Pipeline Runner |
| Mock Agent | `agent/resume_optimization_mock.go` | 规则评分兜底 |
| Pipeline | `pipeline/pipeline.go` | Prompt→LLM→Parse→Validate→Repair |
| Prompt | `prompts/resume_optimization_v1.md` | |
| Schema | `prompts/schema/resume_optimization_schema.json` | |

### 三类 Agent 一致性

```
✅ 三类 Agent 共享同一 Pipeline Runner
✅ 三类 Agent 共享同一 LLM Client (HunyuanClient)
✅ 三类 Agent 共享 LLM→Mock fallback 模式
✅ 三类 Agent 均返回 data.result + data.meta
✅ meta 包含 agent_name / generated_by / agent_status / fallback_reason / error_code / repair_applied / schema_valid / request_id
```

---

## 九、前端 Agent 展示边界

| 检查项 | 结果 |
|------|:--:|
| `agentAdapter.ts` `toAgentMetaViewModel()` 产品化 meta | ✅ |
| `common.ts` `formatAgentSourceLabel()` 基于 generated_by + agent_status | ✅ |
| agents.ts `withMockFallback` 兜底 | ✅（仅异常兜底） |
| 三页面前端不暴露 `agent_status` / `fallback_reason` 原始值 | ✅ |
| 三页面前端不暴露 API error / 401 / Unauthorized | ✅ |
| authStorage / client.ts 401 一致性处理 | ✅ |
| 三页面不手写 token | ✅ |

---

## 十、main.go Wiring

`server/cmd/server/main.go` 的 LLM 初始化逻辑：

```go
if llmEnabled {
    llmCfg := llm.Config{...}  // from cfg.LLM
    llmClient := llm.NewHunyuanClient(llmCfg, nil)
    pipelineRunner = pipeline.NewRunner(llmClient, promptLoader, schemaLoader, validator, repairAgent)
}
// 三类 Service 均传入 pipelineRunner + llmEnabled
matchService := NewMatchService(..., pipelineRunner, llmEnabled)
recommendationService := NewRecommendationService(..., pipelineRunner, llmEnabled)
resumeOptimizationService := NewResumeOptimizationService(..., pipelineRunner, llmEnabled)
```

---

## 十一、P0 / P1 / P2

| 等级 | 数量 | 内容 |
|:--:|:--:|------|
| P0 | **0** | — |
| P1 | **0** | — |
| P2 | **4** | ① 真实 Key 未配置（`LLM_ENABLE=false`） ② 真实 LLM 未调用过 ③ 错误 Key / 超时 / 非 JSON 未实测 ④ 上线前日志观测字段可增强 |

---

## 十二、结论

**pass** ✅ — LLM 基础设施、三类 Agent 链路、Key 安全边界、前端展示边界已全部摸清。

可以进入 **11-1 Key / Provider 配置核验与真实 LLM 首调验证**。

### 建议 11-1 执行内容

1. 确认 `LLM_API_KEY` 来源（TokenHub / Hunyuan）
2. 配置 `.env` 中 `LLM_ENABLE=true` + 真实 Key
3. `make llm-smoke` 验证连通性
4. 对三类 Agent 各做一次真实 LLM 调用
5. 验证 `generated_by=llm_agent` `agent_status=success`
6. 验证前端 meta 展示
