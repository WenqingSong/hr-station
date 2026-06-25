// ─── API Client 基础层 ─────────────────────────────────────────────────────────

import type { ApiResponse } from '../types';
import type { ApiRequestOptions } from './types';
import { ApiClientError } from './errors';
import { clearAuthToken, getAuthToken } from './authStorage';

// ─── Base URL ──────────────────────────────────────────────────────────────────

const DEFAULT_API_BASE_URL = '/api';

export const API_BASE_URL: string =
  (typeof import.meta !== 'undefined' &&
    import.meta.env?.VITE_API_BASE_URL?.trim?.()) ||
  DEFAULT_API_BASE_URL;

// ─── Helpers ───────────────────────────────────────────────────────────────────

/** 拼接 baseUrl + path，去除重复 / */
function joinUrl(baseUrl: string, path: string): string {
  const base = baseUrl.replace(/\/+$/, '');
  const p = (path || '').replace(/^\/+/, '');
  return `${base}/${p}`;
}

/** 类型守卫：判断值是否符合 ApiResponse 结构 */
function isApiResponse<TData>(value: unknown): value is ApiResponse<TData> {
  if (value === null || value === undefined) return false;
  if (typeof value !== 'object') return false;
  const obj = value as Record<string, unknown>;
  return typeof obj.code === 'number';
}

/** 解析 JSON 响应体 */
async function parseJsonResponse(response: Response): Promise<unknown> {
  const text = await response.text();
  if (!text) {
    throw new ApiClientError('Empty response body', {
      status: response.status,
    });
  }
  try {
    return JSON.parse(text);
  } catch {
    throw new ApiClientError('Failed to parse response JSON', {
      status: response.status,
      details: text.substring(0, 200),
    });
  }
}

/**
 * 构建请求 headers。
 * - 默认 Content-Type: application/json
 * - 调用方传入的 headers 优先（包括显式 Authorization）
 * - 如果调用方未传 Authorization，从 storage 中注入 token
 */
function buildHeaders(explicitHeaders?: Record<string, string>): Record<string, string> {
  const base: Record<string, string> = {
    'Content-Type': 'application/json',
  };

  const merged = { ...base, ...explicitHeaders };

  // 仅当调用方未显式传 Authorization 时，自动从 storage 注入
  if (!merged.Authorization && !merged.authorization) {
    const token = getAuthToken();
    if (token) {
      merged['Authorization'] = `Bearer ${token}`;
    }
  }

  return merged;
}

// ─── 核心请求 ──────────────────────────────────────────────────────────────────

/**
 * 底层请求函数：发起 fetch 并返回完整 ApiResponse。
 * - HTTP 非 2xx → 抛出 ApiClientError
 * - JSON 解析失败 → 抛出 ApiClientError
 * - 响应不符合 { code, message, data } → 抛出 ApiClientError
 * - code !== 0 → 抛出 ApiClientError
 */
export async function requestRaw<TData, TBody = unknown>(
  path: string,
  options?: ApiRequestOptions<TBody>,
): Promise<ApiResponse<TData>> {
  const url = joinUrl(API_BASE_URL, path);
  const method = options?.method ?? 'GET';

  const init: RequestInit = {
    method,
    headers: buildHeaders(options?.headers),
    signal: options?.signal,
  };

  // GET/HEAD 不应带 body
  if (options?.body !== undefined && method !== 'GET' && method !== 'HEAD') {
    init.body = JSON.stringify(options.body);
  }

  let response: Response;
  try {
    response = await fetch(url, init);
  } catch (err) {
    throw new ApiClientError(
      err instanceof Error ? err.message : 'Network request failed',
    );
  }

  if (!response.ok) {
    // 尝试解析业务 code，区分 session-replaced (41005) 和普通 token 过期
    let bizCode: number | undefined;
    let bizMessage: string | undefined;
    try {
      const errorBody = await parseJsonResponse(response);
      if (typeof (errorBody as any)?.code === "number") {
        bizCode = (errorBody as any).code;
        bizMessage = (errorBody as any).message;
      }
    } catch {
      // 无法解析 body 时不阻塞
    }

    if (response.status === 401 && bizCode === 41005) {
      // 账号在其他设备登录 → 清 token + 派发专用事件
      clearAuthToken();
      if (typeof window !== "undefined") {
        window.dispatchEvent(new Event("offerr-auth-replaced"));
      }
      throw new ApiClientError(
        bizMessage || "账号已在其他设备登录，请重新登录",
        { status: response.status, code: bizCode },
      );
    }

    // 401 统一清 token + 通知 React 组件刷新登录态（K8-5-36-C）
    if (response.status === 401) {
      clearAuthToken();
      if (typeof window !== "undefined") {
        window.dispatchEvent(new Event("offerr-auth-expired"));
      }
    }
    throw new ApiClientError(
      `HTTP ${response.status}: ${response.statusText}`,
      { status: response.status },
    );
  }

  const parsed = await parseJsonResponse(response);

  if (!isApiResponse<TData>(parsed)) {
    throw new ApiClientError('Invalid response structure: missing code field', {
      status: response.status,
    });
  }

  if (parsed.code !== 0) {
    throw new ApiClientError(parsed.message || 'API error', {
      status: response.status,
      code: parsed.code,
      details: parsed.data,
    });
  }

  return parsed;
}

// ─── 快捷方法 ──────────────────────────────────────────────────────────────────

/**
 * 请求并返回 ApiResponse.data。
 * 字段校验、错误处理由 requestRaw 承担。
 */
export async function request<TData, TBody = unknown>(
  path: string,
  options?: ApiRequestOptions<TBody>,
): Promise<TData> {
  const result = await requestRaw<TData, TBody>(path, options);
  return result.data;
}

/** GET 请求 */
export function get<TData>(
  path: string,
  options?: Omit<ApiRequestOptions, 'method' | 'body'>,
): Promise<TData> {
  return request<TData>(path, { ...options, method: 'GET' });
}

/** POST 请求 */
export function post<TData, TBody = unknown>(
  path: string,
  body?: TBody,
  options?: Omit<ApiRequestOptions<TBody>, 'method' | 'body'>,
): Promise<TData> {
  return request<TData, TBody>(path, { ...options, method: 'POST', body });
}

/** PUT 请求 */
export function put<TData, TBody = unknown>(
  path: string,
  body?: TBody,
  options?: Omit<ApiRequestOptions<TBody>, 'method' | 'body'>,
): Promise<TData> {
  return request<TData, TBody>(path, { ...options, method: 'PUT', body });
}

/** PATCH 请求 */
export function patch<TData, TBody = unknown>(
  path: string,
  body?: TBody,
  options?: Omit<ApiRequestOptions<TBody>, 'method' | 'body'>,
): Promise<TData> {
  return request<TData, TBody>(path, { ...options, method: 'PATCH', body });
}
