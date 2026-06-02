import React, { useState, useEffect } from "react";
import { Badge } from "./ui/Badge";
import { JobCard } from "./JobCard";
import { getJobs, type JobListParams } from "../../api/jobs";
import { toJobCardViewModelsFromApi, type JobCardViewModel } from "../../adapters";

interface JobListProps {
  params: JobListParams;
  onParamsChange: (p: JobListParams) => void;
}

export function JobList({ params, onParamsChange }: JobListProps) {
  const [jobCards, setJobCards] = useState<JobCardViewModel[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setLoading(true);
    setError(null);
    getJobs(params)
      .then((res) => setJobCards(toJobCardViewModelsFromApi(res.items)))
      .catch((e) => setError(e instanceof Error ? e.message : "加载失败"))
      .finally(() => setLoading(false));
  }, [params.direction, params.city, params.recruit_type, params.category, params.keyword]);

  const activeFilters = (Object.keys(params) as (keyof JobListParams)[])
    .filter((k) => params[k] && k !== "limit" && k !== "offset")
    .map((k) => ({ key: k, value: params[k] as string }));

  const clearAll = () => onParamsChange({});

  return (
    <main className="flex-1 max-w-[820px] pb-24">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900 mb-2">找到更适合你的实习岗位</h1>
        <p className="text-gray-500">浏览当前开放岗位，按方向、城市和岗位类别筛选适合你的机会。</p>
      </div>

      {activeFilters.length > 0 && (
        <div className="flex items-center gap-2 mb-6">
          <span className="text-sm text-gray-500 mr-2">当前筛选：</span>
          {activeFilters.map((f) => (
            <Badge key={f.key} variant="blue" className="px-3 py-1 bg-blue-50/50 cursor-pointer" onClick={() => onParamsChange({ ...params, [f.key]: undefined })}>
              {f.value} ✕
            </Badge>
          ))}
          <button className="text-sm text-gray-400 hover:text-gray-600 ml-2" onClick={clearAll}>清空全部</button>
        </div>
      )}

      <div className="space-y-4">
        {loading && (
          <p className="text-sm text-gray-400 py-12 text-center">正在加载岗位列表…</p>
        )}
        {error && (
          <div className="text-center py-12">
            <p className="text-sm text-gray-500 mb-3">岗位数据加载失败，请稍后重试。</p>
            <button className="text-sm text-[#0052D9] hover:underline" onClick={clearAll}>清空筛选条件并重试</button>
          </div>
        )}
        {!loading && !error && jobCards.length === 0 && (
          <p className="text-sm text-gray-400 py-12 text-center">暂无符合筛选条件的岗位</p>
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
    </main>
  );
}
