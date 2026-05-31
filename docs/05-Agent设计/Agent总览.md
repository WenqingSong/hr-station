# Agent 总览

> 最后更新：2026-05-31（阶段 4-3.9.1 Agent 输出到前端模块映射方案）

## 当前状态

阶段 4-3 已开始，4-3.0 已完成方案设计，4-3.1 已通过真实 LLM Smoke 前置验收，4-3.2 已完成 MatchAnalysisAgent 接入 Pipeline，4-3.3 已完成匹配分析接口联调与 fallback 验收，4-3.4 已完成 RecommendationAgent 接入 Pipeline，4-3.5 已完成推荐接口联调与 fallback 验收，4-3.5.B 已完成战略差异化与杀手锏确认，4-3.5.C 已完成投递边界与优化停止条件设计，4-3.5.D 已完成 JD 证据链匹配与 START 简历优化规则，4-3.5.E 已完成 Agent 输出 Quality Gate 设计，4-3.5.F 已完成学生求职极端情况处理边界，4-3.5.G 已完成灵活在线简历 Profile 与纯文字 Agent 上下文设计。MatchAnalysisAgent 已接入并通过接口验收。RecommendationAgent 已接入并通过接口验收。ResumeOptimizationAgent 尚未接入。后续业务 Agent 输入层统一为 ResumeContextPayload（ResumeContextText + profile_summary + privacy_flags），替代旧的 `json.Marshal(resumeVersion)`。Agent 输出需从"给结论"升级为"给 JD 证据链和可核实建议"，并明确 apply_readiness 和 optimization_stop_signal。Agent 输出需要通过 Quality Gate 内容质量门禁。极端情况下需识别 candidate_case_type 并采取对应处理策略。4-3.5.H 已完成战略转型阶段汇总与工程评审，评审结论：不立即大改工程，先进入 4-3.6 ResumeOptimizationAgent 现状核对与接入方案。4-3.6 已完成 ResumeOptimizationAgent 现状核对。MatchAnalysisAgent 和 RecommendationAgent 已接入并验收。ResumeOptimizationAgent 当前仅有 Prompt + Schema，缺 6 项核心业务代码，不具备直接接入 Pipeline 条件。4-3.7 已补齐简历优化 Mock 业务链路（Mock Agent / Model / Service / Handler / Route / 接口文档 / 测试）。当前三类业务能力中，匹配分析和岗位推荐已接入 Pipeline，简历优化已具备 Mock 业务链路。4-3.8.0 已完成三类 Agent LLM/fallback 状态核验与验收矩阵冻结，核心发现：ResumeOptimizationAgent 仍 Mock-only，需 4-3.8.1-R 完成 Pipeline 最小集成。

- **RecommendationMockAgent**（`server/internal/agent/recommendation_mock.go`）：基于规则打分生成推荐结果
- **MatchMockAgent**（`server/internal/agent/match_mock.go`）：基于规则打分生成匹配分析结果（含 JD-简历证据链）
- **VersionCompareMockAgent**（`server/internal/agent/version_compare_mock.go`）：基于规则打分生成版本对比结果

阶段 4 目标：在保留 Mock Agent 兜底的基础上，引入 Hunyuan API 作为 LLM Provider，让推荐、匹配分析、版本对比逐步从 Mock 输出升级为真实模型生成的结构化 JSON 输出。推荐 / 匹配 / 版本对比按顺序逐步接入。LLM 调用失败时自动回退 Mock Agent。

## Agent 列表

| Agent | 说明 | 当前状态 | 详细文档 |
|-------|------|---------|---------|
| 简历提取 Agent | 从结构化简历中提取关键信息 | 阶段 3 暂不独立实现 | [简历提取Agent.md](./简历提取Agent.md) |
| 岗位推荐 Agent | 基于简历推荐匹配岗位 | Mock（阶段 3 ✅）/ Pipeline 接入（阶段 4-3.4 ✅）/ 接口验收（阶段 4-3.5 ✅） | [岗位推荐Agent.md](./岗位推荐Agent.md) |
| 岗位匹配 Agent | 分析简历与岗位的匹配度 | Mock（阶段 3 ✅）/ Pipeline 接入（阶段 4-3.2 ✅） | [岗位匹配Agent.md](./岗位匹配Agent.md) |
| 版本对比 Agent | 对比两个版本的差异 | Mock（阶段 3 ✅）/ 真实 LLM（阶段 4-6） | [版本对比Agent.md](./版本对比Agent.md) |

## 设计原则

