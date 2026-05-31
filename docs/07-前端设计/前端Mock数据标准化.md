# 前端 Mock 数据标准化

> 阶段：5-1.2 | 状态：已完成 | 日期：2026-05-31

## 1. 本阶段目标

基于 5-1.1 已建立的 TypeScript 类型契约，创建统一的前端 Mock 数据目录，将原本散落在 9 个页面、25+ 组件中的硬编码数据集中管理，为后续 5-1.4 页面改造和 5-2 API Client 接入打基础。

## 2. 新增 Mock 文件

| 文件 | 行数 | 说明 |
|------|:--:|------|
| `web/src/mocks/common.ts` | 20 | 共享工具函数（`createMockAgentMeta`） |
| `web/src/mocks/jobs.ts` | 189 | 岗位列表（6 个）+ 详情（3 个）+ 快捷查找 |
| `web/src/mocks/resumes.ts` | 360 | 3 个简历版本 + 弹窗选项 + 历史比较 |
| `web/src/mocks/recommendation.ts` | 149 | Top 5 推荐结果 + AgentResponseEnvelope |
| `web/src/mocks/matchAnalysis.ts` | 96 | 匹配分析结果 + AgentResponseEnvelope |
| `web/src/mocks/resumeOptimization.ts` | 124 | 简历优化结果 + AgentResponseEnvelope |
| `web/src/mocks/index.ts` | 6 | 统一导出 |

**合计：7 个文件，944 行**

## 3. Mock 数据覆盖范围

| 页面 | Mock 数据来源 | 关键导出 |
|------|------|------|
| 岗位列表页 | `jobs.ts` | `mockJobs` |
| 岗位详情页 | `jobs.ts` | `mockJobDetails`、`mockSelectedJob`、`getMockJobById()` |
| 岗位推荐页 | `recommendation.ts` | `mockRecommendationResult`、`mockRecommendationResponse` |
| 岗位匹配分析页 | `matchAnalysis.ts` | `mockMatchAnalysisResult`、`mockMatchAnalysisResponse` |
| 简历优化页 | `resumeOptimization.ts` | `mockResumeOptimizationResult`、`mockResumeOptimizationResponse` |
| 我的简历管理首页 | `resumes.ts` | `mockCurrentResume`、`mockResumeVersions` |
| 简历编辑/添加页 | `resumes.ts` | `mockCurrentResume`（写入参照） |
| 选择简历弹窗 | `resumes.ts` | `mockResumeOptions` |
| 历史简历比较页 | `resumes.ts` | `mockResumeComparison` |

## 4. 统一 Demo 数据

- 候选人：测试候选人 A
- 当前简历：V2 后端开发简历（`mockCurrentResume`）
- 历史简历：V1 通用后端简历
- 方向不匹配简历：V3 产品方向简历
- 核心技能：Go / MySQL / Redis / Docker / 微服务 / RBAC / 接口开发 / 缓存优化
- 核心项目：Go 微服务订单系统、RBAC 权限管理系统、缓存优化项目
- 联系方式已脱敏：182****9164 / candidate@example.com

## 5. Agent Mock Response 结构

所有三类 Agent Mock 响应统一使用 `AgentResponseEnvelope<T>` 结构：

```ts
{
  result: T,    // 业务结果
  meta: {       // 系统元数据
    agent_name: string,
    generated_by: 'mock_agent',
    agent_status: 'mock_success',
    fallback_reason: 'none',
    error_code: 'none',
    repair_applied: false,
    schema_valid: true,
    request_id: string,
  }
}
```

`meta` 通过 `common.ts` 的 `createMockAgentMeta()` 统一生成，避免 3 处重复。

## 6. 与 types 的关系

- 所有 Mock 数据均从 `web/src/types/` 导入对应类型
- 不使用 `any`
- 不引入 React
- 字段完全匹配 5-1.1 类型定义

## 7. 明确本阶段不做

- ❌ 不修改页面组件
- ❌ 不创建 `web/src/adapters/`
- ❌ 不创建 `web/src/api/`
- ❌ 不接后端接口
- ❌ 不引入 fetch / axios

## 8. 后续 5-1.3 / 5-1.4 输入

- 5-1.3 Adapter：可基于 `mocks/` 和 `types/` 定义数据转换函数签名
- 5-1.4 页面改造：可将页面内硬编码数据替换为 `import { mockJobs } from '../mocks'`
