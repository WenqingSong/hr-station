# 全局布局 / Navbar / 路由现状核验

> 阶段：5-3.0 | 状态：已完成 | 日期：2026-05-31

## 1. 本阶段目标

在正式修改 Layout / Navbar / 页面骨架前，只读核验当前前端真实结构，确认 App 入口、路由、Navbar、页面骨架状态，输出后续 5-3.1~5-3.5 拆分建议。

## 2. 当前 App / 路由 / 页面切换现状

### 入口链路

```text
main.tsx → App.tsx → RouterProvider → createBrowserRouter(routes)
```

- ✅ 使用 `react-router`（react-router v7），非 useState 状态切换
- ✅ App.tsx 仅一行 `<RouterProvider router={router} />`，无多余逻辑
- ✅ routes.tsx 中 RootLayout 包裹 Global Navbar + `<Outlet />`
- ✅ 9 个页面路由，含 catch-all `*` → redirect to `/`

### 路由表 vs 冻结规范

| 当前路由 | 当前页面 | 冻结路由 | 差异 |
|------|------|------|------|
| `/` (index) | JobListPage | `/jobs` | ✅ 默认首页 |
| `/jobs` | JobListPage | `/jobs` | ✅ 一致 |
| `/job/:id` | JobDetailPage | `/jobs/:id` | ⚠️ 单数 vs 复数（`job/:id` vs `jobs/:id`） |
| `/recommendations` | JobRecommendationPage | `/recommend` | ❌ 不一致 |
| `/match-analysis` | JobMatchAnalysisPage | `/match` | ❌ 不一致 |
| `/resume-optimization` | ResumeOptimizationPage | `/resume-optimization` | ✅ 一致 |
| `/my-resumes` | ResumeManagementPage | `/my-resumes` | ✅ 一致 |
| `/my-resumes/edit` | MyResumesPage | `/my-resumes/edit` | ✅ 一致 |
| `/my-resumes/compare` | ResumeComparisonPage | `/resume-comparison` | ⚠️ 路径不一致 |
| `/resume-comparison` | redirect → `/my-resumes/compare` | — | ✅ 兼容重定向 |

## 3. Navbar 现状

### Navbar 文件清单

| 文件 | 状态 | 是否接入路由 |
|------|------|:--:|
| `web/src/app/components/Navbar.tsx` | **正在使用**（routes.tsx RootLayout 引入） | ✅ |
| `web/src/app/components/layout/Navbar.tsx` | **冗余副本**（layout/Layout.tsx 引入，但 Layout 未接入路由） | ❌ 未接入 |
| `web/src/app/components/layout/Layout.tsx` | **孤立组件**（含自己的 Navbar + Outlet 容器，但未在任何路由中使用） | ❌ 未接入 |

### 当前 Navbar 内容

**`components/Navbar.tsx`（唯一实际生效的 Navbar）**：

```
左侧：「Offer 捕手」学生求职匹配智能体
右侧：我的简历
```

**`components/layout/Navbar.tsx`**：完全相同的代码副本。

## 4. Navbar 与冻结规则差异

| 规则 | 当前状态 | 结论 |
|------|------|:--:|
| 左侧：「Offer 捕手」学生求职匹配智能体 | ✅ 已实现 | pass |
| 右侧：我的简历 | ✅ 已实现 | pass |
| 禁止：岗位浏览 | ❌ 未出现 | pass |
| 禁止：岗位推荐 | ❌ 未出现 | pass |
| 禁止：首页 | ❌ 未出现 | pass |
| 禁止：查公司 / 招聘活动 / 关键问答 | ❌ 未出现 | pass |
| 禁止：实习生 | ❌ 未出现 | pass |
| 禁止：中文/En / 帮助 / Beta | ❌ 未出现 | pass |
| 禁止：头像 / 用户名 / 下拉菜单 | ❌ 未出现 | pass |
| Navbar 只允许一条 | ⚠️ 代码中存在两套 Navbar 源码，但路由只加载一套 | **partial_pass** |
| 无两层 Navbar | ✅ 只有一条 | pass |

