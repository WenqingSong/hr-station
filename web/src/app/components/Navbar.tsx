import { useState, useEffect } from "react";
import { FileText, LogIn, LogOut, Sparkles } from "lucide-react";
import { Link, useNavigate } from "react-router";
import { getAuthToken, clearAuthToken } from "../../api/authStorage";

export function Navbar() {
  const navigate = useNavigate();
  const [isLoggedIn, setIsLoggedIn] = useState(false);

  useEffect(() => {
    function checkAuth() {
      setIsLoggedIn(!!getAuthToken());
    }
    checkAuth();
    // 监听 storage 变化（跨 tab 同步）
    window.addEventListener("storage", checkAuth);
    return () => window.removeEventListener("storage", checkAuth);
  }, []);

  function handleLogout() {
    clearAuthToken();
    setIsLoggedIn(false);
    navigate("/");
  }

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

        <div className="flex items-center gap-4">
          <Link
            to="/recommend"
            className="inline-flex shrink-0 items-center gap-1.5 text-sm font-medium text-slate-600 transition-colors hover:text-[#0052D9]"
          >
            <Sparkles className="h-4 w-4" />
            推荐岗位
          </Link>
          <Link
            to="/resumes"
            className="inline-flex shrink-0 items-center gap-1.5 text-sm font-medium text-slate-600 transition-colors hover:text-[#0052D9]"
          >
            <FileText className="h-4 w-4" />
            我的简历
          </Link>
          {isLoggedIn ? (
            <button
              onClick={handleLogout}
              className="inline-flex shrink-0 items-center gap-1.5 text-sm font-medium text-slate-600 transition-colors hover:text-[#0052D9]"
            >
              <LogOut className="h-4 w-4" />
              退出
            </button>
          ) : (
            <Link
              to="/login"
              className="inline-flex shrink-0 items-center gap-1.5 text-sm font-medium text-slate-600 transition-colors hover:text-[#0052D9]"
            >
              <LogIn className="h-4 w-4" />
              登录
            </Link>
          )}
        </div>
      </div>
    </nav>
  );
}
