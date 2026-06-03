import type { ReactNode } from "react";

/**
 * K8-5-2-B：统一右侧 AI 辅助栏 Sticky Shell（不接入页面，仅创建组件）
 *
 * 职责：
 * - 右侧 AI 辅助栏的 sticky 行为、宽度安全、最大高度、响应式降级
 * - 不关心具体 AI 内容
 *
 * 桌面端（xl+：≥1280px）：
 * - position: sticky, top 偏移避开导航栏
 * - max-height: calc(100vh - header space), overflow-y: auto
 *
 * 中屏/移动端（<xl）：
 * - position: static，正常流
 *
 * 注意：
 * - 不使用 fixed / absolute
 * - 不脱离布局流
 * - 不在此组件内渲染 Header / Navbar
 *
 * K8-5-3：断点从 lg 改为 xl，仅大屏 sticky，避免中屏挤压主内容。
 */
export interface AISidebarShellProps {
  /** 右侧 AI 模块内容 */
  children: ReactNode;
  /** 附加 className */
  className?: string;
  /** 桌面端 sticky 的 top 偏移（默认 xl:top-24，匹配导航栏高度） */
  stickyTopClassName?: string;
}

export function AISidebarShell({
  children,
  className = "",
  stickyTopClassName = "xl:top-24",
}: AISidebarShellProps) {
  return (
    <aside
      className={[
        "w-full",
        "min-w-0",
        // 桌面端 sticky + 高度约束 + 内部滚动（K8-5-3：xl 断点）
        "xl:sticky",
        stickyTopClassName,
        "xl:max-h-[calc(100vh-7rem)]",
        "xl:overflow-y-auto",
        // 避免内容贴边
        "xl:pr-1",
        className,
      ]
        .filter(Boolean)
        .join(" ")}
    >
      {children}
    </aside>
  );
}