### Navbar 差异矩阵

| 文件 | 当前 Navbar 内容 | 是否符合冻结规则 | 问题 | 后续修复阶段 |
|------|------------|:--:|------|------|
| `components/Navbar.tsx` | 左侧：品牌 / 右侧：我的简历 | ✅ | 无 | — |
| `components/layout/Navbar.tsx` | 左侧：品牌 / 右侧：我的简历（副本） | ✅ 内容正确 | 冗余副本，未接入路由但占据代码空间 | 5-3.1 清理 |
| `components/layout/Layout.tsx` | 含 Navbar + Outlet | — | 未接入路由的孤立组件 | 5-3.1 清理 |

## 5. 页面清单与入口方式

| 页面 | 当前文件 | 当前入口方式 | 是否符合冻结形态 | 主要问题 | 后续阶段 |
|------|------|------|:--:|------|------|
| 岗位列表页 | `pages/JobListPage.tsx` + 3组件 | 路由 `/jobs` + Navbar 品牌 Link | ✅ | 无 | — |
| 岗位详情页 | `pages/JobDetailPage.tsx` + 3组件 | 路由 `/job/:id` | ✅ | 路由单数 vs 复数 | 5-3.2 路由规范化 |
| 岗位推荐页 | `pages/JobRecommendationPage.tsx` + 4组件 | 路由 `/recommendations` | ⚠️ | 路由与冻结规范不一致 | 5-3.2 路由规范化 |
| 岗位匹配分析页 | `pages/JobMatchAnalysisPage.tsx` + 4组件 | 路由 `/match-analysis` | ⚠️ | 路由与冻结规范不一致 | 5-3.2 路由规范化 |
| 简历优化页 | `pages/ResumeOptimizationPage.tsx` + 4组件 | 路由 `/resume-optimization` | ✅ | 路由一致 | — |
| 我的简历管理首页 | `pages/ResumeManagementPage.tsx` | 路由 `/my-resumes` + Navbar 我的简历 | ✅ | 无 | — |
| 简历编辑/添加页 | `pages/MyResumesPage.tsx` | 路由 `/my-resumes/edit` | ✅ | 单列长表单 ✅ | — |
| 选择简历弹窗 | `components/SelectResumeModal.tsx` | 弹窗组件 | ✅ | 作为弹窗正确 | — |
| 历史简历比较页 | `pages/ResumeComparisonPage.tsx` | 路由 `/my-resumes/compare` + `/resume-comparison` 重定向 | ⚠️ | 路由路径与冻结不严格一致 | 5-3.2 路由规范化 |

## 6. 页面关系核验

| 流程 | 当前是否可达 | 当前实现方式 | 是否符合冻结关系 | 问题 | 修复建议 |
|------|:--:|------|:--:|------|------|
| 岗位列表 → 岗位推荐 | ✅ | JobList 右侧 AIDecisionWorkspace 按钮 → `/recommendations` | ✅ | 路由需规范化为 `/recommend` | 5-3.2 |
| 岗位列表 → 岗位详情 | ✅ | JobCard 点击 → `/job/:id` | ✅ | 路由需规范化为 `/jobs/:id` | 5-3.2 |
| 岗位详情 → 岗位匹配分析 | ✅ | 右侧 JobDetailAIDecision 按钮 → `/match-analysis` | ✅ | 路由需规范化为 `/match` | 5-3.2 |
| 岗位匹配分析 → 简历优化 | ✅ | MatchAnalysisDetail 按钮 → `/resume-optimization` | ✅ | 无 | — |
| Navbar 我的简历 → 我的简历管理首页 | ✅ | Link → `/my-resumes` | ✅ | 无 | — |
| 我的简历管理 → 修改简历 | ✅ | Link → `/my-resumes/edit` | ✅ | 无 | — |
| 我的简历管理 → 添加简历 | ✅ | 同 → `/my-resumes/edit` | ✅ | 无 | — |
| 我的简历管理 → 历史比较 | ✅ | Link → `/my-resumes/compare` | ✅ | 无 | — |
| 分析前 → 选择简历弹窗 | ✅ | SelectResumeModal（在 ResumeManagementPage 内） | ⚠️ | 弹窗仅在简历管理页内，岗位推荐/匹配分析页未接入 | 5-3.3/5-3.4 接入 |

