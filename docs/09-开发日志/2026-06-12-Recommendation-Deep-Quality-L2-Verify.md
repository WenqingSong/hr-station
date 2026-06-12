# 2026-06-12 Recommendation-Deep-Quality-L2-Verify — Deep 输出质量真实运行验收

## 性质
运行级验收（不修改代码）。

## 验证结果

### Deep API
- HTTP 200, generated_by=llm_agent, 5 items
- 总耗时约 146 秒（2.4 分钟）
- Cascade 全链路: L0 gate → L1 Discovery(29.5s) → L2 Rerank(30.3s) → L3 Synthesis ×5(~80s) → L4 Calibration(6.8s)

### 新字段返回
- candidate_core_strength ✅
- fit_confidence_reason ✅
- job_specific_reason ✅（与 recommendation_reason 不同）
- suggested_action ✅（与 application_strategy 不同）
- job_specific_risks ✅（与 risk_points 不同）
- global_summary ✅（150字左右，有方向性判断）
- candidate_positioning ✅

### 字段重复检查
5 个岗位 4 组字段对比全部 FALSE：
- reason_vs_specific: 全部不重复 ✅
- suggested_vs_strategy: 全部不重复 ✅
- strategy_vs_improvement: 全部不重复 ✅
- risk 副本: 全部不重复 ✅

### 低级问题
- raw code 扫描: 0 暴露 ✅
- mock/fallback: 0 暴露 ✅
- 套话禁止列表: 0 命中 ✅

### Fast 回归
- HTTP 200, generated_by=fast_tag_engine ✅

### 构建
- go build / go vet / go test / npm build 全部通过

### 最终判断
L2 正式收口 ✅ — Deep 结果质量已达到可给评委展示水平。
