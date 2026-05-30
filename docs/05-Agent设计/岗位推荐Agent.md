# 岗位推荐 Agent

> 最后更新：2026-05-30（阶段 4-3.0 业务 Agent 接入 LLM 方案）

## 1. 当前实现状态

阶段 3-6 已完成，Mock Agent 已实现。文件：`server/internal/agent/recommendation_mock.go`。当前使用简单规则打分生成推荐结果，不调用真实 LLM。

## 2. 阶段 3 Mock Agent 目标

基于用户简历内容和岗位列表，生成结构化的推荐结果：

- 从 jobs 表获取所有岗位
- 基于简历的 target_city / target_role / skills_json 与岗位进行规则打分
- 按分数排序取 top 5
- 生成推荐等级（strong_recommend / recommend / try）
- 生成推荐理由、匹配亮点、风险点、建议动作
- 结果有产品感，便于前端展示

## 3. 输入数据

| 字段 | 类型 | 说明 |
|------|------|------|
| resume | ResumeVersion | 当前简历版本数据 |
| jobs | []*Job | 岗位列表 |

## 4. 输出结构

```json
[
  {
    "job_id": 1,
    "title": "Go 后端开发工程师",
    "department": "技术部",
    "city": "深圳",
    "category": "后端开发",
    "job_type": "校招",
    "priority": "strong_recommend",
    "recommendation_score": 89,
    "recommendation_reason": "目标城市与岗位城市一致...",
    "match_highlights": ["目标城市一致", "技能交集"],
    "risk_points": ["Redis 经验可进一步强化"],
    "suggested_action": "建议优先查看该岗位详情..."
  }
]
```

## 5. Mock 推荐规则

### 打分规则

| 规则 | 加分 | 说明 |
|------|------|------|
| target_city 与 job.city 一致 | +25 | 城市匹配 |
| target_role 与 job.category 相关 | +30 | 方向匹配 |
| target_role 关键词在 title/category 中出现 | +10 | 关键词匹配 |
| 技能交集（每项） | +5 | 最大 30 |
| 高技能重叠奖励 | +5 | 交集 > 6 |
| 关键技术同时匹配（Go/MySQL/Redis/Docker/K8s 等） | +3/项 | 额外加分 |

### 优先级生成

| 优先级 | 分数范围 | 值 |
|--------|---------|-----|
| 强推荐 | ≥50 | `strong_recommend` |
| 推荐 | 30-49 | `recommend` |
| 可尝试 | <30 | `try` |

### 推荐理由生成

基于实际匹配项动态拼接：
- 城市一致 → 提及城市匹配
- 方向一致 → 提及方向匹配
- 技能交集 → 列出关键重叠技能

### 匹配亮点生成

- 城市匹配 → "目标城市与简历 target_city 一致"
- 方向匹配 → "目标岗位与简历 target_role 接近"
- 技能交集 → "技能关键词与 JD 要求存在交集"
- 有实习经历 → "项目经历中包含相关实习或工作经验"

### 风险点生成

- 找出岗位要求但简历缺失的技能
- 城市不一致 → 提及工作地点与目标城市不同
- 无明显风险时 → "项目经历偏校园级别"

### 建议动作生成

- strong_recommend → 建议优先查看并做匹配分析
- recommend → 建议关注并了解具体要求
- try → 建议补充技能后再关注

## 6. 当前不调用真实模型

- 不读取 LLM_API_KEY
- 不调用外部 API
- 不使用真实 Prompt
- 不生成简历内容
- 所有推荐基于本地规则计算

## 7. 4-1.3 适配方向

4-1.3.6 已完成，推荐候选池已纳入 tencent_join 岗位：

1. 推荐候选池从 8 条种子岗位扩展到 8+186 条（含腾讯岗位）
2. data_quality=full_jd 的岗位优先参与推荐，基于 data_quality 评分
3. summary_only 岗位降权（-10）
4. 推荐服务仍只读取本地 jobs 表
5. JD keyword matching 已适配 description/requirements
6. 推荐结果新增 external_source / external_id / data_quality / source_url 响应字段
7. generated_by 仍为 mock_agent
8. 适配方案见 `docs/02-后端设计/阶段4-1.3腾讯招聘信息适配方案.md`

## 8. 后续真实 Agent 设计方向

阶段 4 将替换为真实 LLM：

1. 使用 Prompt 模板构建推荐请求
2. 调用 LLM API 生成更自然的推荐理由
3. 基于完整简历内容做深度分析
4. 生成更精准的匹配评估
5. 保持输出结构不变，便于无缝替换

## 9. Prompt Template 登记状态

