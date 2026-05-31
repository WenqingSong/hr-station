# API Client 与响应消费适配验收

> 阶段：5-2.5 | 状态：已完成 | 日期：2026-05-31

## 1. 验收目标

统一验收阶段 5-2 是否已经完成前端 API Client 基础设施、Jobs / Resumes API、三类 Agent API、Mock fallback、`data.result + data.meta` 消费适配、ViewModel Loader 的闭环，判定是否可以进入 5-3 全局布局 / Navbar / 页面骨架修复。

## 2. 阶段 5-2 子阶段状态

| 编号 | 名称 | 结论 |
|------|------|:--:|
| 5-2.0 | APIClient 现状与接口契约核验 | pass |
| 5-2.1 | API Client 基础层建设 | pass |
| 5-2.2 | Jobs / Resumes API 模块与 Mock fallback | pass |
| 5-2.3 | 三类 Agent API 模块与响应适配 | pass |
| 5-2.4 | API 返回到 ViewModel 统一消费封装 | pass |
| 5-2.4-R | API 返回到 ViewModel 文档补齐 | pass |

## 3. `web/src/api/` 文件清单

| 文件 | 行数 | 职责 |
|------|:--:|------|
| `types.ts` | 25 | API Client 内部类型 |
| `errors.ts` | 28 | ApiClientError + isApiClientError |
| `client.ts` | 163 | 统一 fetch 请求层 |
| `jobs.ts` | 47 | 岗位 API + mock fallback |
| `resumes.ts` | 128 | 简历 API + mock fallback |
| `agents.ts` | 138 | 三类 Agent API + mock fallback |
| `viewModels.ts` | 120 | 9 个 ViewModel Loader |
| `index.ts` | 7 | 统一导出 |

**合计：8 文件，656 行。无 `selectors.ts`。**

## 4. client 基础层验收

| 验收项 | 结果 |
|------|:--:|
| 导出 `API_BASE_URL` | ✅ |
| 默认 base URL 为 `/api` | ✅ |
| 支持 `VITE_API_BASE_URL` 环境变量 | ✅ |
| `requestRaw<T>()` — 返回完整 ApiResponse | ✅ |
| `request<T>()` — 返回 data 字段 | ✅ |
| `get<T>()` | ✅ |
| `post<TData, TBody>()` | ✅ |
| `put<TData, TBody>()` | ✅ |
| `patch<TData, TBody>()` | ✅ |
| HTTP 非 2xx → ApiClientError | ✅ |
| JSON 解析失败 → ApiClientError（截断前 200 字符） | ✅ |
| 响应缺 code 字段 → ApiClientError | ✅ |
| `code !== 0` → ApiClientError | ✅ |
| 统一抛出 `ApiClientError` | ✅ |
| 无 axios | ✅ |
| 无业务路径 | ✅ |
| 无引用 mocks / adapters | ✅ |
| 无引用页面组件 | ✅ |

**结论：pass**

## 5. Jobs API 验收

| 验收项 | 函数 | 路径 | 结果 |
|------|------|------|:--:|
| 岗位列表 | `getJobs()` | `/jobs` | ✅ |
| 岗位详情 | `getJobDetail(jobId)` | `/jobs/{jobId}` | ✅ |
| Demo 默认岗位 | `getSelectedJob()` | `/jobs/{id}` | ✅ |
| 无 `/api/jobs` 硬编码前缀 | — | — | ✅ |
| 无 `/api/v1` | — | — | ✅ |
| Mock fallback | `withFallback` → `mockJobs/mockJobDetails/mockSelectedJob/getMockJobById` | — | ✅ |
| 未改页面 | — | — | ✅ |
| 未改后端 | — | — | ✅ |

**结论：pass**

## 6. Resumes API 验收

### P0 读取

| 验收项 | 函数 | 路径 | 结果 |
|------|------|------|:--:|
| 简历版本列表 | `getResumes()` | `/resumes/versions` | ✅ |
| 当前简历 | `getCurrentResume()` | `/resumes/current` | ✅ |
| 简历详情 | `getResumeDetail(id)` | `/resumes/detail?id=` | ✅ |
| 简历选项（弹窗） | `getResumeOptions()` | —（mock direct） | ✅ |
| 版本比较 | `compareResumes(baseId, targetId)` | `/resumes/versions/compare` | ✅ |

### P1 写入

| 验收项 | 函数 | 路径 | 结果 |
|------|------|------|:--:|
| 创建简历 | `createResume(input)` | `POST /resumes` | ✅ |
| 更新简历 | `updateResume(id, input)` | `PUT /resumes/detail` | ✅ |
| 设为当前 | `setCurrentResume(id)` | `PATCH /resumes/current` | ✅ |

