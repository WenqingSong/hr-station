import React, { useState } from "react";
import { JobFilterSidebar, type JobFilters } from "../components/JobFilterSidebar";
import { JobList } from "../components/JobList";
import { AIDecisionWorkspace } from "../components/AIDecisionWorkspace";
import type { JobListParams } from "../../api/jobs";

export function JobListPage() {
  const [keyword, setKeyword] = useState("");
  const [filters, setFilters] = useState<JobFilters>({});

  const params: JobListParams = {
    keyword: keyword || undefined,
    direction: filters.direction,
    city: filters.city,
    recruit_type: filters.recruit_type,
    category: filters.category,
  };

  return (
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
          onParamsChange={(p) => {
            setKeyword(p.keyword || "");
            setFilters({ direction: p.direction, city: p.city, recruit_type: p.recruit_type, category: p.category });
          }}
        />
        <AIDecisionWorkspace />
      </div>
    </div>
  );
}
