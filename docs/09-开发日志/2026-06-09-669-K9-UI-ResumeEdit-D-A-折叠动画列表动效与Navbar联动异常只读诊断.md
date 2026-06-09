# K9-UI-ResumeEdit-D-A：折叠动画、列表动效与 Navbar 联动异常只读诊断

> 日期：2026-06-09 | 性质：只读诊断，未修改任何源码

---

## 一、当前状态确认

| 项 | 结果 |
| -- | -- |
| Navbar 文件 | `web/src/app/components/Navbar.tsx`（173 行） |
| Navbar 定位 | `fixed inset-x-0 top-0 z-50` |
| Navbar 显隐逻辑 | 基于 scroll direction：向下隐藏 / 向上显示 |
| 编辑页布局 | `ResumeEditorShell` → `ContentPageShell` → body/window 滚动 |
| 父模块动画 | `grid-template-rows: 0fr → 1fr`，240ms |
| AnimatedListItem | 新组件（61 行），enter: RAF+state，exit: max-h transition |
| motion 库 | v12.23.24 已安装，零使用 |

---

## 二、Navbar 显隐逻辑诊断

### 2.1 Navbar scroll listener 逻辑

`Navbar.tsx` L33-56：

```js
const handleScroll = () => {
  const currentScrollY = window.scrollY;
  if (currentScrollY > lastScrollY.current && currentScrollY > 80) {
    setHidden(true);   // 向下滚动且超过 80px → 隐藏
  } else if (currentScrollY < lastScrollY.current) {
    setHidden(false);  // 向上滚动 → 显示
  }
  if (currentScrollY <= 20) {
    setHidden(false);  // 接近顶部 → 强制显示
  }
  lastScrollY.current = currentScrollY;
};
window.addEventListener("scroll", handleScroll, { passive: true });
```

| 项 | 当前实现 | 问题 | 建议 |
| -- | -- | -- | -- |
| Navbar 是否 fixed | ✅ `fixed inset-x-0 top-0 z-50` | — | — |
| 是否根据 scroll direction 显隐 | ✅ 是 | **无法区分真实用户滚动 vs 页面布局变化的 scroll 事件** | 增加最小滚动距离阈值（如 30px）过滤布局 shift |
| 显隐依据 | `scrollY > lastScrollY` 方向判断 | **无 threshold 过滤短距离布局抖动** | `Math.abs(delta) > 30` 才改变显隐状态 |
| threshold | 仅有 `>80` 的单向阈值 | `80` 防止顶部误隐藏，但不足以防御模块折叠导致的 scrollY 变化 | 见上 |
| 模块折叠触发 scroll 事件 | ✅ **必然触发** | 折叠后页面高度减小 → 浏览器自动调整 scrollY → 触发 `handleScroll` | 核心根因 |

### 2.2 完整因果链

```
用户点击折叠模块（如教育经历，约 600px 内容）
  ↓
setOpen(false) → React re-render
  ↓
body 从 grid-rows-[1fr] → grid-rows-[0fr] CSS transition 240ms
  ↓
页面总高度减小 ~500px
  ↓
浏览器在 transition 期间持续触发 scroll 事件（scrollY 随高度变化而变化）
  ↓
Navbar handleScroll 检测到 scrollY 变化 → 误判为"用户向上滚动"
  ↓
Navbar setHidden(false) → translate-y-0（显示）
  ↓
用户看到：折叠模块 → Navbar 突然出现 ❌
```

**展开模块时反向触发**：
```
展开 → 页面高度增加 → scrollY 增大 → 误判为"向下滚动>80" → setHidden(true) → Navbar 消失 ❌
```

### 2.3 根因确认

| 问题 | 证据 | 根因 | 建议 |
| -- | -- | -- | -- |
| 折叠触发 Navbar 显示 | scroll listener 无最小 delta 阈值 | 页面高度变化 ≈ 用户向上滚动 | **增加 30px delta threshold** |
| 展开触发 Navbar 隐藏 | scroll 事件在 transition 期间持续触发 | 240ms 内高度持续变化，多次触发 handler | threshold + **编辑页禁用 auto-hide** |

