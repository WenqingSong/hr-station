# K8-5-28-B：前端 Agent response types / API / adapter 补齐 evidence 字段

## 本阶段性质

前端数据层补齐。只修 types / adapter / ViewModel，不改页面 UI、后端、SQL、Agent/Prompt/Schema/Mock。

## 修改清单

| 文件 | 修改 |
|------|------|
| `web/src/types/agent.ts` | `TopFixItem` + `resume_module?`；`EvidenceItem` + `resume_module?`；新增 `JDEvidenceChainItem`；`ResumeOptimizationResult` + `jd_evidence_chain` |
| `web/src/types/resume.ts` | `FieldChange` + `module_name?` |
| `web/src/adapters/agentAdapter.ts` | `TopFixItemViewModel` / `EvidenceItemViewModel` + `resumeModule?`；converter 透传 `resume_module` |
| `web/src/adapters/resumeAdapter.ts` | `ResumeComparisonViewModel.fieldChanges` + `moduleName?`；mapping 透传 `fc.module_name` |
| `web/src/adapters/resumeOptimizationAdapter.ts` | 新增 `JDEvidenceChainViewModel`；`ResumeOptimizationPageViewModel` + `jdEvidenceChain`；converter 映射 `jd_evidence_chain` |

无后端修改。

## evidence 字段补齐总表

| 字段 | 后端有 | 前端 type | adapter/ViewModel | 页面 UI |
|------|:---:|:---:|:---:|:---:|
| `TopFixItem.resume_module` | ✅ | ✅ | ✅ | ⬜ 留给 C/D/E |
| `EvidenceItem.resume_module` | ✅ | ✅ | ✅ | ⬜ 留给 C/D |
| `FieldChange.module_name` | ✅ | ✅ | ✅ | ⬜ 留给 F |
| `JDEvidenceChain` (优化) | ✅ | ✅ | ✅ | ⬜ 留给 E |
| `FieldChange.change_type` | ✅ | ✅（已有） | ✅（已有） | ✅（已有） |
| `evidence_chain` (推荐/匹配) | ✅ | ✅（已有） | ✅（已有） | ✅（已有） |
| `top_fix_items` | ✅ | ✅（已有） | ✅（已有） | ⚠️ 推荐页未展示 |

## 推荐链路结果

| 字段 | 结果 |
|------|:---:|
| `resume_module` (TopFixItem) | ✅ 类型 + adapter 已补齐 |
| `resume_module` (EvidenceItem) | ✅ 类型 + adapter 已补齐 |
| `evidence_chain` | ✅ 已有 |
| `top_fix_items` | ✅ 已有（类型/adapter） |

## 匹配链路结果

| 字段 | 结果 |
|------|:---:|
| `resume_module` (TopFixItem) | ✅ |
| `resume_module` (EvidenceItem) | ✅ |
| `evidence_chain` | ✅ 已有 |
| `top_fix_items` | ✅ 已有 |

## 优化链路结果

| 字段 | 结果 |
|------|:---:|
| `resume_module` (TopFixItem) | ✅ |
| `jd_evidence_chain` | ✅ 类型 + ViewModel + adapter 已补齐 |
| `top_fix_items` | ✅ 已有 |

## 对比链路结果

| 字段 | 结果 |
|------|:---:|
| `module_name` (FieldChange) | ✅ 类型 + adapter 已补齐 |
| `change_type` (FieldChange) | ✅ 已有 |

## 构建 / API / 测试结果

| 验收项 | 结果 |
|--------|:---:|
| `npm run build` (1.98s) | ✅ |
| `go test ./...` (10 packages) | ✅ |
| `go vet ./...` | ✅ |
| API 200/200/200/401 | ✅ |

## 未修改范围确认

```text
未修改后端 ✅
未修改数据库 SQL ✅
未修改 Agent / Prompt / Schema / Mock ✅
未修改 API 路径 ✅
未新增 /api/resumes/v2 ✅
未修改登录鉴权 ✅
未修改 jobs 页面 ✅
未修改 .env ✅
未安装依赖 ✅
未迁移目录 ✅
未新增上传能力 ✅（type="file"/FormData/upload/multipart/dropzone/drag = 0 命中）
未修改页面 UI 展示 ✅（pages/components = 0 命中）
```

## 风险与边界

```text
页面 evidence 展示增强留给 K8-5-28-C/D/E/F。
Mock API 端到端回归留给 K8-5-28-G。
最终验收留给 K8-5-28-Z。
```

## 未执行 Git 命令

本次未执行任何 Git 命令，也未提交代码。
