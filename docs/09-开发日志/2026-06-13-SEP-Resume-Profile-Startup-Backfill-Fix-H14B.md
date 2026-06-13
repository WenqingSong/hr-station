# SEP-Resume-Profile-Startup-Backfill-Fix-H14B：启动时补齐历史简历 SEP Profile

## 目标

H14 完成了 create/update 异步触发 SEP profile，但缺少启动 backfill。H14B 补充：后端启动后扫描历史简历，对缺少 verified SEP profile 的简历异步构建。

## 修复详情

### 1. ResumeRepository 新增 ListBackfillCandidates

查询 `raw_text IS NOT NULL AND CHAR_LENGTH(raw_text)>50` 的简历，按 `is_current DESC, id ASC`，limit 10。

### 2. ResumeService 新增 BackfillMissingResumeProfilesAsync

| 项目 | 值 |
|------|------|
| 并发 | 1（semaphore channel） |
| 超时 | 120s per resume |
| recover | goroutine 内 recover |
| 阻断启动 | ❌ 不阻断 |
| 幂等 | ✅ 调用 `sepGate.EnsureResumeProfile`（已有 hash/version skip） |

### 3. main.go 调用

在 `resumeService.SetSemanticGate(sepSemanticGate)` 后立即调用 `BackfillMissingResumeProfilesAsync(context.Background())`。

## 启动验证

```
[REC_TRACE][BE] event=startup_backfill_start
[REC_TRACE][BE] event=startup_backfill_candidate_count count=8
[REC_TRACE][BE] event=startup_backfill_build_done resumeId=2 status=verified
```

8 个候选人，逐个异步构建。

## 性能验证

| resumeId | time | gen | items | 结论 |
|:---:|:---:|------|:---:|:---:|
| 1 | 44ms | fast_tag_engine | 5 | ✅ |
| 3 | 45ms | fast_tag_engine | 5 | ✅ |
| 8 | 46ms | fast_tag_engine | 5 | ✅ |
| 4 | 30s | — | timeout | backfill 正在构建中（预期行为） |

## 修改文件

| 文件 | 修改 |
|------|------|
| `server/internal/repository/resume_repository.go` | ListBackfillCandidates |
| `server/internal/service/resume_service.go` | BackfillMissingResumeProfilesAsync + time import |
| `server/cmd/server/main.go` | 调用 backfill |

## 构建

go build/vet ✅
