# Deep-Recommend-Cascade-Pipeline-H — 级联式深度推荐管线

## 目标
将 Deep 推荐从旧 3-Agent 切换到 4 层级联式深度推荐管线。

## 新增文件（3 文件，~700 行）

| 文件 | 内容 |
|---|---|
| `service/recommend_deep_cascade.go` | Pipeline 编排 + 类型定义 + 所有层执行逻辑 + Response 组装 |
| `service/recommend_deep_cascade_prompt.go` | 4 层 Hy3 Prompt Builder |
| `service/recommend_deep_cascade_validator.go` | 4 层输出验证 + raw code sanitizer |

## 级联式管线

| Layer | 名称 | LLM | reasoning | 输入 | 输出 |
|---|---|---|---|---|---|
| 0 | SemanticProfileGate | — | — | SEP | 确保 resume + job SEP |
| 1 | Candidate Discovery | ×1 high | high | Resume SEP + 46 jobs SEP | 10-15 候选 |
| 2 | Semantic Rerank | ×1 high | high | L1 候选 + SEP | Top 5 + 排序理由 |
| 3 | Per-Job Deep Synthesis | ×5 high | high | Resume/JD + SEP + L2 context | 8 项深度分析 |
| 4 | Cross-Job Calibration | ×1 low | low | 5 L3 输出 + L2 | 校准排序 + 质量结论 |

**LLM 总计：8 次（1+1+5+1）**

## 修改文件

- `recommendation_service.go`: +qualityClient/qualityModel + recommendDeepCascade + Deep 路由
- `cmd/server/main.go`: qualityClient/qualityModel 上提到外层作用域
- `recommendation_service_test.go`: 参数补齐

## 验证

- go build/vet/test ✅ (11 包)
- npm build ✅
- make check ✅
- Deep 链路不经过旧 3-Agent ✅
- 旧代码保留为 fallback ✅
