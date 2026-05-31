# 三类 Agent 当前 LLM / fallback 状态核验与验收矩阵

> 创建时间：2026-05-31
> 所属阶段：4-3.8.0
> 核验人：CodeBuddy

## 1. 背景

4-3.7-fix 系列已完成 Prompt / Schema / Go 契约一致性修复和最终闸门验收，4-3.7-fix.4 已完成核心 Agent 返回内容 Smoke 与固定测试简历沉淀。在正式进入 4-3.8 集成验收之前，必须先核验三类 Agent 当前的 LLM / fallback 真实状态，冻结验收矩阵，确定 4-3.8.1 的具体范围。

## 2. 核验范围

| 核验维度 | 核验对象 |
|----------|----------|
| LLM Agent 是否存在 | match_agent.go / recommendation_agent.go / resume_optimization_mock.go |
| Mock Agent 是否存在 | match_mock.go / recommendation_mock.go / resume_optimization_mock.go |
| LLM Pipeline 接入 | MatchService / RecommendationService / ResumeOptimizationService |
| fallback 行为 | Service 层 LLM 优先 + Mock 回退逻辑 |
| JSON Schema 校验 | Pipeline 内 Schema validate |
| JSON repair | Pipeline 内 repair once |
| generated_by 赋值 | LLM success → "llm_agent"，fallback → "mock_agent" |
| 当前测试覆盖 | Agent 层 / Service 层 / Pipeline 层 |

## 3. 当前代码与文件清单

### 3.1 Agent 层

| 文件 | 说明 |
|------|------|
| `server/internal/agent/match_agent.go` | MatchAnalysisAgent（LLM Agent），含 PipelineRunner 接口、MatchFallbackError、PromptValues 组装、字段映射 |
| `server/internal/agent/match_mock.go` | MatchMockAgent（Mock Agent），规则打分 |
| `server/internal/agent/recommendation_agent.go` | RecommendationAgent（LLM Agent），含 PipelineRunner 接口、RecFallbackError、PromptValues 组装、字段映射、候选岗位校验 |
| `server/internal/agent/recommendation_mock.go` | RecommendationMockAgent（Mock Agent），规则打分 |
| `server/internal/agent/resume_optimization_mock.go` | ResumeOptimizationMockAgent（仅 Mock Agent），规则分析 |
| `server/internal/agent/pipeline/pipeline.go` | 统一 Pipeline Runner |
| `server/internal/agent/pipeline/types.go` | RunRequest / RunResult / AgentStatus / ErrorCode / FallbackReason |
| `server/internal/agent/repair/json_repair_agent.go` | JSONRepairAgent（repair once） |

### 3.2 Model 层

| 文件 | 关键字段 |
|------|----------|
| `server/internal/model/match.go` | MatchAnalysisResult 含 ApplyReadiness / OptimizationStopSignal / TopFixItems / GeneratedBy / AgentStatus / RepairApplied / FallbackReason / ErrorCode |
| `server/internal/model/recommendation.go` | RecommendedJob 含 ApplyReadiness / OptimizationStopSignal / TopFixItems / GeneratedBy / AgentStatus / RepairApplied / FallbackReason / ErrorCode |
| `server/internal/model/resume_optimization.go` | ResumeOptimizationResult 含 ApplyReadiness / OptimizationStopSignal / GeneratedBy / AgentStatus / RepairApplied / FallbackRequired / FallbackReason / ErrorCode |

### 3.3 Service 层

| 文件 | LLM 接入状态 |
|------|-------------|
| `server/internal/service/match_service.go` | **已接入**：pipelineRunner + llmEnabled，LLM 优先 + Mock 回退 |
| `server/internal/service/recommendation_service.go` | **已接入**：pipelineRunner + llmEnabled，LLM 优先 + Mock 回退 |
| `server/internal/service/resume_optimization_service.go` | **未接入**：仅 mockAgent，无 pipelineRunner / llmEnabled |

### 3.4 LLM 基础设施

| 文件 | 说明 |
|------|------|
| `server/internal/llm/client.go` | Client 接口（Chat + Provider） |
| `server/internal/llm/noop_client.go` | NoopClient（测试用） |
| `server/internal/llm/hunyuan_client.go` | HunyuanClient（OpenAI-compatible） |
| `server/internal/llm/errors.go` | LLMError / ErrorCode / IsDirectMockFallback |
| `server/internal/llm/types.go` | ChatRequest / ChatResponse / Message / Usage |
| `server/internal/agent/pipeline/pipeline.go` | Pipeline Runner（Prompt→LLM→parse→validate→repair→fallback） |

