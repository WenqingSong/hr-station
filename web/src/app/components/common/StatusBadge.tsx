import type { ReactNode } from "react";

// ─── Types ──────────────────────────────────────────────────────────────

type BadgeTone = "blue" | "emerald" | "amber" | "rose" | "slate";

export interface StatusBadgeProps {
  children: ReactNode;
  tone?: BadgeTone;
  className?: string;
}

// ─── Style maps ─────────────────────────────────────────────────────────

const toneStyles: Record<BadgeTone, string> = {
  blue: "bg-blue-50 text-blue-700 border-blue-200",
  emerald: "bg-emerald-50 text-emerald-700 border-emerald-200",
  amber: "bg-amber-50 text-amber-700 border-amber-200",
  rose: "bg-red-50 text-red-600 border-red-200",
  slate: "bg-slate-50 text-slate-500 border-slate-200",
};

// ─── Component ───────────────────────────────────────────────────────────

export function StatusBadge({ children, tone = "slate", className = "" }: StatusBadgeProps) {
  return (
    <span className={`inline-flex items-center text-xs font-medium rounded-full px-2.5 py-0.5 border ${toneStyles[tone]} ${className}`}>
      {children}
    </span>
  );
}
