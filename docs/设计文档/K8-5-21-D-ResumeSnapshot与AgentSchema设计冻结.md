# K8-5-21-D — ResumeSnapshot / Agent Schema 设计冻结

> 日期：2026-06-04 | 性质：design freeze | 未修改源码

---

## 1. 设计背景

基于 K8-5-21-A/B/C 的冻结成果：
- K8-5-21-A：当前简历字段约 30 个，education/experience/project/skills 以 JSON 存储，Agent 通过 `resume_version_no` 桥接
- K8-5-21-B：新版简历冻结为 10 模块 ~90 字段，定位学生求职画像采集系统
- K8-5-21-C：数据库冻结为主表 + schema_version + 10 JSON 模块 + 摘要列；API 同路径 `schema_version` 区分 V1/V2；`NormalizeResumeToSnapshot()` 统一输出

**本阶段核心问题**：
- 当前无统一 `ResumeSnapshot`：四个 Agent 各自直接使用 `ResumeVersion` 的旧字段 patch，各自组装不同的 prompt 输入
- Prompt 感知 DB 原始 JSON 结构（`target_role`、`skills` 数组、`experience_json` 等），与 B/C 阶段设计的新模块字段脱节
- Mock Agent 仅覆盖旧字段，不感知新版证书/获奖/作品链接/AI 技能
- Response Schema 缺少对 resume evidence 的结构化引用字段

---

## 2. 核心设计结论

```text
1. 所有 Agent 统一接收 ResumeSnapshotV2。
2. DB 原始结构和 V1/V2 差异必须在 Normalize 层消化。
3. Prompt 不直接感知 DB 字段名（如 target_role、experience_json 等）。
4. NormalizeResumeToSnapshot() 在后端 service 层实现，V1/V2 都输出统一的 ResumeSnapshotV2。
5. Snapshot = 11 模块（10 页面模块 + meta）。
6. 手机号/邮箱/内部 ID/登录 token/原始附件 不进入 Agent。
7. Prompt 差异化裁剪：每个功能选择需要的模块，空模块不传或标记"未填写"。
```

---

## 3. 当前 Agent 现状

### 3.1 四核心功能输入输出总表

| 功能 | 当前请求字段 | 当前使用简历方式 | 当前返回字段 | 问题 |
|------|------------|---------------|-------------|------|
| **岗位推荐** | `resume_version_no` → 后端查 `ResumeVersion` | `buildPromptValues` 直接 patch `target_role/target_city/skills/education/experience/project/self_evaluation` | `RecommendedJob[]` (match_score/priority/evidence/risk_points/top_fix_items...) | 无 GPA/证书/奖项/AI技能/作品链接；prompt 感知旧字段名 |
| **岗位匹配** | `resume_version_no` + `jobId` | 同上 | `MatchAnalysisResult` (match_score/evidence_chain/gap_points/risk_points...) | 同上 |
| **简历优化** | `resume_text` + `target_job` (流式) | `buildPromptValues` 用 `resume_text/target_section/target_roles/expected_locations` | `ResumeOptimizationResult` (summary/top_fix_items/jd_evidence_chain...) | 非结构化文本模式，未按新模块字段组织 |
| **简历对比** | `from_version_no` + `to_version_no` | `resumeToMap` 直接 patch 旧字段 | `VersionCompareAnalysis` (score/field_changes/risk_changes...) | 仅对比旧字段，缺 8 个新模块 |

### 3.2 当前 Agent 输入构建方式（均直接使用 DB 旧字段）

```go
// recommendation_agent.go / match_agent.go - buildPromptValues
"version_no":      resume.VersionNo,
"target_role":     resume.TargetRole,
"target_city":     resume.TargetCity,
"skills":          safeParseJSONArray(resume.SkillsJSON),
"education":       safeParseJSONArray(resume.EducationJSON),
"experience":      safeParseJSONArray(resume.ExperienceJSON),
"project":         safeParseJSONArray(resume.ProjectJSON),
"self_evaluation": resume.SelfEvaluation,
```

```go
// version_compare_agent.go - resumeToMap
// 完全相同的字段列表
```

