# 2026-05-29-066-LLMClient与Hunyuan接入方案

> 创建时间：2026-05-29
> 所属阶段：4-2.0
> 任务类型：docs only

## 1. 本次任务

阶段 4-2.0：LLM Client 与 Hunyuan 接入方案设计（docs only）。

目标：设计 4-2 阶段整体工程边界，明确 LLM Client 抽象、Hunyuan Client 接入方式、Prompt Loader、Schema Loader、JSON parse + Schema validate、JSON repair once、Mock fallback 等基础设施的设计方向，为后续 4-2.1 ~ 4-2.10 的代码实现提供方案指导。

## 2. 设计范围

本次只做方案文档设计，不写任何 Go 代码、不修改任何配置文件、不创建任何目录或文件。

## 3. 当前前提

### 3.1 已有资产

- 4 个 Prompt 文件（match_analysis / recommendation / resume_optimization / json_repair）
- 3 个 Schema 文件（match_analysis / recommendation / resume_optimization）
- 3 个 Mock Agent（recommendation_mock / match_mock / version_compare_mock）
- 失败兜底与 Mock 回退策略文档
- .env.example 中已有 10 项 LLM 配置
- config.go 已有 LLMConfig 和 FeatureConfig
- server/internal/llm/ 目录已存在但为空

### 3.2 尚未有

- LLM Client 接口
- Hunyuan Client 实现
- Prompt Loader
- Schema Loader
- JSON Schema Validator
- JSON Repair 调用器
- LLM Agent Pipeline

## 4. 新增方案文档

- `docs/05-Agent设计/阶段4-2LLMClient与Hunyuan接入方案.md`

文档覆盖 22 个章节：阶段目标、当前前提、阶段边界、4-2与4-3分工、总体调用链路、目录结构、环境变量配置、LLM Client 抽象、Hunyuan Client 接入、Prompt Loader、Schema Loader、JSON parse与Schema validate、JSONRepairAgent调用链路、Mock fallback集成、generated_by标记、日志与安全、错误码、Makefile脚本、子任务拆分、验收标准、不做事项、下一步建议。

## 5. 工程边界设计

### 5.1 4-2 边界

4-2 只做基础设施层，不改造业务 Agent：

- LLM 配置
- LLM Client 接口 + Hunyuan Client 实现
- Prompt Loader
- Schema Loader
- JSON parse + Schema validate
- JSON repair once
- Mock fallback decision
- Smoke test

### 5.2 4-3 边界

4-3 才把业务 Agent 接入 LLM：

- MatchAnalysisAgent 调用 LLMRunner
- RecommendationAgent 调用 LLMRunner
- ResumeOptimizationAgent 调用 LLMRunner

### 5.3 核心判断

4-2 产出工具，4-3 产出业务价值。4-2 不改业务 Agent，不改接口结构，不改前端。

## 6. 目录结构建议

```
server/internal/llm/
  client.go           # LLMClient 接口 + 工厂
  types.go            # 请求/响应/错误码类型
  hunyuan_client.go   # Hunyuan Client 实现
  runner.go           # LLMRunner（调用→parse→validate→repair→fallback）

server/internal/prompt/
  loader.go           # PromptLoader
  types.go            # Prompt 模板类型

server/internal/schema/
  loader.go           # SchemaLoader
  validator.go        # SchemaValidator
  types.go            # Schema 类型

server/internal/agent/repair/
  json_repair_agent.go # JSONRepairAgent 调用器
```

本次 4-2.0 只设计目录，不创建目录和文件。

## 7. 调用链路设计

```
业务 Agent（4-3 改造）
→ 构造上下文
→ PromptLoader.Load(promptName)
→ SchemaLoader.Load(schemaName)
→ LLMClient.Chat(ctx, req)
→ 提取 JSON → json.Unmarshal
→ SchemaValidator.Validate(data, schema)
→ 失败？→ JSONRepairAgent 调用一次
  → repair output parse + validate
  → 仍失败？→ Mock fallback
→ 通过？→ 返回 LLM 结果
```

4-2 实现从 PromptLoader 到 fallback decision 的基础设施；4-3 把业务 Agent 接入该链路。

## 8. fallback 策略继承

4-2 继承 4-1.4.11 设计的失败兜底与 Mock 回退策略：

