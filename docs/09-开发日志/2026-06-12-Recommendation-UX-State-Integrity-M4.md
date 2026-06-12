# 2026-06-12 Recommendation-UX-State-Integrity-M4 — 推荐页前端状态完整性治理

## 任务性质
前端状态治理（不改后端/数据库）。

## 修复

### 错误分类
- 新增 `classifyDeepError()` 函数，识别 LLM_UNAVAILABLE / GENERATION_FAILED / NETWORK_ERROR / UNKNOWN
- 检测 503 + "大模型服务" 关键词 → LLM_UNAVAILABLE
- `deep_unavailable` 新阶段

### Deep 不可用处理
- LLM_UNAVAILABLE 时进入 `deep_unavailable` 阶段
- 不清空 fastResult / deepResult
- 右侧控制台显示 "深度推荐：暂不可用" + "深度推荐服务暂不可用"
- 按钮变更为 "稍后重试深度推荐" + "查看快速推荐"

### 能力状态展示
- 右侧控制台新增能力状态区：快速推荐/深度推荐 可用/暂不可用
- agent_status → `agentStatusLabel()` 翻译（cascade_verified → "已完成候选发现、语义重排、逐岗分析和全局校准"）
- generated_by → `sourceLabel()` 翻译（fast_tag_engine → "快速推荐引擎", llm_agent → "深度推荐智能体"）
- mock_agent/mock_quality → 返回空字符串（不展示）

### 错误不污染
- Deep 错误不写入 deepResult → Fast 结果保持不变
- `activeResult` 在非 deep_done 时 fallback 到 fastResult

## 构建
- npm build ✅ (1.96s)
- Raw source scan: 0 user-visible
- UX text scan: 全部状态文案到位
