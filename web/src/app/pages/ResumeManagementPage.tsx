import { useState, useEffect, useMemo } from "react";
import { Link, useNavigate, useLocation } from "react-router";
import { PageLoading, PageError } from "../components/PageState";
import { ContentPageShell } from "../components/layout/ContentPageShell";
import { FileText, Plus, Search, Sparkles, Wand2, GitCompare, Pencil, Trash2, Eye } from "lucide-react";
import { getResumes, deleteResume, setCurrentResume } from "../../api/resumes";
import { AppConfirmDialog } from "../components/common/AppConfirmDialog";
import { OptimizationEntryModal } from "../components/OptimizationEntryModal";
import { ResumePickerModal, type ResumePickerOption } from "../components/ResumePickerModal";
import type { Resume } from "../../types";
import { getAuthRedirectIf401 } from "../utils/authRedirect";
import { isApiClientError } from "../../api/errors";

// ─── Constants ────────────────────────────────────────────────────────────────

const DESKTOP_PAGE_SIZE = 8;
const COMPACT_PAGE_SIZE = 5;

// ─── Types ────────────────────────────────────────────────────────────────────

interface ResumeRow {
  id: number;
  name: string;
  isCurrent: boolean;
  createdAt: string;
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

function toRows(resumes: Resume[]): ResumeRow[] {
  return resumes.map((r) => ({
    id: r.id,
    name: r.name || "",
    isCurrent: r.is_current || false,
    createdAt: r.created_at || r.updated_at || "",
  }));
}

function getOtherResumeId(resumes: Resume[], excludeId: number): number {
  const other = resumes.find(r => r.id !== excludeId);
  return other ? other.id : excludeId;
}

function formatDate(iso: string): string {
  if (!iso) return "";
  try {
    const d = new Date(iso);
    const y = d.getFullYear();
    const m = String(d.getMonth() + 1).padStart(2, "0");
    const day = String(d.getDate()).padStart(2, "0");
    return `${y}-${m}-${day}`;
  } catch {
    return iso.slice(0, 10);
  }
}

// ─── Sub-components ───────────────────────────────────────────────────────────

function EmptyState() {
  return (
    <div className="bg-white rounded-2xl border border-slate-100 shadow-sm p-12 text-center">
      <FileText className="w-12 h-12 text-slate-300 mx-auto mb-4" />
      <h2 className="text-lg font-semibold text-slate-700 mb-2">你还没有创建简历</h2>
      <p className="text-sm text-slate-500 mb-6 max-w-md mx-auto">
        添加一份简历后，可以用于岗位投递、匹配分析和简历优化。
      </p>
      <Link
        to="/resumes/edit"
        className="inline-flex items-center gap-1.5 px-5 py-2 bg-[#0052D9] text-white text-sm font-medium rounded-lg hover:bg-[#003DA5] transition-colors"
      >
        <Plus className="w-4 h-4" />
        添加简历
      </Link>
    </div>
  );
}

interface SmartAssistantProps {
  selectedResumeId: number | null;
  selectedResumeName: string;
  hasResumes: boolean;
  loading: boolean;
  resumes: ResumeRow[];
  onOptimize: () => void;
  onCompare: () => void;
}

function SmartAssistant({ selectedResumeId, selectedResumeName, hasResumes, loading, resumes, onOptimize, onCompare }: SmartAssistantProps) {
  const disabled = !hasResumes || selectedResumeId === null;

  const entries = [
    {
      icon: Sparkles,
      label: "推荐适合的岗位",
      desc: "基于简历能力定位，AI 推荐高匹配度岗位",
      href: selectedResumeId ? `/recommend?resumeId=${selectedResumeId}` : "#",
      disabled,
    },
    {
      icon: Search,
      label: "去岗位广场做匹配",
      desc: "浏览岗位后，可对任意岗位发起匹配分析",
      href: "/jobs",
      disabled: !hasResumes,
    },
    {
      icon: Wand2,
      label: "优化选中简历",
      desc: "AI 分析简历薄弱项，给出可执行优化建议",
      action: "optimize",
      disabled,
    },
    {
      icon: GitCompare,
      label: "查看版本对比",
      desc: "对比不同简历版本的变化与提升",
      action: "compare",
      disabled: !hasResumes || resumes.length < 2,
    },
  ];

  return (
    <div className="bg-white rounded-2xl border border-slate-100 shadow-sm p-5">
      <h2 className="text-[15px] font-semibold text-slate-800 mb-3">智能求职助手</h2>

      {/* Selected resume indicator */}
      <div className="mb-4 px-3 py-3 bg-slate-50 rounded-lg border border-slate-100">
        <p className="text-[11px] text-slate-400 mb-1 font-medium">当前选中简历</p>
        {loading ? (
          <p className="text-sm text-slate-400 animate-pulse">正在切换…</p>
        ) : selectedResumeName ? (
          <p className="text-sm font-semibold text-slate-800">{selectedResumeName}</p>
        ) : (
          <p className="text-sm text-slate-400">点击左侧列表中的简历</p>
        )}
      </div>

      <div className="space-y-2">
        {entries.map((entry) => (
          <div key={entry.label}>
            {entry.disabled ? (
              <div className="flex items-start gap-3 px-3 py-2.5 rounded-lg opacity-50 cursor-not-allowed">
                <entry.icon className="w-4 h-4 text-slate-400 mt-0.5 shrink-0" />
                <div>
                  <p className="text-sm font-medium text-slate-600">{entry.label}</p>
                  <p className="text-xs text-slate-400 mt-0.5">{entry.desc}</p>
                </div>
              </div>
            ) : "action" in entry && entry.action === "optimize" ? (
              <button
                onClick={onOptimize}
                className="flex items-start gap-3 px-3 py-2.5 rounded-lg hover:bg-blue-50/50 transition-colors group w-full text-left"
              >
                <entry.icon className="w-4 h-4 text-[#0052D9] mt-0.5 shrink-0 group-hover:scale-110 transition-transform" />
                <div>
                  <p className="text-sm font-medium text-slate-800 group-hover:text-[#0052D9] transition-colors">{entry.label}</p>
                  <p className="text-xs text-slate-400 mt-0.5">{entry.desc}</p>
                </div>
              </button>
            ) : "action" in entry && entry.action === "compare" ? (
              <button
                onClick={onCompare}
                className="flex items-start gap-3 px-3 py-2.5 rounded-lg hover:bg-blue-50/50 transition-colors group w-full text-left"
              >
                <entry.icon className="w-4 h-4 text-[#0052D9] mt-0.5 shrink-0 group-hover:scale-110 transition-transform" />
                <div>
                  <p className="text-sm font-medium text-slate-800 group-hover:text-[#0052D9] transition-colors">
                    {entry.label}
                  </p>
                  <p className="text-xs text-slate-400 mt-0.5">{entry.desc}</p>
                </div>
              </button>
            ) : (
              <Link
                to={entry.href}
                className="flex items-start gap-3 px-3 py-2.5 rounded-lg hover:bg-blue-50/50 transition-colors group"
              >
                <entry.icon className="w-4 h-4 text-[#0052D9] mt-0.5 shrink-0 group-hover:scale-110 transition-transform" />
                <div>
                  <p className="text-sm font-medium text-slate-800 group-hover:text-[#0052D9] transition-colors">
                    {entry.label}
                  </p>
                  <p className="text-xs text-slate-400 mt-0.5">{entry.desc}</p>
                </div>
              </Link>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}

// ─── Pagination Bar ───────────────────────────────────────────────────────────

function PaginationBar({
  page, totalPages, totalCount, onPrev, onNext,
}: {
  page: number; totalPages: number; totalCount: number;
  onPrev: () => void; onNext: () => void;
}) {
  if (totalPages <= 1) return null;
  return (
    <div className="flex items-center justify-between px-6 py-3 border-t border-slate-100 bg-slate-50/50">
      <span className="text-xs text-slate-500">共 {totalCount} 份简历</span>
      <div className="flex items-center gap-3">
        <span className="text-xs text-slate-500">第 {page} / {totalPages} 页</span>
        <div className="flex gap-1">
          <button
            onClick={onPrev}
            disabled={page <= 1}
            className="px-2.5 py-1 text-xs rounded border border-slate-200 bg-white text-slate-600 hover:bg-slate-100 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
          >
            上一页
          </button>
          <button
            onClick={onNext}
            disabled={page >= totalPages}
            className="px-2.5 py-1 text-xs rounded border border-slate-200 bg-white text-slate-600 hover:bg-slate-100 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
          >
            下一页
          </button>
        </div>
      </div>
    </div>
  );
}

// ─── Action Button ────────────────────────────────────────────────────────────

function ActionBtn({ children, onClick, to, className = "" }: {
  children: React.ReactNode;
  onClick?: (e: React.MouseEvent) => void;
  to?: string;
  className?: string;
}) {
  const base = `flex items-center gap-1 whitespace-nowrap text-[13px] text-slate-400 hover:text-[#0052D9] transition-colors cursor-pointer ${className}`;
  if (to) {
    return <Link to={to} onClick={onClick} className={base}>{children}</Link>;
  }
  return <button onClick={onClick} className={base}>{children}</button>;
}

// ─── Main Component ───────────────────────────────────────────────────────────

export function ResumeManagementPage() {
  const [rows, setRows] = useState<ResumeRow[] | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);
  const [selectedResumeId, setSelectedResumeId] = useState<number | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<ResumeRow | null>(null);
  const [deleteLoading, setDeleteLoading] = useState(false);
  const [assistantLoading, setAssistantLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [showOptimizationModal, setShowOptimizationModal] = useState(false);
  const [showCompareModal, setShowCompareModal] = useState(false);
  const [pickerOptions, setPickerOptions] = useState<ResumePickerOption[]>([]);
  const [pickerLoading, setPickerLoading] = useState(false);
  const [pickerSelectedIds, setPickerSelectedIds] = useState<number[]>([]);
  const navigate = useNavigate();
  const location = useLocation();

  // ── Responsive pageSize ─────────────────────────────────────────────────
  const getInitialPageSize = () => {
    if (typeof window === "undefined") return DESKTOP_PAGE_SIZE;
    return window.matchMedia("(min-width: 1024px)").matches
      ? DESKTOP_PAGE_SIZE
      : COMPACT_PAGE_SIZE;
  };
  const [pageSize, setPageSize] = useState(getInitialPageSize);

  useEffect(() => {
    const media = window.matchMedia("(min-width: 1024px)");
    const handleChange = () => {
      setPageSize(media.matches ? DESKTOP_PAGE_SIZE : COMPACT_PAGE_SIZE);
    };
    if (media.addEventListener) {
      media.addEventListener("change", handleChange);
    } else {
      media.addListener(handleChange);
    }
    return () => {
      if (media.removeEventListener) {
        media.removeEventListener("change", handleChange);
      } else {
        media.removeListener(handleChange);
      }
    };
  }, []);

  const handleCompare = () => {
    setPickerSelectedIds([]);
    setPickerLoading(true);
    setPickerOptions((rows ?? []).map((r) => ({ id: r.id, name: r.name })));
    setTimeout(() => setPickerLoading(false), 100);
    setShowCompareModal(true);
  };

  const handleSelectResume = (id: number) => {
    if (id === selectedResumeId) return;
    setSelectedResumeId(id);
    setAssistantLoading(true);
    // Silent sync to backend
    setCurrentResume(id).catch(() => { /* ignore */ });
    const timer = window.setTimeout(() => setAssistantLoading(false), 200);
    return () => clearTimeout(timer);
  };

  const handleDelete = async () => {
    if (!deleteTarget) return;
    setDeleteLoading(true);
    try {
      await deleteResume(deleteTarget.id);
      const list = await getResumes();
      const updated = sortRows(toRows(list));
      setRows(updated);
      if (list.length === 0) { setSelectedResumeId(null); setPage(1); return; }
      // Keep current page within bounds
      const maxPage = Math.max(1, Math.ceil(updated.length / pageSize));
      if (page > maxPage) setPage(maxPage);
    } catch { /* ignore */ }
    finally { setDeleteLoading(false); setDeleteTarget(null); }
  };

  useEffect(() => {
    getResumes()
      .then((resumes) => {
        const sorted = sortRows(toRows(resumes));
        setRows(sorted);
        if (sorted.length > 0) {
          const cur = sorted.find((r) => r.isCurrent) ?? sorted[0];
          setSelectedResumeId(cur.id);
        }
      })
      .catch((err: unknown) => {
        if (isApiClientError(err) && err.status === 401) {
          const redirect = getAuthRedirectIf401(err, location.pathname, location.search);
          if (redirect) { navigate(redirect, { replace: true }); return; }
        }
        setError(true);
      })
      .finally(() => setLoading(false));
  }, []);

  const hasResumes = useMemo(() => (rows && rows.length > 0) ?? false, [rows]);

  // Pagination
  const totalPages = Math.max(1, Math.ceil((rows?.length ?? 0) / pageSize));
  const currentPage = Math.min(page, totalPages);
  const pagedRows = (rows ?? []).slice((currentPage - 1) * pageSize, currentPage * pageSize);

  // K9-UI-ResumeManagement-B: correct currentPage when pageSize changes
  useEffect(() => {
    const maxPage = Math.max(1, Math.ceil((rows?.length ?? 0) / pageSize));
    setPage((p) => Math.min(p, maxPage));
  }, [rows?.length, pageSize]);

  if (loading) {
    return <PageLoading message="正在加载简历列表…" />;
  }

  if (error) {
    return (
      <ContentPageShell maxWidthClassName="max-w-[1200px]">
        <PageError title="简历列表加载失败，请稍后重试。" description="">
          <Link to="/resumes/edit" className="text-[#0052D9] text-sm hover:underline">去创建简历</Link>
        </PageError>
      </ContentPageShell>
    );
  }

  if (!rows || rows.length === 0) {
    return (
      <ContentPageShell maxWidthClassName="max-w-[1200px]">
        <EmptyState />
      </ContentPageShell>
    );
  }

  return (
    <>
    <ContentPageShell maxWidthClassName="max-w-[1200px]">

      <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-4 mb-6">
        <div>
          <h1 className="text-[22px] font-semibold text-slate-900">我的简历</h1>
          <p className="mt-1.5 text-sm text-slate-500">
            管理你的简历内容，用于岗位投递、匹配分析和简历优化。
          </p>
        </div>
        <Link
          to="/resumes/edit"
          className="inline-flex items-center gap-1.5 px-4 py-2 bg-[#0052D9] text-white text-sm font-medium rounded-lg hover:bg-[#003DA5] transition-colors shrink-0 self-start"
        >
          <Plus className="w-4 h-4" />
          添加简历
        </Link>
      </div>

      <div className="flex flex-col lg:flex-row gap-6">

        <div className="flex-[2] min-w-0">
          <div className="bg-white rounded-2xl border border-slate-100 shadow-sm overflow-hidden">
            <div className="px-6 py-3.5 border-b border-slate-100 flex items-center justify-between">
              <h2 className="text-sm font-semibold text-slate-700">简历列表</h2>
              <span className="text-xs text-slate-400">共 {rows.length} 份简历</span>
            </div>

            <div className="divide-y divide-slate-100">
              {pagedRows.map((row) => {
                const isSelected = selectedResumeId === row.id;
                return (
                  <div
                    key={row.id}
                    onClick={() => handleSelectResume(row.id)}
                    className={`relative flex items-center gap-4 px-6 py-4 cursor-pointer transition-colors ${
                      isSelected ? "bg-blue-50/70" : "hover:bg-slate-50"
                    }`}
                  >
                    {isSelected && (
                      <span className="absolute left-0 top-3 bottom-3 w-[3px] rounded-r-full bg-[#0052D9]" />
                    )}
                    {/* Name column: flex-1, truncate */}
                    <span className={`flex-1 min-w-0 text-[15px] font-semibold truncate ${
                      isSelected ? "text-[#0052D9]" : "text-slate-900"
                    }`}>
                      {row.name}
                    </span>
                    {/* Date column: fixed width */}
                    <span className="w-[5.5rem] shrink-0 text-xs text-slate-400 text-right tabular-nums">
                      {formatDate(row.createdAt)}
                    </span>
                    {/* Actions column: fixed width, right-aligned */}
                    <div className="w-[220px] shrink-0 flex items-center justify-end gap-3">
                      <ActionBtn to={`/resumes/detail/${row.id}`} onClick={(e) => e.stopPropagation()}>
                        <Eye className="w-3.5 h-3.5" />
                        查看详情
                      </ActionBtn>
                      <ActionBtn to={`/resumes/edit?id=${row.id}`} onClick={(e) => e.stopPropagation()}>
                        <Pencil className="w-3.5 h-3.5" />
                        编辑
                      </ActionBtn>
                      <ActionBtn
                        onClick={(e) => { e.stopPropagation(); setDeleteTarget(row); }}
                        className="hover:text-red-500"
                      >
                        <Trash2 className="w-3.5 h-3.5" />
                        删除
                      </ActionBtn>
                    </div>
                  </div>
                );
              })}
            </div>

            <PaginationBar
              page={currentPage} totalPages={totalPages} totalCount={rows.length}
              onPrev={() => setPage(Math.max(1, currentPage - 1))}
              onNext={() => setPage(Math.min(totalPages, currentPage + 1))}
            />
          </div>
        </div>

        <div className="flex-1 min-w-0 lg:max-w-[320px]">
          <SmartAssistant
            selectedResumeId={selectedResumeId}
            selectedResumeName={rows?.find((r) => r.id === selectedResumeId)?.name ?? ""}
            hasResumes={hasResumes}
            loading={assistantLoading}
            resumes={rows ?? []}
            onOptimize={() => setShowOptimizationModal(true)}
            onCompare={handleCompare}
          />
        </div>

      </div>

    </ContentPageShell>
      <AppConfirmDialog
        open={deleteTarget !== null}
        title="删除这份简历？"
        description={deleteTarget?.isCurrent ? "这份简历当前正在使用。删除后将无法恢复，系统会从剩余简历中选择最新的一份。" : "删除后无法恢复。"}
        confirmText="确认删除"
        cancelText="取消"
        variant="danger"
        loading={deleteLoading}
        onConfirm={handleDelete}
        onCancel={() => setDeleteTarget(null)}
      />
      <OptimizationEntryModal
        open={showOptimizationModal}
        onClose={() => setShowOptimizationModal(false)}
        initialResumeId={selectedResumeId}
      />
      <ResumePickerModal
        open={showCompareModal}
        title="选择两份简历进行版本对比"
        description="先选择上一版，再选择当前版。系统会分析当前版相比上一版的提升、风险和下一版优化建议。"
        resumes={pickerOptions}
        loading={pickerLoading}
        selectedIds={pickerSelectedIds}
        selectionMode="compare"
        onSelectionChange={setPickerSelectedIds}
        confirmText="开始对比"
        onConfirm={(ids) => { if (ids.length >= 2) navigate(`/resumes/compare?fromResumeId=${ids[0]}&toResumeId=${ids[1]}`); setShowCompareModal(false); }}
        onClose={() => setShowCompareModal(false)}
      />
    </>
  );
}

// ─── Sort helper ──────────────────────────────────────────────────────────────

function sortRows(rs: ResumeRow[]): ResumeRow[] {
  return [...rs].sort((a, b) => {
    const bTime = new Date(b.createdAt || 0).getTime();
    const aTime = new Date(a.createdAt || 0).getTime();
    return bTime - aTime;
  });
}
