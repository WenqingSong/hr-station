import { useState, useEffect, useRef } from "react";
import { FileText, LogIn, LogOut } from "lucide-react";
import { Link, useNavigate, useLocation, useSearchParams } from "react-router";
import { logout } from "../../api/auth";
import { useAuth } from "../../stores/authStore";
import { AuthDialog } from "./auth/AuthDialog";
import { normalizeRedirect } from "../../utils/redirect";

const HERO_NAV_THRESHOLD = 280;

export function Navbar() {
  const navigate = useNavigate();
  const location = useLocation();
  const [searchParams, setSearchParams] = useSearchParams();
  const { isAuthenticated, clearAuth } = useAuth();
  const [hidden, setHidden] = useState(false);
  const [isHeroTop, setIsHeroTop] = useState(false);
  const [authDialogOpen, setAuthDialogOpen] = useState(false);
  const [authDialogMode, setAuthDialogMode] = useState<"login" | "register">("login");
  const lastScrollY = useRef(0);
  const ticking = useRef(false);
  const redirectAfterLoginRef = useRef<string | null>(null);

  // ── Hero route detection ──
  const isHeroRoute =
    location.pathname === "/" ||
    location.pathname === "/jobs" ||
    /^\/jobs\/[^/]+$/.test(location.pathname) ||
    /^\/job\/[^/]+$/.test(location.pathname);

  useEffect(() => {
    const auth = searchParams.get("auth");
    if (auth === "login" || auth === "register") {
      redirectAfterLoginRef.current = searchParams.get("redirect");
      setAuthDialogMode(auth === "register" ? "register" : "login");
      setAuthDialogOpen(true);
      setSearchParams({}, { replace: true });
    }
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  // 滚动：隐藏/显示 + Hero 顶部判定
  useEffect(() => {
    const handleScroll = () => {
      if (!ticking.current) {
        window.requestAnimationFrame(() => {
          const currentScrollY = window.scrollY;
          const isResumeEditorRoute =
            location.pathname === "/resumes/new" ||
            location.pathname === "/resumes/edit" ||
            location.pathname.startsWith("/resumes/edit?id=");

          if (isResumeEditorRoute) {
            setHidden(false);
            lastScrollY.current = currentScrollY;
            ticking.current = false;
            return;
          }

          const delta = currentScrollY - lastScrollY.current;
          if (Math.abs(delta) < 30) {
            lastScrollY.current = currentScrollY;
            ticking.current = false;
            return;
          }

          if (delta > 0 && currentScrollY > 80) {
            setHidden(true);
          } else if (delta < 0) {
            setHidden(false);
          }
          if (currentScrollY <= 20) {
            setHidden(false);
          }

          // Hero 顶部透明状态
          setIsHeroTop(isHeroRoute && currentScrollY < HERO_NAV_THRESHOLD);

          lastScrollY.current = currentScrollY;
          ticking.current = false;
        });
        ticking.current = true;
      }
    };

    // Immediately set initial state
    setIsHeroTop(isHeroRoute && window.scrollY < HERO_NAV_THRESHOLD);

    window.addEventListener("scroll", handleScroll, { passive: true });
    return () => window.removeEventListener("scroll", handleScroll);
  }, [location.pathname, isHeroRoute]);

  const handleAuthSuccess = () => {
    if (redirectAfterLoginRef.current) {
      const target = normalizeRedirect(redirectAfterLoginRef.current, "/jobs");
      redirectAfterLoginRef.current = null;
      navigate(target, { replace: true });
    }
  };

  async function handleLogout() {
    try { await logout(); } catch (_) { /* ignore if API fails */ }
    clearAuth();
    navigate("/");
  }

  // ════════ Visual: two modes ════════
  const navSurface = isHeroTop
    ? "bg-slate-950/28 border-white/5 text-white backdrop-blur-[2px] shadow-none"
    : "bg-white border-slate-200/70 text-slate-950 shadow-sm";

  const brandClass = isHeroTop
    ? "text-white"
    : "text-slate-950";

  const linkClass = isHeroTop
    ? "text-white/85 hover:text-white"
    : "text-slate-700 hover:text-slate-950";

  const logoutClass = isHeroTop
    ? "text-white/70 hover:text-white"
    : "text-slate-500 hover:text-slate-800";

  const loginBtn = isHeroTop
    ? "border-white/30 text-white/90 hover:bg-white/12 hover:text-white"
    : "border-slate-200/70 text-slate-700 hover:bg-white/80 hover:text-slate-950";

  return (
    <>
    <nav
      className={[
        "fixed inset-x-0 top-0 z-50",
        "h-16 md:h-[68px]",
        "border-b",
        "transition-colors duration-200",
        "transition-transform duration-300 ease-[cubic-bezier(0.22,1,0.36,1)]",
        hidden ? "-translate-y-[110%]" : "translate-y-0",
        navSurface,
      ].join(" ")}
    >
      <div className="mx-auto flex h-full max-w-[1440px] items-center justify-between px-6 md:px-8">
        <Link
          to="/jobs"
          className={`font-brand text-[22px] font-extrabold tracking-[-0.03em] ${brandClass}`}
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
                className={`inline-flex items-center gap-1.5 text-sm font-medium transition-colors ${logoutClass}`}
              >
                <LogOut className="h-4 w-4" />
                退出
              </button>
            </>
          ) : (
            <button
              type="button"
              onClick={() => { setAuthDialogMode("login"); setAuthDialogOpen(true); }}
              className={`inline-flex items-center gap-1.5 rounded-full border px-4 py-2 text-sm font-medium transition ${loginBtn}`}
            >
              <LogIn className="h-4 w-4" />
              登录
            </button>
          )}
        </div>
      </div>
    </nav>

    <AuthDialog
      open={authDialogOpen}
      onOpenChange={setAuthDialogOpen}
      defaultMode={authDialogMode}
      onLoginSuccess={handleAuthSuccess}
    />
    </>
  );
}
