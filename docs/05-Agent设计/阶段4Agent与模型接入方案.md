# 阶段4Agent与模型接入方案

> 最后更新：2026-05-29（阶段 4-1.2.0 腾讯招聘公开接口探测测试）

## 1. 阶段目标

在保留阶段 3 Mock Agent 兜底能力的基础上，引入 Hunyuan API 作为 LLM Provider，让推荐、匹配分析、版本对比逐步从 Mock 输出升级为真实模型生成的结构化 JSON 输出。

## 2. 当前基础

阶段 3 已完成以下交付物：

| 交付物 | 状态 | 说明 |
|--------|------|------|
| 3 个 Mock Agent | ✅ 已完成 | RecommendationMockAgent / MatchMockAgent / VersionCompareMockAgent |
| 14 个 HTTP 接口 | ✅ 已完成 | 认证 + 简历 + 岗位 + 推荐 + 匹配 + 版本对比 |
| 6 张数据库表 | ✅ 已完成 | 含种子数据 |
| 统一响应结构 | ✅ 已完成 | `{code, message, data}` |
| JWT 认证 | ✅ 已完成 | Bearer Token |
| Mock 结果快照保存 | ✅ 已完成 | recommendation_results / match_analyses / version_compare_results |

当前 Mock Agent 的局限：
- 基于简单规则打分，无法理解自然语言
- 无法生成有深度的分析和建议
- 无法根据真实简历内容做个性化推理
- 推荐理由和证据链较为模板化

## 3. 技术路线

```
业务接口
  → Agent Service
    → Prompt Builder
      → LLM Client
        → Hunyuan OpenAI-compatible API
      → JSON 输出解析
    → 字段校验
  → 失败兜底 Mock
  → 统一响应
```

核心流程：

1. Agent Service 接收业务请求
2. Prompt Builder 根据业务类型构建 Prompt（含 JSON 输出格式约束）
3. LLM Client 调用 Hunyuan API
4. 解析 JSON 输出
5. 字段校验（分数范围、必填字段、真实性约束）
6. 校验通过 → 返回结果
7. 校验失败 / LLM 调用失败 → 回退 Mock Agent

## 4. Hunyuan 接入策略

### 4.1 为什么选择 Hunyuan

- 腾讯云官方大模型服务，国内访问稳定
- 提供 OpenAI-compatible API，降低接入复杂度
- 有免费额度可用于 Demo 开发
- 与本项目"腾讯云生态"定位一致

### 4.2 接入方式

Hunyuan 提供 OpenAI-compatible API，可直接使用 HTTP POST 调用：

```
POST https://api.hunyuan.cloud.tencent.com/v1/chat/completions
Authorization: Bearer <API_KEY>
Content-Type: application/json

{
  "model": "hunyuan-turbos-latest",
  "messages": [...],
  "temperature": 0.2,
  "max_tokens": 2048
}
```

### 4.3 关键配置

| 配置项 | 环境变量 | 说明 |
|--------|---------|------|
| Provider | `LLM_PROVIDER` | 固定 `hunyuan` |
| Base URL | `LLM_BASE_URL` | `https://api.hunyuan.cloud.tencent.com/v1` |
| API Key | `LLM_API_KEY` | 只写 `.env`，不提交 |
| Model | `LLM_MODEL` | `hunyuan-turbos-latest` |
| Timeout | `LLM_TIMEOUT_SECONDS` | 20 秒 |
| Max Retries | `LLM_MAX_RETRIES` | 1 次 |
| Max Output Tokens | `LLM_MAX_OUTPUT_TOKENS` | 2048 |
| Temperature | `LLM_TEMPERATURE` | 0.2（低随机性，确保输出稳定） |
| Enable LLM | `ENABLE_LLM_AGENT` | `true` 启用 / `false` 使用 Mock |
| Enable Fallback | `ENABLE_LLM_FALLBACK` | `true` 失败时回退 Mock |

## 5. Agent 改造顺序

按风险从低到高逐步接入：

| 顺序 | Agent | 原因 |
|------|-------|------|
| 1 | 岗位匹配分析 | 核心亮点功能，输入输出最明确，用户最关注 |
| 2 | 岗位推荐 | 与匹配分析逻辑相似，Prompt 模板可复用 |
| 3 | 简历版本对比 | 输入最复杂（两个版本），需更精细的 Prompt |

