# 2026-06-12 Frontend-Design-System-Sweep-C：全站前端成熟化设计统一

## 修复背景

`Frontend-Layout-Fix-B` 已修复四核心 AI 页面的 navbar 间距问题，但全站仍存在设计碎片化问题：
- 页面背景色不一致（RootLayout `#F5F7FA` vs ContentPageShell `#F5F7FB`）
- PageError/PageEmpty 视觉简陋（仅文字文本，无卡片容器）
- PageError 缺少 `onRetry` 属性导致 JobList 重试按钮无人实现
- PageEmpty 缺少 `actionLabel`/`onAction` 属性
- 用户可见的 mock/fallback 标签不够产品化

## 设计统一规则

| 类别 | 规则 |
|---|---|
| 页面背景 | `#F5F7FB`（全站统一） |
| 卡片样式 | `rounded-2xl border border-slate-200 shadow-sm bg-white` |
| 按钮主色 | `#0052D9`（腾讯蓝），hover `#003DA5` |
| 错误态 | `PageError` — 白色卡片 + 红色图标容器 + 标题/描述 + 操作按钮 |
| 空状态 | `PageEmpty` — 白色卡片 + 蓝色图标容器 + 标题/描述 + 操作按钮 |
| Loading | `PageLoading` — 居中 spinner + 加载文案 |

## 修改内容

### PageState.tsx（核心提升）

**PageError 增强**：
- 新增 `onRetry` 属性（可选），传入后自动显示"重新尝试"蓝色主按钮
- 视觉升级为白色卡片容器：`rounded-2xl border border-slate-200 shadow-sm`
- 图标容器：红色背景（`bg-red-50`）+ 红色图标（`text-red-400`）
- 标题从 `<p>` 改为 `<h2>`，`text-base font-semibold text-slate-900`

**PageEmpty 增强**：
- 新增 `actionLabel` 和 `onAction` 属性，自动显示操作按钮
- 同白色卡片容器
- 图标容器：蓝色背景（`bg-blue-50`）+ 蓝色图标（`text-blue-400`）
- 同意标题化处理

### routes.tsx

- RootLayout 背景色从 `bg-[#F5F7FA]` 统一为 `bg-[#F5F7FB]`（与 ContentPageShell/AppBootLoading/ErrorBoundary 一致）

### recommendationAdapter.ts

- `mock_agent` label：`"规则推荐"` → `"基础模型"`
- `mock_quality` label：`"规则兜底"` → `"基础模型"`
- `fallback` label：`"推荐失败 · 已保留快速结果"` → `"推荐生成失败 · 已保留当前结果"`
- `llm_agent` label：`"AI生成"` → `"AI 生成"`（加空格）

### common.ts

- `mock_agent/mock_quality` badge：`"规则兜底"` → `"基础模型"`
- `fallback` badge：`"分析异常 - 已降级"` → `"分析结果降级"`（tone 从 red 改为 orange，更温和）

## 修复效果

| 修复前 | 修复后 |
|---|---|
| PageError 为纯文本 + 图标 | 白色卡片 + 图标容器 + 标题 + 操作按钮 |
| PageEmpty 为纯文本 + 图标 | 白色卡片 + 图标容器 + 标题 + 操作按钮 |
| RootLayout 背景与 ContentPageShell 不一致 | 统一为 `#F5F7FB` |
| 用户看到"规则兜底"等技术术语 | 显示"基础模型"等产品化表述 |
| JobList 的 retry 按钮无法渲染 | 现在正常工作 |

## 构建验证

| 项目 | 结果 |
|---|---|
| `npm build` (vite) | ✅ 通过 |
| `go build` | ✅ 通过 |
| `go vet` | ✅ 通过 |
| `go test` | ✅ 全部通过 |
| `make check` | ✅ 全部通过 |
| `make prod-check` | ❌ 缺少 `.env.prod` + LLM key（环境问题，非本次改动引入） |

## 遗留问题

- `JobList.tsx`、`JobDetailBanner.tsx`、`JobDetailJD.tsx` 使用 `PageError`/`PageEmpty` 时传递了不存在的 props（`onRetry`/`actionLabel`/`onAction`），现在已通过增强组件修复
- `PageContainer` 组件未使用（定义为 `max-w-[1360px] py-6`），保留不做删除
- 卡片圆角仍有多样性（`rounded-lg`/`rounded-xl`/`rounded-2xl`），属于各页面内部组件设计，暂不统一
