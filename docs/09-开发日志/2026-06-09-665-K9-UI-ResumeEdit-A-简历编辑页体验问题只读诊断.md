# K9-UI-ResumeEdit-A：简历编辑页体验问题只读诊断

> 日期：2026-06-09 | 性质：只读诊断，未修改任何源码

---

## 一、当前状态确认

| 项 | 结果 |
| -- | -- |
| 编辑页文件 | `web/src/app/pages/MyResumesPage.tsx`（491 行） |
| 路由 | `/resumes/edit?id=X`（编辑）、`/resumes/new`（新建） |
| 模块组件 | 10 个 Section wrapper 组件，位于 `web/src/app/components/resume-editor/modules/` |
| 折叠组件 | `ResumeEditorSectionCard.tsx`（145 行） |
| 年月选择器 | `YearMonthPicker.tsx`（160 行，自定义组件） |
| Textarea | `web/src/app/components/ui/textarea.tsx`（32 行，基础 wrapper） |
| 右侧状态栏 | `ResumeEditorStatusSidebar.tsx` 组件存在，但 **当前编辑页不渲染**（K8-5-29-E-R 已移除） |
| 其他页面复用 | `YearMonthPicker`、`ResumeEditorSectionCard` 仅用于 MyResumesPage，**无跨页面复用** |
| 示例数据 | `init.sql` 中 user_id=2 Arthur 有 9 份示范简历 |

---

## 二、文件定位

| 功能 | 文件 | 说明 |
| -- | -- | -- |
| 新建/编辑简历页 | `/workspace/web/src/app/pages/MyResumesPage.tsx` | 491 行，单文件承载全部表单状态和逻辑 |
| 页面布局壳 | `/workspace/web/src/app/components/resume-editor/ResumeEditorShell.tsx` | 主栏 + 可选右侧栏（280px），max-w-[1280px] |
| 分区卡片 | `/workspace/web/src/app/components/resume-editor/ResumeEditorSectionCard.tsx` | 可折叠 Section Card，提供展开/折叠、badge、required |
| 简历基础模块 | `/workspace/web/src/app/components/resume-editor/modules/ResumeBasicsSection.tsx` | 不可折叠，必填 |
| 基础信息模块 | `/workspace/web/src/app/components/resume-editor/modules/BasicInfoSection.tsx` | 不可折叠，必填 |
| 教育经历模块 | `/workspace/web/src/app/components/resume-editor/modules/EducationSection.tsx` | 可折叠，默认展开 |
| 项目经历模块 | `/workspace/web/src/app/components/resume-editor/modules/ProjectSection.tsx` | 可折叠，默认展开 |
| 技能信息模块 | `/workspace/web/src/app/components/resume-editor/modules/SkillsSection.tsx` | 可折叠，默认展开 |
| 实习经历模块 | `/workspace/web/src/app/components/resume-editor/modules/InternshipSection.tsx` | 可折叠，默认折叠 |
| 证书与获奖模块 | `/workspace/web/src/app/components/resume-editor/modules/CertificatesSection.tsx` | 可折叠，默认折叠 |
| 作品链接模块 | `/workspace/web/src/app/components/resume-editor/modules/PortfolioExtraSection.tsx` | 可折叠，默认折叠 |
| 年月选择器 | `/workspace/web/src/app/components/resume-editor/YearMonthPicker.tsx` | 自定义，button+popover，YYYY-MM 格式 |
| Textarea UI | `/workspace/web/src/app/components/ui/textarea.tsx` | 基础 textarea wrapper，默认 resize-y / min-h-[120px] |
| 保存卡片 | `/workspace/web/src/app/components/resume-editor/ResumeEditorSaveCard.tsx` | 非固定，随内容流 |
| 右侧状态栏 | `/workspace/web/src/app/components/resume-editor/ResumeEditorStatusSidebar.tsx` | **未渲染** |
| 名称检查前端 API | `/workspace/web/src/api/resumes.ts` | checkResumeName(name, excludeId?) |
| 名称检查后端 Handler | `/workspace/server/internal/http/resume_handler.go` | GET /api/resumes/check-name |
| 名称检查后端 Service | `/workspace/server/internal/service/resume_service.go` | DB UNIQUE(user_id,name) 检查 |

