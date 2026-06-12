# GQCL-Fast-Composer-F — 实现岗位推荐 Fast 输出质量控制层

## 目标
将 Fast 推荐输出正式统一经过 GQCL Fast Output Composer，确保不展示 raw code、debug tag、空字段。

## 修改内容

### 后端：fast_recommend_explainer.go

| 函数 | 修改 |
|---|---|
| `buildFastRecommendReasonExplainer` | 旧版 `"标签匹配：research、redis、algorithm"` → 新版中文完整句子（基于 coreMatch + scaledScore） |
| `buildFastDisplayTags` | 新增：将 matched tags 转为 2-4 个中文短语 |
| `clampScore` | 新增：分数裁剪工具 |

### 前端：recommendationAdapter.ts

| 字段 | 修复前 | 修复后 |
|---|---|---|
| `jobSpecificRisks` | 直接透传 `risk_points` raw codes | 经过 `humanizeRiskPoints()` |
| `resumeEvidence` | 直接透传 ID 列表 | 经过 `humanizeTagList()` 过滤 |
| `whyRankedHere` | 无 fallback | 新增 `humanizeOneSentenceReason` fallback |
| `jobSpecificReason` | 直接透传 | 经过 `humanizeOneSentenceReason()` |

### Composer 架构

```
Fast Ranker (召回/打分/排序)
        ↓
Fast Output Composer
  ├─ Score → Level (优先投递/重点关注/可以备选/谨慎尝试)
  ├─ Tags → humanize → display_tags
  ├─ Matched tags → match_highlights (完整中文句)
  ├─ Missing/risk tags → risk_summaries (完整中文句)
  ├─ Score + highlights → one_sentence_reason
  └─ Score + risk → next_action
        ↓
GQCL Safety Check (raw code 扫描)
        ↓
Frontend Safety Adapter (tagLabelMap.ts 二次防御)
```

## 验证

- go build/vet/test ✅
- npm build ✅
- make check ✅
- 前端 raw code 扫描 ✅（adapter 零命中）
- 后端 "标签匹配：" 扫描 ✅（explainer 零命中）

## 下一步
进入 Fast-Recommend-SEP-Integration-G。
