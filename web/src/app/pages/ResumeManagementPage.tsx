import { useState, useEffect } from "react";
import { Link, useNavigate, useLocation } from "react-router";
import { PageLoading, PageError } from "../components/PageState";
import { ContentPageShell } from "../components/layout/ContentPageShell";
import { FileText, Plus, Search, GitCompare, Wand2, Sparkles, Pencil } from "lucide-react";
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

type Scenario = "match" | "recommend" | "compare" | "optimize" | "other";

interface GroupDef {
  key: Scenario;
  label: string;
  icon: typeof Search;
  accentColor: string;
  badgeBg: string;
  badgeText: string;
  description: string;
}

// ─── Configuration ────────────────────────────────────────────────────────────

/** Default job IDs mapped by resume version_no prefix */
const DEFAULT_JOB_MAP: Record<number, number> = {
  1: 10003, // algorithm → NLP
  2: 10001, // backend dev
  3: 10016, // product manager
  8: 10001, // backend dev optimization
  9: 10001,
};

const GROUPS: GroupDef[] = [
  {
    key: "match",
    label: "岗位匹配分析",
    icon: Search,
    accentColor: "border-l-blue-500",
    badgeBg: "bg-blue-50",
    badgeText: "text-blue-700",
    description: "选择岗位后，分析简历与 JD 的匹配度和风险点。",
  },
  {
    key: "recommend",
    label: "岗位智能推荐",
    icon: Sparkles,
    accentColor: "border-l-emerald-500",
    badgeBg: "bg-emerald-50",
    badgeText: "text-emerald-700",
    description: "基于简历能力画像，推荐更适合投递的岗位。",
  },
  {
    key: "compare",
    label: "简历版本对比",
    icon: GitCompare,
    accentColor: "border-l-violet-500",
    badgeBg: "bg-violet-50",
    badgeText: "text-violet-700",
    description: "对比两版简历变化，识别提升点和潜在风险。",
  },
  {
    key: "optimize",
    label: "简历优化建议",
    icon: Wand2,
    accentColor: "border-l-orange-500",
    badgeBg: "bg-orange-50",
    badgeText: "text-orange-700",
    description: "针对目标岗位，生成可执行的简历补强建议。",
  },
];

// ─── Helpers ──────────────────────────────────────────────────────────────────

function getScenario(versionNo: number): Scenario {
  if (versionNo <= 3) return "match";
  if (versionNo <= 5) return "recommend";
  if (versionNo <= 7) return "compare";
  if (versionNo <= 9) return "optimize";
  return "other";
}

function getGroupDef(key: Scenario): GroupDef | undefined {
  return GROUPS.find((g) => g.key === key);
}

function toRows(versions: ResumeVersionListItem[]): ResumeRow[] {
  return versions
    .map((v) => ({
      id: v.id,
      name: v.name || "",
      versionNo: v.version_no,
    }))
    .sort((a, b) => a.versionNo - b.versionNo);
}

function groupRows(rows: ResumeRow[]): Map<Scenario, ResumeRow[]> {
  const map = new Map<Scenario, ResumeRow[]>();
  for (const r of rows) {
    const s = getScenario(r.versionNo);
    if (!map.has(s)) map.set(s, []);
    map.get(s)!.push(r);
  }
  return map;
}

function getMatchHref(row: ResumeRow): string {
  const jobId = DEFAULT_JOB_MAP[row.versionNo] ?? 10001;
  return `/match?jobId=${jobId}&resumeVersionNo=${row.versionNo}`;
}

function getRecHref(row: ResumeRow): string {
  return `/recommend?resumeVersionNo=${row.versionNo}`;
}

function getOptHref(row: ResumeRow): string {
  const jobId = DEFAULT_JOB_MAP[row.versionNo] ?? 10001;
  return `/resume-optimization?jobId=${jobId}&resumeVersionNo=${row.versionNo}`;
}

// ─── Components ───────────────────────────────────────────────────────────────

