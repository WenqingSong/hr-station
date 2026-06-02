# K5-1：真实 LLM 单 Agent Smoke Test

- **日期**：2026-06-02
- **阶段**：K5-1（blocked）
- **阻断原因**：`.env` 中 `LLM_API_KEY` 为占位符 `sk-your-real-api-key`，非真实 Key

## 1. blocked 原因

当前 `.env` 中：
- `LLM_ENABLE=false` — 未启用
- `LLM_API_KEY=sk-your-real-api-key` — 占位符

无法执行真实 LLM smoke test。

## 2. 已确认就绪

| 项目 | 状态 |
|------|------|
| Match API route | ✅ `POST /api/jobs/{id}/match` |
| MatchAnalysisAgent + Pipeline | ✅ 已接入 |
| MatchMockAgent fallback | ✅ |
| Prompt + Schema | ✅ 完整 |
| 测试数据 jobId=55 | ✅ HTTP 200 |
| go test | ✅ all pass |
| Key 安全 | ✅ 无泄漏 |
| 代码修改 | ✅ 零修改 |

## 3. 需要用户操作

1. 在 `.env` 中设置 `LLM_ENABLE=true`
2. 在 `.env` 中设置 `LLM_API_KEY=<真实 Key>`
3. 执行 `make restart`
4. 获取 token 后调用 `POST /api/jobs/55/match`

## 4. P0/P1/P2

| 级别 | 数量 |
|------|------|
| P0 | 0 |
| P1 | 0 |
| P2 | 3（Key 未配 + 调用未执行 + 推荐/优化未测） |

## 5. 判定

**K5-1 blocked。** 等待用户本地配置真实 LLM API Key。
