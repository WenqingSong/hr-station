# Deep-5-Module-Hy3-Prompt-Design-JM-R2 完成报告

> 执行时间：2026-06-14 01:30
> 类型：Deep 5 模块 Hy3 Prompt 设计（不修改代码）
> 状态：✅ Prompt 设计完成，JM-R3 可执行

---

## 1. 总体结论

| 项目 | 结果 |
|------|------|
| 是否完成 JM-R2 | ✅ 5 Module + 1 Judge Prompt 设计完成 |
| 是否修改代码 | ❌ 未修改 |
| 是否执行 Git | ❌ 未执行 |
| 是否完成 5 模块 Prompt | ✅ 投递决策/命中点/淘汰风险/改写指令/岗位策略 |
| 是否完成最终裁判 Prompt | ✅ Judge merge + consistency check |
| 是否完成 JSON Schema | ✅ 每模块独立 JSON Schema |
| 是否可以进入 JM-R3 | ✅ 可以 |

---

## 2. 统一输入上下文

```json
{
  "resume": {
    "resume_id": 1,
    "resume_name": "简历名称",
    "raw_text": "完整简历原文",
    "sep_profile": {
      "summary": "语义画像摘要",
      "skills": ["Go", "MySQL", "Redis", "Docker", "Nginx"],
      "projects": ["企业课程学习平台"],
      "internships": ["杭州智航星辰 Go 后端实习"],
      "evidence_items": ["实习中开发了RESTful API并上线"],
      "risks": ["K8s经验缺失"]
    }
  },
  "job": {
    "job_id": 10001,
    "title": "软件开发-后台开发方向",
    "direction": "软件开发",
    "category": "后台开发",
    "city": "深圳",
    "raw_jd": "完整岗位 JD 原文",
    "sep_profile": {
      "job_positioning": "后台开发工程师",
      "must_have": ["Go", "MySQL", "Redis"],
      "should_have": ["微服务", "分布式"],
      "nice_to_have": ["Kubernetes", "消息队列"],
      "hidden_expectations": ["后台系统稳定性建设", "大规模服务经验"]
    }
  },
  "fast_result": {
    "screening_score": 76,
    "screening_level": "中等偏高",
    "apply_decision": "revise_before_apply",
    "one_sentence": "方向较一致，但表达需强化",
    "top_reasons": ["有后端项目经历"],
    "primary_risk": "工程复杂度表达不足"
  },
  "constraints": {
    "goal": "帮助学生提高初筛通过率",
    "do_not_fabricate": true,
    "language": "zh-CN",
    "output_format": "json"
  }
}
```

**要求：** Prompt 必须基于完整简历+JD+SEP；Fast 仅参考；无证据则标记"未体现"；方向不匹配敢说换方向。

---

## 3. 模块 1 Prompt：投递决策

### System Prompt

```
你是一个资深大厂校园招聘 HR 初筛专家。

你的任务：判断一个学生当前是否应该投递这个岗位。

你的判断标准：
1. 简历方向是否和岗位方向一致或可迁移。
2. 必备技能要求是否在简历中有可验证的证据（不只是技能列表）。
3. 是否存在会直接影响初筛的硬伤（方向错位、必备技能全缺、核心经历空白）。

你的立场：
- 你服务学生，但不哄骗学生。
- 你宁可建议"先改简历再投"或"换方向"，也不建议盲目投递。
- 如果简历表达有明显问题但经历本身够用 → 建议修改后投递。
- 如果方向完全不匹配 → 必须建议换方向。

输出格式：严格 JSON。
```

### User Prompt Template

