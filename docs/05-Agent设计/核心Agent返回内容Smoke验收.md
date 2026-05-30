# 核心 Agent 返回内容 Smoke 验收

> 创建时间：2026-05-31
> 所属阶段：4-3.7-fix.4
> 验收人：CodeBuddy

## 1. 背景

4-3.7-fix 系列已完成 fix.1 ~ fix.3，确认 Prompt / Schema / Go 契约一致，最终闸门验收已通过，可进入 4-3.8。

但在正式进入 4-3.8 三类业务 Agent 集成验收之前，用户希望增加一个保险阶段：确认核心 Agent 能正常返回内容，同时固化 3 个固定测试简历作为后续稳定输入。

本阶段不是三类 Agent 全量验收，只做一个核心 Agent 的本地确定性路径 Smoke。

## 2. 验收目标

| 验收项 | 说明 |
|--------|------|
| 3 个固定测试简历 | 已创建，不包含真实隐私，覆盖三种典型情况 |
| MatchAnalysisAgent Smoke | 本地 Mock Agent 确定性路径，不调用真实 LLM |
| 返回内容结构 | 包含 apply_readiness / optimization_stop_signal / top_fix_items |
| 匹配依据 / 证据链 | 非空 |
| generated_by 标识 | 不破坏 |

## 3. 固定测试简历

| 文件 | 定位 | 匿名姓名 | 典型特征 |
|------|------|----------|----------|
| `resume_01_go_backend_strong.json` | Go 后端方向，较强匹配 | 测试候选人A | 本科计算机、Go/MySQL/Redis/Docker/JWT、2个项目、1段实习 |
| `resume_02_backend_weak_insufficient.json` | 后端方向意向明确，证据不足 | 测试候选人B | 本科软件工程、Go/Java/MySQL 但无项目细节、只有课程作业 |
| `resume_03_non_internet_mismatch.json` | 非互联网方向，与后端岗位不匹配 | 测试候选人C | 本科市场营销、Office/沟通/策划、无编程项目 |

所有测试简历均为虚构数据，使用 example.com 邮箱，不含真实手机号/身份证/学校真名。

## 4. 本次测试 Agent

| Agent | 路径 | 是否调用真实 LLM |
|-------|------|-------------------|
| MatchAnalysisAgent | MatchMockAgent（本地确定性规则打分） | 否 |

本阶段不测试 RecommendationAgent 和 ResumeOptimizationAgent，留待 4-3.8 统一验收。

## 5. 固定 JD

```
岗位：后台开发实习生
部门：技术部
城市：深圳
方向：后端开发

描述：参与后台系统开发与维护，负责接口设计和实现。

要求：
1. 熟悉 Go 或 Java 至少一种后端语言；
2. 熟悉 HTTP API 开发；
3. 熟悉 MySQL，了解索引、事务；
4. 了解 Redis、缓存、登录鉴权优先；
5. 有项目开发、问题排查、接口联调经验；
6. 具备良好代码习惯和学习能力。

技能关键词：Go, MySQL, Redis, HTTP API, Docker
```

## 6. Smoke 断言

| 序号 | 断言 | 结果 |
|------|------|------|
| 1 | Agent 调用不 panic | ✅ PASS |
| 2 | 返回结果非 nil | ✅ PASS |
| 3 | 返回内容非空（match_score > 0 或 screening_result 非空） | ✅ PASS（match_score=92, screening_result=highly_likely） |
| 4 | apply_readiness 非空 | ✅ PASS（ready_to_apply） |
| 5 | optimization_stop_signal 非空 | ✅ PASS（stop_and_apply） |
| 6 | top_fix_items 字段存在 | ✅ PASS（1 items） |
| 7 | 至少一个 evidence_chain / match_highlights 非空 | ✅ PASS（evidence_chain=7, match_highlights=3） |
| 8 | generated_by 不破坏 | ✅ PASS（Mock 不设置，不为非法值） |

## 7. 测试结果

### Smoke 测试输出

```
=== RUN   TestMatchAnalysisAgentReturnsDecisionFieldsSmoke
    apply_readiness = "ready_to_apply"
    optimization_stop_signal = "stop_and_apply"
    top_fix_items count = 1
    evidence_chain count = 7, match_highlights count = 3
    === Smoke Test Summary ===
    match_score = 92
    screening_result = highly_likely
    apply_readiness = ready_to_apply
    optimization_stop_signal = stop_and_apply
    top_fix_items = 1 items
    evidence_chain = 7 items
--- PASS: TestMatchAnalysisAgentReturnsDecisionFieldsSmoke (0.00s)
```

### Go 测试

```
go test ./... — 全部通过
```

### LLM 配置检查

```
make llm-check — 可用，未调用真实 LLM
```

## 8. 当前限制

1. **本阶段只测试了 MatchAnalysisAgent 的本地确定性（Mock）返回路径**
2. **不代表真实 LLM 质量**：Mock Agent 使用规则打分，不调用真实模型
3. **真实 LLM + fallback 行为放到 4-3.8 统一验收**
4. 只用了 resume_01_go_backend_strong 一份简历做 Smoke，其余两份简历已沉淀为测试资产
5. Mock Agent 不设置 GeneratedBy 字段（空字符串），这是预期行为

## 9. 是否可以进入 4-3.8

**可以进入 4-3.8。**

理由：

1. ✅ 核心 Agent 返回内容 Smoke 通过
2. ✅ apply_readiness / optimization_stop_signal / top_fix_items 关键字段均正常返回
3. ✅ 证据链和匹配依据正常
4. ✅ 3 个固定测试简历已固化，可被后续 4-3.8 复用
5. ✅ go test ./... 通过
6. ✅ make llm-check 可用
7. ✅ 未调用真实 LLM，未修改 Prompt / Schema

## 10. 下一步建议

1. **正式进入 4-3.8**：ResumeOptimizationAgent 接入 Pipeline（LLM 优先 + Mock fallback）
2. **4-3.8 可使用本阶段固化的 3 个测试简历作为稳定输入**
3. **4-3.8 应使用真实 LLM 验证三类 Agent 的完整返回质量**
