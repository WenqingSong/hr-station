# 阶段3核心业务后端Mock方案

> 最后更新：2026-05-29（阶段 3-9 阶段 3 汇总）

## 1. 阶段目标

阶段 3 的核心目标是：**在不接真实模型 API 的前提下，先用后端 Mock 数据跑通 Offer 捕手的核心业务闭环。**

阶段 3 完成后，后端应支持以下主链路：

```
用户注册 / 登录
→ 添加简历版本
→ 查看当前简历
→ 查看岗位列表 / 岗位详情
→ 获取岗位推荐 Mock
→ 进行岗位匹配分析 Mock
→ 添加新简历版本
→ 查看版本对比 Mock
```

阶段 3 不是 Agent 接入阶段，也不是前端接后端阶段。

阶段 3 必须坚持：

1. 推荐、匹配、版本对比先使用 Mock JSON
2. 不调用真实 LLM
3. 不接真实 Agent API
4. 不自动生成优化版简历
5. 不做多简历管理
6. 不上传 PDF / DOC
7. 不做岗位收藏
8. 不做投递系统
9. 不做 HR 后台
10. 不做复杂权限系统

## 2. 当前基础

阶段 2 已完成后端基础设施建设：

| 基础设施 | 状态 | 说明 |
|----------|------|------|
| 配置读取 | ✅ 已完成 | App/Server/Database/JWT/LLM/Feature/Log 七组 |
| 统一响应结构 | ✅ 已完成 | `{code, message, data}` 格式 |
| 基础错误处理 | ✅ 已完成 | AppError + 基础错误码 |
| 中间件链 | ✅ 已完成 | Recover → CORS → RequestLogger |
| MySQL 连接 | ✅ 已完成 | database/sql + MySQL Driver |
| 健康检查 | ✅ 已完成 | app/database 状态聚合 |
| **数据库表结构** | ✅ 已完成 | 6 张表 + 种子数据（3-2） |

阶段 2 未实现的（阶段 3 要做的）：

| 能力 | 状态 | 阶段 |
|------|------|------|
| ~~业务表结构~~ | ✅ 已完成 | 阶段 3-2 |
| ~~用户注册登录 / JWT~~ | ✅ 已完成 | 阶段 3-3 |
| 业务接口 | ❌ 未实现 | 阶段 3 |
| Mock Agent | ❌ 未实现 | 阶段 3 |
| 真实 LLM 接入 | ❌ 未实现 | 阶段 4 |
| 前端接后端 API | ❌ 未实现 | 阶段 5 |

## 3. 阶段 3 任务拆分

### 3-1 创建阶段 3 核心业务后端 Mock 方案 ✅

| 项目 | 说明 |
|------|------|
| 目标 | 规划阶段 3 整体方案，更新所有相关文档 |
| 允许修改 | `docs/` 目录内 Markdown 文档 |
| 禁止事项 | 不修改 Go 代码、不修改 server/、不修改 Makefile、不修改 .env.example、不修改 docker-compose.yml、不写 SQL |
| 验收标准 | 阶段 3 方案文档创建完成，所有相关文档已同步更新 |
| 是否改工程文件 | ❌ 否 |
| 是否需要更新文档 | ✅ 是 |
| 是否依赖前置任务 | ❌ 否 |
| 完成日期 | 2026-05-29 |

### 3-2 设计并写入 server/sql/init.sql 表结构与种子数据 ✅

| 项目 | 说明 |
|------|------|
| 目标 | 编写阶段 3 所需的 6 张数据表结构和种子数据 |
| 允许修改 | `server/sql/init.sql`、`docs/` 目录内文档 |
| 禁止事项 | 不修改 Go 代码、不实现接口、不修改 Makefile、不修改 .env.example |
| 验收标准 | init.sql 包含 6 张表结构和种子数据，可通过 MySQL 导入成功 |
| 是否改工程文件 | ✅ 是（修改 server/sql/init.sql） |
| 是否需要更新文档 | ✅ 是 |
| 是否依赖前置任务 | ✅ 依赖 3-1 |
| 完成日期 | 2026-05-29 |

### 3-3 实现用户注册 / 登录基础接口 ✅

