# Offer 捕手后端架构 UML 总览

> 本目录由 **checker（高级 Go 后端工程师）** 产出，用于沉淀后端架构文档，供其他 agent 协同。
> 范围：`/workspace/server`（Go 后端）。仅做只读分析，未修改任何代码、未执行任何 git 操作。

## 文档目录

| 文件 | 内容 | 图类型 |
|------|------|--------|
| [01-architecture-overview.md](./01-architecture-overview.md) | 架构分层、组件总览、分层依赖规则 | 分层组件图 |
| [02-package-dependencies.md](./02-package-dependencies.md) | 包（package）级别依赖关系 | 包依赖图 |
| [03-model-er.md](./03-model-er.md) | 数据模型实体及关系 | ER 图 |
| [04-service-layer.md](./04-service-layer.md) | Service 业务层类图与依赖 | 类图 |
| [05-agent-pipeline.md](./05-agent-pipeline.md) | Agent / Pipeline / Repair / Guardrail 层 | 类图 + 流程图 |
| [06-llm-layer.md](./06-llm-layer.md) | LLM 客户端抽象与 Provider 实现 | 类图 |
| [07-http-layer.md](./07-http-layer.md) | HTTP 路由、Handler、中间件 | 类图 + 路由表 |
| [08-sequences.md](./08-sequences.md) | 关键业务时序（请求链路 / Pipeline / 分级） | 时序图 |
| [09-findings.md](./09-findings.md) | 架构审查发现与改进建议 | 文本 |

## 一图速览（核心分层）

```
┌─────────────────────────────────────────────────────────────┐
│  cmd/server/main.go          Composition Root（依赖组装）     │
└──────────────────────────────┬──────────────────────────────┘
                               │ 组装所有依赖
┌──────────────────────────────▼──────────────────────────────┐
│  internal/http               HTTP 路由 + Handler（REST API） │
│  internal/middleware         Recover / CORS / Logger / Auth │
└──────────────────────────────┬──────────────────────────────┘
                               │ 调用
┌──────────────────────────────▼──────────────────────────────┐
│  internal/service            业务逻辑层（8 大业务模块）       │
│    Auth / Resume / ResumeProfile / Job / Match /            │
│    Recommendation / VersionCompare / Optimization /        │
│    TagEvidence / SemanticProfile（SEP）                     │
└──────┬───────────────────────────────┬──────────────────────┘
       │ 规则/编排                      │ 数据访问
┌──────▼──────────────────────┐  ┌──────▼─────────────────────┐
│  internal/agent             │  │  internal/repository        │
│   业务 Agent + Pipeline     │  │   10 个实体 Repository      │
│   + Repair + Guardrail      │  └──────┬─────────────────────┘
│   + MockAgent（规则兜底）    │         │ SQL
└──────┬──────────────────────┘  ┌──────▼─────────────────────┐
       │ 调用                     │  internal/model             │
┌──────▼──────────────────────┐  │   实体 + DTO（纯类型）       │
│  internal/llm               │  └────────────────────────────┘
│   Client 接口 + Hy3/Hunyuan │
└─────────────────────────────┘
```

## 关键技术特征（摘要）

1. **分层架构**：`http → service → repository → model/database`，依赖方向自上而下，无反向依赖。
2. **Agent 双轨制**：LLM Agent（`llm_agent`）与规则 Mock Agent（`mock_agent`）并行，LLM 失败自动 `fallback` 到 Mock。
3. **统一 Pipeline**：所有 LLM 业务 Agent 复用 `pipeline.Runner`，执行 `Prompt 渲染 → LLM 调用 → JSON 抽取/解析 → Schema 校验 → 失败修复一次 → 兜底`。
4. **Provider 抽象**：`llm.Client` 接口隔离混元（Hunyuan/OpenAI 兼容）与 Hy3（腾讯云 LKEAP）两种后端。
5. **分级生成（Generation Tier）**：`fast`（规则/标签引擎/SEP）与 `quality`（Deep 3-agent / 级联 LLM）双档。
6. **多态语义画像**：`SemanticProfile` 用 `TargetType + TargetID` 多态指向 Resume/Job。
7. **会话管理**：JWT（无状态身份）+ Redis（有状态会话，支持单点登录顶替）。
8. **统一响应**：`response` 包 + `errors.AppError` 错误码体系。

## 面向其他 agent 的约定

- 所有图均使用 **Mermaid** 语法，可在支持 Mermaid 的 Markdown 渲染器中直接查看。
- 每个文件自包含，可独立阅读。
- 命名遵循 Go 包名与真实类型名，便于与源码一一对应。
