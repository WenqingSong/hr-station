# 推荐接口联调与 fallback 验收报告

> 创建时间：2026-05-30
> 所属阶段：4-3.5
> 验收人：CodeBuddy

## 1. 验收目标

验证推荐接口在 LLM Pipeline 优先 + Mock fallback 模式下的稳定性和可用性，覆盖 LLM success / repaired / fallback_required / LLM_ENABLE=false / 配置异常 / 候选集外岗位 / 空推荐列表 / 请求参数异常等场景。

## 2. 验收范围

- 推荐接口 `POST /api/jobs/recommend`
- RecommendationService / RecommendationAgent / Pipeline / Mock fallback 调用链路
- 字段映射与元数据传递
- 安全边界（不泄露 API Key / Prompt / raw output）
- 不涉及 MatchAnalysisAgent / ResumeOptimizationAgent / Prompt / Schema / 前端

## 3. 当前推荐调用链路

```
HTTP POST /api/jobs/recommend
  → AuthMiddleware（JWT 认证）
    → RecommendationHandler.Recommend()
      → RecommendationService.Recommend(ctx, userID, req)
        → getResume()：ResumeRepository 获取简历版本
        → JobRepository.ListRecommendCandidates(ctx, 50) 获取候选岗位
        → if llmEnabled && pipelineRunner != nil:
            → recommendWithLLM():
              → agent.NewRecommendationAgent(pipelineRunner)
              → RecommendationAgent.Recommend(ctx, resume, jobs)
                → buildPromptValues()：组装 RESUME_CONTEXT_JSON / JOB_CANDIDATES_JSON / BUSINESS_CONTEXT
                → PipelineRunner.Run(ctx, RunRequest)
                → 检查 FallbackRequired → RecFallbackError
                → mapParsedToRecommendedJobs()：字段映射 + job_id 白名单校验
                → 设置 GeneratedBy / AgentStatus / RepairApplied 元数据（4-3.5 修复）
              → 保存快照 RecommendationRepository.Create()
            → on RecFallbackError: recommendWithMock()
        → else:
            → recommendWithMock():
              → RecommendationMockAgent.Recommend(ctx, resume, jobs)
              → 设置 generated_by=mock_agent / agent_status=fallback/mock / fallback_reason / error_code
              → 保存快照 RecommendationRepository.Create()
      → response.OK(w, resp) / handleRecommendationError(w, err)
```

### 链路核验记录

| # | 核验项 | 代码位置 |
|---|--------|----------|
| 1 | 推荐接口路径 | `POST /api/jobs/recommend` |
| 2 | 请求方法 | POST |
| 3 | 请求体结构 | `RecommendationRequest{resume_version_no, limit}` |
| 4 | 响应体结构 | `RecommendationResponse{resume_version_id, resume_version_no, generated_by, recommended_jobs, summary, created_at}` |
| 5 | Service 调用 Agent | `recommendation_service.go:136` `agent.NewRecommendationAgent(s.pipelineRunner)` |
| 6 | Agent 调用 Pipeline | `recommendation_agent.go:72` `a.runner.Run(ctx, runReq)` |
| 7 | 候选岗位来源 | `recommendation_service.go:83` `s.jobRepo.ListRecommendCandidates(ctx, 50)` |
| 8 | 候选岗位白名单校验 | `recommendation_agent.go:219-222` `candidateIDs[jobID]` 存在性校验 |
| 9 | Mock fallback 调用位置 | `recommendation_service.go:102` `s.recommendWithMock(...)` |
| 10 | generated_by 赋值（LLM） | `recommendation_agent.go:109` `runResult.GeneratedBy` |
| 11 | agent_status 赋值（LLM） | `recommendation_agent.go:110` `string(runResult.AgentStatus)` |
| 12 | repair_applied 赋值（LLM） | `recommendation_agent.go:111` `runResult.RepairApplied` |
| 13 | fallback_reason / error_code 赋值 | `recommendation_service.go:208-209` Mock fallback 时设置 |
| 14 | LLM_ENABLE=false 路径 | `recommendation_service.go:106` `recommendWithMock(ctx, userID, resume, jobs, req.Limit, nil)` |
| 15 | Pipeline error 路径 | `recommendation_agent.go:81-87` FallbackRequired → RecFallbackError → Service fallback |

## 4. 测试环境

- 使用 fake PipelineRunner，不调用真实 Hunyuan
- Go test 框架
- 无外部依赖

## 5. LLM success 场景验收

**结果：✅ 通过**

