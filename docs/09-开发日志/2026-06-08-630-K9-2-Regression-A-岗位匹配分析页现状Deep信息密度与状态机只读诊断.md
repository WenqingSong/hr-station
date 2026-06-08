# K9-2-Regression-A 只读诊断报告

## 1. 当前状态确认

- **诊断时间**：2026-06-08
- **诊断页面**：`/match?resumeId=1&jobId=10001` → 岗位匹配分析页
- **LLM 启用状态**：`LLM_ENABLE=false` → 所有请求走 Mock Agent 兜底
- **数据库**：`offer_catcher.match_analyses` 中仅 1 条 quality 记录（id=113），共 ~20 条 fast 记录
- **latest API**：返回 `has_result=true`，读到 `generation_tier=fast`，`generated_by=mock_agent`

---

## 2. 页面现象

| 页面区域 | 现象 | 判断 |
|----------|------|------|
| 顶部标题 | "岗位匹配分析" + 岗位/简历名 | 正常 |
| 顶部状态 | "快速匹配 · 规则分析" | ✅ 正确反映 fast_done 状态 |
| 左侧栏 | 岗位 JD 摘要（title + 核心要求 + JD 关键词） | 核心要求摘要正常，JD 关键词显示"未获取到" |
| 中间区域 | 仅匹配结论（分数 94 + "强匹配，优先投递"）+ 升级提示 | ⚠️ Fast 简化版，缺少 Deep 详情 |
| 右侧栏 | 当前简历 + "生成深度匹配分析"按钮 | 正常，等待用户触发 |
| 信息密度 | **极低**：只有一个分数、一个等级、几条优势/短板/建议 | **根因：页面处于 fast_done 状态** |

---

## 3. 前端结构诊断

| 文件/组件 | 状态 | 问题 |
|-----------|------|------|
| `JobMatchAnalysisPage.tsx` | ✅ 状态机完整（7 阶段） | Fast 渲染访问不存在的 ViewModel 属性 |
| `MatchAnalysisSidebarLeft.tsx` | ✅ 正常展示 JD 摘要 | `jobVm.requirements` 依赖 adapter 计算（已修复） |
| `MatchAnalysisDetail.tsx` | ✅ 6 模块 Deep 布局存在 | 仅在 `!isFast` 时挂载，fast_done 不渲染 |
| `MatchAnalysisSidebarRight.tsx` | ⚠️ 存在于代码库中 | **当前页面未使用**，右侧栏为内联渲染 |
| `MatchAnalysisHeader.tsx` | ⚠️ 存在于代码库中 | **当前页面未使用**，标题为内联渲染 |
| `matchAnalysisAdapter.ts` | ⚠️ 映射不完整 | ViewModel 缺少 `matchSummary`/`matchAdvantages`/`matchRisks`/`matchSuggestions`/`profileDominantProfile` |
| `AgentStepLoadingOverlay` | ✅ 已接入 | `deep_loading` 时 4 步动画（profile→requirements→evidence→suggestions） |
| `AIThinkingLoading` | ❌ 未使用 | 旧版，已废弃 |

### 3.1 页面当前状态机

```
loading_latest → GET /api/jobs/{id}/match/latest?resume_id=X
    │
    ├── hasResult=true && tier=fast → fast_done (展示 Fast 简化版)
    ├── hasResult=true && tier=quality → deep_done (展示 MatchAnalysisDetail)
    └── hasResult=false → fast_loading → POST fast → fast_done

用户点击「生成深度匹配分析」：
    deep_loading (AgentStepLoadingOverlay) → POST quality
    ├── success → deep_done (展示 MatchAnalysisDetail)
    └── error → deep_error (保留当前结果 + 可重试)
```

### 3.2 关键判断

