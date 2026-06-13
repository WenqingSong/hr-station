# Job-Match-Core-Rebuild-Baseline-Audit-JM-R0 完成报告

> 执行时间：2026-06-14 01:21
> 类型：岗位匹配核心功能重构基线审计（只读，不修改代码）
> 状态：✅ 审计完成，JM-R1 可执行

---

## 1. 总体结论

| 项目 | 结果 |
|------|------|
| 是否完成 JM-R0 | ✅ 审计完成 |
| 是否修改代码 | ❌ 未修改 |
| 是否执行 Git | ❌ 未执行 |
| 当前岗位匹配最大问题 | **价值表达错了 — 把"字段完整"当成"产品有用"** |
| 是否确认需要全面重构 | ✅ 是 — 页面、后端输出模块、Deep 生成逻辑均需重构 |
| 是否可以进入 JM-R1 | ✅ 可以 |

---

## 2. 当前页面问题审计

| 问题 | 当前表现 | 为什么影响评委/学生 |
|------|---------|-------------------|
| **Fast 展示过重** | 虽已缩到 Hero+3卡，但卡内仍有 coverage summary/evidence summary 两行 | 对学生而言，"JD要求覆盖摘要"没有决策价值 |
| **Deep 内容浅** | Deep v3 内容来自 SEP 规则引擎（非 LLM），key_risks 从 Fast structured_risks 拷贝，delta_reason 模板化 | 学生读完 Deep 仍然不知道怎么改简历 |
| **面试追问偏题** | interview_risk_questions 虽然在折叠区，但作为模块存在 | 岗位匹配不是面试准备 — 产品边界混淆 |
| **coverage/evidence 无用户价值** | 摘要 4 条覆盖 + 3 条证据，但学生对 "MySQL:65" 无感知 | 学生需要的是"为什么 MySQL 不够"而不是"MySQL 覆盖 65 分" |
| **调试信息污染** | agent_trace/quality_checks 已在折叠区 | 但存在本身就是产品债务 — 不应该花时间生成 |
| **页面缺少高级感** | 当前是"卡片+列表"模式，没有层次感、呼吸感 | 评委第一眼看不出产品深度，用户也很难信任匹配判断 |
| **后端 Deep 不是 LLM** | Deep v3 的 8-layer 是确定性规则引擎（无 LLM 调用） | 内容浅、模板化、不自然、无法给出具体改写建议 |
| **字段驱动设计** | 整个 Deep 输出 Schema 有 25 个顶层字段 | 设计逻辑是"后端有什么就展示什么"，而非"用户需要什么判断" |

**核心结论：当前页面失败不是因为字段少，而是因为价值表达错了。**

---

## 3. 当前 API 输出审计

### Fast 当前输出字段

| 有价值 | 无价值/调试 |
|--------|------------|
| match_score (分数) | generated_by |
| match_level (等级) | agent_status |
| application_decision (投递决策) | fallback_reason |
| decision_label (决策标签) | score_breakdown (5 维对学生太细) |
| summary (一句话结论) | hard_gate (学生不关心 gate 状态) |
| core_advantages (核心优势) | requirement_coverage (全量对学生无决策价值) |
| structured_risks (风险) | resume_evidence (证据列表无叙事) |
| action_suggestions (建议) | — |

### Deep 当前输出字段

| 有价值 | 无价值/调试 |
|--------|------------|
| match_score | generated_by / agent_status |
| match_level | fallback_reason / report_version |
| application_decision / decision_label | score_breakdown (6 维) |
| one_sentence_conclusion | score_alignment (内部审计) |
| job_understanding (must_have) | quality_checks |
| key_risks (风险) | agent_trace |
| resume_rewrite_strategy | interview_risk_questions |
| — | requirement_coverage (全量) |
| — | resume_evidence_mapping (全量) |
| — | core_match_points (信息与优势重复) |
| — | application_strategy |
| — | action_plan |

