import React from "react";
import { useParams } from "react-router";
import { JobDetailBanner } from "../components/JobDetailBanner";
import { JobDetailJD } from "../components/JobDetailJD";
import { JobDetailAIDecision } from "../components/JobDetailAIDecision";

/**
 * K8-5-1：重构岗位详情页布局，对齐腾讯招聘详情页
 * 
 * 设计说明：
 * - 主内容区域（左侧）更宽，优先展示岗位 JD
 * - AI 模块下移到主内容后，作为辅助卡片
 * - 页面居中对齐，卡片宽度、间距、对齐统一
 */
export function JobDetailPage() {
  const { id: jobId } = useParams<{ id: string }>();
  return (
    <>
      <JobDetailBanner jobId={jobId} />
      <div className="mx-auto max-w-[1200px] px-6 md:px-8 py-6">
        {/* 主内容区域 + AI 模块 */}
        <div className="space-y-6">
          {/* 主内容区域：岗位 JD + 招聘部门和工作地 */}
          <div className="flex-1 max-w-[860px] mx-auto">
            <JobDetailJD jobId={jobId} />
          </div>
          
          {/* AI 模块下移到主内容后，作为辅助卡片 */}
          <div className="max-w-[860px] mx-auto">
            <JobDetailAIDecision />
          </div>
        </div>
      </div>
    </>
  );
}