---

## 三、简历名称可用状态诊断

### 3.1 名称检查机制

完整链路：**前端 onBlur → 调用后端 API → 后端 DB 查询 → 返回结果**。

```
用户输入 → onBlur 触发 → checkResumeName(name, excludeId?) 
→ GET /api/resumes/check-name?name=X&exclude_id=Y 
→ service.CheckResumeName() → DB ExistsByName(user_id, name, exclude_id) 
→ 返回 {available: bool, reason: string}
```

- **检查方式**：后端 API，非前端本地判断
- **Debounce**：无。只在 blur 时触发一次
- **编辑模式保护**：通过 `exclude_id` 参数排除当前简历，编辑自己的简历时不会误判为重复
- **DB 层**：UNIQUE(user_id, name) 约束 + repository.ExistsByName() 查询

### 3.2 状态枚举

| 状态 | 触发条件 | 当前文案 | 当前样式 |
| -- | -- | -- | -- |
| `idle` | 初始状态 / 名称被修改后重置 | "输入简历名称后自动检查是否可用" | `bg-slate-50 border-slate-100 text-slate-500`（灰色，不明显） |
| `checking` | onBlur 后 API 调用中 | "正在检查…" | `bg-blue-50 border-blue-100 text-blue-600` |
| `available` | API 返回 available=true | "名称可用，保存后将用于岗位推荐、匹配分析和简历优化建议。" | `bg-green-50 border-green-100 text-green-700` |
| `duplicate` | API 返回 available=false | reason 文本或 "该简历名称已存在" | `bg-red-50 border-red-200 text-red-600` |
| `invalid` | 本地 trim 后为空 | "简历名称不能为空" | `bg-red-50 border-red-200 text-red-600` |
| `error` | API 调用失败 | "暂时无法校验名称，请稍后重试。" | `bg-red-50 border-red-100 text-red-500` |

### 3.3 问题诊断

| 状态 | 当前文案 | 当前样式 | 问题 | 建议 |
| -- | -- | -- | -- | -- |
| idle | "输入简历名称后自动检查是否可用" | 灰色背景，无图标区分 | 信息提示不醒目，灰色容易被忽略 | 改为蓝色提示，增加问号图标 |
| available | "名称可用…"（长文案） | 绿色背景 Sparkles 图标 | ✅ 可用状态表现尚可，但文案过长 | 精简为"名称可用 ✓"，保留小字提示 |
| duplicate | reason 文本 | 红色背景 | ✅ 正确 | 无需修改 |
| invalid | "简历名称不能为空" | 红色背景 | ✅ 正确 | 无需修改 |
| checking | "正在检查…" | 蓝色背景 | 无 spinner 动画 | 加 loading spinner |
| error | "暂时无法校验名称…" | 红色背景 | ✅ 正确 | 无需修改 |

**核心问题**：`idle` 状态下的提示信息视觉强度太弱（灰色），用户可能不知道需要进行名称检查。截图中的"可用不明显"可能是因为用户处于 idle 状态，还没有触发 blur。

### 3.4 误判风险评估

- **编辑当前简历时不会误判**：`exclude_id` 机制正确排除当前简历 ID
- **没有 debounce 保护**：用户快速修改名称后立即点击保存，可能在 save 逻辑中再次触发 check → 已有保存前的二次校验（第 237-256 行）

---

## 四、Textarea 内部滚动诊断

### 4.1 当前 Textarea 使用情况汇总

所有 textarea 均使用基础 `Textarea` UI 组件，并在调用处覆盖 `className` + `rows` + `resize-none`。

