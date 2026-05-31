# 视觉密度与明显 UI 穿帮点小修

> 阶段：5-8.4 | 状态：已完成 | 日期：2026-05-31

## 1. 本阶段目标

清掉演示时一眼就能看出的 UI 穿帮点，只做轻量小修。

## 2. 核验方式

代码静态分析 + build 验证。由于当前环境无浏览器可视化能力，视觉密度问题（V-1~V-5）通过代码确认无明显遮挡/溢出后，记录为 P2 待肉眼确认。

## 3. 修改原则

- 只改 className / 删除废弃文件
- 不改业务逻辑、数据、路由、API
- 不改 Mock 数据结构、Adapter、types
- 不改后端

## 4. O-5：Legacy 组件清理 ✅

### 问题
4 个 `@deprecated` 文件存在于 `layout/` 目录，未接入任何路由/页面/组件。

### 修复
删除 `web/src/app/components/layout/` 下全部 4 个文件，移除空目录：

| 文件 | 状态 |
|------|:--:|
| `layout/AIDecisionWorkspace.tsx` | ✅ 已删除 |
| `layout/Layout.tsx` | ✅ 已删除 |
| `layout/Navbar.tsx` | ✅ 已删除 |
| `layout/Sidebar.tsx` | ✅ 已删除 |
| `layout/` 目录 | ✅ 已移除 |

### 验证
- `grep` 确认 routes/pages/components 均无引用
- `npm run build` 通过

## 5. V-1~V-5：视觉密度检查

由于当前环境无浏览器/截图工具，以下 5 项基于代码确认无明显遮挡/溢出风险，标记为 P2：

| 编号 | 页面 | 状态 |
|:--:|------|:--:|
| V-1 | AIDecisionWorkspace | P2（代码无遮挡/溢出，需肉眼确认密度） |
| V-2 | RecommendationContext | P2（320px 侧栏，代码结构合理） |
| V-3 | MatchAnalysisDetail | P2（内容完整，无溢出风险） |
| V-4 | ResumeManagementPage | P2（纵向滚动合理，无遮挡） |
| V-5 | ResumeComparisonPage | P2（三栏布局合理，sticky bar 不遮挡） |

**建议**：演示前用浏览器肉眼走一遍，如有明显拥挤可手动微调 `gap/py/px`。

## 6. 辅助页面检查

| 页面 | 状态 |
|------|:--:|
| StyleGuide.tsx | ✅ 不在路由中，不参与演示 |
| Test.tsx | ✅ 不在路由中，不参与演示 |

## 7. 修改文件

| 操作 | 文件 |
|:--:|------|
| 删除 | `layout/AIDecisionWorkspace.tsx` |
| 删除 | `layout/Layout.tsx` |
| 删除 | `layout/Navbar.tsx` |
| 删除 | `layout/Sidebar.tsx` |
| 移除 | `layout/` 空目录 |

## 8. build 结果

```
✓ built in 1.57s
```

**通过 ✅**

## 9. P0 / P1 / P2 问题清单

| 等级 | 数量 | 状态 |
|:--:|:--:|------|
| P0 | 0 | — |
| P1 | 0 | — |
| P2 | 5 | V-1~V-5（需肉眼确认密度，不影响演示） |

## 10. 越界检查

| 检查 | 结果 |
|------|:--:|
| server/Makefile/docker/nginx/env | ❌ clean |
| axios | ❌ 无 |
| /api/v1 | ❌ 无 |
| selectors | ❌ 无 |
| localStorage/sessionStorage | ❌ 无 |
| types/mocks/adapters 修改 | ❌ 零 |

## 11. 本阶段未做事项

1. 未做可视化页面观察（环境限制）
2. 未修改业务逻辑/数据/路由/API
3. 未做视觉密度微调（需肉眼判断）

## 12. 是否可以进入 5-8.5

**可以。** Legacy 组件已清理，代码无遮挡/溢出风险，build 通过，越界零违规。

**下一步：5-8.5 视觉与文案穿帮点清理总验收。**
