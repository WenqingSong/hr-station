# Offer捕手｜学生求职匹配与简历优化智能体

> 面向学生的「投递前决策系统」：把一份简历和一批岗位，转化为可解释的匹配结论、可执行的优化建议、可追踪的版本对比。

---

## 给 LLM / AI 编码助手（先读这个）

**如果你是大模型或 AI 编码助手，在阅读源码之前，请先完整阅读 [`docs/LLM_PROJECT_CONTEXT.md`](docs/LLM_PROJECT_CONTEXT.md)。**

该文档是专门为 LLM 编写的项目上下文手册，包含：

- 系统全景、模块边界与请求链路
- 完整 HTTP 路由表、请求/响应契约、业务错误码
- 25 张数据表的职责划分与关键字段
- Fast / Quality 双档 AI 调用链与逐级降级顺序
- Pipeline / Prompt / Schema / Repair / Fallback 的硬性约束
- 前端路由、数据适配层与契约陷阱
- 已知不一致、风险清单与改动后的验证清单

**本项目的高风险改动点**：AI 输出契约（`generated_by` / `agent_status` / `generation_tier`）、`/api` 路径前缀、前端 `/offer/` base path、prompt 与 schema 的文件名映射、`resumes` 关系型数据模型。修改这些内容前，请先阅读上述文档对应章节。

---

## 1. 项目定位

Offer捕手是一个面向学生求职场景的智能体应用。它不是岗位信息的二次聚合，而是帮助学生回答投递前的四个问题：

1. 适合投哪些岗位，为什么；
2. 简历与 JD 之间的证据链是否成立；
3. 现在最该改简历的哪几处；
4. 改到什么程度可以投，什么时候应该先补信息。

产品定位为：**面向学生的投递前决策与简历初筛优化系统**。官方招聘平台承担岗位入口与投递流程，本项目承担投递前分析、JD 证据链、简历优化建议与决策辅助。

当前项目包含一个完整的可运行闭环：岗位浏览 → 简历编辑 → 岗位推荐 → 匹配分析 → 简历优化 → 版本对比，并保留 Mock 规则引擎作为无 LLM 环境下的可演示降级路径。

## 2. 核心能力

| 能力 | 入口 | Fast 档（低延迟、规则驱动） | Quality 档（LLM 深度分析） |
|------|------|---------------------------|--------------------------|
| 岗位推荐 | `POST /api/jobs/recommend` | SEP 语义画像排序 + 解释器 | Deep Cascade（L1–L4）→ 3-Agent → LLM → Mock |
| 岗位匹配分析 | `POST /api/jobs/{id}/match` | SEP Fast 匹配引擎 | Hy3 Screening 5 模块 → 3-Agent → LLM → Mock |
| 简历优化 | `POST /api/resumes/{id}/optimization` | TagEvidence 规则引擎 | 3-Agent → LLM V1 → Mock V2 |
| 简历版本对比 | `POST /api/resumes/compare` | TagEvidence 差异分析 | 3-Agent → LLM V1 → Mock |
| 简历画像 | `GET /api/resumes/{id}/profile/status` | 规则画像（同步） | LLM 画像（异步生成） |
| 岗位语义画像（SEP） | 内嵌于推荐 / 匹配 | 标签 + 证据 | Hy3 生成，落 `semantic_profiles` |

关键设计：

- **双档分层**：`generation_tier` 只有 `fast` / `quality` 两个正式值，`deep` 是 `quality` 的兼容别名。
- **逐级降级**：每一档都有明确的降级顺序，最终回落 Mock 规则引擎。
- **Mock 不是空桩**：`match_mock.go`、`recommendation_mock.go`、`version_compare_mock.go` 合计约 180 KB，是可独立跑通全链路的规则引擎。
- **统一元信息**：所有 AI 响应都带 `meta`（`generated_by` / `agent_status` / `fallback_reason` / `repair_applied` / `schema_valid`），前端据此判断结果可信度。

## 3. 技术栈

