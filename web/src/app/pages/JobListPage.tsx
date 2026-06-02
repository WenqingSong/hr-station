import { useCallback } from "react";
import { useSearchParams } from "react-router";
import { JobFilterSidebar, type JobFilters } from "../components/JobFilterSidebar";
import { JobList } from "../components/JobList";
import { AIDecisionWorkspace } from "../components/AIDecisionWorkspace";
import type { JobListParams } from "../../api/jobs";

const PAGE_SIZE = 10;

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
                next.set("page", "1");
                return next;
              });
            }}
          />
          <AIDecisionWorkspace />
        </div>
      </div>
    </div>
  );
}
