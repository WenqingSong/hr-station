# 10-P-4-5 Jobs 模块索引、字段契约与查询稳定性核验

## 1. 阶段信息

- 阶段：10-P-4-5
- 日期：2026-06-02
- 范围：Jobs 模块后端全量 read-only audit
- 是否修改代码：否
- 是否修改前端：否
- 是否修改后端：否
- 是否修改数据库：否
- 是否接 LLM：否
- 是否 git commit：否

## 2. 本次结论摘要

- Jobs 模块当前**可支撑首页 / 岗位详情主链路**，3 个接口运行正常
- **未发现 P0**（首页核心功能可用）
- **发现 3 个 P1**：`job_type` 无索引影响 facets/筛选查询性能、前端 `Job` 类型与后端 `JobResponse` 不匹配、`recruit_type = job_type` 重复字段设计需澄清
- **发现 4 个 P2**：facets label 映射不完整、keyword 模糊搜索全表扫描风险、`external_source` 错误码未使用 AppError、`status` 筛选未暴露为查询参数
- 接口契约基本稳定，**无需变更路径或参数名**
- **前后端契约存在类型不匹配**：前端 `Job` 类型 `{business_group, locations, type_label, tags, short_description}` 与后端返回 `{department, city, job_type, skill_keywords, description}` 不对齐，但运行时通过 adapter 映射工作
- go build / go vet / go test 全部通过
- 本阶段零代码修改

## 3. Jobs 接口契约矩阵

| 接口 | 请求参数 | 返回字段 | 是否前端使用 | 当前问题 | 设计结论 |
| -- | ---- | ---- | ------ | ---- | ---- |
| `GET /api/jobs` | `keyword`, `direction`, `city`, `recruit_type`, `category`, `external_source`, `limit`, `offset` | `JobResponse`（22 字段） | 是（JobList/JobFilterSidebar） | `recruit_type` 查的是 `job_type` 列；`external_source` 错误用字符串匹配 | 参数名保持，参数→列映射需文档化 |
| `GET /api/jobs/facets` | 无 | `{directions, cities, recruit_types, categories}` 各含 `{value, label, count}` | 是（JobFilterSidebar） | `recruit_types` 按 `job_type` 聚合；job_type 无索引 | facets 名保持，需加索引 |
| `GET /api/jobs/{id}` | path 参数 `id` | 同 `JobResponse` | 是（JobDetailPage/JobDetailBanner） | 无 | 稳定可冻结 |

## 4. 字段设计冻结表

| 字段 | 数据库列 | API 字段 | 前端使用 | 当前状态 | 设计结论 |
| -- | ----- | ------ | ---- | ---- | ---- |
| `id` | `id BIGINT UNSIGNED PK` | `id` (uint64) | 是 | ✅ | 冻结（A） |
| `title` | `title VARCHAR(200)` | `title` | 是 | ✅ | 冻结（A） |
| `department` | `department VARCHAR(200)` | `department` | 前端用 `businessGroup` 映射 | ⚠️ 字段名与前端语义不一致 | 冻结（A），前端 adapter 文档化 |
| `city` | `city VARCHAR(100)` | `city` | 前端用 `locationsText` 聚合 | ⚠️ 单值 vs 数组 | 冻结（A），当前单城市可接受 |
| `direction` | `direction VARCHAR(100)` | `direction` | 筛选 facets 用 | ✅ | 冻结（A） |
| `category` | `category VARCHAR(100)` | `category` | 筛选 facets 用 | ✅ | 冻结（A） |
| `job_type` | `job_type VARCHAR(100)` | `job_type` + `recruit_type`（相同值） | 筛选用 `recruit_type` 参数 | ⚠️ 双字段返回相同值 | 冻结（A），`recruit_type` 明确为 `job_type` 别名 |
| `description` | `description TEXT` | `description` | 前端 adapter 用 `shortDescription` | ⚠️ 完整描述 vs 简短描述 | 冻结（A），adapter 截断或全量 |
| `requirements` | `requirements TEXT` | `requirements` | 前端 `requirements` | ✅ | 冻结（A） |
| `skill_keywords` | `skill_keywords_json JSON` | `skill_keywords` | 前端 adapter 映射为 `tags` | ⚠️ JSON 数组 vs string[] | 冻结（A），后端已序列化 JSON |
| `status` | `status VARCHAR(50) DEFAULT 'active'` | `status` | 前端 `statusLabel` prop | ✅ | 冻结（A） |
| `external_source` | `external_source VARCHAR(50)` | `external_source` | 筛选参数支持 | ✅ | 冻结（A） |
| `is_featured` | `is_featured TINYINT(1)` | `is_featured` | 无直接展示，用于排序 | ✅ | 冻结（A），保持内部排序 |
| `data_quality` | `data_quality VARCHAR(50)` | `data_quality` | 无前端展示 | ✅ | 冻结（A），内部指标 |
| `created_at` / `updated_at` | `DATETIME` | `created_at` / `updated_at` (ISO8601) | 前端无直接展示 | ✅ | 冻结（A） |

