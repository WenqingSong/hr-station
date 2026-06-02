import React, { useState, useEffect } from "react";
import { useSearchParams } from "react-router";
import { MatchAnalysisHeader } from "../components/MatchAnalysisHeader";
import { MatchAnalysisSidebarLeft } from "../components/MatchAnalysisSidebarLeft";
import { MatchAnalysisDetail } from "../components/MatchAnalysisDetail";
import { MatchAnalysisSidebarRight } from "../components/MatchAnalysisSidebarRight";
import { loadMatchAnalysisViewModel, loadResumeOptionViewModelsFromApi, loadResumeVersionDetail } from "../../api";
import { getJobDetail } from "../../api/jobs";
import type { JobApiItem } from "../../api/jobs";
import type { MatchAnalysisPageViewModel, ResumeOptionViewModel } from "../../adapters";
import type { Resume } from "../../types";

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
  const [job, setJob] = useState<JobApiItem | null>(null);
  const [loadingJob, setLoadingJob] = useState(!!jobId);

  // 简历选项 & 选中状态
  const [resumeOptions, setResumeOptions] = useState<ResumeOptionViewModel[]>([]);
  const [selectedOption, setSelectedOption] = useState<ResumeOptionViewModel | null>(null);
  const [loadingOptions, setLoadingOptions] = useState(true);

  // 选中版本的简历详情（10-3-R：使用版本详情替代 getCurrentResume）
  const [selectedResumeDetail, setSelectedResumeDetail] = useState<Resume | null>(null);
  const [loadingResumeDetail, setLoadingResumeDetail] = useState(false);
  const [resumeDetailError, setResumeDetailError] = useState<string | null>(null);

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

  // 加载简历选项列表
  useEffect(() => {
    loadResumeOptionViewModelsFromApi()
      .then((opts) => {
        setResumeOptions(opts);
        if (opts.length === 0) return;
        const targetVersionNo = parseVersionNoFromQuery(resumeId, resumeVersionNo);
        const matched = targetVersionNo
          ? opts.find((o) => Number(o.id) === targetVersionNo)
          : undefined;
        setSelectedOption(matched ?? opts[0]);
      })
      .catch(() => {})
      .finally(() => setLoadingOptions(false));
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  // 根据选中版本号加载简历版本详情（10-3-R：selected resume version detail 对齐）
  useEffect(() => {
    if (!selectedOption) return;
    const versionNo = Number(selectedOption.id);
    if (!versionNo || isNaN(versionNo)) return;

    setLoadingResumeDetail(true);
    setResumeDetailError(null);
    setSelectedResumeDetail(null);
    loadResumeVersionDetail(versionNo)
      .then((detail) => {
        if (!detail) {
          setResumeDetailError("简历版本详情加载失败，请重新选择简历版本后再进行匹配分析。");
        } else {
          setSelectedResumeDetail(detail);
        }
      })
      .catch(() =>
        setResumeDetailError("简历版本详情加载失败，请重新选择简历版本后再进行匹配分析。")
      )
      .finally(() => setLoadingResumeDetail(false));
  }, [selectedOption]);

  // 触发匹配 Agent（版本详情加载成功后才调用）
  useEffect(() => {
    if (!jobId || !selectedOption || !selectedResumeDetail) return;
    const versionNo = Number(selectedOption.id);
    if (!versionNo || isNaN(versionNo)) return;

    setLoadingMatch(true);
    setMatchError(null);
    loadMatchAnalysisViewModel(jobId, { resume_version_no: versionNo })
      .then(setResult)
      .catch(() => setMatchError("匹配分析结果加载失败，请稍后重试。"))
      .finally(() => setLoadingMatch(false));
  }, [jobId, selectedOption, selectedResumeDetail]);

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
  if (loadingJob || loadingOptions) {
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

  // ── 空态：版本详情加载中 ──
  if (loadingResumeDetail) {
    return (
      <div className="max-w-[1440px] mx-auto px-6 py-12 text-center text-gray-400 text-sm">
        正在加载简历版本详情…
      </div>
    );
  }

  // ── 空态：版本详情加载失败 ──
  if (resumeDetailError) {
    return (
      <div className="max-w-[1440px] mx-auto px-6 py-12 text-center">
        <p className="text-gray-500 text-sm mb-3">{resumeDetailError}</p>
        <p className="text-gray-400 text-xs mb-4">请重新选择简历版本后再进行匹配分析。</p>
        <a href="/resumes" className="text-[#0052D9] text-sm underline">管理我的简历</a>
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
      <MatchAnalysisHeader job={job ?? undefined} resume={selectedResumeDetail ?? undefined} jobSource={jobSource} resumeSource={resumeSource} />
      <div className="flex gap-6 mt-6 items-start">
        <MatchAnalysisSidebarLeft job={job ?? undefined} jobSource={jobSource} />
        <MatchAnalysisDetail currentJobId={jobId} currentResumeId={String(selectedOption.id)} data={result} />
        <MatchAnalysisSidebarRight resume={selectedResumeDetail ?? undefined} resumeSource={resumeSource} />
      </div>
    </div>
  );
}
