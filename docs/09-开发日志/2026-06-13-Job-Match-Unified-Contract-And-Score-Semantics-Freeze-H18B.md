# Job-Match-Unified-Contract-And-Score-Semantics-Freeze-H18B 完成报告

> 执行时间：2026-06-13 14:52  
> 类型：设计冻结（不修改代码）  
> 状态：✅ 冻结完成，H18-C 可执行

---

## 1. 总体结论

| 项目 | 结果 |
|------|------|
| 是否完成设计冻结 | ✅ 完成 |
| 是否修改业务代码 | ❌ 未修改 |
| 是否执行 Git | ❌ 未执行 |
| 是否统一 Fast/Deep 分数字义 | ✅ 5 级区间 + 上限规则 |
| 是否冻结投递建议枚举 | ✅ 5 值枚举 + 决策规则 |
| 是否冻结风险等级 | ✅ 结构化 risk + high/medium/low |
| 是否冻结证据链结构 | ✅ 7 字段结构化 evidence |
| 是否冻结要求覆盖结构 | ✅ requirement_coverage 7 字段 |
| 是否冻结 Fast 输出契约 | ✅ fast_sep_match_engine |
| 是否冻结 Deep 输出契约 | ✅ deep_match_agent |
| 是否可以进入 H18-C | ✅ 可以 |

---

## 2. 当前契约问题复查

| 问题 | 当前表现（H18-A 实测） | 冻结后的解决方式 |
|------|----------------------|-----------------|
| Fast 分数虚高 | `min(98, 60 + tagScore*0.8)`，base 60 无意义 | 统一 0-100 分 5 级语义 + hard gate 上限 |
| Fast/Deep 分数冲突 | 同一岗位 Fast=60 Deep=50，无解释 | 统一语义；差距 >15 分时 Deep 必须解释原因 |
| highlights null | 无匹配标签时 `match_highlights: null` | 所有数组默认 `[]`，禁止 null |
| raw tag code 泄露 | `risk_points: ["project_exp", "docker_k8s"]` | 结构化 risk+evidence+coverage，全部 humanized |
| evidence_chain null | Deep 实测 evidence_chain=null | 结构化 `resume_evidence` 7 字段，非空 |
| fit_points null | Deep 实测 fit_points=null | 并入 `core_match_points`，非空 |
| **SEP 未接入** | 匹配用 TagEvidenceService，推荐用 SemanticProfileGate | H18-C 接入 SEP Profile |

---

## 3. 统一分数字义

| 分数区间 | 匹配等级 | 投递含义 | 限制规则 |
|------|------|------|------|
| 90-100 | 高度匹配 | 方向强一致 + 必备覆盖充分 + 证据充实 + 风险可控。建议优先投递 | hard gate 全部通过 |
| 80-89 | 较高匹配 | 方向一致或相邻 + 主要覆盖较好 + 有证据但少量短板。建议投递前微调 | 无 hard gate blocked |
| 70-79 | 中等匹配 | 有相关性但核心或证据不足。可投，建议针对 JD 优化后投递 | 至少有方向相关性 |
| 60-69 | 弱匹配 | 方向或核心能力存在明显短板。不建议优先 | 可能存在 hard gate warning |
| 0-59 | 不建议投递 | 方向/必备/证据明显不匹配，投递命中率极低 | hard gate blocked |

**分数上限规则（Hard Gate Cap）**：

| 条件 | 最高分 |
|------|--------|
| 岗位方向明显不一致 | 65 |
| 缺少核心必备技能（must_have missing） | 75 |
| 只有技能列表，无项目/实习证据 | 72 |
| 无相关项目证据 | 70 |
| 无实习但岗位强实习 | 78 |
| 证据充分但表达弱 | 85 |
| 方向强一致 + 实习证据强 + 风险少 | 可达 90+ |
| hard gate blocked | 60 |

---

## 4. 投递建议枚举

| code | 展示文案 | 分数触发区间 | 允许风险覆盖 |
|------|----------|-------------|-------------|
| `priority_apply` | 优先投递 | 90+ | 无 medium/high 风险时 |
| `apply_directly` | 可以直接投递 | 80-89 | 无 high 风险时 |
| `revise_before_apply` | 修改后投递 | 70-79，或 80+ 但有 medium/high 风险 | 默认中高位默认值 |
| `backup_apply` | 作为备选投递 | 60-69 | 低优先级 |
| `not_recommended` | 暂不建议投递 | 0-59，或 hard gate blocked | 不建议当前投递 |

决策规则：分数为主，风险可降级。80 分但有 high risk → `revise_before_apply`。

---

## 5. 风险结构

| 字段 | 类型 | 说明 |
|------|------|------|
| `risk_name` | string | 风险名称，如"高并发表达不足" |
| `risk_level` | enum | `high` / `medium` / `low` |
| `why_it_matters` | string | 为什么这个风险会影响筛选 |
| `evidence_gap` | string | 简历中缺少哪些具体证据 |
| `fix_suggestion` | string | 如何修复（可执行建议） |

