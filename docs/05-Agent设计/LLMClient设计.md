# LLM Client 设计

> 最后更新：2026-05-30（阶段 4-3.0 业务 Agent 接入 LLM 方案）

## 1. 设计目标

1. 提供统一的 LLM 调用抽象，屏蔽不同 Provider 差异
2. 支持结构化 JSON 输出（Agent 必须只输出 JSON）
3. 支持超时控制与重试
4. 支持失败兜底回退 Mock Agent
5. 不引入复杂框架，使用 Go 标准库 + HTTP Client

## 2. Provider 抽象

```go
// LLMClient 定义 LLM 调用抽象
type LLMClient interface {
    // ChatJSON 发送聊天请求，要求模型只返回 JSON
    // 返回值为模型输出的原始 JSON 字节
    ChatJSON(ctx context.Context, req ChatRequest) ([]byte, error)
}
```

当前只实现 Hunyuan Provider（OpenAI-compatible），后续按需扩展。

### Provider 注册

```go
func NewLLMClient(provider string, cfg *config.Config) LLMClient {
    switch provider {
    case "hunyuan":
        return NewHunyuanClient(cfg)
    default:
        return nil // 未配置则返回 nil，使用 Mock Agent
    }
}
```

## 3. Hunyuan OpenAI-compatible 调用方式

Hunyuan 提供 OpenAI-compatible API：

```
POST {LLM_BASE_URL}/chat/completions
Authorization: Bearer {LLM_API_KEY}
Content-Type: application/json

{
  "model": "{LLM_MODEL}",
  "messages": [
    {"role": "system", "content": "系统 Prompt"},
    {"role": "user", "content": "用户 Prompt"}
  ],
  "temperature": {LLM_TEMPERATURE},
  "max_tokens": {LLM_MAX_OUTPUT_TOKENS}
}
```

响应格式与 OpenAI Chat Completions 一致：
- `choices[0].message.content` 包含模型输出文本
- `usage` 包含 token 使用量

## 4. 请求结构

```go
// ChatRequest LLM 请求
type ChatRequest struct {
    SystemPrompt string // 系统 Prompt
    UserPrompt   string // 用户 Prompt
    Temperature  float64
    MaxTokens    int
}

// ChatMessage 聊天消息
type ChatMessage struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

// ChatCompletionRequest OpenAI-compatible 请求体
type ChatCompletionRequest struct {
    Model       string        `json:"model"`
    Messages    []ChatMessage `json:"messages"`
    Temperature float64       `json:"temperature"`
    MaxTokens   int           `json:"max_tokens"`
}
```

## 5. 响应结构

```go
// ChatCompletionResponse OpenAI-compatible 响应体
type ChatCompletionResponse struct {
    ID      string   `json:"id"`
    Object  string   `json:"object"`
    Choices []Choice `json:"choices"`
    Usage   Usage    `json:"usage"`
}

// Choice 选择项
type Choice struct {
    Index        int     `json:"index"`
    Message      ChatMessage `json:"message"`
    FinishReason string  `json:"finish_reason"`
}

// Usage token 使用量
type Usage struct {
    PromptTokens     int `json:"prompt_tokens"`
    CompletionTokens int `json:"completion_tokens"`
    TotalTokens      int `json:"total_tokens"`
}
```

## 6. 超时与重试

### 超时控制

- 使用 `context.WithTimeout` 设置请求超时
- 默认超时 20 秒（通过 `LLM_TIMEOUT_SECONDS` 配置）
- 超时后返回 error，由调用方决定是否回退 Mock

### 重试策略

- 最大重试次数由 `LLM_MAX_RETRIES` 控制，默认 1 次
- 只对以下情况重试：
  - HTTP 5xx 错误
  - 网络超时
- 不对以下情况重试：
  - HTTP 4xx 错误（认证失败、配额超限等）
  - JSON 解析失败（LLM 输出格式错误，重试无意义）

### 重试间隔

- 使用固定间隔，不引入指数退避
- 间隔 1 秒
- Demo 阶段不需要复杂重试策略

## 7. 错误分类

| 错误类型 | 说明 | 处理方式 |
|---------|------|---------|
| `ErrLLMUnavailable` | LLM 服务不可用（5xx / 网络错误） | 回退 Mock |
| `ErrLLMAuth` | 认证失败（401 / 403） | 回退 Mock，日志告警 |
| `ErrLLMQuota` | 配额超限（429） | 回退 Mock，日志告警 |
| `ErrLLMTimeout` | 请求超时 | 回退 Mock |
| `ErrLLMJSONParse` | JSON 解析失败 | 回退 Mock |
| `ErrLLMFieldValidation` | 字段校验失败 | 回退 Mock |
| `ErrLLMContentViolation` | 输出包含禁止内容（虚构经历等） | 回退 Mock |

