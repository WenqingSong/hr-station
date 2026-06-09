# K9-Final-Nav-C-Fix：Navbar 恢复腾讯校招式透明质感并修复染色问题

> 日期：2026-06-09 | 性质：前端源码修改

---

## 一、C 阶段问题复盘

| 问题 | 原因 | 修复策略 |
| ---- | ---- | ---- |
| K9-Final-Nav-B 白底 Navbar 视觉保守 | 一刀切纯白底，Hero 区域无融合感 | 恢复 Hero 深色透明模式，但避免白半透 |
| 旧版 bg-white/58 蓝色穿透 | 白色半透覆盖蓝色 Hero → 42% 蓝色透过 | **禁白半透**，Hero 顶部只用深色透明 `bg-slate-950/55` |
| 旧版 isTop <= 80 过早切换 | 80px 时仍在 Hero 视觉区域，突然变白底 | **阈值提高到 280px**，在 Hero 视觉区域内保持深色透明 |
| 旧版 bg-slate-950/18 太浅 | 18% 深色效果不明显 | **55% 深色透明**，文字白色清晰可读 |

---

## 二、修改清单

| 文件 | 修改 |
| ---- | ---- |
| `web/src/app/components/Navbar.tsx` | 恢复双模式视觉 + 修复策略 |

---

## 三、Navbar 视觉状态重构

| 状态 | 背景 | 文字 | 边框 | blur | 结果 |
| ---- | ---- | ---- | ---- | ---- | ---- |
| Hero 顶部（scrollY < 280 && Hero 路由） | `bg-slate-950/55` 深色半透明 | `text-white` / `text-white/85` | `border-white/10` 弱白边框 | `backdrop-blur-md` 轻微模糊 | ✅ |
| 内容区/普通页面 | `bg-white` 纯白 | `text-slate-950` / `text-slate-700` | `border-slate-200/70` | 无 | ✅ |

### 视觉分支

```ts
const navSurface = isHeroTop
  ? "bg-slate-950/55 border-white/10 text-white backdrop-blur-md shadow-none"
  : "bg-white border-slate-200/70 text-slate-950 shadow-sm";
```

### 颜色过渡

- `transition-colors duration-200` — 快速平滑切换，不出现拖尾色带
- `transition-transform duration-300` — 滚动显隐动画（保留）

---

## 四、Hero 页面判断逻辑

| 项 | 修改前 | 修改后 |
| ---- | ---- | ---- |
| Hero 路由检测 | 已删除 | 恢复：`/`, `/jobs`, `/jobs/:id` |
| 透明阈值 | 无 | `HERO_NAV_THRESHOLD = 280`（约 Hero 视觉区域高度） |
| isHeroTop | 无 | `isHeroRoute && scrollY < 280` |
| 初始状态 | 无 | 挂载时立即设置 `window.scrollY < 280` |

### 关键代码

```ts
const isHeroRoute =
  location.pathname === "/" ||
  location.pathname === "/jobs" ||
  /^\/jobs\/[^/]+$/.test(location.pathname) ||
  /^\/job\/[^/]+$/.test(location.pathname);

// 后台持续更新
setIsHeroTop(isHeroRoute && currentScrollY < HERO_NAV_THRESHOLD);
```

---

## 五、禁止回归项确认

| 禁止项 | 结果 |
| ---- | ---- |
| `bg-white/58` | ❌ 不存在（0 命中） |
| `bg-white/60` / `bg-white/70` | ❌ 不存在 |
| `bg-slate-950/18` | ❌ 不存在 |
| `isTop <= 80` | ❌ 不存在 |
| `backdrop-blur-xl` 在 Navbar | ❌ 不存在（仅 `backdrop-blur-md`） |

---

## 六、构建 / 检查结果

| 验收项 | 结果 | 归因 |
| ---- | ---- | ---- |
| npm build | ✅ 通过 | 1773 modules, 1.96s |
| go build | ✅ 通过 | — |
| go vet | ✅ 通过 | — |
| go test | 6 FAIL | pre-existing (K9-2) |
| check-encoding | FAIL | pre-existing (init.sql) |
| ESLint / tsc | ✅ 0 错误 | — |

---

## 七、未修改范围确认

- ✅ 未修改数据库结构
- ✅ 未修改 `server/sql/init.sql`
- ✅ 未修改真实简历数据
- ✅ 未执行 `make reset-up`
- ✅ 未执行任何 Git 命令
- ✅ 未新增依赖
- ✅ 未修改后端 API
- ✅ 未修改业务逻辑
- ✅ 未修改路由参数
