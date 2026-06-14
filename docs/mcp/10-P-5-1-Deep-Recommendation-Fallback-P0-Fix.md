# P0 修复记录：岗位 Deep 推荐 Fallback 恢复

## 版本历史

| 轮次 | 日期 | 内容 |
|------|------|------|
| R1 | 2026-06-15 | 初版：修复 fallback 元数据语义、503→200 |
| R2 | 2026-06-15 | 生产定位：修复 ctx cancelled 级联超时、safe context、SEP profile 批量生成 guard |

---

## R2: 生产问题定位与修复

### 一、问题现象

1. 生产环境前端 Deep 推荐失败或一直转圈
2. 生产日志显示：
   ```
   [SEP-Profiler] job 10021: failed profile saved, error=chat failed after 28102ms: LLM_TIMEOUT
   [DeepCascade] L0: resume=9 tags=11 jobs=20/46
   [RecommendationService] deep cascade failed: layer1: LLM_TIMEOUT
   [RecommendationService] context cancelled during quality fallback, final mock fallback
   http: superfluous response.WriteHeader call
   method=POST path=/api/jobs/recommend status=500 duration=1m39s
   ```
3. DeepCascade 串行生成大量缺失 job SEP profile（job 10008-10046），每个耗时 18-38s

### 二、根因

#### 根因 #1: DeepCascade 同步生成 46 个岗位 SEP profile → 级联超时

`RunDeepRecommendationCascade` L0 gate 对每个 job 调用 `EnsureJobProfile(ctx, jid)`，当大部分岗位缺失 SEP profile 时，串行生成导致：
- Layer 0 耗时 >120s
- Cascade 内部 timeout 900s 内可能耗尽
- 或 HTTP server context 被取消

#### 根因 #2: Fallback 使用已取消的 request context

`recommendCore` 4 处 `recommendWithMock(ctx, ...)` 在 `ctx.Err()!=nil` 时仍传入原 ctx：
- Line 302: context cancelled before quality tier
- Line 365: LLM fallback  
- Line 382: final cancelled fallback
- Line 397: general fallback

`recommendWithMock` 内 `s.recommendationRepo.Create(ctx, recSnapshot)` 使用已取消 ctx 导致 DB 保存失败。

#### 根因 #3: Match 服务同样存在 ctx cancelled fallback 问题

`analyzeWithMock` 在 fallback 链路中同样使用原 request context 做 DB 持久化。

### 三、修改文件清单

| 文件 | 修改 |
|------|------|
| `server/internal/service/recommendation_service.go` | 新增 `safeFallbackCtx()`、`countMissingJobProfiles()`、`deepCascadeMissingProfileThreshold=5`；所有 fallback 路径使用 safe context；DeepCascade 前 guard 跳过；DB 保存使用独立 ctx；marshal 失败不阻断响应 |
| `server/internal/service/match_service.go` | 新增 `safeMatchFallbackCtx()`；`analyzeWithMock` DB 保存使用 safe context；marshal 失败不阻断响应 |

### 四、修复详细

#### 4.1 safeFallbackCtx()
```go
func safeFallbackCtx() (context.Context, context.CancelFunc) {
    return context.WithTimeout(context.Background(), 15*time.Second)
}
```

#### 4.2 DeepCascade SEP profile 缺失 guard
在「进入 DeepCascade 之前」检查缺失的 job SEP profile 数量。
阈值：`deepCascadeMissingProfileThreshold = 5`
若 missing > 5：直接跳过 cascade → `qualityFallback` label → `deep_fallback_agent`

`countMissingJobProfiles` 使用 **只读** `semanticRepo.FindLatestByTarget` 查询，**不触发**同步 profile 生成。

#### 4.3 所有 fallback 路径使用 safe context
- LLM disabled/client missing → safe ctx
- Context cancelled guard → safe ctx
- LLM fallback → safe ctx  
- Exhausted/cancelled fallback → safe ctx
- General fallback → safe ctx

#### 4.4 recommendWithMock DB save
- marshal 失败 → log + return response（不返回 error）
- DB Create → safeFallbackCtx() 独立 context
- DB 失败 → log + 不阻止响应

#### 4.5 Match serve analyzeWithMock
- marshal 失败 → log + return response
- DB matchRepo.Create → safeMatchFallbackCtx()
- DB 失败 → log，不阻止响应

### 五、验收命令

