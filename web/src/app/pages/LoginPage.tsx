import { useState } from "react";
import { AppNav } from "../components/AppNav";
import { color, radius, shadow } from "../tokens";

/* ─────────────── Input ─────────────── */

function FormInput({
  label,
  type = "text",
  placeholder,
  value,
  onChange,
}: {
  label: string;
  type?: string;
  placeholder: string;
  value: string;
  onChange: (v: string) => void;
}) {
  const [focused, setFocused] = useState(false);

  return (
    <div style={{ marginBottom: 14 }}>
      <label
        style={{
          display: "block",
          fontSize: 13,
          color: color.textSecondary,
          marginBottom: 6,
        }}
      >
        {label}
      </label>
      <input
        type={type}
        placeholder={placeholder}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        onFocus={() => setFocused(true)}
        onBlur={() => setFocused(false)}
        style={{
          width: "100%",
          height: 40,
          padding: "0 12px",
          fontSize: 14,
          color: color.textPrimary,
          background: "#fff",
          border: `1px solid ${focused ? color.blue : color.border}`,
          borderRadius: 8,
          outline: "none",
          boxSizing: "border-box",
          transition: "border-color 0.15s",
          boxShadow: focused ? `0 0 0 3px rgba(29,111,255,0.08)` : "none",
        }}
      />
    </div>
  );
}

/* ─────────────── Login Card ─────────────── */

type Mode = "login" | "register";

function LoginCard() {
  const [mode, setMode] = useState<Mode>("login");
  const [account, setAccount] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");

  return (
    <div
      style={{
        background: "#fff",
        borderRadius: 16,
        border: `1px solid ${color.border}`,
        boxShadow: shadow.card,
        width: 440,
        padding: "36px 40px 32px",
        boxSizing: "border-box",
      }}
    >
      {/* Header */}
      <div style={{ marginBottom: 28 }}>
        <div
          style={{
            fontSize: 20,
            fontWeight: 700,
            color: color.textPrimary,
            marginBottom: 8,
          }}
        >
          {mode === "login" ? "登录 Offer 捕手" : "注册 Offer 捕手"}
        </div>
        <div
          style={{
            fontSize: 13,
            color: color.textTertiary,
            lineHeight: 1.6,
          }}
        >
          {mode === "login"
            ? "保存你的简历版本、岗位推荐结果和匹配分析记录。"
            : "创建账号，开始你的 AI 求职之旅。"}
        </div>
      </div>

      {/* Form */}
      <div>
        <FormInput
          label="邮箱 / 手机号"
          placeholder="请输入邮箱或手机号"
          value={account}
          onChange={setAccount}
        />
        <FormInput
          label="密码"
          type="password"
          placeholder="请输入密码"
          value={password}
          onChange={setPassword}
        />
        {mode === "register" && (
          <FormInput
            label="确认密码"
            type="password"
            placeholder="请再次输入密码"
            value={confirmPassword}
            onChange={setConfirmPassword}
          />
        )}
      </div>

      {/* Primary Button */}
      <button
        style={{
          width: "100%",
          height: 42,
          background: color.blue,
          color: "#fff",
          fontSize: 14,
          fontWeight: 600,
          border: "none",
          borderRadius: 8,
          cursor: "pointer",
          marginTop: 6,
          marginBottom: 10,
          transition: "background 0.15s",
        }}
        onMouseEnter={(e) => {
          (e.currentTarget as HTMLButtonElement).style.background = color.blueDark;
        }}
        onMouseLeave={(e) => {
          (e.currentTarget as HTMLButtonElement).style.background = color.blue;
        }}
      >
        {mode === "login" ? "登录" : "注册"}
      </button>

      {/* Secondary Button */}
      <button
        onClick={() => {
          setMode(mode === "login" ? "register" : "login");
          setConfirmPassword("");
        }}
        style={{
          width: "100%",
          height: 40,
          background: "#fff",
          color: color.blue,
          fontSize: 14,
          fontWeight: 500,
          border: `1px solid ${color.blue}`,
          borderRadius: 8,
          cursor: "pointer",
          marginBottom: 20,
          transition: "background 0.15s",
        }}
        onMouseEnter={(e) => {
          (e.currentTarget as HTMLButtonElement).style.background = color.blueLight;
        }}
        onMouseLeave={(e) => {
          (e.currentTarget as HTMLButtonElement).style.background = "#fff";
        }}
      >
        {mode === "login" ? "注册账号" : "返回登录"}
      </button>

      {/* Divider */}
      <div
        style={{
          display: "flex",
          alignItems: "center",
          gap: 10,
          marginBottom: 16,
        }}
      >
        <div style={{ flex: 1, height: 1, background: color.border }} />
        <span style={{ fontSize: 12, color: color.textDisabled }}>或</span>
        <div style={{ flex: 1, height: 1, background: color.border }} />
      </div>

      {/* Demo Account Button */}
      <div style={{ textAlign: "center" }}>
        <button
          style={{
            background: "none",
            border: "none",
            fontSize: 13,
            color: color.blue,
            cursor: "pointer",
            padding: "4px 0",
            marginBottom: 8,
          }}
        >
          使用示例账号快速体验
        </button>
        <div
          style={{
            fontSize: 12,
            color: color.textDisabled,
            lineHeight: 1.6,
          }}
        >
          示例账号会自动载入示例简历和岗位库，便于快速体验完整流程。
        </div>
      </div>
    </div>
  );
}

/* ─────────────── Page ─────────────── */

export function LoginPage() {
  return (
    <div style={{ minHeight: "100vh", background: color.pageBg }}>
      <AppNav activeNav="login" />

      <main
        style={{
          paddingTop: "56px",
          minHeight: "100vh",
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          justifyContent: "center",
        }}
      >
        <div
          style={{
            display: "flex",
            flexDirection: "column",
            alignItems: "center",
            padding: "48px 24px 64px",
          }}
        >
          <LoginCard />

          {/* Bottom disclaimer */}
          <p
            style={{
              marginTop: 20,
              fontSize: 12,
              color: color.textDisabled,
              textAlign: "center",
              maxWidth: 440,
              lineHeight: 1.6,
            }}
          >
            本系统仅用于 AI 求职匹配与简历诊断 Demo，不涉及正式投递。
          </p>
        </div>
      </main>
    </div>
  );
}
