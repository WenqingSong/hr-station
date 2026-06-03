import { isApiClientError } from "../../api/errors";

/**
 * 检查错误是否为 401（未登录/token 过期），若是则返回 login redirect URL。
 * 返回 null 表示非 401 错误。
 */
export function getAuthRedirectIf401(
  error: unknown,
  pathname: string,
  search: string,
): string | null {
  if (isApiClientError(error) && error.status === 401) {
    return `/login?redirect=${encodeURIComponent(pathname + search)}`;
  }
  return null;
}
