# H20-C-1：SEP Fast Latency Audit 只读审计

> 日期：2026-06-14
> 类型：只读审计（未修改任何代码）
> 编号：H20-C-1

## 审计结论

**Fast 慢的原因是 Fast Recommend 在同步调用 SEP profile 生成（Hy3 LLM），违反了 SEP 产品冻结原则。**

## 一、代码入口审计

| 链路 | 入口文件/函数 | 是否调用 SEP | 是否同步生成 | 是否触发 LLM |
|---|---|---|---|---|
| Fast 推荐 | `recommendation_service.go:849 fastRecommendWithSEP()` | ✅ `EnsureResumeProfile` + `EnsureJobProfile` | ✅ **同步** | ✅ **Hy3 high reasoning** |
| Fast 匹配 | `match_service.go:196 Analyze()` → `checkFastSEPContext()` | ✅ 只读 `FindLatestByTarget` | ❌ 不生成 | ❌ 不触发 |
| Deep 推荐 | `recommendation_service.go` → quality runner | ✅ 读 SEP tags | ❌ | 取决于 LLM runner |
| Deep 匹配 | `match_service.go:237 runScreeningMatchEntry()` | ✅ 读 SEP | ❌ | 取决于 Hy3 pipeline |

## 二、关键发现：两条路径的 SEP 策略严重不一致

| 维度 | Fast Recommend | Fast Match |
|------|---------------|------------|
| SEP 缺失时 | **同步生成 (Hy3 LLM)** | **返回 profile_pending** (不等待) |
| Tag fallback | 有（回退到 TagEvidenceService） | **无**（H18-C 已移除） |
| Gate 入口 | `SemanticProfileGate`（主动生成） | `checkFastSEPContext`（只读检查） |
| P0 风险 | ✅ 存在 | ❌ 不存在 |

## 三、数据库 Profile 状态

| 项目 | 结果 |
|---|---|
| 简历总数 | 9 |
| 岗位总数 | 46 |
| 有 SEP 的简历 | 2（resume 1, 8） |
| 无 SEP 的简历 | 7（resume 2-7, 9） |
| 有 SEP 的岗位 | 46/46（全部 verified） |
| SEP status 分布 | resume=2 verified, job=46 verified |
| 重复 SEP | 无 |
| pending/building | 0 |
| failed | 0 |

**关键**：7/9 的简历（78%）没有 SEP profile，Fast Recommend 会触发同步 LLM 生成。

## 四、同步阻塞调用链（拆解）

```
Fast Recommend 请求
  → RecommendationService.Recommend()                     [line 131]
    → profileSvc.EnsureProfile(type="fast")                [line 150]  ← 旧系统，规则提取，快
    → recommendCore()                                      [line 198]
      → fastRecommendWithSEP()                              [line 849]
        → semanticGate.EnsureResumeProfile(ctx, resumeID)  [line 857]  ← 🔴 同步阻塞
          → ensureProfile()                                [line 86]
            → GetVerifiedProfile() → 无缓存?               [line 108]
            → profiler.GenerateResumeSemanticProfile()    [line 153]  ← 🔴 Hy3 LLM 调用 (10-60s)
        → for each job: EnsureJobProfile()                 [line 889]  ← 46个 job 全部缓存命中（快）
```

## 五、注释与实现矛盾

```go
// recommendation_service.go:848
// No LLM call in recommendation — SEP profiles are pre-generated via Hy3.
```

**这个注释在当前实现下是 false。** `EnsureResumeProfile` 在没有缓存时会同步调用 Hy3 LLM。

## 六、前端 Loading 审计

| 页面 | loading 状态 | 是否只等 Fast | 是否误等 Deep | 是否轮询 profile |
|---|---|---|---|---|
| 岗位推荐 Fast | `generateFast()` 最多 8 次轮询 (1.8s 间隔) | ✅ 独立 Fast API | ❌ 不等待 Deep | ✅ 轮询 `profileStatus !== "ready"` |
| 岗位匹配 Fast | `loadLatestMatchViewModel()` 无轮询 | ✅ 独立 Fast API | ❌ 不等待 Deep | ❌ 无轮询 |

**前端未误等 Deep**，但前端的 `profileStatus` 轮询关注的是旧 ResumeProfile 系统（规则提取），而非 SEP。旧 profile ready 后，后端在 `fastRecommendWithSEP()` 中才开始同步生成 SEP。

## 七、问题分级

| 等级 | 问题 | 证据 | 影响 |
|---|---|---|---|
| **P0** | Fast Recommend 同步生成 SEP | `semantic_profile_gate.go:153 GenerateResumeSemanticProfile()` 在 Fast 路径被同步调用 | 7/9 简历首次 Fast Recommend 需等 10-60s LLM |
| **P0** | `EnsureResumeProfile()` 注释与实际行为不符 | `recommendation_service.go:848` 注释说 "No LLM call"，但实际会触发 | 误导性代码文档 |
| **P1** | Fast Recommend 与 Fast Match SEP 策略不一致 | Recommend 用 Gate（主动生成），Match 用 `checkFastSEPContext`（只读） | 维护混乱，行为不一致 |
| **P1** | 缺少 profile ready 检查在 Fast Recommend 入口 | Recommend 只检查旧 ResumeProfile ready，不检查 SEP ready | 旧 profile ready 后仍被 SEP 生成阻塞 |
| **P2** | 没有 `updated_at` 被 Fast 请求刷新的证据 | DB 中 SEP `updated_at` 稳定在 6/12-6/14 | 暂未发现重复生成 |
| **P2** | 日志不可观测 | 无性能埋点，无法判断每次耗时分布 | 难以定位生产问题 |

## 八、修复建议（H20-C-2）

1. **`fastRecommendWithSEP()` 改为只读**：用 `semanticRepo.FindLatestByTarget()` 替代 `semanticGate.EnsureResumeProfile()`
2. **SEP 缺失时快速返回**：返回 `profile_pending` 状态，不被 `recommendCore` 的 `recommendCore` 吞掉
3. **后台异步补齐**：用 `go func()` 触发 `EnsureResumeProfile()` 但不等待
4. **统一 Recommend 和 Match 的 SEP 策略**：都使用 `checkFastSEPContext` 模式
5. **删除/修正 false 注释**：`recommendation_service.go:848`
6. **增加耗时埋点**：在 `fastRecommendWithSEP()` 中记录各阶段耗时
