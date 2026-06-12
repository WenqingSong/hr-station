# Recommendation-Display-Source-And-Garbled-Audit-I1：推荐页展示源与乱码只读审计

## 审计目标

只读分析推荐页两个问题：
1. 证据区域出现 `�...` 乱码字符
2. 有 Deep 结果时页面是否默认展示 Fast

**结论：不修改任何文件。**

---

## 1. 乱码根因

### 1.1 现象

Fast 推荐 API 的 `match_highlights`（即前端渲染的"匹配证据"）中包含 `�...`：

```
resume_id=1: "数据分析、�..."
resume_id=8: "后端接口开发�..."
```

`�` 是 Unicode 替换字符 U+FFFD，表示字节序列无效。

### 1.2 根本原因

**`server/internal/service/semantic_profile_gate.go:181-186`** 的 `truncateStr` 函数：

```go
func truncateStr(s string, maxLen int) string {
    if len(s) <= maxLen { return s }
    return s[:maxLen] + "..."  // ← BUG
}
```

Go 的 `len(s)` 返回**字节数**（不是字符数）。中文 UTF-8 编码每个字占 3 字节。

Fast Ranker 在 `server/internal/service/fast_recommend_sep_ranker.go:324` 调用 `truncateStr(ev, 120)`：

```go
s.EvidenceTexts = append(s.EvidenceTexts, truncateStr(ev, 120))
```

当 `ev` 长度为 121~122 字节且第 121 字节恰好落在中文字符中间时，`s[:120]` 会切碎该多字节字符，产生 `�`。

### 1.3 同样受影响的函数

| 文件 | 函数 | 状态 |
|------|------|:---:|
| `semantic_profile_gate.go:181` | `truncateStr(s, maxLen)` | ❌ 字节级截断 |
| `semantic_profile_profiler.go:312` | `truncate(s, maxLen)` | ❌ 字节级截断 |
| `tag_evidence_service.go:771` | `truncateEvidence(text, maxLen)` | ✅ 使用 `[]rune` |

### 1.4 前端防御现状

- `web/src/app/utils/cleanUserText.ts`：**不处理 `�` 字符**（只替换内部术语和 mock 标签）
- `web/src/adapters/tagLabelMap.ts`：**不处理 `�`**（120+ 映射，raw code 检测存在）
- 前端 adapter **不会引入**此乱码——`�` 已在 API response 中

### 1.5 修复方向

**后端修复** `truncateStr` 和 `truncate`，使用 `[]rune` 替代字节切片：

```go
func truncateStr(s string, maxLen int) string {
    runes := []rune(s)
    if len(runes) <= maxLen { return s }
    return string(runes[:maxLen]) + "..."
}
```

**同时**前端 `cleanUserText` 应增加安全网：去除 `\ufffd`。

---

## 2. 展示源分析

### 2.1 当前展示逻辑

页面初始化流程（`JobRecommendationPage.tsx:144-176`）：

```tsx
useEffect(() => {
  loadLatestRecommendationViewModel(rid)  // ← 调用 latest API
    .then((latest) => {
      if (latest.hasResult && !isMockGeneratedBy(genBy)) {
        const tier = (latest.generationTier === "quality") ? "quality" : "fast";
        setActiveTier(tier);
        if (tier === "quality") {
          setDeepResult(latest.result);   // ← 展示 Deep
          setPhase("deep_done");
        } else {
          setFastResult(latest.result);   // ← 展示 Fast
          setPhase("fast_done");
        }
      } else {
        generateFast(rid);  // ← 无历史 → 生成 Fast
      }
    })
```

### 2.2 当前后端 latest API

```
GET /api/jobs/recommend/latest?resume_id=X
```

- 调用 `FindLatestByResumeID(ctx, userID, resumeID)`
- 查询：`WHERE user_id = ? AND resume_id = ? AND status = 'ready' AND generated_by NOT IN ('mock_agent', 'mock_quality') ORDER BY created_at DESC LIMIT 1`
- **不区分 generation_tier**，返回时间上最新的非 mock 结果
- 不支持 `generation_tier` 查询参数

### 2.3 数据库实际状态

| 简历 | Fast 最新非 mock | Quality 最新非 mock | latest API 返回 | 前端展示 |
|------|:---:|:---:|------|:---:|
| resume_id=1 | 无（全部 mock_agent × 30+） | ID 37（llm_agent, 2026-06-07, 1 item） | quality·1 item | **Deep**·1 item |
| resume_id=8 | 无 | ID 47（llm_agent, 2026-06-12, 4 items） | quality·4 items | **Deep**·4 items |

### 2.4 当前是否为 quality-first？

**是。** 当前逻辑已实现 quality-first 策略：
- 如果 latest 是 quality → 展示 Deep
- 如果 latest 是 fast → 展示 Fast
- 如果无历史 → 生成 Fast

### 2.5 仍需改进的问题

