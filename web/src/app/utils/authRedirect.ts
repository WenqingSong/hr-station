import { isApiClientError } from "../../api/errors";

/**
 * 检查错误是否为 401（未登录/token 过期），若是则返回统一 AuthDialog redirect URL。
 * 返回 null 表示非 401 错误。
 *
 * K8-5-36-A：不再跳转到 /login（旧独立 LoginPage）。
 * 改为跳转到首页携带 auth=login 参数，触发 Navbar 中的统一 AuthDialog。
 */
export function getAuthRedirectIf401(
  error: unknown,
  pathname: string,
  search: string,
): string | null {
  if (isApiClientError(error) && error.status === 401) {
    return `/?auth=login&redirect=${encodeURIComponent(pathname + search)}`;
  }
  return null;
}