| 诊断项 | 结果 |
|--------|------|
| 页面进入是否先 GET latest | ✅ 是，`useEffect` 中 `loadLatestMatchViewModel(jobId, rid)` |
| has_result=true 时是否不 POST fast | ✅ 是，直接进入 fast_done 或 deep_done |
| has_result=false 时是否 POST fast | ✅ 是，自动 `loadMatchAnalysisViewModel(jobId, {resume_id, generation_tier:"fast"})` |
| 点击 Deep 是否 POST quality | ✅ 是，`handleDeepAnalysis()` 调用 `generation_tier:"quality"` |
| deep_done 展示哪些区块 | ✅ 6 模块：匹配结论、Top 3 修复、JD 证据链、优势、风险、下一步建议 |
| fast_done 展示哪些区块 | ⚠️ 4 模块：匹配结论、核心优势(前2)、主要短板(前1)、快速建议(前1) + 升级提示 |
| Deep 完整区块是否存在但未渲染 | ✅ 存在（`MatchAnalysisDetail` 组件），但仅在 `!isFast` 时渲染 |
| 是否仍使用 AIThinkingLoading | ❌ 否，已替换为 AgentStepLoadingOverlay |
| 是否已接入 AgentStepLoadingOverlay | ✅ 是，deep_loading 阶段使用 |
| 是否存在旧文案 / mock 字段暴露 | ⚠️ Fast 区访问 `activeResult.matchSummary`/`matchAdvantages`/`matchRisks`/`matchSuggestions`，这些字段不在 ViewModel 中 |

---

## 4. 后端 API 诊断

| API | Fast | Quality | 问题 |
|-----|------|---------|------|
| POST /api/jobs/{id}/match | `generation_tier=fast`, `generated_by=mock_agent` | `generation_tier=quality`, `generated_by=mock_quality` | **两者内容几乎完全相同** |
| GET /api/jobs/{id}/match/latest | 返回 ready 记录（按 resume+job+user） | 同上 | 只查 `status='ready'`，按 `created_at DESC` |

### 4.1 match API 当前字段契约

**后端 Go Model (`MatchAnalysisResult`)**：30+ 字段
- 基础：`match_score`, `match_level`, `screening_result`, `evidence_chain`
- 扩展（omitempty）：`gap_points`, `interview_questions`, `apply_readiness`, `optimization_stop_signal`, `top_fix_items`, `profile_type`, `profile_dominant_profile`

**前端 TS Type (`MatchAnalysisResult`)**：10 字段
- `match_score`, `match_level`, `summary`, `apply_readiness`, `optimization_stop_signal`, `top_fix_items`, `evidence_chain`, `match_highlights`, `risk_items`, `next_actions`

**前端 ViewModel (`MatchAnalysisPageViewModel`)**：10 字段
- `matchScore`, `matchLevel`, `summary`, `applyReadiness`, `optimizationStopSignal`, `topFixItems`, `evidenceChain`, `matchHighlights`, `riskItems`, `nextActions`

### 4.2 关键判断

| 诊断项 | 结果 |
|--------|------|
| fast / quality 是否共用同一套字段 | ✅ 是，共享 `MatchAnalysisResult` 结构体 |
| quality 是否真的生成更完整内容 | ❌ **否。Mock Agent 中 fast 和 quality 调用同一个 `mockAgent.Analyze()`，输出内容完全一致** |
| score / level 是否合理 | ⚠️ 简单方向匹配 + 标签矩阵（35-95），缺少真实 JD 语义分析 |
| 是否消费 ResumeProfile | ✅ 是，通过 `ProfileProvider.EnsureProfile()` 读取标签 |
| 是否消费完整简历内容 | ❌ **否。Mock Agent 只使用简历画像标签，不读取简历原文** |
| 是否消费完整 JD | ⚠️ Mock Agent 使用 JD 的 `direction`/`Department`+技能矩阵，但不读 `description`/`requirements` 原文 |
| 是否有 Hard Guard | ❌ **否。Match 模块没有 Hard Guard 机制。K9-1 有但 K9-2 没有** |
| 是否有三层 Agent 思路 | ❌ **否。LLM Agent 是单次 Prompt 调用，Mock Agent 是规则引擎** |
| 是否只是规则包装 | ✅ **是。Mock Agent 完全是规则匹配（方向+标签矩阵），LLM 也仅在 `LLM_ENABLE=true` 时可用** |
| 是否保存 generation_tier/generated_by/status | ✅ 是，`match_analyses` 表有这些列 |
| latest API 查询维度是否正确 | ✅ 按 user+resume+job+status=ready 查询，按 created_at 降序 |

