# 2026-05-30-069-JSONParse与SchemaValidate工具

> 创建时间：2026-05-30
> 所属阶段：4-2.3
> 任务类型：后端工程实现

## 1. 本次任务

阶段 4-2.3：JSON parse + Schema validate 工具。

目标：让 Go 后端具备从 LLM raw output 中提取 JSON、解析 JSON、并基于既有 Schema 执行结构校验的能力。

## 2. 修改范围

- 新增 `server/internal/llmjson/` 包（types.go、extractor.go、parser.go、extractor_test.go、parser_test.go）
- 新增 `server/internal/schema/validator.go` 和 `validator_test.go`
- 修改 `server/go.mod` 和 `server/go.sum`（新增 santhosh-tekuri/jsonschema/v5 依赖）
- 更新 `docs/05-Agent设计/` 下相关文档
- 更新 `docs/项目状态.md`、`docs/下一步任务.md`、`docs/变更记录.md`、`docs/说明书.md`

## 3. JSON 提取与解析工具

### 包

`server/internal/llmjson/`

### 文件

| 文件 | 说明 |
|------|------|
| types.go | ErrorCode 类型、4 个常量、ParseResult 结构体 |
| extractor.go | ExtractObject / ExtractBestCandidate 函数 |
| parser.go | ParseObject / ExtractAndParse 函数 |

### 支持的输入场景

| 场景 | 处理方式 |
|------|---------|
| 纯 JSON object | 直接返回 |
| ```json code fence | 提取 code fence 内容 |
| 无语言标识 code fence | 提取 code fence 内容 |
| 前后带解释文字 | 通过花括号匹配提取 |
| 多行 JSON | 正常处理 |
| BOM 前缀 | 自动清除 |
| 空字符串 | 返回 JSON_EMPTY |
| 无 JSON object | 返回 JSON_OBJECT_NOT_FOUND |
| 非法 JSON | 返回 JSON_PARSE_FAILED |
| 顶层 array | 返回 JSON_TOP_LEVEL_NOT_OBJECT |

### 错误码设计

| ErrorCode | 含义 |
|-----------|------|
| JSON_EMPTY | raw output 为空 |
| JSON_OBJECT_NOT_FOUND | raw output 中未找到 JSON object |
| JSON_PARSE_FAILED | 提取的 JSON 字符串无法解析 |
| JSON_TOP_LEVEL_NOT_OBJECT | 顶层 JSON 值为 array 而非 object |

### 核心函数

- `ExtractObject(raw string) (string, error)`：从 raw output 提取 JSON 字符串
- `ExtractBestCandidate(raw string) (string, error)`：提取最可能的 JSON 候选（即使无效），供 repair 使用
- `ParseObject(raw string) (map[string]any, error)`：提取并解析 JSON object
- `ExtractAndParse(raw string) *ParseResult`：合并提取与解析，返回结构化结果

## 4. Schema validate 工具

### 包

`server/internal/schema/`（在现有包上扩展）

### 新增文件

| 文件 | 说明 |
|------|------|
| validator.go | Validator 结构体、Validate / ValidateJSON 方法、ValidationResult / ValidationError 类型 |
| validator_test.go | 9 项单元测试（3 个合法对象 + 4 种失败场景 + ValidateJSON） |

### 核心方法

- `NewValidator() *Validator`：创建 Validator
- `Validator.Validate(doc *Document, data any) (*ValidationResult, error)`：校验数据是否符合 Schema
- `Validator.ValidateJSON(doc *Document, jsonBytes []byte) (*ValidationResult, error)`：从 JSON bytes 解析后校验

### ValidationResult 结构

| 字段 | 说明 |
|------|------|
| Success | 是否通过校验 |
| SchemaName | 使用的 Schema 逻辑名称 |
| ErrorCode | 失败时的错误码（SCHEMA_VALIDATION_FAILED / JSON_PARSE_FAILED） |
| ErrorMessage | 摘要错误信息 |
| Errors | 详细错误列表（Path + Message） |

### 与 4-2.2 Schema Loader 的关系

复用 4-2.2 的 `schema.Loader` 加载 Schema 文件，Validator 接收 `*Document` 作为输入。

## 5. 错误码设计

### llmjson 错误码

| ErrorCode | 触发条件 | 后续处理 |
|-----------|---------|---------|
| JSON_EMPTY | raw output 为空 | 直接 Mock fallback |
| JSON_OBJECT_NOT_FOUND | 无 JSON object | 进入 JSON repair |
| JSON_PARSE_FAILED | JSON 解析失败 | 进入 JSON repair |
| JSON_TOP_LEVEL_NOT_OBJECT | 顶层为 array | 进入 JSON repair |

### schema validate 错误码

| ErrorCode | 触发条件 | 后续处理 |
|-----------|---------|---------|
| JSON_PARSE_FAILED | JSON bytes 无法解析 | 进入 JSON repair |
| SCHEMA_VALIDATION_FAILED | Schema 校验失败 | 进入 JSON repair |

## 6. 测试结果

### llmjson 包测试（21 项，全部通过）

| 测试 | 说明 |
|------|------|
| TestExtractObject_PureJSON | 纯 JSON object 成功 |
| TestExtractObject_MarkdownJSONFence | ```json code fence 成功 |
| TestExtractObject_CodeFenceNoLanguage | 无语言 code fence 成功 |
| TestExtractObject_SurroundingText | 前后带解释文字成功 |
| TestExtractObject_MultiLineJSON | 多行 JSON 成功 |
| TestExtractObject_EmptyString | 空字符串返回 JSON_EMPTY |
| TestExtractObject_NoJSONObject | 无 JSON 返回 JSON_OBJECT_NOT_FOUND |
| TestExtractObject_BOMPrefixed | BOM 前缀自动清除 |
| TestExtractObject_NestedJSON | 嵌套 JSON 正常处理 |
| TestParseObject_PureJSON | 纯 JSON 解析成功 |
| TestParseObject_MarkdownFence | Markdown 包裹解析成功 |
| TestParseObject_SurroundingText | 带文字解析成功 |
| TestParseObject_InvalidJSON | 非法 JSON 返回 JSON_PARSE_FAILED |
| TestParseObject_TopLevelArray | 顶层 array 返回 JSON_TOP_LEVEL_NOT_OBJECT |
| TestParseObject_EmptyString | 空字符串解析报错 |
| TestExtractAndParse_Success | 合并提取解析成功 |
| TestExtractAndParse_Empty | 空字符串返回 JSON_EMPTY |
| TestExtractAndParse_NoJSONObject | 无 JSON 返回 JSON_OBJECT_NOT_FOUND |
| TestExtractAndParse_InvalidJSON | 非法 JSON 返回 JSON_PARSE_FAILED |
| TestExtractAndParse_TopLevelArray | 顶层 array 返回 JSON_TOP_LEVEL_NOT_OBJECT |
| TestExtractAndParse_MarkdownFence | Markdown 包裹提取解析成功 |

