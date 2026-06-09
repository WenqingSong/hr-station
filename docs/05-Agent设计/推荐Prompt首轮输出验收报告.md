# 推荐 Prompt 首轮输出验收报告

> 验收时间：2026-05-29
> 所属阶段：4-1.4.5
> 验收对象：server/prompts/recommendation_v1.md
> 验收方式：手动组装 Prompt 输入，调用 Hunyuan 一次，检查输出质量

## 1. 验收目标

验证 `recommendation_v1.md` 是否能让模型输出合法、专业、克制、有四档推荐、可执行查询建议和下一步行动的结构化岗位推荐结果。

## 2. 验收范围

- 1 份简历样本（Arthur V2）
- 4 个候选岗位（1 强匹配 + 1 中弱匹配 + 2 弱匹配）
- 1 次模型调用
- 12 项质量检查

## 3. 验收时间

2026-05-29

## 4. 测试样本

### 简历

- 用户：Arthur（user_id=1）
- 版本：V2（version_id=2）
- target_role：后端开发工程师
- target_city：深圳
- skills：Go, MySQL, Redis, Docker, Kubernetes, etcd, gRPC, Git, Linux
- 项目：校园二手交易平台（Go+Gin）、分布式任务调度系统（etcd）
- 实习：后端开发实习生（Go+微服务）、云原生开发实习生（K8s Operator）

### 候选岗位

| job_id | title | department | 匹配类型 | data_quality |
|--------|-------|------------|---------|-------------|
| 9 | 软件开发-后台开发方向 | CDG | 强匹配 | full_jd |
| 13 | 解决方案-行业咨询方向 | CSIG | 中弱匹配 | full_jd |
| 10 | 算法-多模态方向 | CDG | 弱匹配 | full_jd |
| 11 | 算法-自然语言处理方向 | CDG | 弱匹配 | full_jd |

## 5. Prompt 文件

- 路径：`server/prompts/recommendation_v1.md`
- 版本：v1
- 占位符替换：{{RESUME_CONTEXT_JSON}}、{{JOB_CANDIDATES_JSON}}、{{BUSINESS_CONTEXT}}

## 6. 模型调用方式

- Provider：Hunyuan（via TokenHub）
- Model：hy3-preview
- temperature：0.2
- max_tokens：8192
- prompt_tokens：8692
- completion_tokens：3832
- total_tokens：12524
- finish_reason：stop
- 调用结果：成功

## 7. JSON 合法性检查

| 检查项 | 结果 | 说明 |
|--------|------|------|
| 是否可被 JSON.parse 解析 | ✅ | 直接解析成功 |
| 是否为 JSON object | ✅ | 顶层为对象 |
| 是否没有 Markdown 包裹 | ✅ | 无代码块包裹 |
| 是否没有多余解释文字 | ✅ | 无前后缀 |

## 8. 字段完整性检查

### 顶层字段

| 字段 | 结果 |
|------|------|
| strategy_summary | ✅ |
| recommendations | ✅（4 项） |
| overall_risks | ✅（2 项） |
| resume_focus | ✅（3 项） |
| authenticity_notice | ✅ |
| analysis_basis | ✅ |
| generated_by | ✅ |

### 每个 recommendation 字段

| 字段 | job_id=9 | job_id=13 | job_id=10 | job_id=11 |
|------|----------|-----------|-----------|-----------|
| job_id | ✅ | ✅ | ✅ | ✅ |
| title | ✅ | ✅ | ✅ | ✅ |
| department | ✅ | ✅ | ✅ | ✅ |
| city | ✅ | ✅ | ✅ | ✅ |
| category | ✅ | ✅ | ✅ | ✅ |
| job_type | ✅ | ✅ | ✅ | ✅ |
| external_source | ✅ | ✅ | ✅ | ✅ |
| external_id | ✅ | ✅ | ✅ | ✅ |
| data_quality | ✅ | ✅ | ✅ | ✅ |
| source_url | ✅ | ✅ | ✅ | ✅ |
| recommend_level | ✅ | ✅ | ✅ | ✅ |
| recommend_label | ✅ | ✅ | ✅ | ✅ |
| match_score | ✅ | ✅ | ✅ | ✅ |
| offer_go_summary | ✅ | ✅ | ✅ | ✅ |
| recommend_reason | ✅ | ✅ | ✅ | ✅ |
| evidence | ✅（4条） | ✅（3条） | ✅（3条） | ✅（3条） |
| risk_points | ✅（2条） | ✅（1条） | ✅（1条） | ✅（1条） |
| query_suggestions | ✅（3条） | ✅（2条） | ✅（2条） | ✅（2条） |
| next_action | ✅ | ✅ | ✅ | ✅ |

