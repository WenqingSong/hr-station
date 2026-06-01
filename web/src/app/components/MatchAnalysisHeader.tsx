import React from "react";
import { ChevronLeft, Briefcase, FileText, Activity } from "lucide-react";
import { useNavigate } from "react-router";
import { toJobDetailViewModel, toResumeSummaryViewModel } from "../../adapters";
import type { JobDetail, Resume } from "../../types";

interface MatchAnalysisHeaderProps {
  job?: JobDetail;
  resume?: Resume;
  jobSource?: string;
  resumeSource?: string;
}

export function MatchAnalysisHeader({ job, resume, jobSource = "目标岗位", resumeSource = "当前简历" }: MatchAnalysisHeaderProps) {
  const navigate = useNavigate();
  const jobVm = toJobDetailViewModel(job);
  const resumeVm = toResumeSummaryViewModel(resume);

  return (
    <div className="bg-white rounded-lg p-5 border border-gray-200 shadow-sm flex flex-col gap-4">
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 tracking-tight mb-1.5">岗位匹配分析</h1>
          <p className="text-sm text-gray-500">
            基于目标岗位 JD 和当前简历，分析匹配度、证据链、风险短板和投递前优化优先级。
          </p>
        </div>
        <button 
          onClick={() => navigate(-1)}
          className="flex items-center text-[13px] text-gray-500 hover:text-[#0052D9] transition-colors bg-gray-50 hover:bg-blue-50 px-3 py-1.5 rounded-md border border-gray-100"
        >
          <ChevronLeft className="w-4 h-4 mr-0.5" />
          返回岗位详情
        </button>
      </div>

      <div className="flex flex-wrap items-center gap-6 pt-4 border-t border-gray-100 text-[13px]">
        <div className="flex items-center gap-2">
          <span className="text-gray-400">目标岗位：</span>
          <div className="flex items-center text-gray-800 font-medium">
            <Briefcase className="w-3.5 h-3.5 mr-1.5 text-[#0052D9]" />
            {jobVm.title}
          </div>
        </div>
        <div className="w-px h-3 bg-gray-200"></div>
        <div className="flex items-center gap-2">
          <span className="text-gray-500">分析来源：</span>
          <span className="text-[#0052D9] font-medium bg-blue-50 px-2 py-0.5 rounded-sm text-[12px]">基于目标岗位与当前简历</span>
        </div>

        <div className="w-px h-3 bg-gray-200"></div>
        
        <div className="flex items-center gap-2">
          <span className="text-gray-400">分析状态：</span>
          <div className="flex items-center text-gray-600 bg-gray-50 px-2 py-0.5 rounded border border-gray-100">
            <Activity className="w-3.5 h-3.5 mr-1.5 text-gray-400" />
            分析完成
          </div>
        </div>
      </div>
    </div>
  );
}
