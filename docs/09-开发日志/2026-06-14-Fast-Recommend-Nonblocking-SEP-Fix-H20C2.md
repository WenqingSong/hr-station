# H20-C-2：Fast Recommend Non-blocking SEP Fix

> 日期：2026-06-14
> 类型：P0 后端修复
> 编号：H20-C-2
> 前置：H20-C-1 SEP Fast Latency Audit

## 问题根因（H20-C-1 审计发现）

`fastRecommendWithSEP()` 在 `recommendation_service.go:857` 同步调用 `semanticGate.EnsureResumeProfile()`，对无 SEP 缓存的简历（7/9）触发同步 Hy3 LLM 生成（10-60 秒），违反了 SEP 核心原则。

## 修复概要

### 修改文件

| 文件 | 修改内容 |
|---|---|
| `server/internal/service/semantic_profile_gate.go` | 新增 `EnsureResumeProfileAsync` / `EnsureJobProfileAsync` / `scheduleAsync`，含 in-flight 去重 |
| `server/internal/service/recommendation_service.go` | 重写 `fastRecommendWithSEP()` 为只读模式，新增 `findReadyResumeSEP` / `findReadyJobSEP` helper，新增 `ErrFastRecommendSEPMissing` sentinel error，修正 misleading 注释 |

### 核心修改

1. **`fastRecommendWithSEP()` 改为只读**：
   - `semanticRepo.FindLatestByTarget()` 替代 `semanticGate.EnsureResumeProfile()`
   - `semanticRepo.FindLatestByTarget()` 替代 `semanticGate.EnsureJobProfile()`
   - SEP missing → async schedule + return `ErrFastRecommendSEPMissing`
   - 零同步 Hy3 调用

2. **异步补齐**：
   - `EnsureResumeProfileAsync(resumeID)` / `EnsureJobProfileAsync(jobID)`
   - 使用 `context.Background()` + 2min timeout，不继承请求 context
   - `go func()` + defer panic recover + 日志

3. **in-flight 去重**：
   - `sync.Mutex` + `map[string]struct{}` 按 `resume:N` / `job:N` 去重
   - 同一 key 已有 in-flight 任务时静默丢弃

4. **注释修正**：删除 "No LLM call in recommendation" 误导性注释，更新为真实描述

## 修复后 Fast Recommend 行为

| 场景 | 行为 |
|---|---|
| resume SEP ready + job SEP ready | 直接 SEP 排名推荐，无 LLM |
| resume SEP missing | 异步 schedule resume SEP 生成 + 立降级到 TagEvidenceService |
| 部分 job SEP missing | 异步 schedule job SEP 生成 + 用已有 job SEP，不足时降级 |
| SEP pending/building/failed | 视为不可用，异步 schedule + 降级 |
| content_hash 变化 | 视为 SEP stale，异步 schedule 重建 + 降级 |

## 构建结果

| 检查 | 结果 |
|---|---|
| go build | ✅ PASS |
| go vet | ✅ PASS |
| go test | ✅ PASS (service 0.008s) |
| npm build | ✅ 1.8s |

## 符合 SEP 冻结原则

| 原则 | 满足 |
|---|---|
| SEP 生成一次后复用 | ✅ ready 缓存命中 |
| Fast 只读 ready SEP | ✅ `FindLatestByTarget` only |
| SEP missing 不阻塞 | ✅ 立即 `ErrFastRecommendSEPMissing` |
| 后台异步补齐 | ✅ `EnsureXxxAsync` + goroutine |
| ready 不重复生成 | ✅ hash/status check |
| hash 变化才重建 | ✅ `ensureProfile` 仍保留 hash 检查 |
