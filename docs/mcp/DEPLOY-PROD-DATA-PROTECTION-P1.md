# DEPLOY-PROD-DATA-PROTECTION-P1 — 生产数据保护增强

> 文档负责人：开发环境自动生成  
> 生效日期：当前  
> 前置依赖：`DEPLOY-PROD-SAFETY-P0.md`（P0 安全基线）  
> 状态：已修复（P1 级别）

---

## 1. 背景

`DEPLOY-FIX-P0` 已解决生产部署命令规范、`.env.prod` 加载方式、安全检查脚本等基线问题。

但 P0 后仍存在 3 个生产数据保护风险：

| # | 风险 | 严重度 | 说明 |
|---|------|--------|------|
| 1 | compose project name 未固定 | P1 | 生产目录名变化 → 新 volume → 触发空库初始化 |
| 2 | 默认 prod compose 挂载 init.sql | P1 | 含 `DROP TABLE IF EXISTS`，volume 丢失时清空数据 |
| 3 | 无生产数据库备份脚本 | P1 | 部署前无法快速备份，出问题时无法回滚 |

---

## 2. 本次 P1 修改

| 修改项 | 文件 | 说明 |
|--------|------|------|
| 固定 project name | `docker-compose.prod.yml` | 新增 `name: ${COMPOSE_PROJECT_NAME:-wenqing-hr-demo}` |
| 移除 init.sql 挂载 | `docker-compose.prod.yml` | 默认 prod compose 不再自动挂载 destructive init.sql |
| 新增 bootstrap override | `docker-compose.prod.bootstrap.yml` | 隔离的全新空库初始化配置，默认不使用 |
| 新增 env 变量 | `.env.prod.example` | 新增 `COMPOSE_PROJECT_NAME=wenqing-hr-demo` |
| 新增备份脚本 | `scripts/deploy/prod-backup.sh` | 部署前执行 `mysqldump` → `backups/mysql/` |
| 增强安检脚本 | `scripts/deploy/prod-check.sh` | 检查 project name、禁止 init.sql、备份目录 |
| 新增 Makefile target | `Makefile` | `make prod-backup` |
| 新增备份目录 | `backups/.gitkeep` / `backups/mysql/.gitkeep` | 占位，确保目录结构存在 |
| 新增本文档 | `docs/mcp/DEPLOY-PROD-DATA-PROTECTION-P1.md` | 数据保护专项文档 |

---

## 3. Compose Project Name 风险

### 风险机制

Docker Compose 默认以**父目录名**作为 project name。named volume 的命名格式为：

```
<project_name>_<volume_name>
```

如果生产目录发生变化：

```text
原目录：/opt/wenqing-hr-demo
新目录：/opt/wenqing-hr-demo-new
```

会导致：

- 旧 volume：`wenqing-hr-demo_mysql_data`（已存在，有数据）
- 新 volume：`wenqing-hr-demo-new_mysql_data`（新建，空）

MySQL 看到空 volume → 视为首次初始化 → 如果 init.sql 存在则执行 → **DROP TABLE + CREATE TABLE 覆盖**。

### 修复方案

`docker-compose.prod.yml` 顶部新增固定 project name：

```yaml
name: ${COMPOSE_PROJECT_NAME:-wenqing-hr-demo}
```

- 默认值：`wenqing-hr-demo`
- 可通过 `.env.prod` 中 `COMPOSE_PROJECT_NAME` 覆盖
- 首次部署后**不得随意修改**

### 使用方式

```bash
# 所有生产 compose 命令必须显式指定 project name：
docker compose --project-name wenqing-hr-demo --env-file .env.prod -f docker-compose.prod.yml config
docker compose --project-name wenqing-hr-demo --env-file .env.prod -f docker-compose.prod.yml up -d --build
docker compose --project-name wenqing-hr-demo -f docker-compose.prod.yml ps
docker compose --project-name wenqing-hr-demo -f docker-compose.prod.yml down
```

---

## 4. init.sql 风险

### 风险描述

`server/sql/init.sql`（688 行）包含 **19 处 `DROP TABLE IF EXISTS`**，覆盖所有 15+ 张业务表。

虽然 MySQL Docker 镜像只在**空 volume 首次初始化**时执行 `/docker-entrypoint-initdb.d/` 下的脚本，但在以下场景仍存在风险：

| 场景 | 是否触发 | 风险 |
|------|---------|------|
| `docker compose down -v` 后 `up` | 是 | ❌ 数据丢失 |
| compose project name 变化 | 是（新 volume） | ❌ 数据不可见 + 空库初始化 |
| 手动 `docker volume rm` | 是 | ❌ 数据丢失 |
| 正常重启（已有 volume） | 否 | ✅ 安全 |

### 修复方案

**默认 `docker-compose.prod.yml` 已移除 init.sql 挂载。**

原来的一行：
```yaml
- ./server/sql/init.sql:/docker-entrypoint-initdb.d/init.sql:ro
```

已被移除并替换为注释说明。

### Bootstrap Override（仅限全新空库）

如果需要初始化全新空数据库，使用独立的 override 文件：

```bash
# 仅限全新空 MySQL volume，无任何业务数据时执行
docker compose --project-name wenqing-hr-demo --env-file .env.prod \
  -f docker-compose.prod.yml \
  -f docker-compose.prod.bootstrap.yml \
  up -d --build
```

