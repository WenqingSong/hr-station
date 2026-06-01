# Auth API Client 与类型补齐

> 阶段：8-1 | 状态：pass | 日期：2026-06-01

---

## 一、本阶段目标

补齐前端 Auth API Client 层与 TypeScript 类型定义，为后续登录页和 token 注入提供稳定基础。

---

## 二、后端 Auth 契约来源

| 来源 | 文件 |
|------|------|
| 请求类型 | `server/internal/service/auth_service.go` (RegisterRequest, LoginRequest) |
| 响应类型 | `server/internal/service/auth_service.go` (AuthResponse, AuthUser) |
| 路由 | `server/internal/http/router.go` |
| 8-0 核验 | `docs/11-登录态与用户态/登录态Token用户态现状核验.md` |

---

## 三、接口契约矩阵

| 接口 | 方法 | 请求体 | data 响应 | 是否需要 token |
|------|:--:|------|------|:--:|
| /api/auth/register | POST | `{email, password, name}` | `{token, user:{id,email,name}}` | ❌ |
| /api/auth/login | POST | `{email, password}` | `{token, user:{id,email,name}}` | ❌ |
| /api/auth/me | GET | — | `{id, email, name}` | ✅ Bearer |

---

## 四、新增 Auth 类型

**文件**：`web/src/types/auth.ts`

| 类型 | 字段 | 对齐后端 |
|------|------|:--:|
| `AuthUser` | `id: number, email: string, name: string` | `service.AuthUser` |
| `RegisterRequest` | `email, password, name` | `service.RegisterRequest` |
| `LoginRequest` | `email, password` | `service.LoginRequest` |
| `AuthTokenResponse` | `token: string, user: AuthUser` | `service.AuthResponse` |

### 关键设计决策

```
- AuthUser.id 使用 number（后端 uint64，前端 JS safe）
- me 接口 data 直接返回 AuthUser（不包裹 user 字段）
- 无 localStorage/sessionStorage 相关字段
- 无 mock/demo/fake 字段
```

---

## 五、新增 Auth API Client

**文件**：`web/src/api/auth.ts`

| 函数 | HTTP | 路径 | 请求体 | 返回类型 | fallback |
|------|:--:|------|------|------|:--:|
| `register()` | POST | `/auth/register` | `RegisterRequest` | `AuthTokenResponse` | ❌ 无 |
| `login()` | POST | `/auth/login` | `LoginRequest` | `AuthTokenResponse` | ❌ 无 |
| `getMe()` | GET | `/auth/me` | — (token 参数) | `AuthUser` | ❌ 无 |

### 边界约束

```
✅ Auth API 不使用 withFallback
✅ Auth API 不 import mocks
✅ Auth API 不返回假 token
✅ Auth API 不写 localStorage/sessionStorage
✅ getMe 手动传 Authorization (8-3 改为全局注入)
```

---

## 六、为什么 Auth 不使用 fallback

```
Auth 与业务数据 fallback 本质不同：
- 业务数据 fallback → 展示降级数据，体验略有下降
- Auth fallback → 制造假登录态，严重误导用户以为自己已登录

所以 auth.ts 不使用 withFallback，错误向上传播，
由 8-2 页面和 8-3 全局逻辑处理错误展示。
```

---

## 七、为什么本阶段不做 token storage

```
token storage 属于 8-3 范围：
- 8-1 只提供 auth 函数和类型
- 8-2 页面调用 auth 函数获取 token
- 8-3 才实现 token 存储、恢复、自动注入、401 处理
```

---

## 八、为什么本阶段不做登录页

```
登录/注册页面属于 8-2 范围：
- 8-1 补齐 API 基础层
- 8-2 创建 LoginPage / RegisterPage 组件，接入路由
- 8-3 全局 token 注入和 401 处理
```

---

## 九、修改文件清单

| 文件 | 操作 | 说明 |
|------|:--:|------|
| `web/src/types/auth.ts` | 新增 | 4 个 Auth 类型 |
| `web/src/api/auth.ts` | 新增 | 3 个 Auth API 函数 |
| `web/src/types/index.ts` | 修改 | +1 行 (auth export) |
| `web/src/api/index.ts` | 修改 | +1 行 (auth export) |

---

## 十、build 结果

```
✓ built in 1.94s ✅
```

1680 modules transformed，零 warning，零 error。

---

## 十一、grep 验证

| 检查项 | 结果 |
|------|:--:|
| auth.ts 不使用 withFallback | ✅ |
| auth.ts 不 import mock | ✅ |
| auth.ts 不写 localStorage/sessionStorage | ✅ |
| getMe 含 Authorization header | ✅ (仅手动传参) |
| web/src 无新增 LoginPage/RegisterPage | ✅ |
| web/src 无新增 AuthContext/ProtectedRoute | ✅ |

---

## 十二、P0/P1/P2

| 等级 | 数量 | 说明 |
|:--:|:--:|------|
| P0 | 0 | — |
| P1 | 0 | Auth API 补齐后 P1-A1 已解除 |
| P2 | 0 | — |

8-0 识别的 P1 问题状态变化：
- P1-A1 (auth API 缺失) → ✅ 已解除
- P1-A2 (登录页缺失) → ✅ 8-2 已解除
- P1-A3 (token 不注入) → 留待 8-3
- P1-A4 (401 被 fallback 吞掉) → 留待 8-3
- P1-A5 (路由鉴权缺失) → 留待 8-2/8-3

---

## 十三、越界检查

```
server/ ❌
web/src/app/ ❌
Makefile ❌
.env/.env.example ❌
Prompt/Schema/SQL ❌

零越界 ✅
```

---

## 十四、本阶段未做事项

```
✅ 本阶段不实现登录页
✅ 本阶段不实现注册页
✅ 本阶段不保存 token（localStorage/sessionStorage）
✅ 本阶段不自动注入 Authorization
✅ 本阶段不处理 401
✅ 本阶段不做路由鉴权
✅ 本阶段不修改后端 Auth
✅ 本阶段不改数据库
```

---

## 十五、是否可以进入 8-2

**可以进入 8-2（登录 / 注册页面与路由接入）。**

前提：
- Auth API Client 3 函数已补齐 ✅
- Auth 类型 4 个已补齐 ✅
- build 通过 ✅
- 零越界 ✅
- backend Auth 完整可用 ✅

---

## 十六、最终结论

**pass ✅** — Auth API Client 与类型补齐完成，可进入 8-2。
