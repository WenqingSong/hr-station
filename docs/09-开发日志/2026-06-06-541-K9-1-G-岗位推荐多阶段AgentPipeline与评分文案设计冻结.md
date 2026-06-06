# K9-1-G：岗位推荐多阶段 Agent Pipeline 与评分文案设计冻结

## 背景

K9-1-F 验收：分数全部55，Fast=Quality，理由泛化。

## 设计冻结

- **Fast Pipeline**：2层（规则召回 + 规则精排）
- **Quality Pipeline**：3层（候选召回 + 精排评分 + Judge审查）
- **Judge 9项规则**：方向错配、分数趋同、理由空泛等
- **分数区间**：5档（90-95强匹配 / 80-89高匹配 / 70-79可尝试 / 60-69备选 / <60不推荐）
- **推荐文案**：`recommendation_label` + `recommendation_priority`
- **API 新增字段**：label/priority/pipeline_mode/match_evidence/judge_result
- **前端卡片**：标签优先于分数
- **Deep loading**：智能体深度思考4步阶段文案

## 后续阶段

| 阶段 | 范围 |
| -- | -- |
| K9-1-H | 后端实现 |
| K9-1-I | 前端实现 |
| K9-1-J | 5样例验收 |

## 检查

encoding ✅ 未修改源码/DB
