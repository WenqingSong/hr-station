# Layout / Navbar 基础组件统一

> 阶段：5-3.1 | 状态：已完成 | 日期：2026-05-31

## 1. 本阶段目标

统一全局 Layout / Navbar 生效路径，确保项目只有一套全局 Navbar，新增 PageContainer 统一组件，为后续路由整理和页面骨架统一打基础。

## 2. 改造前 Navbar / Layout 状态

| 文件 | 状态 |
|------|------|
| `components/Navbar.tsx` | **活跃**（routes.tsx RootLayout 引用） |
| `components/layout/Navbar.tsx` | **冗余**（layout/Layout.tsx 引用，但 Layout 未接入路由） |
| `components/layout/Layout.tsx` | **未使用**（定义含 Navbar 的布局但未在路由中使用） |
| `components/layout/AIDecisionWorkspace.tsx` | **冗余**（与 components/AIDecisionWorkspace.tsx 1:1 副本） |
| `components/layout/Sidebar.tsx` | **未使用**（仅 StyleGuide 引用） |

## 3. 保留的唯一 Navbar

**文件**：`web/src/app/components/Navbar.tsx`（routes.tsx 中 RootLayout 引用）

**内容**：
```
左侧：「Offer 捕手」学生求职匹配智能体
右侧：我的简历
```

完全符合冻结规则。无禁止导航项。

## 4. RootLayout 调整

`routes.tsx` 中 RootLayout 无需修改（已经在改造前就是干净状态）：

```tsx
function RootLayout() {
  return (
    <div className="flex min-h-screen flex-col bg-[#F5F7FA] font-sans text-slate-900">
      <Navbar />
      <main className="flex-1">
        <Outlet />
      </main>
    </div>
  );
}
```

- ✅ 只渲染一套全局 Navbar
- ✅ 提供全局背景和上下结构
- ✅ 无业务逻辑
- ✅ 无 AI 决策工作台

## 5. PageContainer

**新增**：`web/src/app/components/PageContainer.tsx`

可配置参数：
- `maxWidth` — 页面最大宽度（默认 `1360px`）
- `paddingY` / `paddingX` — 页面内边距
- `className` — 自定义 className

职责：仅提供布局容器，不管理业务内容。后续 5-3.3/5-3.4 业务页可大规模接入。

## 6. 冗余组件处理策略

| 文件 | 策略 | 原因 |
|------|------|------|
| `components/layout/Navbar.tsx` | 标记 `@deprecated`，保留文件 | 与活跃 Navbar 1:1 副本，仅 layout/Layout.tsx 内部引用 |
| `components/layout/Layout.tsx` | 标记 `@deprecated`，保留文件 | 未接入路由，不被任何活跃页面引用 |
| `components/layout/AIDecisionWorkspace.tsx` | 标记 `@deprecated`，保留文件 | 与 components/AIDecisionWorkspace.tsx 重复 |
| `components/layout/Sidebar.tsx` | 标记 `@deprecated`，保留文件 | 与 JobFilterSidebar 功能重复 |
| `pages/StyleGuide.tsx` | import 修正 | 从 `layout/Sidebar` / `layout/AIDecisionWorkspace` → 活跃组件 |

所有已废弃文件均保留，后续清理阶段（5-8 或后续）可安全删除。

## 7. 未做事项

```text
未改路由路径（留到 5-3.2）
未接 API
未改页面主流程
未改后端
未改 types/mocks/adapters/api
未新增 selectors
未新增 axios
未引入状态管理库
未重写页面视觉
未接入 SelectResumeModal 主流程
未删除 layout/ 目录文件（deprecated 保留）
```

## 8. 5-3.2 输入

```text
1. 路由路径规范化：/recommendations → /recommend，/match-analysis → /match，/job/:id → /jobs/:id
2. 页面内 min-h-screen 容器移除
3. PageContainer 在部分页面先接入验证
```
