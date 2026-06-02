# 10-P-4-5-B Jobs Detail Flow 字段契约修复

## 1. 阶段信息

* 阶段：10-P-4-5-B
* 日期：2026-06-02
* 范围：Jobs Detail Flow 字段契约修复（getJobDetail → toJobDetailViewModelFromApi → 所有 detail 页面/组件）
* 是否修改代码：是
* 是否修改前端：是
* 是否修改后端：否
* 是否修改数据库：否
* 是否接 LLM：否
* 是否 git commit：否

## 2. 本次结论摘要

* Detail Flow 已全部对齐后端 API 类型：`getJobDetail()` / `getSelectedJob()` 返回 `Promise<JobApiItem>`
* 新增 `toJobDetailViewModelFromApi(job?: JobApiItem | null): JobDetailViewModel`，所有字段从后端真实字段计算
* 旧版 `toJobDetailViewModel(job?: JobDetail)` + `toJobCardViewModel(job: Job)` + `toJobCardViewModels(jobs: Job[])` 已全部移除
* 旧版 `LEGACY_CATEGORY_LABELS` / `LEGACY_TYPE_LABELS`（英文 key）已移除
* `JobDetailPage` / `JobDetailBanner` / `JobDetailJD` / `JobDetailAIDecision` — 无需修改（始终使用 `JobDetailViewModel`）
* `JobMatchAnalysisPage` — `JobDetail` → `JobApiItem`
* `ResumeOptimizationPage` — `JobDetail` → `JobApiItem`；`buildTargetJob` 改用 `description` + `requirements` 后端字段
* `MatchAnalysisHeader` / `MatchAnalysisSidebarLeft` / `ResumeOptimizationHeader` / `ResumeOptimizationSidebarLeft` — `JobDetail` → `JobApiItem`
* 未新增假字段、未引入 mock/fallback、未使用 `as any` / `@ts-ignore`
* `npm build` / `npm typecheck` / `go build` / `go vet` / `go test` 全部通过
* 未 commit

## 3. 执行前问题清查

| 问题 | 证据文件 | 影响 | 修复方式 |
| -- | ---- | -- | ---- |
| `getJobDetail()` 返回 `Promise<JobDetail>` | `api/jobs.ts:84` | 调用方使用不存在字段如 `banner_summary`/`responsibilities`/`bonus_points` | 改为返回 `Promise<JobApiItem>` |
| `toJobDetailViewModel` 读取 `JobDetail` 旧字段 | `jobAdapter.ts:104-115` | `banner_summary`/`responsibilities`/`bonus_points` 等字段后端不提供 | 新增 `toJobDetailViewModelFromApi` 从 `JobApiItem` 计算 |
| `MatchAnalysisHeader` / `MatchAnalysisSidebarLeft` 接受 `JobDetail` | 源文件 | 运行时访问不存在字段 | 改为接受 `JobApiItem` |
| `ResumeOptimizationHeader` / `ResumeOptimizationSidebarLeft` 接受 `JobDetail` | 源文件 | 运行时访问不存在字段 | 改为接受 `JobApiItem` |
| `ResumeOptimizationPage.buildTargetJob` 读取 `banner_summary`/`short_description`/`responsibilities`/`bonus_points` | `ResumeOptimizationPage.tsx:58-68` | 字段均不存在 | 改写为使用 `description` + `requirements` |
| `LEGACY_CATEGORY_LABELS` / `LEGACY_TYPE_LABELS` 英文 key 映射 | `jobAdapter.ts:70-84` | 后端返回中文值如"后端开发"，英文 key 映射永不命中 | 移除旧版函数，统一使用中文 key maps |

## 4. Detail API 类型修复

