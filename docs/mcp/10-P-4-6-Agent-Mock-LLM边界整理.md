# 10-P-4-6 Agent / Mock / LLM 边界整理

## 1. 阶段信息

* 阶段：10-P-4-6
* 日期：2026-06-02
* 范围：Agent 层边界清查、Mock/LLM 接口统一、fallback 策略冻结
* 是否修改代码：是（2 个文件，最小整理）
* 是否修改前端：否
* 是否修改后端：是（仅 agent 接口 + 常量使用）
* 是否修改数据库：否
* 是否接真实 LLM：否

## 2. 本次结论摘要

* Agent 层边界已经清晰：4 个 Agent（Recommendation / Match Analysis / Resume Optimization / Version Compare），3 个有 LLM+Mock，1 个仅 Mock
* 3 个 Agent 有统一 interface 定义（`Recommender` / `MatchAnalyzer` / `ResumeOptimizer`），编译时约束 Mock 和 LLM 实现
* Version Compare 缺少 `VersionComparer` 接口，本阶段补齐，使 4 个 Agent 全部接口化
* `AgentResponseMeta`（`model/agent_response.go`）已统一 `generated_by` / `agent_status` / `fallback_reason` / `error_code` / `repair_applied` 元数据
* `generated_by` 常量 `llm_agent` / `mock_agent` 已在 model 层定义，Service 层通过 `NewAgentResponseMeta()` 构建
* `LLM_ENABLE` 默认 `false`，LLM 当前关闭。Mock Agent 作为默认 Demo 路径
* Service 层统一调度：`llmEnabled && pipelineRunner != nil` → LLM 优先 → 失败 fallback Mock → LLM 关闭时直接用 Mock
* Fallback 策略明确：LLM 失败返回 `*XxxFallbackError`，Service 捕获后记录日志，调用 Mock Agent，metadata 标记 `agent_status="fallback"` / `generated_by="mock_agent"`
* Pipeline 层（`agent/pipeline/`）完整：Prompt Render → LLM Call → JSON Parse → Schema Validate → Repair（一次），每步失败有明确的 `FallbackReason` 和 `ErrorCode`
* `ENABLE_MOCK_AGENT` 配置项存在但未被任何代码引用（死配置），本阶段不处理
* 前端 `formatAgentSourceLabel` 将 `mock_success` 映射为"Mock 演示结果"（tone: gray），属前端展示行为，本阶段不修改，记录为已知
* go build / go vet / go test 全部通过（9 packages PASS）
* 未接真实 LLM、未修改前端、未改数据库
* 当前建议进入下一阶段（10-P-4-7 或用户指定优先级）

## 3. 执行前 Agent 链路清查

| Agent 链路 | 当前实现 | 输入 | 输出 | 当前问题 | 风险等级 |
| -------- | ---- | -- | -- | ---- | ---- |
| Recommendation | `RecommendationAgent` (LLM) + `RecommendationMockAgent` (Mock) | `*model.ResumeVersion` + `[]*model.Job` | `[]model.RecommendedJob` | Mock 使用关键词匹配，质量有限（非本阶段问题） | P2（质量） |
| Match Analysis | `MatchAnalysisAgent` (LLM) + `MatchMockAgent` (Mock) | `*model.ResumeVersion` + `*model.Job` | `*model.MatchAnalysisResult` | 同上 | P2（质量） |
| Resume Optimization | `ResumeOptimizationAgent` (LLM) + `ResumeOptimizationMockAgent` (Mock) | `*OptimizeInput`（resume_text/job_title/jd_text/target_section） | `*model.ResumeOptimizationResult` | 同上 | P2（质量） |
| Version Compare | `VersionCompareMockAgent`（仅 Mock，无 LLM） | `*model.ResumeVersion` ×2 | `*model.VersionCompareAnalysis` | 无 `VersionComparer` 接口；无 LLM 实现；`generated_by` 用字符串字面量 | P1（接口缺失） |

**接口覆盖：**

