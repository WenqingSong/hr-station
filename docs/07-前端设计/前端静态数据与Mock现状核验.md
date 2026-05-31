# 前端静态数据与 Mock 现状核验

> 阶段：5-1.0 | 状态：read-only + docs | 日期：2026-05-31

## 一、执行摘要

当前前端 100% 硬编码 Mock 数据，不存在任何 API 调用、类型定义、Mock 数据目录或 Adapter 层。前端页面形态完整（9 页 + 25+ 组件），Navbar 冻结合规，路由完备，但**数据层完全为零**。

## 二、目录结构核验

| 目录 | 是否存在 | 说明 |
|------|:---:|------|
| `web/src/types/` | ❌ **不存在** | 无 TypeScript 接口类型定义 |
| `web/src/mocks/` | ❌ **不存在** | 无统一 Mock 数据文件 |
| `web/src/api/` | ❌ **不存在** | 无 API Client 封装 |
| `web/src/adapters/` | ❌ **不存在** | 无数据适配器层 |
| `web/src/lib/` | ❌ **不存在** | 无工具函数库 |
| `web/src/app/pages/` | ✅ 存在 | 9 个页面文件 |
| `web/src/app/components/` | ✅ 存在 | 25+ 组件（含 business/layout/ui） |
| `web/src/app/routes.tsx` | ✅ 存在 | React Router 7，10 条路由 |
| `web/src/app/App.tsx` | ✅ 存在 | RouterProvider 入口 |

## 三、全局状态总览

| 检查项 | 状态 | 说明 |
|--------|:---:|------|
| 真实路由 | ✅ 存在 | React Router 7，10 条路由 |
| 统一 Layout/Navbar | ✅ 存在 | routes.tsx 中 RootLayout |
| Navbar 是否符合冻结规则 | ✅ 合规 | 单条，左侧品牌 + 右侧"我的简历" |
| 是否有重复/错误 Navbar | 否 | Navbar 独立定义，Layout 版本未使用 |
| API 调用 | ❌ 无 | 零 fetch/axios/import.meta.env |
| TypeScript 类型文件 | ❌ 无 | 类型仅在组件内部内联定义 |
| Mock 数据目录 | ❌ 无 | 数据直接写在组件中 |
| Adapter 层 | ❌ 无 | 无数据转换层 |
| 环境变量 | ❌ 无 | 无 VITE_API_BASE_URL 等 |
| 消费后端 `data.result + data.meta` | ❌ 无 | 纯静态 Mock 展示 |
| fallback 机制 | ❌ 无 | 所有页面固定展示 Mock 数据 |

## 四、逐页硬编码数据核验

### 1. 岗位列表页 (`JobListPage.tsx`)

**数据来源**：`JobList.tsx` 内含 `JOB_MOCKS` 数组（6 个岗位硬编码）
**筛选条件**：`JobFilterSidebar.tsx` 内含 `FILTER_GROUPS`（4 组 17 个选项硬编码）
**AI 工作台**：`AIDecisionWorkspace.tsx` 全量硬编码 + "Mock 演示状态"标签

| 数据类型 | 硬编码内容 | 所在文件 | P |
|------|------|------|:--:|
| 岗位列表 | 6 个岗位（id/title/type/dept/skills/location/desc） | `JobList.tsx` | P0 |
| 筛选条件 | 4 组筛选（求职方向/地点/类型/类别） | `JobFilterSidebar.tsx` | P0 |
| 当前简历摘要 | V2 后端开发简历 + 标签 + 元数据 | `AIDecisionWorkspace.tsx` | P0 |
| 当前选中岗位 | 软件开发-后台开发方向 | `AIDecisionWorkspace.tsx` | P0 |
| 推荐入口文案 | "推荐最适合的 5 个岗位" | `AIDecisionWorkspace.tsx` | P1 |
| 简历风险提示 | 3 个影响因素动态列表 | `AIDecisionWorkspace.tsx` | P1 |
| AI 状态标签 | "Mock 演示状态" | `AIDecisionWorkspace.tsx` | P1 |

