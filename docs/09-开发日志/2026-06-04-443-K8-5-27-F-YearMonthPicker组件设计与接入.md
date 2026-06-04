# K8-5-27-F：YearMonthPicker 组件设计与接入

## 本阶段目标

统一年会选择器组件 YearMonthPicker，替换简历编辑页全部 8 处年月字段。不改后端/payload。

## 组件实现结果

| 项 | 结果 |
|----|:---:|
| 组件路径 | `web/src/app/components/resume-editor/YearMonthPicker.tsx` ✅ |
| Props | `{value, onChange, placeholder, disabled, className}` |
| 格式 | `YYYY-MM` ✅ |
| 面板 | 年切换 + 12 月网格 + 清除/本月 |
| 点击外部关闭 | ✅ useRef + mousedown |
| 视觉 | rounded-xl + border-slate-200 + focus:ring-[#0052D9]/10 |
| 零依赖新增 | ✅ 仅用 lucide-react (项目已有) |

## 字段替换结果（8/8）

| 模块 | 字段 | 修改前 | 修改后 |
|------|------|--------|--------|
| 教育 | startDate | `<Input>` | `<YearMonthPicker>` ✅ |
| 教育 | endDate | `<Input>` | `<YearMonthPicker>` ✅ |
| 项目 | startDate | `<Input>` | `<YearMonthPicker>` ✅ |
| 项目 | endDate | `<Input>` | `<YearMonthPicker>` ✅ |
| 实习 | startDate | `<Input>` | `<YearMonthPicker>` ✅ |
| 实习 | endDate | `<Input>` | `<YearMonthPicker>` ✅ |
| 证书 | issuedAt | `<Input type="month">` | `<YearMonthPicker>` ✅ |
| 获奖 | awardedAt | `<Input type="month">` | `<YearMonthPicker>` ✅ |

## 原生日期控件清理

| 检查项 | 结果 |
|--------|:---:|
| `type="month"` 残留 | ✅ 0 |
| `type="date"` 残留 | ✅ 0 |

## payload / 回填确认

| 项 | 结果 |
|----|:---:|
| `start_date` / `end_date` / `issued_at` / `awarded_at` 字段名不变 | ✅ |
| 格式 `YYYY-MM` 不变 | ✅ |
| `schema_version: 2` 不变 | ✅ |
| edit 回填字符串格式不变 | ✅ |

## 修改清单

| 文件 | 修改 |
|------|------|
| `YearMonthPicker.tsx` | **新增** — 121 行完整组件 |
| `MyResumesPage.tsx` L22 | 新增 import |
| `MyResumesPage.tsx` L1074-1076 | 教育 startDate/endDate → YearMonthPicker |
| `MyResumesPage.tsx` L1189-1191 | 项目 startDate/endDate → YearMonthPicker |
| `MyResumesPage.tsx` L1312-1314 | 实习 startDate/endDate → YearMonthPicker |
| `MyResumesPage.tsx` L1408 | 证书 issuedAt → YearMonthPicker |
| `MyResumesPage.tsx` L1426 | 获奖 awardedAt → YearMonthPicker |

## 构建 / API / 测试结果

| 验收项 | 结果 |
|--------|:---:|
| `npm run build` | ✅ 1.89s |
| `go build ./...` | ✅ |
| `go vet ./...` | ✅ |
| `go test ./...` | ✅ 全部通过 |
| API 200/200/200/401 | ✅ |
| 上传能力 | ✅ 0 |
| 后端无误改 | ✅ |

## 未修改范围确认

```text
未修改后端 ✅
未修改数据库 SQL ✅
未修改 Agent / Prompt / Schema / Mock ✅
未修改 API 路径 ✅
未新增 /api/resumes/v2 ✅
未修改登录鉴权 ✅
未修改 jobs 页面 ✅
未修改 .env ✅
未安装依赖 ✅
未迁移目录 ✅
未新增上传能力 ✅
未修改 payload 字段名 ✅
```

## 风险与边界

```text
payload/adapter 整体验收 → K8-5-27-G
Agent 再设计 → K8-5-26-A
```

## 未执行 Git 命令

本次未执行任何 Git 命令，也未提交代码。
