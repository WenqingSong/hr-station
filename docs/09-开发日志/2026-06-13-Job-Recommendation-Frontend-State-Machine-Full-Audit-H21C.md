# Job-Recommendation-Frontend-State-Machine-Full-Audit-H21C 完成报告

> 执行：2026-06-13 | 类型：只读审计 | 状态：根因已定位

## 1. 总体结论

| 项目 | 结果 |
|------|------|
| 是否完成审计 | ✅ |
| 是否修改代码 | ❌ |
| **当前最大问题** | 3 个：State 1 缺按钮(已修) / Deep 历史不完整误展示 / Deep 1job 毒化状态 |
| 是否纯前端 | ✅ 状态机逻辑缺陷 |
| 是否涉及后端 | ❌ API 正常 |
| 进入 H21-E 契约冻结 | ✅ |

---

## 2. 页面与文件定位

| 项目 | 文件 |
|------|------|
| 推荐页主文件 | `web/src/app/pages/JobRecommendationPage.tsx` |
| API 调用 | `web/src/api/index.ts` → `loadRecommendationViewModel` / `loadLatestRecommendationViewModel` |
| Adapter | `web/src/adapters/recommendationAdapter.ts` |
| Deep loading 组件 | `web/src/app/components/common/DeepAgentLoadingOverlay.tsx` |
| **黄色历史提示** | `JobRecommendationPage.tsx:489-502` |
| **右侧智能分析** | `JobRecommendationPage.tsx:682-848` |

---

## 3. 核心状态变量

| 变量 | 初始值 | 关键更新点 | 风险 |
|------|--------|-----------|------|
| `phase` | `"loading_latest"` | L82/159/174/181/194/226/260/288/295 | Deep 不完整时设 `deep_done` 错误 |
| `deepResult` | `null` | L171/180/277 | 1job 残存阻断 State 1 |
| `deepHistoryIncomplete` | `false` | L182=`true`, L277=`false` | **无 activeTier 门控** |
| `deepReady` | 计算 | `deepJobs.length >= 5` | 1job 时=false 阻断 State 4 |
| `activeTier` | `"fast"` | L172/192/289/290/329 | — |

---

## 4. 页面生命周期（resumeId=2）

| T | 事件 | 代码行 | 说明 |
|---|------|--------|------|
| T0 | mount | 65 | `phase="loading_latest"` |
| T1 | load options | 120 | 加载简历列表 |
| T2 | select resumeId=2 | 128 | `setSelectedOption` |
| **T3** | **load quality history** | **163** | 🔴 串行阻塞 Fast（等待 Deep history 返回） |
| T4 | quality history = 1 job | 168 | `count=1 < 5` |
| **T5** | **set incomplete** | **180-183** | `deepResult=1job`, `phase="deep_done"`, `deepHistoryIncomplete=true` |
| T6 | load fast history | 186 | 串行（在 quality .then 内） |
| T7 | fast history empty → generateFast | 200 | `pollFast()` |
| T8 | fast returns 5 jobs | 194 | `phase="fast_done"`, `activeTier="fast"` |
| **T9** | **State 1 评估** | 687 | `!deepResult` = **false** → State 1 不命中 |
| T10 | State 4 评估 | 771 | `deepReady` = **false** → 不命中 |
| **T11** | **No state matches** | — | 🔴 右侧按钮区域空白 |
| **T12** | **yellow banner** | 490 | `deepHistoryIncomplete=true` → 顶部大黄条 |

---

## 5. 黄色"深度推荐历史不完整"审计

| 项目 | 结果 |
|------|------|
| 控制变量 | `deepHistoryIncomplete` (L81) |
| 设置位置 | L182: Deep history <5 jobs |
| 是否区分 activeTier | ❌ **不区分** — 只要 `deepHistoryIncomplete=true`，无论当前显示 Fast/Deep 都展示 |
| 是否应该在 Fast 页面出现 | ❌ **不应该** — Fast 主页面不需要 Deep 历史警告 |
| 触发条件 | `deepHistoryIncomplete && (` — 无 `isDeepDisplay` / `activeTier` 门控 |
| **修复方向** | 加 `{isDeepDisplay && deepHistoryIncomplete && ...}` 条件 |

---

## 6. 右侧智能分析四状态审计

