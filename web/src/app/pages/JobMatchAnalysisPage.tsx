import React, { useState, useEffect } from "react";
import { Link, useSearchParams, useNavigate, useLocation } from "react-router";
import { loadMatchAnalysisViewModel, loadLatestMatchViewModel, loadResumeOptionViewModelsFromApi, loadResumeVersionDetail } from "../../api";
import { getJobDetail } from "../../api/jobs";
import type { JobApiItem } from "../../api/jobs";
import type { MatchAnalysisPageViewModel, ResumeOptionViewModel } from "../../adapters";
import type { Resume } from "../../types";
import { ContentPageShell } from "../components/layout/ContentPageShell";
import { getAuthRedirectIf401 } from "../utils/authRedirect";
import { humanizeProfile } from "../utils/profileLabel";
import { PageError, PageEmpty } from "../components/PageState";
import { AgentStepLoadingOverlay } from "../components/common/AgentStepLoadingOverlay";
import { MatchAnalysisSkeleton } from "../components/match/MatchAnalysisSkeleton";
import { Loader2, Sparkles, RefreshCw, AlertCircle, ShieldCheck, ShieldX, PenTool, Compass } from "lucide-react";
import { isMockGeneratedBy } from "../utils/agentError";

// ─── Helpers ──────────────────────────────────────────────────────────
function formatDateTime(v?: string | null): string {
  if (!v) return ""; const d = new Date(v); if (Number.isNaN(d.getTime())) return "";
  return `${d.getFullYear()}-${String(d.getMonth()+1).padStart(2,"0")}-${String(d.getDate()).padStart(2,"0")} ${String(d.getHours()).padStart(2,"0")}:${String(d.getMinutes()).padStart(2,"0")}`;
}

type MatchPhase = "loading_latest" | "fast_loading" | "fast_done" | "fast_error" | "deep_loading" | "deep_done" | "deep_error";
function parseResumeIdFromQuery(s: string | null): number | null { if (s && /^\d+$/.test(s)) return Number(s); return null; }

