import React, { useState, useEffect } from "react";
import { Link, useSearchParams, useNavigate, useLocation } from "react-router";
import { loadLatestOptimizationV2ViewModel, loadGenerateOptimizationV2ViewModel, loadResumeOptionViewModelsFromApi, loadResumeVersionDetail } from "../../api";
import { getJobDetail } from "../../api/jobs";
import type { JobApiItem } from "../../api/jobs";
import type { ResumeOptionViewModel } from "../../adapters";
import type { OptimizationV2ViewModel } from "../../api/viewModels";
import type { Resume } from "../../types";
import type { OptimizationMode, GenerationTier } from "../../types";
import type { OptimizationFixItem, OptimizationRewriteExample, OptimizationLongTermItem } from "../../types";
import { ContentPageShell } from "../components/layout/ContentPageShell";
import { getAuthRedirectIf401 } from "../utils/authRedirect";
import { PageLoading } from "../components/PageState";
import { AgentStepLoadingOverlay } from "../components/common/AgentStepLoadingOverlay";
import { Sparkles, RefreshCw, Loader2, Target, Info, FileEdit, AlertTriangle, ArrowUp } from "lucide-react";

// ─── Phase ─────────────────────────────────────────────────────────────

type OptPhase =
  | "loading_latest"
  | "fast_loading"
  | "fast_done"
  | "fast_error"
  | "deep_loading"
  | "deep_done"
  | "deep_error";

// ─── Header Label ──────────────────────────────────────────────────────

function getHeaderLabel(phase: OptPhase, tier: string, genBy: string): string {
  if (phase === "loading_latest") return "正在读取优化结果";
  if (phase === "fast_loading") return "正在生成快速优化建议";
  if (phase === "deep_loading") return "智能体正在深度优化";
  if (phase === "fast_done") return "快速优化 · 规则分析";
  if (phase === "deep_done") {
    if (genBy?.includes("llm")) return "深度优化 · AI生成";
    return "深度优化 · 规则兜底";
  }
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

  // Derived mode
  const parsedResumeId = parseResumeIdFromQuery(resumeIdStr);
  const parsedJobId = jobIdStr ? parseJobIdFromQuery(jobIdStr) : undefined;
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

  // K9-3-E: phase + active result
  const [phase, setPhase] = useState<OptPhase>("loading_latest");
  const [activeResult, setActiveResult] = useState<OptimizationV2ViewModel | null>(null);
  const [activeTier, setActiveTier] = useState<string>("fast");
  const [activeGeneratedBy, setActiveGeneratedBy] = useState<string>("");
  const [optError, setOptError] = useState<string | null>(null);

  // ── URL validation errors ──
  if (!parsedResumeId) {
    return (
      <div className="max-w-[1440px] mx-auto px-6 py-12 text-center">
        <p className="text-gray-500 text-sm mb-3">缺少有效的简历 ID</p>
        <p className="text-gray-400 text-xs mb-4">请通过简历管理页面选择一份简历后再进行优化。</p>
        <Link to="/resumes" className="text-[#0052D9] text-sm underline">管理我的简历</Link>
      </div>
    );
  }
  if (derivedMode === "target_job" && !parsedJobId) {
    return (
      <div className="max-w-[1440px] mx-auto px-6 py-12 text-center">
        <p className="text-gray-500 text-sm mb-3">缺少有效的岗位 ID</p>
        <Link to="/jobs" className="text-[#0052D9] text-sm underline">返回岗位列表</Link>
      </div>
    );
  }
  if (derivedMode === "target_direction" && !targetDirection) {
    return (
      <div className="max-w-[1440px] mx-auto px-6 py-12 text-center">
        <p className="text-gray-500 text-sm mb-3">缺少目标方向参数</p>
        <Link to="/recommend" className="text-[#0052D9] text-sm underline">返回岗位推荐</Link>
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

  // ── Latest-first entry ──
  useEffect(() => {
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
          // No result — auto-generate Fast
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
        setPhase("deep_done");
      })
      .catch(() => setPhase("deep_error"));
  };

  // ── Guards ──
  if (loadingJob || loadingOptions) {
    return <PageLoading message="正在加载岗位和简历信息…" />;
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
        <Link to="/resumes" className="text-[#0052D9] text-sm underline">管理我的简历</Link>
      </div>
    );
  }

  if (phase === "loading_latest" || phase === "fast_loading") {
    return <PageLoading message="正在加载优化建议…" />;
  }

  if (phase === "fast_error" && !activeResult) {
    return (
      <div className="max-w-[1440px] mx-auto px-6 py-12 text-center">
        <p className="text-gray-500 text-sm mb-3">{optError || "优化接口暂时不可用"}</p>
        <Link to="/resumes" className="text-[#0052D9] text-sm underline">管理我的简历</Link>
      </div>
    );
  }

  const isFast = activeTier !== "quality";
  const headerLabel = getHeaderLabel(phase, activeTier, activeGeneratedBy);
  const modeLabel = getModeLabel(derivedMode, targetDirection || undefined, job?.title);

  return (
    <ContentPageShell maxWidthClassName="max-w-[1440px]" contentClassName="pb-12">

      {/* Agent Overlay for deep loading */}
      <AgentStepLoadingOverlay
        open={phase === "deep_loading"}
        title="正在深度优化简历"
        description="Offerr鹅正在分析简历画像、目标方向和优化缺口，生成完整改写建议。"
        steps={[
          { key: "profile", label: "读取简历画像" },
          { key: "analyze", label: "分析目标方向" },
          { key: "gaps", label: "定位优化缺口" },
          { key: "suggestions", label: "生成改写建议" },
        ]}
        status="loading"
      />

      {/* Header */}
      <div className="mb-6">
        <h1 className="text-[22px] font-semibold text-gray-900 flex items-center gap-2">
          <Sparkles className="w-5 h-5 text-[#0052D9]" />
          简历优化
        </h1>
        <div className="flex flex-wrap items-center gap-3 mt-3 text-xs text-gray-500">
          <span className="text-[#0052D9] font-medium">{headerLabel}</span>
          {activeResult && (
            <>
              <span className="text-gray-300">|</span>
              <span>{modeLabel}</span>
              <span className="text-gray-300">|</span>
              <span>简历：{selectedResumeDetail?.real_name || selectedOption.name}</span>
            </>
          )}
        </div>
      </div>

      {activeResult && (
        <div className="flex gap-6 items-start">
          {/* ── Center content: Fast vs Deep ── */}
          {isFast ? (
            <FastView data={activeResult} phase={phase} />
          ) : (
            <DeepView data={activeResult} phase={phase} />
          )}

          {/* ── Right sidebar ── */}
          <RightSidebar
            resumeName={selectedOption.name}
            profileDominantProfile={activeResult.profileDominantProfile}
            optimizationMode={derivedMode}
            targetDirection={targetDirection}
            jobTitle={job?.title}
            phase={phase}
            isFast={isFast}
            onDeepOptimization={handleDeepOptimization}
          />
        </div>
      )}
    </ContentPageShell>
  );
}

