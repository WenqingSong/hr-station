import React, { useState, useEffect } from "react";
import { Link, useSearchParams, useNavigate, useLocation } from "react-router";
import { loadLatestOptimizationV2ViewModel, loadGenerateOptimizationV2ViewModel, loadResumeOptionViewModelsFromApi, loadResumeVersionDetail } from "../../api";
import { getJobDetail } from "../../api/jobs";
import { getLatestMatchAnalysis } from "../../api/agents";
import type { MatchLatestResponse } from "../../api/agents";
import type { JobApiItem } from "../../api/jobs";
import type { ResumeOptionViewModel } from "../../adapters";
import type { OptimizationV2ViewModel } from "../../api/viewModels";
import type { Resume } from "../../types";
import type { OptimizationMode, GenerationTier } from "../../types";
import type { OptimizationFixItem, OptimizationRewriteExample, OptimizationLongTermItem } from "../../types";
import { ContentPageShell } from "../components/layout/ContentPageShell";
import { getAuthRedirectIf401 } from "../utils/authRedirect";
import { humanizeProfile } from "../utils/profileLabel";
import { cleanUserText } from "../utils/cleanUserText";
import { PageLoading } from "../components/PageState";
import { AgentStepLoadingOverlay } from "../components/common/AgentStepLoadingOverlay";
import { BulletList } from "../components/common/BulletList";
import { OptimizationEntryModal } from "../components/OptimizationEntryModal";
import { Sparkles, RefreshCw, Loader2, Target, Info, FileEdit, AlertTriangle, ArrowUp, FileText, Briefcase, Search, ExternalLink, ChevronDown, ChevronUp, Copy, CheckCheck, GraduationCap } from "lucide-react";

// ─── Text Safety ───────────────────────────────────────────────────────

/** K13-Stability-A: Sanitize display text. Safe against non-string types. */
function cleanText(value?: string | null | unknown): string {
  if (!value) return "";
  if (typeof value !== "string") return "";
  const sanitized = value
    .replace(/\uFFFD/g, "")
    .replace(/[\u0000-\u0008\u000B\u000C\u000E-\u001F]/g, "")
    .replace(/\s+/g, " ");
  return cleanUserText(sanitized).trim();
}

// ─── Phase ─────────────────────────────────────────────────────────────

type OptPhase =
  | "idle"              // waiting for job selection
  | "loading_latest"
  | "fast_loading"
  | "fast_done"
  | "fast_error"
  | "deep_loading"
  | "deep_done"
  | "deep_error";

// ─── Header Label ──────────────────────────────────────────────────────

function getHeaderLabel(phase: OptPhase, tier: string, genBy: string): string {
  if (phase === "idle") return "等待选择岗位";
  if (phase === "loading_latest") return "正在读取优化结果";
  if (phase === "fast_loading") return "正在生成快速优化建议";
  if (phase === "deep_loading") return "智能体正在深度优化";
  if (phase === "fast_done") return "快速优化 · 初步预览";
  if (phase === "deep_done") return "深度优化已完成";
  if (phase === "deep_error") return "深度优化生成失败 · 已保留当前结果";
  if (phase === "fast_error") return "快速优化生成失败";
  return "";
}

// ─── Mode Label ────────────────────────────────────────────────────────

function getModeLabel(mode: string, targetDir?: string, jobTitle?: string): string {
  if (mode === "target_job" && jobTitle) return `针对岗位 · ${jobTitle}`;
  if (mode === "target_direction" && targetDir) return `针对方向 · ${targetDir}`;
  return "通用优化";
}

// ─── Level Color ───────────────────────────────────────────────────────

function getLevelColor(level: string): string {
  if (level.includes("优秀") || level.includes("直接投递")) return "text-emerald-600 bg-emerald-50 border-emerald-200";
  if (level.includes("建议补强") || level.includes("补强后投递")) return "text-amber-600 bg-amber-50 border-amber-200";
  if (level.includes("需重点") || level.includes("不建议")) return "text-red-600 bg-red-50 border-red-200";
  return "text-blue-600 bg-blue-50 border-blue-200";
}

// ─── Component ─────────────────────────────────────────────────────────

