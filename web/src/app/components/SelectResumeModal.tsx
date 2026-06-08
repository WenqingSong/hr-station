import React, { useEffect, useState } from "react";
import { useNavigate, useLocation } from "react-router";
import { loadResumeOptionViewModelsFromApi } from "../../api";
import type { ResumeOptionViewModel } from "../../adapters";
import { getAuthRedirectIf401 } from "../utils/authRedirect";
import { ResumePickerModal } from "./ResumePickerModal";
import type { ResumePickerOption } from "./ResumePickerModal";

// ─── Types ────────────────────────────────────────────────────────────────────

type Scene = "recommendation" | "match-analysis" | "resume-optimization";

interface SelectResumeModalProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm?: (resumeId: string) => void;
  scene?: Scene;
  /** 目标岗位名称（可选，用于 header 显示） */
  jobTitle?: string;
  /** K9-2-Entry-B-Fix: 初始选中简历 ID（来自父组件当前简历） */
  initialSelectedResumeId?: number | null;
}

// ─── Scene config ─────────────────────────────────────────────────────────────

const SCENE_CONFIG: Record<
  Scene,
  { title: string; description: string; confirmText: string }
> = {
  recommendation: {
    title: "选择用于推荐的简历",
    description: "系统将从当前岗位池中推荐最值得优先关注的 5 个岗位。",
    confirmText: "确认并开始推荐",
  },
  "match-analysis": {
    title: "选择用于分析的简历",
    description: "系统将基于所选简历与当前岗位 JD 进行匹配分析。",
    confirmText: "确认并开始分析",
  },
  "resume-optimization": {
    title: "选择用于优化的简历",
    description: "系统将基于所选简历与当前岗位 JD 生成定向优化建议。",
    confirmText: "确认并生成建议",
  },
};

// ─── Helpers ──────────────────────────────────────────────────────────────────

function toPickerOption(vm: ResumeOptionViewModel): ResumePickerOption {
  return {
    id: vm.id,
    name: vm.name,
    isCurrent: vm.isCurrent,
    statusLabel: vm.statusLabel,
    directionLabel: vm.directionLabel,
    tags: vm.tags || [],
    projectCount: vm.projectCount,
    updatedAt: vm.updatedAt,
  };
}

// ─── Component ────────────────────────────────────────────────────────────────

/**
 * SelectResumeModal — 岗位详情页简历选择弹窗（thin wrapper）
 *
 * 负责：
 * - 按 scene 配置 title / description / confirmText
 * - 加载简历 options（API）
 * - 管理默认选中逻辑
 * - 调用方通过 onConfirm(resumeId: string) 获取选择结果
 *
 * UI 委托给 ResumePickerModal（Portal + blur + body lock + ARIA）
 */
export function SelectResumeModal({
  isOpen,
  onClose,
  onConfirm,
  scene = "match-analysis",
  jobTitle = "",
  initialSelectedResumeId = null,
}: SelectResumeModalProps) {
  const [resumes, setResumes] = useState<ResumeOptionViewModel[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [selectedIds, setSelectedIds] = useState<number[]>([]);
  const navigate = useNavigate();
  const location = useLocation();

  // ── Load resumes on open ─────────────────────────────────────────────
  useEffect(() => {
    if (!isOpen) {
      setSelectedIds([]);
      setError(null);
      return;
    }
    setLoading(true);
    setError(null);
    loadResumeOptionViewModelsFromApi()
      .then((opts) => {
        setResumes(opts);
        if (opts.length > 0) {
          const initial =
            initialSelectedResumeId != null
              ? opts.find((o) => o.id === initialSelectedResumeId)
              : undefined;
          setSelectedIds(initial ? [initial.id] : [(opts.find((o) => o.isCurrent) || opts[0]).id]);
        } else {
          setSelectedIds([]);
        }
      })
      .catch((err) => {
        const authRedirect = getAuthRedirectIf401(
          err,
          location.pathname,
          location.search,
        );
        if (authRedirect) {
          navigate(authRedirect, { replace: true });
          return;
        }
        setError("简历列表加载失败，请稍后重试。");
      })
      .finally(() => setLoading(false));
  }, [isOpen, initialSelectedResumeId]);

  // ── Retry ────────────────────────────────────────────────────────────
  const handleRetry = () => {
    setLoading(true);
    setError(null);
    loadResumeOptionViewModelsFromApi()
      .then((opts) => {
        setResumes(opts);
        if (opts.length > 0) {
          const initial =
            initialSelectedResumeId != null
              ? opts.find((o) => o.id === initialSelectedResumeId)
              : undefined;
          setSelectedIds(initial ? [initial.id] : [(opts.find((o) => o.isCurrent) || opts[0]).id]);
        } else {
          setSelectedIds([]);
        }
      })
      .catch(() => setError("简历列表加载失败，请稍后重试。"))
      .finally(() => setLoading(false));
  };

  // ── Confirm handler ──────────────────────────────────────────────────
  const handleConfirm = (ids: number[]) => {
    if (ids.length > 0) {
      onConfirm?.(String(ids[0]));
    }
  };

  const cfg = SCENE_CONFIG[scene];
  const pickerOptions: ResumePickerOption[] = resumes.map(toPickerOption);

  return (
    <ResumePickerModal
      open={isOpen}
      title={cfg.title}
      description={cfg.description}
      targetLabel={jobTitle || undefined}
      resumes={pickerOptions}
      loading={loading}
      error={error}
      onRetry={handleRetry}
      selectedIds={selectedIds}
      selectionMode="single"
      onSelectionChange={setSelectedIds}
      confirmText={cfg.confirmText}
      onConfirm={handleConfirm}
      onClose={onClose}
    />
  );
}
