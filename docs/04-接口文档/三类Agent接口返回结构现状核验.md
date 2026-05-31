# 三类 Agent 接口返回结构现状核验

> 创建时间：2026-05-31
> 所属阶段：4-4.0
> 核验人：CodeBuddy
> 核验范围：MatchAnalysis、Recommendation、ResumeOptimization 三类 Agent 接口的真实返回结构、meta 字段位置、fallback / repair 状态字段、接口文档与 Go Model 一致性

## 1. 背景

阶段 4-3 已完成三类业务 Agent 接入 LLM Pipeline，三类 Agent 均具备 LLM + fallback 基础能力。前端产品化阶段即将开始，但在进入前端实现前，必须先收口后端接口返回结构。

前端不能面对三类接口三套不一致返回结构。

本阶段只做现状核验，不做改造。

## 2. 核验范围

| 核验对象 | 范围 |
|---------|------|
| MatchAnalysis 接口 | `POST /api/jobs/{id}/match` |
| Recommendation 接口 | `POST /api/jobs/recommend` |
| ResumeOptimization 接口 | `POST /api/resumes/optimize` |
| Go Model | match.go / recommendation.go / resume_optimization.go |
| Handler | match_handler.go / recommendation_handler.go / resume_optimization_handler.go |
| Service | match_service.go / recommendation_service.go / resume_optimization_service.go |
| Agent | match_agent.go / recommendation_agent.go / resume_optimization_agent.go |
| Pipeline | pipeline/types.go |
| Schema | match_analysis_schema.json / recommendation_schema.json / resume_optimization_schema.json |
| 接口文档 | 匹配分析接口.md / 推荐接口.md / 简历优化接口.md |

## 3. 当前三类接口总览

### 3.1 统一 HTTP Response Envelope

三类接口均使用 `response.OK(w, resp)` 返回，底层封装为：

```json
{
  "code": 0,
  "message": "ok",
  "data": { ... }
}
```

**结论：HTTP 层有统一 envelope（`{code, message, data}`），但 `data` 内部结构三类不一致。**

### 3.2 data 内部结构对比

| 项目 | MatchAnalysis | Recommendation | ResumeOptimization |
|------|--------------|----------------|--------------------|
| 业务结果位置 | `data.result` | `data.recommended_jobs` | `data.result` |
| generated_by 外层 | `data.generated_by` | `data.generated_by` | `data.generated_by` |
| ID 字段 | `data.analysis_id` | `data.resume_version_id` | `data.optimization_id` |
| 辅助字段 | `data.analysis_basis` | `data.summary` | 无 |
| 时间字段 | `data.created_at` | `data.created_at` | `data.created_at` |

**关键差异**：
- Match 和 ResumeOptimization 的业务结果在 `data.result` 中
- Recommendation 的业务结果在 `data.recommended_jobs` 数组中
- Recommendation 的 meta 字段（generated_by 等）在每个 `recommended_jobs[i]` 内部

## 4. MatchAnalysis 接口返回结构现状

### 4.1 Handler 返回路径

```
Handler → Service.Analyze → response.OK(w, MatchResponse)
```

### 4.2 MatchResponse 结构

```go
type MatchResponse struct {
    AnalysisID    uint64                     `json:"analysis_id"`
    GeneratedBy   string                     `json:"generated_by"`
    Result        *model.MatchAnalysisResult `json:"result"`
    AnalysisBasis *model.AnalysisBasis       `json:"analysis_basis"`
    CreatedAt     string                     `json:"created_at"`
}
```

### 4.3 MatchAnalysisResult 关键字段

| 字段 | JSON tag | 类型 | omitempty | 位置 |
|------|----------|------|-----------|------|
| generated_by | generated_by | string | ✅ omitempty | data.result 内 |
| agent_status | agent_status | string | ✅ omitempty | data.result 内 |
| fallback_reason | fallback_reason | string | ✅ omitempty | data.result 内 |
| error_code | error_code | string | ✅ omitempty | data.result 内 |
| repair_applied | repair_applied | bool | ✅ omitempty | data.result 内 |
| apply_readiness | apply_readiness | string | ✅ omitempty | data.result 内 |
| optimization_stop_signal | optimization_stop_signal | string | ✅ omitempty | data.result 内 |
| top_fix_items | top_fix_items | []TopFixItem | ✅ omitempty | data.result 内 |

