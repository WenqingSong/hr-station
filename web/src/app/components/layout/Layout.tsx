// @deprecated — 5-3.1 已废弃。全局布局由 routes.tsx 中 RootLayout 统一管理。
// 保留此文件仅用于历史追溯，后续清理阶段可安全删除。

import { Outlet } from "react-router";
import { Navbar } from "./Navbar";

export function Layout() {
  return (
    <div className="min-h-screen bg-slate-50 text-slate-900 font-sans pt-14">
      <Navbar />
      <div className="max-w-[1600px] mx-auto px-6 py-6 flex gap-6 items-start justify-center">
        <Outlet />
      </div>
    </div>
  );
}
