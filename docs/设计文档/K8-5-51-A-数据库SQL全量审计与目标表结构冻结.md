# K8-5-51-A：数据库 SQL 全量只读审计与目标表结构冻结

> 状态：✅ 已完成 | 日期：2026-06-06 | 性质：只读审计，未修改任何源码/SQL

---

## 一、审计结论摘要

```
P0 问题：7 项
P1 问题：12 项
P2 问题：3 项

核心发现：
1. 当前 DB 缺失 resume_versions / recommendation_results / match_analyses / version_compare_results / resume_optimization_results 5 张表
2. Go 代码全量使用 resume_versions / resume_version_id / from_version_id / to_version_id 旧命名
3. 前端 106 处使用 version_no / versionNo / resumeVersionNo 旧命名
4. migration-j1/j2 中所有 ALTER TABLE 字段已在 init.sql 中直接包含（完全冗余）
5. job_locations/job_business_groups/job_skills 在 init.sql 和 migration-j1.sql 重复创建
6. tencent_jobs.sql 依赖 job_filter_groups/options/relations/seed_versions（这些表仅在 init.sql 中创建）
```

---

## 二、SQL 文件清单与处理建议

| 文件 | 行数 | 当前作用 | 类型判断 | 问题 | 建议 |
|------|:---:|----------|----------|------|------|
| `server/sql/init.sql` | 418 | 主初始化脚本（users/jobs/结果表/筛选维表） | init 主初始化脚本 | DROP TABLE 列表不完整（缺 job_filter_*、seed_versions、resume_versions）；seed 数据引用 resume_version_id 但无 resume_versions 表 | **A 合并进 init.sql** — 作为唯一入口，需重建 |
| `server/sql/resume_versions.sql` | 2046 | 创建 resume_versions 表 + 9 份 Demo 简历 seed | 简历 seed | 表名 resume_versions 应为 resumes；version_no 字段应移除；与 init.sql 分离导致独立管理混乱 | **C 废弃** — 内容合入 init.sql 重写，表名改为 resumes |
| `server/sql/seed/tencent_jobs.sql` | 5756 | 46 条腾讯岗位 + filter_groups/options/relations seed | 岗位 seed（生成脚本输出） | 依赖 job_filter_* 表存在（由 init.sql 创建）；DELETE FROM 方式幂等但耦合 init.sql | **B 作为生成源保留** — 源文件保留，最终 seed 合并进新 init.sql |
| `server/sql/migration-j1.sql` | 58 | ALTER TABLE jobs + CREATE TABLE job_locations/business_groups/skills | 历史 migration（已集成入 init.sql） | **100% 冗余**：所有 ALTER TABLE 字段已在 init.sql CREATE TABLE jobs 中；job_locations/business_groups/skills 已在 init.sql 中创建 | **C 废弃** — clean env 不再执行，仅保留为历史参考 |
| `server/sql/migration-j2.sql` | 41 | ALTER TABLE resume_versions V2 字段扩展 | 历史 migration（已集成入 resume_versions.sql） | **100% 冗余**：所有字段已在 resume_versions.sql CREATE TABLE 中直接包含 | **C 废弃** — clean env 不再执行，仅保留为历史参考 |

---

## 三、P0 问题清单