```go
// resume_optimization_agent.go - buildPromptValues
"resume_text":     input.ResumeText,
"target_section":  input.TargetSection,
"target_roles":    input.TargetRoles,
"expected_locations": input.ExpectedLocations,
"accept_relocation":  input.AcceptRelocation,
```

**关键发现**：四个 Agent 都不存在 `ResumeSnapshot` 统一结构，Prompt 直接感知 `target_role`、`skills`、`experience_json` 等 V1 旧字段名。无 normalize 层。

### 3.3 当前 Prompt 中简历字段引用

| Prompt 文件 | 简历字段 placeholder | 实际注入内容 |
|-------------|-------------------|-------------|
| `recommendation_v1.md` | `{{RESUME_CONTEXT_JSON}}` 含 `target_role/skills/education/experience/project/self_evaluation` | 旧字段 patch |
| `match_analysis_v1.md` | 同上 | 同上 |
| `resume_optimization_v1.md` | `{{RESUME_CONTEXT_JSON}}` 含 `resume_text/target_section/...` | 文本模式 |
| `version_compare_v1.md` | `{{OLD_RESUME_JSON}}` / `{{NEW_RESUME_JSON}}` | 旧字段 patch |

### 3.4 当前 Response 中 evidence 引用方式

| Agent | 当前 evidence 结构 | 是否引用 resume 具体字段 |
|-------|-----------------|----------------------|
| **Recommendation** | `{job_requirement, resume_evidence, judgement, explanation}` | 仅文本级，无结构化字段引用 |
| **Match** | `{jd_requirement, resume_evidence, judgement, explanation}` | 同上 |
| **Optimization** | `{jd_keyword, resume_keyword, alignment_status, gap_analysis, screening_risk}` | keyword 级对齐，但无模块归属 |
| **Comparison** | `{field, change}` | `field` 仅字符串（"target_role"/"skills"/"project"等 V1 字段名） |

---

## 4. ResumeSnapshotV2 结构

### 4.1 完整类型定义

```ts
type ResumeSnapshotV2 = {
  // ─── Meta（元信息，非页面模块） ───
  meta: {
    schemaVersion: 2;
    resumeVersionNo: string;
    resumeName: string;
    targetDirection: string;
    updatedAt?: string;
  };

  // ─── 模块 1: 基础信息 ───
  profile: {
    name?: string;
    school?: string;
    degree?: string;
    major?: string;
    graduationYear?: string;
    currentCity?: string;
    // phone / email 不进入 Agent
  };

  // ─── 模块 2: 求职意向 ───
  jobIntention: {
    targetRoles: string[];
    jobType?: "internship" | "campus" | "daily_internship" | "other";
    expectedCities: string[];
    acceptOtherCities?: boolean;
    earliestStartDate?: string;
    internshipDuration?: string;
    weeklyAvailableDays?: number;
    targetIndustries?: string[];
    targetCompanies?: string[];
  };

  // ─── 模块 3: 教育经历 ───
  educations: Array<{
    school: string;
    degree?: string;
    major?: string;
    department?: string;
    startDate?: string;
    endDate?: string;
    gpa?: string;
    gpaScale?: string;
    rank?: string;
    coreCourses?: string[];
    researchDirection?: string;
    lab?: string;
    paperSummary?: string;
  }>;

  // ─── 模块 4: 实习经历 ───
  internships: Array<{
    company: string;
    role: string;
    startDate?: string;
    endDate?: string;
    workContent?: string;
    techStack?: string[];
    achievements?: string[];
    metrics?: string;
  }>;

  // ─── 模块 5: 项目经历 ───
  projects: Array<{
    name: string;
    role?: string;
    type?: "course" | "competition" | "personal" | "enterprise" | "open_source" | "other";
    startDate?: string;
    endDate?: string;
    techStack?: string[];
    description?: string;
    responsibilities?: string;
    achievements?: string[];
    link?: string;
  }>;

  // ─── 模块 6: 技能信息 ───
  skills: {
    programmingLanguages?: string[];
    backendFrameworks?: string[];
    frontendTechnologies?: string[];
    databases?: string[];
    middleware?: string[];
    cloudPlatforms?: string[];
    devTools?: string[];
    aiTools?: string[];
    aiApplicationScenarios?: string[];
    otherKeywords?: string[];
  };

  // ─── 模块 7: 证书与语言 ───
  certificates: Array<{
    name: string;
    type?: "language" | "cloud" | "software" | "other";
    level?: string;
    score?: string;
    issuedAt?: string;
  }>;

  // ─── 模块 8: 获奖与竞赛 ───
  awards: Array<{
    name: string;
    type?: "scholarship" | "competition" | "honor" | "research" | "other";
    level?: "school" | "city" | "province" | "national" | "international" | "other";
    awardedAt?: string;
    description?: string;
  }>;

  // ─── 模块 9: 作品链接 ───
  portfolio: {
    github?: string;
    gitee?: string;
    cnb?: string;
    blog?: string;
    demoLinks?: string[];
    portfolioLinks?: string[];
    homepage?: string;
  };

  // ─── 模块 10: 补充信息 ───
  extraInfo: {
    selfEvaluation?: string;
    additionalInfo?: string;
  };
};
```