## 4. MatchAnalysisAgent 状态

### 4.1 LLM Agent

- **文件**：`match_agent.go`
- **结构体**：`MatchAnalysisAgent struct { runner PipelineRunner }`
- **接口**：实现 `MatchAnalyzer` 接口的 `Analyze(ctx, resume, job)` 方法
- **Pipeline 调用**：通过 `runner.Run(ctx, runReq)` 调用 Pipeline
- **PromptValues**：RESUME_CONTEXT_JSON + JOB_CONTEXT_JSON + BUSINESS_CONTEXT
- **字段映射**：`mapParsedToMatchResult` 将 `map[string]any` 映射到 `*model.MatchAnalysisResult`
- **Fallback 触发**：`MatchFallbackError`（Pipeline fallback_required 或映射失败）

### 4.2 Mock Agent

- **文件**：`match_mock.go`
- **结构体**：`MatchMockAgent struct{}`
- **接口**：实现 `MatchAnalyzer` 接口
- **方式**：规则打分 + 证据链生成
- **新字段支持**：`determineMatchApplyReadiness` + `generateMatchTopFixItems`

### 4.3 Service 层接入

- **文件**：`match_service.go`
- **LLM 优先逻辑**：`if s.llmEnabled && s.pipelineRunner != nil` → 先调 LLM → 失败调 Mock
- **Mock 直接路径**：LLM 禁用时直接调 Mock
- **generated_by 赋值**：
  - LLM 成功：`result.GeneratedBy = runResult.GeneratedBy`（即 "llm_agent"）
  - Mock 成功：`result.GeneratedBy = "mock_agent"`
  - Mock fallback：`result.AgentStatus = "fallback"`

### 4.4 JSON Schema / repair

- Pipeline 内置 Schema validate + repair once
- MatchAnalysisAgent 不直接处理 JSON Schema / repair
- Pipeline 返回 success / repaired / fallback_required

### 4.5 当前测试覆盖

| 测试文件 | 覆盖路径 |
|----------|---------|
| `match_agent_test.go` | fake PipelineRunner success / fallback / mapping failure |
| `match_service_test.go` | LLM success / repaired / fallback / LLM disabled / config error / Mock stability |
| `pipeline/pipeline_test.go` | LLM success / malformed JSON / Schema fail / repair / fallback / rate limit / unknown target |
| `match_agent_response_smoke_test.go` | Mock Agent 返回内容 Smoke（apply_readiness / optimization_stop_signal / top_fix_items） |

### 4.6 状态判定

**MatchAnalysisAgent：LLM + fallback 已接入**

- ✅ LLM Agent 存在
- ✅ Mock Agent 存在
- ✅ fallback 行为清晰（MatchFallbackError → Mock 回退）
- ✅ JSON Schema 校验（Pipeline 内置）
- ✅ JSON repair（Pipeline 内置 repair once）
- ✅ generated_by 赋值清晰（llm_agent / mock_agent）
- ✅ LLM 成功路径：Pipeline success/repaired → 映射字段 → 返回
- ✅ LLM 失败路径：Pipeline fallback_required → MatchFallbackError → Service 层调 Mock
- ✅ 映射失败路径：mapParsedToMatchResult 失败 → MatchFallbackError → Service 层调 Mock
- ✅ 测试覆盖：fake Pipeline success / fallback / error / service 层多场景

## 5. RecommendationAgent 状态

### 5.1 LLM Agent

- **文件**：`recommendation_agent.go`
- **结构体**：`RecommendationAgent struct { runner PipelineRunner }`
- **接口**：实现 `Recommender` 接口的 `Recommend(ctx, resume, jobs)` 方法
- **Pipeline 调用**：通过 `runner.Run(ctx, runReq)` 调用 Pipeline
- **PromptValues**：RESUME_CONTEXT_JSON + JOB_CANDIDATES_JSON + BUSINESS_CONTEXT
- **字段映射**：`mapParsedToRecommendedJobs` 将 `map[string]any` 映射到 `[]model.RecommendedJob`
- **候选岗位校验**：job_id 白名单过滤，全部无效触发 fallback
- **Fallback 触发**：`RecFallbackError`（Pipeline fallback_required 或映射失败）

### 5.2 Mock Agent

- **文件**：`recommendation_mock.go`
- **结构体**：`RecommendationMockAgent struct{}`
- **接口**：实现 `Recommender` 接口
- **方式**：规则打分排序取 top 5
- **新字段支持**：`determineRecApplyReadiness` + `generateRecTopFixItems`

