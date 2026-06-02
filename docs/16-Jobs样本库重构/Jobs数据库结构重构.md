# J1：Jobs 数据库结构重构

## 1. 目标

基于 `resources/jobs/tencent-campus-jobs.json` 的字段结构重构 jobs 表，为 J2 导入脚本做准备。采用**兼容式重构**：保留旧字段，仅追加新字段和关联表。

## 2. 旧表问题

| 问题 | 详情 |
|------|------|
| 缺少 application_type | JSON 中 46 岗位均为 `应届实习`，旧表无此字段 |
| 缺少 slogan/summary | 岗位宣传语和概要无处存储 |
| 缺少结构化 JD | responsibilities/must_have/nice_to_have/bonus_points 仅能存为单 TEXT |
| 无关联表 | locations/skills/business_groups 无法做筛选聚合 |
| source 命名混乱 | 仅有 `local_mock` / `tencent_join`，无 `manual_curated` |

## 3. 新增字段（17 个）

| 字段 | 类型 | 用途 |
|------|------|------|
| `application_type` | VARCHAR(64) | 应届实习/校园招聘 |
| `application_project` | VARCHAR(64) | 实习生/应届生 |
| `application_subtype` | VARCHAR(64) | 申请子类型 |
| `slogan` | VARCHAR(200) | 岗位宣传语 |
| `summary` | TEXT | 岗位概要 |
| `responsibilities_json` | JSON | 职责列表 |
| `must_have_json` | JSON | 必备要求 |
| `nice_to_have_json` | JSON | 了解项 |
| `bonus_points_json` | JSON | 加分项 |
| `notes_json` | JSON | 备注 |
| `skills_json` | JSON | 技能列表 |
| `organizations_json` | JSON | 事业群 |
| `locations_json` | JSON | 工作城市 |
| `interview_locations_json` | JSON | 面试城市 |
| `facets_json` | JSON | 聚合原始数据 |
| `source_name` | VARCHAR(200) | 数据来源名称 |
| `collected_at` | DATETIME | 采集时间 |

## 4. 新增关联表（3 个）

| 表 | 字段 | 用途 |
|----|------|------|
| `job_locations` | job_id, city | 城市筛选聚合 |
| `job_business_groups` | job_id, code, name | 事业群筛选聚合 |
| `job_skills` | job_id, skill | 技能筛选聚合 |

## 5. JSON → DB 映射

| JSON 路径 | DB 字段/表 |
|-----------|-----------|
| source.source_type | external_source (= manual_curated) |
| source.source_name | source_name |
| source.source_url | source_url |
| source.data_quality | data_quality (= manual_verified) |
| source.collected_at | collected_at |
| job.id | external_id |
| job.title | title |
| job.job_type | job_type |
| job.application_type | application_type |
| job.direction | direction |
| job.category | category |
| job.status | status |
| job.slogan | slogan |
| description.summary | summary |
| description.responsibilities | responsibilities_json |
| requirements.must_have | must_have_json |
| requirements.nice_to_have | nice_to_have_json |
| requirements.bonus_points | bonus_points_json |
| skills | skills_json + job_skills |
| locations | locations_json + job_locations |
| organizations | organizations_json + job_business_groups |
| interview_locations | interview_locations_json |
| facets | facets_json |

## 6. 兼容性

- **旧字段全保留**（department, city, description, requirements, skill_keywords_json 等）
- 新字段全部有 DEFAULT（' ' 或 NULL），不影响旧 INSERT
- 旧 8 条 local_mock seed 保留到 J2
- 后端 Go build ✅，API 仍可正常返回旧岗位

## 7. 未做

- 未导入 JSON（J2）
- 未修改后端 API（J3）
- 未修改前端（J4）
- 未删除旧字段

## 8. P0/P1/P2

| 级别 | 问题 |
|------|------|
| — | 无。全部通过。 |
