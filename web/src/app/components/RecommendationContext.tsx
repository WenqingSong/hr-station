import React from "react";
import { FileText, Sparkles, AlertCircle } from "lucide-react";
import { Badge } from "./ui/Badge";
import { mockCurrentResume } from "../mocks";
import { toResumeSummaryViewModel } from "../adapters";

const resumeVm = toResumeSummaryViewModel(mockCurrentResume);

interface RecommendationContextProps {
  resumeName?: string;
  sourceLabel?: string;
}

export function RecommendationContext({ resumeName = resumeVm.name, sourceLabel = "默认简历" }: RecommendationContextProps) {
  return (
    <aside className="w-[320px] shrink-0 flex flex-col gap-4">
      {/* Resume Card */}
      <div className="bg-white border border-gray-200 rounded-lg p-4 shadow-sm">
        <div className="flex justify-between items-start mb-3">
          <div className="flex items-center gap-2 text-gray-500 text-[13px]">
            <FileText className="h-4 w-4" />
            当前用于推荐的简历
          </div>
        </div>
        <div className="mb-3">
          <h3 className="font-bold text-gray-900 text-[15px] mb-2">{resumeName}</h3>
          <div className="flex flex-wrap gap-1.5">
            {[resumeVm.directionLabel, ...resumeVm.skillTags.slice(0, 4)].map((tag) => (
              <Badge key={tag} variant="outline" className="text-[11px] px-1.5 text-gray-600 bg-gray-50 font-normal">
                {tag}
              </Badge>
            ))}
          </div>
        </div>
        <div className="flex items-center gap-2 text-[12px]">
          <span className="text-gray-400">推荐来源：</span>
          <span className="text-[#0052D9] font-medium">{sourceLabel}</span>
        </div>
      </div>

      {/* Basis Card */}
      <div className="bg-white border border-gray-200 rounded-lg p-5 shadow-sm">
        <h3 className="font-bold text-gray-900 mb-4 text-[15px]">本次推荐依据</h3>
        <ul className="space-y-3 text-[13px] text-gray-600">
          <li className="flex items-start">
            <span className="w-[70px] shrink-0 text-gray-400">求职方向：</span>
            <span className="font-medium text-gray-800">{resumeVm.directionLabel}</span>
          </li>
          <li className="flex items-start">
            <span className="w-[70px] shrink-0 text-gray-400">核心技能：</span>
            <span className="font-medium text-gray-800">{resumeVm.skillTags.slice(0, 3).join(' / ')}</span>
          </li>
          <li className="flex items-start">
            <span className="w-[70px] shrink-0 text-gray-400">项目经历：</span>
            <span className="font-medium text-gray-800 leading-relaxed">{resumeVm.projectCount} 个项目</span>
          </li>
          <li className="flex items-start pt-2 border-t border-gray-50">
            <span className="w-[70px] shrink-0 text-gray-400">目标岗位池：</span>
            <span className="font-medium text-gray-800">软件开发方向</span>
          </li>
        </ul>
      </div>

      {/* Top Fix Preview */}
      <div className="bg-amber-50/30 border border-amber-100/60 rounded-lg p-5 shadow-sm">
        <div className="flex items-center gap-2 mb-3">
          <AlertCircle className="h-4 w-4 text-amber-500" />
          <h3 className="font-bold text-amber-900 text-[14px]">影响推荐排序的 3 个因素</h3>
        </div>
        <ul className="space-y-2 text-[12px] text-amber-800/80 list-disc pl-5 marker:text-amber-300">
          {resumeVm.gapItems.slice(0, 3).map((gap, i) => (
            <li key={i}>{gap}</li>
          ))}
        </ul>
      </div>

      {/* Mock Status */}
      <div className="mt-2 flex justify-center">
        <div className="inline-flex items-center gap-1.5 px-3 py-1.5 text-gray-400 text-[11px] opacity-70">
          <Sparkles className="h-3.5 w-3.5" />
          当前结果来源：Mock 演示
        </div>
      </div>
    </aside>
  );
}
