# Agent 输出到前端模块映射方案

> 创建时间：2026-05-31
> 所属阶段：4-3.9.1
> 设计人：CodeBuddy
> 依据文档：`docs/07-前端设计/三类Agent输出字段与前端现状核验.md`

## 1. 背景

阶段 4-3.9.0 已完成三类 Agent 输出字段与前端现状核验，确认：

- 三类 Agent 后端能力已基本完成（LLM + fallback）
- 当前前端为 100% 纯 Mock 静态 Demo
- 无 API 对接层、无真实路由、无简历优化页面
- Mock 数据枚举与 Agent 输出不一致
- ResumeOptimization Go Model 与 Schema 存在显著差距

本阶段目标：**设计三类 Agent 输出字段如何被拆成前端可展示、可交互、可理解的产品化模块**。

本阶段为 docs only，不做前端实现。

## 2. 设计目标

1. **不展示一段 AI 文本**——把 Agent 返回内容拆成卡片、表格、任务清单、看板、核实清单等产品化模块
2. **三 Agent 形成闭环**——岗位推荐 → 岗位匹配分析 → 简历优化 → 投递前确认
3. **统一展示语言**——三类 Agent 共享字段使用统一前端名称
4. **数据源优先级明确**——Go Model / 接口实际返回优先，Schema 理想字段后置
5. **P0/P1/P2 分层**——先做核心决策模块，再做增强模块

## 3. 数据源优先级

本方案严格遵守以下数据源优先级：

```
真实接口返回结构 / Go Model  >  接口文档  >  Schema  >  Prompt 期望
```

**核心原则**：

- 如果字段在 Go Model 中存在且接口实际返回 → **可作为 P0**
- 如果字段仅在 Schema 中定义但 Go Model 尚未实现 → **只能列为 P1/P2**
- 如果字段仅在 Prompt 中期望 → **列为 P2 后续增强**

### 3.1 各 Agent 数据源可用性

| Agent | Go Model 字段数 | Schema 字段数 | 差距 |
|-------|----------------|--------------|------|
| MatchAnalysisAgent | 21 | 20+ | 基本对齐，Go Model 覆盖 Schema 核心字段 |
| RecommendationAgent | 18 | 20+ | 基本对齐，Go Model 覆盖 Schema 核心字段 |
| ResumeOptimizationAgent | 10 | 18+ | **差距显著**，Go Model 仅有 Mock 字段 |

### 3.2 ResumeOptimization 字段分级

**P0 真实可用字段**（Go Model 已实现，接口已返回）：

| 字段 | JSON tag | 类型 |
|------|----------|------|
| apply_readiness | apply_readiness | string |
| optimization_stop_signal | optimization_stop_signal | string |
| top_fix_items | top_fix_items | []ResumeOptimizationFixItem |
| summary | summary | string |
| long_term_improvement | long_term_improvement | []ResumeOptimizationLongTermItem |
| jd_evidence_chain | jd_evidence_chain | []ResumeOptimizationJDEvidenceItem |
| verification_notice | verification_notice | string |
| generated_by | generated_by | string |
| agent_status | agent_status | string |
| fallback_required | fallback_required | bool |
| repair_applied | repair_applied | bool |

**P1/P2 后续增强字段**（仅存在于 Schema，Go Model 未实现）：

| 字段 | Schema 中存在 | Go Model 中存在 | 优先级 |
|------|-------------|----------------|--------|
| screening_focus | ✅ | ❌ | P1 |
| priority_suggestions | ✅ | ❌ | P1 |
| keyword_alignment | ✅ | ❌ | P1 |
| project_optimization | ✅ | ❌ | P1 |
| skill_section_optimization | ✅ | ❌ | P2 |
| summary_optimization | ✅ | ❌ | P2 |
| missing_evidence_plan | ✅ | ❌ | P1 |
| rewrite_examples | ✅ | ❌ | P1 |
| offer_go_summary | ✅ | ❌ | P1 |
| next_actions | ✅ | ❌ | P1 |
| authenticity_notice | ✅ | ❌ | P1 |
| target_job | ✅ | ❌ | P1（请求体有，响应体无） |

## 4. 三 Agent 求职决策闭环

### 4.1 闭环设计

三类 Agent 不是三个散功能，而是一个**求职决策闭环**：

```
┌─────────────────────────────────────────────────────┐
│                   求职决策闭环                         │
│                                                       │
│  Step 1: 岗位推荐                                     │
│  "先投哪个岗位？"                                      │
│  → RecommendationAgent                               │
│                                                       │
│  Step 2: 岗位匹配分析                                  │
│  "这个岗位现在能不能投？为什么？"                        │
│  → MatchAnalysisAgent                                │
│                                                       │
│  Step 3: 简历优化                                     │
│  "简历怎么改？改到哪里可以停？"                          │
│  → ResumeOptimizationAgent                           │
│                                                       │
│  Step 4: 投递前确认                                    │
│  "修复项确认了吗？真实性确认了吗？可以投了吗？"            │
│  → 三类 Agent 共享字段 + 用户确认状态                    │
│                                                       │
└─────────────────────────────────────────────────────┘
```

### 4.2 闭环流转路径

