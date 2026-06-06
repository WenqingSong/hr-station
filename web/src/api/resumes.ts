// ─── Resumes API（K8-5-48-C：关系型 DTO） ─────────────────────────────────────

import { get, post, put, request } from './client';
import type {
  Resume,
  ResumeOption,
  ResumeComparison,
  SaveResumePayload,
} from '../types';

// ─── 读取 ─────────────────────────────────────────────────────────────────────

/** 获取所有简历列表 */
export async function getResumes(): Promise<Resume[]> {
  return get<Resume[]>('/resumes');
}

/** 获取当前简历 */
export async function getCurrentResume(): Promise<Resume | null> {
  try {
    return await get<Resume>('/resumes/current');
  } catch {
    return null;
  }
}

/** 获取指定简历详情 */
export async function getResumeById(id: number): Promise<Resume> {
  return get<Resume>(`/resumes/${id}`);
}

/** 获取简历选项列表（选择弹窗用） */
export async function getResumeOptions(): Promise<ResumeOption[]> {
  return get<ResumeOption[]>('/resumes');
}

/** 版本对比（from_id / to_id） */
export async function compareResumes(
  fromId: number,
  toId: number,
): Promise<ResumeComparison> {
  return get<ResumeComparison>(
    `/resumes/compare?from_id=${fromId}&to_id=${toId}`,
  );
}

// ─── 写入 ─────────────────────────────────────────────────────────────────────

/** 创建新简历 */
export async function createResume(
  payload: SaveResumePayload,
): Promise<Resume> {
  return post<Resume>('/resumes', payload);
}

/** 更新简历 */
export async function updateResume(
  id: number,
  payload: SaveResumePayload,
): Promise<Resume> {
  return put<Resume>(`/resumes/${id}`, payload);
}

/** 删除简历 */
export async function deleteResume(id: number): Promise<void> {
  return request<void>(`/resumes/${id}`, { method: 'DELETE' });
}

/** 设为当前简历 */
export async function setCurrentResume(id: number): Promise<Resume> {
  return post<Resume>(`/resumes/${id}/set-current`);
}

// ─── Name check（K8-5-49-B） ────────────────────────────────────────────────

export interface ResumeNameCheckResult {
  available: boolean;
  reason?: string;
}

export async function checkResumeName(name: string, excludeId?: number): Promise<ResumeNameCheckResult> {
  let url = `/resumes/check-name?name=${encodeURIComponent(name)}`;
  if (excludeId && excludeId > 0) url += `&exclude_id=${excludeId}`;
  return get<ResumeNameCheckResult>(url);
}

// ─── @deprecated wrappers（K8-5-48-D 清理） ──────────────────────────────────

/** @deprecated use getResumeById */
export const getResumeDetail = getResumeById;

/** @deprecated use getResumes */
export const getResumeVersions = getResumes;
