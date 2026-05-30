# 2026-05-31-110-PromptSchemaGo契约最小一致性补齐

## 1. 本次任务

4-3.7-fix.2：Prompt / Schema / Go 契约最小一致性补齐。

补齐 MatchAnalysis / Recommendation 的 Go model、agent mapping、mock/fallback/fixture 中缺失的 apply_readiness / optimization_stop_signal / top_fix_items 字段，使其与 Prompt / Schema 保持一致。

## 2. 背景

fix.1 已完成状态归一，确认 C.1 为 partial_pass：
- Schema / Prompt 已完整（3 个 Schema + 3 个 Prompt 均已包含新字段）
- ResumeOptimization Go model 已同步
- MatchAnalysis / Recommendation Go model + agent mapping 缺失

本次 fix.2 补齐这一缺口。

## 3. 修改范围

### Go model
- `server/internal/model/match.go`：新增 `TopFixItem` 结构体；`MatchAnalysisResult` 新增 3 字段
- `server/internal/model/recommendation.go`：`RecommendedJob` 新增 3 字段

### Agent mapping
- `server/internal/agent/match_agent.go`：新增 `mapTopFixItems` + 三字段映射
- `server/internal/agent/recommendation_agent.go`：新增 `mapRecTopFixItems` + 三字段映射

### Mock / fallback
- `server/internal/agent/match_mock.go`：新增 `determineMatchApplyReadiness` / `generateMatchTopFixItems`
- `server/internal/agent/recommendation_mock.go`：新增 `determineRecApplyReadiness` / `generateRecTopFixItems`

### 测试 fixture
- `server/internal/agent/match_agent_test.go`：`minimalLLMOutput` 新增 3 字段
- `server/internal/agent/recommendation_agent_test.go`：测试 fixture 新增 3 字段

### 文档
- `docs/项目状态.md`
- `docs/下一步任务.md`
- `docs/变更记录.md`
- `docs/说明书.md`
- `docs/05-Agent设计/PromptSchema战略一致性核验与改动评审.md`

## 4. MatchAnalysis 契约补齐

### Model
```go
// model.TopFixItem（新增，Match/Recommendation 共用）
type TopFixItem struct {
    Issue            string `json:"issue"`
    WhyItMatters     string `json:"why_it_matters"`
    Action           string `json:"action"`
    ExampleReference string `json:"example_reference"`
    SelfCheckQuestion string `json:"self_check_question"`
    EstimatedEffort  string `json:"estimated_effort"`
}

// MatchAnalysisResult 新增字段
ApplyReadiness         string       `json:"apply_readiness,omitempty"`
OptimizationStopSignal string       `json:"optimization_stop_signal,omitempty"`
TopFixItems            []TopFixItem `json:"top_fix_items,omitempty"`
```

### Agent mapping
- `mapParsedToMatchResult`：从 LLM parsed 中提取 `apply_readiness`（默认 `insufficient_information`）、`optimization_stop_signal`（默认 `needs_more_evidence`）、`top_fix_items`
- `mapTopFixItems`：将 Schema 的 top_fix_item 结构映射为 `[]model.TopFixItem`

### Mock / fallback
- `determineMatchApplyReadiness`：基于 score 和 data_quality 判断
  - score >= 85 → `ready_to_apply` / `stop_and_apply`
  - score >= 60 → `quick_fix_before_apply` / `fix_top_3_then_apply`
  - summary_only → `insufficient_information` / `needs_more_evidence`
  - 其他 → `not_ready_for_this_role` / `needs_more_evidence`
- `generateMatchTopFixItems`：生成 1-3 个快速修复项

## 5. Recommendation 契约补齐

### Model
```go
// RecommendedJob 新增字段（per job）
ApplyReadiness         string       `json:"apply_readiness,omitempty"`
OptimizationStopSignal string       `json:"optimization_stop_signal,omitempty"`
TopFixItems            []TopFixItem `json:"top_fix_items,omitempty"`
```

### Agent mapping
- `mapParsedToRecommendedJobs`：从每个 recommendation 中提取三字段
- `mapRecTopFixItems`：将 Schema 的 top_fix_item 映射为 `[]model.TopFixItem`

### Mock / fallback
- `determineRecApplyReadiness`：基于 score 和 data_quality 判断
- `generateRecTopFixItems`：`ready_to_apply` 返回空数组，否则生成 1-3 项

## 6. ResumeOptimization 核验

ResumeOptimization 已在 C.1 阶段完成同步，本次仅核验：
- `ApplyReadiness` ✅ 已存在
- `OptimizationStopSignal` ✅ 已存在
- `TopFixItems` ✅ 已存在（使用 `ResumeOptimizationFixItem`，多了 `learn_more_query` 字段）
- Mock / fallback ✅ 已有合理默认值
- 测试 ✅ 已通过

无需重复修改。

## 7. Mock / fallback / fixture 同步

| 组件 | 同步状态 |
|------|---------|
| MatchMockAgent | ✅ 已同步 |
| RecommendationMockAgent | ✅ 已同步 |
| MatchAnalysisAgent LLM mapping | ✅ 已同步 |
| RecommendationAgent LLM mapping | ✅ 已同步 |
| match_agent_test.go fixture | ✅ 已同步 |
| recommendation_agent_test.go fixture | ✅ 已同步 |
| pipeline_test.go fixture | ✅ 已有（ResumeOptimization 场景） |
| schema/validator_test.go fixture | ✅ 已有 |

## 8. 测试结果

- `go build ./...`：编译通过 ✅
- `go test ./...`：全部通过 ✅
- `make llm-check`：通过（未调用真实 LLM）✅
- Schema JSON 校验：三个 Schema 均合法 ✅
- 字段覆盖检查：`apply_readiness` / `optimization_stop_signal` / `top_fix_items` 在 Prompt / Schema / Go model / agent mapping / mock / test 中全覆盖 ✅

## 9. 未修改内容

- 未修改前端
- 未修改 Makefile
- 未修改 docker-compose
- 未修改 Nginx
- 未修改 Prompt / Schema 主体设计
- 未调用真实 LLM
- 未请求腾讯招聘接口
- 未执行 reset / reset-up
- 未清空 volume

## 10. 当前限制

- TopFixItem 与 ResumeOptimizationFixItem 是两个独立结构体（后者多了 learn_more_query），未来可考虑统一
- apply_readiness / optimization_stop_signal 的枚举校验目前在 Prompt 层面约束，Go 代码中未做严格枚举校验
- Quality Gate / 极端情况字段 / ResumeContextText 输入层改造留待后续阶段

## 11. 是否可以进入 fix.3

可以。C.1 已从 partial_pass 升级为 pass，Go 契约一致性补齐完成。

下一步进入 4-3.7-fix.3 最终闸门验收，确认所有修复稳定后进入 4-3.8。

## 12. 下一步建议

进入 4-3.7-fix.3：进入 4-3.8 前最终闸门验收。
