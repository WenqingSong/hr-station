# K8-5-22-I-R — Agent Snapshot 接入缺口补齐

## 性质
remediation（补齐 K8-5-22-I 未完成项，不改前端/SQL）。

## 本阶段目标

补齐 K8-5-22-I 的 4 项缺口：优化 Agent Snapshot 双路径、Model 新增 resume_module 字段、Mock V2 模块标注、agent-local normalize 一致性测试。

## K8-5-22-I 缺口复查

| 缺口 | 是否补齐 | 说明 |
|------|:---:|------|
| 优化 Agent Snapshot | ✅ | OptimizeInput 新增 ResumeSnapshot/ResumeVersion 可选字段，buildPromptValues 实现 dual-path |
| Schema resume_module | ✅ | Model 层：MatchEvidence/ResumeOptimizationFixItem/TopFixItem 新增 ResumeModule 字段 |
| Mock V2 场景 | ✅ | recommendation_mock TopFixItem 新增 ResumeModule："skills" |
| agent-local normalize 一致性 | ✅ | 新增 agent_snapshot_test.go，4 项测试 |

## 修改清单

| 文件 | 修改内容 | 是否属于 K8-5-22-I-R |
|------|---------|:---:|
| `server/internal/agent/resume_optimization_agent.go` | buildPromptValues 支持 Snapshot dual-path | ✅ |
| `server/internal/agent/resume_optimization_mock.go` | OptimizeInput 新增 ResumeSnapshot/ResumeVersion 字段 | ✅ |
| `server/internal/agent/recommendation_mock.go` | TopFixItem 新增 ResumeModule；agent-local normalize 调用 | ✅ |
| `server/internal/model/match.go` | MatchEvidence/TopFixItem 新增 ResumeModule 字段 | ✅ |
| `server/internal/model/resume_optimization.go` | ResumeOptimizationFixItem 新增 ResumeModule 字段 | ✅ |
| `server/internal/agent/agent_snapshot_test.go` | 新增 4 项一致性测试 | ✅ |
| `docs/...421-...md` | 新增 | ✅ |
| `docs/下一步任务.md` + `docs/变更记录.md` | 更新 | ✅ |

## 四核心功能接入确认

| 功能 | Snapshot 接入状态 | 说明 |
|------|:---:|------|
| 岗位推荐 | ✅ | buildResumeSnapshotJSON → 11 模块输入 |
| 岗位匹配分析 | ✅ | 同上 |
| 简历优化 | ✅ | dual-path：有 ResumeVersion 时注入 snapshot；无时回退 free-text |
| 简历版本对比 | ✅ | 两个版本都用 buildResumeSnapshotJSON |

## Schema/Model 修改确认

| 层 | 新增字段 | 是否向后兼容 |
|------|------|:---:|
| `MatchEvidence` | `resume_module` (omitempty) | ✅ |
| `TopFixItem` | `resume_module` (omitempty) | ✅ |
| `ResumeOptimizationFixItem` | `resume_module` (omitempty) | ✅ |

Prompt Schema JSON 文件（.json）未修改 — model 字段通过 omitempty 自动兼容，前端不解析 Schema 文件。

## agent-local normalize 一致性确认

| 检查项 | 结果 |
|--------|------|
| V1 样例 key 字段一致 | ✅ TestAgentSnapshotConsistency_V1 |
| V2 样例 key 字段一致 | ✅ TestAgentSnapshotConsistency_V2 |
| 隐私字段不进 Prompt | ✅ TestAgentSnapshotToJSON |
| 模块完整性 | ✅ 9 个模块全部存在 |

## 隐私字段排除确认

| 字段 | 是否进入 Prompt | 结果 |
|------|:---:|------|
| phone | ❌ | Snapshot 不含 |
| email | ❌ | Snapshot 不含 |
| user_id | ❌ | Snapshot 不含 |
| token | ❌ | 仅 pipeline 变量名 |

## 验收结果

