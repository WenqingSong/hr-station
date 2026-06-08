import React, {
  useEffect,
  useState,
  useMemo,
  useRef,
  useCallback,
} from "react";
import { createPortal } from "react-dom";
import { Link } from "react-router";
import {
  X,
  Clock,
  Briefcase,
  Search,
  RefreshCw,
  FilePlus,
  ChevronRight,
} from "lucide-react";
import { Skeleton } from "../ui/skeleton";

// ─── Types ────────────────────────────────────────────────────────────────────

export type ResumePickerSelectionMode = "single" | "compare";

export interface ResumePickerOption {
  id: number;
  name: string;
  isCurrent?: boolean;
  statusLabel?: string;
  directionLabel?: string;
  dominantProfile?: string;
  tags?: string[];
  projectCount?: number;
  updatedAt?: string;
}

export interface ResumePickerModalProps {
  open: boolean;

  title: string;
  description?: string;
  targetLabel?: string;

  resumes: ResumePickerOption[];
  loading?: boolean;
  error?: string | null;
  onRetry?: () => void;

  selectedIds: number[];
  selectionMode?: ResumePickerSelectionMode;
  maxSelected?: number;
  onSelectionChange: (ids: number[]) => void;

  confirmText: string;
  confirming?: boolean;
  confirmDisabled?: boolean;
  onConfirm: (ids: number[]) => void;

  onClose: () => void;

