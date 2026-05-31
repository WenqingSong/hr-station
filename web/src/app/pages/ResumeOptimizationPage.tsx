import React from "react";
import { ResumeOptimizationHeader } from "../components/ResumeOptimizationHeader";
import { ResumeOptimizationSidebarLeft } from "../components/ResumeOptimizationSidebarLeft";
import { ResumeOptimizationDetail } from "../components/ResumeOptimizationDetail";
import { ResumeOptimizationSidebarRight } from "../components/ResumeOptimizationSidebarRight";

export function ResumeOptimizationPage() {
  return (
    <div className="min-h-screen bg-[#F3F4F6] font-sans text-gray-900 selection:bg-blue-100 selection:text-blue-900 pb-20">
      <div className="mx-auto max-w-[1440px] px-6 md:px-8 py-6">
        <ResumeOptimizationHeader />
        
        <div className="flex gap-6 mt-6 items-start justify-center">
          <ResumeOptimizationSidebarLeft />
          <ResumeOptimizationDetail />
          <ResumeOptimizationSidebarRight />
        </div>
      </div>
    </div>
  );
}