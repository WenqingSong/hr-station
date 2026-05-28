# Prompt 模板

## 当前状态

待设计

## 模板管理规范

所有 Prompt 模板统一管理，便于版本控制和调优。

### 文件结构

```
server/internal/agent/prompts/
├── resume_extract.md
├── job_recommend.md
├── job_match.md
└── version_compare.md
```

### 模板格式

使用 Go template 语法，变量用 `{{.Variable}}` 标记。

## 设计原则

1. **不虚构**：Prompt 中明确要求不编造不存在的经历
2. **有证据**：所有结论必须基于简历中的真实内容
3. **可追溯**：每条建议都要指出对应的简历原文
4. **有建设性**：不仅指出问题，还给出改进方向

## 后续待补充

- 各 Agent 的 Prompt 模板内容
- 模板版本管理方案
- Prompt 调优记录

## 关联文档

- [Agent总览.md](./Agent总览.md)
- [模型接入.md](./模型接入.md)
