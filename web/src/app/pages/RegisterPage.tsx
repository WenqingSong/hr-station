import { useEffect } from "react";
import { useNavigate, useSearchParams } from "react-router";

/**
 * RegisterPage — K8-5-36-A：已废弃独立注册表单，转为统一 AuthDialog 重定向。
 *
 * 本组件不再是用户可见的独立注册页面。
 * 所有 /register 访问自动重定向到首页并触发 AuthDialog：
 *   /register                      → /?auth=login
 *   /register?redirect=%2Fresumes  → /?auth=login&redirect=%2Fresumes
 *
 * 注册入口在 AuthDialog 内通过 tab 切换提供。
 */
export function RegisterPage() {
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
