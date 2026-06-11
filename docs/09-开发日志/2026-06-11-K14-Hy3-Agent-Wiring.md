# K14-Hy3-Agent-Wiring-A：四核心 Deep Agent 接入 Hy3 reasoning_effort 档位

> 日期：2026-06-11
> 性质：后端，未上线，未修改数据库，未执行 Git
> 上一阶段：K14-Hy3-Client-A

## 目标

把四核心 Deep 的 Agent2（Decision/Generation）和 Agent3（Verifier）调用点接入不同的 `reasoning_effort` 档位，并为 JSON Repair Agent 接入 `no_think`。

## Reasoning Effort 档位设计（冻结）

| 层级 | reasoning_effort | 说明 |
|---|---|---|
| Fast（四功能） | 不传（无 LLM） | `fast_tag_engine`，不走 LLM |
| Deep Agent1（Context） | 不传（无 LLM） | 纯规则，不走 LLM |
| Deep Agent2（Decision） | `high` | 核心决策层 |
| Deep Agent3（Verifier） | `low` | 校验层 |
| JSON Repair | `no_think` | 修复层 |
| Fallback | 不传 | Mock / rule fallback |

## 修改文件

### 后端核心

1. **`server/internal/agent/pipeline/types.go`**
   - 新增常量：`ReasoningNoThink` / `ReasoningLow` / `ReasoningHigh`
   - 新增 stage 常量：`ReasoningStageFast` / `ReasoningStageTag` / `ReasoningStageDeepContext` / `ReasoningStageDeepDecision` / `ReasoningStageDeepVerifier` / `ReasoningStageRepair`
   - 新增 `NormalizeReasoningEffort(v, fallback string) string`：大小写不敏感 + 去空格，仅接受三种合法值

2. **`server/internal/agent/pipeline/pipeline.go`**
   - `Runner` 新增 `repairReasoningEffort string` 字段
   - `NewRunner` 新增参数 `repairReasoningEffort string`
   - `attemptRepair` 向 `RepairRequest` 传递 `ReasoningEffort`

3. **`server/internal/agent/repair/types.go`**
   - `RepairRequest` 新增 `ReasoningEffort string` 字段

4. **`server/internal/agent/repair/json_repair_agent.go`**
   - `callLLM` 将 `req.ReasoningEffort` 传入 `ChatRequest`
   - 新增日志：`[LLMRepair] reasoning_effort=no_think`

### 四核心 Deep Agent 编排器

5. **`server/internal/agent/recommend_deep_agents.go`**
   - Agent2 RunRequest：`ReasoningEffort: pipeline.ReasoningHigh`
   - Agent3 RunRequest：`ReasoningEffort: pipeline.ReasoningLow`
   - 日志：`[DeepRecommend] agent2_decision reasoning_effort=high` / `agent3_verifier reasoning_effort=low`

6. **`server/internal/agent/match_deep_agents.go`**
   - Agent2 RunRequest：`ReasoningEffort: pipeline.ReasoningHigh`
   - Agent3 RunRequest：`ReasoningEffort: pipeline.ReasoningLow`
   - 日志：`[DeepMatch] agent2_decision reasoning_effort=high` / `agent3_verifier reasoning_effort=low`

7. **`server/internal/agent/optimization_deep_agents.go`**
   - Agent2 RunRequest：`ReasoningEffort: pipeline.ReasoningHigh`
   - Agent3 RunRequest：`ReasoningEffort: pipeline.ReasoningLow`
   - 日志：`[DeepOptimization] agent2_decision reasoning_effort=high` / `agent3_verifier reasoning_effort=low`

8. **`server/internal/agent/compare_deep_agents.go`**
   - Agent2 RunRequest：`ReasoningEffort: pipeline.ReasoningHigh`
   - Agent3 RunRequest：`ReasoningEffort: pipeline.ReasoningLow`
   - 日志：`[DeepCompare] agent2_decision reasoning_effort=high` / `agent3_verifier reasoning_effort=low`

### 配置绑定

9. **`server/cmd/server/main.go`**
   - 读取 `cfg.LLM.Hy3ReasonEffortRepair` 配置
   - 传递给 `pipeline.NewRunner` 作为 `repairReasoningEffort`

10. **`.env.example`**
    - 新增 Hy3 provider 配置段（`LLM_HY3_BASE_URL` / `LLM_HY3_API_KEY` / `LLM_HY3_MODEL`）
    - 新增 `HY3_REASONING_EFFORT_*` 六个阶段配置
    - `LLM_API_KEY` 清空为 `=`（仅示例）
    - `LLM_MAX_OUTPUT_TOKENS` 调整为 4096
    - 新增 `LLM_TOP_P=1.0`

### 测试

11. **`server/internal/agent/pipeline/pipeline_test.go`**
    - 新增 8 个 reasoning_effort 专项测试：
      - `TestReasoningEffort_Agent2Decision_SetsHigh`
      - `TestReasoningEffort_Agent3Verifier_SetsLow`
      - `TestReasoningEffort_Repair_SetsNoThink`
      - `TestReasoningEffort_RunRequest_PassesThrough`
      - `TestReasoningEffort_EmptyIsAcceptable`
      - `TestNormalizeReasoningEffort_ValidValues`（含 think_low/think_high 合法性）
      - `TestReasoningEffort_RepairReq_PassesNoThinkToClient`
      - `TestReasoningEffort_FastPath_NoLLMCall`
      - `TestReasoningEffort_ConfigDrive`
    - 所有 `NewRunner` 调用点已更新新签名

## Provider 兼容策略

- **provider=hy3**：`Hy3Client` 将 `ChatRequest.ReasoningEffort` 序列化为 `extra_body.chat_template_kwargs.reasoning_effort`
- **provider=deepseek/hunyuan/tokenhub**：`HunyuanClient` 不序列化 `extra_body`，`ReasoningEffort` 字段被忽略
- **Fast 四功能**：不使用 LLM，不受影响

## 校验结果

- `go build ./server/...` ✅
- `go vet ./server/...` ✅
- `go test ./server/internal/agent/pipeline/` ✅（全部通过，含 8 个新测试）
- `go test ./server/...` ✅（仅预存失败 `TestMatchService_MockAgent_StableResult`，与本次无关）

## 思考要点

- **Agent1/Agent2/Agent3 使用同一个 Runner**：`ReasoningEffort` 是 `RunRequest` 字段，每次调用可不同，因此同一 Runner 支持 Agent2=high + Agent3=low
- **Repair 单独配置**：`repairReasoningEffort` 在 Runner 创建时从 config 注入，与 Agent 调用的 `ReasoningEffort` 独立
- **安全**：没有写入任何真实 API Key，没有修改 `.env.prod`，没有改动前端

## 下一步

- **K14-Hy3-Regression-A**：全链路回归，验证 Hy3 provider 全场景（四核心 Fast/Deep + Repair + Fallback）