风险等级说明：
- `high`：直接影响筛选通过率
- `medium`：投递前需重点补强
- `low`：简历表达优化项

禁止：raw tag code、null、纯关键词、泛泛建议。

---

## 6. 证据结构

| 字段 | 类型 | 说明 |
|------|------|------|
| `evidence_name` | string | 证据名称，如"Go 后端工程实践" |
| `evidence_level` | int 0-5 | 证据强度 |
| `source_type` | enum | `internship` / `project` / `course` / `skill_list` / `self_claimed` |
| `source_title` | string | 来源标题，如"企业课程学习平台" |
| `evidence_text` | string | 证据原文引用 |
| `supports_requirement` | string | 支撑了哪个岗位要求 |
| `confidence` | enum | `high` / `medium` / `low` |

证据等级：
- 5：实习/生产场景 + 明确结果
- 4：项目/实习 + 具体职责
- 3：项目中出现技术使用
- 2：技能列表中出现
- 1：泛泛描述
- 0：无证据

规则：strong match 至少需要 Level 3 证据。Level 1/2 不能支撑"高度匹配"。

---

## 7. 要求覆盖结构

| 字段 | 类型 | 说明 |
|------|------|------|
| `requirement_name` | string | 要求名称 |
| `requirement_type` | enum | `must_have` / `should_have` / `nice_to_have` / `hidden_expectation` |
| `coverage_status` | enum | `strong` / `partial` / `transferable` / `weak` / `missing` / `unknown` |
| `coverage_score` | int 0-100 | 该要求项覆盖分 |
| `matched_evidence` | string[] | 匹配的证据引用 |
| `gap` | string | 缺口描述（有缺口时填写） |
| `suggestion` | string | 补强建议 |

规则：
- `must_have` 缺失显著影响总分
- `nice_to_have` 命中不能掩盖 `must_have` 缺失
- `hidden_expectation` 影响风险但不单独决定分数
- `unknown` ≠ `missing`

---

## 8. Fast Match Result 契约

```json
{
  "generation_tier": "fast",
  "generated_by": "fast_sep_match_engine",
  "match_score": 82,
  "match_level": "较高匹配",
  "application_decision": "revise_before_apply",
  "decision_label": "修改后投递",
  "one_sentence_conclusion": "方向较一致，但需补强稳定性表达。",
  "score_breakdown": {
    "direction_fit": 82,
    "must_have_coverage": 78,
    "evidence_strength": 80,
    "experience_relevance": 75,
    "risk_control": 70
  },
  "hard_gate": {
    "status": "weak_pass",
    "score_cap": 85,
    "warnings": []
  },
  "core_advantages": [],
  "main_gaps": [],
  "risk_points": [],
  "requirement_coverage": [],
  "resume_evidence": [],
  "action_suggestions": []
}
```

| 字段 | 必填 | 默认值 | 说明 |
|------|------|--------|------|
| `generation_tier` | ✅ | `"fast"` | |
| `generated_by` | ✅ | `"fast_sep_match_engine"` | 从 `fast_tag_engine` 改名 |
| `match_score` | ✅ | - | 0-100 |
| `match_level` | ✅ | - | 5 级文案 |
| `application_decision` | ✅ | - | 5 值枚举 |
| `decision_label` | ✅ | - | 中文展示 |
| `one_sentence_conclusion` | ✅ | - | 具体，不可泛 |
| `score_breakdown` | ✅ | - | 5 维度 |
| `hard_gate` | ✅ | - | status + cap + warnings |
| `core_advantages` | ✅ | `[]` | |
| `main_gaps` | ✅ | `[]` | |
| `risk_points` | ✅ | `[]` | 结构化 risk 对象数组 |
| `requirement_coverage` | ✅ | `[]` | 结构化 coverage 对象数组 |
| `resume_evidence` | ✅ | `[]` | 结构化 evidence 对象数组 |
| `action_suggestions` | ✅ | `[]` | 可执行建议 |

---

## 9. Deep Match Report 契约

```json
{
  "generation_tier": "quality",
  "generated_by": "deep_match_agent",
  "agent_status": "match_verified",
  "match_score": 82,
  "match_level": "较高匹配",
  "application_decision": "revise_before_apply",
  "decision_label": "修改后投递",
  "one_sentence_conclusion": "",
  "job_understanding": {
    "job_positioning": "",
    "core_expectations": [],
    "hidden_expectations": []
  },
  "requirement_coverage": [],
  "resume_evidence": [],
  "core_match_points": [],
  "key_risks": [],
  "resume_rewrite_strategy": {
    "skills_reorder": [],
    "project_rewrite_points": [],
    "internship_rewrite_points": [],
    "keywords_to_add": [],
    "content_to_weaken": []
  },
  "interview_risk_questions": [],
  "application_strategy": {
    "best_timing": "",
    "delivery_strategy": "",
    "competing_positioning": ""
  },
  "agent_trace": []
}
```

