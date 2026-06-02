import React, { useState, useEffect, useMemo } from "react";
import { JobCard } from "./JobCard";
import { getJobs, type JobListParams } from "../../api/jobs";
import { toJobCardViewModelsFromApi, type JobCardViewModel } from "../../adapters";
import { X } from "lucide-react";

interface JobListProps {
  params: JobListParams;
  onPageChange: (page: number) => void;
  onParamsChange: (p: JobListParams) => void;
  /** 用于 display 的已选筛选条件（支持中文 label） */
  activeFilters: { key: string; label: string }[];
  onClearFilters: () => void;
}

const PAGE_SIZE = 10;

export function JobList({ params, onPageChange, onParamsChange, activeFilters, onClearFilters }: JobListProps) {
  const [jobCards, setJobCards] = useState<JobCardViewModel[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const page = params.page || 1;
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));

  useEffect(() => {
    setLoading(true);
    setError(null);
    getJobs(params)
      .then((res) => { setJobCards(toJobCardViewModelsFromApi(res.items)); setTotal(res.total); })
      .catch((e) => setError(e instanceof Error ? e.message : "加载失败"))
      .finally(() => setLoading(false));
  }, [params.keyword, params.direction, params.city, params.recruit_type, params.category, params.business_group, params.skill, params.application_project, params.application_subtype, params.page]);

  return (
    <main className="flex-1 max-w-[820px] pb-24">
      {/* ── 列表区顶部 ── */}
      <div className="mb-5">
        <h1 className="text-2xl font-bold text-gray-900 mb-1">找到更适合你的实习岗位</h1>

        {/* ── 筛选状态栏（参考腾讯招聘）── */}
        <div className="flex items-center justify-between flex-wrap gap-2 mt-3">
          <span className="text-[14px] text-gray-500">
            共 <span className="text-gray-800 font-semibold">{total}</span> 个岗位
          </span>
          {activeFilters.length > 0 && (
            <div className="flex items-center gap-2 flex-wrap">
              <div className="flex flex-wrap gap-1.5">
                {activeFilters.map((f) => (
                  <span
                    key={f.key}
                    className="inline-flex items-center gap-1 text-[13px] bg-blue-50 text-[#0052D9] px-2.5 py-1 rounded-md cursor-pointer hover:bg-blue-100 transition-colors"
                    onClick={() => {
                      const cleared: Record<string, undefined> = {};
                      // Find which param key this filter maps to
                      const keyMap: Record<string, string> = {
                        direction: "direction", city: "city", recruit_type: "recruit_type",
                        category: "category", business_group: "business_group", skill: "skill",
                        application_project: "application_project", application_subtype: "application_subtype",
                        keyword: "keyword",
                      };
                      const paramKey = f.key;
                      onParamsChange({ [paramKey]: undefined } as unknown as JobListParams);
                    }}
                  >
                    {f.label}
                    <X className="h-3.5 w-3.5" />
                  </span>
                ))}
              </div>
              <button
                className="text-[13px] text-gray-400 hover:text-[#0052D9] transition-colors shrink-0"
                onClick={onClearFilters}
              >
                清除全部
              </button>
            </div>
          )}
        </div>
      </div>

      {/* ── 岗位卡片列表 ── */}
      <div className="space-y-4">
        {loading && (
          <p className="text-sm text-gray-400 py-12 text-center">正在加载岗位列表…</p>
        )}
        {error && (
          <div className="text-center py-12">
            <p className="text-sm text-gray-500 mb-3">岗位列表加载失败，请稍后重试</p>
            <button className="text-sm text-[#0052D9] hover:underline" onClick={() => onParamsChange({})}>清空筛选并重试</button>
          </div>
        )}
        {!loading && !error && jobCards.length === 0 && (
          <div className="text-center py-12">
            <p className="text-sm text-gray-500 mb-2">没有找到匹配岗位</p>
            <p className="text-xs text-gray-400 mb-3">可尝试放宽筛选条件或搜索其他关键词</p>
            <button className="text-sm text-[#0052D9] hover:underline" onClick={onClearFilters}>清空筛选条件</button>
          </div>
        )}
        {jobCards.map((job) => (
          <JobCard
            key={job.id}
            id={job.id}
            title={job.title}
            direction={job.direction}
            type={job.typeLabel}
            department={job.businessGroup}
            businessGroups={job.businessGroups}
            skills={job.tags}
            location={job.locationsText}
            locations={job.locationsList}
            description={job.shortDescription}
            statusLabel={job.statusLabel}
          />
        ))}
      </div>

      {/* ── 分页 ── */}
      {!loading && !error && totalPages > 1 && (
        <div className="flex items-center justify-center gap-1 mt-8">
          <button
            disabled={page <= 1}
            onClick={() => onPageChange(page - 1)}
            className="px-3 py-1.5 text-sm border border-gray-200 rounded-md disabled:opacity-30 disabled:cursor-not-allowed hover:border-[#0052D9] hover:text-[#0052D9] transition-colors"
          >
            上一页
          </button>
          {Array.from({ length: Math.min(totalPages, 7) }, (_, i) => {
            let pg: number;
            if (totalPages <= 7) { pg = i + 1; }
            else if (page <= 4) { pg = i + 1; }
            else if (page >= totalPages - 3) { pg = totalPages - 6 + i; }
            else { pg = page - 3 + i; }
            return (
              <button
                key={pg}
                disabled={pg === page}
                onClick={() => onPageChange(pg)}
                className={`min-w-[36px] px-2 py-1.5 text-sm border rounded-md transition-colors ${
                  pg === page
                    ? "bg-[#0052D9] text-white border-[#0052D9]"
                    : "border-gray-200 hover:border-[#0052D9] hover:text-[#0052D9]"
                }`}
              >
                {pg}
              </button>
            );
          })}
          <button
            disabled={page >= totalPages}
            onClick={() => onPageChange(page + 1)}
            className="px-3 py-1.5 text-sm border border-gray-200 rounded-md disabled:opacity-30 disabled:cursor-not-allowed hover:border-[#0052D9] hover:text-[#0052D9] transition-colors"
          >
            下一页
          </button>
          <span className="ml-3 text-xs text-gray-400">
            第 {page}/{totalPages} 页
          </span>
        </div>
      )}
    </main>
  );
}
