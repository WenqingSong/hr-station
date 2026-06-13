# Deep-Recommendation-Agent-Status-Flow-Audit-H16A 完成报告

> 执行时间：2026-06-13 12:49  
> 类型：只读审计  
> 范围：前端 JobRecommendationPage + AgentStepLoadingOverlay + 后端 Deep Cascade Pipeline  
> 状态：✅ 完成

---

## 1. 总体结论

| 项目 | 结果 |
|------|------|
| 是否完成审计 | ✅ 完成 |
| 是否修改业务代码 | ❌ 未修改 |
| 是否执行 Git | ❌ 未执行 |
| 是否找到右侧智能分析模块 | ✅ 找到 |
| 是否找到深度推荐 loading 组件 | ✅ 找到 |
| 是否存在假百分比/预计耗时 | ✅ 确认存在 |
| 后端是否有 task_id | ❌ 无 |
| 后端是否有 stage/progress | ❌ 无 |
| 是否支持逐岗 5 个子任务真实状态 | ❌ 不支持实时推送 |
| H16-B 推荐方案 | **方案 C**（安全 UX 改造） |
| 是否可以进入 H16-B | ✅ 可以 |

---

## 2. 前端文件定位

| 模块 | 文件 | 说明 |
|------|------|------|
| 岗位推荐页 | `web/src/app/pages/JobRecommendationPage.tsx` | 781行，主组件 |
| 右侧智能分析 | 同上，第 649-762 行 | 内联 JSX，`<aside>` 块 |
| 深度推荐 loading | `web/src/app/components/common/AgentStepLoadingOverlay.tsx` | 347行，独立组件 |
| 推荐 API 调用 | `web/src/api/viewModels.ts` -> `web/src/api/agents.ts` | loadRecommendationViewModel → runRecommendation |
| Loading steps 数据 | `JobRecommendationPage.tsx` 第 93-99 行 | 硬编码 5 步骤数组 |
| 推荐适配器 | `web/src/adapters/recommendationAdapter.ts` | 346行，数据转换 |

---

## 3. 右侧智能分析审计

| 项目 | 当前实现 | 问题 | H16-B 建议 |
|------|----------|------|-----------|
| "快速推荐：可用" | 第 657 行硬编码文本 `<span>快速推荐：可用</span>` | 永远显示，深度完成后也在 | 条件渲染：`!isDeepDisplay &&` |
| "深度推荐：可用" | 第 662 行根据 `phase` 判断 | 显示逻辑正确 | 保留 |
| displayMode 判断 | `activeTier` + `isDeepDisplay = activeResult === deepResult && deepReady` | 逻辑正确但未联动右侧 | 联动隐藏 fast 可用态 |
| 深度完成态 | 第 687-694 行显示"深度分析已完成" | 正确 | 保留 |
| 底部提示文字 | 第 756-760 行根据相位显示不同文案 | 正确 | 保留 |

**问题确认**：深度完成后（`deep_done` + `isDeepDisplay=true`），右侧仍然显示：
- "快速推荐：可用"（硬编码，未被 `isDeepDisplay` 控制）
- "深度推荐：可用"（也显示，两个并列）

**H16-B 修复**：仅需在第 655 行 `快速推荐：可用` 外包 `{!isDeepDisplay && (` 条件。

---

## 4. loading 弹窗审计

| 项目 | 当前实现 | 是否真实 | 问题 |
|------|----------|----------|------|
| 百分比进度条 | `Math.min(96, Math.floor((displayElapsed / 150) * 96))` - 纯时间驱动 | ❌ 假 | 150秒到96%，毫无意义 |
| 预计耗时 | 硬编码常量 `"预计耗时约 1-3 分钟；首次构建岗位画像时可能更久。"` | ❌ 假 | 写死文字 |
| 已用时 | `setInterval(1000)` 计时器 | ✅ 真实 | 计数正确 |
| 阶段列表 | 硬编码 5 步：准备语义画像/候选岗位发现/语义重排/逐岗深度分析/全局一致性校准 | ⚠️ 后端有对应代码但前端不接收状态 | 标签名匹配后端层级 |
| 阶段完成状态 | `setTimeout` 每 650ms 自动推进 1 步，永不自动完成最后一步 | ❌ 假 | 纯前端动画 |
| timer/setInterval | 有清理（useEffect return），关闭页面可清理 | ✅ 安全 | - |
| 免责声明 | 第 332 行已有 `"当前为预估进度，最终结果以分析完成后为准"` | ✅ 诚实 | - |

