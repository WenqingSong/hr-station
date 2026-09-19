# Refactor Card：IAM 系统实现（agents 服务）

## 0. 任务元数据

- 仓库：https://cnb.cool/go-cloud-devops/hr-station
- 任务名称：IAM 系统实现（agents 服务）
- 任务目录：working/refactor/iam-implementation
- 基准分支：main
- 基准提交：2a7f17d（HEAD，未经改动）
- 当前提交：2a7f17d
- 项目负责人：待指定
- 当前阶段：DRAFT
- 当前负责人：Analyst
- 契约版本：DRAFT
- 契约状态：WAITING_FOR_OWNER_APPROVAL
- 创建时间：2026-09-20
- 最后更新时间：2026-09-20

## 1. 任务目标与负责人决定

### 1.1 要解决的问题

在 `/workspace/agents` 已搭建的 GoFrame 骨架（当前仅含 `/healthz` 健康检查）之上，从零实现一套**规范、生产级**的 IAM（身份认证与访问管理）子系统，为后续其他业务模块（简历、岗位、匹配、推荐、优化、对比）提供统一的认证与授权能力。

旧系统（`/workspace/server`）中已有可用的 Auth 实现（register/login/me/logout + JWT + Redis 会话 + 单设备顶替），但其存在架构性缺陷（手写 `net/http`、巨型组合根、内联鉴权重复、错误码三套并存、测试几乎空白），且**本次重构不照搬旧实现，而是按 GoFrame 规范重新设计与实现**。旧实现仅作为对外行为契约的参照。

### 1.2 目标范围

在 `agents` 服务内实现以下 IAM 能力（本轮聚焦认证，不含 RBAC/权限点授权）：

1. 用户注册（`POST /api/auth/register`）
2. 用户登录（`POST /api/auth/login`）
3. 获取当前用户（`GET /api/auth/me`）
4. 用户登出（`POST /api/auth/logout`）
5. JWT 签发与校验
6. Redis 会话管理与单设备顶替语义
7. 认证中间件（统一鉴权，替换旧系统的内联鉴权重复）
8. 统一响应包装（`{code, message, data}`）与错误语义

### 1.3 禁止修改

- 不修改 `/workspace/server` 旧系统代码。
- 不修改 `agents` 现有 `main.go`、`cmd.go` 的启动/DB 预检骨架约定（可增量扩展，不得破坏现有 `/healthz` 行为）。
- 不实现 RBAC、角色、权限点、多租户、OAuth/短信等超出本轮范围的能力。
- 不迁移旧系统的四大模型功能（Match/Recommend/Optimize/Compare）到 agents。

### 1.4 项目负责人决定

- 选择方案：待定
- 决定理由：待定
- 批准的契约版本：未批准
- 批准状态：未批准
- 批准人：待定
- 批准时间：待定

## 2. 当前实现证据

> 说明：本轮为「从零实现 IAM」。`agents` 服务当前**尚无任何 IAM 代码**，因此「当前实现」证据聚焦于：(a) `agents` 骨架现状；(b) 旧系统 `/workspace/server` 中已存在的 Auth 实现作为行为参照与缺陷来源。

### 2.1 当前调用链

**agents 骨架现状调用链（仅健康检查）**：

```
main.go (加载 .env → setDefaultEnv → setupDatabase → cmd.Main.Run)
  └── internal/cmd/cmd.go
        ├── g.Cfg() 读取 manifest/config/config.yaml
        ├── checkDatabase() 启动前限时 Ping MySQL（失败即 Fatal）
        └── g.Server().Group("/")
              ├── MiddlewareHandlerResponse（统一响应包装）
              ├── Middleware DBHealth（探测 DB 状态写入 ctx）
              └── Bind(health.NewV1())
                    └── ControllerV1.Health → /healthz
```

**旧系统 Auth 调用链（行为参照，不作为实现模板）**：

```
HTTP (net/http + ServeMux, router.go)
  ├── POST /api/auth/register → AuthHandler.Register → AuthService.Register
  ├── POST /api/auth/login    → AuthHandler.Login    → AuthService.Login
  ├── GET  /api/auth/me       → AuthMiddleware → AuthHandler.Me → AuthService.Me
  └── POST /api/auth/logout   → AuthMiddleware → AuthHandler.Logout → AuthService.Logout
        AuthService 依赖 UserRepository(*sql.DB) + session.Manager(*redis.Client)
```

