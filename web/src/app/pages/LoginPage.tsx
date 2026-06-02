import { useState, useEffect, type FormEvent } from "react";
import { Link, useNavigate } from "react-router";
import { login } from "../../api/auth";
import { setAuthToken, getAuthToken } from "../../api/authStorage";
import type { LoginRequest } from "../../types/auth";

export function LoginPage() {
  const navigate = useNavigate();

  // 已登录用户直接跳转首页
  useEffect(() => {
    if (getAuthToken()) navigate("/");
  }, []); // eslint-disable-line react-hooks/exhaustive-deps
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError("");

    if (!email.trim()) {
      setError("请输入邮箱地址");
      return;
    }
    if (!password) {
      setError("请输入密码");
      return;
    }

    setLoading(true);
    try {
      const input: LoginRequest = { email: email.trim(), password };
      const response = await login(input);
      setAuthToken(response.token);
      navigate("/");
    } catch (err: unknown) {
      const msg =
        err instanceof Error ? err.message : "登录失败，请稍后重试";
      // 产品化错误文案，不暴露 401 / Unauthorized / API error
      if (msg.includes("invalid email or password") || msg.includes("401")) {
        setError("邮箱或密码不正确，请重新输入");
      } else {
        setError("登录失败，请稍后重试");
      }
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="flex min-h-[calc(100vh-3.5rem)] items-center justify-center px-4">
      <div className="w-full max-w-sm rounded-lg border border-gray-200 bg-white p-8 shadow-sm">
        <h1 className="mb-6 text-center text-xl font-semibold text-slate-900">
          登录 Offer 捕手
        </h1>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label
              htmlFor="login-email"
              className="mb-1 block text-sm font-medium text-slate-700"
            >
              邮箱
            </label>
            <input
              id="login-email"
              type="email"
              autoComplete="email"
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="your@email.com"
              className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm text-slate-900 placeholder-slate-400 focus:border-[#0052D9] focus:outline-none focus:ring-1 focus:ring-[#0052D9]"
            />
          </div>

          <div>
            <label
              htmlFor="login-password"
              className="mb-1 block text-sm font-medium text-slate-700"
            >
              密码
            </label>
            <input
              id="login-password"
              type="password"
              autoComplete="current-password"
              required
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="请输入密码"
              className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm text-slate-900 placeholder-slate-400 focus:border-[#0052D9] focus:outline-none focus:ring-1 focus:ring-[#0052D9]"
            />
          </div>

          {error && (
            <p className="rounded-md bg-red-50 px-3 py-2 text-sm text-red-600" role="alert">
              {error}
            </p>
          )}

          <button
            type="submit"
            disabled={loading}
            className="w-full rounded-md bg-[#0052D9] px-4 py-2.5 text-sm font-medium text-white transition-colors hover:bg-[#0043B3] disabled:cursor-not-allowed disabled:opacity-50"
          >
            {loading ? "登录中…" : "登录"}
          </button>
        </form>

        <p className="mt-4 text-center text-sm text-slate-500">
          还没有账号？{" "}
          <Link to="/register" className="font-medium text-[#0052D9] hover:underline">
            注册
          </Link>
        </p>
      </div>
    </div>
  );
}
