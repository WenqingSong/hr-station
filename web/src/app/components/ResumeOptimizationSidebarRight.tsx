import React from "react";
import { FileText, Sparkles, Target, TrendingUp } from "lucide-react";
import { Badge } from "./ui/Badge";
import { mockCurrentResume } from "../mocks";
import { toResumeSummaryViewModel } from "../adapters";

const resumeVm = toResumeSummaryViewModel(mockCurrentResume);

export function ResumeOptimizationSidebarRight() {
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
          包含 {resumeVm.projectCount} 个项目经历，最近更新：{resumeVm.updatedAt}
        </p>

        <div className="mt-5 pt-4 border-t border-gray-100">
          <h4 className="font-bold text-gray-900 text-[13px] mb-3 flex items-center gap-1.5">
            <Target className="w-3.5 h-3.5 text-[#0052D9]" />
            当前简历核心内容
          </h4>
          <ul className="space-y-2 text-[12px] text-gray-600 leading-relaxed">
            <li className="flex gap-2">
              <span className="text-gray-400 shrink-0">求职方向：</span>
              <span className="text-gray-800 font-medium">{resumeVm.directionLabel}</span>
            </li>
            <li className="flex gap-2">
              <span className="text-gray-400 shrink-0">技能关键词：</span>
              <span className="text-gray-800">{resumeVm.skillTags.slice(0, 4).join('、')}</span>
            </li>
            <li className="flex gap-2">
              <span className="text-gray-400 shrink-0">项目经历：</span>
              <span className="text-gray-800">{resumeVm.projectCount} 个</span>
            </li>
            <li className="flex gap-2">
              <span className="text-gray-400 shrink-0">简历完整度：</span>
              <span className="text-gray-800">{resumeVm.completeness}%</span>
            </li>
          </ul>
        </div>

        <div className="mt-5 pt-4 border-t border-gray-100">
          <h4 className="font-bold text-gray-700 text-[13px] mb-3 flex items-center gap-1.5">
            <TrendingUp className="w-3.5 h-3.5 text-gray-400" />
            本次优化预计影响
          </h4>
          <ul className="space-y-2 text-[12px] text-gray-500">
            {[
              "后台开发匹配度提升",
              "JD 关键词覆盖更完整",
              "项目经历表达更贴近岗位要求",
              "投递前风险降低"
            ].map((item, i) => (
              <li key={i} className="flex items-center gap-2">
                <div className="w-1 h-1 rounded-full bg-gray-300"></div>
                {item}
              </li>
            ))}
          </ul>
        </div>
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