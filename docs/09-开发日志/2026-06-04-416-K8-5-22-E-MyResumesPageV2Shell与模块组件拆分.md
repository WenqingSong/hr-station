# K8-5-22-E — MyResumesPage V2 Shell 与模块组件拆分

## 性质
implementation（前端页面重构，不改后端/Agent/SQL）。

## 本阶段目标

将 MyResumesPage 从大单文件拆分为 V2 组件结构，为后续 K8-5-22-F/G/H 分批实现 90 字段做准备。

不实现 90 字段，不启用 schema_version=2 写入。

## 新增组件清单

### 布局组件

| 组件 | 文件 | 职责 |
|------|------|------|
| `ResumeEditorShell` | `resume-editor/ResumeEditorShell.tsx` | 整体布局：max-w-[1280px]、标题、左右双栏、保存区 |
| `ResumeEditorSectionCard` | `resume-editor/ResumeEditorSectionCard.tsx` | 通用折叠卡片：标题、描述、badge、折叠/展开、必填标识 |
| `ResumeEditorStatusSidebar` | `resume-editor/ResumeEditorStatusSidebar.tsx` | 右侧状态栏：完成度、AI 就绪、四功能状态、保存状态 |
| `ResumeEditorSaveCard` | `resume-editor/ResumeEditorSaveCard.tsx` | 底部保存卡：取消/保存按钮、成功/错误提示 |
| `ResumeEditorModulePlaceholder` | `resume-editor/ResumeEditorModulePlaceholder.tsx` | 模块占位符：标题、描述、后续阶段标注 |

### 10 模块组件

| 模块 | 文件 | 默认状态 | 实现阶段 |
|------|------|:---:|------|
| 1. 简历基础 | `modules/ResumeBasicsSection.tsx` | 展开（不可折叠） | K8-5-22-E（含 V1 内容） |
| 2. 基础信息 | `modules/BasicInfoSection.tsx` | 展开 | K8-5-22-F |
| 3. 求职意向 | `modules/JobIntentionSection.tsx` | 展开 | K8-5-22-F |
| 4. 教育经历 | `modules/EducationSection.tsx` | 展开 | K8-5-22-F |
| 5. 项目经历 | `modules/ProjectSection.tsx` | 展开 | K8-5-22-G |
| 6. 技能信息 | `modules/SkillsSection.tsx` | 展开 | K8-5-22-G |
| 7. 实习经历 | `modules/InternshipSection.tsx` | 折叠 | K8-5-22-G |
| 8. 证书与语言能力 | `modules/CertificatesSection.tsx` | 折叠（占位） | K8-5-22-H |
| 9. 获奖与竞赛 | `modules/AwardsSection.tsx` | 折叠（占位） | K8-5-22-H |
| 10. 作品链接与补充信息 | `modules/PortfolioExtraSection.tsx` | 折叠（含 V1 自我评价） | K8-5-22-H |

### 栏导出文件

| 文件 | 职责 |
|------|------|
| `resume-editor/index.ts` | 统一导出所有组件和类型 |

## MyResumesPage 接入情况

- **方案**：方案 B（低风险），将现有 V1 内容放入新模块组件，保持所有 state/logic 不变
- **handleSave**：未修改，仍生成 V1 payload
- **createResume/updateResume**：未修改
- **loading/error/retry**：未修改
- **页面结构**：`ResumeEditorShell`（标题 + 两栏 + 保存区）替代旧的 `ContentPageShell`
- **旧模块内容**：8 个 V1 模块内容迁入对应 V2 模块组件，字段和逻辑不变
- **新模块**：证书（8）和获奖（9）为占位符，模块 10 承载旧自我评价 + 占位补充信息

## V1 保存逻辑确认

- `handleSave` 函数未修改 ✅
- V1 payload 结构未改变（name, phone, email, location, target_city, target_role, education, experience, project, skills, self_evaluation） ✅
- `createResume` / `updateResume` 调用未修改 ✅
- 未引入 `schema_version: 2` ✅
- 未引入 `CreateResumeRequestV2` ✅

## 验收结果

| 验收项 | 结果 |
|--------|------|
| `npm run build` | ✅ 1.90s |
| `go test ./...` | ✅ 全部通过 |
| `go build ./...` | ✅ 通过 |
| `go vet ./...` | ✅ 通过 |
| API `/api/jobs?page=1&page_size=5` | ✅ 200 |
| API `/api/jobs/10001` | ✅ 200 |
| API `/api/jobs/facets` | ✅ 200 |
| API `/api/resumes` | ✅ 401（正常） |
| 新组件 grep | ✅ 40+ 处命中 |
| V2 写入禁止检查 | ✅ 0 命中 |
| 后端未误改检查 | ✅ 0 命中 |

## 未修改范围

- 未修改 Go 源码 ✅
- 未修改后端 handler/service/repository/model ✅
- 未修改数据库 SQL ✅
- 未修改 Agent / Prompt / Schema / Mock ✅
- 未修改 types/resume.ts / api/resumes.ts / adapter / viewModels ✅
- 未修改 ResumeManagementPage / ResumeComparisonPage / AI 页面 ✅
- 未新增 API 路径 ✅
- 未启用 schema_version=2 写入 ✅
- 未安装依赖 / 未迁移目录 ✅

## 未执行 Git 命令

本次未执行任何 Git 命令，也未提交代码。

## 风险与边界

- 本阶段只完成 MyResumesPage V2 Shell 和模块边界
- 模块 2-7 仍使用 V1 旧表单内容，V2 新字段留给 F/G
- 模块 8-9 为纯占位符，留给 H
- 模块 10 后半部分（作品链接）为占位符，留给 H
- V2 payload 启用留给后续阶段
