# K8-5-21-C — 数据库与后端 API 设计冻结

> 日期：2026-06-03 | 性质：design freeze | 未修改源码

---

## 1. 设计背景

基于 K8-5-21-A/B：
- 当前 `resume_versions` 单表，16 列，education/experience/project/skills 为 JSON
- 新版 10 模块 ~90 字段
- 需要平衡 Demo 快速迭代与长期可维护性

---

## 2. 当前数据库结构

| 表 | 当前字段 | 问题 |
|----|---------|------|
| `resume_versions` | id, user_id, version_no, name, phone, email, location, target_city, target_role, education_json, experience_json, project_json, skills_json, self_evaluation, raw_text, created_at | **无 schema_version**、无 profiles/certificates/awards/portfolio 字段 |
| `users` | id, phone, password_hash, name, email, created_at, updated_at | 够用 |
| `recommendation_results` | id, user_id, resume_version_id, result_json | 外键关联 resume_versions |
| `match_analyses` | id, user_id, resume_version_id, job_id, score, result_json | 同上 |
| `version_compare_results` | id, from_version_id, to_version_id, result_json | 外键关联 resume_versions |

---

## 3. 存储方案比较

| 方案 | 内容 | 优点 | 风险 | 推荐 |
|:---:|------|------|------|:---:|
| **A** | 全 JSON 放一个字段 | 极简 | 不可查询，Agent 解析成本高 | ❌ |
| **B** | 主表 + 10 个模块 JSON 字段 + 摘要结构化列 | 结构清晰，版本易管理，Agent 直接组装 | JSON 字段多，列数增加 | ✅ **推荐** |
| C | 主表 + 6 个子表（关系型） | 规范可查询 | 开发量大，迁移复杂，不适合 Demo 阶段 | ❌ 后期可选 |

**冻结：方案 B — 主表 + 10 个模块 JSON 字段 + schema_version + 摘要结构化列。**

---

## 4. 推荐存储模型

### 4.1 结构化列（新增+保留）

| 列名 | 类型 | 原因 |
|------|------|------|
| `id` | BIGINT PK | 保留 |
| `user_id` | BIGINT FK | 保留 |
| `version_no` | INT | 保留 |
| `schema_version` | **TINYINT (新增)** | V1=1, V2=2，区分新旧数据 |
| `name` | VARCHAR(100) | 保留 |
| `phone` | VARCHAR(50) | 保留 |
| `email` | VARCHAR(255) | 保留 |
| `location` | VARCHAR(100) | 保留 |
| `target_city` | VARCHAR(100) | 保留 |
| `target_role` | VARCHAR(100) | 保留 |
| `target_direction` | **VARCHAR(100) (新增)** | 替代 target_role |
| `job_type` | **VARCHAR(50) (新增)** | 校招/实习 |
| `expected_cities` | **VARCHAR(255) (新增)** | 摘要，逗号分隔 |
| `school` | **VARCHAR(100) (新增)** | 摘要列 |
| `degree` | **VARCHAR(20) (新增)** | 摘要列 |
| `major` | **VARCHAR(100) (新增)** | 摘要列 |
| `graduation_year` | **VARCHAR(10) (新增)** | 摘要列 |
| `created_at` | DATETIME | 保留 |

### 4.2 JSON 模块字段（新增/替换）

