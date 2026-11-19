# Offer捕手 · LLM 项目上下文手册

> 本文件的目标读者是 **大语言模型 / AI 编码助手**，不是终端用户。
> 目标：让你在不全量扫描仓库的前提下，准确理解系统结构、契约、调用链与风险边界，并做出不破坏现有设计的改动。
>
> 适用基线：`main` 分支，`1fcf889`（2026-11）。
> 事实来源：源码 > 脚本 / Compose / SQL > 本文件 > 其他历史文档。若发现冲突，以代码为准，并在改完后同步本文件。

---

## 0. 使用说明

### 0.1 建议阅读顺序

| 顺序 | 章节 | 目的 |
|------|------|------|
| 1 | §1 项目速览 | 建立产品语义，避免把系统误解为普通招聘网站 |
| 2 | §2 系统全景 | 弄清浏览器、Nginx、前后端、MySQL、Redis、LLM 的边界 |
| 3 | §3 仓库地图 | 知道什么该读、什么可以忽略 |
| 4 | §4 后端架构 + §5 数据模型 | 掌握接口与持久化契约 |
| 5 | §6 AI 子系统 | 本项目最复杂、最容易改错的部分 |
| 6 | §7 前端架构 | 掌握页面、路由、适配层与 base path |
| 7 | §11 风险登记 + §12 变更规程 | 动手前必读 |

### 0.2 六条硬性不变量

改动任何代码前，先记住这六条。破坏其中任意一条，都会造成跨端故障：

1. **响应外壳固定**：所有接口返回 `{ "code": number, "message": string, "data": any }`；成功 `code = 0`。
2. **`/api` 不 strip**：Nginx 与 Vite 都把 `/api/...` 原样转发给后端，后端路由表以 `/api` 开头。
3. **前端 base 固定**：Vite `base` 与 react-router `basename` 都是 `/offer/`；组件内部只写相对路径。
4. **AI 元信息固定**：`meta.generated_by` / `agent_status` / `fallback_reason` / `error_code` 是前端判断结果可信度的唯一依据，字段不得缺失、不得改语义。
5. **Repair 只做一次**：Pipeline 不允许递归修复，`LLM_REPAIR_MAX_ATTEMPTS` 必须为 1。
6. **档位只有两档**：`generation_tier` 正式值只有 `fast` / `quality`；`deep` 仅作为 `quality` 的兼容别名。

### 0.3 不要做的事

- 不要引入 ORM、Web 框架或 LLM SDK 替换现有实现（项目刻意保持窄依赖）。
- 不要把真实 API Key 写入任何被 Git 跟踪的文件。
- 不要在日志中打印简历全文、JD 全文或模型原始输出。
- 不要把 `resume_versions` 当作当前主表（当前主表是 `resumes`，见 §5）。
- 不要只改 Prompt 而不改 Schema，或反之；两者是成对契约。
- 不要在 `quality` 档静默返回 Mock 而不设置 `meta`。

---

## 1. 项目速览

### 1.1 一句话

Offer捕手是一个**面向学生的投递前决策系统**：输入一份结构化简历与岗位库，输出可解释的岗位推荐、匹配分析、简历优化建议和版本对比结论。

### 1.2 它不是什么

- 不是招聘信息聚合站：岗位数据只是输入，核心价值是分析与决策。
- 不是投递系统：不做网申、投递、面试安排。
- 不是通用简历生成器：建议强绑定目标岗位 / 方向与证据链。
- 不是纯 LLM 套壳：Fast 档主要由规则、标签与证据引擎完成；LLM 只在 Quality 档承担深度推理。

### 1.3 用户旅程

```
注册 / 登录
  → 浏览岗位广场（公开）
    → 查看岗位详情（公开）
      → 选择简历
        ├── 岗位推荐（Fast 首屏 / Deep 深度）
        ├── 匹配分析（Fast 初筛 / Deep 深度）
        ├── 简历优化（关联岗位或方向）
        └── 版本对比（两份简历）
  → 根据建议编辑简历 → 再次分析 → 形成迭代闭环
```

### 1.4 术语表

| 术语 | 含义 |
|------|------|
| **Fast 档** | `generation_tier=fast`，低延迟路径，通常不调用 LLM，由 SEP / TagEvidence / 规则引擎完成 |
| **Quality 档** | `generation_tier=quality`，深度路径，调用 LLM，失败时逐级降级 |
| **SEP** | Semantic Evidence Profile，语义画像：简历 / 岗位的标签 + 证据集合，落 `semantic_profiles` / `semantic_profile_tags` |
| **TagEvidence** | 轻量标签证据引擎，Fast 档的规则基础，落 `resume_profiles` / `job_profiles` |
| **Deep Cascade** | 推荐 Quality 档的 L1–L4 级联管线（候选发现 → 语义重排 → 单岗综合 → 跨岗校准） |
| **3-Agent** | 通用的 `Decision → Verifier` 两段式（加上规则 Context 构建，共称 3-Agent）深度路径 |
| **Hy3 Screening** | 匹配 Quality 档的 5 模块筛查管线（决策 / 通过点 / 淘汰风险 / 改写指令 / 岗位策略） |
| **Pipeline** | 统一的 LLM 调用管线：Prompt → LLM → JSON → Schema → Repair → Fallback |
| **Mock Agent** | 完整的规则引擎降级实现，不是空桩 |
| **降级（fallback）** | 从高质量路径退回低质量但可用的路径；必须在 `meta` 中如实标记 |

---

## 2. 系统全景

### 2.1 运行时拓扑（开发环境）

```
┌──────────────────────────────────────────────────────────────┐
│ Browser                                                      │
│   http://localhost/offer/      http://localhost/api/...       │
└───────────────┬──────────────────────────┬───────────────────┘
                │                          │
        ┌───────▼──────────────────────────▼───────┐
        │ Nginx :80 (container)                    │
        │  location /      → host:5173             │
        │  location /api/  → host:8080             │
        │  location /healthz → host:8080/healthz   │
        └───────┬──────────────────────────┬───────┘
                │                          │
      ┌─────────▼─────────┐      ┌─────────▼──────────┐
      │ Vite :5173        │      │ Go Backend :8080   │
      │ React SPA         │      │ net/http           │
      └───────────────────┘      └───┬────────┬───────┘
                                     │        │
                              ┌──────▼──┐  ┌──▼─────┐
                              │ MySQL   │  │ Redis  │
                              │ :3306   │  │ :6379  │
                              └─────────┘  └────────┘
                                     │
                              ┌──────▼──────────────┐
                              │ LLM Provider (HTTP)  │
                              │ OpenAI-compatible    │
                              └─────────────────────┘
```

### 2.2 运行时拓扑（生产环境）

```
Browser → web 容器（Nginx 静态托管 :80）
             └── /api/* → server 容器（Go :8080）
                            ├── mysql 容器（不暴露宿主端口）
                            └── redis 容器（不暴露宿主端口）
```

注意：`web/nginx.conf` **不包含** `/api` 反代，注释声明由外部公共 Gateway 负责。若没有 Gateway，生产 `/api` 会落到 SPA fallback。这是已知部署缺口（见 §11）。

### 2.3 请求处理链路

```
HTTP Request
  → middleware.Recover      捕获 panic，返回 50000
  → middleware.CORS         白名单校验 + OPTIONS 204
  → middleware.RequestLogger 记录 method/path/status/duration
  → http.ServeMux
      → 公开路由
      → middleware.Auth（JWT 解析 + Redis 单会话校验）
      → Handler
          → Service
              → Repository → MySQL
              → Agent / Pipeline → LLM Provider
  → response.OK / response.Error
```

---

## 3. 仓库地图

### 3.1 权威性分级

| 等级 | 路径 | 说明 |
|------|------|------|
| A（事实来源） | `server/internal/**`、`server/prompts/**`、`server/sql/**`、`web/src/**` | 代码与资产本身 |
| A | `Makefile`、`docker-compose*.yml`、`nginx/**`、`scripts/**`、`.cnb.yml` | 工程化事实 |
| A | `README.md`、`docs/LLM_PROJECT_CONTEXT.md` | 当前维护的文档 |
| B（可用但需复核） | `working/backend-architecture-uml/**`、`working/参考数据/**` | 分析产物 / 旧契约参照 |
| C（历史，不可作为事实来源） | `server/README.md`、`web/README.md`、`backups/**`、`server/sql/init.sql.bak` | 已过时或为归档 |
| C | `working/deep-recommend-failure-report/**`、`working/refactor/**` | 历史故障 / 未批准设计 |

### 3.2 关键路径速查

