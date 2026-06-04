# K8-5-28-A：四核心 AI 页面调用链路与 Evidence 展示现状盘点

## 本阶段性质

只读盘点，不修改源码。审计四核心 AI 页面前后端链路与 evidence 字段展示现状。

---

## 四核心页面调用链路总表

| 页面 | API | 请求参数 | response ViewModel | UI 组件 |
|------|-----|---------|-------------------|---------|
| 岗位推荐 | POST `/jobs/recommend` | `resume_version_no`, `limit` | `RecommendationPageViewModel` | `RecommendationDetail` + `RecommendationSidebar` + `RecommendationContext` + `RecommendationHeader` |
| 岗位匹配 | POST `/jobs/{jobId}/match` | `resume_version_no` | `MatchAnalysisPageViewModel` | `MatchAnalysisDetail` + `MatchAnalysisSidebarLeft/Right` + `MatchAnalysisHeader` |
| 简历优化 | POST `/resumes/optimize` | `resume_text`, `target_job{title,jd_text}`, `target_section`, `mode` | `ResumeOptimizationPageViewModel` | `ResumeOptimizationDetail` + `ResumeOptimizationHeader` + `ResumeOptimizationSidebarLeft/Right` |
| 版本对比 | GET `/resumes/versions/compare` | `from_id`, `to_id` | `ResumeComparisonViewModel` | 字段变化表 + 风险变化表 + 建议完成 + 关键改进 |

---

## evidence 字段接收 / 展示总表

| 字段 | 后端有 | 前端类型有 | 页面展示有 | 差距 |
|------|:---:|:---:|:---:|------|
| `TopFixItem.resume_module` | ✅ match.go L42 | ❌ 无 | ❌ 无 | 前后端断开 — 前端未接收 |
| `MatchEvidence.resume_module` | ✅ match.go L22 | ❌ 无 | ❌ 无 | 前后端断开 — 前端未接收 |
| `ResumeOptimizationFixItem.resume_module` | ✅ resume_optimization.go L39 | ❌ 无 | ❌ 无 | 前后端断开 — 前端未接收 |
| `FieldChange.module_name` | ✅ version_compare.go L23 | ❌ 无 | ❌ 无 | 前后端断开 — adapter 也未映射 |
| `FieldChange.change_type` | ✅ version_compare.go L21 | ✅ 有 | ✅ 有 | 展示为"新增/修改/移除/不变" |
| `MatchEvidence.resume_point` | ✅ match.go L20 | ✅ 有（resumeEvidence） | ✅ 有 | 匹配页"简历证据"列 |
| `JDEvidenceItem.resume_evidence` | ✅ resume_optimization.go L51 | ❌ 无 | ❌ 无 | 优化页完全未展示 JDEvidenceChain |
| `SuggestionCompletion.evidence` | ✅ version_compare.go L38 | ✅ 有 | ✅ 有 | 对比页"证据"字段 |
| `evidence_chain` (推荐) | ✅ recommendation.go (via Mock) | ✅ 有 | ✅ 有 | 推荐页"匹配证据"表 |
| `evidence_chain` (匹配) | ✅ match.go L56 | ✅ 有 | ✅ 有 | 匹配页"JD 证据链分析"表 |
| `top_fix_items` (推荐) | ✅ recommendation.go L46 | ✅ 有 | ❌ 无 | 推荐页仅展示 summary+risk，无 fix items |
| `top_fix_items` (匹配) | ✅ match.go L78 | ✅ 有 | ✅ 有 | 匹配页"投递前最应该先修的 3 件事" |
| `top_fix_items` (优化) | ✅ resume_optimization.go L75 | ✅ 有 | ✅ 有 | 优化页"投递前最应该先修的 3 件事" |
| `jd_evidence_chain` (优化) | ✅ resume_optimization.go L81 | ❌ 无 | ❌ 无 | 优化页完全未展示 JD 证据链 |

---

## 岗位推荐页面盘点