| 层 | 技术 |
|----|------|
| 后端 | Go 1.24.1，标准库 `net/http` + `database/sql`，零 Web 框架、零 ORM |
| 前端 | React 18.3.1 + TypeScript + Vite 6 + Tailwind CSS v4 + shadcn/ui |
| 路由 | react-router 7（`createBrowserRouter`，base `/offer/`） |
| 数据库 | MySQL 8.0（25 张表，utf8mb4） |
| 会话 | Redis 7（JWT + 单会话校验，支持异地登录顶号） |
| 反向代理 | Nginx（开发容器；生产由 web 镜像内置 Nginx） |
| LLM 接入 | OpenAI 兼容协议；支持 `tokenhub` / `hunyuan` / `hy3` / `deepseek` |
| 结构化输出 | JSON Schema 校验（santhosh-tekuri/jsonschema v5）+ JSON Repair Agent |
| 工程化 | Makefile + `scripts/dev/*.sh` + Docker Compose + CNB CI |
| 后端依赖 | 以标准库为主，外部仅 MySQL driver、JWT、JSON Schema、bcrypt、godotenv、go-redis |

## 4. 系统架构

### 4.1 开发环境（混合式）

```
浏览器
  │
  ├── http://localhost/offer/ ──► Nginx :80 ──► Vite Dev Server :5173（前端）
  │                                   │
  └── http://localhost/api/... ───────┴──────► Go Backend :8080
                                                   │
                                    ┌──────────────┼──────────────┐
                                    ▼              ▼              ▼
                               MySQL :3306     Redis :6379    LLM Provider
```

- Docker Compose 管理 **MySQL + Redis + Nginx**；
- Backend / Frontend 由 Makefile 以宿主机后台进程启动；
- Nginx 通过 `host.docker.internal` 代理到宿主机服务；
- 前端 dev server 与生产构建均使用 `base: '/offer/'`，因此应用入口是 `/offer/`，不是 `/`。

### 4.2 AI 调用链

```
HTTP Handler
  → Service（tier 选择、上下文构建、结果落库、降级编排）
    → Business Agent（prompt 值组装、结果映射、fallback 判定）
      → pipeline.Runner
          Prompt Render → LLM Chat → JSON 抽取解析 → Schema 校验
          → 失败：JSONRepairAgent 修复一次 → 再解析 + 再校验
          → 仍失败：返回 fallback_required
    → 逐级降级：Deep / Screening → 3-Agent → 单 Agent LLM → Mock 规则引擎
```

统一约束：

- Repair **只允许一次**，不允许递归修复（`LLM_REPAIR_MAX_ATTEMPTS` 强制为 1）；
- 任何 LLM 失败都必须映射为明确的 `fallback_reason` + `error_code`；
- 原始模型输出默认不落日志、不返回前端，避免简历 / JD 泄露；
- 每个 Agent 的 prompt 与 schema 都由名称映射表加载，缺失即视为配置错误。

## 5. 仓库结构

```
.
├── README.md                       # 本文件
├── docs/
│   └── LLM_PROJECT_CONTEXT.md      # 专门给 LLM 的项目上下文手册（必读）
├── Makefile                        # 统一命令入口（help 可查看全部）
├── docker-compose.yml              # 开发环境：MySQL + Redis + Nginx
├── docker-compose.prod.yml         # 生产环境：MySQL + Redis + Server + Web
├── docker-compose.prod.bootstrap.yml # 生产空库初始化 override
├── nginx/                          # 开发环境 Nginx 配置
├── scripts/
│   ├── dev/                        # 开发脚本（17 个）
│   ├── deploy/                     # 部署前检查与备份
│   └── lib.sh                      # CI 依赖等待库
├── server/                         # Go 后端
│   ├── cmd/server/                 # 主入口（依赖装配）
│   ├── cmd/import_jobs/            # 岗位 JSON 入库
│   ├── cmd/generate_jobs_seed/     # 生成 SQL seed
│   ├── cmd/llm-smoke/              # LLM 连通性冒烟
│   ├── cmd/sep-backfill/           # 语义画像补齐
│   ├── internal/
│   │   ├── config/                 # 唯一配置加载点
│   │   ├── http/                   # 路由 + 7 个 Handler
│   │   ├── middleware/             # Recover / CORS / Logger / Auth
│   │   ├── repository/             # 11 个 MySQL Repository
│   │   ├── service/                # 业务编排 + SEP + Deep Cascade
│   │   ├── agent/                  # 业务 Agent + Mock + pipeline + repair
│   │   ├── llm/                    # Provider 抽象（hunyuan / hy3 / noop）
│   │   ├── prompt/ schema/ llmjson/ # Prompt、Schema、JSON 抽取
│   │   └── model/                  # 领域模型
│   ├── prompts/                    # 14 个 Prompt 模板 + 13 个 JSON Schema
│   └── sql/                        # init.sql（25 表）+ migrations + seed
├── web/                            # React 前端
│   ├── src/app/                    # 页面、布局、组件、路由
│   ├── src/api/                    # API 客户端 + ViewModel Loader
│   ├── src/adapters/               # 后端 DTO → 前端 ViewModel
│   ├── src/types/                  # 后端契约类型
│   └── src/stores/                 # authStore（唯一全局状态）
├── agents/                         # 独立 GoFrame 骨架服务（IAM 落地点）
├── resources/                      # 岗位主数据 + 增量 SQL 存档
├── deploy/dumps/                   # 生产全库 dump + 恢复手册
├── backups/                        # 历史备份产物（只读，不参与构建）
└── working/                        # 架构分析 / 故障复盘 / 重构设计（历史记录）
```

