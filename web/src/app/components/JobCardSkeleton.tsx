import React from "react";

export function JobCardSkeleton() {
  return (
    <div className="bg-white border border-gray-100 rounded-2xl p-6 animate-pulse">
      {/* Title */}
      <div className="h-5 bg-gray-200 rounded w-3/4 mb-4" />

      {/* Meta info row: direction | type | department */}
      <div className="flex gap-2 mb-4">
        <div className="h-4 bg-gray-200 rounded w-16" />
        <div className="h-4 bg-gray-200 rounded w-12" />
        <div className="h-4 bg-gray-200 rounded w-28" />
      </div>

      {/* Business group badges */}
      <div className="flex gap-1.5 mb-4">
        <div className="h-5 bg-gray-200 rounded w-14" />
        <div className="h-5 bg-gray-200 rounded w-20" />
        <div className="h-5 bg-gray-200 rounded w-16" />
      </div>

      {/* Location */}
      <div className="flex items-center gap-1.5 mb-4">
        <div className="h-3 w-3 bg-gray-200 rounded" />
        <div className="h-4 bg-gray-200 rounded w-36" />
      </div>

      {/* Description lines */}
      <div className="space-y-2 mb-4">
        <div className="h-4 bg-gray-200 rounded w-full" />
        <div className="h-4 bg-gray-200 rounded w-5/6" />
      </div>

      {/* Skills + action button */}
      <div className="flex items-center justify-between">
        <div className="flex gap-1.5">
          <div className="h-6 bg-gray-200 rounded w-12" />
          <div className="h-6 bg-gray-200 rounded w-16" />
          <div className="h-6 bg-gray-200 rounded w-10" />
        </div>
        <div className="h-8 bg-gray-200 rounded w-20" />
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
