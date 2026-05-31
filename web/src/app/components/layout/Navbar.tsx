// @deprecated — 5-3.1 已废弃。唯一生效 Navbar 见 ../Navbar.tsx。
// 保留此文件仅用于历史追溯，后续清理阶段可安全删除。

import { FileText } from "lucide-react";
import { Link } from "react-router";

export function Navbar() {
  return (
    <nav className="sticky top-0 z-50 h-14 border-b border-gray-200 bg-white">
      <div className="mx-auto flex h-full max-w-[1360px] items-center justify-between px-6">
        <Link to="/jobs" className="flex min-w-0 items-center gap-2 text-[#0052D9]" aria-label="Offer 捕手首页">
          <span className="flex h-5 w-5 shrink-0 items-center justify-center rounded bg-[#0052D9] text-[11px] font-semibold leading-none text-white">
            O
          </span>
          <span className="truncate text-[15px] font-semibold tracking-tight text-slate-900">
            「Offer 捕手」学生求职匹配智能体
          </span>
        </Link>

        <Link
          to="/resumes"
          className="inline-flex shrink-0 items-center gap-1.5 text-sm font-medium text-slate-600 transition-colors hover:text-[#0052D9]"
        >
          <FileText className="h-4 w-4" />
          我的简历
        </Link>
      </div>
    </nav>
  );
}
