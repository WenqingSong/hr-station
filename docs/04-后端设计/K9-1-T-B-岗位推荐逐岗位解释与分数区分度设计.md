# K9-1-T-B：岗位推荐逐岗位解释与分数区分度设计

> 状态：设计冻结 | 日期：2026-06-07 | 阶段：K9-1-T-B

## 1. K9-1-T-A 诊断结论

K9-1-T-A 只读诊断确认当前推荐排序方向已基本正确，但存在以下核心问题：

| # | 问题 | 严重度 |
|---|------|:---:|
| 1 | Top5 推荐卡片高度同质化 | 高 |
| 2 | score 大量触达 95 cap，分数无区分度 | 高 |
| 3 | recommendation_label 大量都是"强匹配，优先投递" | 高 |
| 4 | recommendation_reason 模板化生成 | 高 |
| 5 | match_highlights 通用标签生成 | 中 |
| 6 | risk_points 兜底模板 | 中 |
| 7 | suggested_action 只有轻微差异 | 中 |
| 8 | 右侧"推荐来源"与顶部状态栏重复 | 低 |
| 9 | 右侧"推荐记录"空占位，无真实业务能力 | 低 |
| 10 | Quality 在 LLM disabled 下是 fallback，非真正深度推荐 | 中 |
| 11 | AgentStepLoadingOverlay 当前是前端体验型进度，非后端真实阶段同步 | 低 |

**核心根因**：岗位推荐缺少逐岗位解释能力。系统已知道"大方向适合"，但未解释清楚每个岗位的具体匹配与差异。

## 2. 同质化根因分析

### 2.1 分数同质化

```text
score = 55 + tagMatrixScore (≤37) + evidence bonuses (≤20) + education bonuses (≤8)
       + top spread (5-i, i<5)
clamp: min(55), max(95)
```

当前问题：
- tagMatrixScore 在算法方向上，多个算法岗位得分差异很小
- 5 个强有力的 evidence bonus（project/internship/research/school/degree）使得高分岗位大量涌入 90+
- top spread 直接让 Top1-5 全线逼近 95 cap
- cap=95 过于宽松，无法区分同为"强匹配"的 Top1 和 Top5

### 2.2 文案同质化

`buildReasonV2` 仅依据 dirFit 产出 4 种模板：

```go
dirFit=="strong" → "你的{画像}画像与该岗位高度匹配，技能积累和项目经验可直接复用。"
dirFit=="medium"  → "你的{画像}背景与{类别}岗位有一定关联，技能可部分迁移。"
```

同为 `dirFit=="strong"` 的多个算法岗位（多模态/NLP/CV/数据科学），reason 完全相同。

`buildHighlightsV2` 基于通用标签组合产出，不区分岗位方向子类型。

`buildRisksV2` 兜底模板 `"建议补充与该岗位要求更匹配的项目细节"` 对用户无行动价值。

### 2.3 Label 同质化

`recommendationLabelV2` 条件：

```go
score >= 90 && dirFit == "strong" && tFit == "strong" && risk != "high"
→ "强匹配，优先投递"
```

由于 Top5 算法岗位分数均 ≥90、dirFit=strong、tagFit=strong、risk=low，全部获得 strong label。

## 3. 新增逐岗位解释字段

### 3.1 RecommendedJob 新增字段

```go
type RecommendedJob struct {
    // ... 现有字段保持不变 ...

    // ── K9-1-T-B: 逐岗位解释字段 ──
    WhyRankedHere      string   `json:"why_ranked_here,omitempty"`      // 为什么排在这个名次
    ResumeEvidence     []string `json:"resume_evidence,omitempty"`     // 简历匹配证据
    JobSpecificReason  string   `json:"job_specific_reason,omitempty"`  // 逐岗位推荐理由
    JobSpecificRisks   []string `json:"job_specific_risks,omitempty"`   // 逐岗位风险
    SuggestedImprovement string `json:"suggested_improvement,omitempty"` // 补强建议
}
```

### 3.2 字段说明

