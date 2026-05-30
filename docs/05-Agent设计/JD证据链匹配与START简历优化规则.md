# JD 证据链匹配与 START 简历优化规则

> 创建时间：2026-05-30
> 所属阶段：4-3.5.D
> 依据文档：战略差异化与杀手锏确认、投递边界与优化停止条件设计、阶段4-3方案、岗位匹配Agent、岗位推荐Agent、Agent总览、Prompt模板、JSON输出规范、失败兜底与Mock回退策略

## 1. 背景

4-3.5.B 已完成战略差异化确认，项目定位从"岗位推荐工具"升级为"投递前决策与简历初筛优化系统"。

4-3.5.C 已完成投递边界与优化停止条件设计，定义了 `apply_readiness` 和 `optimization_stop_signal`，明确什么时候可以投、什么时候先改、什么时候不建议投、什么时候信息不足。

三个核心杀手锏中，**JD 证据链匹配**是杀手锏二，是让 Agent 输出从"给结论"升级为"给证据、边界、可核实建议"的关键能力。

当前 Agent 输出存在以下问题：

1. 匹配分析只给"适合/不适合"，但不说明依据哪个 JD 要求、对应哪条简历证据
2. 推荐只给岗位列表，但不说明为什么推荐、对应哪些简历证据
3. 简历优化建议使用"建议改为"，暗示用户必须照抄，而不是提供参考
4. 优化建议没有提醒用户核实，学生可能盲目照抄 AI 生成内容
5. 缺少检查简历表达完整性的框架

本文档定义 JD 证据链匹配和 START 简历优化的完整规则。

## 2. 核心问题：不能只给结论，必须给证据链

当前 Agent 输出的核心问题：

1. **只给结论不给证据**：说"你很适合该岗位"但不说明岗位要求
2. **脱离 JD 做泛泛建议**：说"建议突出项目经历"但不说明对应哪个 JD 要求
3. **"建议改为"式指令**：暗示用户必须照抄，而不是提供参考
4. **没有核实机制**：学生可能盲目照抄 AI 生成的经历或指标
5. **缺少表达完整性检查**：简历经历可能缺少关键要素但无人指出

**核心原则**：

> 每一条关键判断都应尽量绑定：JD 原始要求 → 简历中可见证据 → 当前匹配判断 → 差距或风险 → 可执行优化动作。
> 不允许脱离 JD 做职业规划式长篇建议。

## 3. 设计目标

1. 每条匹配判断都有 JD 依据和简历证据
2. 每条优化建议都绑定 JD 要求
3. 使用"示例参考"替代"建议改为"
4. 每条建议都提醒用户核实
5. 用 START / STAR-T 检查简历经历表达完整性
6. 不编造用户没有的经历、指标、岗位
7. 区分"表达优化"和"事实补充"
8. 为后续 Prompt / Schema / Agent 提供设计方向

## 4. 基本原则

### 4.1 JD 是判断的核心，不是装饰性引用

岗位 JD 是匹配分析、岗位推荐和简历优化的核心依据。Agent 不能脱离 JD 做泛泛建议，也不能只根据简历自行推断用户适合什么。

每一条关键判断都应尽量绑定：

1. JD 原始要求或要求摘要
2. 简历中可见证据
3. 当前匹配判断
4. 差距或风险
5. 可执行优化动作

**禁止**：

1. 只输出"你很适合该岗位"但不说明岗位要求
2. 只输出"建议突出项目经历"但不说明对应哪个 JD 要求
3. 脱离岗位 JD 做职业规划式长篇建议
4. 根据用户未提供的信息进行推断
5. 编造岗位要求

### 4.2 简历证据必须来自用户输入

`resume_evidence` 只能来自用户简历、在线简历 Profile、用户补充信息或系统已保存的结构化经历。

**禁止**：

1. 编造用户没有写过的项目
2. 编造用户没有做过的技术栈
3. 编造用户没有承担过的职责
4. 编造指标，如 QPS、DAU、提升百分比、用户规模
5. 把团队成果写成用户个人成果

**推荐表达**：

如果简历中没有足够证据，应输出"缺少证据"，而不是替用户补出一段看似更强的经历。

### 4.3 建议必须可核实

AI 的建议不一定正确，尤其在简历优化场景中，AI 可能误解用户经历或给出过度包装表达。因此每条关键建议都应提供自查问题、核实提醒或学习关键词。

