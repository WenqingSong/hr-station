# 阶段 4-3 业务 Agent 接入 LLM 方案

> 创建时间：2026-05-30
> 所属阶段：4-3.0
> 当前阶段：4-3.1 已通过（真实 LLM Smoke 前置验收通过，Hunyuan / TokenHub 最小调用可用）
> 依据文档：阶段4-2LLMClient与Hunyuan接入方案、阶段4-2集成验收报告、阶段4-2LLM基础设施汇总、Agent总览、岗位匹配Agent、岗位推荐Agent、LLMClient设计、Hunyuan接入准备、Prompt模板、JSON输出规范、失败兜底与Mock回退策略、match_mock.go、recommendation_mock.go、match_service.go、recommendation_service.go、match.go、recommendation.go、pipeline/pipeline.go、pipeline/types.go

## 1. 阶段目标

4-3 的目标不是重新实现 LLM 基础设施，而是让业务 Agent 复用 4-2 的 Pipeline。

阶段目标：

- MatchAnalysisAgent 接入 Pipeline
- RecommendationAgent 接入 Pipeline
- ResumeOptimizationAgent 根据当前工程现状接入或预留
- fallback_required 后调用原 Mock Agent
- 保持接口稳定
- 保留 Mock fallback
- 不让业务 Agent 直接调用 HunyuanClient
- 不跳过 Schema validate
- 不跳过 JSONRepairAgent repair once

## 2. 当前前提

### 2.1 已有

| 资产 | 位置 | 说明 |
|------|------|------|
| 腾讯真实岗位数据 | jobs 表 tencent_join | data_quality=full_jd |
| 岗位列表 / 详情接口 | GET /api/jobs | 支持 external_source 筛选 |
| 推荐 Mock Agent | server/internal/agent/recommendation_mock.go | 规则打分 |
| 匹配 Mock Agent | server/internal/agent/match_mock.go | 规则打分 + 证据链 |
| 版本对比 Mock Agent | server/internal/agent/version_compare_mock.go | 字段差异对比 |
| 匹配分析 Prompt | server/prompts/match_analysis_v1.md | HR初筛+技术面试官双视角 |
| 推荐 Prompt | server/prompts/recommendation_v1.md | 四档推荐+Offer鹅语气 |
| 简历优化 Prompt | server/prompts/resume_optimization_v1.md | P0/P1/P2+关键词对齐 |
| JSON 修复 Prompt | server/prompts/json_repair_v1.md | 只修结构不重做业务 |
| 匹配分析 Schema | server/prompts/schema/match_analysis_schema.json | 17 required 字段 |
| 推荐 Schema | server/prompts/schema/recommendation_schema.json | 7 required 字段 |
| 简历优化 Schema | server/prompts/schema/resume_optimization_schema.json | 15 required 字段 |
| Prompt Loader | server/internal/prompt/ | 从 server/prompts/ 加载 4 个 Prompt |
| Schema Loader | server/internal/schema/ | 从 server/prompts/schema/ 加载 3 个 Schema |
| JSON parse + Schema validate | server/internal/llmjson/ + schema/validator.go | 提取/解析/校验 |
| LLM Client 抽象 | server/internal/llm/ | Client interface + Config + ErrorCode + NoopClient |
| Hunyuan Client | server/internal/llm/hunyuan_client.go | OpenAI-compatible |
| JSONRepairAgent | server/internal/agent/repair/ | repair once |
| 统一 Pipeline | server/internal/agent/pipeline/ | Prompt→LLM→parse→validate→repair→fallback decision |
| llm-smoke 入口 | server/cmd/llm-smoke/ | make llm-smoke |

### 2.2 尚未有

| 缺失项 | 说明 |
|--------|------|
| 业务 Agent 调用 Pipeline | 当前 service 直接调用 mockAgent |
| 业务接口返回 LLM 结果 | 当前 generated_by 固定为 "mock_agent" |
| fallback_required 调用具体 Mock Agent | Pipeline 只返回 decision |
| 前端展示 agent_status / repair_applied / fallback_reason | 4-4 统一收口 |
| ResumeOptimizationAgent 的业务代码 | 无 service / agent / controller |
| ResumeOptimizationAgent 的 Mock Agent | 无 resume_optimization_mock.go |
| 简历优化接口 | 无 POST /api/resumes/optimization |
| Pipeline RunResult.Parsed 到 model 结构的字段映射 | RunResult.Parsed 是 map[string]any |

## 3. 阶段边界

### 3.1 4-3 做

- 业务 Agent 接入 Pipeline
- 业务上下文组装
- PromptValues 组装
- Pipeline RunResult 映射为业务结果
- fallback_required 调用原 Mock Agent
- 接口稳定性验收

