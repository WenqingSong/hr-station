# Job-Match-Full-Chain-Current-State-Audit-H18A 完成报告

> 执行时间：2026-06-13 14:36  
> 类型：只读审计  
> 范围：岗位匹配全链路（前端+API+后端+DB）  
> 状态：✅ 审计完成

---

## 1. 总体结论

| 项目 | 结果 |
|------|------|
| 是否完成审计 | ✅ 完成 |
| 是否修改代码 | ❌ 未修改 |
| 是否执行 Git | ❌ 未执行 |
| **当前岗位匹配最大问题** | **Fast 用原始标签代码匹配，Deep 用 TagProfile 非 SEP；分数体系不一致；缺少 evidence mapping + hard gate** |
| Fast 是否精准 | ❌ 不精准。纯标签计数公式 `60 + tagScore*0.8`，base 60 无意义，matched/highlights 可为 null |
| Deep 是否精准 | ⚠️ 部分精准。LLM 可感知方向错配，但基于 TagProfile 非 SEP，缺少结构化 evidence/coverage |
| 是否使用 SEP Profile | ❌ **未使用**。匹配用 `TagEvidenceService`（标签画像），推荐用 `SemanticProfileGate`（语义画像）——存在系统性割裂 |
| Fast/Deep 分数是否一致 | ❌ 不一致。Fast 公式 60-98（虚假底分），Deep LLM 给出独立估值（实测 50），分数语义冲突 |
| 是否需要重构 | ✅ 需要全面重构 |
| 推荐下一步 | **H18-B：Fast Match 架构冻结 → H18-C：Fast Evidence Engine 后端重构** |

---

## 2. 环境状态

| 项目 | 结果 |
|------|------|
| backend healthz | ✅ 200, database=ok |
| frontend access | ✅ 200 |
| database | ✅ 可用，match_analyses 表存在但无最近数据 |
| go build | ✅ |
| vite build | ✅ 2.16s |

---

## 3. 前端页面定位

| 模块 | 文件 | 说明 |
|------|------|------|
| 岗位匹配页面 | `web/src/app/pages/JobMatchAnalysisPage.tsx` | 主组件，~500 行 |
| Fast 展示 | 同一文件 L375-430 | 匹配结论卡片（score + level + summary + highlights + risks） |
| Deep 展示 | 同一文件 L431+（Detail 组件） | `MatchAnalysisDetail` 组件展示 Deep 结构化报告 |
| Loading | 使用旧的 `AgentStepLoadingOverlay` | 4 步假进度，未升级到 H17 设计 |
| 路由 | `/match?jobId=X&resumeId=Y` | 从 URL params 读取 |
| API 调用 | `loadMatchAnalysisViewModel(jobId, {resume_id, generation_tier})` | POST `/api/jobs/{id}/match` |

**当前页面问题**：
- Fast 展示简单：score + level + summary + highlights 列表 + risks 列表
- Deep 展示较好但有字段缺失（fit_points/gap_points 可能 null）
- Loading 仍用旧 AgentStepLoadingOverlay（假进度，未受益于 H17 改造）
- 右侧侧边栏复用推荐页组件，但能力状态逻辑偏简单

---

## 4. API 契约

| 项目 | 结果 |
|------|------|
| endpoint | `POST /api/jobs/{jobId}/match` |
| method | POST |
| request | `{resume_id: number, generation_tier?: "fast"|"quality"}` |
| response | `{analysis_id, generated_by, result: MatchAnalysisResult, meta}` |
| Fast/Deep 区分 | `generation_tier` 字段，默认 "fast" |
| latest endpoint | `GET /api/jobs/{jobId}/match/latest?resume_id=X` |
| ViewModel | `toMatchAnalysisPageViewModel()` in matchAnalysisAdapter.ts |

**响应字段（MatchAnalysisResult）**：
- `match_score`, `match_level`, `summary`
- `evidence_chain`, `match_highlights`, `risk_points`
- `gap_points`, `optimization_suggestions`, `next_steps`
- `fit_points`（Deep only）, `resume_evidence`（Deep only）
- `generated_by`, `agent_status`, `fallback_reason`

---

## 5. 后端链路定位

| 层 | 文件/函数 | 说明 |
|------|-----------|------|
| route | `server/cmd/` → `/api/jobs/{id}/match` | HTTP POST |
| handler | `server/internal/http/match_handler.go` | 鉴权 + 参数校验 |
| service | `server/internal/service/match_service.go` | `Analyze()` 主入口 |
| Fast engine | `match_service.go:fastMatchWithTags()` | `TagEvidenceService` + `computeTagMatch()` |
| Deep 3-agent | `match_service.go:matchWith3Agents()` | `agent.RunDeepMatch3Agents()` |
| LLM agent | `server/internal/agent/match_agent.go` | 单 agent（fallback 路径） |
| mock | `server/internal/agent/match_mock.go` | `AnalyzeFast()` / `AnalyzeQuality()` |
| model | `server/internal/model/match.go` | `MatchAnalysisResult` 结构体 |
| repo | `server/internal/repository/match_repository.go` | `match_analyses` 表 CRUD |

---

