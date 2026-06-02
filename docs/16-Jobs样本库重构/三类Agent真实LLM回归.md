# K5-2：三类 Agent 真实 LLM 回归

## 1. 性质

真实 LLM 回归。**未修改代码，未输出真实 Key。**

## 2. 配置

| 项 | 值 |
|----|-----|
| `LLM_ENABLE` | `true` |
| `LLM_PROVIDER` | `hunyuan` |
| `LLM_MODEL` | `hy3-preview` |
| `LLM_TIMEOUT` | `120s`（上调以适配大 Prompt） |
| `LLM_API_KEY` | 已配置（masked） |
| `.env` git tracked | 否 |

## 3. REAL_JOB_ID / REAL_RESUME_VERSION_NO

`55` / `3`

## 4. 三类 Agent 回归结果

### 4.1 Match Analysis Agent

```
POST /api/jobs/55/match  {"resume_version_no": 3}
```

| 字段 | 值 |
|------|-----|
| HTTP | 200 |
| `generated_by` | **`llm_agent`** ✅ |
| `agent_status` | **`success`** ✅ |
| `fallback_reason` | `none` |
| `repair_applied` | `False` |
| `schema_valid` | `True` |
| `match_score` | 55 |
| `evidence_chain` | 6 |
| `top_fix_items` | 3 |

### 4.2 Recommendation Agent

```
POST /api/jobs/recommend  {"resume_version_no": 3, "limit": 3}
```

| 字段 | 值 |
|------|-----|
| HTTP | 200 |
| `generated_by` | **`llm_agent`** ✅ |
| `agent_status` | **`repaired`** ✅ |
| `fallback_reason` | `none` |
| `repair_applied` | `True` |
| `schema_valid` | `True` |
| 推荐数量 | 3 |
| 推荐结果 `generated_by` | `llm_agent` |

**说明：** LLM 首次输出未通过 schema validate，JSONRepairAgent 修复一次后通过。

### 4.3 Resume Optimization Agent

```
POST /api/resumes/optimize
{"resume_text":"Go developer...", "target_job":{...}, "target_section":"skills"}
```

| 字段 | 值 |
|------|-----|
| HTTP | 200 |
| `generated_by` | `mock_agent` ⚠️ |
| `agent_status` | `fallback` ⚠️ |
| `fallback_reason` | `json_repair_schema_failed` |
| `error_code` | `JSON_REPAIR_SCHEMA_FAILED` |
| `repair_applied` | `False` |
| `schema_valid` | `True`（Mock fallback） |
| `apply_readiness` | `insufficient_information` |
| `top_fix_items` | 2 |

**说明：** LLM 返回了不合法的 JSON，JSONRepairAgent 修复失败，Pipeline 正确回退到 Mock Agent。

## 5. 回归汇总

| Agent | LLM | Status | Repair | Fallback | Schema |
|-------|-----|--------|--------|----------|--------|
| Match Analysis | ✅ llm_agent | success | 无 | 无 | True |
| Recommendation | ✅ llm_agent | repaired | ✅ 一次成功 | 无 | True |
| Resume Optimize | ❌ mock_agent | fallback | ❌ 失败 | ✅ Mock | True |

**2/3 Agent 真实 LLM 成功。** OPT LLM 输出格式问题导致 fallback，不影响前端。

## 6. parse / validate / repair / fallback

| 阶段 | Match | Rec | Opt |
|------|-------|-----|-----|
| LLM 调用 | ✅ | ✅ | ✅ |
| JSON parse | ✅ | ✅ 首次失败 | ❌ |
| Schema validate | ✅ | ❌ 首次失败 | ❌ |
| Repair | 未触发 | ✅ 成功 | ❌ 失败 |
| Fallback | 未触发 | 未触发 | ✅ Mock |

## 7. Key 安全检查

| 检查 | 结果 |
|------|------|
| real Key in tracked files | clean ✅ |
| `.env` in git diff | 否 ✅ |
| go test | all pass ✅ |

## 8. P0 / P1 / P2

| 级别 | 编号 | 问题 | 状态 |
|------|------|------|------|
| — | — | **P0 = 0** | ✅ |
| — | — | **P1 = 0** | ✅ |
| P2 | P2-1 | OPT LLM 输出格式问题（json_repair_schema_failed） | Prompt 后续优化 |
| P2 | P2-2 | Rec 首次需要 repair | OK |
| P2 | P2-3 | 浏览器人工验收未执行 | 后置 |
| P2 | P2-4 | timeout 需 120s（大 Prompt 场景） | 配置调整 |

## 9. 判定

**K5-2 = partial_pass。** 

- Match Analysis：真实 LLM 首次通过 ✅
- Recommendation：真实 LLM 通过（repair once）✅
- Resume Optimization：真实 LLM 调用成功但 JSON 格式不符，fallback 可靠 ✅

2/3 Agent 真实 LLM 成功，1 Agent fallback 正常。P0=0 P1=0。

## 10. 是否可以进入下一阶段

可以进入 K5-3（浏览器端到端验收含真实 LLM 结果）或修复 OPT Prompt/Schema 后重测。
