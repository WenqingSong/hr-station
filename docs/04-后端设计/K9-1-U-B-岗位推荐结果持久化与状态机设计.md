# K9-1-U-B：岗位推荐结果持久化与状态机设计

> 状态：设计冻结 | 日期：2026-06-07 | 阶段：K9-1-U-B

## 1. K9-1-U-A 诊断结论

| 级别 | 问题 |
|:--:|------|
| P0 | Quality API 对 resume_id=1 返回 pending/空，前端超时显示"推荐失败" |
| P0 | recommendation_results 只存不读，无 latest 接口 |
| P1 | recommendation_results 表缺 generation_tier / generated_by |
| P2 | Quality profile 异步生成无完成通知 |
| P2 | 保存简历后不触发 profile 生成 |

## 2. recommendation_results 表结构扩展

### 2.1 当前结构

```sql
CREATE TABLE recommendation_results (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT UNSIGNED NOT NULL,
  resume_id BIGINT UNSIGNED NOT NULL,
  result_json JSON NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  KEY idx_rec_user_created (user_id, created_at),
  KEY idx_rec_resume (resume_id)
);
```

### 2.2 扩展设计（冻结）

```sql
-- K9-1-U-B: K9-1-U-C 实现阶段执行（开发库 ALTER，不改 init.sql）
ALTER TABLE recommendation_results
  ADD COLUMN generation_tier VARCHAR(20) NOT NULL DEFAULT 'fast' COMMENT 'fast / quality',
  ADD COLUMN generated_by   VARCHAR(50) NOT NULL DEFAULT 'mock_agent' COMMENT 'mock_agent / mock_quality / llm_agent',
  ADD COLUMN status         VARCHAR(20) NOT NULL DEFAULT 'ready' COMMENT 'ready / error',

  ADD INDEX idx_rec_resume_tier_created (resume_id, generation_tier, created_at),
  ADD INDEX idx_rec_user_resume_created (user_id, resume_id, created_at);
```

### 2.3 Model 扩展

```go
type RecommendationResult struct {
    ID             uint64    `db:"id"`
    UserID         uint64    `db:"user_id"`
    ResumeID       uint64    `db:"resume_id"`
    GenerationTier string    `db:"generation_tier"` // "fast" / "quality"
    GeneratedBy    string    `db:"generated_by"`     // "mock_agent" / "mock_quality" / "llm_agent"
    Status         string    `db:"status"`           // "ready" / "error"
    ResultJSON     []byte    `db:"result_json"`
    CreatedAt      time.Time `db:"created_at"`
}
```

## 3. 推荐结果保存策略

### 3.1 Append-Only + Latest

| 场景 | 行为 |
|------|------|
| Fast 生成成功 | INSERT（generation_tier=fast, status=ready） |
| Quality 生成成功 | INSERT（generation_tier=quality, status=ready） |
| 生成失败 | 不保存 |
| 读取最新 | `SELECT ... WHERE resume_id=X AND status='ready' ORDER BY created_at DESC LIMIT 1` |
| 读取最新 fast | 加 `AND generation_tier='fast'` |
| 读取最新 quality | 加 `AND generation_tier='quality'` |

### 3.2 页面进入决策

```text
1. GET /api/jobs/recommend/latest?resume_id=X
2. 如果 has_result=true → 使用 latest 渲染（fast→Fast简化，quality→Deep完整）
3. 如果 has_result=false → POST /api/jobs/recommend generation_tier=fast，成功后保存并展示 Fast
```

### 3.3 Deep 生成决策

```text
1. POST /api/jobs/recommend generation_tier=quality
2. 成功 → INSERT（generation_tier=quality），前端切换 Deep 完整展示
3. 再次进入 → GET /latest 返回 quality
```

## 4. API 设计

### 4.1 GET /api/jobs/recommend/latest

```json
// has_result=true
{
  "code": 0, "message": "ok",
  "data": {
    "has_result": true,
    "result_id": 123,
    "resume_id": 1,
    "generation_tier": "quality",
    "generated_by": "mock_quality",
    "status": "ready",
    "created_at": "2026-06-07T13:37:39+08:00",
    "profile_type": "quality",
    "profile_dominant_profile": "algorithm_research",
    "result": { "recommended_jobs": [...], "recommendation_summary": "..." }
  }
}

// has_result=false
{
  "code": 0, "message": "ok",
  "data": { "has_result": false }
}
```

后端实现：`RecommendationService.GetLatest(ctx, userID, resumeID)`。

### 4.2 POST /api/jobs/recommend（保持现有）

语义冻结：
- 负责生成 + 保存
- 成功后 `INSERT` recommendation_results（含 generation_tier / generated_by / status）
- 不负责读取历史

## 5. RecommendationResponse 新增字段

```go
type RecommendationResponse struct {
    // ... existing ...
    ResultID       uint64 `json:"result_id,omitempty"`       // 保存后的 recommendation_results.id
    GenerationTier string `json:"generation_tier,omitempty"` // "fast" / "quality"
}
```

## 6. Quality profile pending 处理

### 6.1 问题

quality profile stale → EnsureProfile → pending → 前端超时 → "推荐失败"

### 6.2 冻结处理方案

| 状态 | 后端行为 | 前端行为 |
|------|---------|---------|
| profile.ready | 返回推荐结果 | 正常展示 |
| profile.pending | 返回 HTTP 200 + profile_status=pending + jobs=[] | 显示"正在生成深度画像"，进入 deep_profile_pending |
| profile.failed | 返回 HTTP 200 + profile_status=failed + message | 显示"深度画像生成失败"，可重试 regenerate |
| profile.pending 超时 | — | 显示"深度画像生成较慢 · 已保留快速结果" |
| API error (500/network) | — | 显示"深度推荐生成失败 · 已保留快速结果" |

**禁止**：pending 不直接变成 deep_error，不显示"推荐失败"。

### 6.3 前端新增状态

```ts
type RecPhase =
  | "fast_loading" | "fast_done" | "fast_error"
  | "deep_loading" | "deep_profile_pending" | "deep_done" | "deep_error" | "profile_failed";
```

## 7. ResumeProfile 生成链路

### 7.1 保存简历后

| 行为 | 说明 |
|------|------|
| 生成 fast profile | 同步或快速异步 |
| 标记旧 profile stale | MarkStaleByResumeID |
| 异步生成 quality | 延后执行 |
| 补齐该用户其他缺失 fast profile | 可选，降低用户等待 |

本轮最低要求：**Fast profile 在推荐页进入时保证 ready**。

### 7.2 后续实现建议

- `ResumeService.CreateResume` / `UpdateResume` 后调用 `ResumeProfileService.EnsureProfile(ctx, userID, resumeID, "fast")`
- 启动时可选 backfill：扫描 missing fast profile 并生成
- Quality 只在用户点击 Deep 时触发

## 8. 后续阶段拆分

| 编号 | 名称 | 范围 |
|------|------|------|
| K9-1-U-C | 后端结果持久化与 API 实现 | ALTER 表 + Model/Repo/Service + GET /latest |
| K9-1-U-D | 前端读取 latest + Fast/Deep 展示分层 | 页面进入状态机 + 展示密度差异化 |
| K9-1-U-E | Quality pending 状态与超时文案修复 | deep_profile_pending + 文案优化 |
| K9-1-U-F | 5 类样例持久化与回访验收 | 完整链路验收 |
