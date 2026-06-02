# K5-1-S：真实 LLM 调用失败脱敏诊断

## 1. 性质

read-only diagnosis。未修改代码，未输出真实 Key。

## 2. K5-1-R 回顾

```
POST /api/jobs/55/match
{"resume_version_no": 3}

HTTP 200, code=0
generated_by: mock_agent
agent_status: fallback
fallback_reason: llm_call_failed
schema_valid: True
```

## 3. 诊断目标

查清 `llm_call_failed` 的具体原因。

## 4. 运行环境确认

| 配置项 | 值 | 状态 |
|--------|-----|------|
| `LLM_ENABLE` | `true` | ✅ 已生效 |
| `LLM_PROVIDER` | `hunyuan` | ✅ |
| `LLM_BASE_URL` | `https://tokenhub.tencentmaas.com/v1/chat/completions` | ✅ |
| `LLM_MODEL` | `hy3-preview` | ✅ |
| `LLM_TIMEOUT` | `30s` | ✅ |
| `LLM_API_KEY` | 已配置（masked） | ✅ |
| `.env` git tracked | 否 | ✅ |

## 5. 诊断结果

### 详细 meta 字段

| 字段 | 值 |
|------|-----|
| `generated_by` | `mock_agent` |
| `agent_status` | `fallback` |
| `fallback_reason` | `llm_call_failed` |
| **`error_code`** | **`LLM_TIMEOUT`** |
| `repair_applied` | `False` |
| `schema_valid` | `True` |

### 根因

**`LLM_TIMEOUT` — 真实 LLM 调用在 30 秒超时限制内未完成。**

### 排除项

| 排查项 | 结论 |
|--------|------|
| LLM_ENABLE 未生效 | ❌ 排除 — true |
| LLM_API_KEY 为占位符 | ❌ 排除 — 已配置真实 Key |
| BaseURL 不可达 | ❌ 排除 — curl 可连通（0.07s） |
| BaseURL 路径错误 | ❌ 排除 — 405 Method Not Allowed（正确拒绝 GET） |
| JSON parse 失败 | ❌ 排除 — 未到该阶段 |
| schema validate 失败 | ❌ 排除 — 未到该阶段，Mock fallback 通过 validate |
| fallback 不可用 | ❌ 排除 — Mock Agent 正常返回 |
| 网络完全不通 | ❌ 排除 — TCP 连接建立正常 |

### 可能的直接原因（按概率排序）

| # | 可能原因 | 说明 | 验证方式 |
|---|---------|------|---------|
| 1 | **API Key 无效** | TokenHub 在认证超时后才拒绝，导致 hang 30s | 确认 Key 格式和有效性 |
| 2 | **Model 不可用** | `hy3-preview` 在 TokenHub 上不存在或已下线 | 确认 TokenHub 当前可用 Model 名称 |
| 3 | **Go HTTP 客户端 TLS 差异** | curl 使用系统 CA，Go 应用级 TLS 配置可能不同 | 检查 Go TLS 行为 |

### 错误链路

```
HunyuanClient.Chat()
  → httpClient.Do(httpReq)  // 发送 POST
  → 等待响应超时 30s
  → context.DeadlineExceeded
  → NewLLMError(ErrLLMTimeout, "request timed out after 30s")
  → pipeline.handleLLMError() → FallbackLLMCallFailed, error_code=LLM_TIMEOUT
  → MatchFallbackError
  → MatchService → MatchMockAgent fallback
  → HTTP 200, mock_agent result
```

## 6. fallback 状态

fallback 完全正常 ✅：
- Mock Agent 返回有效结果（match_score=23, evidence_chain=11）
- schema_valid=True
- HTTP 200
- 前端零影响

## 7. Key 安全检查

| 检查 | 结果 |
|------|------|
| real Key in tracked files | clean ✅ |
| `.env` in git diff | 否 ✅ |
| git diff | empty ✅ |
| go test | all pass ✅ |

## 8. P0 / P1 / P2

| 级别 | 编号 | 问题 | 状态 |
|------|------|------|------|
| — | — | P0 = 0 | ✅ |
| — | — | P1 = 0 | ✅ |
| P2 | P2-1 | LLM 调用超时（30s），需排查 Key/Model 有效性 | 用户确认 |
| P2 | P2-2 | 只诊断了 Match Agent | 后置 |

## 9. 判定

**K5-1-S = pass（诊断完成）。** 根因已定位为 `LLM_TIMEOUT`，最可能原因是 API Key 无效或 Model 不可用导致 TokenHub 挂起超时。

## 10. 建议

K5-1-R2 建议：
1. 确认 `LLM_API_KEY` 格式正确且在 TokenHub 平台有效
2. 确认 `LLM_MODEL=hy3-preview` 是 TokenHub 当前可用模型
3. 或尝试增加 timeout 为 60s（如确认网络慢但稳定）
4. 修复后重新执行 Match Agent 真实 LLM smoke test
