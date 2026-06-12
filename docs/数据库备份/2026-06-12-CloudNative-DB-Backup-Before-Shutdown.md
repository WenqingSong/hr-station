# CloudNative-DB-Backup-Before-Shutdown — 环境关闭前数据库备份

执行时间：2026-06-12 14:44

## 容器状态

| 服务 | 状态 |
|---|---|
| offer-catcher-nginx | Up 7min |
| offer-catcher-redis | Up 7min (healthy) |
| offer-catcher-mysql | Up 7min (healthy) |
| Docker volume | `workspace_mysql_data` |

## SQL 增量文件

| 文件 |
|---|
| `resources/sql-add/20260612_2100_sep_recommend_datalayer.sql` |
| `resources/sql-add/20260612_2130_sep_weight_fix.sql` |

## 备份文件

| 类型 | 路径 | 大小 |
|---|---|---|
| 完整数据库 | `backups/db/offer_catcher_full_20260612_144443.sql` | 2.6M |
| SEP + 推荐表 | `backups/db/offer_catcher_sep_recommend_20260612_144445.sql` | 1.2M |
| Schema 结构 | `backups/db/offer_catcher_schema_20260612_144448.sql` | 36K |
| 压缩包 | `backups/db/offer_catcher_db_backup_20260612_144457.tar.gz` | 449K |

## 关键表行数

| 表 | 行数 |
|---|---|
| users | 3 |
| resumes | 8 |
| jobs | 46 |
| semantic_profiles | 47（1 resume + 46 job verified） |
| semantic_profile_tags | 708 |
| recommendation_results | 47 |
| resume_profiles | 16 |
| job_profiles | 46 |

## SEP 数据健康

- resume verified: 1
- job verified: 46/46
- semantic_profile_tags: 708 条
- failed/stale/pending: 0

## 恢复方式

```bash
# 恢复完整数据库
docker compose exec -T mysql mysql -uroot -p"${MYSQL_ROOT_PASSWORD}" < backups/db/offer_catcher_full_20260612_144443.sql

# 仅恢复 SEP + 推荐表
docker compose exec -T mysql mysql -uroot -p"${MYSQL_ROOT_PASSWORD}" < backups/db/offer_catcher_sep_recommend_20260612_144445.sql
```

注意：恢复前确认当前 DB 是否需要覆盖；不要直接覆盖线上库。

## 关闭环境建议

```bash
# 推荐（保留 volume）
make down
# 或
docker compose down

# 禁止（会删除 volume 和数据）
docker compose down -v
make reset
```

## 后续 SQL 必须落盘

`/workspace/resources/sql-add/`
