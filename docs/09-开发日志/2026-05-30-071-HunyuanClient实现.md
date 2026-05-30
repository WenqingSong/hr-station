# 2026-05-30-071-HunyuanClient实现

## 1. 本次任务

阶段 4-2.5：Hunyuan Client 实现

目标：实现一个 HunyuanClient，让它实现 llm.Client 接口，但暂时不接入业务 Agent。

## 2. 修改范围

### 新增
- `server/internal/llm/hunyuan_client.go`：HunyuanClient 实现
- `server/internal/llm/hunyuan_types.go`：OpenAI-compatible 内部请求/响应类型
- `server/internal/llm/hunyuan_client_test.go`：httptest 单元测试

### 未修改
- `.env` / `.env.example`
- `server/prompts/` / `server/prompts/schema/`
- `server/internal/agent/` / `server/internal/service/`
- `server/internal/controller/` / `server/internal/repository/`
- `Makefile` / `scripts/dev/` / `web/`
- `docker-compose.yml` / `README.md`

## 3. HunyuanClient 设计

### 结构体

```go
type HunyuanClient struct {
    config     Config
    httpClient *http.Client
    endpoint   string
}
```

### 构造函数

```go
func NewHunyuanClient(config Config, httpClient *http.Client) (*HunyuanClient, error)
```

- 构造时校验 Config
- httpClient 为空时使用带 timeout 的默认 client
- endpoint 基于 BaseURL 规范化

### 接口实现

- `Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)`
- `Provider() Provider` → 返回 ProviderHunyuan

## 4. HTTP 请求映射

### endpoint 规范化

`normalizeEndpoint(baseURL)` 规则：
- 已以 `/chat/completions` 结尾 → 直接使用
- 以 `/v1` 结尾 → 拼接 `/chat/completions`
- 以 `/v1/` 结尾 → 拼接 `chat/completions`（无双斜杠）
- 其他 → 拼接 `/v1/chat/completions`

### 请求格式

```
POST {endpoint}
Content-Type: application/json
Authorization: Bearer {config.APIKey}

{
  "model": "{effective model}",
  "messages": [...],
  "temperature": {effective temperature},
  "max_tokens": {effective max_tokens}
}
```

### 字段映射

- model: coalesce(req.Model, config.Model)
- temperature: 当 req.Temperature != 0.2（config default）时使用 req.Temperature，否则使用 config.Temperature
- max_tokens: req.MaxTokens > 0 时使用 req.MaxTokens，否则使用 config.MaxOutputTokens
- timeout: req.TimeoutSec > 0 时使用 req.TimeoutSec，否则使用 config.TimeoutSeconds

## 5. HTTP 响应解析

### 解析流程

1. 读取 HTTP 响应体
2. json.Unmarshal 为 chatCompletionResponse
3. 检查 choices 是否为空 → LLM_EMPTY_RESPONSE
4. 提取 choices[0].message.content
5. 检查 content 是否为空 → LLM_EMPTY_RESPONSE
6. 构建 ChatResponse

### ChatResponse 映射

| 字段 | 来源 |
|------|------|
| Provider | ProviderHunyuan |
| Model | chatResp.Model（为空时用 effective model） |
| RawContent | choices[0].message.content |
| FinishReason | choices[0].finish_reason |
| Usage | chatResp.Usage |
| RequestID | chatResp.ID |
| DurationMs | time.Since(start).Milliseconds() |
| RawOutputLength | len(content) |

## 6. 错误处理与错误码

### HTTP status 映射

| HTTP Status | ErrorCode | 说明 |
|-------------|-----------|------|
| 429 | LLM_RATE_LIMIT | 限流 |
| 401/403 | LLM_API_KEY_MISSING | 认证失败（不打印 token） |
| 5xx | LLM_HTTP_ERROR | 服务器错误 |
| 4xx (其他) | LLM_HTTP_ERROR | 客户端错误 |

### 其他错误映射

| 场景 | ErrorCode |
|------|-----------|
| context.DeadlineExceeded | LLM_TIMEOUT |
| context.Canceled | LLM_TIMEOUT |
| HTTP 请求失败 | LLM_HTTP_ERROR |
| 响应 body 为空 | LLM_EMPTY_RESPONSE |
| 响应 JSON 解析失败 | LLM_HTTP_ERROR |
| choices 为空 | LLM_EMPTY_RESPONSE |
| content 为空 | LLM_EMPTY_RESPONSE |
| messages 为空 | LLM_EMPTY_RESPONSE |

### finish_reason=length

