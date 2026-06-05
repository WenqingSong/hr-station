# K8-5-37-D：页面跳转 Loading Overlay 视觉与动效专业精修

> 日期：2026-06-05 | 性质：前端视觉精修 | 不改后端/API/SQL/依赖

---

## 一、本阶段目标

基于 K8-5-37-C 用户浏览器真实验收反馈，对 `PageTransitionOverlay` 进行专业化视觉与动效精修。

**不是新增功能，而是打磨已有组件**。

---

## 二、用户截图问题

```text
1. 遮罩过厚，整页变成灰蓝色雾面，视觉廉价。
2. Navbar 也被压暗，像页面卡死。
3. 中央 Logo + Spinner 太突兀，像默认后台 loading。
4. 文案"正在加载…"太弱，层级不专业。
5. 背景 blur 太重，页面内容被糊成一片。
6. 动效节奏生硬，不像专业 SaaS 产品的页面切换。
```

---

## 三、当前 Overlay 为什么不专业

| 问题 | 根因 |
|------|------|
| 灰蓝雾面 | `bg-[#F5F7FB]/[0.92]` 不透明度太高，色调偏灰蓝 |
| blur 太重 | `backdrop-blur-[2px]` 让底层内容完全模糊 |
| Logo 太大 | 28px 品牌标识在 loading 场景下过度强调品牌 |
| Spinner 太大 | `Loader2 w-6 h-6` 图标型 spinner 太显眼 |
| 文案太弱 | "正在加载…" + `text-gray-400`，语义不准确 + 层级太弱 |
| 无 fade | 直接 opacity jump，没有过渡 |
| 停留时间过长 | SHOW_DELAY=120 + MIN_VISIBLE=360，总共 480ms 感觉拖沓 |

---

## 四、修改文件

| 文件 | 操作 | 说明 |
|------|:---:|------|
| `web/src/app/components/common/PageTransitionOverlay.tsx` | 重写 | 遮罩 / 胶囊 / spinner / 文案 / 动效全部精修 |
| `web/src/app/components/common/RouteProgressBar.tsx` | 微调 | 透明度 1→0.8，辅助补充不抢注意力 |

---

## 五、覆盖范围调整

覆盖范围未变——继续只覆盖 `main` 内容区（`div.flex-1.relative` 容器内），不覆盖 Navbar。

z-index 从 `z-40` 降至 `z-30`，与 Navbar(`z-50`) 之间留出更大安全距离。

---

## 六、视觉参数调整

| 参数 | 修改前 | 修改后 |
|------|--------|--------|
| 遮罩背景 | `bg-[#F5F7FB]/[0.92]` | `bg-slate-50/70` |
| blur | `backdrop-blur-[2px]` | `backdrop-blur-[1px]` |
| z-index | `z-40` | `z-30` |
| 中央元素 | 大 Logo + 大 Spinner + 灰色文案 | 小型白色胶囊 `rounded-full` |
| spinner | `Loader2 w-6 h-6` (10px 图标) | `h-4 w-4 border-2 rounded-full` (细线 spinner) |
| 文案 | "正在加载…" `text-gray-400` | "正在切换页面" `text-slate-700 font-medium` |
| 胶囊 | 无 | `bg-white/90 px-5 py-3 gap-3` |
| ring | 无 | `ring-1 ring-slate-200/80` |
| 阴影 | 无 | `shadow-[0_8px_30px_rgba(15,23,42,0.08)]` |
| 禁用交互 | 无 | `select-none pointer-events-none`（遮罩）；胶囊 `pointer-events-auto` |

---

## 七、动效参数调整

| 参数 | 修改前 | 修改后 |
|------|--------|--------|
| SHOW_DELAY | 120ms | **90ms** |
| MIN_VISIBLE | 360ms | **260ms** |
| fade 过渡 | 无（直接 opacity jump） | `transition-opacity duration-[180ms] ease-out` |
| 胶囊入场 | 无 | `translateY(4px) scale(0.98)` → `translateY(0) scale(1)` |
| RouteProgressBar 透明度 | 1.0 | **0.8** |

总时长对比：
- 修改前：120 + 360 = 480ms
- 修改后：90 + 260 + 180(fade) ≈ 530ms（但更丝滑的 fade 让体验更好）

---

## 八、浏览器验证结果

| 检查项 | 结果 |
|--------|:---:|
| 遮罩不再是灰蓝大雾 | 待用户确认 |
| Navbar 保持清晰不被压暗 | 待用户确认 |
| 中央为小型白色胶囊 + 细线 spinner | 待用户确认 |
| 文案变为"正在切换页面" | 待用户确认 |
| 动画更短更丝滑（90/260/180ms） | 待用户确认 |
| overlay 不会长期停留 | 待用户确认 |
| 不再旧页面/新页面 skeleton 混杂 | 待用户确认 |

**浏览器手工验证待用户确认；代码路径、构建和 API 已通过。**

---

## 九、build / test / API 结果

| 验收项 | 结果 |
|--------|:---:|
| `npm run build` | ✅ 通过（1.98s，1771 modules） |
| `go test ./...` | ✅ 全部通过 |
| `go build ./...` | ✅ 通过 |
| `go vet ./...` | ✅ 通过 |
| `curl /api/jobs?page=1&page_size=5` | ✅ code=0 |
| `curl /api/jobs/10001` | ✅ code=0 |
| `curl /api/jobs/facets` | ✅ code=0 |
| `curl /api/resumes` | ✅ 401（正常） |

---

## 十、未新增依赖确认

```text
未新增第三方 loading 依赖（nprogress、framer-motion、react-spinners、lottie 等）。
```

---

## 十一、未修改范围确认

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

## 十二、风险与边界

```text
本阶段只做 PageTransitionOverlay 视觉和动效精修。
如果后续仍觉得遮罩感强，可以继续降低背景透明度（bg-slate-50/50）或完全移除 blur。
细线 spinner 依赖 border 动画，确认在所有浏览器中都能正常渲染。
```

---

## 十三、Git 约束确认

```text
本次未执行任何 Git 命令，也未提交代码。
```
