import React, { useState, useEffect } from "react";
import { useSearchParams } from "react-router";
import { MatchAnalysisHeader } from "../components/MatchAnalysisHeader";
import { MatchAnalysisSidebarLeft } from "../components/MatchAnalysisSidebarLeft";
import { MatchAnalysisDetail } from "../components/MatchAnalysisDetail";
import { MatchAnalysisSidebarRight } from "../components/MatchAnalysisSidebarRight";
import { loadMatchAnalysisViewModel } from "../../api";
import type { MatchAnalysisPageViewModel } from "../../adapters";
import {
  mockSelectedJob, mockJobDetails, mockCurrentResume, mockResumeVersions,
} from "../../mocks";

/** 将前端 resumeId（如 resume-v2）映射为后端 resume_version_no */
function resolveResumeVersionNo(resumeId: string | null): number {
  if (!resumeId) return 2;
  const matched = resumeId.match(/\d+/);
  return matched ? Number(matched[0]) : 2;
}

export function JobMatchAnalysisPage() {
  const [searchParams] = useSearchParams();
  const jobId = searchParams.get("jobId") || "1";
  const resumeId = searchParams.get("resumeId");
  const [result, setResult] = useState<MatchAnalysisPageViewModel | null>(null);
  const [loading, setLoading] = useState(true);

  // Fallback: 上下文展示仍用 mock 数据
  const gotJob = jobId ? (mockJobDetails[jobId] ?? mockSelectedJob) : mockSelectedJob;
  const gotResume = resumeId
    ? (mockResumeVersions.find((r) => r.id === resumeId) ?? mockCurrentResume)
    : mockCurrentResume;
  const jobSource = "目标岗位";
  const resumeSource = "当前简历";

  useEffect(() => {
    const versionNo = resolveResumeVersionNo(resumeId);
    loadMatchAnalysisViewModel(jobId, { resume_version_no: versionNo })
      .then(setResult)
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [jobId, resumeId]);

  if (loading) {
    return <div className="max-w-[1440px] mx-auto px-6 py-12 text-center text-gray-400 text-sm">正在生成匹配分析…</div>;
  }
  if (!result) {
    return <div className="max-w-[1440px] mx-auto px-6 py-12 text-center"><p className="text-gray-500 text-sm mb-2">匹配分析接口暂时不可用，已展示演示数据。</p></div>;
  }

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