### 3.2 4-3 不做

- 不重新实现 LLM Client
- 不重新实现 HunyuanClient
- 不重新实现 Prompt Loader
- 不重新实现 Schema Validator
- 不重新实现 JSONRepairAgent
- 不大改前端
- 不重构数据库
- 不重写 Prompt / Schema
- 不绕过 Pipeline

## 4. 4-2 与 4-3 的分工

| 阶段 | 范围 | 产出 |
|------|------|------|
| 4-2 | 建底层能力 | Pipeline / PromptLoader / SchemaLoader / HunyuanClient / JSONRepairAgent / fallback decision |
| 4-3 | 业务 Agent 使用底层能力 | MatchAnalysisAgent 接入 Pipeline → 生成真实匹配分析结果；RecommendationAgent 接入 Pipeline → 生成真实推荐结果 |

核心分工：

```
4-2 的 Pipeline 负责：
Prompt → LLM → JSON parse → Schema validate → repair once → fallback_required decision

4-3 的业务 Agent 负责：
准备 PromptValues → 调用 Pipeline → 成功则映射 LLM 结果 → fallback_required 则调用 Mock Agent
```

## 5. 当前业务 Agent 代码现状

### 5.1 MatchMockAgent

- **文件**：`server/internal/agent/match_mock.go`
- **结构体**：`MatchMockAgent struct{}`
- **方法**：`Analyze(resume *model.ResumeVersion, job *model.Job) (*model.MatchAnalysisResult, error)`
- **调用方**：`server/internal/service/match_service.go` 的 `MatchService.Analyze`
- **返回类型**：`*model.MatchAnalysisResult`
- **当前 behavior**：规则打分生成匹配分数、证据链、风险点、P0/P1 建议等
- **当前依赖**：`model.ResumeVersion`, `model.Job`

### 5.2 RecommendationMockAgent

- **文件**：`server/internal/agent/recommendation_mock.go`
- **结构体**：`RecommendationMockAgent struct{}`
- **方法**：`Recommend(resume *model.ResumeVersion, jobs []*model.Job) ([]model.RecommendedJob, error)`
- **调用方**：`server/internal/service/recommendation_service.go` 的 `RecommendationService.Recommend`
- **返回类型**：`[]model.RecommendedJob`
- **当前 behavior**：规则打分排序取 top 5
- **当前依赖**：`model.ResumeVersion`, `[]*model.Job`

### 5.3 MatchService

- **文件**：`server/internal/service/match_service.go`
- **结构体**：`MatchService struct{ resumeRepo, jobRepo, matchRepo, mockAgent }`
- **核心逻辑**：获取简历 → 获取岗位 → `mockAgent.Analyze(resume, job)` → 组装 MatchResponse → 保存快照
- **当前 generated_by**：硬编码 `"mock_agent"`

### 5.4 RecommendationService

- **文件**：`server/internal/service/recommendation_service.go`
- **结构体**：`RecommendationService struct{ resumeRepo, jobRepo, recommendationRepo, mockAgent }`
- **核心逻辑**：获取简历 → 获取候选岗位列表 → `mockAgent.Recommend(resume, jobs)` → 组装 RecommendationResponse → 保存快照
- **当前 generated_by**：硬编码 `"mock_agent"`

### 5.5 Model 结构

- **MatchAnalysisResult**：JobID, JobTitle, ResumeVersionID, ResumeVersionNo, MatchScore, ScreeningResult, EvidenceChain, MatchHighlights, RiskPoints, OptimizationSuggestions, NextSteps, AuthenticityNotice, ExternalSource, ExternalID, DataQuality, SourceURL, IsFeatured
- **RecommendedJob**：JobID, Title, Department, City, Category, JobType, Priority, RecommendationScore, RecommendationReason, MatchHighlights, RiskPoints, SuggestedAction, ExternalSource, ExternalID, DataQuality, SourceURL, IsFeatured
- **MatchResponse**：AnalysisID, GeneratedBy, Result, AnalysisBasis, CreatedAt
- **RecommendationResponse**：ResumeVersionID, ResumeVersionNo, GeneratedBy, RecommendedJobs, Summary, CreatedAt

### 5.6 LLM Pipeline 输出 Schema 与现有 API Model 对比

- Pipeline `RunResult.Parsed` 类型为 `map[string]any`
- LLM Schema 定义的字段结构与 model 层 Go struct 有对应关系，但不可直接断言
- 4-3.2 需要实现 `map[string]any → *model.MatchAnalysisResult` 字段映射
- 4-3.4 需要实现 `map[string]any → []model.RecommendedJob` 字段映射
- 不能直接把 `map[string]any` 原样返回接口

### 5.7 ResumeOptimizationAgent 当前状态

