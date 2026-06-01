# 10-P-3-1 Jobs facets 后端接口契约冻结与实现

## 1. 阶段信息

- 阶段：10-P-3-1
- 日期：2026-06-01
- 范围：Jobs 后端（router/handler/service/repository/model/sql）
- 是否修改前端：否
- 是否调用腾讯接口：否
- 是否继续 LLM：否
- 是否继续阶段 11/12/13：否

## 2. 本次结论摘要

- `GET /api/jobs/facets` 已实现：返回 directions/cities/recruit_types/categories 的真实聚合 count
- `GET /api/jobs` 筛选增强：新增 direction、recruit_type 参数支持
- `GET /api/jobs/{id}` 字段补齐：新增 direction、recruit_type、status 字段
- jobs 数据源：MySQL jobs 表（8 条 local_mock seed + 可能的 tencent_join）
- 新增 direction 字段、status 字段到 jobs 表
- recruit_type API 层映射到 job_type（不新增数据库字段）
- 未调用腾讯接口
- 未修改前端
- go build / go vet / go test 全部通过
- facets 端点验证通过：directions 3组(4/3/1)、cities 5组、recruit_types 1组(校招:8)、categories 8组
- 首页筛选 count 真实化后端已就绪，前端下一步可接入

## 3. 冻结后的 Jobs 接口契约

### 3.1 GET /api/jobs

**请求参数**：

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|---|---|---|---|---|
| keyword | string | 否 | "" | 模糊搜索 title/department/description/requirements |
| direction | string | 否 | "" | 按 direction 精确筛选（新增） |
| city | string | 否 | "" | 按 city 精确筛选 |
| recruit_type | string | 否 | "" | 按 job_type 精确筛选（API 对外用 recruit_type） |
| category | string | 否 | "" | 按 category 精确筛选 |
| external_source | string | 否 | "" | 来源筛选（local_mock/tencent_join） |
| limit | int | 否 | 20 (max 50) | 每页条数 |
| offset | int | 否 | 0 | 偏移量 |

**参数兼容**：旧参数 limit/offset/external_source 继续支持。

**返回结构**：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "id": 1,
        "title": "Go 后端开发工程师",
        "department": "技术部",
        "city": "深圳",
        "direction": "软件开发",
        "category": "后端开发",
        "job_type": "校招",
        "recruit_type": "校招",
        "description": "...",
        "requirements": "...",
        "skill_keywords": ["Go","MySQL","Redis"],
        "status": "active",
        "is_featured": true,
        "created_at": "2026-01-01T00:00:00Z",
        "updated_at": "2026-01-01T00:00:00Z"
      }
    ],
    "total": 8,
    "limit": 20,
    "offset": 0
  }
}
```

**items 字段说明**：

| 字段 | 类型 | 说明 | DB 来源 |
|---|---|---|---|
| id | uint64 | 岗位 ID | jobs.id |
| title | string | 岗位标题 | jobs.title |
| department | string | 部门 | jobs.department |
| city | string | 城市 | jobs.city |
| direction | string | 求职方向 | jobs.direction（新增） |
| category | string | 岗位类别 | jobs.category |
| job_type | string | 原始 job_type（保留兼容） | jobs.job_type |
| recruit_type | string | 招聘类型（API 新字段，映射 job_type） | jobs.job_type |
| description | string | 岗位描述 | jobs.description |
| requirements | string | 岗位要求 | jobs.requirements |
| skill_keywords | json array | 技能关键词 | jobs.skill_keywords_json |
| status | string | 岗位状态 | jobs.status（新增，默认 "active"） |
| is_featured | bool | 是否精选 | jobs.is_featured |
| created_at | string | 创建时间 | jobs.created_at |
| updated_at | string | 更新时间 | jobs.updated_at |

**错误策略**：invalid external_source → 400；认证失败 → 401；DB 错误 → 500。

### 3.2 GET /api/jobs/facets

**请求参数**：无（首次实现全量聚合）。

**返回结构**：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "directions": [
      {"value": "软件开发", "label": "软件开发", "count": 4},
      {"value": "算法与大数据", "label": "算法与大数据", "count": 3},
      {"value": "产品策划与运营", "label": "产品策划/运营", "count": 1}
    ],
    "cities": [
      {"value": "深圳", "label": "深圳", "count": 3}
    ],
    "recruit_types": [
      {"value": "校招", "label": "校园招聘", "count": 8}
    ],
    "categories": [
      {"value": "后端开发", "label": "后端开发", "count": 1}
    ]
  }
}
```