export function ResumeOptimizationPage() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const location = useLocation();

  // URL params
  const resumeIdStr = searchParams.get("resumeId");
  const jobIdStr = searchParams.get("jobId");
  const targetDirection = searchParams.get("targetDirection");
  const explicitMode = searchParams.get("mode") as OptimizationMode | null;

  // Derived
  const parsedResumeId = parseResumeIdFromQuery(resumeIdStr);
  const parsedJobId = jobIdStr ? parseJobIdFromQuery(jobIdStr) : undefined;

  // K9-3-H-B: Don't silently fallback to general. If no jobId and not explicitly general/target_direction, show prompt.
  const needsJobSelection = parsedResumeId != null && parsedJobId == null && !targetDirection && explicitMode !== "general";

  // Actual mode for API calls
  const derivedMode: OptimizationMode = parsedJobId ? "target_job" : targetDirection ? "target_direction" : "general";

  // Job
  const [job, setJob] = useState<JobApiItem | null>(null);
  const [loadingJob, setLoadingJob] = useState(!!parsedJobId);

  // Resume options
  const [resumeOptions, setResumeOptions] = useState<ResumeOptionViewModel[]>([]);
  const [selectedOption, setSelectedOption] = useState<ResumeOptionViewModel | null>(null);
  const [loadingOptions, setLoadingOptions] = useState(true);

  // Resume detail
  const [selectedResumeDetail, setSelectedResumeDetail] = useState<Resume | null>(null);
  const [loadingResumeDetail, setLoadingResumeDetail] = useState(false);
  const [resumeDetailError, setResumeDetailError] = useState<string | null>(null);

  // Phase + active result
  const [phase, setPhase] = useState<OptPhase>(needsJobSelection ? "idle" : "loading_latest");
  const [activeResult, setActiveResult] = useState<OptimizationV2ViewModel | null>(null);
  const [activeTier, setActiveTier] = useState<string>("fast");
  const [activeGeneratedBy, setActiveGeneratedBy] = useState<string>("");
  const [optError, setOptError] = useState<string | null>(null);
  const [deepBackendDone, setDeepBackendDone] = useState(false);

  // Modal
  const [showModal, setShowModal] = useState(false);

  // Match latest
  const [matchLatest, setMatchLatest] = useState<MatchLatestResponse | null>(null);
  const [matchLoaded, setMatchLoaded] = useState(false);

  // ── Show job selection prompt ──
  if (needsJobSelection) {
    return (
      <ContentPageShell maxWidthClassName="max-w-[1280px]" contentClassName="pb-12">
        <div className="py-20 text-center max-w-lg mx-auto">
          <div className="w-16 h-16 mx-auto mb-5 rounded-2xl bg-blue-50 flex items-center justify-center">
            <Briefcase className="w-7 h-7 text-[#0052D9]" />
          </div>
          <h2 className="text-lg font-semibold text-slate-800 mb-2">选择目标岗位以获得更精准的优化建议</h2>
          <p className="text-sm text-slate-500 mb-8 leading-relaxed">
            系统将结合目标岗位 JD 分析你的简历，生成可直接执行的修改建议。
            通用优化只能给出基础建议，针对性较弱。
          </p>
          <div className="flex items-center justify-center gap-3 flex-wrap">
            <button
              onClick={() => setShowModal(true)}
              className="inline-flex items-center gap-2 px-5 py-2.5 bg-[#0052D9] text-white text-sm font-medium rounded-lg hover:bg-[#003DA5] transition-colors"
            >
              <Search className="w-4 h-4" />
              选择目标岗位
            </button>
            <button
              onClick={() => {
                // Navigate to general mode
                if (parsedResumeId) {
                  navigate(`/resume-optimization?resumeId=${parsedResumeId}&mode=general`, { replace: true });
                }
              }}
              className="inline-flex items-center gap-1.5 px-4 py-2.5 text-[13px] text-slate-500 border border-slate-200 rounded-lg hover:bg-slate-50 transition-colors"
            >
              继续通用优化
            </button>
          </div>
          <OptimizationEntryModal
            open={showModal}
            onClose={() => setShowModal(false)}
            initialResumeId={parsedResumeId}
          />
        </div>
      </ContentPageShell>
    );
  }

  // ── URL validation errors ──
  if (!parsedResumeId) {
    return (
      <div className="max-w-[1280px] mx-auto px-6 py-12 text-center">
        <p className="text-gray-500 text-sm mb-3">缺少有效的简历 ID</p>
        <p className="text-gray-400 text-xs mb-4">请通过简历管理页面选择一份简历后再进行优化。</p>
        <Link to="/resumes" className="text-[#0052D9] text-sm underline">管理我的简历</Link>
      </div>
    );
  }

  // ── Load job ──
  useEffect(() => {
    if (!parsedJobId) { setLoadingJob(false); return; }
    getJobDetail(String(parsedJobId))
      .then(setJob)
      .catch(() => {})
      .finally(() => setLoadingJob(false));
  }, [parsedJobId]);

  // ── Load resumes ──
  useEffect(() => {
    loadResumeOptionViewModelsFromApi()
      .then((opts) => {
        setResumeOptions(opts);
        if (opts.length === 0) return;
        const matched = opts.find((o) => o.id === parsedResumeId);
        setSelectedOption(matched ?? null);
      })
      .catch((err) => {
        const authRedirect = getAuthRedirectIf401(err, location.pathname, location.search);
        if (authRedirect) { navigate(authRedirect, { replace: true }); return; }
      })
      .finally(() => setLoadingOptions(false));
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  // ── Load resume detail ──
  useEffect(() => {
    if (!selectedOption) return;
    const id = selectedOption.id;
    if (!id || isNaN(id)) return;
    setLoadingResumeDetail(true);
    setResumeDetailError(null);
    setSelectedResumeDetail(null);
    loadResumeVersionDetail(id)
      .then((detail) => {
        if (!detail) setResumeDetailError("简历详情加载失败，请重新选择简历。");
        else setSelectedResumeDetail(detail);
      })
      .catch((err) => {
        const authRedirect = getAuthRedirectIf401(err, location.pathname, location.search);
        if (authRedirect) { navigate(authRedirect, { replace: true }); return; }
        setResumeDetailError("简历详情加载失败，请重新选择简历。");
      })
      .finally(() => setLoadingResumeDetail(false));
  }, [selectedOption]);

  // ── Load match latest ──
  useEffect(() => {
    if (!parsedJobId || !parsedResumeId) { setMatchLoaded(true); return; }
    setMatchLoaded(false);
    getLatestMatchAnalysis(String(parsedJobId), parsedResumeId)
      .then((ml) => { if (ml.has_result) setMatchLatest(ml); else setMatchLatest(null); })
      .catch(() => setMatchLatest(null))
      .finally(() => setMatchLoaded(true));
  }, [parsedResumeId, parsedJobId]);

  // ── Latest-first entry ──
  useEffect(() => {
    if (phase === "idle") return;
    if (!selectedOption || !selectedResumeDetail) return;
    const rid = selectedOption.id;
    if (!rid || isNaN(rid)) return;

    setPhase("loading_latest");
    const params: { optimization_mode?: string; target_direction?: string; job_id?: number } = {
      optimization_mode: derivedMode,
    };
    if (derivedMode === "target_direction") params.target_direction = targetDirection || "";
    if (derivedMode === "target_job") params.job_id = parsedJobId;

    loadLatestOptimizationV2ViewModel(rid, params)
      .then((latest) => {
        if (latest.hasResult && latest.result) {
          setActiveResult(latest.result);
          setActiveTier(latest.generationTier || "fast");
          setActiveGeneratedBy(latest.generatedBy || "");
          setPhase(latest.generationTier === "quality" ? "deep_done" : "fast_done");
        } else {
          setPhase("fast_loading");
          const genInput = {
            optimization_mode: derivedMode,
            generation_tier: "fast" as GenerationTier,
          } as any;
          if (derivedMode === "target_direction") genInput.target_direction = targetDirection || "";
          if (derivedMode === "target_job") genInput.job_id = parsedJobId;
          loadGenerateOptimizationV2ViewModel(rid, genInput)
            .then((r) => {
              setActiveResult(r);
              setActiveTier("fast");
              setActiveGeneratedBy(r.generatedBy || "");
              setPhase("fast_done");
            })
            .catch(() => {
              setOptError("快速优化生成失败");
              setPhase("fast_error");
            });
        }
      })
      .catch(() => {
        setOptError("读取优化结果失败");
        setPhase("fast_error");
      });
  }, [selectedOption, selectedResumeDetail]); // eslint-disable-line react-hooks/exhaustive-deps

  // ── Deep generation ──
  const handleDeepOptimization = () => {
    if (!selectedOption) return;
    const rid = selectedOption.id;
    if (!rid || isNaN(rid)) return;

    setPhase("deep_loading");
    setDeepBackendDone(false);
    const genInput = {
      optimization_mode: derivedMode,
      generation_tier: "quality" as GenerationTier,
    } as any;
    if (derivedMode === "target_direction") genInput.target_direction = targetDirection || "";
    if (derivedMode === "target_job") genInput.job_id = parsedJobId;
    loadGenerateOptimizationV2ViewModel(rid, genInput)
      .then((r) => {
        setActiveResult(r);
        setActiveTier("quality");
        setActiveGeneratedBy(r.generatedBy || "");
        setDeepBackendDone(true);
      })
      .catch(() => setPhase("deep_error"));
  };

  const handleReselect = () => {
    setShowModal(true);
  };

  // ── Guards ──
  if (loadingJob || loadingOptions) {
    return <PageLoading message="正在加载岗位和简历信息…" />;
  }

  if (resumeOptions.length === 0 || !selectedOption) {
    return (
      <div className="max-w-[1280px] mx-auto px-6 py-12 text-center">
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
      <div className="max-w-[1280px] mx-auto px-6 py-12 text-center">
        <p className="text-gray-500 text-sm mb-3">{resumeDetailError}</p>
        <Link to="/resumes" className="text-[#0052D9] text-sm underline">管理我的简历</Link>
      </div>
    );
  }

  if (phase === "loading_latest" || phase === "fast_loading" || phase === "idle") {
    return <PageLoading message="正在加载优化建议…" />;
  }

  if (phase === "fast_error" && !activeResult) {
    return (
      <div className="max-w-[1280px] mx-auto px-6 py-12 text-center">
        <p className="text-gray-500 text-sm mb-3">{optError || "优化接口暂时不可用"}</p>
        <Link to="/resumes" className="text-[#0052D9] text-sm underline">管理我的简历</Link>
      </div>
    );
  }

  const isFast = activeTier !== "quality";
  const headerLabel = getHeaderLabel(phase, activeTier, activeGeneratedBy);
  const modeLabel = getModeLabel(derivedMode, targetDirection || undefined, job?.title);
  const resumeName = selectedResumeDetail?.real_name || selectedOption.name;

  // Extract skill tags and edu for sidebar
  const skillTags = selectedResumeDetail?.professional_skills
    ? selectedResumeDetail.professional_skills.split(/[,，、\n]/).map((s) => s.trim()).filter(Boolean).slice(0, 5)
    : [];
  const topEdu = (selectedResumeDetail?.educations ?? []).find((e) => e.school || e.major || e.degree);
  const eduLine = topEdu ? [topEdu.degree, topEdu.major, topEdu.school].filter(Boolean).join(" · ") : "";
  const projectCount = (selectedResumeDetail?.projects ?? []).length;
  const internshipCount = (selectedResumeDetail?.internships ?? []).length;

  // Extract job keywords
  const jobKeywords = (job?.skill_keywords ?? []).slice(0, 5);

  // Tier display
  const tierLabel = activeTier === "quality" ? "深度优化" : "快速优化";

  // Match data
  const matchScore = matchLatest?.result?.match_score;
  const matchLevel = matchLatest?.result?.match_level;
  const matchGaps = matchLatest?.result?.gap_points ?? [];
  const matchHighlights = matchLatest?.result?.match_highlights ?? [];
  const matchDelivery = matchLatest?.result?.delivery_strategy;

  return (
    <><ContentPageShell maxWidthClassName="max-w-[1280px]" contentClassName="pb-12">

      {/* Agent Overlay for deep loading */}
      <AgentStepLoadingOverlay
        open={phase === "deep_loading"}
        title="正在生成深度优化报告"
        description="系统正在结合目标岗位 JD、匹配分析结果和当前简历，生成更完整的修改建议。"
        steps={[
          { key: "jd", label: "读取目标岗位 JD 与关键要求" },
          { key: "evidence", label: "对照当前简历中的项目、技能和经历证据" },
          { key: "suggestions", label: "生成可执行的修改建议与改写示例" },
          { key: "strategy", label: "汇总投递重点和后续补强策略" },
        ]}
        backendDone={deepBackendDone}
        onFinish={() => setPhase("deep_done")}
      />

      {/* ── General mode notice ── */}
      {derivedMode === "general" && (
        <div className="bg-amber-50 border border-amber-200 rounded-xl p-4 mb-6 flex items-start gap-3">
          <AlertTriangle className="w-4 h-4 text-amber-600 mt-0.5 shrink-0" />
          <div className="flex-1 min-w-0">
            <p className="text-sm font-medium text-amber-800 mb-1">通用优化模式</p>
            <p className="text-[12px] text-amber-700 leading-relaxed">
              当前未选择目标岗位，以下建议只适合做基础检查。选择目标岗位后，系统会结合 JD 给出更具体的修改建议。
            </p>
          </div>
          <button onClick={() => setShowModal(true)} className="px-3 py-1.5 text-[12px] font-medium text-[#0052D9] border border-blue-200 rounded-lg hover:bg-blue-50 transition-colors shrink-0 ml-2">
            选择目标岗位
          </button>
        </div>
      )}

      {/* ── Top: Optimization Context ── */}
      <div className="bg-white rounded-xl border border-slate-200 shadow-sm mb-6 overflow-hidden">
        <div className="flex flex-col sm:flex-row sm:items-center divide-y sm:divide-y-0 sm:divide-x divide-slate-100">
          {/* Resume */}
          <div className="flex items-start gap-3 px-5 py-4 flex-1 min-w-0">
            <div className="w-9 h-9 rounded-lg bg-blue-50 flex items-center justify-center shrink-0 mt-0.5">
              <FileText className="w-4 h-4 text-[#0052D9]" />
            </div>
            <div className="min-w-0">
              <p className="text-[11px] text-slate-400 font-medium">当前简历</p>
              <p className="text-sm font-semibold text-slate-800 truncate">{resumeName}</p>
              <div className="flex items-center gap-2 text-[11px] text-slate-500 mt-0.5 flex-wrap">
                {activeResult?.profileDominantProfile && (
                  <span className="text-[#0052D9]">简历定位：{humanizeProfile(activeResult.profileDominantProfile)}</span>
                )}
                {eduLine && (
                  <><span className="text-slate-300">|</span>
                  <span className="flex items-center gap-0.5"><GraduationCap className="w-2.5 h-2.5" />{eduLine}</span></>
                )}
                {(projectCount > 0 || internshipCount > 0) && (
                  <><span className="text-slate-300">|</span>
                  <span>{[projectCount > 0 && `${projectCount} 个项目`, internshipCount > 0 && `${internshipCount} 段实习`].filter(Boolean).join(" · ")}</span></>
                )}
              </div>
            </div>
          </div>

          {/* Target Job */}
          <div className="flex items-start gap-3 px-5 py-4 flex-1 min-w-0">
            <div className="w-9 h-9 rounded-lg bg-amber-50 flex items-center justify-center shrink-0 mt-0.5">
              <Briefcase className="w-4 h-4 text-amber-600" />
            </div>
            <div className="min-w-0">
              <p className="text-[11px] text-slate-400 font-medium">目标岗位</p>
              {job?.title ? (
                <>
                  <p className="text-sm font-semibold text-slate-800 truncate">{job.title}</p>
                  <p className="text-[11px] text-slate-500 mt-0.5">
                    {[job.direction, job.city, job.recruit_type, job.department].filter(Boolean).join(" · ")}
                  </p>
                  {jobKeywords.length > 0 && (
                    <div className="flex flex-wrap gap-1 mt-1">
                      {jobKeywords.map((kw, i) => (
                        <span key={i} className="px-1 py-0.5 bg-amber-50 text-amber-700 text-[10px] rounded">{kw}</span>
                      ))}
                    </div>
                  )}
                </>
              ) : derivedMode === "general" ? (
                <p className="text-sm text-slate-500">通用优化（无目标岗位）</p>
              ) : (
                <p className="text-sm text-slate-400 animate-pulse">加载中…</p>
              )}
            </div>
          </div>

          {/* Mode + Tier */}
          <div className="px-5 py-4 shrink-0">
            <div className="flex items-center gap-3">
              <div>
                <p className="text-[11px] text-slate-400 font-medium">优化详情</p>
                <p className="text-sm font-semibold text-slate-800">
                  {derivedMode === "target_job" ? "针对目标岗位" :
                   derivedMode === "target_direction" ? "针对目标方向" : "通用优化"}
                </p>
                <p className="text-[11px] text-slate-500 mt-0.5">
                  <span>{phase === "deep_done" ? "深度优化已完成" : tierLabel}</span>
                  {matchScore != null && (
                    <><span className="text-slate-300 mx-1.5">|</span>
                    <span className="text-[#0052D9]">匹配 {matchScore} 分</span></>
                  )}
                </p>
              </div>
              <button
                onClick={handleReselect}
                className="px-3 py-1.5 text-[12px] text-[#0052D9] border border-blue-200 rounded-lg hover:bg-blue-50 transition-colors shrink-0"
              >
                重新选择
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* ── Deep status notice ── */}
      {activeResult && !isFast && (
        <div className="bg-emerald-50 border border-emerald-200 rounded-xl p-4 mb-6 flex items-start gap-3">
          <div className="w-8 h-8 rounded-lg bg-emerald-100 flex items-center justify-center shrink-0 mt-0.5">
            <CheckCheck className="w-4 h-4 text-emerald-600" />
          </div>
          <div>
            <p className="text-sm font-semibold text-emerald-800 mb-0.5">深度优化报告已完成</p>
            <p className="text-[12px] text-emerald-700 leading-relaxed">
              基于目标岗位 JD、匹配分析结果和当前简历生成。可重新生成或查看匹配分析获取更多依据。
            </p>
          </div>
        </div>
      )}

      {activeResult && (
        <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_320px] items-start">
          {/* ── Center content ── */}
          {isFast ? (
            <FastView data={activeResult} phase={phase} onDeepOptimization={handleDeepOptimization} />
          ) : (
            <DeepView data={activeResult} phase={phase} />
          )}

          {/* ── Right sidebar ── */}
          <RightSidebar
            resumeName={resumeName}
            profileDominantProfile={activeResult.profileDominantProfile}
            optimizationMode={derivedMode}
            targetDirection={targetDirection}
            jobTitle={job?.title}
            jobCity={job?.city}
            jobDirection={job?.direction}
            jobDepartment={job?.department}
            jobRecruitType={job?.recruit_type}
            jobKeywords={jobKeywords}
            jobId={parsedJobId}
            resumeId={parsedResumeId}
            resumeEdu={eduLine}
            projectCount={projectCount}
            internshipCount={internshipCount}
            skillTags={skillTags}
            phase={phase}
            isFast={isFast}
            matchScore={matchScore}
            matchLevel={matchLevel}
            matchGaps={matchGaps}
            matchHighlights={matchHighlights}
            matchLoaded={matchLoaded}
            onDeepOptimization={handleDeepOptimization}
            onReselect={handleReselect}
          />
        </div>
      )}
    </ContentPageShell>
    <OptimizationEntryModal
      open={showModal}
      onClose={() => setShowModal(false)}
      initialResumeId={parsedResumeId}
      initialJobId={parsedJobId}
    />
    </>
  );
}

