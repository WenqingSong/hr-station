# Frontend-UseMemo-Import-Runtime-Fix-H1：修复推荐页 useMemo 未导入导致的运行时崩溃

## 现象

用户真实浏览器访问 `http://localhost:8080/recommend` 时，页面进入 ErrorBoundary，控制台报错：

```
JobRecommendationPage.tsx:112 Uncaught ReferenceError: useMemo is not defined
at JobRecommendationPage
```

## 根因

N2（React Hook 崩溃修复）在 `JobRecommendationPage.tsx:112` 新增了 `useMemo` 调用：

```tsx
const parsedUrlResumeId = useMemo(() => parseResumeIdFromQuery(resumeId), [resumeId]);
```

但顶部 React import 缺少 `useMemo`：

```tsx
// 修复前
import React, { useState, useEffect, useRef } from "react";
```

Vite 使用 esbuild 进行 JSX/TSX 转译，不执行 TypeScript 类型检查，因此 `npm build` 可以通过。但在浏览器运行时，React 需要在模块作用域中引用 `useMemo`，缺失 import 会导致 `ReferenceError`。

## 修复

**文件**：`web/src/app/pages/JobRecommendationPage.tsx`

**修改**：在 React import 中添加 `useMemo`

```diff
- import React, { useState, useEffect, useRef } from "react";
+ import React, { useState, useEffect, useMemo, useRef } from "react";
```

仅 1 行修改，不涉及业务逻辑。

## Hook import 验收

| Hook | 是否使用 | 是否 import |
|------|:---:|:---:|
| useState | ✅（14 处） | ✅ |
| useEffect | ✅（3 处） | ✅ |
| useMemo | ✅（1 处，line 112） | ✅ 已修复 |
| useRef | ✅（4 处） | ✅ |
| useCallback | ❌ 不使用 | N/A |
| useReducer | ❌ 不使用 | N/A |

## 验证结果

| 项目 | 结果 |
|------|:---:|
| npm build | ✅ PASS（2.23s） |
| tsc --noEmit | 35 个预存历史 TS 错误，**无** `Cannot find name 'useMemo'` ✅ |
| useMemo is not defined 扫描 | ✅ 0 命中 |
| Rendered more hooks 扫描 | ✅ 0 命中 |

## 浏览器验证

环境无 Playwright/Chrome。**需用户本机刷新验证**：
1. 打开 `http://localhost:8080/recommend`
2. 打开 `http://localhost:8080/recommend?resumeId=1`
3. 打开 `http://localhost:8080/recommend?resumeId=8`
4. 页面不再显示"页面渲染异常"
5. console 不再出现 `useMemo is not defined`

## 备注

Vite HMR WebSocket 502（`ws://localhost:8080 failed: 502`）是代理配置问题，不影响页面功能，本任务不处理。

## 修改文件

| 文件 | 修改内容 |
|------|------|
| `web/src/app/pages/JobRecommendationPage.tsx` | 第 1 行：React import 添加 `useMemo` |
