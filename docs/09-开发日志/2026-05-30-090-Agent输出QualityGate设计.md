# 2026-05-30-090-Agent 输出 Quality Gate 设计

## 1. 本次任务

阶段 4-3.5.E：Agent 输出 Quality Gate 设计。docs only，不修改任何代码。

## 2. 背景

4-3.5.B 已完成战略差异化确认，项目定位升级为"投递前决策与简历初筛优化系统"。

4-3.5.C 已完成投递边界与优化停止条件设计（杀手锏一）。

4-3.5.D 已完成 JD 证据链匹配与 START 简历优化规则（杀手锏二）。

**Agent 输出 Quality Gate** 是杀手锏三，确保 Agent 输出不仅结构合法，还有可靠内容。

## 3. 核心设计

### 3.1 核心问题

结构合法不等于内容可靠。JSON Schema 只能保证输出结构合法，不能保证内容有业务价值。

### 3.2 Quality Gate 四层结果

| quality_gate_result | 定义 | 后续动作 |
|---------------------|------|---------|
| `pass` | 输出符合业务边界 | 直接返回 |
| `revise_required` | 输出总体可修复 | revise once |
| `fallback_required` | 输出存在严重业务风险 | conservative fallback 或 Mock fallback |
| `manual_review_suggested` | 不确定性高 | 返回保守建议 + 核实提示 |

### 3.3 十项检查维度

1. jd_grounding_check：是否引用 JD
2. resume_evidence_check：是否基于用户简历
3. fabrication_check：是否编造经历
4. apply_boundary_check：是否给出投递边界
5. optimization_stop_check：是否给出优化停止条件
6. example_reference_check：是否使用"示例参考"
7. verification_check：是否包含核实提醒
8. safety_tone_check：是否存在不当表达
9. consistency_check：apply_readiness 与 optimization_stop_signal 是否一致
10. schema_display_check：是否适合前端展示

### 3.4 revise once 设计

- 最多允许一次 revise
- revise 输入包括：原始输出 + failed_checks + JD + 简历上下文 + Schema
- revise 后再检查，仍不合格则进入 fallback

### 3.5 fallback 策略

- conservative_fallback：LLM 输出有一定价值但缺少足够证据
- mock_fallback：LLM 输出存在严重问题
- fallback 不是失败，而是 Demo 稳定性保障

### 3.6 Quality Gate 与现有层的关系

| 层级 | 负责方 | 检查内容 |
|------|--------|---------|
| 结构合法 | JSON Schema | 字段存在、类型正确、枚举合法 |
| 格式修复 | JSON Repair | 修复 Markdown 包裹、语法错误 |
| 内容质量 | Quality Gate | 是否基于 JD、是否基于证据、是否编造、是否泛化 |

## 4. 三类 Agent 质检重点

### 4.1 MatchAnalysisAgent

重点检查：JD 证据链、apply_readiness、optimization_stop_signal、编造简历证据、泛泛匹配

### 4.2 RecommendationAgent

重点检查：推荐岗位来自候选集、编造 job_id/title/department、推荐优先级和理由

### 4.3 ResumeOptimizationAgent

重点检查：使用"示例参考"、包含 verification_notice、基于 JD 和简历证据、不编造经历和指标

## 5. Quality Gate 输出结构

```json
{
  "quality_gate_result": "pass / revise_required / fallback_required / manual_review_suggested",
  "quality_score": 0.0,
  "failed_checks": [
    {
      "check_name": "jd_grounding_check",
      "severity": "high",
      "issue": "缺少 JD 证据链",
      "revision_instruction": "请补充每条建议对应的 JD 要求和简历证据。"
    }
  ],
  "revision_allowed": true,
  "revision_count": 0,
  "final_action": "return / revise_once / conservative_fallback / mock_fallback"
}
```

## 6. 实现路径建议