## 7. 核心业务页骨架核验

### 7.1 岗位列表页

| 验收项 | 状态 |
|------|:--:|
| 左侧筛选（JobFilterSidebar） | ✅ |
| 中间岗位列表（JobList） | ✅ |
| 右侧 AI 决策工作台（AIDecisionWorkspace） | ✅ |
| 右侧突出"推荐最适合的 5 个岗位" | ✅ |
| 未错误出现岗位匹配分析主按钮 | ✅ |
| 未错误出现简历优化主按钮 | ✅ |
| 无重复 Navbar | ✅ |
| 风险：JobFilterSidebar 含"实习生"筛选标签（不在 Navbar 但仍需关注） | ⚠️ P2 |

### 7.2 岗位详情页

| 验收项 | 状态 |
|------|:--:|
| 顶部蓝色岗位 Banner（JobDetailBanner） | ✅ |
| 中间岗位 JD 正文（JobDetailJD） | ✅ |
| 右侧匹配分析入口（JobDetailAIDecision） | ✅ |
| 右侧突出"开始岗位匹配分析" | ✅ |
| 未错误出现岗位推荐主按钮 | ✅ |
| 未错误出现简历优化主按钮 | ✅ |
| 无重复 Navbar | ✅ |

### 7.3 岗位推荐页

| 验收项 | 状态 |
|------|:--:|
| 三栏结构 | ✅ |
| 左侧 Top 5 推荐岗位（RecommendationSidebar） | ✅ |
| 中间 AI 推荐分析（RecommendationDetail） | ✅ |
| 右侧当前简历摘要（RecommendationContext） | ✅ |
| 显示推荐理由 / 风险 / 下一步 | ✅ |
| 无重复 Navbar | ✅ |

### 7.4 岗位匹配分析页

| 验收项 | 状态 |
|------|:--:|
| 三栏结构 | ✅ |
| 左侧岗位 JD 摘要（MatchAnalysisSidebarLeft） | ✅ |
| 中间 AI 匹配分析结果（MatchAnalysisDetail） | ✅ |
| 右侧当前简历摘要（MatchAnalysisSidebarRight） | ✅ |
| 显示匹配分数（82/100） | ✅ |
| 显示投递准备度 | ✅ |
| 显示 Top 3 修复任务 | ✅ |
| 显示 JD 证据链 | ✅ |
| "针对该岗位优化简历"按钮 | ✅ |
| 无重复 Navbar | ✅ |

### 7.5 简历优化页

| 验收项 | 状态 |
|------|:--:|
| 三栏结构 | ✅ |
| 左侧目标岗位 JD（ResumeOptimizationSidebarLeft） | ✅ |
| 中间 AI 简历优化工作台（ResumeOptimizationDetail） | ✅ |
| 右侧当前简历摘要（ResumeOptimizationSidebarRight） | ✅ |
| 显示 Top 3 修复任务 | ✅ |
| 显示关键词对齐 | ✅ |
| 显示项目经历优化建议 | ✅ |
| 保留"AI 不替用户编造经历"边界 | ✅ |
| 无重复 Navbar | ✅ |

## 8. 简历模块骨架核验

### 8.1 我的简历管理首页（ResumeManagementPage）

| 验收项 | 状态 |
|------|:--:|
| 腾讯校园招聘风格详情页 | ✅ |
| 非 Dashboard 化 | ✅ |
| 非三栏看板 | ✅ |
| 无复杂腾讯招聘字段 | ✅ |
| 修改/添加/历史比较入口 | ✅ |
| 无重复 Navbar | ✅ |
| 风险：使用 3 个 `.tsx` 版本（`components/Navbar.tsx` / 含硬编码数据 `/pages/ResumeManagementPage.tsx`） | ⚠️ P1 |

