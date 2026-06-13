# Offer捕手 项目记忆

## 项目概述
- 名称：Offer捕手（Offer Catcher）
- 技术栈：Go 后端 + React/TypeScript 前端 + MySQL + Redis + Nginx
- 核心功能：简历管理、岗位推荐（Fast/Deep）、岗位匹配分析、简历优化

## 用户偏好
- 不允许执行任何 Git 命令（任务中明确禁止）
- 只读审计任务中严禁修改代码
- 前端要追求优秀设计，不满足于低质量实现
- 不要为了兼容旧代码而保留差代码，该重构就重构
- H16-H17 系列任务采用严格的多阶段流程：Audit → Design Freeze → Implement → Acceptance

## 深度推荐架构关键事实（H16A-H17B 已确认）
- Deep Cascade 有 5 层：L0 Gate → L1 候选发现 → L2 语义重排 → L3 逐岗合成×5 → L4 全局校准
- 后端在单个 HTTP POST 内串行执行所有层，完成才返回（约 100s）
- 没有 task_id、stage/progress、SSE、轮询阶段接口
- 响应中有完整 pipeline_trace（9 步含逐岗 job_id 和耗时）
- pipeline_trace 无法用于实时进度展示，但可在完成后展示"分析履历"
- 右侧"快速推荐：可用"在深度完成后应条件隐藏（H16B 已修复）

## AgentStepLoadingOverlay 组件
- 位置：`web/src/app/components/common/AgentStepLoadingOverlay.tsx`
- 被 4 个页面引用：JobRecommendationPage, JobMatchAnalysisPage, ResumeComparisonPage, ResumeOptimizationPage
- 有 `noAutoAdvance` 补丁参数（H16B 新增），用于关闭假进度

## H17 设计决策
- H17-B 冻结方案：新增 `DeepAgentLoadingOverlay.tsx` 专用组件（~200 行），纯 CSS 动画
- 方案 C：纯前端 UX 重构，不改后端
- 中心 96px 呼吸环 + 横向 shimmer pill + breathing dot 岗位矩阵
- 旧 AgentStepLoadingOverlay 保持不变，其他 3 个页面不受影响

## H19-A Deep Match 架构冻结（已完成）
- 当前 Deep: fit_points/evidence_chain=null, risk 为 bare strings, 缺 JD 解构
- 冻结 8 层 Agent: L0 Gate→L8 Calibration, 12 模块输出, deep_match_agent
- 关键新字段: job_understanding, core_match_points, key_risks(6字段), resume_rewrite_strategy, interview_risk_questions(≥5)
- Deep/Fast 共享分数字义, 分差>15 须解释

## H18-D Fast Match Engine（已完成）
- match_sep_engine.go 830 行，8 模块（Normalization+Direction+Gate+Coverage+Evidence+Risk+Calibrator+Assembler）
- 5 维度加权评分，API 实测 5 jobs 分数 29-49 有区分，<0.005s

## H17-D 状态整合（已完成）
- 新增 runId guard（deepRunSeqRef + cancelledDeepRunRef）防止旧请求覆盖 + 取消后强制切换
- isDeepDisplay = phase === "deep_done" && activeTier === "quality" && deepReady
- 右侧 4 种互斥状态：Fast only / Deep loading / Deep displayed / Deep bg complete
- "查看深度推荐"入口 + handleViewDeepBackground()
- 预存 TS 错误已修复（metaRaw as any），推荐页 TS 零错误

## H18-A 岗位匹配审计（已完成）
- 核心发现：匹配用 TagEvidenceService（标签画像），推荐用 SemanticProfileGate（语义画像）——系统性割裂
- Fast Match：`min(98, 60 + tagScore*0.8)`，highlights 可为 null，raw tag code 泄露
- Deep Match：3-agent 但基于 TagProfile，非 SEP
- match_analyses 表存在，SEP resume_profiles/job_profiles 已可用
## H18-B-R 匹配契约冻结（修正版，2026-06-13 22:10）

