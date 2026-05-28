# 岗位匹配 Agent

> 最后更新：2026-05-29（阶段 3-1 创建阶段 3 核心业务后端 Mock 方案）

## 当前状态

阶段 3 方案已创建，Mock 输出结构已规划。阶段 3-7 将实现 Mock Agent。**这是阶段 3 的核心亮点 Agent。**

## 功能

深度分析用户简历与目标岗位的匹配情况：

- 生成匹配度评分
- 构建 JD-简历证据链（逐条对照岗位要求与简历经历）
- 检测表达风险
- 给出 P0/P1 优化建议
- 不自动生成优化版简历

## 输入

| 字段 | 类型 | 说明 |
|------|------|------|
| resume_version | object | 当前简历版本数据 |
| job | object | 目标岗位 JD |

## 阶段 3 Mock 输出结构

```json
{
  "match_score": 72,
  "screening_result": "likely_pass",
  "evidence_chain": [
    {
      "jd_requirement": "熟练掌握 Java 核心技术",
      "resume_evidence": "项目中使用 Java 开发后端服务",
      "match_status": "matched",
      "detail": "简历中体现了 Java 开发经验"
    },
    {
      "jd_requirement": "熟悉分布式系统设计",
      "resume_evidence": null,
      "match_status": "missing",
      "detail": "简历中未体现分布式系统相关经验"
    }
  ],
  "risk_points": [
    {
      "type": "skill_gap",
      "description": "缺乏分布式系统经验，可能影响大型项目参与",
      "severity": "high"
    },
    {
      "type": "expression_risk",
      "description": "项目描述较为笼统，缺乏量化成果",
      "severity": "medium"
    }
  ],
  "optimization_suggestions": [
    {
      "priority": "P0",
      "suggestion": "补充分布式系统相关学习经验或项目实践",
      "target_field": "experience",
      "reason": "岗位要求中明确提及分布式系统"
    },
    {
      "priority": "P1",
      "suggestion": "在项目描述中增加量化指标（如：处理 QPS、用户量等）",
      "target_field": "project",
      "reason": "当前项目描述缺乏具体成果"
    }
  ],
  "next_steps": [
    "根据 P0 建议补充分布式系统经验",
    "优化项目描述，增加量化指标",
    "添加新简历版本后再次匹配"
  ],
  "authenticity_notice": "本分析基于您简历中的真实内容生成，不会为您虚构不存在的经历。优化建议旨在帮助您更好地呈现真实能力。"
}
```

### 证据链状态

| 状态 | 值 | 说明 |
|------|---|------|
| 匹配 | `matched` | 简历中有对应内容 |
| 部分匹配 | `partial_matched` | 简历中有部分对应内容 |
| 缺失 | `missing` | 简历中没有对应内容 |

### 风险类型

| 类型 | 值 | 说明 |
|------|---|------|
| 技能缺口 | `skill_gap` | 岗位要求的技能在简历中缺失 |
| 表达风险 | `expression_risk` | 描述模糊、缺乏量化、过于笼统 |
| 经验不足 | `experience_gap` | 相关经验不足 |

### 风险严重度

| 严重度 | 值 | 说明 |
|--------|---|------|
| 高 | `high` | 可能导致初筛不通过 |
| 中 | `medium` | 建议优化但不紧急 |
| 低 | `low` | 可选优化 |

### 优化建议优先级

| 优先级 | 值 | 说明 |
|--------|---|------|
| P0 | `P0` | 必须优化，否则初筛风险高 |
| P1 | `P1` | 建议优化，可提升匹配度 |

### 初筛判断

| 值 | 说明 |
|---|------|
| `likely_pass` | 匹配度较高，预计可通过初筛 |
| `borderline` | 匹配度一般，初筛有风险 |
| `unlikely_pass` | 匹配度较低，初筛可能不通过 |

## 阶段 3 Mock 实现方式

1. 基于关键词匹配的伪评分
2. 预定义的证据链模板
3. 固定的风险点和优化建议
4. 不请求外部模型
5. 不读取 LLM_API_KEY
6. 必须包含真实性提示

## 关键约束

1. **必须有 JD-简历证据链**：逐条对照岗位要求与简历经历
2. **必须有初筛风险**：识别可能导致初筛失败的问题
3. **必须有 P0/P1 优化建议**：给出明确改进方向
4. **不生成优化版简历**：只给诊断和建议，用户手动修改
5. **必须有真实性提示**：明确告知用户不会虚构经历

## 阶段 4 真实 Agent

阶段 4 将替换为：
1. 真实 LLM API 调用
2. 基于 Prompt 的深度分析
3. 更精准的证据链构建
4. 更个性化的优化建议

## 关联文档

- [Agent总览.md](./Agent总览.md)
- [Prompt模板.md](./Prompt模板.md)
- [../04-接口文档/匹配分析接口.md](../04-接口文档/匹配分析接口.md)
- [../02-后端设计/阶段3核心业务后端Mock方案.md](../02-后端设计/阶段3核心业务后端Mock方案.md)
