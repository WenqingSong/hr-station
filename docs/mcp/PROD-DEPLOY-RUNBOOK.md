# PROD-DEPLOY-RUNBOOK — 正式生产部署手册

> **文档版本**：v1.0（由开发环境 CodeBuddy 根据 P0/P1 改造生成）  
> **目标读者**：生产服务器运维人员 / 生产 AI Agent  
> **前置依赖**：`DEPLOY-PROD-SAFETY-P0.md` + `DEPLOY-PROD-DATA-PROTECTION-P1.md`  
> **状态**：已落盘，未在生产服务器执行

---

## 1. 文档定位

本 Runbook 是 Offer 捕手项目的**正式生产部署手册**。

- 它在开发环境 `/workspace` 生成，用于指导后续生产服务器部署。
- 它不是部署执行记录。
- 它不代表已经连接生产或完成上线。
- 所有命令均需在**生产服务器**上执行。
- 生产 AI Agent 只能按本手册执行，禁止自创命令。

---

## 2. 当前部署架构

```
用户浏览器
    │
    ▼
┌──────────────────────────┐
│  web (Nginx:80)          │  ← 暴露端口 80
│  - 静态文件托管           │
│  - /api/* → server:8080  │
│  - /healthz → server     │
│  - SPA fallback          │
└──────────┬───────────────┘
           │
           ▼
┌──────────────────────────┐
│  server (Go:8080)        │  ← 业务后端
│  - 多阶段构建             │
│  - golang:1.24 → alpine  │
│  - env_file: .env.prod   │
└────┬──────────┬──────────┘
     │          │
     ▼          ▼
┌────────┐  ┌─────────────┐
│ mysql  │  │ redis        │
│ :8.0   │  │ :7-alpine    │
│ 3306   │  │ 6379         │
└────────┘  └─────────────┘
```

| 组件 | 镜像/构建 | 端口 | 健康检查 |
|------|----------|------|---------|
| **web** | `web/Dockerfile` (Vite + Nginx) | **80:80** | `curl -f http://127.0.0.1:80/` |
| **server** | `server/Dockerfile` (Go 1.24 → Alpine) | 8080（内部） | `curl -f http://127.0.0.1:8080/healthz` |
| **mysql** | `mysql:8.0` | 3306（内部） | `mysqladmin ping` |
| **redis** | `redis:7-alpine` | 6379（内部） | `redis-cli ping` |

| 资源 | 名称 | 说明 |
|------|------|------|
| compose project | `wenqing-hr-demo` | **固定**，首次部署后不得修改 |
| MySQL volume | `wenqing-hr-demo_mysql_data` | named volume |
| Redis volume | `wenqing-hr-demo_redis_data` | named volume |
| 网络 | `offer-catcher-prod` | bridge |

| 关键文件 | 用途 |
|---------|------|
| `docker-compose.prod.yml` | 生产服务编排（**不含 init.sql**） |
| `docker-compose.prod.bootstrap.yml` | 全新空库初始化 override（**默认禁止使用**） |
| `.env.prod` | 生产环境变量（仅维护于生产服务器） |
| `.env.prod.example` | 生产 env 模板（占位符，可提交 Git） |
| `web/nginx.conf` | 生产 Nginx 配置（/api 反代 + SPA fallback） |
| `scripts/deploy/prod-check.sh` | 部署前安全检查（只读） |
| `scripts/deploy/prod-backup.sh` | 部署前数据库备份 |

---

## 3. 生产部署铁律

以下规则在**每次生产操作时**必须严格遵守，违反任何一条都可能导致数据丢失或部署失败。

