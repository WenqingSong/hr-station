# K9-Final-Nav-G-B：Navbar Surface 渐变过渡重构

> 日期：2026-06-09 | 性质：前端源码修改（双 Surface opacity 交叉淡入淡出）

---

## 一、修改清单

| 文件 | 修改 |
| ---- | ---- |
| `Navbar.tsx` L10-11,16-19 | 新增 blend 常量 + clamp/getHeroProgress |
| `Navbar.tsx` L29 | 新增 `navHeroProgress` state |
| `Navbar.tsx` L62,112 | scroll handler + init 中更新 progress |
| `Navbar.tsx` L143-197 | nav 改为 4 层背景层 + opacity 控制 |
| `Navbar.tsx` L133-148 | 文字 `isHeroTop` → `navHeroProgress >= 0.55` |

---

## 二、新增过渡参数

| 参数 | 值 | 说明 |
| ---- | ---- | ---- |
| HERO_BLEND_START_Y | 220 | 超过此值完全白底 |
| HERO_BLEND_END_Y | 80 | 低于此值完全透明 |
| 过渡区间 | 80~220 (140px) | 线性插值 |

### navHeroProgress 计算

```
scrollY ≤ 80  → progress = 1.0 → 全透明
scrollY = 150 → progress ≈ 0.5 → 50/50
scrollY ≥ 220 → progress = 0.0 → 全白底
普通页面     → progress = 0.0
```

---

## 三、Surface 结构重构

| Surface | 实现 | opacity |
| ---- | ---- | ---- |
| solidSurface（白底） | `absolute inset-0 bg-white shadow-sm` | `1 - navHeroProgress` |
| heroSurface（透明） | `absolute inset-0 bg-slate-950/28 backdrop-blur-[2px]` | `navHeroProgress` |
| solid border | `absolute h-px bg-slate-200/70` | `1 - navHeroProgress` |
| hero border | `absolute h-px bg-white/5` | `navHeroProgress` |

所有背景层均有 `aria-hidden="true"` + `pointer-events-none`，内容层 `relative z-10`。

---

## 四、文字过渡

| 项 | 处理 |
| ---- | ---- |
| 阈值 | `navHeroProgress >= 0.55` |
| transition | `transition-colors duration-200` |

---

## 五、显隐状态机未改

| 项 | 结果 |
| ---- | ---- |
| SHOW_TOP_Y=20 | ✅ |
| HIDE_START_Y=96 | ✅ |
| SCROLL_DELTA=8 | ✅ |
| 下滑隐藏/上滑显示 | ✅ |

---

## 六、构建

| 验收项 | 结果 |
| ---- | ---- |
| npm build | ✅ 1.92s |
| go build | ✅ |
| go vet | ✅ |
| 禁止参数 | 0 ✅ |

---

## 七、未修改范围确认

- ✅ 未修改数据库/后端/Hero/显隐状态机/Git/P2