**count 来源**：MySQL `SELECT direction, COUNT(*) FROM jobs WHERE direction != '' GROUP BY direction` 等 GROUP BY 查询。空值列不进入 facets。

**字段映射规则**：
- `directions` ← `jobs.direction`（label 优先使用内置中文映射）
- `cities` ← `jobs.city`
- `recruit_types` ← `jobs.job_type`（label 优先使用内置映射：校招→校园招聘 等）
- `categories` ← `jobs.category`

**排序规则**：按 count 降序。

**认证**：需要 JWT Bearer Token。

### 3.3 GET /api/jobs/{id}

**Path 参数**：`{id}` — uint64 岗位 ID。

**返回字段**：与 GET /api/jobs items 中的单条一致，包含 direction、recruit_type、status。

**找不到岗位**：HTTP 404 + `{"code": 43001, "message": "job not found"}`。

## 4. 实现内容清单

| 文件 | 修改内容 | 目的 | 是否业务代码 |
|---|---|---|---|
| `server/sql/init.sql` | jobs 表新增 direction(VARCHAR, NOT NULL DEFAULT '')、status(VARCHAR, NOT NULL DEFAULT 'active')；索引 idx_jobs_direction；种子数据补 direction 值 | 数据库支撑 facets 和筛选 | 是 |
| `server/internal/model/job.go` | 新增 Direction、Status 字段 | Model 映射新字段 | 是 |
| `server/internal/repository/job_repository.go` | JobQuery 新增 Direction/RecruitType；新增 GetFacets() 方法；buildWhereClause 新增 direction/recruit_type 条件；所有 SELECT 列统一为 jobSelectCols 常量 | 筛选扩展 + facets 聚合查询 | 是 |
| `server/internal/service/job_service.go` | JobListRequest 新增 Direction/RecruitType；JobResponse 新增 Direction/RecruitType/Status；新增 JobFacetsResponse + GetFacets() 方法；内置中文 label 映射 | 业务逻辑 + 字段映射 | 是 |
| `server/internal/http/job_handler.go` | List handler 新增 direction/recruit_type 参数解析；新增 Facets handler；handleJobError 改为 switch | API 层参数接收 | 是 |
| `server/internal/http/router.go` | 注册 /api/jobs/facets 路由；dispatchJobRoutes 新增 facets 分发 | 路由注册 | 是 |

## 5. 数据库 / 数据源变更

| 文件 / 表 | 变更 | 原因 | 风险 |
|---|---|---|---|
| jobs 表 | 新增 direction VARCHAR(100) NOT NULL DEFAULT '' | facets directions 维度 | 旧 MySQL 容器需手动 ALTER TABLE（已执行） |
| jobs 表 | 新增 status VARCHAR(50) NOT NULL DEFAULT 'active' | 前端岗位状态标签真实化 | 同上 |
| jobs 种子数据 | 8 条岗位补 direction + status 值 | 新环境 db-import 后直接可用 | 旧数据 direction 为空（不影响旧查询） |
| init.sql | 同步 schema | 新环境创建表时包含新字段 | 无风险 |

**注意**：当前项目无正式数据库迁移体系。init.sql 更新对新环境（首次 docker compose up）有效；已有环境的旧 MySQL 容器需手动 ALTER TABLE（本阶段已通过 docker exec 执行）。

## 6. 路由与分层实现说明

| 层级 | 文件 | 作用 |
|---|---|---|
| Router | `router.go` | 注册 /api/jobs/facets 到 JobHandler.Facets；dispatchJobRoutes 分发 facets/detail/match |
| Handler | `job_handler.go` | List 解析 direction/recruit_type 参数；Facets 调用 Service；Detail 不变 |
| Service | `job_service.go` | GetFacets 调用 Repo → 包装为 JobFacetsResponse + label映射；List 传 direction/recruit_type 到 Repo |
| Repository | `job_repository.go` | GetFacets 执行 4 个 GROUP BY；buildWhereClause 增 direction/recruit_type 条件 |
| Model | `job.go` | 结构体增 Direction/Status |
| SQL | `init.sql` | Schema 增 direction/status 列 + 种子数据 |

