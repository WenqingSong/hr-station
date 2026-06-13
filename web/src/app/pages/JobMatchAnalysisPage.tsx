import React, { useState, useEffect } from "react";
import { Link, useSearchParams, useNavigate, useLocation } from "react-router";
import { MatchAnalysisHeader } from "../components/match/MatchAnalysisHeader";
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
import { MatchAnalysisSkeleton } from "../components/match/MatchAnalysisSkeleton";
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

// ─── Score-to-Level mapping ────────────────────────────────────────────

function matchLevelLabel(score: number): string {
  if (score >= 90) return "高度匹配";
  if (score >= 80) return "较高匹配";
  if (score >= 70) return "中等匹配";
  if (score >= 60) return "弱匹配";
  return "不建议投递";
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
          if (!found) { setUrlResumeNotFound(true); setSelectedOption(null); setLoadingOptions(false); return; }
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
          setPhase("fast_loading");
          loadMatchAnalysisViewModel(jobId, { resume_id: rid, generation_tier: "fast" })
            .then((r) => { setActiveResult(r); setActiveTier("fast"); setActiveGeneratedBy(r.generatedBy || "fast_sep_match_engine"); setPhase("fast_done"); })
            .catch(() => { setMatchError("快速匹配分析生成失败"); setPhase("fast_error"); });
        }
      })
      .catch(() => { setMatchError("读取匹配分析结果失败"); setPhase("fast_error"); });
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
        setActiveResult(r); setActiveTier("quality");
        setActiveGeneratedBy(r.generatedBy || ""); setDeepBackendDone(true);
        if (isMockGeneratedBy(r.generatedBy || "") || r.fallback) { setPhase("deep_error"); }
      })
      .catch(() => { setPhase("deep_error"); });
    const timer = setTimeout(() => abort.abort(), 90000);
    abort.signal.addEventListener("abort", () => clearTimeout(timer));
  };

  // ── Guard states ──
  if (!jobId) {
    return (<ContentPageShell maxWidthClassName="max-w-[1440px]" contentClassName="pb-12"><PageError title="缺少目标岗位" description="请先从岗位列表选择一个目标岗位，再进行匹配分析。"><Link to="/jobs" className="px-4 py-2 bg-[#0052D9] text-white text-sm rounded-lg hover:bg-[#0047BD]">返回岗位列表</Link></PageError></ContentPageShell>);
  }
  if (!loadingJob && jobNotFound) {
    return (<ContentPageShell maxWidthClassName="max-w-[1440px]" contentClassName="pb-12"><PageError title="岗位不存在或已下线" description={`当前链接中的岗位（ID: ${jobId}）无法访问。请返回岗位列表重新选择。`}><Link to="/jobs" className="px-4 py-2 bg-[#0052D9] text-white text-sm rounded-lg hover:bg-[#0047BD]">返回岗位列表</Link></PageError></ContentPageShell>);
  }
  if (loadingJob || loadingOptions) { return <MatchAnalysisSkeleton message="正在加载岗位和简历信息…" />; }
  if (urlResumeNotFound) {
    return (<ContentPageShell maxWidthClassName="max-w-[1440px]" contentClassName="pb-12"><PageError title="简历不存在或不属于当前账号" description="当前链接中的简历无法访问。请从你的简历列表中选择一份简历后再进行岗位匹配分析。"><Link to="/resumes" className="px-4 py-2 bg-gray-100 text-gray-700 text-sm rounded-lg hover:bg-gray-200">返回我的简历</Link></PageError></ContentPageShell>);
  }
  if (resumeOptions.length === 0) {
    return (<ContentPageShell maxWidthClassName="max-w-[1440px]" contentClassName="pb-12"><PageEmpty title="你还没有可用于匹配的简历" description="请先创建或导入一份简历，Offer捕手才能基于你的经历分析岗位匹配度。"><div className="flex justify-center gap-3"><Link to="/resumes/new" className="px-4 py-2 bg-[#0052D9] text-white text-sm rounded-lg hover:bg-[#0047BD]">去创建简历</Link><Link to="/resumes" className="px-4 py-2 bg-gray-100 text-gray-700 text-sm rounded-lg hover:bg-gray-200">管理我的简历</Link></div></PageEmpty></ContentPageShell>);
  }
  if (!selectedOption) {
    return (<ContentPageShell maxWidthClassName="max-w-[1440px]" contentClassName="pb-12"><PageError title="无法选择简历" description="请返回简历列表重新选择。"><Link to="/resumes" className="text-[#0052D9] text-sm underline">管理我的简历</Link></PageError></ContentPageShell>);
  }
  if (loadingResumeDetail) { return <MatchAnalysisSkeleton message="正在加载简历详情…" />; }
  if (resumeDetailError) {
    return (<ContentPageShell maxWidthClassName="max-w-[1440px]" contentClassName="pb-12"><PageError title="简历详情加载失败" description={resumeDetailError}><Link to="/resumes" className="text-[#0052D9] text-sm underline">管理我的简历</Link></PageError></ContentPageShell>);
  }
  if (phase === "loading_latest" || phase === "fast_loading") { return <MatchAnalysisSkeleton message={phase === "fast_loading" ? "正在生成快速匹配分析…" : "正在读取匹配分析结果…"} />; }
  if (phase === "fast_error" && !activeResult) {
    return (<ContentPageShell maxWidthClassName="max-w-[1440px]" contentClassName="pb-12"><PageError title="匹配分析生成失败" description={matchError || "匹配分析接口暂时不可用"}><Link to={`/jobs/${jobId}`} className="text-[#0052D9] text-sm underline">返回岗位详情</Link></PageError></ContentPageShell>);
  }

  const isFast = activeTier !== "quality";
  const headerLabel = getHeaderLabel(phase, activeTier, activeGeneratedBy);

  return (
    <ContentPageShell maxWidthClassName="max-w-[1280px]" contentClassName="pb-12">

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
      <div className="mb-5">
        <h1 className="text-[22px] font-semibold text-gray-900 flex items-center gap-2">
          <Sparkles className="w-5 h-5 text-[#0052D9]" />
          岗位匹配分析
        </h1>
        <div className="flex flex-wrap items-center gap-3 mt-2 text-xs text-gray-500">
          <span className="text-[#0052D9] font-medium">{headerLabel}</span>
          {activeResult && <><span className="text-gray-300">|</span><span>{job?.title || "目标岗位"}</span><span className="text-gray-300">|</span><span>简历：{selectedResumeDetail?.real_name || selectedOption.name}</span></>}
        </div>
      </div>

      {activeResult && (
        <div className="grid grid-cols-[minmax(0,1fr)_300px] gap-6 items-start">
          {/* ═══ Left: Main Content Area ══════════════════════════ */}
          <section className="min-w-0 space-y-5">

            {/* ── Status Notices ── */}
            {activeResult.isProfilePending && (
              <div className="bg-white border border-amber-200 rounded-3xl p-8 text-center shadow-sm">
                <div className="w-16 h-16 mx-auto mb-4 rounded-2xl bg-amber-50 flex items-center justify-center"><Loader2 className="w-7 h-7 text-amber-500 animate-spin" /></div>
                <h3 className="text-[18px] font-semibold text-slate-800 mb-2">语义画像准备中</h3>
                <p className="text-[14px] text-slate-500 max-w-[480px] mx-auto leading-relaxed">{activeResult.profilePendingReason || "系统正在准备简历和岗位的语义画像。画像完成后，重新生成即可获得更准确的岗位匹配分析。"}</p>
                <p className="text-[11px] text-slate-400 mt-3">这不是简历能力判断，而是系统准备状态 — 画像就绪后即可生成真实的岗位匹配分析。</p>
                <button onClick={handleDeepAnalysis} className="mt-5 px-5 py-2.5 bg-[#0052D9] text-white text-[13px] rounded-xl hover:bg-[#0047BD] transition-colors inline-flex items-center gap-2"><RefreshCw className="w-3.5 h-3.5" />重新生成匹配分析</button>
              </div>
            )}
            {activeResult.isDeepLegacyFallback && (
              <div className="bg-amber-50 border border-amber-200 rounded-2xl p-4 flex items-center gap-3"><AlertCircle className="w-4 h-4 text-amber-500 shrink-0" /><div className="text-[12px] text-amber-800">当前为兼容分析结果，可信度低于 Deep v3 语义顾问报告。</div><button onClick={handleDeepAnalysis} className="ml-auto shrink-0 text-[11px] text-amber-700 underline hover:text-amber-900">重新生成</button></div>
            )}
            {activeResult.isDeepMockFallback && (
              <div className="bg-red-50 border border-red-200 rounded-2xl p-4 flex items-center gap-3"><AlertCircle className="w-4 h-4 text-red-500 shrink-0" /><div className="text-[12px] text-red-800">当前为演示兜底结果，不代表真实匹配判断。</div></div>
            )}

            {/* ── Hero Decision Card ─────────────────────────── */}
            {!activeResult.isProfilePending && !activeResult.isDeepMockFallback && (
            <div className="bg-white border border-slate-200 rounded-3xl p-6 shadow-sm">
              <div className="flex items-start gap-5">
                <div className="w-[72px] h-[72px] rounded-2xl bg-gradient-to-br from-blue-500 to-cyan-400 flex items-center justify-center shrink-0 shadow-lg shadow-blue-500/20">
                  <span className="text-[28px] font-bold text-white">{activeResult.matchScore}</span>
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-1">
                    <span className={`text-[20px] font-bold ${activeResult.matchScore >= 80 ? "text-emerald-600" : activeResult.matchScore >= 60 ? "text-amber-600" : "text-slate-700"}`}>{matchLevelLabel(activeResult.matchScore)}</span>
                    {activeResult.decisionLabel && <span className="text-[11px] px-2.5 py-0.5 rounded-full bg-amber-50 text-amber-700 border border-amber-200 font-medium">{activeResult.decisionLabel}</span>}
                  </div>
                  <p className="text-[14px] text-slate-600 leading-relaxed">{activeResult.oneSentenceConclusion || activeResult.summary || "暂无匹配结论"}</p>
                  <div className="flex items-center gap-2 mt-2">
                    {isFast ? <span className="text-[10px] px-2 py-0.5 bg-blue-50 text-blue-500 rounded-full">快速初筛</span> : <span className="text-[10px] px-2 py-0.5 bg-indigo-50 text-indigo-500 rounded-full">深度顾问分析</span>}
                    {activeResult.isDeepLegacyFallback && <span className="text-[10px] px-2 py-0.5 bg-amber-50 text-amber-500 rounded-full">兼容分析</span>}
                  </div>
                </div>
                <div className="flex flex-col gap-2 shrink-0">
                  {isFast ? (
                    <button onClick={handleDeepAnalysis} disabled={phase === "deep_loading"} className="px-4 py-2 bg-[#0052D9] text-white text-[12px] rounded-xl hover:bg-[#0047BD] transition-colors inline-flex items-center gap-1.5">{phase === "deep_loading" ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : <Sparkles className="w-3.5 h-3.5" />}{phase === "deep_loading" ? "分析中" : "深度分析"}</button>
                  ) : (
                    <button onClick={() => { if (!jobId || !selectedOption) return; const rid = selectedOption.id; if (!rid || isNaN(rid)) return; setPhase("fast_loading"); loadMatchAnalysisViewModel(jobId, { resume_id: rid, generation_tier: "fast" }).then(r => { setActiveResult(r); setActiveTier("fast"); setActiveGeneratedBy(r.generatedBy || ""); setPhase("fast_done"); }).catch(() => { setMatchError("快速匹配分析生成失败"); setPhase("fast_error"); }); }} className="px-4 py-2 bg-slate-100 text-slate-700 text-[12px] rounded-xl hover:bg-slate-200 transition-colors inline-flex items-center gap-1.5"><RefreshCw className="w-3.5 h-3.5" />快速重检</button>
                  )}
                </div>
              </div>
            </div>
            )}

            {/* ── Three Key Insight Cards ────────────────────── */}
            {!activeResult.isProfilePending && (
            <div className="grid grid-cols-3 gap-4">
              <div className="bg-white border border-slate-200 rounded-2xl p-4 shadow-sm">
                <h4 className="text-[12px] font-semibold text-emerald-600 mb-2">为什么可以考虑？</h4>
                {activeResult.coreMatchPoints.length > 0 ? (
                  <ul className="space-y-2">{activeResult.coreMatchPoints.slice(0, 3).map((m: any, i: number) => <li key={i} className="flex items-start gap-2 text-[12px] text-slate-700"><span className="w-1.5 h-1.5 rounded-full bg-emerald-400 mt-1.5 shrink-0" /><span className="min-w-0">{m.point_name || m.pointName || m}</span></li>)}</ul>
                ) : activeResult.coreAdvantages.length > 0 ? (
                  <ul className="space-y-2">{activeResult.coreAdvantages.slice(0, 3).map((a, i) => <li key={i} className="flex items-start gap-2 text-[12px] text-slate-700"><span className="w-1.5 h-1.5 rounded-full bg-emerald-400 mt-1.5 shrink-0" /><span className="min-w-0">{a}</span></li>)}</ul>
                ) : activeResult.matchHighlights.length > 0 ? (
                  <ul className="space-y-2">{activeResult.matchHighlights.slice(0, 3).map((h, i) => <li key={i} className="flex items-start gap-2 text-[12px] text-slate-700"><span className="w-1.5 h-1.5 rounded-full bg-emerald-400 mt-1.5 shrink-0" /><span className="min-w-0">{h}</span></li>)}</ul>
                ) : <p className="text-[11px] text-slate-400">当前简历与该岗位的直接匹配点较少。</p>}
              </div>
              <div className="bg-white border border-slate-200 rounded-2xl p-4 shadow-sm">
                <h4 className="text-[12px] font-semibold text-red-500 mb-2">最大风险</h4>
                {activeResult.keyRisks.length > 0 ? (
                  <ul className="space-y-2">{activeResult.keyRisks.slice(0, 3).map((r: any, i: number) => <li key={i} className="flex items-start gap-2 text-[12px] text-slate-700"><span className={`w-1.5 h-1.5 rounded-full mt-1.5 shrink-0 ${(r.risk_level || r.riskLevel) === "high" ? "bg-red-400" : "bg-amber-400"}`} /><span className="min-w-0">{(r.risk_name || r.riskName)}</span></li>)}</ul>
                ) : activeResult.structuredRisks.length > 0 ? (
                  <ul className="space-y-2">{activeResult.structuredRisks.slice(0, 3).map((r, i) => <li key={i} className="flex items-start gap-2 text-[12px] text-slate-700"><span className={`w-1.5 h-1.5 rounded-full mt-1.5 shrink-0 ${r.riskLevel === "high" ? "bg-red-400" : "bg-amber-400"}`} /><span className="min-w-0">{r.riskName}</span></li>)}</ul>
                ) : <p className="text-[11px] text-slate-400">未发现明显高风险，但仍建议结合 JD 优化表达。</p>}
              </div>
              <div className="bg-white border border-slate-200 rounded-2xl p-4 shadow-sm">
                <h4 className="text-[12px] font-semibold text-[#0052D9] mb-2">现在怎么改？</h4>
                {activeResult.actionSuggestions.length > 0 ? (
                  <ol className="space-y-2">{activeResult.actionSuggestions.slice(0, 3).map((a, i) => <li key={i} className="flex items-start gap-2 text-[12px] text-slate-700"><span className="text-[#0052D9] font-bold shrink-0">{i + 1}.</span><span className="min-w-0">{a}</span></li>)}</ol>
                ) : activeResult.nextActions.length > 0 ? (
                  <ol className="space-y-2">{activeResult.nextActions.slice(0, 3).map((a, i) => <li key={i} className="flex items-start gap-2 text-[12px] text-slate-700"><span className="text-[#0052D9] font-bold shrink-0">{i + 1}.</span><span className="min-w-0">{a}</span></li>)}</ol>
                ) : <p className="text-[11px] text-slate-400">建议优先补充与该岗位核心要求相关的项目证据，再进行投递。</p>}
              </div>
            </div>
            )}

            {/* ── Evidence Summary ───────────────────────────── */}
            {!activeResult.isProfilePending && activeResult.requirementCoverage.length > 0 && (
            <div className="grid grid-cols-2 gap-4">
              <div className="bg-white border border-slate-200 rounded-2xl p-4 shadow-sm">
                <h4 className="text-[12px] font-semibold text-slate-400 uppercase tracking-wide mb-2">JD 要求覆盖</h4>
                <div className="space-y-1.5">{activeResult.requirementCoverage.slice(0, 4).map((c: any, i: number) => <div key={i} className="flex items-center gap-2 text-[11px]"><span className={`text-[9px] px-1.5 py-0.5 rounded shrink-0 ${(c.requirement_type || c.requirementType) === "must_have" ? "bg-red-50 text-red-600" : "bg-slate-50 text-slate-500"}`}>{(c.requirement_type || c.requirementType) === "must_have" ? "必备" : (c.requirement_type || c.requirementType) === "should_have" ? "重要" : ""}</span><span className="flex-1 truncate text-slate-700">{(c.requirement_name || c.requirementName)}</span><span className={`shrink-0 font-mono text-[10px] ${(c.coverage_status || c.coverageStatus) === "strong" ? "text-emerald-600" : (c.coverage_status || c.coverageStatus) === "missing" ? "text-red-400" : "text-slate-400"}`}>{c.coverage_score || c.coverageScore || "-"}</span></div>)}</div>
              </div>
              <div className="bg-white border border-slate-200 rounded-2xl p-4 shadow-sm">
                <h4 className="text-[12px] font-semibold text-slate-400 uppercase tracking-wide mb-2">简历证据</h4>
                {activeResult.resumeEvidenceMapping.length > 0 ? (
                  <ul className="space-y-1.5">{activeResult.resumeEvidenceMapping.slice(0, 3).map((ev: any, i: number) => <li key={i} className="flex items-center gap-1.5 text-[11px] text-slate-600"><span className="text-[9px] px-1 py-0.5 bg-slate-100 text-slate-500 rounded shrink-0">L{ev.evidence_level || ev.evidenceLevel || "?"}</span><span className="truncate">{ev.evidence_name || ev.evidenceName || ev.source_title || ev.sourceTitle}</span></li>)}</ul>
                ) : activeResult.resumeEvidence && activeResult.resumeEvidence.length > 0 ? (
                  <ul className="space-y-1.5">{activeResult.resumeEvidence.slice(0, 3).map((e: any, i: number) => <li key={i} className="text-[11px] text-slate-600 truncate">• {typeof e === 'string' ? e : (e.evidence_name || e.evidenceName || '')}</li>)}</ul>
                ) : <p className="text-[11px] text-slate-400">无结构化证据。</p>}
              </div>
            </div>
            )}

            {/* ── Deep Summary ────────────────────────────────── */}
            {!isFast && !activeResult.isProfilePending && activeResult.scoreAlignment && (
            <div className="bg-white border border-indigo-200 rounded-2xl p-4 shadow-sm">
              <h4 className="text-[12px] font-semibold text-indigo-500 mb-3 inline-flex items-center gap-1.5"><span className="w-1.5 h-1.5 rounded-full bg-indigo-400" />Deep 补充判断</h4>
              <div className="flex items-center gap-3 text-[12px] mb-2">
                <span className="text-slate-400">Fast {activeResult.scoreAlignment.fastScore} → Deep {activeResult.scoreAlignment.deepScore}</span>
                <span className={`text-[10px] px-2 py-0.5 rounded-full ${activeResult.scoreAlignment.isConsistent ? "bg-emerald-50 text-emerald-600" : "bg-amber-50 text-amber-600"}`}>{activeResult.scoreAlignment.isConsistent ? "判断一致" : "存在差异"}</span>
              </div>
              {activeResult.scoreAlignment.deltaReason && <p className="text-[11px] text-slate-500">{activeResult.scoreAlignment.deltaReason}</p>}
            </div>
            )}

            {/* ── Collapsible Details ─────────────────────────── */}
            {!activeResult.isProfilePending && (
            <details className="bg-white border border-slate-200 rounded-2xl shadow-sm group">
              <summary className="px-4 py-3 cursor-pointer text-[12px] font-medium text-slate-500 hover:text-slate-700 select-none flex items-center gap-2">展开分析详情</summary>
              <div className="px-4 pb-4 space-y-4 border-t border-slate-100 pt-3">
                {activeResult.requirementCoverage.length > 0 && (
                <div><h5 className="text-[11px] font-semibold text-slate-400 uppercase tracking-wide mb-2">完整要求覆盖</h5>
                  <div className="space-y-1">{activeResult.requirementCoverage.map((c: any, i: number) => <div key={i} className="flex items-center gap-2 text-[11px] border-b border-slate-50 pb-1"><span className={`text-[9px] px-1.5 py-0.5 rounded shrink-0 ${(c.requirement_type || c.requirementType) === "must_have" ? "bg-red-50 text-red-600" : (c.requirement_type || c.requirementType) === "should_have" ? "bg-amber-50 text-amber-600" : "bg-slate-50 text-slate-500"}`}>{(c.requirement_type || c.requirementType) === "must_have" ? "必备" : (c.requirement_type || c.requirementType) === "should_have" ? "重要" : "加分"}</span><span className="flex-1 truncate text-slate-700">{(c.requirement_name || c.requirementName)}</span><span className="text-[10px] text-slate-400">{c.coverage_score || c.coverageScore || "-"}</span></div>)}</div>
                </div>
                )}
                {activeResult.resumeEvidenceMapping.length > 0 && (
                <div><h5 className="text-[11px] font-semibold text-slate-400 uppercase tracking-wide mb-2">完整简历证据</h5>
                  <div className="space-y-2">{activeResult.resumeEvidenceMapping.map((ev: any, i: number) => <div key={i} className="text-[11px] border-b border-slate-50 pb-1"><span className="text-[9px] px-1 py-0.5 bg-slate-100 text-slate-500 rounded mr-2">L{ev.evidence_level || ev.evidenceLevel || "?"}</span><span className="font-medium text-slate-700">{ev.evidence_name || ev.evidenceName}</span>{ev.evidence_text && <p className="text-slate-500 mt-0.5">{ev.evidence_text || ev.evidenceText}</p>}</div>)}</div>
                </div>
                )}
                {activeResult.jobUnderstanding && ((activeResult.jobUnderstanding as any).must_have_requirements?.length > 0) && (
                <div><h5 className="text-[11px] font-semibold text-slate-400 uppercase tracking-wide mb-2">JD 解构</h5>
                  <div className="flex flex-wrap gap-1">{(activeResult.jobUnderstanding as any).must_have_requirements.slice(0, 8).map((r: string, i: number) => <span key={i} className="text-[10px] px-2 py-0.5 bg-red-50 text-red-600 rounded">{r}</span>)}</div>
                </div>
                )}
                {activeResult.interviewRiskQuestions.length > 0 && (
                <div><h5 className="text-[11px] font-semibold text-slate-400 uppercase tracking-wide mb-2">面试准备参考</h5>
                  <div className="space-y-2">{activeResult.interviewRiskQuestions.slice(0, 5).map((q: any, i: number) => <div key={i} className="border border-purple-100 rounded-xl p-3 bg-purple-50/20"><div className="flex items-center gap-2 mb-1"><span className="w-4 h-4 rounded-full bg-purple-100 text-purple-700 text-[10px] font-bold flex items-center justify-center shrink-0">{i + 1}</span><span className="text-[12px] font-medium text-slate-800">{q.question || q.Question}</span></div>{(q.why_asked || q.whyAsked) && <p className="text-[10px] text-slate-400 ml-6">为什么问：{q.why_asked || q.whyAsked}</p>}{(q.answer_direction || q.answerDirection) && <p className="text-[10px] text-blue-600 ml-6 mt-0.5">回答方向：{q.answer_direction || q.answerDirection}</p>}</div>)}</div>
                </div>
                )}
                <details className="mt-2">
                  <summary className="text-[10px] text-slate-400 cursor-pointer hover:text-slate-600">技术调试信息</summary>
                  <div className="mt-2 space-y-1 text-[10px] text-slate-400 font-mono">
                    <p>generated_by: {activeResult.generatedBy || "—"}</p>
                    <p>report_version: {activeResult.reportVersion || "—"}</p>
                    {activeResult.meta && <p>agent_status: {(activeResult.meta as any)?.agentStatus || "—"}</p>}
                    {activeResult.qualityChecks && <p>quality_checks: {Object.entries(activeResult.qualityChecks).filter(([,v]) => !v).map(([k]) => k).join(", ") || "all ✅"}</p>}
                    {activeResult.agentTrace.length > 0 && <div className="mt-1">{activeResult.agentTrace.map((t: any, i: number) => <p key={i}>{(t.stage || t.Stage)}: {(t.status || t.Status)} — {(t.summary || t.Summary)}</p>)}</div>}
                  </div>
                </details>
              </div>
            </details>
            )}
          </section>

          {/* ═══ Right Sidebar ══════════════════════════════════ */}
          <aside className="space-y-4">
            <div className="bg-white border border-gray-200 rounded-2xl p-4">
              <h3 className="text-[11px] font-semibold text-gray-400 uppercase tracking-wide mb-2">当前岗位</h3>
              <p className="text-[13px] font-bold text-gray-900 leading-snug">{job?.title || "目标岗位"}</p>
              {job?.direction && <p className="text-[11px] text-slate-500 mt-0.5">{job.direction}{job.city ? ` · ${job.city}` : ""}</p>}
            </div>
            <div className="bg-white border border-gray-200 rounded-2xl p-4">
              <h3 className="text-[11px] font-semibold text-gray-400 uppercase tracking-wide mb-2">当前简历</h3>
              <p className="text-[13px] font-bold text-gray-900 leading-snug truncate">{selectedOption.name}</p>
              {activeResult?.profileDominantProfile && <p className="text-[11px] text-blue-600 mt-0.5">定位：{humanizeProfile(activeResult.profileDominantProfile)}</p>}
              {selectedOption.tags && selectedOption.tags.length > 0 && (
                <div className="flex flex-wrap gap-1 mt-2">{selectedOption.tags.slice(0, 4).map((t, i) => <span key={i} className="px-1.5 py-0.5 bg-slate-100 text-slate-600 text-[10px] rounded">{t}</span>)}</div>
              )}
              {formatDateTime(selectedOption.updatedAt) && <p className="text-[10px] text-gray-400 mt-2">更新于 {formatDateTime(selectedOption.updatedAt)}</p>}
            </div>
            <div className="bg-white border border-gray-200 rounded-2xl p-4">
              <h3 className="text-[11px] font-semibold text-gray-400 uppercase tracking-wide mb-2">操作</h3>
              {phase === "deep_loading" ? (
                <div className="flex items-center gap-2 text-xs text-gray-500"><Loader2 className="w-3.5 h-3.5 animate-spin" /> 智能体深度分析中…</div>
              ) : phase === "deep_done" ? (
                <div className="space-y-2">
                  <div className="text-xs text-green-700 flex items-center gap-1.5"><RefreshCw className="w-3 h-3" />深度分析已完成</div>
                  <button onClick={handleDeepAnalysis} className="w-full px-3 py-2 bg-slate-100 text-slate-700 text-[12px] rounded-lg hover:bg-slate-200">重新生成深度分析</button>
                </div>
              ) : phase === "deep_error" ? (
                <button onClick={handleDeepAnalysis} className="w-full px-3 py-2 bg-[#0052D9] text-white text-[12px] rounded-lg hover:bg-[#0047BD]">重试深度匹配分析</button>
              ) : (
                <button onClick={handleDeepAnalysis} className="w-full px-3 py-2 bg-[#0052D9] text-white text-[12px] rounded-lg hover:bg-[#0047BD] inline-flex items-center justify-center gap-1.5">
                  <Sparkles className="w-3.5 h-3.5" /> 生成深度分析
                </button>
              )}
              <Link to="/jobs" className="block mt-2 w-full text-center px-3 py-2 bg-slate-50 text-slate-600 text-[11px] rounded-lg hover:bg-slate-100 border border-slate-200">返回岗位列表</Link>
            </div>
          </aside>
        </div>
      )}
    </ContentPageShell>
  );
}
