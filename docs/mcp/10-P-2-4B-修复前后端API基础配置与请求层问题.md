# 10-P-2-4B - 修复前后端 API 基础配置与请求层问题 - MCP 上下文记录

> 执行时间：2026-06-01 | 状态：pass（API fallback 已移除，build 通过）
> 是否修改代码：是（仅请求层和上游调用组件）
> 上一阶段：10-P-2-4A（补齐工程化验证闭环）

---

## 1. 阶段信息

| 项目 | 内容 |
|---|---|
| 当前阶段编号 | 10-P-2-4B（编号冲突：已有 10-P-2-4A=工程化验证闭环，本次用 -B 后缀） |
| 项目名称 | Offer 捕手 — 学生求职匹配智能体 |
| 暂停阶段 | 阶段 11、12、13 |

## 2. 本次目标

移除前端 API 层的静默 fallback 到 mock 数据，让请求失败时错误正确暴露。统一 API baseURL、token 注入、错误处理。修复因 API 签名变更导致的上游调用断裂。

## 3. 本次执行范围

**实际执行了：**
- 移除 `api/jobs.ts` 的 `withFallback` 和 mock import（P0-4 修复）
- 移除 `api/resumes.ts` 的 `withFallback` 和 mock import
- 移除 `api/agents.ts` 的 `withMockFallback` 和 mock import
- 修复 `getSelectedJob` 签名：需要 `jobId` 参数
- 修复 `getResumeDetail` 签名：`resumeId: string` → `versionNo: number`
- 修复 `runRecommendation` 默认 `resume_version_no`: 2→0
- 移除 `runResumeOptimization` 硬编码默认值
- 修复 `viewModels.ts` 的 fallback 链和冲突调用
- 修复 `JobDetailPage`/`JobDetailBanner`/`JobDetailJD` 传递 jobId
- 修复 `AIDecisionWorkspace` 接收 selectedJobId prop
- 验证 `npm run build` (2.10s) 和 `go build` 通过

**仅检查（未修改）：**
- Vite proxy/Nginx/CORS 配置（已确认对齐）
- token 注入逻辑（client.ts 已统一）

## 4. 明确未做事项

| 未做事项 | 原因 |
|---|---|
| 未修复前端组件顶层 import mock（P0-3） | 组件层，非请求层 |
| 未实现 Jobs facets | 非本阶段 |
| 未修改 Agent 业务逻辑 | 禁止 |
| 未继续 LLM | 阶段 11 暂停 |

## 5. 修改文件清单

| 文件 | 修改类型 | 修改说明 |
|---|---|---|
| `web/src/api/jobs.ts` | 请求层 | 移除 withFallback + mock import；getSelectedJob 加 jobId 参数 |
| `web/src/api/resumes.ts` | 请求层 | 移除 withFallback + mock import；getResumeDetail 参数改为 versionNo(number)；setCurrentResume 改为 throw |
| `web/src/api/agents.ts` | 请求层 | 移除 withMockFallback + mock import；默认值清理 |
| `web/src/api/viewModels.ts` | 请求层 | loadSelectedJobDetailViewModel 加 jobId；loadResumeOptionViewModelsFromApi 移除 fallback 链 |
| `web/src/app/components/AIDecisionWorkspace.tsx` | 组件 | 新增 selectedJobId prop（可选） |
| `web/src/app/components/JobDetailBanner.tsx` | 组件 | 新增 jobId prop |
| `web/src/app/components/JobDetailJD.tsx` | 组件 | 新增 jobId prop |
| `web/src/app/pages/JobDetailPage.tsx` | 路由页 | useParams 获取 jobId 并下传子组件 |

## 6. 新增文件清单

无。

## 7. 新增或更新 docs 清单

