# K9-1-H4-B：ResumeProfile 存储与生成后端实现（开发库直改版）

> 日期：2026-06-07 | 状态：已完成

## 1. 设计遵循

基于 K9-1-H4-A 设计冻结，DB-first 模式：不改 init.sql，开发库直接建表。

## 2. DDL

```sql
CREATE TABLE IF NOT EXISTS resume_profiles (
  id BIGINT UNSIGNED AUTO_INCREMENT, resume_id BIGINT UNSIGNED NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  profile_type VARCHAR(20), profile_version VARCHAR(20), status VARCHAR(20),
  dominant_profile VARCHAR(80), primary_tags_json JSON, candidate_tags_json JSON,
  risk_tags_json JSON, evidence_json JSON, profile_summary TEXT,
  source VARCHAR(50), generator VARCHAR(50), error_message TEXT,
  resume_updated_at DATETIME, generated_at DATETIME,
  created_at DATETIME, updated_at DATETIME,
  UNIQUE KEY uk_resume_profile_type (resume_id, profile_type),
  CONSTRAINT fk_resume_profiles_resume FOREIGN KEY (resume_id) REFERENCES resumes(id) ON DELETE CASCADE
);
```

## 3. 实现文件

| 文件 | 说明 |
| -- | -- |
| `model/resume_profile.go` | ResumeProfile + ResumeProfileStatusResponse + ProfileTag |
| `repository/resume_profile_repository.go` | FindByResumeIDAndType, UpsertProfile, MarkStale, CreatePendingIfMissing, UpdateReady, UpdateFailed, ListMissingFastProfiles |
| `service/resume_profile_service.go` | EnsureProfile, GenerateFastProfile, GenerateQualityProfileAsync, BackfillMissingProfiles, MarkStaleOnResumeUpdate, GetProfileStatus, RegenerateProfile + tag extraction |
| HTTP handler | ProfileStatus, ProfileRegenerate (via ResumeByIDDispatcher sub-routes) |
| main.go | Inject ResumeProfileRepo + ResumeProfileService |
| resume_service.go | SetProfileService + triggerProfileGen on Create/Update |

## 4. 验收结果

### 01-算法研究生 (resume_id=7)
- fast_status: ready
- dominant_profile: **algorithm_research** ✅
- primary_tags: 6 tags including algorithm, research_ability, research, data, etc.

### API
- GET /api/resumes/7/profile/status → 200 ✅
- POST /api/resumes/7/profile/regenerate → 200 ✅

### 构建
- go build ✅ go vet ✅ npm build ✅ check-encoding ✅

## 5. 约束
- init.sql 未修改
- 现有 4 份简历未删除
- 未执行 reset-up / Git
