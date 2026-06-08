# K9-DB-Final-Snapshot-A-Fix — 清理乱码并补齐版本对比数据

日期：2026-06-08
性质：数据库清理 + 数据补齐。

## 乱码污染行定位

| bad_ids | resume_id | mode | job_id | tier | count |
|---------|------|------|--------|------|-------|
| 2,3,11,19,40 | 1 | general | 0 | fast | 5 |
| 58,59 | 1 | target_job | 10001 | fast | 2 |
| 5,13,21,43 | 2 | target_direction | 0 | fast | 4 |
| 7,15,23 | 2 | target_job | 10001 | fast | 3 |

**共 14 行**，均为 fast tier。根因：旧版 `focus[:80]` byte-slice 截断中文。

## 修复操作

1. **备份**：mysqldump 到 `backups/sql/k9-bad-rows-*.txt`（25KB）
2. **删除**：精确 DELETE 14 行 → bad_cnt = 0
3. **补齐 compare**：API POST quality 生成 5 个对比场景

## 验证

| 检查项 | 结果 |
|------|------|
| optimization bad_cnt | 0 ✅ |
| compare bad_cnt | 0 ✅ |
| optimization total | 63 行 |
| K9-4 compare API | ✅ 5 场景 quality OK |
| npm build | ✅ |
| go build / go vet | ✅ |
| check-encoding | ✅ OK |

## 未修改范围

- 未修改源码
- 未修改 init.sql
- 未修改数据库结构
- 未删除真实简历
- 未 reset-up / 未 Git
- 仅 DELETE 已备份污染行 + API 新增
