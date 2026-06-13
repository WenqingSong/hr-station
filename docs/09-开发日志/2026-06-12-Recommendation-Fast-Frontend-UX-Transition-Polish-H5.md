# Recommendation-Fast-Frontend-UX-Transition-Polish-H5：Fast 轻量化展示 + 新旧结果无闪屏

## 目标

Fast 模式轻量化（隐藏证据/风险/策略），Deep 保留完整；新旧结果切换不闪旧内容。

---

## 修改详情

### 1. isFastDisplay 明确区分

新增 `isFastDisplay = !isDeepDisplay && activeResult !== null`，用于 Fast 模式专属展示分支。

### 2. Fast 中间详情轻量化

**文件**：`JobRecommendationPage.tsx`

| 区域 | 修复前 | 修复后 |
|------|------|------|
| 匹配证据 | Fast 显示 3 条 `matchHighlights` | **仅 Deep 显示**（`isDeepDisplay &&` guard） |
| 主要风险 | 仅 `jobSpecificRisks`（Deep 字段） | 不变 + **加 `isDeepDisplay &&` guard** |
| Fast 专属卡片 | 无 | 新增：轻量卡片"快速推荐用于岗位初筛…深度推荐会进一步分析…" |

### 3. 新旧切换无闪屏

| 修复项 | 修复位置 | 效果 |
|------|------|------|
| `generateFast` 开始时 | line 228: `setSelJobId(null)` | 清除旧岗位引用 |
| `handleDeepRecommendation` 开始时 | line 256: `setSelJobId(null)` | 清除旧岗位引用 |
| Fast 成功后 | result→selJobId→phase 顺序 | 先提交结果再切 done |
| Deep 成功后 | deepResult→selJobId→backEndDone→phase | 相同顺序 + 显式 `setPhase("deep_done")` |
| Loading 中间区 | `fast_loading/deep_loading` 时显示 loading 动画 | 不暴露旧 job 详情 |

---

## Fast / Deep 展示契约

| 模式 | 展示内容 | 隐藏内容 |
|------|------|------|
| Fast | rank + score + title + 一句话 reason + ≤3 tags + 核心判断 + Fast 轻量提示 | match_highlights / resume_evidence / risk_points / job_specific_risks / application_strategy |
| Deep | Fast 内容 + 推荐依据 + 匹配证据 + 主要风险 + 行动建议 + AI 洞察卡片 | — |

---

## 构建

| 项目 | 结果 |
|------|:---:|
| npm build | ✅ PASS (2.10s) |
| isFastDisplay scan | ✅ defined |
| selJobId reset scan | ✅ 2 sites |
| isDeepDisplay guards | ✅ evidence+risks+strategy |

---

## 修改文件

| 文件 | 修改内容 |
|------|------|
| `web/src/app/pages/JobRecommendationPage.tsx` | isFastDisplay + Fast 轻量卡片 + evidence/risks Deep-only + selJobId reset ×2 + loading guard + result→phase 顺序 |

---

## 浏览器验证

| 场景 | 预期 |
|------|------|
| /recommend 快速推荐 | 左侧榜单 + 中间只有核心判断和 Fast 轻量卡片 |
| 重新生成 Fast | middle 区显示 loading 动画（不闪旧内容） |
| 深度推荐 | 完整证据/风险/策略 |
| console | 无 useMemo/Rendered more hooks/ErrorBoundary |
