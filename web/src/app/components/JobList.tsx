import React, { useState, useEffect } from "react";
import { JobCard } from "./JobCard";
import { getJobs, type JobListParams } from "../../api/jobs";
import { toJobCardViewModelsFromApi, type JobCardViewModel } from "../../adapters";

interface JobListProps {
  params: JobListParams;
  onPageChange: (page: number) => void;
  onParamsChange: (p: JobListParams) => void;
}

const PAGE_SIZE = 10;

export function JobList({ params, onPageChange, onParamsChange }: JobListProps) {
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
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900 mb-2">找到更适合你的实习岗位</h1>
        <p className="text-gray-500">浏览当前开放岗位，按方向、城市和岗位类别筛选适合你的机会。</p>
      </div>

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
            <button className="text-sm text-[#0052D9] hover:underline" onClick={() => onParamsChange({})}>清空筛选条件</button>
          </div>
        )}
        {jobCards.map((job) => (
          <JobCard
            key={job.id}
            id={job.id}
            title={job.title}
            type={job.typeLabel}
            department={job.businessGroup}
            skills={job.tags}
            location={job.locationsText}
            description={job.shortDescription}
            statusLabel={job.statusLabel}
          />
        ))}
      </div>

      {/* Pagination */}
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
            共 {total} 个岗位 · 第 {page}/{totalPages} 页
          </span>
        </div>
      )}
    </main>
  );
}
