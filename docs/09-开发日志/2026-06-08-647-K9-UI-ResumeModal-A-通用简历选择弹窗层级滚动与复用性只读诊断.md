# K9-UI-ResumeModal-A：通用简历选择弹窗层级、滚动与复用性只读诊断

> 日期：2026-06-08 | 性质：只读诊断 | 零源码修改

---

## 1. 当前状态确认

| 项 | 结果 |
|----|:--:|
| SelectResumeModal 文件 | `web/src/app/components/SelectResumeModal.tsx`（301 行） |
| 弹窗类型 | 自定义弹窗（非 shadcn Dialog） |
| 使用 Portal | ❌ **否**。渲染在 `JobDetailAIDecision` 组件树内 |
| body scroll lock | ❌ **无**。全项目无 `useLockBodyScroll` 或 `body.style.overflow` |
| ARIA 支持 | ❌ **零**。无 `role="dialog"` / `aria-modal` / `aria-labelledby` / ESC / focus trap |
| 覆盖层 blur | ❌ **无**。`bg-slate-950/30`（30% 不透明度黑色，无 blur） |
| Panel 背景 | `bg-white`（完全 opaque，非半透明） |

---

## 2. 截图问题复盘

| 问题 | 判断 |
|------|------|
| 弹窗白底上能看到背景 slogan/标题 | 🔴 **属实**。根因：**遮罩层 `bg-slate-950/30` 透明度只有 30%，无 `backdrop-blur`，导致背景内容 70% 可见地透过深色遮罩层** |
| 弹窗外仍可滚动 | 🔴 **属实**。全项目无 body scroll lock，弹窗打开后 body 仍可滚动 |
| 弹窗内滚动条贴边 | 🟡 **属实**。`overflow-y-auto` 在内容区域，滚动条贴边 |
| 双滚动（body + modal） | 🔴 **属实**。body 和弹窗内容区域同时可滚 |
| 背景文字"透到"弹窗白底上 | ⚠️ **视觉误判**。Panel 使用 `bg-white`（完全不透明），文字不可能穿透 panel。用户体验到的是遮罩层透明度低 + 无 blur 导致透过 **遮罩区域可清晰看到背景内容**，视觉上形成"背景穿透"印象 |

---

## 3. 文件定位

| 功能 | 文件 | 说明 |
|------|------|------|
| 当前简历选择弹窗 | `web/src/app/components/SelectResumeModal.tsx` | 301 行，独立自定义弹窗 |
| 岗位详情右侧决策组件 | `web/src/app/components/ai/JobDetailAIDecision.tsx` | 237 行，弹窗宿主 |
| 弹窗触发入口 | `JobDetailAIDecision.tsx` L219/227 | `setModalAction("match")` / `setModalAction("optimize")` 触发 |
| 已有 Dialog 基础组件 | `web/src/app/components/ui/dialog.tsx` | shadcn/Radix Dialog，支持 Portal + overlay + content |
| 已有 ConfirmDialog | `web/src/app/components/common/AppConfirmDialog.tsx` | 简单确认弹窗，z-[9999] + backdrop-blur |
| 已有 AuthDialog | `web/src/app/components/auth/AuthDialog.tsx` | 登录弹窗，使用 DialogPortal + backdrop-blur-md |
| Portal 工具 | `DialogPortal` (from `ui/dialog.tsx`) | Radix Portal 包装 |
| body scroll lock 工具 | **❌ 不存在** | 全项目无此能力 |

---

## 4. 当前弹窗结构诊断

| 项 | 当前 | 问题 |
|----|------|------|
| **Props** | `isOpen`, `onClose`, `onConfirm?`, `scene?`, `jobTitle?`, `initialSelectedResumeId?` | 合理，但缺 `selectionMode` |
| **State** | `selected`(number\|null), `resumeOptions`, `loading`, `error`, `submitting`, `searchQuery` | 单数 selected，不支持双选 |
| **selected 逻辑** | 单选（`setSelected(resume.id)`），默认选 `isCurrent` 或第一 | 仅单选 |
| **多选** | ❌ 不支持 | 无法用于简历比较 |
| **内部 API 请求** | ✅ 是。`loadResumeOptionViewModelsFromApi()` 在 `useEffect` | 弹窗内发请求，非纯展示 |
| **耦合度** | `scene` prop 定义了 3 种场景文案 | 与 `JobDetailAIDecision.handlerConfirm` 通过 `onConfirm(resumeId)` 弱耦合 |

