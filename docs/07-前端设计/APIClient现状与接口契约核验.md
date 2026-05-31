# API Client 现状与接口契约核验

> 阶段：5-2.0 | 状态：已完成 | 日期：2026-05-31

## 1. 本阶段目标

在正式创建 `web/src/api/` 前，核验前后端接口契约对齐状态，明确 5-2.1~5-2.5 设计输入。

## 2. 前端 API Client 现状

| 检查项 | 状态 |
|------|:--:|
| `web/src/api/` 目录 | ❌ 不存在 |
| `fetch(` 调用 | ❌ 无 |
| `axios` | ❌ 无 |
| `VITE_API_BASE_URL` | ❌ 未定义 |
| `import.meta.env` | ❌ 未使用 |
| API Client 封装 | ❌ 无 |
| 页面直接调后端 | ❌ 无 |
| types/mocks/adapters 就绪 | ✅ 三层完整（21 文件 / 1975 行） |
| tsconfig `@/*` 别名 | ✅ 已配置 |

**结论：前端 API 层完全为零，types/mocks/adapters 基础设施完备。**

## 3. 后端接口实际状态

### 全局路由表

| 路径 | 方法 | 认证 | Handler |
|------|:--:|:--:|------|
| `/api/auth/register` | POST | 否 | AuthHandler.Register |
| `/api/auth/login` | POST | 否 | AuthHandler.Login |
| `/api/auth/me` | GET | 是 | AuthHandler.Me |
| `/api/jobs` | GET | 是 | JobHandler.List |
| `/api/jobs/{id}` | GET | 是 | JobHandler.Detail |
| `/api/jobs/{id}/match` | POST | 是 | MatchHandler.Analyze |
| `/api/jobs/recommend` | POST | 是 | RecommendationHandler.Recommend |
| `/api/resumes` | POST | 是 | ResumeHandler.Create |
| `/api/resumes/current` | GET | 是 | ResumeHandler.Current |
| `/api/resumes/detail` | GET | 是 | ResumeHandler.Detail |
| `/api/resumes/versions` | GET | 是 | ResumeHandler.Versions |
| `/api/resumes/versions/{id}` | GET | 是 | ResumeHandler.VersionDetail |
| `/api/resumes/versions/compare` | GET | 是 | VersionCompareHandler.Compare |
| `/api/resumes/optimize` | POST | 是 | ResumeOptimizationHandler.Optimize |

**Base URL：`/api`（不带 `/v1`）**

### 响应格式

```
{ "code": 0, "message": "ok", "data": {...} }
```

## 4. Jobs 接口核验

| 接口 | 路径 | 方法 | 后端 | 前端 Job 类型 | 匹配 |
|------|------|:--:|:--:|------|:--:|
| 岗位列表 | `/api/jobs` | GET | JobHandler.List | `Job[]` | ✅ |
| 岗位详情 | `/api/jobs/{id}` | GET | JobHandler.Detail | `JobDetail` | ✅ |

**Mock fallback**：`mockJobs` / `mockJobDetails`

## 5. Resumes 接口核验

| 接口 | 路径 | 方法 | 后端 | 前端类型 | P0/P1 |
|------|------|:--:|:--:|------|:--:|
| 当前简历 | `/api/resumes/current` | GET | ResumeHandler.Current | `Resume` | P0 |
| 简历列表 | `/api/resumes/versions` | GET | ResumeHandler.Versions | `Resume[]` | P0 |
| 简历详情 | `/api/resumes/detail` + `/api/resumes/versions/{id}` | GET | ResumeHandler.Detail/VersionDetail | `Resume` | P1 |
| 创建简历 | `/api/resumes` | POST | ResumeHandler.Create | — | P1 |
| 版本比较 | `/api/resumes/versions/compare` | GET | VersionCompareHandler.Compare | `ResumeVersionComparison` | P1 |

**Mock fallback**：`mockCurrentResume` / `mockResumeOptions` / `mockResumeComparison`

## 6. 三类 Agent 接口核验