### 2.2 当前依赖关系

**agents 服务**（`/workspace/agents/go.mod`）：
- `github.com/gogf/gf/v2 v2.10.3` + `contrib/drivers/mysql/v2`
- `github.com/joho/godotenv v1.5.1`
- **缺失**：JWT 库、bcrypt 库、Redis 客户端库（当前 go.mod 均未引入，需新增）

**旧系统 Auth 依赖**（`/workspace/server`）：
- JWT：`github.com/golang-jwt/jwt/v5`
- 密码：`golang.org/x/crypto/bcrypt`
- 会话：`github.com/redis/go-redis/v9`
- DB：`database/sql` + `github.com/go-sql-driver/mysql`（原生 SQL，无 ORM）

### 2.3 当前测试情况

- `agents` 现有测试仅 2 个文件：
  - `internal/cmd/cmd_test.go`：进程内启动 HTTP 服务测 `/healthz`（DB 指向不可达地址，断言 database 状态为 ok/unavailable）。
  - `internal/cmd/config_test.go`：校验 `config.yaml` 静态配置可解析。
- `agents` **无任何 IAM 相关测试**。
- 旧系统 Auth 测试仅覆盖纯函数 `validatePhone` / `validatePassword`（`auth_service_test.go`），**缺**：Register/Login/Me/Logout、JWT 签发解析、bcrypt、会话顶替、中间件、Repository 的测试（`session/`、`middleware/`、`http/`、`repository/` 目录均无 `*_test.go`）。

### 2.4 代码证据

| ID | 结论 | 文件和位置 | 证据 | 状态 |
|---|---|---|---|---|
| E-001 | agents 服务无任何 IAM 代码，仅有健康检查 | `/workspace/agents/internal/controller/health/health.go`、`/workspace/agents/api/health/v1/health.go` | 仅 `Health` 控制器与 `HealthReq/HealthRes` | 已验证 |
| E-002 | agents 已引入 GoFrame MySQL 驱动，gdb 可用 | `/workspace/agents/go.mod:6-9` | 依赖 `gf/v2`、`contrib/drivers/mysql/v2` | 已验证 |
| E-003 | agents go.mod 缺 JWT/bcrypt/Redis 依赖 | `/workspace/agents/go.mod` | 无 `golang-jwt`、`x/crypto/bcrypt`、`go-redis` | 已验证 |
| E-004 | agents 数据库为独立库 `offer_catcher_re` | `/workspace/agents/sql/init.sql`、`/workspace/agents/main.go:40` | `CREATE DATABASE offer_catcher_re`，`AGENTS_DB_NAME` 默认 `offer_catcher_re` | 已验证 |
| E-005 | agents 启动前限时 Ping MySQL，失败即退出 | `/workspace/agents/internal/cmd/cmd.go:30-34,59-82` | `checkDatabase` + `pingDatabase` 5s 超时 Fatal | 已验证 |
| E-006 | 旧系统 Auth 四个接口已实现 | `/workspace/server/internal/http/router.go:52-57` | register/login 公开，me/logout 走 `AuthMiddleware` | 已验证 |
| E-007 | 旧系统 JWT 用 HS256，含 user_id/phone/session_id | `/workspace/server/internal/service/auth_service.go:233-249` | `jwt.SigningMethodHS256`，Issuer=`offer-catcher` | 已验证 |
| E-008 | 旧系统密码用 bcrypt DefaultCost | `/workspace/server/internal/service/auth_service.go:103,140` | `bcrypt.GenerateFromPassword` / `CompareHashAndPassword` | 已验证 |
| E-009 | 旧系统会话存 Redis，双 Key 设计 | `/workspace/server/internal/session/session.go:11-21` | `auth:session:{uid}:{sid}` + `auth:user_current:{uid}` | 已验证 |
| E-010 | 旧系统单设备顶替：新登录删除旧会话 | `/workspace/server/internal/session/session.go:108-129` | `ReplaceUserSession` 读旧指针→删旧→建新→更新指针 | 已验证 |
| E-011 | 旧系统顶替返回 401 语义 | `/workspace/server/internal/middleware/auth.go:49-53` | `errors.Is(err, ErrSessionReplaced)` → 41005 | 已验证 |
| E-012 | 旧系统内联鉴权重复（match 路由手写鉴权） | `/workspace/server/internal/http/router.go:145-201` | `dispatchJobWildcard` 内手写 ParseToken+ValidateSession，未复用中间件 | 已验证 |
| E-013 | 旧系统错误码 410xx 定义 | `/workspace/server/internal/errors/error.go:71-77,106-110` | 41002 无效凭据 / 41003 重复手机号 / 41004 会话无效(未用) / 41005 会话被顶替 | 已验证 |
| E-014 | 旧系统 users 表结构（phone 唯一索引） | `/workspace/server/sql/init.sql:661-671` | 字段 id/email/phone/password_hash/name/created_at/updated_at，`uk_users_phone` | 已验证 |
| E-015 | 旧系统配置默认值不安全 | `/workspace/server/internal/config/config.go:181-182,198,249-260` | `JWT_SECRET` 默认 `change-me-in-local`，JWT 168h，session TTL 604800s | 已验证 |
| E-016 | 旧系统 UserRepository 手写 SQL，无测试 | `/workspace/server/internal/repository/user_repository.go` | `database/sql` 手写 INSERT/SELECT，无 `*_test.go` | 已验证 |
| E-017 | 旧系统重复定义 Claims | `/workspace/server/internal/model/user.go:36-42` 与 `/workspace/server/internal/service/auth_service.go:34-39` | 两份 Claims，业务使用 service 包内那份 | 已验证 |
| E-018 | 旧系统 Redis 连接失败仅 warning 不退出 | `/workspace/server/cmd/server/main.go:59-61` | Redis 不可用时 session 校验全部失败返回 401 | 推断（据 main.go 注释与代码） |

