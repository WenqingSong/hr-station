# 第 1 阶段分析：项目全景与迭代记录

> 分析时间：2026-06-23
> 源文档：`docs/` 下 850+ Markdown 文件
> 核心阅读：`说明书.md`、`项目状态.md`、`00-项目总览/`、`08-架构决策记录/`、`09-开发日志/`（550+日志）、`变更记录.md`

---

## 一、项目概述

### 1.1 项目名称与定位

**Offer 捕手**（又名 Offer 鹅 / Offerr鹅）是一个**面向校招求职者的 AI 岗位推荐与简历匹配分析平台**。

核心理念：**不替用户编造经历，而是帮用户更好地呈现真实经历。**

### 1.2 一句话定位

面向大学生求职场景的 AI 岗位推荐与简历匹配分析 Demo，核心亮点是「JD-简历证据链 + 初筛风险 + 简历版本迭代」。

### 1.3 技术栈

| 层级 | 技术栈 |
|------|--------|
| 前端 | React 18.3 + Vite 6.3.5 + TypeScript + TailwindCSS 4.1 + shadcn/ui |
| 后端 | Go（标准库 net/http） + MySQL 8.0 |
| AI/LLM | 腾讯混元 Hy3 + DeepSeek V4 Pro（OpenAI-compatible API） |
| 部署 | Docker Compose + Nginx + Makefile（混合式启动：容器化 MySQL+Nginx + 宿主机 Go/Vite） |
| 数据源 | 腾讯招聘官网公开岗位数据（46 条手动精选岗位） |

---

## 二、产品定位与核心价值

### 2.1 目标用户

- 正在准备校招的应届毕业生
- 有简历但不知道投什么岗位的求职者
- 已有目标岗位但不确定匹配度的求职者
- 想要通过迭代优化简历提升匹配度的求职者

### 2.2 解决的痛点

| 痛点 | 说明 | 解决方案 |
|------|------|---------|
| 海投无反馈 | 不知道哪些岗位真正适合自己，盲目投递效率低 | 基于简历的岗位智能推荐 |
| 匹配度不明 | 无法量化评估自己与岗位的匹配程度 | JD-简历证据链 + 多维度评分 |
| 简历不会改 | 不知道简历哪里有问题、怎么改才能提升匹配度 | P0/P1/P2 三级优化建议 + before/after 示例 |
| 迭代无追踪 | 修改简历后无法对比改进效果 | 版本迭代 + 版本对比（V1→V2→V3） |

### 2.3 核心差异点

1. **JD-简历证据链**：不是简单的匹配度评分，而是逐条展示岗位要求与简历内容的对应关系，每条绑定真实证据
2. **初筛风险提示**：识别简历中可能导致初筛被淘汰的表达风险（模糊描述、关键技能缺失等）
3. **简历版本迭代**：支持"分析→修改→再分析"的迭代闭环，追踪改进效果
4. **不代写**：不替用户生成优化版简历，只给诊断和建议，保护求职真实性

---

## 三、四大核心功能体系

### 3.1 四主 Agent 架构（K6-0 冻结）

| # | 核心功能 | 主 Agent | 目标 |
|---|---------|---------|------|
| 1 | 简历版本对比 | Resume Version Compare Agent | 比较不同简历版本变化、提升点、退化点，判断是否更贴近目标岗位 |
| 2 | 岗位智能推荐 | Job Recommendation Agent | 基于当前简历，推荐更适合投递的岗位，给出推荐理由和风险提示 |
| 3 | **岗位匹配分析** ⭐ | Job Match Analysis Agent | 分析简历与岗位JD匹配度，输出匹配分、初筛判断、证据链、投递决策 |
| 4 | 简历优化建议 | Resume Optimization Agent | 针对岗位给出可执行的简历修改建议、关键词补强和改写方向 |

### 3.2 岗位匹配分析详解（核心主推功能）

岗位匹配分析是产品的核心竞争力，分为 **Fast（秒级初筛）** 和 **Deep（LLM 深度分析）** 双阶段：

**Fast 模式（语义画像引擎）：**
- 秒级返回初筛判断
- 6 模块：结论卡、五维评分+硬门槛、核心优势与短板、风险诊断、要求覆盖、行动建议
- 匹配等级 5 级：高度匹配 / 建议投递 / 修改后投递 / 备选 / 不建议投递

**Deep 模式（5 模块 Hy3 深度分析）：**
- A. 投递决策：分数 + 建议（可直接投递/修改后投递/备选/不建议投递/建议换方向）
- B. 为什么可能过初筛：简历命中点，绑定真实证据
- C. 为什么可能被筛掉：淘汰风险识别，说明影响和修复方向
- D. 简历具体怎么改：before/after 改写示例，基于学生真实经历
- E. 下一步投递策略：推荐更适合岗位方向，避免方向错位

