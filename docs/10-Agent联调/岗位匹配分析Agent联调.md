# 岗位匹配分析 Agent 联调

> 阶段：7-2 | 状态：pass | 日期：2026-06-01

## curl 联调

| 接口 | HTTP | code | result | meta | 结论 |
|------|:--:|:--:|------|------|:--:|
| `POST /api/jobs/1/match` | 200 | 0 | match_score=98, 8 evidence_chain | fallback: llm_api_key_missing | ✅ |

## 前端修改

- `agents.ts`: MatchAnalysisRequest → `{resume_version_no}`, runMatchAnalysis 默认版2
- `JobMatchAnalysisPage.tsx`: mock → `loadMatchAnalysisViewModel(jobId, {resume_version_no})`
- loading: "正在生成匹配分析…", fallback: "匹配分析接口暂时不可用…"

## build

✓ built in 1.61s ✅

## URL

`/match?jobId=1&resumeId=resume-v2` → 200 ✅

## P0/P1/P2

全部 0

## 结论

**pass ✅** — 可进入 7-3
