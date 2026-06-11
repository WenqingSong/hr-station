import React, { useState, useEffect } from "react";
import { Link, useSearchParams, useNavigate, useLocation } from "react-router";
import { MatchAnalysisHeader } from "../components/match/MatchAnalysisHeader";
import { MatchAnalysisSidebarLeft } from "../components/match/MatchAnalysisSidebarLeft";
import { MatchAnalysisDetail } from "../components/match/MatchAnalysisDetail";
import { MatchAnalysisSidebarRight } from "../components/match/MatchAnalysisSidebarRight";
import { loadMatchAnalysisViewModel, loadLatestMatchViewModel, loadResumeOptionViewModelsFromApi, loadResumeVersionDetail, type MatchLatestResponse } from "../../api";
import { getJobDetail } from "../../api/jobs";
import type { JobApiItem } from "../../api/jobs";
import type { MatchAnalysisPageViewModel, ResumeOptionViewModel } from "../../adapters";
import type { Resume } from "../../types";
import { ContentPageShell } from "../components/layout/ContentPageShell";
import { getAuthRedirectIf401 } from "../utils/authRedirect";
import { humanizeProfile } from "../utils/profileLabel";
import { PageLoading } from "../components/PageState";
import { AgentStepLoadingOverlay } from "../components/common/AgentStepLoadingOverlay";
import AgentErrorState from "../components/common/AgentErrorState";
import { BulletList } from "../components/common/BulletList";
import { MatchAnalysisSkeleton } from "../components/match/MatchAnalysisSkeleton";
import { Badge } from "../components/ui/Badge";
import { Loader2, Sparkles, RefreshCw, AlertCircle } from "lucide-react";
import { isMockGeneratedBy } from "../utils/agentError";

// ─── Helpers ─────────────────────────────────────────────────────────────