**结论**：岗位数据、筛选条件、简历摘要、AI 工作台数据全部硬编码在组件内部。无 API 调用。Navbar 合规。

---

### 2. 岗位详情页 (`JobDetailPage.tsx`)

**数据来源**：`JobDetailBanner.tsx` + `JobDetailJD.tsx` + `JobDetailAIDecision.tsx` 全部硬编码

| 数据类型 | 硬编码内容 | 所在文件 | P |
|------|------|------|:--:|
| 岗位标题/Banner | "软件开发-后台开发方向" + 标签 + 事业群/地点 | `JobDetailBanner.tsx` | P0 |
| JD 正文 | 岗位描述/要求/加分项/地点/投递说明 | `JobDetailJD.tsx` | P0 |
| 当前简历摘要 | V2 后端开发简历 + 标签 | `JobDetailAIDecision.tsx` | P0 |
| 岗位匹配分析入口 | "开始岗位匹配分析" 按钮 | `JobDetailAIDecision.tsx` | P1 |
| AI 状态标签 | "当前结果来源：Mock 演示" | `JobDetailAIDecision.tsx` | P1 |

**结论**：所有数据硬编码。无 API 调用。入口按钮导航正确。

---

### 3. 岗位推荐页 (`JobRecommendationPage.tsx`)

**数据来源**：`RecommendationSidebar.tsx` 内含 `RECOMMENDATIONS` 数组（5 个推荐岗位硬编码）

| 数据类型 | 硬编码内容 | 所在文件 | P |
|------|------|------|:--:|
| Top 5 推荐岗位 | 5 个岗位（id/rank/title/level/keywords/location/reason/risk） | `RecommendationSidebar.tsx` | P0 |
| 选中岗位分析 | 岗位信息 + "强推荐" + 推荐理由 3 条 | `RecommendationDetail.tsx` | P0 |
| 匹配证据表格 | 3 行（岗位要求/简历证据/证据强度） | `RecommendationDetail.tsx` | P0 |
| 主要风险 | 3 个风险项 | `RecommendationDetail.tsx` | P0 |
| 投递建议 | 2 句建议文案 | `RecommendationDetail.tsx` | P1 |
| 当前简历摘要 | V2 后端开发简历 + 标签 | `RecommendationContext.tsx` | P0 |
| 推荐依据 | 求职方向/核心技能/项目经历/目标岗位池 | `RecommendationContext.tsx` | P0 |
| 影响排序因素 | 3 个因素列表 | `RecommendationContext.tsx` | P1 |
| 页面标题 | "AI 岗位推荐" + 筛选范围/岗位池 | `RecommendationHeader.tsx` | P1 |

**结论**：未消费 RecommendationAgent 返回，无 `data.result + data.meta`，无 meta 状态标签。Mock 状态写在组件底部。

---

### 4. 岗位匹配分析页 (`JobMatchAnalysisPage.tsx`)

**数据来源**：全部 5 个组件硬编码

| 数据类型 | 硬编码内容 | 所在文件 | P |
|------|------|------|:--:|
| 匹配分数 | 82/100 | `MatchAnalysisDetail.tsx` | P0 |
| 投递准备度 | "先修复后投递" | `MatchAnalysisDetail.tsx` | P0 |
| 优化停止信号 | "修复 Top 3 后再投递" | `MatchAnalysisDetail.tsx` | P0 |
| Top 3 快速修复 | 3 个任务（标题/优先级/影响/核实） | `MatchAnalysisDetail.tsx` | P0 |
| JD 证据链表格 | 4 行（JD 要求/简历证据/强度/风险说明） | `MatchAnalysisDetail.tsx` | P0 |
| 优势匹配 | 3 个优势项 | `MatchAnalysisDetail.tsx` | P0 |
| 风险短板 | 3 个风险项 | `MatchAnalysisDetail.tsx` | P0 |
| 下一步建议 | 1 句话文案 + 3 个按钮 | `MatchAnalysisDetail.tsx` | P1 |
| 岗位 JD 摘要 | 岗位信息 + 核心要求 5 条 + JD 关键词 8 个 | `MatchAnalysisSidebarLeft.tsx` | P0 |
| 当前简历摘要 | V2 后端开发简历 + 核心证据 5 项 + 缺口 3 项 | `MatchAnalysisSidebarRight.tsx` | P0 |
| 页面标题 | "岗位匹配分析" + 目标岗位/简历/状态 | `MatchAnalysisHeader.tsx` | P1 |

