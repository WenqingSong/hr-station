# SEP-Profile-PreGeneration-And-Async-Build-Fix-H14：SEP Profile 预生成与异步构建修复

## 目标

将 SEP 语义画像从"推荐时懒加载"改为"简历保存时预生成"。

## 修复详情

### 1. triggerProfileGen 接入 SEP Gate

**文件**：`server/internal/service/resume_service.go:369-390`

已有 `triggerProfileGen`（在 create/update 时异步调用）新增 SEP `semanticGate.EnsureResumeProfile` 异步生成。当简历保存后，后台 goroutine 同时生成 old-style fast profile + SEP semantic profile。

### 2. ResumeService 注入 SEP Gate

新增 `sepGate` 字段 + `SetSemanticGate` 方法。`main.go` 中在 `sepSemanticGate` 创建后注入到 `resumeService`。

### 3. 推荐接口兜底 trace

推荐接口中 `EnsureResumeProfile` 的 `SemanticProfileGateResult.Reused` 字段区分：
- `Reused=true` → `recommend_profile_ready_before_request`（正常路径）
- `Reused=false` → `recommend_profile_missing_sync_fallback`（兜底构建）

---

## 修改文件

| 文件 | 修改 |
|------|------|
| `server/internal/service/resume_service.go` | sepGate 字段 + SetSemanticGate + triggerProfileGen 接 SEP |
| `server/cmd/server/main.go` | resumeService.SetSemanticGate(sepSemanticGate) |
| `server/internal/service/recommendation_service.go` | recommend sync_fallback trace |

## 构建

| 项目 | 结果 |
|------|:---:|
| go build/vet | ✅ |
| npm build | ✅ (1.95s) |