| 字段 | 位置（行号） | 当前组件 | min-h | rows | resize | 是否内部滚动 | 建议 AutoGrow |
| -- | -- | -- | -- | -- | -- | -- | -- |
| 论文 / 学术成果 | L403 | `<Textarea>` | 72px | 2 | resize-none | 是（72px 太小） | ✅ 强烈建议 |
| 项目描述 | L419 | `<Textarea>` | 120px | 4 | resize-none | 是（内容超 4 行时） | ✅ 强烈建议 |
| 实习描述 | L441 | `<Textarea>` | 120px | 4 | resize-none | 是（内容超 4 行时） | ✅ 强烈建议 |
| AI 应用技能 | L429 | `<Textarea>` | 88px | 3 | resize-none | 是（内容超 3 行时） | ✅ 建议（但非高频长文本） |
| 证书说明 | L458 | `<Textarea>` | 72px | 2 | resize-none | 是（72px 太小） | ⚠️ 中等（通常较短） |
| 获奖说明 | L460 | `<Textarea>` | 72px | 2 | resize-none | 是（72px 太小） | ⚠️ 中等（通常较短） |
| 补充信息 | L474 | `<Textarea>` | 140px | 未设 | 未设 resize-none（继承 resize-y） | 是（但用户可拖拽） | ✅ 强烈建议 |

### 4.2 根因分析

1. **所有长文本字段均设置 `resize-none`**：禁止用户手动拖拽调整高度
2. **固定 rows + min-h**：rows=2~4 对应约 48~96px 的文字高度，当内容超过该高度后 textarea 内部出现滚动条
3. **Textarea 基础组件默认 `resize-y`**：但调用处全部覆盖为 `resize-none`
4. **不存在 AutoGrowTextarea 组件**：需要新增

### 4.3 适合封装 AutoGrowTextarea

| 评估项 | 结论 |
| -- | -- |
| 前端是否方便实现 | ✅ 是，通过 ref 监听 scrollHeight 即可 |
| 后端是否需改动 | ❌ 不需要，纯前端 |
| 数据库是否需改动 | ❌ 不需要 |
| 是否需要依赖 | ❌ 不需要新依赖，原生的 textarea + ref |
| 是否影响性能 | ⚠️ 极轻微（scrollHeight 读取），可加 throttle |

### 4.4 推荐策略

- **封装 `AutoGrowTextarea` 组件**：基于现有 `Textarea`，增加 `maxRows` 属性控制最大行数
- **项目描述 + 实习描述 + 论文摘要 + 补充信息**：全部使用 AutoGrowTextarea，maxRows=12
- **证书说明 + 获奖说明 + AI 应用技能**：保持当前固定高度，这些字段通常不会太长
- **保留 `resize-y` 作为兜底**：给用户额外的控制权

---

## 五、年月选择器尺寸跳动诊断

### 5.1 当前组件结构

`YearMonthPicker` 是一个自定义 button-trigger 组件：

```tsx
// Trigger button
<button className="flex h-10 w-full items-center rounded-xl ...">
  <span className="flex-1 truncate">
    {display || placeholder}  // display="2026.06" 或 placeholder="开始时间"
  </span>
  <svg>...</svg>               // 下拉箭头
</button>
```

### 5.2 使用位置

`YearMonthPicker` 出现在以下 3 处（均在 MyResumesPage.tsx 内）：

| 场景 | 位置 | placeholder | 实际值格式 |
| -- | -- | -- | -- |
| 教育经历 - 开始/结束时间 | L387 | "开始时间" / "结束时间" | "2026.06" |
| 项目经历 - 开始/结束时间 | L417 | "开始时间" / "结束时间" | "2026.06" |
| 实习经历 - 开始/结束时间 | L440 | "开始时间" / "结束时间" | "2026.06" |

### 5.3 根因分析

| 组件 | 当前宽高策略 | 跳动原因 | 建议固定尺寸 |
| -- | -- | -- | -- |
| YearMonthPicker trigger | `w-full`（flex-1 自适应）+ `h-10` | placeholder "开始时间"（4 字 28px）→ 实际值 "2026.06"（7 字 ~56px），内容宽度变化导致按钮在 flex 容器中的占比变化，进而使两个 picker 之间的间距和相对位置变动 | 固定 `w-[136px]` |

