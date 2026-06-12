# Recommendation-Entry-And-Deep-State-Fix-L2：修复推荐页入口与 Deep 优先状态机

## 修复目标

基于 L1 审计，修复 `/jobs` 入口不带 resumeId、空态过早闪现、activeResult 不优先 Deep 三个问题。

---

## 修复详情

### 1. /jobs 入口显式带 resumeId

**文件**：`web/src/app/components/ai/AIDecisionWorkspace.tsx:115`

```diff
- onClick={() => navigate("/recommend")}
+ onClick={() => navigate(`/recommend?resumeId=${resumeVm.id}`)}
```

### 2. /recommend 无 resumeId 时安全重定向

**文件**：`web/src/app/pages/JobRecommendationPage.tsx:137-143`

URL 无 resumeId 且简历列表已加载时，自动 `navigate(`/recommend?resumeId=${cur.id}`, { replace: true })`，避免依赖隐式 is_current 回退。

### 3. latest loading guard（isLatestLoading）

新增 `isLatestLoading` 状态：
- quality-first init 启动时设为 `true`
- quality 完整（≥5）→ `false`
- quality 不完整 → fast found → `false`
- quality 不完整 → generateFast 完成 → `false`
- generateFast 成功/失败 → `false`

空态判断前检查 `isLatestLoading`：若为 `true` 则显示"正在恢复推荐结果…"而非"暂无推荐结果"。

### 4. activeResult Deep 优先

```ts
const deepJobs = deepResult?.recommendedJobs ?? [];
const fastJobs = fastResult?.recommendedJobs ?? [];
const deepReady = deepJobs.length >= 5;
const fastReady = fastJobs.length > 0;
const activeResult = deepReady ? deepResult
  : fastReady ? fastResult
  : deepResult ?? fastResult;
```

不依赖 phase → Deep 完整存在时，永远优先 Deep。

### 5. isDeepDisplay 基于实际结果

```diff
- const isDeepDisplay = phase === "deep_done";
+ const isDeepDisplay = activeResult === deepResult && deepReady;
```

右侧控制台模式标识跟随实际展示结果。

### 6. 空态 guard

空态仅在 `isLatestLoading === false && (!activeResult || jobs.length === 0)` 时出现。

---

## 修改文件

| 文件 | 修改内容 |
|------|------|
| `web/src/app/components/ai/AIDecisionWorkspace.tsx` | navigate 带 resumeId |
| `web/src/app/pages/JobRecommendationPage.tsx` | 无 resumeId 重定向 + isLatestLoading + activeResult Deep-first + 空态 guard |

---

## API 验证

| resumeId | latest quality items | generated_by | 结论 |
|:---:|:---:|------|:---:|
| 1 | 5 | llm_agent | ✅ |
| 8 | 5 | llm_agent | ✅ |

---

## 构建

| 项目 | 结果 |
|------|:---:|
| npm build | ✅ PASS (2.25s) |

---

## 浏览器验证

| 场景 | 预期 |
|------|------|
| `/jobs` → 推荐适合的岗位 | URL → `/recommend?resumeId=1`，展示 Deep Top5 |
| `/recommend` | → 自动 replace 到 `/recommend?resumeId=1` |
| `/recommend?resumeId=8` | 展示 Deep Top5 |
| 0 条 Deep + 0 条 Fast | 显示"正在恢复推荐结果…"→"暂无推荐结果"|
