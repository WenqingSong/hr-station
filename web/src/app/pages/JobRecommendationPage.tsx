import React, { useState, useEffect, useRef } from "react";
import { Link, useSearchParams, useNavigate, useLocation } from "react-router";
import { Sparkles, Loader2, AlertCircle, RefreshCw } from "lucide-react";
import { loadRecommendationViewModel, loadResumeOptionViewModelsFromApi } from "../../api";
import type { RecommendationPageViewModel, ResumeOptionViewModel, RecommendedJobViewModel, RecommendationPriorityTone } from "../../adapters";
import { PageLoading, PageEmpty } from "../components/PageState";
import { ContentPageShell } from "../components/layout/ContentPageShell";
import { getAuthRedirectIf401 } from "../utils/authRedirect";
import { Badge } from "../components/ui/Badge";

// ─── Helpers ──────────────────────────────────────────────────────────────────

function parseResumeIdFromQuery(resumeId: string | null): number | null {
  if (resumeId && /^\d+$/.test(resumeId)) return Number(resumeId);
  return null;
}

type RecPhase =
  | "fast_loading" | "profile_generating" | "fast_done" | "fast_error"
  | "deep_loading" | "deep_done" | "deep_error" | "profile_failed";

function priorityBadgeVariant(tone: RecommendationPriorityTone): "green" | "blue" | "amber" | "secondary" | "danger" {
  switch (tone) {
    case "strong": return "blue";
    case "high": return "secondary";
    case "try": return "amber";
    case "backup": return "secondary";
    case "not_recommended": return "danger";
  }
}

