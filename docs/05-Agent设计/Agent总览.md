# Agent 总览

> 最后更新：2026-05-29（阶段 4-1.4.4B 登记岗位推荐 Agent 指令模板）

## 当前状态

阶段 4-1.4 进行中，4-1.4.1B 已完成。三个 Mock Agent 仍正常运行。第一个正式 Agent Prompt Template 已登记（`server/prompts/match_analysis_v1.md`），后续将通过离线测试确定真实 Agent 输出质量，不做 MCP 记忆，不做多轮 Agent。

- **RecommendationMockAgent**（`server/internal/agent/recommendation_mock.go`）：基于规则打分生成推荐结果
- **MatchMockAgent**（`server/internal/agent/match_mock.go`）：基于规则打分生成匹配分析结果（含 JD-简历证据链）
- **VersionCompareMockAgent**（`server/internal/agent/version_compare_mock.go`）：基于规则打分生成版本对比结果

阶段 4 目标：在保留 Mock Agent 兜底的基础上，引入 Hunyuan API 作为 LLM Provider，让推荐、匹配分析、版本对比逐步从 Mock 输出升级为真实模型生成的结构化 JSON 输出。推荐 / 匹配 / 版本对比按顺序逐步接入。LLM 调用失败时自动回退 Mock Agent。

## Agent 列表

| Agent | 说明 | 当前状态 | 详细文档 |
|-------|------|---------|---------|
| 简历提取 Agent | 从结构化简历中提取关键信息 | 阶段 3 暂不独立实现 | [简历提取Agent.md](./简历提取Agent.md) |
| 岗位推荐 Agent | 基于简历推荐匹配岗位 | Mock（阶段 3 ✅）/ 真实 LLM（阶段 4-5） | [岗位推荐Agent.md](./岗位推荐Agent.md) |
| 岗位匹配 Agent | 分析简历与岗位的匹配度 | Mock（阶段 3 ✅）/ 真实 LLM（阶段 4-4） | [岗位匹配Agent.md](./岗位匹配Agent.md) |
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

## 阶段 4-1.4 Prompt / Agent 真实接入（进行中）

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
  - 下一步为 4-1.4.6A 人工定稿推荐 JSON Schema

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
