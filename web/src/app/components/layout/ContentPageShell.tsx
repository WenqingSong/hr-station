import type { ReactNode } from "react";

/**
 * K8-5-15-C：普通内容页统一 Shell
 *
 * 职责：
 * - 提供 Navbar offset（pt-24）
 * - 统一背景 bg-[#F5F7FB]
 * - 统一 max-width 和左右 padding
 *
 * 不适用于 Hero 页面（/jobs、/jobs/:id）。
 */
export type ContentPageShellProps = {
  children: ReactNode;
  className?: string;
  maxWidthClassName?: string;
  contentClassName?: string;
};

export function ContentPageShell({
  children,
  className = "",
  maxWidthClassName = "max-w-[1280px]",
  contentClassName = "",
}: ContentPageShellProps) {
  return (
    <div className={`min-h-screen bg-[#F5F7FB] pt-24 pb-12 px-4 sm:px-6 lg:px-8 ${className}`}>
      <div className={`mx-auto w-full ${maxWidthClassName} ${contentClassName}`}>
        {children}
      </div>
    </div>
  );
}
