// @5-3.1 — 统一页面容器，提供一致的最大宽度 / padding / 间距。
// 后续 5-3.3/5-3.4 业务页大规模接入前先建立组件。

import React from "react";

interface PageContainerProps {
  /** 最大宽度，默认 1360px */
  maxWidth?: string;
  /** 上边距 */
  paddingY?: string;
  /** 左右 padding */
  paddingX?: string;
  children: React.ReactNode;
  className?: string;
}

export function PageContainer({
  maxWidth = "max-w-[1360px]",
  paddingY = "py-6",
  paddingX = "px-6",
  children,
  className = "",
}: PageContainerProps) {
  return (
    <main className={`mx-auto w-full ${maxWidth} ${paddingY} ${paddingX} ${className}`}>
      {children}
    </main>
  );
}