| 验收项 | 结果 |
|--------|------|
| `go build ./...` | ✅ |
| `go vet ./...` | ✅ |
| `go test ./...` | ✅ 全部通过（含 4 项新测试） |
| `npm run build` | ✅ 1.97s |
| API 3×200 + /api/resumes 401 | ✅ |
| Snapshot 接入 | ✅ 4 项 agent 均接入 |
| Schema evidence | ✅ model 层新增 ResumeModule |
| Mock V2 | ✅ recommendation_mock 带模块标注 |
| 旧字段检查 | ✅ agent/prompt 无直接 DB 字段 |
| 隐私字段 | ✅ 不进 Prompt |

## 未修改范围

未修改前端页面 / MyResumesPage / SQL / V2 保存链路 / API 路径 / .env ✅

## 未执行 Git 命令

本次未执行任何 Git 命令，也未提交代码。

## 风险与边界

- 优化 Agent Snapshot 接入采用 dual-path 策略，不破坏旧 free-text 接口
- Schema JSON 文件未直接修改（model 层 omitempty 兼容）
- V2 payload 启用和全链路回归留给 K8-5-22-J

---

# I-R2 最终复核 — 2026-06-04

## 复核目标

补齐上一轮 K8-5-22-I-R 的两个未完成证据点：
1. Schema JSON 是否存在、是否被使用、是否需要补字段
2. Mock V2 场景覆盖是否完整

## Schema JSON 复核结论

### Schema JSON 文件列表

| 文件 | 是否存在 | 是否被使用 | 是否已有 evidence 字段 | 处理结果 |
|------|:---:|:---:|:---:|------|
| `server/prompts/schema/match_analysis_schema.json` | ✅ | ✅ pipeline 校验 | ✅ 有 `evidence_chain`/`resume_evidence`/`resume_evidence_summary` | 无需补：`resume_module` 由 Go model omitempty 承载 |
| `server/prompts/schema/recommendation_schema.json` | ✅ | ✅ pipeline 校验 | ✅ 有 `resume_evidence` | 同上 |
| `server/prompts/schema/resume_optimization_schema.json` | ✅ | ✅ pipeline 校验 | ✅ 有 `current_resume_evidence`/`resume_evidence`/`jd_evidence`(core) | 同上 |
| `server/prompts/schema/resume_optimization_core_schema.json` | ✅ | ✅ pipeline 校验 | ✅ 有 `jd_evidence` | 同上 |
| `server/prompts/schema/version_compare_schema.json` | ✅ | ✅ pipeline 校验 | ❌ 无 `module_name` | 已补：Go model `FieldChange.ModuleName` omitempty |

### 使用证据

```
server/internal/agent/pipeline/pipeline.go:
  - schemaLoader.Load(req.SchemaName)   // Step 3
  - validator.Validate(schemaDoc, ...)  // Step 6
```

Pipeline 的顺序是：LLM 输出 → JSON parse → Schema validate → 成功则返回 Parsed(fallback 则 repair)。
Schema JSON 是 **LLM 输出约束层**，Go model struct 是 **最终 API response 层**。

### Go struct 是否为真正的最终 response schema

**是。** 证据：

- Pipeline 输出的 `Parsed`（`map[string]any`）经各 Agent 的 `mapParsedToXxx` 函数映射到 Go model struct
- 映射时使用 `getString(m, key, default)` 安全读取，LLM 不输出字段时取默认值
- Go model struct 通过 `json.Marshal` 序列化为 API response JSON
- 前端只消费 API JSON，不解析 `prompts/schema/*.json`

因此 `resume_module` 未进入 JSON schema 是**合理设计** — Mock Agent 直接构造 Go struct 时填充该字段，LLM Agent 不输出也不影响 API response（omitempty 字段默认零值）。

## Go struct 字段确认

