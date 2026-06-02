# K5-0：真实 LLM 接入前置准入核验

## 1. 性质

**read-only audit。** 未修改任何业务代码，未调用真实 LLM，未配置真实 API Key。

## 2. REAL_JOB_ID

`55`（软件开发-后台开发方向）

## 3. REAL_RESUME_VERSION_NO

`3`（v3 代表版本）

## 4. 当前 LLM_ENABLE 状态

**LLM_ENABLE = false**，默认关闭。

| 来源 | 值 |
|------|-----|
| `DefaultConfig()` | `Enable: false` |
| `.env.example` | `LLM_ENABLE=false` |
| 服务启动 | `LLM_ENABLE=false` → `llmEnabled=false` → 仅 Mock Agent |

真实 LLM 接入时只需设 `LLM_ENABLE=true` + 有效 `LLM_BASE_URL` + `LLM_API_KEY`，零代码变更。

## 5. Provider / BaseURL / Model 配置

| 配置项 | 值 | 来源 |
|--------|-----|------|
| LLM_PROVIDER | `hunyuan` | `.env.example` |
| LLM_BASE_URL | `https://tokenhub.tencentmaas.com/v1/chat/completions` | `.env.example` |
| LLM_MODEL | `hy3-preview` | `.env.example` |
| LLM_TEMPERATURE | `0.2` | `.env.example` |
| LLM_MAX_OUTPUT_TOKENS | `8192` | `.env.example` |
| LLM_TIMEOUT_SECONDS | `30` | `.env.example` |
| LLM_REPAIR_ENABLE | `true` | `.env.example` |
| LLM_REPAIR_MAX_ATTEMPTS | `1` | `.env.example` |
| LLM_FALLBACK_ENABLE | `true` | `.env.example` |
| LLM_LOG_RAW_OUTPUT | `false` | `.env.example` |
| LLM_PROMPT_DIR | `server/prompts` | `.env.example` |
| LLM_SCHEMA_DIR | `server/prompts/schema` | `.env.example` |

## 6. API Key 安全状态

| 检查项 | 结果 |
|--------|------|
| `.env` 是否被 git 追踪 | 否（仅 `.env.example` 被追踪） |
| `.env.example` 中的 Key | 占位符 `sk-your-real-api-key` |
| 真实 Key 扫描 (`sk-` / `AKIA`) | clean — 仅 docs 历史引用 |
| `Config.APIKey` 的 JSON tag | `json:"-"` ✅ |
| `MaskedAPIKey()` 日志脱敏 | ✅ 保留最后 4 位 |
| `safeErr()` + `redactSensitive()` | ✅ 管道层脱敏 |
| `LLMError` 不包含 Key / Prompt / 简历 | ✅ 原则已固化 |

**结论：无真实 Key 泄漏风险。**

## 7. Client 架构

| 组件 | 文件 | 状态 |
|------|------|------|
| `Client` 接口 | `server/internal/llm/client.go` | ✅ `Chat(ctx, ChatRequest) (*ChatResponse, error)` + `Provider()` |
| `HunyuanClient` | `server/internal/llm/hunyuan_client.go` | ✅ OpenAI-compatible，含 timeout、HTTP status 分类、error wrapping |
| `NoopClient` | `server/internal/llm/noop_client.go` | ✅ 不访问网络，用于测试 |
| `Config` | `server/internal/llm/config.go` | ✅ 15 字段，`Enable`/`Provider`/`BaseURL`/`APIKey`/`Model`/`Temperature`/`MaxOutputTokens`/`TimeoutSeconds`/`RepairEnable`/`RepairMaxAttempts`/`FallbackEnable`/`LogRawOutput`/`PromptDir`/`SchemaDir` |
| `LLMError` | `server/internal/llm/errors.go` | ✅ 9 种 ErrorCode + `IsDirectMockFallback()` |
| Timeout 控制 | `context.WithTimeout` | ✅ 30s 默认 |
| APIKey 安全 | `json:"-"` + `MaskedAPIKey()` | ✅ |

## 8. Agent Pipeline

**共享 Pipeline Runner**（`server/internal/agent/pipeline/pipeline.go`），三类 Agent 共用同一套流程：

```
Prompt 渲染 → LLM Chat → JSON 提取/解析 → Schema 校验
  → 失败：JSONRepairAgent 修复一次 → 修复后 parse + validate
  → 仍失败：return fallback_required → 调用方回退到 Mock Agent
```

| Agent | 文件 | LLM Agent | Mock Fallback | 共用 Pipeline |
|-------|------|-----------|---------------|--------------|
| 岗位匹配分析 | `match_agent.go` | ✅ `MatchAnalysisAgent` | ✅ `MatchMockAgent` | ✅ |
| 岗位推荐 | `recommendation_agent.go` | ✅ `RecommendationAgent` | ✅ `RecommendationMockAgent` | ✅ |
| 简历优化 | `resume_optimization_agent.go` | ✅ `ResumeOptimizationAgent` | ✅ `ResumeOptimizationMockAgent` | ✅ |

### Pipeline 流程详解

