# 第 2 阶段分析：后端架构与 Agent 系统设计

> 分析时间：2026-06-23
> 源文档：`02-后端设计/`、`04-后端设计/`、`05-Agent设计/`(35文件)、`03-数据库设计/`、`04-接口文档/`、`17-Agent架构冻结/`、`架构设计/`

---

## 一、后端架构全景

### 1.1 技术选型

| 技术 | 选择 | 理由 |
|------|------|------|
| 语言 | Go | 高性能、编译型、适合 API 服务 |
| HTTP 框架 | **Go 标准库 `net/http`** | 不引入 Gin/Echo/Fiber，保持最小依赖 |
| 数据库驱动 | `database/sql` + `go-sql-driver/mysql` | 不引入 GORM/sqlx，裸 SQL + 标准库 |
| 数据库 | MySQL 8.0 | 结构化简历数据，Docker 容器化 |
| 缓存 | Redis 7（预留，阶段 7） | 阶段 1-6 不引入 |

**关键决策**：全项目不引入第三方 Web 框架和 ORM，使用 Go 标准库 + 轻量驱动，极致精简依赖。

### 1.2 分层架构

```
Handler（HTTP 请求处理）  ← 6 个 handler 文件
    ↓
Service（业务逻辑编排）   ← 6 个 service 文件
    ↓                       ↓
Repository（数据访问）   Agent Pipeline（LLM 调用）
    ↓                       ↓
Model（数据模型）        LLM Client 层
    ↓
MySQL 数据库
```

| 层 | 职责 | 文件数 | 关键约束 |
|----|------|--------|---------|
| Handler | HTTP 解析、参数校验、响应格式化 | 6 | 不堆业务逻辑 |
| Service | 业务编排、Agent 调用、fallback 决策 | 6 | 依赖 Repository + Pipeline |
| Repository | 数据库 CRUD，裸 SQL | 6 | 不包含业务逻辑 |
| Model | 数据结构，对应 DB 表 | 6+ | Go struct + db tag |
| Middleware | Recover / CORS / Logger / Auth | 4 | 链式执行 |
| Agent Pipeline | Prompt→LLM→parse→validate→repair→fallback | 5+ | 统一模型调用入口 |
| LLM Client | 抽象接口 + HunyuanClient + NoopClient | 3 | 业务 Agent 不直接依赖 Hunyuan |

### 1.3 中间件链

```
请求 → Recover（panic 捕获，返回 50000）
     → CORS（环境变量配置化，CORS_ALLOWED_ORIGINS）
     → RequestLogger（method/path/status/duration/remote/user_agent）
     → Auth（JWT Bearer Token 验证，注入 user_id）
     → Handler
```

### 1.4 配置体系

7 组结构化配置：

| 配置组 | 字段数 | 说明 |
|--------|--------|------|
| App | 2 | 应用名称、运行环境 |
| Server | 1 | HTTP 端口 |
| Database | 6 | MySQL 连接 + DSN |
| JWT | 2 | Secret + 过期时间 |
| LLM | 14 | 核心 LLM 配置（Provider/Key/Model/Temp/Timeout/Repair/Fallback 等） |
| Feature | 多个 | 功能开关（EnableMockAgent, EnableLLMAgent 等） |
| Log | 1 | 日志级别 |

### 1.5 统一响应结构

```json
// 成功
{ "code": 0, "message": "ok", "data": { ... } }

// 错误
{ "code": 40001, "message": "invalid parameter", "data": null }
```

错误码体系：0(成功) / 40000(请求错误) / 40001(参数错误) / 40100(未认证) / 40300(无权限) / 40400(资源不存在) / 50000(内部错误) / 50300(依赖不可用)

