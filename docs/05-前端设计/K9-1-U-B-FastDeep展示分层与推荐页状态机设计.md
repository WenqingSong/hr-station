# K9-1-U-B：Fast/Deep 展示分层与推荐页状态机设计

> 状态：设计冻结 | 日期：2026-06-07 | 阶段：K9-1-U-B

## 1. K9-1-U-A 诊断结论

- P0：推荐结果只存不读，每次 POST 重新生成
- P1：Fast/Deep 展示密度完全相同
- P2：Quality profile pending → 前端超时 → "推荐失败"

## 2. 前端页面进入状态机

### 2.1 新进入流程

```
/resume?resumeId=X
  → GET /api/jobs/recommend/latest?resume_id=X
  → has_result=true?
    YES → latest.gen_tier=='quality'? Deep完整 : Fast简化
    NO  → POST /api/jobs/recommend gen_tier=fast
           → fast_done → Fast简化展示
```

### 2.2 状态定义

```ts
type RecPhase =
  | "loading_latest"           // 正在读取最新结果
  | "fast_loading"             // Fast 生成中
  | "fast_done"                // Fast 就绪（简化展示）
  | "fast_error"               // Fast 失败
  | "deep_loading"             // Deep 生成中（overlay）
  | "deep_profile_pending"     // Deep profile 未就绪
  | "deep_done"                // Deep 就绪（完整展示）
  | "deep_error"               // Deep 失败
  | "profile_failed";
```

### 2.3 activeResult 规则

```ts
const activeResult = phase === "deep_done" && deepResult
  ? deepResult
  : fastResult;
// deep_error / deep_profile_pending → fallback to fastResult
```

## 3. Fast 简化展示设计

### 3.1 左侧卡片（Fast）

```
┌─────────────────────────────────────┐
│ [高匹配]                       90分 │
│ #1 NLP工程师                        │
│ 北京 · 算法 · AI平台部              │
│                                     │
│ 该岗位与你的研究方向存在较大交集...   │ ← whyRankedHere 1行
│                                     │
│ [深度学习模型训练经验]               │ ← 1条 evidence tag
│                                     │
│ ⚠ 缺少NLP专项项目                   │ ← 1条 risk
└─────────────────────────────────────┘
```

**Fast 不展示**：多条 evidence、完整 matchHighlights、jobSpecificReason 正文。

### 3.2 中间详情（Fast）

```
一句话总结：whyRankedHere
核心匹配点：matchHighlights[0:2]
主要风险：jobSpecificRisks[0:2]
[进入岗位详情] [进行匹配分析]
```

**Fast 不展示**：jobSpecificReason / resumeEvidence 列表 / suggestedImprovement。

### 3.3 Fast 触发条件

- `latest.generation_tier === 'fast'`
- 首次进入且无 latest 结果 → Fast 生成后展示

## 4. Deep 完整展示设计

### 4.1 左侧卡片（Deep）

```
┌─────────────────────────────────────┐
│ [高匹配]                       90分 │
│ #1 NLP工程师                        │
│ 北京 · 算法 · AI平台部              │
│                                     │
│ 该岗位在NLP方向与你的简历证据匹配... │ ← whyRankedHere
│                                     │
│ [深度学习模型训练] [实验评估能力]    │ ← 2条 evidence tag
│                                     │
│ ⚠ 缺少NLP专项项目或论文             │ ← risk[0]
└─────────────────────────────────────┘
```

### 4.2 中间详情（Deep）

```
一句话总结：whyRankedHere
为什么推荐这个岗位：jobSpecificReason
简历匹配证据：resumeEvidence（列表）
匹配亮点：matchHighlights
主要风险：jobSpecificRisks（完整）
补强建议：suggestedImprovement
下一步建议：suggestedAction
```

### 4.3 Deep 触发条件

- `latest.generation_tier === 'quality'`
- 用户点击"生成深度推荐"成功

## 5. 顶部状态栏文案

| 场景 | 文案 |
|------|------|
| Fast 初始 | 快速推荐 · 规则推荐 |
| Fast LLM | 快速推荐 · AI生成 |
| Deep loading | 智能体正在深度分析 |
| Deep profile pending | 正在生成深度画像 |
| Deep done (LLM disabled) | 深度推荐 · 规则兜底 |
| Deep done (LLM) | 深度推荐 · AI生成 |
| Deep profile timeout | 深度画像生成较慢 · 已保留快速结果 |
| Deep API failure | 深度推荐生成失败 · 已保留快速结果 |

**不再统一使用**："推荐失败 · 已保留快速结果"（太严重）。

## 6. headerLabel 派生规则

```ts
function getHeaderLabel(phase: RecPhase, generationBy?: string): string {
  if (phase === "deep_loading") return "智能体正在深度分析";
  if (phase === "deep_profile_pending") return "正在生成深度画像";
  if (phase === "deep_done") {
    return generationBy?.startsWith("llm") ? "深度推荐 · AI生成" : "深度推荐 · 规则兜底";
  }
  if (phase === "deep_error") return "深度推荐生成失败 · 已保留快速结果";
  return "快速推荐 · 规则推荐";
}
```

## 7. Quality profile pending 处理流程

```
点击"生成深度推荐"
  → POST /api/jobs/recommend generation_tier=quality
  → profile_status === "ready" 且 jobs 非空
    → setPhase("deep_done"), 完整展示
  → profile_status === "pending"
    → setPhase("deep_profile_pending"), 轮询 profile/status
    → 轮询 N 次后仍 pending
      → 文案："深度画像生成较慢 · 已保留快速结果"
      → 不 setPhase("deep_error") — 保持 fastResult 展示
  → profile_status === "failed"
    → 文案："深度画像生成失败，可重试"
  → API error
    → setPhase("deep_error"), 文案："深度推荐生成失败 · 已保留快速结果"
```

## 8. 展示模式 token

```ts
type DisplayMode = "fast_compact" | "deep_full";

// derived from:
const displayMode: DisplayMode =
  phase === "deep_done" && deepResult ? "deep_full" : "fast_compact";
```

页面 JSX 用 `displayMode` 控制区块显隐，而非每个字段单独判断。

## 9. 修改文件清单（K9-1-U-D 实现阶段）

| 文件 | 修改 |
|------|------|
| `web/src/types/agent.ts` | 新增 latest result 类型 |
| `web/src/api/viewModels.ts` | 新增 `loadLatestRecommendation` + `loadRecommendationViewModel` 逻辑 |
| `web/src/adapters/recommendationAdapter.ts` | 新增 latest result adapter |
| `web/src/app/pages/JobRecommendationPage.tsx` | 进入状态机 + 展示分层 |

## 10. 构建 / 约束

- 不改后端、不改 DB、不改 init.sql
- npm build / go build / vet / check-encoding 保持通过