**问题总结：**
- Deep 真正对学生有价值的是：分数、投递决策、风险诊断、简历改写 — 但这 4 块恰好是规则引擎生成的最弱部分
- 规则引擎生成的大段结构化数据（coverage/evidence/trace）对学生零价值
- Deep 缺少了最核心的：**LLM 驱动的具体改写指令和岗位策略**

---

## 4. 当前代码结构审计

| 文件 | 行数 | 当前职责 | 后续处理 |
|------|------|---------|---------|
| `match_service.go` | 951 | Fast/Deep 路由 + context build + legacy fallback | 保留核心路由，简化 |
| `match_sep_engine.go` | 1077 | Fast 确定性匹配引擎（direction/gate/coverage/evidence/risk/calibrator） | Fast 保留作为快速初筛 |
| `match_deep_agent.go` | 1030 | Deep v3 8-layer 确定性 pipeline + legacy buildDeepMatchReport | **核心重构文件**：改为 LLM 驱动 |
| `match_deep_schema.go` | ~210 | DeepMatchV3Report 等所有深对 schema 类型 | 需要新 Schema |
| `match_deep_prompt.go` | — | LLM prompt 定义（声明但未使用） | **核心重构文件**：Hy3 Prompt |
| `JobMatchAnalysisPage.tsx` | 448 | Fast/Deep 页面渲染 | **核心重构文件**：全新设计 |
| `matchAnalysisAdapter.ts` | 213 | 后端字段→ViewModel 映射 | 需要大幅精简 |
| `match_deep_agents.go` (agent/) | — | 3-agent tag-based pipeline (legacy) | 可以删除或彻底隔离 |
| `match_agent.go` (agent/) | — | 单 LLM match_analysis agent | 可以删除或改作他用 |

---

## 5. 新岗位匹配定义

```
岗位匹配 = 初筛通过率诊断 + 简历命中率提升方案
```

**不是：** 简历和 JD 的相似度报告
**不是：** 后端字段的全面展示
**不是：** 面试准备工具
**不是：** Agent 调试信息展示

---

## 6. 新 Fast 定位

```
Fast = 快速初筛判断（秒级）

只输出：
1. 一句话结论（50字以内）
2. 是否建议投递（priority_apply / apply / revise / skip）
3. 1-2 个关键原因（为什么是这个判断）

不输出：
- 完整 coverage 表
- 完整 evidence 列表
- 5 维 score_breakdown
- hard_gate
- Agent 信息
- 面试追问
```

---

## 7. 新 Deep 定位

```
Deep = 初筛命中率提升方案（LLM 驱动，非规则引擎）

5 个核心模块：

1. 投递决策
   - 这个岗位该不该投（priority_apply / apply_directly / revise_before_apply / backup_apply / not_recommended）
   - 必须给出 2-3 句明确理由
   - 不允许泛泛 "匹配度一般"

2. 初筛命中点（最多 3 条）
   - HR/ATS 系统为什么可能让你通过初筛
   - 每条必须绑定简历真实证据
   - 不允许编造
   - 示例：你的 Go 后端实习包含 RESTful API、权限校验、上线部署 → 这直接覆盖后台开发岗位的 2 个核心关键词

3. 初筛淘汰风险（最多 3 条）
   - 为什么可能被筛掉
   - 按严重程度排序（最高风险最先）
   - 必须说明具体会卡在哪个筛选环节
   - 必须给出降低风险的具体动作
   - 示例：K8s 缺失 → 部分后台岗位会用 K8s 做关键词过滤 → 你虽然没有生产 K8s 经验，但在 Docker 部署和 Nginx 配置方面有实践，可以在简历里补充 K8s 基础概念和 minikube 实验

4. 简历改写指令（2-5 条）
   - 不是 "建议优化"——是可直接照着改的句子
   - 每条指令必须说明改写后为什么能提高命中率
   - 示例：将第 2 个项目描述的第一句改为 "负责企业课程平台后端核心模块，独立完成播放鉴权、课程权限、学习进度、资源下载等 6 个模块的接口开发与上线部署" → 这样改可以把 "参与" 升级为 "负责"，关键词密度提升 3 倍

5. 岗位策略建议
   - 这个岗位是主投 / 备选 / 不建议
   - 如果是备选/不建议，推荐 1-2 个更适合的岗位方向
   - 告诉学生下一步投递节奏
```