1. **先 Mock 后接模型**：阶段 3 使用 Mock Agent 跑通闭环，阶段 4 接真实 LLM（参见 ADR-003）
2. **统一接口**：所有 Agent 遵循统一的输入输出接口，方便后续替换
3. **Prompt 管理**：Prompt 模板独立管理，便于调优和版本控制
4. **真实性保护**：Agent 不虚构经历，不替用户编造内容
5. **失败兜底**：LLM 调用失败时自动回退 Mock Agent，不阻断用户操作

## 阶段 4 LLM Agent 接入方向

### 接入策略

- Provider 首选 Hunyuan（腾讯云混元）
- 使用 OpenAI-compatible API，降低接入复杂度
- 低温度（0.2）确保输出稳定
- 强制 JSON 输出 + 字段校验
- 失败自动回退 Mock Agent

### Agent 改造顺序

| 顺序 | Agent | 原因 |
|------|-------|------|
| 1 | 岗位匹配分析 | 核心亮点功能，输入输出最明确 |
| 2 | 岗位推荐 | 与匹配分析逻辑相似 |
| 3 | 简历版本对比 | 输入最复杂，需更精细的 Prompt |

### 开关控制

```
ENABLE_LLM_AGENT=false → 始终使用 Mock Agent
ENABLE_LLM_AGENT=true → 优先使用真实 LLM
ENABLE_LLM_FALLBACK=true → LLM 失败时回退 Mock Agent
```

### 来源标识

响应中通过 `generated_by` 字段区分：
- `"mock_agent"`：Mock Agent 生成
- `"llm_agent"`：真实 LLM 生成
- `"llm_agent_fallback"`：LLM 失败后回退 Mock

## 阶段 3 Mock Agent 约束

| 约束 | 说明 |
|------|------|
| 不调用真实 LLM | Mock Agent 使用固定 JSON 或简单规则生成结果 |
| 不读取 LLM_API_KEY | 不访问 LLM 配置 |
| 不使用真实 Prompt | Prompt 仅作为结构化输出参考 |
| 结果必须结构化 | 输出为固定格式的 JSON 结构 |
| 阶段 4 替换 | 真实 LLM 接入时只需替换 Agent 实现 |

## 阶段 3 Mock Agent 实现

### 岗位推荐 Mock（3-6 实现）

- 文件：`server/internal/agent/recommendation_mock.go`
- 输入：简历版本 + 岗位列表
- 输出：推荐岗位列表 + 推荐理由 + 匹配亮点 + 风险点 + 优先级
- 方式：固定 JSON 或简单规则匹配

### 岗位匹配 Mock（3-7 实现）

- 文件：`server/internal/agent/match_mock.go`
- 输入：简历版本 + 目标岗位 JD
- 输出：匹配分数 + 初筛判断 + JD-简历证据链 + 风险点 + P0/P1 优化建议 + 下一步建议 + 真实性提示
- 方式：固定 JSON 或简单规则生成

### 版本对比 Mock（3-8 实现）

- 文件：`server/internal/agent/version_compare_mock.go`
- 输入：旧版本简历 + 新版本简历
- 输出：分数变化 + 字段变化 + 风险变化 + 建议完成情况 + 总结
- 方式：字段差异对比 + 固定评价

### 简历提取 Agent（阶段 3 暂不独立实现）

- 简历信息提取功能（关键技能标签、求职方向、表达风险）在阶段 3 暂不独立实现
- 阶段 3 简历详情页直接展示结构化数据
- 阶段 4 接入真实 LLM 后再实现简历提取 Agent

## 阶段 4-1.3 腾讯岗位数据适配（已完成）

4-1.3 已完成，腾讯岗位后端适配闭环已形成。

## 阶段 4-2 LLM Client 抽象与 Hunyuan Client 接入（已完成）

4-2 已正式完成，LLM 基础设施已闭环。4-3 已开始，MatchAnalysisAgent / RecommendationAgent / ResumeOptimizationAgent 将逐步接入 Pipeline。

## 阶段 4-3 业务 Agent 接入真实 LLM（进行中）

4-3.0 已完成业务 Agent 接入 LLM 方案设计（docs only），方案文档见 `docs/05-Agent设计/阶段4-3业务Agent接入LLM方案.md`。

关键方案决策：