建议字段：

- `self_check_question`：帮助用户确认建议是否适用的自查问题
- `learn_more_query`：用户可进一步搜索学习的方向或关键词
- `verification_notice`：提醒用户 AI 建议仅供参考，需自行核实

### 4.4 "示例参考"替代"建议改为"

简历优化输出中不使用"建议改为"作为强指令表达。统一使用"示例参考"。

原因：

1. 避免用户误以为必须照抄
2. 避免 AI 替用户编造经历
3. 强化"基于真实经历核实后使用"的边界
4. 降低过度承诺和误导风险

推荐结构：

```text
原表达
问题
影响初筛的原因
优化方向
示例参考
使用前请核实
自查问题
学习关键词
```

### 4.5 START / STAR-T 是检查框架，不是造句模板

START / STAR-T 技术岗表达法用于检查简历经历是否完整，不是让 AI 为用户编造完整故事。

如果用户没有提供 Result，Agent 不能编造结果。
如果用户没有提供 Tech，Agent 只能提示补充真实技术栈，不能假设用户用过某项技术。

## 5. JD 证据链标准结构

### 5.1 jd_evidence_chain 定义

`jd_evidence_chain` 是 JD 要求 → 简历证据 → 匹配判断 → 差距风险 → 优化动作的标准链路结构。

### 5.2 结构定义

```json
{
  "jd_requirement": "岗位要求摘要",
  "jd_requirement_source": "来自岗位 JD 的要求原文或摘要",
  "resume_evidence": "简历中对应证据",
  "evidence_strength": "strong / medium / weak / missing",
  "match_judgement": "matched / partially_matched / missing / risky",
  "gap_analysis": "差距说明",
  "screening_risk": "可能影响初筛的风险",
  "action_suggestion": "建议动作",
  "example_reference": "示例参考",
  "self_check_question": "自查问题",
  "learn_more_query": "学习关键词"
}
```

### 5.3 字段说明

#### jd_requirement

- **含义**：岗位 JD 要求的摘要
- **数据来源**：从岗位 JD 的 description / requirements 中提取
- **是否允许 Agent 生成**：允许总结，但不允许编造 JD 中没有的要求
- **禁止内容**：编造岗位要求、夸大岗位门槛、添加歧视性要求
- **示例**：`岗位要求 Go 后端开发经验`

#### jd_requirement_source

- **含义**：来自岗位 JD 的要求原文或摘要
- **数据来源**：直接引用 JD 中的原始表述
- **是否允许 Agent 生成**：允许压缩，但不允许歪曲原意
- **禁止内容**：脱离 JD 原意、编造不存在的岗位要求
- **示例**：`JD 原文："熟悉 Go 语言开发，有后端项目经验"`

#### resume_evidence

- **含义**：简历中与该 JD 要求对应的证据
- **数据来源**：只能来自用户简历内容
- **是否允许 Agent 生成**：允许整理和引用，不允许编造
- **禁止内容**：编造用户没有的项目、技术栈、职责、指标
- **示例**：`简历中技能栏包含 Go，项目经历中有 GoFrame 后台管理系统`

#### evidence_strength

- **含义**：简历证据对该 JD 要求的支撑强度
- **数据来源**：Agent 基于 JD 要求和简历证据判断
- **是否允许 Agent 生成**：允许判断
- **禁止内容**：无
- **示例**：`strong`

#### match_judgement

- **含义**：该 JD 要求的匹配判断
- **数据来源**：Agent 基于 evidence_strength 判断
- **是否允许 Agent 生成**：允许判断
- **禁止内容**：无
- **示例**：`matched`

#### gap_analysis

- **含义**：当前匹配差距说明
- **数据来源**：Agent 分析
- **是否允许 Agent 生成**：允许
- **禁止内容**：评价用户本人能力
- **示例**：`简历中缺少高并发场景的描述`

#### screening_risk

- **含义**：该差距可能影响初筛的风险类型和说明
- **数据来源**：Agent 分析
- **是否允许 Agent 生成**：允许
- **禁止内容**：恐吓用户、评价人格
- **示例**：`keyword_missing：初筛关键词匹配可能不通过`

#### action_suggestion

