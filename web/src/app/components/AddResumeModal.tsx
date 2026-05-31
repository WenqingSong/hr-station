import React, { useState } from "react";
import { X } from "lucide-react";

interface AddResumeModalProps {
  isOpen: boolean;
  onClose: () => void;
}

const inputCls =
  "w-full px-3 py-2 border border-gray-200 rounded-md text-sm text-gray-900 bg-white focus:outline-none focus:ring-1 focus:ring-[#0052D9] focus:border-[#0052D9] transition-colors placeholder-gray-400";

const CREATE_METHODS = [
  {
    value: "blank",
    label: "从空白简历创建",
    desc: "从零开始填写所有简历信息",
  },
  {
    value: "copy",
    label: "复制当前简历创建",
    desc: "以「V2 后端开发简历」为基础，复制后独立编辑",
  },
  {
    value: "paste",
    label: "粘贴简历文本快速创建",
    desc: "将已有简历文字粘贴后由系统解析填入各字段",
  },
];

export function AddResumeModal({ isOpen, onClose }: AddResumeModalProps) {
  const [resumeName, setResumeName] = useState("");
  const [jobDirection, setJobDirection] = useState("后端开发");
  const [createMethod, setCreateMethod] = useState("blank");
  const [setAsCurrent, setSetAsCurrent] = useState(true);

  if (!isOpen) return null;

  const handleCreate = () => {
    if (!resumeName.trim()) return;
    onClose();
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      {/* Overlay */}
      <div
        className="absolute inset-0 bg-black/40 backdrop-blur-[1px]"
        onClick={onClose}
      />

      {/* Modal */}
      <div className="relative bg-white rounded-xl shadow-2xl w-full max-w-[500px] mx-4 z-10">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-5 border-b border-gray-100">
          <h2 className="text-base text-gray-900">新增简历</h2>
          <button
            onClick={onClose}
            className="p-1 text-gray-400 hover:text-gray-600 transition-colors rounded-md hover:bg-gray-50"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Body */}
        <div className="px-6 py-5 space-y-5">
          <p className="text-sm text-gray-500 leading-relaxed">
            创建一份新的简历版本，用于岗位推荐、岗位匹配分析和简历优化。
          </p>

          {/* Resume Name */}
          <div>
            <label className="block text-sm text-gray-600 mb-1.5">
              简历名称 <span className="text-red-400">*</span>
            </label>
            <input
              className={inputCls}
              placeholder="例如：V4 后端开发简历"
              value={resumeName}
              onChange={(e) => setResumeName(e.target.value)}
              autoFocus
            />
          </div>

          {/* Job Direction */}
          <div>
            <label className="block text-sm text-gray-600 mb-1.5">求职方向</label>
            <select
              className={inputCls}
              value={jobDirection}
              onChange={(e) => setJobDirection(e.target.value)}
            >
              <option>后端开发</option>
              <option>云计算开发</option>
              <option>算法与大数据</option>
              <option>产品策划</option>
              <option>设计</option>
              <option>前端开发</option>
            </select>
          </div>

          {/* Create Method */}
          <div>
            <label className="block text-sm text-gray-600 mb-2">创建方式</label>
            <div className="space-y-2">
              {CREATE_METHODS.map((opt) => (
                <label
                  key={opt.value}
                  className={`flex items-start gap-3 px-4 py-3 rounded-md border cursor-pointer transition-colors ${
                    createMethod === opt.value
                      ? "border-[#0052D9] bg-blue-50"
                      : "border-gray-200 hover:border-gray-300 hover:bg-gray-50"
                  }`}
                >
                  <input
                    type="radio"
                    name="createMethod"
                    value={opt.value}
                    checked={createMethod === opt.value}
                    onChange={() => setCreateMethod(opt.value)}
                    className="mt-0.5 accent-[#0052D9] flex-shrink-0"
                  />
                  <div>
                    <p
                      className={`text-sm ${
                        createMethod === opt.value ? "text-[#0052D9]" : "text-gray-800"
                      }`}
                    >
                      {opt.label}
                    </p>
                    <p className="text-xs text-gray-400 mt-0.5 leading-relaxed">
                      {opt.desc}
                    </p>
                  </div>
                </label>
              ))}
            </div>
          </div>

          {/* Set as Current */}
          <label className="flex items-center gap-2.5 cursor-pointer">
            <input
              type="checkbox"
              checked={setAsCurrent}
              onChange={(e) => setSetAsCurrent(e.target.checked)}
              className="accent-[#0052D9] w-4 h-4"
            />
            <span className="text-sm text-gray-700">创建后设为当前简历</span>
          </label>
        </div>

        {/* Footer */}
        <div className="flex items-center justify-end gap-2 px-6 py-4 border-t border-gray-100">
          <button
            onClick={onClose}
            className="px-4 py-2 text-sm text-gray-600 border border-gray-200 rounded-md hover:bg-gray-50 transition-colors"
          >
            取消
          </button>
          <button
            onClick={handleCreate}
            disabled={!resumeName.trim()}
            className="px-4 py-2 text-sm text-white bg-[#0052D9] rounded-md hover:bg-blue-700 transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
          >
            创建简历
          </button>
        </div>
      </div>
    </div>
  );
}
