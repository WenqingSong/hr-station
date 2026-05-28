import { useState } from "react";
import { MapPin, X, FileText, CheckCircle2, AlertCircle } from "lucide-react";
import { color } from "../tokens";
import { PageShell, PageHeader } from "../components/layout/PageShell";
import { Card, CardHeader, CardDivider } from "../components/ui/Card";
import { Button } from "../components/ui/Button";
import { StatusTag, TextTag } from "../components/ui/StatusTag";
import { ScoreRing } from "../components/ui/ScoreRing";

/* ─────────────── Types & Data ─────────────── */

type RecommendLevel = "priority" | "try" | "low";
type RiskLevel = "low" | "mid" | "mid-high" | "high";

interface RecommendJob {
  rank: number;
  title: string;
  type: string;
  locations: string[];
  score: number;
  level: RecommendLevel;
  risk: RiskLevel;
  reasons: string[];
  shortfalls: string[];
  desc: string;
  requirements: string;
  tags: string[];
}

const LEVEL_MAP: Record<RecommendLevel, { label: string; status: "success" | "info" | "neutral" }> = {
  priority: { label: "优先推荐", status: "success" },
  try:      { label: "可以尝试", status: "info" },
  low:      { label: "暂不优先", status: "neutral" },
};

const RISK_MAP: Record<RiskLevel, { label: string; status: "success" | "warn" | "danger" | "neutral" }> = {
  low:      { label: "低",   status: "success" },
  mid:      { label: "中",   status: "warn" },
  "mid-high":{ label: "中高", status: "danger" },
  high:     { label: "高",   status: "danger" },
};

const RECOMMENDATIONS: RecommendJob[] = [
  {
    rank: 1,
    title: "后端开发实习生",
    type: "技术 · 应届实习",
    locations: ["深圳", "北京", "上海"],
    score: 86,
    level: "priority",
    risk: "mid",
    tags: ["Go", "MySQL", "Redis", "后台系统"],
    reasons: [
      "简历中出现 Go、MySQL、Redis 等核心技能",
      "有后台管理系统项目经历",
      "项目方向与岗位 JD 较一致",
    ],
    shortfalls: [
      "Redis 只有关键词，缺少业务场景描述",
      "性能优化和高并发经验无明确证据",
    ],
    desc: "负责核心后台服务的设计与开发，参与系统架构优化，与产品、前端协作完成业务迭代。",
    requirements: "熟悉 Go / Java / Python 等后端语言之一；了解 MySQL、Redis 等常见中间件；有实际项目经验优先。",
  },
  {
    rank: 2,
    title: "云计算开发实习生",
    type: "技术 · 应届实习",
    locations: ["深圳", "北京"],
    score: 78,
    level: "try",
    risk: "mid-high",
    tags: ["Go", "Kubernetes", "分布式", "容器"],
    reasons: [
      "后端基础技能可迁移到云平台开发",
      "Go 语言与云方向技术栈有重叠",
    ],
    shortfalls: [
      "容器、K8s、云原生部署证据不足",
      "缺少分布式系统实战经历",
    ],
    desc: "负责云平台基础组件开发与优化，参与容器调度、服务治理等核心方向研究与实现。",
    requirements: "了解 Kubernetes 或 Docker；熟悉 Go 语言；有分布式系统背景者优先。",
  },
  {
    rank: 3,
    title: "测试开发实习生",
    type: "技术 · 应届实习",
    locations: ["深圳", "成都"],
    score: 72,
    level: "try",
    risk: "mid",
    tags: ["Python", "接口测试", "自动化测试"],
    reasons: [
      "后端接口和数据库能力可迁移到测试开发",
      "有 Python 基础",
    ],
    shortfalls: [
      "测试工具使用经验不足",
      "缺少自动化测试或 CI/CD 项目经历",
    ],
    desc: "负责测试用例设计与自动化脚本开发，参与 CI/CD 流程搭建，协同保障产品质量。",
    requirements: "了解 Python 或 Java；具备基础测试理论；有自动化测试项目经验者优先。",
  },
  {
    rank: 4,
    title: "数据开发实习生",
    type: "技术 · 应届实习",
    locations: ["北京", "上海"],
    score: 58,
    level: "low",
    risk: "high",
    tags: ["SQL", "Python", "数仓", "ETL"],
    reasons: [
      "数据库基础知识有一定相关性",
    ],
    shortfalls: [
      "缺少数仓、ETL、数据处理项目经历",
      "未提及 Spark、Hive 等大数据工具",
      "数据方向整体证据薄弱",
    ],
    desc: "参与数据平台核心管道建设，负责数据清洗、指标体系搭建及数据质量保障工作。",
    requirements: "熟练使用 SQL；了解 Python；有 Hive / Spark 使用经验者优先。",
  },
];

