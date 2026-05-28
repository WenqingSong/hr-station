import {
  FileText,
  CheckCircle2,
  AlertCircle,
  AlertTriangle,
} from "lucide-react";
import { color } from "../tokens";
import { PageShell, PageHeader } from "../components/layout/PageShell";
import { Card, CardHeader, CardDivider } from "../components/ui/Card";
import { Button } from "../components/ui/Button";
import { TextTag } from "../components/ui/StatusTag";

/* ─────────────── Version Info Bar ─────────────── */

function VersionBar() {
  const items = [
    {
      label: "当前版本",
      value: (
        <TextTag
          label="V2"
          style={{ color: color.blue, background: color.blueLight, fontWeight: 700 }}
        />
      ),
    },
    { label: "最近添加", value: "2026-05-29" },
    { label: "最近目标岗位", value: "后端开发实习生" },
    {
      label: "最近匹配度",
      value: (
        <span style={{ fontWeight: 700, color: color.blue, fontSize: 15 }}>
          84 / 100
        </span>
      ),
    },
  ];

  return (
    <Card style={{ padding: "14px 20px", marginBottom: 14 }}>
      <div style={{ display: "flex", alignItems: "center", gap: 0, flexWrap: "wrap" }}>
        {items.map((item, i) => (
          <div
            key={item.label}
            style={{
              display: "flex",
              alignItems: "center",
              gap: 8,
              paddingRight: i < items.length - 1 ? 24 : 0,
              marginRight: i < items.length - 1 ? 24 : 0,
              borderRight: i < items.length - 1 ? `1px solid ${color.border}` : "none",
              padding: "4px 0",
            }}
          >
            <span style={{ fontSize: 12, color: color.textTertiary, whiteSpace: "nowrap" }}>
              {item.label}
            </span>
            <div style={{ fontSize: 14, color: color.textPrimary, fontWeight: 500 }}>
              {item.value}
            </div>
          </div>
        ))}
      </div>
    </Card>
  );
}

/* ─────────────── Left: Resume Content ─────────────── */

function SectionLabel({ children }: { children: React.ReactNode }) {
  return (
    <div
      style={{
        fontSize: 11,
        fontWeight: 700,
        color: color.textTertiary,
        letterSpacing: "0.6px",
        marginBottom: 10,
        paddingBottom: 7,
        borderBottom: `1px solid ${color.divider}`,
      }}
    >
      {children}
    </div>
  );
}

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <div
      style={{
        display: "flex",
        gap: 0,
        marginBottom: 6,
        fontSize: 13,
      }}
    >
      <span style={{ color: color.textTertiary, minWidth: 88, flexShrink: 0 }}>{label}</span>
      <span style={{ color: color.textPrimary }}>{value}</span>
    </div>
  );
}

