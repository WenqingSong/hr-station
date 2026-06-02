// ─── Jobs API ──────────────────────────────────────────────────────────────────

import { get } from './client';

// ─── 后端原始类型（对齐 GET /api/jobs 和 GET /api/jobs/{id} 返回结构）─────────

/** 后端 JobResponse 对应的类型 */
export interface JobApiItem {
  id: number;
  title: string;
  department: string;
  city: string;
  direction: string;
  category: string;
  job_type: string;
  recruit_type: string;
  application_type?: string;
  application_project?: string;
  application_subtype?: string;
  slogan?: string;
  summary?: string;
  description: string;
  requirements: string;
  skill_keywords: string[];
  skills?: string[];
  responsibilities?: string[];
  must_have?: string[];
  nice_to_have?: string[];
  bonus_points?: string[];
  notes?: string[];
  organizations?: { code: string; name: string }[];
  locations?: string[];
  interview_locations?: string[];
  status: string;
  external_source?: string;
  external_id?: string;
  source_name?: string;
  source_url?: string;
  data_quality?: string;
  is_featured?: boolean;
  created_at?: string;
  updated_at?: string;
}

/** GET /api/jobs 的 data 字段类型 */
export interface JobListResponse {
  items: JobApiItem[];
  total: number;
  limit: number;
  offset: number;
}

/** 筛选聚合项 */
export interface JobFacetItem {
  value: string;
  label: string;
  count: number;
}

/** 筛选聚合结果 */
export interface JobFacets {
  directions: JobFacetItem[];
  cities: JobFacetItem[];
  recruit_types: JobFacetItem[];
  categories: JobFacetItem[];
  application_projects?: JobFacetItem[];
  application_subtypes?: JobFacetItem[];
  job_types?: JobFacetItem[];
  business_groups?: JobFacetItem[];
  skills?: JobFacetItem[];
}

/** 岗位列表查询参数 */
export interface JobListParams {
  keyword?: string;
  direction?: string;
  city?: string;
  recruit_type?: string;
  category?: string;
  business_group?: string;
  skill?: string;
  application_type?: string;
  application_project?: string;
  application_subtype?: string;
  page?: number;
  page_size?: number;
}

/**
 * 获取岗位列表。
 * 调用 GET /api/jobs。返回 {items, total, limit, offset}。
 * 失败时错误直接向上传播。
 */
export async function getJobs(params?: JobListParams): Promise<JobListResponse> {
  const qs = buildQuery(params);
  return get<JobListResponse>(`/jobs${qs}`);
}

/**
 * 获取筛选聚合。
 * 调用 GET /api/jobs/facets。失败时错误直接向上传播。
 */
export async function getJobFacets(): Promise<JobFacets> {
  return get<JobFacets>('/jobs/facets');
}

/**
 * 获取岗位详情。
 * 调用 GET /api/jobs/{jobId}。失败时错误直接向上传播。
 */
export async function getJobDetail(jobId: string): Promise<JobApiItem> {
  return get<JobApiItem>(`/jobs/${jobId}`);
}

/**
 * 获取指定岗位详情（由调用方传入 jobId）。
 * 调用 GET /api/jobs/{id}。失败时错误直接向上传播。
 */
export async function getSelectedJob(jobId: string): Promise<JobApiItem> {
  return get<JobApiItem>(`/jobs/${jobId}`);
}

/** 将 JobListParams 转为查询字符串 */
function buildQuery(params?: JobListParams): string {
  if (!params) return '';
  const parts: string[] = [];
  if (params.keyword) parts.push(`keyword=${encodeURIComponent(params.keyword)}`);
  if (params.direction) parts.push(`direction=${encodeURIComponent(params.direction)}`);
  if (params.city) parts.push(`city=${encodeURIComponent(params.city)}`);
  if (params.recruit_type) parts.push(`recruit_type=${encodeURIComponent(params.recruit_type)}`);
  if (params.category) parts.push(`category=${encodeURIComponent(params.category)}`);
  if (params.business_group) parts.push(`business_group=${encodeURIComponent(params.business_group)}`);
  if (params.skill) parts.push(`skill=${encodeURIComponent(params.skill)}`);
  if (params.application_type) parts.push(`application_type=${encodeURIComponent(params.application_type)}`);
  if (params.application_project) parts.push(`application_project=${encodeURIComponent(params.application_project)}`);
  if (params.application_subtype) parts.push(`application_subtype=${encodeURIComponent(params.application_subtype)}`);
  if (params.page) parts.push(`page=${params.page}`);
  if (params.page_size) parts.push(`page_size=${params.page_size}`);
  return parts.length ? `?${parts.join('&')}` : '';
}
