// ─── Redirect 安全校验工具（K8-5-46-B） ────────────────────────────────────────

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
  // 必须以 / 开头，不能以 // 开头
  if (!value.startsWith("/") || value.startsWith("//")) return false;
  // 禁止 http/https/javascript
  if (/^(https?:|javascript:)/i.test(value)) return false;
  // 禁止反斜杠和控制字符
  if (/[\\<>"'\x00-\x1f]/.test(value)) return false;
  // 路径前缀在白名单内
  return SAFE_PATH_PREFIXES.some((prefix) => value.startsWith(prefix));
}

/** 规范化 redirect，非法时 fallback */
export function normalizeRedirect(
  value: string | null | undefined,
  fallback: string = DEFAULT_REDIRECT,
): string {
  if (!isSafeRedirect(value)) return fallback;

  // 特殊处理：/resumes/edit?versionNo=0 等非法参数替换为 /resumes
  const v = value!;
  if (v.startsWith("/resumes/edit")) {
    const clean = normalizeResumeEditRedirect(v);
    return clean;
  }

  return v;
}

/**
 * /resumes/edit?versionNo=1  → 保留
 * /resumes/edit?versionNo=0  → /resumes
 * /resumes/edit?versionNo=abc → /resumes
 * /resumes/edit              → 保留（新增）
 */
export function normalizeResumeEditRedirect(value: string): string {
  try {
    const url = new URL(value, "http://localhost");
    const versionNo = url.searchParams.get("versionNo");
    if (versionNo !== null) {
      const parsed = Number(versionNo);
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
 * 从 pathname + search 构造，非法时 fallback。
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

  // /resumes/edit 特殊处理 versionNo
  if (pathname.startsWith("/resumes/edit")) {
    return normalizeResumeEditRedirect(full);
  }

  return full;
}

/** 校验 versionNo 参数是否为合法正整数 */
export function isValidVersionNo(raw: string | null): boolean {
  if (raw === null) return true; // 无参数 = 合法（新增简历）
  const parsed = Number(raw);
  return Number.isInteger(parsed) && parsed > 0;
}