**分类统计：** A（已实现，可冻结）14 个、B（需修复）0 个、C（需新设计）0 个

## 5. 查询参数设计冻结表

| 参数 | 当前后端支持 | 当前前端使用 | 是否保留 | 设计结论 | 验收标准 |
| -- | ------ | ------ | ---- | ---- | ---- |
| `keyword` | ✅ LIKE 多字段模糊 | ✅ `JobFilterSidebar` | 是 | 保留 | LIKE 安全，参数化查询无注入 |
| `direction` | ✅ `WHERE direction = ?` | ✅ `JobFilterSidebar` | 是 | 保留 | 有 idx_jobs_direction 索引 |
| `city` | ✅ `WHERE city = ?` | ✅ `JobFilterSidebar` | 是 | 保留 | 有 idx_jobs_city 索引 |
| `recruit_type` | ✅ `WHERE job_type = ?` | ✅ `JobFilterSidebar` | 是 | 保留，文档化映射 | **job_type 无索引**（P1） |
| `category` | ✅ `WHERE category = ?` | ✅ `JobFilterSidebar` | 是 | 保留 | 有 idx_jobs_category 索引 |
| `external_source` | ✅ `WHERE external_source = ?` | ❌ 前端未用 | 保留 | 后端能力保留 | 有 idx_jobs_external_source 索引 |
| `status` | ❌ 未暴露 | ❌ 前端未用 | 可选 | 后续按需添加（P2） | 无需本阶段 |
| `limit` | ✅ 最大 50，默认 20 | ✅ `JobListParams` | 是 | 保留 | service 层硬限制 |
| `offset` | ✅ 默认 0 | ✅ `JobListParams` | 是 | 保留 | 安全 |
| `page` / `page_size` | ❌ 不支持 | ❌ 前端未用 | 否 | 不添加，用 limit/offset | 无需 |

## 6. 索引与查询性能核验

| 查询场景 | 当前 WHERE / GROUP BY / ORDER BY | 当前索引 | 风险等级 | 设计结论 |
| ---- | ------------------------------ | ---- | ---- | ---- |
| city 筛选 | `WHERE city = ?` | `idx_jobs_city` | — | ✅ 已覆盖 |
| direction 筛选 | `WHERE direction = ?` | `idx_jobs_direction` | — | ✅ 已覆盖 |
| category 筛选 | `WHERE category = ?` | `idx_jobs_category` | — | ✅ 已覆盖 |
| job_type 筛选 | `WHERE job_type = ?` | **无索引** | P1 | **需添加** `idx_jobs_job_type` |
| external_source 筛选 | `WHERE external_source = ?` | `idx_jobs_external_source` | — | ✅ 已覆盖 |
| status 筛选 | 未暴露为参数 | 无 | P2 | 后续按需 |
| keyword 模糊搜索 | `LIKE %keyword%` 四字段 | `idx_jobs_title` 仅覆盖 title | P2 | 全表扫描风险，8 条 seed 数据无影响 |
| facets GROUP BY direction | `GROUP BY direction` | `idx_jobs_direction` | — | ✅ 覆盖 |
| facets GROUP BY city | `GROUP BY city` | `idx_jobs_city` | — | ✅ 覆盖 |
| facets GROUP BY job_type | `GROUP BY job_type` | **无索引** | P1 | **需添加** `idx_jobs_job_type` |
| facets GROUP BY category | `GROUP BY category` | `idx_jobs_category` | — | ✅ 覆盖 |
| ORDER BY is_featured DESC, id ASC | `is_featured` + `id` | `idx_jobs_is_featured` + 主键 | — | ✅ 覆盖 |

**索引修复设计（下一个 prompt 执行）：**

```sql
ALTER TABLE jobs ADD INDEX idx_jobs_job_type (job_type);
ALTER TABLE jobs ADD INDEX idx_jobs_status (status);
```

- `idx_jobs_job_type`：覆盖 `recruit_type` 筛选 + facets GROUP BY job_type
- `idx_jobs_status`：预留 status 筛选能力

仅需 2 个索引，不盲目加联合索引。

## 7. 前后端契约一致性核验

