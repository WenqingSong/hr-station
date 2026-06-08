import React from "react";
import { ContentPageShell } from "../layout/ContentPageShell";
import { Sparkles } from "lucide-react";

interface MatchAnalysisSkeletonProps {
  message?: string;
}

/**
 * MatchAnalysisSkeleton — 匹配分析页三栏骨架屏
 *
 * 模拟真实页面三栏布局，在数据加载期间保持页面结构稳定：
 * - 页面标题区
 * - 左侧：岗位 JD 摘要骨架（w-[320px]）
 * - 中间：AI 匹配分析报告骨架（flex-1）
 * - 右侧：当前简历与深度分析卡片骨架（w-[300px]）
 */
export function MatchAnalysisSkeleton({ message }: MatchAnalysisSkeletonProps) {
  return (
    <ContentPageShell maxWidthClassName="max-w-[1440px]" contentClassName="pb-12">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-[22px] font-semibold text-gray-900 flex items-center gap-2">
          <Sparkles className="w-5 h-5 text-[#0052D9]" />
          岗位匹配分析
        </h1>
        <div className="flex items-center gap-3 mt-3">
          <span className="text-[#0052D9] font-medium text-xs flex items-center gap-1.5">
            <span className="inline-block h-3 w-3 animate-spin rounded-full border-2 border-blue-100 border-t-[#0052D9]" />
            {message || "正在加载…"}
          </span>
        </div>
      </div>

      {/* Three-column skeleton */}
      <div className="flex gap-6 items-start">
        {/* ── Left sidebar — Job JD skeleton ── */}
        <aside className="w-[320px] shrink-0">
          <div className="bg-white border border-gray-200 rounded-lg p-4 animate-pulse">
            {/* Title */}
            <div className="h-4 w-20 bg-gray-100 rounded mb-4" />
            {/* Job meta lines */}
            <div className="h-5 w-3/4 bg-gray-100 rounded mb-2" />
            <div className="h-3.5 w-1/2 bg-gray-100 rounded mb-4" />
            {/* Requirements label */}
            <div className="h-4 w-16 bg-gray-100 rounded mb-3" />
            {/* Requirement lines */}
            <div className="space-y-2 mb-4">
              <div className="h-3 w-full bg-gray-100 rounded" />
              <div className="h-3 w-5/6 bg-gray-100 rounded" />
              <div className="h-3 w-4/5 bg-gray-100 rounded" />
              <div className="h-3 w-3/4 bg-gray-100 rounded" />
            </div>
            {/* Keywords */}
            <div className="h-4 w-16 bg-gray-100 rounded mb-2" />
            <div className="flex flex-wrap gap-1">
              <div className="h-5 w-16 bg-gray-100 rounded" />
              <div className="h-5 w-12 bg-gray-100 rounded" />
              <div className="h-5 w-20 bg-gray-100 rounded" />
              <div className="h-5 w-14 bg-gray-100 rounded" />
            </div>
          </div>
        </aside>

        {/* ── Center — AI report skeleton ── */}
        <div className="flex-1 min-w-0 max-w-[680px]">
          <div className="bg-white border border-gray-200 rounded-lg shadow-sm overflow-hidden animate-pulse">
            {/* Decision card header */}
            <div className="bg-[#F4F8FF] p-5 border-b border-blue-100">
              <div className="h-5 w-32 bg-blue-100/60 rounded mb-3" />
              <div className="flex items-center gap-3 mb-2">
                <div className="h-5 w-16 bg-blue-100/60 rounded" />
                <div className="h-5 w-16 bg-blue-100/60 rounded" />
              </div>
              <div className="h-4 w-3/4 bg-blue-100/60 rounded" />
            </div>
            {/* Report sections */}
            <div className="p-5 space-y-5">
              {/* Section 1 */}
              <div>
                <div className="h-4 w-24 bg-gray-100 rounded mb-3" />
                <div className="space-y-2">
                  <div className="h-3.5 w-full bg-gray-100 rounded" />
                  <div className="h-3.5 w-5/6 bg-gray-100 rounded" />
                </div>
              </div>
              {/* Section 2 */}
              <div>
                <div className="h-4 w-28 bg-gray-100 rounded mb-3" />
                <div className="space-y-2">
                  <div className="h-3.5 w-full bg-gray-100 rounded" />
                  <div className="h-3.5 w-4/5 bg-gray-100 rounded" />
                  <div className="h-3.5 w-3/4 bg-gray-100 rounded" />
                </div>
              </div>
              {/* Section 3 */}
              <div>
                <div className="h-4 w-20 bg-gray-100 rounded mb-3" />
                <div className="space-y-2">
                  <div className="h-3.5 w-full bg-gray-100 rounded" />
                  <div className="h-3.5 w-2/3 bg-gray-100 rounded" />
                </div>
              </div>
              {/* Section 4 — suggestion */}
              <div>
                <div className="h-4 w-20 bg-gray-100 rounded mb-3" />
                <div className="h-3.5 w-full bg-gray-100 rounded" />
                <div className="h-3.5 w-1/2 bg-gray-100 rounded mt-2" />
              </div>
            </div>
          </div>
        </div>

        {/* ── Right sidebar — Resume + Deep button skeleton ── */}
        <aside className="w-[300px] shrink-0 flex flex-col gap-4">
          {/* Resume card */}
          <div className="bg-white border border-gray-200 rounded-lg p-4 animate-pulse">
            <div className="h-3.5 w-16 bg-gray-100 rounded mb-3" />
            <div className="h-4.5 w-24 bg-gray-100 rounded mb-1" />
            <div className="h-3.5 w-32 bg-gray-100 rounded mb-2" />
            {/* Tags */}
            <div className="flex flex-wrap gap-1 mb-2">
              <div className="h-5 w-12 bg-gray-100 rounded" />
              <div className="h-5 w-16 bg-gray-100 rounded" />
              <div className="h-5 w-10 bg-gray-100 rounded" />
            </div>
            <div className="h-3.5 w-40 bg-gray-100 rounded mb-1" />
          </div>

          {/* Resume evidence skeleton */}
          <div className="bg-white border border-gray-200 rounded-lg p-4 animate-pulse">
            <div className="h-3.5 w-20 bg-gray-100 rounded mb-3" />
            <div className="space-y-1.5">
              <div className="h-3 w-full bg-gray-100 rounded" />
              <div className="h-3 w-5/6 bg-gray-100 rounded" />
              <div className="h-3 w-4/5 bg-gray-100 rounded" />
              <div className="h-3 w-3/4 bg-gray-100 rounded" />
            </div>
          </div>

          {/* Deep analysis card */}
          <div className="bg-white border border-gray-200 rounded-lg p-4 animate-pulse">
            <div className="h-3.5 w-20 bg-gray-100 rounded mb-3" />
            <div className="h-3 w-3/4 bg-gray-100 rounded mb-2" />
            <div className="h-3 w-5/6 bg-gray-100 rounded mb-3" />
            <div className="h-9 w-full bg-gray-100 rounded-lg" />
          </div>
        </aside>
      </div>
    </ContentPageShell>
  );
}