**关键发现：当前项目中不存在简历优化的业务代码。**

具体核查结果：

| 检查项 | 结果 |
|--------|------|
| resume_optimization_mock.go | ❌ 不存在 |
| resume_optimization_service.go | ❌ 不存在 |
| resume_optimization_handler.go | ❌ 不存在 |
| resume_optimization model | ❌ 不存在 |
| POST /api/resumes/optimization | ❌ 不存在 |
| 版本对比 Mock Agent | ✅ 存在 version_compare_mock.go，但这是"版本对比"，不是"简历优化" |

结论：ResumeOptimizationAgent 目前只有 Prompt + Schema 资产，没有对应业务接口、service、agent 或 controller。不能强行造新接口。

## 6. MatchAnalysisAgent 接入方案

### 6.1 接入目标

MatchAnalysisAgent 是 4-3 优先接入对象。

- 匹配分析接口优先调用 Pipeline
- Pipeline 使用 match_analysis_v1.md
- Pipeline 使用 match_analysis_schema.json
- Pipeline TargetOutputType=match_analysis
- Pipeline 成功或 repaired 时返回 LLM 结果
- Pipeline fallback_required 时调用原 match_mock.go
- 接口不崩
- Mock fallback 保底

### 6.2 PromptValues 来源

MatchAnalysis Prompt 有三个占位符：

| 占位符 | 数据来源 | 说明 |
|--------|---------|------|
| `{{RESUME_CONTEXT_JSON}}` | model.ResumeVersion 序列化为 JSON | 包含 target_role, skills, projects, education, experience 等 |
| `{{JOB_CONTEXT_JSON}}` | model.Job 序列化为 JSON | 包含 title, description, requirements, data_quality 等 |
| `{{BUSINESS_CONTEXT}}` | 业务 Agent 硬编码/常量 | 目标：帮助学生判断是否值得投递该岗位；约束：不能建议虚构经历 |

组装方式：

```text
PromptValues = map[string]string{
    "RESUME_CONTEXT_JSON": json.Marshal(resume),
    "JOB_CONTEXT_JSON":    json.Marshal(job),
    "BUSINESS_CONTEXT":    matchAnalysisBusinessContext,
}
```

注意：PromptValues 中不要放 API Key；日志中不要输出完整简历或完整 JD。

### 6.3 Pipeline 调用

```text
runReq := pipeline.RunRequest{
    AgentName:        "MatchAnalysisAgent",
    PromptName:       prompt.PromptMatchAnalysis,
    SchemaName:       schema.SchemaMatchAnalysis,
    TargetOutputType: pipeline.TargetMatchAnalysis,
    PromptValues:     promptValues,
    Metadata:         map[string]string{"job_id": strconv.FormatUint(job.ID, 10)},
}

runResult := pipeline.Runner.Run(ctx, runReq)
```

### 6.4 LLM 结果字段映射

Pipeline 成功时 `runResult.Parsed` 是 `map[string]any`，需要映射到 `*model.MatchAnalysisResult`：

```text
runResult.Parsed["match_score"]        → MatchAnalysisResult.MatchScore (int)
runResult.Parsed["screening_result"]   → MatchAnalysisResult.ScreeningResult (string)
runResult.Parsed["evidence_chain"]     → MatchAnalysisResult.EvidenceChain ([]MatchEvidence)
runResult.Parsed["match_highlights"]   → MatchAnalysisResult.MatchHighlights ([]string)
runResult.Parsed["risk_points"]        → MatchAnalysisResult.RiskPoints ([]string)
runResult.Parsed["optimization_suggestions"] → MatchAnalysisResult.OptimizationSuggestions ([]MatchSuggestion)
runResult.Parsed["next_steps"]         → MatchAnalysisResult.NextSteps ([]string)
runResult.Parsed["authenticity_notice"] → MatchAnalysisResult.AuthenticityNotice (string)
...
```

4-3.2 需实现一个 `mapToMatchAnalysisResult(parsed map[string]any, job *model.Job, resume *model.ResumeVersion) (*model.MatchAnalysisResult, error)` 函数。

如果映射失败（字段缺失、类型不匹配），应视为 fallback_required，调用 Mock Agent。

### 6.5 fallback_required 策略

```text
Pipeline RunResult.FallbackRequired == true
→ 调用现有 MatchMockAgent.Analyze(resume, job)
→ generated_by = "mock_agent"
→ 可在内部或返回 meta 中记录 fallback_reason / error_code
```

### 6.6 Service 层改造方向

