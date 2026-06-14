import React, { useState, useEffect } from "react";
import { Link, useSearchParams, useNavigate, useLocation } from "react-router";
import { loadMatchAnalysisViewModel, loadLatestMatchViewModel, loadResumeOptionViewModelsFromApi, loadResumeVersionDetail } from "../../api";
import { getJobDetail } from "../../api/jobs";
import type { JobApiItem } from "../../api/jobs";
import type { MatchAnalysisPageViewModel, ResumeOptionViewModel, ScreeningEliminationRiskVM, ResumeRewriteCommandVM } from "../../adapters";
import type { Resume } from "../../types";
import { ContentPageShell } from "../components/layout/ContentPageShell";
import { getAuthRedirectIf401 } from "../utils/authRedirect";
import { humanizeProfile } from "../utils/profileLabel";
import { PageError, PageEmpty } from "../components/PageState";
import { UnifiedAgentLoadingOverlay } from "../components/common/UnifiedAgentLoadingOverlay";
import { MatchAnalysisSkeleton } from "../components/match/MatchAnalysisSkeleton";
import { AppButton, PrimaryButton, SecondaryButton } from "../components/common/AppButton";
import { SectionCard } from "../components/common/SectionCard";
import { StatusBadge } from "../components/common/StatusBadge";
import { AiNotice } from "../components/common/AiNotice";
import { Loader2, Sparkles, RefreshCw, AlertCircle, ShieldCheck, ShieldX, PenTool, Compass, ChevronDown, ChevronUp, CheckCircle2, Info, FileText, Search, MessageSquare, Bug, Target } from "lucide-react";
import { isMockGeneratedBy } from "../utils/agentError";

// ─── Helpers ──────────────────────────────────────────────────────────
function formatDateTime(v?: string | null): string {
  if (!v) return ""; const d = new Date(v); if (Number.isNaN(d.getTime())) return "";
  return `${d.getFullYear()}-${String(d.getMonth()+1).padStart(2,"0")}-${String(d.getDate()).padStart(2,"0")} ${String(d.getHours()).padStart(2,"0")}:${String(d.getMinutes()).padStart(2,"0")}`;
}

type MatchPhase = "loading_latest" | "fast_loading" | "fast_done" | "fast_error" | "deep_loading" | "deep_done" | "deep_error";
function parseResumeIdFromQuery(s: string | null): number | null { if (s && /^\d+$/.test(s)) return Number(s); return null; }

// ─── Suggestion Priority Classification ───────────────────────────────
type PriorityLevel = "must_fix" | "suggested" | "optional";
type PriorityItem = { level: PriorityLevel; risk?: ScreeningEliminationRiskVM; command?: ResumeRewriteCommandVM; title: string; description: string; };

function classifyPriority(items: { risks: ScreeningEliminationRiskVM[]; commands: ResumeRewriteCommandVM[] }): { mustFix: PriorityItem[]; suggested: PriorityItem[]; optional: PriorityItem[] } {
  const mustFix: PriorityItem[] = [];
  const suggested: PriorityItem[] = [];
  const optional: PriorityItem[] = [];

  // Classify risks by severity
  for (const r of items.risks) {
    const item: PriorityItem = { level: r.severity === "high" ? "must_fix" : r.severity === "medium" ? "suggested" : "optional", risk: r, title: r.riskTitle, description: r.fixDirection || r.currentResumeProblem };
    if (item.level === "must_fix") mustFix.push(item);
    else if (item.level === "suggested") suggested.push(item);
    else optional.push(item);
  }

  // Distribute commands: first 2→must_fix, next 3→suggested, rest→optional
  items.commands.forEach((c, i) => {
    const item: PriorityItem = { level: i < 2 ? "must_fix" : i < 5 ? "suggested" : "optional", command: c, title: c.targetSection, description: c.problem };
    if (item.level === "must_fix") mustFix.push(item);
    else if (item.level === "suggested") suggested.push(item);
    else optional.push(item);
  });

  return { mustFix: mustFix.slice(0, 3), suggested: suggested.slice(0, 3), optional: optional.slice(0, 3) };
}

// ─── Readiness Boundary Logic ─────────────────────────────────────────
type ReadinessLevel = { canApply: boolean; label: string; description: string; mustComplete: string; bonus: string; boundary: string; cardBg: string; labelBg: string; textColor: string; iconColor: string; };
function getReadinessLevel(score: number): ReadinessLevel {
  if (score >= 80) return {
    canApply: true, label: "可投递", iconColor: "text-emerald-600", textColor: "text-emerald-700",
    cardBg: "bg-emerald-50/60 border-emerald-200", labelBg: "bg-emerald-100 text-emerald-700 border-emerald-300",
    description: "已具备较高投递基础，完成「必须修改」项后即可进入投递。",
    mustComplete: "完成「必须修改」的 1-2 项关键修改。",
    bonus: "「建议补强」「可选优化」属于加分项，不必等待全部完成。",
    boundary: "不需要为追求完美而无限修改。完成后即可投递。"
  };
  if (score >= 70) return {
    canApply: true, label: "重点优化后可投", iconColor: "text-blue-600", textColor: "text-blue-700",
    cardBg: "bg-blue-50/60 border-blue-200", labelBg: "bg-blue-100 text-blue-700 border-blue-300",
    description: "可以作为重点优化方向后投递，但需要先完成关键修改。",
    mustComplete: "建议先完成 2-3 个关键修改项。",
    bonus: "完成后即可进入投递，不必无限修改。",
    boundary: "如果投递截止日期较近，优先完成「必须修改」项即可。"
  };
  if (score >= 60) return {
    canApply: false, label: "需补强核心证据", iconColor: "text-amber-600", textColor: "text-amber-700",
    cardBg: "bg-amber-50/60 border-amber-200", labelBg: "bg-amber-100 text-amber-700 border-amber-300",
    description: "建议先补强核心证据再投递，否则初筛淘汰风险较高。",
    mustComplete: "建议补充岗位核心要求的 2-3 项证据。",
    bonus: "如果短期无法补齐，建议将此岗位作为备选。",
    boundary: "不建议在证据不足的情况下投递，优先关注更匹配的岗位。"
  };
  return {
    canApply: false, label: "暂不建议优先投递", iconColor: "text-red-600", textColor: "text-red-700",
    cardBg: "bg-red-50/60 border-red-200", labelBg: "bg-red-100 text-red-700 border-red-300",
    description: "当前匹配度较低，不建议将此岗位作为优先投递方向。",
    mustComplete: "如果不是强相关方向，建议转投更匹配的岗位。",
    bonus: "如果仍想尝试，需要从核心能力层面重新梳理简历。",
    boundary: "不做无限修改。先确认方向是否匹配，再决定是否投入。"
  };
}