### 3.3 三大战略杀手锏（4-3.5 阶段产出）

| 杀手锏 | 说明 |
|--------|------|
| 投递边界与优化停止条件 | `apply_readiness` 四档判断 + `optimization_stop_signal`：区分 Top 3 快速优化项与长期提升项 |
| JD 证据链匹配 | `jd_evidence_chain` + `evidence_strength` + `match_judgement` + `screening_risk` |
| Agent 输出质量门禁 | `quality_gate_result` 四级枚举 + 十项检查维度 + revise once 约束 + conservative/mock fallback |

### 3.4 功能边界（ADR 决定）

| 边界 | 说明 |
|------|------|
| 单主简历 | 用户只有一份主简历，V1→V2→V3 版本迭代 |
| 所有字段可选填 | 降低填写压力 |
| 不上传 PDF/DOC | 只支持结构化表单录入 |
| 不自动生成简历 | 只给诊断和建议，用户手动修改 |
| 先 Mock 后接模型 | 前期 Mock Agent 跑通闭环，后续接入真实 LLM |

---

## 四、技术架构概览

### 4.1 整体架构

```
前端（React + Vite + TypeScript）
    ↕ REST API (JSON, JWT Auth)
后端（Go, net/http）
    ├── Model 层（6 个数据模型）
    ├── Repository 层（6 个数据访问）
    ├── Service 层（6 个业务逻辑）
    ├── Handler 层（6 个 HTTP 处理）
    ├── Middleware 链（Recover → CORS → RequestLogger → Auth）
    ├── Agent Pipeline（Prompt → LLM → JSON parse → Schema validate → repair → fallback）
    └── LLM Client 层（Client 接口 + HunyuanClient + NoopClient）
    ↕
MySQL 8.0（8 张核心表）
    ↕ 低频同步
腾讯招聘官网公开岗位数据源（46 条岗位）
```

### 4.2 后端分层架构

| 层级 | 文件数 | 职责 |
|------|--------|------|
| `cmd/server` | 1 | 程序入口，初始化配置、数据库、中间件、路由 |
| `internal/config` | 1 | 7 组结构化配置（App/Server/Database/JWT/LLM/Feature/Log） |
| `internal/model` | 6 | 数据模型定义 |
| `internal/repository` | 6 | 数据库访问层 |
| `internal/service` | 6 | 业务逻辑层 |
| `internal/http` | 6 | HTTP Handler + Router |
| `internal/middleware` | 4 | Recover / CORS / RequestLogger / Auth |
| `internal/agent` | 7+ | Mock Agent（3）+ LLM Agent（3）+ Repair Agent + Pipeline |
| `internal/llm` | 3 | LLM Client 抽象 / HunyuanClient / NoopClient |

### 4.3 数据库设计

| 表名 | 说明 | 种子数据 |
|------|------|---------|
| users | 用户注册、登录 | 2 条 |
| resume_versions | 简历版本管理 | 2 条（V1 + V2） |
| jobs | 岗位列表和详情 | 8 条 local_mock + 46 条 manual_curated |
| recommendation_results | 推荐结果快照 | 运行时生成 |
| match_analyses | 匹配分析结果 | 运行时生成 |
| version_compare_results | 版本对比结果 | 运行时生成 |

### 4.4 Agent Pipeline 安全链路

```
业务 Agent 构造 PromptValues
    → Prompt Loader（加载 .md 模板 + 占位符替换）
    → Schema Loader（加载 JSON Schema）
    → llm.Client.Chat（Hunyuan / DeepSeek）
    → llmjson Extract / Parse（提取 JSON + 清理 Markdown fence）
    → Schema Validate（draft 2020-12）
    → 失败 → JSONRepairAgent repair once
    → 再失败 → Mock Agent fallback
```

关键安全设计：
- `LLM_API_KEY` Go struct tag `json:"-"` 防止序列化泄露
- `MaskedAPIKey()` 日志脱敏
- `safeErr()` 不在错误中泄露 Key
- `LLM_REPAIR_MAX_ATTEMPTS=1` 禁止递归修复
- `LLM_ENABLE=false` 默认关闭
- `.env` 不在 Git 追踪中

---

## 五、关键架构决策（ADR）

### ADR-001：技术栈选择
- 前端：React 18 + Vite + TypeScript + TailwindCSS + shadcn/ui（沿用 Figma Make 生成的代码基础）
- 后端：Go + 标准库 net/http
- 数据库：MySQL 8.0
- AI：先 Mock 后接模型

### ADR-002：单主简历设计
- 用户只有一份主简历，通过 V1→V2→V3 版本迭代管理
- 历史版本只读，不可编辑、不可删除
- 简化 UI 和状态管理，Demo 阶段聚焦核心闭环