### 4.2 模块总表

| 序号 | 模块 | 字段数 | 进入 Agent | 说明 |
|:---:|------|:---:|:---:|------|
| — | **meta** | 5 | ✅ 元信息 | `schemaVersion`, `resumeVersionNo`, `resumeName`, `targetDirection`, `updatedAt` |
| 1 | **profile** | 6 | ✅ | 不含 phone/email |
| 2 | **jobIntention** | 9 | ✅ | 推荐/匹配核心 |
| 3 | **educations** | 数组(每项13) | ✅ | 含 GPA/排名/课程/院系/论文 |
| 4 | **internships** | 数组(每项7) | ✅ | 含技术栈/成果/量化 |
| 5 | **projects** | 数组(每项10) | ✅ | 含类型/成果/链接 |
| 6 | **skills** | 10 | ✅ | 分层技能(AI/云/语言/框架/数据库...) |
| 7 | **certificates** | 数组(每项5) | ✅ | 含语言/云/软件证书 |
| 8 | **awards** | 数组(每项5) | ✅ | 含奖学金/竞赛/荣誉 |
| 9 | **portfolio** | 7 | ✅ | 含 GitHub/博客/主页/demo |
| 10 | **extraInfo** | 2 | ✅ | 自我评价/补充 |
| 合计 | 11 模块 | ~90 字段 | | phone/email 仅前端展示，不进 Agent |

### 4.3 命名与默认值规则

```text
1. 字段名使用 camelCase 英文语义，不混用旧字段名（如不叫 target_role、experience_json）。
2. 所有数组字段默认空数组 []。
3. 所有可选文本字段默认空字符串 "" 或省略。
4. Agent 看到的是标准化后的 Snapshot，不是 DB 原始结构。
5. meta.moduleVersion 固定为 2，表示 ResumeSnapshotV2。
```

---

## 5. V1 → V2 Normalize 策略

### 5.1 映射表

| V1 字段/路径 | V2 字段 | 默认值 / 处理 |
|-------------|--------|-------------|
| `name` | `meta.resumeName` | 直接映射 |
| `target_role` | `meta.targetDirection` + `jobIntention.targetRoles[0]` | 拆分为方向 + 数组第 1 项 |
| `target_city` | `jobIntention.expectedCities[0]` | 单个拆入数组 |
| `location` | `profile.currentCity` | 直接映射 |
| `profile.name` (raw_text) | `profile.name` | 无则 "" |
| `profile.school` (raw_text) | `profile.school` | 无则 "" |
| `profile.degree` (education_json) | `profile.degree` | 取第一条 education.degree |
| `profile.major` (raw_text) | `profile.major` | 无则 "" |
| `profile.graduation_year` | `profile.graduationYear` | 无则 "" |
| `phone` | **不进入 Snapshot** | 前端展示用 |
| `email` | **不进入 Snapshot** | 前端展示用 |
| `education_json` (V1 数组) | `educations[]` | 字段映射：school→school, major→major, degree→degree, start→startDate, end→endDate, courses→coreCourses。缺失字段给 "" |
| `experience_json` (V1 数组) | `internships[]` | 字段映射：company→company, role→role, start→startDate, end→endDate, description→workContent。techStack/achievements/metrics 给 [] |
| `project_json` (V1 数组) | `projects[]` | 字段映射：name→name, description→description, tech_stack→techStack, responsibilities→responsibilities。role/type/achievements/link 给默认值 |
| `skills_json` (V1 `string[]`) | `skills.otherKeywords` | V1 扁平技能全部归入 otherKeywords，其他分类字段给 [] |
| `self_evaluation` | `extraInfo.selfEvaluation` | 直接映射 |
| `raw_text` | **不进入 Snapshot** | 预留文本 |
| `user_id` / `id` | **不进入 Snapshot** | DB 内部 ID |
| schema_version=1 | `meta.schemaVersion = 2` | 统一输出 V2 |
| 8 个新模块（certificates/awards/portfolio 等） | 对应 V2 字段 | **全部给空数组 / 空默认值** |