### 8.2 简历编辑/添加页（MyResumesPage）

| 验收项 | 状态 |
|------|:--:|
| 单列长表单 | ✅ |
| 基础信息/求职方向/教育/实习/项目/技能/自评 | ✅ |
| 无头像上传 | ✅ |
| 无附件上传 | ✅ |
| 无富文本编辑器 | ✅ |
| 无复杂身份字段 | ✅ |

### 8.3 选择简历弹窗（SelectResumeModal）

| 验收项 | 状态 |
|------|:--:|
| 作为弹窗（非完整页面） | ✅ |
| 默认 V2 | ✅ |
| V3 不建议 | ✅ |
| 取消/确认并开始分析按钮 | ✅ |
| 被岗位推荐和岗位匹配分析复用 | ⚠️ 仅内嵌于 ResumeManagementPage（`useState(false)`），岗位推荐/匹配分析等页面未接入弹窗 | P0—5-3.3 |

### 8.4 历史简历比较页（ResumeComparisonPage）

| 验收项 | 状态 |
|------|:--:|
| 三栏结构（旧版本 / 差异 / 新版本） | ✅ |
| 非新增 Agent | ✅ |
| 非代码 diff | ✅ |
| 无复杂编辑器 | ✅ |

## 9. 页面骨架差异矩阵

| 页面 | 当前骨架 | 冻结骨架 | 差异 | 风险等级 | 修复阶段 |
|------|----|----|------|:--:|------|
| 岗位列表页 | 三栏（筛选 + 列表 + AI 工作台） | 三栏 | 基本一致 | P2 | — |
| 岗位详情页 | Banner + 双栏（JD + AI 工作台） | Banner + 双栏 | 基本一致 | P2 | — |
| 岗位推荐页 | 三栏 | 三栏 | 基本一致 | P2 | 5-3.3 |
| 岗位匹配分析页 | 三栏 | 三栏 | 基本一致 | P2 | 5-3.3 |
| 简历优化页 | 三栏 | 三栏 | 基本一致 | P2 | 5-3.3 |
| 我的简历管理首页 | 单列详情页 | 单列详情页 | 基本一致 | P1 | 5-3.4 |
| 简历编辑/添加页 | 单列长表单 | 单列长表单 | 基本一致 | P1 | 5-3.4 |
| 选择简历弹窗 | 弹窗（仅简历管理页内） | 全局可复用弹窗 | 未全局接入 | **P0** | 5-3.3 |
| 历史简历比较页 | 三栏 | 三栏 | 基本一致 | P1 | 5-3.4 |

## 10. 重复 Navbar / Header 清单

| 文件 | 重复类型 | 当前表现 | P0 | 处理建议 |
|------|----|----|:--:|----|
| `components/layout/Navbar.tsx` | 冗余 Navbar 副本 | 与 `components/Navbar.tsx` 完全一致的代码，被 `layout/Layout.tsx` 引用但未接入路由 | 否 | 5-3.1 删除 |
| `components/layout/Layout.tsx` | 冗余 Layout 组件 | 定义了含 Navbar 的布局但未在路由中使用 | 否 | 5-3.1 删除 |
| `components/layout/AIDecisionWorkspace.tsx` | 冗余 AI 工作台副本 | `components/AIDecisionWorkspace.tsx` 的 1:1 副本 | 否 | 5-3.1 删除 |
| `components/layout/Sidebar.tsx` | 孤立筛选组件 | 与 JobFilterSidebar 相似的筛选栏但未接入 | 否 | 5-3.1 删除 |
| 各页面 `min-h-screen` | 页面内重复全屏容器 | 每个页面各自声明 `min-h-screen bg-[...]`，与 RootLayout 的 `min-h-screen flex-col` 形成嵌套 | 否 | 5-3.2 统一 |

## 11. 5-3 后续拆分建议

### 5-3.1 Layout / Navbar 基础组件统一

