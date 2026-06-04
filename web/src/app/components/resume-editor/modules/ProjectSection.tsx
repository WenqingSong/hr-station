import React from "react";
import {
  ResumeEditorSectionCard,
  type ResumeEditorSectionCardProps,
} from "../ResumeEditorSectionCard";

export interface ProjectSectionProps
  extends Omit<
    ResumeEditorSectionCardProps,
    "title" | "description" | "badge" | "defaultOpen" | "collapsible" | "required" | "children"
  > {
  children?: React.ReactNode;
}

/**
 * Module 5: 项目经历
 *
 * Core module — expanded by default.
 * Fields (K8-5-22-G): name, role, type, startDate, endDate,
 *   techStack, description, responsibilities, achievements, link.
 */
export function ProjectSection({
  children,
  ...rest
}: ProjectSectionProps) {
  return (
    <ResumeEditorSectionCard
      title="项目经历"
      description="项目经历是技术岗位匹配的重要依据，建议写清楚技术栈、职责和量化结果。"
      badge="推荐"
      defaultOpen={true}
      {...rest}
    >
      {children}
    </ResumeEditorSectionCard>
  );
}
