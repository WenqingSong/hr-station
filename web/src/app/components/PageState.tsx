import React from "react";
import { Link } from "react-router";
import { Loader2, AlertCircle, FileText } from "lucide-react";

// ─── PageLoading ────────────────────────────────────────────────────────────
//
// 统一的页面级加载状态。
// 替代各处内联的 "正在加载…" 文本。

interface PageLoadingProps {
  message?: string;
  /** 页面最大宽度，默认 1440px */
  maxWidth?: string;
  className?: string;
}

export function PageLoading({
  message = "正在加载…",
  maxWidth = "max-w-[1440px]",
  className = "",
}: PageLoadingProps) {
  return (
    <div className={`min-h-[calc(100vh-80px)] flex items-center justify-center ${maxWidth} mx-auto px-6 ${className}`}>
      <div className="flex flex-col items-center gap-3">
        <Loader2 className="w-6 h-6 text-[#0052D9] animate-spin" />
        <p className="text-sm text-gray-400">{message}</p>
      </div>
    </div>
  );
}

// ─── PageError ──────────────────────────────────────────────────────────────
//
// 统一的页面级错误状态。
// 可传 title / description 和 action slot，默认无按钮。

interface PageErrorProps {
  title?: string;
  description?: string;
  maxWidth?: string;
  className?: string;
  /** 重试按钮回调，传入后显示"重新尝试"按钮 */
  onRetry?: () => void;
  children?: React.ReactNode;
}

export function PageError({
  title = "数据加载失败",
  description = "请稍后重试。",
  maxWidth = "max-w-[1440px]",
  className = "",
  onRetry,
  children,
}: PageErrorProps) {
  return (
    <div className={`${maxWidth} mx-auto px-6 py-16 ${className}`}>
      <div className="max-w-lg mx-auto bg-white rounded-2xl border border-slate-200 shadow-sm p-8 text-center">
        <div className="w-12 h-12 mx-auto mb-4 rounded-xl bg-red-50 flex items-center justify-center">
          <AlertCircle className="w-6 h-6 text-red-400" />
        </div>
        <h2 className="text-base font-semibold text-slate-900 mb-2">{title}</h2>
        {description && <p className="text-sm text-slate-500 mb-6 leading-relaxed">{description}</p>}
        <div className="flex items-center justify-center gap-3">
          {onRetry && (
            <button onClick={onRetry} className="px-5 py-2 bg-[#0052D9] text-white text-sm font-medium rounded-lg hover:bg-[#003DA5] transition-colors">
              重新尝试
            </button>
          )}
          {children}
        </div>
      </div>
    </div>
  );
}

// ─── PageEmpty ──────────────────────────────────────────────────────────────
//
// 统一的页面级空状态（无数据）。
// 替代各页面的"暂无数据"内联块。

interface PageEmptyProps {
  title?: string;
  description?: string;
  maxWidth?: string;
  className?: string;
  /** 主操作按钮文案 */
  actionLabel?: string;
  /** 主操作按钮回调 */
  onAction?: () => void;
  children?: React.ReactNode;
}

export function PageEmpty({
  title = "暂无数据",
  description,
  maxWidth = "max-w-[1440px]",
  className = "",
  actionLabel,
  onAction,
  children,
}: PageEmptyProps) {
  return (
    <div className={`${maxWidth} mx-auto px-6 py-16 ${className}`}>
      <div className="max-w-lg mx-auto bg-white rounded-2xl border border-slate-200 shadow-sm p-8 text-center">
        <div className="w-12 h-12 mx-auto mb-4 rounded-xl bg-blue-50 flex items-center justify-center">
          <FileText className="w-6 h-6 text-blue-400" />
        </div>
        <h2 className="text-base font-semibold text-slate-900 mb-2">{title}</h2>
        {description && <p className="text-sm text-slate-500 mb-6 leading-relaxed">{description}</p>}
        <div className="flex items-center justify-center gap-3">
          {onAction && actionLabel && (
            <button onClick={onAction} className="px-5 py-2 bg-[#0052D9] text-white text-sm font-medium rounded-lg hover:bg-[#003DA5] transition-colors">
              {actionLabel}
            </button>
          )}
          {children}
        </div>
      </div>
    </div>
  );
}
