# K8-5-2-C：岗位详情页接入统一 AI Sticky Shell

**日期**：2026-06-03
**任务编号**：K8-5-2-C
**任务名称**：岗位详情页接入统一 AI Sticky Shell
**状态**：✅ 已完成

---

## 一、本阶段目标

基于 K8-5-2-B 已创建的布局 Shell，将岗位详情页接入 `TwoColumnPageShell` + `AISidebarShell`，修复右侧 AI 模块 sticky 布局。

---

## 二、接入前问题

| 问题 | 位置 | 描述 |
|------|------|------|
| 内联 grid 布局 | `JobDetailPage.tsx` | 手写 `grid grid-cols-[1fr_340px]`，逻辑散落在页面中 |
| 重复 sticky 声明 | `JobDetailPage.tsx` + `JobDetailAIDecision.tsx` | page 层有 `lg:sticky lg:top-24`，AI 组件内部也有 `sticky top-24`，两处同时维护 |
| 布局职责混乱 | `JobDetailAIDecision.tsx` | AI 组件内部含 `w-[340px] shrink-0 sticky top-24 self-start`，承担了页面级布局职责 |
| max-width 不统一 | `JobDetailPage.tsx` | 使用 `max-w-[1200px]`，与 Shell 的 `max-w-6xl` (1152px) 略有差异 |

---

## 三、接入方案

### TwoColumnPageShell 接入方式

```tsx
<TwoColumnPageShell
  contentClassName="space-y-6"
  aside={
    <AISidebarShell>
      <JobDetailAIDecision />
    </AISidebarShell>
  }
>
  <JobDetailJD />
</TwoColumnPageShell>
```

- 移除 `JobDetailPage` 的内联 `grid` 布局
- 移除页面层的 `max-w-[1200px]` wrapper（由 Shell 统一管理）
- 移除页面层的 `lg:sticky lg:top-24`（由 Shell 统一管理）

### AISidebarShell 接入方式

- 将 `JobDetailAIDecision` 包裹在 `AISidebarShell` 中
- Shell 提供 `aside` 语义标签 + `lg:sticky lg:top-24 lg:max-h-[calc(100vh-7rem)] lg:overflow-y-auto`
- 移动端自动降级为 `static` 正常流

### JobDetailAIDecision 职责收敛

| 移除的类 | 原因 | 现在由谁管理 |
|---------|------|-----------|
| `w-[340px]` | 宽度管理 | `TwoColumnPageShell` 的 `xl:grid-cols-[minmax(0,1fr)_340px]` |
| `shrink-0` | flex 收缩控制 | `TwoColumnPageShell` 的 grid 布局 |
| `sticky` | sticky 定位 | `AISidebarShell` |
| `top-24` | top 偏移 | `AISidebarShell` |
| `self-start` | 对齐方式 | `TwoColumnPageShell` 的 `items-start` |
| `aside` | 语义标签 | `AISidebarShell` 的 `<aside>` |

保留的内部样式：`flex flex-col gap-4`（内容内部间距）

### desktop / mobile 行为

| | 桌面端（lg+） | 移动端（<lg） |
|--|------------|-------------|
| 布局 | 两列（JD + AI） | 单列（AI 在 JD 下方） |
| AI 模块 | sticky, top-24, max-height 约束 | static 正常流 |
| 宽度 | 右侧栏 340px (xl) / 320px (lg) | 100% |

---

## 四、修改文件清单

| 文件 | 修改内容 | 原因 | 是否属于 K8-5-2-C |
|------|----------|------|-------------------|
| `web/src/app/pages/JobDetailPage.tsx` | 接入 `TwoColumnPageShell` + `AISidebarShell`，移除内联 grid 布局和 sticky 逻辑 | 统一布局规范 | ✅ 是 |
| `web/src/app/components/JobDetailAIDecision.tsx` | 移除 `w-[340px] shrink-0 sticky top-24 self-start aside`，改为纯 `div` | 职责收敛，布局由 Shell 管理 | ✅ 是 |

---

## 五、API 验收结果

| API | 结果 |
|-----|------|
| `curl "http://localhost:8080/api/jobs/10001"` | ✅ code=0, fields 完整 |

---

## 六、构建验收结果

| 验收项 | 结果 |
|--------|------|
| `cd web && npm run build` | ✅ 通过（1.72s，1669 modules） |
| Lint (JobDetailPage) | ✅ 0 errors |
| Lint (JobDetailAIDecision) | ✅ 0 errors |

---

## 七、未修改范围确认

- `JobListPage.tsx` → 未修改
- `JobList.tsx` → 未修改
- `AIDecisionWorkspace.tsx` → 未修改
- `JobFilterSidebar.tsx` → 未修改
- 后端所有文件 → 未修改
- 其余页面 → 未修改

---

## 八、实机验证结果或限制

**需用户实机验证以下项**（当前环境无法浏览器验证）：

1. 右侧 AI 投递决策在桌面端位于右侧栏 ✅（代码层面确认）
2. 主 JD 内容和右侧 AI 顶部对齐 ✅（Shell 使用 `items-start`)
3. 向下滚动时右侧 AI sticky 生效 ✅（`AISidebarShell` 提供 `lg:sticky lg:top-24`）
4. 右侧 AI 不遮挡 header ✅（`lg:top-24` = 96px，Navbar 高 56px）
5. 右侧 AI 不掉到页面下方 ✅（`lg:max-h-[calc(100vh-7rem)]` 约束）
6. 小屏下 AI 在主内容下方 ✅（`grid-cols-1` 降级）
7. 页面无横向滚动条 ✅（`min-w-0` 保护）
8. 返回岗位列表正常 ✅（Banner 中 `navigate("/jobs")` 未改动）

---

## 九、风险与边界

1. **sticky 行为仍需用户实机确认**：代码层面已正确，但需要浏览器验证滚动行为
2. **`max-w-6xl` (1152px) vs 旧 `max-w-[1200px]` (1200px)**：宽度略有收窄，视觉上更居中
3. **岗位列表页右侧 AI 还未接入统一 Shell**：待 K8-5-2-D 处理
4. **其他页面右侧 AI 还未治理**：待后续阶段处理

---

## 十、本次未执行 Git 命令

本次未执行任何 Git 命令，也未提交代码。

---

**任务完成时间**：2026-06-03
**执行人**：AI Assistant (DeepSeek-V4-Pro)
**审核人**：待用户验收