**结论**：未消费 MatchAnalysisAgent 返回，无 `data.result + data.meta`。"针对该岗位优化简历"入口存在且正确。

---

### 5. 简历优化页 (`ResumeOptimizationPage.tsx`)

**数据来源**：`ResumeOptimizationDetail.tsx`（含可折叠 P1 模块）等全部硬编码

| 数据类型 | 硬编码内容 | 所在文件 | P |
|------|------|------|:--:|
| 优化结论 | 诊断摘要 | `ResumeOptimizationDetail.tsx` | P0 |
| 投递准备度 | "先修复后投递" | `ResumeOptimizationDetail.tsx` | P0 |
| 优化停止信号 | "修复 Top 3 后再投递" | `ResumeOptimizationDetail.tsx` | P0 |
| Top 3 快速修复 | 3 个任务（标题/优先级/影响/核实/desc） | `ResumeOptimizationDetail.tsx` | P0 |
| 需要核实信息清单 | 5 项核实条目 | `ResumeOptimizationDetail.tsx` | P0 |
| 修改前后对比示例 | 原简历表达 → 优化建议 → 优化后示例 | `ResumeOptimizationDetail.tsx` | P0 |
| JD 关键词对齐建议 (P1) | 5 行表格（可折叠） | `ResumeOptimizationDetail.tsx` | P1 |
| 项目经历优化建议 (P1) | 2 个项目详细建议（可折叠） | `ResumeOptimizationDetail.tsx` | P1 |
| 投递前确认 | 4 个确认项 + 3 个按钮 | `ResumeOptimizationDetail.tsx` | P0 |
| 岗位 JD 摘要 | 岗位信息 + 核心要求 + JD 关键词 | `ResumeOptimizationSidebarLeft.tsx` | P0 |
| 当前简历摘要 | V2 后端开发简历 + 核心内容 + 优化影响 | `ResumeOptimizationSidebarRight.tsx` | P0 |
| 页面标题 | "简历优化" + 目标岗位/简历/状态 | `ResumeOptimizationHeader.tsx` | P1 |

**结论**：P0 模块完整（优化结论/交付准备度/停止信号/Top3/核实清单/对比示例/确认清单），P1 模块以可折叠方式存在。未消费 ResumeOptimizationAgent 返回，无 meta 状态标签。**符合"投递前简历修复工作台"定位**。

---

### 6. 我的简历管理首页 (`ResumeManagementPage.tsx`)

**数据来源**：组件内部常量 `resumeVersions`、`skills`、`projects` 全部硬编码

