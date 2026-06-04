# K8-5-22-C — NormalizeResumeToSnapshotV2 后端归一化层

> 日期：2026-06-04 | 性质：implementation | 新增 model + service + test

---

## 1. 本阶段目标

基于 K8-5-22-A 的数据库字段和 K8-5-22-B 的后端 V2 支持，新增 NormalizeResumeToSnapshotV2 归一化层，把 V1/V2 ResumeVersion 统一转换成 ResumeSnapshotV2，供后续 Agent 接入使用。

---

## 2. 新增文件清单

| 文件 | 修改内容 | 属于 K8-5-22-C |
|------|---------|:---:|
| `server/internal/model/resume_snapshot.go` | **新增** — ResumeSnapshotV2 + 14 个子类型（Meta/Profile/JobIntention/Education/Internship/Project/Skills/Certificate/Award/Portfolio/ExtraInfo） | ✅ |
| `server/internal/service/resume_snapshot_normalizer.go` | **新增** — NormalizeResumeToSnapshotV2 + normalizeV1ToSnapshot + normalizeV2ToSnapshot + 6 个 JSON 解析/转换 helper | ✅ |
| `server/internal/service/resume_snapshot_normalizer_test.go` | **新增** — 12 项测试：V1 最小/V2 完整/V2 fallback/空JSON/非法JSON/schemaVersion=0/空V2模块/V1字符串技能/V2对象技能/null JSON/phone email 不入 Snapshot | ✅ |

---

## 3. ResumeSnapshotV2 结构确认

| 模块 | 字段摘要 | 用途 |
|------|---------|------|
| Meta | schema_version, resume_version_no, resume_name, target_direction | 元信息 |
| Profile | name, school, degree, major, graduation_year, current_city | 基础信息（不含 phone/email） |
| JobIntention | target_roles, job_type, expected_cities, accept_other_cities, earliest_start_date, internship_duration, weekly_available_days, target_industries, target_companies | 求职意向 |
| Educations | []{school, degree, major, department, start_date, end_date, gpa, gpa_scale, rank, core_courses, research_direction, lab, paper_summary} | 教育经历 |
| Internships | []{company, role, start_date, end_date, work_content, tech_stack, achievements, metrics} | 实习经历 |
| Projects | []{name, role, type, start_date, end_date, tech_stack, description, responsibilities, achievements, link} | 项目经历 |
| Skills | {programming_languages, backend_frameworks, frontend_technologies, databases, middleware, cloud_platforms, dev_tools, ai_tools, ai_application_scenarios, other_keywords} | 技能信息 |
| Certificates | []{name, type, level, score, issued_at} | 证书与语言 |
| Awards | []{name, type, level, awarded_at, description} | 获奖与竞赛 |
| Portfolio | {github, gitee, cnb, blog, demo_links, portfolio_links, homepage} | 作品链接 |
| ExtraInfo | {self_evaluation, additional_info} | 补充信息 |

共 11 个模块，~90 字段。

---

## 4. V1 → V2 映射规则

| V1 字段 | V2 Snapshot 字段 | 结果 |
|---------|-----------------|:---:|
| VersionNo | meta.resume_version_no | ✅ 直接映射 |
| Name | meta.resume_name | ✅ 直接映射 |
| TargetRole | meta.target_direction + jobIntention.targetRoles[0] | ✅ 拆分为方向 + 数组 |
| TargetCity | jobIntention.expectedCities[0] | ✅ 单值拆入数组 |
| Location | profile.currentCity | ✅ 直接映射 |
| EducationJSON | educations[]（字段映射：school→school, major→major, degree→degree, start→start_date, end→end_date, courses→core_courses） | ✅ JSON 解析映射 |
| ExperienceJSON | internships[]（字段映射：company→company, role→role, start→start_date, end→end_date, description→work_content） | ✅ JSON 解析映射 |
| ProjectJSON | projects[]（字段映射：name→name, description→description, tech_stack→techStack, responsibilities→responsibilities） | ✅ JSON 解析映射 |
| SkillsJSON | skills.otherKeywords（V1 扁平数组全部归入 otherKeywords） | ✅ 字符串数组 |
| SelfEvaluation | extraInfo.selfEvaluation | ✅ 直接映射 |
| Phone/Email/ID/RawText | 不进入 Snapshot | ✅ 排除 |
| Certificates/Awards/Portfolio 等新模块 | 空数组/空默认值 | ✅ |

