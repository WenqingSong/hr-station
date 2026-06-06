# 2026-06-06-508-K8-5-44-B：按新契约改写 resume_versions.sql seed

> 状态：✅ 已完成 | 性质：SQL seed 重写 | 修改：1 文件

---

## 一、修改内容

重建 `server/sql/resume_versions.sql`，按 K8-5-44-A 新字段契约重写 9 份 demo 简历 seed。

## 二、字段精简结果

| 模块 | 新字段 | 删除/合并 |
|------|--------|----------|
| profile | name, phone, email, city | 删 school/degree/major/graduation_year（教育经历已有） |
| education | 全保留（13字段） | 删 related_courses |
| project | name, role, start_date, end_date, description | 删 type/tech_stack/responsibilities/achievements/link → 合并进 description |
| skills | professional_skills, ai_application | 删 11 个旧分类字段 → 合并 |
| internship | company, position, start_date, end_date, description | 删 work_content/tech_stack/achievements/metrics → 合并进 description |
| certificates | {language_exams, certificates[{name, description}]} | 删 type/level/score/issued_at → 合并进 description |
| awards | [{name, description}] | 删 type/level/awarded_at → 合并进 description |
| portfolio | {works[{name, url, password}], homepage} | 删 github/gitee/cnb/blog/demo_links/portfolio_links |
| extra | {additional_info} | 删 self_evaluation → 合并进 additional_info |

## 三、静态检查

| 检查 | 结果 |
|------|:---:|
| 旧字段残留 | 0 处 ✅ |
| additional_info | 9 份 ✅ |
| professional_skills | 9 份 ✅ |
| ai_application | 9 份 ✅ |
| 9 简历名称 | 全部存在 ✅ |
| go build | ✅ |
| go vet | ✅ |
| npm run build | ✅ |

## 四、约束确认

- 未改 init.sql
- 未改前端/后端
- 未执行 make reset-up
- 未执行 Git 命令
