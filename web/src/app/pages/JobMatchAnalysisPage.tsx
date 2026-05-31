import React from "react";
import { useSearchParams } from "react-router";
import { MatchAnalysisHeader } from "../components/MatchAnalysisHeader";
import { MatchAnalysisSidebarLeft } from "../components/MatchAnalysisSidebarLeft";
import { MatchAnalysisDetail } from "../components/MatchAnalysisDetail";
import { MatchAnalysisSidebarRight } from "../components/MatchAnalysisSidebarRight";
import {
  mockSelectedJob, mockJobDetails, mockCurrentResume, mockResumeVersions,
} from "../mocks";

export function JobMatchAnalysisPage() {
  const [searchParams] = useSearchParams();
  const jobId = searchParams.get("jobId");
  const resumeId = searchParams.get("resumeId");

  const gotJob = jobId ? (mockJobDetails[jobId] ?? mockSelectedJob) : mockSelectedJob;
  const gotResume = resumeId
    ? (mockResumeVersions.find((r) => r.id === resumeId) ?? mockCurrentResume)
    : mockCurrentResume;

  const jobSource = jobId && mockJobDetails[jobId] ? "已选择岗位" : "默认岗位";
  const resumeSource = resumeId && mockResumeVersions.some((r) => r.id === resumeId) ? "已选择简历" : "默认简历";

  return (
    <div className="mx-auto max-w-[1440px] px-6 md:px-8 py-6 pb-12">
      <MatchAnalysisHeader job={gotJob} resume={gotResume} jobSource={jobSource} resumeSource={resumeSource} />
      <div className="flex gap-6 mt-6 items-start">
        <MatchAnalysisSidebarLeft job={gotJob} jobSource={jobSource} />
        <MatchAnalysisDetail currentJobId={gotJob.id} currentResumeId={gotResume.id} />
        <MatchAnalysisSidebarRight resume={gotResume} resumeSource={resumeSource} />
      </div>
    </div>
  );
}