### 2.5 根因

1. **agents 服务是全新空白骨架**：仅有健康检查，IAM 需从零实现，这是本轮工作的直接前提。
2. **旧系统 Auth 架构性缺陷**（本轮需避免重蹈）：
   - 手写 `net/http` + `ServeMux`，鉴权逻辑内联重复（`dispatchJobWildcard` 手写 ParseToken/ValidateSession）。
   - 巨型组合根 `main.go`（280 行手写装配），依赖通过构造器传递。
   - 错误语义分散：`errors.AppError` 中文文案与英文 message 混用，`CodeSessionInvalid`(41004) 定义未用。
   - 安全默认值不当（`JWT_SECRET` 默认弱密钥）。
   - 测试几乎空白，关键路径（会话顶替、中间件、JWT、Repository）无覆盖。
3. **Redis 会话是认证正确性的关键外部依赖**，旧系统对 Redis 故障无降级策略（连接失败仅 warning，导致所有受保护接口 401）。

### 2.6 影响范围

- 本轮仅影响 `agents` 服务，不触碰 `server` 旧系统。
- 对外影响：新增 `/api/auth/*` 四接口，需与旧系统前端契约对齐（`{code,message,data}` 包装、`{token,user}` 返回、410xx 错误码语义、单设备顶替 401）。
- 后续业务模块（简历/岗位/匹配/推荐/优化/对比）将依赖本 IAM 提供的统一鉴权中间件与用户上下文注入。

### 2.7 未验证事项

- 旧系统前端是否依赖 `AuthTokenResponse.user` 的具体字段形状（含 email/name 是否必须），需前端确认。
- 单设备顶替语义是否被产品接受为最终语义（当前为「单用户单会话」），需负责人确认。
- agents 服务与旧 `server` 服务未来是否共享同一 MySQL/Redis 实例、是否共享用户表，需负责人确认（当前 agents 使用独立库 `offer_catcher_re`）。
- JWT 过期时间（旧系统 168h）与 session TTL（旧系统 604800s）是否沿用，需负责人确认。
- 是否需要在 agents 侧保留旧系统 `users` 表的 seed 数据（旧 init.sql 有 3 个测试用户），需负责人确认。

## 3. 候选方案与推荐方案

### 3.1 方案 A：GoFrame 规范分层 + gdb DAO + Redis 会话（推荐）

