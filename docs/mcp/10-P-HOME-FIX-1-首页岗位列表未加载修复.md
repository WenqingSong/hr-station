# 10-P-HOME-FIX-1 首页岗位列表未加载修复

## 1. 阶段信息

- **阶段编号**：10-P-HOME-FIX-1
- **日期**：2026-06-02
- **范围**：首页岗位列表公开访问 + handler 内鉴权移除
- **是否修改业务代码**：是
- **修改模式**：最小修复（仅 1 个文件，+3 / -19 行）

## 2. 本次结论摘要

- **根因**：`job_handler.go` 的 `List`/`Detail`/`Facets` 三个 handler 内部调用了 `middleware.UserIDFromContext(r.Context())`，对 context 中没有 user_id 的请求返回 401。由于 `router.go` 对这些路由未套 auth middleware，未登录用户的请求 context 中无 user_id，被 handler 内部鉴权拒绝。
- **修复**：移除三个公开 handler 中的 `UserIDFromContext` 校验（3 处各 4 行），移除未使用的 `middleware` import。
- **路由层已正确**：`router.go` 中 `/api/jobs`、`/api/jobs/facets` 已是公开路由（`HandleFunc`，无 auth middleware），`/api/jobs/recommend` 和 `/api/jobs/{id}/match` 保持受保护。
- **验证通过**：全部 11 项验收标准通过。

## 3. 执行前状态

| 检查项 | 结果 |
|--------|------|
| git status | 干净 |
| docker compose ps | 3 个服务 healthy (mysql, redis, nginx) |
| healthz | 200 OK |
| curl localhost:5173/ | 200 OK (Vite dev server) |
| `npm run build` | ✅ |
| `npm run typecheck` | ✅ |
| `go build ./cmd/server` | ✅ |
| `go test ./...` | ✅ (10 packages PASS) |

## 4. 问题复现结果

```bash
# 不带 token请求公开岗位列表 → 401
curl -i http://localhost:8080/api/jobs
# HTTP/1.1 401 Unauthorized
# {"code":40100,"message":"unauthorized","data":null}

# 不带 token请求 facets → 401
curl -i http://localhost:8080/api/jobs/facets
# HTTP/1.1 401 Unauthorized

# 不带 token请求岗位详情 → 401
curl -i http://localhost:8080/api/jobs/1
# HTTP/1.1 401 Unauthorized
```

前端页面现象：
- 筛选区域显示"筛选加载失败"
- 岗位列表显示"岗位数据加载失败，请稍后重试"
- 用户处于未登录状态

## 5. 根因分析

### 路由层面（已正确）

`router.go` 第 69-70 行已将 jobs list/facets 设为公开路由：

```go
mux.HandleFunc("/api/jobs", deps.JobHandler.List)           // public
mux.HandleFunc("/api/jobs/facets", deps.JobHandler.Facets)  // public
```

未套 `deps.AuthMiddleware`，路由层不会拒绝未登录请求。

### Handler 层面（根因）

`job_handler.go` 的三个 handler 内部仍有 `UserIDFromContext` 校验：

```go
// List handler (line 31-33)
if _, ok := middleware.UserIDFromContext(r.Context()); !ok {
    response.Error(w, apperror.ErrUnauthorized.HTTPStatus, ...)
    return
}
```

由于路由层没有 auth middleware → context 中没有注入 user_id → `UserIDFromContext` 返回 false → 返回 401。

**三层鉴权叠加**：
1. 路由层（router.go）：✅ 已是公开的（未套 auth middleware）
2. Handler 层（job_handler.go）：❌ 内部 self-check user_id → 401
3. Service/Repository 层：不依赖 user_id（查询公开岗位数据）

**根因结论**：路由层暴露正确，但 handler 内部的遗留鉴权检查阻断了未登录请求。

## 6. 路由公开 / 私有边界设计

| 路径 | 公开/私有 | 鉴权方式 |
|------|-----------|----------|
| GET /api/jobs | 公开 | 无 |
| GET /api/jobs/facets | 公开 | 无 |
| GET /api/jobs/{id} | 公开 | 无（wildcard dispatch 中的 public 分支） |
| POST /api/jobs/recommend | 私有 | `deps.AuthMiddleware` |
| POST /api/jobs/{id}/match | 私有 | `dispatchJobWildcard` 内 inline auth |

## 7. 后端修复内容

### `server/internal/http/job_handler.go`（+3 / -19）

三处修改：

**1. List handler** — 移除 `UserIDFromContext` 校验（第 31-34 行）：

```go
// 移除：
if _, ok := middleware.UserIDFromContext(r.Context()); !ok {
    response.Error(w, apperror.ErrUnauthorized.HTTPStatus, ...)
    return
}
```

**2. Detail handler** — 同上（第 83-86 行）

**3. Facets handler** — 同上（第 117-120 行）

**4. 移除未使用的 import**：