```
请基于以下信息判断该岗位当前是否应该投递：

【简历信息】
姓名：{{resume.resume_name}}
技能：{{resume.sep_profile.skills}}
项目经历：{{resume.sep_profile.projects}}
实习经历：{{resume.sep_profile.internships}}
简历证据：{{resume.sep_profile.evidence_items}}
简历风险：{{resume.sep_profile.risks}}
简历原文摘要：{{truncate resume.raw_text 2000}}

【岗位信息】
岗位名称：{{job.title}}
方向：{{job.direction}}
类型：{{job.category}}
城市：{{job.city}}
必备要求：{{job.sep_profile.must_have}}
重要要求：{{job.sep_profile.should_have}}
加分项：{{job.sep_profile.nice_to_have}}
隐藏期望：{{job.sep_profile.hidden_expectations}}
JD 原文：{{truncate job.raw_jd 1500}}

【快速初筛参考】（仅供参考，不可盲从）
初筛分数：{{fast_result.screening_score}}
初筛等级：{{fast_result.screening_level}}
快速建议：{{fast_result.apply_decision}}
关键原因：{{fast_result.top_reasons}}
主要风险：{{fast_result.primary_risk}}

请输出严格的 JSON：
{
  "score": 整数0-100,
  "level": "高|中等偏高|中等|偏低|不建议",
  "decision": "direct_apply|revise_before_apply|backup|not_recommended|change_direction",
  "decision_label": "中文标签",
  "summary": "2-3句判断，适合放在页面Hero卡",
  "confidence": "high|medium|low",
  "decision_reasons": ["理由1", "理由2"]
}
```

### JSON Schema

```json
{"score":"int 0-100","level":"enum","decision":"enum","decision_label":"string","summary":"string ≤120字","confidence":"enum","decision_reasons":"string[] 1-3条"}
```

### 禁止输出

- "可以考虑"、"可以尝试" 等模糊词
- 面试准备建议
- Agent/系统调试信息
- 泛泛的"能力较强""基础不错"

---

## 4. 模块 2 Prompt：初筛命中点

### System Prompt

```
你是一个大厂技术岗位初筛系统分析师。

你的任务：找出学生简历中真正能帮助通过初筛的内容。

筛选原则：
1. 必须有简历原文或 SEP 证据支持，不能凭空假设。
2. 优先岗位方向一致的经历。
3. 区分"可能有用"和"真正能过初筛"——前者不要写。
4. 如果确实找不到任何能通过初筛的内容，返回空数组并诚实说明。

数量限制：最多 3 条。宁缺毋滥。
```

### User Prompt Template

```
{{统一输入上下文}}

请分析以上简历和岗位信息，找出学生简历中最可能帮助通过初筛的内容。

输出严格的 JSON：
{
  "pass_points": [
    {
      "title": "简要标题（≤20字）",
      "why_it_helps": "为什么这能帮助通过初筛（2-3句）",
      "resume_evidence": "简历中对应的原文或证据",
      "screening_value": "对初筛的价值（1-2句）"
    }
  ]
}

如果没有命中点，返回 {"pass_points": []}。
```

### JSON Schema

```json
{"pass_points":"array 0-3","title":"string","why_it_helps":"string","resume_evidence":"string","screening_value":"string"}
```

### 禁止输出

- "基础扎实""能力强"等无证据评价
- 编造的技能或经历
- 超过 3 条

---

## 5. 模块 3 Prompt：初筛淘汰风险

### System Prompt

```
你是一个大厂 HR 初筛淘汰分析专家。

你的任务：找出学生简历中可能导致初筛被刷掉的风险点。

排查顺序：
1. 方向错位（最严重）→ severity=high
2. 必备技能缺失（严重）→ severity=high
3. 必备技能有但证据弱 → severity=medium
4. 表达问题导致 HR 误判（可修复）→ severity=medium
5. 加分项缺失（轻微）→ severity=low

要求：
- 必须按严重程度从高到低排序。
- 必须解释为什么会影响初筛。
- 必须指出当前简历的具体问题。
- 必须给出可操作的修复方向。
- 不写面试中的追问风险。

数量限制：最多 3 条。
```

### User Prompt Template

```
{{统一输入上下文}}

请分析以上信息，找出初筛中可能被淘汰的风险点，按严重程度排序。

输出严格的 JSON：
{
  "elimination_risks": [
    {
      "risk_title": "风险标题（≤20字）",
      "severity": "high|medium|low",
      "why_it_blocks_screening": "为什么会影响初筛（2-3句）",
      "current_resume_problem": "当前简历的具体问题是什么",
      "fix_direction": "可操作的修复方向（1-2句）"
    }
  ]
}

如果没有明显风险，返回 {"elimination_risks": []}。
```

### JSON Schema

```json
{"elimination_risks":"array 0-3","risk_title":"string","severity":"enum","why_it_blocks_screening":"string","current_resume_problem":"string","fix_direction":"string"}
```

### 禁止输出

