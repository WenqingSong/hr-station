# 10-P-4-5-C Jobs 主链路契约统一验收与 commit 确认

## 1. 阶段信息

* 阶段：10-P-4-5-C
* 日期：2026-06-02
* 范围：10-P-4-5-A/B 统一验收，设计冻结确认，commit 建议
* 是否修改业务代码：否
* 是否修改前端：否
* 是否修改后端：否 
* 是否修改数据库：否
* 是否接 LLM：否
* 是否 git commit：否

## 2. 本次结论摘要

* 10-P-4-5-A 通过统一验收：`idx_jobs_job_type` 索引已补齐，`external_source` 已改 AppError，List Flow 已对齐 `JobApiItem`
* 10-P-4-5-B 通过统一验收：Detail Flow 已对齐 `JobApiItem`，全部旧版 adapter 函数已移除，9 个页面/组件已迁移
* Jobs List Flow 已对齐后端 `JobApiItem`：`getJobs()` → `JobListResponse`，adapter `toJobCardViewModelFromApi` 映射正确
* Jobs Detail Flow 已对齐后端 `JobApiItem`：`getJobDetail()` → `JobApiItem`，adapter `toJobDetailViewModelFromApi` 从真实字段计算
* 后端 `idx_jobs_job_type` 已确认存在于 `init.sql:111`
* `external_source` AppError 已确认：service 返回 `ErrInvalidExternalSource`，handler 统一 `handleJobError`，零字符串匹配
* mock / fallback / as any 零反弹（活跃代码）
* 旧 API 字段（`banner_summary`/`business_group`/`type_label`/`short_description`）在活跃代码中零泄漏
* 所有旧版 adapter 函数（`toJobCardViewModel(job: Job)` 等）已全部移除
* `go build` / `go vet` / `go test` 全部通过（9 packages PASS）
* `npm build` / `npm typecheck` 全部通过（零错误）
* 10-P-4-5-A 已 commit：`208fa13 feat: 对齐岗位列表接口与前端适配层`
* 10-P-4-5-B 已 commit：`4dc26a4 refactor: 移除旧版 Job 类型适配，统一使用 JobApiItem`
* 工作区干净，零未提交改动
* **建议用户确认已完成的 commit 状态**

## 3. 10-P-4-5-A 验收结果

| 验收项 | 结果 | 证据 | 是否通过 |
| --- | -- | -- | ---- |
| `idx_jobs_job_type` 索引已存在 | `KEY idx_jobs_job_type (job_type)` | `server/sql/init.sql:111` | ✅ |
| `idx_jobs_status` 未添加 | grep 零匹配 | init.sql 无 `idx_jobs_status` | ✅ |
| `external_source` 使用 AppError | service L107: `return nil, apperror.ErrInvalidExternalSource` | `server/internal/service/job_service.go:107` | ✅ |
| handler 使用 `handleJobError` | handler L69: `handleJobError(w, err)` | `server/internal/http/job_handler.go:69` | ✅ |
| 零字符串匹配 `err.Error()` | grep 零匹配 | `grep -rn "err.Error()" server/internal/http/job_handler.go`: 零结果 | ✅ |
| `JobApiItem` 类型已定义 | `export interface JobApiItem { id: number; title: string; ... }` | `web/src/api/jobs.ts:8-25` | ✅ |
| `JobListResponse` 类型已定义 | `export interface JobListResponse { items: JobApiItem[]; ... }` | `web/src/api/jobs.ts:28-33` | ✅ |
| `getJobs()` 返回 `JobListResponse` | `Promise<JobListResponse>` | `web/src/api/jobs.ts:66` | ✅ |
| `toJobCardViewModelFromApi` 映射正确 | department→businessGroup, city→locationsText, skill_keywords→tags, recruit_type→typeLabel, description→shortDescription, status→statusLabel | `web/src/adapters/jobAdapter.ts:43-54` | ✅ |
| `JobList.tsx` 使用 `res.items` | `.then((res) => setJobCards(toJobCardViewModelsFromApi(res.items)))` | `web/src/app/components/JobList.tsx:31` | ✅ |
| `recruit_type`/`job_type` 冻结 | API 对外 `recruit_type` = 数据库内部 `job_type` | MCP 文档化 | ✅ |
| `status` 冻结 | 只返回不筛选 | 无 handler 支持 status 参数 | ✅ |
| pagination 冻结 | `limit/offset`，无 `page/page_size` | handler/service 仅使用 limit/offset | ✅ |

