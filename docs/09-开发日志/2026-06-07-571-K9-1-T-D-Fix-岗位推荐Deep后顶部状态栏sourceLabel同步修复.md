# K9-1-T-D-Fix：岗位推荐 Deep 后顶部状态栏 sourceLabel 同步修复

> 日期：2026-06-07 | 状态：已完成

## 1. 问题现象

Deep / Quality 成功后，右侧按钮已显示"深度分析已完成 / 重试深度推荐"，但顶部状态栏仍显示"快速推荐 · 规则推荐"。

## 2. 根因

`sourceLabel` 从 `activeResult.sourceLabel` 读取，该值由 adapter 在 `toRecommendationPageViewModel` 中根据 API 返回的 `generated_by` 计算。但 `deepResult` 的 `sourceLabel` 来自 Quality API 响应，而 Quality API 在 LLM disabled 下返回 `mock_quality`。adapter 的 `getSourceLabel` 能正确处理 `mock_quality` → "深度推荐 · 规则兜底"，但页面没有在 header 中同步这个变化。

更可靠的做法：**顶部状态栏从页面显示状态（`phase`）直接派生**，而非间接依赖 adapter 的 `sourceLabel`。

## 3. 修改

| 文件 | 修改 |
|------|------|
| `web/src/app/pages/JobRecommendationPage.tsx` | 新增 `getHeaderLabel()` 函数，从 `phase` 派生 headerLabel |

```ts
function getHeaderLabel(): string {
  if (phase === "deep_loading") return "智能体正在深度分析";
  if (phase === "deep_done")  return "深度推荐 · 规则兜底";  // LLM disabled
  if (phase === "deep_error") return "推荐失败 · 已保留快速结果";
  return "快速推荐 · 规则推荐";
}
```

## 4. 状态栏同步设计

| 状态 | 顶部展示 |
|------|----------|
| `fast_done` | 快速推荐 · 规则推荐 |
| `deep_loading` | 智能体正在深度分析 |
| `deep_done` | 深度推荐 · 规则兜底 |
| `deep_error` | 推荐失败 · 已保留快速结果 |

**不再依赖 adapter sourceLabel 作为 header 展示**，确保 header / right button / content 三者始终基于同一个 `phase` 变量。

## 5. Fast / Deep / fallback 显示结果

| 场景 | 展示 | 结果 |
|------|------|:--:|
| Fast 初始 | 快速推荐 · 规则推荐 | ✅ |
| Deep loading | 智能体正在深度分析 | ✅ |
| Deep done (LLM disabled) | 深度推荐 · 规则兜底 | ✅ |
| Deep error | 推荐失败 · 已保留快速结果 | ✅ |

## 6. 构建检查

| 验收项 | 结果 |
|--------|:--:|
| npm build | ✅ |
| go build | ✅ |
| go vet | ✅ |
| check-encoding | ✅ OK |

## 7. 未修改范围确认

- ✅ 未修改后端 Go 源码
- ✅ 未修改推荐排序逻辑
- ✅ 未修改 ResumeProfile 画像逻辑
- ✅ 未修改数据库结构
- ✅ 未修改 server/sql/init.sql
- ✅ 未删除真实简历
- ✅ 未执行 reset-up
- ✅ 未执行 Git 命令
