import { useState, useEffect, type FormEvent } from "react";
import { Loader2 } from "lucide-react";
import { login, register } from "../../../api/auth";
import { useAuth } from "../../../stores/authStore";
import type { LoginRequest, RegisterRequest } from "../../../types/auth";
import { Dialog, DialogPortal, DialogOverlay } from "../ui/dialog";
import { Input } from "../ui/input";
import { PasswordInput } from "../ui/password-input";

type AuthDialogMode = "login" | "register";

export type AuthDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  defaultMode?: AuthDialogMode;
  onLoginSuccess?: () => void;
};

/**
 * K8-5-13-E：AuthDialog 玻璃模糊视觉精修 + 响应式
 *
 * 设计说明：
 * - 自定义 DialogOverlay（backdrop-blur 玻璃模糊）
 * - 自定义 content div（替换 DialogContent，不依赖全局 Dialog 样式）
 * - max-h + overflow-y-auto 处理小屏高度溢出
 * - 桌面/平板/手机三档响应式 padding
 */
export function AuthDialog({
  open,
  onOpenChange,
  defaultMode = "login",
  onLoginSuccess,
}: AuthDialogProps) {
  const [mode, setMode] = useState<AuthDialogMode>(defaultMode);
  const { setAuthenticated } = useAuth();

  // ── Login fields ──
  const [loginPhone, setLoginPhone] = useState("");
  const [loginPassword, setLoginPassword] = useState("");
  const [loginLoading, setLoginLoading] = useState(false);
  const [loginError, setLoginError] = useState("");

  // ── Demo reviewer one-click login ──
  const [demoLoading, setDemoLoading] = useState(false);

  // ── Register fields ──
  const [regPhone, setRegPhone] = useState("");
  const [regPassword, setRegPassword] = useState("");
  const [regConfirm, setRegConfirm] = useState("");
  const [regLoading, setRegLoading] = useState(false);
  const [regError, setRegError] = useState("");

  // 弹窗打开时重置为 defaultMode
  useEffect(() => {
    if (open) {
      setMode(defaultMode);
      setLoginError("");
      setRegError("");
    }
  }, [open, defaultMode]);

  // ── 登录 ──
  async function handleLogin(e: FormEvent) {
    e.preventDefault();
    setLoginError("");

    const phone = loginPhone.trim();
    if (!phone) { setLoginError("请输入手机号"); return; }
    if (!/^1[3-9]\d{9}$/.test(phone)) { setLoginError("手机号格式不正确"); return; }
    if (!loginPassword) { setLoginError("请输入密码"); return; }

    setLoginLoading(true);
    try {
      const input: LoginRequest = { phone, password: loginPassword };
      const res = await login(input);
      setAuthenticated(res.token, res.user);
      onLoginSuccess?.();
      onOpenChange(false);
    } catch (err: unknown) {
      // K15-P0-R2: 展示后端真实错误消息，而非只展示泛化提示
      const msg = err instanceof Error ? err.message : "登录失败，请稍后重试";
      if (msg.includes("401") || msg.includes("unauthorized")) { setLoginError("手机号或密码错误"); }
      else if (msg.includes("手机号") || msg.includes("密码") || msg.includes("invalid") || msg.includes("格式") || msg.includes("长度")) { setLoginError(msg); }
      else if (msg.startsWith("HTTP ")) { setLoginError("服务异常，请稍后重试"); }
      else { setLoginError(msg); }
    } finally {
      setLoginLoading(false);
    }
  }

  // ── Demo reviewer one-click login ──
  async function handleDemoLogin() {
    setLoginError("");
    setDemoLoading(true);
    try {
      const res = await login({ phone: "13900139000", password: "123456" });
      setAuthenticated(res.token, res.user);
      onLoginSuccess?.();
      onOpenChange(false);
    } catch (err: unknown) {
      setLoginError("示范账号登录失败，请稍后重试");
    } finally {
      setDemoLoading(false);
    }
  }

  // ── 注册 ──
  async function handleRegister(e: FormEvent) {
    e.preventDefault();
    setRegError("");

    const phone = regPhone.trim();
    if (!phone) { setRegError("请输入手机号"); return; }
    if (!/^1[3-9]\d{9}$/.test(phone)) { setRegError("手机号格式不正确"); return; }
    if (!regPassword) { setRegError("请输入密码"); return; }
    if (regPassword.length < 6) { setRegError("密码长度不能少于 6 位"); return; }
    if (regPassword !== regConfirm) { setRegError("两次输入的密码不一致"); return; }

    setRegLoading(true);
    try {
      const input: RegisterRequest = { phone, password: regPassword };
      const res = await register(input);
      setAuthenticated(res.token, res.user);
      onLoginSuccess?.();
      onOpenChange(false);
    } catch (err: unknown) {
      // K15-P0-R2: 展示后端真实错误消息，保留前端已知语义的友好化映射
      const msg = err instanceof Error ? err.message : "注册失败，请稍后重试";
      if (msg.includes("已被注册") || msg.includes("already registered") || msg.includes("duplicate")) {
        setRegError("该手机号已被注册，请直接登录");
      } else if (msg.includes("手机号") || msg.includes("phone") && msg.includes("invalid")) {
        setRegError(msg); // 展示后端的手机号校验详情
      } else if (msg.includes("密码") || msg.includes("password") || msg.includes("长度") || msg.includes("length")) {
        setRegError(msg); // 展示后端的密码校验详情
      } else if (msg.startsWith("HTTP ")) {
        setRegError("注册服务异常，请稍后重试");
      } else {
        setRegError(msg); // 默认透传后端消息
      }
    } finally {
      setRegLoading(false);
    }
  }

  // 切换模式：保留手机号，清空密码/error
  function switchMode(to: AuthDialogMode) {
    setMode(to);
    setLoginError("");
    setRegError("");
    if (to === "register") {
      setLoginPassword("");
      setRegPhone(loginPhone);
      setRegPassword("");
      setRegConfirm("");
    } else {
      setRegPassword("");
      setRegConfirm("");
      setLoginPhone(regPhone);
      setLoginPassword("");
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogPortal>
        {/* K8-5-17-A：柔和渐暗 + 玻璃模糊 Overlay */}
        <DialogOverlay
          className="fixed inset-0 z-50 bg-slate-950/35 backdrop-blur-md data-[state=open]:animate-auth-overlay-in data-[state=closed]:animate-auth-overlay-out"
        />

        {/* K8-5-17-A：镜头聚焦 Content — 动画外壳（overflow-hidden 防滚动条闪烁） */}
        <div
          className="fixed top-[50%] left-[50%] z-50 w-[min(92vw,420px)] max-h-[min(92vh,720px)] translate-x-[-50%] translate-y-[-50%] rounded-3xl border-0 bg-white/88 backdrop-blur-xl ring-1 ring-white/60 shadow-[0_24px_80px_rgba(15,23,42,0.22)] overflow-hidden data-[state=open]:animate-auth-dialog-in data-[state=closed]:animate-auth-dialog-out"
          style={{ backgroundColor: "rgba(255,255,255,0.88)" }}
        >
          {/* 滚动容器：auth-dialog-scroll 隐藏滚动条视觉 + 保留滚动能力 */}
          <div className="auth-dialog-scroll max-h-[min(92vh,720px)] overflow-y-auto overflow-x-hidden overscroll-contain">
          {/* ── Close button (top-right) ── */}
          <button
            type="button"
            onClick={() => onOpenChange(false)}
            className="absolute top-4 right-4 inline-flex h-8 w-8 items-center justify-center rounded-full text-slate-400 transition-all duration-200 hover:scale-105 hover:bg-slate-100 hover:text-slate-600 active:scale-95"
            aria-label="关闭"
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
          </button>

          {/* ── Header ── */}
          <div className="px-5 pt-6 pb-4 sm:px-7 sm:pt-7">
            <p
              className="text-[22px] font-extrabold tracking-[-0.03em] text-slate-950 mb-1"
              style={{
                fontFamily:
                  '"Alimama ShuHeiTi", "MiSans", "HarmonyOS Sans SC", "Source Han Sans SC", "Noto Sans SC", "Inter", system-ui, sans-serif',
              }}
            >
              Offer捕手
            </p>
            <p className="text-[13px] font-medium text-slate-500 mb-3">
              学生求职匹配与简历优化智能体
            </p>
            <h2 className="text-lg font-semibold text-slate-950">
              {mode === "login" ? "登录" : "注册"}
            </h2>
            <p className="mt-1 text-sm leading-6 text-slate-500">
              {mode === "login" ? "欢迎回来，登录你的账号" : "创建一个账号，开始使用 Offer捕手"}
            </p>
          </div>

          {/* ── Body ── */}
          <div key={mode} className="animate-auth-panel-in px-5 pb-6 sm:px-7 sm:pb-7">
            {mode === "login" ? (
              <form onSubmit={handleLogin} className="space-y-4">
                <div>
                  <label htmlFor="auth-login-phone" className="mb-1.5 block text-sm font-medium text-slate-700">
                    手机号
                  </label>
                  <Input
                    id="auth-login-phone"
                    value={loginPhone}
                    onChange={(e) => setLoginPhone(e.target.value)}
                    placeholder="请输入手机号"
                    autoComplete="tel"
                  />
                </div>
                <div>
                  <label htmlFor="auth-login-password" className="mb-1.5 block text-sm font-medium text-slate-700">
                    密码
                  </label>
                  <PasswordInput
                    id="auth-login-password"
                    value={loginPassword}
                    onChange={(e) => setLoginPassword(e.target.value)}
                    placeholder="请输入密码"
                    autoComplete="current-password"
                  />
                </div>

                {loginError && (
                  <p className="animate-auth-error-in rounded-xl bg-red-50/90 px-4 py-3 text-sm text-red-600 ring-1 ring-red-100" role="alert">
                    {loginError}
                  </p>
                )}

                <button
                  type="submit"
                  disabled={loginLoading || demoLoading}
                  className="flex h-11 w-full items-center justify-center gap-2 rounded-xl bg-[#0052D9] text-sm font-medium text-white shadow-sm transition-colors hover:bg-[#0047BD] disabled:cursor-not-allowed disabled:opacity-50"
                >
                  {loginLoading && <Loader2 className="h-4 w-4 animate-spin" />}
                  {loginLoading ? "登录中…" : "登录"}
                </button>

                {/* Demo reviewer one-click experience */}
                <div className="relative">
                  <div className="absolute inset-0 flex items-center" aria-hidden="true">
                    <div className="w-full border-t border-gray-200" />
                  </div>
                  <div className="relative flex justify-center">
                    <span className="bg-white/88 px-3 text-xs text-gray-400">或</span>
                  </div>
                </div>

                <button
                  type="button"
                  onClick={handleDemoLogin}
                  disabled={loginLoading || demoLoading}
                  className="flex h-11 w-full items-center justify-center gap-2 rounded-xl border-2 border-[#0052D9] bg-white text-sm font-medium text-[#0052D9] shadow-sm transition-all hover:bg-blue-50 disabled:cursor-not-allowed disabled:opacity-50"
                >
                  {demoLoading && <Loader2 className="h-4 w-4 animate-spin" />}
                  {demoLoading ? "正在进入…" : "🎯 评委一键体验"}
                </button>
                <p className="text-center text-[11px] text-gray-400 -mt-2">
                  使用示范账号快速查看完整 Demo 数据
                </p>

                <p className="text-center text-sm text-slate-500">
                  还没有账号？{" "}
                  <button type="button" onClick={() => switchMode("register")} className="font-medium text-[#0052D9] hover:text-[#0047BD] hover:underline">
                    注册
                  </button>
                </p>
              </form>
            ) : (
              <form onSubmit={handleRegister} className="space-y-4">
                <div>
                  <label htmlFor="auth-reg-phone" className="mb-1.5 block text-sm font-medium text-slate-700">
                    手机号
                  </label>
                  <Input
                    id="auth-reg-phone"
                    value={regPhone}
                    onChange={(e) => setRegPhone(e.target.value)}
                    placeholder="请输入手机号"
                    autoComplete="tel"
                  />
                </div>
                <div>
                  <label htmlFor="auth-reg-password" className="mb-1.5 block text-sm font-medium text-slate-700">
                    密码
                  </label>
                  <PasswordInput
                    id="auth-reg-password"
                    value={regPassword}
                    onChange={(e) => setRegPassword(e.target.value)}
                    placeholder="至少 6 位"
                    autoComplete="new-password"
                  />
                </div>
                <div>
                  <label htmlFor="auth-reg-confirm" className="mb-1.5 block text-sm font-medium text-slate-700">
                    确认密码
                  </label>
                  <PasswordInput
                    id="auth-reg-confirm"
                    value={regConfirm}
                    onChange={(e) => setRegConfirm(e.target.value)}
                    placeholder="请再次输入密码"
                    autoComplete="new-password"
                  />
                </div>

                {regError && (
                  <p className="animate-auth-error-in rounded-xl bg-red-50/90 px-4 py-3 text-sm text-red-600 ring-1 ring-red-100" role="alert">
                    {regError}
                  </p>
                )}

                <button
                  type="submit"
                  disabled={regLoading}
                  className="flex h-11 w-full items-center justify-center gap-2 rounded-xl bg-[#0052D9] text-sm font-medium text-white shadow-sm transition-colors hover:bg-[#0047BD] disabled:cursor-not-allowed disabled:opacity-50"
                >
                  {regLoading && <Loader2 className="h-4 w-4 animate-spin" />}
                  {regLoading ? "注册中…" : "注册"}
                </button>

                <p className="text-center text-sm text-slate-500">
                  已有账号？{" "}
                  <button type="button" onClick={() => switchMode("login")} className="font-medium text-[#0052D9] hover:text-[#0047BD] hover:underline">
                    登录
                  </button>
                </p>
              </form>
            )}
          </div>
          </div>{/* /滚动容器 */}
        </div>{/* /动画外壳 */}
      </DialogPortal>
    </Dialog>
  );
}
