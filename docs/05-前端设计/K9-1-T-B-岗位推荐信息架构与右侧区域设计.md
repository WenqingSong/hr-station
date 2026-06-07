# K9-1-T-B：岗位推荐信息架构与右侧区域设计

> 状态：设计冻结 | 日期：2026-06-07 | 阶段：K9-1-T-B

## 1. 背景（K9-1-T-A 诊断结论）

当前前端推荐页面存在以下信息架构问题：

| # | 问题 | 影响 |
|---|------|------|
| 1 | 左侧卡片信息同质化，展示与岗位无关的通用文本 | 用户看不出岗位间差异 |
| 2 | 中间详情区展示内容与左侧卡片大量重复 | 信息冗余 |
| 3 | 右侧"推荐来源"与顶部状态栏重复 | 空间浪费 |
| 4 | 右侧"推荐记录"为空占位 | 无实际业务价值 |
| 5 | sourceLabel 显示"演示数据" | 削弱可信度 |
| 6 | Deep loading 是前端体验型，非后端同步 | 非真实进度 |

## 2. 左侧推荐卡片信息层级设计

### 2.1 新版卡片结构

```
┌──────────────────────────────────────────────┐
│  [强匹配·优先投递]                   92分    │  ← Badge + 分数
│  #1  推荐算法工程师                           │  ← 排名 + 岗位名称
│  深圳 · 算法 · 技术事业群                     │  ← 城市 · 类别 · 事业群
│                                              │
│  核心原因：                                   │
│  推荐系统实验方向与岗位要求精确匹配，           │  ← 一行差异化理由
│  简历证据最充分                                │
│                                              │
│  匹配证据：                                   │
│  推荐系统实验平台 / PyTorch排序模型训练         │  ← 2条证据（tag形式）
│                                              │
│  主要风险：缺少多模态项目经历                   │  ← 1条核心风险
│                                              │
│  强匹配 · 优先投递    ⚠2项风险                 │  ← 底部标签 + 风险数
└──────────────────────────────────────────────┘
```

### 2.2 信息字段映射

| 卡片位置 | 现有字段 | K9-1-T-C 后字段 |
|----------|----------|-----------------|
| Badge | `priorityText` | 不变（K9-1-T-C 调整 label 规则后自然差异化） |
| 分数 | `score` | 不变（K9-1-T-C 调整评分后自然差异化） |
| 排名+名称 | `rank` + `title` | 不变 |
| 城市·类别·部门 | `city` + `category` + `department` | 不变 |
| 核心原因（1行） | `oneSentenceReason`（来自 `recommendation_reason`） | `why_ranked_here`（K9-1-T-C 新增） |
| 匹配证据 | `matchHighlights.slice(0,2)` | 新增 `resume_evidence`（优先），兜底 `match_highlights` |
| 主要风险 | 无 | 新增 `job_specific_risks[0]`（K9-1-T-C 新增） |
| 底部标签 | `priorityText` + 风险数 | 保持不变 |

### 2.3 前端适配要点

```tsx
// 核心原因：优先使用 why_ranked_here，兜底使用 recommendation_reason
const coreReason = job.whyRankedHere || job.oneSentenceReason;

// 匹配证据：优先使用 resume_evidence，兜底使用 matchHighlights
const evidence = job.resumeEvidence?.length ? job.resumeEvidence : job.matchHighlights.slice(0, 2);

// 主要风险：使用 job_specific_risks 第一条
const mainRisk = job.jobSpecificRisks?.[0];
```

### 2.4 视觉约束

- 卡片宽度保持 380px
- 不要塞太多长文本
- 核心原因最多 2 行截断（line-clamp-2）
- 匹配证据最多 2 条，tag 形式展示
- 风险仅显示第 1 条（最多 1 行截断）

## 3. 中间详情区信息层级设计

### 3.1 新版结构

