import type { ReactNode } from "react";
import { color, layout } from "../../tokens";
import { AppNav } from "../AppNav";

interface PageShellProps {
  children: ReactNode;
  /** 当前激活的导航项 key */
  activeNav?: "home" | "resume" | "jobs" | "none";
}

/**
 * 所有页面的统一外壳：
 *   顶部导航 + 浅灰背景 + 居中内容区
 *
 * 用法：
 *   <PageShell activeNav="home">
 *     <PageHeader title="..." subtitle="..." />
 *     ...内容...
 *   </PageShell>
 */
export function PageShell({ children, activeNav = "none" }: PageShellProps) {
  return (
    <div style={{ minHeight: "100vh", background: color.pageBg }}>
      <AppNav activeNav={activeNav} />
      <main style={{ paddingTop: layout.navHeight }}>
        <div
          style={{
            maxWidth: layout.contentMaxWidth,
            margin: "0 auto",
            padding: "28px 24px 64px",
          }}
        >
          {children}
        </div>
      </main>
    </div>
  );
}

/**
 * 页面顶部标题区，极简，不做营销感
 */
interface PageHeaderProps {
  title: string;
  subtitle?: string;
  action?: ReactNode;
}

export function PageHeader({ title, subtitle, action }: PageHeaderProps) {
  return (
    <div
      style={{
        display: "flex",
        alignItems: "flex-end",
        justifyContent: "space-between",
        marginBottom: 20,
      }}
    >
      <div>
        <h1
          style={{
            fontSize: 20,
            fontWeight: 600,
            color: color.textPrimary,
            lineHeight: 1.4,
            marginBottom: subtitle ? 4 : 0,
          }}
        >
          {title}
        </h1>
        {subtitle && (
          <p style={{ fontSize: 13, color: color.textTertiary }}>{subtitle}</p>
        )}
      </div>
      {action && <div>{action}</div>}
    </div>
  );
}

/**
 * 模块分组标题（卡片内部或模块间）
 */
interface SectionTitleProps {
  title: string;
  extra?: ReactNode;
}

export function SectionTitle({ title, extra }: SectionTitleProps) {
  return (
    <div
      style={{
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
        marginBottom: 12,
      }}
    >
      <span style={{ fontSize: 15, fontWeight: 600, color: color.textPrimary }}>
        {title}
      </span>
      {extra && (
        <span style={{ fontSize: 13, color: color.textSecondary }}>{extra}</span>
      )}
    </div>
  );
}
