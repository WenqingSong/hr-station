# K8-5-37-C：页面跳转整体 Loading Overlay 设计修正

> 日期：2026-06-05 | 性质：前端新增 + 修改 | 不改后端/API/SQL/依赖

---

## 一、本阶段目标

基于 K8-5-37-B 完成后用户真实体验反馈，修正 Loading 设计方向：

从"局部 skeleton / 顶部进度条"为主，调整为"页面切换整体 Loading Overlay"为主。

---

## 二、用户反馈的问题

K8-5-37-B 已完成 RouteProgressBar + AIThinkingLoading + 内联文案替换，但用户真实体验后发现：

```text
当前 loading 更像"模块级加载"：
1. 顶部 Hero 已经渲染出来。
2. 中间内容是 Skeleton。
3. 右侧卡片是局部 loading。
4. 整个页面没有"正在切换页面"的统一感。
5. 页面跳转时会出现新旧页面、局部骨架、右侧卡片混杂的割裂感。
```

用户真正想要的是：

```text
页面跳转时，整个页面短暂进入统一 loading 状态。
不是每个模块各自显示 skeleton。
```

---

## 三、为什么 RouteProgressBar + 局部 Skeleton 不够

| 问题 | 原因 |
|------|------|
| 模块级割裂 | 不同模块各自管理 loading 状态，时序不统一 |
| Hero 先渲染 | 页面级组件在数据到之前就渲染了 Hero 等固定结构 |
| 骨架碎片化 | 每个模块独立显示 skeleton，没有全局统一遮罩 |
| 进度条太弱 | RouteProgressBar 只有 2px，不足以形成"整页切换"的视觉印象 |

需要的不是"每个模块告诉用户它在加载"，而是"告诉用户整个页面正在切换"。

---

## 四、新增 PageTransitionOverlay 设计

### 定位

- **覆盖范围**：main 内容区（`absolute inset-0` 在 `flex-1 relative` 容器内），保留 Navbar 可见
- **不覆盖 Navbar**：Navbar 是稳定导航，不应频繁闪烁
- **不覆盖 AuthDialog**：AuthDialog 使用 `DialogPortal` 渲染到 `document.body`，天然高于页面层

### 视觉

| 属性 | 值 |
|------|-----|
| 背景 | `bg-[#F5F7FB]/[0.92]` + `backdrop-blur-[2px]` |
| 内容 | Offerr鹅 品牌标识 + Loader2 spinner + "正在加载…" |
| z-index | `z-40`（低于 Navbar z-50、RouteProgressBar z-9999、AppBootLoading z-99999） |
| 动画 | 无额外动画，仅 spinner 自旋 |

### 显示策略

```text
1. 路由 location 变化时，不立即显示。
2. 延迟 120ms（SHOW_DELAY）后仍处于切换中，才显示 overlay。
3. 一旦显示，至少保持 360ms（MIN_VISIBLE）。
4. 跳过首次 mount（useRef），不拦截首屏加载。
5. total duration 控制在 480ms - 600ms。
```

### 实现参数

```ts
const SHOW_DELAY = 120;  // 避免极快切换时闪烁
const MIN_VISIBLE = 360;  // 保证可见时长，避免一闪而过
```

---

## 五、接入位置

**文件**：`web/src/app/routes.tsx` — `RootLayout`

**修改前**：
```jsx
<main className="flex-1">
  <Outlet />
</main>
```

**修改后**：
```jsx
<div className="flex-1 relative">
  <PageTransitionOverlay />
  <Outlet />
</div>
```

`PageTransitionOverlay` 作为 `Outlet` 的兄弟节点，通过 `absolute inset-0` 覆盖主要内容区。React Router 的 `useLocation()` hook 在 RootLayout（路由组件）中可用。

---

## 六、Loading 分工最终明确

| 场景 | 组件 | z-index / 说明 |
|------|------|--------|
| 应用启动 / Auth Bootstrap | AppBootLoading | `fixed inset-0 z-[99999]` |
| 路由跳转顶部进度补充 | RouteProgressBar | `fixed top-0 z-[9999]` |
| **页面跳转整体遮罩** | **PageTransitionOverlay** | `absolute z-40`（本阶段新增） |
| Navbar | Navbar | `fixed z-50`（不被覆盖） |
| Auth 弹窗 | AuthDialog (DialogPortal) | 独立 Portal（不被覆盖） |
| 页面数据慢加载 | PageLoading / Skeleton | 正常文档流 |
| AI 生成中 | AIThinkingLoading | 正常文档流 |
| 按钮操作 | 按钮内部 Loader2 | 按钮内部 |

---

## 七、浏览器验证结果

| 检查项 | 结果 |
|--------|:---:|
| 从首页点击 /jobs，出现短暂整体 overlay | 待用户确认 |
| 从 /jobs 点击 /resumes，出现短暂整体 overlay | 待用户确认 |
| 从 /resumes 点击推荐/匹配/优化入口，出现短暂整体 overlay | 待用户确认 |
| overlay 不会长期卡住（max ~480ms） | 待用户确认 |
| overlay 不会遮挡 AuthDialog（DialogPortal） | 待用户确认 |
| overlay 不会导致页面布局跳动 | 待用户确认 |
| 页面切换时不再明显出现旧页面+新页面 skeleton 混杂 | 待用户确认 |

**浏览器手工验证待用户确认；代码路径、构建和 API 已通过。**

---

## 八、build / test / API 结果

| 验收项 | 结果 |
|--------|:---:|
| `npm run build` | ✅ 通过（2.17s，1771 modules） |
| `go test ./...` | ✅ 全部通过 |
| `go build ./...` | ✅ 通过 |
| `go vet ./...` | ✅ 通过 |
| `curl /api/jobs?page=1&page_size=5` | ✅ code=0 |
| `curl /api/jobs/10001` | ✅ code=0 |
| `curl /api/jobs/facets` | ✅ code=0 |
| `curl /api/resumes` | ✅ 401（正常） |

---

## 九、未新增依赖确认

```text
未新增第三方 loading 依赖（nprogress、framer-motion、react-spinners、lottie 等）。
```

---

## 十、未修改范围确认

```text
未修改后端 Go 源码
未修改数据库 SQL
未修改 seed 数据
未修改 Agent / Prompt / Schema / Mock
未修改 API 路径
未修改后端登录鉴权
未修改 jobs 业务接口
未修改简历业务接口
未修改 .env
未安装依赖
未迁移目录
未新增上传能力
未执行 make reset-up
未写入数据库
```

---

## 十一、风险与边界

```text
当前 Overlay 基于 location change 触发，使用固定的 showDelay + minVisible，
属于前端体验层修正，不依赖真实数据加载状态。

如果后续引入 React Router Data Router，可再升级为基于 useNavigation().state === "loading"
的真实 pending 状态，实现更精准的 overlay 控制。
```

---

## 十二、Git 约束确认

```text
本次未执行任何 Git 命令，也未提交代码。
```
