# K8-5-12-B：统一 Loading / Empty / Error 设计规范冻结（仅规范，不改代码）

## 一、本次范围确认
只做设计规范冻结，未修改任何源码。

## 二、统一状态设计规范

### 1. 页面级 Loading
- **组件**：`PageLoading`（已有）
- **视觉**：居中 `Loader2 h-6 w-6 animate-spin text-[#0052D9]` + `text-sm text-gray-400`
- **文案**：可自定义（默认"正在加载…"）
- **适用**：详情页、推荐页、匹配/优化/对比页初次加载

### 2. 区域级 Skeleton
- **已有**：`JobCardSkeleton` + `JobListSkeleton`（保留）
- **新增**：`DetailSkeleton`（标题+多行文本），`AIResultSkeleton`（分析卡片），`ResumeSkeleton`（表单骨架）
- **视觉**：`bg-gray-100 rounded animate-pulse`（复用 `ui/Skeleton` 基类）
- **规则**：卡片页用卡片骨架，详情用标题+文本，AI 结果用结果卡骨架

### 3. 按钮 Pending
- **组件**：不独立抽象，在各按钮内局部实现
- **视觉**：`Loader2 h-4 w-4 animate-spin` + disabled + 文案变为"生成中…/分析中…/保存中…"
- **规则**：`disabled:cursor-not-allowed disabled:opacity-50`
- **参考**：Login/Register 的按钮 pending 为最佳实践

### 4. AI 生成中
- **组件**：`AIProgressState`（新增）
- **视觉**：白卡 `rounded-2xl bg-white ring-1 ring-gray-100 p-6` + 步骤列表 + spinner
- **文案**：三步 — `正在读取简历` → `正在分析岗位` → `正在生成建议`，每步有灰色→蓝色切换
- **规则**：不能只有按钮禁用，必须有进度反馈

### 5. Empty State
- **组件**：`PageEmpty`（已有，完善文案）
- **视觉**：图标 + `text-sm font-semibold text-gray-900` 标题 + `text-sm text-gray-500` 说明 + CTA
- **文案规则**：必须具体，不能只写"暂无数据"
  - 无简历 → "还没有可用于分析的简历" + "去创建简历"
  - 无岗位 → "未找到匹配的岗位" + "清除筛选"
  - 无推荐 → "当前没有推荐结果" + "上传简历后重试"

### 6. Error State
- **组件**：`PageError`（已有）
- **视觉**：`AlertCircle h-10 w-10 text-gray-300` + 标题 + 说明 + 重试按钮
- **规则**：所有失败操作提供"重试"按钮，不允许静默失败

### 7. Retry 策略
- 所有可重试操作必须有重试按钮
- 重试时显示 pending 状态
- 重试按钮复用原请求逻辑

### 8. 登录态 Loading
- 按钮 pending + disabled
- 成功后短时显示"正在进入…"
- 防重复提交

## 三、页面映射表

| 页面 | Loading | Skeleton | Pending | Empty | Error | Retry |
|------|---------|----------|---------|-------|-------|-------|
| /jobs | ✅ JobListSkeleton | ✅ | — | ✅ | ✅ PageError | ✅ |
| /jobs/:id | ✅ PageLoading | DetailSkeleton(新) | — | ✅ JD不存在 | ✅ PageError | — |
| /recommend | ✅ PageLoading | — | ✅ 生成中… | ✅ 无推荐 | ✅ PageError | ✅ |
| /match | ✅ PageLoading | AIResultSkeleton(新) | ✅ 分析中… | ✅ 无简历 | ✅ PageError | ✅ |
| /resume-optimization | ✅ PageLoading | AIResultSkeleton(新) | ✅ 优化中… | ✅ 无简历 | ✅ PageError | ✅ |
| /resumes/compare | ✅ PageLoading | — | ✅ 对比中… | ✅ 无版本 | ✅ PageError | ✅ |
| /resumes | ✅ PageLoading | ResumeSkeleton(新) | ✅ 保存中… | ✅ 无简历 | ✅ PageError | — |
| /resumes/edit | ✅ PageLoading | ResumeSkeleton(新) | ✅ 保存中… | — | ✅ PageError | ✅ |
| /login | — | — | ✅ 登录中… | — | ✅ 红色提示 | — |
| /register | — | — | ✅ 注册中… | — | ✅ 红色提示 | — |

