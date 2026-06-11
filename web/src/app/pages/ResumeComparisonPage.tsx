import { useState, useEffect, useCallback, useRef } from "react";
import { Link, useSearchParams, useNavigate, useLocation } from "react-router";
import { Loader2, Sparkles, RefreshCw, Target, CheckCheck, GitCompare, Search, ExternalLink, ChevronRight } from "lucide-react";
import { loadLatestCompareV2ViewModel, loadCompareV2ViewModel } from "../../api";
import { loadResumeOptionViewModelsFromApi } from "../../api";
import type { CompareV2ViewModel } from "../../api/viewModels";
import type { ResumeOptionViewModel } from "../../adapters";
import { ContentPageShell } from "../components/layout/ContentPageShell";
import { PageLoading, PageError } from "../components/PageState";
import { AgentStepLoadingOverlay, type AgentLoadingStep } from "../components/common/AgentStepLoadingOverlay";
import { BulletList } from "../components/common/BulletList";
import { ResumePickerModal, type ResumePickerOption } from "../components/ResumePickerModal";
import { getAuthRedirectIf401 } from "../utils/authRedirect";
import { humanizeProfile, humanizeProfileChange } from "../utils/profileLabel";
import { cleanUserText } from "../utils/cleanUserText";
import { getResumes } from "../../api/resumes";
import type { Resume } from "../../types";

// ─── Text safety ────────────────────────────────────────────────────────

/** K13-Stability-A: Safe against non-string types. */
function cleanText(value?: string | null): string {
  if (!value) return "";
  if (typeof value !== "string") return "";
  // Step 1: sanitize control chars
  const sanitized = value
    .replace(/\uFFFD/g, "")
    .replace(/[\u0000-\u0008\u000B\u000C\u000E-\u001F]/g, "")
    .replace(/\s+/g, " ");
  // Step 2: apply user-facing text cleaning (画像→简历定位, 标签→关键词, etc.)
  const cleaned = cleanUserText(sanitized);
  // Step 3: page-specific low-value replacements
  return cleaned
    .replace(/\bfrom版本\b/g, "原版本")
    .replace(/\bto版本\b/g, "目标版本")
    .replace(/保留from版本/g, "保留原版本")
    .replace(/保留to版本/g, "保留目标版本")
    .replace(/目标版本在整体质量上更优/g, "新版更适合作为当前主投递版本")
    .trim();
}

function isLowValueCompareText(text: string): boolean {
  if (!text) return true;
  const low = text.trim();
  return (
    low.includes("整体结构框架") ||
    low.includes("具体模块内容") ||
    low === "保留原版本的整体结构框架" ||
    low === "保留目标版本的整体结构框架" ||
    low === "保留原版本的具体模块内容" ||
    low === "保留目标版本的具体模块内容" ||
    low.length < 5
  );
}

function filterLowValue(items: string[]): string[] {
  const result = items.filter((s) => !isLowValueCompareText(s));
  return result.length > 0 ? result : [];
}

// ─── Helpers ───────────────────────────────────────────────────────────

function getMode(urlJobId: string | null, urlDir: string | null): { mode: string; dir: string; jobId: number | null } {
  if (urlJobId && /^\d+$/.test(urlJobId)) return { mode: "target_job", dir: "", jobId: Number(urlJobId) };
  if (urlDir && urlDir.trim()) return { mode: "target_direction", dir: urlDir.trim(), jobId: null };
  return { mode: "general", dir: "", jobId: null };
}

type ComparePhase =
  | "idle"
  | "loading_latest" | "fast_loading" | "fast_done" | "fast_error"
  | "deep_loading" | "deep_done" | "deep_error";