function formatDateTime(value?: string | null): string {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`;
}

// ─── Phase ─────────────────────────────────────────────────────────────

type MatchPhase =
  | "loading_latest"
  | "fast_loading"
  | "fast_done"
  | "fast_error"
  | "deep_loading"
  | "deep_done"
  | "deep_error";

function parseResumeIdFromQuery(resumeId: string | null): number | null {
  if (resumeId && /^\d+$/.test(resumeId)) return Number(resumeId);
  return null;
}

// ─── Header Label ──────────────────────────────────────────────────────

function getHeaderLabel(phase: MatchPhase, tier: string | undefined, genBy: string | undefined): string {
  if (phase === "loading_latest") return "正在读取匹配分析结果";
  if (phase === "fast_loading") return "正在生成快速匹配分析";
  if (phase === "deep_loading") return "正在深度分析中";
  if (phase === "fast_done") return "快速匹配 · 标准匹配分析";
  if (phase === "deep_done") {
    if (isMockGeneratedBy(genBy)) return "深度匹配 · 模型连接失败";
    if (genBy === "llm_agent") return "深度匹配 · 智能匹配分析";
    return "深度匹配 · 基础匹配分析";
  }
  if (phase === "deep_error") return "深度匹配分析生成失败 · 已保留当前结果";
  if (phase === "fast_error") return "快速匹配分析生成失败";
  return "";
}

// ─── Component ─────────────────────────────────────────────────────────

export function JobMatchAnalysisPage() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const location = useLocation();
  const jobId = searchParams.get("jobId") || "";
  const resumeIdStr = searchParams.get("resumeId");

  const [job, setJob] = useState<JobApiItem | null>(null);
  const [loadingJob, setLoadingJob] = useState(!!jobId);
  const [jobNotFound, setJobNotFound] = useState(false);

  const [resumeOptions, setResumeOptions] = useState<ResumeOptionViewModel[]>([]);
  const [selectedOption, setSelectedOption] = useState<ResumeOptionViewModel | null>(null);
  const [loadingOptions, setLoadingOptions] = useState(true);
  const [urlResumeNotFound, setUrlResumeNotFound] = useState(false);

  const [selectedResumeDetail, setSelectedResumeDetail] = useState<Resume | null>(null);
  const [loadingResumeDetail, setLoadingResumeDetail] = useState(false);
  const [resumeDetailError, setResumeDetailError] = useState<string | null>(null);

  // K9-2-E: phase + active result
  const [phase, setPhase] = useState<MatchPhase>("loading_latest");
  const [activeResult, setActiveResult] = useState<MatchAnalysisPageViewModel | null>(null);
  const [activeTier, setActiveTier] = useState<string>("fast");
  const [activeGeneratedBy, setActiveGeneratedBy] = useState<string>("fast_tag_engine");
  const [matchError, setMatchError] = useState<string | null>(null);
  const [deepBackendDone, setDeepBackendDone] = useState(false);

  // ── Load job ──
  useEffect(() => {
    if (!jobId) { setLoadingJob(false); return; }
    getJobDetail(jobId)
      .then((j) => { if (j) setJob(j); else setJobNotFound(true); })
      .catch(() => { setJobNotFound(true); })
      .finally(() => setLoadingJob(false));
  }, [jobId]);

  // ── Load resumes ──
  useEffect(() => {
    loadResumeOptionViewModelsFromApi()
      .then((opts) => {
        setResumeOptions(opts);
        if (opts.length === 0) { setLoadingOptions(false); return; }
        const targetId = parseResumeIdFromQuery(resumeIdStr);
        if (targetId) {
          const found = opts.find((o) => o.id === targetId);
          if (!found) {
            setUrlResumeNotFound(true);
            setSelectedOption(null);
            setLoadingOptions(false);
            return;
          }
          setSelectedOption(found);
          navigate(`/match?jobId=${jobId}&resumeId=${found.id}`, { replace: true });
        } else {
          const cur = opts.find((o) => o.isCurrent) ?? opts[0];
          setSelectedOption(cur);
          navigate(`/match?jobId=${jobId}&resumeId=${cur.id}`, { replace: true });
        }
        setLoadingOptions(false);
      })
      .catch((err) => {
        const authRedirect = getAuthRedirectIf401(err, location.pathname, location.search);
        if (authRedirect) { navigate(authRedirect, { replace: true }); return; }
        setLoadingOptions(false);
      });
  }, []);

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
        if (!detail) setResumeDetailError("简历版本详情加载失败");
        else setSelectedResumeDetail(detail);
      })
      .catch((err) => {
        const authRedirect = getAuthRedirectIf401(err, location.pathname, location.search);
        if (authRedirect) { navigate(authRedirect, { replace: true }); return; }
        setResumeDetailError("简历版本详情加载失败");
      })
      .finally(() => setLoadingResumeDetail(false));
  }, [selectedOption]);

  // ── Latest-first entry ──
  useEffect(() => {
    if (!jobId || !selectedOption || !selectedResumeDetail) return;
    const rid = selectedOption.id;
    if (!rid || isNaN(rid)) return;

    setPhase("loading_latest");
    loadLatestMatchViewModel(jobId, rid)
      .then((latest) => {
        if (latest.hasResult && latest.result) {
          setActiveResult(latest.result);
          setActiveTier(latest.generationTier || "fast");
          setActiveGeneratedBy(latest.generatedBy || "");
          setPhase(latest.generationTier === "quality" ? "deep_done" : "fast_done");
        } else {
          // No result — auto-generate Fast
          setPhase("fast_loading");
          loadMatchAnalysisViewModel(jobId, { resume_id: rid, generation_tier: "fast" })
            .then((r) => {
              setActiveResult(r);
              setActiveTier("fast");
              setActiveGeneratedBy(r.generatedBy || "fast_tag_engine");
              setPhase("fast_done");
            })
            .catch(() => {
              setMatchError("快速匹配分析生成失败");
              setPhase("fast_error");
            });
        }
      })
      .catch(() => {
        setMatchError("读取匹配分析结果失败");
        setPhase("fast_error");
      });
  }, [jobId, selectedOption, selectedResumeDetail]);

  // ── Deep generation ──
  const handleDeepAnalysis = () => {
    if (!jobId || !selectedOption) return;
    const rid = selectedOption.id;
    if (!rid || isNaN(rid)) return;

    setPhase("deep_loading");
    setDeepBackendDone(false);
    const abort = new AbortController();
    loadMatchAnalysisViewModel(jobId, { resume_id: rid, generation_tier: "quality" }, abort.signal)
      .then((r) => {
        setActiveResult(r);
        setActiveTier("quality");
        const genBy = r.generatedBy || "";
        setActiveGeneratedBy(genBy);
        setDeepBackendDone(true);
        // If backend returned mock/fallback, treat as error
        if (isMockGeneratedBy(genBy) || r.fallback) {
          setPhase("deep_error");
        }
      })
      .catch(() => {
        setPhase("deep_error");
      });
    // K13-Stability-A: auto-abort after 90s to stay under backend LLM_TIMEOUT_SECONDS=120
    const timer = setTimeout(() => abort.abort(), 90000);
    // Cleanup timeout on success (abort resolves the promise → .catch sets deep_error)
    abort.signal.addEventListener("abort", () => clearTimeout(timer));
  };

  // ── Guard states ──
  // Missing jobId
  if (!jobId) {
    return (
      <div className="max-w-[1440px] mx-auto px-6 py-16 text-center">
        <AlertCircle className="w-10 h-10 text-gray-400 mx-auto mb-4" />
        <h2 className="text-lg font-semibold text-gray-900 mb-2">缺少目标岗位</h2>
        <p className="text-sm text-gray-500 max-w-md mx-auto mb-6">
          请先从岗位列表选择一个岗位，再进行匹配分析。
        </p>
        <Link to="/jobs" className="px-4 py-2 bg-[#0052D9] text-white text-sm rounded-lg hover:bg-[#0047BD]">返回岗位列表</Link>
      </div>
    );
  }

  // Job not found or load failed
  if (!loadingJob && jobNotFound) {
    return (
      <div className="max-w-[1440px] mx-auto px-6 py-16 text-center">
        <AlertCircle className="w-10 h-10 text-amber-500 mx-auto mb-4" />
        <h2 className="text-lg font-semibold text-gray-900 mb-2">岗位不存在或已下线</h2>
        <p className="text-sm text-gray-500 max-w-md mx-auto mb-6">
          当前链接中的岗位（ID: {jobId}）无法访问。请返回岗位列表重新选择。
        </p>
        <Link to="/jobs" className="px-4 py-2 bg-[#0052D9] text-white text-sm rounded-lg hover:bg-[#0047BD]">返回岗位列表</Link>
      </div>
    );
  }

  if (loadingJob || loadingOptions) {
    return <MatchAnalysisSkeleton message="正在加载岗位和简历信息…" />;
  }

  // URL resumeId doesn't belong to current user
  if (urlResumeNotFound) {
    return (
      <div className="max-w-[1440px] mx-auto px-6 py-16 text-center">
        <AlertCircle className="w-10 h-10 text-amber-500 mx-auto mb-4" />
        <h2 className="text-lg font-semibold text-gray-900 mb-2">简历不存在或不属于当前账号</h2>
        <p className="text-sm text-gray-500 max-w-md mx-auto mb-6">
          当前链接中的简历无法访问。请从你的简历列表中选择一份简历后再进行岗位匹配分析。
        </p>
        <Link to="/resumes" className="px-4 py-2 bg-gray-100 text-gray-700 text-sm rounded-lg hover:bg-gray-200">返回我的简历</Link>
      </div>
    );
  }

  // No resumes at all
  if (resumeOptions.length === 0) {
    return (
      <div className="max-w-[1440px] mx-auto px-6 py-16 text-center">
        <AlertCircle className="w-10 h-10 text-gray-400 mx-auto mb-4" />
        <h2 className="text-lg font-semibold text-gray-900 mb-2">你还没有可用于匹配的简历</h2>
        <p className="text-sm text-gray-500 max-w-md mx-auto mb-6">
          请先创建或导入一份简历，Offer捕手才能基于你的经历分析岗位匹配度。
        </p>
        <div className="flex justify-center gap-3">
          <Link to="/resumes/new" className="px-4 py-2 bg-[#0052D9] text-white text-sm rounded-lg hover:bg-[#0047BD]">去创建简历</Link>
          <Link to="/resumes" className="px-4 py-2 bg-gray-100 text-gray-700 text-sm rounded-lg hover:bg-gray-200">管理我的简历</Link>
        </div>
      </div>
    );
  }

  if (!selectedOption) {
    return (
      <div className="max-w-[1440px] mx-auto px-6 py-12 text-center">
        <p className="text-gray-500 text-sm mb-3">无法选择简历</p>
        <Link to="/resumes" className="text-[#0052D9] text-sm underline">管理我的简历</Link>
      </div>
    );
  }

  if (loadingResumeDetail) {
    return <MatchAnalysisSkeleton message="正在加载简历详情…" />;
  }

  if (resumeDetailError) {
    return (
      <div className="max-w-[1440px] mx-auto px-6 py-12 text-center">
        <p className="text-gray-500 text-sm mb-3">{resumeDetailError}</p>
        <Link to="/resumes" className="text-[#0052D9] text-sm underline">管理我的简历</Link>
      </div>
    );
  }

  if (phase === "loading_latest" || phase === "fast_loading") {
    return <MatchAnalysisSkeleton message={phase === "fast_loading" ? "正在生成快速匹配分析…" : "正在读取匹配分析结果…"} />;
  }

  if (phase === "fast_error" && !activeResult) {
    return (
      <div className="max-w-[1440px] mx-auto px-6 py-12 text-center">
        <p className="text-gray-500 text-sm mb-3">{matchError || "匹配分析接口暂时不可用"}</p>
        <Link to={`/jobs/${jobId}`} className="text-[#0052D9] text-sm underline">返回岗位详情</Link>
      </div>
    );
  }

  const isFast = activeTier !== "quality";
  const headerLabel = getHeaderLabel(phase, activeTier, activeGeneratedBy);

  return (
    <ContentPageShell maxWidthClassName="max-w-[1440px]" contentClassName="pb-12">

      {/* Agent Overlay for deep loading */}
      <AgentStepLoadingOverlay
        open={phase === "deep_loading"}
        title="正在深度匹配分析"
        description="Offer捕手正在结合简历定位、岗位要求和匹配证据，生成完整分析报告。"
        steps={[
          { key: "profile", label: "读取简历定位与技能信息" },
          { key: "requirements", label: "拆解岗位要求" },
          { key: "evidence", label: "计算匹配证据" },
          { key: "suggestions", label: "生成差距与补强建议" },
        ]}
        backendDone={deepBackendDone}
        onFinish={() => setPhase("deep_done")}
      />

      {/* Header */}
      <div className="mb-6">
        <h1 className="text-[22px] font-semibold text-gray-900 flex items-center gap-2">
          <Sparkles className="w-5 h-5 text-[#0052D9]" />
          岗位匹配分析
        </h1>
        <div className="flex flex-wrap items-center gap-3 mt-3 text-xs text-gray-500">
          <span className="text-[#0052D9] font-medium">{headerLabel}</span>
          {activeResult && (
            <>
              <span className="text-gray-300">|</span>
              <span>{job?.title || "目标岗位"}</span>
              <span className="text-gray-300">|</span>
              <span>简历：{selectedResumeDetail?.real_name || selectedOption.name}</span>
            </>
          )}
        </div>
      </div>

      {activeResult && (
        <div className="flex gap-6 items-start">
          {/* Left sidebar — job */}
          <MatchAnalysisSidebarLeft job={job ?? undefined} jobSource="目标岗位" jobRequirements={activeResult?.jobRequirements} />

          {/* Center detail — Fast vs Deep */}
          {isFast ? (
            /* ── Fast simplified ── */
            <div className="flex-1 min-w-0 max-w-[680px] bg-white border border-gray-200 rounded-lg shadow-sm overflow-hidden">
              <div className="bg-[#F4F8FF] p-5 border-b border-blue-100">
                <h2 className="text-lg font-bold text-gray-900 mb-2">匹配结论</h2>
                <div className="flex items-center gap-3 mb-2">
                  <Badge variant={activeResult.matchScore >= 80 ? "default" : activeResult.matchScore >= 70 ? "secondary" : "outline"} className="text-xs px-2">
                    {activeResult.matchLevel || "待评估"}
                  </Badge>
                  <span className="text-sm font-semibold text-gray-700">{activeResult.matchScore || 0} 分</span>
                </div>
                {activeResult.summary && (
                  <p className="text-sm text-gray-600">{activeResult.summary}</p>
                )}
              </div>
              <div className="p-5 space-y-4">
                {/* Core advantages */}
                {activeResult.matchHighlights && activeResult.matchHighlights.length > 0 && (
                  <section>
                    <h3 className="font-bold text-sm text-gray-900 mb-2">核心优势</h3>
                    <BulletList items={activeResult.matchHighlights.slice(0, 2).map(a => a)} tone="green" />
                  </section>
                )}
                {/* Main gap */}
                {activeResult.gapPoints && activeResult.gapPoints.length > 0 && (
                  <section>
                    <h3 className="font-bold text-sm text-amber-600 mb-2">主要短板</h3>
                    <p className="text-[13px] text-gray-700">{activeResult.gapPoints[0]}</p>
                  </section>
                )}
                {/* Main risk */}
                {activeResult.riskPoints && activeResult.riskPoints.length > 0 && (
                  <section>
                    <h3 className="font-bold text-sm text-amber-600 mb-2">风险提醒</h3>
                    <p className="text-[13px] text-gray-700">{activeResult.riskPoints[0]}</p>
                  </section>
                )}
                {/* Quick suggestion */}
                {activeResult.improvementSuggestions && activeResult.improvementSuggestions.length > 0 && (
                  <section>
                    <h3 className="font-bold text-sm text-gray-900 mb-2">快速建议</h3>
                    <p className="text-[13px] text-gray-700">{activeResult.improvementSuggestions[0]?.title || activeResult.improvementSuggestions[0]?.action}</p>
                  </section>
                )}
                {/* Upsell to Deep */}
                <div className="bg-blue-50/50 border border-blue-100 rounded-md p-3 mt-4">
                  <p className="text-[12px] text-blue-700">
                    点击右侧「生成深度匹配分析」，查看岗位要求拆解、简历证据对应、风险与补强建议、面试表达点和投递策略。
                  </p>
                </div>
              </div>
            </div>
          ) : (
            /* ── Deep complete ── */
            <MatchAnalysisDetail currentJobId={jobId} currentResumeId={String(selectedOption.id)} data={activeResult} />
          )}

          {/* Right sidebar — resume details + deep button */}
          <aside className="w-[300px] shrink-0 flex flex-col gap-4">
            <div className="bg-white border border-gray-200 rounded-lg p-4">
              <h3 className="text-[13px] font-medium text-gray-500 mb-3">当前简历</h3>
              <p className="text-sm font-bold text-gray-900 mb-1">{selectedOption.name}</p>
              {activeResult?.profileDominantProfile && (
                <p className="text-[11px] text-[#0052D9] mt-1 mb-2">简历定位：{humanizeProfile(activeResult.profileDominantProfile)}</p>
              )}
              {selectedOption.directionLabel && (
                <p className="text-[11px] text-gray-500">学校/专业：{selectedOption.directionLabel}</p>
              )}
              {selectedOption.tags && selectedOption.tags.length > 0 && (
                <div className="flex flex-wrap gap-1 mt-2">
                  {selectedOption.tags.slice(0, 4).map((t, i) => (
                    <span key={i} className="px-1.5 py-0.5 bg-slate-100 text-slate-600 text-[10px] rounded">{t}</span>
                  ))}
                </div>
              )}
              <p className="text-[11px] text-gray-400 mt-2">
                {selectedOption.projectCount > 0 ? `${selectedOption.projectCount} 个项目经历` : ""}
                {formatDateTime(selectedOption.updatedAt) ? ` · 更新于 ${formatDateTime(selectedOption.updatedAt)}` : ""}
              </p>
            </div>

            {activeResult?.resumeEvidence && activeResult.resumeEvidence.length > 0 && (
              <div className="bg-white border border-gray-200 rounded-lg p-4">
                <h3 className="text-[13px] font-medium text-gray-500 mb-2">相关简历证据</h3>
                <BulletList items={activeResult.resumeEvidence.slice(0, 4).map(ev => ev)} tone="green" size="sm" />
              </div>
            )}

            <div className="bg-white border border-gray-200 rounded-lg p-4">
              <h3 className="text-[13px] font-medium text-gray-500 mb-2">深度匹配分析</h3>
              <p className="text-[11px] text-gray-400 mb-3">
                {isFast
                  ? "基于简历定位、岗位JD和匹配证据生成完整分析报告。"
                  : "当前已展示深度匹配分析结果，可重新生成。"}
              </p>
              {phase === "deep_loading" ? (
                <div className="flex items-center gap-2 text-xs text-gray-500">
                  <Loader2 className="w-3.5 h-3.5 animate-spin" /> 智能体深度分析中…
                </div>
              ) : phase === "deep_done" ? (
                <div className="text-xs text-green-700 flex items-center gap-1.5">
                  <RefreshCw className="w-3 h-3" /> 深度分析已完成
                  <button onClick={handleDeepAnalysis} className="ml-1 text-[#0052D9] hover:underline">重新生成</button>
                </div>
              ) : phase === "deep_error" ? (
                <button onClick={handleDeepAnalysis} className="w-full flex items-center justify-center gap-1.5 px-3 py-2 bg-[#0052D9] text-white text-sm rounded-lg hover:bg-[#0047BD]">
                  <RefreshCw className="w-3.5 h-3.5" /> 重试深度匹配分析
                </button>
              ) : (
                <button onClick={handleDeepAnalysis} className="w-full flex items-center justify-center gap-1.5 px-3 py-2 bg-[#0052D9] text-white text-sm rounded-lg hover:bg-[#0047BD]">
                  <Sparkles className="w-3.5 h-3.5" /> 生成深度匹配分析
                </button>
              )}
            </div>
          </aside>
        </div>
      )}
    </ContentPageShell>
  );
}
