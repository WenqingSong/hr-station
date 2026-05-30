# 岗位匹配Agent

> 最后更新：2026-05-30（阶段 4-3.5.B 战略差异化与杀手锏确认）

## 1. 当前实现状态

阶段 3-7 已完成，岗位匹配 Mock Agent 已实现（`server/internal/agent/match_mock.go`）。阶段 4-3.2 已完成 MatchAnalysisAgent 接入 Pipeline（`server/internal/agent/match_agent.go`），MatchService 支持 LLM Pipeline 优先 + Mock 回退。LLM_ENABLE=true 时优先调用 Pipeline，fallback_required 或映射失败时自动回退 Mock Agent；LLM_ENABLE=false 时直接走 Mock Agent。

## 2. 阶段 3 Mock Agent 目标

1. 生成匹配度评分（0-100）
2. 构建 JD-简历证据链（逐条对照岗位要求与简历经历）
3. 生成初筛判断
4. 检测风险点
5. 给出 P0/P1 优化建议
6. 给出下一步建议
7. 包含真实性提示
8. 不自动生成优化版简历

## 3. 输入数据

| 字段 | 类型 | 说明 |
|------|------|------|
| resume_version | object | 当前简历版本数据 |
| job | object | 目标岗位 JD |

## 4. 输出结构

```json
{
  "job_id": 1,
  "job_title": "Go 后端开发工程师",
  "resume_version_id": 2,
  "resume_version_no": 2,
  "match_score": 78,
  "screening_result": "likely",
  "evidence_chain": [...],
  "match_highlights": [...],
  "risk_points": [...],
  "optimization_suggestions": [...],
  "next_steps": [...],
  "authenticity_notice": "以下建议仅用于表达优化，请只补充真实经历，不要虚构项目或能力。"
}
```

## 5. Mock 匹配规则

| 规则 | 加分 | 说明 |
|------|------|------|
| target_city = job.city | +15 | 城市匹配 |
| target_role 与 category/title 相关 | +25/+8 | 方向匹配 |
| 技能交集（每项） | +5 | 最大 35 |
| 高技能重叠奖励 | +5 | 交集 > 5 |
| 关键技术同时匹配 | +3/项 | 最大 15 |
| 有实习/工作经历 | +5 | |
| 有项目经历 | +5 | |

### 证据链判断

| 判断 | 值 | 说明 |
|------|---|------|
| 强匹配 | `strong_match` | 简历中有明确对应内容 |
| 部分匹配 | `partial_match` | 简历中有部分对应内容 |
| 缺失 | `missing` | 简历中没有对应内容 |

### 初筛判断

| 分数 | 判断 | 说明 |
|------|------|------|
| 85-100 | `highly_likely` | 匹配度高 |
| 70-84 | `likely` | 匹配度较高 |
| 55-69 | `possible` | 匹配度一般 |
| 0-54 | `risky` | 匹配度低 |

### 优化建议优先级

| 优先级 | 说明 |
|--------|------|
| P0 | 必须优化，否则初筛风险高 |
| P1 | 建议优化，可提升匹配度 |

## 6. JD-简历证据链设计

证据链是本 Agent 的核心输出，逐条对照岗位要求与简历经历：

1. 对岗位的每个技能关键词，检查简历中是否有对应内容
2. 对城市要求，检查简历 target_city 是否匹配
3. 对方向要求，检查简历 target_role 是否匹配
4. 每条证据包含：requirement（岗位要求）、resume_point（简历证据）、judgement（判断）

## 7. P0/P1 优化建议设计

- **P0**：基于缺失的关键技能或岗位明确要求但简历缺失的内容
- **P1**：基于可提升的量化指标或工作地点不一致等问题
- 每条建议包含：priority、title、reason、action
- action 是可操作的具体建议，不是模糊的"提升自己"

## 8. 真实性提示

每条匹配分析结果都包含 authenticity_notice 字段，明确告知用户：
- 建议仅用于表达优化
- 请只补充真实经历
- 不要虚构项目或能力

## 9. 当前不调用真实模型

1. 不读取 LLM_API_KEY
2. 不调用外部 API
3. 不使用真实 Prompt
4. 结果基于简单规则打分
5. 阶段 4 才替换为真实 LLM

## 10. 4-1.3 适配方向

4-1.3.6 已完成，匹配分析已使用腾讯岗位 description/requirements：