```text
MatchService 新增 pipelineRunner 依赖
→ Analyze 方法中：
  1. 获取简历、岗位（不变）
  2. 组装 PromptValues
  3. pipelineRunner.Run(ctx, runReq)
  4. switch runResult.AgentStatus:
     - success / repaired:
       映射 Parsed → MatchAnalysisResult
       generated_by = "llm_agent"
       如果映射失败，fallback 到 Mock
     - fallback_required:
       mockAgent.Analyze(resume, job)
       generated_by = "mock_agent"
  5. 组装 MatchResponse
  6. 保存快照
```

## 7. RecommendationAgent 接入方案

### 7.1 接入目标

RecommendationAgent 是第二优先级接入对象。

- 推荐接口优先调用 Pipeline
- Pipeline 使用 recommendation_v1.md
- Pipeline 使用 recommendation_schema.json
- Pipeline TargetOutputType=recommendation
- Pipeline 成功或 repaired 时返回 LLM 推荐结果
- Pipeline fallback_required 时调用原 recommendation_mock.go
- 推荐列表稳定返回
- Mock fallback 保底

### 7.2 PromptValues 来源

推荐 Prompt 有三个占位符：

| 占位符 | 数据来源 | 说明 |
|--------|---------|------|
| `{{RESUME_CONTEXT_JSON}}` | model.ResumeVersion 序列化为 JSON | 包含 target_role, skills 等 |
| `{{JOB_CANDIDATES_JSON}}` | 候选岗位列表序列化为 JSON | 从 jobRepo.ListRecommendCandidates 获取 |
| `{{BUSINESS_CONTEXT}}` | 业务 Agent 硬编码/常量 | 目标：帮助学生缩小投递范围；约束：不能编造岗位 |

### 7.3 候选岗位来源

**关键约束**：继续复用当前 RecommendationService / repository 的候选池逻辑。

- 不在 4-3.4 重新设计岗位召回
- 不要让 LLM 自己编造岗位
- LLM 只能在给定 JOB_CANDIDATES_JSON 中推荐

### 7.4 岗位推荐完整性约束

**岗位推荐不能依赖模型生成不存在的岗位。**

- job_id / title / external_id 必须来自输入候选集
- LLM 输出中如果出现候选集之外的 job_id，该推荐项应被丢弃
- 如果 LLM 输出的推荐列表为空或全部无效，应 fallback 到 Mock Agent

### 7.5 LLM 结果字段映射

Pipeline 成功时 `runResult.Parsed` 需要映射到 `[]model.RecommendedJob`：

推荐 Schema 顶层包含 `recommended_jobs` 数组，需要：

```text
parsed["recommended_jobs"] → 遍历每项 → mapToRecommendedJob(item, candidateJobMap)
```

其中 `candidateJobMap` 用于验证 job_id 的合法性，并补充 ExternalSource / ExternalID / DataQuality 等字段。

### 7.6 fallback_required 策略

```text
Pipeline RunResult.FallbackRequired == true
→ 调用现有 RecommendationMockAgent.Recommend(resume, jobs)
→ generated_by = "mock_agent"
→ 保持接口稳定
```

### 7.7 Service 层改造方向

```text
RecommendationService 新增 pipelineRunner 依赖
→ Recommend 方法中：
  1. 获取简历、候选岗位列表（不变）
  2. 组装 PromptValues（RESUME_CONTEXT_JSON + JOB_CANDIDATES_JSON + BUSINESS_CONTEXT）
  3. pipelineRunner.Run(ctx, runReq)
  4. switch runResult.AgentStatus:
     - success / repaired:
       映射 Parsed → []RecommendedJob
       验证 job_id 合法性
       generated_by = "llm_agent"
       如果映射失败或 job_id 全部无效，fallback 到 Mock
     - fallback_required:
       mockAgent.Recommend(resume, jobs)
       generated_by = "mock_agent"
  5. 组装 RecommendationResponse
  6. 保存快照
```

## 8. ResumeOptimizationAgent 现状与接入策略

### 8.1 当前现状

**ResumeOptimizationAgent 目前不具备业务接入条件。**

| 资产 | 状态 |
|------|------|
| Prompt 文件 resume_optimization_v1.md | ✅ 已存在 |
| Schema 文件 resume_optimization_schema.json | ✅ 已存在 |
| Pipeline TargetOutputType=resume_optimization | ✅ Pipeline 已支持 |
| resume_optimization_mock.go | ❌ 不存在 |
| resume_optimization_service.go | ❌ 不存在 |
| resume_optimization_handler.go | ❌ 不存在 |
| resume_optimization model | ❌ 不存在 |
| POST /api/resumes/optimization 接口 | ❌ 不存在 |

### 8.2 接入策略

**不强行造新接口。**

- 4-3.6 先做现状核对与接入方案
- 4-3.7 可以只做后端能力预留或文档登记
- 真正前端化和接口化留到后续阶段

