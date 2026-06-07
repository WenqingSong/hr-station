# K9-1-Regression-G：岗位推荐 Deep 详情字段与信息密度修复

> 日期：2026-06-07
> 性质：后端实现，修改 `recommendation_mock.go`
> 结论：✅ **Deep 详情字段已填充，信息密度恢复**

---

## 一、问题复现

| 页面 | 现象 | 结论 |
|------|------|:--:|
| `/recommend?resumeId=1` Deep 详情区 | 排名第 1 位、综合匹配度 90、基于简历完整内容分析、建议查看岗位详情 | P0 空壳 |
| 全部 5 类 Deep 结果 | 6 个字段全是占位文案 | P0 信息密度为 0 |

---

## 二、根因分析

| 根因 | 证据 |
|------|------|
| `RecommendQualityWithFeatureCards` 构造 `RecommendedJob` 时字段为占位文案 | `recommendation_mock.go:1953-1962` |
| `WhyRankedHere`: `fmt.Sprintf("排在第 %d 位", i+1)` | 纯排名，无推荐理由 |
| `RecommendationReason`: `"基于完整特征分析，匹配度 %d 分"` | 空泛 |
| `MatchHighlights`: `[]string{"基于简历完整内容分析"}` | 占位 |
| `SuggestedAction`: `"建议查看岗位详情。"` | 唯一建议 |
| `ResumeEvidence`: 未设置 | 缺失 |
| `JobSpecificRisks`: 未设置 | 缺失 |
| `SuggestedImprovement`: 未设置 | 缺失 |

---

## 三、修改清单

| 文件 | 修改 |
|------|------|
| `server/internal/agent/recommendation_mock.go` | 新增 8 个 builder 函数 + 修改 `RecommendQualityWithFeatureCards` |

### 新增 Builder 函数

| Builder | 行数 | 说明 |
|---------|:---:|------|
| `buildWhyRankedHere` | ~30 | 基于方向匹配 + guard 生成 7 种场景的推荐理由 |
| `buildResumeEvidence` | ~55 | 基于 StrongCapabilities + 方向匹配生成 2-4 条证据 |
| `buildJobSpecificReason` | ~22 | 基于方向 + score 生成具体分析段落 |
| `buildJobSpecificRisks` | ~25 | guard reasons + 方向风险 |
| `buildSuggestedImprovement` | ~22 | 6 种场景的具体补强建议 |
| `buildMatchHighlights` | ~35 | 4 种方向的匹配亮点 + R4/R5 跨方向迁移 |
| `buildSuggestedAction` | ~15 | 基于 score 分层的建议 |
| `buildRecommendationReason` | ~15 | 基于方向特征的总结理由 |

### 字段填充对照

| 字段 | 修复前 | 修复后 |
|------|--------|--------|
| `why_ranked_here` | `"排在第 N 位"` | `"该岗位方向（XX）与简历中的算法建模…高度相关，因此进入 Top5"` |
| `resume_evidence` | 空 | 2-4 条具体证据，含方向匹配证据 |
| `job_specific_reason` | `"综合匹配度 NN"` | `"简历背景与 XX 岗位在算法建模…有明确对应关系，推荐分数 90 反映方向高度一致"` |
| `job_specific_risks` | 空 | 1-3 条 guard reasons + 方向风险 |
| `suggested_improvement` | 空 | `"建议在简历中补充模型评估指标（准确率/召回率/AUC）…"` |
| `match_highlights` | `["基于简历完整内容分析"]` | 2 条具体亮点 |
| `suggested_action` | `"建议查看岗位详情。"` | `"建议重点查看…使用匹配分析…提前准备针对性简历"` |
| `recommendation_reason` | `"基于完整特征分析，匹配度 %d"` | `"基于算法方向特征分析…明确信号，匹配度 %d 分"` |

---

## 四、5 类 API 验收

| R | Top1 | Score | Label | Evidence | Risks | Highlights | Why | 结论 |
|---|------|-------|-------|:---:|:---:|:---:|-----|:--:|
| 1 | 算法-多模态方向 | 90 | 强匹配，优先关注 | 3 | 1 | 2 | "与算法建模…高度相关" | ✅ |
| 2 | 后台开发方向 | 90 | 强匹配，优先关注 | 3 | 1 | 2 | "与后台开发…高度匹配" | ✅ |
| 3 | 用户研究 | 95 | 强匹配，优先关注 | 4 | 2 | 2 | "与需求分析…密切关联" | ✅ |
| 4 | 解决方案 | 70 | 可尝试，需要补强 | 2 | 1 | 2 | "通用能力部分匹配" | ✅ |
| 5 | 解决方案 | 70 | 可尝试，需要补强 | 2 | 1 | 2 | "通用能力部分匹配" | ✅ |

**全部 25 个 job 验收通过：evidence≥2, risks≥1, highlights=2, 无占位文案。**

---

## 五、DB 一致性验收

| resume_id | job_count | why_ranked (前40字) | evidence | risks | highlights | 结论 |
|-----------|-----------|------|:---:|:---:|:---:|:--:|
| 1 | 5 | "与算法建模、深度学习训练及实验评估经历高度相关" | 3 | 1 | 2 | ✅ |
| 2 | 5 | "与后台开发、数据库设计及接口开发能力高度匹配" | 3 | 1 | 2 | ✅ |
| 3 | 5 | "与需求分析、用户研究及项目推进经验密切关联" | 4 | 2 | 2 | ✅ |
| 4 | 5 | "与通用能力部分匹配，方向关联性尚可" | 2 | 1 | 2 | ✅ |
| 5 | 5 | "与通用能力部分匹配，方向关联性尚可" | 2 | 1 | 2 | ✅ |

---

## 六、测试结果

| 测试 | 结果 | 归因 |
|------|:--:|------|
| go build ./... | ✅ PASS | — |
| go vet ./... | ✅ PASS | — |
| npm build | ✅ PASS | — |
| go test ./... | ❌ 4 FAIL | K9-2 MatchAnalysis pre-existing |
| check-encoding | ❌ 1 FAIL | init.sql pre-existing |
| POST quality ×5 | ✅ PASS | 25/25 jobs 字段完整 |
| GET latest ×5 | ✅ PASS | 全部复用 quality |

---

## 七、旧文案检查

| 检查项 | 结果 |
|--------|:--:|
| "排名第 X 位" 不再作为推荐理由 | ✅ |
| "基于简历完整内容分析" 不再作为证据 | ✅ |
| "建议查看岗位详情。" 不再作为唯一建议 | ✅ |
| mock_quality 仅内部代码用，不展示 | ✅ |
| 旧失败文案不出现 | ✅ |

---

## 八、未修改范围确认

```text
✅ 未修改 K9-2 岗位匹配分析
✅ 未修改 K9-3 简历优化建议
✅ 未修改 K9-4 简历版本对比
✅ 未修改 server/sql/init.sql
✅ 未修改数据库结构
✅ 未删除真实简历
✅ 未执行 reset-up
✅ 未执行 Git 命令
```

---

## 九、构建清单

| 项目 | 状态 |
|------|:--:|
| go build ./... | ✅ |
| go vet ./... | ✅ |
| npm build | ✅ |
| go test ./... | ⚠️ 4 pre-existing (K9-2) |
| check-encoding | ⚠️ 1 pre-existing (init.sql) |

---

## 十、结论

```text
✅ K9-1-Regression-G 通过
Deep quality 详情字段全部填充完毕，信息密度已恢复。
25 个 job 全部非空壳，evidence/risks/highlights 达标。
可进入 K9-1-Regression-G-Check：岗位推荐 Deep 详情密度浏览器真实验收。
```
