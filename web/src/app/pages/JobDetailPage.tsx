import React from "react";
import { JobDetailBanner } from "../components/JobDetailBanner";
import { JobDetailJD } from "../components/JobDetailJD";
import { JobDetailAIDecision } from "../components/JobDetailAIDecision";

export function JobDetailPage() {
  return (
    <>
      <JobDetailBanner />
      <div className="mx-auto max-w-[1440px] px-6 md:px-8 py-6">
        <div className="flex gap-6 items-start">
          <JobDetailJD />
          <JobDetailAIDecision />
        </div>
      </div>
    </>
  );
}