1. 匹配分析已使用腾讯岗位完整 JD（description / requirements）
2. data_quality=full_jd 的岗位支持深度证据链，匹配结果包含 analysis_basis
3. summary_only 岗位返回降级提示，建议补充完整 JD 后再分析
4. 匹配服务仍只读取本地 jobs 表
5. 适配方案见 `docs/02-后端设计/阶段4-1.3腾讯招聘信息适配方案.md`
6. generated_by 仍为 mock_agent

## 11. Prompt Template 登记状态

- 匹配分析 Agent 后续将使用 `server/prompts/match_analysis_v1.md`
- 当前该 Prompt Template 已登记（4-1.4.1A 人工定稿，4-1.4.1B CodeBuddy 登记）
- 模板包含 HR 初筛 + 技术面试官双视角、Offer鹅语气、证据链、P0/P1/P2 优化建议、面试追问预测和真实性提醒
- 当前仍使用 Mock Agent
- **4-1.4.2 已完成：匹配分析 Prompt 首轮输出验收**
  - 验收结论：partial_pass
  - 模型输出 JSON 合法，17 个顶层字段完整，枚举值全部合法
  - 发现问题：rewrite_example 中少量具体数字超出简历事实，证据链缺少 missing 判断
  - 问题不阻塞进入 4-1.4.3A JSON Schema 定稿
  - 验收报告见 `docs/05-Agent设计/匹配分析Prompt首轮输出验收报告.md`
- **4-1.4.3A/B 已完成：匹配分析 JSON Schema 定稿与登记**
  - Schema 路径：`server/prompts/schema/match_analysis_schema.json`
  - 当前匹配分析 Agent 后续将使用：
    - Prompt Template：`server/prompts/match_analysis_v1.md`
    - Output Schema：`server/prompts/schema/match_analysis_schema.json`
  - 当前 Prompt 和 Schema 均已登记
  - 当前仍使用 Mock Agent
  - 当前仍未接入 LLM Client
  - 后续推荐 Prompt 将继续采用人工定稿 + CodeBuddy 登记模式

## 12. 后续真实 Agent 设计方向

1. 基于真实 Prompt 的深度分析
2. 更精准的证据链构建
3. 更个性化的优化建议
4. 支持 NLP 理解岗位描述和简历内容
5. 异步分析支持

## 13. 阶段 4-3.0 接入方案

4-3.0 已完成 MatchAnalysisAgent 接入 LLM 方案设计（docs only），详细方案见 `docs/05-Agent设计/阶段4-3业务Agent接入LLM方案.md`。

关键接入设计：

1. **Service 层注入 pipelineRunner**：MatchService 新增 pipelineRunner 字段，构造函数注入
2. **PromptValues 组装**：Service 从数据库读取简历和岗位后，组装 `RESUME_CONTEXT_JSON`、`JOB_CONTEXT_JSON`、`BUSINESS_CONTEXT` 三个占位符
3. **Pipeline 调用**：AgentName=match_analysis，PromptName=match_analysis，SchemaName=match_analysis
4. **RunResult 处理**：success/repaired → 将 RunResult.Parsed 映射为 model.MatchAnalysisResult；fallback_required → 调用 s.mockAgent.Analyze()
5. **字段映射策略**：map[string]any → typed struct，映射失败触发 Mock fallback
6. **generated_by 规则**：success/repaired → llm_agent，fallback_required → mock_agent
7. **4-3.2 已实现本接入方案**

当前 MatchAnalysisAgent 已接入 Pipeline，支持 LLM 优先 + Mock 回退。

4-3.3 已完成匹配分析接口稳定性验收。MatchAnalysisAgent 的 LLM Pipeline 优先 + Mock fallback 链路已通过联调。6 大验收场景全部通过：LLM success、repaired、fallback_required、LLM_ENABLE=false、配置异常、请求参数异常。

## 阶段 4-3.5.B 战略差异化确认

4-3.5.B 已完成战略差异化确认。后续匹配分析输出需要围绕 JD 证据链展开：

1. 岗位要求 → 简历证据 → 匹配判断 → 风险差距 → 投递建议
2. 需要支持 `apply_readiness` 投递准备度判断（ready_to_apply / quick_fix_before_apply / not_ready_for_this_role / insufficient_information）
3. 需要支持 `optimization_stop_signal` 优化停止信号（stop_and_apply / fix_top_3_then_apply / needs_more_evidence / target_role_mismatch）
4. 每条建议需要附带 `self_check_question`、`learn_more_query`、`verification_notice`
5. 本阶段不修改 Prompt / Schema / 代码，只登记方向