## 4. 10-P-4-5-B 验收结果

| 验收项 | 结果 | 证据 | 是否通过 |
| --- | -- | -- | ---- |
| `getJobDetail()` 返回 `Promise<JobApiItem>` | `export async function getJobDetail(jobId: string): Promise<JobApiItem>` | `web/src/api/jobs.ts:83` | ✅ |
| `getSelectedJob()` 返回 `Promise<JobApiItem>` | `export async function getSelectedJob(jobId: string): Promise<JobApiItem>` | `web/src/api/jobs.ts:91` | ✅ |
| `toJobDetailViewModelFromApi` 已创建 | 从 `JobApiItem` 读取 `description`/`requirements`/`skill_keywords` 计算 ViewModel | `web/src/adapters/jobAdapter.ts:63-87` | ✅ |
| 旧版 adapter 已清理 | `toJobCardViewModel(job: Job)` / `toJobDetailViewModel(job: JobDetail)` 已移除 | grep 零匹配 | ✅ |
| 旧版英文 key 标签 map 已清理 | `LEGACY_CATEGORY_LABELS` / `LEGACY_TYPE_LABELS` 已移除 | grep 零匹配 | ✅ |
| `JobMatchAnalysisPage` 使用 `JobApiItem` | `useState<JobApiItem | null>(null)` | `web/src/app/pages/JobMatchAnalysisPage.tsx:31` | ✅ |
| `ResumeOptimizationPage` 使用 `JobApiItem` | `useState<JobApiItem | null>(null)` | `web/src/app/pages/ResumeOptimizationPage.tsx:76` | ✅ |
| `buildTargetJob` 使用后端字段 | `job.description` / `job.requirements` | `web/src/app/pages/ResumeOptimizationPage.tsx:60-66` | ✅ |
| `MatchAnalysisHeader` 使用 `JobApiItem` | `import type { JobApiItem }` + `toJobDetailViewModelFromApi` | `MatchAnalysisHeader.tsx:5,17` | ✅ |
| `MatchAnalysisSidebarLeft` 使用 `JobApiItem` | `import type { JobApiItem }` + `toJobDetailViewModelFromApi` | `MatchAnalysisSidebarLeft.tsx:5,13` | ✅ |
| `ResumeOptimizationHeader` 使用 `JobApiItem` | `import type { JobApiItem }` + `toJobDetailViewModelFromApi` | `ResumeOptimizationHeader.tsx:5,17` | ✅ |
| `ResumeOptimizationSidebarLeft` 使用 `JobApiItem` | `import type { JobApiItem }` + `toJobDetailViewModelFromApi` | `ResumeOptimizationSidebarLeft.tsx:5,13` | ✅ |
| `JobDetailPage` 无需修改 | 始终使用 `JobDetailViewModel`（非 `JobDetail`） | prop 传递正确 | ✅ |
| 旧 API 字段零泄漏 | `banner_summary`/`business_group`/`type_label`/`short_description` 在 api/adapters/app 中零匹配 | grep 零结果 | ✅ |

## 5. Jobs 字段映射最终冻结表