| 项 | 修改前 | 修改后 | 说明 |
| - | --- | --- | -- |
| `getJobDetail()` 返回类型 | `Promise<JobDetail>` | `Promise<JobApiItem>` | 对齐后端 `GET /api/jobs/{id}` 真实返回结构 |
| `getSelectedJob()` 返回类型 | `Promise<JobDetail>` | `Promise<JobApiItem>` | 同上 |
| `api/jobs.ts` imports | `import type { Job, JobDetail } from '../types'` | 无（移除） | 不再依赖旧 UI 类型 |

## 5. Detail ViewModel 字段映射冻结

| 后端字段 | ViewModel 字段 | 使用位置 | 设计结论 |
| ---- | ------------ | ---- | ---- |
| `department` | `businessGroup` | JobDetailBanner, MatchAnalysisSidebarLeft | 由 `toJobCardViewModelFromApi` 映射，冻结 |
| `city` | `locationsText` | JobDetailBanner, JobDetailJD, Sidebars | 单值城市，冻结 |
| `recruit_type` | `typeLabel`（经 RECRUIT_TYPE_LABELS 映射） | JobDetailBanner, JobDetailAIDecision, Sidebars | 冻结 |
| `category` | `categoryLabel` | JobDetailBanner, JobDetailAIDecision, Sidebars | 后端中文值透传，冻结 |
| `description` | `bannerSummary` / `shortDescription` / `responsibilities` | JobDetailBanner, JobDetailAIDecision, JobDetailJD | `bannerSummary`=全量 description；`shortDescription`=全量；`responsibilities`=按句号/分号拆分 |
| `requirements` | `requirements`（按行拆分） | JobDetailJD, Sidebars | 后端 TEXT 拆分换行行，去序号前缀 |
| `skill_keywords` | `tags` / `keywords` | JobCard, JobDetailViewModel | 冻结 |
| `status` | `statusLabel`（经 STATUS_LABELS 映射） | JobCard | `active` → "招聘中"，冻结 |
| `direction` | — | 筛选 facets 用 | 冻结，ViewModel 未使用 |
| `job_type` | — | 数据库内部字段 | 冻结 |
| — | `bonusPoints` | JobDetailJD | 后端无对应字段，固定空数组 `[]` |
| — | `deliveryNotes` | JobDetailJD | 后端无对应字段，默认 `undefined`，组件兜底文案 |

## 6. 页面 / 组件迁移清单

| 文件 | 修改内容 | 是否用户可见 | 说明 |
| -- | ---- | ------ | -- |
| `web/src/api/jobs.ts` | `getJobDetail()` / `getSelectedJob()` 返回 `Promise<JobApiItem>`；移除 `Job`/`JobDetail` imports | 否（类型层） | API 类型对齐后端 |
| `web/src/adapters/jobAdapter.ts` | 新增 `toJobDetailViewModelFromApi`；移除旧版 `toJobCardViewModel`、`toJobCardViewModels`、`toJobDetailViewModel`；移除英文 key 标签 map | 否（adapter 层） | adapter 统一使用后端原始字段 |
| `web/src/api/viewModels.ts` | `loadJobDetailViewModel` / `loadSelectedJobDetailViewModel` 调用改为 `toJobDetailViewModelFromApi` | 否（loader） | 适配新 adapter |
| `web/src/app/pages/JobMatchAnalysisPage.tsx` | `JobDetail` → `JobApiItem` | 是 | 匹配分析页目标岗位类型对齐 |
| `web/src/app/pages/ResumeOptimizationPage.tsx` | `JobDetail` → `JobApiItem`；`buildTargetJob` 改用 `description` + `requirements` | 是 | 简历优化页目标岗位类型对齐 |
| `web/src/app/components/MatchAnalysisHeader.tsx` | `JobDetail` → `JobApiItem`，`toJobDetailViewModel` → `toJobDetailViewModelFromApi` | 是 | 匹配分析页眉 |
| `web/src/app/components/MatchAnalysisSidebarLeft.tsx` | `JobDetail` → `JobApiItem`，`toJobDetailViewModel` → `toJobDetailViewModelFromApi` | 是 | 匹配分析页左侧栏 |
| `web/src/app/components/ResumeOptimizationHeader.tsx` | `JobDetail` → `JobApiItem`，`toJobDetailViewModel` → `toJobDetailViewModelFromApi` | 是 | 简历优化页眉 |
| `web/src/app/components/ResumeOptimizationSidebarLeft.tsx` | `JobDetail` → `JobApiItem`，`toJobDetailViewModel` → `toJobDetailViewModelFromApi` | 是 | 简历优化页左侧栏 |