## 6. 当前 Fast 匹配算法

| 维度 | 当前实现 | 问题 | 是否重构 |
|------|----------|------|---------|
| **分数公式** | `min(98, 60 + tagScore * 0.8)` | base 60 无语义意义；tagScore 纯标签命中计数；实际分数 60-98 虚高 | ✅ |
| **方向适配** | 无。不检查 direction 匹配 | 简历后端方向匹配前端岗位也得高分 | ✅ |
| **必备要求** | 无 hard gate | 所有岗位 "screening_result": "pass" | ✅ |
| **技能匹配** | `computeTagMatch()` 标签计数 | 原始 tag code 展示（"project_exp"，非人可读） | ✅ |
| **证据映射** | `evidence_chain`: "未匹配"/"已匹配" | 无真实证据映射，不引用简历内容 | ✅ |
| **风险识别** | risk_points = gap_points（完全相同） | 无差异化风险 vs 差距 | ✅ |
| **分数上限** | 98 硬上限 | 无校准，无法区分"完美"和"不错" | ✅ |
| **highlight** | 可为 null | 当无标签命中时 match_highlights=null ← 严重 UX 问题 | ✅ |

**Fast 实测（r3+j10001）**：score=60, level="建议补强后投递", match_highlights=**null**, risk_points=13 条原始 tag code。

---

## 7. 当前 Deep 匹配链路

| 项目 | 当前实现 | 问题 |
|------|----------|------|
| 是否 multi-agent | ✅ 3-agent（Context + Decision + Verifier） | Agent 1 基于 TagProfile 非 SEP |
| 输入上下文 | `ResumeTagProfile` + `JobTagProfile` | ❌ 非 SEP，缺少 semantic evidence |
| prompt 质量 | `matchAnalysisBusinessContext` 基础 prompt | 缺少 JD 深度解构指令 |
| schema | `SchemaMatchAnalysis` | 输出字段有限 |
| JSON repair | ✅ pipeline 内置 | - |
| fallback | ✅ 3-agent → single agent → mock | 级联安全 |
| JD 解构 | ❌ 无 | Deep 未拆解 JD 的 must_have/should_have/risk |
| 证据映射 | ⚠️ 基于标签 | 非语义证据 |
| 风险诊断 | ✅ gap_points（12 项） | 质量一般 |
| 简历建议 | ⚠️ 基础优化建议 | 非定向到 JD |
| 面试预测 | ❌ 无 interview_talking_points | - |

**Deep 实测（r3+j10001）**：score=50, level="low", 正确识别方向错配。但 fit_points=null, evidence_chain=null。

---

## 8. 数据库结构

| 表 | 作用 | 当前数据 | 是否支持重构 |
|------|------|----------|-------------|
| `match_analyses` | 持久化匹配结果 | 表存在但最近无数据 | ✅ 字段足够（score, generation_tier, result_json） |
| `resume_profiles` | SEP 简历画像 | 有数据 | ✅ 推荐已在用 |
| `job_profiles` | SEP 岗位画像 | 46 verified | ✅ 推荐已在用 |
| `resumes` | 简历主表 | 正常 | ✅ |
| `jobs` | 岗位主表 | 正常 | ✅ |

---

## 9. SEP 使用情况 — 🔴 核心发现

| 项目 | 结果 |
|------|------|
| **Fast 是否读 SEP resume profile** | ❌ 否。用 `TagEvidenceService.EnsureResumeTagProfile()`（标签画像） |
| **Fast 是否读 SEP job profile** | ❌ 否。用 `TagEvidenceService.EnsureJobTagProfile()`（标签画像） |
| **Deep 是否读 SEP resume profile** | ❌ 否。Agent 1 用 `ResumeTagProfile`（标签） |
| **Deep 是否读 SEP job profile** | ❌ 否。Agent 1 用 `JobTagProfile`（标签） |
| **是否存在推荐用 SEP、匹配不用 SEP** | ✅ **是！系统性割裂。推荐有 SemanticProfileGate + Deep Cascade（5 层），匹配只有 TagEvidence** |

**根因**：推荐链路已使用 `SemanticProfileGate.EnsureResumeProfile()` / `EnsureJobProfile()`（语义证据画像），但匹配链路停留在 `TagEvidenceService`（纯标签画像）。这是精准度不够的**根本原因**。

---

## 10. API 实测结果

| 模式 | resumeId | jobId | 耗时 | score | generated_by | level | 核心问题 |
|------|----------|-------|------|-------|-------------|-------|----------|
| Fast | 3 | 10001 | 0.01s | 60 | fast_tag_engine | 建议补强后投递 | highlights=null, raw tag codes, 无 evidence |
| Deep | 3 | 10001 | 34.3s | 50 | llm_agent | low | fit_points=null, 基于 TagProfile 非 SEP |

---

## 11. 精准度问题判断

1. **Fast 不精准的核心原因**：纯标签计数，无方向适配、无 hard gate、无证据映射。即使简历和目标岗位完全不匹配（如后端简历匹配前端岗位），也能拿到 60+ 分，因为 base 就是 60。

