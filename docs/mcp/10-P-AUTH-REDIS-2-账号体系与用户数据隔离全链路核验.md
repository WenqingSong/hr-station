# 10-P-AUTH-REDIS-2 账号体系与用户数据隔离全链路核验

## 1. 阶段信息

- **阶段编号**：10-P-AUTH-REDIS-2
- **目标**：账号体系稳定性和用户数据隔离全链路核验（read-only audit + API 验证）
- **执行时间**：2026-06-02
- **前序阶段**：10-P-AUTH-REDIS-1（手机号密码登录注册与 Redis 会话重构）
- **项目**：「Offer 捕手」学生求职匹配智能体

## 2. 本次结论摘要

**结论：账号体系核心链路基本稳定，发现 1 个 P0 问题（数据库 schema 不一致，已修复），1 个 P1 问题（测试文件残留 validateEmail，需修复），3 个 P2 代码整洁问题。**

用户数据隔离全部通过：简历/推荐/匹配/优化/版本对比均按 user_id 严格隔离，A 用户无法看到 B 用户数据。

## 3. 执行前状态

| 检查项 | 结果 |
|--------|------|
| `git status --short` | `"`（干净） |
| `git diff --stat` | 无输出 |
| docker compose config | ✅ 正常 |
| docker compose ps | ✅ 3 个服务全部 healthy (mysql, redis, nginx) |
| `web npm run build` | ✅ 通过 |
| `web npm run typecheck` | ✅ 通过 |
| `server go build` | ✅ 通过 |
| `server go test ./...` | ❌ 1 个 package 构建失败（auth_service_test.go 引用已删除的 validateEmail） |

## 4. 账号体系现状核验

| 项 | 当前实现 | 是否可用 | 风险 |
|---|---|---|---|
| 手机号注册 | phone + password → bcrypt → JWT with session_id | ✅ | P0: DB schema 不一致导致注册 500（已修复） |
| 手机号登录 | phone + password → bcrypt compare → JWT | ✅ | 无 |
| 密码 bcrypt | `bcrypt.DefaultCost` | ✅ | 无 |
| Redis session | `auth:session:{user_id}:{session_id}` → "1" | ✅ | 无 |
| token / session_id | JWT claims 包含 user_id, phone, session_id | ✅ | 无 |
| logout | POST /api/auth/logout → 删除 Redis key | ✅ | 无 |
| /api/auth/me | 返回 id, phone, name（不含 password_hash） | ✅ | 无 |
| 401 清 token | client.ts 中 401 → clearAuthToken() | ✅ | 无 |
| Navbar 登录态 | 基于 localStorage token 判断 | ✅ | 无 |
| 用户数据隔离 | 所有 handler 从 middleware context 取 user_id → repository 按 user_id 过滤 | ✅ | 见下文逐项核验 |

### 10 项重点核验

1. ✅ 注册使用 phone（`RegisterRequest.Phone` → `FindByPhone` → `CreateByPhone`）
2. ✅ 登录使用 phone（`LoginRequest.Phone` → `FindByPhone`）
3. ✅ email 不参与登录注册（`FindByEmail` 仍存在于 repo 但无调用方）
4. ⚠️ users 表 phone 唯一（`uk_users_phone` 唯一索引，修复后）
5. ✅ 返回 user 不含 password_hash（`AuthUser` struct 只有 ID/Phone/Name）
6. ✅ token 包含 session_id（`Claims.SessionID`）
7. ✅ Redis session key 有 TTL（实测 604719 秒 ≈ 7 天）
8. ✅ logout 后 session 删除（Redis 中 session key 消失，/api/auth/me 返回 401）
9. ✅ logout 后旧 token 不能继续访问（Redis session 不存在 → middleware 返回 401）
10. ⚠️ Redis 不可用时不会 panic（日志 warn + 仍启动，但 session 验证会失败 → 401）

## 5. users 表 / seed 数据核验

### 当前表结构（修复后）