const deepSteps: AgentLoadingStep[] = [
  { key: "profiles", label: "读取两份简历版本信息" },
  { key: "diff", label: "分析优势、风险和关键差异" },
  { key: "merge", label: "生成可合并内容与改写示例" },
  { key: "next", label: "汇总下一版简历迭代建议" },
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

  const [fromId, setFromId] = useState<number | null>(fromParam && /^\d+$/.test(fromParam) ? Number(fromParam) : null);
  const [toId, setToId] = useState<number | null>(toParam && /^\d+$/.test(toParam) ? Number(toParam) : null);
  const [resumeLabels, setResumeLabels] = useState<Record<number, string>>({});
  const [loadingResumes, setLoadingResumes] = useState(true);

  const [phase, setPhase] = useState<ComparePhase>(fromParam && toParam ? "loading_latest" : "idle");
  const [fastResult, setFastResult] = useState<CompareV2ViewModel | null>(null);
  const [deepResult, setDeepResult] = useState<CompareV2ViewModel | null>(null);
  const [errorMsg, setErrorMsg] = useState("");
  const [deepErrMsg, setDeepErrMsg] = useState("");
  const [deepBackendDone, setDeepBackendDone] = useState(false);

  // ── Selection modal ──
  const [showPicker, setShowPicker] = useState(false);
  const [pickerOptions, setPickerOptions] = useState<ResumePickerOption[]>([]);
  const [pickerLoading, setPickerLoading] = useState(false);
  const [pickerError, setPickerError] = useState<string | null>(null);
  const [pickerSelectedIds, setPickerSelectedIds] = useState<number[]>([]);

  const loadPickerOptions = useCallback(() => {
    setPickerLoading(true); setPickerError(null);
    loadResumeOptionViewModelsFromApi()
      .then((opts) => setPickerOptions(opts.map((o) => ({ id: o.id, name: o.name, directionLabel: o.directionLabel }))))
      .catch(() => setPickerError("简历列表加载失败"))
      .finally(() => setPickerLoading(false));
  }, []);

  const openPicker = useCallback(() => {
    setPickerSelectedIds([]);
    loadPickerOptions();
    setShowPicker(true);
  }, [loadPickerOptions]);

  useEffect(() => {
    if (!fromParam || !toParam) return;
    const f = Number(fromParam), t = Number(toParam);
    if (!isNaN(f) && !isNaN(t)) { setFromId(f); setToId(t); }
  }, [fromParam, toParam]);

  // ── Load resume labels ──
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

  // ── Validate params ──
  useEffect(() => {
    if (loadingResumes || phase !== "loading_latest") return;
    if (!fromId || !toId) { setPhase("idle"); return; }
    if (fromId === toId) { setErrorMsg("不能对比相同的简历。"); setPhase("fast_error"); return; }
    loadLatest();
  }, [fromId, toId, loadingResumes]);

  async function loadLatest() {
    if (!fromId || !toId) return;
    setPhase("loading_latest"); setErrorMsg("");
    try {
      const latest = await loadLatestCompareV2ViewModel({ fromResumeId: fromId, toResumeId: toId, comparisonMode: mode, targetDirection: dir, jobId: jobId ?? undefined });
      if (latest.hasResult && latest.result) {
        if (latest.generationTier === "quality") { setDeepResult(latest.result); setPhase("deep_done"); }
        else { setFastResult(latest.result); setPhase("fast_done"); }
        return;
      }
    } catch { /* fall through */ }
    generateFast();
  }

  function generateFast() {
    if (!fromId || !toId) return;
    setPhase("fast_loading"); setErrorMsg("");
    loadCompareV2ViewModel({ fromResumeId: fromId, toResumeId: toId, comparisonMode: mode, generationTier: "fast", targetDirection: dir, jobId: jobId ?? undefined })
      .then((vm) => { setFastResult(vm); setPhase("fast_done"); })
      .catch(() => { setErrorMsg("快速对比生成失败，请稍后重试。"); setPhase("fast_error"); });
  }

  async function handleDeep() {
    if (!fromId || !toId) return;
    setPhase("deep_loading"); setDeepErrMsg(""); setDeepBackendDone(false);
    const abort = new AbortController();
    const timer = setTimeout(() => abort.abort(), 90000);
    abort.signal.addEventListener("abort", () => clearTimeout(timer));
    try {
      const vm = await loadCompareV2ViewModel({ fromResumeId: fromId, toResumeId: toId, comparisonMode: mode, generationTier: "quality", targetDirection: dir, jobId: jobId ?? undefined }, abort.signal);
      setDeepResult(vm); setDeepBackendDone(true);
    } catch {
      setPhase("deep_error");
      setDeepErrMsg("深度对比生成失败，当前快速对比结果仍可参考。");
    }
  }

  const activeResult: CompareV2ViewModel | null = phase === "deep_done" && deepResult ? deepResult : fastResult;
  const isDeep = phase === "deep_done";
  const isFast = phase === "fast_done";
  const fromLabel = fromId ? resumeLabels[fromId] || `简历 ${fromId}` : "";
  const toLabel = toId ? resumeLabels[toId] || `简历 ${toId}` : "";

  // ── No params: show selection guide ──
  if (!loadingResumes && (phase === "idle" || (!fromId || !toId))) {
    return (
      <ContentPageShell maxWidthClassName="max-w-[1280px]" contentClassName="pb-12">
        <div className="py-20 text-center max-w-lg mx-auto">
          <div className="w-16 h-16 mx-auto mb-5 rounded-2xl bg-blue-50 flex items-center justify-center">
            <GitCompare className="w-7 h-7 text-[#0052D9]" />
          </div>
          <h2 className="text-lg font-semibold text-slate-800 mb-2">选择两份简历进行版本对比</h2>
          <p className="text-sm text-slate-500 mb-8 leading-relaxed">
            系统会分析两份简历的优势、风险、可合并内容和下一版迭代建议。
          </p>
          <button onClick={openPicker} className="inline-flex items-center gap-2 px-5 py-2.5 bg-[#0052D9] text-white text-sm font-medium rounded-lg hover:bg-[#003DA5] transition-colors">
            <Search className="w-4 h-4" /> 选择两份简历
          </button>
        </div>
        <ResumePickerModal
          open={showPicker}
          title="选择两份简历进行版本对比"
          description="先选择上一版，再选择当前版。系统会分析当前版相比上一版的提升、风险和下一版优化建议。"
          resumes={pickerOptions}
          loading={pickerLoading}
          error={pickerError}
          onRetry={loadPickerOptions}
          selectedIds={pickerSelectedIds}
          selectionMode="compare"
          onSelectionChange={setPickerSelectedIds}
          confirmText="开始对比"
          onConfirm={(ids) => { if (ids.length >= 2) navigate(`/resumes/compare?fromResumeId=${ids[0]}&toResumeId=${ids[1]}`); setShowPicker(false); }}
          onClose={() => setShowPicker(false)}
        />
      </ContentPageShell>
    );
  }

  // ── Error guards ──
  if (loadingResumes) return <PageLoading message="正在加载简历信息…" />;
  if (fromId === toId) return <PageError title="不能对比相同的简历" description="请选择两份不同的简历。"><Link to="/resumes" className="text-[#0052D9] text-sm underline">返回我的简历</Link></PageError>;
  if (phase === "loading_latest") return <PageLoading message="正在读取对比结果…" />;
  if (phase === "fast_loading") return <div className="max-w-[1280px] mx-auto px-6 py-32 text-center"><Loader2 className="w-10 h-10 animate-spin text-[#0052D9] mx-auto mb-4" /><h2 className="text-lg font-bold text-gray-900 mb-2">正在生成快速对比</h2><p className="text-sm text-gray-500 max-w-md mx-auto">系统正在对比两份简历的结构、方向和关键差异。</p></div>;
  if (phase === "fast_error") return <div className="max-w-[1280px] mx-auto px-6 py-12 text-center"><p className="text-gray-500 text-sm mb-3">{errorMsg}</p><button onClick={generateFast} className="text-[#0052D9] text-sm underline">重新尝试</button></div>;
  if (!activeResult) return <div className="max-w-[1280px] mx-auto px-6 py-12 text-center"><p className="text-gray-500 text-sm mb-3">暂无对比结果</p><Link to="/resumes" className="text-[#0052D9] text-sm underline">返回我的简历</Link></div>;

  const V = activeResult;
  const winnerLabel = V.winnerResumeId === fromId ? fromLabel : toLabel;

  return (
    <ContentPageShell maxWidthClassName="max-w-[1280px]" contentClassName="pb-12">

      {/* Overlay */}
      <AgentStepLoadingOverlay
        open={phase === "deep_loading"}
        title="正在生成深度对比报告"
        description="系统正在结合两份简历的定位和内容差异，生成完整对比分析。"
        steps={deepSteps}
        backendDone={deepBackendDone}
        onFinish={() => setPhase("deep_done")}
      />

      {/* ── Deep status bar ── */}
      {isDeep && (
        <div className="bg-emerald-50 border border-emerald-200 rounded-xl p-4 mb-6 flex items-start gap-3">
          <CheckCheck className="w-4 h-4 text-emerald-600 mt-0.5 shrink-0" />
          <div>
            <p className="text-sm font-semibold text-emerald-800 mb-0.5">深度对比报告已完成</p>
            <p className="text-[12px] text-emerald-700 leading-relaxed">基于两份简历的定位和内容差异生成。可重新生成或查看进一步分析。</p>
          </div>
        </div>
      )}

      {/* ── Top Decision Card ── */}
      <div className="bg-white rounded-xl border border-slate-200 shadow-sm mb-6 overflow-hidden">
        <div className="flex flex-col sm:flex-row sm:items-center divide-y sm:divide-y-0 sm:divide-x divide-slate-100">
          <div className="px-5 py-4 flex-1 min-w-0">
            <div className="flex items-center gap-3">
              <div className="flex items-center gap-1.5 text-sm font-semibold text-slate-800">
                <span className="truncate">{fromLabel}</span>
                <ChevronRight className="w-3 h-3 text-slate-300" />
                <span className="truncate">{toLabel}</span>
              </div>
            </div>
            <p className="text-[11px] text-slate-400 mt-1">
              {isDeep ? "深度对比已完成" : "快速对比 · 初步预览"}
              {V.fromProfileDominantProfile && <><span className="text-slate-300 mx-1.5">|</span>{humanizeProfileChange(V.fromProfileDominantProfile, V.toProfileDominantProfile)}</>}
            </p>
          </div>
          <div className="px-5 py-4 shrink-0">
            <div className="flex items-center gap-4">
              <div>
                <div className="text-[11px] text-slate-400 font-medium">对比结论</div>
                <div className="flex items-baseline gap-1">
                  <span className="text-xl font-bold text-[#0052D9]">{V.compareScore}</span>
                  <span className="text-xs text-slate-400">/100</span>
                </div>
                <p className="text-[11px] text-slate-600 mt-0.5 max-w-[200px] truncate">{V.compareLevel}</p>
              </div>
              <div className="bg-[#F4F8FF] rounded-lg px-3 py-2 border border-blue-100">
                <span className="text-[10px] text-[#0052D9] font-medium">推荐使用</span>
                <p className="text-sm font-semibold text-slate-800">{winnerLabel}</p>
                <p className="text-[10px] text-slate-500 max-w-[180px] truncate">{cleanText(V.winnerLabel)}</p>
              </div>
              <button onClick={openPicker} className="px-3 py-1.5 text-[12px] text-[#0052D9] border border-blue-200 rounded-lg hover:bg-blue-50 transition-colors shrink-0">重新选择</button>
            </div>
          </div>
        </div>
      </div>

      {isFast && (
        /* ── Fast: Deep CTA Card ── */
        <div className="bg-gradient-to-br from-[#F0F4FF] to-white border border-blue-200 rounded-xl p-6 shadow-sm mb-6">
          <div className="flex items-start gap-4">
            <div className="w-12 h-12 rounded-xl bg-[#0052D9]/10 flex items-center justify-center shrink-0">
              <Sparkles className="w-6 h-6 text-[#0052D9]" />
            </div>
            <div className="flex-1 min-w-0">
              <h3 className="text-base font-bold text-slate-800 mb-1">{phase === "deep_error" ? "深度对比报告生成失败" : "生成深度对比报告"}</h3>
              <p className="text-[13px] text-slate-500 leading-relaxed mb-4">
                {phase === "deep_error" ? "深度对比生成失败，已保留当前快速对比结果。请稍后重试。" : "当前结果是快速对比，只展示基础差异和初步推荐。生成深度对比后，系统会输出：两个版本的优势和风险对照、可保留内容、合并建议、改写示例，以及下一版简历迭代建议。"}
              </p>
              <button onClick={handleDeep} disabled={phase === "deep_loading"} className={`inline-flex items-center gap-2 px-6 py-3 text-sm font-semibold rounded-lg transition-all ${phase === "deep_loading" ? "bg-slate-200 text-slate-400 cursor-not-allowed" : "bg-[#0052D9] text-white hover:bg-[#003DA5] shadow-md hover:shadow-lg"}`}>
                {phase === "deep_loading" ? <><Loader2 className="w-4 h-4 animate-spin" /> 正在生成深度对比报告…</> : phase === "deep_error" ? <><RefreshCw className="w-4 h-4" /> 重新生成深度对比报告</> : <><Sparkles className="w-4 h-4" /> 生成深度对比报告</>}
              </button>
            </div>
          </div>
        </div>
      )}

      <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_320px] items-start">
        {/* ── Main Report ── */}
        <div className="min-w-0 space-y-4">
          {/* Iteration Summary */}
          {V.oneSentenceSummary && <div className="bg-blue-50/50 border border-blue-100 rounded-lg p-4"><p className="text-[13px] text-blue-800 leading-relaxed font-medium">{cleanText(V.oneSentenceSummary)}</p></div>}

          {/* Key Differences */}
          {V.keyDifferences.length > 0 && (
            <section className="bg-white border border-gray-200 rounded-lg p-5 shadow-sm">
              <h3 className="font-bold text-sm text-gray-900 mb-3 flex items-center gap-1.5"><span className="w-1 h-3.5 bg-[#0052D9] rounded-full" />本次迭代结论</h3>
              <BulletList items={V.keyDifferences.map(d => cleanText(d))} tone="blue" />
            </section>
          )}

          {/* 新版明显提升 (to_strengths) */}
          {V.toStrengths.length > 0 && (
            <section className="bg-white border border-emerald-200 rounded-lg p-5 shadow-sm">
              <h3 className="font-bold text-sm text-emerald-800 mb-3">新版明显提升</h3>
              <BulletList items={V.toStrengths.map(s => cleanText(s))} tone="green" size="sm" />
            </section>
          )}

          {/* 新版仍存在的问题 (to_risks) */}
          {V.toRisks.length > 0 && (
            <section className="bg-amber-50/50 border border-amber-100 rounded-lg p-5">
              <h3 className="font-bold text-sm text-amber-800 mb-2">新版仍存在的问题</h3>
              <BulletList items={V.toRisks.map(r => cleanText(r))} tone="amber" size="sm" />
            </section>
          )}

          {/* Deep sections */}
          {isDeep && (
            <div className="space-y-4">
              {/* 旧版值得保留 (filtered keepFromItems) */}
              {(() => { const items = filterLowValue(V.keepFromItems); if (items.length > 0) return <section className="bg-white border border-gray-200 rounded-lg p-5 shadow-sm"><h3 className="font-bold text-sm text-gray-900 mb-2">旧版值得保留的内容</h3><BulletList items={items.map(k => cleanText(k))} tone="slate" size="sm" /></section>; if (V.keepFromItems.length > 0) return null; return null; })()}

              {/* 新版必须保留 (filtered keepToItems) */}
              {(() => { const items = filterLowValue(V.keepToItems); if (items.length > 0) return <section className="bg-white border border-gray-200 rounded-lg p-5 shadow-sm"><h3 className="font-bold text-sm text-gray-900 mb-2">新版必须保留的内容</h3><BulletList items={items.map(k => cleanText(k))} tone="slate" size="sm" /></section>; if (V.keepToItems.length > 0) return null; return null; })()}

              {/* 下一版合并方案 (merge + next) */}
              {V.mergeSuggestions.length > 0 && (
                <section className="bg-green-50/50 border border-green-200 rounded-lg p-5"><h3 className="font-bold text-sm text-green-900 mb-3 flex items-center gap-1.5"><span className="w-1 h-3.5 bg-green-500 rounded-full" />下一版合并方案</h3>
                  <BulletList items={V.mergeSuggestions.map(m => cleanText(m))} tone="green" />
                  {V.nextVersionSuggestions.length > 0 && <div className="mt-3 pt-3 border-t border-green-200"><BulletList items={V.nextVersionSuggestions.map(n => cleanText(n))} tone="green" /></div>}
                </section>
              )}

              {/* 改写示例 */}
              {V.rewriteExamples.length > 0 && (
                <section className="bg-white border border-gray-200 rounded-lg p-5 shadow-sm"><h3 className="font-bold text-sm text-gray-900 mb-3 flex items-center gap-1.5"><span className="w-1 h-3.5 bg-purple-500 rounded-full" />可直接复制的改写示例</h3>
                  <div className="space-y-3">{V.rewriteExamples.map((ex, i) => (
                    <div key={i} className="bg-gray-50 rounded-lg p-3 border border-gray-100">
                      <div className="text-[11px] text-purple-600 font-medium mb-1">{cleanText(ex.section)}</div>
                      <div className="grid grid-cols-2 gap-3 text-[12px]">
                        <div><span className="text-gray-400 text-[10px]">旧版本</span><p className="text-gray-500 mt-0.5">{cleanText(ex.fromVersion)}</p></div>
                        <div><span className="text-gray-400 text-[10px]">当前版本</span><p className="text-gray-700 mt-0.5 font-medium">{cleanText(ex.toVersion)}</p></div>
                      </div>
                      {ex.suggestedNext && <p className="text-[11px] text-blue-600 mt-2 border-t border-gray-200 pt-2">下一版建议: {cleanText(ex.suggestedNext)}</p>}
                      {ex.reason && <p className="text-[10px] text-gray-400 mt-1">原因: {cleanText(ex.reason)}</p>}
                    </div>
                  ))}</div>
                </section>
              )}
            </div>
          )}

          {/* 初筛通过率影响 */}
          {V.deliveryAdvice && <div className="bg-blue-50/30 border border-blue-100 rounded-lg p-4"><h3 className="font-bold text-sm text-blue-800 mb-1">初筛通过率影响</h3><p className="text-[13px] text-blue-700">{cleanText(V.deliveryAdvice)}</p></div>}

          {/* Old cache hint for fast reports */}
          {isFast && (
            <div className="bg-gray-50 border border-gray-100 rounded-lg p-3 text-center">
              <p className="text-[11px] text-gray-400">当前展示的是快捷对比结果。如内容不够具体，建议点击上方按钮生成深度对比报告，获取更完整的迭代诊断。</p>
            </div>
          )}
        </div>

        {/* ── Right Action Panel ── */}
        <aside className="shrink-0 flex flex-col gap-4">
          {/* Compare objects */}
          <div className="bg-white border border-gray-200 rounded-lg overflow-hidden">
            <div className="px-4 py-3 border-b border-gray-100"><h3 className="text-[13px] font-medium text-gray-500">对比对象</h3></div>
            <div className="p-4">
              <p className="text-sm font-semibold text-gray-900 mb-1">{fromLabel} → {toLabel}</p>
              {V.fromProfileDominantProfile && <p className="text-[11px] text-[#0052D9]">{humanizeProfile(V.fromProfileDominantProfile)} → {humanizeProfile(V.toProfileDominantProfile)}</p>}
            </div>
          </div>

          {/* Status */}
          <div className="bg-white border border-gray-200 rounded-lg overflow-hidden">
            <div className="px-4 py-3 border-b border-gray-100"><h3 className="text-[13px] font-medium text-gray-500">当前结果</h3></div>
            <div className="p-4">
              <div className="flex items-center gap-2 mb-2">
                <div className={`w-2 h-2 rounded-full ${isDeep ? "bg-emerald-400" : "bg-amber-400"}`} />
                <span className="text-[12px] text-gray-600">{isDeep ? "深度对比已完成" : "快速对比"}</span>
              </div>
              <p className="text-2xl font-bold text-[#0052D9]">{V.compareScore} <span className="text-xs text-gray-400">/100</span></p>
              <p className="text-[12px] text-gray-600 mt-1">推荐：{winnerLabel}</p>
            </div>
          </div>

          {/* Actions */}
          <div className="bg-white border border-gray-200 rounded-lg overflow-hidden">
            <div className="px-4 py-3 border-b border-gray-100"><h3 className="text-[13px] font-medium text-gray-500">对比操作</h3></div>
            <div className="p-4 space-y-3">
              {phase === "deep_loading" ? (
                <button disabled className="w-full flex items-center justify-center gap-2 px-4 py-3 bg-slate-200 text-slate-500 text-sm font-semibold rounded-lg cursor-not-allowed"><Loader2 className="w-4 h-4 animate-spin" /> 正在生成深度对比报告…</button>
              ) : isDeep ? (
                <button onClick={handleDeep} className="w-full flex items-center justify-center gap-2 px-4 py-3 bg-white border border-blue-200 text-[#0052D9] text-sm font-semibold rounded-lg hover:bg-blue-50 transition-colors"><RefreshCw className="w-4 h-4" /> 重新生成深度对比报告</button>
              ) : phase === "deep_error" ? (
                <button onClick={handleDeep} className="w-full flex items-center justify-center gap-2 px-4 py-3 bg-[#0052D9] text-white text-sm font-semibold rounded-lg hover:bg-[#003DA5] transition-colors"><RefreshCw className="w-4 h-4" /> 重新生成深度对比报告</button>
              ) : (
                <button onClick={handleDeep} className="w-full flex items-center justify-center gap-2 px-4 py-3 bg-[#0052D9] text-white text-sm font-semibold rounded-lg hover:bg-[#003DA5] shadow-sm transition-colors"><Sparkles className="w-4 h-4" /> 生成深度对比报告</button>
              )}
              <div className="border-t border-gray-100" />
              <button onClick={openPicker} className="w-full flex items-center gap-2 px-2 py-1.5 text-[12px] text-gray-500 hover:text-[#0052D9] hover:bg-gray-50 rounded transition-colors"><Search className="w-3 h-3" /> 重新选择两份简历</button>
              <Link to="/resumes" className="w-full flex items-center gap-2 px-2 py-1.5 text-[12px] text-gray-500 hover:text-[#0052D9] hover:bg-gray-50 rounded transition-colors"><ExternalLink className="w-3 h-3" /> 返回我的简历</Link>
            </div>
          </div>
        </aside>
      </div>

      {/* Modal for reselect */}
      <ResumePickerModal
        open={showPicker}
        title="选择两份简历进行版本对比"
        description="先选择上一版，再选择当前版。系统会分析当前版相比上一版的提升、风险和下一版优化建议。"
        resumes={pickerOptions}
        loading={pickerLoading}
        error={pickerError}
        onRetry={loadPickerOptions}
        selectedIds={pickerSelectedIds}
        selectionMode="compare"
        onSelectionChange={setPickerSelectedIds}
        confirmText="开始对比"
        onConfirm={(ids) => { if (ids.length >= 2) navigate(`/resumes/compare?fromResumeId=${ids[0]}&toResumeId=${ids[1]}`); setShowPicker(false); }}
        onClose={() => setShowPicker(false)}
      />
    </ContentPageShell>
  );
}
