# Recommendation-Reopen-Deep-Display-Audit-J1：重新打开推荐页仍展示 Fast 的只读审计

## 审计目标

分析用户 `/recommend` 无 resumeId 参数时为何展示 Fast 而非 Deep。

---

## 1. /recommend resumeId 解析

### 结论

访问 `http://localhost:8080/recommend`（无 `?resumeId=X`）时：

| 步骤 | 代码位置 | 行为 |
|------|------|------|
| 1. 解析 URL resumeId | `line 69-112` | `parseResumeIdFromQuery(null)` → `null`（无有效 resumeId） |
| 2. 加载简历列表 | `line 117-141` | `loadResumeOptionViewModelsFromApi()` 获取所有简历 |
| 3. 选择简历 | `line 135` | `parsedUrlResumeId` 为 null → 走 else → `opts.find((o) => o.isCurrent) ?? opts[0]` |

**当前 `is_current=1` 的简历**：`resume_id=1`（名称：01-算法研究生-岗位匹配，user_id=2）

---

## 2. 数据库推荐结果状态

### resume_id=1（/recommend 实际使用）

| 项目 | 值 |
|------|------|
| latest quality（非 mock） | ID 37，llm_agent，**arr_len=1**，created_at=2026-06-07 |
| latest fast（非 mock） | **无**（全部 mock_agent） |
| quality 是否完整 | ❌ 仅 1 条（<5） |
| quality 是否 llm_agent | ✅ |
| I2 重新生成的 Deep 是否保存 | ❌ POST /recommend 结果未写入 recommendation_results |

### resume_id=8

| 项目 | 值 |
|------|------|
| latest quality（非 mock） | ID 47，llm_agent，**arr_len=4**，created_at=2026-06-12 |
| latest fast（非 mock） | **无** |
| quality 是否完整 | ❌ 仅 4 条（<5） |

---

## 3. latest API 结果

| API | resumeId=1 | resumeId=8 |
|------|:---:|:---:|
| latest quality | has=true, llm_agent, quality | has=true, llm_agent, quality |
| latest fast | has=false | has=false |
| 是否按 generation_tier 过滤 | ✅ quality 仅返回 quality；fast 仅返回 fast | ✅ 同左 |

---

## 4. 前端展示逻辑

| 项目 | 当前实现 | 问题 |
|------|------|------|
| loadLatestRecommendationViewModel | 传 tier="quality" ✅ | — |
| activeTier 初始值 | `"fast"`（line 82） | 等待 latest 异步回调后更新 |
| quality-first 策略 | ✅ 先查 quality（line 151） | — |
| quality 完整判断 | `count >= 5` → Deep；`count < 5` → 不完整 | **I2 设计**：<5 即 fallback fast |
| Deep 不完整处理 | `setDeepErrorMessage(...)` 后 fallback fast | deepErrorMessage 被 fast_done 覆盖，不可见 |
| Deep 生成成功后 | `setDeepResult + activeTier="quality"` ✅ | — |

---

## 5. 根因判断

1. **用户访问 `/recommend` → 前端解析出 `resume_id=1`**（is_current 简历）。
2. **I2 quality-first 初始化**：先调 `loadLatestRecommendationViewModel(1, "quality")`。
3. **DB `FindLatestByResumeID` 返回 ID 37**：llm_agent，但 `result_json` 仅含 1 条岗位（旧数据，2026-06-07）。
4. **I2 完整性判断 `count=1 < 5`** → 判定为"不完整"，set deepResult + deepErrorMessage，**立即 fallback 到 latest fast**。
5. **`loadLatestRecommendationViewModel(1, "fast")`** → has_result=false（所有 fast 记录均为 mock_agent，被 `NOT IN ('mock_agent','mock_quality')` 过滤）。
6. **fallback 链到底** → `generateFast(1)` → 生成并展示 Fast。
7. **结果**：页面最终展示 Fast，`activeTier="fast"`，右侧显示"快速推荐"。

**根本矛盾**：I2 将 quality `< 5` 条当作"不完整应 fallback"，但没有在 fallback 到 Fast 后保留 deepResult 和 deepErrorMessage 的可见提示。用户看到的是纯 Fast 页面，不知道有 Deep 历史也不完整。

**I2 Deep 重新生成未持久化**：POST `/api/jobs/recommend?generation_tier=quality` 返回了 5 条完整结果，但未写入 `recommendation_results` 表。latest 端点始终返回旧记录。

---

## 6. 修复方案建议

| 优先级 | 问题 | 修复位置 | 修复动作 |
|:----:|------|------|------|
| P0 | resume_id=1 quality 仅 1 条（<5），导致 quality-first 永远 fallback | 前端 + DB | ①重新生成完整 5 条 Deep 并**确保写入 recommendation_results**；②如无法写入，前端直接使用 POST 结果实时渲染 |
| P0 | quality 不完整时 fallback 到 Fast 没有任何用户可见提示 | `JobRecommendationPage.tsx` | 在 Fast 展示状态增加"历史深度推荐不完整"横幅 |
| P1 | Deep POST 成功返回 5 条，但未持久化到 recommendation_results | 后端 quality 推荐流程 | 确保 quality 成功时 `result_json` 写入包含 5 条完整`recommended_jobs`（而非单条修复记录） |
| P1 | resume_id=1 无真实 fast 历史 → 总是实时生成 | 不影响功能 | 可接受 |

---

## 7. 是否需要重新生成 Deep

| resumeId | 当前 quality items_len | 是否需要重新生成 | 原因 |
|:---:|:---:|:---:|------|
| 1 | 1 | ✅ 是 | 仅 1 条（I2 重新生成的结果未持久化） |
| 8 | 4 | ✅ 是 | 4 条（仍 <5，未达完整性阈值） |

---

## 8. 构建结果

| 项目 | 结果 |
|------|:---:|
| npm build | ✅ PASS（I2 构建状态未变） |
| go build | ✅ PASS |
| go vet | ✅ PASS |
| go test | ✅ ALL OK |

---

## 9. 下一步

- **推荐任务名**：`Recommendation-Reopen-Deep-Display-Fix-J2`
- **是否允许直接修复**：是
- **是否需要后端接口变更**：是（需要保证 quality 生成结果写入 recommendation_results 为 5 条完整数据）
- **是否需要前端展示逻辑变更**：是（Fast 状态需显示"Deep 不完整"提示；或从不完整 quality 不 fallback 并展示既有结果+重新生成引导）
- **是否需要重新生成 Deep**：是（双简历均不符合 ≥5 完整性要求）