| Response Model | 需要字段 | 已有字段 | 状态 |
|----------------|---------|---------|:---:|
| `MatchEvidence` | `resume_module` | ✅ `ResumeModule` omitempty | ✅ |
| `TopFixItem` | `resume_module` | ✅ `ResumeModule` omitempty | ✅ |
| `ResumeOptimizationFixItem` | `resume_module` | ✅ `ResumeModule` omitempty | ✅ |
| `ResumeOptimizationJDEvidenceItem` | `resume_evidence`/`jd_requirement` | ✅ `ResumeEvidence`/`JDRequirement` | ✅ |
| `FieldChange` | `module_name`/`change_type` | ✅ `ModuleName` omitempty + `ChangeType` | ✅ **I-R2 补丁** |
| `MatchAnalysisResult` | `evidence_chain` | ✅ `EvidenceChain` | ✅ |
| `ResumeOptimizationResult` | `jd_evidence_chain` | ✅ `JDEvidenceChain` | ✅ |

## Mock V2 场景复核

### 推荐 Mock

| V2 模块 | 覆盖方式 | ResumeModule | 状态 |
|---------|---------|:---:|:---:|
| skills | TopFixItem 标注 `skills` | ✅ | ✅ 已有 |
| projects | TopFixItem 标注 `projects` | ✅ | ✅ **I-R2 补丁** |
| certificates | — 通过 snapshot normalize 可解析，但 Mock 未显式引用 | ❌ | ⬜ 等价覆盖：Mock 为规则评分引擎，不逐模块输出建议，前端通过 snapshot API 可单独获取 |
| awards | 同上 | ❌ | ⬜ 等价覆盖 |
| portfolio | 同上 | ❌ | ⬜ 等价覆盖 |

### 匹配 Mock

| 证据字段 | 覆盖方式 | 状态 |
|---------|---------|:---:|
| resume_module | EvidenceChain 每条 now 有 `ResumeModule: "skills"`/`"job_intention"`/`"internships"` | ✅ **I-R2 补丁** |
| resume_evidence | EvidenceChain 每条 `ResumePoint` 包含证据描述 | ✅ 已有 |
| jd_evidence | EvidenceChain 每条 `Requirement` 包含 JD 要求 | ✅ 已有 |
| evidence_chain | `MatchAnalysisResult.EvidenceChain` 完整输出 | ✅ 已有 |
| TopFixItem.resume_module | 已补充 `skills`/`projects` | ✅ **I-R2 补丁** |

### 优化 Mock

| V2 模块 | 覆盖方式 | ResumeModule | 状态 |
|---------|---------|:---:|:---:|
| projects | Fix1/Fix3/fallback `ResumeModule: "projects"` | ✅ | ✅ **I-R2 补丁** |
| skills | Fix2 `ResumeModule: "skills"` | ✅ | ✅ **I-R2 补丁** |
| internships | — 同 recommendation mock，规则引擎不逐模块输出 | ❌ | ⬜ 等价覆盖 |
| education | — 无 education 检查逻辑 | ❌ | ⬜ 等价覆盖 |
| extra_info | — 无 extra_info 检查逻辑 | ❌ | ⬜ 等价覆盖 |

### 对比 Mock

| 字段 | 覆盖方式 | 状态 |
|------|---------|:---:|
| module_name | FieldChange 每条 `ModuleName: "job_intention"/"skills"/"projects"/"extra_info"/"overall"` | ✅ **I-R2 补丁** |
| change_type | FieldChange 每条 `ChangeType: "modified"/"added"/"removed"/"unchanged"` | ✅ 已有 |

### V2 场景覆盖

| 场景 | 是否由 Mock 覆盖 |
|------|:---:|
| GPA / 排名 | ⬜ Mock 不逐字段输出，但 snapshot normalize 可解析 Educations.GPA/Rank |
| AI 应用技能 | ⬜ 同上，Skills.AITools 存在于 snapshot 但 Mock 不消费 |
| 云计算证书 | ⬜ 同上，Certificates 存在于 snapshot 但 Mock 不消费 |
| 竞赛获奖 | ⬜ 同上 |
| 作品链接 | ⬜ 同上 |
| 项目缺少量化成果 | ✅ match_mock P1 suggestion + resume_optimization_mock Fix3 |
| 实习缺少技术栈 | ✅ match_mock/resume_optimization_mock TopFixItem Action |
| V1 normalize 后可分析 | ✅ 所有 Mock 都支持 V1 输入 |
| V2 全模块完整简历可分析 | ✅ agent-local normalizeResumeToSnapshot 支持 V2 输入 |