### 6.1 RecommendationAgent

| 项 | 值 |
|------|------|
| 路径 | `POST /api/jobs/recommend` |
| 请求体 | `{ resume_version_no?: int, limit?: int }` |
| 返回结构 | `ApiResponse<{ result: RecommendationResult, meta: AgentResponseMeta }>` |
| data.result.recommended_jobs | ✅ |
| data.meta | ✅ (8字段) |
| 前端类型匹配 | ✅ `RecommendationResult` |
| **Mock fallback** | `mockRecommendationResponse` |

### 6.2 MatchAnalysisAgent

| 项 | 值 |
|------|------|
| 路径 | `POST /api/jobs/{id}/match` |
| 请求体 | `{ resume_version_no?: int }` (可选) |
| 返回结构 | `ApiResponse<{ result: MatchAnalysisResult, meta: AgentResponseMeta }>` |
| match_score + match_level | ✅ |
| top_fix_items + evidence_chain | ✅ |
| data.meta | ✅ (8字段) |
| 前端类型匹配 | ✅ `MatchAnalysisResult` |
| **Mock fallback** | `mockMatchAnalysisResponse` |

### 6.3 ResumeOptimizationAgent

| 项 | 值 |
|------|------|
| 路径 | `POST /api/resumes/optimize` |
| 请求体 | `{ resume_text, target_job, user_intention }` |
| 返回结构 | `ApiResponse<{ result: ResumeOptimizationResult, meta: AgentResponseMeta }>` |
| summary + top_fix_items | ✅ |
| keyword_alignment + project_optimization | ✅ |
| rewrite_examples + final_checklist | ✅ |
| data.meta | ✅ (8字段) |
| 前端类型匹配 | ✅ `ResumeOptimizationResult` |
| **Mock fallback** | `mockResumeOptimizationResponse` |

## 7. data.result + data.meta 契约核验

### 后端 Go Model (agent_response.go)

```
agent_name / generated_by / agent_status / fallback_reason /
error_code / repair_applied / schema_valid / request_id
```

### 前端 TypeScript (agent.ts)

```
agent_name / generated_by / agent_status / fallback_reason /
error_code / repair_applied / schema_valid / request_id
```

**结论：前后端 8 字段完全一致。**

### 已知限制

| 限制 | 说明 | 方案 |
|------|------|------|
| request_id 为空字符串 | 后端返回空字符串 | 前端默认值 |
| schema_valid 默认 true | Mock 场景未实际校验 | 前端以 mock_success 为正常态 |

## 8. API Client 设计输入

### 8.1 基础层 (5-2.1)

```
web/src/api/client.ts   — request<T>, get<T>, post<T>, put<T>, patch<T>
web/src/api/errors.ts   — ApiError class, isApiError guard
web/src/api/types.ts    — import from web/src/types/api.ts
web/src/api/index.ts    — 统一导出
```

关键设计：
- 原生 `fetch`，不引入 axios
- `VITE_API_BASE_URL` 环境变量
- 统一 `code !== 0` 错误处理
- JWT token 注入（Bearer header）
- 不打印敏感信息

### 8.2 Jobs/Resumes API (5-2.2)

```
web/src/api/jobs.ts
  - getJobs(params?)
  - getJobDetail(jobId: string)

web/src/api/resumes.ts
  - getCurrentResume()        [P0]
  - getResumeVersions()       [P0]
  - getResumeDetail(id)       [P1]
  - createResume(data)        [P1]
  - compareResumes(params)    [P1]
```

Mock fallback 优先级：API 成功 → mock fallback

### 8.3 Agents API (5-2.3)

```
web/src/api/agents.ts
  - runRecommendation(input?)   [P0]
  - runMatchAnalysis(input?)    [P0]
  - runResumeOptimization(input)[P0]
```

返回类型：`ApiResponse<AgentResponseEnvelope<TResult>>`

Mock fallback：`mockRecommendationResponse` / `mockMatchAnalysisResponse` / `mockResumeOptimizationResponse`

