# K9-1-Regression-D — Deep 推荐最小可信重构 ✅

**日期**: 2026-06-07  
**阶段**: 确定性重构  
**关联**: K9-1（岗位推荐）

---

## 1. 问题复现

| 场景 | 结果 |
|---|---|
| R5 quality 重构前 | Top1=算法-计算机视觉方向, score=85, 高匹配 ❌ |
| R5 quality 重构后 | Top1=解决方案-行业咨询方向, score=95 ✅ |

---

## 2. 重构

### 新增能力
- **Runtime Resume Feature Card**: 基于完整简历文本构建（IsTechnical/IsAlgorithm/IsPublicService等）
- **Runtime Job Feature Card**: 岗位方向检测（IsAlgorithm/IsComputerVision/IsTechnical等）
- **Hard Guard**: 6 条规则限制非技术简历+技术岗组合
- **RecommendQualityWithFeatureCards**: Feature Card + Guard + Scoring 流水线

### Hard Guard 规则

| 规则 | 处理 |
|---|---|
| 非技术简历+技术岗 | score ≤ 70 |
| 无算法证据+算法岗 | score ≤ 60 |
| 无CV证据+CV岗 | score ≤ 55 |
| 公共事业简历+算法/CV | score ≤ 50 |
| 排斥互联网快节奏 | score ≤ 60 |

### 5 类验收

| R | Profile | Top1 重构前 | Top1 重构后 | 结论 |
|---|---|---|---|---|
| 1 | algorithm | 算法-多模态 | 算法-多模态 ✅ | ✅ |
| 2 | backend | 后台开发 | 后台开发 ✅ | ✅ |
| 3 | product | 技术研究 | 解决方案 ✅ | ✅ |
| 4 | product | 技术研究 | 解决方案 ✅ | ✅ |
| 5 | business_solution | **算法-CV ❌** | **解决方案 ✅** | ✅ |

## 3. 构建 — go build / go vet / npm build ✅