| Agent | 接口名 | 定义文件 | Mock 实现 | LLM 实现 |
| ----- | ------ | ----- | ------- | ------- |
| Recommendation | `Recommender` | `recommendation_agent.go:18` | `RecommendationMockAgent` | `RecommendationAgent` |
| Match Analysis | `MatchAnalyzer` | `match_agent.go:18` | `MatchMockAgent` | `MatchAnalysisAgent` |
| Resume Optimization | `ResumeOptimizer` | `resume_optimization_agent.go:17` | `ResumeOptimizationMockAgent` | `ResumeOptimizationAgent` |
| Version Compare | `VersionComparer` | **本阶段新增** → `version_compare_mock.go:14` | `VersionCompareMockAgent` | 无（待实现） |

## 4. Agent 边界设计冻结

| 层级 | 设计结论 | 说明 |
| -- | ---- | -- |
| Mock Agent | 规则引擎，基于真实 job/resume 输入做确定性分析。`generated_by="mock_agent"`，`agent_status="mock_success"` | 本地 Demo 默认路径，输出稳定可预测 |
| LLM Agent | Pipeline Runner 驱动：Prompt Render → LLM Call → JSON Parse → Schema Validate → Repair。`generated_by="llm_agent"`，`agent_status="success"/"repaired"` | LLM_ENABLE=true 时优先使用 |
| Service 调度层 | 每个 Service 持有 `mockAgent` + `pipelineRunner` + `llmEnabled`。`Analyze/Recommend/Optimize` 统一入口 | 调度逻辑集中，不散落在 handler |
| generated_by | `"llm_agent"` = LLM 生成（含 repair）；`"mock_agent"` = Mock 生成或 LLM fallback | 常量定义在 `model/agent_response.go:28-31` |
| provider | 当前未暴露到 API 响应。Pipeline 内部记录 `LLMProvider`（如"hunyuan"），仅在日志中可见 | 后续可加入 AgentResponseMeta |
| fallback | LLM 失败 → Service 捕获 `*XxxFallbackError` → 日志记录 → 调用 Mock → metadata 标记 `agent_status="fallback"` / `generated_by="mock_agent"` | 不允许静默 fallback |
| 错误处理 | LLM Agent 返回专用 `*XxxFallbackError`（含 FallbackReason + ErrorCode）；Handler 统一使用 `apperror.AsAppError()` | 10-P-4-2 模式已应用 |

## 5. 修改内容清单

| 文件 | 修改内容 | 是否业务代码 | 说明 |
| -- | ---- | ------ | -- |
| `server/internal/agent/version_compare_mock.go` | 新增 `VersionComparer` 接口；添加编译时接口检查 `var _ VersionComparer = (*VersionCompareMockAgent)(nil)` | 否（接口定义） | 补齐 Version Compare 接口，使 4 个 Agent 全部接口化，后续可新增 LLM 实现 |
| `server/internal/service/version_compare_service.go` | `GeneratedBy: "mock_agent"` → `GeneratedBy: model.GeneratedByMockAgent` | 否（常量替换） | 使用 model 层常量，消除魔法字符串 |

（仅 2 文件，+11 -1 行，最小整理）

## 6. generated_by / provider 设计

| 场景 | generated_by | provider | 说明 |
| -- | ------------ | -------- | -- |
| LLM_ENABLE=false（默认） | `mock_agent` | — | Mock Agent 直接调用，`agent_status="mock_success"` |
| LLM_ENABLE=true 且成功 | `llm_agent` | hunyuan(默认) | Pipeline 完整成功，`agent_status="success"` |
| LLM_ENABLE=true JSON 解析失败但 Repair 成功 | `llm_agent` | hunyuan(默认) | `agent_status="repaired"`，`repair_applied=true` |
| LLM_ENABLE=true 且失败 | `mock_agent` | —（Mock fallback） | `agent_status="fallback"`，`fallback_reason` 记录原因（e.g. `"llm_call_failed"`） |
| Mock 默认路径（所有 Agent） | `mock_agent` | — | `agent_status="mock_success"` |
| Version Compare | `mock_agent` | — | 仅 Mock 实现，无 LLM 路径。`generated_by` 已改用 `model.GeneratedByMockAgent` 常量 |

