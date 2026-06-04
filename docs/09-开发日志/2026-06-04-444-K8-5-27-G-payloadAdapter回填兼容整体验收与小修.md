# K8-5-27-G：payload / adapter / 回填兼容整体验收与小修

## 本阶段目标

对 K8-5-27 系列（B→F）产生的 payload/adapter/回填变化做全链路一致性复核。

## payload 复核结果

| 字段 | 当前结构 | 结果 |
|------|---------|:---:|
| `schema_version` | `2` | ✅ |
| `target_direction` | `""` | ✅ |
| `target_role` | `""` | ✅ |
| `job_intention` | 不进入 payload | ✅ |
| `skills` | `{development_languages, ai_application}`（无 language_exam） | ✅ |
| `certificates[0]` | `type:"language_exam"` 外语考试前置 | ✅ |
| `certificates[1..]` | 普通证书（type≠language_exam） | ✅ |
| `awards` | 5 字段不变 | ✅ |
| `portfolio_links` | `{works, homepage}` | ✅ |
| `extra_info` | `{self_evaluation, additional_info}` | ✅ |
| `achievements` | `string[].filter(Boolean)` | ✅ |
| `start_date/end_date/issued_at/awarded_at` | YYYY-MM 字符串不变 | ✅ |

## adapter / 回填兼容结果

| 结构 | 结果 |
|------|:---:|
| V1 `skills[]` → `developmentLanguages` | ✅ `skills.filter(Boolean)` |
| V2 `skills.programming_languages` → `developmentLanguages` | ✅ L435 |
| V2 `skills.other_keywords` → `developmentLanguages` fallback | ✅ L438 |
| V2 `skills.ai_tools + ai_application_scenarios` → `aiApplication` | ✅ L439-440 |
| 旧 `skills.language_exam` → `languageExamType/Score` | ✅ L444-445 |
| 新 `certificates type=language_exam` → `languageExamType/Score` | ✅ L453-458 |
| 旧 portfolio `github/gitee/cnb` → works | ✅ L502-504 |
| 旧 portfolio `demo_links/portfolio_links` → works | ✅ L505-506 |
| 旧 portfolio `blog/homepage` → homepage | ✅ L515 |
| 旧 `achievements` string → `toAchievementList()` | ✅ L781 |
| 新 `achievements` string[] → 直接回填 | ✅ |
| certificates/awards → 对应 state | ✅ |
| extra_info.self_evaluation/additional_info → 自评模块 | ✅ |

## YearMonthPicker 复核

| 检查项 | 结果 |
|--------|:---:|
| YearMonthPicker 8 处接入 | ✅ L1074/1076/1189/1191/1312/1314/1408/1426 |
| `type="month"` 残留 | ✅ 0 |
| `type="date"` 残留 | ✅ 0 |

## 旧字段 UI 残留

| 命中 | 位置 | 类型 |
|------|:---:|:---:|
| `github` | L502-504 | 回填 helper ✅ |
| `gitee` | L503 | 回填 helper ✅ |
| `cnb` | L504 | 回填 helper ✅ |
| `开发语言` | L1261 | helper text（非 UI label）✅ |
| `github.com/xxx` | L1244/1460 | placeholder ✅ |
| `cnb.cool/xxx` | L1455 | placeholder ✅ |

**0 个旧字段 UI 输入残留 ✅**

## 小修清单

**无源码修改** — K8-5-27-B~F 产生的所有改动均为设计预期，payload/adapter/回填链路一致。

## 上传能力禁止确认

| 检查项 | 结果 |
|--------|:---:|
| `type="file"` / `FormData` / `upload` / `multipart` / `dropzone` / `drag` | ✅ 0 |

## 构建 / API / 测试结果

| 验收项 | 结果 |
|--------|:---:|
| `npm run build` | ✅ 2.04s |
| `go build ./...` | ✅ |
| `go vet ./...` | ✅ |
| `go test ./...` | ✅ 全部通过 |
| API 200/200/200/401 | ✅ |
| 后端无误改 | ✅（仅 ResumeSnapshotV2.AIApplicationScenarios 既有字段） |

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
```

## 风险与边界

```text
真实登录态端到端回归 → K8-5-27-Z
Agent 再设计 → K8-5-26-A
```

## 未执行 Git 命令

本次未执行任何 Git 命令，也未提交代码。