---

## 5. 字段密度诊断

| 字段 | Fast API | Quality API | Fast 前端展示 | Deep 前端展示 | 问题 |
|------|----------|-------------|---------------|---------------|------|
| match_score | ✅ 94 | ✅ 94 | ✅ 显示 | ✅ 显示 | |
| match_level | ✅ "强匹配，优先投递" | ✅ 同 | ✅ 显示 | ❌ 未使用 | |
| summary | ❌ API 中无 | ❌ API 中无 | ❌ `matchSummary` undefined | ✅ 使用 `data.summary` | fast 访问不存在的字段 |
| evidence_chain | ❌ null | ❌ null | ❌ | ⚠️ 展示但为空 | Mock 不生成 evidence_chain |
| match_highlights | ✅ 2 条 | ✅ 2 条 | ❌ `matchAdvantages` undefined | ✅ 在"你的优势"中展示 | fast 用错属性名 |
| risk_points | ✅ 1 条 | ✅ 1 条 | ❌ `matchRisks` undefined | ✅ 在"主要风险"中展示 | fast 用错属性名 |
| optimization_suggestions | ✅ 1 条 | ✅ 1 条 | ❌ `matchSuggestions` undefined | ❌ 未使用 | fast 用错属性名 |
| top_fix_items | ❌ null | ❌ null | ❌ | ❌ 空数组 | Mock 不生成 |
| gap_points | ❌ null | ❌ null | ❌ | ❌ | Mock 不生成 |
| interview_questions | ❌ null | ❌ null | ❌ | ❌ | Mock 不生成 |
| delivery_strategy | ❌ null | ❌ null | ❌ | ❌ | Mock 有函数但未放入结果 |
| apply_readiness | ❌ 空 | ❌ 空 | ❌ | ⚠️ 显示默认值 | Mock 不生成 |
| optimization_stop_signal | ❌ 空 | ❌ 空 | ❌ | ⚠️ 显示默认值 | Mock 不生成 |
| next_steps | ❌ null | ❌ null | ❌ | ⚠️ 显示默认文案 | Mock 不生成 |
| profile_dominant_profile | ✅ algorithm_research | ✅ algorithm_research | ✅ 右侧显示 | ❌ | |

### 5.1 结论

**Fast 和 Quality API 返回字段密度完全相同**。因为 `LLM_ENABLE=false`，两者都走 `matchMockAgent.Analyze()`（Mock Agent），唯一的区别是 `generated_by` 字段值（`mock_agent` vs `mock_quality`）。

**Quality 没有生成更丰富的内容**。`evidence_chain`、`top_fix_items`、`gap_points`、`interview_questions`、`delivery_strategy`、`apply_readiness`、`optimization_stop_signal`、`next_steps` 全部为 null/空。

---

## 6. latest-first 状态机诊断

| 场景 | 行为 | 结论 |
|------|------|------|
| 首次访问，DB 已有 fast ready 记录 | latest 返回 fast → 进入 fast_done | ✅ 正常 |
| 首次访问，DB 无记录 | latest 返回 has_result=false → 自动 POST fast → fast_done | ✅ 正常 |
| 已有 quality ready 记录（DB id=113） | latest 返回 quality → 进入 deep_done | ✅ 正常（但 quality 内容=fast 内容） |
| 已有 fast ready 后又点 Deep | POST quality → deep_done | ✅ 状态流转正确，但内容无变化 |

---

## 7. DB 持久化诊断

| 项 | 结果 |
|----|------|
| 是否有 quality ready 记录 | ✅ id=113, generation_tier=quality, generated_by=mock_quality, status=ready |
| metadata 与 result_json 是否一致 | ✅ generation_tier/generated_by 列值一致 |
| Deep 字段是否为空 | ❌ evidence_chain=null, gap_points=NULL, delivery_strategy=NULL, apply_readiness=NULL, top_fix_items=NULL |
| latest 是否会读到空壳 fast | ⚠️ 会！当前 latest 返回 id=113 (quality) 但其内容与 fast (id=112) 完全相同 |
| 是否存在旧结构污染 | ❌ 不存在 |

