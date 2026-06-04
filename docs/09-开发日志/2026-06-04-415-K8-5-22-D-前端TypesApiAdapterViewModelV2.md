# K8-5-22-D — 前端 Types / API / Adapter / ViewModel V2

## 性质
implementation（仅前端数据层，不改页面/后端/Agent/SQL）。

## 本阶段目标

让前端 types / api / adapter / viewModel 具备 V2 字段识别能力，但不改页面 UI，不启用新版 10 模块表单。

## 修改文件

| 文件 | 修改内容 |
|------|----------|
| `web/src/types/resume.ts` | 新增 10 个 V2 模块类型 + V2 request 类型 + Resume 接口扩展 V2 字段 + ResumeVersionListItem 扩展 V2 字段 |
| `web/src/api/resumes.ts` | createResume/updateResume 支持 V2 payload 类型；导出 V2 request 类型 |
| `web/src/adapters/resumeAdapter.ts` | adapter 函数 V2-aware；新增 getV2School/getV2Degree/getV2Major 等 V2→V1 fallback helper；toResumeOptionViewModelsFromVersions 支持 V2 摘要字段；ViewModel 新增 schemaVersion |
| `web/src/api/viewModels.ts` | 无修改（已通过 adapter 兼容 V2） |

## 新增/扩展前端类型

### V2 模块类型（10 个）

| 类型 | 用途 |
|------|------|
| `ResumeSchemaVersion` | 1\|2 schema 版本 |
| `ResumeProfileV2` | 模块 1：基础信息 |
| `ResumeJobIntentionV2` | 模块 2：求职意向 |
| `ResumeEducationV2` | 模块 3：教育经历（扩展版） |
| `ResumeInternshipV2` | 模块 4：实习经历 |
| `ResumeProjectV2` | 模块 5：项目经历（扩展版） |
| `ResumeSkillsV2` | 模块 6：技能（分类版） |
| `ResumeCertificateV2` | 模块 7：证书与语言 |
| `ResumeAwardV2` | 模块 8：获奖与竞赛 |
| `ResumePortfolioV2` | 模块 9：作品链接 |
| `ResumeExtraInfoV2` | 模块 10：补充信息 |

### V2 Request 类型

| 类型 | 用途 |
|------|------|
| `CreateResumeRequestV1` | V1 写操作 payload |
| `CreateResumeRequestV2` | V2 写操作 payload（schema_version=2, 10 模块） |
| `CreateResumeRequest` | Union type: V1 \| V2 |

### Resume 接口扩展

Resume 接口新增 15 个可选 V2 字段：
- `schema_version`、`target_direction`、`school`、`degree`、`major`、`graduation_year`（摘要列）
- `profile_json`、`job_intention`、`internships`、`certificates`、`awards`、`portfolio_links`、`extra_info`（JSON 列）
- `internships`（对应后端 response `internships` 字段）

### ResumeVersionListItem 扩展

新增 5 个可选 V2 字段：`schema_version`、`target_direction`、`expected_cities`、`school`、`degree`

## V1/V2 兼容策略

### adapter fallback 逻辑

```
V2 字段优先 → V1 字段 fallback:

- target_direction → direction_label → direction
- school → profile.school → education[0].school
- degree → profile.education → education[0].degree
- major → profile.major → education[0].major
- graduation_year → profile.graduation_year
- expected_cities → (V2 summary column, comma-separated)
```

### ViewModel 扩展

| ViewModel | 新增字段 | 说明 |
|-----------|---------|------|
| ResumeSummaryViewModel | `schemaVersion?: number` | 标识 V1/V2 |
| ResumeManagementViewModel | `schemaVersion?: number` | 管理页感知 V2 |
| ResumeOptionViewModel | `schemaVersion?: number` | 选择弹窗感知 V2 |

### API 层

- `createResume(input: Partial<Resume> | CreateResumeRequestV2)`：接受 V1 或 V2 payload
- `updateResume()`：同上
- 类型导出 `CreateResumeRequest`、`CreateResumeRequestV1`、`CreateResumeRequestV2`
- 不新增 API 路径，不新增 `/api/resumes/v2`

## 页面未修改确认

- `MyResumesPage`：未修改 ✅
- `ResumeManagementPage`：未修改 ✅
- `ResumeComparisonPage`：未修改 ✅
- `SelectResumeModal`：未修改 ✅
- `JobRecommendationPage`、`JobMatchAnalysisPage`、`ResumeOptimizationPage`：未修改 ✅

## 后端/Agent/SQL 未修改确认

- Go 源码：未修改 ✅
- Agent / Prompt / Schema / Mock：未修改 ✅
- SQL / migration：未修改 ✅
- `.env` / `.env.example`：未修改 ✅

## 验收结果

| 验收项 | 结果 |
|--------|------|
| `npm run build` | ✅ 2.37s |
| `go test ./...` | ✅ 全部通过 |
| `go build ./...` | ✅ 通过 |
| `go vet ./...` | ✅ 通过 |
| API `/api/jobs?page=1&page_size=5` | ✅ 200 |
| API `/api/jobs/10001` | ✅ 200 |
| API `/api/jobs/facets` | ✅ 200 |
| API `/api/resumes` | ✅ 401（正常，需认证） |
| 前端 V2 类型 grep 可见 | ✅ 20+ 处命中 |
| 页面未误改 grep | ✅ 0 命中 |
| 后端未误改 grep | ✅ 0 命中 |

## 未执行 Git 命令

本次未执行任何 Git 命令，也未提交代码。

## 风险与边界

- 本阶段只完成前端数据层 V2 识别能力
- MyResumesPage V2 Shell 留给 K8-5-22-E
- 10 模块表单实现留给 K8-5-22-F/G/H
- Agent 接入留给 K8-5-22-I
- 前端尚未发送 schema_version=2
