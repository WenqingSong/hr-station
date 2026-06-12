# 2026-06-12 Recommendation-Source-Governance-M3 — 推荐结果来源治理

## 任务性质
来源治理 + LLM Capability Gate + 历史过滤 + backend 方向诊断。

## 修复

### 1. LLM Capability Gate
- quality tier 入口新增 `!s.llmEnabled` 和 `s.qualityClient == nil` 显式检查
- LLM 不可用时返回 503 + 明确中文错误文案
- 不再依赖 `enableMockAgent` 决定是否 mock quality 结果

### 2. 历史结果过滤
- `FindLatestByResumeID` 增加 `generated_by NOT IN ('mock_agent', 'mock_quality')` 条件
- 用户可见历史不再包含 mock 数据

### 3. Backend 方向诊断
- jobs 表中存在 "软件开发-后台开发方向"（id=10001, category="后台开发"）
- M2B 的 direction gate 已正确识别 backend direction
- 因 tag 数量略少（3 vs 4），Rank 5 排在 AI应用/测试开发/运营/全栈之后
- 判断合理，无需硬编码排序

## DB 状态
- mock_agent: 40 条历史（旧，已被过滤）
- llm_agent: 6 条（真实 Deep 结果）
- 新查询不再展示 mock

## 验证
- Fast API: scores [98,93,91,88,85], range=13, 10001 进 Top5
- Latest API: 返回 llm_agent 结果，mock 被过滤
- raw code/mock/fallback: 0