**根因确认**：
- Trigger 使用 `w-full`，在 flex 容器中自适应
- Placeholder "开始时间"（4 中文字符）约 56px 宽
- 实际值 "2026.06"（7 个 ASCII 字符）约 56px 宽
- **问题不在于 placeholder 和值之间的宽度差异（它们恰好宽度接近），而在于 `w-full` 在 flex 容器中导致两个 picker 之间的 `gap-2` 分配比例变化** — 当值非空时两个 picker 内容宽度相等，但当任一为空时 placeholder 宽度可能导致轻微偏移
- 更准确的根因：两个 `YearMonthPicker` 放在 `flex items-center gap-2` 的容器中，每个 `w-full` 尝试平分剩余空间，但 placeholder "开始时间" / "结束时间" 和实际值 "2026.06" / "2025.12" 的渲染宽度可能存在 ±2px 差异，导致轻微跳动

### 5.4 修复建议

- 给 `YearMonthPicker` **添加固定的 `min-width`**（如 `w-[136px]`），保证两个 picker 宽度始终一致
- 或者给外层 flex 容器改为 `grid grid-cols-[1fr_auto_1fr]` 布局，消除 flex-grow 比例分配的不确定性
- 所有 3 处使用 `YearMonthPicker` 的位置（教育、项目、实习）统一修复

---

## 六、模块展开/折叠动画诊断

### 6.1 当前折叠机制

`ResumeEditorSectionCard` 使用 **条件渲染** (`{!isCollapsed && (...)}`) 而非 CSS 动画：

```tsx
// L135-142
{!isCollapsed && (
  <div className="px-6 pb-6">
    {description && <p className="...">{description}</p>}
    <div className={cn(!description && "pt-2")}>{children}</div>
  </div>
)}
```

### 6.2 展开/折叠时 header 下内容变化

| 状态 | Header 下方显示内容 | 结构 |
| -- | -- | -- |
| 折叠 | description 文本（line-clamp-1），在独立 div 中，px-6 pb-1 | `<div className="px-6 pb-1"><p>desc</p></div>` |
| 展开 | body div（px-6 pb-6），内含 description（-mt-2） + children | `<div className="px-6 pb-6"><p className="-mt-2">desc</p><div>children</div></div>` |

### 6.3 问题诊断

| 模块 | 当前折叠方式 | 问题 | 建议动画方案 |
| -- | -- | -- | -- |
| 教育经历 | 条件渲染 + ChevronDown 旋转 | 折叠时 description 显示在 header 下方（pb-1），展开时 description 移动到 body 内部（pb-6），视觉上副标题位置发生 **±5px 垂直位移** | height 动画 + grid-template-rows |
| 项目经历 | 同上 | 同上 | 同上 |
| 技能信息 | 同上 | 同上 | 同上 |
| 实习经历 | 同上 | 同上 | 同上 |
| 证书与获奖 | 同上 | 同上 | 同上 |
| 作品链接 | 同上 | 同上 | 同上 |

### 6.4 副标题位移根因

**原因1 (核心)**：折叠时 description 在 `px-6 pb-1` 的独立容器中（下 padding 4px），展开时 description 在 `px-6 pb-6` 的容器中（下 padding 24px）且带 `-mt-2`（负 margin 8px），两个容器的视觉起始位置有细微差异。

**原因2**：折叠时 description 使用 `line-clamp-1`（单行截断），展开时 description 可能换行，行数变化导致视觉差异。

### 6.5 当前动画能力

| 动画项 | 是否实现 | 说明 |
| -- | -- | -- |
| ChevronDown 旋转 | ✅ | `transition-transform duration-200` + `rotate-180` |
| Body 展开动画 | ❌ | 条件渲染，无过渡 |
| Description 切换动画 | ❌ | display block/none 切换 |

### 6.6 推荐方案

- **统一 `CollapsibleSection` 动画**：将条件渲染改为 `grid-template-rows: 0fr → 1fr` 过渡（Tailwind v4 支持）
- **Description 始终在 body 内部**：折叠时 hidden 而不是移动位置
- **当前动画已具备**：Chevron 旋转 animation 已经正确（duration-200）
- **Body 展开建议**：使用 CSS `grid-template-rows` transition（或 max-height transition），200ms ease-out
- **不引入新依赖**：Tailwind v4 已内置 `transition-all`

---

## 七、"一键体验创建新简历"可行性

### 7.1 现有示例数据

