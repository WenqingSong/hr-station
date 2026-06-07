# K9-3-B: Resume Optimization Frontend Design

## 1. Page State Machine

```
loading_latest → GET /latest
  → has_result=true + tier=fast → fast_done
  → has_result=true + tier=quality → deep_done
  → has_result=false → fast_loading → POST fast → fast_done

Click Deep:
  deep_loading → POST quality → deep_done
```

Phases: loading_latest, fast_loading, fast_done, fast_error, deep_loading, deep_done, deep_error

## 2. Entry URLs

| URL | Mode |
|-----|------|
| /resume-optimization?resumeId=X | general |
| /resume-optimization?resumeId=X&targetDirection=backend_engineering | target_direction |
| /resume-optimization?resumeId=X&jobId=Y | target_job |

## 3. Fast Display (Simplified)

**Left/Summary**: optimization_level + score, one_sentence_summary, 3-5 quick_fix_items, 1-2 risk_points, delivery_focus

**Right**: Resume info + "Generate Deep Optimization" button

Fast does NOT show: structure/content/keyword/evidence fix items, rewrite examples, long-term items

## 4. Deep Display (Complete)

Sections: Optimization conclusion → Structure fixes → Content fixes → Keyword fixes → Evidence fixes → Risk points → **Rewrite examples (visual focus)** → Delivery focus → Long-term improvements

Right: Resume info + "Regenerate Deep Optimization" button

## 5. Header Labels

| Phase | Label |
|-------|-------|
| fast_done | 快速优化 · 规则分析 |
| deep_done (mock) | 深度优化 · 规则兜底 |
| deep_done (llm) | 深度优化 · AI生成 |
| deep_error | 深度优化生成失败 · 已保留当前结果 |

No "mock_agent"/"演示数据"/"推荐失败" in display

## 6. AgentStepLoadingOverlay

4 steps: 读取简历画像 → 分析目标方向 → 定位优化缺口 → 生成改写建议

Replaces AIThinkingLoading

## 7. Rewrite Examples Visual Design

Each rewrite example card: section label, "Before" (strikethrough/amber), "After" (green highlight), "Reason" (italic/small)

This is the core visual differentiation from Fast mode.