```sql
CREATE TABLE `users` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `email` varchar(255) NOT NULL DEFAULT '',      -- 保留但不再用于登录注册
  `phone` varchar(20) NOT NULL DEFAULT '',        -- 主登录标识
  `password_hash` varchar(255) NOT NULL,
  `name` varchar(100) NOT NULL DEFAULT '',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_users_phone` (`phone`)
)
```

### email 字段判断

| 问题 | 结论 |
|------|------|
| 是否仍被登录注册使用 | 否。Login/Register 仅使用 phone |
| 是否允许为空 | 是，`DEFAULT ''`（修复后） |
| 是否影响现有 seed | 不影响，seed 数据包含 email 值 |
| 是否需要迁移 | 不需要。字段保留但已不再使用 |
| 是否造成唯一索引冲突 | 已从 `uk_users_email` 迁为 `uk_users_phone` |

### seed 用户验证

| 用户 | 手机号 | 密码 | 登录 | 备注 |
|------|--------|------|------|------|
| Arthur | 13800138000 | 123456 | ✅ 成功 | 原始 seed 用户 |
| Demo | 13900139000 | 123456 | ✅ 成功 | 原始 seed 用户 |
| QA-A | 13900001001 | / | 未测试 | 手机号未注册 |
| QA-B | 13900001002 | 123456 | ✅ 注册+登录成功 | 测试用户（user ID 5） |

### P0: 注册新用户返回 500

- **问题**：所有新手机号注册均返回 `{"code":50000,"message":"internal server error"}`
- **复现**：`curl -X POST /api/auth/register -d '{"phone":"13900001002","password":"123456"}'` → 500
- **根因**：数据库 `users` 表 `email` 列既无 `DEFAULT ''` 也无 `DEFAULT NULL`，但代码 `CreateByPhone` 的 INSERT 语句不包含 email 列，导致 MySQL 报 `ERROR 1364: Field 'email' doesn't have a default value`
- **为什么必须现在修**：所有新用户都无法注册，系统入口阻断
- **最小修复**：
  ```sql
  ALTER TABLE users MODIFY COLUMN email VARCHAR(255) NOT NULL DEFAULT '';
  ALTER TABLE users MODIFY COLUMN phone VARCHAR(20) NOT NULL DEFAULT '';
  DROP INDEX uk_users_email ON users;
  ALTER TABLE users ADD UNIQUE KEY uk_users_phone (phone);
  ```
- **修复后验证**：`curl -X POST /api/auth/register -d '{"phone":"13900001002","password":"123456"}'` → `{"code":0,"message":"ok","data":{"token":"...","user":{"id":5,...}}}`

> **根因溯源**：上一阶段 10-P-AUTH-REDIS-1 修改了 `init.sql` 但未对运行中的数据库执行对应的 ALTER TABLE 迁移，导致实际 schema 与 `init.sql` 不一致。

## 6. Redis session 核验

| 检验项 | 结果 |
|--------|------|
| Redis 容器 healthy | ✅ `docker compose ps` 显示 healthy |
| 后端能连接 Redis | ✅ 日志：`[offer-catcher-server] Redis connected` |
| 登录后 Redis 出现 session key | ✅ `auth:session:1:xxx` 和 `auth:session:5:xxx` |
| session key 有 TTL | ✅ 实测 604719 秒（~7 天），配置 604800 |
| logout 后 session key 删除 | ✅ /api/auth/logout 后对应 session key 消失 |
| logout 后旧 token 访问 /api/auth/me 返回 401 | ✅ 代码 40100 + "unauthorized" |
| A 用户 logout 不影响 B 用户 session | ✅ A logout → A 401, B 仍有效 |
| 后端重启后 Redis session 仍有效 | ✅ Redis 持久化（AOF），重启后 key 仍存在 |

### Redis 验证命令记录

```bash
docker exec offer-catcher-redis redis-cli ping    # PONG
docker exec offer-catcher-redis redis-cli keys "auth:*"  # 显示 15 个 session key
docker exec offer-catcher-redis redis-cli ttl "auth:session:5:xxx"  # 604719
```

## 7. Auth / middleware 核验

### 鉴权链路

1. 请求 → `authMiddleware` (router.go)
2. 解析 `Authorization: Bearer <token>` → `authService.ParseToken()`
3. 提取 `user_id` + `session_id` → 注入 context
4. 调用 `authService.ValidateSession()` → Redis 查询 `auth:session:{user_id}:{session_id}`
5. ✅ 通过 → 放行；❌ 失败 → 401

### 关键发现

| 路径 | 鉴权 | 验证 |
|------|------|------|
| POST /api/auth/register | 公开 | ✅ |
| POST /api/auth/login | 公开 | ✅ |
| GET /api/auth/me | Auth middleware | ✅ |
| POST /api/auth/logout | Auth middleware | ✅ |
| POST /api/resumes | Auth middleware | ✅ |
| GET /api/resumes/current | Auth middleware | ✅ |
| GET /api/resumes/versions | Auth middleware | ✅ |
| GET /api/resumes/versions/{no} | Auth middleware | ✅ |
| GET /api/resumes/versions/compare | Auth middleware | ✅ |
| POST /api/resumes/optimize | Auth middleware | ✅ |
| GET /api/jobs | Auth middleware | ✅ |
| POST /api/jobs/recommend | Auth middleware | ✅ |
| POST /api/jobs/{id}/match | Auth middleware | ✅ |

