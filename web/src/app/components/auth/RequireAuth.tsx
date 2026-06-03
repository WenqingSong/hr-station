import type { ReactNode } from "react";
import { Navigate, useLocation } from "react-router";
import { getAuthToken } from "../../../api/authStorage";

type RequireAuthProps = {
  children: ReactNode;
};

/**
 * RequireAuth — 路由级登录鉴权守卫。
 *
 * 行为：
 * - 调用 getAuthToken() 同步检查 localStorage 中的 token。
 * - 无 token → <Navigate to="/login?redirect=<当前路径>" replace />
 * - 有 token → 渲染 children。
 *
 * 不做：
 * - 不弹 AuthDialog（AuthDialog 绑定 Navbar state）。
 * - 不修改 client.ts 401 逻辑。
 * - 不做 token 过期校验（token 过期由后端 401 驱动，归 E 阶段）。
 */
export function RequireAuth({ children }: RequireAuthProps) {
  const location = useLocation();
  const token = getAuthToken();

  if (!token) {
    const redirect = encodeURIComponent(location.pathname + location.search);
    return <Navigate to={`/login?redirect=${redirect}`} replace />;
  }

  return <>{children}</>;
}
