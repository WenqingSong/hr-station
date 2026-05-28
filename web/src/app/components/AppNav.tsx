import { color, layout, shadow } from "../tokens";

interface NavItem {
  key: string;
  label: string;
  href: string;
}

const NAV_ITEMS: NavItem[] = [
  { key: "home",   label: "首页",   href: "#" },
  { key: "resume", label: "我的简历", href: "#resume" },
  { key: "jobs",   label: "岗位列表", href: "#jobs" },
  { key: "login",  label: "登录",   href: "#login" },
];

interface AppNavProps {
  activeNav?: string;
}

/**
 * 全局统一顶部导航
 * 风格：腾讯校招内页导航
 * - 白底 56px 高
 * - 左：Offer 捕手
 * - 右：导航项（|分隔） + 登录按钮
 */
export function AppNav({ activeNav = "home" }: AppNavProps) {
  return (
    <nav
      style={{
        position: "fixed",
        top: 0,
        left: 0,
        right: 0,
        zIndex: 200,
        height: layout.navHeight,
        background: color.cardBg,
        boxShadow: shadow.nav,
        display: "flex",
        alignItems: "center",
      }}
    >
      <div
        style={{
          maxWidth: layout.contentMaxWidth,
          margin: "0 auto",
          padding: layout.contentPadding,
          width: "100%",
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
        }}
      >
        {/* Logo */}
        <a
          href="#"
          style={{ textDecoration: "none", display: "flex", alignItems: "center", gap: 8 }}
        >
          <div
            style={{
              width: 26,
              height: 26,
              borderRadius: 6,
              background: color.blue,
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              flexShrink: 0,
            }}
          >
            <svg width="13" height="13" viewBox="0 0 14 14" fill="none">
              <path
                d="M7 1L12 4V10L7 13L2 10V4L7 1Z"
                stroke="white"
                strokeWidth="1.3"
                fill="none"
              />
              <circle cx="7" cy="7" r="1.8" fill="white" />
            </svg>
          </div>
          <span
            style={{
              fontSize: 15,
              fontWeight: 700,
              color: color.textPrimary,
              letterSpacing: "-0.2px",
            }}
          >
            Offer 捕手
          </span>
        </a>

        {/* Right */}
        <div style={{ display: "flex", alignItems: "center" }}>
          {/* When on the login page, show all 4 nav items; otherwise show only the first 3 */}
          {NAV_ITEMS
            .filter((item) => activeNav === "login" || item.key !== "login")
            .map((item, i, arr) => {
              const isActive = activeNav === item.key;
              return (
                <span key={item.key} style={{ display: "flex", alignItems: "center" }}>
                  <a
                    href={item.href}
                    style={{
                      fontSize: 14,
                      color: isActive ? color.blue : color.textSecondary,
                      fontWeight: isActive ? 600 : 400,
                      textDecoration: "none",
                      padding: "0 14px",
                      height: layout.navHeight,
                      display: "flex",
                      alignItems: "center",
                      borderBottom: isActive ? `2px solid ${color.blue}` : "2px solid transparent",
                      transition: "color 0.15s",
                    }}
                    onMouseEnter={(e) => {
                      if (!isActive)
                        (e.currentTarget as HTMLAnchorElement).style.color = color.blue;
                    }}
                    onMouseLeave={(e) => {
                      if (!isActive)
                        (e.currentTarget as HTMLAnchorElement).style.color = color.textSecondary;
                    }}
                  >
                    {item.label}
                  </a>
                  {i < arr.length - 1 && (
                    <span
                      style={{
                        display: "inline-block",
                        width: 1,
                        height: 14,
                        background: color.border,
                        margin: "0 2px",
                      }}
                    />
                  )}
                </span>
              );
            })}

          {/* Blue 登录 button — hidden when already on the login page */}
          {activeNav !== "login" && (
            <button
              style={{
                marginLeft: 16,
                height: 34,
                padding: "0 18px",
                fontSize: 13,
                fontWeight: 500,
                color: "#fff",
                background: color.blue,
                border: "none",
                borderRadius: "6px",
                cursor: "pointer",
                transition: "background 0.15s",
                flexShrink: 0,
              }}
              onMouseEnter={(e) => {
                (e.currentTarget as HTMLButtonElement).style.background = color.blueDark;
              }}
              onMouseLeave={(e) => {
                (e.currentTarget as HTMLButtonElement).style.background = color.blue;
              }}
            >
              登录
            </button>
          )}
        </div>
      </div>
    </nav>
  );
}
