# 2026-06-12 Recommendation-Error-System-And-Deep-Crash-Fix-N2 — 修复推荐页 React 崩溃 + 报错体系

## 任务性质
前端 crash 修复 + 报错体系完善。

## 根因
`useEffect` 内调用 `navigate`（line 138/143）在 React.StrictMode 下导致组件 double-mount 时 hook 注册顺序不一致，触发 "Rendered more hooks" crash。

## 修复

### 1. Hook 稳定性
- 移除 `useEffect` 中的 `navigate`（原有 2 处 → 仅保留 auth redirect 1 处）
- 改用 `useMemo` 解析 URL `parsedUrlResumeId`
- useEffect 去重：`parsedUrlResumeId` 替换 `resumeId`
- StrictMode 保留 ✅

### 2. API 错误日志
- Fast catch: `console.groupCollapsed("[RecommendPage] Fast API failed")` + error + context
- Deep catch: `console.groupCollapsed("[RecommendPage] Deep API failed")` + classified error + context

### 3. ErrorBoundary 升级
- 用户标题："页面加载异常" → "页面渲染异常"
- 用户文案：不再误导为"结果生成失败"
- console: `console.groupCollapsed` + error + componentStack

### 验证
- npm build ✅ (2.62s), go build ✅
- resume_id=1 Fast: HTTP 200, 11893 bytes
- resume_id=8 Fast: HTTP 200, 12816 bytes
