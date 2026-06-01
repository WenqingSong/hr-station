import React from "react";
import { useParams } from "react-router";
import { JobDetailBanner } from "../components/JobDetailBanner";
import { JobDetailJD } from "../components/JobDetailJD";
import { JobDetailAIDecision } from "../components/JobDetailAIDecision";

export function JobDetailPage() {
  const { id: jobId } = useParams<{ id: string }>();
  return (
    <>
      <JobDetailBanner jobId={jobId} />
      <div className="mx-auto max-w-[1440px] px-6 md:px-8 py-6">
        <div className="flex gap-6 items-start">
          <JobDetailJD jobId={jobId} />
          <JobDetailAIDecision />
        </div>
      </div>
    </>
  );
}