2. **Deep 不精准的核心原因**：虽然用了 LLM 可以做方向判断，但输入只是 `TagProfile`（标签列表），不是 `SemanticProfile`（语义证据），导致 LLM 看到的上下文极度有限——只有标签名称，没有 resume/JD 原文证据。

3. **分数体系不一致**：Fast 公式 `60 + tagScore*0.8` 给出 60-98，Deep LLM 独立判断给出独立分数。同一岗位 Fast=60 Deep=50，产品体验割裂。

4. **match_highlights 可为 null**：Fast 场景下若无匹配标签，highlights 直接 null——前端展示"匹配结论"但无任何证据支撑。

5. **risk_points 与 gap_points 完全相同**：Fast 不区分风险和差距，全部混在一起，用户无法区分"硬伤"和"可补"。

6. **无 JD requirement coverage**：当前匹配不拆解 JD 的 must_have/should_have/nice_to_have，无法给出具体覆盖度。

7. **缺少 hard gate**（方向 gate、学历 gate、技能 gate）：所有岗位 screening_result 都是 "pass"，没有真正筛掉明显不合适的。

8. **evidence_chain 生成质量低下**：Fast 的 evidence 都是 "未匹配"/"已匹配" 二元判断，不引用任何简历内容。Deep 的 evidence_chain 直接为 null。

9. **tag code 泄露到前端**：risk_points、gap_points 包含 `project_exp`、`docker_k8s` 等原始代码，用户无法理解。

10. **推荐用 SEP，匹配不用 SEP，是最大的架构割裂**。推荐系统的 Deep Cascade 已实现 5 层语义分析，而匹配系统停留在标签层面。

---

## 12. 重构建议

| 阶段 | 目标 | 是否必要 | 优先级 |
|------|------|---------|-------|
| **H18-B** | Fast Match 架构冻结（设计 Evidence-Grounded Fast Match Engine 规范） | ✅ 必须 | P0 |
| **H18-C** | Fast Match Engine 后端重构（接入 SEP Profile + Hard Gate + Evidence Mapping + 新分数） | ✅ 必须 | P0 |
| **H18-D** | Fast Match 前端展示重构（结构化 evidence 展示 + tag humanize + score 说明） | ✅ 必须 | P0 |
| **H18-E** | Fast Match 验收 | ✅ 必须 | P0 |
| **H19-A** | Deep Match 架构冻结（Multi-Agent Match Consultant 设计） | ✅ 必须 | P1 |
| **H19-B** | Deep Match Agent Schema/Prompt 重构（接入 SEP + JD Deconstruction + Evidence Report） | ✅ 必须 | P1 |
| **H19-C** | Deep Match 前端报告页重构（结构化 Agent Report 页面） | ✅ 必须 | P1 |
| **H19-D** | Deep Match 验收 | ✅ 必须 | P1 |

**建议先修 Fast**：因为 Fast 是用户最常见的交互（秒级返回），且当前分数虚高问题最影响用户信任。

**建议先统一数据契约**：Fast 和 Deep 应共享同一套分数语义（如 coverage/evidence/risk 三维），避免评分矛盾。

**建议先接入 SEP**：匹配应使用 `SemanticProfileGate` 替代 `TagEvidenceService`，这是精准度提升的最短路径。

**最大风险**：SEP Profile 的 job_profiles 目前只有 46 个 verified，若目标岗位不在池中需先构建。建议重构时增加 lazy profile 构建。

---

## 13. docs

- 新增：`docs/09-开发日志/2026-06-13-Job-Match-Full-Chain-Current-State-Audit-H18A.md`
- 更新：`docs/变更记录.md`、`docs/下一步任务.md`

---

## 14. 给用户的结论

1. **当前岗位匹配达不到用户和 HR 预期的根本原因有三**：(1) Fast 是纯标签计数公式 `60 + tagScore*0.8`，base 60 无意义且 highlights 可为 null；(2) Deep 基于 TagProfile 而非 SEP SemanticProfile，输入上下文不足；(3) 推荐已用 SEP，匹配不用 SEP——系统性割裂。

2. **Fast（标签）和 Deep（LLM）使用完全不同的方法论**，分数语义不一致。Fast 虚高（公式底分 60），Deep 偏向保守（LLM 独立估值），同一岗位得分无法比较。

3. **不能只改页面文案**：当前问题深度在于算法层（缺 SEP、缺 evidence mapping、缺 hard gate、缺 JD deconstruction），前端展示只是冰山一角。

4. **第一阶段应先修 Fast Match Engine**：接入 SEP SemanticProfile → 引入 Hard Gate（方向/学历/技能）→ 建立 evidence mapping（非标签计数）→ 重构分数公式 → 输出结构化 evidence/risk/coverage。这是提升用户信任的最短路径。

5. **第二阶段再修 Deep Match Consultant**：基于 SEP 输入 → JD 深度解构 → Multi-Agent 证据链 → 结构化 Match Report（coverage matrix + evidence mapping + risk diagnosis + interview prediction）。

6. **loading 体验也需要升级**：匹配页目前仍用旧 `AgentStepLoadingOverlay`（假进度），可借鉴 H17 的 `DeepAgentLoadingOverlay` 设计。
