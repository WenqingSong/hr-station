# ResumeOptimizationAgent 现状核对与接入方案

> 创建时间：2026-05-30
> 所属阶段：4-3.6（4-3.7 已补齐 Mock 业务链路）
> 依据文档：阶段4-3业务Agent接入LLM方案、Agent总览、Prompt模板、JSON输出规范、失败兜底与Mock回退策略、阶段4-3.5战略转型汇总与工程评审、战略差异化与杀手锏确认、投递边界与优化停止条件设计、JD证据链匹配与START简历优化规则、Agent输出QualityGate设计、学生求职极端情况处理边界、灵活在线简历Profile与纯文字Agent上下文设计

## 1. 背景

4-3.5.H 已完成战略转型阶段汇总与工程评审，评审结论：不立即大改工程，先进入 4-3.6 ResumeOptimizationAgent 现状核对与接入方案。

MatchAnalysisAgent 和 RecommendationAgent 已分别完成 Pipeline 接入和接口验收（4-3.2 ~ 4-3.5）。ResumeOptimizationAgent 是阶段 4-3 三类业务 Agent 的最后一个，当前需要明确其工程现状，判断是否具备直接接入 Pipeline 的条件。

## 2. 核对目标

1. 确认 Prompt / Schema 资产是否存在
2. 确认 Agent 工程代码是否存在
3. 确认 Service / Controller / Model 是否存在
4. 确认 Mock / fallback 是否存在
5. 确认接口和路由是否存在
6. 明确是否已有独立"简历优化业务接口"
7. 明确是否只有"版本对比"能力
8. 明确是否能直接接入 Pipeline
9. 给出 4-3.7 阶段建议

## 3. 当前 Prompt / Schema 资产核对

| 检查项 | 结果 | 说明 |
|--------|------|------|
| `server/prompts/resume_optimization_v1.md` | ✅ 存在 | 简历优化 Agent 指令模板，4-1.4.7A 人工定稿，4-1.4.7B CodeBuddy 登记 |
| 是否被 Prompt Loader 支持 | ✅ 支持 | Prompt Loader 支持逻辑名称 `resume_optimization` |
| 是否在 docs 中登记 | ✅ 已登记 | 4-1.4.7B 已登记，Prompt模板.md 中已记录 |
| 是否经过 4-1.4 首轮验收 | ✅ partial_pass | 4-1.4.8 已完成首轮输出验收 |
| 当前是否已接入 Go 后端 | ❌ 未接入 | 仅 Prompt 文件存在，未被 Service / Agent 层调用 |
| `server/prompts/schema/resume_optimization_schema.json` | ✅ 存在 | 简历优化 JSON Schema，4-1.4.9A 人工定稿，4-1.4.9B CodeBuddy 登记 |
| 是否 JSON 合法 | ✅ 合法 | 15 个顶层 required 字段，10 个 $defs 子结构 |
| 是否被 Schema Loader 支持 | ✅ 支持 | Schema Loader 支持逻辑名称 `resume_optimization` |
| 是否在 docs 中登记 | ✅ 已登记 | 4-1.4.9B 已登记，JSON输出规范.md 中已记录 |
| 是否经过 4-1.4 首轮验收 | ✅ Schema 已通过 | Schema 已吸收首轮验收问题 |
| 当前是否已接入 Go 后端 | ❌ 未接入 | 仅 Schema 文件存在，未被 Pipeline 外的 Service 调用 |

**Prompt 资产总结**：Prompt + Schema 双资产完备，Pipeline 已支持 `TargetResumeOptimization`，但尚未被业务代码调用。

## 4. 当前 Agent 工程代码核对

