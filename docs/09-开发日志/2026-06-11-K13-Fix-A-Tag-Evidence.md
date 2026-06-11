# K13-Fix-A-Tag-Evidence：Tag Agent 与 Evidence 底座

**日期**: 2026-06-11  
**状态**: ✅ 已完成  
**类型**: 架构修复（第一阶段）

## 1. 执行摘要

建立统一的 Tag Agent + Evidence 底座，为后续 Fast 链路和 Deep Agent 1 提供公共数据层。

## 2. Tag Agent 设计

### 2.1 架构

```
Resume Tag → extractTags() [规则引擎] → ResumeTagProfile
            ├── primary_tags (方向/能力/领域)
            ├── skill_tags (从 structured data)
            ├── project_tags → EvidenceItem[]
            ├── internship_tags → EvidenceItem[]
            ├── education_tags
            └── evidence (skill/project/internship/education/certificate)

Job Tag → extractJobTags() [规则引擎] → JobTagProfile
            ├── primary_tags (方向/技能/要求)
            ├── skill_tags (JD 关键词)
            ├── requirement_tags (学历/经验等)
            ├── risk_tags
            └── evidence (keyword/requirement/responsibility)
```

### 2.2 关键文件

| 文件 | 说明 |
|------|------|
| `server/internal/model/tag_evidence.go` | 数据契约（TagItem, EvidenceItem, ResumeTagProfile, JobTagProfile, JobCandidate） |
| `server/internal/model/job_profile.go` | JobProfile DB model |
| `server/internal/service/tag_evidence_service.go` | TagEvidenceService（约800行） |
| `server/internal/service/tag_evidence_service_test.go` | 7项测试 |
| `server/internal/repository/job_profile_repository.go` | job_profiles CRUD |
| `server/internal/repository/resume_profile_repository.go` | resume_profiles UpsertProfile |

## 3. Resume tags 字段

| 字段 | 类型 | 来源 |
|------|------|------|
| `primary_tags` | []TagItem | extractTags() → buildPrimaryTags() |
| `candidate_tags` | []TagItem | extractTags() 全套 |
| `skill_tags` | []TagItem | resume.ProfessionalSkills 解析 |
| `project_tags` | []TagItem | resume.Projects 结构化 |
| `internship_tags` | []TagItem | resume.Internships 结构化 |
| `education_tags` | []TagItem | resume.Educations |
| `intent_tags` | []TagItem | resume job intention |
| `risk_tags` | []TagItem | extractRiskTags() |
| `evidence` | ResumeEvidence | buildResumeEvidence() |

## 4. Job tags 字段

| 字段 | 类型 | 来源 |
|------|------|------|
| `primary_tags` | []TagItem | extractJobTags() |
| `skill_tags` | []TagItem | JD skill_keywords |
| `requirement_tags` | []TagItem | JD requirements 解析 |
| `risk_tags` | []TagItem | 方向偏差风险 |
| `evidence` | JobEvidence | buildJobEvidence() |

## 5. Evidence 字段

### ResumeEvidence
- `skill_evidence`: 从 ProfessionalSkills 提取
- `project_evidence`: 从 Projects 提取（最多5条）
- `internship_evidence`: 从 Internships 提取（最多5条）
- `education_evidence`: 从 Educations 提取
- `certificate_evidence`: 从 Certificates 提取
- `project_count`: len(Projects)
- `internship_count`: len(Internships)
- `skill_count`: len(split(Skills))

### JobEvidence
- `keyword_evidence`: 从 skill_keywords 提取
- `requirement_evidence`: 从 requirements 提取
- `responsibility_evidence`: 从 description 提取

## 6. RecallJobsByTags

```go
RecallJobsByTags(ctx, resumeID, userID, limit) → []JobCandidate
```

流程：
1. 读取 ResumeTagProfile
2. 遍历所有 JobTagProfile
3. computeTagMatch()：比对 resume tags vs job tags
4. 计算 tag_score、matched_tags、missing_tags、evidence_refs
5. 按 score 降序排序，取前 limit

## 7. Fast 后续使用

```go
rp := tagEvidence.GetResumeTagProfile(ctx, resumeID, userID)
jobs := tagEvidence.RecallJobsByTags(ctx, resumeID, userID, 20)
// → Fast 使用 tag_score + matched_tags 快速推荐
```

## 8. Deep Agent 1 后续使用

```go
ctx := TagEvidenceContext{
    ResumeProfile: rp,
    JobProfiles:   topNJobProfiles,
}
// → Deep Agent 1 基于 tag context + evidence 做候选召回和上下文压缩
```

## 9. 验证结果

| 项目 | 结果 |
|------|:--:|
| resume 8 evidence_json 非 {} | ✅ project_evidence=3, internship_evidence=1 |
| project_count=3, internship_count=1 | ✅ 非 0/0 |
| job 10001 evidence 非 {} | ✅ kw=9, req=5, resp=1 |
| job_profiles backfill 46 个 | ✅ |
| resume_profiles backfill 8 个 | ✅ |
| go test tag_evidence tests 7/7 | ✅ |
| go build + go vet | ✅ |
| 不影响四核心接口 | ✅ |
| 前端 UI 不变 | ✅ |
