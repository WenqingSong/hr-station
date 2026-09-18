# 09 — 架构审查发现与建议

> checker（高级 Go 后端工程师）视角的客观审查结论。每个发现均标注证据来源。

## 一、架构优点（做得好的地方）

| # | 优点 | 证据 |
|---|------|------|
| 1 | **分层清晰，依赖单向向下，无循环依赖** | `internal/` 包结构：http → service → agent/repository → llm/model；未发现 repository 反向依赖 service |
| 2 | **LLM Provider 抽象干净** | `llm.Client` 接口 + `HunyuanClient`/`Hy3Client`/`NoopClient` 三实现，Provider 工厂 `newLLMClient` 单点切换 |
| 3 | **业务 Agent 依赖倒置** | 业务 Agent 依赖 `PipelineRunner` 接口（`match_agent.go:28`），而非具体 `pipeline.Runner` 或 `llm.Client` |
| 4 | **统一 Pipeline + 修复 + 兜底模式** | `pipeline.Runner.Run` 固化流程，鲁棒性高，LLM 失败自动转 Mock |
| 5 | **统一响应/错误码体系** | `response` 包 + `errors.AppError`，错误码分段（400xx/410xx/420xx/430xx） |
| 6 | **会话顶替（单点登录）设计完整** | `session.Manager` 的 `auth:session:*` + `auth:user_current:*` 双 key + `ErrSessionReplaced` |
| 7 | **安全脱敏意识强** | `LLMError` 不变量、`Config.APIKey` 不序列化、`redactSensitive` 多层脱敏 |

## 二、问题与风险（按严重度排序）

### 🔴 P1 — `main.go` 巨型组合根，职责过重

**证据**：`cmd/server/main.go` 约 280 行，不仅做依赖装配，还内联了业务逻辑：
- 启动两个后台回填 goroutine（TagEvidence 岗位画像回填、简历画像回填，`main.go:175-203`）。
- 手动拼接 `llm.Config` 并复制 fast/quality 两个 client。

**影响**：装配逻辑与业务启动逻辑耦合，难以测试；新增模块持续向 main.go 追加代码。

**建议**：
1. 抽取 `wire*` 函数（`wireAuth()`、`wireMatch()`、`wireRecommendation()` 等），main 只做编排。
2. 回填逻辑下沉到独立 service 方法，main 仅 `go service.StartBackfillJobs(...)`。
3. 可选引入 `google/wire` 做编译期依赖注入。

### 🔴 P2 — Service 构造函数参数过多（长参数列表坏味道）

**证据**：
- `NewMatchService` 12 个参数（`match_service.go`）
- `NewRecommendationService` 15 个参数（`recommendation_service.go`）
- `NewVersionCompareService` 12 个参数（`version_compare_service.go`）

**影响**：参数顺序易错、可读性差、参数变更波及所有调用方。

**建议**：引入依赖结构体（Options/Config struct）：
```go
type MatchServiceDeps struct {
    ResumeRepo  *repository.ResumeRepository
    JobRepo     *repository.JobRepository
    MatchRepo   *repository.MatchRepository
    MockAgent   *agent.MatchMockAgent
    FastRunner  agent.PipelineRunner
    QualityRunner agent.PipelineRunner
    Flags       AgentFlags          // llmEnabled/fallbackEnable/enableMockAgent 三合一
    TagEvidence *TagEvidenceService
    SemanticRepo *repository.SemanticProfileRepository
    Hy3Client   llm.Client
}
```

### 🟠 P3 — `service` 层与 `llm.Client` 直接耦合，且存在 Pipeline 重复实现

**证据**：
- `MatchService`、`RecommendationService`、`SemanticProfileProfiler` 直接持有 `llm.Client`（构造函数签名）。
- `match_deep_agent.go`（5 模块 Hy3）、`recommend_deep_cascade.go`（级联）直接拼 prompt + 调 `llm.Client` + 手动 JSON 提取/修复，与 `pipeline.Runner` 的功能重叠。

**影响**：
- 业务 Agent 走 Pipeline、Deep 路径走裸 `llm.Client`，两套 LLM 调用范式并存。
- JSON 提取/修复逻辑在 `service` 层重复出现（`match_deep_repair.go`），破坏单一职责。

**建议**：将 Deep 5 模块 / DeepCascade 的 LLM 调用收敛为受 `pipeline.Runner` 或独立的 LLM 执行器统一调度，`service` 层只做编排与业务映射。

