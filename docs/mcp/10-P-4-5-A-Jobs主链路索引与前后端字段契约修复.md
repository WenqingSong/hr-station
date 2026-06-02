# 10-P-4-5-A Jobs 主链路索引与前后端字段契约修复

## 1. 阶段信息

* 阶段：10-P-4-5-A
* 日期：2026-06-02
* 范围：Jobs 主链路索引补齐 + external_source AppError + 前端类型契约修复
* 是否修改代码：是
* 是否修改前端：是
* 是否修改后端：是
* 是否修改数据库：是（仅索引 DDL）
* 是否接 LLM：否
* 是否 git commit：否

## 2. 本次结论摘要

* `idx_jobs_job_type` 已补齐，`init.sql` 新增单列索引，覆盖 `recruit_type` 筛选 + facets GROUP BY job_type
* `external_source` 错误已改为 AppError：service 返回 `ErrInvalidExternalSource` sentinel，handler 统一使用 `handleJobError(w, err)`，消除字符串匹配
* 前端 Jobs 列表类型已对齐后端：新增 `JobApiItem` / `JobListResponse` 类型，`getJobs()` 返回 `JobListResponse`
* adapter 新增 `toJobCardViewModelFromApi()`，从 `JobApiItem` 读取后端字段并映射为 ViewModel
* 旧版 adapter 函数保留向后兼容（Detail flow 尚未迁移）
* `recruit_type` = `job_type`（API 对外 = 数据库内部），设计已冻结
* `status` 当前只作为返回字段，不开放筛选参数，设计已冻结
* `go build` / `go vet` / `go test` 全部通过（9 packages PASS）
* `npm build` / `npm typecheck` 全部通过（零错误）
* 未 commit

## 3. 10-P-4-5 问题修复对照表

| 问题编号 | 原问题 | 修复方式 | 当前状态 |
| ---- | --- | ---- | ---- |
| P1-J1 | `job_type` 无索引 | `init.sql` 新增 `KEY idx_jobs_job_type (job_type)` | ✅ 已修复 |
| P1-J2 | 前端 `Job` 类型与后端 `JobResponse` 不匹配 | 新增 `JobApiItem` / `JobListResponse` 类型，List 流程 adapter 改用 `toJobCardViewModelFromApi` | ✅ 已修复（List 流程） |
| P1-J3 | `recruit_type` = `job_type` 重复字段语义澄清 | MCP 冻结设计：recruit_type 是 API 对外字段，job_type 是数据库内部字段 | ✅ 已冻结 |
| P2-J3 | `external_source` 错误用字符串匹配 | service 返回 `ErrInvalidExternalSource` sentinel AppError，handler 用 `handleJobError` 统一处理 | ✅ 已修复 |
| — | `status` 设计 | status 当前只返回，不开放筛选，不添加 `idx_jobs_status` | ✅ 已冻结 |

## 4. 后端 schema / 索引变更

| 表 | 索引 | 目的 | 是否影响旧环境 |
| - | -- | -- | ------- |
| `jobs` | `idx_jobs_job_type` (`job_type`) | 覆盖 `recruit_type` 筛选（WHERE job_type = ?）+ facets GROUP BY job_type | 是，需手动迁移 |

**旧环境手动迁移 SQL：**

```sql
ALTER TABLE jobs ADD INDEX idx_jobs_job_type (job_type);
```

## 5. 后端错误处理变更

| 场景 | 修改前 | 修改后 | 响应结构 |
| -- | --- | --- | ---- |
| `external_source` 参数非法 | `job_service.go:107` 返回 `fmt.Errorf("invalid external_source")`，`job_handler.go:69` 用 `err.Error() == "invalid external_source"` 字符串匹配，返回 40001 | `job_service.go:107` 返回 `apperror.ErrInvalidExternalSource` sentinel，`job_handler.go:68-70` 统一调用 `handleJobError(w, err)` | `{code: 40001, message: "invalid external_source", data: null}` |

## 6. 前端类型契约变更

