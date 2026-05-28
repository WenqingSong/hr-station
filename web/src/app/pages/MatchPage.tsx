import { useState } from "react";
import {
  FileText,
  Briefcase,
  CheckCircle2,
  AlertTriangle,
  ArrowRight,
  ShieldCheck,
} from "lucide-react";
import { color } from "../tokens";
import { PageShell, PageHeader } from "../components/layout/PageShell";
import { Card, CardHeader, CardDivider } from "../components/ui/Card";
import { Button } from "../components/ui/Button";
import { StatusTag, TextTag } from "../components/ui/StatusTag";
import { ScoreRing } from "../components/ui/ScoreRing";

/* ─────────────── Types ─────────────── */

type EvidenceStrength = "strong" | "mid" | "weak" | "none";
type RiskLevel = "low" | "mid" | "high";

interface EvidenceRow {
  requirement: string;
  evidence: string;
  strength: EvidenceStrength;
  risk: RiskLevel;
  suggestion: string;
}

const STRENGTH_MAP: Record<EvidenceStrength, { label: string; status: "success" | "warn" | "danger" | "neutral" }> = {
  strong: { label: "强", status: "success" },
  mid:    { label: "中", status: "warn" },
  weak:   { label: "弱", status: "danger" },
  none:   { label: "无", status: "neutral" },
};

const RISK_MAP: Record<RiskLevel, { label: string; status: "success" | "warn" | "danger" }> = {
  low:  { label: "低", status: "success" },
  mid:  { label: "中", status: "warn" },
  high: { label: "高", status: "danger" },
};

const EVIDENCE_ROWS: EvidenceRow[] = [
  {
    requirement: "熟悉 Go",
    evidence: "GoFrame 后台项目开发",
    strength: "strong",
    risk: "low",
    suggestion: "可补充负责的具体模块和接口场景",
  },
  {
    requirement: "熟悉 MySQL",
    evidence: "使用 MySQL 完成业务数据存储",
    strength: "mid",
    risk: "mid",
    suggestion: "补充表设计、索引优化、事务相关经历",
  },
  {
    requirement: "熟悉 Redis",
    evidence: "只出现 Redis 关键词",
    strength: "weak",
    risk: "high",
    suggestion: "补充真实缓存业务场景，如缓存击穿处理",
  },
  {
    requirement: "后台系统开发经验",
    evidence: "后台管理系统项目",
    strength: "mid",
    risk: "mid",
    suggestion: "补充个人贡献和可量化的业务结果",
  },
  {
    requirement: "性能优化经验",
    evidence: "无明确证据",
    strength: "none",
    risk: "high",
    suggestion: "不建议虚构，建议补充真实实践后再填写",
  },
];

/* ─────────────── Match Object Info ─────────────── */

function MatchInfoBar() {
  return (
    <Card style={{ padding: "14px 20px" }}>
      <div style={{ display: "flex", alignItems: "center", gap: 24, flexWrap: "wrap" }}>
        <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
          <span style={{ fontSize: 12, color: color.textTertiary }}>当前简历版本</span>
          <TextTag label="V2" style={{ color: color.blue, background: color.blueLight, fontWeight: 700 }} />
        </div>
        <div style={{ width: 1, height: 16, background: color.border }} />
        <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
          <Briefcase size={14} color={color.textTertiary} />
          <span style={{ fontSize: 14, fontWeight: 600, color: color.textPrimary }}>
            后端开发实习生
          </span>
          <span style={{ fontSize: 13, color: color.textSecondary }}>
            技术 · 应届实习 · 深圳 / 北京 / 上海
          </span>
        </div>
        <div style={{ marginLeft: "auto" }}>
          <span style={{ fontSize: 12, color: color.textTertiary }}>分析时间：</span>
          <span style={{ fontSize: 12, color: color.textSecondary }}>2026-05-29 14:20</span>
        </div>
      </div>
    </Card>
  );
}

/* ─────────────── Match Overview ─────────────── */