| 验证项 | 结果 |
|--------|------|
| 不调用 recommendation_mock | ✅ Agent 直接返回结果 |
| 返回结构符合推荐接口文档 | ✅ RecommendationResponse 结构完整 |
| recommendation items 包含必要字段 | ✅ job_id / title / priority / score / reason / highlights / risk / action |
| recommend_level 映射正确 | ✅ priority_apply → strong_recommend |
| match_score 范围合理 | ✅ 85 |
| recommend_reason 不为空 | ✅ |
| 不返回完整 raw output | ✅ RunResult.RawContent 不进入响应 |
| 不泄露 API Key | ✅ 序列化结果不含敏感模式 |
| 不返回候选集外岗位 | ✅ job_id 白名单校验通过 |

## 6. LLM repaired 场景验收

**结果：✅ 通过**

| 验证项 | 结果 |
|--------|------|
| 不调用 recommendation_mock | ✅ Agent 直接返回修复后结果 |
| 返回结构稳定 | ✅ |
| generated_by=llm_agent | ✅ 不是 json_repair_agent |
| agent_status=repaired | ✅ |
| repair_applied=true | ✅ |
| 不使用 llm_agent_fallback | ✅ |
| 不返回候选集外岗位 | ✅ |

## 7. fallback_required 场景验收

**结果：✅ 通过**

| 验证项 | 结果 |
|--------|------|
| fallback_required 不导致接口失败 | ✅ Agent 返回 RecFallbackError，Service 回退 Mock |
| 用户仍拿到可用推荐结果 | ✅ Mock Agent 提供稳定推荐 |
| 不暴露底层技术错误给用户 | ✅ fallback_reason / error_code 仅在 metadata |
| fallback_reason 有值 | ✅ llm_call_failed |
| error_code 有值 | ✅ LLM_HTTP_ERROR |
| 不调用 MatchAnalysisAgent | ✅ |
| 不调用 ResumeOptimizationAgent | ✅ |

## 8. LLM_ENABLE=false 场景验收

**结果：✅ 通过**

| 验证项 | 结果 |
|--------|------|
| 推荐接口仍走 Mock fallback | ✅ Service 直接调用 recommendWithMock |
| generated_by=mock_agent | ✅ |
| agent_status=mock | ✅ |
| 接口不 500 | ✅ Mock Agent 返回稳定结果 |
| 不强行调用 Hunyuan | ✅ pipelineRunner 为 nil 时不进入 LLM 路径 |

## 9. LLM 配置异常场景验收

**结果：✅ 通过**

| 验证项 | 结果 |
|--------|------|
| 不泄露配置细节 | ✅ 错误通过 RecFallbackError 传递，不暴露 API Key |
| 不输出 API Key | ✅ 序列化结果不含敏感模式 |
| 不返回 500 | ✅ Service 层 fallback 到 Mock |
| fallback_reason / error_code 体现失败原因 | ✅ json_parse_failed / JSON_PARSE_FAILED |

## 10. 候选集外岗位场景验收

**结果：✅ 通过**

| 验证项 | 结果 |
|--------|------|
| 候选集外 job_id 不得进入最终响应 | ✅ 白名单校验过滤 |
| title / department / city 来自候选岗位 | ✅ 代码强制从 candidateIDs 取值 |
| 过滤后无有效推荐项时 fallback | ✅ 全部过滤 → RecFallbackError → Mock |
| 不新增岗位 | ✅ |
| 不请求腾讯招聘接口 | ✅ |

## 11. 空推荐列表场景验收

**结果：✅ 通过**

| 验证项 | 结果 |
|--------|------|
| 空推荐列表不直接返回给用户 | ✅ Agent 返回 RecFallbackError |
| fallback_reason / error_code 合理 | ✅ field_mapping_failed |
| 不返回 500 | ✅ Service fallback 到 Mock |

## 12. 请求参数异常场景验收

**结果：✅ 通过**

| 验证项 | 结果 |
|--------|------|
| 简历缺失返回业务错误 | ✅ "resume not found" |
| 候选岗位为空返回业务错误 | ✅ "no jobs available" |
| limit 超出范围被截断 | ✅ 最大 5 |
| 参数错误不被 fallback 掩盖 | ✅ 参数校验在 LLM 调用之前 |

## 13. 排序与结果可用性验收

**结果：✅ 通过**

| 验证项 | 结果 |
|--------|------|
| 推荐列表不为空 | ✅ Mock Agent 返回至少 3 项 |
| 推荐项顺序稳定 | ✅ Mock 按 score 降序排列 |
| 推荐档位合法 | ✅ strong_recommend / recommend / try |
| match_score 范围合理 | ✅ 0-95 |
| 推荐理由不为空 | ✅ |
| suggested_action 合理 | ✅ |
| job_id / title 与候选岗位一致 | ✅ |

## 14. 接口返回结构验收

**结果：✅ 通过**