### ADR-003：先 Mock 后接模型
- 阶段 3 使用 Mock Agent 跑通核心业务闭环
- 阶段 4 接入真实 LLM
- 降低不确定性，保证产品闭环先行

### ADR-004：不自动生成简历
- 保护求职真实性，避免 AI 虚构经历
- 系统只给诊断和建议，用户手动修改
- 降低法律风险，突出产品差异化

---

## 六、详细迭代记录

> 以下记录覆盖从 2026-05-28 到 2026-06-14 的全部 13 个阶段、550+ 条开发日志。

### 阶段 0：项目审计与文档中枢（2026-05-28）

| 编号 | 任务 | 说明 |
|------|------|------|
| 0-1 | 项目只读审计 | 完成项目结构扫描、技术栈判断、页面识别、风险梳理 |
| 0-2 | 创建 docs 中文文档体系 | 建立 29 个子目录、850+ Markdown 文档的文档中枢 |
| 0-3 | 编写产品范围文档 | 完善产品定位、用户路径、功能范围、不做范围 |
| 0-4 | 编写前端工程化改造计划 | 形成六阶段前端改造计划 |
| 0-5 | 编写工程执行规范 | 建立 8 阶段工程执行规范、15 项 CodeBuddy 禁止事项、文档更新规范 |
| 0-6 | 阶段 0 汇总 | 产出 54 个核心文档，确认可进入阶段 1 |

**关键产出**：4 个 ADR（架构决策记录） + 5 个产品/工程规则文档 + 完整的文档体系框架

### 阶段 1：工程初始化（2026-05-28）

| 编号 | 任务 | 说明 |
|------|------|------|
| 1-1 | 创建工程初始化方案 | 规划阶段 1 整体方案和修改范围 |
| 1-2 | 创建 Makefile | 19 个命令的统一开发入口 |
| 1-3 | 创建 .env.example | 6 组环境变量，按阶段启用 |
| 1-4 | 编写 docker-compose.yml | MySQL 8.0 服务编排，含 healthcheck、数据卷、网络 |
| 1-5 | 初始化 Go 后端工程 | cmd/server + internal 分层目录 + `/healthz` 接口 |
| 1-6 | 编写前后端运行说明 | 12 章节完整运行说明 |
| 1-7 | 阶段 1 汇总 | 确认可进入阶段 2 |

**交付物**：Makefile（19 命令）、.env.example、docker-compose.yml、Go 工程骨架、运行说明文档

### 阶段 2：后端基础设施（2026-05-28）

| 编号 | 任务 | 说明 |
|------|------|------|
| 2-1 | 创建阶段 2 方案 | 规划后端基础设施方案 |
| 2-2 | 增强配置加载 | 7 组结构化配置，从 .env 读取 |
| 2-3 | 统一响应结构 | `{code, message, data}` + AppError + 错误码 |
| 2-4 | 中间件实现 | Recover + CORS + RequestLogger 中间件链 |
| 2-5 | MySQL 数据库连接 | database/sql + go-sql-driver/mysql，连接池配置 |
| 2-6 | 增强健康检查 | `/healthz` 返回 app/database 状态 |
| 2-7 | 阶段 2 汇总 | 确认可进入阶段 3 |

**交付物**：config.go、response.go、error.go、3 个中间件、mysql.go、health service、统一路由

### 阶段 3：核心业务后端 Mock（2026-05-29）

| 编号 | 任务 | 说明 |
|------|------|------|
| 3-1 | Mock 方案设计 | 规划阶段 3 整体方案 |
| 3-2 | 数据库初始化 | 6 张表 + 种子数据写入 init.sql |
| 3-3 | 用户注册/登录 | JWT 签发与验证，Auth 中间件 |
| 3-4 | 简历版本接口 | 5 个接口：Create / Current / Detail / Versions / VersionDetail |
| 3-5 | 岗位接口 | List / Detail，支持分页和筛选 |
| 3-6 | 岗位推荐 Mock | RecommendationMockAgent（规则打分：城市/方向/技能）
| 3-7 | 匹配分析 Mock | MatchMockAgent（规则打分 + 证据链 + P0/P1 建议）
| 3-8 | 版本对比 Mock | VersionCompareMockAgent（字段差异 + 分数变化）
| 3-9 | 阶段 3 汇总 | 14 个 curl 可验证接口，后端 Mock 闭环完成 |

**交付物**：
- 14 个 API 接口（健康检查 + 认证 + 简历 + 岗位 + 推荐 + 匹配 + 对比）
- 6 层架构：Model × 6 / Repository × 6 / Service × 6 / Handler × 6 / Agent Mock × 3 / Middleware × 1
- 统一响应结构 + JWT 鉴权

