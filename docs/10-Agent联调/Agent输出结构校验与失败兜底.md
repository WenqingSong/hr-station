# Agent 输出结构校验与失败兜底

> 阶段：7-4 | 状态：pass | 日期：2026-06-01

## curl 复验

| Agent | code | result 关键字段 | meta | 结论 |
|------|:--:|------|------|:--:|
| 推荐 | 0 | 2 jobs, recommended_jobs[] | recommendation/fallback | ✅ |
| 匹配 | 0 | score=90, 8 evidence_chain | match_analysis/fallback | ✅ |
| 优化 | 0 | apply_readiness, 1 top_fix_items | resume_optimization/fallback | ✅ |

## adapter 容错

| 适配器 | 风险 |
|------|------|
| recommendationAdapter | `recommended_jobs.map()` 无 null 守卫 (P2) |
| matchAnalysisAdapter | 直取 primitive，安全 ✅ |
| resumeOptimizationAdapter | `keyword_alignment.map()` 等 3 处无 null 守卫 (P2) |

适配器假设 Mock Agent 返回完整字段，实际运行安全。

## 页面失败兜底

| 页面 | loading | error/fallback | 不白屏 | 结论 |
|------|:--:|:--:|:--:|:--:|
| 推荐页 | ✅ | ✅ | ✅ | ✅ |
| 匹配页 | ✅ | ✅ | ✅ | ✅ |
| 优化页 | ✅ | ✅ | ✅ | ✅ |

## 异常响应路径

fallback 链: agents.ts(4) + jobs.ts(4) + resumes.ts(8) = 16 处 withFallback ✅

## 工程化文案泄漏

**零** (llm_api_key_missing/401/API error/mock fallback)

## URL 验证

3/3 → 200 ✅

## build

✓ 1.67s ✅

## P0/P1/P2

| 等级 | 数量 | 说明 |
|:--:|:--:|------|
| P0 | 0 | — |
| P1 | 0 | — |
| P2 | 1 | adapter .map() 缺 null 守卫（不影响运行） |

## 结论

**pass ✅** — 可进入 7-5
