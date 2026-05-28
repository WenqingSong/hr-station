import { useState } from "react";
import { Search, MapPin, X, ChevronRight } from "lucide-react";
import { color } from "../tokens";
import { PageShell, PageHeader } from "../components/layout/PageShell";
import { Card, CardHeader, CardDivider } from "../components/ui/Card";
import { Button } from "../components/ui/Button";
import { StatusTag, TextTag } from "../components/ui/StatusTag";

/* ─────────────── Data ─────────────── */

export interface Job {
  id: number;
  title: string;
  type: string;
  locations: string[];
  tags: string[];
  summary: string;
  desc: string;
  requirements: string;
}

const JOBS: Job[] = [
  {
    id: 1,
    title: "后端开发实习生",
    type: "技术 · 应届实习",
    locations: ["深圳", "北京", "上海"],
    tags: ["Go", "MySQL", "Redis", "后台系统", "接口开发"],
    summary: "参与后台系统开发、接口设计与业务联调，要求熟悉至少一门后端语言，了解数据库和缓存。",
    desc: "负责核心后台服务的设计与开发，参与系统架构优化，与产品、前端协作完成业务迭代。",
    requirements: "熟悉 Go / Java / Python 等后端语言之一；了解 MySQL、Redis 等常见中间件；有实际项目经验优先。",
  },
  {
    id: 2,
    title: "云计算开发实习生",
    type: "技术 · 应届实习",
    locations: ["深圳", "北京"],
    tags: ["Kubernetes", "Go", "分布式", "微服务", "容器"],
    summary: "参与云原生平台核心模块研发，要求了解 K8s 或容器技术，有 Go 开发经验者优先。",
    desc: "负责云平台基础组件开发与优化，参与容器调度、服务治理等核心方向研究与实现。",
    requirements: "了解 Kubernetes 或 Docker；熟悉 Go 语言；有分布式系统背景者优先。",
  },
  {
    id: 3,
    title: "数据开发实习生",
    type: "技术 · 应届实习",
    locations: ["北京", "上海"],
    tags: ["Python", "Spark", "Hive", "SQL", "数据仓库"],
    summary: "负责数据仓库建设与 ETL 流程开发，要求熟悉 SQL，了解 Spark 或 Hive 等大数据框架。",
    desc: "参与数据平台核心管道建设，负责数据清洗、指标体系搭建及数据质量保障工作。",
    requirements: "熟练使用 SQL；了解 Python；有 Hive / Spark 使用经验者优先。",
  },
  {
    id: 4,
    title: "算法实习生",
    type: "技术 · 应届实习",
    locations: ["深圳", "北京"],
    tags: ["Python", "PyTorch", "NLP", "机器学习", "模型训练"],
    summary: "参与推荐 / NLP / CV 等方向模型研究与工程化落地，要求有深度学习项目经验。",
    desc: "负责算法模型训练与调优，参与工程化部署，与业务团队联合推进效果提升。",
    requirements: "熟悉 Python；了解 PyTorch 或 TensorFlow；有完整 ML 项目经历者优先。",
  },
  {
    id: 5,
    title: "产品经理实习生",
    type: "产品 · 应届实习",
    locations: ["深圳", "上海"],
    tags: ["产品设计", "竞品分析", "需求文档", "用户调研", "原型"],
    summary: "参与产品需求调研与方案设计，撰写 PRD，协助产品迭代推进，有互联网产品实习经验者优先。",
    desc: "协助产品经理完成需求分析、产品方案设计与跨团队协作推进，参与产品数据分析。",
    requirements: "逻辑清晰，善于沟通；熟悉 Axure / Figma 等原型工具；有产品实习经历优先。",
  },
  {
    id: 6,
    title: "测试开发实习生",
    type: "技术 · 应届实习",
    locations: ["深圳", "成都"],
    tags: ["Python", "Java", "自动化测试", "CI/CD", "接口测试"],
    summary: "参与自动化测试框架建设与用例设计，要求了解测试基础，有编程能力，具备质量保障意识。",
    desc: "负责测试用例设计与自动化脚本开发，参与 CI/CD 流程搭建，协同保障产品质量。",
    requirements: "了解 Python 或 Java；具备基础测试理论；有自动化测试项目经验者优先。",
  },
];

/* ─────────────── Job Detail Modal ─────────────── */

