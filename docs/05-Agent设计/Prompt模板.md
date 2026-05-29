# Prompt 模板

> 最后更新：2026-05-29（阶段 4-1.4.2 匹配分析 Prompt 首轮输出验收）

## 当前状态

阶段 4-1.4.2 已完成，匹配分析 Prompt 首轮输出验收结论：**partial_pass**。

Prompt 将作为独立产品资产管理：
- 目录 `server/prompts/`（替代原 `server/internal/agent/prompts/` 规划）
- ✅ `match_analysis_v1.md` 已登记 — Match Analysis Agent Prompt Template（4-1.4.1A 人工定稿，4-1.4.1B CodeBuddy 登记）
  - 模板包含 HR 初筛 + 技术面试官双视角
  - 模板包含 Offer鹅展示语气、证据链、P0/P1/P2 优化建议、面试追问预测和真实性提醒
  - 模板要求严格 JSON 输出，包含完整 JSON 结构定义和枚举约束
  - 当前尚未接入 Go 后端和 Hunyuan
  - **4-1.4.2 首轮输出验收结论：partial_pass**
    - JSON 输出合法，17 个顶层字段完整，枚举值全部合法
    - 证据链真实引用 JD 和简历，Offer鹅语气亲和专业，分数克制（72 分）
    - 发现问题：rewrite_example 中少量具体数字超出简历事实，证据链缺少 missing 判断
    - 问题不阻塞进入 4-1.4.3A JSON Schema 定稿
    - 验收报告见 `docs/05-Agent设计/匹配分析Prompt首轮输出验收报告.md`
- 4-1.4.4 开始创建推荐 Prompt 文件 `recommendation_v1.md`
- 4-1.4.7 开始创建简历优化 Prompt 文件 `resume_optimization_v1.md`
- JSON 修复 Prompt `json_repair_v1.md` 为 P2 优先级，后续按需创建
- JSON Schema 文件将存放在 `server/prompts/schema/` 目录
- 方案文档见 `docs/05-Agent设计/阶段4-1.4Prompt与Agent真实接入方案.md`

后续 Prompt 将以腾讯真实岗位 JD 为输入：
- 推荐 Prompt 和匹配 Prompt 需要分别设计
- JSON 输出 Schema 需要固定
- 失败兜底和 Mock 回退需要保留
- data_quality 分层已建立，Prompt 可针对不同质量等级差异化处理

## 模板管理规范

所有 Prompt 模板统一管理，便于版本控制和调优。

### 文件结构

```
server/internal/agent/prompts/
├── resume_extract.md     # 简历提取 Prompt（阶段 4）
├── job_recommend.md      # 岗位推荐 Prompt（阶段 4）
├── job_match.md          # 岗位匹配 Prompt（阶段 4）
└── version_compare.md    # 版本对比 Prompt（阶段 4）
```

### 模板格式

使用 Go template 语法，变量用 `{{.Variable}}` 标记。

## 设计原则

1. **不虚构**：Prompt 中明确要求不编造不存在的经历
2. **有证据**：所有结论必须基于简历中的真实内容
3. **可追溯**：每条建议都要指出对应的简历原文
4. **有建设性**：不仅指出问题，还给出改进方向

## 阶段 4 Prompt 必须包含的约束

1. **只输出 JSON**：必须要求模型只输出 JSON，不要 Markdown、不要解释性前后缀、不要代码块
2. **禁止事项**：不虚构经历、不编造不存在的项目或能力、不生成优化版简历
3. **真实性约束**：所有分析必须基于用户简历中实际存在的内容
4. **分数范围**：分数必须是 0-100 整数
5. **必填字段**：明确列出输出 JSON 中必须包含的字段
6. **JSON 结构示例**：在 Prompt 中提供完整的 JSON 输出结构示例

### Prompt 通用结构

```
你是一个简历分析助手。

## 约束
1. 只输出 JSON，不要输出任何其他内容
2. 不要输出 Markdown 格式
3. 不要输出解释性前后缀
4. 不要输出代码块包裹
5. 不要虚构用户简历中不存在的经历
6. 不要自动生成优化版简历
7. 所有分析必须基于用户简历中实际存在的内容

## 输出格式
{JSON 结构示例}

## 输入数据
{业务数据}

## 分析要求
{具体分析要求}
```

## 各 Agent Prompt 方向

### 简历提取 Prompt

- 输入：结构化简历数据
- 输出：关键技能标签、求职方向推断、表达风险列表
- 阶段 4 暂不独立实现

### 岗位推荐 Prompt

- 输入：简历提取结果 + 岗位列表
- 输出：推荐岗位列表 + 推荐理由 + 推荐等级
- 4-5 实现

### 岗位匹配 Prompt

- 输入：简历数据 + 目标岗位 JD
- 输出：匹配分数 + 证据链 + 风险点 + 优化建议
- **核心 Prompt**，4-4 重点实现
- 必须包含 JD-简历证据链构建要求

### 版本对比 Prompt

- 输入：旧版本 + 新版本简历
- 输出：差异分析 + 改进评价 + 仍需改进
- 4-6 实现

## 后续待补充

- 各 Agent 的真实 Prompt 模板内容（4-4 / 4-5 / 6 实现时编写）
- 模板版本管理方案
- Prompt 调优记录

## 关联文档

- [Agent总览.md](./Agent总览.md)
- [阶段4Agent与模型接入方案.md](./阶段4Agent与模型接入方案.md)
- [LLMClient设计.md](./LLMClient设计.md)
- [JSON输出规范.md](./JSON输出规范.md)
- [模型接入.md](./模型接入.md)
- [../02-后端设计/阶段3核心业务后端Mock方案.md](../02-后端设计/阶段3核心业务后端Mock方案.md)