| 后端字段 | API 字段 | ViewModel 字段 | 使用位置 | 结论 |
| ---- | ------ | ------------ | ---- | -- |
| `department` | `JobApiItem.department` | `businessGroup` | JobCard, JobDetailBanner, JobDetailAIDecision, Sidebars | 冻结（adapter 映射） |
| `city` | `JobApiItem.city` | `locationsText` | JobCard, JobDetailBanner, JobDetailJD, Sidebars | 冻结（单值城市） |
| `direction` | `JobApiItem.direction` | — | facets 筛选 | 冻结（筛选用） |
| `category` | `JobApiItem.category` | `categoryLabel` | JobCard, JobDetailBanner, JobDetailAIDecision, Sidebars | 冻结（中文值透传） |
| `job_type` | `JobApiItem.job_type` | — | 数据库内部字段 | 冻结（= recruit_type） |
| `recruit_type` | `JobApiItem.recruit_type` | `typeLabel`（经 RECRUIT_TYPE_LABELS 映射） | JobCard, JobDetailBanner, JobDetailAIDecision, Sidebars | 冻结（API 对外字段） |
| `description` | `JobApiItem.description` | `shortDescription` / `bannerSummary` / `responsibilities`（拆分） | JobCard, JobDetailBanner, JobDetailJD, Sidebars | 冻结 |
| `requirements` | `JobApiItem.requirements` | `requirements`（按行拆分） | JobDetailJD, Sidebars | 冻结（TEXT 拆行） |
| `skill_keywords` | `JobApiItem.skill_keywords` (string[]) | `tags` / `keywords` | JobCard, JobDetailViewModel | 冻结 |
| `status` | `JobApiItem.status` | `statusLabel`（经 STATUS_LABELS 映射） | JobCard | 冻结（active→招聘中） |
| `external_source` | `JobApiItem.external_source` | — | 后端筛选参数 | 冻结 |
| `is_featured` | `JobApiItem.is_featured` | — | 后端排序 | 冻结 |
| `created_at` | `JobApiItem.created_at` | — | 后端字段 | 冻结 |
| `updated_at` | `JobApiItem.updated_at` | — | 后端字段 | 冻结 |

## 6. 设计冻结结论

* **recruit_type/job_type 关系**：API 对外字段 `recruit_type` = 数据库内部字段 `job_type`。后端同时返回两字段（值相同）。前端筛选参数使用 `recruit_type`。facets 返回 `recruit_types`（按 `job_type` GROUP BY）。已冻结，不修改。
* **status 设计**：当前仅作为返回字段出现在 API 响应中，不开放为查询筛选参数。前端 `statusLabel` 由 adapter 从 `Status.active` 映射为"招聘中"。不添加 `idx_jobs_status`。已冻结，不修改。
* **pagination 设计**：当前使用 `limit/offset`（service 层硬限制 max 50）。不新增 `page/page_size` 参数。已冻结，不修改。
* **keyword 搜索设计**：当前使用 `LIKE %keyword%` 四字段模糊搜索。当前数据量可接受，不做全文检索。已冻结，不修改。
* **facets 设计**：`GET /api/jobs/facets` 返回 `directions`/`cities`/`recruit_types`/`categories` 四个维度，各含 `{value, label, count}`。字段名、聚合方式已冻结，不修改。
* **索引设计**：当前 8 个索引（city/direction/category/title/external_source/data_quality/is_featured/job_type）。仅 `idx_jobs_job_type` 为本阶段新增。不新增其他索引。

## 7. 反弹检查结果

| 检查项 | 结果 | 说明 |
| --- | -- | -- |
| `as any`（活跃代码） | 零匹配 | 仅 MyResumesPage.tsx（P2 死代码）残留 2 处，非本阶段引入 |
| `@ts-ignore` | 零匹配 | — |
| mock import | 零匹配 | — |
| `withFallback` / `withMockFallback` / `fallbackToMock` | 零匹配 | — |
| "演示数据" / "Mock 演示"（活跃组件） | 零匹配 | `common.ts:34` 的 `Mock 演示结果` 是 Agent 适配器 label，非 Jobs 相关，非本阶段引入 |
| `FILTER_GROUPS` | 零匹配 | — |
| `banner_summary` / `business_group` / `type_label` / `short_description`（API 层） | 零匹配 | 已全部移除 |
| `responsibilities` / `bonusPoints`（API 层） | 零匹配 | 仅 ViewModel 字段，来源为后端真实字段或空值 |
| `err.Error()` 字符串匹配（Jobs 相关） | 零匹配 | handler 统一使用 `handleJobError` + `errors.As` |