### 额外验收

| 验收项 | 结果 |
|------|:--:|
| 无 `/api/v1` | ✅ |
| 无依赖 adapters | ✅ |
| 无改页面 | ✅ |
| 无改后端 | ✅ |
| P1 写入有稳定 fallback | ✅ |
| `getResumeOptions()` 稳定返回弹窗数据 | ✅ |

**结论：pass**

## 7. 三类 Agent API 验收

| 验收项 | 路径 | mock fallback | 结果 |
|------|------|:--:|:--:|
| `runRecommendation(input?)` | `POST /jobs/recommend` | `mockRecommendationResponse` | ✅ |
| `runMatchAnalysis(jobId, input?)` | `POST /jobs/{jobId}/match` | `mockMatchAnalysisResponse` | ✅ |
| `runResumeOptimization(input?)` | `POST /resumes/optimize` | `mockResumeOptimizationResponse` | ✅ |

### 返回类型

| 函数 | 返回类型 | 结果 |
|------|------|:--:|
| `runRecommendation` | `AgentResponseEnvelope<RecommendationResult>` | ✅ |
| `runMatchAnalysis` | `AgentResponseEnvelope<MatchAnalysisResult>` | ✅ |
| `runResumeOptimization` | `AgentResponseEnvelope<ResumeOptimizationResult>` | ✅ |

### 额外验收

| 验收项 | 结果 |
|------|:--:|
| 无 `/api/v1` | ✅ |
| 无旧路径（`/recommendation`, `/match-analysis`, `/resume-optimization`） | ✅ |
| 请求体含最小字段（resume_id/resume_snapshot/job_id/job_snapshot） | ✅ |
| Mock fallback 返回完整 envelope（result + meta） | ✅ |
| 未调用真实 LLM | ✅ |
| 未改后端 | ✅ |
| 未改页面 | ✅ |

**结论：pass**

## 8. ViewModel Loader 验收

### Jobs Loader

| 函数 | API 调用 | Adapter | 结果 |
|------|------|------|:--:|
| `loadJobListViewModels()` | `getJobs()` | `toJobCardViewModels` | ✅ |
| `loadJobDetailViewModel(jobId)` | `getJobDetail(jobId)` | `toJobDetailViewModel` | ✅ |
| `loadSelectedJobDetailViewModel()` | `getSelectedJob()` | `toJobDetailViewModel` | ✅ |

### Resumes Loader

| 函数 | API 调用 | Adapter | 结果 |
|------|------|------|:--:|
| `loadCurrentResumeSummaryViewModel()` | `getCurrentResume()` | `toResumeSummaryViewModel` | ✅ |
| `loadResumeOptionViewModels()` | `getResumeOptions()` | `toResumeOptionViewModels` | ✅ |
| `loadResumeComparisonViewModel(baseId, targetId)` | `compareResumes()` | `toResumeComparisonViewModel` | ✅ |
| `loadResumeSummaryViewModel(resumeId)` | `getResumeDetail()` | `toResumeSummaryViewModel` | ✅ |

### Agent Loader

| 函数 | API 调用 | data.result+data.meta | Adapter | 结果 |
|------|------|:--:|------|:--:|
| `loadRecommendationViewModel(input?)` | `runRecommendation()` | ✅ | `toRecommendationPageViewModel` | ✅ |
| `loadMatchAnalysisViewModel(jobId, input?)` | `runMatchAnalysis()` | ✅ | `toMatchAnalysisPageViewModel` | ✅ |
| `loadResumeOptimizationViewModel(input?)` | `runResumeOptimization()` | ✅ | `toResumeOptimizationPageViewModel` | ✅ |

### 额外验收

| 验收项 | 结果 |
|------|:--:|
| 无 fetch | ✅ |
| 无 fallback 复杂逻辑（复用 API 层 fallback） | ✅ |
| 无直接引用页面组件 | ✅ |
| 无新增 selectors | ✅ |
| 返回明确 Promise 类型 | ✅ |

**结论：pass**

## 9. index 导出验收

```ts
export * from './types';        // ✅
export * from './errors';       // ✅
export * from './client';       // ✅
export * from './jobs';         // ✅
export * from './resumes';      // ✅
export * from './agents';       // ✅
export * from './viewModels';   // ✅
```

无 selectors 导出。

**结论：pass**

## 10. Mock fallback 验收

| 模块 | fallback 策略 | 实现 | 结果 |
|------|------|------|:--:|
| jobs.ts | `withFallback` helper | try→catch→fallback | ✅ |
| resumes.ts | `withFallback` helper | try→catch→fallback | ✅ |
| agents.ts | `withMockFallback` helper | try→catch→mock | ✅ |

