package service

import (
	"testing"

	"offer-catcher/server/internal/model"
)

// Test 1: Ordinary backend job with no education requirement → filter out education advice
func TestFilterEducationAdvice_OrdinaryBackendJob_NoEduReq(t *testing.T) {
	job := &model.Job{
		Title:        "软件开发-后台开发方向",
		Direction:    "软件开发",
		Description:  "负责实现和优化公司的产品功能",
		Requirements: "扎实的编程能力，熟练掌握 Go/Java，熟悉 TCP/UDP 协议，了解 MySQL/Redis",
	}
	resume := &model.Resume{
		RawText: "教育经历：清华大学 计算机科学与技术 本科 GPA 3.8",
		Educations: []model.ResumeEducation{
			{School: "清华大学", Degree: "本科", Major: "计算机科学与技术"},
		},
	}
	result := &model.OptimizationResultV2{
		OptimizationScore: 75,
		RiskPoints: []string{
			"简历未明确标注学历信息，不符合岗位学历要求",
			"项目经历描述未突出与岗位要求的匹配度",
		},
		QuickFixItems: []model.OptimizationFixItem{
			{Section: "教育背景", Issue: "未明确标注学历层次", Action: "补充本科、硕士学历信息"},
			{Section: "项目经历", Issue: "缺少量化指标", Action: "补充QPS等量化数据"},
		},
		ContentFixItems: []model.OptimizationFixItem{
			{Section: "教育背景", Issue: "无学历层次、院校信息", Action: "补充学历信息"},
		},
		RewriteExamples: []model.OptimizationRewriteExample{
			{Section: "教育背景", Before: "清华大学", After: "清华大学 计算机本科 硕士 GPA 3.8", Reason: "完善教育信息"},
		},
		DeliveryFocus: "请重点检查学历信息是否完整",
	}

	filterEducationOptimizationAdvice(result, resume, job)

	// Should NOT allow education advice for this backend job
	if shouldAllowEducationAdvice(resume, job) {
		t.Error("shouldAllowEducationAdvice returned true for ordinary backend job with no high-edu JD")
	}

	// All risk_points about education should be filtered
	for _, r := range result.RiskPoints {
		if isEducationAdviceText(r) {
			t.Errorf("risk_point still contains education advice after filter: %s", r)
		}
	}
	if len(result.RiskPoints) != 1 {
		t.Errorf("expected 1 risk_point after filter, got %d: %v", len(result.RiskPoints), result.RiskPoints)
	}

	// QuickFixItems: education fix should be filtered
	for _, item := range result.QuickFixItems {
		combined := item.Section + " " + item.Issue + " " + item.Action
		if isEducationAdviceText(combined) {
			t.Errorf("quick_fix still contains education advice: %s / %s", item.Section, item.Issue)
		}
	}
	if len(result.QuickFixItems) != 1 {
		t.Errorf("expected 1 quick_fix after filter, got %d", len(result.QuickFixItems))
	}

	// ContentFixItems: education should be filtered
	if len(result.ContentFixItems) != 0 {
		t.Errorf("expected 0 content_fix after filter, got %d: %v", len(result.ContentFixItems), result.ContentFixItems)
	}

	// RewriteExamples: education rewrite should be filtered
	if len(result.RewriteExamples) != 0 {
		t.Errorf("expected 0 rewrite_examples after filter, got %d", len(result.RewriteExamples))
	}

	// DeliveryFocus: should be replaced
	if isEducationAdviceText(result.DeliveryFocus) {
		t.Errorf("delivery_focus still has education advice: %s", result.DeliveryFocus)
	}
}

