# 10-P-5-2-A 前端主路径缺口与 Navbar 用户态复验修复

## 1. 阶段信息

* 阶段：10-P-5-2-A
* 日期：2026-06-02
* 范围：岗位详情→简历优化直达入口、Navbar useLocation 当前页刷新
* 是否修改代码：是
* 是否修改前端：是
* 是否修改后端：否
* 是否修改数据库：否
* 是否接真实 LLM：否

## 2. 本次结论摘要

* 已补齐岗位详情 → 简历优化直接入口：`JobDetailAIDecision` 卡片新增"生成简历优化建议"按钮
* 已保留岗位详情 → 匹配分析入口：双按钮并存（"分析岗位匹配度" + "生成简历优化建议"）
* 已复用真实简历选择：`SelectResumeModal` 扩展 `resume-optimization` scene，通过 `handleConfirm` 根据 `modalAction` 分发跳转
* Navbar 登录态已支持当前页更新：新增 `useLocation()`，每次路由变化重新读取 token
* 未新增后端接口
* 未影响主链路：匹配分析路线完整保留
* npm build 1.59s（1676 modules）+ npm typecheck 零错误
* go build + go test（11 packages PASS）全部通过
* 建议 10-P-5-2 封存，建议进入 10-P-5-3

## 3. 执行前问题确认

| 问题 | 证据 | 用户影响 | 修复方式 |
| -- | -- | ---- | ---- |
| 岗位详情 → 简历优化缺直接入口 | `JobDetailAIDecision.tsx` 仅有 `handleConfirmMatch` → `/match`，无 optimize 导航 | 用户无法从岗位详情直达简历优化 | 新增 handleConfirm + modalAction 分发 + SelectResumeModal resume-optimization scene |
| Navbar 只依赖 storage event 存在风险 | `Navbar.tsx:16` 仅 `window.addEventListener("storage", checkAuth)` ，storage event 不在同窗口触发 | 登录成功 navigate("/") 后 Navbar 不刷新，仍显示"登录" | 新增 `useLocation()` 监听路由变化，每次变化重新 `getAuthToken()` |

## 4. 岗位详情 AI 入口修复

| 项 | 修改前 | 修改后 | 说明 |
| - | --- | --- | -- |
| 匹配分析入口 | 单按钮"开始岗位匹配分析"，点击打开 SelectResumeModal | 双按钮之一"分析岗位匹配度"，点击 `setModalAction("match")` | 保留原有功能 |
| 简历优化入口 | 不存在 | 双按钮之二"生成简历优化建议"，点击 `setModalAction("optimize")` | 本次新增 |
| SelectResumeModal 复用 | 1 个 Modal，`scene="match-analysis"`，`onConfirm={handleConfirmMatch}` | 2 个 Modal：`scene="match-analysis"` + `scene="resume-optimization"`，共享 `handleConfirm`，根据 `modalAction` 分发 | `SelectResumeModal` 新增 `resume-optimization` scene |
| 跳转 query 参数 | `/match?jobId=X&resumeId=X` | `/match?jobId=X&resumeId=X` 或 `/resume-optimization?jobId=X&resumeId=X` | 与 `ResumeOptimizationPage` 的 `useSearchParams().get("jobId")` 一致 |
| 缺简历时处理 | 显示"简历数据暂未加载" warning | 保留 | 不伪造数据 |

## 5. Navbar 用户态修复

| 场景 | 修改前 | 修改后 | 说明 |
| -- | --- | --- | -- |
| 未登录 | 显示"登录" | 显示"登录" | 不变 |
| 登录后 navigate("/") | ⚠️ 仍显示"登录"（storage event 不同窗口触发） | ✅ 显示"退出"（location 变化触发 `isLoggedIn(true)`） | `useEffect([location])` 每次路由变化重新读 token |
| 退出后 | ✅ 显示"登录"（`handleLogout` 立即 `setIsLoggedIn(false)`） | ✅ 显示"登录" | 不变 |
| 路由变化后 | ❌ 不重新读取 token | ✅ 重新读取 token | `useEffect(() => { setIsLoggedIn(!!getAuthToken()); }, [location])` |
| 跨 tab storage | ✅ 通过 `window.addEventListener("storage")` 同步 | ✅ 保留 | 不受影响 |
| 是否依赖后端 me 接口 | 否 | 否 | 仅依赖 localStorage token 存在性判断 |

## 6. 主交互路径复验

| 路径 | 是否可走 | 证据 | 结论 |
| -- | ---- | -- | -- |
| 岗位详情 → 匹配分析 | ✅ | `handleConfirm` + `modalAction === "match"` → `/match?jobId=X&resumeId=X` | ✅ 完整 |
| 岗位详情 → 简历优化 | ✅ | `handleConfirm` + `modalAction === "optimize"` → `/resume-optimization?jobId=X&resumeId=X` | ✅ 本次补齐 |
| 匹配分析 → 简历优化 | ✅ | `MatchAnalysisDetail` → `navigate(/resume-optimization?jobId=X&resumeId=X)` | ✅ 保持 |
| 简历优化 → 版本对比 | ✅ | `ResumeOptimizationDetail` "版本对比"按钮 (10-P-5-2 已修复) | ✅ 保持 |