---

## 三、编辑页滚动容器与布局诊断

| 检查项 | 当前实现 | 问题 | 建议 |
| -- | -- | -- | -- |
| 页面使用 body/window 滚动 | ✅ `ContentPageShell(max-w-[1280px])`，无内部滚动容器 | 模块高度变化直接影响 `window.scrollY` | 无法避免（body scroll 的固有特性） |
| 右侧栏 sticky | ✅ `lg:sticky lg:top-24` | — | — |
| 折叠模块导致页面总高度变化 | ✅ **是** | grid-rows 0fr→1fr 改变 section 高度，页面总高度随之变化 | 这是 grid-rows 动画的副作用 |
| scroll anchoring 行为 | 浏览器默认 `overflow-anchor: auto` | 浏览器可能自动锚定滚动位置，导致在不同高度的 content 间切换时 scrollY 跳变 | 可在 section body 设置 `overflow-anchor: none` |
| scrollIntoView / 自动聚焦 | ❌ 没有 | — | — |

### 结论

**编辑页使用 body scroll + grid-rows 高度动画** 是 Navbar 联动的结构根因。页面高度的任何变化都会触发 scroll 事件并被 Navbar 误判。

---

## 四、父模块折叠动画"不够丝滑"诊断

| 问题 | 证据 | 根因 | 建议 |
| -- | -- | -- | -- |
| 页面总高度剧烈变化 | `grid-rows-[1fr]` → `grid-rows-[0fr]`：教育模块折叠时页面高度瞬间减小 ~600px | grid-rows 1fr 对应内容实际高度，折叠后变为 0 | **在 section 上设置 `overflow-anchor: none`**，减少浏览器自动锚定 |
| opacity 导致内容发白 | body 同时应用 `opacity: 0→1` 过渡 | 折叠时 description + children 整体变透明，看起来发白 | 去掉 opacity transition，仅保留 grid-rows |
| 性能：grid-template-rows 对复杂表单 | 教育经历含 15+ input + AutoGrowTextarea + YearMonthPicker | grid-template-rows transition 要求浏览器在每帧重新计算布局 | 对大模块可禁用高度动画，改为 fade-only |
| 折叠时内容仍可见一瞬间 | `overflow-hidden` 在 `grid-rows-[0fr]` 的子元素上 | 当 grid row 高度 > 0 但 overflow-hidden 生效前，可能有 1 帧内容泄露 | 对 `grid-rows-[0fr]` 状态也保证 opacity-0 |
| collapsed 状态保留 min-height | ❌ 没有，`0fr` = 0 高度 | 无 | — |

---

## 五、AnimatedListItem 文字卡顿/闪烁诊断

### 5.1 Enter 动画诊断

| 问题 | 证据 | 根因 | 修复建议 |
| -- | -- | -- | -- |
| 首帧闪烁 | `shouldAnimate=false` → `opacity-0 scale-[0.985] translate-y-1.5`，RAF → `shouldAnimate=true` → `opacity-100 scale-100 translate-y-0` | React 18 批量更新可能导致首帧用默认 `opacity-100` 渲染，然后下一帧才切到 `opacity-0` 初始状态 | 去掉 RAF+state 方案，改用 CSS `@keyframes animation` |
| scale 导致文字模糊 | `scale-[0.985]` → `scale-100` 过渡 | `transform: scale(0.985)` 触发 GPU 合成层，亚像素渲染使文字模糊 | **移除 scale**，仅用 `opacity + translateY` |
| `mounted` 未使用 | `useState(false)` + `useEffect(() => setMounted(true), [])` | 声明了但从未在 JSX 使用 | 删除无用代码 |
| max-h 过大 | `max-h-[2000px]` 远大于实际内容（教育经历 ~500px，证书 ~100px） | CSS max-height 过渡上限与内容无关，但 2000px 不造成性能问题 | 可接受，但建议改为 `max-h-[900px]` |