// ─── Fast View ──────────────────────────────────────────────────────────

function FastView({ data, phase, onDeepOptimization }: { data: OptimizationV2ViewModel; phase: OptPhase; onDeepOptimization: () => void }) {
  const levelColor = getLevelColor(data.optimizationLevel);
  const topFixes = (data.quickFixItems ?? []).slice(0, 3);
  const isLoading = phase === "deep_loading";
  const hasError = phase === "deep_error";

  return (
    <div className="min-w-0 space-y-6">
      {/* ── Primary CTA: Upgrade to Deep ── */}
      <div className="bg-gradient-to-br from-[#F0F4FF] to-white border border-blue-200 rounded-xl p-6 shadow-sm">
        <div className="flex items-start gap-4">
          <div className="w-12 h-12 rounded-xl bg-[#0052D9]/10 flex items-center justify-center shrink-0">
            <Sparkles className="w-6 h-6 text-[#0052D9]" />
          </div>
          <div className="flex-1 min-w-0">
            <h3 className="text-base font-bold text-slate-800 mb-1">
              {hasError ? "深度优化报告生成失败" : "生成深度优化报告"}
            </h3>
            {hasError ? (
              <p className="text-[13px] text-slate-500 leading-relaxed mb-4">
                深度优化生成失败，已保留当前快速优化结果。请稍后重试。
              </p>
            ) : (
              <p className="text-[13px] text-slate-500 leading-relaxed mb-4">
                当前结果是快速预览，适合快速判断修改方向。生成深度优化后，系统会结合目标岗位 JD、匹配分析结果和简历证据，输出：投递前必须修改项、JD 要求与证据缺口、可直接参考的改写示例、关键词与证据补强建议，以及投递重点策略。
              </p>
            )}
            <button
              onClick={onDeepOptimization}
              disabled={isLoading}
              className={`inline-flex items-center gap-2 px-6 py-3 text-sm font-semibold rounded-lg transition-all ${
                isLoading
                  ? "bg-slate-200 text-slate-400 cursor-not-allowed"
                  : "bg-[#0052D9] text-white hover:bg-[#003DA5] shadow-md hover:shadow-lg"
              }`}
            >
              {isLoading ? (
                <><Loader2 className="w-4 h-4 animate-spin" /> 正在生成深度优化报告…</>
              ) : hasError ? (
                <><RefreshCw className="w-4 h-4" /> 重新生成深度优化报告</>
              ) : (
                <><Sparkles className="w-4 h-4" /> 生成深度优化报告</>
              )}
            </button>
          </div>
        </div>
      </div>

      {/* ── Fast Result Preview ── */}
      <div className="bg-white border border-gray-200 rounded-lg shadow-sm overflow-hidden">
        <div className="bg-slate-50 px-4 py-2 border-b border-gray-100 flex items-center gap-2">
          <span className="text-[11px] font-medium text-slate-500">快速优化结果 · 初步预览</span>
          {data.optimizationScore > 0 && (
            <span className="ml-auto text-[11px] text-slate-400">{data.optimizationScore} 分</span>
          )}
        </div>
        <div className="bg-[#F4F8FF] p-5 border-b border-blue-100">
          <h2 className="text-lg font-bold text-gray-900 mb-2">优化结论</h2>
          <div className="flex items-center gap-3 mb-2">
            <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium border ${levelColor}`}>
              {data.optimizationLevel || "待评估"}
            </span>
          </div>
          {data.oneSentenceSummary && (
            <p className="text-sm text-gray-600 leading-relaxed">{data.oneSentenceSummary}</p>
          )}
        </div>

        <div className="p-5 space-y-4">
          {topFixes.length > 0 && (
            <section>
              <h3 className="font-bold text-sm text-gray-900 mb-2 flex items-center gap-1.5">
                <Target className="w-3.5 h-3.5 text-red-500" />
                投递前先改 {topFixes.length} 件事
              </h3>
              <ul className="space-y-2">
                {topFixes.map((item, i) => (
                  <li key={i} className="text-[13px] text-gray-700 flex items-start gap-2 bg-gray-50 p-2.5 rounded-md border border-gray-100">
                    <span className="text-[#0052D9] font-medium shrink-0 mt-0.5">{i + 1}.</span>
                    <div>
                      <span className="font-medium text-gray-800">{cleanText(item.issue)}</span>
                      {item.action && <span className="text-gray-500 ml-1">→ {cleanText(item.action)}</span>}
                    </div>
                  </li>
                ))}
              </ul>
            </section>
          )}

          {data.riskPoints && data.riskPoints.length > 0 && (
            <section>
              <h3 className="font-bold text-sm text-amber-600 mb-2 flex items-center gap-1.5">
                <AlertTriangle className="w-3.5 h-3.5" />
                主要风险
              </h3>
              <BulletList items={data.riskPoints.slice(0, 2).map(r => cleanText(r))} tone="amber" size="sm" />
            </section>
          )}

          {data.deliveryFocus && (
            <section>
              <h3 className="font-bold text-sm text-gray-900 mb-2 flex items-center gap-1.5">
                <ArrowUp className="w-3.5 h-3.5 text-[#0052D9]" />
                投递重点
              </h3>
              <p className="text-[13px] text-gray-700 bg-blue-50/30 p-3 rounded-md border border-blue-100">{cleanText(data.deliveryFocus)}</p>
            </section>
          )}
        </div>
      </div>
    </div>
  );
}

// ─── Deep View ──────────────────────────────────────────────────────────

function DeepView({ data, phase }: { data: OptimizationV2ViewModel; phase: OptPhase }) {
  const levelColor = getLevelColor(data.optimizationLevel);

  // Collect all fix items for "Top 3 must-fix"
  const allFixes = [
    ...((data.quickFixItems ?? []).map((f) => ({ ...f, _src: "quick" }))),
    ...((data.evidenceFixItems ?? []).map((f) => ({ ...f, _src: "evidence" }))),
    ...((data.keywordFixItems ?? []).map((f) => ({ ...f, _src: "keyword" }))),
  ];
  const top3 = allFixes.slice(0, 3);

  // Other sections for reorganization
  const keywords = data.keywordFixItems ?? [];
  const evidences = data.evidenceFixItems ?? [];
  const contentFixes = data.contentFixItems ?? [];
  const structureFixes = data.structureFixItems ?? [];

  return (
    <div className="min-w-0 bg-white border border-gray-200 rounded-lg shadow-sm overflow-hidden">
      {/* Score & summary */}
      <div className="bg-gradient-to-br from-[#F4F8FF] to-white p-5 border-b border-blue-100">
        <h2 className="text-lg font-bold text-gray-900 mb-2">优化结论</h2>
        <div className="flex items-center gap-3 mb-2">
          <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium border ${levelColor}`}>
            {data.optimizationLevel || "待评估"}
          </span>
          <span className="text-sm font-semibold text-gray-700">{data.optimizationScore || 0} 分</span>
        </div>
        {data.oneSentenceSummary && (
          <p className="text-sm text-gray-600">{data.oneSentenceSummary}</p>
        )}
      </div>

      <div className="p-5 space-y-6">
        {/* ── 1. Top 3 Must-Fix ── */}
        {top3.length > 0 && (
          <section>
            <h3 className="font-bold text-base text-gray-900 mb-3 flex items-center gap-2">
              <Target className="w-4 h-4 text-red-500" />
              投递前必须修改
            </h3>
            <div className="space-y-2">
              {top3.map((item, i) => (
                <div key={i} className="flex items-start gap-3 bg-red-50/30 border border-red-100 p-3 rounded-lg">
                  <span className="flex-shrink-0 w-6 h-6 rounded-full bg-red-100 text-red-600 flex items-center justify-center text-xs font-bold">
                    P{i + 1}
                  </span>
                  <div className="flex-1 min-w-0">
                    <p className="text-[13px] font-semibold text-gray-800">{cleanText(item.issue)}</p>
                    {item.section && (
                      <p className="text-[11px] text-gray-400 mt-0.5">建议修改位置：{cleanText(item.section)}</p>
                    )}
                    {item.action && (
                      <p className="text-[12px] text-gray-600 mt-1">{cleanText(item.action)}</p>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </section>
        )}

        {/* ── 2. JD Evidence (keyword + evidence) ── */}
        {(evidences.length > 0 || keywords.length > 0) && (
          <section>
            <h3 className="font-bold text-sm text-gray-900 mb-3 flex items-center gap-1.5">
              <Info className="w-3.5 h-3.5 text-[#0052D9]" />
              JD 要求补强
            </h3>
            <div className="space-y-2">
              {evidences.slice(0, 4).map((item, i) => (
                <div key={`ev_${i}`} className="text-[13px] text-gray-700 flex items-start gap-2 bg-gray-50 p-2.5 rounded-md border border-gray-100">
                  <span className="text-[#0052D9] font-medium shrink-0 mt-0.5">{i + 1}.</span>
                  <div>
                    <span className="font-medium text-gray-800">{cleanText(item.issue)}</span>
                    {item.action && <span className="text-gray-500 ml-1">→ {cleanText(item.action)}</span>}
                  </div>
                </div>
              ))}
              {keywords.slice(0, 3).map((item, i) => (
                <div key={`kw_${i}`} className="text-[13px] text-gray-700 flex items-start gap-2 bg-gray-50 p-2.5 rounded-md border border-gray-100">
                  <span className="text-[#0052D9] font-medium shrink-0 mt-0.5">{evidences.slice(0, 4).length + i + 1}.</span>
                  <div>
                    <span className="font-medium text-gray-800">{cleanText(item.section)}：{cleanText(item.issue)}</span>
                    {item.action && <span className="text-gray-500 ml-1">→ {cleanText(item.action)}</span>}
                  </div>
                </div>
              ))}
            </div>
          </section>
        )}

        {/* ── 3. Structure + Content (lower priority) ── */}
        {(structureFixes.length > 0 || contentFixes.length > 0) && (
          <SectionCollapse title="结构与表达优化" defaultOpen={false}>
            <div className="space-y-2 pt-2">
              {structureFixes.map((item, i) => (
                <div key={`st_${i}`} className="text-[13px] text-gray-700 flex items-start gap-2 bg-gray-50 p-2.5 rounded-md border border-gray-100">
                  <span className="text-[#0052D9] font-medium shrink-0 mt-0.5">{i + 1}.</span>
                  <div>
                    <span className="font-medium text-gray-800">{cleanText(item.section)}：{cleanText(item.issue)}</span>
                    {item.action && <span className="text-gray-500 ml-1">→ {cleanText(item.action)}</span>}
                  </div>
                </div>
              ))}
              {contentFixes.map((item, i) => (
                <div key={`ct_${i}`} className="text-[13px] text-gray-700 flex items-start gap-2 bg-gray-50 p-2.5 rounded-md border border-gray-100">
                  <span className="text-[#0052D9] font-medium shrink-0 mt-0.5">{structureFixes.length + i + 1}.</span>
                  <div>
                    <span className="font-medium text-gray-800">{cleanText(item.section)}：{cleanText(item.issue)}</span>
                    {item.action && <span className="text-gray-500 ml-1">→ {cleanText(item.action)}</span>}
                  </div>
                </div>
              ))}
            </div>
          </SectionCollapse>
        )}

        {/* ── 4. Rewrite Examples (Visual Highlight) ── */}
        {data.rewriteExamples && data.rewriteExamples.length > 0 && (
          <section>
            <h3 className="font-bold text-base text-gray-900 mb-4 flex items-center gap-2">
              <Sparkles className="w-4 h-4 text-[#0052D9]" />
              改写示例
            </h3>
            <div className="space-y-4">
              {data.rewriteExamples.map((ex, i) => (
                <RewriteExampleCard key={i} ex={ex} />
              ))}
            </div>
          </section>
        )}

        {/* ── 5. Risk Points ── */}
        {data.riskPoints && data.riskPoints.length > 0 && (
          <section>
            <h3 className="font-bold text-sm text-amber-600 mb-2 flex items-center gap-1.5">
              <AlertTriangle className="w-3.5 h-3.5" />
              风险提醒
            </h3>
            <BulletList items={data.riskPoints.map(r => cleanText(r))} tone="amber" size="sm" />
          </section>
        )}

        {/* ── 6. Delivery Focus ── */}
        {data.deliveryFocus && (
          <section>
            <h3 className="font-bold text-sm text-gray-900 mb-2 flex items-center gap-1.5">
              <ArrowUp className="w-3.5 h-3.5 text-[#0052D9]" />
              投递重点
            </h3>
            <p className="text-[13px] text-gray-700 bg-blue-50/50 p-3 rounded-md border border-blue-100">{cleanText(data.deliveryFocus)}</p>
          </section>
        )}

        {/* ── 7. Long Term (degraded) ── */}
        {data.longTermImprovements && data.longTermImprovements.length > 0 && (
          <SectionCollapse title="长期补强建议" defaultOpen={false}>
            <div className="pt-2">
              <BulletList 
                items={data.longTermImprovements.map(item => (
                  <span className="font-medium text-gray-800">{cleanText(item.direction)}{item.note && <span className="text-gray-500 font-normal"> — {cleanText(item.note)}</span>}</span>
                ))} 
                tone="slate" 
                size="sm" 
              />
            </div>
          </SectionCollapse>
        )}
      </div>
    </div>
  );
}

// ─── Section Collapse ──────────────────────────────────────────────────────────

function SectionCollapse({ title, children, defaultOpen = false }: { title: string; children: React.ReactNode; defaultOpen?: boolean }) {
  const [open, setOpen] = useState(defaultOpen);
  return (
    <section className="border border-gray-100 rounded-lg overflow-hidden">
      <button
        onClick={() => setOpen(!open)}
        className="w-full flex items-center justify-between p-3 bg-gray-50 hover:bg-gray-100 transition-colors text-left"
      >
        <span className="text-[13px] font-medium text-gray-600">{title}</span>
        {open ? <ChevronUp className="w-3.5 h-3.5 text-gray-400" /> : <ChevronDown className="w-3.5 h-3.5 text-gray-400" />}
      </button>
      {open && <div className="p-3">{children}</div>}
    </section>
  );
}

// ─── Right Sidebar ──────────────────────────────────────────────────────

function RightSidebar({
  resumeName, profileDominantProfile, optimizationMode, targetDirection,
  jobTitle, jobCity, jobDirection, jobDepartment, jobRecruitType, jobKeywords,
  jobId, resumeId, resumeEdu, projectCount, internshipCount, skillTags, phase, isFast,
  matchScore, matchLevel, matchGaps, matchHighlights, matchLoaded,
  onDeepOptimization, onReselect,
}: {
  resumeName: string;
  profileDominantProfile: string;
  optimizationMode: string;
  targetDirection?: string | null;
  jobTitle?: string;
  jobCity?: string;
  jobDirection?: string;
  jobDepartment?: string;
  jobRecruitType?: string;
  jobKeywords: string[];
  jobId?: number;
  resumeId?: number | null;
  resumeEdu: string;
  projectCount: number;
  internshipCount: number;
  skillTags: string[];
  phase: OptPhase;
  isFast: boolean;
  matchScore?: number;
  matchLevel?: string;
  matchGaps: string[];
  matchHighlights: string[];
  matchLoaded: boolean;
  onDeepOptimization: () => void;
  onReselect: () => void;
}) {
  const hasResult = phase !== "fast_error" && phase !== "fast_loading" && phase !== "loading_latest" && phase !== "idle";

  return (
    <aside className="shrink-0 flex flex-col gap-4">
      {/* Current Resume */}
      <div className="bg-white border border-gray-200 rounded-lg overflow-hidden">
        <div className="px-4 py-3 border-b border-gray-100">
          <h3 className="text-[13px] font-medium text-gray-500">当前简历</h3>
        </div>
        <div className="p-4">
          <p className="text-sm font-bold text-gray-900 mb-1">{resumeName}</p>
          {profileDominantProfile && (
            <p className="text-[11px] text-[#0052D9]">简历定位：{humanizeProfile(profileDominantProfile)}</p>
          )}
          {(resumeEdu || projectCount > 0 || internshipCount > 0) && (
            <div className="flex items-center gap-2 text-[11px] text-gray-500 mt-2 flex-wrap">
              {resumeEdu && <span><GraduationCap className="w-2.5 h-2.5 inline mr-0.5" />{resumeEdu}</span>}
              {projectCount > 0 && <span>{projectCount} 个项目</span>}
              {internshipCount > 0 && <span>{internshipCount} 段实习</span>}
            </div>
          )}
          {skillTags.length > 0 && (
            <div className="flex flex-wrap gap-1 mt-2">
              {skillTags.map((tag, i) => (
                <span key={i} className="px-1.5 py-0.5 bg-slate-100 text-slate-600 text-[10px] rounded">
                  {tag}
                </span>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Target Job */}
      <div className="bg-white border border-gray-200 rounded-lg overflow-hidden">
        <div className="px-4 py-3 border-b border-gray-100">
          <h3 className="text-[13px] font-medium text-gray-500">目标岗位</h3>
        </div>
        <div className="p-4">
          {optimizationMode === "target_job" ? (
            <>
              <p className="text-sm font-bold text-gray-900 mb-1">{jobTitle || "加载中…"}</p>
              <p className="text-[11px] text-gray-500">
                {[jobDirection, jobCity, jobRecruitType, jobDepartment].filter(Boolean).join(" · ") || "—"}
              </p>
              {jobKeywords.length > 0 && (
                <div className="flex flex-wrap gap-1 mt-2">
                  {jobKeywords.map((kw, i) => (
                    <span key={i} className="px-1.5 py-0.5 bg-amber-50 text-amber-700 text-[10px] rounded">{kw}</span>
                  ))}
                </div>
              )}
            </>
          ) : optimizationMode === "target_direction" ? (
            <>
              <p className="text-sm font-bold text-gray-900 mb-1">{targetDirection || "未知方向"}</p>
              <p className="text-[11px] text-gray-500">针对目标方向优化</p>
            </>
          ) : (
            <>
              <p className="text-sm text-gray-500 mb-1">通用优化</p>
              <p className="text-[11px] text-amber-600">
                选择目标岗位后建议更具体
              </p>
            </>
          )}
        </div>
      </div>

      {/* Match Analysis Card */}
      {jobId && resumeId && (
        <div className="bg-white border border-gray-200 rounded-lg overflow-hidden">
          <div className="px-4 py-3 border-b border-gray-100">
            <h3 className="text-[13px] font-medium text-gray-500">匹配分析依据</h3>
          </div>
          <div className="p-4">
            {!matchLoaded ? (
              <p className="text-[12px] text-gray-400">加载中…</p>
            ) : matchScore != null ? (
              <>
                <div className="flex items-center gap-2 mb-2">
                  <span className="text-2xl font-bold text-[#0052D9]">{matchScore}</span>
                  <span className="text-[11px] text-gray-400">分</span>
                  {matchLevel && (
                    <span className="px-1.5 py-0.5 bg-blue-50 text-[#0052D9] text-[10px] rounded font-medium">{matchLevel}</span>
                  )}
                </div>
                <BulletList
                  items={matchGaps.slice(0, 2).map(g => { const t = cleanText(g); return t.length > 80 ? t.slice(0, 80) + "…" : t; })}
                  tone="amber"
                  size="xs"
                />
                <BulletList
                  items={matchHighlights.slice(0, 1).map(h => { const t = cleanText(h); return t.length > 80 ? t.slice(0, 80) + "…" : h; })}
                  tone="green"
                  size="xs"
                />
                <Link
                  to={`/match?resumeId=${resumeId}&jobId=${jobId}`}
                  className="mt-2 inline-flex items-center gap-1 text-[11px] text-[#0052D9] hover:underline"
                >
                  查看匹配分析 <ExternalLink className="w-3 h-3" />
                </Link>
              </>
            ) : (
              <>
                <p className="text-[12px] text-gray-500 mb-2">尚未生成匹配分析</p>
                <p className="text-[10px] text-gray-400 mb-2">可先生成匹配分析，获得更完整的优化依据。</p>
                <Link
                  to={`/match?resumeId=${resumeId}&jobId=${jobId}`}
                  className="inline-flex items-center gap-1 px-2.5 py-1 text-[11px] bg-[#0052D9] text-white rounded-md hover:bg-[#003DA5] transition-colors"
                >
                  生成匹配分析
                </Link>
              </>
            )}
          </div>
        </div>
      )}

      {/* Actions Panel */}
      <div className="bg-white border border-gray-200 rounded-lg overflow-hidden">
        <div className="px-4 py-3 border-b border-gray-100">
          <h3 className="text-[13px] font-medium text-gray-500">优化操作</h3>
        </div>
        <div className="p-4 space-y-3">
          {/* Status */}
          <div className="flex items-center gap-2">
            <div className={`w-2 h-2 rounded-full ${isFast ? "bg-amber-400" : "bg-emerald-400"}`} />
            <span className="text-[12px] text-gray-600">
              {isFast ? "当前：快速优化" : "深度优化已完成"}
            </span>
          </div>

          {/* Primary CTA */}
          {phase === "deep_loading" ? (
            <button disabled className="w-full flex items-center justify-center gap-2 px-4 py-3 bg-slate-200 text-slate-500 text-sm font-semibold rounded-lg cursor-not-allowed">
              <Loader2 className="w-4 h-4 animate-spin" /> 正在生成深度优化报告…
            </button>
          ) : phase === "deep_done" || !isFast ? (
            <button onClick={onDeepOptimization} className="w-full flex items-center justify-center gap-2 px-4 py-3 bg-white border border-blue-200 text-[#0052D9] text-sm font-semibold rounded-lg hover:bg-blue-50 transition-colors">
              <RefreshCw className="w-4 h-4" /> 重新生成深度优化报告
            </button>
          ) : phase === "deep_error" ? (
            <button onClick={onDeepOptimization} className="w-full flex items-center justify-center gap-2 px-4 py-3 bg-[#0052D9] text-white text-sm font-semibold rounded-lg hover:bg-[#003DA5] transition-colors">
              <RefreshCw className="w-4 h-4" /> 重新生成深度优化报告
            </button>
          ) : (
            <button onClick={onDeepOptimization} className="w-full flex items-center justify-center gap-2 px-4 py-3 bg-[#0052D9] text-white text-sm font-semibold rounded-lg hover:bg-[#003DA5] shadow-sm transition-colors">
              <Sparkles className="w-4 h-4" /> 生成深度优化报告
            </button>
          )}

          {/* Divider */}
          <div className="border-t border-gray-100" />

          {/* Secondary actions */}
          <div className="space-y-1">
            {jobId && resumeId && (
              <Link to={`/match?resumeId=${resumeId}&jobId=${jobId}`} className="w-full flex items-center gap-2 px-2 py-1.5 text-[12px] text-gray-500 hover:text-[#0052D9] hover:bg-gray-50 rounded transition-colors">
                <ExternalLink className="w-3 h-3" /> 查看匹配分析
              </Link>
            )}
            {jobId && (
              <Link to={`/jobs/${jobId}`} className="w-full flex items-center gap-2 px-2 py-1.5 text-[12px] text-gray-500 hover:text-[#0052D9] hover:bg-gray-50 rounded transition-colors">
                <Briefcase className="w-3 h-3" /> 返回岗位详情
              </Link>
            )}
            <button onClick={onReselect} className="w-full flex items-center gap-2 px-2 py-1.5 text-[12px] text-gray-500 hover:text-[#0052D9] hover:bg-gray-50 rounded transition-colors">
              <Search className="w-3 h-3" /> 重新选择简历/岗位
            </button>
          </div>
        </div>
      </div>
    </aside>
  );
}

// ─── Rewrite Example Card ──────────────────────────────────────────────────────

function RewriteExampleCard({ ex }: { ex: OptimizationRewriteExample }) {
  const [copied, setCopied] = useState(false);
  const handleCopy = () => {
    if (!ex.after) return;
    navigator.clipboard.writeText(ex.after).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }).catch(() => {});
  };

  return (
    <div className="border border-blue-200 rounded-lg overflow-hidden bg-white shadow-sm">
      <div className="bg-blue-50/50 px-4 py-2 border-b border-blue-100 flex items-center justify-between gap-2">
        <div className="flex items-center gap-2 min-w-0">
          {ex.section && <span className="text-xs font-bold text-[#0052D9] shrink-0">{ex.section}</span>}
          {ex.reason && <span className="text-xs text-gray-500 truncate">{cleanText(ex.reason)}</span>}
        </div>
        <button
          onClick={handleCopy}
          className="shrink-0 flex items-center gap-1 text-[11px] text-[#0052D9] hover:text-[#003DA5] transition-colors"
        >
          {copied ? <><CheckCheck className="w-3 h-3" /> 已复制</> : <><Copy className="w-3 h-3" /> 复制修改后</>}
        </button>
      </div>
      <div className="grid grid-cols-2 divide-x divide-gray-100">
        <div className="p-4 bg-gray-50/50">
          <div className="text-[11px] font-medium text-gray-400 mb-1.5">修改前</div>
          <p className="text-[13px] text-gray-600 whitespace-pre-wrap leading-relaxed">{cleanText(ex.before)}</p>
        </div>
        <div className="p-4 bg-blue-50/20">
          <div className="text-[11px] font-medium text-[#0052D9] mb-1.5">修改后</div>
          <p className="text-[13px] text-gray-800 whitespace-pre-wrap leading-relaxed">{cleanText(ex.after)}</p>
        </div>
      </div>
    </div>
  );
}

// ─── Helpers ────────────────────────────────────────────────────────────

function parseResumeIdFromQuery(v: string | null): number | null {
  if (v && /^\d+$/.test(v)) return Number(v);
  return null;
}

function parseJobIdFromQuery(v: string | null): number | null {
  if (v && /^\d+$/.test(v)) return Number(v);
  return null;
}
