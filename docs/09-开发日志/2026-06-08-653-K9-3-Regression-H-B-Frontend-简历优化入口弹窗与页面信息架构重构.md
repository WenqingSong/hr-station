# K9-3-Regression-H-B-Frontend — 简历优化入口弹窗与页面信息架构重构

日期：2026-06-08
性质：前端重构。5 文件修改 + 2 文件新增。

## 背景

K9-3-Regression-H-A 诊断发现：
- 页面 `max-w-[1440px]` 过宽，视觉偏左
- 4 个入口中 2 个不传 jobId → 静默降级为 general 模式
- 无 JobPicker/JobSelector 组件
- 无 OptimizationEntryModal
- target_job 后端已支持但前端未作为主路径
- 右侧栏信息密度极低
- 旧版 `resume-optimization/` 下 4 个组件为 dead code

## 变更内容

### 1. 新增 JobPicker

**文件**：`web/src/app/components/JobPicker.tsx`（175 行）

能力：
- 搜索岗位（标题/方向/城市/部门/类别）
- 展示岗位标题/招聘类型/城市/方向/部门/类别
- 单选，蓝色选中态
- loading / error / empty / search empty 四种状态
- 调用 `getJobs()` API（不新增后端）

### 2. 新增 OptimizationEntryModal

**文件**：`web/src/app/components/OptimizationEntryModal.tsx`（230 行）

能力：
- Portal + `backdrop-blur-sm` + body lock + ESC/遮罩关闭 + ARIA
- 两步式交互：Step 1 选择简历 → Step 2 选择目标岗位
- 支持 `initialResumeId` 自动预选（跳过 step 1）
- 确认后 navigate 到 `/resume-optimization?resumeId=X&jobId=Y`
- "继续通用优化" 作为次要入口（弱化）
- 主按钮："开始针对岗位优化"

### 3. 重构 ResumeOptimizationPage

**文件**：`web/src/app/pages/ResumeOptimizationPage.tsx`（重写，原 661 行 → 530 行）

核心改动：

#### 3.1 不再静默 general
- `resumeId` 存在但无 `jobId` 时 → 不再自动 POST general
- 显示"选择目标岗位"引导页，提供"选择目标岗位"和"继续通用优化"两个按钮
- `mode=general` URL 参数可显式进入 general

#### 3.2 布局重构
- `max-w-[1280px]` 居中（原 `max-w-[1440px]`）
- `lg:grid-cols-[minmax(0,1fr)_320px]` 两栏
- 新增顶部"当前优化对象"卡片：简历 × 岗位 × 模式 + 重新选择按钮

#### 3.3 右侧栏升级
- 当前简历卡片：名称 + 画像 + 技能标签
- 目标岗位卡片：标题 + 方向/城市 + 优化模式
- 操作按钮区：
  - 重新选择简历/岗位
  - 生成/重试/重新生成深度优化
  - 进入匹配分析（有 jobId+resumeId 时）
  - 返回岗位详情（有 jobId 时）

#### 3.4 DeepView 信息重组
- Top 3 必须修改（P1/P2/P3 优先级标签）
- JD 要求补强（evidence + keyword 合并展示）
- 结构与表达优化（折叠）
- 改写示例（原有视觉重点保留）
- 风险提醒
- 投递重点
- 长期补强建议（折叠/降级）

#### 3.5 状态机改造
- 新增 `"idle"` phase：等待用户选择岗位
- `"loading_latest"` 仅在 mode 确定后进入
- general 需要显式 URL 参数或用户点击"继续通用优化"

### 4. 改造 ResumeManagementPage

**文件**：`web/src/app/pages/ResumeManagementPage.tsx`

- 新增 `OptimizationEntryModal` 导入和状态
- SmartAssistant 添加 `onOptimize` 回调
- "优化选中简历" 从 `<Link to={...}>` 改为 `onClick → 打开弹窗`
- 主组件渲染 `OptimizationEntryModal`

### 5. 改造 ResumeDetailPage

**文件**：`web/src/app/pages/ResumeDetailPage.tsx`

- 新增 `OptimizationEntryModal` 导入和状态
- "去简历优化" 从 `<Link to={...}>` 改为 `onClick → 打开弹窗`
- 主组件渲染 `OptimizationEntryModal`

## 入口链路对比

| 入口 | 修复前 | 修复后 |
|------|--------|--------|
| 岗位详情页 | navigate 直接跳转 | 不变（已有 jobId+resumeId） |
| 匹配分析页 | navigate 直接跳转 | 不变（已有 jobId+resumeId） |
| 简历管理页 | `/resume-optimization?resumeId=X` → general | 打开 OptimizationEntryModal → 选岗位 |
| 简历详情页 | `/resume-optimization?resumeId=X` → general | 打开 OptimizationEntryModal → 选岗位 |

## target_job 主路径

| 场景 | 结果 |
|------|------|
| `resumeId+jobId` | ✅ 自动 target_job latest-first |
| `resumeId` only | ✅ 引导选择岗位，不自动 general |
| 用户点击"继续通用优化" | ✅ 显式 general |
| `mode=general` URL | ✅ 显式 general |

## 构建 / 测试

| 验收项 | 结果 |
|--------|------|
| npm build | ✅ 619.35 KB JS |
| go build ./... | ✅ |
| go vet ./... | ✅ |
| go test ./... | 6 FAIL — **pre-existing**（agent/service K9-2 测试） |
| check-encoding | ✅ OK |

## 未修改范围确认

- 未修改 K9-1 岗位推荐业务逻辑
- 未修改 K9-2 岗位匹配分析后端
- 未修改 K9-3 后端 API / Service / Agent / Repository / Model
- 未修改 K9-4 简历版本对比
- 未修改 server/sql/init.sql
- 未修改数据库结构
- 未删除真实简历
- 未执行 reset-up
- 未执行 Git 命令
