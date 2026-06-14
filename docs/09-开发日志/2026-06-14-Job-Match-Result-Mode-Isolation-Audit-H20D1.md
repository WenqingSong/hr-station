# H20-D-1：Job Match Result Mode Isolation Audit

> 日期：2026-06-14
> 类型：只读审计
> 编号：H20-D-1

## 结论：Fast/Deep 结果不隔离，确认存在覆盖

### 证据链路

1. **DB 层**: `match_analyses` 表有 `generation_tier` 字段，但 `FindLatestByResumeAndJob` 查询**不带 tier 过滤**，`ORDER BY created_at DESC LIMIT 1` 返回跨 tier 的最新行。
2. **真实数据**: resume=2 job=10001 下同时有 28 条 fast + 6 条 quality，最新记录是 quality 层级。
3. **前端层**: 只有单一 `activeResult/activeTier` 状态，无法同时持有 Fast/Deep 两套结果。
4. **API 层**: `GET /api/jobs/{id}/match/latest` 返回唯一 latest 结果，不区分 tier。
5. **保存层**: Fast SEP Engine 写 `generated_by=mock_agent`（BUG），Deep Match V3 写 `generation_tier=quality`。

### 覆盖发生位置

| 层级 | 覆盖原因 |
|---|---|
| DB 保存 | 无 UNIQUE（resume_id+job_id+tier），每次 INSERT 新行 ✅ 不覆盖单行 |
| 后端查询 | `FindLatestByResumeAndJob` 不带 tier 过滤 → 返回最新行（可能是 Deep） |
| API | `/latest` 返回唯一结果 |
| 前端 | 单一 `activeResult` 状态 |

### 唯一索引

```
PRIMARY KEY (id) — auto_increment only
idx_match_user_created (user_id, created_at)
idx_match_job (job_id)
idx_match_resume (resume_id)
```
**无唯一约束防止重复**，每代生成新行。

### 问题分级

| 等级 | 问题 |
|---|---|
| P0 | `FindLatestByResumeAndJob` 不带 tier 过滤，Deep 覆盖 Fast 默认视图 |
| P0 | 前端单 `activeResult`，Deep 后丢失 Fast |
| P1 | Fast SEP Engine `generated_by=mock_agent`（应为 `fast_sep_match_engine`） |
| P1 | API `/latest` 不区分 tier |
| P2 | 无 UNIQUE (resume_id, job_id, generation_tier) |

### H20-D-2 修复建议

1. `match_analyses` 不新增字段（已有 `generation_tier`）
2. `FindLatestByResumeAndJob` 增加 `generation_tier` 参数
3. 前端拆分为 `fastResult / deepResult` + `activeMode`
4. 页面默认加载 `mode=fast` latest
5. Fast SEP Engine 修复 `generated_by` 写入
