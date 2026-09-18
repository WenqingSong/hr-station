# 02 — 包依赖关系图

基于 `import` 语句静态梳理的内部包依赖关系（仅内部包，不含第三方库）。

## 1. 内部包依赖图

```mermaid
graph LR
    %% 顶层入口
    MAIN["cmd/server"]
    AGENT["internal/agent"]
    PIPELINE["internal/agent/pipeline"]
    REPAIR["internal/agent/repair"]
    GUARD["internal/agent/guardrail"]
    CFG["internal/config"]
    DB["internal/database"]
    HEALTH["internal/health"]
    HTTP["internal/http"]
    LLM["internal/llm"]
    MW["internal/middleware"]
    MODEL["internal/model"]
    PROMPT["internal/prompt"]
    REPO["internal/repository"]
    RESP["internal/response"]
    SCHEMA["internal/schema"]
    SVC["internal/service"]
    SESS["internal/session"]
    LLMJSON["internal/llmjson"]
    ERR["internal/errors"]

    MAIN --> AGENT
    MAIN --> PIPELINE
    MAIN --> REPAIR
    MAIN --> CFG
    MAIN --> DB
    MAIN --> HEALTH
    MAIN --> HTTP
    MAIN --> LLM
    MAIN --> MW
    MAIN --> MODEL
    MAIN --> PROMPT
    MAIN --> REPO
    MAIN --> SCHEMA
    MAIN --> SVC
    MAIN --> SESS

    HTTP --> ERR
    HTTP --> HEALTH
    HTTP --> MW
    HTTP --> RESP
    HTTP --> SVC

    MW --> ERR
    MW --> RESP
    MW --> SVC
    MW --> SESS

    SVC --> AGENT
    SVC --> LLM
    SVC --> MODEL
    SVC --> REPO
    SVC --> PIPELINE
    SVC --> SCHEMA

    AGENT --> PIPELINE
    AGENT --> REPAIR
    AGENT --> GUARD
    AGENT --> MODEL

    PIPELINE --> REPAIR
    PIPELINE --> LLM
    PIPELINE --> LLMJSON
    PIPELINE --> PROMPT
    PIPELINE --> SCHEMA

    REPAIR --> LLM
    REPAIR --> LLMJSON
    REPAIR --> PROMPT
    REPAIR --> SCHEMA

    GUARD --> MODEL

    REPO --> MODEL
    REPO --> DB

    DB --> CFG

    SESS --> MODEL

    LLM --> CFG

    RESP --> ERR
```

## 2. 依赖矩阵（关键关系）

| 源包 | 依赖的内部包 | 备注 |
|------|-------------|------|
| `cmd/server` | 几乎全部 | Composition Root，唯一允许大面积依赖 |
| `http` | errors, health, middleware, response, service | Handler 只调 Service |
| `middleware` | errors, response, service, session | Auth 中间件依赖 AuthService + Session |
| `service` | agent, llm, model, repository, pipeline, schema | 核心编排层 |
| `agent` | pipeline, repair, guardrail, model | 业务 Agent |
| `agent/pipeline` | repair, llm, llmjson, prompt, schema | 统一管线 |
| `agent/repair` | llm, llmjson, prompt, schema | JSON 修复 |
| `agent/guardrail` | model | 本地护栏，仅操作 DTO |
| `repository` | model, database | 纯 SQL 层 |
| `database` | config | MySQL 连接封装 |
| `llm` | config | LLM 配置 |
| `session` | model | Redis 会话（间接 model 常量） |
| `model` | （无） | 纯类型，零内部依赖 |
| `errors` / `response` / `llmjson` / `schema` / `prompt` / `health` | （几乎无业务依赖） | 底层基础设施 |

## 3. 依赖方向总览（分层环检测）

```mermaid
graph TB
    direction TB
    L0["L0 纯类型 / 基础设施<br/>model · errors · response · llmjson · schema · prompt · health"]
    L1["L1 数据与外部适配<br/>config · database · session · llm"]
    L2["L2 数据访问<br/>repository"]
    L3["L3 Agent 运行时<br/>agent/pipeline · agent/repair · agent/guardrail"]
    L4["L4 业务 Agent<br/>agent"]
    L5["L5 业务服务<br/>service"]
    L6["L6 HTTP 接入<br/>middleware · http"]
    L7["L7 入口<br/>cmd/server"]

    L7 --> L6
    L6 --> L5
    L5 --> L4
    L5 --> L3
    L4 --> L3
    L5 --> L2
    L3 --> L1
    L3 --> L0
    L2 --> L1
    L2 --> L0
    L1 --> L0
```

**结论**：依赖严格单向向下，无循环依赖。`service` 同时依赖 `agent`（L4）与 `agent/pipeline`（L3），符合"服务层编排 Agent 与管线"的设计。

## 4. 依赖健康度观察

- ✅ 未发现 `repository` → `service` 或 `http` → `repository` 的越层依赖。
- ✅ 业务 Agent 依赖 `PipelineRunner` 接口（定义于 `agent` 包），而非具体 `pipeline.Runner` 结构体，依赖倒置合理。
- ⚠️ `service` 包文件数量多（37 个），且 `match_service.go` / `recommendation_service.go` 直接持有 `llm.Client`，使 `service` 与 `llm` 强耦合（仅 Deep/SEP 路径需要，可考虑收敛到 Agent 层）。
- ⚠️ `agent` 包同时包含 `mock_helpers.go`、`agent_snapshot.go`、`compare_deep_agents.go` 等多种职责，包内聚度偏弱（详见 09-findings）。