### 1.6 后端接口清单（14 个）

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | `/healthz` | 健康检查（app + database 状态） | ❌ |
| POST | `/api/auth/register` | 用户注册 | ❌ |
| POST | `/api/auth/login` | 用户登录 | ❌ |
| GET | `/api/auth/me` | 获取当前用户 | ✅ |
| POST | `/api/resumes` | 新增简历版本 | ✅ |
| GET | `/api/resumes/current` | 获取当前简历 | ✅ |
| GET | `/api/resumes/detail` | 获取简历详情 | ✅ |
| GET | `/api/resumes/versions` | 获取版本列表 | ✅ |
| GET | `/api/resumes/versions/{versionNo}` | 获取指定版本 | ✅ |
| GET | `/api/jobs` | 岗位列表（支持 external_source 筛选） | ✅ |
| GET | `/api/jobs/{id}` | 岗位详情 | ✅ |
| POST | `/api/jobs/recommend` | 岗位推荐（LLM + Mock fallback） | ✅ |
| POST | `/api/jobs/{id}/match` | 岗位匹配分析（LLM + Mock fallback） | ✅ |
| GET/POST | `/api/resumes/versions/compare` | 版本对比 | ✅ |

---

## 二、Agent 系统深度设计

### 2.1 四主 Agent 架构

```
Offer Catcher 四主 Agent 架构

1. Resume Version Compare Agent     ← 简历版本对比（P1：唯一未 LLM 化）
2. Job Recommendation Agent         ← 岗位智能推荐（Fast/Deep 双阶段）
3. Job Match Analysis Agent    ⭐   ← 岗位匹配分析（核心亮点）
4. Resume Optimization Agent        ← 简历优化建议（3 子 Agent 拆分）
```

### 2.2 Agent 设计原则

1. **先 Mock 后接模型**：阶段 3 Mock Agent 跑通闭环，阶段 4 接真实 LLM
2. **统一接口**：所有 Agent 遵循统一输入输出接口，Service 层通过 `pipelineRunner` 调用
3. **Prompt 独立管理**：模板在 `server/prompts/` 目录，作为产品资产独立版本控制
4. **真实性保护**：Agent 不虚构经历、不替代编造内容、不自动生成简历
5. **失败兜底**：LLM 调用失败自动回退 Mock Agent，不阻断用户操作
6. **业务 Agent 不直接依赖 Hunyuan**：所有模型调用通过 Pipeline Runner 完成

### 2.3 Agent 状态矩阵

| 核心功能 | 页面 | 接口 | Mock | LLM Agent | Prompt | Schema | 状态 |
|---------|------|------|------|-----------|--------|--------|------|
| 版本对比 | ✅ | ✅ | ✅ | ❌ 未 LLM 化 | ❌ | ❌ | P1 待做 |
| 岗位推荐 | ✅ | ✅ | ✅ | ✅ repaired | ✅ | ✅ | ✅ 完成 |
| 匹配分析 | ✅ | ✅ | ✅ | ✅ success | ✅ | ✅ | ✅ 完成 |
| 简历优化 | ✅ | ✅ | ✅ | ✅ fallback | ✅ | ✅ | ⚠️ LLM 不稳定 |

### 2.4 Prompt 资产体系

| Prompt 名称 | 文件 | 用途 | 首轮验收 |
|------------|------|------|---------|
| match_analysis | `match_analysis_v1.md` | HR初筛+技术面试官双视角匹配分析 | partial_pass |
| recommendation | `recommendation_v1.md` | 四档推荐+Offer捕手语气 | partial_pass |
| resume_optimization | `resume_optimization_v1.md` | P0/P1/P2优化+关键词对齐 | partial_pass |
| json_repair | `json_repair_v1.md` | JSON结构修复（不重做业务分析） | N/A |

**Prompt 通用结构**：
```
1. 角色定义（你是一个简历分析助手）
2. 严格约束（只输出JSON/不虚构经历/不自动生成简历）
3. 输出格式（完整 JSON Schema 草案）
4. 输入数据（业务上下文占位符）
5. 分析要求（具体任务描述 + 证据链构建规则）
```