---

## 5. 前端 API 审计

| 项目 | 结果 |
|------|------|
| Deep 是否同步等待最终响应 | ❌ 不是。前端用 `setTimeout` 轮询，检查 `profileStatus === "ready"` |
| 是否有 task_id | ❌ 无。前端不接收也不存储 task_id |
| 是否轮询 stage | ❌ 不轮询 stage。只轮询 `profileStatus` |
| 是否 SSE/EventSource | ❌ 无 |
| 是否能接收 per-job 状态 | ❌ 无接口 |
| 轮询最大次数 | 10 次 (MAX_QUALITY_POLL)，每次 2s 间隔 |
| 完整 Deep 请求耗时（实测） | 103.6秒 (HTTP 200) |

---

## 6. 后端 Deep 推荐审计

| 项目 | 结果 |
|------|------|
| Deep 入口函数 | `RecommendationService.recommendCore()` → `recommendDeepCascade()` |
| 是否同步生成 | ✅ 是。单次 HTTP POST，内部阻塞直到所有层完成 |
| 是否分阶段执行 | ✅ 是。5 层级联：L0 Gate → L1 候选发现 → L2 语义重排 → L3 逐岗合成×5 → L4 全局校准 |
| 是否候选发现 Agent | ✅ `runLayer1Discovery()` - Hy3 LLM ×1 |
| 是否语义重排 Agent | ✅ `runLayer2Rerank()` - Hy3 LLM ×1 |
| 是否逐岗分析 Agent | ✅ `runLayer3Synthesis()` - Hy3 LLM ×5 (for-loop) |
| 是否全局校准 Agent | ✅ `runLayer4Calibration()` - Hy3 LLM ×1 |
| 是否 per-job loop | ✅ `for _, tj := range top5 { runLayer3Synthesis(...) }` |
| 是否存 task/status | ❌ 不存中间 task。只存最终 `recommendation_results` |
| Trace 数据 | ✅ 生成 `[]DeepCascadeStep` trace 数组，存于 `pipeline_trace` 字段 |
| 是否有中间状态推送 | ❌ 无。所有 trace 只在最终响应中一并返回 |
| Agent 状态字段 | `agent_status: "cascade_verified"` (在 result 和每个 job 上) |

---

## 7. Deep API 响应结构（实测）

| 字段 | 是否存在 | 说明 |
|------|----------|------|
| task_id | ❌ 无 | 无异步任务机制 |
| status | ❌ 无（只有 profile_status） | profile_status="verified" |
| current_stage | ❌ 无 | 无阶段字段 |
| progress | ❌ 无 | 无进度字段 |
| generation_tier | ✅ | "quality" |
| pipeline_mode | ✅ | "cascade" |
| pipeline_trace | ✅ | JSON 数组，含 7 个 DeepCascadeStep（L0+L1+L2+L3×5+L4），每步有 layer/status/duration_ms/tokens/detail |
| quality_status | ✅ | "cascade_verified" |
| recommended_jobs | ✅ | 5 个完整岗位 |
| per-job agent_status | ✅ | "cascade_verified" |
| per-job analysis | ✅ | 每个岗位有 why_ranked_here, resume_evidence, job_specific_reason, job_specific_risks, candidate_core_strength, fit_confidence_reason 等 |
| global_summary | ✅ | L4 全局摘要 |
| candidate_positioning | ✅ | L4 候选人定位 |

---

## 8. 逐岗分析 5 子任务可行性

| 判断项 | 结果 |
|------|------|
| Top5 是否固定 | ✅ 是。Layer 2 输出严格 5 个 TopJobs |
| 后端是否逐个分析岗位 | ✅ 是。`for _, tj := range top5 { runLayer3Synthesis(...) }` 依次调用 LLM |
| 是否能实时回传每个岗位完成 | ❌ 不能。后端在单个 HTTP 请求内串行执行全部层，完成才返回 |
| 是否可以前端真实点亮 5 子任务 | ❌ 不能。后端无中间状态推送机制 |
| 如果不能，最小安全方案 | **方案 C**：删除假进度条，改为"已提交分析任务，等待模型返回"，5 子项只显示为"待分析/统一生成中"，不逐个点亮 |

