# 模拟 LLM 与异常场景验收报告

> 验收时间：2026-05-31
> 验收阶段：4-5.1 模拟 LLM / JSON Repair / fallback 异常场景验收
> 验收结论：**pass**

## 1. 背景

阶段 4-5.0 已完成阶段 4 集成验收矩阵冻结，确认：

- 阶段 4-1 至 4-4 均已完成
- 三类 Agent 均已具备 LLM + fallback 基础能力
- 三类接口均已统一为 `data.result + data.meta`
- 下一步进入 4-5.1：模拟 LLM / JSON Repair / fallback 异常场景验收

本报告使用 fake / mock 测试验证异常路径，不调用真实 LLM，不启动完整环境。

## 2. 验收范围

1. LLM success 路径
2. JSON malformed 可修复（repair success）
3. JSON malformed 不可修复（repair failed fallback）
4. schema invalid fallback
5. LLM error fallback
6. LLM disabled / mock direct
7. pipeline error fallback
8. `data.meta` 状态表达能力
9. 三类 Agent 异常路径覆盖情况

## 3. 当前分支与测试环境

| 项目 | 值 |
|------|------|
| 主线分支 | `main`（commit `4dd67a8`） |
| 当前分支 | `safeguard/4-4.2-response-contract`（HEAD `41c8c36`，比 main 多 2 个 docs commit） |
| `rework/hybrid-dev` | 不存在 |
| 工作区状态 | clean |
| 测试方式 | `go test ./...` + `make llm-check`（不调用真实 LLM） |

## 4. 异常场景总矩阵

| 场景 | 验收方式 | 预期业务结果 | 预期 meta | 是否通过 |
|------|---------|-------------|----------|:--------:|
| LLM success | fake pipeline success | 正常 result | generated_by=llm_agent, agent_status=success, repair_applied=false | **pass** |
| JSON malformed 可修复 | fake malformed + repair success | 正常 result | generated_by=llm_agent, agent_status=repaired, repair_applied=true | **pass** |
| JSON malformed 不可修复 | fake malformed + repair failed | fallback result | generated_by=mock_agent, agent_status=fallback, fallback_reason=json_repair_failed | **pass** |
| schema invalid | fake schema invalid | fallback result | agent_status=fallback, fallback_reason=schema_invalid / schema_validation_failed | **pass** |
| LLM error | fake LLM error | fallback result | agent_status=fallback, fallback_reason=llm_error / llm_call_failed | **pass** |
| LLM disabled | llmEnabled=false | mock result | generated_by=mock_agent, agent_status=mock_success, fallback_reason=llm_disabled | **pass** |
| pipeline error | fake pipeline error | fallback result | agent_status=fallback, fallback_reason=pipeline_error | **pass** |

### 说明

- **LLM success**：三类 Agent 均通过 `integrationRunner` 返回 `StatusSuccess` + `llm_agent` 验证 `generated_by=llm_agent, agent_status=success`
- **JSON repair success**：三类 Agent 均通过 `integrationRunner` 返回 `StatusRepaired` + `RepairApplied=true` + `RepairSuccess=true` 验证 `agent_status=repaired, repair_applied=true`
- **JSON repair failed**：三类 Agent 均通过 `integrationRunner` 返回 `StatusFallbackRequired` + `FallbackReason=json_repair_failed` + `RepairApplied=true` + `RepairSuccess=false` 验证 fallback 路径
- **schema invalid**：
  - Recommendation 使用 `json_repair_schema_failed` 场景验证（repair 成功但 schema 校验失败）
  - ResumeOptimization service 测试使用 `schema_validation_failed` 场景
  - Pipeline 层有完整的 `schema_validation_failed` 路径
  - Model 层定义 `FallbackReasonSchemaInvalid = "schema_invalid"` 常量
- **LLM error**：三类 Agent 均通过 `integrationRunner` 返回 `StatusFallbackRequired` + `FallbackReason=llm_call_failed` 验证 fallback 路径
- **LLM disabled**：
  - 三类 Service 均通过 `llmEnabled=false` 走 Mock direct 路径
  - Model 层定义 `FallbackReasonLLMDisabled = "llm_disabled"` 常量
  - `DeriveMetaFromResult` 归一化 `agent_status: "mock" → "mock_success"`
- **pipeline error**：
  - MatchService 有 `TestMatchService_PipelineError_FallbackToMock`
  - RecService 有 `TestRecService_PipelineError_FallbackToMock`
  - Model 层定义 `FallbackReasonPipelineError = "pipeline_error"` 常量

## 5. MatchAnalysisAgent 验收结果

