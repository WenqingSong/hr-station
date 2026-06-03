import { useCallback, useMemo } from "react";
import { useSearchParams } from "react-router";
import { JobFilterSidebar, type JobFilters } from "../components/jobs/JobFilterSidebar";
import { JobList } from "../components/jobs/JobList";
import { JobHeroBanner } from "../components/jobs/JobHeroBanner";
import { AIDecisionWorkspace } from "../components/AIDecisionWorkspace";
import { AISidebarShell } from "../layouts/AISidebarShell";
import type { JobListParams } from "../../api/jobs";

const PAGE_SIZE = 10;

// K8-4-4-R: 筛选键到中文组名的映射
const GROUP_NAMES: Record<string, string> = {
  keyword: '搜索',
  direction: '求职方向',
  city: '工作城市',
  recruit_type: '岗位类型',
  category: '岗位类别',
  business_group: '事业群',
  skill: '技能关键词',
  application_project: '应聘项目',
  application_subtype: '应聘类型',
  job_category: '岗位类别',
};

/** 将 filters + keyword 转为显示用的标签列表（支持逗号多选，含分组中文前缀） */
function filtersToDisplay(keyword: string, filters: JobFilters): { key: string; label: string; group: string }[] {
  const entries: { key: string; label: string; group: string }[] = [];

  // 搜索关键词优先展示
  if (keyword) {
    entries.push({ key: 'keyword', label: keyword, group: GROUP_NAMES.keyword });
  }

  // 筛选条件按固定顺序处理
  const keys: (keyof JobFilters)[] = [
    'application_project', 'application_subtype', 'job_category',
    'direction', 'recruit_type', 'category',
    'city', 'business_group', 'skill',
  ];
  for (const key of keys) {
    const val = filters[key];
    if (typeof val === 'string' && val) {
      const groupName = GROUP_NAMES[key] || key;
      val.split(',').forEach(v => {
        if (v.trim()) entries.push({ key, label: v.trim(), group: groupName });
      });
    }
  }
  return entries;
}

export function JobListPage() {
  const [searchParams, setSearchParams] = useSearchParams();

  const keyword = searchParams.get("keyword") || "";
  const filters: JobFilters = {
    direction: searchParams.get("direction") || undefined,
    city: searchParams.get("city") || undefined,
    recruit_type: searchParams.get("recruit_type") || undefined,
    category: searchParams.get("category") || undefined,
    business_group: searchParams.get("business_group") || undefined,
    skill: searchParams.get("skill") || undefined,
    application_project: searchParams.get("application_project") || undefined,
    application_subtype: searchParams.get("application_subtype") || undefined,
    job_category: searchParams.get("job_category") || undefined,
  };
  const page = parseInt(searchParams.get("page") || "1", 10) || 1;

  const setKeyword = useCallback((v: string) => {
    setSearchParams((prev) => { const p = new URLSearchParams(prev); if (v) p.set("keyword", v); else p.delete("keyword"); p.set("page", "1"); return p; });
  }, [setSearchParams]);

  const setFilters = useCallback((f: JobFilters) => {
    setSearchParams((prev) => {
      const p = new URLSearchParams(prev);
      const keys: (keyof JobFilters)[] = ['direction','city','recruit_type','category','business_group','skill','application_project','application_subtype','job_category'];
      for (const k of keys) {
        if (f[k]) p.set(k, f[k]);
        else p.delete(k);
      }
      p.set("page", "1");
      return p;
    });
  }, [setSearchParams]);

  const setPage = useCallback((pg: number) => {
    setSearchParams((prev) => { const p = new URLSearchParams(prev); p.set("page", String(pg)); return p; });
  }, [setSearchParams]);

  // K8-4-4-R: 清除全部同时清除关键词和所有筛选条件
  const clearFilters = useCallback(() => {
    setSearchParams(new URLSearchParams({ page: "1" }));
  }, [setSearchParams]);

  // K8-4-4-R: 将 keyword 一起纳入顶部标签栏展示
  const activeFilterLabels = useMemo(() => filtersToDisplay(keyword, filters), [keyword, filters]);

  const params: JobListParams = {
    keyword: keyword || undefined,
    direction: filters.direction,
    city: filters.city,
    recruit_type: filters.recruit_type,
    category: filters.category,
    business_group: filters.business_group,
    skill: filters.skill,
    application_project: filters.application_project,
    application_subtype: filters.application_subtype,
    job_category: filters.job_category,
    page,
    page_size: PAGE_SIZE,
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="mx-auto max-w-[1440px] px-6 md:px-8 py-6">
        {/* K8-4-4-R12：顶部 Hero Banner */}
        <JobHeroBanner />

        <div className="flex gap-6 items-start mt-6">
          <JobFilterSidebar
            keyword={keyword}
            onKeywordChange={setKeyword}
            filters={filters}
            onFiltersChange={setFilters}
          />
          <JobList
            params={params}
            onPageChange={setPage}
            onParamsChange={(p) => {
              setSearchParams((prev) => {
                const next = new URLSearchParams(prev);
                // 清除关键词
                if ('keyword' in p) {
                  if (p.keyword) next.set("keyword", p.keyword);
                  else next.delete("keyword");
                }
                // 清除其他筛选参数
                const allKeys: (keyof JobFilters)[] = ['direction','city','recruit_type','category','business_group','skill','application_project','application_subtype','job_category'];
                for (const k of allKeys) {
                  if (k in p) {
                    const v = p[k as keyof typeof p];
                    if (v) next.set(k, v as string);
                    else next.delete(k);
                  }
                }
                next.set("page", "1");
                return next;
              });
            }}
            activeFilters={activeFilterLabels}
            onClearFilters={clearFilters}
          />
          <AISidebarShell className="lg:w-[300px] lg:shrink-0">
            <AIDecisionWorkspace />
          </AISidebarShell>
        </div>
      </div>
    </div>
  );
}