1. **业务 Agent 依赖 Pipeline Runner**：Service 层注入 pipelineRunner，不直接依赖 HunyuanClient
2. **PromptValues 由 Service 层组装**：Service 从数据库读取简历/岗位后组装 PromptValues，传给 Pipeline
3. **fallback_required 由 Pipeline 返回**：Pipeline 只返回决策，业务 Agent 调用具体 Mock Agent
4. **generated_by 规则**：success/repaired → llm_agent，fallback_required → mock_agent
5. **RunResult.Parsed → Go Model 映射**：map[string]any → typed struct，映射失败触发 Mock fallback
6. **ResumeOptimizationAgent 无业务代码**：4-3.6 仅做状态评估，不强行造新接口

4-3 子任务（4-3.0 ~ 4-3.9）：

|| 编号 | 任务 | 状态 |
||------|------|------|
|| 4-3.0 | 业务 Agent 接入 LLM 方案 | ✅ 已完成 |
|| 4-3.1 | 真实 LLM Smoke Test | ✅ 已通过 |
|| 4-3.2 | MatchAnalysisAgent 接入 Pipeline | ✅ 已完成 |
|| 4-3.3 | 匹配分析接口联调与 fallback 验收 | ✅ 已完成 |
|| 4-3.4 | RecommendationAgent 接入 Pipeline | ⬜ 待执行 |
|| 4-3.5 | 推荐接口联调与 fallback 验收 | ⬜ 待执行 |
|| 4-3.6 | ResumeOptimizationAgent 状态评估 | ⬜ 待执行 |
|| 4-3.7 | RunResult 字段策略落地 | ⬜ 待执行 |
|| 4-3.8 | 三类业务 Agent 集成验收 | ⬜ 待执行 |
|| 4-3.8.0 | 三类 Agent LLM/fallback 状态核验与验收矩阵冻结 | ✅ 已完成 |
|| 4-3.8.1-R | ResumeOptimizationAgent Pipeline 最小集成 | ⬜ 待执行 |
|| 4-3.9 | 阶段 4-3 汇总 | ⬜ 待执行 |

4-2.8 已完成统一 LLM Agent Pipeline：
- 4-2.0 已完成：LLM Client 与 Hunyuan 接入方案（docs only）
- 4-2.1 已完成：LLM / Hunyuan 环境变量配置项（.env.example 已新增 14 项 LLM_* 配置）
- 4-2.2 已完成：Prompt Loader / Schema Loader（可加载 4 个 Prompt + 3 个 Schema）
- 4-2.3 已完成：JSON parse + Schema validate 工具（llmjson 包 + schema.Validator）
- 4-2.4 已完成：LLM Client 抽象接口（Client + ChatRequest/ChatResponse + Config + ErrorCode + NoopClient）
- 4-2.5 已完成：Hunyuan Client 实现（OpenAI-compatible API，httptest 22 项测试）
- 4-2.6 已完成：LLM Smoke Test / Makefile 接入
- 4-2.7 已完成：JSONRepairAgent 调用链路（repair once + 14 项单元测试）
- 4-2.8 已完成：统一 LLM Agent Pipeline（Prompt→LLM→parse→validate→repair→fallback，18 项单元测试）
- 方案文档见 `docs/05-Agent设计/阶段4-2LLMClient与Hunyuan接入方案.md`
- 4-2 目标不是直接完成业务 Agent 的 LLM 化，而是建立统一的大模型基础设施层
- 4-2 产出：LLMClient 抽象 / HunyuanClient / PromptLoader / SchemaLoader / SchemaValidator / JSONRepairAgent 调用器 / Pipeline Runner / fallback decision
- 4-3 才把具体业务 Agent 接入 LLM

## 阶段 4-1.4 Prompt / Agent 真实接入（已完成）

4-1.4 开始设计 Prompt 和 Agent 的真实接入方案：

- 4-1.4.0 已完成：Prompt / Agent 真实接入方案已确定
- 4-1.4.1A 已完成：人工定稿匹配分析 Agent 指令模板
- 4-1.4.1B 已完成：CodeBuddy 登记匹配分析 Agent 指令模板
- 第一个正式 Agent Prompt Template 已登记：`server/prompts/match_analysis_v1.md`
- 当前 Agent 仍为 Mock Agent，当前仍未接 Hunyuan，当前仍未实现 LLM Client，当前仍未替换 Mock Agent
- Agent 采用 Stateless 单次请求模式，不做 MCP 记忆，不做多轮 Agent
- Prompt 将作为独立产品资产管理，目录 `server/prompts/`
- 4-1.4 优先打磨匹配分析 Prompt（P0），再设计推荐 Prompt（P1）
- 离线测试使用真实腾讯岗位 JD + 简历样本
- 失败兜底策略已设计：LLM 失败回退 Mock Agent，generated_by 区分来源
- 不接 Hunyuan / LLM Client（4-2 才实现）
- 方案文档见 `docs/05-Agent设计/阶段4-1.4Prompt与Agent真实接入方案.md`
- **4-1.4.2 已完成：匹配分析 Prompt 首轮输出验收（partial_pass）**
  - 模型输出 JSON 合法，17 个顶层字段完整，枚举值全部合法
  - 发现问题：rewrite_example 中少量具体数字超出简历事实，证据链缺少 missing 判断
  - 问题不阻塞进入 4-1.4.3A JSON Schema 定稿