| 列名 | JSON 内容 | 旧字段映射 |
|------|---------|-----------|
| `profile_json` | name, school, degree, major, graduation_year, phone, email, city | 部分从旧列迁移 |
| `job_intention_json` | target_roles, job_type, expected_cities, accept_other, onboard_date, internship_duration, weekly_days, target_industries, target_companies | 新 |
| `education_json` | [{school, degree, major, department, start, end, gpa, gpa_scale, ranking, courses, research, advisor, papers}] | 扩展 `education_json` |
| `internship_json` | [{company, role, start, end, content, tech_stack, achievements, metrics}] | 扩展 `experience_json` |
| `project_json` | [{name, role, type, tech_stack, start, end, description, responsibility, achievements, link}] | 扩展 `project_json` |
| `skills_json` | {languages, frameworks, databases, middleware, cloud, tools, ai_skills, other_keywords} | 替换 `skills_json` |
| `certificates_json` | [{type, name, level, score, date}] | **新增** |
| `awards_json` | [{type, name, level, date, description}] | **新增** |
| `portfolio_json` | {github, gitee, blog, homepage, demo_links, other_links}[] | **新增** |
| `extra_json` | {self_evaluation, additional_info} | 迁移 `self_evaluation` |

### 4.3 清理列

| 列 | 处理 |
|----|------|
| `target_city` | 保留（摘要），详细数据在 `job_intention_json` |
| `target_role` | 保留（摘要） |
| `education_json` | V2 内容扩展，V1 保留兼容 |
| `experience_json` | 重命名为 `internship_json` |
| `skills_json` | V2 改为对象结构，V1 兼容 |
| `self_evaluation` | 迁移到 `extra_json` |
| `raw_text` | 保留 |
| `location` | 弃用，迁移到 `profile_json.city` |

---

## 5. Request / Response V2 设计

### 5.1 CreateResumeRequest V2

```go
type CreateResumeRequestV2 struct {
    SchemaVersion  int                     `json:"schema_version"`   // 固定=2
    ResumeName     string                  `json:"resume_name"`
    TargetDirection string                `json:"target_direction"`
    Description    string                  `json:"description"`

    Profile        ResumeProfileV2         `json:"profile"`
    JobIntention   ResumeJobIntentionV2    `json:"job_intention"`
    Educations     []ResumeEducationV2     `json:"educations"`
    Internships    []ResumeInternshipV2    `json:"internships"`
    Projects       []ResumeProjectV2       `json:"projects"`
    Skills         ResumeSkillsV2          `json:"skills"`
    Certificates   []ResumeCertificateV2   `json:"certificates"`
    Awards         []ResumeAwardV2         `json:"awards"`
    PortfolioLinks ResumePortfolioV2       `json:"portfolio_links"`
    ExtraInfo      ResumeExtraInfoV2       `json:"extra_info"`
}
```

### 5.2 Field-Level Structures

```
ResumeProfileV2: name, school, degree, major, department, graduation_year, phone, email, current_city
ResumeJobIntentionV2: target_roles[], job_type, expected_cities[], accept_other, onboard_date, internship_duration, weekly_days, target_industries[], target_companies[]
ResumeEducationV2: school, degree, major, department, start_date, end_date, gpa, gpa_scale, ranking, courses[], research_direction, advisor, papers[]
ResumeInternshipV2: company, role, start_date, end_date, content, tech_stack[], achievements[], metrics[]
ResumeProjectV2: name, role, type, tech_stack[], start_date, end_date, description, responsibility, achievements[], link
ResumeSkillsV2: languages[], frameworks[], databases[], middleware[], cloud[], tools[], ai_skills[], other_keywords[]
ResumeCertificateV2: type, name, level, score, date
ResumeAwardV2: type, name, level, date, description
ResumePortfolioV2: github, gitee, blog, homepage, demo_links[], other_links[]
ResumeExtraInfoV2: self_evaluation, additional_info
```

### 5.3 ResumeVersionResponse V2

