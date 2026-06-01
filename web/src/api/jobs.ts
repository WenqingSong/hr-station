// ─── Jobs API ──────────────────────────────────────────────────────────────────

import { get } from './client';
import type { Job, JobDetail } from '../types';

/**
 * 获取岗位列表。
 * 调用 GET /api/jobs。失败时错误直接向上传播。
 */
export async function getJobs(): Promise<Job[]> {
  return get<Job[]>('/jobs');
}

/**
 * 获取岗位详情。
 * 调用 GET /api/jobs/{jobId}。失败时错误直接向上传播。
 */
export async function getJobDetail(jobId: string): Promise<JobDetail> {
  return get<JobDetail>(`/jobs/${jobId}`);
}

/**
 * 获取指定岗位详情（由调用方传入 jobId）。
 * 调用 GET /api/jobs/{id}。失败时错误直接向上传播。
 */
export async function getSelectedJob(jobId: string): Promise<JobDetail> {
  return get<JobDetail>(`/jobs/${jobId}`);
}
