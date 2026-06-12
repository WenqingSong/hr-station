# 2026-06-12 JobDetail-ResumeCard-Fix-B：修复岗位详情页"当前简历"卡片空字段渲染

## 修复背景

`JobDetail-ResumeCard-Audit-A` 审计发现岗位详情页"当前简历"卡片存在 3 个 P1 + 2 个 P2 问题。类似问题也存在于匹配页和岗位广场页的简历卡片中。

## 修复内容

### JobDetailAIDecision.tsx（岗位详情页右侧栏）

| 问题 | 修复前 | 修复后 |
|---|---|---|
| 空白 Badge | `<Badge>{resumeVm?.directionLabel \|\| ""}</Badge>` 空字符串仍渲染 | `{resumeVm?.directionLabel ? <Badge>...</Badge> : null}` 条件渲染 |
| skillTags 空字符串 | 无 filter | `.filter(Boolean)` 过滤空串 |
| projectCount 误显示 0 | `{resumeVm?.projectCount ?? 0}` 项目经历` | 条件渲染：> 0 显示 "包含 N 个"，= 0 显示 "简历信息待完善" |
| updatedAt 空值 | `最近更新：{resumeVm?.updatedAt \|\| ""}` | 仅在 updatedAt 有值时显示 · 分隔 |
| 简历名称过长 | 无截断 | `line-clamp-2` 两行截断 |
| 切换按钮弱 | `text-[11px] hover:underline` | `rounded-full px-3 py-1 font-medium hover:bg-blue-50` |

### MatchAnalysisSidebarRight.tsx（匹配页右侧栏）

| 问题 | 修复前 | 修复后 |
|---|---|---|
| 空 Badge | `[directionLabel, ...skillTags.slice(0,4)]` 含空字符串 | `.filter(Boolean).slice(0,5)` 过滤后渲染 |

### AIDecisionWorkspace.tsx（岗位广场 AI 栏）

| 问题 | 修复前 | 修复后 |
|---|---|---|
| 空 Badge | `<Badge>{resumeVm.directionLabel}</Badge>` | `{resumeVm.directionLabel ? <Badge>...</Badge> : null}` |
| skillTags 空串 | 无 filter | `.filter(Boolean).slice(0,3)` |

## 构建验证

| 项目 | 结果 |
|---|---|
| `npm build` (vite) | ✅ 通过（1777 modules，2.27s） |
| `go test` | ✅ 全部通过 |
| `make check` | ✅ 全部通过 |

## 未处理项

| 项目 | 原因 |
|---|---|
| 抽 CurrentResumeCard 公共组件 | 本阶段只做最小修复，后续重构任务 |
| 统一卡片圆角 | 涉及面广，单独阶段处理 |
