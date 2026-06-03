import { useState, useEffect, type FormEvent } from "react";
import { Link, useNavigate } from "react-router";
import { login } from "../../api/auth";
import { setAuthToken, getAuthToken } from "../../api/authStorage";
import type { LoginRequest } from "../../types/auth";

export function LoginPage() {
  const navigate = useNavigate();

  useEffect(() => {
    if (getAuthToken()) navigate("/");
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const [phone, setPhone] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError("");

    if (!phone.trim()) { setError("请输入手机号"); return; }
    if (!/^1[3-9]\d{9}$/.test(phone.trim())) { setError("手机号格式不正确"); return; }
    if (!password) { setError("请输入密码"); return; }

    setLoading(true);
    try {
      const input: LoginRequest = { phone: phone.trim(), password };
      const response = await login(input);
      setAuthToken(response.token);
      navigate("/");
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "登录失败，请稍后重试";
      if (msg.includes("手机号或密码错误") || msg.includes("401")) {
        setError("手机号或密码错误");
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
        <h1 className="mb-6 text-center text-xl font-semibold text-slate-900">登录 offerr鹅</h1>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label htmlFor="login-phone" className="mb-1 block text-sm font-medium text-slate-700">手机号</label>
            <input id="login-phone" type="tel" autoComplete="tel" required value={phone} onChange={(e) => setPhone(e.target.value)} placeholder="请输入手机号" className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm text-slate-900 placeholder-slate-400 focus:border-[#0052D9] focus:outline-none focus:ring-1 focus:ring-[#0052D9]" />
          </div>
          <div>
            <label htmlFor="login-password" className="mb-1 block text-sm font-medium text-slate-700">密码</label>
            <input id="login-password" type="password" autoComplete="current-password" required value={password} onChange={(e) => setPassword(e.target.value)} placeholder="请输入密码" className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm text-slate-900 placeholder-slate-400 focus:border-[#0052D9] focus:outline-none focus:ring-1 focus:ring-[#0052D9]" />
          </div>
          {error && <p className="rounded-md bg-red-50 px-3 py-2 text-sm text-red-600" role="alert">{error}</p>}
          <button type="submit" disabled={loading} className="w-full rounded-md bg-[#0052D9] px-4 py-2.5 text-sm font-medium text-white transition-colors hover:bg-[#0043B3] disabled:cursor-not-allowed disabled:opacity-50">{loading ? "登录中…" : "登录"}</button>
        </form>
        <p className="mt-4 text-center text-sm text-slate-500">还没有账号？ <Link to="/register" className="font-medium text-[#0052D9] hover:underline">注册</Link></p>
      </div>
    </div>
  );
}