- 推荐 Agent 后续将使用 `server/prompts/recommendation_v1.md`
- 当前该 Prompt Template 已登记（4-1.4.4A 人工定稿，4-1.4.4B CodeBuddy 登记）
- 模板面向"不知道选择哪个岗位"的学生求职场景
- 模板采用四档推荐：优先投递 / 建议尝试 / 先了解再决定 / 暂不建议
- 模板包含 Offer鹅展示语气、推荐理由、证据链、风险点、查询建议、下一步行动和真实性提醒
- 当前仍使用 Mock Agent
- 当前仍未接入 Go 后端和 Hunyuan
- **4-1.4.5 已完成：推荐 Prompt 首轮输出验收（partial_pass）**
  - 验收报告见 `docs/05-Agent设计/推荐Prompt首轮输出验收报告.md`
  - 模型输出 JSON 合法，字段完整，证据链真实，Offer鹅语气专业
  - 发现轻微问题：四档推荐覆盖不完整，不阻塞后续 Schema 定稿
- 推荐 Agent 后续将使用：
  - Prompt Template：`server/prompts/recommendation_v1.md`
  - Output Schema：`server/prompts/schema/recommendation_schema.json`
- 当前 Prompt 和 Schema 均已登记
- 当前仍使用 Mock Agent
- 当前仍未接入 LLM Client
- 后续简历优化 Prompt 将继续采用人工定稿 + CodeBuddy 登记模式

## 关联文档

- [Agent总览.md](./Agent总览.md)
- [Prompt模板.md](./Prompt模板.md)
- [../04-接口文档/推荐接口.md](../04-接口文档/推荐接口.md)
- [../02-后端设计/阶段3核心业务后端Mock方案.md](../02-后端设计/阶段3核心业务后端Mock方案.md)

## 阶段 4-3.0 接入方案

4-3.0 已完成 RecommendationAgent 接入 LLM 方案设计（docs only），详细方案见 `docs/05-Agent设计/阶段4-3业务Agent接入LLM方案.md`。

关键接入设计：

1. **Service 层注入 pipelineRunner**：RecommendationService 新增 pipelineRunner 字段，构造函数注入
2. **PromptValues 组装**：Service 从数据库读取简历和岗位列表后，组装 `RESUME_CONTEXT_JSON`、`JOB_CANDIDATES_JSON`、`BUSINESS_CONTEXT` 三个占位符
3. **Pipeline 调用**：AgentName=recommendation，PromptName=recommendation，SchemaName=recommendation
4. **RunResult 处理**：success/repaired → 将 RunResult.Parsed 映射为 []model.RecommendedJob；fallback_required → 调用 s.mockAgent.Recommend()
5. **字段映射策略**：map[string]any → typed struct，映射失败触发 Mock fallback
6. **generated_by 规则**：success/repaired → llm_agent，fallback_required → mock_agent
7. **4-3.4 将实现本接入方案**

当前仍使用 Mock Agent，当前仍未接入 Pipeline。

## 阶段 4-3.4 RecommendationAgent 接入 Pipeline

4-3.4 已完成 RecommendationAgent 接入统一 Pipeline。关键实现：

1. **Recommender 接口抽象**：`Recommend(ctx, resume, jobs)` — RecommendationMockAgent 和 RecommendationAgent 共享同一接口
2. **PromptValues 组装**：RESUME_CONTEXT_JSON / JOB_CANDIDATES_JSON / BUSINESS_CONTEXT 三个占位符，与 recommendation_v1.md 对应
3. **Pipeline 调用**：AgentName=RecommendationAgent，PromptName=recommendation，SchemaName=recommendation
4. **候选岗位校验**：LLM 返回的 job_id 必须在候选集中，不存在的 job_id 被过滤；全部过滤后触发 fallback
5. **字段映射**：map[string]any → RecommendedJob，推荐级别映射（priority_apply→strong_recommend、recommended→recommend、explore_first→try）
6. **Job 字段来自候选集**：title/department/city/ExternalSource 等来自数据库候选岗位，不采用 LLM 输出
7. **RecFallbackError**：Pipeline fallback_required 或映射失败时返回，Service 层捕获后调用 Mock
8. **generated_by 规则**：success/repaired → llm_agent，fallback_required → mock_agent
9. **Mock Agent 签名更新**：Recommend 方法新增 context.Context 参数以实现 Recommender 接口
10. **RecommendedJob 新增字段**（omitempty）：RecommendLabel / OfferGoSummary / QuerySuggestions / GeneratedBy / AgentStatus / RepairApplied / FallbackReason / ErrorCode
11. **RecommendationService 改造**：注入 pipelineRunner + llmEnabled，LLM 优先 + Mock 回退，LLM 路径保存快照
12. **12 项 Agent 层测试通过**：LLM 成功、repaired、fallback_required、无效 job_id 过滤、空推荐列表、混合 job_id、PromptValues 组装、安全边界、推荐级别映射、候选字段来源校验等