---

## 5. V2 → V2 映射规则

| V2 字段 | Snapshot 字段 | 策略 |
|---------|-------------|------|
| ProfileJSON | profile | JSON 优先，summary fallback |
| JobIntentionJSON | jobIntention | JSON 优先，summary fallback |
| EducationJSON | educations | JSON 直接映射 |
| InternshipJSON | internships | JSON 直接（fallback ExperienceJSON） |
| ProjectJSON | projects | JSON 直接映射 |
| SkillsJSON | skills（对象格式） | JSON 直接映射 |
| CertificatesJSON | certificates | JSON 直接映射 |
| AwardsJSON | awards | JSON 直接映射 |
| PortfolioJSON | portfolio | JSON 直接映射 |
| ExtraJSON | extraInfo | JSON 优先，self_evaluation 列 fallback |
| TargetDirection | meta.target_direction | 直接 |
| JobType | jobIntention.jobType | fallback |
| ExpectedCities | jobIntention.expectedCities | fallback（逗号分隔解析） |
| School/Degree/Major/GraduationYear | profile.* | fallback |

---

## 6. JSON 解析策略

| 策略 | 说明 |
|------|------|
| decodeJSONObject[T] | 空字符串/"null" → fallback；非法 JSON → error |
| decodeJSONArray[T] | 空字符串/"null" → []；单个对象 → 单元素数组；非法 JSON → error |
| 不 panic | 所有 JSON 解析错误通过 error 返回 |
| 不吞错误 | 非法 JSON 不静默丢弃，向外层报告 |

---

## 7. 测试覆盖确认

| 测试 | 结果 |
|------|:---:|
| V1 最小数据（目标/城市/技能/教育/实习/项目/自我评价） | ✅ PASS |
| V2 完整数据（10 模块全部 JSON） | ✅ PASS |
| V2 summary fallback（JSON 为空，摘要列兜底） | ✅ PASS |
| 空 JSON 安全（不 panic，返回空切片） | ✅ PASS |
| 非法 V1 JSON 返回 error | ✅ PASS |
| 非法 V2 JSON 返回 error | ✅ PASS |
| schema_version=0 按 V1 处理 | ✅ PASS |
| V2 空 JSON 模块安全 | ✅ PASS |
| V1 字符串数组技能 | ✅ PASS |
| V2 对象格式技能 | ✅ PASS |
| null JSON 安全 | ✅ PASS |
| phone/email/user_id/id/raw_text 不进入 Snapshot | ✅ PASS |

---

## 8. 构建 / API / 测试结果

| 验收项 | 结果 |
|--------|------|
| `go build ./...` | ✅ 通过 |
| `go vet ./...` | ✅ 通过 |
| `go test ./...`（model/service） | ✅ 全部通过（12 normalizer + 既有） |
| `npm run build` | ✅ 1.82s |
| `GET /api/jobs?page=1&page_size=5` | ✅ 200 |
| `GET /api/jobs/10001` | ✅ 200 |
| `GET /api/jobs/facets` | ✅ 200 |
| `GET /api/resumes` | ✅ 401（后端保护正常） |
| Normalizer 字段 search | ✅ 结构和方法可见 |
| Agent/Prompt/前端 未误接 | ✅ 无引用 |

---

## 9. 未修改范围确认

```text
未修改前端源码
未修改 TypeScript types
未修改 MyResumesPage
未修改 ResumeManagementPage
未修改 Adapter
未修改 Agent
未修改 Prompt / Schema
未修改 Mock
未修改数据库 SQL
未修改 API 路由
未新增接口
未修改 .env
未安装依赖
未迁移目录
未接入 Agent
```

---

## 10. 风险与边界

```text
本阶段只完成 ResumeSnapshotV2 结构和 NormalizeResumeToSnapshotV2。
Agent 接入留给 K8-5-22-I。
前端 V2 types / adapter / page 留给后续阶段。
V1/V2 全链路回归留给 K8-5-22-J。
```

---

## 11. Git 约束确认

```text
本次未执行任何 Git 命令，也未提交代码。
```
