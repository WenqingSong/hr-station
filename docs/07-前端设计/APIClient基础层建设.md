# API Client 基础层建设

> 阶段：5-2.1 | 状态：已完成 | 日期：2026-05-31

## 1. 本阶段目标

新增统一 API Client 基础层，为后续 5-2.2/5-2.3/5-2.4 业务 API 模块提供基础请求能力。

## 2. 新增文件

| 文件 | 行数 | 说明 |
|------|:--:|------|
| `web/src/api/types.ts` | 25 | API Client 内部类型（HttpMethod / ApiRequestOptions / ApiClientConfig / ApiRequestResult） |
| `web/src/api/errors.ts` | 28 | 统一错误类 ApiClientError + isApiClientError 守卫 |
| `web/src/api/client.ts` | 163 | 核心请求层（requestRaw / request / get / post / put / patch） |
| `web/src/api/index.ts` | 3 | 统一导出 |

**合计：4 文件，219 行**

## 3. base URL 规则

```ts
const DEFAULT_API_BASE_URL = '/api';
export const API_BASE_URL = import.meta.env?.VITE_API_BASE_URL || DEFAULT_API_BASE_URL;
```

- 默认 `/api`，不是 `/api/v1`
- 未配置环境变量时回退到默认值
- 调用时传入相对路径：`/jobs`、`/resumes/current`、`/jobs/recommend`

## 4. 请求能力

| 函数 | 返回 | 说明 |
|------|------|------|
| `requestRaw<T>()` | `ApiResponse<T>` | 完整响应（含 code/message/data） |
| `request<T>()` | `T` | 仅返回 data 字段 |
| `get<T>()` | `T` | GET 快捷方法 |
| `post<T, B>()` | `T` | POST 快捷方法 |
| `put<T, B>()` | `T` | PUT 快捷方法 |
| `patch<T, B>()` | `T` | PATCH 快捷方法 |

## 5. 错误处理策略

| 错误类型 | 处理 |
|------|------|
| 网络错误 | `ApiClientError` + 错误消息 |
| HTTP 非 2xx | `ApiClientError` + status |
| JSON 解析失败 | `ApiClientError` + 前 200 字符 |
| 响应缺 code 字段 | `ApiClientError` |
| `code !== 0` | `ApiClientError` + code/message/details |

关键行为：
- 不打印 token/key 等敏感信息
- JSON 解析失败时截断（前 200 字符）避免大日志
- 错误详情通过 `ApiRequestResult.details` 透传

## 6. 响应结构校验

```
后端响应: { code: 0, message: "ok", data: T }
requestRaw → ApiResponse<T>（含 code/message/data）
request    → T（仅 data）
如 code !== 0 → 抛出 ApiClientError
```

## 7. 本阶段不做

- 不创建 `api/jobs.ts` / `api/resumes.ts` / `api/agents.ts`
- 不接页面
- 不改后端
- 不引入 axios

## 8. 后续 5-2.2/5-2.3/5-2.4 输入

- 5-2.2：`api/jobs.ts` / `api/resumes.ts`，基于 client 封装业务请求 + mock fallback ✅ 已完成
- 5-2.3：`api/agents.ts`，封装三类 Agent API + data.result+data.meta 消费 ✅ 已完成
- 5-2.4：`api/viewModels.ts`，API 返回 → ViewModel 统一封装 ✅ 已完成（9 个 Loader 函数，120 行）
