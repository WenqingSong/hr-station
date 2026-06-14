# H20-C-4：Fast Latency Acceptance & SEP Strategy Freeze

> 日期：2026-06-14
> 类型：最终验收
> 编号：H20-C-4

## SEP 资产状态

| 类型 | 总数 | verified | missing | pending | failed | dup |
|---|---|---|---|---|---|---|
| resume | 9 | **9** | 0 | 0 | 0 | 0 |
| job | 46 | **46** | 0 | 0 | 0 | 0 |

## Backfill 幂等

```
resume: total=9 ready=9 missing=0
job:    total=46 ready=46 missing=0
→ scheduled=0, generated=0, no Hy3 triggered
```

## Fast Recommend 耗时（resume 4, has SEP）

| Round | HTTP | 耗时 | pipeline | profile |
|---|---|---|---|---|
| 1 | 200 | **0.083s** | sep_ranker | ready |
| 2 | 200 | **0.051s** | sep_ranker | ready |
| 3 | 200 | **0.050s** | sep_ranker | ready |

- generated_by: `fast_tag_engine`（P1：语义不清，pipeline 实际是 sep_ranker）
- 零 Hy3/LLM 调用

## Fast Match 耗时（resume 4, job 10001, has SEP）

| Round | HTTP | 耗时 | generated_by |
|---|---|---|---|
| 1 | 200 | **0.005s** | fast_sep_match_engine |
| 2 | 200 | **0.004s** | fast_sep_match_engine |
| 3 | 200 | **0.004s** | fast_sep_match_engine |

- 零 Hy3/LLM 调用

## 日志证据

Fast 请求期间（13:40:56）仅有 `[FastSEPEngine]` 纯规则日志，零 `[LLM-Hy3]`、零 `[SEP-Gate] generating`。

## H20-C-2 未回退

✅ `fastRecommendWithSEP` 仍只调 `EnsureXxxAsync`，零同步 Hy3。

## 构建

| 项目 | 结果 |
|---|---|
| go build | ✅ PASS |
| go vet | ✅ PASS |
| npm build | ✅ 2.06s |

## SEP 策略冻结

- Fast Recommend: 只读 SEP → sep_ranker；missing → tag fallback；async 补齐
- Fast Match: 只读 SEP → fast_sep_match_engine；missing → profile_pending
- SEP 资产: backfill 补齐 + ready 不重复 + hash 变化才重建 + dry-run 可复跑
- 禁止: Fast 请求同步 Hy3 / 同步生成 SEP / ready 重复生成
