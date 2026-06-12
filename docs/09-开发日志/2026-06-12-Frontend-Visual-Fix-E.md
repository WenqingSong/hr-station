# 2026-06-12 Frontend-Visual-Fix-E：修复推荐页残留技术化标签文案

## 修复背景

`Frontend-Visual-Acceptance-D` 验收确认：全站前端整体通过，唯一残留问题是 `JobRecommendationPage.tsx` 中 `getHeaderLabel()` 函数内有 2 处用户可见的技术化标签：

- `"规则兜底"` → 当 Deep LLM 不可用时显示
- `"规则推荐"` → 快速推荐默认标签

这些标签在 Sweep-C 阶段已从 adapter 层修复（`recommendationAdapter.ts` / `common.ts`），但页面内 `getHeaderLabel()` 直接硬编码了旧文案，未与 adapter 同步更新。

## 修改内容

### JobRecommendationPage.tsx（第 290、295 行）

| 位置 | 修改前 | 修改后 |
|---|---|---|
| deep_done / llm 失败分支 | `"深度推荐 · 规则兜底"` | `"深度推荐 · 基础模型"` |
| deep_done / llm 成功分支 | `"深度推荐 · AI生成"` | `"深度推荐 · AI 生成"` |
| default（快速推荐） | `"快速推荐 · 规则推荐（N 个岗位）"` | `"快速推荐 · 基础模型（N 个岗位）"` |

### 代码对比

```tsx
// 修改前
const base = activeGeneratedBy?.startsWith("llm") 
  ? "深度推荐 · AI生成" 
  : "深度推荐 · 规则兜底";
...
default: return `快速推荐 · 规则推荐（${actualCount} 个岗位）`;

// 修改后
const base = activeGeneratedBy?.startsWith("llm") 
  ? "深度推荐 · AI 生成" 
  : "深度推荐 · 基础模型";
...
default: return `快速推荐 · 基础模型（${actualCount} 个岗位）`;
```

## 技术化文案全站复查

修改后扫描确认：

| 关键词 | 用户可见 UI | 状态 |
|---|---|---|
| `规则兜底` | 0 | ✅ 已清除 |
| `规则推荐` | 0 | ✅ 已清除 |
| `mock_agent` / `mock_quality` | 仅在 util/注释中 | ✅ 无 UI 泄漏（cleanUserText.ts 有防御层） |
| `fallback` 用户可见 | 0 | ✅ 无 UI 泄漏 |

## 构建验证

| 项目 | 结果 |
|---|---|
| `npm build` (vite) | ✅ 通过（1777 modules，2.51s） |
| `go build` | ✅ 通过 |
| `go vet` | ✅ 通过 |
| `go test` | ✅ 全部通过 |
| `make check` | ✅ 全部通过 |

## 页面验证

| URL | 结果 |
|---|---|
| `/recommend?resumeId=8` 8080 | ✅ 637 bytes SPA shell |
| `/recommend?resumeId=8` 5173 | ✅ 637 bytes SPA shell |
| 8080 vs 5173 | ✅ MD5 一致 |

## 修改原则

- 仅改 2 行文案，不改变任何业务逻辑
- 不改变 `activeGeneratedBy?.startsWith("llm")` 判断
- 不改变 API、Adapter、ViewModel