| 项目 | 说明 |
|------|------|
| 目标 | 实现 POST /api/auth/register、POST /api/auth/login、GET /api/auth/me，使用最小 JWT |
| 允许修改 | `server/internal/model/user.go`、`server/internal/repository/user_repository.go`、`server/internal/service/auth_service.go`、`server/internal/http/auth_handler.go`、`server/internal/http/router.go`、`server/cmd/server/main.go`、`docs/` |
| 禁止事项 | 不做刷新 token、不做角色权限、不做邮箱验证码、不做找回密码、不引入 Casbin |
| 验收标准 | 注册 / 登录 / me 接口可用，JWT 签发和验证正常，使用统一响应结构 |
| 是否改工程文件 | ✅ 是（新增 model / repository / service / handler 文件） |
| 是否需要更新文档 | ✅ 是 |
| 是否依赖前置任务 | ✅ 依赖 3-2 |
| 完成日期 | 2026-05-29 |

### 3-4 实现简历版本接口 ✅

| 项目 | 说明 |
|------|------|
| 目标 | 实现 POST /api/resumes、GET /api/resumes/current、GET /api/resumes/detail、GET /api/resumes/versions、GET /api/resumes/versions/:versionNo |
| 允许修改 | `server/internal/model/resume.go`、`server/internal/repository/resume_repository.go`、`server/internal/service/resume_service.go`、`server/internal/http/resume_handler.go`、`server/internal/http/router.go`、`docs/` |
| 禁止事项 | 不做多简历管理、不做文件上传、不做 PDF 解析、不自动生成优化版简历 |
| 验收标准 | 简历版本接口可用，用户只有一份主简历，新增简历即新增版本，历史版本只读 |
| 是否改工程文件 | ✅ 是 |
| 是否需要更新文档 | ✅ 是 |
| 是否依赖前置任务 | ✅ 依赖 3-3（需要 JWT 认证） |
| 完成日期 | 2026-05-29 |

### 3-5 实现岗位列表与岗位详情接口 ✅

| 项目 | 说明 |
|------|------|
| 目标 | 实现 GET /api/jobs、GET /api/jobs/:id |
| 允许修改 | `server/internal/model/job.go`、`server/internal/repository/job_repository.go`、`server/internal/service/job_service.go`、`server/internal/http/job_handler.go`、`server/internal/http/router.go`、`docs/` |
| 禁止事项 | 不做收藏、不做投递、不做复杂筛选、不做真实岗位爬取 |
| 验收标准 | 岗位列表和详情接口可用，从 jobs 表读取种子数据，支持简单 keyword/city/category 查询 |
| 是否改工程文件 | ✅ 是 |
| 是否需要更新文档 | ✅ 是 |
| 是否依赖前置任务 | ✅ 依赖 3-3（需要 JWT 认证） |
| 完成日期 | 2026-05-29 |

### 3-6 实现岗位推荐 Mock 接口 ✅

| 项目 | 说明 |
|------|------|
| 目标 | 实现 POST /api/jobs/recommend，返回 Mock 推荐结果 |
| 允许修改 | `server/internal/agent/recommendation_mock.go`、`server/internal/model/recommendation.go`、`server/internal/repository/recommendation_repository.go`、`server/internal/service/recommendation_service.go`、`server/internal/http/recommendation_handler.go`、`server/internal/http/router.go`、`docs/` |
| 禁止事项 | 不接真实 LLM、不读取 LLM_API_KEY、不使用真实 Prompt |
| 验收标准 | 推荐接口可用，返回结构化 Mock 结果（推荐岗位 / 推荐理由 / 匹配亮点 / 风险点 / 优先级） |
| 是否改工程文件 | ✅ 是 |
| 是否需要更新文档 | ✅ 是 |
| 是否依赖前置任务 | ✅ 依赖 3-4（需要简历数据）和 3-5（需要岗位数据） |
| 完成日期 | 2026-05-29 |

### 3-7 实现岗位匹配分析 Mock 接口 ✅

