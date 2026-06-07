# K9-4-B 简历版本对比页面状态机与 Fast/Deep 展示分层设计

**日期**: 2026-06-07  
**阶段**: 设计冻结（K9-4-B）  
**关联**: K9-4-A（只读诊断）、K9-1-U-B（推荐页状态机参考）、K9-3-B（优化页状态机参考）

---

## 1. 页面定位

```
Route: /resumes/compare
Component: ResumeComparisonPage.tsx（464行 → 重建数据流）
```

帮助学生：
1. 比较两份简历，判断哪一版更适合投递
2. 看清差异、优势、风险和修改方向
3. 决定是否合并两版优点生成下一版

---

## 2. 状态机设计

### 2.1 Phase 枚举

```ts
type ComparePhase =
  | "loading_latest"      // 读取最新对比结果
  | "fast_loading"         // 正在生成 Fast 对比
  | "fast_done"            // Fast 对比完成
  | "fast_error"           // Fast 对比失败
  | "deep_loading"         // 正在生成 Deep 对比
  | "deep_done"            // Deep 对比完成
  | "deep_error"           // Deep 对比失败（保留 Fast）
  | "profile_failed";      // 简历画像失败
```

### 2.2 展示层级 Token

```ts
type CompareDisplayMode = "fast_compact" | "deep_full";
```

页面用这一个 token 控制区块显隐，避免逐字段判断。

### 2.3 页面进入流程

```
页面进入
  → parse URL: fromResumeId, toResumeId, jobId, targetDirection
  → derive comparisonMode
  → loading_latest
  → GET /api/resumes/compare/latest?from_resume_id=X&to_resume_id=Y&comparison_mode=M&...

  latest.has_result=true:
    tier=fast  → fast_done
    tier=quality → deep_done

  latest.has_result=false:
    → fast_loading
    → POST /api/resumes/compare (Fast)
    → fast_done  (resp.profile_status=ready)
    → fast_error (resp.profile_status=failed)
```

### 2.4 activeResult 回退规则

```ts
const activeResult = phase === "deep_done" && deepResult 
  ? deepResult 
  : fastResult;
```

Deep 失败/error → 自动回退到 fastResult，不丢失当前页面。

---

## 3. URL 参数设计

### 3.1 支持参数

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `fromResumeId` | number | ✅ | 原版本简历 ID |
| `toResumeId` | number | ✅ | 目标版本简历 ID |
| `targetDirection` | string | ❌ | 目标方向（→ target_direction 模式） |
| `jobId` | number | ❌ | 目标岗位（→ target_job 模式） |

### 3.2 模式推导

```ts
let comparisonMode: string;
if (jobId) comparisonMode = "target_job";
else if (targetDirection) comparisonMode = "target_direction";
else comparisonMode = "general";
```

### 3.3 URL 示例

```
/resumes/compare?fromResumeId=1&toResumeId=2
/resumes/compare?fromResumeId=1&toResumeId=2&targetDirection=backend_engineering
/resumes/compare?fromResumeId=1&toResumeId=2&jobId=10001
```

### 3.4 边界处理

```
fromResumeId 缺失 → PageError "请选择要对比的简历"
toResumeId 缺失 → PageError "请选择要对比的简历"
fromResumeId === toResumeId → PageError "不能对比相同的简历"
jobId 非数字 → PageError "无效的岗位 ID"
不静默 fallback 到第一份简历
```

---

## 4. Fast 对比展示

### 4.1 定位

快速判断哪版更值得继续打磨。规则引擎 + resume profile 驱动。

### 4.2 展示内容