```bash
# 推荐 Deep fallback（即使 Hy3 超时）
curl -i -X POST http://127.0.0.1/api/jobs/recommend \
  -H "Authorization: Bearer <token>" \
  -H 'Content-Type: application/json' \
  -d '{"resume_id":9,"generation_tier":"deep","limit":5}'

# 期望: HTTP 200, generation_tier=deep, generated_by=deep_fallback_agent

# 匹配 Deep
curl -i -X POST http://127.0.0.1/api/jobs/10001/match \
  -H "Authorization: Bearer <token>" \
  -H 'Content-Type: application/json' \
  -d '{"resume_id":9,"generation_tier":"quality"}'

# Fast 推荐
curl -i -X POST http://127.0.0.1/api/jobs/recommend \
  -H "Authorization: Bearer <token>" \
  -H 'Content-Type: application/json' \
  -d '{"resume_id":9,"generation_tier":"fast","limit":5}'
```

### 六、验收结果

| 项目 | 结果 |
|------|------|
| `go vet ./internal/service/...` | ✅ |
| `go build ./...` | ✅ |
| `go test ./internal/service/...` | ✅ 全部通过 |
| `go test ./internal/model/...` | ✅ 全部通过 |
| 前端 TS（修改文件） | ✅ 无新增错误 |

### 七、回滚建议

如果本次修复引入新问题，可以：
1. 注释掉 `safeFallbackCtx()` 调用，改为使用原始 `ctx`
2. 注释掉 `countMissingJobProfiles` 检查，直接进入 cascade
3. 回退 match_service.go 的 safe context 修改

以上回退均指代码逻辑回退，不涉及数据库操作。

---

## R1: 初始修复



线上部署后发现严重 P0 问题：为了避免 Hy3 Deep 推荐接口超时，我们临时采用了最低保险策略，但副作用是「岗位推荐」功能的 Deep 推荐在前端表现为消失/退化，Deep 区域不再正常展示。

## 根因分析（5 个问题）

### P1: LLM 不可用时返回 503（`recommendation_service.go:252-257`）

```go
// 旧代码：直接返回 503
if isQuality && !s.llmEnabled {
    return nil, apperror.New(..., 503)
}
if isQuality && s.qualityClient == nil {
    return nil, apperror.New(..., 503)
}
```

前端将 503 分类为 `LLM_UNAVAILABLE` → phase `deep_unavailable` → Deep 区域隐藏。

### P2: Mock quality fallback 元数据语义错误（`recommendation_service.go:608-654`）

```go
// 旧代码：quality mock 使用了错误的元数据
generatedBy = "mock_quality"       // → 应使用 deep_fallback_agent
pipelineMode = "quality_pipeline"  // → 应使用 deep_fallback
GenerationTier 未在 response 中设置
meta.GeneratedBy = "mock_agent"    // → 在 DeriveMetaFromResult 中硬编码
```

### P3: recommendWithLLM fallback 用 fast 模式（`recommendation_service.go:311`）

```go
// 旧代码：LLM 失败后传给 recommendWithMock的质量模式参数为 false
return s.recommendWithMock(ctx, ..., false, ...)
```
导致 `generated_by=mock_agent`、`pipeline_mode=fast_rule`。

### P4: 前端将 mock fallback 标记为错误（`JobRecommendationPage.tsx:282-286`）

```tsx
// 旧代码：只有 llm_agent 算合法
if (isMockGeneratedBy(genBy) || metaRaw?.fallback) {
    setPhase("deep_error");
    return;
}
```

### P5: 前端 adapter 不识别 fallback agent（`recommendationAdapter.ts:310-311`, `agentError.ts`）

- `isQuality` 只检查 `mock_quality` / `llm_agent`，不识别 `deep_fallback_agent`
- `getSource()` 不识别 `deep_fallback_agent`

## 修改文件清单

| 文件 | 修改类型 | 说明 |
|------|---------|------|
| `server/internal/model/agent_response.go` | 新增常量 | `GeneratedByDeepFallbackAgent`、`WithDeepFallback()` 方法、fallback 原因/错误码常量 |
| `server/internal/service/recommendation_service.go` | 核心修复 | 503 → Mock fallback、修正 fallback 元数据语义 |
| `web/src/app/utils/agentError.ts` | 兼容修复 | 新增 `isDeepTierGeneratedBy()`、扩展 `isDeep` 检查 |
| `web/src/app/pages/JobRecommendationPage.tsx` | 兼容修复 | fallback 不再标记为 `deep_error` |
| `web/src/adapters/recommendationAdapter.ts` | 兼容修复 | 添加 `deep_fallback_agent` 支持 |

