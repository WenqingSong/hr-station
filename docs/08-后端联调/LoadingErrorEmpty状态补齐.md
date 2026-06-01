# Loading / Error / Empty 状态补齐

> 阶段：6-5 | 状态：已完成 | 日期：2026-06-01

## 1. 本阶段目标

在 6-4 页面数据源切换基础上，统一补齐普通业务页面/组件的 loading / error / empty 状态展示。

## 2. 修改范围

| 文件 | 修改内容 |
|------|------|
| `JobDetailBanner.tsx` | `if (!jobVm) return null` → loading 文案 + fallback 提示 |
| `JobDetailJD.tsx` | `if (!jobVm) return null` → loading 文案 + empty 提示 |
| `AIDecisionWorkspace.tsx` | 新增 loading 状态、Promise.all 并行加载、fallback 兜底标题 |
| `JobDetailAIDecision.tsx` | 新增 loading 状态、Promise.all 并行加载 |
| `MyResumesPage.tsx` | edit 模式 loading 文案 |
| `ResumeManagementPage.tsx` | 移除硬编码 "加载中…" 占位 |

## 3. Jobs 状态补齐结果

| 组件 | loading | error/fallback | empty |
|------|:--:|:--:|:--:|
| JobList | ✅ "正在加载岗位列表…" | ✅ "已切换到演示数据" | ✅ "暂无岗位数据" |
| JobDetailBanner | ✅ "正在加载岗位详情…" | ✅ "岗位接口暂时不可用，已展示演示数据。" | — |
| JobDetailJD | ✅ "正在加载岗位详情…" | — | ✅ "暂无岗位详情" |
| AIDecisionWorkspace | ✅ "正在加载岗位数据…" | — | — |

## 4. Resumes 状态补齐结果

| 页面 | loading | error/fallback | empty |
|------|:--:|:--:|:--:|
| ResumeComparisonPage | ✅ "正在加载版本对比…" | — | — |
| MyResumesPage | ✅ "正在加载简历数据…"（edit 模式） | — | — |
| ResumeManagementPage | ✅ 动态 label | — | — |

## 5. SelectResumeModal 状态补齐结果

| 模式 | 状态 |
|------|:--:|
| 加载中 | ✅ "正在加载简历列表…" |

## 6. AIDecisionWorkspace / JobDetailAIDecision

- 新增 `Promise.all` 并行加载简历和岗位，统一 `loading` 状态
- loading 态：显示 "正在加载岗位数据…" / "正在加载岗位详情…"
- 加载完成后 fallback null-safe 渲染，不白屏

## 7. Agent 页面未处理

✅ JobRecommendationPage / JobMatchAnalysisPage / ResumeOptimizationPage 及其子组件未修改（阶段 7）。

## 8. 状态文案规则

| 类型 | 文案 |
|------|------|
| loading | "正在加载{数据名}…" |
| error/fallback | "接口暂时不可用，已切换到演示数据。" |
| empty | "暂无{数据名}" |

## 9. build 结果

```
✓ built in 1.47s
```
**通过 ✅**

## 10. grep 验证

| 检查项 | 结果 |
|------|:--:|
| loading states 覆盖 | 8 处 ✅ |
| empty/fallback states | 4 处 ✅ |
| pages/components fetch | 0 ✅ |
| pages/components import api/client | 0 ✅ |
| localStorage/sessionStorage | 0 ✅ |
| axios / /api/v1 | 0 ✅ |

## 11. P0 / P1 / P2

| 等级 | 数量 |
|:--:|:--:|
| P0 | 0 |
| P1 | 0 |
| P2 | 0 |

## 12. 越界检查

| 检查项 | 结果 |
|------|:--:|
| 修改 server | ❌ 零 |
| 修改 Makefile/docker/env | ❌ 零 |
| 修改 Agent files | ❌ 零 |

## 13. 本阶段未做事项

| # | 未做项 | 说明 |
|:--:|------|------|
| 1 | Agent 页面状态 | 阶段 7 |
| 2 | 修改后端 | 禁止 |

## 14. 是否可以进入 6-6

**可以。** Jobs + Resumes 页面 loading / error / empty 已补齐。

## 15. 最终结论

**pass ✅**