- **含义**：建议的优化动作
- **数据来源**：Agent 基于 JD 和简历分析
- **是否允许 Agent 生成**：允许
- **禁止内容**：建议编造经历
- **示例**：`在项目经历中补充 GoFrame 使用场景和 RBAC 实现细节`

#### example_reference

- **含义**：示例参考表达
- **数据来源**：Agent 生成
- **是否允许 Agent 生成**：允许，但必须提醒核实
- **禁止内容**：编造不存在的经历和指标
- **示例**：`在后台管理系统中使用 GoFrame 完成用户权限模块开发，结合 JWT/RBAC 实现接口权限校验`

#### self_check_question

- **含义**：帮助用户确认建议是否适用的自查问题
- **数据来源**：Agent 生成
- **是否允许 Agent 生成**：允许
- **禁止内容**：无
- **示例**：`请确认你是否确实使用过 GoFrame 框架，而不是只听说过`

#### learn_more_query

- **含义**：用户可进一步搜索学习的方向或关键词
- **数据来源**：Agent 生成
- **是否允许 Agent 生成**：允许
- **禁止内容**：无
- **示例**：`Go 后端实习 项目经历 STAR 写法 权限系统 简历表达`

## 6. evidence_strength 设计

### 6.1 枚举定义

| 值 | 定义 | 说明 |
|---|------|------|
| `strong` | 简历中有直接、具体、可验证的经历证据 | 有明确项目/实习/竞赛支撑，且表达清楚 |
| `medium` | 简历中有相关经历，但表达不够完整 | 有关键词或简短提及，缺少职责/结果/Tech |
| `weak` | 简历中只有关键词或泛泛描述 | 有技能标签但没有项目证据 |
| `missing` | 简历中没有对应证据 | 完全没有相关经历或关键词 |

### 6.2 约束

- `evidence_strength=missing` 时，不允许输出强匹配结论
- `evidence_strength=weak` 时，不允许建议用户直接夸大经历，只能提示补充真实证据
- `evidence_strength` 判断必须基于简历中可见的内容，不能凭推断

### 6.3 示例

**strong**：

```text
JD 要求：Go 后端开发经验
简历证据：项目经历中有"基于 GoFrame 完成后台管理系统用户权限模块开发"
判断：简历中有直接、具体、可验证的经历证据
```

**medium**：

```text
JD 要求：Redis 缓存经验
简历证据：技能栏包含 Redis，但项目经历中没有提及使用场景
判断：有关键词但缺少项目证据支撑
```

**weak**：

```text
JD 要求：高并发系统设计
简历证据：技能栏包含"高并发"标签
判断：只有标签，没有具体项目或场景
```

**missing**：

```text
JD 要求：Kubernetes 容器编排经验
简历证据：简历中没有任何相关内容
判断：完全缺少证据
```

## 7. match_judgement 设计

### 7.1 枚举定义

| 值 | 定义 | 说明 |
|---|------|------|
| `matched` | 该 JD 要求已有明确简历证据支撑 | evidence_strength=strong 时通常为 matched |
| `partially_matched` | 有部分相关经历，但证据不完整 | evidence_strength=medium 时通常为 partially_matched |
| `missing` | 完全缺少证据 | evidence_strength=missing 或 weak 时可能为 missing |
| `risky` | 看似相关，但可能存在夸大、不清楚或要求不完全一致 | 表达模糊、可能过度包装、方向不完全一致 |

### 7.2 与 evidence_strength 的关系

| evidence_strength | 典型 match_judgement |
|---|---|
| `strong` | `matched` |
| `medium` | `partially_matched` |
| `weak` | `missing` 或 `risky` |
| `missing` | `missing` |

### 7.3 约束

- `match_judgement` 是针对单条 JD 要求的判断，不是整体岗位匹配结论
- `risky` 不能用于暗示用户不诚实，只能指出表达可能存在初筛风险

## 8. screening_risk 设计

### 8.1 定义

`screening_risk` 用于说明该条差距为什么会影响初筛。

### 8.2 风险类型枚举

