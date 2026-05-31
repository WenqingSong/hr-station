import React from "react";
import { JobFilterSidebar } from "../components/JobFilterSidebar";
import { JobList } from "../components/JobList";
import { AIDecisionWorkspace } from "../components/AIDecisionWorkspace";

export function JobListPage() {
  return (
    <div className="min-h-screen bg-[#F3F4F6] font-sans text-gray-900 selection:bg-blue-100 selection:text-blue-900">
      <div className="mx-auto max-w-[1440px] px-6 py-8">
        <div className="flex gap-6 items-start">
          <JobFilterSidebar />
          <JobList />
          <AIDecisionWorkspace />
        </div>
      </div>
    </div>
  );
}
