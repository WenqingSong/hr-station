# Hunyuan 接入准备

> 最后更新：2026-05-30（阶段 4-3.0 业务 Agent 接入 LLM 方案）

## 1. 用户需要准备的信息

| 配置项 | 是否必填 | 填写位置 | 说明 | 示例 |
|--------|---------|---------|------|------|
| `LLM_PROVIDER` | 是 | .env | 模型提供商 | `hunyuan` |
| `LLM_BASE_URL` | 是 | .env | Hunyuan API Base URL | `https://api.hunyuan.cloud.tencent.com/v1` |
| `LLM_API_KEY` | 是 | .env | 腾讯云混元 API Key，只填本地 .env | 不要写入文档 |
| `LLM_MODEL` | 是 | .env | 可用模型名称 | `hunyuan-turbos-latest` |
| `LLM_TIMEOUT_SECONDS` | 否 | .env | 请求超时秒数 | `20` |
| `LLM_MAX_RETRIES` | 否 | .env | 失败重试次数 | `1` |
| `LLM_MAX_OUTPUT_TOKENS` | 否 | .env | 最大输出 token | `2048` |
| `LLM_TEMPERATURE` | 否 | .env | 温度 | `0.2` |
| `ENABLE_LLM_AGENT` | 是 | .env | 是否启用真实 LLM | `false` / `true` |
| `ENABLE_LLM_FALLBACK` | 是 | .env | 是否失败回退 Mock | `true` |

## 2. 用户需要完成的腾讯云操作

1. 登录腾讯云控制台
2. 开通混元大模型服务
3. 确认是否有免费额度
4. 创建或获取 API Key
5. 确认可用模型名称（建议 `hunyuan-turbos-latest`）
6. 确认 API Base URL（建议 `https://api.hunyuan.cloud.tencent.com/v1`）
7. 将 Key 和模型配置填写到本地 `.env`
8. 不要把 `.env` 提交到 Git

## 3. 用户填写后的本地 .env 示例

```env
ENABLE_LLM_AGENT=true
ENABLE_LLM_FALLBACK=true
LLM_PROVIDER=hunyuan
LLM_BASE_URL=https://api.hunyuan.cloud.tencent.com/v1
LLM_API_KEY=请填写你的真实Key
LLM_MODEL=hunyuan-turbos-latest
LLM_TIMEOUT_SECONDS=20
LLM_MAX_RETRIES=1
LLM_MAX_OUTPUT_TOKENS=2048
LLM_TEMPERATURE=0.2
```

**重要提醒**：
- `LLM_API_KEY` 只能填写在本地 `.env` 中
- 不要将真实 Key 写入 `.env.example`
- 不要将 `.env` 提交到代码仓库
- 如果 `.env` 已被 git 追踪，需执行 `git rm --cached .env`

## 4. 用户填写完成后的验证方式

本次（4-1）不执行验证。下一步 4-2 实现 LLM Client 后可验证：

1. 检查 `.env` 是否存在
2. 检查 `LLM_API_KEY` 是否非空
3. 使用 LLM Client 测试 `/v1/chat/completions` 接口
4. 要求模型返回固定 JSON 格式
5. JSON 解析成功后才允许进入业务 Agent 改造

## 5. Hunyuan API 调用示例

### 5.1 curl 测试

```bash
curl -s -X POST https://api.hunyuan.cloud.tencent.com/v1/chat/completions \
  -H "Authorization: Bearer $LLM_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "hunyuan-turbos-latest",
    "messages": [
      {"role": "system", "content": "你是一个简历分析助手。只输出 JSON，不要输出其他内容。"},
      {"role": "user", "content": "分析以下简历与岗位的匹配度，输出 JSON：{\"match_score\": 85, \"summary\": \"...\"}"}
    ],
    "temperature": 0.2,
    "max_tokens": 2048
  }'
```

### 5.2 预期响应结构