### 8.3 真实性风险

简历优化涉及真实性风险，必须遵守以下约束：

- rewrite_examples 必须基于真实经历
- 不能鼓励编造项目、指标、实习、规模、结果
- 如果模型输出存在过度包装，必须 fallback 或标记 requires_user_verification
- resume_optimization_schema.json 已包含 `requires_user_verification` 字段约束
- Schema 已禁止编造具体数字、指标、系统规模、职责或成果

### 8.4 ResumeOptimization PromptValues

如果 4-3.7 决定做后端能力预留，PromptValues 来源为：

| 占位符 | 数据来源 |
|--------|---------|
| `{{RESUME_CONTEXT_JSON}}` | model.ResumeVersion |
| `{{TARGET_JOB_CONTEXT_JSON}}` | model.Job（注意：匹配分析用 JOB_CONTEXT_JSON，简历优化用 TARGET_JOB_CONTEXT_JSON） |
| `{{MATCH_ANALYSIS_JSON}}` | 可选，来自匹配分析结果 |
| `{{BUSINESS_CONTEXT}}` | 业务约束 |

## 9. Pipeline 调用位置设计

### 9.1 分层原则

- Controller 不直接调用 Pipeline
- Service 负责业务编排
- Agent 层负责 LLM / Mock 能力
- Pipeline 作为 Agent 内部依赖

### 9.2 推荐结构

```text
service
→ 组装 PromptValues
→ pipelineRunner.Run(ctx, runReq)
→ switch runResult.AgentStatus:
   success / repaired → 映射 LLM 结果
   fallback_required → 调用对应 Mock Agent
```

### 9.3 禁止设计

```text
❌ controller → pipeline
❌ business agent → HunyuanClient
❌ service 直接构造 HTTP 请求调用 LLM
```

业务 Agent 只依赖 Pipeline，不直接依赖 HunyuanClient。

### 9.4 依赖注入方向

```text
main.go 中：
pipelineRunner := pipeline.NewRunner(llmClient, promptLoader, schemaLoader, validator, repairAgent)

matchService := service.NewMatchService(resumeRepo, jobRepo, matchRepo, mockAgent, pipelineRunner)
recommendationService := service.NewRecommendationService(resumeRepo, jobRepo, recommendationRepo, mockAgent, pipelineRunner)
```

Service 同时持有 MockAgent 和 PipelineRunner，根据 Pipeline 结果决定是否 fallback。

## 10. PromptValues 组装策略

### 10.1 核心原则

- PromptValues 必须由业务 Agent / Service 基于现有业务数据组装
- 不能让 Pipeline 读取数据库
- 不能让 Pipeline 自己查岗位
- 不能让 Pipeline 自己查简历
- Pipeline 只负责通用执行链路

### 10.2 MatchAnalysis PromptValues

```text
PromptValues = {
    "RESUME_CONTEXT_JSON":  json.Marshal(resume),
    "JOB_CONTEXT_JSON":     json.Marshal(job),
    "BUSINESS_CONTEXT":     matchAnalysisBusinessContext,
}
```

业务上下文建议值：

```text
目标：帮助学生判断是否值得投递该岗位
目标：提升简历初筛命中率
约束：不能建议用户虚构经历
约束：所有建议必须基于已有真实经历
约束：输出必须是严格 JSON
```

### 10.3 Recommendation PromptValues

```text
PromptValues = {
    "RESUME_CONTEXT_JSON":  json.Marshal(resume),
    "JOB_CANDIDATES_JSON":  json.Marshal(jobs),
    "BUSINESS_CONTEXT":     recommendationBusinessContext,
}
```

### 10.4 ResumeOptimization PromptValues（预留）

```text
PromptValues = {
    "RESUME_CONTEXT_JSON":      json.Marshal(resume),
    "TARGET_JOB_CONTEXT_JSON":  json.Marshal(job),
    "MATCH_ANALYSIS_JSON":      json.Marshal(matchResult),  // 可选
    "BUSINESS_CONTEXT":         resumeOptBusinessContext,
}
```

### 10.5 安全约束

- PromptValues 中不要放 API Key
- 日志中不要输出完整简历
- 日志中不要输出完整 JD
- 必要时只记录长度、hash、job_id、agent_name

## 11. fallback_required 处理策略

### 11.1 核心原则

- Pipeline 只返回 fallback_required decision
- 业务 Agent 才负责调用具体 Mock Agent

### 11.2 处理链路

