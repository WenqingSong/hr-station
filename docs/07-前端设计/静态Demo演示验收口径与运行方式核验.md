# 静态 Demo 演示验收口径与运行方式核验

> 阶段：5-7.0 | 状态：已完成 | 日期：2026-05-31

## 1. 本阶段目标

确定后续 5-7 的验收口径：Demo 运行方式、验收路径、P0/P1/P2 判定标准、子阶段范围。

本阶段为 **read-only + docs**，不实际启动服务，不执行路径验收。

## 2. 构建准入前置条件

**来源：5-6.4 前端 Build 准入总验收**

| 条件 | 状态 | 证据 |
|------|:--:|------|
| Node | v22.22.2 | `node -v` |
| npm | 10.9.7 | `npm -v` |
| package.json 存在 | ✅ | `web/package.json` |
| package-lock.json 存在 | ✅ | `npm install` 生成 |
| node_modules 存在 | ✅ | 285 packages |
| vite binary 存在 | ✅ | `node_modules/.bin/vite` |
| npm run build 通过 | ✅ | `✓ built in 1.72s` |
| web/dist 存在 | ✅ | 560K |
| dist/index.html | ✅ | 含 JS/CSS 引用 |
| JS asset | ✅ | `index-DNRPjRjB.js` |
| CSS asset | ✅ | `index-BPO82uEp.css` |
| 后端越界 | ❌ 无 | `git diff` clean |
| axios / api-v1 | ❌ 无 | `grep` 零匹配 |
| selectors | ❌ 无 | `grep` 零匹配 |
| 页面 import API | ❌ 无 | `grep` 零匹配 |

**结论：build 准入前置条件 ✅ 全部满足，可进入 5-7。**

## 3. Demo 运行方式

### 3.1 当前 scripts

```json
{
  "build": "vite build",
  "dev": "vite"
}
```

| 运行方式 | 是否存在 | 命令 | 是否建议用于 5-7 |
|------|:--:|------|:--:|
| dev | ✅ | `npm run dev` | ❌ 不建议（HMR 开发模式，非静态演示） |
| build | ✅ | `npm run build` | ❌ 已执行（构建阶段已过） |
| preview | ❌ 未定义 | — | ⚠️ 需补充 |

### 3.2 建议方案

`package.json` scripts 中未定义 `preview`，但：

- Vite 6.3.5 内置 `vite preview` 命令
- `web/dist` 已存在且完整
- `node_modules/.bin/vite` 存在

**推荐方案 A（优先）：**

```bash
cd web && npx vite preview --outDir dist
```

或先补充 scripts：

```json
"preview": "vite preview --outDir dist"
```

然后：

```bash
cd web && npm run preview
```

**推荐方案 B（备选）：**

```bash
# 直接用任意静态文件服务器
cd web && npx serve dist
```

### 3.3 运行矩阵

| 参数 | 建议值 |
|------|------|
| 入口 | `web/dist/index.html` |
| 端口 | Vite preview 默认 4173 |
| 后端 | **不需要启动** |
| 真实 API | **不需要接入** |
| 认证 | **不需要** |
| LLM | **不调用** |

**Demotion 性质**：静态 Demo，全部使用 mocks/adapters 本地数据，零后端依赖。

### 3.4 判决

**使用 `npx vite preview` 或补充 `preview` script 后通过 `npm run preview` 验收 dist 静态产物。本阶段不启动。**

## 4. 三条主演示路径

### 4.1 路径一：岗位列表 → 选择简历 → 岗位推荐

```text
/
→ 点击"推荐最适合的 5 个岗位"
→ SelectResumeModal
→ 默认选择 V2 后端开发简历
→ 确认
→ /recommend?resumeId=v2
→ 岗位推荐页展示当前简历与推荐结果
```

- **使用页面**：JobListPage → SelectResumeModal → JobRecommendationPage
- **入口组件**：AIDecisionWorkspace（`setIsResumeModalOpen(true)`）
- **URL 参数**：`/recommend?resumeId=v2`
- **验收阶段**：5-7.1
- **路由**：✅ 已确认（`/` → `createBrowserRouter` path: `"/"` component: JobListPage）

### 4.2 路径二：岗位详情 → 匹配分析 → 简历优化

