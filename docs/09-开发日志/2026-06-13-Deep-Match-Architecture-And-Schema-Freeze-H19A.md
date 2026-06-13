# Deep-Match-Architecture-And-Schema-Freeze-H19A 完成报告

> 执行时间：2026-06-13 15:48  
> 类型：设计冻结（不修改代码）  
> 状态：✅ 冻结完成，H19-B 可执行

---

## 1. 总体结论

| 项目 | 结果 |
|------|------|
| 是否完成设计冻结 | ✅ 完成 |
| 是否修改业务代码 | ❌ 未修改 |
| 是否执行 Git | ❌ 未执行 |
| **当前 Deep 最大问题** | fit_points/evidence_chain 为 null；risk 为 bare strings 无结构化；缺 JD 解构/证据映射/改写策略/面试预测 |
| 是否冻结 Agent 架构 | ✅ 8 层 Agent（L0-L8） |
| 是否冻结输出 Schema | ✅ 12 模块化 Deep Match Report |
| 是否冻结分数字义 | ✅ 与 Fast 统一 5 级语义 |
| 是否冻结前端展示模块 | ✅ 10 模块卡片 |
| 是否可以进入 H19-B | ✅ 可以 |

---

## 2. 当前 Deep 审计结果

| 项目 | 当前实现 | 问题 |
|------|----------|------|
| Deep 入口 | `matchWith3Agents()` → `agent.RunDeepMatch3Agents()` | 3-agent 架构 |
| 是否 multi-agent | ✅ 3-agent（Context+Decision+Verifier） | Agent 1 基于 TagProfile 非 SEP |
| 输入上下文 | `ResumeTagProfile` + `JobTagProfile` | ❌ 非 SEP Profile |
| 是否使用 Resume SEP | ❌ 否 | 用 TagEvidenceService |
| 是否使用 Job SEP | ❌ 否 | 用 TagEvidenceService |
| 是否使用 Fast Result | ❌ 否 | Deep 独立运行 |
| prompt 质量 | `matchAnalysisBusinessContext` 基础 prompt | 缺少 JD 解构指令 |
| schema | `SchemaMatchAnalysis` | 字段有限 |
| JSON repair | ✅ pipeline 内置 | - |
| fallback | ✅ 3-agent → single agent → mock | - |
| **fit_points** | **null** | ❌ P0 |
| **evidence_chain** | **null** | ❌ P0 |
| risk_points | 3 bare strings | ❌ 非结构化 |
| generated_by | `llm_agent` | 需改为 `deep_match_agent` |

---

## 3. 当前 Deep API 实测

| 项目 | 结果 |
|------|------|
| HTTP/code | 200/0 |
| 耗时 | 34.8s |
| score | 35 (low) |
| generated_by | llm_agent |
| agent_status | success |
| fit_points | **null** ❌ |
| evidence_chain | **null** ❌ |
| risk_points | 3 bare strings ❌ |
| JD 解构 | ❌ 无 |
| 证据映射 | ❌ 无 |
| 简历改写策略 | ❌ 无 |
| 面试追问 | ❌ 无 |

---

## 4. 冻结 Agent 架构（8 层）

| 层级 | Agent | 作用 |
|------|-------|------|
| L0 | Input Gate Agent | 检查输入完整性（Resume SEP / Job SEP / Fast Result） |
| L1 | JD Decomposition Agent | 拆解 JD 为 must_have / should_have / nice_to_have / hidden |
| L2 | Resume Evidence Audit Agent | 审计简历证据（从 SEP evidence_text 提取） |
| L3 | Requirement Alignment Agent | JD 要求 × 简历证据逐条对齐（coverage + evidence mapping） |
| L4 | Risk Diagnosis Agent | 识别投递风险（方向/技能/证据/学历/经历） |
| L5 | Application Strategy Agent | 投递策略（时机/竞争定位/推荐度） |
| L6 | Resume Rewrite Strategy Agent | 针对该岗位的简历改写方向 |
| L7 | Interview Question Prediction Agent | 预测面试追问（基于 risk + gap） |
| L8 | Global Calibration Agent | 统一分数/风险/建议，防前后矛盾 |

**H19-B 实现策略**：先以单 LLM 调用按 8 层结构输出，不强制多次 LLM 调用。agent_trace 记录逻辑分层。

---

## 5. 冻结输出 Schema

