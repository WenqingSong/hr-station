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
