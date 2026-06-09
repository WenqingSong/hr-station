import React, { useRef, useLayoutEffect, useState } from "react";
import { cn } from "./utils";

export interface AnimatedListItemProps {
  children: React.ReactNode;
  /** When true, triggers exit animation */
  removing?: boolean;
  className?: string;
}

/**
 * Animated list item with enter/exit transitions.
 *
 * Enter: CSS @keyframes (opacity 0→1 + translateY 6px→0), 180ms, no scale.
 * Exit:  content fade-out (80ms) + measured-height collapse (160ms).
 *
 * No scale transforms (avoids sub-pixel text blur).
 * No max-h:2000px — uses getBoundingClientRect for real height.
 */
export function AnimatedListItem({
  children,
  removing = false,
  className,
}: AnimatedListItemProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const [exitStyle, setExitStyle] = useState<React.CSSProperties>({});
  // Phase 0 = idle / entering, 1 = height locked (measured), 2 = collapsing
  const [exitPhase, setExitPhase] = useState(0);
  const prevRemovingRef = useRef(false);

  useLayoutEffect(() => {
    if (removing && !prevRemovingRef.current) {
      // Entering exit: measure, lock height, then collapse
      const el = containerRef.current;
      if (!el) return;
      const realH = el.getBoundingClientRect().height;
      setExitPhase(1);
      // Lock height at current real value (no transition yet)
      setExitStyle({ height: `${realH}px` });
      // Next frame: trigger collapse with transition
      requestAnimationFrame(() => {
        setExitPhase(2);
        setExitStyle({
          height: "0px",
          marginTop: "0px",
          marginBottom: "0px",
          transition: "height 160ms cubic-bezier(0.22, 1, 0.36, 1), margin 160ms cubic-bezier(0.22, 1, 0.36, 1)",
        });
      });
    } else if (!removing) {
      setExitPhase(0);
      setExitStyle({});
    }
    prevRemovingRef.current = removing;
  }, [removing]);

  return (
    <div
      ref={containerRef}
      className={cn(
        "overflow-hidden",
        !removing && "resume-item-enter",
        className,
      )}
      style={exitStyle}
      aria-hidden={removing || undefined}
    >
      <div
        style={
          exitPhase === 2
            ? {
                animation: "resume-item-exit-content 80ms cubic-bezier(0.22, 1, 0.36, 1) both",
              }
            : undefined
        }
      >
        {children}
      </div>
    </div>
  );
}