```
┌─ Header ──────────────────────────────┐
│ "简历版本对比 · 快速分析"               │
│ 从: Resume Name A (ID:1)              │
│ 到: Resume Name B (ID:2)              │
│ 画像: algorithm_research → backend    │
│ 生成时间: ...                          │
│ "这是快速对比结果，点击深度对比获取更完整分析"│
├─ 三栏布局 ────────────────────────────┤
│ ┌─ Left ──┐ ┌─ Center ───┐ ┌─ Right ┐│
│ │ 版本 A  │ │ 对比分析    │ │ 版本 B  ││
│ │ 概述    │ │             │ │        ││
│ │ 优势:   │ │ compare_score│ │ 优势:  ││
│ │ 1-2条   │ │ compare_level│ │ 1-2条  ││
│ │ 风险:   │ │ winner_label │ │ 风险:  ││
│ │ ≤1条    │ │             │ │ ≤1条   ││
│ │         │ │ 关键差异:    │ │        ││
│ │         │ │ 2-3条       │ │        ││
│ │         │ │             │ │        ││
│ │         │ │ 投递建议     │ │        ││
│ └─────────┘ └──────────────┘ └────────┘
└──────────────────────────────────────┘
```

### 4.3 Fast 不展示

```
- from/to_resume_summary（详情概述）
- keep_from_items / keep_to_items
- merge_suggestions
- rewrite_examples
- next_version_suggestions
- 完整的关键差异列表（仅 2-3 条）
```

---

## 5. Deep 对比展示

### 5.1 定位

完整解释两版差异、保留项、合并策略和下一版改写方向。

### 5.2 展示内容

```
┌─ Header ───────────────────────────────┐
│ "简历版本对比 · 深度分析" / "AI生成"      │
│ 同 Fast header 信息                     │
├─ 三栏布局 ─────────────────────────────┤
│ ┌─ Left ──┐ ┌─ Center ──────────┐ ┌─ Right ┐│
│ │ 版本 A  │ │ 对比分析           │ │ 版本 B  ││
│ │         │ │                   │ │        ││
│ │ 画像:   │ │ ⭐ 结论           │ │ 画像:  ││
│ │ X       │ │ compare_score     │ │ Y      ││
│ │         │ │ compare_level     │ │        ││
│ │ 概述:   │ │ winner_label      │ │ 概述:  ││
│ │ 几行    │ │ one_sentence      │ │ 几行   ││
│ │         │ │                   │ │        ││
│ │ 优势:   │ │ 📊 关键差异       │ │ 优势:  ││
│ │ 完整    │ │ 完整列表          │ │ 完整   ││
│ │         │ │                   │ │        ││
│ │ 风险:   │ │ 🔄 合并建议       │ │ 风险:  ││
│ │ 全部    │ │ keep_from_items   │ │ 全部   ││
│ │         │ │ keep_to_items     │ │        ││
│ │ 建议    │ │ merge_suggestions │ │ 建议   ││
│ │ 保留:   │ │                   │ │ 保留:  ││
│ │ 列表    │ │ ✏️ 改写示例       │ │ 列表   ││
│ │         │ │ rewrite_examples  │ │        ││
│ │         │ │ (before/after)    │ │        ││
│ │         │ │                   │ │        ││
│ │         │ │ 📝 下一版建议     │ │        ││
│ │         │ │ next_version_     │ │        ││
│ │         │ │   suggestions     │ │        ││
│ │         │ │                   │ │        ││
│ │         │ │ 🚀 投递建议       │ │        ││
│ │         │ │ delivery_advice   │ │        ││
│ └─────────┘ └───────────────────┘ └────────┘
└──────────────────────────────────────┘
```

### 5.3 Deep 重点展示

```
merge_suggestions   ← 如何合并两版优点
rewrite_examples    ← before/after + section + reason（最核心差异化）
next_version_suggestions ← 下一版应该怎么改
```

### 5.4 Header 状态栏

```tsx
function getHeaderLabel(): string {
  switch (phase) {
    case "deep_loading": return "智能体正在深度分析两版简历差异";
    case "deep_done": 
      return activeGeneratedBy?.startsWith("llm") 
        ? "深度对比 · AI生成" 
        : "深度对比 · 规则兜底";
    case "deep_error": return "深度对比生成失败 · 已保留快速结果";
    default: return "快速对比 · 规则对比";
  }
}
```

---

