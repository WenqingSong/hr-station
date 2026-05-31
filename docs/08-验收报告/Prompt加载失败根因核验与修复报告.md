# Prompt 加载失败根因核验与修复报告

> 验收时间：2026-05-31
> 验收阶段：4-5.3-R Prompt 加载失败根因核验与最小修复
> 验收结论：**pass**

## 1. 背景

4-5.3 三类 Agent 接口联调验收主体通过，但判定为 `partial_pass`，原因是三类接口均返回 `fallback_reason=prompt_load_failed`。这表明接口契约没问题，但 Prompt 加载链路存在疑点，不属于预期的 API Key 占位符导致的 fallback。

本阶段目标：定位 `prompt_load_failed` 的根因，做最小修复，使三类接口不再因 Prompt 加载失败而 fallback。

## 2. 问题现象

4-5.3 联调时，三类 Agent 接口均返回：

```json
{
  "data": {
    "meta": {
      "agent_status": "fallback",
      "fallback_reason": "prompt_load_failed",
      "error_code": "PROMPT_LOAD_FAILED"
    }
  }
}
```

## 3. 核验范围

1. 三类 Prompt 文件是否存在
2. Prompt loader 的路径策略
3. Schema loader 的路径策略
4. `make up` / `backend.sh` 启动时 server 的实际 cwd
5. 配置默认 PromptDir / SchemaDir 是否与 cwd 匹配
6. `prompt_load_failed` 在代码中的来源

## 4. 当前分支与环境

| 项目 | 值 |
|------|------|
| 主线分支 | `main`（commit `4dd67a8`） |
| 当前分支 | `safeguard/4-4.2-response-contract`（HEAD `0f7fb30`，比 main 多 5 个 docs commit） |
| 工作区状态 | 4 个 Go 文件修改 + 新增 docs |

## 5. Prompt 文件存在性核验

```
$ find server/prompts -maxdepth 3 -type f | sort
server/prompts/json_repair_v1.md
server/prompts/match_analysis_v1.md
server/prompts/recommendation_v1.md
server/prompts/resume_optimization_v1.md
server/prompts/schema/match_analysis_schema.json
server/prompts/schema/recommendation_schema.json
server/prompts/schema/resume_optimization_schema.json
```

**结论：4 个 Prompt 文件 + 3 个 Schema 文件全部存在。**

## 6. Prompt loader 路径策略核验

### 6.1 配置默认值

`server/internal/config/config.go` 第 135-136 行：

```go
PromptDir:  getEnv("LLM_PROMPT_DIR", "server/prompts"),
SchemaDir:  getEnv("LLM_SCHEMA_DIR", "server/prompts/schema"),
```

`.env.example` 设置：

```
LLM_PROMPT_DIR=server/prompts
LLM_SCHEMA_DIR=server/prompts/schema
```

### 6.2 Prompt Loader 实现

`server/internal/prompt/loader.go`：

```go
func NewLoader(baseDir string) *Loader {
    return &Loader{baseDir: baseDir}
}
```

Loader 直接使用 `baseDir`，不做路径解析。

### 6.3 Schema Loader 实现

`server/internal/schema/loader.go`：同上，直接使用 `baseDir`。

### 6.4 `prompt_load_failed` 来源

`server/internal/agent/pipeline/pipeline.go` 第 73-76 行：

```go
renderedPrompt, err := r.promptLoader.Render(req.PromptName, req.PromptValues)
if err != nil {
    return fallbackResult(FallbackPromptLoadFailed, ErrPromptLoadFailed, ...)
}
```

当 `promptLoader.Render` 失败时（即 Prompt 文件不存在），返回 `prompt_load_failed`。

## 7. make up 启动目录核验

### 7.1 backend.sh

`scripts/dev/backend.sh` 第 47-48 行：

```bash
cd "${PROJECT_ROOT}/server"
SERVER_PORT="$SERVER_PORT" nohup go run ./cmd/server >> "$LOG_FILE" 2>&1 &
```

**关键发现：`cd "${PROJECT_ROOT}/server"` 后运行 `go run ./cmd/server`，因此 server 的 cwd 是 `/workspace/server/`。**

### 7.2 路径解析

当 cwd 为 `/workspace/server/` 时：

| 配置路径 | 实际解析 | 是否存在 |
|----------|----------|----------|
| `server/prompts` | `/workspace/server/server/prompts` | ❌ 不存在 |
| `server/prompts/schema` | `/workspace/server/server/prompts/schema` | ❌ 不存在 |
| `prompts` | `/workspace/server/prompts` | ✅ 存在 |
| `prompts/schema` | `/workspace/server/prompts/schema` | ✅ 存在 |

**根因确认：`backend.sh` 将 cwd 切换到 `server/` 目录，但配置默认 `PromptDir=server/prompts` 是相对于项目根目录的路径，导致路径解析失败。**

## 8. 根因分析

### 根因

**`backend.sh` 启动后端时 `cd` 到 `server/` 目录，而 `PromptDir` / `SchemaDir` 配置默认值 `"server/prompts"` / `"server/prompts/schema"` 是相对于项目根目录的路径。两者不匹配，导致 Loader 无法找到 Prompt 文件。**

### 影响范围

- `server/cmd/server/main.go`（通过 `make up` / `backend.sh` 启动）
- `server/cmd/llm-smoke/main.go`（如果也从 `server/` 目录运行）

### 不受影响

- 单元测试：使用 `filepath.Abs("../../prompts")` 相对于测试文件定位，不受 cwd 影响
- 从项目根目录直接 `go run ./cmd/server`：`server/prompts` 路径正确

## 9. 最小修复方案

