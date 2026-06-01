# 10-P-4-3 Auth / 用户态 / 输入校验加固

## 1. 阶段信息

- 阶段：10-P-4-3
- 日期：2026-06-02
- 范围：Auth 输入校验补齐、JWT_SECRET 启动警告、资源归属复查
- 是否修改代码：是
- 是否修改前端：否
- 是否修改后端：是
- 是否修改数据库：否
- 是否接 LLM：否
- 是否 git commit：是

## 2. 本次结论摘要

- email 格式校验已补齐：注册/登录使用 `net/mail.ParseAddress()` 验证格式，自动 TrimSpace
- password 校验已加固：TrimSpace + 最小 6 字符，已抽取为 `validatePassword()`
- JWT_SECRET 启动警告已添加：`Config.Validate()` 检测弱默认值（`change-me-in-local` 等）并打印 warn 日志
- DB_PASSWORD 默认值警告：检测到 `password` 默认值时打印 warn 日志
- user_id 来源统一：所有 handler 通过 `middleware.UserIDFromContext(ctx)` 获取
- 资源归属已复查确认：所有 resume 查询按 `user_id` 过滤（`GetCurrentByUserID`/`ListVersionsByUserID`/`GetByUserIDAndVersionNo`）
- Agent 结果写入均含 `user_id`，存储隔离正确
- 沿用 10-P-4-2 AppError 体系，前端兼容
- go build / go vet / go test 全部通过
- 已 commit

## 3. 执行前问题清查

| 问题 | 证据文件 | 影响 | 风险等级 |
| -- | ---- | -- | ---- |
| email 无格式校验 | `auth_service.go:67` 仅检查空字符串 | 非法邮箱可进入用户表 | P2 |
| password 无 TrimSpace | `auth_service.go:70-74` | 可注册纯空格密码 | P2 |
| JWT_SECRET 默认 `change-me-in-local` | `config.go:115` | 生产环境风险高 | P2 |
| DB_PASSWORD 默认 `password` | `config.go:111` | 弱密码风险 | P2 |
| user_id 来源 | 全部 handler 使用 `middleware.UserIDFromContext()` | 正确 | — |
| 资源归属 | 全部 resume 查询按 user_id 过滤 | 正确 | — |

## 4. Auth 加固内容

| 项 | 修改前 | 修改后 | 文件 | 说明 |
| -- | --- | --- | -- | -- |
| 注册 email 校验 | 仅 `== ""` | `net/mail.ParseAddress()` + TrimSpace | `auth_service.go` | 标准库，零依赖 |
| 注册 password 校验 | 内联 `== ""` + `len < 6` | 提取 `validatePassword()`，TrimSpace | `auth_service.go` | 逻辑不变，更规范 |
| 登录 email 校验 | 仅 `== ""` | `validateEmail()` | `auth_service.go` | 与注册一致 |
| 登录 password 校验 | 仅 `== ""` | TrimSpace + 空检查 | `auth_service.go` | 防止纯空格密码 |
| 注册 email Trim | 无 | `strings.TrimSpace(req.Email)` | `auth_service.go` | 防止首尾空格 |
| 登录 email Trim | 无 | `strings.TrimSpace(req.Email)` | `auth_service.go` | 同上 |

## 5. JWT / 配置加固内容

| 项 | 修改前 | 修改后 | 文件 | 风险 |
| -- | --- | --- | -- | -- |
| JWT_SECRET 默认值 | `change-me-in-local`，无警告 | 同上 + 启动时 `Config.Validate()` 检测弱值 warn | `config.go`, `main.go` | 已降低 |
| JWT_SECRET 为空 | 无检测 | `Validate()` 打印 WARNING | `config.go` | 已覆盖 |
| DB_PASSWORD 默认值 | `password`，无警告 | `Validate()` 检测 warn | `config.go` | 已降低 |
| .env.example JWT 注释 | 简单说明 | 增加 ⚠️ + `openssl rand -hex 32` 生成建议 | `.env.example` | 无 |

## 6. 用户态与资源归属复查

| 模块 / 接口 | user_id 来源 | 是否校验资源归属 | 当前状态 | 风险等级 |
| ------- | ---------- | -------- | ---- | ---- |
| `GET /api/resumes/current` | `middleware.UserIDFromContext` | `repository.GetCurrentByUserID(ctx, userID)` | ✅ 安全 | — |
| `GET /api/resumes/versions` | 同上 | `repository.ListVersionsByUserID(ctx, userID)` | ✅ 安全 | — |
| `GET /api/resumes/versions/{v}` | 同上 | `repository.GetByUserIDAndVersionNo(ctx, userID, v)` | ✅ 安全 | — |
| `GET /api/resumes/versions/compare` | 同上 | `resumeRepo.GetByUserIDAndVersionNo(ctx, userID, ...)` | ✅ 安全 | — |
| `POST /api/resumes` | 同上 | `repository.CreateVersion` 含 userID | ✅ 安全 | — |
| `POST /api/jobs/recommend` | 同上 | `resumeRepo.GetByUserIDAndVersionNo / GetCurrentByUserID` | ✅ 安全 | — |
| `POST /api/jobs/{id}/match` | 同上 | 同上 | ✅ 安全 | — |
| `POST /api/resumes/optimize` | 同上 | `optimizationRepo.Create` 含 userID | ✅ 安全 | — |
| `POST /api/auth/register` | 无（公开） | 无需 | ✅ 安全 | — |
| `POST /api/auth/login` | 无（公开） | 无需 | ✅ 安全 | — |

