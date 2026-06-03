import { useState, useEffect } from "react";
import { Link, useNavigate, useLocation } from "react-router";
import {
  ArrowLeft,
  ArrowRight,
  ChevronDown,
  AlertCircle,
  FileText,
  TrendingUp,
  TrendingDown,
  Minus,
  CheckCircle2,
  Clock,
} from "lucide-react";
import { loadResumeComparisonViewModel } from "../../api";
import { ContentPageShell } from "../components/layout/ContentPageShell";
import { getResumeVersions } from "../../api/index";
import type { ResumeVersionListItem } from "../../types";
import type { ResumeComparisonViewModel } from "../../adapters";
import { PageLoading, PageError, PageEmpty } from "../components/PageState";
import { getAuthRedirectIf401 } from "../utils/authRedirect";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../components/ui/select";

function getTrendIcon(trend: string) {
  switch (trend) {
    case "reduced":
      return <TrendingDown className="w-3.5 h-3.5 text-green-500" />;
    case "increased":
      return <TrendingUp className="w-3.5 h-3.5 text-orange-500" />;
    default:
      return <Minus className="w-3.5 h-3.5 text-gray-400" />;
  }
}

function getStatusBadge(status: string) {
  switch (status) {
    case "completed":
      return <span className="text-[10px] bg-emerald-100 text-emerald-700 px-1.5 py-0.5 rounded">已完成</span>;
    case "partially_completed":
      return <span className="text-[10px] bg-amber-100 text-amber-700 px-1.5 py-0.5 rounded">部分完成</span>;
    case "not_started":
      return <span className="text-[10px] bg-gray-100 text-gray-500 px-1.5 py-0.5 rounded">未开始</span>;
    default:
      return <span className="text-[10px] bg-gray-100 text-gray-500 px-1.5 py-0.5 rounded">{status}</span>;
  }
}

