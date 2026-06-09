import { useState, useEffect, useRef } from "react";
import { FileText, LogIn, LogOut } from "lucide-react";
import { Link, useNavigate, useLocation, useSearchParams } from "react-router";
import { logout } from "../../api/auth";
import { useAuth } from "../../stores/authStore";
import { AuthDialog } from "./auth/AuthDialog";
import { normalizeRedirect } from "../../utils/redirect";

const HERO_NAV_THRESHOLD = 280;
const HERO_BLEND_START_Y = 220;
const HERO_BLEND_END_Y = 80;
const SHOW_TOP_Y = 20;
const HIDE_START_Y = 96;
const SCROLL_DELTA = 8;

function clamp(v: number, min = 0, max = 1) { return Math.min(max, Math.max(min, v)); }
function getHeroProgress(scrollY: number, isHeroRoute: boolean) {
  if (!isHeroRoute) return 0;
  return clamp((HERO_BLEND_START_Y - scrollY) / (HERO_BLEND_START_Y - HERO_BLEND_END_Y));
}

export function Navbar() {
  const navigate = useNavigate();
  const location = useLocation();
  const [searchParams, setSearchParams] = useSearchParams();
  const { isAuthenticated, clearAuth } = useAuth();
  const [hidden, setHidden] = useState(false);
  const [isHeroTop, setIsHeroTop] = useState(false);
  const [navHeroProgress, setNavHeroProgress] = useState(0);
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

          // ── 1. 视觉状态永远优先更新，不受 delta 守卫阻挡 ──
          setIsHeroTop(isHeroRoute && currentScrollY < HERO_NAV_THRESHOLD);
          setNavHeroProgress(getHeroProgress(currentScrollY, isHeroRoute));

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

          // ── 2. 页面顶部强制显示 ──
          if (currentScrollY <= SHOW_TOP_Y) {
            setHidden(false);
            lastScrollY.current = currentScrollY;
            ticking.current = false;
            return;
          }

          const delta = currentScrollY - lastScrollY.current;

          // ── 3. 微小滚动只跳过显隐，不跳过视觉 ──
          if (Math.abs(delta) < SCROLL_DELTA) {
            lastScrollY.current = currentScrollY;
            ticking.current = false;
            return;
          }

          // ── 4. 向下滚动，超过起始距离后隐藏 ──
          if (delta > 0 && currentScrollY > HIDE_START_Y) {
            setHidden(true);
          }

          // ── 5. 向上滚动，显示 ──
          if (delta < 0) {
            setHidden(false);
          }

          lastScrollY.current = currentScrollY;
          ticking.current = false;
        });
        ticking.current = true;
      }
    };

    // Immediately set initial state
    setIsHeroTop(isHeroRoute && window.scrollY < HERO_NAV_THRESHOLD);
    setNavHeroProgress(getHeroProgress(window.scrollY, isHeroRoute));

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

  // ════════ Visual: progress-driven transition ════════
  const isHeroText = navHeroProgress >= 0.55;

  const brandClass = isHeroText
    ? "text-white"
    : "text-slate-950";

  const linkClass = isHeroText
    ? "text-white/85 hover:text-white"
    : "text-slate-700 hover:text-slate-950";

  const logoutClass = isHeroText
    ? "text-white/70 hover:text-white"
    : "text-slate-500 hover:text-slate-800";

  const loginBtn = isHeroText
    ? "border-white/30 text-white/90 hover:bg-white/12 hover:text-white"
    : "border-slate-200/70 text-slate-700 hover:bg-white/80 hover:text-slate-950";

  return (
    <>
    <nav
      className={[
        "fixed inset-x-0 top-0 z-50",
        "h-16 md:h-[68px]",
        "transition-transform duration-300 ease-[cubic-bezier(0.22,1,0.36,1)]",
        hidden ? "-translate-y-[110%]" : "translate-y-0",
      ].join(" ")}
    >
      {/* ── solidSurface: white background ── */}
      <div
        aria-hidden="true"
        className="pointer-events-none absolute inset-0 bg-white shadow-sm"
        style={{ opacity: 1 - navHeroProgress }}
      />
      {/* ── heroSurface: dark translucent ── */}
      <div
        aria-hidden="true"
        className="pointer-events-none absolute inset-0 bg-slate-950/28 backdrop-blur-[2px]"
        style={{ opacity: navHeroProgress }}
      />
      {/* ── solid border ── */}
      <div
        aria-hidden="true"
        className="pointer-events-none absolute inset-x-0 bottom-0 h-px bg-slate-200/70"
        style={{ opacity: 1 - navHeroProgress }}
      />
      {/* ── hero border ── */}
      <div
        aria-hidden="true"
        className="pointer-events-none absolute inset-x-0 bottom-0 h-px bg-white/5"
        style={{ opacity: navHeroProgress }}
      />

      {/* ── Content ── */}
      <div className="relative z-10 mx-auto flex h-full max-w-[1440px] items-center justify-between px-6 md:px-8">
        <Link
          to="/jobs"
          className={`font-brand text-[22px] font-extrabold tracking-[-0.03em] transition-colors duration-200 ${brandClass}`}
          aria-label="Offer捕手首页"
        >
          Offer捕手
        </Link>

        <div className="flex items-center gap-5">
          {isAuthenticated ? (
            <>
              <Link
                to="/resumes"
                className={`inline-flex items-center gap-1.5 text-sm font-medium transition-colors duration-200 ${linkClass}`}
              >
                <FileText className="h-4 w-4" />
                我的简历
              </Link>
              <button
                onClick={handleLogout}
                className={`inline-flex items-center gap-1.5 text-sm font-medium transition-colors duration-200 ${logoutClass}`}
              >
                <LogOut className="h-4 w-4" />
                退出
              </button>
            </>
          ) : (
            <button
              type="button"
              onClick={() => { setAuthDialogMode("login"); setAuthDialogOpen(true); }}
              className={`inline-flex items-center gap-1.5 rounded-full border px-4 py-2 text-sm font-medium transition duration-200 ${loginBtn}`}
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
