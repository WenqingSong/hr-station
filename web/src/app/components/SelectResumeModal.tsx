import React, { useEffect, useState, useMemo } from "react";
import { Link } from "react-router";
import {
  X,
  CheckCircle2,
  Clock,
  Briefcase,
  ChevronRight,
  Search,
  RefreshCw,
  FilePlus,
} from "lucide-react";
import { loadResumeOptionViewModelsFromApi } from "../../api";
import type { ResumeOptionViewModel } from "../../adapters";
import { getAuthRedirectIf401 } from "../utils/authRedirect";
import { useNavigate, useLocation } from "react-router";

// ─── Types ────────────────────────────────────────────────────────────────────

type Scene = "recommendation" | "match-analysis" | "resume-optimization";

interface SelectResumeModalProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm?: (resumeId: string) => void;
  scene?: Scene;
  jobTitle?: string;
  /** K9-2-Entry-B-Fix: 初始选中简历 ID（来自父组件当前简历） */
  initialSelectedResumeId?: number | null;
}

const SCENE_CONFIG: Record<Scene, { label: string; desc: string }> = {
  "recommendation": {
    label: "岗位推荐",
    desc: "系统将从当前岗位池中推荐最值得优先关注的 5 个岗位。",
  },
  "match-analysis": {
    label: "岗位匹配分析",
    desc: "系统将对比岗位 JD 与所选简历，生成匹配度、证据链和投递建议。",
  },
  "resume-optimization": {
    label: "简历优化",
    desc: "系统将基于目标岗位 JD 分析当前简历，给出具体的修改建议和优化方案。",
  },
};

// ─── Component ────────────────────────────────────────────────────────────────

