import React from "react";
import { Link } from "react-router";
import { Sparkles, Loader2 } from "lucide-react";

export interface ResumeEditorSaveCardProps {
  mode: "add" | "edit";
  saving: boolean;
  successMessage: string;
  errorMessage: string;
  onSave: () => void;
}

/**
 * Bottom save card for the resume editor.
 *
 * NOT fixed — sits at the bottom of the page flow.
 *
 * Left: explanation text
 * Right: cancel link + save button
 * Below: success/error messages
 *
 * Preserves the existing save logic (handleSave) from MyResumesPage.
 */
export function ResumeEditorSaveCard({
  mode,
  saving,
  successMessage,
  errorMessage,
  onSave,
}: ResumeEditorSaveCardProps) {
  const isAdd = mode === "add";

  return (
    <section className="rounded-2xl bg-white ring-1 ring-gray-100 p-6">
      <div className="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
        <div className="flex-1">
          <h2 className="text-[16px] font-semibold text-slate-900 mb-1">
            {isAdd ? "创建简历" : "保存简历"}
          </h2>
          <p className="text-[13px] text-slate-500 leading-relaxed">
            保存后，Offerr鹅 会基于最新保存版本进行岗位推荐、匹配分析和优化建议。
          </p>
        </div>
        <div className="flex items-center gap-3 flex-shrink-0">
          <Link
            to="/resumes"
            className="px-4 py-2.5 text-slate-500 text-sm hover:text-slate-800 transition-colors"
          >
            取消
          </Link>
          <button
            className="rounded-xl bg-[#0052D9] px-6 py-2.5 text-sm font-medium text-white hover:bg-[#0047BD] transition-colors disabled:opacity-50 shadow-sm flex items-center gap-1.5"
            onClick={onSave}
            disabled={saving}
          >
            {saving && <Loader2 className="w-4 h-4 animate-spin" />}
            {saving ? "保存中…" : isAdd ? "创建简历" : "保存简历"}
          </button>
        </div>
      </div>

      {/* Success message */}
      {successMessage && (
        <div className="mt-4 rounded-xl bg-green-50 ring-1 ring-green-200 p-3 text-sm text-green-700 flex items-center gap-2">
          <Sparkles className="w-4 h-3.5 flex-shrink-0" />
          {successMessage}
        </div>
      )}

      {/* Error message */}
      {errorMessage && (
        <div className="mt-4 rounded-xl bg-red-50 ring-1 ring-red-200 p-3 text-sm text-red-600 flex items-center gap-2">
          <Sparkles className="w-4 h-4 flex-shrink-0" />
          {errorMessage}
        </div>
      )}
    </section>
  );
}