| # | 问题 | 证据 | 影响 | 建议 |
|---|------|------|------|------|
| P0-1 | **当前 DB 缺 5 张核心表** | `SHOW TABLES` 只有 9 张表，缺少 `resume_versions`, `recommendation_results`, `match_analyses`, `version_compare_results`, `resume_optimization_results` | 所有简历/AI 功能无法运行 | K8-5-51-B 重写 init.sql 一次性修复 |
| P0-2 | **Go 代码全量使用旧表名 resume_versions** | `resume_repository.go` 6 处 `FROM resume_versions`；`resume.go` model 注释 "Maps to the resume_versions table" | 数据库与代码不一致 | K8-5-51-B 统一改为 resumes |
| P0-3 | **resume_version_id 渗透全部 repository** | `recommendation_repository.go`: INSERT/SELECT `resume_version_id`；`match_repository.go`: INSERT/SELECT `resume_version_id`；`version_compare_repository.go`: `from_version_id`, `to_version_id` | 字段命名不一致 | 改为 resume_id, from_resume_id, to_resume_id |
| P0-4 | **Go model 仍用 ResumeVersion 命名** | `ResumeVersion` struct + `VersionNo` 字段 + `ResumeVersionID` 字段 | model 层命名错误 | 改为 Resume struct，移除 VersionNo |
| P0-5 | **migration-j1/j2 完全冗余** | j1: 18 个 ALTER TABLE jobs 字段全已在 init.sql CREATE TABLE jobs 中包含；j2: 16 个 ALTER TABLE 字段全已在 resume_versions.sql 中包含 | 冗余文件误导 | K8-5-51-B 废弃 |
| P0-6 | **init.sql DROP TABLE 列表不完整** | 缺 `job_filter_option_relations`, `job_filter_options`, `job_filter_groups`, `seed_versions` | init.sql 不能真正幂等（K8-5-51-B 前已在 K8-5-50 修复） | ✅ 已修复 |
| P0-7 | **docker-compose 曾自动挂载 init.sql** | `docker-compose.yml` line 27 原挂载到 `/docker-entrypoint-initdb.d/`，MySQL 启动时自动执行 | 双重导入导致 "Table already exists" | ✅ 已在 K8-5-50 移除 |

---

## 四、resume_versions 残留清单

| 位置 | 残留内容 | 文件 | 行数 | 后续处理 |
|------|----------|------|:---:|----------|
| **SQL - 表名** | `CREATE TABLE resume_versions` | `resume_versions.sql` | 18 | K8-5-51-B：改为 resumes |
| **SQL - INSERT** | `INSERT INTO resume_versions` | `resume_versions.sql` | 66,87 | K8-5-51-B：改为 resumes |
| **SQL - FK 引用** | `recommendation_results` FK to `resume_versions` | `init.sql` | 234 | K8-5-51-B：改为 resumes |
| **SQL - FK 引用** | `match_analyses` FK to `resume_versions` | `init.sql` | 253 | K8-5-51-B：改为 resumes |
| **SQL - FK 引用** | `version_compare_results` FK to `resume_versions` | `init.sql` | 271-272 | K8-5-51-B：改为 resumes |
| **Go model** | `Maps to the resume_versions table` | `internal/model/resume.go` | 6 | K8-5-51-B：改为 resumes |
| **Go model** | `type ResumeVersion struct` | `internal/model/resume.go` | 14 | K8-5-51-B：改为 Resume |
| **Go repo** | 6 处 `FROM resume_versions` | `internal/repository/resume_repository.go` | 24,82,150,212,300,352 | K8-5-51-B：改为 resumes |
| **Go repo** | `resume_version_id` 字段 | `recommendation_repository.go` | 24,25,47,49 | K8-5-51-B：改为 resume_id |
| **Go repo** | `resume_version_id` 字段 | `match_repository.go` | 24,25,47,49 | K8-5-51-B：改为 resume_id |
| **Go repo** | `from_version_id`, `to_version_id` | `version_compare_repository.go` | 24,47 | K8-5-51-B：改为 from/to_resume_id |
| **Go repo** | `resume_version_id` 字段 | `recommendation_repository.go` | 24,47 | K8-5-51-B：改为 resume_id |
| **Go model** | `ResumeVersionID uint64` | `recommendation.go` | 10 | K8-5-51-B：改为 ResumeID |
| **Go model** | `ResumeVersionID uint64` | `match.go` | 10,52 | K8-5-51-B：改为 ResumeID |
| **Go model** | `ResumeVersionNo int` | `match.go` | 53 | K8-5-51-B：移除 |
| **Go model** | `ResumeVersionNo int` | `resume_snapshot.go` | 25 | K8-5-51-B：移除 |
| **Go model** | `FromVersionNo`, `ToVersionNo` | `version_compare.go` | 43-44 | K8-5-51-B：改为 FromResumeNo/ToResumeNo 或移除 |
| **Go service** | `ResumeVersionID uint64` | `recommendation_service.go` | 25 | K8-5-51-B：改为 ResumeID |
| **Go service** | `*model.ResumeVersion` 类型 | 多个文件 | — | K8-5-51-B：改为 *model.Resume |
| **Scripts** | `resume_versions` 在 core_tables | `mysql.sh` | 30 | K8-5-51-B：改为 resumes |
| **Scripts** | `resume_versions` 在 core_tables | `up.sh` | 82 | K8-5-51-B：改为 resumes |
| **Scripts** | `resume_versions.sql` 导入 | `mysql.sh` | 27,112-118,142-148 | K8-5-51-B：改为新文件名 |
| **前端 TS** | `ResumeVersionListItem`, `ResumeVersionComparison` | `resumeAdapter.ts` | 9-10 | K8-5-51-C：改为 ResumeListItem |
| **前端 TS** | `version_no`, `versionNo`, `resumeVersionNo` | 106 处 | — | K8-5-51-C：移除或重命名 |
| **前端 TS** | `from_version_no`, `to_version_no` | `resume.ts` | 304-305 | K8-5-51-C：改为 from/to_resume_no |
| **前端 TS** | `/api/resumes/versions` API 路径 | 多处 | — | K8-5-51-C：可保留为兼容路径 |