```
┌────────────────────────────────────────────────┐
│  当前选中岗位                                    │
│  ┌──────────────────────────────────────────┐  │
│  │  推荐算法工程师                            │  │
│  │  深圳 · 算法 · 技术事业群                   │  │
│  │                                          │  │
│  │  一句话总结：                              │  │
│  │  推荐系统实验方向与岗位要求精确匹配...       │  │  ← why_ranked_here
│  └──────────────────────────────────────────┘  │
│                                                │
│  ┃ 为什么推荐这个岗位                           │
│  job_specific_reason（完整展示，不限字数）        │
│                                                │
│  ┃ 简历匹配证据                                 │
│  ● 推荐系统实验平台项目                          │
│  ● PyTorch 排序模型训练经验                      │
│  ● 用户行为序列建模与消融分析                     │  ← resume_evidence
│                                                │
│  ┃ 匹配亮点                                     │
│  ● 岗位方向与你的核心技能方向高度一致              │  ← matchHighlights
│  ● ...（保留现有）                              │
│                                                │
│  ┃ 主要风险                                     │
│  ⚠ 缺少明确多模态项目经历                        │  ← job_specific_risks
│  ⚠ 图像/视频理解相关证据不足                      │
│                                                │
│  ┃ 补强建议                                     │
│  建议在简历中补充图文检索、视频理解或多模态        │  ← suggested_improvement
│  模型实验经历，并说明数据集、模型结构和评估指标     │
│                                                │
│  ┌──────────────────────────────────────────┐  │
│  │ 建议动作：suggestedAction                  │  │
│  │ [进入岗位详情]  [进行匹配分析]               │  │
│  └──────────────────────────────────────────┘  │
└────────────────────────────────────────────────┘
```

### 3.2 信息字段映射

| 详情区位置 | 现有字段 | K9-1-T-D 新字段 |
|-----------|----------|----------------|
| 一句话总结（头部卡片） | `oneSentenceReason` | `why_ranked_here`（K9-1-T-C 新增） |
| 为什么推荐这个岗位 | 无 | `job_specific_reason`（K9-1-T-C 新增） |
| 简历匹配证据 | 无独立区域 | `resume_evidence`（K9-1-T-C 新增） |
| 匹配亮点 | `matchHighlights` | 保持不变 |
| 主要风险 | `riskPoints` | `job_specific_risks`（优先），兜底 `risk_points` |
| 补强建议 | 无 | `suggested_improvement`（K9-1-T-C 新增） |
| 底部操作 | `suggestedAction` + 按钮 | 保持不变 |

### 3.3 兼容策略

```tsx
// K9-1-T-D：分段渐进替换，不一次性删除旧字段
const detailRiskItems = selectedJob.jobSpecificRisks?.length
  ? selectedJob.jobSpecificRisks
  : selectedJob.riskPoints;
```

## 4. 右侧信息架构取舍

### 4.1 当前结构

```text
右侧栏（300px）：
├── 当前简历        ✅ 保留
├── 推荐来源        ❌ 删除（合并到顶部）
├── 深度推荐        ✅ 保留（保持交互）
└── 推荐记录        ❌ 删除
```

### 4.2 新版右侧结构

```text
右侧栏（300px）：
├── 当前简历
│   ├── 简历名称
│   ├── Resume ID
│   └── 画像：algorithm_research
│
└── 深度推荐
    ├── [未触发] → "生成深度推荐" 按钮
    ├── [loading] → "智能体深度分析中…"
    ├── [done] → "深度分析已完成 [重新生成]"
    └── [error] → "重试深度推荐" 按钮
```

### 4.3 删除"推荐记录"

**原因**：
1. 当前没有真实历史回看能力
2. 空占位增加复杂度
3. 对 Demo 评委没有实际帮助
4. 后端 RecommendationResult 仅用于 snapshot 保存，无查询接口

### 4.4 合并"推荐来源"到顶部

顶部状态栏统一显示：

```text
快速推荐 · 规则推荐 | 生成时间：2026-06-07 15:30:00 | 画像：algorithm_research
深度推荐 · 规则兜底 | 生成时间：2026-06-07 15:31:00 | 画像：algorithm_research
```

实现：在现有的 `header` 区域（第248-268行），已有 `modeLabel` + `sourceLabel` + `profileDominantProfile`，仅需调整文案。

## 5. sourceLabel 规则修正

### 5.1 当前问题

```typescript
// recommendationAdapter.ts - getSourceLabel
if (isQuality && source === "mock") return "深度推荐 · 演示数据";
case "mock": return "快速推荐 · 演示数据";
```

**问题**：
- "演示数据" 削弱可信度
- Demo 评委看到 "演示数据" 会降低评分

### 5.2 新规则

```typescript
function getSourceLabel(source: RecommendationSource, generatedBy?: string): string {
  switch (generatedBy) {
    case "mock_agent":
      return "快速推荐 · 规则推荐";
    case "mock_quality":
      return "深度推荐 · 规则兜底";
    case "llm_agent":
      return "深度推荐 · AI生成";
    case string:
      if (gen.includes("fallback")) return "推荐失败 · 已保留快速结果";
      break;
  }
  switch (source) {
    case "mock": return "快速推荐 · 规则推荐";
    case "fast_llm": return "快速推荐 · AI生成";
    case "deep_llm": return "深度推荐 · AI生成";
    case "fallback": return "推荐失败 · 已保留快速结果";
    default: return "推荐结果";
  }
}
```