| 文件 | 修改前 | 修改后 | 说明 |
| -- | --- | --- | -- |
| `web/src/api/jobs.ts` | 无 `JobApiItem` / `JobListResponse`；`getJobs()` 返回 `Promise<Job[]>` | 新增 `JobApiItem`（对齐后端 `JobResponse`）、`JobListResponse`（对齐 `GET /api/jobs` 响应）；`getJobs()` 返回 `Promise<JobListResponse>` | 类型显式对齐后端 JSON |
| `web/src/adapters/jobAdapter.ts` | `toJobCardViewModel` 接受 `Job` 类型（business_group/locations/type_label/tags/short_description），旧版 label maps 用英文 key | 新增 `toJobCardViewModelFromApi(job: JobApiItem)`，从 `department`/`city`/`recruit_type`/`skill_keywords`/`description` 映射；旧版函数保留向后兼容 | List 流程使用新版 |
| `web/src/app/components/JobList.tsx` | `.then((jobs) => setJobCards(toJobCardViewModels(jobs)))` | `.then((res) => setJobCards(toJobCardViewModelsFromApi(res.items)))` | 适配新返回结构 |
| `web/src/api/viewModels.ts` | `loadJobListViewModels` 调用 `getJobs()` → `toJobCardViewModels(jobs)` | `loadJobListViewModels` 调用 `getJobs()` → `toJobCardViewModelsFromApi(res.items)` | 适配新返回结构 |

## 7. 字段映射冻结表

| 后端字段 | 前端 API 类型字段 | 前端 ViewModel 字段 | 设计结论 |
| ---- | ----------- | --------------- | ---- |
| `department` | `JobApiItem.department` | `businessGroup` | 由 adapter `toJobCardViewModelFromApi` 映射，冻结 |
| `city` | `JobApiItem.city` | `locationsText` | 当前单值城市，adapter 直接赋值字符串，冻结 |
| `job_type` | `JobApiItem.job_type` | — | 数据库内部字段，API 同时返回 job_type 和 recruit_type，冻结 |
| `recruit_type` | `JobApiItem.recruit_type` | `typeLabel`（经 `RECRUIT_TYPE_LABELS` 映射） | API 对外字段，值等于 job_type，adapter 做中文映射，冻结 |
| `skill_keywords` | `JobApiItem.skill_keywords` (string[]) | `tags` | adapter 映射 `skill_keywords ?? []` → `tags`，冻结 |
| `description` | `JobApiItem.description` | `shortDescription` | adapter 全量赋值（不截断），CSS 处理溢出，冻结 |
| `status` | `JobApiItem.status` | `statusLabel`（经 `STATUS_LABELS` 映射） | 当前只返回不筛选，active → "招聘中"，冻结 |
| `category` | `JobApiItem.category` | `categoryLabel` | 后端返回中文值（如"后端开发"），adapter 直接透传，冻结 |

## 8. status / recruit_type / pagination 设计冻结

* `status` 是否开放筛选：否。当前仅作为返回字段，不开放为查询参数。
* 是否加 `idx_jobs_status`：否。本阶段不添加。
* `recruit_type` 与 `job_type` 关系：API 对外字段 `recruit_type` = 数据库内部字段 `job_type`。前端筛选参数继续使用 `recruit_type`。facets 返回 `recruit_types`（按 `job_type` GROUP BY）。
* `limit/offset` 与 `page/page_size` 关系：当前使用 `limit/offset`，不新增 `page/page_size` 参数。

## 9. 验证结果

| 命令 | 结果 | 说明 |
| -- | -- | -- |
| `git status --short` 执行前 | 8 文件 modified | 仅限允许修改范围 |
| `go build ./cmd/server` | ✅ PASS | 编译通过 |
| `go vet ./...` | ✅ PASS | 零告警 |
| `go test ./...` | ✅ PASS | 9 packages PASS |
| `npm run build` | ✅ PASS | 1.70s，1676 modules |
| `npm run typecheck` | ✅ PASS | 零错误 |
| `npm run lint` | ❌ 不可用 | 项目无 ESLint 配置（已知 P1） |
| `git status --short` 执行后 | 8 文件 modified | 仅限允许修改范围 |

## 10. 修改文件清单

| 文件 | 修改内容 | 是否业务代码 |
| -- | ---- | ------ |
| `server/sql/init.sql` | 新增 `KEY idx_jobs_job_type (job_type)` | 否（schema） |
| `server/internal/errors/error.go` | 新增 `ErrInvalidExternalSource` sentinel | 否（errors 包） |
| `server/internal/http/job_handler.go` | List handler 错误处理改用 `handleJobError` | 否 |
| `server/internal/service/job_service.go` | external_source 校验改用 `apperror.ErrInvalidExternalSource` | 否 |
| `web/src/api/jobs.ts` | 新增 `JobApiItem`、`JobListResponse` 类型；`getJobs()` 返回类型改为 `JobListResponse` | 否（类型/API 层） |
| `web/src/adapters/jobAdapter.ts` | 新增 `toJobCardViewModelFromApi`、`toJobCardViewModelsFromApi`；更新 label maps | 否（adapter 层） |
| `web/src/api/viewModels.ts` | `loadJobListViewModels` 适配新返回结构 | 否（loader） |
| `web/src/app/components/JobList.tsx` | 调用处适配 `res.items` + `toJobCardViewModelsFromApi` | 是（组件层，仅类型适配） |

