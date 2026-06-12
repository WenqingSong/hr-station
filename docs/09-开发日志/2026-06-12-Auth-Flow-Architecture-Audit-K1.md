# Auth-Flow-Architecture-Audit-K1：登录注册全链路与弹窗组件只读审计

## 审计目标

完整审计登录注册全链路：前端架构、Dialog ref 问题、API client、后端认证安全、越权风险。

---

## 1. 前端认证架构

### 总评：架构合理，职责清晰，无严重设计缺陷

| 模块 | 当前职责 | 问题 | 建议 |
|------|------|------|------|
| AuthProvider (authStore.tsx) | Context+Reducer 全局状态；bootstrap 读 localStorage+调 /auth/me 验证；监听 401 事件 | 职责适中 | — |
| RequireAuth | 未登录→重定向 `/?auth=login&redirect=...`；登录态待确认→PageLoading | ✅ | — |
| Navbar | 根据 isAuthenticated 显示登录/用户菜单；检测 URL ?auth 参数打开弹窗 | 弹窗 open 状态在 Navbar 本地 useState | 可考虑提升到 Context |
| AuthDialog | 登录/注册表单；成功后调 setAuthenticated；有"评委一键体验"硬编码账号 | 使用 `DialogPrimitive.Overlay` 原生组件，无 ref 问题；表单校验完整 | — |
| API client (client.ts) | `buildHeaders()` 从 localStorage 读 token 注入 Authorization；401 时 clearAuthToken+dispatchEvent | ✅ | — |

### 初始化流程

```
App.tsx:
  AuthProvider
    ├── AuthBoot (调用 bootstrapAuth)
    │     ├── 读取 localStorage token
    │     ├── 有 token → GET /auth/me 验证
    │     └── 无 token/fail → isAuthReady=true
    └── RouterProvider (始终挂载)
          └── RequireAuth → isAuthReady ? 渲染子页面 : PageLoading
```

---

## 2. Dialog ref warning

### 根因

**`web/src/app/components/ui/dialog.tsx:33-47`** — `DialogOverlay` 和 `DialogContent` 是普通函数组件，未使用 `React.forwardRef`。

Radix UI 的 `DialogPrimitive` 内部通过 `SlotClone` 和 `asChild` 机制传递 ref，当 ref 到达普通函数组件时抛出 React 警告：

```
Warning: Function components cannot be given refs.
Check the render method of `SlotClone`.
at DialogOverlay
at DialogPortal
at Dialog
at AuthDialog
at Navbar
```

### 影响评估

| 项目 | 判断 |
|------|------|
| 风险等级 | **P2**（开发警告） |
| 是否影响登录弹窗功能 | ❌ 不影响（弹窗可正常打开/关闭/登录） |
| 是否影响 Radix focus trap | ❌ 不影响（Radix Primitive 自己管理） |
| 是否影响 animation | ❌ 不影响 |
| 是否影响评委演示观感 | ⚠️ console 有红字 warning，观感略差 |

### 最小修复

`DialogOverlay` 和 `DialogContent` 使用 `React.forwardRef`：

```tsx
const DialogOverlay = React.forwardRef<HTMLDivElement, React.ComponentProps<typeof DialogPrimitive.Overlay>>(
  ({ className, ...props }, ref) => (
    <DialogPrimitive.Overlay ref={ref} ... />
  )
);
```

---

## 3. 登录态初始化与路由保护

| 场景 | 当前表现 | 评价 |
|------|------|:---:|
| App 启动 | bootstrapAuth → localStorage token → /auth/me 验证 | ✅ |
| 未登录访问保护页 | RequireAuth 重定向 `/?auth=login&redirect=<安全路径>` | ✅ |
| 登录成功回跳 | AuthDialog onLoginSuccess → safeRedirect(redirect) | ✅ |
| token 过期 | client.ts 401 → clearAuthToken + dispatchEvent → authStore.clearAuth | ✅ |
| 退出登录 | POST /auth/logout + clearAuth + navigate("/") | ✅ |

---

## 4. API client 与 401 处理

| 项目 | 当前实现 | 评价 |
|------|------|:---:|
| Authorization 注入 | `buildHeaders()` 从 localStorage 读 token → Bearer header | ✅ |
| 401 处理 | `clearAuthToken()` + `dispatchEvent("offerr-auth-expired")` | ✅ |
| 错误结构 | `{ code, message, data }` 格式，`code===0` 成功 | ✅ |
| token 清理 | localStorage + authStore 双重清理 | ✅ |
| 敏感信息 | token 不在 URL、不在 console.error body | ✅ |

