import React from "react";
import {
  ResumeEditorSectionCard,
  type ResumeEditorSectionCardProps,
} from "../ResumeEditorSectionCard";

export interface BasicInfoSectionProps
  extends Omit<
    ResumeEditorSectionCardProps,
    "title" | "description" | "badge" | "defaultOpen" | "collapsible" | "required" | "children"
  > {
  /** V1 + V2 form content rendered from MyResumesPage */
  children?: React.ReactNode;
}

/**
 * Module 2: 基础信息
 *
 * Core module — expanded by default.
 * Fields (K8-5-22-F): name, school, degree, major, graduation_year, current_city, phone, email.
 */
export function BasicInfoSection({
  children,
  ...rest
}: BasicInfoSectionProps) {
  return (
    <ResumeEditorSectionCard
      title="基础信息"
      description="用于生成简历摘要和岗位匹配时的基础背景判断。"
      badge="必填"
      defaultOpen={true}
      required
      {...rest}
    >
      {children}
    </ResumeEditorSectionCard>
  );
}