---

## 8. 根因判断

| 级别 | 问题 | 证据 |
|------|------|------|
| **P0** | Mock Agent fast/quality 输出完全一致 | API 对比：fast(id=112) 和 quality(id=113) result_json 除 generated_by 外完全相同 |
| **P0** | Mock Agent 不生成 Deep 字段 | DB evidence_chain/gap_points/delivery_strategy/top_fix_items 全部为 null |
| **P1** | Fast 前端使用不存在的 ViewModel 属性 | `matchSummary`/`matchAdvantages`/`matchRisks`/`matchSuggestions` 不在 ViewModel 中 |
| **P1** | 当前页面信息密度低 | fast_done 只展示匹配结论+3条简化信息+升级提示，Deep 未触发 |
| **P2** | 无 Hard Guard | Match Agent 仅有方向冲突 cap 规则，与 K9-1 的 Feature Card + Hard Guard 差距大 |
| **P2** | 无三层 Agent | 后端是单次 Prompt/Mock 调用，没有 Requirement→Evidence→Explanation 分层 |
| **P2** | LLM 未启用（测试环境） | `LLM_ENABLE=false`，所有请求走 Mock |
| **P3** | delivery_strategy 有函数但不返回 | `match_mock.go:405` 有 `buildDeliveryStrategy()` 但未写入结果 |
| **P3** | Fast 前端 direct 访问未适配的属性 | TypeScript 编译期不报错（`any`），运行时全部 undefined |

### 浏览器现象根因

1. **为什么页面只展示"匹配结论"**：因为 latest 返回 fast tier，页面进入 fast_done，只渲染 Fast 简化版
2. **为什么中间详情没有岗位要求拆解、简历证据、风险、补强建议**：因为 Deep 未触发，`MatchAnalysisDetail` 组件未挂载
3. **是因为还在 fast_done？** ✅ 是
4. **是因为 latest 读到 fast？** ✅ 是，latest 返回 tier=fast（id=97，2026-06-07 的记录）
5. **是因为 Deep 没生成？** ⚠️ DB 有 quality 记录(id=113)但它是刚 POST 生成的（2026-06-08 09:11:26）
6. **是因为 Deep 生成了但前端没展示？** ❌ Deep 生成的内容与 fast 完全相同
7. **是因为字段名不一致？** ⚠️ 部分：fast 渲染区使用 `matchAdvantages`/`matchRisks`/`matchSuggestions` 而 ViewModel 是 `matchHighlights`/`riskItems`/`optimizationSuggestions`
8. **是因为 MatchAnalysisDetail 组件未挂载？** ✅ 是，fast_done 状态下 `isFast=true`，不渲染 MatchAnalysisDetail

---

## 9. 与 K9-1 架构差距

| 能力 | K9-1（岗位推荐） | K9-2 当前（匹配分析） | 差距 |
|------|-----------------|---------------------|------|
| Feature Card | ✅ ResumeFeatureCard + JobFeatureCard | ❌ 无 | **巨大** |
| Hard Guard | ✅ applyHardGuard() 7 条规则 | ❌ 仅有方向冲突 cap | **巨大** |
| 三层 Agent | ✅ Mock 中内置 FeatureCard→Guard→Score pipeline | ❌ 单次 Mock/LLM 调用 | **大** |
| Deep 字段 | ✅ per-job explanation (why_ranked_here, resume_evidence, job_specific_risks...) | ❌ evidence_chain/gap_points/top_fix/delivery_strategy 全部 null | **巨大** |
| LLM 质量 | ✅ RecommendQualityWithFeatureCards pipeline | ❌ 单次 LLM Prompt（同样不生成 Deep 字段） | **大** |
| Fast vs Quality 差异 | ✅ Quality 使用更深 pipeline | ❌ 完全相同 | **巨大** |
| 前端信息密度 | ✅ 完整 Deep 布局 | ⚠️ 完整的 Deep 组件存在但未接收有效数据 | **中** |

---

## 10. 重构建议

