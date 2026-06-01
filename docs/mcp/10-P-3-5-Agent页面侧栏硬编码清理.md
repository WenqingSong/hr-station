# 10-P-3-5 Agent 页面侧栏硬编码清理

## 1. 阶段信息

* 阶段：10-P-3-5
* 日期：2026-06-02
* 范围：Agent 相关页面侧栏硬编码、演示标签、JD 关键词、影响因素、默认值清理
* 是否修改前端：是
* 是否修改后端：否
* 是否继续 LLM：否
* 是否继续阶段 11/12/13：否

## 2. 本次结论摘要

* Agent 页面"演示数据"/"当前为演示数据"/"Mock 演示"等用户可见标签全部清理，替换为真实状态文案（"分析完成"/"优化完成"）或移除
* 匹配分析页和简历优化页侧栏硬编码 JD 关键词数组（8 个假关键词 `Go/MySQL/Redis/微服务/高并发/Linux/数据库/接口开发`）已移除，改为空态提示
* 匹配分析页和简历优化页侧栏硬编码 `"后台开发"`/`"校园招聘"` Badge 已移除，改为条件渲染真实数据
* 简历优化侧栏右侧 4 条假"本次优化预计影响"列表已移除
* 岗位详情 AI 决策区"已展示演示数据"错误文案、"后台开发"Badge、"Mock 演示"标签均已清理
* 岗位详情 Banner "已展示演示数据"错误文案、"校园招聘"Badge 均已清理
* SelectResumeModal 默认 hardcoded `jobTitle = "软件开发-后台开发方向"` 已修复为 `jobTitle = ""`（P1 遗留清零）
* RecommendationDetail 默认 `currentResumeId = "resume-v2"` 已修复；硬编码 navigate("/jobs/1") 改为真实 jobId
* RecommendationHeader 默认 `resumeName = "V2 后端开发简历"` 已修复为空字符串
* mock/fallback 未反弹（withFallback 零匹配，组件 import mock 零匹配）
* npm run build 1.88s 通过，go build 通过
* 未修改后端，未修改页面级（pages）文件

## 3. 页面 / 组件清查矩阵

| 页面 / 组件 | 用户可见业务数据 | 修改前来源 | 修改后来源 | 是否仍有硬编码 | 证据文件 |
| ------- | -------- | ----- | ----- | ------- | ---- |
| JobRecommendationPage | 推荐结果 | API → props | 不变 | ❌ 无硬编码 | — |
| JobMatchAnalysisPage | 匹配结果 | API → props | 不变 | ❌ 无硬编码 | — |
| ResumeOptimizationPage | 优化结果 | API → props | 不变 | ❌ 无硬编码 | — |
| JobDetailPage | 组合布局 | — | 不变 | ❌ 无硬编码 | — |
| MatchAnalysisSidebarLeft | JD 关键词 + 标签 | 8 个假关键词 + "后台开发"/"校园招聘" Badge | 空态提示 + 条件渲染真实标签 | ✅ 已清零 | L27-28, L54-58 |
| ResumeOptimizationSidebarLeft | JD 关键词 + 标签 | 同上 | 同上 | ✅ 已清零 | L27-28, L54-58 |
| MatchAnalysisHeader | 分析状态标签 | "演示数据" 硬编码 | "分析完成" | ✅ 已清零 | L57 |
| ResumeOptimizationHeader | 优化状态标签 | "演示数据" 硬编码 | "优化完成" | ✅ 已清零 | L69 |
| MatchAnalysisSidebarRight | 演示数据提示 | "当前为演示数据" 标签 | 已移除 | ✅ 已清零 | L73-78 |
| ResumeOptimizationSidebarRight | 影响因素 + 演示提示 | 4 条假影响因素 + "当前为演示数据" | 优化说明文案 + 移除演示标签 | ✅ 已清零 | L67-84, L87-92 |
| JobDetailAIDecision | 错误文案 + 标签 + Mock 标签 | "已展示演示数据" / "后台开发" / "Mock 演示" | 真实错误文案 + 条件渲染 | ✅ 已清零 | L50, L64, L127-130 |
| JobDetailBanner | 错误文案 + 标签 | "已展示演示数据" / "校园招聘" | 真实错误文案 + 条件渲染 | ✅ 已清零 | L29, L50 |
| SelectResumeModal | 默认 jobTitle | `"软件开发-后台开发方向"` | `""`（空字符串） | ✅ 已清零 | L45 |
| RecommendationDetail | 默认 resumeId + 导航 jobId | `"resume-v2"` + `/jobs/1` | 空字符串 + 真实 firstJob.jobId | ✅ 已清零 | L13, L112 |
| RecommendationHeader | 默认 resumeName | `"V2 后端开发简历"` | 空字符串 | ✅ 已清零 | L10 |

