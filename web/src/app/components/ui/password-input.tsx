import * as React from "react";
import { Eye, EyeOff } from "lucide-react";

import { cn } from "./utils";
import { Input, type InputProps } from "./input";

type PasswordInputProps = Omit<InputProps, "type"> & {
  revealLabel?: string;
  hideLabel?: string;
};

function PasswordInput({
  className,
  revealLabel = "显示密码",
  hideLabel = "隐藏密码",
  disabled,
  ...props
}: PasswordInputProps) {
  const [visible, setVisible] = React.useState(false);
  const Icon = visible ? EyeOff : Eye;

  return (
    <div className="relative">
      <Input
        {...props}
        disabled={disabled}
        type={visible ? "text" : "password"}
        className={cn("pr-12", className)}
      />
      <button
        type="button"
        tabIndex={-1}
        aria-label={visible ? hideLabel : revealLabel}
        disabled={disabled}
        onClick={() => setVisible((value) => !value)}
        className={cn(
          "absolute right-3 top-1/2 flex h-8 w-8 -translate-y-1/2 items-center justify-center rounded-lg text-slate-400 transition-all duration-150 ease-out",
          "hover:bg-slate-100 hover:text-slate-600 active:scale-95",
          "disabled:pointer-events-none disabled:opacity-40",
        )}
      >
        <Icon className="h-4 w-4" aria-hidden="true" />
      </button>
    </div>
  );
}

export { PasswordInput };
export type { PasswordInputProps };
