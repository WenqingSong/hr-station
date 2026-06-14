# DEPLOY-SCHEMA-MIGRATION-CHECK：生产部署前 Schema 迁移检查

## 一、背景

本项目当前 **没有正式的自动化 database migration 系统**（如 golang-migrate / Flyway）。

数据库 schema 通过以下方式管理：

1. **`server/sql/init.sql`** — 全量初始化脚本（包含 25 条 `DROP TABLE IF EXISTS`），**仅适用于全新空库，严禁在生产已有数据的数据库上执行。**
2. **增量 Migration SQL** — 后续字段变更通过手动执行 `.sql` 文件完成。
3. **Go Migration Runner** — 部分 migration 提供可执行 Go 脚本。

生产部署前必须确认增量 migration 是否已执行，否则相关接口可能返回 500 错误。

---

## 二、生产环境是否需要 Migration

### 判断标准

| 场景 | 是否需要 Migration |
|------|-------------------|
| 全新空库部署（bootstrap override） | 不单独需要 — `init.sql` 已包含完整 schema |
| 已运行过 `init.sql` 的旧库 | **需要确认以下增量 migration 是否已执行** |
| 使用过一次 bootstrap 后再未更新 | 大概率需要执行以下增量 migration |

---

## 三、必须检查的增量 Migration

### 3.1 `server/sql/migration-j1.sql` — jobs 表字段扩展

**影响范围：** 岗位列表、岗位详情、岗位筛选、推荐

**变更内容：**
- `jobs` 表新增 14 列（`application_type`, `application_project`, `application_subtype`, `slogan`, `summary`, `responsibilities_json`, `must_have_json`, `nice_to_have_json`, `bonus_points_json`, `notes_json`, `skills_json`, `organizations_json`, `locations_json`, `interview_locations_json`, `facets_json`, `source_name`, `collected_at`）
- 新增 3 张关联表：`job_locations`, `job_business_groups`, `job_skills`
- 新增索引

**未执行后果：**
- 岗位接口返回数据缺失关键字段（slogan, summary, skills_json 等）
- 岗位详情页展示不完整
- 筛选维度缺失（application_type, facets_json）

**检查方法：**
```sql
SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS
WHERE TABLE_NAME = 'jobs' AND COLUMN_NAME = 'application_type';
-- 有结果 → 已执行；空 → 未执行
```

---

### 3.2 `server/sql/migration-j2.sql` — resume_versions 字段扩展

**影响范围：** 简历管理、简历优化、推荐

**变更内容：**
- `resume_versions` 表新增 11 列（`schema_version`, `target_direction`, `job_type`, `expected_cities`, `school`, `degree`, `major`, `graduation_year`, `profile_json`, `job_intention_json`, `internship_json`, `certificates_json`, `awards_json`, `portfolio_json`, `extra_json`）

**未执行后果：**
- 新版简历字段无法存储，创建/更新简历可能报错
- 简历优化、推荐、匹配分析功能可能因字段缺失返回异常

**检查方法：**
```sql
SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS
WHERE TABLE_NAME = 'resume_versions' AND COLUMN_NAME = 'schema_version';
-- 有结果 → 已执行；空 → 未执行
```

---

### 3.3 `server/sql/migration-k13-tag-evidence.sql` — job_profiles 表

**影响范围：** 岗位推荐（Fast / Tag Evidence）

**变更内容：**
- 新增 `job_profiles` 表（`CREATE TABLE IF NOT EXISTS`，**安全**）
- 存储岗位持久化标签与证据数据

**未执行后果：**
- Tag Evidence 服务无法使用岗位标签进行匹配
- Fast 推荐可能降级到更简单的匹配逻辑

**检查方法：**
```sql
SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES
WHERE TABLE_NAME = 'job_profiles';
-- 有结果 → 已执行；空 → 未执行
```

---

### 3.4 `server/sql/migrations/20260611_k13_result_tables_add_agent_fields.sql` — Agent 字段

**影响范围：** 推荐结果持久化、匹配分析

