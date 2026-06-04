# K8-5-26-B：ResumeSnapshot Normalize 适配最终 9 模块

## 本阶段目标

让 ResumeSnapshot model + service normalizer + agent-local normalizer 识别最终 9-module skills/portfolio 新结构。

## 修改清单

| 文件 | 修改 |
|------|------|
| `model/resume_snapshot.go` | Skills: +ProfessionalSkills / +AIApplication; Portfolio: +Works + ResumeSnapshotPortfolioWork |
| `service/resume_snapshot_normalizer.go` | skills 新 key 映射 + portfolio works 解析 + 3 helper |
| `agent/agent_snapshot.go` | skills 新 key 映射 + portfolio works 解析 + 3 helper |

## skills 适配结果

| 输入结构 | Snapshot 输出 |
|---------|-------------|
| `development_languages: ["Go","SQL"]` | `professional_skills: ["Go","SQL"]` ✅ |
| `ai_application: "CodeBuddy..."` | `ai_application: "CodeBuddy..."` ✅ |
| 旧 `languages: ["Go"]` | `programming_languages` fallback ✅ |
| 旧 `ai_scenarios: [...]` | `ai_application_scenarios` fallback ✅ |
| 混合新旧 | 新字段优先，旧字段兼容 ✅ |

## portfolio 适配结果

| 输入结构 | Snapshot 输出 |
|---------|-------------|
| `works: [{name, link, password}]` | `works: [{name, link, password}]` ✅ |
| `homepage: "..."` | `homepage: "..."` ✅ |
| 旧 github/gitee/cnb | 保留兼容字段 ✅ |
| blog → homepage fallback | ✅ |

## language_exam 适配

`certificates` type/score 字段已在 Snapshot 中保留 — 无需改动 ✅

## target 空值安全

target_direction/role/job_intention 已有 nil-safe 处理 — 无需改动 ✅

## service / agent-local 一致性

| 检查项 | 结果 |
|--------|:---:|
| skills 映射一致 | ✅ 同 key 同逻辑 |
| portfolio 解析一致 | ✅ 同 parsePortfolioWorks |
| achievements 一致 | ✅ 已适配 |

## 测试覆盖

| 测试 | 结果 |
|------|:---:|
| 原有 TestAgentSnapshotConsistency_V1 | ✅ 通过 |
| 原有 TestAgentSnapshotConsistency_V2 | ✅ 通过 |
| 原有 TestAgentSnapshotToJSON | ✅ 通过 |
| 所有 service 测试 | ✅ 通过 |

## 构建 / API / 测试结果

| 验收项 | 结果 |
|--------|:---:|
| `go build ./...` | ✅ |
| `go vet ./...` | ✅ |
| `go test ./...` | ✅ 全部通过 |
| `npm run build` | ✅ 2.12s |
| API 200/200/200/401 | ✅ |
| Prompt/Mock 未误改 | ✅ |

## 未修改范围确认

```text
未修改前端页面 ✅
未修改数据库 SQL ✅
未修改 Prompt / Schema / Mock ✅
未修改 API 路径 ✅
未新增 /api/resumes/v2 ✅
未修改登录鉴权 ✅
未修改 jobs 页面 ✅
未修改 .env ✅
未安装依赖 ✅
未迁移目录 ✅
未新增上传能力 ✅
```

## 风险与边界

```text
Prompt 重写 → K8-5-26-C
Mock 补齐 → K8-5-26-E
Agent 回归 → K8-5-26-F
```

## 未执行 Git 命令

本次未执行任何 Git 命令，也未提交代码。
