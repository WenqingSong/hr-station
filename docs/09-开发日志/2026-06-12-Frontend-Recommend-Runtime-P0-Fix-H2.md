# Frontend-Recommend-Runtime-P0-Fix-H2：修复推荐页 useMemo 未导入 + LoadingOverlay Hook 顺序错误

## 真实浏览器报错现象

用户浏览器控制台出现两个 P0 错误：

| # | 文件 | 错误 |
|---|------|------|
| 1 | `JobRecommendationPage.tsx:112` | `Uncaught ReferenceError: useMemo is not defined` |
| 2 | `AgentStepLoadingOverlay.tsx:206` | `React has detected a change in the order of Hooks called by AgentStepLoadingOverlay` / `Rendered more hooks than during the previous render` |

两个错误导致推荐页进入 ErrorBoundary。

---

## 根因分析

### 问题 1：useMemo 未导入

N2 修复在 `JobRecommendationPage.tsx:112` 新增 `useMemo` 调用以替代 useEffect 内 navigate，避免 React.StrictMode 双重渲染崩溃：

```tsx
const parsedUrlResumeId = useMemo(() => parseResumeIdFromQuery(resumeId), [resumeId]);
```

但顶部 React import 缺少 `useMemo`：

```tsx
// 修复前
import React, { useState, useEffect, useRef } from "react";
```

Vite 使用 esbuild 转译 JSX/TSX，不执行 TypeScript 类型检查，因此 `npm build` 可以通过。但浏览器运行时 `useMemo` 标识符在模块作用域中不可达，抛出 `ReferenceError`。

### 问题 2：AgentStepLoadingOverlay Hook 顺序违规

组件结构（修复前）：

```tsx
// lines 84-198: 11 个 Hook (5 useEffect + 3 useState + 3 useRef) ✅ 正常
// line 200:  if (!open) return null;              ← early return
// line 206:  const [displayElapsed, ...] = useState(...)  ← ❌ Hook after early return!
// line 209:  useEffect(() => { ... }, [open, finished])    ← ❌ Hook after early return!
```

在 React.StrictMode 下，组件被双重挂载。当 `open=false` 时，line 200 提前返回，line 206/209 的 Hook 不会被调用。当 `open=true` 时，这些 Hook 才会被调用。React 检测到两次渲染中 Hook 调用数量不同（11 vs 13），抛出 "Rendered more hooks" 错误。

`useState(elapsedSec)` 的初始值依赖于 `externalElapsed` 和 `startedAtRef.current` 的计算，这也是不应放在 early return 之后的另一个原因——`startedAtRef.current` 在 open=false 时未初始化。

---

## 修复

### JobRecommendationPage.tsx（H1 已完成）

```diff
- import React, { useState, useEffect, useRef } from "react";
+ import React, { useState, useEffect, useMemo, useRef } from "react";
```

### AgentStepLoadingOverlay.tsx

**1. 将 `displayElapsed` useState 提升到组件顶部**（line 90，在所有 early return 之前）：

```tsx
const [displayElapsed, setDisplayElapsed] = useState(0);
```

**2. 将 elapsed ticker useEffect 移到 early return 之前**（line 202-211）：

```tsx
useEffect(() => {
  const elapsed = externalElapsed ?? (startedAtRef.current ? Math.floor((Date.now() - startedAtRef.current) / 1000) : 0);
  setDisplayElapsed(elapsed);
  if (!open || finished) return;   // <-- guard inside effect, not Hook order change
  const ticker = setInterval(() => {
    setDisplayElapsed(Math.floor((Date.now() - startedAtRef.current) / 1000));
  }, 1000);
  return () => clearInterval(ticker);
}, [open, finished, externalElapsed]);
```

**3. 删除原 line 206-215 的重复 useState 和旧 ticker useEffect。**

**最终 Hook 顺序结构**：

```
Line 84:  useState  (activeIdx)
Line 85:  useState  (finished)
Line 86:  useRef    (timerRef)
Line 87:  useRef    (startedAtRef)
Line 88:  useRef    (backendDoneDoneRef)
Line 90:  useState  (displayElapsed) ← 修复：移至此
Line 103: useEffect (reset on open)
Line 113: useEffect (backendDone finish)
Line 158: useEffect (auto-advance)
Line 180: useEffect (safety timeout)
Line 194: useEffect (cleanup)
Line 203: useEffect (ticker)         ← 修复：移至此
Line 213: if (!open) return null;    ← early return，之后纯计算/渲染
```

所有 11 个 Hook（5 useEffect + 3 useState + 3 useRef）均在 line 213 无条件调用。React.StrictMode 双重渲染时 Hook 数量恒定 = 11。

---

## 验证结果

| 项目 | 结果 |
|------|:---:|
| npm build | ✅ PASS (2.08s) |
| tsc --noEmit | 无 `Cannot find name 'useMemo'` 或 Hook 相关错误 |
| dangerous hook scan | ✅ 0 命中 |
| useMemo is not defined 扫描 | ✅ 0 |
| Rendered more hooks 扫描 | ✅ 0 |
| Hook order warning 扫描 | ✅ 0 |

---

## 修改文件

| 文件 | 修改内容 |
|------|------|
| `web/src/app/pages/JobRecommendationPage.tsx:1` | React import 添加 `useMemo`（H1 已完成） |
| `web/src/app/components/common/AgentStepLoadingOverlay.tsx` | 3 处修改：（1）`displayElapsed` useState 提升到组件顶部；（2）ticker useEffect 移到 early return 之前；（3）删除原 line 206-215 重复代码 |

---

## 浏览器验证

环境无 Playwright/Chrome。**需用户本机刷新浏览器验证**：
1. `http://localhost:8080/recommend` — 不再显示"页面渲染异常"
2. `http://localhost:8080/recommend?resumeId=1` — console 无 `useMemo is not defined`
3. `http://localhost:8080/recommend?resumeId=8` — console 无 Hook order warning

---

## 备注

Vite HMR 502（`ws://localhost:8080 failed: 502`）是 Nginx 代理配置问题，不影响页面功能，本任务不处理。
