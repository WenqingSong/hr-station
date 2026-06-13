# Recommendation-Latency-Trace-Instrumentation-H11：推荐页端到端耗时 Trace 埋点

## 目标

通过前后端统一 traceId 串联，拆分推荐页 30s loading 的真实耗时来源。

## Trace 开关

| 方式 | 命令 |
|------|------|
| URL 参数 | `/recommend?resumeId=3&debugRecommend=1` |
| localStorage | `OFFER_DEBUG_RECOMMEND=1` |

开启后前端 console 打印 `[REC_TRACE][FE]` 日志，后端打印 `[REC_TRACE][BE]` 日志。

## 修改文件

| 文件 | 修改内容 |
|------|------|
| `web/src/app/pages/JobRecommendationPage.tsx` | debug switch + traceFE + 10+ trace 点 |
| `web/src/api/client.ts` | `X-Recommend-Trace-Id` header 注入 |
| `server/internal/http/recommendation_handler.go` | recTraceID + recTraceLog + handler 收/发 trace |
| `server/internal/service/recommendation_service.go` | service start/fast_branch/fast_done/fast_error trace |

## 前端 Trace 点

| 事件 | 位置 |
|------|------|
| page_mount | 模块顶层 |
| load_resume_options_start/done | useEffect |
| init_recommendation_start | useEffect |
| latest_fast_response_received | quality-first init |
| latest_quality_response_received | quality async 加载 |
| generate_fast_start/done | generateFast |
| deep_click_start | handleDeepRecommendation |

## 后端 Trace 点

| 事件 | 位置 |
|------|------|
| handler_recommend_received | POST handler |
| handler_recommend_done/error | POST handler |
| service_recommend_start | Recommend service |
| service_fast_branch | fast SEP 分支 |
| service_fast_done | fast 成功 |
| service_fast_sep_error | fast SEP 失败 |

## 安全性

| 项目 | 结果 |
|------|:---:|
| 打印 token | ❌ 否 |
| 打印 API key | ❌ 否 |
| 打印简历全文 | ❌ 否 |
| 打印手机号/邮箱 | ❌ 否 |
| 仅打印耗时/状态/id | ✅ 是 |

## 复现步骤

1. 浏览器打开 `http://localhost:8080/recommend?resumeId=3&debugRecommend=1`
2. 打开 DevTools Console
3. 刷新页面
4. 等待 loading 结束或卡住 30s
5. 复制所有 `[REC_TRACE]` 日志

## 构建

| 项目 | 结果 |
|------|:---:|
| npm build | ✅ (1.90s) |
| go build | ✅ |
| go vet | ✅ |
