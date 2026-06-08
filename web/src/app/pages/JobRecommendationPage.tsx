import React, { useState, useEffect, useRef } from "react";
import { Link, useSearchParams, useNavigate, useLocation } from "react-router";
import { Sparkles, Loader2, AlertCircle, RefreshCw } from "lucide-react";
import { loadRecommendationViewModel, loadResumeOptionViewModelsFromApi, loadLatestRecommendationViewModel } from "../../api";
import type { RecommendationPageViewModel, ResumeOptionViewModel, RecommendedJobViewModel, RecommendationPriorityTone } from "../../adapters";
import { PageLoading, PageEmpty } from "../components/PageState";
import { ContentPageShell } from "../components/layout/ContentPageShell";
import { AgentStepLoadingOverlay, type AgentLoadingStep } from "../components/common/AgentStepLoadingOverlay";
import { getAuthRedirectIf401 } from "../utils/authRedirect";
import { Badge } from "../components/ui/Badge";

// ─── Helpers ──────────────────────────────────────────────────────────────────

function parseResumeIdFromQuery(resumeId: string | null): number | null {
  if (resumeId && /^\d+$/.test(resumeId)) return Number(resumeId);
  return null;
}

type RecPhase =
  | "loading_latest"
  | "fast_loading" | "fast_done" | "fast_error"
  | "deep_loading" | "deep_profile_pending" | "deep_done" | "deep_error" | "deep_timeout"
  | "profile_failed";

/** Current display tier */
type DisplayTier = "fast" | "quality";

function priorityBadgeVariant(tone: RecommendationPriorityTone): "green" | "blue" | "amber" | "secondary" | "danger" {
  switch (tone) { case "strong": return "blue"; case "high": return "secondary"; case "try": return "amber"; case "backup": return "secondary"; case "not_recommended": return "danger"; }
}

// ─── Main Component ───────────────────────────────────────────────────────────

