import React, { useState, useEffect, useCallback } from "react";
import { Search, Check, ChevronRight } from "lucide-react";
import { getJobFacets, type JobFacets, type JobFacetGroup, type JobFacetOption } from "../../../api/jobs";
import { CollapsibleSection } from "../CollapsibleSection";

export interface JobFilters {
  direction?: string; city?: string; recruit_type?: string;
  category?: string; business_group?: string; skill?: string;
  application_project?: string; application_subtype?: string;
  job_category?: string;
}

interface JobFilterSidebarProps {
  keyword: string; onKeywordChange: (v: string) => void;
  filters: JobFilters; onFiltersChange: (f: JobFilters) => void;
}

function safeArr<T>(v: T[] | null | undefined): T[] { return Array.isArray(v) ? v : []; }

export function JobFilterSidebar({ keyword, onKeywordChange, filters, onFiltersChange }: JobFilterSidebarProps) {
  const [facets, setFacets] = useState<JobFacets | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [localKeyword, setLocalKeyword] = useState(keyword);

  useEffect(() => { setLocalKeyword(keyword); }, [keyword]);
  useEffect(() => {
    getJobFacets().then(setFacets).catch(() => setError("加载失败")).finally(() => setLoading(false));
  }, []);

  const toggle = useCallback((key: keyof JobFilters, value: string) => {
    const cur = filters[key];
    onFiltersChange({ ...filters, [key]: cur === value ? undefined : value });
  }, [filters, onFiltersChange]);

  const handleSearch = useCallback(() => onKeywordChange(localKeyword), [localKeyword, onKeywordChange]);
  const handleKeyDown = useCallback((e: React.KeyboardEvent) => { if (e.key === "Enter") handleSearch(); }, [handleSearch]);

  const toggleMulti = useCallback((key: keyof JobFilters, value: string) => {
    const cur = filters[key] || '';
    const arr = cur ? cur.split(',') : [];
    const idx = arr.indexOf(value);
    if (idx >= 0) arr.splice(idx, 1);
    else arr.push(value);
    onFiltersChange({ ...filters, [key]: arr.length > 0 ? arr.join(',') : undefined });
  }, [filters, onFiltersChange]);

  const groups: JobFacetGroup[] = safeArr(facets?.groups);
  const useLegacy = groups.length === 0 && facets !== null;

  return (
    <aside className="w-full xl:w-[260px]">
      <div className="bg-white rounded-2xl overflow-hidden">
        {/* Search */}
        <div className="p-4 pb-2">
          <div className="relative flex items-center bg-gray-100 rounded-lg">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
            <input
              type="text" placeholder="搜索工作岗位"
              value={localKeyword} onChange={(e) => setLocalKeyword(e.target.value)}
              onKeyDown={handleKeyDown}
              className="flex-1 h-9 pl-9 pr-[58px] text-[14px] bg-transparent border-0 placeholder:text-gray-400 text-gray-800 focus:outline-none focus:ring-0"
            />
            <button
              onClick={handleSearch}
              className="absolute right-2 top-1/2 -translate-y-1/2 flex items-center gap-0.5 text-[14px] text-[#0052D9] hover:text-[#0045b8] font-medium"
            >
              查看
              <ChevronRight className="h-4 w-4" />
            </button>
          </div>
        </div>

        {/* Filter groups */}
        <div className="px-4 pb-4">
          {loading && <p className="text-[13px] text-gray-400 text-center py-10">正在加载筛选项…</p>}
          {error && <p className="text-[13px] text-gray-400 text-center py-10">{error}</p>}
          {!loading && !error && useLegacy && <LegacyFilters facets={facets!} filters={filters} toggle={toggle} />}
          {!loading && !error && !useLegacy && groups.map((g) => (
            <FacetGroupBlock key={g.code} group={g} filters={filters} toggle={toggle} toggleMulti={toggleMulti} />
          ))}
          {!loading && !error && groups.length === 0 && !useLegacy && (
            <p className="text-[13px] text-gray-400 text-center py-6">暂无筛选项</p>
          )}
        </div>
      </div>
    </aside>
  );
}

// ─── Legacy fallback ──────────────────────────────────────────────────────────