| # | 铁律 | 原因 |
|---|------|------|
| 1 | **禁止 `source .env.prod`** | `DB_DSN` 含 `tcp()` `?` `&` `=` 等 shell 特殊字符 |
| 2 | **必须使用 `--env-file .env.prod`** | Docker Compose 默认只读 `.env`，生产无该文件 |
| 3 | **必须固定 `COMPOSE_PROJECT_NAME=wenqing-hr-demo`** | 目录名变化 → 新 project name → 新 volume → 数据不可见 |
| 4 | **禁止 `docker compose down -v`** | 删除所有 named volume，数据库数据永久丢失 |
| 5 | **禁止 `docker volume rm`** | 直接删除指定 volume |
| 6 | **禁止 `docker system prune -a --volumes`** | 清理所有未使用 volume |
| 7 | **禁止对生产执行 `server/sql/init.sql`** | 含 19 处 `DROP TABLE IF EXISTS` |
| 8 | **禁止对已有数据使用 `docker-compose.prod.bootstrap.yml`** | bootstrap 挂载 init.sql，会清空数据库 |
| 9 | **禁止无备份部署** | 出问题时无法回滚 |
| 10 | **禁止把开发环境 `.env` 上传到生产** | 开发密钥 ≠ 生产密钥 |
| 11 | **禁止打印真实密钥** | 日志/终端/slack 不得包含密码、Token、API Key |

---

## 4. 生产服务器目录假设

| 项 | 值 |
|----|-----|
| 生产项目目录 | `/opt/wenqing-hr-demo` |
| compose project name | `wenqing-hr-demo` |
| 生产 env 文件 | `/opt/wenqing-hr-demo/.env.prod` |
| 备份目录 | `/opt/wenqing-hr-demo/backups/mysql/` |

> **重要**：即使实际服务器目录名不同，也必须使用固定的 `COMPOSE_PROJECT_NAME=wenqing-hr-demo`。  
> 该名称一旦首次生产部署后，不得随意修改，否则会创建新的 named volume。

---

## 5. 首次进入生产服务器的只读确认

登录生产服务器后，**不要执行任何部署命令**。先做只读确认：

```bash
# 切换到项目目录
cd /opt/wenqing-hr-demo

# 确认当前路径
pwd

# 查看目录内容
ls -la

# 确认必需文件
ls -la .env.prod
ls -la docker-compose.prod.yml
ls -la scripts/deploy/prod-check.sh
ls -la scripts/deploy/prod-backup.sh

# 查看当前运行的 compose 项目
docker compose ls

# 查看与 wenqing-hr-demo 相关的 volume
docker volume ls | grep wenqing-hr-demo || echo "(no volumes yet)"

# 查看现有容器
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
```

> 这些命令全部只读，不启动、不停止、不修改任何服务或数据。

---

## 6. .env.prod 创建与检查

### 6.1 从模板创建

```bash
cp .env.prod.example .env.prod
nano .env.prod   # 或 vim
```

### 6.2 必须替换的变量

以下变量在模板中为 `CHANGE_ME` 占位符，**必须替换为真实值**：

```dotenv
COMPOSE_PROJECT_NAME=wenqing-hr-demo          # 固定，不要改

MYSQL_ROOT_PASSWORD="<真实强密码>"             # MySQL root 密码
DB_PASSWORD="<与上面一致>"                     # 后端连接密码
DB_DSN="root:<密码>@tcp(mysql:3306)/offer_catcher?charset=utf8mb4&parseTime=True&loc=Local"
                                               # DB_DSN 必须保留双引号！

JWT_SECRET="<64+ 字符随机密钥>"               # JWT 签名密钥

LLM_API_KEY="<真实 TokenHub API Key>"          # 大模型 API Key
LLM_HY3_API_KEY="<真实 Hy3 API Key>"           # 备用模型 API Key

CORS_ALLOWED_ORIGINS="https://<生产域名>"       # 生产前端域名

JWT_EXPIRE_HOURS=168                           # JWT 过期时间
```

### 6.3 规则

- `.env.prod` 只在生产服务器维护。
- **不得提交到 Git**（`.gitignore` 已排除 `.env.*`）。
- **不得从开发环境复制真实的 `.env`**。
- `DB_DSN` 必须保留双引号。
- **不要 `source .env.prod`**。
- 修改后运行 `scripts/deploy/prod-check.sh` 验证。

---

## 7. 生产前置检查

**每次部署前必须执行，不通过则禁止继续。**

```bash
# 方式一（推荐）
make prod-check

# 方式二
bash scripts/deploy/prod-check.sh
```

### prod-check 检查内容

