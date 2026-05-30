# 阶段 4-2 LLM 基础设施汇总

> 创建时间：2026-05-30
> 所属阶段：4-2.10
> 结论：✅ 阶段 4-2 已完成

**4-3 已开始**：4-3.0 已完成业务 Agent 接入 LLM 方案设计（docs only），方案文档见 `docs/05-Agent设计/阶段4-3业务Agent接入LLM方案.md`。

## 1. 阶段结论

阶段 4-2 已完成。

LLM Client 抽象、HunyuanClient、Prompt Loader、Schema Loader、JSON parse、Schema validate、JSONRepairAgent、统一 Pipeline、LLM smoke test 入口和 fallback_required 决策均已具备。

当前仍未接业务 Agent。
当前仍未修改推荐 / 匹配 / 简历优化接口。
当前仍未改前端。

## 2. 阶段目标回顾

阶段 4-2 的目标不是让业务接口直接调用模型，而是：

1. 先建立统一的大模型基础设施层
2. 让业务 Agent 后续不直接依赖 Hunyuan
3. 让 Prompt / Schema 文件资产真正进入工程链路
4. 让模型输出先经过 JSON parse + Schema validate
5. 让异常输出可以 repair once
6. 让失败情况可以返回 fallback_required 决策
7. 为 4-3 业务 Agent 接入真实 LLM 做准备

核心判断：**4-2 产出工具，4-3 产出业务价值。**

## 3. 已完成子任务

| 编号 | 任务 | 状态 | 关键产物 | 备注 |
|------|------|------|---------|------|
| 4-2.0 | LLM Client 与 Hunyuan 接入方案 | ✅ 已完成 | 方案文档 | docs only |
| 4-2.1 | LLM / Hunyuan 环境变量配置项 | ✅ 已完成 | .env.example 14 项 LLM_* 配置 | 安全修复：清除遗留 API Key |
| 4-2.2 | Prompt Loader / Schema Loader | ✅ 已完成 | prompt/loader.go + schema/loader.go | 4 Prompt + 3 Schema |
| 4-2.3 | JSON parse + Schema validate 工具 | ✅ 已完成 | llmjson/ + schema/validator.go | santhosh-tekuri/jsonschema/v5 |
| 4-2.4 | LLM Client 抽象接口 | ✅ 已完成 | llm/client.go + types + Config + ErrorCode + NoopClient | 27 项测试 |
| 4-2.5 | Hunyuan Client 实现 | ✅ 已完成 | llm/hunyuan_client.go + hunyuan_types.go | httptest 22 项测试 |
| 4-2.6 | LLM Smoke Test / Makefile 接入 | ✅ 已完成 | cmd/llm-smoke + llm_smoke.sh + Makefile llm-* | 真实 smoke 需 RUN_REAL_LLM_SMOKE=1 |
| 4-2.7 | JSONRepairAgent 调用链路 | ✅ 已完成 | agent/repair/ | repair once + 14 项测试 |
| 4-2.8 | 统一 LLM Agent Pipeline | ✅ 已完成 | agent/pipeline/ | 18 项测试 |
| 4-2.9 | 阶段 4-2 集成验收 | ✅ 已完成 | 阶段4-2集成验收报告.md | 全部通过 |
| 4-2.10 | 阶段 4-2 汇总 | ✅ 已完成 | 本文档 | docs only |

## 4. LLM 基础设施资产汇总

| 资产 | 所属目录 | 用途 | 已测试 | 已接业务 Agent |
|------|---------|------|--------|---------------|
| Prompt Loader | server/internal/prompt/ | 从 server/prompts/ 加载 4 个 Prompt | ✅ | ❌ |
| Schema Loader | server/internal/schema/ | 从 server/prompts/schema/ 加载 3 个 Schema | ✅ | ❌ |
| Schema Validator | server/internal/schema/ | JSON Schema 校验（draft 2020-12） | ✅ | ❌ |
| llmjson Extract / Parse | server/internal/llmjson/ | 从 raw output 提取 JSON + parse | ✅ | ❌ |
| llm.Client 抽象 | server/internal/llm/ | 统一 LLM 调用接口 | ✅ | ❌ |
| NoopClient | server/internal/llm/ | 测试实现，Enable=false 返回 LLM_DISABLED | ✅ | ❌ |
| HunyuanClient | server/internal/llm/ | OpenAI-compatible chat completions | ✅ httptest | ❌ |
| LLM Smoke Test | server/cmd/llm-smoke/ | 最小 smoke test Go 命令 | ✅ | ❌ |
| JSONRepairAgent | server/internal/agent/repair/ | repair once 调用链路 | ✅ | ❌ |
| Pipeline Runner | server/internal/agent/pipeline/ | Prompt→LLM→parse→validate→repair→fallback | ✅ | ❌ |
| fallback_required decision | server/internal/agent/pipeline/ | 只返回 fallback 决策，不调用 Mock Agent | ✅ | ❌ |