### 阶段 4：Agent 与模型接入（2026-05-29 ~ 2026-05-31，项目最复杂的阶段）

#### 4-1：数据源与 Prompt 准备（2026-05-29）

| 编号 | 任务 | 说明 |
|------|------|------|
| 4-1.1 | Makefile 升级 | 40+ 命令统一开发入口 + 11 个 scripts/dev 脚本 |
| 4-1.2.0~6 | **腾讯招聘数据源接入** | 探测确认 3 个接口可用（列表+详情+字典）；15 项配置固化；拉取脚本 9 子命令；小规模验收 7 次请求全部通过 |
| 4-1.3.0~7 | **腾讯岗位数据适配** | jobs 表扩展 8 字段 + 3 索引；清洗入库脚本 7 子命令；幂等导入；local_mock 8 条 + tencent_join 5 条共存；接口适配 external_source 筛选、data_quality 感知 |
| 4-1.4.0~13 | **Prompt 与 Schema 资产建设** | 三个业务 Agent Prompt 逐个定稿、CodeBuddy 登记、首轮输出验收（均 partial_pass）；三个 Schema 逐个定稿；JSON 修复 Prompt；失败兜底 20 种错误码 + 6 种直接 Mock + 6 种先 repair |

**关键突破**：
- 岗位候从 8 条种子扩展到 13 条（后扩展到 46 条）
- 完成了 4 个 Prompt + 3 个 Schema 的全套 AI 资产
- 首次验证 Hunyuan 对真实 JD 的输出质量

#### 4-2：LLM 基础设施（2026-05-29~30）

| 编号 | 任务 | 说明 |
|------|------|------|
| 4-2.1 | 环境变量配置 | 14 项 LLM_* 配置，安全清除遗留 API Key |
| 4-2.2 | Prompt/Schema Loader | 可加载 4 Prompt + 3 Schema，支持占位符替换 |
| 4-2.3 | JSON 工具 | llmjson 包：提取 JSON + parse + BOM 清理 + 错误分类 |
| 4-2.4 | LLM Client 抽象 | Client 接口 + Config + ErrorCode + NoopClient（27 项测试） |
| 4-2.5 | Hunyuan Client | OpenAI-compatible API 实现（httptest 22 项测试） |
| 4-2.6 | LLM Smoke Test | cmd/llm-smoke + Makefile llm-* target |
| 4-2.7 | JSONRepairAgent | repair once 调用链路（14 项测试） |
| 4-2.8 | **统一 Pipeline** | Prompt→LLM→parse→validate→repair→fallback（18 项测试） |
| 4-2.9~10 | 集成验收与汇总 | **132 项测试全部通过**，产物：11 项基础设施资产 |

**关键设计**：Pipeline 20+ 错误码全覆盖，3 种状态（success / repaired / fallback_required），禁止 `generated_by=llm_agent_fallback`

#### 4-3：业务 Agent 接入真实 LLM（2026-05-30~31）

| 编号 | 任务 | 说明 |
|------|------|------|
| 4-3.1 | 真实 LLM Smoke | Hunyuan/TokenHub 最小调用可用；.env 自动生成与 Makefile 接入 |
| 4-3.2 | MatchAnalysisAgent 接入 | Pipeline 接入 + 28 项单元测试 |
| 4-3.3 | 匹配分析联调验收 | 6 场景验收 + LLM 快照保存修复 + 安全边界验证 + 15+ 服务层测试 |
| 4-3.4 | RecommendationAgent 接入 | 候选岗位校验 + 字段映射 + RecFallbackError + 12 项 Agent 测试 |
| 4-3.5 | 推荐联调验收 | 8 场景验收 + 元数据修复 + 19 项 service 层测试 |
| **4-3.5.B~H** | **战略转型** | 产品定位升级为"投递前决策与简历初筛优化系统"；三个杀手锏设计（投递边界/证据链/Quality Gate）；学生极端情况 10 种枚举 + 4 级风险 + 7 种策略；灵活在线简历 Profile 设计 |
| 4-3.6 | ResumeOpt 现状核对 | 确认仅有 Prompt+Schema，缺 6 项核心业务代码 |
| 4-3.7 | 简历优化链路补齐 | 新增 Mock Agent / Model / Service / Handler / Route / 测试 |
| **4-3.7.B.R** | **启动部署重构** | **全容器回退为混合式启动**（Docker MySQL+Nginx + 宿主机 Go+Vite）；修复 Vite 子进程残留；CORS 环境变量化；一键启动集成验收 |
| 4-3.7.C~fix | 战略一致性修复 | Prompt/Schema/Go 契约补齐 apply_readiness 等字段；3 个固定测试简历固化；敏感 Key 脱敏 |
| 4-3.8 | 三类 Agent 集成验收 | ResumeOpt Pipeline 最小集成；LLM/fallback 本地测试补齐；三类 Agent 集成验收收口 |
| 4-3.9 | 前端实现拆分 | 三类 Agent 输出到前端模块映射；6 阶段前端实现计划冻结 |

