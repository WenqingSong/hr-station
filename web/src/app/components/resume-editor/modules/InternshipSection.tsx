import React from "react";
import {
  ResumeEditorSectionCard,
  type ResumeEditorSectionCardProps,
} from "../ResumeEditorSectionCard";

export interface InternshipSectionProps
  extends Omit<
    ResumeEditorSectionCardProps,
    "title" | "description" | "badge" | "defaultOpen" | "collapsible" | "required" | "children"
  > {
  children?: React.ReactNode;
}

/**
 * Module 7: 实习经历
 *
 * Optional module — collapsed by default.
 * Fields (K8-5-22-G): company, role, startDate, endDate,
 *   workContent, techStack, achievements, metrics.
 */
export function InternshipSection({
  children,
  ...rest
}: InternshipSectionProps) {
  return (
    <ResumeEditorSectionCard
      title="实习经历"
      description="实习经历会显著影响岗位匹配和简历优化建议，建议写清楚职责、技术和结果。"
      badge="可选"
      defaultOpen={false}
      collapsible
      {...rest}
    >
      {children}
    </ResumeEditorSectionCard>
  );
}
