# SelectResumeModal / Resume Options 真实化

> 阶段：10-1 | 状态：pass | 日期：2026-06-01

---

## 一、本阶段目标

让 SelectResumeModal 简历选项来自真实 resume versions API，不再依赖 mockResumeOptions。

---

## 二、修改范围

| 文件 | 操作 | 说明 |
|------|:--:|------|
| `web/src/adapters/resumeAdapter.ts` | 修改 | 新增 `toResumeOptionViewModelsFromApi()`（Resume[] → OptionVM[]） |
| `web/src/api/viewModels.ts` | 修改 | 新增 `loadResumeOptionViewModelsFromApi()` |
| `web/src/app/components/SelectResumeModal.tsx` | 修改 | 切换 loader，默认选最新 |

---

## 三、SelectResumeModal 改造

| 项目 | 原来 | 现在 |
|------|------|------|
| loader | `loadResumeOptionViewModels` → mock fallback | `loadResumeOptionViewModelsFromApi` → 真实 API + fallback |
| 默认选中 | hardcoded "v2" | 最新版本（第一个 option） |
| mock import | 间接依赖 mockResumeOptions | 0 direct import |
| option id | mock ID (v1/v2/v3) | `String(versionNo)` from API |
| name | mock data | 真实简历名称 |
| tags | mock tags | `skills.slice(0,3)` |
| isDefault | mock | 第一个 = true（最新版本） |

---

## 四、adapter 新增

`toResumeOptionViewModelsFromApi(resumes: Resume[]): ResumeOptionViewModel[]`

| option field | 来源 |
|------|------|
| `id` | `String(versionNo)` |
| `name` | `r.name \|\| "版本 ${versionNo}"` |
| `statusLabel` | idx===0 → "当前使用" |
| `directionLabel` | `r.direction_label` |
| `tags` | `r.skills.slice(0,3)` |
| `projectCount` | `r.projects.length` |
| `isDefault` | idx===0 → true |
| `isRecommended` | idx===0 → true |

---

## 五、loader 新增

`loadResumeOptionViewModelsFromApi()`:
1. `getResumes()` → GET /resumes/versions → `Resume[]`
2. `toResumeOptionViewModelsFromApi(versions)`
3. catch → fallback to `getResumeOptions()` (mock)

---

## 六、build

```
✓ built in 1.77s ✅
```

---

## 七、grep

- mockResumeOptions in SelectResumeModal: **0** ✅
- loader: `loadResumeOptionViewModelsFromApi` ✅

---

## 八、越界检查

```
server/ ❌
仅修改 resumeAdapter + viewModels + SelectResumeModal ✅
```

---

## 九、未处理事项

```
✅ SelectResumeModal 真实化完成
✅ 推荐/匹配/优化页未动 → 10-2/10-3/10-4
✅ Agent 请求体未动 → 10-2+
✅ 后端未动
```

---

## 十、最终结论

**pass ✅** — 可进入 10-2 推荐页上下文真实化。