## 8. JSON 输出解析

### 8.1 提取 JSON

LLM 输出可能包含 Markdown 代码块包裹：

```
```json
{"match_score": 85}
```
```

解析步骤：
1. 尝试直接 `json.Unmarshal`
2. 如果失败，尝试提取 Markdown 代码块中的内容
3. 去除可能的 ```json 和 ``` 标记
4. 再次 `json.Unmarshal`

### 8.2 校验 JSON

详细规范见 [JSON输出规范.md](./JSON输出规范.md)

## 9. 失败兜底

当 LLM 调用失败时：

1. 判断 `ENABLE_LLM_FALLBACK` 是否为 `true`
2. 如果为 `true`，回退到对应的 Mock Agent
3. 如果为 `false`，返回错误响应给前端
4. 响应中通过 `generated_by` 字段标识来源

```go
func (s *MatchService) Analyze(ctx context.Context, ...) (*MatchAnalysis, error) {
    if s.cfg.Feature.EnableLLMAgent {
        result, err := s.llmClient.ChatJSON(ctx, req)
        if err == nil {
            // 解析 + 校验
            analysis, err := parseAndValidate(result)
            if err == nil {
                analysis.GeneratedBy = "llm_agent"
                return analysis, nil
            }
        }
        // LLM 失败
        if s.cfg.Feature.EnableLLMFallback {
            analysis := s.mockAgent.Analyze(...)
            analysis.GeneratedBy = "llm_agent_fallback"
            return analysis, nil
        }
        return nil, err
    }
    // 使用 Mock
    analysis := s.mockAgent.Analyze(...)
    analysis.GeneratedBy = "mock_agent"
    return analysis, nil
}
```

## 10. 不做事项

| 不做 | 原因 |
|------|------|
| 不引入 LangChain 等 Agent 框架 | 项目定位为自研 Agent 层，简单 Prompt + JSON 即可 |
| 不做流式输出 | 当前同步返回即可 |
| 不做 Function Calling | 用 JSON 输出约束替代 |
| 不做 Embedding / RAG | Demo 阶段不需要 |
| 不做多模型并行 | 单模型即可 |
| 不做缓存 | 后续按需添加 |
| 不做 Token 计费统计 | 后续按需添加 |

## 11. 后续实现步骤

| 步骤 | 任务 | 文件 |
|------|------|------|
| 1 | 实现 LLMClient 接口和类型定义 | `server/internal/llm/client.go` |
| 2 | 实现 Hunyuan Client | `server/internal/llm/hunyuan_client.go` |
| 3 | 实现请求/响应类型 | `server/internal/llm/types.go` |
| 4 | 实现 JSON 解析和校验 | `server/internal/llm/json_validator.go` |
| 5 | 在 Config 中增加新字段 | `server/internal/config/config.go` |
| 6 | 在 main.go 中初始化 LLM Client | `server/cmd/server/main.go` |
| 7 | 改造各 Agent Service | `server/internal/service/*.go` |
| 8 | 编写 Prompt 模板 | `server/internal/agent/prompts/*.md` |

## 阶段 4-1.4 更新

### 4-1.4.11 失败兜底与 Mock 回退策略更新

后续 LLM Client 需要支持以下能力：

```text
timeout：超时控制
error_code：错误码返回
raw_output_length：原始输出长度记录
json_parse_result：JSON 解析结果
schema_validate_result：Schema 校验结果
repair_attempted：是否尝试 repair
repair_success：repair 是否成功
fallback_used：是否使用 fallback
fallback_reason：fallback 原因
```

策略文档见 `docs/05-Agent设计/失败兜底与Mock回退策略.md`

### 4-1.4.12 阶段总验收更新

4-1.4 总验收已完成。后续 4-2 实现 LLM Client 时，应按 4-1.4 产出的 Prompt / Schema / fallback 策略接入。

验收报告见 `docs/05-Agent设计/阶段4-1.4总验收报告.md`

### 4-1.4.13 阶段汇总更新

阶段 4-1.4 已正式闭环。下一阶段进入 4-2 实现 LLM Client 抽象与 Hunyuan Client。

4-2 重点：
- LLM Client 抽象接口
- Hunyuan Client 实现
- Prompt Loader + Schema Loader
- JSON parse + Schema validate
- JSON repair 调用链路
- Mock fallback 集成

