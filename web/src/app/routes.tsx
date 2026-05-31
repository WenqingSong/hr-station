import { Navigate, Outlet, createBrowserRouter } from "react-router";
import { Navbar } from "./components/Navbar";
import { JobListPage } from "./pages/JobListPage";
import { JobDetailPage } from "./pages/JobDetailPage";
import { JobRecommendationPage } from "./pages/JobRecommendationPage";
import { JobMatchAnalysisPage } from "./pages/JobMatchAnalysisPage";
import { ResumeOptimizationPage } from "./pages/ResumeOptimizationPage";
import { MyResumesPage } from "./pages/MyResumesPage";
import { ResumeManagementPage } from "./pages/ResumeManagementPage";
import { ResumeComparisonPage } from "./pages/ResumeComparisonPage";

function RootLayout() {
  return (
    <div className="flex min-h-screen flex-col bg-[#F5F7FA] font-sans text-slate-900">
      <Navbar />
      <main className="flex-1">
        <Outlet />
      </main>
    </div>
  );
}

export const router = createBrowserRouter([
  {
    path: "/",
    Component: RootLayout,
    children: [
      { index: true, Component: JobListPage },
      { path: "jobs", Component: JobListPage },
      // ── P0 核心路由（5-3.2 已规范统一） ──
      { path: "jobs/:id", Component: JobDetailPage },
      { path: "recommend", Component: JobRecommendationPage },
      { path: "match", Component: JobMatchAnalysisPage },
      { path: "resume-optimization", Component: ResumeOptimizationPage },
      { path: "my-resumes", Component: ResumeManagementPage },
      { path: "my-resumes/edit", Component: MyResumesPage },
      { path: "my-resumes/compare", Component: ResumeComparisonPage },
      // ── 旧路径兼容重定向（5-3.2 保留，后续清理阶段可移除） ──
      { path: "job/:id", Component: JobDetailPage },
      { path: "recommendations", element: <Navigate to="/recommend" replace /> },
      { path: "match-analysis", element: <Navigate to="/match" replace /> },
      { path: "resume-comparison", element: <Navigate to="/my-resumes/compare" replace /> },
      { path: "*", element: <Navigate to="/" replace /> },
    ],
  },
]);
