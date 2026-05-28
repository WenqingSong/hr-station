import type { ReactNode, MouseEvent } from "react";
import { color, radius } from "../../tokens";

type ButtonVariant = "primary" | "outline" | "ghost";
type ButtonSize = "md" | "sm";

interface ButtonProps {
  children: ReactNode;
  variant?: ButtonVariant;
  size?: ButtonSize;
  onClick?: () => void;
  disabled?: boolean;
  style?: React.CSSProperties;
}

const sizeMap = {
  md: { height: 36, padding: "0 18px", fontSize: 14 },
  sm: { height: 30, padding: "0 14px", fontSize: 13 },
};

/**
 * 统一按钮组件
 * variant:
 *   primary  — 蓝底白字（主操作）
 *   outline  — 白底蓝边蓝字（次操作）
 *   ghost    — 文字按钮（轻操作）
 */
export function Button({
  children,
  variant = "primary",
  size = "md",
  onClick,
  disabled = false,
  style,
}: ButtonProps) {
  const sz = sizeMap[size];

  const base: React.CSSProperties = {
    display: "inline-flex",
    alignItems: "center",
    justifyContent: "center",
    height: sz.height,
    padding: sz.padding,
    fontSize: sz.fontSize,
    fontWeight: 500,
    borderRadius: radius.btn,
    border: "none",
    cursor: disabled ? "not-allowed" : "pointer",
    transition: "background 0.15s, color 0.15s",
    opacity: disabled ? 0.45 : 1,
    lineHeight: 1,
    ...style,
  };

  const variantStyle: React.CSSProperties =
    variant === "primary"
      ? { color: "#fff", background: color.blue }
      : variant === "outline"
      ? {
          color: color.blue,
          background: "transparent",
          border: `1px solid ${color.blue}`,
        }
      : { color: color.textSecondary, background: "transparent" };

  const handleMouseEnter = (e: MouseEvent<HTMLButtonElement>) => {
    if (disabled) return;
    const el = e.currentTarget;
    if (variant === "primary") el.style.background = color.blueDark;
    else if (variant === "outline") el.style.background = color.blueLight;
    else el.style.color = color.blue;
  };

  const handleMouseLeave = (e: MouseEvent<HTMLButtonElement>) => {
    if (disabled) return;
    const el = e.currentTarget;
    if (variant === "primary") el.style.background = color.blue;
    else if (variant === "outline") el.style.background = "transparent";
    else el.style.color = color.textSecondary;
  };

  return (
    <button
      onClick={disabled ? undefined : onClick}
      onMouseEnter={handleMouseEnter}
      onMouseLeave={handleMouseLeave}
      style={{ ...base, ...variantStyle }}
    >
      {children}
    </button>
  );
}
