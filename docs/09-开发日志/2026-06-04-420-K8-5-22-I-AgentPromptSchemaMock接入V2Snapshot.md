# K8-5-22-I — Agent / Prompt / Schema / Mock 接入 V2 Snapshot

## 性质
implementation（后端 Agent 层接入 ResumeSnapshotV2，不改前端/SQL）。

## 本阶段目标

让推荐/匹配/对比三个核心 Agent 统一使用 `NormalizeResumeToSnapshotV2` 生成结构化输入，
替代旧的逐字段手工拼装（target_role/target_city/skills_json 等）。Prompt key 从 RESUME_CONTEXT_JSON 改为 RESUME_SNAPSHOT_JSON。

## 修改文件

| 文件 | 修改内容 |
|------|----------|
| `server/internal/agent/agent_snapshot.go` | **新增**：agent-local normalizeResumeToSnapshot + snapshotToPromptMap + buildResumeSnapshotJSON + inline V1/V2 normalizers + JSON helpers（避免 agent→service 循环引用） |
| `server/internal/agent/recommendation_agent.go` | buildPromptValues 重构：用 buildResumeSnapshotJSON 替代手动拼装 DB 字段；prompts key RESUME_CONTEXT_JSON→RESUME_SNAPSHOT_JSON |
| `server/internal/agent/match_agent.go` | buildPromptValues 重构：同上 |
| `server/internal/agent/version_compare_agent.go` | buildPromptValues 重构：两个版本都用 buildResumeSnapshotJSON；删除 resumeToMap 函数 |
| `server/internal/agent/recommendation_agent_test.go` | 测试 key 更新：RESUME_CONTEXT_JSON→RESUME_SNAPSHOT_JSON |
| `server/internal/agent/match_agent_test.go` | 测试 key 更新：RESUME_CONTEXT_JSON→RESUME_SNAPSHOT_JSON |
| `server/internal/service/agent_snapshot_helper.go` | **新增**：共享 helper（供未来 service 层直接调用），包含 SnapshotToPromptMap + BuildResumeSnapshotJSON |

## 四核心功能接入确认

| 功能 | 是否使用 ResumeSnapshotV2 | 说明 |
|------|:---:|------|
| 岗位推荐 | ✅ | buildResumeSnapshotJSON(resume) → 11 模块结构化输入 |
| 岗位匹配分析 | ✅ | 同上 |
| 简历优化 | ⬜ 不变 | 优化 Agent 使用 free-text 输入（设计意图），不适合结构化 snapshot |
| 简历版本对比 | ✅ | 两个版本都用 buildResumeSnapshotJSON |

## Prompt key 变更

| 旧 key | 新 key | Agent |
|--------|--------|-------|
| `RESUME_CONTEXT_JSON` | `RESUME_SNAPSHOT_JSON` | Recommendation, Match |
| `OLD_RESUME_JSON` | (无变化) | Compare（内容改为 V2 snapshot） |
| `NEW_RESUME_JSON` | (无变化) | Compare（内容改为 V2 snapshot） |

## 旧字段依赖清理

- `resumeToMap` (target_role/target_city/skills 等) — 已删除 ✅
- recommendation_agent/buildPromptValues 不再直接引用 DB 字段 ✅
- match_agent/buildPromptValues 不再直接引用 DB 字段 ✅
- version_compare_agent/buildPromptValues 不再引用 DB 字段 ✅
- Prompt 文件中残留 `target_role_mismatch` 是 Schema 枚举值（业务逻辑），非 DB 字段 ✅

## 隐私字段确认

| 字段 | 是否进入 Prompt | 结果 |
|------|:---:|------|
| phone | ❌ | Snapshot 不含 phone |
| email | ❌ | Snapshot 不含 email |
| user_id | ❌ | Snapshot 不含 user_id |
| token | ❌ | 仅 pipeline 变量名 |

## 验收结果

| 验收项 | 结果 |
|--------|------|
| `go build ./...` | ✅ |
| `go vet ./...` | ✅ |
| `go test ./...` | ✅ 全部通过 |
| `npm run build` | ✅ 1.97s |
| API 3×200 + /api/resumes 401 | ✅ |
| Snapshot 接入 grep | ✅ agent_snapshot.go 10 处 + 各 agent 引用 |
| 旧字段清理 | ✅ agent/prompt 中无直接 DB 字段依赖 |
| 隐私字段 | ✅ 不进 Prompt |
| 前端未误改 | ✅ 前端 evidence_chain 为既有字段 |

## 未修改范围

- 未修改前端页面 ✅
- 未修改 MyResumesPage ✅
- 未修改 ResumeManagementPage / ResumeComparisonPage / AI 页面 ✅
- 未修改数据库 SQL ✅
- 未修改 V2 保存链路 ✅
- 未启用 schema_version=2 前端写入 ✅
- 未新增 API 路径 ✅
- 未修改 .env ✅
- 未安装依赖 / 未迁移目录 ✅

## 未执行 Git 命令

本次未执行任何 Git 命令，也未提交代码。

## 风险与边界

- Agent 层使用 agent-local 的 normalize 实现（避免循环引用 service），与 service 层 NormalizeResumeToSnapshotV2 逻辑一致
- 简历优化 Agent 保持 free-text 输入（设计如此），不在本阶段修改
- Mock Agent 未修改（Mock 输入为 agent 内部生成，不受此变更影响）
- V2 payload 启用和 V1/V2 全链路回归留给 K8-5-22-J