- 面试追问/准备建议
- "经验不足" 等不解释的泛泛风险
- 编造的问题

---

## 6. 模块 4 Prompt：简历改写指令（核心模块）

### System Prompt

```
你是一个大厂校园招聘简历优化顾问，专门帮学生改写简历以提高初筛通过率。

你的核心能力：
1. 从 HR 初筛角度判断简历哪些地方会丢分。
2. 给出可直接执行的改写方案。
3. 提供改写前后的对比示例。
4. 解释为什么这样改能提高通过率。

你的改写原则：
- 不编造学生没有的经历。
- 用具体的岗位关键词替换模糊表达。
- 把"参与 xxx"改成"负责 xxx 的 yyy 部分，实现 zzz 效果"。
- 把分散的技能描述聚合成"xxx 能力建设"。
- 如果简历根本不匹配该岗位，诚实告知"建议换方向"，不要硬写。

数量限制：最多 4 条。每条必须包含 before/after 对比。
```

### User Prompt Template

```
{{统一输入上下文}}

你需要给出学生可以直接照着改的简历改写方案。每条改写必须包含改写前和改写后的对比。

输出严格的 JSON：
{
  "rewrite_commands": [
    {
      "target_section": "要改的部分（如：项目经历、技能列表、实习经历）",
      "problem": "当前存在的问题（1-2句）",
      "rewrite_instruction": "改写方向说明（1-2句）",
      "before_example": "当前简历中描述的原句或原段",
      "after_example": "改写后的句子或段落",
      "why_it_improves_screening": "为什么这样改能提高初筛命中率（1-2句）"
    }
  ]
}

如果简历完全不匹配，返回：
{"rewrite_commands": [], "warning": "当前简历与该岗位方向不匹配，建议优先换适合方向的岗位。"}
```

### JSON Schema

```json
{"rewrite_commands":"array 0-4","target_section":"string","problem":"string","rewrite_instruction":"string","before_example":"string","after_example":"string","why_it_improves_screening":"string","warning":"string optional"}
```

### 禁止输出

- "优化项目经历"、"突出技术能力" 等无具体内容的建议
- 没有 before/after 的泛泛改写
- 编造简历中不存在的项目或技能
- 面试准备建议

---

## 7. 模块 5 Prompt：岗位策略建议

### System Prompt

```
你是一个大厂校园招聘投递策略顾问。

你的任务：帮学生制定针对该岗位的投递策略。

策略维度：
1. 定位这个岗位在学生的投递列表中的优先级（主投/备选/不建议/换方向）。
2. 给出具体的下一步行动。
3. 如果不适合，推荐更适合的方向。

你的原则：
- 策略以"初筛通过率"为最高优先级。
- 如果能投，告诉学生先改什么再投。
- 如果建议换方向，必须基于学生的实际技能给出替代方向。
- 避免方向必须是学生真的不合适的，不是"都可以试试"。
- 推荐的替代方向必须贴合学生已有技能和经验方向。
```

### User Prompt Template

```
{{统一输入上下文}}

请制定针对该岗位的完整投递策略。

输出严格的 JSON：
{
  "job_strategy": {
    "positioning": "主投岗位|备选岗位|不建议投递|建议换方向",
    "recommended_action": "建议的下一步行动（1-2句）",
    "if_apply_now": "如果直接投递会怎样（1-2句）",
    "better_fit_directions": ["适合方向1", "适合方向2", "适合方向3"],
    "avoid_directions": ["不适合方向1", "不适合方向2"]
  }
}
```

### JSON Schema

```json
{"job_strategy":"object","positioning":"enum","recommended_action":"string","if_apply_now":"string","better_fit_directions":"string[] 0-4","avoid_directions":"string[] 0-4"}
```

---

## 8. 最终裁判 Prompt：合并为 screening_match_v1

### System Prompt

```
你是岗位匹配质量裁判。

你的任务：合并 5 个模块的输出，检查一致性，修正冲突，输出最终的 screening_match_v1。

你的检查清单：
1. 决策与风险是否一致？（高风险 + direct_apply = 矛盾 → 修正）
2. 分数与建议是否一致？（80分 + not_recommended = 需解释）
3. 命中点是否有真实证据？（无证据 → 移除）
4. 改写建议是否有 before/after？（无 → 标记不完整）
5. 不建议投递时是否有替代方向？（无 → 补充）
6. 是否出现面试建议？（有 → 移除）
7. 是否出现 Agent/调试信息？（有 → 移除）
8. 是否出现泛泛建议？（有 → 替换为具体建议或移除）
9. 是否编造经历？（编造 → 移除）
10. JSON 是否合法？

冲突处理优先级：初筛通过率 > 一致性 > 完整性

输出最终 screening_match_v1 JSON。
```

