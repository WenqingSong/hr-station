import React, { useState, useEffect } from "react";
import { Link, useSearchParams, useNavigate, useLocation } from "react-router";
import { ResumeOptimizationHeader } from "../components/resume-optimization/ResumeOptimizationHeader";
import { ResumeOptimizationSidebarLeft } from "../components/resume-optimization/ResumeOptimizationSidebarLeft";
import { ResumeOptimizationDetail } from "../components/resume-optimization/ResumeOptimizationDetail";
import { ResumeOptimizationSidebarRight } from "../components/resume-optimization/ResumeOptimizationSidebarRight";
import { loadResumeOptimizationViewModel, loadResumeOptionViewModelsFromApi, loadResumeVersionDetail } from "../../api";
import { getJobDetail } from "../../api/jobs";
import type { JobApiItem } from "../../api/jobs";
import type { ResumeOptimizationPageViewModel, ResumeOptionViewModel } from "../../adapters";
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

export function ResumeOptimizationPage() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const location = useLocation();
  const jobId = searchParams.get("jobId");
  const resumeId = searchParams.get("resumeId");

  const [optimizationMode, setOptimizationMode] = useState<"standard" | "core">("standard");
  const [modeChanged, setModeChanged] = useState(false);

  const [job, setJob] = useState<JobApiItem | null>(null);
  const [loadingJob, setLoadingJob] = useState(!!jobId);

  const [resumeOptions, setResumeOptions] = useState<ResumeOptionViewModel[]>([]);
  const [selectedOption, setSelectedOption] = useState<ResumeOptionViewModel | null>(null);
  const [loadingOptions, setLoadingOptions] = useState(true);

  const [selectedResumeDetail, setSelectedResumeDetail] = useState<Resume | null>(null);
  const [loadingResumeDetail, setLoadingResumeDetail] = useState(false);
  const [resumeDetailError, setResumeDetailError] = useState<string | null>(null);

  const [result, setResult] = useState<ResumeOptimizationPageViewModel | null>(null);
  const [loadingOptimize, setLoadingOptimize] = useState(false);
  const [optimizeError, setOptimizeError] = useState<string | null>(null);

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
          setResumeDetailError("简历详情加载失败，请重新选择简历后再进行优化。");
        } else {
          setSelectedResumeDetail(detail);
        }
      })
      .catch((err) => {
        const authRedirect = getAuthRedirectIf401(err, location.pathname, location.search);
        if (authRedirect) { navigate(authRedirect, { replace: true }); return; }
        setResumeDetailError("简历详情加载失败，请重新选择简历后再进行优化。");
      })
      .finally(() => setLoadingResumeDetail(false));
  }, [selectedOption]);

  useEffect(() => {
    if (!job || !selectedResumeDetail || !selectedOption) return;

    setLoadingOptimize(true);
    setOptimizeError(null);

    // K8-5-48-D: pass resume_id + job_id, backend resolves raw_text
    const request: Parameters<typeof loadResumeOptimizationViewModel>[0] = {
      resume_id: selectedOption.id,
      job_id: job.id,
      target_section: "project",
    };
    if (optimizationMode === "core") {
      request.mode = "core";
    }
    loadResumeOptimizationViewModel(request)
      .then((vm) => {
        setModeChanged(false);
        setResult(vm);
      })
      .catch(() => setOptimizeError("简历优化失败，请稍后重试。"))
      .finally(() => setLoadingOptimize(false));
  }, [job, selectedResumeDetail, selectedOption, optimizationMode]);

  if (!jobId) {
    return (
      <div className="max-w-[1440px] mx-auto px-6 py-12 text-center">
        <p className="text-gray-500 text-sm mb-3">缺少目标岗位</p>
        <p className="text-gray-400 text-xs mb-4">请先选择一个岗位，再进行简历优化。</p>
        <Link to="/jobs" className="text-[#0052D9] text-sm underline">返回岗位列表</Link>
      </div>
    );
  }

  if (loadingJob || loadingOptions) {
    return <PageLoading message="正在加载岗位和简历信息…" />;
  }

  if (!job) {
    return (
      <div className="max-w-[1440px] mx-auto px-6 py-12 text-center">
        <p className="text-gray-500 text-sm mb-3">目标岗位加载失败</p>
        <p className="text-gray-400 text-xs mb-4">请返回岗位列表重新选择岗位。</p>
        <Link to="/jobs" className="text-[#0052D9] text-sm underline">返回岗位列表</Link>
      </div>
    );
  }

  if (resumeOptions.length === 0 || !selectedOption) {
    return (
      <div className="max-w-[1440px] mx-auto px-6 py-12 text-center">
        <p className="text-gray-500 text-sm mb-3">你还没有可用于优化的简历</p>
        <p className="text-gray-400 text-xs mb-4">请先创建一份简历，系统会基于你的简历内容给出优化建议。</p>
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
        <p className="text-gray-400 text-xs mb-4">请重新选择简历后再进行优化。</p>
        <Link to="/resumes" className="text-[#0052D9] text-sm underline">管理我的简历</Link>
      </div>
    );
  }

  if (loadingOptimize) {
    return (
      <AIThinkingLoading
        title="正在生成简历优化建议"
        description="Offerr鹅正在分析简历表达、岗位要求和可补强内容。"
        steps={["识别简历薄弱项", "匹配岗位筛选关注点", "生成可执行优化建议"]}
      />
    );
  }

  if (!result) {
    return (
      <div className="max-w-[1440px] mx-auto px-6 py-12 text-center">
        <p className="text-gray-500 text-sm mb-3">{optimizeError || "优化接口暂时不可用，请稍后重试。"}</p>
        <Link to={`/jobs/${jobId}`} className="text-[#0052D9] text-sm underline">返回岗位详情</Link>
      </div>
    );
  }

  const jobSource = "目标岗位";
  const resumeSource = "";

  const resumeForDisplay: Resume = selectedResumeDetail
    ? selectedResumeDetail
    : undefined as unknown as Resume;

  return (
    <ContentPageShell maxWidthClassName="max-w-[1440px]" contentClassName="pb-12">
      <ResumeOptimizationHeader job={job} resume={resumeForDisplay} jobSource={jobSource} resumeSource={resumeSource} />

      <div className="flex items-center gap-3 mt-5 mb-1">
        <span className="text-[13px] text-gray-500 font-medium">优化模式：</span>
        <div className="flex bg-gray-100 rounded-lg p-0.5 border border-gray-200">
          <button
            type="button"
            className={`px-4 py-1.5 rounded-md text-[13px] font-medium transition-all ${
              optimizationMode === "standard"
                ? "bg-white text-[#0052D9] shadow-sm border border-gray-200"
                : "text-gray-500 hover:text-gray-700"
            }`}
            onClick={() => { if (optimizationMode !== "standard") { setOptimizationMode("standard"); setModeChanged(true); } }}
          >
            标准优化
          </button>
          <button
            type="button"
            className={`px-4 py-1.5 rounded-md text-[13px] font-medium transition-all ${
              optimizationMode === "core"
                ? "bg-white text-[#0052D9] shadow-sm border border-gray-200"
                : "text-gray-500 hover:text-gray-700"
            }`}
            onClick={() => { if (optimizationMode !== "core") { setOptimizationMode("core"); setModeChanged(true); } }}
          >
            Core 优化
          </button>
        </div>
        <span className="text-[12px] text-gray-400">
          {optimizationMode === "core"
            ? "使用 Optimization Core Agent，输出最关键的 2-3 条优化建议"
            : "使用完整简历优化链路，含关键词对齐、项目优化等详细建议"}
        </span>
      </div>

      <div className="flex gap-6 mt-6 items-start">
        <ResumeOptimizationSidebarLeft job={job} jobSource={jobSource} />
        <ResumeOptimizationDetail currentJobId={jobId} currentResumeId={String(selectedOption.id)} data={result} />
        <ResumeOptimizationSidebarRight resume={resumeForDisplay} resumeSource={resumeSource} />
      </div>
    </ContentPageShell>
  );
}
