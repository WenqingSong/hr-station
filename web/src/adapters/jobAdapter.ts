// ─── 岗位数据 Adapter ───────────────────────────────────────────────────────────

import type { Job, JobDetail } from '../types';
import { formatListText } from './common';

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

// ─── 类别/类型轻量映射 ─────────────────────────────────────────────────────────

const CATEGORY_LABELS: Record<string, string> = {
  backend: '后台开发',
  cloud: '云计算',
  algorithm: '算法',
  client: '客户端',
  data: '数据开发',
  product: '产品',
  game: '游戏',
};

const TYPE_LABELS: Record<string, string> = {
  campus: '校园招聘',
  internship: '实习生',
  early_batch: '提前批',
};

// ─── 转换函数 ──────────────────────────────────────────────────────────────────

export function toJobCardViewModel(job: Job): JobCardViewModel {
  return {
    id: job.id,
    title: job.title,
    businessGroup: job.business_group,
    locationsText: formatListText(job.locations, '多地'),
    tags: job.tags,
    shortDescription: job.short_description,
    statusLabel: job.status,
    categoryLabel: CATEGORY_LABELS[job.category] || job.category,
    typeLabel: TYPE_LABELS[job.type_label] || job.type_label,
  };
}

export function toJobCardViewModels(jobs: Job[]): JobCardViewModel[] {
  return jobs.map(toJobCardViewModel);
}

export function toJobDetailViewModel(job: JobDetail): JobDetailViewModel {
  return {
    ...toJobCardViewModel(job),
    bannerSummary: job.banner_summary,
    responsibilities: job.responsibilities,
    requirements: job.requirements,
    bonusPoints: job.bonus_points,
    deliveryNotes: job.delivery_notes,
    keywords: job.tags,
  };
}