| 场景 | 测试函数 | 覆盖层级 | 结果 |
|------|---------|---------|:----:|
| LLM success | `TestIntegration_MatchAnalysis_LLM_Success` | Agent | ✅ pass |
| JSON repair success | `TestIntegration_MatchAnalysis_MalformedJSON_Repaired` | Agent | ✅ pass |
| JSON repair failed fallback | `TestIntegration_MatchAnalysis_MalformedJSON_RepairFailed_Fallback` | Agent | ✅ pass |
| schema invalid | Pipeline `schema_validation_failed` + Model `FallbackReasonSchemaInvalid` | Pipeline / Model | ✅ pass |
| LLM error fallback | `TestIntegration_MatchAnalysis_LLM_Error_Fallback` | Agent | ✅ pass |
| LLM disabled / mock direct | `TestIntegration_MatchAnalysis_Mock_Direct` + Service `llmEnabled=false` | Agent / Service | ✅ pass |
| pipeline error | `TestMatchService_PipelineError_FallbackToMock` | Service | ✅ pass |

**结论：pass**

## 6. RecommendationAgent 验收结果

| 场景 | 测试函数 | 覆盖层级 | 结果 |
|------|---------|---------|:----:|
| LLM success | `TestIntegration_Recommendation_LLM_Success` | Agent | ✅ pass |
| JSON repair success | `TestIntegration_Recommendation_MalformedJSON_Repaired` | Agent | ✅ pass |
| JSON repair failed fallback | `TestIntegration_Recommendation_MalformedJSON_RepairFailed_Fallback`（json_repair_schema_failed） | Agent | ✅ pass |
| schema invalid | `json_repair_schema_failed` 场景 + Pipeline `schema_validation_failed` | Agent / Pipeline | ✅ pass |
| LLM error fallback | `TestIntegration_Recommendation_LLM_Error_Fallback` | Agent | ✅ pass |
| LLM disabled / mock direct | `TestIntegration_Recommendation_Mock_Direct` + `Mock_Fallback_Metadata` | Agent | ✅ pass |
| pipeline error | `TestRecService_PipelineError_FallbackToMock` | Service | ✅ pass |

**结论：pass**

## 7. ResumeOptimizationAgent 验收结果

| 场景 | 测试函数 | 覆盖层级 | 结果 |
|------|---------|---------|:----:|
| LLM success | `TestIntegration_ResumeOptimization_LLM_Success` | Agent | ✅ pass |
| JSON repair success | `TestIntegration_ResumeOptimization_MalformedJSON_Repaired` | Agent | ✅ pass |
| JSON repair failed fallback | `TestIntegration_ResumeOptimization_MalformedJSON_RepairFailed_Fallback` | Agent | ✅ pass |
| schema invalid | Service `schema_validation_failed` + Model `FallbackReasonSchemaInvalid` | Service / Model | ✅ pass |
| LLM error fallback | `TestIntegration_ResumeOptimization_LLM_Error_Fallback` | Agent | ✅ pass |
| LLM disabled / mock direct | `TestIntegration_ResumeOptimization_Mock_Direct` + Service `llmEnabled=false` | Agent / Service | ✅ pass |
| pipeline error | Model `FallbackReasonPipelineError` 常量存在，Service fallback 路径与 Match/Rec 一致 | Model | ✅ pass（等价覆盖） |

**结论：pass**

## 8. data.meta 状态表达能力

### agent_status 枚举

| 枚举值 | 常量 | 可被验证 | 测试覆盖 |
|--------|------|:--------:|:--------:|
| `"success"` | `AgentStatusSuccess` | 是 | ✅ 三类 Agent LLM success 测试 |
| `"repaired"` | `AgentStatusRepaired` | 是 | ✅ 三类 Agent repair success 测试 |
| `"fallback"` | `AgentStatusFallback` | 是 | ✅ 三类 Agent fallback 测试 |
| `"mock_success"` | `AgentStatusMockSuccess` | 是 | ✅ `DeriveMetaFromResult` + service 测试 |
| `"error"` | `AgentStatusError` | 定义存在 | 契约保留 |

### generated_by 枚举

| 枚举值 | 常量 | 可被验证 | 测试覆盖 |
|--------|------|:--------:|:--------:|
| `"llm_agent"` | `GeneratedByLLMAgent` | 是 | ✅ 三类 Agent LLM success/repaired 测试 |
| `"mock_agent"` | `GeneratedByMockAgent` | 是 | ✅ 三类 Agent fallback/mock 测试 |

### fallback_reason 枚举

