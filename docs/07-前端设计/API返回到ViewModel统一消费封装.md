# API 返回到 ViewModel 统一消费封装

> 阶段：5-2.4 | 状态：已完成 | 日期：2026-05-31

## 1. 阶段目标

5-2.4 的目标是新增统一 ViewModel Loader 层，让后续页面不直接调用原始 API 函数，也不直接理解后端原始返回结构，而是通过统一 Loader 获取页面可消费的 ViewModel。

核心链路：

```text
API / Mock fallback 返回原始数据
        ↓
ViewModel Loader
        ↓
Adapter 转换
        ↓
页面 ViewModel
```

## 2. 新增文件

| 文件 | 行数 | 说明 |
|------|:--:|------|
| `web/src/api/viewModels.ts` | 120 | 9 个 Loader 函数，串联 API 调用 + Adapter 转换 |

最小修改：

| 文件 | 说明 |
|------|------|
| `web/src/api/index.ts` | 新增 `export * from './viewModels'` |

## 3. ViewModel Loader 职责

`viewModels.ts` 负责：

1. 调用 `jobs.ts / resumes.ts / agents.ts` 中的 API 函数；
2. 复用这些 API 函数已有的 mock fallback；
3. 调用 `adapters/` 中的转换函数；
4. 返回页面可直接消费的 ViewModel；
5. 不负责页面渲染；
6. 不负责 loading / error UI；
7. 不直接写 fetch；
8. 不直接读后端响应原始结构。

## 4. Jobs Loader

| 函数 | 返回类型 | 说明 |
|------|------|------|
| `loadJobListViewModels()` | `JobCardViewModel[]` | 加载岗位列表 ViewModel |
| `loadJobDetailViewModel(jobId)` | `JobDetailViewModel` | 加载岗位详情 ViewModel |
| `loadSelectedJobDetailViewModel()` | `JobDetailViewModel` | 加载 Demo 默认岗位 ViewModel |

数据流：

```text
getJobs() / getJobDetail() / getSelectedJob()
        ↓
toJobCardViewModels() / toJobDetailViewModel()
        ↓
JobCardViewModel / JobDetailViewModel
```

## 5. Resumes Loader

| 函数 | 返回类型 | 说明 |
|------|------|------|
| `loadCurrentResumeSummaryViewModel()` | `ResumeSummaryViewModel` | 加载当前简历摘要 |
| `loadResumeOptionViewModels()` | `ResumeOptionViewModel[]` | 加载简历选项列表（选择简历弹窗） |
| `loadResumeComparisonViewModel(baseId, targetId)` | `ResumeComparisonViewModel` | 加载简历版本比较 |
| `loadResumeSummaryViewModel(resumeId)` | `ResumeSummaryViewModel` | 加载指定简历摘要 |

数据流：

```text
getCurrentResume() / getResumeOptions() / compareResumes() / getResumeDetail()
        ↓
resumeAdapter（toResumeSummaryViewModel / toResumeOptionViewModels / toResumeComparisonViewModel）
        ↓
ResumeSummaryViewModel / ResumeOptionViewModel / ResumeComparisonViewModel
```

## 6. Agent Loader

| 函数 | 返回类型 | 说明 |
|------|------|------|
| `loadRecommendationViewModel(input?)` | `RecommendationPageViewModel` | 加载岗位推荐页 ViewModel |
| `loadMatchAnalysisViewModel(jobId, input?)` | `MatchAnalysisPageViewModel` | 加载岗位匹配分析页 ViewModel |
| `loadResumeOptimizationViewModel(input?)` | `ResumeOptimizationPageViewModel` | 加载简历优化页 ViewModel |

数据流：

```text
runRecommendation() / runMatchAnalysis() / runResumeOptimization()
        ↓
AgentResponseEnvelope<TResult>
        ↓
response.result + response.meta
        ↓
对应 adapter（toRecommendationPageViewModel / toMatchAnalysisPageViewModel / toResumeOptimizationPageViewModel）
        ↓
页面 ViewModel
```

**重点强调：`data.result + data.meta` 的解析已经在 Agent API + Loader 层完成。后续页面不需要直接理解 `AgentResponseEnvelope`。**

Loader 中 Agent 函数的典型实现：

```ts
export async function loadRecommendationViewModel(
  input?: Partial<RecommendationRequest>,
): Promise<RecommendationPageViewModel> {
  const response = await runRecommendation(input);
  return toRecommendationPageViewModel(response.result, response.meta);
}
```

`response.result` 是业务数据（`RecommendationResult` / `MatchAnalysisResult` / `ResumeOptimizationResult`），`response.meta` 是 8 字段系统状态（`AgentResponseMeta`）。二者在 Loader 层拆包后传入对应 adapter，页面只消费 ViewModel。

## 7. 与 API / adapters 的关系

分工说明：

| 层 | 职责 |
|------|------|
| `client.ts` | 负责底层 request / response 解析（fetch 原生，统一 code!==0 错误处理，JWT 注入） |
| `jobs.ts / resumes.ts / agents.ts` | 负责业务 API 调用 + mock fallback（withFallback 策略） |
| `viewModels.ts` | 负责调用 API 模块并转 ViewModel（串联 API + Adapter） |
| `adapters/` | 负责纯数据转换（原始类型 → ViewModel，含 tone/label 预计算） |
| `pages/components` | 后续只消费 ViewModel（不直接调 API，不理解后端原始结构） |

## 8. 本阶段未做事项

```text
未写 selectors.ts
未改页面
未改后端
未改 types
未改 mocks
未改 adapters
未做 loading/error UI
未做页面接口接入
未做真实联调
未新增 axios
未新增 /api/v1
```

## 9. 后续 5-2.5 输入

5-2.5 需要验收：

1. `web/src/api/` 是否完整（client / errors / types / jobs / resumes / agents / viewModels / index）；
2. `client / jobs / resumes / agents / viewModels` 是否边界清晰；
3. 是否无 axios；
4. 是否无 `/api/v1`；
5. 是否无页面直接 fetch；
6. 是否无后端改动；
7. 是否 docs 缺口已补齐。

✅ 5-2.5 API Client 与响应消费适配验收 **pass**。全链路闭环，具备进入 5-3 条件。详见 `docs/07-前端设计/APIClient与响应消费适配验收.md`。
