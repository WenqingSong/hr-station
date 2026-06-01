# 开发日志 #186：页面数据源从 Mocks 切到 API ViewModels

> 日期：2026-06-01 | 阶段：6-4 | 状态：已完成

## 做了什么

1. 定位全部页面/组件 mock 依赖（grep + 逐文件分析）
2. 切换 6 个普通业务组件从直接 mocks 到 api viewModels loader：
   - `JobList.tsx`：mockJobs → loadJobListViewModels
   - `JobDetailBanner.tsx`：mockSelectedJob → loadSelectedJobDetailViewModel
   - `JobDetailJD.tsx`：mockSelectedJob → loadSelectedJobDetailViewModel
   - `AIDecisionWorkspace.tsx`：mockCurrentResume/mockSelectedJob → loaders
   - `JobDetailAIDecision.tsx`：mockCurrentResume/mockSelectedJob → loaders（支持路由 jobId）
   - `SelectResumeModal.tsx`：mockResumeOptions → loadResumeOptionViewModels
3. 添加最小 loading/error/empty 状态渲染
4. 确认 Agent 页面/组件 8 处保留 mock（阶段 7）
5. `npm run build` 1.51s 通过
6. grep 验证全部通过
7. 越界检查零违规

## 修改的文件

- `web/src/app/components/JobList.tsx`
- `web/src/app/components/JobDetailBanner.tsx`
- `web/src/app/components/JobDetailJD.tsx`
- `web/src/app/components/AIDecisionWorkspace.tsx`
- `web/src/app/components/JobDetailAIDecision.tsx`
- `web/src/app/components/SelectResumeModal.tsx`

## grep 验证

- 普通业务页面/组件 import mocks：**零**（全部切换）
- Agent 页面/组件保留 mocks：8 处（符合预期）
- fetch / api/client / localStorage / axios / /api/v1：**零**

## 未切换

- ResumeManagementPage / MyResumesPage / ResumeComparisonPage：使用本地硬编码数据，需后续阶段重构
- Agent 页面（JobRecommendation / JobMatchAnalysis / ResumeOptimization）：留到阶段 7

## 文件变更

- 修改：web/src/app/components/ (6 文件)
- 新增：docs/08-后端联调/页面数据源从Mocks切到APIViewModels.md
- 新增：docs/09-开发日志/2026-06-01-186-页面数据源从Mocks切到APIViewModels.md
- 更新：docs/项目状态.md / 下一步任务.md / 变更记录.md / 说明书.md

## 下一步

6-5 Loading / Error / Empty 状态补齐
