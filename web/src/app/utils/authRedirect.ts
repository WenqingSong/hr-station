import { isApiClientError } from "../../api/errors";
import { normalizeReturnTo } from "../../utils/redirect";

/**
 * 检查错误是否为 401，若是则返回统一 AuthDialog redirect URL（K8-5-46-B 接入安全校验）。
 * 返回 null 表示非 401 错误。
 */
export function getAuthRedirectIf401(
  error: unknown,
  pathname: string,
  search: string,
): string | null {
  if (isApiClientError(error) && error.status === 401) {
    const returnTo = normalizeReturnTo(pathname, search, "/jobs");
    return `/?auth=login&redirect=${encodeURIComponent(returnTo)}`;
  }
  return null;
}
