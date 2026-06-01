// ─── Resumes API ───────────────────────────────────────────────────────────────

import { get, post } from './client';
import type {
  Resume,
  ResumeOption,
  ResumeVersionComparison,
  ResumeVersionsResponse,
  ResumeVersionListItem,
} from '../types';
import {
  mockCurrentResume,
  mockResumeVersions,
  mockResumeOptions,
  mockResumeComparison,
  getMockResumeById,
} from '../mocks';

/** Mock fallback helper — swallows errors and returns fallback data. */
async function withFallback<T>(requestFn: () => Promise<T>, fallbackFn: () => T): Promise<T> {
  try {
    return await requestFn();
  } catch {
    return fallbackFn();
  }
}

// ─── P0 读取 ──────────────────────────────────────────────────────────────────

/**
 * 获取简历版本列表。
 * 优先调用 GET /resumes/versions，失败返回 mockResumeVersions。
 */
export async function getResumes(): Promise<Resume[]> {
  return withFallback(
    () => get<Resume[]>('/resumes/versions'),
    () => mockResumeVersions,
  );
}

/** 获取版本列表（真实 {versions, total} 结构，10-1-R） */
export async function getResumeVersions(): Promise<ResumeVersionListItem[]> {
  try {
    const response = await get<ResumeVersionsResponse>('/resumes/versions');
    return Array.isArray(response?.versions) ? response.versions : [];
  } catch {
    return []; // API 失败返回空数组，让调用方决定 fallback
  }
}

/**
 * 获取当前简历。
 * 优先调用 GET /resumes/current，失败返回 mockCurrentResume。
 */
export async function getCurrentResume(): Promise<Resume> {
  return withFallback(
    () => get<Resume>('/resumes/current'),
    () => mockCurrentResume,
  );
}

/**
 * 获取简历详情。
 * 优先调用 GET /resumes/detail?id={resumeId}，失败尝试 mock。
 */
export async function getResumeDetail(resumeId: string): Promise<Resume> {
  return withFallback(
    () => get<Resume>(`/resumes/detail?id=${encodeURIComponent(resumeId)}`),
    () => getMockResumeById(resumeId) ?? mockCurrentResume,
  );
}

/**
 * 获取简历选项（选择简历弹窗专用）。
 * 直接返回 mockResumeOptions，后续后端补专用接口再接入。
 */
export async function getResumeOptions(): Promise<ResumeOption[]> {
  return withFallback(
    () => get<ResumeOption[]>('/resumes/versions'),
    () => mockResumeOptions,
  );
}

/**
 * 获取简历版本比较。
 * 后端期望 GET /resumes/versions/compare?from=1&to=2（版本号）。
 * 优先调用真实 API，失败返回 mock。
 * 注意：当前前端使用 resumeId（字符串），后端使用 versionNo（整数），
 * 真实调用时需 adapter 层做 ID→版本号转换或后端增强。
 */
export async function compareResumes(
  baseId: string,
  targetId: string,
): Promise<ResumeVersionComparison> {
  return withFallback(
    () =>
      get<ResumeVersionComparison>(
        `/resumes/versions/compare?from=${encodeURIComponent(baseId)}&to=${encodeURIComponent(targetId)}`,
      ),
    () => mockResumeComparison,
  );
}

// ─── P1 写入（Demo 占位，真实保存后续接入） ─────────────────────────────────

/**
 * 创建新简历。
 * 优先 POST /resumes，失败返回 mockCurrentResume。
 * 当前用于前端 Demo 占位，真实保存能力后续阶段再接。
 */
export async function createResume(input: Partial<Resume>): Promise<Resume> {
  return withFallback(
    () => post<Resume>('/resumes', input),
    () => mockCurrentResume,
  );
}

/**
 * 更新简历。
 * 后端采用版本迭代模型：每次 POST /resumes 生成新版本代替修改。
 * 优先 POST /resumes，失败返回 mock。
 */
export async function updateResume(
  _resumeId: string,
  input: Partial<Resume>,
): Promise<Resume> {
  return withFallback(
    () => post<Resume>('/resumes', input),
    () => mockCurrentResume,
  );
}

/**
 * 设为当前简历。
 * 后端设计：current = 最新版本（按 version_no DESC LIMIT 1），无需单独设置接口。
 * 当前保留为 Demo 占位，直接返回 mock。
 */
export async function setCurrentResume(resumeId: string): Promise<Resume> {
  // 后端无 set-current 接口，前端 Demo 场景直接返回 mock。
  // 如需切换当前简历，通过再次 POST /resumes 创建新版本即可自动成为 current。
  return getMockResumeById(resumeId) ?? mockCurrentResume;
}
