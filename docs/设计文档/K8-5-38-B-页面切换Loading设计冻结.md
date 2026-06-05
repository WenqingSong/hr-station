# K8-5-38-B：页面切换 Loading 设计冻结

> 日期：2026-06-05 | 性质：设计冻结 | 不改源码

---

## 一、K8-5-37 Overlay 方案失败原因

K8-5-37 系列（A→B→C→D）实现了 `PageTransitionOverlay`：页面路由切换时，在 main 内容区上方覆盖一个半透明遮罩 + 中央白色加载胶囊。

经过 C→D 两轮视觉精修后，用户浏览器验收仍认为不优雅。根因分析：

| 失败维度 | 原因 |
|----------|------|
| **设计模型错误** | Overlay 适合模态阻塞（付款中/文件处理中），不适合普通 SPA 页面跳转 |
| **三层叠加** | RouteProgressBar（顶部蓝条）+ PageTransitionOverlay（中央胶囊）+ 页面 Skeleton（底部）= 3 层同时渲染 |
| **触发不真实** | location-change 驱动的固定 90/260ms 延迟，不反映真实数据加载状态 — 像"演戏" |
| **认知冲突** | 遮罩覆盖旧页面 → 用户感觉"被卡住了"而非"在过渡" |
| **业界不通用** | GitHub/Linear/Notion/Stripe/Vercel 等专业 SaaS 在普通页面跳转时全部使用顶部进度条，无一家用中央遮罩胶囊 |

---

## 二、为什么普通页面跳转不适合中央 Overlay

```
1. SPA 路由切换 ≠ 系统级阻塞
   - 路由切换是前端行为，不需要"阻止用户操作"的遮罩
   - 页面跳转是自然的、高频的，不该每次都像"系统卡住"

2. 每个页面已有自身 skeleton
   - JobCardSkeleton、DetailSkeleton、ResumeSkeleton、PageLoading 等
   - Overlay 叠在 skeleton 上是视觉冗余

3. 无法获取真实 pending
   - 当前是 createBrowserRouter 非 Data Router
   - useNavigation().state 不可用
   - 固定延迟的 overlay 是"伪造的 loading"

4. 用户心智模型
   - Overlay = "阻塞等待"（焦虑）
   - 进度条 + 骨架 = "自然过渡"（期待）
```

---

## 三、新 Loading 分工冻结

| 场景 | 组件 | 结论 | 理由 |
|------|------|:--:|------|
| 应用启动 / Auth Bootstrap | `AppBootLoading` | ✅ 保留 | 全屏品牌体验，唯一使用场景 |
| **普通页面跳转** | **`RouteProgressBar`** | ✅ **唯一全局反馈** | 顶部 2px，轻量不打断 |
| ~~普通页面跳转中央遮罩~~ | ~~`PageTransitionOverlay`~~ | ❌ **停用** | 模型不适合，与 skeleton 冲突 |
| 页面数据加载 | `PageLoading` / `Skeleton` | ✅ 保留 | 页面级数据加载，不涉及路由跳转 |
| AI 生成中 | `AIThinkingLoading` | ✅ 保留 | 专属 AI 场景，卡片式步骤 |
| 按钮提交 | `Loader2` / 按钮 loading | ✅ 保留 | 交互反馈 |
| 错误状态 | `PageError` | ✅ 保留 | — |
| 空状态 | `PageEmpty` | ✅ 保留 | — |

---

## 四、PageTransitionOverlay 处理策略

| 项 | 处理 |
|----|------|
| 组件文件 | 保留 `web/src/app/components/common/PageTransitionOverlay.tsx`，不删除 |
| 组件头部 | 添加 `@deprecated` 注释，标记"当前不启用，保留供未来可能场景" |
| 挂载点 | 从 `RootLayout`（`routes.tsx`）中移除渲染 |
| 引用检查 | 确保无其他文件引用（仅 `routes.tsx` 引用） |
| 文档 | 在变更记录中标记为 deprecated |

---

## 五、RouteProgressBar 设计标准

| 参数 | 标准 |
|------|------|
| 高度 | 2px |
| 颜色 | `#0052D9` |
| 透明度 | 0.8 |
| 位置 | `fixed top-0 left-0 right-0` |
| z-index | `z-[9999]` |
| 动画 | 0→70%（30ms）→100%（180ms）→隐藏（380ms） |
| 总时长 | ~380ms |
| 触发 | `useLocation()` 监听 `pathname` + `search` |
| 交互 | `pointer-events-none` |

**下一阶段不需要修改 RouteProgressBar**，当前参数已满足标准。

---

## 六、页面 Skeleton / PageLoading 边界

保留不变：

```
岗位列表：    JobCardSkeleton（5 张卡片骨架）
岗位详情：    DetailSkeleton + Banner PageLoading
岗位推荐：    PageLoading("正在加载…")
岗位匹配：    PageLoading ×2 + AIThinkingLoading
简历优化：    PageLoading ×2 + AIThinkingLoading
版本对比：    PageLoading("正在加载…") + PageError/PageEmpty
我的简历：    PageLoading("正在加载…") + PageError
编辑简历：    ResumeSkeleton（全页骨架） + PageError
```

Skeleton 是页面数据加载状态，不是路由跳转状态，必须保留。

---

## 七、页面内容过渡（可选增强）

下阶段可选增强方向（不做强制要求）：

```
页面内容淡入：
  opacity 0 → 1
  translate-y: 4px → 0
  duration: 160ms - 220ms
  ease-out
```

但约束：
- 如果 RootLayout 能统一包裹 Outlet 实现，优先统一方案
- 如果做不到统一，不做单个页面改造
- 不要为了 fade-in 大改所有页面

**下一阶段优先不做 fade-in**，先确保 RouteProgressBar 作为主反馈稳定后再考虑。

---

## 八、下一阶段实现边界

### K8-5-38-C：移除页面跳转 Overlay 并冻结 RouteProgressBar 跳转反馈

#### 允许修改

```text
web/src/app/routes.tsx                            (1 处：移除 PageTransitionOverlay 导入和渲染)
web/src/app/components/common/PageTransitionOverlay.tsx  (1 处：头部加 @deprecated 注释)
docs/09-开发日志/*
docs/下一步任务.md
docs/变更记录.md
```

#### 禁止修改

```text
不新增新 loading 组件
不改业务页面（JobListPage/JobDetailPage/...）
不改 AI 页面（JobRecommendationPage/JobMatchAnalysisPage/ResumeOptimizationPage/...）
不改后端 Go 源码
不改数据库 SQL
不改 Agent / Prompt / Schema / Mock
不改 API 路径
不改 .env
不引入依赖
不大改路由结构
不删除 PageTransitionOverlay 组件文件
```

---

## 九、总结

```
1. PageTransitionOverlay → 停用（deprecated），从 routes.tsx 移除挂载
2. RouteProgressBar → 普通页面跳转唯一全局反馈
3. PageLoading / Skeleton → 保留为页面数据加载
4. AppBootLoading / AIThinkingLoading → 保留边界不变
5. 页面内容 fade-in → 可选增强，下一阶段优先不做
```