### 🟠 P4 — SEP（语义画像）接入不一致

**证据**：
- `RecommendationService` 注入了 `SemanticProfileGate` + `SemanticProfileRepository`。
- `MatchService` 只注入了 `SemanticProfileRepository`（`match_service.go` 构造签名），**未注入 Gate**，SEP 画像读取走 Repository 直读，生成依赖外部回填。

**影响**：Fast Match 在 SEP 画像缺失时的行为依赖后台回填是否及时完成，与 Fast Recommend 的"门禁自动生成"语义不一致，可能导致行为差异。

**建议**：统一 SEP 接入策略——要么 Match 也走 Gate（同步/异步确保画像），要么明确 Match 的"只读 + 缺失回退标签引擎"语义并文档化。

### 🟡 P5 — `agent` 包职责过多，内聚度偏弱

**证据**：`internal/agent/` 单包同时包含：
- 业务 Agent（`match_agent.go`、`recommendation_agent.go` 等）
- Mock Agent（`*_mock.go`）
- Deep 3-agent（`*_deep_agents.go`）
- 快照工具（`agent_snapshot.go`）、mock 工具（`mock_helpers.go`）

**建议**：按职责拆分 `agent/business`、`agent/mock`、`agent/deep` 子包，保留 `PipelineRunner` 接口在包根。

### 🟡 P6 — 多套生成路径并存，链路复杂度高

**证据**：Fast 链路（SEP 引擎 / 标签引擎 / Mock）+ Quality 链路（Deep 3-agent / DeepCascade / 5 模块 Hy3 / Mock），同一业务域存在 3~4 套生成实现。

**影响**：新增/排查功能需理解多条链路；fallback 组合爆炸（`fallbackEnable` × `enableMockAgent` × `llmEnabled` × tier）。

**建议**：为每条链路建立统一的状态机/决策表（参考 `AgentStatus` + `GeneratedBy` 的已有雏形），并补充链路决策文档，减少隐性分支。

### 🟡 P7 — 三套错误码并存，映射成本高

**证据**：
- `errors.AppError`：int 码（40000/410xx/...）
- `llm.ErrorCode` + `pipeline` 错误码：string 码（`LLM_TIMEOUT`/`JSON_PARSE_FAILED`/...）
- `model.AgentResponseMeta`：又一套字符串状态（`fallback_reason`/`error_code`）

**影响**：跨层需多次映射（`mapLLMErrorCodeToPipeline`、`mapRepairErrorCode` 等），心智负担高。

**建议**：建立集中的错误码映射表，明确"业务码 vs LLM 码 vs 前端展示码"的边界，减少散落的 switch。

### 🟡 P8 — 向后兼容 shim 累积

**证据**：
- `model.Resume` 混入大量旧字段（`VersionNo`、`Location`、`SelfEvaluation`、`TargetRole` 等"兼容旧字段"）。
- `ResumeVersion = Resume` 别名、`VersionCompareResult = ResumeCompareResult` 别名（标记 deprecated）。
- `repository` 的 `FindByID`/`CreateVersion` 兼容包装（`resume_repository.go:495-513`）。

**建议**：制定 shim 清理计划，随版本迭代逐步移除，避免新旧模型长期并存。

## 三、优先级行动清单

| 优先级 | 事项 | 收益 | 成本 |
|--------|------|------|------|
| P1 | 拆分 `main.go`，抽取 wire 函数 | 可测试性、可维护性 | 中 |
| P1 | Service 构造函数改 Options 结构体 | 可读性、扩展性 | 中 |
| P2 | 收敛 Deep/SEP 的裸 `llm.Client` 调用 | 消除重复、统一范式 | 高 |
| P2 | 统一 SEP 接入策略 | 行为一致性 | 中 |
| P3 | 拆分 `agent` 包 | 内聚度 | 中 |
| P3 | 集中错误码映射 | 降低心智负担 | 低 |
| P3 | shim 清理计划 | 降低技术债 | 低 |

## 四、总体评价

该后端整体架构**分层合理、抽象清晰、鲁棒性设计到位**（统一 Pipeline、Provider 抽象、兜底机制、安全脱敏），是**结构健康的工程化 Go 项目**。主要改进空间集中在**组合根过重、Service 参数过长、多套 LLM 调用范式并存、多套生成链路**四类"规模化演进过程中累积"的问题，属典型的"从单体快速迭代到多业务域"阶段的架构债，尚未到需要重构骨架的程度。