function JobDetailModal({
  job,
  onClose,
  onSelect,
  isSelected,
}: {
  job: Job;
  onClose: () => void;
  onSelect: (job: Job) => void;
  isSelected: boolean;
}) {
  return (
    <div
      style={{
        position: "fixed",
        inset: 0,
        zIndex: 300,
        display: "flex",
        alignItems: "flex-start",
        justifyContent: "flex-end",
      }}
      onClick={onClose}
    >
      {/* Overlay */}
      <div
        style={{
          position: "absolute",
          inset: 0,
          background: "rgba(0,0,0,0.18)",
        }}
      />
      {/* Drawer */}
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
        {/* Drawer header */}
        <div
          style={{
            padding: "20px 24px 16px",
            borderBottom: `1px solid ${color.border}`,
            display: "flex",
            alignItems: "flex-start",
            justifyContent: "space-between",
            gap: 12,
            flexShrink: 0,
          }}
        >
          <div>
            <div style={{ fontSize: 17, fontWeight: 700, color: color.textPrimary, marginBottom: 5 }}>
              {job.title}
            </div>
            <div style={{ fontSize: 13, color: color.textSecondary }}>
              {job.type}
            </div>
          </div>
          <button
            onClick={onClose}
            style={{
              background: "none",
              border: "none",
              cursor: "pointer",
              padding: 4,
              color: color.textTertiary,
              flexShrink: 0,
            }}
          >
            <X size={18} />
          </button>
        </div>

        {/* Drawer body */}
        <div style={{ padding: "20px 24px", flex: 1 }}>
          {/* Locations */}
          <div style={{ display: "flex", alignItems: "center", gap: 6, marginBottom: 16 }}>
            <MapPin size={13} color={color.textTertiary} />
            <span style={{ fontSize: 13, color: color.textSecondary }}>
              {job.locations.join(" / ")}
            </span>
          </div>

          {/* Tags */}
          <div style={{ display: "flex", flexWrap: "wrap", gap: 6, marginBottom: 20 }}>
            {job.tags.map((t) => (
              <TextTag key={t} label={t} />
            ))}
          </div>

          <CardDivider />

          <div style={{ marginBottom: 16 }}>
            <div style={{ fontSize: 13, fontWeight: 600, color: color.textPrimary, marginBottom: 8 }}>
              岗位描述
            </div>
            <p style={{ fontSize: 14, color: color.textSecondary, lineHeight: 1.75 }}>
              {job.desc}
            </p>
          </div>

          <div>
            <div style={{ fontSize: 13, fontWeight: 600, color: color.textPrimary, marginBottom: 8 }}>
              岗位要求
            </div>
            <p style={{ fontSize: 14, color: color.textSecondary, lineHeight: 1.75 }}>
              {job.requirements}
            </p>
          </div>
        </div>

        {/* Drawer footer */}
        <div
          style={{
            padding: "16px 24px",
            borderTop: `1px solid ${color.border}`,
            flexShrink: 0,
          }}
        >
          <Button
            variant={isSelected ? "outline" : "primary"}
            size="md"
            onClick={() => { onSelect(job); onClose(); }}
            style={{ width: "100%" }}
          >
            {isSelected ? "已选择此岗位" : "选择该岗位"}
          </Button>
        </div>
      </div>
    </div>
  );
}

/* ─────────────── Job Card ─────────────── */

function JobCard({
  job,
  isSelected,
  onDetail,
  onSelect,
}: {
  job: Job;
  isSelected: boolean;
  onDetail: () => void;
  onSelect: () => void;
}) {
  return (
    <div
      style={{
        background: "#fff",
        borderRadius: 10,
        border: `1px solid ${isSelected ? color.blue : color.border}`,
        padding: "18px 20px",
        transition: "border-color 0.15s, box-shadow 0.15s",
        cursor: "default",
      }}
      onMouseEnter={(e) => {
        if (!isSelected) (e.currentTarget as HTMLDivElement).style.borderColor = "#C0D8FF";
        (e.currentTarget as HTMLDivElement).style.boxShadow = "0 2px 12px rgba(0,0,60,0.07)";
      }}
      onMouseLeave={(e) => {
        if (!isSelected) (e.currentTarget as HTMLDivElement).style.borderColor = color.border;
        (e.currentTarget as HTMLDivElement).style.boxShadow = "none";
      }}
    >
      {/* Top row */}
      <div
        style={{
          display: "flex",
          alignItems: "flex-start",
          justifyContent: "space-between",
          marginBottom: 8,
          gap: 12,
        }}
      >
        <div style={{ flex: 1 }}>
          <div style={{ display: "flex", alignItems: "center", gap: 8, marginBottom: 4 }}>
            <span style={{ fontSize: 15, fontWeight: 700, color: color.textPrimary }}>
              {job.title}
            </span>
            {isSelected && <StatusTag status="info" label="已选择" dot={false} />}
          </div>
          <div
            style={{
              display: "flex",
              alignItems: "center",
              gap: 8,
              fontSize: 13,
              color: color.textSecondary,
            }}
          >
            <span>{job.type}</span>
            <span style={{ color: color.border }}>|</span>
            <span style={{ display: "flex", alignItems: "center", gap: 3 }}>
              <MapPin size={12} />
              {job.locations.join(" / ")}
            </span>
          </div>
        </div>
        <ChevronRight size={16} color={color.textDisabled} style={{ marginTop: 2, flexShrink: 0 }} />
      </div>

      {/* Tags */}
      <div style={{ display: "flex", flexWrap: "wrap", gap: 5, marginBottom: 10 }}>
        {job.tags.map((t) => (
          <TextTag key={t} label={t} />
        ))}
      </div>

      {/* Summary */}
      <p
        style={{
          fontSize: 13,
          color: color.textSecondary,
          lineHeight: 1.65,
          marginBottom: 14,
          display: "-webkit-box",
          WebkitLineClamp: 2,
          WebkitBoxOrient: "vertical",
          overflow: "hidden",
        }}
      >
        {job.summary}
      </p>

      {/* Actions */}
      <div
        style={{
          display: "flex",
          gap: 8,
          paddingTop: 12,
          borderTop: `1px solid ${color.divider}`,
        }}
      >
        <Button variant="ghost" size="sm" onClick={onDetail}>
          查看详情
        </Button>
        <Button
          variant={isSelected ? "outline" : "outline"}
          size="sm"
          onClick={onSelect}
          style={
            isSelected
              ? { color: color.blue, borderColor: color.blue, background: color.blueLight }
              : {}
          }
        >
          {isSelected ? "已选择" : "选择该岗位"}
        </Button>
      </div>
    </div>
  );
}