function LegacyFilters({ facets, filters, toggle }: { facets: JobFacets; filters: JobFilters; toggle: (k: keyof JobFilters, v: string) => void }) {
  const f = facets;
  return <div className="space-y-1">{
    [
      f.application_projects && safeArr(f.application_projects).length>0 && ['应聘项目','application_project',safeArr(f.application_projects),filters.application_project],
      f.application_subtypes && safeArr(f.application_subtypes).length>0 && ['应聘类型','application_subtype',safeArr(f.application_subtypes),filters.application_subtype],
      f.job_types && safeArr(f.job_types).length>0 && ['岗位类型','recruit_type',safeArr(f.job_types),filters.recruit_type],
      f.directions && safeArr(f.directions).length>0 && ['求职方向','direction',safeArr(f.directions),filters.direction],
      f.cities && safeArr(f.cities).length>0 && ['工作地点','city',safeArr(f.cities),filters.city],
      f.categories && safeArr(f.categories).length>0 && ['岗位类别','category',safeArr(f.categories),filters.category],
      f.business_groups && safeArr(f.business_groups).length>0 && ['事业群','business_group',safeArr(f.business_groups),filters.business_group],
      f.skills && safeArr(f.skills).length>0 && ['技能关键词','skill',safeArr(f.skills).slice(0,20),filters.skill],
    ].filter(Boolean).map(([title,key,items,sel]: any) => (
      <CollapsibleFilterGroup key={key} title={title} items={items} selectedValue={sel} onSelect={(v:string)=>toggle(key,v)} />
    ))}</div>;
}

// ─── Facet group block (new groups[]) ─────────────────────────────────────────

function FacetGroupBlock({ group, filters, toggle, toggleMulti }: {
  group: JobFacetGroup; filters: JobFilters; toggle: (k: keyof JobFilters, v: string) => void; toggleMulti: (k: keyof JobFilters, v: string) => void
}) {
  const [expanded, setExpanded] = useState(false);
  const [showAll, setShowAll] = useState(false);
  const [expandedParentCodes, setExpandedParentCodes] = useState<Set<string>>(new Set());
  const multiSelectGroups = new Set(['city', 'business_group', 'skill']);

  const isMulti = multiSelectGroups.has(group.code);
  const isJobCategory = group.code === 'job_category';
  const filterVal = filters[group.code as keyof JobFilters] || '';
  const selectedVals = new Set(filterVal ? filterVal.split(',') : []);
  const groupSelectedCount = selectedVals.size;

  const visibleOpts = showAll ? group.options : group.options.slice(0, group.max_visible_options);
  const hasMore = group.options.length > group.max_visible_options;

  const handleSelect = (value: string) => {
    (isMulti ? toggleMulti : toggle)(group.code as keyof JobFilters, value);
  };

  return (
    <div className="border-b border-gray-100 last:border-b-0">
      <button onClick={() => setExpanded((v) => !v)} className="w-full flex items-center justify-between py-3.5 text-left">
        <span className="text-[15px] font-bold text-gray-900">{group.name}</span>
        <span className="flex items-center gap-1.5">
          {!expanded && groupSelectedCount > 0 && (
            <span className="text-[11px] text-[#0052D9] bg-blue-50 px-1.5 py-0.5 rounded-full font-medium">
              {groupSelectedCount}
            </span>
          )}
          {/* 箭头旋转动画 */}
          <span className={`text-[17px] font-light text-gray-400 select-none w-4 inline-flex items-center justify-center transition-transform duration-200 ${expanded ? 'rotate-0' : ''}`}>
            {expanded ? '−' : '+'}
          </span>
        </span>
      </button>
      {/* 一级分组展开动画 */}
      <CollapsibleSection open={expanded} duration={220}>
        <div className="pb-3 space-y-0">
          {isJobCategory ? (
            visibleOpts.map((parent) => (
              <JobCategoryParentRow
                key={parent.code}
                parent={parent}
                selectedVals={selectedVals}
                onSelect={handleSelect}
                isChildExpanded={expandedParentCodes.has(parent.code)}
                onToggleChildren={() => {
                  setExpandedParentCodes((prev) => {
                    const next = new Set(prev);
                    if (next.has(parent.code)) next.delete(parent.code);
                    else next.add(parent.code);
                    return next;
                  });
                }}
              />
            ))
          ) : (
            visibleOpts.map((o) => (
              <OptionRow key={o.value} option={o} isSelected={selectedVals.has(o.value)} onSelect={handleSelect} />
            ))
          )}
          {hasMore && (
            <button onClick={() => setShowAll((v) => !v)} className="text-[13px] text-[#0052D9] hover:underline px-2 py-1 mt-1">
              {showAll ? '收起' : `展开更多 (+${group.options.length - group.max_visible_options} 项)`}
            </button>
          )}
        </div>
      </CollapsibleSection>
    </div>
  );
}

// ─── 岗位类别父级行 ───────────────────────────────────────────────────────────

