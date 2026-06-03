import type { ReactNode } from "react";

/**
 * K8-5-2-B：统一右侧 AI 辅助栏 Sticky Shell（不接入页面，仅创建组件）
 *
 * 职责：
 * - 右侧 AI 辅助栏的 sticky 行为、宽度安全、最大高度、响应式降级
 * - 不关心具体 AI 内容
 *
 * 桌面端（lg+）：
 * - position: sticky, top 偏移避开导航栏
 * - max-height: calc(100vh - header space), overflow-y: auto
 *
 * 移动端（<lg）：
 * - position: static，正常流
 *
 * 注意：
 * - 不使用 fixed / absolute
 * - 不脱离布局流
 * - 不在此组件内渲染 Header / Navbar
 */
export interface AISidebarShellProps {
  /** 右侧 AI 模块内容 */
  children: ReactNode;
  /** 附加 className */
  className?: string;
  /** 桌面端 sticky 的 top 偏移（默认 lg:top-24，匹配导航栏高度） */
  stickyTopClassName?: string;
}

export function AISidebarShell({
  children,
  className = "",
  stickyTopClassName = "lg:top-24",
}: AISidebarShellProps) {
  return (
    <aside
      className={[
        "w-full",
        "min-w-0",
        // 桌面端 sticky + 高度约束 + 内部滚动
        "lg:sticky",
        stickyTopClassName,
        "lg:max-h-[calc(100vh-7rem)]",
        "lg:overflow-y-auto",
        // 避免内容贴边
        "lg:pr-1",
        className,
      ]
        .filter(Boolean)
        .join(" ")}
    >
      {children}
    </aside>
  );
}