### 修正原因
上一版 H18-B 报告声称 H18-C→H23-B 已完成但未提供真实代码依据。本次基于全量只读审计重做。

### 冻结成果
- 统一分数字义：5 级（90-100 高度匹配 / 80-89 较高匹配 / 70-79 中等匹配 / 60-69 弱匹配 / 0-59 不建议投递）
- 8 条 Score Cap 规则（mismatch≤65, mustHaveMiss≤75, noEvidence≤72, noProject≤70, noInternship≤78, weakExpr≤85, blocked≤60, strongMatch≥90）
- 投递建议 5 枚举：priority_apply / apply_directly / revise_before_apply / backup_apply / not_recommended
- 结构化 risk（5 字段 + high/medium/low）/ evidence（7 字段 0-5 级）/ requirement_coverage（7 字段 + 6 状态）
- Fast 契约：fast_sep_match_engine 16 字段全 `[]` 禁止 null
- Deep 契约：deep_match_agent 10 模块（JD解构/证据审计/要求覆盖/风险诊断/简历改写/面试追问/投递策略/行动规划/Agent trace）

### 审计发现的关键不一致（3 项，待 H18-C/D 修复）
1. **后端 matchLevel 80/65/50/35 vs 前端 matchLevelLabel 90/80/70/60**（需在 H18-D 修复）
2. **SEP engine `SupportsRequirement: t.Category` 泄露 raw tag code**（需在 H18-D 修复，改为 humanized 映射）
3. **tag fallback 仍用 `fast_tag_fallback_engine`**（需在 H18-C 迁移到 fast_sep_match_engine）

### 仍活跃的遗留依赖
- TagEvidenceService：3 条 active fallback 路径
- matchWith3Agents：Deep SEP 内部支柱
- analyzeWithLLM / mockAgent：最终兜底

### 后续阶段
**全部完成** ✅ H18-C→F + H19-A→D + H20-A PASS

## H20-A Job Match Full-Chain Final Acceptance（PASS — 可上线演示）

### 验收总结
- Fast 9/9 + Deep 4/4 = 13 tests ALL PASS
- Fast/Deep 全部 delta ≤ 9，is_consistent=true
- 6 状态验证：success + profile_pending + legacy_fallback + mock_fallback
- 0 raw code，0 null，0 P0/P1
- 岗位匹配完成全链路收口：Fast 秒级初筛 + Deep 8-layer 顾问报告

## H19-D Deep Match Precision Calibration（已完成，ALL PASS）

### 验收结果
- 5 样例 ALL PASS：强75/邻60/不41/不42/证52
- Fast/Deep 全部 delta ≤ 9，is_consistent=true
- Schema 25 字段完整性 ✅，0 raw code，0 null
- 新增 deep_calib_test.go：5 样例 + schema + alignment + pending

## H19-C Deep Match Frontend Report Refactor（已完成，2026-06-13 23:44）

### 实现内容
- adapter 新增 scoreAlignment/qualityChecks/applicationStrategyObj + isDeepV3/isDeepLegacyFallback/isDeepProfilePending/isDeepMockFallback
- Deep 区域重构为 v3 10-module: 决策卡/分数对齐/JD解构/要求覆盖/证据映射/核心匹配/风险诊断/改写策略/面试追问/Agent履历
- 4 状态分流: deep_v3完整 / profile_pending橙色 / legacy_fallback琥珀色 / mock_fallback红色
- Module 2: Fast/Deep 对比 + deltaReason
- Module 10: 8-stage agent_trace + qualityChecks folded
- 构建: vite ✅ 2.17s

## H19-B Deep Match Agent Backend Refactor（已完成，2026-06-13 23:30）

### 实现内容
- 新增 deep_match_v3 Schema (25字段+score_alignment+quality_checks)
- buildDeepMatchContextV3: 读取 SEP + Fast Result, TagEvidenceService 已移除
- runDeepMatchV3: 8层流水线 L0-L8
- generated_by 收口: v3=deep_match_agent, SEP缺失=deep_match_profile_pending, legacy=deep_match_legacy_fallback
- 修复 sanitizeDeepResponse 不覆盖 legacy fallback 标记
- 构建全通过(vite 1.93s/go test)

