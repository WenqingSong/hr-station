# Docker Compose 说明

> 最后更新：2026-05-30（阶段 4-3.7.B.R.3 Nginx 代理与 CORS 复核修正）

## 1. 概述

当前混合式开发环境中，Docker Compose **仅管理** MySQL 和 Nginx 两个服务。Backend / Frontend 不由 Compose 管理，而是由 Makefile 以本机后台进程方式启动。

## 2. 服务列表

| 服务 | 镜像 | 端口 | 说明 |
|------|------|------|------|
| MySQL | `mysql:8.0` | 3306 | 数据库服务，数据持久化到 `mysql_data` volume |
| Nginx | `nginx:alpine` | 80 | 统一反向代理，通过 `host.docker.internal` 代理到宿主机 |

## 3. 不包含的服务

| 服务 | 运行方式 | 说明 |
|------|---------|------|
| Backend | 本机后台进程 | `go run ./cmd/server`，PID 管理在 `.run/server.pid` |
| Frontend | 本机后台进程 | `npm run dev -- --host 0.0.0.0`，PID 管理在 `.run/web.pid` |

## 4. 配置文件

- `docker-compose.yml`：开发环境主配置（MySQL + Nginx）
- 环境变量从根目录 `.env` 读取

## 5. 常用命令

```bash
make infra-up      # 启动 Docker Compose 基础设施
make infra-down    # 停止 Docker Compose 服务（保留 volume）
make infra-status  # 查看 Docker Compose 状态
make infra-logs    # 查看 Docker Compose 日志
make ps            # 同 infra-status
```

## 6. Volume 管理

| Volume | 说明 | 持久化 |
|--------|------|--------|
| `mysql_data` | MySQL 数据 | ✅ `make down` 不删除，`make reset` 删除 |

## 7. 网络

所有服务连接到 `offer_catcher_net` bridge 网络。

## 8. 与方案 A 的区别

方案 A（全容器开发）的 `docker-compose.dev.yml` 包含 4 个服务（MySQL + Nginx + Backend + Frontend），当前混合式方案只有 2 个服务（MySQL + Nginx）。

方案 A 已通过 `backup/full-container-dev` 分支保留。

## 9. R.3 更新

4-3.7.B.R.3 已复核 Docker Compose 仍只管理 MySQL + Nginx。Nginx 通过 host.docker.internal 代理到宿主机 backend/frontend。后端 CORS 配置已通过环境变量配置化，不影响 Docker Compose 配置（后端不在容器中运行）。

## 10. R.4 更新

4-3.7.B.R.4 已完成混合式一键启动集成验收。Docker Compose 仍只管理 MySQL + Nginx，`make up` / `make down` 链路稳定。
