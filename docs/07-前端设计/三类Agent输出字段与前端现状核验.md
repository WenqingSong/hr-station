# 三类 Agent 输出字段与前端现状核验

> 创建时间：2026-05-31
> 所属阶段：4-3.9.0
> 核验人：CodeBuddy

## 1. 背景

阶段 4-3.8 已完成三类业务 Agent 集成验收收口，三类 Agent 均已具备 LLM + fallback 基础能力，统一具备关键战略字段（`apply_readiness`、`optimization_stop_signal`、`top_fix_items`、`generated_by`、`agent_status`）。

现在要进入前端产品化设计前的核验阶段，目标不是让前端简单展示一段 AI 文本，而是把 Agent 返回的结构化结果拆成多个产品化模块。但在设计映射方案前，必须先查清当前前端真实情况。

本阶段为 read-only + docs，不做前端实现。

## 2. 核验范围

| 核验维度 | 核验对象 |
|----------|----------|
| 三类 Agent 输出字段 | Schema + Go Model + API 响应 |
| 前端目录结构 | web/src/ |
| 前端页面现状 | 8 个页面组件 |
| 前端组件现状 | 自定义组件 + shadcn/ui |
| 前端路由现状 | useState 切换，无真实路由 |
| Mock 数据现状 | 硬编码在页面组件中 |
| API 对接现状 | 无 API client，无 axios/fetch，无 VITE_API_BASE_URL |
| 三类 Agent 前端承接能力 | 逐一评估 |

## 3. 当前 Agent 输出字段总览

### 3.1 MatchAnalysisAgent 输出字段

**Schema 文件**：`server/prompts/schema/match_analysis_schema.json`（VALID）

| 字段 | 类型 | 说明 | 前端候选模块 |
|------|------|------|-------------|
| match_score | integer (0-100) | 辅助匹配分数 | 匹配分数展示 |
| match_level | enum | strong_match/potential_match/explore_first/not_recommended | 匹配等级标签 |
| match_label | enum | 优先投递/建议尝试/先了解再决定/暂不建议 | 用户可见标签 |
| screening_result | enum | likely/uncertain/unlikely | 初筛结果 |
| apply_readiness | enum | ready_to_apply/quick_fix_before_apply/not_ready_for_this_role/insufficient_information | 投递准备度卡片 |
| optimization_stop_signal | enum | stop_and_apply/fix_top_3_then_apply/needs_more_evidence/target_role_mismatch | 优化停止信号卡片 |
| top_fix_items | array (max 3) | 快速修复项 | Top 3 快速修复任务 |
| offer_go_summary | string | OfferGo 语气摘要 | 摘要卡片 |
| jd_key_requirements | array (3-6) | JD 关键要求 | JD 要求列表 |
| resume_evidence_summary | array (2-5) | 简历证据摘要 | 证据摘要 |
| evidence_chain | array (3-6) | JD-简历证据链 | JD 证据链表格 |
| matched_points | array (2-5) | 匹配亮点 | 核心优势 |
| gap_points | array (2-5) | 差距点 | 主要差距 |
| risk_points | array (2-5) | 风险点 | 风险提示卡 |
| resume_optimization | object (P0/P1/P2) | 简历优化建议 | 优化建议 |
| interview_questions | array (3-6) | 面试追问 | 面试追问模块 |
| authenticity_notice | string | 真实性提醒 | 真实性保护 |
| next_actions | array (3-5) | 下一步行动 | 下一步行动模块 |
| analysis_basis | object | 分析依据 | 数据来源说明 |
| generated_by | string | 生成者标识 | 来源标识 |

**Go Model**：`server/internal/model/match.go` → `MatchAnalysisResult`

