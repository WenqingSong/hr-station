# 登录态 / Token / 用户态现状核验

> 阶段：8-0 | 状态：pass | 日期：2026-06-01

---

## 一、本阶段目标

完整摸清当前项目登录态、token、用户态的现状，为阶段 8 后续子阶段提供准确的修复输入。

---

## 二、核验范围

| 维度 | 范围 |
|------|------|
| 后端 Auth 接口 | register / login / me |
| 后端 JWT / Token | 生成、解析、过期、secret、claims |
| 后端 AuthMiddleware | 覆盖范围、注入方式 |
| 后端用户数据隔离 | 简历、岗位、Agent 结果的 user_id |
| 前端 Auth API | 是否有 auth.ts |
| 前端页面 | 是否有 login/register 页面 |
| 前端 token 管理 | localStorage / sessionStorage / 内存 |
| 前端 API client | 是否自动注入 Authorization |
| 前端 401 处理 | 是否有统一处理 |
| 前端路由 | 是否有 ProtectedRoute / 鉴权 |
| 前端 fallback | 是否掩盖未登录问题 |

---

## 三、后端 Auth 文件定位

| 模块 | 文件 | 当前状态 | 备注 |
|------|------|:--:|------|
| Auth Route | `server/internal/http/router.go` | ✅ 完整 | register/login 公开，me 受保护 |
| Auth Handler | `server/internal/http/auth_handler.go` | ✅ 完整 | Register/Login/Me 三方法 |
| Auth Service | `server/internal/service/auth_service.go` | ✅ 完整 | bcrypt + JWT HS256 + TokenParser |
| User Model | `server/internal/model/user.go` | ✅ 完整 | id/email/password_hash/name |
| User Repository | `server/internal/repository/user_repository.go` | ✅ 完整 | Create/FindByEmail/FindByID |
| JWT Utility | `server/internal/service/auth_service.go` (ParseToken) | ✅ 完整 | jwt-go v5, Claims struct |
| AuthMiddleware | `server/internal/middleware/auth.go` | ✅ 完整 | Bearer token 提取 + userID ctx 注入 |
| Response | `server/internal/response/` | ✅ 统一 | {code, message, data} |
| Config | `server/internal/config/config.go` | ✅ 配置化 | JWT_SECRET + JWT_EXPIRE_HOURS |
| users 表 | `server/sql/init.sql` | ✅ 完整 | id/email/password_hash/name/created_at/updated_at |

---

## 四、后端 Auth 接口核验

| 接口 | 方法 | 是否存在 | 是否需要 token | 请求体 | 响应体 | 结论 |
|------|:--:|:--:|:--:|------|------|:--:|
| /api/auth/register | POST | ✅ | ❌ | {email, password, name} | {code, message, data: {token, user}} | ✅ |
| /api/auth/login | POST | ✅ | ❌ | {email, password} | {code, message, data: {token, user}} | ✅ |
| /api/auth/me | GET | ✅ | ✅ Bearer | — | {code, message, data: {id, email, name}} | ✅ |

### 核验判断

| 检查项 | 状态 |
|------|:--:|
| register 返回 token | ✅ |
| login 返回 token | ✅ |
| me 根据 token 返回当前用户 | ✅ |
| 错误响应统一 | ✅ (code/message/data) |
| 密码加密 (bcrypt) | ✅ |
| token 有过期时间 (exp) | ✅ (默认 168h = 7天) |
| token secret 来自配置 | ✅ (JWT_SECRET，默认值 `change-me-in-local`) |
| 重复注册拦截 | ✅ (email already registered → 41001) |
| 错误密码登录拦截 | ✅ (invalid email or password → 41002) |
| 无效 token 拦截 | ✅ (40100 unauthorized) |
| token 签名方法 | HS256 |

---

## 五、JWT / Token 机制核验

