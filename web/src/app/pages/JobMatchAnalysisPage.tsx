import React, { useState, useEffect } from "react";
import { Link, useSearchParams, useNavigate, useLocation } from "react-router";
import { MatchAnalysisHeader } from "../components/match/MatchAnalysisHeader";
import { MatchAnalysisSidebarLeft } from "../components/match/MatchAnalysisSidebarLeft";
import { MatchAnalysisSidebarRight } from "../components/match/MatchAnalysisSidebarRight";
import { loadMatchAnalysisViewModel, loadLatestMatchViewModel, loadResumeOptionViewModelsFromApi, loadResumeVersionDetail, type MatchLatestResponse } from "../../api";
import { getJobDetail } from "../../api/jobs";
import type { JobApiItem } from "../../api/jobs";
import type { MatchAnalysisPageViewModel, ResumeOptionViewModel } from "../../adapters";
import type { Resume } from "../../types";
import { ContentPageShell } from "../components/layout/ContentPageShell";
import { getAuthRedirectIf401 } from "../utils/authRedirect";
import { humanizeProfile } from "../utils/profileLabel";
import { PageLoading, PageError, PageEmpty } from "../components/PageState";
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

// ─── H23-B: Score-to-Level mapping ──────────────────────────────────────

function matchLevelLabel(score: number): string {
  if (score >= 90) return "高度匹配";
  if (score >= 80) return "较高匹配";
  if (score >= 70) return "中等匹配";
  if (score >= 60) return "弱匹配";
  return "边缘匹配";
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
  const [activeGeneratedBy, setActiveGeneratedBy] = useState<string>("fast_sep_match_engine");
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
              setActiveGeneratedBy(r.generatedBy || "fast_sep_match_engine");
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
      <ContentPageShell maxWidthClassName="max-w-[1440px]" contentClassName="pb-12">
        <PageError
          title="缺少目标岗位"
          description="请先从岗位列表选择一个目标岗位，再进行匹配分析。"
        >
          <Link to="/jobs" className="px-4 py-2 bg-[#0052D9] text-white text-sm rounded-lg hover:bg-[#0047BD]">返回岗位列表</Link>
        </PageError>
      </ContentPageShell>
    );
  }

  // Job not found or load failed
  if (!loadingJob && jobNotFound) {
    return (
      <ContentPageShell maxWidthClassName="max-w-[1440px]" contentClassName="pb-12">
        <PageError
          title="岗位不存在或已下线"
          description={`当前链接中的岗位（ID: ${jobId}）无法访问。请返回岗位列表重新选择。`}
        >
          <Link to="/jobs" className="px-4 py-2 bg-[#0052D9] text-white text-sm rounded-lg hover:bg-[#0047BD]">返回岗位列表</Link>
        </PageError>
      </ContentPageShell>
    );
  }

  if (loadingJob || loadingOptions) {
    return <MatchAnalysisSkeleton message="正在加载岗位和简历信息…" />;
  }

  // URL resumeId doesn't belong to current user
  if (urlResumeNotFound) {
    return (
      <ContentPageShell maxWidthClassName="max-w-[1440px]" contentClassName="pb-12">
        <PageError
          title="简历不存在或不属于当前账号"
          description="当前链接中的简历无法访问。请从你的简历列表中选择一份简历后再进行岗位匹配分析。"
        >
          <Link to="/resumes" className="px-4 py-2 bg-gray-100 text-gray-700 text-sm rounded-lg hover:bg-gray-200">返回我的简历</Link>
        </PageError>
      </ContentPageShell>
    );
  }

  // No resumes at all
  if (resumeOptions.length === 0) {
    return (
      <ContentPageShell maxWidthClassName="max-w-[1440px]" contentClassName="pb-12">
        <PageEmpty
          title="你还没有可用于匹配的简历"
          description="请先创建或导入一份简历，Offer捕手才能基于你的经历分析岗位匹配度。"
        >
          <div className="flex justify-center gap-3">
            <Link to="/resumes/new" className="px-4 py-2 bg-[#0052D9] text-white text-sm rounded-lg hover:bg-[#0047BD]">去创建简历</Link>
            <Link to="/resumes" className="px-4 py-2 bg-gray-100 text-gray-700 text-sm rounded-lg hover:bg-gray-200">管理我的简历</Link>
          </div>
        </PageEmpty>
      </ContentPageShell>
    );
  }

  if (!selectedOption) {
    return (
      <ContentPageShell maxWidthClassName="max-w-[1440px]" contentClassName="pb-12">
        <PageError title="无法选择简历" description="请返回简历列表重新选择。">
          <Link to="/resumes" className="text-[#0052D9] text-sm underline">管理我的简历</Link>
        </PageError>
      </ContentPageShell>
    );
  }

  if (loadingResumeDetail) {
    return <MatchAnalysisSkeleton message="正在加载简历详情…" />;
  }

  if (resumeDetailError) {
    return (
      <ContentPageShell maxWidthClassName="max-w-[1440px]" contentClassName="pb-12">
        <PageError title="简历详情加载失败" description={resumeDetailError}>
          <Link to="/resumes" className="text-[#0052D9] text-sm underline">管理我的简历</Link>
        </PageError>
      </ContentPageShell>
    );
  }

  if (phase === "loading_latest" || phase === "fast_loading") {
    return <MatchAnalysisSkeleton message={phase === "fast_loading" ? "正在生成快速匹配分析…" : "正在读取匹配分析结果…"} />;
  }

  if (phase === "fast_error" && !activeResult) {
    return (
      <ContentPageShell maxWidthClassName="max-w-[1440px]" contentClassName="pb-12">
        <PageError
          title="匹配分析生成失败"
          description={matchError || "匹配分析接口暂时不可用"}
        >
          <Link to={`/jobs/${jobId}`} className="text-[#0052D9] text-sm underline">返回岗位详情</Link>
        </PageError>
      </ContentPageShell>
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
            /* ── H18-E Fast: 6-card modular report ── */
            <div className="flex-1 min-w-0 max-w-[680px] flex flex-col gap-4">
              {/* Card 1: Top Conclusion */}
              <div className="bg-white border border-slate-200 rounded-2xl p-5 shadow-sm">
                <div className="flex items-start gap-4">
                  <div className="w-16 h-16 rounded-2xl bg-gradient-to-br from-blue-500 to-cyan-400 flex items-center justify-center shrink-0 shadow-md shadow-blue-500/20">
                    <span className="text-xl font-bold text-white">{activeResult.matchScore}</span>
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-0.5">
                      <span className="text-[15px] font-bold text-slate-800">{matchLevelLabel(activeResult.matchScore)}</span>
                      {activeResult.decisionLabel && (
                        <span className="text-[11px] px-2 py-0.5 rounded-full bg-amber-50 text-amber-700 border border-amber-200">{activeResult.decisionLabel}</span>
                      )}
                    </div>
                    <p className="text-[13px] text-slate-600 leading-relaxed">{activeResult.summary || "暂无匹配结论"}</p>
                    <p className="text-[10px] text-slate-400 mt-1">{activeResult.matchScore >= 60 ? "基于语义画像快速分析" : "当前匹配度较低，建议针对性补强后再投递"}</p>
                  </div>
                </div>
              </div>

              {/* Card 2: Score Breakdown + Hard Gate */}
              {activeResult.scoreBreakdown && (
                <div className="bg-white border border-slate-200 rounded-2xl p-4 shadow-sm">
                  <h3 className="text-[11px] font-semibold text-slate-400 uppercase tracking-wide mb-3">分数拆解</h3>
                  <div className="grid grid-cols-5 gap-2 text-center">
                    {[
                      { k: "方向", v: activeResult.scoreBreakdown.directionFit },
                      { k: "覆盖", v: activeResult.scoreBreakdown.mustHaveCoverage },
                      { k: "证据", v: activeResult.scoreBreakdown.evidenceStrength },
                      { k: "经历", v: activeResult.scoreBreakdown.experienceRelevance },
                      { k: "风险", v: activeResult.scoreBreakdown.riskControl },
                    ].map(({ k, v }) => (
                      <div key={k}>
                        <div className="w-full h-1.5 bg-slate-100 rounded-full overflow-hidden mb-1">
                          <div className="h-full bg-[#0052D9] rounded-full" style={{ width: `${v ?? 0}%` }} />
                        </div>
                        <span className="text-[10px] text-slate-500">{k}</span>
                        <span className="text-[11px] font-mono text-slate-700 ml-1">{v ?? "-"}</span>
                      </div>
                    ))}
                  </div>
                  {activeResult.hardGate && (
                    <div className="mt-3 pt-3 border-t border-slate-100">
                      <div className="flex items-center gap-2 text-[11px]">
                        <span className="text-slate-400">质量门槛：</span>
                        <span className={`font-medium ${activeResult.hardGate.status === "passed" ? "text-emerald-600" : activeResult.hardGate.status === "blocked" ? "text-red-500" : "text-amber-600"}`}>
                          {activeResult.hardGate.status === "passed" ? "通过" : activeResult.hardGate.status === "blocked" ? "不建议" : "受限"}
                        </span>
                        {activeResult.hardGate.scoreCap < 100 && (
                          <span className="text-slate-400">· 分数上限 {activeResult.hardGate.scoreCap}</span>
                        )}
                      </div>
                    </div>
                  )}
                </div>
              )}

              {/* Card 3: Core Advantages + Main Gaps */}
              <div className="grid grid-cols-2 gap-4">
                <div className="bg-white border border-slate-200 rounded-2xl p-4 shadow-sm">
                  <h3 className="text-[11px] font-semibold text-slate-400 uppercase tracking-wide mb-2">核心优势</h3>
                  {(activeResult.coreAdvantages?.length ?? 0) > 0 ? (
                    <ul className="space-y-1.5">
                      {activeResult.coreAdvantages.slice(0, 4).map((a, i) => (
                        <li key={i} className="flex items-start gap-1.5 text-[12px] text-slate-700">
                          <span className="w-1 h-1 rounded-full bg-emerald-400 mt-1.5 shrink-0" />{a}
                        </li>
                      ))}
                    </ul>
                  ) : (
                    <p className="text-[11px] text-slate-400">当前简历暂未体现足够强的岗位匹配优势。</p>
                  )}
                </div>
                <div className="bg-white border border-slate-200 rounded-2xl p-4 shadow-sm">
                  <h3 className="text-[11px] font-semibold text-amber-600 uppercase tracking-wide mb-2">主要短板</h3>
                  <ul className="space-y-1.5">
                    {activeResult.mainGaps.slice(0, 4).map((g, i) => (
                      <li key={i} className="flex items-start gap-1.5 text-[12px] text-slate-700">
                        <span className="w-1 h-1 rounded-full bg-amber-400 mt-1.5 shrink-0" />{g}
                      </li>
                    ))}
                    {activeResult.mainGaps.length === 0 && <li className="text-[11px] text-slate-400">未识别明显短板</li>}
                  </ul>
                </div>
              </div>

              {/* Card 4: Risk Diagnosis */}
              <div className="bg-white border border-slate-200 rounded-2xl p-4 shadow-sm">
                <h3 className="text-[11px] font-semibold text-slate-400 uppercase tracking-wide mb-3">风险诊断</h3>
                {activeResult.structuredRisks.length > 0 ? (
                  <div className="space-y-3">
                    {activeResult.structuredRisks.slice(0, 3).map((r, i) => (
                      <div key={i} className="border border-slate-100 rounded-xl p-3">
                        <div className="flex items-center gap-2 mb-1">
                          <span className={`text-[10px] px-1.5 py-0.5 rounded-full ${r.riskLevel === "high" ? "bg-red-50 text-red-600" : r.riskLevel === "medium" ? "bg-amber-50 text-amber-600" : "bg-slate-50 text-slate-500"}`}>
                            {r.riskLevel === "high" ? "高风险" : r.riskLevel === "medium" ? "中风险" : "低风险"}
                          </span>
                          <span className="text-[13px] font-medium text-slate-800">{r.riskName}</span>
                        </div>
                        <p className="text-[11px] text-slate-500">{r.whyItMatters || r.evidenceGap}</p>
                        {r.fixSuggestion && <p className="text-[11px] text-blue-600 mt-1">{r.fixSuggestion}</p>}
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className="text-[11px] text-slate-400">当前未识别明显高风险，但仍建议结合 JD 微调简历。</p>
                )}
              </div>

              {/* Card 5: Requirement Coverage */}
              {activeResult.requirementCoverage.length > 0 && (
                <div className="bg-white border border-slate-200 rounded-2xl p-4 shadow-sm">
                  <h3 className="text-[11px] font-semibold text-slate-400 uppercase tracking-wide mb-3">要求覆盖</h3>
                  <div className="space-y-2">
                    {activeResult.requirementCoverage.slice(0, 6).map((c, i) => (
                      <div key={i} className="flex items-center gap-2 text-[11px]">
                        <div className="w-16 shrink-0">
                          <div className="w-full h-1 bg-slate-100 rounded-full overflow-hidden">
                            <div className={`h-full rounded-full ${c.coverageStatus === "strong" ? "bg-emerald-500" : c.coverageStatus === "partial" ? "bg-blue-400" : "bg-slate-300"}`} style={{ width: `${c.coverageScore}%` }} />
                          </div>
                        </div>
                        <span className="w-12 shrink-0 text-slate-400 text-[10px]">{c.requirementType === "must_have" ? "必备" : c.requirementType === "should_have" ? "重要" : "加分"}</span>
                        <span className="flex-1 text-slate-700 truncate">{c.requirementName}</span>
                        <span className={`shrink-0 ${c.coverageStatus === "strong" ? "text-emerald-600" : c.coverageStatus === "missing" ? "text-red-400" : "text-slate-400"}`}>{c.coverageScore}</span>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* Card 6: Action Suggestions + Deep upsell */}
              <div className="bg-white border border-slate-200 rounded-2xl p-4 shadow-sm">
                <h3 className="text-[11px] font-semibold text-slate-400 uppercase tracking-wide mb-2">下一步建议</h3>
                {activeResult.actionSuggestions.length > 0 ? (
                  <ul className="space-y-1.5 mb-3">
                    {activeResult.actionSuggestions.slice(0, 4).map((s, i) => (
                      <li key={i} className="flex items-start gap-1.5 text-[12px] text-slate-700">
                        <span className="text-[#0052D9] font-bold mt-0.5">{i + 1}.</span>{s}
                      </li>
                    ))}
                  </ul>
                ) : (
                  <p className="text-[11px] text-slate-400 mb-3">建议优先补充与该岗位核心要求相关的项目证据，再进行投递。</p>
                )}
                <div className="bg-blue-50/50 border border-blue-100 rounded-lg p-3 mt-2">
                  <p className="text-[11px] text-blue-700">
                    想进一步查看 JD 解构、简历改写方向和面试追问预测，可生成深度匹配分析。
                  </p>
                </div>
              </div>
            </div>
          ) : (
            /* ── H19-C Deep: 10-module Agent Report ── */
            <div className="flex-1 min-w-0 max-w-[720px] flex flex-col gap-4">
              {/* Card 1: Top Conclusion */}
              <div className="bg-white border border-slate-200 rounded-2xl p-5 shadow-sm">
                <div className="flex items-center gap-1 mb-3">
                  <span className="text-[10px] px-2 py-0.5 bg-blue-50 text-blue-600 rounded-full">深度岗位顾问报告</span>
                  {activeGeneratedBy === "deep_match_agent" && <span className="text-[10px] px-2 py-0.5 bg-emerald-50 text-emerald-600 rounded-full">多 Agent 分析</span>}
                </div>
                <div className="flex items-start gap-4">
                  <div className="w-20 h-20 rounded-2xl bg-gradient-to-br from-indigo-500 to-purple-400 flex items-center justify-center shrink-0 shadow-md shadow-indigo-500/20">
                    <span className="text-2xl font-bold text-white">{activeResult.matchScore}</span>
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-0.5">
                      <span className="text-[16px] font-bold text-slate-800">{matchLevelLabel(activeResult.matchScore)}</span>
                    </div>
                    <p className="text-[13px] text-slate-600 leading-relaxed">{activeResult.summary || "深度分析完成"}</p>
                    <p className="text-[10px] text-slate-400 mt-1">基于语义画像多 Agent 深度分析</p>
                  </div>
                </div>
              </div>

              {/* Card 2: JD Requirements + Gaps */}
              <div className="bg-white border border-slate-200 rounded-2xl p-4 shadow-sm">
                <h3 className="text-[11px] font-semibold text-slate-400 uppercase tracking-wide mb-3">JD 要求与差距</h3>
                {activeResult.gapPoints && activeResult.gapPoints.length > 0 ? (
                  <div className="grid grid-cols-2 gap-x-4 gap-y-1.5">
                    {activeResult.gapPoints.slice(0, 8).map((g, i) => (
                      <div key={i} className="flex items-start gap-1.5 text-[12px] text-slate-700">
                        <span className={`w-1 h-1 rounded-full mt-1.5 shrink-0 ${i < 3 ? "bg-red-400" : "bg-amber-400"}`} />{g}
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className="text-[11px] text-slate-400">当前未返回结构化 JD 要求分析，建议重新生成深度分析。</p>
                )}
              </div>

              {/* Card 3: Evidence Map — H22-E: use DeepMatchReport resume_evidence_mapping */}
              <div className="bg-white border border-slate-200 rounded-2xl p-4 shadow-sm">
                <h3 className="text-[11px] font-semibold text-slate-400 uppercase tracking-wide mb-3">简历证据映射</h3>
                {(activeResult.resumeEvidenceMapping && activeResult.resumeEvidenceMapping.length > 0) ? (
                  <div className="space-y-2">
                    {activeResult.resumeEvidenceMapping.slice(0, 6).map((ev: any, i: number) => (
                      <div key={i} className="border border-slate-100 rounded-xl p-3">
                        <div className="flex items-center gap-2 mb-1">
                          <span className={`text-[10px] px-1.5 py-0.5 rounded ${(ev.evidence_level || ev.evidenceLevel) >= 4 ? "bg-emerald-50 text-emerald-600" : (ev.evidence_level || ev.evidenceLevel) >= 3 ? "bg-blue-50 text-blue-600" : "bg-slate-50 text-slate-500"}`}>
                            L{ev.evidence_level || ev.evidenceLevel || "?"}
                          </span>
                          <span className="text-[12px] font-medium text-slate-700">
                            {ev.evidence_name || ev.evidenceName || ev.source_title || ev.sourceTitle}
                          </span>
                          {ev.source_type && (
                            <span className="text-[10px] text-slate-400 ml-auto">{ev.source_type || ev.sourceType}</span>
                          )}
                        </div>
                        {ev.evidence_text && (
                          <p className="text-[11px] text-slate-600 leading-relaxed">{ev.evidence_text || ev.evidenceText}</p>
                        )}
                      </div>
                    ))}
                  </div>
                ) : activeResult.evidenceChain && activeResult.evidenceChain.length > 0 ? (
                  <div className="space-y-2">
                    {activeResult.evidenceChain.slice(0, 5).map((ev, i) => (
                      <div key={i} className="flex items-center gap-2 text-[11px]">
                        <span className="w-24 shrink-0 text-slate-400 truncate">{ev.jdRequirement || ev.requirement}</span>
                        <span className={`px-1.5 py-0.5 rounded text-[10px] ${(ev as any).evidenceStrengthLabel === "匹配" ? "bg-emerald-50 text-emerald-600" : "bg-amber-50 text-amber-600"}`}>{(ev as any).evidenceStrengthLabel === "匹配" ? "已匹配" : "部分"}</span>
                        <span className="flex-1 text-slate-600 truncate">{ev.resumeEvidence || ev.resumePoint}</span>
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className="text-[11px] text-slate-400">本次分析未返回结构化证据，建议重新生成深度报告获取完整证据映射。</p>
                )}
              </div>

              {/* Card 4: Core Match + Advantages */}
              <div className="grid grid-cols-2 gap-4">
                <div className="bg-white border border-slate-200 rounded-2xl p-4 shadow-sm">
                  <h3 className="text-[11px] font-semibold text-slate-400 uppercase tracking-wide mb-2">核心匹配点</h3>
                  {activeResult.fitPoints && activeResult.fitPoints.length > 0 ? (
                    <ul className="space-y-1.5">
                      {activeResult.fitPoints.slice(0, 4).map((f, i) => (
                        <li key={i} className="flex items-start gap-1.5 text-[12px] text-slate-700">
                          <span className="w-1 h-1 rounded-full bg-emerald-400 mt-1.5 shrink-0" />{f}
                        </li>
                      ))}
                    </ul>
                  ) : activeResult.matchHighlights && activeResult.matchHighlights.length > 0 ? (
                    <ul className="space-y-1.5">
                      {activeResult.matchHighlights.slice(0, 4).map((h, i) => (
                        <li key={i} className="flex items-start gap-1.5 text-[12px] text-slate-700">
                          <span className="w-1 h-1 rounded-full bg-emerald-400 mt-1.5 shrink-0" />{h}
                        </li>
                      ))}
                    </ul>
                  ) : (
                    <p className="text-[11px] text-slate-400">当前未识别明显匹配优势。</p>
                  )}
                </div>
                <div className="bg-white border border-slate-200 rounded-2xl p-4 shadow-sm">
                  <h3 className="text-[11px] font-semibold text-amber-600 uppercase tracking-wide mb-2">主要短板</h3>
                  {activeResult.gapPoints && activeResult.gapPoints.length > 0 ? (
                    <ul className="space-y-1.5">
                      {activeResult.gapPoints.slice(0, 4).map((g, i) => (
                        <li key={i} className="flex items-start gap-1.5 text-[12px] text-slate-700">
                          <span className="w-1 h-1 rounded-full bg-amber-400 mt-1.5 shrink-0" />{g}
                        </li>
                      ))}
                    </ul>
                  ) : (
                    <p className="text-[11px] text-slate-400">未识别明显短板</p>
                  )}
                </div>
              </div>

              {/* Card 5: Risk Diagnosis */}
              <div className="bg-white border border-slate-200 rounded-2xl p-4 shadow-sm">
                <h3 className="text-[11px] font-semibold text-slate-400 uppercase tracking-wide mb-3">风险诊断</h3>
                {activeResult.riskPoints && activeResult.riskPoints.length > 0 ? (
                  <div className="space-y-2">
                    {activeResult.riskPoints.slice(0, 4).map((r, i) => (
                      <div key={i} className="border border-slate-100 rounded-xl p-3 flex items-start gap-3">
                        <span className={`mt-0.5 text-[10px] px-1.5 py-0.5 rounded-full ${typeof r === 'string' && r.length > 15 ? "bg-amber-50 text-amber-600" : "bg-red-50 text-red-500"}`}>
                          {typeof r === 'string' && r.length > 15 ? "中风险" : "高风险"}
                        </span>
                        <div>
                          <span className="text-[12px] text-slate-700">{r}</span>
                        </div>
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className="text-[11px] text-slate-400">当前未识别明显高风险。</p>
                )}
              </div>

              {/* Card 6: Delivery Strategy */}
              {activeResult.deliveryStrategy && (
                <div className="bg-white border border-slate-200 rounded-2xl p-4 shadow-sm">
                  <h3 className="text-[11px] font-semibold text-slate-400 uppercase tracking-wide mb-2">投递策略</h3>
                  <p className="text-[12px] text-slate-700 leading-relaxed">{activeResult.deliveryStrategy}</p>
                </div>
              )}

              {/* Card 7: Interview Talking Points */}
              {activeResult.interviewTalkingPoints && activeResult.interviewTalkingPoints.length > 0 && (
                <div className="bg-white border border-slate-200 rounded-2xl p-4 shadow-sm">
                  <h3 className="text-[11px] font-semibold text-slate-400 uppercase tracking-wide mb-3">面试表达要点</h3>
                  <div className="space-y-2">
                    {activeResult.interviewTalkingPoints.slice(0, 6).map((pt, i) => (
                      <div key={i} className="flex items-start gap-2 bg-purple-50/30 px-3 py-2 rounded-lg border border-purple-100/30">
                        <span className="w-5 h-5 rounded-full bg-purple-100 text-purple-700 text-[10px] font-bold flex items-center justify-center shrink-0 mt-0.5">{i + 1}</span>
                        <p className="text-[12px] text-slate-700">{pt}</p>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* H22-G: Requirement Coverage Card */}
              {activeResult.requirementCoverage && activeResult.requirementCoverage.length > 0 && (
                <div className="bg-white border border-slate-200 rounded-2xl p-4 shadow-sm">
                  <h3 className="text-[11px] font-semibold text-slate-400 uppercase tracking-wide mb-3">岗位要求覆盖</h3>
                  <div className="space-y-2">
                    {(activeResult.requirementCoverage as any[]).slice(0, 8).map((c: any, i: number) => (
                      <div key={i} className="flex items-center gap-2 text-[11px] border-b border-slate-50 pb-2">
                        <span className={`text-[10px] px-1.5 py-0.5 rounded shrink-0 ${
                          c.requirement_type === "must_have" ? "bg-red-50 text-red-600" :
                          c.requirement_type === "should_have" ? "bg-amber-50 text-amber-600" : "bg-slate-50 text-slate-500"
                        }`}>{c.requirement_type === "must_have" ? "必备" : c.requirement_type === "should_have" ? "重要" : "加分"}</span>
                        <span className="flex-1 text-slate-700 truncate">{c.requirement_name || c.requirementName}</span>
                        <span className={`shrink-0 font-mono text-[10px] ${
                          (c.coverage_status || c.coverageStatus) === "matched" ? "text-emerald-600" :
                          (c.coverage_status || c.coverageStatus) === "partially_matched" ? "text-blue-600" : "text-slate-400"
                        }`}>{c.coverage_score || c.coverageScore || "-"}</span>
                        <span className="text-[10px] text-slate-400 w-16 text-right shrink-0">{
                          (c.coverage_status || c.coverageStatus) === "matched" ? "已覆盖" :
                          (c.coverage_status || c.coverageStatus) === "partially_matched" ? "部分" : "缺失"
                        }</span>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* H22-G: Interview Risk Questions Card */}
              {activeResult.interviewRiskQuestions && activeResult.interviewRiskQuestions.length > 0 && (
                <div className="bg-white border border-slate-200 rounded-2xl p-4 shadow-sm">
                  <h3 className="text-[11px] font-semibold text-slate-400 uppercase tracking-wide mb-3">面试追问预测</h3>
                  <div className="space-y-2">
                    {(activeResult.interviewRiskQuestions as any[]).slice(0, 5).map((q: any, i: number) => (
                      <div key={i} className="border border-purple-100 rounded-xl p-3 bg-purple-50/20">
                        <div className="flex items-center gap-2 mb-1">
                          <span className="w-4 h-4 rounded-full bg-purple-100 text-purple-700 text-[10px] font-bold flex items-center justify-center shrink-0">{i + 1}</span>
                          <span className="text-[12px] font-medium text-slate-800">{q.question || q.Question}</span>
                          {q.riskLevel && (
                            <span className={`text-[9px] px-1 py-0.5 rounded ml-auto ${q.riskLevel === "high" ? "bg-red-50 text-red-500" : "bg-amber-50 text-amber-500"}`}>
                              {q.riskLevel === "high" ? "高风险" : "中风险"}
                            </span>
                          )}
                        </div>
                        {(q.whyAsked || q.why_asked) && (
                          <p className="text-[10px] text-slate-400 ml-6">为什么问：{q.whyAsked || q.why_asked}</p>
                        )}
                        {(q.answerStrategy || q.answer_strategy) && (
                          <p className="text-[10px] text-blue-600 ml-6 mt-0.5">回答方向：{q.answerStrategy || q.answer_strategy}</p>
                        )}
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* Card 8: Action + Rewrite suggestions */}
              <div className="bg-white border border-slate-200 rounded-2xl p-4 shadow-sm">
                <h3 className="text-[11px] font-semibold text-slate-400 uppercase tracking-wide mb-2">下一步行动建议</h3>
                {activeResult.nextActions && activeResult.nextActions.length > 0 ? (
                  <ul className="space-y-1.5">
                    {activeResult.nextActions.slice(0, 5).map((a, i) => (
                      <li key={i} className="flex items-start gap-1.5 text-[12px] text-slate-700">
                        <span className="text-[#0052D9] font-bold mt-0.5">{i + 1}.</span>{a}
                      </li>
                    ))}
                  </ul>
                ) : (
                  <p className="text-[11px] text-slate-400">建议优先补充与该岗位核心要求相关的项目证据。</p>
                )}
              </div>

              {/* Card 9: Regenerate CTA */}
              <div className="bg-blue-50/50 border border-blue-100 rounded-2xl p-4 text-center">
                <p className="text-[11px] text-blue-700 mb-2">多 Agent 深度分析已完成。可重新生成获取最新分析结果。</p>
              </div>
            </div>
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
                <BulletList items={activeResult.resumeEvidence.slice(0, 4).map(ev => typeof ev === 'string' ? ev : (ev as any).evidence_name || (ev as any).evidence_text || (ev as any).source_title || '')} tone="green" size="sm" />
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