先做轻量规则版本（纯代码检查，不调 LLM）：
1. 关键词匹配：检查是否包含 JD 关键词
2. 字段存在性检查：检查 apply_readiness 等是否存在
3. 一致性检查：检查 apply_readiness 与 optimization_stop_signal 是否矛盾
4. 禁止词检查：检查"建议改为"等不当表述
5. 编造检测：检查示例参考是否包含用户简历中没有的具体数字

## 7. 对 Prompt / Schema / 前端的影响

- 后续可评估新增字段：quality_gate_result、quality_score、failed_checks、revision_allowed、revision_count、final_action、quality_gate_summary
- 本阶段不修改 Prompt / Schema / 代码 / 前端
- 是否真正修改，放到 4-3.5.H 统一评审

## 8. 当前不做事项

- 不修改 Prompt 文件
- 不修改 Schema 文件
- 不修改 Go 代码
- 不修改前端代码
- 不实现 Quality Gate
- 不接入 Pipeline
- 不调用真实 Hunyuan

## 9. 新增文档

- `docs/05-Agent设计/Agent输出QualityGate设计.md`

## 10. 更新文档

- 更新战略差异化与杀手锏确认：4-3.5.E 已完成
- 更新投递边界与优化停止条件设计：4-3.5.E 关联
- 更新 JD 证据链匹配与 START 简历优化规则：4-3.5.E 关联
- 更新阶段 4-3 业务 Agent 接入 LLM 方案：4-3.5.E 已完成
- 更新岗位匹配 Agent：后续需通过 Quality Gate
- 更新岗位推荐 Agent：后续需通过 Quality Gate
- 更新 Agent 总览：4-3.5.E 已完成
- 更新 Prompt 模板：后续 Prompt 需配合 Quality Gate 检查项
- 更新 JSON 输出规范：后续 Schema 可评估增加 quality_gate_result 等字段
- 更新失败兜底与 Mock 回退策略：Quality Gate 是内容层兜底
- 更新匹配分析接口文档：后续接口可考虑展示 Quality Gate 结果
- 更新推荐接口文档：后续接口可考虑展示 Quality Gate 结果
- 更新项目状态：4-3.5.E 已完成
- 更新下一步任务：4-3.5.E 已完成，4-3.5.F 待执行
- 更新变更记录：4-3.5.E 变更记录
- 更新说明书：4-3.5.E 已完成

## 11. 自检

- [x] 未修改任何 Go 代码
- [x] 未修改任何 Prompt 文件
- [x] 未修改任何 Schema 文件
- [x] 未修改前端代码
- [x] 未调用真实 Hunyuan
- [x] 未进入 4-3.6
- [x] Quality Gate 四种结果明确定义
- [x] 十项检查维度已定义
- [x] revise once 设计已明确
- [x] fallback 策略已设计
- [x] 与 JSON Repair / Schema Validate 区别已明确
- [x] 三类 Agent 质检重点已定义
- [x] 输出结构已定义
- [x] 对 Prompt / Schema / 前端影响已登记
- [x] 当前不做事项已列出
- [x] 下一步建议已明确

## 12. 下一步

**下一步进入 4-3.5.F：学生求职极端情况处理边界。**

4-3.5.F 的核心内容：

1. 设计信息不足时的处理策略
2. 设计简历很弱时的最小改进建议
3. 设计岗位不匹配时的解释策略
4. 设计海外/地点/学历/实习资格等条件不明确时的处理
5. 仍然以 docs only 为主

## 13. 结论

阶段 4-3.5.E 已完成。项目已完成 Agent 输出 Quality Gate 设计，明确后续 Agent 输出不仅要通过 JSON Schema，还需要通过内容质量门禁，检查是否有 JD 证据链、简历证据、投递边界、优化停止条件、示例参考、用户核实提醒，以及是否存在编造经历、泛化建议或不当承诺；本阶段未修改代码、Prompt、Schema 或前端，下一步进入 4-3.5.F 学生求职极端情况处理边界。
