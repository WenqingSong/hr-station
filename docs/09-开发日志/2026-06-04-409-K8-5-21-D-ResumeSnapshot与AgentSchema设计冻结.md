# K8-5-21-D — ResumeSnapshot 与 Agent Schema 设计冻结

> 日期：2026-06-04 | 性质：design freeze | 未修改源码

---

## 1. 本阶段范围

只做 ResumeSnapshot / Agent Schema 设计冻结，不改源码。冻结 12 项设计决策。

---

## 2. 读取文件

### 设计文档
- `/workspace/docs/设计文档/K8-5-21-B-新版简历字段体系与模块设计冻结.md`
- `/workspace/docs/设计文档/K8-5-21-C-数据库与后端API设计冻结.md`
- `/workspace/docs/09-开发日志/2026-06-03-406-K8-5-21-A-当前简历数据链路现状盘点.md`

### Agent 代码
- `server/internal/agent/recommendation_agent.go` — `buildPromptValues` 直接用旧字段
- `server/internal/agent/match_agent.go` — 同上
- `server/internal/agent/version_compare_agent.go` — `resumeToMap` 直接用旧字段
- `server/internal/agent/resume_optimization_agent.go` — 文本模式 `buildPromptValues`
- `server/internal/agent/resume_optimization_mock.go` — 规则匹配
- `server/internal/agent/version_compare_mock.go` — 旧字段 score 计算

### Prompt 文件
- `server/prompts/recommendation_v1.md`
- `server/prompts/match_analysis_v1.md`
- `server/prompts/resume_optimization_v1.md`
- `server/prompts/resume_optimization_core_v1.md`
- `server/prompts/version_compare_v1.md`

### Schema 文件
- `server/prompts/schema/recommendation_schema.json`
- `server/prompts/schema/match_analysis_schema.json`
- `server/prompts/schema/resume_optimization_schema.json`
- `server/prompts/schema/version_compare_schema.json`

### 模型与前端
- `server/internal/model/resume.go` — `ResumeVersion` struct
- `web/src/api/agents.ts` — Agent API 类型
- `web/src/types/agent.ts` — 前端 Agent 类型
- `web/src/types/resume.ts` — 前端简历类型
- `web/src/adapters/agentAdapter.ts` — Agent ViewModel

---

## 3. 核心设计结论

| 决策 | 冻结 |
|------|------|
| **统一 Snapshot** | 所有 Agent 统一接收 `ResumeSnapshotV2`（11 模块：meta + 10 页面模块） |
| **Normalize 层** | `NormalizeResumeToSnapshot()` 在后端 service 层，V1/V2 都输出统一 V2 |
| **Prompt 隔离** | Prompt 不感知 DB 字段名（`target_role`/`experience_json`），只看到 Snapshot |
| **隐私字段不进入 Agent** | phone/email/user_id/token/raw_text 不进 Agent |
| **Prompt 差异化裁剪** | 不同功能选择需要的模块，空模块标记"未填写" |
| **Response evidence 归属** | 实现阶段新增 `resume_module` 字段（可选），不改现有 Schema |
| **Mock 扩展** | 10 个新场景覆盖证书/获奖/AI技能/作品链接/空缺 |
| **V1→V2 映射** | 24 个字段级映射 + 8 个新模块默认空值 |

---

## 4. 发现的关键问题

| 问题 | 级别 | 说明 |
|------|:---:|------|
| 无统一 ResumeSnapshot | P0 | 四个 Agent 各自 patch 旧字段，结构不一致 |
| Prompt 感知 DB 字段名 | P0 | `target_role`/`skills_json` 等旧名在 Prompt 中硬编码 |
| Mock 仅覆盖旧字段 | P1 | 证书/获奖/AI技能/作品链接无 Mock |
| 简历优化用文本模式 | P2 | 非结构化，无法按模块精度优化 |
| Version Compare 仅对比 V1 字段 | P2 | 对比报告不完整 |

---

## 5. 构建 / API / 测试结果

| 验收项 | 结果 |
|--------|------|
| `npm run build` | ✅ 通过 (1.91s) |
| `GET /api/jobs?page=1&page_size=5` | ✅ 200 |
| `GET /api/jobs/10001` | ✅ 200 |
| `GET /api/jobs/facets` | ✅ 200 |
| `GET /api/resumes` | ✅ 401（后端保护正常） |
| `go test ./...` | ✅ 全部通过 |

---

## 6. 未修改范围确认

```text
未修改源码/样式/前端页面/后端/数据库/SQL/接口/.env/LLM 配置/Prompt/Schema
未新增字段/接口/Agent/Mock
未安装依赖/未迁移目录
```

---

## 7. Git 约束确认

```text
本次未执行任何 Git 命令，也未提交代码。
```