```text
/jobs/:id
→ 点击"开始岗位匹配分析"
→ SelectResumeModal
→ 确认
→ /match?jobId=xxx&resumeId=v2
→ 岗位匹配分析页展示岗位、简历、分析来源
→ 点击"针对该岗位优化简历"
→ /resume-optimization?jobId=xxx&resumeId=v2
→ 简历优化页展示目标岗位、当前简历、优化建议
```

- **使用页面**：JobDetailPage → SelectResumeModal → JobMatchAnalysisPage → ResumeOptimizationPage
- **入口组件**：JobDetailAIDecision（`useParams().id` → `setIsResumeModalOpen(true)`）
- **URL 参数**：`/match?jobId=xxx&resumeId=v2` → `/resume-optimization?jobId=xxx&resumeId=v2`
- **验收阶段**：5-7.2
- **路由**：✅ 已确认（`/jobs/:id` → JobDetailPage, `/match` → JobMatchAnalysisPage, `/resume-optimization` → ResumeOptimizationPage）

### 4.3 路径三：我的简历 → 编辑 / 添加 / 历史比较

```text
Navbar 我的简历
→ /resumes
→ 修改简历 /resumes/edit
→ 添加简历 /resumes/new
→ 历史简历比较 /resumes/compare
```

- **使用页面**：ResumeManagementPage → MyResumesPage → ResumeComparisonPage
- **入口**：Navbar（`to="/resumes"`）
- **URL 参数**：无（使用 `useLocation` 模式区分 edit/new）
- **验收阶段**：5-7.3

### 4.4 冻结确认

| 检查项 | 结果 | 证据 |
|------|:--:|------|
| `createBrowserRouter` 存在 | ✅ | `routes.tsx` |
| `RouterProvider` 存在 | ✅ | `App.tsx` |
| `/` 路由 | ✅ | index → JobListPage |
| `/jobs/:id` 路由 | ✅ | JobDetailPage |
| `/recommend` 路由 | ✅ | JobRecommendationPage |
| `/match` 路由 | ✅ | JobMatchAnalysisPage |
| `/resume-optimization` 路由 | ✅ | ResumeOptimizationPage |
| `/resumes` 路由 | ✅ | ResumeManagementPage |
| `/resumes/edit` 路由 | ✅ | MyResumesPage |
| `/resumes/new` 路由 | ✅ | MyResumesPage（useLocation 区分） |
| `/resumes/compare` 路由 | ✅ | ResumeComparisonPage |
| `SelectResumeModal` 已接入 | ✅ | AIDecisionWorkspace + JobDetailAIDecision |
| `useSearchParams` resumeId | ✅ | JobRecommendationPage |
| `useSearchParams` jobId/resumeId | ✅ | JobMatchAnalysisPage + ResumeOptimizationPage |
| 导航按钮文案存在 | ✅ | "推荐最适合的 5 个岗位" / "开始岗位匹配分析" / "针对该岗位优化简历" |
| Navbar "我的简历" | ✅ | `to="/resumes"` |

**三条主演示路径已冻结 ✅。**

## 5. P0 / P1 / P2 判定标准

### 5.1 P0：演示阻断

满足任一项即 P0：

| # | 条件 | 
|---|------|
| 1 | 页面打不开 |
| 2 | 主路由不可达 |
| 3 | 点击主按钮无反应 |
| 4 | SelectResumeModal 不弹出 |
| 5 | 确认后不跳转 |
| 6 | 跳转后页面白屏 |
| 7 | /recommend /match /resume-optimization 完全不显示上下文 |
| 8 | 简历管理入口不可达 |
| 9 | 页面运行时控制台明显报错 |
| 10 | 构建产物缺失 |

**P0 必须立即处理，不能进入下一条路径验收。**

### 5.2 P1：演示穿帮 / 体验风险

满足任一项可判为 P1：

| # | 条件 |
|---|------|
| 1 | 页面能走通，但文案让评委误解 |
| 2 | 默认简历 / 当前简历文案不统一 |
| 3 | fallback 文案太技术化 |
| 4 | 推荐页 → 匹配页只传 resumeId 不传 jobId |
| 5 | 按钮层级不自然 |
| 6 | 返回路径不够顺 |
| 7 | 某个卡片信息和当前上下文不完全一致 |
| 8 | 演示态保存提示不够清楚 |

