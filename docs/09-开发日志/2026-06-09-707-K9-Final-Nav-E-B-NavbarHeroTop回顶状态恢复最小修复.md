# K9-Final-Nav-E-B：Navbar HeroTop 回顶状态恢复最小修复

> 日期：2026-06-09 | 性质：前端源码修改（handleScroll 结构重排）

---

## 一、问题根因修复

| 问题 | 修改前 | 修改后 |
| ---- | ---- | ---- |
| `setIsHeroTop` 被 delta 守卫阻断 | L76 在 L59-63 delta<30 early return 之后 | **L49** 在 delta 守卫之前（最先执行） |
| 回顶时 isHeroTop 不更新 | delta<30 → return → 跳过 setIsHeroTop | 每次 rAF 回调优先评估 isHeroTop |
| `currentScrollY<=20` 被 delta 阻断 | L71 在 delta 守卫之后 | **L64** 独立提前执行，强制显示 |

### 修改前后对比

```
修改前:                          修改后:
rAF callback {                  rAF callback {
  read scrollY                    read scrollY
  check editor route              
  calc delta                      
  if |delta|<30 → RETURN  ❌     |setIsHeroTop ← NOW AT TOP
  setHidden logic                 check editor route
  setIsHeroTop ← TOO LATE        |if scrollY≤20 → show + return
  update ref                      if |delta|<30 → return
}                                 setHidden logic
                                  update ref
                                }
```

---

## 二、修改清单

| 文件 | 修改 |
| ---- | ---- |
| `Navbar.tsx` L41-84 | handleScroll 内重排：`setIsHeroTop` 提到最前面 + `currentScrollY<=20` 提权 |

---

## 三、未修改视觉参数确认

| 参数 | 当前值 | 是否保持 |
| ---- | ---- | ---- |
| Hero bg | `bg-slate-950/28` | ✅ |
| Hero blur | `backdrop-blur-[2px]` | ✅ |
| Hero border | `border-white/5` | ✅ |
| Solid bg | `bg-white` | ✅ |
| 阈值 | `HERO_NAV_THRESHOLD = 280` | ✅ |

---

## 四、源码检查

| 检查项 | 结果 | 归因 |
| ---- | ---- | ---- |
| 禁止参数全部 0 | ✅ | — |
| `setIsHeroTop` 在 delta 之前 | ✅ L49（最先执行） | ✅ |
| `scrollY<=20` 强制显示 | ✅ L64-65 | ✅ |

---

## 五、构建 / 检查结果

| 验收项 | 结果 | 归因 |
| ---- | ---- | ---- |
| npm build | ✅ 通过 | 1.92s |
| go build | ✅ 通过 | — |
| go vet | ✅ 通过 | — |
| ESLint | ✅ 0 错误 | — |

---

## 六、未修改范围确认

- ✅ 未修改数据库/后端/Hero/视觉参数/P2
- ✅ 未执行 Git / reset-up
- ✅ 未新增依赖