| 项目 | 说明 |
|------|------|
| 目标 | 实现 POST /api/jobs/:id/match，返回 Mock 匹配分析结果（核心亮点） |
| 允许修改 | `server/internal/agent/match_mock.go`、`server/internal/model/match.go`、`server/internal/repository/match_repository.go`、`server/internal/service/match_service.go`、`server/internal/http/match_handler.go`、`server/internal/http/router.go`、`docs/` |
| 禁止事项 | 不接真实 LLM、不自动生成优化版简历、不读取 LLM_API_KEY |
| 验收标准 | 匹配分析接口可用，返回结构化 Mock 结果（匹配分数 / 初筛判断 / JD-简历证据链 / 风险点 / P0-P1 优化建议 / 下一步建议 / 真实性提示） |
| 是否改工程文件 | ✅ 是 |
| 是否需要更新文档 | ✅ 是 |
| 是否依赖前置任务 | ✅ 依赖 3-4（需要简历数据）和 3-5（需要岗位数据） |
| 完成日期 | 2026-05-29 |

### 3-8 实现简历版本对比 Mock 接口 ✅

| 项目 | 说明 |
|------|------|
| 目标 | 实现 GET /api/resumes/versions/compare?from=1&to=2，返回 Mock 版本对比结果 |
| 允许修改 | `server/internal/agent/version_compare_mock.go`、`server/internal/model/version_compare.go`、`server/internal/repository/version_compare_repository.go`、`server/internal/service/version_compare_service.go`、`server/internal/http/version_compare_handler.go`、`server/internal/http/router.go`、`docs/` |
| 禁止事项 | 不接真实 LLM、不自动修改简历 |
| 验收标准 | 版本对比接口可用，返回结构化 Mock 结果（分数变化 / 字段变化 / 风险变化 / 建议完成情况 / 总结） |
| 是否改工程文件 | ✅ 是 |
| 是否需要更新文档 | ✅ 是 |
| 是否依赖前置任务 | ✅ 依赖 3-4（需要简历版本数据） |
| 完成日期 | 2026-05-29 |

### 3-9 阶段 3 汇总 ✅

| 项目 | 说明 |
|------|------|
| 目标 | 汇总阶段 3 成果，确认可进入阶段 4 |
| 允许修改 | `docs/` 目录内 Markdown 文档 |
| 禁止事项 | 不修改 Go 代码 |
| 验收标准 | 阶段 3 所有验收标准通过 |
| 是否改工程文件 | ❌ 否 |
| 是否需要更新文档 | ✅ 是 |
| 是否依赖前置任务 | ✅ 依赖 3-1 到 3-8 全部完成 |
| 完成日期 | 2026-05-29 |

## 阶段 3 完成结论

**✅ 阶段 3 已全部完成，可以进入阶段 4：前端接入后端 API 与演示链路联调。**

阶段 3 全部 9 个子任务已完成，后端已具备完整 Mock 业务闭环能力，所有接口可通过 curl 或 API 工具跑通主链路。

### 交付物完成情况

| 交付物 | 状态 |
|--------|------|
| 6 张数据库表 + 种子数据 | ✅ 已完成 |
| 14 个 HTTP 接口 | ✅ 已完成 |
| 3 个 Mock Agent | ✅ 已完成 |
| JWT 认证 + 统一响应 | ✅ 已完成 |
| 完整文档体系 | ✅ 已完成 |

### 阶段 3 验收结果

所有 16 项验收标准已通过（详见 `docs/00-项目总览/阶段3汇总.md`）。

### 阶段 4 入口说明

阶段 4 目标：前端接入后端 API 与演示链路联调。详见 `docs/下一步任务.md`。

## 4. 数据库表范围规划

阶段 3 计划创建 6 张表，已写入 `server/sql/init.sql`（3-2 ✅ 已完成）。

### 4.1 users

用于用户注册、登录、基础身份识别。

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT UNSIGNED AUTO_INCREMENT | 主键 |
| email | VARCHAR(255) UNIQUE | 登录邮箱 |
| password_hash | VARCHAR(255) | 密码哈希 |
| name | VARCHAR(100) | 用户名 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

不做：多角色 / RBAC / Casbin / 邮箱验证 / 短信验证 / 找回密码 / 刷新 token

### 4.2 resume_versions

用于单主简历的版本化管理。

