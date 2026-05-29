# LLM Client 设计

> 最后更新：2026-05-29（阶段 4-1 Hunyuan API 接入准备）

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

## 关联文档

- [阶段4Agent与模型接入方案.md](./阶段4Agent与模型接入方案.md)
- [Hunyuan接入准备.md](./Hunyuan接入准备.md)
- [JSON输出规范.md](./JSON输出规范.md)
- [Agent总览.md](./Agent总览.md)