| 字段 | 类型 | 说明 | 生成要求 |
|------|------|------|----------|
| `why_ranked_here` | string | 为什么该岗位排在当前名次 | 逐岗位生成，不能相同；引用岗位方向和候选人证据 |
| `resume_evidence` | []string | 从简历中抽取的匹配证据 | 每岗位 2-3 条；来自简历画像/项目/实习/技能/研究方向；不要空泛写"能力较强" |
| `job_specific_reason` | string | 逐岗位推荐理由 | 与岗位 title/category/direction 强相关；不同算法岗位之间也要不同；不能只写"画像高度匹配" |
| `job_specific_risks` | []string | 逐岗位风险 | 风险因岗位而不同；不能总是"建议补充更匹配的项目细节"；要对学生有行动价值 |
| `suggested_improvement` | string | 补强建议 | 可执行；和岗位风险对应；不要只写"建议继续提升" |

### 3.3 示例输出

以算法研究生简历为例：

**Top1：算法-多模态方向（92分）**

```json
{
  "why_ranked_here": "该岗位在算法方向、推荐系统实验和模型评估能力上与简历证据匹配度最高，因此排在 Top1。",
  "resume_evidence": [
    "推荐系统实验平台项目（PyTorch + 排序模型训练）",
    "用户行为序列建模与消融分析（AUC/NDCG评估）",
    "深度学习与特征工程项目经验"
  ],
  "job_specific_reason": "多模态岗位需要模型实验和跨模态理解能力，你的推荐系统实验和深度学习基础具备迁移价值，但简历中还缺少图文或视频理解项目。",
  "job_specific_risks": [
    "缺少明确多模态项目经历",
    "图像/视频理解相关证据不足"
  ],
  "suggested_improvement": "建议在简历中补充图文检索、视频理解或多模态模型实验经历，并说明数据集、模型结构和评估指标。"
}
```

**Top2：算法-自然语言处理方向（90分）**

```json
{
  "why_ranked_here": "NLP方向与你的研究背景有较大交集，但匹配证据强度略低于Top1推荐系统方向。",
  "resume_evidence": [
    "深度学习模型训练经验",
    "用户行为序列建模（可迁移至序列标注/语言模型）"
  ],
  "job_specific_reason": "NLP岗位与推荐系统在模型训练和数据实验上有共通方法论，但你目前缺少NLP专项项目（如文本分类、命名实体识别、预训练语言模型微调）。",
  "job_specific_risks": [
    "缺少NLP专项项目或论文",
    "未涉及Transformer/预训练语言模型（Bert/GPT等）"
  ],
  "suggested_improvement": "建议增加至少1个NLP方向项目（如文本分类/情感分析/问答系统），展示对Transformer架构的实际使用经验。"
}
```

**Top5：算法-数据科学方向（82分）**

```json
{
  "why_ranked_here": "数据科学方向与算法研究有一定关联，但简历中以模型实验为主，数据工程和业务分析证据偏弱。",
  "resume_evidence": [
    "数据分析基础（Pandas/NumPy/SQL）",
    "模型评估方法论"
  ],
  "job_specific_reason": "数据科学岗位需要较强的数据工程和业务分析能力，你的简历在模型实验方面扎实，但在ETL、数据管道和业务指标分析方面的证据不足。",
  "job_specific_risks": [
    "缺少数据工程和pipeline项目",
    "缺少业务分析和A/B测试经验"
  ],
  "suggested_improvement": "建议增加数据清洗/ETL项目，或展示在业务场景中使用数据驱动决策的经历。"
}
```

## 4. 分数区分度规则

### 4.1 当前问题

```text
Top5 全部 95 分（撞 cap），实际得分差异被抹平。
```

### 4.2 新评分规则

```
final_score = 50
            + tag_matrix (≤37，保持现有权重矩阵)
            + evidence_bonus (≤15，收紧上限)
            + direction_precision_bonus (≤8，新增方向精度奖分)
            - risk_penalty (≤12，新增风险罚分)
            - conflict_penalty (≤10，现有冲突降分)
clamp: [55, 95]
```

