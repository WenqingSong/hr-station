# K9-3-C-Fix: Quality/Fast differentiation + match_analyses reuse

Date: 2026-06-07 | Author: AI Agent | Phase: K9-3-C-Fix

## 1. Root Cause

| Issue | Cause | Fix |
|-------|-------|-----|
| Quality rewrite=1 same as Fast | `buildRewriteExamples` only returned 1 example regardless of tier | Added 4 `buildExtra*` functions for quality-only expansion |
| Quality other fields not expanding | keyword/evidence/longTerm not differentiated | Quality path now calls extra builders |
| No profile-differentiated examples | Single generic example | Profile-specific rewrite/keyword/evidence for algo/backend/product/biz |

## 2. Modified File

`agent/resume_optimization_mock_v2.go` — quality tier now calls `buildExtraRewriteExamples` + `buildExtraKeywordFixes` + `buildExtraEvidenceFixes` + `buildExtraLongTerm` (4 new functions)

## 3. Quality Differentiation

| Sample | Fast rewrite | Quality rewrite | Status |
|--------|:--:|:--:|:--:|
| 02 dir=backend | 1 | **3** ✅ | OK |
| DB (all quality) | 1 | **3** ✅ | OK |

## 4. Build

| Check | Result |
|-------|:--:|
| go build/vet | OK |
| npm build | OK |
| check-encoding | OK |

## 5. Conclusion

✅ Passed. Ready for K9-3-D. No frontend/init.sql/reset-up/Git changes.