| 值 | 含义 | 说明 |
|---|------|------|
| `keyword_missing` | 关键词缺失 | 初筛关键词匹配不通过 |
| `responsibility_unclear` | 职责不清 | HR 无法判断你的实际贡献 |
| `result_missing` | 结果缺失 | 无法评估你的产出和影响 |
| `tech_stack_not_supported` | 技术栈未体现 | JD 要求的技术栈在简历中没有证据 |
| `scope_unclear` | 范围不清 | 不清楚是个人负责还是团队整体 |
| `experience_missing` | 经历缺失 | 缺少该岗位核心方向的任何经历 |
| `role_mismatch` | 角色不匹配 | 简历方向与岗位核心要求不一致 |
| `authenticity_risk` | 真实性风险 | 表达可能存在夸大或无法核实 |
| `location_or_preference_risk` | 地点或意向风险 | 工作地点或求职意向不匹配 |

### 8.3 约束

`screening_risk` 不是吓唬用户，而是帮助用户理解为什么该问题会影响 HR 初筛或简历筛选。

每条 `screening_risk` 应包含：

1. 风险类型
2. 风险说明
3. 对初筛的影响

### 8.4 示例

```json
{
  "risk_type": "keyword_missing",
  "description": "JD 要求 Kubernetes 经验，但简历中未体现",
  "impact": "初筛关键词匹配可能不通过，建议在技能栏或项目经历中补充（仅当你确实使用过时）"
}
```

## 9. START / STAR-T 技术岗表达法

### 9.1 方法定义

START / STAR-T 是技术岗简历表达法，用于检查简历经历是否完整。

| 要素 | 英文 | 说明 |
|------|------|------|
| S | Situation | 项目 / 业务背景 |
| T | Task | 你承担的任务 |
| A | Action | 你采取的行动 |
| R | Result | 结果 / 产出 / 影响 |
| T | Tech | 技术栈 / 工程方法 / 工具链 |

也可写作 **STAR-T 技术岗表达法**：Situation / Task / Action / Result / Tech。

### 9.2 各要素说明

#### S - Situation

**说明**：项目 / 业务背景。用户需要说明这个项目解决什么问题，或者处于什么业务场景。

**常见缺失**：只写"做了一个管理系统"，但没有说明系统服务什么场景。

**示例**：

```text
完整：在校园二手交易平台项目中（S）
缺失：做了一个管理系统
```

#### T - Task

**说明**：个人承担的任务。用户需要说明自己负责哪个模块、哪类接口、哪块功能。

**常见缺失**：只写"参与项目开发"，没有说明个人职责边界。

**示例**：

```text
完整：负责用户权限模块的设计与实现（T）
缺失：参与项目开发
```

#### A - Action

**说明**：具体采取的行动。包括技术选型、实现方式、问题处理、协作过程。

**常见缺失**：只写"负责开发"，没有说明如何开发。

**示例**：

```text
完整：使用 GoFrame 实现 RBAC 权限控制，结合 JWT 完成接口鉴权（A）
缺失：负责开发
```

#### R - Result

**说明**：结果 / 产出 / 影响。可以是功能上线、性能提升、联调完成、问题减少、业务流程打通等。

**注意**：没有真实数据时，不允许编造百分比或规模。可以使用"完成""支持""沉淀""联调通过"等保守表达。

**常见缺失**：只写"做了什么"，没有写"产生了什么效果"。

**示例**：

```text
完整：上线后权限相关工单减少，系统稳定性提升（R）
可接受的保守表达：完成权限模块开发并上线，支持 5 个角色的权限管理（R）
不允许：上线后 QPS 提升 300%，DAU 增长 50%（编造数据）
```

#### T - Tech

**说明**：技术栈 / 工程方法 / 工具链。用于说明该经历和技术岗位 JD 的匹配依据。

**常见缺失**：堆技术名词，但没有说明技术如何被使用。

**示例**：

```text
完整：使用 GoFrame 框架 + MySQL + Redis + Docker（T）
缺失：熟悉 Go、MySQL、Redis、Docker（技术堆砌）
```

### 9.3 START 缺失检查

Agent 应检查每段经历是否缺少 S/T/A/R/T 中的关键要素，并给出补充提示。

`start_missing_parts` 字段用于标识当前经历中缺少的 START 要素。

```json
{
  "start_missing_parts": ["Task", "Result", "Tech"]
}
```

### 9.4 约束

- 如果用户没有提供 Result，Agent 不能编造结果
- 如果用户没有提供 Tech，Agent 只能提示补充真实技术栈，不能假设用户用过某项技术
- START 检查是辅助用户完善表达，不是让 AI 替用户编造完整故事
- `start_missing_parts` 只能包含 `Situation` / `Task` / `Action` / `Result` / `Tech`