在 `prompt.NewLoader` 和 `schema.NewLoader` 中添加 `resolveDir` 函数，当 `baseDir` 不存在时尝试去掉 `"server/"` 前缀。

### 修复代码

```go
func resolveDir(baseDir string) string {
    // Try the baseDir as-is first.
    if fi, err := os.Stat(baseDir); err == nil && fi.IsDir() {
        return baseDir
    }

    // If baseDir starts with "server/" or "server\", try without the prefix.
    stripped := ""
    if strings.HasPrefix(baseDir, "server/") {
        stripped = baseDir[7:]
    } else if strings.HasPrefix(baseDir, "server\\") {
        stripped = baseDir[7:]
    }
    if stripped != "" {
        if fi, err := os.Stat(stripped); err == nil && fi.IsDir() {
            return stripped
        }
    }

    // Return original — will fail later with a clear file-not-found error.
    return baseDir
}
```

### 修改位置

1. `server/internal/prompt/loader.go` — 添加 `resolveDir` + 修改 `NewLoader`
2. `server/internal/schema/loader.go` — 添加 `resolveDir` + 修改 `NewLoader`

### 修复策略说明

- 优先使用 `baseDir` 原始值（从项目根目录启动时正确）
- 如果原始值不存在且以 `"server/"` 开头，尝试去掉前缀（从 `server/` 目录启动时正确）
- 两种场景均兼容，不破坏现有测试

## 10. 修改文件

| 文件 | 修改内容 |
|------|----------|
| `server/internal/prompt/loader.go` | 添加 `resolveDir` 函数，`NewLoader` 调用 `resolveDir` |
| `server/internal/schema/loader.go` | 添加 `resolveDir` 函数，`NewLoader` 调用 `resolveDir` |
| `server/internal/prompt/loader_test.go` | 添加 `TestResolveDir_*` 和 `TestNewLoader_ResolvesServerPrefix` 4 个测试 |
| `server/internal/schema/loader_test.go` | 添加 `TestResolveDir_*` 和 `TestNewLoader_ResolvesServerPrefix` 4 个测试 |

## 11. 测试结果

| 测试 | 结果 |
|------|:----:|
| `go test ./internal/prompt/...` | ✅ 全部通过（含 4 个新增 resolveDir 测试） |
| `go test ./internal/schema/...` | ✅ 全部通过（含 4 个新增 resolveDir 测试） |
| `go test ./...` | ✅ 全部通过 |
| `make llm-check` | ✅ 通过 |
| 敏感信息检查 | ✅ 无泄露 |

## 12. 三类接口回归结果

| Agent | fallback_reason（修复前） | fallback_reason（修复后） | 结论 |
|-------|:---:|:---:|:---:|
| MatchAnalysis | `prompt_load_failed` | `llm_api_key_missing` | ✅ 修复成功 |
| Recommendation | `prompt_load_failed` | `llm_api_key_missing` | ✅ 修复成功 |
| ResumeOptimization | `prompt_load_failed` | `llm_api_key_missing` | ✅ 修复成功 |

说明：修复后 Prompt 加载成功，但 LLM 调用因 API Key 为占位符而认证失败，这是预期行为。

### 回归验证矩阵

| 验收项 | MatchAnalysis | Recommendation | ResumeOptimization | 结论 |
|--------|:---:|:---:|:---:|:---:|
| Prompt 文件存在 | pass | pass | pass | **pass** |
| Prompt loader 可定位 | pass | pass | pass | **pass** |
| 接口 HTTP 可访问 | pass | pass | pass | **pass** |
| data.result | pass | pass | pass | **pass** |
| data.meta | pass | pass | pass | **pass** |
| fallback_reason 不再是 prompt_load_failed | pass | pass | pass | **pass** |
| go test 覆盖 | pass | pass | pass | **pass** |

## 13. 当前限制

| 限制 | 严重度 | 说明 |
|------|--------|------|
| `request_id` 为空字符串 | 🟢 低 | 契约允许，后续 Handler 层接入 |
| `schema_valid` 默认 true | 🟡 中 | 未实现实际 Schema 校验逻辑 |
| LLM 因 API Key 占位符走 fallback | 🟢 低 | 配置真实 Key 后可正常使用 |
| 首次 `make up` 可能需 `make restart-soft` | 🟢 低 | 后端已在运行时 make up 不会重启 |

## 14. 4-5.3-R 最终结论

**pass**

判定依据：

1. ✅ 已定位 `prompt_load_failed` 根因（cwd 与配置路径不匹配）
2. ✅ 已完成最小修复（`resolveDir` 自动路径解析）
3. ✅ 三类 Prompt 文件存在
4. ✅ Prompt loader 可在 `server/` cwd 下加载三类 Prompt
5. ✅ 三类接口 curl 后不再出现 `prompt_load_failed`，变为 `llm_api_key_missing`
6. ✅ 三类接口仍返回 `data.result + data.meta`
7. ✅ `go test ./...` 通过
8. ✅ `make llm-check` 可用
9. ✅ 未调用真实 LLM
10. ✅ 未修改 Prompt / Schema / 前端 / Makefile / docker-compose / Nginx
11. ✅ `make down` 成功，volume 保留
12. ✅ 文档和开发日志已更新

## 15. 下一步建议

1. **回到 4-5.3 重新判定**：4-5.3 原判定 `partial_pass` 的核心限制已修复，可升级为 `pass`
2. **进入 4-5.4**：阶段 4 总收口与阶段 5 准入评审
3. **配置有效 API Key 后**：可重新执行真实 LLM Smoke
4. **后续 Handler 层**：补充 `request_id` 生成逻辑和 `schema_valid` 实际校验
