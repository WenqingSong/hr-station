# K6-0：四主 Agent 架构冻结与现状核验

## 1. 性质

read-only audit + docs。**未修改代码、Prompt、Schema。**

## 2. 四个核心功能标准命名

| # | 核心功能 | 主 Agent |
|---|---------|---------|
| 1 | 简历版本对比 | Resume Version Compare Agent |
| 2 | 岗位智能推荐 | Job Recommendation Agent |
| 3 | 岗位匹配分析 | Job Match Analysis Agent |
| 4 | 简历优化建议 | Resume Optimization Agent |

## 3. 四主 Agent 架构

```
Offer Catcher 四主 Agent 架构

1. Resume Version Compare Agent
   功能：简历版本对比
   目标：比较不同简历版本的变化、提升点、退化点、仍需补强点，并判断是否更贴近目标岗位。

2. Job Recommendation Agent
   功能：岗位智能推荐
   目标：基于当前简历，推荐更适合投递的岗位，并给出推荐理由、匹配标签和风险提示。

3. Job Match Analysis Agent
   功能：岗位匹配分析
   目标：分析当前简历与当前岗位 JD 的匹配度，输出匹配分、优势、风险、证据链和投递建议。

4. Resume Optimization Agent
   功能：简历优化建议
   目标：针对当前岗位，给出可执行的简历修改建议、JD 证据、关键词补强和改写方向。
```

### 子能力 Agent（增强能力，非独立主功能）

- **Evidence Chain Agent**：证据链增强，服务匹配分析和简历优化
- **Rewrite Agent**：简历表达改写增强，服务简历优化
- **Optimization Core Agent**：简历优化首屏核心建议，服务简历优化

**原则：** 四个主 Agent 是产品主功能；子 Agent 是增强能力；后续所有开发必须先映射到四个核心功能之一。

## 4. 四主功能现状核验矩阵

| 核心功能 | 页面 | 接口 | Agent | Mock fallback | 真实 LLM | Prompt | Schema | 真实数据 | 可浏览器验收 |
|---------|------|------|-------|--------------|---------|--------|--------|---------|------------|
| 简历版本对比 | ✅ ResumeComparisonPage | ✅ GET /resumes/versions/compare | ⚠️ 仅 Mock | ✅ | ❌ **无 LLM Agent** | ❌ **无 Prompt** | ❌ **无 Schema** | ✅ | ✅ |
| 岗位智能推荐 | ✅ JobRecommendationPage | ✅ POST /jobs/recommend | ✅ RecommendationAgent | ✅ | ✅ repaired | ✅ recommendation_v1 | ✅ recommendation_schema | ✅ | ✅ |
| 岗位匹配分析 | ✅ JobMatchAnalysisPage | ✅ POST /jobs/{id}/match | ✅ MatchAnalysisAgent | ✅ | ✅ success | ✅ match_analysis_v1 | ✅ match_analysis_schema | ✅ | ✅ |
| 简历优化建议 | ✅ ResumeOptimizationPage | ✅ POST /resumes/optimize | ✅ ResumeOptimizationAgent | ✅ | ⚠️ fallback | ✅ resume_optimization_v1 | ✅ resume_optimization_schema | ✅ | ✅ |

## 5. 简历版本对比专项核验

### 结论：**B. 页面真实化但未 Agent 化**

| 维度 | 结果 |
|------|------|
| 前端页面 | ✅ `ResumeComparisonPage.tsx` — 使用 `loadResumeComparisonViewModel`，数据来自真实 API |
| 后端接口 | ✅ `GET /api/resumes/versions/compare` |
| Mock Agent | ✅ `VersionCompareMockAgent` — 规则打分 |
| LLM Agent | ❌ **不存在** — 无 `version_compare_agent.go` |
| Pipeline 集成 | ❌ Service 无 `pipelineRunner` / `llmEnabled` |
| Prompt | ❌ 无 `version_compare_v1.md` |
| Schema | ❌ 无 `version_compare_schema.json` |
| Server main.go | ❌ 无 LLM Pipeline 初始化代码（仅 Match/Rec/Opt 有） |

**对比其他三个 Agent：**

| | Version Compare | Match | Rec | Opt |
|---|:---:|:---:|:---:|:---:|
| Mock | ✅ | ✅ | ✅ | ✅ |
| LLM Agent | ❌ | ✅ | ✅ | ✅ |
| Pipeline | ❌ | ✅ | ✅ | ✅ |
| Prompt | ❌ | ✅ | ✅ | ✅ |
| Schema | ❌ | ✅ | ✅ | ✅ |

### 优先级

**P1** — 四主 Agent 中唯一未完成 LLM 化的功能。建议作为 K6-1 首任务。

## 6. K5 模型状态与 K6 架构状态边界

```
K5 已完成：
- Match Analysis：真实 LLM success（hy3-preview）
- Recommendation：真实 LLM repaired（hy3-preview）
- Resume Optimization：仍 fallback（hy3-preview JSON 不稳定 / deepseek 推理超时）

K6 不继续调模型：
- K6 先冻结产品和 Agent 架构
- 模型选择后置到 K6-5
- 不把模型问题混为 Agent 架构问题
```

## 7. P0 / P1 / P2

| 级别 | 编号 | 问题 |
|------|------|------|
| — | — | P0 = 0 |
| P1 | P1-1 | 简历版本对比未 Agent 化（无 LLM/Prompt/Schema/Pipeline） |
| P2 | P2-1 | Resume Optimization 真实 LLM 仍不稳定 |
| P2 | P2-2 | 浏览器最终演示链路未验收 |

## 8. K6 后续计划

```
K6-1：简历版本对比 LLM Agent 化（Prompt + Schema + Pipeline 集成）
K6-2：简历优化 Core / Evidence / Rewrite 子 Agent 拆分设计
K6-3：Resume Optimization Core Agent 最小实现
K6-4：四主功能页面联调验收
K6-5：hy3-preview vs DeepSeek V4 Pro 最终模型对比
K6-6：最终演示链路准入验收
```

## 9. 判定

**K6-0 = pass。** 四主 Agent 架构已冻结，现状核验完成，差距（版本对比未 Agent 化）已明确。P0=0 P1=1。

## 10. 是否可以进入 K6-1

**可以。** K6-1：简历版本对比 LLM Agent 化。
