import React, { useState, useEffect } from "react";
import { Search, Check, X } from "lucide-react";
import { getJobFacets, type JobFacets } from "../../api/jobs";

export interface JobFilters {
  direction?: string;
  city?: string;
  recruit_type?: string;
  category?: string;
  business_group?: string;
  skill?: string;
  application_project?: string;
  application_subtype?: string;
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
  const [localKeyword, setLocalKeyword] = useState(keyword);

  useEffect(() => { setLocalKeyword(keyword); }, [keyword]);

  useEffect(() => {
    getJobFacets()
      .then(setFacets)
      .catch(() => setFacetsError("筛选条件加载失败"))
      .finally(() => setFacetsLoading(false));
  }, []);

  const toggleFilter = (key: keyof JobFilters, value: string) => {
    onFiltersChange({ ...filters, [key]: filters[key] === value ? undefined : value });
  };

  const handleSearch = () => { onKeywordChange(localKeyword); };
  const handleKeyDown = (e: React.KeyboardEvent) => { if (e.key === "Enter") handleSearch(); };

  const activeCount = Object.values(filters).filter(Boolean).length;

  return (
    <aside className="w-[260px] shrink-0 sticky top-24 self-start bg-white border border-gray-200 rounded-lg">
      {/* Search */}
      <div className="p-4 pb-3 border-b border-gray-100">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
          <input
            type="text"
            placeholder="搜索岗位 / 技能关键词"
            value={localKeyword}
            onChange={(e) => setLocalKeyword(e.target.value)}
            onKeyDown={handleKeyDown}
            className="w-full pl-9 pr-3 py-2 text-sm border border-gray-300 rounded-md focus:outline-none focus:ring-1 focus:ring-[#0052D9] focus:border-[#0052D9] placeholder:text-gray-400"
          />
        </div>
      </div>

      {/* Active filter chips */}
      {activeCount > 0 && (
        <div className="px-4 py-2 border-b border-gray-100">
          <div className="flex items-center justify-between mb-2">
            <span className="text-xs text-gray-500">已选 {activeCount} 项</span>
            <button
              className="text-xs text-[#0052D9] hover:underline"
              onClick={() => onFiltersChange({})}
            >
              清空筛选
            </button>
          </div>
          <div className="flex flex-wrap gap-1.5">
            {Object.entries(filters).map(([key, val]) =>
              val ? (
                <span key={key} className="inline-flex items-center gap-1 text-xs bg-blue-50 text-[#0052D9] px-2 py-0.5 rounded-full cursor-pointer hover:bg-blue-100 transition-colors"
                  onClick={() => toggleFilter(key as keyof JobFilters, val)}>
                  {val}
                  <X className="h-3 w-3" />
                </span>
              ) : null
            )}
          </div>
        </div>
      )}

      {/* Facets */}
      <div className="p-4 pt-3">
        {facetsLoading && (
          <p className="text-xs text-gray-400 text-center py-8">正在加载筛选项…</p>
        )}
        {facetsError && (
          <p className="text-xs text-gray-400 text-center py-8">{facetsError}</p>
        )}
        {facets && (
          <div className="space-y-5">
            {facets.application_projects && facets.application_projects.length > 0 && (
              <FilterGroup
                title="应聘项目"
                items={facets.application_projects}
                selectedValue={filters.application_project}
                onSelect={(v) => toggleFilter("application_project", v)}
              />
            )}
            {facets.application_subtypes && facets.application_subtypes.length > 0 && (
              <FilterGroup
                title="应聘类型"
                items={facets.application_subtypes}
                selectedValue={filters.application_subtype}
                onSelect={(v) => toggleFilter("application_subtype", v)}
              />
            )}
            {facets.job_types && facets.job_types.length > 0 && (
              <FilterGroup
                title="岗位类型"
                items={facets.job_types}
                selectedValue={filters.recruit_type}
                onSelect={(v) => toggleFilter("recruit_type", v)}
              />
            )}
            {facets.directions.length > 0 && (
              <FilterGroup
                title="求职方向"
                items={facets.directions}
                selectedValue={filters.direction}
                onSelect={(v) => toggleFilter("direction", v)}
              />
            )}
            {facets.cities.length > 0 && (
              <FilterGroup
                title="工作地点"
                items={facets.cities}
                selectedValue={filters.city}
                onSelect={(v) => toggleFilter("city", v)}
              />
            )}
            {facets.recruit_types.length > 0 && (
              <FilterGroup
                title="招聘类型"
                items={facets.recruit_types}
                selectedValue={filters.recruit_type}
                onSelect={(v) => toggleFilter("recruit_type", v)}
              />
            )}
            {facets.categories.length > 0 && (
              <FilterGroup
                title="岗位类别"
                items={facets.categories}
                selectedValue={filters.category}
                onSelect={(v) => toggleFilter("category", v)}
              />
            )}
            {facets.business_groups && facets.business_groups.length > 0 && (
              <FilterGroup
                title="事业群"
                items={facets.business_groups}
                selectedValue={filters.business_group}
                onSelect={(v) => toggleFilter("business_group", v)}
              />
            )}
            {facets.skills && facets.skills.length > 0 && (
              <FilterGroup
                title="技能关键词"
                items={facets.skills.slice(0, 20)}
                selectedValue={filters.skill}
                onSelect={(v) => toggleFilter("skill", v)}
              />
            )}
            {!facets.directions.length && !facets.cities.length && !facets.recruit_types.length && !facets.categories.length && !(facets.business_groups||[]).length && !(facets.skills||[]).length && (
              <p className="text-xs text-gray-400 text-center py-4">暂无筛选项</p>
            )}
          </div>
        )}
      </div>
    </aside>
  );
}

function FilterGroup({
  title,
  items,
  selectedValue,
  onSelect,
}: {
  title: string;
  items: { value: string; label: string; count: number }[];
  selectedValue?: string;
  onSelect: (value: string) => void;
}) {
  return (
    <div>
      <h3 className="text-xs font-semibold text-gray-500 uppercase tracking-wide mb-3">{title}</h3>
      <div className="space-y-0.5">
        {items.map((item) => {
          const isSelected = selectedValue === item.value;
          return (
            <button
              key={item.value}
              onClick={() => onSelect(item.value)}
              className={`w-full flex items-center justify-between px-2 py-1.5 rounded-md text-sm transition-colors ${
                isSelected
                  ? "bg-blue-50 text-[#0052D9] font-medium"
                  : "text-gray-700 hover:bg-gray-50"
              }`}
            >
              <span className="flex items-center gap-2 truncate">
                {isSelected && <Check className="h-3.5 w-3.5 shrink-0" />}
                <span className="truncate">{item.label}</span>
              </span>
              {item.count > 0 && (
                <span className={`text-xs shrink-0 ml-2 ${isSelected ? "text-[#0052D9]" : "text-gray-400"}`}>
                  {item.count}
                </span>
              )}
            </button>
          );
        })}
      </div>
    </div>
  );
}
