# K9-Final-Nav-F-A：Navbar 上滑显示下滑隐藏交互只读诊断

> 日期：2026-06-09 | 性质：只读诊断，未修改源码

---

## 一、Navbar 当前状态机定位

| 逻辑 | 位置 | 当前实现 |
| ---- | ---- | ---- |
| isHeroRoute | L25-29 | `/`, `/jobs`, `/jobs/:id`, `/job/:id` ✅ |
| isHeroTop | L49 | `scrollY < 280` ✅（最先执行） |
| hidden state | L16 | `useState(false)` |
| lastScrollY ref | L20 | `useRef(0)` |
| rAF throttle | L21,44 | `ticking` flag — 一次只允许一个 pending 回调 |
| 强制显示 | L64-68 | `scrollY <= 20` → `setHidden(false)` + return |
| delta 过滤 | L71-76 | `|delta| < 30` → return（**跳过显隐**） |
| 向下隐藏 | L78-79 | `delta > 0 && scrollY > 80` → `setHidden(true)` |
| 向上显示 | L80-82 | `delta < 0` → `setHidden(false)` |
| 视觉 variant | L113-115 | `isHeroTop ? hero : solid` ✅ |
| transform | L142 | `hidden ? "-translate-y-[110%]" : "translate-y-0"` ✅ |

---

## 二、显隐状态机诊断

| 检查项 | 当前实现 | 是否符合目标 |
| ---- | ---- | ---- |
| 是否区分滚动方向 | ✅ L78-82 delta > 0 向下 / < 0 向上 | ✅ 方向区分正确 |
| 向下→隐藏 | ✅ `delta > 0 && scrollY > 80` → `setHidden(true)` | ⚠️ 需 scrollY > 80 才触发 |
| 向上→显示 | ✅ `delta < 0` → `setHidden(false)` | ⚠️ |
| 顶部强制显示 | ✅ `scrollY <= 20` → `setHidden(false)` | ✅ |
| **delta 阈值 30px** | 🔴 L72 `|delta| < 30` 跳过所有显隐逻辑 | **🔴 主因** |
| rAF ticking 可能丢失事件 | ⚠️ L21,44 rAF 排队时新 scroll 事件被跳过 | ⚠️ |
| lastScrollY 更新一致性 | ⚠️ L66/L73/L84 多处更新但逻辑路径不同 | ⚠️ |

### 根因分析

```
慢速滚动场景（最常见的问题）：

1. 用户缓慢向下滚动，每帧 scrollY 变化 5-15px
2. 每次 scroll 事件中 |delta| < 30 → 进入 early return（L72）
3. setHidden 永远不会被调用
4. Navbar 在慢速滚动时从不隐藏 — 不符合目标

目标：用户只要持续向下滚动足够距离，Navbar 就应该隐藏。
当前：只有"快速"滚动（delta >= 30px 的帧）才会触发隐藏。
```

```
rAF 延迟导致 delta 失真：

1. ticking = true ← scroll 事件 A 进入 rAF 队列
2. 用户继续滚动 40px，事件 B/C 被跳过（ticking 仍为 true）
3. rAF 回调执行：currentScrollY = 100, lastScrollY = 30
4. delta = 70 → 通过检查 → 正确隐藏 ✅
5. 但中间 40px 的滚动期间用户看到 Navbar 还在...

问题：大 delta 时逻辑正确，但响应不够及时。
小 delta 时逻辑完全被跳过。
```

---

## 三、视觉状态机诊断

| 检查项 | 当前实现 | 是否风险 |
| ---- | ---- | ---- |
| hero/solid 只由 isHeroTop 决定 | ✅ L113-115 | ✅ 无风险 |
| visible 不参与背景 | ✅ L142 只控 transform | ✅ 无风险 |
| 隐藏后仍更新 isHeroTop | ✅ L49 每次 rAF 先执行 | ✅ 无风险 |

**结论：视觉与显隐已完全分离，架构正确。问题只在显隐状态机。**

---

## 四、参数调优分析

| 参数 | 当前值 | 问题 | 建议值 |
| ---- | ---- | ---- | ---- |
| Delta 阈值 | **30** | 太大，慢速滚动从不触发显隐 | **8~10** |
| 隐藏起始 Y | **80** | 太小，顶部轻微滚动就隐藏 | **96~100** |
| 强制显示 Y | **20** | 合理 ✅ | 保持 20 |
| rAF 节流 | ticking flag | 没问题，但可以简化 | 保留或使用 passive |

---

## 五、修复方案评估

| 方案 | 是否推荐 | 理由 | 风险 |
| ---- | ---- | ---- | ---- |
| A：仅调整参数 | ⚠️ 次推荐 | delta 30→10、HIDE_START 80→96，改 2 个数 | 仍需防范边界 |
| B：重写 handleScroll | ✅ **推荐** | 逻辑更清晰：isHeroTop 先 → force-show → delta 过滤显隐；用常量控制所有阈值 | 比 A 多改 5 行，但可维护性更好 |

### 推荐方案 B

**核心改动**：
```
SHOW_TOP_Y   = 20   （顶部强制显示）
HIDE_START_Y = 96   （超过此位置才考虑隐藏）
SCROLL_DELTA = 8    （显隐过滤阈值，原 30 太大）

1. setIsHeroTop 先执行 ✅（已正确）
2. scrollY <= 20 → 强制显示 + return
3. |delta| < 8 → 仅更新 lastScrollY，跳过显隐
4. delta > 0 && scrollY > 96 → hidden
5. delta < 0 → visible
```

---

## 六、K9-Final-Nav-F-B 执行方案

**修改**：仅 `Navbar.tsx` handleScroll 内部（~15 行重写）

| 改动 | 说明 |
| ---- | ---- |
| `30` → `8` | delta 阈值减小 |
| `80` → `96` | 隐藏起始 Y 提高 |
| 提取为常量 | `SHOW_TOP_Y`, `HIDE_START_Y`, `SCROLL_DELTA` |
| L64-68 独立提前 | force-show 保持优先 |

**不改**：视觉参数、Hero、isHeroTop 逻辑、ticking/rAF 架构。

---

## 七、未修改范围确认

- ✅ 未修改源码/数据库/Git/Hero/视觉参数