| # | 检查项 | 不通过 |
|---|--------|--------|
| 1 | `.env.prod` 是否存在 | ERROR → exit 1 |
| 2 | `docker-compose.prod.yml` 是否存在 | ERROR → exit 1 |
| 3 | `COMPOSE_PROJECT_NAME` 是否稳定 | ERROR/WARN |
| 4 | 默认 compose 是否误挂载 `init.sql` | ERROR → exit 1 |
| 5 | 必需变量是否存在（MYSQL_ROOT_PASSWORD、DB_NAME、DB_DSN、JWT_SECRET、LLM_*） | ERROR/WARN |
| 6 | Feature 标志是否正确（ENABLE_MOCK_AGENT=false 等） | WARN |
| 7 | `.env.prod` source 安全性提醒 | INFO |
| 8 | compose config 是否可解析 | ERROR → exit 1 |
| 9 | `backups/mysql/` 目录是否存在 | WARN |

> **prod-check 不部署、不启动容器、不修改数据库。**  
> **prod-check 不通过 → 禁止继续上线。**

---

## 8. 上线前数据库备份

### 8.1 执行备份

```bash
# 方式一（推荐）
make prod-backup

# 方式二
bash scripts/deploy/prod-backup.sh
```

### 8.2 备份信息

| 项 | 说明 |
|----|------|
| 备份路径 | `backups/mysql/` |
| 文件格式 | `offer_catcher_YYYYmmdd_HHMMSS.sql.gz` |
| 备份内容 | 全量逻辑备份（`--single-transaction --routines --triggers`） |
| 备份方式 | `docker exec` 容器内 `mysqldump` → gzip 压缩 |
| 安全性 | 不 `source .env.prod`，不打印密码 |

### 8.3 两种场景

| 场景 | 处理 |
|------|------|
| **已有生产数据升级** | **必须先备份成功。** 备份失败 → 不得继续部署 |
| **首次全新空库部署** | 可能无可备份数据，备份失败允许跳过（需人工确认） |

---

## 9. 正式生产启动命令

### 唯一推荐命令

```bash
docker compose \
  --project-name wenqing-hr-demo \
  --env-file .env.prod \
  -f docker-compose.prod.yml \
  up -d --build
```

### 参数说明

| 参数 | 必要性 | 说明 |
|------|--------|------|
| `--project-name wenqing-hr-demo` | **必须** | 固定 project name，防止目录变化创建新 volume |
| `--env-file .env.prod` | **必须** | 加载生产环境变量（Docker Compose 不会自动读 `.env.prod`） |
| `-f docker-compose.prod.yml` | **必须** | 指定生产 compose |
| `up -d --build` | 部署 | 后台启动 + 构建镜像 |

### 禁止事项

- ❌ 不使用 `docker-compose.prod.bootstrap.yml`（会挂载 init.sql）
- ❌ 不 `source .env.prod` 后再启动
- ❌ 不使用 `down -v`
- ❌ 不带 `--env-file` 直接启动

---

## 10. 启动后状态检查

### 容器状态

```bash
docker compose \
  --project-name wenqing-hr-demo \
  --env-file .env.prod \
  -f docker-compose.prod.yml \
  ps
```

期望输出：所有 4 个 service 均为 `Up` 或 `healthy`。

### 日志检查

```bash
# server 日志（最近 100 行）
docker compose \
  --project-name wenqing-hr-demo \
  --env-file .env.prod \
  -f docker-compose.prod.yml \
  logs --tail=100 server

# web 日志（最近 100 行）
docker compose \
  --project-name wenqing-hr-demo \
  --env-file .env.prod \
  -f docker-compose.prod.yml \
  logs --tail=100 web
```

### 关键日志检查点

- server 日志应包含：
  - `MySQL connected` — 数据库连接成功
  - `Redis connected` — Redis 连接成功（若失败仅 warning，不 fatal）
  - `LLM pipeline initialized` — LLM 初始化成功
  - `Starting on :8080 (env=production, logLevel=...)` — 服务启动
- web 日志无异常错误

> **警告**：日志中如果出现密钥，不能外传。

---

## 11. Smoke Test

### 本地检查（生产服务器内）

```bash
# 健康检查
curl -i http://127.0.0.1/healthz
# 期望: HTTP 200

# 首页
curl -i http://127.0.0.1/
# 期望: HTTP 200 (返回 HTML)

# API 连通性
curl -i http://127.0.0.1/api/jobs
# 期望: HTTP 200 (返回 JSON)
```

