import React from "react";
import { useParams } from "react-router";
import { JobDetailBanner } from "../components/jobs/JobDetailBanner";
import { JobDetailJD } from "../components/jobs/JobDetailJD";
import { JobDetailAIDecision } from "../components/JobDetailAIDecision";
import { TwoColumnPageShell } from "../layouts/TwoColumnPageShell";
import { AISidebarShell } from "../layouts/AISidebarShell";

/**
 * K8-5-2-C：接入统一 TwoColumnPageShell + AISidebarShell
 *
 * 设计说明：
 * - 使用 TwoColumnPageShell 管理两栏布局
 * - 使用 AISidebarShell 管理右侧 AI 的 sticky 行为
 * - JobDetailAIDecision 不再承担页面级布局职责
 */
export function JobDetailPage() {
  const { id: jobId } = useParams<{ id: string }>();
  return (
    <>
      <JobDetailBanner jobId={jobId} />
      <TwoColumnPageShell
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
