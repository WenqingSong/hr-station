# K8-5-51-B：数据库 Clean 初始化最终结构

> 状态：✅ 已完成 | 日期：2026-06-06

---

## 一、任务结论

数据库 clean 初始化已收口，**init.sql 成为唯一入口**。后端六层架构重构暂停。

---

## 二、15 张目标表

| 表名 | 用途 | 最终记录数 |
|------|------|:--------:|
| `users` | 用户注册/登录 | 2 |
| `resumes` | **简历主表（替代 resume_versions）** | 2 |
| `jobs` | 岗位列表和详情 | 46 |
| `job_locations` | 岗位工作城市关联 | — |
| `job_business_groups` | 岗位事业群关联 | — |
| `job_skills` | 岗位技能关联 | — |
| `job_filter_groups` | 筛选分组元数据 | 6 |
| `job_filter_options` | 筛选选项元数据 | — |
| `job_filter_option_relations` | 岗位-筛选项关联 | — |
| `recommendation_results` | 岗位推荐结果快照 | 1 |
| `match_analyses` | 岗位匹配分析结果 | 1 |
| `resume_compare_results` | **简历对比结果（替代 version_compare_results）** | 1 |
| `resume_optimization_results` | 简历优化结果快照 | 0 |
| `seed_versions` | Seed 版本追踪 | 0 |

**确认不存在：** `resume_versions`, `version_compare_results`

---

## 三、字段命名映射

| 旧 | 新 |
|----|-----|
| `resume_versions` | `resumes` |
| `resume_version_id` | `resume_id` |
| `from_version_id` | `from_resume_id` |
| `to_version_id` | `to_resume_id` |
| `version_compare_results` | `resume_compare_results` |
| `version_no` | **移除** |

---

## 四、修改清单

| 文件 | 修改类型 |
|------|------|
| `server/sql/init.sql` | 重写：15 表 + 完整 seed + 唯一入口 |
| `server/sql/resume_versions.sql` | 追加废弃标记 |
| `server/sql/migration-j1.sql` | 追加废弃标记 |
| `server/sql/migration-j2.sql` | 追加废弃标记 |
| `server/internal/model/resume.go` | ResumeVersion → Resume（别名）+ 移除 VersionNo DB 列 |
| `server/internal/model/recommendation.go` | ResumeVersionID → ResumeID |
| `server/internal/model/match.go` | ResumeVersionID → ResumeID + MatchAnalysisResult 更新 |
| `server/internal/model/version_compare.go` | ResumeCompareResult 新类型 + VersionCompareAnalysis 字段更新 |
| `server/internal/repository/resume_repository.go` | 表名 resumes + 移除 version_no + 新增 FindByUserIDAndID |
| `server/internal/repository/recommendation_repository.go` | resume_id |
| `server/internal/repository/match_repository.go` | resume_id |
| `server/internal/repository/version_compare_repository.go` | ResumeCompareRepository + resume_compare_results |
| `server/internal/service/resume_service.go` | 适配新 repo 方法，VersionNo 仅 API 兼容 |
| 6 个 agent 文件 | 字段引用更新 |
| 4 个 handler 文件 | 字段引用更新 |
| `scripts/dev/mysql.sh` | 简化导入逻辑 |
| `scripts/dev/up.sh` | 更新 core_tables |
| `Makefile` | 更新帮助文本 |
