import { FileText, Lightbulb, BarChart2, ChevronRight } from "lucide-react";
import { color } from "../tokens";
import { Card, CardHeader } from "./ui/Card";
import { Button } from "./ui/Button";
import { StatusTag } from "./ui/StatusTag";
import { PageHeader } from "./layout/PageShell";

const ENTRIES = [
  {
    key: "resume",
    icon: FileText,
    title: "添加简历",
    desc: "创建当前简历版本",
    btn: "开始填写",
    variant: "outline" as const,
  },
  {
    key: "recommend",
    icon: Lightbulb,
    title: "岗位推荐",
    desc: "查看更适合的岗位",
    btn: "查看推荐",
    variant: "primary" as const,
  },
  {
    key: "match",
    icon: BarChart2,
    title: "岗位匹配",
    desc: "分析目标岗位匹配度",
    btn: "立即分析",
    variant: "outline" as const,
  },
];

export function HomeContent() {
  return (
    <>
      <PageHeader
        title="Offer 捕手"
        subtitle="基于简历的岗位推荐与岗位匹配 Demo"
      />

      {/* Function entry cards */}
      <div
        style={{
          display: "grid",
          gridTemplateColumns: "repeat(3, 1fr)",
          gap: 16,
          marginBottom: 16,
        }}
      >
        {ENTRIES.map((entry) => {
          const Icon = entry.icon;
          return (
            <Card key={entry.key} hoverable>
              <div
                style={{
                  display: "flex",
                  alignItems: "flex-start",
                  justifyContent: "space-between",
                  marginBottom: 16,
                }}
              >
                <div
                  style={{
                    width: 38,
                    height: 38,
                    borderRadius: 8,
                    background: color.blueLight,
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "center",
                  }}
                >
                  <Icon size={17} color={color.blue} strokeWidth={1.8} />
                </div>
                <ChevronRight size={15} color={color.textDisabled} />
              </div>

              <div
                style={{
                  fontSize: 15,
                  fontWeight: 600,
                  color: color.textPrimary,
                  marginBottom: 5,
                }}
              >
                {entry.title}
              </div>
              <div
                style={{
                  fontSize: 13,
                  color: color.textTertiary,
                  marginBottom: 20,
                }}
              >
                {entry.desc}
              </div>

              <Button variant={entry.variant} size="sm">
                {entry.btn}
              </Button>
            </Card>
          );
        })}
      </div>

      {/* Status card */}
      <Card>
        <CardHeader
          title="最近分析记录"
          extra={
            <a
              href="#"
              style={{
                color: color.blue,
                textDecoration: "none",
                display: "flex",
                alignItems: "center",
                gap: 2,
                fontSize: 13,
              }}
            >
              查看全部 <ChevronRight size={13} />
            </a>
          }
        />

        <div
          style={{
            display: "grid",
            gridTemplateColumns: "repeat(4, 1fr)",
          }}
        >
          {[
            { label: "当前简历版本", content: <span style={{ fontSize: 15, fontWeight: 600, color: color.textPrimary }}>V2</span> },
            {
              label: "最近目标岗位",
              content: (
                <span style={{ fontSize: 15, fontWeight: 600, color: color.textPrimary }}>
                  后端开发实习生
                </span>
              ),
            },
            {
              label: "最近匹配度",
              content: (
                <span style={{ fontSize: 15, fontWeight: 700, color: color.blue }}>
                  84 / 100
                </span>
              ),
            },
            {
              label: "当前状态",
              content: <StatusTag status="warn" label="建议优化后投递" />,
            },
          ].map((item, i) => (
            <div
              key={item.label}
              style={{
                paddingRight: i < 3 ? 24 : 0,
                marginRight: i < 3 ? 24 : 0,
                borderRight: i < 3 ? `1px solid ${color.divider}` : "none",
              }}
            >
              <div
                style={{
                  fontSize: 12,
                  color: color.textTertiary,
                  marginBottom: 8,
                }}
              >
                {item.label}
              </div>
              {item.content}
            </div>
          ))}
        </div>
      </Card>
    </>
  );
}