## 10. 简历优化输出标准结构

### 10.1 结构定义

后续 ResumeOptimizationAgent 的建议结构：

```json
{
  "target_jd_requirement": "目标岗位要求",
  "current_resume_expression": "当前简历表达",
  "problem": "当前问题",
  "why_it_matters": "为什么影响初筛",
  "start_missing_parts": ["Task", "Result", "Tech"],
  "optimization_direction": "优化方向",
  "example_reference": "示例参考",
  "verification_notice": "使用前请核实",
  "self_check_question": "自查问题",
  "learn_more_query": "学习关键词",
  "estimated_effort": "within_10_min / within_30_min / within_60_min / long_term"
}
```

### 10.2 字段说明

| 字段 | 说明 | 来源 |
|------|------|------|
| `target_jd_requirement` | 目标岗位要求 | 来自 JD |
| `current_resume_expression` | 当前简历表达 | 来自简历原文 |
| `problem` | 当前问题 | Agent 分析 |
| `why_it_matters` | 为什么影响初筛 | Agent 分析，应引用 screening_risk |
| `start_missing_parts` | START 缺失要素 | Agent 检查 |
| `optimization_direction` | 优化方向 | Agent 建议 |
| `example_reference` | 示例参考 | Agent 生成，必须提醒核实 |
| `verification_notice` | 使用前请核实 | 固定提醒 |
| `self_check_question` | 自查问题 | Agent 生成 |
| `learn_more_query` | 学习关键词 | Agent 生成 |
| `estimated_effort` | 预计耗时 | Agent 评估 |

### 10.3 约束

- `example_reference` 不是最终简历文本
- 用户不能直接复制未核实内容
- Agent 不应把不存在的经历写进 `example_reference`
- `verification_notice` 应为固定提醒或类似表达

## 11. 示例参考输出规范

### 11.1 规范要点

1. 使用"示例参考"而非"建议改为"
2. 每条示例参考必须附带核实提醒
3. 不能编造用户没有的经历和指标
4. 示例应基于用户简历中的真实经历框架
5. 保守表达优于包装过度

### 11.2 固定核实提醒

每条示例参考应附带以下核实提醒之一：

```text
仅当你确实使用过 [技术/方法] 并负责过 [职责] 时，才可以参考该表达。
```

或：

```text
以下示例仅供表达参考，AI 可能理解错误。请基于你的真实经历核实后使用，不要加入没有实际发生过的经历、指标或职责。
```

### 11.3 示例参考案例

#### 示例 1：表达太泛

**当前表达**：`负责后台接口开发。`

**问题**：缺少技术栈、职责边界和结果。

**示例参考**：`基于 GoFrame 完成后台管理系统用户、角色、权限相关接口开发，结合 JWT / RBAC 完成接口权限校验，并与前端完成联调。`

**核实提醒**：`仅当你确实使用过 GoFrame、JWT、RBAC 并负责过相关接口时，才可以参考该表达。`

---

#### 示例 2：缺少结果

**当前表达**：`使用 Redis 做缓存。`

**问题**：没有说明缓存用于什么场景，也没有说明结果。

**示例参考**：`在课程平台项目中使用 Redis 缓存高频查询数据，降低重复数据库查询，并配合缓存过期策略保证数据更新。`

**核实提醒**：`如果没有实际做过缓存策略设计，不要写"缓存过期策略"。`

---

#### 示例 3：技术堆砌

**当前表达**：`熟悉 Go、MySQL、Redis、Docker、Linux。`

**问题**：技术栈堆砌，缺少项目证据。

**示例参考**：`在后台管理系统项目中使用 Go 完成接口开发，使用 MySQL 设计业务表结构，并使用 Redis 处理登录态或高频查询场景。`

**核实提醒**：`请根据你实际做过的模块替换示例中的场景和技术。`

## 12. JD 证据链与 apply_readiness 的关系

### 12.1 核心规则

`apply_readiness` 不应只依赖一个总分。它应该综合多个 `jd_evidence_chain` 的结果判断。

### 12.2 判断规则

**ready_to_apply**：

- 多数核心 JD 要求为 `matched` 或 `partially_matched`
- `missing` 项不是核心硬要求
- 不存在严重 `risky` 判断

**quick_fix_before_apply**：