**关键调整**：4-3.5 阶段完成了重大战略转型——从"匹配评分工具"升级为"投递决策系统"，新增投递边界（apply_readiness）、优化停止信号（optimization_stop_signal）、证据强度（evidence_strength）、质量门禁（Quality Gate）等核心概念。同时 4-3.7.B.R 阶段将全容器部署回退为更灵活的混合式启动。

#### 4-4：接口返回契约统一（2026-05-31）

| 编号 | 任务 | 说明 |
|------|------|------|
| 4-4.0 | 接口返回结构核验 | 发现 agent_status Mock direct 值不一致、meta 字段位置不一致等问题 |
| 4-4.1 | 统一返回契约冻结 | `data.result + data.meta` 分层、状态枚举、兼容策略 |
| 4-4.2 | 三类接口结构改造 | 补齐 generated_by、agent_status、fallback_reason 等 meta 字段 |
| 4-4.3 | 后端集成验收 | 三类接口 data.result + data.meta 验收通过 |

#### 4-5：阶段 4 集成验收（2026-05-31）

| 编号 | 任务 | 说明 |
|------|------|------|
| 4-5.0 | 验收矩阵冻结 | 明确"下一阶段可进入什么"的判断标准 |
| 4-5.1 | 模拟异常验收 | LLM 失败、JSON 异常、repair、fallback 全场景覆盖 |
| 4-5.2 | 真实 LLM Smoke | partial_pass（API Key 为占位符，认证失败） |
| 4-5.3 | 三类 Agent 联调 | pass（修复 prompt_load_failed） |
| 4-5.4 | 阶段 4 总收口 | **pass_with_known_limitations**：三类 Agent、LLM Client、Hunyuan Client、JSON Repair、Mock fallback、统一接口返回契约均已完成 |

### 阶段 5：前端产品体验闭环（2026-05-31，单日完成 52 个子任务）

| 阶段 | 核心内容 |
|------|---------|
| **5-1** | 前端数据契约与 Mock 标准化：TypeScript 类型定义 + Mock 数据标准化 + Adapter/Selector 统一 |
| **5-2** | API Client 建设：fetch 原生封装（非 axios）+ Jobs/Resumes/Agents API 模块 + ViewModel 统一消费（9 个 Loader 函数） |
| **5-3** | 布局/路由/骨架统一：Navbar 去重（4 冗余组件废弃）+ 简历路由补齐 + 5 页骨架统一 1440px |
| **5-4** | 页面交互闭环：SelectResumeModal 全局接入 + 三页 URL 参数真实传递 + 简历管理操作闭环 |
| **5-5** | 前端工程产物完整性核验：4 层 29 文件齐全，8 项越界零违规 |
| **5-6** | Build 准入验收：P0 build 风险解除，13 项验收全部通过 |
| **5-7** | 静态 Demo 演示验收：3 条演示路径 9 URL 全部 200，P0=0 |
| **5-8** | 视觉与文案穿帮点清理：10 项问题修复，4 个 legacy 组件删除 |
| **5-9** | 阶段 5 总验收：前端静态闭环/build/Demo/文案/联调前准入全部完成 |

**关键数据**：build 1.51s，dist 完整 560K，11 URL 全部 200，P0=0 P1=0 P2=6

### 阶段 6：普通业务 API 联调（2026-06-01）

| 编号 | 任务 | 说明 |
|------|------|------|
| 6-1 | Jobs API 联调 | curl 验证 GET /api/jobs + GET /api/jobs/1 均 200 |
| 6-2 | Resumes API 联调 | 4 个读接口均 200，参数名对齐修复 |
| 6-3 | 简历写操作联调 | POST /api/resumes V1→V2 自增验证通过 |
| 6-4 | 数据源切换 | **6 组件从 Mocks 切到 API ViewModels** |
| 6-5 | Loading/Error/Empty | 全覆盖状态补齐 |
| 6-6~7 | 总验收 | 8 场景 fallback 通过，可进入阶段 7 |

### 阶段 7：Agent/LLM 真实接入（2026-06-01）

| 编号 | 任务 | 说明 |
|------|------|------|
| 7-0 | Agent 接口现状核验 | 响应契约确认 |
| 7-1 | 岗位推荐 Agent 联调 | curl 200，请求体对齐，推荐页接入 |
| 7-2 | 岗位匹配分析 Agent 联调 | curl 200，match_score=98，匹配页接入 |
| 7-3 | 简历优化 Agent 联调 | curl 200，请求体对齐，优化页接入 |
| 7-4 | Agent 输出结构校验 | 失败兜底补齐 |
| 7-5 | LLM 稳定性验收 | 无 Key fallback 3/3 通过 |
| 7-6 | 上线级核验 | P0=0 P1=5 P2=3，Demo 可演示 |