export function SelectResumeModal({
  isOpen,
  onClose,
  onConfirm,
  scene = "match-analysis",
  jobTitle = "",
  initialSelectedResumeId = null,
}: SelectResumeModalProps) {
  const [selected, setSelected] = useState<number | null>(null);
  const [resumeOptions, setResumeOptions] = useState<ResumeOptionViewModel[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const navigate = useNavigate();
  const location = useLocation();

  // ── Load resumes on open ─────────────────────────────────────────────
  useEffect(() => {
    if (!isOpen) {
      setSelected(null);
      setSearchQuery("");
      setError(null);
      setSubmitting(false);
      return;
    }
    setLoading(true);
    setError(null);
    loadResumeOptionViewModelsFromApi()
      .then((opts) => {
        setResumeOptions(opts);
        // Default priority: initialSelectedResumeId > isCurrent > first
        if (opts.length > 0) {
          const initial = initialSelectedResumeId != null ? opts.find(o => o.id === initialSelectedResumeId) : undefined;
          if (initial) {
            setSelected(initial.id);
          } else {
            const current = opts.find(o => o.isCurrent);
            setSelected(current ? current.id : opts[0].id);
          }
        } else {
          setSelected(null);
        }
      })
      .catch((err) => {
        const authRedirect = getAuthRedirectIf401(err, location.pathname, location.search);
        if (authRedirect) { navigate(authRedirect, { replace: true }); return; }
        setError("简历列表加载失败，请稍后重试。");
      })
      .finally(() => setLoading(false));
  }, [isOpen, initialSelectedResumeId]);

  // ── Search ───────────────────────────────────────────────────────────
  const filteredOptions = useMemo(() => {
    let list = resumeOptions;
    if (searchQuery.trim()) {
      const q = searchQuery.trim().toLowerCase();
      list = list.filter(o =>
        o.name.toLowerCase().includes(q) ||
        o.directionLabel.toLowerCase().includes(q) ||
        (o.tags || []).some(t => t.toLowerCase().includes(q))
      );
    }
    return list;
  }, [resumeOptions, searchQuery]);

  if (!isOpen) return null;

  const sceneInfo = SCENE_CONFIG[scene];
  const canConfirm = !loading && !submitting && !error && selected !== null && filteredOptions.some(o => o.id === selected);

  function handleConfirm() {
    if (!canConfirm || selected === null) return;
    setSubmitting(true);
    try {
      onConfirm?.(String(selected));
    } finally {
      // onConfirm navigates away
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/30 px-4 py-8">
      {/* Modal */}
      <div className="relative bg-white w-full max-w-2xl z-10 shadow-lg border border-slate-200 rounded-2xl flex flex-col max-h-[85vh]">

        {/* ── Header ────────────────────────────────────────────────── */}
        <div className="flex items-start justify-between px-6 pt-5 pb-4 border-b border-slate-100 flex-shrink-0">
          <div>
            <h2 className="font-semibold text-slate-900 text-base">选择用于分析的简历</h2>
            <p className="text-[13px] text-slate-500 mt-1">
              系统将基于所选简历进行{sceneInfo.label}
              {jobTitle && <> — 目标岗位：<span className="text-slate-700 font-medium">{jobTitle}</span></>}
            </p>
          </div>
          <button onClick={onClose} className="p-1.5 text-slate-300 hover:text-slate-500 transition-colors flex-shrink-0 ml-3 rounded-lg hover:bg-slate-100">
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* ── Search ────────────────────────────────────────────────── */}
        <div className="px-6 pt-4 pb-3 flex-shrink-0">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
            <input
              type="text"
              placeholder="搜索简历名、方向或技能关键词…"
              value={searchQuery}
              onChange={e => setSearchQuery(e.target.value)}
              className="w-full pl-9 pr-4 py-2 text-[13px] bg-slate-50 border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-100 focus:border-blue-300 placeholder:text-slate-400"
            />
          </div>
        </div>

        {/* ── Content ────────────────────────────────────────────────── */}
        <div className="px-6 pb-2 overflow-y-auto flex-1 min-h-0">
          {loading ? (
            <div className="py-12 space-y-3">
              {[1,2,3].map(i => (
                <div key={i} className="h-20 bg-slate-100 rounded-lg animate-pulse" />
              ))}
              <p className="text-center text-[13px] text-slate-400 pt-2">正在读取你的简历…</p>
            </div>
          ) : error ? (
            <div className="py-10 text-center">
              <p className="text-[13px] text-slate-600 mb-1">{error}</p>
              <button
                onClick={() => { setError(null); setLoading(true); loadResumeOptionViewModelsFromApi().then(opts => { setResumeOptions(opts); if (opts.length > 0) { const initial = initialSelectedResumeId != null ? opts.find(o => o.id === initialSelectedResumeId) : undefined; setSelected(initial ? initial.id : (opts.find(o=>o.isCurrent) || opts[0]).id); } else setSelected(null); }).catch(() => setError("简历列表加载失败，请稍后重试。")).finally(() => setLoading(false)); }}
                className="mt-2 inline-flex items-center gap-1 text-[13px] text-[#0052D9] hover:underline"
              >
                <RefreshCw className="w-3.5 h-3.5" /> 重试
              </button>
            </div>
          ) : filteredOptions.length === 0 ? (
            searchQuery.trim() ? (
              <div className="py-10 text-center">
                <p className="text-[13px] text-slate-500 mb-1">没有找到匹配的简历</p>
                <p className="text-[12px] text-slate-400 mb-3">换个关键词，或清空搜索条件。</p>
                <button onClick={() => setSearchQuery("")} className="text-[13px] text-[#0052D9] hover:underline">清空搜索</button>
              </div>
            ) : (
              <div className="py-10 text-center">
                <p className="text-[13px] text-slate-500 mb-1">还没有可用于分析的简历</p>
                <p className="text-[12px] text-slate-400 mb-3">请先添加或完善一份简历，再进行岗位匹配分析。</p>
                <Link to="/resumes/new" onClick={onClose} className="inline-flex items-center gap-1.5 text-[13px] text-[#0052D9] hover:underline font-medium">
                  <FilePlus className="w-3.5 h-3.5" /> 去添加简历
                </Link>
              </div>
            )
          ) : (
            <div className="space-y-2 pb-2">
              {filteredOptions.map((resume) => {
                const isSelected = selected === resume.id;
                return (
                  <div
                    key={resume.id}
                    onClick={() => setSelected(resume.id)}
                    className={`flex items-start gap-3 px-4 py-3 border rounded-xl cursor-pointer transition-all ${
                      isSelected
                        ? "border-[#0052D9] bg-blue-50/50 shadow-sm"
                        : "border-slate-200 bg-white hover:border-blue-200 hover:bg-slate-50"
                    }`}
                  >
                    {/* Radio */}
                    <div className="flex-shrink-0 mt-1">
                      <div className={`w-4 h-4 rounded-full border-2 flex items-center justify-center transition-colors ${isSelected ? "border-[#0052D9]" : "border-slate-300"}`}>
                        {isSelected && <div className="w-2 h-2 rounded-full bg-[#0052D9]" />}
                      </div>
                    </div>

                    {/* Info */}
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 mb-1.5 flex-wrap">
                        <span className="text-[14px] font-semibold text-slate-900">{resume.name}</span>
                        {resume.isCurrent && (
                          <span className="inline-flex items-center gap-1 px-1.5 py-0.5 bg-blue-50 border border-blue-100 text-[#0052D9] text-[11px] rounded font-medium">
                            <CheckCircle2 className="w-3 h-3" /> 当前使用
                          </span>
                        )}
                        {resume.statusLabel && !resume.isCurrent && (
                          <span className="px-1.5 py-0.5 bg-slate-100 text-slate-500 text-[11px] rounded">{resume.statusLabel}</span>
                        )}
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
                          <span className="text-slate-400">{resume.tags.slice(0, 3).join(" · ")}</span>
                        )}
                      </div>

                      <div className="flex items-center gap-2 text-[11px] text-slate-400">
                        {resume.projectCount > 0 && <span>{resume.projectCount} 个项目</span>}
                        {resume.projectCount > 0 && resume.updatedAt && <span className="w-px h-3 bg-slate-200" />}
                        {resume.updatedAt && (
                          <span className="flex items-center gap-1">
                            <Clock className="w-3 h-3" />
                            {resume.updatedAt}
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

        {/* ── Footer ────────────────────────────────────────────────── */}
        <div className="flex items-center justify-between px-6 py-4 border-t border-slate-100 flex-shrink-0 bg-white rounded-b-2xl">
          <Link
            to="/resumes"
            onClick={onClose}
            className="flex items-center gap-1 text-[12px] text-slate-400 hover:text-[#0052D9] transition-colors"
          >
            管理我的简历
            <ChevronRight className="w-3 h-3" />
          </Link>

          <div className="flex items-center gap-3">
            <button
              onClick={onClose}
              className="px-4 py-2 text-[13px] text-slate-500 hover:text-slate-800 transition-colors rounded-lg hover:bg-slate-100"
            >
              取消
            </button>
            <button
              onClick={handleConfirm}
              disabled={!canConfirm}
              className={`px-5 py-2 text-[13px] font-medium rounded-lg transition-colors ${
                canConfirm
                  ? "bg-[#0052D9] text-white hover:bg-[#003DA5] cursor-pointer"
                  : "bg-slate-200 text-slate-400 cursor-not-allowed"
              }`}
            >
              {submitting ? "正在跳转…" : "确认并开始分析"}
            </button>
          </div>
        </div>

      </div>
    </div>
  );
}