| 数据类型 | 硬编码内容 | 所在文件 | P |
|------|------|------|:--:|
| 简历版本列表 | 3 个版本（v1/v2/v3） | `ResumeManagementPage.tsx` | P0 |
| 当前简历 | V2 后端开发简历 | `ResumeManagementPage.tsx` | P0 |
| 简历完整度 | 78% | `ResumeManagementPage.tsx` | P0 |
| 蓝色个人信息卡 | 姓名/年龄/学历/学校/专业/毕业/联系方式/城市/求职方向 | `ResumeManagementPage.tsx` | P0 |
| 求职意向 | 方向/类型/目标岗位/城市 | `ResumeManagementPage.tsx` | P0 |
| 教育经历 | 1 条 | `ResumeManagementPage.tsx` | P0 |
| 实习经历 | 1 条（ExampleTech） | `ResumeManagementPage.tsx` | P0 |
| 项目经历 | 3 个项目 | `ResumeManagementPage.tsx` | P0 |
| 技能信息 | 8 个技能 + 语言能力 | `ResumeManagementPage.tsx` | P0 |
| 获奖证书 | 2 项 | `ResumeManagementPage.tsx` | P1 |
| 自我评价 | 2 句话 | `ResumeManagementPage.tsx` | P0 |
| 操作按钮 | 修改简历/添加简历/历史比较 | `ResumeManagementPage.tsx` | P1 |
| AI 状态条 | "当前简历可用于" 提示 | `ResumeManagementPage.tsx` | P1 |

**结论**：字段符合冻结最小边界，无头像上传/附件上传/复杂字段。`SelectResumeModal` 已集成。页面形态优秀。

---

### 7. 简历编辑/添加页 (`MyResumesPage.tsx`)

**数据来源**：所有表单初始值通过 `useState()` 硬编码

| 数据类型 | 硬编码内容 | 所在文件 | P |
|------|------|------|:--:|
| 简历名称 | "V2 后端开发简历" | `MyResumesPage.tsx` | P0 |
| 求职方向 | "后端开发" | `MyResumesPage.tsx` | P0 |
| 基础信息 | 姓名/年龄/学历/学校/专业/毕业/手机/邮箱 | `MyResumesPage.tsx` | P0 |
| 教育经历 | 1 条（某大学/软件工程） | `MyResumesPage.tsx` | P0 |
| 实习经历 | 1 条（ExampleTech） | `MyResumesPage.tsx` | P0 |
| 项目经历 | 3 个项目 | `MyResumesPage.tsx` | P0 |
| 技能关键词 | 8 个 + 输入框 | `MyResumesPage.tsx` | P0 |
| 自我评价 | 2 句话 | `MyResumesPage.tsx` | P0 |
| 保存逻辑 | 纯前端 state（无 localStorage/API 持久化） | `MyResumesPage.tsx` | P0 |

**结论**：字段符合 Demo 最小字段（简历名称/方向/姓名/年龄/学历/学校/专业/毕业/手机/邮箱/教育/实习/项目/技能/评价）。**无头像上传、附件上传、复杂目标岗位、富文本编辑器**。保存按钮存在但无真实持久化。底部按钮 "保存并设为当前简历" 为假响应。

---

### 8. 选择简历弹窗 (`SelectResumeModal.tsx`)

**数据来源**：组件内部 `RESUME_OPTIONS` 数组硬编码

| 数据类型 | 硬编码内容 | 所在文件 | P |
|------|------|------|:--:|
| 简历列表 | 3 个选项（v2 推荐/v1 历史/v3 方向不匹配） | `SelectResumeModal.tsx` | P0 |
| 默认选中 | V2 后端开发简历 | `SelectResumeModal.tsx` | P0 |
| V3 方向不匹配 | "方向不匹配" 标签 + 橙色样式 | `SelectResumeModal.tsx` | P0 |
| 场景配置 | 2 个场景（recommendation/match-analysis） | `SelectResumeModal.tsx` | P1 |
| 引导文案 | "建议选择与目标岗位方向最接近..." | `SelectResumeModal.tsx` | P1 |

**结论**：弹窗符合冻结形态。已在 `ResumeManagementPage` 中被调用（"开始 AI 分析"按钮触发）。岗位推荐页和匹配分析页通过 `SelectResumeModal` 复用。**底部有 "管理我的简历" 弱链接**。

---

### 9. 历史简历比较页 (`ResumeComparisonPage.tsx`)

**数据来源**：组件内部全部硬编码

