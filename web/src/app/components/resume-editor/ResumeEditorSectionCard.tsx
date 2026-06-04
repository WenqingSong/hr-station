import React, { useState } from "react";
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
  children,
}: ResumeEditorSectionCardProps) {
  const [open, setOpen] = useState(defaultOpen);
  const isCollapsed = collapsible && !open;

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
        onClick={() => collapsible && setOpen(!open)}
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

      {/* ── Description (when collapsed, show hint) ──────────────── */}
      {description && !open && collapsible && (
        <div className="px-6 pb-1">
          <p className="text-[13px] text-slate-400 line-clamp-1">{description}</p>
        </div>
      )}

      {/* ── Body ────────────────────────────────────────────────── */}
      {!isCollapsed && (
        <div className="px-6 pb-6">
          {description && (
            <p className="text-[13px] text-slate-500 mb-5 -mt-2">{description}</p>
          )}
          <div className={cn(!description && "pt-2")}>{children}</div>
        </div>
      )}
    </section>
  );
}
