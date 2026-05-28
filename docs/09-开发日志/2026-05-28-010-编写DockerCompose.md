# 2026-05-28-010-编写DockerCompose

## 1. 本次任务

基于阶段 1 工程初始化方案，编写项目根目录下的 `docker-compose.yml`，用于启动本地开发所需的 MySQL 服务。

## 2. 修改范围

- 修改 `/workspace/docker-compose.yml`（从空内容重写为 MySQL 服务配置）
- 更新 `docs/` 目录下相关文档

## 3. 新增内容

| 文件 | 说明 |
|------|------|
| `/workspace/docker-compose.yml` | MySQL 8.0 服务 + healthcheck + mysql_data 数据卷 + offer_catcher_net 网络 |
| `docs/09-开发日志/2026-05-28-010-编写DockerCompose.md` | 本开发日志 |

## 4. 修改内容

| 文件 | 修改说明 |
|------|---------|
| `docs/06-工程部署/Docker说明.md` | 从规划更新为实际配置说明（11 章节：服务配置、环境变量来源、数据卷、健康检查、常用命令、CNB 恢复流程、不含前后端原因、扩展计划、风险注意） |
| `docs/06-工程部署/环境变量.md` | 补充 Docker Compose 变量引用方式（`${VAR:-default}` 语法）、Compose 中 DB_* 与 MYSQL_* 变量关系说明、安全注意事项增加 MySQL 密码和 docker-compose.yml 不写真实密码 |
| `docs/06-工程部署/Makefile说明.md` | 更新 Docker 命令说明（make up 启动 MySQL、make down 不删卷、make reset 删除数据卷），更新常见问题（make up 现在启动 MySQL、make down 不删数据、如何完全清理数据） |
| `docs/06-工程部署/阶段1工程初始化方案.md` | 标记 1-4 已完成 |
| `docs/项目状态.md` | 标记 1-4 已完成，更新 Docker/Makefile 状态表，增加 MySQL 服务配置说明 |
| `docs/下一步任务.md` | 标记 1-4 已完成 |
| `docs/变更记录.md` | 追加 1-4 变更记录 |

## 5. 删除内容

- 无

## 6. 关键决策

1. **第一版只稳定 MySQL**：不加入 web/server 服务，降低复杂度，便于开发调试
2. **使用 mysql:8.0 镜像**：MySQL 8.0 稳定可靠，支持 utf8mb4
3. **字符集 utf8mb4 + 时区 Asia/Shanghai**：通过 command 参数配置，确保中文和时区正确
4. **healthcheck 配置**：mysqladmin ping 检查，start_period 30s，retries 10，确保 MySQL 完全就绪才标记 healthy
5. **环境变量使用 ${VAR:-default} 语法**：兼容 .env 不存在的情况
6. **mysql_data 数据卷**：持久化 MySQL 数据，`make down` 不删除
7. **offer_catcher_net 网络**：bridge 网络，为后续扩展预留
8. **不使用 version 字段**：Docker Compose V2 不再需要 version，更简洁

## 7. 未修改内容

- 未修改 `Makefile`
- 未修改 `.env.example`
- 未创建真实 `.env`
- 未修改前端代码
- 未修改后端代码
- 未初始化 Go 后端工程
- 未安装任何依赖
- 未执行 `make up` 或 `docker compose up`

## 8. 当前风险

| 风险 | 说明 |
|------|------|
| MySQL 使用弱密码 | 本地开发用 `password`，生产环境必须替换 |
| 端口 3306 冲突 | 如果宿主机 3306 已被占用，需修改 MYSQL_PORT |
| docker-compose.yml 未实际启动验证 | 未执行 `make up`，需后续验证 |
| .env 不存在 | 需用户执行 `make init` 或手动创建，否则使用默认值 |

## 9. 如何验证

1. 检查 `docker-compose.yml` 是否存在且内容有效：`cat docker-compose.yml`
2. 检查 `docker-compose.yml` 语法：`docker compose config`（如环境有 Docker）
3. 检查只包含 MySQL 服务，不包含 web/server
4. 检查包含 `mysql_data` 数据卷
5. 检查包含 healthcheck
6. 检查环境变量使用 `${VAR:-default}` 语法
7. 检查不包含真实密钥
8. 后续执行 `make up` 验证 MySQL 可启动

## 10. 下一步建议

1. **1-5**：初始化 Go 后端工程目录
2. **1-6**：编写前后端运行说明
3. **验证**：在 1-5 完成后，执行 `make init && make up` 验证完整恢复流程
