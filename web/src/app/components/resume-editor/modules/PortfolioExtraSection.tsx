import React from "react";
import {
  ResumeEditorSectionCard,
  type ResumeEditorSectionCardProps,
} from "../ResumeEditorSectionCard";

export interface PortfolioExtraSectionProps
  extends Omit<
    ResumeEditorSectionCardProps,
    "title" | "description" | "badge" | "defaultOpen" | "collapsible" | "required" | "children"
  > {
  children?: React.ReactNode;
}

/**
 * Module 8/9: 作品或个人主页 / 自我评价 (K8-5-25-E: split)
 */
export function PortfolioExtraSection({
  children,
  ...rest
}: PortfolioExtraSectionProps) {
  return (
    <ResumeEditorSectionCard
      title="作品链接与补充信息"
      description="补充 GitHub、博客、作品集链接和个人补充信息，让简历更立体。"
      badge="可选"
      defaultOpen={false}
      collapsible
      {...rest}
    >
      {children}
    </ResumeEditorSectionCard>
  );
}