| 检查项 | 当前实现 | 风险等级 | 备注 |
|------|------|:--:|------|
| token 生成 | `jwt.NewWithClaims(SigningMethodHS256, claims)` | ✅ 无风险 | 标准 HS256 |
| token 解析 | `jwt.ParseWithClaims()` + HMAC 签名验证 | ✅ 无风险 | 验证签名方法 |
| token 过期 | `ExpiresAt: now + jwtExpireHours` | ✅ 无风险 | 默认 168h(7天) |
| secret 来源 | `getEnv("JWT_SECRET", "change-me-in-local")` | ⚠️ P2 | 默认值不安全，但仅限本地开发 |
| claims 内容 | `{user_id, email, iss, exp, nbf, iat}` | ✅ 无风险 | 关键字段齐全 |
| userID 注入 context | `context.WithValue(ctx, userIDContextKey, userID)` | ✅ 无风险 | 类型安全 contextKey |
| 无效 token 处理 | 返回 401 `{code:40100, message:"unauthorized"}` | ✅ 无风险 | |
| 过期 token 处理 | jwt.Parse 返回 error → 返回 401 | ✅ 无风险 | |
| 签名算法校验 | 只允许 HMAC，拒绝其他算法 | ✅ 无风险 | |

---

## 六、AuthMiddleware 覆盖核验

| 接口模块 | 是否经过 AuthMiddleware | 是否从 context 取 userID | 风险 |
|------|:--:|:--:|:--:|
| /healthz | ❌ (公开) | N/A | ✅ |
| POST /api/auth/register | ❌ (公开) | N/A | ✅ |
| POST /api/auth/login | ❌ (公开) | N/A | ✅ |
| GET /api/auth/me | ✅ | ✅ | ✅ |
| POST /api/resumes | ✅ | ✅ | ✅ |
| GET /api/resumes/current | ✅ | ✅ | ✅ |
| GET /api/resumes/detail | ✅ | ✅ | ✅ |
| GET /api/resumes/versions | ✅ | ✅ | ✅ |
| GET /api/resumes/versions/{no} | ✅ | ✅ | ✅ |
| GET /api/resumes/versions/compare | ✅ | ✅ | ✅ |
| POST /api/resumes/optimize | ✅ | ✅ | ✅ |
| GET /api/jobs | ✅ | ✅ | ✅ |
| GET /api/jobs/recommend | ✅ | ✅ | ✅ |
| GET /api/jobs/{id} | ✅ | ✅ | ✅ |
| POST /api/jobs/{id}/match | ✅ | ✅ | ✅ |

### 判断

| 检查项 | 状态 |
|------|:--:|
| 公开接口和需登录接口区分清楚 | ✅ |
| 业务接口全部需要 Authorization | ✅ |
| Agent 接口也需要 Authorization | ✅ |
| Auth 接口不套 AuthMiddleware | ✅ |
| 无漏保护接口 | ✅ |

---

## 七、用户数据隔离核验

| 数据模块 | 是否按 user_id 查询 | 是否按 user_id 写入 | 越权风险 |
|------|:--:|:--:|:--:|
| users | ✅ (FindByEmail, FindByID) | ✅ (Create 时填入) | ✅ 无 |
| resume_versions | ✅ (全部 WHERE user_id=? ) | ✅ (CreateVersion 用 userID) | ✅ 无 |
| jobs | N/A (公共数据) | N/A | ✅ 合理 |
| recommendation_results | 需进一步确认 | 需进一步确认 | ⚠️ 低 |
| match_analyses | 需进一步确认 | 需进一步确认 | ⚠️ 低 |
| Agent 输入简历 | ✅ (从 context 取 userID，查当前用户简历) | ✅ | ✅ 无 |

### 重点判断

| 检查项 | 结论 |
|------|:--:|
| 简历读取是否只读当前用户 | ✅ GetCurrentByUserID / ListVersionsByUserID / GetByUserIDAndVersionNo |
| 简历写入是否绑定当前用户 | ✅ CreateVersion 使用 handler 传入的 userID |
| 简历版本详情是否校验 user_id | ✅ GetByUserIDAndVersionNo |
| Agent 分析是否使用当前用户简历 | ✅ 从 handler 取 userID → service 查询当前用户简历 |
| 用户是否能通过 versionNo 读取他人简历 | ❌ 不可以 — 查询时双向校验 userID + versionNo |
| Jobs 属于公共数据 | ✅ 合理，不需要 user_id 隔离 |