| 步骤 | 说明 |
|------|------|
| 1. 请求验证 | `agent_name`/`prompt_name`/`schema_name`/`target_output_type` 必填 |
| 2. Prompt 渲染 | `prompt.Loader.Render()` → 失败则 fallback |
| 3. Schema 加载 | `schema.Loader.Load()` → 失败则 fallback |
| 4. LLM 调用 | `llmClient.Chat()` → 失败则 fallback |
| 5. JSON 提取+解析 | `llmjson.ExtractAndParse()` → 失败则 repair once |
| 6. Schema 校验 | `validator.Validate()` → 失败则 repair once |
| 7. JSON Repair | `repairAgent.Repair()` → 仅一次 → 失败则 fallback |
| 8. 返回结果 | AgentStatus: success / repaired / fallback_required |

### Fallback 路径

- LLM 错误 → `handleLLMError()` → `MatchFallbackError` / `RecFallbackError` / `ResumeOptFallbackError`
- Service 层捕获 FallbackError → 调用对应 Mock Agent → 前端不受影响
- `LLM_ENABLE=false` 时，`llmEnabled=false`，Service 层不创建 LLM Agent，直接走 Mock

## 9. Prompt / Schema 状态

| 类型 | 文件 | 关联 Agent |
|------|------|-----------|
| Prompt | `server/prompts/match_analysis_v1.md` | MatchAnalysisAgent |
| Prompt | `server/prompts/recommendation_v1.md` | RecommendationAgent |
| Prompt | `server/prompts/resume_optimization_v1.md` | ResumeOptimizationAgent |
| Prompt | `server/prompts/json_repair_v1.md` | JSONRepairAgent |
| Schema | `server/prompts/schema/match_analysis_schema.json` | MatchAnalysisAgent |
| Schema | `server/prompts/schema/recommendation_schema.json` | RecommendationAgent |
| Schema | `server/prompts/schema/resume_optimization_schema.json` | ResumeOptimizationAgent |

**结论：4 个 Prompt + 3 个 Schema 全部存在，三类 Agent 各有专属 Prompt 和 Schema。**

## 10. fallback / repair / schema validate 状态

| 能力 | 实现 | 状态 |
|------|------|------|
| Schema validate | `schema.Validator` — JSON Schema 校验 | ✅ |
| JSON repair once | `repair.Agent` — RepairMaxAttempts=1，不允许递归修复 | ✅ |
| Mock fallback | 每类 Agent 各有 Mock Agent | ✅ |
| Fallback enabled | `FallbackEnable: true` | ✅ |
| Fallback 路径闭环 | LLM error → FallbackError → Mock Agent → 前端正常展示 | ✅ |
| 日志脱敏 | `safeErr()` + `redactSensitive()` + `MaskedAPIKey()` | ✅ |
| 空结果处理 | Chat 返回空 → `ErrLLMEmptyResponse` → fallback | ✅ |

## 11. 是否调用真实 LLM

**否。** 本阶段未调用真实 LLM。

- `LLM_ENABLE=false`（默认）
- 未配置真实 `LLM_API_KEY`
- 未执行 `make llm-smoke`
- 服务运行在 Mock Agent 模式下

## 12. go test / grep 结果

### go test

```
ok  offer-catcher/server/internal/errors
ok  offer-catcher/server/internal/llm
ok  offer-catcher/server/internal/llmjson
ok  offer-catcher/server/internal/model
ok  offer-catcher/server/internal/prompt
ok  offer-catcher/server/internal/schema
ok  offer-catcher/server/internal/service
```

All pass ✅

### API Key 扫描

```
.git 之外无 sk- / AKIA 真实 Key
仅 .env.example 含占位符 sk-your-real-api-key
```

### git 状态

```
工作区干净，无未提交改动
.env 未被 git 追踪
```

## 13. P0 / P1 / P2

| 级别 | 编号 | 问题 | 状态 |
|------|------|------|------|
| — | — | **P0 = 0** | ✅ |
| — | — | **P1 = 0** | ✅ |
| P2 | P2-1 | 真实 LLM 尚未启用（LLM_ENABLE=false） | K5-1 解决 |
| P2 | P2-2 | 真实 LLM 调用未验证 | K5-1 解决 |
| P2 | P2-3 | 日志可读性可增强 | 后续 |
| P2 | P2-4 | 成本控制策略可增强 | 后续 |
| P2 | P2-5 | VersionCompare Agent 无 LLM 实现（仅 Mock） | 后续 |

## 14. 判定

**K5-0 pass。** 项目已具备安全接入真实 LLM 的基础条件：

- LLM Client 架构完整（接口 + HunyuanClient + NoopClient）✅
- Agent Pipeline 闭环（Prompt → LLM → parse → validate → repair → fallback）✅
- 三类 Agent 共用安全 pipeline，每类都有 Mock fallback ✅
- API Key 安全（json:"-" + MaskedAPIKey + safeErr）✅
- .env.example 仅占位符，.env 未追踪 ✅
- P0=0 P1=0 ✅

## 15. 是否可以进入 K5-1

**可以。** 下一步 K5-1：配置真实 LLM 环境变量并进行最小接通验证。
