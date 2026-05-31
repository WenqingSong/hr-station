import React, { useState } from "react";
import { Activity, Wrench, Sparkles, AlertCircle, RefreshCw, FileEdit, CheckCircle2, ChevronRight, CheckSquare, ChevronDown, ChevronUp } from "lucide-react";
import { Button } from "./ui/Button";
import { Badge } from "./ui/Badge";
import { useNavigate } from "react-router";

const CollapsibleSection = ({ title, icon: Icon, children, defaultOpen = false }: { title: string, icon: any, children: React.ReactNode, defaultOpen?: boolean }) => {
  const [isOpen, setIsOpen] = useState(defaultOpen);
  return (
    <section className="border border-gray-200 rounded-md overflow-hidden bg-white shadow-sm">
      <button 
        onClick={() => setIsOpen(!isOpen)}
        className="w-full flex items-center justify-between p-4 bg-gray-50 hover:bg-gray-100 transition-colors"
      >
        <div className="font-bold text-gray-700 text-[14px] flex items-center gap-2">
          <Icon className="w-4 h-4 text-gray-500" />
          {title}
        </div>
        {isOpen ? <ChevronUp className="w-4 h-4 text-gray-400" /> : <ChevronDown className="w-4 h-4 text-gray-400" />}
      </button>
      {isOpen && <div className="p-4 border-t border-gray-200">{children}</div>}
    </section>
  );
};