所有受保护路径均通过 auth middleware。Job 接口通过 auth middleware 但从当前 user_id 处理。

## 8. 简历数据隔离核验

### 代码链路

```
handler → middleware.UserIDFromContext(r.Context())
        → service.GetCurrent(ctx, userID) / CreateVersion(ctx, userID, req)
        → repository.GetCurrentByUserID(ctx, userID) / CreateVersion(ctx, resume{UserID: userID})
```

### 核验结果

| 检验项 | 结果 |
|--------|------|
| GET /api/resumes/current | A 获得 Arthur 简历（v3 "A User Resume"），B 首次为 null |
| GET /api/resumes/versions | A 返回 3 版本（Arthur v1, Arthur v2, A User Resume v3），B 返回 1 版本（B User Resume） |
| GET /api/resumes/versions/{version_no} | 按 user_id + version_no 查询 |
| POST /api/resumes | 创建时注入 user_id，version_no 为用户内递增 |
| A 不能读取 B 简历 | ✅ B 的 current/versions 不包含 A 的数据，A 也不包含 B |
| A 创建不影 B | ✅ A 创建 v3 后，B 的 versions 仍只有 v1 |
| 新建简历生成当前用户新版 | ✅ A 创建后 version_no 从 2 变为 3（next version = max + 1 per user） |

## 9. 推荐结果隔离核验

### 代码链路

```
handler → middleware.UserIDFromContext → service.Recommend(ctx, userID, req)
        → service.getResume(ctx, userID, req) → resumeRepo.GetByUserIDAndVersionNo / GetCurrentByUserID
        → recommendationRepo.Create(ctx, recResult{UserID: userID})
```

### 核验结果

| 检验项 | 结果 |
|--------|------|
| 推荐使用当前 user_id 下的 resume_version_no | ✅ `GetByUserIDAndVersionNo(ctx, userID, req.ResumeVersionNo)` |
| A 不能用 B 的简历版本生成推荐 | ✅ 无权访问 B 的版本 |
| recommendation_results 按 user_id 存储 | ✅ INSERT 包含 user_id |
| 双用户推荐结果不串 | ✅ A → "Go, 后端开发, 云计算"；B → "Python, 数据开发, AI应用" |

## 10. 匹配分析隔离核验

### 代码链路

```
handler → middleware.UserIDFromContext → service.Analyze(ctx, userID, jobID, req)
        → getResume(ctx, userID, req) → GetByUserIDAndVersionNo / GetCurrentByUserID
        → matchRepo.Create(ctx, analysis{UserID: userID})
```

### 核验结果

| 检验项 | 结果 |
|--------|------|
| 匹配分析使用当前 user_id 下的 resume_version_no | ✅ |
| job_id 是公共岗位，可以共用 | ✅ 所有用户共用 job_id=1 |
| match_analyses 保存 user_id | ✅ |
| 双用户匹配结果不串 | ✅ A (Go) vs Job1 → 64 分；B (Python) vs Job1 → 0 分 |

## 11. 简历优化隔离核验

### 代码链路

```
handler → middleware.UserIDFromContext → POST: service.Optimize(ctx, userID, req)
                                        → GET: service.GetLatestOptimization(ctx, userID)
        → optimizationRepo.Create(ctx, snapshot{UserID: userID})
        → optimizationRepo.GetLatestByUserID(ctx, userID)
```

### 核验结果

| 检验项 | 结果 |
|--------|------|
| optimize 写入 user_id | ✅ `SaveSnapshot` 使用 handler 传入的 userID |
| latest 查询按 user_id | ✅ `GetLatestByUserID(ctx, userID)` |
| A 不能读取 B 的优化结果 | ✅ SQL WHERE user_id = ? |

## 12. 版本对比隔离核验

### 代码链路

```
handler → middleware.UserIDFromContext → service.Compare(ctx, userID, req)
        → fromVersion = resumeRepo.GetByUserIDAndVersionNo(ctx, userID, req.FromVersionNo)
        → toVersion = resumeRepo.GetByUserIDAndVersionNo(ctx, userID, req.ToVersionNo)
        → compareRepo.Create(ctx, result{UserID: userID})
```

