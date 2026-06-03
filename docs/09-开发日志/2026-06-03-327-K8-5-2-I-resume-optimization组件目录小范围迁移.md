# K8-5-2-I：resume-optimization 组件目录小范围迁移

**日期：** 2026-06-03
**性质：** frontend directory restructure
**范围：** 4 文件迁移 + 1 页面 import 修复
**上游依赖：** K8-5-2-A ~ K8-5-2-H 全部 pass

---

## 1. 本阶段目标

只迁移简历优化 resume-optimization 相关组件到 `web/src/app/components/resume-optimization/`。
不改组件逻辑、不改页面布局、不改业务逻辑、只修 import。

---

## 2. 迁移前 resume-optimization 组件清单

| 组件 | 旧路径 |
|------|--------|
| ResumeOptimizationHeader | `web/src/app/components/ResumeOptimizationHeader.tsx` |
| ResumeOptimizationDetail | `web/src/app/components/ResumeOptimizationDetail.tsx` |
| ResumeOptimizationSidebarLeft | `web/src/app/components/ResumeOptimizationSidebarLeft.tsx` |
| ResumeOptimizationSidebarRight | `web/src/app/components/ResumeOptimizationSidebarRight.tsx` |

四个组件均平铺在 `components/` 根目录下。

---

## 3. 迁移后目录结构

```
web/src/app/components/resume-optimization/
  ResumeOptimizationHeader.tsx
  ResumeOptimizationDetail.tsx
  ResumeOptimizationSidebarLeft.tsx
  ResumeOptimizationSidebarRight.tsx
```

无 `index.ts` barrel 导出——与 jobs/ai/match/recommendation 子目录一致。

---

## 4. 未迁移 SelectResumeModal 说明

`SelectResumeModal.tsx` 仍保留在 `web/src/app/components/` 根目录。
原因：该组件是跨页面复用的简历选择弹窗，被多个页面（JobRecommendationPage、ResumeOptimizationPage 等）引用，不应归入单个功能子目录。

---

## 5. import 修复说明

### 新文件内部 import（原平铺→加深一层）

| 文件 | 修改项 |
|------|--------|
| `ResumeOptimizationHeader.tsx` | `../../adapters`→`../../../adapters`；`../../api/jobs`→`../../../api/jobs`；`../../types`→`../../../types` |
| `ResumeOptimizationDetail.tsx` | `./ui/Button`→`../ui/Button`；`./ui/Badge`→`../ui/Badge`；`../../adapters`→`../../../adapters` |
| `ResumeOptimizationSidebarLeft.tsx` | `./ui/Badge`→`../ui/Badge`；`../../adapters`→`../../../adapters`；`../../api/jobs`→`../../../api/jobs` |
| `ResumeOptimizationSidebarRight.tsx` | `./ui/Badge`→`../ui/Badge`；`../../adapters`→`../../../adapters`；`../../types`→`../../../types` |

### 页面 import

| 页面 | 修改项 |
|------|--------|
| `pages/ResumeOptimizationPage.tsx` | 4 处 import：`../components/ResumeOptimization*` → `../components/resume-optimization/ResumeOptimization*` |

---

## 6. 修改文件清单

| 文件 | 修改内容 | 原因 | 是否属于 K8-5-2-I |
|------|----------|------|-------------------|
| `components/resume-optimization/ResumeOptimizationHeader.tsx` | 新建，3 处 import 适配 | 目录迁移 | 是 |
| `components/resume-optimization/ResumeOptimizationDetail.tsx` | 新建，3 处 import 适配 | 目录迁移 | 是 |
| `components/resume-optimization/ResumeOptimizationSidebarLeft.tsx` | 新建，3 处 import 适配 | 目录迁移 | 是 |
| `components/resume-optimization/ResumeOptimizationSidebarRight.tsx` | 新建，3 处 import 适配 | 目录迁移 | 是 |
| `pages/ResumeOptimizationPage.tsx` | 4 处页面 import 更新 | 页面引用新路径 | 是 |
| `components/ResumeOptimizationHeader.tsx` | 删除 | 已迁移 | 是 |
| `components/ResumeOptimizationDetail.tsx` | 删除 | 已迁移 | 是 |
| `components/ResumeOptimizationSidebarLeft.tsx` | 删除 | 已迁移 | 是 |
| `components/ResumeOptimizationSidebarRight.tsx` | 删除 | 已迁移 | 是 |

注意：`SelectResumeModal.tsx` 未移动（跨页面复用组件），按计划保留。

---

## 7. 引用检查结果

```
grep -rn "components/ResumeOptimizationHeader\|components/ResumeOptimizationDetail\|components/ResumeOptimizationSidebarLeft\|components/ResumeOptimizationSidebarRight" web/src
```

**结果：NO_MATCHES_FOUND** — 旧路径引用无残留 ✅

新路径全部正确：

```
pages/ResumeOptimizationPage.tsx:3:import ... from "../components/resume-optimization/ResumeOptimizationHeader";
pages/ResumeOptimizationPage.tsx:4:import ... from "../components/resume-optimization/ResumeOptimizationSidebarLeft";
pages/ResumeOptimizationPage.tsx:5:import ... from "../components/resume-optimization/ResumeOptimizationDetail";
pages/ResumeOptimizationPage.tsx:6:import ... from "../components/resume-optimization/ResumeOptimizationSidebarRight";
```

---

## 8. build 结果

| 阶段 | 命令 | 结果 | 耗时 |
|------|------|------|------|
| 第一次构建（删旧文件前） | `cd web && npm run build` | ✅ pass | 1.50s |
| 第二次构建（删旧文件后） | `cd web && npm run build` | ✅ pass | 1.53s |

---

## 9. API 验收

| 接口 | 结果 |
|------|------|
| `curl "http://localhost:8080/api/jobs?page=1&page_size=5"` | ✅ code:0 |
| `curl "http://localhost:8080/api/jobs/10001"` | ✅ code:0 |

---

## 10. 页面验证结果或限制

本次为纯目录重组，两次 build 通过、旧路径引用无残留、API 正常、JS/CSS 体积与迁移前一致。未在当前环境打开浏览器实机验证，推荐在条件允许时打开简历优化页确认 Header、左侧 Sidebar、中间优化详情、右侧 Sidebar、Core/标准优化模式切换均正常。

---

## 11. K8-5-2 系列收尾总结

本次是 K8-5-2 组件目录迁移系列的最后一个子任务。

完成状态汇总：

| 子任务 | 迁移组件 | 迁移数 | 状态 |
|--------|---------|--------|------|
| K8-5-2-E | jobs | 8 | ✅ pass |
| K8-5-2-F | ai | 2 | ✅ pass |
| K8-5-2-G | match | 4 | ✅ pass |
| K8-5-2-H | recommendation | 4 | ✅ pass |
| K8-5-2-I | resume-optimization | 4 | ✅ pass |

**总计：22 个组件迁移到 5 个功能子目录。**

`components/` 根目录剩余文件（均不出现在任何功能子目录）：
- `CollapsibleSection.tsx` — 通用折叠组件
- `Navbar.tsx` — 全局导航栏
- `PageContainer.tsx` — 通用页面容器
- `PageState.tsx` — 通用页面状态组件
- `SelectResumeModal.tsx` — 跨页面复用简历选择弹窗
- `figma/` — Figma 相关
- `ui/` — 通用 UI 组件库

---

## 12. 本次未执行 Git 命令说明

本次未执行任何 Git 命令，也未提交代码。