### 5.2 Normalize 实现位置

```
后端 service 层：
  func NormalizeResumeToSnapshot(dbRow *model.ResumeVersion) *ResumeSnapshotV2

调用链：
  Agent.Recommend/Match/Optimize/Compare
    → service 层先调 NormalizeResumeToSnapshot()
    → Agent buildPromptValues(snapshot)
    → Prompt 注入展开后的 snapshot JSON
```

### 5.3 兼容原则

```text
1. V1 数据只读（历史版本不可变），归一化后供 Agent 使用。
2. V2 数据直接按新字段组装 Snapshot。
3. Agent 永远只接收 ResumeSnapshotV2，不感知 V1/V2 差异。
4. Prompt 不需要知道"这个数据是 V1 归一化来的还是 V2 直接取的"。
```

---

## 6. 四核心功能字段使用策略

### 6.1 简历版本对比

| 模块 | 对比关注点 | 进入 Agent |
|------|----------|:---:|
| **meta** | 简历名称变更、目标方向调整 | ✅ |
| **profile** | 学校/学历/专业/毕业时间字段变化 | ✅ |
| **jobIntention** | 目标岗位/城市/岗位类型/可到岗时间变化 | ✅ |
| **educations** | 教育经历新增/删除、GPA/排名/课程补充 | ✅ |
| **internships** | 实习经历新增/删除、技术栈/成果/量化数据变化 | ✅ |
| **projects** | 项目新增/删除、技术栈变化、成果量化变化、职责表达变化 | ✅ |
| **skills** | 技能分类变化、AI 工具能力变化、云平台/中间件补充 | ✅ |
| **certificates** | 证书新增/删除/等级变化 | ✅ |
| **awards** | 获奖新增/删除/级别变化 | ✅ |
| **portfolio** | 作品链接新增/删除 | ✅ |
| **extraInfo** | 自我评价内容变化 | ✅ |

**对比策略**：全部 11 模块进入 Agent，LLM 逐模块分析字段级变化、评估改进方向、标记缺失项。

### 6.2 岗位智能推荐

| 模块 | 推荐使用方式 | 进入 Agent |
|------|------------|:---:|
| **jobIntention** | `targetRoles` 决定方向过滤；`expectedCities` 决定地域过滤；`weeklyAvailableDays` / `internshipDuration` 决定实习岗位可行性；`acceptOtherCities` 扩展推荐范围 | ✅ |
| **profile** | `school/degree/major/graduationYear` 决定匹配候选池（如校招/实习/日常实习） | ✅ |
| **skills** | `programmingLanguages/backendFrameworks/databases/cloudPlatforms` 决定技术方向匹配；`aiTools/aiApplicationScenarios` 决定 AI 岗位加权 | ✅ |
| **projects** | 项目类型/技术栈/描述与岗位 JD 的交集 | ✅ |
| **internships** | 实习公司/岗位/技术栈与岗位的关联度 | ✅ |
| **certificates** | `cloud` 类型证书 → 云计算岗位加权；`language` 类型证书 → 外企/出海岗位加权 | ✅ |
| **awards** | 竞赛获奖 → 技术岗位加权；奖学金 → 学术能力参考 | ✅ |
| **portfolio** | GitHub → 代码能力验证（可选） | ✅ |
| **educations** | GPA/排名 → 部分外企硬门槛；课程 → 方向匹配 | ✅ |
| **extraInfo** | 自我评价补充意向信息 | ✅ |