**关键设计决策**：用户只有一份主简历，不设计多简历 `resumes` 表。每次添加简历就是新增一个版本。历史版本只读。

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT UNSIGNED AUTO_INCREMENT | 主键 |
| user_id | BIGINT UNSIGNED | 用户 ID（外键） |
| version_no | INT | 版本号（1, 2, 3...） |
| name | VARCHAR(100) | 姓名 |
| phone | VARCHAR(50) | 电话 |
| email | VARCHAR(255) | 邮箱 |
| location | VARCHAR(100) | 所在地 |
| target_city | VARCHAR(100) | 目标城市 |
| target_role | VARCHAR(100) | 目标岗位 |
| education_json | JSON | 教育经历（数组） |
| experience_json | JSON | 工作经历（数组） |
| project_json | JSON | 项目经历（数组） |
| skills_json | JSON | 技能列表 |
| self_evaluation | TEXT | 自我评价 |
| raw_text | TEXT | 原始文本（可选） |
| created_at | DATETIME | 创建时间 |

说明：
1. 所有字段可选
2. 结构复杂字段先用 JSON 存储
3. 阶段 3 不做文件上传
4. 阶段 3 不做 PDF 解析
5. 阶段 3 不自动生成优化版简历
6. 不单独建 education_experiences / work_experiences / project_experiences 表，先用 JSON 简化

### 4.3 jobs

用于岗位列表和岗位详情。

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT UNSIGNED AUTO_INCREMENT | 主键 |
| title | VARCHAR(200) | 岗位名称 |
| department | VARCHAR(200) | 部门 |
| city | VARCHAR(100) | 城市 |
| category | VARCHAR(100) | 方向分类 |
| job_type | VARCHAR(100) | 岗位类型（校招/实习/社招） |
| description | TEXT | 岗位描述 |
| requirements | TEXT | 岗位要求 |
| skill_keywords_json | JSON | 技能关键词 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

### 4.4 recommendation_results

用于保存岗位推荐 Mock 结果快照。

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT UNSIGNED AUTO_INCREMENT | 主键 |
| user_id | BIGINT UNSIGNED | 用户 ID |
| resume_version_id | BIGINT UNSIGNED | 简历版本 ID |
| result_json | JSON | 推荐结果 JSON |
| created_at | DATETIME | 创建时间 |

### 4.5 match_analyses

用于保存岗位匹配分析 Mock 结果。

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT UNSIGNED AUTO_INCREMENT | 主键 |
| user_id | BIGINT UNSIGNED | 用户 ID |
| resume_version_id | BIGINT UNSIGNED | 简历版本 ID |
| job_id | BIGINT UNSIGNED | 岗位 ID |
| score | INT | 匹配分数（0-100） |
| result_json | JSON | 分析结果 JSON |
| created_at | DATETIME | 创建时间 |

### 4.6 version_compare_results

用于保存简历版本对比 Mock 结果。

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT UNSIGNED AUTO_INCREMENT | 主键 |
| user_id | BIGINT UNSIGNED | 用户 ID |
| from_version_id | BIGINT UNSIGNED | 旧版本 ID |
| to_version_id | BIGINT UNSIGNED | 新版本 ID |
| result_json | JSON | 对比结果 JSON |
| created_at | DATETIME | 创建时间 |

### 4.7 与阶段 2 规划的差异说明

| 调整项 | 原规划 | 阶段 3 实际 | 原因 |
|--------|--------|------------|------|
| 简历主表 | `resumes` + `resume_versions` | 仅 `resume_versions` | 用户只有一份主简历，无需 resumes 表 |
| 经历子表 | 3 张独立表 | JSON 字段 | Demo 阶段简化，降低 SQL 复杂度 |
| 推荐结果 | `job_recommendations` | `recommendation_results` | 更明确的命名 |
| 匹配证据 | `match_evidences` 独立表 | 嵌入 `match_analyses.result_json` | 证据链作为分析结果的一部分存储 |
| 版本对比 | 无 | `version_compare_results` | 新增，支持版本迭代闭环 |

## 5. 后端目录扩展规划

阶段 3 逐步扩展后端目录。3-2 已完成 SQL，3-3 起开始创建 Go 文件。

