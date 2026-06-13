# Deep-Recommendation-Loading-FullChain-Audit-H17A 完成报告

> 执行时间：2026-06-13 13:26  
> 类型：只读审计  
> 范围：AgentStepLoadingOverlay + JobRecommendationPage + Deep API  
> 状态：✅ 审计完成

---

## 1. 总体结论

| 项目 | 结果 |
|------|------|
| 是否完成审计 | ✅ 完成 |
| 是否修改代码 | ❌ 未修改 |
| 是否执行 Git | ❌ 未执行 |
| 当前"都不动"的根因 | **UX 设计问题，非 Bug**：95% 视觉空间是静态灰字，仅 40px spinner 有动效 |
| spinner 是否真的不转 | ❌ **spinner 在转**（`animate-spin` 类存在，Tailwind 内置 support） |
| elapsed 是否更新 | ✅ **在更新**（`setInterval` 每秒更新 `displayElapsed` state） |
| Agent steps 是否真实状态 | ❌ 不真实。`noAutoAdvance=true` 下全部 pending，静态无推进 |
| Deep API 是否正常 | ✅ 正常（200, code=0, cascade_verified, 5 items, ~115s） |
| 后端是否支持实时 stage | ❌ 不支持实时推送。但响应中包含完整 `pipeline_trace`（9 步含逐岗耗时） |
| H17-B 推荐方案 | **方案 C**（纯前端高级 UX 重构） |
| 是否可以进入 H17-B | ✅ 可以 |

---

## 2. 前端定位

| 模块 | 文件 | 关键状态/函数 |
|------|------|-------------|
| 推荐页 | `web/src/app/pages/JobRecommendationPage.tsx` | `phase`, `deepBackendDone`, `handleDeepRecommendation()` |
| loading overlay | `web/src/app/components/common/AgentStepLoadingOverlay.tsx` | `noAutoAdvance=true`, `finished`, `displayElapsed`, `stepStatus()` |
| Deep 请求 | `handleDeepRecommendation()` (L262) → `loadRecommendationViewModel()` → `runRecommendation()` → `POST /api/jobs/recommend` | async `pollQuality()` 轮询 |
| 右侧状态 | 同一文件 L649-773 | `isDeepDisplay`, `phase`, `deepResult` |

---

## 3. Loading 行为审计

| 项目 | 当前实现 | 是否真实 | 问题 |
|------|----------|----------|------|
| spinner | `Loader2 w-10 h-10 animate-spin` (L300) | ✅ 真实 CSS 动画 | 占模态框面积 ~2%，淹没在文字中 |
| 已用时 | `setInterval` 每秒更新 `displayElapsed` (L234) | ✅ 真实计时 | 文字太小（`text-[11px] text-slate-400`），不显眼 |
| Agent steps | `noAutoAdvance=true` → 全部 `pending` (L268-270) | ✅ 诚实但静态 | 4 Agent + 5 子任务共 9 行灰字灰圈，零动态 |
| 5 个岗位子任务 | `children` 渲染，全 pending (L361-378) | ✅ 诚实但无意义 | 所有子任务显示相同灰色圆圈，用户感到"什么都没发生" |
| noAutoAdvance | `true` (L785) | ✅ 正确关闭假进度 | 副作用：无任何推进视觉反馈 |
| backendDone | `false` 直到 API 返回 (L275) | ✅ 正确 | Deep ~115s，用户需等 2 分钟看静态界面 |
| 返回快速推荐 | `onCancel` → `setPhase("fast_done")` (L788) | ✅ 功能正常 | 文字在底部 `text-xs underline`，不够显眼 |

---

## 4. "不动"的根因判断

| 类型 | 是否命中 | 证据 |
|------|----------|------|
| A. spinner 动画缺失 | ❌ 否 | `animate-spin` 类存在，Tailwind 4.1.12 内置支持 |
| B. elapsed 不更新 | ❌ 否 | `setInterval` L234 每秒触发 `setDisplayElapsed()` |
| **C. steps 全 pending 导致静态** | ✅ **主因** | `noAutoAdvance=true` → `stepStatus()` 返回 `"pending"` 给所有 13 个 item |
| D. Deep 请求阻塞主线程 | ❌ 否 | `async/await`，HTTP POST 不阻塞 React 渲染 |
| **E. 视觉设计太弱** | ✅ **主因** | 420px 模态框内仅 1 个 40px spinner 有动画，其余全部为 `text-slate-300/400` 静态灰字 |

