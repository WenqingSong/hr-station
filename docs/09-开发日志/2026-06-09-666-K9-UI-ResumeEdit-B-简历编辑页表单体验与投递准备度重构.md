# K9-UI-ResumeEdit-B：简历编辑页表单体验与投递准备度重构

> 日期：2026-06-09 | 性质：前端重构，未修改后端/数据库

---

## 核心变更

### 1. 新增 AutoGrowTextarea 组件

- 文件：`web/src/app/components/ui/auto-grow-textarea.tsx`（新建，92 行）
- 实现：`ref + scrollHeight` 自动增高，支持 `minRows` / `maxRows` 控制
- 保持与 Textarea 一致的圆角、边框、focus 样式
- 不引入新依赖，纯原生 textarea + React hooks

### 2. YearMonthPicker 固定宽度

- 文件：`web/src/app/components/resume-editor/YearMonthPicker.tsx`
- 修改：`w-full` → `w-[136px]`（trigger 按钮）
- 影响：教育经历 / 项目经历 / 实习经历共 3 处 6 个 picker

### 3. ResumeEditorSectionCard 折叠动画

- 文件：`web/src/app/components/resume-editor/ResumeEditorSectionCard.tsx`
- 修改：条件渲染 → CSS `grid-template-rows: 0fr → 1fr` 过渡动画
- Description 始终在 body 内部，不再在折叠/展开时移动到不同容器
- 保留 ChevronDown 旋转动画（`transition-transform duration-200`）
- 影响：6 个可折叠模块（教育 / 项目 / 技能 / 实习 / 证书获奖 / 作品链接）

### 4. 简历名称可用状态重做

- 文件：`web/src/app/pages/MyResumesPage.tsx`
- 修改：每个状态有专用 icon（HelpCircle / Loader2 / CheckCircle2 / XCircle / AlertCircle）
- 状态条固定 `min-h-[48px]`，防止状态切换时布局跳动
- idle 从灰色改为蓝色提示（`bg-blue-50/60 border-blue-100 text-blue-600`）
- checking 状态加 `animate-spin` 动画
- 文案精简：available 改为"名称可用"（新建）/ "当前名称可继续使用"（编辑）

### 5. Textarea → AutoGrowTextarea 替换

- 文件：`web/src/app/pages/MyResumesPage.tsx`
- 替换字段：
  - 论文/学术成果：minRows=2 maxRows=10
  - 项目描述：minRows=4 maxRows=12
  - 实习描述：minRows=4 maxRows=12
  - AI 应用技能：minRows=3 maxRows=8
  - 证书说明：minRows=2 maxRows=6
  - 获奖说明：minRows=2 maxRows=6
  - 补充信息：minRows=4 maxRows=15
- 所有字段移除 `resize-none`，AutoGrowTextarea 内置 `resize-none`

### 6. "填入体验内容"按钮

- 文件：`web/src/app/pages/MyResumesPage.tsx`
- 位置：新建页顶部蓝色提示条中
- 行为：调用 `getResumeById(201)` → adapter 转换 → 填入所有表单 state → **不自动保存**
- 有 loading 状态和错误提示
- 编辑已有简历时不显示该按钮

### 7. 投递准备度右侧栏

- 文件：`web/src/app/components/resume-editor/ResumeEditorStatusSidebar.tsx`（完全重写，146 行）
- 标题："投递准备度"
- 展示：进度环 + 百分比 + 分类完成状态列表 + 下一步建议
- 分类：基础信息 / 教育经历 / 技能能力 / 实践经历 / AI 应用 / 证书获奖
- 纯前端计算（useMemo），不请求后端
- Shell 中 sidebar 传递至 `ResumeEditorShell`

### 8. Shell 侧栏接入

- 文件：`web/src/app/pages/MyResumesPage.tsx`
- `ResumeEditorShell` 新增 `sidebar` prop 传入 `ResumeEditorStatusSidebar`

---

## 修改清单

| 文件 | 操作 | 行数变化 |
| -- | -- | -- |
| `web/src/app/components/ui/auto-grow-textarea.tsx` | **新建** | +92 |
| `web/src/app/components/resume-editor/YearMonthPicker.tsx` | 修改 1 行 | `w-full` → `w-[136px]` |
| `web/src/app/components/resume-editor/ResumeEditorSectionCard.tsx` | 重写 body 渲染 | ~15 行变化 |
| `web/src/app/components/resume-editor/ResumeEditorStatusSidebar.tsx` | 完全重写 | 186→146 行 |
| `web/src/app/pages/MyResumesPage.tsx` | 大量修改 | ~150 行新增 |
| `web/src/api/resumes.ts` | 未修改 | — |
| `server/` | 未修改 | — |
| `server/sql/init.sql` | 未修改 | — |

---

## 构建 / 检查结果

| 检查 | 结果 | 说明 |
| -- | -- | -- |
| npm build | ✅ 通过 | 644 kB JS bundle |
| go build | ✅ 通过 | 零错误 |
| go vet | ✅ 通过 | 零错误 |
| go test | 6 FAIL | **pre-existing**（K9-2 agent/service 测试） |
| check-encoding | ✅ 通过 | 无编码错误 |
| lint (edited files) | ✅ 通过 | 零 lint 错误 |

---

## 禁用文案检查

- 所有 `undefined` 命中均为合法 TypeScript 代码（`?? undefined`、`|| undefined`）
- 无 "Invalid Date" / "mock_agent" / "mock_quality" 用户可见文案
- 无 "演示数据" / "假数据" 用户可见文案
- "填入体验内容" 文案符合规范，未使用禁用词

---

## 未修改范围确认

- ✅ 未修改后端 API
- ✅ 未修改数据库结构
- ✅ 未修改 `server/sql/init.sql`
- ✅ 未修改真实简历数据
- ✅ 未执行 `make reset-up`
- ✅ 未执行 Git 命令
- ✅ 未修改 `.env` / `.env.example`
- ✅ K9-1 / K9-2 / K9-3 / K9-4 核心业务逻辑未受影响
