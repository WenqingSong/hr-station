# 页面数据源从 Mocks 切到 API ViewModels

> 阶段：6-4 | 状态：已完成 | 日期：2026-06-01

## 1. 本阶段目标

将普通业务页面/组件的数据来源从直接 mocks 切换到 api viewModels loader。

## 2. 修改范围

| 文件 | 原数据源 | 新数据源 | 切换方式 |
|------|------|------|------|
| `JobList.tsx` | `mockJobs` → `toJobCardViewModels` | `loadJobListViewModels()` | useEffect + useState |
| `JobDetailBanner.tsx` | `mockSelectedJob` → `toJobDetailViewModel` | `loadSelectedJobDetailViewModel()` | useEffect + useState |
| `JobDetailJD.tsx` | `mockSelectedJob` → `toJobDetailViewModel` | `loadSelectedJobDetailViewModel()` | useEffect + useState |
| `AIDecisionWorkspace.tsx` | `mockCurrentResume` / `mockSelectedJob` | `loadCurrentResumeSummaryViewModel` / `loadSelectedJobDetailViewModel` | useEffect + useState |
| `JobDetailAIDecision.tsx` | `mockCurrentResume` / `mockSelectedJob` | `loadCurrentResumeSummaryViewModel` / `loadJobDetailViewModel(jobId)` | useEffect + useState（从路由取 jobId） |
| `SelectResumeModal.tsx` | `mockResumeOptions` → `toResumeOptionViewModels` | `loadResumeOptionViewModels()` | useEffect（弹窗打开时加载） |

## 3. API ViewModel Loader 设计

复用现有 `web/src/api/viewModels.ts`，已有 loader：
- `loadJobListViewModels()` → `JobCardViewModel[]`
- `loadJobDetailViewModel(jobId)` → `JobDetailViewModel`
- `loadSelectedJobDetailViewModel()` → `JobDetailViewModel`
- `loadCurrentResumeSummaryViewModel()` → `ResumeSummaryViewModel`
- `loadResumeOptionViewModels()` → `ResumeOptionViewModel[]`

每条 loader 内部调用 api 层（已有 withFallback），失败时自动 fallback mock。

## 4. Jobs 页面切换结果

| 页面/组件 | 切换 | 状态 |
|------|:--:|:--:|
| JobList（岗位列表） | mockJobs → loadJobListViewModels | ✅ 已切换 |
| JobDetailBanner（详情横幅） | mockSelectedJob → loadSelectedJobDetailViewModel | ✅ 已切换 |
| JobDetailJD（JD 详情） | mockSelectedJob → loadSelectedJobDetailViewModel | ✅ 已切换 |
| AIDecisionWorkspace（决策侧栏） | mockCurrentResume/mockSelectedJob → loaders | ✅ 已切换 |
| JobDetailAIDecision（详情决策） | mockCurrentResume/mockSelectedJob → loaders（支持路由 jobId） | ✅ 已切换 |

## 5. Resumes 页面切换结果

| 页面/组件 | 切换 | 状态 |
|------|:--:|:--:|
| SelectResumeModal | mockResumeOptions → loadResumeOptionViewModels | ✅ 已切换 |
| ResumeManagementPage | 未直接 import mocks（使用内嵌硬编码数据） | ⚠️ 留到后续阶段 |
| MyResumesPage | useState 本地管理，未 import mocks | ⚠️ 留到后续阶段 |
| ResumeComparisonPage | JSX 内硬编码，未 import mocks | ⚠️ 留到后续阶段 |

ResumeManagementPage / MyResumesPage / ResumeComparisonPage 使用**本地硬编码数据**而非 mocks 导入，本质上也需切换但结构重构更复杂，本阶段先完成直接 mock 导入的切换。

## 6. SelectResumeModal 切换结果

- 弹窗打开时通过 `loadResumeOptionViewModels()` 加载
- 空数据时显示"正在加载简历列表…"
- 加载失败时 fallback mock 数据（api 层 withFallback 保证）

