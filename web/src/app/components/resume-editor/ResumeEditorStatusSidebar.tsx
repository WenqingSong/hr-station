import React from "react";
import {
  CheckCircle2,
  AlertCircle,
  Clock,
  Sparkles,
  FileText,
  Search,
  Zap,
  BarChart3,
} from "lucide-react";

export interface ResumeEditorStatusSidebarProps {
  completeness: number;
  /** Whether the resume has enough info for analysis */
  aiReady: boolean;
  /** Count of required fields still missing */
  missingRequired: number;
  /** Count of recommended fields to fill */
  recommendedToFill: number;
  /** Function statuses */
  aiFunctions?: {
    recommendation: boolean;
    matchAnalysis: boolean;
    resumeOptimization: boolean;
    versionComparison: boolean;
  };
  /** Whether currently saving */
  saving?: boolean;
}

/**
 * Right sidebar for the resume editor.
 *
 * Shows completion status, readiness, missing required/recommended fields,
 * and function availability.
 *
 * Desktop: sticky on right. Mobile: placed above the main form.
 */
export function ResumeEditorStatusSidebar({
  completeness,
  aiReady,
  missingRequired,
  recommendedToFill,
  aiFunctions,
  saving,
}: ResumeEditorStatusSidebarProps) {
  return (
    <div className="bg-white rounded-2xl ring-1 ring-gray-100 overflow-hidden">
      {/* ── Header ──────────────────────────────────────────────── */}
      <div className="px-5 py-4 border-b border-gray-100">
        <h3 className="text-sm font-semibold text-slate-900">简历状态</h3>
      </div>

      <div className="px-5 py-4 space-y-4">
        {/* Completion */}
        <div>
          <div className="flex items-center justify-between mb-2">
            <span className="text-xs text-slate-500">完成度</span>
            <span className="text-sm font-semibold text-[#0052D9]">
              {completeness}%
            </span>
          </div>
          <div className="w-full h-1.5 bg-gray-100 rounded-full overflow-hidden">
            <div
              className="h-full bg-[#0052D9] rounded-full transition-all duration-300"
              style={{ width: `${Math.max(completeness, 5)}%` }}
            />
          </div>
        </div>

        {/* Readiness */}
        <div
          className={`flex items-center gap-2 rounded-xl px-3 py-2 text-xs ${
            aiReady
              ? "bg-green-50 text-green-700 ring-1 ring-green-100"
              : "bg-orange-50 text-orange-600 ring-1 ring-orange-100"
          }`}
        >
          {aiReady ? (
            <CheckCircle2 className="w-3.5 h-3.5 flex-shrink-0" />
          ) : (
            <AlertCircle className="w-3.5 h-3.5 flex-shrink-0" />
          )}
          <span>{aiReady ? "信息完整，可用于分析" : "当前信息尚不完整"}</span>
        </div>

        {/* Missing / recommended */}
        <div className="space-y-2">
          {missingRequired > 0 && (
            <div className="flex items-center justify-between text-xs">
              <span className="text-slate-500 flex items-center gap-1.5">
                <AlertCircle className="w-3 h-3 text-red-400" />
                必填缺失
              </span>
              <span className="font-medium text-red-600">{missingRequired} 项</span>
            </div>
          )}
          {recommendedToFill > 0 && (
            <div className="flex items-center justify-between text-xs">
              <span className="text-slate-500 flex items-center gap-1.5">
                <Clock className="w-3 h-3 text-amber-400" />
                推荐补充
              </span>
              <span className="font-medium text-amber-600">
                {recommendedToFill} 项
              </span>
            </div>
          )}
          {missingRequired === 0 && recommendedToFill === 0 && (
            <div className="text-xs text-slate-400">
              所有必填项已完成，推荐项也已补充。
            </div>
          )}
        </div>

        {/* Divider */}
        <div className="border-t border-gray-100" />

        {/* Functions */}
        {aiFunctions && (
          <div className="space-y-2.5">
            <span className="text-xs font-medium text-slate-500">
              四核心功能状态
            </span>
            <FunctionItem
              icon={<Sparkles className="w-3.5 h-3.5" />}
              label="岗位推荐"
              ready={aiFunctions.recommendation}
            />
            <FunctionItem
              icon={<Search className="w-3.5 h-3.5" />}
              label="岗位匹配分析"
              ready={aiFunctions.matchAnalysis}
            />
            <FunctionItem
              icon={<Zap className="w-3.5 h-3.5" />}
              label="简历优化"
              ready={aiFunctions.resumeOptimization}
            />
            <FunctionItem
              icon={<BarChart3 className="w-3.5 h-3.5" />}
              label="版本对比"
              ready={aiFunctions.versionComparison}
            />
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

function FunctionItem({
  icon,
  label,
  ready,
}: {
  icon: React.ReactNode;
  label: string;
  ready: boolean;
}) {
  return (
    <div className="flex items-center justify-between text-xs">
      <span className="text-slate-500 flex items-center gap-1.5">
        {icon}
        {label}
      </span>
      {ready ? (
        <CheckCircle2 className="w-3 h-3 text-green-500" />
      ) : (
        <Clock className="w-3 h-3 text-gray-300" />
      )}
    </div>
  );
}
