// ─── Resumes API ───────────────────────────────────────────────────────────────

import { get, post } from './client';
import type {
  Resume,
  ResumeOption,
  ResumeVersionComparison,
  ResumeVersionsResponse,
  ResumeVersionListItem,
} from '../types';

// ─── P0 读取 ──────────────────────────────────────────────────────────────────

/**
 * 获取简历版本列表。
 * 调用 GET /api/resumes/versions。失败时错误直接向上传播。
 */
export async function getResumes(): Promise<Resume[]> {
  return get<Resume[]>('/resumes/versions');
}

/** 获取版本列表（真实 {versions, total} 结构）。
 * 调用 GET /api/resumes/versions。失败时错误直接向上传播。
 */
export async function getResumeVersions(): Promise<ResumeVersionListItem[]> {
  const response = await get<ResumeVersionsResponse>('/resumes/versions');
  return Array.isArray(response?.versions) ? response.versions : [];
}

/**
 * 获取当前简历。
 * 调用 GET /api/resumes/current。失败时错误直接向上传播。
 */
export async function getCurrentResume(): Promise<Resume> {
  return get<Resume>('/resumes/current');
}

/**
 * 获取简历详情（指定版本）。
 * 调用 GET /api/resumes/versions/{versionNo}。失败时错误直接向上传播。
 */
export async function getResumeDetail(versionNo: number): Promise<Resume> {
  return get<Resume>(`/resumes/versions/${versionNo}`);
}

/**
 * 获取简历选项列表（选择简历弹窗专用）。
 * 调用 GET /api/resumes/versions。失败时错误直接向上传播。
 */
export async function getResumeOptions(): Promise<ResumeOption[]> {
  return get<ResumeOption[]>('/resumes/versions');
}

/**
 * 获取简历版本比较。
 * 调用 GET /api/resumes/versions/compare?from={v1}&to={v2}。失败时错误直接向上传播。
 */
export async function compareResumes(
  baseId: string,
  targetId: string,
): Promise<ResumeVersionComparison> {
  return get<ResumeVersionComparison>(
    `/resumes/versions/compare?from=${encodeURIComponent(baseId)}&to=${encodeURIComponent(targetId)}`,
  );
}

// ─── P1 写入 ───────────────────────────────────────────────────────────────────

/**
 * 创建新简历。
 * 调用 POST /api/resumes。失败时错误直接向上传播。
 */
export async function createResume(input: Partial<Resume>): Promise<Resume> {
  return post<Resume>('/resumes', input);
}

/**
 * 更新简历（后端采用版本迭代模型：POST 生成新版本）。
 * 调用 POST /api/resumes。失败时错误直接向上传播。
 */
export async function updateResume(
  _resumeId: string,
  input: Partial<Resume>,
): Promise<Resume> {
  return post<Resume>('/resumes', input);
}

/**
 * 设为当前简历。
 * 后端设计：current = 最新版本。前端通过再次 POST /resumes 创建新版本即可自动成为 current。
 * 无独立 set-current 接口，调用方应使用 createResume/updateResume 代替。
 */
export async function setCurrentResume(_resumeId: string): Promise<Resume> {
  throw new Error('setCurrentResume: 后端无此接口，请使用 createResume/updateResume 创建新版本');
}