| 你要找 | 去哪里 |
|--------|--------|
| 后端入口与依赖装配 | `server/cmd/server/main.go` |
| 路由表 | `server/internal/http/router.go` |
| 中间件 | `server/internal/middleware/` |
| 配置项 | `server/internal/config/config.go` + `.env.example` |
| 数据库 DDL | `server/sql/init.sql`（25 表） |
| 增量迁移 | `server/sql/migrations/` |
| 岗位 seed | `server/sql/seed/tencent_jobs.sql` |
| 业务编排 | `server/internal/service/`（37 个 Go 文件，含 10 个测试） |
| LLM 管线 | `server/internal/agent/pipeline/` |
| Prompt / Schema | `server/prompts/*.md`、`server/prompts/schema/*.json` |
| LLM Provider | `server/internal/llm/` |
| 前端路由 | `web/src/app/routes.tsx` |
| 前端页面 | `web/src/app/pages/`（12 个） |
| 前端 API | `web/src/api/` |
| DTO → ViewModel | `web/src/adapters/` |
| 全局状态 | `web/src/stores/authStore.tsx` |
| 开发命令 | `Makefile` + `make help` |
| 生产 compose | `docker-compose.prod.yml` |
| 生产恢复 | `deploy/dumps/PROD_RESTORE_FROM_LOCAL_DUMP.md` |
| 独立 agents 模块 | `agents/README.md` |

### 3.3 可以直接忽略的目录

`backups/`（历史备份）、`deploy/dumps/*.sql`（生产快照）、`working/`（分析记录）、`server/sql/init.sql.bak`、仓库根与 `server/` 下的编译产物。它们不参与构建，也不应作为设计依据。

---

## 4. 后端架构

### 4.1 分层与依赖方向

```
cmd/server
  └── internal/http（Handler）
        └── internal/service（业务编排）
              ├── internal/repository（SQL）
              ├── internal/agent（业务 Agent / Mock）
              │     └── internal/agent/pipeline
              │           ├── internal/prompt
              │           ├── internal/schema
              │           ├── internal/llmjson
              │           ├── internal/llm
              │           └── internal/agent/repair
              └── internal/model（领域模型）
```

约束：

- Repository **没有 interface 抽象**，全部是 `struct + *sql.DB`；Service 直接依赖具体类型。
- `agent.PipelineRunner` 是少数接口之一，用于让 Service 可注入 fast / quality 两个 Runner。
- `model` 不依赖任何上层包。
- `internal/service` 体量最大，包含 4 类能力编排 + SEP + Deep Cascade + Hy3 Screening。

### 4.2 启动装配顺序（`cmd/server/main.go`）

1. `config.Load()` + `config.Validate()`（弱口令仅告警）
2. `database.NewMySQL`，5s 超时 `Ping`，失败 `Fatal`
3. `health.NewService`
4. Redis 连接；失败只 `warn`，服务继续启动（会话校验将失效）
5. `session.NewManager`
6. Auth：`UserRepository → AuthService → AuthHandler → middleware.Auth`
7. Resume：`ResumeRepository → ResumeService`
8. ResumeProfile：`ResumeProfileRepository → ResumeProfileService`，并注入 ResumeService
9. Job：`JobRepository → JobService → JobHandler`
10. LLM（仅 `LLM_ENABLE=true`）：
    - 创建 fast / quality 两个 `llm.Client`
    - 加载 `prompt.Loader`、`schema.Loader/Validator`、`repair.Agent`
    - 创建两个 `pipeline.Runner`
    - 创建 SEP：`SemanticProfileRepository → SemanticProfileService → SemanticProfileProfiler → SemanticProfileGate`
11. `ProfileAgent`（LLM 开启时）注入 ResumeProfileService
12. `JobProfileRepository → TagEvidenceService`，启动两个后台回填 goroutine（岗位 tag profile、简历 evidence）
13. 创建 Match / Recommendation / Optimization / Compare 服务与 Handler
14. `middleware.InitCORS` → `NewRouter` → `ListenAndServe`

关键含义：**没有 LLM 也能启动**，此时 Quality 档会按配置直接报错或回落 Mock；Redis 不可用也能启动，但所有鉴权接口会失败。

### 4.3 完整路由表

所有路由注册在 `server/internal/http/router.go`。方法校验在各 Handler 内部完成。