**注意**：`generated_by` 同时存在于 `data.generated_by`（Service 层设置）和 `data.result.generated_by`（Agent 层设置），存在重复。

### 4.4 完整返回结构示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "analysis_id": 12,
    "generated_by": "mock_agent",
    "result": {
      "job_id": 1,
      "job_title": "Go 后端开发工程师",
      "match_score": 78,
      "screening_result": "likely",
      "evidence_chain": [...],
      "match_highlights": [...],
      "risk_points": [...],
      "optimization_suggestions": [...],
      "next_steps": [...],
      "authenticity_notice": "...",
      "match_level": "",
      "match_label": "",
      "offer_go_summary": "",
      "gap_points": [],
      "interview_questions": [],
      "apply_readiness": "quick_fix_before_apply",
      "optimization_stop_signal": "fix_top_3_then_apply",
      "top_fix_items": [...],
      "generated_by": "mock_agent",
      "agent_status": "mock",
      "repair_applied": false,
      "fallback_reason": "",
      "error_code": ""
    },
    "analysis_basis": {...},
    "created_at": "..."
  }
}
```

## 5. Recommendation 接口返回结构现状

### 5.1 Handler 返回路径

```
Handler → Service.Recommend → response.OK(w, RecommendationResponse)
```

### 5.2 RecommendationResponse 结构

```go
type RecommendationResponse struct {
    ResumeVersionID uint64                  `json:"resume_version_id"`
    ResumeVersionNo int                     `json:"resume_version_no"`
    GeneratedBy     string                  `json:"generated_by"`
    RecommendedJobs []*model.RecommendedJob `json:"recommended_jobs"`
    Summary         string                  `json:"summary"`
    CreatedAt       string                  `json:"created_at"`
}
```

### 5.3 RecommendedJob 关键字段

| 字段 | JSON tag | 类型 | omitempty | 位置 |
|------|----------|------|-----------|------|
| generated_by | generated_by | string | ✅ omitempty | data.recommended_jobs[i] 内 |
| agent_status | agent_status | string | ✅ omitempty | data.recommended_jobs[i] 内 |
| fallback_reason | fallback_reason | string | ✅ omitempty | data.recommended_jobs[i] 内 |
| error_code | error_code | string | ✅ omitempty | data.recommended_jobs[i] 内 |
| repair_applied | repair_applied | bool | ✅ omitempty | data.recommended_jobs[i] 内 |
| apply_readiness | apply_readiness | string | ✅ omitempty | data.recommended_jobs[i] 内 |
| optimization_stop_signal | optimization_stop_signal | string | ✅ omitempty | data.recommended_jobs[i] 内 |
| top_fix_items | top_fix_items | []TopFixItem | ✅ omitempty | data.recommended_jobs[i] 内 |

**注意**：`generated_by` 同时存在于 `data.generated_by` 和每个 `data.recommended_jobs[i].generated_by`，存在重复且粒度不同。

### 5.4 完整返回结构示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "resume_version_id": 2,
    "resume_version_no": 2,
    "generated_by": "mock_agent",
    "recommended_jobs": [
      {
        "job_id": 1,
        "title": "Go 后端开发工程师",
        "priority": "strong_recommend",
        "recommendation_score": 95,
        "recommendation_reason": "...",
        "match_highlights": [...],
        "risk_points": [...],
        "suggested_action": "...",
        "recommend_label": "",
        "offer_go_summary": "",
        "query_suggestions": [],
        "apply_readiness": "quick_fix_before_apply",
        "optimization_stop_signal": "fix_top_3_then_apply",
        "top_fix_items": [...],
        "generated_by": "mock_agent",
        "agent_status": "mock",
        "repair_applied": false,
        "fallback_reason": "",
        "error_code": ""
      }
    ],
    "summary": "...",
    "created_at": "..."
  }
}
```

