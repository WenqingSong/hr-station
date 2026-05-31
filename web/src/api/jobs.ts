// ─── Jobs API ──────────────────────────────────────────────────────────────────

import { get } from './client';
import type { Job, JobDetail } from '../types';
import { mockJobs, mockJobDetails, mockSelectedJob, getMockJobById } from '../mocks';

/** Mock fallback helper — swallows errors and returns fallback data. */
async function withFallback<T>(requestFn: () => Promise<T>, fallbackFn: () => T): Promise<T> {
  try {
    return await requestFn();
  } catch {
    return fallbackFn();
  }
}

/**
 * 获取岗位列表。
 * 优先调用 GET /jobs，失败返回 mockJobs。
 */
export async function getJobs(): Promise<Job[]> {
  return withFallback(
    () => get<Job[]>('/jobs'),
    () => mockJobs,
  );
}

/**
 * 获取岗位详情。
 * 优先调用 GET /jobs/{jobId}，失败尝试 mock，兜底返回 mockSelectedJob。
 */
export async function getJobDetail(jobId: string): Promise<JobDetail> {
  return withFallback(
    () => get<JobDetail>(`/jobs/${jobId}`),
    () => getMockJobById(jobId) ?? mockSelectedJob,
  );
}

/**
 * 获取 Demo 默认岗位。
 * 直接返回 mockSelectedJob。
 */
export async function getSelectedJob(): Promise<JobDetail> {
  return withFallback(
    () => get<JobDetail>(`/jobs/${mockSelectedJob.id}`),
    () => mockSelectedJob,
  );
}
