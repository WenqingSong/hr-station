# 开发日志 083 - MatchAnalysisAgent 接入 Pipeline

> 日期：2026-05-30
> 阶段：4-3.2
> 任务：MatchAnalysisAgent 接入 Pipeline

## 1. 任务目标

将匹配分析链路从纯 Mock Agent 升级为 LLM Pipeline 优先 + Mock fallback。具体包括：
- 创建 `MatchAnalyzer` 接口，Mock 和 LLM Agent 共享同一契约
- 实现 `MatchAnalysisAgent`，调用 Pipeline 的 `match_analysis` prompt/schema
- 组装 PromptValues（RESUME_CONTEXT_JSON, JOB_CONTEXT_JSON, BUSINESS_CONTEXT）
- 处理 Pipeline 的 success/repaired/fallback_required 三种结果
- 将 `map[string]any` Pipeline 输出映射为 `model.MatchAnalysisResult`
- Pipeline fallback_required 时调用原始 `match_mock.go`
- 支持 LLM_ENABLE=false 直接走 Mock 路径
- 编写单元测试（fake Pipeline，不调用真实 Hunyuan）

## 2. 修改范围

### 新增文件

| 文件 | 说明 |
|------|------|
| `server/internal/agent/match_agent.go` | MatchAnalysisAgent 实现 |
| `server/internal/agent/match_agent_test.go` | 28 项单元测试 |

### 修改文件

| 文件 | 变更 |
|------|------|
| `server/internal/model/match.go` | MatchAnalysisResult 新增 10 个 LLM-enriched 字段（omitempty） |
| `server/internal/service/match_service.go` | 重写：pipelineRunner + llmEnabled 注入，Analyze() 支持 LLM 优先 + Mock 回退 |
| `server/internal/agent/match_mock.go` | Analyze 签名新增 ctx 参数 |
| `server/internal/config/config.go` | LLMConfig 从 3 字段扩展到 14 字段 |
| `server/cmd/server/main.go` | LLM 基础设施初始化 + MatchService 依赖注入 |

### 不修改

- `.env` / `.env.example`：已有 14 项 LLM_* 配置
- Prompt 文件：`match_analysis_v1.md` 不修改
- Schema 文件：`match_analysis_schema.json` 不修改
- RecommendationAgent：不在本任务范围
- 前端代码：不修改
- 不进入 4-3.3

## 3. 关键设计决策

### 3.1 MatchAnalyzer 接口抽象

```go
type MatchAnalyzer interface {
    Analyze(ctx context.Context, resume *model.ResumeVersion, job *model.Job) (*model.MatchAnalysisResult, error)
}
```

Mock 和 LLM Agent 均实现此接口，便于后续替换和测试。

### 3.2 PipelineRunner 接口抽象

```go
type PipelineRunner interface {
    Run(ctx context.Context, req pipeline.RunRequest) *pipeline.RunResult
}
```

将 `MatchAnalysisAgent` 依赖从具体 `*pipeline.Runner` 改为接口，便于测试注入 fake runner。

### 3.3 MatchFallbackError 自定义错误

```go
type MatchFallbackError struct {
    FallbackReason string
    ErrorCode      string
    ErrorMessage   string
}
```

Pipeline 返回 fallback_required 或字段映射失败时，返回此错误。MatchService 捕获后回退 Mock Agent。

### 3.4 字段映射策略

LLM Schema 输出与 Mock Model 存在结构差异，映射函数处理：
- `evidence_chain.judgement`：matched→strong_match, partial→partial_match
- `screening_result`：likely→likely, uncertain→possible, risky→risky
- `matched_points`：提取 point 字符串 → match_highlights
- `resume_optimization.p0/p1/p2`：扁平化为 MatchSuggestion 列表
- 映射失败触发 MatchFallbackError → Mock 回退

### 3.5 LLM-enriched 字段使用 omitempty

新增字段全部使用 `json:"field_name,omitempty"`，不破坏现有 API 消费者。

### 3.6 config.LLMConfig 扩展

从 3 字段（BaseURL, APIKey, Model）扩展到 14 字段，与 `llm.Config` 对齐。新增 `getEnvAsFloat64` 辅助函数。

## 4. 验证方式

### 编译验证

```bash
cd server && go build ./...
# 通过，无错误
```

### 单元测试

```bash
cd server && go test ./internal/agent/... -v -count=1
# 28 项测试全部通过
```

测试覆盖：
- MatchAnalysisAgent.Analyze 成功（success/repaired/fallback_required/字段映射失败）
- 字段映射函数（mapScreeningResult, mapJudgement, mapEvidenceChain, mapMatchedPoints, mapGapPoints, mapRiskPoints, mapOptimizationSuggestions, mapInterviewQuestions, mapNextActions）
- 类型安全辅助函数（getString, getInt, getSlice）
- MatchFallbackError.Error()
- 接口合规性（MatchMockAgent/MatchAnalysisAgent 实现 MatchAnalyzer）
- PromptValues 组装
- 分数裁剪（0-100）
- Benchmark

### 全量测试

```bash
cd server && go test ./... -count=1
# 所有包测试通过，无回归
```

## 5. 未完成项

- 4-3.3 匹配分析接口联调与 fallback 验收（真实 LLM 调用验证）
- RecommendationAgent 接入 Pipeline（4-3.4）
- RunResult 字段策略统一落地（4-3.7）

## 6. 风险与注意事项

1. LLM_ENABLE=true 需要配置正确的 LLM_BASE_URL / LLM_API_KEY / LLM_MODEL
2. Pipeline 调用增加延迟（~2-5 秒），接口响应时间会变长
3. 新增字段 omitempty 不影响现有前端，但 4-3.3 联调时需确认
4. config.LLMConfig 扩展后，旧 .env 中没有的 LLM_* 变量会使用默认值
