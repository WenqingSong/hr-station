# 2026-05-31-115-三类AgentLLMFallback本地集成测试补齐

## 1. 本次任务

阶段 4-3.8.1：三类 Agent LLM / fallback 本地集成测试补齐。

目标：使用 fake Pipeline / mock / fallback 统一验证 MatchAnalysisAgent、RecommendationAgent、ResumeOptimizationAgent 的 LLM success、malformed JSON / repair / fallback、LLM error fallback、Mock direct 等行为，并断言 apply_readiness、optimization_stop_signal、top_fix_items 与 generated_by / agent_status。

## 2. 背景

- 4-3.8.0 已完成：三类 Agent 当前 LLM / fallback 状态核验与验收矩阵冻结
- 4-3.8.1-R 已完成：ResumeOptimizationAgent Pipeline 最小集成，三类 Agent 均具备 LLM + fallback 基础能力
- 前置测试覆盖：MatchAnalysis 和 Recommendation 有 Agent + Service 层测试，ResumeOptimization 有 Agent + Service 层测试
- 缺口：缺少统一集成测试、缺少 malformed JSON / repair / fallback 场景、缺少 apply_readiness / optimization_stop_signal / top_fix_items 统一断言

## 3. 修改范围

- 新增：`server/internal/agent/agent_pipeline_integration_test.go`
- 更新：`docs/05-Agent设计/三类Agent当前LLMFallback状态核验与验收矩阵.md`
- 更新：`docs/项目状态.md`
- 更新：`docs/下一步任务.md`
- 更新：`docs/变更记录.md`
- 更新：`docs/说明书.md`

## 4. 测试设计

### Fake Pipeline / Fake Runner

使用 `integrationRunner` 统一 fake runner，支持三种行为：
- **success**：返回合法 Parsed + generated_by=llm_agent + agent_status=success
- **repaired**：返回合法 Parsed + generated_by=llm_agent + agent_status=repaired + repair_applied=true
- **fallback**：返回 fallback_required + generated_by=mock_agent + error 信息

### 测试矩阵

| Agent              | LLM success | malformed JSON repair | repair failed fallback | LLM error fallback | Mock direct | generated_by / agent_status |
| ------------------ | ----------- | -------------------- | ---------------------- | ------------------ | ----------- | --------------------------- |
| MatchAnalysis      | ✅          | ✅                    | ✅                      | ✅                  | ✅           | ✅                           |
| Recommendation     | ✅          | ✅                    | ✅                      | ✅                  | ✅           | ✅                           |
| ResumeOptimization | ✅          | ✅                    | ✅                      | ✅                  | ✅           | ✅                           |

## 5. MatchAnalysisAgent 测试

- `TestIntegration_MatchAnalysis_LLM_Success`：LLM success，断言 generated_by=llm_agent, agent_status=success, apply_readiness 非空, optimization_stop_signal 非空, top_fix_items 非空, evidence_chain 非空, offer_go_summary 非空
- `TestIntegration_MatchAnalysis_LLM_Error_Fallback`：LLM error，断言 MatchFallbackError
- `TestIntegration_MatchAnalysis_MalformedJSON_Repaired`：repair success，断言 generated_by=llm_agent, agent_status=repaired, repair_applied=true, 战略字段保留
- `TestIntegration_MatchAnalysis_MalformedJSON_RepairFailed_Fallback`：repair failed，断言 MatchFallbackError + fallback_reason=json_repair_failed
- `TestIntegration_MatchAnalysis_Mock_Direct`：Mock 直接返回，断言 score 范围、核心字段非空
- `TestIntegration_MatchAnalysis_Mock_Fallback_Metadata`：Mock fallback，断言 generated_by=mock_agent, agent_status=fallback, 战略字段保留

## 6. RecommendationAgent 测试

- `TestIntegration_Recommendation_LLM_Success`：LLM success，断言 generated_by=llm_agent, agent_status=success, apply_readiness/optimization_stop_signal/top_fix_items 非空, recommend_reason/offer_go_summary 非空
- `TestIntegration_Recommendation_LLM_Error_Fallback`：LLM error，断言 RecFallbackError
- `TestIntegration_Recommendation_MalformedJSON_Repaired`：repair success，断言同 MatchAnalysis
- `TestIntegration_Recommendation_MalformedJSON_RepairFailed_Fallback`：repair failed，断言 RecFallbackError + fallback_reason=json_repair_schema_failed
- `TestIntegration_Recommendation_Mock_Direct`：Mock 直接返回
- `TestIntegration_Recommendation_Mock_Fallback_Metadata`：Mock fallback metadata

## 7. ResumeOptimizationAgent 测试

- `TestIntegration_ResumeOptimization_LLM_Success`：LLM success，断言 generated_by=llm_agent, agent_status=success, apply_readiness/optimization_stop_signal/top_fix_items 非空, summary/verification_notice 非空
- `TestIntegration_ResumeOptimization_LLM_Error_Fallback`：LLM error，断言 ResumeOptFallbackError
- `TestIntegration_ResumeOptimization_MalformedJSON_Repaired`：repair success
- `TestIntegration_ResumeOptimization_MalformedJSON_RepairFailed_Fallback`：repair failed，断言 ResumeOptFallbackError
- `TestIntegration_ResumeOptimization_Mock_Direct`：Mock 直接返回
- `TestIntegration_ResumeOptimization_Mock_Fallback_Metadata`：Mock fallback metadata

## 8. 跨 Agent 统一测试

- `TestIntegration_GeneratedBy_AgentStatus_Rules`：验证 generated_by / agent_status 规则一致性
- `TestIntegration_StrategicFields_AllAgents`：三类 Agent 统一断言 apply_readiness / optimization_stop_signal / top_fix_items
- `TestIntegration_DefaultStrategicFields_WhenMissing`：LLM 输出缺字段时默认值验证
- `TestIntegration_AllAgents_InterfaceCompliance`：接口合规性交叉检查

## 9. 测试结果

- `go test ./...`：✅ 全部通过
- `make llm-check`：✅ 通过
- JSON Schema 校验：✅ 3 个 Schema 均 VALID
- 未调用真实 LLM

## 10. 未修改内容

- 未修改 Prompt
- 未修改 Schema
- 未修改前端
- 未修改 Makefile
- 未修改 docker-compose
- 未修改 Nginx
- 未调用真实 LLM
- 未请求腾讯招聘接口
- 未执行 reset / reset-up
- 未清空 volume

## 11. 当前限制

- malformed JSON 测试通过 fake Pipeline 模拟，未使用真实 JSON repair Agent
- 当前测试在 Agent 层面验证，Service 层的完整集成测试（含 DB 交互）不在本阶段范围
- Recommendation 的 apply_readiness / optimization_stop_signal / top_fix_items 在 Agent 层已有，但 Service 层已有测试未显式断言

## 12. 是否可以进入 4-3.8.2

可以。三类 Agent 均已通过本地集成测试补齐，覆盖 LLM success / error fallback / malformed JSON repair / Mock direct / Mock fallback 等场景。

## 13. 下一步建议

进入 4-3.8.2：三类 Agent 集成验收收口。
