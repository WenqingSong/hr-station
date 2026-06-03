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
  const [scrolled, setScrolled] = useState(false);
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

  // 滚动隐藏 / 显示 + 顶部透明状态
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
          setScrolled(currentScrollY > 10);
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

  const brandFontStyle: React.CSSProperties = {
    fontFamily:
      '"Alimama ShuHeiTi", "HarmonyOS Sans SC", "MiSans", "Source Han Sans SC", "Noto Sans SC", "Inter", system-ui, sans-serif',
    letterSpacing: "-0.02em",
  };

  return (
    <nav
      className={`fixed top-0 left-0 right-0 z-50 h-16 transition-all duration-300 ease-out border-b ${
        hidden ? "-translate-y-full" : "translate-y-0"
      } ${
        scrolled
          ? "bg-white/75 backdrop-blur-xl border-white/20 shadow-[0_1px_3px_rgba(0,0,0,0.05)]"
          : "bg-white/40 backdrop-blur-md border-white/10"
      }`}
    >
      <div className="mx-auto flex h-full max-w-[1440px] items-center justify-between px-6 md:px-8">
        <Link
          to="/jobs"
          className="text-[20px] md:text-[22px] font-extrabold text-slate-900"
          style={brandFontStyle}
          aria-label="Offer 捕手首页"
        >
          Offer 捕手
        </Link>

        <div className="flex items-center gap-5">
          {isLoggedIn ? (
            <>
              <Link
                to="/resumes"
                className="inline-flex items-center gap-1.5 text-sm font-medium text-slate-600 transition-colors hover:text-[#0052D9]"
              >
                <FileText className="h-4 w-4" />
                我的简历
              </Link>
              <button
                onClick={handleLogout}
                className="inline-flex items-center gap-1.5 text-sm font-medium text-slate-500 transition-colors hover:text-[#0052D9]"
              >
                <LogOut className="h-4 w-4" />
                退出
              </button>
            </>
          ) : (
            <Link
              to="/login"
              className="inline-flex items-center gap-1.5 text-sm font-medium text-slate-600 transition-colors hover:text-[#0052D9]"
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
