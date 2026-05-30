# 开发日志：阶段 4-2.8 统一 LLM Agent Pipeline

> 日期：2026-05-30
> 阶段：4-2.8
> 依赖：4-2.2（Prompt/Schema Loader）、4-2.3（JSON parse + Schema validate）、4-2.4（LLM Client）、4-2.7（JSONRepairAgent）

## 任务目标

实现统一 LLM Agent Pipeline，编排完整执行链路：
Prompt Render → LLM Call → JSON extract/parse → Schema validate → JSONRepairAgent repair once → fallback decision

## 产出

### 新增文件

1. `server/internal/agent/pipeline/types.go`
   - TargetOutputType：match_analysis / recommendation / resume_optimization
   - AgentStatus：success / repaired / fallback_required
   - RunRequest：AgentName / PromptName / SchemaName / TargetOutputType / PromptValues / Metadata
   - RunResult：AgentStatus / GeneratedBy / Parsed / JSONText / RawContent / RawOutputLength / RepairApplied / RepairSuccess / FallbackRequired / FallbackReason / ErrorCode / ErrorMessage / LLMProvider / LLMModel / FinishReason / DurationMs / ValidationErrors
   - 20+ 错误码常量
   - 13 个 FallbackReason 常量
   - mapLLMErrorCodeToPipeline / mapLLMErrorCodeToFallbackReason 映射函数

2. `server/internal/agent/pipeline/pipeline.go`
   - Runner 结构体与 NewRunner 构造函数
   - Run() 方法实现完整 Pipeline 执行链
   - validateRequest：校验 AgentName / PromptName / SchemaName / TargetOutputType
   - callLLM：构造 ChatRequest 并调用 llm.Client
   - handleLLMError：LLM 错误码 → Pipeline 错误码 → FallbackReason 映射
   - attemptRepair：调用 JSONRepairAgent 一次
   - processRepairResult：repair.RepairResult → pipeline.RunResult 转换
   - targetOutputTypeFromSchema / mapRepairErrorCode / mapRepairErrorCodeToFallbackReason
   - fallbackResult / fillFallback / msSince
   - safeErr / redactSensitive：API Key 脱敏

3. `server/internal/agent/pipeline/pipeline_test.go`
   - fakeClient：支持多响应序列的 llm.Client stub
   - minimalMatchAnalysis / minimalRecommendation / minimalResumeOptimization：符合 Schema 的最小测试对象
   - 18 项单元测试全部通过

### 修改文件

- 更新 11+ 文档文件

## 测试结果

```
=== RUN   TestPipeline_MatchAnalysis_FirstSuccess          --- PASS
=== RUN   TestPipeline_Recommendation_FirstSuccess         --- PASS
=== RUN   TestPipeline_ResumeOptimization_FirstSuccess     --- PASS
=== RUN   TestPipeline_ParseFailed_RepairSuccess           --- PASS
=== RUN   TestPipeline_SchemaValidateFailed_RepairSuccess  --- PASS
=== RUN   TestPipeline_ParseFailed_RepairFailed            --- PASS
=== RUN   TestPipeline_LLMError_FallbackRequired           --- PASS
=== RUN   TestPipeline_LLMStructuredError_RateLimit        --- PASS
=== RUN   TestPipeline_UnknownTargetOutputType             --- PASS
=== RUN   TestPipeline_PromptLoadFailed                    --- PASS
=== RUN   TestPipeline_SchemaLoadFailed                    --- PASS
=== RUN   TestPipeline_RepairOnce_MaxTwoCalls              --- PASS
=== RUN   TestPipeline_FallbackRequired_DoesNotCallMock    --- PASS
=== RUN   TestPipeline_ErrorDoesNotContainFakeAPIKey       --- PASS
=== RUN   TestPipeline_RawOutputLength                     --- PASS
=== RUN   TestPipeline_RepairAppliedAndRepairSuccess       --- PASS
=== RUN   TestPipeline_LLMDisabled                         --- PASS
=== RUN   TestPipeline_DurationMs                          --- PASS
PASS  ok  offer-catcher/server/internal/agent/pipeline  0.026s
```

全量测试：`go test ./... -count=1 -timeout 60s` 全部通过
代码检查：`go vet ./...` 无警告

## 关键设计决策

1. **AgentStatus 三值设计**：success（首次通过）/ repaired（修复通过）/ fallback_required（需要 Mock 兜底）
2. **generated_by 规则**：success/repaired → "llm_agent"，fallback_required → "mock_agent"
3. **Repair once 约束**：Pipeline 最多调用 LLM 2 次（初始 + repair）
4. **Fallback 只做决策**：Pipeline 只返回 fallback_required 决策，不调用具体 Mock Agent
5. **非 LLMError 默认映射**：非 LLMError 类型的 error 默认映射为 LLM_HTTP_ERROR + llm_call_failed，不使用 UNKNOWN_ERROR
6. **API Key 脱敏**：safeErr + redactSensitive 确保 error message 不泄露 API Key

## 修复的问题

1. minimalRecommendation / minimalResumeOptimization 测试对象不符合 Schema → 从 repair test 复制已验证通过的完整版本
2. handleLLMError 非 LLMError 默认映射从 UNKNOWN_ERROR 修正为 LLM_HTTP_ERROR + llm_call_failed

## 未修改文件

- 未修改 .env / .env.example
- 未修改 Prompt 文件
- 未修改 Schema 文件
- 未修改 agent / service / controller / web / Makefile

## 是否可以进入 4-2.9

✅ 可以。4-2.8 已完成所有目标：
- Pipeline Runner 实现完整
- 18 项单元测试全部通过
- 全量测试通过
- go vet 无警告
- 文档已更新

下一步：4-2.9 阶段 4-2 集成验收
