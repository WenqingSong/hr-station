# Job-Match-Screening-Oriented-Schema-Freeze-JM-R1 完成报告

> 执行时间：2026-06-14 01:26
> 类型：岗位匹配初筛导向新 Schema 冻结（设计冻结，不修改代码）
> 状态：✅ 冻结完成，JM-R2 可执行

---

## 1. 总体结论

| 项目 | 结果 |
|------|------|
| 是否完成 JM-R1 | ✅ Schema 冻结完成 |
| 是否修改代码 | ❌ 未修改 |
| 是否执行 Git | ❌ 未执行 |
| 是否冻结新 Fast Schema | ✅ 8 字段极简 Schema |
| 是否冻结新 Deep Schema | ✅ 5 模块 screening_match_v1 Schema |
| 是否废弃旧 Deep 10-module | ✅ 废弃 |
| 是否废弃面试追问主展示 | ✅ 废弃 |
| 是否废弃 Agent Trace / quality_checks 主展示 | ✅ 废弃 |
| 是否可以进入 JM-R2 | ✅ 可以 |

---

## 2. 新岗位匹配定义

```
岗位匹配 = 初筛通过率诊断 + 简历命中率提升方案
```

**核心不是判断"简历和 JD 有多像"，而是判断：**
1. 这个岗位投递是否值得
2. 当前简历能否通过初筛
3. 哪些内容会帮助通过初筛
4. 哪些内容会导致被筛掉
5. 简历应该如何立刻改写
6. 如果不适合，应该换什么方向

---

## 3. 新 Fast 定位与 Schema

```
Fast = 秒级初筛判断（不做长模块展示）
```

Fast 只输出 8 个字段：

| # | 字段 | 含义 | 要求 |
|---|------|------|------|
| 1 | `mode` | `"fast"` | 固定值 |
| 2 | `screening_score` | 初筛通过概率 0-100 | 整数 |
| 3 | `screening_level` | 高 / 中等偏高 / 中等 / 偏低 / 不建议 | 中文等级 |
| 4 | `apply_decision` | direct_apply / revise_before_apply / backup / not_recommended / change_direction | 投递决策 code |
| 5 | `decision_label` | 可以投递 / 修改后投递 / 备选投递 / 不建议投递 / 建议换方向 | 中文 |
| 6 | `one_sentence` | 一句话判断 ≤80 字 | 必须具体 |
| 7 | `top_reasons` | 1-2 条关键原因 | 必须绑定证据 |
| 8 | `primary_risk` | 最大风险 | 只保留最关键 |
| 9 | `next_action` | 下一步动作 | 必须可执行 |

**Fast 禁止输出到主页面：** 完整 coverage、完整 evidence、多个风险长列表、面试建议、Agent 信息、调试字段

---

## 4. 新 Deep 5 模块

```
Deep = 初筛命中率提升方案（LLM 驱动，非规则引擎）
version: "screening_match_v1"
```

| # | 模块 | 解决的问题 | 核心字段 | 条数限制 |
|---|------|---------|---------|---------|
| 1 | **screening_decision** | 该不该投？ | score + level + decision + decision_label + summary + confidence | 1 |
| 2 | **screening_pass_points** | 为什么能过初筛？ | title + why_it_helps + resume_evidence + screening_value | ≤3 |
| 3 | **screening_elimination_risks** | 为什么可能被筛？ | risk_title + severity(high/medium/low) + why_it_blocks_screening + current_resume_problem + fix_direction | ≤3 |
| 4 | **resume_rewrite_commands** | 简历具体怎么改？ | target_section + problem + rewrite_instruction + before_example + after_example + why_it_improves_screening | ≤4 |
| 5 | **job_strategy** | 投递策略？ | positioning(主投/备选/不建议/换方向) + recommended_action + if_apply_now + better_fit_directions + avoid_directions | 1 |

---

## 5. 决策枚举

### apply_decision

| code | label |
|------|-------|
| `direct_apply` | 可以直接投递 |
| `revise_before_apply` | 修改后投递 |
| `backup` | 备选投递 |
| `not_recommended` | 不建议投递 |
| `change_direction` | 建议换方向 |

### screening_level

| 值 |
|---|
| 高 |
| 中等偏高 |
| 中等 |
| 偏低 |
| 不建议 |

### severity / confidence

| 值 |
|---|
| high |
| medium |
| low |

---

## 6. 内容质量标准

### 必须遵守

1. 每一句都服务初筛通过率
2. 每个建议都能被学生直接执行
3. 每个风险都说明为什么影响初筛
4. 每个命中点都绑定真实简历证据
5. 简历改写指令必须包含 before/after 示例
6. 岗位策略必须给出具体方向和行动

### 禁止输出

```
1. "建议提升技术能力"
2. "建议优化项目经验"
3. "你的经历比较匹配"（无证据）
4. "可以加强简历表达"
5. "面试时可以重点准备 xxx"
6. "本报告由 deep_match_v3 生成"
7. "Agent 已完成 JD_DECOMPOSITION"
```

