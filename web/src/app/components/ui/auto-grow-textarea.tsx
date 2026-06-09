import React, { useRef, useEffect } from "react";
import { cn } from "./utils";

export interface AutoGrowTextareaProps
  extends Omit<React.TextareaHTMLAttributes<HTMLTextAreaElement>, "rows"> {
  /** Minimum number of visible rows before overflow (used to calculate min-height) */
  minRows?: number;
  /** Maximum number of visible rows before scroll (prevents infinite growth) */
  maxRows?: number;
  error?: boolean;
}

/**
 * Self-growing textarea — auto-expands as content grows.
 *
 * Uses ref + scrollHeight to calculate required height.
 * Maintains the same visual style (rounded-xl / border / focus ring) as Textarea.
 * No internal scrollbar when content fits within maxRows.
 */
function AutoGrowTextarea({
  className,
  error = false,
  minRows = 2,
  maxRows = 12,
  value,
  onChange,
  ...props
}: AutoGrowTextareaProps) {
  const ref = useRef<HTMLTextAreaElement>(null);

  // Single-line height estimate: 1rem (14px) + 1.5rem (py-3) + border
  const lineHeight = 20; // approximate px for text-sm leading-relaxed
  const paddingY = 24;   // py-3 = 12px top + 12px bottom
  const minHeight = minRows * lineHeight + paddingY;

  const syncHeight = () => {
    const el = ref.current;
    if (!el) return;
    // Reset to auto so scrollHeight recalculates correctly on shrink
    el.style.height = "auto";
    const scrollH = el.scrollHeight;
    const maxH = maxRows * lineHeight + paddingY;
    const h = Math.max(minHeight, Math.min(scrollH, maxH));
    el.style.height = `${h}px`;
    // Internal scrollbar only when content exceeds maxRows
    el.style.overflowY = scrollH > maxH ? "auto" : "hidden";
  };

  // Sync on value change
  useEffect(() => {
    syncHeight();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [value]);

  // Sync on mount for initial content
  useEffect(() => {
    syncHeight();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <textarea
      ref={ref}
      data-slot="auto-grow-textarea"
      aria-invalid={error ? true : props["aria-invalid"] || undefined}
      value={value}
      onChange={onChange}
      rows={1} // base rows, height is controlled by JS
      className={cn(
        // Keep same visual as Textarea
        "w-full rounded-xl border border-slate-200 bg-white px-4 py-3 text-sm text-slate-900 shadow-none outline-none transition-all duration-200 ease-out",
        "placeholder:text-slate-400",
        "hover:border-slate-300",
        "focus:border-[#0052D9] focus:ring-4 focus:ring-[#0052D9]/10",
        "disabled:cursor-not-allowed disabled:bg-slate-50 disabled:text-slate-400 disabled:opacity-70",
        "resize-none",
        error &&
          "border-red-300 focus:border-red-500 focus:ring-red-500/10",
        className,
      )}
      style={{
        minHeight: `${minHeight}px`,
        maxHeight: `${maxRows * lineHeight + paddingY}px`,
      }}
      {...props}
    />
  );
}

export { AutoGrowTextarea };
