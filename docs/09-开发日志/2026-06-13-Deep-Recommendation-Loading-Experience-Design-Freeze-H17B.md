# Deep-Recommendation-Loading-Experience-Design-Freeze-H17B 完成报告

> 执行时间：2026-06-13 13:39  
> 类型：设计方案冻结（不修改代码）  
> 状态：✅ 冻结完成，H17-C 可执行

---

## 1. 总体结论

| 项目 | 结果 |
|------|------|
| 是否完成设计冻结 | ✅ 完成 |
| 是否修改业务代码 | ❌ 未修改 |
| 是否执行 Git | ❌ 未执行 |
| 最终选择方案 | **方案 C（纯前端 UX 重构）**，新增 `DeepAgentLoadingOverlay` 专用组件，保留旧组件向后兼容 |
| 是否改后端 | ❌ 不改 |
| 是否改接口 | ❌ 不改 |
| 是否改 DB | ❌ 不改 |
| H17-C 是否可以执行 | ✅ 可以，修改边界已冻结 |

---

## 2. 当前组件引用范围

| 组件 | 引用文件 | 是否只被推荐页使用 | 结论 |
|------|----------|-------------------|------|
| AgentStepLoadingOverlay | `JobRecommendationPage.tsx` | ❌ 共享 | 保留，向后兼容 |
| AgentStepLoadingOverlay | `JobMatchAnalysisPage.tsx` | ❌ 共享 | 不受影响 |
| AgentStepLoadingOverlay | `ResumeComparisonPage.tsx` | ❌ 共享 | 不受影响 |
| AgentStepLoadingOverlay | `ResumeOptimizationPage.tsx` | ❌ 共享 | 不受影响 |

**重构策略**：新增 `DeepAgentLoadingOverlay.tsx` 专用于深度推荐 loading，推荐页切换过去。`AgentStepLoadingOverlay` 保持不变（仍被 3 个其他页面使用）。

---

## 3. 最终 Loading 设计方案

| 模块 | 冻结方案 |
|------|----------|
| 弹窗尺寸 | 480px，min-h 自适应 |
| 圆角 | 24px（rounded-3xl） |
| 背景 | `bg-white/95` + 极浅蓝渐变 `from-white via-blue-50/20 to-white` |
| 边框 | `border border-slate-100 ring-1 ring-black/3` |
| 阴影 | `shadow-2xl shadow-blue-900/5` |
| 遮罩 | `bg-slate-900/15 backdrop-blur-md` |
| 中心动态主视觉 | 96px 蓝色呼吸环（CSS `@keyframes breathe` pulse + 渐变边框旋转 ring）+ 内层 48px 圆点 |
| 标题 | "深度推荐智能体正在分析" `text-[18px] font-bold` |
| 副文案 | "基于简历画像、岗位池和岗位证据进行多 Agent 协同分析" `text-[13px] text-slate-500` |
| 已用时 | `已用时 01:14` `text-[13px] font-mono text-slate-500` + 微弱脉冲点 |
| Agent 链路 | 横向 4 个 pill 标签（shimmer pulse），每 pill 有 tooltip 说明 |
| Top5 岗位矩阵 | 横向 5 个小芯片，统一显示"分析中"+ breathing dot |
| 底部提示 | "深度推荐通常需要几十秒到数分钟，完成后将自动展示结果。" |
| 返回按钮 | "返回快速推荐结果"，底部居中，`text-sm text-slate-400 hover:text-slate-700` |

### 3.1 中心呼吸环 CSS（纯 Tailwind + inline keyframes）

```css
@keyframes breathe {
  0%, 100% { transform: scale(1); opacity: 0.6; }
  50% { transform: scale(1.08); opacity: 1; }
}
@keyframes orbit {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
@keyframes shimmer {
  0%, 100% { background-position: 200% 0; }
  50% { background-position: -200% 0; }
}
```

调用方式：在组件内用 `<style>` 标签或 Tailwind `animate-[breathe_2s_ease-in-out_infinite]`。

---

## 4. Agent 链路设计

| Agent | 展示方式 | 是否真实状态 | 是否允许动效 |
|--------|----------|-------------|-------------|
| 候选发现 | 横向 pill `<span>候选发现</span>` | ❌ 不追踪真实完成 | ✅ shimmer pulse |
| 语义重排 | 横向 pill `<span>语义重排</span>` | ❌ 不追踪真实完成 | ✅ shimmer pulse |
| 逐岗分析 | 横向 pill `<span>逐岗分析</span>` | ❌ 不追踪真实完成 | ✅ shimmer pulse（高亮） |
| 全局校准 | 横向 pill `<span>全局校准</span>` | ❌ 不追踪真实完成 | ✅ shimmer pulse |

**布局**：`flex gap-2 justify-center`，每个 pill `px-3 py-1.5 rounded-full text-[11px] font-medium`。

**逐岗分析** pill 使用略深的蓝色（`bg-blue-100 text-blue-700`），暗示这是当前最重的阶段，但不伪造它已完成。

