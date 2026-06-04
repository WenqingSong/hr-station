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
 * Contains: resume name only.
 *
 * K8-5-24-B: Removed target direction & description — recommendation
 * should not require users to pre-specify a job direction.
 */
export function ResumeBasicsSection({ children, ...rest }: ResumeBasicsSectionProps) {
  return (
    <ResumeEditorSectionCard
      title="简历基础"
      description="给这份简历起一个便于区分的名称。"
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
