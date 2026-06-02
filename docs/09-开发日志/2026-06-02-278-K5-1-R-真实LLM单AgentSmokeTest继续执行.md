# K5-1-R：真实 LLM 单 Agent Smoke Test 继续执行

- **日期**：2026-06-02
- **阶段**：K5-1-R（partial_pass）

## 1. 执行条件

- `LLM_ENABLE=true`，真实 Key 已配置在本地 `.env`
- 未修改代码，未输出真实 Key

## 2. 测试请求

```
POST /api/jobs/55/match
{"resume_version_no": 3}
jobId=55, resume v3
```

## 3. 结果

```
HTTP 200
code: 0
generated_by: mock_agent
agent_status: fallback
fallback_reason: llm_call_failed
schema_valid: True
```

**真实 LLM 调用失败**（llm_call_failed），fallback 到 Mock Agent 正常工作。

## 4. 核心字段

match_score=23, evidence_chain=11, top_fix_items=2, schema_valid=True

## 5. 安全

- Key scan: clean
- `.env` not tracked
- 零代码修改

## 6. P0/P1/P2

| 级别 | 数量 |
|------|------|
| P0 | 0 |
| P1 | 0 |
| P2 | 3（LLM 调用失败 + 只测 1 Agent + 浏览器未验） |

## 7. 判定

**K5-1-R = partial_pass。** 真实 LLM 失败但 fallback 可靠，前端不受影响。

## 8. 本次更新文件

| 文件 | 操作 |
|------|------|
| `docs/16-Jobs样本库重构/真实LLM单AgentSmokeTest.md` | 重写为 K5-1-R 完整报告 |
| `docs/09-开发日志/2026-06-02-278-K5-1-R-真实LLM单AgentSmokeTest继续执行.md` | 新增（本文件） |
| `docs/说明书.md` | 更新 |
| `docs/项目状态.md` | 更新 |
| `docs/下一步任务.md` | 更新 |
| `docs/变更记录.md` | 更新 |