- **目标**：删除冗余 Navbar 副本和未使用 Layout 组件，统一全局布局
- **是否改前端**：是（删除文件）
- **是否改后端**：否
- **主要文件**：
  - 删除 `components/layout/Navbar.tsx`
  - 删除 `components/layout/Layout.tsx`
  - 删除 `components/layout/AIDecisionWorkspace.tsx`
  - 删除 `components/layout/Sidebar.tsx`
  - 确认 `components/Navbar.tsx` 为唯一 Navbar
- **验收标准**：find web/src -type f | grep -Ei "nav|navbar" 只返回 `components/Navbar.tsx`；layout/ 目录可清除
- **禁止范围**：不改页面、不改后端、不改 API

### 5-3.2 页面路由 / 页面状态骨架整理

- **目标**：路由路径规范化对齐冻结规范；统一页面容器结构
- **是否改前端**：是
- **是否改后端**：否
- **主要文件**：
  - `routes.tsx`：`/job/:id` → `/jobs/:id`，`/recommendations` → `/recommend`，`/match-analysis` → `/match`
  - 各页面：移除各自的 `min-h-screen` 外层容器
- **验收标准**：路由表对齐冻结规范；页面不再各自设全屏高度
- **禁止范围**：不改后端、不改 types/mocks/adapters/api

### 5-3.3 核心业务页骨架统一

- **目标**：SelectResumeModal 接入岗位推荐页和岗位匹配分析页；统一三栏布局 Standard3ColLayout 组件
- **是否改前端**：是
- **是否改后端**：否
- **主要文件**：
  - JobRecommendationPage → 集成 SelectResumeModal
  - JobMatchAnalysisPage → 集成 SelectResumeModal
  - 可选：提取 Standard3ColLayout 通用骨架
- **验收标准**：岗位推荐、匹配分析前弹出选择简历弹窗；三栏页面使用统一骨架
- **禁止范围**：不改后端、不改数据层

### 5-3.4 简历模块骨架统一

- **目标**：ResumeManagementPage 接入 mocks+adapters 替换内联硬编码
- **是否改前端**：是
- **是否改后端**：否
- **主要文件**：
  - `pages/ResumeManagementPage.tsx`（替换 `position`、`skills`、`projects` 硬编码为 mocks+adapters）
  - `pages/ResumeComparisonPage.tsx`（确认骨架完整）
- **验收标准**：简历管理页数据来自统一 mocks+adapters
- **禁止范围**：不改后端、不改 API

### 5-3.5 全局布局 / Navbar / 页面骨架验收

- **目标**：验收 5-3 全阶段成果
- **是否改前端**：否（docs only）
- **是否改后端**：否
- **验收标准**：Navbar 唯一且符合冻结规则；路由规范化；SelectResumeModal 全局可复用；简历模块直连 mocks
- **禁止范围**：不改代码

## 12. 本阶段未做事项

```text
未修改前端代码
未删除不使用的组件
未修改 Navbar
未修改路由
未修改页面结构
未启动项目
未调用真实 LLM
```

## 13. 是否可以进入 5-3.1

**可以。**

当前 Navbar 内容已完全符合冻结规则（左侧品牌 + 右侧我的简历），路由体系完整（react-router + 9 页面），核心业务 5 页骨架均三栏且无重复 Navbar。主要待处理问题：

1. **代码冗余**（P1）：两套 Navbar 源码、未使用 Layout/Sidebar 组件 → 5-3.1 清理
2. **路由不一致**（P1）：`/recommendations`→`/recommend`、`/match-analysis`→`/match`、`/job/:id`→`/jobs/:id` → 5-3.2 规范
3. **SelectResumeModal 未全局接入**（P0）：仅在简历管理页内，岗位推荐/匹配分析页需接入 → 5-3.3

## 14. 验收结论

**pass** — 全局布局 / Navbar / 路由 / 页面骨架现状核验完成，已定位所有差异点和冗余组件，具备进入 5-3.1 条件。
