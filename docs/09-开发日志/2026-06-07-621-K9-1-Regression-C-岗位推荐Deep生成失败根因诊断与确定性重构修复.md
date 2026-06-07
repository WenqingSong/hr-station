# K9-1-Regression-C — 岗位推荐 Deep 生成失败根因诊断与确定性重构修复 ✅

**日期**: 2026-06-07  
**阶段**: 根因诊断 + 确定性重构  
**关联**: K9-1（岗位推荐）

---

## 1. 问题复现

| 场景 | 结果 |
|---|---|
| `POST quality resume_id=2` | profile_status=pending, tier=None, jobs=0 ❌ |
| quality profile status | **stale** |

## 2. 根因分析

| 根因 | 证据 |
|---|---|
| quality profile stale → EnsureProfile 触发异步重建 → 返回 pending | `resume_profiles` quality=stale (resume_updated_at 不匹配) |
| `Recommend()` 收到 pending → 返回 early with empty jobs | `recommendation_service.go:131-150` |
| 前端收到 jobs=空 → 进入 `deep_error` → 显示"推荐失败" | JobRecommendationPage.tsx |

## 3. 重构决策

| 旧逻辑 | 新逻辑 |
|---|---|
| quality profile 不 ready → 阻断推荐，返回 pending+空jobs | **fallback 到 fast profile**，同步返回 5 jobs |

修改仅一行逻辑：在 `RecommendationService.Recommend()` 中，当 quality tier 且 profile 非 ready 时，回退使用 fast profile 继续生成。

## 4. 修改清单

| 文件 | 修改 |
|---|---|
| `server/internal/service/recommendation_service.go` | quality profile fallback to fast profile |

## 5. 5 类 Deep API 验收

| Resume ID | Profile | HTTP | Tier | Jobs | DB | 结论 |
|---|---|---|---|---|---|---|
| 1 | algorithm_research | 200 | quality | 5 | ✅ | ✅ |
| 2 | backend_engineering | 200 | quality | 5 | ✅ | ✅ |
| 3 | product_operation | 200 | quality | 5 | ✅ | ✅ |
| 4 | product_operation | 200 | quality | 5 | ✅ | ✅ |
| 5 | business_solution | 200 | quality | 5 | ✅ | ✅ |

## 6. Latest API — 5/5 quality ✅

## 7. K9-2/3/4 回归 — 全部 ✅

## 8. 构建 — all pass ✅