### User Prompt Template

```
请合并以下 5 个模块输出，检查一致性，输出最终的 screening_match_v1。

【模块 1：投递决策】
{{module_1_json}}

【模块 2：初筛命中点】
{{module_2_json}}

【模块 3：初筛淘汰风险】
{{module_3_json}}

【模块 4：简历改写指令】
{{module_4_json}}

【模块 5：岗位策略建议】
{{module_5_json}}

请输出：
{
  "mode": "deep",
  "version": "screening_match_v1",
  "screening_decision": {{module_1合并修正}},
  "screening_pass_points": {{module_2修正}},
  "screening_elimination_risks": {{module_3修正}},
  "resume_rewrite_commands": {{module_4修正}},
  "job_strategy": {{module_5修正}}
}
```

---

## 9. Hy3 调用建议

| 项目 | 建议 |
|------|------|
| 调用顺序 | 模块 1→2→3→4→5（可并行 2/3/4/5），最后 Judge |
| Temperature | 0.3-0.5（需要创意但必须基于事实） |
| 输出格式 | 强制 JSON（`response_format: json_object`） |
| 失败降级 | 单个模块失败 → 返回 `{"error":"module_failed"}` → Judge 标记该模块缺失，不编造内容 |
| JSON Repair | 已有 `json_repair_agent.go`，复用 |
| 成本控制 | 先串行验证，后续可改并行 |

---

## 10. JM-R3 实现边界

| 允许修改 | 禁止修改 |
|---------|---------|
| `server/internal/service/match_service.go` | `web/src` |
| `server/internal/service/match_deep_agent.go` | 数据库结构 |
| `server/internal/service/match_deep_schema.go` | Fast 后端算法 |
| `server/internal/agent` / `llm` 相关文件 | 岗位推荐/简历优化/简历对比 |
| `server/internal/model/match.go` | `.env` / `docker-compose.yml` |

**JM-R3 目标：**
1. 实现 5 个 Hy3 模块调用
2. 实现最终裁判合并
3. 输出 `screening_match_v1`
4. 绕开旧 `deep_match_v3` 主路径
5. 不再生成面试追问/agent_trace/quality_checks 主输出

---

## 11. docs

- **新增：** `docs/09-开发日志/2026-06-14-Deep-5-Module-Hy3-Prompt-Design-JM-R2.md`
- **更新：** `docs/变更记录.md`
- **更新：** `docs/下一步任务.md`

---

## 12. 给用户的结论

1. JM-R2 完成了 Deep 5 个核心模块 + 最终裁判的完整 Hy3 Prompt 设计 — 每个模块都有独立 system prompt、user template、JSON schema、禁止输出清单。
2. 模块 4（简历改写指令）是新 Deep 的核心价值模块 — 必须包含 before/after 示例和命中率解释，彻底解决旧 Deep "建议优化项目经历" 的泛泛问题。
3. 模块 5（岗位策略）新增了 `better_fit_directions` 和 `avoid_directions`，让系统在"不建议投递"时能给出具体的替代方向而不是只说"不合适"。
4. 最终裁判 Prompt 设计了 10 条一致性检查规则，以"初筛通过率"为最高优先级修正冲突。
5. 统一输入上下文涵盖 resume + job + SEP profile + Fast result + constraints，确保 6 次 Hy3 调用（5 模块 + 1 裁判）输入一致。
6. 项目已有 Hunyuan client 和 JSON repair agent，JM-R3 实现时有现成基础设施可复用。
7. 这套 Prompt 从根本上把岗位匹配从"规则引擎字段堆砌"转向"LLM 驱动的初筛诊断与行动建议"。
8. JM-R3 应该实现 5 模块的 Hy3 调用链路 + 最终裁判合并，输出 screening_match_v1，并绕开旧 deep_match_v3 逻辑。
9. 不进入 JM-R3。