- 做法：
  - 遵循 GoFrame v2 规范分层：`api/`（请求响应结构）→ `internal/controller/`（HTTP 控制器）→ `internal/logic/`（业务逻辑）→ `internal/dao/` + `internal/model/`（gdb 数据访问）→ `internal/middleware/`（鉴权中间件）。
  - 使用 `gdb`（已随 MySQL 驱动引入）替代手写 `database/sql`；`users` 表结构用 `gf gen dao` 或手写 model 生成。
  - JWT 引入 `golang-jwt/jwt/v5`，bcrypt 引入 `golang.org/x/crypto/bcrypt`。
  - 会话沿用 Redis 双 Key 设计，引入 `go-redis/v9`，封装 `internal/logic/session` 或 `internal/service/session`。
  - 鉴权统一走 GoFrame 路由组中间件（替换旧系统内联鉴权），用户信息通过 GoFrame 请求上下文注入。
  - 错误语义统一收敛到一套 `internal/errors`（保留 410xx 业务码与前端可读文案）。
  - 配置全部走 `manifest/config/config.yaml` + 环境变量，JWT secret 不设弱默认值，生产环境强制校验。

- 优点：
  - 完全契合 agents 现有 GoFrame 骨架与「生产级、规范」目标。
  - gdb 自带连接池、事务、结构化查询，可测试性显著优于手写 SQL。
  - 分层清晰，后续业务模块可复用鉴权中间件与用户上下文。
  - 错误语义单一化，消除旧系统三套错误码并存问题。

- 缺点：
  - 需新增 JWT/bcrypt/Redis 三个依赖，Redis 仍是外部强依赖。
  - 从零编写工作量中等（controller/logic/dao/middleware/errors/session 六层 + 测试）。

- 风险：
  - Redis 故障导致鉴权不可用（需明确 fallback 语义）。
  - 前端契约对齐风险（`{token,user}` 字段形状、单设备顶替 401）。

- 成本：中（约 6 层代码 + 契约/集成测试，但骨架已就绪，gdb/响应包装中间件已存在，复用度高）。

- 回滚方式：IAM 为独立新增模块，不触碰既有 `/healthz`，删除新增包即可回滚；DB 迁移需保留幂等 down 脚本。

### 3.2 方案 B：GoFrame 分层 + 无状态 JWT（移除 Redis 会话）

- 做法：
  - 同方案 A 的分层与 gdb 依赖，但**不引入 Redis**，JWT 完全无状态（仅凭签名与过期时间校验，不维护服务端会话）。
  - 登出退化为「客户端丢弃 token」，服务端无法主动失效；**放弃单设备顶替语义**。

- 优点：
  - 移除 Redis 外部依赖，部署与测试更简单，无 Redis 故障风险。
  - 实现成本最低。

- 缺点：
  - 丢失「账号已在其他设备登录」的顶替能力，与旧系统前端契约（41005）不兼容。
  - 登出无法服务端生效，安全上不可主动吊销 token。
  - 若未来需要会话管理，需返工重引入 Redis。

- 风险：前端/产品若依赖单设备登录语义，将产生契约回归。

- 成本：低，但牺牲能力与契约兼容。

- 回滚方式：易回滚，但若要恢复顶替语义需重引入 Redis。

### 3.3 方案对比

| 维度 | 方案 A（Redis 会话） | 方案 B（无状态 JWT） |
|---|---|---|
| 与旧系统契约兼容（单设备顶替 41005） | 完全兼容 | 不兼容 |
| 外部依赖 | MySQL + Redis | 仅 MySQL |
| 登出服务端生效 | 是 | 否 |
| 实现成本 | 中 | 低 |
| 可测试性 | 高（Redis 可 mock，gdb 可测） | 高 |
| 风险 | Redis 故障需 fallback | 丢失顶替语义 |
| 回滚难度 | 低（独立模块） | 低 |

### 3.4 推荐方案

**方案 A**：GoFrame 规范分层 + gdb DAO + Redis 会话（保留单设备顶替语义）。

### 3.5 推荐理由

