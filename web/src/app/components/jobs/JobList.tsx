import React, { useState, useEffect, useCallback } from "react";
import { JobCard } from "./JobCard";
import { JobListSkeleton } from "./JobCardSkeleton";
import { JobPagination } from "./JobPagination";
import { getJobs, type JobListParams } from "../../../api/jobs";
import { toJobCardViewModelsFromApi, type JobCardViewModel } from "../../../adapters";
import { PageError, PageEmpty } from "../PageState";
import { X, RefreshCw } from "lucide-react";

interface JobListProps {
  params: JobListParams;
  onPageChange: (page: number) => void;
  onParamsChange: (p: JobListParams) => void;
  /** 用于 display 的已选筛选条件（含分组中文前缀） */
  activeFilters: { key: string; label: string; group: string }[];
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

  // K8-4-4-R4: 状态拆分
  const isInitialLoading = loading && jobCards.length === 0;
  const isRefetching = loading && jobCards.length > 0;
  const isEmpty = !loading && !error && jobCards.length === 0;
  const isError = Boolean(error);

  // K8-5-12-D2: retry by re-triggering request with current params
  const handleRetry = useCallback(() => onParamsChange({}), [onParamsChange]);

  useEffect(() => {
    setLoading(true);
    setError(null);
    getJobs(params)
      .then((res) => { setJobCards(toJobCardViewModelsFromApi(res.items)); setTotal(res.total); })
      .catch((e) => setError(e instanceof Error ? e.message : "加载失败"))
      .finally(() => setLoading(false));
  }, [params.keyword, params.direction, params.city, params.recruit_type, params.category, params.business_group, params.skill, params.job_category, params.application_project, params.application_subtype, params.page]);

  return (
    <main className="min-w-0 w-full pb-24">
      {/* ── 第 2 层：岗位列表大容器 ── */}
      <div className="bg-[#F0F2F5] rounded-[24px] p-5 sm:p-6">

        {/* ── 容器顶部：共 X 个岗位 + 已选标签 ── */}
        <div className="flex items-center justify-between flex-wrap gap-2 mb-5">
          <div className="flex items-center gap-2">
            <span className="text-[14px] text-gray-500">
              共 <span className="text-gray-800 font-semibold">{isInitialLoading ? "..." : total}</span> 个岗位
            </span>
            {isRefetching && (
              <span className="inline-flex items-center gap-1 text-[13px] text-[#0052D9]">
                <RefreshCw className="h-3.5 w-3.5 animate-spin" />
                <span className="hidden sm:inline">更新中</span>
              </span>
            )}
          </div>
          {activeFilters.length > 0 && (
            <div className="flex items-center gap-2 flex-wrap">
              <div className="flex flex-wrap gap-1.5">
                {activeFilters.map((f) => (
                  <span
                    key={f.key + f.label}
                    className="inline-flex items-center gap-1 text-[13px] bg-white border border-gray-200 text-gray-700 px-2.5 py-1 rounded-lg cursor-pointer hover:bg-gray-50 hover:border-gray-300 transition-colors"
                    onClick={() => {
                      onParamsChange({ [f.key]: undefined } as unknown as JobListParams);
                    }}
                  >
                    <span className="text-gray-500">{f.group}：</span>
                    <span>{f.label}</span>
                    <X className="h-3.5 w-3.5 text-gray-400 hover:text-gray-600" />
                  </span>
                ))}
              </div>
              <button
                className="text-[13px] text-gray-400 hover:text-gray-600 transition-colors shrink-0"
                onClick={onClearFilters}
              >
                清除全部
              </button>
            </div>
          )}
        </div>

        {/* ── 岗位卡片列表 ── */}
        <div className="space-y-3">
          {/* Initial loading：骨架屏 */}
          {isInitialLoading && <JobListSkeleton />}

          {/* Error state — K8-5-12-D2: 统一 PageError */}
          {isError && (
            <PageError
              title="岗位列表加载失败"
              description={error || "暂时无法获取岗位数据，请稍后重试。"}
              onRetry={handleRetry}
            />
          )}

          {/* Empty state — K8-5-12-D2: 统一 PageEmpty */}
          {isEmpty && (
            <PageEmpty
              title="未找到匹配岗位"
              description="可以尝试调整关键词、城市、岗位类别或其他筛选条件。"
              actionLabel="清除筛选"
              onAction={onClearFilters}
            />
          )}

          {/* Job cards (hidden during initial loading, dimmed during refetch) */}
          {!isInitialLoading && !isError && jobCards.map((job) => (
            <div key={job.id} className={isRefetching ? "opacity-50 transition-opacity" : ""}>
              <JobCard
                id={job.id}
                title={job.title}
                jobTypeLabel={job.jobTypeLabel}
                applicationType={job.applicationType}
                organizationCodes={job.organizationCodes || []}
                locations={job.locationsList}
              />
            </div>
          ))}
        </div>

        {/* ── 分页 ── */}
        {!isInitialLoading && !isError && totalPages > 1 && (
          <JobPagination
            currentPage={page}
            totalPages={totalPages}
            onPageChange={onPageChange}
          />
        )}
      </div>
    </main>
  );
}