> `working/`、`backups/`、`server/sql/init.sql.bak` 属于历史产物；`server/README.md` 与 `web/README.md` 已过时，不要作为事实来源。

## 6. 快速启动

### 6.1 前置条件

| 依赖 | 版本建议 |
|------|---------|
| Docker + Docker Compose v2 | 必需 |
| Go | 1.24+ |
| Node.js / npm | Node 22+ / npm 10+ |
| make、curl、lsof | 必需 |

### 6.2 启动

```bash
# 1. 生成 .env（如果不存在；make up 也会自动执行）
cp .env.example .env

# 2. 按需填写本地敏感变量
#    LLM_API_KEY / JWT_SECRET / DB_PASSWORD

# 3. 环境体检（只读，不启动服务）
make doctor

# 4. 一键启动（基础设施 + 后端 + 前端）
make up

# 5. 查看状态
make status
```

`make up` 的实际步骤：确保 `.env` 存在 → 启动 Compose（MySQL / Redis / Nginx）→ 等待 MySQL 就绪 → 按需导入 `server/sql/init.sql` → 应用岗位 seed → 启动 Go 后端 → 健康检查 → 安装前端依赖并启动 Vite → 终检。

## 7. 访问地址

| 入口 | 地址 | 说明 |
|------|------|------|
| 应用（Nginx） | http://localhost/offer/ | 推荐入口，`/api` 同源代理，无 CORS |
| 应用（Vite 直连） | http://localhost:5173/offer/ | 前端 hot reload |
| API（Nginx） | http://localhost/api/ | 不 strip `/api` 前缀 |
| 健康检查 | http://localhost/healthz | 服务 + 数据库状态 |
| 后端直连 | http://localhost:8080 | 调试用，需要 CORS |
| Swagger UI | http://localhost:8080/swagger/ | 仅覆盖部分接口 |

> 前端以 `/offer/` 为 base path，`http://localhost/` 不是应用入口。`make status` 打印的 `http://localhost/` 为历史遗留展示，实际请使用 `/offer/`。

## 8. 常用命令

`make help` 是完整命令清单，以下为高频子集。

### 生命周期

```bash
make up              # 启动完整开发环境
make down            # 停止服务，保留 volume
make restart         # 软重启，不清 volume
make status          # 查看所有服务状态
make health          # 逐项健康检查，失败返回非 0
make stop            # 只停本地前后端，不动 Docker
```

### 日志

```bash
make logs-backend    # Go 后端日志
make logs-web        # Vite 日志
make logs-mysql      # MySQL 日志
make logs-agent      # 过滤 agent / llm 相关日志
make nginx-logs      # Nginx 日志
```

### 质量

```bash
make lint            # gofmt + shell 语法检查
make test            # go test ./...
make build           # 前端构建 + 后端编译
make ci              # doctor + lint + test + build
```

