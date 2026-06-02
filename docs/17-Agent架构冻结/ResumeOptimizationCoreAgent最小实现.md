# K6-3：Resume Optimization Core Agent 最小实现

## 1. 目标

实现 Optimization Core Agent（首屏核心优化建议），作为 Resume Optimization 主 Agent 的子能力。

## 2. 修改文件清单

| 文件 | 操作 |
|------|------|
| `server/prompts/resume_optimization_core_v1.md` | **新增** — 短 Prompt (~30 行)，JSON-only |
| `server/prompts/schema/resume_optimization_core_schema.json` | **新增** — 7 required 字段 |
| `server/internal/agent/resume_optimization_core_agent.go` | **新增** — CoreLLMAgent + CoreMockAgent |
| `server/internal/service/resume_optimization_service.go` | **修改** — +Mode 字段，+optimizeCore() |
| `server/internal/prompt/types.go` | **修改** — +PromptResumeOptimizationCore |
| `server/internal/prompt/loader.go` | **修改** — +mapping |
| `server/internal/prompt/loader_test.go` | **修改** — 5→6 names |
| `server/internal/schema/types.go` | **修改** — +SchemaResumeOptimizationCore |
| `server/internal/schema/loader.go` | **修改** — +mapping |
| `server/internal/schema/loader_test.go` | **修改** — 4→5 names |
| `server/internal/agent/pipeline/types.go` | **修改** — +TargetResumeOptimizationCore |
| `server/internal/agent/pipeline/pipeline.go` | **修改** — +case |
| `server/internal/agent/repair/types.go` | **修改** — +TargetResumeOptimizationCore |
| `server/internal/agent/repair/json_repair_agent.go` | **修改** — +case |

## 3. 接口设计

```
POST /api/resumes/optimize  {"mode":"core", ...}  → Core Agent
POST /api/resumes/optimize  (no mode)              → 原完整 Agent（兼容）
```

## 4. Core API 验证

```
POST /api/resumes/optimize {"mode":"core", ...}
HTTP 200
generated_by: llm_agent ✅
agent_status: success ✅
fallback_reason: none
schema_valid: True ✅
top_fix_items: 3 ✅
apply_readiness: needs_revision
summary: "简历当前仅体现 Go、Redis、MySQL 三项技术..."
```

## 5. 旧接口兼容

旧 `POST /api/resumes/optimize` 无 mode 时走原完整流程，未修改。

## 6. go test

ALL PASS ✅

## 7. 前端

未修改。首屏接入后置。

## 8. P0/P1/P2

P0=0 ✅ P1=0 ✅ P2=2（前端未接入 + 浏览器未验）

## 9. 判定

**K6-3 = pass。** Core Agent 真实 LLM 首次通过。可进入 K6-4。
