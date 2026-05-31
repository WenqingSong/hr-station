import * as React from "react";
import { cn } from "../../utils/cn";

interface TagProps extends React.HTMLAttributes<HTMLSpanElement> {
  variant?: "default" | "blue" | "outline";
}

export function Tag({ className, variant = "default", ...props }: TagProps) {
  const variants = {
    default: "bg-slate-100 text-slate-700",
    blue: "bg-blue-50 text-blue-700",
    outline: "border border-slate-200 text-slate-600 bg-white",
  };

  return (
    <span
      className={cn(
        "inline-flex items-center px-2.5 py-0.5 rounded text-xs font-medium",
        variants[variant],
        className
      )}
      {...props}
    />
  );
}