每个 pill 的 `<title>` 属性包含 tooltip 说明。

---

## 5. Top5 岗位分析矩阵设计

| 子项 | 展示文案 | 状态 | 是否伪造完成 |
|------|----------|------|-------------|
| 岗位 1 | `岗位 1` · `分析中` | 统一 pending breathing dot | ❌ 不伪造 |
| 岗位 2 | `岗位 2` · `分析中` | 统一 pending breathing dot | ❌ 不伪造 |
| 岗位 3 | `岗位 3` · `分析中` | 统一 pending breathing dot | ❌ 不伪造 |
| 岗位 4 | `岗位 4` · `分析中` | 统一 pending breathing dot | ❌ 不伪造 |
| 岗位 5 | `岗位 5` · `分析中` | 统一 pending breathing dot | ❌ 不伪造 |

**布局**：`flex gap-2 justify-center flex-wrap`，每个卡片 68-72px 宽。

**视觉**：`bg-slate-50 border border-slate-100 rounded-xl px-3 py-2 text-center`。

**"分析中" breathing dot**：`w-1.5 h-1.5 rounded-full bg-blue-400 animate-[breathe_1.5s_ease-in-out_infinite]`。

5 个子卡片**不会逐个变成"已完成"**，Deep 完成后 loading overlay 关闭直接展示最终结果。

---

## 6. 删除旧逻辑清单

| 旧逻辑 | 是否删除 | 原因 |
|--------|----------|------|
| 假百分比（`displayPct`） | ❌ 保留在旧组件 | 旧组件仍被 3 页使用，不可删；新组件不展示 |
| 预计耗时（`estimatedTimeText`） | ❌ 保留在旧组件 | 同上 |
| `setTimeout` 自动推进阶段 | ❌ 保留在旧组件 | 同上 |
| `noAutoAdvance` 补丁参数 | ❌ 保留在旧组件 | 向后兼容；新组件不需要此参数 |
| 长列表 pending 主体（13 行灰色步骤） | ❌ 旧组件保留 | 新组件不展示长列表 |
| `deepSteps` 数组（含 children） | ✅ 简化 | 新组件不需要 step 数组，改为固定 Agent 链路展示 |
| `DEEP_ESTIMATED_TIME` 常量 | ✅ 已删除 | H16B 已删 |
| 右侧"快速推荐：可用" | ✅ 已条件隐藏 | H16B 已修复 |

**关键原则**：旧组件 `AgentStepLoadingOverlay` 代码**一行不改**。新组件 `DeepAgentLoadingOverlay` 从零实现，只做深度推荐 loading，干净无债。

---

## 7. Deep 完成态设计

| 区域 | 冻结方案 |
|------|----------|
| 右侧智能分析（capability status） | `{!isDeepDisplay && (` 条件隐藏"快速推荐：可用"，H16B 已实现 |
| 快速推荐并列状态 | 不出现，H16B 已实现 |
| 模式指示器 | "深度推荐" + 绿色圆点 |
| 深度完成文案 | "✅ 深度分析已完成" + agentStatus 说明 |
| 重新生成按钮 | "重新生成深度推荐"，border 样式 |
| pipeline_trace 展示 | **H17-C 不做**（避免 UI 复杂化），留作后续迭代 |

---

## 8. H17-C 实现边界

| 项目 | 结论 |
|------|------|
| **允许修改文件** | 仅 3 个 |
| 1 | `web/src/app/pages/JobRecommendationPage.tsx`（切换引用 + 简化参数） |
| 2 | `web/src/app/components/common/DeepAgentLoadingOverlay.tsx`（**新增**） |
| 3 | `web/src/types/agent.ts`（如果需要类型，可选） |
| **不可修改文件** | 所有其他文件，尤其是旧 `AgentStepLoadingOverlay.tsx` |
| 是否新增组件 | ✅ `DeepAgentLoadingOverlay.tsx`（~200 行） |
| 是否清理旧组件 | ❌ 不改，向后兼容 |
| 是否保留兼容逻辑 | ✅ 旧组件保留全部逻辑 |
| 推荐页改造量 | 替换 `AgentStepLoadingOverlay` 为 `DeepAgentLoadingOverlay`，简化 `deepSteps` 为固定文案 |
| 是否引入新依赖 | ❌ 不引入。CSS 动画用 Tailwind `animate-[...]` 实现 |
| 风险点 | 极低。全新组件，零依赖，不影响旧组件和其他 3 个页面 |

---

## 9. 验收标准（H17-C 完成后必须满足）