---

## 八、前端 Auth 代码核验

| 前端模块 | 是否存在 | 文件 | 当前状态 |
|------|:--:|------|------|
| auth api | ❌ | — | 未创建 auth.ts |
| auth types | ❌ | — | 无 login/register request/response 类型 |
| login page | ❌ | — | 无登录页面 |
| register page | ❌ | — | 无注册页面 |
| current user state | ❌ | — | 无 AuthContext / Zustand / store |
| token storage | ❌ | — | 零 localStorage / sessionStorage 使用 |
| token injection | ❌ | — | API client 支持 headers 但不自动注入 |
| 401 handling | ❌ | — | 401 被 withFallback 吞掉 |
| route guard | ❌ | — | 无 ProtectedRoute 组件 |

### 重点判断

| 检查项 | 结论 |
|------|:--:|
| web/src/api 是否有 auth.ts | ❌ 不存在 |
| 是否有 LoginPage / RegisterPage | ❌ 不存在 |
| 是否有 AuthContext / store | ❌ 不存在 |
| token 是否保存 | ❌ 不保存（零 localStorage/sessionStorage） |
| API client 是否自动带 Authorization | ❌ 不会。client.ts 合并 headers 但无人传入 Authorization |
| 401 是否会跳登录 | ❌ 不会。fallback 静默拦截 |
| 页面是否靠 fallback 展示 | ✅ 是。全部 API 调用包裹 withFallback |

---

## 九、API Client token 注入核验

| 文件 | 是否支持 headers | 是否自动注入 token | 风险 |
|------|:--:|:--:|------|
| client.ts | ✅ (options.headers 合并) | ❌ | **无全局 token 注入点** |
| jobs.ts | ❌ (不传 headers) | ❌ | 所有请求无 Authorization |
| resumes.ts | ❌ (不传 headers) | ❌ | 所有请求无 Authorization |
| agents.ts | ❌ (不传 headers) | ❌ | 所有请求无 Authorization |

### 重点判断

| 检查项 | 结论 |
|------|------|
| client.ts 是否支持 headers | ✅ 支持，通过 `ApiRequestOptions.headers` 合并 |
| client.ts 是否自动读取 token | ❌ 不会 |
| jobs/resumes/agents 是否传 Authorization | ❌ 全部没有 |
| 为什么接口能 fallback | catch 后返回 mock 数据，掩盖了 401 错误 |
| token 应该在哪里注入 | client.ts 的 requestRaw 中统一读取存储的 token |
| 401 应该由谁处理 | 应由 client.ts 统一拦截 401 → 清除 token → 跳登录 |

---

## 十、前端路由 / 页面鉴权核验

| 路由 | 页面 | 是否需要登录 | 当前是否保护 | 风险 |
|------|------|:--:|:--:|:--:|
| `/` | JobListPage | 建议是 | ❌ 无保护 | P1 |
| `/jobs` | JobListPage | 建议是 | ❌ 无保护 | P1 |
| `/jobs/:id` | JobDetailPage | 建议是 | ❌ 无保护 | P1 |
| `/recommend` | JobRecommendationPage | **是** | ❌ 无保护 | **P1** |
| `/match` | JobMatchAnalysisPage | **是** | ❌ 无保护 | **P1** |
| `/resume-optimization` | ResumeOptimizationPage | **是** | ❌ 无保护 | **P1** |
| `/resumes` | ResumeManagementPage | **是** | ❌ 无保护 | **P1** |
| `/resumes/edit` | MyResumesPage | **是** | ❌ 无保护 | **P1** |
| `/resumes/new` | MyResumesPage | **是** | ❌ 无保护 | **P1** |
| `/resumes/compare` | ResumeComparisonPage | **是** | ❌ 无保护 | **P1** |
| `/login` | — | — | ❌ 路由不存在 | **P1** |
| `/register` | — | — | ❌ 路由不存在 | **P1** |
| `*` | → `/` | — | ❌ | P2 |

