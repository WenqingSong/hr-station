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
      description="Offerr鹅 会优先根据求职方向、城市和岗位类型筛选更适合你的岗位。"
      badge="推荐"
      defaultOpen={true}
      {...rest}
    >
      {children}
    </ResumeEditorSectionCard>
  );
}
