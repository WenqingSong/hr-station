import React from "react";
import {
  ResumeEditorSectionCard,
  type ResumeEditorSectionCardProps,
} from "../ResumeEditorSectionCard";

export interface JobIntentionSectionProps
  extends Omit<
    ResumeEditorSectionCardProps,
    "title" | "description" | "badge" | "defaultOpen" | "collapsible" | "required" | "children"
  > {
  children?: React.ReactNode;
}

/**
 * Module 3: 求职意向
 *
 * Core module — expanded by default.
 * Fields (K8-5-22-F): direction, jobType, cities, acceptOtherCities,
 *   earliestStartDate, internshipDuration, weeklyAvailableDays, targetIndustries, targetCompanies.
 */
export function JobIntentionSection({
  children,
  ...rest
}: JobIntentionSectionProps) {
  return (
    <ResumeEditorSectionCard
      title="求职意向"
      description="如果你已经有明确方向，可以填写；暂时不确定也可以留空，Offerr鹅 会基于教育、技能和项目经历推荐适合的岗位。"
      badge="推荐"
      defaultOpen={true}
      {...rest}
    >
      {children}
    </ResumeEditorSectionCard>
  );
}