1. Deep loading 弹窗中心有明显的 96px 呼吸环动效
2. 没有百分比进度条和百分比数字
3. 没有预计耗时文案
4. 已用时 `已用时 mm:ss` 清晰可见且每秒更新
5. Agent 链路是横向 4 个 pill 标签带 shimmer pulse，不是 13 行灰色静态列表
6. Top5 岗位矩阵 5 个小卡片，统一显示"分析中"+ breathing dot
7. 不伪造逐个岗位完成（5 个卡片不逐个变绿/变完成）
8. Deep 完成后 loading 弹窗关闭，结果正常展示
9. Deep 完成后右侧不显示"快速推荐：可用"并列状态
10. Fast 初始推荐不受影响（点击"生成深度推荐"前）
11. Deep 结果展示不受影响（Top5 列表、中间详情正常）
12. Console 无新增 React Hook / TypeError / ErrorBoundary 错误
13. npm build / go build / go vet / go test 全部通过
14. 旧 `AgentStepLoadingOverlay` 未被修改，其他 3 个页面功能不受影响
15. 点击"返回快速推荐结果" overlay 正常关闭，Deep 请求静默完成但结果不自动展示

---

## 10. 暂停行为审计（取消/返回按钮文案）

| 行为 | 实际情况 | 结论 |
|------|----------|------|
| 点击"返回快速推荐结果" | `setPhase("fast_done")` → overlay 关闭 | ✅ |
| Deep 请求是否仍在后台 | `pollQuality()` 未被 abort，async 继续执行 | ✅ 是 |
| Deep 完成后是否自动展示 | `setDeepResult(vm)` 被调用，但 `phase` 已是 fast_done | ❌ 不自动展示 |
| 是否可重新生成获取结果 | 用户再次点击"生成深度推荐"会重新发起请求 | ✅ 可以 |
| 是否应该写"后台继续分析" | 不应写。结果不会自动展示，用户需要主动重新发起 | ❌ 不可用此文案 |
| **推荐按钮文案** | "返回快速推荐结果" | ✅ 准确 |
| **推荐提示文案** | "深度推荐通常需要几十秒到数分钟，完成后将自动展示结果。" | ✅ 准确（适用于正常流程） |

---

## 11. H17-C 实现文件改造清单

### 11.1 新增文件
```
web/src/app/components/common/DeepAgentLoadingOverlay.tsx
```

### 11.2 修改文件
```
web/src/app/pages/JobRecommendationPage.tsx
```

### 11.3 JobRecommendationPage 改动点

| 改动 | 说明 |
|------|------|
| import 新增 | `import { DeepAgentLoadingOverlay } from "../components/common/DeepAgentLoadingOverlay"` |
| import 可移除 | `type AgentLoadingStep`（不再需要） |
| `deepSteps` 数组 | 可删除（新组件内置 Agent 链路） |
| `deepElapsedSec` + timer | 保留（elapsed 计时逻辑不变） |
| `deepBackendDone` | 保留（控制 overlay 关闭） |
| `<AgentStepLoadingOverlay ...>` | 替换为 `<DeepAgentLoadingOverlay ...>` |
| 新组件 props | `open`, `elapsedSeconds`, `backendDone`, `onFinish`, `onCancel`, `cancelLabel` |

### 11.4 DeepAgentLoadingOverlay 组件结构

```
DeepAgentLoadingOverlay
├── 遮罩层（blur + semi-transparent）
├── 弹窗卡片
│   ├── 中心呼吸环动效（96px CSS animation）
│   ├── 标题 + 副文案
│   ├── 已用时 行
│   ├── Agent 链路（横向 4 pill + shimmer）
│   ├── Top5 岗位矩阵（5 卡片 + breathing dot）
│   ├── 底部提示文字
│   └── 返回按钮
└── maxDuration safety timeout（15min）
```

---

## 12. docs

- 新增：`docs/09-开发日志/2026-06-13-Deep-Recommendation-Loading-Experience-Design-Freeze-H17B.md`
- 更新：`docs/变更记录.md` `docs/下一步任务.md`

---

## 13. 给用户的结论

1. **最终设计方案比现在优秀的原因**：用 96px 中心呼吸环 + 横向 shimmer pill 链路 + breathing dot 岗位矩阵，替代了 1 个 40px spinner + 13 行灰色静态文字。用户一眼就能看到"系统在运行"。

2. **不做后端实时 stage 的原因**：后端单个 HTTP POST 内串行执行 5 层 Deep Cascade，完成才返回。为 loading UX 加异步任务/SSE/轮询端点属于过度工程，当前 100s 耗时适合用纯前端动效表达"等待"而非伪造"进度"。

3. **为什么这种设计是"诚实但有动态感"**：不伪造百分比、不伪造阶段完成、不伪造岗位逐个分析完成。用呼吸环/pulse/shimmer 等纯装饰动效表达"系统正在工作"，而不是表达"完成了多少"。

4. **H17-C 应该新增独立组件**：`DeepAgentLoadingOverlay.tsx` 从零编写（~200 行），不修改旧 `AgentStepLoadingOverlay.tsx`，不影响其他 3 个页面。

5. **H17-C 实现风险极低**：修改范围仅 2-3 个文件，无新依赖，纯 CSS 动画，不改后端。可以立即进入 H17-C 实现。
