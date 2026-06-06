// ─── Redirect 安全校验工具（K8-5-48-C：versionNo → id） ──────────────────────

const DEFAULT_REDIRECT = "/jobs";

/** 已知前端路由前缀白名单 */
const SAFE_PATH_PREFIXES = [
  "/jobs",
  "/recommend",
  "/match",
  "/resumes",
  "/resume-optimization",
  "/login",
];

/** 检查 redirect 是否为站内安全路径 */
export function isSafeRedirect(value: string | null | undefined): boolean {
  if (!value || typeof value !== "string") return false;
  if (!value.startsWith("/") || value.startsWith("//")) return false;
  if (/^(https?:|javascript:)/i.test(value)) return false;
  if (/[\\<>"'\x00-\x1f]/.test(value)) return false;
  return SAFE_PATH_PREFIXES.some((prefix) => value.startsWith(prefix));
}

/** 规范化 redirect，非法时 fallback */
export function normalizeRedirect(
  value: string | null | undefined,
  fallback: string = DEFAULT_REDIRECT,
): string {
  if (!isSafeRedirect(value)) return fallback;

  const v = value!;
  if (v.startsWith("/resumes/edit")) {
    const clean = normalizeResumeEditRedirect(v);
    return clean;
  }

  return v;
}

/**
 * /resumes/edit?id=1  → 保留
 * /resumes/edit?id=0  → /resumes
 * /resumes/edit?id=abc → /resumes
 * /resumes/edit       → 保留（新增）
 */
export function normalizeResumeEditRedirect(value: string): string {
  try {
    const url = new URL(value, "http://localhost");
    const id = url.searchParams.get("id");
    if (id !== null) {
      const parsed = Number(id);
      if (!Number.isInteger(parsed) || parsed <= 0) {
        return "/resumes";
      }
    }
    return value;
  } catch {
    return "/resumes";
  }
}

/**
 * 为 RequireAuth 生成安全的 returnTo 值。
 */
export function normalizeReturnTo(
  pathname: string,
  search?: string,
  fallback: string = DEFAULT_REDIRECT,
): string {
  if (!pathname.startsWith("/") || pathname.startsWith("//")) return fallback;
  if (/[\\<>"'&\x00-\x1f]/.test(pathname)) return fallback;

  const full = pathname + (search || "");
  if (!isSafeRedirect(pathname)) return fallback;

  if (pathname.startsWith("/resumes/edit")) {
    return normalizeResumeEditRedirect(full);
  }

  return full;
}

/** 校验 resumeId 参数是否为合法正整数 */
export function isValidResumeId(raw: string | null): boolean {
  if (raw === null) return true; // 无参数 = 合法（新增简历）
  const parsed = Number(raw);
  return Number.isInteger(parsed) && parsed > 0;
}

/** @deprecated use isValidResumeId */
export const isValidVersionNo = isValidResumeId;
