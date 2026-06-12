# Recommendation-Entry-And-Deep-State-Audit-L1：岗位页入口到推荐页 Deep 展示链路只读审计

## 审计目标

用户从 `/jobs` 右侧栏点击"推荐适合的岗位"进入 `/recommend`，期望展示已保存的 Deep 结果，实际看到"暂无推荐结果"。

---

## 1. /jobs 入口审计

| 项目 | 当前实现 | 问题 | 建议 |
|------|------|------|------|
| 按钮位置 | `AIDecisionWorkspace.tsx:117` — 岗位列表页右侧 AI 面板 | — | — |
| 跳转目标 | `navigate("/recommend")` | **不带 resumeId** | `/recommend?resumeId=currentId` |
| 是否带 resumeId | ❌ 否 | `/recommend` 依赖 `is_current` 简历，可能选取错误简历 | 显式传 resumeId |
| 是否触发生成 | ❌ 否，仅导航 | — | — |

**结论**：`/jobs` → `/recommend` 跳转不带 resumeId。`/recommend` 页面回退到 `is_current` 简历（query: `opts.find(o => o.isCurrent) ?? opts[0]`，line 135）。当前 `is_current` 简历为 `resume_id=1`。

---

## 2. /recommend 初始化审计

| 项目 | 当前实现 | 评价 |
|------|------|:---:|
| resumeId 选择 | URL 无参数 → `is_current` 或 `opts[0]` | ✅ |
| latest quality 调用 | `loadLatestRecommendationViewModel(rid, "quality")` | ✅ J2 实现 |
| latest fast 调用 | quality 不完整时 fallback；quality 无→fallback | ✅ |
| loading 状态 | "loading_latest" phase → `PageLoading` | ✅ |
| empty 状态 | line 355: `!activeResult \|\| jobs.length === 0` → "暂无推荐结果" | ⚠️ 在异步完成前可能闪过 |
| activeResult 计算 | `phase==="deep_done" && deepResult ? deepResult : fastResult` | ✅ |

### 当前数据验证

| 项目 | resumeId=1 | resumeId=8 |
|------|:---:|:---:|
| DB latest quality id | 53（5 items） | 49（5 items） |
| DB quality items_len | 5 | 5 |
| API latest quality | has=true, list, 5 items | has=true, list, 5 items |
| API result structure | `[{job1},...,{job5}]` | `[{job1},...,{job5}]` |
| latest fast | has=false | has=false |

**结论**：DB 和 API 均返回完整 5 条 quality 结果。

---

## 3. 前端 adapter/viewModel 验证

| 项目 | 当前实现 | 评价 |
|------|------|:---:|
| data.result array | J2 viewModels.ts:300 `Array.isArray(raw.result) ? {recommended_jobs: raw.result}` | ✅ 正确包装 |
| recommended_jobs 提取 | `toRecommendationPageViewModel` → `Array.isArray(result.recommended_jobs)` | ✅ |
| 完整性判断 | `count >= 5` → Deep；`< 5` → 不完整 fallback | ✅ |

**结论**：adapter 路径正确，5 条数据应被正确解析。

---

## 4. 空状态根因判断

### 触发条件

**文件**：`web/src/app/pages/JobRecommendationPage.tsx:355`

```tsx
if (!activeResult || jobs.length === 0) {
  return <PageEmpty title="暂无推荐结果" .../>
}
```

**activeResult 计算**（line 288）：
```tsx
const activeResult = phase === "deep_done" && deepResult ? deepResult : fastResult;
```

### 为何可能进入空状态

对于当前正确的数据（DB 有 5 条 quality），quality-first 流程应为：
1. `loadLatestRecommendationViewModel(1, "quality")` → `count=5 >= 5`
2. → `setDeepResult(vm)` + `setPhase("deep_done")` + `setActiveTier("quality")`
3. → `activeResult = deepResult` （5 jobs）
4. → 跳过空状态 → 展示 Deep ✅

如果用户看到"暂无推荐结果"，最可能原因：
- **时序问题**：`selectedOption` 异步加载期间，React 渲染中间状态时 `deepResult` 和 `fastResult` 均为 null，导致 `activeResult = null`
- **phase 未更新**：quality-first init 的 useEffect 未执行（`selectedOption` 未触发 effect 重跑）
- **loading 阶段未完全阻塞**：理论上 `loadingOptions` 和 `loading_latest` guard 应遮挡

### 当前验证

在 J2 修复后的环境下（DB 有 5 条 quality，后端已重启），`/recommend` 初始化应正确展示 Deep。用户截图问题可能是 J2 修复前的状态，或前端构建未刷新。

---

## 5. 构建结果

| 项目 | 结果 |
|------|:---:|
| npm build | ✅ PASS（1.97s） |
| go build | ✅ PASS |
| go vet | ✅ PASS |

---

## 6. 修复方案建议

| 优先级 | 问题 | 修复位置 | 修复动作 |
|:----:|------|------|------|
| P1 | `/jobs` 按钮跳转不带 resumeId | `AIDecisionWorkspace.tsx:117` | `navigate("/recommend")` → `navigate("/recommend?resumeId=${currentResumeId}")` |
| P1 | 空状态可能在 quality-first 异步完成前闪过 | `JobRecommendationPage.tsx:355` | 增加 `isLatestLoading` guard |
| P2 | deepResult 已有但 phase 非 deep_done 时误判 | `JobRecommendationPage.tsx:288` | `activeResult` 计算增加 fallback：`deepResult ?? fastResult` |

---

## 7. 下一步

- **推荐任务名**：`Recommendation-Entry-And-Deep-State-Fix-L2`
- **是否允许直接修复**：是
- **是否需要前端修改**：是（导航 URL + 状态 guard）
- **是否需要后端修改**：否
- **是否需要重新生成 Deep**：否（数据已完整）