```json
{
  "id": "chatcmpl-xxx",
  "object": "chat.completion",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "{\"match_score\": 85, \"summary\": \"...\"}"
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 100,
    "completion_tokens": 50,
    "total_tokens": 150
  }
}
```

## 6. 安全注意事项

1. **不要在 `.env.example` 中写入真实 API Key**
2. **不要将 `.env` 提交到代码仓库**
3. **API Key 应妥善保管，不要分享给他人**
4. **注意 API 调用费用，建议关注免费额度使用情况**
5. **生产环境应使用更严格的密钥管理方案**

## 阶段 4-2 更新

4-2.0 已完成接入方案，详细方案见 `docs/05-Agent设计/阶段4-2LLMClient与Hunyuan接入方案.md`。

Hunyuan 接入后的核心约束：

1. **Hunyuan 接入后必须经过 LLM Client 抽象层**：业务 Agent 不直接依赖 Hunyuan
2. **必须支持 timeout、error_code、raw_output_length**
3. **必须支持 JSON parse、Schema validate**
4. **必须支持 JSON repair once**
5. **必须支持 Mock fallback**
6. **不打印 API Key**
7. **.env.example 中 API Key 必须为空**

4-2.5 才开始实现 Hunyuan Client。

## 阶段 4-2.5 更新

4-2.5 已完成 HunyuanClient HTTP 调用封装。当前仅通过 httptest 验证，不调用真实 Hunyuan。真实连通性验证放到 4-2.6。

新增：
- `server/internal/llm/hunyuan_client.go`：HunyuanClient 实现 llm.Client 接口
  - 支持 OpenAI-compatible chat completions 请求与响应映射
  - 支持 BaseURL endpoint 规范化（/v1、/v1/、/v1/chat/completions）
  - 支持 Authorization Bearer Token
  - 支持 per-request timeout 和 context cancellation
  - 支持 HTTP status 错误分类（429/5xx/401/403）
  - 支持空响应错误分类
  - 支持 finish_reason 记录和 usage 解析
  - 不在错误中泄露 API Key
- `server/internal/llm/hunyuan_types.go`：OpenAI-compatible 内部请求/响应类型（不导出）
- `server/internal/llm/hunyuan_client_test.go`：httptest 单元测试（22 项全部通过）
- 无新增第三方依赖，使用 Go 标准库 net/http / encoding/json / context

## 阶段 4-2.7 更新

4-2.7 已完成 JSON repair once 调用链路。后续 Hunyuan 输出发生 JSON parse / Schema validate 失败时，可进入 JSONRepairAgent 修复一次。repair 调用仍通过 llm.Client 抽象层完成。

新增：
- `server/internal/agent/repair/` 包：JSONRepairAgent repair once 调用链路
- repair 只执行一次，不递归修复
- repair 失败后返回结构化错误，Mock fallback 留给后续 Pipeline 决策

## 阶段 4-2.6 更新

4-2.6 已接入 Hunyuan smoke test。真实调用必须显式设置 `RUN_REAL_LLM_SMOKE=1`。smoke test 只使用最小 Prompt，不携带简历、JD 或业务数据。

新增：
- `server/cmd/llm-smoke/main.go`：最小 LLM smoke test Go 命令
- `scripts/dev/llm_smoke.sh`：LLM smoke test 脚本（help / check / smoke）
- Makefile `llm-help` / `llm-check` / `llm-smoke` target

执行方式：
```bash
make llm-check                              # 检查 LLM 配置（不请求外部模型）
RUN_REAL_LLM_SMOKE=1 make llm-smoke         # 执行真实 HunyuanClient 连通性验证
```

## 阶段 4-2.4 更新

4-2.4 已完成 LLM Client 抽象。Hunyuan Client 后续必须实现 `llm.Client` 接口（`Chat(ctx, ChatRequest) (*ChatResponse, error)` + `Provider() Provider`）。