#### 4.2.1 evidence_bonus 收紧

| 条件 | 旧 | 新 |
|------|----|-----|
| project_evidence | +6 | +4 |
| internship_evidence | +5 | +3 |
| research_evidence | +5 | +3 |
| school_level ≥ 985 | +4 | +3 |
| degree_level ≥ 硕士 | +4 | +2 |

总上限从 ~20 降至 ≤15，减少高分堆积。

#### 4.2.2 新增 direction_precision_bonus

当候选人简历中的 domain tag 与岗位 domain tag 精确对应时加分：

| 候选人 tag | 岗位 tag | 加分 |
|-----------|----------|:---:|
| recommendation_system | recommendation_system | +8 |
| nlp | nlp | +8 |
| multimodal | multimodal | +8 |
| search_algorithm | search_algorithm | +8 |
| data_mining | data_science | +5 |
| machine_learning_platform | 任意算法岗位 | +3 |

此加分替代旧的 top spread，使得精确匹配的方向获得更高分，而非同质化加分。

#### 4.2.3 新增 risk_penalty

| 条件 | 扣分 |
|------|:---:|
| dirFit="weak" 但 score>78 | 扣至≤74 |
| 岗位要求 specific domain experience 但简历缺少 | -4 |
| 岗位偏工程但简历偏研究 | -3 |
| 缺少岗位关键技能证据 | -3 |

#### 4.2.4 去除 top spread

去掉 `score += (5 - i)` 的粗暴加分。排序仅依赖真实评分+profile-aware tie-break。

### 4.3 预期分差

对于算法研究生（algorithm_research），5 个算法方向岗位的预期分数分布：

```text
Top1: 算法-推荐系统/多模态（精确方向匹配+高证据）  → 92-94
Top2: 算法-NLP（方向有关联+中等证据）              → 88-91
Top3: 算法-数据科学（方向较近+较少证据）             → 83-86
Top4: 算法-计算机图形学（方向有一定差距）            → 78-82
Top5: AI应用开发（方向可迁移但偏离研究主线）         → 75-80
```

关键约束：
- 不去强制硬编码排名分数，但通过精密加分+风险扣分实现自然梯度
- 同方向岗位之间，根据岗位细分差异（precision_bonus + risk_penalty）拉开分

## 5. Label 约束规则

### 5.1 新规则

| Label | priority | 触发条件 | 数量约束 |
|-------|----------|----------|:---:|
| strong | 强匹配，优先投递 | score≥90 AND dirFit="strong" AND tagFit="strong" AND risk="low" AND precision_bonus≥5 | ≤2个 |
| high | 高匹配，重点关注 | score≥82 AND dirFit≠"weak" AND tagFit≠"weak" AND risk≠"high" | 无限制 |
| try | 可尝试，需要补强 | score≥72 AND dirFit≠"conflict" | 无限制 |
| backup | 备选参考 | score≥60 | 无限制 |
| not_recommended | 不推荐 | score<60 OR dirFit="weak" AND tagFit="weak" | 无限制 |

### 5.2 关键变更

1. **strong 上限**：同一个推荐列表中 strong 不超过 2 个
2. **风险硬降级**：如果 risk="medium" 且 precision_bonus<5，最多 high（不能 strong）
3. **方向证据不足**：缺少具体方向证据时最多 try
4. **Quality fallback 标签克制**：Quality fallback 时不能表现得比 Fast 更"神奇"

### 5.3 实现策略

在 `buildResultsV2` 遍历后增加 post-processing：

```go
// Post-process: enforce strong count ≤ 2
strongCount := 0
for i := range results {
    if results[i].RecommendationPriority == "strong" {
        strongCount++
        if strongCount > 2 {
            results[i].RecommendationLabel = "高匹配，重点关注"
            results[i].RecommendationPriority = "high"
        }
    }
}
```

## 6. Content Generator 升级

### 6.1 新增 Builders

