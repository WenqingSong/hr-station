import type { ReactNode } from "react";
import { Navigate, useLocation } from "react-router";
import { useAuth } from "../../../stores/authStore";
import { PageLoading } from "../PageState";
import { normalizeReturnTo } from "../../../utils/redirect";

type RequireAuthProps = {
  children: ReactNode;
};

/**
 * RequireAuth — 路由级登录鉴权守卫（K8-5-46-B 接入 redirect 安全校验）。
 *
 * 行为：
 *   - isAuthReady 为 false → 显示 loading
 *   - isAuthReady 为 true && isAuthenticated 为 false → 跳转 /?auth=login&redirect=<安全路径>
 *   - isAuthenticated 为 true → 渲染 children
 */
export function RequireAuth({ children }: RequireAuthProps) {
  const location = useLocation();
  const { isAuthenticated, isAuthReady } = useAuth();

  if (!isAuthReady) {
    return <PageLoading message="正在验证登录状态…" />;
  }

  if (!isAuthenticated) {
    const returnTo = normalizeReturnTo(location.pathname, location.search, "/jobs");
    const redirect = encodeURIComponent(returnTo);
    return <Navigate to={`/?auth=login&redirect=${redirect}`} replace />;
  }

  return <>{children}</>;
}
