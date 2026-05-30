# 2026-05-30-073-JSONRepairAgent调用链路

## 1. 本次任务

实现阶段 4-2.7：JSONRepairAgent 调用链路。

当业务 LLM 输出发生 JSON parse 失败或 Schema validate 失败时，调用 JSONRepairAgent 尝试修复一次，然后对 repair output 再次执行 JSON extract / parse / Schema validate。

## 2. 修改范围

### 新增文件
- `server/internal/agent/repair/types.go`：TargetOutputType、RepairRequest、RepairResult、7 个错误码
- `server/internal/agent/repair/json_repair_agent.go`：Agent 结构体、NewAgent、Repair 方法
- `server/internal/agent/repair/json_repair_agent_test.go`：14 项单元测试

### 修改文件
- 11 个文档文件（docs/）

### 未修改
- .env / .env.example / Prompt 文件 / Schema 文件 / 业务 Agent / service / controller / web / Makefile / docker-compose.yml

## 3. JSONRepairAgent 设计

JSONRepairAgent 只负责 JSON 结构修复，不负责业务分析。核心约束：

- 只修复结构，不重新做业务分析
- 只执行一次，不递归，不重试
- 通过 llm.Client 调用模型，不直接依赖 HunyuanClient
- repair output 必须再次经过 JSON extract / parse / Schema validate
- repair 成功后 generated_by 仍保持 llm_agent
- repair 失败返回结构化错误，不执行 Mock fallback

### Agent 结构体

```go
type Agent struct {
    llmClient    llm.Client
    promptLoader *prompt.Loader
    schemaLoader *schema.Loader
    validator    *schema.Validator
}
```

### 核心方法

```go
func (a *Agent) Repair(ctx context.Context, req RepairRequest) *RepairResult
```

Repair 方法执行步骤：
1. 验证 TargetOutputType
2. 加载并渲染 json_repair Prompt
3. 调用 llm.Client 一次
4. 对 repair output 执行 ExtractAndParse
5. 对 parsed object 执行 Schema validate

## 4. Prompt 渲染

加载 `json_repair` Prompt（json_repair_v1.md），替换 5 个占位符：

| 占位符 | 来源 |
|--------|------|
| `{{TARGET_OUTPUT_TYPE}}` | RepairRequest.TargetOutputType |
| `{{TARGET_SCHEMA_JSON}}` | RepairRequest.TargetSchemaJSON |
| `{{BROKEN_MODEL_OUTPUT}}` | RepairRequest.BrokenOutput |
| `{{VALIDATION_ERRORS}}` | RepairRequest.ValidationErrors |
| `{{REPAIR_CONTEXT}}` | RepairRequest.RepairContext |

不把 API Key 放入 Prompt。不把完整简历或 JD 放入 repair context。

## 5. LLM 调用链路

JSONRepairAgent 通过 `llm.Client` 调用模型：

- System message: "你是 JSONRepairAgent，只输出严格 JSON。"
- User message: 渲染后的 repair prompt
- Temperature: 0（修复场景需要确定性）
- MaxTokens: 8192
- Metadata 包含 agent_name / prompt_name / target_output_type（不含 API Key）

## 6. Repair 后处理

repair raw output 返回后，必须执行：

1. `llmjson.ExtractAndParse(repairRawOutput)` — 提取并解析 JSON
2. `schema.Validator.Validate(targetSchema, parsedObject)` — Schema 校验

如果 parse 失败：返回 `JSON_REPAIR_PARSE_FAILED`
如果 validate 失败：返回 `JSON_REPAIR_SCHEMA_FAILED`
如果成功：返回 `RepairResult{Success: true, RepairSuccess: true, Parsed: ..., RepairedJSON: ...}`

## 7. repair once 约束

- Repair 方法只调用 llm.Client 一次
- 如果 repair 后仍 parse 失败，不再 repair
- 如果 repair 后 Schema validate 失败，不再 repair
- 不允许递归
- 不允许 loop retry
- 测试中验证 fakeClient.calls == 1

## 8. 错误码设计

| 错误码 | 含义 |
|--------|------|
| JSON_REPAIR_FAILED | 通用 repair 失败 |
| JSON_REPAIR_PARSE_FAILED | repair output JSON parse 失败 |
| JSON_REPAIR_SCHEMA_FAILED | repair output Schema validate 失败 |
| JSON_REPAIR_LLM_FAILED | LLM 调用失败 |
| JSON_REPAIR_PROMPT_LOAD_FAILED | Prompt 加载失败 |
| JSON_REPAIR_SCHEMA_LOAD_FAILED | Schema 加载失败 |
| JSON_REPAIR_TARGET_INVALID | TargetOutputType 不合法 |