| State | 条件 | 命中 | 问题 |
|-------|------|------|------|
| State 1 | `!isDeepDisplay && !deepResult` | ❌ 不命中 | `deepResult` 被 1job 污染为 truthy |
| State 2 | `phase==="deep_loading"` | ❌ | — |
| State 3 | `isDeepDisplay` | ❌ | phase="fast_done" |
| State 4 | `!isDeepDisplay && deepResult && deepReady` | ❌ | deepReady=false (1<5) |
| **结果** | **无状态命中** | 🔴 | 右侧无按钮、无状态 |

---

## 7. Fast 首次加载慢审计

| 项目 | 结果 |
|------|------|
| Fast API resumeId=2 | 20s (后端 profile 构建) |
| 是否被 Deep history 阻塞 | ✅ **是** — `loadLatestRecommendationViewModel(quality)` 先于 Fast 串行执行 |
| resumeId=2 触发请求数 | 3+（resume options + quality history + fast history + fast generate） |
| 更可能是前端还是后端 | 后端（20s profile 构建）+ 前端串行延迟 |
| H21-D 后端应检查 | resumeId=2 的 profile 是否已缓存 / SEP 是否需首次构建 |

---

## 8. 问题清单

| 等级 | 问题 | 触发条件 | 根因 | 修复方向 |
|------|------|----------|------|---------|
| **P0** | **State 1 缺按钮** | initial/fast_done + deepResult=null | H17-D 遗漏渲染 | ✅ H21-B 已修复 |
| **P0** | **Deep 1job 历史毒化状态** | 加载到 <5 job Deep 历史 | `deepResult` 被设非 null → State 1/4 都不命中 | 清除 deepResult 当 incomplete 或新增 State |
| **P1** | **Fast 页面误显 Deep 历史不完整** | deepHistoryIncomplete=true + any phase | 无 `activeTier`/`isDeepDisplay` 门控 | 加 `{isDeepDisplay && ...}` |
| **P0** | **Deep history 串行阻塞 Fast** | 首次加载 | `loadLatest(quality)` 串行在 Fast 之前 | 并行加载或先 Fast 后 Deep history |

---

## 9. H21-E 状态机契约草案

| 状态 | 主区域 | 右侧按钮 | 顶部 warning | DeepResult 赋值 |
|------|--------|----------|-------------|----------------|
| `booting` | loading | — | — | null |
| `loading_fast` | spinner | — | — | null |
| `fast_ready_deep_not_generated` | Fast | 生成深度推荐 | ❌ | null |
| `fast_ready_deep_history_incomplete` | Fast | 生成深度推荐 | ❌ | null（或仅存储不误判） |
| `deep_loading` | Fast+overlay | 返回快速推荐 | ❌ | null |
| `deep_ready_background` | Fast | 查看深度推荐 + 重新生成 | ❌ | >=5 job ViewModel |
| `deep_displayed` | Deep | 重新生成 | ❌ | >=5 job |
| `deep_failed` | Fast | 重试 | ❌ | — |
| `deep_unavailable` | Fast | 稍后重试 | ❌ | — |

**关键原则**：
1. `deepHistoryIncomplete` 只在 `isDeepDisplay` 时展示
2. Deep 历史 <5 jobs 时不设 `deepResult`（避免污染状态判断）
3. Fast 页面永远不显示 Deep 历史警告
4. Deep history 加载不应串行阻塞 Fast

---

## 10. API 验证

| 项目 | HTTP | time | 结果 |
|------|------|------|------|
| Fast resumeId=2 | 200 | 20s | code=0, 5 items |

---

## 11. 构建

| 项目 | 结果 |
|------|------|
| vite build | ✅ 2.29s |

---

## 12. 给用户的结论

1. **三个根因互相叠加**：State 1 缺按钮(P0)+ `deepResult` 被 1job 历史毒化导致无状态命中(P0)+ 黄色 warning 无 activeTier 门控(P1)。

2. **黄色"深度推荐历史不完整"不该在 Fast 页面出现**：`deepHistoryIncomplete` 在 L490 无任何 phase/tier 门控，只要为 true 就在页面顶部渲染。

3. **1job Deep 历史毒化 `deepResult`**：L180 `setDeepResult(1job)` 后 `deepResult` 非 null，导致 State 1 (`!deepResult`) 不命中；`deepReady` 又因 <5 为 false 导致 State 4 也不命中——右侧完全空白。

4. **Fast 首次慢的主因在后端**：resumeId=2 的 Fast 推荐本身耗时 20s（profile 构建）。前端 Deep history 串行加载加剧了等待。

5. **必须先冻结状态机契约再修**：H21-E 应明确 9 个状态、deepResult 何时应 null、历史 incomplete 何时展示。然后 H21-F 落地修复。
