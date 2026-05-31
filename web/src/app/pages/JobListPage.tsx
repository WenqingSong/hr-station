import React from "react";
import { JobFilterSidebar } from "../components/JobFilterSidebar";
import { JobList } from "../components/JobList";
import { AIDecisionWorkspace } from "../components/AIDecisionWorkspace";

export function JobListPage() {
  return (
    <div className="mx-auto max-w-[1440px] px-6 md:px-8 py-6">
      <div className="flex gap-6 items-start">
        <JobFilterSidebar />
        <JobList />
        <AIDecisionWorkspace />
      </div>
    </div>
  );
}