### 6.3 岗位匹配分析

| 模块 | 匹配使用方式 | 进入 Agent |
|------|------------|:---:|
| **JD 本身** | 核心要求、技术栈、岗位职责、任职要求 | ✅ |
| **jobIntention** | 目标方向是否与岗位一致；城市是否匹配 | ✅ |
| **skills** | 分层技能与 JD 关键词逐个对齐（programmingLanguages 对 JD 语言要求，cloudPlatforms 对 JD 云要求等） | ✅ |
| **projects** | `techStack + description` 与 JD 要求逐项 evidence mapping | ✅ |
| **internships** | `techStack + workContent` 与 JD 要求匹配 | ✅ |
| **educations** | `school/degree/major/GPA` 是否满足 JD 学历门槛 | ✅ |
| **certificates** | `cloud` 证书 → 匹配 JD 云平台要求；`language` 证书 → 匹配 JD 语言要求 | ✅ |
| **awards** | 竞赛/荣誉 → 匹配 JD 专业技能要求 | ✅ |
| **portfolio** | GitHub → 匹配 JD 代码能力要求 | ✅ |
| **extraInfo** | 自我评价中是否体现岗位相关能力 | ✅ |

### 6.4 简历优化建议

| 模块 | 优化使用方式 | 进入 Agent |
|------|------------|:---:|
| **所有模块** | 针对目标岗位 JD，逐模块检查：<br>1. 是否有直接证据（matched）<br>2. 是否有间接证据（partial）<br>3. 是否缺失（missing） | ✅ |
| **projects** | P0：缺少成果量化 → 建议补充；P1：表达偏弱 → 建议展开；P2：顺序问题 → 建议重排 | ✅ |
| **skills** | P0：岗位核心技能缺失 → 建议补充证据；P1：技能太散 → 建议归类；P2：overstated → 建议诚实表达 | ✅ |
| **internships** | P0：缺少技术栈 → 建议补充；P1：缺少成果 → 建议量化 | ✅ |
| **extraInfo** | P1：太泛 → 建议改为和目标岗位相关的总结 | ✅ |
| **certificates/awards/portfolio** | P2：有空缺 → 建议补充相关证书或作品 | ✅ |

---

## 7. 字段进入 Agent 策略

### 7.1 必须进入 Agent（11 个模块全部）

| 字段/模块 | 是否进入 Agent | 原因 |
|----------|:---:|------|
| `meta` | ✅ | 识别版本、简历名称、方向 |
| `profile`（不含 phone/email） | ✅ | 身份/学历/城市匹配 |
| `jobIntention` | ✅ | 推荐/匹配核心 |
| `educations` | ✅ | 学历匹配 |
| `internships` | ✅ | 经历匹配 |
| `projects` | ✅ | 能力验证 |
| `skills` | ✅ | 技能匹配 |
| `certificates` | ✅ | 专业/语言匹配 |
| `awards` | ✅ | 学术/竞赛能力 |
| `portfolio` | ✅ | 代码验证 |
| `extraInfo` | ✅ | 优化建议素材 |

### 7.2 不进入 Agent

| 字段 | 原因 |
|------|------|
| `phone` | 隐私，前端展示用 |
| `email` | 隐私，前端展示用 |
| `user_id` / `id` | 内部 DB 主键 |
| 登录 token | 安全 |
| 原始附件 (raw_text) | 非结构化，不入 Prompt |
| 身份证/证件 | ATS 字段，未采集 |
| 性别/年龄/QQ/微信 | 非 AI 必要 |

### 7.3 仅前端展示不入 Agent

| 字段 | Agent 使用 | 前端展示 |
|------|:---:|:---:|
| `profile.phone` | ❌ | ✅ 脱敏展示 |
| `profile.email` | ❌ | ✅ |
| `meta.createdAt` | ❌ | ✅ |
| `meta.updatedAt` | ❌ | ✅ |
| `resumeVersionNo` | ✅（仅 meta 中） | ✅ |

---

## 8. Prompt 组织策略

### 8.1 通用 Prompt 输入块

每个 Agent 的 Prompt 中，`{{RESUME_CONTEXT_JSON}}` 替换为以下标准化结构：