| 字段 | Go 类型 | JSON tag |
|------|---------|----------|
| MatchScore | int | match_score |
| ScreeningResult | string | screening_result |
| EvidenceChain | []MatchEvidence | evidence_chain |
| MatchHighlights | []string | match_highlights |
| RiskPoints | []string | risk_points |
| OptimizationSuggestions | []MatchSuggestion | optimization_suggestions |
| NextSteps | []string | next_steps |
| AuthenticityNotice | string | authenticity_notice |
| MatchLevel | string | match_level (omitempty) |
| MatchLabel | string | match_label (omitempty) |
| OfferGoSummary | string | offer_go_summary (omitempty) |
| GapPoints | []string | gap_points (omitempty) |
| InterviewQuestions | []string | interview_questions (omitempty) |
| ApplyReadiness | string | apply_readiness (omitempty) |
| OptimizationStopSignal | string | optimization_stop_signal (omitempty) |
| TopFixItems | []TopFixItem | top_fix_items (omitempty) |
| GeneratedBy | string | generated_by (omitempty) |
| AgentStatus | string | agent_status (omitempty) |
| RepairApplied | bool | repair_applied (omitempty) |
| FallbackReason | string | fallback_reason (omitempty) |
| ErrorCode | string | error_code (omitempty) |

### 3.2 RecommendationAgent 输出字段

**Schema 文件**：`server/prompts/schema/recommendation_schema.json`（VALID）

每个推荐岗位（recommendation_item）包含：

| 字段 | 类型 | 说明 | 前端候选模块 |
|------|------|------|-------------|
| job_id | integer | 岗位 ID | 岗位标识 |
| title | string | 岗位标题 | 岗位卡片标题 |
| department | string | 部门 | 岗位卡片副标题 |
| city | string | 城市 | 岗位卡片 |
| category | string | 分类 | 岗位卡片 |
| recommend_level | enum | priority_apply/recommended/explore_first/not_recommended | 推荐等级展示 |
| recommend_label | enum | 优先投递/建议尝试/先了解再决定/暂不建议 | 用户可见标签 |
| match_score | integer (0-100) | 辅助分数 | 分数环 |
| apply_readiness | enum | 投递准备度 | 投递准备度卡片 |
| optimization_stop_signal | enum | 优化停止信号 | 优化停止信号卡片 |
| top_fix_items | array (max 3) | 快速修复项 | Top 3 快速修复任务 |
| offer_go_summary | string | OfferGo 摘要 | 摘要 |
| recommend_reason | string | 推荐理由 | 推荐理由 |
| evidence | array (2-4) | 证据链 | JD 证据链 |
| risk_points | array (1-4) | 风险点 | 风险提示 |
| query_suggestions | array (2-4) | 搜索建议 | 用户搜索建议 |
| next_action | object | 下一步行动 | 下一步行动模块 |

**Go Model**：`server/internal/model/recommendation.go` → `RecommendedJob`

| 字段 | Go 类型 | JSON tag |
|------|---------|----------|
| JobID | uint64 | job_id |
| Title | string | title |
| Department | string | department |
| City | string | city |
| Category | string | category |
| Priority | string | priority |
| RecommendationScore | int | recommendation_score |
| RecommendationReason | string | recommendation_reason |
| MatchHighlights | []string | match_highlights |
| RiskPoints | []string | risk_points |
| SuggestedAction | string | suggested_action |
| RecommendLabel | string | recommend_label (omitempty) |
| OfferGoSummary | string | offer_go_summary (omitempty) |
| QuerySuggestions | []string | query_suggestions (omitempty) |
| ApplyReadiness | string | apply_readiness (omitempty) |
| OptimizationStopSignal | string | optimization_stop_signal (omitempty) |
| TopFixItems | []TopFixItem | top_fix_items (omitempty) |
| GeneratedBy | string | generated_by (omitempty) |
| AgentStatus | string | agent_status (omitempty) |

### 3.3 ResumeOptimizationAgent 输出字段

**Schema 文件**：`server/prompts/schema/resume_optimization_schema.json`（VALID）

