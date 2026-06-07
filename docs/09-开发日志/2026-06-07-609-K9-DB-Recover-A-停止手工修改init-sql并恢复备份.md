# K9-DB-Recover-A — 停止手工修改 init.sql 并恢复备份 ✅

**日期**: 2026-06-07  
**阶段**: 纠偏恢复  
**关联**: K9-1-Regression-A（只读诊断）、K9-1-Regression-B-Fix（已中止）

---

## 1. 停止原因

| 问题 | 说明 |
|---|---|
| K9-1-Regression-A 诊断结论正确 | `init.sql` 仅含 resume_id=1 的 1 条 recommendation_results |
| 手工拼接 SQL/JSON 转义复杂 | Python `json.dumps()` 转义、MySQL `QUOTE()` 转义、mysqldump 转义三套格式不一致 |
| 当前仍处于开发阶段 | 不应反复手工修改 `init.sql` |
| 当前数据库运行状态才是可信状态 | 所有推荐/匹配/优化数据都在运行数据库中 |
| 最合理做法 | 功能开发完成后一次性 `mysqldump` 替换 `init.sql` |

---

## 2. init.sql 手工修改风险

在 K9-1-Regression-B-Fix 中尝试了两种方式写入 seed：
1. Python `json.dumps()` → 转义格式不符 MySQL 预期，导入临时库报错 `Invalid JSON text`
2. MySQL `QUOTE()` + `GROUP_CONCAT` → 格式正确但 `group_concat_max_len` 需调参，整体复杂度过高

**结论**: 停止手工拼接 SQL seed，回归一步到位策略。

---

## 3. 备份与恢复

| 项目 | 值 |
|---|---|
| 使用的备份文件 | `backups/sql/init.sql.bak.20260607-194337` |
| 备份大小 | 1.5M，597 行 |
| 备份内容 | K9-DB-Snapshot-A 后 + K9-1-Regression-B-Fix 前的干净版本 |
| 恢复后 init.sql | 597 行，无 K9-1-Regression-B 污染标记 |

---

## 4. 当前数据库完整性确认

| 表/数据 | 确认结果 |
|---|---|
| resumes | 5 条 ✅（id=1/2/3/4/5） |
| resume_profiles | 10 条 ✅（5×fast ready + 5×quality stale） |
| recommendation_results | 48 条 ✅（含 K9-1-Regression-B-Fix 中生成的 46/47/48） |
| match_analyses | 111 条 ✅ |
| resume_optimization_results | 53 条 ✅ |
| 当前运行数据库未重置 | ✅ 未执行 reset-up |

---

## 5. API / 页面轻量验证

| 项 | 结果 |
|---|---|
| GET `/api/jobs/recommend/latest?resume_id=5` | `has_result: true`, 5 jobs ✅ |
| npm build | ✅ 通过 |
| go build | ✅ 通过 |
| go vet | ✅ 通过 |
| check-encoding | ⚠️ FAIL — 已知 mysqldump JSON 列中文双转义误报，非本次引入 |

---

## 6. 后续数据库策略

| 阶段 | 策略 |
|---|---|
| **开发阶段（当前）** | 继续使用当前运行数据库推进功能，不手工修改 `init.sql` |
| **K9 全部功能完成后** | **K9-DB-Final-Snapshot**：`mysqldump` 当前完整数据库 → 临时库验证 → 整体替换 `init.sql` |
| **最终快照原则** | 不手写 JSON seed，不局部修补，一步到位 mysqldump |

---

## 7. 未修改范围确认

- [x] 未修改前端源码
- [x] 未修改后端业务源码
- [x] 未修改数据库结构
- [x] 未删除当前数据库数据
- [x] 未清空业务表
- [x] 未删除真实简历
- [x] 未修改 recommendation_results 表结构
- [x] 未修改 ResumeProfile 画像生成逻辑
- [x] 未修改推荐排序逻辑
- [x] 未执行 `reset-up`
- [x] 未执行 Git 命令
- [x] 已恢复 `init.sql` 到备份状态
- [x] 已清理临时文件
- [x] 已删除临时库 `offer_catcher_init_check`
