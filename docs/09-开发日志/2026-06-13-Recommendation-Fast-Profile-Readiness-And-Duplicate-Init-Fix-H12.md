# Recommendation-Fast-Profile-Readiness-And-Duplicate-Init-Fix-H12：修复首次画像等待与重复初始化

## 目标

基于 H11 trace 发现，修复两个核心问题：
1. React StrictMode 导致 init/generateFast 重复执行（每次 2 次）
2. 首次简历画像构建（30-40s）缺乏用户提示

## 修复详情

### 1. 防重 guard

| 项目 | 修复 |
|------|------|
| init 阶段 | `initInFlightRef` + `initializedResumeRef` 双重 guard |
| generateFast | `fastInFlightRef` guard，执行中跳过重复调用 |
| handleFastRecommendation | fastInFlightRef 已执行时 return |
| StrictMode 重复 mount | guard 阻止第二次全链路调用 |

新增 trace 事件：`init_skip_in_flight`、`init_skip_already_done`、`generate_fast_skip_in_flight`。

### 2. 画像构建提示

| 阶段 | 文案 |
|------|------|
| 0-3s | "正在快速筛选岗位…" |
| 3-15s | "首次正在建立简历语义画像…首次分析可能需要 30-60 秒…" |
| >15s | "仍在建立简历语义画像…" |

请求成功后自动清除 timer 和 hint。

### 3. 后端耗时拆分

新增 `[REC_TRACE][BE]` 事件：`profile_build_start` / `profile_build_done` / `profile_build_fail` / `fast_rank_done`。

---

## 修改文件

| 文件 | 修改 |
|------|------|
| `web/src/app/pages/JobRecommendationPage.tsx` | initInFlightRef + fastInFlightRef + fastLoadingHint + 文案 |
| `server/internal/service/recommendation_service.go` | profile_build + fast_rank trace |

## 构建

| 项目 | 结果 |
|------|:---:|
| npm build | ✅ (1.90s) |
| go build/vet | ✅ |
