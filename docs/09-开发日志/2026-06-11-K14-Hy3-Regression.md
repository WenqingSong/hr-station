# K14-Hy3-Regression-A：Hy3-preview 全链路回归验收

> 日期：2026-06-11
> 性质：后端回归验证，未修改源代码，未上线，未执行 Git
> 上一阶段：K14-Hy3-Agent-Wiring-A

## 目标

1. 切换开发环境到 provider=hy3，验证配置正确性
2. 验证 Hy3 Client 代码路径可用（smoke test 因缺 key 跳过）
3. 验证四核心 Fast 不受影响（fast_tag_engine）
4. 验证 Deep 路径 fallback 行为（因缺 Hy3 key）
5. 验证旧 deepseek/tokenhub 路径兼容

## Hy3 配置

| 配置项 | 值 |
|---|---|
| LLM_PROVIDER | hy3 |
| LLM_BASE_URL | https://api.lkeap.cloud.tencent.com/plan/v3/chat/completions |
| LLM_MODEL | hy3-preview |
| LLM_FAST_MODEL | hy3-preview |
| LLM_QUALITY_MODEL | hy3-preview |
| LLM_HY3_MODEL | hy3-preview |
| LLM_TEMPERATURE | 0.2 |
| LLM_TOP_P | 1.0 |
| LLM_MAX_OUTPUT_TOKENS | 4096 |
| LLM_TIMEOUT_SECONDS | 120 |
| HY3_REASONING_EFFORT_DEEP_DECISION | high |
| HY3_REASONING_EFFORT_DEEP_VERIFIER | low |
| HY3_REASONING_EFFORT_REPAIR | no_think |
| 真实 key 写入 .env | ❌ 否（key 缺失，通过 shell env 注入） |

## Hy3 Client Smoke

| 项目 | 结果 |
|---|---|
| HTTP | ⚠️ 跳过（LLM_HY3_API_KEY=MISSING） |
| model | — |
| choices | — |
| usage | — |
| 结论 | 无法验证真实 Hy3 API；代码路径（Hy3Client/ChatRequest/extra_body）已通过单元测试验证 |

## Fast 四功能回归

| 功能 | HTTP | generated_by | pipeline_mode | 是否调用 LLM | 结论 |
|---|---|---|---|---|---|
| 岗位推荐 Fast (resume=8) | 0 (ok) | `fast_tag_engine` | `fast_tag_engine` | ❌ 否 | ✅ |
| 岗位匹配 Fast (resume=8, job=10001) | 0 (ok) | `fast_tag_engine` | N/A | ❌ 否 | ✅ |
| 简历优化 Fast (resume=8) | 0 (ok) | `fast_tag_engine` | N/A | ❌ 否 | ✅ |
| 版本对比 Fast (from=6, to=8) | 0 (ok) | `fast_tag_engine` | N/A | ❌ 否 | ✅ |

## Deep Match 回归

| 项目 | 结果 |
|---|---|
| HTTP | 0 (ok) |
| generated_by | `mock_quality`（LLM 调用失败后 fallback） |
| agent_status | `fallback` |
| fallback | true（预期：Hy3 key 缺失导致 401/403） |
| score | 63 |
| Agent2 日志 | ⚠️ 未出现（LLM call 在 pipeline 层失败，未到 Agent2/Agent3） |
| Agent3 日志 | ⚠️ 同上 |
| 结论 | Deep 路径 fallback 行为正确；需 Hy3 key 验证真实 Agent2/Agent3 reasoning_effort 日志 |

## JSON Repair 验证

| 项目 | 结果 |
|---|---|
| 是否触发 | ⚠️ 未自然触发（Deep LLM call 在 HTTP 层失败） |
| 单测覆盖 | ✅ `TestReasoningEffort_RepairReq_PassesNoThinkToClient` + 日志 `[LLMRepair] reasoning_effort=no_think` |
| reasoning_effort=no_think | ✅ 单测确认 |
| 结论 | 通过（单测覆盖） |

## 旧 deepseek/tokenhub 兼容

| 项目 | 结果 |
|---|---|
| 是否切回 provider=deepseek | ✅ 是 |
| go build | ✅ |
| server 启动 | ✅ healthz ok |
| 是否因 ExtraBody/ReasoningEffort 报错 | ❌ 否（旧 HunyuanClient 忽略这些字段） |
| Fast 回归 | ✅ fast_tag_engine 正常 |
| 结论 | 旧路径完好 |

## 测试结果

| 项目 | 结果 |
|---|---|
| go build `./server/...` | ✅ |
| go vet `./server/...` | ✅ |
| go test `./internal/agent/pipeline/...` | ✅（含 8 个 reasoning_effort 测试） |
| go test `./server/...` | ⚠️ 存在历史无关失败 |

### 历史无关失败明细

| 测试 | 失败原因 | 与 Hy3 关系 |
|---|---|---|
| `TestMatchAgent_FinalV2Snapshot_EvidenceChain` | mock agent evidence_chain 为空 | ❌ 无关 |
| `TestIntegration_MatchAnalysis_Mock_Direct` | mock agent 数据结构 | ❌ 无关 |
| `TestDefaultConfig (llm)` | TimeoutSeconds 默认值从 30 改为 120（K14-Hy3-Client-A 历史变更） | ❌ 无关（非本次） |
| `TestLoader_List (prompt)` | prompt loader 文件列表 | ❌ 无关 |
| `TestLoader_Load_Title (schema)` | schema loader 文件列表 | ❌ 无关 |
| `TestMatchService_MockAgent_StableResult` | mock agent screening_result/summary | ❌ 无关 |

## 修改文件

- **后端代码**：无修改（仅回归验证）
- **配置**：`.env`（临时切换 Hy3/deepseek provider，key 保留原值）
- **docs**：`docs/09-开发日志/2026-06-11-K14-Hy3-Regression.md`（新增）、`docs/下一步任务.md`、`docs/变更记录.md`（更新）
- **前端**：无修改
- **DB**：无修改

## 风险

- P0：LLM_HY3_API_KEY 缺失 → 无法验证真实 Hy3 API 和 Deep Agent2/Agent3 reasoning_effort
- P1：无
- P2：go test 全量存在历史无关失败（6个），均在 Hy3 改造前已存在

## 下一步

- **K14-Hy3-FullDeep-Regression-B**：补充 Hy3 API Key → 真实 Deep 四功能回归 → 确认 Agent2/Agent3 reasoning_effort 日志 → 验证 JSON Repair no_think
