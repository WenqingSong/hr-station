import { useState } from "react";
import { FileText, CheckCircle2, AlertTriangle, ChevronRight } from "lucide-react";
import { color } from "../tokens";
import { PageShell, PageHeader } from "../components/layout/PageShell";
import { Card, CardHeader, CardDivider } from "../components/ui/Card";
import { Button } from "../components/ui/Button";
import { TextTag } from "../components/ui/StatusTag";

/* ─────────────── Data ─────────────── */

interface Version {
  id: string;
  date: string;
  job: string;
  score: number;
  status: string;
  statusColor: string;
  statusBg: string;
}

const VERSIONS: Version[] = [
  {
    id: "V1",
    date: "5 月 28 日",
    job: "后端开发实习生",
    score: 72,
    status: "原始版本",
    statusColor: color.textSecondary,
    statusBg: "#F2F3F5",
  },
  {
    id: "V2",
    date: "5 月 28 日",
    job: "后端开发实习生",
    score: 84,
    status: "手动优化",
    statusColor: color.successText,
    statusBg: color.successBg,
  },
  {
    id: "V3",
    date: "5 月 29 日",
    job: "云计算开发实习生",
    score: 78,
    status: "调整方向",
    statusColor: color.blue,
    statusBg: color.blueLight,
  },
];

interface CompareMetric {
  label: string;
  v1: number;
  v2: number;
  unit?: string;
  higherIsBetter: boolean;
}

const ALL_METRICS: CompareMetric[] = [
  { label: "综合匹配度",  v1: 72, v2: 84, unit: "/ 100", higherIsBetter: true },
  { label: "强证据项数",  v1: 2,  v2: 4,  unit: "项",    higherIsBetter: true },
  { label: "弱证据项数",  v1: 4,  v2: 2,  unit: "项",    higherIsBetter: false },
  { label: "高风险项数",  v1: 3,  v2: 1,  unit: "项",    higherIsBetter: false },
];

/* ─────────────── Info Bar ─────────────── */

function InfoBar() {
  return (
    <Card style={{ padding: "14px 20px", marginBottom: 14 }}>
      <div style={{ display: "flex", alignItems: "center", gap: 0, flexWrap: "wrap" }}>
        {[
          {
            label: "最新版本",
            content: (
              <TextTag label="V3" style={{ color: color.blue, background: color.blueLight, fontWeight: 700 }} />
            ),
          },
          { label: "简历方向", content: "后端开发方向" },
          { label: "最近目标岗位", content: "后端开发实习生" },
          {
            label: "最近匹配度",
            content: (
              <span style={{ fontWeight: 700, color: color.blue, fontSize: 15 }}>84 / 100</span>
            ),
          },
        ].map((item, i, arr) => (
          <div
            key={item.label}
            style={{
              display: "flex",
              alignItems: "center",
              gap: 8,
              paddingRight: i < arr.length - 1 ? 24 : 0,
              marginRight: i < arr.length - 1 ? 24 : 0,
              borderRight: i < arr.length - 1 ? `1px solid ${color.border}` : "none",
              padding: "4px 0",
            }}
          >
            <span style={{ fontSize: 12, color: color.textTertiary, whiteSpace: "nowrap" }}>
              {item.label}
            </span>
            <div style={{ fontSize: 14, color: color.textPrimary, fontWeight: 500 }}>
              {item.content}
            </div>
          </div>
        ))}
      </div>
    </Card>
  );
}

/* ─────────────── Version List Table ─────────────── */

