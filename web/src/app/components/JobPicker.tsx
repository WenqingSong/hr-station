import React, { useState, useEffect, useCallback, useRef } from "react";
import { Search, Briefcase, MapPin, Building2, Tag, RefreshCw, ChevronLeft, ChevronRight } from "lucide-react";
import type { JobApiItem } from "../../api/jobs";
import { getJobs } from "../../api/jobs";

// ─── Types ────────────────────────────────────────────────────────────────────

export interface JobPickerProps {
  selectedJobId?: number | null;
  onSelect: (job: JobApiItem) => void;
  emptyMessage?: string;
}

// ─── Constants ────────────────────────────────────────────────────────────────

const PAGE_SIZE = 10;

// ─── Helpers ──────────────────────────────────────────────────────────────────

/** Generate pagination page numbers with ellipsis, same logic as JobPagination. */
function generatePages(current: number, total: number): (number | "...")[] {
  if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1);

  const pages: (number | "...")[] = [];
  if (current <= 4) {
    for (let i = 1; i <= 6; i++) pages.push(i);
    pages.push("...", total);
  } else if (current >= total - 3) {
    pages.push(1, "...");
    for (let i = total - 5; i <= total; i++) pages.push(i);
  } else {
    pages.push(1, "...");
    for (let i = current - 2; i <= current + 2; i++) pages.push(i);
    pages.push("...", total);
  }
  return pages;
}

// ─── Component ────────────────────────────────────────────────────────────────

