import * as React from "react";
import { cn } from "../../utils/cn";
import { Sparkles, Wrench, AlertTriangle, PlayCircle, Info } from "lucide-react";

export type AIStateType = "generated" | "repaired" | "fallback" | "mock" | "abnormal";

interface BadgeProps extends React.HTMLAttributes<HTMLDivElement> {
  state: AIStateType;
  children?: React.ReactNode;
}

const config = {
  generated: {
    icon: Sparkles,
    styles: "bg-purple-50 text-purple-600 border-purple-100",
    defaultText: "LLM 生成",
  },
  repaired: {
    icon: Wrench,
    styles: "bg-blue-50 text-blue-600 border-blue-100",
    defaultText: "已自动修复",
  },
  fallback: {
    icon: AlertTriangle,
    styles: "bg-orange-50 text-orange-600 border-orange-100",
    defaultText: "兜底结果",
  },
  mock: {
    icon: PlayCircle,
    styles: "bg-slate-100 text-slate-500 border-slate-200",
    defaultText: "Mock 演示",
  },
  abnormal: {
    icon: Info,
    styles: "bg-red-50 text-red-600 border-red-100",
    defaultText: "异常可用",
  },
};

export function AIBadge({ state, children, className, ...props }: BadgeProps) {
  const { icon: Icon, styles, defaultText } = config[state];
  return (
    <div
      className={cn(
        "inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-medium border",
        styles,
        className
      )}
      {...props}
    >
      <Icon className="w-3 h-3" />
      <span>{children || defaultText}</span>
    </div>
  );
}