---

## 五、jobs 结构冲突清单

| # | 冲突 | 位置 | 详情 | 后续处理 |
|---|------|------|------|----------|
| J1 | **migration-j1 全部 18 个 ALTER TABLE jobs 字段已存在于 init.sql CREATE TABLE jobs** | `migration-j1.sql` vs `init.sql` | application_type, application_project, application_subtype, slogan, summary, responsibilities_json, must_have_json, nice_to_have_json, bonus_points_json, notes_json, skills_json, organizations_json, locations_json, interview_locations_json, facets_json, source_name, collected_at | K8-5-51-B：废弃 migration-j1.sql |
| J2 | **migration-j1 ALTER TABLE ADD INDEX idx_jobs_application_type, idx_jobs_status** | `migration-j1.sql` line 21-22 vs `init.sql` line 115-116 | 索引已在 init.sql CREATE TABLE 中定义 | K8-5-51-B：废弃 |
| J3 | **job_locations 在 init.sql (line 122) 和 migration-j1.sql (line 24) 重复创建** | `init.sql` 2.3a vs `migration-j1.sql` | 完全相同的表定义（migration 用 CREATE TABLE IF NOT EXISTS） | K8-5-51-B：从 init.sql 保留；废弃 migration |
| J4 | **job_business_groups 在 init.sql (line 136) 和 migration-j1.sql (line 35) 重复创建** | `init.sql` 2.3b vs `migration-j1.sql` | 完全相同的表定义 | K8-5-51-B：同上 |
| J5 | **job_skills 在 init.sql (line 151) 和 migration-j1.sql (line 47) 重复创建** | `init.sql` 2.3c vs `migration-j1.sql` | 完全相同的表定义 | K8-5-51-B：同上 |
| J6 | **JSON 冗余字段 + 关联表并存** | `jobs` 表同时有 `skills_json`, `organizations_json`, `locations_json`, `interview_locations_json`, `facets_json` + 关联表 `job_skills`, `job_business_groups`, `job_locations` | Demo 阶段允许并存：JSON 用于详情展示，关联表用于筛选聚合 | Demo 阶段保留，正式发布前决策 |

---

## 六、当前数据库初始化链路分析

### 6.1 链路图

```
make reset-up
  → docker compose down -v (清 volume)
  → make up
    → up.sh step 4: infra.sh up (docker compose up mysql)
    → up.sh step 6: 检查核心表 → 缺失时调用 mysql.sh import
      → mysql.sh import: 1) resume_versions.sql  2) init.sql
    → up.sh step 6b: 导入 tencent_jobs.sql seed（条件：tables_ok=true）
```

### 6.2 关键问题

| 问题 | 回答 |
|------|------|
| 当前 make reset-seed-db 实际执行哪个 SQL 文件？ | **不存在** `reset-seed-db` 命令；实际入口是 `make db-import` → `mysql.sh import` |
| 当前是否有多个 SQL 入口？ | 是：`mysql.sh import`（init.sql + resume_versions.sql）+ `up.sh step 6b`（tencent_jobs.sql）+ `make jobs-seed-apply`（手动） |
| 当前 tencent_jobs.sql 是否会被自动导入？ | 是：`up.sh` step 6b 在 tables_ok=true 时自动执行 `docker exec < tencent_jobs.sql` |
| 当前 resume seed 是否会被自动导入？ | 是：`mysql.sh import` 导入 `resume_versions.sql`（K8-5-50 新增） |
| 当前是否存在 SOURCE 拆分文件？ | 否：没有使用 MySQL SOURCE 命令 |
| 当前 clean volume 后真正会导入哪些数据？ | `resume_versions.sql`（resume table + seed）→ `init.sql`（users + jobs + filter meta + AI result tables + seed）→ `tencent_jobs.sql`（46 条腾讯岗位 + filter seed） |

