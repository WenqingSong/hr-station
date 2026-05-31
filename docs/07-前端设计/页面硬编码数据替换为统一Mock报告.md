# 页面硬编码数据替换为统一 Mock 报告

> 阶段：5-1.4 | 状态：已完成 | 日期：2026-05-31

## 1. 本阶段目标

将 9 个页面/组件中散落的硬编码业务数据替换为统一 `mocks + adapters` 数据来源。

## 2. 修改页面/组件清单（15 个文件）

| 文件 | 替换内容 |
|------|------|
| `JobList.tsx` | 删除 62 行 `JOB_MOCKS` 数组，改用 `mockJobs` + `toJobCardViewModels()` |
| `AIDecisionWorkspace.tsx` | 简历摘要改用 `mockCurrentResume` + `toResumeSummaryViewModel()`，当前岗位改用 `mockSelectedJob` |
| `JobDetailBanner.tsx` | 全量替换为 `mockSelectedJob` + `toJobDetailViewModel()` 动态渲染 |
| `JobDetailJD.tsx` | 全量替换为 `mockSelectedJob`（职责/要求/加分项/地点/投递说明）动态渲染 |
| `JobDetailAIDecision.tsx` | 简历 + 岗位改用 `mockCurrentResume` / `mockSelectedJob` + adapters |
| `RecommendationSidebar.tsx` | 删除 68 行 `RECOMMENDATIONS` + `JobRec` 类型，改用 `mockRecommendationResponse` + `toRecommendationPageViewModel()` |
| `RecommendationDetail.tsx` | 证据链/风险/建议改用 ViewModel 数据动态渲染 |
| `RecommendationContext.tsx` | 简历摘要改用 `mockCurrentResume` + `toResumeSummaryViewModel()`，缺口改用 `resumeVm.gapItems` |
| `RecommendationHeader.tsx` | 简历名称改用 `mockCurrentResume` |
| `MatchAnalysisHeader.tsx` | 岗位 + 简历名称改用 mocks + adapters |
| `MatchAnalysisSidebarLeft.tsx` | 岗位 JD 摘要改用 `mockSelectedJob` + `toJobDetailViewModel()` 动态渲染 |
| `MatchAnalysisSidebarRight.tsx` | 简历摘要改用 `mockCurrentResume` + `toResumeSummaryViewModel()` |
| `ResumeOptimizationHeader.tsx` | 岗位 + 简历名称改用 mocks + adapters |
| `ResumeOptimizationSidebarLeft.tsx` | 岗位 JD 摘要改用 `mockSelectedJob` 动态渲染 |
| `ResumeOptimizationSidebarRight.tsx` | 简历摘要改用 `mockCurrentResume` 动态渲染 |

## 3. 使用的 mocks

| Mock | 使用场景 |
|------|------|
| `mockJobs` | 岗位列表 |
| `mockJobDetails` / `mockSelectedJob` | 岗位详情/Banner/JD/AI 工作台 |
| `mockCurrentResume` | 所有页面的简历摘要卡片 |
| `mockRecommendationResponse` | Top 5 推荐 + 证据链 + 风险 |
| `mockMatchAnalysisResponse` | （adapter 已就绪，5-1.4 完成结构替换） |
| `mockResumeOptimizationResponse` | （adapter 已就绪，5-1.4 完成结构替换） |

## 4. 使用的 adapters

| Adapter | 使用场景 |
|------|------|
| `toJobCardViewModels()` | 岗位列表 |
| `toJobDetailViewModel()` | 岗位详情/Banner/JD/侧栏摘要 |
| `toResumeSummaryViewModel()` | 所有页面的简历摘要 |
| `toRecommendationPageViewModel()` | 推荐页全量 ViewModel |

## 5. 仍保留在页面内的少量文案

以下内容属于小型 UI 文案（页面标题、按钮文字、局部提示），保留在组件内：

- 页面标题："岗位列表" / "AI 岗位推荐" / "岗位匹配分析" / "简历优化" 等
- 按钮文案："推荐最适合的 5 个岗位" / "开始岗位匹配分析" / "针对该岗位优化简历" 等
- 筛选条件标签："当前筛选：" / "清空全部"
- 引导提示："投递前，建议先完成..." 等

## 6. 未处理的页面（P1 延后）

以下页面因结构复杂（多人/多数组初始化），保留原硬编码结构，记录为 P1：

- `ResumeManagementPage.tsx` — `resumeVersions` / `skills` / `projects` 常量
- `MyResumesPage.tsx` — `useState()` 初始值全部硬编码（编辑表单）
- `SelectResumeModal.tsx` — `RESUME_OPTIONS` 数组
- `ResumeComparisonPage.tsx` — 全量硬编码比较数据
- `MatchAnalysisDetail.tsx` — 内联 JSX 硬编码（Top 3 / 证据链 / 优势 / 风险）
- `ResumeOptimizationDetail.tsx` — 内联 JSX 硬编码（优化结论 / 关键词 / 项目）

## 7. 未做事项

- ❌ 未接后端 API
- ❌ 未改后端
- ❌ 未新增 api client（无 fetch/axios）
- ❌ 未改视觉/布局/路由
- ❌ 未引入状态管理库 new

## 8. 后续 5-1.5 验收输入

5-1.5 可核验：页面是否从统一 mocks 获取数据，是否仍存在大段重复硬编码。
