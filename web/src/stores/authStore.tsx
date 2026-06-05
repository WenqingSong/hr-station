// ─── Auth Store — 统一登录态状态管理（React Context + useReducer） ────────────
//
// 职责：
//   - 提供全局登录态单一来源
//   - App 启动时调用 /auth/me 做 auth bootstrap
//   - 监听 offerr-auth-expired 事件（client.ts 401 触发），全局清理
//   - 所有页面/组件统一从此读取登录态
//
// K8-5-36-C：统一登录态 Auth Bootstrap、401 全局清理与统一登录弹窗链路修复
//
// 使用：
//   - App.tsx 用 <AuthProvider> 包裹整个应用
//   - 组件内用 const { isAuthenticated, user, ... } = useAuth()

import {
  createContext,
  useContext,
  useReducer,
  useEffect,
  useCallback,
  type ReactNode,
} from "react";
import { getAuthToken, setAuthToken, clearAuthToken } from "../api/authStorage";
import { getMe } from "../api/auth";
import type { AuthUser } from "../types/auth";

// ─── State ──────────────────────────────────────────────────────────────────────

export type AuthState = {
  token: string | null;
  user: AuthUser | null;
  isAuthenticated: boolean;
  isAuthReady: boolean;
  isBootstrapping: boolean;
};

type AuthAction =
  | { type: "SET_AUTHENTICATED"; token: string; user: AuthUser | null }
  | { type: "SET_AUTH_READY" }
  | { type: "SET_BOOTSTRAPPING" }
  | { type: "CLEAR_AUTH" };

function authReducer(state: AuthState, action: AuthAction): AuthState {
  switch (action.type) {
    case "SET_BOOTSTRAPPING":
      return { ...state, isBootstrapping: true };
    case "SET_AUTHENTICATED":
      return {
        ...state,
        token: action.token,
        user: action.user,
        isAuthenticated: true,
        isAuthReady: true,
        isBootstrapping: false,
      };
    case "SET_AUTH_READY":
      return { ...state, isAuthReady: true, isBootstrapping: false };
    case "CLEAR_AUTH":
      // 幂等：如果已经未认证，no-op
      if (!state.isAuthenticated && state.isAuthReady && !state.token && !state.user) {
        return state;
      }
      return {
        ...state,
        token: null,
        user: null,
        isAuthenticated: false,
        isAuthReady: true,
        isBootstrapping: false,
      };
    default:
      return state;
  }
}

const initialState: AuthState = {
  token: getAuthToken(),
  user: null,
  isAuthenticated: false,
  isAuthReady: false,
  isBootstrapping: false,
};

// ─── Context ────────────────────────────────────────────────────────────────────

export type AuthContextValue = AuthState & {
  /** 登录/注册成功后调用：写入 token 并标记已认证 */
  setAuthenticated: (token: string, user?: AuthUser | null) => void;
  /** 退出登录 / token 失效时调用：清空 token + user + 标记未认证 */
  clearAuth: () => void;
  /** App 启动时执行一次：校验 localStorage 中的 token 是否有效 */
  bootstrapAuth: () => Promise<void>;
};

const AuthContext = createContext<AuthContextValue | null>(null);

// ─── Provider ───────────────────────────────────────────────────────────────────

export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, dispatch] = useReducer(authReducer, initialState);

  const setAuthenticated = useCallback(
    (token: string, user: AuthUser | null = null) => {
      setAuthToken(token);
      dispatch({ type: "SET_AUTHENTICATED", token, user });
    },
    [],
  );

  const clearAuth = useCallback(() => {
    clearAuthToken();
    dispatch({ type: "CLEAR_AUTH" });
  }, []);

  const bootstrapAuth = useCallback(async () => {
    dispatch({ type: "SET_BOOTSTRAPPING" });
    const token = getAuthToken();
    if (!token) {
      dispatch({ type: "SET_AUTH_READY" });
      return;
    }
    try {
      const user = await getMe(token);
      dispatch({ type: "SET_AUTHENTICATED", token, user });
    } catch (_err) {
      // /api/auth/me 失败（token 过期/无效/网络错误）→ 清理
      clearAuthToken();
      dispatch({ type: "CLEAR_AUTH" });
    }
  }, []);

  // ── 监听全局 auth-expired 事件（client.ts 401 触发） ──
  useEffect(() => {
    function handleAuthExpired() {
      clearAuth();
    }
    window.addEventListener("offerr-auth-expired", handleAuthExpired);
    return () => window.removeEventListener("offerr-auth-expired", handleAuthExpired);
  }, [clearAuth]);

  const value: AuthContextValue = {
    ...state,
    setAuthenticated,
    clearAuth,
    bootstrapAuth,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

// ─── Hook ───────────────────────────────────────────────────────────────────────

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return ctx;
}