## 9. 枚举合法性检查

| 检查项 | 结果 | 说明 |
|--------|------|------|
| recommend_level | ✅ | priority_apply / not_recommended |
| recommend_label | ✅ | 优先投递 / 暂不建议 |
| generated_by | ✅ | llm_agent |
| evidence[].judgement | ✅ | matched / partial / missing |
| risk_points[].severity | ✅ | high / medium / low |
| next_action.priority | ✅ | p0 / p1 / p2 |

## 10. 四档推荐质量检查

| job_id | recommend_level | recommend_label | match_score | 评估 |
|--------|----------------|-----------------|-------------|------|
| 9 | priority_apply | 优先投递 | 88 | ✅ 正确识别强匹配 |
| 13 | not_recommended | 暂不建议 | 42 | ⚠️ 可考虑 explore_first |
| 10 | not_recommended | 暂不建议 | 35 | ✅ 正确识别弱匹配 |
| 11 | not_recommended | 暂不建议 | 33 | ✅ 正确识别弱匹配 |

**问题**：4 个岗位只使用了 2 个推荐等级（priority_apply 和 not_recommended），缺少 recommended 和 explore_first。原因：测试样本中 3 个岗位均为弱匹配算法/咨询方向，确实与后端开发差距较大。岗位 13（score=42）可争论应为 explore_first。

**结论**：推荐等级区分基本合理，但四档覆盖不完整，与测试样本覆盖有关。

## 11. 证据链质量检查

| job_id | evidence 数量 | judgement 类型 | 是否有 missing | 评估 |
|--------|-------------|---------------|-------------|------|
| 9 | 4 | matched/partial/missing | ✅ | 正确识别 Go/MySQL 匹配，Redis 部分匹配，TCP/UDP 缺失 |
| 13 | 3 | partial/missing | ✅ | 正确识别沟通能力、商业意识缺失 |
| 10 | 3 | partial/missing | ✅ | 正确识别 CV/DL 缺失 |
| 11 | 3 | partial/missing | ✅ | 正确识别 NLP/DL 缺失 |

**结论**：证据链质量好。模型正确使用 "简历中暂未看到直接证据" 表达缺失证据，没有编造简历内容。

## 12. query_suggestions 检查

| job_id | 数量 | 示例 | 是否具体 | 评估 |
|--------|------|------|---------|------|
| 9 | 3 | Go 后端开发 校招 面试题、MySQL Redis 后端项目 简历怎么写、TCP UDP 网络编程 面试 | ✅ | 具体且相关 |
| 13 | 2 | 解决方案架构 实习 日常工作、行业咨询 岗位 技能要求 | ✅ | 具体且相关 |
| 10 | 2 | 多模态算法 实习 校招 要求、计算机视觉 入门 学习路线 | ✅ | 具体且相关 |
| 11 | 2 | 自然语言处理 实习 校招 要求、NLP 入门 学习路线 项目实践 | ✅ | 具体且相关 |

**结论**：query_suggestions 质量好，查询词具体，没有空泛表达。

## 13. next_action 检查

| job_id | priority | action | 是否具体 | 评估 |
|--------|----------|--------|---------|------|
| 9 | p0 | 梳理并补充简历中 Go 后端项目的量化指标，增加网络协议相关的学习经历描述 | ✅ | 具体且可执行 |
| 13 | p2 | 先查询该岗位的日常工作内容和能力要求，判断是否长期感兴趣 | ✅ | 合理 |
| 10 | p2 | 先查询多模态算法岗位的日常工作内容和入门要求 | ✅ | 合理 |
| 11 | p2 | 先查询 NLP 算法岗位的日常工作内容和入门要求 | ✅ | 合理 |

