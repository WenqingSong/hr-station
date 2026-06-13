# Deep-Match-Content-Quality-ReadOnly-Audit-H22A

> 2026-06-13 | 只读审计 | 根因已定位

## 1. 总体结论

| 项目 | 结果 |
|------|------|
| 是否完成审计 | ✅ |
| 修改代码 | ❌ |
| Deep 内容质量 | **差** — 3-agent legacy pipeline, 非 12 模块报告 |
| 最大问题 | DeepMatchReport 新字段从未进入 API 响应 |
| 后端/前端 | 后端(新字段未生成) + 前端未接入 |
| 进入 H22-B | ✅ |

## 2. API 字段审计 (3 Deep jobs, resumeId=3)

| 字段 | 状态 |
|------|------|
| generated_by | deep_match_agent ✅ |
| score/level | 35-55, 弱/中等匹配 ✅ |
| gap_points | 11-14 items ✅ (legacy) |
| risk_points | 2-5 humanized strings ✅ |
| evidence_chain | [] 空 🔴 |
| **job_understanding** | **null** 🔴 |
| **resume_rewrite_strategy** | **null** 🔴 |
| **interview_risk_questions** | **[]** 🔴 |
| **agent_trace** | **[]** 🔴 |
| **requirement_coverage** | **null** 🔴 |

## 3. 根因链路

```
3-agent legacy pipeline (match_deep_agents.go)
  -> gap_points/risk_points (字符串列表)
  -> fit_points=null, evidence_chain=null

sanitizeDeepResponse (match_deep_agent.go)
  -> generated_by=deep_match_agent, null->[] 修复 ✅
  -> 但 DeepMatchReport 新字段(job_understanding 等) 从未填充 🔴

API 响应
  -> data.result 仅有 legacy 字段
  -> 新 12 模块字段全部 null/空

前端
  -> evidence_chain=[] -> 空态文案
  -> job_understanding 等 null -> 不渲染
```

## 4. 内容质量评分 (1-5)

| 维度 | 分 | 证据 |
|------|-----|------|
| JD 理解 | 2 | 推荐算法岗位的 gaps 含 Docker/K8s/Linux, 而非召回/排序/CTR |
| 简历证据可信度 | 1 | evidence_chain=[] |
| 岗位-简历映射 | 2 | 仅有 gap 列表, 无 requirement_coverage |
| 分数-结论一致性 | 2 | "中等匹配" vs "整体匹配度较低" 冲突 |
| 风险诊断深度 | 2 | 5 条 risk 仅标题("Linux技能缺失风险") |
| 修改建议 | 1 | 无 resume_rewrite_strategy |
| 面试追问 | 1 | 无 interview_risk_questions |
| Fast/Deep 差异 | 3 | Deep 有更多 gaps(14 vs 5) |
| 真人顾问感 | 2 | 模板化 gap 列表 |

## 5. 问题清单

| 等级 | 问题 | 根因 |
|------|------|------|
| P0 | new 12-module fields all NULL | DeepMatchReport not in API path |
| P0 | evidence_chain=[] | 3-agent doesn't produce structured evidence |
| P1 | job_understanding=null | Not generated |
| P1 | recommends Golang/Docker for ML job | Prompt lacks SEP semantic context |

## 6. 数据库

最近 record: mock_quality (旧), no deep_match_agent records. result_json 为 legacy 结构.

## 7. 下一步

| 阶段 | 优先级 |
|------|-------|
| H22-B: Deep 后端链路审计 (3-agent prompt I/O) | P0 |
| H22-C: Deep 输出契约冻结 (12 module -> API) | P0 |
| H22-D: Prompt/Schema/Agent 重构 | P0 |
| H22-E: 前端新字段接入 | P1 |

## 8. 结论

1. Deep Match 本质是 3-agent legacy pipeline, 输出 gap/risk strings, 非结构化报告.
2. DeepMatchReport 8 模块 Schema 已定义但从未进入 API 响应.
3. evidence_chain=[] 导致前端"简历证据映射"空态.
4. Prompt 无 SEP 语义上下文导致偏题分析.
5. 前端已尽力用 legacy 字段渲染 9 模块, 新字段缺失时展示专业空态.
6. H22-B 须审计 3-agent pipeline 实际 LLM I/O 确认断层位置.