  searchPlaceholder?: string;
  footerLeft?: React.ReactNode;
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

function formatDateTime(value?: string | null): string {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`;
}

const TITLE_ID = "resume-picker-title";

// ─── Component ────────────────────────────────────────────────────────────────

export function ResumePickerModal({
  open,
  title,
  description,
  targetLabel,
  resumes,
  loading = false,
  error = null,
  onRetry,
  selectedIds,
  selectionMode = "single",
  maxSelected,
  onSelectionChange,
  confirmText,
  confirming = false,
  confirmDisabled = false,
  onConfirm,
  onClose,
  searchPlaceholder = "搜索简历名、方向或技能关键词…",
  footerLeft,
}: ResumePickerModalProps) {
  const [searchQuery, setSearchQuery] = useState("");
  const overlayRef = useRef<HTMLDivElement>(null);

  const resolvedMax = maxSelected ?? (selectionMode === "compare" ? 2 : 1);

  // ── Reset search on open/close ───────────────────────────────────────
  useEffect(() => {
    if (!open) setSearchQuery("");
  }, [open]);

  // ── Body scroll lock ─────────────────────────────────────────────────
  useEffect(() => {
    if (!open) return;

    const prevOverflow = document.body.style.overflow;
    const prevPaddingRight = document.body.style.paddingRight;
    const scrollbarWidth = window.innerWidth - document.documentElement.clientWidth;

    document.body.style.overflow = "hidden";
    if (scrollbarWidth > 0) {
      document.body.style.paddingRight = `${scrollbarWidth}px`;
    }

    return () => {
      document.body.style.overflow = prevOverflow;
      document.body.style.paddingRight = prevPaddingRight;
    };
  }, [open]);

  // ── ESC close ────────────────────────────────────────────────────────
  useEffect(() => {
    if (!open) return;
    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [open, onClose]);

  // ── Click overlay to close ───────────────────────────────────────────
  const handleOverlayMouseDown = useCallback(
    (e: React.MouseEvent) => {
      if (e.target === overlayRef.current) onClose();
    },
    [onClose],
  );

  // ── Search filter ────────────────────────────────────────────────────
  const filteredResumes = useMemo(() => {
    if (!searchQuery.trim()) return resumes;
    const q = searchQuery.trim().toLowerCase();
    return resumes.filter(
      (r) =>
        r.name.toLowerCase().includes(q) ||
        (r.directionLabel || "").toLowerCase().includes(q) ||
        (r.dominantProfile || "").toLowerCase().includes(q) ||
        (r.tags || []).some((t) => t.toLowerCase().includes(q)),
    );
  }, [resumes, searchQuery]);

  // ── Toggle selection ─────────────────────────────────────────────────
  const toggleSelection = useCallback(
    (id: number) => {
      const current = new Set(selectedIds);
      if (current.has(id)) {
        current.delete(id);
      } else {
        if (current.size >= resolvedMax) {
          // In single mode: replace; in compare mode: replace last
          if (selectionMode === "single") {
            current.clear();
          } else {
            // Remove first to keep max 2 (replace older)
            const arr = Array.from(current);
            current.delete(arr[0]);
          }
        }
        current.add(id);
      }
      onSelectionChange(Array.from(current));
    },
    [selectedIds, resolvedMax, selectionMode, onSelectionChange],
  );

  // ── Confirm state ────────────────────────────────────────────────────
  const minRequired = selectionMode === "compare" ? 2 : 1;
  const canConfirm =
    !loading &&
    !confirming &&
    !error &&
    selectedIds.length >= minRequired &&
    !confirmDisabled;

  if (!open) return null;

  return createPortal(
    <div className="fixed inset-0 z-[10000]" ref={overlayRef}>
      {/* ── Overlay ── */}
      <div
        className="absolute inset-0 bg-slate-950/40 backdrop-blur-sm"
        onMouseDown={handleOverlayMouseDown}
      />

      {/* ── Panel container ── */}
      <div className="relative z-[10001] flex min-h-full items-center justify-center px-4 py-8">
        <section
          role="dialog"
          aria-modal="true"
          aria-labelledby={TITLE_ID}
          className="flex w-full max-w-3xl flex-col overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-2xl"
          style={{ maxHeight: "min(760px, calc(100vh - 64px))" }}
        >
          {/* ── Header ── */}
          <header className="shrink-0 border-b border-slate-100 px-6 pt-5 pb-4">
            <div className="flex items-start justify-between">
              <div>
                <h2 id={TITLE_ID} className="font-semibold text-slate-900 text-base">
                  {title}
                </h2>
                {(description || targetLabel) && (
                  <p className="text-[13px] text-slate-500 mt-1">
                    {description}
                    {targetLabel && (
                      <>
                        {" "}— 目标岗位：<span className="text-slate-700 font-medium">{targetLabel}</span>
                      </>
                    )}
                  </p>
                )}
              </div>
              <button
                onClick={onClose}
                className="p-1.5 text-slate-300 hover:text-slate-500 transition-colors flex-shrink-0 ml-3 rounded-lg hover:bg-slate-100"
                aria-label="关闭"
              >
                <X className="w-4 h-4" />
              </button>
            </div>
          </header>

          {/* ── Search ── */}
          <div className="shrink-0 px-6 pt-4 pb-3">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
              <input
                type="text"
                placeholder={searchPlaceholder}
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full pl-9 pr-4 py-2 text-[13px] bg-slate-50 border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-100 focus:border-blue-300 placeholder:text-slate-400"
              />
            </div>
          </div>

          {/* ── Body / List ── */}
          <div className="min-h-0 flex-1 overflow-y-auto px-6 py-4">
            {loading ? (
              <div className="py-12 space-y-3">
                {[1, 2, 3].map((i) => (
                  <div key={i} className="h-20 bg-slate-100 rounded-lg animate-pulse" />
                ))}
                <p className="text-center text-[13px] text-slate-400 pt-2">
                  正在读取你的简历…
                </p>
              </div>
            ) : error ? (
              <div className="py-10 text-center">
                <p className="text-[13px] text-slate-600 mb-1">{error}</p>
                {onRetry && (
                  <button
                    onClick={onRetry}
                    className="mt-2 inline-flex items-center gap-1 text-[13px] text-[#0052D9] hover:underline"
                  >
                    <RefreshCw className="w-3.5 h-3.5" /> 重试
                  </button>
                )}
              </div>
            ) : filteredResumes.length === 0 ? (
              searchQuery.trim() ? (
                <div className="py-10 text-center">
                  <p className="text-[13px] text-slate-500 mb-1">没有找到匹配的简历</p>
                  <p className="text-[12px] text-slate-400 mb-3">
                    换个关键词，或清空搜索条件。
                  </p>
                  <button
                    onClick={() => setSearchQuery("")}
                    className="text-[13px] text-[#0052D9] hover:underline"
                  >
                    清空搜索
                  </button>
                </div>
              ) : (
                <div className="py-10 text-center">
                  <p className="text-[13px] text-slate-500 mb-1">
                    还没有可用于分析的简历
                  </p>
                  <p className="text-[12px] text-slate-400 mb-3">
                    请先添加或完善一份简历，再进行操作。
                  </p>
                  <Link
                    to="/resumes/new"
                    onClick={onClose}
                    className="inline-flex items-center gap-1.5 text-[13px] text-[#0052D9] hover:underline font-medium"
                  >
                    <FilePlus className="w-3.5 h-3.5" /> 去添加简历
                  </Link>
                </div>
              )
            ) : (
              <div className="space-y-2 pb-2">
                {filteredResumes.map((resume) => {
                  const idx = selectedIds.indexOf(resume.id);
                  const isSelected = idx !== -1;

                  return (
                    <div
                      key={resume.id}
                      onClick={() => toggleSelection(resume.id)}
                      className={`flex items-start gap-3 px-4 py-3 border rounded-xl cursor-pointer transition-all ${
                        isSelected
                          ? "border-[#0052D9] bg-blue-50/50 shadow-sm"
                          : "border-slate-200 bg-white hover:border-blue-200 hover:bg-slate-50"
                      }`}
                    >
                      {/* Selection indicator */}
                      <div className="flex-shrink-0 mt-1">
                        {selectionMode === "compare" ? (
                          <div
                            className={`w-5 h-5 rounded-md border-2 flex items-center justify-center text-[11px] font-semibold transition-colors ${
                              isSelected
                                ? "border-[#0052D9] bg-[#0052D9] text-white"
                                : "border-slate-300 text-slate-300"
                            }`}
                          >
                            {isSelected ? idx + 1 : ""}
                          </div>
                        ) : (
                          <div
                            className={`w-4 h-4 rounded-full border-2 flex items-center justify-center transition-colors ${
                              isSelected
                                ? "border-[#0052D9]"
                                : "border-slate-300"
                            }`}
                          >
                            {isSelected && (
                              <div className="w-2 h-2 rounded-full bg-[#0052D9]" />
                            )}
                          </div>
                        )}
                      </div>

                      {/* Info */}
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2 mb-1.5 flex-wrap">
                          <span className="text-[14px] font-semibold text-slate-900">
                            {resume.name}
                          </span>

                        </div>

                        <div className="flex items-center gap-2 text-[12px] text-slate-500 mb-1.5 flex-wrap">
                          {resume.directionLabel && (
                            <>
                              <span className="flex items-center gap-1">
                                <Briefcase className="w-3 h-3" />
                                {resume.directionLabel}
                              </span>
                              <span className="w-px h-3 bg-slate-200" />
                            </>
                          )}
                          {resume.tags && resume.tags.length > 0 && (
                            <span className="text-slate-400">
                              {resume.tags.slice(0, 3).join(" · ")}
                            </span>
                          )}
                        </div>

                        <div className="flex items-center gap-2 text-[11px] text-slate-400">
                          {resume.projectCount != null && resume.projectCount > 0 && (
                            <span>{resume.projectCount} 个项目</span>
                          )}
                          {resume.projectCount != null &&
                            resume.projectCount > 0 &&
                            formatDateTime(resume.updatedAt) && <span className="w-px h-3 bg-slate-200" />}
                          {formatDateTime(resume.updatedAt) && (
                            <span className="flex items-center gap-1">
                              <Clock className="w-3 h-3" />
                              {formatDateTime(resume.updatedAt)}
                            </span>
                          )}
                        </div>
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </div>

          {/* ── Footer ── */}
          <footer className="shrink-0 border-t border-slate-100 bg-white px-6 py-4">
            <div className="flex items-center justify-between">
              {footerLeft || (
                <Link
                  to="/resumes"
                  onClick={onClose}
                  className="flex items-center gap-1 text-[12px] text-slate-400 hover:text-[#0052D9] transition-colors"
                >
                  管理我的简历
                  <ChevronRight className="w-3 h-3" />
                </Link>
              )}

              <div className="flex items-center gap-3">
                <button
                  onClick={onClose}
                  className="px-4 py-2 text-[13px] text-slate-500 hover:text-slate-800 transition-colors rounded-lg hover:bg-slate-100"
                >
                  取消
                </button>
                <button
                  onClick={() => onConfirm(selectedIds)}
                  disabled={!canConfirm}
                  className={`px-5 py-2 text-[13px] font-medium rounded-lg transition-colors ${
                    canConfirm
                      ? "bg-[#0052D9] text-white hover:bg-[#003DA5] cursor-pointer"
                      : "bg-slate-200 text-slate-400 cursor-not-allowed"
                  }`}
                >
                  {confirming ? "正在跳转…" : confirmText}
                </button>
              </div>
            </div>
          </footer>
        </section>
      </div>
    </div>,
    document.body,
  );
}
