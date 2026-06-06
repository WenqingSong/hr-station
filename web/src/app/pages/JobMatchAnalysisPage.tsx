import React, { useState, useEffect } from "react";
import { Link, useSearchParams, useNavigate, useLocation } from "react-router";
import { MatchAnalysisHeader } from "../components/match/MatchAnalysisHeader";
import { MatchAnalysisSidebarLeft } from "../components/match/MatchAnalysisSidebarLeft";
import { MatchAnalysisDetail } from "../components/match/MatchAnalysisDetail";
import { MatchAnalysisSidebarRight } from "../components/match/MatchAnalysisSidebarRight";
import { loadMatchAnalysisViewModel, loadResumeOptionViewModelsFromApi, loadResumeVersionDetail } from "../../api";
import { getJobDetail } from "../../api/jobs";
import type { JobApiItem } from "../../api/jobs";
import type { MatchAnalysisPageViewModel, ResumeOptionViewModel } from "../../adapters";
import type { Resume } from "../../types";
import { ContentPageShell } from "../components/layout/ContentPageShell";
import { getAuthRedirectIf401 } from "../utils/authRedirect";
import { PageLoading } from "../components/PageState";
import { AIThinkingLoading } from "../components/common/AIThinkingLoading";

/** 从 URL 查询参数中解析 resumeId（K8-5-48-C） */
function parseResumeIdFromQuery(resumeId: string | null): number | null {
  if (resumeId && /^\d+$/.test(resumeId)) return Number(resumeId);
  return null;
}