1. 旧系统对外契约明确包含「单设备顶替、41005 账号已在其他设备登录」语义，前端已按此实现，方案 A 可无损对齐，避免契约回归。
2. 方案 A 与 agents 现有 GoFrame 骨架、已引入的 MySQL 驱动、统一响应包装中间件高度契合，复用度最高，符合「规范、生产级」目标。
3. Redis 会话是成熟方案，其故障可通过明确的 fallback 语义（见第 4 章）约束，风险可控；方案 B 虽省 Redis 但直接放弃既有能力，得不偿失。
4. 方案 A 的分层与统一鉴权中间件，是后续业务模块接入 IAM 的必要前提。

### 3.6 需要负责人决定的事项

1. **会话语义**：是否确认沿用「单用户单会话（新登录顶替旧会话，旧 token 返回 401）」？
2. **外部依赖**：agents 是否共享旧 `server` 的 MySQL/Redis 实例？用户数据表是否独立（当前 agents 库为 `offer_catcher_re`）？
3. **配置参数**：JWT 过期时间（旧系统 168h）与 session TTL（旧系统 604800s）是否沿用？
4. **前端契约**：`AuthTokenResponse.user` 字段形状（id/phone/name/email）以哪个为准？
5. **安全默认值**：JWT secret 是否要求生产环境强制非弱密钥（旧系统默认 `change-me-in-local` 仅告警不阻断）？
6. **是否本轮引入 RBAC/权限点**（默认不引入，仅认证）。

## 4. 重构契约

- 契约版本：DRAFT
- 契约状态：WAITING_FOR_APPROVAL

### 4.1 输入

| 接口 | 输入 |
|---|---|
| `POST /api/auth/register` | `{ "phone": string, "password": string }`（phone 中国大陆手机号 `^1[3-9]\d{9}$`，password ≥ 6 位） |
| `POST /api/auth/login` | `{ "phone": string, "password": string }` |
| `GET /api/auth/me` | 无 body，`Authorization: Bearer <token>` |
| `POST /api/auth/logout` | 无 body，`Authorization: Bearer <token>` |

### 4.2 输出

统一包装 `{ "code": int, "message": string, "data": object|null }`。

- `register` / `login` 成功：`data = { "token": string, "user": { "id", "phone", "name" } }`（`email` 字段待负责人确认是否保留）。
- `me` 成功：`data = { "id", "phone", "name" }`。
- `logout` 成功：`data = { "message": "已退出登录" }`。
- 失败：`data = null`，`code` 与 `message` 见 4.6。

### 4.3 接口边界

- `register` / `login` 为公开接口（无鉴权中间件）。
- `me` / `logout` 为受保护接口（统一走鉴权中间件，不内联鉴权）。
- 鉴权中间件职责：解析 `Authorization: Bearer <token>` → 校验 JWT 签名/过期 → 校验 Redis 会话与当前指针 → 将 `userID` / `sessionID` 注入请求上下文；任何一步失败即返回对应错误，不进入业务控制器。
- 后续业务模块（简历/岗位/匹配等）复用同一鉴权中间件与上下文键，不得复制鉴权逻辑。

### 4.4 数据结构或表结构

**users 表**（`offer_catcher_re` 库，与旧系统字段语义对齐）：

