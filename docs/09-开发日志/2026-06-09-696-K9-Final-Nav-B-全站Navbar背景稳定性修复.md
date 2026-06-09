# K9-Final-Nav-B：全站 Navbar 背景稳定性修复

> 日期：2026-06-09 | 性质：前端源码修改

---

## 一、修改清单

| 文件 | 修改 |
| ---- | ---- |
| `web/src/app/components/Navbar.tsx` | 背景统一为纯白 + 删除双模式背景逻辑 + 简化代码 |

### 不改文件

- `JobHeroBanner.tsx` — 内容区已有 `pt-24`，Navbar 纯白后蓝色不穿透
- `JobDetailBanner.tsx` — 内容区已有 `pt-20`，同上
- `routes.tsx` — RootLayout 不变
- `ContentPageShell.tsx` — `pt-24` + `bg-[#F5F7FB]` 不变

---

## 二、Navbar 背景策略修复

| 项 | 修改前 | 修改后 | 结果 |
| ---- | ---- | ---- | ---- |
| 背景 | `bg-white/58`（58% 半透）或 `bg-slate-950/18`（18% 深色半透） | `bg-white`（固体白色） | ✅ |
| backdrop-blur | `backdrop-blur-xl` | **移除** | ✅ |
| 双模式分支 | `isHeroNavbarRoute` + `useHeroNavbar` + `isTop` 控制 6 组颜色 | **全部删除** | ✅ |
| 背景过渡动画 | `transition-[transform,background-color,border-color,box-shadow,color]` | `transition-transform`（仅 translate-y） | ✅ |
| Logo 颜色 | 条件：`text-white` 或 `text-slate-950` | 固定：`text-slate-950` | ✅ |
| 链接颜色 | 条件：`text-white/88` 或 `text-slate-700` | 固定：`text-slate-700` | ✅ |
| 退出颜色 | 条件：`text-white/72` 或 `text-slate-500` | 固定：`text-slate-500` | ✅ |
| 登录按钮 | 条件：白字半透玻璃 或 深色字半透 | 固定：深色字半透（保留 `backdrop-blur`） | ✅ |
| border | 条件：`border-white/10` 或 `border-slate-200/40` | 固定：`border-slate-200/70` | ✅ |
| shadow | 条件：`shadow-none` 或 大阴影 | 固定：`shadow-sm` | ✅ |
| 滚动显隐动画 | ✅ 保留 | ✅ 保留（delta 30px + 80px 阈值） | ✅ |
| z-index | `z-50` | `z-50`（不变） | ✅ |
| 高度 | `h-16 md:h-[68px]` | 不变 | ✅ |

### 代码变化统计

- 删除：`isTop` state、`isHeroNavbarRoute`、`useHeroNavbar`、`brandTextShadow`、6 个条件 class 变量
- 简化的 nav className：从 6 个拼接项 → 7 个固定项
- 净减 ~50 行

---

## 三、Hero 叠层修复

| 页面 / 组件 | 修改前 | 修改后 | 结果 |
| ---- | ---- | ---- | ---- |
| JobHeroBanner | `bg-[#001a4d]` 从 y=0 开始 + `pt-24` 内容偏移 | **无需修改** — Navbar 纯白后蓝色从后方不透出 | ✅ |
| JobDetailBanner | 蓝色渐变从 y=0 开始 + `pt-20` 内容偏移 | **无需修改** — 同上 | ✅ |

### 原理

Navbar 是 `fixed z-50` 覆盖在页面最上层。修改前 Navbar 使用 `bg-white/58`（42% 透明），Hero Banner 的蓝色背景从透明部分穿透，形成异常浅蓝色带。修改后 Navbar 使用 `bg-white`（完全不透明），Hero 背景无法穿透 — 无论 Hero Banner 从什么位置开始都不会再有问题。

---

## 四、普通页面回归

| 页面 | 结果 |
| ---- | ---- |
| `/recommend` | ✅ ContentPageShell `bg-[#F5F7FB]` + `pt-24`，Navbar 纯白 |
| `/match` | ✅ 同上 |
| `/resume-optimization` | ✅ 同上 |
| `/resumes` | ✅ 同上 |
| `/resumes/compare` | ✅ 同上 |
| `/resumes/edit` | ✅ ResumeEditorShell，编辑页 Navbar 始终显示 |

---

## 五、源码检查

| 检查项 | 结果 | 归因 |
| ---- | ---- | ---- |
| Navbar 中 `bg-white/58` | ❌ 已删除（0 命中） | ✅ |
| Navbar 中 `bg-slate-950/18` | ❌ 已删除（0 命中） | ✅ |
| Navbar 中 `backdrop-blur-xl` | ❌ 已删除（0 命中） | ✅ |
| Navbar 中 `isHeroNavbarRoute` | ❌ 已删除（0 命中） | ✅ |
| Navbar 中 `brandTextShadow` | ❌ 已删除（0 命中） | ✅ |
| Navbar 中 `isTop` | ❌ 已删除（0 命中） | ✅ |
| Navbar `bg-white` solid | ✅ 存在 | ✅ |

---

## 六、构建 / 检查结果

| 验收项 | 结果 | 归因 |
| ---- | ---- | ---- |
| npm build | ✅ 通过 | 1773 modules, 2.06s |
| go build | ✅ 通过 | — |
| go vet | ✅ 通过 | — |
| go test | 6 FAIL | pre-existing（K9-2 match agent/service） |
| check-encoding | FAIL | pre-existing（init.sql dump 编码） |
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
- ✅ 未删除有效内容（仅删除双模式背景逻辑）