汇总文档见 `docs/00-项目总览/阶段4-1.4Prompt与Agent真实接入汇总.md`

4-1.4 不实现 LLM Client，4-2 才实现 LLM Client 与 Hunyuan Client：

- LLM Client 将依赖 4-1.4 固定的 Prompt 与 Schema
- config.go 需在 4-2 中增加 EnableLLMAgent / EnableLLMFallback / LLMProvider / LLMTimeoutSeconds / LLMMaxRetries / LLMMaxOutputTokens / LLMTemperature 等 7 个字段
- server/internal/llm/ 当前仅存在 .gitkeep，4-2 从零实现
- 方案文档见 `docs/05-Agent设计/阶段4-1.4Prompt与Agent真实接入方案.md`

## 阶段 4-2 更新

4-2.0 已完成方案设计，详细方案见 `docs/05-Agent设计/阶段4-2LLMClient与Hunyuan接入方案.md`。

关键设计决策：

1. **LLM Client 后续应作为统一模型调用抽象**：业务 Agent 不直接依赖 Hunyuan，只依赖 LLM Client / Runner 抽象
2. **Hunyuan Client 是 LLM Client 的 provider 实现**：通过工厂函数 NewLLMClient(provider, cfg) 创建
3. **LLMRunner 统一调用链路**：Chat → JSON parse → Schema validate → JSON repair once → Mock fallback
4. **Prompt Loader 从 server/prompts/ 加载 .md 文件**：不把 Prompt 写死在 Go 代码
5. **Schema Loader 从 server/prompts/schema/ 加载 JSON Schema**：推荐使用 santhosh-tekuri/jsonschema/v5 作为校验库
6. **4-2 只建基础设施，4-3 才改造业务 Agent**

## 阶段 4-2.1 更新

4-2.1 已固化 LLM_* 配置项到 `.env.example`，共 14 项配置：

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| LLM_ENABLE | false | 是否启用 LLM |
| LLM_PROVIDER | hunyuan | LLM provider |
| LLM_BASE_URL | （空） | API Base URL |
| LLM_API_KEY | （空） | API Key，只放在 .env |
| LLM_MODEL | （空） | 模型名称 |
| LLM_TEMPERATURE | 0.2 | 温度 |
| LLM_MAX_OUTPUT_TOKENS | 8192 | 最大输出 token |
| LLM_TIMEOUT_SECONDS | 30 | 请求超时（秒） |
| LLM_REPAIR_ENABLE | true | 是否启用 JSON repair |
| LLM_REPAIR_MAX_ATTEMPTS | 1 | repair 最大尝试次数 |
| LLM_FALLBACK_ENABLE | true | 是否启用 Mock fallback |
| LLM_LOG_RAW_OUTPUT | false | 是否记录原始输出 |
| LLM_PROMPT_DIR | server/prompts | Prompt 文件目录 |
| LLM_SCHEMA_DIR | server/prompts/schema | Schema 文件目录 |

安全修复：清除了 `.env.example` 中遗留的真实 API Key。

## 阶段 4-2.2 更新

4-2.2 已完成 Prompt Loader / Schema Loader，后续 LLM Client / Pipeline 应复用 Loader 加载 Prompt 和 Schema。

Prompt Loader 能力：
- 从 server/prompts/ 加载 4 个 Prompt 文件（match_analysis / recommendation / resume_optimization / json_repair）
- 支持占位符替换（{{KEY}} 格式）
- 只允许预定义逻辑名称，防止路径穿越
- 单元测试全部通过（10 项）

Schema Loader 能力：
- 从 server/prompts/schema/ 加载 3 个 Schema 文件
- JSON 合法性检查（顶层必须为 object）
- 提供 Title() / RequiredCount() / Draft() 等元信息方法
- 只允许预定义逻辑名称，防止路径穿越
- 单元测试全部通过（11 项）

当前仍未实现模型调用。

## 阶段 4-2.3 更新

4-2.3 已完成 JSON parse + Schema validate 基础工具，后续 LLM Client / Pipeline 需要复用该工具处理 raw output。

新增 Go 包：
- `server/internal/llmjson/`：从 LLM raw output 提取 JSON object（支持 Markdown code fence、前后解释文字、BOM 清理）、JSON parse、错误码分类（JSON_EMPTY / JSON_OBJECT_NOT_FOUND / JSON_PARSE_FAILED / JSON_TOP_LEVEL_NOT_OBJECT）
- `server/internal/schema/validator.go`：Validator 使用 santhosh-tekuri/jsonschema/v5 实现 JSON Schema 校验，支持 draft 2020-12，返回结构化 ValidationResult（Success / ErrorCode / Errors）

