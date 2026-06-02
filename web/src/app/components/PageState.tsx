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
    <div className={`${maxWidth} mx-auto px-6 py-16 text-center ${className}`}>
      <Loader2 className="w-6 h-6 text-[#0052D9] mx-auto mb-3 animate-spin" />
      <p className="text-sm text-gray-400">{message}</p>
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
  children?: React.ReactNode;
}

export function PageError({
  title = "数据加载失败",
  description = "请稍后重试。",
  maxWidth = "max-w-[1440px]",
  className = "",
  children,
}: PageErrorProps) {
  return (
    <div className={`${maxWidth} mx-auto px-6 py-16 text-center ${className}`}>
      <AlertCircle className="w-10 h-10 text-gray-300 mx-auto mb-3" />
      <p className="text-gray-500 text-sm mb-1">{title}</p>
      {description && <p className="text-gray-400 text-xs mb-4">{description}</p>}
      {children}
    </div>
  );
}

// ─── PageEmpty ──────────────────────────────────────────────────────────────
//
// 统一的页面级空状态（无数据）。
// 替代各页面的“暂无数据”内联块。

interface PageEmptyProps {
  title?: string;
  description?: string;
  maxWidth?: string;
  className?: string;
  children?: React.ReactNode;
}

export function PageEmpty({
  title = "暂无数据",
  description,
  maxWidth = "max-w-[1440px]",
  className = "",
  children,
}: PageEmptyProps) {
  return (
    <div className={`${maxWidth} mx-auto px-6 py-16 text-center ${className}`}>
      <FileText className="w-10 h-10 text-gray-300 mx-auto mb-3" />
      <p className="text-gray-500 text-sm mb-1">{title}</p>
      {description && <p className="text-gray-400 text-xs mb-4">{description}</p>}
      {children}
    </div>
  );
}
