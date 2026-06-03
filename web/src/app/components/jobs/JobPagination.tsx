import React from "react";
import { ChevronLeft, ChevronRight } from "lucide-react";

interface JobPaginationProps {
  currentPage: number;
  totalPages: number;
  onPageChange: (page: number) => void;
}

/**
 * 生成腾讯招聘风格的分页页码数组
 * 规则：
 * 1. totalPages <= 7: 全部展示
 * 2. 当前页靠前: 1 2 3 4 5 6 ... last
 * 3. 当前页在中间: 1 ... left cur right ... last
 * 4. 当前页靠后: 1 ... last-5 last-4 last-3 last-2 last-1 last
 */
function generatePaginationPages(current: number, total: number): (number | string)[] {
  if (total <= 7) {
    return Array.from({ length: total }, (_, i) => i + 1);
  }

  const pages: (number | string)[] = [];

  if (current <= 4) {
    // 当前页靠前：1 2 3 4 5 6 ... last
    for (let i = 1; i <= 6; i++) {
      pages.push(i);
    }
    pages.push('...');
    pages.push(total);
  } else if (current >= total - 3) {
    // 当前页靠后：1 ... last-5 last-4 last-3 last-2 last-1 last
    pages.push(1);
    pages.push('...');
    for (let i = total - 5; i <= total; i++) {
      pages.push(i);
    }
  } else {
    // 当前页在中间：1 ... current-2 current-1 current current+1 current+2 ... last
    pages.push(1);
    pages.push('...');
    for (let i = current - 2; i <= current + 2; i++) {
      pages.push(i);
    }
    pages.push('...');
    pages.push(total);
  }

  return pages;
}

export function JobPagination({ currentPage, totalPages, onPageChange }: JobPaginationProps) {
  if (totalPages <= 1) return null;

  const pages = generatePaginationPages(currentPage, totalPages);
  const isFirstPage = currentPage <= 1;
  const isLastPage = currentPage >= totalPages;

  return (
    <nav className="flex items-center justify-end mt-5 pt-3" aria-label="分页导航">
      <div className="inline-flex items-center bg-white rounded-md overflow-hidden">
        {/* 左箭头 */}
        <button
          disabled={isFirstPage}
          onClick={() => !isFirstPage && onPageChange(currentPage - 1)}
          className={`flex items-center justify-center w-7 h-7 transition-colors ${
            isFirstPage
              ? 'text-gray-300 cursor-not-allowed'
              : 'text-gray-500 hover:bg-gray-50'
          }`}
          aria-label="上一页"
        >
          <ChevronLeft className="h-3 w-3" />
        </button>

        {/* 页码 */}
        {pages.map((item, index) => {
          if (item === '...') {
            return (
              <span
                key={`ellipsis-${index}`}
                className="flex items-center justify-center w-7 h-7 text-gray-400 text-[13px] select-none"
              >
                ...
              </span>
            );
          }

          const pageNum = item as number;
          const isActive = pageNum === currentPage;

          return (
            <button
              key={pageNum}
              onClick={() => onPageChange(pageNum)}
              className={`flex items-center justify-center w-7 h-7 text-[13px] transition-colors ${
                isActive
                  ? 'bg-[#0052D9] text-white'
                  : 'text-gray-700 hover:bg-gray-50'
              }`}
              aria-current={isActive ? 'page' : undefined}
            >
              {pageNum}
            </button>
          );
        })}

        {/* 右箭头 */}
        <button
          disabled={isLastPage}
          onClick={() => !isLastPage && onPageChange(currentPage + 1)}
          className={`flex items-center justify-center w-7 h-7 transition-colors ${
            isLastPage
              ? 'text-gray-300 cursor-not-allowed'
              : 'text-gray-500 hover:bg-gray-50'
          }`}
          aria-label="下一页"
        >
          <ChevronRight className="h-3 w-3" />
        </button>
      </div>
    </nav>
  );
}
