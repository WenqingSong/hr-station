# 版本对比 Agent

> 最后更新：2026-05-29（阶段 3-1 创建阶段 3 核心业务后端 Mock 方案）

## 当前状态

阶段 3 方案已创建，Mock 输出结构已规划。阶段 3-8 将实现 Mock Agent。

## 功能

对比两个简历版本的差异：

- 逐字段对比差异
- 分析变化原因
- 评估改进效果
- 指出仍需改进的地方
- 服务于"分析 → 修改 → 新版本 → 对比"的产品闭环

## 输入

| 字段 | 类型 | 说明 |
|------|------|------|
| from_version | object | 旧版本简历数据 |
| to_version | object | 新版本简历数据 |

## 阶段 3 Mock 输出结构

```json
{
  "from_version": {
    "version_no": 1,
    "name": "张三"
  },
  "to_version": {
    "version_no": 2,
    "name": "张三"
  },
  "score_change": {
    "from_score": 60,
    "to_score": 72,
    "change": "+12"
  },
  "field_changes": [
    {
      "field": "experience",
      "change_type": "added",
      "description": "新增了一段实习经历"
    },
    {
      "field": "project",
      "change_type": "modified",
      "description": "项目描述增加了量化指标"
    }
  ],
  "risk_changes": [
    {
      "risk": "项目描述缺乏量化指标",
      "from_severity": "medium",
      "to_severity": "resolved",
      "description": "项目描述已补充量化指标"
    }
  ],
  "suggestion_completion": [
    {
      "suggestion": "补充分布式系统相关学习经验或项目实践",
      "priority": "P0",
      "status": "completed",
      "description": "已在实习经历中补充相关内容"
    }
  ],
  "summary": "V2 相比 V1 在项目描述和实习经历方面有明显改进，匹配度从 60 提升至 72。建议继续优化技能描述。"
}
```

### 字段变化类型

| 类型 | 值 | 说明 |
|------|---|------|
| 新增 | `added` | 新增了字段内容 |
| 修改 | `modified` | 修改了字段内容 |
| 删除 | `removed` | 删除了字段内容 |

### 风险变化说明

| 严重度变化 | 说明 |
|-----------|------|
| `medium` → `resolved` | 风险已解决 |
| `high` → `medium` | 风险有所缓解 |
| `high` → `high` | 风险未改善 |
| `resolved` → `medium` | 新增风险 |

### 建议完成状态

| 状态 | 值 | 说明 |
|------|---|------|
| 已完成 | `completed` | 已完成该建议 |
| 部分完成 | `partially_completed` | 部分完成 |
| 未开始 | `not_started` | 未开始 |

## 阶段 3 Mock 实现方式

1. 基于字段值的简单对比
2. 预定义的改进效果评价
3. 固定的改进建议
4. 不请求外部模型
5. 不读取 LLM_API_KEY

## 阶段 4 真实 Agent

阶段 4 将替换为：
1. 真实 LLM API 调用
2. 基于 Prompt 的深度对比
3. 更精准的变化分析
4. 更个性化的改进建议

## 关联文档

- [Agent总览.md](./Agent总览.md)
- [Prompt模板.md](./Prompt模板.md)
- [../04-接口文档/匹配分析接口.md](../04-接口文档/匹配分析接口.md)
- [../02-后端设计/阶段3核心业务后端Mock方案.md](../02-后端设计/阶段3核心业务后端Mock方案.md)