## 6. AgentStepLoadingOverlay

### 6.1 Deep 生成时复用

```tsx
const compareSteps: AgentLoadingStep[] = [
  { key: "resume-profiles", label: "读取两版简历画像" },
  { key: "analyze-diff", label: "分析关键差异" },
  { key: "judge-winner", label: "判断更优版本" },
  { key: "merge-suggest", label: "生成合并与改写建议" },
];

<AgentStepLoadingOverlay
  open={phase === "deep_loading"}
  title="智能体正在深度分析两版简历差异"
  steps={compareSteps}
  status={phase === "deep_done" ? "success" : "loading"}
/>
```

### 6.2 要求

```
1. Deep loading 时保留当前 Fast 页面作为背景（已有 fast_done 状态）
2. 成功后关闭 overlay → 切换到 deep_done
3. 失败后关闭 overlay → 切换到 deep_error，保留 fast_done 结果
4. 不再使用普通 spinner
```

---

## 7. 入口设计

### 7.1 简历管理页

```
保留现有 "查看版本对比" 按钮（GitCompare 图标）
改为: 需要用户先选择两份简历。

实现方案:
  - 新增 "对比简历" 按钮，disabled 直到选中 2 份简历
  - 或: 当前选中的简历作为 from，弹出选择器选 to
  - URL: /resumes/compare?fromResumeId=X&toResumeId=Y
```

### 7.2 简历详情页（新增入口）

```
可在 ResumeDetailPage 右侧操作区新增 "与另一份简历对比" 入口。
URL: /resumes/compare?fromResumeId={当前详情页ID}&toResumeId={用户选择}
```

### 7.3 简历优化页（可选，本阶段不强制）

```
优化完成后，可对比优化前后的简历版本。
此入口留待 K9-4-E 实现时评估。
```

---

## 8. 展示模式差异总表

| 字段 / 区块 | Fast (compact) | Deep (full) |
|---|---|---|
| Header（两版简历名+ID+画像） | ✅ | ✅ |
| compare_score + compare_level | ✅ | ✅ |
| winner_resume_id + winner_label | ✅ | ✅ |
| one_sentence_summary | ✅ | ✅ |
| key_differences | 2-3 条 | 全部 |
| from_strengths | 1-2 条 | 全部 |
| to_strengths | 1-2 条 | 全部 |
| from_risks | ≤1 条 | 全部 |
| to_risks | ≤1 条 | 全部 |
| delivery_advice | 简短 | 完整 |
| from_resume_summary | — | ✅ |
| to_resume_summary | — | ✅ |
| keep_from_items | — | ✅ |
| keep_to_items | — | ✅ |
| merge_suggestions | — | ✅ **核心** |
| rewrite_examples (before/after) | — | ✅ **最核心** |
| next_version_suggestions | — | ✅ |
| Deep 生成入口按钮 | ✅ | —（替换为"重新生成"） |
| AgentStepLoadingOverlay | 仅在 deep_loading 时显示 | |

---

## 9. 页面状态 Loading/Error 处理

| Phase | 展示 |
|---|---|
| `loading_latest` | PageLoading "正在读取对比结果…" |
| `fast_loading` | Loader2 居中 + "正在生成快速对比" + 说明文案 |
| `fast_error` / `profile_failed` | AlertCircle + errorMessage + "重新尝试"按钮 |
| `deep_loading` | AgentStepLoadingOverlay（保留背景 fast_done 页面） |
| `deep_error` | 保留 fast_done 结果 + 小 banner "深度对比生成失败" |
| 无对比结果 | PageEmpty + 提示 |

---

## 10. 复用组件

| 组件 | 用途 |
|---|---|
| `AgentStepLoadingOverlay` | Deep 加载进度 |
| `PageLoading` | 初始加载 |
| `PageError` / `PageEmpty` | 异常状态 |
| `ContentPageShell` | 统一布局壳 |
| `Badge` | Profile 标签展示 |
| `Loader2` / `AlertCircle` / `RefreshCw` | 图标 |
