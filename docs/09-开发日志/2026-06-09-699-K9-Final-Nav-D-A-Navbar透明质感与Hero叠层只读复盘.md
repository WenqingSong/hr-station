# K9-Final-Nav-D-A：Navbar 透明质感与 Hero 叠层只读复盘

> 日期：2026-06-09 | 性质：只读诊断，未修改任何源码

---

## 一、当前文件定位

| 文件 | 作用 | 当前关键实现 |
| ---- | ---- | ---- |
| `Navbar.tsx` | 固定顶部导航 | L107-109：`bg-slate-950/55` + `backdrop-blur-md` Hero 态 / `bg-white` solid 态 |
| `JobHeroBanner.tsx` | 首页岗位列表 Hero | `bg-[#001a4d]` 深蓝底 + `radial-gradient` + 光斑 + `pt-24` 内容偏移 |
| `JobDetailBanner.tsx` | 岗位详情 Hero | `bg-gradient-to-br from-[#0052D9] via-[#0066FF] to-[#0040A8]` + `pt-20` |
| `routes.tsx` | 路由 | RootLayout：`<Navbar />` + `<main><Outlet /></main>`，背景 `bg-[#F5F7FA]` |

---

## 二、三版 Navbar 效果复盘

### 版本 0：原始半透明版（历史）

| 项 | 判断 |
| ---- | ---- |
| 优点 | Hero 顶部视觉融合好，接近腾讯校招风格 |
| 缺点 | `bg-white/58` 白色半透压蓝色 Hero → 42% 蓝色穿透 → 异常浅蓝色带 |
| 具体 bug 根因 | **白色半透明背景覆盖蓝色 Hero** — `bg-white/58` 让蓝色从后方 42% 穿透，形成浅蓝污染 |
| 哪些值得保留 | Hero 顶部的透明融合设计思路（方向对，参数错） |
| 哪些必须禁止 | **白色半透明**（`bg-white/58`、`bg-white/60`、`bg-white/70`）压蓝色 Hero；`isTop <= 80` 过早切换 |

### 版本 1：纯白 Navbar（Nav-B）

| 项 | 判断 |
| ---- | ---- |
| 优点 | 稳定，无穿透，无异常色带 |
| 缺点 | Hero 顶部非常割裂 — Navbar 是白条压在蓝色 Hero 上，像后台管理系统 |
| 为什么用户觉得丑 | 纯白 `bg-white` solid 与蓝色 Hero 形成强烈断裂感，无融合无质感 |
| 哪些可以保留 | Solid 态 `bg-white` 用于内容区和普通页面（正确） |
| 哪些不应继续 | Hero 区域使用纯白 Navbar（完全错误） |

### 版本 2：当前深色透明（Nav-C-Fix）

| 项 | 判断 |
| ---- | ---- |
| 优点 | 方向正确（Hero 深色透明），无白色半透染色，文字白色可读 |
| 缺点 | **视觉太重** — 变成一整条厚重深蓝/深黑横条，不像腾讯校招的轻薄透明 |
| 为什么变成厚重深蓝条 | `bg-slate-950/55`（55% 不透明度太高）+ `backdrop-blur-md`（强模糊让背景失去亮点）+ `border-white/10`（可见分割线）三者叠加 |
| opacity 是否过高 | **是** — 55% 的 `rgb(2,6,23)` 压在任何蓝色 Hero 上都形成几乎实色的深色条 |
| blur 是否过强 | **是** — `backdrop-blur-md` 抹掉了 Hero 的 bright highlight 区域，让暗色均匀填充 |
| 是否像腾讯校招官网 | **不像** — 腾讯校招 Navbar 是 15-25% 的轻暗化，几乎不可见的叠加层 |

---

## 三、当前"厚重深蓝条"根因分析