**最终根因**：C + E 叠加。`noAutoAdvance` 关掉了假进度推进（正确），但替代设计是纯静态文案列表。用户等待 ~2 分钟看 13 行灰色文字 + 1 个小转圈，体验自然"都不动了"。

另外值得注意：父组件 `deepElapsedSec` timer（`setInterval(() => setDeepElapsedSec(s => s+1), 1000)` L268）和 overlay 内部 `displayElapsed` timer（L234）是两个独立计时器。虽然两者都在工作，但 overlay 从 `externalElapsed` 初始化时会触发额外的 `useEffect` 重新运行。不过这不影响显示正确性。

---

## 5. 后端 Deep 链路

| 项目 | 结果 |
|------|------|
| 是否单 POST 同步等待 | ✅ 是 |
| 是否有 task_id | ❌ 无 |
| 是否有 stage/progress | ❌ 无 |
| **是否有 pipeline_trace** | ✅ **有！** 9 步含逐岗 job_id + 耗时 |
| 是否 L3 逐岗循环 | ✅ `for _, tj := range top5 { runLayer3Synthesis(...) }` |
| 是否可实时推送 | ❌ 不可。所有 trace 仅在最终响应中返回 |

**实测 pipeline_trace（resumeId=3）：**

| 层级 | 状态 | 耗时 | 详情 |
|------|------|------|------|
| 0_gate | ok | 30ms | resume=3 tags=15 jobs_ready=46 |
| 1_discovery | ok | 18,532ms | - |
| 2_rerank | ok | 25,052ms | - |
| 3_synthesis | ok | 15,275ms | job=10016 |
| 3_synthesis | ok | 14,295ms | job=10037 |
| 3_synthesis | ok | 10,533ms | job=10014 |
| 3_synthesis | ok | 12,625ms | job=10007 |
| 3_synthesis | ok | 11,873ms | job=10044 |
| 4_calibration | ok | 6,460ms | - |

> 注：pipeline_trace 虽然在最终响应中可用，但无法在 loading 期间实时使用。可作为完成后的"分析履历"展示。

---

## 6. Deep API 实测

| 项目 | 结果 |
|------|------|
| resumeId | 3 |
| total time | 114.7s |
| HTTP/code | 200 / 0 |
| generated_by | llm_agent |
| agent_status | cascade_verified |
| items | 5 |
| pipeline_trace | ✅ 9 步 |
| 后端错误 | ❌ 无 |

---

## 7. UX 问题清单

| 等级 | 问题 | 说明 | 是否阻塞 |
|------|------|------|---------|
| P0 | - | 无功能不可用问题 | - |
| **P1** | 弹窗 95% 视觉空间完全静态 | 仅 40px spinner 有动画，用户感知为"卡住" | 影响用户体验 |
| **P1** | 13 行步骤全部灰色 pending | 无"正在工作中"的视觉信号，视觉疲劳 | 影响感知 |
| **P1** | 缺少动态呼吸/扫描/脉冲效果 | 现代化 loading 应有节奏感，而非纯静态列表 | 影响高级感 |
| P2 | spinner 相对模态框太小 | 420px 宽模态框，40px spinner 只占 9.5% 宽 | 优化建议 |
| P2 | 已用时文字太小太淡 | `text-[11px] text-slate-400`，淹没在大量文字中 | 优化建议 |
| P2 | 5 个岗位子任务在 loading 期间无意义 | 全部 pending 灰字，用户不知"5 岗位"正在被分析 | 优化建议 |
| P2 | 文案偏工程化 | "候选发现 Agent"、"语义重排 Agent"对普通用户不够友好 | 优化建议 |

---

## 8. H17-B 建议方案

### 选择：**方案 C（纯前端高级 UX 重构）**

### 理由：