## 11. 未做事项

* 未接 LLM
* 未新增接口
* 未修改 Jobs 成功业务行为
* 未添加 status 筛选
* 未添加 `idx_jobs_status`
* 未修改 Auth / Resumes / Agent
* 未修改 Detail 流程（getJobDetail → JobDetail → toJobDetailViewModel 仍使用旧类型）
* 未修改 `JobMatchAnalysisPage.tsx` / `ResumeOptimizationPage.tsx`（不在本阶段 scope）
* 未修改 `MatchAnalysisHeader.tsx` / `MatchAnalysisSidebarLeft.tsx` / `ResumeOptimizationHeader.tsx` / `ResumeOptimizationSidebarLeft.tsx`（不在本阶段 scope）
* 未 git commit

## 12. 当前风险与后续建议

* **已解决**：`idx_jobs_job_type` 索引补齐 → facets/筛选查询可走索引
* **已解决**：`external_source` 错误改用 AppError → 消除字符串匹配
* **已解决**：前端 List 流程类型对齐后端 → `JobApiItem` / `toJobCardViewModelFromApi`
* **仍需下一条修复 prompt**：Detail 流程（`getJobDetail` → `JobDetail` → `toJobDetailViewModel`）仍使用旧类型，`JobDetail` 字段 `banner_summary`/`responsibilities`/`bonus_points` 等与后端 `JobResponse` 不对齐。涉及文件：`JobMatchAnalysisPage.tsx`、`ResumeOptimizationPage.tsx` 及 4 个 sidebar/header 组件。需下一条 prompt 明确 scope 后修复。
* **不需要处理，原因是**：`page/page_size` 参数当前不引入，`limit/offset` 足够

## 13. docs 更新清单

| 文件 | 新增 / 更新 | 目的 |
| -- | ------- | -- |
| `docs/mcp/项目上下文.md` | 更新 | 更新当前阶段为 10-P-4-5-A |
| `docs/mcp/10-P-4-5-A-Jobs主链路索引与前后端字段契约修复.md` | 新增 | 本阶段 MCP 记录 |
| `docs/09-开发日志/2026-06-02-244-10-P-4-5-A-Jobs主链路修复.md` | 新增 | 开发日志 |

## 14. 建议 commit 信息

* 是否已 commit：否
* 建议 commit message：

```text
fix(jobs): align job contract and add job type index
```

## 15. CodeBuddy 重启恢复摘要

10-P-4-5-A 完成 Jobs 主链路索引与前后端字段契约修复。后端：`init.sql` 新增 `idx_jobs_job_type` 索引，`external_source` 错误改用 `ErrInvalidExternalSource` sentinel AppError（消除字符串匹配），handler 统一使用 `handleJobError`。前端：新增 `JobApiItem` / `JobListResponse` 类型对齐后端 JSON，`getJobs()` 返回 `JobListResponse`，adapter 新增 `toJobCardViewModelFromApi` 从后端原始字段映射 ViewModel。`recruit_type`/`job_type`/`status`/分页设计已冻结。go build/vet/test + npm build/typecheck 全部通过。未 commit。Detail 流程（`getJobDetail` → `JobDetail`）尚未迁移，需下一条 prompt 处理。

下次继续前读取：`docs/mcp/项目上下文.md`、`docs/mcp/10-P-4-5-A-Jobs主链路索引与前后端字段契约修复.md`、`docs/mcp/10-P-4-5-Jobs模块索引字段契约与查询稳定性核验.md`。

## 16. 最终判断

1. `idx_jobs_job_type` 是否已补齐？是。
2. 前端 Jobs 类型是否已对齐后端？是（List 流程已对齐；Detail 流程待下一条）。
3. `external_source` 错误是否已改为 AppError？是。
4. `recruit_type/job_type` 关系是否已冻结？是。
5. `status` 是否已冻结为只返回、不筛选？是。
6. 验证命令是否全部通过？是（go build/vet/test + npm build/typecheck 全部通过）。
7. 是否修改前端？是（4 个文件：api/jobs.ts、adapters/jobAdapter.ts、api/viewModels.ts、components/JobList.tsx）。
8. 是否修改后端？是（4 个文件：init.sql、errors/error.go、http/job_handler.go、service/job_service.go）。
9. 是否修改数据库 schema？是（仅新增索引）。
10. 是否 git commit？否。
11. 本阶段是否可以通过？是。
12. 是否建议用户下一步确认 commit？是。
