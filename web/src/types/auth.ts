// ─── Auth 类型 ──────────────────────────────────────────────────────────────────
//
// 字段与后端 Auth handler / service / model 契约对齐：
//   - RegisterRequest  ← service.RegisterRequest  { email, password, name }
//   - LoginRequest     ← service.LoginRequest      { email, password }
//   - AuthTokenResponse ← service.AuthResponse     { token, user }
//   - AuthUser         ← service.AuthUser / model.User  { id, email, name }
//   - me 接口 data 直接返回 AuthUser（不包裹在 user 字段中）

/** 用户信息 */
export interface AuthUser {
  id: number;
  email: string;
  name: string;
}

/** 注册请求 */
export interface RegisterRequest {
  email: string;
  password: string;
  name: string;
}

/** 登录请求 */
export interface LoginRequest {
  email: string;
  password: string;
}

/** 注册 / 登录成功后返回的 token + user */
export interface AuthTokenResponse {
  token: string;
  user: AuthUser;
}
