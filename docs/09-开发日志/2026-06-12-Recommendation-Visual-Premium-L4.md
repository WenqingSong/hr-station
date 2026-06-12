# 2026-06-12 Recommendation-Visual-Premium-L4 — 岗位推荐页高级视觉与评委观感升级

## 任务性质
前端视觉提升（不改后端/数据库）。

## 视觉升级内容

### 顶部洞察卡
- 替换旧蓝渐变条为白底 "AI 推荐洞察" 卡片
- 双层网格：候选人定位 + 推荐洞察
- 右上角 "级联管线已完成" badge

### 左侧推荐榜单
- 新增"推荐榜单"标题 + Fast/Deep 标签
- Rank 数字圈 + 文案标签（#1 "最优先" → #5 "备选"）
- 分数展示（大号数字 + 分字）
- candidate_core_strength 优先展示
- 标签改用圆角 pill 样式

### 中间评估报告
- 区域标题改为 uppercase 小标
- "核心判断"：浅蓝背景 + fit_confidence_reason
- "推荐依据"：job_specific_reason 完整展示
- "匹配证据"：浅绿背景 + 圆点列表
- "主要风险"：浅橙背景 + 圆点列表
- "行动建议"：投递策略 + 简历补强双栏
- 按钮改为圆角 2xl

### 右侧智能分析控制台
- 重命名为"智能分析"
- 模式指示灯（绿点=Deep，蓝点=Fast）
- 完成态摘要
- 统一按钮样式
- 底部轻量提示文字

### 字段新增
- candidatePositioning 从 API response 透传到 ViewModel → 顶部洞察卡展示

## 构建
- npm build ✅ (2.0s)
- raw code scan ✅ 0
- visual text scan ✅ 全部区块到位
