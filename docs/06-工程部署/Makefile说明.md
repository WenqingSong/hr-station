# Makefile 说明

> 最后更新：2026-05-29（阶段 4-1.1.3 修复 Makefile 全量测试问题）

## 1. 设计目标

1. Makefile 作为统一命令入口，不直接写大量复杂 shell 逻辑
2. 具体执行逻辑下沉到 `scripts/dev/*.sh` 脚本
3. 云环境重启后快速恢复开发环境
4. 前端、后端、Docker、数据库、Agent 分开控制
5. 后续阶段持续扩展
6. 新模型 / 新开发者快速接手
7. 危险操作必须 `CONFIRM=yes` 确认
8. 不在脚本中打印真实密钥

## 2. 总体结构

```
Makefile                    # 统一入口 / 命令编排层
scripts/dev/common.sh       # 通用函数（log、load_env、mask_secret 等）
scripts/dev/doctor.sh       # 开发环境检查
scripts/dev/env.sh          # 环境初始化与配置展示
scripts/dev/infra.sh         # Docker Compose 基础设施管理
scripts/dev/mysql.sh        # 数据库导入/重置/shell/检查
scripts/dev/backend.sh      # 后端启动/停止/重启/日志/状态/健康检查
scripts/dev/web.sh          # 前端安装/启动/停止/重启/日志/状态/健康检查
scripts/dev/agent.sh        # Agent/LLM 环境检查
scripts/dev/logs.sh         # 统一日志查看
scripts/dev/quality.sh      # 代码质量检查
scripts/dev/clean.sh        # 清理操作
```

## 3. 云原生环境重启恢复流程

CNB 环境重启后，按以下顺序恢复：

```bash
make doctor     # 1. 检查环境
make init       # 2. 初始化（补齐 .env、创建目录）
make up         # 3. 一键启动完整开发环境
make health     # 4. 健康检查
```

如果前端依赖缺失：

```bash
make web-install
make web-up
```

如果数据库为空：

```bash
make db-import
```

如果想回到干净 Demo 数据：

```bash
make db-reset CONFIRM=yes
```

## 4. 常用命令

| 命令 | 说明 | 只读 | 有副作用 |
|------|------|------|---------|
| `make help` | 查看所有命令 | ✅ | ❌ |
| `make doctor` | 检查开发环境 | ✅ | ❌ |
| `make init` | 初始化开发环境 | ❌ | ⚠️ |
| `make check` | doctor + 健康检查 | ✅ | ❌ |
| `make status` | 查看所有服务运行状态 | ✅ | ❌ |
| `make health` | 查看所有服务健康状态 | ✅ | ❌ |

## 5. 生命周期命令

| 命令 | 说明 | 危险级别 |
|------|------|---------|
| `make up` | 一键启动完整开发环境 | 🟢 安全 |
| `make stop` | 停止前后端进程，保留 Docker 数据 | 🟢 安全 |
| `make down` | 停止所有服务，保留数据库 volume | 🟢 安全 |
| `make restart` | stop + up | 🟢 安全 |
| `make rebuild` | 清理运行态并重建 | 🔴 需 CONFIRM=yes |

### `make up` 执行顺序

```
1. 环境检查（doctor）
2. 环境初始化（init）
3. 启动基础设施（infra-up）
4. 检查数据库（db-check）
5. 启动后端（backend-up）
6. 启动前端（web-up）
7. 健康检查
```

### `make stop` vs `make down`

- `make stop`：只停止本地前后端进程，Docker 容器继续运行
- `make down`：停止前后端 + 停止 Docker 容器，但**不删除 volume**，数据保留

## 6. 基础设施命令

| 命令 | 说明 |
|------|------|
| `make infra-up` | 启动 Docker Compose 基础设施（MySQL） |
| `make infra-down` | 停止 Docker Compose 服务，保留 volume |
| `make infra-status` | 查看 Docker Compose 状态 |
| `make infra-logs` | 查看 Docker Compose 日志 |

## 7. 数据库命令

| 命令 | 说明 | 危险级别 |
|------|------|---------|
| `make db-import` | 导入 server/sql/init.sql | 🟢 可重复执行 |
| `make db-reset` | 重新导入初始化 SQL | 🔴 需 CONFIRM=yes |
| `make db-shell` | 进入 MySQL shell | 🟢 安全 |
| `make db-check` | 检查 6 张核心表记录数 | 🟢 只读 |

`db-check` 检查的 6 张核心表：users、resume_versions、jobs、recommendation_results、match_analyses、version_compare_results。

## 8. 后端命令

| 命令 | 说明 |
|------|------|
| `make backend-up` | 后台启动 Go 后端 |
| `make backend-stop` | 停止 Go 后端 |
| `make backend-restart` | 重启 Go 后端 |
| `make backend-logs` | 查看后端日志 |
| `make backend-status` | 查看后端进程状态 |
| `make backend-health` | curl /healthz |

后端以后台进程方式启动：
- PID 文件：`.run/server.pid`
- 日志文件：`.logs/server.log`
- 启动命令：`cd server && SERVER_PORT=$SERVER_PORT go run ./cmd/server`
- 不产生 `server/server` 编译产物
- 健康检查：`http://localhost:${SERVER_PORT}/healthz`