### 外部域名检查（如果已配置）

```bash
curl -i https://YOUR_DOMAIN/healthz
curl -i https://YOUR_DOMAIN/
curl -i https://YOUR_DOMAIN/api/jobs
```

### 人工页面验收（浏览器）

| 页面 | 检查项 | 核心功能 |
|------|--------|---------|
| 首页 `/` | 能打开，岗位列表正常加载 | 岗位列表展示 |
| 登录页 | 能注册/登录，登录后状态保持 | 用户认证 |
| 我的简历 `/my-resumes` | 能创建、查看、编辑简历 | 简历管理 |
| 岗位推荐 | 选择简历后能生成推荐结果 | **四核心之一** ✅ |
| 岗位匹配分析 | 选择简历+岗位后能生成匹配分析 | **四核心之一** ✅ |
| 简历优化建议 | 选择简历+岗位后能生成优化建议 | **四核心之一** ✅ |
| 简历版本对比 | 选择两个版本能生成对比报告 | **四核心之一** ✅ |

> **四核心功能必须至少各做一次 smoke test。**

---

## 12. 旧数据与旧 result_json 说明

项目历史可能存在以下表中的生成结果：

- `resume_optimization_results` — 简历优化结果
- `match_analyses` — 匹配分析结果
- `recommendation_results` — 推荐结果
- `resume_compare_results` — 版本对比结果

### 部署规则

1. 部署不会自动清理旧结果，数据全部保留。
2. 之前生成的旧 `result_json` 可能仍然保留旧版本文案。
3. 修复后的逻辑只在**重新生成**后体现在新结果里。
4. **不要直接 `DELETE` / `UPDATE` 生产结果表。**
5. 如需清理旧结果，必须另开数据治理任务，不在部署流程中处理。

---

## 13. 失败处理

### 13.1 compose 变量为空

**现象**：
```
WARN[0000] The "MYSQL_ROOT_PASSWORD" variable is not set
WARN[0000] The "DB_NAME" variable is not set
```

**处理**：
1. 确认命令是否带有 `--env-file .env.prod`
2. 确认 `.env.prod` 是否存在
3. 确认变量名是否拼写正确
4. 运行 `scripts/deploy/prod-check.sh` 定位缺失变量

### 13.2 source .env.prod 报错

**现象**：bash 语法错误、`unexpected token`、`(` 解析失败等

**处理**：
- **不要 source .env.prod**
- 使用 `--env-file .env.prod` 参数
- 原因：`DB_DSN` 含 `tcp()` `?` `&` `=` 等 shell 特殊字符

### 13.3 MySQL 连接失败

**现象**：server 日志 `Failed to initialize database` 或 `Failed to ping database`

**处理**：
1. 检查 `.env.prod` 中 `DB_DSN` 是否使用 `mysql:3306`（非 `127.0.0.1`）
2. 检查 `MYSQL_ROOT_PASSWORD` 与 `DB_PASSWORD` 是否一致
3. 检查 `DB_NAME` 是否为 `offer_catcher`
4. 检查 mysql 容器是否 healthy：`docker compose --project-name wenqing-hr-demo -f docker-compose.prod.yml ps`
5. 检查 mysql 日志：`docker compose ... logs mysql`

### 13.4 前端页面打开但 API 失败

**现象**：首页能打开，但岗位列表为空或 API 请求 502/504

**处理**：
1. 确认 `web/nginx.conf` 是否代理 `/api` 到 `server:8080`
2. 确认 server 容器是否 healthy
3. 直接请求 server 验证：`curl http://127.0.0.1:8080/api/jobs`（需要从宿主机或 web 容器内执行）
4. 检查 server 日志中的数据库连接状态

### 13.5 端口冲突

```bash
# 检查 80 端口占用
ss -lntp | grep ':80'
```

**处理**：
- 先确认占用来源（可能是旧的 nginx 或其他服务）
- **禁止乱杀进程**，先确认后再人工处理
- 非本项目占用时需人工评估

### 13.6 容器启动但反复重启

