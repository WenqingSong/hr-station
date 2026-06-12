# 数据库状态快照 — DB-State-Snapshot-A

> 执行时间：2026-06-12 10:30 UTC+8
> 执行环境：开发环境（Docker MySQL 容器 `offer-catcher-mysql`）
> 数据库名：`offer_catcher`
> 连接方式：`docker exec mysql -uroot`（密码从 .env 读取，不输出）

本快照是 SEP / GQCL / 岗位推荐 Agent 全链路重构前的数据库基线。
后续上线前必须重新阅读本文件，再设计迁移计划。
**不允许直接覆盖线上库。不允许仅依赖 init.sql。**

---

## ⚠️ SQL 迁移规则（必须遵守）

1. 后续新增表、补字段、索引、迁移 SQL，**都必须放入 `/workspace/resources/sql-add`**。
2. **不允许只改 `server/sql/init.sql`**。
3. **不允许只在代码里自动建表**。
4. SQL 文件命名：`YYYYMMDD_HHMM_sep_xxx.sql`，如 `20260612_1800_sep_semantic_profiles.sql`。

当前位置：`/workspace/resources/sql-add/`（已存在，当前为空目录）

---

## 一、当前环境

| 项目 | 值 |
|---|---|
| Docker 容器 | `offer-catcher-mysql` (Up, healthy) |
| MySQL 版本 | 8.0 |
| 数据库 | `offer_catcher` |
| 表总数 | 23 |
| 用户数 | 2 |
| 简历数 | 8（其中 4 份标记 is_current=1） |
| 岗位数 | 46 |
| 线上库可访问 | ❌ 当前环境无法访问线上库，本快照仅记录开发库 |

---

## 二、SQL 目录

| 路径 | 状态 | 说明 |
|---|---|---|
| `server/sql/init.sql` | ✅ 存在 | 全量建表脚本 |
| `server/sql/migrations/` | ✅ 3 文件 | 增量迁移：k13 tag-evidence + prod job_profiles + agent fields |
| `server/sql/seed/` | ✅ 1 文件 | `tencent_jobs.sql` |
| `resources/sql-add/` | ✅ 空目录 | 后续增量 SQL 落盘处 |
| `backups/sql/` | ✅ 存在 | 历史备份 |

---

## 三、表清单与行数

| # | 表名 | 行数 | 用途 |
|---|---|---|---|
| 1 | `users` | 2 | 用户表 |
| 2 | `resumes` | 8 | 简历主表 |
| 3 | `jobs` | 46 | 岗位主表 |
| 4 | `resume_profiles` | 16 | 简历标签画像 |
| 5 | `job_profiles` | 46 | 岗位标签画像 |
| 6 | `recommendation_results` | 43 | 推荐结果快照 |
| 7 | `match_analyses` | 113 | 匹配分析结果 |
| 8 | `resume_optimization_results` | 60 | 优化结果 |
| 9 | `resume_compare_results` | 0 | 对比结果（空） |
| 10 | `resume_educations` | 31 | 简历教育经历 |
| 11 | `resume_projects` | 31 | 简历项目经历 |
| 12 | `resume_internships` | 31 | 简历实习经历 |
| 13 | `resume_works` | 12 | 作品链接 |
| 14 | `resume_certificates` | 8 | 证书 |
| 15 | `resume_awards` | 7 | 获奖 |
| 16 | `resume_language_exams` | 3 | 语言考试 |
| 17 | `job_locations` | — | 岗位地点 |
| 18 | `job_skills` | — | 岗位技能 |
| 19 | `job_filter_options` | — | 筛选选项 |
| 20 | `job_filter_option_relations` | — | 筛选关联 |
| 21 | `job_filter_groups` | — | 筛选组 |
| 22 | `job_business_groups` | — | 事业群 |
| 23 | `seed_versions` | — | 数据版本 |

> 标记 `—` 的表属于辅助/配置表，未统计行数。`agent_runs`、`agent_results` 表不存在。

---

## 四、简历表（resumes）

### 结构
```
id (bigint PK), user_id (bigint), name (varchar 100), is_current (tinyint),
real_name (varchar 50), phone (varchar 50), email (varchar 255), city (varchar 100),
professional_skills (text), ai_application (text), homepage (varchar 500),
additional_info (text), raw_text (mediumtext), created_at, updated_at
```

