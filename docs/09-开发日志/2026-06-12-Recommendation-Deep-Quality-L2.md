# 2026-06-12 Recommendation-Deep-Quality-L2 — Deep Cascade 输出质量升级

## 任务性质

后端 Deep Cascade prompt 和字段映射优化，提升 Deep 推荐结果的专业度和可信度。

## 修复目标

1. 优化 Layer 3 prompt，让每个岗位输出字段明确分工
2. 优化 Layer 4 prompt，让全局校准更像专业评审判断
3. 修复字段映射，避免 LLM 同一段文案映射到多个前端字段
4. 增强 validator，拦截重复字段、套话、空泛建议、raw code

## 修改文件

| 文件 | 修改内容 |
|---|---|
| `server/internal/service/recommend_deep_cascade.go` | L3 struct +5 字段(JobSpecificReason/SuggestedAction/JobSpecificRisks/CandidateCoreStrength/FitConfidenceReason)；L4 struct +5 字段；buildRecommendedJobsFromCascade 字段独立映射；RunDeepRecommendationCascade 返回 L4 结果 |
| `server/internal/service/recommend_deep_cascade_prompt.go` | L3 prompt 全面重写（17条规则+反例约束+45字限制+字段不重复要求）；L4 prompt 重写（校准器身份+5段式global_summary+新字段输出） |
| `server/internal/service/recommend_deep_cascade_validator.go` | validateLayer3 增加字段重复检测+套话检测+display_tags raw code检测；新增 sameText/strListsEqual/isBoilerplateText 工具函数；sanitizeCascadeOutputs 覆盖新字段 |
| `server/internal/model/recommendation.go` | RecommendedJob 新增 DisplayTags/ApplicationStrategy/CandidateCoreStrength/FitConfidenceReason |
| `server/internal/service/recommendation_service.go` | RecommendationBusinessResult 新增 L4 校准字段；recommendDeepCascade 传递 L4 结果+覆盖新字段 sanitize |
| `web/src/adapters/recommendationAdapter.ts` | ViewModel 新增 candidateCoreStrength/fitConfidenceReason 字段映射 |

## Layer 3 字段分工

| 字段 | 角色 | 约束 |
|---|---|---|
| one_sentence_conclusion | 一句话结论（卡片首屏） | ≤45字，有判断 |
| job_specific_reason | 详细推荐理由 | 包含岗位要求-简历证据对应，≠conclusion |
| risk_points | 主要风险 | 说明为什么是风险 |
| job_specific_risks | 更细风险 | 可空，不可与 risk_points 相同 |
| application_strategy | 投递策略（岗位级） | 80-150字，不可只说"建议投递" |
| suggested_action | 马上做什么 | ≤25字 |
| resume_improvement_hint | 简历补强建议 | 80-150字，具体到模块/表达 |
| why_ranked_here | 排序原因 | 相对于其他 Top5 |
| candidate_core_strength | 核心竞争力 | ≤40字 |
| fit_confidence_reason | 评分置信度 | ≤60字 |

## Validator 增强

- 字段非空检测
- one_sentence_conclusion≠job_specific_reason 检测
- application_strategy≠suggested_action 检测
- application_strategy≠resume_improvement_hint 检测
- risk_points≠job_specific_risks 检测（非空时）
- 套话黑名单："存在一定关联""建议进一步确认""具备一定基础"等
- display_tags raw code 检测
- global_summary 长度检查（≤250字）

## 构建验证

```
go build ./... ✅
go vet ./... ✅
go test ./... ✅ (0.007s)
npm build ✅ (2.02s)
```

## 待验证

因环境 LLM_API_KEY 未配置，后端暂无法启动，Deep API 真实验证待后续完成。