每个 Agent 的改造步骤：

1. 编写该 Agent 的 Prompt 模板（要求只输出 JSON）
2. 在 Service 中判断 `ENABLE_LLM_AGENT`
3. 若启用，构建 Prompt → 调用 LLM Client → 解析 JSON → 校验
4. 若失败且 `ENABLE_LLM_FALLBACK=true`，回退 Mock Agent
5. 若未启用，直接使用 Mock Agent

## 6. Prompt 模板管理

### 6.1 模板目录

```
server/internal/agent/prompts/
├── job_match.md           # 岗位匹配 Prompt（4-4）
├── job_recommend.md       # 岗位推荐 Prompt（4-5）
└── version_compare.md     # 版本对比 Prompt（4-6）
```

### 6.2 模板格式

使用 Go template 语法，变量用 `{{.Variable}}` 标记。

### 6.3 Prompt 必须包含的约束

1. **只输出 JSON**：不输出 Markdown、不输出解释性前后缀、不输出代码块
2. **真实性约束**：不虚构经历、不编造不存在的项目或能力
3. **禁止自动生成简历**：不输出优化版简历内容
4. **分数范围**：0-100 整数
5. **必填字段**：每个 Agent 的 JSON 结构有明确必填字段
6. **证据链**：匹配分析必须有 JD-简历证据链

## 7. JSON 输出校验

### 7.1 校验流程

```
LLM 原始输出
  → 提取 JSON（去除可能的 Markdown 代码块包裹）
  → JSON 解析
  → 字段完整性校验
  → 分数范围校验（0-100）
  → 真实性校验（不包含虚构内容）
  → 通过 / 失败
```

### 7.2 校验失败处理

- JSON 解析失败 → 回退 Mock
- 缺必填字段 → 回退 Mock
- 分数超范围 → 截断到 0-100
- 包含"自动生成简历" → 回退 Mock
- 包含虚构经历 → 回退 Mock

详细规范见 [JSON输出规范.md](./JSON输出规范.md)

## 8. 失败兜底策略

### 8.1 兜底条件

以下情况触发兜底：

| 场景 | 说明 | 兜底方式 |
|------|------|---------|
| LLM 调用超时 | 请求超过 `LLM_TIMEOUT_SECONDS` | 回退 Mock |
| LLM 返回错误 | HTTP 4xx/5xx | 回退 Mock |
| JSON 解析失败 | 输出不是合法 JSON | 回退 Mock |
| 字段校验失败 | 缺必填字段 | 回退 Mock |
| 真实性校验失败 | 输出虚构内容 | 回退 Mock |
| LLM 未启用 | `ENABLE_LLM_AGENT=false` | 使用 Mock |
| API Key 为空 | `LLM_API_KEY=""` | 使用 Mock |

### 8.2 兜底行为

- 兜底时使用阶段 3 Mock Agent 的原有逻辑
- 响应中通过 `generated_by` 字段区分来源：
  - `"mock_agent"`：Mock Agent 生成
  - `"llm_agent"`：真实 LLM 生成
  - `"llm_agent_fallback"`：LLM 失败后回退 Mock
- 兜底对用户透明，不抛出额外错误

### 8.3 开关控制

```
ENABLE_LLM_AGENT=false → 始终使用 Mock Agent
ENABLE_LLM_AGENT=true, ENABLE_LLM_FALLBACK=true → 优先 LLM，失败回退 Mock
ENABLE_LLM_AGENT=true, ENABLE_LLM_FALLBACK=false → 优先 LLM，失败直接报错
```

## 9. 环境变量设计

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `ENABLE_LLM_AGENT` | `false` | 是否启用真实 LLM Agent |
| `ENABLE_LLM_FALLBACK` | `true` | LLM 失败是否回退 Mock |
| `LLM_PROVIDER` | `hunyuan` | LLM 提供商 |
| `LLM_BASE_URL` | `https://api.hunyuan.cloud.tencent.com/v1` | API Base URL |
| `LLM_API_KEY` | 空 | API Key，只写 `.env` |
| `LLM_MODEL` | `hunyuan-turbos-latest` | 模型名称 |
| `LLM_TIMEOUT_SECONDS` | `20` | 超时秒数 |
| `LLM_MAX_RETRIES` | `1` | 最大重试次数 |
| `LLM_MAX_OUTPUT_TOKENS` | `2048` | 最大输出 token |
| `LLM_TEMPERATURE` | `0.2` | 温度 |