| 检查项 | 当前实现 | 是否导致厚重 |
| ---- | ---- | ---- |
| 背景透明度 | `bg-slate-950/55`（55% 不透明） | **🔴 主因** — 55% 几乎达到"半实色" |
| blur 强度 | `backdrop-blur-md` | **🔴 辅助** — 抹掉 Hero 亮部，让暗色填充更均匀 → 更"实" |
| border | `border-white/10`（10% 白边框） | ⚠️ 轻微 — 增加视觉分离线 |
| shadow | `none`（Hero 态） | ✅ 无影响 |
| Navbar 高度 | `h-16 / md:h-[68px]`（64px / 68px） | ✅ 合理，非主要因素 |
| transition | `duration-200` | ✅ 合理 |

### 根因定量分析

```
Hero 颜色: bg-[#001a4d] = rgb(0, 26, 77)
叠加图层: bg-slate-950/55 = rgba(2, 6, 23, 0.55)

合成结果 ≈ rgb(1.1, 14.2, 46.3) — 这是一个非常深暗的颜色

再加上 backdrop-blur-md 抹掉 Hero 的亮部光斑:
→ 用户看到的是一条均匀的深暗色横条

腾讯校招参考: ~rgba(0,0,0,0.18~0.25) + 极轻或零 blur
→ 合成结果 ≈ rgb(0, 20, 62) — 仅比 Hero 略暗，像一层"柔光滤镜"
```

### 结论

根因不是单一的 opacity 值，而是 **opacity 过高 × blur 过强 × border 可见** 三者叠加。本质是 Navbar 太重，覆盖了 Hero 视觉层次。

---

## 四、腾讯校招式目标视觉 token

| 项 | 候选值 | 推荐值 | 理由 |
| ---- | ---- | ---- | ---- |
| Hero 背景 | `bg-slate-950/24` / `/28` / `/32` / `/40` | **`bg-slate-950/28`** | 28% 是 sweet spot — 给 Hero 可见的一层微妙暗化，但不形成明显深色横条 |
| blur | `none` / `backdrop-blur-[2px]` / `backdrop-blur-sm` | **`backdrop-blur-[2px]`** 或 **`backdrop-blur-sm`** | 2px 足够柔化背景边缘，但不消除亮部；sm 是可用的备选 |
| border | `none` / `border-white/5` / `border-white/8` | **`none`** 或 **`border-white/5`** | 腾讯校招无可见边框；如果必须保留，用极淡的 5% |
| shadow | `none` | **`none`** | Hero 态不需要阴影 |
| height | `h-16 md:h-[68px]` | **不变** | 高度合理 |
| transition | `150ms` / `200ms` | **`transition-colors duration-200`** | 保持当前不变 |
| white state | `bg-white` / `bg-white/96` | **`bg-white`** | solid 白底不变 |
| threshold | `220` / `280` / `sentinel` | **`280`** | 当前合理，覆盖 Hero 主体 + 部分过渡 |

### 推荐参数组合

```
Hero 态:
  bg-slate-950/28
  backdrop-blur-[2px]  (备选: backdrop-blur-sm)
  border-white/5       (备选: none)
  shadow-none
  text-white

Solid 态:
  bg-white
  border-slate-200/70
  shadow-sm
  text-slate-950
```

---

## 五、新 Navbar 状态机设计

| 状态 | 触发条件 | 背景 | 文字 | 边框 | blur | 说明 |
| ---- | ---- | ---- | ---- | ---- | ---- | ---- |
| HeroTop | `isHeroRoute && scrollY < 280` | `bg-slate-950/28` | `text-white` / `text-white/85` | `border-white/5` | `backdrop-blur-[2px]` | 轻薄暗化，白色文字清晰 |
| Solid | 其他所有情况 | `bg-white` | `text-slate-950` / `text-slate-700` | `border-slate-200/70` | 无 | 纯白底 + 轻阴影 |

### 状态逻辑（保持当前架构）

```ts
const isHeroRoute = ...;  // 不变
const isHeroTop = isHeroRoute && scrollY < 280; // 不变
const navVariant = isHeroTop ? "hero" : "solid";
```

**不新增复杂度**：仍然是简单的 hero/solid 两态，只调整参数。

---

## 六、禁用策略