## 四、组件演进建议

| 组件 | 建议 | 原因 |
|------|------|------|
| `PageLoading` | 保留 + 全面接入 | 设计完善，仅缺接入 |
| `PageError` | 保留 + 各页面接入 children 重试按钮 | 同上 |
| `PageEmpty` | 保留 + 完善各页面具体文案 | 同上 |
| `JobCardSkeleton` | 保留 | 岗位列表专用 |
| `ui/Skeleton` | 保留为基础 | 所有新骨架的基类 |
| `DetailSkeleton` | 新增 | 详情页标题+文本骨架 |
| `AIResultSkeleton` | 新增 | AI 结果卡骨架 |
| `AIProgressState` | 新增 | AI 生成中步骤反馈 |
| `ResumeSkeleton` | 新增 | 简历表单骨架 |
| LoadingButton | **暂不抽象** | 按钮 pending 场景差异大，先在局部实现 |

## 五、设计 token

| 类型 | 推荐样式 |
|------|---------|
| spinner | `Loader2 h-4 w-4 animate-spin`（按钮）/ `h-6 w-6`（页面）|
| 页面 loading | `py-16 text-center` + `text-sm text-gray-400` |
| skeleton | `bg-gray-100 rounded animate-pulse` |
| error 标题 | `text-sm font-semibold text-gray-900` |
| error 说明 | `text-sm text-gray-500` |
| empty 标题 | `text-sm font-semibold text-gray-900` |
| empty 说明 | `text-sm text-gray-500` |
| 重试按钮 | 蓝色主按钮 `bg-[#0052D9]` / 灰色次按钮 |
| AI pending 卡 | `rounded-2xl bg-white ring-1 ring-gray-100 p-6` |
| 按钮 disabled | `disabled:cursor-not-allowed disabled:opacity-50` |

## 六、P0/P1/P2 优先级

| 级别 | 问题 | 修复阶段 |
|------|------|---------|
| **P0** | AI 生成中无骨架/无步骤反馈 | K8-5-12-E |
| **P0** | AI 三页面按钮无 pending/disabled | K8-5-12-E |
| **P1** | 详情页/推荐页未用 PageLoading | K8-5-12-D |
| **P1** | 详情页无 DetailSkeleton | K8-5-12-C/D |
| **P1** | 空状态/错误态未统一 PageState | K8-5-12-D/E/F |
| **P2** | 骨架动画统一 shimmer | K8-5-12-Z |

## 七、后续 prompt 拆分

| 编号 | 内容 | 主要文件 |
|------|------|---------|
| **K8-5-12-C** | DetailSkeleton / AIResultSkeleton / AIProgressState / ResumeSkeleton 组件补齐 | `PageState.tsx` / 新增骨架文件 |
| **K8-5-12-D** | 岗位列表+详情 loading 体验修复 | `JobListPage` / `JobDetailPage` / `JobDetailBanner` / `JobDetailJD` |
| **K8-5-12-E** | AI 三页面 pending/error/retry 修复 | `JobRecommendationPage` / `MatchAnalysisPage` / `ResumeOptimizationPage` |
| **K8-5-12-F** | 简历页面 loading/empty/error 修复 | `ResumeManagementPage` / `MyResumesPage` / `ResumeComparisonPage` |
| **K8-5-12-Z** | 全站加载体验验收与冻结 | 只读验收 |

## 八、未修改范围

- 未修改源码/CSS/后端/数据库/接口/.env/LLM
- 未安装依赖/未迁移目录

## 九、Git 约束确认

本次未执行任何 Git 命令，也未提交代码。
