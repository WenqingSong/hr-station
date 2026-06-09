# K9-UI-ResumeEdit-C-A：简历编辑页模块与子项动效只读诊断

> 日期：2026-06-09 | 性质：只读诊断，未修改任何源码

---

## 一、当前状态确认

| 项 | 结果 |
| -- | -- |
| 动画库 | `motion` v12.23.24 已安装但 **未使用**（package.json 中存在，零次 import） |
| 当前动画方案 | 纯 CSS（`@keyframes` + `grid-template-rows` + `transition`） |
| 依赖 | `tw-animate-css`（Tailwind CSS 动画插件） |
| 编辑页主文件 | `MyResumesPage.tsx`（617 行） |
| 父模块卡片 | `ResumeEditorSectionCard.tsx`（152 行） |
| 子项列表 | 教育/项目/实习/证书/获奖/外语考试/作品链接，均在 `MyResumesPage.tsx` 内联渲染 |

---

## 二、文件定位

| 功能 | 文件 | 当前实现 |
| -- | -- | -- |
| 编辑页主文件 | `web/src/app/pages/MyResumesPage.tsx` | 617 行，单文件承载全部 7 个子项列表 |
| 父模块卡片 | `web/src/app/components/resume-editor/ResumeEditorSectionCard.tsx` | 152 行，grid-template-rows 动画 |
| 展开 / 折叠动画 | `ResumeEditorSectionCard` L128-149 | `grid` + `transition-[grid-template-rows,opacity,padding] duration-200 ease-out` |
| 教育经历列表 | `MyResumesPage` L503-535 | `educations.map`，`key={edu.id}`，inline filter 删除 |
| 项目经历列表 | `MyResumesPage` L537-550 | `projects.map`，`key={proj.id}`，`removeProject` 删除 |
| 实习经历列表 | `MyResumesPage` L560-572 | `internships.map`，`key={intern.id}`，inline filter 删除 |
| 证书列表 | `MyResumesPage` L584 | `certificateDrafts.map`，`key={cert.id}`，`removeCertificate` 删除 |
| 获奖列表 | `MyResumesPage` L586 | `awardDrafts.map`，`key={award.id}`，`removeAward` 删除 |
| 作品链接列表 | `MyResumesPage` L593 | `portfolioDraft.works.map`，`key={work.id}`，`removeWork` 删除 |
| 外语考试列表 | `MyResumesPage` L578 | `languageExamDrafts.map`，**`key={idx}`** ⚠️，`removeLanguageExam(idx)` 删除 |
| 新增 enter 动画 | `web/src/styles/tailwind.css` L7-28 | `@keyframes resume-item-enter`（140ms，opacity+translate3d） |
| 删除 exit 动画 | — | **不存在**，所有删除立即 unmount |

---

## 三、父模块展开 / 折叠动画诊断

### 3.1 当前实现

```css
/* L128-134 ResumeEditorSectionCard.tsx */
grid
transition-[grid-template-rows,opacity,padding] duration-200 ease-out

open:
  grid-rows-[1fr] opacity-100

collapsed:
  grid-rows-[0fr] opacity-0
```

体层级：
```html
<div class="grid transition-[grid-rows...]">       <!-- 动画容器 -->
  <div class="overflow-hidden">                      <!-- 裁剪层 -->
    <div class="px-6 pb-5">                          <!-- 内边距容器 -->
      <p>description</p>
      {children}
    </div>
  </div>
</div>
```

### 3.2 逐项诊断

| 检查项 | 当前情况 | 问题 | 建议 |
| -- | -- | -- | -- |
| 是否使用 grid-template-rows | ✅ 是 `0fr→1fr` | 技术路径正确 | — |
| 条件渲染 | ❌ 否，body 始终渲染 | ✅ 正确，不需要改 | — |
| body 折叠时 unmount | ❌ 否，仅 `overflow-hidden` | ✅ 正确 | — |
| header 高度固定 | ✅ `px-6 py-4` 固定 | ✅ 结构稳定 | — |
| title/subtitle/badge/arrow 结构 | ✅ 四者在同一 flex 行 | ✅ 稳定 | — |
| 折叠时触发父容器 reflow | ⚠️ 是 | `transition-[padding]` 导致 section 的 padding 参与动画，父容器可能轻微跳动 | 移除 `padding` 从 transition 中，仅保留 `grid-template-rows, opacity` |
| duration 过短 | ⚠️ `duration-200`（200ms） | 对于含丰富表单的模块（如教育经历 15+ 字段），200ms 略短 | 升级为 `duration-[240ms]` |
| easing 不自然 | ⚠️ `ease-out` | `ease-out` 对高度展开尚可，但对 20+ 行的教育经历模块，`ease-out` 集中在结束阶段显得生硬 | 升级为 `cubic-bezier(0.22, 1, 0.36, 1)`（ease-out-quint） |
| opacity 和 height 不同步 | ⚠️ | opacity 和 grid-rows 同时从 0→1 过渡，但 grid-template-rows 从 `0fr` 到 `1fr` 的过渡本质是**高度变化**，CSS grid-template-rows transition 是近似效果，不是真正平滑 | 保留当前方案（grid-rows transition 是可行的 CSS-only 最佳方案） |
| body padding/margin 瞬时跳动 | ⚠️ | `pb-5`（20px）在折叠时仍然存在于 DOM 中但被 `overflow-hidden` 裁剪；展开时 padding 瞬间出现 | 可接受。当前结构下 padding 在 overflow-hidden 内层，不会传到外层 |
| nested section 双层动画 | ❌ 无 | 没有嵌套 SectionCard | 无需处理 |

