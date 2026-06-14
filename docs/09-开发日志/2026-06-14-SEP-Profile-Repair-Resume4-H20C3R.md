# H20-C-3R：SEP Profile Repair & Resume 4 Recovery

> 日期：2026-06-14
> 类型：SEP 修复
> 编号：H20-C-3R
> 前置：H20-C-3（SEP Asset Warmup）

## 背景

H20-C-3 补齐到 8/9 时，resume 4 因 Hy3 输出 `tags[10].evidence_text` 空被 validator 拒绝，整份 profile 失败。

## 根因

`callLLMForProfile()` → `SanitizeLLMOutput()` → `ValidateSemanticProfileOutput()`：validator 对任何空 `evidence_text` 的 tag 返回 error，导致整份 profile 失败。无 repair 或 drop 机制。

## 修复

### 1. `repairSemanticProfileTags()` 新增

- 位于 `semantic_profile_profiler.go`
- 在 sanitize 后、validate 前执行
- 对每个 `evidence_text` 空的 tag：
  - 如果 `Explanation` >= 30 字符 → 用 Explanation 作为 evidence_text fallback
  - 否则 → drop 该 tag
- 记录 `dropped` count + `repair warnings`
- 单个坏 tag 不导致整份 profile 失败

### 2. `callLLMForProfile` 调用链

```
Sanitize → Repair → Validate → persist
```

### 3. `--target-id` 支持

- `BackfillOptions.TargetID` 新增
- CLI `--target-id=4` 只处理指定 resume/job
- `Makefile sep-backfill-resume4` 快捷目标

## 执行结果

```
target=resume target-id=4
scheduled=1 generated=1 gen_failed=0 skipped=0
duration=18s
tags=16
```

无 repair warning（Hy3 第二次输出正常），直接 validated。

## DB 复查

| 类型 | 总数 | verified |
|---|---|---|
| resume | 9 | **9** |
| job | 46 | **46** |

## Dry-run 幂等确认

```
resume: total=9 ready=9 missing=0
job:    total=46 ready=46 missing=0
```

## 构建

| 项目 | 结果 |
|---|---|
| go build | ✅ PASS |
| go vet | ✅ PASS |
| go test | ✅ PASS |
| npm build | ✅ 1.97s |

## H20-C-2 未回退

✅ `fastRecommendWithSEP` 仍只调 `EnsureXxxAsync`，无同步生成。
