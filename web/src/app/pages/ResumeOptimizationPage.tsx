import React from "react";
import { useSearchParams } from "react-router";
import { ResumeOptimizationHeader } from "../components/ResumeOptimizationHeader";
import { ResumeOptimizationSidebarLeft } from "../components/ResumeOptimizationSidebarLeft";
import { ResumeOptimizationDetail } from "../components/ResumeOptimizationDetail";
import { ResumeOptimizationSidebarRight } from "../components/ResumeOptimizationSidebarRight";
import {
  mockSelectedJob, mockJobDetails, mockCurrentResume, mockResumeVersions,
} from "../mocks";

export function ResumeOptimizationPage() {
  const [searchParams] = useSearchParams();
  const jobId = searchParams.get("jobId");
  const resumeId = searchParams.get("resumeId");

  const gotJob = jobId ? (mockJobDetails[jobId] ?? mockSelectedJob) : mockSelectedJob;
  const gotResume = resumeId
    ? (mockResumeVersions.find((r) => r.id === resumeId) ?? mockCurrentResume)
    : mockCurrentResume;

  const jobSource = jobId && mockJobDetails[jobId] ? "来自岗位匹配分析" : "默认岗位";
  const resumeSource = resumeId && mockResumeVersions.some((r) => r.id === resumeId) ? "来自岗位匹配分析" : "默认简历";

  return (
    <div className="mx-auto max-w-[1440px] px-6 md:px-8 py-6 pb-12">
      <ResumeOptimizationHeader job={gotJob} resume={gotResume} jobSource={jobSource} resumeSource={resumeSource} />
      <div className="flex gap-6 mt-6 items-start">
        <ResumeOptimizationSidebarLeft job={gotJob} jobSource={jobSource} />
        <ResumeOptimizationDetail currentJobId={gotJob.id} currentResumeId={gotResume.id} />
        <ResumeOptimizationSidebarRight resume={gotResume} resumeSource={resumeSource} />
      </div>
    </div>
  );
}