## 7. mock / fallback / 旧字段反弹检查

| 检查项 | 结果 | 说明 |
| --- | -- | -- |
| `as any` / `@ts-ignore` | 零匹配（活跃代码） | 仅 MyResumesPage（P2 死代码）残留 |
| `withFallback` / `withMockFallback` / `fallbackToMock` | 零匹配 | 无反弹 |
| mock import | 零匹配 | 无反弹 |
| `banner_summary` / `responsibilities` / `bonus_points` | 仅 ViewModel 字段名 | `JobAdapter.ts` ViewModel 定义 + `JobDetailJD.tsx` 消费，来源均为后端真实字段 |
| `business_group` / `type_label` / `short_description` (API 旧字段) | 零匹配 | 已完全移除 |

## 8. 验证结果

| 命令 | 结果 | 说明 |
| -- | -- | -- |
| `git status --short` 执行前 | 9 文件 modified | 仅前端 Detail Flow 文件 |
| `npm run build` | ✅ PASS | 1.85s，1676 modules |
| `npm run typecheck` | ✅ PASS | 零错误 |
| `go build ./cmd/server` | ✅ PASS | 编译通过 |
| `go vet ./...` | ✅ PASS | 零告警 |
| `go test ./...` | ✅ PASS | 9 packages PASS |
| 反弹检查 | ✅ | 零 as any/mock/fallback 反弹；旧 API 字段零泄露 |
| `git status --short` 执行后 | 9 文件 modified | 仅前端 Detail Flow 文件 |

## 9. 修改文件清单

| 文件 | 修改内容 | 是否业务代码 |
| -- | ---- | ------ |
| `web/src/api/jobs.ts` | `getJobDetail()` / `getSelectedJob()` 返回 `Promise<JobApiItem>`；移除旧 imports | 否（API 层） |
| `web/src/adapters/jobAdapter.ts` | 新增 `toJobDetailViewModelFromApi`；移除旧版 adapter 函数和英文 key 标签 map | 否（adapter 层） |
| `web/src/api/viewModels.ts` | `loadJobDetailViewModel` / `loadSelectedJobDetailViewModel` 改用新 adapter | 否（loader） |
| `web/src/app/pages/JobMatchAnalysisPage.tsx` | `JobDetail` → `JobApiItem` | 是 |
| `web/src/app/pages/ResumeOptimizationPage.tsx` | `JobDetail` → `JobApiItem`；`buildTargetJob` 改用 `description` + `requirements` | 是 |
| `web/src/app/components/MatchAnalysisHeader.tsx` | `JobDetail` → `JobApiItem`；`toJobDetailViewModel` → `toJobDetailViewModelFromApi` | 是 |
| `web/src/app/components/MatchAnalysisSidebarLeft.tsx` | `JobDetail` → `JobApiItem`；`toJobDetailViewModel` → `toJobDetailViewModelFromApi` | 是 |
| `web/src/app/components/ResumeOptimizationHeader.tsx` | `JobDetail` → `JobApiItem`；`toJobDetailViewModel` → `toJobDetailViewModelFromApi` | 是 |
| `web/src/app/components/ResumeOptimizationSidebarLeft.tsx` | `JobDetail` → `JobApiItem`；`toJobDetailViewModel` → `toJobDetailViewModelFromApi` | 是 |

## 10. 未做事项

* 未修改后端接口
* 未修改数据库 schema
* 未新增接口
* 未接 LLM
* 未修改 Jobs 成功业务行为
* 未修改 Auth / Resumes / Agent
* 未 git commit

