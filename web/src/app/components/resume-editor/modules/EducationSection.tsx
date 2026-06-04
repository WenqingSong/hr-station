import React from "react";
import {
  ResumeEditorSectionCard,
  type ResumeEditorSectionCardProps,
} from "../ResumeEditorSectionCard";

export interface EducationSectionProps
  extends Omit<
    ResumeEditorSectionCardProps,
    "title" | "description" | "badge" | "defaultOpen" | "collapsible" | "required" | "children"
  > {
  children?: React.ReactNode;
  onAdd?: () => void;
  addLabel?: string;
}

/**
 * Module 4: 教育经历
 *
 * Core module — expanded by default.
 * Fields (K8-5-22-F): school, degree, major, department, startDate, endDate,
 *   gpa, gpaScale, rank, coreCourses, researchDirection, lab, paperSummary.
 */
export function EducationSection({
  children,
  onAdd,
  addLabel = "添加教育经历",
  ...rest
}: EducationSectionProps) {
  return (
    <ResumeEditorSectionCard
      title="教育经历"
      description="补充教育背景，帮助 Offerr鹅 判断你的学历、专业和岗位匹配基础。"
      badge="推荐"
      defaultOpen={true}
      {...rest}
    >
      {children}
    </ResumeEditorSectionCard>
  );
}