后端停止逻辑（4-1.1.3 修复）：
- 先终止子进程（`pgrep -P`），再终止主进程
- 兜底：如果端口仍有 go/server 相关残留进程，自动清理
- 停止后端口应完全释放，`backend-restart` 不再因端口残留失败

## 9. 前端命令

| 命令 | 说明 |
|------|------|
| `make web-install` | 在 web/ 下执行 npm install |
| `make web-up` | 后台启动前端开发服务 |
| `make web-stop` | 停止前端服务 |
| `make web-restart` | 重启前端服务 |
| `make web-logs` | 查看前端日志 |
| `make web-status` | 查看前端进程状态 |
| `make web-health` | 检查前端端口 |

前端以后台进程方式启动：
- PID 文件：`.run/web.pid`
- 日志文件：`.logs/web.log`
- `node_modules` 不存在时提示执行 `make web-install`
- 不自动安装依赖

## 10. Agent / LLM 命令

| 命令 | 说明 |
|------|------|
| `make agent-env` | 检查 LLM 相关环境变量（不打印真实 Key） |
| `make agent-check` | 检查 ENABLE_LLM_AGENT / LLM_PROVIDER / LLM_MODEL |
| `make agent-test` | 当前只提示 LLM Client 尚未实现 |

安全规则：
- 不打印真实 `LLM_API_KEY` 值，只显示「已配置/未配置」
- 如果 `LLM_MODEL=hhy3-preview`，提示疑似拼写错误
- `agent-test` 当前不调用 Hunyuan API，阶段 4-2 后启用

## 11. 日志命令

| 命令 | 说明 |
|------|------|
| `make logs` | 查看日志入口提示 |
| `make logs-backend` | tail .logs/server.log |
| `make logs-web` | tail .logs/web.log |
| `make logs-mysql` | docker compose logs -f mysql |
| `make logs-agent` | 当前查看后端日志中的 Agent 相关内容 |

日志目录 `.logs/` 由 `make init` 自动创建。

## 12. 质量检查命令

| 命令 | 说明 |
|------|------|
| `make lint` | Go gofmt 检查 + Shell 语法检查 |
| `make test` | cd server && go test ./... |
| `make build` | 前端 build + 后端 go build（输出到 /tmp） |
| `make ci` | doctor + lint + test + build |

`make build` 注意事项：
- 后端 go build 输出到 `/tmp/offer-catcher-server`，不残留 `server/server`
- 前端如果 `node_modules` 不存在，提示先执行 `make web-install`

## 13. 清理命令

| 命令 | 说明 | 危险级别 |
|------|------|---------|
| `make clean` | 清理运行文件和构建产物，不删除数据库 volume | 🟢 安全 |
| `make clean-logs` | 删除 .logs/ | 🟢 安全 |
| `make clean-pids` | 删除 .run/ | 🟢 安全 |
| `make clean-volume` | 删除 Docker volume | 🔴 需 CONFIRM=yes |

`make clean` 清理内容：
- `.run/*.pid`
- `.logs/*.log`
- `web/dist/`
- `server/bin/`
- `server/server`（残留二进制）

## 14. 危险操作确认

以下命令必须 `CONFIRM=yes`：

```bash
make rebuild CONFIRM=yes       # 清理运行态并重建
make db-reset CONFIRM=yes      # 重新导入初始化 SQL
make clean-volume CONFIRM=yes  # 删除 Docker volume
```

没有确认时输出提示并拒绝执行：

```text
危险操作：该命令会重置本地开发数据。
如确认执行，请使用：make xxx CONFIRM=yes
```

## 15. 常见问题

### Q: CNB 环境重启后如何恢复？

```bash
make doctor → make init → make up → make health
```

### Q: `make init` 会覆盖 .env 吗？

不会。`.env` 已存在时跳过，不会覆盖。

### Q: `make up` 启动失败怎么办？

1. 先执行 `make doctor` 检查环境
2. 如果 MySQL 未就绪，等待后重试：`make backend-up`
3. 如果前端依赖缺失：`make web-install && make web-up`
4. 如果数据库为空：`make db-import`

### Q: `make down` 会删除数据库数据吗？

不会。`make down` 只停止服务和容器，不删除 volume。

### Q: 如何完全清理数据库？

```bash
make clean-volume CONFIRM=yes
```

这将删除 `mysql_data` 数据卷，MySQL 中所有数据将丢失。

### Q: 后端日志在哪里？

```bash
make logs-backend    # 实时查看
cat .logs/server.log # 查看全部
```

### Q: 前端依赖缺失怎么办？

```bash
make web-install
```

### Q: `make agent-check` 显示 LLM_API_KEY 值吗？

不会。只显示「已配置/未配置」，不打印真实 Key。

## 关联文档

- [工程执行规范.md](./工程执行规范.md)
- [环境变量.md](./环境变量.md)
- [部署说明.md](./部署说明.md)
- [Docker说明.md](./Docker说明.md)
