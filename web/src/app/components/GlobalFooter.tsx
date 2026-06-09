import React from "react";
import { Mail, ExternalLink } from "lucide-react";

export function GlobalFooter() {
  return (
    <footer className="border-t border-slate-200/70 bg-slate-50/80">
      <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8 md:py-10">
        <div className="flex flex-col gap-6 md:flex-row md:justify-between md:items-start">
          {/* Left: brand + contact */}
          <div className="flex-1 min-w-0">
            <p className="text-sm font-semibold text-slate-800">
              Offer捕手 · AI 求职智能体
            </p>
            <p className="mt-1.5 max-w-md text-[13px] leading-relaxed text-slate-500">
              帮助学生更快找到适合岗位，更准优化简历，提高投递命中率。
            </p>
            <a
              href="mailto:1784870640@qq.com"
              className="mt-3 inline-flex items-center gap-1.5 text-[13px] text-slate-500 transition-colors hover:text-slate-700"
            >
              <Mail className="h-3.5 w-3.5" />
              1784870640@qq.com
            </a>
          </div>

          {/* Right: feedback */}
          <div className="flex-shrink-0">
            <p className="text-[13px] text-slate-500 leading-relaxed max-w-xs">
              如果你对这个产品原型有使用建议，或希望帮助我们继续优化体验，欢迎填写体验反馈问卷。感谢你的体验与支持。
            </p>
            <a
              href="https://docs.qq.com/form/page/DREp0UFRmZ0NVVXFt"
              target="_blank"
              rel="noopener noreferrer"
              className="mt-3 inline-flex items-center gap-1.5 rounded-full border border-blue-200 bg-blue-50 px-4 py-2 text-sm font-semibold text-blue-700 transition hover:bg-blue-100"
            >
              填写体验反馈问卷
              <ExternalLink className="h-3.5 w-3.5" />
            </a>
          </div>
        </div>

        {/* Bottom: copyright */}
        <div className="mt-8 border-t border-slate-200/50 pt-5 text-center text-xs text-slate-400">
          © 2026 weniqngsongceshi.online. 保留所有权利，侵权必究。
        </div>
      </div>
    </footer>
  );
}