| 数据源 | 说明 |
| -- | -- |
| `init.sql` 中 9 份示范简历 | user_id=2（Arthur），id=201~209，完整关系型数据 |
| API 获取方式 | `getResumeById(201~209)` |
| 数据质量 | 高质量，已通过多轮验收 |

### 7.2 方案评估

| 方案 | 优点 | 风险 | 是否推荐 |
| -- | -- | -- | -- |
| **A：填入表单不自动保存** | 用户可预览体验内容，决定是否保存；不污染数据库 | 表单可能需要 adapter 转换 | ✅ 强烈推荐 |
| B：自动创建新简历并跳转编辑页 | 流程简单，一键完成 | 会污染当前用户数据库；可能与现有名称冲突 | ❌ 不推荐 |
| C：直接跳转到示范账号的简历 | 展示最完整数据 | 破坏当前登录态 | ❌ 不推荐 |
| D：在简历管理页增加"创建示例简历"按钮 | 不影响编辑页 UI | 用户需要多一步操作 | ⚠️ 备选 |

### 7.3 推荐实现细节

- **按钮位置**：新建简历页（/resumes/new）顶部说明栏中，紧接"当前简历可用于：岗位推荐/匹配分析/优化"提示
- **按钮文案**："✨ 填入体验内容"
- **行为**：调用 API 获取 demo 简历数据 → 通过 adapter 转换为 ViewModel → 填入所有表单 state → **不自动保存**
- **数据来源**：读取一份示范简历（如 id=201），通过 `getResumeById(201)` 获取
- **不污染数据库**：仅前端表单填入，不调用 create/update API

---

## 八、"投递准备度 / 简历完善度"可行性

### 8.1 当前状态

- `ResumeEditorStatusSidebar` 组件已存在（186 行），包含 completeness 进度条、aiReady 状态、缺失必填/推荐补充计数、四核心功能状态
- **但该组件当前不在编辑页渲染**（K8-5-29-E-R 移除）
- 原组件计算的 `completeness` 通过 props 传入，非实时计算

### 8.2 推荐计算方案

按分类权重计算，总分 100%：

| 分类 | 字段 | 权重 | 计入条件 |
| -- | -- | -- | -- |
| **基础必需项** | 简历名称 | 5% | 非空 |
| | 姓名 | 10% | 非空 |
| | 当前城市 | 5% | 非空 |
| | 手机号 | 2.5% | 非空 |
| | 邮箱 | 2.5% | 非空 |
| **技能信息** | 专业技能 / 工具能力 | 15% | 非空 |
| | AI 应用技能 | 5% | 非空（推荐项，非必需） |
| **教育经历** | 至少 1 条教育记录（学校+专业） | 15% | school 和 major 非空 |
| | 研究方向 / 实验室 / 论文 | 5% | 任意一项非空 |
| **项目/实习/调研** | 至少 1 条项目/实习记录 | 15% | 名称非空 |
| **证书与获奖** | 至少 1 条证书/获奖/外语考试 | 5% | 名称非空 |
| **作品链接** | 至少 1 条作品链接或主页 | 5% | url 或 homepage 非空 |
| **补充信息** | 补充信息 | 5% | 非空 |
| **求职意向** | 求职意向（方向/城市） | 5% | 非空 |

### 8.3 设计原则

- **不要叫"编写进度"**，建议叫 **"投递准备度"** 或 **"简历完善度"**
- 分三档显示：
  - 基础必需项（名称+姓名+教育+技能+项目/实习）→ 占 67.5%
  - 类型适用项（研究方向/论文 — 研究型必填，本科可选）→ 占 5%
  - 增强建议项（城市/联系方式/证书/作品/补充信息/求职意向）→ 占 27.5%
- **本科简历不填研究方向/导师/实验室/论文仍可 100%**：这 5% 权重不计入本科生的总分
- **研究型简历**：研究方向/论文计入核心项
- **项目/实习/调研**：至少存在一类（15% 权重）
- **增强项**：计入总进度但是可选项（用户可选择忽略）
- **实时计算**：基于表单 state 纯前端计算，无需后端

### 8.4 可行性结论

