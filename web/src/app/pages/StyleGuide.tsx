import { Button } from "../components/ui/Button";
import { Tag } from "../components/ui/Tag";
import { AIBadge } from "../components/ui/AIBadge";
import { JobFilterSidebar } from "../components/JobFilterSidebar";
import { AIDecisionWorkspace } from "../components/AIDecisionWorkspace";
import { DeliveryReadinessCard } from "../components/business/DeliveryReadinessCard";
import { OptimizationStopSignalCard } from "../components/business/OptimizationStopSignalCard";
import { TopFixTaskCard } from "../components/business/TopFixTaskCard";
import { JDEvidenceTable } from "../components/business/JDEvidenceTable";
import { RiskWarningCard } from "../components/business/RiskWarningCard";
import { CurrentResumeSummaryCard } from "../components/business/CurrentResumeSummaryCard";
import { JobCard } from "../components/business/JobCard";
import { FileText, UserPlus, Zap } from "lucide-react";

export function StyleGuide() {
  return (
    <>
      <JobFilterSidebar />
      
      <main className="flex-1 max-w-[820px] min-w-[680px] space-y-8 pb-20">
        
        {/* Section 1: Typography & Basic Components */}
        <section className="bg-white p-6 rounded-lg border border-slate-200">
          <h2 className="text-xl font-semibold text-slate-800 mb-6 pb-2 border-b border-slate-100">
            UI 组件与样式规范
          </h2>
          
          <div className="space-y-6">
            <div>
              <h3 className="text-sm font-medium text-slate-500 mb-3">按钮 Buttons</h3>
              <div className="flex flex-wrap gap-3">
                <Button variant="primary">主按钮 (Primary)</Button>
                <Button variant="secondary">次按钮 (Secondary)</Button>
                <Button variant="weak">弱按钮 (Weak)</Button>
                <Button variant="outline">描边按钮 (Outline)</Button>
              </div>
            </div>

            <div>
              <h3 className="text-sm font-medium text-slate-500 mb-3">标签 Tags</h3>
              <div className="flex flex-wrap gap-2">
                <Tag>C/C++</Tag>
                <Tag>WXG</Tag>
                <Tag variant="blue">高并发</Tag>
                <Tag variant="outline">工作地点: 深圳</Tag>
              </div>
            </div>

            <div>
              <h3 className="text-sm font-medium text-slate-500 mb-3">AI 状态徽标 Badges</h3>
              <div className="flex flex-wrap gap-3">
                <AIBadge state="generated" />
                <AIBadge state="repaired" />
                <AIBadge state="fallback" />
                <AIBadge state="mock" />
                <AIBadge state="abnormal" />
              </div>
            </div>
          </div>
        </section>

        {/* Section 2: Business Components */}
        <section className="bg-white p-6 rounded-lg border border-slate-200">
          <h2 className="text-xl font-semibold text-slate-800 mb-6 pb-2 border-b border-slate-100">
            业务卡片展示
          </h2>
          
          <div className="space-y-6">
            <JobCard 
              title="软件开发-后台开发方向"
              tags={["C/C++开发", "CDG", "WXG", "TEG", "后台", "微服务"]}
              location="深圳, 北京, 广州"
              status="招聘中"
              deadline="3天后截止"
              isHot
            />

            <div className="grid grid-cols-2 gap-4">
              <CurrentResumeSummaryCard />
              <div className="space-y-4">
                <RiskWarningCard 
                  risks={[
                    "核心技能「C++」无直接项目证据支撑",
                    "缺少「高并发」处理经验的相关描述"
                  ]}
                />
              </div>
            </div>

            <JDEvidenceTable 
              data={[
                { jdReq: "熟练掌握 C/C++/Go 语言", evidence: "Go 微服务项目经验，熟练使用 Go", strength: "strong" },
                { jdReq: "了解分布式系统设计", evidence: "使用过 Redis, Kafka，但未提及系统设计", strength: "weak", risk: "可能在系统设计面试中处于劣势" },
                { jdReq: "熟悉 Linux 操作系统", evidence: "", strength: "missing", risk: "需补充 Linux 相关开发环境经验" }
              ]}
            />
          </div>
        </section>
      </main>

      {/* Right Sidebar: AI Decision Workspace */}
      <AIDecisionWorkspace>
        <div className="space-y-4">
          <Button variant="primary" className="w-full gap-2">
            <Zap className="w-4 h-4" />
            岗位匹配分析
          </Button>

          <DeliveryReadinessCard status="fix_first" />
          
          <TopFixTaskCard 
            tasks={[
              { id: "1", title: "补充 Linux 使用经验", priority: "high", impact: "通过率 +15%", needsVerification: false },
              { id: "2", title: "量化微服务项目 QPS 收益", priority: "high", impact: "面试邀请 +20%", needsVerification: true },
              { id: "3", title: "精简非技术类校园经历", priority: "medium", impact: "阅读体验改善", needsVerification: false }
            ]}
          />
          
          <OptimizationStopSignalCard type="fix_needed" />

          {/* Demonstration of another state */}
          <div className="mt-8 pt-4 border-t border-slate-100 border-dashed">
            <div className="text-xs text-slate-400 mb-3 text-center">其它状态演示</div>
            <DeliveryReadinessCard status="ready" />
            <div className="mt-4">
              <OptimizationStopSignalCard type="ready" />
            </div>
          </div>
        </div>
      </AIDecisionWorkspace>
    </>
  );
}