### 数据

```bash
make db-import           # 导入 server/sql/init.sql
make db-check            # 检查核心表与记录数
make db-shell            # 进入 MySQL shell
make jobs-import-preview # 预览岗位导入
make jobs-import         # 导入 resources/jobs 岗位数据
make jobs-seed-apply     # 应用岗位 SQL seed
```

### LLM

```bash
make llm-check                        # 只校验配置，不请求外部模型
RUN_REAL_LLM_SMOKE=1 make llm-smoke   # 显式开启真实连通性测试
make agent-env                        # 查看脱敏后的 LLM 配置
```

### 破坏性命令（必须显式确认）

```bash
make reset CONFIRM=yes        # 清空开发 volume（含 MySQL 数据）
make reset-up CONFIRM=yes     # 清空 volume 后重启
make clean-volume CONFIRM=yes # 删除 Docker volume，之后还需输入 YES
```

## 9. 配置与环境变量

### 9.1 文件规则

| 文件 | 用途 | 是否提交 |
|------|------|---------|
| `.env.example` | 开发变量模板 | 提交，只放占位符 |
| `.env` | 本地运行时配置 | 不提交 |
| `.env.local` | 本地真实配置 | 不提交 |
| `.env.prod.example` | 生产变量模板 | 提交，敏感值为 `CHANGE_ME` |
| `.env.prod` | 生产真实配置 | 不提交 |

`ensure-env` 的优先级：`.env` 已存在 → `.env.local` → `.env.example`。复制模板后会对缺失的双模型变量做幂等补齐（`LLM_FAST_MODEL` / `LLM_QUALITY_MODEL` / `LLM_DEFAULT_TIER`）。

### 9.2 关键变量

| 变量 | 代码默认值 | 说明 |
|------|-----------|------|
| `APP_ENV` | `local` | 运行环境标识 |
| `SERVER_PORT` | `8080` | 后端端口 |
| `FRONTEND_URL` | `http://localhost:5173` | 后端反代前端的地址 |
| `DB_DSN` | `root:password@tcp(127.0.0.1:3306)/offer_catcher?...` | 后端实际使用 DSN（优先于 DB_HOST 等） |
| `REDIS_HOST` / `REDIS_PORT` | `127.0.0.1` / `6379` | 会话存储 |
| `REDIS_SESSION_TTL` | `604800` | 会话有效期（秒） |
| `JWT_SECRET` | `change-me-in-local` | JWT 签名密钥，生产必须替换 |
| `JWT_EXPIRE_HOURS` | `168` | Token 有效期 |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:5173,http://127.0.0.1:5173` | 逗号分隔，禁止生产使用 `*` |
| `LLM_ENABLE` | `false` | 是否启用真实 LLM |
| `LLM_PROVIDER` | `hunyuan` | `tokenhub` / `hunyuan` / `hy3` / `deepseek` |
| `LLM_FAST_MODEL` / `LLM_QUALITY_MODEL` | 空 | 为空时回落 `LLM_MODEL` |
| `LLM_DEFAULT_TIER` | `fast` | 未显式指定 tier 时的默认档 |
| `LLM_FALLBACK_ENABLE` | `true` | LLM 失败是否允许降级 |
| `ENABLE_MOCK_AGENT` | `true` | 是否允许最终 Mock 兜底 |
| `LLM_REPAIR_MAX_ATTEMPTS` | `1` | 必须为 1，禁止递归修复 |
| `LLM_PROMPT_DIR` | `server/prompts` | Prompt 模板目录 |
| `LLM_SCHEMA_DIR` | `server/prompts/schema` | JSON Schema 目录 |
| `LOG_LEVEL` | `debug` | 日志级别 |

> `.env.example` 与代码默认值存在差异（例如模板中 `LLM_ENABLE=true`、`ENABLE_MOCK_AGENT=false`）。以实际 `.env` 为准；不确定时执行 `make agent-env` 查看脱敏后的生效配置。

### 9.3 安全规则