**Prompt 演化路径**：
- 4-1.4：人工定稿 → CodeBuddy 登记 → 首轮输出验收（partial_pass）
- 4-3.5.C：新增 apply_readiness / optimization_stop_signal / top_fix_items
- 4-3.5.D：新增 jd_evidence_chain / evidence_strength / START 检查规则
- 4-3.5.F：新增 candidate_case_type 10 种极端情况处理规则
- 4-3.7.C.1：三个 Schema 最小一致性修复，补齐战略字段
- JM-R2：Deep 5 模块 Hy3 Prompt 专向设计

### 2.5 Schema 资产体系

| Schema 名称 | 文件 | Required 字段数 | 约束 |
|------------|------|:---:|------|
| match_analysis_schema | `match_analysis_schema.json` | 17 | 含 apply_readiness + optimization_stop_signal + top_fix_items |
| recommendation_schema | `recommendation_schema.json` | 7 | 四档推荐固定枚举 + recommend_level 与 label 一致性 |
| resume_optimization_schema | `resume_optimization_schema.json` | 15 | 含 P0/P1/P2 优先级 + keyword_alignment + rewrite_examples |

**Schema 演化**：
- 全部使用 JSON Schema draft 2020-12
- Schema 禁止 `additionalProperties`
- 首轮验收问题已全部被 Schema 吸收（rewrite_example 禁止编造数字、evidence_chain 支持 missing 等）

---

## 三、LLM Pipeline：核心安全链路

### 3.1 完整调用链路

```
业务 Service 构造 PromptValues
    ↓
Pipeline Runner.Run()
    ↓
Prompt Loader（加载 .md 模板 + {{占位符}} 替换）
    ↓
Schema Loader（加载 JSON Schema）
    ↓
llm.Client.Chat（HunyuanClient / NoopClient）
    ↓
llmjson.ExtractAndParse（去除 Markdown fence + BOM 清理 + JSON parse）
    ↓
Schema Validator（santhosh-tekuri/jsonschema/v5, draft 2020-12）
    ↓
    ├── 通过 → AgentStatus = "success", generated_by = "llm_agent"
    ├── 失败 → JSONRepairAgent.Repair()（仅一次，非递归）
    │       ↓
    │       ├── 修复成功 → AgentStatus = "repaired", generated_by = "llm_agent"
    │       └── 修复失败 → AgentStatus = "fallback_required"
    └── fallback_required → Service 层调用对应 Mock Agent
                           → generated_by = "mock_agent"
```

### 3.2 Pipeline 关键约束

| 约束 | 说明 |
|------|------|
| JSON repair 只尝试一次 | `LLM_REPAIR_MAX_ATTEMPTS = 1`，禁止递归修复 |
| Repair 不改变 generated_by | 修复成功仍标记为 `llm_agent` |
| Pipeline 只返回 fallback 决策 | 不直接调用 Mock Agent（4-3 层 Service 处理） |
| 禁止的 generated_by 值 | `llm_agent_fallback`、`json_repair_agent` 均禁止 |
| API Key 保护 | Go struct `json:"-"` + `MaskedAPIKey()` + `safeErr()` + `redactSensitive()` |
| LLM_ENABLE 默认 false | 不配置真实 Key 时系统自动走 Mock |
| finish_reason=length | 不视为硬错误，返回 ChatResponse，Pipeline 决定是否 repair |
| 不在错误中泄露 Key | 所有错误处理经过脱敏 |

### 3.3 20+ 错误码覆盖

| 错误码类别 | 直接 Mock | 先 Repair |
|-----------|:---:|:---:|
| LLM_DISABLED / CONFIG_MISSING / API_KEY_MISSING | ✅ | |
| LLM_TIMEOUT / HTTP_ERROR / RATE_LIMIT / EMPTY_RESPONSE | ✅ | |
| JSON_PARSE_FAILED / MARKDOWN_WRAPPED | | ✅ |
| SCHEMA_VALIDATION_FAILED / REQUIRED_FIELD_MISSING | | ✅ |
| JSON_REPAIR_FAILED / REPAIR_SCHEMA_FAILED | ✅ | |
| BUSINESS_SAFETY_RISK（虚构经历/造假） | ✅ | |

