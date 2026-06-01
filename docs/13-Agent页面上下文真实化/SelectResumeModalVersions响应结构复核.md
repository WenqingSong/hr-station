# SelectResumeModal Versions 响应结构复核

> 阶段：10-1-R | 状态：pass | 日期：2026-06-01

---

## 一、本阶段目标

修正 SelectResumeModal 对 `/api/resumes/versions` 响应结构的处理，确保主路径来自真实 `response.versions[]`。

---

## 二、10-1 partial_pass 原因

`getResumes()` 声明返回 `Resume[]`，但后端实际返回 `{versions: [{id, version_no, name, ...}], total}`。类型不匹配导致 API 成功但数据不可用。

---

## 三、后端 versions 响应结构

```
GET /api/resumes/versions → {code:0, message:"ok", data: {
  versions: [{id, version_no, name, target_city, target_role, created_at}],
  total: N
}}
```

---

## 四、修正摘要

| 文件 | 修正 |
|------|------|
| `types/resume.ts` | 新增 `ResumeVersionListItem` + `ResumeVersionsResponse` |
| `api/resumes.ts` | 新增 `getResumeVersions()`（正确解析 `response.versions`） |
| `adapters/resumeAdapter.ts` | `toResumeOptionViewModelsFromVersions()` ← `ResumeVersionListItem[]` |
| `api/viewModels.ts` | `loadResumeOptionViewModelsFromApi()` 改用 `getResumeVersions()` |

### 修正前后对比

| 项目 | 修正前 (10-1) | 修正后 (10-1-R) |
|------|------|------|
| API 返回类型 | `Resume[]`（错） | `ResumeVersionsResponse`（对） |
| adapter 输入 | `Resume[]`（错） | `ResumeVersionListItem[]`（对） |
| option.id | `String(versionNo)`（猜测） | `String(v.version_no)`（真实） |
| option.name | `r.name` | `v.name`（真实） |
| option.tags | `r.skills` | `[v.target_city]`（真实可用字段） |

---

## 五、curl 验证

| 检查项 | 结果 |
|------|:--:|
| versions: 2, total: 2 | ✅ |
| v2: V2 Updated (Go Engineer) | ✅ |
| v1: V1 (Dev) | ✅ |

---

## 六、grep

| 检查项 | 结果 |
|------|:--:|
| ResumeVersionListItem / ResumeVersionsResponse | ✅ 存在 |
| getResumeVersions() | ✅ |
| mockResumeOptions in SelectResumeModal | 0 ✅ |

---

## 七、build

```
✓ built in 1.69s ✅
```

---

## 八、P0/P1/P2

- P0: 0
- P1: 0（versions 结构已正确）
- P2: 0

10-1 由 partial_pass → **修正为 pass**。

---

## 九、越界

```
server/ ❌ Agent pages ❌ agents.ts ❌ auth ❌ ✅
```

---

## 十、最终结论

**pass ✅** — 可进入 10-2。
