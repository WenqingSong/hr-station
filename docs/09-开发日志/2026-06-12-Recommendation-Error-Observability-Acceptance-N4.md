# 2026-06-12 Recommendation-Error-Observability-Acceptance-N4 — 报错可观测性验收

## 最终状态：PASS

### ErrorBoundary
- groupCollapsed + console.error + console.info componentStack ✅
- 用户文案 "页面渲染异常" ✅

### API 错误日志
- Fast catch: groupCollapsed + error + context{action, resumeId, tier} ✅
- Deep catch: groupCollapsed + classifyDeepError + context ✅

### API 错误场景
- unauth: 40100 "unauthorized" ✅
- invalid resume_id: 42001 "resume version not found" ✅
- invalid tier: 40001 "invalid generation_tier" ✅
- latest invalid: 42001 ✅

### LLM Gate
- 4 层：llmEnabled → qualityClient → cascade → classifyDeepError ✅

### 敏感信息
- console 日志仅 error object + context，无 token/key/password ✅

### 构建
- npm build + go build/vet 全部通过 ✅