与已有变量 `ENABLE_MOCK_AGENT` 的关系：
- `ENABLE_MOCK_AGENT=true`：阶段 3 控制 Mock Agent 启用
- `ENABLE_LLM_AGENT=false`：阶段 4 控制是否启用真实 LLM
- 默认情况下两个都是 false/true，即使用 Mock Agent
- 当 `ENABLE_LLM_AGENT=true` 时，优先使用真实 LLM
- 当 `ENABLE_LLM_AGENT=true` 且 `ENABLE_LLM_FALLBACK=true` 时，LLM 失败回退 Mock

## 10. 阶段 4 任务拆分

| 编号 | 任务 | 说明 | 是否修改 Go 代码 |
|------|------|------|-----------------|
| 4-1 | 创建 Hunyuan API 接入准备方案与配置清单 | 配置 `.env.example`、编写接入准备文档 | ❌ 否 |
| 4-2 | 实现 LLM Client 抽象与 Hunyuan Client | `server/internal/llm/` 目录 | ✅ 是 |
| 4-3 | 实现 JSON 输出解析、校验与失败兜底 | `server/internal/llm/json_validator.go` | ✅ 是 |
| 4-4 | 接入岗位匹配分析真实 LLM Agent | 改造 `match_service` 和 `match_mock` | ✅ 是 |
| 4-5 | 接入岗位推荐真实 LLM Agent | 改造 `recommendation_service` 和 `recommendation_mock` | ✅ 是 |
| 4-6 | 接入简历版本对比真实 LLM Agent | 改造 `version_compare_service` 和 `version_compare_mock` | ✅ 是 |
| 4-7 | 阶段 4 汇总 | 确认所有 Agent 可用，Mock 兜底正常 | ❌ 否 |

## 11. 验收标准

| 序号 | 验收标准 |
|------|---------|
| 1 | `ENABLE_LLM_AGENT=false` 时，所有接口行为与阶段 3 一致 |
| 2 | `ENABLE_LLM_AGENT=true` 且 API Key 有效时，3 个 Agent 能调用真实 LLM |
| 3 | LLM 输出 JSON 解析成功后，响应结构与阶段 3 兼容 |
| 4 | LLM 调用失败时，`ENABLE_LLM_FALLBACK=true` 自动回退 Mock |
| 5 | `generated_by` 字段正确区分 `mock_agent` / `llm_agent` / `llm_agent_fallback` |
| 6 | 所有 Prompt 模板包含"只输出 JSON"和"不虚构经历"约束 |
| 7 | JSON 校验覆盖必填字段、分数范围、真实性 |
| 8 | 不修改已有 API 路径和响应主结构 |
| 9 | 不自动生成优化版简历 |
| 10 | 不修改用户简历内容 |
| 11 | 未填写 API Key 时，系统正常使用 Mock Agent 运行 |
| 12 | 文档、变更记录、开发日志完整 |

## 12. 不做事项

| 不做 | 原因 |
|------|------|
| 不接复杂平台型 Agent | 项目定位为自研 Agent 层 |
| 不重写后端接口 | 保持阶段 3 接口路径和响应主结构不变 |
| 不重写数据库表结构 | 数据层无需变更 |
| 不改前端页面 | 前端接入作为阶段 5 |
| 不自动生成优化版简历 | 核心约束，保护真实性 |
| 不修改用户简历内容 | 核心约束 |
| 不去掉 Mock 兜底 | Mock Agent 作为 LLM 失败的保障 |
| 不引入复杂 Agent 框架 | 用简单 Prompt + JSON 即可满足需求 |
| 不做异步 Agent 调用 | 当前同步返回即可 |
| 不接非 Hunyuan 的 LLM Provider | 后续按需扩展 |

## 关联文档

- [Hunyuan接入准备.md](./Hunyuan接入准备.md)
- [LLMClient设计.md](./LLMClient设计.md)
- [JSON输出规范.md](./JSON输出规范.md)
- [Agent总览.md](./Agent总览.md)
- [Prompt模板.md](./Prompt模板.md)
- [../02-后端设计/阶段3核心业务后端Mock方案.md](../02-后端设计/阶段3核心业务后端Mock方案.md)
