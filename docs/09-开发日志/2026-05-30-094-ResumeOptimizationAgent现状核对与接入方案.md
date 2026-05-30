# 2026-05-30-094-ResumeOptimizationAgent现状核对与接入方案

## 1. 本次任务

4-3.6：ResumeOptimizationAgent 现状核对与接入方案。

核对当前项目是否存在简历优化的完整工程链路（Prompt / Schema / Agent / Service / Controller / Model / Mock / Route / 接口文档），明确是否具备直接接入 Pipeline 的条件，给出 4-3.7 阶段建议。

## 2. 核对范围

- Prompt / Schema 资产
- Agent 工程代码
- Service 层
- Controller / Route
- Model 层
- Mock / fallback
- 接口文档
- 版本对比能力（与简历优化的边界）

## 3. Prompt / Schema 核对结果

| 资产 | 状态 |
|------|------|
| `server/prompts/resume_optimization_v1.md` | ✅ 存在，4-1.4.7A 人工定稿，4-1.4.8 首轮验收 partial_pass |
| `server/prompts/schema/resume_optimization_schema.json` | ✅ 存在，4-1.4.9A 人工定稿，15 个 required 字段 |
| Prompt Loader 支持 | ✅ 逻辑名称 `resume_optimization` |
| Schema Loader 支持 | ✅ 逻辑名称 `resume_optimization` |
| Pipeline 支持 | ✅ `TargetResumeOptimization` |
| Repair 支持 | ✅ `TargetResumeOptimization` |

**结论：Prompt + Schema 双资产完备，Pipeline 已预留支持。**

## 4. Agent 工程核对结果

| 文件 | 状态 |
|------|------|
| `server/internal/agent/resume_optimization_agent.go` | ❌ 不存在 |
| `server/internal/agent/resume_optimization_mock.go` | ❌ 不存在 |
| `server/internal/agent/version_compare_mock.go` | ✅ 存在（但这是版本对比，不是简历优化） |

**结论：无 ResumeOptimizationAgent 和 ResumeOptimizationMockAgent。**

## 5. Service / Controller / Model 核对结果

| 层级 | 文件 | 状态 |
|------|------|------|
| Service | `resume_optimization_service.go` | ❌ 不存在 |
| Service | `version_compare_service.go` | ✅ 存在（版本对比） |
| Controller | `resume_optimization_handler.go` | ❌ 不存在 |
| Controller | `version_compare_handler.go` | ✅ 存在（版本对比） |
| Model | `resume_optimization.go` | ❌ 不存在 |
| Model | `version_compare.go` | ✅ 存在（版本对比） |
| Route | `POST /api/resumes/optimization` | ❌ 不存在 |
| Route | `GET /api/resumes/versions/compare` | ✅ 存在（版本对比） |

**结论：无简历优化业务代码，只有版本对比。两者不是同一个能力。**

## 6. 接口文档核对结果

| 文档 | 状态 |
|------|------|
| `docs/04-接口文档/简历优化接口.md` | ❌ 不存在 |
| `docs/04-接口文档/简历接口.md` | ✅ 存在（覆盖简历版本管理和版本对比） |

**结论：无独立简历优化接口文档。**

## 7. 版本对比能力核对结果

版本对比（VersionCompare）和简历优化（ResumeOptimization）的区别：

| 维度 | 版本对比 | 简历优化 |
|------|---------|---------|
| 输入 | 旧版本 + 新版本 | 当前简历 + 目标岗位 JD |
| 核心问题 | "我改了什么、变好了吗" | "怎么改才能通过初筛、什么时候可以投" |
| 是否需要 JD | ❌ | ✅ |
| 真实性风险 | 低 | 高 |
| 战略差异化关联 | 无 | 核心杀手锏之一 |

**结论：版本对比不能替代简历优化。当前只有版本对比能力，没有简历优化能力。**

## 8. 是否可直接接入 Pipeline

**结论：不能。**

缺少 6 项核心业务代码：
1. ResumeOptimizationAgent（LLM 版）
2. ResumeOptimizationMockAgent（fallback 保障）
3. ResumeOptimizationService
4. ResumeOptimizationHandler
5. ResumeOptimization Model（Request / Response）
6. 路由注册

对比 MatchAnalysisAgent / RecommendationAgent 的接入路径，它们在阶段 3 已有完整的 Mock Agent + Service + Handler + Model + Route，4-3 只需新增 LLM Agent 和改造 Service。但 ResumeOptimizationAgent 连阶段 3 的业务代码都没有。

## 9. 4-3.7 建议

**主推荐：4-3.7 先补 ResumeOptimization mock / service / controller，再接 Pipeline。**

分步执行：
- 4-3.7.A：创建 ResumeOptimizationMockAgent + Model + Service + Handler + Route
- 4-3.7.B：ResumeOptimizationAgent 接入 Pipeline
- 4-3.7.C：简历优化接口联调与 fallback 验收

接入时必须吸收 4-3.5.B ~ 4-3.5.G 的战略边界：
- 不做无限简历优化
- 不直接输出"建议改为"
- 使用"示例参考"
- 保留用户核实提醒
- 围绕 JD 证据链
- 区分 Top 3 快速优化项和长期提升项

## 10. 文档更新

本次新增：
- `docs/05-Agent设计/ResumeOptimizationAgent现状核对与接入方案.md`
- `docs/09-开发日志/2026-05-30-094-ResumeOptimizationAgent现状核对与接入方案.md`

本次修改：
- 更新阶段 4-3 业务 Agent 接入 LLM 方案
- 更新 Agent 总览
- 更新 Prompt 模板文档
- 更新 JSON 输出规范
- 更新失败兜底与 Mock 回退策略
- 更新阶段 4-3.5 战略转型汇总与工程评审
- 更新项目状态
- 更新下一步任务
- 更新说明书

## 11. 未修改内容

1. 未修改 Go 代码
2. 未修改 Prompt 文件
3. 未修改 Schema 文件
4. 未修改前端
5. 未调用 Hunyuan
6. 未请求腾讯招聘接口
7. 未进入 4-3.7
8. 本次只做 ResumeOptimizationAgent 现状核对与接入方案

## 12. 当前风险

1. **ResumeOptimizationAgent 补齐业务代码工作量较大**：需要从零创建 Mock Agent / Service / Handler / Model / Route
2. **简历优化涉及真实性风险**：必须确保 Mock Agent 和 LLM Agent 都不编造经历
3. **当前 Schema 不完全匹配新战略字段**：缺少 apply_readiness / optimization_stop_signal / jd_evidence_chain 等字段，但不阻塞当前接入
4. **版本对比与简历优化的边界需在接口层面明确**：避免前端混用

## 13. 是否可以进入下一步

可以进入 4-3.7。4-3.7 应先补简历优化业务代码（Mock Agent / Service / Handler / Model / Route），再接入 Pipeline。

## 14. 下一步建议

进入 4-3.7.A：创建 ResumeOptimizationMockAgent + Model + Service + Handler + Route。