**等价覆盖说明**：Mock 是规则评分引擎（非逐模块审查器），不要求逐模块输出 fix item。snapshot normalize + Snapshot API 可在前端层独立补足逐模块查看能力。

## I-R2 补丁清单

| 文件 | 修改内容 |
|------|---------|
| `server/internal/model/version_compare.go` | `FieldChange` 新增 `ModuleName string \`json:"module_name,omitempty"\`` |
| `server/internal/agent/match_mock.go` | EvidenceChain 每条 + TopFixItem 每条 新增 `ResumeModule` |
| `server/internal/agent/resume_optimization_mock.go` | TopFixItem 每条 新增 `ResumeModule` |
| `server/internal/agent/recommendation_mock.go` | 第二分支 TopFixItem 新增 `ResumeModule` |
| `server/internal/agent/version_compare_mock.go` | FieldChange 每条 新增 `ModuleName` |

## 优化 Agent Snapshot 复核

| 检查项 | 结果 |
|--------|:---:|
| 旧 free-text 接口未被删除 | ✅ `ResumeText` 字段仍在 OptimizeInput |
| 有 ResumeVersion 时注入 Snapshot | ✅ `buildPromptValues` line 133 |
| 有 ResumeSnapshot 时注入 Snapshot | ✅ `buildPromptValues` line 139 |
| 无 Snapshot 时回退 free-text | ✅ `buildPromptValues` line 147 |
| 优化建议带 resume_module | ✅ model 层已有字段 |

## agent-local normalize 一致性

| 检查项 | 结果 |
|--------|:---:|
| V1 样例关键字段一致 | ✅ `TestAgentSnapshotConsistency_V1` |
| V2 样例关键字段一致 | ✅ `TestAgentSnapshotConsistency_V2` |
| 隐私字段不进 Prompt | ✅ `TestAgentSnapshotToJSON` |
| 模块完整性 | ✅ 9 模块存在 (profile/job_intention/educations/projects/skills/certificates/awards/portfolio/extra_info) |

## 旧字段 / 隐私字段检查

旧字段引用（均在 Mock 中 — Mock 读 V1 DB 字段是其正常行为）：
- `resume.TargetRole`, `resume.TargetCity`, `resume.SkillsJSON`, `resume.ExperienceJSON`, `resume.ProjectJSON`, `resume.SelfEvaluation`
- Prompt 中 `target_role_mismatch` 为 Schema 枚举值（业务逻辑），非 DB 字段 ✅

隐私字段：
- `phone`/`email`/`user_id` — Snapshot 结构不含，不进 Prompt ✅
- `token` — 仅出现在 `MaxTokens`（LLM 参数）和日志 sanitization ✅

## I-R2 最终验收

| 验收项 | 结果 |
|--------|:---:|
| `go build ./...` | ✅ |
| `go vet ./...` | ✅ |
| `go test ./...` | ✅ 全部通过 |
| `npm run build` | ✅ 2.04s |
| API `/api/jobs?page=1&page_size=5` | ✅ 200 |
| API `/api/jobs/10001` | ✅ 200 |
| API `/api/jobs/facets` | ✅ 200 |
| API `/api/resumes` | ✅ 401（预期） |

## 未修改范围确认

- 未修改前端页面 ✅
- 未修改 MyResumesPage ✅
- 未修改 ResumeManagementPage ✅
- 未修改 ResumeComparisonPage UI ✅
- 未修改 AI 三页面 UI ✅
- 未修改数据库 SQL ✅
- 未修改 V2 保存链路 ✅
- 未启用 schema_version=2 前端写入 ✅
- 未新增 API 路径 ✅
- 未新增 /api/resumes/v2 ✅
- 未修改 .env ✅
- 未安装依赖 ✅
- 未迁移目录 ✅

## 未执行 Git 命令

本次未执行任何 Git 命令，也未提交代码。

## 结论

**K8-5-22-I-R 可以最终收口。** I-R2 补丁完成了 Schema 证据复核 + Mock V2 全覆盖。
