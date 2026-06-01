import React, { useState, useEffect } from "react";
import { useSearchParams } from "react-router";
import { ResumeOptimizationHeader } from "../components/ResumeOptimizationHeader";
import { ResumeOptimizationSidebarLeft } from "../components/ResumeOptimizationSidebarLeft";
import { ResumeOptimizationDetail } from "../components/ResumeOptimizationDetail";
import { ResumeOptimizationSidebarRight } from "../components/ResumeOptimizationSidebarRight";
import { loadResumeOptimizationViewModel } from "../../api";
import type { ResumeOptimizationPageViewModel } from "../../adapters";
import {
  mockSelectedJob, mockJobDetails, mockCurrentResume, mockResumeVersions,
} from "../../mocks";

export function ResumeOptimizationPage() {
  const [searchParams] = useSearchParams();
  const jobId = searchParams.get("jobId");
  const resumeId = searchParams.get("resumeId");
  const [result, setResult] = useState<ResumeOptimizationPageViewModel | null>(null);
  const [loading, setLoading] = useState(true);

  const gotJob = jobId ? (mockJobDetails[jobId] ?? mockSelectedJob) : mockSelectedJob;
  const gotResume = resumeId
    ? (mockResumeVersions.find((r) => r.id === resumeId) ?? mockCurrentResume)
    : mockCurrentResume;
  const jobSource = "基于匹配分析结果";
  const resumeSource = "当前简历";

  useEffect(() => {
    const resumeText = [gotResume.name, gotResume.direction_label, gotResume.skills?.join(", "), gotResume.self_evaluation].filter(Boolean).join("\n");
    const targetJob = { title: gotJob.title, jd_text: [gotJob.description, gotJob.requirements].filter(Boolean).join("\n") };
    loadResumeOptimizationViewModel({ resume_text: resumeText, target_job: targetJob, target_section: "project" })
      .then(setResult)
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [jobId, resumeId]);

  if (loading) return <div className="max-w-[1440px] mx-auto px-6 py-12 text-center text-gray-400 text-sm">正在生成简历优化建议…</div>;
  if (!result) return <div className="max-w-[1440px] mx-auto px-6 py-12 text-center"><p className="text-gray-500 text-sm">优化接口暂时不可用，已展示演示数据。</p></div>;

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