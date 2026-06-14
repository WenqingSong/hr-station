import type { ReactNode } from "react";

// ─── Types ──────────────────────────────────────────────────────────────

type CardVariant = "default" | "soft";

export interface SectionCardProps {
  children: ReactNode;
  title?: string;
  description?: string;
  icon?: ReactNode;
  extra?: ReactNode;
  variant?: CardVariant;
  className?: string;
  titleClassName?: string;
  descriptionClassName?: string;
}

// ─── Component ───────────────────────────────────────────────────────────

const variantStyles: Record<CardVariant, string> = {
  default: "bg-white border border-slate-200/70",
  soft: "bg-slate-50/80 border border-slate-200/50",
};

export function SectionCard({
  children,
  title,
  description,
  icon,
  extra,
  variant = "default",
  className = "",
  titleClassName = "",
  descriptionClassName = "",
}: SectionCardProps) {
  return (
    <div className={`rounded-2xl shadow-sm p-5 sm:p-6 ${variantStyles[variant]} ${className}`}>
      {(title || icon || extra) && (
        <div className="flex items-center gap-2.5 mb-4">
          {icon && <span className="w-8 h-8 rounded-lg bg-slate-100 flex items-center justify-center shrink-0">{icon}</span>}
          <div className="flex-1 min-w-0">
            {title && <h3 className={`text-[14px] font-semibold text-slate-950 ${titleClassName}`}>{title}</h3>}
            {description && <p className={`text-[13px] text-slate-500 leading-6 mt-0.5 ${descriptionClassName}`}>{description}</p>}
          </div>
          {extra && <div className="shrink-0">{extra}</div>}
        </div>
      )}
      {children}
    </div>
  );
}