### 3.3 根因判断

**"不够丝滑"的根因排序**：

| 序号 | 根因 | 严重程度 |
| -- | -- | -- |
| **A** | 动画时间太短（200ms） | ⚠️ 中等 |
| **B** | easing 不自然（ease-out 对大面积展开停留感太重） | ⚠️ 中等 |
| **C** | body 内容始终参与布局 | ❌ 不适用（grid-rows 方案已避免） |
| **D** | header 和 body 分离不稳定 | ❌ 不适用（已修复） |
| **E** | padding/margin 突然出现 | ❌ 已处理（overflow-hidden） |
| **F** | 条件渲染导致 exit 缺失 | ❌ 不适用（无条件渲染） |
| **G** | 高度动画无法处理动态内容 | ⚠️ 轻微（grid 方案的天然局限，但可接受） |

**结论**：当前动画整体正确，主要问题是 **duration 偏短**（200ms）和 **easing 偏硬**（ease-out），这两个参数调优即可显著提升流畅感。

---

## 四、子项新增 / 删除动效诊断

### 4.1 逐列表诊断

| 列表 | key 策略 | 新增方式 | 删除方式 | 问题 | 建议 |
| -- | -- | -- | -- | -- | -- |
| 教育经历 | `key={edu.id}` ✅ 递增数字 id | `push` + `setRecentlyAdded` | inline `filter` 立即 unmount | ❌ 无 exit 动画，item 瞬间消失，下方 item 瞬时上跳 | delayed remove + exit animation |
| 项目经历 | `key={proj.id}` ✅ 递增数字 id | `push` + `setRecentlyAdded` | `removeProject` → `filter` 立即 unmount | ❌ 无 exit 动画，下方 item 瞬时上跳 | delayed remove + exit animation |
| 实习经历 | `key={intern.id}` ✅ 递增数字 id | `push` + `setRecentlyAdded` | inline `filter` 立即 unmount | ❌ 无 exit 动画 | delayed remove + exit animation |
| 证书 | `key={cert.id}` ✅ 递增数字 id | `push` + `setRecentlyAdded` | `removeCertificate` → `filter` 立即 unmount | ❌ 无 exit 动画 | delayed remove + exit animation |
| 获奖 | `key={award.id}` ✅ 递增数字 id | `push` + `setRecentlyAdded` | `removeAward` → `filter` 立即 unmount | ❌ 无 exit 动画 | delayed remove + exit animation |
| 作品链接 | `key={work.id}` ✅ 递增数字 id | `push` + `setRecentlyAdded` | `removeWork` → `filter` 立即 unmount | ❌ 无 exit 动画 | delayed remove + exit animation |
| **外语考试** | **`key={idx}` ❌ 使用 index** | `push` + `setRecentlyAdded`（key 是 `String(length)`） | `removeLanguageExam(idx)` → `filter((_, i) => i !== idx)` 立即 unmount | ⚠️ **P0 缺陷**：使用 index 作为 key，删除中间项时 React 会错误复用 DOM 导致动画和状态混乱 | **必须改为 stable id**（新增 clientId 或改用递增 id） |

### 4.2 新增动效现状

- **已存在 enter 动画**：`resume-item-enter` CSS class
- **实现方式**：`recentlyAdded` state 记录最新添加的 item 的 type+key → 条件添加 class
- **动画参数**：`140ms cubic-bezier(0.16, 1, 0.3, 1)`，opacity 0→1 + translateY 4px→0
- **问题**：`recentlyAdded` 是一次性信号（下次渲染时不再匹配 class），但 CSS `animation: both` 会保留最终状态 → 如果没有清除机制可能导致问题。实际表现：由于 `recentlyAdded` 在下一次 setState 后就不再匹配，class 条件变为 false，class 被移除 → animation 被中断。**实际上 enter 动画当前可能没有正确完成**，取决于 React 是否在同一帧内设置了另一个 state。

