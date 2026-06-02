import React, { useState, useEffect, useCallback } from "react";
import { Search, Check, X, ChevronUp, ChevronDown } from "lucide-react";
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

/** 安全获取数组，防止 null/undefined 导致崩溃 */
function safeItems<T>(v: T[] | null | undefined): T[] {
  return Array.isArray(v) ? v : [];
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

  const toggleFilter = useCallback((key: keyof JobFilters, value: string) => {
    onFiltersChange({ ...filters, [key]: filters[key] === value ? undefined : value });
  }, [filters, onFiltersChange]);

  const handleSearch = useCallback(() => { onKeywordChange(localKeyword); }, [localKeyword, onKeywordChange]);
  const handleKeyDown = useCallback((e: React.KeyboardEvent) => { if (e.key === "Enter") handleSearch(); }, [handleSearch]);

  const activeCount = Object.values(filters).filter(Boolean).length;

  // 安全获取各 facet 数组
  const safeDirections = safeItems(facets?.directions);
  const safeCities = safeItems(facets?.cities);
  const safeJobTypes = safeItems(facets?.job_types);
  const safeCategories = safeItems(facets?.categories);
  const safeBusinessGroups = safeItems(facets?.business_groups);
  const safeSkills = safeItems(facets?.skills);
  const safeAppProjects = safeItems(facets?.application_projects);
  const safeAppSubtypes = safeItems(facets?.application_subtypes);

  const hasAnyFacets = safeDirections.length > 0 || safeCities.length > 0 || safeJobTypes.length > 0 ||
    safeCategories.length > 0 || safeBusinessGroups.length > 0 || safeSkills.length > 0 ||
    safeAppProjects.length > 0 || safeAppSubtypes.length > 0;

  return (
    <aside className="w-[280px] shrink-0">
      <div className="bg-white border border-gray-100 rounded-2xl shadow-sm overflow-hidden">
        {/* ── 搜索框 ── */}
        <div className="p-5 pb-4">
          <div className="relative">
            <Search className="absolute left-4 top-1/2 -translate-y-1/2 h-[18px] w-[18px] text-gray-400" />
            <input
              type="text"
              placeholder="搜索岗位 / 技能关键词"
              value={localKeyword}
              onChange={(e) => setLocalKeyword(e.target.value)}
              onKeyDown={handleKeyDown}
              className="w-full h-12 pl-11 pr-4 text-[15px] bg-gray-50 border border-gray-200 rounded-xl
                placeholder:text-gray-400 text-gray-800
                focus:outline-none focus:ring-2 focus:ring-[#0052D9]/20 focus:border-[#0052D9] focus:bg-white
                hover:border-gray-300 transition-colors"
            />
          </div>
        </div>

        {/* ── 已选筛选标签 ── */}
        {activeCount > 0 && (
          <div className="px-5 pb-4">
            <div className="flex items-center justify-between mb-2.5">
              <span className="text-[13px] text-gray-500">已选 {activeCount} 项</span>
              <button
                className="text-[13px] text-[#0052D9] hover:text-[#0045b8] transition-colors"
                onClick={() => onFiltersChange({})}
              >
                清空筛选
              </button>
            </div>
            <div className="flex flex-wrap gap-1.5">
              {Object.entries(filters).map(([key, val]) =>
                val && typeof val === "string" ? (
                  <span
                    key={key}
                    className="inline-flex items-center gap-1 text-[13px] bg-blue-50 text-[#0052D9] px-2.5 py-1 rounded-lg cursor-pointer hover:bg-blue-100 transition-colors"
                    onClick={() => toggleFilter(key as keyof JobFilters, val)}
                  >
                    {val}
                    <X className="h-3.5 w-3.5" />
                  </span>
                ) : null
              )}
            </div>
          </div>
        )}

        <div className="border-t border-gray-100" />

        {/* ── 筛选分组 ── */}
        <div className="p-5">
          {facetsLoading && (
            <p className="text-[13px] text-gray-400 text-center py-10">正在加载筛选项…</p>
          )}
          {facetsError && (
            <p className="text-[13px] text-gray-400 text-center py-10">{facetsError}</p>
          )}
          {facets && !facetsLoading && (
            <div className="space-y-1">
              {safeAppProjects.length > 0 && (
                <CollapsibleFilterGroup
                  title="应聘项目"
                  items={safeAppProjects}
                  selectedValue={filters.application_project}
                  onSelect={(v) => toggleFilter("application_project", v)}
                />
              )}
              {safeAppSubtypes.length > 0 && (
                <CollapsibleFilterGroup
                  title="应聘类型"
                  items={safeAppSubtypes}
                  selectedValue={filters.application_subtype}
                  onSelect={(v) => toggleFilter("application_subtype", v)}
                />
              )}
              {safeJobTypes.length > 0 && (
                <CollapsibleFilterGroup
                  title="岗位类型"
                  items={safeJobTypes}
                  selectedValue={filters.recruit_type}
                  onSelect={(v) => toggleFilter("recruit_type", v)}
                />
              )}
              {safeDirections.length > 0 && (
                <CollapsibleFilterGroup
                  title="求职方向"
                  items={safeDirections}
                  selectedValue={filters.direction}
                  onSelect={(v) => toggleFilter("direction", v)}
                />
              )}
              {safeCities.length > 0 && (
                <CollapsibleFilterGroup
                  title="工作地点"
                  items={safeCities}
                  selectedValue={filters.city}
                  onSelect={(v) => toggleFilter("city", v)}
                />
              )}
              {safeCategories.length > 0 && (
                <CollapsibleFilterGroup
                  title="岗位类别"
                  items={safeCategories}
                  selectedValue={filters.category}
                  onSelect={(v) => toggleFilter("category", v)}
                />
              )}
              {safeBusinessGroups.length > 0 && (
                <CollapsibleFilterGroup
                  title="事业群"
                  items={safeBusinessGroups}
                  selectedValue={filters.business_group}
                  onSelect={(v) => toggleFilter("business_group", v)}
                />
              )}
              {safeSkills.length > 0 && (
                <CollapsibleFilterGroup
                  title="技能关键词"
                  items={safeSkills.slice(0, 20)}
                  selectedValue={filters.skill}
                  onSelect={(v) => toggleFilter("skill", v)}
                />
              )}
              {!hasAnyFacets && (
                <p className="text-[13px] text-gray-400 text-center py-6">暂无筛选项</p>
              )}
            </div>
          )}
        </div>
      </div>
    </aside>
  );
}

/** 可展开/收起的筛选分组 */
function CollapsibleFilterGroup({
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
  const [expanded, setExpanded] = useState(true);
  const safeList = safeItems(items);

  const selectedCount = safeList.filter((it) => it.value === selectedValue).length;

  return (
    <div className="border-b border-gray-50 last:border-b-0">
      <button
        onClick={() => setExpanded((v) => !v)}
        className="w-full flex items-center justify-between py-3 text-left group"
      >
        <span className="text-[13px] font-semibold text-gray-700">{title}</span>
        <span className="flex items-center gap-1">
          {!expanded && selectedCount > 0 && (
            <span className="text-[11px] text-[#0052D9] bg-blue-50 px-1.5 py-0.5 rounded-full font-medium">
              {selectedValue}
            </span>
          )}
          {expanded ? (
            <ChevronUp className="h-4 w-4 text-gray-400 group-hover:text-gray-600 transition-colors" />
          ) : (
            <ChevronDown className="h-4 w-4 text-gray-400 group-hover:text-gray-600 transition-colors" />
          )}
        </span>
      </button>

      {expanded && (
        <div className="pb-3 space-y-0.5">
          {safeList.map((item) => {
            const isSelected = selectedValue === item.value;
            return (
              <button
                key={item.value}
                onClick={() => onSelect(item.value)}
                className={`w-full flex items-center justify-between px-3 py-2 rounded-lg text-[14px] transition-colors text-left
                  ${isSelected
                    ? "bg-blue-50 text-[#0052D9] font-medium"
                    : "text-gray-600 hover:bg-gray-50"
                  }`}
              >
                <span className="flex items-center gap-2.5 truncate min-w-0">
                  <span className={`shrink-0 flex h-[18px] w-[18px] items-center justify-center rounded border-2 transition-colors ${
                    isSelected
                      ? "bg-[#0052D9] border-[#0052D9]"
                      : "border-gray-300"
                  }`}>
                    {isSelected && <Check className="h-3 w-3 text-white" strokeWidth={3} />}
                  </span>
                  <span className="truncate">{item.label}</span>
                </span>
                {item.count > 0 && (
                  <span className={`text-[12px] shrink-0 ml-2 ${isSelected ? "text-[#0052D9]" : "text-gray-400"}`}>
                    {item.count}
                  </span>
                )}
              </button>
            );
          })}
        </div>
      )}
    </div>
  );
}