| 前端位置 | 使用字段 / 参数 | 后端是否稳定提供 | 问题 | 设计结论 |
| ---- | --------- | -------- | -- | ---- |
| `api/jobs.ts` | `JobListParams`: keyword/direction/city/recruit_type/category/limit/offset | ✅ 全部支持 | `recruit_type` 后端查 `job_type` 列，已文档化 | B |
| `api/jobs.ts` | `JobFacets`: directions/cities/recruit_types/categories | ✅ 全匹配 | `recruit_types` value 是 `job_type` 值 | B |
| `adapters/jobAdapter.ts` | `job.business_group` → 后端 `department` | ✅ 返回 `department` | 前端类型 `Job` 有 `business_group` 但实际后端 JSON key 是 `department` | B（运行时通过 JSON 映射工作） |
| `adapters/jobAdapter.ts` | `job.locations` → 后端 `city` | ✅ 返回 `city` | 前端类型声明 `locations: string[]` 但实际是单值 `city`，adapter 用 `formatListText` 兜底 | B |
| `adapters/jobAdapter.ts` | `job.tags` → 后端 `skill_keywords` | ✅ 返回 `skill_keywords` (JSON array) | 前端类型 `tags: string[]`，后端 `skill_keywords` 是 JSON，运行时反序列化 | B |
| `adapters/jobAdapter.ts` | `job.type_label` → 后端 `recruit_type`（job_type 值） | ✅ 返回 job_type | 前端 type_label 用了后端 recruit_type（=job_type），再用 TYPE_LABELS map | B |
| `components/JobFilterSidebar.tsx` | facets `directions/cities/recruit_types/categories` + `count` | ✅ 全部来自 `GET /api/jobs/facets` | count 真实 | ✅ |
| `components/JobCard.tsx` | `statusLabel` prop（可选） | ✅ 后端返回 `status` | 需 adapter 做中文映射 | B |
| `components/JobDetailBanner.tsx` | `title/typeLabel/categoryLabel/businessGroup/locationsText/bannerSummary` | ✅ 全由 adapter 从 `JobDetail` 映射 | 稳定 | ✅ |

**核心发现：前端 TypeScript 类型 `Job`（`types/job.ts`）与后端实际返回 JSON 结构存在 4 处不匹配，但运行时通过 Go JSON tag → JSON key → TypeScript 隐式类型断言工作。建议后续前端类型与后端对齐（P1）。**

## 8. 当前 P0 / P1 / P2 / P3 问题清单

### 8.1 P0

**当前未发现 Jobs 模块 P0。** 首页 / 岗位列表 / 岗位详情核心功能正常，go build/vet/test 通过。

### 8.2 P1

| 编号 | 问题 | 证据 | 影响 | 设计结论 | 下一个 prompt |
| -- | -- | -- | -- | ---- | ---------- |
| P1-J1 | `job_type` 列无索引 | `init.sql` jobs 表 7 个索引不含 job_type；`buildWhereClause:202` 用 `WHERE job_type = ?`；facets L142 用 `GROUP BY job_type` | facets 和筛选查询走全表扫描 | 添加 `idx_jobs_job_type` | 是，建索引 |
| P1-J2 | 前端 `Job` 类型与后端 `JobResponse` 契约不匹配 | 前端 `types/job.ts` 声明 `{business_group, locations, type_label, tags, short_description}`，后端返回 `{department, city, job_type, skill_keywords, description}` | TypeScript 类型检查不覆盖实际 JSON 结构 | 前端类型对齐后端，或在 adapter 层显式声明 | 是，前端类型修正 |
| P1-J3 | `recruit_type` 同时作为 API 字段和 facet 名，但实际等于 `job_type` | `toJobResponse:222` 注释 `// API 层 recruit_type 映射到 job_type`；facets 也按 `job_type` GROUP BY | 字段语义混淆，"招聘类型"= "校招" 但数据库值也是 "校招"，双字段返回相同值 | 文档化并保持；前端已有 `typeLabel` 再做一次 TYPE_LABELS 映射 | 否，文档化 |

### 8.3 P2 / P3

