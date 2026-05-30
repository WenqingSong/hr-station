# 2026-05-30-085-RecommendationAgent接入Pipeline

## 1. 本次任务

阶段 4-3.4：RecommendationAgent 接入 Pipeline。将岗位推荐从纯 Mock Agent 升级为 LLM Pipeline 优先 + Mock fallback。

## 2. 修改范围

- `server/internal/agent/recommendation_agent.go`（新增）
- `server/internal/agent/recommendation_agent_test.go`（新增）
- `server/internal/agent/recommendation_mock.go`（修改：Recommend 方法签名新增 ctx 参数）
- `server/internal/model/recommendation.go`（修改：新增 LLM-enriched 字段）
- `server/internal/service/recommendation_service.go`（修改：注入 pipelineRunner + llmEnabled）
- `server/cmd/server/main.go`（修改：RecommendationService 移到 LLM 初始化后创建）
- docs 更新 13 份文档

## 3. 当前推荐链路

```
POST /api/jobs/recommend
→ RecommendationHandler.Recommend
→ RecommendationService.Recommend
→ recommendWithLLM (llmEnabled=true)
  → RecommendationAgent.Recommend(ctx, resume, jobs)
  → pipelineRunner.Run(ctx, RunRequest)
  → success → mapParsedToRecommendedJobs → resp(generated_by=llm_agent)
  → repaired → mapParsedToRecommendedJobs → resp(generated_by=llm_agent, repair_applied=true)
  → fallback_required → RecFallbackError → recommendWithMock
→ recommendWithMock (llmEnabled=false 或 fallback)
  → mockAgent.Recommend(ctx, resume, jobs)
  → resp(generated_by=mock_agent, fallback_reason=xxx)
```

## 4. 接入方案

与 MatchAnalysisAgent 同构：

1. 定义 Recommender 接口（`Recommend(ctx, resume, jobs)`）
2. RecommendationAgent 实现 Recommender，内部调用 PipelineRunner
3. RecommendationMockAgent 实现 Recommender（新增 ctx 参数）
4. RecommendationService 持有 pipelineRunner + llmEnabled + mockAgent
5. LLM 优先，fallback 时调用 mockAgent

## 5. PromptValues 组装

三个占位符，与 recommendation_v1.md 对应：

- **RESUME_CONTEXT_JSON**：简历安全子集（version_no, target_role, target_city, skills, education, experience, project, self_evaluation）
- **JOB_CANDIDATES_JSON**：候选岗位列表（job_id, title, department, city, category, description, requirements, skill_keywords, external_id, data_quality）
- **BUSINESS_CONTEXT**：业务规则约束（只能基于候选岗位推荐，不要编造岗位）

## 6. 候选岗位校验

推荐链路的核心风险是模型编造岗位。实现以下校验：

1. LLM 返回的每个推荐项必须有 job_id
2. job_id 必须在候选集中（candidateIDs map 白名单过滤）
3. 无效 job_id 的推荐项被跳过
4. 全部过滤后（0 个有效推荐）触发 fallback
5. 推荐项的 title/department/city 等字段来自数据库候选岗位，不采用 LLM 输出
6. 测试覆盖"LLM 返回不存在 job_id"场景

## 7. Pipeline 调用

```go
runReq := pipeline.RunRequest{
    AgentName:        "RecommendationAgent",
    PromptName:       prompt.PromptRecommendation,
    SchemaName:       schema.SchemaRecommendation,
    TargetOutputType: pipeline.TargetRecommendation,
    PromptValues:     promptValues,
    Metadata:         map[string]string{...},
}
runResult := a.runner.Run(ctx, runReq)
```

## 8. LLM 成功 / repaired / fallback_required 处理

- **success**：mapParsedToRecommendedJobs → generated_by=llm_agent
- **repaired**：mapParsedToRecommendedJobs → generated_by=llm_agent, repair_applied=true
- **fallback_required**：返回 RecFallbackError → Service 调用 recommendWithMock

## 9. Mock fallback 集成

- RecFallbackError 结构体（FallbackReason / ErrorCode / ErrorMessage）
- Service 层捕获 RecFallbackError 后调用 recommendWithMock
- Mock 路径设置 AgentStatus=fallback / FallbackReason / ErrorCode
- LLM_ENABLE=false 时直接走 Mock，不初始化 Pipeline