| 检查项 | 结果 | 说明 |
|--------|------|------|
| `server/internal/agent/resume_optimization_agent.go` | ❌ 不存在 | 无 ResumeOptimizationAgent 实现 |
| `server/internal/agent/resume_optimization_mock.go` | ❌ 不存在 | 无简历优化 Mock Agent |
| `server/internal/agent/version_compare_mock.go` | ✅ 存在 | 但这是"版本对比"，不是"简历优化" |
| Pipeline 是否支持 resume_optimization | ✅ 支持 | `TargetResumeOptimization` 已定义在 `pipeline/types.go` |
| Pipeline 测试是否覆盖 resume_optimization | ✅ 覆盖 | `pipeline_test.go` 中有 `TestPipeline_ResumeOptimization_FirstSuccess` |
| Repair 是否支持 resume_optimization | ✅ 支持 | `repair/json_repair_agent.go` 支持 `TargetResumeOptimization` |
| 是否存在单元测试 | ⚠️ 仅 Pipeline/Repair 级别 | Pipeline 和 Repair 的测试使用了 `minimalResumeOptimization()` fixture，但无业务 Agent 级别测试 |

**Agent 代码总结**：无 `resume_optimization_agent.go`，无 `resume_optimization_mock.go`。Pipeline 和 Repair 层面已预留 `TargetResumeOptimization` 支持，但缺少业务 Agent 实现。

## 5. 当前 Service 层核对

| 检查项 | 结果 | 说明 |
|--------|------|------|
| `server/internal/service/resume_optimization_service.go` | ❌ 不存在 | 无简历优化服务 |
| `server/internal/service/version_compare_service.go` | ✅ 存在 | 但这是"版本对比"服务，不是"简历优化"服务 |
| 是否已经调用 Agent | ❌ 否 | 无简历优化服务，无 Agent 调用 |
| 是否已经调用 Pipeline | ❌ 否 | 无简历优化服务，无 Pipeline 调用 |

**Service 层总结**：无简历优化服务，只有版本对比服务。两者不是同一个能力。

## 6. 当前 Controller / Route 核对

| 检查项 | 结果 | 说明 |
|--------|------|------|
| `server/internal/http/resume_optimization_handler.go` | ❌ 不存在 | 无简历优化 HTTP Handler |
| `server/internal/http/version_compare_handler.go` | ✅ 存在 | 但这是"版本对比"Handler |
| 简历优化接口路由 | ❌ 不存在 | 无 `POST /api/resumes/optimization` |
| 版本对比接口路由 | ✅ 存在 | `GET /api/resumes/versions/compare` |

**Controller / Route 总结**：无简历优化接口和路由。

## 7. 当前 Model 层核对

| 检查项 | 结果 | 说明 |
|--------|------|------|
| `server/internal/model/resume_optimization.go` | ❌ 不存在 | 无简历优化 Model |
| `server/internal/model/version_compare.go` | ✅ 存在 | 但这是"版本对比"Model |
| `ResumeOptimizationRequest` | ❌ 不存在 | 无简历优化请求结构 |
| `ResumeOptimizationResponse` | ❌ 不存在 | 无简历优化响应结构 |
| `generated_by` 字段 | ✅ 已在 match.go / recommendation.go 中 | 但简历优化无 Model 定义 |

**Model 层总结**：无简历优化相关 Model。

## 8. 当前 Mock / fallback 核对

| 检查项 | 结果 | 说明 |
|--------|------|------|
| `server/internal/agent/resume_optimization_mock.go` | ❌ 不存在 | 无简历优化 Mock |
| `server/internal/agent/version_compare_mock.go` | ✅ 存在 | 但这是"版本对比"Mock，不是"简历优化"Mock |
| fallback 触发条件 | ❌ 未定义 | 无简历优化 fallback 逻辑 |
| generated_by 是否正确 | ❌ 不适用 | 无简历优化业务代码 |
| 是否已有 Mock fallback 稳定性测试 | ❌ 不适用 | 无简历优化业务代码 |

**Mock / fallback 总结**：无简历优化 Mock Agent，无法提供 fallback 保障。这是接入 Pipeline 的关键缺失项。

## 9. 当前接口文档核对

