# 简历优化 Agent 联调

> 阶段：7-3 | 状态：pass | 日期：2026-06-01

## curl 联调

`POST /api/resumes/optimize` → 200, code:0, apply_readiness, top_fix_items, fallback: llm_api_key_missing ✅

请求体: `{resume_text, target_job: {title, jd_text}}` (target_job 为对象非字符串)

## 前端修改

- `agents.ts`: ResumeOptimizationRequest → `{resume_text, target_job, target_section, user_intention}`; 移除旧字段; 移除未使用的 Resume/JobDetail/mock imports
- `ResumeOptimizationPage.tsx`: mock → `loadResumeOptimizationViewModel({resume_text, target_job, ...})`; loading/fallback 状态

## build

✓ built in 1.81s ✅

## URL

`/resume-optimization?jobId=1&resumeId=resume-v2` → 200 ✅

## P0/P1/P2

全部 0

## 结论

**pass ✅** — 可进入 7-4
