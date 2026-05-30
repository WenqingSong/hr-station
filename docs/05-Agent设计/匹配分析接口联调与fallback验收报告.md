# 匹配分析接口联调与 fallback 验收报告

> 创建时间：2026-05-30
> 所属阶段：4-3.3
> 验收人：CodeBuddy

## 1. 验收目标

验证匹配分析接口在以下场景下都稳定：

1. LLM success — Pipeline 返回 AgentStatus=success
2. JSON repair 成功 — Pipeline 返回 AgentStatus=repaired
3. fallback_required — Pipeline 返回 fallback_required，回退 Mock Agent
4. LLM_ENABLE=false — 直接走 Mock Agent
5. LLM 配置异常 — 配置缺失/错误时稳定回退 Mock
6. Pipeline / Agent 异常 — 运行时异常时不 500

## 2. 验收范围

- 匹配分析接口 `POST /api/jobs/{id}/match`
- MatchAnalysisAgent（LLM Pipeline 路径）
- MatchMockAgent（Mock fallback 路径）
- MatchService（LLM 优先 + Mock 回退）
- 字段映射（LLM Schema → model.MatchAnalysisResult）
- generated_by / agent_status / repair_applied / fallback_reason / error_code 字段
- 安全边界（不泄露 API Key、不输出完整 raw output）

## 3. 当前匹配分析调用链路

```
Client → POST /api/jobs/{id}/match
  → MatchHandler.Analyze (server/internal/http/match_handler.go)
    → MatchService.Analyze (server/internal/service/match_service.go)
      → if llmEnabled && pipelineRunner != nil:
        → analyzeWithLLM:
          → agent.NewMatchAnalysisAgent(pipelineRunner)
          → MatchAnalysisAgent.Analyze(ctx, resume, job):
            → buildPromptValues(RESUME_CONTEXT_JSON, JOB_CONTEXT_JSON, BUSINESS_CONTEXT)
            → pipelineRunner.Run(ctx, runReq)
            → if FallbackRequired: return MatchFallbackError
            → mapParsedToMatchResult(parsed, resume, job)
            → set GeneratedBy/AgentStatus/RepairApplied
          → if MatchFallbackError: fall through to analyzeWithMock
          → if success: save snapshot to match_analyses table
      → else:
        → analyzeWithMock:
          → mockAgent.Analyze(ctx, resume, job)
          → set GeneratedBy="mock_agent", AgentStatus="mock"/"fallback"
          → save snapshot to match_analyses table
    → MatchResponse{AnalysisID, GeneratedBy, Result, AnalysisBasis, CreatedAt}
```

关键代码位置：

| 项目 | 位置 |
|------|------|
| 匹配分析接口路径 | POST /api/jobs/{id}/match |
| 请求方法 | POST |
| 请求体 | `{resume_version_no?: int}` |
| 响应体 | `{code, message, data: {analysis_id, generated_by, result, analysis_basis, created_at}}` |
| Service 调用 Agent | match_service.go:82-93 |
| Agent 调用 Pipeline | match_agent.go:77 |
| Mock fallback 调用位置 | match_service.go:92, 96 |
| generated_by 赋值 | match_agent.go:106 (LLM), match_service.go:170 (Mock) |
| agent_status 赋值 | match_agent.go:107 (LLM), match_service.go:172/177 (Mock) |
| repair_applied 赋值 | match_agent.go:108 (LLM), match_service.go:175 (Mock) |
| fallback_reason / error_code 赋值 | match_service.go:173-174 (Mock fallback) |
| LLM_ENABLE=false 路径 | match_service.go:96 (直接 Mock) |
| Pipeline error 路径 | match_agent.go:86-92 → match_service.go:89-92 (回退 Mock) |

## 4. 测试环境

| 项目 | 说明 |
|------|------|
| 测试方式 | 单元测试 + fake Pipeline（不调用真实 Hunyuan） |
| 测试框架 | Go testing |
| fake Pipeline | fakePipelineRunner 实现 agent.PipelineRunner 接口 |
| 测试覆盖包 | server/internal/agent, server/internal/service |
| 是否调用真实 Hunyuan | 否 |
| 是否调用真实腾讯接口 | 否 |

## 5. LLM success 场景验收

**测试用例**：`TestMatchService_LLM_Success`

**Pipeline 模拟**：
- AgentStatus = success
- GeneratedBy = "llm_agent"
- Parsed = 完整匹配分析 LLM 输出
- FallbackRequired = false

**验收结果**：