## 4. 移除 / 替换的硬编码业务数据

| 文件 | 原硬编码内容 | 替换方式 | 当前状态 |
| -- | ------ | ---- | ---- |
| `MatchAnalysisSidebarLeft.tsx` | `"后台开发"` Badge + `"校园招聘"` Badge | 条件渲染 jobVm.typeLabel/jobVm.categoryLabel | ✅ 已移除 |
| `MatchAnalysisSidebarLeft.tsx` | `["Go","MySQL","Redis","微服务","高并发","Linux","数据库","接口开发"]` 8 个假 JD 关键词 | 空态提示"关键词提取需要真实 JD 匹配分析结果，当前未获取到。" | ✅ 已移除 |
| `ResumeOptimizationSidebarLeft.tsx` | `"后台开发"` Badge + `"校园招聘"` Badge | 条件渲染真实标签 | ✅ 已移除 |
| `ResumeOptimizationSidebarLeft.tsx` | `["Go","MySQL","Redis","微服务","高并发","Linux","数据库","接口开发"]` 8 个假 JD 关键词 | 空态提示 | ✅ 已移除 |
| `MatchAnalysisHeader.tsx` | `"演示数据"` 状态标签 | `"分析完成"` | ✅ 已替换 |
| `ResumeOptimizationHeader.tsx` | `"演示数据"` 状态标签 | `"优化完成"` | ✅ 已替换 |
| `MatchAnalysisSidebarRight.tsx` | `"当前为演示数据"` Sparkles 标签 | 整个标签区移除 | ✅ 已移除 |
| `ResumeOptimizationSidebarRight.tsx` | 4 条假影响因素（"后台开发匹配度提升"/"JD 关键词覆盖更完整"/"项目经历表达更贴近岗位要求"/"投递前风险降低"） | 替换为优化说明文案 | ✅ 已移除 |
| `ResumeOptimizationSidebarRight.tsx` | `"当前为演示数据"` Sparkles 标签 | 整个标签区移除 | ✅ 已移除 |
| `JobDetailAIDecision.tsx` | `"简历接口暂时不可用，已展示演示数据。"` 错误文案 | `"简历数据暂未加载，请先创建或选择一份简历。"` | ✅ 已替换 |
| `JobDetailAIDecision.tsx` | `"后台开发"` Badge | 条件渲染真实 jobVm.typeLabel/categoryLabel | ✅ 已移除 |
| `JobDetailAIDecision.tsx` | `"当前结果来源：Mock 演示"` 底部标签 | 整个标签区移除 | ✅ 已移除 |
| `JobDetailBanner.tsx` | `"岗位接口暂时不可用，已展示演示数据。"` 错误文案 | `"岗位数据加载失败，请稍后重试。"` | ✅ 已替换 |
| `JobDetailBanner.tsx` | `"校园招聘"` Badge | 条件渲染真实 typeLabel/categoryLabel | ✅ 已移除 |
| `SelectResumeModal.tsx` | `jobTitle = "软件开发-后台开发方向"` 默认值 | `jobTitle = ""`（空字符串，条件渲染隐藏） | ✅ 已修复 |
| `RecommendationDetail.tsx` | `currentResumeId = "resume-v2"` 默认值 | `currentResumeId = ""` | ✅ 已修复 |
| `RecommendationDetail.tsx` | `navigate("/jobs/1")` 硬编码导航 | `navigate(\`/jobs/${firstJob.jobId}\`)` | ✅ 已修复 |
| `RecommendationHeader.tsx` | `resumeName = "V2 后端开发简历"` 默认值 | `resumeName = ""` | ✅ 已修复 |