### 阶段 8：登录态与用户态闭环（2026-06-01）

| 编号 | 任务 | 说明 |
|------|------|------|
| 8-1 | Auth API Client | auth.ts 3 函数 + types/auth.ts 4 类型 |
| 8-2 | 登录/注册页面 | LoginPage + RegisterPage + /login + /register 路由 |
| 8-3 | 全局 Token 管理 | authStorage + Authorization 自动注入 + 401 清 token |
| 8-4 | 用户态回归验收 | P0=0 P1=0 P2=3，可进入阶段 9 |

### 阶段 9：简历数据真实化与保存闭环（2026-06-01~02）

| 编号 | 任务 | 说明 |
|------|------|------|
| 9-0 | 现状核验 | 后端 API 就绪，前端 **10 处硬编码** |
| 9-1 | 简历管理页硬编码清理 | 11 处硬编码清理完毕，新增 ManagementVM + loader |
| 9-2 | 简历编辑初始值接入 | MyResumesPage edit 完整回填，add 空表单 |
| 9-3 | 保存真实联调 | handleSave → createResume/updateResume，演示文案清零 |
| 9-4 | 保存后刷新 | allSettled refresh，curl V1→V2 验证通过 |
| 9-5 | 链路回归验收 | P0=0 P1=0 P2=2，阶段 9 收口 |

### 阶段 10：Agent 页面上下文真实化（2026-06-02）

| 编号 | 任务 | 说明 |
|------|------|------|
| 10-0 | 现状核验 | 主结果 real，上下文 mock |
| 10-1 | SelectResumeModal 真实化 | API versions→OptionVM，默认最新版本 |
| 10-2 | 推荐页上下文真实化 | mock→real，versionNo 真实传递 |
| 10-3 | 匹配页上下文真实化 | mock→real，jobId+versionNo 真实传递 |
| 10-4 | 简历优化页上下文真实化 | resume_text/target_job 真实构建 |
| 10-5 | 一致性回归 | 三页一致，P0=0 P1=0 P2=3 |

### Jobs 样本库重构系列（J0→K5，穿插于阶段 6~11 之间）

| 阶段 | 说明 | 关键成果 |
|------|------|---------|
| J0 | 岗位资源文件结构校验 | 确认 resources/jobs/tencent-campus-jobs.json (162K, 46 岗位) |
| J1 | Jobs 数据库结构重构 | 新增 17 字段 + 3 关联表 |
| J2 | JSON 导入脚本 | 46 条 manual_curated，180 locations，194 BGs，1402 skills |
| J3 | Jobs API 适配 | 新字段/新筛选/新 facets |
| J4 | 首页前端适配 | 8 组筛选 + 46 条岗位 |
| J5 | 浏览器全量验收 | J0→J5 收口 |
| K1 | JobDetailPage 详情展示 | 完整 JD 字段 |
| K2 | 浏览器验收 + Agent 联调 | 入口参数正确 |
| K3 | Jobs+Resume+Agent 主路径 | 全真实上下文 |
| K4 | Agent 结果页端到端验收 | /match + /resume-optimization 链路验收 |
| K5-0 | 真实 LLM 接入前置核验 | P0=0 P1=0，可进入 K5-1 |
| K6-0 | 四主 Agent 架构冻结 | 明确版本对比未 LLM 化，产品架构正式冻结 |
| K6-2 | 简历优化子 Agent 拆分 | 3 子 Agent 设计 + 接口草案 + 实施计划 |

### 阶段 11：真实 LLM 验证与上线准入（2026-06-13~14，进行中）

这是项目最后冲刺的阶段，以 **H 系列**和 **JM 系列**任务为主：

#### H18 系列：Fast Match 全链路

| 编号 | 任务 | 成果 |
|------|------|------|
| H18-B-R | 统一契约与分数字义冻结 | 5 级分数语义 + 8 条 score cap + 5 值投递建议枚举 |
| H18-C | SEP 上下文接入 | 消除 TagEvidenceService 对 Fast 主路径的强依赖 |
| H18-D | 核心算法重构 | matchLevel 阈值统一 90/80/70/60；raw code 泄露修复；risk 结构化 5 字段；coverage 6 状态；Hard Gate 8 条规则；新字段 5 个 |
| H18-E | 前端报告页重构 | 从 4 段式 → 6 模块化卡片；5 新类型定义 + 8 新字段 |
| H18-F | 精准度校准验收 | 5 样例全部 PASS；24 组测试分数 28-58，15 种不同值 |

