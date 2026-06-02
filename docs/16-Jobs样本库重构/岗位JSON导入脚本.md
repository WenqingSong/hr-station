# J2：岗位 JSON 导入脚本

## 1. 前置条件

| 条件 | 状态 |
|------|------|
| J0 资源校验 | ✅ |
| J1 数据库结构重构 | ✅ |
| `resources/jobs/tencent-campus-jobs.json` | ✅ 162KB, 46 岗位 |

## 2. 导入脚本

- **路径**：`server/cmd/import_jobs/main.go`
- **参数**：`--file`（JSON 路径）、`--dry-run`（仅验证）、`--replace`（事务导入）

## 3. 导入结果

| 指标 | 结果 |
|------|------|
| deleted_old_jobs | 8（旧 local_mock） |
| jobs (manual_curated) | 46 |
| job_locations | 180 |
| job_business_groups | 194 |
| job_skills | 1402 |
| source | manual_curated / manual_verified |

## 4. API 验证

- `GET /api/jobs` → code=0, total=46 ✅
- `GET /api/jobs/facets` → directions=28, categories=46 ✅
- facets cities 仅 3 个（因旧 API 查询 `jobs.city` 而非 `job_locations`）→ **记录到 J3**

## 5. 命令

```bash
cd server
go run ./cmd/import_jobs --file ../resources/jobs/tencent-campus-jobs.json --dry-run
go run ./cmd/import_jobs --file ../resources/jobs/tencent-campus-jobs.json --replace
```

## 6. 未做事项

- 未修改前端
- 未修改 Jobs API
- 未修改 Agent/LLM/Auth
- 未修改 resources JSON

## 7. P0/P1/P2

| 级别 | 问题 |
|------|------|
| — | 无。J2-R 已修复 FK 安全，orphan=0 |
