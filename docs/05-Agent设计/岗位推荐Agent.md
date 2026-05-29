# 岗位推荐 Agent

> 最后更新：2026-05-29（阶段 4-1.3.0 腾讯招聘信息适配方案）

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

4-1.3 只让推荐 Mock Agent 使用更真实的岗位数据，不接 Hunyuan：

1. 推荐候选池从 8 条种子岗位扩展到 8+186 条（含腾讯岗位）
2. data_quality=full_jd 的岗位优先参与推荐
3. summary_only 岗位降权或仅做轻量推荐
4. 推荐服务仍只读取本地 jobs 表
5. 适配方案见 `docs/02-后端设计/阶段4-1.3腾讯招聘信息适配方案.md`

## 8. 后续真实 Agent 设计方向

阶段 4 将替换为真实 LLM：

1. 使用 Prompt 模板构建推荐请求
2. 调用 LLM API 生成更自然的推荐理由
3. 基于完整简历内容做深度分析
4. 生成更精准的匹配评估
5. 保持输出结构不变，便于无缝替换

## 关联文档

- [Agent总览.md](./Agent总览.md)
- [Prompt模板.md](./Prompt模板.md)
- [../04-接口文档/推荐接口.md](../04-接口文档/推荐接口.md)
- [../02-后端设计/阶段3核心业务后端Mock方案.md](../02-后端设计/阶段3核心业务后端Mock方案.md)
