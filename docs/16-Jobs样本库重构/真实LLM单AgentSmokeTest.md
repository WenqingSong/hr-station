# K5-1-R2：真实 LLM 直连验证与 Match Agent 重测

## 1. 性质

真实 LLM 最小直连 + Match Agent 重测。**未修改代码，未输出真实 Key。**

## 2. K5-1-S 回顾

K5-1-S 诊断根因：`LLM_TIMEOUT` — 30s 超时。BaseURL 可达，排除网络问题。最可能原因：前次测试使用了旧 Key（服务器未重启）。

## 3. 执行前状态

| 项 | 值 |
|----|-----|
| `LLM_ENABLE` | `true` |
| `LLM_PROVIDER` | `hunyuan` |
| `LLM_BASE_URL` | `https://tokenhub.tencentmaas.com/v1/chat/completions` |
| `LLM_MODEL` | `hy3-preview` |
| `LLM_API_KEY` | 已配置（masked，未输出） |
| `.env` git tracked | 否 |

## 4. 最小 LLM 直连请求

```json
{
  "model": "hy3-preview",
  "messages": [{"role": "user", "content": "Return only this JSON: {\"ok\":true}"}],
  "temperature": 0
}
```

```
HTTP 200 ✅
Time: 2.29s
Content: {"ok": true}
finish_reason: stop
usage: prompt_tokens=25, completion_tokens=6, total_tokens=31
```

**结论：LLM 直连成功。** TokenHub hy3-preview 可用。

## 5. Match Agent 重测

```
POST /api/jobs/55/match
{"resume_version_no": 3}
```

### 结果

```
HTTP: 200
code: 0
generated_by: llm_agent        ← 真实 LLM！
agent_status: success          ← 首次通过，无需 repair
fallback_reason: none
error_code: none
repair_applied: False
schema_valid: True
```

### 核心字段 vs Mock

| 字段 | 真实 LLM | Mock (K5-1-R) |
|------|---------|---------------|
| `match_score` | **55** | 23 |
| `evidence_chain` | **5** | 11 |
| `top_fix_items` | **3** | 2 |
| `match_highlights` | **2** | 1 |
| `apply_readiness` | **quick_fix_before_apply** | not_ready_for_this_role |
| `optimization_stop_signal` | **fix_top_3_then_apply** | needs_more_evidence |

### 真实 LLM 内容样例

**Match highlights:**
- "掌握Go语言，符合岗位要求的开发语言方向"
- "掌握Redis，匹配岗位了解NoSQL、Key-Value存储原理的要求"

**Risk points:**
- "仅列出Go、Redis技能，无项目支撑，技术面试时无法回答相关追问"
- "简历信息过少，HR无法在30秒内判断匹配度"

**Top fix item [0]:**
- "简历中缺少项目经历、教育经历、工作/实习经历的具体信息，仅列出技能无法证明能力"
- "补充至少1个与Go、Redis相关的真实项目经历，包含项目背景、你的职责、使用的技术栈"

## 6. 判定

| 检查项 | 结果 |
|--------|------|
| HTTP 200 | ✅ |
| generated_by ≠ mock_agent | ✅ llm_agent |
| agent_status ≠ fallback | ✅ success |
| fallback_reason 为空 | ✅ none |
| parse 成功 | ✅ |
| schema_valid=true | ✅ |
| 包含 core fields | ✅ match_score/evidence_chain/top_fix_items |
| 无 repair | ✅ |

**真实 LLM Match Agent：完全成功。**

## 7. Key 安全检查

| 检查 | 结果 |
|------|------|
| real Key in tracked files | clean ✅ |
| `.env` in git diff | 否 ✅ |
| git diff | docs only ✅ |

## 8. P0 / P1 / P2

| 级别 | 编号 | 问题 | 状态 |
|------|------|------|------|
| — | — | **P0 = 0** | ✅ |
| — | — | **P1 = 0** | ✅ |
| P2 | P2-1 | 只测了 Match Agent | 后置 |
| P2 | P2-2 | 浏览器人工验收未执行 | 后置 |

## 9. 判定

**K5-1-R2 = pass。** 真实 LLM Match Agent 首次通过，`generated_by=llm_agent`，`agent_status=success`，`schema_valid=True`。

## 10. 是否可以进入 K5-2

**可以。** 下一步扩展测试到 Recommendation Agent 和 Resume Optimization Agent。