| 字段 | 类型 | 说明 | 前端候选模块 |
|------|------|------|-------------|
| optimization_summary | string | 优化总览 | 简历优化工作台 |
| offer_go_summary | string | OfferGo 摘要 | 摘要卡片 |
| target_job | object | 目标岗位 | 目标岗位信息 |
| apply_readiness | enum | 投递准备度 | 投递准备度卡片 |
| optimization_stop_signal | enum | 优化停止信号 | 优化停止信号卡片 |
| top_fix_items | array (max 3) | 快速修复项 | Top 3 快速修复任务 |
| screening_focus | array (1-5) | 初筛重点 | 初筛重点列表 |
| priority_suggestions | object (P0/P1/P2) | 分优先级的优化建议 | 简历优化工作台 |
| keyword_alignment | array (3-8) | 关键词对齐 | 关键词对齐表 |
| project_optimization | array (1-5) | 项目经历优化 | 项目优化模块 |
| skill_section_optimization | array (1-4) | 技能栏优化 | 技能栏优化模块 |
| summary_optimization | array (1-3) | 摘要优化 | 摘要优化模块 |
| missing_evidence_plan | array (0-5) | 缺失证据计划 | 用户核实清单 |
| rewrite_examples | array (2-6) | 重写示例 | 示例参考卡片 |
| next_actions | array (3-6) | 下一步行动 | 下一步行动模块 |
| authenticity_notice | string | 真实性提醒 | 真实性保护 |
| analysis_basis | object | 分析依据 | 数据来源说明 |
| generated_by | string | 生成者标识 | 来源标识 |

**Go Model**：`server/internal/model/resume_optimization.go` → `ResumeOptimizationResult`

| 字段 | Go 类型 | JSON tag |
|------|---------|----------|
| AgentStatus | string | agent_status |
| GeneratedBy | string | generated_by |
| FallbackRequired | bool | fallback_required |
| RepairApplied | bool | repair_applied |
| ApplyReadiness | string | apply_readiness |
| OptimizationStopSignal | string | optimization_stop_signal |
| Summary | string | summary |
| TopFixItems | []ResumeOptimizationFixItem | top_fix_items |
| LongTermImprovement | []ResumeOptimizationLongTermItem | long_term_improvement |
| JDEvidenceChain | []ResumeOptimizationJDEvidenceItem | jd_evidence_chain |
| VerificationNotice | string | verification_notice |
| FallbackReason | string | fallback_reason (omitempty) |
| ErrorCode | string | error_code (omitempty) |

**⚠️ 关键发现**：ResumeOptimizationAgent 的 Go Model 与 Schema 存在显著差距——Go Model 仅有 summary/top_fix_items/long_term_improvement/jd_evidence_chain/verification_notice，缺少 Schema 中定义的 screening_focus/priority_suggestions/keyword_alignment/project_optimization/skill_section_optimization/summary_optimization/missing_evidence_plan/rewrite_examples/offer_go_summary/next_actions/authenticity_notice/analysis_basis 等字段。这是因为当前 Go Model 对应的是 Mock Agent 输出，LLM Agent 输出需要后续映射改造。

## 4. 当前前端目录结构

```
web/src/
├── app/
│   ├── App.tsx                          # 主入口，useState 切换页面
│   ├── components/
│   │   ├── AppNav.tsx                   # 顶部导航
│   │   ├── HomeContent.tsx              # 首页内容
│   │   ├── figma/
│   │   │   └── ImageWithFallback.tsx    # 图片降级
│   │   ├── layout/
│   │   │   └── PageShell.tsx            # 页面壳
│   │   └── ui/
│   │       ├── Button.tsx               # 自定义按钮
│   │       ├── Card.tsx                 # 自定义卡片
│   │       ├── ScoreRing.tsx            # 分数环
│   │       ├── StatusTag.tsx            # 状态标签
│   │       └── accordion.tsx ~ tooltip.tsx  # shadcn/ui 组件（40+）
│   ├── pages/
│   │   ├── AddResumePage.tsx            # 添加简历
│   │   ├── JobsPage.tsx                 # 岗位列表
│   │   ├── LoginPage.tsx                # 登录页
│   │   ├── MatchPage.tsx                # 岗位匹配分析
│   │   ├── RecommendPage.tsx            # 岗位推荐
│   │   ├── ResumeDetailPage.tsx          # 简历详情
│   │   ├── ResumePage.tsx               # 我的简历
│   │   └── VersionsPage.tsx             # 版本迭代
│   └── tokens.ts                        # 设计 token（颜色等）
├── imports/
│   └── image.png
├── main.tsx                             # React 入口
└── styles/
    ├── fonts.css
    ├── globals.css
    ├── index.css
    ├── tailwind.css
    └── theme.css
```

