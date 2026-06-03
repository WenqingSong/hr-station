import React, { useState, useEffect, useCallback } from "react";
import { Search, Check, ChevronUp, ChevronDown } from "lucide-react";
import { getJobFacets, type JobFacets, type JobFacetGroup, type JobFacetOption } from "../../api/jobs";

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

  // multi-select toggle (comma-separated)
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
    <aside className="w-[280px] shrink-0">
      <div className="bg-white border border-gray-100 rounded-2xl shadow-sm overflow-hidden">
        {/* Search */}
        <div className="p-5 pb-4">
          <div className="relative">
            <Search className="absolute left-4 top-1/2 -translate-y-1/2 h-[18px] w-[18px] text-gray-400" />
            <input
              type="text" placeholder="搜索岗位 / 技能关键词"
              value={localKeyword} onChange={(e) => setLocalKeyword(e.target.value)}
              onKeyDown={handleKeyDown}
              className="w-full h-12 pl-11 pr-4 text-[15px] bg-gray-50 border border-gray-200 rounded-xl placeholder:text-gray-400 text-gray-800 focus:outline-none focus:ring-2 focus:ring-[#0052D9]/20 focus:border-[#0052D9] focus:bg-white hover:border-gray-300 transition-colors"
            />
          </div>
        </div>

        {/* Filter groups */}
        <div className="p-5">
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
  // K8-4-4-R2: 父级 option children 展开状态（二级默认关闭）
  const [expandedParentCodes, setExpandedParentCodes] = useState<Set<string>>(new Set());
  const multiSelectGroups = new Set(['city', 'business_group', 'skill']);

  const isMulti = multiSelectGroups.has(group.code);
  const isJobCategory = group.code === 'job_category';
  const filterVal = filters[group.code as keyof JobFilters] || '';
  const selectedVals = new Set(filterVal ? filterVal.split(',') : []);
  // K8-4-4-R2: 本组内选中数量（用于折叠徽标）
  const groupSelectedCount = selectedVals.size;

  const visibleOpts = showAll ? group.options : group.options.slice(0, group.max_visible_options);
  const hasMore = group.options.length > group.max_visible_options;

  const handleSelect = (value: string) => {
    (isMulti ? toggleMulti : toggle)(group.code as keyof JobFilters, value);
  };

  return (
    <div className="border-b border-gray-50 last:border-b-0">
      <button onClick={() => setExpanded((v) => !v)} className="w-full flex items-center justify-between py-3 text-left group">
        <span className="text-[13px] font-semibold text-gray-700">{group.name}</span>
        <span className="flex items-center gap-1">
          {!expanded && groupSelectedCount > 0 && (
            <span className="text-[11px] text-[#0052D9] bg-blue-50 px-1.5 py-0.5 rounded-full font-medium">
              {groupSelectedCount}
            </span>
          )}
          {expanded ? <ChevronUp className="h-4 w-4 text-gray-400 group-hover:text-gray-600" /> : <ChevronDown className="h-4 w-4 text-gray-400 group-hover:text-gray-600" />}
        </span>
      </button>
      {expanded && (
        <div className="pb-3 space-y-0.5">
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
            <button onClick={() => setShowAll((v) => !v)} className="text-[13px] text-[#0052D9] hover:underline px-3 py-1">
              {showAll ? '收起' : `展开更多 (+${group.options.length - group.max_visible_options} 项)`}
            </button>
          )}
        </div>
      )}
    </div>
  );
}

