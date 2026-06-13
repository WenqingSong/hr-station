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
## H18-B 匹配契约冻结（已完成）
- 统一分数字义：5 级 0-100（高/较高/中等/弱/不建议）+ 8 hard gate cap
- 投递建议 5 枚举：priority_apply ~ not_recommended
- 结构化 risk（5 字段）/ evidence（7 字段 0-5 级）/ coverage（7 字段 6 状态）
- Fast 契约：fast_sep_match_engine 15 字段全 `[]` 禁止 null
- Deep 契约：deep_match_agent 12 模块，比 Fast 多 JD 解构/改写/面试预测/投递策略/agent_trace
- 后端不改 DB，复用 result_json
- 后续 9 阶段：H18-C ~ H20-A，先 Fast 后 Deep，先 SEP 后算法