```text
Pipeline RunResult.AgentStatus = success
→ 使用 LLM 结果
→ generated_by = "llm_agent"
→ agent_status = "success"
→ repair_applied = false

Pipeline RunResult.AgentStatus = repaired
→ 使用 LLM repair 后结果
→ generated_by = "llm_agent"
→ agent_status = "repaired"
→ repair_applied = true

Pipeline RunResult.AgentStatus = fallback_required
→ 调用对应 Mock Agent
→ 返回 Mock 结果
→ generated_by = "mock_agent"
→ agent_status = "fallback"
→ fallback_reason = 来自 Pipeline RunResult.FallbackReason
→ error_code = 来自 Pipeline RunResult.ErrorCode
```

### 11.3 fallback_required 不是失败

- fallback_required 是稳定性机制
- 用户侧应看到稳定业务结果，而不是底层技术错误
- 内部日志可记录 fallback_reason / error_code

### 11.4 LLM 结果映射失败的 fallback

即使 Pipeline 返回 success / repaired，如果 `map[string]any` → Model 结构映射失败（字段缺失、类型不匹配），也应 fallback 到 Mock Agent：

```text
llmResult, err := mapToMatchAnalysisResult(runResult.Parsed, job, resume)
if err != nil {
    // 映射失败，fallback 到 Mock
    llmResult, _ = mockAgent.Analyze(resume, job)
    generatedBy = "mock_agent"
    // 记录映射失败原因
}
```

## 12. Mock fallback 集成策略

### 12.1 MatchAnalysisAgent fallback

```text
fallback 时调用：mockAgent.Analyze(resume, job)
返回：*model.MatchAnalysisResult
generated_by = "mock_agent"
```

Mock Agent 保持现有逻辑不变，不修改。

### 12.2 RecommendationAgent fallback

```text
fallback 时调用：mockAgent.Recommend(resume, jobs)
返回：[]model.RecommendedJob
generated_by = "mock_agent"
```

Mock Agent 保持现有逻辑不变，不修改。

### 12.3 ResumeOptimizationAgent fallback

当前不存在 resume_optimization_mock.go，4-3.7 需要决定：

- 选项 A：创建简单的确定性 Mock Agent（返回通用优化建议模板）
- 选项 B：暂不实现，4-3.7 只做后端 Pipeline 能力预留
- 推荐选择 B，不强行造接口

## 13. 返回结构与字段策略

### 13.1 当前接口结构

**匹配分析接口** (POST /api/jobs/{id}/match)：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "analysis_id": 12,
    "generated_by": "mock_agent",
    "result": { ... MatchAnalysisResult ... },
    "analysis_basis": { ... },
    "created_at": "..."
  }
}
```

**推荐接口** (POST /api/jobs/recommend)：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "resume_version_id": 2,
    "resume_version_no": 2,
    "generated_by": "mock_agent",
    "recommended_jobs": [ ... ],
    "summary": "...",
    "created_at": "..."
  }
}
```

### 13.2 字段策略

4-3 优先保证现有接口稳定。建议策略：

| 字段 | 4-3 处理 | 4-4 处理 |
|------|---------|---------|
| generated_by | ✅ 立即使用（llm_agent / mock_agent） | 保持 |
| agent_status | 4-3 内部记录，4-4 统一接口返回 | 统一收口 |
| repair_applied | 4-3 内部记录，4-4 统一接口返回 | 统一收口 |
| fallback_reason | 4-3 内部记录，4-4 统一接口返回 | 统一收口 |
| error_code | 4-3 内部记录，4-4 统一接口返回 | 统一收口 |

### 13.3 禁止行为

- 不能为了 LLM 接入一次性破坏前端已有字段
- 不能把 Pipeline map[string]any 直接返回给前端
- 需要做模型层结构映射
- LLM 结果如果无法完整映射到现有 Model 结构，宁可 fallback 到 Mock

## 14. generated_by / agent_status / repair_applied 规则

### 14.1 统一规则

| 场景 | generated_by | agent_status | repair_applied |
|------|-------------|-------------|----------------|
| LLM 首次成功 | llm_agent | success | false |
| LLM repair 成功 | llm_agent | repaired | true |
| Pipeline fallback_required + Mock fallback 成功 | mock_agent | fallback | 视情况 |
| LLM 结果映射失败 + Mock fallback | mock_agent | fallback | false |

### 14.2 禁止使用的值

```text
❌ generated_by = llm_agent_fallback
❌ generated_by = json_repair_agent
❌ generated_by = repair_agent
❌ generated_by = pipeline
```

## 15. 日志与安全策略

### 15.1 允许记录

- request_id
- agent_name
- prompt_name
- schema_name
- target_output_type
- job_id
- resume_id 或 resume_length
- llm_enabled
- agent_status
- repair_applied
- fallback_required
- fallback_reason
- error_code
- duration_ms
- raw_output_length

### 15.2 禁止记录