---

## 5. 层级 / 毛玻璃诊断

| 项 | 当前 | 问题 |
|----|------|------|
| **遮罩层 z-index** | `z-50` | 🔴 **与 Navbar z-50 相同**，可能有 stacking 冲突 |
| **Panel z-index** | `z-10`（子元素内，实际无效果） | 🟡 子元素 z-index 在父 stacking context 内无意义 |
| **是否 Portal** | ❌ 不 Portal，渲染在 `JobDetailAIDecision > div.flex.flex-col` | 🟡 虽用 `fixed` 定位，但无 parent transform/filter 阻断，当前可工作 |
| **遮罩透明度** | `bg-slate-950/30` | 🔴 **30% 遮罩太弱**，背景内容 70% 可见 |
| **遮罩 backdrop-blur** | ❌ 无 | 🔴 对比 `AuthDialog`（`backdrop-blur-md`）和 `AppConfirmDialog`（`backdrop-blur-md`），本弹窗缺少 blur |
| **Panel 背景** | `bg-white` | ✅ 完全不透明，文字不会真的穿透 panel |

### 5.1 与现有组件对比

| 组件 | Portal | overlay blur | overlay opacity | z-index |
|------|:--:|:--:|:--:|:--:|
| **SelectResumeModal** | ❌ | ❌ 无 | 30% | z-50 |
| **AuthDialog** | ✅ DialogPortal | ✅ `backdrop-blur-md` | 35% | z-50 |
| **AppConfirmDialog** | ❌（但 z-[9999]） | ✅ `backdrop-blur-md` | 30% | z-[9999] |
| **shadcn Dialog** | ✅ DialogPortal | ❌ 无 | 50% | z-50 |

---

## 6. 背景文字透出根因判断

| 根因 | 证据 |
|------|------|
| 🔴 **遮罩层透明度太低 + 无毛玻璃** | `bg-slate-950/30` 只过滤 30% 光线，70% 背景内容可见。`AuthDialog` 同样的 35% 但有 `backdrop-blur-md` 消除了文字可读性 |
| 🟡 **遮罩 z-50 = Navbar z-50** | 理论上可能冲突，但 Navbar 是 `fixed inset-x-0 top-0`，modal overlay 是 `fixed inset-0`，不完全重叠 |
| ✅ **Panel bg-white 不穿透** | 文字不可能穿透 `bg-white`。视觉误判——实际是通过遮罩区域看到背景文字 |

**结论**：不是 panel 透明，是 **overlay 遮罩效果不够强**（30% 黑色 + 无 blur），让用户透过遮罩看到背景 slogan/标题。

---

## 7. 滚动问题诊断

| 项 | 当前 | 问题 |
|----|------|------|
| **body scroll lock** | ❌ 无 | 🔴 弹窗打开后 body 仍可滚动 |
| **body + modal 双滚动** | ✅ 存在 | 🔴 body 和弹窗内容区域同时可滚 |
| **弹窗内容区域滚动** | `overflow-y-auto flex-1 min-h-0` | ✅ 滚动容器隔离正确 |
| **Footer 是否 sticky** | `flex-shrink-0` 固定底部（flex column） | ✅ 不会跟着内容滚动 |
| **弹窗高度限制** | `max-h-[85vh]` + flex column | ✅ 不会超出 viewport |
| **滚动条贴边** | 浏览器默认 scrollbar | 🟡 无美化处理 |

### 7.1 滚动根因

A. ✅ **body 未锁定滚动** — 无 `body.style.overflow = "hidden"`  
B. ❌ overlay 不滚动（fixed 定位）  
C. ❌ panel 已限制 `max-h-[85vh]`  
D. ✅ list 滚动区已隔离（`overflow-y-auto flex-1 min-h-0`）  
E. ❌ footer 为 `flex-shrink-0`，不跟随滚动  
F. 🟡 滚动条无视觉处理

---

## 8. 通用化复用性诊断