## 7. fallback 策略冻结

1. **LLM 关闭时是否允许 Mock**：允许。`LLM_ENABLE=false` 时 Service 直接调用 Mock Agent。这是默认行为。
2. **LLM 启用失败时是否 fallback**：允许且显式。Service 捕获 `*XxxFallbackError`，记录 `log.Printf("[XxxService] LLM pipeline fallback: %v", fallbackErr)`，调用 Mock Agent，metadata 标记 `agent_status="fallback"` / `generated_by="mock_agent"`。
3. **fallback 是否返回给前端**：是，但不可伪装。前端通过 `AgentResponseMeta` 可以知道生成来源。当前前端 `formatAgentSourceLabel` 对 `mock_success` 显示"Mock 演示结果"（gray tone），对 `fallback` 显示"兜底结果"（orange tone）。
4. **是否记录 fallback 来源**：是。后端日志输出 `[XxxService] LLM pipeline fallback: ...`，响应中 `meta.fallback_reason` + `meta.error_code` 标识具体原因。
5. **是否允许用户可见 Mock 文案**：当前前端展示，但建议后续阶段去技术化（将"Mock 演示结果"改为"规则分析结果"或"本地分析"等非技术文案）。本阶段不修改前端。

## 8. Agent 返回结构稳定性

| Agent | 返回结构 | 是否稳定 | 风险 | 结论 |
| ----- | ---- | ---- | -- | -- |
| Recommendation | `RecommendationResponse` → `{meta, result, recommended_jobs: []RecommendedJob}` | ✅ 稳定 | — | Mock 和 LLM 返回同结构 |
| Match Analysis | `MatchResponse` → `{meta, result, match_id}` | ✅ 稳定 | — | Mock 和 LLM 返回同结构 |
| Resume Optimization | `ResumeOptimizationResponse` → `{meta, result, optimization_id}` | ✅ 稳定 | — | Mock 和 LLM 返回同结构 |
| Version Compare | `VersionCompareResponse` → `{compare_id, generated_by, result, created_at}` | ✅ 稳定 | 无 `meta` 字段（仅 `generated_by` 字符串），与其他 3 个 Agent 不一致 | 低风险，后续 LLM 接入时统一 |

## 9. 测试补充

| 文件 | 测试内容 | 结果 |
| -- | ---- | -- |
| 无新增测试 | agent 层测试已覆盖：agent_pipeline_integration_test.go（LLM success/repaired/fallback/mock_direct/mock_fallback_metadata 5 场景），match_agent_test.go（接口实现），recommendation_agent_test.go，resume_optimization_agent_test.go，resume_optimization_mock_test.go，service/agent_response_test.go（metadata 构建），service/match_service_test.go（Service 调度 + MockAgent_StableResult） | ✅ 覆盖充分 |

**未新增测试原因**：现有 19 个测试文件已覆盖 Agent 主链路（Pipeline 集成、Mock/Metadata、Service 调度、接口实现）。本阶段仅新增 1 个接口定义 + 1 个常量替换，无新逻辑需要测试。

## 10. 验证结果

| 命令 | 结果 | 说明 |
| -- | -- | -- |
| `git status --short` 执行前 | 2 文件 modified | 仅 agent 接口 + service 常量 |
| `go build ./cmd/server` | ✅ PASS | 编译通过 |
| `go vet ./...` | ✅ PASS | 零告警 |
| `go test ./...` | ✅ PASS | 9 packages PASS（含 agent + agent/pipeline + agent/repair + service） |
| 反弹检查 | ✅ | 零 as any/mock/fallback 反弹；前端"Mock 演示结果"为预存行为，非本阶段引入 |
| `npm run build` | 未执行 | 本阶段未修改前端 |
| `npm run typecheck` | 未执行 | 本阶段未修改前端 |
| `git status --short` 执行后 | 2 文件 modified | 仅 agent 接口 + service 常量 |
| `git diff --stat` | +11 -1 行 | 最小整理 |

## 11. 未做事项

