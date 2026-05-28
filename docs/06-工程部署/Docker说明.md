# Docker 说明

> 最后更新：2026-05-28（阶段 2-6 增强健康检查接口）

## 1. 当前 Docker Compose 目标

阶段 1 Docker Compose 第一版只包含 MySQL 服务，不包含前端和后端。

核心目标：

1. 为本地开发提供稳定的 MySQL 数据库
2. 使用 Docker volume 持久化 MySQL 数据
3. 配合 `make up / make down` 命令管理服务生命周期
4. 适配 CNB 云原生开发环境重启后的快速恢复

## 2. 当前服务列表

| 服务 | 镜像 | 端口 | 状态 |
|------|------|------|------|
| mysql | mysql:8.0 | 3306:3306 | ✅ 已配置 |

不包含 web / server 服务。

## 3. MySQL 服务配置

| 项目 | 值 | 说明 |
|------|------|------|
| 服务名 | `mysql` | Docker Compose 内部服务名 |
| 镜像 | `mysql:8.0` | MySQL 8.0 官方镜像 |
| 容器名 | `offer-catcher-mysql` | 便于识别 |
| 数据库 | `offer_catcher` | 初始化数据库名 |
| 用户名 | `root` | 默认 root 用户 |
| 密码 | 从 .env 读取 MYSQL_ROOT_PASSWORD，默认 `password` | 本地开发示例值 |
| 端口 | 3306:3306 | 宿主机:容器 |
| 字符集 | utf8mb4 | 通过 command 参数配置 |
| 排序规则 | utf8mb4_unicode_ci | 通过 command 参数配置 |
| 时区 | Asia/Shanghai (+08:00) | 容器时区和 MySQL 时区 |
| 数据卷 | `mysql_data` | 持久化到 Docker volume |
| 网络 | `offer_catcher_net` | bridge 网络 |

### MySQL command 参数

```yaml
command:
  - --character-set-server=utf8mb4
  - --collation-server=utf8mb4_unicode_ci
  - --default-time-zone=+08:00
```

这些参数确保 MySQL 8.0 使用 utf8mb4 字符集和正确的时区。

## 4. 环境变量来源

Docker Compose 默认读取项目根目录 `.env` 文件中的环境变量。

| 变量 | 默认值 | 来源 | 说明 |
|------|--------|------|------|
| `MYSQL_ROOT_PASSWORD` | `password` | .env | MySQL root 密码，应与 DB_PASSWORD 保持一致 |
| `MYSQL_DATABASE` | `offer_catcher` | .env | MySQL 初始化数据库名，应与 DB_NAME 保持一致 |
| `MYSQL_PORT` | `3306` | .env | MySQL 暴露到宿主机的端口 |

### 变量引用方式

`docker-compose.yml` 中使用 `${VAR:-default}` 语法：

- 如果 `.env` 存在且包含变量，使用 `.env` 中的值
- 如果 `.env` 不存在或变量未定义，使用默认值

### .env 生成方式

1. `.env.example` 是模板，已提交到仓库
2. `.env` 由 `make init` 从 `.env.example` 复制生成
3. `.env` 不应提交到仓库

## 5. 数据卷说明

| 卷名 | 挂载点 | 用途 |
|------|--------|------|
| `mysql_data` | `/var/lib/mysql` | MySQL 数据持久化 |

### 数据卷保护规则

| 命令 | 是否删除数据卷 | 说明 |
|------|--------------|------|
| `make down` | ❌ 不删除 | 只停止容器，数据保留在 mysql_data 卷中 |
| `make restart` | ❌ 不删除 | 只重启容器，数据保留 |
| `make reset` | ✅ 删除 | ⚠️ 危险操作，执行 `docker compose down -v`，删除所有数据卷 |

### 如何确认数据卷存在

```bash
docker volume ls | grep mysql_data
```

## 6. 健康检查说明

MySQL 服务配置了 healthcheck：

```yaml
healthcheck:
  test: ["CMD-SHELL", "mysqladmin ping -h localhost -uroot -p$${MYSQL_ROOT_PASSWORD} || exit 1"]
  interval: 10s
  timeout: 5s
  retries: 10
  start_period: 30s
```

| 参数 | 值 | 说明 |
|------|------|------|
| test | mysqladmin ping | 通过 ping 命令检查 MySQL 是否就绪 |
| interval | 10s | 每 10 秒检查一次 |
| timeout | 5s | 单次检查超时 5 秒 |
| retries | 10 | 连续 10 次失败才标记为 unhealthy |
| start_period | 30s | 容器启动后 30 秒内不计入重试次数 |

### 查看健康状态

```bash
make status
# 或
docker compose ps
```

## 7. 常用命令

