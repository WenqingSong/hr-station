export interface AuthUser {
  id: number;
  phone: string;
  name: string;
}
export interface LoginRequest {
  phone: string;
  password: string;
}
export interface RegisterRequest {
  phone: string;
  password: string;
}
export interface AuthTokenResponse {
  token: string;
  user: AuthUser;
}
