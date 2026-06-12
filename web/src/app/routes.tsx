import { Navigate, Outlet, createBrowserRouter } from "react-router";
import { Navbar } from "./components/Navbar";
import { GlobalFooter } from "./components/GlobalFooter";
import { RequireAuth } from "./components/auth/RequireAuth";
import { RouteProgressBar } from "./components/common/RouteProgressBar";
import { ErrorBoundary } from "./components/common/ErrorBoundary";
import { JobListPage } from "./pages/JobListPage";
import { JobDetailPage } from "./pages/JobDetailPage";
import { JobRecommendationPage } from "./pages/JobRecommendationPage";
import { JobMatchAnalysisPage } from "./pages/JobMatchAnalysisPage";
import { ResumeOptimizationPage } from "./pages/ResumeOptimizationPage";
import { MyResumesPage } from "./pages/MyResumesPage";
import { ResumeManagementPage } from "./pages/ResumeManagementPage";
import { ResumeComparisonPage } from "./pages/ResumeComparisonPage";
import { ResumeDetailPage } from "./pages/ResumeDetailPage";
import { LoginPage } from "./pages/LoginPage";
import { RegisterPage } from "./pages/RegisterPage";
import { NotFoundPage } from "./pages/NotFoundPage";

/**
 * K13-Stability-A: 为四个核心页面包裹 ErrorBoundary。
 * 防止运行时异常导致整个页面白屏。
 * 异常时显示友好错误提示和返回按钮，不破坏 Navbar/Footer 结构。
 */
function CorePageWrapper({ children }: { children: React.ReactNode }) {
  return <ErrorBoundary fallbackPath="/resumes">{children}</ErrorBoundary>;
}

function RootLayout() {
  return (
    <div className="flex min-h-screen flex-col bg-[#F5F7FB] font-sans text-slate-900">
      <RouteProgressBar />
      <Navbar />
      <main className="flex-1">
        <Outlet />
      </main>
      <GlobalFooter />
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
      // ── P0 AI 模块（K8-5-16-D：RequireAuth 保护 + K13 ErrorBoundary） ──
      { path: "recommend", element: <RequireAuth><CorePageWrapper><JobRecommendationPage /></CorePageWrapper></RequireAuth> },
      { path: "match", element: <RequireAuth><CorePageWrapper><JobMatchAnalysisPage /></CorePageWrapper></RequireAuth> },
      { path: "resume-optimization", element: <RequireAuth><CorePageWrapper><ResumeOptimizationPage /></CorePageWrapper></RequireAuth> },
      // ── P1 简历模块（K8-5-16-C：RequireAuth 保护 + K13 ErrorBoundary） ──
      { path: "resumes", element: <RequireAuth><ResumeManagementPage /></RequireAuth> },
      { path: "resumes/edit", element: <RequireAuth><MyResumesPage /></RequireAuth> },
      { path: "resumes/new", element: <RequireAuth><MyResumesPage /></RequireAuth> },
      { path: "resumes/compare", element: <RequireAuth><CorePageWrapper><ResumeComparisonPage /></CorePageWrapper></RequireAuth> },
      { path: "resumes/detail/:id", element: <RequireAuth><ResumeDetailPage /></RequireAuth> },
      // ── Auth 页面（8-2 新增） ──
      { path: "login", Component: LoginPage },
      { path: "register", Component: RegisterPage },
      // ── 旧路径兼容重定向（5-3.2/5-3.2-R 保留，后续清理阶段可移除） ──
      { path: "job/:id", Component: JobDetailPage },
      { path: "recommendations", element: <Navigate to="/recommend" replace /> },
      { path: "match-analysis", element: <Navigate to="/match" replace /> },
      { path: "my-resumes", element: <Navigate to="/resumes" replace /> },
      { path: "my-resumes/edit", element: <Navigate to="/resumes/edit" replace /> },
      { path: "my-resumes/new", element: <Navigate to="/resumes/new" replace /> },
      { path: "my-resumes/compare", element: <Navigate to="/resumes/compare" replace /> },
      { path: "resume-comparison", element: <Navigate to="/resumes/compare" replace /> },
      { path: "*", Component: NotFoundPage },
    ],
  },
]);
