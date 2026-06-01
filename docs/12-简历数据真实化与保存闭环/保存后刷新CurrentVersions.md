# 保存后刷新 Current / Versions

> 阶段：9-4 | 状态：pass | 日期：2026-06-01

---

## 一、本阶段目标

补齐 9-3 之后保存成功的刷新闭环，确保 current 和 versions 在保存后能反映最新状态。

---

## 二、9-3 保存后遗留问题

| 问题 | 9-3 状态 |
|------|:--:|
| 保存成功但没有刷新 current | ❌ |
| 保存成功但没有刷新 versions | ❌ |
| 用户保存后误以为数据已同步 | ❌ |

---

## 三、修改范围

| 文件 | 操作 |
|------|:--:|
| `web/src/app/pages/MyResumesPage.tsx` | 修改 |

---

## 四、保存后刷新策略（策略 A）

保存成功后立即执行：
```ts
await Promise.allSettled([getCurrentResume(), getResumes()]);
```

- `getCurrentResume()` → 确认后端 current 已更新为最新版本
- `getResumes()` (即 versions) → 确认版本列表已包含新版本
- 使用 `allSettled` 而非 `all`，避免刷新失败阻断成功提示

---

## 五、add 模式保存后刷新

- `createResume(payload)` 成功
- `Promise.allSettled([getCurrentResume(), getResumes()])` 刷新
- 成功提示：`"简历已创建成功，当前简历已更新。"`

## 六、edit 模式保存后刷新

- `updateResume("", payload)` 成功
- `Promise.allSettled([getCurrentResume(), getResumes()])` 刷新
- 成功提示：`"修改已保存，已生成新的简历版本。当前简历已更新。"`

---

## 七、current 刷新验证

| 方式 | 实现 |
|------|------|
| 保存后立即刷新 | `getCurrentResume()` in handleSave |
| /resumes 重新进入 | `ResumeManagementPage` 每次 mount 调用 `loadResumeManagementPageViewModel()` |
| /resumes/edit 重新进入 | 每次 mount 调用 `loadResumeManagementPageViewModel()` |

## 八、versions 刷新验证

| 方式 | 实现 |
|------|------|
| 保存后立即刷新 | `getResumes()` (即 versions) in handleSave |
| /resumes 重新进入 | 不直接展示 versions，但 API 已刷新 |

---

## 九、setCurrentResume

**未处理。** 后端 current=latest version，保存成功后新版本自然成为 current。

---

## 十、保存失败行为

| 场景 | 行为 |
|------|------|
| API 失败 | `setSaveErrorMsg("保存失败...")`，**不刷新** current/versions |
| 401 | `setSaveErrorMsg("登录状态已失效...")`，不刷新 |
| 网络错误 | `setSaveErrorMsg("保存失败...")` |

---

## 十一、curl 验证结果

| 场景 | 结果 |
|------|:--:|
| create V1 | version_no=1 ✅ |
| current after V1 | version_no=1, name=V1 ✅ |
| versions after V1 | total=1 ✅ |
| create V2 | version_no=2 ✅ |
| current after V2 | version_no=2, name=V2 Updated ✅ |
| versions after V2 | total=2 ✅ |

---

## 十二、build 结果

```
✓ built in 1.64s ✅
```

---

## 十三、grep 验证

| 检查项 | 结果 |
|------|:--:|
| getCurrentResume/getResumes 调用 | ✅ lines 15, 291 |
| createResume/updateResume 保留 | ✅ lines 285-287 |
| setCurrentResume 调用 | 0 ✅ |
| "已保存本次演示修改" | 0 ✅ |

---

## 十四、P0/P1/P2

| 等级 | 数量 | 说明 |
|:--:|:--:|------|
| P0 | 0 | — |
| P1 | 0 | current/versions 刷新已实现 |
| P2 | 0 | — |

---

## 十五、越界检查

```
server/ ❌
仅修改 MyResumesPage.tsx ✅
```

---

## 十六、最终结论

**pass ✅** — 可进入 9-5 简历链路回归验收。
