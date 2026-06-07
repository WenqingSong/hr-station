# K9-2-E: Frontend match page latest-first + Fast/Deep display layering

Date: 2026-06-07 | Author: AI Agent | Phase: K9-2-E

## 1. Files Modified

| File | Change |
|------|--------|
| web/src/api/agents.ts | +getLatestMatchAnalysis + MatchLatestResponse type |
| web/src/api/viewModels.ts | +loadLatestMatchViewModel + LatestMatchViewModel |
| web/src/app/pages/JobMatchAnalysisPage.tsx | Rewrite: latest-first + phase state machine + Fast/Deep layering + AgentStepLoadingOverlay |

## 2. Page Entry Flow

| State | Action |
|-------|--------|
| loading_latest | GET /latest |
| latest=fast | fast_done, Fast simplified |
| latest=quality | deep_done, Deep complete (MatchAnalysisDetail) |
| no latest | POST fast, fast_done |

## 3. Fast vs Deep

| Mode | Left | Center | Right |
|------|------|--------|-------|
| Fast | Job sidebar | Score + level + 2 advantages + 1 gap + 1 suggestion + upsell | Resume + Deep button |
| Deep | Job sidebar | MatchAnalysisDetail (full) | Resume + Regenerate btn |

## 4. Header Labels

| Phase | Label |
|-------|-------|
| fast_done | 快速匹配 · 规则分析 |
| deep_done (mock_quality) | 深度匹配 · 规则兜底 |
| deep_done (llm) | 深度匹配 · AI生成 |
| deep_error | 深度匹配分析生成失败 · 已保留当前结果 |

## 5. AgentStepLoadingOverlay

4-step: profile → requirements → evidence → suggestions. Replaces AIThinkingLoading.

## 6. Build

| Check | Result |
|-------|:--:|
| npm build | OK |
| go build/vet | OK |
| check-encoding | OK |

## 7. Unchanged

No backend / init.sql / reset-up / Git changes.
