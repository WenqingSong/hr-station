# Agent 页面上下文真实化现状核验

> 阶段：10-0 | 状态：pass | 日期：2026-06-01

---

## 一、本阶段目标

查清 Agent 相关页面还有哪些上下文仍是静态/mock 数据，为阶段 10 后续修复提供准确输入。

---

## 二、核验范围

| 维度 | 方式 |
|------|:--:|
| Agent API 请求结构 | grep + read |
| Agent 页面 mock import | grep |
| 推荐页上下文 | grep + read |
| 匹配页上下文 | grep + read |
| 优化页上下文 | grep + read |
| ResumeComparisonPage | grep |
| SelectResumeModal | grep |
| 后端 Agent curl | curl |
| build | npm run build |

---

## 三、Agent API 请求结构核验

| Agent | API 函数 | 请求字段 | 默认值 | 是否硬编码 | 风险 |
|------|------|------|------|:--:|:--:|
| 推荐 | `runRecommendation` | `resume_version_no` | **2** | ✅ 硬编码 | **P1** |
| 推荐 | `runRecommendation` | `limit` | 5 | ✅ 合理 | — |
| 匹配 | `runMatchAnalysis` | `resume_version_no` | **2** | ✅ 硬编码 | **P1** |
| 优化 | `runResumeOptimization` | `resume_text` | **"候选人后端开发经验"** | ✅ 硬编码 | **P1** |
| 优化 | `runResumeOptimization` | `target_job` | **{title:"后端开发工程师", jd_text:"要求Go/MySQL/Redis等后端能力"}** | ✅ 硬编码 | **P1** |

### 关键发现

```
推荐/匹配的 resume_version_no 默认值 2 是硬编码；
优化页的 resume_text 和 target_job 默认值是硬编码 mock 文案。
虽然页面调用时可能传入真实值，但默认值本质仍是 mock。
```

---

## 四、Agent 页面 mock import 扫描

| 文件 | mock 导入 | 用途 | 用户可见 | 风险 |
|------|------|------|:--:|:--:|
| JobRecommendationPage | `mockCurrentResume`, `mockResumeVersions` | 上下文展示（简历名称/版本） | **是** | **P1** |
| JobMatchAnalysisPage | `mockSelectedJob`, `mockJobDetails`, `mockCurrentResume`, `mockResumeVersions` | 上下文展示（目标岗位+当前简历） | **是** | **P1** |
| ResumeOptimizationPage | `mockSelectedJob`, `mockJobDetails`, `mockCurrentResume`, `mockResumeVersions` | 上下文展示（目标岗位+当前简历） | **是** | **P1** |
| ResumeComparisonPage | 0 mock import ✅ | — | — | ✅ |
| SelectResumeModal | `loadResumeOptionViewModels` → `getResumeOptions` → **mockResumeOptions fallback** | 简历选择列表 | **是** | P2 |

### 额外发现

```
resumeId → version_no 映射硬编码：
  JobRecommendationPage: resolveResumeVersionNo() (line 12)
  JobMatchAnalysisPage: resolveResumeVersionNo() (line 14)

这两个函数通过解析 "resume-v2" → 2 来映射，属于阶段 5 的兼容方案。
```

---

## 五、推荐页上下文核验

| 检查项 | 当前来源 | 是否真实 | 风险 |
|------|------|:--:|:--:|
| 推荐结果 | `loadRecommendationViewModel` → Agent API | ✅ 真实 | — |
| 当前简历名称 | `mockResumeVersions.find(id)` / `mockCurrentResume` | ❌ mock | **P1** |
| 当前简历版本 | `resolveResumeVersionNo(resumeId)` → v2 | ❌ 硬编码 | **P1** |
| resume_version_no | `resolveResumeVersionNo(resumeId)` | ❌ 硬编码映射 | **P1** |
| resumeId URL | `searchParams.get("resumeId")` → "resume-v2" | ⚠️ 历史格式 | P2 |

**结论：推荐页主结果已真实化，但页面上下文（简历名称、版本信息）仍来自 mock，与 Agent 请求使用的数据可能不一致。**

---

## 六、匹配页上下文核验

| 检查项 | 当前来源 | 是否真实 | 风险 |
|------|------|:--:|:--:|
| 匹配结果 | `loadMatchAnalysisViewModel` → Agent API | ✅ 真实 | — |
| 目标岗位展示 | `mockJobDetails[jobId]` / `mockSelectedJob` | ❌ mock | **P1** |
| 当前简历展示 | `mockResumeVersions.find(id)` / `mockCurrentResume` | ❌ mock | **P1** |
| jobId | URL params → 真实 job id | ✅ 真实 | — |
| resume_version_no | `resolveResumeVersionNo(resumeId)` | ❌ 硬编码映射 | **P1** |