```
server/internal/model/
  user.go                    # 用户数据结构（3-3）
  resume.go                  # 简历数据结构（3-4）
  job.go                     # 岗位数据结构（3-5）
  recommendation.go          # 推荐结果数据结构（3-6）
  match.go                   # 匹配分析数据结构（3-7）
  version_compare.go         # 版本对比数据结构（3-8）

server/internal/repository/
  user_repository.go         # 用户数据访问（3-3）
  resume_repository.go       # 简历数据访问（3-4）
  job_repository.go          # 岗位数据访问（3-5）
  recommendation_repository.go  # 推荐结果数据访问（3-6）
  match_repository.go        # 匹配分析数据访问（3-7）
  version_compare_repository.go # 版本对比数据访问（3-8）

server/internal/service/
  auth_service.go            # 认证业务（3-3）
  resume_service.go          # 简历业务（3-4）
  job_service.go             # 岗位业务（3-5）
  recommendation_service.go  # 推荐业务（3-6）
  match_service.go           # 匹配分析业务（3-7）
  version_compare_service.go # 版本对比业务（3-8）

server/internal/http/
  auth_handler.go            # 认证 HTTP 处理（3-3）
  resume_handler.go          # 简历 HTTP 处理（3-4）
  job_handler.go             # 岗位 HTTP 处理（3-5）
  recommendation_handler.go  # 推荐 HTTP 处理（3-6）
  match_handler.go            # 匹配 HTTP 处理（3-7）
  version_compare_handler.go # 版本对比 HTTP 处理（3-8）
  router.go                  # 路由注册（已有，3-3 起扩展）

server/internal/agent/
  recommendation_mock.go     # 推荐 Mock Agent（3-6）
  match_mock.go              # 匹配 Mock Agent（3-7）
  version_compare_mock.go    # 版本对比 Mock Agent（3-8）

server/internal/middleware/
  auth.go                    # JWT Auth 中间件（3-3）
```

### 分层职责

| 层 | 职责 | 说明 |
|----|------|------|
| Model | 数据结构定义 | 对应数据库表，定义 Go struct |
| Repository | 数据访问 | 封装 SQL 查询，操作数据库 |
| Service | 业务编排 | 编排 Repository 和 Agent，处理业务逻辑 |
| HTTP Handler | HTTP 入参出参与路由处理 | 解析请求、调用 Service、格式化响应 |
| Agent Mock | Mock 数据生成 | 生成推荐、匹配、版本对比的结构化结果 |

## 6. 接口范围规划

### 6.1 认证接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/auth/register` | 用户注册 |
| POST | `/api/auth/login` | 用户登录 |
| GET | `/api/auth/me` | 获取当前用户信息 |

### 6.2 简历接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/resumes` | 新增简历版本 |
| GET | `/api/resumes/current` | 获取当前（最新版本）简历 |
| GET | `/api/resumes/detail` | 获取简历详情 |
| GET | `/api/resumes/versions` | 获取版本列表 |
| GET | `/api/resumes/versions/:versionNo` | 获取指定版本 |

### 6.3 岗位接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/jobs` | 岗位列表 |
| GET | `/api/jobs/:id` | 岗位详情 |

### 6.4 岗位推荐 Mock 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/jobs/recommend` | 获取推荐岗位 Mock |

### 6.5 岗位匹配分析 Mock 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/jobs/:id/match` | 岗位匹配分析 Mock |

### 6.6 简历版本对比 Mock 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/resumes/versions/compare?from=1&to=2` | 版本对比 Mock |

### 6.7 API 路径前缀说明

阶段 3 统一使用 `/api/` 前缀（不带 `/v1/`），简化 Demo 阶段路径。

## 7. Mock Agent 边界

阶段 3 使用 Mock Agent，不调用真实 LLM。

### 7.1 岗位推荐 Mock

**输入**：当前简历版本 + 岗位列表

**输出**：推荐岗位列表 + 推荐理由 + 匹配亮点 + 风险点 + 优先级

### 7.2 岗位匹配 Mock

**输入**：当前简历版本 + 目标岗位 JD

**输出**：匹配分数 + 初筛判断 + JD-简历证据链 + 风险点 + P0/P1 优化建议 + 下一步建议 + 真实性提示

### 7.3 版本对比 Mock

**输入**：旧版本简历 + 新版本简历

**输出**：字段变化 + 能力表达变化 + 风险变化 + 建议完成情况 + 总结

### 7.4 Mock 实现方式

1. 阶段 3 Mock Agent 可以是固定 JSON
2. 也可以是简单规则生成
3. 不请求外部模型
4. 不读取 LLM_API_KEY
5. 不使用真实 Prompt
6. 阶段 4 再替换为真实 Agent / LLM

