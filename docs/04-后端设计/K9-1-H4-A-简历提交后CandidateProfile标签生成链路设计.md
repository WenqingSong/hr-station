# K9-1-H4-A：简历提交后 CandidateProfile 标签生成链路设计

> 状态：设计冻结 | 日期：2026-06-06

## 1. 背景

### 1.1 当前架构问题

K9-1-H2 / H3 / H4 已确认岗位推荐排序异常的根本原因不在分数，而在架构层面：

1. **标签只在推荐时才临时提取** — CandidateProfile 仅作为推荐 Agent 的内部临时变量，没有持久化
2. **多个 Agent 无法复用同一份画像** — 岗位匹配分析、简历优化、版本对比都要重新解析
3. **推荐链路没有"标签是否已准备好"的状态** — 推荐结果依赖标签但不知道标签状态
4. **用户保存简历后系统没有自动生成画像** — 所有推导都在推荐请求时才计算
5. **后端没有检测"哪些简历缺少画像"的补偿机制**

### 1.2 架构调整方向

**核心理念**：标签画像应在简历保存时就开始生成，不应等到用户点击岗位推荐时才临时分析。

```
修复前：用户点推荐 → 临时 extract tags → 推荐 → 丢失
修复后：保存简历 → buildRawText → gen profile → persist → 推荐/匹配/优化/对比 复用
```

## 2. 整体架构

### 2.1 ResumeProfile 实体

```
resume_profiles (
  resume_id    ── reservation base
  profile_type  ── "fast" | "quality"
  status        ── pending / ready / failed / stale
  dominant_profile ── algorithm_research / backend_engineering / ...
  primary_tags_json ── JSON array
  candidate_tags_json ── full tag set JSON
  risk_tags_json     ── JSON array
  evidence_json      ── evidence map
  profile_summary    ── AI generated summary
  source       ── "rule" | "agent" | "merged"
  generator    ── "rule_extractor" | "mock_agent" | "llm_agent"
  error_message
  resume_updated_at  ── 用于检测 stale
  generated_at
  profile_version    ── schema version
)
```

### 2.2 两种 Profile 分型

| 维度 | Fast Profile | Quality Profile |
| -- | -- | -- |
| 用途 | 快速推荐（Fast推荐） | 深度推荐/匹配分析/优化/对比 |
| 触发时机 | 简历保存后立即同步 | 简历保存后异步 |
| 生成方式 | 纯规则抽取 | Agent 结构化抽取 + Judge |
| 生成耗时 | < 300ms | 异步，无上限 |
| source | rule | agent / merged |
| 失败影响 | 推荐入口阻塞 | Fast 仍可用 |

### 2.3 推荐链路消费 Profile

```
岗位推荐入口
→ EnsureProfile(resume_id, "fast")
→ status=ready  → 使用 profile tags 推荐
→ status=pending → 返回 profile_generating（前端 loading）
→ status=missing → 创建 pending + 异步生成 + 返回 profile_generating
→ status=failed  → 返回 profile_failed（前端可重试）
→ status=stale   → 触发重新生成 + 返回 profile_generating
```

## 3. Profile 状态机

```
                    ┌─────────┐
                    │ missing │ ← 从未生成过
                    └────┬────┘
                         │ EnsureProfile / 简历保存
                    ┌────▼────┐
                    │ pending │ ← 已创建任务，正在生成
                    └────┬────┘
                    ╱         ╲
             成功  ╱           ╲ 失败
           ┌───▼───┐      ┌───▼───┐
           │ ready │      │ failed │ ──→ 可重试
           └───┬───┘      └───────┘
               │ 简历更新
          ┌────▼────┐
          │  stale  │ ──→ 触发重新生成 → pending
          └─────────┘
```

### 3.1 状态含义

| 状态 | 含义 | 前端行为 |
| -- | -- | -- |
| missing | DB 无记录 | 触发生成 + loading |
| pending | 正在生成 | loading + 轮询 |
| ready | 生成完成 | 使用画像 |
| failed | 生成失败 | 错误提示 + 重试按钮 |
| stale | 简历已更新，画像过期 | 触发生成 + loading |

## 4. 简历保存触发链路

### 4.1 Create Resume

```
POST /api/resumes
  → CreateRelational(ResumeRelationalRequest)
    → repo.Create(resume)              // INSERT resumes
      → buildRawText(resume)           // 生成 raw_text
      → INSERT resume_educations/projects/...
    → go profileSvc.EnsureProfile(ctx, resume.ID, "fast")    // 同步规则生成
    → go profileSvc.GenerateQualityAsync(ctx, resume.ID)     // 异步 agent 生成
    → return ResumeResponse（不等待 quality）
```

**同步/异步策略**：

- Fast Profile：同步规则生成，控制在 300ms 内，保证简历保存后立即可用
- Quality Profile：异步 goroutine 生成，不阻塞保存 API

### 4.2 Update Resume

