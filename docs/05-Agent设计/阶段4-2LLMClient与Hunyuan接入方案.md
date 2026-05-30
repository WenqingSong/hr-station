# 阶段 4-2 LLM Client 与 Hunyuan 接入方案

> 创建时间：2026-05-29
> 所属阶段：4-2.0
> 依据文档：阶段4-1.4方案、阶段4-1.4汇总、Agent总览、Prompt模板、JSON输出规范、LLMClient设计、Hunyuan接入准备、失败兜底与Mock回退策略、config.go、main.go、.env.example、go.mod
> 最后更新：2026-05-30（阶段 4-2.6 LLM Smoke Test / Makefile 接入）

## 1. 阶段目标

4-2 的目标不是直接完成业务 Agent 的 LLM 化，而是建立统一的大模型基础设施层。

具体目标：

1. **统一 LLM Client 抽象**：定义与 Provider 无关的 LLM 调用接口
2. **接入 Hunyuan Client**：基于 OpenAI-compatible API 实现第一个 Provider
3. **支持 Prompt 文件加载**：从 `server/prompts/` 加载 `.md` Prompt 模板
4. **支持 Schema 文件加载**：从 `server/prompts/schema/` 加载 JSON Schema
5. **支持 JSON parse**：从模型 raw output 提取合法 JSON
6. **支持 Schema validate**：使用 JSON Schema 校验模型输出
7. **支持 JSON repair once**：JSONRepairAgent 调用一次修复
8. **支持 Mock fallback 决策**：通用 fallback 判断逻辑
9. **为 4-3 业务 Agent 接入真实 LLM 做准备**：4-2 只建基础设施，4-3 才改造业务 Agent

## 2. 当前前提

### 2.1 已有