## 6. ResumeOptimization 接口返回结构现状

### 6.1 Handler 返回路径

```
Handler → Service.Optimize → response.OK(w, ResumeOptimizationResponse)
```

### 6.2 ResumeOptimizationResponse 结构

```go
type ResumeOptimizationResponse struct {
    OptimizationID uint64                          `json:"optimization_id"`
    GeneratedBy    string                          `json:"generated_by"`
    Result         *model.ResumeOptimizationResult `json:"result"`
    CreatedAt      string                          `json:"created_at"`
}
```

### 6.3 ResumeOptimizationResult 关键字段

| 字段 | JSON tag | 类型 | omitempty | 位置 |
|------|----------|------|-----------|------|
| generated_by | generated_by | string | ❌ 必填 | data.result 内 |
| agent_status | agent_status | string | ❌ 必填 | data.result 内 |
| fallback_required | fallback_required | bool | ❌ 必填 | data.result 内 |
| repair_applied | repair_applied | bool | ❌ 必填 | data.result 内 |
| fallback_reason | fallback_reason | string | ✅ omitempty | data.result 内 |
| error_code | error_code | string | ✅ omitempty | data.result 内 |
| apply_readiness | apply_readiness | string | ❌ 必填 | data.result 内 |
| optimization_stop_signal | optimization_stop_signal | string | ❌ 必填 | data.result 内 |
| top_fix_items | top_fix_items | []ResumeOptimizationFixItem | ❌ 必填 | data.result 内 |
| summary | summary | string | ❌ 必填 | data.result 内 |
| long_term_improvement | long_term_improvement | []ResumeOptimizationLongTermItem | ❌ 必填 | data.result 内 |
| jd_evidence_chain | jd_evidence_chain | []ResumeOptimizationJDEvidenceItem | ❌ 必填 | data.result 内 |
| verification_notice | verification_notice | string | ❌ 必填 | data.result 内 |

**关键差异**：ResumeOptimization 的 meta 字段（generated_by / agent_status / fallback_required / repair_applied）是**必填字段**（无 omitempty），而 Match/Recommendation 是 omitempty。