## 5. 空态 / 错误态说明

| 场景 | 当前展示 | 是否会显示 mock |
| -- | ---- | ---------- |
| JD 关键词无数据 | "关键词提取需要真实 JD 匹配分析结果，当前未获取到。" | 否 |
| 未选择简历页面 | 匹配/优化页调用方在父页面处理空态 | 否 |
| 推荐结果为空 | "暂无推荐结果，请先选择简历并发起推荐。" | 否 |
| 岗位 Banner 加载失败 | "岗位数据加载失败，请稍后重试。" | 否 |
| JobDetail 简历未加载 | "简历数据暂未加载，请先创建或选择一份简历。" | 否 |
| 接口失败（鉴权 401） | API client 统一抛出错误，无 fallback | 否 |

## 6. API / 数据来源说明

| 模块 | 当前数据来源 | 是否真实后端 | 说明 |
| -- | ------ | ------ | -- |
| 推荐 | `POST /jobs/recommend` | ✅ 是（后端 Mock Agent） | Mock Agent 基于真实简历/岗位数据做规则推荐 |
| 匹配分析 | `POST /jobs/{id}/match` | ✅ 是（后端 Mock Agent） | 同上，规则匹配 |
| 简历优化 | `POST /resumes/optimize` | ✅ 是（后端 Mock Agent） | 同上，规则优化 |
| 岗位详情 AI 入口 | `GET /jobs/{id}` + `GET /resumes/current` | ✅ 是 | 岗位/简历数据来自真实 API |
| 选择简历弹窗 | `GET /api/resumes/versions` | ✅ 是 | 简历版本列表来自真实 API |

## 7. 修改内容清单

| 文件 | 修改内容 | 目的 | 是否业务代码 |
| -- | ---- | -- | ------ |
| `web/src/app/components/MatchAnalysisSidebarLeft.tsx` | 移除硬编码 `"后台开发"`/`"校园招聘"` Badge；移除假 JD 关键词数组，改为空态 | 清理侧栏假数据 | 是 |
| `web/src/app/components/ResumeOptimizationSidebarLeft.tsx` | 同上 | 清理侧栏假数据 | 是 |
| `web/src/app/components/MatchAnalysisHeader.tsx` | `"演示数据"` → `"分析完成"` | 清理演示标签 | 是 |
| `web/src/app/components/ResumeOptimizationHeader.tsx` | `"演示数据"` → `"优化完成"` | 清理演示标签 | 是 |
| `web/src/app/components/MatchAnalysisSidebarRight.tsx` | 移除 `"当前为演示数据"` 标签区；移除未使用的 `Sparkles` import | 清理演示标签 | 是 |
| `web/src/app/components/ResumeOptimizationSidebarRight.tsx` | 移除 4 条假影响因素 + `"当前为演示数据"` 标签；替换为优化说明文案；移除未使用的 `Sparkles` import | 清理假影响因素 + 演示标签 | 是 |
| `web/src/app/components/JobDetailAIDecision.tsx` | 替换 `"已展示演示数据"` 错误文案；移除 `"后台开发"` Badge；移除 `"Mock 演示"` 标签 | 清理演示标签和假 Badge | 是 |
| `web/src/app/components/JobDetailBanner.tsx` | 替换 `"已展示演示数据"` 错误文案；移除 `"校园招聘"` Badge | 清理演示标签和假 Badge | 是 |
| `web/src/app/components/SelectResumeModal.tsx` | `jobTitle` 默认值从 `"软件开发-后台开发方向"` 改为 `""` | 修复 P1 遗留 | 是 |
| `web/src/app/components/RecommendationDetail.tsx` | `currentResumeId` 默认从 `"resume-v2"` 改为 `""`；navigate 从 `"/jobs/1"` 改为真实 `firstJob.jobId` | 清理硬编码默认值 | 是 |
| `web/src/app/components/RecommendationHeader.tsx` | `resumeName` 默认从 `"V2 后端开发简历"` 改为 `""` | 清理硬编码默认值 | 是 |

