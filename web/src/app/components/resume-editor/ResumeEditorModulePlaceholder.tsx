import React from "react";
import { Sparkles } from "lucide-react";

export interface ResumeEditorModulePlaceholderProps {
  title: string;
  description: string;
  /** Which future phase will implement this module */
  phase?: string;
}

/**
 * Placeholder for a module that will be implemented in a future phase (F/G/H).
 *
 * Shows a gentle hint about what content will go here.
 */
export function ResumeEditorModulePlaceholder({
  title,
  description,
  phase,
}: ResumeEditorModulePlaceholderProps) {
  return (
    <div className="rounded-xl bg-blue-50/50 border border-blue-100/50 px-4 py-5 text-center">
      <Sparkles className="w-4 h-4 text-[#0052D9]/30 mx-auto mb-2" />
      <p className="text-sm font-medium text-slate-600 mb-1">{title}</p>
      <p className="text-xs text-slate-400 leading-relaxed max-w-md mx-auto">
        {description}
      </p>
      {phase && (
        <p className="mt-2 text-[11px] text-slate-300">{phase}</p>
      )}
    </div>
  );
}
