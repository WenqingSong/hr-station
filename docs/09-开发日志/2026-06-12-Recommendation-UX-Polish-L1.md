# 2026-06-12 Recommendation-UX-Polish-L1 — 岗位推荐页 P0/P1 展示问题修复

## 任务性质

前端产品化打磨（第一轮），不修改后端算法和数据库。

## 修复目标

1. Fast 推荐中 raw code / 乱码 / 空 badge / undefined/null 等低级展示问题
2. 扩展 tagLabelMap，保证常见后端 code 不会裸露
3. recommendationAdapter 字段去重，避免同一段内容在多个模块重复展示
4. Deep Loading 弹窗增加预估时间，步骤文案对齐真实 Deep Cascade
5. Deep Loading 超时从 90s 调整为 900s
6. 右侧"生成深度推荐"按钮改为正式 outline button
7. 左侧 Top5 卡片降噪，限制 badge 数量

## 修改文件

| 文件 | 修改内容 |
|---|---|
| `web/src/adapters/tagLabelMap.ts` | 扩展 TAG_LABEL_MAP 从 33→120+ 条目；增加 isInvalidValue 函数；humanizeMatchHighlights/humanizeRiskPoints 去除强制兜底文案 |
| `web/src/adapters/recommendationAdapter.ts` | 新增 normalizeText/isSameText/dedupeTextList/isMeaningfulText 工具函数；toRecommendedJobViewModel 中实现字段去重逻辑 |
| `web/src/app/components/common/AgentStepLoadingOverlay.tsx` | maxDurationMs 90s→900s；增加 estimatedTimeText/onCancel/cancelLabel props；footer 增加预估时间和取消按钮渲染 |
| `web/src/app/pages/JobRecommendationPage.tsx` | deepSteps 4 步文案对齐 Deep Cascade Layer 1-4；增加 DEEP_ESTIMATED_TIME；左侧卡片限制 evidence≤2/tags≤3+ overflow count；右侧按钮改为 outline 设计；中间详情区优化空字段展示 |

## Fields Deduplication Strategy

| 字段组 | 修复前 | 修复后 |
|---|---|---|
| recommendationReason / jobSpecificReason | 可能出现相同 LLM 输出 | 若 humanize 后相同，jobSpecificReason 置空，仅保留 oneSentenceReason |
| riskPoints / jobSpecificRisks | 相同数组可能重复渲染 | riskSet 比对后，若全覆盖则清空 riskPoints |
| suggestedAction / applicationStrategy | 同一后端字段 | 区分优先级：suggestedImprovement≠suggestedAction 时保留 improvement |
| matchHighlights / riskPoints 兜底 | 空数组强显示 boilerplate | 去除兜底文案，页面自行处理空状态 |

## Deep Loading Enhancement

- 步骤：`Layer 1 候选发现` → `Layer 2 语义重排` → `Layer 3 逐岗分析` → `Layer 4 全局校准`
- 预估时间：`预计耗时约 1-2 分钟；若系统正在首次构建岗位画像，可能需要更长时间。`
- 超时：900s（覆盖完整 Deep Cascade 全链路）
- 取消：可取消返回快速推荐结果

## Left Cards Simplification

- evidence badges: max 2（原 1-2） 
- displayTags: max 3，超出显示 +N 计数
- 去除卡片中的风险单行文本、长段补强建议

## Build Verification

```
npm --prefix web run build → ✅ passed (1.96s)
```

Raw code scan: ✅ 零 raw code 泄露（仅在 tagLabelMap.ts 定义映射中出现，渲染逻辑无裸露）
