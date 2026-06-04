import React from "react";
import {
  ResumeEditorSectionCard,
  type ResumeEditorSectionCardProps,
} from "../ResumeEditorSectionCard";

export type ResumeBasicsSectionProps = Omit<
  ResumeEditorSectionCardProps,
  "title" | "description" | "badge" | "defaultOpen" | "collapsible" | "required" | "children"
> & {
  children: React.ReactNode;
};

/**
 * Module 1: 简历基础
 *
 * Core module — always expanded, non-collapsible.
 * Contains: resume name, target direction, description.
 *
 * In K8-5-22-E, the existing V1 form content for this module is
 * passed as children from MyResumesPage.
 */
export function ResumeBasicsSection({ children, ...rest }: ResumeBasicsSectionProps) {
  return (
    <ResumeEditorSectionCard
      title="简历基础"
      description="先确定这份简历的求职定位，后续推荐、匹配和优化都会基于这里的信息。"
      badge="必填"
      defaultOpen={true}
      collapsible={false}
      required
      {...rest}
    >
      {children}
    </ResumeEditorSectionCard>
  );
}