```bash
# 查看容器详情
docker ps -a --format "table {{.Names}}\t{{.Status}}"

# 查看退出容器日志
docker logs <container_name> --tail=50
```

常见原因：
- server：数据库连接失败 → 检查 DB_DSN
- web：依赖 server healthy，server 未就绪则持续等待
- mysql：密码为空 → 检查 --env-file

---

## 14. 回滚方案

### 原则

1. **优先回滚镜像/代码版本，不回滚数据库。**
2. **禁止使用 `down -v` 回滚。**
3. 如果只是新版本容器异常，重新部署上一版本代码。
4. 如果涉及数据异常，先保留现场并使用备份进行人工恢复评估。
5. 当前已有备份脚本但**没有 restore 脚本**，恢复必须人工谨慎执行。
6. **数据库恢复属于高风险操作，不能由 AI 自动执行。**

### 应用层回滚

```bash
# 1. 切回上一版本代码（Git checkout 或重新拉取旧版）
git checkout <上一版本commit>

# 2. 重新执行安全检查
bash scripts/deploy/prod-check.sh

# 3. 重新部署
docker compose \
  --project-name wenqing-hr-demo \
  --env-file .env.prod \
  -f docker-compose.prod.yml \
  up -d --build

# 4. Smoke test 验证
curl -i http://127.0.0.1/healthz
```

### 绝对禁止的回滚操作

| 禁止操作 | 原因 |
|---------|------|
| `docker compose down -v` | 删除所有 volume 数据 |
| `docker volume rm` | 直接删除 volume |
| 重新执行 `server/sql/init.sql` | 清空所有业务表 |
| 使用 `docker-compose.prod.bootstrap.yml` | 挂载 init.sql，触发 DROP TABLE |

---

## 15. 全新空库初始化说明

### 默认行为

默认生产部署**不使用** `docker-compose.prod.bootstrap.yml`，因此不会自动初始化数据库。

### Bootstrap 使用条件（全部满足才允许）

- [ ] 确认为**全新生产环境**（从未有过数据）
- [ ] 确认**没有任何历史数据**
- [ ] 确认**没有已有 mysql volume**
- [ ] 用户**明确接受**初始化 seed 数据
- [ ] 已**人工确认** `init.sql` 含 `DROP TABLE IF EXISTS` 风险

### Bootstrap 命令（仅限以上条件全部满足）

```bash
docker compose \
  --project-name wenqing-hr-demo \
  --env-file .env.prod \
  -f docker-compose.prod.yml \
  -f docker-compose.prod.bootstrap.yml \
  up -d --build
```

> ⚠️ **已有生产数据时绝对禁止使用此命令。**  
> ⚠️ **此命令会执行 init.sql 中的 DROP TABLE + CREATE TABLE。**

---

## 16. 标准部署流程总览

```
─────────────────────────────────────────────────────────────
                   生产部署标准流程
─────────────────────────────────────────────────────────────

  Step 1  登录生产服务器
          ssh user@production-server

  Step 2  进入项目目录
          cd /opt/wenqing-hr-demo

  Step 3  确认代码和文件（只读）
          git log -1 --oneline
          ls -la .env.prod docker-compose.prod.yml

  Step 4  创建/检查 .env.prod
          cp .env.prod.example .env.prod   # 首次
          nano .env.prod                    # 填入真实密钥

  Step 5  生产前置检查
          bash scripts/deploy/prod-check.sh
          ▸ 检查不通过 → 修复 → 重新检查 → 不通过禁止部署

  Step 6  数据库备份
          bash scripts/deploy/prod-backup.sh
          ▸ 已有数据 → 备份必须成功
          ▸ 异常 → 停止，人工排查

  Step 7  构建并启动
          docker compose --project-name wenqing-hr-demo \
            --env-file .env.prod \
            -f docker-compose.prod.yml \
            up -d --build

  Step 8  容器状态检查
          docker compose --project-name wenqing-hr-demo \
            --env-file .env.prod \
            -f docker-compose.prod.yml ps
          ▸ 期望: 4 services up / healthy

  Step 9  日志检查
          docker compose ... logs --tail=100 server
          docker compose ... logs --tail=100 web
          ▸ 确认 MySQL connected / Redis connected / LLM initialized

  Step 10 Smoke Test
          curl -i http://127.0.0.1/healthz   → HTTP 200
          curl -i http://127.0.0.1/           → HTTP 200
          curl -i http://127.0.0.1/api/jobs   → HTTP 200

  Step 11 人工页面验收
          浏览器打开: http://YOUR_DOMAIN/
          ▸ 首页 / 登录 / 简历管理 / 岗位推荐 / 匹配分析 / 优化 / 对比

  Step 12 记录结果
          记录: 部署时间 / 操作用户 / 前置检查结果 / 备份路径 / 验收结果
─────────────────────────────────────────────────────────────
```