**复查结论：所有接口 user_id 来源统一，资源归属校验正确，无越权风险。**

## 7. 修改文件清单

| 文件 | 修改内容 | 是否业务代码 | 说明 |
| -- | ---- | ------ | -- |
| `server/internal/service/auth_service.go` | 新增 `validateEmail()`/`validatePassword()`；Register/Login 使用 | 否（校验层） | +43/-11 |
| `server/internal/config/config.go` | 新增 `Validate()` 方法 | 否（配置层） | +19 |
| `server/cmd/server/main.go` | `cfg.Validate()` 调用 | 否 | +1 |
| `.env.example` | JWT_SECRET 注释增强 | 否 | +3/-1 |

## 8. 测试补充

未新增测试文件。现有 service 测试（含 auth_service 相关测试）全部 PASS。email/password 校验逻辑为通用标准库调用，风险低。

## 9. 验证结果

| 命令 / 检查项 | 结果 | 说明 |
| ------- | -- | -- |
| git status 执行前 | 干净 | 零修改 |
| go build ./cmd/server | ✅ PASS | 编译通过 |
| go vet ./... | ✅ PASS | 零告警 |
| go test ./... | ✅ PASS | 9 packages PASS |
| git diff --stat | 4 files, +55 -11 | 仅 server/ + .env.example |
| git status 执行后 | 符合边界 | 无前端/数据库/Agent 修改 |

## 10. 未做事项

- 未修改前端
- 未修改数据库 schema
- 未修改 Jobs / Resumes / Agent 正常业务逻辑
- 未接 LLM
- 未做上线部署
- 未处理 Agent Provider 边界
- 未处理 lint / 前端工程问题

## 11. 当前风险与后续建议

- **JWT_SECRET 默认值仍存在**：`change-me-in-local` 仍为默认值，但现启动时有 warn 日志。本地 Demo 不影响使用，生产环境需设置强 secret
- **DB_PASSWORD 默认值仍存在**：`password` 仍为默认值，本地 Demo 可接受，生产需修改
- **email 校验使用 `net/mail.ParseAddress`**：覆盖标准格式，复杂场景（如 Unicode）够用
- **无资源归属越权**：全链路验证通过
- **后续建议**：Agent 模块（10-P-4-2 后续子阶段）、测试补齐

## 12. docs 更新清单

| 文件 | 新增 / 更新 | 目的 |
| -- | ------- | -- |
| `docs/mcp/项目上下文.md` | 更新 | 更新阶段状态 |
| `docs/mcp/10-P-4-3-Auth用户态输入校验加固.md` | 新增 | 本阶段 MCP 记录 |
| `docs/09-开发日志/2026-06-02-242-10-P-4-3-Auth加固.md` | 新增 | 开发日志 |

## 13. commit 信息

- 是否 commit：是
- commit message：`feat(server): harden auth validation and user context`
- commit hash：见下方

## 14. CodeBuddy 重启恢复摘要

10-P-4-3 完成 Auth/用户态/输入校验加固：email 格式校验（net/mail）、password TrimSpace+最小长度、JWT_SECRET 弱默认值启动 warn、Config.Validate() 新增。全量资源归属复查确认安全：所有 resume 查询按 user_id 过滤，Agent 结果存储含 user_id。纠错沿用 10-P-4-2 AppError 体系。go build/vet/test 全部通过。

下次继续前读取：`docs/mcp/项目上下文.md`、本文件、`docs/mcp/10-P-4-0-后端现状清查与专业化问题分级.md`。

## 15. 最终判断

1. **Auth 输入校验是否已加固？** **是。** email 格式 + password 规范校验。

2. **JWT_SECRET 风险是否已降低？** **是。** 启动时 warn 检测。

3. **user_id 来源是否清晰？** **是。** 统一通过 middleware.UserIDFromContext。

4. **核心资源归属是否已复查？** **是。** 全部接口确认隔离开正确。

5. **是否保持前端兼容？** **是。** 响应结构不变。

6. **go build / go vet / go test 是否通过？** **是。**

7. **是否修改前端？** **否。**

8. **是否修改数据库？** **否。**

9. **是否接入 LLM？** **否。**

10. **是否已 commit？** **是。**

11. **本阶段是否可以通过？** **是。**