当前所有基础设施均未直接接业务 Agent。

## 5. 工程目录与模块汇总

| 目录 / 文件 | 职责 |
|------------|------|
| server/internal/prompt/ | Prompt Loader：从 server/prompts/ 加载 .md 模板，支持占位符替换 |
| server/internal/schema/ | Schema Loader + Validator：从 server/prompts/schema/ 加载 JSON Schema，使用 santhosh-tekuri/jsonschema/v5 校验 |
| server/internal/llmjson/ | JSON 提取 / Parse：从 LLM raw output 提取 JSON object，支持 Markdown code fence、前后解释文字、BOM 清理 |
| server/internal/llm/ | LLM Client 抽象：Client 接口 + ChatRequest/ChatResponse + Config + ErrorCode + LLMError + NoopClient + HunyuanClient |
| server/internal/agent/repair/ | JSONRepairAgent：repair once 调用链路，通过 llm.Client 调用模型 |
| server/internal/agent/pipeline/ | Pipeline Runner：Prompt Render → LLM Call → JSON parse → Schema validate → repair once → fallback decision |
| server/cmd/llm-smoke/ | 最小 LLM smoke test Go 命令：验证 HunyuanClient 连通性 |
| scripts/dev/llm_smoke.sh | LLM smoke test 脚本：help / check / smoke |
| Makefile | llm-help / llm-check / llm-smoke target |

## 6. 环境变量与安全策略汇总

4-2.1 已固化 14 项 LLM_* 配置：

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| LLM_ENABLE | false | 是否启用 LLM |
| LLM_PROVIDER | hunyuan | 模型提供商 |
| LLM_BASE_URL | （空） | OpenAI-compatible API 地址 |
| LLM_API_KEY | （空） | API Key，只放在 .env |
| LLM_MODEL | （空） | 模型名称 |
| LLM_TEMPERATURE | 0.2 | 温度，低温度保稳定 |
| LLM_MAX_OUTPUT_TOKENS | 8192 | 最大输出 token |
| LLM_TIMEOUT_SECONDS | 30 | 请求超时（秒） |
| LLM_REPAIR_ENABLE | true | 是否启用 JSON repair |
| LLM_REPAIR_MAX_ATTEMPTS | 1 | repair 最大尝试次数（固定为 1） |
| LLM_FALLBACK_ENABLE | true | 是否启用 Mock fallback |
| LLM_LOG_RAW_OUTPUT | false | 是否记录原始输出（调试用） |
| LLM_PROMPT_DIR | server/prompts | Prompt 文件目录 |
| LLM_SCHEMA_DIR | server/prompts/schema | Schema 文件目录 |

安全策略强调：

1. LLM_API_KEY 只能写入本地 .env
2. .env.example 必须为空值或占位符
3. docs / 日志 / Git / 聊天记录不得记录真实 API Key
4. LLM_ENABLE 默认 false
5. LLM_FALLBACK_ENABLE 默认 true
6. LLM_REPAIR_MAX_ATTEMPTS 固定为 1
7. LLM_LOG_RAW_OUTPUT 默认 false

## 7. Prompt / Schema 加载能力汇总

### Prompt Loader 支持

| PromptName | 文件路径 |
|-----------|---------|
| match_analysis | server/prompts/match_analysis_v1.md |
| recommendation | server/prompts/recommendation_v1.md |
| resume_optimization | server/prompts/resume_optimization_v1.md |
| json_repair | server/prompts/json_repair_v1.md |

### Schema Loader 支持

| SchemaName | 文件路径 |
|-----------|---------|
| match_analysis | server/prompts/schema/match_analysis_schema.json |
| recommendation | server/prompts/schema/recommendation_schema.json |
| resume_optimization | server/prompts/schema/resume_optimization_schema.json |

关键说明：

- 后续业务 Agent 不需要把 Prompt 写死在代码里
- 后续业务 Agent 只需要传 PromptName、SchemaName 和 PromptValues
- json_repair 不需要独立业务 Schema，它使用目标业务 Schema 作为修复目标

## 8. JSON parse / Schema validate 能力汇总

LLM raw output 不允许直接返回业务接口。必须经过：

```
raw output
→ extract JSON object（去除 Markdown code fence / BOM / 前后解释文字）
→ json.Unmarshal
→ top-level object check
→ Schema validate
```

支持的异常识别：

