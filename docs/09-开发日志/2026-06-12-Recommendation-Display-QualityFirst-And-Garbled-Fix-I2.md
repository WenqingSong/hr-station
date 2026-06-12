# Recommendation-Display-QualityFirst-And-Garbled-Fix-I2：修复推荐页乱码 + Deep 优先展示逻辑

## 修复目标

基于 I1 审计结论，修复推荐页三个问题：
1. Fast API `match_highlights` 中的 `�` 乱码
2. 前端 cleanUserText 无 `\ufffd` 防御
3. Deep/Fast 优先展示策略 + latest API tier 支持

---

## 修复详情

### 1. 后端 UTF-8 截断乱码（P0）

**根因**：`truncateStr(s, maxLen)` 使用 `s[:maxLen]` 对字符串按字节切片，中文字符（UTF-8 3 字节）被切碎产生 `�`（U+FFFD）。

| 文件 | 行号 | 修复 |
|------|------|------|
| `server/internal/service/semantic_profile_gate.go` | 181-186 | `s[:maxLen]` → `string([]rune(s)[:maxLen])` + `maxLen<=0` guard |
| `server/internal/service/semantic_profile_profiler.go` | 312-315 | 同上 |

**验证**：
- Fast API resume_id=1: 0 `\ufffd` in all 5 match_highlights ✅
- Fast API resume_id=8: 0 `\ufffd` in all 5 match_highlights ✅

### 2. 前端 cleanUserText 兜底清洗（P1）

| 文件 | 修改 |
|------|------|
| `web/src/app/utils/cleanUserText.ts` | 新增 5 条规则：`\ufffd`→空、`\uFFFD`→空、`undefined`→空、`NaN`→空、`[object Object]`→空、多空格→单空格 |

### 3. Latest API 支持 generation_tier（P1）

| 文件 | 修改 |
|------|------|
| `server/internal/repository/recommendation_repository.go:68` | `FindLatestByResumeID` 新增 `tier string` 参数，动态拼接 SQL `AND generation_tier = ?` |
| `server/internal/service/recommendation_service.go:682` | `GetLatestRecommendation` 新增 `tier string` 参数并透传 |
| `server/internal/http/recommendation_handler.go:103` | 读取 `r.URL.Query().Get("generation_tier")` 并传入 service |
| `web/src/api/agents.ts:61` | `getLatestRecommendation` 新增可选 `tier` 参数 |
| `web/src/api/viewModels.ts:293` | `loadLatestRecommendationViewModel` 新增可选 `tier` 参数 |

**验证**：
- `latest?resume_id=1&generation_tier=quality` → has_result=true ✅
- `latest?resume_id=1&generation_tier=fast` → has_result=false（无真实 fast 历史）✅
- `latest?resume_id=8&generation_tier=quality` → has_result=true ✅
- `latest?resume_id=8&generation_tier=fast` → has_result=false ✅
- 无乱码 ✅

### 4. 前端 quality-first 初始化（P1）

| 文件 | 修改 |
|------|------|
| `web/src/app/pages/JobRecommendationPage.tsx:144` | 重写 entry flow：先查 latest quality，完整（≥5）→展示 Deep；不完整→fallback latest fast；无 fast→生成 Fast |

**流程**：
```
loadLatestRecommendationViewModel(rid, "quality")
  ├─ quality 存在 && jobs >= 5 → setDeepResult + phase="deep_done" → 展示 Deep ✅
  ├─ quality 存在 && jobs < 5 → setDeepResult + deepErrorMessage（不完整提示）→ fallback fast
  └─ quality 不存在/失败 → fallback fast
      ├─ fast 存在 → setFastResult + phase="fast_done" → 展示 Fast ✅
      └─ fast 不存在 → generateFast → 生成并展示 Fast ✅
```

### 5. Deep 不完整提示（P1）

右侧控制台在 Deep done 状态下显示 `deepErrorMessage`（当 quality 历史不完整时）。

### 6. Deep 重新生成

通过 POST `/api/jobs/recommend` 重新生成 Deep：
- resume_id=1: HTTP 200, code=0, llm_agent, 5 items ✅
- resume_id=8: HTTP 200, code=0, llm_agent, 5 items ✅

---

## Deep / Fast 展示源修复对比

| 场景 | 修复前 | 修复后 |
|------|------|------|
| 只有 Fast | latest 返回 fast → 展示 Fast | latest quality 无 → latest fast → Fast ✅ |
| 有完整 Deep (≥5) | latest 返回 quality → Deep | latest quality 返回 5 → Deep ✅ |
| Deep 不完整 (<5) | latest 返回 quality → 伪装完整 Deep | quality 不完整 → fallback fast + 提示 ✅ |
| Deep 生成成功后 | setDeepResult + phase="deep_done" → Deep | 同左 ✅ |
| Deep 失败后 | 保留 Fast | 同左 ✅ |
| 重新打开页面 | 依赖最近非 mock | quality-first（先查 quality，再 fast）✅ |

---

## 验证结果

| 项目 | 结果 |
|------|:---:|
| typecheck | ✅ 无 target 错误 |
| npm build | ✅ PASS（1.96s） |
| go build | ✅ PASS |
| go vet | ✅ PASS |
| go test | ✅ ALL OK |
| Fast `\ufffd` scan | ✅ 0（resume 1 + 8） |
| latest tier filter | ✅ quality/fast 正确区分 |
| byte truncate scan | ✅ NO_BYTE_TRUNCATE |
| bad visible text scan | ✅ 无用户暴露 |

---

## 修改文件

| 文件 | 修改内容 | 说明 |
|------|------|------|
| `server/internal/service/semantic_profile_gate.go` | truncateStr 改为 rune-safe | P0 乱码修复 |
| `server/internal/service/semantic_profile_profiler.go` | truncate 改为 rune-safe | P0 乱码修复 |
| `server/internal/repository/recommendation_repository.go` | FindLatestByResumeID + tier | P1 API |
| `server/internal/service/recommendation_service.go` | GetLatestRecommendation + tier | P1 API |
| `server/internal/http/recommendation_handler.go` | 读取 generation_tier 查询参数 | P1 API |
| `web/src/app/utils/cleanUserText.ts` | +5 条清洗规则 | P1 防御 |
| `web/src/api/agents.ts` | getLatestRecommendation + tier | P1 前端 API |
| `web/src/api/viewModels.ts` | loadLatestRecommendationViewModel + tier | P1 前端 API |
| `web/src/app/pages/JobRecommendationPage.tsx` | quality-first + 不完整检测 + 提示 UI | P1 展示策略 |

---

## 剩余问题

| 等级 | 问题 | 是否阻塞 | 建议 |
|:----:|------|:----:|------|
| P2 | latest quality 保存格式为单条记录（非完整 5 岗位 session） | 否 | Deep 重新生成可获取完整 5 条；后续优化保存格式 |
| P2 | 无 Fast/Deep 手动切换按钮 | 否 | P3 增强 |