## 8. 未修复 / 保留问题

| 问题 | 风险等级 | 未修原因 | 后续处理方向 |
| -- | ---- | ---- | ------ |
| 后端 Mock Agent 仍使用规则引擎（非 LLM） | P1 | 阶段 11 暂停，不可接真实 LLM | 待阶段 11 恢复 |
| `MyResumesSidebar.tsx` "当前为演示数据" | P2 | 该组件为死代码（无引用） | 后续清理阶段删除 |
| `MyResumesDetail.tsx` 硬编码关键词 | P2 | 同上（死代码） | 同上 |
| `MyResumesList.tsx` 硬编码关键词 | P2 | 同上（死代码） | 同上 |
| `AIBadge.tsx` 默认文本 "Mock 演示" | P2 | 该组件无活跃引用 | 后续清理 |
| typecheck/lint 缺失 | P1 | 已封存于 10-P-3-0 | 后续添加 `tsc --noEmit` |

## 9. 验证结果

| 命令 / 检查项 | 目的 | 结果 |
| -------- | -- | -- |
| `git status --short` 执行前 | 初始状态 | 干净（零输出） |
| `npm run build` | 前端编译验证 | ✅ 1.88s, 1676 modules, 零错误 |
| `go build ./cmd/server` | 后端编译验证 | ✅ PASS，零错误 |
| `git status --short` 执行后 | 修改范围 | 11 文件（均在 web/src/app/components/） |
| `grep -rn "演示数据\|当前为演示\|Mock 演示\|已展示演示\|当前结果来源" web/src/app/components/ | grep -v "MyResumes\|AIBadge"` | 演示标签检查 | 零匹配（仅死代码文件残留） |
| `grep -rn '"Go".*"MySQL".*"Redis".*"微服务"\|"后台开发"\|"校园招聘"' web/src/app/components/ | grep -v "MyResumes"` | 假关键词/Badge 检查 | 零匹配 |
| `grep -rn 'jobTitle.*=.*"软件开发-后台开发方向"' web/src/app/ | grep -v docs` | SelectResumeModal 默认值 | 零匹配 |
| `grep -rn '"V2 后端开发简历"\|"resume-v2"' web/src/app/components/ | grep -v "MyResumes"` | 硬编码名称检查 | 零匹配 |
| `grep -rn '"后台开发匹配度提升"\|"投递前风险降低"' web/src/app/ | grep -v docs` | 假影响因素检查 | 零匹配 |
| `grep -rn "withFallback\|withMockFallback\|fallbackToMock" web/src/app/ | grep -v docs` | mock/fallback 反弹检查 | 零匹配 |
| `grep -rn "from.*mocks" web/src/app/components/ | grep -v "MyResumes\|StyleGuide"` | component mock import 检查 | 零匹配 |

## 10. 未做事项

* 未修改后端
* 未接真实 LLM
* 未修改 Jobs / Resumes 后端
* 未修改首页 / 简历管理页 / 简历版本对比页主流程
* 未修改页面级文件（JobRecommendationPage / JobMatchAnalysisPage / ResumeOptimizationPage 等）
* 未删除 MyResumes* 死代码
* 未继续阶段 11/12/13
* 未做上线部署
* 未私自规划后续 prompt
* 未 git commit

