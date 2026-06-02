// ─── 岗位数据 Adapter ───────────────────────────────────────────────────────────

import type { JobApiItem } from '../api/jobs';

// ─── ViewModel 类型 ────────────────────────────────────────────────────────────

export interface JobCardViewModel {
  id: string;
  title: string;
  businessGroup: string;
  businessGroups: string[];
  locationsText: string;
  locationsList: string[];
  tags: string[];
  shortDescription: string;
  statusLabel: string;
  categoryLabel: string;
  typeLabel: string;
  direction: string;
  applicationType: string;
}

export interface JobDetailViewModel extends JobCardViewModel {
  bannerSummary: string;
  slogan: string;
  summary: string;
  applicationType: string;
  applicationProject: string;
  applicationSubtype: string;
  direction: string;
  category: string;
  responsibilities: string[];
  mustHave: string[];
  niceToHave: string[];
  bonusPoints: string[];
  notes: string[];
  organizations: { code: string; name: string }[];
  locations: string[];
  interviewLocations: string[];
  skills: string[];
  externalId: string;
  sourceName: string;
  deliveryNotes?: string;
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

/** 安全解析 JSON 数组 */
function parseStringArray(raw: unknown): string[] {
  if (Array.isArray(raw)) return raw.filter((v): v is string => typeof v === "string");
  return [];
}

/** 安全解析 organizations [{code, name}] → name 列表 */
function parseOrgNames(raw: unknown): string[] {
  if (Array.isArray(raw)) {
    return raw
      .filter((v): v is Record<string, unknown> => v !== null && typeof v === "object")
      .map((v) => (v as Record<string, unknown>).name)
      .filter((v): v is string => typeof v === "string");
  }
  return [];
}

/** 从后端 JobApiItem 生成 JobCardViewModel */
export function toJobCardViewModelFromApi(job: JobApiItem): JobCardViewModel {
  const locationsRaw = parseStringArray((job as Record<string, unknown>).locations);
  const orgsRaw = parseOrgNames((job as Record<string, unknown>).organizations);
  return {
    id: String(job.id),
    title: job.title,
    businessGroup: job.department,
    businessGroups: orgsRaw.length > 0 ? orgsRaw : (job.department ? [job.department] : []),
    locationsText: job.city,
    locationsList: locationsRaw.length > 0 ? locationsRaw : (job.city ? [job.city] : []),
    tags: Array.isArray(job.skill_keywords) ? job.skill_keywords : [],
    shortDescription: typeof job.description === "string" ? job.description : '',
    statusLabel: STATUS_LABELS[job.status] || job.status,
    categoryLabel: job.category,
    typeLabel: RECRUIT_TYPE_LABELS[job.recruit_type] || job.recruit_type || '',
    direction: job.direction || '',
    applicationType: job.application_type || '',
  };
}

/** 批量转换 */
export function toJobCardViewModelsFromApi(jobs: JobApiItem[]): JobCardViewModel[] {
  return jobs.map(toJobCardViewModelFromApi);
}

/** 从后端 JobApiItem 生成 JobDetailViewModel */
export function toJobDetailViewModelFromApi(job?: JobApiItem | null): JobDetailViewModel {
  const empty = {
    id: '', title: '', businessGroup: '', locationsText: '', tags: [],
    shortDescription: '', statusLabel: '', categoryLabel: '', typeLabel: '',
    bannerSummary: '', slogan: '', summary: '',
    applicationType: '', applicationProject: '', applicationSubtype: '',
    direction: '', category: '',
    responsibilities: [], mustHave: [], niceToHave: [], bonusPoints: [], notes: [],
    organizations: [], locations: [], interviewLocations: [], skills: [],
    externalId: '', sourceName: '',
  };
  if (!job) return empty;
  const cardVm = toJobCardViewModelFromApi(job);
  return {
    ...cardVm,
    bannerSummary: job.description ?? '',
    slogan: job.slogan ?? '',
    summary: job.summary ?? '',
    applicationType: job.application_type ?? '',
    applicationProject: job.application_project ?? '',
    applicationSubtype: job.application_subtype ?? '',
    direction: job.direction ?? '',
    category: job.category ?? '',
    responsibilities: job.responsibilities ?? [],
    mustHave: job.must_have ?? [],
    niceToHave: job.nice_to_have ?? [],
    bonusPoints: job.bonus_points ?? [],
    notes: job.notes ?? [],
    organizations: job.organizations ?? [],
    locations: job.locations ?? [],
    interviewLocations: job.interview_locations ?? [],
    skills: job.skills ?? [],
    externalId: job.external_id ?? '',
    sourceName: job.source_name ?? '',
  };
}
