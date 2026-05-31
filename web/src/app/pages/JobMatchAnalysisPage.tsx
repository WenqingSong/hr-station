import React from "react";
import { MatchAnalysisHeader } from "../components/MatchAnalysisHeader";
import { MatchAnalysisSidebarLeft } from "../components/MatchAnalysisSidebarLeft";
import { MatchAnalysisDetail } from "../components/MatchAnalysisDetail";
import { MatchAnalysisSidebarRight } from "../components/MatchAnalysisSidebarRight";

export function JobMatchAnalysisPage() {
  return (
    <div className="mx-auto max-w-[1440px] px-6 md:px-8 py-6 pb-12">
      <MatchAnalysisHeader />
      <div className="flex gap-6 mt-6 items-start">
        <MatchAnalysisSidebarLeft />
        <MatchAnalysisDetail />
        <MatchAnalysisSidebarRight />
      </div>
    </div>
  );
}
