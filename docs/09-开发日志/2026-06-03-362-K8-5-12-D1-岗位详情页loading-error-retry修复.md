# K8-5-12-D1：岗位详情页 loading / error / retry 修复

## 一、当前问题
D-A 诊断：Banner/JD 均用内联文本 loading + 空 catch + 无 retry。

## 二、修复方案

| 组件 | Loading | Error | Retry |
|------|---------|-------|-------|
| JobDetailBanner | `PageLoading` | `PageError(onRetry)` | `retryCount` + `handleRetry` |
| JobDetailJD | `DetailSkeleton(sections=2)` | `PageError(onRetry)` | `retryCount` + `handleRetry` |

## 三、修改清单

| 文件 | 修改 |
|------|------|
| `JobDetailBanner.tsx` | 空 catch→setError；loading→PageLoading；!jobVm→PageError(onRetry)；新增 retryCount |
| `JobDetailJD.tsx` | 同上；loading→DetailSkeleton |
| `ui/DetailSkeleton.tsx` | 重建（K8-5-12-C 文件丢失） |

## 四、验收

| 检查项 | 结果 |
|--------|------|
| `npm run build` | pass（1.62s） |
| 空 catch 清除 | pass |
| Banner: PageLoading + PageError + retry | pass |
| JD: DetailSkeleton + PageError + retry | pass |
| Hero 字段不回退 | pass |
| 正文结构完整 | pass |
| API 三端点 | pass |

## 五、未修改范围
- 未统一详情页请求 / 未改后端 / 未改 Hero 字段 / 未改正文结构

## 六、Git 约束确认
本次未执行任何 Git 命令，也未提交代码。