## 5. 当前前端页面现状

| 页面 | 文件 | 与 Agent 的关系 | 当前状态 |
|------|------|----------------|---------|
| 首页 | HomeContent.tsx | 总览入口 | 硬编码 Mock |
| 我的简历 | ResumePage.tsx | 无直接关联 | 硬编码 Mock |
| 添加简历 | AddResumePage.tsx | 无直接关联 | 硬编码 Mock |
| 简历详情 | ResumeDetailPage.tsx | 无直接关联 | 硬编码 Mock |
| **岗位列表** | JobsPage.tsx | 推荐前置 | 硬编码 Mock，有岗位卡片+选择功能 |
| **岗位推荐** | RecommendPage.tsx | **RecommendationAgent** | 硬编码 Mock，有推荐卡片+排名+分数环+推荐等级+风险+推荐理由+短板+详情抽屉 |
| **岗位匹配分析** | MatchPage.tsx | **MatchAnalysisAgent** | 硬编码 Mock，有匹配总览+分数环+核心优势+主要风险+JD-简历证据链表格+真实性保护+P0/P1优化建议+下一步建议 |
| 版本迭代 | VersionsPage.tsx | 无直接关联 | 硬编码 Mock |
| 登录页 | LoginPage.tsx | 无直接关联 | 硬编码 Mock |

**⚠️ 无简历优化页面**：当前前端没有 `ResumeOptimizationPage` 或类似页面，无法承接 ResumeOptimizationAgent 输出。

## 6. 当前前端组件现状

### 6.1 自定义业务组件

| 组件 | 文件 | 说明 | 可复用性 |
|------|------|------|---------|
| Button | Button.tsx | 自定义按钮 | ✅ 高 |
| Card | Card.tsx | 自定义卡片（含 CardHeader, CardDivider） | ✅ 高 |
| ScoreRing | ScoreRing.tsx | 分数环组件 | ✅ 高 |
| StatusTag | StatusTag.tsx | 状态标签（StatusTag + TextTag） | ✅ 高 |
| PageShell | PageShell.tsx | 页面壳（含 PageHeader） | ✅ 高 |
| AppNav | AppNav.tsx | 顶部导航 | ✅ 高 |

### 6.2 shadcn/ui 组件（40+）

完整 shadcn/ui 组件库已安装，包括：

| 类别 | 组件 |
|------|------|
| 表单 | input, textarea, select, checkbox, radio-group, switch, slider, form |
| 数据展示 | table, card, badge, avatar, progress, tabs, accordion, separator |
| 反馈 | alert, alert-dialog, dialog, drawer, sheet, sonner, tooltip |
| 导航 | navigation-menu, menubar, breadcrumb, pagination, sidebar |
| 布局 | resizable, scroll-area, aspect-ratio, collapsible |

**结论**：shadcn/ui 组件库足以支撑卡片、表格、看板、任务列表等产品化模块的实现。

## 7. 当前前端路由现状

| 核验项 | 结果 |
|--------|------|
| 是否有路由 | ❌ 无真实路由 |
| 路由方案 | useState 切换页面（Demo 预览模式） |
| react-router 依赖 | ✅ 已安装 `react-router@7.13.0`，但未使用 |
| 是否单页 Demo | ✅ 是 |
| 是否多页面 | ❌ 否，所有页面通过 useState 条件渲染 |
| 是否有推荐入口 | ✅ 页面存在但仅 Demo |
| 是否有匹配分析入口 | ✅ 页面存在但仅 Demo |
| 是否有简历优化入口 | ❌ 无页面 |

