import { useState } from "react";
import { PageShell } from "./components/layout/PageShell";
import { HomeContent } from "./components/HomeContent";
import { ResumePage } from "./pages/ResumePage";
import { AddResumePage } from "./pages/AddResumePage";
import { ResumeDetailPage } from "./pages/ResumeDetailPage";
import { JobsPage } from "./pages/JobsPage";
import { RecommendPage } from "./pages/RecommendPage";
import { MatchPage } from "./pages/MatchPage";
import { VersionsPage } from "./pages/VersionsPage";
import { LoginPage } from "./pages/LoginPage";
import { color } from "./tokens";

type Page =
  | "home"
  | "resume"
  | "resume-empty"
  | "resume-add"
  | "resume-add-first"
  | "resume-detail"
  | "resume-detail-empty"
  | "jobs"
  | "recommend"
  | "recommend-empty"
  | "match"
  | "versions"
  | "versions-empty"
  | "login";

const PAGE_OPTIONS: { key: Page; label: string }[] = [
  { key: "home",                label: "首页" },
  { key: "resume",              label: "我的简历" },
  { key: "resume-empty",        label: "简历（空）" },
  { key: "resume-add",          label: "添加简历" },
  { key: "resume-add-first",    label: "添加简历（首次）" },
  { key: "resume-detail",       label: "简历详情" },
  { key: "resume-detail-empty", label: "详情（无简历）" },
  { key: "jobs",                label: "岗位列表" },
  { key: "recommend",           label: "岗位推荐" },
  { key: "recommend-empty",     label: "推荐（无简历）" },
  { key: "match",               label: "岗位匹配分析" },
  { key: "versions",            label: "版本迭代" },
  { key: "versions-empty",      label: "迭代（空）" },
  { key: "login",               label: "登录页" },
];

export default function App() {
  const [page, setPage] = useState<Page>("resume-detail");

  return (
    <>
      {/* Demo switcher */}
      <div
        style={{
          position: "fixed",
          bottom: 20,
          right: 20,
          zIndex: 999,
          background: "#1F2329",
          borderRadius: 10,
          padding: "10px 14px",
          display: "flex",
          alignItems: "center",
          gap: 6,
          boxShadow: "0 4px 20px rgba(0,0,0,0.25)",
          flexWrap: "wrap",
          maxWidth: 660,
        }}
      >
        <span
          style={{
            fontSize: 11,
            color: "#86909C",
            lineHeight: "26px",
            paddingRight: 4,
            whiteSpace: "nowrap",
          }}
        >
          预览：
        </span>
        {PAGE_OPTIONS.map((p) => (
          <button
            key={p.key}
            onClick={() => setPage(p.key)}
            style={{
              height: 26,
              padding: "0 10px",
              fontSize: 11,
              fontWeight: 500,
              borderRadius: 5,
              border: "none",
              cursor: "pointer",
              background: page === p.key ? color.blue : "#2D3139",
              color: page === p.key ? "#fff" : "#86909C",
              transition: "all 0.15s",
              whiteSpace: "nowrap",
            }}
          >
            {p.label}
          </button>
        ))}
      </div>

      {page === "home"                && <PageShell activeNav="home"><HomeContent /></PageShell>}
      {page === "resume"              && <ResumePage hasResume={true} />}
      {page === "resume-empty"        && <ResumePage hasResume={false} />}
      {page === "resume-add"          && <AddResumePage isFirstVersion={false} />}
      {page === "resume-add-first"    && <AddResumePage isFirstVersion={true} />}
      {page === "resume-detail"       && <ResumeDetailPage hasResume={true} />}
      {page === "resume-detail-empty" && <ResumeDetailPage hasResume={false} />}
      {page === "jobs"                && <JobsPage />}
      {page === "recommend"           && <RecommendPage hasResume={true} />}
      {page === "recommend-empty"     && <RecommendPage hasResume={false} />}
      {page === "match"               && <MatchPage />}
      {page === "versions"            && <VersionsPage hasVersions={true} />}
      {page === "versions-empty"      && <VersionsPage hasVersions={false} />}
      {page === "login"               && <LoginPage />}
    </>
  );
}
