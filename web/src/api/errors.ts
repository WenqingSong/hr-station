// ─── API 统一错误 ───────────────────────────────────────────────────────────────

/** API Client 统一错误 */
export class ApiClientError extends Error {
  public readonly status?: number;
  public readonly code?: string | number;
  public readonly details?: unknown;

  constructor(
    message: string,
    options?: {
      status?: number;
      code?: string | number;
      details?: unknown;
    },
  ) {
    super(message);
    this.name = 'ApiClientError';
    this.status = options?.status;
    this.code = options?.code;
    this.details = options?.details;
  }
}

/** 类型守卫：是否为 ApiClientError */
export function isApiClientError(error: unknown): error is ApiClientError {
  return error instanceof ApiClientError;
}
