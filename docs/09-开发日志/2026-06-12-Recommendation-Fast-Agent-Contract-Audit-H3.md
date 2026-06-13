# Recommendation-Fast-Agent-Contract-Audit-H3：审计 Fast 推荐 Agent 输出与前端消费契约

## 审计目标

确认 Fast 推荐是否输出过重文本、前端是否错误展示 Deep 风格内容、新旧结果切换时是否有闪屏。

---

## 1. Fast 后端链路

| 项目 | 结果 |
|------|------|
| Fast 入口函数 | SEP Ranker: `fastRecommendWithSEP()` (`recommendation_service.go:848`) |
| 回退入口 | `fastRecommendWithTags()` (`recommendation_service.go:750`) — TagEvidence |
| 是否调用 LLM | ❌ 不调用，纯规则匹配（SEP 标签 + Career Direction Gate + 加权评分） |
| generated_by | `fast_tag_engine` |
| 主要排序依据 | Career Direction（+40/+22/-10/-40）→ TechMatch → EngMatch → ExpSignal → Domain → RiskPenalty |
| 当前输出字段 | `recommendation_reason`, `match_highlights`, `risk_points`, `resume_evidence`, `display_tags`, `why_ranked_here`, `suggested_action` |
| 耗时 | ~0.5-1.5s（SEP 标签匹配 + 排序，无 LLM） |

---

## 2. Fast API 输出审计

| 字段 | 当前返回 | 适合 Fast 展示 | 建议 |
|------|:---:|:---:|------|
| job_id / title / score | ✅ | ✅ | 保留 |
| display_tags | ✅ (2-4 个) | ✅ | 保留（≤3 个） |
| recommendation_reason | ✅（一句话，~30-60 字） | ✅ | 保留（左侧卡片/oneSentenceReason） |
| match_highlights | ✅（≤3 条，每条约 80-120 字） | ⚠️ 偏重 | Fast 模式隐藏 |
| risk_points | ✅（1-2 条模板化文本） | ⚠️ 模板化 | Fast 模式隐藏 |
| resume_evidence | ✅（原始 SEP evidence） | ❌ 技术文本 | Fast 模式隐藏 |
| why_ranked_here | ✅（方向语义匹配总结） | ✅ | 保留（一句话） |
| suggested_action | ✅ | ❌ 不适合 | Fast 模式隐藏 |
| job_specific_reason | ❌ 不返回 | — | — |
| application_strategy | ❌ 不返回 | — | — |

**结论**：Fast 返回了 `match_highlights`、`risk_points`、`resume_evidence`、`suggested_action`，这些语义过重，应该只在 Deep 模式下展示。

---

## 3. 前端 Adapter 契约

| 项目 | 当前情况 | 问题 | 建议 |
|------|------|------|------|
| Fast/Deep 是否区分 | ✅ `getSource(generatedBy)` → `fast_tag` / `deep_llm` | — | — |
| 是否补 Deep 字段 | ❌ adapter 不区分 tier，统一处理 | Fast 字段走 Deep 路径（humanizeMatchHighlights 等） | 按 `isQuality` 分支 |
| 是否生成 fallback 文案 | ✅ humanizeXxx 函数生成兜底 | Fast 不需要 | 保留 humanize 但跳过过度处理 |
| 是否有 raw source 防御 | ✅ `isRawTagCode` 检查 | — | — |
| 是否建议改 adapter | ✅ 是——Fast 模式应跳过 evidence/risks/highlights 的人性化处理 | | |

---

## 4. 推荐页展示审计

| 区域 | 当前展示 | 问题 | Fast 模式建议 |
|------|------|------|------|
| 左侧榜单 | rank + score + title + 一句话 reason + 3 tags | 信息密度合理 | 保留现状 |
| 中间详情 | 核心判断 + 匹配证据(3条) + 主要风险(3条) + 行动建议 | **过重！** 像 Deep 分析 | **Fast：只显示核心判断**；Deep：完整报告 |
| 右侧控制台 | 当前展示源 + 智能分析 | ✅ | 保留 |
| Loading | "正在生成快速推荐…" | ✅ | 保留 |

### Fast/Deep 中间详情的当前分支

| 区域 | Fast 模式 | Deep 模式 |
|------|:---:|:---:|
| 核心判断 (line 517-525) | `oneSentenceReason` | `oneSentenceReason \|\| jobSpecificReason` |
| 推荐依据 (line 567) | ❌ 不显示 | `jobSpecificReason` |
| 匹配证据 (line 575) | ✅ `matchHighlights`（最多3条） | `resumeEvidence` |
| 主要风险 (line 592) | ❌ 不显示（仅 jobSpecificRisks） | `jobSpecificRisks` |
| 行动建议 (line 609) | ❌ 不显示 | `applicationStrategy \|\| suggestedImprovement` |

