# DEPLOY-PROD-SAFETY-P0 — 生产部署安全基线

> 文档负责人：开发环境自动生成  
> 生效日期：当前  
> 状态：已修复（P0 级别）

---

## 1. 背景

Offer 捕手项目的生产自动部署一直失败。经过 `DEPLOY-DIAG-DEV-1` 只读调查，确认根因集中在环境变量加载机制与 Docker Compose 的交互方式上。

### 调查结论摘要

| 根因 | 严重度 | 说明 |
|------|--------|------|
| Compose 默认只读 `.env`，不自动读 `.env.prod` | P0 | 生产无 `.env`，变量插值为空 |
| compose 文件注释给出的命令缺少 `--env-file .env.prod` | P0 | 误导 AI 和人工部署 |
| `.env.prod` 不能 `source`（含 shell 特殊字符） | P0 | 曾尝试 `source .env.prod` 失败 |
| 生产 compose 挂载 `init.sql`（含 DROP TABLE） | P1 | volume 丢失时有数据风险 |
| 无标准生产部署脚本 | P1 | 每次靠 AI 临时拼命令 |

---

## 2. 核心结论（部署安全铁律）

| # | 规则 | 原因 |
|---|------|------|
| 1 | **生产命令必须显式指定 `--env-file .env.prod`** | Docker Compose 默认只读取 `.env`，生产无此文件 |
| 2 | **禁止 `source .env.prod`** | `DB_DSN` 含 `tcp()` `?` `&` `=` 等 shell 特殊字符 |
| 3 | **禁止 `docker compose down -v`** | 会删除 `mysql_data` / `redis_data` volume，丢失所有数据 |
| 4 | **禁止删除生产 volume** | `docker volume rm` 同样致命 |
| 5 | **禁止手动执行 `init.sql`** | 含 `DROP TABLE IF EXISTS`，会清空所有业务表 |
| 6 | **禁止用 dev 命令操作生产** | `make reset` / `make clean-volume` / `make db-reset` 都会删数据 |

---

## 3. 正确命令

### 验证配置（不启动容器）

```bash
# 方式一：Makefile（推荐，不打印完整密钥）
make prod-config

# 方式二：直接运行（会打印完整 config，包含敏感值）
docker compose --env-file .env.prod -f docker-compose.prod.yml config

# 方式三：安全检查脚本（只读，不部署）
bash scripts/deploy/prod-check.sh
make prod-check
```

### 构建并启动

```bash
docker compose --env-file .env.prod -f docker-compose.prod.yml up -d --build
```

### 查看状态

```bash
docker compose -f docker-compose.prod.yml ps
```

### 查看日志

```bash
docker compose -f docker-compose.prod.yml logs -f
```

### 停止服务（保留数据）

```bash
docker compose -f docker-compose.prod.yml down
# 注意：不加 -v，保留 volume 数据
```

---

## 4. 错误命令（禁止执行）

| 错误命令 | 后果 | 为何错误 |
|---------|------|---------|
| `docker compose -f docker-compose.prod.yml up -d --build` | MySQL 密码为空，部署失败 | 缺少 `--env-file .env.prod` |
| `source .env.prod` | bash 语法错误 | `DB_DSN` 含特殊字符 |
| `source .env.prod && docker compose ...` | 先 source 失败，后续全部失效 | 同上 |
| `docker compose down -v` | **删除所有数据库数据** | 删除 named volume |
| `docker volume rm <prod_volume>` | **删除所有数据库数据** | 直接删除 volume |
| `make reset` | 执行 dev `docker compose down -v` | dev 命令用于生产 |
| `make clean-volume` | 执行 dev `docker compose down -v` | 同上 |
| `make db-reset` | 重新导入含 DROP TABLE 的 init.sql | 清空业务数据 |
| `docker exec ... mysql ... < server/sql/init.sql` | 手动执行初始化 SQL | 清空所有表 |

---

## 5. `.env.prod.example` 说明

