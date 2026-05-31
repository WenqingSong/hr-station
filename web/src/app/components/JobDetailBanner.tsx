import React from "react";
import { ArrowLeft } from "lucide-react";
import { useNavigate } from "react-router";
import { Badge } from "./ui/Badge";
import { mockSelectedJob } from "../mocks";
import { toJobDetailViewModel } from "../adapters";

const jobVm = toJobDetailViewModel(mockSelectedJob);

export function JobDetailBanner() {
  const navigate = useNavigate();
  
  return (
    <div className="w-full bg-[#0052D9] relative overflow-hidden text-white">
      <div className="absolute inset-0 bg-gradient-to-br from-[#0052D9] via-[#0040A8] to-[#002B73]"></div>
      <div className="absolute top-0 right-[10%] w-[600px] h-[600px] rounded-full bg-white opacity-[0.03] blur-3xl mix-blend-overlay pointer-events-none"></div>

      <div className="mx-auto max-w-[1200px] px-6 py-6 relative z-10 flex flex-col items-start">
        <button 
          onClick={() => navigate("/")}
          className="flex items-center gap-1.5 text-blue-100 hover:text-white transition-colors mb-4 text-sm font-medium"
        >
          <ArrowLeft className="w-4 h-4" />
          返回岗位列表
        </button>
        
        <h1 className="text-[28px] font-bold mb-3 tracking-tight">{jobVm.title}</h1>
        
        <div className="flex items-center gap-3 mb-4">
          <Badge className="bg-white/10 text-white border-transparent hover:bg-white/20 font-normal px-2.5 py-0.5 rounded-sm">{jobVm.typeLabel}</Badge>
          <Badge className="bg-white/10 text-white border-transparent hover:bg-white/20 font-normal px-2.5 py-0.5 rounded-sm">{jobVm.categoryLabel}</Badge>
          <Badge className="bg-white/10 text-white border-transparent hover:bg-white/20 font-normal px-2.5 py-0.5 rounded-sm">校园招聘</Badge>
        </div>

        <div className="flex flex-col gap-2.5 text-sm text-blue-50/90 max-w-3xl">
          <div className="flex items-start gap-3">
            <span className="opacity-60 shrink-0 w-[70px]">事业群：</span>
            <span>{jobVm.businessGroup}</span>
          </div>
          <div className="flex items-start gap-3">
            <span className="opacity-60 shrink-0 w-[70px]">工作地点：</span>
            <span>{jobVm.locationsText}</span>
          </div>
          <div className="flex items-start gap-3 mt-1.5">
            <span className="opacity-60 shrink-0 w-[70px]">岗位简介：</span>
            <p className="leading-relaxed">{jobVm.bannerSummary}</p>
          </div>
        </div>
      </div>
    </div>
  );
}
