# H20-E-2：Minimal Unified UI Components

> 日期：2026-06-14
> 类型：组件抽取
> 编号：H20-E-2
> 前置：H20-E-1（审计）

## 抽取的组件

| 组件 | 文件 | 用途 |
|---|---|---|
| UnifiedAgentLoadingOverlay | `common/UnifiedAgentLoadingOverlay.tsx` | 统一蓝调科技感 loading（以 DeepAgentLoadingOverlay 为母版） |
| AppButton / PrimaryButton / SecondaryButton | `common/AppButton.tsx` | 统一主/次按钮 |
| SectionCard | `common/SectionCard.tsx` | 统一白色圆角卡片（`rounded-2xl shadow-sm border-slate-200/70`） |
| StatusBadge | `common/StatusBadge.tsx` | 统一状态标签（blue/emerald/amber/rose/slate） |
| AiNotice | `common/AiNotice.tsx` | 统一 AI 生成边界提示 |

## 未侵入现有页面

所有现有页面 import 保持不变：`AgentStepLoadingOverlay`、`DeepAgentLoadingOverlay` 仍正常导出。H20-E-3 负责将匹配页接入统一组件。

## 构建

| 项目 | 结果 |
|---|---|
| npm build | ✅ 2.10s |