| Step | 入口 | Agent 输出 | 关键决策 | 下一步 |
|------|------|-----------|---------|--------|
| 1 | 岗位推荐看板 | recommendations + recommend_level + apply_readiness | 先投哪个 | 点击岗位进入 Step 2 |
| 2 | 岗位匹配分析报告 | match_score + evidence_chain + apply_readiness + optimization_stop_signal | 能不能投 | 需要优化进入 Step 3，可直接投进入 Step 4 |
| 3 | 简历优化工作台 | summary + top_fix_items + jd_evidence_chain + verification_notice | 怎么改 | 修复后重新 Step 2 或进入 Step 4 |
| 4 | 投递前确认清单 | apply_readiness + top_fix_items 完成状态 + verification_notice | 是否投递 | 投递或返回 Step 3 |

### 4.3 Agent 在闭环中的角色

| Agent | 闭环角色 | 核心问题 | 核心输出 |
|-------|---------|---------|---------|
| RecommendationAgent | 方向选择 | 先投哪个？ | 岗位推荐看板 + 推荐等级 + 证据链 |
| MatchAnalysisAgent | 决策判断 | 能不能投？ | 匹配分数 + 证据链 + 投递准备度 + 面试追问 |
| ResumeOptimizationAgent | 行动指导 | 怎么改？ | Top Fix 任务 + JD 证据链 + 核实清单 |

## 5. 页面级信息架构

### 5.1 页面总览

| 页面 | 对应 Agent | 核心模块 | 当前是否已有 | 后续动作 |
|------|-----------|---------|------------|---------|
| 岗位推荐看板 | RecommendationAgent | 岗位卡片组 + 推荐分组 + 证据链 + Top Fix | ⚠️ RecommendPage 部分承接 | 升级为推荐看板 |
| 岗位匹配分析报告 | MatchAnalysisAgent | 匹配总览 + 证据链表格 + 风险卡 + Top Fix + 面试追问 | ⚠️ MatchPage 部分承接 | 升级增加战略字段模块 |
| 简历优化工作台 | ResumeOptimizationAgent | 优化总览 + Top Fix + JD 证据链 + 核实清单 | ❌ 完全缺失 | **新增页面** |
| 投递前确认清单 | 三类 Agent 共享 | 准备度 + 修复确认 + 真实性确认 | ❌ 完全缺失 | **新增页面/区域** |
| 求职决策工作台 | 三类 Agent 共享 | 流程入口 + 最近决策 + 当前状态 | ❌ 完全缺失 | **新增页面/区域** |

### 5.2 岗位推荐看板

**对应 Agent**：RecommendationAgent

**页面布局**：

```
┌────────────────────────────────────────────────┐
│  岗位推荐看板                                    │
│                                                  │
│  ┌─ 策略摘要 ──────────────────────────────┐    │
│  │ strategy_summary (OfferGo 语气)          │    │
│  └──────────────────────────────────────────┘    │
│                                                  │
│  ┌─ 推荐分组看板 ───────────────────────────┐    │
│  │                                          │    │
│  │  [优先投递]  [建议尝试]  [先了解]  [暂不建议] │    │
│  │    卡片1      卡片2     卡片3     卡片4     │    │
│  │                                          │    │
│  └──────────────────────────────────────────┘    │
│                                                  │
│  ┌─ 整体风险 ──────────────────────────────┐    │
│  │ overall_risks                            │    │
│  └──────────────────────────────────────────┘    │
│                                                  │
│  ┌─ 简历优化焦点 ──────────────────────────┐    │
│  │ resume_focus                             │    │
│  └──────────────────────────────────────────┘    │
└────────────────────────────────────────────────┘
```

**核心模块**：

| 模块 | 展示形态 | 字段来源 | 优先级 |
|------|---------|---------|--------|
| 策略摘要卡片 | 文本卡 | strategy_summary | P0 |
| 推荐分组看板 | 四列看板 + 卡片 | recommendations (recommend_level 分组) | P0 |
| 岗位卡片 | 卡片 + 分数环 + 标签 | title/department/city/match_score/recommend_label | P0 |
| 投递准备度标签 | 标签/徽章 | apply_readiness | P0 |
| 优化停止信号标签 | 标签/徽章 | optimization_stop_signal | P0 |
| Top 3 快速修复 | 折叠任务卡 | top_fix_items | P0 |
| 岗位证据链 | 可展开表格 | evidence | P0 |
| 岗位风险 | 风险标签组 | risk_points | P1 |
| 搜索建议 | 链接组 | query_suggestions | P1 |
| 下一步行动 | 行动卡 | next_action | P1 |
| 多岗位对比 | 横向对比表 | 多个 recommendation_item | P2 |
| 整体风险 | 风险卡组 | overall_risks | P1 |
| 简历优化焦点 | 焦点列表 | resume_focus | P1 |

### 5.3 岗位匹配分析报告

**对应 Agent**：MatchAnalysisAgent

**页面布局**：

