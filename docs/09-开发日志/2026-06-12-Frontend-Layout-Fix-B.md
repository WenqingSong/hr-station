# 2026-06-12 Frontend-Layout-Fix-B：四核心 AI 页面状态态布局修复

## 修复背景

`Frontend-Layout-Audit-A` 审计发现：推荐页、匹配页、优化页、对比页四个核心 AI 页面的 error / empty / guard / loading 状态大量使用裸 `max-w` + `py-12/py-16/py-32` 容器，未包裹 `ContentPageShell`，导致内容紧贴或部分被 fixed navbar（68px）遮挡。

## 修复原则

1. **统一容器**：所有非 Hero 页面状态态（error / empty / guard / loading）必须使用 `ContentPageShell`（提供 `pt-24` navbar 偏移）
2. **错误态**：统一使用 `PageError` 组件
3. **空状态**：统一使用 `PageEmpty` 组件
4. **Loading 态**：优先使用 `PageLoading`（自包含 `min-h-[calc(100vh-80px)]`），保持内联 loading 时包裹 `ContentPageShell`
5. **不改变业务逻辑、API 调用、后端代码**

## 修复内容

### JobRecommendationPage.tsx（推荐页）
- `urlResumeNotFound`：裸 div → `ContentPageShell + PageError`
- `no resumes`：裸 div → `ContentPageShell + PageEmpty`
- `!selectedOption`：`PageEmpty` 裸用 → `ContentPageShell + PageError`
- `fast_loading`：裸 div py-32 → `PageLoading`
- `fast_error / profile_failed`：裸 div py-12 → `ContentPageShell + PageError`（保留"重新尝试"+"返回我的简历"按钮）
- `no result`：裸 div py-12 → `ContentPageShell + PageEmpty`

### JobMatchAnalysisPage.tsx（匹配页）
- `!jobId`：裸 div py-16 → `ContentPageShell + PageError`
- `jobNotFound`：裸 div py-16 → `ContentPageShell + PageError`
- `urlResumeNotFound`：裸 div py-16 → `ContentPageShell + PageError`
- `no resumes`：裸 div py-16 → `ContentPageShell + PageEmpty`
- `!selectedOption`：裸 div py-12 → `ContentPageShell + PageError`
- `resumeDetailError`：裸 div py-12 → `ContentPageShell + PageError`
- `fast_error`：裸 div py-12 → `ContentPageShell + PageError`

### ResumeOptimizationPage.tsx（优化页）
- `!parsedResumeId`：裸 div py-12 → `ContentPageShell + PageError`
- `no resumes`：裸 div py-12 → `ContentPageShell + PageError`
- `resumeDetailError`：裸 div py-12 → `ContentPageShell + PageError`
- `fast_error`：裸 div py-12 → `ContentPageShell + PageError`

### ResumeComparisonPage.tsx（对比页）
- `fromId === toId`：裸 `PageError` → `ContentPageShell + PageError`
- `fast_loading`：裸 div py-32 → `PageLoading`
- `fast_error`：裸 div py-12 → `ContentPageShell + PageError`
- `no activeResult`：裸 div py-12 → `ContentPageShell + PageError`

## 布局规范

| 组件 | 用途 | navbar 偏移 | 独立使用？ |
|---|---|---|---|
| `ContentPageShell` | 页面外层 Shell | `pt-24` (96px) | 必须用于所有非 Hero 页面 |
| `PageError` | 错误态卡片 | 无（依赖 ContentPageShell） | 需包裹在 ContentPageShell 内 |
| `PageEmpty` | 空状态卡片 | 无（依赖 ContentPageShell） | 需包裹在 ContentPageShell 内 |
| `PageLoading` | 加载态（居中 spinner） | `min-h-[calc(100vh-80px)]` | ✅ 可独立使用 |

## 构建验证

| 项目 | 结果 |
|---|---|
| `npm build` (vite) | ✅ 通过 |
| `go build` | ✅ 通过 |
| `go vet` | ✅ 通过 |
| `go test` | ✅ 通过 |
| `make check` | ✅ 通过 |
| `make prod-check` | ❌ 缺少 .env.prod（环境配置问题） |

## 风险

- 无业务逻辑变更，无后端变更
- 完全向后兼容正常态布局
- 按钮和文案均保留