**发现**：Fast 模式的"匹配证据"区域显示了 3 条 `matchHighlights`（后端拼接的字符串，长约 80-120 字），这对 Fast 定位过重。

---

## 5. 新旧结果闪屏根因

### 根因：状态提交顺序 + activeResult 依赖 phase

**闪屏链路**（以点击"生成深度推荐"为例）：

1. 用户当前在 `fast_done`，`activeResult = fastResult`（Fast Top5）
2. 点击 Deep → `setPhase("deep_loading")`——此时 `activeResult` 仍为 `fastResult`
3. L2 的 `activeResult` 逻辑：`deepReady?deepResult : fastReady?fastResult : deepResult??fastResult`
4. 如果 `deepResult` 已有旧值（如上次的 Deep 4 条），它可能立刻被选中，导致中间详情的 selectedJob 仍指向旧 Deep 的岗位
5. `setDeepBackendDone(true)` + `setDeepResult(vm)`——新 Deep 写入
6. `setPhase("deep_done")`——下一帧 reactive 切换

**关键矛盾**：
- `setDeepResult(new)` 和 `setPhase("deep_done")` 是同一个事件循环但两次 setState
- 中间有一次渲染 `deepResult=new` 但 `phase` 还是旧值
- `activeResult` 计算基于 `deepReady`（≥5），不依赖 phase，**应不受影响**

**当前 L2 修复后的评估**：L2 的 `activeResult` 不再依赖 phase，改为基于 `deepReady/fastReady`，已消除闪烁问题的主要来源。残留问题：`selectedJob` 可能在切换时短暂指向旧 Deep 的岗位（因为 `selJobId` 未被重置）。

---

## 6. Fast / Deep 最终产品契约建议

| 层级 | 定位 | 展示内容 | 不应该展示 |
|------|------|------|------|
| Fast | 岗位初筛榜单（秒级） | rank + score + title + 一句话 reason + ≤3 tags | match_highlights / resume_evidence / risk_points / strategy |
| Deep | 深度分析报告（分钟级） | 核心判断 + 推荐依据 + 匹配证据 + 主要风险 + 行动建议 + 候选定位 | — |

---

## 7. 后续任务拆分

| 任务 | 是否需要 | 原因 | 目标 |
|:---:|:---:|------|------|
| H4 Fast Agent 输出契约重构 | ⚠️ 可选 | Fast 后端输出已足够轻量化（无 LLM），字段主要是前端过度展示 | 如需缩短后端文案生成耗时，可精简 evidence 拼接 |
| H5 前端 UX 与无闪屏切换 | ✅ 必须 | Fast 中间详情过度展示 Deep 字段 + 切换时 selJobId 残留 | Fast hide evidence/risks + reset selJobId before switching |
| H6 浏览器最终验收 | ✅ | 收口 | 完整验证 |

---

## 8. 构建检查

| 项目 | 结果 |
|------|:---:|
| npm build | ✅ PASS (1.85s) |
| go build | ✅ PASS |

---

## 9. 风险等级

| 等级 | 问题 | 是否阻塞演示 | 建议 |
|:----:|------|:----:|------|
| P1 | Fast 中间详情显示 match_highlights（过重） | 否，但影响专业观感 | H5 hide for Fast |
| P2 | 新旧切换时 selJobId 未重置 | 否 | H5 resetOnNewResult |
| P2 | Fast API 耗时 ~1-2s（可接受） | 否 | H4 可选优化 |

---

## 10. 给用户的结论

Fast 推荐后端已是纯规则匹配（不调 LLM），评分通过 Career Direction Gate + SEP 标签矩阵完成，耗时在 0.5-1.5s 区间，符合"秒级轻量"定位。Fast 不调用 LLM，不生成长文本证据——问题在前端展示层：中间详情区域在 Fast 模式下显示了 3 条 `match_highlights`（技术文本），让 Fast 看起来像 Deep 分析。adapter 不区分 Fast/Deep 统一走人性化处理，加重了 Fast 的视觉负担。建议先做 H5：前端按 `isDeepDisplay` 分支，Fast 模式隐藏证据/风险/策略区，只留核心判断。H4 可跳过（后端无需改契约），或与 H5 合并为一次前端重构。闪屏问题 L2 已大为缓解（activeResult Deep-first），残留的 selJobId 不为旧结果可在 H5 一并修复。**建议先 H5，再 H6 最终验收。**