## 7. 修改内容清单

| 文件 | 修改内容 | 是否业务代码 |
| -- | ---- | ------ |
| `web/src/app/components/Navbar.tsx` | 新增 `useLocation`；登录态判断从仅 mount + storage → 增加 location 变化触发 | 是 |
| `web/src/app/components/JobDetailAIDecision.tsx` | 新增 `modalAction` 状态；`handleConfirm` 统一分发 match/optimize；双按钮 UI；2 个 SelectResumeModal | 是 |
| `web/src/app/components/SelectResumeModal.tsx` | 新增 `resume-optimization` scene；扩展 `Scene` 类型 + `SCENE_CONFIG`；jobTitle 显示条件从 `match-analysis` → 非 recommendation | 是 |
| `docs/mcp/项目上下文.md` | 更新当前阶段为 10-P-5-2-A | 否 |
| `docs/mcp/10-P-5-2-A-前端主路径缺口与Navbar用户态复验修复.md` | 新增 | 否 |
| `docs/09-开发日志/2026-06-02-254-10-P-5-2-A-主路径与Navbar修复.md` | 新增 | 否 |

## 8. 验证结果

| 命令 | 结果 | 说明 |
| -- | -- | -- |
| `git status --short` 执行前 | 3 files M (JobDetailAIDecision / Navbar / SelectResumeModal) + docs | — |
| `npm run build` | ✅ PASS | 1.59s，1676 modules |
| `npm run typecheck` | ✅ PASS | 零错误 |
| `go build ./cmd/server` | ✅ PASS | — |
| `go test ./...` | ✅ PASS | 11 packages |
| StyleGuide / AIBadge / mocks 反弹 | ✅ 零匹配 | 无反弹 |
| Mock / 演示 / 测试数据反弹 | ✅ 零匹配 | 无反弹 |
| 内部 `href` 检查 | ✅ 零匹配 | 无新的 `<a href>` |
| resume-optimization?jobId / match?jobId / useLocation / getAuthToken | ✅ 均有引用 | 路径参数 + Navbar 刷新机制确认 |
| `git status --short` 执行后 | +3 M 文件 (JobDetailAIDecision / Navbar / SelectResumeModal)，+51 -20 lines | 3 files changed |

## 9. 未做事项

* 未修改后端代码
* 未修改数据库 schema
* 未新增接口
* 未接真实 LLM
* 未做样式系统统一
* 未做页面大重构
* 未修改业务数据源
* 未执行任何版本控制写入操作

## 10. 当前风险与后续建议

* **已解决**：岗位详情→简历优化 直接入口补齐
* **已解决**：Navbar `useLocation` 路由变化触发 token 重读
* **已解决**：`SelectResumeModal` resume-optimization scene 扩展
* **不需要处理，原因是**：匹配分析入口已原样保留，不受影响
* **不需要处理，原因是**：`generated_by`/`agent_status` 仅内部元数据，不暴露 UI
* **需要下一阶段处理，阶段名是 10-P-5-3**：前端样式系统与基础组件统一

## 11. docs 更新清单

| 文件 | 新增 / 更新 | 目的 |
| -- | ------- | -- |
| `docs/mcp/项目上下文.md` | 更新 | 更新当前阶段为 10-P-5-2-A |
| `docs/mcp/10-P-5-2-A-前端主路径缺口与Navbar用户态复验修复.md` | 新增 | 本阶段 MCP 记录 |
| `docs/09-开发日志/2026-06-02-254-10-P-5-2-A-主路径与Navbar修复.md` | 新增 | 开发日志 |

## 12. CodeBuddy 重启恢复摘要

10-P-5-2-A 修复了 10-P-5-2 验收发现的两个缺口：
- 岗位详情→简历优化：JobDetailAIDecision 新增双按钮 UI（匹配分析 + 简历优化），新增 modalAction 状态分发，SelectResumeModal 扩展 resume-optimization scene
- Navbar 登录态：新增 useLocation() 监听路由变化，每次路由变化重新 getAuthToken()，确保登录/退出后当前页立即刷新
- 10-P-5-2 可以封存

下次继续前读取：`docs/mcp/项目上下文.md`、`docs/mcp/10-P-5-2-A-前端主路径缺口与Navbar用户态复验修复.md`。

## 13. 最终判断

1. 岗位详情 → 简历优化直接入口是否已补齐？是。
2. 岗位详情 → 匹配分析入口是否保留？是。
3. 是否复用真实简历选择？是（SelectResumeModal + loadResumeOptionViewModelsFromApi）。
4. Navbar 登录态当前页更新是否可靠？是（useLocation 监听路由变化 + storage event 跨 tab 同步）。
5. 是否新增后端接口？否。
6. 是否出现 Mock / 演示反弹？否。
7. 主链路是否受影响？否（匹配分析入口原样保留，新增优化入口为补充）。
8. npm build / typecheck 是否通过？是（1.59s build, 零错误 typecheck）。
9. 后端基础验证是否通过？是（go build + go test 11 packages PASS）。
10. 本阶段是否可以通过？是。
11. 10-P-5-2 是否可以封存？是。
12. 是否建议进入 10-P-5-3？是。
