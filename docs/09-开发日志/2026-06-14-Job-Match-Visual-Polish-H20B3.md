# H20-B-3：Job Match Visual Polish & Screenshot Acceptance

> 日期：2026-06-14
> 类型：前端视觉精修
> 编号：H20-B-3

## 目标

在 H20-B-2-R 功能重构完成后，对页面做视觉精修，使其达到成熟产品观感。

## 审计发现（H20-B-2-R 遗留视觉问题）

| 问题 | 发现 |
|---|---|
| Hero 分数过小 | `text-[32px]`，不够醒目 |
| 脆弱的字符串拼接 | `readiness.bgColor.replace("bg-","bg-").replace("-50","-100")` 依赖字符串替换 |
| Pass Points 全量展示 | 主视图未限制 Top 3，所有命中证据默认铺开 |
| Risks 全量展示 | 主视图未限制 Top 3，所有风险默认铺开 |
| MetaActionBar 小屏拥挤 | 无 `flex-wrap`，长内容挤压按钮 |
| AI Disclaimer 色调过蓝 | `bg-blue-50/50` + `border-blue-100` 喧宾夺主 |
| Hero 卡片过高 | `p-8` + `gap-6` 压迫感强 |
| 缺少面试准备折叠区 | 无"面试准备参考"模块 |

## 精修内容

### 1. Hero 分数放大
- `text-[32px]` → `text-5xl`（48px）
- 圆圈 `w-[80px] h-[80px]` → `w-[96px] h-[96px]`
- 卡片 `p-8` → `p-6`，`gap-6` → `gap-5`，减少视觉压迫
- AI 轻提示 `mt-5 pt-4` → `mt-4 pt-3`，`text-xs` → `text-[11px]`

### 2. ReadinessLevel 类型重构
- 移除 `color`/`bgColor` 字符串拼接
- 新增 `cardBg`、`labelBg`、`iconColor`、`textColor` 独立属性
- 所有颜色值预计算，零运行时字符串替换

### 3. 主视图 Top-N 控制
- Pass Points：主视图 `.slice(0, 3)`，超量项放入「完整命中证据」折叠区
- Elimination Risks：主视图 `.slice(0, 3)`，超量项放入「全部风险详情」折叠区
- 正文完整展示，不限条内截断

### 4. 新增折叠区
- 「完整命中证据」：Top 3 外的所有 pass points
- 「全部风险详情」：Top 3 外的所有 risks
- 「面试准备参考」：placeholder（后续版本填充）

### 5. MetaActionBar 响应式
- 添加 `flex-wrap`
- 按钮文字简化：`生成初筛命中率提升方案` → `初筛提升方案`，`重新生成分析` → `重新生成`
- 简历名/岗位名添加 `truncate max-w-[180px/140px]` 防止溢出
- 分隔线 `hidden sm:block` 小屏隐藏
- 图标缩小 0.5px

### 6. AI Disclaimer 色调
- `bg-blue-50/50 border-blue-100` → `bg-slate-50 border-slate-200`
- 图标 `bg-blue-100 text-blue-500` → `bg-slate-200 text-slate-400`
- 标题 `text-sm text-slate-800` → `text-xs text-slate-500`
- 正文 `text-sm leading-7 text-slate-600` → `text-xs leading-6 text-slate-500`

## 构建结果

| 检查 | 结果 |
|---|---|
| npm build | ✅ 1.86s |
| go build | ✅ PASS |
| go vet | ✅ PASS |
| go test | ✅ PASS |

## 静态扫描

| 扫描项 | 结果 |
|---|---|
| Hero 字号 | ✅ `text-5xl` |
| 省略/截断 | ✅ 仅 MetaActionBar 短标签（允许单行） |
| 右侧栏 `<aside>` | ✅ 无 |
| 脆弱的字符串拼接 | ✅ 已移除 |
| raw code | ✅ CLEAN |