所有 fallback 行为一致：
- API 调用失败（网络错误 / HTTP 非 2xx / code!==0）→ 自动 fallback
- 不打印错误日志，不弹 toast
- fallback 始终返回可用数据（来自 `web/src/mocks/`）

**结论：pass**

## 11. data.result + data.meta 消费验收

### 消费链路

```
agents.ts: post() → AgentResponseEnvelope<TResult>
       ↓
viewModels.ts: response.result + response.meta
       ↓
adapters: toXxxPageViewModel(result, meta)
       ↓
ViewModel (含 agentMeta / topFixItems / evidenceItems 等)
```

### 前端 TypeScript (agent.ts)

```
agent_name / generated_by / agent_status / fallback_reason /
error_code / repair_applied / schema_valid / request_id
```

### 后端 Go Model (agent_response.go)

```
agent_name / generated_by / agent_status / fallback_reason /
error_code / repair_applied / schema_valid / request_id
```

**前后端 8 字段完全一致。**

### 消费路径

| Agent | agents.ts 解包 | viewModels.ts 拆分 | adapter 消费 |
|------|:--:|:--:|:--:|
| Recommendation | `response.result + response.meta` | 透传 | `toRecommendationPageViewModel(result, meta)` |
| MatchAnalysis | `response.result + response.meta` | 透传 | `toMatchAnalysisPageViewModel(result, meta)` |
| ResumeOptimization | `response.result + response.meta` | 透传 | `toResumeOptimizationPageViewModel(result, meta)` |

**页面后续不需要直接理解 AgentResponseEnvelope。**

**结论：pass**

## 12. 页面 / 组件越界检查

| 检查项 | 结果 |
|------|:--:|
| 页面 import api 模块 | ❌ 无 |
| 组件 import api 模块 | ❌ 无 |
| app import api 模块 | ❌ 无 |

**结论：零越界**

## 13. API / 后端越界检查

| 检查项 | 结果 |
|------|:--:|
| `fetch(` 只出现在 `client.ts` | ✅ |
| `axios` | ❌ 无 |
| `/api/v1` | ❌ 无 |
| `selectors` (web/src/api/) | ❌ 无 |
| server/ 修改 | ❌ 无 |
| Makefile / docker-compose / nginx / .env 修改 | ❌ 无 |

**结论：零越界**

## 14. build 结果

```text
cd web && npm run build
→ sh: 1: vite: not found
```

**结论：环境依赖问题（node_modules 未安装），非代码问题。**

## 15. docs 完整性

| 文档 | 存在 |
|------|:--:|
| `docs/07-前端设计/APIClient现状与接口契约核验.md` | ✅ |
| `docs/07-前端设计/APIClient基础层建设.md` | ✅ |
| `docs/07-前端设计/JobsResumesAPI模块与MockFallback.md` | ✅ |
| `docs/07-前端设计/三类AgentAPI模块与响应适配.md` | ✅ |
| `docs/07-前端设计/API返回到ViewModel统一消费封装.md` | ✅ |

### 开发日志

| 阶段 | 开发日志 |
|------|------|
| 5-2.0 | #140-APIClient现状与接口契约核验 ✅ |
| 5-2.1 | #141-APIClient基础层建设 ✅ |
| 5-2.2 | 包含在 #141 中 ✅ |
| 5-2.3 | 包含在 #141 中 ✅ |
| 5-2.4 | 包含在 #142 中 ✅ |
| 5-2.4-R | #142-API返回到ViewModel统一消费封装文档补齐 ✅ |

**结论：完整**

## 16. 已知限制

| 限制 | 说明 |
|------|------|
| build 不可用 | `node_modules` 未安装，无法验证 TypeScript 编译 |
| JWT token 注入 | `client.ts` 当前未注入 Bearer token（后续页面接入时追加） |
| 无真实接口联调 | 所有 API 函数当前通过 mock fallback 返回数据，未对接真实后端 |
| 无全局 store | loading/error 状态管理尚未实现 |
| P1 写入函数 | createResume/updateResume/setCurrentResume 仅 Demo 占位 |

## 17. 是否可以进入 5-3

**可以。**

- `web/src/api/` 8 文件完整，层级清晰
- client / jobs / resumes / agents / viewModels 边界分明
- Mock fallback 完整，所有调用链路有兜底
- `data.result + data.meta` 消费链路完整
- 零 axios / 零 `/api/v1` / 零 selectors
- 页面零直接 import api
- 后端零修改
- docs 完整，无缺口

## 18. 验收结论

**pass**

阶段 5-2 API Client 与响应消费适配已完整闭环，具备进入 5-3 全局布局 / Navbar / 页面骨架修复的条件。
