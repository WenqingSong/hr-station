import React, { useState, useEffect, useMemo, useRef } from "react";
import { Link, useSearchParams, useNavigate, useLocation } from "react-router";
import { Sparkles, Loader2, AlertCircle, RefreshCw, CheckCircle2 } from "lucide-react";
import { loadRecommendationViewModel, loadResumeOptionViewModelsFromApi, loadLatestRecommendationViewModel } from "../../api";
import type { RecommendationPageViewModel, ResumeOptionViewModel, RecommendedJobViewModel, RecommendationPriorityTone } from "../../adapters";
import { PageLoading, PageEmpty, PageError } from "../components/PageState";
import { ContentPageShell } from "../components/layout/ContentPageShell";
import { DeepAgentLoadingOverlay } from "../components/common/DeepAgentLoadingOverlay";
import AgentErrorState from "../components/common/AgentErrorState";
import { BulletList } from "../components/common/BulletList";
import { getAuthRedirectIf401 } from "../utils/authRedirect";
import { humanizeProfile } from "../utils/profileLabel";
import { Badge } from "../components/ui/Badge";
import { isMockGeneratedBy, ContractError } from "../utils/agentError";

// ─── Helpers ──────────────────────────────────────────────────────────────────

function parseResumeIdFromQuery(resumeId: string | null): number | null {
  if (resumeId && /^\d+$/.test(resumeId)) return Number(resumeId);
  return null;
}

type RecPhase =
  | "loading_latest"
  | "fast_loading" | "fast_done" | "fast_error"
  | "deep_loading" | "deep_profile_pending" | "deep_done" | "deep_error" | "deep_timeout" | "deep_unavailable"
  | "profile_failed";

/** M4: error classification */
type DeepErrorType = "LLM_UNAVAILABLE" | "GENERATION_FAILED" | "NETWORK_ERROR" | "GATEWAY_TIMEOUT" | "UNKNOWN";

function classifyDeepError(err: any): DeepErrorType {
  const msg = JSON.stringify(err).toLowerCase();
  // K15-P0-D6: handle 504 Gateway Timeout specifically
  if (/504|gateway timeout|gateway-timeout/i.test(msg)) return "GATEWAY_TIMEOUT";
  if (/503|模型未启用|大模型服务|服务未配置|深度推荐需要|llm unavailable|not available/i.test(msg)) return "LLM_UNAVAILABLE";
  if (/401|unauthorized|token/i.test(msg)) return "UNKNOWN";
  if (/timeout|fetch failed|network/i.test(msg)) return "NETWORK_ERROR";
  return "GENERATION_FAILED";
}

/** M4: human-readable generated_by label */
function sourceLabel(genBy?: string, agentStatus?: string): string {
  if (!genBy) return "";
  if (genBy === "fast_tag_engine") return "快速推荐引擎";
  if (genBy === "llm_agent") return "深度推荐智能体";
  if (genBy === "deep_fallback_agent") return agentStatus === "fallback" ? "深度推荐（兜底）" : "深度推荐基础模型";
  if (genBy === "mock_quality") return "深度推荐基础模型";
  if (genBy.includes("mock")) return ""; // never show mock
  return genBy;
}