/* ─────────────── Right Panel ─────────────── */

type PanelState = "no-resume" | "no-job" | "ready";

function RightPanel({
  state,
  selectedJob,
}: {
  state: PanelState;
  selectedJob: Job | null;
}) {
  const hasResume = state !== "no-resume";
  const hasJob = state === "ready" && selectedJob !== null;

  return (
    <div style={{ width: 260, flexShrink: 0 }}>
      <Card style={{ padding: "18px 20px" }}>
        {/* Resume section */}
        <div style={{ marginBottom: 16 }}>
          <div style={{ fontSize: 12, color: color.textTertiary, marginBottom: 6 }}>
            当前简历
          </div>
          {hasResume ? (
            <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
              <TextTag
                label="V2"
                style={{ color: color.blue, background: color.blueLight, fontWeight: 600 }}
              />
              <span style={{ fontSize: 13, color: color.textSecondary }}>简历已就绪</span>
            </div>
          ) : (
            <div style={{ display: "flex", alignItems: "center", gap: 6 }}>
              <span style={{ fontSize: 13, color: color.danger }}>暂无简历</span>
            </div>
          )}
        </div>

        <CardDivider />

        {/* Job section */}
        <div style={{ marginBottom: 16 }}>
          <div style={{ fontSize: 12, color: color.textTertiary, marginBottom: 6 }}>
            当前选择岗位
          </div>
          {hasJob ? (
            <div style={{ fontSize: 14, fontWeight: 600, color: color.textPrimary }}>
              {selectedJob!.title}
            </div>
          ) : (
            <span style={{ fontSize: 13, color: color.textTertiary }}>尚未选择岗位</span>
          )}
        </div>

        <CardDivider />

        {/* Action buttons */}
        <div style={{ display: "flex", flexDirection: "column", gap: 10, marginTop: 4 }}>
          {/* 岗位推荐 */}
          <div>
            <Button
              variant="primary"
              size="md"
              disabled={!hasResume}
              style={{ width: "100%", marginBottom: 4 }}
            >
              岗位推荐
            </Button>
            <p style={{ fontSize: 11, color: color.textTertiary, lineHeight: 1.55, margin: 0 }}>
              {!hasResume
                ? "请先添加简历后使用"
                : "基于当前简历，查看更适合优先投递的岗位"}
            </p>
          </div>

          {/* 岗位匹配 */}
          <div>
            <Button
              variant="outline"
              size="md"
              disabled={!hasResume || !hasJob}
              style={{ width: "100%", marginBottom: 4 }}
            >
              岗位匹配
            </Button>
            <p style={{ fontSize: 11, color: color.textTertiary, lineHeight: 1.55, margin: 0 }}>
              {!hasResume
                ? "请先添加简历后使用"
                : !hasJob
                ? "请选择一个岗位后再进行匹配"
                : "分析匹配度、证据链与初筛风险"}
            </p>
          </div>
        </div>
      </Card>
    </div>
  );
}

/* ─────────────── Search Bar ─────────────── */