- **4-1.4.3A/B 已完成：匹配分析 JSON Schema 定稿与登记**
  - 匹配分析 Agent 当前已有第一个完整 Prompt + Schema 资产组合：
    - `match_analysis_v1.md`
    - `match_analysis_schema.json`
  - Schema 已吸收 4-1.4.2 首轮验收问题
  - 当前仍未接 Hunyuan
  - 当前仍未实现 LLM Client
  - 当前仍未替换 Mock Agent
  - 下一步为 4-1.4.4A 人工定稿岗位推荐 Agent 指令模板
- **4-1.4.4A/B 已完成：岗位推荐 Agent 指令模板定稿与登记**
  - 第二个正式 Agent Prompt Template 已登记：`server/prompts/recommendation_v1.md`
  - 当前已有：
    - 匹配分析 Prompt + Schema（match_analysis_v1.md + match_analysis_schema.json）
    - 岗位推荐 Prompt（recommendation_v1.md）
  - 当前仍未接 Hunyuan
  - 当前仍未实现 LLM Client
  - 当前仍未替换 Mock Agent
  - **4-1.4.5 已完成：推荐 Prompt 首轮输出验收（partial_pass）**
    - 模型输出 JSON 合法，7 个顶层字段完整，4 个推荐项各含 19 个必要字段
    - 证据链真实，Offer鹅语气专业，分数克制（88/42/35/33）
    - 发现轻微问题：四档推荐覆盖不完整，不阻塞 Schema 定稿
    - 验收报告见 `docs/05-Agent设计/推荐Prompt首轮输出验收报告.md`
  - **4-1.4.6A/B 已完成：推荐 JSON Schema 定稿与登记**
    - 推荐 Agent 当前已有完整 Prompt + Schema 资产组合：
      - `recommendation_v1.md`
      - `recommendation_schema.json`
    - 当前已有：
      - 匹配分析 Prompt + Schema
      - 岗位推荐 Prompt + Schema
    - 当前仍未接 Hunyuan
    - 当前仍未实现 LLM Client
    - 当前仍未替换 Mock Agent
  - **4-1.4.7A/B 已完成：简历优化 Agent 指令模板定稿与登记**
    - 第三个正式 Agent Prompt Template 已登记：`server/prompts/resume_optimization_v1.md`
    - 当前已有：
      - 匹配分析 Prompt + Schema（match_analysis_v1.md + match_analysis_schema.json）
      - 岗位推荐 Prompt + Schema（recommendation_v1.md + recommendation_schema.json）
      - 简历优化 Prompt（resume_optimization_v1.md）
    - 当前仍未接 Hunyuan
    - 当前仍未实现 LLM Client
    - 当前仍未替换 Mock Agent
  - **4-1.4.8 已完成：简历优化 Prompt 首轮输出验收（partial_pass）**
    - 模型输出 JSON 合法，15 个顶层字段完整
    - P0/P1/P2 优化建议具体且围绕目标岗位
    - keyword_alignment 正确区分 present/partial/missing
    - rewrite_examples 全部 requires_user_verification=true
    - 验收报告见 `docs/05-Agent设计/简历优化Prompt首轮输出验收报告.md`
  - **4-1.4.9A/B 已完成：简历优化 JSON Schema 定稿与登记**
    - 简历优化 Agent 当前已有完整 Prompt + Schema 资产组合：
      - `resume_optimization_v1.md`
      - `resume_optimization_schema.json`
    - 当前已有：
      - 匹配分析 Prompt + Schema（match_analysis_v1.md + match_analysis_schema.json）
      - 岗位推荐 Prompt + Schema（recommendation_v1.md + recommendation_schema.json）
      - 简历优化 Prompt + Schema（resume_optimization_v1.md + resume_optimization_schema.json）
    - 当前仍未接 Hunyuan
    - 当前仍未实现 LLM Client
    - 当前仍未替换 Mock Agent
    - 下一步为 4-1.4.10A 人工定稿 JSON 修复 Agent 指令模板
  - **4-1.4.10A/B 已完成：JSON 修复 Agent 指令模板定稿与登记**
    - 第四个正式 Agent Prompt Template 已登记：`server/prompts/json_repair_v1.md`
    - JSONRepairAgent 定位为结构修复器，不是业务生成器
    - 当前已有：
      - 匹配分析 Prompt + Schema（match_analysis_v1.md + match_analysis_schema.json）
      - 岗位推荐 Prompt + Schema（recommendation_v1.md + recommendation_schema.json）
      - 简历优化 Prompt + Schema（resume_optimization_v1.md + resume_optimization_schema.json）
      - JSON 修复 Prompt（json_repair_v1.md）
    - 当前仍未接 Hunyuan
    - 当前仍未实现 LLM Client
    - 当前仍未替换 Mock Agent
    - 下一步为 4-1.4.11 失败兜底与 Mock 回退策略
  - **4-1.4.11 已完成：失败兜底与 Mock 回退策略**
    - 当前已有三类业务 Agent Prompt + Schema，以及 JSON Repair Prompt
    - 后续真实接入时采用：业务 Agent LLM 输出 → Schema 校验 → JSON Repair 一次 → Mock fallback
    - 策略文档见 `docs/05-Agent设计/失败兜底与Mock回退策略.md`
    - 当前仍未接 Hunyuan
    - 当前仍未实现 LLM Client
    - 当前仍未替换 Mock Agent
    - 下一步为 4-1.4.12 阶段 4-1.4 总验收
  - **4-1.4.12 已完成：阶段 4-1.4 总验收**
    - Prompt 资产、Schema 资产、首轮输出验收、JSON Repair、fallback 策略均已通过总验收
    - 当前已有：
      - MatchAnalysisAgent Prompt + Schema
      - RecommendationAgent Prompt + Schema
      - ResumeOptimizationAgent Prompt + Schema
      - JSONRepairAgent Prompt
      - 失败兜底与 Mock 回退策略
    - 验收报告见 `docs/05-Agent设计/阶段4-1.4总验收报告.md`
    - 当前仍未接 Hunyuan
    - 当前仍未实现 LLM Client
    - 当前仍未替换 Mock Agent
    - 下一步为 4-1.4.13 阶段 4-1.4 汇总
  - **4-1.4.13 已完成：阶段 4-1.4 汇总**
    - 阶段 4-1.4 已正式闭环
    - 汇总文档见 `docs/00-项目总览/阶段4-1.4Prompt与Agent真实接入汇总.md`
    - 当前 Agent 资产体系已形成：
      - MatchAnalysisAgent Prompt + Schema
      - RecommendationAgent Prompt + Schema
      - ResumeOptimizationAgent Prompt + Schema
      - JSONRepairAgent Prompt
      - Mock fallback 策略
    - 当前仍未接 Hunyuan
    - 当前仍未实现 LLM Client
    - 当前仍未替换 Mock Agent
    - 下一阶段为 4-2：LLM Client 抽象与 Hunyuan Client 接入

