# 2026-05-31-114-ResumeOptimizationAgentPipeline最小集成

## 1. 本次任务

阶段 4-3.8.1-R：ResumeOptimizationAgent Pipeline 最小集成

将 ResumeOptimizationAgent 从 Mock-only 补齐为 LLM Pipeline + Mock fallback，使三类业务 Agent 均具备 LLM + fallback 基础能力。

## 2. 背景

前置核验阶段 4-3.8.1-R-A 确认 ResumeOptimizationAgent 仍为 Mock-only：
- 不存在 `resume_optimization_agent.go`（LLM Agent）
- ResumeOptimizationService 没有 PipelineRunner / llmEnabled
- 没有 LLM success 路径
- 没有 LLM error → Mock fallback 路径
- 没有 ResumeOptFallbackError

MatchAnalysisAgent 和 RecommendationAgent 已完成 LLM + fallback 接入，ResumeOptimizationAgent 是最后一个。

## 3. 云原生开发环境重启说明

云原生开发环境已重启，本阶段重新读取当前代码和文档，未依赖旧终端状态。

## 4. 修改范围

- `server/internal/agent/resume_optimization_agent.go`（新增）
- `server/internal/agent/resume_optimization_agent_test.go`（新增）
- `server/internal/service/resume_optimization_service.go`（修改）
- `server/internal/service/resume_optimization_service_test.go`（修改）
- `server/cmd/server/main.go`（最小修改：注入 pipelineRunner + llmEnabled）
- `docs/`（状态文档更新）

## 5. ResumeOptimizationAgent 新增

### 5.1 LLM Agent

新增 `ResumeOptimizationAgent`，实现 `ResumeOptimizer` 接口：
- 构造函数 `NewResumeOptimizationAgent(runner PipelineRunner)`
- `Optimize(ctx, input)` 方法
- 使用 `TargetResumeOptimization` 复用已有 Pipeline target
- 使用 `prompt.PromptResumeOptimization` + `schema.SchemaResumeOptimization`
- PromptValues：RESUME_CONTEXT_JSON / TARGET_JOB_CONTEXT_JSON / MATCH_ANALYSIS_JSON / BUSINESS_CONTEXT

### 5.2 字段映射

从 Pipeline Parsed (map[string]any) → model.ResumeOptimizationResult：
- optimization_summary → Summary
- apply_readiness → ApplyReadiness
- optimization_stop_signal → OptimizationStopSignal
- top_fix_items → TopFixItems
- priority_suggestions → LongTermImprovement
- keyword_alignment → JDEvidenceChain
- authenticity_notice → VerificationNotice

### 5.3 ResumeOptimizer 接口

统一接口，Mock Agent 和 LLM Agent 均实现：
```go
type ResumeOptimizer interface {
    Optimize(ctx context.Context, input *OptimizeInput) (*model.ResumeOptimizationResult, error)
}
```

## 6. Service Pipeline 注入

ResumeOptimizationService 新增：
- `pipelineRunner agent.PipelineRunner`
- `llmEnabled bool`

构造函数：`NewResumeOptimizationService(repo, mockAgent, pipelineRunner, llmEnabled)`

行为：
- llmEnabled=false → 直接走 Mock
- llmEnabled=true + Pipeline 成功 → 返回 LLM 结果
- llmEnabled=true + Pipeline 失败 → 走 Mock fallback

## 7. Mock fallback 保留

- Mock Agent 完全保留，未修改
- Fallback 路径：当 LLM 失败时，Service 调用 `s.mockAgent.Optimize()`
- Fallback 结果设置 `agent_status=fallback`、`fallback_reason`、`error_code`
- `generated_by` 保持 `mock_agent`（与 Match/Recommendation 一致）

## 8. generated_by 规则

| 场景 | generated_by | agent_status |
|------|-------------|-------------|
| LLM 成功 | llm_agent | success / repaired |
| Mock 直接返回 | mock_agent | mock_success |
| LLM 失败后 fallback | mock_agent | fallback |

## 9. 测试结果

### Agent 测试（8 项）
- TestResumeOptimizationAgent_LLMSuccess ✅
- TestResumeOptimizationAgent_FallbackRequired ✅
- TestResumeOptimizationAgent_RepairedSuccess ✅
- TestResumeOptimizationAgent_MappingFailure ✅
- TestResumeOptimizationAgent_PreservesCriticalFields ✅
- TestResumeOptFallbackError_Error ✅
- TestResumeOptimizationAgent_EnsureInterface ✅
- TestResumeOptimizationAgent_DefaultFieldsWhenMissing ✅

### Service 测试（5 项新增 LLM + fallback）
- TestResumeOptimizationService_LLMDisabled_UseMock ✅
- TestResumeOptimizationService_LLMEnabled_PipelineSuccess ✅
- TestResumeOptimizationService_LLMEnabled_PipelineFallbackToMock ✅
- TestResumeOptimizationService_LLMEnabled_NilRunner_UseMock ✅
- TestResumeOptimizationService_LLMFallback_PreservesTopFixItems ✅

### go test ./...
全部通过

### make llm-check
通过

## 10. 未修改内容

- 未修改 Prompt / Schema
- 未修改前端
- 未修改 Makefile / docker-compose / Nginx
- 未调用真实 LLM
- 未请求腾讯招聘接口
- 未执行 reset / reset-up
- 未清空 volume

## 11. 当前限制

- ResumeOptimizationAgent 字段映射中，keyword_alignment → JDEvidenceChain 的对齐状态映射是简化的，后续 4-3.8.1 集成测试可验证真实 LLM 输出的映射质量
- MATCH_ANALYSIS_JSON 当前传空 `{}`，后续可考虑支持传入已有匹配分析结果

## 12. 是否可以进入 4-3.8.1

可以。三类业务 Agent 均已完成 LLM Pipeline 最小集成：
- MatchAnalysisAgent：LLM + fallback ✅
- RecommendationAgent：LLM + fallback ✅
- ResumeOptimizationAgent：LLM + fallback ✅

## 13. 下一步建议

进入 4-3.8.1：三类 Agent LLM / fallback 本地集成测试补齐
