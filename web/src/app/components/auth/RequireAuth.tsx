import type { ReactNode } from "react";
import { Navigate, useLocation } from "react-router";
import { useAuth } from "../../../stores/authStore";
import { PageLoading } from "../PageState";

type RequireAuthProps = {
  children: ReactNode;
};

/**
 * RequireAuth — 路由级登录鉴权守卫（K8-5-36-C 接入统一 AuthContext）。
 *
 * 行为：
 *   - isAuthReady 为 false（bootstrap 未完成）→ 显示 loading
 *   - isAuthReady 为 true && isAuthenticated 为 false → 跳转 /?auth=login&redirect=<当前路径>
 *   - isAuthenticated 为 true → 渲染 children
 *
 * 不再直接读 localStorage token。
 * 不再跳转到 /login（旧独立 LoginPage）。
 */
export function RequireAuth({ children }: RequireAuthProps) {
  const location = useLocation();
  const { isAuthenticated, isAuthReady } = useAuth();

  if (!isAuthReady) {
    return <PageLoading message="正在验证登录状态…" />;
  }

  if (!isAuthenticated) {
    const redirect = encodeURIComponent(location.pathname + location.search);
    return <Navigate to={`/?auth=login&redirect=${redirect}`} replace />;
  }

  return <>{children}</>;
}
