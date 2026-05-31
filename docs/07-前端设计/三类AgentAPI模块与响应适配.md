# 三类 Agent API 模块与响应适配

> 阶段：5-2.3 | 状态：已完成 | 日期：2026-05-31

## 1. 本阶段目标

新增三类 Agent API 模块，统一消费 `AgentResponseEnvelope<TResult>`（`data.result + data.meta`），并支持 mock fallback。

## 2. 新增文件

| 文件 | 行数 | 说明 |
|------|:--:|------|
| `web/src/api/agents.ts` | 138 | 三类 Agent 函数 + 请求类型 + 类型别名 |

## 3. 三类函数

### runRecommendation(input?)

| 项 | 值 |
|------|------|
| 路径 | `POST /jobs/recommend` |
| 默认输入 | `resume_id/resume_snapshot=mockCurrentResume, limit=5` |
| 返回 | `AgentResponseEnvelope<RecommendationResult>` |
| fallback | `mockRecommendationResponse` |

### runMatchAnalysis(jobId, input?)

| 项 | 值 |
|------|------|
| 路径 | `POST /jobs/{jobId}/match` |
| 默认输入 | `job_id=jobId, resume_id/resume_snapshot=mockCurrentResume, job_snapshot=mockSelectedJob` |
| 返回 | `AgentResponseEnvelope<MatchAnalysisResult>` |
| fallback | `mockMatchAnalysisResponse` |

### runResumeOptimization(input?)

| 项 | 值 |
|------|------|
| 路径 | `POST /resumes/optimize` |
| 默认输入 | `job_id=mockSelectedJob.id, resume_id/resume_snapshot=mockCurrentResume, job_snapshot=mockSelectedJob` |
| 返回 | `AgentResponseEnvelope<ResumeOptimizationResult>` |
| fallback | `mockResumeOptimizationResponse` |

## 4. 请求体最小字段

| 字段 | 类型 | 说明 |
|------|------|------|
| `resume_id` | `string?` | 简历 ID |
| `resume_snapshot` | `Resume?` | Demo 稳定运行 |
| `job_id` | `string` | 岗位 ID |
| `job_snapshot` | `JobDetail?` | Demo 稳定运行 |
| `limit` | `number?` | 推荐数量 |

## 5. data.result + data.meta 消费方式

```ts
const response = await runRecommendation();
// response.result → RecommendationResult (业务数据)
// response.meta → AgentResponseMeta (8字段系统状态)
```

## 6. 本阶段未做事项

- 未写 ViewModel loader（5-2.4）
- 未改页面/组件
- 未改后端
- 未改 types/mocks/adapters
- 未引入 axios 或 /api/v1

## 7. 后续 5-2.4 输入

可基于 `runRecommendation()` 等函数 + adapters 封装 `loadRecommendationViewModel()` 等 loader。
