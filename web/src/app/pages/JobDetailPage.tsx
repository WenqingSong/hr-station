import React from "react";
import { JobDetailBanner } from "../components/JobDetailBanner";
import { JobDetailJD } from "../components/JobDetailJD";
import { JobDetailAIDecision } from "../components/JobDetailAIDecision";

export function JobDetailPage() {
  return (
    <div className="min-h-screen bg-[#F3F4F6] font-sans text-gray-900 selection:bg-blue-100 selection:text-blue-900">
      <JobDetailBanner />
      
      <div className="mx-auto max-w-[1200px] px-6 py-8">
        <div className="flex gap-6 items-start justify-center">
          <JobDetailJD />
          <JobDetailAIDecision />
        </div>
      </div>
    </div>
  );
}
