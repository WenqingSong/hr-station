# Jobs / Resumes API 模块与 Mock fallback

> 阶段：5-2.2 | 状态：已完成 | 日期：2026-05-31

## 1. 本阶段目标

新增 Jobs / Resumes 业务 API 模块，封装真实后端路径 + mock fallback，为后续页面接入提供统一入口。

## 2. 新增 API 文件

| 文件 | 行数 | 说明 |
|------|:--:|------|
| `web/src/api/jobs.ts` | 47 | 岗位 API（getJobs / getJobDetail / getSelectedJob） |
| `web/src/api/resumes.ts` | 128 | 简历 API（P0 读取 5 个 + P1 写入 3 个） |
| `web/src/api/index.ts` | 5 | 追加 jobs/resumes 导出 |

**合计：3 文件（2 新 + 1 改），180 行新增**

## 3. Jobs API 函数

| 函数 | 后端路径 | fallback |
|------|------|------|
| `getJobs()` | `GET /jobs` | `mockJobs` |
| `getJobDetail(jobId)` | `GET /jobs/{jobId}` | `getMockJobById(jobId) ?? mockSelectedJob` |
| `getSelectedJob()` | `GET /jobs/{id}` | `mockSelectedJob` |

## 4. Resumes API 函数

### P0 读取

| 函数 | 后端路径 | fallback |
|------|------|------|
| `getResumes()` | `GET /resumes/versions` | `mockResumeVersions` |
| `getCurrentResume()` | `GET /resumes/current` | `mockCurrentResume` |
| `getResumeDetail(id)` | `GET /resumes/detail?id=` | `getMockResumeById() ?? mockCurrentResume` |
| `getResumeOptions()` | —（直接 mock） | `mockResumeOptions` |
| `compareResumes(baseId, targetId)` | `GET /resumes/versions/compare` | `mockResumeComparison` |

### P1 写入（Demo 占位）

| 函数 | 后端路径 | fallback |
|------|------|------|
| `createResume(input)` | `POST /resumes` | `mockCurrentResume` |
| `updateResume(id, input)` | `PUT /resumes/detail?id=` | `getMockResumeById() ?? mockCurrentResume` |
| `setCurrentResume(id)` | `PATCH /resumes/current` | `getMockResumeById() ?? mockCurrentResume` |

## 5. Mock fallback 策略

所有函数使用统一 `withFallback()` helper：

```ts
try { return await requestFn(); } catch { return fallbackFn(); }
```

- 网络错误 / HTTP 非 2xx / code!==0 → 自动 fallback
- 不打印错误日志，不弹 toast
- fallback 始终返回可用数据

## 6. 本阶段未做事项

- 未写 Agents API（留到 5-2.3）
- 未写 ViewModel loader（留到 5-2.4）
- 未改页面/组件
- 未改后端
- 未改 types/mocks/adapters
- 未引入 axios
- 未使用 /api/v1

## 7. 后续 5-2.3 输入

可基于现有 `client.get/post()` + `withFallback` 模式构建三类 Agent API。