```sql
CREATE TABLE `users` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `email` varchar(255) NOT NULL DEFAULT '',
  `phone` varchar(20) NOT NULL DEFAULT '',
  `password_hash` varchar(255) NOT NULL,
  `name` varchar(100) NOT NULL DEFAULT '',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_users_phone` (`phone`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';
```

**JWT Claims**：`user_id`、`phone`、`session_id` + 标准 `exp/iat/nbf/iss`（issuer=`offer-catcher`，HS256）。

**Redis 会话 Key**（双 Key，沿用旧系统语义）：
- `auth:session:{userID}:{sessionID}` = `"1"`（活跃会话，TTL = session TTL）
- `auth:user_current:{userID}` = `{sessionID}`（当前会话指针，TTL = session TTL）

### 4.5 行为不变量

1. 密码必须以 bcrypt 哈希存储，禁止明文落库；校验用 `CompareHashAndPassword`。
2. 登录/注册成功必定创建新会话、签发新 token，并顶替旧会话（单用户单会话）。
3. 任意请求通过鉴权中间件时，必须同时满足：JWT 有效 + Redis 会话存在且为 `"1"` + 当前指针等于 token 内 `sessionID`。
4. 顶替后旧 token 请求返回 401（41005），不进入业务逻辑。
5. 登出仅清除当前会话；旧设备登出不得删除新设备的当前指针（防止误删）。
6. 鉴权中间件注入的 `userID`/`sessionID` 类型稳定，业务层依赖的上下文键唯一且不冲突。
7. 错误响应统一走 `{code,message,data}`，HTTP 状态码与业务码分离（见 4.6）。

### 4.6 错误语义

| 业务码 | HTTP 状态 | 场景 | message（前端可读） |
|---|---|---|---|
| 40001 | 400 | 参数非法（手机号格式/密码长度等） | 对应中文提示 |
| 40100 | 401 | 缺 token / token 解析失败 / 会话无效 | unauthorized |
| 41002 | 401 | 手机号或密码错误 | 手机号或密码错误 |
| 41003 | 409 | 手机号已注册 | 该手机号已被注册 |
| 41005 | 401 | 会话被顶替 | 账号已在其他设备登录，请重新登录 |
| 40400 | 404 | 用户不存在（me） | not found |
| 50000 | 500 | 内部错误 | internal server error |

> 注：旧系统 `CodeSessionInvalid`(41004) 已定义但从未使用，本轮收敛为「40100/会话无效」，不引入无用码。

### 4.7 fallback 语义

- **Redis 不可用**：鉴权中间件在会话校验失败且非「顶替」语义时，统一返回 40100（unauthorized），**不降级放行**（安全优先）；同时记录明确日志。启动阶段 Redis 故障是否阻断启动，由负责人决定（旧系统为仅告警不阻断）。
- **MySQL 不可用**：沿用 agents 现有「启动前限时 Ping 失败即 Fatal」约定；运行期 DB 异常返回 50000，不吞错。
- **JWT secret 缺失/弱密钥**：生产环境（非 local）必须启动失败或强制告警并拒绝服务，不得静默使用弱默认值（待负责人确认强度策略）。

### 4.8 允许修改

- 在 `agents` 下新增 `api/auth/`、`internal/controller/auth/`、`internal/logic/auth/`、`internal/logic/session/`、`internal/dao/`、`internal/model/`、`internal/middleware/auth.go`、`internal/errors/` 等包。
- 修改 `manifest/config/config.yaml`（新增 JWT/session 配置项）。
- 修改 `go.mod`/`go.sum`（新增 jwt、bcrypt、redis 依赖）。
- 修改 `sql/init.sql` 或新增迁移脚本（建 `users` 表）。
- 修改 `main.go`/`cmd.go` 以装配 IAM 依赖（保持既有 `/healthz` 与启动预检行为不变）。

### 4.9 禁止修改

- 禁止修改 `/workspace/server` 旧系统代码。
- 禁止破坏既有 `/healthz` 接口及其响应形状（`{service,env,app,database}`）。
- 禁止删除/改动现有 `DBHealth` 中间件与启动 DB 预检语义。
- 禁止在业务控制器内内联鉴权（必须走统一中间件）。
- 禁止明文存储密码、禁止弱 JWT 默认密钥用于生产。

### 4.10 范围外事项

- RBAC、角色、权限点、多租户、OAuth、短信/邮箱验证码。
- 四大模型功能（匹配/推荐/优化/对比）的迁移。
- 前端改造与联调（本轮仅后端）。
- 旧 `server` 服务的下线或数据迁移（不在本轮）。

### 4.11 回滚要求

- IAM 为独立新增模块，回滚 = 移除新增包与路由注册，不影响 `/healthz`。
- DB 迁移需提供幂等 down 脚本（`DROP TABLE IF EXISTS users`）。
- 会话/用户数据不回滚到旧系统（旧系统与 agents 库隔离，无交叉污染）。

## 5. 验收矩阵

| ID | 场景 | 初始条件 | 操作 | 预期结果 | 测试层级 | 是否需要真实外部服务 |
|---|---|---|---|---|---|---|
| AC-001 | 健康检查仍可用 | 服务启动 | GET /healthz | 返回 200，`{service,env,app,database}` 结构不变 | 契约 | 否（DB 状态可为 unavailable） |
| AC-002 | 注册成功 | 合法 phone/password，手机号未注册 | POST /api/auth/register | code=0，返回 `{token,user}`，users 表新增记录，密码为 bcrypt 哈希 | 集成 | 是（MySQL + Redis） |
| AC-003 | 注册重复手机号 | 手机号已存在 | POST /api/auth/register | code=41003，HTTP 409 | 集成 | 是（MySQL） |
| AC-004 | 注册非法手机号 | 空/格式错误 phone | POST /api/auth/register | code=40001，HTTP 400 | 单元 | 否 |
| AC-005 | 注册弱密码 | 密码 < 6 位 | POST /api/auth/register | code=40001，HTTP 400 | 单元 | 否 |
| AC-006 | 登录成功 | 已注册用户 | POST /api/auth/login | code=0，返回 `{token,user}` | 集成 | 是（MySQL + Redis） |
| AC-007 | 登录错误密码 | 已注册用户 | 错误密码登录 | code=41002，HTTP 401 | 集成 | 是（MySQL） |
| AC-008 | 登录不存在手机号 | 手机号未注册 | 登录 | code=41002，HTTP 401 | 集成 | 是（MySQL） |
| AC-009 | me 返回当前用户 | 有效 token | GET /api/auth/me | code=0，返回当前用户信息 | 集成 | 是（MySQL + Redis） |
| AC-010 | me 无 token | 无 Authorization | GET /api/auth/me | code=40100，HTTP 401 | 契约 | 否 |
| AC-011 | me 非法 token | 伪造/过期 token | GET /api/auth/me | code=40100，HTTP 401 | 单元 | 否（mock JWT） |
| AC-012 | 单设备顶替 | 用户登录两次 | 第二次登录后，用第一个 token 访问 me | code=41005，HTTP 401，message=账号已在其他设备登录 | 集成 | 是（Redis） |
| AC-013 | 登出后访问失效 | 有效 token 登出 | 登出后用原 token 访问 me | code=40100，HTTP 401 | 集成 | 是（Redis） |
| AC-014 | 登出不影响新会话 | 登录→登出→再登录 | 登出后，新登录 token 正常访问 | code=0 | 集成 | 是（Redis） |
| AC-015 | JWT 签名算法混淆防护 | 构造 HS/RS 混淆 token | 访问受保护接口 | code=40100，拒绝非 HMAC 算法 | 单元 | 否 |
| AC-016 | bcrypt 校验 | 已知哈希与明文 | CompareHashAndPassword | 匹配返回 nil，不匹配返回错误 | 单元 | 否 |
| AC-017 | 会话指针唯一性 | 多会话并发登录 | 校验当前指针 | 仅最新 session 有效 | 单元（mock Redis） | 否 |
| AC-018 | 错误响应统一包装 | 任意失败请求 | 触发错误 | 返回 `{code,message,data:null}`，HTTP 状态正确 | 契约 | 否 |
| AC-019 | 中间件注入上下文 | 有效 token | 受保护接口内读取 userID/sessionID | 注入正确，类型稳定 | 单元 | 否（mock） |
| AC-020 | Redis 故障 fallback | Redis 不可达 | 访问受保护接口 | 返回 40100，不降级放行，有明确日志 | 契约/单元 | 否（mock Redis 失败） |
| AC-021 | users 表结构契约 | 空库执行迁移 | 建表 | users 表字段/索引符合 4.4 | 契约 | 是（MySQL） |
| AC-022 | 配置缺 JWT secret 行为 | 生产 env，缺 secret | 启动 | 按负责人策略：拒绝启动或强制告警 | 契约/单元 | 否 |

## 6. Coder 实现记录

由 Coder 编写。

## 7. Cleaner 初次审查

由 Cleaner 编写。

## 8. Fixer 修复记录

由 Fixer 编写。

## 9. Cleaner 复审结果

由 Cleaner 编写。

## 10. Deliverer 交付验收

由 Deliverer 编写。

## 11. 最终交付决定

由项目负责人编写。

## 12. 流转记录

| 时间 | 契约版本 | 当前提交 | 执行角色 | 执行结果 | 下一状态 | 下一负责人 |
|---|---|---|---|---|---|---|
| 2026-09-20 | DRAFT | 2a7f17d | Analyst | 完成证据梳理、候选方案、契约与验收矩阵起草 | WAITING_FOR_OWNER_APPROVAL | 项目负责人 |
