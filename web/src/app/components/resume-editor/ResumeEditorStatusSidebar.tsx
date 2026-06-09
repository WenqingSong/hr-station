import React from "react";
import { CheckCircle2, AlertCircle, ArrowRight, Sparkles } from "lucide-react";

export interface ResumeEditorStatusSidebarProps {
  /** Overall readiness score (0–100) */
  readiness: number;
  /** Per-category statuses */
  categories: ReadinessCategory[];
  /** Next-step suggestion string */
  nextStep?: string;
  /** Whether currently saving */
  saving?: boolean;
}

export interface ReadinessCategory {
  label: string;
  status: "complete" | "partial" | "missing";
  /** Optional detail text */
  detail?: string;
}

/**
 * "投递准备度" right sidebar for the resume editor.
 *
 * Shows a progress ring + category breakdown.
 * Pure UI — receives computed data via props.
 */
export function ResumeEditorStatusSidebar({
  readiness,
  categories,
  nextStep,
  saving,
}: ResumeEditorStatusSidebarProps) {
  const safeReadiness = Math.max(0, Math.min(100, Math.round(readiness)));

  return (
    <div className="bg-white rounded-2xl ring-1 ring-gray-100 overflow-hidden">
      {/* ── Header ──────────────────────────────────────────────── */}
      <div className="px-5 py-4 border-b border-gray-100">
        <h3 className="text-sm font-semibold text-slate-900">投递准备度</h3>
      </div>

      <div className="px-5 py-4 space-y-4">
        {/* Progress ring + score */}
        <div className="flex items-center gap-4">
          {/* Circular progress */}
          <div className="relative w-16 h-16 flex-shrink-0">
            <svg className="w-16 h-16 -rotate-90" viewBox="0 0 64 64">
              <circle
                cx="32" cy="32" r="27"
                fill="none"
                stroke="#f1f5f9"
                strokeWidth="5"
              />
              <circle
                cx="32" cy="32" r="27"
                fill="none"
                stroke={safeReadiness >= 100 ? "#22c55e" : safeReadiness >= 50 ? "#0052D9" : "#f59e0b"}
                strokeWidth="5"
                strokeLinecap="round"
                strokeDasharray={`${safeReadiness * 1.696} 169.6`}
                className="transition-all duration-500 ease-out"
              />
            </svg>
            <div className="absolute inset-0 flex items-center justify-center">
              <span className="text-sm font-bold text-slate-800">{safeReadiness}</span>
            </div>
          </div>

          <div className="min-w-0">
            <p className="text-[13px] font-medium text-slate-700">
              {safeReadiness >= 100
                ? "准备充分"
                : safeReadiness >= 60
                  ? "持续完善中"
                  : safeReadiness >= 30
                    ? "需要补充"
                    : "刚刚开始"}
            </p>
            <p className="text-[11px] text-slate-400 mt-0.5 leading-relaxed">
              {safeReadiness >= 100
                ? "简历信息完整，可用于投递分析。"
                : "继续补充信息，提升岗位匹配精度。"}
            </p>
          </div>
        </div>

        {/* Category list */}
        <div className="space-y-1.5">
          {categories.map((cat) => (
            <div
              key={cat.label}
              className="flex items-center justify-between py-1.5"
            >
              <div className="flex items-center gap-2 min-w-0">
                {cat.status === "complete" ? (
                  <CheckCircle2 className="w-3.5 h-3.5 text-green-500 flex-shrink-0" />
                ) : cat.status === "partial" ? (
                  <AlertCircle className="w-3.5 h-3.5 text-amber-400 flex-shrink-0" />
                ) : (
                  <div className="w-3.5 h-3.5 rounded-full border border-gray-200 flex-shrink-0" />
                )}
                <span className="text-xs text-slate-600 truncate">
                  {cat.label}
                </span>
              </div>
              <span
                className={`text-[11px] font-medium flex-shrink-0 ml-2 ${
                  cat.status === "complete"
                    ? "text-green-600"
                    : cat.status === "partial"
                      ? "text-amber-600"
                      : "text-slate-400"
                }`}
              >
                {cat.status === "complete"
                  ? "已完成"
                  : cat.status === "partial"
                    ? "待完善"
                    : "待补充"}
              </span>
            </div>
          ))}
        </div>

        {/* Next step */}
        {nextStep && (
          <div className="rounded-xl bg-blue-50 ring-1 ring-blue-100 px-3 py-2.5">
            <div className="flex items-start gap-1.5">
              <ArrowRight className="w-3.5 h-3.5 text-[#0052D9] flex-shrink-0 mt-0.5" />
              <p className="text-xs text-[#0052D9] leading-relaxed">{nextStep}</p>
            </div>
          </div>
        )}

        {/* Saving indicator */}
        {saving && (
          <div className="border-t border-gray-100 pt-3">
            <div className="flex items-center gap-2 text-xs text-blue-600">
              <div className="w-3 h-3 border-2 border-blue-400 border-t-transparent rounded-full animate-spin" />
              正在保存…
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
