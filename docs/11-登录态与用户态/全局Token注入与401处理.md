# 全局 Token 注入与 401 处理

> 阶段：8-3 | 状态：pass | 日期：2026-06-01

---

## 一、本阶段目标

实现 token 持久化、API client 自动注入 Authorization、401 统一清 token，让前端从"能登录"升级到"登录态可以被 API client 使用"。

---

## 二、修改范围

| 文件 | 操作 | 说明 |
|------|:--:|------|
| `web/src/api/authStorage.ts` | 新增 | token storage（localStorage 读写） |
| `web/src/api/client.ts` | 修改 | 自动注入 Authorization + 401 清 token |
| `web/src/api/auth.ts` | 修改 | 更新注释（8-3 已完成） |
| `web/src/app/pages/LoginPage.tsx` | 修改 | 登录成功后保存 token |

---

## 三、authStorage 设计

**文件**：`web/src/api/authStorage.ts`

| 函数 | 功能 | 特点 |
|------|------|------|
| `getAuthToken()` | 从 localStorage 读取 token | SSR 安全，无效值返回 null |
| `setAuthToken(token)` | 写入 token 到 localStorage | SSR 安全，空值不写 |
| `clearAuthToken()` | 删除 token | 401 时调用 |

设计要点：
- key 名称：`offer_catcher_auth_token`
- 不存 user 大对象 ✅
- 不存 password ✅
- 不存 refresh token ✅
- 不使用 sessionStorage ✅
- 单 key 策略 ✅
- SSR/build 安全（检查 typeof window） ✅

---

## 四、client.ts Authorization 自动注入

**修改点**：

1. 新增 `buildHeaders()` 函数：
   - 默认 `Content-Type: application/json`
   - 调用方传入的 headers 优先（包括显式 Authorization）
   - 如果调用方未传 Authorization，从 `getAuthToken()` 注入

2. 401 处理：
   - HTTP 401 时调用 `clearAuthToken()`
   - 不跳转页面（无 `window.location.href`）
   - 不弹 toast
   - 抛出 `ApiClientError` 让调用方处理

### 覆盖策略

```
if (显式 headers 中有 Authorization) → 使用显式 headers 的值
else if (storage 中有 token) → 注入 Bearer <token>
else → 不注入 Authorization
```

### 影响范围

| API 调用 | 之前 | 之后 |
|------|------|------|
| jobs.ts 所有函数 | 无 Authorization → 401 → fallback | 自动带 Authorization → 200 → 真实数据 |
| resumes.ts 所有函数 | 同 | 自动带 Authorization |
| agents.ts 所有函数 | 同 | 自动带 Authorization |
| auth.ts register/login | 无 Authorization（正常） | 不变（无 token 时不注入） |
| auth.ts getMe(token) | 显式传 Authorization | 显式优先级更高，不变 |

---

## 五、401 处理策略

```
HTTP 401 时：
1. clearAuthToken() — 清除 localStorage 中的 token
2. 抛出 ApiClientError({status: 401}) — 让调用方感知
3. 不执行 window.location.href = "/login" — client 层不做路由跳转
4. 不弹 toast/alert
```

路由跳转留到 8-4 或后续阶段处理。

---

## 六、LoginPage 保存 token 结果

修改前：
```ts
await login(input);
navigate("/");
```

修改后：
```ts
const response = await login(input);
setAuthToken(response.token);
navigate("/");
```

登录成功后 token 已持久化，刷新页面后可恢复。

---

## 七、RegisterPage 策略

**不变**。注册成功后 navigate("/login")，不保存 token。
用户需要到登录页重新输入密码登录。

---

## 八、为什么本阶段不做 ProtectedRoute

```
ProtectedRoute 需要 wait for 用户态初始化（getMe 校验 token 有效性）。
这属于 8-4 用户态回归验收的范围。
8-3 先解决基础设施：token 存了、client 注入了、401 清 token。
```

---

## 九、为什么 client.ts 不直接跳转

```
client.ts 是 API 层，不应直接控制路由（解耦原则）。
跳转策略由页面或上层 Guard 决定。
```

---

## 十、curl 验证

复用 8-0 验证结果（12/12 pass）。

---

## 十一、URL 验证

前端路由 `/login` `/register` 在 8-2 已验证。

---

## 十二、build 结果

```
✓ built in 1.70s ✅
```

1683 modules transformed，零 warning，零 error。

---

## 十三、grep 验证

| 检查项 | 结果 |
|------|:--:|
| localStorage 只在 authStorage.ts（+auth.ts 注释） | ✅ |
| 无 sessionStorage（仅注释） | ✅ |
| 无 ProtectedRoute/AuthGuard/AuthContext | ✅ |
| client.ts/LoginPage 无 location.href/alert | ✅ |
| 页面无工程化错误文案泄漏 | ✅ |
| client.ts 含 getAuthToken/Authorization 注入/clearAuthToken | ✅ |
| LoginPage 含 setAuthToken | ✅ |

---

## 十四、P0/P1/P2

| 等级 | 数量 | 说明 |
|:--:|:--:|------|
| P0 | 0 | — |
| P1 | 5→1 | A1(8-1),A2(8-2),**A3(8-3)**,**A4(8-3)**,A5→8-4 |
| P2 | 3 | 不变 |

8-0 识别的 P1 已解除 4/5：
- P1-A1 (auth API) → ✅ 8-1
- P1-A2 (登录页) → ✅ 8-2
- P1-A3 (token 不存储) → ✅ 8-3
- P1-A4 (Authorization 不注入 + 401) → ✅ 8-3
- P1-A5 (路由鉴权) → 8-4

---

## 十五、越界检查

```
server/ ❌
Makefile ❌
.env/.env.example ❌
Prompt/Schema/SQL ❌

仅修改：authStorage(new) + client.ts + auth.ts + LoginPage + docs ✅
```

---

## 十六、本阶段未做事项

```
✅ 本阶段实现 token storage
✅ 本阶段实现 Authorization 自动注入
✅ 本阶段实现 401 清 token
✅ 本阶段不实现 ProtectedRoute
✅ 本阶段不实现 AuthContext
✅ 本阶段不实现完整 currentUser
✅ 本阶段不做用户菜单
✅ 本阶段不做退出登录 UI
✅ 本阶段不做简历真实保存
✅ 本阶段不做 Agent 上下文真实化
✅ 本阶段不修改后端
```

---

## 十七、是否可以进入 8-4

**可以进入 8-4（用户态回归验收）。**

前提：
- token storage 已实现 ✅
- Authorization 自动注入已完成 ✅
- 401 清 token 已完成 ✅
- LoginPage 登录后保存 token ✅
- build 通过 ✅
- 零越界 ✅

---

## 十八、最终结论

**pass ✅** — 全局 token 注入与 401 处理完成，可进入 8-4。
