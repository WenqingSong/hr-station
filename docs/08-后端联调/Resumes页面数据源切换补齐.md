# Resumes 页面数据源切换补齐

> 阶段：6-4-R | 状态：已完成 | 日期：2026-06-01

## 1. 本阶段目标

补齐 6-4 遗留的 Resumes 页面（ResumeManagementPage / MyResumesPage / ResumeComparisonPage）数据源从硬编码切换到 api viewModels。

## 2. 6-4 partial_pass 原因

6-4 中 Jobs 页面（5 组件）和 SelectResumeModal 已完成切换，但 Resumes 三个页面仍使用本地硬编码数据，属于 P2 遗留。

## 3. 修改范围

| 文件 | 原数据源 | 新数据源 |
|------|------|------|
| `ResumeManagementPage.tsx` | JSX 硬编码版本信息 | `loadCurrentResumeSummaryViewModel()` |
| `MyResumesPage.tsx` | useState 硬编码初始值 | edit 模式：`loadCurrentResumeSummaryViewModel()` 预填 |
| `ResumeComparisonPage.tsx` | JSX 硬编码 V1/V2 对比 | `loadResumeComparisonViewModel("1","2")` |

## 4. Resume viewModel loader 补齐结果

复用现有 loader，无需新增：
- `loadCurrentResumeSummaryViewModel()` — 已有 ✅
- `loadResumeComparisonViewModel()` — 已有 ✅

## 5. ResumeManagementPage 切换结果

- `useEffect` 调用 `loadCurrentResumeSummaryViewModel()` 加载当前简历
- 版本标签（3 处）从固定硬编码改为动态 `loadedVersionLabel`
- 版本列表下拉、项目列表、技能列表等 UI 结构保留

## 6. MyResumesPage 切换结果

- edit 模式：`useEffect` 加载 current resume，回填 `resumeName`
- add 模式：保持空白默认值
- 保存按钮保持演示态（"已保存本次演示修改"）

## 7. ResumeComparisonPage 切换结果

- `useEffect` 调用 `loadResumeComparisonViewModel("1","2")` 加载对比数据
- baseName / targetName（7 处）从硬编码改为动态 `comparison?.xxx`
- loading 态：显示"正在加载版本对比…"
- V1/V2 展示结构保留（关键词变化表、项目变化等）

## 8. MyResumesSidebar 切换结果

MyResumesSidebar 未被任何页面引用，无需处理。

## 9. token / auth 边界

无 token → API 401 → `withFallback` → mock 数据 → 页面不崩

## 10. mock fallback 保留

✅ 所有 loader 通过 api 层 `withFallback` 保证。

## 11. build 结果

```
✓ built in 1.59s
```
**通过 ✅**

## 12. grep 验证

| 检查项 | 结果 |
|------|:--:|
| Resumes 3 页面 import mocks | 0 ✅ |
| pages/components 直接 fetch | 0 ✅ |
| pages/components import api/client | 0 ✅ |
| localStorage/sessionStorage | 0 ✅ |
| axios / /api/v1 | 0 ✅ |

## 13. P0 / P1 / P2

| 等级 | 数量 | 说明 |
|:--:|:--:|------|
| P0 | 0 | — |
| P1 | 0 | — |
| P2 | 0 | — |

## 14. 越界检查

| 检查项 | 结果 |
|------|:--:|
| 修改 server | ❌ 零 |
| 修改 Makefile/docker/env | ❌ 零 |
| 修改 Agent files | ❌ 零 |

## 15. 本阶段未做事项

| # | 未做项 | 说明 |
|:--:|------|------|
| 1 | 页面视觉重构 | 只切数据源 |
| 2 | 系统性状态补齐 | 6-5 |
| 3 | 修改后端 | 禁止 |

## 16. 是否可以进入 6-5

**可以。** Resumes 三个页面已切到 api viewModels，6-4 P2 已清零。

## 17. 最终结论

**pass ✅**