| 检查项 | 结果 |
|--------|------|
| generated_by = "llm_agent" | ✅ 通过 |
| agent_status = "success" | ✅ 通过 |
| repair_applied = false | ✅ 通过 |
| fallback_reason 为空 | ✅ 通过 |
| error_code 为空 | ✅ 通过 |
| match_score 有值 | ✅ 通过 |
| screening_result 正确映射 | ✅ 通过 |
| evidence_chain 非空 | ✅ 通过 |
| 不调用 match_mock | ✅ 通过 |
| 不泄露 API Key | ✅ 通过 |
| 不返回完整 raw output | ✅ 通过 |

## 6. LLM repaired 场景验收

**测试用例**：`TestMatchService_LLM_Repaired`

**Pipeline 模拟**：
- AgentStatus = repaired
- GeneratedBy = "llm_agent"
- RepairApplied = true
- RepairSuccess = true
- FallbackRequired = false

**验收结果**：

| 检查项 | 结果 |
|--------|------|
| generated_by = "llm_agent"（非 json_repair_agent） | ✅ 通过 |
| agent_status = "repaired" | ✅ 通过 |
| repair_applied = true | ✅ 通过 |
| fallback_reason 为空 | ✅ 通过 |
| 返回结构稳定 | ✅ 通过 |
| repair 成功仍视为 LLM 结果 | ✅ 通过 |

## 7. fallback_required 场景验收

**测试用例**：`TestMatchService_LLM_FallbackRequired`

**Pipeline 模拟**：
- AgentStatus = fallback_required
- FallbackRequired = true
- FallbackReason = "llm_call_failed"
- ErrorCode = "LLM_HTTP_ERROR"

**验收结果**：

| 检查项 | 结果 |
|--------|------|
| 返回 MatchFallbackError | ✅ 通过 |
| fallback_reason = "llm_call_failed" | ✅ 通过 |
| error_code = "LLM_HTTP_ERROR" | ✅ 通过 |
| Service 层回退 Mock Agent | ✅ 通过 |
| Mock 结果 generated_by = "mock_agent" | ✅ 通过 |
| Mock 结果 agent_status = "fallback" | ✅ 通过 |
| Mock 结果 fallback_reason 有值 | ✅ 通过 |
| Mock 结果 error_code 有值 | ✅ 通过 |
| 用户拿到可用匹配分析结果 | ✅ 通过 |
| 不暴露底层技术错误给用户 | ✅ 通过（仅 meta 字段保留） |

## 8. LLM_ENABLE=false 场景验收

**测试用例**：`TestMatchService_LLMDisabled_UseMock`

**验收结果**：

| 检查项 | 结果 |
|--------|------|
| pipelineRunner = nil | ✅ 通过 |
| llmEnabled = false | ✅ 通过 |
| Mock Agent 可正常调用 | ✅ 通过 |
| 不调用 Hunyuan | ✅ 通过 |
| generated_by = "mock_agent" | ✅ 通过 |
| agent_status = "mock" | ✅ 通过 |
| 接口不 500 | ✅ 通过 |
| match_score 在 0-100 范围 | ✅ 通过 |

## 9. LLM 配置异常场景验收

**测试用例**：`TestMatchService_LLMConfigError_FallbackToMock`, `TestMatchService_PipelineError_FallbackToMock`

**Pipeline 模拟**：
- 配置缺失：FallbackReason = "llm_config_missing", ErrorCode = "LLM_CONFIG_MISSING"
- Pipeline 错误：FallbackReason = "json_parse_failed", ErrorCode = "JSON_PARSE_FAILED"

**验收结果**：

| 检查项 | 结果 |
|--------|------|
| 不泄露配置细节 | ✅ 通过（ErrorMessage 不含 API Key） |
| 不输出 API Key | ✅ 通过 |
| 不返回 500 | ✅ 通过（回退 Mock） |
| fallback_reason 体现配置失败 | ✅ 通过 |
| error_code 体现配置失败 | ✅ 通过 |
| Pipeline JSON parse 失败时回退 Mock | ✅ 通过 |

## 10. 请求参数异常场景验收

**测试用例**：`TestMatchService_InvalidJobID`

**验收结果**：

| 检查项 | 结果 |
|--------|------|
| job_id 无效时 handler 返回参数错误 | ✅ 通过（handler 层校验） |
| resume 缺失时 service 返回错误 | ✅ 通过（service 层校验） |
| 参数错误不被 fallback 掩盖 | ✅ 通过 |
| 仍返回业务参数错误 | ✅ 通过 |

## 11. 接口返回结构验收

**测试用例**：`TestMatchService_Response_Structure`, `TestMatchService_LLMFieldMapping_Complete`, `TestMatchService_Omitempty_Fields`

**验收结果**：

