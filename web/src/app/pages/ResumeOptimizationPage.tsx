import React from "react";
import { ResumeOptimizationHeader } from "../components/ResumeOptimizationHeader";
import { ResumeOptimizationSidebarLeft } from "../components/ResumeOptimizationSidebarLeft";
import { ResumeOptimizationDetail } from "../components/ResumeOptimizationDetail";
import { ResumeOptimizationSidebarRight } from "../components/ResumeOptimizationSidebarRight";

export function ResumeOptimizationPage() {
  return (
    <div className="mx-auto max-w-[1440px] px-6 md:px-8 py-6 pb-12">
      <ResumeOptimizationHeader />
      <div className="flex gap-6 mt-6 items-start">
        <ResumeOptimizationSidebarLeft />
        <ResumeOptimizationDetail />
        <ResumeOptimizationSidebarRight />
      </div>
    </div>
  );
}