```go
func buildWhyRankedHere(rank int, cp *candidateTagProfile, sj scoredJobV2, allRanked []scoredJobV2) string
func buildResumeEvidence(cp *candidateTagProfile, sj scoredJobV2) []string
func buildJobSpecificReason(cp *candidateTagProfile, sj scoredJobV2) string
func buildJobSpecificRisks(cp *candidateTagProfile, sj scoredJobV2) []string
func buildSuggestedImprovement(cp *candidateTagProfile, sj scoredJobV2) string
```

### 6.2 现有 Builder 保持兼容

- `buildReasonV2` → 仅映射到 `RecommendationReason`（一句话概述），兼容现有前端
- `buildHighlightsV2` / `buildRisksV2` / `buildActionV2` → 保留兼容，但新增字段优先使用

## 7. Deep / Quality 状态设计

### 7.1 LLM disabled 现状

```text
当前环境 LLM disabled → Quality Pipeline 走 mock_quality fallback
qualityMode=true 仅触发 Judge 审查 + 潜在 rerank
Quality 与 Fast 使用同一个 tagMatrixScore 引擎
```

### 7.2 设计冻结

| 场景 | generated_by | 前端显示 |
|------|--------------|----------|
| Fast (LLM enabled) | llm_agent | 快速推荐 · AI生成 |
| Fast (LLM disabled) | mock_agent | 快速推荐 · 规则推荐 |
| Quality (LLM enabled) | llm_agent | 深度推荐 · AI生成 |
| Quality (LLM disabled) | mock_quality | 深度推荐 · 规则兜底 |
| Fallback | mock_agent | 推荐失败 · 已保留快速结果 |

**约束**：
- 如果 Quality 与 Fast 完全相同（当前 LLM disabled 状态），在设计文档中明确标记为已知限制
- 不做假数据伪装成"AI深度生成"
- 后续打开 LLM 后再专项验收 K9-1-T-E：LLM Quality 深度推荐真实验收

## 8. 后续阶段拆分

### K9-1-T-C：后端逐岗位解释与分数区分度实现

| 范围 | 说明 |
|------|------|
| 修改 `recommendation_mock.go` | 新增 5 个 explanation builder + 调整评分公式 |
| 修改 `recommendation.go` (model) | 新增 5 个字段 |
| 修改 content generators | buildReasonV2/buildHighlightsV2/buildRisksV2 增加逐岗位差异 |
| 调整 score cap/spread | 去除 top spread，增加 precision_bonus + risk_penalty |
| 调整 label 限制 | strong ≤ 2 个 |
| API 验证 | 5 类样例验收 |

**禁止**：
- 不改前端大布局
- 不改 DB
- 不改 init.sql

### K9-1-T-D：前端推荐页信息架构与卡片展示收口

| 范围 | 说明 |
|------|------|
| adapter 映射新增字段 | `recommendationAdapter.ts` 增加 why_ranked_here 等映射 |
| types 新增字段 | `agent.ts` RecommendedJobItem 增加 5 个新字段 |
| 左侧卡片展示 | 核心原因/证据/风险简短展示 |
| 中间详情区 | 完整解释展示 |
| 删除推荐记录 | 移除右侧"推荐记录"区域 |
| 合并推荐来源 | 合并到顶部状态栏 |
| Deep fallback 文案修正 | "深度推荐 · 规则兜底"替代"演示数据" |
| sourceLabel 规则修正 | 去掉"演示数据"文案 |

## 9. 文档输出清单

| 文档 | 路径 |
|------|------|
| 后端设计冻结 | `docs/04-后端设计/K9-1-T-B-岗位推荐逐岗位解释与分数区分度设计.md` |
| 前端设计冻结 | `docs/05-前端设计/K9-1-T-B-岗位推荐信息架构与右侧区域设计.md` |
| 开发日志 | `docs/09-开发日志/2026-06-07-566-K9-1-T-B-岗位推荐逐岗位解释分数区分度与右侧信息架构设计冻结.md` |
| 任务清单更新 | `docs/下一步任务.md` |
| 变更记录更新 | `docs/变更记录.md` |
