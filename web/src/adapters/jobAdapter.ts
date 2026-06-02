// ─── 岗位数据 Adapter ───────────────────────────────────────────────────────────

import type { JobApiItem } from '../api/jobs';

// ─── ViewModel 类型 ────────────────────────────────────────────────────────────

export interface JobCardViewModel {
  id: string;
  title: string;
  businessGroup: string;
  locationsText: string;
  tags: string[];
  shortDescription: string;
  statusLabel: string;
  categoryLabel: string;
  typeLabel: string;
}

export interface JobDetailViewModel extends JobCardViewModel {
  bannerSummary: string;
  responsibilities: string[];
  requirements: string[];
  bonusPoints: string[];
  deliveryNotes?: string;
  keywords: string[];
}

// ─── 标签映射（对齐后端 seed 数据值）─────────────────────────────────────────────

const RECRUIT_TYPE_LABELS: Record<string, string> = {
  '校招': '校园招聘',
  '实习': '实习生',
  '提前批': '提前批',
};

const STATUS_LABELS: Record<string, string> = {
  'active': '招聘中',
};

// ─── 后端原始类型 → ViewModel ──────────────────────────────────────────────────

/** 从后端 JobApiItem 生成 JobCardViewModel */
export function toJobCardViewModelFromApi(job: JobApiItem): JobCardViewModel {
  return {
    id: String(job.id),
    title: job.title,
    businessGroup: job.department,
    locationsText: job.city,
    tags: job.skill_keywords ?? [],
    shortDescription: job.description ?? '',
    statusLabel: STATUS_LABELS[job.status] || job.status,
    categoryLabel: job.category,
    typeLabel: RECRUIT_TYPE_LABELS[job.recruit_type] || job.recruit_type,
  };
}

/** 批量转换 */
export function toJobCardViewModelsFromApi(jobs: JobApiItem[]): JobCardViewModel[] {
  return jobs.map(toJobCardViewModelFromApi);
}

/** 从后端 JobApiItem 生成 JobDetailViewModel */
export function toJobDetailViewModelFromApi(job?: JobApiItem | null): JobDetailViewModel {
  if (!job) {
    return {
      id: '', title: '', businessGroup: '', locationsText: '', tags: [],
      shortDescription: '', statusLabel: '', categoryLabel: '', typeLabel: '',
      bannerSummary: '', responsibilities: [], requirements: [], bonusPoints: [],
      keywords: [],
    };
  }
  const cardVm = toJobCardViewModelFromApi(job);
  const reqLines = job.requirements
    ? job.requirements.split('\n').map(l => l.replace(/^\d+[\.\)、\s]+/, '').trim()).filter(Boolean)
    : [];
  return {
    ...cardVm,
    bannerSummary: job.description ?? '',
    responsibilities: job.description
      ? job.description.split(/[。；]/).map(s => s.trim()).filter(Boolean)
      : [],
    requirements: reqLines,
    bonusPoints: [],
    deliveryNotes: undefined,
    keywords: job.skill_keywords ?? [],
  };
}
