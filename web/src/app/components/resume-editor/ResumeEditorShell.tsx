import React from "react";
import { ContentPageShell } from "../layout/ContentPageShell";

export interface ResumeEditorShellProps {
  title: string;
  subtitle: string;
  sidebar?: React.ReactNode;
  saveCard?: React.ReactNode;
  children: React.ReactNode;
}

/**
 * Main layout shell for the V2 resume editor.
 *
 * Layout:
 * - max-w-[1280px]
 * - Desktop: grid with main column + right sidebar (w-[280px], sticky)
 * - Mobile: single column, sidebar moves to top
 * - Bottom: save card
 *
 * Preserves existing ContentPageShell wrapper for consistency with the app.
 */
export function ResumeEditorShell({
  title,
  subtitle,
  sidebar,
  saveCard,
  children,
}: ResumeEditorShellProps) {
  const hasSidebar = !!sidebar;

  return (
    <ContentPageShell maxWidthClassName="max-w-[1280px]">
      {/* ── Page Header ────────────────────────────────────────────── */}
      <div className="pt-8 pb-4">
        <h1 className="text-2xl font-bold text-slate-950">{title}</h1>
        <p className="mt-2 text-sm text-slate-500">{subtitle}</p>
      </div>

      {/* ── Main layout: two-column on desktop, single on mobile ──── */}
      <div className="flex flex-col lg:flex-row lg:gap-6">
        {/* Main form area */}
        <div className="flex-1 min-w-0 space-y-5">{children}</div>

        {/* Right sidebar — desktop: sticky, mobile: top/shrink */}
        {hasSidebar && (
          <aside className="lg:w-[280px] lg:flex-shrink-0 mb-4 lg:mb-0 lg:order-last order-first">
            <div className="lg:sticky lg:top-24 space-y-4">{sidebar}</div>
          </aside>
        )}
      </div>

      {/* ── Bottom Save Card ───────────────────────────────────────── */}
      {saveCard && <div className="mt-8">{saveCard}</div>}
    </ContentPageShell>
  );
}