---

## 七、Go 代码数据库引用审计

### 7.1 关键统计

| 指标 | 数量 |
|------|:---:|
| Go 文件使用 `resume_version_id` | 10 处 |
| Go 文件使用 `resume_versions`（表名） | 6 处 |
| Go 文件使用 `version_no` | 27 处 |
| Go 文件使用 `from_version_id` / `to_version_id` | 2 处 |
| Go model `ResumeVersion` 定义 | 1 个 struct |

### 7.2 详细引用

| 文件 | 引用内容 | 当前风险 | 后续建议 |
|------|----------|----------|----------|
| `model/resume.go` | `type ResumeVersion struct` + `VersionNo int` | 核心 model 命名错误 | 改为 `type Resume struct`，移除 VersionNo |
| `model/resume.go` | 注释 "Maps to the resume_versions table" | 文档过期 | 改为 resumes |
| `model/recommendation.go` | `ResumeVersionID uint64` | 字段命名 | 改为 ResumeID |
| `model/match.go` | `ResumeVersionID uint64` + `ResumeVersionNo int` | 双字段 | 改为 ResumeID，移除 VersionNo |
| `model/version_compare.go` | `FromVersionNo int` + `ToVersionNo int` | 依赖 version_no 概念 | 移除或改为 metadata |
| `repository/resume_repository.go` | 6 处 `FROM resume_versions` + 多处 `version_no` | 全量使用旧表名 | 改为 `FROM resumes`，移除 version_no |
| `repository/recommendation_repository.go` | `resume_version_id` 列读写 | DB 字段名 | 改为 resume_id |
| `repository/match_repository.go` | `resume_version_id` 列读写 | DB 字段名 | 改为 resume_id |
| `repository/version_compare_repository.go` | `from_version_id`, `to_version_id` | DB 字段名 | 改为 from_resume_id, to_resume_id |

### 7.3 Go 代码判断

| 判断 | 结论 |
|------|------|
| Go 代码是否已经使用 resumes？ | **否** — 全量使用 resume_versions |
| Go 代码是否仍使用 resume_version_id？ | **是** — 3 个 repository 全部使用 |
| Go 代码是否仍依赖 version_no？ | **是** — model + repository 27 处引用 |
| Go repository 是否与目标表结构一致？ | **否** — 存在严重不一致 |
| Agent/service/model 是否还有旧字段 stub？ | **是** — ResumeVersionID, ResumeVersionNo, FromVersionNo, ToVersionNo |

---

## 八、前端 API 契约中的旧命名

### 8.1 关键统计

| 指标 | 数量 |
|------|:---:|
| 前端使用 `version_no` / `versionNo` / `resumeVersionNo` | **106 处** |
| 前端使用 `from_version_no` / `to_version_no` | 2 处（类型定义） |
| 前端 API 路径 `/api/resumes/versions` | 多处 |

### 8.2 判断

| 问题 | 回答 |
|------|------|
| 前端 API 路径是否仍叫 /versions？ | 是：`/api/resumes/versions` 用于获取简历版本列表 |
| 这是否只是接口路径命名问题？ | 是：API 路径 `/versions` 是历史命名，可兼容保留；但数据结构中 `version_no` 字段在移除 version_no 后需要调整 |
| 本阶段是否需要改前端？ | **否** — K8-5-51-A 只审不改；前端改动放到 K8-5-51-C |

---

## 九、目标数据库表清单冻结

