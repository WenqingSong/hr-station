# K3：Jobs + Resume + Agent 主路径联调

## 1. REAL_JOB_ID

55（软件开发-后台开发方向，skills=9）

## 2. REAL_RESUME_VERSION_NO

3 个版本（Arthur v1/v2/v3 via seed data）

## 3. 匹配分析链路

| 步骤 | 数据来源 | 结果 |
|------|---------|------|
| 读取 jobId | URL ?jobId=55 | ✅ |
| 加载岗位 | getJobDetail(jobId) → /api/jobs/55 | ✅ |
| 加载简历选项 | loadResumeOptionViewModelsFromApi() → /api/resumes/versions | ✅ |
| 加载版本详情 | loadResumeVersionDetail(versionNo) | ✅ |
| 调用 Agent | loadMatchAnalysisViewModel(jobId, {resume_version_no}) | ✅ |
| Agent 上下文 | 真实 jobId + 真实 resume_version_no | ✅ |

## 4. 简历优化链路

| 步骤 | 数据来源 | 结果 |
|------|---------|------|
| 读取 jobId | URL ?jobId=55 | ✅ |
| 加载岗位 | getJobDetail(jobId) | ✅ |
| 加载版本详情 | loadResumeVersionDetail(versionNo) | ✅ |
| buildTargetJob | job.title + job.description + job.requirements | ✅ |
| buildResumeText | resume.name/education/experiences/projects/skills | ✅ |
| 调用 Agent | loadResumeOptimizationViewModel({resume_text, target_job}) | ✅ |

## 5. 上下文来源

- 匹配分析：真实 jobId=55 + 真实 resume_version_no=3
- 简历优化：真实 job detail(title/desc/reqs) + 真实 resume detail(教育/经历/项目/技能)
- fallback/mock agent 可接受，但请求上下文来自真实数据

## 6. 验证

- build 1.64s ✅
- grep CLEAN（无静态 jobId/mock/错误泄漏）✅
- 未修改后端/JSON ✅

## 7. P2

Agent 走 fallback/mock（真实 LLM 后置）

## 8. 判定

**K3 pass。零前端代码修改。**