| 错误码 | 触发条件 |
|--------|---------|
| JSON_EMPTY | 空字符串 |
| JSON_OBJECT_NOT_FOUND | 未找到 JSON object |
| JSON_PARSE_FAILED | json.Unmarshal 失败 |
| JSON_TOP_LEVEL_NOT_OBJECT | 顶层为 array 而非 object |
| SCHEMA_VALIDATION_FAILED | Schema 校验失败 |

## 9. LLM Client / HunyuanClient 能力汇总

- **llm.Client 是统一模型调用抽象**：业务 Agent 后续只依赖 llm.Client，不直接依赖 Hunyuan
- **HunyuanClient 是 llm.Client 的 provider 实现**：基于 OpenAI-compatible chat completions
- 支持 endpoint normalization（/v1、/v1/、/v1/chat/completions）
- 支持 Authorization Bearer
- 支持 timeout / HTTP error / rate limit / empty response 错误处理
- HTTP 429→LLM_RATE_LIMIT、5xx→LLM_HTTP_ERROR、401/403→LLM_API_KEY_MISSING
- 使用 httptest 完成验证（22 项测试）
- finish_reason=length 不视为硬错误，返回 ChatResponse，Pipeline 决定是否 repair
- 不在错误中泄露 API Key

真实 Hunyuan smoke test 入口已存在，但本阶段集成验收中真实 smoke 未必执行，取决于本地 .env 与 RUN_REAL_LLM_SMOKE=1。

## 10. LLM Smoke Test 能力汇总

| 命令 | 说明 |
|------|------|
| make llm-help | 显示 LLM 相关命令帮助 |
| make llm-check | 检查 LLM 配置（不请求外部模型） |
| make llm-smoke | 执行最小 smoke test（需 RUN_REAL_LLM_SMOKE=1 才触发真实调用） |

关键约束：

- make llm-check 不请求外部模型
- make llm-smoke 默认不请求真实模型
- RUN_REAL_LLM_SMOKE=1 make llm-smoke 才触发一次真实 smoke
- smoke test 使用最小 Prompt
- 不携带简历、JD、岗位数据或用户隐私
- 不输出 API Key
- 不输出完整 raw output

## 11. JSONRepairAgent 能力汇总

- 已接入 json_repair_v1.md
- 支持 5 个占位符替换：TARGET_OUTPUT_TYPE / TARGET_SCHEMA_JSON / BROKEN_MODEL_OUTPUT / VALIDATION_ERRORS / REPAIR_CONTEXT
- 通过 llm.Client 调用模型，不直接依赖 HunyuanClient
- repair 后重新 extract / parse / schema validate
- repair 只执行一次
- repair 失败不递归
- repair 失败不重试
- repair 失败交给后续 Pipeline fallback_required 决策
- repair 成功后 generated_by 仍为 llm_agent

7 个 repair 错误码：

| 错误码 | 含义 |
|--------|------|
| JSON_REPAIR_FAILED | repair 通用失败 |
| JSON_REPAIR_PARSE_FAILED | repair output parse 失败 |
| JSON_REPAIR_SCHEMA_FAILED | repair output schema validate 失败 |
| JSON_REPAIR_LLM_FAILED | repair LLM 调用失败 |
| JSON_REPAIR_PROMPT_LOAD_FAILED | repair Prompt 加载失败 |
| JSON_REPAIR_SCHEMA_LOAD_FAILED | repair Schema 加载失败 |
| JSON_REPAIR_TARGET_INVALID | target output type 无效 |

## 12. 统一 Pipeline 能力汇总

完整链路：

```
PromptName + SchemaName + PromptValues
→ Prompt Loader
→ Prompt Render
→ Schema Loader
→ llm.Client.Chat
→ ChatResponse.RawContent
→ llmjson Extract / Parse
→ Schema validate
→ 如果失败：JSONRepairAgent repair once
  → repair output Extract / Parse
  → repair output Schema validate
→ AgentStatus: success / repaired / fallback_required
```

关键约束：

- Pipeline 只做 fallback_required decision
- Pipeline 不调用具体 Mock Agent
- 具体 Mock fallback 放到 4-3 业务 Agent 接入阶段

三种状态：

| AgentStatus | 触发条件 | generated_by |
|------------|---------|-------------|
| success | 首次 parse + validate 通过 | llm_agent |
| repaired | repair 成功 | llm_agent |
| fallback_required | LLM 错误或 repair 失败 | mock_agent |

禁止出现的 generated_by 值：llm_agent_fallback、json_repair_agent

20+ Pipeline 错误码覆盖：PIPELINE_REQUEST_INVALID / PROMPT_LOAD_FAILED / PROMPT_RENDER_FAILED / SCHEMA_LOAD_FAILED / LLM_* 系列 / JSON_* 系列 / SCHEMA_VALIDATION_FAILED / JSON_REPAIR_* 系列 / TARGET_OUTPUT_TYPE_INVALID / UNKNOWN_ERROR

