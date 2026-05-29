# 岗位匹配Agent

> 最后更新：2026-05-29（阶段 4-1.4.1B 登记匹配分析 Agent 指令模板）

## 1. 当前实现状态

阶段 3-7 已完成，岗位匹配 Mock Agent 已实现（`server/internal/agent/match_mock.go`）。Mock Agent 使用简单规则打分生成结构化匹配分析结果，包含 JD-简历证据链、初筛判断、风险点、P0/P1 优化建议、下一步建议和真实性提示。不调用真实 LLM。

## 2. 阶段 3 Mock Agent 目标

1. 生成匹配度评分（0-100）
2. 构建 JD-简历证据链（逐条对照岗位要求与简历经历）
3. 生成初筛判断
4. 检测风险点
5. 给出 P0/P1 优化建议
6. 给出下一步建议
7. 包含真实性提示
8. 不自动生成优化版简历

## 3. 输入数据

| 字段 | 类型 | 说明 |
|------|------|------|
| resume_version | object | 当前简历版本数据 |
| job | object | 目标岗位 JD |

## 4. 输出结构

```json
{
  "job_id": 1,
  "job_title": "Go 后端开发工程师",
  "resume_version_id": 2,
  "resume_version_no": 2,
  "match_score": 78,
  "screening_result": "likely",
  "evidence_chain": [...],
  "match_highlights": [...],
  "risk_points": [...],
  "optimization_suggestions": [...],
  "next_steps": [...],
  "authenticity_notice": "以下建议仅用于表达优化，请只补充真实经历，不要虚构项目或能力。"
}
```

## 5. Mock 匹配规则

| 规则 | 加分 | 说明 |
|------|------|------|
| target_city = job.city | +15 | 城市匹配 |
| target_role 与 category/title 相关 | +25/+8 | 方向匹配 |
| 技能交集（每项） | +5 | 最大 35 |
| 高技能重叠奖励 | +5 | 交集 > 5 |
| 关键技术同时匹配 | +3/项 | 最大 15 |
| 有实习/工作经历 | +5 | |
| 有项目经历 | +5 | |

### 证据链判断

| 判断 | 值 | 说明 |
|------|---|------|
| 强匹配 | `strong_match` | 简历中有明确对应内容 |
| 部分匹配 | `partial_match` | 简历中有部分对应内容 |
| 缺失 | `missing` | 简历中没有对应内容 |

### 初筛判断

| 分数 | 判断 | 说明 |
|------|------|------|
| 85-100 | `highly_likely` | 匹配度高 |
| 70-84 | `likely` | 匹配度较高 |
| 55-69 | `possible` | 匹配度一般 |
| 0-54 | `risky` | 匹配度低 |

### 优化建议优先级

| 优先级 | 说明 |
|--------|------|
| P0 | 必须优化，否则初筛风险高 |
| P1 | 建议优化，可提升匹配度 |

## 6. JD-简历证据链设计

证据链是本 Agent 的核心输出，逐条对照岗位要求与简历经历：

1. 对岗位的每个技能关键词，检查简历中是否有对应内容
2. 对城市要求，检查简历 target_city 是否匹配
3. 对方向要求，检查简历 target_role 是否匹配
4. 每条证据包含：requirement（岗位要求）、resume_point（简历证据）、judgement（判断）

## 7. P0/P1 优化建议设计

- **P0**：基于缺失的关键技能或岗位明确要求但简历缺失的内容
- **P1**：基于可提升的量化指标或工作地点不一致等问题
- 每条建议包含：priority、title、reason、action
- action 是可操作的具体建议，不是模糊的"提升自己"

## 8. 真实性提示

每条匹配分析结果都包含 authenticity_notice 字段，明确告知用户：
- 建议仅用于表达优化
- 请只补充真实经历
- 不要虚构项目或能力

## 9. 当前不调用真实模型

1. 不读取 LLM_API_KEY
2. 不调用外部 API
3. 不使用真实 Prompt
4. 结果基于简单规则打分
5. 阶段 4 才替换为真实 LLM

## 10. 4-1.3 适配方向

4-1.3.6 已完成，匹配分析已使用腾讯岗位 description/requirements：

1. 匹配分析已使用腾讯岗位完整 JD（description / requirements）
2. data_quality=full_jd 的岗位支持深度证据链，匹配结果包含 analysis_basis
3. summary_only 岗位返回降级提示，建议补充完整 JD 后再分析
4. 匹配服务仍只读取本地 jobs 表
5. 适配方案见 `docs/02-后端设计/阶段4-1.3腾讯招聘信息适配方案.md`
6. generated_by 仍为 mock_agent

## 11. Prompt Template 登记状态

- 匹配分析 Agent 后续将使用 `server/prompts/match_analysis_v1.md`
- 当前该 Prompt Template 已登记（4-1.4.1A 人工定稿，4-1.4.1B CodeBuddy 登记）
- 模板包含 HR 初筛 + 技术面试官双视角、Offer鹅语气、证据链、P0/P1/P2 优化建议、面试追问预测和真实性提醒
- 当前仍使用 Mock Agent
- 后续 4-1.4.2 将固定匹配分析 JSON Schema
- 后续 4-1.4.3 将进行匹配分析 Prompt 首轮输出验收

## 12. 后续真实 Agent 设计方向

1. 基于真实 Prompt 的深度分析
2. 更精准的证据链构建
3. 更个性化的优化建议
4. 支持 NLP 理解岗位描述和简历内容
5. 异步分析支持