**App.tsx 路由机制**：

```typescript
type Page = "home" | "resume" | ... | "recommend" | "match" | ... | "login";

// 底部浮动 Demo 切换器
const [page, setPage] = useState<Page>("resume-detail");

// 条件渲染
{page === "recommend" && <RecommendPage hasResume={true} />}
{page === "match" && <MatchPage />}
```

## 8. 当前 Mock 数据现状

### 8.1 Mock 数据位置

**所有 Mock 数据硬编码在页面组件内部**，没有独立的 Mock 数据文件或目录。

| 页面 | Mock 数据位置 | 数据结构 |
|------|--------------|---------|
| JobsPage.tsx | `JOBS: Job[]` | 硬编码 6 个岗位（id, title, type, locations, tags, summary, desc, requirements） |
| RecommendPage.tsx | `RECOMMENDATIONS: RecommendJob[]` | 硬编码 4 个推荐（rank, title, type, locations, score, level, risk, reasons, shortfalls, desc, requirements, tags） |
| MatchPage.tsx | `EVIDENCE_ROWS: EvidenceRow[]` | 硬编码 5 条证据链（requirement, evidence, strength, risk, suggestion） |

### 8.2 Mock 数据与 Agent 输出的差距

| 维度 | 当前 Mock 数据 | Agent 输出 | 差距 |
|------|---------------|-----------|------|
| 推荐等级 | level: "priority"/"try"/"low" | recommend_level: "priority_apply"/"recommended"/"explore_first"/"not_recommended" | ❌ 枚举不一致 |
| 推荐标签 | 无 | recommend_label: "优先投递"等 | ❌ 缺失 |
| 分数 | score (0-100) | match_score (0-100) | ⚠️ 字段名不一致 |
| 风险等级 | risk: "low"/"mid"/"mid-high"/"high" | severity: "high"/"medium"/"low" | ❌ 枚举不一致 |
| 投递准备度 | 无 | apply_readiness | ❌ 缺失 |
| 优化停止信号 | 无 | optimization_stop_signal | ❌ 缺失 |
| Top Fix | 无 | top_fix_items | ❌ 缺失 |
| 证据链 | 简化版 (requirement/evidence/strength/risk/suggestion) | 完整版 (jd_requirement/resume_evidence/judgement/explanation) | ❌ 字段名和结构不一致 |
| 证据强度 | "strong"/"mid"/"weak"/"none" | "strong"/"medium"/"weak"/"missing" | ❌ 枚举不一致 |
| 匹配判断 | 无 | judgement: "matched"/"partial"/"missing" | ❌ 缺失 |
| 面试追问 | 无 | interview_questions | ❌ 缺失 |
| OfferGo 摘要 | 无 | offer_go_summary | ❌ 缺失 |
| 下一步行动 | 硬编码步骤列表 | next_actions (结构化) | ❌ 结构不一致 |
| 生成来源 | 无 | generated_by / agent_status | ❌ 缺失 |
| 简历优化 | 无独立页面 | ResumeOptimizationResult 全部字段 | ❌ 完全缺失 |

**结论**：当前 Mock 数据是前端设计时的"想法原型"，与后端 Agent 输出差距很大。后续需要：
1. 统一枚举值
2. 对齐字段名
3. 补齐缺失的战略字段
4. 新增简历优化页面的 Mock 数据

## 9. 当前 API 对接现状

| 核验项 | 结果 |
|--------|------|
| 是否有 API client | ❌ 无 |
| 是否使用 axios | ❌ 否 |
| 是否使用 fetch | ❌ 否 |
| 是否使用 VITE_API_BASE_URL | ❌ 否 |
| 是否使用 import.meta.env | ❌ 否（源码中无） |
| 是否调用后端 /api | ❌ 否 |
| 是否对接推荐接口 | ❌ 否 |
| 是否对接匹配分析接口 | ❌ 否 |
| 是否对接简历优化接口 | ❌ 否 |
| 是否仍为纯 Mock 前端 | ✅ 是 |