- `LLM_API_KEY`、`JWT_SECRET`、`DB_PASSWORD`、`MYSQL_ROOT_PASSWORD` 只能写入本地 `.env` / `.env.prod`；
- 生产环境通过环境变量或 Secret 注入，不使用 `.env` 文件；
- 不要把真实 Key 写入 `.env.example`、README、日志、Git；
- `prod-check.sh` 会校验必填变量、占位符、危险 compose 配置。

## 10. LLM 与 Agent 体系（速览）

| 组件 | 位置 | 职责 |
|------|------|------|
| `pipeline.Runner` | `server/internal/agent/pipeline` | 固定 9 步：校验 → 渲染 Prompt → 载入 Schema → 调 LLM → 解析 JSON → 校验 → 修复一次 → 复校验 → 返回状态 |
| `repair.Agent` | `server/internal/agent/repair` | 单次 JSON 修复，使用 `json_repair_v1.md` |
| `llm.Client` | `server/internal/llm` | Provider 抽象，`hunyuan` / `hy3` / `noop` |
| `prompt.Loader` | `server/internal/prompt` | Prompt 名称 → 文件映射 + `{{KEY}}` 渲染 |
| `schema.Loader/Validator` | `server/internal/schema` | Schema 名称 → 文件映射 + JSON Schema 校验 |
| `SemanticProfileGate` | `server/internal/service` | 简历 / 岗位语义画像的生成、缓存、失效与异步去重 |

双档语义：

| 档位 | 触发 | 典型耗时 | 是否调用 LLM |
|------|------|---------|-------------|
| `fast` | 默认 / 首屏 | 毫秒级到秒级 | 通常不调用，规则 + 标签 + 证据 |
| `quality` | 用户显式请求深度分析 | 数秒到数十秒 | 调用，按能力走 Deep 级联或 3-Agent |

