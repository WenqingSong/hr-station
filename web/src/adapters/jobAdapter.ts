// ─── 岗位数据 Adapter ───────────────────────────────────────────────────────────

import type { Job, JobDetail } from '../types';
import type { JobApiItem } from '../api/jobs';
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

/**
 * 从后端 JobApiItem 生成 JobCardViewModel。
 * 字段映射：department → businessGroup, city → locationsText,
 * skill_keywords → tags, recruit_type → typeLabel, description → shortDescription。
 */
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

// ─── 旧版转换函数（保留向后兼容，供 Detail flow 等未迁移代码使用）──────────────

const LEGACY_CATEGORY_LABELS: Record<string, string> = {
  backend: '后台开发',
  cloud: '云计算',
  algorithm: '算法',
  client: '客户端',
  data: '数据开发',
  product: '产品',
  game: '游戏',
};

const LEGACY_TYPE_LABELS: Record<string, string> = {
  campus: '校园招聘',
  internship: '实习生',
  early_batch: '提前批',
};

export function toJobCardViewModel(job: Job): JobCardViewModel {
  return {
    id: job.id,
    title: job.title,
    businessGroup: job.business_group,
    locationsText: formatListText(job.locations, '多地'),
    tags: job.tags,
    shortDescription: job.short_description,
    statusLabel: job.status,
    categoryLabel: LEGACY_CATEGORY_LABELS[job.category] || job.category,
    typeLabel: LEGACY_TYPE_LABELS[job.type_label] || job.type_label,
  };
}

export function toJobCardViewModels(jobs: Job[]): JobCardViewModel[] {
  return jobs.map(toJobCardViewModel);
}

export function toJobDetailViewModel(job?: JobDetail | null): JobDetailViewModel {
  if (!job) return { id: '', title: '', businessGroup: '', locationsText: '', tags: [], shortDescription: '', statusLabel: '', categoryLabel: '', typeLabel: '', bannerSummary: '', responsibilities: [], requirements: [], bonusPoints: [], keywords: [] };
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