### schema validator 测试（9 项，全部通过）

| 测试 | 说明 |
|------|------|
| TestValidator_Validate_MatchAnalysis_Valid | match_analysis 最小合法对象校验通过 |
| TestValidator_Validate_Recommendation_Valid | recommendation 最小合法对象校验通过 |
| TestValidator_Validate_ResumeOptimization_Valid | resume_optimization 最小合法对象校验通过 |
| TestValidator_Validate_MissingRequiredField | 缺少 required 字段校验失败 |
| TestValidator_Validate_EnumInvalid | 枚举不合法校验失败 |
| TestValidator_Validate_AdditionalProperties | additionalProperties 校验失败 |
| TestValidator_Validate_TypeError | 类型错误校验失败 |
| TestValidator_ValidateJSON_Valid | ValidateJSON 合法对象校验通过 |
| TestValidator_ValidateJSON_InvalidJSON | ValidateJSON 非法 JSON 报错 |

## 7. 第三方依赖说明

新增依赖：`github.com/santhosh-tekuri/jsonschema/v5 v5.3.1`

选择原因：
1. 项目的 3 个 Schema 文件均使用 `$schema: "https://json-schema.org/draft/2020-12/schema"`，需要支持 draft 2020-12
2. santhosh-tekuri/jsonschema/v5 是纯 Go 实现，无 CGO 依赖
3. 活跃维护，API 简洁
4. 项目 Go 版本为 1.25.0，完全兼容
5. 备选方案 xeipuuv/gojsonschema 不支持 draft 2020-12，维护缓慢

## 8. 文档更新

- 更新阶段 4-2 LLM Client 与 Hunyuan 接入方案
- 更新 JSON 输出规范
- 更新 LLM Client 设计
- 更新 Hunyuan 接入准备
- 更新失败兜底与 Mock 回退策略
- 更新 Agent 总览
- 更新 Prompt 模板
- 更新项目状态
- 更新下一步任务
- 更新变更记录
- 更新说明书

## 9. 未修改内容

- 未修改 .env
- 未修改 .env.example
- 未修改 Prompt 文件
- 未修改 Schema 文件
- 未修改 config.go
- 未修改业务 Agent
- 未修改 Makefile
- 未修改 scripts/dev/
- 未修改前端
- 未调用 Hunyuan
- 未请求腾讯招聘接口
- 未实现 LLM Client
- 未实现 Hunyuan Client
- 未实现 JSONRepairAgent 调用
- 未实现 Mock fallback

## 10. 当前风险

| 风险 | 严重度 | 说明 |
|------|--------|------|
| config.go 尚未读取 LLM_* 配置项 | 🟡 中 | 后续 4-2.4+ 逐步实现 |
| santhosh-tekuri/jsonschema 的错误路径格式 | 🟢 低 | 使用 InstanceLocation 字段，需后续确认与 JSON repair 的对接 |
| Schema 校验失败时错误信息可能过多 | 🟢 低 | 后续可根据需要裁剪或摘要 |

## 11. 是否可以进入 4-2.4

✅ 可以进入 4-2.4。

JSON parse + Schema validate 工具已实现并通过测试，后续 4-2.4 可以基于这些工具构建 LLM Client 抽象接口。

## 12. 下一步建议

1. **4-2.4**：LLM Client 抽象接口
2. **4-2.5**：Hunyuan Client 实现
3. **4-2.6**：LLM Smoke Test / Makefile 接入
