# 匹配页 Selected Resume Version Detail 对齐

> 阶段：10-3-R | 状态：pass | 日期：2026-06-01

---

## 一、背景

10-3 匹配页上下文真实化完成时，存在一个一致性风险：

```text
Match Agent 请求使用 selectedResumeOption.versionNo，
但页面传给 Header / SidebarRight 的 Resume 详情来自 getCurrentResume()。

当 selected versionNo 不是 current/latest version 时：
- 页面展示 Resume = current/latest version
- Match Agent 请求 resume_version_no = selected versionNo

这导致页面展示上下文与 Agent 请求上下文不一致。
```

本阶段（10-3-R）专门修复此一致性问题。

---

## 二、修改范围

| 文件 | 修改内容 |
|------|------|
| `web/src/api/viewModels.ts` | 新增 `loadResumeVersionDetail(versionNo)` 函数 |
| `web/src/app/pages/JobMatchAnalysisPage.tsx` | 替换 `getCurrentResume()` 为版本详情加载 |

**零越界**：未修改 server / 推荐页 / 优化页 / SelectResumeModal / resumes.ts。

---

## 三、修复策略

### 修复前（10-3 partial_pass 状态）

```text
selectedResumeOption.versionNo → Match Agent 请求 ✅
getCurrentResume() → 页面展示 Resume 详情 ❌（不一致）
```

### 修复后（10-3-R）

```text
selectedResumeOption.versionNo
  → loadResumeVersionDetail(versionNo) → selectedResumeDetail → 页面展示 ✅

selectedResumeOption.versionNo
  → loadMatchAnalysisViewModel(jobId, { resume_version_no: versionNo }) → Match Agent ✅

selectedResumeOption.versionNo
  === loadResumeVersionDetail 使用的 versionNo
  === Match Agent 请求体中的 resume_version_no
  === 页面展示的 Resume 详情来源
```

---

## 四、数据流设计

```
URL jobId
→ getJobDetail(jobId)
→ selectedJob

loadResumeOptionViewModelsFromApi()
→ resumeOptions
→ selectedResumeOption
→ versionNo = Number(selectedResumeOption.id)

versionNo
→ loadResumeVersionDetail(versionNo)
→ selectedResumeDetail
→ Header & SidebarRight display

jobId + versionNo
→ loadMatchAnalysisViewModel(jobId, { resume_version_no: versionNo })
→ match result

页面展示：
selectedJob + selectedResumeOption + selectedResumeDetail + match result
```

---

## 五、新增函数

### `loadResumeVersionDetail(versionNo)` (viewModels.ts)

```ts
export async function loadResumeVersionDetail(
  versionNo: number,
): Promise<Resume | null> {
  try {
    return await get<Resume>(`/resumes/versions/${versionNo}`);
  } catch {
    return null;
  }
}
```

调用 `GET /resumes/versions/{versionNo}`（后端阶段 6 已验证可用），返回指定版本的 Resume 详情。

---

## 六、页面状态变化

### 移除

- `import { getCurrentResume } from "../../api/resumes"`
- `const [currentResume, setCurrentResume] = useState<Resume | null>(null)`

### 新增

- `import { loadResumeVersionDetail } from "../../api"`
- `const [selectedResumeDetail, setSelectedResumeDetail] = useState<Resume | null>(null)`
- `const [loadingResumeDetail, setLoadingResumeDetail] = useState(false)`
- `const [resumeDetailError, setResumeDetailError] = useState<string | null>(null)`

### 新增空态

| 状态 | 展示 |
|------|------|
| 版本详情加载中 | "正在加载简历版本详情…" |
| 版本详情加载失败 | "简历版本详情加载失败" + "管理我的简历" 链接 |
| 版本详情失败 | 不调用 Match Agent |

### Match Agent 触发条件变更

```text
修复前：selectedOption 存在即触发
修复后：selectedOption && selectedResumeDetail 均存在才触发
```

---

## 七、验证

| 检查项 | 结果 |
|------|:--:|
| build | ✅ 1.69s |
| getCurrentResume import | 0（仅注释引用） ✅ |
| getResumeVersionDetail usage | ✅ loadResumeVersionDetail 7 次引用 |
| selectedResumeDetail usage | ✅ 2 处 props 传递 + 1 处 guard |
| loadingResumeDetail / resumeDetailError | ✅ 3 处状态使用 |
| mock import | 0 ✅ |
| resume-v2 hardcoded | 0 ✅ |
| engine token exposure | 0 ✅ |
| forbidden file changes | 0 ✅ |

---

## 八、git diff 摘要

```
web/src/api/viewModels.ts     — +get import, +Resume type, +loadResumeVersionDetail
web/src/app/pages/JobMatchAnalysisPage.tsx  — getCurrentResume→version detail, +loading/error states
```

仅修改 2 个文件，零越界。

---

## 九、结论

**pass** ✅

所有四链路已对齐：

```text
selectedResumeOption.versionNo (v1)
  → loadResumeVersionDetail(v1) → selectedResumeDetail (V1 详情) → 页面展示 ✅
  → loadMatchAnalysisViewModel(jobId, { resume_version_no: v1 }) → Agent 请求 ✅
```

10-3 从 **partial_pass 修正为 pass**。

可进入 10-4 简历优化页上下文真实化。