详见 `docs/05-Agent设计/战略差异化与杀手锏确认.md`

## 阶段 4-3.5.C 投递边界与优化停止条件设计

4-3.5.C 已完成投递边界与优化停止条件设计。匹配分析后续应支持投递准备度判断：

1. apply_readiness 枚举：ready_to_apply / quick_fix_before_apply / not_ready_for_this_role / insufficient_information
2. optimization_stop_signal 枚举：stop_and_apply / fix_top_3_then_apply / needs_more_evidence / target_role_mismatch
3. apply_readiness=quick_fix_before_apply 时输出 Top 3 快速优化项（top_fix_items）
4. 长期提升项（long_term_improvement）与快速优化项分开展示
5. 每条建议附带 self_check_question / learn_more_query / verification_notice
6. 优化停止信号避免用户陷入无限优化循环

本阶段不修改 Prompt / Schema / 代码，只登记方向。

详见 `docs/05-Agent设计/投递边界与优化停止条件设计.md`

## 阶段 4-3.5.D JD 证据链匹配与 START 简历优化规则

4-3.5.D 已完成 JD 证据链匹配与 START 简历优化规则设计。匹配分析后续应围绕 JD 证据链输出：

1. jd_evidence_chain：JD 要求 → 简历证据 → 证据强度 → 匹配判断 → 差距风险 → 优化动作
2. evidence_strength：strong / medium / weak / missing
3. match_judgement：matched / partially_matched / missing / risky
4. screening_risk：keyword_missing / responsibility_unclear / result_missing 等九类风险
5. 每条建议附带 self_check_question / learn_more_query / verification_notice
6. 使用"示例参考"而非"建议改为"

本阶段不修改 Prompt / Schema / 代码，只登记方向。

详见 `docs/05-Agent设计/JD证据链匹配与START简历优化规则.md`

## 阶段 4-3.5.E Agent 输出 Quality Gate 设计

4-3.5.E 已完成 Agent 输出 Quality Gate 设计。后续匹配分析输出需要通过 Quality Gate 检查：

1. 是否有 JD 证据链（jd_grounding_check）
2. 是否有 apply_readiness（apply_boundary_check）
3. 是否有 optimization_stop_signal（optimization_stop_check）
4. 是否编造简历证据（fabrication_check）
5. 是否避免泛泛匹配（jd_grounding_check）
6. apply_readiness 与 optimization_stop_signal 是否一致（consistency_check）
7. 是否包含核实提醒（verification_check）
8. 是否存在不当表达（safety_tone_check）

fabrication_check 一旦失败应直接 fallback_required。revise_required 时最多重试一次。

本阶段不修改 Prompt / Schema / 代码，只登记方向。

详见 `docs/05-Agent设计/Agent输出QualityGate设计.md`

## 阶段 4-3.5.F 学生求职极端情况处理边界

4-3.5.F 已完成学生求职极端情况处理边界设计。后续匹配分析需识别 candidate_case_type：

1. 信息不足时不强行给高分，返回 apply_readiness=insufficient_information
2. 弱简历时给出最小可行动作，不否定学生
3. 岗位不匹配时表达为"当前简历 × 当前岗位 JD 不匹配"，不做人格判断
4. 海外/地点问题只处理投递条件与岗位适配
5. 优秀候选人只提示方向，不做长期职业规划
6. 真实性风险时保守处理，建议只保留实际参与的内容
7. 过度优化时明确告知可以投递
8. 造假请求必须拒绝，提供基于真实经历的替代建议

详见 `docs/05-Agent设计/学生求职极端情况处理边界.md`

## 阶段 4-3.5.G 灵活在线简历 Profile 与纯文字 Agent 上下文设计

4-3.5.G 已完成灵活在线简历 Profile 与纯文字 Agent 上下文设计。后续匹配分析输入统一为 ResumeContextPayload：

1. **ResumeContextText** 替代旧的 `json.Marshal(resumeVersion)`，结构化纯文字格式，Agent 可稳定消费
2. **profile_summary** 提供简历概况信号（section_count / has_project / has_internship 等），快速判断 candidate_case_type
3. **privacy_flags** 确认隐私字段已排除（姓名/手机/邮箱不发给 LLM）
4. **CURRENT_TARGET_SECTION**（仅简历优化 Agent）：针对具体 Section 给出优化建议

PromptValues 变化：`RESUME_CONTEXT_JSON` → `RESUME_CONTEXT_TEXT` + `PROFILE_SUMMARY_JSON` + `PRIVACY_FLAGS_JSON`