```go
type ResumeVersionResponseV2 struct {
    ID             uint64                  `json:"id"`
    VersionNo      int                     `json:"version_no"`
    SchemaVersion  int                     `json:"schema_version"`
    ResumeName     string                  `json:"resume_name"`
    // 摘要列（列表页可直接用）
    TargetDirection string               `json:"target_direction"`
    JobType        string                  `json:"job_type"`
    ExpectedCities string                  `json:"expected_cities"`
    School         string                  `json:"school"`
    Degree         string                  `json:"degree"`
    Major          string                  `json:"major"`
    GraduationYear string                  `json:"graduation_year"`
    // 完整模块数据
    Profile        ResumeProfileV2         `json:"profile"`
    JobIntention   ResumeJobIntentionV2    `json:"job_intention"`
    Educations     []ResumeEducationV2     `json:"educations"`
    Internships    []ResumeInternshipV2    `json:"internships"`
    Projects       []ResumeProjectV2       `json:"projects"`
    Skills         ResumeSkillsV2          `json:"skills"`
    Certificates   []ResumeCertificateV2   `json:"certificates"`
    Awards         []ResumeAwardV2         `json:"awards"`
    PortfolioLinks ResumePortfolioV2       `json:"portfolio_links"`
    ExtraInfo      ResumeExtraInfoV2       `json:"extra_info"`
    CreatedAt      string                  `json:"created_at"`
}
```

---

## 6. API 版本策略

| 方案 | 结论 |
|------|------|
| A：直接扩展现有 `/api/resumes` | ❌ 旧 payload 混入风险 |
| B：新增 `/api/resumes/v2` | ❌ 路径膨胀 |
| **C：同路径 + `schema_version: 2`** | ✅ **推荐** |

**冻结**：`POST /api/resumes` 接受 `schema_version` 字段。`schema_version=1` 或不传 → 旧逻辑；`schema_version=2` → 新逻辑。Response 也返回 `schema_version`。

### 向后兼容

```go
if req.SchemaVersion < 2 {
    // 旧 payload → 旧 service 逻辑（V1 adapter）
    return s.CreateVersionV1(ctx, userID, req)
}
// 新 payload → 新 service 逻辑
return s.CreateVersionV2(ctx, userID, req)
```

---

## 7. 旧数据兼容策略

| 旧字段 | 新字段 | 兼容方式 |
|--------|--------|---------|
| `schema_version` 缺失 | — | 读取时默认=1 |
| `target_role` | `target_direction` / `job_intention.target_roles` | V1→V2 adapter 映射 |
| `location` | `profile.city` | V1→V2 adapter 迁移 |
| `target_city` | `job_intention.expected_cities` | V1→V2 adapter 映射 |
| `education_json` (V1) | `education_json` (V2) | V1 嵌套结构 → V2 字段扩展 |
| `experience_json` | `internship_json` | V1 字段名映射 |
| `skills_json` (V1 array) | `skills_json` (V2 object) | V1 `string[]` → V2 `other_keywords[]` |
| `self_evaluation` (列) | `extra_json.self_evaluation` | V1→V2 adapter 迁移 |

**原则**：
- V1 数据只读（历史版本不可变）
- 读取 V1 时通过 `NormalizeResumeToSnapshot()` 统一成 V2 格式
- 写入始终用最新版本

---

## 8. 数据库迁移策略

| 阶段 | 操作 | 风险 |
|:---:|------|------|
| **1** | ALTER TABLE 新增列：schema_version, target_direction, job_type, expected_cities, school, degree, major, graduation_year, profile_json, job_intention_json, internship_json, certificates_json, awards_json, portfolio_json, extra_json | 低（新增列，NULL 默认） |
| 2 | 保留旧列：education_json, experience_json, project_json, skills_json, self_evaluation, raw_text, location | 无 |
| 3 | 后端开发 NormalizeResumeToSnapshot() | 中（需测试 V1/V2 全路径） |
| 4 | 前端切 V2 payload | 高（需联调） |
| 5 | Agent 切 V2 ResumeSnapshot | 高（Prompt 联动） |
| 6 | 清理旧列（可选，后期） | 低 |

---

## 9. ResumeSnapshot 读取策略

