# 三类 Agent LLM / fallback 本地集成测试补齐

## 1. 背景

阶段 4-3.8.1：三类 Agent LLM / fallback 本地集成测试补齐。

前置条件：
- MatchAnalysisAgent：LLM + fallback 已接入（4-3.8.1-R 之前已完成）
- RecommendationAgent：LLM + fallback 已接入（4-3.8.1-R 之前已完成）
- ResumeOptimizationAgent：LLM + fallback 已接入（4-3.8.1-R 完成）

## 2. 测试范围

三类 Agent 的统一本地集成测试，覆盖：
1. LLM success 行为
2. LLM error fallback 行为
3. Malformed JSON / JSON repair / fallback 行为
4. Mock direct 行为
5. Mock fallback 行为
6. generated_by / agent_status 标识
7. apply_readiness / optimization_stop_signal / top_fix_items 战略字段

## 3. Fake Pipeline / Fake Runner 设计

使用 `integrationRunner` 结构体，实现 `PipelineRunner` 接口：
- 可配置返回 success / repaired / fallback_required 三种状态
- 不依赖真实 LLM 调用
- 不依赖数据库或外部服务

测试数据：
- `intTestResume()`：标准测试简历
- `intTestJob()`：标准测试岗位
- `intTestJobs()`：标准测试岗位列表
- `intTestLLMMatchOutput()`：完整 LLM Match 输出（含战略字段）
- `intTestLLMRecOutput()`：完整 LLM Recommendation 输出（含战略字段）
- `intTestLLMResumeOptOutput()`：完整 LLM ResumeOptimization 输出（含战略字段）
- `intTestOptimizeInput()`：标准 OptimizeInput

## 4. MatchAnalysisAgent 测试结果

| 测试场景 | 测试函数 | 结果 |
|----------|----------|------|
| LLM success | TestIntegration_MatchAnalysis_LLM_Success | ✅ pass |
| LLM error fallback | TestIntegration_MatchAnalysis_LLM_Error_Fallback | ✅ pass |
| Malformed JSON → repair success | TestIntegration_MatchAnalysis_MalformedJSON_Repaired | ✅ pass |
| Malformed JSON → repair failed → fallback | TestIntegration_MatchAnalysis_MalformedJSON_RepairFailed_Fallback | ✅ pass |
| Mock direct | TestIntegration_MatchAnalysis_Mock_Direct | ✅ pass |
| Mock fallback metadata | TestIntegration_MatchAnalysis_Mock_Fallback_Metadata | ✅ pass |

## 5. RecommendationAgent 测试结果

| 测试场景 | 测试函数 | 结果 |
|----------|----------|------|
| LLM success | TestIntegration_Recommendation_LLM_Success | ✅ pass |
| LLM error fallback | TestIntegration_Recommendation_LLM_Error_Fallback | ✅ pass |
| Malformed JSON → repair success | TestIntegration_Recommendation_MalformedJSON_Repaired | ✅ pass |
| Malformed JSON → repair failed → fallback | TestIntegration_Recommendation_MalformedJSON_RepairFailed_Fallback | ✅ pass |
| Mock direct | TestIntegration_Recommendation_Mock_Direct | ✅ pass |
| Mock fallback metadata | TestIntegration_Recommendation_Mock_Fallback_Metadata | ✅ pass |

## 6. ResumeOptimizationAgent 测试结果

| 测试场景 | 测试函数 | 结果 |
|----------|----------|------|
| LLM success | TestIntegration_ResumeOptimization_LLM_Success | ✅ pass |
| LLM error fallback | TestIntegration_ResumeOptimization_LLM_Error_Fallback | ✅ pass |
| Malformed JSON → repair success | TestIntegration_ResumeOptimization_MalformedJSON_Repaired | ✅ pass |
| Malformed JSON → repair failed → fallback | TestIntegration_ResumeOptimization_MalformedJSON_RepairFailed_Fallback | ✅ pass |
| Mock direct | TestIntegration_ResumeOptimization_Mock_Direct | ✅ pass |
| Mock fallback metadata | TestIntegration_ResumeOptimization_Mock_Fallback_Metadata | ✅ pass |

## 7. generated_by / agent_status 标准

| 场景 | generated_by | agent_status |
|------|-------------|-------------|
| LLM success | llm_agent | success |
| LLM repaired | llm_agent | repaired |
| LLM error → fallback required | mock_agent（Pipeline 层） | fallback_required（Pipeline 层） |
| Mock fallback（Service 层） | mock_agent | fallback |
| Mock direct（llmEnabled=false） | mock_agent | mock_success |

## 8. LLM / fallback 测试矩阵

| Agent              | LLM success | malformed JSON repair | repair failed fallback | LLM error fallback | Mock direct | generated_by / agent_status |
| ------------------ | ----------- | -------------------- | ---------------------- | ------------------ | ----------- | --------------------------- |
| MatchAnalysis      | ✅ pass     | ✅ pass               | ✅ pass                 | ✅ pass             | ✅ pass      | ✅ pass                      |
| Recommendation     | ✅ pass     | ✅ pass               | ✅ pass                 | ✅ pass             | ✅ pass      | ✅ pass                      |
| ResumeOptimization | ✅ pass     | ✅ pass               | ✅ pass                 | ✅ pass             | ✅ pass      | ✅ pass                      |

## 9. 测试命令与结果

```bash
cd server && go test ./...
# 全部通过

make llm-check
# 通过

python3 -m json.tool server/prompts/schema/resume_optimization_schema.json >/dev/null
python3 -m json.tool server/prompts/schema/match_analysis_schema.json >/dev/null
python3 -m json.tool server/prompts/schema/recommendation_schema.json >/dev/null
# 3 个 Schema 均 VALID
```

## 10. 当前限制

- malformed JSON 测试通过 fake Pipeline 模拟，未使用真实 JSON repair Agent
- Service 层的完整集成测试（含 DB 交互）不在本阶段范围
- 未调用真实 LLM

## 11. 是否可以进入 4-3.8.2

可以。三类 Agent 均已通过本地集成测试补齐。

## 12. 下一步建议

进入 4-3.8.2：三类 Agent 集成验收收口。
