# 2026-05-30-106-PromptSchema战略一致性核验与改动评审

## 1. 本次任务

阶段 4-3.7.C：Prompt / Schema 战略一致性核验与改动评审。

## 2. 背景

项目定位已从"岗位推荐工具"升级为"投递前决策与简历初筛优化系统"，4-3.5.B ~ 4-3.5.G 完成了三个杀手锏和配套设计的战略文档，但 Prompt 和 Schema 尚未同步更新。需要在 4-3.8 前核验一致性。

## 3. 核验范围

- 三类业务 Agent Prompt：match_analysis_v1 / recommendation_v1 / resume_optimization_v1
- 三类业务 Agent Schema：match_analysis_schema.json / recommendation_schema.json / resume_optimization_schema.json
- 对照 4-3.5.B ~ 4-3.5.G 战略设计文档的 10 个维度

## 4. 核验结果

### 已支持的战略能力

| 维度 | 状态 |
|------|------|
| JD 证据链 | ✅ Match/Rec 完整支持，ResumeOptimization 间接支持 |
| 示例参考与用户核实 | ✅ ResumeOptimization 完整支持（can_use_directly / requires_user_verification / verification_note）|
| 不虚构经历 | ✅ 三个 Prompt 均明确要求 |
| 不承诺通过初筛 | ✅ 三个 Prompt 均明确要求 |
| 不替用户做人生决策 | ✅ 三个 Prompt 均明确要求 |
| generated_by / fallback | ✅ Schema 枚举正确，后端 fallback 逻辑清晰 |
| P0/P1/P2 优先级 | ✅ 三个 Agent 均支持 |
| 分析依据 | ✅ 三个 Agent 均有 analysis_basis |

### 未支持的战略能力

| 维度 | 状态 | 严重度 |
|------|------|--------|
| apply_readiness | ❌ 三个 Agent 均缺失 | 🔴 高 |
| optimization_stop_signal | ❌ 三个 Agent 均缺失 | 🔴 高 |
| START / STAR-T 表达规则 | ❌ 三个 Prompt 均未提及 | 🟡 中 |
| Quality Gate 字段 | ❌ 三个 Schema 均不包含 | 🟡 中 |
| 极端情况字段 | ❌ 三个 Schema 均不包含 | 🟡 中 |
| ResumeContextText 输入 | ❌ 三个 Prompt 均使用旧格式 | 🟡 中 |

## 5. 三类 Agent 评审

| Agent | 当前评估 | 结论 |
|-------|----------|------|
| MatchAnalysisAgent | 证据链、P0/P1/P2、真实性规则、改写示例均已完整；缺 apply_readiness 和 optimization_stop_signal | 需要最小修复 |
| RecommendationAgent | 证据链、推荐级别、风险点、查询建议均已完整；缺 apply_readiness 和 optimization_stop_signal | 需要最小修复 |
| ResumeOptimizationAgent | 用户核实机制最完整（verification_note）；缺 apply_readiness、optimization_stop_signal、Pipeline 未接入 | 需要最小修复 |

## 6. 当前缺口

核心缺口 2 个：

1. **apply_readiness**：投递边界判断字段，战略杀手锏一的核心输出，当前无等价替代
2. **optimization_stop_signal**：优化停止信号，三个 Agent 统一缺失，无等价替代

其余缺口（START / Quality Gate / 极端情况 / ResumeContextText）不阻塞 4-3.8，可在后续增强。

## 7. 是否需要 C.1

**需要插入 4-3.7.C.1 Prompt / Schema 最小一致性修复。**

理由：
- apply_readiness 和 optimization_stop_signal 是"投递前决策系统"核心差异化能力
- 缺少这两个字段，4-3.8 无法验证投递边界和优化停止条件
- 修复范围可控：仅新增 2-3 个字段 + 对应 Prompt 指引

C.1 最小修复范围：
- 三个 Schema 新增 `apply_readiness` + `optimization_stop_signal`
- 三个 Prompt 新增投递边界和优化停止条件指引
- ResumeOptimization Schema 新增 `quick_fix_items`

## 8. 是否可以进入 4-3.8

**完成 C.1 后可以进入 4-3.8。**

当前三个 Agent 的 Prompt / Schema 在证据链、优先级、真实性规则、用户核实等维度已非常完整，只有投递边界和优化停止条件两个核心字段缺失。C.1 修复后即可支撑 4-3.8 集成验收。

## 9. 未修改内容

- 未修改 Prompt 文件
- 未修改 Schema 文件
- 未修改业务代码
- 未修改接口
- 未修改前端
- 未修改 Makefile
- 未调用真实 LLM
- 未请求腾讯招聘接口
- 未清空 volume
- 未启动项目
- 未执行 make up / down / restart

## 10. 下一步建议

1. **进入 4-3.7.C.1**：Prompt / Schema 最小一致性修复
2. **C.1 完成后进入 4-3.8**：ResumeOptimizationAgent 接入 Pipeline
3. **后续增强**：START 规则 / Quality Gate / 极端情况 / ResumeContextText
