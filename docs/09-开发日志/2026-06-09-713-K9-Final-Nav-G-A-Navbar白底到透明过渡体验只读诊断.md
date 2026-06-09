# K9-Final-Nav-G-A：Navbar 白底到透明过渡体验只读诊断

> 日期：2026-06-09 | 性质：只读诊断，未修改源码

---

## 一、Navbar 当前实现定位

| 逻辑 | 位置 | 当前实现 |
| ---- | ---- | ---- |
| Hero 判断 | L52 | `scrollY < 280` → boolean |
| 背景 class | L122-124 | `isHeroTop ? heroClass : solidClass` **二值** |
| 文字颜色 | L126-140 | 5 组 `isHeroTop ? whiteVariant : darkVariant` |
| border | L122-124 | 同上二值 |
| shadow | L122-124 | 同上二值 |
| blur | L122-124 | 同上二值（2px vs none） |
| transition | L149 | `transition-colors duration-200` |
| progress/opacity | — | **不存在** 🔴 |
| 中间过渡态 | — | **不存在** 🔴 |

---

## 二、二值硬切诊断

| 检查项 | 当前实现 | 是否硬切 |
| ---- | ---- | ---- |
| 背景色 | `isHeroTop ? "bg-slate-950/28" : "bg-white"` | **🔴 是** — 一个阈值点翻转 |
| 文字色 | `isHeroTop ? "text-white" : "text-slate-950"` | **🔴 是** |
| border | `isHeroTop ? "border-white/5" : "border-slate-200/70"` | **🔴 是** |
| shadow | `isHeroTop ? "shadow-none" : "shadow-sm"` | **🔴 是** |
| blur | `isHeroTop ? "backdrop-blur-[2px]" : none` | **🔴 是** |
| transition | `duration-200` | ⚠️ 200ms 太短，且 backdrop-filter 不可过渡 |
| progress | 不存在 | 🔴 |
| scroll 区间渐变 | 不存在 | 🔴 |

### 根因

```
isHeroTop = scrollY < 280  — 一个 boolean
所有视觉属性在 279→281 时一次性翻转
transition-colors duration-200 仅给 200ms 缓动
backdrop-filter 无法过渡，出现/消失是瞬间的
```

**结论：突兀感来自二值 boolean 切换 + 200ms 过渡不够 + backdrop-filter 不可过渡。**

---

## 三、transition 能力诊断

| 项 | 当前实现 | 影响 |
| ---- | ---- | ---- |
| background-color | `transition-colors` ✅ | 可过渡，但 200ms 太短 |
| color | `transition-colors` ✅ | 可过渡 |
| border-color | `transition-colors` ✅ | 可过渡 |
| box-shadow | `transition-colors` ✅ | 可过渡 |
| transform | `transition-transform` ✅ | 显隐用 |
| **backdrop-filter** | **不可过渡** 🔴 | 瞬间出现/消失 |

---

## 四、过渡方案评估

| 方案 | 是否推荐 | 理由 | 风险 |
| ---- | ---- | ---- | ---- |
| A：单层+更长的 transition | ❌ | backdrop-filter 仍不可过渡；仍是二值硬切 | 不治本 |
| B：三态状态机 | ❌ | 状态分支复杂，易与显隐耦合 | 维护成本高 |
| C：双 Surface opacity 交叉淡入淡出 | ✅ **强烈推荐** | 两个绝对定位背景层通过 opacity 交叉淡入；天然支持任何属性组合；backdrop-filter 通过 opacity 平滑出现 | 实现约 40 行 |
| D：CSS variable 动态 rgba | ⚠️ | 平滑但文字/border/shadow 仍需同步 | 不如 C 全面 |

---

## 五、推荐过渡区间设计

| 参数 | 推荐值 | 理由 |
| ---- | ---- | ---- |
| HERO_BLEND_START_Y | **220** | 超过此值为完全白底 |
| HERO_BLEND_END_Y | **80** | 低于此值为完全透明 |
| progress 计算 | `clamp((220-scrollY)/140, 0, 1)` | 线性插值区间 80-220（140px 过渡带） |
| HERO_NAV_THRESHOLD | **保留** 280 用于其他逻辑 | 不影响 blend |

### navHeroProgress 计算

```ts
function getNavHeroProgress(scrollY: number): number {
  if (!isHeroRoute) return 0;
  const raw = (HERO_BLEND_START_Y - scrollY) / (HERO_BLEND_START_Y - HERO_BLEND_END_Y);
  return Math.min(1, Math.max(0, raw));
}
```

```
scrollY  ≤ 80  → progress = 1.0 → 完全透明态
scrollY  = 150 → progress ≈ 0.5 → 50/50 blend
scrollY  ≥ 220 → progress = 0.0 → 完全白底态
```

---

## 六、文字颜色过渡建议

**方案 4（推荐）**：背景优先修，文字用 progress ≥ 0.55 阈值 + `transition-colors duration-200`

```
理由：
- 背景硬切是主要突兀来源（视觉面积大）
- 文字颜色切换用 200ms transition 兜住即可
- 不需要复杂的 color-mix/rgba 插值
```

---

## 七、K9-Final-Nav-G-B 执行方案

**修改**：仅 `Navbar.tsx`（~40 行新增）

| 改动 | 说明 |
| ---- | ---- |
| 新增常量 | `HERO_BLEND_START_Y=220`, `HERO_BLEND_END_Y=80` |
| 新增 state | `navHeroProgress: number` (0~1) |
| scroll 更新 | 每次 rAF 先计算 progress |
| 双背景层 | 两个绝对定位 div：solidSurface + heroSurface |
| opacity | solid: `1-progress`, hero: `progress` |
| 文字 | progress ≥ 0.55 ? white : dark（保持 transition） |
| 内容层 | `relative z-10` 在背景层之上 |

**不改**：显隐逻辑、SHOW_TOP_Y/HIDE_START_Y/SCROLL_DELTA、Hero Banner、视觉颜色值。

---

## 八、未修改范围确认

- ✅ 未修改源码/数据库/Git/Hero/视觉参数/显隐状态机
