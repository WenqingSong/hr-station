# 2026-05-29-068-PromptLoader与SchemaLoader

> 创建时间：2026-05-29
> 所属阶段：4-2.2
> 任务类型：后端工程实现

## 1. 本次任务

阶段 4-2.2：Prompt Loader / Schema Loader。

目标：让 Go 后端具备从 `server/prompts/` 加载 Prompt 文件、从 `server/prompts/schema/` 加载 Schema 文件的基础能力。

## 2. 修改范围

- 新增 `server/internal/prompt/` 包（types.go、loader.go、loader_test.go）
- 新增 `server/internal/schema/` 包（types.go、loader.go、loader_test.go）
- 更新 `docs/05-Agent设计/` 下相关文档
- 更新 `docs/项目状态.md`、`docs/下一步任务.md`、`docs/变更记录.md`、`docs/说明书.md`

## 3. Prompt Loader 实现

### 包

`server/internal/prompt/`

### 文件

| 文件 | 说明 |
|------|------|
| types.go | PromptName 类型、4 个常量、Template 结构体 |
| loader.go | Loader 结构体、NewLoader / Load / Render / List / ResolvePath 方法 |
| loader_test.go | 10 项单元测试 |

### 支持的逻辑名称

| PromptName | 文件 |
|-----------|------|
| match_analysis | match_analysis_v1.md |
| recommendation | recommendation_v1.md |
| resume_optimization | resume_optimization_v1.md |
| json_repair | json_repair_v1.md |

### 核心方法

- `NewLoader(baseDir string) *Loader`：创建 Loader
- `Load(name PromptName) (*Template, error)`：加载 Prompt 模板
- `Render(name PromptName, values map[string]string) (string, error)`：加载模板并替换 `{{KEY}}` 占位符
- `List() []PromptName`：列出所有已知逻辑名称
- `ResolvePath(name PromptName) (string, error)`：解析文件路径

### 占位符规则

- 格式：`{{KEY}}`
- Render 只做字符串替换
- 不对 JSON 内容做二次转义
- 未提供的占位符保留原样，不报错

## 4. Schema Loader 实现

### 包

`server/internal/schema/`

### 文件

| 文件 | 说明 |
|------|------|
| types.go | SchemaName 类型、3 个常量、Document 结构体 |
| loader.go | Loader 结构体、NewLoader / Load / List / ResolvePath 方法、Document 元信息方法 |
| loader_test.go | 11 项单元测试 |

### 支持的逻辑名称

| SchemaName | 文件 |
|-----------|------|
| match_analysis | match_analysis_schema.json |
| recommendation | recommendation_schema.json |
| resume_optimization | resume_optimization_schema.json |

### 核心方法

- `NewLoader(baseDir string) *Loader`：创建 Loader
- `Load(name SchemaName) (*Document, error)`：加载 Schema 文件并解析 JSON
- `List() []SchemaName`：列出所有已知逻辑名称
- `ResolvePath(name SchemaName) (string, error)`：解析文件路径
- `Document.Title()`：读取 Schema 的 `title` 字段
- `Document.RequiredCount()`：读取 `required` 数组长度
- `Document.Draft()`：读取 `$schema` URI

### JSON 合法性检查

- 使用 `encoding/json` 解析
- 顶层必须是 object（`map[string]any`）
- 不做完整 Schema validate
- 不引入第三方依赖

## 5. 路径安全设计

- 只允许预定义的逻辑名称（PromptName / SchemaName 常量）
- 不允许外部直接传文件路径
- ResolvePath 基于 baseDir + 映射文件名构建
- 使用 `filepath.Abs` 验证路径不穿越 baseDir
- 未知逻辑名称返回明确错误

## 6. 测试结果

### Prompt Loader 测试（10 项，全部通过）

| 测试 | 说明 |
|------|------|
| TestLoader_Load_MatchAnalysis | 能加载 match_analysis_v1.md |
| TestLoader_Load_Recommendation | 能加载 recommendation_v1.md |
| TestLoader_Load_ResumeOptimization | 能加载 resume_optimization_v1.md |
| TestLoader_Load_JSONRepair | 能加载 json_repair_v1.md |
| TestLoader_Render | 占位符替换正常 |
| TestLoader_Render_MissingPlaceholder | 未提供占位符保留原样 |
| TestLoader_List | 返回 4 个 PromptName |
| TestLoader_ResolvePath | 路径解析正确 |
| TestLoader_Load_UnknownName | 未知名称返回错误 |
| TestLoader_ResolvePath_UnknownName | 未知名称返回错误 |

### Schema Loader 测试（11 项，全部通过）

| 测试 | 说明 |
|------|------|
| TestLoader_Load_MatchAnalysis | 能加载 match_analysis_schema.json |
| TestLoader_Load_Recommendation | 能加载 recommendation_schema.json |
| TestLoader_Load_ResumeOptimization | 能加载 resume_optimization_schema.json |
| TestLoader_Load_AllSchemasAreValidJSON | 3 个 Schema 都是合法 JSON |
| TestLoader_Load_Title | 能读取 title |
| TestLoader_Load_RequiredCount | 能读取 required 字段数量 |
| TestLoader_Load_UnknownName | 未知名称返回错误 |
| TestLoader_List | 返回 3 个 SchemaName |
| TestLoader_Load_InvalidJSON | 非法 JSON 返回错误 |
| TestLoader_ResolvePath_UnknownName | 未知名称返回错误 |
| TestDocument_Draft | 能读取 $schema URI |

## 7. 文档更新

- 更新阶段 4-2 LLM Client 与 Hunyuan 接入方案
- 更新 Prompt 模板文档
- 更新 JSON 输出规范
- 更新 LLM Client 设计
- 更新 Hunyuan 接入准备
- 更新 Agent 总览
- 更新失败兜底与 Mock 回退策略（时间戳）
- 更新项目状态
- 更新下一步任务
- 更新变更记录
- 更新说明书

## 8. 未修改内容

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

## 9. 当前风险

| 风险 | 严重度 | 说明 |
|------|--------|------|
| config.go 尚未读取 LLM_* 配置项 | 🟡 中 | 后续 4-2.4+ 逐步实现 |
| 占位符格式 {{KEY}} 较简单 | 🟢 低 | 当前满足需求，后续可升级为 Go template |
| Schema Loader 不做完整 validate | 🟢 低 | 4-2.3 实现 |

## 10. 是否可以进入 4-2.3

✅ 可以进入 4-2.3。

Prompt Loader 和 Schema Loader 已实现并通过测试，后续 4-2.3 可以基于 Schema Loader 实现 JSON parse + Schema validate 工具。

## 11. 下一步建议

1. **4-2.3**：JSON parse + Schema validate 工具
2. **4-2.4**：LLM Client 抽象接口
3. **4-2.5**：Hunyuan Client 实现
