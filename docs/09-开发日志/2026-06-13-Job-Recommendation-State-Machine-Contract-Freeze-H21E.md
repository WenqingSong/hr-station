# Job-Recommendation-State-Machine-Contract-Freeze-H21E 完成报告

> 执行：2026-06-13 | 类型：设计冻结 | 状态：✅

## 1. 总体结论

| 项目 | 结果 |
|------|------|
| 是否完成契约冻结 | ✅ |
| 是否修改业务代码 | ❌ |
| 冻结 deepResult 合法定义 | ✅ >=5 jobs 才可 set |
| 冻结 9 状态 | ✅ |
| 冻结 H21-F 修复边界 | ✅ 仅 `JobRecommendationPage.tsx` |
| 进入 H21-F | ✅ |

---

## 2. H21-C 问题 → 契约解决

| 问题 | 根因（代码行） | 契约解决 |
|------|---------------|----------|
| State 1 缺按钮 | L686-704 漏渲染 | 9 状态明确要求 `fast_ready_*` 必须有 CTA 按钮 |
| Deep 1job 毒化 deepResult | **L180** `setDeepResult(1job)` | **deepResult 仅完整 Deep(>=5)可 set** |
| Fast 误显 Deep history warning | **L490** 无门控 | 加 `activeTier === "deep"` 或 `isDeepDisplay` |
| Fast 串行阻塞 | L163 quality-first | 并行：Fast 优先展示，Deep history 后台加载 |

---

## 3. deepResult 合法定义

**完整 Deep**：`recommendedJobs.length >= 5`，`generated_by` 非 mock。

**不完整 Deep (1-4 jobs)**：
- ❌ 不能 `setDeepResult(incomplete)`
- ✅ 可设 `deepHistoryIncomplete=true` + `deepHistoryCount=N`
- ✅ `deepResult` 保持 `null`

**原因**：`deepResult` 非 null 会污染 State 1 条件 `!deepResult`，且 `deepReady`=<5 又阻断 State 4 → 右侧空白。

---

## 4. deepHistoryIncomplete 显示规则

| 条件 | 允许显示 |
|------|---------|
| `activeTier === "deep"` 或 `isDeepDisplay` | ✅ 顶部黄色 banner |
| `activeTier === "fast"` | ❌ 顶部大黄条 ❌ |
| `activeTier === "fast"` + `deepHistoryIncomplete=true` | ✅ 右侧弱提示（State 3） |

---

## 5. 9 状态契约

| # | 状态 | 主区域 | 右侧按钮 | 顶部 warning | deepResult |
|---|------|--------|----------|-------------|-----------|
| 0 | `booting` | 骨架屏 | — | — | null |
| 1 | `loading_fast` | spinner | — | — | null |
| 2 | `fast_ready_deep_not_generated` | Fast | **生成深度推荐** | ❌ | null |
| 3 | `fast_ready_deep_history_incomplete` | Fast | **重新生成完整深度推荐** | ❌ (仅右侧弱提示) | null |
| 4 | `deep_loading` | Fast+overlay | 返回快速推荐 | ❌ | null |
| 5 | `deep_ready_background` | Fast | 查看深度推荐 + 重新生成 | ❌ | >=5 ViewModel |
| 6 | `deep_displayed` | Deep | 重新生成深度推荐 | 仅 Deep 自身有问题时 | >=5 ViewModel |
| 7 | `deep_failed` | Fast | 重试深度推荐 + 返回快速推荐 | ❌ | — |
| 8 | `deep_unavailable` | Fast | 稍后重试 + 查看快速推荐 | 轻量提示 | — |

---

## 6. 状态变量契约

| 变量 | 冻结语义 | 禁止行为 |
|------|----------|---------|
| `phase` | 9 值枚举 | ❌ `<5 jobs` 时设 `deep_done` |
| `activeTier` | `"fast"` / `"deep"` | ❌ Deep 后台完成不强制切 |
| `deepResult` | 仅 `>=5` jobs | ❌ `<5` 时 set |
| `deepHistoryIncomplete` | Boolean flag | ❌ Fast 页面无门控展示 |
| `activeResult` | `activeTier` 优先 | ❌ `deepResult` 非 null 就抢占 |

---

## 7. 请求时序契约

```
页面进入 /recommend?resumeId=2：
1. loadResumeOptions()
2. 并行：
   ├── loadLatestRecommendation(quality) — 历史 meta
   └── loadLatestRecommendation(fast) || generateFast()
3. Fast 结果优先 → phase="fast_ready_*"
4. Deep 不完整 → deepHistoryIncomplete=true, deepResult=null
5. ❌ 不允许 quality 串行阻塞 Fast
```

```
生成 Deep：
1. phase="deep_loading", activeTier 保持 fast
2. overlay DeepAgentLoadingOverlay
3. 用户未返回 → Deep 完成 → phase="deep_displayed"
4. 用户返回 Fast → Deep 完成 → deepReadyInBackground=true
5. ❌ <5 jobs 结果不 setDeepResult
```

---

## 8. H21-F 最小修复边界

| 文件 | 修复 | 行数 |
|------|------|------|
| `JobRecommendationPage.tsx` | L180: `setDeepResult(1job)` → 仅记录 incomplete | 1 |
| 同上 | L490: `deepHistoryIncomplete &&` → 加 `isDeepDisplay &&` | 1 |
| 同上 | State 3 新增 `fast_ready_deep_history_incomplete` 状态块 | ~20 |
| 同上 | L163-222 请求时序优化（可选，H21-D 后做） | ~15 |

---

## 9. H21-F 验收用例

| Case | 场景 | 预期 |
|------|------|------|
| C1 | resumeId=2, Deep 历史 1 job | 顶部无黄条，右侧"重新生成完整深度推荐" |
| C2 | resumeId=3, Deep 未生成 | 右侧"生成深度推荐" |
| C3 | 点击生成 Deep | overlay 正常，右侧"返回快速推荐" |
| C4 | Deep 完成，用户未切 | 右侧"查看深度推荐" |
| C5 | 点击查看 Deep | 主区域 Deep，右侧"重新生成" |
| C6 | 刷新页面 | deepResult 不被 1job 污染 |

---

## 10. 构建

| 项目 | 结果 |
|------|------|
| vite build | ✅ 2.40s |

---

## 11. docs

- 新增：`docs/09-开发日志/2026-06-13-Job-Recommendation-State-Machine-Contract-Freeze-H21E.md`
- 更新：`docs/变更记录.md`、`docs/下一步任务.md`

---

## 12. 给用户的结论

1. **必须先冻结状态机**：3 个根因（缺按钮/毒化 deepResult/warning 无门控）共享同一个源头 — 没有明确契约约束状态转换和变量赋值。

2. **H21-F 只改 `JobRecommendationPage.tsx`**：L180 停用 `<5` 的 `setDeepResult`，L490 加 `isDeepDisplay` 门控，State 3 补 `fast_ready_deep_history_incomplete` 块。

3. **不动后端**：API 正常，问题是前端逻辑污染。Fast 20s 慢属后端 profile 构建延迟（H21-D 审计）。

4. **9 状态契约确保不再回归**：每个状态的主区域、按钮、warning、deepResult 赋值都有明确规则。
