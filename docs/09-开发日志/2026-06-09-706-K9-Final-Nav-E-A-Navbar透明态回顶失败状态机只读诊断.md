# K9-Final-Nav-E-A：Navbar 透明态回顶失败状态机只读诊断

> 日期：2026-06-09 | 性质：只读诊断，未修改源码

---

## 一、Navbar 状态机定位

| 逻辑 | 当前位置 | 当前实现 |
| ---- | ---- | ---- |
| Hero route 判断 | L25-29 `isHeroRoute` | `/`, `/jobs`, `/jobs/:id`, `/job/:id` |
| scrollY 读取 | L46 | `window.scrollY` |
| HeroTop 状态设置 | L76 | `setIsHeroTop(isHeroRoute && scrollY < 280)` |
| 显隐状态（hidden） | L66-72 | delta > 0 && scrollY>80 → hide; delta < 0 → show; scrollY≤20 → show |
| 背景 variant | L107-109 | `isHeroTop ? hero(透明) : solid(白底)` |
| scroll listener | L88 | `window.addEventListener("scroll", ...)` |
| cleanup | L89 | `removeEventListener` |
| pathname 变化处理 | L90 | `useEffect` deps: `[location.pathname, isHeroRoute]` |
| delta 守卫 | **L59-63** | **`|delta| < 30` → early return（不更新 isHeroTop）** |

---

## 二、显隐状态与视觉状态耦合诊断

| 检查项 | 当前实现 | 是否风险 |
| ---- | ---- | ---- |
| 显隐是否独立于视觉 | ✅ `hidden` 只控制 `-translate-y-[110%]`，不参与 `navSurface` | ✅ |
| 视觉是否独立于显隐 | ✅ `navSurface` 只由 `isHeroTop` 决定 | ✅ |
| 但两者共享一个 handler | ✅ 同一个 `handleScroll` 统一处理 | 正常 |
| **delta<30 时会同时跳过两者** | 🔴 L59-63 early return **阻止所有状态更新** | **🔴 根因** |

### 关键代码

```js
// L59-63: THE ROOT CAUSE
const delta = currentScrollY - lastScrollY.current;
if (Math.abs(delta) < 30) {
  lastScrollY.current = currentScrollY;  // ref gets updated
  ticking.current = false;
  return;  // ← *** BOTH setHidden AND setIsHeroTop are SKIPPED ***
}

// L66-72: hidden state update (UNREACHABLE when |delta| < 30)
if (delta > 0 && currentScrollY > 80) setHidden(true);
else if (delta < 0) setHidden(false);
if (currentScrollY <= 20) setHidden(false);

// L76: isHeroTop update (UNREACHABLE when |delta| < 30)
setIsHeroTop(isHeroRoute && currentScrollY < HERO_NAV_THRESHOLD);
```

---

## 三、回顶失败原因排查

| 可能原因 | 是否存在 | 证据 |
| ---- | ---- | ---- |
| **delta<30 early return 跳过 isHeroTop 更新** | **🔴 是** | L59-63 在 `setIsHeroTop`（L76）之前 return |
| lastScrollY 只在 early return 内更新 | ⚠️ 是 | L61 更新 ref 但不评估 isHeroTop |
| scroll 事件绑定 window，滚动容器正确 | ✅ 正常 | RootLayout 无内部 `overflow` 容器 |
| pathname 识别正确 | ✅ 正常 | L25-29 覆盖 `/`, `/jobs`, `/jobs/:id` |
| 视觉 class 正确 | ✅ 正常 | L107-109 `isHeroTop` 驱动 |

### 根因确认

**`delta < 30` 的 early return 是唯一根因。**

回顶失败场景还原：
1. 用户在页面底部（scrollY > 280），Navbar 白底
2. 用户开始向上滚回顶部
3. 滚动事件产生一系列 scroll 事件，某些事件 delta 较小（< 30px）
4. 小 delta 事件进入 early return → **不执行 `setIsHeroTop`**
5. 即使最终 scrollY ≈ 0，如果最后一次有效的 scroll 事件 delta < 30 → `isHeroTop` 保持 false
6. Navbar 维持白底，不复原透明态