| 方案 | 说明 | 是否推荐 |
|------|------|---------|
| **A: 只补前端字段** | 修复 Fast 的 ViewModel 属性名，让 fast_done 至少能正常展示现有数据 | ❌ 不推荐单独做，意义太小 |
| **B: 只补后端 Deep 字段** | Mock Agent 中补齐 evidence_chain/gap_points/top_fix/delivery_strategy 生成 | ⚠️ 部分推荐，但不够 |
| **C: 引入 Feature Card + Hard Guard** | 仿照 K9-1 构建 ResumeFeatureCard/JobFeatureCard，加入 Hard Guard 决策 | ✅ 推荐，解决评分合理性问题 |
| **D: 三层 Agent 重构** | Requirement Agent → Evidence Scoring Agent → Explanation Agent | ✅ 推荐，解决 Deep 字段贫瘠问题 |
| **E: 完整 K9-1 对齐** | C + D 组合：Feature Card + Hard Guard + 三层 Agent | ✅ **最推荐** |

### 最小可落地方案（推荐顺序）

1. **第一步**：后端 Mock Agent 补齐 Deep 字段生成（evidence_chain/gap_points/top_fix/delivery_strategy/apply_readiness）
2. **第二步**：后端引入 Feature Card + Hard Guard（复用 K9-1 的 buildResumeFeatureCard/buildJobFeatureCard/applyHardGuard）
3. **第三步**：后端 quality tier 使用三层 Agent（Requirement→Evidence→Explanation）
4. **第四步**：前端修复 Fast 属性名，并确保 Deep 字段完整展示

### 需要几个 prompt

- **当前**：1 个（`match_analysis_prompt`，单次 LLM 调用）
- **建议**：3 个
  1. Requirement Agent prompt：拆岗位要求→匹配维度→评分标准
  2. Evidence Scoring Agent prompt：简历证据逐项验证→缺口识别→风险评分
  3. Explanation Agent prompt：生成完整报告→面试表达点→投递策略

### 推荐阶段拆分

| 阶段 | 名称 | 目标 |
|------|------|------|
| K9-2-Regression-B | DeepFields | 后端补齐 Deep 字段生成 + Feature Card + Hard Guard |
| K9-2-Regression-C | ThreeLayerAgent | 后端 quality tier 三层 Agent prompt |
| K9-2-Regression-D | FrontendAlign | 前端修复属性名 + 完整信息密度展示 |

---

## 11. 建议后续阶段

| 阶段 | 名称 | 目标 |
|------|------|------|
| K9-2-Regression-B | DeepFields + FeatureCard + HardGuard | 后端 Mock quality 生成完整 Deep 字段；引入 Feature Card + Hard Guard |
| K9-2-Regression-C | ThreeLayerAgent | 3 个分层 Agent prompt，quality tier 使用三层 pipeline |
| K9-2-Regression-D | FrontendDensity | 前端修复 Fast 属性名对齐 ViewModel；Deep 6 模块完整展示新字段 |

---

## 12. 是否允许进入执行阶段

**建议 K9-2-Regression-B-Rebuild**

根因判断：
- ❌ 不是前端展示问题（Deep 组件完整但无有效数据）
- ❌ 不是后端字段薄（**确实薄**，但更根本的是架构弱）
- ✅ **候选/评分/证据全弱**：Mock Agent 没有 evidence_chain（证据链完全为空）；没有 gap_points（缺口分析为空）；没有 top_fix_items（修复建议为空）；没有 three-layer analysis

---

## 13. 未修改范围确认

**未修改任何源码** ✅
**未修改 K9-1 岗位推荐** ✅
**未修改 K9-3 简历优化建议** ✅
**未修改 K9-4 简历版本对比** ✅
**未修改 server/sql/init.sql** ✅
**未修改数据库结构** ✅
**未删除真实简历** ✅
**未执行 reset-up** ✅
**未执行 Git 命令** ✅

---

## 14. docs 更新清单

新增：
- `docs/09-开发日志/2026-06-08-630-K9-2-Regression-A-岗位匹配分析页现状Deep信息密度与状态机只读诊断.md`

待更新：
- `docs/下一步任务.md`
- `docs/变更记录.md`

---

## 15. Git 约束确认

未执行任何 Git 命令 ✅
