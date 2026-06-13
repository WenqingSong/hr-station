# Deep-Match-Output-Contract-Freeze-H22C

> 2026-06-13 | 设计冻结 | ✅

## 1. 总体结论

| 项目 | 结果 |
|------|------|
| 完成契约冻结 | ✅ |
| 修改业务代码 | ❌ |
| 冻结 API/DeepMatchReport/DB/legacy/H22-D边界 | ✅ |
| 进入 H22-D | ✅ |

## 2. H22-A/H22-B 问题 → 契约解决

| 问题 | 根因 | 契约解决 |
|------|------|---------|
| Deep 像模板 | 3-agent 只产 gap/risk 列表 | DeepMatchReport 12 模块为主，legacy 为辅 |
| evidence_chain=[] | Prompt 不要求结构化 evidence | `resume_evidence_mapping` 至少 3 条，不足时 agent_status≠match_verified |
| job_understanding=null | deepMatchSystemPrompt 死代码 | 激活 Prompt，JD 解构为必选模块 |
| DeepMatchReport 死代码 | Prompt 未接入执行路径 | H22-D 重构 runDeepMatchSEP 调用链 |
| generated_by 假 Deep | sanitizeDeepResponse 仅换标签 | fallback=legacy 的 agent_status 必须 legacy_fallback |

## 3. API 输出契约

```json
{
  "code": 0,
  "data": {
    "generated_by": "deep_match_agent",
    "agent_status": "match_verified",
    "result": {
      "report_version": "deep_match_v2",
      "score": 0,
      "match_level": "medium",
      "application_decision": "revise_before_apply",
      "one_sentence_conclusion": "",
      "score_explanation": "",
      "job_understanding": {...},
      "requirement_coverage": [...],
      "resume_evidence_mapping": [...],
      "core_match_points": [...],
      "key_risks": [...],
      "resume_rewrite_strategy": [...],
      "interview_risk_questions": [...],
      "action_plan": [...],
      "agent_trace": [...],
      "legacy": { "fit_points":[], "gap_points":[], ... }
    }
  }
}
```

**禁止**：`evidence_mapping=[]` 同时 `agent_status=match_verified`；任何数组 null。

## 4. DeepMatchReport 核心模块

| 模块 | 关键字段 | 最小条数 | 核心约束 |
|------|---------|---------|---------|
| **job_understanding** | must_have/should_have/nice_to_have/hidden | - | 不得复制 JD；推荐算法岗位必须优先分析算法能力 |
| **requirement_coverage** | coverage_status/matched_evidence_ids/gap/why_it_matters | must_have 全覆盖 | missing 必须说明影响和修复 |
| **resume_evidence_mapping** | source_text/interpretation/limitations | **≥3** | source_text 必须来自简历原文，不得编造 |
| **core_match_points** | why_it_matches/evidence_ids | ≥2 | 不得只写标题，必须关联证据 |
| **key_risks** | why_it_matters/possible_interviewer_concern/fix_suggestion | ≥2 | 不得只写"学历风险"标题 |
| **resume_rewrite_strategy** | section/problem/rewrite_direction/example_bullet | ≥3 | example_bullet 不得编造不存在经历 |
| **interview_risk_questions** | question/why_asked/answer_strategy | ≥3 | 必须与岗位/简历差距相关 |
| **action_plan** | step/priority/time_cost/expected_impact | ≥3 | 不得泛泛"建议补充" |
| **agent_trace** | agent_name/task/status/summary | ≥6 | 包含 JD/Evidence/Alignment/Risk/Strategy/Calibration |

## 5. score / level / decision 语义

| score | match_level | decision 倾向 |
|-------|-------------|--------------|
| 90-100 | high | priority_apply / apply_directly |
| 80-89 | good | apply_directly / revise_before_apply |
| 70-79 | medium | revise_before_apply |
| 60-69 | low | backup_apply |
| 0-59 | not_recommended | not_recommended |

**一致性规则**：score/level/conclusion/decision 必须一致。禁止"53=中等匹配"又说"整体较低"。
**gate 规则**：evidence_mapping 为空→最高 apply_directly；must_have 缺→最高 revise_before_apply。

## 6. legacy 兼容策略

- `data.result.legacy` 由 DeepMatchReport 反向生成（fit_points←core_match_points 等）
- legacy fallback（旧 3-agent）标记 `agent_status=legacy_fallback` 或 `partial_report`
- 禁止 legacy+match_verified 组合

## 7. DB result_json 契约

- 保存完整 DeepMatchReport JSON，与 API result 同构
- `generated_by=deep_match_agent`，`generation_tier=quality`
- 刷新页面可从 result_json 恢复

## 8. H22-D 重构边界

| 文件 | 允许 | 目标 |
|------|------|------|
| match_deep_prompt.go | ✅ | 激活 deepMatchSystemPrompt 进入执行路径 |
| match_deep_agent.go | ✅ | runDeepMatchSEP 调用 buildDeepMatchPrompt |
| match_deep_repair.go | ✅ | repair→DeepMatchReport，非 legacy |
| match_deep_schema.go | ✅ | 补齐缺失子类型 |
| match_service.go | ✅ | Deep 路径路由调整 |
| model/match.go | 必要时 | 补齐 response 字段 |
| web/src | ❌ | H22-E 再做前端接入 |

禁止：只改 generated_by、legacy 包装新字段、continue 死代码。

## 9. H22-D 验收用例

| Case | 场景 | 核心验证 |
|------|------|---------|
| C1 | 推荐算法岗位 | job_understanding 含 PyTorch/召回/排序；evidence≥3；risks 有 fix_suggestion |
| C2 | 后端开发岗位 | evidence 映射 Go/MySQL/Redis；risks 含 why_it_matters |
| C3 | 方向不匹配 | score<60；decision=not_recommended/backup_apply |

通用：generated_by=deep_match_agent；evidence_mapping≥3；job_understanding 非空；result_json 落库。

## 10. 构建

vite ✅ 2.34s / go ✅

## 11. docs

- 新增：`docs/09-开发日志/2026-06-13-Deep-Match-Output-Contract-Freeze-H22C.md`
- 更新：`docs/变更记录.md`、`docs/下一步任务.md`

## 12. 结论

1. **H22-C 冻结了 Deep Match 完整输出契约**：API/DeepMatchReport 12 模块/DB/legacy/H22-D 边界。
2. **H22-D 必须激活 deepMatchSystemPrompt**：当前完整 Schema 存在于代码中但从未被调用。
3. **不能继续在 legacy 外包装 match_verified**：fallback 必须标记 legacy_fallback/partial_report。
4. **evidence_mapping 是核心质量指标**：空时不能标记 match_verified。
5. **score/level/conclusion 必须一致**：禁止"53=中等匹配+整体较低"的产品冲突。
