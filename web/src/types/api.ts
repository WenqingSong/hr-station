// ─── 通用 API 响应类型 ────────────────────────────────────────────────────────

/** HTTP 外层统一响应 code */
export type ApiCode = number;

/**
 * 标准 API 外层响应 envelope。
 * - code: 0 = 成功，非 0 = 错误
 * - message: 成功时为 "ok"，错误时为描述
 * - data: 业务数据容器
 */
export interface ApiResponse<TData = unknown> {
  code: ApiCode;
  message: string;
  data: TData;
}

/** 分页结果 */
export interface PaginatedResult<TItem> {
  items: TItem[];
  total: number;
  page?: number;
  page_size?: number;
}

/** API 错误信息 */
export interface ApiErrorInfo {
  code: string;
  message: string;
  details?: string;
}