| 数据类型 | 硬编码内容 | 所在文件 | P |
|------|------|------|:--:|
| 版本对比选择 | V1 通用后端 vs V2 后端开发 | `ResumeComparisonPage.tsx` | P1 |
| 版本变化总览 | 3 个度量卡片 | `ResumeComparisonPage.tsx` | P1 |
| 新增内容 | 3 项 | `ResumeComparisonPage.tsx` | P1 |
| 强化内容 | 3 项 | `ResumeComparisonPage.tsx` | P1 |
| 关键词变化表 | 6 行关键词对比 | `ResumeComparisonPage.tsx` | P1 |
| 项目经历变化 | 3 个项目变化描述 | `ResumeComparisonPage.tsx` | P1 |
| 仍需补强内容 | 3 个缺口项 | `ResumeComparisonPage.tsx` | P1 |
| 建议结论 | 1 段文案 + 3 个按钮 | `ResumeComparisonPage.tsx` | P1 |
| 旧/新版本摘要 | 左右两栏 V1/V2 摘要 | `ResumeComparisonPage.tsx` | P1 |

**结论**：纯 P1 增强页，不新增 Agent。页面结构完整，数据全部硬编码。底部有固定 Action Bar。

---

## 五、硬编码数据汇总矩阵

| 页面/组件 | 文件路径 | 硬编码数据类型 | 当前问题 | 迁移目标 | P |
| ------- | ---- | ------- | ---- | ------ | :--: |
| 岗位列表 | `JobList.tsx` | 岗位数据(6条) | 组件内 `JOB_MOCKS` 数组 | `mocks/jobs.ts` | P0 |
| 岗位筛选 | `JobFilterSidebar.tsx` | 筛选条件(4组17项) | 组件内 `FILTER_GROUPS` 数组 | `mocks/jobs.ts` + `types/job.ts` | P0 |
| AI 工作台-简历 | `AIDecisionWorkspace.tsx` | 当前简历摘要 | 标签/名称/技能/项目数硬编码 | `mocks/resumes.ts` | P0 |
| AI 工作台-选中岗位 | `AIDecisionWorkspace.tsx` | 当前选中岗位信息 | 硬编码 title + labels | `types/job.ts` | P0 |
| AI 工作台-风险 | `AIDecisionWorkspace.tsx` | 简历风险因素(3条) | 硬编码列表 | `adapters/agentAdapter.ts` | P1 |
| 岗位详情 Banner | `JobDetailBanner.tsx` | 岗位标题/标签/信息 | 全部硬编码 | `mocks/jobs.ts` | P0 |
| 岗位详情 JD | `JobDetailJD.tsx` | JD 正文(5段) | 全部硬编码 | `mocks/jobs.ts` | P0 |
| 岗位详情 AI 工作台 | `JobDetailAIDecision.tsx` | 简历摘要/岗位信息 | 全部硬编码 | `mocks/resumes.ts` | P0 |
| Top5 推荐列表 | `RecommendationSidebar.tsx` | 5个推荐岗位 | 组件内 `RECOMMENDATIONS` | `mocks/recommendation.ts` | P0 |
| 推荐分析详情 | `RecommendationDetail.tsx` | 推荐理由/证据链/风险/建议 | 全部硬编码 | `mocks/recommendation.ts` | P0 |
| 推荐依据 | `RecommendationContext.tsx` | 简历摘要/推荐依据/风险因素 | 全部硬编码 | `mocks/recommendation.ts` | P0 |
| 推荐页头 | `RecommendationHeader.tsx` | 页面标题/简历/筛选范围 | 硬编码 | `adapters/agentAdapter.ts` | P1 |
| 匹配结论 | `MatchAnalysisDetail.tsx` | 匹配分数82/准备度/停止信号 | 全部硬编码 | `mocks/matchAnalysis.ts` | P0 |
| 匹配证据链 | `MatchAnalysisDetail.tsx` | JD 证据链表格(4行) | 全部硬编码 | `mocks/matchAnalysis.ts` | P0 |
| 匹配优势/风险 | `MatchAnalysisDetail.tsx` | 优势3项+风险3项 | 全部硬编码 | `mocks/matchAnalysis.ts` | P0 |
| 匹配 Top3 修复 | `MatchAnalysisDetail.tsx` | 3个快速修复任务 | 全部硬编码 | `mocks/matchAnalysis.ts` | P0 |
| 匹配左侧栏 | `MatchAnalysisSidebarLeft.tsx` | JD 摘要/关键词 | 全部硬编码 | `mocks/matchAnalysis.ts` | P0 |
| 匹配右侧栏 | `MatchAnalysisSidebarRight.tsx` | 简历摘要/证据/缺口 | 全部硬编码 | `mocks/resumes.ts` | P0 |
| 匹配页头 | `MatchAnalysisHeader.tsx` | 岗位/简历/分析状态 | 硬编码 | `adapters/agentAdapter.ts` | P1 |
| 优化结论 | `ResumeOptimizationDetail.tsx` | 诊断摘要/准备度/停止信号 | 全部硬编码 | `mocks/resumeOptimization.ts` | P0 |
| 优化 Top3 修复 | `ResumeOptimizationDetail.tsx` | 3个修复任务 | 全部硬编码 | `mocks/resumeOptimization.ts` | P0 |
| 优化对比示例 | `ResumeOptimizationDetail.tsx` | 修改前后对比 | 全部硬编码 | `mocks/resumeOptimization.ts` | P0 |
| 优化核实清单 | `ResumeOptimizationDetail.tsx` | 5项核实条目 | 全部硬编码 | `mocks/resumeOptimization.ts` | P0 |
| 优化确认清单 | `ResumeOptimizationDetail.tsx` | 4项确认+3按钮 | 全部硬编码 | `mocks/resumeOptimization.ts` | P0 |
| 优化关键词对齐(P1) | `ResumeOptimizationDetail.tsx` | 5行关键词表格 | 硬编码(可折叠) | `mocks/resumeOptimization.ts` | P1 |
| 优化项目建议(P1) | `ResumeOptimizationDetail.tsx` | 2个项目详细建议 | 硬编码(可折叠) | `mocks/resumeOptimization.ts` | P1 |
| 优化左侧栏 | `ResumeOptimizationSidebarLeft.tsx` | JD 摘要/关键词 | 全部硬编码 | `mocks/resumeOptimization.ts` | P0 |
| 优化右侧栏 | `ResumeOptimizationSidebarRight.tsx` | 简历摘要/核心内容 | 全部硬编码 | `mocks/resumes.ts` | P0 |
| 优化页头 | `ResumeOptimizationHeader.tsx` | 目标岗位/简历/状态 | 硬编码 | `adapters/agentAdapter.ts` | P1 |
| 简历管理首页 | `ResumeManagementPage.tsx` | 简历版本/个人信息/模块数据 | 全部硬编码 | `mocks/resumes.ts` | P0 |
| 简历编辑页 | `MyResumesPage.tsx` | 表单初始值(全部) | 全部 useState 硬编码 | `mocks/resumes.ts` | P0 |
| 选择简历弹窗 | `SelectResumeModal.tsx` | 3个简历选项 | 组件内 `RESUME_OPTIONS` | `mocks/resumes.ts` | P0 |
| 历史比较页 | `ResumeComparisonPage.tsx` | 全部比较数据 | 全部硬编码 | `mocks/resumes.ts` | P1 |
| 所有页面 | 全部 | AI meta 状态 | "Mock 演示状态" 硬编码 | `adapters/agentAdapter.ts` | P1 |
| 所有页面 | 全部 | 页面按钮文案 | 硬编码 | — | P2 |