#### 公开

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/healthz` | 服务 + DB 健康检查；DB 异常时 HTTP 200 + `message=degraded` |
| GET | `/api/jobs` | 岗位分页列表 |
| GET | `/api/jobs/facets` | 筛选项聚合 |
| GET | `/api/jobs/{id}` | 岗位详情 |
| GET | `/swagger/doc.json` | OpenAPI（仅覆盖部分接口） |
| GET | `/swagger/` | Swagger UI（CDN 资源） |
| ANY | `/` | 反代到 `FRONTEND_URL`（含 WebSocket HMR） |

#### 认证

| 方法 | 路径 | 请求体 / 说明 |
|------|------|--------------|
| POST | `/api/auth/register` | `{phone, password}`，返回 `{token, user}` |
| POST | `/api/auth/login` | `{phone, password}`，会顶掉旧会话 |
| GET | `/api/auth/me` | 需要 Bearer Token |
| POST | `/api/auth/logout` | 销毁 Redis 会话 |

#### 简历

| 方法 | 路径 | 说明 |
|------|------|------|
| GET / POST | `/api/resumes` | 列表 / 新建（关系型：主表 + 7 子表） |
| GET | `/api/resumes/check-name` | 名称重复校验，支持 `exclude_id` |
| GET | `/api/resumes/current` | 当前默认简历 |
| GET | `/api/resumes/detail` | `current` 的语义别名 |
| GET | `/api/resumes/versions` | 版本列表（legacy 语义，实为简历列表） |
| GET | `/api/resumes/versions/{versionNo}` | 指定版本 |
| GET / POST | `/api/resumes/versions/compare` | 版本对比（legacy 路径） |
| GET | `/api/resumes/compare/latest` | 最近对比结果 |
| GET / POST | `/api/resumes/compare` | 版本对比（V2 主路径） |
| GET / POST | `/api/resumes/optimize` | 优化 V1（legacy） |
| GET | `/api/resumes/{id}` | 简历详情 |
| PUT | `/api/resumes/{id}` | 更新简历 |
| DELETE | `/api/resumes/{id}` | 删除简历 |
| POST | `/api/resumes/{id}/set-current` | 设为当前简历 |
| GET | `/api/resumes/{id}/profile/status` | 简历画像状态 |
| POST | `/api/resumes/{id}/profile/regenerate` | 重新生成画像，body `{profile_type}` |
| GET | `/api/resumes/{id}/optimization/latest` | 最近优化结果 V2 |
| POST | `/api/resumes/{id}/optimization` | 生成优化 V2 |

`/api/resumes/{id}` 的二级分发在 `ResumeByIDDispatcher` 中手写完成，不支持任意子路径。

#### 岗位 AI

| 方法 | 路径 | 请求体 / 说明 |
|------|------|--------------|
| POST | `/api/jobs/recommend` | `{resume_id, limit, generation_tier, force_regenerate}` |
| GET | `/api/jobs/recommend/latest` | `resume_id`，可选 `generation_tier` |
| POST | `/api/jobs/{id}/match` | `{resume_id, generation_tier}` |
| GET | `/api/jobs/{id}/match/latest` | `resume_id`，可选 `tier` / `generation_tier` / `quality` |

`/api/jobs/{id}/match` 与 `/api/jobs/{id}/match/latest` 在 `dispatchJobWildcard` 中**内联实现鉴权**，未复用 `middleware.Auth`，且未注入 `session_id`。

#### 岗位列表查询参数

`keyword`、`direction`、`city`、`recruit_type`、`category`、`business_group`、`skill`、`application_project`、`application_subtype`、`application_type`、`external_source`、`job_category`、`page`、`page_size`、`limit`、`offset`。

- `city` / `business_group` / `skill` / `job_category` 支持逗号分隔多选。
- `job_category` 中父类（技术 / 产品 / 设计 / 市场 / 职能）映射到 `jobs.job_type`，其余映射到 `jobs.category`。
- `limit` / `offset` 存在时覆盖 `page` / `page_size`。

### 4.4 中间件与鉴权

顺序：`Recover → CORS → RequestLogger → ServeMux → Auth（按路由）`。

| 中间件 | 行为 |
|--------|------|
| `Recover` | `defer recover()`，打印 `[PANIC]` + stack，返回 50000 |
| `CORS` | 白名单来自 `CORS_ALLOWED_ORIGINS`；**额外硬编码放行任意端口的 localhost / 127.0.0.1 与 `*.cnb.run`**；OPTIONS 预检返回 204 |
| `RequestLogger` | 包装 ResponseWriter，记录 method/path/status/duration/remote/user_agent |
| `Auth` | 解析 `Authorization: Bearer <jwt>` → `ParseToken` → Redis `ValidateSession` → 注入 `user_id` / `session_id` |

鉴权语义：

- JWT Claims：`{user_id, phone, session_id, iat, exp}`。
- Redis 保存「当前有效 session_id」；新登录会顶掉旧会话。
- 旧会话再次请求 → `41005`（账号已在其他设备登录），前端据此清 token 并提示。
- 其他 401 → 前端清 token 并派发 `offerr-auth-expired`。

### 4.5 统一响应与错误码

成功：

```json
{ "code": 0, "message": "ok", "data": {} }
```

失败：

```json
{ "code": 42001, "message": "resume version not found", "data": null }
```

错误码段位（`server/internal/errors/error.go`）：

| 段位 | 含义 |
|------|------|
| `40000` | bad request |
| `40001` | invalid parameter |
| `40100` | unauthorized |
| `40300` | forbidden |
| `40400` | not found |
| `40500` | method not allowed |
| `41002` | 手机号或密码错误 |
| `41003` | 手机号已注册 |
| `41004` | 会话无效 |
| `41005` | 会话被顶替 |
| `42001` | 简历版本不存在 |
| `42002` | 简历不存在 |
| `42003` / `42004` | 对比的 from / to 简历不存在 |
| `42005` | 简历名称重复 |
| `43001` | 岗位不存在 |
| `50000` | internal error |
| `50300` | dependency unavailable |

### 4.6 配置系统

唯一加载点：`server/internal/config/config.go`。

- 使用 `godotenv` 依次尝试 `.env`、`../.env`、`../../.env`、`../../../.env`，失败仅告警。
- 环境变量优先于 `.env`。
- 后端**只读 `DB_DSN`** 建立连接；`DB_HOST` / `DB_PORT` 等仅作用于脚本与 Compose。
- `Config.Validate()` 只做弱口令告警，不阻止启动。

代码默认值 vs 模板值（易踩坑）：

| 变量 | 代码默认 | `.env.example` |
|------|---------|----------------|
| `LLM_ENABLE` | `false` | `true` |
| `LLM_PROVIDER` | `hunyuan` | `tokenhub` |
| `LLM_FALLBACK_ENABLE` | `true` | `false` |
| `ENABLE_MOCK_AGENT` | `true` | `false` |
| `LLM_MAX_OUTPUT_TOKENS` | `4096` | `4096` |
| `LLM_TIMEOUT_SECONDS` | `120` | `180` |
| `LLM_MODEL` | 空 | `hy3-preview`（legacy fallback） |

---

## 5. 数据模型

### 5.1 表清单（25 张，`server/sql/init.sql`）

#### 用户

| 表 | 关键字段 | 约束 |
|----|---------|------|
| `users` | `id`, `phone`, `password_hash`, `name`, `email` | UK `phone` |

#### 岗位

| 表 | 关键字段 | 说明 |
|----|---------|------|
| `jobs` | `title`, `department`, `city`, `direction`, `category`, `job_type`, `application_type`, `description`, `requirements`, 多个 `*_json` 字段, `external_source`, `external_id`, `data_quality`, `status` | 岗位主表，字段极宽，含原始 JSON |
| `job_locations` | `job_id`, `city` | 多城市 |
| `job_business_groups` | `job_id`, `code`, `name` | 事业群 |
| `job_skills` | `job_id`, `skill` | 技能标签 |
| `job_profiles` | `job_id`, `direction`, `category`, `*_tags_json`, `evidence_json`, `source`, `version` | 岗位标签证据画像 |
| `job_filter_groups` | `code`, `name`, `display_order`, `visible` | 筛选项分组 |
| `job_filter_options` | `group_id`, `code`, `label`, `value`, `parent_option_id` | 筛选项 |
| `job_filter_option_relations` | `job_id`, `option_id` | 岗位与筛选项关联 |

#### 简历（1 + 7 子表）

| 表 | 关键字段 |
|----|---------|
| `resumes` | `id`, `user_id`, `name`, `is_current`, `real_name`, `phone`, `email`, `city`, `professional_skills`, `raw_text` |
| `resume_educations` | `school`, `degree`, `major`, `department`, `start_date`, `end_date`, `gpa`, `gpa_scale`, `rank_text`, `core_courses`, `research_direction`, `advisor`, `lab`, `paper_summary`, `sort_order` |
| `resume_projects` | `name`, `role`, `start_date`, `end_date`, `description`, `sort_order` |
| `resume_internships` | `company`, `position`, `start_date`, `end_date`, `description`, `sort_order` |
| `resume_language_exams` | `exam_type`, `score`, `sort_order` |
| `resume_certificates` | `name`, `description`, `sort_order` |
| `resume_awards` | `name`, `description`, `sort_order` |
| `resume_works` | `name`, `url`, `password`, `sort_order` |

约束：所有子表 `FK resume_id → resumes(id) ON DELETE CASCADE`；`resumes` 有 `UK (user_id, name)`。

#### AI 结果

| 表 | 关键字段 | 说明 |
|----|---------|------|
| `match_analyses` | `user_id`, `resume_id`, `job_id`, `score`, `result_json`, `generation_tier`, `generated_by`, `status`, `profile_type`, `profile_dominant_profile` | 匹配结果 |
| `recommendation_results` | `user_id`, `resume_id`, `result_json`, `generation_tier`, `generated_by`, `status`, `profile_type` | 推荐结果 |
| `resume_optimization_results` | `user_id`, `resume_id`, `job_id`, `result_json`, `optimization_mode`, `generation_tier`, `generated_by`, `target_direction` | 优化结果 |
| `resume_compare_results` | `user_id`, `from_resume_id`, `to_resume_id`, `result_json` | 对比结果 |
| `resume_profiles` | `resume_id`, `user_id`, `profile_type`, `status`, `dominant_profile`, `*_tags_json`, `evidence_json`, `source`, `generator` | 简历标签画像，`UK (resume_id, profile_type)` |

结果表统一追加的 agent 元字段（`sql/migrations/20260611_k13_result_tables_add_agent_fields.sql`）：

`agent_status`、`fallback`、`repair_applied`、`error_code`、`duration_ms`、`model_name`、`result_version`。

#### SEP

| 表 | 关键字段 | 说明 |
|----|---------|------|
| `semantic_profiles` | `target_type`（`resume`/`job`）, `target_id`, `profile_version`, `prompt_version`, `schema_version`, `content_hash`, `model_provider`, `model_name`, `reasoning_effort`, `summary_sentence`, `*_json`, `status` | 语义画像主表；唯一键含 `content_hash`，内容变化即产生新版本 |
| `semantic_profile_tags` | `profile_id`, `target_type`, `target_id`, `tag_code`, `tag_label`, `category`, `confidence`, `evidence_text`, `explanation`, `source_section`, `is_user_visible`, `weight` | 标签明细 |

标签 `category` 枚举：`role_direction`、`technical_skill`、`engineering_capability`、`domain_context`、`experience_signal`、`risk_signal`、`preference_signal`。

`semantic_profiles.status`：`pending` / `processing` / `verified` / `failed` / `stale`。

#### 运维

| 表 | 说明 |
|----|------|
| `seed_versions` | seed 幂等追踪（`seed_key`, `seed_version`, `checksum`, `applied_at`） |

### 5.2 核心关系

```
users 1 ── n resumes
resumes 1 ── n resume_educations / projects / internships / language_exams / certificates / awards / works
resumes 1 ── n resume_profiles
resumes 1 ── n match_analyses / recommendation_results / resume_optimization_results / resume_compare_results
jobs 1 ── n job_locations / job_business_groups / job_skills
jobs 1 ── 1 job_profiles
jobs 1 ── n match_analyses / resume_optimization_results
semantic_profiles 1 ── n semantic_profile_tags
```

### 5.3 `resumes` 与 `resume_versions`

- `resume_versions` 是历史表结构，`sql/resume_versions.sql` 明确说明已被 `resumes` 取代（K8-5-51-B）。
- 当前代码只读写 `resumes` 与 7 张子表。
- `model.Resume` 保留了 `SchemaVersion`、多个 `*JSON` 等 shim 字段，仅用于兼容旧快照，不代表数据库仍有这些列。
- **不要新增对 `resume_versions` 的写入**。

### 5.4 迁移与 seed 策略

- 没有迁移框架。`sql/migrations/*.sql` 由人工执行。
- `sql/migrations/20260611_k13_run.go` 是带 `//go:build ignore` 的一次性执行器：先查 `information_schema.COLUMNS`，列不存在才 `ALTER TABLE ... ADD COLUMN`，因此可重复执行。MySQL 8 不原生支持 `ADD COLUMN IF NOT EXISTS`，不要照搬 MariaDB 写法。
- `init.sql` 含 `DROP TABLE`，只用于开发环境初始化或生产**空库** bootstrap。
- 岗位 seed 通过 `seed_versions` 判断是否已应用，避免重复导入。

---

## 6. AI 子系统（核心）

### 6.1 组件分层

```
Service 层（业务编排 / tier 选择 / 落库 / 降级）
  └── Business Agent（prompt 值组装 / 结果映射 / fallback 判定）
        └── pipeline.Runner（统一管线）
              ├── prompt.Loader
              ├── schema.Loader + schema.Validator
              ├── llm.Client（hunyuan / hy3 / noop）
              ├── llmjson.ExtractAndParse
              └── repair.Agent（仅一次）
```

### 6.2 Pipeline 的 9 个步骤

实现：`server/internal/agent/pipeline/pipeline.go`。

1. 校验 `RunRequest`：`AgentName`、`PromptName`、`SchemaName`、`TargetOutputType` 必填且合法。
2. `prompt.Loader.Render`：加载 `.md` 并替换 `{{KEY}}` 占位符。
3. `schema.Loader.Load`：加载目标 JSON Schema。
4. `llm.Client.Chat`：两条消息（system + user），记录 `[LLMRequest]` 诊断日志。
5. `llmjson.ExtractAndParse`：三级抽取（整体 JSON → code fence → 花括号配平）。
6. `schema.Validator.Validate`：JSON Schema 校验。
7. 失败则 `repair.Agent.Repair`：渲染 `json_repair_v1.md`，单次 LLM 修复。
8. 修复结果再次 parse + validate。
9. 返回 `RunResult`。

状态机：

| 状态 | 含义 | `GeneratedBy` |
|------|------|--------------|
| `success` | 首次 parse + validate 通过 | `llm_agent` |
| `repaired` | 首次失败，修复后通过 | `llm_agent` |
| `fallback_required` | 修复仍失败 / LLM 调用失败 / 配置缺失 | `mock_agent` |

`RunResult` 关键字段：`AgentStatus`、`GeneratedBy`、`Parsed`、`JSONText`、`RepairApplied`、`RepairSuccess`、`FallbackRequired`、`FallbackReason`、`ErrorCode`、`DurationMs`、`RawOutputLength`、`LLMProvider`、`LLMModel`、`FinishReason`、`ValidationErrors`。

注意：

- Pipeline **没有重试循环**，修复只调用一次；`repair.Agent.Repair` 内部使用 `context.Background()`，不受请求超时约束（已知风险）。
- `RawContent` 只用于内部诊断，不返回、不默认落日志。

### 6.3 Provider 层

实现：`server/internal/llm/`。

| Provider | 文件 | 说明 |
|----------|------|------|
| `hunyuan` / `deepseek` / `tokenhub` | `hunyuan_client.go` | 共用 OpenAI 兼容实现，按 endpoint 规范化 URL |
| `hy3` | `hy3_client.go` | LKEAP `hy3-preview`，注入 `reasoning_effort` 到 `extra_body.chat_template_kwargs` |
| `noop` | `noop_client.go` | 测试 / 占位，开启时返回固定 `{"noop":true}` |

Endpoint 规范化规则：

- 以 `/chat/completions` 结尾 → 原样使用；
- 以 `/v1` 结尾 → 追加 `/chat/completions`；
- 其他 → 追加 `/v1/chat/completions`。

错误分类：`LLM_DISABLED`、`LLM_CONFIG_MISSING`、`LLM_API_KEY_MISSING`、`LLM_AUTH_FAILED`、`LLM_TIMEOUT`、`LLM_HTTP_ERROR`、`LLM_RATE_LIMIT`、`LLM_EMPTY_RESPONSE`、`LLM_OUTPUT_TRUNCATED`。

Fast / Quality 使用不同模型：

- `LLM_FAST_MODEL` 为空 → 回落 `LLM_MODEL`
- `LLM_QUALITY_MODEL` 为空 → 回落 `LLM_MODEL`

### 6.4 Prompt 与 Schema 资产

`server/prompts/` 下 14 个 Prompt、`schema/` 下 13 个 Schema。

主要映射：

| Prompt | 用途 | 对应 Schema |
|--------|------|------------|
| `match_analysis_v1.md` | 匹配分析单 Agent | `match_analysis_schema.json` |
| `recommendation_v1.md` | 推荐单 Agent | `recommendation_schema.json` |
| `resume_optimization_v1.md` | 优化 V1 | `resume_optimization_schema.json` |
| `resume_optimization_core_v1.md` | 优化精简输出 | `resume_optimization_core_schema.json` |
| `version_compare_v1.md` | 版本对比 V1 | `version_compare_schema.json` |
| `json_repair_v1.md` | JSON 修复 | 无（使用目标 Schema 校验） |
| `match_decision_v1.md` / `match_verifier_v1.md` | 3-Agent 匹配 | 对应 `*_schema.json` |
| `recommendation_decision_v1.md` / `recommendation_verifier_v1.md` | 3-Agent 推荐 | 对应 `*_schema.json` |
| `optimization_decision_v1.md` / `optimization_verifier_v1.md` | 3-Agent 优化 | 对应 `*_schema.json` |
| `compare_decision_v1.md` / `compare_verifier_v1.md` | 3-Agent 对比 | 对应 `*_schema.json` |

`prompts/_shared/` 是共享约束片段：

- `frontend_contract.md`：`generated_by` / `agent_status` / `generation_tier` 取值；数组字段不得为 null。
- `evidence_rules.md`：证据链规则。
- `json_output_rules.md`：只输出 JSON、禁止 Markdown 等。

Prompt 加载是**名称映射**，不是路径扫描（`server/internal/prompt/loader.go`）。新增 Prompt 必须同时改映射表，否则 `PROMPT_LOAD_FAILED`。

### 6.5 Fast / Quality 双档总览

| 能力 | Fast | Quality |
|------|------|---------|
| 推荐 | `fastRecommendWithSEP` → `fastRecommendWithTags` → Mock | `recommendDeepCascade` → `recommendWith3Agents` → `recommendWithLLM` → Mock |
| 匹配 | `fastMatchWithSEP` → `buildFastSEPPendingResponse` | `runHy3ScreeningPipeline` → `matchWith3Agents` → `analyzeWithLLM` → Mock |
| 优化 | `fastOptimizeWithTags` → Mock V2 | `optimizeWith3Agents` → `ResumeOptimizationAgent` V1 → Mock V2 |
| 对比 | `fastCompareWithTags` → Mock | `compareWith3Agents` → `Compare` V1 → Mock |

共同规则：

- `generation_tier` 未传 → `fast`。
- Quality 失败后是否允许 Mock，取决于 `LLM_FALLBACK_ENABLE` 与 `ENABLE_MOCK_AGENT`。
- 每个结果落库时都要写 `generation_tier`、`generated_by`、`profile_type`、`profile_dominant_profile`。

**落库矩阵（重要，容易误解）**：

| 能力 | Fast 结果落库 | Quality 结果落库 | `/latest?tier=fast` |
|------|--------------|-----------------|---------------------|
| 推荐 | 否（SEP 与 Tag 路径均直接返回） | 是（Deep Cascade / LLM / Mock 兜底） | 查不到 Fast 结果 |
| 匹配 | 否（`analysis_id=0`） | 是 | 查不到 Fast 结果 |
| 优化 | 是（`fastOptimizeWithTags` 内部 `saveOptV2Snapshot`） | 是 | 可查到 |
| 对比 | 是（`fastCompareWithTags` 内部 `saveCompareV2`） | 是 | 可查到 |

这意味着「Fast 是即时计算，Quality 是持久结果」只适用于推荐与匹配；优化与对比的 Fast 也会持久化。改动 `/latest` 或落库逻辑时务必按此矩阵确认。

### 6.6 四条端到端调用链

#### 6.6.1 匹配分析 `POST /api/jobs/{id}/match`

```
dispatchJobWildcard（内联鉴权，注入 user_id + job_id）
→ MatchHandler.Analyze
→ MatchService.Analyze
   → getResume / jobRepo.FindByID
   ├ fast:
   │   → checkFastSEPContext（查 semantic_profiles ×2）
   │   → ready：fastMatchWithSEP → runFastSEPMatch
   │        → 直接返回，analysis_id = 0，不写 match_analyses
   │   → 未就绪：buildFastSEPPendingResponse（score=0，不写库）
   ├ quality:
   │   → runScreeningMatchEntry → runHy3ScreeningPipeline
   │        M1 Decision → M2 PassPoints → M3 EliminationRisks
   │        → M4 ResumeRewriteCommands → M5 JobStrategy → Judge
   │   → 失败：matchWith3Agents（Context → Decision → Verifier）
   │   → 失败：analyzeWithLLM（单 Agent Pipeline）
   │   → 失败：analyzeWithMock
   → quality / mock 路径写 match_analyses
→ response.OK({analysis_id, result, analysis_basis, meta})
```

SEP Fast 引擎内部顺序（`match_sep_engine.go`）：

```
normalizeTerm / termMatch（同义词归一）
→ detectDirection / analyzeDirectionFit
→ checkHardGate
→ buildRequirementCoverage
→ mapResumeEvidence
→ detectMatchRisks
→ calibrateFastMatchScore（clamp 0–100）
→ decideApplication
→ buildConclusion
→ assembleFastMatchResult
```

#### 6.6.2 岗位推荐 `POST /api/jobs/recommend`

```
RecommendationHandler.Recommend
→ RecommendationService.Recommend
   → limit 夹紧（≤5）
   → profileSvc.EnsureProfile（fast / quality）
   → quality 且非 force_regenerate：查缓存命中直接返回
   → recommendCore
      → getResume + jobRepo.ListRecommendCandidates(50)
      ├ fast: fastRecommendWithSEP
      │       → SemanticProfileGate 异步确保画像
      │       → 标签归一 + 打分排序
      │       → 解释器生成理由 / 标签 / 建议
      ├ quality: recommendDeepCascade
      │       L0 SemanticProfileGate
      │       L1 Candidate Discovery（high ×1）
      │       L2 Semantic Rerank（high ×1）
      │       L3 Per-Job Synthesis（high ×5）
      │       L4 Cross-Job Calibration（low ×1）
      ├ 回落: fastRecommendWithTags → 3-Agent → LLM → Mock
   → 仅 quality / LLM / Mock 兜底路径写 recommendation_results
   → fast SEP / fast Tag 结果直接返回，不落库
```

#### 6.6.3 简历优化 `POST /api/resumes/{id}/optimization`

```
ResumeByIDDispatcher（注入 resumeID）
→ ResumeOptimizationHandler.OptimizeV2
→ ResumeOptimizationService.OptimizeV2
   → 校验 optimization_mode（general / target_direction / target_job）
   → 校验 target_direction / job_id 与对应模式匹配
   ├ fast: fastOptimizeWithTags（TagEvidence）
   ├ quality: optimizeWith3Agents
   │          → 失败：ResumeOptimizationAgent V1
   │          → 失败：Mock V2
   → filterEducationOptimizationAdvice（无教育经历时过滤学历类建议）
   → saveOptV2Snapshot（持久化 + 初始化结果快照）
```

#### 6.6.4 版本对比 `POST /api/resumes/compare`

```
VersionCompareHandler.Compare（POST → CompareV2）
→ VersionCompareService.CompareV2
   → 校验 from / to 存在且不相等
   → 校验 comparison_mode（general / target_direction / target_job）
   → 确保两份简历的 fast profile
   ├ fast: fastCompareWithTags（TagEvidence diff）
   ├ quality: compareWith3Agents
   │          → 失败：Compare V1（LLM）
   │          → 失败：Mock
   → saveCompareV2
```

### 6.7 SEP 语义画像

核心服务：

| 组件 | 职责 |
|------|------|
| `SemanticProfileService` | 内容哈希、查询 verified 画像、标记 stale、状态更新 |
| `SemanticProfileProfiler` | 调用 Hy3 生成画像与标签 |
| `SemanticProfileGate` | 确保画像可用：hash → 查 verified → 标签数 < 5 视为损坏 → 标记 stale → 生成 |
| `SemanticProfileValidator` | 校验 LLM 输出结构与标签合法性 |

关键行为：

- 画像按 `target_type + target_id + profile_version + schema_version + content_hash` 唯一。
- 内容变化 → 新 hash → 旧画像自然失效。
- `EnsureResumeProfileAsync` / `EnsureJobProfileAsync` 使用 `asyncInFlight` map 去重，后台 goroutine 有 2 分钟超时。
- Gate 层把「verified 但标签数 < 5」视为损坏画像，标记失败并触发重建。
- Fast 匹配另有一条更低的门槛：简历 / 岗位画像各至少 3 个标签，否则视为未就绪，返回 `profile_pending`。

### 6.8 Deep 级联与 3-Agent

| 名称 | 位置 | 适用 | 结构 |
|------|------|------|------|
| Deep Cascade | `recommend_deep_cascade.go` | 推荐 Quality | L0 Gate → L1 候选发现 → L2 语义重排 → L3 单岗综合 → L4 跨岗校准 |
| Hy3 Screening | `match_deep_agent.go` | 匹配 Quality | M1–M5 五模块 + 本地 Judge |
| 3-Agent（通用） | `*_deep_agents.go` | 四类能力 | `BuildContext`（纯规则）→ `DecisionAgent`（high）→ `VerifierAgent`（low） |
| 单 Agent | `*_agent.go` | 最后一级 LLM 降级 | Pipeline + 单 Prompt / Schema |
| Mock | `*_mock*.go` | 最终兜底 | 规则引擎 |

3-Agent 的通用语义：

- Context 层是纯规则，不调用 LLM，用于压缩输入、生成候选与基线分。
- Decision 层负责主判断，使用 `high` reasoning effort。
- Verifier 层负责一致性校验与修正，使用 `low` reasoning effort。
- 任一层失败，整条 3-Agent 链路失败，进入下一级降级。

### 6.9 输出契约（前端最敏感）

所有 AI 响应都在 `data.meta` 中带统一元信息（`model.AgentResponseMeta`）：

| 字段 | 取值 |
|------|------|
| `agent_name` | `match_analysis` / `recommendation` / `resume_optimization` |
| `generated_by` | `llm_agent` / `mock_agent` / `deep_fallback_agent` / `mock_quality`（legacy）/ `fast_tag_engine` / `fast_sep_match_engine` / `hy3_screening_match_agent` 等 |
| `agent_status` | `success` / `repaired` / `fallback` / `mock_success` / `error` |
| `fallback_reason` | `none` / `llm_disabled` / `llm_error` / `llm_timeout` / `json_parse_failed` / `json_repair_failed` / `schema_invalid` / `pipeline_error` / `llm_api_key_missing` / `llm_auth_failed` / `deep_cascade_failed` / `quality_exhausted` / `quality_cancelled` 等 |
| `error_code` | `none` / `LLM_DISABLED` / `LLM_ERROR` / `LLM_TIMEOUT` / `JSON_PARSE_FAILED` / `JSON_REPAIR_FAILED` / `SCHEMA_INVALID` / `PIPELINE_ERROR` / `MOCK_FALLBACK_USED` / `VALIDATION_ERROR` 等 |
| `repair_applied` | bool |
| `schema_valid` | bool |
| `request_id` | 预留，当前多为空 |

前端判定（`web/src/app/utils/agentError.ts`）：

- `mock_agent` 与 `mock_quality` 一律被 `isMockGeneratedBy` 判定为 Mock，直接抛 `RealAgentError`。
- `fallback === true` 一律视为降级结果，直接抛错。
- `quality` 档只接受 `llm_agent` / `deep_fallback_agent`；其他值（包括 `fast_tag_engine`）会抛错。
- `fast` 档不强制特定 `generated_by`，只要不是 Mock 且 `fallback !== true` 即通过；实际常见值为 `fast_tag_engine` / `fast_sep_match_engine`。
- 注意：`isDeepTierGeneratedBy` 仍把 `mock_quality` 列为 Deep 值，但 `isMockGeneratedBy` 会先将其拦截，二者语义不一致，属于历史遗留。
- 数组字段（如 `evidence`、`risks`、`suggestions`）**不得为 null**，空数组也必须返回 `[]`。

### 6.10 成本、超时、日志与脱敏

- Token 统计存在（`llm.Usage`），但**只写日志，不落库、不计费**。
- 超时：`HunyuanClient` 使用 `LLM_TIMEOUT_SECONDS` 或单请求覆盖；Hy3 同。
- Repair 使用 `context.Background()`，没有超时继承。
- 日志标签：`[LLMRequest]`、`[LLM-Hy3]`、`[SEP-Profiler]`、`[DeepCascade]`、`[OptimizeV2]`、`[VersionCompareService]`。
- 错误信息经过 `safeErr` 截断 200 字符 + `redactSensitive` 抹除 `key=` / `token=` / `sk-` / `api_key=` / `apikey=`。
- `LLM_LOG_RAW_OUTPUT=false` 时不得打印模型原始输出。

---

## 7. 前端架构

### 7.1 技术栈与入口

- React 18.3.1（运行时）+ TypeScript 6 + Vite 6.3.5 + Tailwind v4 + shadcn/ui。
- 入口：`web/src/main.tsx` → `App.tsx`（AuthProvider + AuthBoot + RouterProvider）。
- 路由：`web/src/app/routes.tsx`，`createBrowserRouter`，`basename: "/offer/"`。
- 构建：`vite.config.ts` 中 `base: "/offer/"`，别名 `@ → ./src`，dev proxy `/api` 与 `/healthz` → `:8080`。

### 7.2 路由表

| 路径 | 页面 | 保护 |
|------|------|------|
| `/`、`/jobs` | 岗位广场 | 公开 |
| `/jobs/:id`、`/job/:id` | 岗位详情 | 公开 |
| `/recommend` | 岗位推荐 | RequireAuth + ErrorBoundary |
| `/match` | 匹配分析 | RequireAuth + ErrorBoundary |
| `/resume-optimization` | 简历优化 | RequireAuth + ErrorBoundary |
| `/resumes` | 简历管理 | RequireAuth |
| `/resumes/edit`、`/resumes/new` | 简历编辑 / 新建 | RequireAuth |
| `/resumes/compare` | 版本对比 | RequireAuth + ErrorBoundary |
| `/resumes/detail/:id` | 简历详情 | RequireAuth |
| `/login`、`/register` | 重定向壳，交给 Navbar + AuthDialog | 公开 |
| `*` | 404 | — |

旧路径 `/recommendations`、`/match-analysis`、`/my-resumes*`、`/resume-comparison` 均重定向到新路径。

### 7.3 页面清单

| 页面 | 文件 | 核心内容 |
|------|------|---------|
| 岗位广场 | `JobListPage.tsx` | 左筛选（关键词 + facets）、中岗位卡片流 + 分页、右 AI 决策侧栏 |
| 岗位详情 | `JobDetailPage.tsx` | JD / 要求 / 技能 / 投递说明 + 右侧 AI 决策入口 |
| 岗位推荐 | `JobRecommendationPage.tsx` | Top5 榜单 + 单岗评估报告 + Fast / Deep 控制台 |
| 匹配分析 | `JobMatchAnalysisPage.tsx` | 分数 Hero + 三卡结论 + 投递准备度 + 证据链 + Deep 折叠区 |
| 简历优化 | `ResumeOptimizationPage.tsx` | 上下文条 + Fast / Deep 双视图 + JD 证据链 + 改写示例 |
| 版本对比 | `ResumeComparisonPage.tsx` | 结论卡 + 六段式报告 + 前→后改写示例 |
| 简历管理 | `ResumeManagementPage.tsx` | 简历列表 + 智能求职助手入口 |
| 简历编辑 | `MyResumesPage.tsx` | 分区表单 + 右侧准备度侧栏 |
| 简历详情 | `ResumeDetailPage.tsx` | 只读简历展示 |

### 7.4 数据流

```
页面
  → src/api/*（请求）
  → src/api/normalizeAgentResponse.ts（兼容 data.result / data 两种形状）
  → src/adapters/*Adapter.ts（DTO → ViewModel）
  → src/api/viewModels.ts（统一 Loader）
  → 页面渲染
```

- 简单读写页面直接用 `useEffect + api/* + useState`。
- AI 页面必须走 normalizer + adapter，避免后端字段演进直接击穿 UI。
- 状态管理只有 `authStore`（Context + useReducer）；其余为页面级状态。
- 没有 react-query / SWR / zustand / redux。

### 7.5 鉴权与会话

- Token 存 `localStorage`，key 为 `offer_catcher_auth_token`（`api/authStorage.ts`）。
- 请求头由 `api/client.ts` 自动注入 `Authorization: Bearer <token>`。
- 401 处理：
  - `code === 41005` → 清 token + 派发 `offerr-auth-replaced`（异地登录）。
  - 其他 401 → 清 token + 派发 `offerr-auth-expired`。
- `RequireAuth` 基于 `authStore`，未登录会带上 `redirect` 参数跳回首页登录弹窗。
- 跳转白名单在 `utils/redirect.ts`，防止开放重定向。

### 7.6 设计系统

- Tailwind v4 CSS-first，无 `tailwind.config.js`。
- 主题变量在 `styles/theme.css`；`@custom-variant dark`。
- 品牌色：`#0052D9`，页面底色 `#F5F7FB`，列表容器 `#F0F2F5`。
- `src/app/components/ui/` 有 54 个文件（含 `utils.ts`），但实际高频使用的只有 Button / Badge / Input / Select / Dialog / Skeleton 等十余个。
- 字体：品牌字体栈 Alimama ShuHeiTi / MiSans / HarmonyOS Sans SC；正文 HarmonyOS Sans SC / Noto Sans SC。

### 7.7 前端契约陷阱

1. `base: "/offer/"` 与 `basename: "/offer/"` 必须同时存在；只改一个会导致路由或静态资源 404。
2. 组件内部路径必须写相对路径，不能手工加 `/offer/`。
3. 后端 `data.result` 与 `data` 两种形状并存，必须经过 `normalizeAgentResponse`。
4. 数组字段可能为 null（历史问题），adapter 必须做兜底。
5. `mock_agent` / `mock_quality` / `fallback=true` 结果不能当作正常结果展示；`quality` 档还要求 `generated_by` 必须是 `llm_agent` 或 `deep_fallback_agent`。
6. `react` 18 运行时搭配 `@types/react` 19，类型检查存在错配风险。

---

## 8. 工程化与运行

### 8.1 Makefile 心智模型

Makefile 是**薄编排层**，实际逻辑都在 `scripts/dev/*.sh`。`make help` 是唯一可靠的命令清单。

| 分组 | 代表命令 |
|------|---------|
| 环境 | `doctor`、`init`、`ensure-env`、`check`、`status`、`health` |
| 生命周期 | `up`、`stop`、`down`、`restart`、`reset`、`reset-up`、`rebuild` |
| 基础设施 | `infra-up`、`infra-down`、`infra-status`、`infra-logs` |
| 数据库 | `db-import`、`db-reset`、`db-shell`、`db-check` |
| 后端 | `backend-up`、`backend-stop`、`backend-restart`、`backend-logs`、`backend-health` |
| 前端 | `web-install`、`web-up`、`web-stop`、`web-restart`、`web-logs` |
| 日志 | `logs-backend`、`logs-web`、`logs-mysql`、`logs-agent`、`nginx-logs` |
| LLM | `llm-check`、`llm-smoke`、`agent-env`、`agent-check` |
| SEP | `sep-backfill-dry-run`、`sep-backfill`、`sep-backfill-resume4` |
| 质量 | `lint`、`test`、`build`、`ci` |
| 岗位数据 | `jobs-fetch*`、`jobs-import*`、`jobs-seed-*` |
| 生产 | `prod-config`、`prod-check`、`prod-backup` |
| 清理 | `clean`、`clean-logs`、`clean-pids`、`clean-volume` |

危险命令需要 `CONFIRM=yes`；`clean-volume` 还需要交互输入 `YES`。

### 8.2 脚本职责

| 脚本 | 职责 |
|------|------|
| `common.sh` | 颜色、日志、`load_env`、`mask_secret`、PID 树清理、端口检查、危险确认 |
| `up.sh` | 11 步一键启动 + 失败汇总 |
| `doctor.sh` | 只读环境体检 |
| `ensure_env.sh` | `.env` 就绪 + 双模型变量幂等补齐 |
| `infra.sh` | Compose 薄封装 |
| `backend.sh` | Go 进程管理（nohup + PID + 端口兜底） |
| `web.sh` | Vite 进程管理（自动 npm ci / install） |
| `mysql.sh` | 导入 / 重置 / shell / 表检查 |
| `db-backup.sh` | 本地库备份（当前未被 Makefile 引用） |
| `quality.sh` | lint / test / build / ci |
| `llm_smoke.sh` | LLM 配置检查 + 受控 smoke |
| `tencent_jobs.sh` | 腾讯招聘公开接口低频拉取 |
| `tencent_jobs_import.sh` | raw JSON → SQL → MySQL |
| `logs.sh` | 日志分流 |
| `clean.sh` | 清理运行态 / volume |
| `agent.sh` | LLM 变量展示与校验（`test` 为空实现） |
| `prod-check.sh` | 9 段生产前只读检查 |
| `prod-backup.sh` | `mysqldump | gzip` → `backups/mysql/` |

`load_env` 逐行解析 `.env` 而不是 `source`，因为 `DB_DSN` 中的 `tcp(...)` / `?` / `&` 会破坏 shell 解析。这是历史故障的直接产物，不要改回 `source`。

### 8.3 Compose 三件套

| 文件 | 服务 | 说明 |
|------|------|------|
| `docker-compose.yml` | mysql、redis、nginx | 开发；MySQL 与 Redis 暴露宿主端口，Nginx 反代宿主前后端 |
| `docker-compose.prod.yml` | mysql、redis、server、web | 生产；数据库不暴露端口，web 暴露 80 |
| `docker-compose.prod.bootstrap.yml` | 仅覆盖 mysql | 空库首启挂载 `init.sql` |

生产禁令：不 `source .env.prod`、不省略 `--env-file` 与 `--project-name`、不 `down -v`、不删 volume、不在有数据的库上叠加 bootstrap。

### 8.4 环境变量分区

| 分区 | 代表变量 |
|------|---------|
| 应用 | `APP_NAME`、`APP_ENV`、`SERVER_PORT`、`FRONTEND_URL` |
| 数据库 | `DB_DSN`、`DB_HOST`、`DB_PORT`、`DB_NAME`、`DB_USER`、`DB_PASSWORD`、`MYSQL_*` |
| Redis | `REDIS_HOST`、`REDIS_PORT`、`REDIS_PASSWORD`、`REDIS_DB`、`REDIS_SESSION_TTL` |
| JWT | `JWT_SECRET`、`JWT_EXPIRE_HOURS` |
| LLM | `LLM_*`、`HY3_REASONING_EFFORT_*` |
| CORS | `CORS_ALLOWED_ORIGINS`、`CORS_ALLOWED_METHODS`、`CORS_ALLOWED_HEADERS`、`CORS_ALLOW_CREDENTIALS` |
| 腾讯数据源 | `TENCENT_JOBS_*` |
| 调试 | `LOG_LEVEL`、`ENABLE_MOCK_AGENT` |
| 生产 | `COMPOSE_PROJECT_NAME` |

`VITE_API_BASE_URL` 只在构建时注入前端；开发默认 `/api`。

### 8.5 CI（`.cnb.yml`）

- PR：启动依赖 → `make test`。
- main push：`make test` → 前端 `npm ci && npm run build` → 构建并推送 server / web 镜像（SHA + latest tag）→ 同步到 GitHub 镜像。
- CI **不部署生产**；生产仍是服务器上手工 `docker compose up -d --build`。
- prod compose 使用 `build:`，不引用 CI 推送的镜像 tag，存在两套构建路径。

---

## 9. 生产部署与运维

### 9.1 标准流程

```bash
cp .env.prod.example .env.prod
# 填充所有 CHANGE_ME

make prod-check        # 只读检查，必须 PASSED
make prod-config       # compose 配置校验
make prod-backup       # 备份到 backups/mysql/

docker compose --project-name wenqing-hr-demo --env-file .env.prod \
  -f docker-compose.prod.yml up -d --build

docker compose --project-name wenqing-hr-demo --env-file .env.prod \
  -f docker-compose.prod.yml ps
```

### 9.2 空库首次初始化

```bash
docker compose --project-name wenqing-hr-demo --env-file .env.prod \
  -f docker-compose.prod.yml -f docker-compose.prod.bootstrap.yml up -d --build
```

只能在 `mysql_data` 为空时使用。`init.sql` 含 `DROP TABLE`。

### 9.3 全库替换

见 `deploy/dumps/PROD_RESTORE_FROM_LOCAL_DUMP.md`，关键点：

- 先备份，再 `stop server web`（保留 mysql / redis）。
- `DROP DATABASE` + `CREATE DATABASE utf8mb4` + `gunzip -c | mysql`。
- 验证表数 25、关键表行数、`AUTO_INCREMENT`。
- 回滚用 SQL 级恢复，不用 volume 物理回滚。

### 9.4 生产基线（dump 校验清单）

| 指标 | 值 |
|------|----|
| 表数 | 25 |
| `jobs` | 46（AUTO_INCREMENT 10047） |
| `users` | 3–4 |
| `resumes` | 8–9 |
| `semantic_profiles` | 49 |
| `semantic_profile_tags` | 737 |
| `match_analyses` | 113 |
| `resume_optimization_results` | 63 |

---

## 10. 质量现状与测试

### 10.1 测试分布（32 个 `_test.go`）

| 包 | 重点 |
|----|------|
| `internal/llm` | httptest 打桩，覆盖错误映射与超时 |
| `internal/llmjson` | JSON 抽取与解析 |
| `internal/prompt` | 名称映射与占位符渲染 |
| `internal/schema` | Schema 加载与校验 |
| `internal/agent/pipeline` | 管线状态机（最大单测文件） |
| `internal/agent/repair` | 修复流程 |
| `internal/agent` | 四类业务 Agent + Mock 回归 |
| `internal/model` | Agent 元信息派生 |
| `internal/service` | Match / Recommendation / Optimization 编排与标定 |

无测试：`repository`、`http`、`middleware`、`session`、`config`、`database`、`health`、`cmd/*`。

### 10.2 运行

```bash
cd server && go test ./...
cd server && go test ./internal/agent/pipeline -v
make test
make lint
make build
make ci
```

### 10.3 回归样例

`server/testdata/resumes/`：

- `resume_01_go_backend_strong.json`：强匹配
- `resume_02_backend_weak_insufficient.json`：弱匹配
- `resume_03_non_internet_mismatch.json`：方向不匹配

用于匹配质量标定与回归。

---

## 11. 已知不一致与风险登记

> 这些是当前代码的真实状态，不是建议。改动相关区域时请一并修正或明确规避。

| 编号 | 位置 | 问题 | 影响 |
|------|------|------|------|
| R1 | `.env.example` | 存在疑似真实格式的 `LLM_API_KEY` | 密钥泄露，应立即轮换并替换为占位符 |
| R2 | `server/internal/http/router.go` | `/api/jobs/{id}/match*` 内联复制 Auth 逻辑，未注入 `session_id` | 鉴权行为不一致，后续维护易漏改 |
| R3 | `server/internal/service/match_service.go` | legacy LLM 落库硬编码 `generation_tier=fast` 与 `generated_by=fast_sep_match_engine` | 结果元信息与实际路径不符 |
| R4 | `web/nginx.conf` + `docker-compose.prod.yml` | web 镜像不含 `/api` 反代，却直接暴露 80 | 无外部 Gateway 时生产 `/api` 不可用 |
| R5 | `.gitignore` | 未忽略 `deploy/dumps/*.sql` | 含真实业务数据的 dump 已入库 |
| R6 | 仓库根 / `server/` | 提交了 14 MB 级编译产物与 `init.sql.bak` | 仓库膨胀 |
| R7 | `Makefile` | `.PHONY` 声明了不存在的 `jobs-import-sql`；`rebuild`、`sep-backfill-resume4` 未声明 | `make jobs-import-sql` 报错 |
| R8 | `Makefile` / `README` | 基础设施描述为 MySQL + Nginx，实际含 Redis；`make status` 打印 `http://localhost/`，实际入口是 `/offer/` | 误导使用者 |
| R9 | `server/internal/middleware/cors.go` | 硬编码放行任意端口 localhost 与 `*.cnb.run` | 生产 CORS 实际比配置更宽 |
| R10 | `server/internal/agent/pipeline/pipeline.go` | Repair 使用 `context.Background()` | 修复请求不受超时约束，可能长时间挂起 |
| R11 | `sql/init.sql` / 迁移 | 无迁移框架，依赖人工执行 | 环境间 schema 漂移风险 |
| R12 | `server/README.md`、`web/README.md` | 严重过时 | 误导后续开发者 |
| R13 | `web/package.json` | React 18 运行时 + `@types/react` 19 | 类型检查错配 |
| R14 | 全局 | 无限流、无 request-id 贯穿、无 metrics / tracing | 生产可观测性与抗压能力不足 |
| R15 | `server/sql/init.sql` | 含 `DROP TABLE` | 误用于生产会清库 |
| R16 | 四类能力的 Fast 分支 | 落库策略不一致：推荐 / 匹配的 Fast 不落库，优化 / 对比的 Fast 落库 | `/latest?tier=fast` 在推荐与匹配上通常查不到结果；见 §6.5 落库矩阵 |

---

## 12. 变更规程（给 LLM 的操作手册）

### 12.1 新增一个 HTTP 接口

1. 在 `server/internal/http/` 新增 Handler 或扩展已有 Handler。
2. 在 `router.go` 注册路由；注意更具体的路径必须写在通配路径之前。
3. 需要鉴权的路由包 `middleware.Auth`；不要在 Handler 内手写 token 解析。
4. Handler 只做参数解析与错误映射，业务逻辑放 Service。
5. 统一用 `response.OK` / `response.Error` / `response.WriteAppError`。
6. 补充 `server/docs/swagger.json`（如果该接口需要对外文档）。
7. 更新 `README.md` 与本文档的路由表。

### 12.2 新增一个 AI 能力

1. 在 `server/prompts/` 新增 Prompt `.md`。
2. 在 `server/prompts/schema/` 新增 JSON Schema。
3. 在 `internal/prompt/loader.go` 的 `PromptName` 映射中注册。
4. 在 `internal/schema/loader.go` 的 `SchemaName` 映射中注册。
5. 在 `internal/agent/pipeline/types.go` 增加 `TargetOutputType` 并加入 `validTargetOutputTypes`。
6. 新增 Business Agent，实现输入组装 → `runner.Run` → 结果映射 → fallback 判定。
7. 在 Service 中编排 fast / quality 双档与降级顺序。
8. 落库时写全 `generation_tier`、`generated_by`、`profile_type`、`profile_dominant_profile`。
9. 返回 `meta`，数组字段初始化为空数组。
10. 补 Pipeline / Agent / Service 三层测试。

### 12.3 修改 Prompt 或 Schema

- 两者必须成对修改；Schema 是最终裁判。
- 修改后运行 `go test ./internal/agent/... ./internal/schema/...`。
- 检查 `prompts/_shared/frontend_contract.md` 是否仍与新的字段契约一致。
- 不要删除已有字段，前端适配层可能仍依赖；如需废弃，先加新字段并保留旧字段一个版本。

### 12.4 修改数据库表

1. 开发环境改 `server/sql/init.sql`（保持新环境可直接初始化）。
2. 同时新增 `server/sql/migrations/<date>_<name>.sql`（供已有环境升级）。
3. 更新 `internal/model/` 与 `internal/repository/`。
4. 检查所有 `SELECT *` 或固定列列表是否需要同步。
5. 更新本文档 §5。
6. 生产变更必须先备份（`make prod-backup`），再按迁移脚本执行。

### 12.5 修改前端页面或路由

1. 路由统一在 `web/src/app/routes.tsx`；需要登录的页面包 `RequireAuth`。
2. 核心 AI 页面包 `CorePageWrapper`（ErrorBoundary）。
3. 不要在组件内手写 `/offer/` 前缀。
4. 新接口调用放 `src/api/`，不要在页面里直接 `fetch`。
5. 新后端字段先加到 `src/types/`，再在 `src/adapters/` 做映射，最后页面消费。
6. 运行 `npm run typecheck`（虽然存在 React 类型错配的历史问题）。

### 12.6 修改鉴权或会话

- JWT Claims 与 Redis session key 是跨端契约，改动需要同步前端 401 处理逻辑。
- `41005` 必须保持为「异地登录」语义，前端据此提示。
- 不要移除 `middleware.Auth` 的 Redis 校验，否则顶号能力失效。

---

## 13. 改动后的验证清单

### 13.1 后端

```bash
cd server
gofmt -l .                 # 应无输出
go build ./cmd/server
go test ./...
```

手工验证：

```bash
curl -s http://localhost:8080/healthz
curl -s "http://localhost:8080/api/jobs?page=1&page_size=5"
curl -s -X POST http://localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"phone":"...","password":"..."}'
```

AI 接口验证需要 Bearer Token，且要检查返回的 `meta.generated_by` / `agent_status` 是否符合预期路径。

### 13.2 前端

```bash
cd web
npm run typecheck
npm run build
```

浏览器验证：

- `http://localhost/offer/` 能打开岗位广场。
- 登录后 `/offer/recommend`、`/offer/match`、`/offer/resume-optimization`、`/offer/resumes/compare` 可访问。
- 刷新深层路由不 404（SPA fallback）。
- Fast / Deep 切换时 UI 状态正确，降级结果有提示。

### 13.3 契约

- 所有成功响应 `code === 0`。
- 所有 AI 响应含 `meta` 且字段完整。
- 数组字段不为 null。
- `generation_tier` 只出现 `fast` / `quality`。
- 错误响应 `data === null`。

### 13.4 工程化

```bash
make lint
make test
make build
make prod-check        # 涉及生产改动时
```

---

## 14. 附录

### 14.1 关键文件索引

| 主题 | 文件 |
|------|------|
| 后端入口 | `server/cmd/server/main.go` |
| 路由 | `server/internal/http/router.go` |
| 配置 | `server/internal/config/config.go` |
| 错误码 | `server/internal/errors/error.go` |
| 响应 | `server/internal/response/response.go` |
| Auth 中间件 | `server/internal/middleware/auth.go` |
| CORS | `server/internal/middleware/cors.go` |
| Pipeline | `server/internal/agent/pipeline/pipeline.go` |
| Pipeline 类型 | `server/internal/agent/pipeline/types.go` |
| Repair | `server/internal/agent/repair/json_repair_agent.go` |
| Prompt Loader | `server/internal/prompt/loader.go` |
| Schema Loader | `server/internal/schema/loader.go` |
| LLM Client | `server/internal/llm/client.go` |
| Match 编排 | `server/internal/service/match_service.go` |
| Match Fast 引擎 | `server/internal/service/match_sep_engine.go` |
| Match Deep | `server/internal/service/match_deep_agent.go` |
| Recommendation 编排 | `server/internal/service/recommendation_service.go` |
| Deep Cascade | `server/internal/service/recommend_deep_cascade.go` |
| Optimization 编排 | `server/internal/service/resume_optimization_service.go` |
| Compare 编排 | `server/internal/service/version_compare_service.go` |
| SEP | `server/internal/service/semantic_profile_*.go` |
| TagEvidence | `server/internal/service/tag_evidence_service.go` |
| Agent 元信息 | `server/internal/model/agent_response.go` |
| 前端路由 | `web/src/app/routes.tsx` |
| 前端 API 客户端 | `web/src/api/client.ts` |
| 前端 Agent API | `web/src/api/agents.ts` |
| 前端适配层 | `web/src/api/viewModels.ts` + `web/src/adapters/` |
| 前端鉴权 | `web/src/stores/authStore.tsx` |
| 数据库 DDL | `server/sql/init.sql` |
| 生产 compose | `docker-compose.prod.yml` |
| 生产检查 | `scripts/deploy/prod-check.sh` |

### 14.2 常用调试命令

```bash
make status
make health
make logs-backend
make logs-agent
make db-check
make db-shell
make llm-check
RUN_REAL_LLM_SMOKE=1 make llm-smoke
make sep-backfill-dry-run
```

### 14.3 判断一次 AI 请求走了哪条路径

按顺序看：

1. 请求体 `generation_tier`。
2. 响应 `data.meta.generated_by`。
3. 响应 `data.meta.agent_status`。
4. 响应 `data.meta.fallback_reason` / `error_code`。
5. 后端日志中的 `[LLMRequest]`、`[DeepCascade]`、`[SEP-Profiler]`、`[OptimizeV2]` 标记。

常见组合：

| 组合 | 含义 |
|------|------|
| `fast_tag_engine` + `success` | Fast 档规则引擎正常 |
| `fast_sep_match_engine` + `success` | Fast 档 SEP 匹配正常 |
| `llm_agent` + `success` | 单 Agent LLM 首次通过 |
| `llm_agent` + `repaired` | LLM 输出被 Repair 修复后通过 |
| `hy3_screening_match_agent` + `success` | Hy3 五模块筛查成功 |
| `hy3_screening_match_agent_degraded` + `fallback` | 部分模块回落规则 |
| `deep_fallback_agent` + `fallback` | Deep 链路失败，回落 |
| `mock_agent` + `mock_success` | 最终 Mock 兜底 |
| `mock_quality` + `fallback` | 质量档 Mock（legacy 标记）；前端会按 Mock 报错 |

---

**维护约定**：本文件与 `README.md` 是活文档。任何涉及路由、表结构、Agent 契约、环境变量、启动方式、部署流程的改动，都必须在同一个提交内同步更新。