## 11. 当前风险与不确定项

* **后端 Mock Agent 仍存在** — 匹配分析/推荐/优化接口返回的是 Mock Agent 规则引擎结果，非 LLM。但数据基于真实简历/岗位差异计算，前端不编造结果，不符合"假数据"定义
* **MyResumes* 死代码仍有硬编码** — `MyResumesSidebar.tsx`（"当前为演示数据"）、`MyResumesDetail.tsx`（假关键词）、`MyResumesList.tsx`（假关键词）但均为零引用死代码，不影响用户可见页面
* **AIBadge.tsx 默认文本 "Mock 演示" 未改** — 该组件无活跃引用，不影响用户可见页面。后续清理阶段统一处理
* **typecheck/lint 缺失** — P1 遗留，已封存于 10-P-3-0
* **接口需 JWT 鉴权** — 未登录时 401 由 API client 统一处理，无 fallback，正确暴露错误
* **mock/fallback 未反弹** — grep 确认后零匹配

## 12. docs 更新清单

| 文件 | 新增 / 更新 | 目的 |
| -- | ------- | -- |
| `docs/mcp/项目上下文.md` | 更新 | 更新当前阶段为 10-P-3-5 |
| `docs/mcp/10-P-3-5-Agent页面侧栏硬编码清理.md` | 新增 | 本阶段 MCP 记录（即本文件） |
| `docs/09-开发日志/2026-06-02-236-10-P-3-5-Agent页面侧栏硬编码清理.md` | 新增 | 开发日志 |

## 13. CodeBuddy 重启恢复摘要

10-P-3-5 完成 Agent 页面侧栏硬编码清理。11 个组件文件共修复 20 处硬编码/演示标签：MatchAnalysisSidebarLeft 和 ResumeOptimizationSidebarLeft 的假 JD 关键词（8 个）和假 Badge（"后台开发"/"校园招聘"）；MatchAnalysisHeader 和 ResumeOptimizationHeader 的"演示数据"状态标签；MatchAnalysisSidebarRight 和 ResumeOptimizationSidebarRight 的"当前为演示数据"提示及假影响因素；JobDetailAIDecision 的"已展示演示数据"/"Mock 演示"及假 Badge；JobDetailBanner 的"已展示演示数据"/"校园招聘"；SelectResumeModal 的硬编码 jobTitle（P1 遗留已清零）；RecommendationDetail 和 RecommendationHeader 的硬编码默认值。所有修改均在组件层，未触及页面级文件。npm build 1.88s 通过，mock/fallback 未反弹。剩余 P1/P2：后端 Mock Agent、MyResumes* 死代码、typecheck 缺失。阶段 11/12/13 仍暂停。

下次继续前必须读取：`docs/mcp/项目上下文.md`、本文件、`docs/mcp/10-P-3-0-前端低级和严重问题清查.md`。

## 14. 最终判断

1. **Agent 页面用户可见演示标签是否已清理？** **是。** 所有活跃组件中的"演示数据"/"Mock 演示"/"当前为演示数据"/"已展示演示数据"均已清理或替换。

2. **Agent 侧栏硬编码 JD 关键词 / 影响因素是否已清理？** **是。** 假 JD 关键词数组（2 处各 8 个）和假影响因素（4 条）均已移除。

3. **SelectResumeModal 默认 jobTitle 是否已修复？** **是。** 默认值从 `"软件开发-后台开发方向"` 改为 `""`，P1 遗留清零。

4. **API fallback / mock import 是否反弹？** **否。** grep 确认零匹配。

5. **是否修改后端？** **否。**

6. **是否接入真实 LLM？** **否。** 阶段 11 暂停。

7. **是否可以进入 typecheck / lint 与 as any 收敛阶段？** **是。** P1-F7（typecheck/lint 缺失）为下一个可处理的工程化问题。

8. **本阶段是否可以通过？** **是。** 全部验收标准满足。
