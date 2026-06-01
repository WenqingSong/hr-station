import React, { useState, useEffect } from "react";
import { Search } from "lucide-react";
import { Checkbox } from "./ui/Checkbox";
import { getJobFacets, type JobFacets } from "../../api/jobs";

export interface JobFilters {
  direction?: string;
  city?: string;
  recruit_type?: string;
  category?: string;
}

interface JobFilterSidebarProps {
  keyword: string;
  onKeywordChange: (v: string) => void;
  filters: JobFilters;
  onFiltersChange: (f: JobFilters) => void;
}

export function JobFilterSidebar({ keyword, onKeywordChange, filters, onFiltersChange }: JobFilterSidebarProps) {
  const [facets, setFacets] = useState<JobFacets | null>(null);
  const [facetsLoading, setFacetsLoading] = useState(true);
  const [facetsError, setFacetsError] = useState<string | null>(null);

  useEffect(() => {
    getJobFacets()
      .then(setFacets)
      .catch((e) => setFacetsError(e instanceof Error ? e.message : "加载筛选失败"))
      .finally(() => setFacetsLoading(false));
  }, []);

  const toggleFilter = (key: keyof JobFilters, value: string) => {
    const newVal = filters[key] === value ? undefined : value;
    onFiltersChange({ ...filters, [key]: newVal });
  };

  const isActive = (key: keyof JobFilters, value: string) => filters[key] === value;

  return (
    <aside className="w-[260px] shrink-0 sticky top-24 self-start bg-white border border-gray-200 p-4 rounded-md">
      <div className="relative mb-6">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
        <input
          type="text"
          placeholder="搜索岗位 / 技能关键词"
          value={keyword}
          onChange={(e) => onKeywordChange(e.target.value)}
          className="w-full pl-9 pr-3 py-2 text-sm border border-gray-300 rounded-md focus:outline-none focus:ring-1 focus:ring-[#0052D9] focus:border-[#0052D9] placeholder:text-gray-400"
        />
      </div>

      {facetsLoading && (
        <p className="text-xs text-gray-400 text-center py-4">正在加载筛选项…</p>
      )}
      {facetsError && (
        <p className="text-xs text-gray-400 text-center py-4">筛选加载失败</p>
      )}
      {facets && (
        <div className="space-y-6">
          {facets.directions.length > 0 && (
            <FilterGroup
              title="求职方向"
              items={facets.directions}
              activeValue={filters.direction}
              onToggle={(v) => toggleFilter("direction", v)}
            />
          )}
          {facets.cities.length > 0 && (
            <FilterGroup
              title="工作地点"
              items={facets.cities}
              activeValue={filters.city}
              onToggle={(v) => toggleFilter("city", v)}
            />
          )}
          {facets.recruit_types.length > 0 && (
            <FilterGroup
              title="招聘类型"
              items={facets.recruit_types}
              activeValue={filters.recruit_type}
              onToggle={(v) => toggleFilter("recruit_type", v)}
            />
          )}
          {facets.categories.length > 0 && (
            <FilterGroup
              title="岗位类别"
              items={facets.categories}
              activeValue={filters.category}
              onToggle={(v) => toggleFilter("category", v)}
            />
          )}
          {!facets.directions.length && !facets.cities.length && !facets.recruit_types.length && !facets.categories.length && (
            <p className="text-xs text-gray-400 text-center py-4">暂无筛选项</p>
          )}
        </div>
      )}
    </aside>
  );
}

function FilterGroup({
  title,
  items,
  activeValue,
  onToggle,
}: {
  title: string;
  items: { value: string; label: string; count: number }[];
  activeValue?: string;
  onToggle: (value: string) => void;
}) {
  return (
    <div>
      <div className="flex items-center justify-between mb-3">
        <h3 className="text-sm font-bold text-gray-900">{title}</h3>
      </div>
      <div className="space-y-3">
        {items.map((item) => (
          <div
            key={item.value}
            className="flex items-center justify-between group cursor-pointer"
            onClick={() => onToggle(item.value)}
          >
            <div className="flex items-center gap-2">
              <Checkbox
                id={`${title}-${item.value}`}
                checked={activeValue === item.value}
                onChange={() => onToggle(item.value)}
              />
              <label
                htmlFor={`${title}-${item.value}`}
                className="text-sm text-gray-600 cursor-pointer group-hover:text-[#0052D9]"
              >
                {item.label}
              </label>
            </div>
            <span className="text-xs text-gray-400">{item.count > 0 ? item.count : ""}</span>
          </div>
        ))}
      </div>
    </div>
  );
}