**关键结论**：
- 后端 `recommend_deep_cascade.go` 确实按 5 层 Agent 阶段执行（L0-L4），L3 也确实逐岗位调用 LLM
- 但所有执行在单个 HTTP POST 请求内串行完成（context.WithTimeout 900s）
- 没有任何中间状态同步给前端，前端轮询也只能等到 profile_status 变成 "ready" 后才拿到全部结果
- `pipeline_trace` 在最后一次返回时才包含所有层完成记录

---

## 9. H16-B 实现建议

### 选择：**方案 C（安全 UX 改造）**

**理由**：后端虽是同步 Deep Cascade，但**没有异步任务机制、没有中间状态推送、没有 SSE、没有轮询端点**。为了 UI 加这些基础设施属于过度工程。

**具体建议**：

| 改动项 | 说明 |
|--------|------|
| 删除假百分比进度条 | 删除 `AgentStepLoadingOverlay` 中 L3 进度条区块（第 277-289 行） |
| 删除预计耗时文字 | 不传 `estimatedTimeText` prop，或改为"正在调用多阶段 AI 分析" |
| 保留已用时 | `已用 mm:ss` 保留，真实有用 |
| 保留阶段列表 | 保留 5 步标签，改为"等待模型返回"所有 pending 态 |
| 不逐个点亮 5 子任务 | 所有步骤保持 pending，只在 backendDone 时全部变 done |
| 隐藏右侧"快速推荐：可用" | 深度完成后条件隐藏 |
| 不后端改造 | 不新增 task 表、不新增轮询端点、不引入 SSE |
| 前端改造量 | 约 20-30 行改动 |

**改动不涉及**：
- ❌ 后端新增 task_id
- ❌ 新增 stage 状态接口
- ❌ 新增前端轮询
- ❌ SSE/EventSource
- ❌ DB schema 变更

---

## 10. 构建检查

| 项目 | 结果 |
|------|------|
| go build | ✅ 通过 |
| go vet | ✅ 通过 |
| go test | ✅ 通过 (service, schema, agent cached) |
| npm/tsc | ⚠️ 有预存 TypeError（ResumeComparisonPage.tsx, ResumeOptimizationPage.tsx 等），但与推荐页无关 |

---

## 11. docs

- 新增：`docs/09-开发日志/2026-06-13-Deep-Recommendation-Agent-Status-Flow-Audit-H16A.md`
- 更新：`docs/变更记录.md` `docs/下一步任务.md`（需后续 H16-B 时更新）

---

## 12. 给用户的结论

1. **当前深度推荐 loading 看起来假**是因为：百分比进度条由纯前端 `setInterval` + 时间公式 `Math.floor((elapsed/150)*96)` 生成，阶段步骤由 650ms `setTimeout` 自动推进，完全不是从后端接收的真实状态。

2. **后端确实有多 Agent 阶段**：Deep Cascade Pipeline 有 5 层（Layer 0-4），L3 确实逐个岗位调用 LLM（for-loop ×5），所有层完成会生成 `pipeline_trace`。

3. **但后端阶段状态不会实时推送**：所有层在单个 HTTP POST 请求内串行执行，完成才一次性返回，前端无法逐阶段逐岗位接收。

4. **逐岗 5 个子加载不能做真实同步**：因为后端没有异步任务机制、没有 task_id、没有 SSE、没有轮询端点。即便做轮询也只能等到"完成"或"未完成"两个状态。

5. **H16-B 最稳方案是方案 C**：只做前端 UX 改造——删除假百分比和预计耗时，保留已用时和 Agent 阶段说明，确保不伪造中间状态。同时修复右侧"智能分析"模块在深度完成后仍显示"快速推荐：可用"的问题。

6. **不建议为了 UI 改动后端**：新增 task 表、异步任务、SSE 或阶段轮询端点，会显著增加系统复杂度，与当前产品阶段不匹配。当深度推荐耗时稳定在 1-2 分钟时，"等待最终结果"的 UX 是足够的。

7. **H16-B 可以立即进入**，预计改动量约 30 行前端代码。
