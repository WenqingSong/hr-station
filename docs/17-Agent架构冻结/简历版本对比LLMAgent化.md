# K6-1：简历版本对比 LLM Agent 化

## 1. 目标

补齐四主 Agent 中唯一未 LLM 化的 Version Compare Agent。

## 2. 新增文件

| 文件 | 说明 |
|------|------|
| `server/internal/agent/version_compare_agent.go` | VersionCompareLLMAgent + FallbackError + PromptValues + Field mapping |
| `server/prompts/version_compare_v1.md` | 简洁 Prompt，仅 JSON，无 skeleton |
| `server/prompts/schema/version_compare_schema.json` | 9 required 字段，克制设计 |

## 3. 修改文件

| 文件 | 说明 |
|------|------|
| `server/internal/model/version_compare.go` | 新增 LLM 字段（overall_change, recommendation, improved_items 等） |
| `server/internal/service/version_compare_service.go` | 接入 pipelineRunner + llmEnabled，LLM 优先 Mock fallback |
| `server/cmd/server/main.go` | 传入 pipelineRunner 和 llmEnabled |
| `server/internal/prompt/types.go` | +PromptVersionCompare |
| `server/internal/prompt/loader.go` | +version_compare_v1.md mapping |
| `server/internal/schema/types.go` | +SchemaVersionCompare |
| `server/internal/schema/loader.go` | +version_compare_schema.json mapping |
| `server/internal/agent/pipeline/types.go` | +TargetCompare |
| `server/internal/agent/pipeline/pipeline.go` | +targetOutputTypeFromSchema |
| `server/internal/agent/repair/types.go` | +TargetCompare |
| `server/internal/agent/repair/json_repair_agent.go` | +schemaNameFromTarget |
| `server/internal/prompt/loader_test.go` | 4→5 names |
| `server/internal/schema/loader_test.go` | 3→4 names |

## 4. API 验证

```
GET /api/resumes/versions/compare?from=2&to=3
HTTP 200, code=0
generated_by: llm_agent ✅
overall_change: clear_regression
improved: 0, weakened: 6, missing: 3, next_actions: 3
```

## 5. Fallback 确认

LLM 失败 → Mock Agent 正常返回。前端对比页不受影响。

## 6. go test

ALL PASS ✅

## 7. P0/P1/P2

P0=0 P1=0 P2=1（浏览器未验）

## 8. 判定

**K6-1 = pass。** 四主 Agent 全部 LLM 化完成。
