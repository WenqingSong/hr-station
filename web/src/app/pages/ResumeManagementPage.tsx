import { useState, useEffect, useMemo } from "react";
import { Link, useNavigate, useLocation } from "react-router";
import { PageLoading, PageError } from "../components/PageState";
import { ContentPageShell } from "../components/layout/ContentPageShell";
import { FileText, Plus, Search, Sparkles, Wand2, GitCompare, Pencil } from "lucide-react";
import { getResumeVersions } from "../../api/resumes";
import type { ResumeVersionListItem } from "../../types";
import { getAuthRedirectIf401 } from "../utils/authRedirect";
import { isApiClientError } from "../../api/errors";

// ─── Types ────────────────────────────────────────────────────────────────────

interface ResumeRow {
  id: number;
  name: string;
  versionNo: number;
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

function toRows(versions: ResumeVersionListItem[]): ResumeRow[] {
  return versions
    .map((v) => ({
      id: v.id,
      name: v.name || "",
      versionNo: v.version_no,
    }))
    .sort((a, b) => b.versionNo - a.versionNo); // 最新在前
}

// ─── Sub-components ───────────────────────────────────────────────────────────

function EmptyState() {
  return (
    <div className="bg-white rounded-2xl ring-1 ring-gray-100 p-12 text-center">
      <FileText className="w-12 h-12 text-gray-300 mx-auto mb-4" />
      <h2 className="text-lg font-semibold text-gray-700 mb-2">你还没有创建简历</h2>
      <p className="text-sm text-gray-500 mb-6 max-w-md mx-auto">
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
  selectedVersionNo: number | null;
  hasResumes: boolean;
}

function SmartAssistant({ selectedVersionNo, hasResumes }: SmartAssistantProps) {
  const disabled = !hasResumes || selectedVersionNo === null;

  const entries = [
    {
      icon: Sparkles,
      label: "推荐适合的岗位",
      desc: "基于简历能力画像，AI 推荐高匹配度岗位",
      href: selectedVersionNo ? `/recommend?resumeVersionNo=${selectedVersionNo}` : "#",
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
      label: "优化当前简历",
      desc: "AI 分析简历薄弱项，给出可执行优化建议",
      href: selectedVersionNo ? `/resume-optimization?resumeVersionNo=${selectedVersionNo}` : "#",
      disabled,
    },
    {
      icon: GitCompare,
      label: "查看版本对比",
      desc: "对比不同简历版本的变化与提升",
      href: "/resumes/compare",
      disabled: !hasResumes,
    },
  ];

  return (
    <div className="bg-white rounded-2xl ring-1 ring-gray-100 p-5">
      <h2 className="text-[15px] font-semibold text-gray-800 mb-1">智能求职助手</h2>
      <p className="text-xs text-gray-400 mb-4 leading-relaxed">
        选择一份简历后，可使用智能推荐、匹配分析和优化建议。
      </p>
      <div className="space-y-2">
        {entries.map((entry) => (
          <div key={entry.label}>
            {entry.disabled ? (
              <div className="flex items-start gap-3 px-3 py-2.5 rounded-lg opacity-50 cursor-not-allowed">
                <entry.icon className="w-4 h-4 text-gray-400 mt-0.5 shrink-0" />
                <div>
                  <p className="text-sm font-medium text-gray-600">{entry.label}</p>
                  <p className="text-xs text-gray-400 mt-0.5">{entry.desc}</p>
                </div>
              </div>
            ) : (
              <Link
                to={entry.href}
                className="flex items-start gap-3 px-3 py-2.5 rounded-lg hover:bg-blue-50/50 transition-colors group"
              >
                <entry.icon className="w-4 h-4 text-[#0052D9] mt-0.5 shrink-0 group-hover:scale-110 transition-transform" />
                <div>
                  <p className="text-sm font-medium text-gray-800 group-hover:text-[#0052D9] transition-colors">
                    {entry.label}
                  </p>
                  <p className="text-xs text-gray-400 mt-0.5">{entry.desc}</p>
                </div>
              </Link>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}

// ─── Main Component ───────────────────────────────────────────────────────────

export function ResumeManagementPage() {
  const [rows, setRows] = useState<ResumeRow[] | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);
  const [selectedVersionNo, setSelectedVersionNo] = useState<number | null>(null);
  const navigate = useNavigate();
  const location = useLocation();

  useEffect(() => {
    getResumeVersions()
      .then((versions) => {
        const sorted = toRows(versions);
        setRows(sorted);
        if (sorted.length > 0) {
          setSelectedVersionNo(sorted[0].versionNo);
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

  // ── Loading ──
  if (loading) {
    return (
      <ContentPageShell maxWidthClassName="max-w-[1040px]">
        <PageLoading message="正在加载简历列表…" />
      </ContentPageShell>
    );
  }

  // ── Error ──
  if (error) {
    return (
      <ContentPageShell maxWidthClassName="max-w-[1040px]">
        <PageError title="简历列表加载失败，请稍后重试。" description="">
          <Link to="/resumes/edit" className="text-[#0052D9] text-sm hover:underline">去创建简历</Link>
        </PageError>
      </ContentPageShell>
    );
  }

  // ── Empty ──
  if (!rows || rows.length === 0) {
    return (
      <ContentPageShell maxWidthClassName="max-w-[1040px]">
        <EmptyState />
      </ContentPageShell>
    );
  }

  return (
    <ContentPageShell maxWidthClassName="max-w-[1040px]">

      {/* ── Header ──────────────────────────────────────────────────────── */}
      <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-4 mb-6">
        <div>
          <h1 className="text-[22px] font-semibold text-gray-900">我的简历</h1>
          <p className="mt-1.5 text-sm text-gray-500">
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

      {/* ── Two-column Layout ────────────────────────────────────────────── */}
      <div className="flex flex-col lg:flex-row gap-6">

        {/* ── Left: Resume List ──────────────────────────────────────── */}
        <div className="flex-[2] min-w-0">
          <div className="bg-white rounded-2xl ring-1 ring-gray-100 overflow-hidden">
            {/* List header */}
            <div className="px-5 py-3.5 border-b border-gray-100 flex items-center justify-between">
              <h2 className="text-sm font-semibold text-gray-700">简历列表</h2>
              <span className="text-xs text-gray-400">共 {rows.length} 份简历</span>
            </div>

            {/* List rows */}
            <div className="divide-y divide-gray-100">
              {rows.map((row) => {
                const isSelected = selectedVersionNo === row.versionNo;
                return (
                  <div
                    key={row.id}
                    onClick={() => setSelectedVersionNo(row.versionNo)}
                    className={`flex items-center justify-between px-5 py-4 cursor-pointer transition-colors group ${
                      isSelected
                        ? "bg-blue-50/60 border-l-[3px] border-l-[#0052D9] pl-[17px]"
                        : "hover:bg-gray-50 border-l-[3px] border-l-transparent pl-[17px]"
                    }`}
                  >
                    <span className={`text-sm font-medium truncate ${
                      isSelected ? "text-[#0052D9]" : "text-gray-900"
                    }`}>
                      {row.name}
                    </span>
                    <Link
                      to={`/resumes/edit?versionNo=${row.versionNo}`}
                      onClick={(e) => e.stopPropagation()}
                      className="flex items-center gap-1 text-xs text-gray-400 hover:text-[#0052D9] transition-colors ml-3 shrink-0"
                    >
                      <Pencil className="w-3 h-3" />
                      编辑
                    </Link>
                  </div>
                );
              })}
            </div>
          </div>
        </div>

        {/* ── Right: Smart Assistant ────────────────────────────────────── */}
        <div className="flex-1 min-w-0 lg:max-w-[320px]">
          <SmartAssistant selectedVersionNo={selectedVersionNo} hasResumes={hasResumes} />
        </div>

      </div>

    </ContentPageShell>
  );
}
