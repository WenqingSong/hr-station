# Recommendation-Reopen-Deep-Display-Fix-J2：修复 Deep 结果持久化与重新打开默认展示 Deep

## 修复目标

基于 J1 审计结论（recommendDeepCascade 未保存到 DB → quality 历史不完整 → quality-first 永远 fallback Fast），修复 Deep 持久化与展示源。

---

## 根因

**`recommendDeepCascade` 是 quality 主线路径，但完全没有调用 `recommendationRepo.Create` 保存结果。** 只有降级路径 `recommendWithLLM` 和 `recommendWithMock` 会在故障场景下保存（且每次仅存单条修复记录）。

调用链：
```
POST /api/jobs/recommend (quality)
  → recommendCore
    → recommendDeepCascade()  ← 主线，生成 5 条推荐，❌ 不保存
    → recommendWith3Agents() ← 降级，❌ 也不保存
    → recommendWithLLM()     ← 再次降级，✅ 保存（单条）
    → recommendWithMock()    ← 最终降级，✅ 保存（单条）
```

因为主线不保存，latest quality 永远读取降级路径的单条记录，前端 I2 quality-first 检测到 `<5` → fallback Fast。

---

## 修复详情

### 1. 后端：recommendDeepCascade 添加持久化

**文件**：`server/internal/service/recommendation_service.go:433-452`

在构建完 `RecommendationResponse` 后，`return resp, nil` 之前添加持久化逻辑：

```go
if len(recommendedJobs) > 0 {
    resultJSON, _ := json.Marshal(recommendedJobs)  // 完整 5-item 数组
    recResult := &model.RecommendationResult{
        GenerationTier: "quality",
        GeneratedBy:    "llm_agent",
        Status:         "ready",
        ResultJSON:     resultJSON,
        ...
    }
    s.recommendationRepo.Create(ctx, recResult)
}
```

### 2. 前端：viewModel 适配 latest 数组格式

**文件**：`web/src/api/viewModels.ts:294-306`

latest API 返回的 `result_json` 是 `[job1, job2, ...]` 数组格式。adapter `toRecommendationPageViewModel` 期望 `{recommended_jobs: [...]}` 对象。新增包装：

```ts
const resultObj = Array.isArray(raw.result)
    ? { recommended_jobs: raw.result }
    : raw.result;
```

### 3. 前端：Deep 不完整提示不被 Fast 覆盖

**文件**：`web/src/app/pages/JobRecommendationPage.tsx`

- 新增 `deepHistoryIncomplete` + `deepHistoryIncompleteCount` 状态（独立于 `deepErrorMessage`）
- quality 不完整时设置这两个状态而非 `deepErrorMessage`
- Deep 成功时清除这两个状态
- 新增琥珀色横幅：顶部"深度推荐历史不完整" + 右侧控制台提示（Fast/Fast+Deep_done 模式下均可见）

---

## DB 持久化验证

| resumeId | 最新 quality id | result_json arr_len | generated_by | 结论 |
|:---:|:---:|:---:|------|:---:|
| 1 | 48 | **5** | llm_agent | ✅ 完整 5 条 |
| 8 | 49 | **5** | llm_agent | ✅ 完整 5 条 |

旧记录（ID 37 的 1 条、IDs 43-47 的 4 条）均保留，不做删除。

---

## Deep 重新生成结果

| resumeId | HTTP/code | generated_by | items | scores | bad text | 结论 |
|:---:|:---:|------|:---:|------|:---:|:---:|
| 1 | 200/0 | llm_agent | 5 | [88,84,83,78,80] | 0 | ✅ |
| 8 | 200/0 | llm_agent | 5 | [88,83,80,77,74] | 0 | ✅ |

---

## 前端展示逻辑

| 场景 | 修复前 | 修复后 |
|------|------|------|
| /recommend 无 resumeId | quality 1条→fallback→Fast（无提示） | quality 5条→Deep ✅ |
| latest quality 完整 (≥5) | fallback Fast（因无持久化） | 展示 Deep ✅ |
| latest quality 不完整 (<5) | fallback Fast（无提示） | 展示 Fast + 琥珀色"不完整"横幅 ✅ |
| latest quality 不存在 | 展示 Fast | 展示 Fast ✅ |
| Deep 成功后 | 展示 Deep | 展示 Deep + 清除 incomplete flag ✅ |

---

## 修改文件

| 文件 | 修改内容 |
|------|------|
| `server/internal/service/recommendation_service.go:433-452` | recommendDeepCascade 添加持久化逻辑 |
| `web/src/app/pages/JobRecommendationPage.tsx` | +2 状态变量 + quality 不完整改用持久标记 + Deep 成功清除 + 不完整横幅 UI |
| `web/src/api/viewModels.ts:298-300` | latest 结果数组→对象包装 |

---

## 构建与测试

| 项目 | 结果 |
|------|:---:|
| npm build | ✅ PASS (2.35s) |
| go build | ✅ PASS |
| go vet | ✅ PASS |
| go test | ✅ ALL OK |

---

## 剩余问题

| 等级 | 问题 | 建议 |
|:----:|------|------|
| P2 | `recommendWith3Agents` 降级路径也未保存 | 后续补充 |