**结论**：当前前端为 100% 纯 Mock 静态 Demo，没有任何 API 对接能力。需要从零搭建 API client 层。

## 10. RecommendationAgent 前端承接能力

### 10.1 评估

| 承接能力 | 评估 | 说明 |
|----------|------|------|
| 岗位列表 | ✅ 已具备 | RecommendPage 有 RecommendCard 列表 |
| 岗位卡片 | ✅ 已具备 | 含标题、类型、城市、标签、分数环 |
| 推荐等级展示 | ⚠️ 部分具备 | 有 level 展示，但枚举值与 Agent 不一致 |
| 岗位排序 | ⚠️ 部分具备 | 有 rank 排序，但未使用 recommend_level |
| 多岗位对比 | ❌ 不具备 | 无对比功能 |
| 岗位风险展示 | ⚠️ 部分具备 | 有 risk 展示，但枚举值与 Agent 不一致 |
| 岗位下一步行动 | ❌ 不具备 | 无 next_action 展示 |

### 10.2 后续需要映射的字段

| Agent 字段 | 前端当前状态 | 差距 |
|-----------|-------------|------|
| recommendations | 有 RecommendCard 列表 | 字段名/枚举需对齐 |
| recommend_level | 有 level，枚举不一致 | priority/try/low → priority_apply/recommended/explore_first/not_recommended |
| recommend_label | 无 | 需新增 |
| match_score | 有 score | 字段名需对齐 |
| apply_readiness | 无 | 需新增投递准备度卡片 |
| optimization_stop_signal | 无 | 需新增优化停止信号卡片 |
| top_fix_items | 无 | 需新增 Top 3 快速修复任务 |
| evidence | 有简化版 reasons/shortfalls | 需升级为证据链表格 |
| risk_points | 有 risk，枚举不一致 | 需对齐 |
| query_suggestions | 无 | 需新增搜索建议 |
| next_action | 无 | 需新增下一步行动 |

## 11. MatchAnalysisAgent 前端承接能力

### 11.1 评估

| 承接能力 | 评估 | 说明 |
|----------|------|------|
| 匹配分数展示 | ✅ 已具备 | MatchPage 有 ScoreRing + 72 分 |
| 投递准备度展示 | ❌ 不具备 | 无 apply_readiness 展示 |
| 优化停止信号展示 | ❌ 不具备 | 无 optimization_stop_signal 展示 |
| JD 证据链表格 | ⚠️ 部分具备 | 有 EvidenceTable，但字段/枚举与 Agent 不一致 |
| 风险点列表 | ⚠️ 部分具备 | 有主要风险展示，但为简单字符串列表 |
| Top Fix 任务 | ❌ 不具备 | 无 top_fix_items 展示 |
| 面试追问 | ❌ 不具备 | 无 interview_questions 展示 |
| 下一步行动 | ⚠️ 部分具备 | 有 NextStepsCard，但为硬编码步骤 |

### 11.2 后续需要映射的字段

| Agent 字段 | 前端当前状态 | 差距 |
|-----------|-------------|------|
| match_score | 有 ScoreRing | ✅ 可直接对齐 |
| match_level / match_label | 无 | 需新增匹配等级标签 |
| screening_result | 无 | 需新增初筛结果 |
| apply_readiness | 无 | 需新增投递准备度卡片 |
| optimization_stop_signal | 无 | 需新增优化停止信号卡片 |
| top_fix_items | 无 | 需新增 Top 3 快速修复任务 |
| evidence_chain | 有 EvidenceTable | 字段/枚举需对齐（requirement→jd_requirement, strength→judgement 等） |
| matched_points | 有核心优势列表 | 结构需对齐 |
| gap_points | 无独立展示 | 需新增 |
| risk_points | 有主要风险列表 | 结构需对齐（string → {risk, severity, reason, mitigation}） |
| interview_questions | 无 | 需新增面试追问模块 |
| next_actions | 有 NextStepsCard | 需升级为结构化行动 |
| offer_go_summary | 无 | 需新增摘要卡片 |
| authenticity_notice | 有 AuthenticityCard | 内容需对齐 |