#### H19 系列：Deep Match 后端重构

| 编号 | 任务 | 成果 |
|------|------|------|
| H19-A | Schema/Prompt 冻结 | 13 项输入上下文 + 8 层 Agent 流程 + 25 顶层字段 Schema + 10 前端展示模块 |
| H19-B | 后端重构 | 新增 4 文件（schema/repair/prompt/agent）；null 修复；generated_by 收口；风险 humanized |
| H19-C | 前端报告页重构 | Deep 9 模块卡片布局；4 状态分流；分数对齐卡 |
| H19-D | 精准度校准验收 | 5 样例 ALL PASS；Fast/Deep delta ≤ 9，is_consistent=true |

#### H20 系列：岗位匹配产品化与契约化（持续进行中）

| 编号 | 任务 | 状态 |
|------|------|------|
| H20-A | **全链路最终验收** | ✅ PASS — 13 tests PASS，0 P0/P1，可上线演示 |
| H20-B-1 | 页面产品目标冻结 | ✅ 页面只回答 4 问：该不该投/为什么/最大风险/怎么改 |
| H20-B-2/2-R | 页面布局重构 | ✅ 680px 窄中栏 → 1280px 8+4 栅格；6 状态兼容 |
| H20-B-3 | 视觉精修验收 | ✅ Hero 分数放大、ReadinessLevel 类型重构、3 折叠区 |
| **H20-C-1** | **SEP Fast Latency 审计** | ✅ 发现 P0：Fast Recommend 同步调 Hy3（7/9 简历受影响） |
| **H20-C-2** | **P0 修复** | ✅ Fast Recommend 不再同步调 EnsureResumeProfile，改为只读+降级 |
| H20-C-3 | SEP 资产补齐 | ✅ backfill CLI + 幂等预热（6/7 成功，46 岗位已 ready） |
| H20-C-3R | Profile 修复 | ✅ 新增 repair 链 + resume 4 修复（16 tags, 18s） |
| H20-C-4 | 延迟验收 + 策略冻结 | ✅ Fast Recommend 50-83ms, Fast Match 4-5ms, SEP 9/9+46/46 |
| **H20-D-1** | **结果隔离审计** | ✅ 发现 P0：Deep 结果成为 latest 后页面默认展示 Deep |
| **H20-D-2** | **隔离修复** | ✅ 后端新增 tier 过滤 + API tier 参数；前端 fastResult/deepResult 拆分 |
| H20-E-1 | 统一设计语言审计 | ✅ 抽取推荐页为设计母版，审计匹配页跳脱点 |
| H20-E-2 | 统一 UI 组件 | ✅ 5 个组件：UnifiedAgentLoadingOverlay/AppButton/SectionCard/StatusBadge/AiNotice |
| H20-E-3 | 视觉语言对齐 | ✅ 1440px + 蓝调科技感 Loading + PrimaryButton/SecondaryButton |
| H20-E-3R | 视觉平衡修复 | ✅ Fast 页面新增核心三卡 + 优先级摘要 + 匹配依据摘要 |
| H20-E-3R-T | 排版对齐 | ✅ 全量文字尺寸缩小对齐推荐页设计标准 |
| H20-E-3R2 | **数据感知 UI 修复** | ✅ 根因：Fast 不生产 Deep 专属字段 → 自适应渲染 + 空状态降级 |
| H20-E-4 | 截图验收 | ⚠️ 下一步 |

#### JM 系列：岗位匹配核心重构（JM-R0→JM-R6→JM-DEEP-OPT1）

| 编号 | 任务 | 成果 |
|------|------|------|
| JM-R0 | 核心重构基线审计 | 发现"价值表达错误"——把"字段完整"当成"产品有用"；Deep v3 是规则引擎无 LLM |
| JM-R1 | 新 Schema 冻结 | Fast 8 字段极简 + Deep 5 模块 screening_match_v1；废弃面试追问/Agent Trace |
| JM-R2 | Deep 5 模块 Prompt | 5 模块 + 1 裁判完整 system/user prompt + JSON schema + 禁止项 |
| JM-R3-R | **真 Hy3 生成链路** | 确认旧链为纯规则管道，修复为真实 Hy3 5 模块 + Judge 生成（+450 行代码） |
| JM-R5-R | **Hy3 输出校准** | Case 1 score 58→78, not_recommended→revise_before_apply；修改 5 模块 prompt |
| JM-R6 | 最终演示验收 | 三案例全部达标：C1 78/revise, C2 33/change_direction, C3 42/not_recommended |
| JM-FIX1 | 锁定主案例 | Resume 2(周明远) 数据支撑度 24/25 最高，锁定 × Job 10001 |
| JM-FIX2 | 按主案例重做页面 | 新增 B 核心三卡、更新 A-F 区域标题、清理不存在经历 |
| JM-DEEP-OPT1 | 全局输出质量优化 | 8 组合回归矩阵审计 + Prompt 全局校准；修复算法→后端误判(78→52)和算法子领域误判(48→78) |

