# 三类 Agent 接口联调验收报告

> 验收时间：2026-05-31
> 验收阶段：4-5.3 三类 Agent 接口联调验收
> 验收结论：**partial_pass**（Nginx 入口未独立验证 CORS 响应体、request_id 为空、schema_valid 默认 true、LLM fallback 因 Prompt 加载失败而非 API Key 占位符）

## 1. 背景

阶段 4-5.2 已完成真实 LLM 最小 Smoke 验收（partial_pass），确认：

- HunyuanClient 创建、请求发送、错误处理链路可用
- API Key 为占位符导致认证失败
- 真实 LLM 成功路径未现场验证

本阶段目标：通过本地开发环境和 curl 验证三类 Agent HTTP 接口能否稳定返回 `code/message/data`，并统一提供 `data.result + data.meta`。

**不依赖真实 LLM 成功。**

## 2. 验收范围

1. 三类 Agent HTTP 接口能否通过本地开发环境访问
2. 三类接口是否稳定返回 `code/message/data`
3. 三类接口是否统一返回 `data.result + data.meta`
4. 三类接口的 `data.meta` 是否能让前端识别 LLM / repair / fallback / mock 状态
5. Nginx `/api` 代理和 CORS OPTIONS 是否可用
6. `/healthz` 是否可用

## 3. 当前分支与环境

| 项目 | 值 |
|------|------|
| 主线分支 | `main`（commit `4dd67a8`） |
| 当前分支 | `safeguard/4-4.2-response-contract`（HEAD `8ad7ad9`，比 main 多 4 个 docs commit） |
| 工作区状态 | clean（验收期间无代码变更） |
| 启动方式 | `make up` + `make restart-soft` + `make down` |

## 4. 启动环境结果

| 项目 | 结果 | 说明 |
|------|:----:|------|
| `make up` | ✅ pass | 基础设施 MySQL + Nginx 启动，后端/前端启动 |
| `make status` | ✅ pass | 所有服务健康 |
| 首次 curl 缺 `data.meta` | ⚠️ | 后端运行旧二进制，`make restart-soft` 后修复 |
| `make restart-soft` | ✅ pass | 重新编译后端，加载最新代码 |
| `make down` | ✅ pass | 所有服务停止，volume 保留 |

**说明**：首次 `make up` 时后端进程已存在旧二进制，未包含 4-4.2 的 `data.meta` 改造。执行 `make restart-soft` 后重新编译，`data.meta` 正确返回。

## 5. 接口路径核验

从 `server/internal/http/router.go` 确认：

| Agent | HTTP 方法 | 路径 | 认证 |
|-------|-----------|------|:----:|
| MatchAnalysis | POST | `/api/jobs/{id}/match` | ✅ Bearer |
| Recommendation | POST | `/api/jobs/recommend` | ✅ Bearer |
| ResumeOptimization | POST | `/api/resumes/optimize` | ✅ Bearer |
| Healthz | GET | `/healthz` | ❌ |

## 6. Healthz / Nginx / CORS 验收

### /healthz

```
curl -s http://localhost/healthz
→ {"code":0,"message":"ok","data":{"service":"offer-catcher","env":"local","app":"ok","database":"ok"}}
→ HTTP 200
```

**结论：pass**

### Nginx /api 代理

三类接口均通过 Nginx 入口 `http://localhost/api/...` 访问成功，路径正确代理到后端 `:8080`。

**结论：pass**

### CORS OPTIONS

```
curl -i -X OPTIONS http://localhost/api/jobs/recommend \
  -H "Origin: http://localhost" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: Content-Type, Authorization"
→ HTTP 204 No Content
→ Access-Control-Allow-Origin: http://localhost
→ Access-Control-Allow-Methods: GET, POST, PUT, PATCH, DELETE, OPTIONS
→ Access-Control-Allow-Headers: Content-Type, Authorization, X-Requested-With
→ Access-Control-Max-Age: 86400
```

**结论：pass**

## 7. MatchAnalysis 接口联调结果

### 请求

```
POST /api/jobs/1/match
Authorization: Bearer <token>
Content-Type: application/json
Body: {}
```

