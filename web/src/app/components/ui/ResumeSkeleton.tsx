import { Skeleton } from "./skeleton";

/**
 * K8-5-14-H：简历编辑页骨架加载
 *
 * 用于 MyResumesPage 初次加载时展示，模拟表单卡片结构。
 */

function CardSkeleton({ rows = 4 }: { rows?: number }) {
  return (
    <div className="rounded-2xl bg-white ring-1 ring-gray-100 p-6 space-y-4">
      <Skeleton className="h-5 w-24" />
      <Skeleton className="h-4 w-64" />
      <div className="grid grid-cols-2 gap-4">
        {Array.from({ length: rows }).map((_, i) => (
          <Skeleton key={i} className="h-10 rounded-xl" />
        ))}
      </div>
    </div>
  );
}

export function ResumeSkeleton() {
  return (
    <div className="min-h-screen bg-[#F5F7FB]">
      {/* PageHeader skeleton */}
      <div className="mx-auto max-w-[1280px] px-4 pt-8 pb-4 sm:px-6 lg:px-8">
        <Skeleton className="h-4 w-24 mb-3" />
        <Skeleton className="h-7 w-40 mb-2" />
        <Skeleton className="h-4 w-80" />
      </div>

      {/* Main layout skeleton */}
      <div className="mx-auto max-w-[1280px] px-4 pb-24 sm:px-6 lg:px-8">
        <div className="grid grid-cols-1 gap-6 xl:grid-cols-[minmax(0,1fr)_320px] xl:items-start">
          <div className="space-y-5">
            <CardSkeleton rows={4} />
            <CardSkeleton rows={6} />
            <CardSkeleton rows={3} />
            <CardSkeleton rows={2} />
          </div>
          <aside className="hidden xl:block xl:sticky xl:top-24">
            <div className="rounded-2xl bg-white ring-1 ring-gray-100 p-5 space-y-4">
              <Skeleton className="h-5 w-20" />
              <Skeleton className="h-2 w-full" />
              <Skeleton className="h-4 w-36" />
              <Skeleton className="h-4 w-28" />
              <Skeleton className="h-20 w-full rounded-xl" />
            </div>
          </aside>
        </div>
      </div>
    </div>
  );
}