#### 其他系列

| 系列 | 内容 |
|------|------|
| **Recommendation-UX** | 岗位推荐页：tagLabelMap 120+ 映射、Deep Loading 超时适配 90s→900s、右侧按钮 redesign、卡片降噪 |
| **H21 系列** | 推荐页状态机：9 状态冻结 + Deep 按钮修复 + 状态机契约修复 |
| **K13-Deploy-Stable** | 生产环境上线部署：⚠️ 待执行 |
| **DEPLOY-PRECHECK-1** | 上线部署前审计：25 表，P0 禁止 init.sql 直接执行 |

---

## 七、项目成果与数据

### 7.1 工程规模

| 维度 | 数据 |
|------|------|
| 后端文件 | 146 个 Go 源文件 |
| 前端文件 | 139 个 TSX + 40 个 TS |
| 文档文件 | 850+ Markdown（含 550+ 开发日志） |
| 数据库表 | 25 张（含历史增量） |
| API 接口 | 14+ 个 REST 接口 |
| 岗位数据 | 46 条腾讯校园招聘手动精选岗位 |
| 种子简历 | 9 份不同背景的示范简历 |
| 开发周期 | 约 18 天（2026-05-28 ~ 2026-06-14） |
| 开发日志 | 550+ 条，条条有编号、有结论 |
| 阶段划分 | 13 个阶段，11 个已完成，2 个进行中/待执行 |

### 7.2 测试数据

| 测试类型 | 数量/结果 |
|----------|----------|
| Go 单元测试 | 132+ 项（4-2 阶段全量测试） |
| Agent 本地集成测试 | 三类 Agent LLM success / malformed / fallback / Mock 全覆盖 |
| 浏览器 URL 验收 | 11 URL 全部 HTTP 200 |
| Fast Match 精准度 | 24 组测试，15 种不同值 |
| Deep Match 精准度 | 5 样例 ALL PASS，Fast/Deep delta ≤ 9 |
| Demo 案例验收 | 3 经典案例全部达标 |

### 7.3 性能数据

| 指标 | 数值 |
|------|------|
| 前端 Build | 1.51s ~ 3.05s |
| Dist 产物 | 560K |
| Fast Recommend | 50-83ms |
| Fast Match | 4-5ms（ready SEP 场景） |
| SEP Profile 生成 | 18s（含 Hy3 LLM 调用） |

### 7.4 关键设计产出

| 资产类型 | 数量 | 说明 |
|----------|------|------|
| ADR 架构决策 | 4 个 | 技术栈、单主简历、先Mock后模型、不自动生成简历 |
| Prompt 模板 | 4 个 | match_analysis、recommendation、resume_optimization、json_repair |
| JSON Schema | 3 个 | 对应三个业务 Agent |
| 错误码 | 20+ 种 | Pipeline 全覆盖 |
| Quality Gate 维度 | 10 项 | 四级枚举 + revise once |
| 极端情况处理 | 10 种枚举 + 4 级风险 + 7 种策略 |

---

## 八、亮点总结（简历可用提炼）

1. **完整的产品思维**：从阶段 0 只读审计开始，经历 13 个阶段，覆盖产品定位、架构设计、前后端开发、AI Agent 集成、工程部署、文档管理全流程
2. **AI Agent 深度实践**：设计并实现了四主 Agent 架构，完成了 Prompt 工程、LLM Pipeline、JSON Repair、Quality Gate、Mock Fallback 等完整的 LLM 工程化链路
3. **Fast/Deep 双阶段策略**：Fast 秒级语义匹配 + Deep 5 模块 LLM 深度分析，兼顾体验与质量
4. **文档驱动的工程管理**：850+ 文档、550+ 开发日志、逐条变更记录，每个任务有编号、有结论、可追溯
5. **安全与合规设计**：API Key 序列化保护、日志脱敏、.env 不入库、LLM 输出质量门禁、不代写简历的伦理边界
6. **快速迭代**：18 天内完成 13 阶段 550+ 任务，从零搭建到可上线 Demo

---

## 九、进入第 2 阶段分析的准备

第 2 阶段将深入分析：
- 后端架构分层设计（`02-后端设计/`、`04-后端设计/`）
- Agent 系统详细设计（`05-Agent设计/` 35 个文件）
- 数据库设计（`03-数据库设计/`）
- Fast/Deep 引擎实现细节
- LLM Pipeline 实现机制