## 阶段 4-3.4 更新

4-3.4 已完成：RecommendationAgent 接入 Pipeline。

- MatchAnalysisAgent 已接入并通过接口验收（4-3.2 + 4-3.3）
- RecommendationAgent 已接入 Pipeline（4-3.4），推荐链路已从纯 Mock 升级为 LLM Pipeline 优先 + Mock fallback
- 推荐结果必须来自候选岗位集，不允许模型编造岗位
- ResumeOptimizationAgent 尚未接入
- 下一步为 4-3.5 推荐接口联调与 fallback 验收

## 关联文档

- [阶段4Agent与模型接入方案.md](./阶段4Agent与模型接入方案.md)
- [Hunyuan接入准备.md](./Hunyuan接入准备.md)
- [LLMClient设计.md](./LLMClient设计.md)
- [JSON输出规范.md](./JSON输出规范.md)
- [Prompt模板.md](./Prompt模板.md)
- [模型接入.md](./模型接入.md)
- [岗位推荐Agent.md](./岗位推荐Agent.md)
- [岗位匹配Agent.md](./岗位匹配Agent.md)
- [版本对比Agent.md](./版本对比Agent.md)
- [../08-架构决策记录/ADR-003-先Mock后接模型.md](../08-架构决策记录/ADR-003-先Mock后接模型.md)
- [../02-后端设计/阶段3核心业务后端Mock方案.md](../02-后端设计/阶段3核心业务后端Mock方案.md)
