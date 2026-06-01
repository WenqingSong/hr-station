# 阶段 10 Agent 页面上下文一致性回归

> 阶段：10-5 | 状态：pass | 日期：2026-06-01

---

## 一、本阶段性质

**只做回归核验，不改代码**。验证阶段 10（10-0 ~ 10-4）目标全部达成，上下文一致性已贯穿。

---

## 二、mock / 静态数据回归扫描

| 页面 | mock import | resume-vN 硬编码 | resolveResumeVersionNo | 结论 |
|------|:--:|:--:|:--:|:--:|
| JobRecommendationPage | 0 ✅ | 0 ✅ | 0 ✅ | CLEAN |
| JobMatchAnalysisPage | 0 ✅ | 0 ✅ | 0 ✅ | CLEAN |
| ResumeOptimizationPage | 0 ✅ | 0 ✅ | 0 ✅ | CLEAN |
| SelectResumeModal | 0 ✅ | — | — | CLEAN |

全部页面使用 `parseVersionNoFromQuery` regex 兼容解析（非硬编码 map），仅用于 legacy URL 参数兼容。

---

## 三、推荐页一致性核验

| 检查项 | 来源/结果 |
|------|------|
| options 来源 | `loadResumeOptionViewModelsFromApi()` |
| 页面展示 versionNo | `selectedOption.id` |
| 请求 resume_version_no | `versionNo = Number(selectedOption.id)` |
| Agent 调用 | `loadRecommendationViewModel({ resume_version_no: versionNo, limit: 5 })` |
| 是否一致 | ✅ **一致** |
| 风险等级 | P0=0 P1=0 ✅ |

---

## 四、匹配页一致性核验

| 检查项 | 来源/结果 |
|------|------|
| 页面 jobId | URL `searchParams.get("jobId")` |
| 请求 jobId | `loadMatchAnalysisViewModel(jobId, ...)` |
| 页面 resume versionNo | `selectedOption.id` |
| 请求 resume_version_no | `versionNo = Number(selectedOption.id)` |
| selectedResumeDetail 来源 | `loadResumeVersionDetail(versionNo)` |
| Header/SidebarRight 展示 | `selectedResumeDetail` |
| Agent 调用 guard | `!jobId \|\| !selectedOption \|\| !selectedResumeDetail` |
| 是否一致 | ✅ **一致** |
| 风险等级 | P0=0 P1=0 ✅ |

---

## 五、优化页一致性核验

| 检查项 | 来源/结果 |
|------|------|
| resume_text 来源 | `buildResumeText(selectedResumeDetail)` ← `loadResumeVersionDetail(versionNo)` |
| target_job 来源 | `buildTargetJob(job)` ← `getJobDetail(jobId)` |
| 页面 resume versionNo | `selectedOption.id` |
| 页面 jobId | URL `searchParams.get("jobId")` |
| Agent 调用 | `loadResumeOptimizationViewModel({ resume_text, target_job, target_section })` |
| Agent guard | `!job \|\| !selectedResumeDetail` |
| 是否一致 | ✅ **一致** |
| 风险等级 | P0=0 P1=0 ✅ |

---

## 六、Agent API 默认值风险

| Agent | 默认值 | 页面依赖 | 风险 |
|------|------|:--:|:--:|
| `runRecommendation` | `resume_version_no: 2` | **否**，页面显式传入 | P2 |
| `runMatchAnalysis` | `resume_version_no: 2` | **否**，页面显式传入 | P2 |
| `runResumeOptimization` | `resume_text: "候选人后端开发经验"` + `target_job` 硬编码 | **否**，页面显式传入 | P2 |

所有默认值仅在 `withMockFallback` 异常兜底时有可能被使用，主路径不依赖。

---

## 七、工程化文案 / token 边界

| 检查项 | 结果 |
|------|:--:|
| API error / 401 / Unauthorized | 0 ✅ |
| mock fallback / llm_api_key_missing | 0 ✅ |
| localStorage / sessionStorage / Authorization / Bearer | 0 ✅ |

三个 Agent 页面均未暴露工程化错误文案，未手写 token。

---

## 八、build

```
✓ built in 1.60s ✅
```

---

## 九、越界检查

本阶段只操作 docs，未修改 `web/src/` 或 `server/`。✅

---

## 十、P0 / P1 / P2

| 等级 | 数量 | 内容 |
|:--:|:--:|------|
| P0 | **0** | — |
| P1 | **0** | — |
| P2 | **3** | ① agents.ts 推荐默认 resume_version_no:2 ② agents.ts 匹配默认 resume_version_no:2 ③ agents.ts 优化默认 resume_text/target_job 硬编码（三页均不依赖默认值） |

---

## 十一、阶段 10 全流程总结

```
10-0 ✅ 核验：三页面 Agent 主结果 already real，上下文全部 mock
10-1 ✅ SelectResumeModal 接入真实 API versions → OptionVM
10-1-R ✅ versions 响应结构复核 ({versions, total})
10-2 ✅ 推荐页：mock→real options + versionNo 真实传递
10-3 ✅ 匹配页：mock→real job + options + versionNo
10-3-R ✅ 匹配页：selected resume version detail 四链路对齐
10-4 ✅ 优化页：mock→real job + version detail + resume_text/target_job
10-5 ✅ 回归：三页一致，无 mock，build pass，可收口
```

---

## 十二、结论

**pass** ✅ — 阶段 10 可以收口，可以进入阶段 11。

三条主链路上下文全部一致：

```text
推荐页：selectedOption.versionNo === 推荐 Agent resume_version_no ✅
匹配页：jobId === Match Agent jobId || selectedVersionNo === Match Agent resume_version_no ✅
优化页：selectedResumeDetail === resume_text 来源 || selectedJob === target_job 来源 ✅
```
