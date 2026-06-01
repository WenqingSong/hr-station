import { useState, type FormEvent } from "react";
import { Link, useNavigate } from "react-router";
import { register } from "../../api/auth";
import type { RegisterRequest } from "../../types/auth";

export function RegisterPage() {
  const navigate = useNavigate();
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError("");

    if (!name.trim()) {
      setError("请输入姓名");
      return;
    }
    if (!email.trim()) {
      setError("请输入邮箱地址");
      return;
    }
    if (!password) {
      setError("请输入密码");
      return;
    }
    if (password.length < 6) {
      setError("密码长度不能少于 6 位");
      return;
    }
    if (password !== confirmPassword) {
      setError("两次输入的密码不一致");
      return;
    }

    setLoading(true);
    try {
      const input: RegisterRequest = {
        name: name.trim(),
        email: email.trim(),
        password,
      };
      // 调用 register，但不持久化 token（留到 8-3）
      await register(input);
      // 注册成功后跳登录页（本阶段不保存 token）
      navigate("/login");
    } catch (err: unknown) {
      const msg =
        err instanceof Error ? err.message : "注册失败，请稍后重试";
      // 产品化错误文案
      if (msg.includes("email already registered")) {
        setError("该邮箱已被注册，请直接登录");
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
        <h1 className="mb-6 text-center text-xl font-semibold text-slate-900">
          注册 Offer 捕手
        </h1>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label
              htmlFor="register-name"
              className="mb-1 block text-sm font-medium text-slate-700"
            >
              姓名
            </label>
            <input
              id="register-name"
              type="text"
              autoComplete="name"
              required
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="你的姓名"
              className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm text-slate-900 placeholder-slate-400 focus:border-[#0052D9] focus:outline-none focus:ring-1 focus:ring-[#0052D9]"
            />
          </div>

          <div>
            <label
              htmlFor="register-email"
              className="mb-1 block text-sm font-medium text-slate-700"
            >
              邮箱
            </label>
            <input
              id="register-email"
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
              htmlFor="register-password"
              className="mb-1 block text-sm font-medium text-slate-700"
            >
              密码
            </label>
            <input
              id="register-password"
              type="password"
              autoComplete="new-password"
              required
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="至少 6 位"
              className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm text-slate-900 placeholder-slate-400 focus:border-[#0052D9] focus:outline-none focus:ring-1 focus:ring-[#0052D9]"
            />
          </div>

          <div>
            <label
              htmlFor="register-confirm-password"
              className="mb-1 block text-sm font-medium text-slate-700"
            >
              确认密码
            </label>
            <input
              id="register-confirm-password"
              type="password"
              autoComplete="new-password"
              required
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              placeholder="请再次输入密码"
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
            {loading ? "注册中…" : "注册"}
          </button>
        </form>

        <p className="mt-4 text-center text-sm text-slate-500">
          已有账号？{" "}
          <Link to="/login" className="font-medium text-[#0052D9] hover:underline">
            登录
          </Link>
        </p>
      </div>
    </div>
  );
}