| 检查项 | 结果 | 说明 |
|--------|------|------|
| `docs/04-接口文档/简历优化接口.md` | ❌ 不存在 | 无简历优化接口文档 |
| `docs/04-接口文档/简历版本对比接口.md` | ❌ 不存在（含在简历接口.md 中） | 版本对比接口文档在简历接口.md 内 |
| `docs/04-接口文档/简历接口.md` | ✅ 存在 | 覆盖简历版本管理和版本对比 |
| 文档和工程是否一致 | ⚠️ 部分一致 | 版本对比文档和工程一致，但无简历优化文档 |

**接口文档总结**：无独立简历优化接口文档。

## 10. 当前是否已有独立简历优化业务接口

**结论：没有。**

证据：

- 文件路径：无 `server/internal/http/resume_optimization_handler.go`
- 代码位置：无简历优化 Service / Agent / Model
- 文档位置：无 `docs/04-接口文档/简历优化接口.md`
- 路由位置：无 `POST /api/resumes/optimization` 路由

当前项目只有以下简历相关接口：

| 接口 | 说明 |
|------|------|
| `POST /api/resumes` | 新增简历版本 |
| `GET /api/resumes/current` | 获取当前简历 |
| `GET /api/resumes/detail` | 获取简历详情 |
| `GET /api/resumes/versions` | 获取版本列表 |
| `GET /api/resumes/versions/{versionNo}` | 获取指定版本 |
| `GET /api/resumes/versions/compare` | 版本对比（Mock） |

其中版本对比接口（`GET /api/resumes/versions/compare`）只是比较两个版本的文本差异，不是基于目标岗位 JD 的简历优化。

## 11. 当前是否只有版本对比能力

**结论：是。**

当前"简历相关 Agent 能力"只有版本对比（VersionCompareMockAgent），但版本对比和简历优化是两个不同的能力：

| 维度 | 版本对比 | 简历优化 |
|------|---------|---------|
| 输入 | 旧版本简历 + 新版本简历 | 当前简历 + 目标岗位 JD |
| 输出 | 分数变化 + 字段变化 + 风险变化 | P0/P1/P2 优化建议 + JD 证据链 + apply_readiness + 示例参考 |
| 核心问题 | "我改了什么、变好了吗" | "我该怎么改才能通过初筛、什么时候可以投" |
| 是否需要目标岗位 JD | ❌ 不需要 | ✅ 必须有 |
| 是否涉及真实性风险 | ⚠️ 低 | 🔴 高（不能编造经历、指标） |
| 战略差异化关联 | ❌ 无 | ✅ 核心杀手锏之一 |

**版本对比不能替代简历优化。** 版本对比只告诉用户改了什么，简历优化才是基于目标岗位 JD 帮助用户判断怎么改、什么时候改完可以投。

## 12. ResumeOptimizationAgent 与版本对比的边界

### 12.1 版本对比（VersionCompare）

- 比较两份简历文本差异
- 告诉用户哪里变好了、哪里变差了
- 不需要目标岗位 JD
- 不做"建议改为"
- 不涉及 JD 证据链
- 不涉及投递边界

### 12.2 简历优化（ResumeOptimization）

- 基于目标岗位 JD，告诉用户当前简历哪里影响初筛
- 给出优化方向：不是"建议改为"，而是"示例参考"
- 围绕 JD 证据链：JD 要求 → 简历证据 → 匹配判断 → 差距风险 → 示例参考 → 核实提醒
- 明确投递边界：apply_readiness / optimization_stop_signal
- 区分 Top 3 快速优化项和长期提升项
- 不能编造经历、指标、项目、职责
- 每条建议必须可核实

### 12.3 关键区别

```
版本对比：比较 → 评价 → "你改了这些，效果如何"
简历优化：分析 → 定位 → "针对这个岗位，你的简历哪里影响初筛，怎么改，改到什么程度可以投"
```

两者不是同一个能力，不能混用。

## 13. 是否具备直接接入 Pipeline 的条件

**结论：不具备。**

当前 ResumeOptimizationAgent 接入 Pipeline 需要以下前提：