输出契约：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "result": {},
    "meta": {
      "agent_name": "match_analysis",
      "generated_by": "llm_agent",
      "agent_status": "success",
      "fallback_reason": "none",
      "error_code": "none",
      "repair_applied": false,
      "schema_valid": true,
      "request_id": ""
    }
  }
}
```

前端判定规则（`web/src/app/utils/agentError.ts`）：

- `mock_agent` 与 `mock_quality` 一律视为 Mock 结果并报错；
- `fallback = true` 一律视为降级结果；
- `quality` 档只接受 `llm_agent` / `deep_fallback_agent`；
- `fast` 档不强制特定 `generated_by`，实际常见值为 `fast_tag_engine` / `fast_sep_match_engine`。

## 11. 数据与岗位数据

- `server/sql/init.sql`：25 张表的 DDL + 基础 seed，是开发环境唯一初始化入口；
- `server/sql/migrations/`：增量迁移（结果表 agent 元字段、`job_profiles` 等）；
- `server/sql/seed/tencent_jobs.sql`：由 `resources/jobs/tencent-campus-jobs.json` 生成的岗位 seed；
- `resources/jobs/tencent-campus-jobs.json`：人工整理的腾讯校招 / 实习岗位主数据源；
- `scripts/dev/tencent_jobs.sh`：低频拉取腾讯招聘公开接口原始 JSON；
- `scripts/dev/tencent_jobs_import.sh`：清洗 raw JSON → 生成 `INSERT ... ON DUPLICATE KEY UPDATE` → 导入。

核心表分组：

| 域 | 表 |
|----|----|
| 用户 | `users` |
| 岗位 | `jobs`、`job_locations`、`job_business_groups`、`job_skills`、`job_profiles` |
| 筛选项 | `job_filter_groups`、`job_filter_options`、`job_filter_option_relations` |
| 简历 | `resumes` + 7 张子表（教育 / 项目 / 实习 / 语言 / 证书 / 获奖 / 作品） |
| AI 结果 | `match_analyses`、`recommendation_results`、`resume_optimization_results`、`resume_compare_results`、`resume_profiles` |
| SEP | `semantic_profiles`、`semantic_profile_tags` |
| 运维 | `seed_versions` |

## 12. 测试与质量

```bash
make test    # cd server && go test ./...（32 个测试文件）
make lint    # gofmt -l . + bash -n
make build   # web: npm run build；server: go build ./cmd/server
make ci      # doctor + lint + test + build
```

测试重点覆盖：LLM 客户端与错误映射、Pipeline 状态机、JSON 抽取与修复、Schema 校验、Prompt 加载、四类业务 Agent、Match / Recommendation / Optimization 服务编排。

测试缺口：`repository`、`http`、`middleware`、`session`、`config`、`database`、`health` 目前没有单元测试；`cmd/*` 没有测试。

## 13. 生产部署

### 13.1 前置检查

```bash
cp .env.prod.example .env.prod        # 填写所有 CHANGE_ME
make prod-check                       # 只读安全检查，必须 PASSED
make prod-config                      # 校验 compose 配置，不启动
make prod-backup                      # 导出生产库到 backups/mysql/
```

### 13.2 标准发布

```bash
docker compose --project-name wenqing-hr-demo --env-file .env.prod \
  -f docker-compose.prod.yml up -d --build

docker compose --project-name wenqing-hr-demo --env-file .env.prod \
  -f docker-compose.prod.yml ps

curl -f http://localhost/healthz
```

### 13.3 空库首次初始化

仅首次、仅空库时叠加 bootstrap override：

```bash
docker compose --project-name wenqing-hr-demo --env-file .env.prod \
  -f docker-compose.prod.yml -f docker-compose.prod.bootstrap.yml up -d --build
```

`init.sql` 含 `DROP TABLE`，`mysql_data` 已有数据后严禁再叠加 bootstrap。

### 13.4 生产安全禁令

- 不要 `source .env.prod`（`DB_DSN` 含 `tcp()`、`?`、`&` 等特殊字符）；
- 不要省略 `--env-file .env.prod` 和 `--project-name`；
- 不要执行 `docker compose down -v`；
- 不要删除 `mysql_data` / `redis_data`；
- 不要在生产手动执行 `server/sql/init.sql`。

全库替换与回滚流程见 [`deploy/dumps/PROD_RESTORE_FROM_LOCAL_DUMP.md`](deploy/dumps/PROD_RESTORE_FROM_LOCAL_DUMP.md)。

## 14. agents/ 模块

`agents/` 是一个独立的 Go Module（GoFrame v2），与 `server/` 物理隔离，使用独立数据库 `offer_catcher_re`，默认端口 `18080`。当前只有 `/healthz`，定位是 IAM 与新架构的落地点，尚未接入前端与网关。

```bash
cd agents
cp .env.example .env      # 填写 DB_HOST / DB_PORT / DB_USER / DB_PASSWORD
make db-init              # 创建 offer_catcher_re
make start                # 编译 + 启动 + 健康检查
make test                 # 配置加载 + /healthz 接口测试
```

根目录 Makefile 与 `agents/` 无关；`make agent-*` 指的是主服务的 LLM 配置检查。

## 15. 常见问题

### Q1：`make up` 之后访问哪个地址？

应用是 http://localhost/offer/，API 是 http://localhost/api/，健康检查是 http://localhost/healthz。

### Q2：`make down` 会删数据库吗？

不会。`down` 只停止服务和容器，保留 volume。只有 `reset` / `reset-up` / `clean-volume` 会删除 volume。

### Q3：为什么后端和前端不放进 Compose？

Go 后端在宿主机运行更利于 IDE 断点调试与热重载；Vite 在宿主机运行的文件监听更可靠。Compose 只承担有状态基础设施与统一入口。

### Q4：为什么前端要在 `/offer/` 下？

前端 `vite.config.ts` 的 `base` 与 `routes.tsx` 的 `basename` 都是 `/offer/`，用于生产环境在公共网关下挂载子路径。开发环境同样需要带前缀访问。

### Q5：CORS 报错怎么办？

优先通过 `http://localhost/offer/` 访问，此时走 Nginx 同源代理，不需要 CORS。直连 `:8080` 调试时，确保 `CORS_ALLOWED_ORIGINS` 包含前端 Origin。

### Q6：如何配置 LLM？

在本地 `.env` 设置：

```env
LLM_ENABLE=true
LLM_PROVIDER=tokenhub
LLM_BASE_URL=https://your-openai-compatible-endpoint/v1/chat/completions
LLM_API_KEY=sk-your-local-key
LLM_FAST_MODEL=your-fast-model
LLM_QUALITY_MODEL=your-quality-model
```

然后执行 `make llm-check`；确需真实请求时执行 `RUN_REAL_LLM_SMOKE=1 make llm-smoke`。

### Q7：为什么质量档没有返回 LLM 结果？

按以下顺序排查：`LLM_ENABLE` 是否为 true、Provider 与 BaseURL / Key 是否匹配、`generation_tier` 是否传 `quality`、日志中是否出现 `fallback_reason`、`ENABLE_MOCK_AGENT` 与 `LLM_FALLBACK_ENABLE` 的取值。

### Q8：端口 5173 / 8080 被占用怎么办？

```bash
make backend-stop
make web-stop
make doctor
```

### Q9：数据库数据乱了如何回到干净状态？

```bash
make reset-up CONFIRM=yes
```

该命令会清空开发 volume 并重新导入 seed，属于破坏性操作。

### Q10：如何看某次 AI 调用为什么降级？

```bash
make logs-agent
```

关注 `[LLMRequest]`、`fallback_reason`、`error_code`、`[DeepCascade]`、`[SEP-Profiler]` 等日志标记。

## 16. 已知限制与风险

1. `make status` / `make help` 仍把基础设施描述为「MySQL + Nginx」，实际包含 Redis。
2. 根目录 `Makefile` 的 `.PHONY` 中声明了不存在的 `jobs-import-sql`；`rebuild`、`sep-backfill-resume4` 未列入 `.PHONY`。
3. `server/README.md`（阶段 1 骨架）与 `web/README.md`（Figma 导出说明）已完全过时。
4. `server/internal/http/router.go` 在 `/api/jobs/{id}/match` 分支内联复制了 Auth 中间件逻辑，未复用 `middleware.Auth`。
5. `server/internal/service/match_service.go` 的 legacy LLM 落库分支硬编码了 `generation_tier=fast` 与 `generated_by=fast_sep_match_engine`，与真实执行路径不一致。
6. 生产 `web/nginx.conf` 不含 `/api` 反代，`docker-compose.prod.yml` 却直接暴露 `web:80`；若没有外部 Gateway，生产 `/api` 请求会落到 SPA fallback。
7. `.gitignore` 未忽略 `deploy/dumps/*.sql`，2.6 MB 含真实业务数据的 dump 已入库。
8. `server/` 根目录与仓库根目录存在历史编译产物（`offer-catcher-server`、`sep-backfill`、`sep_gate_probe` 等）。
9. `.env.example` 中存在疑似真实格式的 API Key，应立即轮换并替换为占位符。
10. `ENABLE_MOCK_AGENT`、`LLM_FALLBACK_ENABLE` 在模板与代码中的默认值方向不一致，部署前需明确口径。
11. 缺少限流、请求 ID 贯穿、metrics / tracing 与统一 DB 迁移框架。
12. 前端依赖 `react` 18 运行时搭配 `@types/react` 19，类型检查存在潜在错配。

## 17. 文档索引

| 文档 | 面向对象 | 内容 |
|------|---------|------|
| [`docs/LLM_PROJECT_CONTEXT.md`](docs/LLM_PROJECT_CONTEXT.md) | **LLM / AI 编码助手** | 全量项目上下文、契约、调用链、风险与变更规程 |
| [`agents/README.md`](agents/README.md) | 后端 | GoFrame agents 模块说明 |
| [`deploy/dumps/PROD_RESTORE_FROM_LOCAL_DUMP.md`](deploy/dumps/PROD_RESTORE_FROM_LOCAL_DUMP.md) | 运维 | 生产全库替换与回滚手册 |
| [`error-cases.md`](error-cases.md) | 全体 | 60 条历史错误案例（P0/P1/P2） |
| `working/backend-architecture-uml/` | 架构 | 分层、ER、时序、Agent Pipeline 分析 |
| `working/参考数据/` | 前端 / 后端 | 旧系统对外契约参照 |

---

维护建议：改动接口、表结构、Agent 契约或启动方式时，同步更新本 README 与 `docs/LLM_PROJECT_CONTEXT.md`，避免文档与实现再次漂移。