| 编号 | 问题 | 等级 | 证据 | 影响 | 设计结论 | 下一个 prompt |
| -- | -- | -- | -- | -- | ---- | ---------- |
| P2-J1 | facets label 映射不完整 | `directionLabels` 仅 4 项映射，`recruitTypeLabels` 仅 3 项 | 新增 direction/job_type 值时 label 回退为 value | 按需扩展 map | 否，后续 |
| P2-J2 | keyword 模糊搜索全表扫描 | `LIKE %keyword%` 四字段，仅 title 有索引 | 数据量 < 1000 时无影响 | 目前可接受 | 否，后续 |
| P2-J3 | `external_source` 错误用字符串匹配 | `job_handler.go:69` `err.Error() == "invalid external_source"` | 错误消息变更静默失败 | 改为 AppError | 是，改用 AppError |
| P2-J4 | `status` 筛选参数未暴露 | handler/service 不支持 status 过滤 | 功能受限 | 按需添加 | 否，后续 |
| P3-J1 | `department` → 前端叫 `businessGroup` 语义不一致 | 后端 `department`，前端 adapter 映射为 `businessGroup` | 团队理解成本 | 保持 | 否 |
| P3-J2 | seed 注释含 "Demo" 字样 | `init.sql:96` `是否Demo精选岗位` | 不专业 | 改为 "是否精选岗位" | 否，后续 |

## 9. 需要进入下一条 prompt 的修复项

| 优先级 | 修复项 | 是否改代码 | 是否改 schema | 验收标准 |
| --- | --- | ----- | ---------- | ---- |
| P1 | 添加 `idx_jobs_job_type` 索引 | 否（仅 DDL） | 是 | `EXPLAIN` 确认 facets GROUP BY job_type 走索引 |
| P1 | 添加 `idx_jobs_status` 索引 | 否（仅 DDL） | 是 | 同上 |
| P1 | 前端 `Job` 类型对齐后端 | 是（前端 types） | 否 | TypeScript 类型与后端 JSON key 一致 |
| P2 | `external_source` 错误改用 AppError | 是（handler） | 否 | `handleJobError` 正确分发 40001 |

## 10. 验证命令结果

| 命令 | 目的 | 结果 |
| -- | -- | -- |
| `git status --short` 执行前 | 初始状态 | 仅 `项目上下文.md` 修改 |
| `go build ./cmd/server` | 后端编译 | ✅ PASS |
| `go vet ./...` | 静态分析 | ✅ 零告警 |
| `go test ./...` | 测试 | ✅ 9 packages PASS |
| `grep "job_type\|recruit_type" server/` | 字段映射搜索 | 确认 `recruit_type = job_type` |
| `grep "KEY\|INDEX" server/sql/init.sql` | 索引清查 | 确认 job_type 无索引 |
| `git status --short` 执行后 | 零业务代码修改 | 仅 docs 变更 |

## 11. 本阶段未做事项

- 未修改前端
- 未修改后端业务代码
- 未修改数据库 schema
- 未新增索引
- 未新增接口
- 未接 LLM
- 未 git commit

## 12. docs 更新清单

| 文件 | 新增 / 更新 | 目的 |
| -- | ------- | -- |
| `docs/mcp/项目上下文.md` | 更新 | 更新阶段状态 |
| `docs/mcp/10-P-4-5-Jobs模块索引字段契约与查询稳定性核验.md` | 新增 | 本阶段 MCP 记录 |
| `docs/09-开发日志/2026-06-02-244-10-P-4-5-Jobs模块核验.md` | 新增 | 开发日志 |

## 13. CodeBuddy 重启恢复摘要

10-P-4-5 完成 Jobs 模块全量 read-only audit。3 个接口（list/facets/detail）契约稳定，参数名保持一致。22 个字段中 14 个可冻结。发现 3 个 P1：job_type 无索引（影响 facets/筛选性能）、前端 Job 类型与后端 JobResponse 不匹配（4 处字段名差异但运行时通过 adapter 工作）、recruit_type 双字段重复需文档化。发现 4 个 P2：facets label 映射不全、keyword LIKE 全表扫描、external_source 错误用字符串匹配、status 筛选未暴露。下一个 prompt 建议修复 P1 索引 + 前端类型对齐。本阶段未修改代码。

## 14. 最终判断

1. **Jobs 模块是否支撑当前首页 / 岗位详情主链路？** **是。** 3 个接口正常运行，前端 10-P-3 真实化已验证通过。

2. **当前是否发现 Jobs P0？** **否。** 无阻塞主链路的缺陷。

3. **接口契约是否已经设计冻结？** **是。** 参数名、返回结构、facets 格式均可冻结，无需变更。`recruit_type = job_type` 映射已文档化。

4. **是否存在必须修复的 P1？** **是。** `job_type` 索引缺失（P1-J1）和前端类型对齐（P1-J2）建议在下一条修复。

5. **是否需要下一条修复 prompt？** **是。** 建议执行 10-P-4-6（Jobs 索引 + 类型契约修复）。

6. **是否修改代码？** **否。** 本阶段零代码修改。

7. **是否 git commit？** **否。** 本阶段仅更新 docs。

8. **本阶段是否可以通过？** **是。** 审计完整，设计冻结结论明确，修复项已列表。
