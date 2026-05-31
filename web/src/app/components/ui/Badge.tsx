import * as React from "react";
import { cn } from "../../utils";

export interface BadgeProps extends React.HTMLAttributes<HTMLDivElement> {
  variant?: "default" | "secondary" | "outline" | "blue" | "green";
}

function Badge({ className, variant = "default", ...props }: BadgeProps) {
  return (
    <div
      className={cn(
        "inline-flex items-center rounded-sm px-2 py-0.5 text-xs font-semibold transition-colors focus:outline-none focus:ring-2 focus:ring-blue-600 focus:ring-offset-2",
        {
          "border-transparent bg-gray-100 text-gray-800 hover:bg-gray-200": variant === "default",
          "border-transparent bg-gray-50 text-gray-600": variant === "secondary",
          "text-gray-800 border border-gray-200": variant === "outline",
          "border-transparent bg-blue-50 text-[#0052D9]": variant === "blue",
          "border-transparent bg-green-50 text-green-700": variant === "green",
        },
        className
      )}
      {...props}
    />
  );
}

export { Badge };
