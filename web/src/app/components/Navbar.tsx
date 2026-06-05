import { useState, useEffect, useRef } from "react";
import { FileText, LogIn, LogOut } from "lucide-react";
import { Link, useNavigate, useLocation, useSearchParams } from "react-router";
import { logout } from "../../api/auth";
import { useAuth } from "../../stores/authStore";
import { AuthDialog } from "./auth/AuthDialog";

export function Navbar() {
  const navigate = useNavigate();
  const location = useLocation();
  const [searchParams, setSearchParams] = useSearchParams();
  const { isAuthenticated, clearAuth } = useAuth();
  const [hidden, setHidden] = useState(false);
  const [isTop, setIsTop] = useState(true);
  const [authDialogOpen, setAuthDialogOpen] = useState(false);
  const [authDialogMode, setAuthDialogMode] = useState<"login" | "register">("login");
  const lastScrollY = useRef(0);
  const ticking = useRef(false);
  const redirectAfterLoginRef = useRef<string | null>(null);

  useEffect(() => {
    const auth = searchParams.get("auth");
    if (auth === "login" || auth === "register") {
      redirectAfterLoginRef.current = searchParams.get("redirect");
      setAuthDialogMode(auth === "register" ? "register" : "login");
      setAuthDialogOpen(true);
      // 清理 URL 参数，避免刷新时反复触发
      setSearchParams({}, { replace: true });
    }
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

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

  const handleAuthSuccess = () => {
    // AuthDialog 已调用 useAuth().setAuthenticated() 更新全局状态
    // 此处仅处理 redirect 跳转
    if (redirectAfterLoginRef.current) {
      const target = redirectAfterLoginRef.current;
      redirectAfterLoginRef.current = null;
      navigate(target);
    }
  };

  async function handleLogout() {
    try { await logout(); } catch (_) { /* ignore if API fails */ }
    clearAuth();
    navigate("/");
  }

  // K8-5-15-C-R：路由语义判断 — Hero 页面顶部才允许白字/透明 Navbar
  const isHeroNavbarRoute =
    location.pathname === "/" ||
    location.pathname === "/jobs" ||
    /^\/jobs\/[^/]+$/.test(location.pathname) ||
    /^\/job\/[^/]+$/.test(location.pathname);

  const useHeroNavbar = isHeroNavbarRoute && isTop;

  const brandTextShadow: React.CSSProperties | undefined = useHeroNavbar
    ? { textShadow: "0 1px 12px rgba(0,0,0,0.18)" }
    : undefined;

  // Hero 页顶部：深色透明玻璃悬浮在 Banner 上，白字
  // 普通内容页：白底毛玻璃 + 深色字
  const navSurfaceClass = useHeroNavbar
    ? "border-white/10 bg-slate-950/18 text-white shadow-none backdrop-blur-md backdrop-saturate-150"
    : "border-slate-200/40 bg-white/58 text-slate-950 shadow-[0_8px_24px_rgba(15,23,42,0.06)] backdrop-blur-xl backdrop-saturate-150";

  const brandTextClass = useHeroNavbar
    ? "text-white"
    : "text-slate-950";

  const linkClass = useHeroNavbar
    ? "text-white/88 hover:text-white"
    : "text-slate-700 hover:text-slate-950";

  const logoutLinkClass = useHeroNavbar
    ? "text-white/72 hover:text-white"
    : "text-slate-500 hover:text-slate-950";

  const loginButtonClass = useHeroNavbar
    ? "rounded-full border border-white/25 bg-white/8 px-4 py-2 text-sm font-medium text-white/92 backdrop-blur transition hover:bg-white/14 hover:text-white"
    : "rounded-full border border-slate-200/70 bg-white/50 px-4 py-2 text-sm font-medium text-slate-700 backdrop-blur transition hover:bg-white/80 hover:text-slate-950";

  return (
    <>
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
          {isAuthenticated ? (
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
            <button
              type="button"
              onClick={() => { setAuthDialogMode("login"); setAuthDialogOpen(true); }}
              className={`inline-flex items-center gap-1.5 ${loginButtonClass}`}
            >
              <LogIn className="h-4 w-4" />
              登录
            </button>
          )}
        </div>
      </div>
    </nav>

    {/* K8-5-13-D：登录弹窗 — K8-5-36-A 支持 URL 参数驱动自动打开与 mode 切换 */}
    <AuthDialog
      open={authDialogOpen}
      onOpenChange={setAuthDialogOpen}
      defaultMode={authDialogMode}
      onLoginSuccess={handleAuthSuccess}
    />
    </>
  );
}
