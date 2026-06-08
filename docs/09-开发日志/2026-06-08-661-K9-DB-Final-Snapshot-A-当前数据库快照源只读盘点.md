# K9-DB-Final-Snapshot-A — 当前数据库快照源只读盘点

日期：2026-06-08
性质：只读盘点，零修改。

## 一、数据库连接信息

| 项 | 当前值 |
|------|------|
| MySQL 容器名 | `offer-catcher-mysql` |
| Docker service 名 | `mysql` |
| 镜像 | `mysql:8.0` |
| 端口 | `3306` |
| 数据库名 | `offer_catcher` |
| 用户名 | `root` |
| 密码 | `***` |
| 字符集 | `utf8mb4` / `utf8mb4_unicode_ci` |
| init.sql 位置 | `server/sql/init.sql` |
| make up 使用 init.sql | ❌ 当前 docker-compose 没有挂载 init.sql volume |

## 二、init.sql 当前状态

| 项 | 结果 |
|------|------|
| 存在 | ✅ |
| 行数 | 597 |
| 大小 | 1.5MB |
| 风格 | ✅ mysqldump（`DROP TABLE IF EXISTS` + `LOCK TABLES WRITE` + `/*!40101 SET`） |
| 手工污染 | ❌ 无，纯 mysqldump 输出 |
| 含四大核心表 | ✅ users/resumes/jobs/recommendation_results/match_analyses/resume_optimization_results/resume_compare_results |
| 建议备份 | ✅ 是，替换前必须先备份 |

## 三、核心表行数

| 表 | 行数 | 判断 |
|------|------|------|
| users | 2 | ✅ |
| resumes | 8 | ✅ 含 3 个 V2/V3 衍生版 |
| resume_profiles | 16 | ✅ 8 fast + 5 quality stale + 3 quality ready |
| jobs | 46 | ✅ 腾讯岗位 |
| recommendation_results | 42 | ✅ fast 40 + quality 2 |
| match_analyses | 113 | ✅ fast 107 + quality 6 |
| resume_optimization_results | 62 | ⚠️ fast 26 + quality 36 |
| resume_compare_results | 0 | ⚠️ 空（需 POST quality 生成） |

## 四、5 类简历与画像完整性

| resume_id | profile_type | dominant_profile | status | 结论 |
|------|------|------|------|------|
| 1 | fast | algorithm_research | ready | ✅ |
| 1 | quality | algorithm_research | stale | ⚠️ 需刷新 |
| 2 | fast | backend_engineering | ready | ✅ |
| 2 | quality | backend_engineering | ready | ✅ |
| 3 | fast | product_operation | ready | ✅ |
| 3 | quality | product_operation | ready | ✅ |
| 4 | fast | career_exploration | ready | ✅ |
| 4 | quality | career_exploration | stale | ⚠️ 需刷新 |
| 5 | fast | business_solution | ready | ✅ |
| 5 | quality | business_solution | stale | ⚠️ 需刷新 |

## 五、四大核心功能数据完整性

| 功能 | fast | quality | latest ready | 结论 |
|------|------|---------|-------------|------|
| K9-1 岗位推荐 | ✅ 40 | ✅ 2 | ✅ | |
| K9-2 岗位匹配分析 | ✅ 107 | ✅ 6 | ✅ | |
| K9-3 简历优化建议 | ✅ 26 | ✅ 36 | ⚠️ 14 rows 有乱码 | 需清理 |
| K9-4 简历版本对比 | 0 | 0 | ❌ 空 | 需 POST quality 生成 |

## 六、脏数据 / 临时数据 / 乱码风险

| 检查项 | 结果 | 是否阻断 |
|------|------|------|
| tmp/temp/test 表 | ✅ 无 | — |
| 非 ready 状态 | quality profiles 5 stale | ⚠️ 建议刷新 |
| 乱码 `�` | **14 rows** in resume_optimization_results | ⚠️ 阻断 |
| 空 result_json | ✅ 无 | — |

**乱码来源**：K9-3-Regression-H-F 修复前的 fast 模式 `focus[:80]` byte-slice 截断产生的旧数据。后端源码已修复，但 DB 中旧数据仍需清理。

## 七、当前数据库是否适合作为最终快照源

**结论：暂不适合（需先清理乱码数据）**

原因：
1. `resume_optimization_results` 有 14 行含 `�` 字符
2. `resume_compare_results` 为空，需先生成几条对比数据
3. 3 份 quality profile 为 stale，建议刷新

## 八、下一阶段推荐 mysqldump 命令

```bash
docker compose exec -T mysql mysqldump \
  -u root -p"${MYSQL_ROOT_PASSWORD}" \
  --default-character-set=utf8mb4 \
  --single-transaction \
  --routines \
  --triggers \
  --hex-blob \
  --set-gtid-purged=OFF \
  --no-tablespaces \
  offer_catcher \
  > server/sql/init.sql
```

## 九、下一阶段验证方案

1. `cp server/sql/init.sql backups/init.sql.$(date +%Y%m%d-%H%M%S)`
2. 执行 mysqldump 导出
3. `docker compose exec -T mysql mysql -u root -ppassword -e "CREATE DATABASE offer_catcher_snap_check"`
4. `docker compose exec -T mysql mysql -u root -ppassword offer_catcher_snap_check < server/sql/init.sql`
5. 校验核心表行数
6. 校验乱码是否清零
7. 校验 API 可读取
8. npm build + go build

## 十、风险提示

| 风险 | 处理建议 |
|------|---------|
| 14 行乱码数据 | K9-DB-Final-Snapshot-B 前先重新生成对应 optimization results |
| resume_compare_results=0 | POST quality general 1/2 等生成几条对比数据 |
| quality profiles 5 stale | 通过 API 触发 quality profile 刷新 |
| check-encoding 会命中旧 SQL dump | pre-existing，导出新 dump 后自然消除 |

## 十一、是否进入 K9-DB-Final-Snapshot-B

**⚠️ 暂不通过**。需先清理乱码数据并补全 contrast 数据，再进入 B 阶段。

建议前置阶段：
- **K9-DB-Final-Snapshot-A-Fix**：清理乱码 + 刷新 profile + 生成 compare 数据

## 十二、未修改范围确认

- 未修改任何源码
- 未修改当前数据库数据
- 未修改 server/sql/init.sql
- 未 reset-up
- 未执行 Git 命令