// ─── Fast View ──────────────────────────────────────────────────────────

function FastView({ data, phase }: { data: OptimizationV2ViewModel; phase: OptPhase }) {
  const levelColor = getLevelColor(data.optimizationLevel);
  return (
    <div className="flex-1 min-w-0 max-w-[720px] bg-white border border-gray-200 rounded-lg shadow-sm overflow-hidden">
      {/* Score & level */}
      <div className="bg-[#F4F8FF] p-5 border-b border-blue-100">
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

      <div className="p-5 space-y-4">
        {/* Quick Fix Items */}
        {data.quickFixItems && data.quickFixItems.length > 0 && (
          <section>
            <h3 className="font-bold text-sm text-gray-900 mb-2 flex items-center gap-1.5">
              <FileEdit className="w-3.5 h-3.5 text-[#0052D9]" />
              快速修改项
            </h3>
            <ul className="space-y-2">
              {data.quickFixItems.slice(0, 5).map((item, i) => (
                <li key={i} className="text-[13px] text-gray-700 flex items-start gap-2 bg-gray-50 p-2.5 rounded-md border border-gray-100">
                  <span className="text-[#0052D9] font-medium shrink-0 mt-0.5">{i + 1}.</span>
                  <div>
                    <span className="font-medium text-gray-800">{item.section}：</span>
                    <span>{item.issue}</span>
                    <span className="text-gray-500 ml-1">→ {item.action}</span>
                  </div>
                </li>
              ))}
            </ul>
          </section>
        )}

        {/* Risk Points */}
        {data.riskPoints && data.riskPoints.length > 0 && (
          <section>
            <h3 className="font-bold text-sm text-amber-600 mb-2 flex items-center gap-1.5">
              <AlertTriangle className="w-3.5 h-3.5" />
              主要风险
            </h3>
            <ul className="space-y-1">
              {data.riskPoints.slice(0, 2).map((r, i) => (
                <li key={i} className="text-[13px] text-gray-700 flex items-start gap-2">
                  <span className="text-amber-500 mt-1 shrink-0">●</span> {r}
                </li>
              ))}
            </ul>
          </section>
        )}

        {/* Delivery Focus */}
        {data.deliveryFocus && (
          <section>
            <h3 className="font-bold text-sm text-gray-900 mb-2 flex items-center gap-1.5">
              <ArrowUp className="w-3.5 h-3.5 text-[#0052D9]" />
              投递重点
            </h3>
            <p className="text-[13px] text-gray-700 bg-blue-50/50 p-3 rounded-md border border-blue-100">{data.deliveryFocus}</p>
          </section>
        )}

        {/* Upsell */}
        <div className="bg-blue-50/50 border border-blue-100 rounded-md p-3 mt-4">
          <p className="text-[12px] text-blue-700">
            当前为快速优化结果，适合先判断最该改的几处。生成深度优化后可查看改写示例和完整优化路径。
          </p>
        </div>
      </div>
    </div>
  );
}