## 12. ResumeOptimizationAgent 前端承接能力

### 12.1 评估

| 承接能力 | 评估 | 说明 |
|----------|------|------|
| 简历优化总览 | ❌ 不具备 | 无简历优化页面 |
| 简历模块诊断 | ❌ 不具备 | 无 |
| 技能栏优化 | ❌ 不具备 | 无 |
| 项目经历优化 | ❌ 不具备 | 无 |
| 关键词对齐 | ❌ 不具备 | 无 |
| 示例参考 | ❌ 不具备 | 无 |
| 用户核实清单 | ❌ 不具备 | 无 |
| 缺失证据计划 | ❌ 不具备 | 无 |
| Top Fix 任务 | ❌ 不具备 | 无 |

### 12.2 后续需要映射的字段

| Agent 字段 | 前端当前状态 | 差距 |
|-----------|-------------|------|
| optimization_summary | 无 | 需新增简历优化工作台 |
| apply_readiness | 无 | 需新增投递准备度卡片 |
| optimization_stop_signal | 无 | 需新增优化停止信号卡片 |
| top_fix_items | 无 | 需新增 Top 3 快速修复任务 |
| screening_focus | 无 | 需新增初筛重点 |
| priority_suggestions | 无 | 需新增分优先级优化建议 |
| keyword_alignment | 无 | 需新增关键词对齐表 |
| project_optimization | 无 | 需新增项目优化模块 |
| skill_section_optimization | 无 | 需新增技能栏优化模块 |
| missing_evidence_plan | 无 | 需新增用户核实清单 |
| rewrite_examples | 无 | 需新增示例参考卡片 |
| authenticity_notice | 无 | 需新增真实性提醒 |
| next_actions | 无 | 需新增下一步行动 |

**结论**：ResumeOptimizationAgent 完全没有前端承接能力，需要从零设计简历优化页面。

## 13. 初步 UI 模块候选映射

| # | UI 模块 | 来源 Agent | 来源字段 | 前端当前是否已有类似承载 | 后续实现优先级 |
|---|---------|-----------|---------|----------------------|--------------|
| 1 | 投递准备度卡片 | 三类 Agent 共有 | apply_readiness | ❌ 无 | **P0** |
| 2 | 优化停止信号卡片 | 三类 Agent 共有 | optimization_stop_signal | ❌ 无 | **P0** |
| 3 | Top 3 快速修复任务 | 三类 Agent 共有 | top_fix_items | ❌ 无 | **P0** |
| 4 | JD 证据链表格 | MatchAnalysis + Recommendation | evidence_chain / evidence | ⚠️ 有简化版，需对齐 | **P0** |
| 5 | 岗位推荐看板 | Recommendation | recommendations + recommend_level + match_score | ⚠️ 有 RecommendCard，需对齐 | **P0** |
| 6 | 多岗位对比表 | Recommendation | 多个 recommendation_item 横向对比 | ❌ 无 | P1 |
| 7 | 简历优化工作台 | ResumeOptimization | optimization_summary + priority_suggestions + keyword_alignment + project_optimization + skill_section_optimization | ❌ 完全缺失 | **P0** |
| 8 | 示例参考卡片 | ResumeOptimization | rewrite_examples | ❌ 无 | P1 |
| 9 | 用户核实清单 | ResumeOptimization | missing_evidence_plan + verification_notice | ❌ 无 | P1 |
| 10 | 风险提示卡 | MatchAnalysis + Recommendation | risk_points | ⚠️ 有简化版 | P1 |
| 11 | 面试追问模块 | MatchAnalysis | interview_questions | ❌ 无 | P2 |
| 12 | 下一步行动模块 | 三类 Agent 共有 | next_actions | ⚠️ 有硬编码版 | P1 |

### P0 模块说明（4-3.9.1 需重点设计）