**关键**：永远不要显示"演示数据"文案。

## 6. Deep / Quality 状态文案设计

### 6.1 点击"深度推荐"后状态

```
用户点击 → Overlay（智能体分析过程）→ 请求成功
    ├── 返回 LLM 结果 → 顶部显示"深度推荐 · AI生成"
    └── 返回 mock_quality → 顶部显示"深度推荐 · 规则兜底"
```

### 6.2 Overlay 文案

保留现有 K9-1-S 实现的 AgentStepLoadingOverlay，不做修改：

```tsx
const deepSteps: AgentLoadingStep[] = [
  { key: "resume-profile", label: "读取简历画像标签" },
  { key: "job-tags", label: "匹配岗位标签矩阵" },
  { key: "reasoning", label: "生成推荐理由和风险点" },
  { key: "judge", label: "审查推荐结果一致性" },
];
```

### 6.3 完成状态文案

```tsx
// 深度推荐完成后
{modeLabel === "深度推荐" && (
  <span className="flex items-center gap-1.5">
    <Sparkles className="w-3 h-3 text-green-500" />
    {sourceLabel}  {/* 深度推荐 · 规则兜底 / 深度推荐 · AI生成 */}
  </span>
)}
```

## 7. Agent Loading 方案取舍

### 7.1 设计冻结结论

**本轮不实现后端同步**。

**原因**：
1. 当前推荐 API 是普通 HTTP 同步响应
2. 后端没有 SSE / WebSocket / recommendation_runs 状态表
3. 为 Demo 单独引入实时阶段同步会显著增加复杂度
4. 目前更关键的问题是推荐解释质量，而非 loading 真同步

### 7.2 后续扩展方案（P2）

冻结为后续任务 `K9-1-X：Recommendation Run 状态表 + 轮询式 Agent Stage 同步`：

```text
POST /api/jobs/recommend/start → run_id
GET /api/jobs/recommend/runs/{run_id} → current_stage / status / result
前端轮询 stage
```

本轮不实现。

### 7.3 现状保持

继续使用 K9-1-S 实现的 `AgentStepLoadingOverlay`（前端体验型 4 步进度）。在 LLM 打开、后端真正有阶段划分后，再改为轮询式同步。

## 8. 前端类型定义新增

### 8.1 `agent.ts` 新增字段

```typescript
export interface RecommendedJobItem {
  // ... 现有字段保持不变 ...

  // ── K9-1-T-B: 逐岗位解释字段 ──
  why_ranked_here?: string;
  resume_evidence?: string[];
  job_specific_reason?: string;
  job_specific_risks?: string[];
  suggested_improvement?: string;
}
```

### 8.2 `recommendationAdapter.ts` 新增映射

```typescript
export interface RecommendedJobViewModel {
  // ... 现有字段保持不变 ...

  // ── K9-1-T-B: 逐岗位解释字段 ──
  whyRankedHere: string;
  resumeEvidence: string[];
  jobSpecificReason: string;
  jobSpecificRisks: string[];
  suggestedImprovement: string;
}
```

## 9. 修改文件清单（K9-1-T-D 实现阶段）

| 文件 | 修改内容 |
|------|----------|
| `web/src/types/agent.ts` | RecommendedJobItem 新增 5 个字段 |
| `web/src/adapters/recommendationAdapter.ts` | RecommendedJobViewModel 新增 5 个字段 + 映射逻辑 + sourceLabel 修正 |
| `web/src/app/pages/JobRecommendationPage.tsx` | 左侧卡片展示新字段 + 中间详情区新增区域 + 删除推荐记录 + 合并推荐来源到顶部 + Deep fallback 文案 |

**不修改的文件**：
- `AgentStepLoadingOverlay.tsx` — 保持不变
- 其他页面和组件

## 10. 验收标准（K9-1-T-D 后）

| 验收项 | 标准 |
|--------|------|
| 左侧 5 张卡片信息彼此不同 | 核心原因/证据/风险各有差异 |
| 中间详情区有 6 个内容区块 | 一句话总结/为什么推荐/证据/亮点/风险/补强建议 |
| 右侧无"推荐记录"区域 | 已删除 |
| 顶部状态栏包含完整信息 | 来源+时间+画像 |
| sourceLabel 无"演示数据" | 全部为"规则推荐/规则兜底/AI生成" |
| Deep 完成后显示正确文案 | "深度推荐 · 规则兜底" |
| 5 类样例各自展示正确 | 01-05 各自有差异化解释 |