```json
{
  "meta": { "schemaVersion": 2, "resumeVersionNo": "1", "resumeName": "xxx", "targetDirection": "后端开发" },
  "profile": { "name": "...", "school": "...", "degree": "硕士", "major": "...", "graduationYear": "2027", "currentCity": "深圳" },
  "jobIntention": { "targetRoles": ["后端开发"], "jobType": "internship", "expectedCities": ["深圳", "北京"], "weeklyAvailableDays": 5, "internshipDuration": "3个月" },
  "educations": [{ ... }],
  "internships": [{ ... }],
  "projects": [{ ... }],
  "skills": { "programmingLanguages": [...], "backendFrameworks": [...], "databases": [...], "cloudPlatforms": [...], "aiTools": [...] },
  "certificates": [{ ... }],
  "awards": [{ ... }],
  "portfolio": { "github": "...", "blog": "..." },
  "extraInfo": { "selfEvaluation": "..." }
}
```

### 8.2 字段裁剪策略

```text
1. 空模块不传或标记 "未填写"：
   - 如 certificates 为空数组，传 "未填写" 或 []。
   - AI 看到空模块后应明确"简历中暂未包含证书/获奖信息"，不编造。

2. 过长字段做摘要：
   - workContent/description 超过 500 字？保留全文，Prompt 中不做截断（让 LLM 自己处理）。
   - 但每个经历最多保留关键字段，避免 Prompt 无限膨胀。

3. 作品链接只传标题和 URL：
   - github/blog 只传链接，不抓取网页内容。
   - Prompt 中明确告知："URL 仅供参考，不作内容抓取"。

4. phone/email 不传（见 7.2）。

5. 每个经历最多保留关键字段：
   - internships：company, role, startDate, endDate, workContent, techStack, achievements, metrics
   - 其他字段（如 performance rating 等不存在的）省略。
```

### 8.3 功能差异化 Prompt 策略

| 功能 | Prompt 重点 | 裁剪规则 |
|------|-----------|---------|
| **推荐** | 重点看 `jobIntention` + `skills` + `projects`；次要看 `internships` + `certificates` | profile 可精简到 school/degree/major/graduationYear |
| **匹配** | 重点看 JD evidence + resume evidence 逐条对照；`skills` 分层对齐 JD 关键词 | 全部模块保留，但 portfolios 可降为"有无 GitHub"布尔 |
| **优化** | 重点看缺失项、弱表达、可量化空间；逐模块扫描 P0/P1/P2 | 全部模块保留，portfolio 可选 |
| **对比** | 重点看字段变化、能力增强、风险变化；两版本逐模块 diff | 全部模块保留，不裁剪 |

### 8.4 Prompt 模板变更

本阶段**不改 Prompt 文件**，仅冻结策略。实现阶段：

```text
- recommendation_v1.md：{{RESUME_CONTEXT_JSON}} 不再传旧字段，传 ResumeSnapshotV2
- match_analysis_v1.md：同上
- resume_optimization_v1.md：{{RESUME_CONTEXT_JSON}} 从文本模式切为 Snapshot 模式，或新增 optional "resume_snapshot" 字段
- version_compare_v1.md：{{OLD_RESUME_JSON}} / {{NEW_RESUME_JSON}} 改用 Snapshot 结构
```

---

## 9. Response Schema 边界

### 9.1 当前 Response 中 evidence 引用方式

| Agent | 当前 evidence 字段 | 不足 |
|-------|------------------|------|
| **Recommendation** | `evidence[]: {job_requirement, resume_evidence, judgement, explanation}` | `resume_evidence` 是自由文本，无模块归属 |
| **Match** | `evidence_chain[]: {jd_requirement, resume_evidence, judgement, explanation}` | 同上 |
| **Optimization** | `keyword_alignment[]: {keyword, status, resume_evidence, suggested_action}` | `resume_evidence` 仅指关键词，无字段归属 |
| **Comparison** | `improved_items[]: {field, change}` | `field` 是 V1 旧字段名（"target_role"、"skills"），未来需映射为新模块名 |

### 9.2 Response 需要新增的 evidence 字段

