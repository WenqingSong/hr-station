import React from "react";
import { MatchAnalysisHeader } from "../components/MatchAnalysisHeader";
import { MatchAnalysisSidebarLeft } from "../components/MatchAnalysisSidebarLeft";
import { MatchAnalysisDetail } from "../components/MatchAnalysisDetail";
import { MatchAnalysisSidebarRight } from "../components/MatchAnalysisSidebarRight";

export function JobMatchAnalysisPage() {
  return (
    <div className="min-h-screen bg-[#F3F4F6] font-sans text-gray-900 selection:bg-blue-100 selection:text-blue-900 pb-20">
      <div className="mx-auto max-w-[1440px] px-6 md:px-8 py-6">
        <MatchAnalysisHeader />
        
        <div className="flex gap-6 mt-6 items-start justify-center">
          <MatchAnalysisSidebarLeft />
          <MatchAnalysisDetail />
          <MatchAnalysisSidebarRight />
        </div>
      </div>
    </div>
  );
}