- **位置**：项目根目录 `.env.prod.example`
- **用途**：生产环境变量模板，所有值使用 `CHANGE_ME` 占位符
- **不应包含**：任何真实密钥、密码、API Key
- **可以提交到 Git**：是（仅模板）
- **真实 `.env.prod`**：仅存在于生产服务器 `/opt/wenqing-hr-demo/.env.prod`，不应提交到 Git

### 从模板创建真实 `.env.prod`

```bash
cp .env.prod.example .env.prod
# 然后编辑 .env.prod，替换所有 CHANGE_ME 占位符为真实值
```

### 格式注意事项

- `DB_DSN` 必须用双引号包围（含 `tcp()` `?` `&` `=` 等特殊字符）
- 不要用 shell `source` 加载 `.env.prod`
- Docker Compose 的 `--env-file` 可以正确处理引号包围的值

---

## 6. init.sql 风险说明

### 当前状态

`docker-compose.prod.yml` 第 31 行挂载了 `server/sql/init.sql`：

```yaml
volumes:
  - ./server/sql/init.sql:/docker-entrypoint-initdb.d/init.sql:ro
```

`init.sql` 包含 **19 处 `DROP TABLE IF EXISTS`**，覆盖所有 15+ 张业务表。

### 何时触发

MySQL Docker 镜像仅在 **volume 首次初始化（数据目录为空）** 时执行 `/docker-entrypoint-initdb.d/` 下的脚本。

### 安全场景

| 场景 | 是否触发 init.sql | 安全 |
|------|------------------|------|
| 已有数据的 volume，重新 `up` | 否 | ✅ |
| 已有数据的 volume，重启容器 | 否 | ✅ |
| `docker compose down`（不带 `-v`）后 `up` | 否 | ✅ |
| 全新空 volume，首次 `up` | 是 | ⚠️ 正常初始化 |
| `docker compose down -v` 后 `up` | 是 | ❌ 数据丢失 |
| compose project name 变更 | 是（新 volume） | ❌ 数据不可见 |

### 防护措施

1. **绝对禁止 `docker compose down -v`**
2. **不要手动删除 `mysql_data` volume**
3. **生产 compose 命令必须固定 project name**（当前未固定，依赖默认 project name）
4. **P1 阶段应考虑移除 init.sql 挂载，改用显式迁移**

---

## 7. 部署前置检查工具

| 命令 | 功能 | 是否部署 |
|------|------|---------|
| `make prod-config` | 验证 compose config（输出脱敏） | ❌ 只读 |
| `make prod-check` | 完整安全校验（变量、flags、config） | ❌ 只读 |
| `bash scripts/deploy/prod-check.sh` | 同 `make prod-check`，脚本版 | ❌ 只读 |

### prod-check 检查项目

1. `.env.prod` 是否存在
2. `docker-compose.prod.yml` 是否存在
3. 必需变量是否设置：`MYSQL_ROOT_PASSWORD`、`DB_NAME`、`DB_DSN`、`JWT_SECRET`
4. LLM 配置是否完整：`LLM_ENABLE`、`LLM_PROVIDER`、`LLM_MODEL`、`LLM_API_KEY`
5. 生产 Feature 标志是否正确：`ENABLE_MOCK_AGENT=false`、`LLM_FALLBACK_ENABLE=false`
6. Compose config 能否通过
7. init.sql 挂载风险警告
8. 部署安全禁令提醒

---

## 8. 部署前检查清单（Deploy Checklist）

在每次生产部署前，确认以下所有项：

- [ ] 已将 `.env.prod.example` 复制为 `.env.prod` 并填入真实密钥
- [ ] 已确认 `.env.prod` 不使用 `source` 加载
- [ ] `make prod-check` 通过
- [ ] `docker compose --env-file .env.prod -f docker-compose.prod.yml config` 通过
- [ ] 已有最新数据库备份（`backups/db/`）
- [ ] 确认不会执行 `down -v`
- [ ] 确认 compose project name 未变更
- [ ] 部署命令记忆为：`docker compose --env-file .env.prod -f docker-compose.prod.yml up -d --build`

