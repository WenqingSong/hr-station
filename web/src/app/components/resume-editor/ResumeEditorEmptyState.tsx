import React from "react";

export interface ResumeEditorEmptyStateProps {
  title: string;
  description: string;
  actionLabel: string;
  onAction: () => void;
}

/**
 * Unified empty state card for repeatable list modules.
 *
 * Shows centered title + description + action button in a
 * rounded-2xl bg-slate-50/80 container with min-height.
 * Enter animation uses CSS transition (opacity + translateY).
 */
export function ResumeEditorEmptyState({
  title,
  description,
  actionLabel,
  onAction,
}: ResumeEditorEmptyStateProps) {
  return (
    <div
      className="rounded-xl bg-slate-50/80 border border-slate-100 min-h-[180px] flex flex-col items-center justify-center text-center px-6 py-8"
    >
      <p className="text-sm text-slate-500 mb-1">{title}</p>
      <p className="text-xs text-slate-400 mb-4">{description}</p>
      <button
        type="button"
        onClick={onAction}
        className="text-[13px] font-medium text-[#0052D9] hover:underline"
      >
        {actionLabel}
      </button>
    </div>
  );
}
