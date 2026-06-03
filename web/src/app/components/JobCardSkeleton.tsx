import React from "react";

export function JobCardSkeleton() {
  return (
    <div className="bg-white border border-gray-100 rounded-2xl p-6 animate-pulse">
      {/* Title */}
      <div className="h-[22px] bg-gray-200 rounded w-3/4 mb-3" />

      {/* Meta info row: 技术 | 软件开发 | 校园招聘 */}
      <div className="flex gap-1.5 mb-2.5">
        <div className="h-4 bg-gray-200 rounded w-10" />
        <div className="h-4 bg-gray-200 rounded w-2" />
        <div className="h-4 bg-gray-200 rounded w-14" />
        <div className="h-4 bg-gray-200 rounded w-2" />
        <div className="h-4 bg-gray-200 rounded w-16" />
      </div>

      {/* Organization codes: CDG CSIG IEG */}
      <div className="h-4 bg-gray-200 rounded w-2/3 mb-2.5" />

      {/* Location */}
      <div className="h-4 bg-gray-200 rounded w-1/2 mb-3" />

      {/* Description lines */}
      <div className="space-y-2 mb-4">
        <div className="h-4 bg-gray-200 rounded w-full" />
        <div className="h-4 bg-gray-200 rounded w-4/5" />
      </div>

      {/* Bottom bar: 查看详情 */}
      <div className="pt-3.5 border-t border-gray-50 flex justify-end">
        <div className="h-4 bg-gray-200 rounded w-20" />
      </div>
    </div>
  );
}

export function JobListSkeleton() {
  return (
    <div className="space-y-4">
      {Array.from({ length: 5 }).map((_, i) => (
        <JobCardSkeleton key={i} />
      ))}
    </div>
  );
}