function MatchOverview() {
  const pros = [
    "Go 与后台项目经历匹配",
    "MySQL 有基础使用经历",
    "项目方向与岗位较一致",
  ];
  const risks = [
    "Redis 只有关键词，没有业务场景",
    "性能优化无明确证据",
    "项目贡献表达不够清晰",
  ];

  return (
    <Card>
      <CardHeader title="匹配总览" />
      <div style={{ display: "flex", gap: 24, alignItems: "flex-start" }}>
        {/* Score block */}
        <div
          style={{
            display: "flex",
            flexDirection: "column",
            alignItems: "center",
            gap: 10,
            paddingRight: 24,
            borderRight: `1px solid ${color.divider}`,
            flexShrink: 0,
            minWidth: 110,
          }}
        >
          <ScoreRing score={72} size={80} strokeWidth={7} />
          <div style={{ textAlign: "center" }}>
            <div style={{ fontSize: 12, color: color.textTertiary }}>综合匹配度</div>
          </div>
        </div>

        {/* Verdict + pros/risks */}
        <div style={{ flex: 1 }}>
          {/* Verdict */}
          <div
            style={{
              display: "flex",
              alignItems: "center",
              gap: 10,
              marginBottom: 16,
              padding: "10px 14px",
              background: color.warnBg,
              borderRadius: 8,
              border: `1px solid #FFE4B5`,
            }}
          >
            <AlertTriangle size={15} color={color.warn} />
            <span style={{ fontSize: 13, fontWeight: 600, color: color.warnText }}>
              B 类，可以投递，但建议优化简历后再投
            </span>
          </div>

          {/* Pros & Risks */}
          <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "0 24px" }}>
            <div>
              <div
                style={{
                  fontSize: 12,
                  fontWeight: 600,
                  color: color.successText,
                  marginBottom: 10,
                  display: "flex",
                  alignItems: "center",
                  gap: 5,
                }}
              >
                <div style={{ width: 3, height: 12, borderRadius: 2, background: color.success }} />
                核心优势
              </div>
              {pros.map((p) => (
                <div key={p} style={{ display: "flex", alignItems: "flex-start", gap: 7, marginBottom: 7 }}>
                  <CheckCircle2 size={13} color={color.success} style={{ marginTop: 2, flexShrink: 0 }} />
                  <span style={{ fontSize: 13, color: color.textSecondary, lineHeight: 1.6 }}>{p}</span>
                </div>
              ))}
            </div>
            <div>
              <div
                style={{
                  fontSize: 12,
                  fontWeight: 600,
                  color: color.dangerText,
                  marginBottom: 10,
                  display: "flex",
                  alignItems: "center",
                  gap: 5,
                }}
              >
                <div style={{ width: 3, height: 12, borderRadius: 2, background: color.danger }} />
                主要风险
              </div>
              {risks.map((r) => (
                <div key={r} style={{ display: "flex", alignItems: "flex-start", gap: 7, marginBottom: 7 }}>
                  <AlertTriangle size={13} color={color.warn} style={{ marginTop: 2, flexShrink: 0 }} />
                  <span style={{ fontSize: 13, color: color.textSecondary, lineHeight: 1.6 }}>{r}</span>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </Card>
  );
}

/* ─────────────── Evidence Table ─────────────── */

const COL_WIDTHS = ["22%", "24%", "9%", "9%", "36%"];
const HEADERS = ["岗位要求", "简历证据", "证据强度", "初筛风险", "优化建议"];

function EvidenceTable() {
  const [hovered, setHovered] = useState<number | null>(null);

  return (
    <Card>
      <CardHeader title="JD-简历证据链" />
      <p style={{ fontSize: 13, color: color.textTertiary, marginBottom: 16, lineHeight: 1.6 }}>
        系统将岗位要求与简历中的证据逐项对齐，帮助你判断哪些能力有证据支撑，哪些存在初筛风险。
      </p>

      <div style={{ overflowX: "auto" }}>
        <table style={{ width: "100%", borderCollapse: "collapse" }}>
          {/* Header */}
          <thead>
            <tr style={{ background: "#FAFBFC" }}>
              {HEADERS.map((h, i) => (
                <th
                  key={h}
                  style={{
                    width: COL_WIDTHS[i],
                    padding: "10px 14px",
                    textAlign: "left",
                    fontSize: 12,
                    fontWeight: 600,
                    color: color.textSecondary,
                    borderBottom: `1px solid ${color.border}`,
                    borderTop: `1px solid ${color.border}`,
                    whiteSpace: "nowrap",
                  }}
                >
                  {h}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {EVIDENCE_ROWS.map((row, i) => {
              const s = STRENGTH_MAP[row.strength];
              const r = RISK_MAP[row.risk];
              const isLast = i === EVIDENCE_ROWS.length - 1;
              return (
                <tr
                  key={i}
                  style={{
                    background: hovered === i ? color.pageBg : "#fff",
                    transition: "background 0.12s",
                  }}
                  onMouseEnter={() => setHovered(i)}
                  onMouseLeave={() => setHovered(null)}
                >
                  <td
                    style={{
                      padding: "13px 14px",
                      fontSize: 13,
                      color: color.textPrimary,
                      fontWeight: 500,
                      borderBottom: isLast ? "none" : `1px solid ${color.divider}`,
                      verticalAlign: "top",
                    }}
                  >
                    {row.requirement}
                  </td>
                  <td
                    style={{
                      padding: "13px 14px",
                      fontSize: 13,
                      color: row.evidence === "无明确证据" ? color.textDisabled : color.textSecondary,
                      borderBottom: isLast ? "none" : `1px solid ${color.divider}`,
                      verticalAlign: "top",
                      fontStyle: row.evidence === "无明确证据" ? "italic" : "normal",
                    }}
                  >
                    {row.evidence}
                  </td>
                  <td
                    style={{
                      padding: "13px 14px",
                      borderBottom: isLast ? "none" : `1px solid ${color.divider}`,
                      verticalAlign: "top",
                    }}
                  >
                    <StatusTag status={s.status} label={s.label} dot={false} />
                  </td>
                  <td
                    style={{
                      padding: "13px 14px",
                      borderBottom: isLast ? "none" : `1px solid ${color.divider}`,
                      verticalAlign: "top",
                    }}
                  >
                    <StatusTag status={r.status} label={r.label} dot={false} />
                  </td>
                  <td
                    style={{
                      padding: "13px 14px",
                      fontSize: 13,
                      color: color.textSecondary,
                      lineHeight: 1.65,
                      borderBottom: isLast ? "none" : `1px solid ${color.divider}`,
                      verticalAlign: "top",
                    }}
                  >
                    {row.suggestion}
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </Card>
  );
}

/* ─────────────── Authenticity Card ─────────────── */

function AuthenticityCard() {
  const categories = [
    { label: "可直接优化表达", status: "success" as const, desc: "改善描述方式，不改变事实" },
    { label: "需要用户确认",   status: "warn"    as const, desc: "系统存疑，请核实后再填" },
    { label: "当前无证据，不建议虚构", status: "neutral" as const, desc: "如实留空，避免初筛被识别" },
  ];

  return (
    <Card>
      <CardHeader
        title="真实性保护"
        extra={
          <div style={{ display: "flex", alignItems: "center", gap: 5 }}>
            <ShieldCheck size={14} color={color.blue} />
            <span style={{ fontSize: 12, color: color.blue }}>系统承诺</span>
          </div>
        }
      />
      <div
        style={{
          padding: "12px 16px",
          background: color.blueLight,
          borderRadius: 8,
          marginBottom: 16,
          fontSize: 13,
          color: color.blue,
          lineHeight: 1.65,
        }}
      >
        系统不会帮助用户虚构经历。涉及数据、规模、性能提升、业务结果等内容时，如简历中没有证据，将标记为「需要用户确认」或「当前无证据」。
      </div>
      <div style={{ display: "flex", gap: 10, flexWrap: "wrap" }}>
        {categories.map((c) => (
          <div
            key={c.label}
            style={{
              display: "flex",
              flexDirection: "column",
              gap: 5,
              padding: "10px 14px",
              border: `1px solid ${color.border}`,
              borderRadius: 8,
              flex: 1,
              minWidth: 160,
            }}
          >
            <StatusTag status={c.status} label={c.label} />
            <span style={{ fontSize: 12, color: color.textTertiary, lineHeight: 1.5 }}>{c.desc}</span>
          </div>
        ))}
      </div>
    </Card>
  );
}

/* ─────────────── Optimization Suggestions ─────────────── */

function OptimizationCard() {
  const p0 = [
    "项目经历只写了技术栈，没有体现个人贡献和职责边界。",
    "Redis 只出现关键词，没有说明具体业务场景。",
  ];
  const p1 = [
    "MySQL 经历缺少表设计、索引优化、事务相关描述。",
    "后台系统项目缺少业务目标和可量化结果描述。",
  ];

  return (
    <Card>
      <CardHeader title="简历优化建议" />
      <div style={{ display: "flex", flexDirection: "column", gap: 14 }}>
        {/* P0 */}
        <div
          style={{
            border: `1px solid #FFD4CC`,
            borderRadius: 8,
            overflow: "hidden",
          }}
        >
          <div
            style={{
              padding: "8px 14px",
              background: color.dangerBg,
              display: "flex",
              alignItems: "center",
              gap: 8,
            }}
          >
            <span
              style={{
                fontSize: 11,
                fontWeight: 700,
                background: color.danger,
                color: "#fff",
                borderRadius: 3,
                padding: "1px 6px",
              }}
            >
              P0
            </span>
            <span style={{ fontSize: 12, fontWeight: 600, color: color.dangerText }}>
              必改项
            </span>
          </div>
          <div style={{ padding: "10px 14px 12px" }}>
            {p0.map((item, i) => (
              <div
                key={i}
                style={{
                  display: "flex",
                  alignItems: "flex-start",
                  gap: 8,
                  marginBottom: i < p0.length - 1 ? 8 : 0,
                }}
              >
                <span
                  style={{
                    fontSize: 12,
                    color: color.dangerText,
                    fontWeight: 700,
                    flexShrink: 0,
                    marginTop: 1,
                  }}
                >
                  {i + 1}.
                </span>
                <span style={{ fontSize: 13, color: color.textSecondary, lineHeight: 1.65 }}>
                  {item}
                </span>
              </div>
            ))}
          </div>
        </div>

        {/* P1 */}
        <div
          style={{
            border: `1px solid #C0D8FF`,
            borderRadius: 8,
            overflow: "hidden",
          }}
        >
          <div
            style={{
              padding: "8px 14px",
              background: color.blueLight,
              display: "flex",
              alignItems: "center",
              gap: 8,
            }}
          >
            <span
              style={{
                fontSize: 11,
                fontWeight: 700,
                background: color.blue,
                color: "#fff",
                borderRadius: 3,
                padding: "1px 6px",
              }}
            >
              P1
            </span>
            <span style={{ fontSize: 12, fontWeight: 600, color: color.blue }}>
              建议优化
            </span>
          </div>
          <div style={{ padding: "10px 14px 12px" }}>
            {p1.map((item, i) => (
              <div
                key={i}
                style={{
                  display: "flex",
                  alignItems: "flex-start",
                  gap: 8,
                  marginBottom: i < p1.length - 1 ? 8 : 0,
                }}
              >
                <span
                  style={{
                    fontSize: 12,
                    color: color.blue,
                    fontWeight: 700,
                    flexShrink: 0,
                    marginTop: 1,
                  }}
                >
                  {i + 1}.
                </span>
                <span style={{ fontSize: 13, color: color.textSecondary, lineHeight: 1.65 }}>
                  {item}
                </span>
              </div>
            ))}
          </div>
        </div>
      </div>
    </Card>
  );
}

/* ─────────────── Next Steps ─────────────── */

const STEPS = [
  { num: "1", label: "查看分析结果" },
  { num: "2", label: "手动修改简历" },
  { num: "3", label: "添加新版本" },
  { num: "4", label: "再次匹配" },
];

function NextStepsCard() {
  return (
    <Card>
      <CardHeader title="下一步建议" />
      <p style={{ fontSize: 13, color: color.textSecondary, lineHeight: 1.7, marginBottom: 20 }}>
        你可以根据以上建议自行修改简历，然后通过「添加新版本」提交。系统会将新简历保存为下一版本，并用于下一次岗位推荐或匹配分析。
      </p>
      <div style={{ display: "flex", alignItems: "center", gap: 0, flexWrap: "wrap" }}>
        {STEPS.map((step, i) => (
          <div key={step.num} style={{ display: "flex", alignItems: "center" }}>
            <div
              style={{
                display: "flex",
                flexDirection: "column",
                alignItems: "center",
                gap: 6,
              }}
            >
              <div
                style={{
                  width: 32,
                  height: 32,
                  borderRadius: "50%",
                  background: i === 0 ? color.blue : "#F2F3F5",
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "center",
                  fontSize: 13,
                  fontWeight: 700,
                  color: i === 0 ? "#fff" : color.textSecondary,
                }}
              >
                {step.num}
              </div>
              <span style={{ fontSize: 12, color: i === 0 ? color.blue : color.textSecondary, whiteSpace: "nowrap" }}>
                {step.label}
              </span>
            </div>
            {i < STEPS.length - 1 && (
              <div style={{ display: "flex", alignItems: "center", padding: "0 10px", paddingBottom: 22 }}>
                <ArrowRight size={14} color={color.textDisabled} />
              </div>
            )}
          </div>
        ))}
      </div>
    </Card>
  );
}

/* ─────────────── Bottom Actions ─────────────── */

function BottomActions() {
  return (
    <div
      style={{
        display: "flex",
        alignItems: "center",
        gap: 10,
        padding: "16px 24px",
        background: "#fff",
        borderRadius: 10,
        border: `1px solid ${color.border}`,
      }}
    >
      <Button variant="primary" size="md">添加新版本</Button>
      <Button variant="outline" size="md">保存为最近一次分析</Button>
      <Button variant="outline" size="md">返回岗位列表</Button>
      <Button variant="ghost" size="md">我的简历</Button>
    </div>
  );
}

/* ─────────────── Empty States ─────────────── */

function EmptyState({ type }: { type: "no-resume" | "no-job" }) {
  const config = {
    "no-resume": {
      icon: FileText,
      title: "请先添加简历",
      desc: "添加简历后，系统才能分析你与目标岗位的匹配度。",
      btn: "添加简历",
    },
    "no-job": {
      icon: Briefcase,
      title: "请先选择目标岗位",
      desc: "选择岗位后，系统才能生成匹配分析报告。",
      btn: "返回岗位列表",
    },
  }[type];

  const Icon = config.icon;

  return (
    <Card style={{ textAlign: "center", padding: "60px 24px" }}>
      <div
        style={{
          width: 52,
          height: 52,
          borderRadius: 12,
          background: "#F2F3F5",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          margin: "0 auto 18px",
        }}
      >
        <Icon size={22} color={color.textTertiary} strokeWidth={1.5} />
      </div>
      <div style={{ fontSize: 15, fontWeight: 600, color: color.textPrimary, marginBottom: 8 }}>
        {config.title}
      </div>
      <div
        style={{
          fontSize: 13,
          color: color.textTertiary,
          lineHeight: 1.7,
          maxWidth: 320,
          margin: "0 auto 28px",
        }}
      >
        {config.desc}
      </div>
      <Button variant="primary" size="md">{config.btn}</Button>
    </Card>
  );
}

/* ─────────────── State Switcher ─────────────── */

type PageState = "full" | "no-resume" | "no-job";

function StateSwitcher({ state, onChange }: { state: PageState; onChange: (s: PageState) => void }) {
  const options: { key: PageState; label: string }[] = [
    { key: "full",      label: "完整分析" },
    { key: "no-resume", label: "无简历" },
    { key: "no-job",    label: "无目标岗位" },
  ];
  return (
    <div
      style={{
        display: "flex",
        alignItems: "center",
        gap: 6,
        background: color.pageBg,
        border: `1px solid ${color.border}`,
        borderRadius: 6,
        padding: "4px 6px",
        marginBottom: 14,
        alignSelf: "flex-start",
      }}
    >
      <span style={{ fontSize: 11, color: color.textTertiary, paddingRight: 4 }}>状态：</span>
      {options.map((o) => (
        <button
          key={o.key}
          onClick={() => onChange(o.key)}
          style={{
            height: 24,
            padding: "0 10px",
            fontSize: 11,
            fontWeight: 500,
            borderRadius: 4,
            border: "none",
            cursor: "pointer",
            background: state === o.key ? color.blue : "transparent",
            color: state === o.key ? "#fff" : color.textSecondary,
            transition: "all 0.15s",
          }}
        >
          {o.label}
        </button>
      ))}
    </div>
  );
}

/* ─────────────── Page ─────────────── */

export function MatchPage() {
  const [pageState, setPageState] = useState<PageState>("full");

  return (
    <PageShell activeNav="jobs">
      <PageHeader
        title="岗位匹配分析"
        subtitle="分析当前简历与目标岗位 JD 的匹配度，识别初筛风险和优化方向。"
      />

      <StateSwitcher state={pageState} onChange={setPageState} />

      {pageState === "full" && (
        <div style={{ display: "flex", flexDirection: "column", gap: 14 }}>
          <MatchInfoBar />
          <MatchOverview />
          <EvidenceTable />
          <AuthenticityCard />
          <OptimizationCard />
          <NextStepsCard />
          <BottomActions />
        </div>
      )}

      {pageState === "no-resume" && <EmptyState type="no-resume" />}
      {pageState === "no-job"    && <EmptyState type="no-job" />}
    </PageShell>
  );
}
