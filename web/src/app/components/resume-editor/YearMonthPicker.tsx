import React, { useState, useRef, useEffect, useLayoutEffect, useCallback } from "react";
import { createPortal } from "react-dom";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { cn } from "../ui/utils";

type YearMonthPickerProps = {
  value?: string;
  onChange: (value: string) => void;
  placeholder?: string;
  disabled?: boolean;
  className?: string;
};

const MONTHS = ["1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12"];
const PANEL_WIDTH = 256;
const ESTIMATED_PANEL_HEIGHT = 320;
const VIEWPORT_PADDING = 8;

function parseYearMonth(value?: string) {
  if (!value) return { year: new Date().getFullYear(), month: 0 };
  const m = value.match(/^(\d{4})-(\d{2})$/);
  if (m) return { year: Number(m[1]), month: Number(m[2]) };
  return { year: new Date().getFullYear(), month: 0 };
}

function formatYearMonth(year: number, month: number) {
  return `${year}-${String(month).padStart(2, "0")}`;
}

export function YearMonthPicker({
  value,
  onChange,
  placeholder = "选择年月",
  disabled = false,
  className,
}: YearMonthPickerProps) {
  const [open, setOpen] = useState(false);
  const parsed = parseYearMonth(value);
  const [viewYear, setViewYear] = useState(parsed.year);
  const containerRef = useRef<HTMLDivElement>(null);
  const triggerRef = useRef<HTMLButtonElement>(null);
  const panelRef = useRef<HTMLDivElement>(null);
  const [panelPos, setPanelPos] = useState<{ left: number; top: number }>({ left: 0, top: 0 });

  // Calculate panel position relative to viewport (for position:fixed)
  const calcPanelPos = useCallback(() => {
    const trigger = triggerRef.current;
    if (!trigger) return;
    const rect = trigger.getBoundingClientRect();

    // fixed 定位：使用 viewport 坐标，不要加 scrollX/scrollY
    let left = rect.left;
    let top = rect.bottom + 4;

    // 右侧边界保护
    left = Math.min(Math.max(VIEWPORT_PADDING, left), window.innerWidth - PANEL_WIDTH - VIEWPORT_PADDING);

    // 下方空间不足时向上翻转
    if (top + ESTIMATED_PANEL_HEIGHT > window.innerHeight - VIEWPORT_PADDING) {
      top = Math.max(VIEWPORT_PADDING, rect.top - ESTIMATED_PANEL_HEIGHT - 4);
    }

    setPanelPos({ left, top });
  }, []);

  // 打开时用 useLayoutEffect 计算位置，避免面板先出现在 (0,0) 再跳动
  useLayoutEffect(() => {
    if (!open) return;
    calcPanelPos();
  }, [open, calcPanelPos]);

  // 打开期间监听 scroll / resize 实时重定位
  useEffect(() => {
    if (!open) return;
    window.addEventListener("scroll", calcPanelPos, true);
    window.addEventListener("resize", calcPanelPos);
    return () => {
      window.removeEventListener("scroll", calcPanelPos, true);
      window.removeEventListener("resize", calcPanelPos);
    };
  }, [open, calcPanelPos]);

  // 外部点击关闭 — 同时排除 trigger 容器和 portal 面板内部点击
  useEffect(() => {
    if (!open) return;
    function handler(e: MouseEvent) {
      const target = e.target as Node;
      const insideTrigger = containerRef.current?.contains(target) ?? false;
      const insidePanel = panelRef.current?.contains(target) ?? false;
      if (!insideTrigger && !insidePanel) {
        setOpen(false);
      }
    }
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, [open]);

  // Sync viewYear when value changes
  useEffect(() => {
    if (value) setViewYear(parseYearMonth(value).year);
  }, [value]);

  function selectMonth(monthNum: number) {
    onChange(formatYearMonth(viewYear, monthNum));
    setOpen(false);
  }

  function clear() {
    onChange("");
    setOpen(false);
  }

  function selectCurrent() {
    const now = new Date();
    onChange(formatYearMonth(now.getFullYear(), now.getMonth() + 1));
    setOpen(false);
  }

  const display = value ? `${parsed.year}.${String(parsed.month).padStart(2, "0")}` : "";

  return (
    <div ref={containerRef} className={cn("relative", className)}>
      {/* Trigger */}
      <button
        ref={triggerRef}
        type="button"
        disabled={disabled}
        onClick={() => !disabled && setOpen((prev) => !prev)}
        className={cn(
          "flex h-10 w-[136px] items-center rounded-xl border border-slate-200 bg-white px-3 text-sm text-left transition-all duration-200",
          "focus:outline-none focus:ring-4 focus:ring-[#0052D9]/10",
          disabled ? "cursor-not-allowed bg-gray-50 text-slate-400" : "cursor-pointer hover:border-slate-300",
        )}
      >
        <span className={cn("flex-1 truncate", !display && "text-slate-400")}>
          {display || placeholder}
        </span>
        <svg className="ml-2 h-3.5 w-3.5 flex-shrink-0 text-slate-400" viewBox="0 0 12 12" fill="none">
          <path d="M1 4l5 5 5-5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
        </svg>
      </button>

      {/* Panel — rendered to body via portal to escape stacking context */}
      {open &&
        createPortal(
          <div
            ref={panelRef}
            className="fixed w-[256px] rounded-2xl border border-slate-200 bg-white p-3 shadow-xl"
            style={{ left: panelPos.left, top: panelPos.top, zIndex: 99999 }}
          >
          {/* Year nav */}
          <div className="flex items-center justify-between mb-3">
            <button
              type="button"
              onClick={() => setViewYear((y) => y - 1)}
              className="flex h-7 w-7 items-center justify-center rounded-lg text-slate-500 hover:bg-slate-100 hover:text-slate-700"
              aria-label="上一年"
            >
              <ChevronLeft className="h-4 w-4" />
            </button>
            <span className="text-sm font-semibold text-slate-800">{viewYear}</span>
            <button
              type="button"
              onClick={() => setViewYear((y) => y + 1)}
              className="flex h-7 w-7 items-center justify-center rounded-lg text-slate-500 hover:bg-slate-100 hover:text-slate-700"
              aria-label="下一年"
            >
              <ChevronRight className="h-4 w-4" />
            </button>
          </div>

          {/* Month grid */}
          <div className="grid grid-cols-4 gap-1.5 mb-3">
            {MONTHS.map((m, i) => {
              const isSelected = parsed.month === i + 1 && parsed.year === viewYear;
              return (
                <button
                  key={m}
                  type="button"
                  onClick={() => selectMonth(i + 1)}
                  className={cn(
                    "h-9 rounded-lg text-xs font-medium transition-colors",
                    isSelected
                      ? "bg-[#0052D9] text-white"
                      : "text-slate-600 hover:bg-slate-100",
                  )}
                >
                  {m}月
                </button>
              );
            })}
          </div>

          {/* Footer */}
          <div className="flex items-center justify-between border-t border-slate-100 pt-2">
            <button
              type="button"
              onClick={clear}
              className="text-xs text-slate-400 hover:text-slate-600 transition-colors px-1"
            >
              清除
            </button>
            <button
              type="button"
              onClick={selectCurrent}
              className="text-xs text-[#0052D9] hover:text-[#0047BD] font-medium transition-colors px-1"
            >
              本月
            </button>
          </div>
        </div>,
        document.body,
      )}
    </div>
  );
}
