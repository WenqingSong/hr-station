# H20-E-1：Unified Visual Language Audit

> 日期：2026-06-14
> 类型：只读审计
> 编号：H20-E-1

## 核心发现

**岗位匹配页（JobMatchAnalysisPage）视觉语言基本正确但存在局部跳脱**，主要问题不是方向性错误而是缺少统一标准的约束。

## 图1-4 统一设计语言标准

| 维度 | 标准 | 来源 |
|---|---|---|
| 页面背景 | `bg-[#F5F7FB]` | ContentPageShell / RootLayout |
| 页面最大宽度 | `max-w-[1440px]` | 推荐页、岗位列表、PageState |
| 卡片背景 | `bg-white` | 全部 |
| 卡片圆角 | `rounded-2xl`（标准）/ `rounded-3xl`（大型面板） | 全部 |
| 卡片阴影 | `shadow-sm` | 全部 |
| 卡片边框 | `border border-slate-200` | 全部 |
| 主按钮 | `bg-[#0052D9] text-white text-sm rounded-xl hover:bg-[#0047BD]` | 全部 |
| 次按钮 | `bg-slate-100 text-slate-700 text-sm rounded-xl` | 匹配页已对齐 |
| 主标题 | `text-[22px] font-semibold` / `text-xl font-semibold` | 推荐页/匹配页 |
| 正文 | `text-sm leading-7` / `text-[13px]` | 匹配页更舒展 |
| 辅助 | `text-xs text-slate-400/500` | 全部 |
| Loading 遮罩 | `bg-slate-900/20 backdrop-blur-md` | 共用 |
| Loading 卡片 | 推荐页蓝色渐变科技感；匹配页白+灰步进式 | 不统一 |
| Footer | `bg-slate-50/80 border-t border-slate-200/70` | 全局 |

## 图5-7 跳脱点

| 维度 | 跳脱表现 | 影响 |
|---|---|---|
| 页面宽度 | 匹配页 `max-w-[1280px]` vs 推荐页 `max-w-[1440px]` | 略窄但可接受 |
| 内容密度 | 匹配页单列宽松（好）；推荐页三栏密集 | 无跳脱 |
| 字体系统 | 匹配页使用标准 Tailwind 字号；推荐页大量 `text-[11px/13px]` | 匹配页更规范 |
| 卡片系统 | 一致：`rounded-2xl shadow-sm border-slate-200` | ✅ |
| 蓝色使用 | `#0052D9` 一致；分数圆渐变 `from-blue-500 to-cyan-400` 略亮 | 轻微 |
| 操作区 | 按钮样式 ✅；但缺少推荐页的"返回"统一入口感 | 轻微 |
| Loading | 推荐页用 DeepAgentLoadingOverlay（科技感）；匹配页用 AgentStepLoadingOverlay（步进式） | 风格不一致 |
| AI 提示 | 匹配页 slate 调提示条 vs 推荐页无同等元素 | 无对比基准 |
| 骨架屏 | MatchAnalysisSkeleton 用 `max-w-[1440px]` 但页面用 `max-w-[1280px]` | 尺寸不匹配 |

## Loading 对齐结论

- 遮罩层一致（`bg-slate-900/20 backdrop-blur-md`）✅
- 卡片风格不统一：推荐页蓝色科技感 vs 匹配页白底灰调步进
- **建议**：统一为推荐页的 DeepAgentLoadingOverlay 风格，或至少统一卡片白底蓝调

## H20-E-2 建议

| 建议 | 是否已有 | 行动 |
|---|---|---|
| 统一 max-width 为 1440px | ❌ 匹配页 1280px | 改匹配页 |
| 统一 Loading Overlay 风格 | 两套并存 | 对齐推荐页 |
| 骨架屏 max-width 对齐 | ❌ | 修复 |
| PrimaryButton 组件 | ❌ 散落各处 | 可抽取 |

## H20-E-3 建议

- 匹配页向推荐页靠拢（max-width, loading）
- 匹配页保留字体舒展度优势（`text-sm leading-7` 优于 `text-[13px]`）
- 长内容折叠保留 ✅