```
┌────────────────────────────────────────────────┐
│  岗位匹配分析报告                                │
│                                                  │
│  ┌─ 匹配总览 ──────────────────────────────┐    │
│  │ 匹配分数(环) + 匹配等级标签 + 初筛判断    │    │
│  │ 投递准备度 + 优化停止信号                  │    │
│  └──────────────────────────────────────────┘    │
│                                                  │
│  ┌─ OfferGo 摘要 ──────────────────────────┐    │
│  │ offer_go_summary                         │    │
│  └──────────────────────────────────────────┘    │
│                                                  │
│  ┌─ Top 3 快速修复 ────────────────────────┐    │
│  │ top_fix_items (任务卡 / checklist)       │    │
│  └──────────────────────────────────────────┘    │
│                                                  │
│  ┌─ JD 证据链表格 ─────────────────────────┐    │
│  │ JD 要求 | 简历证据 | 判断 | 风险 | 建议  │    │
│  └──────────────────────────────────────────┘    │
│                                                  │
│  ┌─ 核心优势 / 主要差距 ───────────────────┐    │
│  │ matched_points / gap_points              │    │
│  └──────────────────────────────────────────┘    │
│                                                  │
│  ┌─ 风险提示 ──────────────────────────────┐    │
│  │ risk_points (带严重度)                    │    │
│  └──────────────────────────────────────────┘    │
│                                                  │
│  ┌─ 优化建议 ──────────────────────────────┐    │
│  │ optimization_suggestions (P0/P1/P2)      │    │
│  └──────────────────────────────────────────┘    │
│                                                  │
│  ┌─ 面试追问 ──────────────────────────────┐    │
│  │ interview_questions (手风琴/折叠)         │    │
│  └──────────────────────────────────────────┘    │
│                                                  │
│  ┌─ 真实性保护 + 下一步行动 ────────────────┐    │
│  │ authenticity_notice + next_actions       │    │
│  └──────────────────────────────────────────┘    │
└────────────────────────────────────────────────┘
```

**核心模块**：

| 模块 | 展示形态 | 字段来源 | 优先级 |
|------|---------|---------|--------|
| 匹配分数环 | ScoreRing | match_score | P0 |
| 匹配等级标签 | 标签 | match_level + match_label | P0 |
| 初筛判断标签 | 标签 | screening_result | P0 |
| 投递准备度卡片 | 卡片 | apply_readiness | P0 |
| 优化停止信号卡片 | 卡片 | optimization_stop_signal | P0 |
| Top 3 快速修复任务 | 任务卡 / checklist | top_fix_items | P0 |
| OfferGo 摘要 | 文本卡 | offer_go_summary | P0 |
| JD 证据链表格 | 表格 | evidence_chain | P0 |
| 核心优势 | 列表 | matched_points / match_highlights | P1 |
| 主要差距 | 列表 | gap_points | P1 |
| 风险提示卡 | 卡片组 | risk_points | P1 |
| 优化建议 | P0/P1/P2 分组 | optimization_suggestions | P1 |
| 面试追问 | 折叠面板 | interview_questions | P2 |
| 真实性保护 | 提示卡 | authenticity_notice | P1 |
| 下一步行动 | 行动列表 | next_steps / next_actions | P1 |
| 结果来源状态 | 标签 | generated_by + agent_status | P0 |

### 5.4 简历优化工作台

**对应 Agent**：ResumeOptimizationAgent

**页面布局**：

```
┌────────────────────────────────────────────────┐
│  简历优化工作台                                  │
│                                                  │
│  ┌─ 优化总览 ──────────────────────────────┐    │
│  │ summary + 投递准备度 + 优化停止信号       │    │
│  │ 结果来源状态标签                          │    │
│  └──────────────────────────────────────────┘    │
│                                                  │
│  ┌─ Top 3 快速修复 ────────────────────────┐    │
│  │ top_fix_items (任务卡 / checklist)       │    │
│  │ [修复完成] [跳过] [核实] 按钮             │    │
│  └──────────────────────────────────────────┘    │
│                                                  │
│  ┌─ JD 证据链表格 ─────────────────────────┐    │
│  │ JD 要求 | 简历证据 | 强度 | 判断 | 风险  │    │
│  └──────────────────────────────────────────┘    │
│                                                  │
│  ┌─ 长期提升方向 ──────────────────────────┐    │
│  │ long_term_improvement (折叠)             │    │
│  └──────────────────────────────────────────┘    │
│                                                  │
│  ┌─ 用户核实清单 ──────────────────────────┐    │
│  │ verification_notice + 核实勾选项          │    │
│  └──────────────────────────────────────────┘    │
│                                                  │
│  ┌─ [P1] 初筛重点 ─────────────────────────┐    │
│  │ screening_focus (Go Model 未实现，P1)    │    │
│  └──────────────────────────────────────────┘    │
│                                                  │
│  ┌─ [P1] 分优先级建议 ─────────────────────┐    │
│  │ priority_suggestions (Go Model 未实现)    │    │
│  └──────────────────────────────────────────┘    │
│                                                  │
│  ┌─ [P1] 关键词对齐 ───────────────────────┐    │
│  │ keyword_alignment (Go Model 未实现)       │    │
│  └──────────────────────────────────────────┘    │
│                                                  │
│  ┌─ [P1] 项目经历优化 ─────────────────────┐    │
│  │ project_optimization (Go Model 未实现)    │    │
│  └──────────────────────────────────────────┘    │
│                                                  │
│  ┌─ [P1] 示例参考 ─────────────────────────┐    │
│  │ rewrite_examples (Go Model 未实现)        │    │
│  └──────────────────────────────────────────┘    │
└────────────────────────────────────────────────┘
```

**核心模块**：