function VersionListCard() {
  const cols = ["版本", "添加时间", "最近目标岗位", "最近匹配度", "状态"];
  const widths = ["12%", "16%", "28%", "18%", "18%"];

  return (
    <Card>
      <CardHeader title="版本列表" />
      <div style={{ overflowX: "auto" }}>
        <table style={{ width: "100%", borderCollapse: "collapse" }}>
          <thead>
            <tr style={{ background: "#FAFBFC" }}>
              {cols.map((c, i) => (
                <th
                  key={c}
                  style={{
                    width: widths[i],
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
                  {c}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {VERSIONS.map((v, i) => {
              const isLatest = v.id === "V3";
              const isLast = i === VERSIONS.length - 1;
              return (
                <tr
                  key={v.id}
                  style={{ background: isLatest ? "#FAFBFF" : "#fff" }}
                >
                  {/* Version */}
                  <td style={{ padding: "13px 14px", borderBottom: isLast ? "none" : `1px solid ${color.divider}`, verticalAlign: "middle" }}>
                    <div style={{ display: "flex", alignItems: "center", gap: 6 }}>
                      <TextTag
                        label={v.id}
                        style={{ color: color.blue, background: color.blueLight, fontWeight: 700 }}
                      />
                      {isLatest && (
                        <span style={{ fontSize: 11, color: color.blue, background: color.blueLight, borderRadius: 3, padding: "1px 6px", fontWeight: 500 }}>
                          当前
                        </span>
                      )}
                    </div>
                  </td>
                  {/* Date */}
                  <td style={{ padding: "13px 14px", fontSize: 13, color: color.textSecondary, borderBottom: isLast ? "none" : `1px solid ${color.divider}`, verticalAlign: "middle" }}>
                    {v.date}
                  </td>
                  {/* Job */}
                  <td style={{ padding: "13px 14px", fontSize: 13, color: color.textPrimary, borderBottom: isLast ? "none" : `1px solid ${color.divider}`, verticalAlign: "middle" }}>
                    {v.job}
                  </td>
                  {/* Score */}
                  <td style={{ padding: "13px 14px", borderBottom: isLast ? "none" : `1px solid ${color.divider}`, verticalAlign: "middle" }}>
                    <span style={{ fontSize: 15, fontWeight: 700, color: color.blue }}>
                      {v.score}
                    </span>
                    <span style={{ fontSize: 12, color: color.textTertiary }}> / 100</span>
                  </td>
                  {/* Status */}
                  <td style={{ padding: "13px 14px", borderBottom: isLast ? "none" : `1px solid ${color.divider}`, verticalAlign: "middle" }}>
                    <span
                      style={{
                        fontSize: 12,
                        color: v.statusColor,
                        background: v.statusBg,
                        borderRadius: 4,
                        padding: "3px 8px",
                        fontWeight: 500,
                      }}
                    >
                      {v.status}
                    </span>
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

/* ─────────────── Version Comparison ─────────────── */

const VERSION_IDS = VERSIONS.map((v) => v.id);

function CompareCard() {
  const [left, setLeft] = useState("V1");
  const [right, setRight] = useState("V2");

  const leftData = VERSIONS.find((v) => v.id === left)!;
  const rightData = VERSIONS.find((v) => v.id === right)!;

  function getMetricValue(metric: CompareMetric, versionId: string) {
    if (versionId === "V1") return metric.v1;
    if (versionId === "V2") return metric.v2;
    // V3 — fabricate plausible numbers
    if (metric.label === "综合匹配度") return 78;
    if (metric.label === "强证据项数") return 3;
    if (metric.label === "弱证据项数") return 3;
    if (metric.label === "高风险项数") return 2;
    return metric.v2;
  }

  const selectStyle: React.CSSProperties = {
    height: 30,
    padding: "0 10px",
    fontSize: 13,
    color: color.blue,
    background: color.blueLight,
    border: `1px solid #C0D8FF`,
    borderRadius: 6,
    cursor: "pointer",
    fontWeight: 600,
    appearance: "auto" as const,
    outline: "none",
  };

  return (
    <Card>
      {/* Header with selectors */}
      <div
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          marginBottom: 16,
          paddingBottom: 14,
          borderBottom: `1px solid ${color.divider}`,
          flexWrap: "wrap",
          gap: 10,
        }}
      >
        <span style={{ fontSize: 15, fontWeight: 600, color: color.textPrimary }}>
          版本对比
        </span>
        <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
          <span style={{ fontSize: 12, color: color.textTertiary }}>选择对比版本：</span>
          <select style={selectStyle} value={left} onChange={(e) => setLeft(e.target.value)}>
            {VERSION_IDS.map((id) => (
              <option key={id} value={id}>{id}</option>
            ))}
          </select>
          <span style={{ fontSize: 13, color: color.textTertiary, padding: "0 4px" }}>对比</span>
          <select style={selectStyle} value={right} onChange={(e) => setRight(e.target.value)}>
            {VERSION_IDS.map((id) => (
              <option key={id} value={id}>{id}</option>
            ))}
          </select>
        </div>
      </div>

      {/* Comparison table */}
      <table style={{ width: "100%", borderCollapse: "collapse" }}>
        <thead>
          <tr style={{ background: "#FAFBFC" }}>
            {[
              { label: "指标", width: "30%" },
              { label: `${left} · ${VERSIONS.find((v) => v.id === left)?.status}`, width: "23%" },
              { label: `${right} · ${VERSIONS.find((v) => v.id === right)?.status}`, width: "23%" },
              { label: "变化", width: "24%" },
            ].map((col) => (
              <th
                key={col.label}
                style={{
                  width: col.width,
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
                {col.label}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {ALL_METRICS.map((metric, i) => {
            const lv = getMetricValue(metric, left);
            const rv = getMetricValue(metric, right);
            const diff = rv - lv;
            const isPositive = metric.higherIsBetter ? diff > 0 : diff < 0;
            const isNeutral = diff === 0;
            const changeColor = isNeutral
              ? color.textTertiary
              : isPositive
              ? color.successText
              : color.dangerText;
            const changePrefix = diff > 0 ? "+" : "";
            const isLast = i === ALL_METRICS.length - 1;

            return (
              <tr key={metric.label}>
                <td style={{ padding: "13px 14px", fontSize: 13, fontWeight: 500, color: color.textPrimary, borderBottom: isLast ? "none" : `1px solid ${color.divider}` }}>
                  {metric.label}
                </td>
                <td style={{ padding: "13px 14px", fontSize: 14, color: color.textSecondary, borderBottom: isLast ? "none" : `1px solid ${color.divider}` }}>
                  {lv}
                  {metric.unit && <span style={{ fontSize: 12, color: color.textTertiary, marginLeft: 3 }}>{metric.unit}</span>}
                </td>
                <td style={{ padding: "13px 14px", fontSize: 14, color: color.textSecondary, borderBottom: isLast ? "none" : `1px solid ${color.divider}` }}>
                  {rv}
                  {metric.unit && <span style={{ fontSize: 12, color: color.textTertiary, marginLeft: 3 }}>{metric.unit}</span>}
                </td>
                <td style={{ padding: "13px 14px", borderBottom: isLast ? "none" : `1px solid ${color.divider}` }}>
                  <span
                    style={{
                      fontSize: 14,
                      fontWeight: 700,
                      color: changeColor,
                      background: isNeutral ? "#F2F3F5" : isPositive ? color.successBg : color.dangerBg,
                      borderRadius: 4,
                      padding: "2px 8px",
                    }}
                  >
                    {isNeutral ? "—" : `${changePrefix}${diff}${metric.unit ? " " + metric.unit : ""}`}
                  </span>
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </Card>
  );
}

/* ─────────────── Change Reasons ─────────────── */

const REASONS = [
  "项目经历补充了负责模块和个人贡献的描述，HR 可以清楚判断参与深度。",
  "Redis 从关键词变成了具体业务场景，证据由「弱」升级为「中」。",
  "MySQL 增加了表设计、接口联调和查询优化相关描述，证据强度有所提升。",
];

function ChangeReasonsCard() {
  return (
    <Card>
      <CardHeader title="变化原因" subtitle="V1 → V2" />
      <div
        style={{
          padding: "12px 16px",
          background: color.successBg,
          borderRadius: 8,
          marginBottom: 14,
          display: "flex",
          alignItems: "center",
          gap: 8,
        }}
      >
        <CheckCircle2 size={14} color={color.success} style={{ flexShrink: 0 }} />
        <span style={{ fontSize: 13, color: color.successText, fontWeight: 500 }}>
          本次综合匹配度提升 +12，主要来自以下改进
        </span>
      </div>
      <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
        {REASONS.map((r, i) => (
          <div
            key={i}
            style={{ display: "flex", alignItems: "flex-start", gap: 10 }}
          >
            <span
              style={{
                fontSize: 11,
                fontWeight: 700,
                color: color.successText,
                background: color.successBg,
                borderRadius: 3,
                padding: "2px 6px",
                flexShrink: 0,
                marginTop: 2,
              }}
            >
              {i + 1}
            </span>
            <span style={{ fontSize: 13, color: color.textSecondary, lineHeight: 1.65 }}>
              {r}
            </span>
          </div>
        ))}
      </div>
    </Card>
  );
}

/* ─────────────── Still Needs Improvement ─────────────── */

const IMPROVEMENTS = [
  "高并发经验仍缺少真实证据，建议在有实际经历后再补充。",
  "性能优化数据仍需用户确认，如无真实压测数据，不建议填写具体百分比。",
  "云原生部署经历未体现，如转向云计算岗位，建议补充相关实践。",
];

function ImprovementCard() {
  return (
    <Card>
      <CardHeader title="仍需改进" />
      <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
        {IMPROVEMENTS.map((item, i) => (
          <div
            key={i}
            style={{
              display: "flex",
              alignItems: "flex-start",
              gap: 10,
              padding: "10px 14px",
              background: color.warnBg,
              borderRadius: 8,
              border: `1px solid #FFE4B5`,
            }}
          >
            <AlertTriangle size={14} color={color.warn} style={{ marginTop: 2, flexShrink: 0 }} />
            <span style={{ fontSize: 13, color: color.textSecondary, lineHeight: 1.65 }}>
              {item}
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
      <Button variant="primary"  size="md">添加新版本</Button>
      <Button variant="outline"  size="md">返回我的简历</Button>
      <Button variant="outline"  size="md">去岗位列表继续匹配</Button>
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
      <div style={{ fontSize: 15, fontWeight: 600, color: color.textPrimary, marginBottom: 8 }}>
        当前还没有简历版本
      </div>
      <div
        style={{
          fontSize: 13,
          color: color.textTertiary,
          lineHeight: 1.7,
          maxWidth: 340,
          margin: "0 auto 28px",
        }}
      >
        请先添加第一版简历，系统会基于你的简历内容进行岗位推荐和岗位匹配分析。
      </div>
      <Button variant="primary" size="md">添加简历</Button>
    </Card>
  );
}

/* ─────────────── Page ─────────────── */

export function VersionsPage({ hasVersions = true }: { hasVersions?: boolean }) {
  return (
    <PageShell activeNav="resume">
      <PageHeader
        title="简历版本迭代"
        subtitle="对比不同版本的匹配表现，查看优化是否带来提升。"
      />

      {hasVersions ? (
        <div style={{ display: "flex", flexDirection: "column", gap: 14 }}>
          <InfoBar />
          <VersionListCard />
          <CompareCard />
          <ChangeReasonsCard />
          <ImprovementCard />
          <BottomActions />
        </div>
      ) : (
        <EmptyState />
      )}
    </PageShell>
  );
}
