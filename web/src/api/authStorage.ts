// ─── Auth Token Storage ────────────────────────────────────────────────────────
//
// 职责：读写 localStorage 中的 auth token。
//
// 边界：
//   - 只存 token 字符串，不存 user 大对象
//   - SSR / build 安全（检查 typeof window）
//   - 不使用 sessionStorage
//   - 单 key 策略

const AUTH_TOKEN_KEY = "offer_catcher_auth_token";

function storageAvailable(): boolean {
  try {
    return typeof window !== "undefined" && Boolean(window.localStorage);
  } catch {
    return false;
  }
}

/** 获取已存储的 token，不存在返回 null */
export function getAuthToken(): string | null {
  if (!storageAvailable()) return null;
  const value = window.localStorage.getItem(AUTH_TOKEN_KEY);
  return value && value.length > 0 ? value : null;
}

/** 存储 token */
export function setAuthToken(token: string): void {
  if (!storageAvailable()) return;
  window.localStorage.setItem(AUTH_TOKEN_KEY, token);
}

/** 清除 token（登录过期 / 退出登录 / 401 时调用） */
export function clearAuthToken(): void {
  if (!storageAvailable()) return;
  window.localStorage.removeItem(AUTH_TOKEN_KEY);
}