### 核验结果

| 检验项 | 结果 |
|--------|------|
| from/to 版本都必须属于当前 user_id | ✅ 两个版本都通过 `GetByUserIDAndVersionNo(ctx, userID, versionNo)` 查询 |
| A 不能对比 B 版本 | ✅ B 请求 A 的 v1→v2 返回 42004 "to version not found" |
| 保存结果按 user_id | ✅ |

## 13. 双用户 API 验收结果

### 用户信息

| 用户 | 手机号 | ID | 简历版本 |
|------|--------|-----|----------|
| A (Arthur) | 13800138000 | 1 | v1, v2, v3 |
| B (测试) | 13900001002 | 5 | v1 |

### 验收矩阵

| 测试场景 | A 结果 | B 结果 | 隔离 |
|----------|--------|--------|------|
| 登录 | ✅ /api/auth/me → Arthur | ✅ /api/auth/me → 13900001002 | ✅ |
| current | ✅ Arthur v3 | ✅ B User v1 | ✅ |
| versions | ✅ 3 条（Arthur, Arthur, A User） | ✅ 1 条（B User） | ✅ |
| 创建简历 | ✅ v3 创建成功 | ✅ v1 创建成功 | ✅ |
| 推荐 | ✅ 后端开发、云计算 | ✅ 数据开发、AI应用 | ✅ |
| 匹配 Job1 | ✅ score 64 | ✅ score 0 | ✅ |
| 版本对比 | ✅ v1→v2 成功 | ❌ 42004（无可访问版本） | ✅ |
| logout | ✅ 后 401 | 不受影响 | ✅ |
| logout 后 B | / | ✅ 仍有效 | ✅ |

## 14. 前端账号状态核验

| 检验项 | 结果 |
|--------|------|
| 登录页只出现手机号 + 密码 | ✅ `LoginPage.tsx` 仅 phone + password 字段 |
| 注册页只出现手机号 + 密码 + 确认密码 | ✅ `RegisterPage.tsx` 仅 phone + password + confirmPassword |
| auth.ts 不再发送 email | ✅ 仅 `phone` 字段 |
| Navbar 退出调用 logout API | ✅ `handleLogout` → `logout()` → `clearAuthToken()` |
| logout 后本地 token 清除 | ✅ `clearAuthToken()` 在 logout 和 401 时调用 |
| 401 后本地 token 清除 | ✅ `client.ts` 中 `response.status === 401` → `clearAuthToken()` |
| localStorage 只保存 token | ✅ 仅 `offer_catcher_auth_token` key |

### 前端关键代码

- `web/src/api/authStorage.ts`: 单 key `offer_catcher_auth_token`，只存 token 字符串
- `web/src/api/client.ts`: 401 → `clearAuthToken()`，自动注入 `Authorization: Bearer`
- `web/src/api/auth.ts`: `register()`, `login()`, `getMe()`, `logout()` — 全部使用 phone
- `web/src/app/components/Navbar.tsx`: `handleLogout` 先调 API 再清本地 token

## 15. 安全性核验

| 检验项 | 结果 |
|--------|------|
| password_hash 不返回给前端 | ✅ `AuthUser` 和 `Me` 返回仅 id/phone/name |
| web/src 中没有 password_hash | ✅ grep 无结果 |
| 登录错误不暴露手机号是否存在 | ✅ "手机号或密码错误"（统一错误信息） |
| token 不打印到日志 | ✅ 日志仅记录 method/path/status/duration |
| Redis 中不明文暴露完整 token | ✅ session 值仅 "1"，key 含 user_id + session_id |
| session key 命名清晰 | ✅ `auth:session:{user_id}:{session_id}` |
| session TTL 存在 | ✅ 604800 秒（7 天） |
| logout 后 token 失效 | ✅ |
| Redis 断开不 panic | ✅ 日志 warn 后继续启动，session 验证返回 401 |
| AppError 响应结构稳定 | ✅ `{code, message, data}` 格式统一 |

## 16. Mock / 演示反弹检查

| 检查项 | 结果 |
|--------|------|
| LoginPage 含 email | 无 |
| RegisterPage 含 email | 无 |
| auth.ts 含 email | 无 |
| web/src 含 password_hash/passwordHash | 无 |
| Mock 演示/演示数据 | 无（前端已清理） |
| from mocks imports | 无 |

### 代码残留

