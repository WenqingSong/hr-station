// ─── Jobs API ──────────────────────────────────────────────────────────────────

import { get } from './client';

// ─── 后端原始类型 ─────────────────────────────────────────────────────────────

export interface JobApiItem {
  id: number; title: string; department: string; city: string;
  direction: string; category: string; job_type: string; recruit_type: string;
  application_type?: string; application_project?: string; application_subtype?: string;
  slogan?: string; summary?: string; description: string; requirements: string;
  skill_keywords: string[]; skills?: string[];
  responsibilities?: string[]; must_have?: string[]; nice_to_have?: string[];
  bonus_points?: string[]; notes?: string[];
  delivery_notes?: string;
  organizations?: { code: string; name: string }[];
  locations?: string[]; interview_locations?: string[];
  status: string; external_source?: string; external_id?: string;
  source_name?: string; source_url?: string; data_quality?: string;
  is_featured?: boolean; created_at?: string; updated_at?: string;
}

export interface JobListResponse { items: JobApiItem[]; total: number; limit: number; offset: number }

export interface JobFacetItem { value: string; label: string; count: number }

// ─── K8-4-4: products facets 新结构 ────────────────────────────────────────────

export interface JobFacetGroup {
  code: string; name: string; display_order: number;
  default_expanded: boolean; visible: boolean; max_visible_options: number;
  options: JobFacetOption[];
}

export interface JobFacetOption {
  code: string; label: string; value: string; count: number;
  display_order: number; visible: boolean;
  children?: JobFacetOption[] | null;
}

export interface JobFacets {
  // K8-4-4 新字段
  groups?: JobFacetGroup[];
  // 旧兼容字段
  directions: JobFacetItem[]; cities: JobFacetItem[];
  recruit_types: JobFacetItem[]; categories: JobFacetItem[];
  application_projects?: JobFacetItem[]; application_subtypes?: JobFacetItem[];
  job_types?: JobFacetItem[]; business_groups?: JobFacetItem[]; skills?: JobFacetItem[];
}

export interface JobListParams {
  keyword?: string; direction?: string; city?: string;
  recruit_type?: string; category?: string; business_group?: string;
  skill?: string; job_category?: string;
  application_type?: string; application_project?: string; application_subtype?: string;
  page?: number; page_size?: number;
}

export async function getJobs(params?: JobListParams): Promise<JobListResponse> {
  return get<JobListResponse>(`/jobs${buildQuery(params)}`);
}

export async function getJobFacets(): Promise<JobFacets> {
  return get<JobFacets>('/jobs/facets');
}

export async function getJobDetail(jobId: string): Promise<JobApiItem> {
  return get<JobApiItem>(`/jobs/${jobId}`);
}

export async function getSelectedJob(jobId: string): Promise<JobApiItem> {
  return get<JobApiItem>(`/jobs/${jobId}`);
}

function buildQuery(params?: JobListParams): string {
  if (!params) return '';
  const parts: string[] = [];
  const push = (k: string, v: string | number | undefined) => {
    if (v !== undefined && v !== '') parts.push(`${k}=${encodeURIComponent(String(v))}`);
  };
  push('keyword', params.keyword);
  push('direction', params.direction);
  push('city', params.city);
  push('recruit_type', params.recruit_type);
  push('category', params.category);
  push('business_group', params.business_group);
  push('skill', params.skill);
  push('job_category', params.job_category);
  push('application_type', params.application_type);
  push('application_project', params.application_project);
  push('application_subtype', params.application_subtype);
  push('page', params.page);
  push('page_size', params.page_size);
  return parts.length ? `?${parts.join('&')}` : '';
}
