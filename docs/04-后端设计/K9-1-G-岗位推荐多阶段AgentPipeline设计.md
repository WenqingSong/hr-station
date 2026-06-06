# K9-1-G：岗位推荐多阶段 Agent Pipeline 与评分文案设计

> 状态：设计冻结 | 日期：2026-06-06

## 1. 背景与问题（K9-1-F 验收结论）

| 已修复 | 仍存在问题 |
| -- | -- |
| ✅ Top1 不再是后台开发 | ❌ 分数全部 55，无区分度 |
| ✅ 后台开发已排除出 Top5 | ❌ Fast = Quality，完全相同 |
| ✅ Quality gen=mock_quality | ❌ 推荐理由泛化 |

## 2. Fast Pipeline（2层）

### Layer 1：规则召回

先用候选画像 + 岗位摘要做排序，从全部岗位中召回 10-15 个。

### Layer 2：规则精排

对召回候选精排评分，输出 Top5 + 分数 + 理由 + 亮点 + 风险 + 建议。

## 3. Quality Pipeline（3层）

### Layer 1：候选召回
输入简历短画像 + 岗位压缩卡片，召回 10-15 个候选。

### Layer 2：精排评分
输入完整简历 + 候选岗位详情，输出 Top5 含分数/标签/证据。

### Layer 3：Judge 审查
检查方向错配、分数趋同、理由空泛、风险不真实等 9 项规则。
- pass → 直接返回
- minor → 规则修正
- major → 最多重试 1 次，失败则 fallback

## 4. 分数区间与推荐文案

| 分数 | 推荐文案 | priority |
| -- | -- | -- |
| 90-95 | 强匹配，优先投递 | strong |
| 80-89 | 高匹配，重点关注 | high |
| 70-79 | 可尝试，需要补强 | try |
| 60-69 | 备选参考，不建议优先 | backup |
| <60 | 不推荐 | not_recommended |

对应字段：`recommendation_label` (string) + `recommendation_priority` (enum)

## 5. 前端卡片展示优先级

推荐标签文案 > 岗位标题 > 分数 > 一句话理由 > 匹配点 > 风险提醒

## 6. 智能体深度思考 Loading

全局毛玻璃层 + 4步阶段文案（解析画像→召回候选→精排评分→校验质量），每步使用 spinner/check/wait 状态。

## 7. API Schema 新增字段

| 字段 | 类型 | 说明 |
| -- | -- | -- |
| `recommendation_label` | string | "强匹配，优先投递"等 |
| `recommendation_priority` | string | strong/high/try/backup/not_recommended |
| `pipeline_mode` | string | fast_rule / quality_pipeline / fallback |
| `match_evidence` | []Evidence | resume_signal + job_requirement + strength + explanation |
| `judge_result` | JudgeResult | pass/severity/issues/fix_strategy |

## 8. Fast vs Quality 差异

| 维度 | Fast | Quality |
| -- | -- | -- |
| 层数 | 2 | 3 |
| Judge | 轻量 guardrail | 完整 Judge |
| 证据 | 无 | match_evidence |
| gen_by | mock_agent | mock_quality |
| pipeline | fast_rule | quality_pipeline |

## 9. K9-1-H 后端实现边界

实现 3 层 pipeline、评分、Judge、新增 API 字段。不改前端/DB（不做 ALTER）。

## 10. K9-1-I 前端实现边界

卡片展示 label 优先、Deep 智能体 loading、全量替换。

## 11. K9-1-J 验收样例

5 份简历：01-算法研究生、02-211后台开发、03-二本产品经理、04-职业探索型、05-不适合互联网。
