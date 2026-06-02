# K5-2-O：Resume Optimization Agent JSON 输出失败脱敏诊断

## 1. 性质

read-only diagnosis。**未修改代码，未输出真实 Key。**

## 2. 失败症状

```
POST /api/resumes/optimize
generated_by: mock_agent
agent_status: fallback
fallback_reason: json_repair_schema_failed
error_code: JSON_REPAIR_SCHEMA_FAILED
schema_valid: True (Mock fallback)
```

**失败阶段：** Schema validate → repair attempt → repair also failed → fallback.

## 3. 根因分析

### Schema 复杂度

Resume Optimization Schema 有 **18 个 required 顶级字段** + 嵌套 required 子字段：

| 类别 | 字段 | 说明 |
|------|------|------|
| 简单 required | `optimization_summary`, `offer_go_summary`, `target_job`, `apply_readiness`, `optimization_stop_signal`, `authenticity_notice`, `generated_by` | 7 个标量/字符串 |
| 复杂 required | `top_fix_items` (array), `screening_focus` (array), `keyword_alignment` (array), `project_optimization` (array), `skill_section_optimization` (array), `summary_optimization` (array), `missing_evidence_plan` (array), `rewrite_examples` (array), `next_actions` (array) | 9 个数组 |
| 嵌套 required | `priority_suggestions` (object, 3 sub-required: p0/p1/p2) | 1 个含子 required 的对象 |
| 嵌套 required | `analysis_basis` (object, 5 sub-required) | 1 个含子 required 的对象 |

**合计 required 值：** 18 顶级 + 3 (priority_suggestions 子字段) + 5 (analysis_basis 子字段) = **26+ 个必填值**

### 对比其他 Agent

| Agent | Required 顶级字段 | Required 子字段 |
|-------|------------------|----------------|
| Match Analysis | ~12 | ~3 |
| Recommendation | ~5 (per job) | ~3 (per job) |
| **Resume Optimization** | **18** | **8 (nested)** |

OPT Schema 比 Match/Rec 复杂 **3-4 倍**。

### 修复失败路径

```
LLM Chat → JSON parse ✅
         → Schema validate ❌ (missing fields / type mismatches)
         → JSONRepairAgent repair 
           → Repair Prompt: "不能编造新的业务事实"
           → 无法补齐缺失的复杂字段（project_optimization, analysis_basis 等）
           → 修复后仍 Schema validate ❌
           → json_repair_schema_failed
         → fallback → Mock Agent ✅
```

### 可能的具体缺失字段

基于 Schema 复杂度分析，最可能缺失的字段：

1. **`offer_go_summary`** — 简单字符串，LLM 可能空输出
2. **`analysis_basis`** — 5 个子 required 字段，LLM 容易遗漏子字段
3. **`priority_suggestions.p0/p1/p2`** — 3 个子 required 字段，LLM 可能只输出部分
4. **`project_optimization`** / **`skill_section_optimization`** / **`summary_optimization`** — 数组，在输入无相关内容时 LLM 输出 `[]` 或缺失
5. **`target_job`** — type 标记为 `?`，存在 Schema 定义问题

## 4. 判定

| 检查项 | 结论 |
|--------|------|
| LLM 调用成功 | ✅ |
| JSON parse 成功 | ✅ |
| Schema validate 失败 | ✅（18 required fields 太严格） |
| Repair 失败 | ✅（修复 Agent 受"不编造"约束，无法补全缺失内容） |
| Fallback 可靠 | ✅（Mock Agent 正常返回） |

## 5. 修复建议

| 方案 | 性质 | 工作量 |
|------|------|--------|
| K5-2-O-R：削减 OPT Schema required 字段 | Schema 修改 | 中 |
| K5-2-O-R：增加 Prompt JSON 输出约束 | Prompt 修改 | 小 |
| K5-2-O-R：优化 repair 策略 | 代码修改 | 中 |
| K5-2-S：Schema 宽容度评估 | Docs only | 小 |

**优先建议：** K5-2-O-R — 削减 OPT Schema required 字段。当前 18 required 字段过于严格，建议先移除 `offer_go_summary`, `screening_focus`, `summary_optimization`, `skill_section_optimization`, `analysis_basis` 的 required 约束。

## 6. Key 安全检查

| 检查 | 结果 |
|------|------|
| Key scan | clean ✅ |
| `.env` in git diff | 否 ✅ |

## 7. P0 / P1 / P2

| 级别 | 编号 | 问题 |
|------|------|------|
| — | — | P0 = 0 ✅ |
| — | — | P1 = 0 ✅ |
| P2 | P2-1 | OPT Schema 18 required 字段导致 LLM 输出无法通过 validate |
| P2 | P2-2 | Repair 受"不编造"约束，无法补全缺失字段 |
| P2 | P2-3 | 需要 Schema/Prompt 优化 |

## 8. 最终判定

**K5-2-O = pass（诊断完成）。** 根因：OPT Schema 18 required 字段过于复杂，LLM 输出无法满足。需要 K5-2-O-R 修复。
