import { useState, useEffect, type FormEvent } from "react";
import { Link, useNavigate, useLocation } from "react-router";
import { register } from "../../api/auth";
import { setAuthToken, getAuthToken } from "../../api/authStorage";
import type { RegisterRequest } from "../../types/auth";
import { Input } from "../components/ui/input";
import { PasswordInput } from "../components/ui/password-input";

/** 安全读取 redirect 参数，只允许站内路径 */
function getSafeRedirect(search: string): string {
  const params = new URLSearchParams(search);
  const redirect = params.get("redirect");
  if (!redirect) return "/jobs";
  if (!redirect.startsWith("/")) return "/jobs";
  if (redirect.startsWith("//")) return "/jobs";
  if (redirect.includes("://")) return "/jobs";
  return redirect;
}

export function RegisterPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const safeRedirect = getSafeRedirect(location.search);

  useEffect(() => {
    if (getAuthToken()) navigate(safeRedirect);
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const [phone, setPhone] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError("");

    if (!phone.trim()) { setError("请输入手机号"); return; }
    if (!/^1[3-9]\d{9}$/.test(phone.trim())) { setError("手机号格式不正确"); return; }
    if (!password) { setError("请输入密码"); return; }
    if (password.length < 6) { setError("密码长度不能少于 6 位"); return; }
    if (password !== confirmPassword) { setError("两次输入的密码不一致"); return; }

    setLoading(true);
    try {
      const input: RegisterRequest = { phone: phone.trim(), password };
      const response = await register(input);
      setAuthToken(response.token);
      navigate(safeRedirect);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "注册失败，请稍后重试";
      if (msg.includes("已被注册")) {
        setError("该手机号已被注册，请直接登录");
      } else {
        setError("注册失败，请稍后重试");
      }
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="flex min-h-[calc(100vh-3.5rem)] items-center justify-center px-4">
      <div className="w-full max-w-sm rounded-lg border border-gray-200 bg-white p-8 shadow-sm">
        <h1 className="mb-6 text-center text-xl font-semibold text-slate-900">注册 Offerr鹅</h1>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label htmlFor="register-phone" className="mb-1 block text-sm font-medium text-slate-700">手机号</label>
            <Input id="register-phone" autoComplete="tel" value={phone} onChange={(e) => setPhone(e.target.value)} placeholder="请输入手机号" />
          </div>
          <div>
            <label htmlFor="register-password" className="mb-1 block text-sm font-medium text-slate-700">密码</label>
            <PasswordInput id="register-password" autoComplete="new-password" value={password} onChange={(e) => setPassword(e.target.value)} placeholder="至少 6 位" />
          </div>
          <div>
            <label htmlFor="register-confirm-password" className="mb-1 block text-sm font-medium text-slate-700">确认密码</label>
            <PasswordInput id="register-confirm-password" autoComplete="new-password" value={confirmPassword} onChange={(e) => setConfirmPassword(e.target.value)} placeholder="请再次输入密码" />
          </div>
          {error && <p className="rounded-md bg-red-50 px-3 py-2 text-sm text-red-600" role="alert">{error}</p>}
          <button type="submit" disabled={loading} className="w-full rounded-md bg-[#0052D9] px-4 py-2.5 text-sm font-medium text-white transition-colors hover:bg-[#0043B3] disabled:cursor-not-allowed disabled:opacity-50">{loading ? "注册中…" : "注册"}</button>
        </form>
        <p className="mt-4 text-center text-sm text-slate-500">已有账号？ <Link to="/login" className="font-medium text-[#0052D9] hover:underline">登录</Link></p>
      </div>
    </div>
  );
}
