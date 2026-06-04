import React from "react";
import {
  ResumeEditorSectionCard,
  type ResumeEditorSectionCardProps,
} from "../ResumeEditorSectionCard";

export interface AwardsSectionProps
  extends Omit<
    ResumeEditorSectionCardProps,
    "title" | "description" | "badge" | "defaultOpen" | "collapsible" | "required" | "children"
  > {
  children?: React.ReactNode;
  onAdd?: () => void;
  addLabel?: string;
}

/**
 * Module 9: 获奖与竞赛
 *
 * New module — collapsed by default.
 * Fields (K8-5-22-H): name, type, level, awardedAt, description (array).
 */
export function AwardsSection({
  children,
  onAdd,
  addLabel = "添加获奖",
  ...rest
}: AwardsSectionProps) {
  return (
    <ResumeEditorSectionCard
      title="获奖与竞赛"
      description="补充竞赛获奖和学术荣誉，帮助岗位匹配时展示综合竞争力。"
      badge="可选"
      defaultOpen={false}
      collapsible
      {...rest}
    >
      {children}
    </ResumeEditorSectionCard>
  );
}
