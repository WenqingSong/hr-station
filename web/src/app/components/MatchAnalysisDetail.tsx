import React from "react";
import { CheckCircle2, ShieldAlert, Zap, ArrowRight, CheckSquare, Sparkles, Activity, PlusCircle } from "lucide-react";
import { Button } from "./ui/Button";
import { Badge } from "./ui/Badge";
import { useNavigate } from "react-router";

export function MatchAnalysisDetail() {
  const navigate = useNavigate();

  return (
    <div className="flex-1 min-w-[640px] max-w-[760px] bg-white border border-gray-200 rounded-lg shadow-sm overflow-hidden flex flex-col">
      {/* Module 1: Conclusion */}
      <div className="bg-gradient-to-br from-[#F4F8FF] to-white p-6 border-b border-blue-100">
        <h2 className="text-[18px] font-bold text-gray-900 mb-5 flex items-center gap-2">
          <Activity className="w-5 h-5 text-[#0052D9]" />
          匹配结论
        </h2>
        
        <div className="flex flex-col gap-5">
          <div className="grid grid-cols-3 gap-4">
            <div className="bg-white p-4 rounded-lg border border-blue-100 shadow-sm flex flex-col items-center justify-center text-center">
              <div className="text-[13px] text-gray-500 mb-1">匹配分数</div>
              <div className="text-3xl font-black text-[#0052D9] flex items-baseline">
                82<span className="text-[12px] text-gray-400 font-medium ml-1">/ 100</span>
              </div>
            </div>
            
            <div className="bg-white p-4 rounded-lg border border-amber-100 shadow-sm flex flex-col items-center justify-center text-center">
              <div className="text-[13px] text-gray-500 mb-1">投递准备度</div>
              <div className="text-[16px] font-bold text-amber-600">先修复后投递</div>
            </div>
            
            <div className="bg-white p-4 rounded-lg border border-gray-200 shadow-sm flex flex-col items-center justify-center text-center">
              <div className="text-[13px] text-gray-500 mb-1">优化停止信号</div>
              <div className="text-[15px] font-bold text-gray-800">修复 Top 3 后再投递</div>
            </div>
          </div>
          
          <p className="text-[14px] text-gray-700 leading-relaxed bg-white/60 p-3.5 rounded border border-blue-50">
            <span className="font-semibold text-gray-900">一句话总结：</span>
            你的 Go 后端、MySQL、Redis 和微服务项目经历与岗位核心要求较匹配，但高并发压测、Linux 使用经验和线上问题排查证据仍需补强。
          </p>
        </div>
      </div>

      <div className="p-6 space-y-8">
        {/* Module 2: Top 3 Fixes */}
        <section>
          <h3 className="font-bold text-gray-900 mb-4 text-[16px] flex items-center gap-2">
            <span className="w-1 h-4 bg-[#0052D9] rounded-full inline-block"></span>
            投递前最应该先修的 3 件事
          </h3>
          <div className="space-y-3">
            {[
              {
                title: "补充 Linux 使用经验",
                priority: "高",
                impact: "提升后台开发岗位可信度",
                verify: "是"
              },
              {
                title: "量化微服务项目 QPS / 响应时间 / 缓存命中率",
                priority: "高",
                impact: "增强高并发和性能优化证据",
                verify: "是"
              },
              {
                title: "补充线上问题排查或接口优化案例",
                priority: "中",
                impact: "增强工程实践与稳定性意识",
                verify: "是"
              }
            ].map((task, i) => (
              <div key={i} className="flex items-start gap-3 p-4 bg-gray-50 border border-gray-100 rounded-md hover:border-blue-100 transition-colors">
                <CheckSquare className="w-5 h-5 text-gray-300 shrink-0 mt-0.5" />
                <div className="flex-1">
                  <div className="text-[15px] font-bold text-gray-900 mb-2">{task.title}</div>
                  <div className="grid grid-cols-1 sm:grid-cols-3 gap-y-2 gap-x-4 text-[12px]">
                    <div className="flex items-center gap-1.5">
                      <span className="text-gray-400">优先级：</span>
                      <span className={task.priority === "高" ? "text-red-600 font-medium" : "text-amber-600 font-medium"}>
                        {task.priority}
                      </span>
                    </div>
                    <div className="flex items-center gap-1.5 sm:col-span-2">
                      <span className="text-gray-400">预计影响：</span>
                      <span className="text-gray-700">{task.impact}</span>
                    </div>
                    <div className="flex items-center gap-1.5 sm:col-span-3">
                      <span className="text-gray-400">需要用户核实：</span>
                      <span className="text-gray-700">{task.verify}</span>
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </section>

        {/* Module 3: Evidence Chain */}
        <section>
          <h3 className="font-bold text-gray-900 mb-4 text-[16px] flex items-center gap-2">
            <span className="w-1 h-4 bg-[#0052D9] rounded-full inline-block"></span>
            JD 证据链分析
          </h3>
          <div className="border border-gray-200 rounded-md overflow-hidden">
            <div className="grid grid-cols-12 gap-4 bg-gray-50 p-3 text-[12px] font-medium text-gray-600 border-b border-gray-200">
              <div className="col-span-3">JD 要求</div>
              <div className="col-span-4">简历证据</div>
              <div className="col-span-2 text-center">证据强度</div>
              <div className="col-span-3">风险说明</div>
            </div>
            <div className="divide-y divide-gray-100 text-[12px] leading-relaxed">
              {/* Row 1 */}
              <div className="grid grid-cols-12 gap-4 p-3 items-start hover:bg-gray-50 transition-colors">
                <div className="col-span-3 text-gray-800">熟悉 Go / Java / C++ 至少一门语言</div>
                <div className="col-span-4 text-gray-600">Go 微服务项目，包含接口开发和数据库设计</div>
                <div className="col-span-2 flex justify-center">
                  <Badge className="bg-green-50 text-green-700 border border-green-200 font-normal">强</Badge>
                </div>
                <div className="col-span-3 text-gray-400">-----</div>
              </div>
              {/* Row 2 */}
              <div className="grid grid-cols-12 gap-4 p-3 items-start hover:bg-gray-50 transition-colors">
                <div className="col-span-3 text-gray-800">理解数据库和缓存</div>
                <div className="col-span-4 text-gray-600">使用 MySQL、Redis 完成订单缓存与查询优化</div>
                <div className="col-span-2 flex justify-center">
                  <Badge className="bg-green-50 text-green-700 border border-green-200 font-normal">强</Badge>
                </div>
                <div className="col-span-3 text-gray-500">可以补充性能数据</div>
              </div>
              {/* Row 3 */}
              <div className="grid grid-cols-12 gap-4 p-3 items-start hover:bg-gray-50 transition-colors">
                <div className="col-span-3 text-gray-800">高并发或分布式系统经验</div>
                <div className="col-span-4 text-gray-600">有微服务项目，但缺少 QPS、压测、性能优化数据</div>
                <div className="col-span-2 flex justify-center">
                  <Badge className="bg-blue-50 text-[#0052D9] border border-blue-200 font-normal">中</Badge>
                </div>
                <div className="col-span-3 text-amber-600">建议量化项目指标</div>
              </div>
              {/* Row 4 */}
              <div className="grid grid-cols-12 gap-4 p-3 items-start hover:bg-gray-50 transition-colors">
                <div className="col-span-3 text-gray-800">熟悉 Linux / 后台服务部署</div>
                <div className="col-span-4 text-gray-600">简历中未明确体现 Linux 使用经验</div>
                <div className="col-span-2 flex justify-center">
                  <Badge className="bg-amber-50 text-amber-700 border border-amber-200 font-normal">弱</Badge>
                </div>
                <div className="col-span-3 text-amber-600">需要补充 Linux 环境开发或部署经历</div>
              </div>
            </div>
          </div>
        </section>

        <div className="grid grid-cols-2 gap-6">
          {/* Module 4: Strengths */}
          <section>
            <h3 className="font-bold text-gray-900 mb-3 text-[15px] flex items-center gap-2">
              <CheckCircle2 className="w-4 h-4 text-green-600" />
              你的优势
            </h3>
            <ul className="space-y-2 text-[13px] text-gray-700 bg-green-50/30 p-4 rounded-md border border-green-100/50">
              <li className="flex items-start gap-2">
                <div className="w-1.5 h-1.5 rounded-full bg-green-500 shrink-0 mt-1.5"></div>
                后端开发方向与岗位方向一致
              </li>
              <li className="flex items-start gap-2">
                <div className="w-1.5 h-1.5 rounded-full bg-green-500 shrink-0 mt-1.5"></div>
                Go / MySQL / Redis 技能与岗位要求高度相关
              </li>
              <li className="flex items-start gap-2">
                <div className="w-1.5 h-1.5 rounded-full bg-green-500 shrink-0 mt-1.5"></div>
                微服务项目能支撑后台系统设计与接口开发经历
              </li>
            </ul>
          </section>

          {/* Module 5: Risks */}
          <section>
            <h3 className="font-bold text-gray-900 mb-3 text-[15px] flex items-center gap-2">
              <ShieldAlert className="w-4 h-4 text-amber-500" />
              主要风险
            </h3>
            <ul className="space-y-2 text-[13px] text-amber-800 bg-amber-50/30 p-4 rounded-md border border-amber-100/50">
              <li className="flex items-start gap-2">
                <div className="w-1.5 h-1.5 rounded-full bg-amber-400 shrink-0 mt-1.5"></div>
                高并发压测结果描述不足
              </li>
              <li className="flex items-start gap-2">
                <div className="w-1.5 h-1.5 rounded-full bg-amber-400 shrink-0 mt-1.5"></div>
                Linux 使用经验不够明确
              </li>
              <li className="flex items-start gap-2">
                <div className="w-1.5 h-1.5 rounded-full bg-amber-400 shrink-0 mt-1.5"></div>
                线上问题排查案例缺失
              </li>
            </ul>
          </section>
        </div>

        {/* Module 6: Actions */}
        <section className="bg-gray-50 rounded-md p-5 border border-gray-200 mt-8">
          <div className="flex items-start gap-3 mb-5">
            <Zap className="w-5 h-5 text-[#0052D9] shrink-0 mt-0.5" />
            <div className="text-[14px] text-gray-700 leading-relaxed">
              <p className="font-bold text-gray-900 mb-1">下一步建议</p>
              建议先修复 Top 3 简历问题，再进入投递。当前岗位值得优先关注，但简历中仍有几处关键证据需要补强。
            </div>
          </div>
          
          <div className="flex items-center gap-4">
            <Button 
              className="flex-1 bg-[#0052D9] hover:bg-[#0040A8] text-white h-11"
              onClick={() => navigate("/resume-optimization")}
            >
              <PlusCircle className="w-4 h-4 mr-2" />
              针对该岗位优化简历
            </Button>
            <Button 
              variant="outline" 
              className="px-6 border-gray-300 text-gray-700 hover:bg-gray-50 h-11"
              onClick={() => navigate(-1)}
            >
              返回岗位详情
            </Button>
            <Button 
              variant="outline" 
              className="px-6 border-transparent text-[#0052D9] hover:bg-blue-50 h-11"
              onClick={() => navigate("/recommendations")}
            >
              查看推荐岗位
            </Button>
          </div>
        </section>
      </div>
    </div>
  );
}
