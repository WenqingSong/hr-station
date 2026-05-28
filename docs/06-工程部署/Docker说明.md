# Docker 说明

> 最后更新：2026-05-28（阶段 1-3 补充 MySQL 环境变量说明）

## 当前状态

待编写（docker-compose.yml 存在但内容为空，阶段 1-4 执行）

## 第一版规划：只稳定 MySQL

阶段 1 Docker Compose 第一版只包含 MySQL 服务。

### 原因

1. CNB 开发环境中，前端和后端更适合通过本地命令开发，便于热更新
2. MySQL 更适合通过 Docker 保持一致
3. 阶段 1 先降低复杂度
4. 阶段 2 后端基础设施再连接数据库
5. 阶段 7 交付部署时再考虑完整容器化 web/server

### docker-compose.yml 规划

```yaml
version: '3.8'

services:
  mysql:
    image: mysql:8.0
    container_name: offer-catcher-mysql
    restart: unless-stopped
    ports:
      - "${MYSQL_PORT:-3306}:3306"
    environment:
      MYSQL_ROOT_PASSWORD: ${MYSQL_ROOT_PASSWORD:-password}
      MYSQL_DATABASE: ${MYSQL_DATABASE:-offer_catcher}
      MYSQL_CHARSET: utf8mb4
      MYSQL_COLLATION: utf8mb4_unicode_ci
    volumes:
      - mysql_data:/var/lib/mysql
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  mysql_data:
```

### MySQL 环境变量

Docker Compose 中 MySQL 服务使用以下环境变量（来自 `.env`）：

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `MYSQL_ROOT_PASSWORD` | `password` | MySQL root 密码，应与 DB_PASSWORD 保持一致 |
| `MYSQL_DATABASE` | `offer_catcher` | MySQL 初始化数据库名，应与 DB_NAME 保持一致 |
| `MYSQL_PORT` | `3306` | MySQL 暴露到宿主机的端口 |

这些变量已在 `.env.example` 中定义，阶段 1-4 编写 `docker-compose.yml` 时将通过 `${VAR:-default}` 语法引用。

### MySQL 服务配置

| 项目 | 值 |
|------|------|
| 服务名 | mysql |
| 镜像 | mysql:8.0 |
| 容器名 | offer-catcher-mysql |
| 数据库 | offer_catcher |
| 用户名 | root |
| 密码 | 从 .env 读取 MYSQL_ROOT_PASSWORD，默认 password |
| 端口 | 3306:3306 |
| 数据卷 | mysql_data |
| 字符集 | utf8mb4 |
| 排序规则 | utf8mb4_unicode_ci |

## 数据卷保护

| 命令 | 是否删除数据卷 | 说明 |
|------|--------------|------|
| `make down` | ❌ 不删除 | 只停止容器，数据保留 |
| `make restart` | ❌ 不删除 | 只重启容器，数据保留 |
| `make reset` | ✅ 删除 | ⚠️ 危险操作，删除容器和数据卷 |

## 各阶段 Docker 扩展计划

| 阶段 | 服务 | 说明 |
|------|------|------|
| 阶段 1 | mysql | 第一版只稳定 MySQL |
| 阶段 2 | 无 | 后端连已有 MySQL |
| 阶段 3 | 无 | 业务接口 |
| 阶段 4 | 无 | Agent / 模型 |
| 阶段 7 | redis、web、server | 完整容器化部署 |

## 本地运行 vs Docker Compose 运行

| 组件 | 本地运行（推荐开发） | Docker Compose 运行（部署） |
|------|-------------------|--------------------------|
| 前端 | `make dev-web`（npm run dev） | web 容器 |
| 后端 | `make dev-server`（go run） | server 容器 |
| MySQL | `make up`（Docker） | mysql 容器 |

开发环境推荐：MySQL 用 Docker，前端和后端用本地命令。

## 关联文档

- [阶段1工程初始化方案.md](./阶段1工程初始化方案.md)
- [环境变量.md](./环境变量.md)
- [Makefile说明.md](./Makefile说明.md)
- [部署说明.md](./部署说明.md)