| 模块 | 展示形态 | 字段来源 | 优先级 | 数据源 |
|------|---------|---------|--------|--------|
| 优化总览 | 文本卡 + 标签 | summary + apply_readiness + optimization_stop_signal | P0 | ✅ Go Model |
| Top 3 快速修复任务 | 任务卡 / checklist | top_fix_items | P0 | ✅ Go Model |
| JD 证据链表格 | 表格 | jd_evidence_chain | P0 | ✅ Go Model |
| 长期提升方向 | 折叠列表 | long_term_improvement | P0 | ✅ Go Model |
| 用户核实清单 | 核实卡 + 勾选 | verification_notice | P0 | ✅ Go Model |
| 结果来源状态 | 标签 | generated_by + agent_status | P0 | ✅ Go Model |
| 初筛重点 | 列表 | screening_focus | **P1** | ❌ 仅 Schema |
| 分优先级建议 | P0/P1/P2 分组 | priority_suggestions | **P1** | ❌ 仅 Schema |
| 关键词对齐 | 对齐表 | keyword_alignment | **P1** | ❌ 仅 Schema |
| 项目经历优化 | 手风琴 | project_optimization | **P1** | ❌ 仅 Schema |
| 技能栏优化 | 列表 | skill_section_optimization | **P2** | ❌ 仅 Schema |
| 摘要优化 | 列表 | summary_optimization | **P2** | ❌ 仅 Schema |
| 示例参考卡片 | 前后对比卡 | rewrite_examples | **P1** | ❌ 仅 Schema |
| 缺失证据计划 | 行动列表 | missing_evidence_plan | **P1** | ❌ 仅 Schema |
| OfferGo 摘要 | 文本卡 | offer_go_summary | **P1** | ❌ 仅 Schema |
| 下一步行动 | 行动列表 | next_actions | **P1** | ❌ 仅 Schema |
| 真实性提醒 | 提示卡 | authenticity_notice | **P1** | ❌ 仅 Schema |

### 5.5 投递前确认清单

**对应 Agent**：三类 Agent 共享字段 + 用户确认状态

**页面布局**：

```
┌────────────────────────────────────────────────┐
│  投递前确认清单                                  │
│                                                  │
│  ┌─ 投递准备度 ────────────────────────────┐    │
│  │ apply_readiness → 可视化状态              │    │
│  └──────────────────────────────────────────┘    │
│                                                  │
│  ┌─ Top Fix 完成确认 ──────────────────────┐    │
│  │ ☐ 修复项 1 完成                           │    │
│  │ ☐ 修复项 2 完成                           │    │
│  │ ☐ 修复项 3 完成                           │    │
│  └──────────────────────────────────────────┘    │
│                                                  │
│  ┌─ 真实性确认 ────────────────────────────┐    │
│  │ ☐ 示例内容已核实                          │    │
│  │ ☐ 不存在夸大经历                          │    │
│  │ ☐ 技术是真实使用过的                      │    │
│  │ ☐ 项目职责是真实承担的                    │    │
│  │ ☐ 数字结果有依据                          │    │
│  └──────────────────────────────────────────┘    │
│                                                  │
│  ┌─ 面试准备确认 ──────────────────────────┐    │
│  │ ☐ 能回答面试追问                          │    │
│  │ ☐ 能解释简历中的技术细节                  │    │
│  └──────────────────────────────────────────┘    │
│                                                  │
│  ┌─ 投递决策 ──────────────────────────────┐    │
│  │ [可以投递] / [再优化一下] / [暂不投递]    │    │
│  └──────────────────────────────────────────┘    │
└────────────────────────────────────────────────┘
```

### 5.6 求职决策工作台

**对应 Agent**：三类 Agent 共享

**页面布局**：

```
┌────────────────────────────────────────────────┐
│  求职决策工作台                                  │
│                                                  │
│  ┌─ 求职流程入口 ──────────────────────────┐    │
│  │ [岗位推荐] → [匹配分析] → [简历优化]     │    │
│  │                                    → [投递] │    │
│  └──────────────────────────────────────────┘    │
│                                                  │
│  ┌─ 最近决策 ──────────────────────────────┐    │
│  │ 最近一次推荐/分析/优化结果摘要            │    │
│  └──────────────────────────────────────────┘    │
│                                                  │
│  ┌─ 当前状态 ──────────────────────────────┐    │
│  │ 推荐状态 / 分析状态 / 优化状态            │    │
│  └──────────────────────────────────────────┘    │
└────────────────────────────────────────────────┘
```

## 6. 三类 Agent 共享字段统一展示

以下字段三类 Agent 都要统一展示，使用统一的前端名称和展示形态：

| 字段 | 统一前端名称 | 展示形态 | 枚举映射 |
|------|------------|---------|---------|
| apply_readiness | 投递准备度 | 卡片/标签 | ready_to_apply → "可以投递" / quick_fix_before_apply → "先修几项再投" / not_ready_for_this_role → "暂不建议主投" / insufficient_information → "信息不足" |
| optimization_stop_signal | 优化停止信号 | 卡片/标签 | stop_and_apply → "停止优化，开始投递" / fix_top_3_then_apply → "修完 Top 3 再投" / needs_more_evidence → "需要补真实证据" / target_role_mismatch → "目标岗位不匹配" |
| top_fix_items | Top 3 快速修复任务 | 任务卡 / checklist | 每项含 issue/why_it_matters/action/example_reference/self_check_question/estimated_effort |
| generated_by | 结果来源 | 小标签 | llm_agent → "AI 分析" / mock_agent → "规则分析" / llm_agent_fallback → "AI 降级分析" |
| agent_status | 兜底状态 | 小标签 | success → 不显示 / mock_success → "规则模式" / repair_applied → "修复后输出" / fallback → "降级输出" |

### 6.1 投递准备度统一展示规范

