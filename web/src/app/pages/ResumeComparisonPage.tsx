import { useState, useEffect, useRef } from "react";
import { Link, useSearchParams, useNavigate, useLocation } from "react-router";
import { Loader2, AlertCircle, Sparkles, RefreshCw, ArrowRight } from "lucide-react";
import { loadLatestCompareV2ViewModel, loadCompareV2ViewModel } from "../../api";
import type { CompareV2ViewModel } from "../../api/viewModels";
import { ContentPageShell } from "../components/layout/ContentPageShell";
import { PageLoading, PageError, PageEmpty } from "../components/PageState";
import { AgentStepLoadingOverlay, type AgentLoadingStep } from "../components/common/AgentStepLoadingOverlay";
import { getAuthRedirectIf401 } from "../utils/authRedirect";
import { getResumes } from "../../api/resumes";
import type { Resume } from "../../types";

// ─── Helpers ───────────────────────────────────────────────────────────

function getMode(urlJobId: string | null, urlDir: string | null): { mode: string; dir: string; jobId: number | null } {
  if (urlJobId && /^\d+$/.test(urlJobId)) return { mode: "target_job", dir: "", jobId: Number(urlJobId) };
  if (urlDir && urlDir.trim()) return { mode: "target_direction", dir: urlDir.trim(), jobId: null };
  return { mode: "general", dir: "", jobId: null };
}

function getHeaderLabel(phase: ComparePhase, genBy: string): string {
  switch (phase) {
    case "loading_latest": return "正在读取对比结果…";
    case "fast_loading": return "正在生成快速对比…";
    case "fast_done": return "快速对比 · 规则分析";
    case "deep_loading": return "智能体正在深度对比…";
    case "deep_done": return genBy?.startsWith("llm") ? "深度对比 · AI生成" : "深度对比 · 规则兜底";
    case "deep_error": return "深度对比生成失败 · 已保留当前结果";
    case "fast_error": return "快速对比生成失败";
    default: return "";
  }
}

type ComparePhase =
  | "loading_latest" | "fast_loading" | "fast_done" | "fast_error"
  | "deep_loading" | "deep_done" | "deep_error";

const deepSteps: AgentLoadingStep[] = [
  { key: "profiles", label: "读取两版简历画像" },
  { key: "diff", label: "分析关键差异" },
  { key: "winner", label: "判断更优版本" },
  { key: "merge", label: "生成合并与改写建议" },
];

// ─── Main ───────────────────────────────────────────────────────────────

