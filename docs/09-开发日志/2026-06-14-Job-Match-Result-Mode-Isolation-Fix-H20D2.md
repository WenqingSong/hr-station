# H20-D-2：Job Match Result Mode Isolation Fix

> 日期：2026-06-14
> 类型：后端+前端修复
> 编号：H20-D-2
> 前置：H20-D-1（审计）

## 修复目标

Fast/Deep 岗位匹配结果独立保存、独立读取、独立展示。重新进入页面默认 Fast。

## 修改文件

| 文件 | 修改 |
|---|---|
| `server/internal/repository/match_repository.go` | 新增 `FindLatestByResumeAndJobAndTier()` |
| `server/internal/service/match_service.go` | `GetLatestMatchAnalysis` 增加 tier 参数；新增 `normalizeMatchGenerationTier()`；修复 `analyzeWithLLM` generated_by |
| `server/internal/http/match_handler.go` | `Latest` handler 读取 `tier` query param，默认 fast |
| `web/src/api/agents.ts` | `getLatestMatchAnalysis` 新增 `tier=fast` 参数 |
| `web/src/api/viewModels.ts` | `loadLatestMatchViewModel` 新增 `tier=fast` 参数 |
| `web/src/app/pages/JobMatchAnalysisPage.tsx` | 拆分 `fastResult`/`deepResult`/`deepExists`；mounted 默认 fast；新增 `switchToFast`；Deep 存在时显示"查看深度分析"按钮 |

## API 验证

```bash
# Default (tier=fast): id=104 gen_by=mock_agent ✅
# tier=fast:          id=104 gen_by=mock_agent ✅
# tier=quality:       id=89  gen_by=mock_quality ✅
```

## 构建

| 项目 | 结果 |
|---|---|
| go build | ✅ PASS |
| go vet | ✅ PASS |
| npm build | ✅ 1.91s |