| 能力 | 当前支持 | 问题 |
|------|:--:|------|
| 单选简历 | ✅ | — |
| 双选简历 | ❌ | `selected` 为单数 `number\|null`，不可多选 |
| 禁止选择相同简历 | ❌ | 无此概念（仅单选） |
| 初始选中 | ✅ | `initialSelectedResumeId` |
| 当前使用简历同步 | ✅ | 默认选 `isCurrent` |
| 搜索 | ✅ | `searchQuery` + `useMemo` 过滤 |
| 多简历滚动 | ✅ | `overflow-y-auto` |
| loading/error/empty | ✅ | 三种状态全有 |
| 自定义标题/说明 | ❌ | 硬编码「选择用于分析的简历」+ scene 文案 |
| 自定义 confirm 文案 | ❌ | 硬编码「确认并开始分析」 |
| 自定义场景 action | ❌ | `onConfirm` 仅返回 resumeId，行为在调用方 |
| 选择后不路由跳转 | ⚠️ | `onConfirm` 可选，但 `JobDetailAIDecision.handlerConfirm` 会用 `navigate()` |

---

## 9. 现有可复用组件诊断

| 组件 | 是否可复用 | 说明 |
|------|:--:|------|
| `ui/dialog.tsx` (shadcn Dialog) | ✅ 可复用 | Portal + overlay + content + 动画 + Radix 无障碍，但样式可能与需求不同 |
| `auth/AuthDialog.tsx` | ⚠️ 部分 | 有 Portal + blur overlay 经验，但产品样式强绑定 |
| `common/AppConfirmDialog.tsx` | ⚠️ 部分 | 简单 fixed 弹窗，z-[9999] + blur overlay，但无滚动列表 |
| `storage/authStore.tsx` | ❌ | 不相关 |
| Radix Portal | ✅ 可复用 | 项目已安装 `@radix-ui/react-dialog` |

---

## 10. 推荐 Modal 架构

### 层级结构（建议）

```
createPortal(
  <div className="fixed inset-0 z-[1000]">     ← portal 到 body，z 最高层
    <div className="absolute inset-0 bg-slate-950/30 backdrop-blur-sm" />  ← 遮罩 blur
    <div className="relative z-[1001] flex min-h-full items-center justify-center p-4">
      <div className="w-full max-w-2xl rounded-2xl bg-white shadow-2xl border flex flex-col max-h-[min(760px,calc(100vh-64px))]">
        {/* HEADER: shrink-0 */}
        {/* LIST: flex-1 overflow-y-auto min-h-0 */}
        {/* FOOTER: shrink-0 border-t */}
      </div>
    </div>
  </div>,
  document.body
)
```

### 层级要点

| 层级 | z-index | 说明 |
|------|:--:|------|
| Portal container | `z-[1000]` | 高于 Navbar(z-50)、高于 RouteProgressBar(z-9999) 需调整或低于 |
| Overlay | `z-[1000]`（absolute） | 仅在 portal 内相对定位 |
| Panel | `z-[1001]`（relative） | 高于 overlay，确保不会被遮罩盖住 |
| Overlay blur | `backdrop-blur-sm` 或 `backdrop-blur-md` | **必须**：参考 AuthDialog/AppConfirmDialog |
| Overlay 颜色 | `bg-slate-950/40` 或 `bg-slate-950/50` | **建议**：40-50% 比当前 30% 更强 |
| Panel 背景 | `bg-white`（完全不透明） | **保持不变** |
| Panel 不要 blur | ❌ 不要 | 只 overlay blur |

---

## 11. 推荐 ResumePicker 抽象

### 建议方案 B：ResumePickerModal

```typescript
interface ResumePickerModalProps {
  open: boolean;

  // ── Content ──
  title: string;
  description?: string;
  resumes: ResumeOptionViewModel[];   // 数据由调用方注入，不内部请求
  loading?: boolean;
  error?: string | null;
  onRetry?: () => void;

  // ── Selection ──
  selectedIds: number[];
  selectionMode: "single" | "compare";
  maxSelected?: number;               // 默认 selectionMode 决定：single=1, compare=2
  onSelectionChange: (ids: number[]) => void;

  // ── Actions ──
  confirmText: string;
  confirmDisabled?: boolean;
  onConfirm: (ids: number[]) => void;
  onClose: () => void;

  // ── Optional ──
  showSearch?: boolean;               // 默认 true
  slotFooterLeft?: ReactNode;         // 如「管理我的简历」链接
}
```