## 六、缺失目录/文件清单

| 文件 | 当前存在 | 新增 | 建议阶段 | 说明 |
|------|:---:|:---:|------|------|
| `web/src/types/` | ❌ | 是 | 5-1.1 | 类型根目录 |
| `web/src/types/job.ts` | ❌ | 是 | 5-1.1 | 岗位相关类型 |
| `web/src/types/resume.ts` | ❌ | 是 | 5-1.1 | 简历相关类型 |
| `web/src/types/agent.ts` | ❌ | 是 | 5-1.1 | Agent 返回/props 类型 |
| `web/src/types/api.ts` | ❌ | 是 | 5-1.1 | API 请求/响应包装类型 |
| `web/src/types/index.ts` | ❌ | 是 | 5-1.1 | 统一导出 |
| `web/src/mocks/` | ❌ | 是 | 5-1.2 | Mock 数据根目录 |
| `web/src/mocks/jobs.ts` | ❌ | 是 | 5-1.2 | 岗位 Mock 数据 |
| `web/src/mocks/resumes.ts` | ❌ | 是 | 5-1.2 | 简历 Mock 数据 |
| `web/src/mocks/recommendation.ts` | ❌ | 是 | 5-1.2 | 岗位推荐 Mock |
| `web/src/mocks/matchAnalysis.ts` | ❌ | 是 | 5-1.2 | 匹配分析 Mock |
| `web/src/mocks/resumeOptimization.ts` | ❌ | 是 | 5-1.2 | 简历优化 Mock |
| `web/src/mocks/index.ts` | ❌ | 是 | 5-1.2 | Mock 统一导出 + selectors |
| `web/src/adapters/` | ❌ | 可延后 | 5-1.3 | Adapter 根目录 |
| `web/src/adapters/jobAdapter.ts` | ❌ | 可延后 | 5-1.3 | 岗位数据适配 |
| `web/src/adapters/recommendationAdapter.ts` | ❌ | 可延后 | 5-1.3 | 推荐结果适配 |
| `web/src/adapters/matchAnalysisAdapter.ts` | ❌ | 可延后 | 5-1.3 | 匹配分析适配 |
| `web/src/adapters/resumeOptimizationAdapter.ts` | ❌ | 可延后 | 5-1.3 | 简历优化适配 |
| `web/src/adapters/resumeAdapter.ts` | ❌ | 可延后 | 5-1.3 | 简历数据适配 |
| `web/src/api/` | ❌ | 是 | 后续 | API Client 根目录 |

