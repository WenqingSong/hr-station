import React from "react";
import { ChevronLeft, Briefcase, FileText, Activity } from "lucide-react";
import { useNavigate } from "react-router";
import { mockSelectedJob, mockCurrentResume } from "../mocks";
import { toJobDetailViewModel, toResumeSummaryViewModel } from "../adapters";

const jobVm = toJobDetailViewModel(mockSelectedJob);
const resumeVm = toResumeSummaryViewModel(mockCurrentResume);

export function ResumeOptimizationHeader() {
  const navigate = useNavigate();

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
          <span className="text-gray-400">当前简历：</span>
          <div className="flex items-center text-gray-700 font-medium">
            <FileText className="w-3 h-3 mr-1 text-[#0052D9]" />
            {resumeVm.name}
          </div>
        </div>
        <div className="w-px h-3 bg-gray-200"></div>
        <div className="flex items-center gap-1.5">
          <span className="text-gray-400">优化状态：</span>
          <div className="flex items-center text-gray-500">
            <Activity className="w-3 h-3 mr-1 text-gray-400" />
            Mock 演示结果 / 兜底结果
          </div>
        </div>
      </div>
    </div>
  );
}