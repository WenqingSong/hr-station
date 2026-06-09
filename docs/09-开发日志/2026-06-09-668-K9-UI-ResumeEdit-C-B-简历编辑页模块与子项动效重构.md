# K9-UI-ResumeEdit-C-B：简历编辑页模块与子项动效重构

> 日期：2026-06-09 | 性质：前端动效重构，未修改后端/数据库

---

## 核心变更

### 1. 父模块展开/折叠动画升级

- 文件：`web/src/app/components/resume-editor/ResumeEditorSectionCard.tsx`
- `duration-200` → `duration-[240ms]`（200ms→240ms）
- `ease-out` → `ease-[cubic-bezier(0.22,1,0.36,1)]`（更自然的 ease-out-quint 曲线）
- 移除 `padding` 从 `transition-[...]` 范围，仅保留 `grid-template-rows, opacity`
- 6 个可折叠模块统一升级

### 2. 新增 AnimatedListItem 组件

- 文件：`web/src/app/components/ui/animated-list-item.tsx`（新建，60 行）
- Enter：CSS transition（opacity 0→1 + translateY 6px→0 + scale 0.985→1），180ms
- Exit：CSS transition（opacity→0 + translateY→-4px + scale→0.985 + max-h→0），160ms
- 零新依赖，纯 CSS transition
- 支持 `motion-reduce` 无障碍

### 3. Delayed remove 移除动画系统

- 文件：`web/src/app/pages/MyResumesPage.tsx`
- 新增 `removingItemKeys: Set<string>` state
- 新增 `markRemoving(key, removeFn)` 工具函数
- 删除按钮点击 → 标记 removing → 170ms exit 动画 → 真正 filter 移除
- 删除按钮 removing 状态 disabled + opacity-50

### 4. 7 类列表全部接入 AnimatedListItem

| 列表 | key 策略 | AnimatedListItem | delayed remove |
| -- | -- | -- | -- |
| 教育经历 | `key={edu.id}` ✅ | ✅ | ✅ |
| 项目经历 | `key={proj.id}` ✅ | ✅ | ✅ |
| 实习经历 | `key={intern.id}` ✅ | ✅ | ✅ |
| 证书 | `key={cert.id}` ✅ | ✅ | ✅ |
| 获奖 | `key={award.id}` ✅ | ✅ | ✅ |
| 作品链接 | `key={work.id}` ✅ | ✅ | ✅ |
| 外语考试 | **`key={le.id}` ✅（修复）** | ✅ | ✅ |

### 5. 外语考试 key 修复（P0）

- `interface LanguageExamDraft` 新增 `id: number`
- `addLanguageExam` 改为递增 id（`Math.max(...ids) + 1`）
- `removeLanguageExam(id)` / `updateLanguageExam(id, ...)` 改为按 id 操作
- 渲染 `key={le.id}`（不再使用 index）
- 编辑模式回填 + 演示填充均补 id

### 6. 保存安全

- `handleSave` 生成 payload 前过滤 `removingItemKeys` 中的 7 类 item
- Key 命名：`education:${id}` / `project:${id}` / `internship:${id}` / `certificate:${id}` / `award:${id}` / `work:${id}` / `language:${id}`
- 确保了 exit 动画进行中的删除项不会被误保存

---

## 修改清单

| 文件 | 操作 | 说明 |
| -- | -- | -- |
| `web/src/app/components/ui/animated-list-item.tsx` | **新建** | 60 行 |
| `web/src/app/components/resume-editor/ResumeEditorSectionCard.tsx` | 修改 2 行 | duration + easing |
| `web/src/app/pages/MyResumesPage.tsx` | 大量修改 | ~200 行变化 |

---

## 构建/检查结果

| 检查 | 结果 | 说明 |
| -- | -- | -- |
| npm build | ✅ | 647 kB JS |
| go build | ✅ | 零错误 |
| go vet | ✅ | 零错误 |
| go test | 6 FAIL | pre-existing（K9-2 agent/service） |
| check-encoding | ✅ | 仅文档中已知乱码引用 |
| motion imports | ✅ | 零新增 |

---

## 未修改范围确认

- ✅ 未修改后端 API
- ✅ 未修改数据库结构
- ✅ 未修改 `server/sql/init.sql`
- ✅ 未引入新依赖
- ✅ 未使用 motion/framer-motion
- ✅ 未执行 Git 命令