function agentStatusLabel(s?: string): string {
  if (!s) return "";
  if (s === "cascade_verified") return "已完成候选发现、语义重排、逐岗分析和全局校准";
  if (s === "cascade_repaired") return "系统已自动修复少量格式问题";
  return "";
}

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
  // J2: tracks incomplete Deep history separately from transient errors
  const [deepHistoryIncomplete, setDeepHistoryIncomplete] = useState(false);
  const [deepHistoryIncompleteCount, setDeepHistoryIncompleteCount] = useState(0);
  // L2: guards empty state from flashing before latest load completes
  const [isLatestLoading, setIsLatestLoading] = useState(false);

  // K9-1-U-D: track which tier is currently displayed
  const [activeTier, setActiveTier] = useState<DisplayTier>("fast");
  const [activeGeneratedBy, setActiveGeneratedBy] = useState<string>("");

  const [selJobId, setSelJobId] = useState<number | null>(null);
  const [deepBackendDone, setDeepBackendDone] = useState(false);

  // H17-D: Deep run guard — prevents stale requests from overwriting user choice
  const deepRunSeqRef = useRef(0);
  const cancelledDeepRunRef = useRef(0);
  const [deepReadyInBackground, setDeepReadyInBackground] = useState(false);
  // H17-D: phaseRef syncs latest phase for async closures
  const phaseRef = useRef<RecPhase>(phase);
  useEffect(() => { phaseRef.current = phase; }, [phase]);

  // L3: elapsed seconds timer for deep loading
  const [deepElapsedSec, setDeepElapsedSec] = useState(0);
  const deepElapsedRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const fastPollRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const qualityPollRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const fastAttemptsRef = useRef(0);
  const qualityAttemptsRef = useRef(0);
  const MAX_FAST_POLL = 8;
  const MAX_QUALITY_POLL = 10;

  useEffect(() => { return () => { if (fastPollRef.current) clearTimeout(fastPollRef.current); if (qualityPollRef.current) clearTimeout(qualityPollRef.current); if (deepElapsedRef.current) clearInterval(deepElapsedRef.current); }; }, []);

  // N2: parse resumeId from URL once via useMemo — avoid useEffect+StrictMode hook crash
  const parsedUrlResumeId = useMemo(() => parseResumeIdFromQuery(resumeId), [resumeId]);

  // ── Load resume options ──────────────────────────────────────────────────
  const [urlResumeNotFound, setUrlResumeNotFound] = useState(false);

  useEffect(() => {
    loadResumeOptionViewModelsFromApi()
      .then((opts) => {
        setResumeOptions(opts);
        if (opts.length === 0) {
          if (parsedUrlResumeId) setUrlResumeNotFound(true);
          setSelectedOption(null); setLoadingOptions(false); return;
        }
        if (parsedUrlResumeId) {
          const found = opts.find((o) => o.id === parsedUrlResumeId);
          if (!found) {
            setUrlResumeNotFound(true);
            setSelectedOption(null);
            setLoadingOptions(false);
            return;
          }
          setSelectedOption(found);
        } else {
          // L2: no resumeId in URL — redirect to URL with current resumeId
          const cur = opts.find((o) => o.isCurrent) ?? opts[0];
          if (cur) {
            navigate(`/recommend?resumeId=${cur.id}`, { replace: true });
            setLoadingOptions(false);
            return;
          }
          setSelectedOption(null);
          setLoadingOptions(false);
          return;
        }
        setLoadingOptions(false);
      })
      .catch((err) => { const r = getAuthRedirectIf401(err, location.pathname, location.search); if (r) { navigate(r, { replace: true }); return; } setLoadingOptions(false); });
  }, [parsedUrlResumeId]);

  // ── I2+J2+L2: Quality-first entry — try latest quality, fallback to fast ─
  useEffect(() => {
    if (!selectedOption) return;
    const rid = selectedOption.id;
    if (!rid || isNaN(rid)) return;
    setPhase("loading_latest"); setErrorMessage("");
    setIsLatestLoading(true);

    // 1. Try latest quality first
    loadLatestRecommendationViewModel(rid, "quality")
      .then((latest) => {
        const genBy = latest.generatedBy ?? "";
        if (latest.hasResult && latest.result && latest.result.recommendedJobs.length > 0
            && !isMockGeneratedBy(genBy)) {
          const count = latest.result.recommendedJobs.length;
          if (count >= 5) {
            // Complete Deep quality result — show as Deep
            setDeepResult(latest.result);
            setActiveTier("quality");
            setActiveGeneratedBy(genBy);
            setPhase("deep_done");
            setIsLatestLoading(false);
            setSelJobId(latest.result.recommendedJobs[0].jobId);
            return;
          }
          // H21-F: Incomplete quality (1-4 jobs) — record only, do NOT pollute deepResult
          setDeepHistoryIncomplete(true);
          setDeepHistoryIncompleteCount(count);
          // deepResult stays null; phase will be overwritten by Fast fallback below
        }
        // 2. Fallback: try latest fast
        loadLatestRecommendationViewModel(rid, "fast")
          .then((fastLatest) => {
            const fGen = fastLatest.generatedBy ?? "";
            if (fastLatest.hasResult && fastLatest.result && fastLatest.result.recommendedJobs.length > 0
                && !isMockGeneratedBy(fGen)) {
              setFastResult(fastLatest.result);
              setActiveTier("fast");
              setActiveGeneratedBy(fGen);
              setPhase("fast_done");
              setIsLatestLoading(false);
              setSelJobId(fastLatest.result.recommendedJobs[0].jobId);
              return;
            }
            // No cached fast result → generate fresh Fast
            generateFast(rid);
          })
          .catch(() => generateFast(rid));
      })
      .catch(() => {
        // Quality latest failed → try fast latest
        loadLatestRecommendationViewModel(rid, "fast")
          .then((fastLatest) => {
            const fGen = fastLatest.generatedBy ?? "";
            if (fastLatest.hasResult && fastLatest.result && fastLatest.result.recommendedJobs.length > 0
                && !isMockGeneratedBy(fGen)) {
              setFastResult(fastLatest.result);
              setActiveTier("fast");
              setActiveGeneratedBy(fGen);
              setPhase("fast_done");
              setIsLatestLoading(false);
              setSelJobId(fastLatest.result.recommendedJobs[0].jobId);
              return;
            }
            generateFast(rid);
          })
          .catch(() => generateFast(rid));
      });
  }, [selectedOption]);

  function generateFast(rid: number) {
    setPhase("fast_loading"); fastAttemptsRef.current = 0;
    const pollFast = () => {
      loadRecommendationViewModel({ resume_id: rid, limit: 5, generation_tier: "fast" as any })
        .then((vm) => {
          if (vm.profileStatus === "ready") { setFastResult(vm); setActiveTier("fast"); setPhase("fast_done"); setIsLatestLoading(false); if (vm.recommendedJobs.length > 0) setSelJobId(vm.recommendedJobs[0].jobId); return; }
          if (vm.profileStatus === "failed") { setPhase("profile_failed"); setIsLatestLoading(false); setErrorMessage(vm.profileMessage || "简历信息分析失败，请稍后重试。"); return; }
          fastAttemptsRef.current++;
          if (fastAttemptsRef.current >= MAX_FAST_POLL) { setPhase("fast_error"); setErrorMessage("简历分析时间较长，可稍后重试。"); return; }
          fastPollRef.current = setTimeout(pollFast, 1800);
        })
        .catch((err) => {
          console.groupCollapsed("[RecommendPage] Fast API failed");
          console.error(err);
          console.info("context", { action: "generateFast", resumeId: rid, tier: "fast" });
          console.groupEnd();
          if (err instanceof ContractError) {
            setPhase("fast_error"); setErrorMessage("数据解析失败，请刷新页面后重试。");
          } else {
            setPhase("fast_error"); setErrorMessage("推荐结果加载失败，请稍后重试。");
          }
        });
    };
    pollFast();
  }

  // ── Deep recommendation ──────────────────────────────────────────────────
  async function handleDeepRecommendation() {
    if (!selectedOption) return;

    // H17-D: increment run sequence to guard against stale requests
    const runId = ++deepRunSeqRef.current;
    cancelledDeepRunRef.current = 0;
    setDeepReadyInBackground(false);

    setPhase("deep_loading"); setDeepErrorMessage(""); setDeepBackendDone(false); qualityAttemptsRef.current = 0;
    // L3: start elapsed timer
    setDeepElapsedSec(0);
    if (deepElapsedRef.current) clearInterval(deepElapsedRef.current);
    deepElapsedRef.current = setInterval(() => setDeepElapsedSec((s) => s + 1), 1000);

    const pollQuality = async () => {
      try {
        const vm = await loadRecommendationViewModel({ resume_id: selectedOption.id, limit: 5, generation_tier: "quality" as any });
        // H17-D: discard stale request results
        if (runId !== deepRunSeqRef.current) return;

        if (vm.profileStatus === "ready" && vm.recommendedJobs.length > 0) {
          if (deepElapsedRef.current) { clearInterval(deepElapsedRef.current); deepElapsedRef.current = null; }
          // H17-D: check if user cancelled this specific run
          const wasCancelled = cancelledDeepRunRef.current === runId;
          const currentPhase = phaseRef.current;
          setDeepBackendDone(true); setDeepResult(vm); setDeepHistoryIncomplete(false); setDeepHistoryIncompleteCount(0);
          const metaRaw = vm.meta as any;
          const genBy = metaRaw?.generated_by ?? "";
          setActiveGeneratedBy(genBy);
          // K15-P0: deep_fallback_agent is a valid Deep result — do NOT treat as error
          // Only pure fast-tier mock_agent in a Deep context is an error
          if (genBy === "mock_agent") {
            setPhase("deep_error"); setDeepErrorMessage("AI 模型连接失败，当前显示的不是 AI 生成结果。");
            return;
          }
          if (vm.recommendedJobs.length > 0) setSelJobId(vm.recommendedJobs[0].jobId);
          // H17-D: if user returned to Fast, don't force-switch — show background complete entry
          if (wasCancelled || currentPhase === "fast_done") {
            setActiveTier("fast");
            setPhase("fast_done");
            setDeepReadyInBackground(true);
            return;
          }
          setActiveTier("quality");
          setPhase("deep_done");
          return;
        }
        // H17-D: guard against stale poll results
        if (runId !== deepRunSeqRef.current) return;

        if (vm.profileStatus === "pending") { setPhase("deep_profile_pending"); return; }
        if (vm.profileStatus === "failed") { setPhase("deep_error"); setDeepErrorMessage(vm.profileMessage || "深度分析生成失败，已保留当前结果。"); return; }
        qualityAttemptsRef.current++;
        if (qualityAttemptsRef.current >= MAX_QUALITY_POLL) { setPhase("deep_timeout"); return; }
        qualityPollRef.current = setTimeout(pollQuality, 2000);
      } catch (caught: any) {
        if (deepElapsedRef.current) { clearInterval(deepElapsedRef.current); deepElapsedRef.current = null; }
        // H17-D: discard errors from stale requests
        if (runId !== deepRunSeqRef.current) return;
        console.groupCollapsed("[RecommendPage] Deep API failed");
        console.error(caught);
        console.info("context", { action: "generateDeep", resumeId: selectedOption?.id, tier: "quality" });
        console.groupEnd();
        const errType = classifyDeepError(caught);
        // K15-P0-D6: distinct handling for gateway timeout vs other errors
        if (errType === "LLM_UNAVAILABLE") {
          setPhase("deep_unavailable"); setDeepErrorMessage("深度推荐需要调用大模型服务，当前服务未连接成功。你仍可使用快速推荐结果。");
        } else if (errType === "GATEWAY_TIMEOUT") {
          setPhase("deep_timeout"); setDeepErrorMessage("深度推荐生成时间较长，已超过网关等待，建议稍后重试。已为你保留快速推荐结果。");
        } else if (errType === "NETWORK_ERROR") {
          setPhase("deep_error"); setDeepErrorMessage("网络连接异常，深度推荐暂时不可用。已为你保留快速推荐结果，可稍后重试。");
        } else {
          setPhase("deep_error"); setDeepErrorMessage("深度分析暂时生成失败，已为你保留快速推荐结果，可稍后重试。");
        }
      }
    };
    if (qualityPollRef.current) clearTimeout(qualityPollRef.current);
    pollQuality();
  }

  // H17-D: view Deep result when it completed in background
  function handleViewDeepBackground() {
    setActiveTier("quality");
    setPhase("deep_done");
    setDeepReadyInBackground(false);
    if (deepResult?.recommendedJobs?.length) setSelJobId(deepResult.recommendedJobs[0].jobId);
  }

  // H17-D: activeResult respects user's tier choice — only show Deep when deep_done phase
  const deepJobs = deepResult?.recommendedJobs ?? [];
  const fastJobs = fastResult?.recommendedJobs ?? [];
  const deepReady = deepJobs.length >= 5;
  const fastReady = fastJobs.length > 0;
  // Show Deep when user explicitly chose it (deep_done phase + activeTier quality + deep ready)
  const isDeepDisplay = (phase === "deep_done" && activeTier === "quality" && deepReady);
  const activeResult = isDeepDisplay ? deepResult
    : fastReady ? fastResult
    : deepResult ?? fastResult;
  const jobs = activeResult?.recommendedJobs ?? [];
  const selectedJob = selJobId ? (jobs.find((j) => j.jobId === selJobId) ?? jobs[0]) : jobs[0];

  // ── Loading states ───────────────────────────────────────────────────────
  if (loadingOptions) return <PageLoading message="正在加载简历信息…" />;

  // User has resumes but the URL resumeId doesn't belong to them
  if (urlResumeNotFound) {
    return (
      <ContentPageShell maxWidthClassName="max-w-[1440px]" contentClassName="pb-12">
        <PageError
          title="简历不存在或不属于当前账号"
          description="当前链接中的简历无法访问。请从你的简历列表中选择一份简历后再生成岗位推荐。"
        >
          <div className="flex justify-center gap-3">
            <Link to="/resumes" className="px-4 py-2 bg-gray-100 text-gray-700 text-sm rounded-lg hover:bg-gray-200">返回我的简历</Link>
          </div>
        </PageError>
      </ContentPageShell>
    );
  }

  // No resumes at all — show empty state with CTA
  if (resumeOptions.length === 0) {
    return (
      <ContentPageShell maxWidthClassName="max-w-[1440px]" contentClassName="pb-12">
        <PageEmpty
          title="你还没有可用于推荐的简历"
          description="请先创建或导入一份简历，Offer捕手才能基于你的经历推荐合适岗位。"
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
        <PageError title="无法选择简历" description="请返回简历列表重新选择。" />
      </ContentPageShell>
    );
  }

  if (phase === "loading_latest") return <PageLoading message="正在读取推荐结果…" />;
  if (phase === "fast_loading") {
    return <PageLoading message="正在生成快速推荐，系统正在提取你的教育、项目、实习经历和技能关键词…" />;
  }
  if (phase === "fast_error" || phase === "profile_failed") {
    return (
      <ContentPageShell maxWidthClassName="max-w-[1440px]" contentClassName="pb-12">
        <PageError
          title={errorMessage || "推荐结果加载失败"}
          description="请确认网络连接正常、服务器运行中，或尝试切换简历后重试。"
        >
          <div className="flex justify-center gap-3">
            <button onClick={() => setSelectedOption({ ...selectedOption })} className="px-4 py-2 bg-[#0052D9] text-white text-sm rounded-lg hover:bg-[#0047BD]">重新尝试</button>
            <Link to="/resumes" className="px-4 py-2 bg-gray-100 text-gray-700 text-sm rounded-lg hover:bg-gray-200">返回我的简历</Link>
          </div>
        </PageError>
      </ContentPageShell>
    );
  }
  // L2: only show empty after latest load completes (guard against flash)
  if (isLatestLoading) return <PageLoading message="正在恢复推荐结果…" />;
  if (!activeResult || jobs.length === 0) {
    return (
      <ContentPageShell maxWidthClassName="max-w-[1440px]" contentClassName="pb-12">
        <PageEmpty title="暂无推荐结果" description="当前简历还没有可展示的推荐结果，你可以先生成快速推荐，或生成深度推荐。">
          <Link to="/resumes" className="text-[#0052D9] text-sm underline">切换简历</Link>
        </PageEmpty>
      </ContentPageShell>
    );
  }

  const displayName = selectedOption.name;
  const nowStr = new Date().toLocaleString("zh-CN");

  // K9-1-U-D: header label from phase + activeTier
  function getHeaderLabel(): string {
    switch (phase) {
      case "deep_loading": return "智能体正在深度分析";
      case "deep_profile_pending": return "正在生成深度分析";
      case "deep_done": {
        const base = activeGeneratedBy?.startsWith("llm") ? "深度推荐 · AI 生成" : "深度推荐 · 基础模型";
        return `${base}（${actualCount} 个岗位）`;
      }
      case "deep_timeout": return "深度分析生成较慢 · 已保留当前结果";
      case "deep_error": return "深度推荐生成失败 · 已保留当前结果";
      case "deep_unavailable": return "深度推荐服务暂不可用 · 快速推荐仍可使用";
      default: return sourceLabel(activeGeneratedBy) || `快速推荐 · 基础模型（${actualCount} 个岗位）`;
    }
  }
  const isDeepUnavailable = phase === "deep_unavailable";
  const actualCount = jobs.length;
  const headerLabel = getHeaderLabel();

  return (
    <ContentPageShell maxWidthClassName="max-w-[1440px]" contentClassName="pb-12">
      {/* L4: Top Insight Card */}
      <div className="mb-6">
        <h1 className="text-[22px] font-semibold text-slate-800 flex items-center gap-2 mb-1">
          <Sparkles className="w-5 h-5 text-[#0052D9]" />Offer捕手岗位推荐
        </h1>
        <p className="text-sm text-slate-500">
          基于 <span className="font-medium text-slate-700">{displayName}</span> 的简历画像，为您推荐 {actualCount} 个{isDeepDisplay ? "深度分析" : "快速匹配"}岗位
        </p>
        <div className="flex flex-wrap items-center gap-3 mt-2 text-xs text-slate-400">
          <span className="flex items-center gap-1.5 text-[#0052D9] font-medium">{headerLabel}</span>
          <span className="text-slate-300">|</span><span>{nowStr}</span>
          {activeResult.profileDominantProfile && (<><span className="text-slate-300">|</span><span className="text-[#0052D9]">{humanizeProfile(activeResult.profileDominantProfile)}</span></>)}
        </div>

        {/* L4: AI Insight Card (replaces old blue gradient) */}
        {isDeepDisplay && (activeResult.globalSummary || activeResult.candidatePositioning) && (
          <div className="mt-4 bg-white border border-slate-200 rounded-2xl p-5 shadow-sm">
            <div className="flex items-center gap-2 mb-3">
              <Sparkles className="w-4 h-4 text-[#0052D9]" />
              <span className="text-sm font-semibold text-slate-800">AI 推荐洞察</span>
              {activeResult.qualityStatus && (
                <span className="text-[10px] px-2 py-0.5 bg-blue-50 text-blue-600 rounded-full ml-auto">
                  级联管线已完成
                </span>
              )}
            </div>
            <div className="grid grid-cols-[auto_1fr] gap-x-4 gap-y-2">
              {activeResult.candidatePositioning && (
                <>
                  <span className="text-xs font-medium text-slate-400 whitespace-nowrap pt-0.5">候选人定位</span>
                  <p className="text-[13px] text-slate-700 leading-relaxed">{activeResult.candidatePositioning}</p>
                </>
              )}
              {activeResult.globalSummary && (
                <>
                  <span className="text-xs font-medium text-slate-400 whitespace-nowrap pt-0.5">推荐洞察</span>
                  <p className="text-[13px] text-slate-700 leading-relaxed">{activeResult.globalSummary}</p>
                </>
              )}
            </div>
          </div>
        )}

        {/* J2: Incomplete Deep history warning — H21-F: gated to Deep display only */}
        {isDeepDisplay && deepHistoryIncomplete && (
          <div className="mt-4 bg-amber-50 border border-amber-200 rounded-2xl p-4 shadow-sm">
            <div className="flex items-start gap-3">
              <AlertCircle className="w-5 h-5 text-amber-500 shrink-0 mt-0.5" />
              <div>
                <p className="text-[13px] font-medium text-amber-800">深度推荐历史不完整</p>
                <p className="text-[12px] text-amber-600 mt-0.5">
                  之前保存的深度推荐仅有 {deepHistoryIncompleteCount} 个岗位（不足 5 个），建议重新生成以获取完整的岗位推荐分析。
                </p>
              </div>
            </div>
          </div>
        )}

        {/* Fast mode hint */}
        {!isDeepDisplay && (
          <div className="mt-4 bg-white border border-slate-200 rounded-2xl p-4 shadow-sm">
            <div className="flex items-start gap-3">
              <div className="w-8 h-8 rounded-full bg-blue-50 flex items-center justify-center shrink-0 mt-0.5">
                <Sparkles className="w-4 h-4 text-[#0052D9]" />
              </div>
              <div>
                <p className="text-[13px] font-medium text-slate-700">快速推荐</p>
                <p className="text-[12px] text-slate-400 mt-0.5">基于简历画像和岗位画像进行快速排序，适合先了解方向。想获得更完整的证据、风险和投递策略，可以生成深度推荐。</p>
              </div>
            </div>
          </div>
        )}
      </div>

      <div className="flex gap-6 items-start">
        {/* ── Left: Top5 Ranking ─────────────────────────────────────────── */}
        <div className="w-[370px] shrink-0 flex flex-col gap-3">
          <div className="flex items-center justify-between px-1 mb-1">
            <span className="text-xs font-medium text-slate-400 uppercase tracking-wide">推荐榜单</span>
            <span className="text-[10px] text-slate-300">{isDeepDisplay ? "Deep" : "Fast"} · Top {actualCount}</span>
          </div>
          {jobs.map((job) => {
            const isSel = selectedJob && job.jobId === selectedJob.jobId;
            const badgeVar = priorityBadgeVariant(job.priorityTone);
            const showTags = job.displayTags.slice(0, 3);
            const rankLabel = job.rank === 1 ? "最优先" : job.rank === 2 ? "推荐" : job.rank === 3 ? "推荐" : job.rank === 4 ? "可考虑" : "备选";
            const rankTone = job.rank <= 2 ? "text-[#0052D9]" : job.rank <= 3 ? "text-slate-500" : "text-slate-400";
            return (
              <div key={job.jobId} onClick={() => setSelJobId(job.jobId)}
                className={`group relative p-4 bg-white border rounded-2xl cursor-pointer transition-all duration-200 hover:shadow-md ${
                  isSel ? "border-[#0052D9] bg-blue-50/40 shadow-sm" : "border-slate-200 hover:border-slate-300"}`}>
                {isSel && <div className="absolute left-0 top-3 bottom-3 w-0.5 bg-[#0052D9] rounded-full" />}
                {/* Rank + Score row */}
                <div className="flex items-center justify-between mb-2">
                  <div className="flex items-center gap-2">
                    <span className={`text-xs font-bold ${rankTone} bg-white border ${isSel ? "border-[#0052D9]/20" : "border-slate-200"} rounded-full w-6 h-6 flex items-center justify-center`}>{job.rank}</span>
                    <span className={`text-[11px] font-medium ${rankTone}`}>{rankLabel}</span>
                  </div>
                  <div className="flex items-center gap-1.5">
                    <span className={`text-[13px] font-bold ${isSel ? "text-[#0052D9]" : "text-slate-700"}`}>{job.score}</span>
                    <span className="text-[10px] text-slate-400">分</span>
                  </div>
                </div>
                {/* Title + location */}
                <h3 className={`font-bold text-[14px] truncate mb-1 ${isSel ? "text-[#0052D9]" : "text-slate-800"}`}>{job.title}</h3>
                <div className="text-[11px] text-slate-400 mb-2">{job.city} · {job.department}</div>
                {/* Core strength or one-sentence reason */}
                {(job.candidateCoreStrength || job.oneSentenceReason) && (
                  <p className="text-[12px] text-slate-500 line-clamp-2 leading-relaxed mb-2">
                    {job.candidateCoreStrength || job.oneSentenceReason}
                  </p>
                )}
                {/* Tags */}
                {showTags.length > 0 && (
                  <div className="flex flex-wrap gap-1">
                    {showTags.map((t, i) => (
                      <span key={i} className="text-[10px] px-1.5 py-0.5 bg-slate-100 text-slate-500 rounded-md">{t}</span>
                    ))}
                    {job.displayTags.length > 3 && <span className="text-[10px] text-slate-300 self-center">+{job.displayTags.length - 3}</span>}
                  </div>
                )}
              </div>
            );
          })}
        </div>

        {/* ── Center: Assessment Report ─────────────────────────────────── */}
        <div className="flex-1 min-w-0 max-w-[700px] bg-white border border-slate-200 rounded-2xl shadow-sm overflow-hidden">
          {selectedJob && (<>
            {/* Job header */}
            <div className="px-6 py-5 border-b border-slate-100">
              <div className="flex items-center justify-between mb-1">
                <h2 className="text-lg font-bold text-slate-800">{selectedJob.title}</h2>
                <div className="flex items-center gap-2 shrink-0">
                  <span className="text-[22px] font-bold text-[#0052D9]">{selectedJob.score}</span>
                  <span className="text-[11px] text-slate-400">分</span>
                </div>
              </div>
              <div className="text-[12px] text-slate-400">{selectedJob.city} · {selectedJob.category} · {selectedJob.department}</div>
            </div>

            <div className="p-6 space-y-6">
              {/* Core Judgement */}
              <section>
                <h3 className="text-[11px] font-semibold text-slate-400 uppercase tracking-wide mb-2">核心判断</h3>
                <div className="bg-blue-50/60 border border-blue-100/60 rounded-xl p-4">
                  <p className="text-[13px] text-slate-700 leading-relaxed">{selectedJob.oneSentenceReason || selectedJob.jobSpecificReason}</p>
                  {selectedJob.fitConfidenceReason && (
                    <p className="text-[11px] text-slate-400 mt-1.5">{selectedJob.fitConfidenceReason}</p>
                  )}
                </div>
              </section>

              {/* Why Recommended (Deep only) */}
              {isDeepDisplay && selectedJob.jobSpecificReason && (
                <section>
                  <h3 className="text-[11px] font-semibold text-slate-400 uppercase tracking-wide mb-2">推荐依据</h3>
                  <p className="text-[13px] text-slate-600 leading-relaxed">{selectedJob.jobSpecificReason}</p>
                </section>
              )}

              {/* Evidence */}
              {(isDeepDisplay ? selectedJob.resumeEvidence.length > 0 : selectedJob.matchHighlights.length > 0) && (
                <section>
                  <h3 className="text-[11px] font-semibold text-slate-400 uppercase tracking-wide mb-2">匹配证据</h3>
                  <div className="bg-emerald-50/40 border border-emerald-100/40 rounded-xl p-4">
                    <ul className="space-y-1.5">
                      {(isDeepDisplay ? selectedJob.resumeEvidence : selectedJob.matchHighlights).slice(0, 3).map((e, i) => (
                        <li key={i} className="flex items-start gap-2 text-[13px] text-slate-700">
                          <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 mt-1.5 shrink-0" />
                          {e}
                        </li>
                      ))}
                    </ul>
                  </div>
                </section>
              )}

              {/* Risks */}
              {selectedJob.jobSpecificRisks.length > 0 && (
                <section>
                  <h3 className="text-[11px] font-semibold text-slate-400 uppercase tracking-wide mb-2">主要风险</h3>
                  <div className="bg-amber-50/40 border border-amber-100/40 rounded-xl p-4">
                    <ul className="space-y-1.5">
                      {selectedJob.jobSpecificRisks.slice(0, 3).map((r, i) => (
                        <li key={i} className="flex items-start gap-2 text-[13px] text-amber-800">
                          <span className="w-1.5 h-1.5 rounded-full bg-amber-400 mt-1.5 shrink-0" />
                          {r}
                        </li>
                      ))}
                    </ul>
                  </div>
                </section>
              )}

              {/* Strategy + Improvement */}
              {isDeepDisplay && (selectedJob.applicationStrategy || selectedJob.suggestedImprovement) && (
                <section>
                  <h3 className="text-[11px] font-semibold text-slate-400 uppercase tracking-wide mb-2">行动建议</h3>
                  <div className="bg-slate-50 border border-slate-100 rounded-xl p-4 space-y-3">
                    {selectedJob.applicationStrategy && (
                      <div>
                        <span className="text-[11px] font-medium text-slate-500">投递策略</span>
                        <p className="text-[13px] text-slate-700 mt-0.5">{selectedJob.applicationStrategy}</p>
                      </div>
                    )}
                    {selectedJob.suggestedImprovement && (
                      <div>
                        <span className="text-[11px] font-medium text-slate-500">简历补强</span>
                        <p className="text-[13px] text-slate-700 mt-0.5">{selectedJob.suggestedImprovement}</p>
                      </div>
                    )}
                  </div>
                </section>
              )}

              {/* Next action */}
              <div className="flex gap-3 pt-2">
                <Link to={`/jobs/${selectedJob.jobId}`} className="flex-1 text-center h-10 flex items-center justify-center bg-[#0052D9] text-white text-[13px] rounded-xl hover:bg-[#0047BD] transition-colors font-medium">进入岗位详情</Link>
                <Link to={`/match?jobId=${selectedJob.jobId}&resumeId=${selectedOption.id}`} className="flex-1 text-center h-10 flex items-center justify-center border border-slate-200 text-slate-600 text-[13px] rounded-xl hover:bg-slate-50 transition-colors">匹配分析</Link>
              </div>
            </div>
          </>)}
        </div>

        {/* ── Right: Analysis Console ────────────────────────────────────── */}
        <aside className="w-[310px] shrink-0 flex flex-col gap-4">
          {/* Resume card */}
          <div className="bg-white border border-slate-200 rounded-2xl p-5 shadow-sm">
            <h3 className="text-[11px] font-semibold text-slate-400 uppercase tracking-wide mb-3">当前简历</h3>
            <p className="text-sm font-bold text-slate-800">{selectedOption.name}</p>
            {activeResult.profileDominantProfile && (
              <p className="text-[12px] text-slate-500 mt-1">{humanizeProfile(activeResult.profileDominantProfile)}</p>
            )}
          </div>

          {/* M4: Analysis Console with capability status — H17-D: 4 clear states */}
          <div className="bg-white border border-slate-200 rounded-2xl p-5 shadow-sm">
            <h3 className="text-[11px] font-semibold text-slate-400 uppercase tracking-wide mb-3">智能分析</h3>

            {/* ── H21-F: State 1a — Fast ready + Deep history incomplete ── */}
            {!isDeepDisplay && !deepResult && deepHistoryIncomplete && phase !== "deep_loading" && phase !== "deep_unavailable" && (
              <>
                <div className="space-y-1.5 mb-4 text-[11px]">
                  <div className="flex items-center gap-2">
                    <span className="w-1.5 h-1.5 rounded-full bg-blue-400" />
                    <span className="text-slate-500">快速推荐：可用</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="w-1.5 h-1.5 rounded-full bg-amber-400" />
                    <span className="text-slate-500">深度推荐：历史不完整</span>
                  </div>
                </div>
                <div className="flex items-center gap-2 mb-4">
                  <span className="w-2 h-2 rounded-full bg-blue-400" />
                  <span className="text-[13px] font-medium text-slate-700">快速推荐</span>
                </div>
                <div className="mb-4 p-2.5 bg-amber-50 border border-amber-100 rounded-xl text-[11px]">
                  <p className="text-amber-700">历史深度推荐仅有 {deepHistoryIncompleteCount} 个岗位，建议重新生成完整分析。</p>
                </div>
                <button
                  type="button"
                  onClick={handleDeepRecommendation}
                  className="w-full h-9 flex items-center justify-center gap-1.5 bg-[#0052D9] text-white text-[12px] rounded-xl hover:bg-[#0047BD] transition-colors font-medium"
                >
                  <Sparkles className="w-3 h-3" />
                  重新生成完整深度推荐
                </button>
              </>
            )}

            {/* ── State 1b: Fast only (no Deep result, not loading, not incomplete) ── */}
            {!isDeepDisplay && !deepResult && !deepHistoryIncomplete && phase !== "deep_loading" && phase !== "deep_unavailable" && (
              <>
                <div className="space-y-1.5 mb-4 text-[11px]">
                  <div className="flex items-center gap-2">
                    <span className="w-1.5 h-1.5 rounded-full bg-blue-400" />
                    <span className="text-slate-500">快速推荐：可用</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="w-1.5 h-1.5 rounded-full bg-slate-300" />
                    <span className="text-slate-500">深度推荐：未生成</span>
                  </div>
                </div>
                <div className="flex items-center gap-2 mb-4">
                  <span className="w-2 h-2 rounded-full bg-blue-400" />
                  <span className="text-[13px] font-medium text-slate-700">快速推荐</span>
                </div>
                <button
                  type="button"
                  onClick={handleDeepRecommendation}
                  className="w-full h-9 flex items-center justify-center gap-1.5 bg-[#0052D9] text-white text-[12px] rounded-xl hover:bg-[#0047BD] transition-colors font-medium"
                >
                  <Sparkles className="w-3 h-3" />
                  生成深度推荐
                </button>
              </>
            )}

            {/* ── State 2: Deep loading ── */}
            {phase === "deep_loading" && (
              <>
                <div className="space-y-1.5 mb-4 text-[11px]">
                  <div className="flex items-center gap-2">
                    <span className="w-1.5 h-1.5 rounded-full bg-slate-300" />
                    <span className="text-slate-400">快速推荐：后台可用</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="w-1.5 h-1.5 rounded-full bg-blue-400" style={{ animation: "h17c-pulse-dot 1.5s ease-in-out infinite" }} />
                    <span className="text-slate-700 font-medium">深度推荐：分析中</span>
                  </div>
                </div>
                <div className="flex items-center gap-2 mb-4">
                  <span className="w-2 h-2 rounded-full bg-blue-400" />
                  <span className="text-[13px] font-medium text-slate-700">深度推荐智能体</span>
                </div>
                <div className="flex items-center gap-2 text-xs text-slate-500 mb-4">
                  <Loader2 className="w-3.5 h-3.5 animate-spin" />深度分析进行中…
                </div>
                <button onClick={() => {
                  if (deepElapsedRef.current) clearInterval(deepElapsedRef.current);
                  cancelledDeepRunRef.current = deepRunSeqRef.current;
                  setActiveTier("fast");
                  setPhase("fast_done");
                  setDeepErrorMessage("已返回快速推荐结果。深度推荐完成后可手动查看。");
                }} className="w-full h-9 flex items-center justify-center gap-1.5 border border-slate-200 text-slate-500 text-[12px] rounded-xl hover:bg-slate-50 transition-colors">
                  返回快速推荐结果
                </button>
              </>
            )}

            {/* ── State 3: Deep displayed ── */}
            {isDeepDisplay && (
              <>
                <div className="space-y-1.5 mb-4 text-[11px]">
                  <div className="flex items-center gap-2">
                    <span className="w-1.5 h-1.5 rounded-full bg-slate-300" />
                    <span className="text-slate-400">快速推荐：后台可用</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="w-1.5 h-1.5 rounded-full bg-emerald-400" />
                    <span className="text-slate-500">深度推荐：可用</span>
                  </div>
                </div>
                <div className="flex items-center gap-2 mb-4">
                  <span className="w-2 h-2 rounded-full bg-emerald-500" />
                  <span className="text-[13px] font-medium text-slate-700">深度推荐</span>
                  {activeGeneratedBy && (
                    <span className="text-[10px] text-slate-400 ml-auto">{sourceLabel(activeGeneratedBy)}</span>
                  )}
                </div>
                <div className="space-y-2 mb-4 text-[11px] text-slate-500">
                  <p className="flex items-center gap-1.5"><CheckCircle2 className="w-3 h-3 text-emerald-500" />深度分析已完成</p>
                  {activeResult.agentStatus && agentStatusLabel(activeResult.agentStatus) && (
                    <p className="text-slate-400 pl-5">{agentStatusLabel(activeResult.agentStatus)}</p>
                  )}
                </div>
                <button onClick={handleDeepRecommendation} className="w-full h-9 flex items-center justify-center gap-1.5 border border-[#0052D9] text-[#0052D9] text-[12px] rounded-xl hover:bg-blue-50 transition-colors font-medium mb-2">
                  <RefreshCw className="w-3 h-3" />重新生成深度推荐
                </button>
              </>
            )}

            {/* ── State 4: Fast displayed but Deep completed in background ── */}
            {!isDeepDisplay && deepResult && deepReady && phase !== "deep_loading" && phase !== "deep_unavailable" && (
              <>
                <div className="space-y-1.5 mb-4 text-[11px]">
                  <div className="flex items-center gap-2">
                    <span className="w-1.5 h-1.5 rounded-full bg-blue-400" />
                    <span className="text-slate-500">快速推荐：可用</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="w-1.5 h-1.5 rounded-full bg-emerald-400" />
                    <span className="text-slate-500">深度推荐：已完成</span>
                  </div>
                </div>
                <div className="flex items-center gap-2 mb-4">
                  <span className="w-2 h-2 rounded-full bg-blue-400" />
                  <span className="text-[13px] font-medium text-slate-700">快速推荐</span>
                </div>
                <div className="mb-4 p-2.5 bg-emerald-50 border border-emerald-100 rounded-xl text-[11px]">
                  <p className="font-medium text-emerald-700">深度推荐已完成</p>
                  <p className="text-emerald-500 mt-0.5">可查看刚生成的深度推荐结果。</p>
                </div>
                <div className="space-y-2">
                  <button onClick={handleViewDeepBackground} className="w-full h-9 flex items-center justify-center gap-1.5 bg-[#0052D9] text-white text-[12px] rounded-xl hover:bg-[#0047BD] transition-colors font-medium">
                    <Sparkles className="w-3 h-3" />查看深度推荐
                  </button>
                  <button onClick={handleDeepRecommendation} className="w-full h-9 flex items-center justify-center gap-1.5 border border-[#0052D9] text-[#0052D9] text-[12px] rounded-xl hover:bg-blue-50 transition-colors font-medium">
                    <RefreshCw className="w-3 h-3" />重新生成深度推荐
                  </button>
                </div>
              </>
            )}

            {/* ── Deep unavailable ── */}
            {phase === "deep_unavailable" && (
              <>
                <div className="mb-4 p-2.5 bg-amber-50 border border-amber-100 rounded-xl text-[11px]">
                  <p className="font-medium text-amber-700">深度推荐服务暂不可用</p>
                  <p className="text-amber-500 mt-0.5">{deepErrorMessage || "快速推荐仍可使用。"}</p>
                </div>
                <div className="space-y-2">
                  <button onClick={handleDeepRecommendation} className="w-full h-9 flex items-center justify-center gap-1.5 border border-slate-200 text-slate-500 text-[12px] rounded-xl hover:bg-slate-50 transition-colors">
                    <RefreshCw className="w-3 h-3" />稍后重试深度推荐
                  </button>
                  <button onClick={() => { if (fastResult) setPhase("fast_done"); }} className="w-full h-9 flex items-center justify-center gap-1.5 bg-[#0052D9] text-white text-[12px] rounded-xl hover:bg-[#0047BD] transition-colors font-medium">
                    查看快速推荐
                  </button>
                </div>
              </>
            )}

            {/* ── Deep error / timeout ── */}
            {(phase === "deep_error" || phase === "deep_timeout") && (
              <>
                <div className="mb-4 p-2.5 bg-amber-50 border border-amber-100 rounded-xl text-[11px] text-amber-700">
                  <p className="font-medium">{phase === "deep_timeout" ? "生成时间较长" : "暂时生成失败"}</p>
                  <p className="text-amber-500 mt-0.5">{deepErrorMessage || "快速推荐结果仍可参考。"}</p>
                </div>
                <div className="space-y-2">
                  <button onClick={handleDeepRecommendation} className="w-full h-9 flex items-center justify-center gap-1.5 bg-[#0052D9] text-white text-[12px] rounded-xl hover:bg-[#0047BD] transition-colors font-medium">
                    <RefreshCw className="w-3 h-3" />重试深度推荐
                  </button>
                  <button onClick={() => { if (fastResult) setPhase("fast_done"); }} className="w-full h-9 flex items-center justify-center gap-1.5 border border-slate-200 text-slate-500 text-[12px] rounded-xl hover:bg-slate-50 transition-colors">
                    返回快速推荐
                  </button>
                </div>
              </>
            )}

            {/* Footer hint */}
            <p className="text-[10px] text-slate-400 mt-3 text-center leading-relaxed">
              {isDeepDisplay
                ? "当前结果已完成多 Agent 分析，可重新生成获取最新推荐。"
                : deepReadyInBackground || (deepResult && deepReady)
                ? "深度推荐已完成，可点击上方按钮查看。"
                : phase === "deep_unavailable"
                ? "深度推荐需要大模型服务，当前服务未连接成功。"
                : "深度推荐会调用多 Agent 协同分析链路，完成后将自动展示结果。"}
            </p>
          </div>
        </aside>
      </div>

      <DeepAgentLoadingOverlay
        open={phase === "deep_loading"}
        elapsedSeconds={deepElapsedSec}
        onBackToFast={() => {
          if (deepElapsedRef.current) clearInterval(deepElapsedRef.current);
          // H17-D: mark current run as cancelled so completion won't force-switch
          cancelledDeepRunRef.current = deepRunSeqRef.current;
          setActiveTier("fast");
          setPhase("fast_done");
          setDeepErrorMessage("已返回快速推荐结果。深度推荐完成后可手动查看。");
        }}
      />
    </ContentPageShell>
  );
}