## 七、页面与数据源现状表

| 页面 | 当前数据来源 | 统一 mock | 有类型 | 有 adapter | 接 API | 后续处理 |
|------|------|:---:|:---:|:---:|:---:|------|
| 岗位列表页 | 组件内 `JOB_MOCKS` + `FILTER_GROUPS` | ❌ | ❌ | ❌ | ❌ | P0 标准化 |
| 岗位详情页 | 组件内全部硬编码 | ❌ | ❌ | ❌ | ❌ | P0 标准化 |
| 岗位推荐页 | 组件内 `RECOMMENDATIONS` + 全部硬编码 | ❌ | ❌ | ❌ | ❌ | P0 标准化 |
| 岗位匹配分析页 | 全部组件内硬编码 | ❌ | ❌ | ❌ | ❌ | P0 标准化 |
| 简历优化页 | 全部组件内硬编码 | ❌ | ❌ | ❌ | ❌ | P0 标准化 |
| 我的简历管理首页 | 组件内常量硬编码 | ❌ | ❌ | ❌ | ❌ | P0 标准化 |
| 简历编辑/添加页 | useState 初始值硬编码 | ❌ | ❌ | ❌ | ❌ | P0 标准化 |
| 选择简历弹窗 | 组件内 `RESUME_OPTIONS` | ❌ | ❌ | ❌ | ❌ | P0 标准化 |
| 历史简历比较页 | 组件内全部硬编码 | ❌ | ❌ | ❌ | ❌ | P1 增强 |

## 八、5-1 后续拆分建议