export function ResumeOptimizationDetail() {
  const navigate = useNavigate();

  return (
    <div className="flex-1 min-w-[660px] max-w-[780px] bg-white border border-gray-200 rounded-lg shadow-sm overflow-hidden flex flex-col">
      {/* Module 1: Optimization Conclusion (P0) */}
      <div className="bg-gradient-to-br from-[#F4F8FF] to-white p-6 border-b border-blue-100">
        <h2 className="text-[16px] font-bold text-gray-900 mb-4 flex items-center gap-2">
          <Activity className="w-4 h-4 text-[#0052D9]" />
          优化结论
        </h2>
        
        <div className="flex flex-col gap-4">
          <div className="grid grid-cols-2 gap-4">
            <div className="bg-white px-5 py-4 rounded-lg border-l-4 border-l-amber-500 border-y border-r border-gray-200 shadow-sm flex items-center justify-between">
              <div className="text-[13px] text-gray-500">投递准备度</div>
              <div className="text-[15px] font-bold text-amber-600">先修复后投递</div>
            </div>
            
            <div className="bg-white px-5 py-4 rounded-lg border-l-4 border-l-[#0052D9] border-y border-r border-gray-200 shadow-sm flex items-center justify-between">
              <div className="text-[13px] text-gray-500">优化停止信号</div>
              <div className="text-[15px] font-bold text-[#0052D9]">修复 Top 3 后再投递</div>
            </div>
          </div>
          
          <p className="text-[13px] text-gray-700 leading-relaxed bg-blue-50/50 p-3 rounded">
            <span className="font-bold text-gray-900 mr-2">诊断摘要：</span>
            当前简历与后台开发岗位方向匹配，但 Linux 使用经验、高并发压测结果和线上问题排查案例证据不足。建议先完成 3 个关键修复，再进入投递。
          </p>
        </div>
      </div>

      <div className="p-6 space-y-8">
        {/* Module 2: Top 3 Quick Fixes (P0) */}
        <section>
          <h3 className="font-bold text-gray-900 mb-4 text-[16px] flex items-center gap-2">
            <Wrench className="w-4 h-4 text-[#0052D9]" />
            投递前最应该先修的 3 件事
          </h3>
          <div className="space-y-4">
            {[
              {
                title: "补充 Linux 使用经验",
                priority: "高",
                impact: "提升后台开发岗位可信度",
                verify: "是",
                desc: "如果你确实在项目开发或部署中使用过 Linux，请补充具体命令、环境或排查经历；如果没有，不要编造。"
              },
              {
                title: "量化微服务项目 QPS / 响应时间 / 缓存命中率",
                priority: "高",
                impact: "增强高并发和性能优化证据",
                verify: "是",
                desc: "如果项目有压测或性能数据，请补充真实指标；如果没有，可以描述优化思路，但不要虚构数据。"
              },
              {
                title: "补充线上问题排查或接口优化案例",
                priority: "中",
                impact: "增强工程实践与稳定性意识",
                verify: "是",
                desc: "补充一次接口慢查询、缓存失效、权限校验、异常排查或日志定位经历。"
              }
            ].map((task, i) => (
              <div key={i} className="flex items-start gap-3 p-4 bg-gray-50 border border-gray-200 rounded-md">
                <div className="flex items-center justify-center w-6 h-6 rounded-full bg-white border border-gray-300 text-gray-700 text-[13px] font-bold shrink-0">
                  {i + 1}
                </div>
                <div className="flex-1">
                  <div className="text-[15px] font-bold text-gray-900 mb-2">{task.title}</div>
                  <div className="grid grid-cols-1 sm:grid-cols-3 gap-y-2 gap-x-4 text-[13px] mb-3">
                    <div className="flex items-center gap-1.5">
                      <span className="text-gray-500">优先级：</span>
                      <span className={task.priority === "高" ? "text-red-600 font-medium" : "text-blue-600 font-medium"}>
                        {task.priority}
                      </span>
                    </div>
                    <div className="flex items-center gap-1.5 sm:col-span-2">
                      <span className="text-gray-500">预计影响：</span>
                      <span className="text-gray-800">{task.impact}</span>
                    </div>
                  </div>
                  <div className="bg-amber-50/50 border border-amber-100/50 p-2.5 rounded text-[12px] text-gray-700 flex items-start gap-2">
                    <AlertCircle className="w-3.5 h-3.5 text-amber-500 shrink-0 mt-0.5" />
                    <div>
                      <span className="font-semibold text-amber-800 mr-1">需要用户核实：</span>
                      {task.desc}
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </section>

        {/* Module 5: Verification Checklist (P0) */}
        <section>
          <h3 className="font-bold text-gray-900 mb-4 text-[16px] flex items-center gap-2">
            <CheckSquare className="w-4 h-4 text-[#0052D9]" />
            需要你核实后再写入简历的信息
          </h3>
          <div className="bg-orange-50/40 border border-orange-100 rounded-md p-4">
            <ul className="space-y-3 text-[13px]">
              {[
                "是否真实使用过 Linux 开发或部署环境",
                "是否有项目压测数据，如 QPS、响应时间、缓存命中率",
                "是否有线上问题排查或日志定位经历",
                "是否参与过服务拆分、接口设计或性能优化",
                "是否能证明 MySQL / Redis 的实际使用效果"
              ].map((item, i) => (
                <li key={i} className="flex items-center justify-between gap-4 border-b border-orange-100/50 pb-2 last:border-0 last:pb-0">
                  <div className="flex items-center gap-2 text-gray-800">
                    <div className="w-1.5 h-1.5 rounded-full bg-orange-400 shrink-0"></div>
                    {item}
                  </div>
                  <Badge className="bg-orange-100 text-orange-700 border-orange-200 font-normal shrink-0">待核实</Badge>
                </li>
              ))}
            </ul>
          </div>
        </section>

        {/* Module 6: Before & After Example (P0) */}
        <section>
          <h3 className="font-bold text-gray-900 mb-4 text-[16px] flex items-center gap-2">
            <Sparkles className="w-4 h-4 text-[#0052D9]" />
            修改前后对比示例
          </h3>
          <div className="flex items-stretch border border-gray-200 rounded-md overflow-hidden bg-gray-50">
            <div className="flex-1 p-4 border-r border-gray-200">
              <div className="text-[12px] font-bold text-gray-500 mb-2">原简历表达</div>
              <div className="text-[13px] text-gray-700">完成订单接口开发，使用 MySQL 和 Redis。</div>
            </div>
            <div className="flex flex-col items-center justify-center px-4 bg-white">
              <div className="text-[11px] text-[#0052D9] font-medium mb-1 text-center">优化建议</div>
              <ChevronRight className="w-4 h-4 text-gray-300" />
              <div className="text-[10px] text-gray-500 mt-1 text-center w-24">补充职责、技术选择、优化目标和结果指标</div>
            </div>
            <div className="flex-[1.5] p-4 bg-blue-50/30 border-l border-blue-100">
              <div className="text-[12px] font-bold text-[#0052D9] mb-2 flex justify-between">
                优化后示例
              </div>
              <div className="text-[13px] text-gray-800 leading-relaxed">
                负责订单服务核心接口设计与实现，使用 MySQL 完成订单数据建模，并结合 Redis 缓存优化高频查询路径；如有真实压测数据，可补充 QPS、响应时间或缓存命中率等指标。
              </div>
              <div className="text-[11px] text-amber-600 mt-3 pt-3 border-t border-blue-100/50 flex items-center gap-1">
                <AlertCircle className="w-3.5 h-3.5" />
                需用户核实真实经历和数据后使用。
              </div>
            </div>
          </div>
        </section>

        <div className="h-px bg-gray-200 w-full my-6"></div>

        {/* Collapsible P1 Modules */}
        <div className="space-y-4">
          {/* Module 3: Keyword Alignment (P1) */}
          <CollapsibleSection title="展开查看：JD 关键词对齐建议" icon={RefreshCw}>
            <div className="border border-gray-200 rounded-md overflow-hidden">
              <div className="grid grid-cols-12 gap-4 bg-gray-50 p-3 text-[13px] font-medium text-gray-600 border-b border-gray-200">
                <div className="col-span-3">JD 关键词</div>
                <div className="col-span-3">简历当前覆盖</div>
                <div className="col-span-6">建议动作</div>
              </div>
              <div className="divide-y divide-gray-100 text-[13px] leading-relaxed">
                <div className="grid grid-cols-12 gap-4 p-3 items-center hover:bg-gray-50 transition-colors">
                  <div className="col-span-3 font-medium text-gray-800">Go</div>
                  <div className="col-span-3"><Badge className="bg-green-50 text-green-700 border-green-200 font-normal">已覆盖</Badge></div>
                  <div className="col-span-6 text-gray-600">保留并突出项目中的接口开发职责</div>
                </div>
                <div className="grid grid-cols-12 gap-4 p-3 items-center hover:bg-gray-50 transition-colors">
                  <div className="col-span-3 font-medium text-gray-800">MySQL / Redis</div>
                  <div className="col-span-3"><Badge className="bg-green-50 text-green-700 border-green-200 font-normal">已覆盖</Badge></div>
                  <div className="col-span-6 text-gray-600">补充缓存命中率或查询优化效果</div>
                </div>
                <div className="grid grid-cols-12 gap-4 p-3 items-center hover:bg-gray-50 transition-colors">
                  <div className="col-span-3 font-medium text-gray-800">Linux</div>
                  <div className="col-span-3"><Badge className="bg-amber-50 text-amber-700 border-amber-200 font-normal">弱覆盖</Badge></div>
                  <div className="col-span-6 text-gray-600">补充开发、部署或问题排查场景</div>
                </div>
                <div className="grid grid-cols-12 gap-4 p-3 items-center hover:bg-gray-50 transition-colors">
                  <div className="col-span-3 font-medium text-gray-800">高并发</div>
                  <div className="col-span-3"><Badge className="bg-red-50 text-red-600 border-red-200 font-normal">未明确覆盖</Badge></div>
                  <div className="col-span-6 text-gray-600">补充压测指标或性能优化过程</div>
                </div>
                <div className="grid grid-cols-12 gap-4 p-3 items-center hover:bg-gray-50 transition-colors">
                  <div className="col-span-3 font-medium text-gray-800">微服务</div>
                  <div className="col-span-3"><Badge className="bg-green-50 text-green-700 border-green-200 font-normal">已覆盖</Badge></div>
                  <div className="col-span-6 text-gray-600">说明服务拆分、接口设计和治理方式</div>
                </div>
              </div>
            </div>
          </CollapsibleSection>

          {/* Module 4: Project Experience Optimization (P1) */}
          <CollapsibleSection title="展开查看：项目经历详细优化建议 (2条)" icon={FileEdit}>
            <div className="space-y-4">
              <div className="border border-gray-200 rounded-md p-4">
                <div className="font-bold text-gray-900 text-[15px] mb-3">项目 1：Go 微服务订单系统</div>
                <div className="space-y-3 text-[13px]">
                  <div className="flex gap-3">
                    <span className="text-gray-500 shrink-0 w-[68px]">当前问题：</span>
                    <span className="text-gray-800">描述偏功能实现，缺少性能指标和稳定性证据。</span>
                  </div>
                  <div className="flex gap-3">
                    <span className="text-gray-500 shrink-0 w-[68px]">建议改法：</span>
                    <span className="text-[#0052D9] font-medium">突出接口设计、数据库设计、Redis 缓存、错误处理、权限校验和服务拆分。</span>
                  </div>
                  <div className="flex gap-3 bg-gray-50 p-3 rounded border border-gray-100">
                    <span className="text-gray-500 shrink-0 w-[68px]">示例表达：</span>
                    <div className="text-gray-700 leading-relaxed">
                      将“完成订单接口开发”优化为“负责订单服务核心接口设计与实现，使用 MySQL 完成订单数据建模，结合 Redis 缓存降低高频查询压力，并通过统一错误处理和权限校验提升接口稳定性”。
                      <div className="text-[11px] text-amber-600 mt-2 flex items-center gap-1">
                        <AlertCircle className="w-3 h-3" />
                        需用户核实后使用
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <div className="border border-gray-200 rounded-md p-4">
                <div className="font-bold text-gray-900 text-[15px] mb-3">项目 2：RBAC 权限管理系统</div>
                <div className="space-y-3 text-[13px]">
                  <div className="flex gap-3">
                    <span className="text-gray-500 shrink-0 w-[68px]">当前问题：</span>
                    <span className="text-gray-800">与后台开发相关，但没有突出权限校验、接口安全和系统稳定性。</span>
                  </div>
                  <div className="flex gap-3">
                    <span className="text-gray-500 shrink-0 w-[68px]">建议改法：</span>
                    <span className="text-[#0052D9] font-medium">强���角色权限模型、接口鉴权、操作日志和异常处理。</span>
                  </div>
                  <div className="flex gap-3 bg-gray-50 p-3 rounded border border-gray-100">
                    <span className="text-gray-500 shrink-0 w-[68px]">示例表达：</span>
                    <div className="text-gray-700 leading-relaxed">
                      将“实现权限管理功能”优化为“设计并实现基于角色的权限控制流程，完成用户、角色、菜单与接口权限的关联校验，并通过操作日志和异常处理提升后台系统可维护性”。
                      <div className="text-[11px] text-amber-600 mt-2 flex items-center gap-1">
                        <AlertCircle className="w-3 h-3" />
                        需用户核实后使用
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </CollapsibleSection>
        </div>

        {/* Module 7: Pre-delivery Confirmation & Actions (P0) */}
        <section className="bg-white border border-gray-200 shadow-sm rounded-md p-5 mt-8">
          <div className="mb-5">
            <h3 className="font-bold text-gray-900 mb-3 text-[15px] flex items-center gap-2">
              投递前确认
            </h3>
            <div className="grid grid-cols-2 gap-3 text-[13px] text-gray-700">
              <div className="flex items-center gap-2">
                <CheckCircle2 className="w-4 h-4 text-gray-300" /> Top 3 修复任务是否完成
              </div>
              <div className="flex items-center gap-2">
                <CheckCircle2 className="w-4 h-4 text-gray-300" /> 需要核实的信息是否确认
              </div>
              <div className="flex items-center gap-2">
                <CheckCircle2 className="w-4 h-4 text-gray-300" /> 是否保留真实经历边界
              </div>
              <div className="flex items-center gap-2">
                <CheckCircle2 className="w-4 h-4 text-gray-300" /> 是否准备进入岗位投递
              </div>
            </div>
          </div>
          
          <div className="flex items-center gap-4 pt-5 border-t border-gray-100">
            <Button 
              className="flex-[2] bg-[#0052D9] hover:bg-[#0040A8] text-white h-11 shadow-sm"
            >
              <Sparkles className="w-4 h-4 mr-2" />
              生成优化后简历草稿
            </Button>
            <Button 
              variant="outline" 
              className="flex-1 border-gray-300 text-gray-700 hover:bg-gray-50 h-11"
              onClick={() => navigate(-1)}
            >
              返回岗位匹配分析
            </Button>
            <Button 
              variant="outline" 
              className="flex-1 border-transparent text-[#0052D9] hover:bg-blue-50 h-11"
              onClick={() => navigate("/job/1")}
            >
              查看岗位详情
            </Button>
          </div>
        </section>
      </div>
    </div>
  );
}