export function JobRecommendationPage() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const location = useLocation();
  const resumeId = searchParams.get("resumeId");

  const [resumeOptions, setResumeOptions] = useState<ResumeOptionViewModel[]>([]);
  const [selectedOption, setSelectedOption] = useState<ResumeOptionViewModel | null>(null);
  const [loadingOptions, setLoadingOptions] = useState(true);

  const [phase, setPhase] = useState<RecPhase>("loading_latest");
  const [fastResult, setFastResult] = useState<RecommendationPageViewModel | null>(null);
  const [deepResult, setDeepResult] = useState<RecommendationPageViewModel | null>(null);
  const [errorMessage, setErrorMessage] = useState("");
  const [deepErrorMessage, setDeepErrorMessage] = useState("");

  // K9-1-U-D: track which tier is currently displayed
  const [activeTier, setActiveTier] = useState<DisplayTier>("fast");
  const [activeGeneratedBy, setActiveGeneratedBy] = useState<string>("");

  const [selJobId, setSelJobId] = useState<number | null>(null);
  const [deepOverlayStatus, setDeepOverlayStatus] = useState<"loading" | "success" | "error">("loading");

  const deepSteps: AgentLoadingStep[] = [
    { key: "resume-profile", label: "读取简历画像标签" },
    { key: "job-tags", label: "匹配岗位标签矩阵" },
    { key: "reasoning", label: "生成推荐理由和风险点" },
    { key: "judge", label: "审查推荐结果一致性" },
  ];

  const fastPollRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const qualityPollRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const fastAttemptsRef = useRef(0);
  const qualityAttemptsRef = useRef(0);
  const MAX_FAST_POLL = 8;
  const MAX_QUALITY_POLL = 10;

  useEffect(() => { return () => { if (fastPollRef.current) clearTimeout(fastPollRef.current); if (qualityPollRef.current) clearTimeout(qualityPollRef.current); }; }, []);

  // ── Load resume options ──────────────────────────────────────────────────
  useEffect(() => {
    loadResumeOptionViewModelsFromApi()
      .then((opts) => {
        setResumeOptions(opts);
        if (opts.length === 0) { setSelectedOption(null); setLoadingOptions(false); return; }
        const target = parseResumeIdFromQuery(resumeId);
        if (target) { const found = opts.find((o) => o.id === target); setSelectedOption(found ?? opts[0]); }
        else { const cur = opts.find((o) => o.isCurrent); setSelectedOption(cur ?? opts[0]); }
        setLoadingOptions(false);
      })
      .catch((err) => { const r = getAuthRedirectIf401(err, location.pathname, location.search); if (r) { navigate(r, { replace: true }); return; } setLoadingOptions(false); });
  }, []);

  // ── K9-1-U-D: Entry flow — try latest first, generate Fast if none ────────
  useEffect(() => {
    if (!selectedOption) return;
    const rid = selectedOption.id;
    if (!rid || isNaN(rid)) return;
    setPhase("loading_latest"); setErrorMessage("");

    loadLatestRecommendationViewModel(rid)
      .then((latest) => {
        if (latest.hasResult && latest.result && latest.result.recommendedJobs.length > 0) {
          const tier = (latest.generationTier === "quality") ? "quality" : "fast";
          setActiveTier(tier);
          setActiveGeneratedBy(latest.generatedBy ?? "");
          if (tier === "quality") {
            setDeepResult(latest.result);
            setPhase("deep_done");
          } else {
            setFastResult(latest.result);
            setPhase("fast_done");
          }
          setSelJobId(latest.result.recommendedJobs[0].jobId);
          return;
        }
        // No saved result → generate Fast
        generateFast(rid);
      })
      .catch(() => {
        // Latest API failed → fallback to generating Fast
        generateFast(rid);
      });
  }, [selectedOption]);

  function generateFast(rid: number) {
    setPhase("fast_loading"); fastAttemptsRef.current = 0;
    const pollFast = () => {
      loadRecommendationViewModel({ resume_id: rid, limit: 5, generation_tier: "fast" as any })
        .then((vm) => {
          if (vm.profileStatus === "ready") { setFastResult(vm); setActiveTier("fast"); setPhase("fast_done"); if (vm.recommendedJobs.length > 0) setSelJobId(vm.recommendedJobs[0].jobId); return; }
          if (vm.profileStatus === "failed") { setPhase("profile_failed"); setErrorMessage(vm.profileMessage || "简历画像生成失败，请稍后重试。"); return; }
          fastAttemptsRef.current++;
          if (fastAttemptsRef.current >= MAX_FAST_POLL) { setPhase("fast_error"); setErrorMessage("画像生成时间较长，可稍后重试。"); return; }
          fastPollRef.current = setTimeout(pollFast, 1800);
        })
        .catch(() => { setPhase("fast_error"); setErrorMessage("推荐结果加载失败，请稍后重试。"); });
    };
    pollFast();
  }

  // ── Deep recommendation ──────────────────────────────────────────────────
  async function handleDeepRecommendation() {
    if (!selectedOption) return;
    setPhase("deep_loading"); setDeepErrorMessage(""); setDeepOverlayStatus("loading"); qualityAttemptsRef.current = 0;

    const pollQuality = async () => {
      try {
        const vm = await loadRecommendationViewModel({ resume_id: selectedOption.id, limit: 5, generation_tier: "quality" as any });
        if (vm.profileStatus === "ready" && vm.recommendedJobs.length > 0) {
          setDeepOverlayStatus("success"); setDeepResult(vm); setActiveTier("quality");
          setActiveGeneratedBy(vm.meta?.generated_by ?? "");
          if (vm.recommendedJobs.length > 0) setSelJobId(vm.recommendedJobs[0].jobId);
          return;
        }
        if (vm.profileStatus === "pending") { setPhase("deep_profile_pending"); return; }
        if (vm.profileStatus === "failed") { setDeepOverlayStatus("error"); setPhase("deep_error"); setDeepErrorMessage(vm.profileMessage || "深度画像生成失败，已保留当前结果。"); return; }
        qualityAttemptsRef.current++;
        if (qualityAttemptsRef.current >= MAX_QUALITY_POLL) { setPhase("deep_timeout"); return; }
        qualityPollRef.current = setTimeout(pollQuality, 2000);
      } catch { setDeepOverlayStatus("error"); setPhase("deep_error"); setDeepErrorMessage("深度分析暂时生成失败，当前结果仍可参考。"); }
    };
    if (qualityPollRef.current) clearTimeout(qualityPollRef.current);
    pollQuality();
  }

  const activeResult = phase === "deep_done" && deepResult ? deepResult : fastResult;
  const jobs = activeResult?.recommendedJobs ?? [];
  const selectedJob = selJobId ? jobs.find((j) => j.jobId === selJobId) ?? jobs[0] : jobs[0];

  // ── Loading states ───────────────────────────────────────────────────────
  if (loadingOptions) return <PageLoading message="正在加载简历信息…" />;
  if (resumeOptions.length === 0 || !selectedOption) {
    return (<PageEmpty title="你还没有可用于推荐的简历" description="请先创建一份简历，系统会基于你的简历生成岗位推荐。"><Link to="/resumes/new" className="text-[#0052D9] text-sm underline">去添加简历</Link></PageEmpty>);
  }
  if (phase === "loading_latest") return <PageLoading message="正在读取推荐结果…" />;
  if (phase === "fast_loading") {
    return (<div className="max-w-[1440px] mx-auto px-6 py-32 text-center"><Loader2 className="w-10 h-10 animate-spin text-[#0052D9] mx-auto mb-4" /><h2 className="text-lg font-bold text-gray-900 mb-2">正在生成快速推荐</h2><p className="text-sm text-gray-500 max-w-md mx-auto">系统正在提取你的教育、项目、实习经历和技能标签进行快速岗位推荐。</p></div>);
  }
  if (phase === "fast_error" || phase === "profile_failed") {
    return (<div className="max-w-[1440px] mx-auto px-6 py-12 text-center"><AlertCircle className="w-8 h-8 text-amber-500 mx-auto mb-3" /><p className="text-gray-500 text-sm mb-3">{errorMessage}</p><button onClick={() => setSelectedOption({ ...selectedOption })} className="text-[#0052D9] text-sm underline">重新尝试</button></div>);
  }
  if (!activeResult || jobs.length === 0) {
    return (<div className="max-w-[1440px] mx-auto px-6 py-12 text-center"><p className="text-gray-500 text-sm">暂无推荐结果。</p></div>);
  }

  const displayName = selectedOption.name;
  const nowStr = new Date().toLocaleString("zh-CN");

  // K9-1-U-D: header label from phase + activeTier
  function getHeaderLabel(): string {
    switch (phase) {
      case "deep_loading": return "智能体正在深度分析";
      case "deep_profile_pending": return "正在生成深度画像";
      case "deep_done": return activeGeneratedBy?.startsWith("llm") ? "深度推荐 · AI生成" : "深度推荐 · 规则兜底";
      case "deep_timeout": return "深度画像生成较慢 · 已保留当前结果";
      case "deep_error": return "深度推荐生成失败 · 已保留当前结果";
      default: return "快速推荐 · 规则推荐";
    }
  }
  const headerLabel = getHeaderLabel();
  const isDeepDisplay = phase === "deep_done";

  return (
    <ContentPageShell maxWidthClassName="max-w-[1440px]" contentClassName="pb-12">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-[22px] font-semibold text-gray-900 flex items-center gap-2"><Sparkles className="w-5 h-5 text-[#0052D9]" />Offerr鹅岗位推荐</h1>
        <p className="mt-1.5 text-sm text-gray-500">基于简历 <span className="font-medium text-gray-700">{displayName}</span>（ID:{selectedOption.id}），从岗位池中推荐 Top 5 岗位。</p>
        <div className="flex flex-wrap items-center gap-3 mt-3 text-xs text-gray-500">
          <span className="flex items-center gap-1.5 text-[#0052D9] font-medium">{headerLabel}</span>
          <span className="text-gray-300">|</span><span>生成时间：{nowStr}</span>
          {activeResult.profileDominantProfile && (<><span className="text-gray-300">|</span><span className="text-[#0052D9]">画像：{activeResult.profileDominantProfile}</span></>)}
        </div>
        {/* Fast mode hint */}
        {!isDeepDisplay && (<p className="mt-2 text-[11px] text-gray-400">这是快速推荐结果，点击"生成深度推荐"可获取更完整的岗位匹配解释。</p>)}
      </div>

      <div className="flex gap-6 items-start">
        {/* ── Left cards ────────────────────────────────────────────────── */}
        <div className="w-[380px] shrink-0 flex flex-col gap-3">
          {jobs.map((job) => {
            const isSel = selectedJob && job.jobId === selectedJob.jobId;
            const badgeVar = priorityBadgeVariant(job.priorityTone);
            const evidence = isDeepDisplay ? job.resumeEvidence.slice(0, 2) : job.resumeEvidence.slice(0, 1);
            return (
              <div key={job.jobId} onClick={() => setSelJobId(job.jobId)}
                className={`relative p-4 bg-white border rounded-lg cursor-pointer transition-all duration-200 hover:shadow-md ${isSel ? "border-[#0052D9] bg-[#F4F8FF]" : "border-gray-200 hover:border-[#0052D9]"}`}>
                {isSel && <div className="absolute left-0 top-0 bottom-0 w-1 bg-[#0052D9] rounded-l-lg" />}
                <div className="flex items-center gap-2 mb-1">
                  <Badge variant={badgeVar} className="text-[11px] font-medium px-2">{job.priorityText}</Badge>
                  <span className={`text-[11px] text-gray-400 ${isSel ? "text-[#0052D9]" : ""}`}>{job.score}分</span>
                </div>
                <h3 className={`font-bold text-[14px] truncate mb-1 ${isSel ? "text-[#0052D9]" : "text-gray-900"}`}><span className="text-xs text-gray-400 mr-1">#{job.rank}</span>{job.title}</h3>
                <div className="flex flex-wrap items-center gap-1.5 mb-2 text-[11px] text-gray-500"><span>{job.city}</span><span className="text-gray-300">·</span><span>{job.category}</span><span className="text-gray-300">·</span><span>{job.department}</span></div>
                {(job.whyRankedHere || job.oneSentenceReason) && (<p className="text-[12px] text-gray-600 line-clamp-2 leading-relaxed mb-2">{job.whyRankedHere || job.oneSentenceReason}</p>)}
                {evidence.length > 0 && (<div className="flex flex-wrap gap-1 mb-1.5">{evidence.map((e, i) => (<Badge key={i} variant="outline" className="text-[10px] font-normal px-1.5 py-0 bg-blue-50/30 text-blue-700 border-blue-100">{e}</Badge>))}</div>)}
                {isDeepDisplay && job.jobSpecificRisks[0] && (<p className="text-[10px] text-amber-600 line-clamp-1">{job.jobSpecificRisks[0]}</p>)}
              </div>
            );
          })}
        </div>

        {/* ── Center detail ─────────────────────────────────────────────── */}
        <div className="flex-1 min-w-0 max-w-[680px] bg-white border border-gray-200 rounded-lg shadow-sm overflow-hidden">
          {selectedJob && (<>
            <div className="bg-[#F4F8FF] p-5 border-b border-blue-100">
              <div className="text-[12px] text-gray-400 mb-1">当前选中岗位</div>
              <h2 className="text-lg font-bold text-gray-900 mb-3">{selectedJob.title}</h2>
              <div className="flex flex-wrap items-center gap-2 mb-3 text-xs text-gray-500"><span>{selectedJob.city}</span><span className="text-gray-300">·</span><span>{selectedJob.category}</span><span className="text-gray-300">·</span><span>{selectedJob.department}</span></div>
              <div className="bg-white rounded-md p-3 border border-blue-50"><span className="text-xs font-semibold text-gray-900">一句话总结：</span><span className="text-sm text-gray-700">{selectedJob.whyRankedHere || selectedJob.oneSentenceReason}</span></div>
            </div>
            <div className="p-5 space-y-5">
              {/* Deep: full explanation */}
              {isDeepDisplay && selectedJob.jobSpecificReason && (<section><h3 className="font-bold text-sm text-gray-900 mb-2 flex items-center gap-1.5"><span className="w-1 h-3.5 bg-[#0052D9] rounded-full" />为什么推荐这个岗位</h3><p className="text-[13px] text-gray-700 leading-relaxed">{selectedJob.jobSpecificReason}</p></section>)}
              {isDeepDisplay && selectedJob.resumeEvidence.length > 0 && (<section><h3 className="font-bold text-sm text-gray-900 mb-2 flex items-center gap-1.5"><span className="w-1 h-3.5 bg-green-500 rounded-full" />简历匹配证据</h3><div className="flex flex-wrap gap-1.5">{selectedJob.resumeEvidence.map((e, i) => (<Badge key={i} variant="outline" className="text-[11px] font-normal px-2 py-0.5 bg-green-50/50 text-green-800 border-green-200">{e}</Badge>))}</div></section>)}
              {/* Fast: simplified highlights */}
              {!isDeepDisplay && selectedJob.matchHighlights.length > 0 && (<section><h3 className="font-bold text-sm text-gray-900 mb-2 flex items-center gap-1.5"><span className="w-1 h-3.5 bg-blue-400 rounded-full" />核心匹配点</h3><ul className="space-y-1.5">{selectedJob.matchHighlights.slice(0, 2).map((h, i) => (<li key={i} className="text-[13px] text-gray-700 flex items-start gap-2"><span className="text-green-500 mt-1 shrink-0">●</span> {h}</li>))}</ul></section>)}
              {/* Risks */}
              {selectedJob.jobSpecificRisks.length > 0 && (<section><h3 className="font-bold text-sm text-gray-900 mb-2 flex items-center gap-1.5"><span className="w-1 h-3.5 bg-amber-400 rounded-full" />主要风险</h3><div className="bg-amber-50/30 border border-amber-100/50 rounded-md p-3"><ul className="space-y-1.5">{(isDeepDisplay ? selectedJob.jobSpecificRisks : selectedJob.jobSpecificRisks.slice(0, 2)).map((r, i) => (<li key={i} className="flex items-center gap-2 text-[13px] text-amber-700"><div className="w-1.5 h-1.5 rounded-full bg-amber-400/80 shrink-0" /> {r}</li>))}</ul></div></section>)}
              {/* Deep: improvement */}
              {isDeepDisplay && selectedJob.suggestedImprovement && (<section><h3 className="font-bold text-sm text-gray-900 mb-2 flex items-center gap-1.5"><span className="w-1 h-3.5 bg-[#0052D9] rounded-full" />补强建议</h3><div className="bg-blue-50/30 border border-blue-100/50 rounded-md p-3"><p className="text-[13px] text-blue-800 leading-relaxed">{selectedJob.suggestedImprovement}</p></div></section>)}
              {/* Action */}
              <section className="bg-gray-50 rounded-md p-4 border border-gray-100">
                <div className="text-[13px] text-gray-700 leading-relaxed mb-3"><span className="font-medium text-gray-900">下一步建议：</span>{selectedJob.suggestedAction}</div>
                <div className="flex gap-3">
                  <Link to={`/jobs/${selectedJob.jobId}`} className="flex-1 text-center px-3 py-2 bg-[#0052D9] text-white text-xs rounded-lg hover:bg-[#0047BD]">进入岗位详情</Link>
                  <Link to={`/match?jobId=${selectedJob.jobId}&resumeId=${selectedOption.id}`} className="flex-1 text-center px-3 py-2 border border-[#0052D9] text-[#0052D9] text-xs rounded-lg hover:bg-blue-50">进行匹配分析</Link>
                </div>
              </section>
            </div>
          </>)}
        </div>

        {/* ── Right sidebar ─────────────────────────────────────────────── */}
        <aside className="w-[300px] shrink-0 flex flex-col gap-4">
          <div className="bg-white border border-gray-200 rounded-lg p-4">
            <h3 className="text-[13px] font-medium text-gray-500 mb-2">当前简历</h3>
            <p className="text-sm font-bold text-gray-900 mb-1">{selectedOption.name}</p>
            <p className="text-[11px] text-gray-400">Resume ID: {selectedOption.id}</p>
            {activeResult.profileDominantProfile && (<p className="text-[11px] text-[#0052D9] mt-1">画像：{activeResult.profileDominantProfile}</p>)}
          </div>

          <div className="bg-white border border-gray-200 rounded-lg p-4">
            <h3 className="text-[13px] font-medium text-gray-500 mb-2">深度推荐</h3>
            <p className="text-[11px] text-gray-400 mb-3">基于当前简历画像和岗位标签生成更完整的匹配解释。</p>
            {phase === "deep_loading" || phase === "deep_profile_pending" ? (
              <div className="flex items-center gap-2 text-xs text-gray-500"><Loader2 className="w-3.5 h-3.5 animate-spin" /> {phase === "deep_profile_pending" ? "正在生成深度画像" : "智能体深度分析中…"}</div>
            ) : phase === "deep_done" ? (
              <div className="text-xs text-green-700 flex items-center gap-1.5"><RefreshCw className="w-3 h-3" /> 深度分析已完成<button onClick={handleDeepRecommendation} className="ml-1 text-[#0052D9] hover:underline">重新生成</button></div>
            ) : phase === "deep_error" || phase === "deep_timeout" ? (
              <button onClick={handleDeepRecommendation} className="w-full flex items-center justify-center gap-1.5 px-3 py-2 bg-[#0052D9] text-white text-sm rounded-lg hover:bg-[#0047BD]"><RefreshCw className="w-3.5 h-3.5" /> 重试深度推荐</button>
            ) : (
              <button onClick={handleDeepRecommendation} className="w-full flex items-center justify-center gap-1.5 px-3 py-2 bg-[#0052D9] text-white text-sm rounded-lg hover:bg-[#0047BD]"><Sparkles className="w-3.5 h-3.5" /> 生成深度推荐</button>
            )}
          </div>
        </aside>
      </div>

      <AgentStepLoadingOverlay
        open={phase === "deep_loading" || deepOverlayStatus === "success"}
        title="智能体正在深度分析岗位匹配"
        description="系统正在结合你的候选人标签、岗位标签和匹配证据，生成更高质量的推荐结果。"
        steps={deepSteps}
        status={deepOverlayStatus === "success" ? "success" : phase === "deep_error" ? "error" : "loading"}
        errorText={deepErrorMessage} successDelayMs={400}
        onClose={() => { if (deepOverlayStatus === "success") { setPhase("deep_done"); setDeepOverlayStatus("loading"); } }}
      />
    </ContentPageShell>
  );
}
