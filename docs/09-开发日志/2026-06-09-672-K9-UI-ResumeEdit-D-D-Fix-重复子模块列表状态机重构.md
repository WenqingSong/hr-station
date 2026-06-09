# K9-UI-ResumeEdit-D-D-Fix：重复子模块列表状态机重构

> 日期：2026-06-09 | 性质：前端状态机重构，未修改后端/数据库

---

## D-C-Fix 失败复盘

| 问题 | 原因 |
| -- | -- |
| D-C-Fix 仍然出现裸露态 | 空状态条件 `list.length === 0` 在最后一项正在 exiting 时仍为 false → 列表容器渲染 + add button 可见 |
| 根因 | 状态机缺 `transitioningToEmpty` 态 — 当 `activeItems.length === 0 && removingItems.length > 0` 时应提前渲染 empty state |

## 修复策略

**关键变化**：empty state 不再等 `items.length === 0` 才渲染。当 `active items` 全部清空（全部在 exiting），立刻渲染 empty state，让退出中的旧卡片和 empty state 共存于舞台。

## 状态机

```
hasActive? → 渲染列表 + add button
allRemoving? → 渲染 exiting items + empty state（add button 只在 empty state 内）
isEmpty? → 渲染 empty state
```

## 实现

每个列表用 IIFE 计算 `hasActive` / `hasRemoving`：

```tsx
{(() => {
  const hasActive = items.some(i => !removingItemKeys.has(key(i)));
  const hasRemoving = items.some(i => removingItemKeys.has(key(i)));
  if (items.length === 0 || (!hasActive && hasRemoving)) return <ResumeEditorEmptyState ... />;
  return (list + add btn);
})()}
```

## 新增组件

| 文件 | 说明 |
| -- | -- |
| `web/src/app/components/resume-editor/ResumeEditorEmptyState.tsx` | 统一空状态卡片（min-h-[180px]，bg-slate-50/80） |

## 7 类列表覆盖

| 模块 | 状态机 | 结果 |
| -- | -- | -- |
| 教育经历 | IIFE + ResumeEditorEmptyState | ✅ |
| 项目经历 | 同上 | ✅ |
| 实习经历 | 同上 | ✅ |
| 证书与获奖 | IIFE 双数组计算 | ✅ |
| 作品链接 | IIFE + ResumeEditorEmptyState | ✅ |
| 外语考试 | IIFE + ResumeEditorEmptyState | ✅ |

## 构建

| 检查 | 结果 |
| -- | -- |
| npm build | ✅ 649 kB |
| go build | ✅ |
| go vet | ✅ |
| lint | ✅ |