### 数据
| id | is_current | created | updated |
|----|-----------|---------|---------|
| 1 | 1 | 2026-06-07 | 2026-06-07 |
| 2 | 0 | 2026-06-07 | 2026-06-07 |
| 3 | 0 | 2026-06-07 | 2026-06-07 |
| 4 | 0 | 2026-06-07 | 2026-06-07 |
| 5 | 0 | 2026-06-07 | 2026-06-07 |
| 6 | 1 | 2026-06-08 | 2026-06-08 |
| 7 | 1 | 2026-06-08 | 2026-06-08 |
| 8 | 1 | 2026-06-08 | 2026-06-08 |

> ⚠️ 4 份简历 is_current=1，不是唯一当前值。此状态取决于应用层实现。

---

## 五、岗位表（jobs）

- 总数：46 个岗位（来自腾讯校招 seed 数据）
- 含方向：后端、前端、算法、数据、产品、运营、设计、AI 等

---

## 六、Profile 表

### 6.1 resume_profiles

**结构核心字段**：
```
id, resume_id, user_id, profile_type(varchar 20), profile_version(varchar 20),
status(varchar 20), dominant_profile(varchar 80),
primary_tags_json, candidate_tags_json, risk_tags_json, evidence_json,
profile_summary(text), source(varchar 50), generator(varchar 50), error_message(text),
resume_updated_at, generated_at, created_at, updated_at
```

**数据分布**：
| profile_type | version | status | source | 数量 |
|---|---|---|---|---|
| fast | v1 | ready | rule | 8 |
| quality | v1 | stale | rule | 6 |
| quality | v1 | ready | rule | 2 |

**缺失字段**：`content_hash`、`model_name`、`reasoning_effort`

### 6.2 job_profiles

**结构核心字段**：
```
id, job_id, direction, category,
primary_tags_json, skill_tags_json, requirement_tags_json,
responsibility_tags_json, risk_tags_json, evidence_json,
source(varchar 50), version(varchar 50), created_at, updated_at
```

**数据分布**：
| version | source | 数量 |
|---|---|---|
| tag_agent_v1 | rule | 46 |

**缺失字段**：`profile_status`、`content_hash`、`model_name`

---

## 七、结果表

### 7.1 recommendation_results（43 条）
```
id, user_id, resume_id, result_json(json), created_at, generation_tier, generated_by,
status, profile_type, profile_dominant_profile, updated_at
```

- 最近推荐：2026-06-12 resume_id=8
- generated_by 含：`llm_agent`、`fast_tag_engine`、`mock_agent`
- 所有记录 `result_json IS NOT NULL`

### 7.2 match_analyses（113 条）
```
id, user_id, resume_id, job_id, score(int), result_json(json), created_at,
generation_tier, generated_by, status, profile_type, profile_dominant_profile, updated_at
```

- 最近匹配皆 `mock_agent`（均为早期测试数据）
- 含 Fast 和 Quality tier

### 7.3 resume_optimization_results（60 条）
```
id, user_id, resume_id, job_id, result_json(json), created_at, optimization_mode,
generation_tier, generated_by, status, profile_type, profile_dominant_profile,
target_direction, updated_at
```

- 含 `fast`/`quality` tier，`target_job`/`target_direction` mode
- generated_by 含 `mock_agent`、`mock_quality`

### 7.4 resume_compare_results（0 条）
- 表存在但无数据

---

## 八、SEP 重构相关发现

| 待补充字段 | resume_profiles | job_profiles |
|---|---|---|
| `content_hash` | ❌ 缺失 | ❌ 缺失 |
| `model_name` | ❌ 缺失 | ❌ 缺失 |
| `reasoning_effort` | ❌ 缺失 | ❌ 缺失 |
| `profile_status` | ✅ 有 `status`（ready/stale/pending） | ❌ 缺失 |
| 独立 tags 表 | ❌ 标签以 JSON 存储 | ❌ 同上 |

**SEP 需要新建的表**（后续落盘 `/workspace/resources/sql-add`）：
1. `semantic_profiles` — 统一简历/岗位语义画像
2. `semantic_profile_tags` — 标签详情（label/confidence/evidence/explanation）

---

## 九、迁移注意事项

1. **所有增量 SQL 必须落盘 `/workspace/resources/sql-add`**，不修改 `init.sql`
2. SEP 新表与旧 `resume_profiles`/`job_profiles` 共存，渐进迁移
3. `resume_profiles` 存量 16 条、`job_profiles` 存量 46 条需评估是否迁移至 SEP 新表
4. 结果表（recommendation/match/optimize）当前仅存 JSON 快照，不需要立即迁移
5. 线上库可能结构与开发库一致，但需单独验证后再设计迁移计划
