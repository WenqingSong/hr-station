import React from "react";
import { FileText, CheckCircle2, AlertCircle } from "lucide-react";
import { Badge } from "../ui/Badge";
import { toResumeSummaryViewModel } from "../../../adapters";
import type { Resume } from "../../../types";

// ── Helpers ──

function formatDateTime(value?: string | null): string {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`;
}

interface MatchAnalysisSidebarRightProps {
  resume?: Resume;
  resumeSource?: string;
}

export function MatchAnalysisSidebarRight({ resume, resumeSource = "当前简历" }: MatchAnalysisSidebarRightProps) {
  const resumeVm = toResumeSummaryViewModel(resume);
  return (
    <aside className="w-[320px] shrink-0 flex flex-col gap-4">
      {/* Resume Card */}
      <div className="bg-white border border-gray-200 rounded-lg p-5 shadow-sm">
        <div className="flex justify-between items-start mb-4">
          <div className="flex items-center gap-2 text-gray-500 text-[13px]">
            <FileText className="h-4 w-4" />
            当前简历摘要
          </div>
          <button className="text-[12px] text-[#0052D9] hover:underline">
            切换简历
          </button>
        </div>
        <div className="mb-4">
          <h3 className="font-bold text-gray-900 text-[16px] mb-3">{resumeVm.name}</h3>
          <div className="flex flex-wrap gap-1.5">
            {[resumeVm.directionLabel, ...resumeVm.skillTags.slice(0, 4)].map((tag) => (
              <Badge key={tag} variant="outline" className="text-[11px] px-1.5 text-gray-600 bg-gray-50 font-normal">
                {tag}
              </Badge>
            ))}
          </div>
        </div>
        <p className="text-[12px] text-gray-400">
          包含 {resumeVm.projectCount} 个项目经历，最近更新：{formatDateTime(resumeVm.updatedAt) || "—"}
        </p>

        <div className="mt-5 pt-4 border-t border-gray-100">
          <h4 className="font-bold text-gray-900 text-[13px] mb-3 flex items-center gap-1.5">
            <CheckCircle2 className="w-3.5 h-3.5 text-green-600" />
            简历核心证据
          </h4>
          <ul className="space-y-2 text-[12px] text-gray-600">
            {resumeVm.skillTags.slice(0, 5).map((item, i) => (
              <li key={i} className="flex items-center gap-2">
                <div className="w-1 h-1 rounded-full bg-gray-300"></div>
                {item}
              </li>
            ))}
          </ul>
        </div>

        <div className="mt-5 pt-4 border-t border-gray-100">
          <h4 className="font-bold text-gray-900 text-[13px] mb-3 flex items-center gap-1.5">
            <AlertCircle className="w-3.5 h-3.5 text-amber-500" />
            当前简历缺口
          </h4>
          <ul className="space-y-2 text-[12px] text-amber-700/90">
            {resumeVm.gapItems.slice(0, 3).map((item, i) => (
              <li key={i} className="flex items-center gap-2">
                <div className="w-1.5 h-1.5 rounded-full bg-amber-400"></div>
                {item}
              </li>
            ))}
          </ul>
        </div>
      </div>

    </aside>
  );
}