| 问题 | 严重程度 | 说明 |
|------|:---:|------|
| latest API 不支持按 tier 查询 | P1 | 无法独立获取 fast/quality 历史 |
| quality 结果仅 1~4 条单岗位 | P1 | 非完整 5 岗位推荐会话（由 quality 流程保存方式所致） |
| Deep 成功后不立即持久化前端 activeTier | P2 | 下次打开依赖 latest API 查询 |
| 无显式切换按钮 | P2 | 用户无法手动切换 Fast/Deep 视图 |
| resume_id=1 quality 结果始于 June 7 | P2 | 旧数据，可能需重新生成 |

### 2.6 用户截图中展示 Fast 的可能原因

根据代码分析，以下场景会导致展示 Fast：
1. **latest API 返回 fast 记录**：如果 DB 中最近一条非 mock 记录是 fast → 展示 Fast。但当前 DB 中 resume_id=1 和 8 的 fast 全为 mock_agent（被过滤），不会触发。
2. **latest API 失败** → fallback 到 `generateFast`。
3. **latest 结果 `recommendedJobs.length === 0`** → fallback 到 `generateFast`。
4. **用户没有已保存的历史** → 初次生成 Fast。

最可能原因：用户场景中 latest API 因某种原因未返回可用 quality 结果，或 DB 在截图时刻的 quality 记录已被删除/未生成。

---

## 3. 乱码来源判定

| 层级 | 是否存在乱码 | 证据 | 结论 |
|------|:---:|------|------|
| DB result_json | 否 | ID 37/47 的 result_json 为正常 JSON | 乱码不来自存储 |
| API 原始 JSON | **是** | Fast `/recommend` 的 `match_highlights` 含 `�...` | **乱码源头** |
| adapter 映射 | 否 | `toRecommendationPageViewModel` 正确传递字符串 | 不引入新乱码 |
| UI 渲染 | 否 | React 直接渲染字符串，无额外编码转换 | 不引入新乱码 |
| cleanUserText | 否 | 不处理 `�`（也无防御） | 现有清洗不纠正 |

---

## 4. 具体乱码字段

| 字段路径 | 示例 | 来源 | 建议 |
|------|------|------|------|
| `data.recommended_jobs[].match_highlights[]` | `"数据分析、�..."` | `truncateStr(ev, 120)` 字节截断 | 修复 `truncateStr` + 前端 cleanUserText 防御 |

---

## 5. 专业化产品逻辑建议

### 当前状态评估
- 同一简历 Fast/Deep 共存：✅ 支持（`fastResult` / `deepResult` 独立 state）
- 有 Deep 时默认展示 Deep：✅ 已实现（latest tier=quality → phase=deep_done）
- Deep 失败保留 Fast：✅ 已实现（`deep_error` 保留 `fastResult`）
- 重新打开恢复 Deep：✅ 已实现（latest API 返回最近非 mock 记录）
- 显示展示源：✅ 右侧控制台显示"深度推荐"或"快速推荐"（line 605）
- 手动切换：❌ 不支持

### 建议方案

**最小改动（前端 quality-first 增强）**：
1. 页面初始化后，如有 Deep 历史 → 主动查询 quality latest 展示
2. Fast 总是调用 POST `/recommend` 实时生成（不依赖历史）
3. Deep 优先读取 latest quality → 无历史时给出引导

**最稳方案（演示阶段）**：
- 保持当前 quality-first 逻辑不变
- 确保 resume_id=1 和 8 都有 5 岗位 Deep 结果
- 确保片段不出现乱码

---

## 6. 构建结果

| 项目 | 结果 |
|------|:---:|
| npm build | ✅ PASS（2.38s） |
| go build | ✅ PASS |
| go vet | ✅ PASS |
| go test | ✅ ALL OK（10 packages） |

---

## 7. 修复方案建议

| 优先级 | 问题 | 修复位置 | 修复动作 |
|:----:|------|------|------|
| P0 | match_highlights 含 `�` | `server/internal/service/semantic_profile_gate.go:185` | `truncateStr` 改用 `[]rune` |
| P0 | match_highlights 含 `�` | `server/internal/service/semantic_profile_profiler.go:314` | `truncate` 改用 `[]rune` |
| P1 | cleanUserText 无防御 | `web/src/app/utils/cleanUserText.ts` | 增加 `�` → 空字符串的替换规则 |
| P1 | quality 结果仅 1~4 条（非 5 岗位完整输出） | 后端 quality 生成/保存逻辑 | 确保完整 5 岗位保存 |
| P2 | latest API 不支持 tier 参数 | `server/internal/http/recommendation_handler.go` | 增加 `generation_tier` 查询参数 |
| P2 | 无法手动切换 Fast/Deep | `web/src/app/pages/JobRecommendationPage.tsx` | 右侧控制台增加"查看快速推荐"按钮 |

---

## 8. 下一步

- **推荐任务名**：`Recommendation-Display-QualityFirst-And-Garbled-Fix-I2`
- **是否允许直接修复**：是（P0 仅 2 行 Go 代码 + 1 行前端防御规则）
- **是否需要先重跑 Deep**：建议对 resume_id=1 重新生成（现有 1 条旧数据）
- **是否需要后端接口变更**：可选（latest API 增加 tier 参数为 P2）