function ResumeContentCard() {
  return (
    <Card style={{ flex: 1 }}>
      <CardHeader title="简历内容" subtitle="V2 · 2026-05-29" />

      {/* Basic Info */}
      <div style={{ marginBottom: 20 }}>
        <SectionLabel>基本信息</SectionLabel>
        <InfoRow label="姓名" value="宋文清" />
        <InfoRow label="所在城市" value="合肥" />
        <InfoRow label="邮箱" value="1784870640@qq.com" />
      </div>

      {/* Job Intent */}
      <div style={{ marginBottom: 20 }}>
        <SectionLabel>求职意向</SectionLabel>
        <InfoRow label="目标方向" value="后端开发" />
        <InfoRow label="期望城市" value="深圳" />
        <InfoRow label="求职类型" value="实习 / 校招" />
      </div>

      {/* Education */}
      <div style={{ marginBottom: 20 }}>
        <SectionLabel>教育经历</SectionLabel>
        <InfoRow label="学校" value="某某大学" />
        <InfoRow label="专业" value="计算机相关专业" />
        <InfoRow label="学历" value="本科" />
        <InfoRow label="时间" value="2022.09 — 2026.06" />
      </div>

      {/* Projects */}
      <div style={{ marginBottom: 20 }}>
        <SectionLabel>项目经历</SectionLabel>
        <div
          style={{
            background: color.pageBg,
            borderRadius: 7,
            padding: "12px 14px",
          }}
        >
          <div
            style={{
              fontSize: 13,
              fontWeight: 600,
              color: color.textPrimary,
              marginBottom: 8,
            }}
          >
            后台管理系统
          </div>
          <div style={{ marginBottom: 6 }}>
            <span style={{ fontSize: 12, color: color.textTertiary }}>使用技术：</span>
            <span style={{ fontSize: 13, color: color.textSecondary }}>
              Go、MySQL、Redis、GoFrame
            </span>
          </div>
          <div>
            <span style={{ fontSize: 12, color: color.textTertiary }}>个人负责：</span>
            <span style={{ fontSize: 13, color: color.textSecondary }}>
              权限管理、角色管理、数据查询接口开发
            </span>
          </div>
        </div>
      </div>

      {/* Work */}
      <div style={{ marginBottom: 20 }}>
        <SectionLabel>实习 / 工作经历</SectionLabel>
        <div style={{ fontSize: 13, color: color.textDisabled, fontStyle: "italic" }}>
          暂无填写
        </div>
      </div>

      {/* Self Assessment */}
      <div>
        <SectionLabel>自我评价</SectionLabel>
        <p
          style={{
            fontSize: 13,
            color: color.textSecondary,
            lineHeight: 1.7,
            margin: 0,
          }}
        >
          希望从事后端开发方向，熟悉 Go 和数据库，有后台系统项目经验，在校期间专注于后端技术栈学习与实践。
        </p>
      </div>
    </Card>
  );
}

/* ─────────────── Right: AI Extraction ─────────────── */

function ExtractionCard() {
  const skills = ["Go", "MySQL", "Redis", "Linux", "GoFrame"];
  const directions = ["后台管理系统", "权限管理", "接口开发", "数据查询"];
  const abilities = ["后端开发", "数据库使用", "缓存使用", "接口开发"];

  const pros = [
    "后端技术栈集中",
    "有后台系统完整项目经历",
    "项目方向与后端岗位相关",
  ];
  const cons = [
    "Redis 只有关键词，缺少具体业务场景",
    "项目贡献表达不够清晰",
    "缺少量化结果描述",
  ];

  return (
    <Card style={{ width: 340, flexShrink: 0 }}>
      <CardHeader title="AI 简历信息提取" />

      {/* Skills */}
      <div style={{ marginBottom: 18 }}>
        <div style={{ fontSize: 12, color: color.textTertiary, marginBottom: 8 }}>
          识别到的技能关键词
        </div>
        <div style={{ display: "flex", flexWrap: "wrap", gap: 5 }}>
          {skills.map((s) => (
            <TextTag
              key={s}
              label={s}
              style={{ color: color.blue, background: color.blueLight }}
            />
          ))}
        </div>
      </div>

      {/* Directions */}
      <div style={{ marginBottom: 18 }}>
        <div style={{ fontSize: 12, color: color.textTertiary, marginBottom: 8 }}>
          项目方向
        </div>
        <div style={{ display: "flex", flexWrap: "wrap", gap: 5 }}>
          {directions.map((d) => (
            <TextTag key={d} label={d} />
          ))}
        </div>
      </div>

      {/* Abilities */}
      <div style={{ marginBottom: 18 }}>
        <div style={{ fontSize: 12, color: color.textTertiary, marginBottom: 8 }}>
          岗位相关能力
        </div>
        <div style={{ display: "flex", flexWrap: "wrap", gap: 5 }}>
          {abilities.map((a) => (
            <TextTag key={a} label={a} />
          ))}
        </div>
      </div>

      <CardDivider />

      {/* Pros */}
      <div style={{ marginBottom: 16 }}>
        <div
          style={{
            fontSize: 12,
            fontWeight: 600,
            color: color.successText,
            marginBottom: 8,
            display: "flex",
            alignItems: "center",
            gap: 5,
          }}
        >
          <div
            style={{ width: 3, height: 11, borderRadius: 2, background: color.success }}
          />
          当前优势
        </div>
        {pros.map((p) => (
          <div
            key={p}
            style={{ display: "flex", alignItems: "flex-start", gap: 6, marginBottom: 6 }}
          >
            <CheckCircle2
              size={13}
              color={color.success}
              style={{ marginTop: 2, flexShrink: 0 }}
            />
            <span style={{ fontSize: 13, color: color.textSecondary, lineHeight: 1.6 }}>
              {p}
            </span>
          </div>
        ))}
      </div>

      {/* Cons */}
      <div>
        <div
          style={{
            fontSize: 12,
            fontWeight: 600,
            color: color.warnText,
            marginBottom: 8,
            display: "flex",
            alignItems: "center",
            gap: 5,
          }}
        >
          <div
            style={{ width: 3, height: 11, borderRadius: 2, background: color.warn }}
          />
          当前短板
        </div>
        {cons.map((c) => (
          <div
            key={c}
            style={{ display: "flex", alignItems: "flex-start", gap: 6, marginBottom: 6 }}
          >
            <AlertCircle
              size={13}
              color={color.warn}
              style={{ marginTop: 2, flexShrink: 0 }}
            />
            <span style={{ fontSize: 13, color: color.textSecondary, lineHeight: 1.6 }}>
              {c}
            </span>
          </div>
        ))}
      </div>
    </Card>
  );
}