| 项 | 结论 |
| -- | -- |
| 当前页面有右侧栏 | ❌ 无（组件存在但已移除） |
| 右侧栏组件设计就绪 | ✅ `ResumeEditorStatusSidebar` 已设计（186 行） |
| 表单数据支持实时计算 | ✅ 全部表单 state 在前端可用 |
| 是否需要后端 API | ❌ 不需要，纯前端 |
| 是否适合下一阶段实现 | ✅ 是，配合右侧栏恢复 |

---

## 九、页面信息架构诊断

### 9.1 当前编辑页结构

| 区域 | 当前内容 | 问题 | 建议 |
| -- | -- | -- | -- |
| 顶部说明区 | 标题 + 副标题 + "返回"链接 + 蓝色提示条 | ✅ 良好 | 可增加"填入体验内容"按钮 |
| 简历基础 | 简历名称 input + 状态提示 | 名称状态 idle 时太弱 | 强化 idle 状态视觉 |
| 基础信息 | 姓名* / 城市 / 手机号 / 邮箱 | ✅ 良好 | 无需修改 |
| 教育经历 | 学校* / 专业* / 学历 / 院系 / 时间 / GPA / 排名 / 主修课程 / 方向 / 导师 / 实验室 / 论文 | ✅ 良好 | textarea 需要 AutoGrow |
| 实习经历 | 公司* / 职位* / 时间 / 描述 | ✅ 良好 | textarea 需要 AutoGrow |
| 项目经历 | 名称* / 时间 / 角色 / 描述 | ✅ 良好 | textarea 需要 AutoGrow |
| 技能信息 | 专业技能 / AI 应用技能 | ✅ 良好 | --- |
| 证书与获奖 | 外语考试 / 证书 / 获奖 | ✅ 良好 | --- |
| 作品链接与补充信息 | 作品链接 / 个人主页 / 补充信息 | ✅ 良好 | textarea 需要 AutoGrow |
| 右侧栏 | **无（未渲染）** | ❌ 缺失 | 需要恢复并升级 |
| 底部保存区 | `ResumeEditorSaveCard`（非固定） | ⚠️ 不在 viewport 底部固定，长表单需滚动到底才能保存 | 可考虑 sticky bottom |

### 9.2 布局评估

| 维度 | 当前状态 | 问题 | 建议 |
| -- | -- | -- | -- |
| 最大宽度 | `max-w-[1280px]` | ✅ 合理 | 不修改 |
| 双栏布局 | 已支持（`ResumeEditorShell` 主栏 + 280px 侧栏） | 当前侧栏未启用 | 启用右侧"投递准备度"栏 |
| 移动端 | 已有响应式（lg:flex-row） | ✅ 良好 | 侧栏在移动端置于顶部 |
| 保存按钮 | 非固定，位于内容流底部 | 可能被用户忽略 | 保持当前设计（非 fixed），保存卡片有足够视觉体量 |
| 品牌一致性 | 使用 Offerr鹅 蓝 `#0052D9` + 白卡 `bg-white ring-1 ring-gray-100` | ✅ 一致 | 无需修改 |

---

## 十、全链路重构判断

| 层级 | 是否需要修改 | 理由 |
| -- | -- | -- |
| 后端 API | ❌ 不需要 | 名称检查 API 已完善；无需新增接口 |
| 数据库结构 | ❌ 不需要 | 无需新增字段 |
| 前端页面（MyResumesPage） | ✅ 需要 | Textarea AutoGrow、名称状态优化、填入体验内容按钮 |
| 通用 Textarea 组件 | ✅ 需要 | 新增 `AutoGrowTextarea` 组件 |
| 年月选择器组件 | ✅ 需要 | 固定宽度修复 |
| 折叠组件（SectionCard） | ✅ 需要 | 展开动画修复 |
| 示例数据模板 | ❌ 不需要 | 直接用现有 API 读取示范简历 |
| 右侧准备度栏 | ✅ 需要 | 恢复 `ResumeEditorStatusSidebar` 并升级为投递准备度 |

**结论：下一阶段只改前端**，后端和数据库零改动。

---

## 十一、推荐执行方案

### 11.1 最小修复方案（P0）

修复 4 个截图中的体验问题：