```go
// 移除：
"offer-catcher/server/internal/middleware"
```

## 8. 前端修复内容

**无修改。** 前端 `JobListPage.tsx`、`JobList.tsx`、`JobFilterSidebar.tsx` 代码逻辑正确：

- `client.ts`：无 token 时不发 Authorization header，后端公开接口正常返回
- `JobList.tsx`：`getJobs()` → `GET /api/jobs` → 200 返回岗位列表
- `JobFilterSidebar.tsx`：`getJobFacets()` → `GET /api/jobs/facets` → 200 返回筛选项
- `AIDecisionWorkspace.tsx`：`isLoggedIn` 区分三种状态，未登录显示"登录后使用 AI 求职决策"

## 9. 数据库 / seed 核验

```
id    title               city   category    job_type  status
1     Go 后端开发工程师      深圳    后端开发      校招      active
2     云计算研发工程师        深圳    云计算        校招      active
3     AI 应用开发工程师      北京    AI应用       校招      active
4     数据开发工程师         上海    数据开发      校招      active
5     测试开发工程师         广州    测试开发      校招      active
```

jobs 表 8 条数据，全部 active，facets 所需字段均有值。非 seed 缺失导致。

## 10. API 验收结果

| 接口 | 预期 | 实际 | 状态 |
|------|------|------|------|
| GET /api/jobs (无 token) | 200, code=0, 8 jobs | 200, code=0, total=8 | ✅ |
| GET /api/jobs/facets (无 token) | 200, code=0 | 200, code=0, 8 categories | ✅ |
| GET /api/jobs/1 (无 token) | 200, code=0 | 200, code=0, title="Go 后端开发工程师" | ✅ |
| POST /api/jobs/recommend (无 token) | 401 | 401 | ✅ |
| POST /api/jobs/1/match (无 token) | 401 | 401 | ✅ |
| POST /api/jobs/recommend (有 token) | 200 | 200, summary 正常 | ✅ |
| POST /api/auth/login | 200 | 200, token 返回 | ✅ |
| GET /api/auth/me (无 token) | 401 | 401 | ✅ |

## 11. 浏览器验收情况

当前环境无法完成浏览器视觉验收，只完成 curl/API/代码路径验收。

已验证：
- 后端所有 8 个测试接口行为正确
- 前端构建 + typecheck 通过
- 路由公开/私有边界清晰

## 12. Mock / 演示反弹检查

- `web/src` 中无 "Mock 演示"/"演示数据" 残留 ✅
- 无 `from ".*/mocks"` 引用 ✅
- 无 `withFallback`/`withMockFallback`/`fallbackToMock` ✅

## 13. 验证结果

| 检查项 | 结果 |
|--------|------|
| `go build ./cmd/server` | ✅ |
| `go test ./...` | ✅ (10 packages PASS) |
| `npm run build` | ✅ |
| `npm run typecheck` | ✅ |
| AuthMiddleware 仅用于受保护路由 | ✅ |
| UserIDFromContext 仅用于受保护 handler | ✅ (job_handler.go 中已移除) |

## 14. 当前 P0 / P1 / P2 / P3

| 级别 | 问题 | 状态 |
|------|------|------|
| P0 | 首页岗位列表 401（handler 内鉴权） | ✅ 已修复 |
| P0 | 首页筛选加载失败 | ✅ 已修复（同根因） |
| P0 | 岗位详情 401 | ✅ 已修复（同根因） |
| — | 无其他新增问题 | — |

## 15. 修改文件清单

| 文件 | 修改 | 行数 |
|------|------|------|
| `server/internal/http/job_handler.go` | 移除 3 处 UserIDFromContext 校验 + import | +3 / -19 |

## 16. 本阶段未做事项

- 未修改前端代码
- 未修改路由代码（已正确）
- 未修改数据库
- 未接真实 LLM
- 未做浏览器视觉验收
- 未做版本控制写入

## 17. docs 更新清单

| 文件 | 操作 |
|------|------|
| `docs/mcp/10-P-HOME-FIX-1-首页岗位列表未加载修复.md` | 新增 |
| `docs/mcp/项目上下文.md` | 待更新 |

## 18. CodeBuddy 重启恢复摘要

- 当前阶段：10-P-HOME-FIX-1（首页岗位列表未加载修复）
- 根因：`job_handler.go` handler 内部 `UserIDFromContext` 鉴权拒绝了未登录请求
- 修复：1 个文件，移除 3 处暗含鉴权
- 公开/私有边界：`router.go` 路由层已正确，handler 层现也正确
- 构建/测试：全部通过

## 19. 最终判断

**建议通过本阶段。**

根因明确（handler 内鉴权与路由层公开定义冲突），修复最小（1 个文件，仅删除冗余校验），验证全面（8 个 API 接口全部行为正确），无副作用（build/test/typecheck 全绿）。
