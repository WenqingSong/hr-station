# H20-E-3：Job Match Visual Language Alignment

> 日期：2026-06-14
> 类型：页面对齐
> 编号：H20-E-3
> 前置：H20-E-1（审计）/ H20-E-2（组件）

## 对齐内容

| 项目 | 修改前 | 修改后 |
|---|---|---|
| 页面宽度 | `max-w-[1280px]` | `max-w-[1440px]` |
| loading | `AgentStepLoadingOverlay` (白灰步进) | `UnifiedAgentLoadingOverlay` (蓝调科技感) |
| 主按钮 | 手写 `bg-[#0052D9]` className | `PrimaryButton` / `SecondaryButton` / `AppButton` |
| 决策标签 | 手写 `rounded-full` span | `StatusBadge` (emerald/blue/amber/slate) |
| Hero AI 提示 | 手写 div | `AiNotice variant="inline"` |
| 底部 AI 声明 | 手写 div | `AiNotice tone="slate"` |

## 构建

| 项目 | 结果 |
|---|---|
| npm build | ✅ 1.88s |