export function JobMatchAnalysisPage() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const location = useLocation();
  const jobId = searchParams.get("jobId");
  const resumeId = searchParams.get("resumeId");

  const [job, setJob] = useState<JobApiItem | null>(null);
  const [loadingJob, setLoadingJob] = useState(!!jobId);

  const [resumeOptions, setResumeOptions] = useState<ResumeOptionViewModel[]>([]);
  const [selectedOption, setSelectedOption] = useState<ResumeOptionViewModel | null>(null);
  const [loadingOptions, setLoadingOptions] = useState(true);

  const [selectedResumeDetail, setSelectedResumeDetail] = useState<Resume | null>(null);
  const [loadingResumeDetail, setLoadingResumeDetail] = useState(false);
  const [resumeDetailError, setResumeDetailError] = useState<string | null>(null);

  const [result, setResult] = useState<MatchAnalysisPageViewModel | null>(null);
  const [loadingMatch, setLoadingMatch] = useState(false);
  const [matchError, setMatchError] = useState<string | null>(null);

  useEffect(() => {
    if (!jobId) { setLoadingJob(false); return; }
    getJobDetail(jobId)
      .then(setJob)
      .catch(() => {})
      .finally(() => setLoadingJob(false));
  }, [jobId]);

  useEffect(() => {
    loadResumeOptionViewModelsFromApi()
      .then((opts) => {
        setResumeOptions(opts);
        if (opts.length === 0) return;
        const targetResumeId = parseResumeIdFromQuery(resumeId);
        const matched = targetResumeId
          ? opts.find((o) => o.id === targetResumeId)
          : undefined;
        setSelectedOption(matched ?? opts[0]);
      })
      .catch((err) => {
        const authRedirect = getAuthRedirectIf401(err, location.pathname, location.search);
        if (authRedirect) { navigate(authRedirect, { replace: true }); return; }
      })
      .finally(() => setLoadingOptions(false));
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    if (!selectedOption) return;
    const id = selectedOption.id;
    if (!id || isNaN(id)) return;

    setLoadingResumeDetail(true);
    setResumeDetailError(null);
    setSelectedResumeDetail(null);
    loadResumeVersionDetail(id)
      .then((detail) => {
        if (!detail) {
          setResumeDetailError("简历版本详情加载失败，请重新选择简历后再进行匹配分析。");
        } else {
          setSelectedResumeDetail(detail);
        }
      })
      .catch((err) => {
        const authRedirect = getAuthRedirectIf401(err, location.pathname, location.search);
        if (authRedirect) { navigate(authRedirect, { replace: true }); return; }
        setResumeDetailError("简历版本详情加载失败，请重新选择简历后再进行匹配分析。");
      })
      .finally(() => setLoadingResumeDetail(false));
  }, [selectedOption]);

  useEffect(() => {
    if (!jobId || !selectedOption || !selectedResumeDetail) return;
    const id = selectedOption.id;
    if (!id || isNaN(id)) return;

    setLoadingMatch(true);
    setMatchError(null);
    loadMatchAnalysisViewModel(jobId, { /* resume_id when backend supports */ })
      .then(setResult)
      .catch(() => setMatchError("匹配分析结果加载失败，请稍后重试。"))
      .finally(() => setLoadingMatch(false));
  }, [jobId, selectedOption, selectedResumeDetail]);

  if (!jobId) {
    return (
      <div className="max-w-[1440px] mx-auto px-6 py-12 text-center">
        <p className="text-gray-500 text-sm mb-3">缺少目标岗位</p>
        <p className="text-gray-400 text-xs mb-4">请先选择一个岗位，再进行匹配分析。</p>
        <Link to="/jobs" className="text-[#0052D9] text-sm underline">返回岗位列表</Link>
      </div>
    );
  }

  if (loadingJob || loadingOptions) {
    return <PageLoading message="正在加载岗位和简历信息…" />;
  }

  if (resumeOptions.length === 0 || !selectedOption) {
    return (
      <div className="max-w-[1440px] mx-auto px-6 py-12 text-center">
        <p className="text-gray-500 text-sm mb-3">你还没有可用于匹配的简历</p>
        <p className="text-gray-400 text-xs mb-4">请先创建一份简历，系统会基于你的简历分析岗位匹配度。</p>
        <Link to="/resumes/new" className="text-[#0052D9] text-sm underline">去添加简历</Link>
      </div>
    );
  }

  if (loadingResumeDetail) {
    return <PageLoading message="正在加载简历详情…" />;
  }

  if (resumeDetailError) {
    return (
      <div className="max-w-[1440px] mx-auto px-6 py-12 text-center">
        <p className="text-gray-500 text-sm mb-3">{resumeDetailError}</p>
        <p className="text-gray-400 text-xs mb-4">请重新选择简历后再进行匹配分析。</p>
        <Link to="/resumes" className="text-[#0052D9] text-sm underline">管理我的简历</Link>
      </div>
    );
  }

  if (loadingMatch) {
    return (
      <AIThinkingLoading
        title="正在分析岗位匹配度"
        description="Offerr鹅正在结合简历内容和岗位要求，生成匹配分、优势与风险点。"
        steps={["读取简历关键信息", "对齐岗位 JD 要求", "生成可解释分析结果"]}
      />
    );
  }

  if (!result) {
    return (
      <div className="max-w-[1440px] mx-auto px-6 py-12 text-center">
        <p className="text-gray-500 text-sm mb-3">{matchError || "匹配分析接口暂时不可用，请稍后重试。"}</p>
        <Link to={`/jobs/${jobId}`} className="text-[#0052D9] text-sm underline">返回岗位详情</Link>
      </div>
    );
  }

  const jobSource = "目标岗位";
  const resumeSource = "";

  return (
    <ContentPageShell maxWidthClassName="max-w-[1440px]" contentClassName="pb-12">
      <MatchAnalysisHeader job={job ?? undefined} resume={selectedResumeDetail ?? undefined} jobSource={jobSource} resumeSource={resumeSource} />
      <div className="flex gap-6 mt-6 items-start">
        <MatchAnalysisSidebarLeft job={job ?? undefined} jobSource={jobSource} />
        <MatchAnalysisDetail currentJobId={jobId} currentResumeId={String(selectedOption.id)} data={result} />
        <MatchAnalysisSidebarRight resume={selectedResumeDetail ?? undefined} resumeSource={resumeSource} />
      </div>
    </ContentPageShell>
  );
}