- 有匹配基础，但存在 2-3 个 `weak` / `partially_matched` 的表达问题
- 可通过短期修改改善
- 问题主要是 `keyword_missing` / `responsibility_unclear` / `result_missing` 类型

**not_ready_for_this_role**：

- 核心 JD 要求多为 `missing`
- 存在 `experience_missing` 或 `role_mismatch` 类型风险
- 不是简单改写能解决的问题

**insufficient_information**：

- 简历证据过少，无法形成可靠 `jd_evidence_chain`
- 多数 JD 要求的 `evidence_strength` 为 `missing`，但不是因为能力不匹配，而是因为信息不足

### 12.3 约束

- 不能只看 match_score 数字判断 apply_readiness
- 必须看 jd_evidence_chain 的分布
- 即使总分较高，如果核心硬要求全部 missing，也不应输出 ready_to_apply

## 13. JD 证据链与 RecommendationAgent 的关系

### 13.1 核心规则

RecommendationAgent 不应只给岗位排序。每个推荐岗位应尽量给出推荐证据。

### 13.2 推荐字段方向

| 字段 | 说明 |
|------|------|
| `recommend_reason` | 为什么推荐该岗位 |
| `evidence` | 对应哪些简历证据 |
| `why_this_job_first` | 为什么先投这个岗位 |
| `apply_readiness` | 投递准备度 |
| `recommendation_confidence` | 推荐置信度 |

### 13.3 约束

- 推荐岗位仍必须来自候选岗位集，不允许 LLM 编造岗位
- 每个推荐都应关联至少 1 条 jd_evidence_chain
- 推荐不是只告诉用户投什么，还要告诉用户：
  1. 先投哪个
  2. 为什么先投
  3. 哪个岗位需要先补简历
  4. 哪个岗位暂不建议投

## 14. JD 证据链与 ResumeOptimizationAgent 的关系

### 14.1 核心规则

ResumeOptimizationAgent 的输出必须围绕 JD 证据链展开，不能脱离 JD 做泛泛建议。

### 14.2 输出方向

1. 每条优化建议必须绑定 `target_jd_requirement`
2. 每条建议必须说明 `why_it_matters`（引用 screening_risk）
3. 必须使用 `example_reference` 而非"建议改为"
4. 必须附带 `verification_notice` / `self_check_question` / `learn_more_query`
5. 必须检查 START 缺失要素
6. 必须给出 `optimization_stop_signal`

### 14.3 约束

- 简历优化不是无限改写
- 先修最影响初筛的点
- 修完就投
- 长期提升另列
- 不能编造经历和指标

## 15. JD 证据链与 Quality Gate 的关系

### 15.1 后续 Quality Gate 检查项

后续 Quality Gate 可以检查：

| 序号 | 检查项 | 说明 |
|------|--------|------|
| 1 | 是否每条建议绑定 JD 要求 | 不能脱离 JD 做泛泛建议 |
| 2 | 是否每条建议有简历证据 | 不能凭空建议 |
| 3 | 是否出现泛泛建议 | 如"提升自己""多学习"等无操作性的建议 |
| 4 | 是否使用"示例参考"而不是"建议改为" | 避免指令式表达 |
| 5 | 是否包含核实提醒 | verification_notice |
| 6 | 是否编造经历或指标 | 不允许 AI 替用户编造 |
| 7 | 是否缺少 learn_more_query | 每条建议应提供学习方向 |
| 8 | 是否给出与 apply_readiness 矛盾的建议 | 如 ready_to_apply 但建议继续大幅优化 |

### 15.2 当前状态

本阶段只做设计确认，不实现 Quality Gate。

4-3.5.E 已完成 Agent 输出 Quality Gate 设计。Quality Gate 可检查 Agent 输出是否缺少 JD 证据链、是否泛泛建议、是否编造经历、是否缺少核实提醒、是否使用"建议改为"而非"示例参考"。fabrication_check 一旦失败应直接 fallback_required。详见 `docs/05-Agent设计/Agent输出QualityGate设计.md`

---

4-3.5.F 已完成学生求职极端情况处理边界设计。极端情况对 JD 证据链的影响：