### 部署后验证

- [ ] `docker compose -f docker-compose.prod.yml ps` — 所有服务 running/healthy
- [ ] `curl -f http://localhost/healthz` — HTTP 200
- [ ] `curl -f http://localhost/api/healthz` — HTTP 200
- [ ] 浏览器访问首页正常
- [ ] 登录功能正常
- [ ] 数据库数据未丢失

---

## 9. AI 自动部署约束（AI Agent 必须遵守）

当 AI Agent 被要求执行生产部署时，必须：

1. **先执行 `make prod-check` 或 `scripts/deploy/prod-check.sh`**
2. **检查 .env.prod 存在且无占位符**
3. **确认备份已完成**
4. **命令必须显式包含 `--env-file .env.prod`**
5. **不得执行以下任何操作**：
   - `docker compose down -v`
   - `docker volume rm`
   - `source .env.prod`
   - `make reset`
   - `make clean-volume`
   - `make db-reset`
   - 直接执行 `server/sql/init.sql`
6. **如部署失败，先运行 prod-check 诊断，不要盲目重试**

---

## 10. 本次修复记录（P0 基线）

| 修改项 | 文件 | 说明 |
|--------|------|------|
| 修正 compose 注释 | `docker-compose.prod.yml` | 添加 `--env-file .env.prod`，补充安全禁令 |
| 新增 env 模板 | `.env.prod.example` | 占位符变量，DB_DSN 加引号 |
| 增强 prod-config | `Makefile` | 不打印完整 config，脱敏输出 |
| 增强 prod-check | `Makefile` | 新增 DB_DSN/DB_NAME 检查、placeholder 检测、init.sql 警告 |
| 新增安检脚本 | `scripts/deploy/prod-check.sh` | 独立可执行的部署前只读检查 |
| 新增安全准则 | `docs/mcp/DEPLOY-PROD-SAFETY-P0.md` | 本文档 |

### 本次未做事项

- 未连接生产服务器
- 未执行任何部署命令（up/down/build/restart）
- 未修改数据库
- 未创建真实 `.env.prod`
- 未移除 init.sql 挂载（P1 任务）
- 未执行 Git 操作

---

## 11. 仍然存在的风险（P1/P2 待处理）

| 风险 | 级别 | 后续任务 |
|------|------|---------|
| prod compose 仍挂载 init.sql | P1 | DEPLOY-FIX-P1：移除或改为条件初始化 |
| 没有自动备份脚本 | P1 | 新增 `scripts/deploy/prod-backup.sh` |
| 没有增量迁移体系 | P1 | 设计 migration 机制替代 init.sql |
| compose 未固定 project name | P1 | 添加 `COMPOSE_PROJECT_NAME` 或 `-p` 固定名称 |
| 生产 `.env.prod` 需人工创建 | P2 | 后续可在部署脚本中自动校验和创建 |
| 没有一键安全部署命令 | P2 | `make prod-deploy`（含 backup → up → smoke test） |
| 没有部署后自动 smoke test | P2 | 自动检查 healthz + 核心页面 |

---

## 12. 相关文件索引

| 文件 | 路径 |
|------|------|
| 生产 compose | `docker-compose.prod.yml` |
| 生产 env 模板 | `.env.prod.example` |
| 部署安检脚本 | `scripts/deploy/prod-check.sh` |
| 部署安全准则 | `docs/mcp/DEPLOY-PROD-SAFETY-P0.md` |
| 只读诊断报告 | `docs/mcp/DEPLOY-DIAG-DEV-1`（上一轮） |
| Makefile | `Makefile`（含 `prod-config` / `prod-check`） |
| 开发环境变量 | `docs/06-工程部署/环境变量.md` |
| Docker 说明 | `docs/06-工程部署/DockerCompose说明.md` |