### 5.3 Service 层接入

- **文件**：`recommendation_service.go`
- **LLM 优先逻辑**：`if s.llmEnabled && s.pipelineRunner != nil` → 先调 LLM → 失败调 Mock
- **Mock 直接路径**：LLM 禁用时直接调 Mock
- **generated_by 赋值**：
  - LLM 成功：`recommendedJobs[i].GeneratedBy = runResult.GeneratedBy`
  - Mock 成功：`recommendedJobs[i].GeneratedBy = "mock_agent"`
  - Mock fallback：`recommendedJobs[i].AgentStatus = "fallback"`

### 5.4 JSON Schema / repair

- Pipeline 内置 Schema validate + repair once
- RecommendationAgent 不直接处理 JSON Schema / repair
- Pipeline 返回 success / repaired / fallback_required

### 5.5 当前测试覆盖

| 测试文件 | 覆盖路径 |
|----------|---------|
| `recommendation_agent_test.go` | fake PipelineRunner success / fallback / mapping failure / job_id validation |
| `recommendation_service_test.go` | LLM success / repaired / fallback / LLM disabled / config error / candidate filter / empty list |
| `pipeline/pipeline_test.go` | 共享 Pipeline 测试（同 MatchAnalysis） |

### 5.6 状态判定

**RecommendationAgent：LLM + fallback 已接入**

- ✅ LLM Agent 存在
- ✅ Mock Agent 存在
- ✅ fallback 行为清晰（RecFallbackError → Mock 回退）
- ✅ JSON Schema 校验（Pipeline 内置）
- ✅ JSON repair（Pipeline 内置 repair once）
- ✅ generated_by 赋值清晰（llm_agent / mock_agent）
- ✅ LLM 成功路径：Pipeline success/repaired → 映射字段 → 候选校验 → 返回
- ✅ LLM 失败路径：Pipeline fallback_required → RecFallbackError → Service 层调 Mock
- ✅ 映射失败路径：mapParsedToRecommendedJobs 失败 → RecFallbackError → Service 层调 Mock
- ✅ 候选集外岗位路径：job_id 不在候选集 → 过滤 → 全部无效 → fallback
- ✅ 测试覆盖：fake Pipeline success / fallback / error / service 层多场景 / 候选校验

## 6. ResumeOptimizationAgent 状态

### 6.1 LLM Agent

- **不存在**。当前项目中没有 `resume_optimization_agent.go` 文件。
- Pipeline `types.go` 中已定义 `TargetResumeOptimization`，Pipeline 支持 resume_optimization 类型。
- 但没有实际的 ResumeOptimizationAgent 结构体来调用 Pipeline。

### 6.2 Mock Agent

- **文件**：`resume_optimization_mock.go`
- **结构体**：`ResumeOptimizationMockAgent struct{}`
- **接口**：自定义 `Optimize(ctx, input)` 方法
- **方式**：规则分析（关键词匹配、经验链生成）
- **新字段支持**：`determineApplyReadiness` + `generateTopFixItems`
- **generated_by**：硬编码 `"mock_agent"`
- **agent_status**：硬编码 `"mock_success"`

### 6.3 Service 层接入

- **文件**：`resume_optimization_service.go`
- **无 LLM 接入**：`ResumeOptimizationService` 只有 `optimizationRepo` + `mockAgent`
- **没有 pipelineRunner 字段**
- **没有 llmEnabled 字段**
- **直接调用 Mock Agent**：`s.mockAgent.Optimize(ctx, input)`

### 6.4 关键缺失项

| 缺失项 | 说明 |
|--------|------|
| ResumeOptimizationAgent（LLM Agent） | 不存在 `resume_optimization_agent.go` |
| PipelineRunner 依赖 | Service 未注入 pipelineRunner |
| llmEnabled 开关 | Service 未注入 llmEnabled |
| Recommender 类似接口 | 无 `Optimizer` 或类似接口抽象 |
| PromptValues 组装 | 无 `RESUME_CONTEXT_JSON` / `TARGET_JOB_CONTEXT_JSON` / `MATCH_ANALYSIS_JSON` / `BUSINESS_CONTEXT` |
| mapParsedToResult | 无 `map[string]any → ResumeOptimizationResult` 映射函数 |
| FallbackError | 无 `OptimizationFallbackError` 或类似结构 |
| LLM 成功路径 | 不存在 |
| LLM fallback 路径 | 不存在 |

### 6.5 当前测试覆盖