1. **insufficient_information**：简历证据过少，无法形成可靠 jd_evidence_chain，多数 JD 要求的 evidence_strength 为 missing
2. **weak_resume**：证据链中 weak / partially_matched 占多数，但不是方向不匹配
3. **target_role_mismatch**：核心 JD 要求多为 missing，存在 experience_missing 或 role_mismatch 类型 screening_risk
4. **authenticity_risk**：简历表达存在真实性风险，evidence_strength=weak 但 match_judgement=risky
5. **unsafe_user_request**：用户要求编造证据链中不存在的证据，必须拒绝

极端情况下 JD 证据链的 evidence_strength 和 match_judgement 应反映实际情况，不能因用户请求而强行调整。详见 `docs/05-Agent设计/学生求职极端情况处理边界.md`

## 16. 对 Prompt / Schema 的后续影响

### 16.1 可评估新增字段

| 字段 | 所属 Agent | 说明 |
|------|-----------|------|
| `jd_evidence_chain` | 匹配分析 / 推荐 / 简历优化 | JD 证据链 |
| `jd_requirement` | 匹配分析 / 简历优化 | JD 要求摘要 |
| `jd_requirement_source` | 匹配分析 / 简历优化 | JD 原始要求 |
| `resume_evidence` | 匹配分析 / 简历优化 | 简历对应证据 |
| `evidence_strength` | 匹配分析 / 简历优化 | 证据强度 |
| `match_judgement` | 匹配分析 | 匹配判断 |
| `gap_analysis` | 匹配分析 / 简历优化 | 差距说明 |
| `screening_risk` | 匹配分析 / 简历优化 | 初筛风险 |
| `action_suggestion` | 匹配分析 / 简历优化 | 建议动作 |
| `example_reference` | 简历优化 | 示例参考 |
| `verification_notice` | 匹配分析 / 简历优化 | 核实提醒 |
| `self_check_question` | 匹配分析 / 简历优化 | 自查问题 |
| `learn_more_query` | 匹配分析 / 简历优化 | 学习关键词 |
| `start_missing_parts` | 简历优化 | START 缺失要素 |
| `estimated_effort` | 简历优化 | 预计耗时 |

### 16.2 重要说明

本阶段不修改 Prompt 文件和 Schema 文件。仅登记后续优化方向。

是否真正修改 Prompt / Schema，放到 4-3.5.H 或后续 Prompt/Schema 评审阶段决定。

## 17. 对前端展示的影响

### 17.1 匹配分析页面

- 证据链可视化（JD 要求 ←→ 简历证据 ←→ 匹配判断）
- evidence_strength 颜色标识（strong=绿 / medium=黄 / weak=橙 / missing=灰）
- screening_risk 风险标签
- 每条优化建议附带"示例参考"和"使用前请核实"提示
- START 缺失检查提示

### 17.2 推荐页面

- 推荐理由增加 JD 证据链展开
- 推荐置信度提示
- 每个推荐岗位的投递准备度标签

### 17.3 简历优化页面

- 改写建议标注"示例参考"而非"优化后"
- 增加 START 缺失检查提示
- 增加自查问题和学习关键词
- 每条建议附带 verification_notice
- screening_risk 可视化

### 17.4 重要说明

前端展示改造属于阶段 5/6 范围，本阶段不修改前端代码。

## 18. 当前不做事项

| 不做 | 原因 |
|------|------|
| 不修改 Prompt 文件 | 本阶段只做设计登记 |
| 不修改 Schema 文件 | 本阶段只做设计登记 |
| 不修改 Go 代码 | 本阶段只做 docs only |
| 不修改前端代码 | 本阶段只做 docs only |
| 不实现 Quality Gate | 只做设计确认，4-3.5.E 再评估 |
| 不实现 jd_evidence_chain 字段 | 只做结构设计 |
| 不实现 START 检查 | 只做规则设计 |
| 不调用真实 Hunyuan | 本阶段不涉及 LLM 调用 |
| 不进入 4-3.6 | 本阶段是设计，不是代码开发 |

## 19. 是否需要立即改工程

**本阶段暂不立即改工程代码。**

先完成 JD 证据链匹配与 START 简历优化规则设计。

后续是否修改 Prompt / Schema / Agent 输出结构，放到 4-3.5.H 统一评审。

## 20. 下一步建议

**下一步进入 4-3.5.E：Agent 输出 Quality Gate 设计。**

4-3.5.E 的核心内容：

1. 设计 Quality Gate 检查项
2. 设计重试策略
3. 设计保守 fallback
4. 仍然以 docs only 为主，不急于改代码

不进入 4-3.6。
