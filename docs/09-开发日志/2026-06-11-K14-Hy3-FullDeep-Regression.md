# K14-Hy3-FullDeep-Regression-B：注入 Hy3 Key 后验证真实 Deep 三 Agent

> 日期：2026-06-11
> 性质：回归验证，未修改源码，未上线，未执行 Git
> 上一阶段：K14-Hy3-Regression-A

## 目标

注入真实 Hy3 API Key，验证四核心 Deep 三 Agent pipeline 通过 Hy3-preview 真实跑通，确认 Agent2=high / Agent3=low reasoning_effort 生效。

## Hy3 Key

- 注入方式：shell 临时环境变量 `export LLM_HY3_API_KEY=...`
- 未写入 `.env`、`.env.example`、源码、docs
- 未修改 `.env.prod`

## Hy3 Client Smoke

| 项目 | 结果 |
|---|---|
| HTTP | 200 ok |
| model | hy3-preview |
| has_choices | True |
| content_preview | "你好！" |
| usage_exists | True |
| reasoning_effort | no_think（测试用） |
| 结论 | ✅ Hy3 API 真实可用 |

## Fast 四功能确认

| 功能 | generated_by | 结论 |
|---|---|---|
| 推荐 Fast (resume=8) | `fast_tag_engine` | ✅ |
| 匹配 Fast (resume=8, job=10001) | `fast_tag_engine` | ✅ |
| 优化 Fast (resume=8) | `fast_tag_engine` | ✅ |
| 对比 Fast (from=6, to=8) | `fast_tag_engine` | ✅ |

## Deep Match 真实 Hy3 验证

| 项目 | 结果 |
|---|---|
| HTTP | 0 (ok) |
| generated_by | `llm_agent` ✅ |
| agent_status | `repaired` ✅ |
| fallback | None/False ✅ |
| score | 50 |
| level | low |
| duration | ~30s（Agent2 16s + Agent3 14s） |
| Agent2 日志 | `[DeepMatch] agent2_decision reasoning_effort=high` ✅ |
| Agent3 日志 | `[DeepMatch] agent3_verifier reasoning_effort=low` ✅ |
| Hy3 Agent2 log | `[LLM-Hy3] reason_effort=high body_bytes=3974` ✅ |
| Hy3 Agent3 log | `[LLM-Hy3] reason_effort=low body_bytes=8323` ✅ |
| 结论 | ✅ Deep Match 真实 Hy3 跑通，Agent2=high / Agent3=low 全部确认 |

## JSON Repair

| 项目 | 结果 |
|---|---|
| 是否触发 | ⚠️ 未自然触发（LLM 输出合法 JSON，无需修复） |
| 单测 | ✅ `TestReasoningEffort_RepairReq_PassesNoThinkToClient` 覆盖 |
| 结论 | 通过（单测覆盖，不作为阻断） |

## 旧 TokenHub 兼容

| 项目 | 结果 |
|---|---|
| go build (deepseek) | ✅ |
| ReasoningEffort/ExtraBody 报错 | ❌ 否 |
| 切回 Hy3 | ✅ |

## 构建与测试

| 项目 | 结果 |
|---|---|
| go build `./server/...` | ✅ |
| go vet `./server/...` | ✅ |
| go test `./internal/agent/pipeline/...` | ✅（8 reasoning_effort 测试全部通过） |
| go test `./server/...` | ⚠️ 6 个历史无关失败 |

### 历史无关失败

| 测试 | 原因 | 与 Hy3 |
|---|---|---|
| TestMatchAgent_FinalV2Snapshot_EvidenceChain | mock 数据 | ❌ |
| TestMatchAgent_FinalV2Snapshot_TargetRoleEmptySmoke | mock 数据 | ❌ |
| TestIntegration_MatchAnalysis_Mock_Direct | mock 数据 | ❌ |
| TestIntegration_MatchAnalysis_Mock_Fallback_Metadata | mock 数据 | ❌ |
| TestMatchAnalysisAgentReturnsDecisionFieldsSmoke | mock 数据 | ❌ |
| TestDefaultConfig (llm) | TimeoutSeconds 默认值 | ❌ |
| TestLoader_List (prompt) | 文件列表 | ❌ |
| TestLoader_Load_Title + TestLoader_List (schema) | 文件列表 | ❌ |
| TestMatchService_MockAgent_StableResult | mock 数据 | ❌ |

## 修改文件

- 后端源码：无修改
- 配置：`.env`（临时切换 Hy3/deepseek，未写入 Key）
- docs：`docs/09-开发日志/2026-06-11-K14-Hy3-FullDeep-Regression.md`（新增）、`docs/下一步任务.md`、`docs/变更记录.md`（更新）
- 前端：无修改
- DB：无修改

## 风险

- P0：无
- P1：无
- P2：go test 全量有历史无关失败（9 个），不影响 Hy3 功能

## 下一步

- **K14-Hy3-Prod-Prepare**：确认生产环境 Hy3 Key 准备、更新 `.env.prod`、执行生产部署
