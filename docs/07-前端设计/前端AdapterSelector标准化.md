# 前端 Adapter / Selector 标准化

> 阶段：5-1.3 | 状态：已完成 | 日期：2026-05-31

## 1. 本阶段目标

基于 5-1.1 types + 5-1.2 mocks，新增统一的 Adapter / Selector 层，将原始数据（Mock 或未来后端响应）转换为页面友好 ViewModel，为后续 5-1.4 页面改造提供稳定数据转换函数。

## 2. 新增 Adapter 文件

| 文件 | 行数 | 说明 |
|------|:--:|------|
| `web/src/adapters/common.ts` | 100 | 通用工具（statusBadge、formatListText、formatAgentSourceLabel、priority/evidence tone 映射） |
| `web/src/adapters/agentAdapter.ts` | 98 | Agent 公共字段（meta → AgentMetaViewModel、TopFix → TopFixItemViewModel、Evidence → EvidenceItemViewModel） |
| `web/src/adapters/jobAdapter.ts` | 77 | 岗位（Job → JobCardViewModel、JobDetail → JobDetailViewModel） |
| `web/src/adapters/resumeAdapter.ts` | 118 | 简历（Resume → ResumeSummaryViewModel、ResumeOption → ResumeOptionViewModel、ResumeVersionComparison → ResumeComparisonViewModel） |
| `web/src/adapters/recommendationAdapter.ts` | 70 | 岗位推荐（RecommendationResult + meta → RecommendationPageViewModel） |
| `web/src/adapters/matchAnalysisAdapter.ts` | 48 | 匹配分析（MatchAnalysisResult + meta → MatchAnalysisPageViewModel） |
| `web/src/adapters/resumeOptimizationAdapter.ts` | 100 | 简历优化（ResumeOptimizationResult + meta → ResumeOptimizationPageViewModel） |
| `web/src/adapters/index.ts` | 7 | 统一导出 |

**合计：8 个文件，618 行**

## 3. ViewModel 设计原则

- **页面不消费原始数据**：所有页面通过 ViewModel 消费数据
- **tone/label 预计算**：优先级、证据强度、Agent 状态在 adapter 中完成 tone 和 label 映射
- **纯函数**：所有 adapter 函数为纯函数，输入 → 输出，无副作用
- **类型安全**：所有输入/输出使用 5-1.1 定义的类型，不使用 `any`
- **可组合**：高级 adapter（如 recommendationAdapter）组合基础 adapter（agentAdapter）
- **兼容未来后端**：接收 `AgentResponseMeta` 可选参数，与 `AgentResponseEnvelope` 对齐

## 4. Agent meta 展示映射

| 条件 | 标签 | tone |
|------|------|:--:|
| `llm_agent + success` | LLM 生成 | blue |
| `llm_agent + repaired` | 已自动修复 | green |
| `agent_status=fallback` | 兜底结果 | orange |
| `agent_status=mock_success` | Mock 演示结果 | gray |
| `agent_status=error` | 异常但可用 | red |
| 默认 | 状态未知 | gray |

### 优先级 tone 映射

| 优先级 | tone |
|--------|:--:|
| high | orange |
| medium | blue |
| low | gray |

### 证据强度 tone 映射

| 强度 | tone | 中文 |
|------|:--:|------|
| strong | green | 强 |
| medium | blue | 中 |
| weak | orange | 弱 |
| none | red | 无 |

## 5. 页面与 Adapter 对应关系

| 页面 | Adapter | 输入 | 输出 |
|------|------|------|------|
| 岗位列表页 | `jobAdapter` | `Job[]` | `JobCardViewModel[]` |
| 岗位详情页 | `jobAdapter` | `JobDetail` | `JobDetailViewModel` |
| 岗位推荐页 | `recommendationAdapter` + `agentAdapter` | `RecommendationResult + meta` | `RecommendationPageViewModel` |
| 岗位匹配分析页 | `matchAnalysisAdapter` + `agentAdapter` | `MatchAnalysisResult + meta` | `MatchAnalysisPageViewModel` |
| 简历优化页 | `resumeOptimizationAdapter` + `agentAdapter` | `ResumeOptimizationResult + meta` | `ResumeOptimizationPageViewModel` |
| 我的简历管理首页 | `resumeAdapter` | `Resume` | `ResumeSummaryViewModel` |
| 简历编辑/添加页 | `resumeAdapter` | `Resume` | `ResumeSummaryViewModel` |
| 选择简历弹窗 | `resumeAdapter` | `ResumeOption[]` | `ResumeOptionViewModel[]` |
| 历史简历比较页 | `resumeAdapter` | `ResumeVersionComparison` | `ResumeComparisonViewModel` |

## 6. 明确本阶段不做

- ❌ 不修改页面组件
- ❌ 不创建 API Client（无 fetch/axios）
- ❌ 不接后端接口
- ❌ 不替换页面硬编码数据（留到 5-1.4）
- ❌ 不修改 types 或 mocks
- ❌ 不引入状态管理库
- ❌ 不引入 React

## 7. 后续 5-1.4 输入

5-1.4 可将页面内硬编码数据替换为：

```ts
// 旧方式（页面内硬编码）
const JOB_MOCKS = [{ id: 'job-1', ... }];

// 新方式（5-1.4 后将做）
import { mockJobs } from '../mocks';
import { toJobCardViewModels } from '../adapters';
const jobCards = toJobCardViewModels(mockJobs);
```
