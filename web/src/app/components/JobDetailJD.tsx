import React from "react";

export function JobDetailJD() {
  return (
    <div className="flex-1 bg-white border border-gray-200 rounded-lg p-10 min-w-[760px] max-w-[820px] shadow-sm">
      <div className="space-y-10 text-gray-700 text-[15px] leading-relaxed">
        {/* Section 1 */}
        <section>
          <h3 className="font-bold text-gray-900 mb-4 text-[17px] flex items-center gap-2">
            <span className="w-1 h-4 bg-[#0052D9] rounded-full inline-block"></span>
            岗位描述
          </h3>
          <ul className="list-disc pl-5 space-y-2.5 marker:text-gray-300">
            <li>负责后台服务的设计、开发和性能优化，支撑高并发、高可用业务场景；</li>
            <li>参与系统架构设计，提升服务稳定性、可扩展性和可维护性；</li>
            <li>与产品、算法、客户端、测试等团队协作，完成业务需求落地；</li>
            <li>参与线上问题定位和性能优化，保障系统稳定运行；</li>
            <li>关注工程质量，推动代码规范、接口设计和服务治理优化。</li>
          </ul>
        </section>

        {/* Section 2 */}
        <section>
          <h3 className="font-bold text-gray-900 mb-4 text-[17px] flex items-center gap-2">
            <span className="w-1 h-4 bg-[#0052D9] rounded-full inline-block"></span>
            岗位要求
          </h3>
          <p className="mb-3 font-medium text-gray-800">必须具备：</p>
          <ul className="list-disc pl-5 space-y-2.5 marker:text-gray-300">
            <li>扎实的计算机基础，熟悉数据结构、操作系统、网络、数据库等基础知识；</li>
            <li>熟悉 C/C++ / Java / Go 中至少一门开发语言；</li>
            <li>理解 TCP / UDP、HTTP、RPC、缓存、数据库等常见后端技术；</li>
            <li>具备良好的编码习惯、问题定位能力和学习能力；</li>
            <li>有项目实践或后端系统开发经验者优先。</li>
          </ul>
        </section>

        {/* Section 3 */}
        <section>
          <h3 className="font-bold text-gray-900 mb-4 text-[17px] flex items-center gap-2">
            <span className="w-1 h-4 bg-[#0052D9] rounded-full inline-block"></span>
            加分项或注意事项
          </h3>
          <ul className="list-disc pl-5 space-y-2.5 marker:text-gray-300">
            <li>有高并发、分布式系统、微服务、缓存、消息队列相关经验；</li>
            <li>有云原生、容器、Kubernetes、DevOps 相关经验；</li>
            <li>有开源项目、竞赛、实习或真实项目经验；</li>
            <li>对复杂业务系统和工程质量有持续优化意识。</li>
          </ul>
        </section>

        {/* Section 4 */}
        <section>
          <h3 className="font-bold text-gray-900 mb-4 text-[17px] flex items-center gap-2">
            <span className="w-1 h-4 bg-[#0052D9] rounded-full inline-block"></span>
            工作地点
          </h3>
          <p className="pl-3">深圳 / 北京 / 上海 / 广州 / 成都</p>
        </section>

        <div className="pt-6 mt-6 border-t border-gray-100"></div>

        {/* Section 5 */}
        <section className="bg-[#F4F8FF] p-5 rounded-lg border border-blue-100">
          <h3 className="font-bold text-[#0052D9] mb-2 text-sm flex items-center gap-1.5">
            投递说明
          </h3>
          <p className="text-sm text-gray-600 leading-relaxed">
            该岗位支持使用当前简历进行 AI 匹配分析，分析结果仅作为投递前决策参考，不替代用户最终选择。
          </p>
        </section>
      </div>
    </div>
  );
}
