# Agent 总览

## 当前状态

待设计（前期使用 Mock Agent）

## Agent 列表

| Agent | 说明 | 当前状态 | 详细文档 |
|-------|------|---------|---------|
| 简历提取 Agent | 从结构化简历中提取关键信息 | Mock | [简历提取Agent.md](./简历提取Agent.md) |
| 岗位推荐 Agent | 基于简历推荐匹配岗位 | Mock | [岗位推荐Agent.md](./岗位推荐Agent.md) |
| 岗位匹配 Agent | 分析简历与岗位的匹配度 | Mock | [岗位匹配Agent.md](./岗位匹配Agent.md) |
| 版本对比 Agent | 对比两个版本的差异 | Mock | [版本对比Agent.md](./版本对比Agent.md) |

## 设计原则

1. **先 Mock 后接模型**：前期使用硬编码/规则引擎的 Mock Agent，跑通产品闭环
2. **统一接口**：所有 Agent 遵循统一的输入输出接口，方便后续替换
3. **Prompt 管理**：Prompt 模板独立管理，便于调优和版本控制
4. **真实性保护**：Agent 不虚构经历，不替用户编造内容

## 后续待补充

- Agent 统一接口定义
- Mock Agent 实现方案
- 真实模型接入方案

## 关联文档

- [模型接入.md](./模型接入.md)
- [Prompt模板.md](./Prompt模板.md)
- [../08-架构决策记录/ADR-003-先Mock后接模型.md](../08-架构决策记录/ADR-003-先Mock后接模型.md)