## 修复详细

### 后端修改

#### 1. Model 层常量（`agent_response.go`）

- 新增 `GeneratedByDeepFallbackAgent = "deep_fallback_agent"`
- 新增 `WithDeepFallback(reason, errorCode string)` 方法
- 新增 fallback 原因常量：`llm_api_key_missing`、`llm_client_missing`、`deep_cascade_failed`、`quality_exhausted`、`quality_cancelled`
- 新增错误码常量

#### 2. recommendCore — LLM 不可用改为 fallback

当 `isQuality && !s.llmEnabled` 或 `isQuality && s.qualityClient == nil` 时：
- 旧逻辑：返回 503
- 新逻辑：检查 `enableMockAgent`，若启用则调用 `recommendWithMock(qualityMode=true)` 返回 200 OK fallback
- 仅在 Mock Agent 关闭时才返回 503

#### 3. recommendWithMock — Deep fallback 语义元数据

| 场景 | generated_by | pipeline_mode | generation_tier |
|------|-------------|--------------|-----------------|
| Fast mock | `mock_agent` | `fast_rule` | `fast` |
| Quality+fallback | `deep_fallback_agent` | `deep_fallback` | `deep` |
| Quality 直接 mock | `mock_quality` | `mock_deep_fallback` | `deep` |

meta 使用 `WithDeepFallback()` 方法设置 `generated_by=deep_fallback_agent`、`agent_status=fallback`、`fallback_reason`、`error_code`。

#### 4. recommendWithLLM fallback — qualityMode=true

将 `recommendWithMock(ctx, ..., false, ...)` 改为 `recommendWithMock(ctx, ..., true, ...)`。

### 前端修改

#### 1. agentError.ts

- 新增 `isDeepTierGeneratedBy()` 函数
- `assertRealAgentResult()` 的 `isDeep` 检查扩展为接受 `llm_agent` / `deep_fallback_agent` / `mock_quality`

#### 2. JobRecommendationPage.tsx

- Deep poll 回调中移除 `isMockGeneratedBy(genBy) || metaRaw?.fallback` 检测
- 只拦截 `genBy === "mock_agent"`（纯 fast mock 在 Deep 上下文中是错误）
- `sourceLabel()` 新增 `deep_fallback_agent` 显示

#### 3. recommendationAdapter.ts

- `getSource()` 添加 `deep_fallback_agent → "fallback"`
- `getSourceLabel()` 添加 `deep_fallback_agent → "深度推荐 · 基础模型"`
- `isQuality` 检查新增 `deep_fallback_agent`

## 验收命令

```bash
curl -i -X POST http://127.0.0.1/api/jobs/recommend \
  -H "Authorization: Bearer <token>" \
  -H 'Content-Type: application/json' \
  -d '{"resume_id":9,"generation_tier":"deep","limit":5}'
```

期望响应：
- HTTP 200
- `data.recommended_jobs` 非空
- `data.generation_tier` 为 `deep` 或 `quality`，不能是 `fast`
- `data.pipeline_mode` 不能是 `fast_rule`
- `generated_by` 不能是 `fast_tag_engine`
- `meta.agent_status` 或 `recommended_jobs[*].agent_status` 能体现 fallback
- 前端 Deep 推荐区域正常展示

## 验证结果

| 项目 | 结果 |
|------|------|
| `go vet ./internal/model/... ./internal/service/...` | ✅ 通过 |
| `go build ./...` | ✅ 通过 |
| `go test ./internal/service/... -run "TestRec"` | ✅ 12 个测试全部通过 |
| `go test ./internal/model/...` | ✅ 全部通过 |
| 前端 TypeScript（修改文件） | ✅ 无新增错误 |

## 不受影响的功能

- Fast 推荐（`generation_tier=fast`）逻辑完全不变
- 岗位匹配分析 fast/deep 路径不变
- 简历优化建议不变
- 简历版本对比 fast/quality 不变
- 数据库数据不修改

## 约束检查

| 约束 | 状态 |
|------|------|
| 禁止 Git 操作 | ✅ 未执行任何 git 命令 |
| 禁止修改 .env / .env.prod / 密钥文件 | ✅ 未修改 |
| 不删除数据库数据 | ✅ 未执行 docker compose down -v |
| 不扩大修改范围 | ✅ 仅修改推荐链路相关文件 |