```json
{
  "generation_tier": "quality",
  "generated_by": "deep_match_agent",
  "agent_status": "match_verified",
  "match_score": 0,
  "match_level": "",
  "application_decision": "",
  "decision_label": "",
  "one_sentence_conclusion": "",
  "score_breakdown": {},
  "job_understanding": {},
  "requirement_coverage": [],
  "resume_evidence": [],
  "core_match_points": [],
  "key_risks": [],
  "resume_rewrite_strategy": {},
  "application_strategy": {},
  "interview_risk_questions": [],
  "agent_trace": []
}
```

**硬要求**：所有数组 `[]`、generated_by=`deep_match_agent`、agent_status=`match_verified`、禁止 null、禁止 raw tag code。

---

## 6. 冻结关键字段结构

| 字段 | 结构 | 要求 |
|------|------|------|
| `job_understanding` | positioning + core_responsibilities + must/important/bonus/hidden | 解释 JD 真正在看什么，非复制 |
| `requirement_coverage` | 7 字段（沿用 H18-B） | ≥5 条，覆盖 must/should/nice |
| `resume_evidence` | 7 字段 0-5 级（沿用 H18-B） | ≥3 条或如实标注证据不足 |
| `core_match_points` | point_name + strength(strong/partial/transferable/weak) + evidence | 具体匹配点 |
| `key_risks` | 6 字段（比 Fast 多 interview_impact） | ≥2 条，深度解释 |
| `resume_rewrite_strategy` | skills_reorder/project/internship/keywords/weaken | 具体到模块和方向 |
| `application_strategy` | timing + delivery + competing | 投递策略 |
| `interview_risk_questions` | question + why_asked + answer_direction | ≥5 条，与风险关联 |
| `agent_trace` | stage/agent/status/summary | 9 阶段枚举 |

---

## 7. Deep 与 Fast 关系

| 维度 | Fast | Deep |
|------|------|------|
| 定位 | 秒级投递判断 | 深度求职顾问报告 |
| 输出 | 6 模块 | 10 模块 |
| JD 解构 | ❌ | ✅ `job_understanding` |
| 简历改写 | ❌ | ✅ `resume_rewrite_strategy` |
| 面试预测 | ❌ | ✅ `interview_risk_questions` |
| 分数字义 | 统一 5 级 | 统一 5 级（Deep 可更严格但须解释） |

**分差 >15 规则**：Deep 必须输出 `score_delta_explanation` 说明原因。

---

## 8. H19-B 执行边界

| 项目 | 允许？ |
|------|--------|
| 改 Deep Prompt | ✅ |
| 改 Deep Schema | ✅ |
| 接 SEP Profile | ✅ |
| 接 Fast Result | ✅ |
| 改 Fast Engine | ❌ |
| 改数据库 | ❌ |
| 改前端 | ❌ |

---

## 9. H19-C 执行边界

| 项目 | 允许？ |
|------|--------|
| 改 Deep 前端展示 | ✅ |
| 新增 Deep 报告组件 | ✅ |
| 改 Fast 算法 | ❌ |
| 改后端 | ❌ |

---

## 10. 构建检查

| 项目 | 结果 |
|------|------|
| vite build | ✅ 2.29s |
| go build | ✅ |
| go vet | ✅ |

---

## 11. docs

- 新增：`docs/09-开发日志/2026-06-13-Deep-Match-Architecture-And-Schema-Freeze-H19A.md`
- 更新：`docs/变更记录.md`、`docs/下一步任务.md`

---

## 12. 给用户的结论

1. **当前 Deep 还不够优秀**：fit_points=null, evidence_chain=null, risk 为 bare strings, 无 JD 解构/证据映射/改写策略/面试预测。只是 Fast 的长文本版。

2. **新 Deep 架构 8 层 Agent**：从 Input Gate 到 Global Calibration，每层有明确的责任和输出字段，系统化覆盖 JD 解构→证据审计→要求对齐→风险诊断→策略→改写→面试全链路。

3. **关键改进**：(a) 接入 SEP Profile 替代 TagProfile → 真正语义证据；(b) 接入 Fast Result 作为基准参照 → 分数可解释；(c) 结构化 null→[] 风险→6 字段对象 → 用户能理解每个风险。

4. **为什么不能只改 Prompt**：Prompt 只能提升文字质量，无法解决 null fields、缺字段、SEP 不接入、Fast 不参照等结构性问题。需要 Schema + Pipeline + Context 三重改造。

5. **H19-B 应实现**：重写 Prompt/Schema → 接入 SEP Resume+Job Profile + Fast Result → 单 LLM 调用输出 12 模块 → JSON repair → fallback → `deep_match_agent`。