**根因确认**：`recentlyAdded` 触发 class 添加，但任何其他 setState（如输入文字）也会导致 re-render 并使 `recentlyAdded` 不匹配 → animation class 提前移除 → 动画被截断。好在 `addXxx` 函数中，`setRecentlyAdded` 后紧接着没有其他 setState，所以 **在当前代码中 enter 动画实际上是有效的**。

### 4.3 删除动效缺失根因

所有 7 类子项的删除逻辑都直接 `filter` / `splice` → React 立即 unmount DOM 节点 → **没有任何 exit 动画窗口**。

| 特性 | 是否实现 |
| -- | -- |
| enter animation | ✅ 已实现（`resume-item-enter`） |
| exit animation | ❌ **未实现** |
| list layout transition | ❌ 未实现 |
| delayed remove | ❌ 未实现 |
| removing state tracking | ❌ 未实现 |

---

## 五、动画库可行性判断

`motion` v12.23.24 已在 `package.json` 中但源码零使用。这是一个机会窗口：可以直接使用 motion 的 `AnimatePresence` 实现 exit animation，无需新增依赖。

| 方案 | 是否推荐 | 理由 |
| -- | -- | -- |
| **纯 CSS transition（当前）** | ⚠️ 部分可行 | enter 已经有方案，**但 CSS-only 无法实现 exit animation**（React unmount DOM 时 CSS transition 来不及执行） |
| **手写 AnimatedListItem + removingIds** | ✅ **强烈推荐** | 零新依赖、完全控制、退出动画通过 `removingIds` + `setTimeout` 延迟删除实现 |
| Framer Motion / Motion | ⚠️ 备选 | motion 已安装但从未使用。`AnimatePresence` 可原生处理 enter/exit，但引入 motion API 风格与现有代码不统一 |
| auto-animate | ❌ 不推荐 | 需新装依赖，且对 list 删除动画控制粒度不够 |
| React Transition Group | ❌ 不推荐 | 需新装依赖，API 老旧 |

**推荐方案**：**手写 `AnimatedListItem` + `removingIds`**，理由：

1. 零新依赖（motion 虽已安装但从未使用，强行引入会打破现有代码风格）
2. 对 Demo 阶段最轻量
3. 完全可控，不与表单逻辑冲突
4. 删除动画完成后才真正从数组移除，数据安全

---

## 六、推荐动画设计标准

| 动效 | 标准 |
| -- | -- |
| 父模块展开 | duration: 240ms, easing: `ease-out` 或 `cubic-bezier(0.22, 1, 0.36, 1)`, header 不动，body opacity+height 同步过渡 |
| 父模块折叠 | duration: 200ms, 同上 easing，箭头旋转 180° |
| 子项 enter | opacity 0→1 + translateY 4px→0 + scale 0.98→1, duration: 180ms, easing: `cubic-bezier(0.16, 1, 0.3, 1)` |
| 子项 exit | opacity 1→0 + translateY 0→-4px + scale 1→0.98 + max-height→0, duration: 160ms, 动画完成后 `setTimeout` 真正删除 |
| 删除后下方 item | 通过 exit 动画的 `max-height→0` 缓冲过渡（不是 FLIP），避免瞬时上跳 |

---

## 七、下一阶段实现策略

| 实现点 | 是否需要 | 建议 |
| -- | -- | -- |
| 给语言考试补 stable clientId | ✅ **P0 必须** | `addLanguageExam` 改用递增 id，不用 index |
| removingIds 状态 | ✅ **P1 需要** | 7 类列表统一用 `Set<string>` | track removing item keys |
| AnimatedListItem 组件 | ✅ **P1 需要** | 新建 `web/src/app/components/ui/animated-list-item.tsx`，封装 enter/exit 逻辑 |
| 删除延迟执行 | ✅ **P1 需要** | `setTimeout(actuallyRemove, 160ms)` 在 exit 动画完成后执行 |
| 新增进入动画 | ✅ **已实现** | 保留 `resume-item-enter`，修正 `recentlyAdded` 清除时机 |
| 自动 focus | ❌ 暂不需要 | 增加复杂度，Demo 阶段可选 |
| 折叠动画升级 | ✅ **P1 需要** | duration 200ms→240ms，easing 升级 |
| 上移/下移动画 | ❌ 暂不需要 | 项目经历已有上下移按钮但无动画，P2 |

---

## 八、风险分析