export function JobPicker({
  selectedJobId,
  onSelect,
  emptyMessage = "没有找到匹配的岗位",
}: JobPickerProps) {
  const [jobs, setJobs] = useState<JobApiItem[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [keyword, setKeyword] = useState("");
  const [page, setPage] = useState(1);
  const debounceRef = useRef<ReturnType<typeof setTimeout>>();

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));

  // ── Reusable loader (server-side) ──
  const loadJobs = useCallback(
    (params: { keyword?: string; page?: number; page_size?: number }) => {
      setLoading(true);
      setError(null);
      getJobs(params)
        .then((res) => {
          setJobs(res.items || []);
          setTotal(res.total || 0);
        })
        .catch(() => setError("岗位列表加载失败"))
        .finally(() => setLoading(false));
    },
    [],
  );

  // ── Initial load ──
  useEffect(() => {
    loadJobs({ page: 1, page_size: PAGE_SIZE });
  }, [loadJobs]);

  // ── Debounced server-side keyword search ──
  useEffect(() => {
    if (debounceRef.current) clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(() => {
      setPage(1);
      loadJobs({ keyword: keyword.trim() || undefined, page: 1, page_size: PAGE_SIZE });
    }, 300);
    return () => {
      if (debounceRef.current) clearTimeout(debounceRef.current);
    };
  }, [keyword, loadJobs]);

  // ── Page change ──
  const handlePageChange = (newPage: number) => {
    if (newPage < 1 || newPage > totalPages) return;
    setPage(newPage);
    loadJobs({ keyword: keyword.trim() || undefined, page: newPage, page_size: PAGE_SIZE });
  };

  // ── States ──
  if (loading) {
    return (
      <div className="py-6 text-center">
        <RefreshCw className="w-4 h-4 animate-spin text-slate-400 mx-auto mb-2" />
        <p className="text-[13px] text-slate-400">正在加载岗位列表…</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="py-6 text-center">
        <p className="text-[13px] text-slate-500 mb-1">{error}</p>
        <button
          onClick={() => loadJobs({ page: 1, page_size: PAGE_SIZE })}
          className="text-[13px] text-[#0052D9] hover:underline"
        >
          重试
        </button>
      </div>
    );
  }

  if (total === 0 && !loading) {
    return (
      <div className="py-6 text-center">
        <p className="text-[13px] text-slate-500 mb-1">暂无可用岗位</p>
        <p className="text-[12px] text-slate-400">请稍后重试</p>
      </div>
    );
  }

  return (
    <div className="flex flex-col min-h-0">
      {/* ── Search ── */}
      <div className="shrink-0 px-1 pb-3">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
          <input
            type="text"
            placeholder="搜索岗位名称、方向、城市、部门…"
            value={keyword}
            onChange={(e) => setKeyword(e.target.value)}
            className="w-full pl-9 pr-4 py-2 text-[13px] bg-slate-50 border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-100 focus:border-blue-300 placeholder:text-slate-400"
          />
        </div>
      </div>

      {/* ── List ── */}
      <div className="flex-1 min-h-0 overflow-y-auto space-y-1.5">
        {jobs.length === 0 ? (
          <div className="py-8 text-center">
            <p className="text-[13px] text-slate-500 mb-1">{emptyMessage}</p>
            {keyword.trim() && (
              <button
                onClick={() => setKeyword("")}
                className="text-[12px] text-[#0052D9] hover:underline"
              >
                清空搜索
              </button>
            )}
          </div>
        ) : (
          jobs.map((job) => {
            const isSelected = selectedJobId === job.id;
            return (
              <div
                key={job.id}
                onClick={() => onSelect(job)}
                className={`flex items-start gap-3 px-3 py-2.5 border rounded-lg cursor-pointer transition-all ${
                  isSelected
                    ? "border-[#0052D9] bg-blue-50/50 shadow-sm"
                    : "border-slate-200 bg-white hover:border-blue-200 hover:bg-slate-50"
                }`}
              >
                {/* Selection indicator */}
                <div className="flex-shrink-0 mt-0.5">
                  <div
                    className={`w-4 h-4 rounded-full border-2 flex items-center justify-center transition-colors ${
                      isSelected ? "border-[#0052D9]" : "border-slate-300"
                    }`}
                  >
                    {isSelected && <div className="w-2 h-2 rounded-full bg-[#0052D9]" />}
                  </div>
                </div>

                {/* Info */}
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-1 flex-wrap">
                    <span className="text-[14px] font-semibold text-slate-900 truncate">
                      {job.title}
                    </span>
                    {job.recruit_type && (
                      <span className="px-1.5 py-0.5 bg-blue-50 text-[#0052D9] text-[10px] rounded font-medium shrink-0">
                        {job.recruit_type}
                      </span>
                    )}
                  </div>
                  <div className="flex items-center gap-2 text-[12px] text-slate-500 flex-wrap">
                    {job.city && (
                      <span className="flex items-center gap-1">
                        <MapPin className="w-3 h-3" />
                        {job.city}
                      </span>
                    )}
                    {job.direction && (
                      <span className="flex items-center gap-1">
                        <Briefcase className="w-3 h-3" />
                        {job.direction}
                      </span>
                    )}
                    {job.department && (
                      <span className="flex items-center gap-1">
                        <Building2 className="w-3 h-3" />
                        {job.department}
                      </span>
                    )}
                    {job.category && (
                      <span className="flex items-center gap-1">
                        <Tag className="w-3 h-3" />
                        {job.category}
                      </span>
                    )}
                  </div>
                </div>
              </div>
            );
          })
        )}
      </div>

      {/* ── Pagination ── */}
      {totalPages > 1 && (
        <div className="shrink-0 pt-3">
          <div className="flex items-center justify-between text-[12px] text-slate-400">
            <span>共 {total} 个岗位</span>
            <div className="inline-flex items-center gap-0.5">
              {/* Prev */}
              <button
                disabled={page <= 1}
                onClick={() => handlePageChange(page - 1)}
                className={`flex items-center justify-center w-6 h-6 rounded transition-colors ${
                  page <= 1
                    ? "text-slate-300 cursor-not-allowed"
                    : "text-slate-500 hover:bg-slate-100"
                }`}
                aria-label="上一页"
              >
                <ChevronLeft className="w-3.5 h-3.5" />
              </button>

              {/* Page numbers */}
              {generatePages(page, totalPages).map((item, idx) => {
                if (item === "...") {
                  return (
                    <span
                      key={`ellipsis-${idx}`}
                      className="flex items-center justify-center w-6 h-6 text-slate-400 select-none"
                    >
                      ...
                    </span>
                  );
                }
                const isActive = item === page;
                return (
                  <button
                    key={item}
                    onClick={() => handlePageChange(item)}
                    className={`flex items-center justify-center w-6 h-6 text-[12px] rounded transition-colors ${
                      isActive
                        ? "bg-[#0052D9] text-white"
                        : "text-slate-600 hover:bg-slate-100"
                    }`}
                    aria-current={isActive ? "page" : undefined}
                  >
                    {item}
                  </button>
                );
              })}

              {/* Next */}
              <button
                disabled={page >= totalPages}
                onClick={() => handlePageChange(page + 1)}
                className={`flex items-center justify-center w-6 h-6 rounded transition-colors ${
                  page >= totalPages
                    ? "text-slate-300 cursor-not-allowed"
                    : "text-slate-500 hover:bg-slate-100"
                }`}
                aria-label="下一页"
              >
                <ChevronRight className="w-3.5 h-3.5" />
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
