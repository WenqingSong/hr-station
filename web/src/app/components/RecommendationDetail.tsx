import React from "react";
import { CheckCircle2, ShieldAlert, Zap, ArrowRight } from "lucide-react";
import { Button } from "./ui/Button";
import { Badge } from "./ui/Badge";
import { useNavigate } from "react-router";

export function RecommendationDetail() {
  const navigate = useNavigate();

  return (
    <div className="flex-1 min-w-[560px] max-w-[680px] bg-white border border-gray-200 rounded-lg shadow-sm overflow-hidden">
      {/* Top Info Card */}
      <div className="bg-[#F4F8FF] p-6 border-b border-blue-100">
        <div className="flex items-center gap-2 text-gray-500 text-[13px] mb-2">
          当前选中岗位
        </div>
        <h2 className="text-xl font-bold text-gray-900 mb-4">软件开发-后台开发方向</h2>
        
        <div className="bg-white rounded-md p-4 border border-blue-50 shadow-sm flex items-start gap-4">
          <div className="shrink-0 flex flex-col items-center justify-center bg-green-50 text-green-700 border border-green-100 rounded p-2 min-w-[80px]">
            <CheckCircle2 className="w-5 h-5 mb-1" />
            <span className="text-[13px] font-bold">强推荐</span>
          </div>
          <p className="text-[14px] text-gray-700 leading-relaxed pt-0.5">
            <span className="font-semibold text-gray-900">一句话总结：</span>
            你的 Go 后端、MySQL、Redis 和微服务项目经历与该岗位核心要求高度相关，建议优先关注。
          </p>
        </div>
      </div>

      <div className="p-6 space-y-8">
        {/* Module 1: Reasons */}
        <section>
          <h3 className="font-bold text-gray-900 mb-4 text-[16px] flex items-center gap-2">
            <span className="w-1 h-4 bg-[#0052D9] rounded-full inline-block"></span>
            推荐理由
          </h3>
          <ul className="space-y-3">
            {[
              "后端开发方向与当前简历求职方向一致",
              "简历中 Go / MySQL / Redis 证据较强",
              "项目经历能覆盖岗位对后台系统设计和接口开发的要求"
            ].map((reason, i) => (
              <li key={i} className="flex items-start gap-3 text-[14px] text-gray-700 bg-gray-50 p-3 rounded-md">
                <span className="flex items-center justify-center w-5 h-5 rounded-full bg-blue-100 text-[#0052D9] text-[12px] font-bold shrink-0 mt-0.5">
                  {i + 1}
                </span>
                {reason}
              </li>
            ))}
          </ul>
        </section>

        {/* Module 2: Evidence */}
        <section>
          <h3 className="font-bold text-gray-900 mb-4 text-[16px] flex items-center gap-2">
            <span className="w-1 h-4 bg-[#0052D9] rounded-full inline-block"></span>
            匹配证据
          </h3>
          
          <div className="border border-gray-200 rounded-md overflow-hidden">
            <div className="grid grid-cols-12 gap-4 bg-gray-50 p-3 text-[13px] font-medium text-gray-600 border-b border-gray-200">
              <div className="col-span-5">岗位要求</div>
              <div className="col-span-5">简历证据</div>
              <div className="col-span-2 text-center">证据强度</div>
            </div>
            
            <div className="divide-y divide-gray-100 text-[13px]">
              <div className="grid grid-cols-12 gap-4 p-3 items-center hover:bg-gray-50 transition-colors">
                <div className="col-span-5 text-gray-800">熟悉 Go / Java / C++ 至少一门语言</div>
                <div className="col-span-5 text-gray-600">Go 微服务项目，包含接口开发和数据库设计</div>
                <div className="col-span-2 flex justify-center">
                  <Badge className="bg-green-50 text-green-700 border border-green-200 font-normal">强</Badge>
                </div>
              </div>
              
              <div className="grid grid-cols-12 gap-4 p-3 items-center hover:bg-gray-50 transition-colors">
                <div className="col-span-5 text-gray-800">理解数据库和缓存</div>
                <div className="col-span-5 text-gray-600">使用 MySQL、Redis 完成订单缓存与查询优化</div>
                <div className="col-span-2 flex justify-center">
                  <Badge className="bg-green-50 text-green-700 border border-green-200 font-normal">强</Badge>
                </div>
              </div>
              
              <div className="grid grid-cols-12 gap-4 p-3 items-center hover:bg-gray-50 transition-colors">
                <div className="col-span-5 text-gray-800">高并发或分布式经验</div>
                <div className="col-span-5 text-gray-600">有微服务项目，但缺少压测数据</div>
                <div className="col-span-2 flex justify-center">
                  <Badge className="bg-blue-50 text-[#0052D9] border border-blue-200 font-normal">中</Badge>
                </div>
              </div>
            </div>
          </div>
        </section>

        {/* Module 3: Risks */}
        <section>
          <h3 className="font-bold text-gray-900 mb-4 text-[16px] flex items-center gap-2">
            <span className="w-1 h-4 bg-amber-400 rounded-full inline-block"></span>
            主要风险与提醒
          </h3>
          <div className="bg-amber-50/30 border border-amber-100/50 rounded-md p-4">
            <ul className="space-y-2.5">
              {[
                "高并发压测结果描述不足",
                "Linux 使用经验证据不够明确",
                "缺少线上问题排查案例"
              ].map((risk, i) => (
                <li key={i} className="flex items-center gap-2.5 text-[13px] text-amber-700">
                  <div className="w-1.5 h-1.5 rounded-full bg-amber-400/80 shrink-0"></div>
                  {risk}
                </li>
              ))}
            </ul>
          </div>
        </section>

        {/* Module 4: Suggestions & Actions */}
        <section className="bg-gray-50 rounded-md p-5 border border-gray-100">
          <div className="flex items-start gap-3 mb-5">
            <Zap className="w-5 h-5 text-[#0052D9] shrink-0 mt-0.5" />
            <div className="text-[14px] text-gray-700 leading-relaxed">
              <p className="font-medium text-gray-900 mb-1">投递建议</p>
              建议优先进入岗位详情页，再进行岗位匹配分析。<br/>
              如果时间有限，可先修复 Top 3 简历问题后投递。
            </div>
          </div>
          
          <div className="flex gap-4">
            <Button 
              className="flex-1 bg-[#0052D9] hover:bg-[#0040A8] text-white h-11"
              onClick={() => navigate("/job/1")}
            >
              进入岗位详情
              <ArrowRight className="w-4 h-4 ml-2" />
            </Button>
            <Button 
              variant="outline" 
              className="flex-1 border-[#0052D9] text-[#0052D9] hover:bg-blue-50 h-11"
              onClick={() => navigate("/match-analysis")}
            >
              进行岗位匹配分析
            </Button>
          </div>
        </section>
      </div>
    </div>
  );
}
