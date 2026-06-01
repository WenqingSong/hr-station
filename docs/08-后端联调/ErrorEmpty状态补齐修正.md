# Error / Empty 状态补齐修正

> 阶段：6-5-R | 状态：已完成 | 日期：2026-06-01

## 1. 本阶段目标

修正 6-5 遗留的 error / empty / fallback notice 缺失。

## 2. 6-5 partial_pass 原因

6-5 中 JobList/JobDetailBanner/JD/AIDecisionWorkspace/JobDetailAIDecision 已补 loading，但以下 error/empty/fallback 未补齐：
- ResumeManagementPage：无 error/empty
- MyResumesPage：无 fallback notice
- ResumeComparisonPage：无 empty/fallback（仅 loading）
- SelectResumeModal：loading/empty 混淆
- AIDecisionWorkspace/JobDetailAIDecision：无 fallback notice

## 3. 修改范围

| 文件 | 修改 |
|------|------|
| `ResumeManagementPage.tsx` | 新增 `resumeError` state + empty label |
| `MyResumesPage.tsx` | 新增 `resumeError` + fallback notice |
| `ResumeComparisonPage.tsx` | 新增 `error` state + empty/fallback 分支 |
| `SelectResumeModal.tsx` | 新增 `loading` state，区分 loading/empty |
| `AIDecisionWorkspace.tsx` | 新增 fallback notice |
| `JobDetailAIDecision.tsx` | 新增 fallback notice |

## 4–9. 修正结果

| 组件 | 新增状态 |
|------|------|
| ResumeManagementPage | error: "暂无简历数据" |
| MyResumesPage | fallback: "简历接口暂时不可用，已展示演示数据。" |
| ResumeComparisonPage | empty: "暂无可比较的简历版本" / fallback: "版本对比接口暂时不可用…" |
| SelectResumeModal | loading/empty 分离，empty: "暂无可选择的简历版本" |
| AIDecisionWorkspace | fallback: "简历接口暂时不可用，已展示演示数据。" |
| JobDetailAIDecision | fallback: "简历接口暂时不可用，已展示演示数据。" |

## 10. Agent 页面

未处理（阶段 7）。

## 11. build 结果

```
✓ built in 1.53s
```
**通过 ✅**

## 12. grep 验证

| 检查项 | 结果 |
|------|:--:|
| empty states | 3 处 |
| fallback notices | 6 处 |
| 工程化文案泄漏 | 0 |
| fetch / api-client / localStorage / axios / /api/v1 | 0 |

## 13. P0/P1/P2

P0=0, P1=0, P2=0

## 14. 越界检查

server:❌, Makefile:❌, docker:❌, nginx:❌, .env:❌

## 15–16. 结论

**pass ✅** — 6-5 遗留 error/empty 全部修正，可进入 6-6。