Hunyuan 私有请求 / 响应结构不应泄漏到业务 Agent 层。业务 Agent 只依赖 `llm.Client`、`llm.ChatRequest`、`llm.ChatResponse` 等统一抽象类型。

新增：
- `server/internal/llm/` 包：Client 接口 + 基础类型 + Config + ErrorCode + NoopClient
- 错误码：LLM_DISABLED / LLM_CONFIG_MISSING / LLM_API_KEY_MISSING / LLM_TIMEOUT / LLM_HTTP_ERROR / LLM_RATE_LIMIT / LLM_EMPTY_RESPONSE / LLM_OUTPUT_TRUNCATED / UNKNOWN_ERROR
- `IsDirectMockFallback()`：判断哪些 LLM 错误应直接 Mock fallback

## 阶段 4-2.3 更新

4-2.3 已完成 Hunyuan raw output 后处理工具。后续 Hunyuan 返回内容不得直接进入业务接口，必须先经过 JSON parse + Schema validate。

新增 Go 包：
- `server/internal/llmjson/`：从 raw output 提取 JSON object、JSON parse、错误码分类
- `server/internal/schema/validator.go`：JSON Schema 校验（使用 santhosh-tekuri/jsonschema/v5，支持 draft 2020-12）

## 阶段 4-2.2 更新

4-2.2 已完成 Prompt / Schema 加载能力。Hunyuan 接入时不应把 Prompt 写死在代码中，而应通过 Prompt Loader 加载。

新增 Go 包：
- `server/internal/prompt/`：Prompt Loader（从 server/prompts/ 加载 .md 文件）
- `server/internal/schema/`：Schema Loader（从 server/prompts/schema/ 加载 .json 文件）

## 阶段 4-2.1 更新

4-2.1 已完成 Hunyuan 相关环境变量配置项，详细配置见 `docs/06-工程部署/环境变量.md`。

关键更新：

