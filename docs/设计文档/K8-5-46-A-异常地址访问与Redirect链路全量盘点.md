# K8-5-46-A：异常地址访问与 Redirect 链路全量盘点

> 状态：✅ 已完成 | 日期：2026-06-06 | 性质：只读审计 | 修改：0 行源码

---

## 一、异常 URL 现象

```
http://localhost:8080/?auth=login&redirect=%2Fresumes%2Fedit%3FversionNo%3D0
```

解码后：
```
/?auth=login&redirect=/resumes/edit?versionNo=0
```

页面停留在首页 `/jobs`，未显示登录弹窗或错误页。

---

## 二、当前路由表

| 路径 | 页面 | 登录保护 | 异常处理 |
|------|------|:---:|------|
| `/` | JobListPage | ❌ 公开 | — |
| `/jobs` | JobListPage | ❌ 公开 | — |
| `/jobs/:id` | JobDetailPage | ❌ 公开 | — |
| `/recommend` | JobRecommendationPage | ✅ RequireAuth | 未登录→`/?auth=login&redirect=...` |
| `/match` | JobMatchAnalysisPage | ✅ RequireAuth | 同上 |
| `/resume-optimization` | ResumeOptimizationPage | ✅ RequireAuth | 同上 |
| `/resumes` | ResumeManagementPage | ✅ RequireAuth | 同上 |
| `/resumes/edit` | MyResumesPage | ✅ RequireAuth | 同上 |
| `/resumes/new` | MyResumesPage | ✅ RequireAuth | 同上 |
| `/resumes/compare` | ResumeComparisonPage | ✅ RequireAuth | 同上 |
| `/login` | LoginPage | ❌ 公开 | redirect wrapper → `/?auth=login` |
| `/register` | RegisterPage | ❌ 公开 | redirect wrapper → `/?auth=login` |
| `*` (catch-all) | Navigate to `/` | ❌ 公开 | **静默跳转首页，无 404 页面** |

---

## 三、Auth / Redirect 链路详查

| 链路 | 当前行为 | 风险 |
|------|---------|------|
| 未登录访问受保护页面 | `RequireAuth` → `Navigate to /?auth=login&redirect=<原文编码>` | **redirect 参数零校验**——任何 URL 片段都原样编码 |
| redirect 参数生成 | `encodeURIComponent(location.pathname + location.search)` | 携带所有 query params，包括 versionNo=0 |
| redirect 参数验证 | **不存在** `getSafeRedirect` / `isSafeRedirect` | 无白名单、无路径校验、无参数合法性检查 |
| login 成功后跳转 | `AuthDialog` 读取 `redirectAfterLoginRef` → `navigate(redirect)` | 直接跳转，不验证目标 |
| 401 处理 | `client.ts` 触发 `offerr-auth-expired` 事件 | 简单 redirect 到 `/?auth=login` |
| localStorage 残留 token | `authStore.bootstrapAuth()` 通过 `/auth/me` 校验 | 无效 token 被清除，风险低 |

## 四、versionNo 参数处理

| 场景 | 当前行为 | 问题 |
|------|---------|------|
| `/resumes/edit` (无 versionNo) | 加载 current 简历 | ✅ 正常 |
| `/resumes/edit?versionNo=1` | 加载版本 1 → API → 渲染 | ✅ 正常 |
| `/resumes/edit?versionNo=0` | regex 通过，`Number("0")=0`，`>0` 失败 → 同无 versionNo → 加载 current | ⚠️ 不报错但不合理 |
| `/resumes/edit?versionNo=abc` | regex 失败 → `editingVersionNo=null` → 同无 versionNo | ⚠️ 静默降级 |
| `/resumes/edit?versionNo=9999` | API 404 → 显示"版本不存在" | ✅ 正常 |
| 未登录访问以上任意 | `RequireAuth` → `/?auth=login&redirect=...` | ❌ **redirect 不校验，versionNo=0 被编码传递** |
| 未登录且 versionNo=0 | redirect=`/resumes/edit?versionNo=0` → 登录后跳回 → 同无 versionNo | ❌ 冗余 roundtrip |