export function ResumeComparisonPage() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const location = useLocation();
  const fromParam = searchParams.get("fromResumeId");
  const toParam = searchParams.get("toResumeId");
  const dirParam = searchParams.get("targetDirection");
  const jobParam = searchParams.get("jobId");

  const { mode, dir, jobId } = getMode(jobParam, dirParam);

  const [fromResumeId, setFromResumeId] = useState<number | null>(fromParam && /^\d+$/.test(fromParam) ? Number(fromParam) : null);
  const [toResumeId, setToResumeId] = useState<number | null>(toParam && /^\d+$/.test(toParam) ? Number(toParam) : null);
  const [resumeLabels, setResumeLabels] = useState<Record<number, string>>({});
  const [loadingResumes, setLoadingResumes] = useState(true);

  const [phase, setPhase] = useState<ComparePhase>("loading_latest");
  const [fastResult, setFastResult] = useState<CompareV2ViewModel | null>(null);
  const [deepResult, setDeepResult] = useState<CompareV2ViewModel | null>(null);
  const [errorMsg, setErrorMsg] = useState("");
  const [deepErrMsg, setDeepErrMsg] = useState("");
  const [deepOverlayStatus, setDeepOverlayStatus] = useState<"loading" | "success" | "error">("loading");

  // ── Load resume labels ────────────────────────────────────────────
  useEffect(() => {
    getResumes()
      .then((list: Resume[]) => {
        const map: Record<number, string> = {};
        list.forEach((r) => { map[r.id] = r.name || `简历 ${r.id}`; });
        setResumeLabels(map);
      })
      .catch((err) => {
        const r = getAuthRedirectIf401(err, location.pathname, location.search);
        if (r) { navigate(r, { replace: true }); return; }
      })
      .finally(() => setLoadingResumes(false));
  }, []);

  // ── Validate params ───────────────────────────────────────────────
  useEffect(() => {
    if (loadingResumes) return;
    if (!fromResumeId || !toResumeId) return;
    if (fromResumeId === toResumeId) { setErrorMsg("不能对比相同的简历。"); setPhase("fast_error"); return; }
    loadLatest();
  }, [fromResumeId, toResumeId, loadingResumes]);

  async function loadLatest() {
    if (!fromResumeId || !toResumeId) return;
    setPhase("loading_latest"); setErrorMsg("");
    try {
      const latest = await loadLatestCompareV2ViewModel({
        fromResumeId, toResumeId,
        comparisonMode: mode,
        targetDirection: dir,
        jobId: jobId ?? undefined,
      });
      if (latest.hasResult && latest.result) {
        const tier = latest.generationTier === "quality" ? "quality" : "fast";
        if (tier === "quality") {
          setDeepResult(latest.result); setPhase("deep_done");
        } else {
          setFastResult(latest.result); setPhase("fast_done");
        }
        return;
      }
    } catch { /* fall through to generate */ }
    generateFast();
  }

  function generateFast() {
    if (!fromResumeId || !toResumeId) return;
    setPhase("fast_loading"); setErrorMsg("");
    loadCompareV2ViewModel({
      fromResumeId, toResumeId,
      comparisonMode: mode,
      generationTier: "fast",
      targetDirection: dir,
      jobId: jobId ?? undefined,
    })
      .then((vm) => { setFastResult(vm); setPhase("fast_done"); })
      .catch(() => { setPhase("fast_error"); setErrorMsg("快速对比生成失败，请稍后重试。"); });
  }

  async function handleDeep() {
    if (!fromResumeId || !toResumeId) return;
    setPhase("deep_loading"); setDeepErrMsg(""); setDeepOverlayStatus("loading");
    try {
      const vm = await loadCompareV2ViewModel({
        fromResumeId, toResumeId,
        comparisonMode: mode,
        generationTier: "quality",
        targetDirection: dir,
        jobId: jobId ?? undefined,
      });
      setDeepResult(vm); setDeepOverlayStatus("success");
    } catch {
      setDeepOverlayStatus("error"); setPhase("deep_error");
      setDeepErrMsg("深度对比暂时生成失败，当前快速对比结果仍可参考。");
    }
  }

  const activeResult: CompareV2ViewModel | null = phase === "deep_done" && deepResult ? deepResult : fastResult;
  const isDeep = phase === "deep_done";
  const headerLabel = getHeaderLabel(phase, activeResult?.generatedBy ?? "");
  const fromLabel = fromResumeId ? resumeLabels[fromResumeId] || `简历 ${fromResumeId}` : "?";
  const toLabel = toResumeId ? resumeLabels[toResumeId] || `简历 ${toResumeId}` : "?";

  // ── States ─────────────────────────────────────────────────────────
  if (loadingResumes) return <PageLoading message="正在加载简历信息…" />;
  if (!fromResumeId || !toResumeId) {
    return <PageError title="请选择要对比的两份简历" description="URL 需要 fromResumeId 和 toResumeId 参数。">
      <Link to="/resumes" className="text-[#0052D9] text-sm underline">返回我的简历</Link>
    </PageError>;
  }
  if (fromResumeId === toResumeId) {
    return <PageError title="不能对比相同的简历" description="请选择两份不同的简历。">
      <Link to="/resumes" className="text-[#0052D9] text-sm underline">返回我的简历</Link>
    </PageError>;
  }
  if (phase === "loading_latest") return <PageLoading message="正在读取对比结果…" />;
  if (phase === "fast_loading") {
    return <div className="max-w-[1440px] mx-auto px-6 py-32 text-center">
      <Loader2 className="w-10 h-10 animate-spin text-[#0052D9] mx-auto mb-4" />
      <h2 className="text-lg font-bold text-gray-900 mb-2">正在生成快速对比</h2>
      <p className="text-sm text-gray-500 max-w-md mx-auto">系统正在对比两份简历的结构、方向和关键差异。</p>
    </div>;
  }
  if (phase === "fast_error") {
    return <div className="max-w-[1440px] mx-auto px-6 py-12 text-center">
      <AlertCircle className="w-8 h-8 text-amber-500 mx-auto mb-3" />
      <p className="text-gray-500 text-sm mb-3">{errorMsg}</p>
      <button onClick={generateFast} className="text-[#0052D9] text-sm underline">重新尝试</button>
    </div>;
  }
  if (!activeResult) {
    return <PageEmpty title="暂无对比结果" description="请稍后重试。">
      <Link to="/resumes" className="text-[#0052D9] text-sm underline">返回我的简历</Link>
    </PageEmpty>;
  }

  const V = activeResult;

  return (
    <ContentPageShell maxWidthClassName="max-w-[1440px]" contentClassName="pb-12">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-[22px] font-semibold text-gray-900 flex items-center gap-2"><Sparkles className="w-5 h-5 text-[#0052D9]" />简历版本对比</h1>
        <p className="mt-1.5 text-sm text-gray-500">
          从 <span className="font-medium text-gray-700">{fromLabel}</span> 到 <span className="font-medium text-gray-700">{toLabel}</span>
          {mode !== "general" && <span className="ml-1">（{mode === "target_direction" ? `方向: ${dir}` : `岗位 ID: ${jobId}`}）</span>}
        </p>
        <div className="flex flex-wrap items-center gap-3 mt-3 text-xs text-gray-500">
          <span className="flex items-center gap-1.5 text-[#0052D9] font-medium">{headerLabel}</span>
          <span className="text-gray-300">|</span><span>对比分: {V.compareScore}</span>
          {V.fromProfileDominantProfile && <><span className="text-gray-300">|</span><span>画像: {V.fromProfileDominantProfile} → {V.toProfileDominantProfile}</span></>}
        </div>
        {!isDeep && <p className="mt-2 text-[11px] text-gray-400">当前为快速对比结果，点击"生成深度对比"可查看合并策略和改写示例。</p>}
      </div>

      <div className="flex gap-6 items-start">
        {/* Left: Winner summary */}
        <div className="w-[340px] shrink-0">
          <div className="bg-white border border-gray-200 rounded-lg p-5 shadow-sm space-y-4">
            <div className="text-xs text-gray-400">对比结论</div>
            <div className="text-2xl font-bold text-[#0052D9]">{V.compareScore}<span className="text-sm font-normal text-gray-400"> /100</span></div>
            <div className="text-sm font-semibold text-gray-900">{V.compareLevel}</div>
            <div className="bg-[#F4F8FF] rounded-md p-3 border border-blue-100">
              <span className="text-[11px] font-semibold text-[#0052D9]">推荐使用：</span>
              <span className="text-sm text-gray-800 ml-1">{V.winnerResumeId === fromResumeId ? fromLabel : toLabel}</span>
              <p className="text-[11px] text-gray-500 mt-1">{V.winnerLabel}</p>
            </div>
            <p className="text-[13px] text-gray-600 leading-relaxed">{V.oneSentenceSummary}</p>
            {V.deliveryAdvice && <div className="bg-blue-50/50 rounded-md p-3 text-[12px] text-blue-700">{V.deliveryAdvice}</div>}
          </div>
        </div>

        {/* Center: Detail */}
        <div className="flex-1 min-w-0 space-y-4">
          {/* Key Differences */}
          {V.keyDifferences.length > 0 && (
            <section className="bg-white border border-gray-200 rounded-lg p-5 shadow-sm">
              <h3 className="font-bold text-sm text-gray-900 mb-3">关键差异</h3>
              <ul className="space-y-2">{V.keyDifferences.map((d, i) => <li key={i} className="flex items-start gap-2 text-[13px] text-gray-700"><span className="text-[#0052D9] mt-1 shrink-0">●</span>{d}</li>)}</ul>
            </section>
          )}

          {/* Strengths */}
          <div className="grid grid-cols-2 gap-4">
            <section className="bg-white border border-gray-200 rounded-lg p-5 shadow-sm">
              <h3 className="font-bold text-sm text-gray-900 mb-3">{fromLabel} 优势</h3>
              <ul className="space-y-1.5">{V.fromStrengths.map((s, i) => <li key={i} className="text-[12px] text-gray-600">• {s}</li>)}</ul>
            </section>
            <section className="bg-white border border-gray-200 rounded-lg p-5 shadow-sm">
              <h3 className="font-bold text-sm text-gray-900 mb-3">{toLabel} 优势</h3>
              <ul className="space-y-1.5">{V.toStrengths.map((s, i) => <li key={i} className="text-[12px] text-gray-600">• {s}</li>)}</ul>
            </section>
          </div>

          {/* Risks */}
          <div className="grid grid-cols-2 gap-4">
            {(V.fromRisks.length > 0 || V.toRisks.length > 0) && (
              <section className="bg-amber-50/50 border border-amber-100 rounded-lg p-5">
                <h3 className="font-bold text-sm text-amber-800 mb-2">{fromLabel} 风险</h3>
                <ul className="space-y-1">{V.fromRisks.map((r, i) => <li key={i} className="text-[12px] text-amber-700">• {r}</li>)}</ul>
              </section>
            )}
            {(V.fromRisks.length > 0 || V.toRisks.length > 0) && (
              <section className="bg-amber-50/50 border border-amber-100 rounded-lg p-5">
                <h3 className="font-bold text-sm text-amber-800 mb-2">{toLabel} 风险</h3>
                <ul className="space-y-1">{V.toRisks.map((r, i) => <li key={i} className="text-[12px] text-amber-700">• {r}</li>)}</ul>
              </section>
            )}
          </div>

          {/* Deep sections */}
          {isDeep && (
            <>
              {V.keepFromItems.length > 0 && (
                <section className="bg-white border border-gray-200 rounded-lg p-5 shadow-sm">
                  <h3 className="font-bold text-sm text-gray-900 mb-2">建议保留 {fromLabel} 的内容</h3>
                  <ul className="space-y-1">{V.keepFromItems.map((k, i) => <li key={i} className="text-[12px] text-gray-600">• {k}</li>)}</ul>
                </section>
              )}
              {V.keepToItems.length > 0 && (
                <section className="bg-white border border-gray-200 rounded-lg p-5 shadow-sm">
                  <h3 className="font-bold text-sm text-gray-900 mb-2">建议保留 {toLabel} 的内容</h3>
                  <ul className="space-y-1">{V.keepToItems.map((k, i) => <li key={i} className="text-[12px] text-gray-600">• {k}</li>)}</ul>
                </section>
              )}
              {V.mergeSuggestions.length > 0 && (
                <section className="bg-green-50/50 border border-green-200 rounded-lg p-5">
                  <h3 className="font-bold text-sm text-green-900 mb-3 flex items-center gap-1.5"><span className="w-1 h-3.5 bg-green-500 rounded-full" />合并建议</h3>
                  <ul className="space-y-2">{V.mergeSuggestions.map((m, i) => <li key={i} className="flex items-start gap-2 text-[13px] text-green-800"><span className="text-green-500 mt-0.5">→</span>{m}</li>)}</ul>
                </section>
              )}
              {V.nextVersionSuggestions.length > 0 && (
                <section className="bg-white border border-gray-200 rounded-lg p-5 shadow-sm">
                  <h3 className="font-bold text-sm text-gray-900 mb-3 flex items-center gap-1.5"><span className="w-1 h-3.5 bg-[#0052D9] rounded-full" />下一版建议</h3>
                  <ul className="space-y-1.5">{V.nextVersionSuggestions.map((n, i) => <li key={i} className="text-[13px] text-gray-700">• {n}</li>)}</ul>
                </section>
              )}
              {V.rewriteExamples.length > 0 && (
                <section className="bg-white border border-gray-200 rounded-lg p-5 shadow-sm">
                  <h3 className="font-bold text-sm text-gray-900 mb-3 flex items-center gap-1.5"><span className="w-1 h-3.5 bg-purple-500 rounded-full" />改写示例</h3>
                  <div className="space-y-3">
                    {V.rewriteExamples.map((ex, i) => (
                      <div key={i} className="bg-gray-50 rounded-lg p-3 border border-gray-100">
                        <div className="text-[11px] text-purple-600 font-medium mb-1">{ex.section}</div>
                        <div className="grid grid-cols-2 gap-3 text-[12px]">
                          <div>
                            <span className="text-gray-400 text-[10px]">原版本</span>
                            <p className="text-gray-500 mt-0.5">{ex.fromVersion}</p>
                          </div>
                          <div>
                            <span className="text-gray-400 text-[10px]">目标版本</span>
                            <p className="text-gray-700 mt-0.5 font-medium">{ex.toVersion}</p>
                          </div>
                        </div>
                        {ex.suggestedNext && <p className="text-[11px] text-blue-600 mt-2 border-t border-gray-200 pt-2">下一版建议: {ex.suggestedNext}</p>}
                        {ex.reason && <p className="text-[10px] text-gray-400 mt-1">原因: {ex.reason}</p>}
                      </div>
                    ))}
                  </div>
                </section>
              )}
            </>
          )}
        </div>

        {/* Right sidebar */}
        <aside className="w-[280px] shrink-0 flex flex-col gap-4">
          <div className="bg-white border border-gray-200 rounded-lg p-4">
            <h3 className="text-[13px] font-medium text-gray-500 mb-2">对比信息</h3>
            <p className="text-xs text-gray-400">From: {fromLabel} (ID:{fromResumeId})</p>
            <p className="text-xs text-gray-400 mt-1">To: {toLabel} (ID:{toResumeId})</p>
            {V.fromProfileDominantProfile && <p className="text-xs text-[#0052D9] mt-1">画像: {V.fromProfileDominantProfile} → {V.toProfileDominantProfile}</p>}
          </div>
          <div className="bg-white border border-gray-200 rounded-lg p-4">
            <h3 className="text-[13px] font-medium text-gray-500 mb-2">深度对比</h3>
            <p className="text-[11px] text-gray-400 mb-3">完整分析差异、保留项和合并策略。</p>
            {phase === "deep_loading" ? (
              <div className="flex items-center gap-2 text-xs text-gray-500"><Loader2 className="w-3.5 h-3.5 animate-spin" /> 智能体深度分析中…</div>
            ) : isDeep ? (
              <div className="text-xs text-green-700 flex items-center gap-1.5"><RefreshCw className="w-3 h-3" /> 深度分析已完成<button onClick={handleDeep} className="ml-1 text-[#0052D9] hover:underline">重新生成</button></div>
            ) : phase === "deep_error" ? (
              <button onClick={handleDeep} className="w-full flex items-center justify-center gap-1.5 px-3 py-2 bg-[#0052D9] text-white text-sm rounded-lg hover:bg-[#0047BD]"><RefreshCw className="w-3.5 h-3.5" /> 重试深度对比</button>
            ) : (
              <button onClick={handleDeep} className="w-full flex items-center justify-center gap-1.5 px-3 py-2 bg-[#0052D9] text-white text-sm rounded-lg hover:bg-[#0047BD]"><Sparkles className="w-3.5 h-3.5" /> 生成深度对比</button>
            )}
          </div>
          <Link to="/resumes" className="text-xs text-gray-400 hover:text-[#0052D9] text-center mt-2">返回我的简历</Link>
        </aside>
      </div>

      <AgentStepLoadingOverlay
        open={phase === "deep_loading" || deepOverlayStatus === "success"}
        title="智能体正在深度分析两版简历差异"
        description="系统正在结合简历画像和内容差异，生成完整对比分析。"
        steps={deepSteps}
        status={deepOverlayStatus === "success" ? "success" : phase === "deep_error" ? "error" : "loading"}
        errorText={deepErrMsg} successDelayMs={400}
        onClose={() => { if (deepOverlayStatus === "success") { setPhase("deep_done"); setDeepOverlayStatus("loading"); } }}
      />
    </ContentPageShell>
  );
}
