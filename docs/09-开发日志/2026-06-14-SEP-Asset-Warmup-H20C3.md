# H20-C-3：SEP Asset Warmup

> 日期：2026-06-14
> 类型：SEP 资产补齐与预热
> 编号：H20-C-3
> 前置：H20-C-1（审计）/ H20-C-2（P0修复）

## 一、背景

H20-C-1 审计发现 9 份简历中仅 2 份有 SEP profile，H20-C-2 将 Fast Recommend 改为非阻塞后，大部分简历会降级到 TagEvidenceService。本任务补齐缺失的 SEP profiles 使其成为稳定资产。

## 二、新增能力

### 2.1 SemanticProfileBackfillService

- **文件**：`server/internal/service/semantic_profile_backfill.go`
- **核心方法**：
  - `Scan(ctx, targetType)` — 只读扫描，返回所有 profile 覆盖状态
  - `BackfillMissing(ctx, opts)` — 扫描 + 生成缺失 profile

### 2.2 控制参数

| 参数 | 默认 | 说明 |
|---|---|---|
| `--target` | all | resume / job / all |
| `--dry-run` | false | 只扫描不生成 |
| `--limit` | 0 (无限制) | 最多生成 N 个 |
| `--concurrency` | 1 | 并发数（最多 2） |
| `--rebuild-stale` | false | 是否重建 stale profile |

### 2.3 幂等性

- ready/verified profile 永远 skip
- content_hash 通过 `GetVerifiedProfile` 检查
- 可安全复跑，不会重复生成

## 三、实际执行结果

### Dry-run（执行前）

```
resume: total=9 ready=2 missing=7 stale=0 failed=0
job:    total=46 ready=46 missing=0 stale=0 failed=0
```

### Execute

```
target=all concurrency=1 duration=2m58s
scheduled=7 generated=6 gen_failed=1 skipped=48
```

- resume 4 因 Hy3 输出验证失败（`evidence_text empty for tag "distributed_system_design"`）
- 其余 6 份简历全部成功生成
- 48 个 profile（2 ready resume + 46 ready job）全部正确跳过

### DB 复查

| 类型 | 总数 | ready/verified | missing | failed |
|---|---|---|---|---|
| resume | 9 | 8 | 1（resume 4） | 0 |
| job | 46 | 46 | 0 | 0 |

### Dry-run（执行后，验证幂等）

```
resume: total=9 ready=8 missing=1
job:    total=46 ready=46 missing=0
```

## 四、H20-C-2 未回退验证

| 检查 | 结果 |
|---|---|
| `fastRecommendWithSEP` 仍只调 `EnsureXxxAsync`（无同步） | ✅ |
| Fast 请求仍不调用 `GenerateXxxSemanticProfile` | ✅ |
| backfill 命令独立于 Fast 请求链路 | ✅ |

## 五、构建结果

| 项目 | 结果 |
|---|---|
| go build | ✅ PASS |
| go vet | ✅ PASS |
| go test | ✅ PASS |
| npm build | ✅ 1.90s |