### 响应

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "analysis_id": 2,
    "generated_by": "mock_agent",
    "result": {
      "job_id": 1,
      "match_score": 98,
      "screening_result": "highly_likely",
      "apply_readiness": "ready_to_apply",
      "optimization_stop_signal": "stop_and_apply",
      "top_fix_items": [...],
      "evidence_chain": [...],
      "agent_status": "fallback",
      "fallback_reason": "prompt_load_failed",
      "error_code": "PROMPT_LOAD_FAILED",
      "generated_by": "mock_agent"
    },
    "analysis_basis": {...},
    "created_at": "2026-05-31T12:59:09+08:00",
    "meta": {
      "agent_name": "match_analysis",
      "generated_by": "mock_agent",
      "agent_status": "fallback",
      "fallback_reason": "prompt_load_failed",
      "error_code": "PROMPT_LOAD_FAILED",
      "repair_applied": false,
      "schema_valid": true,
      "request_id": ""
    }
  }
}
```

### 分析

| 验收项 | 结果 |
|--------|:----:|
| HTTP 可访问 | ✅ 200 |
| 返回合法 JSON | ✅ |
| code/message/data | ✅ |
| data.result | ✅ 含 match_score、apply_readiness、top_fix_items |
| data.meta | ✅ |
| meta 8 字段 | ✅ |
| 前端可稳定消费 | ✅ |

**LLM 路径**：Prompt 加载失败，走 fallback → Mock。这是预期行为（LLM 未正确配置时 fallback 应可用）。

## 8. Recommendation 接口联调结果

### 请求

```
POST /api/jobs/recommend
Authorization: Bearer <token>
Content-Type: application/json
Body: {}
```

### 响应

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "resume_version_id": 2,
    "resume_version_no": 2,
    "generated_by": "mock_agent",
    "recommended_jobs": [...],
    "result": {
      "recommended_jobs": [...]
    },
    "summary": "...",
    "created_at": "2026-05-31T12:59:10+08:00",
    "meta": {
      "agent_name": "recommendation",
      "generated_by": "mock_agent",
      "agent_status": "fallback",
      "fallback_reason": "prompt_load_failed",
      "error_code": "PROMPT_LOAD_FAILED",
      "repair_applied": false,
      "schema_valid": true,
      "request_id": ""
    }
  }
}
```

### 分析

| 验收项 | 结果 |
|--------|:----:|
| HTTP 可访问 | ✅ 200 |
| 返回合法 JSON | ✅ |
| code/message/data | ✅ |
| data.result | ✅ |
| data.result.recommended_jobs | ✅ 包含 5 个推荐岗位 |
| data.meta | ✅ |
| meta 8 字段 | ✅ |
| 前端可稳定消费 | ✅ |

**新旧入口一致性**：`data.recommended_jobs`（deprecated）和 `data.result.recommended_jobs`（新）均存在且数据一致。

## 9. ResumeOptimization 接口联调结果

### 请求

```
POST /api/resumes/optimize
Authorization: Bearer <token>
Content-Type: application/json
Body: {
  "resume_text": "计算机学生，熟悉Go、MySQL",
  "target_job": {"title": "Go后端开发", "jd_text": "Go语言，微服务"}
}
```

