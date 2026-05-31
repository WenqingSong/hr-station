import React from "react";
import { CheckCircle2, AlertCircle, FileText, Briefcase, Calendar, FolderGit2 } from "lucide-react";
import { Badge } from "./ui/Badge";
import { Button } from "./ui/Button";

export function MyResumesDetail({ onOpenSelectModal }: { onOpenSelectModal?: () => void }) {
  return (
    <div className="flex-1 min-w-[560px] max-w-[720px] bg-white border border-gray-200 rounded-lg shadow-sm overflow-hidden flex flex-col">
      <div className="p-6 border-b border-gray-100 bg-gradient-to-r from-blue-50/50 to-white">
        <div className="flex items-start justify-between mb-4">
          <h2 className="text-[20px] font-bold text-gray-900 flex items-center gap-2">
            V2 后端开发简历
          </h2>
        </div>
        <div className="flex flex-wrap gap-2">
          <Badge className="bg-blue-100 text-[#0052D9] border-blue-200 font-medium px-2 py-0.5">当前使用</Badge>
          <Badge variant="outline" className="text-gray-600 bg-white font-normal px-2 py-0.5">后端开发</Badge>
          <Badge variant="outline" className="text-gray-600 bg-white font-normal px-2 py-0.5">2026 校园招聘</Badge>
          <Badge variant="outline" className="text-gray-600 bg-white font-normal px-2 py-0.5">Go</Badge>
          <Badge variant="outline" className="text-gray-600 bg-white font-normal px-2 py-0.5">MySQL</Badge>
          <Badge variant="outline" className="text-gray-600 bg-white font-normal px-2 py-0.5">Redis</Badge>
          <Badge variant="outline" className="text-gray-600 bg-white font-normal px-2 py-0.5">微服务</Badge>
        </div>
      </div>

      <div className="p-6 space-y-8">
        {/* Module 1: Basic Info */}
        <section>
          <h3 className="font-bold text-gray-900 text-[15px] mb-4 flex items-center gap-2">
            <FileText className="w-4 h-4 text-gray-400" />
            基础信息
          </h3>
          <div className="grid grid-cols-2 gap-y-3 gap-x-6 text-[13px]">
            <div className="flex items-start gap-2">
              <span className="text-gray-400 shrink-0">简历名称：</span>
              <span className="text-gray-800 font-medium">V2 后端开发简历</span>
            </div>
            <div className="flex items-start gap-2">
              <span className="text-gray-400 shrink-0">求职方向：</span>
              <span className="text-gray-800">后端开发</span>
            </div>
            <div className="flex items-start gap-2">
              <span className="text-gray-400 shrink-0">目标岗位：</span>
              <span className="text-gray-800">软件开发-后台开发方向 / 云计算开发实习生</span>
            </div>
            <div className="flex items-start gap-2">
              <span className="text-gray-400 shrink-0">更新时间：</span>
              <span className="text-gray-800">2 小时前</span>
            </div>
            <div className="flex items-start gap-2">
              <span className="text-gray-400 shrink-0">项目经历：</span>
              <span className="text-gray-800">3 个</span>
            </div>
            <div className="flex items-start gap-2">
              <span className="text-gray-400 shrink-0">当前状态：</span>
              <span className="text-[#0052D9] font-medium flex items-center gap-1">
                <CheckCircle2 className="w-3.5 h-3.5" />
                可用于 AI 分析
              </span>
            </div>
          </div>
        </section>

        {/* Module 2: Skills */}
        <section>
          <h3 className="font-bold text-gray-900 text-[15px] mb-4 flex items-center gap-2">
            <Briefcase className="w-4 h-4 text-gray-400" />
            技能关键词
          </h3>
          <div className="flex flex-wrap gap-2">
            {["Go", "MySQL", "Redis", "Docker", "微服务", "RBAC", "接口开发", "缓存优化"].map(skill => (
              <Badge key={skill} variant="outline" className="bg-gray-50 text-gray-700 border-gray-200 font-normal px-2.5 py-1 text-[12px]">
                {skill}
              </Badge>
            ))}
          </div>
        </section>

        {/* Module 3: Projects */}
        <section>
          <h3 className="font-bold text-gray-900 text-[15px] mb-4 flex items-center gap-2">
            <FolderGit2 className="w-4 h-4 text-gray-400" />
            项目经历摘要
          </h3>
          <div className="space-y-4">
            <div className="p-4 rounded-md border border-gray-200 bg-gray-50/50">
              <div className="font-bold text-gray-900 text-[14px] mb-2">项目 1：Go 微服务订单系统</div>
              <div className="text-[12px] text-gray-500 mb-2">关键词：Go / MySQL / Redis / 微服务</div>
              <p className="text-[13px] text-gray-700 leading-relaxed">
                负责订单服务核心接口设计与实现，使用 MySQL 完成订单数据建模，结合 Redis 缓存优化高频查询路径。
              </p>
            </div>
            <div className="p-4 rounded-md border border-gray-200 bg-gray-50/50">
              <div className="font-bold text-gray-900 text-[14px] mb-2">项目 2：RBAC 权限管理系统</div>
              <div className="text-[12px] text-gray-500 mb-2">关键词：权限控制 / 接口鉴权 / 后台系统</div>
              <p className="text-[13px] text-gray-700 leading-relaxed">
                设计并实现用户、角色、菜单与接口权限的关联校验，提升后台系统权限管理的可维护性。
              </p>
            </div>
            <div className="p-4 rounded-md border border-gray-200 bg-gray-50/50">
              <div className="font-bold text-gray-900 text-[14px] mb-2">项目 3：缓存优化项目</div>
              <div className="text-[12px] text-gray-500 mb-2">关键词：Redis / 查询优化 / 性能优化</div>
              <p className="text-[13px] text-gray-700 leading-relaxed">
                针对高频查询场景设计缓存策略，减少重复数据库访问。
              </p>
            </div>
          </div>
        </section>

        {/* Module 4: Gaps */}
        <section>
          <h3 className="font-bold text-gray-900 text-[15px] mb-4 flex items-center gap-2">
            <AlertCircle className="w-4 h-4 text-gray-400" />
            简历当前缺口
          </h3>
          <div className="bg-amber-50 border border-amber-100 rounded-md p-4">
            <ul className="space-y-2 text-[13px] text-amber-800">
              <li className="flex items-center gap-2">
                <div className="w-1.5 h-1.5 rounded-full bg-amber-400 shrink-0"></div>
                Linux 使用经验描述不足
              </li>
              <li className="flex items-center gap-2">
                <div className="w-1.5 h-1.5 rounded-full bg-amber-400 shrink-0"></div>
                高并发压测数据缺失
              </li>
              <li className="flex items-center gap-2">
                <div className="w-1.5 h-1.5 rounded-full bg-amber-400 shrink-0"></div>
                线上问题排查案例较少
              </li>
            </ul>
          </div>
        </section>

        <div className="h-px bg-gray-200 w-full"></div>

        {/* Module 5: Actions */}
        <section className="flex flex-col gap-3">
          <div className="flex items-center gap-3">
            <Button variant="outline" className="flex-1 bg-gray-50 text-gray-400 border-gray-200 cursor-not-allowed h-10">
              当前简历已启用
            </Button>
            <Button 
              variant="outline" 
              className="flex-1 border-[#0052D9] text-[#0052D9] hover:bg-blue-50 h-10"
              onClick={onOpenSelectModal}
            >
              用于岗位推荐
            </Button>
            <Button 
              className="flex-1 bg-[#0052D9] hover:bg-[#0040A8] text-white h-10"
              onClick={onOpenSelectModal}
            >
              用于岗位匹配分析
            </Button>
          </div>
          <div className="text-[12px] text-gray-400 text-center flex items-center justify-center gap-1.5">
            <AlertCircle className="w-3.5 h-3.5" />
            若要针对某个岗位优化简历，请先进入岗位匹配分析页。
          </div>
        </section>
      </div>
    </div>
  );
}