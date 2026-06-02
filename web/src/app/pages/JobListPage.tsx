import { useCallback, useMemo } from "react";
import { useSearchParams } from "react-router";
import { JobFilterSidebar, type JobFilters } from "../components/JobFilterSidebar";
import { JobList } from "../components/JobList";
import { AIDecisionWorkspace } from "../components/AIDecisionWorkspace";
import type { JobListParams } from "../../api/jobs";

const PAGE_SIZE = 10;

/** 将 filters 转为显示用的标签列表 */
function filtersToDisplay(filters: JobFilters): { key: string; label: string }[] {
  const entries: { key: string; label: string }[] = [];
  if (typeof filters.direction === "string" && filters.direction) entries.push({ key: "direction", label: filters.direction });
  if (typeof filters.city === "string" && filters.city) entries.push({ key: "city", label: filters.city });
  if (typeof filters.recruit_type === "string" && filters.recruit_type) entries.push({ key: "recruit_type", label: filters.recruit_type });
  if (typeof filters.category === "string" && filters.category) entries.push({ key: "category", label: filters.category });
  if (typeof filters.business_group === "string" && filters.business_group) entries.push({ key: "business_group", label: filters.business_group });
  if (typeof filters.skill === "string" && filters.skill) entries.push({ key: "skill", label: filters.skill });
  if (typeof filters.application_project === "string" && filters.application_project) entries.push({ key: "application_project", label: filters.application_project });
  if (typeof filters.application_subtype === "string" && filters.application_subtype) entries.push({ key: "application_subtype", label: filters.application_subtype });
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
  };
  const page = parseInt(searchParams.get("page") || "1", 10) || 1;

  const setKeyword = useCallback((v: string) => {
    setSearchParams((prev) => { const p = new URLSearchParams(prev); if (v) p.set("keyword", v); else p.delete("keyword"); p.set("page", "1"); return p; });
  }, [setSearchParams]);

  const setFilters = useCallback((f: JobFilters) => {
    setSearchParams((prev) => {
      const p = new URLSearchParams(prev);
      if (f.direction) p.set("direction", f.direction); else p.delete("direction");
      if (f.city) p.set("city", f.city); else p.delete("city");
      if (f.recruit_type) p.set("recruit_type", f.recruit_type); else p.delete("recruit_type");
      if (f.category) p.set("category", f.category); else p.delete("category");
      if (f.business_group) p.set("business_group", f.business_group); else p.delete("business_group");
      if (f.skill) p.set("skill", f.skill); else p.delete("skill");
      if (f.application_project) p.set("application_project", f.application_project); else p.delete("application_project");
      if (f.application_subtype) p.set("application_subtype", f.application_subtype); else p.delete("application_subtype");
      p.set("page", "1");
      return p;
    });
  }, [setSearchParams]);

  const setPage = useCallback((pg: number) => {
    setSearchParams((prev) => { const p = new URLSearchParams(prev); p.set("page", String(pg)); return p; });
  }, [setSearchParams]);

  const clearFilters = useCallback(() => {
    setSearchParams((prev) => { const p = new URLSearchParams(prev); p.set("page", "1"); return p; });
  }, [setSearchParams]);

  const activeFilterLabels = useMemo(() => filtersToDisplay(filters), [filters]);

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
    page,
    page_size: PAGE_SIZE,
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="mx-auto max-w-[1440px] px-6 md:px-8 py-6">
        <div className="flex gap-6 items-start">
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
                if (p.keyword) next.set("keyword", p.keyword); else next.delete("keyword");
                // Handle single filter clear
                const cleared = Object.keys(p).filter((k) => p[k as keyof JobListParams] === undefined);
                cleared.forEach((k) => next.delete(k));
                next.set("page", "1");
                return next;
              });
            }}
            activeFilters={activeFilterLabels}
            onClearFilters={clearFilters}
          />
          <AIDecisionWorkspace />
        </div>
      </div>
    </div>
  );
}