**绝对禁止**对有数据的生产环境使用 bootstrap override。

---

## 5. 备份流程

### 部署前正确步骤

```bash
# 1. 安全检查（只读，不部署）
make prod-check
# 或
bash scripts/deploy/prod-check.sh

# 2. 数据库备份
make prod-backup
# 或
bash scripts/deploy/prod-backup.sh

# 3. 构建并部署
docker compose --project-name wenqing-hr-demo --env-file .env.prod -f docker-compose.prod.yml up -d --build

# 4. 验证
docker compose --project-name wenqing-hr-demo -f docker-compose.prod.yml ps
curl -f http://localhost/healthz
```

### 备份文件

- 路径：`backups/mysql/`
- 文件名：`offer_catcher_YYYYmmdd_HHMMSS.sql.gz`
- 内容：全量逻辑备份（含存储过程、触发器）
- 操作：`docker exec` 容器内执行 `mysqldump` → gzip 压缩
- 安全：不 source `.env.prod`，不打印密码

### 重要说明

- 备份脚本**只读数据库**，不执行任何写操作
- 备份脚本**不在本次执行**（当前是开发环境）
- 部署前建议始终执行备份

---

## 6. 禁止事项（数据保护铁律）

| # | 禁止操作 | 后果 | P0/P1 状态 |
|---|---------|------|-----------|
| 1 | `docker compose down -v` | 删除所有 volume 数据 | P0 ✅ |
| 2 | `docker volume rm <name>` | 删除指定 volume | P0 ✅ |
| 3 | `source .env.prod` | bash 语法错误 | P0 ✅ |
| 4 | 省略 `--env-file .env.prod` | MySQL 密码为空 | P0 ✅ |
| 5 | 省略 `--project-name` | 目录变化 → 新 volume | P1 ✅ |
| 6 | 对生产手动执行 `init.sql` | DROP TABLE 清空数据 | P0 ✅ |
| 7 | 对生产使用 bootstrap override | DROP TABLE 清空数据 | P1 ✅ |
| 8 | 修改 `COMPOSE_PROJECT_NAME` 后重新部署 | 新 volume → 数据不可见 | P1 ✅ |
| 9 | 无备份直接部署 | 出问题无法回滚 | P1 ✅ |
| 10 | `make reset` / `make clean-volume` / `make db-reset` | 删除 dev volume（可能在生产误执行） | P0 ✅ |

---

## 7. 工具索引

| 工具 | 路径 | 功能 |
|------|------|------|
| 生产 compose | `docker-compose.prod.yml` | 生产服务编排（已固定 project name，不含 init.sql） |
| Bootstrap override | `docker-compose.prod.bootstrap.yml` | 全新空库初始化（仅限无数据场景） |
| 生产 env 模板 | `.env.prod.example` | 含 COMPOSE_PROJECT_NAME 等所有必需变量 |
| 安检脚本 | `scripts/deploy/prod-check.sh` | 部署前只读检查（9 步） |
| 备份脚本 | `scripts/deploy/prod-backup.sh` | 部署前数据库备份 |
| 备份目录 | `backups/mysql/` | 备份文件存储 |
| P0 安全基线 | `docs/mcp/DEPLOY-PROD-SAFETY-P0.md` | 生产部署安全基线 |
| 本文档 | `docs/mcp/DEPLOY-PROD-DATA-PROTECTION-P1.md` | 生产数据保护增强 |

---

## 8. 本次未做事项

- ✅ 未连接生产服务器
- ✅ 未执行部署（up/down/build/restart）
- ✅ 未执行备份（prod-backup.sh 仅创建，未运行）
- ✅ 未修改数据库
- ✅ 未修改 `server/sql/init.sql`
- ✅ 未创建真实 `.env.prod`
- ✅ 未执行 Git 操作

---

## 9. 仍然存在的风险（P2+ 待处理）

| 风险 | 级别 | 后续任务 |
|------|------|---------|
| 无正式数据库迁移体系 | P2 | 引入 Goose / golang-migrate 替代全量 init.sql |
| 无正式 prod-deploy runbook | P2 | 一键安全部署脚本（backup → check → up → smoke test） |
| 生产现有 volume 名称需人工确认 | P2 | 上线前人工确认现有 compose project name 和 volume |
| 首次生产空库初始化需人工处理 | P2 | 使用 bootstrap override 前人工确认无数据 |
| 备份保留策略未定义 | P2 | 定义备份保留天数、自动清理旧备份 |
| 备份恢复脚本未实现 | P2 | `scripts/deploy/prod-restore.sh` |

---

## 10. 后续建议

```text
DEPLOY-FIX-P2：一键安全部署脚本
  - 新增 scripts/deploy/prod-deploy.sh
  - 自动按序执行：check → backup → up → smoke test
  - 每次部署前自动备份

DB-MIGRATION-P0：正式迁移体系
  - 引入 Goose 或 golang-migrate
  - 取代 init.sql 全量建表方式
  - 支持增量迁移 (up/down)

PROD-DEPLOY-RUNBOOK：正式上线手册
  - 整合 P0 + P1 + P2 所有内容
  - 面向运维团队的一键部署流程
```
