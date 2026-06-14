# H20-E-3R / H20-E-3R-T — Job Match Fast View Visual Balance Fix & Typography Alignment

> 2026-06-14 | 前端专项优化 | ✅ 已完成

## H20-E-3R-T: Typography Alignment (追加)

文字尺寸全量缩小对齐推荐页（JobRecommendationPage）设计标准：

| 层级 | 修改前 | 修改后 | 对齐标准 |
|------|--------|--------|---------|
| Score | `text-5xl` (48px) | `text-[36px]` (36px) | 圈从 96×96→84×84 |
| 页面标题 | `text-xl` (20px) | `text-lg` (18px) | Token #9 |
| 模块标题 | `text-base` (16px) | `text-[14px]` | 推荐页卡片标题 |
| 正文 | `text-sm leading-7` | `text-[13px] leading-6` | Token #12 |
| 辅助 | `text-xs` (12px) | `text-[11px]` | 推荐页 compact |
| 标签/徽章 | `text-xs` | `text-[11px]` | Token #14 |

**SectionCard.tsx 小修**：新增 `titleClassName` / `descriptionClassName` prop；默认 title `text-[14px]`（原 `text-base` 16px）；默认 description `text-[13px]`（原 `text-sm` 14px）。

## 问题背景

岗位匹配 Fast 页面存在以下问题：
- 核心三卡（为什么匹配/最大短板/现在怎么做）仅对 Deep 模式开放，Fast 模式跳过
- 证据与覆盖区、折叠详情区也仅对 Deep 开放
- 导致 Fast 页面内容空白严重：Hero → 空白 → ReadinessBoundary → PriorityActionPlan → AiNotice
- 页面像"没加载完整"，大面积空白

## 修复方案

### 1. 布局修复
- 外层保持 `max-w-[1440px]`，TopBar 保持全宽
- Fast 内容区新增内层容器 `max-w-[1200px] space-y-5`
- Deep 内容区保持 `space-y-6`（不变）

### 2. Fast 核心三卡（新增）
- **为什么匹配**：数据源 `screeningPassPoints` → `decisionReasons` → `summary`，最多展示 2 条，绿色勾图标
- **最大短板**：数据源 `screeningEliminationRisks` top 2，amber 色圆点，表达为"需要补强"
- **现在怎么做**：数据源 `resumeRewriteCommands` top 3，蓝色编号，短标题

### 3. Fast 修改优先级摘要（新增）
- 展示：必须 Top 1-2 + 建议 Top 1-2 + 可选 Top 1
- 紧凑标签式布局（"必须"/"建议"/"可选" badge）
- 空态：显示引导进入深度分析的文案

### 4. Fast 匹配依据摘要（新增）
- 展示筛选命中证据 Top 3
- 含标题 + 简要说明（whyItHelps）
- 数据源：`screeningPassPoints` → `decisionReasons`
- 空态：引导进入深度分析

### 5. Deep 页面完全不变
- 三卡保持原有标题和布局
- PriorityActionPlan 保持完整详细展开
- Evidence & Coverage、折叠区、面试准备等全部不变
- AI Notice 位置不变

## 改动文件

| 文件 | 修改内容 |
|------|---------|
| `web/src/app/pages/JobMatchAnalysisPage.tsx` | 新增 Fast 三卡、优先级摘要、证据摘要；提取 TopBar 为全宽；Fast 内容区新增 `max-w-[1200px]` 内层容器；PriorityActionPlan 加 `!isFast` 守卫 |
| `docs/09-开发日志/2026-06-14-Job-Match-Fast-View-Visual-Balance-Fix-H20E3R.md` | 新建 |
| `docs/变更记录.md` | 更新 |
| `docs/下一步任务.md` | 更新 |

## 构建检查

- npm build: ✅ 3.10s，零 lint 错误

## 静态扫描

- Fast 核心三卡：✅ 已确认（SectionCard ×3）
- 修改优先级：✅ 已确认（SectionCard，hasPriorityContent 守卫）
- 证据摘要：✅ 已确认（SectionCard）
- 截断风险：✅ PASS — 无 line-clamp/truncate
- Loading：✅ UnifiedAgentLoadingOverlay 保持不变
