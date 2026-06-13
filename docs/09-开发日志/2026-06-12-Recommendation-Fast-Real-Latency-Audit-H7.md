# Recommendation-Fast-Real-Latency-Audit-H7：Fast 推荐真实耗时链路审计

## 审计目标

用户报告点击"快速推荐"后约 30s loading。拆解完整耗时链路。

---

## 1. Fast API 真实耗时（curl 直测）

| resumeId | run1 | run2 | run3 | avg |
|:---:|:---:|:---:|:---:|:---:|
| 1 | 0.045s | 0.044s | 0.043s | **0.044s** |
| 8 | 0.046s | 0.048s | 0.046s | **0.047s** |
| 4 | 0.063s | 0.042s | 0.042s | **0.049s** |

**结论：Fast API 耗时仅 ~40-63ms。后端极快，不是瓶颈。**

---

## 2. 30s 根因：用户点击的不是 Fast，是 Deep

### 按钮与行为对照

| 页面标签 | 右侧按钮文案 | onClick | 耗时 |
|------|------|------|:---:|
| "快速推荐"（fast_done 阶段） | **"生成深度推荐"**（蓝底白字） | `handleDeepRecommendation()` | **1-3 分钟** |
| "深度推荐"（deep_done 阶段） | "重新生成深度推荐" | `handleDeepRecommendation()` | 1-3 分钟 |

**不存在"快速推荐"或"重新生成快速推荐"按钮。**

当用户看到页面上方写着"快速推荐"，下方一个大蓝按钮写着"生成深度推荐"时，用户可能以为是在"快速推荐"状态下重新生成推荐，但实际触发的是 Deep Cascade 全链路（Layer 0-4 + 质量轮询）。

### Deep 耗时构成

`handleDeepRecommendation` 触发后：
1. `setPhase("deep_loading")` → AgentStepLoadingOverlay 打开（5 阶段动画）
2. POST `/api/jobs/recommend?generation_tier=quality` → LLM Cascade 调用
3. 轮询：`MAX_QUALITY_POLL=10 × 2000ms = 最大 20s`
4. 加上 LLM 实际响应时间 → **总计 20-90s**

用户感知的 30s 完全符合 Deep 调用耗时。

---

## 3. Fast 前端链路审计

| 项目 | 当前情况 | 是否导致延迟 |
|------|------|:---:|
| generateFast | API 调用 → profileStatus==="ready" → 立即 setPhase("fast_done") | ❌ 无延迟 |
| API 耗时 | 40-63ms | ❌ 无延迟 |
| 轮询 | 仅当 `profileStatus !== "ready"` 时触发（1800ms×8） | 正常情况不触发 |
| latest/quality 并发 | 无——generateFast 只 POST fast | ❌ 无额外请求 |
| LoadingOverlay | Fast 用 PageLoading 简单文字，非 AgentStepLoadingOverlay | ❌ 无阶段动画延迟 |

**Fast 本身是真正的"秒级轻量"：API 40-63ms + React 渲染 < 100ms。**

---

## 4. 为什么用户感到 30s

| 原因 | 详细说明 |
|------|------|
| 按钮语义模糊 | 页面标签"快速推荐" + 按钮"生成深度推荐" → 用户以为点击的是"快速推荐"更新 |
| 实际触发 Deep | `handleDeepRecommendation()` → 完整 LLM Cascade |
| Deep 耗时 20-90s | AgentStepLoadingOverlay 5 阶段动画 + 轮询 + LLM 调用 |
| 无 Fast 重新生成入口 | 页面没有 "重新生成快速推荐" 按钮 |

---

## 5. 结论

**Fast 后端和前端均无性能问题。30s 不是 Fast 耗时，是 Deep 耗时。** 根本原因是 UI 缺少"重新生成快速推荐"按钮，用户只能点"生成深度推荐"（唯一蓝色 CTA），误以为在重新生成 Fast。

---

## 6. 修复建议（H8）

| 优先级 | 修复 | 位置 |
|:----:|------|------|
| P1 | Fast 模式增加"重新生成快速推荐"按钮 | 右侧 actions 区域 fast_done 分支 |
| P1 | 避免 Fast 唯一 CTA 是"生成深度推荐" | 同上 |

---

## 7. 构建

| 项目 | 结果 |
|------|:---:|
| npm build | ✅ PASS |
| go build | ✅ PASS |

---

## 8. 下一步

- **推荐任务名**：`Recommendation-Fast-CTA-Action-Fix-H8`
- 增加 Fast 重新生成按钮 + 修复按钮标签混淆
