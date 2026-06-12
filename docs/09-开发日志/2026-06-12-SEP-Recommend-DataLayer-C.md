# 2026-06-12 SEP-Recommend-DataLayer-C — SEP 语义画像数据层实现

## 本阶段目标

岗位推荐 Agent 全链路重构第一个实现阶段。为 SEP 语义证据画像层建立数据层基础。

## 数据库基线（DB-State-Snapshot-A 确认）

- 开发库 23 张表，执行后 25 张表（+2 SEP 表）
- 旧 `resume_profiles`/`job_profiles` 保留不变
- SQL 落盘 `/workspace/resources/sql-add`

## 新增文件

| 文件 | 内容 |
|---|---|
| `resources/sql-add/20260612_2100_sep_recommend_datalayer.sql` | 增量建表 SQL（含安全注释） |
| `server/internal/model/semantic_profile.go` | SemanticProfile + SemanticProfileTag 结构体 + 常量 |
| `server/internal/repository/semantic_profile_repository.go` | 数据访问层（CRUD + 事务 UpsertTags） |
| `server/internal/service/semantic_profile_service.go` | 服务骨架（hash + create pending + save verified + read） |

## 新增表

### semantic_profiles
```
id, target_type, target_id, profile_version, prompt_version, schema_version,
content_hash, model_provider, model_name, reasoning_effort,
summary_sentence, strengths_json, risks_json, fit_directions_json, profile_json,
status (pending/processing/verified/failed/stale), error_message, created_at, updated_at
```
索引：uk (target_type, target_id, profile_version, schema_version, content_hash)

### semantic_profile_tags
```
id, profile_id, target_type, target_id,
tag_code, tag_label, category (7 categories),
confidence, evidence_text, explanation, source_section,
is_user_visible, weight, created_at, updated_at
```
FK: profile_id → semantic_profiles(id) ON DELETE CASCADE

## 未做事项

- ❌ 未实现 Hy3 Profiler（下一阶段）
- ❌ 未接入 SemanticProfileGate
- ❌ 未接入 Fast/Deep 推荐
- ❌ 未迁移旧数据
- ❌ 未修改旧 profile 表

## 构建验证

| 项目 | 结果 |
|---|---|
| SQL 应用 | ✅ 开发库 25 表 |
| go build | ✅ 通过 |
| go vet | ✅ 通过 |
| go test | ✅ 全部通过（11 包） |