| 测试文件 | 覆盖路径 |
|----------|---------|
| `resume_optimization_mock_test.go` | Mock Agent 正常返回 / generated_by / agent_status / fallback_required / repair_applied / apply_readiness |
| `resume_optimization_service_test.go` | Mock Agent 正常 / 短简历 / 缺少 target_job / 缺少 title 和 JDText |

### 6.6 状态判定

**ResumeOptimizationAgent：LLM + Mock fallback 已接入（4-3.8.1-R 已完成）**

- ✅ Mock Agent 存在
- ✅ LLM Agent 已创建（`resume_optimization_agent.go`）
- ✅ 已接入 Pipeline（使用 `TargetResumeOptimization`）
- ✅ fallback 行为完整（LLM 失败 → Mock 回退，`ResumeOptFallbackError`）
- ✅ Service 已注入 `pipelineRunner` + `llmEnabled`
- ✅ `generated_by` 规则清晰（llm_agent / mock_agent / fallback 时 mock_agent + agent_status=fallback）
- ✅ 测试覆盖 LLM success / fallback / mock-only
- ❌ generated_by 仅硬编码 "mock_agent"
- ❌ 无 JSON repair 测试
- ❌ 无 fake LLM / mock LLM / malformed JSON / error fallback 测试

## 7. LLM / fallback 行为矩阵

| Agent | 当前状态 | LLM success | malformed JSON | LLM error | fallback | generated_by | 是否需要 8.1-R |
|-------|----------|-------------|----------------|-----------|----------|--------------|---------------|
| MatchAnalysis | LLM + fallback 已接入 | 已有（fake Pipeline） | 已有（Pipeline repair） | 已有（Pipeline fallback） | 已有（MatchFallbackError→Mock） | 已有（llm_agent / mock_agent） | 否 |
| Recommendation | LLM + fallback 已接入 | 已有（fake Pipeline） | 已有（Pipeline repair） | 已有（Pipeline fallback） | 已有（RecFallbackError→Mock） | 已有（llm_agent / mock_agent） | 否 |
| ResumeOptimization | **LLM + Mock fallback** | ✅ 存在 | ✅ 存在 | ✅ 支持 | ✅ 支持 | LLM 优先 + Mock fallback | 否 |

## 8. 当前测试覆盖

### 8.1 MatchAnalysisAgent

| 路径 | 测试文件 | 状态 |
|------|----------|------|
| LLM success（fake） | match_agent_test.go / match_service_test.go | ✅ 已有 |
| LLM repaired（fake） | pipeline_test.go | ✅ 已有 |
| LLM error → fallback | match_agent_test.go / match_service_test.go / pipeline_test.go | ✅ 已有 |
| malformed JSON → repair | pipeline_test.go | ✅ 已有 |
| Schema validation fail → repair | pipeline_test.go | ✅ 已有 |
| repair fail → fallback | pipeline_test.go | ✅ 已有 |
| LLM disabled → Mock | match_service_test.go | ✅ 已有 |
| LLM config missing → Mock | match_service_test.go | ✅ 已有 |
| field mapping fail → fallback | match_agent_test.go | ✅ 已有 |
| Mock only（LLM_ENABLE=false） | match_service_test.go | ✅ 已有 |
| generated_by 断言 | match_service_test.go | ✅ 已有 |
| apply_readiness 断言 | match_agent_response_smoke_test.go | ✅ 已有 |
| optimization_stop_signal 断言 | match_agent_response_smoke_test.go | ✅ 已有 |
| top_fix_items 断言 | match_agent_response_smoke_test.go | ✅ 已有 |

### 8.2 RecommendationAgent

| 路径 | 测试文件 | 状态 |
|------|----------|------|
| LLM success（fake） | recommendation_agent_test.go / recommendation_service_test.go | ✅ 已有 |
| LLM error → fallback | recommendation_agent_test.go / recommendation_service_test.go | ✅ 已有 |
| LLM disabled → Mock | recommendation_service_test.go | ✅ 已有 |
| field mapping fail → fallback | recommendation_agent_test.go | ✅ 已有 |
| candidate filter（无效 job_id） | recommendation_agent_test.go | ✅ 已有 |
| all recommendations filtered → fallback | recommendation_agent_test.go | ✅ 已有 |
| generated_by 断言 | recommendation_service_test.go | ✅ 已有 |
| apply_readiness 断言 | recommendation_agent_test.go | ✅ 已有 |
| optimization_stop_signal 断言 | recommendation_agent_test.go | ✅ 已有 |
| top_fix_items 断言 | recommendation_agent_test.go | ✅ 已有 |