// Test 2: JD has "硕士及以上" → allow education advice
func TestFilterEducationAdvice_ResearchJob_MasterRequired(t *testing.T) {
	job := &model.Job{
		Title:        "算法研究员-多模态方向",
		Direction:    "算法研究",
		Description:  "负责多模态大模型的研究与开发",
		Requirements: "硕士及以上学历，博士优先，有顶会论文发表经验",
	}
	resume := &model.Resume{
		RawText: "教育经历：某大学 计算机科学 本科",
		Educations: []model.ResumeEducation{
			{School: "某大学", Degree: "本科", Major: "计算机科学"},
		},
	}
	result := &model.OptimizationResultV2{
		RiskPoints: []string{"学历不满足岗位硕士及以上要求"},
		QuickFixItems: []model.OptimizationFixItem{
			{Section: "教育背景", Issue: "学历层次不足", Action: "岗位要求硕士及以上，当前为本科"},
		},
	}

	// Should ALLOW education advice for this research job
	if !shouldAllowEducationAdvice(resume, job) {
		t.Error("shouldAllowEducationAdvice returned false for research job with explicit master requirement")
	}

	filterEducationOptimizationAdvice(result, resume, job)

	// All education advice should be preserved
	if len(result.RiskPoints) != 1 {
		t.Error("risk_points for research job should be preserved")
	}
	if len(result.QuickFixItems) != 1 {
		t.Error("quick_fix_items for research job should be preserved")
	}
}

// Test 3: JD has "本科及以上" + resume has bachelor → do NOT allow education advice
func TestFilterEducationAdvice_BachelorOrAbove_HasBachelor(t *testing.T) {
	job := &model.Job{
		Title:        "前端开发工程师",
		Requirements: "本科及以上学历，计算机相关专业优先",
	}
	resume := &model.Resume{
		RawText: "教育经历：某大学 软件工程 本科",
		Educations: []model.ResumeEducation{
			{School: "某大学", Degree: "本科", Major: "软件工程"},
		},
	}
	result := &model.OptimizationResultV2{
		RiskPoints: []string{"学历要求为本科及以上，当前学历不明确"},
		QuickFixItems: []model.OptimizationFixItem{
			{Section: "教育背景", Issue: "未明确标注学历层次", Action: "补充学历信息"},
		},
	}

	if shouldAllowEducationAdvice(resume, job) {
		t.Error("shouldAllowEducationAdvice should return false when JD only says '本科及以上' and resume has bachelor")
	}

	filterEducationOptimizationAdvice(result, resume, job)
	if len(result.RiskPoints) != 0 {
		t.Errorf("risk_points should be empty, got: %v", result.RiskPoints)
	}
	if len(result.QuickFixItems) != 0 {
		t.Errorf("quick_fix_items should be empty, got: %v", result.QuickFixItems)
	}
}

// Test 4: isEducationAdviceText keyword matching
func TestIsEducationAdviceText(t *testing.T) {
	tests := []struct {
		text     string
		expected bool
	}{
		{"学历要求不满足", true},
		{"硕士学历要求不满足", true},
		{"博士优先", true},
		{"985院校优先", true},
		{"211院校", true},
		{"双一流", true},
		{"名校背景", true},
		{"教育经历缺失", true},
		{"学历背景不足", true},
		{"学历层次不清晰", true},
		{"学历风险", true},
		{"学位不足", true},
		{"研究生要求", true},
		{"学校背景不匹配", true},
		// Should NOT match technical terms
		{"熟练掌握Java、MySQL、Redis", false},
		{"项目缺少量化指标", false},
		{"专业技能描述不够具体", false},
		{"接口QPS需优化", false},
		{"Go后端开发经验丰富", false},
	}
	for _, tt := range tests {
		got := isEducationAdviceText(tt.text)
		if got != tt.expected {
			t.Errorf("isEducationAdviceText(%q) = %v, want %v", tt.text, got, tt.expected)
		}
	}
}

// Test 5: hasResumeEducation detection
func TestHasResumeEducation(t *testing.T) {
	tests := []struct {
		resume   *model.Resume
		expected bool
	}{
		{
			&model.Resume{Educations: []model.ResumeEducation{{School: "清华", Degree: "本科"}}},
			true,
		},
		{
			&model.Resume{School: "清华", Degree: "本科"},
			true,
		},
		{
			&model.Resume{RawText: "本科 计算机科学 教育经历"},
			true,
		},
		{
			&model.Resume{RawText: "熟练掌握 Go 和 MySQL"},
			false,
		},
		{
			&model.Resume{},
			false,
		},
		{
			nil,
			false,
		},
	}
	for i, tt := range tests {
		got := hasResumeEducation(tt.resume)
		if got != tt.expected {
			t.Errorf("test %d: hasResumeEducation = %v, want %v", i, got, tt.expected)
		}
	}
}