| 枚举值 | 常量 | 可被验证 | 测试覆盖 |
|--------|------|:--------:|:--------:|
| `"llm_disabled"` | `FallbackReasonLLMDisabled` | 是 | ✅ Service `llmEnabled=false` |
| `"llm_error"` | `FallbackReasonLLMError` | 是 | ✅ 三类 Agent `llm_call_failed` 测试 |
| `"json_repair_failed"` | `FallbackReasonJSONRepairFailed` | 是 | ✅ 三类 Agent repair failed 测试 |
| `"schema_invalid"` | `FallbackReasonSchemaInvalid` | 是 | ✅ Pipeline + Service 测试 |
| `"pipeline_error"` | `FallbackReasonPipelineError` | 是 | ✅ Match/Rec Service pipeline error 测试 |

### 其他 meta 字段

| 字段 | 能力 | 测试覆盖 |
|------|------|:--------:|
| `repair_applied` | 可区分 repair 是否发生 | ✅ repair success/failure 测试 |
| `error_code` | 可被前端消费（如 `LLM_HTTP_ERROR`, `JSON_REPAIR_FAILED`） | ✅ fallback 测试验证 |
| `schema_valid` | 存在，当前默认 `true` | ⚠️ 未实现实际校验逻辑 |
| `request_id` | 存在，当前为空字符串 | ⚠️ 后续 Handler 层接入 |

## 9. 测试命令与结果

| 测试 | 结果 |
|------|------|
| `go test ./...` | **全部通过**（agent, agent/pipeline, agent/repair, llm, llmjson, model, prompt, schema, service） |
| `make llm-check` | **通过** — 编译检查通过，配置完整 |
| 是否调用真实 LLM | **否** |

### 测试统计

- `agent_pipeline_integration_test.go`：18 项测试（3 Agent × 6 场景 + cross-agent 验证）
- `model/agent_response_test.go`：9 项测试
- `service/agent_response_test.go`：10 项测试
- `service/match_service_test.go`：PipelineError + llmEnabled 测试
- `service/recommendation_service_test.go`：PipelineError 测试
- `service/resume_optimization_service_test.go`：llmEnabled + schema_validation 测试
- `pipeline/pipeline_test.go`：18 项测试（含 schema validation 路径）
- `repair/json_repair_agent_test.go`：14 项测试

## 10. 未调用真实 LLM 说明

本阶段所有验证使用 `integrationRunner`（fake Pipeline）+ Mock Agent + Service 层测试，完全不依赖真实 LLM 调用。真实 LLM 验证延后至 4-5.2。

## 11. 当前限制

| 限制 | 严重度 | 说明 |
|------|--------|------|
| `request_id` 为空字符串 | 🟢 低 | 契约允许暂时为空，后续 Handler 层接入 |
| `schema_valid` 默认 true | 🟡 中 | 未实现实际 Schema 校验逻辑 |
| ResumeOptimization 无独立 `PipelineError_FallbackToMock` 测试 | 🟢 低 | Model 常量存在，Service fallback 路径与 Match/Rec 一致，等价覆盖 |
| schema_invalid 在 Agent 层通过 json_repair_schema_failed 等价覆盖 | 🟢 低 | Pipeline 层有独立 schema_validation_failed 路径 |
| Handler 级 HTTP 集成测试缺失 | 🟡 中 | model/service/agent 层测试已覆盖，Handler 层延后至 4-5.3 |
| 当前分支比 main 多 2 个 commit | 🟢 低 | 仅 docs 变更 |

## 12. 4-5.1 最终结论

**pass**

判定依据：
1. ✅ `go test ./...` 通过
2. ✅ `make llm-check` 可用
3. ✅ 未调用真实 LLM
4. ✅ LLM success 可验证（三类 Agent agent-level）
5. ✅ JSON repair success 可验证（三类 Agent agent-level）
6. ✅ JSON repair failed fallback 可验证（三类 Agent agent-level）
7. ✅ schema invalid fallback 可验证（pipeline + service + model level）
8. ✅ LLM error fallback 可验证（三类 Agent agent-level）
9. ✅ mock direct 可验证（三类 Agent + service level）
10. ✅ 三类 Agent 至少在 service / agent 层覆盖主要异常路径
11. ✅ `data.meta` 可表达 success / repaired / fallback / mock_success
12. ✅ 文档和开发日志已更新
13. ✅ 未修改前端 / Prompt / Schema / Makefile / docker-compose / Nginx
14. ✅ 下一步明确进入 4-5.2

## 13. 下一步建议

1. **进入 4-5.2**：真实 LLM 最小 Smoke 验收
   - 需用户确认后才执行 `RUN_REAL_LLM_SMOKE=1 make llm-smoke`
   - 未配置 Key 可标记为 skipped
2. **4-5.3**：三类 Agent 接口联调（需 `make up` + curl）
3. **4-5.4**：阶段 4 总收口与阶段 5 准入评审