function JobCategoryParentRow({ parent, selectedVals, onSelect, isChildExpanded, onToggleChildren }: {
  parent: JobFacetOption; selectedVals: Set<string>; onSelect: (v: string) => void;
  isChildExpanded: boolean; onToggleChildren: () => void;
}) {
  const children = safeArr(parent.children);
  const hasChildren = children.length > 0;
  const isSelected = selectedVals.has(parent.value);

  return (
    <div>
      <div className="flex items-center">
        <button
          onClick={() => onSelect(parent.value)}
          className={`flex-1 flex items-center justify-between py-1.5 text-[14px] transition-colors text-left
            ${isSelected ? "text-[#0052D9] font-medium" : "text-gray-700 hover:bg-gray-50"}`}
        >
          <span className="flex items-center gap-2.5 truncate min-w-0">
            <Checkbox isSelected={isSelected} />
            <span className="truncate">{parent.label}</span>
          </span>
          {parent.count > 0 && (
            <span className={`text-[12px] shrink-0 ml-2 ${isSelected ? "text-[#0052D9]" : "text-gray-400"}`}>{parent.count}</span>
          )}
        </button>
        {hasChildren && (
          <button
            onClick={(e) => { e.stopPropagation(); onToggleChildren(); }}
            className="shrink-0 flex items-center justify-center w-7 h-7 text-gray-400 hover:text-gray-600 transition-colors"
            title={isChildExpanded ? '收起子项' : '展开子项'}
          >
            <span className={`text-[15px] font-light select-none transition-transform duration-200 ${isChildExpanded ? 'rotate-0' : ''}`}>
              {isChildExpanded ? '−' : '+'}
            </span>
          </button>
        )}
      </div>
      {/* 二级子项展开动画 */}
      <CollapsibleSection open={isChildExpanded} duration={180}>
        <div className="space-y-0">
          {children.map((child) => (
            <OptionRow
              key={child.value}
              option={child}
              isSelected={selectedVals.has(child.value)}
              onSelect={onSelect}
              depth={1}
            />
          ))}
        </div>
      </CollapsibleSection>
    </div>
  );
}

// ─── Single option row ────────────────────────────────────────────────────────

function OptionRow({ option, isSelected, onSelect, depth = 0 }: {
  option: JobFacetOption; isSelected: boolean; onSelect: (v: string) => void; depth?: number
}) {
  return (
    <button
      onClick={() => onSelect(option.value)}
      className={`w-full flex items-center justify-between py-1.5 text-[14px] transition-colors text-left
        ${isSelected ? "text-[#0052D9] font-medium" : "text-gray-700 hover:bg-gray-50"}`}
      style={{ paddingLeft: `${8 + depth * 24}px` }}
    >
      <span className="flex items-center gap-2.5 truncate min-w-0">
        <Checkbox isSelected={isSelected} />
        <span className="truncate">{option.label}</span>
      </span>
      {option.count > 0 && (
        <span className={`text-[12px] shrink-0 ml-2 ${isSelected ? "text-[#0052D9]" : "text-gray-400"}`}>{option.count}</span>
      )}
    </button>
  );
}

// ─── Checkbox ─────────────────────────────────────────────────────────────────

function Checkbox({ isSelected }: { isSelected: boolean }) {
  return (
    <span className={`shrink-0 flex h-4 w-4 items-center justify-center rounded border transition-colors ${
      isSelected ? "bg-[#0052D9] border-[#0052D9]" : "border-gray-300 bg-white"}`}>
      {isSelected && <Check className="h-2.5 w-2.5 text-white" strokeWidth={3} />}
    </span>
  );
}

// ─── Old collapsible group (legacy fallback) ──────────────────────────────────

function CollapsibleFilterGroup({ title, items, selectedValue, onSelect }: {
  title: string; items: { value: string; label: string; count: number }[]; selectedValue?: string; onSelect: (value: string) => void;
}) {
  const [expanded, setExpanded] = useState(false);
  const list = safeArr(items);
  const sel = list.filter((it) => it.value === selectedValue).length;

  return (
    <div className="border-b border-gray-100 last:border-b-0">
      <button onClick={() => setExpanded((v) => !v)} className="w-full flex items-center justify-between py-3.5 text-left">
        <span className="text-[15px] font-bold text-gray-900">{title}</span>
        <span className="flex items-center gap-1.5">
          {!expanded && sel > 0 && <span className="text-[11px] text-[#0052D9] bg-blue-50 px-1.5 py-0.5 rounded-full font-medium">{selectedValue}</span>}
          <span className={`text-[17px] font-light text-gray-400 select-none w-4 inline-flex items-center justify-center transition-transform duration-200 ${expanded ? 'rotate-0' : ''}`}>
            {expanded ? '−' : '+'}
          </span>
        </span>
      </button>
      <CollapsibleSection open={expanded} duration={220}>
        <div className="pb-3 space-y-0">
          {list.map((item) => {
            const s = selectedValue === item.value;
            return (
              <button key={item.value} onClick={() => onSelect(item.value)}
                className={`w-full flex items-center justify-between py-1.5 text-[14px] transition-colors text-left ${s ? "text-[#0052D9] font-medium" : "text-gray-700 hover:bg-gray-50"}`}>
                <span className="flex items-center gap-2.5 truncate min-w-0">
                  <Checkbox isSelected={s} />
                  <span className="truncate">{item.label}</span>
                </span>
                {item.count > 0 && <span className={`text-[12px] shrink-0 ml-2 ${s ? "text-[#0052D9]" : "text-gray-400"}`}>{item.count}</span>}
              </button>
            );
          })}
        </div>
      </CollapsibleSection>
    </div>
  );
}
