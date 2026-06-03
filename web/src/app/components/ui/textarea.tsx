import * as React from "react";

import { cn } from "./utils";

export interface TextareaProps
  extends React.TextareaHTMLAttributes<HTMLTextAreaElement> {
  error?: boolean;
}

function Textarea({ className, error = false, ...props }: TextareaProps) {
  return (
    <textarea
      data-slot="textarea"
      aria-invalid={error ? true : props["aria-invalid"] || undefined}
      className={cn(
        "min-h-[120px] w-full rounded-xl border border-slate-200 bg-white px-4 py-3 text-sm text-slate-900 shadow-none outline-none transition-all duration-200 ease-out",
        "placeholder:text-slate-400",
        "hover:border-slate-300",
        "focus:border-[#0052D9] focus:ring-4 focus:ring-[#0052D9]/10",
        "disabled:cursor-not-allowed disabled:bg-slate-50 disabled:text-slate-400 disabled:opacity-70",
        "resize-y",
        error &&
          "border-red-300 focus:border-red-500 focus:ring-red-500/10",
        className,
      )}
      {...props}
    />
  );
}

export { Textarea };
