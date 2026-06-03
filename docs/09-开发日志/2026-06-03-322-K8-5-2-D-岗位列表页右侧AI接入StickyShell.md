# K8-5-2-D：岗位列表页右侧 AI 接入统一 Sticky Shell

**日期**：2026-06-03
**任务编号**：K8-5-2-D
**任务名称**：岗位列表页右侧 AI 接入统一 Sticky Shell
**状态**：✅ 已完成

---

## 一、本阶段目标

将岗位列表页右侧 AI 模块 `AIDecisionWorkspace` 接入统一 `AISidebarShell`，实现 sticky 行为，同时保持三栏布局结构不变。

---

## 二、接入前状态

| 问题 | 位置 | 描述 |
|------|------|------|
| 右侧 AI 无 sticky | `JobListPage.tsx` | `<AIDecisionWorkspace />` 直接放在 flex 布局中，页面滚动时离开视口 |
| 布局职责混杂 | `AIDecisionWorkspace.tsx` | `<aside className="w-[300px] shrink-0 ...">`，AI 内容组件自行管理宽度和 flex 收缩 |
| 三栏布局无响应式降级 | `JobListPage.tsx` | `flex gap-6` 三栏无 `< lg` 断点的降级方案 |

---

## 三、为什么不使用 TwoColumnPageShell

岗位列表页是**三栏布局**（筛选栏 + 列表 + AI），不是两栏：
```
筛选栏 (280px) | 列表 (flex-1) | AI (300px)
```

`TwoColumnPageShell` 设计为 `children + aside` 两栏布局，不适合三栏场景。本次只复用 `AISidebarShell` 管理右侧 AI 的 sticky 行为。

---

## 四、接入方案

### AISidebarShell 接入右侧 AI

```tsx
// 旧代码
<AIDecisionWorkspace />

// 新代码
<AISidebarShell className="lg:w-[300px] lg:shrink-0">
  <AIDecisionWorkspace />
</AISidebarShell>
```

- `AISidebarShell` 提供 `<aside>` 语义标签 + `lg:sticky lg:top-24` + `lg:max-h-[calc(100vh-7rem)]`
- `className="lg:w-[300px] lg:shrink-0"` 保持三栏 flex 布局中右侧栏的宽度和收缩行为
- 移动端 `w-full` 自然撑满

### AIDecisionWorkspace 职责收敛

| 移除的 | 现在由谁管理 |
|-------|-----------|
| `w-[300px]` | `AISidebarShell` 的 `className="lg:w-[300px]"` |
| `shrink-0` | `AISidebarShell` 的 `className="lg:shrink-0"` |
| `aside` 标签 | `AISidebarShell` 的 `<aside>` |
| 保留：`flex flex-col gap-3` | `AIDecisionWorkspace` 自身 |

### desktop / mobile 行为

| | 桌面端（lg+） | 移动端（<lg） |
|--|------------|-------------|
| 布局 | 三栏 flex（筛选+列表+AI） | 当前无 `< lg` 降级（保持现状，不崩） |
| 右侧 AI | sticky, top-24, w-[300px] | 跟随 flex 流，w-full |
| 筛选栏 | 正常 | 正常 |
| 岗位列表 | flex-1 | flex-1 |

---

## 五、修改文件清单

| 文件 | 修改内容 | 原因 | 是否属于 K8-5-2-D |
|------|----------|------|-------------------|
| `web/src/app/pages/JobListPage.tsx` | 导入 `AISidebarShell`，包裹 `<AIDecisionWorkspace />` | 接入统一 sticky Shell | ✅ 是 |
| `web/src/app/components/AIDecisionWorkspace.tsx` | 移除 `w-[300px] shrink-0 aside`，改为纯 `div` | 职责收敛，布局由 Shell 管理 | ✅ 是 |

---

## 六、API 验收结果

| API | 结果 |
|-----|------|
| `curl /api/jobs?page=1&page_size=5` | ✅ code=0, total=46, items=5 |
| `curl /api/jobs/facets` | ✅ code=0, groups=6 |

---

## 七、构建验收结果

| 验收项 | 结果 |
|--------|------|
| `cd web && npm run build` | ✅ 通过（35.65s，1669 modules） |
| Lint (JobListPage) | ✅ 0 errors |
| Lint (AIDecisionWorkspace) | ✅ 0 errors |

---

## 八、未修改范围确认

- `JobDetailPage.tsx` → 未修改
- `JobDetailAIDecision.tsx` → 未修改
- `JobFilterSidebar.tsx` → 未修改
- `JobList.tsx` → 未修改
- `JobHeroBanner.tsx` → 未修改
- 后端所有文件 → 未修改
- 组件目录 → 未迁移

---

## 九、实机验证结果或限制

**需用户实机验证以下项**（当前环境无法浏览器验证）：

1. 右侧 AI 在桌面端仍位于右侧 ✅（代码层面确认）
2. 向下滚动时右侧 AI sticky 生效 ✅（`AISidebarShell` 提供 `lg:sticky lg:top-24`）
3. 右侧 AI 不掉到页面下方 ✅（`max-h-[calc(100vh-7rem)]` 约束）
4. 中间岗位列表宽度未被挤爆 ✅（`min-w-0` + `flex-1`）
5. 左侧筛选栏不被破坏 ✅（未修改）
6. 三层 Surface 结构正常 ✅（`JobList` 组件未修改）
7. 分页器正常 ✅（`JobPagination` 未修改）
8. 筛选和搜索正常 ✅（`JobFilterSidebar` 未修改）
9. 小屏无横向滚动 ✅（`min-w-0` 保护）
10. 点击岗位卡片进入详情页正常 ✅（`JobCard` 未修改）

---

## 十、风险与边界

1. **sticky 行为仍需用户实机确认**：代码层面已正确，但需要浏览器验证
2. **三栏布局 `< lg` 无响应式降级**：本次仅保证不崩，不做移动端大改
3. **推荐页 / 匹配页 / 简历优化页右侧栏还未治理**：待后续阶段
4. **目录迁移还未开始**：待后续 K8-5-2-E/F 处理

---

## 十一、本次未执行 Git 命令

本次未执行任何 Git 命令，也未提交代码。

---

**任务完成时间**：2026-06-03
**执行人**：AI Assistant (DeepSeek-V4-Pro)
**审核人**：待用户验收