| 检查项 | 当前情况 | 问题 | 建议 |
|--------|---------|------|------|
| 展示推荐岗位 | ✅ 展示 Top 5，含推荐等级/关键词/所在地/原因/风险标签 | — | — |
| 展示推荐原因 | ✅ "一句话总结" + reason | — | — |
| 展示 top_fix_items | ❌ 未展示 | 后端已有 top_fix_items 含 resume_module，前端未渲染 | K8-5-28-C 补 |
| 展示 resume_module | ❌ 未展示 | 前端 types 无 resume_module 字段 | K8-5-28-B 补类型 |
| 展示 evidence_chain | ✅ 展示"匹配证据"表（岗位要求/简历证据/证据强度） | — | — |
| 展示 JD/简历证据 | ✅ jdRequirement + resumeEvidence + evidenceStrengthLabel | — | — |
| 能体现 AI 技能/证书/项目证据 | ❌ 不能 | evidence 内容来自 Mock，但不区分模块 | — |

---

## 岗位匹配页面盘点

| 检查项 | 当前情况 | 问题 | 建议 |
|--------|---------|------|------|
| 展示匹配分 | ✅ 大号数字 + /100 | — | — |
| 展示 evidence_chain | ✅ "JD 证据链分析"表（JD 要求/简历证据/证据强度/风险说明） | 缺 resume_module 列 | K8-5-28-D 补 |
| 展示 resume_evidence | ✅ "简历证据"列 | — | — |
| 展示 jd_evidence | ✅ "JD 要求"列 | — | — |
| 展示 resume_module | ❌ 未展示 | 前端 EvidenceItem 无 resume_module | K8-5-28-B 补类型 |
| 展示 top_fix_items | ✅ "投递前最应该先修的 3 件事" | 缺 resume_module 标签 | K8-5-28-D 补 |
| 展示 gap/风险 | ✅ 匹配结论 + 优势 + 主要风险 | — | — |

---

## 简历优化页面盘点

| 检查项 | 当前情况 | 问题 | 建议 |
|--------|---------|------|------|
| 展示 fix_items | ✅ "投递前最应该先修的 3 件事" | 缺 resume_module 标签 | K8-5-28-E 补 |
| 展示 resume_module | ❌ 未展示 | 前端 TopFixItem 无 resume_module | K8-5-28-B 补类型 |
| 展示 JD 证据链 | ❌ 完全未展示 | 后端 JDEvidenceChain 完整，前端 types/adapter/UI 都缺 | K8-5-28-E 补 |
| 按模块展示优化建议 | ❌ 未按模块分组 | 关键词对齐 + 项目优化 + 改写示例混排 | K8-5-28-E 补模块分组 |
| 区分有 JD / 无 JD | ⚠️ 页面支持但无标识 | API 传 `target_job` 可为空，UI 不区分 | — |
| 展示优先级 | ✅ P0/P1/P2 通过 color/title 区分 | — | — |
| 展示 verification_notice | ✅ "需要你核实后再写入简历的信息" | — | — |
| 展示 rewrite_examples | ✅ "修改前后对比示例" | — | — |

---

## 简历对比页面盘点

| 检查项 | 当前情况 | 问题 | 建议 |
|--------|---------|------|------|
| 展示 field_changes | ✅ "字段变化明细"表（字段/旧值/新值/变化类型/影响） | 缺 module_name 列 | K8-5-28-F 补 |
| 展示 module_name | ❌ 未展示 | 前端 types 无 module_name，adapter 未映射 | K8-5-28-B 补类型 |
| 展示 change_type | ✅ 展示为"新增/修改/移除/不变"标签 | — | — |
| 按 9 模块展示变化 | ❌ 未按模块分组 | 字段变化按 field 名排平铺，无模块分组 | K8-5-28-F 补模块分组 |
| 展示 improved / weakened / risk | ✅ 关键改进 + 风险变化 + 剩余风险 | — | — |
| 展示 suggestion_completion | ✅ "优化建议完成情况" + evidence 字段 | — | — |
| 展示 score | ✅ 评分变化（before → after + delta） | — | — |

---

## 影响范围判断