### 5.2 Exit 动画诊断（核心问题）

| 问题 | 证据 | 根因 | 修复建议 |
| -- | -- | -- | -- |
| **文字被挤压卡顿** | `max-h-[2000px]` → `max-h-0` transition 160ms | 内容实际高度 ~500px，前 75% 时间（~300px / 500px）max-h 仍然 > 内容高度 → 无视觉变化。最后 40ms max-h < 内容高度 → 文字被快速挤压 | **改为 measured real height**：挂载时 `ref.getBoundingClientRect().height`，用 inline style `height: ${realH}px` → transition 到 `height: 0` |
| 多个 item 同时删除 layout shift | 删除 2+ item 时，每个 `max-h: 2000px→0` 同时进行 | 页面总高度在多个 animation 期间大幅变化 → scroll 事件 → Navbar 联动 | 用 real height 代替 max-h 2000px |
| 删除后下方 item 上跳 | transition 结束后 item unmount，布局空间瞬时释放 | delayed remove 的 setTimeout 170ms 后真正 filter，但 CSS transition 也是 160ms，两者几乎同时完成 | 理论上时间匹配，实际问题来自 max-h 过大导致实际 collapse 时间不等 |

### 5.3 根因总结

| 问题排名 | 根因 | 影响 |
| -- | -- | -- |
| **P0** | `max-h-[2000px]` 远大于实际内容高度 → exit 动画前 75% 无变化，最后瞬间压缩 | exit 看起来"卡顿不流畅" |
| **P0** | scale transform 导致文字亚像素模糊 | enter 和 exit 期间文字发虚 |
| **P1** | RAF + state enter 动画方案有闪烁风险 | enter 偶现一帧"全显示→隐藏→显示" |
| **P2** | `mounted` state 无用代码 | 轻微维护负担 |

---

## 六、删除与保存安全诊断

| 检查项 | 当前情况 | 风险 | 建议 |
| -- | -- | -- | -- |
| 所有列表进入 removing | ✅ `removingItemKeys` Set 统一管理 | — | — |
| delayed remove 期间布局参与 | ✅ item 仍在 DOM 中（AnimatedListItem 包装），高度从 2000px→0 过渡 | 超大 max-h 导致布局空间不准确 | 改为 real height |
| 保存过滤 removing | ✅ `handleSave` 中 7 类列表全过滤 | — | ✅ |
| timeout 在 unmount 后执行 | ⚠️ **存在风险** | `setTimeout` 在 `markRemoving` 中无清理，如果组件在 170ms 内 unmount，仍会调用 `removeFn` → Warning | 在 useEffect 中返回 cleanup |
| 快速添加/删除 setState warning | ⚠️ **可能** | 连续快速删除多个 item 时，`setRemovingItemKeys` 多次调用 | Set 防重逻辑生效，但 React 18 批量更新可能合并 |
| 删除大量 reflow | ✅ **是** | 多个 item 同时从 max-h:2000px→0 过渡，每帧触发 layout | 改为 real height |

---

## 七、是否使用 motion 判断

| 方案 | 是否推荐 | 理由 |
| -- | -- | -- |
| 继续手写 CSS（修复后） | ✅ **推荐** | 修复方案轻量（real height + 移除 scale + @keyframes enter），无需改变代码范式 |
| 使用 motion AnimatePresence | ❌ 不推荐 | motion 已安装但零使用；引入会改变项目动画范式；AnimatePresence 仍有 layout 问题；增加学习成本 |
| 只用 CSS + measured height | ✅ **推荐（与方案 1 一致）** | 关键修复就是改用 real height |
| 完全取消复杂子项动画 | ❌ 过度 | 问题可修复，不需要取消 |

---