### 8.4 ViewModel Loader (5-2.4)

```
web/src/api/selectors.ts
  - loadRecommendationViewModel()
  - loadMatchAnalysisViewModel()
  - loadResumeOptimizationViewModel()
  - loadJobListViewModels()
  - loadJobDetailViewModel()
  - loadResumeOptionViewModels()
```

组合 API 调用 + adapter 转换 → 页面直接消费 ViewModel

## 9. 接口契约矩阵

| 领域 | 前端需要 | 后端接口 | 返回结构 | 类型覆盖 | fallback | 5-2 处理 |
|------|------|------|------|:--:|:--:|------|
| Jobs 列表 | `Job[]` | `GET /api/jobs` | `{code,message,data:Job[]}` | ✅ | mockJobs | 封装 + fallback |
| Jobs 详情 | `JobDetail` | `GET /api/jobs/{id}` | `{code,message,data:JobDetail}` | ✅ | mockJobDetails | 封装 + fallback |
| Resumes 当前 | `Resume` | `GET /api/resumes/current` | `{code,message,data:Resume}` | ✅ | mockCurrentResume | P0 封装 + fallback |
| Resumes 列表 | `Resume[]` | `GET /api/resumes/versions` | `{code,message,data:Resume[]}` | ✅ | mockResumeOptions | P0 封装 + fallback |
| Resumes 比较 | `ResumeVersionComparison` | `GET /api/resumes/versions/compare` | `{code,message,data:...}` | ✅ | mockResumeComparison | P1 |
| Recommendation | `RecommendationResult+meta` | `POST /api/jobs/recommend` | `data.result+data.meta` | ✅ | mockRecommendationResponse | P0 封装 + fallback |
| MatchAnalysis | `MatchAnalysisResult+meta` | `POST /api/jobs/{id}/match` | `data.result+data.meta` | ✅ | mockMatchAnalysisResponse | P0 封装 + fallback |
| ResumeOptimization | `ResumeOptimizationResult+meta` | `POST /api/resumes/optimize` | `data.result+data.meta` | ✅ | mockResumeOptimizationResponse | P0 封装 + fallback |
| API Client 基础层 | `request<T>` | — | — | ✅ | — | 5-2.1 创建 |
| ViewModel Loader | API→ViewModel | — | — | ✅ | — | 5-2.4 创建 |

## 10. 前端 API 模块规划

| 阶段 | 文件 | 主要函数 | fallback | 接页面 |
|------|------|------|:--:|:--:|
| 5-2.1 | `api/client.ts` / `errors.ts` / `index.ts` | `request`/`get`/`post` | — | 否 |
| 5-2.2 | `api/jobs.ts` / `api/resumes.ts` | `getJobs`/`getResumes` | ✅ | 否 |
| 5-2.3 | `api/agents.ts` | `runRecommendation`/`runMatchAnalysis`/`runResumeOptimization` | ✅ | 否 |
| 5-2.4 | `api/selectors.ts` | `load*ViewModel` 函数族 | ✅ | 否 |
| 5-2.5 | docs only | 验收 | — | 否 |

## 11. P0 / P1 边界

### P0

- 统一 `request<T>` client（fetch 原生，无 axios）
- 统一 `ApiResponse<T>` + `AgentResponseEnvelope<T>` 解析
- jobs 读接口封装
- resumes 读接口 P0 封装 + mock fallback
- 三类 Agent API 封装 + mock fallback
- API 返回 → ViewModel loader
- JWT token 注入

### P1

- resume 写接口（create/update）
- resume 版本比较接口
- 复杂错误重试
- 请求取消 AbortController
- 全局 loading store
- token 刷新
- 附件上传

## 12. 是否可以进入 5-2.1

**可以。**

- 后端 14 个接口全部就绪，路径确认
- `data.result + data.meta` 契约前后端 8 字段一致
- 前端 types/mocks/adapters 三层完整
- API Client 设计输入明确
- P0/P1 边界清晰