| 检查项 | 结果 |
|--------|------|
| 原有核心字段仍存在 | ✅ 通过 |
| 新增字段不破坏旧字段 | ✅ 通过（omitempty） |
| generated_by 合理 | ✅ 通过（llm_agent / mock_agent） |
| agent_status 合理 | ✅ 通过（success / repaired / fallback / mock） |
| repair_applied 合理 | ✅ 通过 |
| fallback_reason / error_code 仅在 fallback 语义下出现 | ✅ 通过 |
| match_score 合理（0-100） | ✅ 通过 |
| match_level / match_label 合理 | ✅ 通过（LLM 输出时存在） |
| strengths / gaps / risk_points 安全 | ✅ 通过 |
| authenticity_notice 存在且不鼓励造假 | ✅ 通过 |
| 不返回完整 raw output | ✅ 通过 |
| 不返回完整 Prompt | ✅ 通过 |
| 不返回 API Key | ✅ 通过 |

## 12. 安全边界验收

**测试用例**：`TestMatchService_NoSensitiveDataLeak`

**验收结果**：

| 检查项 | 结果 |
|--------|------|
| 不泄露 API Key | ✅ 通过 |
| 不输出完整 raw output | ✅ 通过 |
| 不输出完整 Prompt | ✅ 通过 |
| 不输出完整简历 | ✅ 通过（仅安全子集进入 PromptValues） |
| 不输出完整 JD | ✅ 通过（仅安全字段进入 PromptValues） |
| 不输出 Authorization header | ✅ 通过 |
| Pipeline safeErr/redactSensitive | ✅ 通过（pipeline 包已实现） |

## 13. 测试结果

| 包 | 测试数 | 通过 | 失败 |
|----|--------|------|------|
| server/internal/agent | 28+ | ✅ 全部 | 0 |
| server/internal/agent/pipeline | 18+ | ✅ 全部 | 0 |
| server/internal/agent/repair | 14+ | ✅ 全部 | 0 |
| server/internal/llm | 27+ | ✅ 全部 | 0 |
| server/internal/llmjson | 21+ | ✅ 全部 | 0 |
| server/internal/prompt | 10+ | ✅ 全部 | 0 |
| server/internal/schema | 9+ | ✅ 全部 | 0 |
| server/internal/service | 15+ | ✅ 全部 | 0 |
| **总计** | **142+** | **✅ 全部通过** | **0** |

编译：`go build ./...` ✅ 通过

## 14. 少量修复记录

| 修复项 | 文件 | 说明 |
|--------|------|------|
| LLM 成功路径缺少快照保存 | match_service.go | `analyzeWithLLM` 成功时未保存匹配分析快照到 match_analyses 表，已修复：添加与 Mock 路径相同的快照保存逻辑 |

## 15. 当前风险

| 风险 | 严重度 | 说明 |
|------|--------|------|
| LLM 调用增加延迟 | 🟡 中 | Pipeline 调用约 2-5 秒，接口响应时间会变长 |
| LLM 输出格式不稳定 | 🟡 中 | 已有 repair once + Mock fallback 兜底 |
| 真实 LLM 联调未执行 | 🟡 中 | 当前验收使用 fake Pipeline，未调用真实 Hunyuan |
| analyzeWithLLM 中 MatchFallbackError 类型断言 | 🟢 低 | 当前实现正确：err 已是 *MatchFallbackError 或被包装 |

## 16. 是否可以进入 4-3.4

**可以进入 4-3.4。**

理由：

1. 匹配分析接口在 6 大验收场景下均通过
2. LLM success / repaired / fallback_required 三路径验证通过
3. LLM_ENABLE=false 时接口稳定
4. 配置异常时接口不 500，正确回退 Mock
5. 安全边界验证通过（不泄露 API Key、不输出完整 raw output）
6. 字段映射完整，omitempty 不破坏现有消费者
7. 修复了 LLM 成功路径快照保存缺失的问题

## 17. 验收结论

**匹配分析接口联调与 fallback 验收通过。**

- 是否真实调用 Hunyuan：否（使用 fake Pipeline）
- 是否使用 fake Pipeline：是
- success 是否通过：✅ 是
- repaired 是否通过：✅ 是
- fallback_required 是否通过：✅ 是
- LLM_ENABLE=false 是否通过：✅ 是
- 配置异常是否通过：✅ 是
- 请求参数异常是否通过：✅ 是
- 是否修改推荐 / 简历优化：否
- 是否修改 Prompt / Schema：否
- 是否修改前端：否
- 是否可以进入 4-3.4：✅ 是

## 18. 下一步建议

1. 进入 4-3.4：RecommendationAgent 接入 Pipeline
2. 如需验证真实 LLM，可在 4-3.4 之前或之后执行一次真实调用
3. 建议在 4-3.8 统一做三类业务 Agent 集成验收时包含真实 LLM 调用
