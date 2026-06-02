# K3：Jobs + Resume + Agent 主路径联调

## 1. REAL_JOB_ID

`55`（软件开发-后台开发方向，skills=9）

## 2. REAL_RESUME_VERSION_NO

**REAL_RESUME_VERSION_NO = 3**

可用的简历版本：v1 / v2 / v3（Arthur 的 3 个版本，由 seed data 提供）。

本次验收以 **v3** 作为代表版本，真实版本号 `resume_version_no=3`，通过 `GET /api/resumes/versions/3` 获取真实简历详情。

## 3. K3 原 partial_pass 原因

K3 在技术验收完成后标记为 `partial_pass`，原因：

1. 未新增 K3 开发日志；
2. `docs/说明书.md` 未明确更新 K3 状态；
3. `REAL_RESUME_VERSION_NO` 表述不精确，只写了"3 个版本（v1/v2/v3）"，没有明确本次验收使用的具体版本号；
4. Agent 仍走 fallback/mock agent，需明确 LLM 后置。

## 4. K3-D 补齐内容

本阶段（K3-D）为 docs-only 补齐：

1. 补齐 K3 开发日志；
2. 明确 `REAL_RESUME_VERSION_NO = 3`；
3. 在 `docs/说明书.md` 新增 K3 专节；
4. 更新 `docs/项目状态.md` / `docs/下一步任务.md` / `docs/变更记录.md`；
5. 未修改任何业务代码。

## 5. 匹配分析链路

```
/jobs/55
→ /match?jobId=55
→ JobMatchAnalysisPage 通过 useSearchParams 读取 jobId=55
→ getJobDetail(55) 加载真实岗位
→ loadResumeOptionViewModelsFromApi() 加载真实简历版本
→ loadResumeVersionDetail(3) 加载真实简历详情
→ loadMatchAnalysisViewModel(jobId=55, resume_version_no=3)
→ Match Agent 请求上下文来自真实岗位 + 真实简历版本
```

| 步骤 | 数据来源 | 结果 |
|------|---------|------|
| 读取 jobId | URL `?jobId=55` | ✅ |
| 加载岗位 | `getJobDetail(jobId)` → `/api/jobs/55` | ✅ |
| 加载简历选项 | `loadResumeOptionViewModelsFromApi()` → `/api/resumes/versions` | ✅ |
| 加载版本详情 | `loadResumeVersionDetail(versionNo)` → 真实 v3 详情 | ✅ |
| 调用 Agent | `loadMatchAnalysisViewModel(jobId, {resume_version_no})` | ✅ |
| Agent 上下文 | 真实 jobId=55 + 真实 resume_version_no=3 | ✅ |

**确认：**
- 没有静态 jobId；
- 没有写死 resume_version_no；
- 没有 mockSelectedJob / mockJobDetails 主路径；
- Agent 可以走 fallback/mock agent，但上下文来自真实数据。

## 6. 简历优化链路

```
/jobs/55
→ /resume-optimization?jobId=55
→ ResumeOptimizationPage 通过 useSearchParams 读取 jobId=55
→ getJobDetail(55) 加载真实岗位
→ loadResumeVersionDetail(3) 加载真实简历详情
→ buildTargetJob(jobDetail) 生成真实 target_job
→ buildResumeText(resumeDetail) 生成真实 resume_text
→ loadResumeOptimizationViewModel({ resume_text, target_job })
→ Optimization Agent 请求上下文来自真实岗位 + 真实简历
```

| 步骤 | 数据来源 | 结果 |
|------|---------|------|
| 读取 jobId | URL `?jobId=55` | ✅ |
| 加载岗位 | `getJobDetail(jobId)` → `/api/jobs/55` | ✅ |
| 加载版本详情 | `loadResumeVersionDetail(versionNo)` → 真实 v3 详情 | ✅ |
| buildTargetJob | `job.title` + `job.responsibilities` + `job.must_have` | ✅ |
| buildResumeText | `resume.name/education/experiences/projects/skills` | ✅ |
| 调用 Agent | `loadResumeOptimizationViewModel({resume_text, target_job})` | ✅ |

**确认：**
- `resume_text` 不来自 mock；
- `target_job` 不来自 mock；
- 页面显示的岗位标题 / 优化目标来自真实 job detail；
- 真实 LLM 未启用，LLM 后置。

## 7. Agent 请求上下文来源

| Agent | 上下文字段 | 来源 |
|-------|-----------|------|
| 匹配分析 | `jobId=55` + `resume_version_no=3` | 真实岗位 API + 真实简历版本 API |
| 简历优化 | `resume_text` + `target_job` | 真实简历 v3 详情 + 真实岗位详情 |

Agent 当前走 fallback/mock agent，但请求上下文**全部来自真实数据**，不存在 mock 上下文。

## 8. fallback/mock agent 使用情况

| 项 | 说明 |
|----|------|
| Agent 执行路径 | fallback/mock agent（真实 LLM 后置） |
| 请求上下文 | 真实（job detail + resume detail） |
| 是否存在 mock 上下文 | 否 |
| 真实 LLM 接入 | 未在本阶段处理，后置于 K4 及之后 |
| Mock 数据污染 | 无 — 前端代码无 mock import，所有数据从 API 获取 |

## 9. 验证

- `npm run build` 1.64s ✅
- grep CLEAN（无静态 jobId / mock / 错误泄漏）✅
- 未修改前端 ✅
- 未修改后端 ✅
- 未修改 resources JSON ✅

## 10. P0 / P1 / P2

| 级别 | 问题 | 状态 |
|------|------|------|
| — | P0 = 0 | ✅ |
| — | P1 = 0 | ✅ |
| P2 | Agent 走 fallback/mock（真实 LLM 后置） | 后置于 K4 |
| P2 | 浏览器人工点击验收未执行 | 后置于 K4 |
| P2 | Agent 结果质量后续再优化 | 后置 |

## 11. 判定

**K3 pass。** 零前端代码修改。零后端代码修改。K3 从 `partial_pass` 修正为 `pass`。

Jobs + Resume + Agent 主路径已基于真实岗位和真实简历上下文贯通。

## 12. 下一阶段

**K4：Agent 结果页浏览器端到端验收与交互优化。**

目标：
1. 浏览器打开 `/match?jobId=55`，选择真实简历版本，触发匹配分析，验证结果页可读；
2. 浏览器打开 `/resume-optimization?jobId=55`，选择真实简历版本，触发简历优化，验证结果页可读；
3. 真实 LLM 仍后置。