/* ─────────────── Expression Risk Card ─────────────── */

const RISKS = [
  "Redis 只出现关键词，缺少真实业务场景，HR 很难判断实际使用深度。",
  "项目描述偏技术栈罗列，个人贡献边界不够清晰，建议补充「负责什么、完成了什么」。",
  "缺少量化结果，如接口数量、数据规模、响应时间等可量化内容。",
];

function ExpressionRiskCard() {
  return (
    <Card>
      <CardHeader title="表达风险" />
      <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
        {RISKS.map((risk, i) => (
          <div
            key={i}
            style={{
              display: "flex",
              alignItems: "flex-start",
              gap: 12,
              padding: "11px 14px",
              background: color.warnBg,
              borderRadius: 8,
              border: `1px solid #FFE4B5`,
            }}
          >
            <AlertTriangle
              size={14}
              color={color.warn}
              style={{ marginTop: 2, flexShrink: 0 }}
            />
            <span style={{ fontSize: 13, color: color.textSecondary, lineHeight: 1.65 }}>
              {risk}
            </span>
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
      <Button variant="outline" size="md">去岗位推荐</Button>
      <Button variant="outline" size="md">去岗位列表匹配</Button>
      <Button variant="ghost"   size="md">查看版本迭代</Button>
    </div>
  );
}

/* ─────────────── Empty State ─────────────── */

function EmptyState() {
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
        <FileText size={22} color={color.textTertiary} strokeWidth={1.5} />
      </div>
      <div
        style={{ fontSize: 15, fontWeight: 600, color: color.textPrimary, marginBottom: 8 }}
      >
        当前还没有简历
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
        请先添加简历，系统会基于你的简历内容进行岗位推荐和岗位匹配分析。
      </div>
      <Button variant="primary" size="md">添加简历</Button>
    </Card>
  );
}

/* ─────────────── Page ─────────────── */

export function ResumeDetailPage({ hasResume = true }: { hasResume?: boolean }) {
  return (
    <PageShell activeNav="resume">
      <PageHeader
        title="简历详情"
        subtitle="查看当前简历内容、系统识别结果和表达风险。"
      />

      {hasResume ? (
        <div style={{ display: "flex", flexDirection: "column", gap: 14 }}>
          <VersionBar />

          {/* Two-column area */}
          <div style={{ display: "flex", gap: 14, alignItems: "flex-start" }}>
            <ResumeContentCard />
            <ExtractionCard />
          </div>

          <ExpressionRiskCard />
          <BottomActions />
        </div>
      ) : (
        <EmptyState />
      )}
    </PageShell>
  );
}