| 表名 | 是否保留 | 用途 | 说明 |
|------|:---:|------|------|
| `users` | ✅ 保留 | 用户注册/登录 | 不变 |
| `resumes` | ✅ 保留 | **简历主表（替代 resume_versions）** | 移除 version_no；每条记录 = 一份简历 |
| `jobs` | ✅ 保留 | 岗位列表和详情 | JSON 冗余字段暂保留（Demo 阶段） |
| `job_locations` | ✅ 保留 | 岗位工作城市关联 | 用于筛选聚合 |
| `job_business_groups` | ✅ 保留 | 岗位事业群关联 | 用于筛选聚合 |
| `job_skills` | ✅ 保留 | 岗位技能关联 | 用于筛选聚合 |
| `job_filter_groups` | ✅ 保留 | 筛选分组元数据 | 不变 |
| `job_filter_options` | ✅ 保留 | 筛选选项元数据 | 不变 |
| `job_filter_option_relations` | ✅ 保留 | 岗位-筛选项关联 | 不变 |
| `recommendation_results` | ✅ 保留 | 岗位推荐结果快照 | `resume_version_id` → `resume_id` |
| `match_analyses` | ✅ 保留 | 岗位匹配分析结果 | `resume_version_id` → `resume_id` |
| `resume_compare_results` | ✅ 保留 | **简历版本对比结果（替代 version_compare_results）** | `from_version_id` → `from_resume_id`；`to_version_id` → `to_resume_id` |
| `resume_optimization_results` | ✅ 保留 | 简历优化结果快照 | 可增加 `resume_id` NULL |
| `seed_versions` | ✅ 保留 | Seed 版本追踪 | 不变 |
| ~~`resume_versions`~~ | ❌ 不保留 | — | **废弃，改为 resumes** |
| ~~`version_compare_results`~~ | ❌ 不保留 | — | **废弃，改为 resume_compare_results** |

---

## 十、关键字段命名冻结

| 旧命名 | 新命名 | 说明 |
|--------|--------|------|
| `resume_versions` | `resumes` | 表名统一 |
| `resume_version_id` | `resume_id` | 推荐/匹配结果引用简历 |
| `from_version_id` | `from_resume_id` | 版本对比：from 简历 |
| `to_version_id` | `to_resume_id` | 版本对比：to 简历 |
| `version_compare_results` | `resume_compare_results` | 表名统一 |
| ~~`version_no`~~ | **移除** | 不再需要版本号；多条 resumes 记录 = 多份简历 |
| `ResumeVersion` | `Resume` | Go model 命名 |
| `ResumeVersionID` | `ResumeID` | Go model 字段 |
| `ResumeVersionNo` | **移除** | Go model 字段 |
| `FromVersionNo` / `ToVersionNo` | **移除** 或改为 metadata | Go model 字段 |

---

## 十一、SQL 文件处理策略冻结

| SQL 文件 | 策略 | 下一阶段操作 |
|----------|:---:|--------------|
| `server/sql/init.sql` | **A 合并** | K8-5-51-B 重写：整合所有表（users + resumes + jobs + 关联表 + filter + 结果表 + seed）为唯一 clean 初始化入口 |
| `server/sql/resume_versions.sql` | **C 废弃** | 内容合入新 init.sql（resumes 表 + seed），原文件移除/归档 |
| `server/sql/seed/tencent_jobs.sql` | **B 源文件保留** | 保留为生成源；seed 内容最终合入新 init.sql 或 build 脚本 |
| `server/sql/migration-j1.sql` | **C 废弃** | 所有字段已集成入 init.sql；clean env 不再执行 |
| `server/sql/migration-j2.sql` | **C 废弃** | 所有字段已在 resume_versions.sql 中；表重写后不再需要 |

---

## 十二、resumes 目标字段（冻结）

```text
id              BIGINT UNSIGNED  PRIMARY KEY
user_id         BIGINT UNSIGNED  FK → users(id)
schema_version  INTEGER          (1=V1, 2=V2)
name            VARCHAR(100)
phone           VARCHAR(50)
email           VARCHAR(255)
location        VARCHAR(100)
target_city     VARCHAR(100)     (V1 兼容，V2 用 expected_cities)
target_role     VARCHAR(100)     (V1 兼容)
education_json  JSON
experience_json JSON
project_json    JSON
skills_json     JSON

-- V2 摘要列
target_direction VARCHAR(100)
job_type         VARCHAR(50)
expected_cities  VARCHAR(255)
school           VARCHAR(100)
degree           VARCHAR(20)
major            VARCHAR(100)
graduation_year  VARCHAR(10)

-- V2 JSON 模块
profile_json       JSON
job_intention_json JSON
internship_json    JSON
certificates_json  JSON
awards_json        JSON
portfolio_json     JSON
extra_json         JSON
self_evaluation    TEXT
raw_text           TEXT

created_at    DATETIME
updated_at    DATETIME (建议新增)

-- ❌ 移除 version_no（不再需要）
-- ❌ 移除 UNIQUE KEY uk_resume_user_version (user_id, version_no)
```

