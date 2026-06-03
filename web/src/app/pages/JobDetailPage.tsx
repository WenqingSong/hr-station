import React from "react";
import { useParams } from "react-router";
import { JobDetailBanner } from "../components/jobs/JobDetailBanner";
import { JobDetailJD } from "../components/jobs/JobDetailJD";
import { JobDetailAIDecision } from "../components/ai/JobDetailAIDecision";
import { TwoColumnPageShell } from "../layouts/TwoColumnPageShell";
import { AISidebarShell } from "../layouts/AISidebarShell";

/**
 * K8-5-2-C：接入统一 TwoColumnPageShell + AISidebarShell
 * K8-5-9-B：布局修复——放宽宽度 + 右侧 AI 取消内部滚动
 *
 * 设计说明：
 * - 使用 TwoColumnPageShell 管理两栏布局（max-w-[1280px]）
 * - 使用 AISidebarShell 管理右侧 AI 的 sticky 行为（不内部滚动）
 * - JobDetailAIDecision 不承担页面级布局职责
 */
export function JobDetailPage() {
  const { id: jobId } = useParams<{ id: string }>();
  return (
    <>
      <JobDetailBanner jobId={jobId} />
      <TwoColumnPageShell
        maxWidthClassName="max-w-[1280px]"
        gridColumnsClassName="xl:grid-cols-[minmax(0,1fr)_320px]"
        className="pt-8 md:pt-10"
        contentClassName="space-y-6"
        aside={
          <AISidebarShell>
            <JobDetailAIDecision />
          </AISidebarShell>
        }
      >
        <JobDetailJD />
      </TwoColumnPageShell>
    </>
  );
}
