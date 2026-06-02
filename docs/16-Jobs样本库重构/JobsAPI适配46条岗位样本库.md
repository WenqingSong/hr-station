# J3：Jobs API 适配 46 条岗位样本库

## 1. 修改了哪些后端文件

| 文件 | 修改 |
|------|------|
| `model/job.go` | 新增 15 个 J1 字段 |
| `repository/job_repository.go` | 扩展 select 列 + city/bg/skill 筛选 + facets 从关联表聚合 |
| `service/job_service.go` | 扩展 JobResponse 含完整 JD 字段 + JobFacetsResponse 新增 business_groups/skills 等 |
| `handler/job_handler.go` | 新增 business_group/skill/application_* 查询参数 |

## 2. 没有修改

- 前端 ✅
- 导入脚本 ✅
- resources JSON ✅
- Agent/LLM/Auth ✅

## 3. API 参数

| 参数 | 来源表 | 已验证 |
|------|--------|--------|
| keyword | jobs.title/description/requirements/skills_json | ✅ total=6 for "Go" |
| direction | jobs.direction | ✅ |
| city | job_locations.city | ✅ total=44 for "深圳总部" |
| category | jobs.category | ✅ |
| business_group | job_business_groups.code | ✅ total=37 for "CSIG" |
| skill | job_skills.skill | ✅ |
| page/page_size | — | ✅ page=2 returns different items |

## 4. Facets 聚合来源

| 聚合 | 来源 |
|------|------|
| cities | job_locations (17 个城市) |
| business_groups | job_business_groups (8 个 BG) |
| skills | job_skills (50 个 skills) |
| directions/categories/job_types | jobs 表 |

## 5. Detail 返回字段

标题、application_type、slogan、locations、organizations、responsibilities、must_have、nice_to_have、bonus_points、skills、source/data_quality 全返回。

## 6. 验证结果

- `go build ./cmd/server` ✅
- `go test ./...` ✅ (10 packages PASS)
- facets cities = 17 (旧 API 仅 3 个) ✅
- city 筛选通过 job_locations ✅
- business_group 筛选通过 job_business_groups ✅

## 7. P0/P1/P2

| 级别 | 问题 |
|------|------|
| — | 无 |
