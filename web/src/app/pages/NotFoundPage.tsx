import { Link } from "react-router";

/**
 * NotFoundPage — K8-5-46-B：catch-all 404 专页。
 * 替代旧的静默 Navigate to "/"。
 */
export function NotFoundPage() {
  return (
    <div className="flex min-h-[60vh] items-center justify-center px-6">
      <div className="mx-auto max-w-lg text-center">
        <p className="text-6xl font-bold text-slate-200 dark:text-slate-700">404</p>
        <h1 className="mt-4 text-xl font-semibold text-slate-700">页面不存在</h1>
        <p className="mt-2 text-sm text-slate-500">
          你访问的地址不存在，或链接已经失效。
        </p>
        <div className="mt-8 flex items-center justify-center gap-3">
          <Link
            to="/jobs"
            className="inline-flex h-10 items-center rounded-lg bg-[#0052D9] px-6 text-sm font-medium text-white hover:bg-blue-700 transition-colors"
          >
            返回岗位广场
          </Link>
          <Link
            to="/resumes"
            className="inline-flex h-10 items-center rounded-lg border border-slate-200 bg-white px-6 text-sm font-medium text-slate-700 hover:bg-slate-50 transition-colors"
          >
            我的简历
          </Link>
        </div>
      </div>
    </div>
  );
}
