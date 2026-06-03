# K8-5-2-H：recommendation 组件目录小范围迁移

**日期：** 2026-06-03
**性质：** frontend directory restructure
**范围：** 4 文件迁移 + 1 页面 import 修复
**上游依赖：** K8-5-2-A ~ K8-5-2-G 全部 pass

---

## 1. 本阶段目标

只迁移岗位智能推荐 recommendation 相关组件到 `web/src/app/components/recommendation/`。
不改组件逻辑、不改页面布局、不改业务逻辑、只修 import。

---

## 2. 迁移前 recommendation 组件清单

| 组件 | 旧路径 |
|------|--------|
| RecommendationHeader | `web/src/app/components/RecommendationHeader.tsx` |
| RecommendationSidebar | `web/src/app/components/RecommendationSidebar.tsx` |
| RecommendationDetail | `web/src/app/components/RecommendationDetail.tsx` |
| RecommendationContext | `web/src/app/components/RecommendationContext.tsx` |

四个组件均平铺在 `components/` 根目录下。

---

## 3. 迁移后目录结构

```
web/src/app/components/recommendation/
  RecommendationHeader.tsx
  RecommendationSidebar.tsx
  RecommendationDetail.tsx
  RecommendationContext.tsx
```

无 `index.ts` barrel 导出——与 jobs/ai/match 子目录一致。

---

## 4. 未迁移其他组件说明

| 组件 | 理由 |
|------|------|
| `ResumeOptimization*` | 属于 K8-5-2-I resumes 迁移 |
| `jobs/*` | 已在 K8-5-2-E 完成 |
| `ai/*` | 已在 K8-5-2-F 完成 |
| `match/*` | 已在 K8-5-2-G 完成 |

---

## 5. import 修复说明

### 新文件内部 import（原平铺→加深一层）

| 文件 | 修改项 |
|------|--------|
| `RecommendationHeader.tsx` | 无内部相对路径 import，无需修复 |
| `RecommendationSidebar.tsx` | `./ui/Badge` → `../ui/Badge`；`../utils` → `../../utils`；`../../adapters` → `../../../adapters` |
| `RecommendationDetail.tsx` | `./ui/Button` → `../ui/Button`；`./ui/Badge` → `../ui/Badge`；`../../adapters` → `../../../adapters` |
| `RecommendationContext.tsx` | `./ui/Badge` → `../ui/Badge`；`../../adapters` → `../../../adapters` |

### 页面 import

| 页面 | 修改项 |
|------|--------|
| `pages/JobRecommendationPage.tsx` | 4 处 import：`../components/Recommendation*` → `../components/recommendation/Recommendation*` |

---

## 6. 修改文件清单

| 文件 | 修改内容 | 原因 | 是否属于 K8-5-2-H |
|------|----------|------|-------------------|
| `components/recommendation/RecommendationHeader.tsx` | 新建（无 import 变更） | 迁移到子目录 | 是 |
| `components/recommendation/RecommendationSidebar.tsx` | 新建，3 处 import 适配 | 目录迁移 | 是 |
| `components/recommendation/RecommendationDetail.tsx` | 新建，3 处 import 适配 | 目录迁移 | 是 |
| `components/recommendation/RecommendationContext.tsx` | 新建，2 处 import 适配 | 目录迁移 | 是 |
| `pages/JobRecommendationPage.tsx` | 4 处页面 import 更新 | 页面引用新路径 | 是 |
| `components/RecommendationHeader.tsx` | 删除 | 已迁移 | 是 |
| `components/RecommendationSidebar.tsx` | 删除 | 已迁移 | 是 |
| `components/RecommendationDetail.tsx` | 删除 | 已迁移 | 是 |
| `components/RecommendationContext.tsx` | 删除 | 已迁移 | 是 |

---

## 7. 引用检查结果

```
grep -rn "components/RecommendationHeader\|components/RecommendationSidebar\|components/RecommendationDetail\|components/RecommendationContext" web/src
```

**结果：NO_MATCHES_FOUND** — 旧路径引用无残留 ✅

新路径全部正确：

```
pages/JobRecommendationPage.tsx:3:import ... from "../components/recommendation/RecommendationHeader";
pages/JobRecommendationPage.tsx:4:import ... from "../components/recommendation/RecommendationSidebar";
pages/JobRecommendationPage.tsx:5:import ... from "../components/recommendation/RecommendationDetail";
pages/JobRecommendationPage.tsx:6:import ... from "../components/recommendation/RecommendationContext";
```

---

## 8. build 结果

| 阶段 | 命令 | 结果 | 耗时 |
|------|------|------|------|
| 第一次构建（删旧文件前） | `cd web && npm run build` | ✅ pass | 1.68s |
| 第二次构建（删旧文件后） | `cd web && npm run build` | ✅ pass | 1.58s |

---

## 9. API 验收

| 接口 | 结果 |
|------|------|
| `curl "http://localhost:8080/api/jobs?page=1&page_size=5"` | ✅ code:0 |
| `curl "http://localhost:8080/api/jobs/10001"` | ✅ code:0 |

---

## 10. 页面验证结果或限制

本次为纯目录重组（不改变组件逻辑/布局/样式），且：
- 两次 `npm run build` 均通过（1669 modules）
- 旧路径引用无残留
- API 接口正常
- JS/CSS 体积与迁移前完全一致

页面实机验证未执行，但基于 build 通过 + import 正确 + API 正常可判定迁移正确。

---

## 11. 本次未执行 Git 命令说明

本次未执行任何 Git 命令，也未提交代码。
