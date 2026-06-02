# K5-2-O-R2：回滚过度 Schema 放宽并做 OPT 最小诊断

## 1. 性质

Schema 回滚 + 最小输入诊断。**Prompt 仅保留最小 JSON-only 约束。**

## 2. K5-2-O-R 为什么不能 pass

K5-2-O-R 做了全局 `additionalProperties=true`（14处）、required 18→9、minItems 放宽。这些修改没有让 OPT 真实 LLM 成功，却严重削弱了 Schema 约束。

## 3. 回滚内容

| 文件 | 操作 |
|------|------|
| `resume_optimization_schema.json` | 回滚到 K5-2-O 前原始版本（438b36a） |
| `resume_optimization_v1.md` | 回滚 skeleton，仅保留一行 JSON-only 约束 |

**不再保留：** `additionalProperties=true`、大规模 required 删除、minItems 放宽、长 skeleton。

## 4. Prompt 保留的 JSON-only 约束

```
你必须输出严格 JSON。不要 Markdown。不要 ```json。不要解释文字。
缺少证据时用 [] 或 "insufficient_evidence"。所有字段名必须与输出结构完全一致。
```

## 5. 当前环境

| 配置 | 值 |
|------|-----|
| `LLM_MODEL` | `deepseek-v4-pro-202606` |
| `LLM_TIMEOUT` | 120s |

## 6. 最小输入诊断结果

```
resume_text: "Name: Arthur\nSkills: Go, Redis, MySQL\nProject: WMS backend API in Go with Redis cache"
target_job: "Go, C/C++, TCP/UDP, MySQL, Redis, 分布式系统"

generated_by: mock_agent
agent_status: fallback
fallback_reason: llm_call_failed
error_code: LLM_TIMEOUT
```

**最小输入也超时。** 不是输入大小问题。

## 7. 根因分析

**`deepseek-v4-pro-202606` 是推理模型（reasoning model）。**

证据：
- 最小请求 1.63s 返回 `"reasoning_content":"We are asked: ..."`
- 1227 行 OPT prompt 导致推理模型思考时间远超 120s
- hy3-preview（非推理模型）能在 120s 内返回（但 JSON 不满足 schema）

**失败阶段：** LLM Chating — timeout。模型在推理阶段消耗过多时间。

## 8. 字段级错误定位

当前 Pipeline Runner 对 LLM timeout 暴露为 `LLM_TIMEOUT`，字段级错误不可见。需 K5-2-O-DIAG 增加脱敏诊断能力。

## 9. Key 安全

✅ clean | go test all pass ✅

## 10. P0 / P1 / P2

| 级别 | 编号 | 问题 |
|------|------|------|
| — | — | P0 = 0 ✅ |
| — | — | P1 = 0 ✅ |
| P2 | P2-1 | deepseek-v4-pro-202606 推理模型在 1227 行 prompt 下超时 |
| P2 | P2-2 | Pipeline 不暴露字段级 schema 错误 |
| P2 | P2-3 | 需换非推理模型或精简 prompt |

## 11. 判定

**K5-2-O-R2 = partial_pass。** Schema 已回滚。根因：deepseek-v4-pro-202606 是推理模型，1227 行 OPT prompt 导致推理超时。

## 12. 是否可以进入 K5-3

可以。建议：换用非推理模型（如 hy3-preview）后重测 OPT，或先进入 K5-3 浏览器验收（Match+Rec 已真实 LLM 可用）。