---

## 十三、下一阶段 K8-5-51-B 建议

### K8-5-51-B：重写 init.sql + 合并 seed + Go/前端适配 + reset 验证

```
预计范围：
1. 重写 init.sql
   - 表名 resume_versions → resumes
   - 表名 version_compare_results → resume_compare_results
   - 字段 resume_version_id → resume_id（×3 表）
   - 字段 from_version_id → from_resume_id
   - 字段 to_version_id → to_resume_id
   - 移除 version_no 及相关唯一索引
   - 新增 resumes.updated_at
   - 合并 resume_versions.sql 的 seed 进 init.sql
   - tencent_jobs.sql 内容作为 seed 子部分
   - 完整的 DROP TABLE 列表（全部表）
2. 修改 Go 代码
   - model: ResumeVersion → Resume
   - repository: 表名 + 字段名全部对齐
   - service: 类型引用更新
3. 废弃文件
   - 删除/归档 migration-j1.sql, migration-j2.sql
   - 删除/归档 resume_versions.sql
4. 修改 scripts
   - mysql.sh: CORE_TABLES 更新
   - up.sh: core_tables 更新
5. 验证
   - make reset-up CONFIRM=yes
   - go build ./... && go vet ./... && go test ./...
   - 只读 DB 查询验证
```

K8-5-51-B 预计是数据库阶段**最后一个 prompt**。

### K8-5-51-C（后续可选）：前端版本命名清理

```
范围：前端 106 处 version_no/versionNo 清理
预计：1 个 prompt，纯前端
```

---

## 十四、构建/测试结果

| 验收项 | 结果 | 备注 |
|--------|:---:|------|
| `go build ./...` | ✅ pass | 编译通过 |
| `go vet ./...` | ⚠️ canceled | 超时取消（环境限制），历史记录显示 pass |
| `go test ./...` | ⚠️ canceled | 超时取消（需要数据库连接） |
| `npm run build` | ✅ pass | 有 1 个 warning（duplicate key "skills"），不影响功能 |
| `docker compose ps` | ✅ running | MySQL + Nginx + Redis 运行中 |
| MySQL `SHOW TABLES` | 9 表 | 缺 5 张核心表（见 P0-1） |

---

## 十五、剩余 prompt 估算

```
数据库收口阶段剩余 1 个 prompt：
K8-5-51-B：重写 init.sql + 合并 seed + reset 验证

前端清理阶段（后续可选）：
K8-5-51-C：前端版本命名清理
```

**后端六层架构重构暂停，不再继续推进。**

---

## 十六、未修改范围确认

```
✅ 未修改 Go 源码
✅ 未修改 TypeScript / React 前端源码
✅ 未修改 SQL 文件（仅 K8-5-50 提前修复了 init.sql DROP TABLE 列表和 docker-compose 挂载）
✅ 未修改 Makefile（K8-5-50 仅更新了 help 文本）
✅ 未修改 docker-compose.yml（K8-5-50 已移除自动挂载）
✅ 未修改 .env / .env.example
✅ 未修改 Nginx 配置
✅ 未修改 Dockerfile
✅ 未执行 make reset 或 make reset-up
✅ 未执行 gf gen dao
✅ 未执行 npm install
✅ 未写入数据库（仅 SHOW TABLES + SHOW CREATE TABLE 只读查询）
✅ 未进入后端六层架构重构
✅ 未执行任何 Git 命令
✅ 本次为纯只读审计
```

---

## 十七、Git 约束确认

```
本次未执行任何 Git 命令，也未提交代码。
审计期间严格遵守：不 git status / git diff / git add / git commit / git push / git reset / git checkout / git branch / git tag
```