/* ─────────────── Basis Card ─────────────── */

function BasisCard() {
  const skills = ["Go", "MySQL", "Redis", "Linux", "GoFrame"];
  const directions = ["后台管理系统", "权限管理", "接口开发", "数据查询"];

  return (
    <Card>
      <CardHeader
        title="推荐依据"
        subtitle="V2 · 2026-05-29"
      />
      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "14px 40px" }}>
        <div>
          <div style={{ fontSize: 12, color: color.textTertiary, marginBottom: 7 }}>
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
        <div>
          <div style={{ fontSize: 12, color: color.textTertiary, marginBottom: 7 }}>
            项目方向
          </div>
          <div style={{ display: "flex", flexWrap: "wrap", gap: 5 }}>
            {directions.map((d) => (
              <TextTag key={d} label={d} />
            ))}
          </div>
        </div>
        <div>
          <div style={{ fontSize: 12, color: color.textTertiary, marginBottom: 6 }}>
            当前优势
          </div>
          <div style={{ display: "flex", alignItems: "flex-start", gap: 6 }}>
            <CheckCircle2 size={13} color={color.success} style={{ marginTop: 2, flexShrink: 0 }} />
            <span style={{ fontSize: 13, color: color.textSecondary, lineHeight: 1.6 }}>
              后端技术栈集中，有完整后台项目经历
            </span>
          </div>
        </div>
        <div>
          <div style={{ fontSize: 12, color: color.textTertiary, marginBottom: 6 }}>
            主要短板
          </div>
          <div style={{ display: "flex", flexDirection: "column", gap: 4 }}>
            {["高并发经验无证据", "云原生部署经历不足", "性能优化描述欠缺"].map((s) => (
              <div key={s} style={{ display: "flex", alignItems: "flex-start", gap: 6 }}>
                <AlertCircle size={13} color={color.warn} style={{ marginTop: 2, flexShrink: 0 }} />
                <span style={{ fontSize: 13, color: color.textSecondary }}>{s}</span>
              </div>
            ))}
          </div>
        </div>
      </div>
    </Card>
  );
}

/* ─────────────── Level Legend ─────────────── */

function LevelLegend() {
  return (
    <div
      style={{
        background: "#FAFBFC",
        border: `1px solid ${color.border}`,
        borderRadius: 8,
        padding: "10px 16px",
        display: "flex",
        alignItems: "center",
        gap: 20,
        flexWrap: "wrap",
      }}
    >
      <span style={{ fontSize: 12, color: color.textTertiary, flexShrink: 0 }}>推荐等级：</span>
      {[
        { label: "优先推荐", desc: "简历证据与岗位要求较匹配", status: "success" as const },
        { label: "可以尝试", desc: "方向相关，但存在部分短板", status: "info" as const },
        { label: "暂不优先", desc: "当前证据不足，建议补充后再投", status: "neutral" as const },
      ].map((item) => (
        <div key={item.label} style={{ display: "flex", alignItems: "center", gap: 6 }}>
          <StatusTag status={item.status} label={item.label} dot={false} />
          <span style={{ fontSize: 12, color: color.textTertiary }}>{item.desc}</span>
        </div>
      ))}
    </div>
  );
}

/* ─────────────── Job Detail Drawer ─────────────── */

