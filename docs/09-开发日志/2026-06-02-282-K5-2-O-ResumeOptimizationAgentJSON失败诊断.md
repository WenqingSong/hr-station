# K5-2-O：OPT JSON 输出失败脱敏诊断

- **日期**：2026-06-02
- **阶段**：K5-2-O（diagnosis, pass）

## 1. 根因

OPT Schema 18 required 字段（26+ 含嵌套）过于复杂，LLM 输出无法满足 Schema validate。Repair 受"不编造"约束，无法补全缺失字段 → `json_repair_schema_failed`。

## 2. 对比

| Agent | Required 顶级 |
|-------|-------------|
| Match | ~12 |
| Recommendation | ~5 (per job) |
| **OPT** | **18** |

OPT 3-4x 复杂于其他 Agent。

## 3. 修复方向

削减 OPT Schema required 字段，或增强 Prompt JSON 输出约束。

## 4. P0/P1/P2

P0=0 P1=0 P2=3

## 5. 判定

**K5-2-O = pass（诊断完成）。** 建议 K5-2-O-R 修复 Schema。