**结论**：next_action 质量好，优先级合理，动作具体。

## 14. Offer捕手语气检查

| 检查项 | 结果 | 说明 |
|--------|------|------|
| 是否先给结论 | ✅ | 每条 offer_go_summary 先说推荐/不推荐 |
| 是否说明原因 | ✅ | 说明匹配/不匹配的具体原因 |
| 是否给下一步方向 | ✅ | 每条都有投递前建议 |
| 是否亲和但不卖萌 | ✅ | 无卖萌表达 |
| 是否避免绝对化表达 | ✅ | 无"绝对适合""一定通过"等表达 |
| 是否专业克制 | ✅ | 判断基于证据，不过度鼓励 |

## 15. 专业性检查

| 检查项 | 结果 | 说明 |
|--------|------|------|
| HR 初筛视角 | ✅ | 正确识别简历与岗位的匹配/不匹配点 |
| 职业规划视角 | ✅ | strategy_summary 明确主攻方向和了解方向 |
| 技术岗位理解视角 | ✅ | 正确区分后端开发与算法岗位的核心差异 |
| 整体投递策略 | ✅ | strategy_summary 涵盖主攻和了解方向 |
| 整体风险 | ✅ | overall_risks 指出网络协议缺失和非技术岗位风险 |
| 简历优化重点 | ✅ | resume_focus 给出 3 条具体优化建议 |
| Demo 展示价值 | ✅ | 输出结构化程度高，适合前端展示 |

## 16. 真实性检查

| 检查项 | 结果 | 说明 |
|--------|------|------|
| authenticity_notice | ✅ | 包含"不要虚构经历，不要编造项目指标"等提示 |
| 不编造岗位 | ✅ | 所有 job_id 均来自输入 |
| 不编造经历 | ✅ | 缺失证据使用"简历中暂未看到直接证据" |
| 不承诺通过初筛 | ✅ | 无绝对承诺 |
| 不编造项目数据 | ✅ | 无超出简历事实的编造 |

## 17. 发现的问题

### 问题 1：四档推荐覆盖不完整

- **现象**：4 个岗位只使用了 priority_apply 和 not_recommended 两档，缺少 recommended 和 explore_first
- **影响**：中等。未能完全展示四档推荐的区分能力
- **原因**：测试样本中 3/4 为弱匹配算法岗位，确实不适合推荐；岗位 13（score=42）可考虑 explore_first
- **建议**：4-1.4.6A Schema 定稿时不阻塞；后续可补充中匹配岗位测试样本

### 问题 2：岗位 13 分类偏保守

- **现象**：解决方案-行业咨询方向（score=42）被标记为 not_recommended，但该岗位要求"理工科专业、熟悉云计算"与简历有一定交集
- **影响**：轻微。可论证为 not_recommended 或 explore_first
- **建议**：不阻塞；Prompt 中 explore_first 的适用条件可进一步细化

## 18. 不记录内容

- 未记录完整 API Key
- 未记录完整请求体
- 未记录完整模型原始输出
- 未记录完整 JD 长文本
- 未记录完整简历长文本

## 19. 验收结论

**partial_pass**

模型调用成功，JSON 合法，核心字段完整，枚举值全部合法，证据链真实引用 JD 和简历，Offer捕手语气亲和专业，分数克制（88/42/35/33），无编造内容，authenticity_notice 包含完整性提醒。

存在轻微问题：四档推荐覆盖不完整（仅使用 2/4 档），岗位 13 分类偏保守。问题不阻塞进入 4-1.4.6A 推荐 Schema 定稿。

## 20. 是否建议进入 4-1.4.6A

✅ 建议进入 4-1.4.6A。

四档覆盖不完整与测试样本覆盖有关，不影响 Schema 定稿。后续接入真实 LLM Agent 时可补充更多样本测试。

## 21. 下一步建议

1. **4-1.4.6A**：人工定稿推荐 JSON Schema — 基于本次验收结果和 recommendation_v1.md 输出结构固定 Schema
2. **4-1.4.6B**：CodeBuddy 登记推荐 JSON Schema
3. 后续可补充中匹配岗位样本，验证 recommended 和 explore_first 档位输出质量