**变更内容：**
- `recommendation_results` 表新增 7 列（`agent_status`, `fallback`, `repair_applied`, `error_code`, `duration_ms`, `model_name`, `result_version`）
- `match_analyses` 表新增对应列

**未执行后果：**
- **写入推荐结果时 SQL 错误 → 接口返回 500**
- 这是**最可能导致生产 500 的未执行 migration**

**检查方法：**
```sql
SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS
WHERE TABLE_NAME = 'recommendation_results' AND COLUMN_NAME = 'agent_status';
-- 有结果 → 已执行；空 → 未执行
```

---

### 3.5 `resources/sql-add/` 目录 — SEP 推荐数据层

**影响范围：** SEP 语义画像推荐

**文件列表：**
- `resources/sql-add/20260612_2100_sep_recommend_datalayer.sql`
- `resources/sql-add/20260612_2130_sep_weight_fix.sql`

**未执行后果：**
- SEP-based Fast 推荐可能无法正常工作
- 语义标签相关表可能缺失

---

## 四、禁止执行的操作

| 操作 | 原因 |
|------|------|
| 在生产直接执行 `server/sql/init.sql` | 包含 25 条 `DROP TABLE IF EXISTS`，会**删除所有生产数据** |
| 使用 `docker-compose.prod.bootstrap.yml` 重新部署 | 会挂载 `init.sql` 并执行 `DROP TABLE` |
| 在生产执行 `docker compose down -v` | 删除所有 named volume（数据库数据永久丢失） |

---

## 五、推荐的生产 Schema 检查流程

1. **部署前只读检查：** 先通过 `INFORMATION_SCHEMA` 查询确认各表字段是否存在
2. **按优先级执行：** 优先执行 `k13_result_tables_add_agent_fields`（影响推荐稳定性）
3. **逐条验证：** 每执行一条 migration，验证结果
4. **备份先行：** 执行任何 SQL 前先做数据库备份
5. **避免 init.sql：** 永远不要在生产环境执行完整的 `init.sql`

---

## 六、后续生产部署 Prompt 必须包含的 Schema 检查步骤

```text
1. 通过 INFORMATION_SCHEMA 检查 recommendation_results 表是否有 agent_status 列
2. 如缺失 → 执行 migrations/20260611_k13_result_tables_add_agent_fields.sql
3. 检查 jobs 表是否有 application_type 列
4. 如缺失 → 执行 migration-j1.sql → migration-j2.sql
5. 检查 job_profiles 表是否存在
6. 如缺失 → 执行 migration-k13-tag-evidence.sql
7. 确认 migration-j1, migration-j2 的 ALTER 语句执行结果
8. 全部通过 → 可以启动应用
```

---

## 七、相关文件清单

| 文件路径 | 类型 | 风险等级 |
|---------|------|---------|
| `server/sql/init.sql` | 全量初始化 | ⛔ 禁止生产执行 |
| `server/sql/init.sql.bak` | 备份 | ℹ️ 参考 |
| `server/sql/migration-j1.sql` | 增量 ALTER | ⚠️ 可能缺失 |
| `server/sql/migration-j2.sql` | 增量 ALTER | ⚠️ 可能缺失 |
| `server/sql/migration-k13-tag-evidence.sql` | 增量 CREATE | ✅ 安全（IF NOT EXISTS）|
| `server/sql/migrations/20260611_k13_result_tables_add_agent_fields.sql` | 增量 ALTER | 🔴 关键，缺失导致 500 |
| `server/sql/migrations/20260611_k13_run.go` | Go Migration Runner | ℹ️ 可选运行 |
| `resources/sql-add/20260612_2100_sep_recommend_datalayer.sql` | SEP 数据层 | ⚠️ 可能缺失 |
| `resources/sql-add/20260612_2130_sep_weight_fix.sql` | SEP 权重修复 | ⚠️ 可能缺失 |
| `server/sql/seed/tencent_jobs.sql` | 岗位种子数据 | ℹ️ 生产通常已有数据 |
| `server/sql/resume_versions.sql` | 简历种子数据 | ℹ️ 生产通常已有数据 |
