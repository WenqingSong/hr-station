# K8-5-22-B — 后端 Resume V2 Request / Response / Service

> 日期：2026-06-04 | 性质：implementation | 修改 Go 后端 Resume 链路

---

## 1. 本阶段目标

基于 K8-5-22-A 的数据库字段（schema_version + 7 摘要列 + 7 模块 JSON 列），完成后端 Resume 写入/读取链路支持 V2 字段。

**不改前端、不改 Agent、不改 SQL。**

---

## 2. 当前状态确认

K8-5-22-B 是后端 V2 支持阶段，基于 K8-5-22-A 的数据库字段。
- K8-5-22-A 已在 `resume_versions` 表新增 16 个 V2 列
- Go 源码尚未修改任何 V2 相关代码
- 本阶段目标：让后端 Resume 模型/请求/响应/服务/仓储支持 V2

---

## 3. 修改方案

| 区域 | 修改内容 | 原因 |
|------|---------|------|
| **Model** | `ResumeVersion` 新增 16 个 V2 字段 | 映射 DB 新增列 |
| **Service** | `CreateResumeRequest` 新增 12 个 V2 字段 | 支持 V2 payload |
| **Service** | `ResumeVersionResponse` 新增 16 个 V2 字段 | 返回 V2 数据 |
| **Service** | `VersionListItem` 新增 5 个 V2 摘要字段 | 列表返回 V2 摘要 |
| **Service** | `CreateVersion` 新增 V1/V2 路由 + backfill 逻辑 | schema_version=2 时写 V2 字段并回填 V1 |
| **Service** | 新增 `extractStringFromJSON` / `extractStringArrayFromJSON` 辅助函数 | V2 JSON 解析回填 |
| **Repository** | 4 个 SQL 查询全部新增 V2 列 SELECT/INSERT/SCAN | 数据库读写不漏字段 |

---

## 4. 修改清单

| 文件 | 修改内容 | 是否属于 K8-5-22-B |
|------|---------|:---:|
| `server/internal/model/resume.go` | ResumeVersion 新增 16 个 V2 字段 | ✅ |
| `server/internal/service/resume_service.go` | CreateResumeRequest / ResumeVersionResponse / VersionListItem 扩展 V2 字段；CreateVersion 新增 V1/V2 路由 + backfillV1FromV2；新增 extractStringFromJSON / extractStringArrayFromJSON 辅助函数 | ✅ |
| `server/internal/repository/resume_repository.go` | CreateVersion / FindByID / GetCurrentByUserID / ListVersionsByUserID / GetByUserIDAndVersionNo 全部新增 V2 列 SELECT/INSERT/SCAN | ✅ |
| `server/internal/http/resume_handler.go` | 未修改（handler 仅透传 request/response 类型，V2 字段自动支持） | ❌ 无需修改 |
| 前端源码 | 未修改 | ❌ |
| Agent/ Prompt / Schema | 未修改 | ❌ |
| SQL | 未修改 | ❌ |

---

## 5. 新增后端字段

### ResumeVersion model 新增

| 字段 | Go 类型 | DB 列 | 用途 |
|------|---------|------|------|
| SchemaVersion | int | schema_version | V1=1, V2=2 |
| TargetDirection | string | target_direction | 目标方向 |
| JobType | string | job_type | 岗位类型 |
| ExpectedCities | string | expected_cities | 期望城市 |
| School | string | school | 学校 |
| Degree | string | degree | 学历 |
| Major | string | major | 专业 |
| GraduationYear | string | graduation_year | 毕业时间 |
| ProfileJSON | []byte | profile_json | 基础信息 |
| JobIntentionJSON | []byte | job_intention_json | 求职意向 |
| InternshipJSON | []byte | internship_json | 实习经历 |
| CertificatesJSON | []byte | certificates_json | 证书与语言 |
| AwardsJSON | []byte | awards_json | 获奖与竞赛 |
| PortfolioJSON | []byte | portfolio_json | 作品链接 |
| ExtraJSON | []byte | extra_json | 补充信息 |

### CreateResumeRequest V2 新增

| JSON 字段 | Go 字段 | 用途 |
|-----------|---------|------|
| schema_version | SchemaVersion | V1/V2 标识 |
| resume_name | ResumeName | V2 简历名称 |
| target_direction | TargetDirection | V2 目标方向 |
| description | Description | V2 描述 |
| profile | Profile | V2 基础信息 JSON |
| job_intention | JobIntention | V2 求职意向 JSON |
| educations | Educations | V2 教育经历 JSON |
| internships | Internships | V2 实习经历 JSON |
| certificates | Certificates | V2 证书 JSON |
| awards | Awards | V2 获奖 JSON |
| portfolio_links | PortfolioLinks | V2 作品链接 JSON |
| extra_info | ExtraInfo | V2 补充信息 JSON |