### 3.4 HunyuanClient 设计

- 基于 OpenAI-compatible API（`POST /v1/chat/completions`）
- 支持 endpoint normalization（`/v1`、`/v1/`、`/v1/chat/completions` 自动补全）
- HTTP 429 → `LLM_RATE_LIMIT`、5xx → `LLM_HTTP_ERROR`、401/403 → `LLM_API_KEY_MISSING`
- httptest 22 项测试全部通过
- 真实 Smoke：provider=hunyuan, model=hy3-preview, finish_reason=stop, json_parse_success=true, duration_ms=1327

---

## 四、数据库设计

### 4.1 核心表结构

| 表名 | 行数 | 核心字段 | 设计决策 |
|------|:---:|------|---------|
| users | 2 | id, email, password_hash, name | 最小用户模型，不做 RBAC |
| resume_versions | 2+ | id, user_id, version_no, education_json, experience_json, project_json, skills_json | **单主简历** + JSON 字段存储子结构 |
| jobs | 46 | id, title, description, requirements, external_source, data_quality, raw_json | 腾讯岗位完整 JD + 本地 Mock 共存 |
| recommendation_results | 运行时 | user_id, resume_id, result_json, generation_tier, generated_by, status | **Append-Only** + latest 读取 |
| match_analyses | 运行时 | user_id, resume_id, job_id, score, result_json, generation_tier, generated_by, status | Fast/Deep 分层存储 |
| version_compare_results | 运行时 | user_id, from_version_id, to_version_id, result_json | 版本对比快照 |

### 4.2 数据库设计原则

1. **单主简历**：不设 `resumes` 多简历表，`resume_versions (user_id, version_no)` 唯一约束
2. **JSON 字段存储子结构**：教育/工作/项目经历存储在 JSON 字段中，Demo 阶段不过度范式化
3. **外部数据源标记**：`external_source`（local_mock / tencent_join）+ `data_quality`（full_jd / summary_only / manual_enriched）
4. **结果快照存储**：推荐/匹配/对比结果持久化到 result_json，支持历史回溯
5. **幂等导入**：`external_source + external_id` 唯一索引，INSERT ON DUPLICATE KEY UPDATE

### 4.3 K9 系列表结构升级

| 版本 | 新增字段 | 说明 |
|------|---------|------|
| K9-1-U-B | generation_tier, generated_by, status | 推荐结果增加生成方式标记 |
| K9-2-B | generation_tier, generated_by, status, profile_type | 匹配分析增加 Fast/Deep 分层 |
| K9-3-B | optimization_mode, generated_by, status | 简历优化增加三模式设计 |

---

## 五、接口设计

### 5.1 统一返回契约（4-4 冻结）

```
{
  "code": 0,
  "message": "ok",
  "data": {
    "result": { ... },         // 业务结果
    "meta": {                  // 元信息
      "generated_by": "llm_agent" | "mock_agent",
      "agent_status": "success" | "repaired" | "fallback",
      "fallback_reason": "...",
      "error_code": "...",
      "repair_applied": false,
      "schema_valid": true,
      "request_id": "..."
    }
  }
}
```

### 5.2 Fast/Deep 双阶段接口设计

**岗位匹配分析**：
- `POST /api/jobs/{id}/match` — 生成分析（`generation_tier`: fast/quality）
- `GET /api/jobs/{id}/match/latest?resume_id=X&tier=fast` — 读取最新结果

**岗位推荐**：
- `POST /api/jobs/recommend` — 生成推荐（`generation_tier`: fast/quality）
- `GET /api/jobs/recommend/latest?resume_id=X` — 读取最新结果