// ─── Component: Collapsible Section ───────────────────────────────────
function CollapsibleSection({ title, icon, children, defaultOpen = false }: { title: string; icon: React.ReactNode; children: React.ReactNode; defaultOpen?: boolean; }) {
  const [open, setOpen] = useState(defaultOpen);
  return (
    <div className="bg-white border border-slate-200 rounded-2xl shadow-sm overflow-hidden">
      <button onClick={() => setOpen(!open)} className="w-full flex items-center justify-between px-6 py-4 hover:bg-slate-50 transition-colors">
        <div className="flex items-center gap-3">
          <span className="w-8 h-8 rounded-lg bg-slate-100 flex items-center justify-center">{icon}</span>
          <span className="text-base font-semibold text-slate-950">{title}</span>
        </div>
        {open ? <ChevronUp className="w-5 h-5 text-slate-400" /> : <ChevronDown className="w-5 h-5 text-slate-400" />}
      </button>
      {open && <div className="px-6 pb-6">{children}</div>}
    </div>
  );
}

// ─── Main Component ───────────────────────────────────────────────────
export function JobMatchAnalysisPage() {
  const [sp] = useSearchParams(); const nav = useNavigate(); const loc = useLocation();
  const jobId = sp.get("jobId") || ""; const resumeIdStr = sp.get("resumeId");
  const [job, setJob] = useState<JobApiItem | null>(null); const [loadingJob, setLoadingJob] = useState(!!jobId); const [jobNotFound, setJobNotFound] = useState(false);
  const [resumeOptions, setResumeOptions] = useState<ResumeOptionViewModel[]>([]); const [selectedOption, setSelectedOption] = useState<ResumeOptionViewModel | null>(null); const [loadingOptions, setLoadingOptions] = useState(true);
  const [selectedResumeDetail, setSelectedResumeDetail] = useState<Resume | null>(null); const [loadingResumeDetail, setLoadingResumeDetail] = useState(false);
  const [phase, setPhase] = useState<MatchPhase>("loading_latest");
  // Tier-isolated results: fast and deep are stored independently
  const [fastResult, setFastResult] = useState<MatchAnalysisPageViewModel | null>(null);
  const [deepResult, setDeepResult] = useState<MatchAnalysisPageViewModel | null>(null);
  const [deepExists, setDeepExists] = useState(false); // deep result is available in DB but not currently active
  const [activeTier, setActiveTier] = useState<"fast" | "quality">("fast");
  const [activeGenBy, setActiveGenBy] = useState("");
  const [matchError, setMatchError] = useState<string | null>(null);
  const [deepBackendDone, setDeepBackendDone] = useState(false);

  // Derive activeResult from tier-separated stores
  const activeResult = activeTier === "quality" && deepResult ? deepResult : fastResult;

  useEffect(() => { if (!jobId) { setLoadingJob(false); return; } getJobDetail(jobId).then(j => { if (j) setJob(j); else setJobNotFound(true); }).catch(() => setJobNotFound(true)).finally(() => setLoadingJob(false)); }, [jobId]);
  useEffect(() => { loadResumeOptionViewModelsFromApi().then(opts => { setResumeOptions(opts); if (!opts.length) { setLoadingOptions(false); return; } const tid = parseResumeIdFromQuery(resumeIdStr); if (tid) { const f = opts.find(o => o.id === tid); if (f) { setSelectedOption(f); nav(`/match?jobId=${jobId}&resumeId=${f.id}`, { replace: true }); } setLoadingOptions(false); } else { const cur = opts.find(o => o.isCurrent) ?? opts[0]; setSelectedOption(cur); nav(`/match?jobId=${jobId}&resumeId=${cur.id}`, { replace: true }); setLoadingOptions(false); } }).catch(err => { if (!getAuthRedirectIf401(err, loc.pathname, loc.search)) setLoadingOptions(false); }); }, []);
  useEffect(() => { if (!selectedOption) return; const id = selectedOption.id; setLoadingResumeDetail(true); loadResumeVersionDetail(id).then(d => { if (d) setSelectedResumeDetail(d); }).finally(() => setLoadingResumeDetail(false)); }, [selectedOption]);

  // Mount: always load fast first. Optionally check if deep exists in background.
  useEffect(() => { if (!jobId || !selectedOption || !selectedResumeDetail) return; const rid = selectedOption.id;
    setPhase("loading_latest");
    // Primary: load fast latest
    loadLatestMatchViewModel(jobId, rid, "fast").then(fastLatest => {
      if (fastLatest.hasResult && fastLatest.result) {
        setFastResult(fastLatest.result);
        setActiveTier("fast");
        setActiveGenBy(fastLatest.generatedBy || "");
        setPhase("fast_done");
      } else {
        // No fast result cached — generate fresh fast
        setPhase("fast_loading");
        loadMatchAnalysisViewModel(jobId, { resume_id: rid, generation_tier: "fast" }).then(r => {
          setFastResult(r); setActiveTier("fast"); setActiveGenBy(r.generatedBy || ""); setPhase("fast_done");
        }).catch(() => { setMatchError("快速分析失败"); setPhase("fast_error"); });
      }
      // Background: check if deep exists (don't affect current view)
      loadLatestMatchViewModel(jobId, rid, "quality").then(deepLatest => {
        if (deepLatest.hasResult && deepLatest.result) {
          setDeepResult(deepLatest.result);
          setDeepExists(true);
        }
      }).catch(() => {});
    }).catch(() => { setMatchError("读取分析结果失败"); setPhase("fast_error"); });
  }, [jobId, selectedOption, selectedResumeDetail]);

  // Switch to fast view
  const switchToFast = () => {
    if (!fastResult && jobId && selectedOption) {
      const rid = selectedOption.id;
      setPhase("fast_loading");
      loadMatchAnalysisViewModel(jobId, { resume_id: rid, generation_tier: "fast" }).then(r => {
        setFastResult(r); setActiveTier("fast"); setActiveGenBy(r.generatedBy || ""); setPhase("fast_done");
      }).catch(() => { setMatchError("快速分析失败"); setPhase("fast_error"); });
    } else {
      setActiveTier("fast");
      setPhase("fast_done");
    }
  };

  const handleDeep = () => { if (!jobId || !selectedOption) return; const rid = selectedOption.id; setPhase("deep_loading"); setDeepBackendDone(false); const abort = new AbortController(); loadMatchAnalysisViewModel(jobId, { resume_id: rid, generation_tier: "quality" }, abort.signal).then(r => { setDeepResult(r); setDeepExists(true); setActiveTier("quality"); setActiveGenBy(r.generatedBy || ""); setDeepBackendDone(true); if (isMockGeneratedBy(r.generatedBy || "") || r.reportVersion !== "screening_match_v1") setPhase("deep_error"); }).catch(() => setPhase("deep_error")); const t = setTimeout(() => abort.abort(), 90000); abort.signal.addEventListener("abort", () => clearTimeout(t)); };

  // ─── Early Returns ──────────────────────────────────────────────────
  if (!jobId) return <ContentPageShell maxWidthClassName="max-w-[1440px]"><PageError title="缺少目标岗位" description="请选择岗位后再进行匹配分析。"><Link to="/jobs" className="px-4 py-2 bg-[#0052D9] text-white text-sm rounded-lg">返回岗位列表</Link></PageError></ContentPageShell>;
  if (!loadingJob && jobNotFound) return <ContentPageShell maxWidthClassName="max-w-[1440px]"><PageError title="岗位不存在" description="请返回岗位列表重新选择。"><Link to="/jobs" className="px-4 py-2 bg-[#0052D9] text-white text-sm rounded-lg">返回岗位列表</Link></PageError></ContentPageShell>;
  if (loadingJob || loadingOptions) return <MatchAnalysisSkeleton message="正在加载…" />;
  if (!resumeOptions.length) return <ContentPageShell maxWidthClassName="max-w-[1440px]"><PageEmpty title="你还没有可用于匹配的简历" description="请先创建或导入一份简历。"><Link to="/resumes/new" className="px-4 py-2 bg-[#0052D9] text-white text-sm rounded-lg">去创建简历</Link></PageEmpty></ContentPageShell>;
  if (loadingResumeDetail || phase === "loading_latest" || phase === "fast_loading") return <MatchAnalysisSkeleton message="正在生成分析…" />;
  if (phase === "fast_error" && !activeResult) return <ContentPageShell maxWidthClassName="max-w-[1440px]"><PageError title="分析失败" description={matchError || "请稍后重试。"}><Link to={`/jobs/${jobId}`} className="text-[#0052D9] text-sm underline">返回岗位详情</Link></PageError></ContentPageShell>;

  // ─── Computed Values ─────────────────────────────────────────────────
  const isFast = activeTier !== "quality";
  const sd = activeResult?.screeningDecision;
  const title = isFast ? "快速初筛判断" : "初筛命中率提升方案";
  const score = sd?.score ?? activeResult?.matchScore ?? 0;
  const readiness = getReadinessLevel(score);
  const isMock = activeResult?.isDeepMockFallback || activeResult?.isDeepLegacyFallback;
  const risks = activeResult?.screeningEliminationRisks || [];
  const commands = activeResult?.resumeRewriteCommands || [];
  const priority = classifyPriority({ risks, commands });
  const hasPriorityContent = priority.mustFix.length > 0 || priority.suggested.length > 0 || priority.optional.length > 0;

  // ─── Decision label color ───────────────────────────────────────────
  const decisionLabelColor = (sd?.decision || "").includes("direct")
    ? "bg-emerald-50 text-emerald-700 border-emerald-200"
    : (sd?.decision || "").includes("revise")
      ? "bg-blue-50 text-blue-700 border-blue-200"
      : "bg-amber-50 text-amber-700 border-amber-200";

  return (
    <ContentPageShell maxWidthClassName="max-w-[1440px]" contentClassName="pb-12">
      <UnifiedAgentLoadingOverlay open={phase === "deep_loading"} title="深度匹配智能体正在分析" description="正在基于当前简历和岗位要求，生成岗位匹配结论、风险诊断和简历修改建议。" phaseLabel="Deep Match Agent" backendDone={deepBackendDone} onFinish={() => setPhase("deep_done")} />

      {/* ── 0. Profile Pending ───────────────────────────────────────── */}
      {activeResult?.isProfilePending && (
        <div className="bg-white border border-amber-200 rounded-3xl p-8 text-center">
          <div className="w-16 h-16 mx-auto mb-4 rounded-2xl bg-amber-50 flex items-center justify-center"><Loader2 className="w-7 h-7 text-amber-500 animate-spin" /></div>
          <h3 className="text-lg font-semibold text-slate-800 mb-2">语义画像准备中</h3>
          <p className="text-sm text-slate-500 max-w-md mx-auto">这不是简历能力判断，而是系统准备状态 — 画像就绪后即可生成初筛分析。</p>
          <button onClick={handleDeep} className="mt-5 px-5 py-2.5 bg-[#0052D9] text-white text-sm rounded-xl">重新生成</button>
        </div>
      )}

      {/* ── Main Content (Single Column, No Sidebar) ──────────────────── */}
      {activeResult && !activeResult.isProfilePending && (
        <div className="space-y-6">

          {/* ══════════════════════════════════════════════════════════════
              1. TOP INFO BAR — replaces former right sidebar
             ══════════════════════════════════════════════════════════════ */}
          <div className="bg-white border border-slate-200 rounded-2xl px-5 py-3.5 flex items-center gap-4 flex-wrap shadow-sm">
            {/* Job info */}
            <div className="flex items-center gap-2.5 min-w-0">
              <div className="w-8 h-8 rounded-lg bg-blue-50 flex items-center justify-center shrink-0">
                <FileText className="w-3.5 h-3.5 text-blue-500" />
              </div>
              <div className="min-w-0">
                <p className="text-xs text-slate-400 leading-none mb-0.5">岗位</p>
                <p className="text-sm font-semibold text-slate-900 whitespace-normal break-words">{job?.title || "—"}</p>
                {(job?.direction || job?.city) && (
                  <p className="text-xs text-slate-500 mt-0.5">{job?.direction}{job?.city ? ` · ${job.city}` : ""}{job?.type ? ` · ${job.type}` : ""}</p>
                )}
              </div>
            </div>
            {/* Divider */}
            <div className="w-px h-8 bg-slate-200 shrink-0 hidden sm:block" />
            {/* Resume info */}
            <div className="flex items-center gap-2.5 min-w-0">
              <div className="w-8 h-8 rounded-lg bg-slate-100 flex items-center justify-center shrink-0">
                <Search className="w-3.5 h-3.5 text-slate-500" />
              </div>
              <div className="min-w-0">
                <p className="text-xs text-slate-400 leading-none mb-0.5">简历</p>
                <p className="text-sm font-semibold text-slate-900 whitespace-normal break-words">{selectedOption?.name}</p>
                {activeResult?.profileDominantProfile && (
                  <p className="text-xs text-blue-600 mt-0.5">{humanizeProfile(activeResult.profileDominantProfile)}</p>
                )}
                {formatDateTime(selectedOption?.updatedAt) && (
                  <p className="text-xs text-slate-400 mt-0.5">更新于 {formatDateTime(selectedOption?.updatedAt)}</p>
                )}
              </div>
            </div>
            {/* Spacer */}
            <div className="flex-1 min-w-[20px]" />
            {/* Action buttons */}
            <div className="flex items-center gap-2 shrink-0">
              {isFast ? (
                <>
                  {deepExists && (
                    <AppButton variant="secondary" onClick={() => { setActiveTier("quality"); setPhase("deep_done"); }} icon={<Sparkles className="w-3.5 h-3.5" />}>
                      查看深度分析
                    </AppButton>
                  )}
                  <PrimaryButton onClick={handleDeep} icon={<PenTool className="w-3.5 h-3.5" />}>
                    生成深度分析
                  </PrimaryButton>
                </>
              ) : (
                <>
                  <PrimaryButton onClick={handleDeep} icon={<RefreshCw className="w-3.5 h-3.5" />}>
                    重新生成
                  </PrimaryButton>
                  <SecondaryButton onClick={switchToFast} icon={<RefreshCw className="w-3.5 h-3.5" />}>
                    快速初筛
                  </SecondaryButton>
                </>
              )}
              <Link to={`/jobs/${jobId}`} className="px-3.5 py-2 bg-slate-50 text-slate-600 text-sm rounded-lg hover:bg-slate-100 border border-slate-200 whitespace-nowrap">
                返回岗位
              </Link>
            </div>
          </div>

          {/* ══════════════════════════════════════════════════════════════
              2. DECISION HERO — the 5-second decision card
             ══════════════════════════════════════════════════════════════ */}
          <div className="bg-white border border-slate-200 rounded-3xl p-6 shadow-sm">
            <div className="flex items-start gap-5">
              {/* Score Circle */}
              <div className="w-[96px] h-[96px] rounded-2xl bg-gradient-to-br from-blue-500 to-cyan-400 flex items-center justify-center shrink-0 shadow-lg shadow-blue-500/20">
                <span className="text-5xl font-semibold text-white tracking-tight">{score}</span>
              </div>
              {/* Decision Content */}
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-3 mb-2">
                  <h2 className="text-xl font-semibold text-slate-950 tracking-tight">{sd?.level ?? activeResult?.matchLevel}</h2>
                  {activeResult?.decisionLabel && (
                    <StatusBadge tone={(sd?.decision||"").includes("direct")?"emerald":(sd?.decision||"").includes("revise")?"blue":"amber"}>{activeResult.decisionLabel}</StatusBadge>
                  )}
                  {isMock && <StatusBadge tone="amber">演示数据</StatusBadge>}
                  {activeResult?.isDeepLegacyFallback && !isMock && <StatusBadge tone="slate">兼容模式</StatusBadge>}
                </div>
                <p className="text-sm leading-7 text-slate-600">{activeResult.summary}</p>
                {(sd?.decisionReasons || []).length > 0 && (
                  <div className="flex flex-wrap gap-1.5 mt-3">
                    {sd!.decisionReasons.slice(0, 3).map((r, i) => (
                      <span key={i} className="text-xs text-slate-500 bg-slate-50 px-2 py-0.5 rounded">{r}</span>
                    ))}
                  </div>
                )}
              </div>
            </div>
            {/* AI boundary light prompt below hero */}
            <AiNotice variant="inline" tone="slate">AI 生成分析，仅供投递前参考，不替代官方要求与真实经历核实</AiNotice>
          </div>

          {/* ══════════════════════════════════════════════════════════════
              3. CORE THREE CARDS (Deep only — quick decision summary)
             ══════════════════════════════════════════════════════════════ */}
          {!isFast && (
            <div className="grid grid-cols-1 md:grid-cols-3 gap-5">
              {/* Why worth applying */}
              <div className="bg-white border border-slate-200 rounded-2xl p-6 shadow-sm">
                <div className="flex items-center gap-2.5 mb-4">
                  <span className="w-9 h-9 rounded-lg bg-emerald-100 flex items-center justify-center"><ShieldCheck className="w-4 h-4 text-emerald-600" /></span>
                  <h3 className="text-base font-semibold text-slate-950">为什么值得投</h3>
                </div>
                {(activeResult?.screeningPassPoints || []).length > 0 ? (
                  <ul className="space-y-2.5">
                    {(activeResult!.screeningPassPoints || []).slice(0, 3).map((p, i) => (
                      <li key={i} className="text-sm leading-6 text-slate-700 pl-3 border-l-2 border-emerald-200">{p.title}</li>
                    ))}
                  </ul>
                ) : (
                  <p className="text-sm text-slate-400">暂无命中信息</p>
                )}
              </div>

              {/* Maximum risk */}
              <div className="bg-white border border-slate-200 rounded-2xl p-6 shadow-sm">
                <div className="flex items-center gap-2.5 mb-4">
                  <span className="w-9 h-9 rounded-lg bg-amber-100 flex items-center justify-center"><AlertCircle className="w-4 h-4 text-amber-600" /></span>
                  <h3 className="text-base font-semibold text-slate-950">最大初筛风险</h3>
                </div>
                {(risks).length > 0 ? (
                  <ul className="space-y-2.5">
                    {risks.slice(0, 2).map((r, i) => (
                      <li key={i} className="text-sm leading-6 text-slate-700 pl-3 border-l-2 border-amber-300">
                        <span className="font-semibold">{r.riskTitle}</span>
                        <br />
                        <span className="text-xs text-slate-500">{r.currentResumeProblem}</span>
                      </li>
                    ))}
                  </ul>
                ) : (
                  <p className="text-sm text-slate-400">未发现显著风险</p>
                )}
              </div>

              {/* What to fix first */}
              <div className="bg-white border border-slate-200 rounded-2xl p-6 shadow-sm">
                <div className="flex items-center gap-2.5 mb-4">
                  <span className="w-9 h-9 rounded-lg bg-blue-100 flex items-center justify-center"><PenTool className="w-4 h-4 text-blue-600" /></span>
                  <h3 className="text-base font-semibold text-slate-950">现在先改哪里</h3>
                </div>
                {(commands).length > 0 ? (
                  <ul className="space-y-2.5">
                    {commands.slice(0, 3).map((c, i) => (
                      <li key={i} className="text-sm leading-6 text-slate-700 pl-3 border-l-2 border-blue-300">
                        {c.targetSection}
                        <span className="text-slate-500"> — {c.problem}</span>
                      </li>
                    ))}
                  </ul>
                ) : (
                  <p className="text-sm text-slate-400">暂未生成修改建议</p>
                )}
                {activeResult?.jobStrategy?.recommendedAction && (
                  <p className="text-sm text-blue-600 mt-3 font-medium leading-6">{activeResult.jobStrategy.recommendedAction}</p>
                )}
              </div>
            </div>
          )}

          {/* ══════════════════════════════════════════════════════════════
              4. READINESS BOUNDARY — "when can I stop modifying"
             ══════════════════════════════════════════════════════════════ */}
          <div className={`border rounded-2xl p-6 shadow-sm ${readiness.cardBg}`}>
            <div className="flex items-center gap-2.5 mb-4">
              <Target className={`w-5 h-5 ${readiness.iconColor}`} />
              <h3 className="text-base font-semibold text-slate-950">改到什么程度就可以投？</h3>
              <span className={`text-xs px-2.5 py-0.5 rounded-full font-medium border ${readiness.labelBg}`}>{readiness.label}</span>
            </div>
            <p className="text-sm leading-7 text-slate-700 mb-4">{readiness.description}</p>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div className="flex items-start gap-2">
                <CheckCircle2 className="w-4 h-4 text-emerald-500 mt-0.5 shrink-0" />
                <div>
                  <p className="text-xs font-semibold text-slate-500 mb-0.5">投递前事项</p>
                  <p className="text-sm leading-6 text-slate-800">{readiness.mustComplete}</p>
                </div>
              </div>
              <div className="flex items-start gap-2">
                <Info className="w-4 h-4 text-blue-500 mt-0.5 shrink-0" />
                <div>
                  <p className="text-xs font-semibold text-slate-500 mb-0.5">加分项说明</p>
                  <p className="text-sm leading-6 text-slate-800">{readiness.bonus}</p>
                </div>
              </div>
              <div className="flex items-start gap-2">
                <AlertCircle className="w-4 h-4 text-amber-500 mt-0.5 shrink-0" />
                <div>
                  <p className="text-xs font-semibold text-slate-500 mb-0.5">修改边界</p>
                  <p className="text-sm leading-6 text-slate-800">{readiness.boundary}</p>
                </div>
              </div>
            </div>
          </div>

          {/* ══════════════════════════════════════════════════════════════
              5. PRIORITY ACTION PLAN — must fix / suggested / optional
             ══════════════════════════════════════════════════════════════ */}
          {hasPriorityContent && (
            <div className="bg-white border border-slate-200 rounded-3xl p-6 shadow-sm">
              <h3 className="text-base font-semibold text-slate-950 mb-1 flex items-center gap-2"><Target className="w-4 h-4 text-slate-600" />修改优先级</h3>
              <p className="text-sm text-slate-500 leading-6 mb-5">以下按优先级排列，建议优先完成「必须修改」项后再进入投递。</p>
              <div className="space-y-5">
                {/* Must Fix */}
                {priority.mustFix.length > 0 && (
                  <div>
                    <h4 className="text-sm font-semibold text-red-600 mb-3 flex items-center gap-2"><span className="w-2 h-2 rounded-full bg-red-500" />必须修改（影响是否建议投递）</h4>
                    <div className="space-y-2.5">
                      {priority.mustFix.map((item, i) => (
                        <div key={i} className="border border-red-200 rounded-xl p-4 bg-red-50/30">
                          <p className="text-sm font-medium text-slate-900 mb-1">{item.risk?.riskTitle || item.command?.targetSection || item.title}</p>
                          <p className="text-sm leading-6 text-slate-700">{item.risk?.fixDirection || item.risk?.currentResumeProblem || item.command?.problem || item.description}</p>
                          {item.risk?.whyItBlocksScreening && (
                            <p className="text-xs text-slate-500 mt-1">{item.risk.whyItBlocksScreening}</p>
                          )}
                          {item.command?.rewriteInstruction && (
                            <p className="text-xs text-blue-600 mt-1">{item.command.rewriteInstruction}</p>
                          )}
                        </div>
                      ))}
                    </div>
                  </div>
                )}

                {/* Suggested */}
                {priority.suggested.length > 0 && (
                  <div>
                    <h4 className="text-sm font-semibold text-amber-600 mb-3 flex items-center gap-2"><span className="w-2 h-2 rounded-full bg-amber-500" />建议补强（提高命中率，非投递硬门槛）</h4>
                    <div className="space-y-2.5">
                      {priority.suggested.map((item, i) => (
                        <div key={i} className="border border-amber-200 rounded-xl p-4 bg-amber-50/30">
                          <p className="text-sm font-medium text-slate-900 mb-1">{item.risk?.riskTitle || item.command?.targetSection || item.title}</p>
                          <p className="text-sm leading-6 text-slate-700">{item.risk?.fixDirection || item.risk?.currentResumeProblem || item.command?.problem || item.description}</p>
                          {item.command?.rewriteInstruction && (
                            <p className="text-xs text-blue-600 mt-1">{item.command.rewriteInstruction}</p>
                          )}
                        </div>
                      ))}
                    </div>
                  </div>
                )}

                {/* Optional */}
                {priority.optional.length > 0 && (
                  <div>
                    <h4 className="text-sm font-semibold text-slate-500 mb-3 flex items-center gap-2"><span className="w-2 h-2 rounded-full bg-slate-400" />可选优化（锦上添花）</h4>
                    <div className="space-y-2.5">
                      {priority.optional.map((item, i) => (
                        <div key={i} className="border border-slate-200 rounded-xl p-4 bg-slate-50/50">
                          <p className="text-sm font-medium text-slate-900 mb-1">{item.risk?.riskTitle || item.command?.targetSection || item.title}</p>
                          <p className="text-sm leading-6 text-slate-700">{item.risk?.fixDirection || item.risk?.currentResumeProblem || item.command?.problem || item.description}</p>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </div>

              {/* If just one category has content, add the boundary note */}
              <p className="text-xs text-slate-400 mt-5 pt-4 border-t border-slate-100">
                完成「必须修改」项后即可准备投递，其余属于加分项，不必追求完美。
              </p>
            </div>
          )}

          {/* ══════════════════════════════════════════════════════════════
              6. EVIDENCE & COVERAGE (Deep mode — top-N summary)
             ══════════════════════════════════════════════════════════════ */}
          {!isFast && (
            <div className="space-y-5">
              {/* Job Strategy */}
              {activeResult?.jobStrategy && (
                <div className="bg-white border border-slate-200 rounded-2xl p-6 shadow-sm">
                  <h3 className="text-base font-semibold text-slate-950 mb-1 flex items-center gap-2"><Compass className="w-4 h-4 text-indigo-500" />下一步怎么投</h3>
                  <p className="text-sm text-slate-500 leading-6 mb-4">基于当前匹配结果的投递策略建议。</p>
                  <div className="flex items-center gap-2 mb-4">
                    <span className={`text-xs px-3 py-1 rounded-full font-medium ${activeResult.jobStrategy.positioning.includes("主投") ? "bg-emerald-50 text-emerald-700" : activeResult.jobStrategy.positioning.includes("备选") ? "bg-amber-50 text-amber-700" : "bg-red-50 text-red-600"}`}>
                      {activeResult.jobStrategy.positioning}
                    </span>
                  </div>
                  <p className="text-sm leading-7 text-slate-700 mb-3 whitespace-pre-wrap break-words">{activeResult.jobStrategy.recommendedAction}</p>
                  {activeResult.jobStrategy.ifApplyNow && (
                    <p className="text-sm text-slate-500 leading-6 mb-3">如果现在直接投：{activeResult.jobStrategy.ifApplyNow}</p>
                  )}
                  {activeResult.jobStrategy.betterFitDirections.length > 0 && (
                    <div className="mb-2">
                      <span className="text-xs font-medium text-slate-500">更适合方向：</span>
                      <div className="flex flex-wrap gap-1 mt-1">{activeResult.jobStrategy.betterFitDirections.map((d, i) => <span key={i} className="text-xs px-2 py-0.5 bg-blue-50 text-blue-600 rounded-full">{d}</span>)}</div>
                    </div>
                  )}
                  {activeResult.jobStrategy.avoidDirections.length > 0 && (
                    <div>
                      <span className="text-xs font-medium text-slate-500">不建议优先投：</span>
                      <div className="flex flex-wrap gap-1 mt-1">{activeResult.jobStrategy.avoidDirections.map((d, i) => <span key={i} className="text-xs px-2 py-0.5 bg-slate-100 text-slate-500 rounded-full">{d}</span>)}</div>
                    </div>
                  )}
                </div>
              )}

              {/* Pass Points — Top 3 only in main view */}
              <div className="bg-white border border-slate-200 rounded-2xl p-6 shadow-sm">
                <h3 className="text-base font-semibold text-slate-950 mb-1 flex items-center gap-2"><ShieldCheck className="w-4 h-4 text-emerald-500" />匹配优势（Top 3）</h3>
                <p className="text-sm text-slate-500 leading-6 mb-4">简历中能支撑该岗位初筛的核心证据。</p>
                {(activeResult?.screeningPassPoints || []).length > 0 ? (
                  <div className="space-y-4">
                    {(activeResult!.screeningPassPoints || []).slice(0, 3).map((p, i) => (
                      <div key={i} className="border border-emerald-100 rounded-xl p-4 bg-emerald-50/30">
                        <div className="flex items-center gap-2 mb-2">
                          <span className="text-xs font-semibold text-emerald-600 bg-emerald-100 px-2 py-0.5 rounded">命中 {i + 1}</span>
                          <span className="text-sm font-medium text-slate-800">{p.title}</span>
                        </div>
                        <p className="text-sm leading-7 text-slate-600 whitespace-pre-wrap break-words mb-2">{p.whyItHelps}</p>
                        {p.resumeEvidence && (
                          <div className="text-xs text-blue-600 bg-blue-50 px-3 py-2 rounded-lg border border-blue-100 font-mono leading-relaxed whitespace-pre-wrap break-words">{p.resumeEvidence}</div>
                        )}
                        <p className="text-xs text-emerald-600 mt-2">{p.screeningValue}</p>
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className="text-sm text-slate-400">当前简历缺少能支撑该岗位初筛的明确证据。</p>
                )}
              </div>

              {/* Elimination Risks — Top 3 only in main view */}
              <div className="bg-white border border-slate-200 rounded-2xl p-6 shadow-sm">
                <h3 className="text-base font-semibold text-slate-950 mb-1 flex items-center gap-2"><ShieldX className="w-4 h-4 text-red-500" />主要风险（Top 3）</h3>
                <p className="text-sm text-slate-500 leading-6 mb-4">投递前最该补强的表达，按风险等级排列。</p>
                {risks.length > 0 ? (
                  <div className="space-y-4">
                    {risks.slice(0, 3).map((r, i) => (
                      <div key={i} className="border rounded-xl p-4" style={{ borderColor: r.severity === "high" ? "#fca5a5" : r.severity === "medium" ? "#fdba74" : "#e2e8f0" }}>
                        <div className="flex items-center gap-2 mb-2">
                          <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${r.severity === "high" ? "bg-red-50 text-red-600" : r.severity === "medium" ? "bg-amber-50 text-amber-600" : "bg-slate-50 text-slate-500"}`}>
                            {r.severity === "high" ? "高风险" : r.severity === "medium" ? "中风险" : "低风险"}
                          </span>
                          <span className="text-sm font-medium text-slate-800">{r.riskTitle}</span>
                        </div>
                        <p className="text-sm leading-7 text-slate-600 whitespace-pre-wrap break-words mb-1">{r.whyItBlocksScreening}</p>
                        {r.currentResumeProblem && <p className="text-sm leading-6 text-slate-500 mb-1">当前问题：{r.currentResumeProblem}</p>}
                        {r.fixDirection && <p className="text-sm leading-6 text-blue-600">修复方向：{r.fixDirection}</p>}
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className="text-sm text-slate-400">未发现明显的初筛淘汰风险。</p>
                )}
              </div>
            </div>
          )}

          {/* ══════════════════════════════════════════════════════════════
              7. DETAILED ANALYSIS ACCORDION (collapsible deep content)
             ══════════════════════════════════════════════════════════════ */}
          {!isFast && (
            <div className="space-y-3">
              {/* Full Pass Points — all evidence, collapsible */}
              {(activeResult?.screeningPassPoints || []).length > 3 && (
                <CollapsibleSection title="完整命中证据" icon={<ShieldCheck className="w-4 h-4 text-emerald-500" />} defaultOpen={false}>
                  <p className="text-sm text-slate-500 leading-6 mb-4">所有匹配优势的完整证据列表（含 Top 3 之外的项目）。</p>
                  <div className="space-y-4">
                    {(activeResult!.screeningPassPoints || []).slice(3).map((p, i) => (
                      <div key={i} className="border border-emerald-100 rounded-xl p-4 bg-emerald-50/30">
                        <div className="flex items-center gap-2 mb-2">
                          <span className="text-xs font-semibold text-emerald-600 bg-emerald-100 px-2 py-0.5 rounded">命中 {i + 4}</span>
                          <span className="text-sm font-medium text-slate-800">{p.title}</span>
                        </div>
                        <p className="text-sm leading-7 text-slate-600 whitespace-pre-wrap break-words mb-2">{p.whyItHelps}</p>
                        {p.resumeEvidence && (
                          <div className="text-xs text-blue-600 bg-blue-50 px-3 py-2 rounded-lg border border-blue-100 font-mono leading-relaxed whitespace-pre-wrap break-words">{p.resumeEvidence}</div>
                        )}
                      </div>
                    ))}
                  </div>
                </CollapsibleSection>
              )}

              {/* Full Risks — beyond Top 3 */}
              {risks.length > 3 && (
                <CollapsibleSection title="全部风险详情" icon={<ShieldX className="w-4 h-4 text-red-400" />} defaultOpen={false}>
                  <p className="text-sm text-slate-500 leading-6 mb-4">超过 Top 3 的风险详情。</p>
                  <div className="space-y-4">
                    {risks.slice(3).map((r, i) => (
                      <div key={i} className="border rounded-xl p-4" style={{ borderColor: r.severity === "high" ? "#fca5a5" : r.severity === "medium" ? "#fdba74" : "#e2e8f0" }}>
                        <div className="flex items-center gap-2 mb-2">
                          <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${r.severity === "high" ? "bg-red-50 text-red-600" : r.severity === "medium" ? "bg-amber-50 text-amber-600" : "bg-slate-50 text-slate-500"}`}>
                            {r.severity === "high" ? "高风险" : r.severity === "medium" ? "中风险" : "低风险"}
                          </span>
                          <span className="text-sm font-medium text-slate-800">{r.riskTitle}</span>
                        </div>
                        <p className="text-sm leading-7 text-slate-600 whitespace-pre-wrap break-words mb-1">{r.whyItBlocksScreening}</p>
                        {r.currentResumeProblem && <p className="text-sm leading-6 text-slate-500 mb-1">当前问题：{r.currentResumeProblem}</p>}
                        {r.fixDirection && <p className="text-sm leading-6 text-blue-600">修复方向：{r.fixDirection}</p>}
                      </div>
                    ))}
                  </div>
                </CollapsibleSection>
              )}

              {/* Rewrite Commands (collapsible — long content) */}
              {commands.length > 0 && (
                <CollapsibleSection title="简历具体怎么改" icon={<PenTool className="w-4 h-4 text-[#0052D9]" />} defaultOpen={false}>
                  <p className="text-sm text-slate-500 leading-6 mb-4">把普通经历改成更容易通过初筛的表达。每条都有修改前/修改后对照。</p>
                  <div className="space-y-5">
                    {commands.map((c, i) => (
                      <div key={i} className="border border-blue-100 rounded-2xl p-4 bg-blue-50/20">
                        <div className="flex items-center gap-2 mb-3">
                          <span className="text-xs font-semibold text-white bg-[#0052D9] w-6 h-6 rounded-full flex items-center justify-center">{i + 1}</span>
                          <span className="text-sm font-medium text-slate-800">{c.targetSection}</span>
                        </div>
                        <p className="text-sm leading-7 text-slate-700 mb-2 whitespace-pre-wrap break-words"><strong>问题：</strong>{c.problem}</p>
                        <p className="text-sm leading-7 text-slate-700 mb-3 whitespace-pre-wrap break-words"><strong>改写方向：</strong>{c.rewriteInstruction}</p>
                        <div className="grid grid-cols-1 lg:grid-cols-2 gap-3 mb-3">
                          <div className="bg-red-50/50 border border-red-100 rounded-lg p-3">
                            <p className="text-xs text-red-500 font-medium mb-0.5">修改前</p>
                            <p className="text-sm leading-6 text-slate-600 whitespace-pre-wrap break-words">{c.beforeExample || "—"}</p>
                          </div>
                          <div className="bg-emerald-50/50 border border-emerald-100 rounded-lg p-3">
                            <p className="text-xs text-emerald-600 font-medium mb-0.5">修改后</p>
                            <p className="text-sm leading-6 text-slate-700 whitespace-pre-wrap break-words">{c.afterExample || "—"}</p>
                          </div>
                        </div>
                        <p className="text-sm leading-6 text-blue-700 bg-blue-50 px-3 py-2 rounded-lg whitespace-pre-wrap break-words">💡 {c.whyItImprovesScreening}</p>
                      </div>
                    ))}
                  </div>
                </CollapsibleSection>
              )}

              {/* Interview Prep Reference — placeholder for future */}
              <CollapsibleSection title="面试准备参考" icon={<MessageSquare className="w-4 h-4 text-slate-400" />} defaultOpen={false}>
                <p className="text-sm text-slate-500 leading-6">
                  此区域仅作提示：若该岗位进入面试阶段，建议基于上述风险与匹配优势提前准备回答方向。
                  具体面试准备建议将在后续版本中提供。
                </p>
              </CollapsibleSection>

              {/* Tech Debug Info (collapsible, default closed) */}
              <CollapsibleSection title="技术详情与调试信息" icon={<Bug className="w-4 h-4 text-slate-400" />} defaultOpen={false}>
                <div className="space-y-3 text-sm text-slate-600">
                  {activeGenBy && (
                    <div className="flex gap-3">
                      <span className="text-xs font-medium text-slate-400 shrink-0">生成来源</span>
                      <span className="font-mono text-xs text-slate-600">{activeGenBy}</span>
                    </div>
                  )}
                  {activeResult?.reportVersion && (
                    <div className="flex gap-3">
                      <span className="text-xs font-medium text-slate-400 shrink-0">报告版本</span>
                      <span className="font-mono text-xs text-slate-600">{activeResult.reportVersion}</span>
                    </div>
                  )}
                  <div className="flex gap-3">
                    <span className="text-xs font-medium text-slate-400 shrink-0">分析层级</span>
                    <span className="text-xs text-slate-600">{isFast ? "Fast 快速初筛" : "Deep 深度分析"}</span>
                  </div>
                  <p className="text-xs text-slate-400 leading-6">此区域仅供技术与调试参考，正常使用时不需要关注。</p>
                </div>
              </CollapsibleSection>

              {/* Mock / Legacy fallback warning */}
              {isMock && (
                <div className="bg-amber-50 border border-amber-200 rounded-2xl p-4 flex items-start gap-3">
                  <AlertCircle className="w-5 h-5 text-amber-500 mt-0.5 shrink-0" />
                  <div>
                    <p className="text-sm font-medium text-amber-800 mb-1">演示数据提示</p>
                    <p className="text-sm leading-6 text-amber-700">当前分析结果基于模拟数据生成，非真实 AI 判断，仅供参考页面效果。</p>
                  </div>
                </div>
              )}
              {activeResult?.isDeepLegacyFallback && !isMock && (
                <div className="bg-slate-50 border border-slate-200 rounded-2xl p-4 flex items-start gap-3">
                  <Info className="w-5 h-5 text-slate-500 mt-0.5 shrink-0" />
                  <div>
                    <p className="text-sm font-medium text-slate-700 mb-1">兼容模式提示</p>
                    <p className="text-sm leading-6 text-slate-600">当前分析使用兼容模式生成，分析深度和准确性可能低于最新版本。</p>
                  </div>
                </div>
              )}
            </div>
          )}

          {/* ══════════════════════════════════════════════════════════════
              8. AI DISCLAIMER — full boundary statement at page bottom
             ══════════════════════════════════════════════════════════════ */}
          <AiNotice tone="slate">
            本分析由 AI 根据你当前选择的简历与岗位描述生成，用于帮助你判断投递优先级和优化表达。
            请不要编造经历，最终投递内容应以你的真实经历和岗位官方要求为准。分析结论仅供投递前参考，
            不替代真实经历核实、岗位官方要求和人工判断。
          </AiNotice>

        </div>
      )}
    </ContentPageShell>
  );
}