| 前提 | 当前状态 | 说明 |
|------|---------|------|
| Prompt 文件 | ✅ 已存在 | resume_optimization_v1.md |
| Schema 文件 | ✅ 已存在 | resume_optimization_schema.json |
| Pipeline 支持 | ✅ 已支持 | TargetResumeOptimization |
| Repair 支持 | ✅ 已支持 | TargetResumeOptimization |
| ResumeOptimizationAgent（Go 代码） | ❌ 不存在 | 需要 PromptValues 组装 + Pipeline 调用 + 字段映射 |
| ResumeOptimizationMockAgent | ❌ 不存在 | 需要 fallback 保障 |
| ResumeOptimizationService | ❌ 不存在 | 需要业务编排（LLM 优先 + Mock fallback） |
| ResumeOptimization Model | ❌ 不存在 | 需要 Request / Response 结构 |
| ResumeOptimization Handler | ❌ 不存在 | 需要 HTTP Handler |
| 路由 | ❌ 不存在 | 需要注册路由 |
| 接口文档 | ❌ 不存在 | 需要接口文档 |

**缺 6 项核心业务代码，不能直接接 Pipeline。**

参考 MatchAnalysisAgent 和 RecommendationAgent 的接入路径：

```
4-3.2 MatchAnalysisAgent 接入 Pipeline
  → 已有 MatchMockAgent（阶段 3-7 创建）
  → 已有 MatchService（阶段 3-7 创建）
  → 已有 MatchHandler（阶段 3-7 创建）
  → 已有 Match Model（阶段 3-7 创建）
  → 已有匹配分析接口和路由（阶段 3-7 创建）
  → 4-3.2 只需新增 MatchAnalysisAgent（LLM 版）+ 改造 Service
```

```
4-3.4 RecommendationAgent 接入 Pipeline
  → 已有 RecommendationMockAgent（阶段 3-6 创建）
  → 已有 RecommendationService（阶段 3-6 创建）
  → 已有 RecommendationHandler（阶段 3-6 创建）
  → 已有 Recommendation Model（阶段 3-6 创建）
  → 已有推荐接口和路由（阶段 3-6 创建）
  → 4-3.4 只需新增 RecommendationAgent（LLM 版）+ 改造 Service
```

对比之下：

```
ResumeOptimizationAgent 接入 Pipeline
  → ❌ 无 ResumeOptimizationMockAgent
  → ❌ 无 ResumeOptimizationService
  → ❌ 无 ResumeOptimizationHandler
  → ❌ 无 ResumeOptimization Model
  → ❌ 无简历优化接口和路由
  → 需要先补齐所有业务代码，才能接入 Pipeline
```

## 14. 理想接入链路

```
Controller (ResumeOptimizationHandler)
→ ResumeOptimizationService
→ ResumeOptimizationAgent
→ Pipeline.Run
→ JSON parse
→ Schema validate
→ repair once
→ fallback_required decision
→ Mock fallback 或 conservative fallback
→ 返回用户可用结果
```

### 14.1 每层需要创建的内容

| 层级 | 需要创建 | 说明 |
|------|---------|------|
| Model | `ResumeOptimizationRequest` | 用户 ID + 简历版本 ID + 目标岗位 ID + 可选匹配分析 ID |
| Model | `ResumeOptimizationResult` | 基于 resume_optimization_schema.json 的 Go 结构 |
| Model | `ResumeOptimizationResponse` | 接口响应结构 |
| Mock Agent | `ResumeOptimizationMockAgent` | 确定性 Mock 输出，用于 fallback |
| LLM Agent | `ResumeOptimizationAgent` | PromptValues 组装 + Pipeline 调用 + 字段映射 + ResumeOptFallbackError |
| Service | `ResumeOptimizationService` | LLM 优先 + Mock fallback + 快照保存 |
| Handler | `ResumeOptimizationHandler` | HTTP 请求处理 |
| Route | `POST /api/resumes/optimization` | 路由注册 |

## 15. 输入设计建议

