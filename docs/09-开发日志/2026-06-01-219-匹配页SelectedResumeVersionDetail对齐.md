# 匹配页 Selected Resume Version Detail 对齐

> 阶段：10-3-R | 日期：2026-06-01 | 状态：pass

---

## 一、背景

10-3 匹配页上下文真实化 partial_pass 遗留问题：页面展示的 Resume 详情来自 `getCurrentResume()`（latest），而 Match Agent 请求使用 `selectedResumeOption.versionNo`（可能是历史版本），存在一致性风险。

---

## 二、修改内容

### 2.1 viewModels.ts

- 新增 `get` import from `./index`
- 新增 `Resume` type import
- 新增 `loadResumeVersionDetail(versionNo: number): Promise<Resume | null>` 函数

### 2.2 JobMatchAnalysisPage.tsx

| 改动 | 说明 |
|------|------|
| 移除 `getCurrentResume` import | 不再使用 current/latest resume |
| 新增 `loadResumeVersionDetail` import | 获取指定版本详情 |
| 移除 `useCallback` import | 不再需要 |
| 状态 `currentResume` → `selectedResumeDetail` | 对齐命名 |
| 状态 `loadingResume` → `loadingOptions` | 拆分 loading 语义 |
| 新增 `loadingResumeDetail` / `resumeDetailError` | 版本详情产品化状态 |
| useEffect 2：拆分为 options 加载 | 仅加载选项列表，选择 match option |
| useEffect 3：新增 version detail 加载 | 根据 selectedOption.id 获取版本详情 |
| useEffect 4：Match Agent guard 增强 | `!selectedResumeDetail` 不调用 Agent |
| 新增空态：loadingResumeDetail | "正在加载简历版本详情…" |
| 新增空态：resumeDetailError | 错误文案 + "管理我的简历" 链接 |
| Header / SidebarRight props | `currentResume` → `selectedResumeDetail` |

---

## 三、验证

- build: 1.69s ✅
- grep: getCurrentResume 仅注释 ✅
- grep: selectedResumeDetail / loadingResumeDetail / resumeDetailError 完整覆盖 ✅
- grep: 0 mock / 0 hardcoded / 0 token in page ✅
- 零越界 ✅

---

## 四、结论

10-3 partial_pass → **pass**。

可进入 10-4 简历优化页上下文真实化。
