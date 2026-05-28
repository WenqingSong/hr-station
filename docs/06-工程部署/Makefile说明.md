# Makefile 说明

> 最后更新：2026-05-28（阶段 1-2 创建 Makefile）

## 1. 设计目标

1. 云环境重启后快速恢复开发环境
2. 本地开发命令统一入口
3. 前端、后端、Docker 服务分开控制
4. 后续阶段持续扩展
5. 新模型 / 新开发者快速接手

## 2. CNB 云原生开发环境重启恢复流程

CNB 环境重启后，按以下顺序恢复：

```bash
# 1. 检查环境（只读，不修改）
make doctor

# 2. 初始化（幂等，可重复执行）
make init

# 3. 启动 Docker 服务
make up

# 4. 启动前端开发（在新终端）
make dev-web

# 5. 启动后端开发（在新终端）
make dev-server
```

辅助命令：

```bash
make status      # 查看容器状态
make logs        # 查看服务日志
make logs-mysql  # 查看 MySQL 日志
make down        # 停止服务（不删卷）
make restart     # 重启服务
```

## 3. 命令总览

| 命令 | 目标 | 只读 | 有副作用 | 适合重启后 | 说明 |
|------|------|------|---------|-----------|------|
| `make help` | 查看所有命令 | ✅ | ❌ | ✅ | 列出所有 make 目标和说明 |
| `make doctor` | 只读检查开发环境 | ✅ | ❌ | ✅ | 检查 Go/Node/Docker/端口/依赖 |
| `make init` | 初始化开发环境 | ❌ | ⚠️ | ✅ | 复制 .env、提示安装依赖（幂等） |
| `make up` | 启动 Docker Compose 服务 | ❌ | ⚠️ | ✅ | 启动 MySQL |
| `make down` | 停止服务（不删卷） | ❌ | ⚠️ | ❌ | docker compose down |
| `make restart` | 重启服务 | ❌ | ⚠️ | ✅ | down + up |
| `make status` | 查看容器状态 | ✅ | ❌ | ✅ | docker compose ps |
| `make logs` | 查看全部服务日志 | ✅ | ❌ | ✅ | docker compose logs -f |
| `make logs-mysql` | 查看 MySQL 日志 | ✅ | ❌ | ✅ | docker compose logs -f mysql |
| `make dev-web` | 启动前端开发服务 | ❌ | ⚠️ | ✅ | cd web && npm run dev |
| `make dev-server` | 启动后端开发服务 | ❌ | ⚠️ | ✅ | cd server && go run ./cmd/server |
| `make dev` | 输出开发启动提示 | ✅ | ❌ | ✅ | 打印开发环境启动步骤 |
| `make build-web` | 构建前端 | ❌ | ⚠️ | ❌ | cd web && npm run build |
| `make build-server` | 构建后端 | ❌ | ⚠️ | ❌ | cd server && go build ./cmd/server |
| `make build` | 构建前后端 | ❌ | ⚠️ | ❌ | build-web + build-server |
| `make test-server` | 运行后端测试 | ✅ | ❌ | ✅ | cd server && go test ./... |
| `make check` | 执行基础检查 | ✅ | ❌ | ✅ | doctor + test-server |
| `make clean` | 清理构建产物 | ❌ | ⚠️ | ❌ | 删除 web/dist、server/bin |
| `make reset` | ⚠️ 危险：停止并清理容器卷 | ❌ | 🔴 | ❌ | docker compose down -v（删除数据） |

## 4. 环境检查命令

### `make help`

输出所有可用命令及说明，是最安全的入口命令。

### `make doctor`

只读检查开发环境，不做任何修改。检查项：

```
--- 工具检查 ---
✓ node 是否安装及版本
✓ npm 是否安装及版本
✓ go 是否安装及版本
✓ docker 是否安装及版本
✓ docker compose 是否可用

--- 项目文件检查 ---
✓ .env 是否存在
✓ .env.example 是否存在
✓ docker-compose.yml 是否存在
✓ web/package.json 是否存在
✓ web/node_modules 是否存在
✓ server/go.mod 是否存在

--- Docker 服务状态 ---
✓ 容器运行状态

--- 端口占用检查 ---
✓ 端口 3306 是否可用（MySQL）
✓ 端口 5173 是否可用（前端）
✓ 端口 8080 是否可用（后端）
```

## 5. 初始化命令

### `make init`

用于 CNB 环境重启后恢复基础开发状态。行为：