1. 后端没有实时 stage/task 接口，改后端为 loading UX 加异步任务属于过度工程
2. 当前问题 100% 是 UX 设计问题，不是功能缺位
3. `pipeline_trace` 可在结果完成后作为"分析履历"展示，提升专业感
4. 纯前端重构无后端风险，修改范围可控

### 修改范围：

| 文件 | 改动 |
|------|------|
| `AgentStepLoadingOverlay.tsx` | 新增 `agentMode` prop → 通用优雅 loading 模板 |
| `JobRecommendationPage.tsx` | 调整调用参数，利用 pipeline_trace 做完成后的"分析履历" |

### 是否改后端：❌ 不改

### 是否改接口：❌ 不改

### 是否改 DB：❌ 不改

### 设计原则：

1. **中心大尺寸动态视觉**（替代 40px spinner）：
   - 80-100px 呼吸脉冲圆环（pulsing ring with breathing effect）
   - 或 4 个 Agent 图标聚合的 scanning orbit 动效
   - 或大脑/芯片类 SVG 逐层点亮动画

2. **节奏层次**：
   - L1：中心大型动画（最抢眼）
   - L2：已用时 + 状态摘要（1 行显著文字）
   - L3：Agent 链路标签（缩小、淡化、仅说明性质，不强吸引注意）

3. **诚实表达**：
   - 显示"深度推荐智能体正在多阶段分析中"
   - 不伪造百分比、不假装阶段完成
   - 用动效的"活力感"替代假进度的"推进感"

4. **简化的步骤展示**：
   - 将 13 行（4 Agent + 5 子 + 4 描述）精简为 4 行 Agent 标签 + 1 行子任务说明
   - 用 `animate-pulse` 或微弱的 bg 闪动代替绿色 check
   - 5 个子任务默认收缩，hover/click 展开

5. **5 子任务改为更有意义的状态**：
   - "Top5 岗位分析矩阵 · 统一分析中"
   - 在 Deep 完成后可展开展示每个岗位的 pipeline_trace 耗时

6. **更强的取消/返回入口**：
   - 不再用底部 underline 小字，改为更明显的按钮或卡片

### 预期效果：

- 用户一眼看到中心大型动效 → "系统在工作"
- 已用时清晰可读 → "有进展反馈"
- Agent 步骤简化淡化 → "知道在做什么，但不会被静态文字干扰"
- 完成后展示 pipeline 分析履历 → "专业感"

### 风险：

- 纯前端 CSS/SVG 动画，无后端依赖风险
- 不新增状态变量，无 React Hook 风险
- 改动集中在一个通用组件内，不影响其他页面

---

## 9. 构建检查

| 项目 | 结果 |
|------|------|
| vite build | ✅ 1.92s |
| go build | ✅ 通过 |
| go vet | ✅ 通过 |
| go test | ✅ 通过（model/prompt/schema/service cached） |

---

## 10. 给用户的结论

1. **"都不动了"不是 bug，是 UX 设计问题**：spinner 确实在转（`animate-spin` 正常），计时器也在更新（`displayElapsed` 每秒渲染），但模态框 95% 的视觉空间是静态灰色文字，仅 1 个 40px 小转圈有动效。

2. **H16B 的 `noAutoAdvance=true` 在正确方向上前进了一步**（删除了假百分比），但替代品设计不足——13 行 pending 灰色步骤 + 1 个 spinner = 用户感觉系统卡死。

3. **后端没有实时 stage，也不应该为 loading UX 加异步任务系统**。实测 114s 完成 5 层 Deep Cascade，pipeline_trace 可在响应后用于完成态"分析履历"展示。

4. **H17-B 应选方案 C：纯前端高级 UX 重构**。核心思路是用"动态活力感"替代"伪进度推进感"——中心大型呼吸脉冲/扫描动效 + 精简步骤 + 保留真实已用时。

5. **后端 pipeline_trace 是个宝**：9 步逐岗耗时数据可在 Deep 完成后作为"分析履历"展示，既提升专业感又不伪造中间状态。

6. H17-B 预计修改 2 个文件（AgentStepLoadingOverlay.tsx + JobRecommendationPage.tsx），不改后端。可以立即进入。