## 7. Agent 页面未切换说明

以下页面/组件保留原有 mock 逻辑（阶段 7 处理）：

| 文件 | 保留的 mock |
|------|------|
| `JobMatchAnalysisPage.tsx` | mockSelectedJob / mockCurrentResume / mockResumeVersions |
| `JobRecommendationPage.tsx` | mockCurrentResume / mockResumeVersions |
| `ResumeOptimizationPage.tsx` | mockSelectedJob / mockCurrentResume / mockResumeVersions |
| `MatchAnalysisDetail.tsx` | mockMatchAnalysisResponse |
| `RecommendationContext.tsx` | mockCurrentResume |
| `RecommendationDetail.tsx` | mockRecommendationResponse |
| `RecommendationSidebar.tsx` | mockRecommendationResponse |
| `ResumeOptimizationDetail.tsx` | mockResumeOptimizationResponse |

## 8. token / auth 边界说明

- 当前页面不做 token 注入
- 无 token → API 返回 401 → api 层 `withFallback` 捕获 → 返回 mock 数据
- 页面不崩，展示 mock 数据
- 不写 localStorage/sessionStorage

## 9. mock fallback 保留

✅ 所有 api 层的 `withFallback` 机制完整保留，后端不可用时自动降级 mock。

## 10. loading / error / empty 状态

| 组件 | loading | error | empty |
|------|:--:|:--:|:--:|
| JobList | ✅ "正在加载岗位列表…" | ✅ "已切换到演示数据" | ✅ "暂无岗位数据" |
| JobDetailBanner | ✅ 返回 null（骨架） | ✅ 返回 null | — |
| JobDetailJD | ✅ 返回 null | ✅ 返回 null | — |
| SelectResumeModal | ✅ "正在加载简历列表…" | ✅ 空列表 | ✅ 空列表 |

## 11. build 结果

```
✓ built in 1.51s
```

**通过 ✅**

## 12. grep 验证

| 检查项 | 结果 |
|------|:--:|
| 普通业务页面/组件 import mocks | ✅ 零（全部切换） |
| Agent 页面/组件保留 mocks | ✅ 符合预期（8 处） |
| 页面/组件直接 fetch | ✅ 零 |
| 页面/组件 import api/client | ✅ 零 |
| localStorage/sessionStorage | ✅ 零 |
| axios | ✅ 零 |
| /api/v1 | ✅ 零 |

## 13. P0 / P1 / P2

| 等级 | 数量 | 说明 |
|:--:|:--:|------|
| P0 | 0 | — |
| P1 | 0 | — |
| P2 | 1 | ResumeManagementPage/MyResumesPage/ResumeComparisonPage 使用本地硬编码数据，需后续阶段切换 |

## 14. 越界检查

| 检查项 | 结果 |
|------|:--:|
| 修改 server | ❌ 零 |
| 修改 Makefile/docker/env | ❌ 零 |
| 修改 Agent API/files | ❌ 零 |

## 15. 本阶段未做事项

| # | 未做项 | 说明 |
|:--:|------|------|
| 1 | Agent 页面切换 | 留到阶段 7 |
| 2 | 简历管理页/比较页切换 | 硬编码数据，结构重构留到后续 |
| 3 | token 注入 | 无 token 则 fallback mock |
| 4 | 修改后端 | 禁止 |

## 16. 是否可以进入 6-5

**可以。** Jobs 页面（5 组件）和 SelectResumeModal 已从 mocks 切到 api viewModels，Agent 页面未误切，mock fallback 保留，build 通过，无 P0。

## 17. 最终结论

**pass ✅**

- JobList / JobDetailBanner / JobDetailJD / AIDecisionWorkspace / JobDetailAIDecision 已切换 ✅
- SelectResumeModal 已切换 ✅
- Agent 页面未误切 ✅
- mock fallback 保留 ✅
- loading/error/empty 最小状态到位 ✅
- build 1.51s 通过 ✅
- 零越界 ✅