**结论：匹配页主结果已真实化，但目标岗位和当前简历的上下文展示仍来自 mock，与 Agent 请求使用的数据可能不一致。**

---

## 七、优化页上下文核验

| 检查项 | 当前来源 | 是否真实 | 风险 |
|------|------|:--:|:--:|
| 优化结果 | `loadResumeOptimizationViewModel` → Agent API | ✅ 真实 | — |
| resume_text | 页面组件组装 → 传给 viewModel | ⚠️ 部分 mock | **P1** |
| target_job | 页面组件组装 → 传给 viewModel | ⚠️ 部分 mock | **P1** |
| 当前简历展示 | `mockResumeVersions.find(id)` / `mockCurrentResume` | ❌ mock | **P1** |
| 目标岗位展示 | `mockJobDetails[jobId]` / `mockSelectedJob` | ❌ mock | **P1** |

**结论：优化页主结果已真实化，但 resume_text/target_job 组装来源仍包含 mock，页面上下文展示全来自 mock。**

---

## 八、ResumeComparisonPage 核验

| 数据块 | 当前来源 | 是否真实 | 风险 |
|------|------|:--:|:--:|
| 版本名称 | `loadResumeComparisonViewModel("1","2")` → API | ✅ | ✅ |
| 对比主结果 | API 返回 | ✅ | ✅ |
| 新增/强化/风险模板 | 部分硬编码字符串 | ⚠️ | P2 |

**结论：比较页已接入真实 API，但部分展示模板仍有硬编码。不属于 Agent 上下文核心问题，列为 P2。**

---

## 九、SelectResumeModal / Resume Options 核验

| 检查项 | 当前来源 | 是否真实 | 风险 |
|------|------|:--:|:--:|
| 简历选项列表 | `loadResumeOptionViewModels` → `getResumeOptions` → fallback `mockResumeOptions` | ⚠️ fallback→mock | P2 |
| 选项数据结构 | `ResumeOption[]`（与后端 `VersionListResponse` 类型不匹配） | ❌ | P2 |

**结论：有 token 时可调真实 API，类型不匹配是隐藏风险。留到 10-1 处理。**

---

## 十、后端 Agent curl 回归

| Agent | HTTP | code | agent_status | 结论 |
|------|:--:|:--:|------|:--:|
| recommend | 200 | 0 | fallback | ✅ |
| match | 200 | 0 | fallback | ✅ |
| optimize | 200 | 0 | fallback | ✅ |

---

## 十一、build 结果

```
✓ built in 1.60s ✅
```

---

## 十二、P0/P1/P2

| 等级 | 数量 | 内容 |
|:--:|:--:|------|
| P0 | **0** | — |
| P1 | **3** | ① 三页面上下文展示仍用 mock resume/job ② agents.ts 默认值硬编码 ③ resumeId→version_no 硬编码映射 |
| P2 | **3** | ① SelectResumeModal 类型不匹配 ② ResumeComparisonPage 模板硬编码 ③ URL 参数格式历史遗留 |

---

## 十三、阶段 10 后续拆分建议

| 阶段 | 目标 | 是否改代码 | 依赖 | 验收标准 |
|:--:|------|:--:|------|------|
| 10-1 | SelectResumeModal / Resume Options 真实化 | 是 | 10-0 | 简历选项来自真实 API，类型匹配 |
| 10-2 | 推荐页上下文真实化 | 是 | 10-1 | 上下文展示与 Agent 请求数据一致 |
| 10-3 | 匹配页上下文真实化 | 是 | 10-1 | 目标岗位+当前简历上下文真实 |
| 10-3-R | 匹配页 selected resume version detail 对齐 | 是 | 10-3 | 四链路 versionNo 一致，10-3 partial_pass→pass |
| 10-4 | 简历优化页上下文真实化 | 是 | 10-1 | resume_text/target_job 从真实数据组装 |
| 10-5 | Agent 页面上下文一致性回归 | 否 | 10-4 | 全链路：请求=展示 |

---

## 十四、未处理事项

```
✅ 只核验
✅ 未改代码
✅ 未改后端
✅ 未调真实 LLM
```

---

## 十五、最终结论

**pass ✅** — Agent 页面上下文现状已全部摸清，主结果已真实化但上下文全为 mock。可进入 10-1。