不视为硬错误，返回 ChatResponse（FinishReason="length"），由 Pipeline 判断是否需要 repair。

## 7. 安全边界

1. **错误字符串不包含 API Key**：所有错误路径均不打印 token
2. **Authorization header 不写入错误**：HTTP 错误只记录 status code
3. **Config 不完整时错误不打印完整配置**：只指出缺失字段
4. **hunyuan_types.go 不导出**：OpenAI-compatible 私有结构不暴露到业务层
5. **测试使用 fake key**：不写入真实 API Key

## 8. 测试结果

执行命令：`cd server && go test ./internal/llm -v`

| 测试类别 | 测试项 | 结果 |
|---------|--------|------|
| 构造函数 | ValidConfig | PASS |
| 构造函数 | MissingAPIKey | PASS |
| 构造函数 | MissingBaseURL | PASS |
| 构造函数 | MissingModel | PASS |
| 构造函数 | Provider | PASS |
| Endpoint | V1 | PASS |
| Endpoint | V1TrailingSlash | PASS |
| Endpoint | ChatCompletions | PASS |
| Endpoint | NoDoubleSlash | PASS |
| Endpoint | BareHost | PASS |
| Chat 成功 | Success | PASS |
| Chat 成功 | RequestModelOverride | PASS |
| Chat 成功 | TemperatureZero | PASS |
| Chat 成功 | MaxTokensOverride | PASS |
| Chat 成功 | FinishReasonLength | PASS |
| Chat 成功 | DefaultHTTPClient | PASS |
| Chat 错误 | EmptyMessages | PASS |
| Chat 错误 | RateLimit429 | PASS |
| Chat 错误 | ServerError500 | PASS |
| Chat 错误 | EmptyBody | PASS |
| Chat 错误 | EmptyChoices | PASS |
| Chat 错误 | EmptyContent | PASS |
| Chat 错误 | NonJSONResponse | PASS |
| Chat 错误 | ContextTimeout | PASS |
| Chat 安全 | ErrorNoAPIKey | PASS |
| Chat 安全 | AuthHeaderNotInError | PASS |
| Chat 安全 | ConfigIncomplete_NoFullConfig | PASS |
| Chat 安全 | 401AuthFailed | PASS |

Hunyuan 新增 22 项测试，全部通过。
加上 4-2.4 的 27 项测试，llm 包共 49 项测试，全部通过。

全量测试 `go test ./...` 通过，无失败。

## 9. 文档更新

- 更新 `docs/05-Agent设计/阶段4-2LLMClient与Hunyuan接入方案.md`
- 更新 `docs/05-Agent设计/LLMClient设计.md`
- 更新 `docs/05-Agent设计/Hunyuan接入准备.md`
- 更新 `docs/05-Agent设计/JSON输出规范.md`
- 更新 `docs/05-Agent设计/失败兜底与Mock回退策略.md`
- 更新 `docs/项目状态.md`
- 更新 `docs/下一步任务.md`
- 更新 `docs/变更记录.md`
- 更新 `docs/说明书.md`

## 10. 未修改内容

- 未修改 `.env` / `.env.example`
- 未写入真实 API Key
- 未调用真实 Hunyuan
- 未修改 Prompt 文件
- 未修改 Schema 文件
- 未修改业务 Agent
- 未实现 JSONRepairAgent 调用
- 未实现 Mock fallback
- 未实现统一 Pipeline
- 未修改 Makefile / scripts.dev / docker-compose.yml
- 未修改前端

## 11. 当前风险

- 低风险：Temperature 默认值判断使用了简单启发式（req.Temperature != 0.2 时覆盖），后续 Pipeline 应统一默认值注入
- 低风险：HunyuanClient 仅通过 httptest 验证，尚未对真实 Hunyuan API 发起请求，4-2.6 需要执行 smoke test
- 低风险：如果真实 Hunyuan 响应结构与 OpenAI-compatible 格式有差异，4-2.6 smoke test 时可能需要微调

## 12. 是否可以进入 4-2.6

可以进入 4-2.6。

理由：
1. HunyuanClient 已完整实现 llm.Client 接口
2. 请求/响应映射已实现
3. 错误码分类已实现
4. 安全边界已验证
5. 22 项 httptest 单元测试全部通过
6. 无新增第三方依赖
7. 未修改任何禁止修改的文件

## 13. 下一步建议

1. **4-2.6**：LLM Smoke Test / Makefile 接入 — 使用真实 Hunyuan API 验证连通性
2. **4-2.7**：JSONRepairAgent 调用链路
3. **4-2.8**：统一 LLM Agent Pipeline
