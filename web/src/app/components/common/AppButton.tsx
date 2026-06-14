import type { ReactNode, ButtonHTMLAttributes } from "react";

// ─── Types ──────────────────────────────────────────────────────────────

type ButtonVariant = "primary" | "secondary" | "ghost";

export interface AppButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  children: ReactNode;
  icon?: ReactNode;
}

// ─── Style maps ─────────────────────────────────────────────────────────

const variantStyles: Record<ButtonVariant, string> = {
  primary:
    "bg-[#0052D9] text-white hover:bg-[#0047BD] disabled:opacity-60",
  secondary:
    "bg-slate-100 text-slate-700 border border-slate-200 hover:bg-slate-200 disabled:opacity-60",
  ghost:
    "bg-transparent text-slate-500 hover:text-slate-700 hover:bg-slate-50 disabled:opacity-60",
};

// ─── Component ───────────────────────────────────────────────────────────

export function AppButton({
  variant = "primary",
  children,
  icon,
  className = "",
  disabled,
  ...rest
}: AppButtonProps) {
  return (
    <button
      className={`inline-flex items-center justify-center gap-1.5 px-4 py-2 text-sm font-medium rounded-xl transition-colors whitespace-nowrap ${variantStyles[variant]} ${className}`}
      disabled={disabled}
      {...rest}
    >
      {icon}
      {children}
    </button>
  );
}

// Convenience aliases
export function PrimaryButton(props: Omit<AppButtonProps, "variant">) {
  return <AppButton variant="primary" {...props} />;
}

export function SecondaryButton(props: Omit<AppButtonProps, "variant">) {
  return <AppButton variant="secondary" {...props} />;
}
