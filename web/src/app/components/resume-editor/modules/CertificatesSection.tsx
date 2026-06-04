import React from "react";
import {
  ResumeEditorSectionCard,
  type ResumeEditorSectionCardProps,
} from "../ResumeEditorSectionCard";

export interface CertificatesSectionProps
  extends Omit<
    ResumeEditorSectionCardProps,
    "title" | "description" | "badge" | "defaultOpen" | "collapsible" | "required" | "children"
  > {
  children?: React.ReactNode;
  onAdd?: () => void;
  addLabel?: string;
}

/**
 * Module 7: 证书与获奖 (merged)
 *
 * Collapsed by default.
 * K8-5-25-E: merged with AwardsSection.
 */
export function CertificatesSection({
  children,
  onAdd,
  addLabel = "添加证书",
  ...rest
}: CertificatesSectionProps) {
  return (
    <ResumeEditorSectionCard
      title="证书与获奖"
      description="补充语言成绩、专业证书和竞赛获奖，提升简历信息完整度。"
      badge="可选"
      defaultOpen={false}
      collapsible
      {...rest}
    >
      {children}
    </ResumeEditorSectionCard>
  );
}
