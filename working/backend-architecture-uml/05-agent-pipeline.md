# 05 — Agent / Pipeline / Repair / Guardrail 层

来源：`/workspace/server/internal/agent/`（含 `pipeline/`、`repair/`、`guardrail/` 子包）。

## 1. Pipeline 运行时类图

```mermaid
classDiagram
    class PipelineRunner {
        <<interface>>
        +Run(ctx, RunRequest) *RunResult
    }

    class Runner {
        -llmClient llm.Client
        -promptLoader *prompt.Loader
        -schemaLoader *schema.Loader
        -validator *schema.Validator
        -repairAgent *repair.Agent
        -maxTokens int
        -temperature float64
        -repairReasoningEffort string
        +NewRunner(...)
        +Run(ctx, RunRequest) *RunResult
        -validateRequest(req)
        -callLLM(ctx, req, prompt)
        -attemptRepair(...)
        -processRepairResult(...)
    }

    class RepairAgent {
        -llmClient llm.Client
        -promptLoader *prompt.Loader
        -schemaLoader *schema.Loader
        -validator *schema.Validator
        +NewAgent(...)
        +Repair(ctx, RepairRequest) *RepairResult
    }

    class RunRequest {
        +AgentName string
        +PromptName prompt.PromptName
        +SchemaName schema.SchemaName
        +TargetOutputType TargetOutputType
        +PromptValues map~string~string
        +ReasoningEffort string
    }

    class RunResult {
        +AgentStatus AgentStatus
        +GeneratedBy string
        +Parsed map~string~any
        +JSONText string
        +RepairApplied bool
        +RepairSuccess bool
        +FallbackRequired bool
        +FallbackReason string
        +ErrorCode string
        +LLMProvider string
        +LLMModel string
        +DurationMs int64
    }

    PipelineRunner <|.. Runner : 实现
    Runner --> RepairAgent : 修复一次
    Runner --> llm.Client
    Runner --> prompt.Loader
    Runner --> schema.Loader
    Runner --> schema.Validator
    Runner --> llmjson : ExtractAndParse
    RepairAgent --> llm.Client
    RepairAgent --> prompt.Loader
    RepairAgent --> schema.Loader
    RepairAgent --> schema.Validator
    Runner ..> RunRequest
    Runner ..> RunResult
```

## 2. Pipeline 运行流程（统一 LLM 管线）

```mermaid
flowchart TD
    START([Runner.Run]) --> V{validateRequest<br/>agent/prompt/schema/target 非空}
    V -->|非法| FB1["fallback<br/>FallbackRequestInvalid"]
    V -->|合法| PR["promptLoader.Render<br/>渲染 Prompt 模板"]
    PR -->|失败| FB2["fallback<br/>FallbackPromptLoadFailed"]
    PR -->|成功| SL["schemaLoader.Load<br/>加载目标 Schema"]
    SL -->|失败| FB3["fallback<br/>FallbackSchemaLoadFailed"]
    SL -->|成功| CALL["llmClient.Chat<br/>System+User 两条消息"]
    CALL -->|LLM 错误| FB4["fallback<br/>按 ErrorCode 映射<br/>llm_call_failed / llm_timeout ..."]
    CALL -->|成功| EXTRACT["llmjson.ExtractAndParse<br/>JSON 抽取 + 解析"]
    EXTRACT -->|解析失败| REPAIR["repair.Agent.Repair<br/>一次性修复"]
    EXTRACT -->|解析成功| VALIDATE["schema.Validator.Validate<br/>Schema 校验"]
    VALIDATE -->|校验失败| REPAIR
    VALIDATE -->|校验成功| OK1["StatusSuccess<br/>GeneratedBy=llm_agent"]
    REPAIR -->|修复成功| OK2["StatusRepaired<br/>RepairSuccess=true<br/>GeneratedBy=llm_agent"]
    REPAIR -->|修复失败| FB5["StatusFallbackRequired<br/>GeneratedBy=mock_agent<br/>→ 调用方转 Mock"]
```

**核心设计**：
- 管线顺序固定：`Prompt 渲染 → LLM 调用 → JSON 抽取/解析 → Schema 校验 → 失败修复一次 → 兜底`。
- 修复**恰好一次**，无递归、无重试循环。
- `RunResult.AgentStatus` 三态：`success` / `repaired` / `fallback_required`。
- `GeneratedBy`：成功/修复 = `llm_agent`；兜底 = `mock_agent`（由调用方 Service 决定是否切 Mock）。

## 3. TargetOutputType / Schema / Prompt 映射

| TargetOutputType | SchemaName | 用途 |
|------------------|-----------|------|
| `match_analysis` | `match_analysis` | 匹配分析 |
| `recommendation` | `recommendation` | 岗位推荐 |
| `resume_optimization` | `resume_optimization` | 简历优化 |
| `resume_optimization_core` | `resume_optimization_core` | Core 优化 |
| `version_compare` | `version_compare` | 版本对比 |
| `*_decision` / `*_verifier` | 对应 `*_decision` / `*_verifier` | Deep 3-agent 的 Agent2（决策）/ Agent3（校验） |

