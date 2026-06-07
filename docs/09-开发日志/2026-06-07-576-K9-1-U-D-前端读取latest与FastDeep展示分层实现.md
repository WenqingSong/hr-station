# K9-1-U-D：前端读取 latest + Fast/Deep 展示分层实现

> 日期：2026-06-07 | 状态：已完成

## 1. K9-1-U-C-Check 结论

后端持久化通过，latest API 正常，进入前端实现。

## 2. 修改清单

| 文件 | 修改 |
|------|------|
| `web/src/api/agents.ts` | 新增 `getLatestRecommendation` + `LatestRecommendationResponse` 类型 |
| `web/src/api/viewModels.ts` | 新增 `loadLatestRecommendationViewModel` + `LatestRecommendationViewModel` |
| `web/src/app/pages/JobRecommendationPage.tsx` | 重写：latest-first 进入 + Fast/Deep 展示分层 + pending/timeout 处理 |

## 3. latest API 前端接入

| 项 | 结果 |
|----|:--:|
| `getLatestRecommendation(resumeId)` | ✅ 调用 GET /api/jobs/recommend/latest |
| `has_result=true` → 使用 saved result | ✅ 不 POST |
| `has_result=false` → 自动 POST fast | ✅ |
| latest=quality → Deep 展示 | ✅ |

## 4. 页面进入状态机

| 状态 | 行为 |
|------|------|
| loading_latest | GET /latest，有结果→展示，无→生成 Fast |
| fast_done | Fast 简化展示，提示"可点击深度推荐" |
| deep_done | Deep 完整展示 |

## 5. Fast / Deep 展示分层

| 模式 | 左侧卡片 | 中间详情 |
|------|---------|---------|
| Fast | 1 evidence + 无 risk | 一句话总结 + 核心匹配点(2条) + 风险(2条) |
| Deep | 2 evidence + 1 risk | 完整6区块（why/specific/evidence/highlights/risks/improvement） |

## 6. Deep pending / timeout / error 处理

| 状态 | 展示 | 是否保留当前结果 |
|------|------|:--:|
| deep_profile_pending | "正在生成深度画像" | ✅ 保留 |
| deep_timeout | "深度画像生成较慢 · 已保留当前结果" | ✅ 保留 |
| deep_error | "深度推荐生成失败 · 已保留当前结果" | ✅ 保留 |

## 7. 顶部状态栏文案

| 场景 | 文案 |
|------|------|
| Fast | 快速推荐 · 规则推荐 |
| Deep (LLM disabled) | 深度推荐 · 规则兜底 |
| Deep profile pending | 正在生成深度画像 |
| Deep timeout | 深度画像生成较慢 · 已保留当前结果 |
| Deep error | 深度推荐生成失败 · 已保留当前结果 |

## 8. 构建结果

| 验收项 | 结果 |
|--------|:--:|
| npm build | ✅ |
| go build/vet | ✅ |
| check-encoding | ✅ OK |
| "演示数据"残留 | ✅ 0 |

## 9. 未修改范围确认

- ✅ 未修改后端 Go 源码
- ✅ 未修改 server/sql/init.sql
- ✅ 未删除真实简历
- ✅ 未执行 reset-up
- ✅ 未执行 Git 命令
