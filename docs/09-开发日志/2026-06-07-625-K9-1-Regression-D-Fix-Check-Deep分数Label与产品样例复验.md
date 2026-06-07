# K9-1-Regression-D-Fix-Check — Deep 分数 Label 与产品样例复验 ✅

**日期**: 2026-06-07  
**阶段**: 复验  
**关联**: K9-1-Regression-D-Fix

---

## 1. 5 类 Deep API 复验

| R | Profile | Top1 | Score | Label | Top3 方向 | 结论 |
|---|---|---|---|---|---|---|
| 1 | algorithm | 算法-多模态 | 90 | 强匹配 | 算法/NLP/高性能 | ✅ |
| 2 | backend | 后台开发 | 90 | 强匹配 | 后台/测试/客户端 | ✅ |
| 3 | product | **用户研究** | 95 | 强匹配 | 用户研究/项目管理/运营 | ✅ |
| 4 | product | 解决方案 | 70 | 可尝试 | 方案/商务/项目 | ✅ |
| 5 | business | 解决方案 | 70 | 可尝试 | 方案/商务/项目 | ✅ |

## 2. Score/Label 复验

| 项 | 结果 |
|---|---|
| 分数范围 | 70, 90, 92, 95 (3+ distinct) ✅ |
| Label 种类 | 强匹配 / 可尝试 (2+) ✅ |
| R3 Top1 非解决方案 | ✅ 用户研究 |
| R5 无算法/CV | ✅ |

## 3. 构建 — go build/go vet/npm build ✅

## 4. 结论

✅ **通过 → K9-1-Regression-E**
