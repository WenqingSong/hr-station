# 保存后刷新 Current / Versions

> 阶段：9-4 | 日期：2026-06-01 | 编号：212

## 执行摘要

- **目标**：保存成功后刷新 current 和 versions
- **结论**：pass ✅

## 修改文件

- `web/src/app/pages/MyResumesPage.tsx` — handleSave 中添加 `Promise.allSettled([getCurrentResume(), getResumes()])`

## 刷新策略

- 策略 A：保存成功后立即拉取 current + versions
- 使用 allSettled 避免刷新失败阻断成功提示

## curl 验证

| 场景 | 结果 |
|------|:--:|
| V1 create → current/versions | ✅ |
| V2 create → current/versions | ✅ |

## build

✓ 1.64s ✅

## 新增文档

- `docs/12-简历数据真实化与保存闭环/保存后刷新CurrentVersions.md`

## 最终结论

pass ✅ — 可进入 9-5