| Agent | 建议新增 | 用途 |
|-------|---------|------|
| **Recommendation** | `evidence[].resume_module` | 标注 evidence 来自哪个 Snapshot 模块（如 "skills"、"projects"、"certificates"），让前端按模块展示 |
| **Match** | `evidence_chain[].resume_module` | 同上 |
| **Optimization** | `priority_suggestions[].resume_module` | 标注建议针对哪个模块 |
| **Comparison** | `improved_items[].module_name` | 将当前 `field` 改为按模块分组，字段级别粒度保留但不作为唯一标识 |

**实现优先级**：Response Schema 的模块归属字段（`resume_module`）在实现阶段新增，本次仅冻结设计方向。

### 9.3 Recommendation/Match/Optimization 是否引用 snapshot evidence

| 问题 | 结论 |
|------|------|
| Recommendation response 是否需要引用 snapshot evidence | ✅ 需要，已有 `evidence` 字段，仅需加 `resume_module` |
| Match response 是否需要引用 resume evidence | ✅ 需要，已有 `evidence_chain`，同上 |
| Optimization response 是否需要指出缺失字段 | ✅ 需要，已有 `missing_evidence_plan` 和 `keyword_alignment`，实现时改为模块粒度 |
| Comparison response 是否需要区分字段变化类型 | ✅ 需要，已有 `improved_items[]/weakened_items[]`，实现时字段名改为模块名 |

---

## 10. Mock Agent 策略

### 10.1 当前 Mock 覆盖情况

| Mock Agent | 覆盖字段 | 缺失 |
|-----------|---------|------|
| `RecommendationMockAgent` | 基于旧字段 keyword 匹配 | 无证书/获奖/作品/AI技能相关逻辑 |
| `MatchMockAgent` | 基于旧字段 keyword 匹配 | 同上 |
| `ResumeOptimizationMockAgent` | 基于文本 keyword 匹配 | 同上 |
| `VersionCompareMockAgent` | 基于旧字段 score 计算 | 同上 |

### 10.2 Mock 需要补充的场景

| Mock 场景 | 用途 | 优先级 |
|----------|------|:---:|
| 有 GPA / 排名的教育背景 | 验证 Agent 能识别教育强化信号 | P1 |
| 有 AI 应用技能的 resume | 验证 skills.aiTools/aiApplicationScenarios 被正确消费 | P1 |
| 有云计算证书的 resume | 验证 certificates 被正确匹配（cloud 方向加权） | P1 |
| 有竞赛获奖的 resume | 验证 awards 被正确识别（level 分级） | P1 |
| 有作品链接的 resume | 验证 portfolio 被正确展示（GitHub/博客） | P2 |
| 项目缺少量化成果 | 验证 Optimization Agent 的 missing_evidence_plan | P1 |
| 实习缺少技术栈 | 验证 Optimization Agent 的 gap 检测 | P1 |
| 简历字段缺失但仍可分析 | 验证极端情况处理（空 certificates、空 awards） | P2 |
| 纯 V1 旧简历归一化为 V2 | 验证 NormalizeResumeToSnapshot() 对 V1 的默认值填充 | P1 |
| 全模块填写完整的 V2 简历 | 验证全字段 Snapshot 传递完整性 | P1 |

### 10.3 Mock 实现阶段建议

```text
实现阶段（K8-5-21-F 之后）：
1. 基于 ResumeSnapshotV2 结构重写 Mock Agent
2. 每个 Mock 返回结果中引用 snapshot 模块（resume_module 字段）
3. 新增至少 3 个 mock fixtures，覆盖 V1 归一化、V2 完整、V2 空缺三种情况
4. version_compare Mock 扩展为按模块对比
```

---

## 11. LLM Agent / Schema 适配策略

### 11.1 适配路径

```text
1. 后端新增 NormalizeResumeToSnapshot() 函数（service 层）
2. 各 Agent 的 buildPromptValues() 改为接收 ResumeSnapshotV2 而非 *model.ResumeVersion
3. Prompt 文件中的 {{RESUME_CONTEXT_JSON}} 内容从旧字段改为 Snapshot 结构
4. Schema 文件不变（当前 Schema 已定义 evidence/resume_evidence/resume_keyword 等通用字段，仅需 prompt 引导 LLM 按模块展开）
5. 前端 types/agent.ts 不需要立即改（Response Schema 仅在实现阶段调整）
```