- API Key
- Authorization header
- 完整简历
- 完整 JD
- 完整 Prompt
- 完整 raw output
- 用户敏感信息

### 15.3 安全强调

- 真实 Hunyuan smoke 在 4-3.1 执行前仍不得泄露 Key
- 日志中只记录长度、摘要、错误类型
- PromptValues 中不要包含 API Key

## 16. 真实 Hunyuan smoke 前置要求

4-3.1 必须先做真实 LLM Smoke 前置验收。

推荐执行：

```bash
RUN_REAL_LLM_SMOKE=1 make llm-smoke
```

说明：

- 如果真实 smoke 不通过，不建议直接接业务 Agent
- 如果真实 smoke 未执行，至少要保持 LLM_ENABLE=false + Mock fallback 可用
- 真实 smoke 只验证连通性和基本 JSON 输出能力，不验证业务 Prompt 质量
- 业务 Prompt 质量在 4-3.2 / 4-3.4 联调中验证

## 17. 4-3 子任务拆分

| 编号 | 任务 | 说明 |
|------|------|------|
| 4-3.0 | 业务 Agent 接入 LLM 方案 | docs only（本次任务） |
| 4-3.1 | 真实 LLM Smoke 前置验收 | RUN_REAL_LLM_SMOKE=1 make llm-smoke |
| 4-3.2 | MatchAnalysisAgent 接入 Pipeline | Service 层改造 + PromptValues 组装 + Pipeline 调用 + fallback |
| 4-3.3 | 匹配分析接口联调与 fallback 验收 | LLM 成功 / repair / fallback 三路径验收 |
| 4-3.4 | RecommendationAgent 接入 Pipeline | Service 层改造 + PromptValues 组装 + Pipeline 调用 + fallback |
| 4-3.5 | 推荐接口联调与 fallback 验收 | LLM 成功 / repair / fallback 三路径验收 |
| 4-3.6 | ResumeOptimizationAgent 现状核对与接入方案 | 核对代码现状，决定是否预留 |
| 4-3.7 | ResumeOptimizationAgent 接入 Pipeline 或预留 | 根据现状决定 |
| 4-3.8 | 三类业务 Agent 集成验收 | 全路径验收 |
| 4-3.9 | 阶段 4-3 汇总 | 汇总 |

## 18. 验收标准

4-3 完成时应满足：

| 序号 | 验收标准 |
|------|---------|
| 1 | MatchAnalysisAgent 可通过 Pipeline 使用 LLM |
| 2 | RecommendationAgent 可通过 Pipeline 使用 LLM |
| 3 | ResumeOptimizationAgent 根据实际情况完成接入或明确预留 |
| 4 | LLM 成功时返回 generated_by=llm_agent |
| 5 | repair 成功时仍返回 generated_by=llm_agent |
| 6 | fallback_required 时返回 generated_by=mock_agent |
| 7 | 接口稳定不崩 |
| 8 | Prompt / Schema 不被改坏 |
| 9 | 业务 Agent 不直接依赖 HunyuanClient |
| 10 | 真实 API Key 不泄露 |
| 11 | LLM_ENABLE=false 时接口仍正常（Mock 兜底） |
| 12 | 匹配分析接口联调通过 |
| 13 | 推荐接口联调通过 |
| 14 | LLM 输出不能直接作为 map[string]any 返回前端，必须映射到 Model 结构 |

## 19. 当前不做事项

- 本次 4-3.0 不写代码
- 不修改业务 Agent
- 不修改 service
- 不修改 controller
- 不修改 model
- 不修改 Prompt
- 不修改 Schema
- 不调用 Hunyuan
- 不执行真实 smoke
- 不修改前端
- 不进入 4-3.1
- 不进入 4-3.2

## 20. 下一步建议

**下一步进入 4-3.1：真实 LLM Smoke 前置验收。**

4-3.1.B 已执行但仍未通过：本地 .env 仍然不存在，LLM 配置仍不完整（LLM_ENABLE=false，LLM_BASE_URL / LLM_API_KEY / LLM_MODEL 未配置），真实 LLM smoke 仍未执行。暂不建议进入 4-3.2。

两次核验（4-3.1 和 4-3.1.B）均因配置缺失无法执行真实 smoke。

4-3.1.C 已完成开发环境 .env 自动生成与 Makefile 接入：新增 `scripts/dev/ensure_env.sh` 和 `ensure-env` Makefile target，`make up` / `make restart` / `make llm-check` / `make llm-smoke` 执行前自动确保 `.env` 存在。下一步仍需重新执行真实 LLM smoke。

验收报告见 `docs/05-Agent设计/真实LLMSmoke前置验收报告.md`。