> `prompt` 包与 `schema` 包各自维护 `PromptName` / `SchemaName` 到文件名的映射，`pipeline` 通过 `TargetOutputType` 桥接两者。

## 4. 业务 Agent 类图

```mermaid
classDiagram
    class MatchAnalysisAgent {
        -runner PipelineRunner
        +NewMatchAnalysisAgent(runner)
        +Generate(...) → model.MatchAnalysisResult
    }
    class RecommendationAgent {
        -runner PipelineRunner
        +NewRecommendationAgent(runner)
        +Generate(...) → []model.RecommendedJob
    }
    class ResumeOptimizationAgent {
        -runner PipelineRunner
        +NewResumeOptimizationAgent(runner)
    }
    class CoreLLMAgent {
        -runner PipelineRunner
        +NewCoreLLMAgent(runner)
    }
    class ProfileAgent {
        -runner PipelineRunner
        +NewProfileAgent(runner)
        +ExtractCandidateProfile(...)
    }
    class VersionCompareLLMAgent {
        -runner PipelineRunner
        +NewVersionCompareLLMAgent(runner)
    }

    MatchAnalysisAgent ..> PipelineRunner
    RecommendationAgent ..> PipelineRunner
    ResumeOptimizationAgent ..> PipelineRunner
    CoreLLMAgent ..> PipelineRunner
    ProfileAgent ..> PipelineRunner
    VersionCompareLLMAgent ..> PipelineRunner
```

**说明**：所有业务 Agent 统一依赖 `PipelineRunner` 接口（在 `match_agent.go` 中定义），**不直接依赖 `llm.Client`**。LLM 调用被封装在 `pipeline.Runner` 内部。

## 5. Deep 3-agent 结构（quality 档）

```mermaid
flowchart LR
    subgraph Deep3["Deep 3-Agent 模式（quality 档）"]
        A1["Agent1（纯规则）<br/>BuildXxxContext / 差距分析"]
        A2["Agent2（LLM 决策）<br/>*_decision 管线"]
        A3["Agent3（LLM 校验）<br/>*_verifier 管线"]
        GUARD["guardrail 本地护栏<br/>(结果后处理安全网)"]
    end
    A1 --> A2
    A2 --> A3
    A3 --> GUARD
    GUARD --> OUT["最终结果<br/>Status=verified_by_local_guard"]
```

| 业务域 | Agent1（规则上下文） | Agent2（LLM 决策） | Agent3（LLM 校验） | 文件 |
|--------|---------------------|-------------------|-------------------|------|
| 匹配 | `BuildMatchContext` | `match_decision` | `match_verifier` | `match_deep_agents.go` |
| 推荐 | 候选池构建 | `recommendation_decision` | `recommendation_verifier` | `recommend_deep_agents.go` |
| 优化 | 差距分析 | `optimization_decision` | `optimization_verifier` | `optimization_deep_agents.go` |
| 对比 | 规则 diff | `compare_decision` | `compare_verifier` | `compare_deep_agents.go` |

> `guardrail` 是**本地确定性护栏**，作为 Agent3 的兜底安全网与结果后处理，**补充但不替代** LLM Verifier。提供 `NormalizeScore`、`ValidateScoreDistribution`、`ValidateWinnerID`、`NormalizeMeta` 等纯函数。

## 6. Mock Agent（规则兜底）

| Mock | 构造 | 职责 |
|------|------|------|
| `MatchMockAgent` | `NewMatchMockAgent()` | 规则生成匹配分析（`AnalyzeFast`/`AnalyzeQuality`） |
| `RecommendationMockAgent` | `NewRecommendationMockAgent()` | 规则岗位推荐（标签抽取/打分/理由） |
| `ResumeOptimizationMockAgent` | `NewResumeOptimizationMockAgent()` | V1 规则优化建议 |
| `ResumeOptimizationMockAgentV2` | `NewResumeOptimizationMockAgentV2()` | 三模式（general/target_direction/target_job）规则优化，可注入 ProfileProvider 与 MatchLatestProvider |
| `VersionCompareMockAgent` | `NewVersionCompareMockAgent()` | 规则版本对比 |
| `CoreMockAgent` | `NewCoreMockAgent()` | Core 模式纯规则兜底 |

**共同点**：全部纯规则实现，不调用 LLM/Pipeline；部分可选注入 `ProfileProvider` 复用画像数据。

## 7. Fast vs Quality 生成链路

```mermaid
flowchart TD
    REQ["业务请求<br/>携带 tier（fast/quality）"]
    REQ --> TIER{tier?}
    TIER -->|fast| FAST["Fast 链路<br/>SEP 语义 / 标签引擎（TagEvidence）<br/>或规则 Mock"]
    TIER -->|quality| QUALITY["Quality 链路<br/>Deep 3-agent / DeepCascade / 5 模块 Hy3<br/>+ 本地裁判"]
    FAST -->|LLM 失败| MOCK["Mock Agent 兜底"]
    QUALITY -->|级联失败/超时| MOCK
    MOCK --> DONE["持久化 ResultJSON 快照<br/>+ GenerationTier 标记"]
    FAST --> DONE
    QUALITY --> DONE
```
