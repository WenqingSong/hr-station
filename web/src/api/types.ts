// ─── API Client 内部类型 ────────────────────────────────────────────────────────

import type { ApiResponse } from '../types';

/** HTTP 方法 */
export type HttpMethod = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE' | 'HEAD';

/** API 请求选项 */
export interface ApiRequestOptions<TBody = unknown> {
  method?: HttpMethod;
  body?: TBody;
  headers?: Record<string, string>;
  signal?: AbortSignal;
}

/** API Client 配置 */
export interface ApiClientConfig {
  baseUrl: string;
}

/** 请求结果：解析后的 data + 原始 raw */
export interface ApiRequestResult<TData> {
  data: TData;
  raw: ApiResponse<TData>;
}