export function ResumeComparisonPage() {
  const [versions, setVersions] = useState<ResumeVersionListItem[]>([]);
  const [baseVersion, setBaseVersion] = useState<ResumeVersionListItem | null>(null);
  const [targetVersion, setTargetVersion] = useState<ResumeVersionListItem | null>(null);
  const [comparison, setComparison] = useState<ResumeComparisonViewModel | null>(null);
  const [versionsLoading, setVersionsLoading] = useState(true);
  const [versionsError, setVersionsError] = useState(false);
  const [compareLoading, setCompareLoading] = useState(false);
  const [compareError, setCompareError] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();

  // Load version list
  useEffect(() => {
    getResumeVersions()
      .then((list) => {
        setVersions(list);
        // Auto-select: base = earliest (last in array), target = latest (first in array)
        if (list.length >= 2) {
          setBaseVersion(list[list.length - 1]);
          setTargetVersion(list[0]);
        } else if (list.length === 1) {
          setBaseVersion(list[0]);
          setTargetVersion(list[0]);
        }
      })
      .catch((err) => {
        const authRedirect = getAuthRedirectIf401(err, location.pathname, location.search);
        if (authRedirect) { navigate(authRedirect, { replace: true }); return; }
        setVersionsError(true);
      })
      .finally(() => setVersionsLoading(false));
  }, []);

  // Run comparison when both versions selected
  useEffect(() => {
    if (!baseVersion || !targetVersion || baseVersion.version_no === targetVersion.version_no) return;

    setCompareLoading(true);
    setCompareError(false);
    loadResumeComparisonViewModel(
      String(baseVersion.version_no),
      String(targetVersion.version_no),
      baseVersion.name || `V${baseVersion.version_no}`,
      targetVersion.name || `V${targetVersion.version_no}`,
    )
      .then(setComparison)
      .catch(() => setCompareError(true))
      .finally(() => setCompareLoading(false));
  }, [baseVersion?.version_no, targetVersion?.version_no]);

  // --- Version list loading ---
  if (versionsLoading) {
    return <PageLoading message="正在加载简历版本…" maxWidth="max-w-[1360px]" />;
  }

  // --- Version list error ---
  if (versionsError) {
    return (
      <PageError
        title="简历版本数据加载失败，请稍后重试。"
        description=""
        maxWidth="max-w-[1360px]"
      >
        <Link to="/resumes" className="text-[#0052D9] text-sm hover:underline">返回我的简历</Link>
      </PageError>
    );
  }

  // --- No versions ---
  if (versions.length === 0) {
    return (
      <PageEmpty
        title="暂无可比较的简历版本"
        description=""
        maxWidth="max-w-[1360px]"
      >
        <Link to="/resumes" className="text-[#0052D9] text-sm hover:underline">返回我的简历</Link>
      </PageEmpty>
    );
  }

  // --- Only one version ---
  if (versions.length === 1) {
    return (
      <PageEmpty
        title="仅有一个简历版本，无法进行比较。"
        description="请先创建更多版本。"
        maxWidth="max-w-[1360px]"
      >
        <Link to="/resumes" className="text-[#0052D9] text-sm hover:underline">返回我的简历</Link>
      </PageEmpty>
    );
  }

  const baseLabel = baseVersion?.name || `V${baseVersion?.version_no}`;
  const targetLabel = targetVersion?.name || `V${targetVersion?.version_no}`;

  return (
    <ContentPageShell>

        {/* Header */}
        <div className="mb-8">
          <Link to="/resumes" className="flex items-center text-sm text-slate-500 hover:text-blue-600 transition-colors mb-4 group">
            <ArrowLeft className="w-4 h-4 mr-1 transition-transform group-hover:-translate-x-0.5" />
            返回我的简历
          </Link>
          <h1 className="text-2xl font-semibold text-slate-900 tracking-tight mb-1">历史简历比较</h1>
          <p className="text-sm text-slate-500">
            比较不同简历版本的变化，查看版本间差异分析。
          </p>
        </div>

        {/* Version Selection */}
        <div className="bg-white rounded-lg border border-gray-200 shadow-sm p-4 mb-8 flex flex-col md:flex-row items-start md:items-center justify-between gap-4">
          <div className="flex items-center gap-4 flex-wrap">
            <div className="flex flex-col">
              <span className="text-xs text-slate-500 mb-1">旧版本（基准）</span>
              <Select
                value={baseVersion ? String(baseVersion.version_no) : undefined}
                onValueChange={(v) => {
                  const ver = versions.find((vv) => vv.version_no === Number(v));
                  if (ver) setBaseVersion(ver);
                }}
              >
                <SelectTrigger className="max-w-[200px] bg-slate-50 text-sm font-medium text-slate-700">
                  <SelectValue placeholder="选择版本" />
                </SelectTrigger>
                <SelectContent>
                  {versions.map((v) => (
                    <SelectItem key={v.version_no} value={String(v.version_no)}>
                      V{v.version_no} {v.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="text-slate-300 pt-4">
              <ArrowRight className="w-5 h-5" />
            </div>

            <div className="flex flex-col">
              <span className="text-xs text-slate-500 mb-1">新版本（目标）</span>
              <Select
                value={targetVersion ? String(targetVersion.version_no) : undefined}
                onValueChange={(v) => {
                  const ver = versions.find((vv) => vv.version_no === Number(v));
                  if (ver) setTargetVersion(ver);
                }}
              >
                <SelectTrigger className="max-w-[200px] bg-blue-50 text-sm font-medium text-blue-700">
                  <SelectValue placeholder="选择版本" />
                </SelectTrigger>
                <SelectContent>
                  {versions.map((v) => (
                    <SelectItem key={v.version_no} value={String(v.version_no)}>
                      V{v.version_no} {v.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>
        </div>

        {/* Comparison Content */}
        {compareLoading ? (
          <PageLoading message="正在生成版本对比…" maxWidth="max-w-[1360px]" className="py-12" />
        ) : compareError ? (
          <PageError
            title="版本对比接口暂时不可用，请稍后重试。"
            description=""
            maxWidth="max-w-[1360px]"
          >
            <Link to="/resumes" className="text-[#0052D9] text-sm hover:underline">返回我的简历</Link>
          </PageError>
        ) : !comparison ? (
          <div className="max-w-[1360px] mx-auto py-12 text-center">
            <p className="text-gray-400 text-sm mb-2">暂无可展示的对比数据</p>
            <Link to="/resumes" className="text-[#0052D9] text-sm hover:underline">返回我的简历</Link>
          </div>
        ) : (
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">

            {/* Left: Base version card */}
            <div className="lg:col-span-3">
              <div className="bg-white rounded-xl border border-gray-200 p-5 shadow-sm sticky top-20">
                <h3 className="font-semibold text-slate-900 mb-2">旧版本：{baseLabel}</h3>
                <span className="inline-flex px-2 py-0.5 rounded text-xs font-medium bg-slate-100 text-slate-600 mb-4">
                  历史版本
                </span>

                <div className="space-y-4 text-sm">
                  <div>
                    <div className="text-slate-500 text-xs mb-1">基准评分</div>
                    <div className="text-2xl font-bold text-slate-700">{comparison.scoreBefore}</div>
                  </div>

                  {comparison.fieldChanges.length > 0 && (
                    <div className="pt-2 border-t border-gray-100">
                      <div className="text-slate-500 text-xs mb-2">版本基本信息</div>
                      <div className="space-y-1.5 text-xs">
                        {comparison.fieldChanges
                          .filter((fc) => fc.field !== "version")
                          .map((fc, i) => (
                            <div key={i} className="flex items-start gap-1.5">
                              <span className="text-slate-400 shrink-0 mt-0.5">•</span>
                              <span className="text-slate-600">{fc.field}：{fc.before || "—"}</span>
                            </div>
                          ))}
                      </div>
                    </div>
                  )}
                </div>
              </div>
            </div>

            {/* Center: Diff Analysis */}
            <div className="lg:col-span-6 space-y-6">

              {/* Score Overview */}
              <section className="bg-white rounded-xl border border-gray-200 p-6 shadow-sm">
                <h2 className="text-base font-semibold text-slate-900 mb-4 flex items-center">
                  <FileText className="w-4 h-4 mr-2 text-[#0052D9]" />
                  版本变化总览
                </h2>

                <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                  <div className="bg-[#F5F7FA] rounded-xl p-5 border border-gray-100 flex flex-col items-center justify-center text-center">
                    <div className="text-xs text-slate-500 mb-2 font-medium">评分变化</div>
                    <div className="flex items-center justify-center gap-3">
                      <span className="text-slate-400 text-xl font-semibold">{comparison.scoreBefore}</span>
                      <span className="text-slate-300">→</span>
                      <span className="text-3xl font-bold" style={{ color: comparison.scoreChange >= 0 ? "#0052D9" : "#ef4444" }}>
                        {comparison.scoreAfter}
                      </span>
                    </div>
                    <div className="text-xs mt-2" style={{ color: comparison.scoreChange >= 0 ? "#16a34a" : "#ef4444" }}>
                      {comparison.scoreChange >= 0 ? "+" : ""}{comparison.scoreChange}
                    </div>
                  </div>

                  <div className="bg-[#F5F7FA] rounded-xl p-5 border border-gray-100 flex flex-col items-center justify-center text-center">
                    <div className="text-xs text-slate-500 mb-2 font-medium">字段变化</div>
                    <div className="text-3xl font-bold text-[#0052D9]">{comparison.fieldChanges.length}</div>
                    <div className="text-xs text-slate-400 mt-1">个字段</div>
                  </div>

                  <div className="bg-[#F5F7FA] rounded-xl p-5 border border-gray-100 flex flex-col items-center justify-center text-center">
                    <div className="text-xs text-slate-500 mb-2 font-medium">风险状态</div>
                    <div className="text-3xl font-bold text-[#0052D9]">{comparison.riskChanges.length}</div>
                    <div className="text-xs text-slate-400 mt-1">项变化</div>
                  </div>
                </div>
              </section>

              {/* Field Changes Table */}
              {comparison.fieldChanges.length > 0 && (
                <section className="bg-white rounded-xl border border-gray-200 p-6 shadow-sm overflow-hidden">
                  <h2 className="text-base font-semibold text-slate-900 mb-4">字段变化明细</h2>
                  <div className="overflow-x-auto border border-gray-100 rounded-lg">
                    <table className="w-full text-sm text-left">
                      <thead className="bg-slate-50 text-slate-500 text-xs uppercase border-b border-gray-100">
                        <tr>
                          <th className="px-4 py-3 font-medium">字段</th>
                          <th className="px-4 py-3 font-medium">旧值</th>
                          <th className="px-4 py-3 font-medium">新值</th>
                          <th className="px-4 py-3 font-medium">变化类型</th>
                          <th className="px-4 py-3 font-medium">影响</th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-gray-100">
                        {comparison.fieldChanges.map((fc, i) => (
                          <tr key={i} className={i % 2 === 0 ? "bg-white hover:bg-slate-50/50" : "bg-slate-50/30 hover:bg-slate-50/50"}>
                            <td className="px-4 py-3 font-medium text-slate-800">{fc.field}</td>
                            <td className="px-4 py-3 text-slate-500 max-w-[160px] truncate">{fc.before || "—"}</td>
                            <td className="px-4 py-3 text-blue-600 font-medium max-w-[160px] truncate">{fc.after || "—"}</td>
                            <td className="px-4 py-3">
                              <span className={`text-xs px-2 py-0.5 rounded ${
                                fc.changeType === "added" ? "bg-emerald-50 text-emerald-600" :
                                fc.changeType === "modified" ? "bg-blue-50 text-blue-600" :
                                fc.changeType === "removed" ? "bg-red-50 text-red-600" :
                                "bg-gray-50 text-gray-500"
                              }`}>
                                {fc.changeType === "added" ? "新增" :
                                 fc.changeType === "modified" ? "修改" :
                                 fc.changeType === "removed" ? "移除" :
                                 fc.changeType === "unchanged" ? "不变" : fc.changeType}
                              </span>
                            </td>
                            <td className="px-4 py-3 text-xs text-slate-600 max-w-[100px]">{fc.impact}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </section>
              )}

              {/* Key Improvements */}
              {comparison.keyImprovements.length > 0 && (
                <section className="bg-white rounded-xl border border-gray-200 p-6 shadow-sm">
                  <h2 className="text-base font-semibold text-slate-900 mb-4">关键改进</h2>
                  <ul className="space-y-2">
                    {comparison.keyImprovements.map((item, i) => (
                      <li key={i} className="flex items-start gap-2 text-sm text-slate-700">
                        <CheckCircle2 className="w-4 h-4 text-emerald-500 shrink-0 mt-0.5" />
                        {item}
                      </li>
                    ))}
                  </ul>
                </section>
              )}

              {/* Suggestion Completion */}
              {comparison.suggestionCompletion.length > 0 && (
                <section className="bg-white rounded-xl border border-gray-200 p-6 shadow-sm">
                  <h2 className="text-base font-semibold text-slate-900 mb-4">优化建议完成情况</h2>
                  <div className="space-y-3">
                    {comparison.suggestionCompletion.map((sc, i) => (
                      <div key={i} className="flex items-start gap-3 p-3 rounded-lg bg-slate-50 border border-gray-100">
                        <div className="shrink-0 mt-0.5">
                          {getStatusBadge(sc.status)}
                        </div>
                        <div className="flex-1 min-w-0">
                          <div className="text-sm font-medium text-slate-800 mb-1">{sc.suggestion}</div>
                          {sc.evidence && (
                            <div className="text-xs text-slate-500">{sc.evidence}</div>
                          )}
                        </div>
                      </div>
                    ))}
                  </div>
                </section>
              )}

              {/* Risk Changes */}
              {comparison.riskChanges.length > 0 && (
                <section className="bg-white rounded-xl border border-gray-200 p-6 shadow-sm">
                  <h2 className="text-base font-semibold text-slate-900 mb-4">风险变化</h2>
                  <div className="overflow-x-auto border border-gray-100 rounded-lg">
                    <table className="w-full text-sm text-left">
                      <thead className="bg-slate-50 text-slate-500 text-xs uppercase border-b border-gray-100">
                        <tr>
                          <th className="px-4 py-3 font-medium">风险项</th>
                          <th className="px-4 py-3 font-medium">之前</th>
                          <th className="px-4 py-3 font-medium">之后</th>
                          <th className="px-4 py-3 font-medium">趋势</th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-gray-100">
                        {comparison.riskChanges.map((rc, i) => (
                          <tr key={i} className="hover:bg-slate-50/50">
                            <td className="px-4 py-3 font-medium text-slate-800">{rc.risk}</td>
                            <td className="px-4 py-3 text-slate-500">{rc.before || "—"}</td>
                            <td className="px-4 py-3 text-slate-700">{rc.after || "—"}</td>
                            <td className="px-4 py-3 flex items-center gap-1">
                              {getTrendIcon(rc.trend)}
                              <span className="text-xs text-slate-500 capitalize">{rc.trend}</span>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </section>
              )}

              {/* Remaining Risks */}
              {comparison.remainingRisks.length > 0 && (
                <section className="bg-orange-50/50 rounded-xl border border-orange-100 p-6">
                  <h2 className="text-base font-semibold text-orange-800 mb-3 flex items-center">
                    <AlertCircle className="w-4 h-4 mr-2 text-orange-500" />
                    仍然需要补强的内容
                  </h2>
                  <ul className="text-sm text-orange-700 space-y-2 list-disc pl-5 marker:text-orange-400">
                    {comparison.remainingRisks.map((risk, i) => (
                      <li key={i}>{risk}</li>
                    ))}
                  </ul>
                </section>
              )}

              {/* Summary + Authenticity Notice */}
              {(comparison.summary || comparison.authenticityNotice) && (
                <section className="bg-blue-50/50 rounded-xl border border-blue-100 p-6">
                  {comparison.summary && (
                    <>
                      <h2 className="text-base font-semibold text-blue-900 mb-3">对比总结</h2>
                      <p className="text-sm text-blue-800 mb-4 leading-relaxed">{comparison.summary}</p>
                    </>
                  )}
                  {comparison.authenticityNotice && (
                    <p className="text-xs text-blue-600/60">{comparison.authenticityNotice}</p>
                  )}
                </section>
              )}

              {/* Empty state for zero field changes */}
              {comparison.fieldChanges.length === 0 && comparison.keyImprovements.length === 0 && comparison.suggestionCompletion.length === 0 && comparison.riskChanges.length === 0 && (
                <div className="text-center py-8 text-gray-400 text-sm">
                  两个版本之间未检测到显著差异。
                </div>
              )}

            </div>

            {/* Right: Target version card */}
            <div className="lg:col-span-3">
              <div className="bg-white rounded-xl border border-blue-200 ring-1 ring-blue-50 p-5 shadow-md sticky top-20">
                <h3 className="font-semibold text-slate-900 mb-2">新版本：{targetLabel}</h3>
                <span className="inline-flex px-2 py-0.5 rounded text-xs font-medium bg-emerald-50 text-emerald-600 border border-emerald-100 mb-4">
                  目标版本
                </span>

                <div className="space-y-4 text-sm">
                  <div>
                    <div className="text-slate-500 text-xs mb-1">版本评分</div>
                    <div className="text-2xl font-bold text-[#0052D9]">{comparison.scoreAfter}</div>
                    <div className="text-xs mt-1" style={{ color: comparison.scoreChange >= 0 ? "#16a34a" : "#ef4444" }}>
                      较基准 {comparison.scoreChange >= 0 ? "+" : ""}{comparison.scoreChange}
                    </div>
                  </div>

                  <div className="pt-2 border-t border-gray-100">
                    <div className="text-slate-500 text-xs mb-1.5">版本信息</div>
                    <div className="flex items-center gap-1.5 text-xs text-slate-600">
                      <Clock className="w-3 h-3" />
                      {comparison.createdAt ? comparison.createdAt.substring(0, 10) : "—"}
                    </div>
                  </div>
                </div>
              </div>
            </div>

          </div>
        )}

      {/* Bottom bar */}
      {(baseVersion && targetVersion && comparison) && (
        <div className="fixed bottom-0 left-0 right-0 bg-white border-t border-gray-200 p-4 shadow-[0_-4px_6px_-1px_rgba(0,0,0,0.05)] z-40">
          <div className="max-w-[1360px] mx-auto px-6 flex flex-col md:flex-row items-center justify-between gap-4">
            <div className="flex items-center text-sm text-slate-500">
              当前比较：
              <span className="font-medium text-slate-700 ml-1">{baseLabel}</span>
              <ArrowRight className="w-3.5 h-3.5 mx-2 text-slate-400" />
              <span className="font-medium text-blue-600">{targetLabel}</span>
            </div>

            <div className="flex items-center gap-3">
              <Link to="/resumes" className="px-4 py-2 text-sm font-medium text-slate-600 bg-white border border-gray-300 rounded hover:bg-gray-50 transition-colors">
                返回我的简历
              </Link>
            </div>
          </div>
        </div>
      )}
    </ContentPageShell>
  );
}