### 方案对比

| 方案 | 内容 | 是否推荐 | 理由 |
|------|------|:--:|------|
| A：只重构 SelectResumeModal | 就地改，不加 Portal，不改签名 | ❌ | 治标不治本，层级/滚动/通用性不动 |
| **B：ResumePickerModal** | 新通用组件 + Portal + blur + body lock + single/compare | ✅ **强烈推荐** | 解决所有 P0 问题 + 为 compare 预留 |
| C：BaseModal + ResumePickerContent | 拆成更小颗粒 | ⚠️ | 过度工程化，当前场景不需要两层抽象 |

---

## 12. 推荐修复方案

### B. ResumePickerModal 重构

| 步骤 | 内容 | 影响 |
|:--:|------|------|
| 1 | 新建 `web/src/app/components/common/ResumePickerModal.tsx` | 新文件 |
| 2 | 使用 `createPortal` → `document.body` | 脱离组件树 stacking context |
| 3 | overlay: `bg-slate-950/40 backdrop-blur-sm` | 解决背景穿透 |
| 4 | body scroll lock: `useEffect(() => { document.body.style.overflow = "hidden"; return () => ... })` | 解决双滚动 |
| 5 | 添加 `role="dialog" aria-modal="true"` + ESC 关闭 + 点击遮罩关闭 | 解决无障碍 |
| 6 | Footer sticky（`shrink-0 border-t pt-4`） | 解决视觉稳定 |
| 7 | selectionMode: "single" \| "compare" | 预留双选 |
| 8 | 数据由调用方注入，弹窗不再发请求 | 消除内部请求耦合 |
| 9 | `JobDetailAIDecision` 改用新组件 | 替换旧 SelectResumeModal |
| 10 | 旧 `SelectResumeModal.tsx` 保留为 deprecated | 不破坏其他使用者 |

---

## 13. 不建议现在做的内容

| 项 | 原因 |
|----|------|
| Framer Motion 动画库 | 当前 CSS transition 足够 |
| 拖动/排序简历列表 | 超出当前范围 |
| 简历预览/详情在弹窗内 | 过度工程化 |
| 完全重构为 Radix Dialog | 当前自定义 fixed 方案简单可控，Radix 适合标准场景 |

---

## 14. 后续阶段建议

| 阶段 | 名称 | 目标 |
|:--:|------|------|
| **K9-UI-ResumeModal-B** | 通用 ResumePickerModal 重构 | Portal + blur overlay + body lock + single/compare + ARIA |
| K9-UI-ResumeModal-C | 接入 JobDetailAIDecision match/optimize 并验收 | 替换旧弹窗，验证层级、滚动、毛玻璃 |
| K9-UI-ResumeModal-D | 预留简历比较双选能力 | 组件层 support compare mode，页面接入暂缓 |

**推荐**：K9-UI-ResumeModal-B 和 C 可以合并为 1 个执行 prompt（先建新组件，再替换旧引用）。

---

## 15. 是否允许进入执行阶段

- ✅ **通过**：进入 **K9-UI-ResumeModal-B：通用 ResumePickerModal 重构**

---

## 16. 未修改范围确认

- ✅ 未修改任何源码（纯只读诊断）
- ✅ 未修改 K9-1 岗位推荐
- ✅ 未修改 K9-2 岗位匹配分析后端
- ✅ 未修改 K9-3 简历优化建议
- ✅ 未修改 K9-4 简历版本对比
- ✅ 未修改 `server/sql/init.sql`
- ✅ 未修改数据库结构
- ✅ 未删除真实简历
- ✅ 未执行 `make reset-up`
- ✅ 未执行任何 Git 命令

---

## 17. docs 更新清单

| 操作 | 文件 |
|:--:|------|
| ✅ 新增 | `docs/09-开发日志/2026-06-08-647-K9-UI-ResumeModal-A-通用简历选择弹窗层级滚动与复用性只读诊断.md`（本文件） |
| ⬜ 待更新 | `docs/下一步任务.md` |
| ⬜ 待更新 | `docs/变更记录.md` |

---

## 18. Git 约束确认

零 Git 命令执行。
