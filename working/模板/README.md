# 后端迁移参照（仅读参照，不是新架构模板）

> 本目录是**旧系统（Go net/http + database/sql 单体）的行为参照**，供 GoFrame 重构时对齐「接口契约、数据结构、输入输出形状」。
> 产出原则：**只读、不修改旧代码、不使用 git 命令**。

## 目录

| 文件 | 内容 |
|---|---|
| [01-接口清单.md](./01-接口清单.md) | 现有全部 HTTP 接口（路径 / Method / 鉴权 / 请求响应要点） |
| [02-示例数据.md](./02-示例数据.md) | 可用的示例用户、简历（两份）、岗位数据 |
| [03-模型功能-输入输出样例.md](./03-模型功能-输入输出样例.md) | 匹配 / 推荐 / 优化 / 对比四个模型功能的请求与响应 JSON 样例 |

---

## 核心声明（必读）

### 1. 旧系统是「行为参考」，不是「架构模板」

新系统迁移到 **GoFrame** 时，**只对齐旧系统的对外行为**，**禁止照搬旧系统的内部结构**。具体红线：

| 旧系统做法（行为参考） | 新系统禁止照搬 |
|---|---|
| 对外返回 `{code, message, data}` 统一包装 | ✅ 必须保持一致（前端契约） |
| 四个功能的 JSON 字段名（`match_score` / `recommended_jobs` / `optimization_mode` / `winner_resume_id` 等） | ✅ 必须保持一致（前端契约） |
| `generation_tier` 取 `fast` / `quality`，`generated_by` 取 `llm_agent` / `mock_agent` 等枚举 | ✅ 必须保持一致（业务语义） |
| 巨型 `main.go` 组合根（280 行手写装配） | ❌ 用 GoFrame `cmd` + service 注册表替代 |
| 手写 `net/http` + `http.ServeMux` 路由 | ❌ 用 `ghttp.Server` + 路由组替代 |
| 手写 `database/sql` Repository | ❌ 用 `gdb` / `gf gen dao` 替代 |
| 手写 `godotenv` + `getEnv*` 配置 | ❌ 用 `gcfg` + `config.yaml` 替代 |
| Service 构造器 12~15 个参数 | ❌ 用 GoFrame service 接口注册 / 依赖注入替代 |
| `errors.AppError` int 码 + `llm.ErrorCode` + `AgentResponseMeta` 三套错误码并存 | ❌ 统一收敛，但**保留 `meta` 的 JSON 形状**给前端 |
| `agent` 单包塞进业务/Mock/Deep/快照工具 | ❌ 拆分到 `logic/agent` 子包 |

### 2. 必须保留的领域资产（GoFrame 不覆盖，属核心竞争力）

- `llm.Client` 接口 + `HunyuanClient` / `Hy3Client` / `NoopClient`
- `pipeline.Runner` 统一管线（渲染 → 调用 → JSON 抽取 → Schema 校验 → 修复一次 → 兜底）
- `repair.Agent`（JSON 修复）、`guardrail`（本地确定性护栏）
- `prompt` / `schema` / `llmjson` + `server/prompts/` 全部 Prompt/Schema 资产
- 6 个 Mock Agent 的**规则逻辑**
- Deep 3-agent / DeepCascade / SEP 语义画像引擎

### 3. 四个模型功能 = 四张结果快照表

旧系统的「四大模型功能」对应四张 `ResultJSON` 快照表，迁移时表结构语义需保留：

| 模型功能 | 核心 DTO | 快照表 |
|---|---|---|
| 岗位匹配 Match | `model.MatchAnalysisResult` | `match_analyses` |
| 岗位推荐 Recommendation | `model.RecommendedJob` | `recommendation_results` |
| 简历优化 Optimization | `model.OptimizationResultV2` | `resume_optimizations` |
| 版本对比 Compare | `model.CompareResponseV2` | `resume_compares`（V2） |

### 4. 使用建议（面向并行 agent）

- **对齐契约**：新 controller 的输出 JSON 必须与 `03-模型功能-输入输出样例.md` 中 `data` 字段形状一致。
- **对齐枚举**：`generation_tier`、`generated_by`、`agent_status`、`fallback_reason`、`error_code`、`optimization_mode`、`comparison_mode`、`screening_result`、`apply_readiness` 的取值以本目录样例为准。
- **数据源**：岗位 seed 来自 `resources/jobs/tencent-campus-jobs.json`（46 个岗位），无需重造。