---

## 8. 必须删除/下沉内容

| 内容 | 处理 |
|------|------|
| 面试追问 | **删除** — 不属于岗位匹配，移入独立面试准备模块 |
| Agent Trace | **删除** — 调试信息，不对用户展示 |
| quality_checks | **删除** — 技术审计，不对用户展示 |
| 完整 coverage 表 | **删除主展示** — 后端仍需计算但不默认展示 |
| 完整 evidence 列表 | **删除主展示** — 证据融入命中点/风险/改写指令的自然语言 |
| generated_by/report_version | **删除** — 不对用户展示 |
| score_breakdown 5/6维 | **下沉** — 折叠或仅内部使用 |
| score_alignment | **下沉** — 折叠或仅内部使用 |
| hard_gate | **下沉** — 融入投递决策理由，不单独展示 |

---

## 9. 后续重构计划

| 阶段 | 目标 | 后端 | 前端 | Prompt | 验收标准 |
|------|------|:---:|:---:|:---:|------|
| **JM-R1** | 新 Schema 冻结（Fast/Deep 输出契约 + 页面模块） | ❌ | ❌ | ❌ | 冻结 Fast 3 字段 + Deep 5 模块输出 Schema |
| **JM-R2** | Deep 多模块 Hy3 Prompt 设计 | ❌ | ❌ | ✅ | 每模块有独立 prompt，输出稳定结构化 JSON |
| **JM-R3** | Deep 多模块后端生成链路实现 | ✅ | ❌ | ✅ | Deep API 输出 5 模块 LLM 生成内容 |
| **JM-R4** | 岗位匹配页面重新设计 | ❌ | ✅ | ❌ | 页面聚焦 5 模块，Fast 极简 |
| **JM-R5** | 经典案例内容校准 | ✅ | ✅ | ✅ | 5-10 组真实 case 内容质量验证 |
| **JM-R6** | 最终演示验收 | ❌ | ❌ | ❌ | 全链路可用，评委可演示 |

---

## 10. docs

- **新增：** `docs/09-开发日志/2026-06-14-Job-Match-Core-Rebuild-Baseline-Audit-JM-R0.md`
- **更新：** `docs/变更记录.md`
- **更新：** `docs/下一步任务.md`

---

## 11. 给用户的结论

1. 当前岗位匹配功能虽然在工程上完成了 H18-H20 的全链路重构（Fast SEP + Deep v3 pipeline），但在产品层面存在致命问题：**价值表达错误 — 把"字段完整"当成了"产品有用"。**
2. Deep v3 的 8-layer pipeline 是确定性规则引擎，没有 LLM 调用 — 导致内容浅、模板化、无法给出学生真正需要的"具体改写指令"和"岗位策略建议"。
3. 当前页面虽已做了 H20-B-2 的产品化布局，但内容本身没有变 — 仍然以 coverage/evidence/risk 等后端字段为展示单元，而非以"初筛命中率"为核心叙事。
4. 面试追问、Agent Trace、quality_checks 等调试/跨场景内容持续污染模块边界。
5. 新方向将岗位匹配重新定义为"初筛通过率诊断 + 简历命中率提升方案" — Fast 只做秒级初筛判断（3 字段），Deep 聚焦 5 个核心模块（投递决策/命中点/淘汰风险/改写指令/岗位策略）。
6. Deep 必须从规则引擎转向 Hy3 LLM 驱动 — 只有 LLM 才能产出自然、具体、可执行的改写指令和策略建议。
7. JM-R1 应该冻结新输出 Schema — Fast 精简为 3 字段，Deep 定义为 5 模块 LLM 输出。
8. JM-R2 是核心：为 Deep 5 模块分别设计 Hy3 Prompt，确保每模块输出稳定结构化 JSON。
9. 本次重构的最终目标是：让学生看完岗位匹配后真正知道该怎么改简历、该不该投、如果不投该换什么方向。
10. 不进入 JM-R1。
