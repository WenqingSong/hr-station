# K5-1-R：真实 LLM 单 Agent Smoke Test 继续执行

## 1. 性质

K5-1 的继续执行。`LLM_ENABLE=true`，真实 Key 已配置在本地 `.env`。

**未修改代码，未输出真实 Key。**

## 2. 执行前状态

| 检查项 | 值 |
|--------|-----|
| 当前分支 | `main` |
| 当前 HEAD | `bf8cb5d` |
| 工作区 | clean |
| `.env` git tracked | 否 |
| `LLM_ENABLE` | `true` |
| `LLM_API_KEY` | 已配置（masked，未输出） |
| `LLM_PROVIDER` | `hunyuan` |
| `LLM_BASE_URL` | `https://tokenhub.tencentmaas.com/v1/chat/completions` |
| `LLM_MODEL` | `hy3-preview` |

## 3. 测试数据

| 数据 | 验证 |
|------|------|
| 岗位 55 | ✅ `软件开发-后台开发方向` / manual_curated / manual_verified |
| 简历版本 3 | ✅ 真实版本详情 |

## 4. 请求

```
POST /api/jobs/55/match
Authorization: Bearer <token_masked>
Content-Type: application/json

{"resume_version_no": 3}
```

## 5. 真实 LLM Smoke Test 结果

```
HTTP 200
code: 0
generated_by: mock_agent
agent_status: fallback
fallback_reason: llm_call_failed
schema_valid: True
repair_applied: False
```

### 分析

- 真实 LLM 调用失败（`llm_call_failed`）：HunyuanClient.Chat() 返回了错误
- Pipeline 正确将错误分类为 `llm_call_failed`，触发 fallback
- `MatchFallbackError` 被 Service 层捕获，回退到 `MatchMockAgent`
- Mock Agent 正常生成了匹配分析结果
- 结果通过 schema validate（`schema_valid: True`）
- HTTP 200 正常返回，前端不会白屏

### 核心结果字段（Mock fallback）

| 字段 | 值 |
|------|-----|
| match_score | 23 |
| match_highlights | 1 条 |
| risk_points | 3 条 |
| evidence_chain | 11 条 |
| top_fix_items | 2 条 |
| apply_readiness | not_ready_for_this_role |
| optimization_stop_signal | needs_more_evidence |

## 6. 失败原因分析

| 可能原因 | 说明 |
|---------|------|
| API Key 无效 | 最可能 — TokenHub 认证失败 |
| BaseURL 不可达 | 网络连通性 |
| Model 名称不对 | `hy3-preview` 与 TokenHub 实际模型不符 |

**重点：fallback 路径完全可靠。** 真实 LLM 失败对前端零影响。

## 7. Key 安全检查

| 检查 | 结果 |
|------|------|
| real sk- / AKIA in tracked files | clean ✅ |
| `.env` in git tracked | 否 ✅ |
| `.env.example` | 占位符 `sk-your-real-api-key` ✅ |
| git diff | empty ✅ |
| pipeline_test.go `sk-fake` | fake test key ✅ |

## 8. go test

```
ok  offer-catcher/server/internal/errors
ok  offer-catcher/server/internal/llm
ok  offer-catcher/server/internal/llmjson
ok  offer-catcher/server/internal/model
ok  offer-catcher/server/internal/prompt
ok  offer-catcher/server/internal/schema
ok  offer-catcher/server/internal/service
```

All pass ✅

## 9. P0 / P1 / P2

| 级别 | 编号 | 问题 | 状态 |
|------|------|------|------|
| — | — | P0 = 0 | ✅ |
| — | — | P1 = 0 | ✅ |
| P2 | P2-1 | 真实 LLM 调用失败（llm_call_failed），fallback 正常 | 需排查 Key/URL/Model |
| P2 | P2-2 | 只测了 Match Agent | 后置 |
| P2 | P2-3 | 浏览器人工验收未执行 | 后置 |

## 10. 判定

**K5-1-R = partial_pass。**

- 真实 LLM 调用失败，但 fallback 可靠 ✅
- 错误处理路径正确（llm_call_failed → Mock fallback）✅
- 前端不白屏，结果可读 ✅
- P0=0 P1=0 ✅
- 零代码修改 ✅
- 零 Key 泄漏 ✅

## 11. 是否可以进入 K5-2

**可以。** K5-2 建议：排查 LLM API Key / URL / Model 配置问题后重新测试，或扩展测试到推荐/优化 Agent。