## 8. 核心业务链路

阶段 3 完成后，后端主链路应可通过 curl 或 API 工具跑通：

```
1. POST /api/auth/register        → 注册用户
2. POST /api/auth/login           → 登录获取 JWT
3. POST /api/resumes              → 添加简历 V1
4. GET  /api/resumes/current       → 查看当前简历
5. GET  /api/jobs                  → 查看岗位列表
6. GET  /api/jobs/:id              → 查看岗位详情
7. POST /api/jobs/recommend        → 获取岗位推荐 Mock
8. POST /api/jobs/:id/match        → 岗位匹配分析 Mock
9. POST /api/resumes              → 添加简历 V2
10. GET /api/resumes/versions/compare?from=1&to=2  → 版本对比 Mock
```

## 9. 阶段 3 不做事项

| 不做事项 | 说明 |
|---------|------|
| 不接真实 LLM | 阶段 3 使用 Mock Agent，阶段 4 接真实模型 |
| 不接真实 Agent API | 不调用外部 Agent 服务 |
| 不自动生成优化版简历 | 只给诊断和建议，用户手动修改 |
| 不上传 PDF / DOC | 只支持结构化表单录入 |
| 不做多简历管理 | 用户只有一份主简历，通过版本迭代 |
| 不做岗位收藏 | Demo 阶段岗位数据量有限 |
| 不做投递追踪 | 本项目不是投递系统 |
| 不做面试模拟 | 超出当前产品范围 |
| 不做 HR 后台 | 面向求职者，不做双边平台 |
| 不做复杂权限系统 | 单用户场景，无需权限体系 |
| 不做 RBAC / Casbin | Demo 不需要角色权限 |
| 不做刷新 token | 阶段 3 只做最小 JWT |
| 不做邮箱验证 | Demo 简化认证 |
| 不做真实岗位爬取 | 使用种子数据 |
| 不做前端接后端 | 阶段 5 才让前端正式接后端 |
| 不修改前端页面业务逻辑 | 阶段 3 只做后端 |

## 10. 阶段 3 验收标准

| 序号 | 验收标准 |
|------|---------|
| 1 | ✅ `server/sql/init.sql` 已包含表结构和种子数据 |
| 2 | 用户注册 / 登录 / me 接口可用 |
| 3 | 简历版本接口可用 |
| 4 | 岗位列表 / 详情接口可用 |
| 5 | 岗位推荐 Mock 接口可用 |
| 6 | 岗位匹配分析 Mock 接口可用 |
| 7 | 简历版本对比 Mock 接口可用 |
| 8 | 所有接口使用统一响应结构 |
| 9 | 需要登录的接口具备基础鉴权 |
| 10 | 推荐、匹配、版本对比不调用真实模型 |
| 11 | 后端主链路可通过 curl 或 API 工具跑通 |
| 12 | 文档、变更记录、开发日志完整 |
| 13 | 可以进入阶段 4：Agent 与真实模型接入，或阶段 5：前端接后端 |

## 11. 阶段 4 / 阶段 5 入口条件

### 进入阶段 4 的条件

1. 阶段 3 全部验收标准通过
2. Mock 接口已跑通核心链路
3. 接口结构稳定，不需要大幅调整
4. 准备好 LLM API Key 和模型配置

### 进入阶段 5 的条件

1. 阶段 3 或阶段 4 已完成
2. 后端接口稳定可用
3. API 文档完善
4. 前端路由和 API Client 准备好接入

### 并行可能

阶段 3 完成后，可以：
- 先进入阶段 4（替换 Mock 为真实 LLM）
- 或先进入阶段 5（前端接后端 Mock 接口）
- 两个阶段可以一定程度并行

## 关联文档

- [后端架构.md](./后端架构.md)
- [目录结构.md](./目录结构.md)
- [服务分层.md](./服务分层.md)
- [阶段2后端基础设施方案.md](./阶段2后端基础设施方案.md)
- [../03-数据库设计/表结构设计.md](../03-数据库设计/表结构设计.md)
- [../04-接口文档/接口总览.md](../04-接口文档/接口总览.md)
- [../05-Agent设计/Agent总览.md](../05-Agent设计/Agent总览.md)
