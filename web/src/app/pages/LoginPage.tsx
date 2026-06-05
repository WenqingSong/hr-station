import { useEffect } from "react";
import { useNavigate, useSearchParams } from "react-router";

/**
 * LoginPage — K8-5-36-A：已废弃独立登录表单，转为统一 AuthDialog 重定向。
 *
 * 本组件不再是用户可见的独立登录页面。
 * 所有 /login 访问自动重定向到首页并触发统一登录弹窗：
 *   /login                      → /?auth=login
 *   /login?redirect=%2Fresumes  → /?auth=login&redirect=%2Fresumes
 *
 * 统一登录体验由 Navbar + AuthDialog 提供。
 * "评委一键体验"保留在 AuthDialog 内。
 */
export function LoginPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();

  useEffect(() => {
    const redirect = searchParams.get("redirect");
    const target = redirect
      ? `/?auth=login&redirect=${encodeURIComponent(redirect)}`
      : "/?auth=login";
    navigate(target, { replace: true });
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  return null;
}