| 禁止项 | 理由 |
| ---- | ---- |
| `bg-white/58`、`bg-white/60`、`bg-white/70` | 白色半透明覆盖蓝色 Hero → 蓝色穿透污染 |
| `bg-slate-950/50` 及以上重度深色 | 形成厚重深色横条 |
| `backdrop-blur-md` / `backdrop-blur-xl` | 过度模糊消除 Hero 亮部，让暗色条更"实" |
| `isTop <= 80` 过早切换 | Hero 内部就变白底 |
| 纯白 `bg-white` 覆盖 Hero 顶部 | 白条压蓝底，视觉割裂 |
| 6 组以上颜色分支 | 复杂度失控 |

### 允许

| 允许项 | 范围 |
| ---- | ---- |
| Hero 顶部轻薄深色透明 | `bg-slate-950/24` ~ `bg-slate-950/32` |
| 轻 blur | `backdrop-blur-[2px]` 或 `backdrop-blur-sm` |
| Solid 白底 | `bg-white` |
| 简单 hero/solid 两态 | 2 组视觉参数 |
| 150-200ms 过渡 | `transition-colors duration-200` |

---

## 七、是否需要改 Hero

| 项 | 是否需要改 | 理由 |
| ---- | ---- | ---- |
| JobHeroBanner padding | **不需要** | 已有 `pt-24`，Navbar 轻薄后视觉自然 |
| JobDetailBanner padding | **不需要** | 已有 `pt-20`，同上 |
| Hero 背景亮度 | **不需要** | `bg-[#001a4d]` 深蓝适合强对比，Navbar 轻薄后反差刚好 |
| Hero 高度 | **不需要** | `lg:min-h-[380px]` 合理，280px 阈值覆盖 Hero 主体 |
| Hero 顶部遮罩 | **不需要** | Navbar 自身轻薄到 `bg-slate-950/28` 已足够 |

**结论**：不改 Hero。Navbar 参数调整即可达到目标效果。

---

## 八、K9-Final-Nav-D-B 执行方案

| 模块 | 修改方向 | 行数 |
| ---- | ---- | ---- |
| `Navbar.tsx` L108 — Hero 背景 | `bg-slate-950/55 → bg-slate-950/28` | 1 行 |
| `Navbar.tsx` L108 — blur | `backdrop-blur-md → backdrop-blur-[2px]`（或 `sm`） | 1 行 |
| `Navbar.tsx` L108 — border | `border-white/10 → border-white/5`（或删除） | 1 行 |

### 推荐最终 class

```tsx
// Hero 态
"bg-slate-950/28 border-white/5 text-white backdrop-blur-[2px] shadow-none"

// Solid 态（不变）
"bg-white border-slate-200/70 text-slate-950 shadow-sm"
```

### 不改范围

- 不改 `isHeroRoute` / `isHeroTop` / `HERO_NAV_THRESHOLD` 逻辑
- 不改 Solid 态样式
- 不改文字颜色分支
- 不改滚动显隐逻辑
- 不改 Hero Banner
- 不改 DB / API / 路由

---

## 九、浏览器验收标准

| 场景 | 标准 |
| ---- | ---- |
| Hero 顶部 | Navbar 轻薄透明 — 不是厚重深蓝横条，不是纯白大条 |
| Hero 顶部 | 一层微微暗化的薄膜，白色文字清晰可读 |
| Hero 顶部 | 不出现白色半透明染蓝 bug |
| Hero 顶部 | 视觉接近腾讯校招官网的轻薄质感 |
| 滚动切换 | `transition-colors duration-200` 自然 |
| 内容区 | 白底稳定，不异常变色 |
| 普通页面 | 白底稳定（recommend/match/optimization/resumes） |
| 移动端 | 不横向溢出 |
| Console | 无新增错误 |

---

## 十、未修改范围确认

- ✅ 未修改任何源码
- ✅ 未修改数据库
- ✅ 未修改 `server/sql/init.sql`
- ✅ 未执行 `make reset-up`
- ✅ 未执行任何 Git 命令
- ✅ 仅执行只读查看 + docs 更新
