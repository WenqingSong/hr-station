# SEP-Profile-Persistence-And-PreGeneration-Audit-H13：SEP Profile 持久化与预生成链路审计

## 审计结论：Profile 持久化正常，缺失预生成

| 项目 | 结论 |
|------|------|
| 是否完成审计 | ✅ |
| resume_id=3 profile 是否存在 | ✅ ID 78，verified，tags=15 |
| profile 是否每次重建 | ❌ 首次构建后持久化，后续 45ms |
| 是否存在保存后生成 | ❌ 无 |
| 是否存在启动预生成 | ❌ 无 |
| 是否同步构建 | ✅ `EnsureResumeProfile` 同步等待 LLM |
| 根因 | **首次访问时 profile 缺失 → 同步构建 30-60s → 持久化 → 后续 45ms** |
| H14 修复 | 简历保存后异步预生成 profile |

---

## 数据库表

| 表 | 作用 | 数据 |
|------|------|:---:|
| `semantic_profiles` | SEP 语义画像（简历/岗位） | resume: 7 verified + 12 stale；job: 46 verified |
| `semantic_profile_tags` | 画像标签 | 随 profile |
| `resume_profiles` | 旧版简历 profile | 有 |
| `job_profiles` | 旧版岗位 profile | 有 |

## resume profile 数据

| resumeId | verified 数 | stale 数 | 最近 verified updated | 后续访问 |
|:---:|:---:|:---:|------|:---:|
| 3 | 1 (ID 78) | 5 | 00:42:08 | **45ms** ✅ |
| 4 | 1 (ID 69) | 1 | 00:00:15 | 45ms ✅ |
| 1 | 1 (ID 67) | 3 | 23:48:20 | 45ms ✅ |
| 8 | 1 (ID 80) | 3 | 00:53:01 | 45ms ✅ |

## job profile 数据

46 jobs, all verified. ✅

## 推荐接口 profile readiness

| 项目 | 实现 | 问题 |
|------|------|------|
| fast 读 profile | `EnsureResumeProfile(ctx, resumeID)` | ✅ 读优先 |
| profile missing | **同步 LLM 构建**（30-60s） | ❌ 阻塞用户 |
| profile stale | 先 mark stale 再重建 | ⚠️ 每次 stale→rebuilt |
| status != ready | 同步等待完成 | ❌ |
| 是否 fallback | SEP fail → `fastRecommendWithTags` | ✅ |

## resume_id=3 连续请求

| run | time | code | gen | profile | 结论 |
|:---:|:---:|:---:|------|------|------|
| 1 | 45ms | 0 | fast_tag_engine | ready | profile 已持久化 ✅ |
| 2 | 44ms | 0 | fast_tag_engine | ready | ✅ |
| 3 | 45ms | 0 | fast_tag_engine | ready | ✅ |

## 根因

Profile **持久化正常**——首次构建后 DB 写入 verified 记录，后续 45ms 秒级复用。问题在于**纯滞后（无预生成）**：简历创建/导入后没有异步触发 SEP profile 构建，首次推荐访问时 `EnsureResumeProfile` 同步等待 LLM 完成（30-60s）。12 条 stale 记录说明 hash 不一致导致多次重建后才得到稳定 verified。

## H14 修复

| 修复项 | 优先级 |
|------|:---:|
| 简历保存后异步生成 SEP profile | P1 |
| 启动预热：扫描无 verified profile 的简历异步生成 | P1 |
| 推荐接口 profile missing 返回明确状态 | P2 |