## 13. 集成验收结果汇总

4-2.9 集成验收结论：**全部通过**。

| 验收项 | 结果 |
|--------|------|
| 基础设施资产齐全 | ✅ 11 项资产全部存在 |
| Prompt / Schema 加载通过 | ✅ 4 Prompt + 3 Schema |
| JSON parse / Schema validate 通过 | ✅ 9 种 JSON 场景 + 4 种错误码 |
| LLM Client / HunyuanClient 通过 | ✅ httptest 22 项 |
| llm-smoke 入口通过 | ✅ make llm-help / llm-check 可用 |
| JSONRepairAgent 通过 | ✅ 14 项测试 |
| Pipeline 通过 | ✅ 18 项测试 |
| fallback_required 决策通过 | ✅ 11 种失败场景覆盖 |
| 安全边界合规 | ✅ 未泄露 API Key |
| 核心模块测试 | ✅ 132 项全部通过 |
| go test ./... | ✅ 全部通过 |
| go vet ./... | ✅ 无警告 |
| 真实 Hunyuan smoke | ❌ 未执行（LLM_ENABLE=false） |

## 14. 当前未完成内容

1. 业务 Agent 尚未接入真实 LLM
2. MatchAnalysisAgent 尚未接 Pipeline
3. RecommendationAgent 尚未接 Pipeline
4. ResumeOptimizationAgent 尚未接 Pipeline
5. 具体 Mock fallback 尚未在业务 Agent 中落地
6. 推荐 / 匹配 / 简历优化接口尚未使用 Pipeline
7. 前端尚未展示 agent_status / repair_applied / fallback_required
8. 真实业务 Prompt 质量仍需在 4-3 / 4-4 中联调
9. 真实 Hunyuan smoke 如未执行，需要后续补执行

## 15. 当前风险与注意事项

1. LLM API Key 不能写入 docs、日志、.env.example、Git 或聊天记录
2. 业务 Agent 接入前应先执行一次真实 smoke
3. Pipeline 当前只返回 fallback_required，不产生具体 Mock 结果
4. 4-3 接业务 Agent 时不能绕过 Pipeline 直接调用 HunyuanClient
5. 4-3 接业务 Agent 时要控制 Prompt 输入长度
6. 4-3 接业务 Agent 时要避免把完整简历、完整 JD 写入日志
7. 前端展示字段需要在 4-4 统一收口
8. 如果 Hunyuan 输出稳定性不足，必须保留 Mock fallback

## 16. 与前序阶段的关系

| 前序阶段 | 提供的能力 |
|---------|----------|
| 4-1.2 | 提供腾讯招聘原始数据接入 |
| 4-1.3 | 提供腾讯岗位入库和后端岗位接口适配 |
| 4-1.4 | 提供 Prompt / Schema / JSON Repair Prompt / fallback 策略 |
| 4-2 | 把这些资产转化为可复用的 LLM 基础设施 |

## 17. 对后续 4-3 的意义

4-3 不需要从零实现模型调用；4-3 不需要重新设计 Prompt Loader；4-3 不需要重新实现 Schema 校验；4-3 不需要重新实现 JSON repair；4-3 只需要让业务 Agent 构造 PromptValues，然后调用统一 Pipeline；4-3 需要把 fallback_required 和具体 Mock Agent 结合起来。

建议 4-3 拆分方向：

| 编号 | 任务 | 说明 |
|------|------|------|
| 4-3.0 | 业务 Agent 接入 LLM 方案 | docs only |
| 4-3.1 | MatchAnalysisAgent 接入 Pipeline | 优先接入 |
| 4-3.2 | 匹配分析接口联调与 fallback 验收 | |
| 4-3.3 | RecommendationAgent 接入 Pipeline | |
| 4-3.4 | 推荐接口联调与 fallback 验收 | |
| 4-3.5 | ResumeOptimizationAgent 接入 Pipeline 或预留 | |
| 4-3.6 | 三类业务 Agent 集成验收 | |
| 4-3.7 | 阶段 4-3 汇总 | |

## 18. 是否可以进入 4-3

可以进入 4-3 准备阶段。

但下一步应先做 4-3.0 docs only 接入方案，不要直接改业务 Agent。

## 19. 下一步建议

1. 阶段 4-2 已完成
2. 4-3.0 已完成：业务 Agent 接入 LLM 方案（docs only）
3. 下一阶段进入 4-3.1：真实 LLM Smoke Test
4. 优先接 MatchAnalysisAgent（4-3.2）
5. 再接 RecommendationAgent（4-3.4）
6. 最后处理 ResumeOptimizationAgent（4-3.6 状态评估）