| 枚举值 | 前端标签 | 颜色 | 图标 | 说明文案 |
|--------|---------|------|------|---------|
| ready_to_apply | 可以投递 | 🟢 绿色 | ✓ | 核心能力和证据已经足够，建议投递 |
| quick_fix_before_apply | 先修几项再投 | 🟡 黄色 | ⚡ | 有匹配基础，完成快速修复后投递 |
| not_ready_for_this_role | 暂不建议主投 | 🔴 红色 | ✗ | 与岗位关键要求差距明显 |
| insufficient_information | 信息不足 | ⚪ 灰色 | ? | 信息不够，无法稳定判断 |

### 6.2 优化停止信号统一展示规范

| 枚举值 | 前端标签 | 颜色 | 说明文案 |
|--------|---------|------|---------|
| stop_and_apply | 可以投了 | 🟢 | 简历已够用，不要再过度优化 |
| fix_top_3_then_apply | 修完再投 | 🟡 | 完成 Top 3 修复项即可投递 |
| needs_more_evidence | 需要补证据 | 🔴 | 问题不是表达能解决的，需要真实经历 |
| target_role_mismatch | 方向不匹配 | 🔴 | 目标岗位与简历证据不匹配 |

### 6.3 结果来源状态标签展示规范

| generated_by | agent_status | 前端展示 | 优先级 |
|-------------|-------------|---------|--------|
| llm_agent | success | "AI 分析" (无特殊标记) | P0 |
| llm_agent | repair_applied | "AI 分析（已修复）" | P0 |
| mock_agent | mock_success | "规则分析" (降级提示) | P0 |
| llm_agent (fallback) | fallback | "AI 降级分析" | P0 |

前端应该在页面底部或结果来源处显示：**"AI 建议仅供参考，请基于真实经历核实后使用"**

## 7. RecommendationAgent 字段到模块映射

| 字段 | UI 模块 | 展示形态 | 当前前端承接 | 优先级 | 数据源 |
|------|--------|---------|------------|--------|--------|
| recommendations | 岗位推荐看板 | 四列看板 + 卡片 | ⚠️ 有 RecommendCard 列表 | P0 | ✅ Go Model |
| recommend_level | 推荐等级标签/分组 | 看板分组 + 卡片标签 | ⚠️ 有 level，枚举不一致 | P0 | ✅ Go Model (priority) |
| recommend_label | 用户可见标签 | 卡片标签 | ❌ 缺失 | P0 | ✅ Go Model |
| match_score | 分数环 | ScoreRing | ⚠️ 有 score，字段名不一致 | P0 | ✅ Go Model (recommendation_score) |
| apply_readiness | 投递准备度卡片 | 卡片/标签 | ❌ 缺失 | P0 | ✅ Go Model |
| optimization_stop_signal | 优化停止信号卡片 | 卡片/标签 | ❌ 缺失 | P0 | ✅ Go Model |
| top_fix_items | Top 3 快速修复任务 | 任务卡 / checklist | ❌ 缺失 | P0 | ✅ Go Model |
| offer_go_summary | OfferGo 摘要卡 | 文本卡 | ❌ 缺失 | P0 | ✅ Go Model |
| recommend_reason | 推荐理由 | 文本/卡片 | ⚠️ 有 reasons | P0 | ✅ Go Model (recommendation_reason) |
| evidence | 岗位证据链 | 可展开表格 | ⚠️ 有简化版 reasons/shortfalls | P0 | ⚠️ Schema 有，Go Model 需映射 |
| risk_points | 风险提示 | 风险标签组 | ⚠️ 有 risk，枚举不一致 | P1 | ✅ Go Model |
| query_suggestions | 搜索建议 | 链接组 | ❌ 缺失 | P1 | ✅ Go Model |
| next_action | 下一步行动 | 行动卡 | ❌ 缺失 | P1 | ✅ Go Model (suggested_action) |
| strategy_summary | 策略摘要卡 | 文本卡 | ❌ 缺失 | P0 | ⚠️ Schema 有，Go Model 需映射 |
| overall_risks | 整体风险 | 风险卡组 | ❌ 缺失 | P1 | ⚠️ Schema 有，Go Model 需映射 |
| resume_focus | 简历优化焦点 | 焦点列表 | ❌ 缺失 | P1 | ⚠️ Schema 有，Go Model 需映射 |
| generated_by | 结果来源标签 | 小标签 | ❌ 缺失 | P0 | ✅ Go Model |
| agent_status | 兜底状态标签 | 小标签 | ❌ 缺失 | P0 | ✅ Go Model |

**Go Model 与 Schema 字段名映射**：

| Schema 字段 | Go Model 字段 (JSON tag) | 差距 |
|------------|------------------------|------|
| recommend_level | priority | 字段名不同，需映射 |
| match_score | recommendation_score | 字段名不同，需映射 |
| recommend_reason | recommendation_reason | 基本一致 |
| next_action | suggested_action | 结构不同：Schema 是 object，Go Model 是 string |
| evidence | 无直接字段 | Schema 有 evidence_item，Go Model 无独立 evidence |
| strategy_summary | 无 | Go Model 无此字段 |
| overall_risks | 无 | Go Model 无此字段 |
| resume_focus | 无 | Go Model 无此字段 |

## 8. MatchAnalysisAgent 字段到模块映射

