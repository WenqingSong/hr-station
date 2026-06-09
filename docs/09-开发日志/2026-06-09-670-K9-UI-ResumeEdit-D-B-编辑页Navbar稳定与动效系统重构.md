# K9-UI-ResumeEdit-D-B：编辑页 Navbar 稳定与动效系统重构

> 日期：2026-06-09 | 性质：前端动效+Navbar 修复，未修改后端/数据库

---

## 核心变更

### 1. Navbar 稳定性修复

- 文件：`web/src/app/components/Navbar.tsx`
- 编辑页路由（`/resumes/new`、`/resumes/edit`、`/resumes/edit?id=`）禁用 auto-hide：Navbar 始终固定显示
- 非编辑页增加 30px delta 阈值：`Math.abs(delta) < 30` 时不改变显隐状态
- 保留原有顶部阈值（scrollY≤20 强制显示）
- Effect 依赖 `location.pathname`

### 2. 父模块折叠动画修复

- 文件：`web/src/app/components/resume-editor/ResumeEditorSectionCard.tsx`
- 移除 opacity transition（仅保留 `grid-template-rows`）
- 添加 `overflow-anchor: none`（inline style + Tailwind class 双保险）
- 保留 240ms + cubic-bezier 过渡

### 3. AnimatedListItem 完全重写

- 文件：`web/src/app/components/ui/animated-list-item.tsx`（完全重写，78 行）
- **移除 scale**：避免亚像素文本模糊
- **移除 max-h-[2000px]**：exit 改用 `getBoundingClientRect().height` 测量真实高度
- **Enter**：CSS @keyframes `resume-item-enter`（opacity 0→1 + translateY 6px→0），180ms
- **Exit 两阶段**：
  - Phase 1：锁定测量高度
  - Phase 2：容器高度→0（160ms）+ 内容 fade-out（80ms translateY -4px + opacity 0）
- 移除无用的 `mounted` state

### 4. CSS 动效更新

- 文件：`web/src/styles/tailwind.css`
- `resume-item-enter` keyframes：移除 scale，duration 140ms→180ms
- 新增 `resume-item-exit-content` keyframes（fade-out + translateY -4px，80ms）

### 5. timeout cleanup

- 文件：`web/src/app/pages/MyResumesPage.tsx`
- 新增 `removeTimeoutsRef` 存储 timeout ID
- `useEffect` cleanup 清理所有 pending timeout
- 新增 `useRef` import

---

## 修改清单

| 文件 | 操作 | 说明 |
| -- | -- | -- |
| `web/src/app/components/Navbar.tsx` | ~25 行变化 | 编辑页禁用 auto-hide + delta 阈值 |
| `web/src/app/components/resume-editor/ResumeEditorSectionCard.tsx` | ~5 行变化 | 移除 opacity + overflow-anchor |
| `web/src/app/components/ui/animated-list-item.tsx` | 完全重写（78 行） | real height exit + @keyframes enter |
| `web/src/styles/tailwind.css` | ~15 行变化 | enter 180ms + exit content keyframes |
| `web/src/app/pages/MyResumesPage.tsx` | ~15 行变化 | timeout cleanup + useRef import |

---

## 构建/检查结果

| 检查 | 结果 | 说明 |
| -- | -- | -- |
| npm build | ✅ | 647 kB |
| go build | ✅ | 零错误 |
| go vet | ✅ | 零错误 |
| go test | 6 FAIL | pre-existing（K9-2） |
| check-encoding | ✅ | 通过 |
| motion imports | ✅ | 零新增 |

---

## 未修改范围确认

- ✅ 未修改后端 API
- ✅ 未修改数据库
- ✅ 未引入新依赖
- ✅ 未使用 motion/framer-motion
- ✅ 未执行 Git 命令
