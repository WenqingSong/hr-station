# K5-1-R2：真实 LLM 直连验证与 Match Agent 重测

- **日期**：2026-06-02
- **阶段**：K5-1-R2（pass）

## 1. 目标

先做最小 LLM 直连验证，再重测 Match Agent。

## 2. 最小直连请求

```
POST hy3-preview, "Return only this JSON: {\"ok\":true}"
HTTP 200, 2.29s, {"ok":true} ✅
```

## 3. Match Agent 重测

```
POST /api/jobs/55/match
generated_by=llm_agent, agent_status=success ✅
match_score=55, evidence_chain=5, top_fix_items=3
fallback=none, repair=none, schema_valid=True
```

## 4. P0/P1/P2

P0=0, P1=0, P2=2

## 5. 判定

**K5-1-R2 = pass。** 真实 LLM 首次通过。可进入 K5-2。
