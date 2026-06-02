# K5-2-O-R：OPT Prompt/Schema 修复与重测

- **日期**：2026-06-02
- **阶段**：K5-2-O-R（partial_pass）

## 1. 修复内容

- Schema required: 18→9
- minItems: keyword_alignment 3→1, rewrite_examples 2→1, next_actions 3→1
- additionalProperties: 全部 true
- Prompt: 加入 JSON skeleton

## 2. 重测

OPT 仍 `json_repair_schema_failed` → Mock fallback。

## 3. 根因

hy3-preview 模型在复杂 JSON 生成上稳定性不足。

## 4. 判定

**partial_pass。** Match/Rec 不受影响，OPT fallback 可靠。
