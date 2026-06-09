import React from "react";

type BulletTone = "blue" | "green" | "amber" | "red" | "purple" | "slate";

interface BulletListProps {
  items: Array<string | React.ReactNode>;
  tone?: BulletTone;
  size?: "xs" | "sm" | "md";
  className?: string;
  itemClassName?: string;
}

const toneClassMap: Record<BulletTone, string> = {
  blue: "text-blue-600",
  green: "text-emerald-600",
  amber: "text-amber-600",
  red: "text-red-600",
  purple: "text-purple-600",
  slate: "text-slate-500",
};

const dotSizeClassMap: Record<"xs" | "sm" | "md", string> = {
  xs: "h-1 w-1 mt-[0.35em]",
  sm: "h-1.5 w-1.5 mt-[0.38em]",
  md: "h-2 w-2 mt-[0.42em]",
};

const textSizeClassMap: Record<"xs" | "sm" | "md", string> = {
  xs: "text-[11px]",
  sm: "text-[12px]",
  md: "text-[13px]",
};

export function BulletList({
  items,
  tone = "blue",
  size = "md",
  className = "",
  itemClassName = "",
}: BulletListProps) {
  if (!items || items.length === 0) return null;

  const dotColorClass = toneClassMap[tone] || toneClassMap.blue;
  const dotSizeClass = dotSizeClassMap[size] || dotSizeClassMap.md;

  return (
    <ul className={`space-y-1.5 ${className}`}>
      {items.map((item, index) => (
        <BulletItem
          key={index}
          tone={tone}
          size={size}
          className={itemClassName}
        >
          {item}
        </BulletItem>
      ))}
    </ul>
  );
}

interface BulletItemProps {
  children: React.ReactNode;
  tone?: BulletTone;
  size?: "xs" | "sm" | "md";
  className?: string;
}

export function BulletItem({
  children,
  tone = "blue",
  size = "md",
  className = "",
}: BulletItemProps) {
  const dotColorClass = toneClassMap[tone] || toneClassMap.blue;
  const dotSizeClass = dotSizeClassMap[size] || dotSizeClassMap.md;
  const textSizeClass = textSizeClassMap[size] || textSizeClassMap.md;

  return (
    <li className={`flex items-start gap-2.5 leading-6 ${className}`}>
      <span
        aria-hidden="true"
        className={`shrink-0 rounded-full bg-current ${dotColorClass} ${dotSizeClass}`}
      />
      <span className={`min-w-0 flex-1 ${textSizeClass} text-gray-700`}>
        {children}
      </span>
    </li>
  );
}