| 字段 | UI 模块 | 展示形态 | 当前前端承接 | 优先级 | 数据源 |
|------|--------|---------|------------|--------|--------|
| match_score | 匹配分数环 | ScoreRing | ✅ 有 ScoreRing | P0 | ✅ Go Model |
| match_level | 匹配等级标签 | 标签 | ❌ 缺失 | P0 | ✅ Go Model |
| match_label | 匹配用户标签 | 标签 | ❌ 缺失 | P0 | ✅ Go Model |
| screening_result | 初筛判断标签 | 标签 | ❌ 缺失 | P0 | ✅ Go Model |
| apply_readiness | 投递准备度卡片 | 卡片/标签 | ❌ 缺失 | P0 | ✅ Go Model |
| optimization_stop_signal | 优化停止信号卡片 | 卡片/标签 | ❌ 缺失 | P0 | ✅ Go Model |
| top_fix_items | Top 3 快速修复任务 | 任务卡 / checklist | ❌ 缺失 | P0 | ✅ Go Model |
| offer_go_summary | OfferGo 摘要卡 | 文本卡 | ❌ 缺失 | P0 | ✅ Go Model |
| jd_key_requirements | JD 关键要求 | 列表 | ❌ 缺失 | P1 | ⚠️ Schema 有，Go Model 需映射 |
| resume_evidence_summary | 证据摘要 | 列表 | ❌ 缺失 | P1 | ⚠️ Schema 有，Go Model 需映射 |
| evidence_chain | JD 证据链表格 | 表格 | ⚠️ 有 EvidenceTable，结构不一致 | P0 | ✅ Go Model |
| matched_points / match_highlights | 核心优势 | 列表 | ⚠️ 有优势列表，结构不一致 | P1 | ✅ Go Model (match_highlights) |
| gap_points | 主要差距 | 列表 | ❌ 缺失 | P1 | ✅ Go Model |
| risk_points | 风险提示卡 | 卡片组 | ⚠️ 有风险列表，结构不一致 | P1 | ✅ Go Model (string[]) |
| optimization_suggestions | 优化建议 | P0/P1/P2 分组 | ⚠️ 有 P0/P1 建议，结构不一致 | P1 | ✅ Go Model |
| interview_questions | 面试追问 | 折叠面板 | ❌ 缺失 | P2 | ✅ Go Model (string[]) |
| authenticity_notice | 真实性保护 | 提示卡 | ⚠️ 有 AuthenticityCard | P1 | ✅ Go Model |
| next_steps / next_actions | 下一步行动 | 行动列表 | ⚠️ 有 NextStepsCard | P1 | ✅ Go Model (next_steps) |
| generated_by | 结果来源标签 | 小标签 | ❌ 缺失 | P0 | ✅ Go Model |
| agent_status | 兜底状态标签 | 小标签 | ❌ 缺失 | P0 | ✅ Go Model |

**Go Model 与 Schema 结构差异**：

| 字段 | Go Model 类型 | Schema 类型 | 差距 |
|------|-------------|------------|------|
| evidence_chain | []MatchEvidence (requirement/resume_point/judgement) | []evidence_chain_item (jd_requirement/resume_evidence/judgement/explanation) | 字段名和结构不一致 |
| matched_points | []string (match_highlights) | []matched_point (point/why_it_matters/evidence) | Go Model 是简单字符串，Schema 是结构体 |
| gap_points | []string | []gap_point (gap/impact/how_to_improve) | Go Model 是简单字符串，Schema 是结构体 |
| risk_points | []string | []risk_point (risk/severity/reason/mitigation) | Go Model 是简单字符串，Schema 是结构体 |
| interview_questions | []string | []interview_question (question/why_ask/how_to_prepare/related_jd_requirement) | Go Model 是简单字符串，Schema 是结构体 |

**结论**：MatchAnalysis Go Model 的核心字段已对齐 Schema，但部分字段（matched_points/gap_points/risk_points/interview_questions）在 Go Model 中为简单字符串数组，而 Schema 定义为结构体数组。前端 P0 模块应基于 Go Model 实际返回结构设计，结构体字段扩展列为 P1/P2。

## 9. ResumeOptimizationAgent 字段到模块映射

### 9.1 P0 真实可用字段映射

| 字段 | UI 模块 | 展示形态 | 当前前端承接 | 数据源 |
|------|--------|---------|------------|--------|
| summary | 优化总览卡 | 文本卡 | ❌ 无页面 | ✅ Go Model |
| apply_readiness | 投递准备度卡片 | 卡片/标签 | ❌ | ✅ Go Model |
| optimization_stop_signal | 优化停止信号卡片 | 卡片/标签 | ❌ | ✅ Go Model |
| top_fix_items | Top 3 快速修复任务 | 任务卡 / checklist | ❌ | ✅ Go Model |
| long_term_improvement | 长期提升方向 | 折叠列表 | ❌ | ✅ Go Model |
| jd_evidence_chain | JD 证据链表格 | 表格 | ❌ | ✅ Go Model |
| verification_notice | 用户核实清单 | 核实卡 | ❌ | ✅ Go Model |
| generated_by | 结果来源标签 | 小标签 | ❌ | ✅ Go Model |
| agent_status | 兜底状态标签 | 小标签 | ❌ | ✅ Go Model |

### 9.2 P1 后续增强字段映射（Go Model 未实现）

