# K9-Final-Nav-D-B：腾讯校招式轻薄透明 Navbar 重构

> 日期：2026-06-09 | 性质：前端源码修改（3 个 class 参数）

---

## 一、修改清单

| 文件 | 修改 |
| ---- | ---- |
| `web/src/app/components/Navbar.tsx` L108 | 3 个 class 参数调整 |

### 未改文件

- JobHeroBanner.tsx ✅
- JobDetailBanner.tsx ✅
- routes.tsx ✅
- 所有普通页面 ✅

---

## 二、三个视觉参数调整

| 参数 | 修改前 | 修改后 | 理由 |
| ---- | ---- | ---- | ---- |
| Hero 背景透明度 | `bg-slate-950/55`（55%） | `bg-slate-950/28`（28%） | /55 形成厚重深蓝实色横条；/28 是轻薄暗化 sweet spot |
| blur 强度 | `backdrop-blur-md` | `backdrop-blur-[2px]` | md 抹掉 Hero 亮部让暗色填充；2px 仅轻微柔化 |
| 边框强度 | `border-white/10`（10% 白） | `border-white/5`（5% 白） | Hero 顶部不需要可见分割线 |

### 最终 L108

```tsx
? "bg-slate-950/28 border-white/5 text-white backdrop-blur-[2px] shadow-none"
```

---

## 三、未扩大范围确认

| 项 | 结果 |
| ---- | ---- |
| HERO_NAV_THRESHOLD | ✅ 不变（280） |
| isHeroRoute / isHeroTop | ✅ 不变 |
| Solid 态 bg-white | ✅ 不变 |
| 普通页面 Navbar | ✅ 不变 |
| 滚动显隐逻辑 | ✅ 不变 |
| Hero Banner | ✅ 未改 |
| 状态机 | ✅ 仍为 hero/solid 两态 |

---

## 四、源码检查

| 检查项 | 结果 | 归因 |
| ---- | ---- | ---- |
| `bg-slate-950/55` | ❌ 0 命中 | ✅ |
| `backdrop-blur-md` | ❌ 0 命中 | ✅ |
| `border-white/10` | ❌ 0 命中 | ✅ |
| `bg-white/58` | ❌ 0 命中 | ✅ |
| `bg-slate-950/18` | ❌ 0 命中 | ✅ |
| `bg-slate-950/28` | ✅ 存在（L108） | ✅ |
| `backdrop-blur-[2px]` | ✅ 存在（L108） | ✅ |
| `border-white/5` | ✅ 存在（L108） | ✅ |
| `bg-white` solid | ✅ 保留 | ✅ |
| `HERO_NAV_THRESHOLD = 280` | ✅ 保留 | ✅ |

---

## 五、构建 / 检查结果

| 验收项 | 结果 | 归因 |
| ---- | ---- | ---- |
| npm build | ✅ 通过 | 1773 modules, 2.03s |
| go build | ✅ 通过 | — |
| go vet | ✅ 通过 | — |
| go test | 6 FAIL | pre-existing (K9-2) |
| check-encoding | FAIL | pre-existing (init.sql) |

---

## 六、未修改范围确认

- ✅ 未修改数据库结构
- ✅ 未修改 `server/sql/init.sql`
- ✅ 未修改真实简历数据
- ✅ 未执行 `make reset-up`
- ✅ 未执行任何 Git 命令
- ✅ 未新增依赖
- ✅ 未修改后端 API
- ✅ 未修改业务逻辑
- ✅ 未修改路由参数
- ✅ 未修改 Hero Banner
- ✅ 未修改 Navbar 状态机
- ✅ 仅修改 Navbar.tsx L108 的 3 个 class 参数