### 11.2 对 Schema 的影响

| Schema 文件 | 当前 | 需改 | 说明 |
|------------|------|:---:|------|
| `recommendation_schema.json` | evidence 无模块归属 | ⚠️ | 可选新增 `resume_module`（非强制），不改 Schema 也可通过 prompt 引导 |
| `match_analysis_schema.json` | evidence_chain 无模块归属 | ⚠️ | 同上 |
| `resume_optimization_schema.json` | keyword_alignment 无模块归属 | ⚠️ | 同上 |
| `version_compare_schema.json` | improved_items.field 为 V1 字段名 | ⚠️ | 实现时可新增 `module_name` 替代 `field` |

**结论**：Schema 文件本身不需大改，主要变化在 Prompt 的输入结构。实现阶段可以为 Schema 新增 `resume_module`（可选字段），不改现有字段。

---

## 12. 问题分级

| 级别 | 问题 | 影响 | 后续阶段 |
|:---:|------|------|:---:|
| **P0** | 没有统一 ResumeSnapshotV2，四个 Agent 会各自理解简历字段，后续必乱 | 推荐/匹配/优化/对比使用不同字段名、不同结构，Prompt 不可维护 | **本阶段已冻结 Snapshot** ✅ |
| **P0** | Prompt 直接读取 DB JSON（`target_role`/`skills`/`experience_json`），会导致 V1/V2 兼容失败 | V1 数据与 V2 Agent 预期结构不一致，fallback 率上升 | E/F 引入 Normalize + Prompt 适配 |
| **P1** | 不做字段裁剪会导致 Prompt 过长、噪声过多 | 90 字段全量输入会使 token 成本翻倍，LLM 推理质量下降 | E/F 实现 Prompt 裁剪 |
| **P1** | Response 不引用 resume evidence 模块归属，用户不信任 AI 结论 | 用户无法在简历上定位 AI 建议对应的模块 | E/F Schema 新增 resume_module |
| **P1** | Mock 不覆盖新字段（证书/获奖/AI技能），会导致页面验收质量低 | 前端看不到新模块的 Agent 结果展示，视觉验收失真 | E/F 更新 Mock |
| P2 | 简历优化 Agent 当前用文本模式，不是结构化 Snapshot | 优化建议无法精确到字段级 | E/F 切换为 Snapshot |
| P2 | Version Compare 仅对比 V1 字段，无法感知新模块变化 | 对比报告不完整 | E/F 用 Snapshot 对比 |

---

## 13. 后续阶段建议

| 编号 | 名称 | 内容 |
|------|------|------|
| **K8-5-21-E** | 前端页面结构设计冻结 | 10 模块表单 layout + 字段分组 + UI 规范 + 前端 types 对齐 Snapshot |
| **K8-5-21-F** | 分阶段实现计划冻结 | C→D→E 实现顺序、验收标准、Mock 更新计划、风险回归 |
| **K8-5-21-G** | 阶段性代码实现（后续拆分） | NormalizeResumeToSnapshot() 实现 + Agent buildPromptValues 改造 + Prompt 更新 + Mock 更新 |

---

## 14. 构建 / API / 测试结果

| 验收项 | 结果 |
|--------|------|
| `npm run build` | ✅ 通过 (1.91s) |
| `GET /api/jobs?page=1&page_size=5` | ✅ 200 |
| `GET /api/jobs/10001` | ✅ 200 |
| `GET /api/jobs/facets` | ✅ 200 |
| `GET /api/resumes` | ✅ 401（后端保护正常） |
| `go test ./...` | ✅ 全部通过 |

---

## 15. 未修改范围确认

```text
未修改源码
未修改样式
未修改前端页面
未修改后端
未修改数据库
未修改 SQL
未修改接口
未修改 .env
未修改 LLM 配置
未修改 Prompt / Schema 文件
未新增字段
未新增接口
未修改 Agent
未修改 Mock
未安装依赖
未迁移目录
```

---

## 16. Git 约束确认

```text
本次未执行任何 Git 命令，也未提交代码。
```