export function JobMatchAnalysisPage() {
  const [sp] = useSearchParams(); const nav = useNavigate(); const loc = useLocation();
  const jobId = sp.get("jobId") || ""; const resumeIdStr = sp.get("resumeId");
  const [job, setJob] = useState<JobApiItem | null>(null); const [loadingJob, setLoadingJob] = useState(!!jobId); const [jobNotFound, setJobNotFound] = useState(false);
  const [resumeOptions, setResumeOptions] = useState<ResumeOptionViewModel[]>([]); const [selectedOption, setSelectedOption] = useState<ResumeOptionViewModel | null>(null); const [loadingOptions, setLoadingOptions] = useState(true);
  const [selectedResumeDetail, setSelectedResumeDetail] = useState<Resume | null>(null); const [loadingResumeDetail, setLoadingResumeDetail] = useState(false);
  const [phase, setPhase] = useState<MatchPhase>("loading_latest"); const [activeResult, setActiveResult] = useState<MatchAnalysisPageViewModel | null>(null);
  const [activeTier, setActiveTier] = useState("fast"); const [activeGenBy, setActiveGenBy] = useState(""); const [matchError, setMatchError] = useState<string | null>(null); const [deepBackendDone, setDeepBackendDone] = useState(false);

  useEffect(() => { if (!jobId) { setLoadingJob(false); return; } getJobDetail(jobId).then(j => { if (j) setJob(j); else setJobNotFound(true); }).catch(() => setJobNotFound(true)).finally(() => setLoadingJob(false)); }, [jobId]);
  useEffect(() => { loadResumeOptionViewModelsFromApi().then(opts => { setResumeOptions(opts); if (!opts.length) { setLoadingOptions(false); return; } const tid = parseResumeIdFromQuery(resumeIdStr); if (tid) { const f = opts.find(o => o.id === tid); if (f) { setSelectedOption(f); nav(`/match?jobId=${jobId}&resumeId=${f.id}`, { replace: true }); } setLoadingOptions(false); } else { const cur = opts.find(o => o.isCurrent) ?? opts[0]; setSelectedOption(cur); nav(`/match?jobId=${jobId}&resumeId=${cur.id}`, { replace: true }); setLoadingOptions(false); } }).catch(err => { if (!getAuthRedirectIf401(err, loc.pathname, loc.search)) setLoadingOptions(false); }); }, []);
  useEffect(() => { if (!selectedOption) return; const id = selectedOption.id; setLoadingResumeDetail(true); loadResumeVersionDetail(id).then(d => { if (d) setSelectedResumeDetail(d); }).finally(() => setLoadingResumeDetail(false)); }, [selectedOption]);
  useEffect(() => { if (!jobId || !selectedOption || !selectedResumeDetail) return; const rid = selectedOption.id; setPhase("loading_latest"); loadLatestMatchViewModel(jobId, rid).then(l => { if (l.hasResult && l.result) { setActiveResult(l.result); setActiveTier(l.generationTier || "fast"); setActiveGenBy(l.generatedBy || ""); setPhase(l.generationTier === "quality" ? "deep_done" : "fast_done"); } else { setPhase("fast_loading"); loadMatchAnalysisViewModel(jobId, { resume_id: rid, generation_tier: "fast" }).then(r => { setActiveResult(r); setActiveTier("fast"); setActiveGenBy(r.generatedBy || ""); setPhase("fast_done"); }).catch(() => { setMatchError("快速分析失败"); setPhase("fast_error"); }); } }).catch(() => { setMatchError("读取分析结果失败"); setPhase("fast_error"); }); }, [jobId, selectedOption, selectedResumeDetail]);

  const handleDeep = () => { if (!jobId || !selectedOption) return; const rid = selectedOption.id; setPhase("deep_loading"); setDeepBackendDone(false); const abort = new AbortController(); loadMatchAnalysisViewModel(jobId, { resume_id: rid, generation_tier: "quality" }, abort.signal).then(r => { setActiveResult(r); setActiveTier("quality"); setActiveGenBy(r.generatedBy || ""); setDeepBackendDone(true); if (isMockGeneratedBy(r.generatedBy || "") || r.reportVersion !== "screening_match_v1") setPhase("deep_error"); }).catch(() => setPhase("deep_error")); const t = setTimeout(() => abort.abort(), 90000); abort.signal.addEventListener("abort", () => clearTimeout(t)); };

  if (!jobId) return <ContentPageShell maxWidthClassName="max-w-[1280px]"><PageError title="缺少目标岗位" description="请选择岗位后再进行匹配分析。"><Link to="/jobs" className="px-4 py-2 bg-[#0052D9] text-white text-sm rounded-lg">返回岗位列表</Link></PageError></ContentPageShell>;
  if (!loadingJob && jobNotFound) return <ContentPageShell maxWidthClassName="max-w-[1280px]"><PageError title="岗位不存在" description="请返回岗位列表重新选择。"><Link to="/jobs" className="px-4 py-2 bg-[#0052D9] text-white text-sm rounded-lg">返回岗位列表</Link></PageError></ContentPageShell>;
  if (loadingJob || loadingOptions) return <MatchAnalysisSkeleton message="正在加载…" />;
  if (!resumeOptions.length) return <ContentPageShell maxWidthClassName="max-w-[1280px]"><PageEmpty title="你还没有可用于匹配的简历" description="请先创建或导入一份简历。"><Link to="/resumes/new" className="px-4 py-2 bg-[#0052D9] text-white text-sm rounded-lg">去创建简历</Link></PageEmpty></ContentPageShell>;
  if (loadingResumeDetail || phase === "loading_latest" || phase === "fast_loading") return <MatchAnalysisSkeleton message="正在生成分析…" />;
  if (phase === "fast_error" && !activeResult) return <ContentPageShell maxWidthClassName="max-w-[1280px]"><PageError title="分析失败" description={matchError || "请稍后重试。"}><Link to={`/jobs/${jobId}`} className="text-[#0052D9] text-sm underline">返回岗位详情</Link></PageError></ContentPageShell>;

  const isFast = activeTier !== "quality";
  const sd = activeResult?.screeningDecision;
  const title = isFast ? "快速初筛判断" : "初筛命中率提升方案";

  return (
    <ContentPageShell maxWidthClassName="max-w-[1280px]" contentClassName="pb-12">
      <AgentStepLoadingOverlay open={phase === "deep_loading"} title="正在深度分析" description="Offer捕手正在生成初筛命中率提升方案。" steps={[{key:"profile",label:"读取简历定位"},{key:"jd",label:"分析岗位要求"},{key:"diagnosis",label:"初筛诊断"},{key:"rewrite",label:"生成改写方案"}]} backendDone={deepBackendDone} onFinish={() => setPhase("deep_done")} />

      <div className="mb-5 flex items-center justify-between">
        <div><h1 className="text-[22px] font-semibold text-gray-900 flex items-center gap-2"><Sparkles className="w-5 h-5 text-[#0052D9]" />{title}</h1><p className="text-xs text-gray-500 mt-1">{job?.title || ""} · 简历：{selectedResumeDetail?.real_name || selectedOption?.name}</p></div>
        <div className="flex gap-2">{isFast ? <button onClick={handleDeep} className="px-4 py-2 bg-[#0052D9] text-white text-[13px] rounded-xl hover:bg-[#0047BD] flex items-center gap-1.5"><PenTool className="w-3.5 h-3.5" />生成提升方案</button> : <button onClick={handleDeep} className="px-4 py-2 bg-slate-100 text-slate-700 text-[13px] rounded-xl hover:bg-slate-200 flex items-center gap-1.5"><RefreshCw className="w-3.5 h-3.5" />重新生成</button>}{!isFast && <button onClick={() => { setPhase("fast_loading"); loadMatchAnalysisViewModel(jobId, { resume_id: selectedOption!.id, generation_tier: "fast" }).then(r => { setActiveResult(r); setActiveTier("fast"); setPhase("fast_done"); }).catch(() => setPhase("fast_error")); }} className="px-4 py-2 bg-slate-100 text-slate-700 text-[13px] rounded-xl hover:bg-slate-200 flex items-center gap-1.5"><RefreshCw className="w-3.5 h-3.5" />快速重检</button>}</div>
      </div>

      {/* profile pending */}
      {activeResult?.isProfilePending && (
        <div className="bg-white border border-amber-200 rounded-3xl p-8 text-center"><div className="w-16 h-16 mx-auto mb-4 rounded-2xl bg-amber-50 flex items-center justify-center"><Loader2 className="w-7 h-7 text-amber-500 animate-spin" /></div><h3 className="text-lg font-semibold text-slate-800 mb-2">语义画像准备中</h3><p className="text-sm text-slate-500 max-w-md mx-auto">这不是简历能力判断，而是系统准备状态 — 画像就绪后即可生成初筛分析。</p><button onClick={handleDeep} className="mt-5 px-5 py-2.5 bg-[#0052D9] text-white text-sm rounded-xl">重新生成</button></div>
      )}

      {activeResult && !activeResult.isProfilePending && (
        <div className="grid grid-cols-1 xl:grid-cols-[minmax(0,1fr)_340px] gap-6 items-start">
          <section className="min-w-0 space-y-5">

            {/* A: Decision Hero */}
            <div className="bg-white border border-slate-200 rounded-3xl p-6 shadow-sm">
              <div className="flex items-start gap-5">
                <div className="w-[72px] h-[72px] rounded-2xl bg-gradient-to-br from-blue-500 to-cyan-400 flex items-center justify-center shrink-0 shadow-lg shadow-blue-500/20"><span className="text-[28px] font-bold text-white">{sd?.score ?? activeResult.matchScore}</span></div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-1">
                    <span className="text-xl font-bold text-slate-800">{sd?.level ?? activeResult.matchLevel}</span>
                    {activeResult.decisionLabel && <span className={`text-[11px] px-2.5 py-0.5 rounded-full font-medium ${(sd?.decision || "").includes("direct") ? "bg-emerald-50 text-emerald-700 border-emerald-200" : (sd?.decision || "").includes("revise") ? "bg-blue-50 text-blue-700 border-blue-200" : "bg-amber-50 text-amber-700 border-amber-200"}`}>{activeResult.decisionLabel}</span>}
                  </div>
                  <p className="text-sm text-slate-600 leading-relaxed">{activeResult.summary}</p>
                  {(sd?.decisionReasons || []).length > 0 && <div className="flex flex-wrap gap-1.5 mt-2">{sd!.decisionReasons.slice(0,3).map((r,i) => <span key={i} className="text-[11px] text-slate-500 bg-slate-50 px-2 py-0.5 rounded">{r}</span>)}</div>}
                </div>
              </div>
            </div>

            {!isFast && (<>
              {/* B: Pass Points */}
              <div className="bg-white border border-slate-200 rounded-3xl p-5 shadow-sm">
                <h3 className="text-[15px] font-semibold text-emerald-700 mb-1 flex items-center gap-2"><ShieldCheck className="w-4 h-4" />为什么可能过初筛</h3>
                <p className="text-[11px] text-slate-400 mb-4">系统只展示真正有助于 HR/ATS 初筛识别的简历证据。</p>
                {(activeResult.screeningPassPoints || []).length > 0 ? (
                  <div className="space-y-3">{activeResult.screeningPassPoints!.slice(0,3).map((p,i) => (
                    <div key={i} className="border border-emerald-100 rounded-xl p-4 bg-emerald-50/30">
                      <div className="flex items-center gap-2 mb-2"><span className="text-xs font-bold text-emerald-600 bg-emerald-100 px-2 py-0.5 rounded">命中点 {i+1}</span><span className="text-[13px] font-medium text-slate-800">{p.title}</span></div>
                      <p className="text-[12px] text-slate-600 leading-relaxed mb-2">{p.whyItHelps}</p>
                      {p.resumeEvidence && <div className="text-[11px] text-blue-600 bg-blue-50 px-3 py-2 rounded-lg border border-blue-100 font-mono">{p.resumeEvidence}</div>}
                      <p className="text-[11px] text-emerald-600 mt-2">{p.screeningValue}</p>
                    </div>))}</div>
                ) : <p className="text-[12px] text-slate-400">当前简历缺少能支撑该岗位初筛的明确证据。</p>}
              </div>

              {/* C: Elimination Risks */}
              <div className="bg-white border border-slate-200 rounded-3xl p-5 shadow-sm">
                <h3 className="text-[15px] font-semibold text-red-600 mb-1 flex items-center gap-2"><ShieldX className="w-4 h-4" />为什么可能被筛掉</h3>
                <p className="text-[11px] text-slate-400 mb-4">优先展示最影响初筛通过率的问题。</p>
                {(activeResult.screeningEliminationRisks || []).length > 0 ? (
                  <div className="space-y-3">{activeResult.screeningEliminationRisks!.slice(0,3).map((r,i) => (
                    <div key={i} className="border rounded-xl p-4" style={{ borderColor: r.severity==="high"?"#fca5a5":"#fdba74" }}>
                      <div className="flex items-center gap-2 mb-2"><span className={`text-[10px] px-2 py-0.5 rounded-full font-medium ${r.severity==="high"?"bg-red-50 text-red-600":"bg-amber-50 text-amber-600"}`}>{r.severity==="high"?"高风险":"中风险"}</span><span className="text-[13px] font-medium text-slate-800">{r.riskTitle}</span></div>
                      <p className="text-[12px] text-slate-600 mb-1">{r.whyItBlocksScreening}</p>
                      {r.currentResumeProblem && <p className="text-[11px] text-slate-400 mb-1">当前问题：{r.currentResumeProblem}</p>}
                      {r.fixDirection && <p className="text-[11px] text-blue-600">修复方向：{r.fixDirection}</p>}
                    </div>))}</div>
                ) : <p className="text-[12px] text-slate-400">未发现明显的初筛淘汰风险。</p>}
              </div>

              {/* D: Rewrite Commands */}
              <div className="bg-white border border-slate-200 rounded-3xl p-5 shadow-sm">
                <h3 className="text-[15px] font-semibold text-[#0052D9] mb-1 flex items-center gap-2"><PenTool className="w-4 h-4" />简历具体怎么改</h3>
                <p className="text-[11px] text-slate-400 mb-4">不是泛泛建议，而是可以直接照着改的简历命中率提升指令。</p>
                {(activeResult.resumeRewriteCommands || []).length > 0 ? (
                  <div className="space-y-5">{activeResult.resumeRewriteCommands!.slice(0,4).map((c,i) => (
                    <div key={i} className="border border-blue-100 rounded-2xl p-4 bg-blue-50/20">
                      <div className="flex items-center gap-2 mb-3"><span className="text-xs font-bold text-white bg-[#0052D9] w-6 h-6 rounded-full flex items-center justify-center">{i+1}</span><span className="text-[13px] font-medium text-slate-800">{c.targetSection}</span></div>
                      <p className="text-[12px] text-slate-500 mb-2"><strong>问题：</strong>{c.problem}</p>
                      <p className="text-[12px] text-slate-700 mb-3"><strong>改写方向：</strong>{c.rewriteInstruction}</p>
                      <div className="grid grid-cols-1 lg:grid-cols-2 gap-3 mb-3">
                        <div className="bg-red-50/50 border border-red-100 rounded-lg p-3"><p className="text-[10px] text-red-500 font-medium mb-0.5">修改前</p><p className="text-[11px] text-slate-600 leading-relaxed">{c.beforeExample || "—"}</p></div>
                        <div className="bg-emerald-50/50 border border-emerald-100 rounded-lg p-3"><p className="text-[10px] text-emerald-600 font-medium mb-0.5">修改后</p><p className="text-[11px] text-slate-700 leading-relaxed">{c.afterExample || "—"}</p></div>
                      </div>
                      <p className="text-[11px] text-blue-700 bg-blue-50 px-3 py-1.5 rounded-lg">💡 {c.whyItImprovesScreening}</p>
                    </div>))}</div>
                ) : <p className="text-[12px] text-slate-400">暂未生成可执行的简历改写指令，请重新生成分析。</p>}
              </div>

              {/* E: Job Strategy */}
              {activeResult.jobStrategy && (
                <div className="bg-white border border-slate-200 rounded-3xl p-5 shadow-sm">
                  <h3 className="text-[15px] font-semibold text-indigo-600 mb-1 flex items-center gap-2"><Compass className="w-4 h-4" />下一步投递策略</h3>
                  <div className="flex items-center gap-2 mt-3 mb-3"><span className={`text-xs px-3 py-1 rounded-full font-medium ${activeResult.jobStrategy.positioning.includes("主投")?"bg-emerald-50 text-emerald-700":activeResult.jobStrategy.positioning.includes("备选")?"bg-amber-50 text-amber-700":"bg-red-50 text-red-600"}`}>{activeResult.jobStrategy.positioning}</span></div>
                  <p className="text-[13px] text-slate-700 mb-2">{activeResult.jobStrategy.recommendedAction}</p>
                  {activeResult.jobStrategy.ifApplyNow && <p className="text-[11px] text-slate-400 mb-3">如果现在直接投：{activeResult.jobStrategy.ifApplyNow}</p>}
                  {activeResult.jobStrategy.betterFitDirections.length > 0 && <div className="mb-2"><span className="text-[11px] text-slate-500">更适合方向：</span><div className="flex flex-wrap gap-1 mt-1">{activeResult.jobStrategy.betterFitDirections.map((d,i) => <span key={i} className="text-[11px] px-2 py-0.5 bg-blue-50 text-blue-600 rounded-full">{d}</span>)}</div></div>}
                  {activeResult.jobStrategy.avoidDirections.length > 0 && <div><span className="text-[11px] text-slate-500">不建议优先投：</span><div className="flex flex-wrap gap-1 mt-1">{activeResult.jobStrategy.avoidDirections.map((d,i) => <span key={i} className="text-[11px] px-2 py-0.5 bg-slate-100 text-slate-500 rounded-full">{d}</span>)}</div></div>}
                </div>
              )}
            </>)}

            {/* Fast mode: just the hero, nothing else */}
          </section>

          {/* Right sidebar */}
          <aside className="space-y-4 xl:sticky xl:top-6 self-start">
            <div className="bg-white border border-gray-200 rounded-2xl p-4"><h3 className="text-[11px] font-semibold text-gray-400 uppercase tracking-wide mb-2">当前岗位</h3><p className="text-[13px] font-bold text-gray-900">{job?.title || "—"}</p>{job?.direction && <p className="text-[11px] text-slate-500 mt-0.5">{job.direction}{job.city ? ` · ${job.city}` : ""}</p>}</div>
            <div className="bg-white border border-gray-200 rounded-2xl p-4"><h3 className="text-[11px] font-semibold text-gray-400 uppercase tracking-wide mb-2">当前简历</h3><p className="text-[13px] font-bold text-gray-900 truncate">{selectedOption?.name}</p>{activeResult?.profileDominantProfile && <p className="text-[11px] text-blue-600 mt-0.5">定位：{humanizeProfile(activeResult.profileDominantProfile)}</p>}{formatDateTime(selectedOption?.updatedAt) && <p className="text-[10px] text-gray-400 mt-2">更新于 {formatDateTime(selectedOption?.updatedAt)}</p>}</div>
            <div className="bg-white border border-gray-200 rounded-2xl p-4">{isFast ? <button onClick={handleDeep} className="w-full px-3 py-2 bg-[#0052D9] text-white text-[12px] rounded-xl hover:bg-[#0047BD] flex items-center justify-center gap-1.5"><PenTool className="w-3.5 h-3.5" />生成初筛提升方案</button> : <button onClick={handleDeep} className="w-full px-3 py-2 bg-[#0052D9] text-white text-[12px] rounded-xl hover:bg-[#0047BD] flex items-center justify-center gap-1.5"><RefreshCw className="w-3.5 h-3.5" />重新生成分析</button>}<Link to="/jobs" className="block mt-2 w-full text-center px-3 py-2 bg-slate-50 text-slate-600 text-[11px] rounded-lg hover:bg-slate-100 border border-slate-200">返回岗位列表</Link></div>
          </aside>
        </div>
      )}
    </ContentPageShell>
  );
}