## 8. 验证命令结果

| 命令 | 结果 | 说明 |
| -- | -- | -- |
| `git status --short` 执行前 | 干净 | 零修改 |
| `git diff --stat` 执行前 | 零输出 | 10-P-4-5-A/B 已全部 commit |
| `go build ./cmd/server` | ✅ PASS | 编译通过 |
| `go vet ./...` | ✅ PASS | 零告警 |
| `go test ./...` | ✅ PASS | 9 packages PASS |
| `npm run build` | ✅ PASS | 1.63s，1676 modules |
| `npm run typecheck` | ✅ PASS | 零错误 |
| `npm run lint` | ❌ 不可用 | 项目无 ESLint 配置（已知 P1，非本阶段引入） |
| 反弹检查 | ✅ | 零 as any/mock/fallback/旧字段反弹 |
| `git status --short` 执行后 | 干净 | 零新修改 |

## 9. 当前修改范围

| 文件 | 来源阶段 | 修改类型 | 是否允许 |
| -- | ---- | ---- | ---- |
| `server/sql/init.sql` | 10-P-4-5-A | 新增 `idx_jobs_job_type` 索引 | ✅ |
| `server/internal/errors/error.go` | 10-P-4-5-A | 新增 `ErrInvalidExternalSource` | ✅ |
| `server/internal/http/job_handler.go` | 10-P-4-5-A | List error 改用 `handleJobError` | ✅ |
| `server/internal/service/job_service.go` | 10-P-4-5-A | external_source 改用 sentinel | ✅ |
| `web/src/api/jobs.ts` | 10-P-4-5-A + B | 新增 `JobApiItem`/`JobListResponse`；getJobs/getJobDetail/getSelectedJob 全部返回 `JobApiItem` | ✅ |
| `web/src/adapters/jobAdapter.ts` | 10-P-4-5-A + B | 新增 `toJobCardViewModelFromApi`/`toJobDetailViewModelFromApi`；移除全部旧版函数 | ✅ |
| `web/src/api/viewModels.ts` | 10-P-4-5-A + B | 适配新版 adapter | ✅ |
| `web/src/app/components/JobList.tsx` | 10-P-4-5-A | 适配新版返回结构 | ✅ |
| `web/src/app/components/MatchAnalysisHeader.tsx` | 10-P-4-5-B | `JobDetail` → `JobApiItem` | ✅ |
| `web/src/app/components/MatchAnalysisSidebarLeft.tsx` | 10-P-4-5-B | `JobDetail` → `JobApiItem` | ✅ |
| `web/src/app/components/ResumeOptimizationHeader.tsx` | 10-P-4-5-B | `JobDetail` → `JobApiItem` | ✅ |
| `web/src/app/components/ResumeOptimizationSidebarLeft.tsx` | 10-P-4-5-B | `JobDetail` → `JobApiItem` | ✅ |
| `web/src/app/pages/JobMatchAnalysisPage.tsx` | 10-P-4-5-B | `JobDetail` → `JobApiItem` | ✅ |
| `web/src/app/pages/ResumeOptimizationPage.tsx` | 10-P-4-5-B | `JobDetail` → `JobApiItem`；`buildTargetJob` 改用后端字段 | ✅ |
| `docs/mcp/*` + `docs/09-开发日志/*` | 10-P-4-5-A/B/C | MCP 记录 + 开发日志 | ✅ |

## 10. 是否建议 commit

* 是否建议 commit：**已在 10-P-4-5-A 和 10-P-4-5-B 阶段分别 commit**
* commit 历史：
  - `208fa13 feat: 对齐岗位列表接口与前端适配层`（10-P-4-5-A）
  - `4dc26a4 refactor: 移除旧版 Job 类型适配，统一使用 JobApiItem`（10-P-4-5-B）
