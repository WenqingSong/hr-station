import React from "react";
import { ChevronLeft, Briefcase, FileText, Activity } from "lucide-react";
import { useNavigate } from "react-router";
import { toJobDetailViewModelFromApi, toResumeSummaryViewModel } from "../../adapters";
import type { JobApiItem } from "../../api/jobs";
import type { Resume } from "../../types";

interface ResumeOptimizationHeaderProps {
  job?: JobApiItem;
  resume?: Resume;
  jobSource?: string;
  resumeSource?: string;
}

export function ResumeOptimizationHeader({ job, resume, jobSource = "目标岗位", resumeSource = "当前简历" }: ResumeOptimizationHeaderProps) {
  const navigate = useNavigate();
  const jobVm = toJobDetailViewModelFromApi(job);
  const resumeVm = toResumeSummaryViewModel(resume);

  return (
    <div className="bg-white rounded-lg p-4 border border-gray-200 shadow-sm flex flex-col gap-3">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <h1 className="text-[18px] font-bold text-gray-900 tracking-tight">简历优化</h1>
          <div className="h-3.5 w-px bg-gray-300"></div>
          <p className="text-[13px] text-gray-500">
            基于目标岗位 JD 和当前简历，生成投递前最关键的修复任务和核实清单。
          </p>
        </div>
        <button 
          onClick={() => navigate(-1)}
          className="flex items-center text-[12px] text-gray-600 hover:text-[#0052D9] transition-colors bg-white hover:bg-blue-50 px-3 py-1.5 rounded border border-gray-200 shadow-sm"
        >
          <ChevronLeft className="w-3.5 h-3.5 mr-0.5" />
          返回岗位匹配分析
        </button>
      </div>

      <div className="flex flex-wrap items-center gap-5 bg-gray-50/80 px-3 py-2 rounded text-[12px] border border-gray-100">
        <div className="flex items-center gap-1.5">
          <span className="text-gray-400">目标岗位：</span>
          <div className="flex items-center text-gray-700 font-medium">
            <Briefcase className="w-3 h-3 mr-1 text-[#0052D9]" />
            {jobVm.title}
          </div>
        </div>
        <div className="w-px h-3 bg-gray-200"></div>
        <div className="flex items-center gap-1.5">
          <span className="text-gray-500">优化来源：</span>
          <span className="text-[#0052D9] font-medium bg-blue-50 px-1.5 py-0.5 rounded-sm">基于匹配分析结果</span>
        </div>
        <div className="w-px h-3 bg-gray-200"></div>
        <div className="flex items-center gap-1.5">
          <span className="text-gray-400">当前简历：</span>
          <div className="flex items-center text-gray-700 font-medium">
            <FileText className="w-3 h-3 mr-1 text-[#0052D9]" />
            {resumeVm.name}
          </div>
        </div>
        <div className="w-px h-3 bg-gray-200"></div>
        <div className="flex items-center gap-1.5">
          <span className="text-gray-500">简历来源：</span>
          <span className="text-[#0052D9] font-medium bg-blue-50 px-1.5 py-0.5 rounded-sm">{resumeSource}</span>
        </div>
        <div className="w-px h-3 bg-gray-200"></div>
        <div className="flex items-center gap-1.5">
          <span className="text-gray-400">优化状态：</span>
          <div className="flex items-center text-gray-500">
            <Activity className="w-3 h-3 mr-1 text-gray-400" />
            优化完成
          </div>
        </div>
      </div>
    </div>
  );
}