### 15.1 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| user_id | uint | ✅ | 从 JWT context 获取 |
| resume_version_id | uint | ✅ | 简历版本 ID |
| job_id | uint | ✅ | 目标岗位 ID |
| match_analysis_id | uint | ❌ | 可选，已有匹配分析结果时传入 |

### 15.2 PromptValues 来源

| 占位符 | 数据来源 | 说明 |
|--------|---------|------|
| `{{RESUME_CONTEXT_JSON}}` | model.ResumeVersion 序列化为 JSON | 当前仍使用 JSON 序列化，后续迁移为 ResumeContextText |
| `{{TARGET_JOB_CONTEXT_JSON}}` | model.Job 序列化为 JSON | 目标岗位完整 JD |
| `{{MATCH_ANALYSIS_JSON}}` | 可选，来自匹配分析结果 | 如果有已有匹配分析，传入辅助优化 |
| `{{BUSINESS_CONTEXT}}` | 业务 Agent 硬编码/常量 | 目标和约束 |

### 15.3 后续迁移方向

根据 4-3.5.G 灵活在线简历 Profile 设计，输入应从 `RESUME_CONTEXT_JSON` 迁移为：

```
{{RESUME_CONTEXT_TEXT}} + {{PROFILE_SUMMARY_JSON}} + {{PRIVACY_FLAGS_JSON}}
```

当前阶段仍使用 `RESUME_CONTEXT_JSON`，迁移放到后续单独评估。

## 16. 输出设计原则

必须符合 4-3.5 战略转型成果：

1. **不写"建议改为"**：使用"示例参考"
2. **必须有 verification_notice**：提醒用户 AI 建议仅供参考，需自行核实
3. **必须有 self_check_question**：帮助用户确认建议是否适用
4. **必须有 learn_more_query**：用户可进一步搜索学习的方向
5. **必须区分 top_fix_items 与 long_term_improvement**：快速优化项与长期提升项分开展示
6. **必须有 optimization_stop_signal**：明确优化停止信号
7. **必须基于 JD 证据链**：每条建议都要绑定 JD 要求和简历证据
8. **不能编造经历、指标、项目、职责**：真实性红线

如果当前 Schema 不支持上述所有字段（如 apply_readiness / optimization_stop_signal / jd_evidence_chain），应在接入方案中标记为后续改造项。

## 17. fallback 设计原则

| 场景 | 处理方式 |
|------|---------|
| LLM 不可用 | → ResumeOptimizationMockAgent fallback |
| JSON parse 失败 | → repair once |
| Schema validate 失败 | → repair once |
| repair 失败 | → ResumeOptimizationMockAgent fallback |
| 内容质量不达标 | → 后续 Quality Gate 或 conservative fallback |
| 字段映射失败 | → ResumeOptimizationMockAgent fallback |

关键前提：必须先创建 `ResumeOptimizationMockAgent`，否则无法提供 fallback 保障。

## 18. 与 4-3.5 战略转型成果的关系

### 18.1 投递边界（4-3.5.C）

- ResumeOptimizationAgent 应输出 `apply_readiness` 和 `optimization_stop_signal`
- 但当前 Schema 未包含这些字段
- **接入方案建议**：在 4-3.7 接入时评估是否需要扩展 Schema，如果不扩展则标记为后续改造

### 18.2 JD 证据链（4-3.5.D）

- ResumeOptimizationAgent 应围绕 JD 证据链输出优化建议
- 当前 Schema 的 `keyword_alignment` / `project_optimization` / `skill_section_optimization` 等字段部分体现了证据链思路
- 但缺少结构化的 `jd_evidence_chain` 字段
- **接入方案建议**：利用现有 Schema 字段实现 JD 证据链逻辑，`jd_evidence_chain` 作为后续扩展

### 18.3 Quality Gate（4-3.5.E）

- 4-3.5.H 已评审：不立即实现完整 Quality Gate
- ResumeOptimizationAgent 接入时暂不接入 Quality Gate
- **接入方案建议**：后续如实现 Quality Gate，可接入简历优化链路

### 18.4 极端情况处理（4-3.5.F）

