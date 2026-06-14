import type { ReactNode } from "react";
import { Info } from "lucide-react";

// ─── Types ──────────────────────────────────────────────────────────────

type NoticeVariant = "inline" | "block";
type NoticeTone = "slate" | "blue";

export interface AiNoticeProps {
  children: ReactNode;
  variant?: NoticeVariant;
  tone?: NoticeTone;
  className?: string;
}

// ─── Component ───────────────────────────────────────────────────────────

const toneStyles: Record<NoticeTone, string> = {
  slate: "bg-slate-50 border-slate-200 text-slate-500",
  blue: "bg-blue-50/50 border-blue-100 text-slate-600",
};

export function AiNotice({ children, variant = "block", tone = "slate", className = "" }: AiNoticeProps) {
  const base = `flex items-start gap-3 border rounded-2xl ${toneStyles[tone]} ${className}`;
  const isInline = variant === "inline";

  return (
    <div className={`${base} ${isInline ? "p-2 px-3" : "p-4"}`}>
      <div className={`rounded-lg bg-slate-200 flex items-center justify-center shrink-0 ${isInline ? "w-6 h-6 mt-0" : "w-7 h-7 mt-0.5"}`}>
        <Info className={`text-slate-400 ${isInline ? "w-3 h-3" : "w-3.5 h-3.5"}`} />
      </div>
      <p className={`${isInline ? "text-xs leading-5" : "text-xs leading-6"}`}>
        {children}
      </p>
    </div>
  );
}
