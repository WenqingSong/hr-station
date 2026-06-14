# H20-E-3R2 — Job Match Fast View Data-Aware UI Repair

> 2026-06-14 | 前端专项修复 | ✅ 已完成

## 问题根源

Fast 页面出现大量"暂无"空卡，原因是 **Fast SEP 引擎不生产 Deep 专属字段**：

| ViewModel 字段 | Fast SEP 是否填充 | 说明 |
|---------------|-------------------|------|
| `screeningPassPoints` | ❌ 空数组 | Deep 专属（Hy3 Module 2） |
| `screeningEliminationRisks` | ❌ 空数组 | Deep 专属（结构化风险） |
| `resumeRewriteCommands` | ❌ 空数组 | Deep 专属（改写建议） |
| `jobStrategy` | ❌ null | Deep 专属 |
| `screeningDecision` | ✅ 有数据 | score, level, summary, decisionLabel |

Fast 引擎自己生产 `fit_points`、`risk_points`（legacy 字符串）、`main_gaps`、`action_suggestions`、`optimization_suggestions`等原生字段，但 adapter 未映射到 ViewModel。

所以即使 Fast 有数据，前端也读不到对应的 ViewModel 字段，导致所有模块都降级为"暂无"空态。

## 修复方案：数据感知渲染

**核心原则**：有数据才展示模块，无数据不放大空卡，稀疏时用 Deep CTA 承接。

### 1. 新增数据感知标识

```ts
const hasPassPoints = screeningPassPoints.length > 0
const hasRisks = screeningEliminationRisks.length > 0
const hasCommands = resumeRewriteCommands.length > 0
const fastCardCount = (hasPassPoints ? 1:0) + (hasRisks ? 1:0) + (hasCommands ? 1:0)
```

### 2. 核心三卡 → 自适应

- `fastCardCount >= 1`: 只展示有数据的卡片，grid column 自适应（1/2/3 列）
- `fastCardCount === 0`: 不展示空三卡，改为单个轻量摘要卡（从 summary + decisionReasons 派生）

### 3. Priority / Evidence → 条件渲染

- Priority：`hasPriorityContent &&` — 无数据时完全隐藏
- Evidence：`hasPassPoints &&` — 无真实证据时完全隐藏

### 4. 新增 Deep CTA

当 Fast 数据稀疏（无 Priority、无 Evidence、无三卡）时，展示轻量 CTA：

```
想看更完整的匹配分析？
生成深度分析可获取详细的风险诊断、证据链和修改建议。
[生成深度匹配分析]
```

### 5. 宽度收口

Fast 内层容器从 1200px → 1120px，避免单列卡片过宽。

## 构建

- npm build: ✅ 3.05s，零错误
- lint: ✅ 零警告

## 静态扫描

- 空态文案：仅 Deep 保留，Fast 大空卡已消除 ✅
- 条件渲染：fastCardCount/hasPassPoints/hasRisks/hasCommands 全部就位 ✅
- 宽度：外层 1440px + 内层 1120px ✅
- 截断：零 ✅
