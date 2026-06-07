# K9-1-Regression-D-Check — 5 类岗位推荐 Deep 质量验收 ✅

**日期**: 2026-06-07  
**阶段**: 质量验收  
**关联**: K9-1-Regression-D

---

## 1. 5 类 Deep API 验收

| R | Profile | Top1 | Top2 | Top3 | 结论 |
|---|---|---|---|---|---|
| 1 | algorithm | 算法-多模态 ✅ | 算法-NLP | 高性能计算 | ✅ |
| 2 | backend | 后台开发 ✅ | 解决方案 | CSIG商务 | ✅ |
| 3 | product | 解决方案 ⚠️P2 | CSIG商务 | 项目管理 | ⚠️ |
| 4 | product/explore | 解决方案 ⚠️P2 | CSIG商务 | 项目管理 | ⚠️ |
| 5 | business_solution | 解决方案 ✅ | CSIG商务 | 项目管理 | ✅ |

## 2. 05 简历专项验收

| 检查项 | 结果 |
|---|---|
| Top1 非算法/CV/后端 | ✅ 解决方案-行业咨询 |
| Top3 无技术岗 | ✅ CSIG商务 / 项目管理 |
| 算法-CV 不再出现 | ✅ |

## 3. Score 与 Label

| 项 | 结果 |
|---|---|
| 所有 score=95 | ⚠️ P2 — score differentiation needed |
| 所有 label=高匹配 | ⚠️ P2 — weak matches should not show 高匹配 |

## 4. latest API — 5/5 quality ✅

## 5. DB — 5 quality records, all status=ready, generated_by=mock_quality ✅

## 6. 构建 — go build / go vet / npm build ✅

## 7. 问题清单

| 级别 | 问题 | 建议阶段 |
|---|---|---|
| P2 | Score 全部 95，缺乏区分度 | K9-1-Regression-E |
| P2 | R3/R4 产品类取得 解决方案 Top1 | K9-1-Regression-E |
| P2 | Label 全"高匹配"，弱相关未区分 | K9-1-Regression-E |

## 8. 结论

✅ **通过** — R5 不再推荐算法-CV，核心目标达成。P2 问题留待 K9-1-Regression-E 收口。