LLM Client 本身仍未实现。

## 阶段 4-2.4 更新

4-2.4 已完成 LLM Client 抽象接口，后续业务 Agent 只依赖 `llm.Client` 接口，不直接依赖 Hunyuan。

新增 Go 包：
- `server/internal/llm/` 包：统一 LLM Client 抽象和基础类型
  - `Client` 接口：`Chat(ctx, ChatRequest) (*ChatResponse, error)` + `Provider() Provider`
  - `ChatRequest` / `ChatResponse` / `Message` / `Usage` / `Provider` / `Role`
  - `Config` 结构体：覆盖 .env.example 全部 14 项 LLM 配置，含 `DefaultConfig()` 和 `Validate()`
  - `ErrorCode`：9 个错误码（LLM_DISABLED / LLM_CONFIG_MISSING / LLM_API_KEY_MISSING / LLM_TIMEOUT / LLM_HTTP_ERROR / LLM_RATE_LIMIT / LLM_EMPTY_RESPONSE / LLM_OUTPUT_TRUNCATED / UNKNOWN_ERROR）
  - `LLMError`：结构化错误，支持 `Unwrap()` / `IsErrorCode()` / `IsDirectMockFallback()`
  - `NoopClient`：不访问外部网络的测试实现，Enable=false 时返回 LLM_DISABLED
- 单元测试 27 项全部通过

核心设计决策：
1. **业务 Agent 后续只依赖 llm.Client**，不直接依赖 Hunyuan
2. **Hunyuan Client 将作为 llm.Client 的 provider 实现放到 4-2.5**
3. **ChatResponse.RawContent 后续必须经过 4-2.3 的 llmjson + Schema validate 处理**
4. **llm.Config.Validate() 在 Enable=true 时强制要求 Provider/BaseURL/APIKey/Model**
5. **IsDirectMockFallback() 判断哪些错误码应直接 Mock fallback**

## 阶段 4-2.5 更新

4-2.5 已完成 Hunyuan Client 实现。HunyuanClient 作为 llm.Client 的 provider 实现，业务 Agent 仍只依赖 llm.Client，不直接依赖 Hunyuan 私有 HTTP 结构。

新增 Go 文件：
- `server/internal/llm/hunyuan_client.go`：HunyuanClient 结构体，实现 llm.Client 接口
  - NewHunyuanClient(config, httpClient)：构造函数，支持注入 http.Client 便于 httptest
  - Chat(ctx, ChatRequest)：发送 OpenAI-compatible chat completions 请求，解析响应
  - Provider()：返回 ProviderHunyuan
  - normalizeEndpoint(baseURL)：规范化 BaseURL，兼容 /v1、/v1/、/v1/chat/completions
- `server/internal/llm/hunyuan_types.go`：OpenAI-compatible 请求/响应内部类型（不导出）
  - chatCompletionRequest / chatCompletionResponse / chatChoice / chatMessage / chatCompletionUsage
- `server/internal/llm/hunyuan_client_test.go`：httptest 单元测试（22 项）

核心设计决策：
1. **HunyuanClient 不暴露私有 HTTP 结构**：chatCompletionRequest / chatCompletionResponse 不导出
2. **finish_reason=length 不视为硬错误**：返回 ChatResponse，Pipeline 决定是否 repair
3. **401/403 映射为 LLM_API_KEY_MISSING**：错误信息写 "authentication failed"，不打印 token
4. **支持注入 http.Client**：便于 httptest 和后续集成测试
5. **当前仅通过 httptest 验证**，不调用真实 Hunyuan，真实连通性放到 4-2.6

## 阶段 4-2.7 更新

4-2.7 已完成 JSONRepairAgent 调用链路。JSONRepairAgent 通过 `llm.Client` 调用模型，不直接依赖 HunyuanClient。repair 成功后业务结果 `generated_by` 仍保持 `llm_agent`。

新增：
- `server/internal/agent/repair/types.go`：TargetOutputType、RepairRequest、RepairResult、7 个错误码
- `server/internal/agent/repair/json_repair_agent.go`：Agent 结构体、NewAgent、Repair 方法（repair once）
- `server/internal/agent/repair/json_repair_agent_test.go`：14 项单元测试（全部通过）

核心设计决策：
1. **JSONRepairAgent 只修复结构，不重新做业务分析**
2. **repair 只执行一次，不递归、不重试**
3. **repair 通过 llm.Client 抽象层调用模型**
4. **repair output 必须再次经过 JSON extract / parse / Schema validate**
5. **错误消息中不泄露 API Key、完整 Prompt、完整 BrokenOutput**
6. **repair 失败只返回结构化错误，不执行 Mock fallback**