| 字段 | UI 模块 | 展示形态 | Schema 定义 | Go Model |
|------|--------|---------|------------|----------|
| screening_focus | 初筛重点 | 列表 | ✅ | ❌ |
| priority_suggestions | 分优先级建议 | P0/P1/P2 分组 | ✅ | ❌ |
| keyword_alignment | 关键词对齐表 | 对齐表格 | ✅ | ❌ |
| project_optimization | 项目经历优化 | 手风琴 | ✅ | ❌ |
| missing_evidence_plan | 缺失证据计划 | 行动列表 | ✅ | ❌ |
| rewrite_examples | 示例参考卡片 | 前后对比卡 | ✅ | ❌ |
| offer_go_summary | OfferGo 摘要 | 文本卡 | ✅ | ❌ |
| next_actions | 下一步行动 | 行动列表 | ✅ | ❌ |
| authenticity_notice | 真实性提醒 | 提示卡 | ✅ | ❌ |

### 9.3 P2 展示增强字段映射（Go Model 未实现）

| 字段 | UI 模块 | 展示形态 | Schema 定义 | Go Model |
|------|--------|---------|------------|----------|
| skill_section_optimization | 技能栏优化 | 列表 | ✅ | ❌ |
| summary_optimization | 摘要优化 | 列表 | ✅ | ❌ |

### 9.4 关键说明

1. **简历优化页面完全缺失**，需要从零新增
2. **P0 实现只能基于 Go Model 实际返回的字段**：summary / apply_readiness / optimization_stop_signal / top_fix_items / long_term_improvement / jd_evidence_chain / verification_notice
3. **Schema 中的 screening_focus / priority_suggestions / keyword_alignment / project_optimization / rewrite_examples 等字段当前 Go Model 不返回**，前端 P0 阶段无法展示这些模块
4. **后续需要后端补齐 ResumeOptimization Go Model**，将 Schema 字段逐步映射到 Go Model，前端才能展示对应模块

## 10. P0 / P1 / P2 模块优先级

### 10.1 P0：必须优先实现

| # | 模块 | 来源 Agent | 来源字段 | 当前前端是否已有类似承载 | 需要动作 |
|---|------|-----------|---------|----------------------|---------|
| 1 | 投递准备度卡片 | 三类 Agent 共有 | apply_readiness | ❌ 无 | 新增组件 |
| 2 | 优化停止信号卡片 | 三类 Agent 共有 | optimization_stop_signal | ❌ 无 | 新增组件 |
| 3 | Top 3 快速修复任务 | 三类 Agent 共有 | top_fix_items | ❌ 无 | 新增组件 |
| 4 | JD 证据链表格 | MatchAnalysis + ResumeOptimization | evidence_chain / jd_evidence_chain | ⚠️ 有简化版 | 升级现有组件 |
| 5 | 岗位推荐看板 | Recommendation | recommendations + recommend_level | ⚠️ 有 RecommendCard | 升级现有页面 |
| 6 | 简历优化工作台基础版 | ResumeOptimization | summary + top_fix_items + jd_evidence_chain + verification_notice | ❌ 完全缺失 | **新增页面** |
| 7 | 用户核实清单基础版 | ResumeOptimization | verification_notice + top_fix_items.self_check_question | ❌ 无 | 新增组件 |
| 8 | 结果来源状态标签 | 三类 Agent 共有 | generated_by + agent_status | ❌ 无 | 新增组件 |

### 10.2 P1：重要增强

| # | 模块 | 来源 Agent | 来源字段 | 说明 |
|---|------|-----------|---------|------|
| 1 | 多岗位对比表 | Recommendation | 多个 recommendation_item 横向对比 | 帮助用户在多个岗位间做决策 |
| 2 | 风险提示卡 | MatchAnalysis + Recommendation | risk_points (结构体版) | 带严重度和缓解措施 |
| 3 | 下一步行动模块 | 三类 Agent 共有 | next_actions / suggested_action | 结构化行动展示 |
| 4 | 示例参考卡片 | ResumeOptimization | rewrite_examples | 前后对比展示（需 Go Model 补齐） |
| 5 | 面试追问模块 | MatchAnalysis | interview_questions (结构体版) | 折叠面板展示 |
| 6 | 简历模块诊断 | ResumeOptimization | priority_suggestions / screening_focus | 需 Go Model 补齐 |
| 7 | 关键词对齐区 | ResumeOptimization | keyword_alignment | 需 Go Model 补齐 |
| 8 | 项目经历优化 | ResumeOptimization | project_optimization | 需 Go Model 补齐 |
| 9 | 缺失证据计划 | ResumeOptimization | missing_evidence_plan | 需 Go Model 补齐 |
| 10 | OfferGo 摘要卡 | MatchAnalysis + ResumeOptimization | offer_go_summary | Recommendation Go Model 有，ResumeOptimization Go Model 无 |
| 11 | 策略摘要卡 | Recommendation | strategy_summary | 需 Go Model 补齐 |
| 12 | 整体风险 | Recommendation | overall_risks | 需 Go Model 补齐 |
| 13 | 简历优化焦点 | Recommendation | resume_focus | 需 Go Model 补齐 |

### 10.3 P2：展示增强

| # | 模块 | 来源 Agent | 来源字段 | 说明 |
|---|------|-----------|---------|------|
| 1 | 雷达图 | MatchAnalysis | match_score + 多维度 | 多维度可视化 |
| 2 | 风险热力标签 | MatchAnalysis + Recommendation | risk_points.severity | 颜色映射 |
| 3 | 投递准备进度条 | 三类 Agent 共有 | apply_readiness | 可视化进度 |
| 4 | 修改前后对比 | ResumeOptimization | rewrite_examples | 需 Go Model 补齐 |
| 5 | 时间成本统计 | 三类 Agent 共有 | top_fix_items.estimated_effort | 汇总预计时间 |
| 6 | 多 Agent 联动时间线 | 三类 Agent 共有 | 分析时间序列 | 展示决策历史 |
| 7 | 技能栏优化 | ResumeOptimization | skill_section_optimization | 需 Go Model 补齐 |
| 8 | 摘要优化 | ResumeOptimization | summary_optimization | 需 Go Model 补齐 |

