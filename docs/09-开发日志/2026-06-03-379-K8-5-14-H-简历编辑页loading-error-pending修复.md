# K8-5-14-H：简历编辑页 loading / error / pending 修复

## 1. 本阶段目标
修复编辑模式的 loading/error/retry 体验。保存 pending/success/failure 已有，无需修改。

## 2. 问题诊断

| 场景 | 原问题 | 修复 |
|------|--------|------|
| 编辑模式加载 | 内联文本"正在加载简历数据…" | 整页 `ResumeSkeleton`（4 卡骨架） |
| 加载失败 | `resumeError=true` 琥珀色静态提示，无 retry | `PageError(onRetry=handleRetryLoad)` + `retryCount` |
| catch 空 | `.catch(() => { setResumeError(true) })` | 改为 `.catch(() => setResumeError("简历加载失败，请稍后重试。"))` |

## 3. 修改清单

| 文件 | 修改 | 属于 K8-5-14-H |
|------|------|:--:|
| `ui/ResumeSkeleton.tsx` | **新建** 4 卡骨架 + header + 侧栏 | 是 |
| `MyResumesPage.tsx` | `resumeError：boolean→string` + `retryCount` + `handleRetryLoad` + 早期返回 skeleton/error | 是 |

## 4. 验收

| 场景 | 结果 |
|------|:--:|
| 加载中 → ResumeSkeleton | ✅ |
| 加载失败 → PageError + retry | ✅ |
| resumeError 有消息 | ✅ |
| retryCount 触发重载 | ✅ |
| 保存中 → disabled + "保存中…" | ✅ |
| 保存成功/失败提示 | ✅ |
| 无空 catch / 无 alert | ✅ |
| `npm run build` | pass |

## 5. 未修改范围
未修改后端/数据库/接口/字段名/表单模块/其他页面。保存 pending 已有，未改。

## 6. Git 约束确认
本次未执行任何 Git 命令，也未提交代码。
