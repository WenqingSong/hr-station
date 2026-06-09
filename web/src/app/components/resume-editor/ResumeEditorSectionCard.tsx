import React, { useState, useEffect, useRef } from "react";
import { ChevronDown } from "lucide-react";
import { cn } from "../ui/utils";

export interface ResumeEditorSectionCardProps {
  title: string;
  description?: string;
  badge?: string;
  defaultOpen?: boolean;
  collapsible?: boolean;
  required?: boolean;
  children: React.ReactNode;
  /**
   * K8-5-41-B：异步回填完成后触发自动展开的 key。
   * 当 autoOpenKey 变化时，如果用户未手动操作过，会自动同步 defaultOpen。
   */
  autoOpenKey?: string | number | null;
}

/**
 * Collapsible section card for resume editor modules.
 *
 * Supports:
 * - Collapsible sections with smooth animation
 * - Badge labels (e.g. "必填", "推荐", "可选")
 * - Required indicator (*)
 * - defaultOpen controls initial state
 * - collapsible=false for core modules that must stay open
 * - autoOpenKey + touched tracking (K8-5-41-B):
 *   autoOpenKey changes → reset open to defaultOpen
 *   user manually toggles → touched=true → blocks auto-open
 *   until autoOpenKey changes again (e.g. switching resume version)
 *
 * Visual style matches the existing resume page:
 * rounded-2xl bg-white ring-1 ring-gray-100
 */
export function ResumeEditorSectionCard({
  title,
  description,
  badge,
  defaultOpen = true,
  collapsible = true,
  required = false,
  autoOpenKey,
  children,
}: ResumeEditorSectionCardProps) {
  const [open, setOpen] = useState(defaultOpen);
  const [touched, setTouched] = useState(false);
  const prevKeyRef = useRef(autoOpenKey);
  const isCollapsed = collapsible && !open;

  // K8-5-41-B/C：autoOpenKey 变化或 defaultOpen 变化时，同步展开状态
  // - autoOpenKey 变化（切换版本）→ 完全重置
  // - defaultOpen 变化（数据回填完成）→ touched=false 时同步
  useEffect(() => {
    const keyChanged = autoOpenKey !== undefined && autoOpenKey !== prevKeyRef.current;

    if (keyChanged) {
      prevKeyRef.current = autoOpenKey;
      setOpen(defaultOpen);
      setTouched(false);
      return;
    }

    // key 没变但 defaultOpen 变了（首次渲染时数据未加载，后续回填完成）
    // → 只要用户未手动操作过，就同步 defaultOpen
    if (!touched) {
      setOpen(defaultOpen);
    }
  }, [autoOpenKey, defaultOpen, touched]);

  const handleToggle = () => {
    if (!collapsible) return;
    setTouched(true);
    setOpen(!open);
  };

  return (
    <section
      className={cn(
        "rounded-2xl bg-white ring-1 ring-gray-100 overflow-hidden transition-all",
      )}
    >
      {/* ── Header ─────────────────────────────────────────────── */}
      <div
        className={cn(
          "flex items-center justify-between px-6 py-4 select-none",
          collapsible && "cursor-pointer hover:bg-gray-50/50 transition-colors",
        )}
        onClick={handleToggle}
      >
        <div className="flex items-center gap-2.5 min-w-0">
          <h2 className="text-[16px] font-semibold text-slate-900 flex items-center gap-1.5">
            {title}
            {required && (
              <span className="text-red-400 text-sm font-normal">*</span>
            )}
          </h2>
          {badge && (
            <span
              className={cn(
                "inline-flex items-center px-2 py-0.5 rounded-full text-[11px] font-medium flex-shrink-0",
                badge === "必填"
                  ? "bg-red-50 text-red-600 ring-1 ring-red-100"
                  : badge === "推荐"
                    ? "bg-amber-50 text-amber-600 ring-1 ring-amber-100"
                    : badge === "可选"
                      ? "bg-blue-50 text-blue-500 ring-1 ring-blue-100"
                      : "bg-gray-50 text-gray-500 ring-1 ring-gray-200",
              )}
            >
              {badge}
            </span>
          )}
        </div>

        {collapsible && (
          <ChevronDown
            className={cn(
              "w-4 h-4 text-gray-400 transition-transform duration-200 flex-shrink-0 ml-3",
              open && "rotate-180",
            )}
          />
        )}
      </div>

      {/* ── Body — grid animation ─────────────────────────────────── */}
      <div
        className={cn(
          "grid overflow-anchor-none transition-[grid-template-rows] duration-[240ms] ease-[cubic-bezier(0.22,1,0.36,1)]",
          open
            ? "grid-rows-[1fr]"
            : "grid-rows-[0fr]",
        )}
        style={{ overflowAnchor: "none" }}
      >
        <div className="overflow-hidden">
          <div className="px-6 pb-5">
            {description && (
              <p className={cn(
                "text-[13px] text-slate-500 mb-5",
                open ? "" : "line-clamp-1"
              )}>
                {description}
              </p>
            )}
            <div className={cn(!description && "pt-2")}>{children}</div>
          </div>
        </div>
      </div>
    </section>
  );
}
