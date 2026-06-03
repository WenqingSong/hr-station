import React from "react";

export interface DetailSkeletonProps {
  sections?: number;
  showAside?: boolean;
}

export function DetailSkeleton({ sections = 3, showAside = false }: DetailSkeletonProps) {
  return (
    <div className={`flex gap-6 ${showAside ? "flex-col xl:flex-row" : "flex-col"}`}>
      <div className="flex-1 rounded-2xl bg-white ring-1 ring-gray-100 p-8 md:p-10 animate-pulse">
        {Array.from({ length: sections }).map((_, si) => (
          <div key={si} className={si > 0 ? "mt-8 pt-2" : ""}>
            <div className="mb-4 h-5 w-32 rounded bg-gray-100" />
            <div className="space-y-2.5">
              <div className="h-4 w-full rounded bg-gray-100" />
              <div className="h-4 w-5/6 rounded bg-gray-100" />
              <div className="h-4 w-3/4 rounded bg-gray-100" />
            </div>
          </div>
        ))}
      </div>
      {showAside && (
        <div className="hidden w-full rounded-2xl bg-white ring-1 ring-gray-100 p-6 animate-pulse xl:block xl:w-[320px]">
          <div className="mb-3 h-5 w-24 rounded bg-gray-100" />
          <div className="mb-4 space-y-2">
            <div className="h-4 w-full rounded bg-gray-100" />
            <div className="h-4 w-2/3 rounded bg-gray-100" />
          </div>
          <div className="h-40 rounded-lg bg-gray-100" />
        </div>
      )}
    </div>
  );
}