### 6.4 完整返回结构示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "optimization_id": 0,
    "generated_by": "mock_agent",
    "result": {
      "agent_status": "mock_success",
      "generated_by": "mock_agent",
      "fallback_required": false,
      "repair_applied": false,
      "apply_readiness": "quick_fix_before_apply",
      "optimization_stop_signal": "fix_top_3_then_apply",
      "summary": "...",
      "top_fix_items": [...],
      "long_term_improvement": [...],
      "jd_evidence_chain": [...],
      "verification_notice": "..."
    },
    "created_at": "..."
  }
}
```

## 7. 三类接口返回结构矩阵

| 项目 | MatchAnalysis | Recommendation | ResumeOptimization | 结论 |
|------|--------------|----------------|--------------------|------|
| 是否有统一 envelope | ✅ 是（{code,message,data}） | ✅ 是 | ✅ 是 | 一致 |
| 业务结果位置 | data.result | data.recommended_jobs | data.result | **不一致** |
| meta 位置 | data.result 内 + data.generated_by | data.recommended_jobs[i] 内 + data.generated_by | data.result 内 + data.generated_by | **不一致** |
| generated_by | data.generated_by + data.result 内（omitempty） | data.generated_by + data.recommended_jobs[i] 内（omitempty） | data.generated_by + data.result 内（必填） | **不一致**：重复 + 必填/选填不同 |
| agent_status | data.result 内（omitempty） | data.recommended_jobs[i] 内（omitempty） | data.result 内（必填） | **不一致**：位置 + 必填/选填不同 |
| fallback_reason | data.result 内（omitempty） | data.recommended_jobs[i] 内（omitempty） | data.result 内（omitempty） | **不一致**：位置不同 |
| error_code | data.result 内（omitempty） | data.recommended_jobs[i] 内（omitempty） | data.result 内（omitempty） | **不一致**：位置不同 |
| repair_applied | data.result 内（omitempty） | data.recommended_jobs[i] 内（omitempty） | data.result 内（必填） | **不一致**：位置 + 必填/选填不同 |
| schema_valid | ❌ 不存在 | ❌ 不存在 | ❌ 不存在 | 一致（均缺失） |
| request_id | ❌ 不存在 | ❌ 不存在 | ❌ 不存在 | 一致（均缺失） |
| fallback_required | ❌ 不存在（用 MatchFallbackError） | ❌ 不存在（用 RecFallbackError） | ✅ data.result 内（必填） | **不一致** |

## 8. 返回路径状态矩阵

| 路径 | 当前是否可识别 | 识别字段 | 前端展示风险 |
|------|--------------|---------|------------|
| LLM success | ✅ 可识别 | generated_by="llm_agent" + agent_status="success" | 🟢 低 |
| JSON repair success | ✅ 可识别 | generated_by="llm_agent" + agent_status="repaired" + repair_applied=true | 🟡 中：repair_applied 在 Match/Rec 中是 omitempty，Mock 时不会返回 |
| LLM error fallback | ✅ 可识别 | generated_by="mock_agent" + agent_status="fallback" + fallback_reason + error_code | 🟡 中：fallback_reason/error_code 位置不一致 |
| Mock direct | ⚠️ 部分可识别 | Match/Rec: agent_status="mock"，ResumeOpt: agent_status="mock_success" | 🔴 高：agent_status 值不一致 |
| handler error | ✅ 可识别 | code≠0 + message | 🟢 低 |

### 8.1 agent_status 值不一致详细分析

| 场景 | MatchAnalysis | Recommendation | ResumeOptimization |
|------|--------------|----------------|--------------------|
| LLM success | "success" | "success" | "success" |
| LLM repaired | "repaired" | "repaired" | "repaired" |
| LLM fallback | "fallback" | "fallback" | "fallback" |
| Mock direct (llmEnabled=false) | **"mock"** | **"mock"** | **"mock_success"** |

**风险**：前端如果统一检查 `agent_status === "mock_success"`，则 Match/Recommendation 的 Mock direct 会漏识别。如果检查 `agent_status === "mock"`，则 ResumeOptimization 会漏识别。

### 8.2 generated_by 双重设置问题

三类接口都存在 `generated_by` 在两个位置设置：

1. **Service 层**设置 `data.generated_by`（MatchResponse / RecommendationResponse / ResumeOptimizationResponse 的 GeneratedBy 字段）
2. **Agent 层**或 **Mock fallback** 设置 `result.generated_by` / `recommended_jobs[i].generated_by`

以 Match 为例：
- LLM 成功：`data.generated_by = result.GeneratedBy = "llm_agent"`
- Mock fallback：`data.generated_by = "mock_agent"`, `result.GeneratedBy = "mock_agent"`（Service 层强制覆盖）
- Mock direct：`data.generated_by = "mock_agent"`, `result.AgentStatus = "mock"` 但 `result.GeneratedBy` 取决于 Mock Agent 初始值

**风险**：前端不知道该读 `data.generated_by` 还是 `data.result.generated_by`。

## 9. 文档与实现一致性矩阵

### 9.1 匹配分析接口文档 vs 实现

| 对象 | 接口文档 | Go Model | Handler | Service | Agent | Schema | 一致性 |
|------|---------|---------|---------|---------|-------|--------|--------|
| response envelope | 描述了 code/message/data | — | response.OK() | — | — | — | ✅ pass |
| generated_by | 文档示例有 data.generated_by | MatchResponse.GeneratedBy + MatchAnalysisResult.GeneratedBy | — | 设置两层 | 设置 result 层 | Schema 中有 | ⚠️ 部分：两层重复 |
| agent_status | 文档后续扩展提到 | MatchAnalysisResult.AgentStatus (omitempty) | — | Mock fallback 设置 | LLM Agent 设置 | Schema 中无 | ⚠️ 部分：文档示例未展示 |
| fallback_reason | 文档后续扩展提到 | MatchAnalysisResult.FallbackReason (omitempty) | — | Mock fallback 设置 | — | Schema 中无 | ⚠️ 部分：文档示例未展示 |
| repair_applied | 文档后续扩展提到 | MatchAnalysisResult.RepairApplied (omitempty) | — | — | LLM Agent 设置 | Schema 中无 | ⚠️ 部分：文档示例未展示 |
| schema_valid | 文档未提及 | ❌ 不存在 | — | — | — | Schema 中无 | ✅ pass（均不存在） |
| request_id | 文档未提及 | ❌ 不存在 | — | — | — | Schema 中无 | ✅ pass（均不存在） |

### 9.2 推荐接口文档 vs 实现

| 对象 | 接口文档 | Go Model | Handler | Service | Agent | Schema | 一致性 |
|------|---------|---------|---------|---------|-------|--------|--------|
| response envelope | 描述了 code/message/data | — | response.OK() | — | — | — | ✅ pass |
| generated_by | 文档示例有 data.generated_by | RecommendationResponse.GeneratedBy + RecommendedJob.GeneratedBy | — | 设置两层 | 设置 job 层 | Schema 中有 | ⚠️ 部分：两层重复 |
| agent_status | 文档后续扩展提到 | RecommendedJob.AgentStatus (omitempty) | — | Mock fallback 设置 | LLM Agent 设置 | Schema 中无 | ⚠️ 部分：文档示例未展示 |
| recommend_level vs priority | Schema 用 recommend_level | Go Model 用 priority | — | Service 映射 | Agent 映射 | Schema 用 recommend_level | ⚠️ 部分：字段名不同 |

### 9.3 简历优化接口文档 vs 实现

| 对象 | 接口文档 | Go Model | Handler | Service | Agent | Schema | 一致性 |
|------|---------|---------|---------|---------|-------|--------|--------|
| response envelope | 描述了 code/message/data | — | response.OK() | — | — | — | ✅ pass |
| generated_by | 文档示例有 data.generated_by + data.result.generated_by | 两层都有 | — | 设置两层 | 设置 result 层 | Schema 中有 | ✅ pass（但重复） |
| agent_status | 文档示例有 data.result.agent_status | ResumeOptimizationResult.AgentStatus（必填） | — | Mock 设置 | LLM Agent 设置 | Schema 中无 | ✅ pass |
| fallback_required | 文档示例有 data.result.fallback_required | ResumeOptimizationResult.FallbackRequired（必填） | — | Mock 设置 | — | Schema 中无 | ✅ pass（但仅此 Agent 有） |
| apply_readiness | 文档示例有 | ResumeOptimizationResult.ApplyReadiness（必填） | — | Mock 设置 | LLM Agent 设置 | Schema 中有 | ✅ pass |
| schema_valid | 文档未提及 | ❌ 不存在 | — | — | — | Schema 中无 | ✅ pass |

### 9.4 综合结论

| Agent | 一致性 | 说明 |
|-------|--------|------|
| MatchAnalysis | ⚠️ 部分 | meta 字段 omitempty，文档未展示 LLM 后新增字段的真实返回示例 |
| Recommendation | ⚠️ 部分 | meta 字段位置在 recommended_jobs[i] 内，与其他两个 Agent 不同；priority vs recommend_level 映射 |
| ResumeOptimization | ✅ 基本一致 | meta 字段必填，文档示例准确；但与 Match/Rec 的 meta 字段风格不同 |

## 10. 当前字段缺口

| 字段 | MatchAnalysis | Recommendation | ResumeOptimization | 影响 |
|------|--------------|----------------|--------------------|------|
| schema_valid | ❌ 缺失 | ❌ 缺失 | ❌ 缺失 | 前端无法判断 LLM 输出是否通过了 Schema 校验 |
| request_id | ❌ 缺失 | ❌ 缺失 | ❌ 缺失 | 前端无法追踪请求链路 |
| fallback_required | ❌ 缺失（用错误传递） | ❌ 缺失（用错误传递） | ✅ 存在（必填） | Match/Rec 的 fallback 状态靠 agent_status 间接判断 |
| agent_status 统一 | "mock" | "mock" | "mock_success" | Mock direct 时 agent_status 值不一致 |

## 11. 前端消费风险

### 11.1 🔴 高风险

1. **agent_status Mock direct 值不一致**：前端需要同时判断 `"mock"` 和 `"mock_success"` 才能识别 Mock direct
2. **meta 字段位置不一致**：Recommendation 的 meta 字段在 `recommended_jobs[i]` 内，Match/ResumeOpt 的在 `result` 内，前端需要三套读取逻辑
3. **generated_by 双重设置**：前端不知道读 `data.generated_by` 还是 `data.result/generated_by`

### 11.2 🟡 中风险

4. **fallback_required 仅 ResumeOptimization 有**：Match/Recommendation 的 fallback 状态需要从 `agent_status="fallback"` 间接判断
5. **omitempty 导致 Mock 时字段缺失**：Match/Rec 的 repair_applied / fallback_reason / error_code 在 Mock direct 时不返回
6. **接口文档示例未更新**：匹配分析和推荐接口文档的 JSON 示例仍展示 Mock 时代的结构，未展示 LLM 后新增的 meta 字段

### 11.3 🟢 低风险

7. **schema_valid 不存在**：前端无法判断输出质量，但可通过 agent_status 间接判断
8. **request_id 不存在**：不影响前端展示，仅影响调试

## 12. 4-4.1 契约设计输入

基于本次核验，4-4.1 统一接口返回契约设计应重点解决：

### 12.1 必须解决

1. **统一 meta 位置**：三类 Agent 的 meta 字段应放在统一位置
   - 建议方案 A：提取到 `data.meta` 统一层
   - 建议方案 B：保持在业务 result 内，但三类对齐字段名和必填性

2. **统一 agent_status 枚举值**：Mock direct 应统一为 `"mock_success"`
   - Match/Rec 当前是 `"mock"`，需改为 `"mock_success"`

3. **消除 generated_by 重复**：选择保留一层
   - 建议保留 `data.generated_by`（外层），删除 result 内的 generated_by
   - 或保留 result 内的，删除外层的

4. **统一 fallback_required**：三类 Agent 都应返回此字段
   - Match/Rec 当前缺失，需补充

### 12.2 建议解决

5. **meta 字段全部非 omitempty**：让前端稳定消费，不需要判断字段是否存在
6. **接口文档 JSON 示例更新**：展示 LLM success / fallback / Mock direct 三种场景的真实返回
7. **新增 schema_valid 字段**：让前端知道输出是否通过了 Schema 校验

### 12.3 可选解决

8. **新增 request_id 字段**：便于前端追踪请求
9. **Recommendation 结构调整**：将 meta 字段从 `recommended_jobs[i]` 提取到外层

## 13. 当前限制

1. 本阶段为 read-only + docs，未修改任何代码
2. 核验基于代码静态分析，未调用真实 LLM 验证实际返回
3. 接口文档示例可能滞后于代码实现
4. Schema 中有但 Go Model 未实现的字段不在本次核验范围

## 14. 是否可以进入 4-4.1

**✅ 可以进入。**

理由：
1. 已查清三类接口当前返回结构
2. 已明确 meta 字段位置和枚举值不一致问题
3. 已明确 fallback / repair 状态字段差异
4. 已明确接口文档与 Go Model 一致性
5. 已形成 4-4.1 契约设计输入

## 15. 下一步建议

进入 **4-4.1 统一接口返回契约设计冻结**，重点解决：

1. 统一 meta 字段位置（建议提取到 `data.meta`）
2. 统一 agent_status 枚举值（Mock direct 统一为 `"mock_success"`）
3. 消除 generated_by 重复
4. 统一 fallback_required 字段
5. 更新接口文档 JSON 示例
6. 冻结契约后，再进入前端实现