**结果隔离**（H20-D-2 修复）：
- 后端 `FindLatestByResumeAndJobAndTier` — 按 tier 过滤查询
- 前端 `fastResult` / `deepResult` 拆分 — 避免 Deep 结果污染 Fast 视图

---

## 六、K9 系列：语义画像与持久化状态机

### 6.1 SEP（Semantic Evidence Profiling Layer）

定位：推荐/匹配/优化/对比的**统一语义画像输入底座**。

| SEP 类型 | 内容 |
|---------|------|
| Resume SEP | 简历语义摘要、核心能力标签（label/code/confidence/evidence）、风险缺口、适配方向、profile_status、content_hash |
| Job SEP | 岗位方向/技能/要求/责任标签、风险信号、JD 摘要 |
| SemanticProfileGate | 核心功能入口 check profile 存在性 + 新鲜度（content_hash 比对） |

**关键数据**：
- 9/9 简历 SEP 全部 ready + 46/46 岗位 SEP 全部 ready（H20-C-4 验收）
- Fast Recommend：50-83ms（ready SEP 场景）
- Fast Match：4-5ms（ready SEP 场景）
- Profile 生成：18s（含 Hy3 LLM 调用）

### 6.2 GQCL（Generation Quality Control Layer）

定位：推荐输出的**统一质量门禁**。

| 组件 | 职责 |
|------|------|
| Fast Output Composer | 将 Fast Ranker raw tag + SEP labels 组合为产品化输出 |
| Deep Quality Gate | 验证 Deep 每层输出满足字段契约、可读性、证据关联性 |
| raw code 禁止 | 任何用户可见输出不得含 `research`/`redis`/`algorithm`/`degree_bachelor` 等 code |
| 前端 Safety Adapter | `tagLabelMap.ts` 120+ 映射统一 humanize 防御 |

### 6.3 推荐 Pipeline（Fast 2层 + Deep 3层）

**Fast Pipeline（2层）**：
- Layer 1：规则召回（从 46 个岗位召回 10-15 个）
- Layer 2：规则精排（输出 Top5 + 分数 + 理由 + 亮点 + 风险）

**Deep Pipeline（3层）**：
- Layer 1：候选召回（基于简历短画像 + 岗位压缩卡片）
- Layer 2：精排评分（完整简历 + 候选岗位详情，输出 Top5 含证据）
- Layer 3：Judge 审查（9 项规则检查：方向错配/分数趋同/理由空泛/风险不真实等）

### 6.4 匹配分析 Fast/Deep 字段差异

| 字段 | Fast | Deep |
|------|:---:|:---:|
| match_score + match_level | ✅ | ✅ |
| one_sentence_summary | ✅ | ✅ |
| fit_points | ≤2 | ✅ |
| resume_evidence | — | ✅ |
| job_requirements | — | ✅ |
| gap_points | ≤1 | ✅ |
| risk_points | — | ✅ |
| improvement_suggestions | ≤1 | ✅ |
| interview_talking_points | — | ✅ |
| delivery_strategy | — | ✅ |

### 6.5 推荐结果状态机（K9-1-U-B）

```
页面进入
  → GET /latest（has_result?）
      ├── has_result=true  → 使用 latest 渲染（fast→Fast，quality→Deep）
      └── has_result=false → POST /api/jobs/recommend generation_tier=fast
                               → 保存后展示 Fast
                                  └── 用户点击 Deep
                                      → POST generation_tier=quality
                                      → 成功后切换 Deep 完整展示
```

前端状态枚举：
```ts
type RecPhase =
  | "fast_loading" | "fast_done" | "fast_error"
  | "deep_loading" | "deep_profile_pending" | "deep_done"
  | "deep_error" | "profile_failed";
```

---

## 七、JM 系列：岗位匹配核心重构

### 7.1 重构基线审计（JM-R0）