* 工作区当前状态：干净，零未提交改动
* 建议用户确认：`git log --oneline -3` 检查 commit 历史，确认无误后 push
* 由于 10-P-4-5-A/B 已分开 commit，无需再次 commit 合并
* 如需合并为单个 commit（squash），建议 message：`fix(jobs): align job contracts and indexes`

## 11. 未做事项

* 未修改业务代码
* 未修 bug
* 未新增接口
* 未接 LLM
* 未自动 commit（本阶段零代码修改）
* 未修改 Auth / Resumes / Agent
* 未添加 status 筛选
* 未添加 idx_jobs_status

## 12. 当前风险与后续建议

* **已解决**：Jobs List Flow 字段契约对齐后端 — 10-P-4-5-A commit `208fa13`
* **已解决**：Jobs Detail Flow 字段契约对齐后端 — 10-P-4-5-B commit `4dc26a4`
* **已解决**：`idx_jobs_job_type` 索引补齐
* **已解决**：`external_source` AppError
* **已解决**：recruit_type/job_type/status/pagination 设计冻结
* **已解决**：旧版 adapter 函数 + 英文 key 标签 map 清理
* **不需要处理，原因是**：`page/page_size` 不引入、`status` 不筛选、keyword 不做全文检索 — 均为明确设计决策
* **需要下一阶段处理，阶段名是 10-P-4-6**：剩余 P1（lint 缺失、VersionCompare 仅 Mock 等），或用户指定的下一个优先级

## 13. docs 更新清单

| 文件 | 新增 / 更新 | 目的 |
| -- | ------- | -- |
| `docs/mcp/项目上下文.md` | 更新 | 更新当前阶段为 10-P-4-5-C |
| `docs/mcp/10-P-4-5-C-Jobs主链路契约统一验收与commit确认.md` | 新增 | 本阶段 MCP 记录 |
| `docs/09-开发日志/2026-06-02-246-10-P-4-5-C-Jobs主链路统一验收.md` | 新增 | 开发日志 |

## 14. CodeBuddy 重启恢复摘要

10-P-4-5-C 完成 Jobs 主链路契约统一验收。10-P-4-5-A（List Flow + 索引 + AppError）和 10-P-4-5-B（Detail Flow）全覆盖验收通过。Jobs List/Detail 两条链路均已对齐后端 `JobApiItem`。`recruit_type`/`job_type`/`status`/`pagination`/`facets`/`keyword` 设计冻结确认。`idx_jobs_job_type` 索引确认，`external_source` AppError 确认。零 mock/fallback/as any 反弹，零旧 API 字段泄漏。go build/vet/test + npm build/typecheck 全部通过。10-P-4-5-A 已 commit `208fa13`，10-P-4-5-B 已 commit `4dc26a4`，工作区干净。10-P-4-5 系列可以封存。用户确认后可以 push。

下次继续前读取：`docs/mcp/项目上下文.md`、本文件。

## 15. 最终判断

1. 10-P-4-5-A 是否通过统一验收？是。
2. 10-P-4-5-B 是否通过统一验收？是。
3. Jobs List Flow 是否已对齐后端 API 类型？是。
4. Jobs Detail Flow 是否已对齐后端 API 类型？是。
5. 后端 idx_jobs_job_type 是否确认？是。
6. external_source AppError 是否确认？是。
7. mock / fallback / as any 是否反弹？否。
8. 验证命令是否全部通过？是。
9. 当前是否建议 commit？已 commit（10-P-4-5-A 在 `208fa13`，10-P-4-5-B 在 `4dc26a4`），工作区干净。
10. 是否已经 commit？是（10-P-4-5-A 和 10-P-4-5-B 分别 commit）。
11. 本阶段是否可以通过？是。
12. 是否允许进入用户确认 commit？已 commit，建议用户确认后 push。