## 阶段 4-2.6 更新

4-2.6 已完成 LLM smoke test 入口和 Makefile 接入。后续可通过 `make llm-check` 检查配置，通过 `RUN_REAL_LLM_SMOKE=1 make llm-smoke` 验证 HunyuanClient 连通性。

新增：
- `server/cmd/llm-smoke/main.go`：最小 LLM smoke test Go 命令
  - 从环境变量读取 LLM_* 配置
  - 构造最小 ChatRequest（temperature=0, max_tokens=256）
  - 调用 HunyuanClient.Chat
  - 尝试 llmjson.ExtractAndParse 解析 raw output
  - 不输出 API Key，不输出完整 raw output（最多 120 字符预览）
- `scripts/dev/llm_smoke.sh`：LLM smoke test 脚本（help / check / smoke）
- Makefile 新增 `llm-help` / `llm-check` / `llm-smoke` target
  - `make llm-check` 不请求外部模型
  - `make llm-smoke` 需要 `RUN_REAL_LLM_SMOKE=1` 才执行真实调用

## 关联文档

- [阶段4Agent与模型接入方案.md](./阶段4Agent与模型接入方案.md)
- [Hunyuan接入准备.md](./Hunyuan接入准备.md)
- [JSON输出规范.md](./JSON输出规范.md)
- [Agent总览.md](./Agent总览.md)

## 阶段 4-3.0 更新

4-3.0 已完成业务 Agent 接入 LLM 方案设计（docs only）。关键设计决策：

1. **业务 Agent 依赖 Pipeline Runner**：Service 层注入 pipelineRunner，不直接依赖 HunyuanClient
2. **Pipeline Runner 依赖 llm.Client**：pipelineRunner.NewRunner(llmClient, promptLoader, schemaLoader, validator, repairAgent)
3. **业务 Agent 不直接调用 llm.Client**：所有模型调用通过 Pipeline Runner 完成
4. **HunyuanClient 仍只作为 llm.Client 的 provider 实现**：业务 Agent 不感知 Provider 差异
5. **4-3.1 需先执行真实 LLM Smoke Test**：验证 HunyuanClient 连通性后再接入业务 Agent
6. **方案文档见 `docs/05-Agent设计/阶段4-3业务Agent接入LLM方案.md`**

## 阶段 4-3.1 更新

4-3.1 已通过 llm-smoke 对 HunyuanClient 真实连通性进行前置验收，但真实 smoke 未执行。

原因：本地 .env 不存在，LLM 配置不完整（LLM_ENABLE=false，LLM_BASE_URL / LLM_API_KEY / LLM_MODEL 未配置）。

结论：
- HunyuanClient 代码逻辑正确（httptest 22 项单元测试通过）
- 真实连通性未验证
- 业务 Agent 接入仍必须通过 Pipeline，不直接调用 HunyuanClient
- 暂不建议进入 4-3.2

验收报告见 `docs/05-Agent设计/真实LLMSmoke前置验收报告.md`。

## 阶段 4-3.1.B 更新

4-3.1.B 再次通过 llm-smoke 对 HunyuanClient 真实连通性进行前置核验，但真实 smoke 仍未执行。

原因：本地 .env 仍然不存在，LLM 配置仍不完整（与前次 4-3.1 一致）。

结论：
- HunyuanClient 代码逻辑正确（httptest 22 项单元测试通过）
- 真实连通性仍未验证
- 两次核验均因配置缺失无法执行
- 业务 Agent 接入仍必须通过 Pipeline，不直接调用 HunyuanClient
- 暂不建议进入 4-3.2

## 阶段 4-3.1 重新执行更新

4-3.1 重新执行仍未能验证 HunyuanClient 真实连通性。

原因：LLM_ENABLE=false，LLM_BASE_URL 未配置。`.env` 已存在（4-3.1.C 自动创建）但只有模板值。

三次核验均因配置不完整无法执行。业务 Agent 接入仍必须通过 Pipeline，不直接调用 HunyuanClient。

## 阶段 4-3.1 通过结果登记更新

4-3.1 已通过 llm-smoke 验证 HunyuanClient 真实连通性。

真实 smoke 结果：
- provider=hunyuan，model=hy3-preview
- finish_reason=stop，json_parse_success=true
- raw_output_length=28，duration_ms=1327

业务 Agent 接入时仍必须调用统一 Pipeline，不直接依赖 HunyuanClient。