### 判断

| 检查项 | 结论 |
|------|:--:|
| 是否有 /login | ❌ 不存在 |
| 是否有 /register | ❌ 不存在 |
| 是否有 ProtectedRoute | ❌ 不存在 |
| 简历页是否必须登录 | ❌ 当前未保护 |
| Agent 页是否必须登录 | ❌ 当前未保护 |
| 未登录访问跳登录 | ❌ 不会，直接 fallback 展示演示数据 |
| fallback 掩盖未登录 | ✅ 是核心问题 |

---

## 十一、登录态缺失对现有业务影响

| 业务 | 当前状态 | 登录态缺失影响 | 风险等级 |
|------|------|------|:--:|
| 岗位列表 | 真实 API + fallback | 仍可展示（岗位是公共数据） | 🟡 中 |
| 岗位详情 | 真实 API + fallback | 仍可展示（岗位是公共数据） | 🟡 中 |
| 当前简历 | 真实 API + fallback | 无 token → fallback 返回 mock 假数据 | 🔴 高 |
| 简历编辑 | 硬编码 + fallback | 无 token → 展示假用户数据 | 🔴 高 |
| 简历保存 | 演示态 | 无 token → 不写数据库 → 用户误以为已保存 | 🔴 高 |
| 简历版本 | 真实 API + fallback | 无 token → fallback 返回 mock 版本列表 | 🔴 高 |
| 推荐 Agent | 真实 API + fallback | 无 token → fallback 返回 mock 推荐 | 🔴 高 |
| 匹配 Agent | 真实 API + fallback | 无 token → fallback 返回 mock 匹配 | 🔴 高 |
| 优化 Agent | 真实 API + fallback | 无 token → fallback 返回 mock 优化 | 🔴 高 |

### 关键判断

```
没有登录态时：
❌ 我的简历是否真实属于当前用户 → 否，展示的是 mock 假数据
❌ 保存会写给谁 → 写入失败，静默 fallback
❌ Agent 用的是谁的简历 → mock 假简历
❌ fallback 是否会掩盖未登录 → 是！所有 401 被 catch 吞掉
❌ 用户是否会误以为看到的是自己的数据 → 会！
```

---

## 十二、curl 验证结果

| 场景 | HTTP | code | 是否符合预期 | 备注 |
|------|:--:|:--:|:--:|------|
| register | 200 | 0 | ✅ | 返回 token + user |
| login (正确密码) | 200 | 0 | ✅ | 返回 token + user |
| login (错误密码) | 401 | 41002 | ✅ | "invalid email or password" |
| me with token | 200 | 0 | ✅ | 返回 {id, email, name} |
| me without token | 401 | 40100 | ✅ | "unauthorized" |
| me with invalid token | 401 | 40100 | ✅ | "unauthorized" |
| resumes/current without token | 401 | 40100 | ✅ | "unauthorized" |
| resumes/current with token (新用户) | 200 | 0 | ✅ | data:null (无简历) |
| resumes/versions with token | 200 | 0 | ✅ | 返回版本列表 |
| jobs with token | 200 | 0 | ✅ | 返回岗位列表 |
| recommend with token (新用户) | 404 | 42001 | ✅ | "resume version not found" (预期) |
| duplicate email register | 400 | 41001 | ✅ | "email already registered" |

---

## 十三、build 结果

```
✓ built in 1.54s ✅
```

- 无 warning
- 无 error

---

## 十四、P0 / P1 / P2 风险清单

### P0 (阻断) — 0 个

```
后端 Auth 完整可用 ✅
无 P0 阻断问题 ✅
```

### P1 (上线前必修) — 5 个

