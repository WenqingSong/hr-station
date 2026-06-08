# K9-DB-Final-Snapshot-B — mysqldump 导出并临时库验证

日期：2026-06-08
性质：数据库导出 + 临时库验证。

## 导出前检查

bad_cnt: **全部为 0** ✅
核心表行数: users=2, resumes=8, profiles=16, jobs=46, rec=42, match=113, opt=63, compare=0

## 操作流程

1. **备份旧 init.sql**: `backups/sql/init.sql.bak.*-before-final-snapshot` (1.5MB)
2. **mysqldump 导出**: `--single-transaction --hex-blob --no-tablespaces` → 798 行, 1.7MB
3. **创建临时库**: `offer_catcher_snapshot_check`
4. **导入**: 无 SQL 语法错误
5. **验证通过**: 全部一致

## 行数对比

| 表 | 当前库 | 临时库 | 一致 |
|------|------|------|------|
| users | 2 | 2 | ✅ |
| resumes | 8 | 8 | ✅ |
| resume_profiles | 16 | 16 | ✅ |
| jobs | 46 | 46 | ✅ |
| recommendation_results | 42 | 42 | ✅ |
| match_analyses | 113 | 113 | ✅ |
| resume_optimization_results | 63 | 63 | ✅ |

## 乱码检查

| 表 | bad_cnt |
|------|------|
| recommendation_results | 0 |
| match_analyses | 0 |
| resume_optimization_results | 0 |
| resume_compare_results | 0 |

## JSON 验证

K9-3 quality JSON_EXTRACT 正常，score 可读。

## 构建

| 项 | 结果 |
|------|------|
| npm build | ✅ |
| go build/go vet | ✅ |
| check-encoding | FAIL — pre-existing (init.generated.current.sql) |

## 结论

**server/sql/init.sql 已正式可用。** 当前运行为最新 mysqldump 快照，通过临时库全量验证。

## 未修改范围

- 未修改源码
- 未修改数据库结构
- 未删除真实简历
- 未 reset-up
- 未执行 Git
- 已备份旧 init.sql
- 已用 mysqldump 替换 server/sql/init.sql