```
PUT /api/resumes/:id
  → UpdateRelational(...)
    → repo.Update(resume) + rebuild raw_text
    → profileSvc.MarkStale(ctx, resume.ID)          // 旧 profile → stale
    → go profileSvc.GenerateFast(ctx, resume.ID)    // 重新生成 fast
    → go profileSvc.GenerateQualityAsync(ctx, resume.ID)  // 重新生成 quality
    → return ResumeResponse
```

### 4.3 Delete Resume

```
DELETE /api/resumes/:id
  → repo.Delete(resume)
  → resume_profiles ON DELETE CASCADE 自动清理
```

## 5. 异步生成与防重机制

### 5.1 防重规则

采用 **UNIQUE(resume_id, profile_type)** + **status 状态判断**：

```go
func (s *ProfileService) EnsureProfile(ctx, resumeID, profileType) (*ProfileStatus, error) {
    profile := s.repo.FindByResumeIDAndType(resumeID, profileType)

    switch profile.Status {
    case "ready":
        if profile.ResumeUpdatedAt == resumes.UpdatedAt {
            return "ready", nil    // 未过期，直接返回
        }
        // 过期 → 标记 stale
        s.repo.UpdateStatus(resumeID, profileType, "stale")

    case "pending":
        return "pending", nil      // 已在生成中，不重复启动

    case "failed":
        // 允许重试
    }

    // 创建 pending 记录，启动生成
    s.repo.Upsert(resumeID, profileType, "pending", ...)
    go s.generate(profileType, resumeID)
    return "pending", nil
}
```

### 5.2 并发安全

- `UNIQUE(resume_id, profile_type)` 保证不会创建重复 pending
- `Upsert` 原子操作，后到的请求看到已有 pending 直接返回
- 异步 goroutine 完成后只更新 `status`，不创建新记录

## 6. 缺失 Profile Backfill 机制

### 6.1 触发时机

1. **后端启动时**：扫描 resumes 表，找出缺少 fast_profile 的记录，异步生成
2. **用户访问推荐页时**：针对目标 resume_id 执行 EnsureProfile
3. **用户访问简历列表时**：可选展示 profile_status，异步补齐缺失

### 6.2 实现

```go
func (s *ProfileService) BackfillMissing(ctx context.Context) error {
    // 1. 查询所有已有 profile 的 resume_id
    profiledIDs := s.repo.FindDistinctResumeIDsWithProfileType("fast")
    // 2. 查询所有简历
    allResumes := s.resumeRepo.FindAll(ctx)
    // 3. 找出缺失的
    for _, r := range allResumes {
        if !contains(profiledIDs, r.ID) {
            go s.GenerateFast(ctx, r.ID)
        }
    }
    return nil
}
```

**注意**：
- Backfill 异步执行，不阻塞启动
- 每用户最多同时生成 N 份（限流）
- 单份失败不影响其他

## 7. 数据库设计

### 7.1 resume_profiles 表

```sql
CREATE TABLE `resume_profiles` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `resume_id`       BIGINT UNSIGNED NOT NULL,
  `user_id`         BIGINT UNSIGNED NOT NULL,

  `profile_type`    VARCHAR(20)  NOT NULL COMMENT 'fast / quality',
  `profile_version` VARCHAR(20)  NOT NULL DEFAULT 'v1',
  `status`          VARCHAR(20)  NOT NULL COMMENT 'pending/ready/failed/stale',

  `dominant_profile`   VARCHAR(80)  NOT NULL DEFAULT '',
  `primary_tags_json`  JSON         NOT NULL,
  `candidate_tags_json` JSON        NOT NULL,
  `risk_tags_json`     JSON         NOT NULL,
  `evidence_json`      JSON         NOT NULL,
  `profile_summary`    TEXT         NULL,

  `source`       VARCHAR(50)  NOT NULL DEFAULT '' COMMENT 'rule/agent/merged',
  `generator`    VARCHAR(50)  NOT NULL DEFAULT '' COMMENT 'rule_extractor/mock_agent/llm_agent',
  `error_message` TEXT         NULL,

  `resume_updated_at` DATETIME NULL COMMENT '简历更新时间，用于检测 stale',
  `generated_at`      DATETIME NULL,
  `created_at`        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_resume_profile_type` (`resume_id`, `profile_type`),
  KEY `idx_profile_user_status` (`user_id`, `status`),
  CONSTRAINT `fk_resume_profiles_resume` FOREIGN KEY (`resume_id`) REFERENCES `resumes`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='简历画像表';