function priorityText(tone: string): string {
  switch (tone) {
    case "strong": return "强匹配，优先投递";
    case "high": return "高匹配，重点关注";
    case "try": return "可尝试，需要补强";
    case "backup": return "备选参考";
    case "not_recommended": return "不推荐";
    default: return tone;
  }
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

  const [phase, setPhase] = useState<RecPhase>("fast_loading");
  const [fastResult, setFastResult] = useState<RecommendationPageViewModel | null>(null);
  const [deepResult, setDeepResult] = useState<RecommendationPageViewModel | null>(null);
  const [errorMessage, setErrorMessage] = useState("");
  const [deepErrorMessage, setDeepErrorMessage] = useState("");

  const [selJobId, setSelJobId] = useState<number | null>(null);

  // Polling refs
  const fastPollRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const qualityPollRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const fastAttemptsRef = useRef(0);
  const qualityAttemptsRef = useRef(0);
  const MAX_FAST_POLL = 8;
  const MAX_QUALITY_POLL = 10;

  // ── Cleanup polling ──────────────────────────────────────────────────────
  useEffect(() => {
    return () => {
      if (fastPollRef.current) clearTimeout(fastPollRef.current);
      if (qualityPollRef.current) clearTimeout(qualityPollRef.current);
    };
  }, []);

  // ── Load resume options ──────────────────────────────────────────────────
  useEffect(() => {
    loadResumeOptionViewModelsFromApi()
      .then((opts) => {
        setResumeOptions(opts);
        if (opts.length === 0) { setSelectedOption(null); setLoadingOptions(false); return; }
        const target = parseResumeIdFromQuery(resumeId);
        setSelectedOption(target ? opts.find((o) => o.id === target) ?? opts[0] : opts[0]);
        setLoadingOptions(false);
      })
      .catch((err) => {
        const r = getAuthRedirectIf401(err, location.pathname, location.search);
        if (r) { navigate(r, { replace: true }); return; }
        setLoadingOptions(false);
      });
  }, []);

  // ── Fast recommendation auto-trigger with polling ────────────────────────
  const fetchFast = async (rid: number) => {
    const vm = await loadRecommendationViewModel({ resume_id: rid, limit: 5, generation_tier: "fast" as any });
    const status = vm.profileStatus;
    if (status === "ready") {
      setFastResult(vm);
      setPhase("fast_done");
      if (vm.recommendedJobs.length > 0) setSelJobId(vm.recommendedJobs[0].jobId);
      return true;
    }
    if (status === "failed") {
      setPhase("profile_failed");
      setErrorMessage(vm.profileMessage || "简历画像生成失败，请稍后重试。");
      return false;
    }
    // pending/generating/stale → poll
    return false;
  };

  useEffect(() => {
    if (!selectedOption) return;
    const rid = selectedOption.id;
    if (!rid || isNaN(rid)) return;
    setPhase("fast_loading"); setErrorMessage(""); fastAttemptsRef.current = 0;
    const pollFast = () => {
      fetchFast(rid).then((done) => {
        if (done) return;
        fastAttemptsRef.current++;
        if (fastAttemptsRef.current >= MAX_FAST_POLL) {
          setPhase("fast_error");
          setErrorMessage("画像生成时间较长，可稍后重试。");
          return;
        }
        setPhase("profile_generating");
        fastPollRef.current = setTimeout(pollFast, 1800);
      }).catch(() => {
        setPhase("fast_error"); setErrorMessage("推荐结果加载失败，请稍后重试。");
      });
    };
    if (fastPollRef.current) clearTimeout(fastPollRef.current);
    pollFast();
    return () => { if (fastPollRef.current) clearTimeout(fastPollRef.current); };
  }, [selectedOption]);

  // ── Deep recommendation with polling ────────────────────────────────────
  async function handleDeepRecommendation() {
    if (!selectedOption || !fastResult) return;
    setPhase("deep_loading"); setDeepErrorMessage(""); qualityAttemptsRef.current = 0;

    const pollQuality = async () => {
      try {
        const vm = await loadRecommendationViewModel({
          resume_id: selectedOption.id, limit: 5, generation_tier: "quality" as any,
        });
        const status = vm.profileStatus;
        if (status === "ready") {
          setDeepResult(vm); setPhase("deep_done");
          if (vm.recommendedJobs.length > 0) setSelJobId(vm.recommendedJobs[0].jobId);
          return;
        }
        if (status === "failed") {
          setPhase("deep_error");
          setDeepErrorMessage(vm.profileMessage || "深度画像生成失败，已保留快速推荐结果。");
          return;
        }
        qualityAttemptsRef.current++;
        if (qualityAttemptsRef.current >= MAX_QUALITY_POLL) {
          setPhase("deep_error");
          setDeepErrorMessage("深度画像生成时间较长，可稍后重试。");
          return;
        }
        qualityPollRef.current = setTimeout(pollQuality, 2000);
      } catch {
        setPhase("deep_error");
        setDeepErrorMessage("深度分析暂时生成失败，快速推荐结果仍可参考。");
      }
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
    return (
      <PageEmpty title="你还没有可用于推荐的简历" description="请先创建一份简历，系统会基于你的简历生成岗位推荐。">
        <Link to="/resumes/new" className="text-[#0052D9] text-sm underline">去添加简历</Link>
      </PageEmpty>
    );
  }

  // Profile generating / fast loading
  if (phase === "fast_loading" || phase === "profile_generating") {
    return (
      <div className="max-w-[1440px] mx-auto px-6 py-32 text-center">
        <Loader2 className="w-10 h-10 animate-spin text-[#0052D9] mx-auto mb-4" />
        <h2 className="text-lg font-bold text-gray-900 mb-2">正在生成简历基础画像</h2>
        <p className="text-sm text-gray-500 leading-relaxed max-w-md mx-auto">
          系统正在提取你的教育经历、项目经历、实习经历和技能标签，用于快速岗位推荐。
        </p>
      </div>
    );
  }
  if (phase === "fast_error" || phase === "profile_failed") {
    return (
      <div className="max-w-[1440px] mx-auto px-6 py-12 text-center">
        <AlertCircle className="w-8 h-8 text-amber-500 mx-auto mb-3" />
        <p className="text-gray-500 text-sm mb-3">{errorMessage}</p>
        <button onClick={() => setSelectedOption({ ...selectedOption })} className="text-[#0052D9] text-sm underline">
          重新尝试
        </button>
      </div>
    );
  }
  if (!activeResult || jobs.length === 0) {
    return (
      <div className="max-w-[1440px] mx-auto px-6 py-12 text-center">
        <p className="text-gray-500 text-sm">暂无推荐结果。</p>
      </div>
    );
  }

  const displayName = selectedOption.name;
  const sourceLabel = activeResult.sourceLabel;
  const nowStr = new Date().toLocaleString("zh-CN");
  const modeLabel = phase === "deep_done" ? "深度推荐" : "快速推荐";

  return (
    <ContentPageShell maxWidthClassName="max-w-[1440px]" contentClassName="pb-12">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-[22px] font-semibold text-gray-900 flex items-center gap-2">
          <Sparkles className="w-5 h-5 text-[#0052D9]" />
          Offerr鹅岗位推荐
        </h1>
        <p className="mt-1.5 text-sm text-gray-500">
          基于简历 <span className="font-medium text-gray-700">{displayName}</span>（ID:{selectedOption.id}），从岗位池中推荐 Top 5 岗位。
        </p>
        <div className="flex flex-wrap items-center gap-3 mt-3 text-xs text-gray-500">
          <span className="flex items-center gap-1.5"><Sparkles className="w-3 h-3 text-amber-400" /> {modeLabel}</span>
          <span className="text-gray-300">|</span>
          <span>{sourceLabel}</span>
          <span className="text-gray-300">|</span>
          <span>生成时间：{nowStr}</span>
          {activeResult.profileDominantProfile && (
            <>
              <span className="text-gray-300">|</span>
              <span className="text-[#0052D9]">画像：{activeResult.profileDominantProfile}</span>
            </>
          )}
        </div>
      </div>

      <div className="flex gap-6 items-start">
        {/* ── Left: Top 5 list ─────────────────────────────────────────── */}
        <div className="w-[380px] shrink-0 flex flex-col gap-3">
          {jobs.map((job) => {
            const isSel = selectedJob && job.jobId === selectedJob.jobId;
            const badgeVar = priorityBadgeVariant(job.priorityTone);
            return (
              <div
                key={job.jobId}
                onClick={() => setSelJobId(job.jobId)}
                className={`relative p-4 bg-white border rounded-lg cursor-pointer transition-all duration-200 hover:shadow-md ${
                  isSel ? "border-[#0052D9] bg-[#F4F8FF]" : "border-gray-200 hover:border-[#0052D9]"
                }`}
              >
                {isSel && <div className="absolute left-0 top-0 bottom-0 w-1 bg-[#0052D9] rounded-l-lg" />}
                {/* Priority badge */}
                <div className="flex items-center gap-2 mb-1">
                  <Badge variant={badgeVar} className="text-[11px] font-medium px-2">
                    {job.priorityText}
                  </Badge>
                  <span className={`text-[11px] text-gray-400 ${isSel ? "text-[#0052D9]" : ""}`}>{job.score}分</span>
                </div>
                {/* Title */}
                <h3 className={`font-bold text-[14px] truncate mb-1 ${isSel ? "text-[#0052D9]" : "text-gray-900"}`}>
                  <span className="text-xs text-gray-400 mr-1">#{job.rank}</span>{job.title}
                </h3>
                <div className="flex flex-wrap items-center gap-1.5 mb-2 text-[11px] text-gray-500">
                  <span>{job.city}</span><span className="text-gray-300">·</span>
                  <span>{job.category}</span><span className="text-gray-300">·</span>
                  <span>{job.department}</span>
                </div>
                <p className="text-[12px] text-gray-600 line-clamp-2 leading-relaxed">{job.oneSentenceReason}</p>
                <div className="flex flex-wrap gap-1 mt-2">
                  {job.matchHighlights.slice(0, 2).map((h, i) => (
                    <Badge key={i} variant="outline" className="text-[10px] font-normal px-1.5 py-0 bg-white text-gray-500">{h}</Badge>
                  ))}
                  {job.riskPoints.length > 0 && (
                    <span className="text-[10px] text-amber-600 ml-1">{job.riskPoints.length}项风险</span>
                  )}
                </div>
              </div>
            );
          })}
        </div>

        {/* ── Center: Detail ──────────────────────────────────────────── */}
        <div className="flex-1 min-w-0 max-w-[680px] bg-white border border-gray-200 rounded-lg shadow-sm overflow-hidden">
          {selectedJob && (
            <>
              <div className="bg-[#F4F8FF] p-5 border-b border-blue-100">
                <div className="text-[12px] text-gray-400 mb-1">当前选中岗位</div>
                <h2 className="text-lg font-bold text-gray-900 mb-3">{selectedJob.title}</h2>
                <div className="flex flex-wrap items-center gap-2 mb-3 text-xs text-gray-500">
                  <span>{selectedJob.city}</span><span className="text-gray-300">·</span>
                  <span>{selectedJob.category}</span><span className="text-gray-300">·</span>
                  <span>{selectedJob.department}</span>
                </div>
                <div className="bg-white rounded-md p-3 border border-blue-50">
                  <span className="text-xs font-semibold text-gray-900">一句话总结：</span>
                  <span className="text-sm text-gray-700">{selectedJob.oneSentenceReason}</span>
                </div>
              </div>

              <div className="p-5 space-y-5">
                <section>
                  <h3 className="font-bold text-sm text-gray-900 mb-2 flex items-center gap-1.5">
                    <span className="w-1 h-3.5 bg-[#0052D9] rounded-full" /> 匹配亮点
                  </h3>
                  <ul className="space-y-1.5">
                    {selectedJob.matchHighlights.map((h, i) => (
                      <li key={i} className="text-[13px] text-gray-700 flex items-start gap-2">
                        <span className="text-green-500 mt-1 shrink-0">●</span> {h}
                      </li>
                    ))}
                  </ul>
                </section>

                <section>
                  <h3 className="font-bold text-sm text-gray-900 mb-2 flex items-center gap-1.5">
                    <span className="w-1 h-3.5 bg-amber-400 rounded-full" /> 主要风险与提醒
                  </h3>
                  <div className="bg-amber-50/30 border border-amber-100/50 rounded-md p-3">
                    <ul className="space-y-1.5">
                      {selectedJob.riskPoints.map((r, i) => (
                        <li key={i} className="flex items-center gap-2 text-[13px] text-amber-700">
                          <div className="w-1.5 h-1.5 rounded-full bg-amber-400/80 shrink-0" /> {r}
                        </li>
                      ))}
                    </ul>
                  </div>
                </section>

                <section className="bg-gray-50 rounded-md p-4 border border-gray-100">
                  <div className="text-[13px] text-gray-700 leading-relaxed mb-3">
                    <span className="font-medium text-gray-900">建议动作：</span>
                    {selectedJob.suggestedAction}
                  </div>
                  <div className="flex gap-3">
                    <a href={`/jobs/${selectedJob.jobId}`} className="flex-1 text-center px-3 py-2 bg-[#0052D9] text-white text-xs rounded-lg hover:bg-[#0047BD]">
                      进入岗位详情
                    </a>
                    <a href={`/match?jobId=${selectedJob.jobId}&resumeId=${selectedOption.id}`} className="flex-1 text-center px-3 py-2 border border-[#0052D9] text-[#0052D9] text-xs rounded-lg hover:bg-blue-50">
                      进行匹配分析
                    </a>
                  </div>
                </section>
              </div>
            </>
          )}
        </div>

        {/* ── Right sidebar ──────────────────────────────────────────── */}
        <aside className="w-[300px] shrink-0 flex flex-col gap-4">
          <div className="bg-white border border-gray-200 rounded-lg p-4">
            <h3 className="text-[13px] font-medium text-gray-500 mb-2">当前简历</h3>
            <p className="text-sm font-bold text-gray-900 mb-1">{selectedOption.name}</p>
            <p className="text-[11px] text-gray-400">Resume ID: {selectedOption.id}</p>
            {activeResult.profileDominantProfile && (
              <p className="text-[11px] text-[#0052D9] mt-1">画像：{activeResult.profileDominantProfile}</p>
            )}
          </div>

          <div className="bg-white border border-gray-200 rounded-lg p-4">
            <h3 className="text-[13px] font-medium text-gray-500 mb-2">推荐来源</h3>
            <p className="text-xs text-[#0052D9] font-medium">{sourceLabel}</p>
            <p className="text-[11px] text-gray-400 mt-1">生成时间：{nowStr}</p>
          </div>

          <div className="bg-white border border-gray-200 rounded-lg p-4">
            <h3 className="text-[13px] font-medium text-gray-500 mb-2">深度推荐</h3>
            {phase === "deep_loading" ? (
              <div className="flex items-center gap-2 text-xs text-gray-500">
                <Loader2 className="w-3.5 h-3.5 animate-spin" /> 智能体深度分析中…
              </div>
            ) : phase === "deep_done" ? (
              <div className="text-xs text-green-700 flex items-center gap-1.5">
                <RefreshCw className="w-3 h-3" /> 深度分析已完成
                <button onClick={handleDeepRecommendation} className="ml-1 text-[#0052D9] hover:underline">重新生成</button>
              </div>
            ) : phase === "deep_error" ? (
              <button onClick={handleDeepRecommendation} className="w-full flex items-center justify-center gap-1.5 px-3 py-2 bg-[#0052D9] text-white text-sm rounded-lg hover:bg-[#0047BD]">
                <RefreshCw className="w-3.5 h-3.5" /> 重试深度推荐
              </button>
            ) : (
              <button onClick={handleDeepRecommendation} className="w-full flex items-center justify-center gap-1.5 px-3 py-2 bg-[#0052D9] text-white text-sm rounded-lg hover:bg-[#0047BD]">
                <Sparkles className="w-3.5 h-3.5" /> 生成深度推荐
              </button>
            )}
          </div>

          <div className="bg-white border border-gray-200 rounded-lg p-4">
            <h3 className="text-[13px] font-medium text-gray-500 mb-2">推荐记录</h3>
            <p className="text-[11px] text-gray-400 leading-relaxed">推荐记录将在生成后用于后续回看。</p>
          </div>
        </aside>
      </div>

      {/* ── Deep loading global overlay ─────────────────────────────── */}
      {phase === "deep_loading" && (
        <div className="fixed inset-0 z-50 bg-slate-900/20 backdrop-blur-md flex items-center justify-center">
          <div className="bg-white rounded-3xl shadow-2xl p-8 max-w-md mx-4 text-center">
            <Loader2 className="w-10 h-10 animate-spin text-[#0052D9] mx-auto mb-4" />
            <h2 className="text-lg font-bold text-gray-900 mb-2">智能体正在深度分析岗位匹配</h2>
            <p className="text-sm text-gray-500 leading-relaxed mb-4">
              系统正在结合你的候选人标签、岗位标签和匹配证据，生成更高质量的推荐结果。
            </p>
            <div className="text-xs text-gray-400 space-y-1">
              <div className="flex items-center gap-2"><Loader2 className="w-3 h-3 animate-spin" /> 读取简历画像标签</div>
              <div className="flex items-center gap-2"><span className="w-3 h-3 rounded-full bg-gray-200" /> 匹配岗位标签矩阵</div>
              <div className="flex items-center gap-2"><span className="w-3 h-3 rounded-full bg-gray-200" /> 生成推荐理由和风险点</div>
              <div className="flex items-center gap-2"><span className="w-3 h-3 rounded-full bg-gray-200" /> 审查推荐结果一致性</div>
            </div>
          </div>
        </div>
      )}
    </ContentPageShell>
  );
}