// ─── K8-4-4-R2: 岗位类别父级行（checkbox + 展开按钮分离）─────────────────────

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
        {/* checkbox + label + count：点击筛选父级 */}
        <button
          onClick={() => onSelect(parent.value)}
          className={`flex-1 flex items-center justify-between px-3 py-2 rounded-l-lg text-[14px] transition-colors text-left
            ${isSelected ? "bg-blue-50 text-[#0052D9] font-medium" : "text-gray-600 hover:bg-gray-50"}`}
        >
          <span className="flex items-center gap-2.5 truncate min-w-0">
            <span className={`shrink-0 flex h-[18px] w-[18px] items-center justify-center rounded border-2 transition-colors ${
              isSelected ? "bg-[#0052D9] border-[#0052D9]" : "border-gray-300"}`}>
              {isSelected && <Check className="h-3 w-3 text-white" strokeWidth={3} />}
            </span>
            <span className="truncate">{parent.label}</span>
          </span>
          {parent.count > 0 && (
            <span className={`text-[12px] shrink-0 ml-2 ${isSelected ? "text-[#0052D9]" : "text-gray-400"}`}>{parent.count}</span>
          )}
        </button>
        {/* 展开/收起 children：不触发筛选 */}
        {hasChildren && (
          <button
            onClick={(e) => { e.stopPropagation(); onToggleChildren(); }}
            className="shrink-0 flex items-center justify-center w-8 h-full min-h-[36px] rounded-r-lg text-gray-400 hover:text-gray-600 hover:bg-gray-50 transition-colors"
            title={isChildExpanded ? '收起子项' : '展开子项'}
          >
            {isChildExpanded
              ? <ChevronUp className="h-4 w-4" />
              : <ChevronDown className="h-4 w-4" />}
          </button>
        )}
      </div>
      {/* 子级列表：仅在展开时渲染 */}
      {isChildExpanded && hasChildren && (
        <div className="space-y-0.5">
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
      )}
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
      className={`w-full flex items-center justify-between px-3 py-2 rounded-lg text-[14px] transition-colors text-left
        ${isSelected ? "bg-blue-50 text-[#0052D9] font-medium" : "text-gray-600 hover:bg-gray-50"}`}
      style={{ paddingLeft: `${12 + depth * 16}px` }}
    >
      <span className="flex items-center gap-2.5 truncate min-w-0">
        <span className={`shrink-0 flex h-[18px] w-[18px] items-center justify-center rounded border-2 transition-colors ${
          isSelected ? "bg-[#0052D9] border-[#0052D9]" : "border-gray-300"}`}>
          {isSelected && <Check className="h-3 w-3 text-white" strokeWidth={3} />}
        </span>
        <span className="truncate">{option.label}</span>
      </span>
      {option.count > 0 && (
        <span className={`text-[12px] shrink-0 ml-2 ${isSelected ? "text-[#0052D9]" : "text-gray-400"}`}>{option.count}</span>
      )}
    </button>
  );
}

// ─── Old collapsible group (legacy fallback) ──────────────────────────────────

function CollapsibleFilterGroup({ title, items, selectedValue, onSelect }: {
  title: string; items: { value: string; label: string; count: number }[]; selectedValue?: string; onSelect: (value: string) => void;
}) {
  const [expanded, setExpanded] = useState(true);
  const list = safeArr(items);
  const sel = list.filter((it) => it.value === selectedValue).length;
  return (
    <div className="border-b border-gray-50 last:border-b-0">
      <button onClick={() => setExpanded((v) => !v)} className="w-full flex items-center justify-between py-3 text-left group">
        <span className="text-[13px] font-semibold text-gray-700">{title}</span>
        <span className="flex items-center gap-1">
          {!expanded && sel > 0 && <span className="text-[11px] text-[#0052D9] bg-blue-50 px-1.5 py-0.5 rounded-full font-medium">{selectedValue}</span>}
          {expanded ? <ChevronUp className="h-4 w-4 text-gray-400" /> : <ChevronDown className="h-4 w-4 text-gray-400" />}
        </span>
      </button>
      {expanded && (
        <div className="pb-3 space-y-0.5">
          {list.map((item) => {
            const s = selectedValue === item.value;
            return (
              <button key={item.value} onClick={() => onSelect(item.value)}
                className={`w-full flex items-center justify-between px-3 py-2 rounded-lg text-[14px] transition-colors text-left ${s ? "bg-blue-50 text-[#0052D9] font-medium" : "text-gray-600 hover:bg-gray-50"}`}>
                <span className="flex items-center gap-2.5 truncate min-w-0">
                  <span className={`shrink-0 flex h-[18px] w-[18px] items-center justify-center rounded border-2 ${s ? "bg-[#0052D9] border-[#0052D9]" : "border-gray-300"}`}>
                    {s && <Check className="h-3 w-3 text-white" strokeWidth={3} />}
                  </span>
                  <span className="truncate">{item.label}</span>
                </span>
                {item.count > 0 && <span className={`text-[12px] shrink-0 ml-2 ${s ? "text-[#0052D9]" : "text-gray-400"}`}>{item.count}</span>}
              </button>
            );
          })}
        </div>
      )}
    </div>
  );
}