核心发现：**"价值表达错误"——把"字段完整"当成"产品有用"**
- Deep v3 是规则引擎（无 LLM），内容浅、模板化
- 面试追问/Agent Trace/quality_checks 持续污染模块边界
- 学生读完 Deep 不知道具体怎么改简历

### 7.2 新方向

**岗位匹配 = 初筛通过率诊断 + 简历命中率提升方案**

- **Fast**（秒级初筛）：3 个核心判断 + 6 模块卡片
- **Deep**（5 模块 Hy3 深度分析）：
  - A. 投递决策（可直接投递/修改后投递/备选/不建议/建议换方向）
  - B. 为什么可能过初筛（命中点 + 真实证据）
  - C. 为什么可能被筛掉（淘汰风险 + 影响说明 + 修复方向）
  - D. 简历具体怎么改（before/after 改写示例）
  - E. 下一步投递策略（推荐更适合岗位方向）

### 7.3 Fast Match 引擎（H18-D 重构）

| 特性 | 重构内容 |
|------|---------|
| matchLevel 阈值 | 统一 90/80/70/60 |
| raw code 泄漏修复 | `humanizeRequirementCategory()` 10 类映射 |
| risk 结构化 | `structured_risks` 5 字段完整输出 |
| coverage 补全 | 6 状态（strong/partial/transferable/weak/missing/unknown） |
| Hard Gate 强化 | 4→8 条规则，新增 internship/weak expression 检测 |
| 新字段 | score_breakdown, hard_gate, core_advantages, main_gaps, action_suggestions |

### 7.4 Deep Match 引擎（JM-R3-R 重构）

**从纯规则引擎 → 真 Hy3 5 模块 LLM 生成**：
- 新增 `runHy3ScreeningPipeline`（450+ 行代码）
- 6 个 prompt builder + JSON repair + Judge 校准
- Fallback 正确标记：verified / degraded / fallback
- 3 案例验收：C1 78/revise, C2 33/change_direction, C3 42/not_recommended

---

## 八、Quality Gate 质量门禁体系

### 8.1 四级枚举

| 级别 | 含义 | 后续动作 |
|------|------|---------|
| pass | 内容质量合格 | 直接返回 |
| revise_required | 需要修正 | revise once，仍不合格则 fallback |
| fallback_required | 存在严重问题 | 直接 Mock fallback |
| manual_review_suggested | 建议人工审查 | 保守输出 |

### 8.2 十项检查维度

1. JD grounding check（是否基于真实 JD）
2. Resume evidence check（是否基于简历证据）
3. Fabrication check（是否编造经历）→ **一旦失败直接 fallback**
4. Apply boundary check（是否给出投递边界）
5. Optimization stop check（是否给出优化停止条件）
6. Example reference check（是否使用示例参考格式）
7. Verification notice check（是否提醒用户核实）
8. Self-check question check（是否包含自查问题）
9. Sensitive attribute check（不评价敏感属性）
10. Schema compliance check（符合 Schema）

### 8.3 极端情况处理

| candidate_case_type | case_risk_level | handling_strategy |
|---------------------|:---:|------|
| normal_case | low | normal_response |
| insufficient_information | medium | clarify_before_recommendation |
| weak_resume | medium | quick_fix_then_apply |
| target_role_mismatch | high | recommend_alternative_jobs |
| overseas_or_location_risk | medium | conservative_advice |
| authenticity_risk | high | conservative_advice |
| over_optimization_risk | medium | conservative_advice |
| unsafe_user_request（要求编造经历） | critical | refuse_fabrication（直接拒绝，不 fallback 到 Mock） |

---

## 九、推荐页重构设计（架构设计冻结 B）

### 9.1 当前问题

| 问题 | 说明 |
|------|------|
| Fast 推荐 | 只有 Ranker 无 Explainer，raw tag code 直出（`research`/`redis`/`algorithm`） |
| Deep 推荐 | 旧 3-Agent 架构，非岗位级深度分析，输出泛化 |
| Profile 层 | 仅规则提取 raw tag，无 content_hash/model_name，质量不可验证 |
| 数据库 | 23 张表，profile 表缺关键字段 |