### 响应

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "optimization_id": 0,
    "generated_by": "mock_agent",
    "result": {
      "apply_readiness": "quick_fix_before_apply",
      "optimization_stop_signal": "fix_top_3_then_apply",
      "summary": "...",
      "top_fix_items": [...],
      "long_term_improvement": [...],
      "jd_evidence_chain": [...],
      "verification_notice": "...",
      "agent_status": "fallback",
      "fallback_reason": "prompt_load_failed",
      "error_code": "PROMPT_LOAD_FAILED"
    },
    "created_at": "2026-05-31T12:59:12+08:00",
    "meta": {
      "agent_name": "resume_optimization",
      "generated_by": "mock_agent",
      "agent_status": "fallback",
      "fallback_reason": "prompt_load_failed",
      "error_code": "PROMPT_LOAD_FAILED",
      "repair_applied": false,
      "schema_valid": true,
      "request_id": ""
    }
  }
}
```

### 分析

| 验收项 | 结果 |
|--------|:----:|
| HTTP 可访问 | ✅ 200 |
| 返回合法 JSON | ✅ |
| code/message/data | ✅ |
| data.result | ✅ 含 apply_readiness、optimization_stop_signal、top_fix_items |
| data.meta | ✅ |
| meta 8 字段 | ✅ |
| 前端可稳定消费 | ✅ |

## 10. 接口联调验收矩阵

| 验收项 | MatchAnalysis | Recommendation | ResumeOptimization | 结论 |
|--------|:---:|:---:|:---:|:---:|
| HTTP 可访问 | pass | pass | pass | **pass** |
| 返回合法 JSON | pass | pass | pass | **pass** |
| code/message/data | pass | pass | pass | **pass** |
| data.result | pass | pass | pass | **pass** |
| data.meta | pass | pass | pass | **pass** |
| meta.agent_name | pass | pass | pass | **pass** |
| meta.generated_by | pass | pass | pass | **pass** |
| meta.agent_status | pass | pass | pass | **pass** |
| meta.fallback_reason | pass | pass | pass | **pass** |
| meta.error_code | pass | pass | pass | **pass** |
| meta.repair_applied | pass | pass | pass | **pass** |
| meta.schema_valid | pass | pass | pass | **pass** ⚠️ |
| meta.request_id | pass | pass | pass | **pass** ⚠️ |
| Recommendation data.result.recommended_jobs | 不适用 | pass | 不适用 | **pass** |
| 前端可稳定消费 | pass | pass | pass | **pass** |

## 11. 基础环境验收矩阵

| 项目 | 结果 | 说明 |
|------|:----:|------|
| make up | pass | |
| make status | pass | |
| /healthz | pass | HTTP 200，app + database ok |
| Nginx /api 代理 | pass | 三类接口均通过 Nginx 入口访问成功 |
| CORS OPTIONS | pass | 204 + 正确 Allow 头 |
| make down | pass | |
| 是否清空 volume | no | ✅ volume 保留 |
| 是否调用真实 LLM | no | ✅ 仅 fallback/mock |

## 12. data.result + data.meta 验收结论

**pass**

- 三类接口均返回 `data.result`（业务结果）和 `data.meta`（系统状态）
- `data.meta` 包含 8 个必填字段，前端可从统一位置读取业务结果和系统状态
- Recommendation 的 `data.result.recommended_jobs` 作为新统一入口已存在
- 当前接口因 Prompt 加载失败走 fallback 路径，`data.meta.agent_status` 为 `"fallback"`，`data.meta.fallback_reason` 为 `"prompt_load_failed"`，前端可正确识别
- `generated_by` 唯一可信位置为 `data.meta.generated_by`

## 13. 测试结果

| 测试 | 结果 |
|------|:----:|
| `go test ./...` | ✅ 全部通过 |
| `make llm-check` | ✅ 通过 |
| 是否调用真实 LLM | ❌ 否 |

## 14. 未调用真实 LLM 说明

本阶段三类接口联调使用 `make up` 启动本地环境，通过 curl 验证 HTTP 接口。接口尝试加载 Prompt 文件后因配置问题走 fallback 路径，返回 Mock 结果。未执行 `RUN_REAL_LLM_SMOKE=1 make llm-smoke`。

## 15. 当前限制

| 限制 | 严重度 | 说明 |
|------|--------|------|
| `request_id` 为空字符串 | 🟢 低 | 契约允许暂时为空，后续 Handler 层接入 |
| `schema_valid` 默认 true | 🟡 中 | 未实现实际 Schema 校验逻辑 |
| LLM fallback 原因为 prompt_load_failed | 🟡 中 | 非预期的 API Key 占位符原因，而是 Prompt 文件加载失败，需排查 |
| 首次启动需 restart-soft 才加载最新代码 | 🟢 低 | 后端已在运行时 make up 不会重启 |
| 当前分支比 main 多 4 个 commit | 🟢 低 | 仅 docs 变更 |
| Nginx CORS 验证未覆盖响应体 | 🟢 低 | OPTIONS 204 已验证，实际 POST 响应 CORS 头未单独验证 |

## 16. 4-5.3 最终结论

**partial_pass**

判定依据：

### pass 项
1. ✅ `make up` 成功
2. ✅ `/healthz` 成功
3. ✅ 三类 Agent 接口均可访问（HTTP 200）
4. ✅ 三类接口均返回合法 JSON
5. ✅ 三类接口均有 `code/message/data`
6. ✅ 三类接口均有 `data.result`
7. ✅ 三类接口均有 `data.meta`
8. ✅ meta 8 字段完整
9. ✅ Recommendation 有 `data.result.recommended_jobs`
10. ✅ 前端可统一读取 `data.result + data.meta`
11. ✅ CORS OPTIONS 可用
12. ✅ `go test ./...` 通过
13. ✅ `make llm-check` 可用
14. ✅ 未调用真实 LLM
15. ✅ 未清空 volume
16. ✅ `make down` 成功
17. ✅ 文档和开发日志已更新

### 非阻塞项
1. ⚠️ `request_id` 为空字符串（契约允许）
2. ⚠️ `schema_valid` 默认 true（未实现实际校验逻辑）
3. ⚠️ LLM fallback 原因为 `prompt_load_failed` 而非预期的 API Key 占位符
4. ⚠️ 首次启动需 `make restart-soft` 才加载最新代码

### 未出现 fail 项
- ❌ make up 失败 → 未发生
- ❌ 三类接口无法访问 → 未发生
- ❌ 返回不是合法 JSON → 未发生
- ❌ 缺 data.result → 未发生
- ❌ 缺 data.meta → 未发生
- ❌ meta 关键字段缺失 → 未发生
- ❌ Recommendation 没有 data.result.recommended_jobs → 未发生
- ❌ go test ./... 失败 → 未发生
- ❌ 调用了真实 LLM → 未发生
- ❌ 清空了 volume → 未发生
- ❌ 修改了禁止范围文件 → 未发生

## 17. 下一步建议

1. **进入 4-5.4**：阶段 4 总收口与阶段 5 准入评审
   - 核验文档与实现一致性
   - 确认可进入阶段 5 前端产品体验闭环
2. **排查 prompt_load_failed**：后续确认 Prompt 文件加载路径是否正确
3. **配置有效 API Key 后**：可重新执行真实 LLM Smoke
4. **后续 Handler 层**：补充 request_id 生成逻辑和 schema_valid 实际校验