// ─── Deep View ──────────────────────────────────────────────────────────

function DeepView({ data, phase }: { data: OptimizationV2ViewModel; phase: OptPhase }) {
  const levelColor = getLevelColor(data.optimizationLevel);

  // Section cards for each fix category
  const fixSections: { title: string; icon: string; items: OptimizationFixItem[] }[] = [
    { title: "快速修改项", icon: "⚡", items: data.quickFixItems },
    { title: "结构优化", icon: "🏗️", items: data.structureFixItems },
    { title: "内容表达优化", icon: "✏️", items: data.contentFixItems },
    { title: "关键词补强", icon: "🔑", items: data.keywordFixItems },
    { title: "证据补强", icon: "📋", items: data.evidenceFixItems },
  ].filter(s => s.items.length > 0);

  return (
    <div className="flex-1 min-w-0 max-w-[780px] bg-white border border-gray-200 rounded-lg shadow-sm overflow-hidden">
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
        {/* Fix item categories */}
        {fixSections.map((section, si) => (
          <section key={si}>
            <h3 className="font-bold text-sm text-gray-900 mb-3 flex items-center gap-1.5">
              <span>{section.icon}</span> {section.title}
            </h3>
            <div className="space-y-2">
              {section.items.map((item, i) => (
                <div key={i} className="text-[13px] text-gray-700 flex items-start gap-2 bg-gray-50 p-2.5 rounded-md border border-gray-100">
                  <span className="text-[#0052D9] font-medium shrink-0 mt-0.5">{i + 1}.</span>
                  <div className="flex-1 min-w-0">
                    <span className="font-medium text-gray-800">{item.section}：</span>
                    <span>{item.issue}</span>
                    {item.action && (
                      <span className="text-gray-500 ml-1">→ {item.action}</span>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </section>
        ))}

        {/* Risk points */}
        {data.riskPoints && data.riskPoints.length > 0 && (
          <section>
            <h3 className="font-bold text-sm text-amber-600 mb-2 flex items-center gap-1.5">
              <AlertTriangle className="w-3.5 h-3.5" />
              风险提醒
            </h3>
            <ul className="space-y-1.5">
              {data.riskPoints.map((r, i) => (
                <li key={i} className="text-[13px] text-gray-700 flex items-start gap-2 bg-amber-50/50 p-2.5 rounded-md border border-amber-100">
                  <span className="text-amber-500 mt-1 shrink-0">●</span> {r}
                </li>
              ))}
            </ul>
          </section>
        )}

        {/* ── Rewrite Examples (Visual Highlight) ── */}
        {data.rewriteExamples && data.rewriteExamples.length > 0 && (
          <section>
            <h3 className="font-bold text-base text-gray-900 mb-4 flex items-center gap-2">
              <Sparkles className="w-4 h-4 text-[#0052D9]" />
              改写示例
            </h3>
            <div className="space-y-4">
              {data.rewriteExamples.map((ex, i) => (
                <div key={i} className="border border-blue-200 rounded-lg overflow-hidden bg-white shadow-sm">
                  <div className="bg-blue-50/50 px-4 py-2 border-b border-blue-100 flex items-center gap-2">
                    <span className="text-xs font-bold text-[#0052D9]">{ex.section}</span>
                    <span className="text-xs text-gray-500">{ex.reason}</span>
                  </div>
                  <div className="grid grid-cols-2 divide-x divide-gray-100">
                    <div className="p-4 bg-gray-50/50">
                      <div className="text-[11px] font-medium text-gray-400 mb-1.5">修改前</div>
                      <p className="text-[13px] text-gray-600 whitespace-pre-wrap leading-relaxed">{ex.before}</p>
                    </div>
                    <div className="p-4 bg-blue-50/20">
                      <div className="text-[11px] font-medium text-[#0052D9] mb-1.5">修改后</div>
                      <p className="text-[13px] text-gray-800 whitespace-pre-wrap leading-relaxed">{ex.after}</p>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </section>
        )}

        {/* Delivery Focus */}
        {data.deliveryFocus && (
          <section>
            <h3 className="font-bold text-sm text-gray-900 mb-2 flex items-center gap-1.5">
              <ArrowUp className="w-3.5 h-3.5 text-[#0052D9]" />
              投递重点
            </h3>
            <p className="text-[13px] text-gray-700 bg-blue-50/50 p-3 rounded-md border border-blue-100">{data.deliveryFocus}</p>
          </section>
        )}

        {/* Long Term Improvements */}
        {data.longTermImprovements && data.longTermImprovements.length > 0 && (
          <section>
            <h3 className="font-bold text-sm text-gray-900 mb-2 flex items-center gap-1.5">
              <Info className="w-3.5 h-3.5 text-gray-400" />
              长期补强建议
            </h3>
            <ul className="space-y-2">
              {data.longTermImprovements.map((item, i) => (
                <li key={i} className="text-[13px] text-gray-700 flex items-start gap-2 bg-gray-50 p-2.5 rounded-md border border-gray-100">
                  <span className="text-gray-400 mt-1 shrink-0">●</span>
                  <div>
                    <span className="font-medium text-gray-800">{item.direction}</span>
                    {item.note && <span className="text-gray-500"> — {item.note}</span>}
                  </div>
                </li>
              ))}
            </ul>
          </section>
        )}
      </div>
    </div>
  );
}

// ─── Right Sidebar ──────────────────────────────────────────────────────

function RightSidebar({
  resumeName,
  profileDominantProfile,
  optimizationMode,
  targetDirection,
  jobTitle,
  phase,
  isFast,
  onDeepOptimization,
}: {
  resumeName: string;
  profileDominantProfile: string;
  optimizationMode: string;
  targetDirection?: string | null;
  jobTitle?: string;
  phase: OptPhase;
  isFast: boolean;
  onDeepOptimization: () => void;
}) {
  return (
    <aside className="w-[300px] shrink-0 flex flex-col gap-4">
      {/* Resume */}
      <div className="bg-white border border-gray-200 rounded-lg p-4">
        <h3 className="text-[13px] font-medium text-gray-500 mb-2">当前简历</h3>
        <p className="text-sm font-bold text-gray-900 mb-1">{resumeName}</p>
        {profileDominantProfile && (
          <p className="text-[11px] text-[#0052D9]">画像：{profileDominantProfile}</p>
        )}
      </div>

      {/* Optimization Mode */}
      <div className="bg-white border border-gray-200 rounded-lg p-4">
        <h3 className="text-[13px] font-medium text-gray-500 mb-2">优化模式</h3>
        <p className="text-sm font-bold text-gray-900">
          {optimizationMode === "target_job" ? "针对目标岗位" :
           optimizationMode === "target_direction" ? "针对目标方向" :
           "通用优化"}
        </p>
        {optimizationMode === "target_direction" && targetDirection && (
          <p className="text-[12px] text-gray-500 mt-1">方向：{targetDirection}</p>
        )}
        {optimizationMode === "target_job" && jobTitle && (
          <p className="text-[12px] text-gray-500 mt-1">岗位：{jobTitle}</p>
        )}
      </div>

      {/* Deep Optimization Button */}
      <div className="bg-white border border-gray-200 rounded-lg p-4">
        <h3 className="text-[13px] font-medium text-gray-500 mb-2">深度优化</h3>
        <p className="text-[11px] text-gray-400 mb-3">
          {isFast
            ? "基于简历画像和目标方向，生成完整的优化建议、改写示例和长期补强计划。"
            : "当前已展示深度优化结果，可重新生成。"}
        </p>
        {phase === "deep_loading" ? (
          <div className="flex items-center gap-2 text-xs text-gray-500">
            <Loader2 className="w-3.5 h-3.5 animate-spin" /> 智能体深度优化中…
          </div>
        ) : phase === "deep_done" || !isFast ? (
          <div className="text-xs text-green-700 flex flex-col gap-1.5">
            <span className="flex items-center gap-1.5">
              <RefreshCw className="w-3 h-3" /> 深度优化已完成
            </span>
            <button onClick={onDeepOptimization} className="text-[#0052D9] hover:underline ml-5">重新生成</button>
          </div>
        ) : phase === "deep_error" ? (
          <button onClick={onDeepOptimization} className="w-full flex items-center justify-center gap-1.5 px-3 py-2 bg-[#0052D9] text-white text-sm rounded-lg hover:bg-[#0047BD]">
            <RefreshCw className="w-3.5 h-3.5" /> 重试深度优化
          </button>
        ) : (
          <button onClick={onDeepOptimization} className="w-full flex items-center justify-center gap-1.5 px-3 py-2 bg-[#0052D9] text-white text-sm rounded-lg hover:bg-[#0047BD]">
            <Sparkles className="w-3.5 h-3.5" /> 生成深度优化
          </button>
        )}
      </div>
    </aside>
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
