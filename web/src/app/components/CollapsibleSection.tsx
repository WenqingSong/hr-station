import React from "react";

interface CollapsibleSectionProps {
  open: boolean;
  children: React.ReactNode;
  /** 动画时长，默认 200ms */
  duration?: number;
}

/**
 * 基于 CSS Grid Rows 的折叠/展开动画容器。
 * 原理：grid-rows-[0fr] → grid-rows-[1fr]，内层 overflow-hidden 实现高度过渡。
 * 支持 height: auto，无需 JS 测量。
 */
export function CollapsibleSection({ open, children, duration = 200 }: CollapsibleSectionProps) {
  return (
    <div
      className="grid transition-[grid-template-rows,opacity] ease-out"
      style={{
        transitionDuration: `${duration}ms`,
        gridTemplateRows: open ? '1fr' : '0fr',
      }}
    >
      <div className="overflow-hidden">
        <div className={`transition-transform ease-out ${open ? 'opacity-100 translate-y-0' : 'opacity-0 -translate-y-1'}`}
          style={{ transitionDuration: `${duration}ms` }}
        >
          {children}
        </div>
      </div>
    </div>
  );
}
