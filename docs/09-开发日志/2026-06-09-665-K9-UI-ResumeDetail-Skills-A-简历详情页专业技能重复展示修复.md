# K9-UI-ResumeDetail-Skills-A：简历详情页专业技能重复展示修复

- **日期**: 2026-06-09
- **状态**: ✅ 已完成
- **类型**: UI 修复

## 问题

简历详情页「能力摘要」模块中，`专业技能 / 工具能力` 同时展示了两份内容：
1. 上方：一整段逗号分隔的长文本技能列表
2. 下方：一组蓝色技能标签

造成明显重复，视觉冗余。

## 根因

`ResumeDetailPage.tsx` 第 268-278 行将 `resume.professional_skills` 字段同时渲染为原始文本和标签：

```tsx
<div className="text-sm text-slate-700 leading-relaxed">{resume.professional_skills}</div>
<TagList items={skillTags} />
```

`skillTags` 是由同一个 `professional_skills` 字段按分隔符拆分而来，因此两条渲染链路表达了完全相同的信息。

## 修复方案

将同时渲染改为互斥选择：
- 有技能标签时（`skillTags.length > 0`）：**只显示标签**
- 无标签时：fallback 显示原始文本（罕见的边缘情况）
- 不再同时渲染两者

```tsx
{skillTags.length > 0 ? (
  <TagList items={skillTags} />
) : (
  <div className="text-sm text-slate-700 leading-relaxed">{resume.professional_skills}</div>
)}
```

## 修改文件

| 文件 | 修改 |
|------|------|
| `web/src/app/pages/ResumeDetailPage.tsx` | 第 272-281 行：移除重复文本渲染，改用互斥标签/文本展示 |

## 未修改范围

- 后端 API：未修改
- 数据库结构：未修改
- `server/sql/init.sql`：未修改
- 真实简历数据：未修改
- 简历编辑页保存逻辑：未影响
- `AI 应用技能` 和 `补充信息` 段落展示：未修改

## 构建验证

| 验收项 | 结果 |
|--------|:--:|
| npm build | ✅ |
| go build | ✅ |
| go vet | ✅ |
| check-encoding | ⚠️ 失败项为 pre-existing（init.sql mysqldump 转义误报） |

## 验收标准

| 项 | 结果 |
|----|:--:|
| 专业技能 / 工具能力 只显示标签 | ✅ |
| 上方逗号分隔长文本不再显示 | ✅ |
| AI 应用技能 仍显示段落 | ✅ |
| 补充信息 仍显示段落 | ✅ |
| 标签换行正常 | ✅ |
| 页面无空白异常 | ✅ |
| 无新增 undefined/Invalid Date 风险 | ✅ |
| 无新增 Console 错误 | ✅ |