## 10. 字段映射

LLM Schema → RecommendedJob 映射：

| Schema 字段 | Model 字段 | 映射规则 |
|---|---|---|
| recommend_level | Priority | priority_apply→strong_recommend, recommended→recommend, explore_first→try |
| match_score | RecommendationScore | 直接映射 |
| recommend_reason | RecommendationReason | 直接映射 |
| evidence | MatchHighlights | 提取 explanation 或 resume_evidence |
| risk_points | RiskPoints | 提取 risk 字符串 |
| next_action | SuggestedAction | 提取 action 字符串 |
| recommend_label | RecommendLabel | 直接映射（omitempty） |
| offer_go_summary | OfferGoSummary | 直接映射（omitempty） |
| query_suggestions | QuerySuggestions | 提取 query 字符串（omitempty） |

Job 字段（title/department/city/ExternalSource/ExternalID/DataQuality/SourceURL/IsFeatured）来自候选集，不采用 LLM 输出。

## 11. 测试结果

12 项 Agent 层测试全部通过：

- TestRecommendationAgent_LLM_Success
- TestRecommendationAgent_LLM_Repaired
- TestRecommendationAgent_FallbackRequired
- TestRecommendationAgent_InvalidJobID_Filtered
- TestRecommendationAgent_EmptyRecommendations
- TestRecommendationAgent_ParsedNil
- TestRecommendationAgent_MixedValidInvalidJobIDs
- TestRecommendationAgent_PromptValuesContainKeys
- TestRecommendationAgent_NoSensitiveDataInResults
- TestRecommendationAgent_RecommendLevelMapping
- TestRecommendationAgent_FieldsFromCandidate
- TestRecommendationAgent_MissingJobID_Skipped

全量测试 `go test ./... -count=1` 全部通过。

## 12. 文档更新

更新 13 份文档：
- 阶段4-3业务Agent接入LLM方案.md
- 岗位推荐Agent.md
- Agent总览.md
- LLMClient设计.md
- Hunyuan接入准备.md
- Prompt模板.md
- JSON输出规范.md
- 失败兜底与Mock回退策略.md
- 推荐接口.md
- 项目状态.md
- 下一步任务.md
- 变更记录.md
- 说明书.md

## 13. 未修改内容

- 未修改 .env / .env.local / .env.example
- 未写入真实 API Key
- 未修改 Prompt 文件
- 未修改 Schema 文件
- 未修改 MatchAnalysisAgent（match_agent.go / match_mock.go / match_service.go）
- 未修改 ResumeOptimizationAgent
- 未修改前端
- 未请求腾讯招聘接口
- 未执行 jobs-fetch / jobs-import
- 本次只接入 RecommendationAgent

## 14. 当前风险

- LLM 编造岗位风险：已通过 job_id 白名单校验缓解，但更复杂的语义级校验留待 4-3.5
- 候选集过大导致 Prompt 过长：当前限制 50 条，后续可优化截断策略
- 推荐级别映射覆盖度：priority_apply / recommended / explore_first 已覆盖，not_recommended 映射为 try

## 15. 是否可以进入 4-3.5

**✅ 可以进入 4-3.5 推荐接口联调与 fallback 验收。**

理由：
1. RecommendationAgent 已接入 Pipeline，12 项 Agent 测试通过
2. 候选岗位校验已实现（job_id 白名单过滤）
3. Mock fallback 已集成
4. 全量测试通过
5. 字段映射和推荐级别映射已验证

## 16. 下一步建议

进入 4-3.5：推荐接口联调与 fallback 验收。验收场景建议包括：

1. LLM success：返回 llm_agent 推荐
2. LLM repaired：返回 llm_agent 推荐 + repair_applied=true
3. LLM fallback_required：回退 Mock 推荐
4. LLM_ENABLE=false：直接 Mock 推荐
5. LLM 配置缺失：Mock fallback
6. LLM 返回不存在 job_id：过滤后 Mock fallback
7. 安全边界验证：无 API Key 泄露
8. generated_by 字段验证
