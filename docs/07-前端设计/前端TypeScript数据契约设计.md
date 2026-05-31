# 前端 TypeScript 数据契约设计

> 阶段：5-1.1 | 状态：已完成 | 日期：2026-05-31

## 1. 本阶段目标

建立统一的前端 TypeScript 数据契约，为后续 5-1.2 Mock 标准化、5-1.3 Adapter、5-1.4 页面改造、5-2 API Client 打基础。

本阶段只新增 TypeScript 类型定义 + `tsconfig.json`，不写 mocks / adapters / API client / 页面改造。

## 2. 新增文件

| 文件 | 行数 | 说明 |
|------|:--:|------|
| `web/tsconfig.json` | 21 | TypeScript 编译配置（首次新增） |
| `web/src/types/api.ts` | 31 | 通用 API 响应类型（ApiResponse / PaginatedResult / ApiErrorInfo） |
| `web/src/types/agent.ts` | 188 | Agent 通用类型 + 三类 Agent Result（meta 枚举、EvidenceItem、TopFixItem、Recommendation/MatchAnalysis/ResumeOptimization Result） |
| `web/src/types/job.ts` | 56 | 岗位类型（Job / JobDetail / JobFilterState / SelectedJobState） |
| `web/src/types/resume.ts` | 108 | 简历类型（Resume / ResumeOption / ResumeVersionComparison 等） |
| `web/src/types/ui.ts` | 25 | 前端展示类型（BadgeTone / StatusBadge / PageAction / DecisionPanelState） |
| `web/src/types/index.ts` | 5 | 统一导出 |

## 3. 核心类型说明

### 3.1 Agent meta 契约

`agent.ts` 中定义了与后端 `AgentResponseMeta` 一致的 8 字段结构：

| 字段 | 类型 | 说明 |
|------|------|------|
| `agent_name` | `string` | Agent 名称 |
| `generated_by` | `AgentGeneratedBy` | 生成来源：llm_agent / mock_agent |
| `agent_status` | `AgentStatus` | 运行状态：success / repaired / fallback / mock_success / error |
| `fallback_reason` | `AgentFallbackReason` | 降级原因：11 种枚举值 |
| `error_code` | `AgentErrorCode` | 错误码：12 种枚举值 |
| `repair_applied` | `boolean` | 是否应用 JSON repair |
| `schema_valid` | `boolean` | Schema 校验是否通过 |
| `request_id` | `string` | 请求追踪 ID |

### 3.2 统一 API 响应包装

```ts
ApiResponse<TData>  // HTTP 外层 { code, message, data }
AgentResponseEnvelope<TResult>  // data 内层 { result, meta }
```

### 3.3 三类 Agent Result

- `RecommendationResult`：含 `recommended_jobs`、`evidence_chain`、`risk_items`、`next_actions` 等
- `MatchAnalysisResult`：含 `match_score`、`evidence_chain`、`match_highlights`、`top_fix_items` 等
- `ResumeOptimizationResult`：含 `keyword_alignment`、`project_optimization`、`rewrite_examples`、`final_checklist` 等

### 3.4 共享业务类型

- `TopFixItem`：快速修复任务（被三类 Agent 共用）
- `EvidenceItem`：JD 证据链条目（被 Recommendation/MatchAnalysis 共用）
- `ApplyReadiness` / `OptimizationStopSignal` / `EvidenceStrength` / `PriorityLevel`：共用枚举

## 4. 页面与类型对应关系

| 页面 | 核心类型 |
|------|------|
| 岗位列表页 | `Job`、`JobFilterState` |
| 岗位详情页 | `JobDetail`、`DecisionPanelState` |
| 岗位推荐页 | `RecommendationResult`、`RecommendedJobItem` |
| 岗位匹配分析页 | `MatchAnalysisResult`、`TopFixItem`、`EvidenceItem` |
| 简历优化页 | `ResumeOptimizationResult`、`RewriteExample`、`KeywordAlignmentItem` |
| 我的简历管理首页 | `Resume`、`ResumeProfile`、`ResumeProject`、`ResumeExperience` |
| 简历编辑/添加页 | `Resume`（写入）、`ResumeEducation`、`ResumeProject` |
| 选择简历弹窗 | `ResumeOption` |
| 历史简历比较页 | `ResumeVersionComparison`、`ResumeKeywordChange`、`ResumeProjectChange` |

## 5. 明确本阶段不做

- ❌ 不创建 `web/src/mocks/`
- ❌ 不创建 `web/src/adapters/`
- ❌ 不创建 `web/src/api/`
- ❌ 不修改 `web/src/pages/` 或 `web/src/components/`
- ❌ 不修改后端
- ❌ 不接真实 API

## 6. 后续 5-1.2 / 5-1.3 / 5-1.4 输入

- 5-1.2 Mock 标准化：可基于 `web/src/types/` 创建类型化的 Mock 数据文件
- 5-1.3 Adapter：可基于 `web/src/types/` 定义 `ApiResponse<T>` → 页面 props 的转换函数签名
- 5-1.4 页面改造：可基于 `web/src/types/` 将硬编码数据迁移为类型化 imports

## 7. 构建状态说明

- `npm run build` 因 `node_modules` 未安装（`sh: vite: not found`）无法执行
- TypeScript 类型文件本身语法正确，`tsconfig.json` 已配置 `include: ["src"]`

## 8. 设计约束

- 所有类型文件不引入 React
- 不依赖页面组件
- 不写 Mock 数据或 API 请求
- 不写 adapter 函数
- 不使用 `any`（兼容未知字段用 `string` 联合或 `unknown`）
- 字段与 5-UX.0.1-A 冻结的 Demo 数据一致
- 不引入腾讯招聘完整字段体系（无头像上传、附件上传、身份证、户籍等）