---

## 5. 登录注册表单体验

| 项目 | 当前实现 | 评价 |
|------|------|:---:|
| loading 状态 | useState isSubmitting → button disabled + "登录中…" | ✅ |
| 防重复提交 | isSubmitting 状态锁 | ✅ |
| 表单校验 | 手机号非空+格式；密码非空+长度≥6；注册确认密码一致性 | ✅ |
| 错误文案 | 中文，具体（如"手机号或密码错误"、"该手机号已注册"） | ✅ |
| 注册成功 | 直接登录（调用 setAuthenticated） | ✅ |
| 退出反馈 | 清晰（清除所有状态 + 跳转首页） | ✅ |
| "评委一键体验" | 硬编码账号 13900139000/123456 | ✅ 适合演示 |

---

## 6. 后端 auth 审计

| 项目 | 当前实现 | 评价 |
|------|------|:---:|
| 登录接口 | POST /api/auth/login，结构 JSON | ✅ |
| 注册接口 | POST /api/auth/register，手机号校验 | ✅ |
| 密码存储 | bcrypt (DefaultCost=10) → password_hash | ✅ 安全 |
| JWT | HS256，7 天过期（可配置），含 user_id+phone+session_id | ✅ |
| 会话管理 | JWT + Redis Session 双重验证，登出立即失效 | ✅ |
| 401 response | JSON `{ code: 40100, message: "unauthorized" }` | ✅ |
| 限流 | **无** | ❌ P2 |
| JWT Secret | 默认 `"change-me-in-local"` | ⚠️ 生产需改 |
| 资源归属校验 | **全部 SQL 带 `WHERE user_id=?`**，覆盖简历/推荐/匹配/优化/对比 | ✅ 全面 |

---

## 7. API 实测

| 场景 | HTTP | code | message | 结论 |
|------|:---:|:---:|------|:---:|
| login success | 200 | 0 | ok (token present) | ✅ |
| login wrong password | 401 | 41002 | "invalid credentials" | ✅ 结构化 |
| login empty body | 400 | 40001 | "invalid parameter" | ✅ 结构化 |
| me with token | 200 | 0 | phone=13900139000 | ✅ |
| me without token | 401 | 40100 | "unauthorized" | ✅ |

---

## 8. 越权风险

| 资源 | 是否校验 user_id | 风险 | 证据 |
|------|:---:|:---:|------|
| resumes | ✅ | 无 | `FindByUserIDAndID` → SQL `WHERE id=? AND user_id=?` |
| recommendation_results | ✅ | 无 | latest API 先 `FindByUserIDAndID` 验证简历归属 |
| match | ✅ | 无 | `FindByUserIDAndID` 验证 |
| optimization | ✅ | 无 | `FindByUserIDAndID` 验证 |
| version_compare | ✅ | 无 | `FindByUserIDAndID` 验证 |

**结论：无越权风险。** 所有受保护资源均通过 SQL 层 `WHERE user_id=?` 保证隔离。

---

## 9. 构建结果

| 项目 | 结果 |
|------|:---:|
| npm build | ✅ PASS (2.12s) |
| go build | ✅ PASS |
| go vet | ✅ PASS |

---

## 10. 修复方案建议

| 优先级 | 问题 | 修复位置 | 修复动作 |
|:----:|------|------|------|
| P2 | DialogOverlay/DialogContent 缺少 forwardRef | `ui/dialog.tsx:33,49` | 包裹 `React.forwardRef` |
| P2 | Navbar 本地 useState 管理弹窗 open | `Navbar.tsx` | 可提升到 AuthContext（非必须） |
| P2 | 无 API 限流 | 后端 middleware | 引入 rate limiter 中间件 |
| P2 | JWT Secret 默认弱值 | `.env` / `config.go` | 生产环境替换为强密钥 |

**无 P0/P1 问题。**

---

## 11. 下一步

- **推荐任务名**：`Auth-Flow-Professional-Fix-K2`
- **是否允许直接修复**：是（P2 级别，不影响核心功能）
- **是否需要后端修改**：可选（限流中间件）
- **是否需要前端修改**：是（Dialog forwardRef）
- **是否需要数据库结构修改**：否
- **是否可以跳过直接进入最终验收**：是（全链路架构合规，无阻塞问题）