---

## 6. V1 兼容策略

| V1 请求字段 | 处理 |
|------------|------|
| schema_version 缺失 | 默认 schemaVersion=1，走 V1 逻辑 |
| name, phone, email, location, target_city, target_role | 保持不变 |
| education, experience, project, skills | normalizeJSONArray 兜底空数组 |
| self_evaluation, raw_text | 保持不变 |

### V2 写入时 V1 回填策略

| V1 字段 | V2 来源 | 方式 |
|---------|---------|------|
| Name | ResumeName | 直接赋值 |
| TargetRole | TargetDirection / job_intention.target_roles[0] | 直接赋值 / JSON 解析 |
| TargetCity | job_intention.expected_cities[0] | JSON 解析 |
| Location | profile.current_city | JSON 解析 |
| Phone | profile.phone | JSON 解析 |
| Email | profile.email | JSON 解析 |
| SelfEvaluation | extra_info.self_evaluation | JSON 解析 |
| EducationJSON | educations | 直接赋值 |
| ExperienceJSON | internships | 直接赋值 |

---

## 7. Repository insert/select/scan 修改情况

| 方法 | 修改 |
|------|------|
| CreateVersion | INSERT 列新增 16 个 V2 列，VALUES 新增 16 个参数 |
| FindByID | SELECT/SCAN 新增 16 个 V2 列 |
| GetCurrentByUserID | SELECT/SCAN 新增 16 个 V2 列 |
| ListVersionsByUserID | SELECT/SCAN 新增 16 个 V2 列（含 rows scan） |
| GetByUserIDAndVersionNo | SELECT/SCAN 新增 16 个 V2 列 |

所有 nullable JSON 字段通过 `sql.NullString` → `fromNullJSON` 安全处理。

---

## 8. Response 返回字段

### ResumeVersionResponse V2 新增

| JSON 字段 | 来源 | 说明 |
|-----------|------|------|
| schema_version | r.SchemaVersion | 始终返回 |
| target_direction | r.TargetDirection | 始终返回 |
| job_type | r.JobType | 始终返回 |
| expected_cities | r.ExpectedCities | 始终返回 |
| school | r.School | 始终返回 |
| degree | r.Degree | 始终返回 |
| major | r.Major | 始终返回 |
| graduation_year | r.GraduationYear | 始终返回 |
| profile | r.ProfileJSON | normalizeResponseJSON 兜底 [] |
| job_intention | r.JobIntentionJSON | normalizeResponseJSON 兜底 [] |
| internships | r.InternshipJSON | normalizeResponseJSON 兜底 [] |
| certificates | r.CertificatesJSON | normalizeResponseJSON 兜底 [] |
| awards | r.AwardsJSON | normalizeResponseJSON 兜底 [] |
| portfolio_links | r.PortfolioJSON | normalizeResponseJSON 兜底 [] |
| extra_info | r.ExtraJSON | normalizeResponseJSON 兜底 [] |

### VersionListItem V2 新增

| JSON 字段 | 来源 |
|-----------|------|
| schema_version | v.SchemaVersion |
| target_direction | v.TargetDirection |
| expected_cities | v.ExpectedCities |
| school | v.School |
| degree | v.Degree |

---

## 9. 构建 / API / 测试结果

| 验收项 | 结果 |
|--------|------|
| `go build ./...` | ✅ 通过 |
| `go test ./...` | ✅ 全部通过 |
| `npm run build` | ✅ 通过 (1.91s) |
| `GET /api/jobs?page=1&page_size=5` | ✅ 200 |
| `GET /api/jobs/10001` | ✅ 200 |
| `GET /api/jobs/facets` | ✅ 200 |
| `GET /api/resumes` | ✅ 401（后端保护正常） |
| V2 字段在 Go 源码中存在 | ✅ 34+ 处 |
| V1 旧字段保留 | ✅ 全部保留 |
| SQL 字段未误改 | ✅ 确认存在 |

---

## 10. 未修改范围确认

```text
未修改前端源码
未修改 TypeScript types
未修改 MyResumesPage
未修改 ResumeManagementPage
未修改 Adapter
未修改 Agent
未修改 Prompt / Schema
未修改 Mock
未修改数据库 SQL
未修改 .env
未安装依赖
未迁移目录
未新增 API 路径
未新增 /api/resumes/v2
未删除 V1 字段
未实现 NormalizeResumeToSnapshotV2
```

---

## 11. 风险与边界

```text
本阶段只完成后端 V2 Request / Response / Service 基础支持。
NormalizeResumeToSnapshotV2 留给 K8-5-22-C。
前端 V2 types / adapter / page 留给后续阶段。
Agent / Prompt / Mock 留给 K8-5-22-I。
```

---

## 12. Git 约束确认

```text
本次未执行任何 Git 命令，也未提交代码。
```
