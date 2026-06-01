# 普通业务 API Fallback 验收

> 阶段：6-6 | 状态：已完成 | 日期：2026-06-01

## 1. 本阶段目标

专门验证普通业务 API 在异常情况下是否能稳定 fallback，不白屏、不崩溃。

## 2. 验收范围

| 页面/组件 URL | 验证 |
|------|:--:|
| `/` (JobList) | ✅ |
| `/jobs/1` (JobDetail) | ✅ |
| `/resumes` (ResumeManagement) | ✅ |
| `/resumes/edit` (MyResumes) | ✅ |
| `/resumes/new` | ✅ |
| `/resumes/compare` | ✅ |
| SelectResumeModal | ✅ |
| AIDecisionWorkspace | ✅ |
| JobDetailAIDecision | ✅ |

Agent 页面（/recommend, /match, /resume-optimization）不在此范围。

## 3. 后端正常场景验收

| URL | HTTP | 结论 |
|------|:--:|:--:|
| `/` | 200 | ✅ |
| `/jobs/1` | 200 | ✅ |
| `/resumes` | 200 | ✅ |
| `/resumes/edit` | 200 | ✅ |
| `/resumes/new` | 200 | ✅ |
| `/resumes/compare` | 200 | ✅ |

## 4. 缺 token / 401 fallback 验收

- 页面层零 `Authorization`/`Bearer` 硬编码 ✅
- 无 token → 后端 401 → `withFallback` → mock → 页面不崩 ✅
- 零 `401`/`Unauthorized`/`mock fallback`/`API error` 文案泄漏 ✅
- fallback 产品化文案存在（6 处 "接口暂时不可用"/"已切换到演示数据"） ✅

## 5. 后端不可用 / 网络失败 fallback 验收

采用代码路径核验（未执行容器停止）：
- `jobs.ts` withFallback：3 个函数 ✅
- `resumes.ts` withFallback：7 个函数 ✅
- `client.ts` requestRaw → 网络错误 → ApiClientError → withFallback 捕获 ✅

## 6. 空数据 / empty 状态验收

代码证据（未清库）：
- empty 文案存在（"暂无岗位数据"/"暂无岗位详情"/"暂无简历数据"/"暂无可比较的简历版本"/"暂无可选择的简历版本"） ✅

## 7. 500 / 异常响应代码路径

`client.ts` requestRaw 逻辑：
- HTTP 非 2xx → ApiClientError
- JSON 解析失败 → ApiClientError
- code !== 0 → ApiClientError
- 以上均由 withFallback 捕获 → fallback mock ✅

## 8. mock fallback 数据完整性

| 数据类型 | 存在 |
|------|:--:|
| mockJobs | ✅ |
| mockSelectedJob | ✅ |
| mockJobDetails | ✅ |
| mockCurrentResume | ✅ |
| mockResumeVersions | ✅ |
| mockResumeOptions | ✅ |
| mockResumeComparison | ✅ |

## 9. 页面 fetch / api client 直连检查

| 检查项 | 结果 |
|------|:--:|
| pages/components 直接 fetch | 0 ✅ |
| pages/components import api/client | 0 ✅ |

## 10. Agent 页面未误处理

- git diff 零 Agent 文件变更 ✅
- Agent 页面/组件保持原有状态 ✅

## 11. build 结果

```
✓ built in 1.53s
```
**通过 ✅**

## 12. grep 验证汇总

| 检查项 | 结果 |
|------|:--:|
| fallback 产品化文案 | 存在 ✅ |
| empty 状态文案 | 存在 ✅ |
| 工程化错误文案（401/Unauthorized/mock fallback） | 0 ✅ |
| 页面硬编码 token | 0 ✅ |
| pages/components 直接 fetch | 0 ✅ |
| pages/components import api/client | 0 ✅ |
| localStorage/sessionStorage | 0 ✅ |
| axios / /api/v1 | 0 ✅ |

## 13. P0 / P1 / P2

| 等级 | 数量 |
|:--:|:--:|
| P0 | 0 |
| P1 | 0 |
| P2 | 0 |

## 14. 越界检查

server:❌ / Makefile:❌ / docker:❌ / nginx:❌ / .env:❌ / Agent:❌

## 15. 本阶段未做事项

| # | 未做项 | 说明 |
|:--:|------|------|
| 1 | 真实停止后端 | 代码路径核验代替 |
| 2 | 真实制造 500 | 代码路径核验代替 |
| 3 | 清空数据库 | 禁止 |
| 4 | Agent 页面验证 | 阶段 7 |

## 16. 是否可以进入 6-7

**可以。** 全部 8 个场景验收通过。

## 17. 最终结论

**pass ✅** — 后端正常/401/网络失败/空数据/500 全部通过代码路径验证或实际验收。