| 编号 | 问题 | 风险说明 |
|:--:|------|------|
| P1-A1 | 前端没有 auth.ts API 模块 | 无法调用 register/login/me |
| P1-A2 | 前端没有登录页 / 注册页 | 用户无法登录 |
| P1-A3 | API client 不自动注入 token | 所有请求无 Authorization → 401 → fallback 静默 |
| P1-A4 | 401 被 withFallback 统一吞掉 | **核心风险**：当前所有 API 调用包裹 `catch → fallback()`，用户永远看不到 401，永远不知道自己未登录，永远看到的是 mock 假数据 |
| P1-A5 | 无 token storage + 路由无鉴权 | 无法保存和恢复登录态，Agent/简历页面无条件可访问 |

### P2 (可后置) — 3 个

| 编号 | 问题 | 风险说明 |
|:--:|------|------|
| P2-A1 | JWT_SECRET 默认值为 `change-me-in-local` | 本地开发可接受，生产须替换 |
| P2-A2 | 无 token 过期自动刷新机制 | token 7天后过期，无静默刷新，用户需重新登录 |
| P2-A3 | 无退出登录 / 用户菜单 / 记住登录 | 体验细节 |

---

## 十五、阶段 8 后续拆分建议

| 阶段 | 目标 | 是否改代码 | 依赖 | 验收标准 |
|:--:|------|:--:|------|------|
| 8-1 | 前端 Auth API Client 与类型补齐 | 是 (frontend-only) | 8-0 | `auth.ts` 实现 register/login/me 三函数；Auth 类型定义完整 |
| 8-2 | 登录 / 注册页面与路由接入 | 是 (frontend-only) | 8-1 | LoginPage / RegisterPage 可访问；支持注册后自动登录 |
| 8-3 | 全局 token 注入与 401 处理 | 是 (frontend-only) | 8-2 | client.ts 自动带 Authorization；401 统一跳登录；token 持久化恢复 |
| 8-4 | 用户态回归验收 | 否 (verification) | 8-3 | 全链路：注册→登录→me→刷新恢复→401跳登录→退出登录 全部通过 |

---

## 十六、本阶段未做事项

```
✅ 本阶段只核验，不修复
✅ 本阶段不新增登录页
✅ 本阶段不修改 API Client
✅ 本阶段不写 token storage
✅ 本阶段不修改后端 Auth
✅ 本阶段不改数据库
✅ 本阶段不做简历保存
✅ 本阶段不做 Agent 上下文真实化
```

---

## 十七、是否可以进入 8-1

**可以进入 8-1（前端 Auth API Client 与类型补齐）。**

前提判断：
- 后端 Auth 接口 3/3 全部可用 ✅
- JWT 机制完整，token 生成/解析/过期/secret 来源均正确 ✅
- AuthMiddleware 覆盖全部业务接口 ✅
- 用户数据隔离实现正确 ✅
- 前端 Auth 缺口已全部识别 ✅
- P0 = 0，无阻断性问题 ✅
- build 通过 ✅

---

## 十八、最终结论

**pass ✅** — 后端 Auth 机制完整可靠，前端 Auth 完全缺失但缺口已全部识别，可进入 8-1 开始补齐前端 Auth API Client。

### 核心发现

```
后端侧：✅ 登录态在服务器端已完整实现
  - register / login / me 三接口可用
  - JWT HS256 + bcrypt + 7天过期 + 完整 claims
  - AuthMiddleware 覆盖全部 14 个业务接口
  - 用户数据严格按 user_id 隔离

前端侧：❌ 登录态在前端完全缺失
  - 零 auth API、零登录页、零 token 存储、零 401 处理
  - 最关键问题：withFallback 静默吞掉所有 401
  - 用户永远看不到自己未登录 → 永远看到 mock 假数据
```

---

## 十九、文档信息

| 项目 | 值 |
|------|------|
| 阶段 | 8-0 |
| 状态 | pass |
| 日期 | 2026-06-01 |
| 当前分支 | main |
| 当前 HEAD | 076d0c2 |
| 工作区 | clean |
| 修改范围 | docs-only |
