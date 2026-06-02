# K5-1：真实 LLM 单 Agent Smoke Test

## 1. 状态

**blocked** — 等待用户在本地 `.env` 配置真实 LLM API Key。

## 2. 阻塞原因

`.env` 当前配置：

| 配置项 | 当前值 | 说明 |
|--------|--------|------|
| `LLM_ENABLE` | `false` | 未启用真实 LLM |
| `LLM_API_KEY` | `sk-your-real-api-key` | **占位符，非真实 Key** |
| `LLM_PROVIDER` | `hunyuan` | ✅ |
| `LLM_BASE_URL` | `https://tokenhub.tencentmaas.com/v1/chat/completions` | ✅ |
| `LLM_MODEL` | `hy3-preview` | ✅ |
| `LLM_TEMPERATURE` | `0.2` | ✅ |
| `LLM_TIMEOUT_SECONDS` | `30` | ✅ |

## 3. 需要用户操作

在本地 `.env` 中修改：

```bash
# 在 /workspace/.env 中修改
LLM_ENABLE=true
LLM_API_KEY=<你的真实 API Key>
```

然后重启服务：

```bash
make restart
```

## 4. 测试数据已就绪

| 数据 | 验证结果 |
|------|---------|
| 岗位 55 | ✅ `软件开发-后台开发方向` |
| 简历版本 3 | ✅ v3 可用 |
| API 基础可用 | ✅ `/api/jobs/55` HTTP 200 |

## 5. Match Analysis Agent 调用路径

```
POST /api/jobs/{id}/match
Content-Type: application/json
Authorization: Bearer <token>

{
  "resume_version_no": 3
}
```

路径：`POST /api/jobs/55/match`

调用链路（LLM_ENABLE=true 时）：

```
MatchHandler → MatchService
  → MatchAnalysisAgent.Analyze(ctx, resume, job)
    → Pipeline Runner.Run()
      → Prompt 渲染 (match_analysis_v1.md)
      → HunyuanClient.Chat() → POST LLM BaseURL
      → JSON parse → Schema validate → repair once → fallback if needed
  → return MatchAnalysisResult + AgentResponseMeta
```

## 6. 代码路径验证

| 检查项 | 结果 |
|--------|------|
| Match API route | ✅ `POST /api/jobs/{id}/match` |
| MatchAnalysisAgent 已接入 Pipeline | ✅ |
| MatchMockAgent 已实现 fallback | ✅ |
| Service 层 llmEnabled 开关 | ✅ |
| Prompt 存在 | ✅ `match_analysis_v1.md` |
| Schema 存在 | ✅ `match_analysis_schema.json` |
| 需要 auth token | ✅ yes |

## 7. 预期行为

### 成功路径（LLM_ENABLE=true + 有效 Key）

- HTTP 200
- `data.result` 包含真实 LLM 输出（match_score, evidence_chain, top_fix_items 等）
- `data.meta.generated_by = "llm_agent"`
- `data.meta.agent_status = "success"` 或 `"repaired"`

### 失败路径（LLM 调用失败）

- `MatchFallbackError` → `MatchMockAgent` fallback
- HTTP 200
- `data.meta.generated_by = "mock_agent"`
- `data.meta.agent_status = "fallback"`
- `data.meta.fallback_reason` 说明原因

## 8. 安全状态

| 检查项 | 结果 |
|--------|------|
| `.env` 未被 git 追踪 | ✅ |
| `.env.example` 仅占位符 | ✅ |
| Key 扫描（tracked files）| ✅ clean |
| `json:"-"` 标记 | ✅ |
| `MaskedAPIKey()` + `safeErr()` | ✅ |

**本阶段未修改代码，未输出真实 Key。**

## 9. P0 / P1 / P2

| 级别 | 编号 | 问题 | 状态 |
|------|------|------|------|
| — | — | P0 = 0 | ✅ |
| — | — | P1 = 0 | ✅ |
| P2 | P2-1 | **阻因：用户未配置真实 API Key** | 用户本地配置 |
| P2 | P2-2 | 真实 LLM 调用未执行 | 等待 Key |
| P2 | P2-3 | 只规划了 Match Agent，推荐/优化未测 | 后置 |

## 10. 判定

**K5-1 blocked。** 未修改代码。等待用户在本地 `.env` 配置真实 `LLM_API_KEY` 并设 `LLM_ENABLE=true` 后方可继续。

## 11. 下一步

用户配置真实 Key 后：

1. `make restart` 重启服务
2. 获取 auth token：`POST /api/auth/login`
3. 调用 `POST /api/jobs/55/match` 验证真实 LLM 输出
4. 验证 parse / schema validate / repair / fallback
5. 将结果填入本文档，从 blocked 更新为 pass / partial_pass
