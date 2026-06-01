import { useState, useEffect } from "react";
import { Link } from "react-router";
import { 
  ArrowLeft, 
  ArrowRight, 
  CheckCircle2, 
  ChevronDown, 
  AlertCircle, 
  Plus, 
  ArrowUpCircle,
  FileText
} from "lucide-react";
import { loadResumeComparisonViewModel } from "../../api";
import type { ResumeComparisonViewModel } from "../../adapters";

export function ResumeComparisonPage() {
  const [comparison, setComparison] = useState<ResumeComparisonViewModel | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadResumeComparisonViewModel("1", "2")
      .then(setComparison)
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  if (loading) {
    return (
      <div className="max-w-[1360px] mx-auto px-6 py-12 text-center text-gray-400">
        正在加载版本对比…
      </div>
    );
  }
  return (
    <>
      <div className="max-w-[1360px] mx-auto px-6 md:px-8 py-6 pb-12">
        
        {/* Header Section */}
        <div className="mb-8">
          <Link to="/resumes" className="flex items-center text-sm text-slate-500 hover:text-blue-600 transition-colors mb-4 group">
            <ArrowLeft className="w-4 h-4 mr-1 transition-transform group-hover:-translate-x-0.5" />
            返回我的简历
          </Link>
          <h1 className="text-2xl font-semibold text-slate-900 tracking-tight mb-1">历史简历比较</h1>
          <p className="text-sm text-slate-500">
            比较不同简历版本的变化，查看关键词、项目经历和岗位适配方向的差异。
          </p>
        </div>

        {/* Version Selection Region */}
        <div className="bg-white rounded-lg border border-gray-200 shadow-sm p-4 mb-8 flex items-center justify-between">
          <div className="flex items-center space-x-6">
            <div className="flex flex-col">
              <span className="text-xs text-slate-500 mb-1">旧版本</span>
              <button className="flex items-center text-sm font-medium text-slate-700 bg-slate-50 hover:bg-slate-100 px-3 py-1.5 rounded border border-gray-200 transition-colors">
                {comparison?.baseName || "V1 通用后端简历"}
                <ChevronDown className="w-4 h-4 ml-2 text-slate-400" />
              </button>
            </div>
            
            <div className="text-slate-300 px-2 mt-4">
              <ArrowRight className="w-5 h-5" />
            </div>

            <div className="flex flex-col">
              <span className="text-xs text-slate-500 mb-1">新版本</span>
              <button className="flex items-center text-sm font-medium text-blue-700 bg-blue-50 hover:bg-blue-100 px-3 py-1.5 rounded border border-blue-200 transition-colors">
                {comparison?.targetName || "V2 后端开发简历"}
                <ChevronDown className="w-4 h-4 ml-2 text-blue-400" />
              </button>
            </div>
          </div>

          <div className="flex items-center text-sm">
            <span className="text-slate-500 mr-3">当前建议版本：<span className="font-medium text-slate-900">{comparison?.targetName || "V2 后端开发简历"}</span></span>
            <span className="inline-flex items-center px-2 py-1 rounded bg-blue-50 text-blue-600 text-xs font-medium border border-blue-100">
              <CheckCircle2 className="w-3.5 h-3.5 mr-1" />
              推荐设为当前简历
            </span>
          </div>
        </div>

        {/* 3-Column Comparison Layout */}
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
          
          {/* Left Column: Old Version Summary */}
          <div className="lg:col-span-3">
            <div className="bg-white rounded-xl border border-gray-200 p-5 shadow-sm sticky top-20">
              <div className="mb-4">
                <h3 className="font-semibold text-slate-900 mb-2">旧版本：{comparison?.baseName || "V1 通用后端简历"}</h3>
                <span className="inline-flex px-2 py-0.5 rounded text-xs font-medium bg-slate-100 text-slate-600">
                  历史版本
                </span>
              </div>
              
              <div className="space-y-4 text-sm">
                <div>
                  <div className="text-slate-500 text-xs mb-1">求职方向</div>
                  <div className="font-medium text-slate-800">通用后端</div>
                </div>
                
                <div>
                  <div className="text-slate-500 text-xs mb-1">关键词 (3)</div>
                  <div className="flex flex-wrap gap-1.5">
                    {["Go", "Java", "MySQL"].map(kw => (
                      <span key={kw} className="px-2 py-0.5 bg-slate-50 border border-gray-100 rounded text-slate-600 text-xs">{kw}</span>
                    ))}
                  </div>
                </div>
                
                <div>
                  <div className="text-slate-500 text-xs mb-1">项目经历</div>
                  <div className="font-medium text-slate-800">2 个项目经历</div>
                </div>
                
                <div className="pt-2 border-t border-gray-100">
                  <div className="text-slate-500 text-xs mb-1">简历摘要</div>
                  <p className="text-slate-600 leading-relaxed text-xs">
                    该版本覆盖基础后端开发能力，但项目经历较少，技术关键词覆盖不完整，对 Go 后端、Redis、微服务等方向的证据较弱。
                  </p>
                </div>

                <div className="pt-2">
                  <div className="text-slate-500 text-xs mb-1.5">主要不足</div>
                  <ul className="text-xs text-slate-600 space-y-1.5 list-disc pl-4 marker:text-slate-300">
                    <li>项目经历数量较少</li>
                    <li>Redis / 微服务证据不足</li>
                    <li>与后台开发岗位 JD 的关键词匹配不够完整</li>
                    <li>缺少简历优化后的投递重点表达</li>
                  </ul>
                </div>
              </div>
            </div>
          </div>

          {/* Middle Column: Diff Analysis */}
          <div className="lg:col-span-6 space-y-6">
            
            {/* Module 1: Overview */}
            <section className="bg-white rounded-xl border border-gray-200 p-6 shadow-sm">
              <h2 className="text-base font-semibold text-slate-900 mb-4 flex items-center">
                <FileText className="w-4 h-4 mr-2 text-[#0052D9]" />
                版本变化总览
              </h2>
              
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <div className="bg-[#F5F7FA] rounded-xl p-5 border border-gray-100 flex flex-col items-center justify-center text-center relative overflow-hidden">
                  <div className="text-xs text-slate-500 mb-3 font-medium relative z-10">关键词覆盖</div>
                  <div className="flex items-center justify-center space-x-3 relative z-10">
                    <div className="text-slate-400 flex items-baseline">
                      <span className="text-xl font-semibold line-through">3</span><span className="text-xs ml-0.5">个</span>
                    </div>
                    <div className="bg-white rounded-full p-1 shadow-sm border border-gray-100">
                      <ArrowRight className="w-3.5 h-3.5 text-slate-400" />
                    </div>
                    <div className="text-[#0052D9] flex items-baseline">
                      <span className="text-3xl font-bold">8</span><span className="text-xs ml-0.5 font-medium">个</span>
                    </div>
                  </div>
                </div>

                <div className="bg-[#F5F7FA] rounded-xl p-5 border border-gray-100 flex flex-col items-center justify-center text-center relative overflow-hidden">
                  <div className="text-xs text-slate-500 mb-3 font-medium relative z-10">项目经历</div>
                  <div className="flex items-center justify-center space-x-3 relative z-10">
                    <div className="text-slate-400 flex items-baseline">
                      <span className="text-xl font-semibold line-through">2</span><span className="text-xs ml-0.5">个</span>
                    </div>
                    <div className="bg-white rounded-full p-1 shadow-sm border border-gray-100">
                      <ArrowRight className="w-3.5 h-3.5 text-slate-400" />
                    </div>
                    <div className="text-[#0052D9] flex items-baseline">
                      <span className="text-3xl font-bold">3</span><span className="text-xs ml-0.5 font-medium">个</span>
                    </div>
                  </div>
                </div>

                <div className="bg-[#F5F7FA] rounded-xl p-5 border border-gray-100 flex flex-col items-center justify-center text-center relative overflow-hidden">
                  <div className="text-xs text-slate-500 mb-3 font-medium relative z-10">适配方向</div>
                  <div className="flex flex-col items-center justify-center space-y-1.5 relative z-10 w-full">
                    <div className="text-slate-400 text-sm line-through">
                      通用后端
                    </div>
                    <div className="text-[#0052D9] text-sm font-semibold whitespace-nowrap">
                      后端开发 / 云计算开发
                    </div>
                  </div>
                </div>
              </div>
            </section>

            {/* Module 2: Added Content */}
            <section className="bg-white rounded-xl border border-gray-200 p-6 shadow-sm">
              <h2 className="text-base font-semibold text-slate-900 mb-4 flex items-center">
                <Plus className="w-4 h-4 mr-2 text-emerald-500" />
                新增内容
              </h2>
              <div className="space-y-4">
                <div className="flex items-start">
                  <div className="bg-emerald-50 p-1.5 rounded mt-0.5 mr-3">
                    <CheckCircle2 className="w-3.5 h-3.5 text-emerald-600" />
                  </div>
                  <div>
                    <h4 className="text-sm font-medium text-slate-900 mb-1">新增缓存优化项目</h4>
                    <p className="text-xs text-slate-600 leading-relaxed">补充 Redis / MySQL 查询优化相关项目经历，增强缓存和性能优化证据。</p>
                  </div>
                </div>
                <div className="flex items-start">
                  <div className="bg-emerald-50 p-1.5 rounded mt-0.5 mr-3">
                    <CheckCircle2 className="w-3.5 h-3.5 text-emerald-600" />
                  </div>
                  <div>
                    <h4 className="text-sm font-medium text-slate-900 mb-1">新增 Docker / 微服务关键词</h4>
                    <p className="text-xs text-slate-600 leading-relaxed">让简历更容易与后台开发和云计算开发岗位 JD 对齐。</p>
                  </div>
                </div>
                <div className="flex items-start">
                  <div className="bg-emerald-50 p-1.5 rounded mt-0.5 mr-3">
                    <CheckCircle2 className="w-3.5 h-3.5 text-emerald-600" />
                  </div>
                  <div>
                    <h4 className="text-sm font-medium text-slate-900 mb-1">新增 RBAC 权限管理表达</h4>
                    <p className="text-xs text-slate-600 leading-relaxed">强化后台系统权限控制、接口鉴权和可维护性相关经历。</p>
                  </div>
                </div>
              </div>
            </section>

            {/* Module 3: Strengthened Content */}
            <section className="bg-white rounded-xl border border-gray-200 p-6 shadow-sm">
              <h2 className="text-base font-semibold text-slate-900 mb-4 flex items-center">
                <ArrowUpCircle className="w-4 h-4 mr-2 text-blue-500" />
                强化内容
              </h2>
              <div className="space-y-4">
                <div className="flex items-start">
                  <div className="bg-blue-50 p-1.5 rounded mt-0.5 mr-3">
                    <ArrowUpCircle className="w-3.5 h-3.5 text-blue-600" />
                  </div>
                  <div>
                    <h4 className="text-sm font-medium text-slate-900 mb-1">Go 微服务订单系统表达更清晰</h4>
                    <p className="text-xs text-slate-600 leading-relaxed">从“完成接口开发”强化为“接口设计、数据库建模、Redis 缓存优化、错误处理和权限校验”。</p>
                  </div>
                </div>
                <div className="flex items-start">
                  <div className="bg-blue-50 p-1.5 rounded mt-0.5 mr-3">
                    <ArrowUpCircle className="w-3.5 h-3.5 text-blue-600" />
                  </div>
                  <div>
                    <h4 className="text-sm font-medium text-slate-900 mb-1">MySQL / Redis 使用证据更明确</h4>
                    <p className="text-xs text-slate-600 leading-relaxed">从普通技术栈展示，强化为具体项目中的缓存优化和查询优化能力。</p>
                  </div>
                </div>
                <div className="flex items-start">
                  <div className="bg-blue-50 p-1.5 rounded mt-0.5 mr-3">
                    <ArrowUpCircle className="w-3.5 h-3.5 text-blue-600" />
                  </div>
                  <div>
                    <h4 className="text-sm font-medium text-slate-900 mb-1">求职方向更聚焦</h4>
                    <p className="text-xs text-slate-600 leading-relaxed">从“通用后端”聚焦为“后端开发方向”。</p>
                  </div>
                </div>
              </div>
            </section>

            {/* Module 4: Keyword Changes Table */}
            <section className="bg-white rounded-xl border border-gray-200 p-6 shadow-sm overflow-hidden">
              <h2 className="text-base font-semibold text-slate-900 mb-4">关键词变化</h2>
              <div className="overflow-x-auto border border-gray-100 rounded-lg">
                <table className="w-full text-sm text-left">
                  <thead className="bg-slate-50 text-slate-500 text-xs uppercase border-b border-gray-100">
                    <tr>
                      <th className="px-4 py-3 font-medium">关键词</th>
                      <th className="px-4 py-3 font-medium">V1 状态</th>
                      <th className="px-4 py-3 font-medium">V2 状态</th>
                      <th className="px-4 py-3 font-medium">变化说明</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-100">
                    <tr className="hover:bg-slate-50/50">
                      <td className="px-4 py-3 font-medium text-slate-800">Go</td>
                      <td className="px-4 py-3 text-slate-500">已覆盖</td>
                      <td className="px-4 py-3 text-blue-600 font-medium">已强化</td>
                      <td className="px-4 py-3 text-slate-600 text-xs">项目表达更具体</td>
                    </tr>
                    <tr className="hover:bg-slate-50/50">
                      <td className="px-4 py-3 font-medium text-slate-800">MySQL</td>
                      <td className="px-4 py-3 text-slate-500">已覆盖</td>
                      <td className="px-4 py-3 text-blue-600 font-medium">已强化</td>
                      <td className="px-4 py-3 text-slate-600 text-xs">增加数据库设计和查询优化描述</td>
                    </tr>
                    <tr className="hover:bg-slate-50/50">
                      <td className="px-4 py-3 font-medium text-slate-800">Redis</td>
                      <td className="px-4 py-3 text-slate-400">弱覆盖</td>
                      <td className="px-4 py-3 text-blue-600 font-medium">已强化</td>
                      <td className="px-4 py-3 text-slate-600 text-xs">新增缓存优化项目</td>
                    </tr>
                    <tr className="hover:bg-slate-50/50 bg-emerald-50/30">
                      <td className="px-4 py-3 font-medium text-slate-800">Docker</td>
                      <td className="px-4 py-3 text-slate-400">未覆盖</td>
                      <td className="px-4 py-3 text-emerald-600 font-medium">新增</td>
                      <td className="px-4 py-3 text-slate-600 text-xs">补充部署与工程化关键词</td>
                    </tr>
                    <tr className="hover:bg-slate-50/50">
                      <td className="px-4 py-3 font-medium text-slate-800">微服务</td>
                      <td className="px-4 py-3 text-slate-400">弱覆盖</td>
                      <td className="px-4 py-3 text-blue-600 font-medium">已强化</td>
                      <td className="px-4 py-3 text-slate-600 text-xs">强化服务拆分与接口设计表达</td>
                    </tr>
                    <tr className="hover:bg-slate-50/50 bg-emerald-50/30">
                      <td className="px-4 py-3 font-medium text-slate-800">RBAC</td>
                      <td className="px-4 py-3 text-slate-400">未覆盖</td>
                      <td className="px-4 py-3 text-emerald-600 font-medium">新增</td>
                      <td className="px-4 py-3 text-slate-600 text-xs">新增权限管理系统经历</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </section>

            {/* Module 5: Project Experience Changes */}
            <section className="bg-white rounded-xl border border-gray-200 p-6 shadow-sm">
              <h2 className="text-base font-semibold text-slate-900 mb-4">项目经历变化</h2>
              <div className="space-y-4">
                <div className="border border-gray-100 rounded-lg p-4 bg-slate-50">
                  <div className="flex justify-between items-start mb-2">
                    <h4 className="text-sm font-semibold text-slate-800">1. Go 微服务订单系统</h4>
                    <span className="text-[10px] bg-blue-100 text-blue-700 px-2 py-0.5 rounded border border-blue-200">强化</span>
                  </div>
                  <p className="text-xs text-slate-600 leading-relaxed">
                    <span className="text-slate-500 line-through mr-1">普通接口开发描述</span> 
                    <ArrowRight className="w-3 h-3 inline text-slate-400 mx-1" />
                    <span className="font-medium text-slate-700">强化为包含接口设计、数据库设计、Redis 缓存、错误处理和权限校验的完整后端项目表达。</span>
                  </p>
                </div>
                
                <div className="border border-gray-100 rounded-lg p-4 bg-slate-50">
                  <div className="flex justify-between items-start mb-2">
                    <h4 className="text-sm font-semibold text-slate-800">2. RBAC 权限管理系统</h4>
                    <span className="text-[10px] bg-blue-100 text-blue-700 px-2 py-0.5 rounded border border-blue-200">强化</span>
                  </div>
                  <p className="text-xs text-slate-600 leading-relaxed">
                    <span className="text-slate-500 line-through mr-1">功能描述</span>
                    <ArrowRight className="w-3 h-3 inline text-slate-400 mx-1" />
                    <span className="font-medium text-slate-700">强化为权限模型、接口鉴权、后台管理系统可维护性表达。</span>
                  </p>
                </div>

                <div className="border border-emerald-100 rounded-lg p-4 bg-emerald-50/50">
                  <div className="flex justify-between items-start mb-2">
                    <h4 className="text-sm font-semibold text-slate-800">3. 缓存优化项目</h4>
                    <span className="text-[10px] bg-emerald-100 text-emerald-700 px-2 py-0.5 rounded border border-emerald-200">新增</span>
                  </div>
                  <p className="text-xs text-slate-600 leading-relaxed">
                    V2 新增项目，用于补充 Redis / MySQL 查询优化证据。
                  </p>
                </div>
              </div>
            </section>

            {/* Module 6: Remaining Risks */}
            <section className="bg-orange-50/50 rounded-xl border border-orange-100 p-6">
              <h2 className="text-base font-semibold text-orange-800 mb-3 flex items-center">
                <AlertCircle className="w-4 h-4 mr-2 text-orange-500" />
                仍然需要补强的内容
              </h2>
              <ul className="text-sm text-orange-700 space-y-2 list-disc pl-5 marker:text-orange-400 mb-3">
                <li>Linux 使用经验仍然描述不足</li>
                <li>高并发压测数据仍然缺失</li>
                <li>线上问题排查案例仍然较少</li>
              </ul>
              <p className="text-xs text-orange-600/80">
                这些问题后续可以在岗位匹配分析和简历优化页继续处理。
              </p>
            </section>

            {/* Module 7: Conclusion & Action */}
            <section className="bg-blue-50/50 rounded-xl border border-blue-100 p-6">
              <h2 className="text-base font-semibold text-blue-900 mb-3">建议结论</h2>
              <p className="text-sm text-blue-800 mb-5 leading-relaxed">
                建议将 <span className="font-semibold">{comparison?.targetName || "V2 后端开发简历"}</span> 设为当前简历，并优先用于后台开发、云计算开发等技术岗位的岗位推荐和岗位匹配分析。
              </p>
              <div className="flex flex-wrap gap-3">
                <button className="bg-[#0052D9] hover:bg-[#0047BA] text-white px-5 py-2 rounded shadow-sm text-sm font-medium transition-colors">
                  设为当前简历
                </button>
                <button className="bg-white border border-blue-200 text-blue-700 hover:bg-blue-50 px-4 py-2 rounded text-sm font-medium transition-colors">
                  用于岗位推荐
                </button>
                <button className="bg-white border border-blue-200 text-blue-700 hover:bg-blue-50 px-4 py-2 rounded text-sm font-medium transition-colors">
                  用于岗位匹配分析
                </button>
              </div>
            </section>

          </div>

          {/* Right Column: New Version Summary */}
          <div className="lg:col-span-3">
            <div className="bg-white rounded-xl border border-blue-200 ring-1 ring-blue-50 p-5 shadow-md sticky top-20">
              <div className="mb-4">
                <h3 className="font-semibold text-slate-900 mb-2">新版本：{comparison?.targetName || "V2 后端开发简历"}</h3>
                <div className="flex gap-2">
                  <span className="inline-flex px-2 py-0.5 rounded text-xs font-medium bg-blue-100 text-blue-700 border border-blue-200">
                    当前使用
                  </span>
                  <span className="inline-flex px-2 py-0.5 rounded text-xs font-medium bg-emerald-50 text-emerald-600 border border-emerald-100">
                    可用于 AI 分析
                  </span>
                </div>
              </div>
              
              <div className="space-y-4 text-sm">
                <div>
                  <div className="text-slate-500 text-xs mb-1">求职方向</div>
                  <div className="font-medium text-slate-800">后端开发 / 云计算开发</div>
                </div>
                
                <div>
                  <div className="text-slate-500 text-xs mb-1">关键词 (8)</div>
                  <div className="flex flex-wrap gap-1.5">
                    {["Go", "MySQL", "Redis", "Docker", "微服务", "RBAC", "接口开发", "缓存优化"].map(kw => (
                      <span key={kw} className="px-2 py-0.5 bg-blue-50 border border-blue-100 rounded text-blue-700 text-xs">{kw}</span>
                    ))}
                  </div>
                </div>
                
                <div>
                  <div className="text-slate-500 text-xs mb-1">项目经历</div>
                  <div className="font-medium text-slate-800">3 个项目经历</div>
                </div>
                
                <div className="pt-2 border-t border-gray-100">
                  <div className="text-slate-500 text-xs mb-1">简历摘要</div>
                  <p className="text-slate-600 leading-relaxed text-xs">
                    该版本强化了 Go 后端、MySQL / Redis、微服务接口设计、RBAC 权限管理和缓存优化等内容，更适合后台开发、云计算开发等技术岗位。
                  </p>
                </div>

                <div className="pt-2">
                  <div className="text-slate-500 text-xs mb-1.5">主要增强</div>
                  <ul className="text-xs text-slate-600 space-y-1.5 list-disc pl-4 marker:text-blue-400">
                    <li>新增缓存优化项目</li>
                    <li>强化 Go 微服务项目表达</li>
                    <li>增加 Redis、Docker 等关键词</li>
                    <li>更贴合后台开发岗位要求</li>
                  </ul>
                </div>
              </div>
            </div>
          </div>

        </div>
      </div>

      {/* Bottom Sticky Action Bar */}
      <div className="fixed bottom-0 left-0 right-0 bg-white border-t border-gray-200 p-4 shadow-[0_-4px_6px_-1px_rgba(0,0,0,0.05)] z-40">
        <div className="max-w-[1360px] mx-auto px-6 flex flex-col md:flex-row items-center justify-between gap-4">
          <div className="flex flex-col md:flex-row md:items-center gap-2 md:gap-6 text-sm">
            <div className="text-slate-500 flex items-center">
              当前比较：
              <span className="font-medium text-slate-700 ml-1">{comparison?.baseName || "V1 通用后端简历"}</span>
              <ArrowRight className="w-3.5 h-3.5 mx-2 text-slate-400" />
              <span className="font-medium text-blue-600">{comparison?.targetName || "V2 后端开发简历"}</span>
            </div>
            <div className="text-xs text-slate-500 border-l-0 md:border-l border-gray-200 md:pl-6">
              <span className="hidden md:inline">推荐下一步：</span>1. 设为当前简历 <span className="mx-1 text-slate-300">|</span> 2. 进行岗位推荐 <span className="mx-1 text-slate-300">|</span> 3. 匹配分析
            </div>
          </div>
          
          <div className="flex items-center gap-3 w-full md:w-auto">
            <Link to="/resumes" className="flex-1 md:flex-none px-4 py-2 text-center text-sm font-medium text-slate-600 bg-white border border-gray-300 rounded hover:bg-gray-50 transition-colors">
              返回我的简历
            </Link>
            <Link to="/recommend" className="flex-1 md:flex-none px-4 py-2 text-center text-sm font-medium text-[#0052D9] bg-blue-50 border border-blue-200 rounded hover:bg-blue-100 transition-colors">
              用于岗位推荐
            </Link>
            <button className="flex-1 md:flex-none px-6 py-2 text-sm font-medium text-white bg-[#0052D9] rounded shadow-sm hover:bg-[#0047BA] transition-colors shadow-blue-500/20">
              设为当前简历
            </button>
          </div>
        </div>
      </div>

    </>
  );
}
