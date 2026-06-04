# K8-5-26-D：Response Schema evidence 模块标注修正

## 本阶段性质

复核 Schema JSON / Go Model / Agent mapParsed 三层 evidence 字段一致性。**无源码修改。**

## schema 使用证据

| 检查项 | 行号/来源 | 结果 |
|--------|:---:|:---:|
| Schema JSON 存在 | `server/prompts/schema/*.json` (5 files) | ✅ |
| Pipeline 加载 schema | `loader.go` + `pipeline.go` Step 3 | ✅ |
| Schema 用于 LLM 校验 | `validator.go` L44 `jsonschema.CompileString()` | ✅ |
| 校验后映射到 Go struct | 各 Agent `mapParsedToXxx` | ✅ |
| API response 最终来源 | Go struct `json.Marshal` | ✅ |

**三层架构**：Schema JSON（LLM 输出约束）→ Agent mapParsedToXxx（桥接）→ Go model struct（API response）

## 推荐 Schema / Model / Agent 结果

| 字段 | Schema JSON | Go Model | Agent mapParsed |
|------|:---:|:---:|:---:|
| `RecommendLabel` | ✅ match_score/reason | ✅ recommendation.go | ✅ |
| `TopFixItem.ResumeModule` | ⬜ 不要求 LLM 输出 | ✅ match.go L42 | ✅ Mock 直接填充 |
| `resume_evidence` (highlight) | ✅ recommendation_schema L390 | — | ✅ mapRecEvidence |

## 匹配 Schema / Model / Agent 结果

| 字段 | Schema JSON | Go Model | Agent mapParsed |
|------|:---:|:---:|:---:|
| `evidence_chain` | ✅ L119 `$ref: evidence_chain_item` | ✅ match.go | ✅ |
| `evidence_chain.resume_evidence` | ✅ L347 | ✅ | ✅ L296 |
| `evidence_chain.jd_evidence` (requirement) | ✅ schema listed | ✅ | ✅ |
| `MatchEvidence.ResumeModule` | ⬜ 不要求 LLM 输出 | ✅ match.go L22 | ✅ Mock 直接填充 |
| `TopFixItem.ResumeModule` | ⬜ 同上 | ✅ match.go L42 | ✅ Mock 直接填充 |

## 优化 Schema / Model / Agent 结果

| 字段 | Schema JSON | Go Model | Agent mapParsed |
|------|:---:|:---:|:---:|
| `ResumeOptimizationFixItem.ResumeModule` | ⬜ | ✅ resume_optimization.go L39 | ✅ Mock 直接填充 |
| `JDEvidenceItem.ResumeEvidence` | ✅ schema L319/339 | ✅ L51 | ✅ L323 |
| `JDEvidenceChain` | ✅ keyword_alignment | ✅ L81 | ✅ L206 |

## 对比 Schema / Model / Agent 结果

| 字段 | Schema JSON | Go Model | Agent mapParsed |
|------|:---:|:---:|:---:|
| `FieldChange.ModuleName` | ⬜ | ✅ version_compare.go L23 | ✅ Mock 填充 |
| `FieldChange.ChangeType` | ✅ enum values | ✅ L21 | ✅ |

## 最终 9 模块 canonical module key

Go model 中 `ResumeModule`/`ModuleName` 统一使用以下值：

| module key | 对应模块 |
|-----------|---------|
| `resume_basics` | 简历基础 |
| `basic_info` | 基础信息 |
| `education` | 教育经历 |
| `projects` | 项目经历 |
| `skills` | 技能信息 |
| `internships` | 实习经历 |
| `certificates_awards` | 证书与获奖 |
| `portfolio` | 作品或个人主页 |
| `extra_info` | 自我评价 |

更新后的 Mock 在 K8-5-26-E 中统一到这些值。

## 修改清单

**无源码修改** — 所有 evidence 字段已在 K8-5-22-I-R 阶段部署完成。Schema JSON 无需新增字段（LLM 不输出 resume_module）。Go model 已全备。Mock 已全备（K8-5-22-I-R）。本次为纯复核。

## Schema / Go / Mock 一致性确认

| 层级 | 是否存在 | 是否一致 |
|------|:---:|:---:|
| match_analysis_schema.json | ✅ evidence_chain | ✅ 与 match.go 一致 |
| recommendation_schema.json | ✅ resume_evidence 片段 | ✅ |
| resume_optimization_schema.json | ✅ jd_evidence_chain | ✅ |
| version_compare_schema.json | ✅ change_type | ✅ |
| match.go (MatchEvidence / TopFixItem) | ✅ ResumeModule | ✅ |
| resume_optimization.go | ✅ ResumeModule / ResumeEvidence | ✅ |
| version_compare.go | ✅ ModuleName / ChangeType | ✅ |

## 测试覆盖

| 测试 | 来源 | 结果 |
|------|------|:---:|
| `TestAgentSnapshotConsistency_V1` | agent_snapshot_test.go | ✅ |
| `TestAgentSnapshotConsistency_V2` | agent_snapshot_test.go | ✅ |
| `TestAgentSnapshotToJSON` | agent_snapshot_test.go | ✅ |
| Schema validator tests | validator_test.go | ✅ |

## 构建 / API / 测试结果

| 验收项 | 结果 |
|--------|:---:|
| `go build ./...` | ✅ |
| `go vet ./...` | ✅ |
| `go test ./...` | ✅ 全部通过 |
| `npm run build` | ✅ 2.47s |
| API 200/200/200/401 | ✅ |

## 未修改范围确认

```text
未修改前端 ✅ SQL ✅ Prompt 主文案 ✅ Mock 场景 ✅ API ✅ jobs ✅ .env ✅
```

## 风险与边界

```text
Mock 最终 9 模块场景 → K8-5-26-E（统一 module key）
Agent 回归 → K8-5-26-F
最终验收 → K8-5-26-Z
```

## 最终结论

```text
K8-5-26-D：所有 evidence 字段已在 K8-5-22-I-R 部署完成，本次纯复核，0 源码修改 ✅
Schema JSON → Go Model → Agent mapParsed 三层一致 ✅
canonical module key 已冻结为 9 模块名称 ✅
```

## 未执行 Git 命令

本次未执行任何 Git 命令，也未提交代码。