### H19-C 待处理
前端接入 deep_match_v3: score_alignment/quality_checks/JD解构/要求覆盖/证据映射/结构化风险/改写策略/面试追问/agent_trace

## H19-A Deep Match Schema/Prompt Contract Freeze（已完成，2026-06-13 23:16）

### 审计发现
- 数据源分裂：3-Agent 使用 TagEvidenceService（旧标签），DeepMatchReport 使用 SEP（新标签）
- Fast 初判未读取：FastResult 字段声明但从未赋值
- 缺失字段：ScoreBreakdown/DecisionLabel/StructuredRisks/HardGate 全部为空
- 4 层降级链：3-agent→LLM→fast fallback→mock_quality

### 冻结内容
- 13 项输入上下文（SEP Profile + Fast Result 首选，Legacy Tag 仅兜底）
- 8 层 Agent 流程（L0 Input Gate → L8 Global Consistency）
- deep_match_v3 Schema：25 顶层字段 + score_alignment + quality_checks
- Prompt 约束：13 必须 + 7 禁止
- 10 个前端展示模块 + 6 维 score_breakdown + 8 条 score cap
- H19-B 目标：移除 TagEvidenceService 依赖，读取 Fast Result，填充全部 deep_match_v3 字段

## H18-F Fast Match Precision Calibration（已完成，PASS_WITH_FIXES）

### 验收结果
- 5 样例全 PASS（强匹配 76/adjacent 58/mismatch 27/mismatch 25/low-ev 53）
- matchLevel 11 边界值验证 ✅、Hard Gate 8 cap 验证 ✅

### P0 修复
1. checkHardGate: 将"实习"加入 hasProjectEvidence keyword（原来只查"项目"）
2. buildRequirementCoverage: 移除错误 jt.EvidenceText 检查（改为纯 evLvl 判定）

### 遗留（H19）
- Fast 整体偏保守（最高 76），adjacent 方向分数偏低
- Deep 仍用 3-agent tag pipeline，需 H19-A 统一契约

### 实现内容
- adapter 新增 isProfilePending/profilePendingReason，检测 fast_sep_profile_pending + profile 三态
- matchLevelLabel "边缘匹配"→"不建议投递"
- profile_pending 专用 UI：橙色状态卡片 + 原因 + 重新生成按钮，不展示 0 分
- Fast 6 模块卡（结论/五维+HardGate/优势+短板/风险/要求覆盖/行动建议）全部接入 H18-D 新字段
- Deep 10-module 完全未修改

### 构建
vite ✅ 1.96s, go build/vest/test ✅

## H18-D Fast Match Engine Core Algorithm Refactor（已完成，2026-06-13 22:32）

### 实现内容
- matchLevel 阈值统一：80/65/50/35→90/80/70/60（与前端一致）
- raw code 泄露修复：新增 `humanizeRequirementCategory()`，10 类 Category→中文映射
- risk 结构化：保留 string[] 兼容，新增 `structured_risks` 5 字段（risk_name/risk_level/why_it_matters/evidence_gap/fix_suggestion）
- coverage 补全：strong/partial → strong/partial/transferable/weak/missing/unknown 6 状态
- 分数校准重写：5 维度 0-100 直接输出 + weighted composite（25%+30%+25%+10%+10%）
- Hard Gate 强化：4→8 条规则（含 internship 检测、weak expression cap）
- model 新增 7 字段 + engine 输出 score_breakdown/hard_gate/core_advantages/main_gaps/action_suggestions

### 构建
vite ✅ 2.12s, go build ✅, go vet ✅, go test ✅

### H18-E 待处理
前端读取新字段：structured_risks/score_breakdown/hard_gate/core_advantages/main_gaps/action_suggestions；区分 profile_pending/success 状态
