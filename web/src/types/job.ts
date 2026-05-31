// ─── 岗位相关类型 ───────────────────────────────────────────────────────────────

/** 岗位类别 */
export type JobCategory =
  | 'backend'
  | 'cloud'
  | 'algorithm'
  | 'client'
  | 'data'
  | 'product'
  | 'game'
  | string;

/** 招聘类型 */
export type JobType = 'campus' | 'internship' | 'early_batch' | string;

/** 岗位基础信息（用于列表） */
export interface Job {
  id: string;
  title: string;
  category: JobCategory;
  type_label: string;
  business_group: string;
  locations: string[];
  tags: string[];
  short_description: string;
  status: string;
}

/** 岗位详情（含完整 JD） */
export interface JobDetail extends Job {
  /** Banner 区域摘要文本 */
  banner_summary: string;
  /** 岗位职责列表 */
  responsibilities: string[];
  /** 岗位要求列表 */
  requirements: string[];
  /** 加分项列表 */
  bonus_points: string[];
  /** 投递说明 */
  delivery_notes?: string;
}

/** 岗位筛选状态 */
export interface JobFilterState {
  keyword: string;
  directions: string[];
  locations: string[];
  job_types: string[];
  categories: string[];
}

/** 当前选中岗位状态 */
export interface SelectedJobState {
  job_id: string;
}