### 推荐输出

```
把"参与课程平台建设"改成"参与企业课程学习平台后端核心链路建设，负责课程权限校验、播放签名鉴权、学习进度一致性与资源下载权限控制，补齐登录态、课程状态、学习权限、小节归属、视频绑定等多层校验，降低越权访问风险。"
```

---

## 7. 页面展示 Schema

### 页面只展示 5 个区域

| 区域 | 内容 | 来源模块 |
|------|------|---------|
| **A. 顶部决策区** | score + decision_label + summary + recommended_action | screening_decision |
| **B. 为什么可能过初筛** | pass points，≤3 条，每条含证据 | screening_pass_points |
| **C. 为什么可能被筛掉** | elimination risks，≤3 条，按 severity 排序 | screening_elimination_risks |
| **D. 简历怎么改** | rewrite commands，≤4 条，before/after 清晰 | resume_rewrite_commands |
| **E. 投递策略** | positioning + recommended_action + better/avoid directions | job_strategy |

### 页面不展示

完整 coverage、完整 evidence、面试追问、Agent Trace、quality_checks、generated_by、report_version、fallback_reason

---

## 8. 旧字段处理策略

| 旧字段 | 处理方式 |
|--------|---------|
| `requirement_coverage` | 后端可保留用于 internal 计算，前端主页面不读取 |
| `resume_evidence_mapping` | 同上 |
| `agent_trace` | 仅日志/db 存储，前端不读取 |
| `quality_checks` | 仅日志/db 存储，前端不读取 |
| `interview_risk_questions` | 移入独立面试准备模块，不在此页面 |
| `score_breakdown` | 后端保留，前端不主展示 |
| `score_alignment` | 后端保留，前端不主展示 |
| `job_understanding` | 融入 5 模块的自然语言 |
| `resume_rewrite_strategy`（旧） | **被 resume_rewrite_commands 取代** |
| `application_strategy`（旧） | **被 job_strategy 取代** |

---

## 9. 经典案例校准目标（JM-R5）

| # | 案例 | 目标 |
|---|------|------|
| 1 | 后端简历 × 后台开发岗位 | 修改后投递 / 主投岗位 — 展示如何提高命中率 |
| 2 | 后端简历 × 算法/咨询岗位 | 不建议投递 / 建议换方向 — 展示系统不盲目鼓励投递 |
| 3 | 薄弱简历 × 后台开发岗位 | 不建议直接投 / 先补项目证据 — 展示系统能给成长路径 |

---

## 10. 后续任务

| 阶段 | 目标 | 改代码 |
|------|------|:---:|
| **JM-R2** | Deep 5 模块 Hy3 Prompt 设计（不写代码） | ❌ |
| **JM-R3** | Deep 5 模块后端生成链路实现 | ✅ |
| **JM-R4** | 岗位匹配页面 A-E 区域重新设计 | ✅ |
| **JM-R5** | 3 案例内容校准 | ✅ |
| **JM-R6** | 最终演示验收 | ❌ |

---

## 11. docs

- **新增：** `docs/09-开发日志/2026-06-14-Job-Match-Screening-Oriented-Schema-Freeze-JM-R1.md`
- **更新：** `docs/变更记录.md`
- **更新：** `docs/下一步任务.md`

---

## 12. 给用户的结论

1. JM-R1 冻结了面向初筛命中率的全新输出 Schema — Fast 精简为 8 字段极简 Schema，Deep 定义 5 个核心模块（决策/命中点/淘汰风险/改写指令/岗位策略）。
2. 新 Schema 的核心改进：每个模块都有明确的"学生问题"对应 — 该不该投、为什么能过、为什么会被筛、怎么改简历、下一步怎么投。
3. 废弃了旧 Deep 10-module 展示结构 — 面试追问、Agent Trace、quality_checks、完整 coverage/evidence 表不再进入学生主视野。
4. 新增 `resume_rewrite_commands` 模块替代旧的 `resume_rewrite_strategy` — 必须包含 before/after 示例和为什么这样改能提高命中率，而非泛泛建议。
5. 新增 `job_strategy` 模块代替旧的 `application_strategy` — 明确主投/备选/不建议/换方向，并给具体替代方向推荐。
6. 新增 `change_direction` 决策枚举 — 当岗位完全不匹配时，不只说"不建议投"，还要说"你应该换什么方向"。
7. 内容质量冻结了 10 条必须 + 7 条禁止 + 1 条推荐示例 — 从根本上杜绝浅、空、无行动价值的输出。
8. 旧字段处理策略明确 — 后端可保留用于计算，但前端主页面不读取、不展示。
9. JM-R2 应该为这 5 个模块分别设计 Hy3 深度思考 Prompt，确保每模块能输出自然、具体、可执行的学生导向内容。
10. 不进入 JM-R2。