## 11. 页面与模块关系

| 页面 | 对应 Agent | 核心模块 | 当前是否已有 | 后续动作 |
|------|-----------|---------|------------|---------|
| 岗位推荐看板 | RecommendationAgent | 策略摘要 + 推荐分组 + 岗位卡片 + 投递准备度 + 优化停止信号 + Top Fix + 证据链 + 搜索建议 + 下一步行动 | ⚠️ RecommendPage 部分承接 | 升级为看板，对齐枚举，新增战略字段模块 |
| 岗位匹配分析报告 | MatchAnalysisAgent | 匹配总览 + 分数环 + 等级标签 + 初筛判断 + 投递准备度 + 优化停止信号 + Top Fix + OfferGo 摘要 + 证据链表格 + 优势/差距 + 风险卡 + 优化建议 + 面试追问 + 真实性保护 + 下一步行动 | ⚠️ MatchPage 部分承接 | 升级增加战略字段模块，对齐证据链结构 |
| 简历优化工作台 | ResumeOptimizationAgent | 优化总览 + 投递准备度 + 优化停止信号 + Top Fix + JD 证据链 + 长期提升 + 核实清单 | ❌ 完全缺失 | **新增页面**，P0 基于 Go Model 实际字段 |
| 投递前确认清单 | 三类 Agent 共享 | 投递准备度 + 修复确认 + 真实性确认 + 面试准备确认 + 投递决策 | ❌ 完全缺失 | **新增页面或区域** |
| 求职决策工作台 | 三类 Agent 共享 | 流程入口 + 最近决策 + 当前状态 | ❌ 完全缺失 | **新增页面或区域** |

## 12. 当前前端差距

| # | 差距 | 严重度 | 说明 |
|---|------|--------|------|
| 1 | 无 API 对接层 | 🔴 高 | 前端 100% 硬编码 Mock，无 axios/fetch/request client |
| 2 | 无真实路由 | 🔴 高 | useState 切换，react-router 已装未用 |
| 3 | 无简历优化页面 | 🔴 高 | ResumeOptimizationAgent 完全无法承接 |
| 4 | Mock 数据枚举不一致 | 🟡 中 | 推荐等级/风险等级/证据强度枚举与 Agent 不一致 |
| 5 | 战略字段完全缺失 | 🔴 高 | apply_readiness / optimization_stop_signal / top_fix_items 无前端承载 |
| 6 | 无统一 Agent 结果展示组件 | 🟡 中 | 缺投递准备度/停止信号/Top Fix/结果来源等统一组件 |
| 7 | 无用户核实清单 | 🟡 中 | 缺核实勾选交互 |
| 8 | 缺多 Agent 闭环路径 | 🟡 中 | 无从推荐→分析→优化→确认的闭环流转 |
| 9 | ResumeOptimization Go Model 差距 | 🔴 高 | Schema 18+ 字段，Go Model 仅 10 字段 |
| 10 | MatchAnalysis 结构体字段简化 | 🟡 中 | risk_points/gap_points/interview_questions Go Model 为 string[]，Schema 为结构体 |

## 13. 不在本阶段实现的内容

本阶段为 docs only，以下内容不在本阶段实现：

1. ❌ 不修改前端代码
2. ❌ 不新增前端组件
3. ❌ 不新增 API client
4. ❌ 不对接后端接口
5. ❌ 不修改后端接口
6. ❌ 不修改 Agent
7. ❌ 不修改 Prompt
8. ❌ 不修改 Schema
9. ❌ 不修改 Go Model
10. ❌ 不修改 Mock 数据
11. ❌ 不补齐 ResumeOptimization Go Model 缺失字段
12. ❌ 不升级 MatchAnalysis Go Model 结构体字段

## 14. 是否可以进入 4-3.9.2

**✅ 可以进入。**

理由：

1. 已设计三 Agent 求职决策闭环
2. 已明确数据源优先级
3. 已设计页面级信息架构（5 个页面）
4. 已设计三类 Agent 共享字段统一展示
5. 已完成三类 Agent 字段到模块映射
6. 已区分 ResumeOptimization 真实可用字段与理想字段
7. 已输出 P0/P1/P2 模块优先级
8. 已输出页面与模块关系
9. 已明确当前前端差距
10. 4-3.9.2 前端实现拆分计划的输入已具备

**进入 4-3.9.2 前的前置条件**：

1. 需确认 ResumeOptimization Go Model 补齐计划（哪些字段需要在 4-3.9.2 之前补齐）
2. 需确认前端技术选型（API client 选 axios 还是 fetch，状态管理选什么）
3. 需确认路由方案（react-router 如何接入）
4. 需确认 Mock 数据对齐策略（统一前端枚举还是后端适配）

## 15. 下一步建议

进入 **4-3.9.2 前端实现拆分计划与验收标准冻结**：

1. 拆分前端实现为可独立交付的子任务
2. 确定每个子任务的验收标准
3. 确定 API 对接层架构
4. 确定路由方案
5. 确定 Mock 数据对齐策略
6. 确定 ResumeOptimization Go Model 补齐优先级
7. 冻结验收标准后进入阶段 5 前端实现