function SearchBar({
  query,
  onQuery,
  total,
  resumeVersion,
}: {
  query: string;
  onQuery: (q: string) => void;
  total: number;
  resumeVersion: string | null;
}) {
  return (
    <Card style={{ padding: "14px 20px", marginBottom: 14 }}>
      <div
        style={{
          display: "flex",
          alignItems: "center",
          gap: 10,
          flexWrap: "wrap",
        }}
      >
        {/* Search input */}
        <div style={{ flex: 1, position: "relative", minWidth: 200 }}>
          <Search
            size={14}
            color={color.textTertiary}
            style={{ position: "absolute", left: 10, top: "50%", transform: "translateY(-50%)" }}
          />
          <input
            style={{
              width: "100%",
              height: 34,
              padding: "0 12px 0 32px",
              fontSize: 13,
              color: color.textPrimary,
              background: color.pageBg,
              border: `1px solid ${color.border}`,
              borderRadius: 6,
              outline: "none",
              boxSizing: "border-box",
            }}
            placeholder="搜索岗位名称 / 技能关键词"
            value={query}
            onChange={(e) => onQuery(e.target.value)}
          />
        </div>
        <Button variant="primary" size="sm">
          搜索
        </Button>

        {/* Divider */}
        <div style={{ width: 1, height: 20, background: color.border }} />

        {/* Resume version */}
        <span style={{ fontSize: 13, color: color.textSecondary, whiteSpace: "nowrap" }}>
          用于分析的简历：
          {resumeVersion ? (
            <TextTag
              label={resumeVersion}
              style={{ color: color.blue, background: color.blueLight, fontWeight: 600, marginLeft: 4 }}
            />
          ) : (
            <span style={{ color: color.danger, marginLeft: 4 }}>暂无简历</span>
          )}
        </span>

        {/* Result count */}
        <span style={{ fontSize: 13, color: color.textTertiary, marginLeft: "auto" }}>
          共 {total} 个岗位
        </span>
      </div>
    </Card>
  );
}

/* ─────────────── State Switcher (demo) ─────────────── */

function StateSwitcher({
  state,
  onChange,
}: {
  state: PanelState;
  onChange: (s: PanelState) => void;
}) {
  const options: { key: PanelState; label: string }[] = [
    { key: "no-resume", label: "无简历" },
    { key: "no-job", label: "有简历·未选岗" },
    { key: "ready", label: "有简历·已选岗" },
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
      <span style={{ fontSize: 11, color: color.textTertiary, paddingRight: 4 }}>右栏状态：</span>
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

export function JobsPage() {
  const [query, setQuery] = useState("");
  const [selectedJob, setSelectedJob] = useState<Job | null>(JOBS[0]);
  const [detailJob, setDetailJob] = useState<Job | null>(null);
  const [panelState, setPanelState] = useState<PanelState>("ready");

  const filtered = JOBS.filter(
    (j) =>
      !query ||
      j.title.includes(query) ||
      j.tags.some((t) => t.includes(query))
  );

  const resumeVersion = panelState === "no-resume" ? null : "V2";
  const effectiveSelected = panelState === "ready" ? selectedJob : null;

  return (
    <PageShell activeNav="jobs">
      <PageHeader
        title="岗位列表"
        subtitle="浏览岗位并选择目标岗位，进行后续匹配分析。"
      />

      {/* Demo state switcher */}
      <StateSwitcher state={panelState} onChange={setPanelState} />

      {/* Search bar */}
      <SearchBar
        query={query}
        onQuery={setQuery}
        total={filtered.length}
        resumeVersion={resumeVersion}
      />

      {/* Main layout */}
      <div style={{ display: "flex", gap: 16, alignItems: "flex-start" }}>
        {/* Left: job list */}
        <div style={{ flex: 1, display: "flex", flexDirection: "column", gap: 10 }}>
          {filtered.length === 0 ? (
            <Card style={{ textAlign: "center", padding: "48px 24px" }}>
              <div style={{ fontSize: 14, color: color.textTertiary }}>
                未找到匹配的岗位，请尝试其他关键词
              </div>
            </Card>
          ) : (
            filtered.map((job) => (
              <JobCard
                key={job.id}
                job={job}
                isSelected={effectiveSelected?.id === job.id}
                onDetail={() => setDetailJob(job)}
                onSelect={() => {
                  setSelectedJob(job);
                  if (panelState === "no-job") setPanelState("ready");
                }}
              />
            ))
          )}
        </div>

        {/* Right: operation panel */}
        <RightPanel state={panelState} selectedJob={effectiveSelected} />
      </div>

      {/* Detail drawer */}
      {detailJob && (
        <JobDetailModal
          job={detailJob}
          onClose={() => setDetailJob(null)}
          onSelect={(job) => {
            setSelectedJob(job);
            if (panelState === "no-job") setPanelState("ready");
          }}
          isSelected={effectiveSelected?.id === detailJob.id}
        />
      )}
    </PageShell>
  );
}
