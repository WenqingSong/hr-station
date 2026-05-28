# 岗位推荐 Agent

> 最后更新：2026-05-29（阶段 3-1 创建阶段 3 核心业务后端 Mock 方案）

## 当前状态

阶段 3 方案已创建，Mock 输出结构已规划。阶段 3-6 将实现 Mock Agent。

## 功能

基于用户简历内容，推荐匹配的岗位：

- 分析简历中的技能、方向、经验水平
- 从岗位库中筛选匹配岗位
- 计算推荐等级（强推荐/推荐/可尝试）
- 生成推荐依据说明

## 输入

| 字段 | 类型 | 说明 |
|------|------|------|
| resume_version | object | 当前简历版本数据 |
| jobs | array | 岗位列表 |

## 阶段 3 Mock 输出结构

```json
{
  "recommended_jobs": [
    {
      "job_id": 1,
      "title": "Java 后端开发工程师",
      "city": "北京",
      "category": "后端开发",
      "recommendation_reason": "简历中的 Java 和 Spring 技能与岗位要求高度匹配",
      "match_highlights": [
        "掌握 Java 核心技术",
        "有 Spring Boot 项目经验",
        "熟悉 MySQL 数据库"
      ],
      "risk_points": [
        "缺乏分布式系统经验"
      ],
      "priority": "strong_recommend"
    }
  ],
  "total": 3
}
```

### 推荐等级定义

| 等级 | 值 | 说明 |
|------|---|------|
| 强推荐 | `strong_recommend` | 匹配度高，建议优先投递 |
| 推荐 | `recommend` | 有一定匹配度，值得尝试 |
| 可尝试 | `try` | 部分匹配，可以作为备选 |

## 阶段 3 Mock 实现方式

1. 基于关键词的简单匹配（简历技能 vs 岗位 skill_keywords）
2. 预定义的推荐结果模板
3. 固定的推荐等级
4. 不请求外部模型
5. 不读取 LLM_API_KEY

## 阶段 4 真实 Agent

阶段 4 将替换为：
1. 真实 LLM API 调用
2. 基于 Prompt 的智能推荐
3. 更精准的匹配度计算
4. 更自然的推荐理由生成

## 关联文档

- [Agent总览.md](./Agent总览.md)
- [Prompt模板.md](./Prompt模板.md)
- [../04-接口文档/推荐接口.md](../04-接口文档/推荐接口.md)
- [../02-后端设计/阶段3核心业务后端Mock方案.md](../02-后端设计/阶段3核心业务后端Mock方案.md)
