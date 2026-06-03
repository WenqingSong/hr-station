import React from "react";

export function JobCardSkeleton() {
  return (
    <div className="bg-white rounded-[18px] py-4 px-5 sm:px-6 animate-pulse">
      {/* Title */}
      <div className="h-[22px] bg-gray-200 rounded w-3/4 mb-2" />

      {/* Meta info row: 技术 | 应届实习 | CDG CSIG IEG */}
      <div className="flex gap-1.5 mb-2">
        <div className="h-4 bg-gray-200 rounded w-10" />
        <div className="h-4 bg-gray-200 rounded w-2" />
        <div className="h-4 bg-gray-200 rounded w-12" />
        <div className="h-4 bg-gray-200 rounded w-2" />
        <div className="h-4 bg-gray-200 rounded w-2/3" />
      </div>

      {/* Location */}
      <div className="h-4 bg-gray-200 rounded w-1/2" />
    </div>
  );
}

export function JobListSkeleton() {
  return (
    <div className="space-y-3">
      {Array.from({ length: 5 }).map((_, i) => (
        <JobCardSkeleton key={i} />
      ))}
    </div>
  );
}