```

### 7.2 与现有表关系

- `resume_profiles` 是独立派生表，不影响 `resumes` 主表结构
- `ON DELETE CASCADE` 保证简历删除时画像自动清理
- 无需修改 `recommendation_results` 表

## 8. 推荐接口适配

### 8.1 新增 profile_status 字段

POST /api/jobs/recommend 响应增加 `profile_status`：

```json
{
  "code": 0,
  "data": {
    "resume_id": 1,
    "profile_status": "ready",      // ready / generating / failed / missing
    "profile_type": "fast",          // fast / quality
    "profile_message": "",           // 当 status != ready 时有用
    "recommended_jobs": [...]
  }
}
```

### 8.2 status 非 ready 时

当 profile 未就绪，不返回推荐结果，由前端 loading / retry。

| profile_status | HTTP code | 行为 |
| -- | -- | -- |
| generating | 200 | 前端轮询，1.5-2s 间隔，最多 15s |
| failed | 200 | 前端显示错误 + 重试按钮 |
| missing | 200 | 后端触发 EnsureProfile，返回 generating |

**不推荐用 HTTP 500 表示 pending**，这是正常业务状态。

## 9. Profile API

### 9.1 GET /api/resumes/:id/profile/status

获取简历画像状态：

```json
{
  "resume_id": 1,
  "fast_status": "ready",
  "quality_status": "pending",
  "fast": {
    "dominant_profile": "algorithm_research",
    "primary_tags": ["algorithm", "research_ability", "modeling_ability"],
    "generated_at": "2026-06-06T10:30:00Z"
  },
  "quality": null
}
```

### 9.2 POST /api/resumes/:id/profile/regenerate

手动触发重生成：

```json
{
  "profile_type": "fast"
}
```

## 10. 前端状态设计

### 10.1 岗位推荐页进入

```
/recommend?resumeId=N
  → 请求推荐接口
  → 后端返回 profile_status=generating
  → 前端全局 loading
```

**loading 文案**：

> 正在生成简历画像
> 系统正在分析你的教育经历、项目经历、实习经历和技能标签。
> 完成后将自动生成岗位推荐。

**轮询逻辑**：
- 每 1.5-2s 重新请求推荐接口
- 最多轮询 15s
- 超时 → "画像生成时间较长，可稍后重试"
- 返回 ready → 进入推荐结果展示

### 10.2 Fast 加载 → Quality 加载

```
Fast ready → 展示快速推荐 → 用户点击"生成深度推荐"
  → 后端返回 quality profile_status=generating
  → 全局毛玻璃 loading
  → 保留 Fast 结果在遮罩下可见
  → quality ready → 全量替换
```

### 10.3 Profile failed

```
profile_status=failed
  → "简历画像生成失败"
  → 按钮："重试生成画像" / "返回编辑简历"
  → retry → POST /api/resumes/:id/profile/regenerate
```

## 11. Fast / Quality 生成方式

### 11.1 Fast Extractor（规则）

复用并增强当前 recommendation_mock.go 中已有的标签提取逻辑：

- 来源：education / project / internship / skills / additional_info / raw_text
- 方法：规则匹配 + confidence 计算
- 输出：CandidateTag JSON + dominant_profile
- 耗时目标：< 300ms

### 11.2 Quality Extractor（Agent）

未来 LLM Agent 实现，当前 Demo 可先复用规则版：

- 输入：raw_text + 结构化字段
- 方法：JSON Schema 约束输出
- 输出：tags + evidence + profile_summary
- 生成方式：异步 goroutine

### 11.3 Merge 策略

如果 fast 和 quality 都存在：

- 推荐：quality 优先
- 解析：fast 作为 fallback
- 冲突：quality + evidence 优先，规则做安全校准

## 12. 边界情况

| 边界 | 处理 |
| -- | -- |
| 简历刚保存，用户立刻进推荐页 | EnsureProfile → pending → 前端 loading |
| 简历更新后旧标签过期 | MarkStale + 触发重新生成 |
| 用户多次点击保存 | 只保留 latest，通过 resume_updated_at 校验 |
| 多个请求同时触发生成 | UNIQUE + status=pending 防重 |
| Fast profile 生成失败 | 标记 failed，前端可重试，不使用空结果推荐 |
| Quality 生成失败 | 标记 failed，Fast 仍可用 |
| 简历删除 | CASCADE 清理 profile |
| 历史简历没有 profile | BackfillMissingProfiles 异步补齐 |
| 多份简历中部分有/部分无 | 推荐页按 resume_id 独立 EnsureProfile |
| Profile 版本升级 | profile_version 不一致 → stale → 重新生成 |

## 13. 后续阶段边界

### K9-1-H4-B：ResumeProfile 存储与生成后端实现

| 允许改 | 不做 |
| -- | -- |
| init.sql + resume_profiles 表 | 不改前端 |
| model/resume_profile.go（新建） | 不接 LLM |
| repository/resume_profile_repo.go（新建） | 不重写推荐 |
| service/profile_service.go（新建） | |
| 简历 create/update 触发 profile | |
| EnsureProfile / Backfill / MarkStale | |
| 推荐接口返回 profile_status | |

### K9-1-H4-C：岗位推荐改为消费 ResumeProfile

| 范围 |
| -- |
| recommendation_service 读取 resume_profiles |
| profile ready → 使用 tags 推荐 |
| profile pending/failed → 按状态返回 |

### K9-1-I：前端状态接入

| 范围 |
| -- |
| 推荐页识别 profile_generating → loading UI |
| 轮询机制 + 超时处理 |
| profile failed → 重试入口 |
| 简历管理页可选画像状态展示 |
