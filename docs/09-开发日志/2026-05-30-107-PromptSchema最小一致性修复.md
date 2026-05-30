# 开发日志：4-3.7.C.1 Prompt / Schema 最小一致性修复

> 日期：2026-05-30
> 阶段：4-3.7.C.1
> 执行者：CodeBuddy

## 任务目标

为三个业务 Agent（MatchAnalysis / Recommendation / ResumeOptimization）的 JSON Schema 和 Prompt 添加 `apply_readiness`、`optimization_stop_signal`、`top_fix_items` 三个字段，支持项目战略定位"投递前决策与简历初筛优化系统"。

## 修改文件清单

### JSON Schema

| 文件 | 修改内容 |
|------|---------|
| `server/prompts/schema/match_analysis_schema.json` | required 新增 3 字段；properties 新增 apply_readiness / optimization_stop_signal / top_fix_items；$defs 新增 top_fix_item |
| `server/prompts/schema/recommendation_schema.json` | recommendation_item.required 新增 3 字段；recommendation_item.properties 新增 3 字段；$defs 新增 top_fix_item |
| `server/prompts/schema/resume_optimization_schema.json` | required 新增 3 字段；properties 新增 3 字段；$defs 新增 top_fix_item |

### Prompt

| 文件 | 修改内容 |
|------|---------|
| `server/prompts/match_analysis_v1.md` | 新增 Section 7.5-7.7；更新 Section 5.4 / 20 / 21 / 22 |
| `server/prompts/recommendation_v1.md` | 新增 Section 9；更新 Section 5.1 / 18 / 20 |
| `server/prompts/resume_optimization_v1.md` | 新增 Section 6；更新 Section 12 / 13 / 15 |

### 测试文件

| 文件 | 修改内容 |
|------|---------|
| `server/internal/schema/validator_test.go` | minimalMatchAnalysis / minimalRecommendation / minimalResumeOptimization 新增 3 字段 |
| `server/internal/agent/repair/json_repair_agent_test.go` | 同上 |
| `server/internal/agent/pipeline/pipeline_test.go` | 同上 |

## 新字段定义

### apply_readiness

4 值枚举，表示投递准备度：

| 值 | 含义 |
|----|------|
| `ready_to_apply` | 核心能力和证据已足够投递 |
| `quick_fix_before_apply` | 整体方向可行，但有 1-3 个快速修复项 |
| `not_ready_for_this_role` | 简历证据和岗位核心要求差距较大 |
| `insufficient_information` | 信息不足以做出可靠判断 |

### optimization_stop_signal

4 值枚举，表示优化停止信号：

| 值 | 含义 |
|----|------|
| `stop_and_apply` | 简历已准备好，不要过度优化 |
| `fix_top_3_then_apply` | 先完成 top_fix_items 再投递 |
| `needs_more_evidence` | 需要真实经验或证据，不是表达能解决的 |
| `target_role_mismatch` | 目标岗位和简历证据不匹配 |

### top_fix_items

0-3 项数组，每项含：

- `issue`：需要修复的具体问题
- `why_it_matters`：为什么对该岗位初筛有影响
- `action`：用户应采取的具体行动
- `example_reference`：修复后可能的样子（仅供参考）
- `self_check_question`：自检修复是否真实准确
- `estimated_effort`：`within_10_min` / `within_30_min` / `within_60_min`

### 映射约束

- `ready_to_apply` → `stop_and_apply`
- `quick_fix_before_apply` → `fix_top_3_then_apply`
- `not_ready_for_this_role` → `needs_more_evidence` 或 `target_role_mismatch`
- `insufficient_information` → `needs_more_evidence`

## 验证结果

- `python3 -m json.tool` 三个 Schema 文件全部通过
- `go test ./...` 全部通过
- `grep -rl apply_readiness server/prompts/` 覆盖 6 个文件
- `grep -rl optimization_stop_signal server/prompts/` 覆盖 6 个文件
- `grep -rl top_fix_items server/prompts/` 覆盖 6 个文件

## 未修改范围

- 未修改业务代码（service / handler / controller）
- 未修改前端
- 未修改 Makefile
- 未调用真实 LLM
- 未清空 volume

## 下一步

进入 4-3.8 三类业务 Agent 集成验收。