**P1 不阻断 5-7，但须记录并进入 5-8 视觉与文案穿帮点清理。**

### 5.3 P2：可延后优化

满足任一项可判为 P2：

| # | 条件 |
|---|------|
| 1 | legacy deprecated 文件仍存在但未接入 |
| 2 | 局部间距不完美但不影响讲解 |
| 3 | 部分文档描述不够细 |
| 4 | dist 是否纳入版本控制尚未最终确认 |
| 5 | 部分开发辅助页面仍存在，如 style-guide |

**P2 不阻断 5-7。**

## 6. 5-7 后续子阶段验收范围

| 子阶段 | 验收路径 | 是否允许改代码 | 输出 |
|------|------|:--:|------|
| 5-7.1 | 岗位列表 → 选择简历 → 岗位推荐 | 原则否 | 路径一验收报告 |
| 5-7.2 | 岗位详情 → 匹配分析 → 简历优化 | 原则否 | 路径二验收报告 |
| 5-7.3 | 我的简历 → 编辑 / 添加 / 历史比较 | 原则否 | 路径三验收报告 |
| 5-7.4 | Demo 演示路径总验收 | 否 | 总验收报告 |

**允许极小修复的条件**：只有 P0 演示阻断，且修复范围很小（如单行 typo、单文件 import 修正），才允许在对应子阶段最小修复。P1 / P2 默认只记录，进入 5-8 统一清理。

## 7. 当前已知限制

来源：5-6.4 + 5-5.2

| 编号 | 限制 | 等级 | 影响 |
|:--:|------|:--:|------|
| R-1 | `npm run preview` 未定义 | P2 | 需补充 script 或使用 `npx vite preview` |
| R-2 | 推荐→匹配仅传 resumeId | P1 | 匹配页 jobId fallback 默认岗位 |
| R-3 | 页面未接真实 API | P1 | 全部使用 mocks/adapters |
| R-4 | Agent 输出仍用 mock | P1 | 形态可演示，非真实推理 |
| R-5 | 简历保存前端演示态 | P1 | 无持久化 |
| R-6 | 4 个 legacy @deprecated 组件 | P2 | 未接入主流程 |
| R-7 | StyleGuide / Test 辅助页面存在 | P2 | 不影响演示 |
| R-8 | dist 不在版本控制 | P2 | 需本地 build |

**以上限制均不阻断 5-7 演示路径验收。**

## 8. 越界检查

| 检查 | 结果 |
|------|:--:|
| `git diff` web/src/ | ❌ clean |
| `git diff` server/ | ❌ clean |
| `git diff` Makefile/docker-compose/nginx/ | ❌ clean |
| axios | ❌ 无 |
| /api/v1 | ❌ 无 |
| selectors | ❌ 无 |
| 页面 import API | ❌ 无 |
| 本阶段修改 web/src | ❌ 无 |
| 本阶段修改 server | ❌ 无 |
| npm install | ❌ 未执行 |
| npm run build | ❌ 未执行 |
| npm run preview | ❌ 未启动 |
| npm run dev | ❌ 未启动 |
| 真实 LLM 调用 | ❌ 无 |
| 清空 volume | ❌ 无 |

## 9. 本阶段未做事项

1. 未启动 `npm run preview` / `npm run dev`
2. 未执行实际路径验收
3. 未执行 5-7.1 / 5-7.2 / 5-7.3
4. 未做视觉清理
5. 未接 API
6. 未新增业务功能
7. 未修改 web/src
8. 未修改 server
9. 未重新 npm install / build

## 10. 是否可以进入 5-7.1

**可以。** 验收口径已形成：

- ✅ 构建准入前置条件全部满足
- ✅ Demo 运行方式已确定（`npx vite preview` 或补充 preview script）
- ✅ 三条主演示路径已冻结
- ✅ P0/P1/P2 判定标准已形成
- ✅ 5-7.1~5-7.4 范围已明确
- ✅ 越界检查零违规
- ✅ 仅修改 docs

**下一步：5-7.1/5-7.2/5-7.3 三条路径全部验收 pass ✅，5-7.4 总验收完成，5-7 阶段结束。**