function DetailDrawer({
  job,
  onClose,
}: {
  job: RecommendJob;
  onClose: () => void;
}) {
  const lvl = LEVEL_MAP[job.level];
  const rsk = RISK_MAP[job.risk];
  return (
    <div
      style={{ position: "fixed", inset: 0, zIndex: 300, display: "flex", alignItems: "flex-start", justifyContent: "flex-end" }}
      onClick={onClose}
    >
      <div style={{ position: "absolute", inset: 0, background: "rgba(0,0,0,0.18)" }} />
      <div
        style={{
          position: "relative",
          width: 440,
          height: "100vh",
          background: "#fff",
          boxShadow: "-4px 0 24px rgba(0,0,0,0.10)",
          overflowY: "auto",
          display: "flex",
          flexDirection: "column",
        }}
        onClick={(e) => e.stopPropagation()}
      >
        <div style={{ padding: "20px 24px 16px", borderBottom: `1px solid ${color.border}`, flexShrink: 0 }}>
          <div style={{ display: "flex", alignItems: "flex-start", justifyContent: "space-between" }}>
            <div>
              <div style={{ display: "flex", alignItems: "center", gap: 8, marginBottom: 5 }}>
                <span style={{ fontSize: 11, fontWeight: 700, color: color.textTertiary, letterSpacing: "0.5px" }}>
                  TOP {job.rank}
                </span>
              </div>
              <div style={{ fontSize: 17, fontWeight: 700, color: color.textPrimary, marginBottom: 5 }}>
                {job.title}
              </div>
              <div style={{ fontSize: 13, color: color.textSecondary }}>{job.type}</div>
            </div>
            <button onClick={onClose} style={{ background: "none", border: "none", cursor: "pointer", padding: 4, color: color.textTertiary }}>
              <X size={18} />
            </button>
          </div>
        </div>
        <div style={{ padding: "20px 24px", flex: 1 }}>
          <div style={{ display: "flex", alignItems: "center", gap: 6, marginBottom: 16 }}>
            <MapPin size={13} color={color.textTertiary} />
            <span style={{ fontSize: 13, color: color.textSecondary }}>{job.locations.join(" / ")}</span>
          </div>
          <div style={{ display: "flex", gap: 6, flexWrap: "wrap", marginBottom: 16 }}>
            {job.tags.map((t) => <TextTag key={t} label={t} />)}
          </div>
          <div style={{ display: "flex", gap: 8, marginBottom: 20 }}>
            <StatusTag status={lvl.status} label={lvl.label} />
            <StatusTag status={rsk.status} label={`初筛风险：${rsk.label}`} />
          </div>
          <CardDivider />
          <div style={{ marginBottom: 16 }}>
            <div style={{ fontSize: 13, fontWeight: 600, color: color.textPrimary, marginBottom: 8 }}>岗位描述</div>
            <p style={{ fontSize: 14, color: color.textSecondary, lineHeight: 1.75 }}>{job.desc}</p>
          </div>
          <div>
            <div style={{ fontSize: 13, fontWeight: 600, color: color.textPrimary, marginBottom: 8 }}>岗位要求</div>
            <p style={{ fontSize: 14, color: color.textSecondary, lineHeight: 1.75 }}>{job.requirements}</p>
          </div>
        </div>
        <div style={{ padding: "16px 24px", borderTop: `1px solid ${color.border}`, flexShrink: 0 }}>
          <Button variant="primary" size="md" style={{ width: "100%" }}>
            进行岗位匹配
          </Button>
        </div>
      </div>
    </div>
  );
}

/* ─────────────── Recommend Card ─────────────── */