| 层级 | 是否需要改 | 原因 |
|------|:---:|------|
| 前端 types (`agent.ts`) | ✅ 是 | TopFixItem / EvidenceItem 缺 `resume_module` 字段 |
| 前端 types (`resume.ts`) | ✅ 是 | FieldChange 缺 `module_name` 字段 |
| 前端 types (`agent.ts`) | ✅ 是 | 缺 JDEvidenceChain 类型（ResumeOptimizationJDEvidenceItem） |
| 前端 adapters (`agentAdapter.ts`) | ✅ 是 | TopFixItem/EvidenceItem ViewModel 缺 resumeModule |
| 前端 adapters (`resumeAdapter.ts`) | ✅ 是 | FieldChange 映射缺 moduleName |
| 前端 adapters (`resumeOptimizationAdapter.ts`) | ✅ 是 | 缺 JDEvidenceChain 映射 |
| 前端页面 UI | ✅ 是 | 推荐页缺 top_fix_items；优化页缺 JD evidence_chain；四页面均缺模块标签 |
| 前端 components | ✅ 是 | 需为每个页面补 module 标签 + table 列 |
| 后端 model | ❌ 不 | 已有全部 evidence 字段 ✅ |
| Agent / Prompt / Schema | ❌ 不 | 已完成 ✅ |
| Mock | ❌ 不 | 已完成 ✅ |
| SQL | ❌ 不 | 不涉及 ✅ |

---

## 后续阶段拆分建议

根据盘点结果，建议以下阶段：

| 编号 | 名称 | 范围 | 优先级 |
|------|------|------|:---:|
| **K8-5-28-B** | 前端 Agent types/API/adapter 补齐 evidence 字段 | types/adapters 层 | P0 |
| **K8-5-28-C** | 岗位推荐页 evidence 展示增强（resume_module + top_fix_items） | recommendation 组件 | P1 |
| **K8-5-28-D** | 岗位匹配页 evidence_chain resume_module 列 + top_fix_items 模块标签 | match 组件 | P1 |
| **K8-5-28-E** | 简历优化页 JDEvidenceChain 展示 + 模块分组 + resume_module 标签 | resume-optimization 组件 | P1 |
| **K8-5-28-F** | 简历版本对比页 module_name 列 + 按模块分组展示 | ResumeComparisonPage | P1 |
| **K8-5-28-G** | 四核心 AI 页面 Mock API 端到端回归 | 全页面 | P2 |
| **K8-5-28-Z** | 四核心 AI 页面 evidence 最终验收 | 验收 | P2 |

建议执行顺序：**B → C/D/E/F 可并行 → G → Z**。

K8-5-28-B 是 P0 最高优先（不补类型和 adapter，C/D/E/F 无法开展）。

---

## 构建 / API / 测试结果

| 验收项 | 结果 |
|--------|:---:|
| `go test ./...` (10 packages) | ✅ 全部通过 |
| `npm run build` (2.22s) | ✅ |
| API /jobs 200 | ✅ |
| API /jobs/10001 200 | ✅ |
| API /jobs/facets 200 | ✅ |
| API /resumes 401 | ✅（正常后端保护） |

---

## 未修改范围确认

```text
未修改前端页面 ✅
未修改后端源码 ✅
未修改数据库 SQL ✅
未修改 Agent / Prompt / Schema / Mock ✅
未修改 API 路径 ✅
未新增 /api/resumes/v2 ✅
未修改登录鉴权 ✅
未修改 jobs 页面 ✅
未修改 .env ✅
未安装依赖 ✅
未迁移目录 ✅
未新增上传能力 ✅
未执行真实写入 ✅
```

---

## 核心发现总结

1. **后端 evidence 字段完整**：resume_module / module_name / ResumeEvidence / JDEvidenceChain / ChangeType 全部存在于后端 Go model。
2. **前端 evidence 字段大量缺失**：TopFixItem / EvidenceItem / FieldChange 三个前端类型均缺少 module 字段。JDEvidenceChain 类型完全缺失。
3. **adapter 层丢弃 evidence**：resumeAdapter 显式丢弃 module_name 映射；agentAdapter 未映射 resume_module。
4. **优化页 JDEvidenceChain 完全未展示**：后端 `JDEvidenceChain` 字段（含 JDRequirement / ResumeEvidence / EvidenceStrength / MatchJudgement / ScreeningRisk）前后端全链路断开。
5. **推荐页 top_fix_items 未展示**：后端有完整 TopFixItem[]（含 ResumeModule），前端未渲染。
6. **四页面均无模块标签**：用户无法看到 AI 判断依据属于简历的哪个模块。

## Git 约束确认

本次未执行任何 Git 命令，也未提交代码。
