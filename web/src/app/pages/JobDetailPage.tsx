import React from "react";
import { useParams } from "react-router";
import { JobDetailBanner } from "../components/JobDetailBanner";
import { JobDetailJD } from "../components/JobDetailJD";
import { JobDetailAIDecision } from "../components/JobDetailAIDecision";

/**
 * K8-5-1-R：修复岗位详情页右侧 AI 模块布局崩坏问题
 * 
 * 设计说明：
 * - 桌面端：左侧主内容（岗位 JD + 招聘部门和工作地），右侧 AI 投递决策栏
 * - 移动端：主内容在上，AI 模块在下
 * - 使用 Grid 布局实现两栏对齐
 * - 右侧 AI 模块使用 sticky 保持可见
 */
export function JobDetailPage() {
  const { id: jobId } = useParams<{ id: string }>();
  return (
    <>
      <JobDetailBanner jobId={jobId} />
      <div className="mx-auto max-w-[1200px] px-6 md:px-8 py-6">
        {/* K8-5-1-R：修复两栏布局，AI 模块回到右侧栏 */}
        <div className="grid grid-cols-1 lg:grid-cols-[1fr_340px] gap-6 items-start">
          {/* 左侧主内容区域：岗位 JD + 招聘部门和工作地 */}
          <div className="min-w-0 space-y-6">
            <JobDetailJD jobId={jobId} />
          </div>
          
          {/* 右侧 AI 投递决策栏 */}
          <div className="min-w-0 lg:sticky lg:top-24">
            <JobDetailAIDecision />
          </div>
        </div>
      </div>
    </>
  );
}