- ResumeOptimizationAgent 应识别极端情况（如信息不足、岗位不匹配）
- 当前 Schema 未包含 `candidate_case_type` 字段
- **接入方案建议**：在 Prompt 中引导模型识别极端情况，输出中用现有字段表达（如 `authenticity_notice`），`candidate_case_type` 作为后续扩展

### 18.5 灵活简历 Profile（4-3.5.G）

- ResumeOptimizationAgent 当前仍使用 `json.Marshal(resumeVersion)` 作为输入
- 后续迁移为 ResumeContextText + profile_summary + privacy_flags
- **接入方案建议**：4-3.7 先用现有输入格式接入，迁移放到后续

## 19. 4-3.7 阶段建议

### 19.1 推荐方案

**4-3.7 应定义为：先补简历优化业务接口方案（docs only），再接 Pipeline。**

理由：

1. 当前缺 6 项核心业务代码（Mock Agent / LLM Agent / Service / Handler / Model / Route）
2. 不能像 MatchAnalysisAgent / RecommendationAgent 那样只改 Service 层就接入 Pipeline
3. 必须先补齐完整的业务链路，才能接入 Pipeline
4. 补业务代码的工作量较大，应先做方案设计

### 19.2 4-3.7 可选定义

| 选项 | 定义 | 说明 |
|------|------|------|
| A | 4-3.7 ResumeOptimizationAgent 接入 Pipeline | 直接写代码补齐所有业务层 + 接入 Pipeline |
| B | 4-3.7 简历优化业务接口方案 docs only | 先设计完整方案，4-3.8 再写代码 |
| C | 4-3.7 ResumeOptimizationAgent 工程预留与版本对比边界确认 | 只做工程预留，不写完整业务代码 |
| D | 4-3.7 先补 ResumeOptimization mock / service / controller，再接 Pipeline | 分步实现：先补 Mock → 再补 Service → 最后接 Pipeline |

### 19.3 主推荐

**推荐选项 D：先补 ResumeOptimization mock / service / controller，再接 Pipeline。**

分步执行：

1. **4-3.7.A**：创建 ResumeOptimizationMockAgent + Model + Service + Handler + Route
2. **4-3.7.B**：ResumeOptimizationAgent 接入 Pipeline
3. **4-3.7.C**：简历优化接口联调与 fallback 验收

这样与 MatchAnalysisAgent / RecommendationAgent 的接入路径保持一致：

```
阶段 3: Mock Agent + Service + Handler + Route（匹配/推荐已有）
阶段 4-3: LLM Agent 接入 Pipeline + 改造 Service（匹配/推荐已完成）
```

但 ResumeOptimizationAgent 缺少"阶段 3"的业务代码，需要先补齐。

## 20. 当前不做事项

1. 不修改 Go 代码
2. 不新增 controller
3. 不新增 service
4. 不新增 model
5. 不新增 route
6. 不接入 Pipeline
7. 不修改业务 Agent
8. 不修改 Prompt 文件
9. 不修改 Schema 文件
10. 不修改前端
11. 不修改数据库
12. 不执行真实 Hunyuan 调用
13. 不请求腾讯招聘接口
14. 不进入 4-3.7

## 21. 结论

1. **当前不具备直接接入 Pipeline 的条件**：缺少 ResumeOptimizationAgent / Mock Agent / Service / Handler / Model / Route 共 6 项核心业务代码
2. **当前只有版本对比能力，不是简历优化能力**：两者是不同能力，不能混用
3. **Prompt + Schema 资产完备**：Pipeline 和 Repair 已预留支持
4. **4-3.7 应先补业务代码再接 Pipeline**：推荐选项 D，分步实现
5. **接入时必须吸收 4-3.5.B ~ 4-3.5.G 的战略边界**：不做无限优化、使用示例参考、保留核实提醒、围绕 JD 证据链、区分 Top 3 和长期提升项
6. **如果当前 Schema 不支持新战略字段，标记为后续改造**：不因字段缺失阻塞当前接入
