import { useState, useEffect, useRef } from "react";
import { FileText, LogIn, LogOut } from "lucide-react";
import { Link, useNavigate, useLocation } from "react-router";
import { getAuthToken, clearAuthToken } from "../../api/authStorage";
import { logout } from "../../api/auth";

export function Navbar() {
  const navigate = useNavigate();
  const location = useLocation();
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [hidden, setHidden] = useState(false);
  const [isTop, setIsTop] = useState(true);
  const lastScrollY = useRef(0);
  const ticking = useRef(false);

  useEffect(() => {
    setIsLoggedIn(!!getAuthToken());
  }, [location]);

  // 监听 storage 变化（跨 tab 同步）
  useEffect(() => {
    function checkAuth() {
      setIsLoggedIn(!!getAuthToken());
    }
    window.addEventListener("storage", checkAuth);
    return () => window.removeEventListener("storage", checkAuth);
  }, []);

  // 滚动隐藏 / 显示 + 顶部判定
  useEffect(() => {
    const handleScroll = () => {
      if (!ticking.current) {
        window.requestAnimationFrame(() => {
          const currentScrollY = window.scrollY;
          if (currentScrollY > lastScrollY.current && currentScrollY > 80) {
            setHidden(true);
          } else if (currentScrollY < lastScrollY.current) {
            setHidden(false);
          }
          if (currentScrollY <= 20) {
            setHidden(false);
          }
          setIsTop(currentScrollY <= 80);
          lastScrollY.current = currentScrollY;
          ticking.current = false;
        });
        ticking.current = true;
      }
    };
    window.addEventListener("scroll", handleScroll, { passive: true });
    return () => window.removeEventListener("scroll", handleScroll);
  }, []);

  async function handleLogout() {
    try { await logout(); } catch (_) { /* ignore if API fails */ }
    clearAuthToken();
    setIsLoggedIn(false);
    navigate("/");
  }

  const brandTextShadow: React.CSSProperties | undefined = isTop
    ? { textShadow: "0 1px 12px rgba(0,0,0,0.18)" }
    : undefined;

  // 顶部：深色透明玻璃悬浮在 Banner 上，白字
  const navSurfaceClass = isTop
    ? "border-white/10 bg-slate-950/18 text-white shadow-none backdrop-blur-md backdrop-saturate-150"
    : "border-slate-200/40 bg-white/58 text-slate-950 shadow-[0_8px_24px_rgba(15,23,42,0.06)] backdrop-blur-xl backdrop-saturate-150";

  const brandTextClass = isTop
    ? "text-white"
    : "text-slate-950";

  const linkClass = isTop
    ? "text-white/88 hover:text-white"
    : "text-slate-700 hover:text-slate-950";

  const logoutLinkClass = isTop
    ? "text-white/72 hover:text-white"
    : "text-slate-500 hover:text-slate-950";

  const loginButtonClass = isTop
    ? "rounded-full border border-white/25 bg-white/8 px-4 py-2 text-sm font-medium text-white/92 backdrop-blur transition hover:bg-white/14 hover:text-white"
    : "rounded-full border border-slate-200/70 bg-white/50 px-4 py-2 text-sm font-medium text-slate-700 backdrop-blur transition hover:bg-white/80 hover:text-slate-950";

  return (
    <nav
      className={[
        "fixed inset-x-0 top-0 z-50",
        "h-16 md:h-[68px]",
        "border-b",
        "transition-[transform,background-color,border-color,box-shadow,color] duration-300 ease-[cubic-bezier(0.22,1,0.36,1)]",
        hidden ? "-translate-y-[110%]" : "translate-y-0",
        navSurfaceClass,
      ].join(" ")}
    >
      <div className="mx-auto flex h-full max-w-[1440px] items-center justify-between px-6 md:px-8">
        <Link
          to="/jobs"
          className={`font-brand text-[22px] font-extrabold tracking-[-0.03em] ${brandTextClass}`}
          style={brandTextShadow}
          aria-label="Offerr鹅首页"
        >
          Offerr鹅
        </Link>

        <div className="flex items-center gap-5">
          {isLoggedIn ? (
            <>
              <Link
                to="/resumes"
                className={`inline-flex items-center gap-1.5 text-sm font-medium transition-colors ${linkClass}`}
              >
                <FileText className="h-4 w-4" />
                我的简历
              </Link>
              <button
                onClick={handleLogout}
                className={`inline-flex items-center gap-1.5 text-sm font-medium transition-colors ${logoutLinkClass}`}
              >
                <LogOut className="h-4 w-4" />
                退出
              </button>
            </>
          ) : (
            <Link
              to="/login"
              className={`inline-flex items-center gap-1.5 ${loginButtonClass}`}
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