详见 `docs/05-Agent设计/灵活在线简历Profile与纯文字Agent上下文设计.md`

## 阶段 4-3.5.H 战略转型阶段汇总与工程评审

4-3.5.H 评审结论：MatchAnalysisAgent 已完成接入与验收，后续再按战略转型方向优化 Prompt / Schema，不在当前阶段立即改动。

详见 `docs/05-Agent设计/阶段4-3.5战略转型汇总与工程评审.md`

本阶段不修改 Prompt / Schema / 代码，只登记方向。

详见 `docs/05-Agent设计/学生求职极端情况处理边界.md`

## 14. 阶段 4-3.2 接入实现

4-3.2 已完成 MatchAnalysisAgent 接入 Pipeline，核心实现：

### 代码变更

| 文件 | 变更 |
|------|------|
| `server/internal/agent/match_agent.go` | 新增：MatchAnalyzer 接口、PipelineRunner 接口、MatchAnalysisAgent、MatchFallbackError、PromptValues 组装、字段映射 |
| `server/internal/agent/match_mock.go` | 修改：Analyze 新增 ctx 参数以实现 MatchAnalyzer 接口 |
| `server/internal/model/match.go` | 修改：新增 MatchLevel/MatchLabel/OfferGoSummary/GapPoints/InterviewQuestions/GeneratedBy/AgentStatus/RepairApplied/FallbackReason/ErrorCode 字段 |
| `server/internal/service/match_service.go` | 重写：MatchService 新增 pipelineRunner + llmEnabled，Analyze() 支持 LLM 优先 + Mock 回退 |
| `server/internal/config/config.go` | 修改：LLMConfig 从 3 字段扩展到 14 字段 |
| `server/cmd/server/main.go` | 修改：LLM 基础设施初始化，MatchService 依赖注入 |
| `server/internal/agent/match_agent_test.go` | 新增：28 项单元测试（fake Pipeline，不调用真实 LLM） |

### MatchAnalyzer 接口

```go
type MatchAnalyzer interface {
    Analyze(ctx context.Context, resume *model.ResumeVersion, job *model.Job) (*model.MatchAnalysisResult, error)
}
```

MatchMockAgent 和 MatchAnalysisAgent 均实现此接口。

### PromptValues 组装

三个占位符对应 match_analysis_v1.md：
- `RESUME_CONTEXT_JSON`：简历安全子集（version_no, target_role, target_city, skills, education, experience, project, self_evaluation）
- `JOB_CONTEXT_JSON`：岗位信息（id, title, department, city, category, description, requirements, data_quality, skill_keywords）
- `BUSINESS_CONTEXT`：业务上下文注入（HR 初筛 + 技术面试官双视角）

### Pipeline 调用

```go
runReq := pipeline.RunRequest{
    AgentName:        "MatchAnalysisAgent",
    PromptName:       prompt.PromptMatchAnalysis,
    SchemaName:       schema.SchemaMatchAnalysis,
    TargetOutputType: pipeline.TargetMatchAnalysis,
    PromptValues:     promptValues,
}
```

### RunResult 处理

| AgentStatus | 处理 |
|-------------|------|
| success | mapParsedToMatchResult → MatchAnalysisResult |
| repaired | mapParsedToMatchResult → MatchAnalysisResult（RepairApplied=true） |
| fallback_required | 返回 MatchFallbackError → Service 回退 Mock Agent |

### 字段映射策略

LLM Schema 输出与 Mock Model 存在结构差异，通过映射函数处理：
- `evidence_chain.judgement`：matched→strong_match, partial→partial_match, missing→missing
- `screening_result`：likely→likely, uncertain→possible, risky→risky
- `matched_points`：提取 point 字符串 → match_highlights
- `gap_points`：提取 gap 字符串（LLM 新增字段）
- `risk_points`：提取 risk 字符串，兼容纯字符串
- `resume_optimization.p0/p1/p2`：扁平化为 MatchSuggestion 列表
- `interview_questions`：提取 question 字符串（LLM 新增字段）
- `next_actions`：提取 action 字符串，兼容纯字符串
- 映射失败触发 MatchFallbackError → Mock 回退

### 开关控制

- `LLM_ENABLE=false`：MatchService 直接走 Mock Agent，不初始化 Pipeline Runner
- `LLM_ENABLE=true`：MatchService 优先走 Pipeline，fallback_required 时回退 Mock