| 资产 | 位置 | 说明 |
|------|------|------|
| 腾讯岗位数据和完整 JD | jobs 表 + tencent_join | data_quality=full_jd |
| 匹配分析 Prompt | server/prompts/match_analysis_v1.md | HR初筛+技术面试官双视角 |
| 岗位推荐 Prompt | server/prompts/recommendation_v1.md | 四档推荐+Offer鹅语气 |
| 简历优化 Prompt | server/prompts/resume_optimization_v1.md | P0/P1/P2+关键词对齐 |
| JSON 修复 Prompt | server/prompts/json_repair_v1.md | 只修结构不重做业务分析 |
| 匹配分析 Schema | server/prompts/schema/match_analysis_schema.json | 17 required 字段 |
| 推荐 Schema | server/prompts/schema/recommendation_schema.json | 7 required 字段 |
| 简历优化 Schema | server/prompts/schema/resume_optimization_schema.json | 15 required 字段 |
| 失败兜底策略 | docs/05-Agent设计/失败兜底与Mock回退策略.md | 20 种错误码 |
| Mock Agent | server/internal/agent/*.go | 推荐Mock+匹配Mock+版本对比Mock |
| Go 后端基础服务 | server/ | 14 个接口，JWT，统一响应 |
| 推荐接口 | POST /api/jobs/recommend | 当前 Mock Agent |
| 匹配接口 | POST /api/jobs/{id}/match | 当前 Mock Agent |
| LLM 配置 | .env.example | ENABLE_LLM_AGENT 等 10 项 |
| config.go | LLMConfig + FeatureConfig | 已有但缺少 EnableLLMAgent 等字段 |
| server/internal/llm/ | 空目录 | 仅有 .gitkeep |

### 2.2 尚未有

| 缺失项 | 说明 |
|--------|------|
| LLM Client | 统一模型调用抽象接口 |
| Hunyuan Client | LLM Client 的 Hunyuan Provider 实现 |
| Prompt Loader | 从 server/prompts/ 加载 .md 文件 |
| Schema Loader | 从 server/prompts/schema/ 加载 JSON Schema |
| Schema Validator | 使用 JSON Schema 校验模型输出 |
| JSON Repair 调用器 | 调用 JSONRepairAgent 一次修复 |
| LLM Agent Pipeline | 统一的 LLM → parse → validate → repair → fallback 链路 |

## 3. 阶段边界

### 3.1 4-2 做什么

- LLM 配置项补齐
- LLM Client 抽象接口
- Hunyuan Client 实现
- Prompt Loader
- Schema Loader
- JSON parse 工具
- Schema validate 工具
- JSON repair once 调用
- Mock fallback decision
- Smoke test
- Makefile 接入

### 3.2 4-2 不做什么

| 不做 | 原因 |
|------|------|
| 不把 MatchAnalysisAgent 接入 LLM | 4-3 才改造业务 Agent |
| 不把 RecommendationAgent 接入 LLM | 4-3 才改造业务 Agent |
| 不把 ResumeOptimizationAgent 接入 LLM | 4-3 才改造业务 Agent |
| 不改前端 | 阶段 5 |
| 不改业务接口返回结构 | 4-4 才统一改 |
| 不做最终演示页面 | 阶段 7 |
| 不改 Prompt | 4-1.4 已定稿 |
| 不改 Schema | 4-1.4 已定稿 |
| 不改 .env.example | 4-2.1 才改 |
| 不请求腾讯招聘接口 | 不在本阶段范围内 |

## 4. 4-2 与 4-3 的分工

| 阶段 | 范围 | 产出 |
|------|------|------|
| 4-2 | 建底层能力 | LLMRunner / PromptLoader / SchemaLoader / SchemaValidator / HunyuanClient / JSONRepairAgent 调用器 / fallback decision |
| 4-3 | 业务 Agent 使用底层能力 | MatchAnalysisAgent 接入 LLM → 生成真实匹配分析结果；RecommendationAgent 接入 LLM → 生成真实推荐结果 |

核心判断：**4-2 产出工具，4-3 产出业务价值。**

4-2 产出 LLMRunner / PromptLoader / SchemaValidator / HunyuanClient 等基础设施后，4-3 的 MatchAnalysisAgent 才调用这些能力生成真实匹配分析结果。

## 5. 总体调用链路

后续目标链路：

```
业务 Agent（4-3 才改造）
→ 构造上下文
→ 选择 Prompt 文件
→ Prompt Loader 加载模板
→ 替换占位符
→ Schema Loader 加载目标 Schema
→ LLM Client 调用 Hunyuan
→ 获取 raw output
→ 提取 JSON
→ JSON parse
→ Schema validate
→ 如果失败：JSONRepairAgent 调用一次
  → repair output parse + validate
  → 如果通过：返回 repaired 结果
  → 如果仍失败：Mock fallback
→ 如果通过：返回 LLM 结果
→ 返回稳定业务结果
```

标注：

- **4-2 只实现从 Prompt Loader 到 fallback decision 的基础设施**
- **4-3 才把业务 Agent 接入该链路**

## 6. 建议目录结构

根据当前项目结构（`server/internal/llm/` 已存在但为空），建议：

```
server/internal/llm/
  client.go           # LLMClient 接口定义 + NewLLMClient 工厂
  types.go            # ChatRequest / ChatResponse / ChatMessage / Usage / ErrorCode 等
  hunyuan_client.go   # Hunyuan Client 实现（OpenAI-compatible）
  runner.go           # LLMRunner：统一调用 → parse → validate → repair → fallback

server/internal/prompt/
  loader.go           # PromptLoader：从 server/prompts/ 加载 .md
  types.go            # PromptTemplate / PromptMapping 等

server/internal/schema/
  loader.go           # SchemaLoader：从 server/prompts/schema/ 加载 .json
  validator.go        # SchemaValidator：JSON Schema 校验
  types.go            # SchemaMapping / ValidationResult 等

server/internal/agent/repair/
  json_repair_agent.go # JSONRepairAgent 调用器（调用 LLM + json_repair_v1.md）
```

当前已存在：

```
server/prompts/
  match_analysis_v1.md
  recommendation_v1.md
  resume_optimization_v1.md
  json_repair_v1.md
  schema/
    match_analysis_schema.json
    recommendation_schema.json
    resume_optimization_schema.json
```

必须说明：**本阶段 4-2.0 只设计目录，不创建目录和文件。**

## 7. 环境变量配置设计

后续应新增 / 补齐的配置项（4-2.1 已完成）：

| 配置项 | 默认值 | 类型 | 说明 |
|--------|--------|------|------|
| LLM_ENABLE | false | 开关 | 是否启用 LLM（false 则直接 Mock） |
| LLM_PROVIDER | hunyuan | 字符串 | 模型提供商 |
| LLM_BASE_URL | （空） | URL | OpenAI-compatible API 地址 |
| LLM_API_KEY | （空） | 密钥 | API Key，只放在 .env |
| LLM_MODEL | （空） | 字符串 | 模型名称 |
| LLM_TEMPERATURE | 0.2 | 浮点 | 温度，低温度保稳定 |
| LLM_MAX_OUTPUT_TOKENS | 8192 | 整数 | 最大输出 token（从 2048 提升） |
| LLM_TIMEOUT_SECONDS | 30 | 整数 | 请求超时（从 20 提升） |
| LLM_REPAIR_ENABLE | true | 开关 | 是否启用 JSON repair |
| LLM_REPAIR_MAX_ATTEMPTS | 1 | 整数 | repair 最大尝试次数（固定为 1） |
| LLM_FALLBACK_ENABLE | true | 开关 | 是否启用 Mock fallback |
| LLM_LOG_RAW_OUTPUT | false | 开关 | 是否记录原始输出（调试用） |
| LLM_PROMPT_DIR | server/prompts | 字符串 | Prompt 文件目录 |
| LLM_SCHEMA_DIR | server/prompts/schema | 字符串 | Schema 文件目录 |

与旧版 `.env.example` 的差异（4-2.1 已完成更新）：

| 旧配置项 | 问题 | 4-2.1 处理 |
|-----------|------|---------|
| ENABLE_LLM_AGENT=false | 命名与 LLM_ENABLE 不一致 | 统一为 LLM_ENABLE=false |
| ENABLE_LLM_FALLBACK=true | 命名与 LLM_FALLBACK_ENABLE 不一致 | 统一为 LLM_FALLBACK_ENABLE=true |
| LLM_MAX_RETRIES=1 | Demo 阶段不重试 | 已移除 |
| LLM_MAX_OUTPUT_TOKENS=2048 | 首轮验收表明 2048 不够 | 已改为 8192 |
| LLM_TIMEOUT_SECONDS=20 | 匹配分析 Prompt token 较多 | 已改为 30 |
| LLM_BASE_URL 含真实 URL | .env.example 不应含真实地址 | 已清除为空 |
| LLM_API_KEY 含真实 Key | 🔴 安全问题 | 已清除为空 |
| 缺少 LLM_REPAIR_ENABLE | - | 已新增，默认 true |
| 缺少 LLM_REPAIR_MAX_ATTEMPTS | - | 已新增，固定 1 |
| 缺少 LLM_LOG_RAW_OUTPUT | - | 已新增，默认 false |
| 缺少 LLM_PROMPT_DIR | - | 已新增，默认 server/prompts |
| 缺少 LLM_SCHEMA_DIR | - | 已新增，默认 server/prompts/schema |

安全约束：

1. **LLM_API_KEY 只能放在 .env**
2. **.env.example 只能放空值或占位符**
3. **docs 和开发日志不得记录真实 API Key**
4. **默认支持关闭 LLM，关闭后直接 Mock fallback**

## 8. LLM Client 抽象设计

### 8.1 接口定义方向

```go
// LLMClient 定义 LLM 调用抽象
type LLMClient interface {
    // Chat 发送聊天请求，返回模型原始输出
    Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
}
```

### 8.2 请求结构

```go
// ChatRequest LLM 请求
type ChatRequest struct {
    SystemPrompt string  // 系统 Prompt（业务 Prompt 模板）
    UserPrompt   string  // 用户 Prompt（业务数据上下文）
    Temperature  float64 // 温度
    MaxTokens    int     // 最大输出 token
}

// ChatMessage 聊天消息
type ChatMessage struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}
```

### 8.3 响应结构

```go
// ChatResponse LLM 响应
type ChatResponse struct {
    ID            string    // 请求 ID
    RawOutput     string    // 模型原始输出文本
    FinishReason  string    // 结束原因：stop / length
    Provider      string    // 提供商：hunyuan
    Model         string    // 实际使用的模型名称
    Usage         Usage     // token 使用量
    DurationMS    int64     // 调用耗时（毫秒）
    ErrorCode     string    // 错误码（成功时为空）
    ErrorMessage  string    // 错误信息（成功时为空）
}

// Usage token 使用量
type Usage struct {
    PromptTokens     int `json:"prompt_tokens"`
    CompletionTokens int `json:"completion_tokens"`
    TotalTokens      int `json:"total_tokens"`
}
```

### 8.4 工厂函数

```go
// NewLLMClient 根据 provider 创建对应的 LLM Client
func NewLLMClient(provider string, cfg *config.Config) LLMClient {
    switch provider {
    case "hunyuan":
        return NewHunyuanClient(cfg)
    default:
        return nil // 未配置则返回 nil，使用 Mock Agent
    }
}
```

### 8.5 核心约束

- **业务 Agent 不直接依赖 Hunyuan**
- **业务 Agent 只依赖 LLM Client / Runner 抽象**
- **LLM Client 不感知业务逻辑（不知道匹配分析/推荐/简历优化的区别）**
- **LLM Client 只负责 Chat 调用和 HTTP 层错误分类**

## 9. Hunyuan Client 接入设计

### 9.1 定位

Hunyuan Client 是 LLM Client 的一个 provider 实现。

### 9.2 调用方式

基于 OpenAI-compatible API：

```
POST {LLM_BASE_URL}/chat/completions
Authorization: Bearer {LLM_API_KEY}
Content-Type: application/json

{
  "model": "{LLM_MODEL}",
  "messages": [
    {"role": "system", "content": "业务 Prompt"},
    {"role": "user", "content": "业务数据上下文"}
  ],
  "temperature": {LLM_TEMPERATURE},
  "max_tokens": {LLM_MAX_OUTPUT_TOKENS}
}
```

### 9.3 需要覆盖的场景

| 场景 | 处理方式 |
|------|---------|
| 正常响应 | 提取 choices[0].message.content |
| HTTP 错误（5xx） | 返回 LLM_HTTP_ERROR |
| 认证失败（401/403） | 返回 LLM_API_KEY_MISSING |
| 限流（429） | 返回 LLM_RATE_LIMIT |
| 空响应 | 返回 LLM_EMPTY_RESPONSE |
| 输出截断（finish_reason=length） | 返回 LLM_OUTPUT_TRUNCATED + 截断内容 |
| 超时 | 返回 LLM_TIMEOUT |
| 网络错误 | 返回 LLM_HTTP_ERROR |

### 9.4 安全要求

- **不打印 API Key**
- **不在日志中记录完整 API Key**
- **不将 API Key 写入任何文档或 .env.example**
- **HTTP 请求 Header 中 Authorization 值不写入日志**

## 10. Prompt Loader 设计

### 10.1 定位

Prompt Loader 从 `server/prompts/` 读取 `.md` 文件，提供统一的 Prompt 加载能力。

### 10.2 需要支持

| 能力 | 说明 |
|------|------|
| 按 PromptName 加载 | 如 Load("match_analysis") |
| 校验文件存在 | 不存在时返回明确错误 |
| 替换占位符 | 支持 `{{.Variable}}` 格式 |
| 保留原始模板 | 每次加载不修改源文件 |
| 不把 Prompt 写死在 Go 代码 | 运行时从文件读取 |
| 支持后续版本化 | 文件名含版本号，如 match_analysis_v1.md |

### 10.3 Prompt 映射

| PromptName | 文件路径 |
|-----------|---------|
| match_analysis | server/prompts/match_analysis_v1.md |
| recommendation | server/prompts/recommendation_v1.md |
| resume_optimization | server/prompts/resume_optimization_v1.md |
| json_repair | server/prompts/json_repair_v1.md |

### 10.4 接口方向

```go
// PromptLoader 加载 Prompt 模板
type PromptLoader struct {
    baseDir string // server/prompts/
}

// Load 加载指定名称的 Prompt 模板
func (l *PromptLoader) Load(name string) (string, error)

// LoadWithVars 加载模板并替换占位符
func (l *PromptLoader) LoadWithVars(name string, vars map[string]string) (string, error)
```

### 10.5 实现方式

- 启动时可通过 `embed` 嵌入或 `os.ReadFile` 运行时读取
- 推荐使用 `os.ReadFile`，便于开发期热更新
- Prompt 文件路径基于配置或相对于可执行文件

## 11. Schema Loader 设计

### 11.1 定位

Schema Loader 从 `server/prompts/schema/` 读取 JSON Schema 文件。

### 11.2 Schema 映射

| SchemaName | 文件路径 |
|-----------|---------|
| match_analysis | server/prompts/schema/match_analysis_schema.json |
| recommendation | server/prompts/schema/recommendation_schema.json |
| resume_optimization | server/prompts/schema/resume_optimization_schema.json |

### 11.3 特殊说明

**json_repair 不需要独立业务输出 Schema**，它使用目标业务 Schema 作为修复目标。

### 11.4 接口方向

```go
// SchemaLoader 加载 JSON Schema
type SchemaLoader struct {
    baseDir string // server/prompts/schema/
}

// Load 加载指定名称的 JSON Schema
func (l *SchemaLoader) Load(name string) ([]byte, error)
```

## 12. JSON parse 与 Schema validate 设计

### 12.1 处理链路

```
raw output（模型原始输出文本）
  → 去除 Markdown code fence（```json ... ```）
  → 提取 JSON object
  → json.Unmarshal
    → 如果失败：进入 JSON repair
  → Schema validate
    → 如果失败：进入 JSON repair
  → 通过：返回 LLM 结果
```

### 12.2 JSON 提取

必须覆盖：

1. **去除 Markdown code fence**：` ```json ... ``` ` 包裹
2. **去除解释性前后缀**：模型可能在 JSON 前后加说明文字
3. **提取 JSON object**：找到第一个 `{` 和最后一个 `}` 之间的内容
4. **去除 BOM / 不可见字符**

### 12.3 Schema validate

需要讨论 Go JSON Schema 校验库选择。

### 12.4 候选 JSON Schema 校验库

| 候选库 | 优点 | 缺点 | 推荐度 |
|--------|------|------|--------|
| github.com/santhosh-tekuri/jsonschema/v5 | 支持 draft-04/06/07/2020-12，纯 Go 实现，活跃维护 | 较新版本 | ⭐⭐⭐ 推荐 |
| github.com/xeipuuv/gojsonschema | 历史较长，使用广泛 | 不支持 draft 2020-12，维护缓慢 | ⭐⭐ 备选 |

**推荐选择**：`github.com/santhosh-tekuri/jsonschema/v5`

推荐理由：

1. 当前 Schema 文件使用 `$schema: "https://json-schema.org/draft/2020-12/schema"`，需要支持 draft 2020-12
2. 纯 Go 实现，无 CGO 依赖
3. 活跃维护，API 简洁
4. 项目 Go 版本为 1.25.0，完全兼容

**但本阶段只提出候选方案，4-2.3 才实际引入依赖。**

### 12.5 校验失败分类

| 校验类型 | 错误码 | 是否进入 repair |
|---------|--------|----------------|
| JSON 解析失败 | JSON_PARSE_FAILED | ✅ |
| Markdown 包裹 | JSON_MARKDOWN_WRAPPED | ✅ |
| Schema 校验失败 | SCHEMA_VALIDATION_FAILED | ✅ |
| 必填字段缺失 | SCHEMA_REQUIRED_FIELD_MISSING | ✅ |
| 枚举不合法 | SCHEMA_ENUM_INVALID | ✅ |
| 类型错误 | SCHEMA_TYPE_INVALID | ✅ |
| 额外字段 | SCHEMA_ADDITIONAL_PROPERTY | ✅ |

## 13. JSONRepairAgent 调用链路设计

### 13.1 Prompt 来源

复用 `server/prompts/json_repair_v1.md`。

### 13.2 调用输入

| 变量 | 说明 |
|------|------|
| TARGET_OUTPUT_TYPE | match_analysis / recommendation / resume_optimization |
| TARGET_SCHEMA_JSON | 对应业务 Schema 的 JSON 字符串 |
| BROKEN_MODEL_OUTPUT | 原始异常输出 |
| VALIDATION_ERRORS | 解析/校验错误信息 |
| REPAIR_CONTEXT | 可选上下文（resume_id、job_id 等） |

### 13.3 核心规则

1. **JSON repair 只调用一次**
2. **repair 不重新做业务分析**
3. **repair 成功后 generated_by 仍为 llm_agent**
4. **repair 失败后 Mock fallback**
5. **不递归 repair**
6. **不无限重试**
7. **repair 只修复结构，不编造业务事实**

### 13.4 调用方式

JSONRepairAgent 本身也需要调用 LLM Client，使用 `json_repair` Prompt 模板。

```
JSONRepairAgent 调用链路：
→ PromptLoader.Load("json_repair")
→ 替换占位符（TARGET_OUTPUT_TYPE / TARGET_SCHEMA_JSON / BROKEN_MODEL_OUTPUT / VALIDATION_ERRORS）
→ LLMClient.Chat(ctx, repairReq)
→ 提取 repair output
→ JSON parse
→ Schema validate（使用目标业务 Schema）
→ 通过：返回 repaired 结果
→ 失败：标记 repair_failed，进入 Mock fallback
```

### 13.5 实现位置建议

```
server/internal/agent/repair/
  json_repair_agent.go
```

## 14. Mock fallback 集成设计

### 14.1 定位

Mock fallback 是最终稳定性保障。

### 14.2 直接触发 Mock 的场景

| 场景 | 错误码 |
|------|--------|
| LLM 被禁用 | LLM_DISABLED |
| LLM 配置缺失 | LLM_CONFIG_MISSING |
| API Key 缺失 | LLM_API_KEY_MISSING |
| 请求超时 | LLM_TIMEOUT |
| HTTP 错误 | LLM_HTTP_ERROR |
| 限流 | LLM_RATE_LIMIT |
| 空响应 | LLM_EMPTY_RESPONSE |
| 上下文过长 | LLM_CONTEXT_TOO_LONG |
| 未知错误 | UNKNOWN_ERROR |

### 14.3 先 repair 再 Mock 的场景

| 场景 | 错误码 |
|------|--------|
| 输出截断 | LLM_OUTPUT_TRUNCATED |
| JSON 解析失败 | JSON_PARSE_FAILED |
| Markdown 包裹 | JSON_MARKDOWN_WRAPPED |
| Schema 校验失败 | SCHEMA_VALIDATION_FAILED |
| 必填字段缺失 | SCHEMA_REQUIRED_FIELD_MISSING |
| 枚举不合法 | SCHEMA_ENUM_INVALID |
| 类型错误 | SCHEMA_TYPE_INVALID |
| 额外字段 | SCHEMA_ADDITIONAL_PROPERTY |

### 14.4 边界说明

- **4-2 只设计和实现通用 fallback decision**
- **4-3 才把具体业务 Mock Agent 接入业务链路**
- fallback decision 只返回"应使用 Mock"，不直接调用具体 Mock Agent

## 15. generated_by / agent_status / fallback 标记设计

### 15.1 generated_by 约定

| 值 | 含义 |
|----|------|
| llm_agent | LLM 生成（含 repair 成功的） |
| mock_agent | Mock Agent 生成 |

不使用的值：

- ❌ `llm_agent_fallback`：fallback 状态由外层 agent_status 表达
- ❌ `json_repair_agent`：repair 不改变业务结果来源
- ❌ `repair_agent`：同上

### 15.2 后续外层 meta 建议

```json
{
  "agent_status": "success | repaired | fallback",
  "fallback_reason": "",
  "error_code": "",
  "repair_applied": false,
  "llm_enabled": true
}
```

说明：如果当前接口暂时没有 meta 字段，4-2 只设计，4-4 再统一改接口返回结构。

## 16. 日志与安全策略

### 16.1 日志记录字段

| 字段 | 说明 |
|------|------|
| request_id | 请求唯一 ID |
| agent_name | 业务 Agent 名称 |
| prompt_name | Prompt 模板名称 |
| schema_name | Schema 名称 |
| provider | 模型提供商 |
| model | 实际使用的模型 |
| llm_enabled | LLM 是否启用 |
| call_duration_ms | 调用耗时（毫秒） |
| raw_output_length | 原始输出长度 |
| json_parse_success | JSON 解析是否成功 |
| schema_validate_success | Schema 校验是否成功 |
| repair_attempted | 是否尝试 repair |
| repair_success | repair 是否成功 |
| fallback_used | 是否使用 fallback |
| fallback_reason | fallback 原因 |
| error_code | 错误码 |

### 16.2 禁止记录

1. **API Key**
2. **完整简历**
3. **完整 JD**
4. **完整模型原始输出**（只记录长度和摘要）
5. **用户敏感信息**

### 16.3 可选记录（调试模式）

- LLM_LOG_RAW_OUTPUT=true 时记录原始输出
- 仅用于开发调试，不得在生产环境开启

## 17. 错误码设计

| 错误码 | 触发阶段 | 处理方式 | 进入 repair | Mock fallback | 用户可见 |
|--------|---------|---------|------------|--------------|---------|
| LLM_DISABLED | 配置检查 | 直接 Mock | ❌ | ✅ | ❌ |
| LLM_CONFIG_MISSING | 配置检查 | 直接 Mock | ❌ | ✅ | ❌ |
| LLM_API_KEY_MISSING | 配置检查 | 直接 Mock | ❌ | ✅ | ❌ |
| LLM_TIMEOUT | LLM 调用 | 直接 Mock | ❌ | ✅ | ❌ |
| LLM_HTTP_ERROR | LLM 调用 | 直接 Mock | ❌ | ✅ | ❌ |
| LLM_RATE_LIMIT | LLM 调用 | 直接 Mock | ❌ | ✅ | ❌ |
| LLM_EMPTY_RESPONSE | LLM 调用 | 直接 Mock | ❌ | ✅ | ❌ |
| LLM_OUTPUT_TRUNCATED | LLM 调用 | 尝试 repair | ✅ | 看结果 | ❌ |
| PROMPT_LOAD_FAILED | Prompt 加载 | 直接 Mock | ❌ | ✅ | ❌ |
| SCHEMA_LOAD_FAILED | Schema 加载 | 直接 Mock | ❌ | ✅ | ❌ |
| JSON_PARSE_FAILED | JSON 解析 | 尝试 repair | ✅ | 看结果 | ❌ |
| SCHEMA_VALIDATION_FAILED | Schema 校验 | 尝试 repair | ✅ | 看结果 | ❌ |
| JSON_REPAIR_FAILED | JSON Repair | 直接 Mock | ❌ | ✅ | ❌ |
| MOCK_FALLBACK_USED | 兜底 | 正常返回 | ❌ | ✅ | ❌ |
| UNKNOWN_ERROR | 未知 | 直接 Mock | ❌ | ✅ | ❌ |

## 18. Makefile / 脚本验收入口设计

本阶段只设计，不修改 Makefile。

### 18.1 建议后续增加

| 命令 | 说明 |
|------|------|
| make llm-check | 检查 LLM 配置是否完整 |
| make llm-smoke | 发送一次简单 LLM 调用验证连通性 |
| make prompt-check | 检查所有 Prompt 文件是否存在且非空 |
| make schema-check | 检查所有 Schema 文件是否合法 JSON |
| make llm-repair-smoke | 验证 JSON repair 一次调用 |

或使用脚本：

```
scripts/dev/llm_smoke.sh
scripts/dev/prompt_check.sh
scripts/dev/schema_check.sh
```

4-2.6 已完成 smoke test 和 Makefile 接入。

## 19. 后续子任务拆分

| 编号 | 任务 | 说明 |
|------|------|------|
| 4-2.0 | LLM Client 与 Hunyuan 接入方案 | docs only（本次任务） |
| 4-2.1 | LLM / Hunyuan 环境变量配置项 | 修改 config.go，补齐配置字段，更新 .env.example |
| 4-2.2 | Prompt Loader / Schema Loader | 实现从 server/prompts/ 和 schema/ 加载文件 |
| 4-2.3 | JSON parse + Schema validate 工具 | 实现 JSON 提取、parse、Schema 校验 |
| 4-2.4 | LLM Client 抽象接口 | 实现 LLMClient interface + types + 工厂函数 |
| 4-2.5 | Hunyuan Client 实现 | 基于 OpenAI-compatible API 实现 Hunyuan Provider |
| 4-2.6 | LLM Smoke Test / Makefile 接入 | 实现 llm-smoke / prompt-check / schema-check |
| 4-2.7 | JSONRepairAgent 调用链路 | 实现 repair once + Mock fallback decision |
| 4-2.8 | 统一 LLM Agent Pipeline | 实现 LLMRunner：Chat → parse → validate → repair → fallback |
| 4-2.9 | 阶段 4-2 集成验收 | 验证所有基础设施能力可用 |
| 4-2.10 | 阶段 4-2 汇总 | 汇总 LLM Client 成果 |

## 20. 验收标准

4-2 完成时应满足：

| 序号 | 验收标准 |
|------|---------|
| 1 | LLM 配置可读取 |
| 2 | Prompt 可从文件加载 |
| 3 | Schema 可从文件加载 |
| 4 | Hunyuan 可通过 LLM Client 调用 |
| 5 | raw output 可提取 JSON |
| 6 | JSON 可 parse |
| 7 | Schema 可 validate |
| 8 | JSON repair 可调用一次 |
| 9 | repair 失败可进入 fallback decision |
| 10 | 不泄露 API Key |
| 11 | 不修改业务 Agent 接口逻辑 |
| 12 | 为 4-3 接入业务 Agent 做好准备 |

## 21. 当前不做事项

本阶段 4-2.0 不：

- 不写 Go 代码
- 不修改 .env.example
- 不实现 Client
- 不调用 Hunyuan
- 不修改 Prompt
- 不修改 Schema
- 不接业务 Agent
- 不改接口
- 不改前端

## 22. 下一步建议

**下一步进入 4-2.1：LLM / Hunyuan 环境变量配置项**

4-2.1 才允许修改 .env.example 和环境变量文档；4-2.2 才开始实现 Prompt Loader / Schema Loader；4-2.4 / 4-2.5 才开始实现 LLM Client / Hunyuan Client。

4-2.3 已完成：JSON parse + Schema validate 工具已实现。
当前已具备从 LLM raw output 提取 JSON、解析 JSON、执行 Schema validate 的基础能力。
- llmjson 包：从 raw output 提取 JSON object（支持 Markdown code fence、前后解释文字），JSON parse，错误码（JSON_EMPTY / JSON_OBJECT_NOT_FOUND / JSON_PARSE_FAILED / JSON_TOP_LEVEL_NOT_OBJECT）
- schema 包新增 Validator：使用 santhosh-tekuri/jsonschema/v5 实现 JSON Schema 校验（支持 draft 2020-12），返回结构化 ValidationResult

4-2.4 已完成：LLM Client 抽象接口已实现。
当前已具备统一模型调用抽象和基础类型定义：
- llm 包：Client 接口（Chat + Provider）、ChatRequest / ChatResponse / Message / Usage / Provider / Role
- llm 包：Config 结构体（覆盖 .env.example 全部 14 项 LLM 配置）、DefaultConfig()、Config.Validate()
- llm 包：ErrorCode（LLM_DISABLED / LLM_CONFIG_MISSING / LLM_API_KEY_MISSING / LLM_TIMEOUT / LLM_HTTP_ERROR / LLM_RATE_LIMIT / LLM_EMPTY_RESPONSE / LLM_OUTPUT_TRUNCATED / UNKNOWN_ERROR）
- llm 包：LLMError（结构化错误 + Unwrap + IsErrorCode + IsDirectMockFallback）
- llm 包：NoopClient（不访问外部网络的测试实现）
- 单元测试 27 项全部通过

当前仍未进行真实 Hunyuan smoke test。

4-2.5 已完成：Hunyuan Client 实现已完成。
当前已具备 Hunyuan / TokenHub OpenAI-compatible chat completions 的 HTTP 调用封装能力：
- HunyuanClient 实现 llm.Client 接口（Chat + Provider）
- 支持通用 ChatRequest 到 OpenAI-compatible 请求体的映射
- 支持 OpenAI-compatible 响应体到 ChatResponse 的解析
- 支持 BaseURL endpoint 规范化（/v1、/v1/、/v1/chat/completions 均可）
- 支持 Authorization Bearer Token
- 支持 per-request timeout 和 context cancellation
- 支持 HTTP status 错误分类（429→LLM_RATE_LIMIT、5xx→LLM_HTTP_ERROR、401/403→LLM_API_KEY_MISSING）
- 支持空响应错误分类（空 body、空 choices、空 content → LLM_EMPTY_RESPONSE）
- 支持 finish_reason=length 记录（不视为硬错误，由 Pipeline 判断是否 repair）
- 支持 usage 解析
- 不在错误中泄露 API Key
- 使用 httptest 进行单元测试，不调用真实 Hunyuan
- 单元测试 22 项全部通过

下一步为 4-2.7：JSONRepairAgent 调用链路

4-2.6 已完成：LLM Smoke Test / Makefile 接入。
当前已具备通过 `make llm-check` 检查 LLM 配置、通过 `RUN_REAL_LLM_SMOKE=1 make llm-smoke` 执行最小 HunyuanClient 连通性验证的能力。
- 新增 `server/cmd/llm-smoke/main.go`：最小 smoke test Go 命令
- 新增 `scripts/dev/llm_smoke.sh`：LLM smoke test 脚本（help / check / smoke）
- 新增 Makefile `llm-help` / `llm-check` / `llm-smoke` target
- `make llm-check` 不请求外部模型
- `make llm-smoke` 需要 `RUN_REAL_LLM_SMOKE=1` 才执行真实调用
- smoke test 使用最小 Prompt，不携带简历、JD 或业务数据
- 不输出 API Key，不输出完整 raw output
- 当前未接业务 Agent
