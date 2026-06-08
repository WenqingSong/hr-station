import React from "react";
import { ArrowLeft, Sparkles, FileText, Target } from "lucide-react";

/**
 * JobDetailPageSkeleton — 岗位详情页统一骨架屏
 *
 * 模拟真实岗位详情页结构：
 * - 顶部蓝色 Hero / Banner 骨架（标题、标签、slogan）
 * - 主体两栏：左侧 JD 骨架 + 右侧 AI 决策卡骨架
 */
export function JobDetailPageSkeleton() {
  return (
    <>
      {/* ── Hero / Banner skeleton ── */}
      <div className="w-full bg-[#0052D9] relative overflow-hidden text-white">
        {/* 背景渐变 */}
        <div className="absolute inset-0 bg-gradient-to-br from-[#0052D9] via-[#0066FF] to-[#0040A8]" />
        {/* 装饰光斑 */}
        <div className="absolute -top-20 -right-20 w-[500px] h-[500px] rounded-full bg-white/5 blur-3xl pointer-events-none" />
        <div className="absolute -bottom-32 -left-32 w-[400px] h-[400px] rounded-full bg-[#00B0F0]/10 blur-3xl pointer-events-none" />

        <div className="mx-auto max-w-[1280px] px-6 pt-20 pb-8 relative z-10 animate-pulse">
          {/* 返回按钮 */}
          <div className="flex items-center gap-1.5 mb-5">
            <ArrowLeft className="w-4 h-4 text-blue-100/60" />
            <div className="h-4 w-24 bg-white/10 rounded" />
          </div>

          {/* 岗位标题 */}
          <div className="h-8 w-3/4 bg-white/15 rounded mb-3" />

          {/* 标签 */}
          <div className="flex items-center gap-2 mb-3">
            <div className="h-6 w-12 bg-white/10 rounded-sm" />
            <div className="h-6 w-16 bg-white/10 rounded-sm" />
          </div>

          {/* Slogan */}
          <div className="h-4 w-2/3 bg-white/10 rounded" />
        </div>

        {/* 底部微光边缘 */}
        <div className="absolute bottom-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-white/20 to-transparent" />
      </div>

      {/* ── Body: Two-column skeleton ── */}
      <div className="mx-auto max-w-[1280px] px-4 sm:px-6 lg:px-8 pt-8 md:pt-10">
        <div className="grid grid-cols-1 gap-6 xl:grid-cols-[minmax(0,1fr)_320px] xl:gap-8 items-start">
          {/* Left: JD skeleton */}
          <div className="flex-1 space-y-6 animate-pulse">
            {/* Card 1: Job description */}
            <div className="bg-white rounded-2xl ring-1 ring-gray-100 p-8 md:p-10">
              {/* Section 1 */}
              <div>
                <div className="h-5 w-24 bg-gray-100 rounded mb-4" />
                <div className="space-y-2.5">
                  <div className="h-4 w-full bg-gray-100 rounded" />
                  <div className="h-4 w-5/6 bg-gray-100 rounded" />
                  <div className="h-4 w-3/4 bg-gray-100 rounded" />
                </div>
              </div>
              {/* Section 2 */}
              <div className="mt-10">
                <div className="h-5 w-24 bg-gray-100 rounded mb-4" />
                <div className="space-y-2.5">
                  <div className="h-4 w-full bg-gray-100 rounded" />
                  <div className="h-4 w-4/5 bg-gray-100 rounded" />
                  <div className="h-4 w-5/6 bg-gray-100 rounded" />
                  <div className="h-4 w-3/5 bg-gray-100 rounded" />
                </div>
              </div>
            </div>
            {/* Card 2: Organization */}
            <div className="bg-white rounded-2xl ring-1 ring-gray-100 p-8 md:p-10">
              <div className="h-5 w-32 bg-gray-100 rounded mb-4" />
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-x-6 gap-y-2.5">
                <div className="h-4 w-full bg-gray-100 rounded" />
                <div className="h-4 w-full bg-gray-100 rounded" />
              </div>
            </div>
          </div>

          {/* Right: AI Decision card skeleton */}
          <aside className="min-w-0 w-full xl:sticky xl:top-24">
            <div className="flex flex-col gap-3 animate-pulse">
              {/* Header */}
              <div className="mb-1">
                <div className="flex items-center gap-2 mb-1">
                  <Sparkles className="h-4 w-4 text-[#0052D9]" />
                  <div className="h-4 w-32 bg-gray-100 rounded" />
                </div>
                <div className="h-3 w-3/4 bg-gray-100 rounded" />
              </div>

              {/* Current Job card */}
              <div className="bg-white ring-1 ring-gray-100 rounded-xl p-4">
                <div className="flex items-center gap-2 mb-3">
                  <Target className="h-3.5 w-3.5 text-gray-300" />
                  <div className="h-3 w-16 bg-gray-100 rounded" />
                </div>
                <div className="h-4 w-3/4 bg-gray-100 rounded mb-2" />
                <div className="flex gap-1.5 mb-2">
                  <div className="h-5 w-12 bg-gray-100 rounded" />
                  <div className="h-5 w-16 bg-gray-100 rounded" />
                </div>
                <div className="h-3 w-1/2 bg-gray-100 rounded" />
              </div>

              {/* Current Resume card */}
              <div className="bg-white ring-1 ring-gray-100 rounded-xl p-4">
                <div className="flex justify-between items-start mb-3">
                  <div className="flex items-center gap-2">
                    <FileText className="h-3.5 w-3.5 text-gray-300" />
                    <div className="h-3 w-16 bg-gray-100 rounded" />
                  </div>
                  <div className="h-3 w-16 bg-gray-100 rounded" />
                </div>
                <div className="h-4 w-24 bg-gray-100 rounded mb-2" />
                <div className="flex gap-1.5 mb-2">
                  <div className="h-5 w-20 bg-gray-100 rounded" />
                  <div className="h-5 w-14 bg-gray-100 rounded" />
                </div>
                <div className="h-3 w-40 bg-gray-100 rounded" />
              </div>

              {/* Action card */}
              <div className="bg-[#F4F8FF]/80 border border-blue-200/60 rounded-xl p-5 text-center">
                <div className="h-4 w-40 bg-gray-200 rounded mx-auto mb-2" />
                <div className="h-3 w-3/4 bg-gray-200 rounded mx-auto mb-4" />
                <div className="flex flex-col gap-2.5">
                  <div className="h-9 w-full bg-[#0052D9]/30 rounded-lg" />
                  <div className="h-9 w-full bg-gray-200 rounded-lg" />
                </div>
              </div>

              {/* Tip card */}
              <div className="bg-orange-50/30 border border-orange-100/50 rounded-xl p-4">
                <div className="flex items-center gap-2 mb-2.5">
                  <div className="h-3.5 w-3.5 bg-orange-200 rounded" />
                  <div className="h-3 w-24 bg-orange-200 rounded" />
                </div>
                <div className="space-y-1.5">
                  <div className="h-3 w-full bg-orange-100 rounded" />
                  <div className="h-3 w-5/6 bg-orange-100 rounded" />
                  <div className="h-3 w-3/4 bg-orange-100 rounded" />
                </div>
              </div>
            </div>
          </aside>
        </div>
      </div>
    </>
  );
}