## 八、专业设计方案

### 8.1 Navbar 联动问题

**推荐方案 A + B 组合**：
1. Navbar 增加 **最小 Delta 阈值 30px**（`Math.abs(currentScrollY - lastScrollY) < 30` 时不改变显隐状态）
2. **编辑页路由禁用 Navbar auto-hide**：在 `/resumes/edit` 和 `/resumes/new` 路由上不响应 scroll-direction 隐/显

原因：
- Navbar auto-hide 在原设计是为了节省阅读空间，对**浏览列表型页面**（如岗位列表）有益
- 对**编辑/表单页面**，Navbar 的跳动反而干扰用户操作
- 增加 Delta 阈值可防御大部分布局 shift

### 8.2 父模块折叠动画

**推荐方案 C + D 组合**：
1. **移除 opacity transition**（避免 body 内容"发白"）
2. **保留 grid-rows 动画但 duration 降低为 200ms**（折中性能和不触发过长 scroll 事件）
3. **section 添加 `overflow-anchor: none`**（阻止浏览器 scroll anchoring 自动调整）

### 8.3 子项新增/删除动画

**推荐方案 A + B + C + D 组合**：
1. **移除 scale transform** → 仅保留 `opacity + translateY`
2. **Exit 改用 measured real height** → `ref.getBoundingClientRect().height` → `height: ${realH}px` → transition → `height: 0px`
3. **Enter 改用 CSS @keyframes animation** → 避免 RAF 闪烁
4. **先 fade/translate，再 collapse** → 两阶段 exit：前 80ms 做 opacity+translateY，后 80ms 做 height collapse

---

## 九、下一阶段建议

### K9-UI-ResumeEdit-D-B：编辑页 Navbar 稳定与动效系统重构

**子任务**：

#### 1. Navbar 稳定策略
- Navbar 增加 `delta > 30` 阈值
- 编辑页路由（`/resumes/edit`、`/resumes/new`）禁用 auto-hide

#### 2. 父模块动画修复
- SectionCard 移除 opacity transition
- 添加 `overflow-anchor: none`

#### 3. AnimatedListItem 重写
- 移除 scale transform
- Exit 改用 measured real height → height 0
- Enter 改用 CSS @keyframes
- Exit 分两阶段：fade(80ms) + collapse(80ms)
- 清理无用 mounted state

#### 4. Delayed remove 安全加固
- useEffect cleanup 中清除 timeout

#### 5. 浏览器验收
- 折叠/展开模块时 Navbar 不跳动
- 子项删除文字不卡顿/不模糊
- 快速连续操作不崩溃
- 保存正常

---

## 十、浏览器验收用例

| 用例 | 验收标准 |
| -- | -- |
| 折叠教育经历模块 | Navbar 不会显示/隐藏跳动 |
| 展开教育经历模块 | Navbar 不变 |
| 删除一条教育经历 | 文字淡出→高度平滑收起，不挤压 |
| 删除一条项目经历 | 同上 |
| 快速连续删除 3 条 | 不崩溃，无 setState warning |
| 删除后立刻保存 | 待删项不保存 |
| 模块折叠/展开反复操作 | Navbar 零次异常跳动 |

---

## 十一、未修改范围确认

- 未修改任何源码
- 未修改数据库
- 未修改 `server/sql/init.sql`
- 未 reset-up
- 未执行 Git 命令

---

## 十二、docs 更新清单

| 文件 | 操作 |
| -- | -- |
| `docs/09-开发日志/2026-06-09-669-K9-UI-ResumeEdit-D-A-折叠动画列表动效与Navbar联动异常只读诊断.md` | 新增 |
| `docs/下一步任务.md` | 更新 |
| `docs/变更记录.md` | 更新 |

---

## 十三、Git 约束确认

- [x] 未执行任何 Git 命令
- [x] 未执行 `make reset-up`
- [x] 未修改任何源码/数据库/init.sql
