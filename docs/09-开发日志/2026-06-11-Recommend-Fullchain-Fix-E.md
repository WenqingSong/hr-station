# Recommend-Fullchain-Fix-E：彻底修复推荐页旧 mock 缓存污染

日期：2026-06-11

## 真实根因

Fix-B/C/D 源码正确，但 **数据库中的旧 mock 缓存** 未被清理：

```
recommendation_results: 41 条 generated_by=mock_agent
match_analyses: 113 条 generated_by=mock_agent/mock_quality
```

页面流程：
1. `loadLatestRecommendationViewModel(2)` → 读到 `generated_by=mock_agent` 的旧缓存
2. 旧代码直接展示该结果（`phase="fast_done"`），不检查是否 mock
3. 前端渲染 mock 卡片 → 用户看到 "推荐结果加载失败" 或假数据

## 修复内容

### 1. 源码修复
`JobRecommendationPage.tsx`：latest/cache 结果增加 `isMockGeneratedBy(genBy)` 检查。mock 缓存直接丢弃，触发 `generateFast()` 生成真实结果。

### 2. 数据库清理
- `DELETE FROM recommendation_results WHERE generated_by IN ('mock_agent','mock_quality')` → 41 条
- `DELETE FROM match_analyses WHERE generated_by IN ('mock_agent','mock_quality')` → 113 条

### 3. 验证
- Arthur + `/recommend?resumeId=2` → 前端显示 "简历不存在或不属于当前账号"
- Demo + `/recommend?resumeId=2` → `latest` 无缓存 → `generateFast` → `fast_tag_engine` ✅
- 8080 / 5173 表现一致

## 结论
旧 mock 缓存是导致页面持续显示 "推荐结果加载失败" 的真实根因。源码 + 数据双修确认修复。