```
DB: resume_versions (V1 or V2)
  ↓ adapter check: schema_version
  ↓
NormalizeResumeToSnapshot(dbRow)
  ↓ V1: 映射+扩展默认值 → V2
  ↓ V2: 直接组装
  ↓
ResumeSnapshotV2 {
  profile, job_intention, educations,
  internships, projects, skills,
  certificates, awards, portfolio_links, extra_info
}
  ↓
Recommendation / Match / Optimization / Comparison Agents
```

**要求**：
- Agent 不直接读 DB 原始 JSON
- `NormalizeResumeToSnapshot()` 在后端 service 层实现
- V1/V2 都输出统一的 ResumeSnapshotV2

---

## 10. 字段校验策略

| 字段 | 校验策略 |
|------|---------|
| `resume_name` | 非空，max 100 |
| `target_direction` | 非空，max 100 |
| `profile.school` | 推荐非空 |
| `profile.degree` | 推荐非空 |
| `profile.phone` | 格式校验（11 位） |
| `profile.email` | email 格式 |
| `job_intention.target_roles` | 至少 1 项 |
| `job_intention.job_type` | 枚举 |
| `internships[].company` | 推荐非空 |
| `projects[].name` | 推荐非空 |
| `skills.*` | 推荐至少 1 类非空 |
| 其他 | **不强制**，优先允许保存，AI 提示补充 |

**Demo 原则**：轻校验、重 AI 提示。不强校验 GPA/证书/奖项为空。

---

## 11. 后端影响范围

| 层级/文件 | 需改 | 说明 |
|-----------|:---:|------|
| `server/sql/init.sql` | ✅ | ALTER 新增列 + 种子数据 |
| `server/sql/migration-j1.sql` | ✅ | 新增 migration |
| `server/internal/model/resume.go` | ✅ | ResumeVersion 新增字段 |
| `server/internal/service/resume_service.go` | ✅ | CreateResumeRequest V2 + 兼容 + NormalizeResumeToSnapshot() |
| `server/internal/repository/resume_repository.go` | ✅ | 新增 JSON 字段读写 |
| `server/internal/http/resume_handler.go` | ✅ | 新增字段解码/编码 |
| `web/src/types/resume.ts` | ✅ | 新类型 |
| `web/src/api/resumes.ts` | ✅ | payload 扩展 |
| `web/src/adapters/resumeAdapter.ts` | ✅ | V2 ViewModel |
| `web/src/api/viewModels.ts` | ✅ | V2 loader |
| Agent service | ✅ | 切换 ResumeSnapshot |
| Agent prompt/schema | ⚠️ | D 阶段 |
| Mock 数据 | ✅ | V2 种子 |

---

## 12. 问题分级

| 级别 | 问题 | 影响 | 后续阶段 |
|:---:|------|------|:---:|
| **P0** | 不先设计 `schema_version` 导致 V1/V2 混乱 | 数据不可识别 | **本阶段已冻结** ✅ |
| **P1** | Agent 无 normalize 层，直接依赖 DB 结构 | Prompt 耦合 DB，后续改 DB 风险高 | D |
| P1 | 摘要列不足影响列表/筛选性能 | 版本列表需 JOIN 或全量加载 | C 实现 |
| P2 | 过强校验降低填写意愿 | 用户流失 | E |

---

## 13. 后续阶段建议

| 编号 | 名称 |
|------|------|
| **K8-5-21-D** | ResumeSnapshot / Agent Schema 设计冻结 |
| **K8-5-21-E** | 前端页面结构设计冻结 |
| **K8-5-21-F** | 分阶段实现计划冻结 |

---

## 14. 构建 / API 结果

| 验收项 | 结果 |
|--------|------|
| `npm run build` | ✅ 通过 (2.42s) |
| `GET /api/jobs` | ✅ 200 |
| `GET /api/resumes` | ✅ 401（后端保护正常） |

---

## 15. 未修改范围确认

```text
未修改源码/样式/前端页面/后端/数据库/SQL/接口/.env/LLM/Prompt/Schema
未新增字段/接口/Agent/Mock
未安装依赖/未迁移目录
```