| 项目 | 工作量 | 风险 |
| -- | -- | -- |
| 名称可用状态视觉强化（idle 改为蓝色提示 + spinner） | 小 | 低 |
| 新增 AutoGrowTextarea 组件 + 接入 4 个长文本字段 | 中 | 低 |
| YearMonthPicker 固定宽度（w-[136px]） | 小 | 低 |
| SectionCard 折叠动画改为 grid-template-rows | 中 | 低 |

### 11.2 产品化方案（P1，推荐）

在最小修复基础上增加：

| 项目 | 工作量 | 风险 |
| -- | -- | -- |
| "✨ 填入体验内容"按钮（新建页顶部） | 中 | 中 |
| 恢复并升级右侧"投递准备度"栏 | 中 | 中 |
| 投递准备度实时计算逻辑 | 中 | 低 |

### 11.3 不建议现在做

| 项目 | 理由 |
| -- | -- |
| 改数据库结构 | 不需要 |
| 接 LLM 自动写简历 | 当前阶段不应接入生成式功能 |
| 自动保存大量示例简历到 DB | 避免污染当前数据库 |
| 移动端深度优化 | 当前优先桌面端体验 |

### 11.4 推荐方案

**推荐产品化方案**，理由：
1. 右侧栏组件已设计完成（186 行），只需恢复并接入
2. 投递准备度纯前端计算，无需后端支持
3. 填入体验内容可利用已有 9 份示范数据，无需额外数据
4. 两个增强项一起做避免重复修改 MyResumesPage

---

## 十二、未修改范围确认

- **未修改任何源码**
- **未修改数据库**
- **未修改 server/sql/init.sql**
- **未执行 reset-up**
- **未执行任何 Git 命令**
- **未修改 .env / .env.example**
- **未修改任何简历数据**

---

## 十三、下一阶段建议

### K9-UI-ResumeEdit-B：简历编辑页表单体验与投递准备度重构

**阶段 B 应执行的内容**：

1. **名称可用状态优化**（小，P0）
   - idle 状态从灰色改为蓝色提示
   - checking 状态加 loading spinner
   - available 文案精简

2. **AutoGrowTextarea 封装与接入**（中，P0）
   - 新建 `web/src/app/components/ui/auto-grow-textarea.tsx`
   - 论文摘要、项目描述、实习描述、补充信息接入
   - 保留 maxRows 限制防止无限增高

3. **YearMonthPicker 宽度修复**（小，P0）
   - 添加 `w-[136px]` 固定宽度
   - 3 处引用统一修复

4. **SectionCard 折叠动画**（中，P0）
   - 将条件渲染改为 CSS grid-template-rows 过渡
   - Description 始终在 body 内，折叠时隐藏

5. **填入体验内容按钮**（中，P1）
   - 新建简历页顶部加按钮
   - 调用 getResumeById 读取示范数据
   - Adapter 转换填入表单 state

6. **投递准备度右侧栏**（中，P1）
   - 恢复 ResumeEditorShell sidebar prop
   - 实现 preparenessScore 纯前端计算
   - 显示：进度环 + 分类面完成情况 + 不足提示

---

## 十四、docs 更新清单

| 文件 | 操作 | 说明 |
| -- | -- | -- |
| `docs/09-开发日志/2026-06-09-665-K9-UI-ResumeEdit-A-简历编辑页体验问题只读诊断.md` | 新增 | 本诊断报告 |
| `docs/下一步任务.md` | 更新 | 新增 K9-UI-ResumeEdit-A 条目 |
| `docs/变更记录.md` | 更新 | 新增诊断记录 |

---

## 十五、Git 约束确认

- [x] 未执行 `git status`
- [x] 未执行 `git diff`
- [x] 未执行 `git add`
- [x] 未执行 `git commit`
- [x] 未执行 `git push`
- [x] 未执行 `git reset`
- [x] 未执行 `git checkout`
- [x] 未执行 `git branch`
- [x] 未执行 `git tag`
- [x] 未执行 `git show`
- [x] 未执行 `git restore`
- [x] 未执行 `make reset-up`
- [x] 未执行 `make reset`
- [x] 未执行 `docker compose down -v`
- [x] 未执行 DROP/DELETE/TRUNCATE 任何数据库操作
