# SEP-Backfill-Priority-And-Recommend-NonBlocking-Fix-H14C：SEP Backfill 优先级与推荐接口非阻塞修复

## 目标

H14/H14B 完成异步生成和启动 backfill，但推荐接口仍可能同步等待 profile 构建。H14C 将推荐接口改为非阻塞。

## 修复详情

### 1. Backfill 优先级优化

`ListBackfillCandidates` 排序改为 `ORDER BY is_current DESC, updated_at DESC, id ASC`，优先当前简历和最近更新的简历。

### 2. 推荐接口非阻塞

新增 `SemanticProfileGate.CheckProfileReady`（只读检查，不触发构建）。推荐接口使用此方法替代 `EnsureResumeProfile`：
- verified 存在 → 正常秒级排序
- verified 缺失 → 返回错误 + 后台异步触发构建

### 3. startup backfill 验证

8 candidates, all skip=8 (already verified). 启动 backfill 无重复构建。

## API 验证

| resumeId | time | code | gen | items | 是否 sync | 结论 |
|:---:|:---:|:---:|------|:---:|:---:|:---:|
| 1 | 44ms | 0 | fast_tag_engine | 5 | ❌ | ✅ |
| 3 | 45ms | 0 | fast_tag_engine | 5 | ❌ | ✅ |
| 8 | 44ms | 0 | fast_tag_engine | 5 | ❌ | ✅ |
| 5 | 59ms | 0 | fast_tag_engine | 5 | ❌ | ✅ |

## 修改文件

| 文件 | 修改 |
|------|------|
| `server/internal/service/semantic_profile_gate.go` | CheckProfileReady |
| `server/internal/service/recommendation_service.go` | 非阻塞 CheckProfileReady |
| `server/internal/repository/resume_repository.go` | backfill 排序 updated_at DESC |

## 构建

go build/vet ✅
