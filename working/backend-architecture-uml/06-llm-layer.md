# 06 — LLM 客户端层类图

来源：`/workspace/server/internal/llm/`（13 个文件）。

## 1. LLM 层类图

```mermaid
classDiagram
    class Client {
        <<interface>>
        +Chat(ctx, ChatRequest) (*ChatResponse, error)
        +Provider() Provider
    }

    class HunyuanClient {
        -config Config
        -httpClient *http.Client
        -endpoint string
        +NewHunyuanClient(config, httpClient)
        +Chat(ctx, req)
        +Provider() ProviderHunyuan
    }

    class Hy3Client {
        -config Config
        -httpDoer *http.Client
        +NewHy3Client(config, httpClient)
        +Chat(ctx, req)
        +Provider() ProviderHy3
    }

    class NoopClient {
        -config Config
        +NewNoopClient(config)
        +Chat(ctx, req)
        +Provider() ProviderNoop
    }

    class Config {
        +Enable bool
        +Provider Provider
        +BaseURL string
        +APIKey string
        +Model string
        +Temperature float64
        +MaxOutputTokens int
        +TimeoutSeconds int
        +RepairEnable bool
        +RepairMaxAttempts int
        +FallbackEnable bool
        +PromptDir string
        +SchemaDir string
        +Validate() error
        +MaskedAPIKey() string
    }

    class ChatRequest {
        +Model string
        +Messages []Message
        +Temperature float64
        +TopP float64
        +MaxTokens int
        +ReasoningEffort string
        +ExtraBody map~string~any
        +TimeoutSec int
        +Metadata map~string~string
    }

    class ChatResponse {
        +Provider Provider
        +Model string
        +RawContent string
        +FinishReason string
        +Usage Usage
        +RequestID string
        +DurationMs int64
        +RawOutputLength int
    }

    class LLMError {
        +Code ErrorCode
        +Message string
        +Cause error
        +Error() string
        +Unwrap() error
    }

    Client <|.. HunyuanClient
    Client <|.. Hy3Client
    Client <|.. NoopClient
    HunyuanClient --> Config
    Hy3Client --> Config
    NoopClient --> Config
    HunyuanClient ..> ChatRequest
    HunyuanClient ..> ChatResponse
    Hy3Client ..> ChatRequest
    Hy3Client ..> ChatResponse
    HunyuanClient ..> LLMError
    Hy3Client ..> LLMError
```

## 2. Provider 工厂（main.go 中的 `newLLMClient`）

```mermaid
flowchart TD
    P{LLM_PROVIDER?}
    P -->|"hy3"| HY3["llm.NewHy3Client(cfg, nil)<br/>腾讯云 LKEAP /plan/v3/chat/completions"]
    P -->|"其他（deepseek/hunyuan/tokenhub）"| HUNY["llm.NewHunyuanClient(cfg, nil)<br/>OpenAI 兼容 /v1/chat/completions"]
```

| Provider 常量 | 值 | 说明 |
|--------------|-----|------|
| `ProviderHunyuan` | `"hunyuan"` | 混元 / TokenHub / DeepSeek（OpenAI 兼容） |
| `ProviderHy3` | `"hy3"` | 腾讯云 LKEAP `hy3-preview` |
| `ProviderNoop` | `"noop"` | 测试用非网络实现 |

## 3. Provider 差异对比

| 维度 | HunyuanClient | Hy3Client |
|------|--------------|-----------|
| 端点 | `normalizeEndpoint` 归一化为 `/v1/chat/completions` | `/plan/v3/chat/completions`（LKEAP） |
| ReasoningEffort | 忽略 | 写入 `extra_body.chat_template_kwargs.reasoning_effort` |
| TopP | 未单独传（OpenAI 兼容体可带） | 支持 `top_p` |
| 错误分类 | 401/403→`LLM_AUTH_FAILED`，429→`LLM_RATE_LIMIT`，5xx→`LLM_HTTP_ERROR` | 同左，`errorFromHTTPStatus` |
| 响应体上限 | 无显式截断 | `io.LimitReader` 1MB |

## 4. ErrorCode 枚举

| ErrorCode | 触发场景 |
|-----------|---------|
| `LLM_DISABLED` | `LLM_ENABLE=false` |
| `LLM_CONFIG_MISSING` | 必填配置缺失 |
| `LLM_API_KEY_MISSING` | API key 为空 |
| `LLM_AUTH_FAILED` | HTTP 401/403（key 无效/不匹配平台） |
| `LLM_TIMEOUT` | 超时 / 取消 |
| `LLM_HTTP_ERROR` | 网络错误 / 5xx |
| `LLM_RATE_LIMIT` | HTTP 429 |
| `LLM_EMPTY_RESPONSE` | 空内容 |
| `LLM_OUTPUT_TRUNCATED` | `finish_reason=length` |
| `UNKNOWN_ERROR` | 未分类 |

**兜底决策辅助**：`llm.IsDirectMockFallback(err)` 判断是否**直接转 Mock**（配置/鉴权/超时/限流/空响应/未知），跳过 JSON 修复。这被 Pipeline 用于决定是走 `attemptRepair` 还是立即 `fallback`。

## 5. 安全设计

- `ChatRequest.Metadata` / `ChatResponse` 明确约束：**绝不携带 API key**。
- `LLMError` 不变量：**绝不包含 API key、完整 Prompt、完整简历、完整 JD**。
- `Config.APIKey` JSON tag 为 `-`（不序列化），`MaskedAPIKey()` 用于日志脱敏。
- Pipeline / Repair 层的 `safeErr` / `redactSensitive` 二次脱敏（`sk-`、`key=`、`token=` 等模式）。

## 6. 与上层的耦合

```mermaid
graph LR
    SERVICE["service（Deep/SEP 路径）"] --> CLIENT["llm.Client"]
    PIPELINE["agent/pipeline.Runner"] --> CLIENT
    REPAIR["agent/repair.Agent"] --> CLIENT
    CLIENT --> HUNY["HunyuanClient"]
    CLIENT --> HY3["Hy3Client"]
```

- `pipeline.Runner` / `repair.Agent` 通过 `llm.Client` 接口调用，是最核心的 LLM 消费方。
- `service` 层仅有 `MatchService`、`RecommendationService`、`SemanticProfileProfiler` 直接持有 `llm.Client`（跑 Deep 5 模块 / DeepCascade / SEP 画像生成）。
- 业务 Agent 不依赖 `llm.Client`（依赖 `PipelineRunner`）。
