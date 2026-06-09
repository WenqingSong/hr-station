# K9-Final-UX-B：全站 BulletList 对齐组件与列表样式统一

> 日期：2026-06-09 | 编号：689 | 阶段：K9-Final-UX | 类型：组件统一 + 页面替换

## 一、任务目标

1. 新增统一 `BulletList` / `BulletItem` 组件
2. 替换 4 个报告型页面中所有 Unicode 圆点（●/•）和散写列表
3. 确保单行/多行文本圆点对齐
4. 不改业务文案、不处理画像/标签/backend_engineering

## 二、新增组件

### BulletList.tsx

路径：`web/src/app/components/common/BulletList.tsx`

**API 设计：**

```tsx
type BulletTone = "blue" | "green" | "amber" | "red" | "purple" | "slate";
type BulletSize = "xs" | "sm" | "md";

interface BulletListProps {
  items: Array<string | React.ReactNode>;
  tone?: BulletTone;     // default: "blue"
  size?: BulletSize;     // default: "md"
  className?: string;
  itemClassName?: string;
}
```

**实现要点：**

- 圆点：`<span aria-hidden="true" className="shrink-0 rounded-full bg-current ${toneClass} ${dotSizeClass}" />`
- 对齐：`items-start`（非 items-center）+ `mt-[0.42em]`（md）/ `mt-[0.38em]`（sm）/ `mt-[0.35em]`（xs）
- 文本容器：`min-w-0 flex-1`
- 不使用 Unicode ●/•、不使用 list-inside

**色彩映射：**

| tone | class |
|------|-------|
| blue | text-blue-600 |
| green | text-emerald-600 |
| amber | text-amber-600 |
| red | text-red-600 |
| purple | text-purple-600 |
| slate | text-slate-500 |

**尺寸映射：**

| size | dot | text | mt |
|------|-----|------|----|
| xs | h-1 w-1 | text-[11px] | mt-[0.35em] |
| sm | h-1.5 w-1.5 | text-[12px] | mt-[0.38em] |
| md | h-2 w-2 | text-[13px] | mt-[0.42em] |

## 三、页面替换结果

### ResumeComparisonPage.tsx（6 处替换）

| 模块 | 替换前 | 替换后 | tone |
|------|--------|--------|------|
| 本次迭代结论 (keyDifferences) | ● + flex items-start | BulletList | blue |
| 新版明显提升 (toStrengths) | ● + flex items-start | BulletList | green / sm |
| 新版仍存在的问题 (toRisks) | ● + flex items-start | BulletList | amber / sm |
| 旧版值得保留 (keepFromItems) | ● + flex items-start | BulletList | slate / sm |
| 新版必须保留 (keepToItems) | ● + flex items-start | BulletList | slate / sm |
| 下一版合并方案 (merge+next) | ● + flex items-start | BulletList | green |

### JobRecommendationPage.tsx（2 处替换）

| 模块 | 替换前 | 替换后 | tone |
|------|--------|--------|------|
| 核心匹配点 (matchHighlights) | — | BulletList | green |
| 主要风险 (jobSpecificRisks) | — | BulletList | amber / sm |

### JobMatchAnalysisPage.tsx（2 处替换）

| 模块 | 替换前 | 替换后 | tone |
|------|--------|--------|------|
| 核心优势 (matchHighlights) | — | BulletList | green |
| 相关简历证据 (resumeEvidence) | — | BulletList | green / sm |

### ResumeOptimizationPage.tsx（5 处替换）

| 模块 | 替换前 | 替换后 | tone |
|------|--------|--------|------|
| 快速预览-主要风险 (riskPoints) | — | BulletList | amber / sm |
| 深度报告-风险提醒 (riskPoints) | <ul><li> + ● fragment | BulletList | amber / sm |
| 长期补强建议 (longTermImprovements) | — | BulletList | slate / sm |
| 右侧栏-匹配差距 (matchGaps) | ● Unicode | BulletList | amber / xs |
| 右侧栏-匹配亮点 (matchHighlights) | ● Unicode | BulletList | green / xs |

## 四、Bug 修复

1. **BulletList.tsx** 第44行：`sizeClassMap` → `dotSizeClassMap`（变量名错误导致运行时崩溃）
2. **ResumeOptimizationPage.tsx** 第823-833行：删除残留的旧 `<ul>` 代码片段（导致 JSX 语法错误）
3. **ResumeOptimizationPage.tsx** 右侧栏：替换 2 处 Unicode `●` 为 BulletList

## 五、构建/检查结果

| 检查 | 结果 | 归因 | 是否阻断 |
|------|------|------|----------|
| npm build | ✅ | — | 否 |
| go build | ✅ | — | 否 |
| go vet | ✅ | — | 否 |
| go test | ❌ 6 FAIL | pre-existing K9-2 agent/service | 否 |
| check-encoding | ❌ mojibake | pre-existing init.sql 数据 | 否 |

## 六、Unicode 圆点残留检查

```
grep -Rni "●|•" web/src/app web/src/components
→ NO_UNICODE_BULLETS_FOUND
```

报告型页面零 Unicode 圆点残留。

## 七、未修改范围确认

- 未修改数据库结构
- 未修改 server/sql/init.sql
- 未修改真实简历数据
- 未 reset-up
- 未执行 Git 命令
- 未新增依赖
- 未修改后端业务逻辑
- 未处理画像/标签/backend_engineering 文案
