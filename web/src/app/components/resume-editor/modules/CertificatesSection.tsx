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
 * Module 8: 证书与语言能力
 *
 * New module — collapsed by default.
 * Fields (K8-5-22-H): name, type, level, score, issuedAt (array).
 */
export function CertificatesSection({
  children,
  onAdd,
  addLabel = "添加证书",
  ...rest
}: CertificatesSectionProps) {
  return (
    <ResumeEditorSectionCard
      title="证书与语言能力"
      description="补充语言成绩（四六级/雅思/托福等）和专业证书，提升简历信息完整度。"
      badge="可选"
      defaultOpen={false}
      collapsible
      {...rest}
    >
      {children}
    </ResumeEditorSectionCard>
  );
}