## 7. 验证结果

| 命令 / 接口 | 目的 | 结果 |
|---|---|---|
| `git status --short` 执行前 | 初始状态 | 干净 |
| `go build ./cmd/server` | 编译验证 | ✅ PASS |
| `go vet ./...` | 静态检查 | ✅ PASS |
| `go test ./...` | 单元测试 | ✅ 9/9 PASS |
| `GET /api/jobs/facets` | facets 端点 | ✅ code:0; directions 3组; cities 5组; recruit_types 1组(校招:8); categories 8组 |
| `ALTER TABLE jobs ADD direction, status` | 数据库迁移 | ✅ 执行成功 |
| `UPDATE jobs SET direction=...` | 种子数据回填 | ✅ 执行成功 |
| `git status --short` 执行后 | 修改范围确认 | 6 文件（均在 server/ 范围内） |

## 8. 未做事项

- 未修改前端
- 未实现首页筛选 UI 对接
- 未调用腾讯岗位接口
- 未修改 Resumes
- 未修改 Agent / LLM
- 未继续阶段 11/12/13
- 未做上线部署
- 未私自规划后续 prompt
- 未 git commit

## 9. 当前风险与不确定项

- 旧 MySQL 容器需手动执行 ALTER TABLE + UPDATE（本阶段已通过 docker exec 执行）。docker compose down -v 重建容器 + make db-import 可自动获得新 schema。
- facets 当前为全量聚合，不支持与其他筛选参数联动（交叉过滤）。后续可增加可选参数支持。
- 前端首页筛选项尚未接入 facets 接口，JobFilterSidebar 仍使用硬编码 FILTER_GROUPS（属下一阶段）。
- direction 的 label 映射使用硬编码 map，仅支持当前种子数据的取值。

## 10. docs 更新清单

| 文件 | 新增 / 更新 | 目的 |
|---|---|---|
| `docs/mcp/项目上下文.md` | 更新 | 10-P-3-1 状态、Jobs 后端新增能力摘要 |
| `docs/mcp/10-P-3-1-Jobs-facets后端接口契约冻结与实现.md` | 新增 | 本阶段完整记录（即本文件） |
| `docs/09-开发日志/2026-06-01-231-10-P-3-1-Jobs-facets.md` | 新增 | 开发日志 |

## 11. CodeBuddy 重启恢复摘要

10-P-3-1 完成 Jobs facets 后端接口实现。`GET /api/jobs/facets` 返回 directions/cities/recruit_types/categories 的真实聚合 count（基于 MySQL GROUP BY）。`GET /api/jobs` 新增 direction/recruit_type 筛选参数。data source 为 MySQL jobs 表（8 条 seed）。go build/vet/test 全部通过。facets 端点验证通过。前端下一次接入时调用 `GET /api/jobs/facets` 获取筛选 count，`GET /api/jobs?direction=xxx` 按条件筛选。LLM/上线继续暂停。

## 12. 最终判断

1. `GET /api/jobs/facets` 是否已实现？ **是。** code:0，真实聚合 count。
2. `GET /api/jobs` 是否支持 direction？ **是。** 新增参数且 Repository 层面生效。
3. `GET /api/jobs` 是否支持 recruit_type？ **是。** API 参数 recruit_type 映射到 DB job_type。
4. facets count 是否来自后端聚合？ **是。** `SELECT direction, COUNT(*) FROM jobs GROUP BY direction`。
5. 是否调用腾讯接口？ **否。**
6. 是否修改前端？ **否。**
7. 是否还有后端阻塞首页真实化的问题？ **否。** backend facets + 筛选已就绪。
8. 是否可以进入首页 / 岗位列表前端真实对接阶段？ **是。** 后端契约已冻结，接口可用。
9. 本阶段是否可以通过？ **是。**