### 5-1.1 TypeScript 数据契约设计与落地

- **目标**：建立 `web/src/types/` 目录，定义 `job.ts`、`resume.ts`、`agent.ts`、`api.ts`、`index.ts`
- **是否改前端**：是（新增 types 文件，不改页面）
- **是否改后端**：否
- **主要文件**：新文件 5 个
- **验收标准**：TypeScript 编译通过，类型覆盖所有页面所需数据域
- **禁止范围**：不改页面组件，不改后端

### 5-1.2 Mock 数据标准化

- **目标**：建立 `web/src/mocks/` 目录，创建 `jobs.ts`、`resumes.ts`、`recommendation.ts`、`matchAnalysis.ts`、`resumeOptimization.ts`、`index.ts`
- **是否改前端**：是（迁移硬编码数据为 mock 文件，页面导入 mock）
- **是否改后端**：否
- **主要文件**：新文件 6 个 + 修改 9 个页面
- **验收标准**：所有页面从统一 mock 文件导入数据，无重复硬编码
- **禁止范围**：不改后端，不接 API

### 5-1.3 Adapter/Selector 标准化

- **目标**：建立 `web/src/adapters/`（可延后到页面接入 API 前）
- **是否改前端**：是（新增 adapter 文件）
- **是否改后端**：否
- **主要文件**：新文件 5 个
- **验收标准**：adapter 函数可将 API 返回转换为组件 props
- **禁止范围**：不改后端，不接真实 API

### 5-1.4 页面硬编码数据替换为统一 mocks

- **目标**：将所有 9 页面的硬编码数据替换为从统一 mocks 导入
- **是否改前端**：是（修改所有页面）
- **是否改后端**：否
- **主要文件**：9 个页面文件 + 25+ 组件文件
- **验收标准**：页面视觉不变，数据源从内联变为 mocks/ 导入
- **禁止范围**：不改后端，不接 API

### 5-1.5 5-1 验收收口

- **目标**：docs only，核验全部标准化结果
- **是否改前端**：否
- **是否改后端**：否
- **主要文件**：docs 文件
- **验收标准**：types/mocks/adapter 完整，页面可独立运行
- **禁止范围**：不改代码

## 九、P0/P1 边界

### P0（5-1 必须完成）

- `web/src/types/` 建立：`job.ts` / `resume.ts` / `agent.ts` / `api.ts` / `index.ts`
- `web/src/mocks/` 建立：`jobs.ts` / `resumes.ts` / `recommendation.ts` / `matchAnalysis.ts` / `resumeOptimization.ts` / `index.ts`
- 所有页面从 `mocks/` 导入数据替代组件内硬编码
- `data.result + data.meta` 的 TypeScript 类型定义

### P1（可延后）

- `web/src/adapters/`（在页面接 API 前需要）
- `web/src/api/`（在接真实后端前需要）
- 历史简历比较页 mock 标准化
- 简历编辑页保存持久化（localStorage mock）
- 真实简历 CRUD
- 用户登录态

## 十、关键观察

1. **页面形态优秀**：9 页均从 Figma Make 生成，与 5-UX.0.1-A 冻结形态完全一致
2. **Navbar 清洁**：单条，左侧品牌 + 右侧"我的简历"，无任何禁止元素
3. **路由完备**：10 条路由覆盖全部页面场景，带 404 兜底
4. **组件拆分好**：business/ 下 6 个可复用组件已定义但未在页面中使用（仍在内联）
5. **数据层完全为零**：types/api/mocks/adapters 全不存在
6. **无任何网络调用**：零 fetch/axios
7. **Mock 标签已标注**：多数页面上有 "Mock 演示状态" 水印
8. **反模式**：岗位 JD/JD 关键词在 3 个不同页面/组件中重复硬编码（JobDetailBanner/MatchAnalysisSidebarLeft/ResumeOptimizationSidebarLeft），迁移到 mocks 后此问题将消失