错误信息不包含 API Key、完整 Prompt、完整 BrokenOutput、完整简历、完整 JD。

`safeErrorMessage` 函数截断长错误消息并脱敏疑似 API Key 模式。

## 9. 测试结果

14 项单元测试，全部通过：

| 测试 | 说明 | 结果 |
|------|------|------|
| TestRepair_MatchAnalysis_Success | 匹配分析 repair 成功 | ✅ |
| TestRepair_Recommendation_Success | 推荐 repair 成功 | ✅ |
| TestRepair_ResumeOptimization_Success | 简历优化 repair 成功 | ✅ |
| TestRepair_UnknownTargetOutputType | unknown target type 返回错误 | ✅ |
| TestRepair_PromptLoadFailed | Prompt 加载失败返回错误 | ✅ |
| TestRepair_LLMError | LLM 返回错误时返回 JSON_REPAIR_LLM_FAILED | ✅ |
| TestRepair_ParseFailed | repair output 非法 JSON 时返回 JSON_REPAIR_PARSE_FAILED | ✅ |
| TestRepair_SchemaValidateFailed | repair output Schema validate 失败时返回 JSON_REPAIR_SCHEMA_FAILED | ✅ |
| TestRepair_OnceConstraint | fakeClient calls == 1 | ✅ |
| TestRepair_Success_ParsedObjectNotEmpty | repair 成功后 parsed object 不为空 | ✅ |
| TestRepair_Success_RawOutputLength | repair 成功后 raw_output_length > 0 | ✅ |
| TestRepair_ErrorDoesNotContainFakeAPIKey | 错误信息不包含 fake API key | ✅ |
| TestRepair_LLMEmptyResponse | LLM 空响应返回错误 | ✅ |
| TestRepair_LLMErrorWithLLMError | LLMError 类型错误映射 | ✅ |

全量测试 `go test ./...` 通过，`go vet ./...` 无警告。

## 10. 文档更新

已更新 11 个文档：
- 阶段 4-2 LLM Client 与 Hunyuan 接入方案
- LLM Client 设计
- Hunyuan 接入准备
- JSON 输出规范
- 失败兜底与 Mock 回退策略
- Prompt 模板
- 项目状态
- 下一步任务
- 变更记录
- 说明书
- 本开发日志

## 11. 未修改内容

- 未修改 .env / .env.example
- 未写入真实 API Key
- 未调用真实 Hunyuan
- 未修改 Prompt 文件
- 未修改 Schema 文件
- 未修改业务 Agent（match_mock / recommendation_mock / version_compare_mock）
- 未实现 Mock fallback
- 未实现统一 LLM Agent Pipeline
- repair 只执行一次
- 业务 Agent 接入放到 4-3

## 12. 当前风险

| 风险 | 严重度 | 说明 |
|------|--------|------|
| repair LLM 调用可能增加延迟 | 🟡 中 | repair 失败时增加一次 LLM 调用延迟 |
| repair 可能改变原始业务语义 | 🟢 低 | Prompt 明确约束只修结构不重做分析 |
| repair 成功率取决于模型能力 | 🟡 中 | 不同模型 repair 效果可能不同 |

## 13. 是否可以进入 4-2.8

**可以。** 4-2.7 已完成 JSONRepairAgent repair once 调用链路。当前已具备：

1. Prompt Loader 加载 repair Prompt
2. llm.Client 调用模型
3. llmjson ExtractAndParse 处理 repair output
4. schema.Validator 校验 repair output
5. repair once 约束（不递归不重试）
6. 结构化错误码

4-2.8 统一 LLM Agent Pipeline 可以基于以上能力构建完整的 Chat → parse → validate → repair → fallback 链路。

## 14. 下一步建议

进入 4-2.8：统一 LLM Agent Pipeline。

4-2.8 将实现 LLMRunner，统一编排：
1. 业务 Agent 构造 ChatRequest
2. llm.Client.Chat
3. llmjson ExtractAndParse
4. schema.Validator.Validate
5. 如果失败：JSONRepairAgent.Repair（repair once）
6. 如果 repair 仍失败：fallback decision
7. 返回稳定业务结果