1. **检查 .env**：如果 `.env` 不存在且 `.env.example` 存在，则复制；如果 `.env.example` 也不存在，输出提示但不报错
2. **检查前端依赖**：如果 `web/node_modules` 不存在，提示安装命令，不自动执行
3. **检查后端依赖**：如果 `server/go.mod` 不存在，提示待阶段 1-5 初始化

**关键特性**：
- 幂等：重复执行不会破坏已有环境
- 不删除任何数据
- 不启动任何服务
- 不强制安装依赖（只提示）

## 6. Docker Compose 命令

### `make up`

启动 Docker Compose 服务。如果 `docker-compose.yml` 不存在或为空，输出清晰提示。

### `make down`

停止 Docker Compose 服务，**不删除数据卷**。数据保留在 `mysql_data` 卷中。

### `make restart`

先 down 再 up，不删除数据卷。

### `make status`

查看容器运行状态。

### `make logs` / `make logs-mysql`

查看服务日志，支持实时跟踪（-f）。

## 7. 前端开发命令

### `make dev-web`

启动前端开发服务。执行 `cd web && npm run dev`。

前置条件：
- `web/package.json` 存在
- `web/node_modules` 存在（如不存在，提示安装）

### 前端脚本说明

| 脚本 | 命令 | 说明 |
|------|------|------|
| `dev` | `vite` | 启动 Vite 开发服务器 |
| `build` | `vite build` | 构建生产版本 |

## 8. 后端开发命令

### `make dev-server`

启动后端开发服务。执行 `cd server && go run ./cmd/server`。

前置条件：
- `server/go.mod` 存在（如不存在，提示待阶段 1-5 初始化）

### `make dev`

不启动任何服务，只输出开发启动提示，说明需在不同终端分别运行 `make dev-web` 和 `make dev-server`。

## 9. 构建与测试命令

### `make build-web`

执行 `cd web && npm run build`，构建前端到 `web/dist/`。

### `make build-server`

执行 `cd server && go build -o bin/server ./cmd/server`，构建后端到 `server/bin/server`。

### `make build`

依次执行 `build-web` 和 `build-server`。

### `make test-server`

执行 `cd server && go test ./...`，运行后端所有测试。

### `make check`

先执行 `doctor` 环境检查，再尝试运行后端测试。

## 10. 清理与危险命令

### `make clean`

只清理构建产物（`web/dist`、`server/bin`），不删除数据库卷。

### `make reset` ⚠️ 危险

该命令用于清理所有 Docker 数据卷。**不会自动执行**，只输出提示：

```bash
docker compose down -v
```

用户需手动确认后执行。执行后 MySQL 中所有数据将被删除。

## 11. 常见问题

### Q: CNB 环境重启后如何恢复？

```bash
make doctor    # 检查环境状态
make init      # 补齐 .env 等基础文件
make up        # 启动 Docker 服务
make dev-web   # 启动前端（新终端）
make dev-server # 启动后端（新终端）
```

### Q: `make init` 如何恢复 `.env`？

`.env.example` 已在阶段 1-3 创建。执行 `make init` 时，如果 `.env` 不存在且 `.env.example` 存在，会自动复制 `.env.example` 为 `.env`。`.env` 不应提交到代码仓库，LLM API Key 等敏感信息需手动在 `.env` 中填写。

### Q: `make up` 提示 `docker-compose.yml` 为空？

`docker-compose.yml` 在阶段 1-4 编写。当前 MySQL 服务尚未配置。

### Q: `make dev-server` 提示 `server/go.mod` 不存在？

Go 后端在阶段 1-5 初始化。当前 server 目录为空。

### Q: 前端依赖丢失怎么办？

```bash
cd web && npm install
```

### Q: 如何避免误删数据库？

- `make down` 不删除数据卷，数据保留
- `make reset` 才会删除数据卷，但不会自动执行
- 只有手动运行 `docker compose down -v` 才会删除数据

## 12. 后续扩展

| 阶段 | 新增命令 | 说明 |
|------|---------|------|
| 阶段 2 | `make migrate` | 运行数据库迁移 |
| 阶段 2 | `make seed` | 填充种子数据 |
| 阶段 7 | `make deploy` | 部署命令 |
| 阶段 7 | `make demo` | 演示启动 |

## 关联文档

- [阶段1工程初始化方案.md](./阶段1工程初始化方案.md)
- [Docker说明.md](./Docker说明.md)
- [环境变量.md](./环境变量.md)
- [部署说明.md](./部署说明.md)
