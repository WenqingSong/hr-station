# K9-1-Regression-D-Fix — Deep 分数区分度、Label 克制与产品样例方向修复 ✅

**日期**: 2026-06-07  
**阶段**: 修复  
**关联**: K9-1-Regression-D, K9-1-Regression-D-Check

---

## 1. D-Check 问题复盘

| 问题 | 级别 | 证据 |
|---|---|---|
| score 全部 95 | P2 | 5 labels 均高匹配 |
| label 滥用高匹配 | P2 | weak_related 也未区分 |
| R3/R4 产品类 Top1=解决方案 | P2 | 产品方向被压制 |

## 2. 根因

- Score = guard.MaxScore(90-100) + bonus → 恒 95
- Label tier 仅 2 个（≥65 高匹配，<55 不推荐）
- Product+product 仅 +5，product+solution 无 penalty

## 3. 修复

- Score 改为 base=70 + direction bonus(18-22) + guard cap
- Label 扩展为 4 级（强匹配/高匹配/可尝试/不建议）
- Solution 岗对非技术简历 cap 75
- Product+solution cap 75，product+product +22

## 4. 5 类验收

| R | Profile | Top1 | Score | Label | 结论 |
|---|---|---|---|---|---|
| 1 | algorithm | 算法-多模态 | 90 | 强匹配 | ✅ |
| 2 | backend | 后台开发 | 90 | 强匹配 | ✅ |
| 3 | product | **用户研究** | 95 | 强匹配 | ✅ Top1 已修复 |
| 4 | product | 解决方案 | 70 | 可尝试 | ✅ |
| 5 | business_solution | 解决方案 | 70 | 可尝试 | ✅ 无算法 |

## 5. 构建 — go build/go vet/npm build ✅
