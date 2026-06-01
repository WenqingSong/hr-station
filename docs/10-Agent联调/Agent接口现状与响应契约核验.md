# Agent 接口现状与响应契约核验

> 阶段：7-0 | 状态：pass | 日期：2026-06-01

## 1. 核验范围

三类 Agent + VersionCompare，涵盖后端 route/handler/service/agent/prompt/schema + 前端 api/types/mocks/adapters。

## 2. 前端 Agent 契约摘要

### 路径

| Agent | 前端函数 | 前端路径 | HTTP |
|------|------|------|:--:|
| 推荐 | `runRecommendation` | `/jobs/recommend` → `/api/jobs/recommend` | POST |
| 匹配 | `runMatchAnalysis` | `/jobs/{jobId}/match` | POST |
| 优化 | `runResumeOptimization` | `/resumes/optimize` | POST |

### 请求体

| Agent | 前端 field | 后端 field | 匹配 |
|------|------|------|:--:|
| 推荐 | `resume_id`, `resume_snapshot`, `limit` | `resume_version_no` (int), `limit` | ❌ mismatch |
| 匹配 | `job_id`, `resume_id`, `resume_snapshot`, `job_snapshot` | `resume_version_no` (int) | ❌ mismatch |
| 优化 | `job_id`, `resume_id`, `resume_snapshot`, `job_snapshot`, `match_analysis_snapshot` | `resume_text` (required), `target_job` (required) | ❌ MAJOR mismatch |

### 响应结构

前端期望：`AgentResponseEnvelope<TResult> = { result: TResult, meta: AgentResponseMeta }`

后端 data 层返回：`{ result, meta }` — 结构与前端一致 ✅

### 前端 fallback

所有三个函数均有 `withMockFallback`，失败 → mock ✅

## 3. 后端 Agent 结构核验

| 模块 | 推荐 | 匹配 | 优化 |
|------|:--:|:--:|:--:|
| Route | `/api/jobs/recommend` | `/api/jobs/{id}/match` | `/api/resumes/optimize` |
| Handler | ✅ | ✅ | ✅ |
| Service | ✅ 双模式 | ✅ 双模式 | ✅ 双模式 |
| Agent (LLM) | ✅ | ✅ | ✅ |
| Agent (Mock) | ✅ | ✅ | ✅ |
| Prompt | ✅ | ✅ | ✅ |
| Schema | ✅ | ✅ | ✅ |
| LLM Client | HunyuanClient | HunyuanClient | HunyuanClient |
| Pipeline | ✅ LLM优先+Mock回退 | ✅ | ✅ |
| Model | ✅ | ✅ | ✅ |
| Repository | ✅ | ✅ | ✅ |

## 4. 推荐 Agent 接口现状

| 项目 | 状态 |
|------|:--:|
| 路径 | `POST /api/jobs/recommend` |
| 请求体 | `{ resume_version_no: int, limit: int }` |
| 响应 | `{ result, meta }` 内含 `recommended_jobs` |
| LLM | ✅ 双模式 |
| 前端路径一致 | ✅ |
| 前端请求体不一致 | ❌ (resume_id+snapshot vs version_no) |

## 5. 匹配分析 Agent 接口现状

| 项目 | 状态 |
|------|:--:|
| 路径 | `POST /api/jobs/{id}/match` |
| 请求体 | `{ resume_version_no: int }` |
| 响应 | `{ result, meta }` 内含 `match_score, evidence_chain` |
| LLM | ✅ 双模式 |
| 前端路径一致 | ✅ |
| 前端请求体不一致 | ❌ (多字段 vs 单字段) |

## 6. 简历优化 Agent 接口现状

| 项目 | 状态 |
|------|:--:|
| 路径 | `POST /api/resumes/optimize` |
| 请求体 | `{ resume_text: required, target_job: { title, jd_text } required }` |
| 响应 | `{ result, meta }` 内含优化建议 |
| LLM | ✅ 双模式 |
| 前端路径一致 | ✅ |
| 前端请求体完全不一致 | ❌ (job_id/snapshots vs resume_text/target_job) |

## 7. 响应结构核验

| 层 | 结构 | code/message/data | result/meta |
|------|:--:|:--:|:--:|
| 推荐 response | ✅ | ✅ | ✅ |
| 匹配 response | ✅ | ✅ | ✅ |
| 优化 response | ✅ | ✅ | ✅ |
| compare response | ✅ | ✅ | mock_agent |
| 前端 types | ✅ (AgentResponseEnvelope) | ✅ | ✅ |

## 8. Prompt / Schema 核验

| Agent | Prompt | Schema |
|------|:--:|:--:|
| 推荐 | `prompts/recommendation_v1.md` | `prompts/schema/recommendation_schema.json` |
| 匹配 | `prompts/match_analysis_v1.md` | `prompts/schema/match_analysis_schema.json` |
| 优化 | `prompts/resume_optimization_v1.md` | `prompts/schema/resume_optimization_schema.json` |

全部存在 ✅

## 9. LLM Client / Mock Agent 核验

| 项目 | 状态 |
|------|:--:|
| LLM Client | HunyuanClient (OpenAI-compatible) ✅ |
| Mock Agent | 三类 + compare 全有 ✅ |
| Pipeline | LLM 优先 → Mock fallback ✅ |
| NoopClient | ✅ (LLM disabled) |
| Prompt Loader | ✅ |
| Schema Loader + Validator | ✅ |
| JSON Extract/Parser | ✅ (内置 OpenAI-compatible + 自定义提取) |
| JSON Repair Agent | ✅ (P2 优先级) |

## 10. 前后端路径对齐核验

| Agent | 后端路径 | 前端路径 | 一致 |
|------|------|------|:--:|
| 推荐 | `POST /api/jobs/recommend` | `/jobs/recommend` (→ /api/jobs/recommend) | ✅ |
| 匹配 | `POST /api/jobs/{id}/match` | `/jobs/{jobId}/match` | ✅ |
| 优化 | `POST /api/resumes/optimize` | `/resumes/optimize` | ✅ |

## 11. Agent 与普通业务 API 边界

| 接口 | 类型 | 阶段 |
|------|------|:--:|
| GET /api/jobs | 普通 | 阶段6 |
| GET /api/jobs/{id} | 普通 | 阶段6 |
| GET/POST /api/resumes/* | 普通 | 阶段6 |
| POST /api/jobs/recommend | Agent | 阶段7 |
| POST /api/jobs/{id}/match | Agent | 阶段7 |
| POST /api/resumes/optimize | Agent | 阶段7 |
| GET /api/resumes/versions/compare | Mock Agent | 阶段6边界 |

## 12. 风险分级

| 编号 | 模块 | 风险 | 等级 |
|:--:|------|------|:--:|
| R1 | 推荐 | 请求体字段 mismatch（resume_id+snapshot vs version_no） | P1 |
| R2 | 匹配 | 请求体字段 mismatch | P1 |
| R3 | 优化 | 请求体完全 mismatch（前端快照 vs 后端文本+JD） | P1 |
| R4 | LLM | API Key 为占位符 | P1 |
| R5 | 优化 | 响应 result 字段需逐项对照 | P2 |

## 13. 是否可以进入 7-1

**可以进入。** 推荐/匹配 Agent 接口存在、路径一致、LLM+Mock 双模式就绪。请求体 mismatch 可在 7-1 中修复前端请求体。

## 14. 越界检查

server:❌ / web/src:❌ / Makefile:❌ — 仅修改 docs ✅

## 15. 最终结论

**pass ✅** — 三类 Agent 接口全部就绪，前后端路径一致，响应结构统一。
