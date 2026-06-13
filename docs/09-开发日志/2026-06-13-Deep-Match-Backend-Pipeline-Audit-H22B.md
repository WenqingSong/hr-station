# Deep-Match-Backend-Pipeline-ReadOnly-Audit-H22B

> 2026-06-13 | 只读审计 | 根因已定位

## 1. 总体结论

| 项目 | 结果 |
|------|------|
| 完成审计 | ✅ |
| 修改代码 | ❌ |
| **主断层** | **F**: DeepMatchReport 完整 Schema 存在但未被真实链路调用 |
| 真实链路 | legacy 3-agent pipeline (tag evidence) → sanitize 包装 |
| DeepMatchReport 新链路 | ❌ 从未进入执行路径 |
| 进入 H22-C | ✅ |

## 2. 真实调用链

```
HTTP POST /api/jobs/{jobId}/match
→ handler
→ MatchService.Analyze() [match_service.go:92]
→ isQuality=true → s.runDeepMatchSEP() [match_service.go:143]
→ s.matchWith3Agents() [match_deep_agent.go:31]
→ agent.RunDeepMatch3Agents() [match_deep_agents.go:185]
→ Agent1: BuildMatchContext (rule-based, tag matching) [match_deep_agents.go:206]
→ Agent2: LLM → prompt.PromptMatchDecision [match_deep_agents.go:229]
→ Agent3: LLM → prompt.PromptMatchVerifier [match_deep_agents.go:260]
→ output: MatchFinalOutput (Score/Level/MatchedPoints/GapPoints/RiskPoints)
→ map to model.MatchAnalysisResult [match_service.go:546-562]
→ sanitizeDeepResponse() [match_deep_agent.go:54]
  → generated_by = "deep_match_agent"
  → null → []
  → risk strings humanized
  (DeepMatchReport fields NOT populated)
→ MatchResponse → API
```

## 3. 关键发现：两个 Prompt 体系

| Prompt | 文件 | 输出 | 是否使用 |
|--------|------|------|---------|
| `deepMatchSystemPrompt` | match_deep_prompt.go:12-60 | DeepMatchReport 12 module | ❌ 从未调用 |
| `prompt.PromptMatchDecision` | prompt/ | MatchDecision (score/points/risks) | ✅ 实际使用 |
| `prompt.PromptMatchVerifier` | prompt/ | MatchFinalOutput (legacy) | ✅ 实际使用 |

**`deepMatchSystemPrompt` 是死代码**：定义在 `match_deep_prompt.go` 中，指定了完整的 12 模块 JSON Schema，但 `buildDeepMatchPrompt()` 从未被 `runDeepMatchSEP` 或 `matchWith3Agents` 调用。

## 4. DeepMatchReport Schema 审计

| 模块 | 是否定义 | 是否填充 | 是否进 API |
|------|---------|---------|-----------|
| job_understanding | ✅ DeepMatchReport | ❌ | ❌ |
| requirement_coverage | ✅ DeepMatchReport | ❌ | ❌ |
| resume_evidence | ✅ DeepMatchReport | ❌ | ❌ |
| core_match_points | ✅ DeepMatchReport | ❌ | ❌ |
| key_risks | ✅ DeepMatchReport | ❌ | ❌ |
| resume_rewrite_strategy | ✅ DeepMatchReport | ❌ | ❌ |
| interview_risk_questions | ✅ DeepMatchReport | ❌ | ❌ |
| agent_trace | ✅ DeepMatchReport | ❌ | ❌ |
| score/level/summary | ✅ legacy | ✅ | ✅ |

## 5. 3-agent 实际输出

| Agent | 输入 | 输出结构 |
|-------|------|---------|
| Agent1 | TagProfile | MatchContext (tag match + evidence snippets) |
| Agent2 | context JSON | **MatchDecision** {Score, Level, Summary, MatchedPoints, GapPoints, RiskPoints, ImprovementSuggestions} |
| Agent3 | decision + context | **MatchFinalOutput** (same fields as MatchDecision + AgentTrace) |

**Prompt 不要求 job_understanding / evidence mapping / rewrite strategy / interview questions。**

## 6. sanitizeDeepResponse 行为

| 操作 | 效果 |
|------|------|
| generated_by → deep_match_agent | ✅ |
| agent_status → match_verified | ✅ |
| null slices → [] | ✅ |
| risk strings humanized | ✅ |
| **DeepMatchReport 字段填充** | ❌ 无 |

## 7. DB 落库

matchWith3Agents 未调用 matchRepo.Create()。DeepMatchReport 未序列化保存。最近 DB 记录为旧 mock_quality。

## 8. 主因判断

| 选项 | 判断 |
|------|------|
| **F (主因)** | ✅ 实际走 legacy 3-agent，仅包装 deep_match_agent |
| A (次因) | ✅ 模型未被要求输出 DeepMatchReport |
| C (次因) | ✅ matchWith3Agents 映射到 legacy MatchAnalysisResult，丢弃 DeepMatchReport 可能性 |

## 9. 问题清单

| 等级 | 问题 | 根因 |
|------|------|------|
| **P0** | DeepMatchReport 从未进入 API | `deepMatchSystemPrompt` 死代码，`buildDeepMatchPrompt` 未被调用 |
| **P0** | evidence_chain=[] | 3-agent 不产出 structured evidence |
| P1 | job_understanding=null | Prompt 不要求 JD 解构 |
| P1 | DB 不保存 Deep 结果 | matchWith3Agents 不调 matchRepo |

## 10. H22-C 契约冻结建议

1. API response 必须直接包含 DeepMatchReport 顶层字段
2. 激活 `deepMatchSystemPrompt` + `buildDeepMatchPrompt` 进入真实链路
3. `sanitizeDeepResponse` 不应只是包装 legacy，应反序列化 LLM 输出为 DeepMatchReport
4. 保留 legacy 字段作为兼容（gap_points/risk_points/summary 从 DeepMatchReport 推导）

## 11. H22-D 重构要点

修改文件：`match_deep_agent.go`（激活 `buildDeepMatchPrompt`）、`match_service.go`（Deep 路径重定向）、`match_deep_repair.go`（repair → DeepMatchReport not legacy）

## 12. 结论

1. **`deepMatchSystemPrompt` 是完整的 12 模块 JSON Schema 但从未被调用。**
2. **真实执行链路是 3-agent tag-based pipeline**，LLM 只被要求输出 gap/risk/suggestion 字符串列表。
3. `generated_by=deep_match_agent` 只是 `sanitizeDeepResponse` 的标签替换。
4. **当前问题主要是后端断层**：Schema 已定义但未接入执行路径。
5. **前端不是主因**：9 模块正确消费可用字段，缺失字段展示专业空态。
6. **H22-C 须冻结新契约**，**H22-D 须激活 deepMatchSystemPrompt 并重构 Deep 路径**。