Deep 比 Fast 多出的能力：
1. `job_understanding` — JD 深度解构
2. `resume_rewrite_strategy` — 简历改写策略
3. `interview_risk_questions` — 面试追问预测
4. `application_strategy` — 投递策略（时机/竞争定位）
5. `agent_trace` — 多 Agent 分析履历

---

## 10. Fast / Deep 关系

| 维度 | Fast | Deep |
|------|------|------|
| 定位 | 秒级投递判断 | 深度求职顾问报告 |
| 耗时 | <3s | 30-120s |
| 核心输出 | score + decision + advantage/gap + action | 全 12 模块报告 |
| JD 解构 | ❌ 不做 | ✅ `job_understanding` |
| 简历改写 | ❌ 不做 | ✅ `resume_rewrite_strategy` |
| 面试预测 | ❌ 不做 | ✅ `interview_risk_questions` |
| 相同字段 | score / level / decision / risk / evidence / coverage | 比 Fast 更详细 |

分数冲突规则：如 Fast 和 Deep 差距 >15 分，Deep 必须解释原因。

---

## 11. 前端展示模块冻结

### Fast 页面模块（5 卡）
1. 匹配结论卡（score + level + one_sentence_conclusion）
2. 投递决策卡（decision + label + 决策原因）
3. 核心优势卡（core_advantages + evidence）
4. 主要短板/风险卡（main_gaps + risk_points）
5. 下一步建议卡（action_suggestions）

### Deep 页面模块（9 卡）
1. 顶部结论卡（score + level + one_sentence_conclusion）
2. 投递决策卡（decision 比 Fast 更详细）
3. JD 解构卡（job_understanding）
4. 要求覆盖卡（requirement_coverage 表格）
5. 证据映射卡（resume_evidence 结构化列表）
6. 风险诊断卡（key_risks 结构化列表）
7. 简历优化方向卡（resume_rewrite_strategy）
8. 面试追问预测卡（interview_risk_questions）
9. Agent 分析履历卡（agent_trace 轻量展示）

---

## 12. 数据库策略

- 优先复用 `match_analyses.result_json`
- H18-C/H18-D：Fast 新结构存入 `result_json`，不改表
- H19-B：Deep 新结构同策略
- 后续仅在查询/索引强依赖时考虑加列

---

## 13. 后续执行计划

| 阶段 | 目标 | 是否进入 | 优先级 |
|------|------|---------|-------|
| **H18-C** | Fast Match SEP Context 接入（接 SemanticProfileGate 替代 TagEvidenceService） | ✅ 可进入 | P0 |
| **H18-D** | Fast Match Engine 核心算法重构（Hard Gate + Evidence + 新分数） | ⬜ 冻结后进入 | P0 |
| **H18-E** | Fast Match 前端展示重构（5 模块卡布局） | ⬜ | P0 |
| **H18-F** | Fast Match 精准度校准验收 | ⬜ | P0 |
| **H19-A** | Deep Match Schema/Prompt 冻结 | ⬜ | P1 |
| **H19-B** | Deep Match Agent 后端重构 | ⬜ | P1 |
| **H19-C** | Deep Match 前端报告页重构（9 模块） | ⬜ | P1 |
| **H19-D** | Deep Match 精准度校准验收 | ⬜ | P1 |
| **H20-A** | 岗位匹配全链路最终验收 | ⬜ | P1 |

优先级：先 Fast 后 Deep / 先契约后算法 / 先后端精准度后前端展示。

---

## 14. 构建检查

| 项目 | 结果 |
|------|------|
| vite build | ✅ 2.86s |
| go build | ✅ |
| go vet | ✅ |

---

## 15. docs

- 新增：`docs/09-开发日志/2026-06-13-Job-Match-Unified-Contract-And-Score-Semantics-Freeze-H18B.md`
- 更新：`docs/变更记录.md`、`docs/下一步任务.md`

---

## 16. 给用户的结论

1. **必须先冻结契约再写代码**：否则 H18-C/H18-D/H19-B 三阶段各自定义字段，后端、前端、用户看到的都是不同的结构，精准度永远说不清楚。

2. **统一分数字义解决 Fast/Deep 冲突**：5 级 0-100 分 + hard gate 上限规则，确保两者用同一套语言说话。差距 >15 分时 Deep 必须解释。

3. **所有数组默认 `[]` 禁止 null**：彻底解决当前 highlights=null / evidence_chain=null / fit_points=null 的问题。

4. **结构化 risk/evidence/coverage 杜绝 raw tag code 泄露**：每个风险必须有 why_it_matters + evidence_gap + fix_suggestion，每项证据必须有 level + source + confidence。

5. **Fast 和 Deep 关系清晰**：Fast 是秒级投递判断（5 模块），Deep 是完整求职顾问报告（9 模块）。Fast 不假装 Deep，Deep 不重复 Fast。

6. **不改数据库**：新结构存入 `match_analyses.result_json`，不新增列，降低重构风险。

7. **H18-C 可以进入**：接入 SEP SemanticProfileGate 替代 TagEvidenceService，这是精准度提升的最短路径。
