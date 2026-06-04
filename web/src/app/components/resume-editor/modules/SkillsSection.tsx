import React from "react";
import {
  ResumeEditorSectionCard,
  type ResumeEditorSectionCardProps,
} from "../ResumeEditorSectionCard";

export interface SkillsSectionProps
  extends Omit<
    ResumeEditorSectionCardProps,
    "title" | "description" | "badge" | "defaultOpen" | "collapsible" | "required" | "children"
  > {
  children?: React.ReactNode;
}

/**
 * Module 5: 技能信息
 *
 * Core module — expanded by default.
 * K8-5-25-D: Tencent-style 3 fields — languageExam, developmentLanguages, aiApplication.
 */
export function SkillsSection({
  children,
  ...rest
}: SkillsSectionProps) {
  return (
    <ResumeEditorSectionCard
      title="技能信息"
      description="补充你掌握的技术、工具和能力标签，Offerr鹅 会基于这些关键词进行岗位推荐、匹配分析和简历优化建议。"
      badge="可选"
      defaultOpen={true}
      {...rest}
    >
      {children}
    </ResumeEditorSectionCard>
  );
}
