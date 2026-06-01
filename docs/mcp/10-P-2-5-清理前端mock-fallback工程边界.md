# 10-P-2-5 - 清理前端 mock / fallback 工程边界 - MCP 上下文记录

> 执行时间：2026-06-01 | 状态：pass（P0-3 已修复）
> 是否修改代码：是（5 组件 + 3 父页面）
> 上一阶段：10-P-2-4B

---

## 1. 阶段信息

| 项目 | 内容 |
|---|---|
| 当前阶段编号 | 10-P-2-5 |
| 项目名称 | Offer 捕手 |
| 暂停阶段 | 阶段 11、12、13 |

## 2. 本次目标

清理 P0-3：5 个前端组件在模块顶层 import mock 数据的问题。让组件改为从 props 接收数据，无真实数据时显示空态。

## 3. 本次执行范围

**实际执行了：**
- 改造 5 个组件：移除 mock import，改为 props 驱动，无数据时显示空态
- 改造 3 个父页面：将 `result` state 传递给对应子组件
- 验证 `npm run build` (2.15s) 和 `go build` 通过
- 验证 mock import 清零

## 4. 明确未做事项

| 未做事项 | 原因 |
|---|---|
| 未实现 Jobs facets | 非本阶段 |
| 未修改后端业务 | 禁止 |
| 未继续 LLM | 阶段 11 暂停 |
| 未 git commit | 禁止 |

## 5. 修改文件清单

| 文件 | 修改内容 |
|---|---|
| `web/src/app/components/RecommendationDetail.tsx` | 移除 mock import；新增 `data?: RecommendationPageViewModel` prop；无数据时显示空态 |
| `web/src/app/components/RecommendationSidebar.tsx` | 同上 |
| `web/src/app/components/RecommendationContext.tsx` | 移除 mock import；新增 `resumeData?: ResumeSummaryViewModel` prop；无数据时不显示简历卡片 |
| `web/src/app/components/MatchAnalysisDetail.tsx` | 移除 mock import；新增 `data?: MatchAnalysisPageViewModel` prop；无数据时显示空态 |
| `web/src/app/components/ResumeOptimizationDetail.tsx` | 移除 mock import；新增 `data?: ResumeOptimizationPageViewModel` prop；无数据时显示空态 |
| `web/src/app/pages/JobRecommendationPage.tsx` | 传递 `result` 给 `RecommendationSidebar` 和 `RecommendationDetail` |
| `web/src/app/pages/JobMatchAnalysisPage.tsx` | 传递 `result` 给 `MatchAnalysisDetail` |
| `web/src/app/pages/ResumeOptimizationPage.tsx` | 传递 `result` 给 `ResumeOptimizationDetail` |

## 6. 新增文件清单

无。

## 7. 新增或更新 docs 清单

| 文件 | 新增/更新 | 目的 |
|---|---|---|
| `docs/mcp/项目上下文.md` | 更新 | 10-P-2-5 状态，P0 全部修复标记 |
| `docs/mcp/10-P-2-5-清理前端mock-fallback工程边界.md` | 新增 | 本阶段 MCP 记录 |
| `docs/09-开发日志/2026-06-01-228-10-P-2-5-清理mock边界.md` | 新增 | 开发日志 |

## 8. 关键决策

| 决策 | 原因 | 影响 |
|---|---|---|
| 组件 props 化而非显式开关 | 父页面已有 `result` state，传递即可 | 零额外依赖 |
| mock 文件不删除 | 仍可用于类型参考和测试 | mocks/*.ts 保留但不再被组件引用 |
| 不实现 VITE_ENABLE_MOCK 开关 | 组件已改为 props 驱动，无需额外开关层 | API 层 fallback 也已移除 |

## 9. 验收证据

| 验收项 | 方法 | 结果 |
|---|---|---|
| git status 初始 | `git status --short` | 干净 |
| mock import 残留（修复前） | `grep -rn 'from.*mocks' web/src/components/` | 5 个组件 |
| mock import 残留（修复后） | 同上 | 零匹配 |
| npm run build | `cd web && npm run build` | 2.15s pass |
| go build | `go build ./cmd/server` | PASS |
| git status 最终 | `git status --short` | 8 文件（5 组件 + 3 页面） |

## 10. 当前风险

| 风险 | 等级 | 阻塞什么 |
|---|---|---|
| 组件现在依赖真实 API 数据，未登录时 API 返回 401 → 页面空态 | P2 | 需登录后才能看到业务数据 |

## 11. 不确定项

无。

## 12. 下次继续前必须读取

1. `docs/mcp/项目上下文.md`
2. `docs/说明书.md`
3. 本文件

## 13. 用户明确要求 / 禁止事项

| 要求/禁止 | 来源 |
|---|---|
| 阶段 11-13 暂停 | 用户全局 |
| 不实现 Jobs facets | 本阶段 prompt |
| 不继续 LLM | 项目背景 |
| 不允许 git commit | 本阶段 prompt |

## 14. CodeBuddy 重启恢复摘要

本阶段改造 5 个核心组件（RecommendationDetail/Sidebar/Context、MatchAnalysisDetail、ResumeOptimizationDetail），移除所有顶层 mock import，改为从 props 接收数据。3 个父页面传递 `result` state 给子组件。npm build 2.15s pass。

P0-3 已修复。4 个 P0 全部解决。mock 文件保留但不再被组件引用。