| 文件 | 新增/更新 | 目的 |
|---|---|---|
| `docs/mcp/项目上下文.md` | 更新 | 10-P-2-4B 状态、API 请求层修复结果、P0-4 标记已修复 |
| `docs/mcp/10-P-2-4B-修复前后端API基础配置与请求层问题.md` | 新增 | 本阶段 MCP 记录 |
| `docs/09-开发日志/2026-06-01-227-10-P-2-4B-修复API请求层.md` | 新增 | 开发日志 |

## 8. 关键决策

| 决策 | 原因 | 影响 |
|---|---|---|
| 使用 10-P-2-4B 编号 | 10-P-2-4A 已被工程化验证闭环占用 | 避免覆盖已有文档 |
| getSelectedJob 改为需要 jobId | 原先使用硬编码 mock ID 调用 | 调用方必须传入真实 jobId |
| setCurrentResume 改为 throw | 后端无此接口，前端不应假装可用 | 调用方必须改为 createResume/updateResume |
| agent 默认 resume_version_no 改为 0 | 0 让后端自动使用当前简历版本 | 比硬编码 2 更合理 |
| viewModels 的 loadResumeOptionViewModelsFromApi 移除 fallback 链 | API 失败应由调用方处理 | 错误正确传播 |

## 9. 验收证据

| 验收项 | 方法 | 结果 |
|---|---|---|
| git status 初始 | `git status --short` | 干净 |
| withFallback 残留 | `grep -rn 'withFallback\|withMockFallback' web/src/api/` | 零匹配 |
| mock import 残留 | `grep -rn 'import.*mocks' web/src/api/` | 零匹配 |
| npm run build | `cd web && npm run build` | 2.10s, 1683 modules |
| go build | `go build ./cmd/server` | PASS |
| git status 最终 | `git status --short` | 8 文件修改（均在预期范围） |

## 10. 当前风险

| 风险 | 严重度 | 阻塞什么 |
|---|---|---|
| P0-3（5组件顶层 import mock）未修复 | P0 | 推荐/匹配/优化页核心内容仍为假数据 |
| JobListPage 未下传 selectedJobId 给 AIDecisionWorkspace | P2 | 首页右侧 AI 面板无法加载当前选中岗位标题 |
| resume 相关 caller 可能使用 string resumeId | P2 | 需改为 number versionNo |

## 11. 不确定项

1. **JobListPage 中 selectedJobId 未传递给 AIDecisionWorkspace**：`JobList.tsx` 有 `selectedJobId` state 但 `JobListPage` 未将其传给 `AIDecisionWorkspace`。本阶段仅做了最小修复（AIDecisionWorkspace 接受可选 prop），未实现完整数据流。
2. **resume 相关调用方是否已适配 versionNo(number)**：`loadResumeSummaryViewModel` 参数改为 number，调用方可能仍传 string。

## 12. 下次继续前必须读取

1. `docs/mcp/项目上下文.md` — 当前阶段和 P0 状态
2. `docs/mcp/10-P-2-4A-补齐工程化验证闭环.md` — 上一阶段工程验证
3. 本文件 — API 请求层修复记录

## 13. 用户明确要求 / 禁止事项

| 要求/禁止 | 来源 |
|---|---|
| 阶段 11-13 暂停 | 用户全局 |
| 不修前端组件 mock import | 本阶段 prompt |
| 不实现 Jobs facets | 本阶段 prompt |
| 不继续 LLM | 项目背景 |
| 不允许 git commit | 本阶段 prompt |

## 14. CodeBuddy 重启恢复摘要

本次执行：移除 jobs.ts/resumes.ts/agents.ts 的全部 withFallback/withMockFallback；移除 api/ 目录全部 mock import；修复上游组件断裂。API 请求失败现在正确抛出而非静默返回假数据。npm build 2.10s pass。

关键修复：P0-4（API fallback 掩盖失败）已解决。P0-3（组件层 mock import）未修复。

输出：3 docs 文件（项目上下文更新 + MCP 记录 + 开发日志）；8 代码文件修改。
