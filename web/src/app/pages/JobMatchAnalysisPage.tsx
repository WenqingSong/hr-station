import React, { useState, useEffect, useCallback } from "react";
import { useSearchParams } from "react-router";
import { MatchAnalysisHeader } from "../components/MatchAnalysisHeader";
import { MatchAnalysisSidebarLeft } from "../components/MatchAnalysisSidebarLeft";
import { MatchAnalysisDetail } from "../components/MatchAnalysisDetail";
import { MatchAnalysisSidebarRight } from "../components/MatchAnalysisSidebarRight";
import { loadMatchAnalysisViewModel, loadResumeOptionViewModelsFromApi } from "../../api";
import { getJobDetail } from "../../api/jobs";
import { getCurrentResume } from "../../api/resumes";
import type { MatchAnalysisPageViewModel, ResumeOptionViewModel } from "../../adapters";
import type { JobDetail, Resume } from "../../types";

/** 从 URL 查询参数中解析 resume_version_no（兼容 legacy resumeId=resume-vN） */
function parseVersionNoFromQuery(resumeId: string | null, resumeVersionNo: string | null): number | null {
  if (resumeVersionNo && /^\d+$/.test(resumeVersionNo)) return Number(resumeVersionNo);
  if (resumeId) {
    if (/^\d+$/.test(resumeId)) return Number(resumeId);
    const match = resumeId.match(/^resume-v(\d+)$/);
    if (match) return Number(match[1]);
  }
  return null;
}

export function JobMatchAnalysisPage() {
  const [searchParams] = useSearchParams();
  const jobId = searchParams.get("jobId");
  const resumeId = searchParams.get("resumeId");
  const resumeVersionNo = searchParams.get("resumeVersionNo");

  // 目标岗位
  const [job, setJob] = useState<JobDetail | null>(null);
  const [loadingJob, setLoadingJob] = useState(!!jobId);

  // 简历选项 & 选中状态
  const [resumeOptions, setResumeOptions] = useState<ResumeOptionViewModel[]>([]);
  const [selectedOption, setSelectedOption] = useState<ResumeOptionViewModel | null>(null);
  const [currentResume, setCurrentResume] = useState<Resume | null>(null);
  const [loadingResume, setLoadingResume] = useState(true);

  // 匹配结果
  const [result, setResult] = useState<MatchAnalysisPageViewModel | null>(null);
  const [loadingMatch, setLoadingMatch] = useState(false);
  const [matchError, setMatchError] = useState<string | null>(null);

  // 加载目标岗位
  useEffect(() => {
    if (!jobId) { setLoadingJob(false); return; }
    getJobDetail(jobId)
      .then(setJob)
      .catch(() => {})
      .finally(() => setLoadingJob(false));
  }, [jobId]);

  // 加载简历选项 + current resume
  useEffect(() => {
    Promise.all([
      loadResumeOptionViewModelsFromApi(),
      getCurrentResume().catch(() => null as Resume | null),
    ]).then(([opts, resume]) => {
      setResumeOptions(opts);
      setCurrentResume(resume);
      if (opts.length === 0) { setLoadingResume(false); return; }
      const targetVersionNo = parseVersionNoFromQuery(resumeId, resumeVersionNo);
      const matched = targetVersionNo
        ? opts.find((o) => Number(o.id) === targetVersionNo)
        : undefined;
      setSelectedOption(matched ?? opts[0]);
      setLoadingResume(false);
    }).catch(() => setLoadingResume(false));
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  // 触发匹配 Agent
  useEffect(() => {
    if (!jobId || !selectedOption) return;
    const versionNo = Number(selectedOption.id);
    if (!versionNo || isNaN(versionNo)) return;

    setLoadingMatch(true);
    setMatchError(null);
    loadMatchAnalysisViewModel(jobId, { resume_version_no: versionNo })
      .then(setResult)
      .catch(() => setMatchError("匹配分析结果加载失败，请稍后重试。"))
      .finally(() => setLoadingMatch(false));
  }, [jobId, selectedOption]);

  // ── 空态：缺 jobId ──
  if (!jobId) {
    return (
      <div className="max-w-[1440px] mx-auto px-6 py-12 text-center">
        <p className="text-gray-500 text-sm mb-3">缺少目标岗位</p>
        <p className="text-gray-400 text-xs mb-4">请先选择一个岗位，再进行匹配分析。</p>
        <a href="/jobs" className="text-[#0052D9] text-sm underline">返回岗位列表</a>
      </div>
    );
  }

  // ── 空态：加载中 ──
  if (loadingJob || loadingResume) {
    return (
      <div className="max-w-[1440px] mx-auto px-6 py-12 text-center text-gray-400 text-sm">
        正在加载岗位和简历信息…
      </div>
    );
  }

  // ── 空态：无简历版本 ──
  if (resumeOptions.length === 0 || !selectedOption) {
    return (
      <div className="max-w-[1440px] mx-auto px-6 py-12 text-center">
        <p className="text-gray-500 text-sm mb-3">你还没有可用于匹配的简历</p>
        <p className="text-gray-400 text-xs mb-4">请先创建一份简历，系统会基于你的简历版本分析岗位匹配度。</p>
        <a href="/resumes/new" className="text-[#0052D9] text-sm underline">去添加简历</a>
      </div>
    );
  }

  // ── 加载匹配结果 ──
  if (loadingMatch) {
    return (
      <div className="max-w-[1440px] mx-auto px-6 py-12 text-center text-gray-400 text-sm">
        正在生成匹配分析…
      </div>
    );
  }

  // ── 匹配失败 ──
  if (!result) {
    return (
      <div className="max-w-[1440px] mx-auto px-6 py-12 text-center">
        <p className="text-gray-500 text-sm mb-2">{matchError || "匹配分析接口暂时不可用，请稍后重试。"}</p>
      </div>
    );
  }

  const jobSource = "目标岗位";
  const resumeSource = `${selectedOption.statusLabel} · ${selectedOption.directionLabel || "求职方向未填写"}`;

  return (
    <div className="mx-auto max-w-[1440px] px-6 md:px-8 py-6 pb-12">
      <MatchAnalysisHeader job={job ?? undefined} resume={currentResume ?? undefined} jobSource={jobSource} resumeSource={resumeSource} />
      <div className="flex gap-6 mt-6 items-start">
        <MatchAnalysisSidebarLeft job={job ?? undefined} jobSource={jobSource} />
        <MatchAnalysisDetail currentJobId={jobId} currentResumeId={String(selectedOption.id)} />
        <MatchAnalysisSidebarRight resume={currentResume ?? undefined} resumeSource={resumeSource} />
      </div>
    </div>
  );
}