### 8.3 ResumeOptimizationAgent

| 路径 | 测试文件 | 状态 |
|------|----------|------|
| Mock 正常返回 | resume_optimization_mock_test.go | ✅ 已有 |
| Mock generated_by 断言 | resume_optimization_mock_test.go | ✅ 已有 |
| Mock apply_readiness 断言 | resume_optimization_mock_test.go | ✅ 已有 |
| Mock optimization_stop_signal | resume_optimization_mock_test.go | ✅ 已有 |
| Mock top_fix_items | resume_optimization_mock_test.go | ✅ 已有 |
| **LLM success（fake）** | — | ❌ 缺失 |
| **LLM error → fallback** | — | ❌ 缺失 |
| **malformed JSON → repair** | — | ❌ 缺失 |
| **fallback generated_by 断言** | — | ❌ 缺失 |

## 9. 4-3.8.1 测试计划

### 9.1 MatchAnalysisAgent 和 RecommendationAgent

这两个 Agent 已完成 LLM + fallback 接入，4-3.8.1 应补齐以下集成验收测试：

1. **统一 fake LLM success 测试**：使用固定简历 + 固定 JD，断言 generated_by="llm_agent"
2. **统一 fake malformed JSON 测试**：Pipeline 返回 malformed JSON → repair → 断言 repair_applied=true
3. **统一 fake LLM error 测试**：Pipeline fallback → Mock → 断言 generated_by="mock_agent"
4. **统一 fallback generated_by 测试**：断言 fallback 时 generated_by="mock_agent" + agent_status="fallback"
5. **统一字段断言**：
   - apply_readiness 非空
   - optimization_stop_signal 非空
   - top_fix_items 字段存在
   - generated_by 值正确
   - evidence / summary / suggestion 字段存在

### 9.2 ResumeOptimizationAgent

此 Agent 仍为 Mock-only，**不具备 4-3.8.1 测试基础**。需要先接入 Pipeline。

## 10. 是否需要 4-3.8.1-R

**需要插入 4-3.8.1-R：ResumeOptimization Pipeline 最小接入。**

理由：

1. ResumeOptimizationAgent 已完成 Pipeline 最小集成（4-3.8.1-R），三类 Agent 均具备 LLM + fallback 能力
2. Pipeline 已支持 `TargetResumeOptimization` 类型
3. Prompt / Schema 已就绪（`resume_optimization_v1.md` + `resume_optimization_schema.json`）
4. Go model 已具备 `generated_by` / `agent_status` / `repair_applied` / `fallback_required` 等字段
5. Service 层已完成改造：注入 `pipelineRunner` + `llmEnabled`，实现 LLM 优先 + Mock 回退

4-3.8.1-R 目标只能是：

1. 创建 `ResumeOptimizationAgent`（LLM Agent），含 `Optimizer` 接口、`OptimizationFallbackError`、PromptValues 组装、字段映射
2. 保留 `ResumeOptimizationMockAgent` 作为 fallback
3. 保持 Prompt / Schema / Go model 一致
4. Service 层注入 PipelineRunner + llmEnabled
5. 不进前端
6. 不改页面
7. 不改启动链路

## 11. 当前风险

| 风险 | 严重度 | 说明 |
|------|--------|------|
| ResumeOptimization LLM 未接入 | 🔴 高 | 三个 Agent 中唯一未接入 LLM Pipeline |
| ResumeOptimization 无 fallback 路径 | 🔴 高 | 没有 LLM Agent 就没有 fallback |
| ResumeOptimization Service 无 LLM 开关 | 🟡 中 | 需要注入 pipelineRunner + llmEnabled |
| TopFixItem 与 ResumeOptimizationFixItem 独立 | 🟢 低 | 后者多一个 learn_more_query 字段，可后续统一 |
| Go 代码未做枚举硬校验 | 🟢 低 | apply_readiness / optimization_stop_signal 枚举由 Prompt 约束 |

## 12. 下一步建议

1. **4-3.8.1-R：ResumeOptimization Pipeline 最小接入** — 创建 ResumeOptimizationAgent、Optimizer 接口、字段映射、Service 层 LLM 优先 + Mock 回退
2. **4-3.8.1：三类 Agent LLM / fallback 本地集成测试补齐** — 统一 fake LLM / malformed JSON / error fallback 测试
3. **4-3.8.2：三类 Agent 统一字段断言验收** — apply_readiness / optimization_stop_signal / top_fix_items / generated_by / evidence / summary / suggestion
