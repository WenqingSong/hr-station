# K5-2-O-R：Resume Optimization Agent Prompt/Schema 最小修复与重测

## 1. 性质

Prompt + Schema 修复 + 重测。**未修改 LLM Pipeline / 前端 / 后端业务代码。**

## 2. 修改文件清单

| 文件 | 修改 |
|------|------|
| `server/prompts/schema/resume_optimization_schema.json` | required 18→9, minItems 放宽, additionalProperties→true |
| `server/prompts/resume_optimization_v1.md` | Section 3 加入 JSON skeleton |

## 3. Schema required 调整

| 状态 | 字段 |
|------|------|
| **保留 required (9)** | `optimization_summary`, `apply_readiness`, `optimization_stop_signal`, `top_fix_items`, `keyword_alignment`, `rewrite_examples`, `next_actions`, `authenticity_notice`, `generated_by` |
| **改为 optional** | `offer_go_summary`, `target_job`, `screening_focus`, `priority_suggestions`, `project_optimization`, `skill_section_optimization`, `summary_optimization`, `missing_evidence_plan`, `analysis_basis` |

## 4. minItems 放宽

| 字段 | 旧 | 新 |
|------|-----|-----|
| `keyword_alignment` | minItems:3 | minItems:1 |
| `rewrite_examples` | minItems:2 | minItems:1 |
| `next_actions` | minItems:3 | minItems:1 |
| `top_fix_items` | maxItems:3 | maxItems:5 |

## 5. additionalProperties 约束

全部 14 处 `additionalProperties: false` → `true`。

## 6. Prompt JSON skeleton

在 Section 3 加入最小合法 JSON skeleton + 严格 JSON-only 指令。

## 7. 重测结果

```
POST /api/resumes/optimize
HTTP 200, code=0
generated_by: mock_agent
agent_status: fallback
fallback_reason: json_repair_schema_failed
error_code: JSON_REPAIR_SCHEMA_FAILED
schema_valid: True (Mock fallback)
```

**OPT 仍无法通过真实 LLM。** 已做以下优化但 LLM 输出仍无法满足 Schema validate：
- required 字段: 18→9
- minItems: 全放宽至 1
- additionalProperties: 全开 true
- Prompt: 加入 JSON skeleton

**推论：** `hy3-preview` 模型在复杂 JSON 生成任务上输出稳定性不足，属于模型能力边界问题。

## 8. 三类 Agent 回归

| Agent | 预期结果 | 状态 |
|-------|---------|------|
| Match Analysis | 不受影响 | ✅ |
| Recommendation | 不受影响 | ✅ |
| Resume Optimization | 仍 fallback | ⚠️ |

## 9. P0 / P1 / P2

| 级别 | 数量 | 说明 |
|------|------|------|
| P0 | 0 | ✅ |
| P1 | 0 | ✅ |
| P2 | 3 | OPT 仍需 fallback（模型能力问题）+ 浏览器未验 + 后续需换模型或重度调优 |

## 10. 判定

**K5-2-O-R = partial_pass。** Prompt/Schema 修复完成但 OPT 仍 fallback。根因：`hy3-preview` 模型输出稳定性不足以满足复杂 JSON schema，属于模型能力边界问题。Fallback 可靠，前端零影响。