---

## 17. 给生产服务器 AI 的执行边界

当 AI Agent 被要求在**生产服务器**上部署时，必须遵守以下边界：

### 允许执行

- `bash scripts/deploy/prod-check.sh` — 只读安全检查
- `bash scripts/deploy/prod-backup.sh` — 数据库备份（只读）
- `make prod-config` — 验证 compose 配置
- `docker compose --project-name ... --env-file ... up -d --build` — 启动部署
- `docker compose ... ps` / `docker compose ... logs` — 状态和日志查看
- `curl` 健康检查和 API 验证

### 绝对禁止

- ❌ 自创部署命令（必须完全按本手册）
- ❌ `source .env.prod`
- ❌ `docker compose down -v`
- ❌ `docker volume rm`
- ❌ 修改数据库（`DELETE`、`UPDATE`、`ALTER`、`DROP`、`TRUNCATE`）
- ❌ 执行 `server/sql/init.sql`
- ❌ 使用 `docker-compose.prod.bootstrap.yml`（除非全新空库且人工确认）
- ❌ 修改 `.env.prod` 中的 `COMPOSE_PROJECT_NAME`
- ❌ 打印或外传敏感信息

### 遇到不确定时

- **立即停止操作**
- **返回当前日志和状态**
- **等待人工决策**
- **不得猜测或尝试执行不确定的命令**

---

## 18. 本文档生成说明

| 项 | 状态 |
|----|------|
| 生成环境 | 开发环境 `/workspace`，CodeBuddy AI Agent |
| 生成时间 | 2026-06-14 |
| 前置依赖 | `DEPLOY-FIX-P0`（部署安全基线）+ `DEPLOY-FIX-P1`（数据保护增强） |
| 是否连接生产 | ❌ 否 |
| 是否执行部署 | ❌ 否 |
| 是否执行备份 | ❌ 否 |
| 是否修改数据库 | ❌ 否 |
| 是否执行 Git | ❌ 否 |
| 当前状态 | 文档已生成，待生产服务器执行验证 |

---

## 附录：快速参考卡片

### 正确命令速查

```bash
# 安全检查（只读）
bash scripts/deploy/prod-check.sh

# 数据库备份
bash scripts/deploy/prod-backup.sh

# 构建并启动
docker compose --project-name wenqing-hr-demo --env-file .env.prod -f docker-compose.prod.yml up -d --build

# 查看状态
docker compose --project-name wenqing-hr-demo -f docker-compose.prod.yml ps

# Smoke test
curl -i http://127.0.0.1/healthz
```

### 禁止命令速查

```bash
source .env.prod                                    # ❌
docker compose -f docker-compose.prod.yml up -d      # ❌ (缺少 --env-file)
docker compose down -v                               # ❌
docker volume rm <name>                              # ❌
docker exec ... mysql ... < server/sql/init.sql       # ❌
make reset / clean-volume / db-reset                 # ❌ (仅限 dev)
```

### 关键文件速查

| 文件 | 用途 |
|------|------|
| `docker-compose.prod.yml` | 生产部署主文件 |
| `.env.prod` | 生产密钥（仅生产服务器） |
| `scripts/deploy/prod-check.sh` | 部署前安全检查 |
| `scripts/deploy/prod-backup.sh` | 数据库备份 |
| `docs/mcp/DEPLOY-PROD-SAFETY-P0.md` | 安全基线文档 |
| `docs/mcp/DEPLOY-PROD-DATA-PROTECTION-P1.md` | 数据保护文档 |
| `docs/mcp/PROD-DEPLOY-RUNBOOK.md` | 本文档 |