## 五、404 / SPA fallback 情况

| 项 | 结果 |
|----|:---:|
| catch-all 路由 | ✅ `path: "*"` → `<Navigate to="/" replace />` |
| 专用 404 页面 | ❌ **不存在**——静默跳转首页 |
| SPA fallback | Vite dev server 自动处理，Nginx proxy 到 Vite |
| 异常地址最终落点 | 首页 `/jobs`——用户看到岗位列表但 URL 仍带 `?auth=login&redirect=...` |

## 六、根因判断

```
根因：redirect 参数零校验，无安全白名单。

证据链：
1. RequireAuth → encodeURIComponent(location.pathname + location.search) → 不校验任何参数
2. getSafeRedirect / isSafeRedirect 在代码库中不存在
3. AuthDialog 登录成功后直接 navigate(redirectAfterLoginRef.current) 不验证
4. versionNo=0 虽被 MyResumesPage 兼容，但作为 redirect 参数仍被携带
5. catch-all 路由静默跳转首页，无 404 专页

具体异常 URL 链路：
/resumes/edit?versionNo=0
→ RequireAuth 未登录捕获
→ /?auth=login&redirect=%2Fresumes%2Fedit%3FversionNo%3D0
→ Navbar 读取 auth=login → 打开 AuthDialog
→ 登录成功 → navigate(/resumes/edit?versionNo=0)
→ MyResumesPage: versionNo=0 → 同 no versionNo → 加载 current
```

---

## 七、推荐实现方案

### 方案 B（推荐）：统一安全修复

| 维度 | 内容 |
|------|------|
| **新增文件** | `web/src/utils/redirect.ts` — `isSafeRedirect()` + `cleanRedirect()` |
| **修改文件** | `RequireAuth.tsx` — redirect 前校验 |
| | `Navbar.tsx` / `AuthDialog.tsx` — 登录后跳转前校验 |
| | `MyResumesPage.tsx` — versionNo≤0 直接 replace 到 `/resumes` |
| | `authRedirect.ts` — `getAuthRedirectIf401` 接入校验 |
| | `routes.tsx` — `*` 改为 `<NotFoundPage />` |
| **新增页面** | `NotFoundPage.tsx` |
| **安全规则** | redirect 必须以 `/` 开头，不含 `//`/`http:`/`https:`，路径在已知路由白名单内 |

### isSafeRedirect 核心逻辑

```ts
const SAFE_PATHS = [/^\/resumes/, /^\/recommend/, /^\/match/, /^\/jobs/, /^\/resume-/];
export function isSafeRedirect(url: string): boolean {
  if (!url.startsWith('/') || url.startsWith('//')) return false;
  if (/https?:\/\//i.test(url)) return false;
  return SAFE_PATHS.some(p => p.test(url));
}
```

### 方案 A（最小修复）

仅 `MyResumesPage` 中 `versionNo ≤ 0` 时 replace 到 `/resumes`。不改 redirect 校验。

### 方案 C（重构）

引入路由守卫和错误边界。改动量大，Demo 阶段不推荐。

---

## 八、下一阶段建议

```
K8-5-46-B：异常地址 Redirect 与非法参数统一修复
  — 新增 isSafeRedirect + 404 页面
  — RequireAuth / AuthDialog / authRedirect 接入安全校验
  — MyResumesPage 非法 versionNo replace 到 /resumes
```

---

## 九、构建结果

| 项 | 结果 |
|----|:---:|
| `npm run build` | ✅ |
| `go build ./...` | ✅ |

---

## 十、未修改范围确认

```
✅ 未修改源码（Go / TS / SQL）
✅ 未修改 seed / DB / API / .env
✅ 未执行 make reset-up
✅ 未执行 Git 命令
```