---

## 四、滚动容器诊断

| 容器 | 是否内部滚动 | 证据 |
| ---- | ---- | ---- |
| RootLayout | ❌ 否 | `flex min-h-screen flex-col` — 无 overflow，自然文档流 |
| `<main>` | ❌ 否 | `flex-1` 无 overflow |
| JobHeroBanner | ❌ 否 | `overflow-hidden` 仅裁剪溢出，不影响滚动 |
| 结论 | ✅ window 滚动 | `window.scrollY` 正确 |

---

## 五、正确状态机设计

### 核心规则

```ts
// 规则 1：isHeroTop 只由两个因素决定
isHeroTop = isHeroRoute && scrollY < HERO_NAV_THRESHOLD;

// 规则 2：只要 scrollY <= 8，强制 visible
if (scrollY <= 8) setHidden(false);

// 规则 3：isHeroTop 的更新不能受 delta/方向守卫阻断
// 每次 scroll 事件（或至少每次 rAF 回调）都应评估 isHeroTop
```

### 建议 handleScroll 结构

```ts
const handleScroll = () => {
  if (!ticking.current) {
    window.requestAnimationFrame(() => {
      const currentScrollY = window.scrollY;

      // 1. ALWAYS update HeroTop state（不受 delta 守卫）
      setIsHeroTop(isHeroRoute && currentScrollY < HERO_NAV_THRESHOLD);

      // 2. Hidden state: suppress edits for resume editor
      if (isResumeEditorRoute) { setHidden(false); ... return; }

      // 3. Delta guard only for HIDDEN logic（not for isHeroTop）
      const delta = currentScrollY - lastScrollY.current;
      if (Math.abs(delta) >= 30) {
        if (delta > 0 && currentScrollY > 80) setHidden(true);
        else if (delta < 0) setHidden(false);
      }

      // 4. Force visible when near top
      if (currentScrollY <= 8) setHidden(false);

      lastScrollY.current = currentScrollY;
      ticking.current = false;
    });
    ticking.current = true;
  }
};
```

**关键改动：`isHeroTop` 更新提到 delta 守卫之前。**

---

## 六、是否需要 IntersectionObserver

| 方案 | 是否推荐 | 理由 | 风险 |
| ---- | ---- | ---- | ---- |
| A：scrollY threshold + 解除 delta 守卫 | ✅ **推荐** | 改动最小（仅把 `setIsHeroTop` 提到 early return 之前）；问题根因明确 | 阈值 280 对详情页 Hero 可能稍粗（但不影响本次修复） |
| B：IntersectionObserver | ❌ 不推荐 | 需要 Hero DOM 标记；改动大；当前问题不是 threshold 精度导致 | 页面 Hero 高度变化时需维护 |

---

## 七、K9-Final-Nav-E-B 执行方案

**修复范围**：仅 `Navbar.tsx` handleScroll 函数内部。

| 修改 | 行号 | 内容 |
| ---- | ---- | ---- |
| 解除 isHeroTop 的 delta 阻断 | L59-76 | 将 `setIsHeroTop(...)` 移到 early return 之前 |
| isResumeEditorRoute 也应更新 isHeroTop | L47-57 | editor 路由也需评估（虽然大概率非 Hero 路由） |

**不改**：
- 视觉参数 ✅
- Hero Banner ✅
- 普通页面 ✅
- 显隐逻辑 ✅
- 阈值 280 ✅

---

## 八、未修改范围确认

- ✅ 未修改任何源码
- ✅ 未修改数据库
- ✅ 未修改 `server/sql/init.sql`
- ✅ 未执行 `make reset-up`
- ✅ 未执行 Git
- ✅ 未修改 Navbar 视觉参数
- ✅ 未修改 Hero
