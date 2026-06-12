import React from "react";
import { Link } from "react-router";
import { AlertTriangle, RefreshCw } from "lucide-react";

interface ErrorBoundaryProps {
  children: React.ReactNode;
  /** 可选：错误发生时的回退页面路径，默认 redirect 到首页 */
  fallbackPath?: string;
}

interface ErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
}

/**
 * K13-Stability-A: React Error Boundary
 *
 * 防止任何页面渲染异常导致整个组件树卸载出现白屏。
 * 捕获异常后显示友好的错误提示和返回/重试按钮。
 */
export class ErrorBoundary extends React.Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.groupCollapsed("[ErrorBoundary] Render error");
    console.error("error:", error);
    console.info("componentStack:", errorInfo?.componentStack);
    console.groupEnd();
  }

  handleRetry = () => {
    this.setState({ hasError: false, error: null });
  };

  render() {
    if (this.state.hasError) {
      return (
        <div className="min-h-screen bg-[#F5F7FB] pt-24 pb-12 px-4 sm:px-6 lg:px-8">
          <div className="max-w-lg mx-auto py-20 text-center">
            <div className="w-16 h-16 mx-auto mb-5 rounded-2xl bg-amber-50 flex items-center justify-center">
              <AlertTriangle className="w-7 h-7 text-amber-600" />
            </div>
            <h2 className="text-lg font-semibold text-slate-800 mb-2">
              页面渲染异常
            </h2>
            <p className="text-sm text-slate-500 mb-2 leading-relaxed">
              页面在渲染过程中出现异常。你可以重试，或返回我的简历后重新进入推荐页。
            </p>
            {this.state.error && (
              <p className="text-xs text-slate-400 mb-6 font-mono truncate max-w-full">
                {this.state.error.message}
              </p>
            )}
            <div className="flex items-center justify-center gap-3 flex-wrap">
              <button
                onClick={this.handleRetry}
                className="inline-flex items-center gap-2 px-5 py-2.5 bg-[#0052D9] text-white text-sm font-medium rounded-lg hover:bg-[#003DA5] transition-colors"
              >
                <RefreshCw className="w-4 h-4" />
                重试
              </button>
              <Link
                to={this.props.fallbackPath ?? "/resumes"}
                className="inline-flex items-center gap-1.5 px-4 py-2.5 text-[13px] text-slate-500 border border-slate-200 rounded-lg hover:bg-slate-50 transition-colors"
              >
                返回我的简历
              </Link>
            </div>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}
