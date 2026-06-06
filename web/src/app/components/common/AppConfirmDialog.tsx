import React from "react";
import { Loader2 } from "lucide-react";
import { cn } from "../ui/utils";

interface AppConfirmDialogProps {
  open: boolean;
  title: string;
  description?: React.ReactNode;
  confirmText?: string;
  cancelText?: string;
  variant?: "default" | "danger";
  loading?: boolean;
  hideCancel?: boolean;
  onConfirm: () => void | Promise<void>;
  onCancel: () => void;
}

export function AppConfirmDialog({
  open, title, description, confirmText = "确认", cancelText = "取消",
  variant = "default", loading = false, hideCancel = false, onConfirm, onCancel,
}: AppConfirmDialogProps) {
  if (!open) return null;

  return (
    <div className="fixed inset-0 z-[9999] flex items-center justify-center">
      {/* Backdrop */}
      <div className="fixed inset-0 bg-slate-950/30 backdrop-blur-md" onClick={onCancel} />
      {/* Card */}
      <div className="relative z-10 w-[min(90vw,400px)] rounded-3xl bg-white shadow-[0_24px_80px_rgba(15,23,42,0.22)] p-6">
        <h2 className="text-base font-semibold text-slate-900 mb-2">{title}</h2>
        {description && (
          <div className="text-sm text-slate-500 leading-relaxed mb-5">{description}</div>
        )}
        <div className="flex items-center justify-end gap-3">
          {!hideCancel && (
            <button
              type="button"
              disabled={loading}
              onClick={onCancel}
              className="px-4 py-2 text-sm font-medium text-slate-600 bg-slate-100 hover:bg-slate-200 rounded-xl transition-colors disabled:opacity-50"
            >
              {cancelText}
            </button>
          )}
          <button
            type="button"
            disabled={loading}
            onClick={onConfirm}
            className={cn(
              "px-4 py-2 text-sm font-medium text-white rounded-xl transition-colors inline-flex items-center gap-1.5 disabled:opacity-50",
              variant === "danger"
                ? "bg-red-600 hover:bg-red-700"
                : "bg-[#0052D9] hover:bg-[#0047BD]"
            )}
          >
            {loading && <Loader2 className="w-3.5 h-3.5 animate-spin" />}
            {confirmText}
          </button>
        </div>
      </div>
    </div>
  );
}