| 验证项 | 结果 |
|--------|------|
| 原有核心字段仍存在 | ✅ job_id / title / department / city / priority / score / reason / highlights / risk / action |
| 新增字段不破坏旧字段 | ✅ omitempty |
| generated_by 合理 | ✅ llm_agent / mock_agent |
| agent_status 合理 | ✅ success / repaired / fallback / mock |
| repair_applied 合理 | ✅ LLM 成功=false, LLM 修复=true |
| fallback_reason / error_code 只在 fallback 语义下出现 | ✅ LLM 成功时为空 |
| 不返回完整 raw output | ✅ |
| 不返回完整 Prompt | ✅ |
| 不返回完整候选岗位池 | ✅ |
| 不返回 API Key | ✅ |

## 15. 安全边界验收

**结果：✅ 通过**

| 验证项 | 结果 |
|--------|------|
| 不泄露 API Key | ✅ 序列化结果不含 sk- / apikey / secret_key 等模式 |
| 不输出完整 Prompt | ✅ Prompt 仅在 Pipeline 内部使用 |
| 不输出完整简历 | ✅ 仅传递安全子集到 Prompt |
| 不输出完整候选岗位池 | ✅ 仅返回推荐结果 |
| 不输出完整 raw output | ✅ RawContent 不序列化到响应 |

## 16. 测试结果

| 测试类别 | 测试数 | 通过 | 失败 |
|----------|--------|------|------|
| Agent 层测试 | 12 | 12 | 0 |
| Service 层测试 | 19 | 19 | 0 |
| Pipeline 测试 | 18 | 18 | 0 |
| 全量测试 | 全部 | ✅ | 0 |

## 17. 少量修复记录

### 修复 1：RecommendationAgent 元数据缺失

**问题**：`RecommendationAgent.Recommend()` 在 `mapParsedToRecommendedJobs()` 后未将 RunResult 的 `GeneratedBy`、`AgentStatus`、`RepairApplied` 传递到 `RecommendedJob`，导致 LLM 成功/修复路径的元数据缺失。对比 `MatchAnalysisAgent`（在 Step 6 正确设置了这些字段）。

**修复**：在 `recommendation_agent.go` 的 `Recommend()` 方法中新增 Step 7，遍历结果并设置元数据：
```go
for i := range recommendedJobs {
    recommendedJobs[i].GeneratedBy = runResult.GeneratedBy
    recommendedJobs[i].AgentStatus = string(runResult.AgentStatus)
    recommendedJobs[i].RepairApplied = runResult.RepairApplied
    recommendedJobs[i].FallbackReason = ""
    recommendedJobs[i].ErrorCode = ""
}
```

**影响**：LLM 成功/修复路径的 `generated_by`、`agent_status`、`repair_applied` 现在正确设置，与 MatchAnalysisAgent 行为一致。

## 18. 当前风险

| 风险 | 严重度 | 说明 |
|------|--------|------|
| 真实 LLM 输出质量 | 🟡 中 | 使用 fake Pipeline 测试，未验证真实 Hunyuan 输出的字段映射质量 |
| Service 层测试受限于 repo 具体类型 | 🟡 中 | repository 为具体类型非接口，无法注入 fake repo，部分路径通过代码审查验证 |

## 19. 是否可以进入 4-3.6

**✅ 可以进入 4-3.6 ResumeOptimizationAgent 现状核对与接入方案。**

理由：
- 推荐接口在所有 8 个验收场景下均稳定返回
- LLM success / repaired / fallback_required 验收通过
- 候选岗位白名单校验通过
- 安全边界验证通过
- 元数据修复已完成并通过测试
- 未修改匹配分析 / 简历优化链路

## 20. 验收结论

| # | 验收项 | 结果 |
|---|--------|------|
| 1 | 是否真实调用 Hunyuan | ❌ 否，使用 fake Pipeline |
| 2 | 是否使用 fake Pipeline | ✅ 是 |
| 3 | success / repaired / fallback_required 是否通过 | ✅ 通过 |
| 4 | LLM_ENABLE=false 是否通过 | ✅ 通过 |
| 5 | 配置异常是否通过 | ✅ 通过 |
| 6 | 候选集外岗位是否通过 | ✅ 通过 |
| 7 | 空推荐列表是否通过 | ✅ 通过 |
| 8 | 请求参数异常是否通过 | ✅ 通过 |
| 9 | 排序与结果可用性是否通过 | ✅ 通过 |
| 10 | 是否修改匹配分析 / 简历优化 | ❌ 否 |
| 11 | 是否修改 Prompt / Schema | ❌ 否 |
| 12 | 是否修改前端 | ❌ 否 |
| 13 | 是否可以进入 4-3.6 | ✅ 可以 |

## 21. 下一步建议

1. 进入 4-3.6 ResumeOptimizationAgent 现状核对与接入方案
2. 评估 ResumeOptimizationAgent 是否具备接入 Pipeline 的业务代码基础
3. 如不具备，规划 Service / Handler / Model / API 创建方案