function RecommendCard({
  job,
  onDetail,
}: {
  job: RecommendJob;
  onDetail: () => void;
}) {
  const lvl = LEVEL_MAP[job.level];
  const rsk = RISK_MAP[job.risk];

  return (
    <Card hoverable>
      <div style={{ display: "flex", gap: 20, alignItems: "flex-start" }}>
        {/* Rank + Score column */}
        <div
          style={{
            display: "flex",
            flexDirection: "column",
            alignItems: "center",
            gap: 10,
            flexShrink: 0,
            paddingRight: 20,
            borderRight: `1px solid ${color.divider}`,
            minWidth: 80,
          }}
        >
          <div
            style={{
              fontSize: 11,
              fontWeight: 700,
              color: job.rank === 1 ? color.blue : color.textTertiary,
              letterSpacing: "0.5px",
              background: job.rank === 1 ? color.blueLight : "#F2F3F5",
              borderRadius: 4,
              padding: "3px 8px",
            }}
          >
            TOP {job.rank}
          </div>
          <ScoreRing score={job.score} size={60} />
        </div>

        {/* Main content */}
        <div style={{ flex: 1, minWidth: 0 }}>
          {/* Title row */}
          <div style={{ display: "flex", alignItems: "flex-start", justifyContent: "space-between", marginBottom: 8, gap: 12 }}>
            <div>
              <div style={{ fontSize: 16, fontWeight: 700, color: color.textPrimary, marginBottom: 4 }}>
                {job.title}
              </div>
              <div style={{ display: "flex", alignItems: "center", gap: 8, fontSize: 13, color: color.textSecondary }}>
                <span>{job.type}</span>
                <span style={{ color: color.border }}>|</span>
                <span style={{ display: "flex", alignItems: "center", gap: 3 }}>
                  <MapPin size={12} />
                  {job.locations.join(" / ")}
                </span>
              </div>
            </div>
            <div style={{ display: "flex", gap: 6, flexShrink: 0, flexWrap: "wrap", justifyContent: "flex-end" }}>
              <StatusTag status={lvl.status} label={lvl.label} />
              <StatusTag status={rsk.status} label={`风险：${rsk.label}`} />
            </div>
          </div>

          {/* Tags */}
          <div style={{ display: "flex", flexWrap: "wrap", gap: 5, marginBottom: 14 }}>
            {job.tags.map((t) => <TextTag key={t} label={t} />)}
          </div>

          <CardDivider />

          {/* Reasons + Shortfalls */}
          <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "10px 24px", marginTop: 10, marginBottom: 14 }}>
            <div>
              <div style={{ fontSize: 12, fontWeight: 600, color: color.successText, marginBottom: 7, display: "flex", alignItems: "center", gap: 4 }}>
                <div style={{ width: 3, height: 11, borderRadius: 2, background: color.success }} />
                推荐理由
              </div>
              {job.reasons.map((r) => (
                <div key={r} style={{ display: "flex", alignItems: "flex-start", gap: 6, marginBottom: 5 }}>
                  <div style={{ width: 4, height: 4, borderRadius: "50%", background: color.textTertiary, marginTop: 6, flexShrink: 0 }} />
                  <span style={{ fontSize: 13, color: color.textSecondary, lineHeight: 1.6 }}>{r}</span>
                </div>
              ))}
            </div>
            <div>
              <div style={{ fontSize: 12, fontWeight: 600, color: color.warnText, marginBottom: 7, display: "flex", alignItems: "center", gap: 4 }}>
                <div style={{ width: 3, height: 11, borderRadius: 2, background: color.warn }} />
                主要短板
              </div>
              {job.shortfalls.map((s) => (
                <div key={s} style={{ display: "flex", alignItems: "flex-start", gap: 6, marginBottom: 5 }}>
                  <div style={{ width: 4, height: 4, borderRadius: "50%", background: color.textTertiary, marginTop: 6, flexShrink: 0 }} />
                  <span style={{ fontSize: 13, color: color.textSecondary, lineHeight: 1.6 }}>{s}</span>
                </div>
              ))}
            </div>
          </div>

          {/* Actions */}
          <div style={{ display: "flex", gap: 8 }}>
            <Button variant="ghost" size="sm" onClick={onDetail}>
              查看岗位详情
            </Button>
            <Button variant="primary" size="sm">
              进行岗位匹配
            </Button>
          </div>
        </div>
      </div>
    </Card>
  );
}

/* ─────────────── Empty State ─────────────── */

function EmptyState() {
  return (
    <Card style={{ textAlign: "center", padding: "60px 24px" }}>
      <div style={{ width: 52, height: 52, borderRadius: 12, background: "#F2F3F5", display: "flex", alignItems: "center", justifyContent: "center", margin: "0 auto 18px" }}>
        <FileText size={22} color={color.textTertiary} strokeWidth={1.5} />
      </div>
      <div style={{ fontSize: 15, fontWeight: 600, color: color.textPrimary, marginBottom: 8 }}>
        请先添加简历
      </div>
      <div style={{ fontSize: 13, color: color.textTertiary, lineHeight: 1.7, maxWidth: 320, margin: "0 auto 28px" }}>
        添加简历后，系统才能基于你的经历推荐适合的岗位。
      </div>
      <Button variant="primary" size="md">
        添加简历
      </Button>
    </Card>
  );
}

/* ─────────────── Bottom Actions ─────────────── */

function BottomActions() {
  return (
    <div style={{ display: "flex", gap: 10, paddingTop: 4 }}>
      <Button variant="outline" size="md">返回岗位列表</Button>
      <Button variant="ghost" size="md">查看我的简历</Button>
    </div>
  );
}

/* ─────────────── Page ─────────────── */

export function RecommendPage({ hasResume = true }: { hasResume?: boolean }) {
  const [detailJob, setDetailJob] = useState<RecommendJob | null>(null);

  return (
    <PageShell activeNav="jobs">
      <PageHeader
        title="岗位推荐"
        subtitle="基于当前简历版本，为你推荐更适合优先查看的岗位。"
      />

      {hasResume ? (
        <div style={{ display: "flex", flexDirection: "column", gap: 14 }}>
          <BasisCard />
          <LevelLegend />
          {RECOMMENDATIONS.map((job) => (
            <RecommendCard
              key={job.rank}
              job={job}
              onDetail={() => setDetailJob(job)}
            />
          ))}
          <BottomActions />
        </div>
      ) : (
        <EmptyState />
      )}

      {detailJob && (
        <DetailDrawer job={detailJob} onClose={() => setDetailJob(null)} />
      )}
    </PageShell>
  );
}