- 20 种错误码
- 6 种直接 Mock 场景（LLM_DISABLED / CONFIG_MISSING / TIMEOUT / HTTP_ERROR / RATE_LIMIT / EMPTY_RESPONSE）
- 6 种先 repair 场景（OUTPUT_TRUNCATED / JSON_PARSE_FAILED / MARKDOWN_WRAPPED / SCHEMA_VALIDATION_FAILED / REQUIRED_FIELD_MISSING / ENUM_INVALID）
- generated_by 只有 llm_agent 和 mock_agent
- repair 成功后 generated_by 仍为 llm_agent
- repair 只执行一次

## 9. 4-2 子任务拆分

| 编号 | 任务 |
|------|------|
| 4-2.0 | LLM Client 与 Hunyuan 接入方案（docs only，本次任务） |
| 4-2.1 | LLM / Hunyuan 环境变量配置项 |
| 4-2.2 | Prompt Loader / Schema Loader |
| 4-2.3 | JSON parse + Schema validate 工具 |
| 4-2.4 | LLM Client 抽象接口 |
| 4-2.5 | Hunyuan Client 实现 |
| 4-2.6 | LLM Smoke Test / Makefile 接入 |
| 4-2.7 | JSONRepairAgent 调用链路 |
| 4-2.8 | 统一 LLM Agent Pipeline |
| 4-2.9 | 阶段 4-2 集成验收 |
| 4-2.10 | 阶段 4-2 汇总 |

## 10. 文档更新

### 新增

- `docs/05-Agent设计/阶段4-2LLMClient与Hunyuan接入方案.md`
- `docs/09-开发日志/2026-05-29-066-LLMClient与Hunyuan接入方案.md`

### 修改

- `docs/项目状态.md`
- `docs/下一步任务.md`
- `docs/变更记录.md`
- `docs/说明书.md`
- `docs/05-Agent设计/Agent总览.md`
- `docs/05-Agent设计/LLMClient设计.md`
- `docs/05-Agent设计/Hunyuan接入准备.md`
- `docs/05-Agent设计/Prompt模板.md`
- `docs/05-Agent设计/JSON输出规范.md`
- `docs/05-Agent设计/失败兜底与Mock回退策略.md`

## 11. 未修改内容

- 未修改 server/ 下任何 Go 代码
- 未修改 server/prompts/ 下任何 Prompt 文件
- 未修改 server/prompts/schema/ 下任何 Schema 文件
- 未修改 .env.example
- 未修改 .env
- 未修改 Makefile
- 未修改 scripts/dev/
- 未修改前端
- 未创建 server/internal/llm/ 下任何文件
- 未创建 server/internal/prompt/
- 未创建 server/internal/schema/
- 未创建 server/internal/agent/repair/

## 12. 当前风险

| 风险 | 严重度 | 说明 |
|------|--------|------|
| .env.example 含真实 API Key | 🔴 高 | LLM_API_KEY 行有真实 sk- 值，4-2.1 必须清空 |
| LLM_BASE_URL 不一致 | 🟡 中 | .env.example 为 tokenhub 完整路径，代码中应为 base URL |
| LLM_MAX_OUTPUT_TOKENS=2048 不够 | 🟡 中 | 首轮验收表明需要 ≥8192 |
| LLM_TIMEOUT_SECONDS=20 可能不够 | 🟡 中 | 匹配分析 Prompt token 较多 |
| Go JSON Schema 库需要引入新依赖 | 🟢 低 | 推荐 santhosh-tekuri/jsonschema/v5 |
| config.go FeatureConfig 缺少 EnableLLMAgent 字段 | 🟡 中 | 4-2.1 需补齐 |

## 13. 是否可以进入 4-2.1

✅ 可以进入 4-2.1。

4-2.0 方案文档已完成，工程边界、接口设计方向、目录结构、环境变量配置、子任务拆分均已明确。

4-2.1 的具体工作：

- 修改 config.go 补齐 EnableLLMAgent / EnableLLMFallback / LLMTimeoutSeconds / LLMMaxOutputTokens / LLMTemperature 等字段
- 修改 .env.example 更新 LLM 配置项（清空真实 API Key，调整默认值）
- 更新环境变量文档

## 14. 下一步建议

1. **4-2.1**：LLM / Hunyuan 环境变量配置项（修改 config.go + .env.example）
2. **4-2.2**：Prompt Loader / Schema Loader
3. **4-2.3**：JSON parse + Schema validate 工具
4. **4-2.4**：LLM Client 抽象接口
5. **4-2.5**：Hunyuan Client 实现
