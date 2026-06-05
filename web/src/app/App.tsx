import { useEffect } from "react";
import { RouterProvider } from "react-router";
import { router } from "./routes";
import { AuthProvider, useAuth } from "../stores/authStore";
import { AppBootLoading } from "./components/common/AppBootLoading";

/**
 * AuthBoot — App 启动时执行 auth bootstrap。
 * 放在 AuthProvider 内部，通过 useAuth() 访问 store。
 * 启动期间显示 AppBootLoading，完成后不阻塞页面渲染。
 */
function AuthBoot() {
  const { bootstrapAuth, isBootstrapping, isAuthReady } = useAuth();

  useEffect(() => {
    bootstrapAuth();
  }, [bootstrapAuth]);

  // K8-5-37-B：bootstrap 进行中 → 显示品牌 loading
  if (isBootstrapping && !isAuthReady) {
    return <AppBootLoading />;
  }

  return null;
}

/**
 * App — K8-5-36-C：包裹 AuthProvider，统一登录态。
 * K8-5-37-B：接入 AppBootLoading 作为 Auth Bootstrap 视觉反馈。
 */
export default function App() {
  return (
    <AuthProvider>
      <AuthBoot />
      <RouterProvider router={router} />
    </AuthProvider>
  );
}
