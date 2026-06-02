# 2026-06-02-245-10-P-4-5-B-Jobs-Detail-Flow字段契约修复

## 一句话摘要

Jobs Detail Flow 全部对齐后端 API 类型：`getJobDetail` 返回 `JobApiItem`，新增 `toJobDetailViewModelFromApi`，迁移 9 个文件，移除全部旧版 adapter。

## 修改文件

| 文件 | 操作 |
| -- | -- |
| `web/src/api/jobs.ts` | `getJobDetail`/`getSelectedJob` 返回 `Promise<JobApiItem>` |
| `web/src/adapters/jobAdapter.ts` | 新增 `toJobDetailViewModelFromApi`；移除旧版 adapter 函数 |
| `web/src/api/viewModels.ts` | 改用 `toJobDetailViewModelFromApi` |
| `web/src/app/pages/JobMatchAnalysisPage.tsx` | `JobDetail` → `JobApiItem` |
| `web/src/app/pages/ResumeOptimizationPage.tsx` | `JobDetail` → `JobApiItem`；`buildTargetJob` 改用后端字段 |
| `web/src/app/components/MatchAnalysisHeader.tsx` | `JobDetail` → `JobApiItem` |
| `web/src/app/components/MatchAnalysisSidebarLeft.tsx` | `JobDetail` → `JobApiItem` |
| `web/src/app/components/ResumeOptimizationHeader.tsx` | `JobDetail` → `JobApiItem` |
| `web/src/app/components/ResumeOptimizationSidebarLeft.tsx` | `JobDetail` → `JobApiItem` |

## 验证

- `npm run build` ✅ 1.85s
- `npm run typecheck` ✅ 零错误
- `go build ./cmd/server` ✅
- `go vet ./...` ✅
- `go test ./...` ✅ 9 packages PASS
- 未 commit