需先创建 `.env.local`（或直接编辑 `.env`）补齐本地 LLM 配置后重新执行 `RUN_REAL_LLM_SMOKE=1 make llm-smoke`，通过后再进入 4-3.2。

4-3.1 重新执行结果：三次核验（4-3.1、4-3.1.B 和重新执行）均因 LLM 配置不完整无法执行真实 smoke。LLM_ENABLE=false，LLM_BASE_URL 未配置。暂不建议进入 4-3.2。

4-3.1 通过结果登记：用户已补齐 `.env` 配置并成功执行 `RUN_REAL_LLM_SMOKE=1 make llm-smoke`。真实 LLM Smoke 前置验收已通过。Hunyuan / TokenHub 最小调用可用。`provider=hunyuan`，`model=hy3-preview`，`json_parse_success=true`。下一步为 4-3.2 MatchAnalysisAgent 接入 Pipeline。

4-3.1.D 已完成 LLM Smoke 环境变量加载修复。问题根因是 `llm_smoke.sh` 使用 `source .env`，遇到 `DB_DSN` 中的 `tcp(...)` 时 bash 解析失败，导致 LLM 配置未加载。修复后 `llm_smoke.sh` 复用 `common.sh` 的 `load_env()` 逐行解析 `.env`。真实 LLM smoke 链路已更稳定。下一步仍为 4-3.2 MatchAnalysisAgent 接入 Pipeline。

4-3.2 已完成 MatchAnalysisAgent 接入 Pipeline。核心实现：MatchAnalyzer 接口抽象、PipelineRunner 接口抽象、MatchAnalysisAgent（PromptValues 组装 + Pipeline 调用 + 字段映射 + MatchFallbackError）、MatchService（LLM 优先 + Mock 回退）、config.LLMConfig 扩展到 14 字段、main.go LLM 基础设施初始化。28 项单元测试通过（fake Pipeline，不调用真实 LLM）。下一步为 4-3.3 匹配分析接口联调与 fallback 验收。

4-3.3 已完成：匹配分析接口联调与 fallback 验收。已验证 LLM success / repaired / fallback_required / LLM_ENABLE=false / 配置异常等路径。修复了 LLM 成功路径快照保存缺失的问题。新增 15+ 项 service 层验收测试。下一步为 4-3.4 RecommendationAgent 接入 Pipeline。

4-3.4 已完成：RecommendationAgent 接入 Pipeline。核心实现：Recommender 接口抽象（RecommendationMockAgent + RecommendationAgent 共享）、PipelineRunner 接口复用、RecommendationAgent（PromptValues 组装 + Pipeline 调用 + 候选岗位校验 + 字段映射 + RecFallbackError）、RecommendationService（LLM 优先 + Mock 回退）、RecommendedJob 新增 LLM-enriched 字段（RecommendLabel / OfferGoSummary / QuerySuggestions / GeneratedBy / AgentStatus / RepairApplied / FallbackReason / ErrorCode，均 omitempty）。12 项 Agent 层单元测试通过（fake Pipeline，不调用真实 LLM）。岗位推荐链路已从纯 Mock 升级为 LLM Pipeline 优先 + Mock fallback。推荐结果必须来自候选岗位集，不允许模型编造岗位（job_id 白名单校验）。下一步为 4-3.5 推荐接口联调与 fallback 验收。

4-3.5 已完成：推荐接口联调与 fallback 验收。已验证 LLM success / repaired / fallback_required / LLM_ENABLE=false / 配置异常 / 候选集外岗位 / 空推荐列表等路径。修复了 RecommendationAgent 元数据缺失问题（GeneratedBy/AgentStatus/RepairApplied 未从 RunResult 传递到 RecommendedJob）。新增 19 项 service 层验收测试 + 12 项 agent 层测试。推荐接口在所有场景下均能稳定返回。下一步为 4-3.6 ResumeOptimizationAgent 现状核对与接入方案。

4-3.5.B 已完成：战略差异化与杀手锏确认。项目定位从"岗位推荐工具"升级为"投递前决策与简历初筛优化系统"。三个核心杀手锏：投递边界与优化停止条件、JD 证据链匹配、Agent 输出 Quality Gate。后续 4-3.5.C ~ 4-3.5.H 将围绕投递边界、JD 证据链、Quality Gate、简历上下文设计进行局部战略校准。

4-3.5.C 已完成：投递边界与优化停止条件设计。已定义 apply_readiness（ready_to_apply / quick_fix_before_apply / not_ready_for_this_role / insufficient_information）和 optimization_stop_signal（stop_and_apply / fix_top_3_then_apply / needs_more_evidence / target_role_mismatch）。后续业务 Agent 输出需要避免无限优化，明确 apply_readiness 与 optimization_stop_signal，区分 Top 3 快速优化项与长期提升项。
