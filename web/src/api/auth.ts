// ─── Auth API ───────────────────────────────────────────────────────────────────
//
// 职责：调用后端 Auth 接口（register / login / me），返回类型化数据。
//
// 本模块严格遵守以下边界：
//   - 不写 token storage（localStorage / sessionStorage）
//   - 不自动注入 Authorization（留到 8-3）
//   - 不使用 fallback / mock（Auth 失败必须让调用方感知）
//   - getMe 接收 token 参数，手动传 Authorization header

import { get, post } from './client';
import type {
  AuthTokenResponse,
  AuthUser,
  LoginRequest,
  RegisterRequest,
} from '../types/auth';

/**
 * 注册新用户。
 * POST /auth/register → { token, user }
 * 失败时错误会向上传播（不 fallback）。
 */
export function register(input: RegisterRequest): Promise<AuthTokenResponse> {
  return post<AuthTokenResponse, RegisterRequest>('/auth/register', input);
}

/**
 * 用户登录。
 * POST /auth/login → { token, user }
 * 失败时错误会向上传播（不 fallback）。
 */
export function login(input: LoginRequest): Promise<AuthTokenResponse> {
  return post<AuthTokenResponse, LoginRequest>('/auth/login', input);
}

/**
 * 获取当前用户信息。
 * GET /auth/me → AuthUser
 * 需要传入有效 token，由调用方通过 Authorization header 传递。
 * 失败时错误会向上传播（不 fallback）。
 */
export function getMe(token: string): Promise<AuthUser> {
  return get<AuthUser>('/auth/me', {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
}
