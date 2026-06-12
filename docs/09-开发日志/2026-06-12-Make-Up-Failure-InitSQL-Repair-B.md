# 2026-06-12 Make-Up-Failure-InitSQL-Repair-B

## 故障现象

执行 `make up` 时，Step 6（数据库检查/导入）报错：

```
[WARN] 缺失表: users resumes jobs recommendation_results match_analyses resume_compare_results
[INFO] 自动导入 server/sql/init.sql...
ERROR 1064 (42000) at line 1:
You have an error in your SQL syntax; check the manual that corresponds to your MySQL server version
for the right syntax to use near 'mysqldump: [Warning] Using a password on the command line interface can be insec' at line 1
```

## 根因

`server/sql/init.sql` 被 **mysqldump stderr warning** 污染了第 1 行：

```
mysqldump: [Warning] Using a password on the command line interface can be insecure.
```

该行不是合法 SQL 语句，MySQL 无法解析，导致整个导入失败。

**污染来源**：`mysqldump -uroot -p"${PASSWORD}" ... > server/sql/init.sql` — MySQL 客户端将 stderr 输出也写入了重定向文件。`backups/db/` 中所有 `.sql` 备份文件同样被污染。

## 影响范围

- `server/sql/init.sql` — 被替换为全量 mysqldump dump（第 1 行 warning 污染）
- `backups/db/offer_catcher_full_20260612_144443.sql` — 第 1 行 warning 污染
- `backups/db/offer_catcher_schema_20260612_144448.sql` — 第 1 行 warning 污染
- `backups/db/offer_catcher_sep_recommend_20260612_144445.sql` — 第 1 行 warning 污染
- 数据库 `offer_catcher` 为空库（0 张表），因为新 volume 创建后 init.sql 导入失败

## 修复操作

### 1. 污染文件备份

修复前将污染文件备份到 `backups/sql-contaminated-before-repair/20260612_152210/`：
- `init.sql.contaminated`
- `offer_catcher_full_20260612_144443.sql.contaminated`
- `offer_catcher_schema_20260612_144448.sql.contaminated`
- `offer_catcher_sep_recommend_20260612_144445.sql.contaminated`

### 2. SQL 文件清理

使用 Python 精确移除 4 个文件首行的 mysqldump/mysql warning。验证清理后：
- 所有文件首行恢复正常 SQL 内容
- `init.sql` 仍包含 25 个 CREATE TABLE + 完整 INSERT INTO 数据
- `backups/db/*.sql` 恢复正常 mysqldump dump 格式

### 3. 数据库导入

执行 `make db-import` 成功导入：
- 26 张表全部恢复
- users: 3, resumes: 8, jobs: 46
- semantic_profiles: 47, semantic_profile_tags: 708
- recommendation_results: 47, match_analyses: 113

### 4. make up 结果

`make up` 中：
- ✅ Step 6 数据库检查通过（不再触发重新导入）
- ✅ Step 7 腾讯岗位 seed 已应用
- ❌ Step 8 后端启动失败（LLM_API_KEY 为空，非 SQL 问题）
- ✅ Step 10 前端启动成功
- ⚠️ 最终因后端失败导致 healthz 不可达

**后端启动失败原因**：`.env` 中 `LLM_ENABLE=true` 但 `LLM_API_KEY=`（空值），与本次 SQL 修复无关。

### 5. 安全备份脚本

新增 `scripts/dev/db-backup.sh`：
- 使用 `MYSQL_PWD` 环境变量代替 `-p` 参数
- `mysqldump` 不再输出 password warning 到 stderr
- 同时通过 `2>/dev/null` 双重保险
- 输出：full dump + SEP tables dump + schema-only dump + tar.gz 归档

## 经验教训

1. **禁止直接使用 `mysqldump -p > file.sql`**：stderr warning 会混入 SQL 文件
2. **必须使用 `MYSQL_PWD=xxx mysqldump -uroot`** 或 `2>/dev/null` 重定向分离 stderr
3. **禁止 `make down -v` / `make reset` / `make clean-volume`**：会删除 volume 导致数据丢失
4. **before 任何危险操作应先执行 `make db-import` 将数据保存到 SQL 文件**

## 后续待处理

- `.env` 中 `LLM_API_KEY` 为空，需填入真实 API Key 或将 `LLM_ENABLE` 设为 `false` 以恢复后端启动