| 命令 | 说明 |
|------|------|
| `make up` | 启动 MySQL 服务（后台运行） |
| `make down` | 停止 MySQL 服务（不删卷） |
| `make restart` | 重启 MySQL 服务 |
| `make status` | 查看容器运行状态 |
| `make logs` | 查看全部服务日志 |
| `make logs-mysql` | 查看 MySQL 日志 |

### 直接使用 docker compose 命令

```bash
docker compose up -d              # 启动
docker compose down               # 停止（不删卷）
docker compose ps                 # 查看状态
docker compose logs -f mysql      # 查看 MySQL 日志
docker compose exec mysql bash    # 进入 MySQL 容器
```

### 连接 MySQL

```bash
# 本地通过 mysql 客户端连接
mysql -h 127.0.0.1 -P 3306 -u root -ppassword offer_catcher

# 通过 Docker 容器内连接
docker compose exec mysql mysql -uroot -ppassword offer_catcher
```

## 8. CNB 环境重启恢复流程

CNB 云原生开发环境重启后，按以下步骤恢复：

```bash
# 1. 检查环境
make doctor

# 2. 初始化（补齐 .env）
make init

# 3. 启动 MySQL 服务
make up

# 4. 确认 MySQL 就绪
make status
```

说明：

1. `make doctor` 检查 `.env`、`.env.example`、`docker-compose.yml` 是否存在
2. `make init` 在 `.env` 不存在时从 `.env.example` 复制生成
3. `make up` 启动 MySQL，数据卷 `mysql_data` 保留之前的数据
4. MySQL 有 healthcheck，启动后约 30 秒内变为 healthy

### 查看 MySQL 状态与日志

```bash
make status       # 查看容器状态（healthy 表示就绪）
make logs-mysql   # 查看 MySQL 日志（实时跟踪）
```

### 避免误删数据卷

1. `make down` 只停止容器，不删除数据卷
2. `make restart` 只重启容器，不删除数据卷
3. `make reset` 会删除数据卷，但不会自动执行（需手动运行 `docker compose down -v`）
4. 日常操作使用 `make down` / `make up` 即可，数据不会丢失

## 9. 不包含前端和后端的原因

第一版 Docker Compose 不包含 web / server 服务，原因：

1. **开发体验**：前端本地 `npm run dev` 有热更新，后端本地 `go run` 有自动重载
2. MySQL 不仅用于后续业务，也已用于 `/healthz` 数据库状态检查
3. `make up` 后可通过 `/healthz` 间接验证数据库连接
2. **调试便利**：本地运行便于断点调试、日志查看
3. **降低复杂度**：阶段 1 先稳定 MySQL，避免同时管理多个服务
4. **阶段规划**：阶段 7 交付部署时再考虑完整容器化

### 本地运行 vs Docker 运行

| 组件 | 本地运行（推荐开发） | Docker 运行（部署） |
|------|-------------------|-------------------|
| 前端 | `make dev-web`（npm run dev） | 阶段 7 web 容器 |
| 后端 | `make dev-server`（go run） | 阶段 7 server 容器 |
| MySQL | `make up`（Docker） | mysql 容器 |

### 本地 Go 后端连接 MySQL

阶段 2-5 后，Go 后端启动时会连接 MySQL。本地开发使用 `127.0.0.1`：

```env
DB_HOST=127.0.0.1
DB_PORT=3306
```

如果 MySQL 未启动，后端启动会失败（这是预期行为）。

### Docker Compose 内 Go 后端连接 MySQL

如果未来后端也放入 Docker Compose，应改为：

```env
DB_HOST=mysql
DB_PORT=3306
DB_DSN=root:password@tcp(mysql:3306)/offer_catcher?charset=utf8mb4&parseTime=True&loc=Local
```

## 10. 后续扩展计划

| 阶段 | 新增服务 | 说明 |
|------|---------|------|
| 阶段 1 | mysql | ✅ 已完成 |
| 阶段 2 | 无 | ✅ 后端已连接 MySQL（2-5 完成），健康检查已增强（2-6 完成） |
| 阶段 3 | 无 | 业务接口 |
| 阶段 4 | 无 | Agent / 模型 |
| 阶段 7 | redis、web、server | 完整容器化部署 |

## 11. 风险与注意事项

| 风险 | 说明 |
|------|------|
| MySQL 密码使用弱密码 | 本地开发用 `password`，生产环境必须替换 |
| `make reset` 会删除数据 | 危险命令，会执行 `docker compose down -v`，删除所有数据卷 |
| 端口冲突 | 如果宿主机 3306 已被占用，需修改 `MYSQL_PORT` |
| 容器未正常退出 | 使用 `make restart` 或 `docker compose restart` 恢复 |
| 数据卷占用磁盘 | 长期开发后数据卷可能变大，可用 `docker system df` 查看 |

## 关联文档

- [阶段1工程初始化方案.md](./阶段1工程初始化方案.md)
- [环境变量.md](./环境变量.md)
- [Makefile说明.md](./Makefile说明.md)
- [部署说明.md](./部署说明.md)
