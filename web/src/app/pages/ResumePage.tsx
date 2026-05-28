import { FileText, GitBranch, BarChart2, ArrowRight, FilePlus2, Clock } from "lucide-react";
import { color } from "../tokens";
import { PageShell, PageHeader } from "../components/layout/PageShell";
import { Card, CardHeader, CardDivider } from "../components/ui/Card";
import { Button } from "../components/ui/Button";
import { StatusTag, TextTag } from "../components/ui/StatusTag";
import { ScoreRing } from "../components/ui/ScoreRing";

/* ─── 有简历状态 ─── */

function ResumeStatusCard() {
  const fields = [
    { label: "当前版本", value: <TextTag label="V2" style={{ color: color.blue, background: color.blueLight, fontWeight: 600 }} /> },
    { label: "首次添加", value: "2026-05-28" },
    { label: "最近添加", value: "2026-05-29" },
    { label: "最近分析时间", value: "2026-05-29 14:20" },
    { label: "最近目标岗位", value: "后端开发实习生" },
    {
      label: "最近匹配度",
      value: <span style={{ fontWeight: 700, color: color.blue }}>84 / 100</span>,
    },
    {
      label: "当前状态",
      value: <StatusTag status="warn" label="建议优化后投递" />,
    },
  ];

  return (
    <Card>
      <CardHeader title="当前简历状态" />
      <div
        style={{
          display: "grid",
          gridTemplateColumns: "repeat(4, 1fr)",
          gap: "16px 0",
        }}
      >
        {fields.map((f, i) => (
          <div
            key={f.label}
            style={{
              paddingRight: 24,
              borderRight: (i + 1) % 4 !== 0 ? `1px solid ${color.divider}` : "none",
              paddingLeft: i % 4 !== 0 ? 24 : 0,
            }}
          >
            <div style={{ fontSize: 12, color: color.textTertiary, marginBottom: 7 }}>
              {f.label}
            </div>
            <div style={{ fontSize: 14, color: color.textPrimary, fontWeight: 500 }}>
              {f.value}
            </div>
          </div>
        ))}
      </div>
    </Card>
  );
}

const OPERATIONS = [
  {
    key: "detail",
    icon: FileText,
    title: "查看简历详情",
    desc: "查看当前简历内容与 AI 提取结果",
    btn: "查看详情",
    variant: "outline" as const,
  },
  {
    key: "add",
    icon: FilePlus2,
    title: "添加新版本",
    desc: "根据建议手动修改后，提交新的简历版本",
    btn: "添加版本",
    variant: "primary" as const,
  },
  {
    key: "versions",
    icon: GitBranch,
    title: "查看版本迭代",
    desc: "对比 V1 / V2 的匹配度和证据变化",
    btn: "查看迭代",
    variant: "outline" as const,
  },
  {
    key: "match",
    icon: BarChart2,
    title: "去岗位匹配",
    desc: "选择目标岗位，分析匹配度与初筛风险",
    btn: "去匹配",
    variant: "outline" as const,
  },
];

function OperationCards() {
  return (
    <div
      style={{
        display: "grid",
        gridTemplateColumns: "repeat(4, 1fr)",
        gap: 14,
      }}
    >
      {OPERATIONS.map((op) => {
        const Icon = op.icon;
        return (
          <Card
            key={op.key}
            hoverable
            style={{
              display: "flex",
              flexDirection: "column",
              justifyContent: "space-between",
              minHeight: 148,
            }}
          >
            <div>
              <div
                style={{
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "space-between",
                  marginBottom: 12,
                }}
              >
                <div
                  style={{
                    width: 34,
                    height: 34,
                    borderRadius: 8,
                    background: op.variant === "primary" ? color.blue : color.blueLight,
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "center",
                  }}
                >
                  <Icon
                    size={16}
                    color={op.variant === "primary" ? "#fff" : color.blue}
                    strokeWidth={1.8}
                  />
                </div>
                <ArrowRight size={14} color={color.textDisabled} />
              </div>
              <div
                style={{
                  fontSize: 14,
                  fontWeight: 600,
                  color: color.textPrimary,
                  marginBottom: 5,
                }}
              >
                {op.title}
              </div>
              <div
                style={{
                  fontSize: 12,
                  color: color.textTertiary,
                  lineHeight: 1.6,
                  marginBottom: 16,
                }}
              >
                {op.desc}
              </div>
            </div>
            <Button variant={op.variant} size="sm">
              {op.btn}
            </Button>
          </Card>
        );
      })}
    </div>
  );
}