1. **投递准备度卡片**：三类 Agent 共有的战略字段，直接决定用户是否应该投递
2. **优化停止信号卡片**：与 apply_readiness 联动，明确告诉用户"可以投了"还是"先改再投"
3. **Top 3 快速修复任务**：最直接可操作的产品化模块，每个修复项含 issue/why_it_matters/action/example_reference/self_check_question/estimated_effort
4. **JD 证据链表格**：MatchPage 已有简化版，需对齐 Agent 输出的 evidence_chain 结构
5. **岗位推荐看板**：RecommendPage 已有卡片，需对齐 recommend_level/label 和新增 apply_readiness
6. **简历优化工作台**：完全缺失，需从零设计，这是 ResumeOptimizationAgent 的主要承载页面

## 14. 当前差距与风险

### 14.1 结构性差距

| 差距 | 严重度 | 说明 |
|------|--------|------|
| 无 API 对接层 | 🔴 高 | 前端 100% 硬编码 Mock，无 axios/fetch/request client |
| 无真实路由 | 🔴 高 | useState 切换，react-router 已装未用 |
| 无简历优化页面 | 🔴 高 | ResumeOptimizationAgent 完全无法承接 |
| Go Model 与 Schema 差距 | 🟡 中 | ResumeOptimization Go Model 仅有 Mock 字段，LLM Agent 输出的 keyword_alignment/project_optimization 等未映射 |

### 14.2 数据差距

| 差距 | 严重度 | 说明 |
|------|--------|------|
| 推荐等级枚举不一致 | 🟡 中 | 前端 priority/try/low vs Agent priority_apply/recommended/explore_first/not_recommended |
| 风险等级枚举不一致 | 🟡 中 | 前端 low/mid/mid-high/high vs Agent high/medium/low |
| 证据强度枚举不一致 | 🟡 中 | 前端 strong/mid/weak/none vs Agent strong/medium/weak/missing |
| 战略字段完全缺失 | 🔴 高 | apply_readiness / optimization_stop_signal / top_fix_items 前端完全无承载 |
| 面试追问缺失 | 🟢 低 | P2 模块 |

### 14.3 能力差距

| 差距 | 严重度 | 说明 |
|------|--------|------|
| 无 API 请求封装 | 🔴 高 | 需从零搭建 |
| 无错误处理 | 🔴 高 | 无 401/403/500 处理 |
| 无 loading 状态 | 🟡 中 | 当前为静态 Demo |
| 无状态管理 | 🟡 中 | 无 zustand/redux 使用 |
| 样式方案混合 | 🟢 低 | 内联 style + Tailwind 混用 |

## 15. 是否可以进入 4-3.9.1

**✅ 可以进入。**

理由：
1. 已核验三类 Agent 输出字段，字段定义清晰
2. 已核验当前前端页面、组件、路由、Mock 数据、API 对接现状
3. 已分别评估三类 Agent 前端承接能力
4. 已输出初步 UI 模块候选映射，P0 模块已标识
5. 已识别关键差距和风险
6. 4-3.9.1 的输入已具备

## 16. 下一步建议

1. **4-3.9.1 Agent 输出到前端模块映射方案**：✅ 已完成。基于本核验报告，已设计三类 Agent 输出字段到前端 UI 模块的正式映射方案
   - 已设计 P0 模块：投递准备度卡片、优化停止信号卡片、Top 3 快速修复任务、JD 证据链表格、岗位推荐看板、简历优化工作台基础版、用户核实清单基础版、结果来源状态标签
   - 已确定枚举值对齐策略
   - 已确定简历优化工作台页面布局
   - 已明确 ResumeOptimization P0 只基于 Go Model 实际字段，Schema 理想字段列为 P1/P2

2. **4-3.9.2 前端实现拆分计划与验收标准冻结**：✅ 已完成。已拆分后续前端产品化实现为 4-4.0 至 4-4.8 阶段，冻结验收标准和不可混做事项
   - 下一步进入 4-4.0：前端数据契约与 Mock 数据标准化