1. `.env.example` 已新增 LLM / Hunyuan Config 分组（14 项配置）
2. 清除了 `.env.example` 中遗留的真实 API Key（安全修复）
3. 清除了 `.env_BASE_URL 中的完整 tokenhub 路径
4. LLM_MAX_OUTPUT_TOKENS 从 2048 提升到 8192
5. LLM_TIMEOUT_SECONDS 从 20 提升到 30
6. 当前只固化配置，不调用 Hunyuan
7. Hunyuan API Key 必须只写入本地 .env

## 阶段 4-1.4 更新

4-1.4 只允许手动 curl 或离线方式测试 Prompt：

- 不在 Go 后端中接入 Hunyuan
- 不记录 API Key 到文档
- .env.example 中 LLM_API_KEY 应保持为空
- curl 测试时使用本地 .env 中的 LLM_API_KEY
- 方案文档见 `docs/05-Agent设计/阶段4-1.4Prompt与Agent真实接入方案.md`

### 4-1.4.2 手动调用验收记录

- 已执行一次匹配分析 Prompt 手动调用验收
- 使用 TokenHub（腾讯混元代理）+ hy3-preview 模型
- 首次调用 max_tokens=4096 导致截断，二次调用 max_tokens=8192 成功
- prompt_tokens=6604, completion_tokens=3618, total_tokens=10222
- LLM_MAX_OUTPUT_TOKENS 生产环境建议设置为 ≥8192
- 不记录 API Key
- 不记录完整模型原始输出
- 验收报告见 `docs/05-Agent设计/匹配分析Prompt首轮输出验收报告.md`

### 4-1.4.5 推荐 Prompt 首轮输出验收记录

- 已执行一次推荐 Prompt 手动调用验收
- 使用 TokenHub（腾讯混元代理）+ hy3-preview 模型
- max_tokens=8192，调用成功（finish_reason=stop）
- prompt_tokens=8692, completion_tokens=3832, total_tokens=12524
- 模型输出 JSON 合法，字段完整，证据链真实
- 不记录 API Key
- 不记录完整模型原始输出
- 验收报告见 `docs/05-Agent设计/推荐Prompt首轮输出验收报告.md`
- 后续推荐 Prompt 的模型调用也建议 max_tokens >= 8192

### 4-1.4.8 简历优化 Prompt 首轮输出验收记录

- 已执行一次简历优化 Prompt 手动调用验收
- 使用 TokenHub（腾讯混元代理）+ hy3-preview 模型
- max_tokens=8192，调用成功（finish_reason=stop）
- prompt_tokens=6723, completion_tokens=4032, total_tokens=10755
- 模型输出 JSON 合法，15 个顶层字段完整
- 不记录 API Key
- 不记录完整模型原始输出
- 验收报告见 `docs/05-Agent设计/简历优化Prompt首轮输出验收报告.md`
- 后续简历优化 Prompt 的模型调用也建议 max_tokens >= 8192

### 4-1.4.13 阶段汇总记录

- 阶段 4-1.4 已正式闭环，汇总文档见 `docs/00-项目总览/阶段4-1.4Prompt与Agent真实接入汇总.md`
- 4 个 Prompt 文件全部存在且已登记
- 3 个 Schema 文件全部存在且 JSON 合法
- 3 个首轮输出验收均为 partial_pass，问题已被 Schema 吸收
- 失败兜底与 Mock 回退策略已通过总验收
- 下一阶段进入 4-2 Hunyuan Client 接入准备与实现
- Hunyuan 接入前的 Prompt / Schema / JSON repair / fallback 策略均已就绪

### 4-1.4.12 阶段总验收记录

- 阶段 4-1.4 总验收已完成，验收报告见 `docs/05-Agent设计/阶段4-1.4总验收报告.md`
- 4 个 Prompt 文件全部存在且已登记
- 3 个 Schema 文件全部存在且 JSON 合法
- 3 个首轮输出验收均为 partial_pass，问题已被 Schema 吸收
- 失败兜底与 Mock 回退策略已通过验收
- 后续 Hunyuan 接入需要加载 Prompt / Schema，并接入 JSON repair once + Mock fallback 链路
- 当前 4-1.4.12 只做总验收，不调用 Hunyuan

### 4-1.4.11 失败兜底与 Mock 回退策略记录

- 失败兜底策略已设计，文档见 `docs/05-Agent设计/失败兜底与Mock回退策略.md`
- Hunyuan 接入后必须接入失败兜底策略
- Hunyuan 超时、HTTP 错误、限流、空响应直接 Mock fallback
- Hunyuan 输出 JSON 不合法或 Schema 不合法时先 JSON repair 一次
- JSON repair 只执行一次，repair 仍失败后 Mock fallback
- 当前 4-1.4.11 只做策略设计，不调用 Hunyuan
- 策略覆盖 20 种错误码，6 种直接 Mock 场景，6 种先 repair 场景

### 4-1.4.10B JSON 修复 Agent 指令模板登记记录

- JSON 修复 Prompt 已登记：`server/prompts/json_repair_v1.md`
- 模板面向"LLM JSON 输出结构修复"场景，支持 match_analysis / recommendation / resume_optimization 三类业务输出修复
- 模板只修复结构，不重新做业务分析，不编造业务事实
- 当前 4-1.4.10B 只登记 Prompt，不调用 Hunyuan
- 后续 JSON repair 调用建议只执行一次
- 如果 JSON repair 仍失败，应进入 Mock Agent fallback

### 4-1.4.9B 简历优化 JSON Schema 登记记录

- 简历优化 JSON Schema 已登记：`server/prompts/schema/resume_optimization_schema.json`
- Schema 验证结果：JSON 合法，15 个顶层 required 字段完整，10 个 $defs 子结构完整
- Schema 已吸收 4-1.4.8 首轮验收问题（允许 rewrite_examples 全部 can_use_directly=false、涉及数字/指标等时必须 requires_user_verification=true、不得编造具体数字/指标/系统规模/职责或成果、p2 允许 0-4 项）
- 当前 4-1.4.9B 只登记 Schema，不再次调用 Hunyuan
- 后续简历优化 Prompt 的模型调用建议 LLM_MAX_OUTPUT_TOKENS ≥8192

### 4-1.4.7B 简历优化 Agent 指令模板登记记录

- 简历优化 Prompt 已登记：`server/prompts/resume_optimization_v1.md`
- 模板面向"目标岗位导向的简历优化"场景，支持简历 + 目标岗位 JD + 可选匹配分析结果输入
- 模板输出 P0/P1/P2 优化建议、关键词对齐、项目优化、技能栏优化、摘要优化、缺失证据补足计划、改写示例和真实性提醒
- 当前 4-1.4.7B 只登记 Prompt，不调用 Hunyuan
- 后续简历优化 Prompt 首轮输出验收建议继续使用 max_tokens >= 8192

### 4-1.4.6B 推荐 JSON Schema 登记记录

- 推荐 JSON Schema 已登记：`server/prompts/schema/recommendation_schema.json`
- Schema 验证结果：JSON 合法，7 个顶层 required 字段完整，7 个 $defs 子结构完整
- Schema 已吸收 4-1.4.5 首轮验收问题（四档推荐固定但不强制覆盖、recommend_level 与 recommend_label 一致、match_score 区间一致）
- 当前 4-1.4.6B 只登记 Schema，不再次调用 Hunyuan
- 后续推荐 Prompt 的模型调用建议 LLM_MAX_OUTPUT_TOKENS ≥8192

### 4-1.4.3B Schema 登记记录

- 匹配分析 JSON Schema 已登记：`server/prompts/schema/match_analysis_schema.json`
- Schema 验证结果：JSON 合法，17 个顶层 required 字段完整，9 个 $defs 子结构完整
- Schema 已吸收 4-1.4.2 首轮验收问题（rewrite_example 禁止编造数字、evidence_chain 支持 missing）
- 当前 4-1.4.3B 只登记 Schema，不再次调用 Hunyuan
- 后续 Hunyuan 调用建议 LLM_MAX_OUTPUT_TOKENS ≥8192

## 关联文档

- [阶段4Agent与模型接入方案.md](./阶段4Agent与模型接入方案.md)
- [LLMClient设计.md](./LLMClient设计.md)
- [JSON输出规范.md](./JSON输出规范.md)
- [../06-工程部署/环境变量.md](../06-工程部署/环境变量.md)

## 阶段 4-3.0 更新

4-3.0 已完成业务 Agent 接入 LLM 方案设计（docs only）。Hunyuan 接入在 4-3 阶段的关键约束：

1. **4-3.1 需先执行真实 LLM Smoke Test**：验证 HunyuanClient 连通性后再接入业务 Agent
2. **业务 Agent 不直接依赖 HunyuanClient**：通过 Pipeline Runner 间接调用
3. **Hunyuan 配置已固化**：14 项 LLM_* 配置已就绪，4-3 不修改 .env.example
4. **4-3.1 将执行 RUN_REAL_LLM_SMOKE=1 make llm-smoke**：确认模型可用后再接入 MatchAnalysisAgent
5. **方案文档见 `docs/05-Agent设计/阶段4-3业务Agent接入LLM方案.md`**

## 阶段 4-3.1 更新

4-3.1 已执行真实 LLM Smoke 前置验收，但真实 smoke 未执行。

原因：本地 .env 不存在，LLM 配置不完整：
- LLM_ENABLE=false
- LLM_BASE_URL 未配置
- LLM_API_KEY 未配置
- LLM_MODEL 未配置

结论：HunyuanClient 真实连通性未验证，暂不建议进入 4-3.2。

验收报告见 `docs/05-Agent设计/真实LLMSmoke前置验收报告.md`。

用户需补齐 .env 配置后重新执行 `RUN_REAL_LLM_SMOKE=1 make llm-smoke`。