### 9.2 重构方向

**从"规则排序+泛化 LLM"升级为"语义证据驱动+级联深度分析"**

- **SEP**：统一语义画像输入底座
- **GQCL**：统一质量门禁（Fast Composer + Deep Quality Gate + raw code 禁止 + 前端 Safety Adapter）
- 旧 `resume_profiles` / `job_profiles` **保留但冻结写入**，新 SEP 数据写入 `semantic_profiles` + `semantic_profile_tags`

---

## 十、工程实践亮点

### 10.1 安全设计

| 措施 | 实现 |
|------|------|
| API Key 保护 | Go struct `json:"-"` + `MaskedAPIKey()` 脱敏 + `safeErr()` |
| .env 不入库 | `.gitignore` 排除 + `.env.example` 仅占位符 |
| 日志安全 | 不记录 API Key / 完整简历 / 完整 JD / 完整模型输出 |
| LLM 默认关闭 | `LLM_ENABLE=false`，不配 Key 走 Mock |
| Repair 防递归 | `LLM_REPAIR_MAX_ATTEMPTS=1` |
| CORS 配置化 | `CORS_ALLOWED_ORIGINS` 环境变量，不允许的 Origin 不返回 CORS 头 |

### 10.2 可观测性

| 日志字段 | 说明 |
|---------|------|
| request_id | 请求唯一 ID |
| agent_name | 业务 Agent 名称 |
| call_duration_ms | LLM 调用耗时 |
| json_parse_success | JSON 解析结果 |
| schema_validate_success | Schema 校验结果 |
| repair_attempted / repair_success | JSON 修复状态 |
| fallback_used / fallback_reason | 降级信息 |
| error_code | 错误码 |

### 10.3 测试覆盖

- **LLM Client 层**：27 项单元测试
- **HunyuanClient**：httptest 22 项测试
- **JSONRepairAgent**：14 项单元测试
- **Pipeline Runner**：18 项单元测试
- **MatchAnalysisAgent**：28 项单元测试 + 15+ 服务层测试
- **RecommendationAgent**：12 项 Agent 测试 + 19 项 service 层测试
- **Fast Match 精准度**：24 组测试，15 种不同值
- **Deep Match 精准度**：5 样例 ALL PASS
- **全量 Go 测试**：`go test ./...` 全部通过

---

## 十一、总结：架构核心亮点

1. **极致精简的技术栈**：Go 标准库 + 裸 SQL，零 Web 框架依赖，零 ORM
2. **完整的 LLM 工程化链路**：Prompt 资产化管理 + Schema 约束 + Pipeline 安全链路 + repair once + Mock fallback
3. **Fast/Deep 双阶段策略**：Fast 秒级语义匹配（4-5ms）+ Deep 5 模块 LLM 深度分析（Hy3）
4. **多层质量防线**：Schema Validate → JSON Repair → Quality Gate（10 维检查）→ 极端情况处理（10 种枚举）
5. **语义画像底座（SEP）**：统一为推荐/匹配/优化/对比提供语义标签，46 岗位 + 9 简历全部就绪
6. **输出质量门禁（GQCL）**：raw code 禁止 + Fast Composer + Deep Quality Gate + 前端 Safety Adapter
7. **结果持久化状态机**：Append-Only 存储 + latest 读取 + Fast/Deep tier 隔离 + 6 状态前端管理
8. **Mock 优先策略**：LLM 失败自动降级，保证 Demo 稳定可演示

---

## 十二、进入第 3 阶段分析

第 3 阶段将分析：前端产品化与工程部署（`01-前端设计/`、`05-前端设计/`、`07-前端设计/`、`06-工程部署/`、`15-前端专业化/`）

