# K9-UI-ResumeModal-B：通用 ResumePickerModal 重构与岗位详情页接入

> 日期：2026-06-08 | 性质：前端重构 | 不改后端

---

## 1. A 阶段诊断问题修复清单

| 问题 | 根因 | 修复结果 |
|------|------|:--:|
| 弹窗层级不专业 | 无 Portal，`z-50` 与 Navbar 同级 | ✅ `createPortal` + `z-[10000]` |
| 背景文字视觉干扰 | `bg-slate-950/30` 无 blur | ✅ `bg-slate-950/40 backdrop-blur-sm` |
| body + modal 双滚动 | 无 body scroll lock | ✅ `document.body.style.overflow = "hidden"` |
| 无 ARIA | 纯 div | ✅ `role="dialog" aria-modal="true" aria-labelledby` |
| 无 ESC/遮罩关闭 | 无键盘事件 | ✅ ESC + 点击遮罩关闭 |
| 复用性差 | 仅单选 | ✅ `selectionMode: "single" | "compare"` |
| 文案/请求耦合 | 硬编码 + 内部 API | ✅ thin wrapper 解耦 |

---

## 2. 修改清单

| 文件 | 修改 | 类型 |
|------|------|:--:|
| `web/src/app/components/ResumePickerModal.tsx` | **新增**：通用简历选择弹窗（~370 行） | 新增 |
| `web/src/app/components/SelectResumeModal.tsx` | **重写**：thin wrapper（~170 行，原 301 行） | 修改 |
| `web/src/app/components/ai/JobDetailAIDecision.tsx` | **无需修改**（接口兼容） | 不变 |

---

## 3. ResumePickerModal 实现

| 能力 | 结果 |
|------|:--:|
| Portal (`createPortal` → `document.body`) | ✅ |
| Body scroll lock（含 scrollbar 补偿） | ✅ |
| Overlay: `bg-slate-950/40 backdrop-blur-sm` | ✅ |
| Panel: `bg-white`（完全不透明，无自身 blur） | ✅ |
| z-index: `z-[10000]` (overlay) + `z-[10001]` (panel) | ✅ |
| ARIA: `role="dialog" aria-modal="true" aria-labelledby` | ✅ |
| ESC 关闭 | ✅ |
| 点击遮罩关闭（`onMouseDown` target 判断） | ✅ |
| Header / body / footer 三段式 | ✅ |
| 列表独立滚动（`overflow-y-auto`） | ✅ |
| Footer 固定（`shrink-0`） | ✅ |
| 搜索（name / direction / profile / tags） | ✅ |
| loading / error(retry) / empty / search-empty | ✅ |
| selectionMode: `"single"` / `"compare"` | ✅ |
| single: 最多 1 个，单选圆点 | ✅ |
| compare: 最多 2 个，序号方框 | ✅ |

---

## 4. SelectResumeModal thin wrapper

保持向后兼容：
- Props: `isOpen`, `onClose`, `onConfirm(resumeId)`, `scene`, `jobTitle`, `initialSelectedResumeId`
- 负责：API 加载 + 默认选中 + scene 文案映射
- UI 委托：`ResumePickerModal`
- `JobDetailAIDecision` **零修改**

---

## 5. 源码检查

| 项 | 结果 |
|----|:--:|
| `createPortal` | ✅ |
| `role="dialog"` | ✅ |
| `aria-modal` | ✅ |
| `backdrop-blur-sm` | ✅ |
| `body.style.overflow = "hidden"` | ✅ |
| `selectionMode` | ✅ |
| `compare` mode 逻辑 | ✅ |
| `z-[10000]` / `z-[10001]` | ✅ |
| 无 `bg-white/90` 或 `bg-white/80` | ✅ |

---

## 6. 构建 / 测试结果

| 验收项 | 结果 | 归因 |
|--------|:--:|------|
| npm build | ✅ | JS ~601KB (+2KB vs 上次) |
| go build ./... | ✅ | — |
| go vet ./... | ✅ | — |
| go test ./... | ⚠️ 6 FAIL | **pre-existing** |
| check-encoding | ⚠️ FAIL | **pre-existing** (SQL dump) |

---

## 7. 问题清单

| 级别 | 问题 | 证据 | 建议阶段 |
|:--:|------|------|:--:|
| P3 | pre-existing go test 6 FAIL | agent/service | 非阻断 |
| P3 | pre-existing check-encoding | SQL dump | 非阻断 |

---

## 8. 是否进入下一阶段

- ✅ **通过**：K9-UI-ResumeModal 全链路（A 诊断 → B 重构）完成

---

## 9. 未修改范围确认

- ✅ 未修改 K9-1/K9-2/K9-3/K9-4
- ✅ 未修改 `server/sql/init.sql` 和数据库结构
- ✅ 未删除真实简历
- ✅ 未执行 `make reset-up`
- ✅ 未执行任何 Git 命令

---

## 10. docs 更新清单

| 操作 | 文件 |
|:--:|------|
| ✅ 新增 | `docs/09-开发日志/2026-06-08-648-K9-UI-ResumeModal-B-通用ResumePickerModal重构与岗位详情页接入.md` |
| ✅ 更新 | `docs/下一步任务.md` |
| ✅ 更新 | `docs/变更记录.md` |

---

## 11. Git 约束确认

零 Git 命令执行。