* 未接真实 LLM
* 未修改前端
* 未修改数据库 schema
* 未新增接口
* 未修改页面 UI
* 未重写 Prompt
* 未改变推荐 / 匹配 / 优化核心算法
* 未处理 `ENABLE_MOCK_AGENT` 死配置（非本阶段范围）
* 未处理前端 `formatAgentSourceLabel` "Mock 演示结果" 文案（前端修改非本阶段范围，记录为已知 P2）

## 12. 当前风险与后续建议

* **已解决**：4 个 Agent 全部接口化，`VersionComparer` 补齐
* **已解决**：`generated_by` 统一使用 model 常量（消除魔法字符串）
* **不需要处理，原因是**：`ENABLE_MOCK_AGENT` 配置死代码 — 非本阶段范围，不影响当前功能
* **需要下一阶段处理，阶段名是 10-P-4-7 或用户指定优先级**：前端 `formatAgentSourceLabel` 将 `mock_success` 展示为"Mock 演示结果" — 建议去技术化（如"规则分析结果"），但不阻塞本阶段通过
* **需要后续 LLM 验证阶段处理，阶段名是 11**：Version Compare 的 LLM Agent 实现；真实 LLM 调用验证

## 13. docs 更新清单

| 文件 | 新增 / 更新 | 目的 |
| -- | ------- | -- |
| `docs/mcp/项目上下文.md` | 更新 | 更新当前阶段为 10-P-4-6 |
| `docs/mcp/10-P-4-6-Agent-Mock-LLM边界整理.md` | 新增 | 本阶段 MCP 记录 |
| `docs/09-开发日志/2026-06-02-247-10-P-4-6-Agent边界整理.md` | 新增 | 开发日志 |

## 14. CodeBuddy 重启恢复摘要

10-P-4-6 完成 Agent / Mock / LLM 边界整理。4 个 Agent 全部审计并接口化：Recommendation / Match Analysis / Resume Optimization 已有完整 LLM+Mock 双实现 + 统一 interface + Pipeline + AgentResponseMeta；Version Compare 补齐 `VersionComparer` 接口。`generated_by` 常量统一到 `model.GeneratedByMockAgent`。`LLM_ENABLE=false` 默认关闭。Service 层统一调度：LLM 优先 → fallback Mock（显式记录日志+metadata）。Fallback 策略明确：agent_status 标记、reason/error_code 可追踪。Pipeline 完整（5 阶段 + 1 次 repair）。go build/vet/test 全部通过（9 packages PASS）。仅修改 2 个文件（+11 -1 行）。未接真实 LLM、未改前端。当前建议进入下一阶段（10-P-4-7 或用户指定优先级）。

下次继续前读取：`docs/mcp/项目上下文.md`、`docs/mcp/10-P-4-6-Agent-Mock-LLM边界整理.md`、`docs/mcp/10-P-4-0-后端现状清查与专业化问题分级.md`。

## 15. 最终判断

1. Agent / Mock / LLM 边界是否已整理？是。4 个 Agent 全部审计并接口化。
2. Recommendation / Match / Optimization / Version Compare 是否覆盖？是。4 个 Agent 全部覆盖。
3. Mock Agent 是否仍可作为默认 Demo 路径？是。`LLM_ENABLE=false` 默认，Mock Agent 作为默认实现。
4. 是否接入真实 LLM？否。
5. fallback 策略是否明确？是。LLM 失败 → `*XxxFallbackError` → Service 捕获 → 日志记录 → Mock fallback → metadata 标记。
6. generated_by / provider 是否明确？是。`model/agent_response.go` 定义常量，`AgentResponseMeta` builder 使用。
7. 前端是否仍不展示 Mock 演示标签？当前前端 `formatAgentSourceLabel` 将 `mock_success` 映射为"Mock 演示结果"。此为预存行为，非本阶段引入，记录为已知 P2。
8. 验证命令是否通过？是。go build/vet/test 全部通过（9 packages PASS）。
9. 是否修改前端？否。
10. 是否修改数据库？否。
11. 本阶段是否可以通过？是。
12. 是否建议进入下一阶段？是。
