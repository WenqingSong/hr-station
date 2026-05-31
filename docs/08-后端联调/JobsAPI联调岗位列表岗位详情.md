# Jobs API 联调：岗位列表 / 岗位详情

> 阶段：6-1 | 状态：已完成 | 日期：2026-05-31

## 1. 本阶段目标

打通 Jobs 普通业务 API 联调，确认前后端通路可用。

## 2. 后端启动方式

```bash
make up
```

启动后 backend 端口 8080，前端经过 Nginx 代理。

## 3. 登录 / token 获取

```
POST /api/auth/register
Body: { "email": "...", "password": "123456", "name": "..." }
→ { code: 0, data: { token: "eyJ...", user: {...} } }
```

JWT token 用于所有业务请求的 `Authorization: Bearer <token>` header。

## 4. Jobs API curl 联调结果

| 接口 | HTTP | code | message | data |
|------|:--:|:--:|------|------|
| `GET /api/jobs` | **200** | 0 | ok | `{ items: [...] }` — 13 个岗位 |
| `GET /api/jobs/1` | **200** | 0 | ok | `{ id, title, department, city, ... }` |

## 5. 前端 jobs API 路径对齐

前端 `web/src/api/jobs.ts`：
- `getJobs()` → `GET /jobs` → 拼接后 `GET /api/jobs` ✅
- `getJobDetail(id)` → `GET /jobs/{id}` → `GET /api/jobs/{id}` ✅

API_BASE_URL 默认 `/api`，路径正确对齐。

## 6. Authorization header 支持

`web/src/api/client.ts` 已支持 `options.headers`，可传入 `Authorization: Bearer <token>`。

阶段 6-1 未实现自动 token 注入（留到用户态任务）。

## 7. Job 字段映射

| 后端字段 | 前端类型 | 状态 |
|------|------|:--:|
| id | jobId | ✅ |
| title | title | ✅ |
| department | — | ⚠️ 后端有，前端 mock 无 |
| city | location | ⚠️ 需 adapter 映射 |
| category | — | ⚠️ 同上 |
| job_type | — | ⚠️ 同上 |
| skill_keywords | skills | ✅ JSON array |
| description | description | ✅ |
| requirements | requirements | ✅ |

后端响应结构 `{ items: [...] }` vs 前端 `Job[]` — 需 adapter 处理 items 数组。

## 8. mock fallback

✅ `withFallback` 机制保留，后端不可用时自动降级 mock 数据。

## 9. build 结果

```
✓ built in 1.63s
```

**通过 ✅**

## 10. 越界检查

✅ 零后端/server/docker/env 修改，零 axios/api-v1/localStorage。

## 11. 是否可以进入 6-2

**可以。** Jobs API 联调成功，前后端通路验证通过。

**下一步：6-2 Resumes API 联调。**
