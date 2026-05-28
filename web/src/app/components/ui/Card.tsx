import type { CSSProperties, ReactNode, MouseEvent } from "react";
import { color, radius, shadow } from "../../tokens";

interface CardProps {
  children: ReactNode;
  style?: CSSProperties;
  hoverable?: boolean;
  padding?: string;
  onClick?: () => void;
}

/**
 * 统一卡片基础组件
 * 所有页面内容区域统一使用此组件承载内容
 */
export function Card({ children, style, hoverable = false, padding, onClick }: CardProps) {
  const handleMouseEnter = (e: MouseEvent<HTMLDivElement>) => {
    if (hoverable) {
      const el = e.currentTarget as HTMLDivElement;
      el.style.boxShadow = shadow.cardHover;
      el.style.borderColor = "#C8D8F8";
    }
  };
  const handleMouseLeave = (e: MouseEvent<HTMLDivElement>) => {
    if (hoverable) {
      const el = e.currentTarget as HTMLDivElement;
      el.style.boxShadow = shadow.card;
      el.style.borderColor = color.border;
    }
  };

  return (
    <div
      onClick={onClick}
      onMouseEnter={handleMouseEnter}
      onMouseLeave={handleMouseLeave}
      style={{
        background: color.cardBg,
        borderRadius: radius.card,
        border: `1px solid ${color.border}`,
        boxShadow: shadow.card,
        padding: padding ?? "20px 24px",
        cursor: onClick || hoverable ? "pointer" : undefined,
        transition: "box-shadow 0.15s, border-color 0.15s",
        ...style,
      }}
    >
      {children}
    </div>
  );
}

/**
 * 卡片内部的 header 行（标题 + 右侧操作区）
 */
interface CardHeaderProps {
  title: string;
  subtitle?: string;
  extra?: ReactNode;
}

export function CardHeader({ title, subtitle, extra }: CardHeaderProps) {
  return (
    <div
      style={{
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
        marginBottom: 16,
        paddingBottom: 14,
        borderBottom: `1px solid ${color.divider}`,
      }}
    >
      <div>
        <span style={{ fontSize: 15, fontWeight: 600, color: color.textPrimary }}>
          {title}
        </span>
        {subtitle && (
          <span style={{ fontSize: 12, color: color.textTertiary, marginLeft: 8 }}>
            {subtitle}
          </span>
        )}
      </div>
      {extra && <div style={{ fontSize: 13, color: color.textSecondary }}>{extra}</div>}
    </div>
  );
}

/**
 * 卡片内部的 Divider
 */
export function CardDivider() {
  return (
    <div
      style={{
        height: 1,
        background: color.divider,
        margin: "14px 0",
      }}
    />
  );
}
