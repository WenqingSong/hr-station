# K9-Final-Nav-F-B：Navbar 上滑显示下滑隐藏状态机重构

> 日期：2026-06-09 | 性质：前端源码修改（handleScroll 逻辑 + 常量提取）

---

## 一、修改清单

| 文件 | 修改 |
| ---- | ---- |
| `Navbar.tsx` L9-12 | 新增 3 个显隐常量 |
| `Navbar.tsx` L43-95 | handleScroll 逻辑重构 |

---

## 二、状态机常量

| 常量 | 修改前 | 修改后 | 作用 |
| ---- | ---- | ---- | ---- |
| SHOW_TOP_Y | 无（硬编码 20） | **20** | 顶部强制显示阈值 |
| HIDE_START_Y | 无（硬编码 80） | **96** | 向下滚动超过此位置才隐藏 |
| SCROLL_DELTA | 无（硬编码 30） | **8** | 显隐过滤阈值（慢速滚动也响应） |
| HERO_NAV_THRESHOLD | 280 | **280**（不变） | Hero 视觉透明阈值 |

---

## 三、handleScroll 逻辑重构

| 步骤 | 逻辑 | 说明 |
| ---- | ---- | ---- |
| 1 | `setIsHeroTop(...)` | 视觉状态最先执行，不受任何守卫阻挡 |
| 2 | `scrollY <= SHOW_TOP_Y` | 顶部强制显示（20px） |
| 3 | `|delta| < SCROLL_DELTA` | 跳过显隐（8px vs 旧 30px — 更灵敏） |
| 4 | `delta > 0 && scrollY > HIDE_START_Y` | 向下滚动隐藏（96px vs 旧 80px） |
| 5 | `delta < 0` | 向上滚动显示（独立 if，去除了 else-if 耦合） |

### 关键改进

- `else if` → 独立 `if`：向上和向下判断互不排斥
- HIDE_START_Y 80→96：减少顶部轻微滚动误触发隐藏
- SCROLL_DELTA 30→8：慢速滚动也响应显隐

---

## 四、未修改视觉参数确认

| 参数 | 当前值 | 保持 |
| ---- | ---- | ---- |
| Hero bg | `bg-slate-950/28` | ✅ |
| Hero blur | `backdrop-blur-[2px]` | ✅ |
| Hero border | `border-white/5` | ✅ |
| Solid bg | `bg-white` | ✅ |
| 阈值 | 280 | ✅ |

---

## 五、源码检查

| 检查项 | 结果 |
| ---- | ---- |
| SHOW_TOP_Y=20 | ✅ L10 |
| HIDE_START_Y=96 | ✅ L11 |
| SCROLL_DELTA=8 | ✅ L12 |
| 禁止视觉参数 | 0 ✅ |
| isHeroTop 在 delta 前 | ✅ L52 |

---

## 六、构建

| 验收项 | 结果 |
| ---- | ---- |
| npm build | ✅ 2.27s |
| go build | ✅ |
| go vet | ✅ |

---

## 七、未修改范围确认

- ✅ 未修改数据库/后端/Hero/视觉参数/Git/P2
