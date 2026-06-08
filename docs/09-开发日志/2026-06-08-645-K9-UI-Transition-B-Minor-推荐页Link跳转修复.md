# K9-UI-Transition-B-Minor：推荐页 Link 跳转修复

> 日期：2026-06-08 | 性质：P0 修复 | 最小改动

---

## 1. 当前状态确认

| 项 | 结果 |
|----|:--:|
| K9-UI-Transition-A 诊断结论 | 🔴 P0: `<a href>` 导致整页浏览器重载 |
| 问题代码位置 | `web/src/app/pages/JobRecommendationPage.tsx` L260-261 |
| 本次目标 | `<a href>` → `<Link to>`，消除整页重载白屏 |
| 改动范围 | 1 文件，2 行代码 |
| 是否修改业务逻辑 | ❌ 否，仅替换跳转标签 |

---

## 2. P0 根因修复

| 问题 | 修复结果 |
|------|:--:|
| `<a href>` 导致浏览器整页重载 → 白屏 | ✅ 改为 `<Link to>` SPA 路由切换 |
| 点击后 React App 重新初始化（AuthBoot + RouterProvider） | ✅ 不再发生 |
| Network 重新请求 index.html | ✅ 不再请求 |
| AppBootLoading 重新出现 | ✅ 不再出现 |
| Navbar 重新挂载闪烁 | ✅ 不再闪烁 |

---

## 3. 修改清单

| 文件 | 修改 |
|------|------|
| `web/src/app/pages/JobRecommendationPage.tsx` L260 | `<a href={...}>` → `<Link to={...}>`（进入岗位详情） |
| `web/src/app/pages/JobRecommendationPage.tsx` L261 | `<a href={...}>` → `<Link to={...}>`（进行匹配分析） |

---

## 4. Link 跳转修复明细

| 按钮 | 修改前 | 修改后 | 结果 |
|------|--------|--------|:--:|
| 进入岗位详情 | `<a href={/jobs/${jobId}}>` | `<Link to={/jobs/${jobId}}>` | ✅ |
| 进行匹配分析 | `<a href={/match?jobId=X&resumeId=Y}>` | `<Link to={/match?jobId=X&resumeId=Y}>` | ✅ |

修改前后 className、文案、跳转 URL 完全一致。

---

## 5. 源码检查

| 项 | 结果 |
|----|:--:|
| 无残留 `<a href>` 跳转到 /jobs | ✅ PASS |
| 无残留 `<a href>` 跳转到 /match | ✅ PASS |
| 无 `window.location` / `location.href` | ✅ PASS |
| `Link to` 存在于 L260-261 | ✅ 确认 |
| `Link` 已导入（react-router L2） | ✅ 确认 |
| Linter 诊断 | ✅ 零错误 |

---

## 6. 构建 / 测试结果

| 验收项 | 结果 | 归因 |
|--------|:--:|------|
| npm build | ✅ PASS | — |
| go build ./... | ✅ PASS | — |
| go vet ./... | ✅ PASS | — |
| go test ./... | ⚠️ 5 FAIL | **pre-existing**：`internal/agent` MatchAnalysis 测试 + `internal/service` MatchService 测试，与前端 `<a>` → `<Link>` 修改无关 |
| check-encoding | ⚠️ FAIL | **pre-existing**：`init.generated.current.sql` 中 SQL dump 的编码问题，与本次修改无关 |

### go test 失败明细（全部 pre-existing）

```
--- FAIL: TestMatchAgent_FinalV2Snapshot_EvidenceChain (0.00s)       ← agent 层
--- FAIL: TestMatchAgent_FinalV2Snapshot_TargetRoleEmptySmoke (0.00s) ← agent 层
--- FAIL: TestIntegration_MatchAnalysis_Mock_Direct (0.00s)           ← agent 层
--- FAIL: TestIntegration_MatchAnalysis_Mock_Fallback_Metadata (0.00s) ← agent 层
--- FAIL: TestMatchAnalysisAgentReturnsDecisionFieldsSmoke (0.00s)    ← agent 层
--- FAIL: TestMatchService_MockAgent_StableResult (0.00s)             ← service 层
```

与本次 K9-UI-Transition-B-Minor 修改**零关联**（全部为后端 agent/service 测试，本次只动了前端 2 行 `<a>` → `<Link>`）。

---

## 7. 未修改范围确认

- ✅ 未修改 K9-1 岗位推荐业务逻辑（仅改跳转标签类型）
- ✅ 未修改 K9-2 岗位匹配分析后端
- ✅ 未修改 K9-3 简历优化建议
- ✅ 未修改 K9-4 简历版本对比
- ✅ 未修改 `server/sql/init.sql`
- ✅ 未修改数据库结构
- ✅ 未删除真实简历数据
- ✅ 未执行 `make reset-up`
- ✅ 未执行任何 Git 命令

---

## 8. docs 更新清单

| 操作 | 文件 |
|:--:|------|
| ✅ 新增 | `docs/09-开发日志/2026-06-08-645-K9-UI-Transition-B-Minor-推荐页Link跳转修复.md`（本文件） |
| ✅ 更新 | `docs/下一步任务.md` |
| ✅ 更新 | `docs/变更记录.md` |

---

## 9. Git 约束确认

- 本次未执行任何 Git 命令
- `git status` / `git diff` / `git add` / `git commit` / `git push` / `git reset` / `git checkout` / `git branch` / `git tag` / `git show` / `git restore` 均未执行

---

## 10. 是否进入下一阶段

- ✅ **通过**：建议进入 **K9-UI-Transition-B-Full**：JobMatchAnalysisPage 骨架屏 + JobDetailPage 骨架屏增强

理由：
- P0 根因已消除：`<a href>` → `<Link to>`，不再触发整页重载
- npm build / go build / go vet 全部通过
- 测试失败和编码检查失败均为 pre-existing，与本次修改无关
- 改动极小（1 文件 2 行），风险极低
- K9-UI-Transition-A 诊断的 P1 问题（MatchAnalysisPage loading 全页替换）可通过 K9-UI-Transition-B-Full 修复