function EmptyState() {
  return (
    <div className="bg-white rounded-2xl ring-1 ring-gray-100 p-12 text-center">
      <FileText className="w-12 h-12 text-gray-300 mx-auto mb-4" />
      <h2 className="text-lg font-semibold text-gray-700 mb-2">你还没有创建简历</h2>
      <p className="text-sm text-gray-500 mb-6 max-w-md mx-auto">
        添加一份简历后，可以在这里管理简历并进入岗位推荐、匹配分析、简历优化和版本对比。
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

function AIFeatureCards() {
  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3 mb-10">
      {GROUPS.map((g) => (
        <div
          key={g.key}
          className="bg-white rounded-2xl ring-1 ring-gray-100 p-4 hover:shadow-[0_4px_16px_rgba(0,82,217,0.08)] transition-shadow cursor-default"
        >
          <div className="flex items-center gap-2 mb-1.5">
            <g.icon className="w-4 h-4 text-[#0052D9]" />
            <h3 className="text-sm font-semibold text-gray-800">{g.label}</h3>
          </div>
          <p className="text-xs text-gray-500 leading-relaxed">{g.description}</p>
        </div>
      ))}
    </div>
  );
}

function GroupSection({
  group,
  rows,
  index,
}: {
  group: GroupDef;
  rows: ResumeRow[];
  index: number;
}) {
  // Compare group: single action button for both rows
  if (group.key === "compare") {
    const from = rows[0]?.versionNo;
    const to = rows[1]?.versionNo;
    const compareHref = from && to ? `/resumes/compare?from=${from}&to=${to}` : "/resumes/compare";
    return (
      <div className="mb-6">
        <div className="flex items-center justify-between mb-3">
          <div className="flex items-center gap-2.5">
            <div className={`w-1 h-5 rounded-full bg-violet-500`} />
            <h2 className="text-[15px] font-semibold text-gray-800">{group.label}</h2>
            <span className={`${group.badgeBg} ${group.badgeText} text-[11px] font-medium px-2 py-0.5 rounded-full`}>
              {rows.length} 份
            </span>
          </div>
          <Link
            to={compareHref}
            className="text-xs font-medium text-[#0052D9] hover:text-[#003DA5] transition-colors"
          >
            对比 {from ?? "?"} 与 {to ?? "?"} →
          </Link>
        </div>
        <div className="bg-white rounded-2xl ring-1 ring-gray-100 divide-y divide-gray-100 overflow-hidden">
          {rows.map((row) => (
            <div key={row.id} className="flex items-center justify-between px-5 py-4 hover:bg-gray-50 transition-colors group">
              <span className="text-sm font-medium text-gray-900 truncate">{row.name}</span>
              <Link
                to="/resumes/edit"
                className="flex items-center gap-1 text-xs text-gray-400 hover:text-[#0052D9] transition-colors ml-3"
              >
                <Pencil className="w-3 h-3" />
                编辑
              </Link>
            </div>
          ))}
        </div>
      </div>
    );
  }

  // Match / Recommend / Optimize: action button per row
  const accentColors: Record<string, string> = {
    match: "bg-blue-500",
    recommend: "bg-emerald-500",
    optimize: "bg-orange-500",
  };
  const getHref = (row: ResumeRow): string => {
    switch (group.key) {
      case "match": return getMatchHref(row);
      case "recommend": return getRecHref(row);
      case "optimize": return getOptHref(row);
      default: return "#";
    }
  };
  const getBtnLabel = (): string => {
    switch (group.key) {
      case "match": return "去匹配岗位";
      case "recommend": return "推荐岗位";
      case "optimize": return "优化简历";
      default: return "";
    }
  };
  const btnLabel = getBtnLabel();
  const accentColor = accentColors[group.key] ?? "bg-gray-500";

  return (
    <div className="mb-6">
      <div className="flex items-center gap-2.5 mb-3">
        <div className={`w-1 h-5 rounded-full ${accentColor}`} />
        <h2 className="text-[15px] font-semibold text-gray-800">{group.label}</h2>
        <span className={`${group.badgeBg} ${group.badgeText} text-[11px] font-medium px-2 py-0.5 rounded-full`}>
          {rows.length} 份
        </span>
      </div>
      <div className="bg-white rounded-2xl ring-1 ring-gray-100 divide-y divide-gray-100 overflow-hidden">
        {rows.map((row) => (
          <div key={row.id} className="flex items-center justify-between px-5 py-4 hover:bg-gray-50 transition-colors group">
            <span className="text-sm font-medium text-gray-900 truncate">{row.name}</span>
            <div className="flex items-center gap-2 ml-3">
              <Link
                to={getHref(row)}
                className="inline-flex items-center px-3.5 py-1.5 bg-[#0052D9] text-white text-xs font-medium rounded-lg hover:bg-[#003DA5] transition-colors"
              >
                {btnLabel}
              </Link>
              <Link
                to="/resumes/edit"
                className="flex items-center gap-1 text-xs text-gray-400 hover:text-[#0052D9] transition-colors"
              >
                <Pencil className="w-3 h-3" />
                编辑
              </Link>
            </div>
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
  const navigate = useNavigate();
  const location = useLocation();

  useEffect(() => {
    getResumeVersions()
      .then((versions) => setRows(toRows(versions)))
      .catch((err: unknown) => {
        if (isApiClientError(err) && err.status === 401) {
          const redirect = getAuthRedirectIf401(err, location.pathname, location.search);
          if (redirect) { navigate(redirect, { replace: true }); return; }
        }
        setError(true);
      })
      .finally(() => setLoading(false));
  }, []);

  // Loading
  if (loading) {
    return (
      <ContentPageShell maxWidthClassName="max-w-[1040px]">
        <PageLoading message="正在加载简历列表…" />
      </ContentPageShell>
    );
  }

  // Error
  if (error) {
    return (
      <ContentPageShell maxWidthClassName="max-w-[1040px]">
        <PageError title="简历数据加载失败，请稍后重试。" description="">
          <Link to="/resumes/edit" className="text-[#0052D9] text-sm hover:underline">去创建简历</Link>
        </PageError>
      </ContentPageShell>
    );
  }

  // Empty
  if (!rows || rows.length === 0) {
    return (
      <ContentPageShell maxWidthClassName="max-w-[1040px]">
        <EmptyState />
      </ContentPageShell>
    );
  }

  const grouped = groupRows(rows);

  return (
    <ContentPageShell maxWidthClassName="max-w-[1040px]">

      {/* ── Title Area ──────────────────────────────────────────────────── */}
      <div className="bg-white rounded-2xl ring-1 ring-gray-100 p-6 mb-6">
        <div className="flex items-start justify-between">
          <div>
            <h1 className="text-[22px] font-semibold text-gray-900">我的简历</h1>
            <p className="mt-1.5 text-sm text-gray-500">
              选择一份示范简历，体验岗位推荐、岗位匹配、简历优化和版本对比
            </p>
          </div>
          <Link
            to="/resumes/edit"
            className="flex items-center gap-1.5 px-4 py-2 bg-[#0052D9] text-white text-sm font-medium rounded-lg hover:bg-[#003DA5] transition-colors flex-shrink-0"
          >
            <Plus className="w-4 h-4" />
            添加简历
          </Link>
        </div>
      </div>

      {/* ── AI Feature Overview ───────────────────────────────────────────── */}
      <AIFeatureCards />

      {/* ── Grouped Resume List ────────────────────────────────────────────── */}
      {GROUPS.map((g, i) => {
        const groupRows = grouped.get(g.key);
        if (!groupRows || groupRows.length === 0) return null;
        return <GroupSection key={g.key} group={g} rows={groupRows} index={i} />;
      })}

    </ContentPageShell>
  );
}