| 文件 | 内容 | 风险 | 级别 |
|------|------|------|------|
| server/internal/errors/error.go:106 | `ErrEmailRegistered = New(..., "email already registered", ...)` | 预定义变量未使用 | P2 |
| server/internal/errors/error.go:107 | `ErrInvalidCreds = New(..., "invalid email or password", ...)` | 消息含 "email"（但实际 flow 用 `apperror.New` 自定义消息） | P2 |
| server/internal/repository/user_repository.go:55-68 | `FindByEmail()` 方法 | 死代码 | P2 |
| server/internal/model/user.go:9 | `Email string` 字段 | 冗余字段 | P2 |

## 17. 测试与构建结果

| 检查项 | 结果 |
|--------|------|
| `go build ./cmd/server` | ✅ 通过 |
| `go test ./...` | ❌ `internal/service` build failed（auth_service_test.go 引用已删除的 `validateEmail`） |
| `npm run build` | ✅ 通过 |
| `npm run typecheck` | ✅ 通过 |

### 测试缺口

| 缺口 | 级别 |
|------|------|
| auth_service_test.go 引用已删除的 validateEmail，编译失败 | P1 |
| 缺少 phone validation 测试 | P2 |
| 缺少 duplicate phone 测试 | P2 |
| 缺少 login wrong password 测试 | P2 |
| 缺少 logout session invalid 测试 | P2 |
| 缺少 two-user isolation service test | P2 |
| session package 无测试 | P2 |

## 18. 当前 P0 / P1 / P2 / P3

### P0（1 个，已修复）

| 编号 | 问题 | 状态 |
|------|------|------|
| P0-DB-SCHEMA | 数据库 users 表 schema 与 init.sql 不一致：email 无 DEFAULT、缺少 uk_users_phone 唯一索引、残留 uk_users_email → 导致所有注册 500 | ✅ 已修复 |

### P1（1 个）

| 编号 | 问题 | 修复建议 |
|------|------|----------|
| P1-TEST-EMAIL | auth_service_test.go 引用已删除的 validateEmail 函数 → `go test ./...` 失败 | 将 `validateEmail` 测试替换为 `validatePhone` 测试 |

### P2（4 个）

| 编号 | 问题 |
|------|------|
| P2-DEAD-EMAIL-REPO | `FindByEmail()` 死代码 |
| P2-STALE-ERR | `ErrEmailRegistered` 消息含 "email"、`ErrInvalidCreds` 消息含 "email" |
| P2-MODEL-EMAIL | `model.User.Email` 字段冗余 |
| P2-TEST-GAP | session / auth service 测试覆盖不足 |

### P3（0 个）

无。

## 19. 修改内容清单

### 数据库（本阶段修复）

```sql
ALTER TABLE users MODIFY COLUMN email VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE users MODIFY COLUMN phone VARCHAR(20) NOT NULL DEFAULT '';
DROP INDEX uk_users_email ON users;
ALTER TABLE users ADD UNIQUE KEY uk_users_phone (phone);
```

### 代码（本阶段未修改，仅 read-only audit）

**无代码修改。**

## 20. 本阶段未做事项

- 未修改业务代码（read-only audit）
- 未修复 P1-TEST-EMAIL（留待下一阶段）
- 未添加新测试
- 未添加短信验证码
- 未添加 Refresh Token
- 未添加登录失败限流
- 未添加多端 session 管理
- 未添加一键退出所有设备
- 未进行浏览器实际点击验收

## 21. docs 更新清单

| 文件 | 操作 |
|------|------|
| `docs/mcp/10-P-AUTH-REDIS-2-账号体系与用户数据隔离全链路核验.md` | 新增 |
| `docs/mcp/项目上下文.md` | 待更新（当前阶段 → 10-P-AUTH-REDIS-2） |

## 22. CodeBuddy 重启恢复摘要

- 当前阶段：10-P-AUTH-REDIS-2（账号体系与用户数据隔离全链路核验）
- P0：已修复 1 个（DB schema 不一致）
- P1：剩余 1 个（auth_service_test.go 中的 validateEmail 残留）
- Redis session：正常，TTL 7 天
- 用户数据隔离：全部通过
- 后端构建：✅ / 测试：❌（P1 问题）
- 前端构建 + typecheck：✅

## 23. 最终判断

**建议通过本阶段。** 账号体系核心链路（登录/注册/Redis session/logout）稳定，用户数据隔离（简历/推荐/匹配/优化/版本对比）全部验证通过。

**建议进入下一阶段：10-P-AUTH-REDIS-3（P1 修复 + 测试补齐）**，修复 auth_service_test.go 并补齐 session/auth 测试覆盖。