function AnalysisSummaryCard() {
  const risks = ["Redis 证据偏弱", "性能优化无明确证据"];

  return (
    <Card>
      <CardHeader
        title="最近一次分析摘要"
        subtitle="后端开发实习生 · V2"
        extra={
          <Button variant="ghost" size="sm">
            查看分析结果 →
          </Button>
        }
      />

      <div style={{ display: "flex", gap: 0, alignItems: "stretch" }}>
        {/* Score block */}
        <div
          style={{
            display: "flex",
            flexDirection: "column",
            alignItems: "center",
            justifyContent: "center",
            paddingRight: 28,
            marginRight: 28,
            borderRight: `1px solid ${color.divider}`,
            gap: 8,
            minWidth: 100,
          }}
        >
          <ScoreRing score={84} size={72} />
          <span style={{ fontSize: 12, color: color.textTertiary }}>综合匹配度</span>
        </div>

        {/* Info block */}
        <div style={{ flex: 1, display: "grid", gridTemplateColumns: "1fr 1fr", gap: "14px 32px" }}>
          <div>
            <div style={{ fontSize: 12, color: color.textTertiary, marginBottom: 6 }}>目标岗位</div>
            <div style={{ fontSize: 14, color: color.textPrimary, fontWeight: 500 }}>
              后端开发实习生
            </div>
          </div>
          <div>
            <div style={{ fontSize: 12, color: color.textTertiary, marginBottom: 6 }}>简历版本</div>
            <TextTag
              label="V2"
              style={{ color: color.blue, background: color.blueLight, fontWeight: 600 }}
            />
          </div>
          <div>
            <div style={{ fontSize: 12, color: color.textTertiary, marginBottom: 6 }}>初筛判断</div>
            <StatusTag status="warn" label="B 类 · 建议优化后投递" />
          </div>
          <div>
            <div style={{ fontSize: 12, color: color.textTertiary, marginBottom: 6 }}>分析时间</div>
            <div
              style={{
                fontSize: 13,
                color: color.textSecondary,
                display: "flex",
                alignItems: "center",
                gap: 5,
              }}
            >
              <Clock size={12} color={color.textTertiary} />
              2026-05-29 14:20
            </div>
          </div>

          {/* Risk row - spans full width */}
          <div style={{ gridColumn: "1 / -1" }}>
            <CardDivider />
            <div style={{ fontSize: 12, color: color.textTertiary, marginBottom: 8 }}>
              主要风险
            </div>
            <div style={{ display: "flex", flexWrap: "wrap", gap: 8 }}>
              {risks.map((r) => (
                <span
                  key={r}
                  style={{
                    fontSize: 13,
                    color: color.warnText,
                    background: color.warnBg,
                    borderRadius: 4,
                    padding: "3px 10px",
                  }}
                >
                  {r}
                </span>
              ))}
            </div>
          </div>
        </div>
      </div>
    </Card>
  );
}

/* ─── 无简历状态 ─── */

function EmptyResumeCard() {
  return (
    <Card
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        padding: "60px 24px",
        textAlign: "center",
        minHeight: 280,
      }}
    >
      <div
        style={{
          width: 56,
          height: 56,
          borderRadius: 12,
          background: "#F2F3F5",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          marginBottom: 20,
        }}
      >
        <FileText size={24} color={color.textTertiary} strokeWidth={1.5} />
      </div>

      <div
        style={{
          fontSize: 15,
          fontWeight: 600,
          color: color.textPrimary,
          marginBottom: 8,
        }}
      >
        当前还没有简历版本
      </div>
      <div
        style={{
          fontSize: 13,
          color: color.textTertiary,
          lineHeight: 1.7,
          maxWidth: 340,
          marginBottom: 28,
        }}
      >
        请先添加第一版简历，系统会自动保存为 V1，并用于后续岗位推荐和岗位匹配分析。
      </div>
      <Button variant="primary" size="md">
        添加简历
      </Button>
    </Card>
  );
}

/* ─── Page ─── */

interface ResumePageProps {
  hasResume?: boolean;
}

export function ResumePage({ hasResume = true }: ResumePageProps) {
  return (
    <PageShell activeNav="resume">
      <PageHeader
        title="我的简历"
        subtitle="查看当前简历版本，并根据岗位匹配建议持续迭代。"
      />

      {hasResume ? (
        <div style={{ display: "flex", flexDirection: "column", gap: 14 }}>
          <ResumeStatusCard />
          <OperationCards />
          <AnalysisSummaryCard />
        </div>
      ) : (
        <EmptyResumeCard />
      )}
    </PageShell>
  );
}
