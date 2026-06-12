# 2026-06-12 Recommendation-Loading-Experience-L3 — Deep 推荐生成过程体验升级

## 任务性质
前端 Loading 体验优化（不改后端/数据库）。

## 修复内容

### Loading 弹窗升级
- 增加已用时间实时显示（每秒更新，`0:00` 格式）
- 增加进度条（0-96% 预估，100% 完成）
- 5 阶段步骤对齐 Deep Cascade Layer 0-4
- 长等待分段提示（60s/180s 阈值）
- 明确标注"当前为预估进度"

### 阶段设计
1. 准备语义画像（Layer 0: SemanticProfileGate）
2. 候选岗位发现（Layer 1: Candidate Discovery）
3. 语义重排（Layer 2: Semantic Rerank）
4. 逐岗深度分析（Layer 3: Per-Job Synthesis ×5）
5. 全局一致性校准（Layer 4: Calibration）

### Deep 完成态
- 右侧栏显示"深度分析已完成" + 状态翻译
- cascade_verified → "已完成候选发现、语义重排、逐岗分析和全局校准"
- 错误态优化：分级展示（失败/超时），提供"重试"和"快速推荐"按钮

### 进度原则
- 明确标注"当前为预估进度，最终结果以分析完成后为准"
- 最大等待时间 900s
- 取消按钮文案改为"返回快速推荐结果"

## 修改文件
- `web/src/app/components/common/AgentStepLoadingOverlay.tsx`: elapsedSeconds prop, 进度条, 已用时间, 长等待提示
- `web/src/app/pages/JobRecommendationPage.tsx`: 5 阶段 steps, 已用时间计时器, 完成状态翻译, 错误态优化

## 构建
- npm build ✅ (1.93s)
- Raw code scan ✅ 0 泄露
- Loading text scan ✅ 全部 5 阶段文案到位