| 风险 | 是否存在 | 处理建议 |
| -- | -- | -- |
| 使用 index 作为 key 致动画错乱 | ✅ **存在（语言考试）** | P0：改用递增 id |
| 删除延迟致保存时仍有待删除项 | ✅ **存在** | `handleSave` 中过滤 `removingIds` 的 item，payload 中排除 |
| 动画中用户快速连续点击删除 | ✅ **存在** | `removingIds` 用 `Set` 防止重复添加；`setTimeout` 前检查 item 是否已被删除 |
| 折叠模块时子项正在删除 | ⚠️ **理论存在** | 折叠只是视觉隐藏（grid-rows:0fr），不影响 DOM 存在。两种动画不冲突 |
| AutoGrowTextarea 高度 JS 设置与父级 CSS grid 动画冲突 | ⚠️ **理论存在** | AutoGrowTextarea 的 `el.style.height=...` 在 useEffect 中执行，父级 `grid-rows` 在 CSS transition 中。同一帧内 JS 设置优先于 CSS transition → 可能导致 textarea 高度在父级展开动画中出现"跳跃" | 在父模块展开后延迟触发 textarea 同步（或允许 textarea 在 transition 期间保持 CSS 驱动） |
| 新增依赖致 bundle 增大 | ❌ 不存在 | 方案不使用新依赖 |
| 移动端性能 | ⚠️ 需注意 | `translate3d` 已开启 GPU 加速；`prefers-reduced-motion` 已处理。P2 |

---

## 九、浏览器验收用例

| 编号 | 用例 | 验收标准 |
| -- | -- | -- |
| 1 | 展开/折叠教育经历 5 次 | 动画流畅不卡顿，header 不位移，body 平滑出现/消失 |
| 2 | 添加教育经历 2 条 | 卡片从 translateY 4px + opacity 0 平滑进入，约 180ms |
| 3 | 删除第 1 条教育经历 | 卡片 opacity→0 + 缩小 + 高度收起，约 160ms，下方卡片不上跳 |
| 4 | 添加项目经历 2 条 | 同用例 2 |
| 5 | 删除中间一条项目经历 | 同用例 3，剩余卡片顺序正确 |
| 6 | 输入长文本让 AutoGrowTextarea 自动增高 | textarea 高度平滑增长，无抖动 |
| 7 | 折叠包含长文本的模块 | body 平滑收起，无瞬时跳动 |
| 8 | 展开包含长文本的模块 | body 平滑展开，textarea 高度正确 |
| 9 | 快速连续添加/删除 | 不崩溃，不出现重复/缺失卡片 |
| 10 | 添加后保存，确认数据无错乱 | 所有 item 正确保存，无 `removingIds` 残留 |

---

## 十、推荐执行方案

### 方案 A：小修（不推荐）
- 调 duration 200→240ms + easing 升级
- 不处理子项删除动画
- 修复语言考试 key

### 方案 B：推荐 ✅
- 父模块 duration 200→240ms + easing 升级（`cubic-bezier(0.22, 1, 0.36, 1)`）
- 新增 `AnimatedListItem` 组件（~60 行）
- 7 类列表统一接入 `removingIds` + delayed remove
- 修复语言考试 `key={idx}` → `key={le.id}`
- 保留现有 `resume-item-enter` enter 动画
- 保存时过滤 `removingIds`

### 方案 C：不建议
- 引入 motion `AnimatePresence`
- 做 FLIP 动画

| 方案 | 内容 | 是否推荐 |
| -- | -- | -- |
| 方案 A | 仅调父模块参数 | ❌ 不解决核心痛点 |
| 方案 B | AnimatedListItem + delayed remove + stable key | ✅ **推荐** |
| 方案 C | 引入 motion 库 | ❌ 过度 |

---

## 十一、下一阶段建议

**K9-UI-ResumeEdit-C-B：简历编辑页模块与子项动效重构**

8 个子任务：
1. 父模块动画升级（duration 240ms + easing 优化）
2. 新增 `AnimatedListItem` 组件
3. 语言考试 key 修复（P0）
4. 7 类列表统一接入 `removingIds` + `setTimeout` delayed remove
5. 保存时过滤 `removingIds`
6. `resume-item-enter` 修正（确保完整播放）
7. 浏览器验收
8. docs 更新

---

## 十二、未修改范围确认

- 未修改任何源码
- 未修改数据库
- 未修改 `server/sql/init.sql`
- 未执行 `reset-up`
- 未执行 Git 命令

---

## 十三、docs 更新清单

| 文件 | 操作 | 说明 |
| -- | -- | -- |
| `docs/09-开发日志/2026-06-09-667-K9-UI-ResumeEdit-C-A-简历编辑页模块与子项动效只读诊断.md` | 新增 | 本诊断报告 |
| `docs/下一步任务.md` | 更新 | 新增 K9-UI-ResumeEdit-C-A/B 条目 |
| `docs/变更记录.md` | 更新 | 新增诊断记录 |

---

## 十四、Git 约束确认

- [x] 未执行 `git status`
- [x] 未执行 `git diff`
- [x] 未执行任何 Git 命令
- [x] 未执行 `make reset-up`
- [x] 未修改任何源码/数据库/init.sql