## 11. 当前风险与后续建议

* **已解决**：Detail Flow 全部对齐后端 `JobApiItem` — `getJobDetail` → adapter → ViewModel 链路完整
* **已解决**：旧版 `JobDetail` 类型不再被详情页面使用（View 组件仅使用 `JobDetailViewModel`）
* **已解决**：`buildTargetJob` 不再依赖 `banner_summary`/`bonus_points` 等不存在字段
* **仍需 10-P-4-5-C 统一验收与 commit**：包括前置 10-P-4-5-A（后端索引+external_source AppError+List Flow）和本次 10-P-4-5-B（Detail Flow）两个阶段的完整验收
* **不需要处理，原因是**：后端接口不修改、数据库不改、LLM 不接、mock/fallback 不引入

## 12. docs 更新清单

| 文件 | 新增 / 更新 | 目的 |
| -- | ------- | -- |
| `docs/mcp/项目上下文.md` | 更新 | 更新当前阶段为 10-P-4-5-B |
| `docs/mcp/10-P-4-5-B-Jobs-Detail-Flow字段契约修复.md` | 新增 | 本阶段 MCP 记录 |
| `docs/09-开发日志/2026-06-02-245-10-P-4-5-B-Jobs-Detail-Flow字段契约修复.md` | 新增 | 开发日志 |

## 13. 建议 commit 信息

* 是否已 commit：否
* 建议 commit message：

```text
fix(jobs): align job detail contract with API response
```

（最终 commit 等 10-P-4-5-C 统一验收后再决定）

## 14. CodeBuddy 重启恢复摘要

10-P-4-5-B 完成 Jobs Detail Flow 字段契约修复。`getJobDetail()` / `getSelectedJob()` 返回类型改为 `JobApiItem`。适配层新增 `toJobDetailViewModelFromApi`（从后端 `description`/`requirements`/`skill_keywords` 等真实字段计算 ViewModel）。移除全部旧版 adapter 函数（`toJobCardViewModel(job: Job)` 等）和英文 key 标签映射。迁移 6 个页面/组件：`JobMatchAnalysisPage`、`ResumeOptimizationPage`、`MatchAnalysisHeader`、`MatchAnalysisSidebarLeft`、`ResumeOptimizationHeader`、`ResumeOptimizationSidebarLeft`。`buildTargetJob` 改为使用 `description` + `requirements` 真实字段。npm build/typecheck + go build/vet/test 全部通过。无 mock/fallback/as any 反弹。未 commit。建议进入 10-P-4-5-C 统一验收与 commit。

下次继续前读取：`docs/mcp/项目上下文.md`、`docs/mcp/10-P-4-5-B-Jobs-Detail-Flow字段契约修复.md`、`docs/mcp/10-P-4-5-A-Jobs主链路索引与前后端字段契约修复.md`。

## 15. 最终判断

1. Detail Flow 是否已对齐后端 API 类型？是。
2. `getJobDetail()` 返回类型是否已修复？是（`Promise<JobApiItem>`）。
3. Detail adapter 是否不再读取旧 API 字段？是（`toJobDetailViewModelFromApi` 所有字段来自 `JobApiItem`）。
4. JobDetail / MatchAnalysis / ResumeOptimization 相关页面是否已迁移？是（9 个文件全部迁移）。
5. mock / fallback 是否反弹？否。
6. 旧字段是否仍在活跃 Detail Flow 中被使用？否（API 层 `banner_summary`/`responsibilities`/`bonus_points` 等已全部移除）。
7. 验证命令是否全部通过？是（npm build/typecheck + go build/vet/test 全部通过）。
8. 是否修改后端？否。
9. 是否修改数据库？否。
10. 是否 git commit？否。
11. 本阶段是否可以通过？是。
12. 是否建议进入 10-P-4-5-C 统一验收与 commit？是。
