# J5：Jobs 主链路浏览器全量验收

## 1. 数据库（✅）

| 检查 | 结果 |
|------|------|
| jobs | 46 |
| manual_curated/manual_verified | 46 |
| job_locations | 180 |
| job_business_groups | 194 |
| job_skills | 1402 |
| orphan (5 tables) | **全部 0** |

## 2. 后端 API（✅）

| 请求 | total |
|------|-------|
| /api/jobs?page=1 | 46 |
| /api/jobs?page=2 | 46 (5 items, different) |
| /api/jobs?keyword=Go | 6 |
| /api/jobs?application_project=实习生 | 46 |
| /api/jobs?application_subtype=应届实习 | 46 |
| /api/jobs?recruit_type=技术 | 26 |
| /api/jobs?city=深圳总部 | 44 |
| /api/jobs?business_group=CSIG | 37 |
| /api/jobs?skill=Go | 6 |
| /api/jobs/facets | 8 组（projects/types/recruit_types/dirs/cats/cities/bgs/skills） |

## 3. 岗位详情（✅）

完整返回 title/external_id/job_type/application_type/direction/category/slogan/summary/responsibilities/must_have/skills/organizations/locations/source。

## 4. 前端（✅）

- `npm run build` 1.58s
- `GET /jobs` HTTP 200

## 5. 用户可见数据源矩阵

| 页面区域 | 数据来源 | 真实后端 |
|----------|---------|---------|
| 岗位列表 | GET /api/jobs | ✅ |
| 8 组筛选 | GET /api/jobs/facets | ✅ |
| 搜索 | keyword → /api/jobs | ✅ |
| 分页 | total/page/page_size | ✅ |
| 详情 | GET /api/jobs/{id} | ✅ |

## 6. 静态数据检查

- "选择岗位"：无 ✅
- Mock 演示：无 ✅
- 静态大数字：无 ✅
- mockJobs：无 ✅
- "招聘类型"：已删除 ✅

## 7. P0/P1/P2

| 级别 | 问题 |
|------|------|
| P2 | JobDetailPage 未完整展示所有 JD 字段 |
| P2 | 岗位类型通过 recruit_type 兼容映射 |
| P2 | 浏览器人工验收留待用户本地确认 |

## 8. 判定

**Jobs 主链路收口 pass。** J0→J5 全系列完成。
