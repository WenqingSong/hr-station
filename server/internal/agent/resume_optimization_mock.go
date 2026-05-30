package agent

import (
	"context"
	"fmt"
	"strings"

	"offer-catcher/server/internal/model"
)

// ResumeOptimizationMockAgent generates mock resume optimization using rule-based analysis.
// It does NOT call any real LLM, Prompt, Schema, or Pipeline.
type ResumeOptimizationMockAgent struct{}

// NewResumeOptimizationMockAgent creates a new ResumeOptimizationMockAgent.
func NewResumeOptimizationMockAgent() *ResumeOptimizationMockAgent {
	return &ResumeOptimizationMockAgent{}
}

// OptimizeInput contains the input for the resume optimization mock agent.
type OptimizeInput struct {
	ResumeText     string
	JobTitle       string
	JDText         string
	TargetSection  string
	TargetRoles    []string
	ExpectedLocations []string
	AcceptRelocation  bool
}

// Optimize generates a mock resume optimization based on resume text and target job JD.
func (a *ResumeOptimizationMockAgent) Optimize(_ context.Context, input *OptimizeInput) (*model.ResumeOptimizationResult, error) {
	resumeLower := strings.ToLower(input.ResumeText)
	jdLower := strings.ToLower(input.JDText + " " + input.JobTitle)

	// Determine apply_readiness based on resume and JD content
	applyReadiness, stopSignal := determineApplyReadiness(input.ResumeText, jdLower)

	// Generate JD evidence chain
	evidenceChain := generateJDEvidenceChain(resumeLower, jdLower, input.JobTitle)

	// Generate top fix items (max 3)
	topFixItems := generateTopFixItems(resumeLower, jdLower, input.JobTitle, input.ResumeText)

	// Generate long-term improvements
	longTermItems := generateLongTermImprovements(resumeLower, jdLower)

	// Generate summary
	summary := generateOptimizationSummary(applyReadiness, len(topFixItems), input.JobTitle)

	notice := "以下建议仅作为简历表达和投递策略参考，AI 可能理解错误。请基于你的真实经历核实后使用，不要加入没有实际发生过的经历、指标或职责。"

	return &model.ResumeOptimizationResult{
		AgentStatus:            "mock_success",
		GeneratedBy:            "mock_agent",
		FallbackRequired:       false,
		RepairApplied:          false,
		ApplyReadiness:         applyReadiness,
		OptimizationStopSignal: stopSignal,
		Summary:                summary,
		TopFixItems:            topFixItems,
		LongTermImprovement:    longTermItems,
		JDEvidenceChain:        evidenceChain,
		VerificationNotice:     notice,
	}, nil
}

// determineApplyReadiness determines apply_readiness and optimization_stop_signal
// based on resume text length and JD keyword overlap.
func determineApplyReadiness(resumeText string, jdLower string) (string, string) {
	resumeLen := len(strings.TrimSpace(resumeText))

	// Very short resume — insufficient information
	if resumeLen < 50 {
		return "insufficient_information", "needs_more_evidence"
	}

	// Check for some keyword overlap
	backendKws := []string{"go", "mysql", "redis", "docker", "后端", "接口", "数据库", "微服务", "分布式"}
	overlap := 0
	resumeLower := strings.ToLower(resumeText)
	for _, kw := range backendKws {
		if strings.Contains(resumeLower, kw) && strings.Contains(jdLower, kw) {
			overlap++
		}
	}

	// Strong overlap — ready to apply
	if overlap >= 4 && resumeLen > 200 {
		return "ready_to_apply", "stop_and_apply"
	}

	// Some overlap — quick fix before apply
	if overlap >= 2 && resumeLen > 100 {
		return "quick_fix_before_apply", "fix_top_3_then_apply"
	}

	// Very little overlap — not ready for this role
	if overlap == 0 && resumeLen > 50 {
		return "not_ready_for_this_role", "target_role_mismatch"
	}

	// Default: quick fix
	return "quick_fix_before_apply", "fix_top_3_then_apply"
}

// generateJDEvidenceChain creates JD evidence chain entries.
func generateJDEvidenceChain(resumeLower string, jdLower string, jobTitle string) []model.ResumeOptimizationJDEvidenceItem {
	var chain []model.ResumeOptimizationJDEvidenceItem

	// Check common backend requirements
	type jdCheck struct {
		requirement string
		keywords    []string
	}

	checks := []jdCheck{
		{"后台开发、接口开发经验", []string{"接口", "后端", "api", "restful", "服务端"}},
		{"数据库和缓存经验", []string{"mysql", "redis", "数据库", "缓存", "mongodb"}},
		{"Go 语言开发能力", []string{"go", "golang"}},
		{"微服务或分布式系统经验", []string{"微服务", "分布式", "grpc", "kubernetes", "docker"}},
		{"项目或实习经历", []string{"项目", "实习", "工作", "经历"}},
	}

	for _, check := range checks {
		jdHas := false
		for _, kw := range check.keywords {
			if strings.Contains(jdLower, kw) {
				jdHas = true
				break
			}
		}

		if !jdHas {
			continue
		}

		resumeHas := false
		var evidence string
		for _, kw := range check.keywords {
			if strings.Contains(resumeLower, kw) {
				resumeHas = true
				evidence = fmt.Sprintf("简历中出现 %s 相关内容", kw)
				break
			}
		}

		if resumeHas {
			chain = append(chain, model.ResumeOptimizationJDEvidenceItem{
				JDRequirement:    check.requirement,
				ResumeEvidence:   evidence,
				EvidenceStrength: "medium",
				MatchJudgement:   "partially_matched",
				GapAnalysis:      "有相关技术栈，但个人职责和结果表达不够清晰。",
				ScreeningRisk:    "responsibility_unclear",
			})
		} else {
			chain = append(chain, model.ResumeOptimizationJDEvidenceItem{
				JDRequirement:    check.requirement,
				ResumeEvidence:   "简历中未体现相关经验",
				EvidenceStrength: "missing",
				MatchJudgement:   "missing",
				GapAnalysis:      fmt.Sprintf("岗位要求%s，但简历中未找到相关证据。", check.requirement),
				ScreeningRisk:    "keyword_missing",
			})
		}
	}

	// Ensure at least 2 entries
	if len(chain) < 2 {
		chain = append(chain, model.ResumeOptimizationJDEvidenceItem{
			JDRequirement:    "岗位方向相关经验",
			ResumeEvidence:   "简历中相关信息有限",
			EvidenceStrength: "weak",
			MatchJudgement:   "partially_matched",
			GapAnalysis:      "当前简历信息不足以形成完整证据链。",
			ScreeningRisk:    "scope_unclear",
		})
	}

	return chain
}

// generateTopFixItems generates up to 3 quick fix items.
// Uses "示例参考" instead of "建议改为".
// Does NOT fabricate metrics (QPS, DAU, percentages).
func generateTopFixItems(resumeLower string, jdLower string, jobTitle string, resumeText string) []model.ResumeOptimizationFixItem {
	var items []model.ResumeOptimizationFixItem

	// Fix 1: Project responsibility clarity
	if strings.Contains(jdLower, "后端") || strings.Contains(jdLower, "接口") || strings.Contains(jdLower, "开发") {
		if strings.Contains(resumeLower, "项目") || strings.Contains(resumeLower, "project") {
			items = append(items, model.ResumeOptimizationFixItem{
				Issue:            "项目职责描述不够清晰",
				WhyItMatters:     "岗位 JD 强调后台接口开发和工程实践，当前表达难以证明个人贡献。",
				Action:           "补充你负责的模块、使用的技术栈和最终产出。",
				ExampleReference: "示例参考：基于 GoFrame 完成后台管理系统用户、角色、权限相关接口开发，并与前端完成联调。",
				SelfCheckQuestion: "请确认你是否确实负责过这些模块，而不是团队整体成果。",
				LearnMoreQuery:   "Go 后端实习 简历 项目经历 STAR 写法",
				EstimatedEffort:  "within_30_min",
			})
		} else {
			items = append(items, model.ResumeOptimizationFixItem{
				Issue:            "缺少项目经历",
				WhyItMatters:     "岗位 JD 要求开发经验，当前简历没有项目经历会影响初筛通过率。",
				Action:           "补充你参与过的课程项目、个人项目或实验室项目。",
				ExampleReference: "示例参考：基于 Gin 框架实现博客系统，包含用户注册登录、文章 CRUD 和评论功能。",
				SelfCheckQuestion: "请确认项目确实是你本人参与完成的。",
				LearnMoreQuery:   "学生简历 项目经历 怎么写",
				EstimatedEffort:  "within_1_hour",
			})
		}
	}

	// Fix 2: Tech stack expression
	missingTechKws := []string{}
	for _, kw := range []string{"go", "mysql", "redis"} {
		if strings.Contains(jdLower, kw) && !strings.Contains(resumeLower, kw) {
			missingTechKws = append(missingTechKws, kw)
		}
	}
	if len(missingTechKws) > 0 {
		kwStr := strings.Join(missingTechKws, "、")
		items = append(items, model.ResumeOptimizationFixItem{
			Issue:            fmt.Sprintf("简历中未体现 %s 相关经验", kwStr),
			WhyItMatters:     fmt.Sprintf("岗位 JD 明确提及 %s，这是初筛关键词。", kwStr),
			Action:           fmt.Sprintf("如果真实使用过 %s，在项目或实习描述中补充使用场景。", kwStr),
			ExampleReference: fmt.Sprintf("示例参考：使用 %s 完成数据持久化，通过连接池优化查询性能。", kwStr),
			SelfCheckQuestion: "请确认你确实在项目中使用过这些技术，而不是仅了解概念。",
			LearnMoreQuery:   fmt.Sprintf("%s 后端开发 实际项目使用场景", kwStr),
			EstimatedEffort:  "within_30_min",
		})
	}

	// Fix 3: Quantifiable results (without fabricating numbers)
	if !hasQuantifiableHint(resumeLower) {
		items = append(items, model.ResumeOptimizationFixItem{
			Issue:            "项目成果缺少量化表达",
			WhyItMatters:     "初筛时量化指标比泛泛描述更有说服力。",
			Action:           "补充接口数量、表设计规模、功能模块数等真实可量化的成果。",
			ExampleReference: "示例参考：独立完成 12 个 RESTful API 接口的开发，涉及 8 张数据库表的设计与优化。",
			SelfCheckQuestion: "请确认所有数字均为真实数据，不要编造指标。",
			LearnMoreQuery:   "后端简历 量化表达 怎么写",
			EstimatedEffort:  "within_30_min",
		})
	}

	// Cap at 3 items
	if len(items) > 3 {
		items = items[:3]
	}

	// Ensure at least 1 item
	if len(items) == 0 {
		items = append(items, model.ResumeOptimizationFixItem{
			Issue:            "简历表达与目标岗位关联度可加强",
			WhyItMatters:     "当前简历内容偏通用，建议围绕目标岗位关键词调整表达。",
			Action:           "在项目描述中突出与目标岗位相关的技术栈和职责。",
			ExampleReference: "示例参考：使用 Go + MySQL 开发后台管理系统，完成权限模块的接口设计与实现。",
			SelfCheckQuestion: "请确认所有描述均为真实经历。",
			LearnMoreQuery:   "简历 关键词优化 投递策略",
			EstimatedEffort:  "within_30_min",
		})
	}

	return items
}

// generateLongTermImprovements generates long-term improvement items.
// These should NOT delay current application.
func generateLongTermImprovements(resumeLower string, jdLower string) []model.ResumeOptimizationLongTermItem {
	var items []model.ResumeOptimizationLongTermItem

	// Check for advanced backend skills
	advancedKws := []string{"压测", "性能优化", "线上问题", "troubleshoot", "监控"}
	hasAdvanced := false
	for _, kw := range advancedKws {
		if strings.Contains(resumeLower, kw) {
			hasAdvanced = true
			break
		}
	}
	if !hasAdvanced {
		items = append(items, model.ResumeOptimizationLongTermItem{
			Direction: "后续可补充压测、性能优化或线上问题处理经验。",
			Note:      "这属于长期提升方向，不建议因为这一点延后当前投递。",
		})
	}

	// Check for open source or community contribution
	if !strings.Contains(resumeLower, "开源") && !strings.Contains(resumeLower, "github") {
		items = append(items, model.ResumeOptimizationLongTermItem{
			Direction: "后续可补充开源项目贡献或技术博客。",
			Note:      "这属于长期提升方向，不建议因为这一点延后当前投递。",
		})
	}

	if len(items) == 0 {
		items = append(items, model.ResumeOptimizationLongTermItem{
			Direction: "持续关注目标岗位方向的新技术和最佳实践。",
			Note:      "这属于长期提升方向，不建议因为这一点延后当前投递。",
		})
	}

	return items
}

// generateOptimizationSummary creates a summary based on apply_readiness and fix count.
func generateOptimizationSummary(applyReadiness string, fixCount int, jobTitle string) string {
	jobDesc := "目标岗位"
	if jobTitle != "" {
		jobDesc = jobTitle
	}

	switch applyReadiness {
	case "ready_to_apply":
		return fmt.Sprintf("当前简历与%s的核心要求匹配度较好，可以投递。", jobDesc)
	case "quick_fix_before_apply":
		return fmt.Sprintf("当前简历与%s有一定匹配基础，建议先完成 %d 个快速优化项后投递。", jobDesc, fixCount)
	case "not_ready_for_this_role":
		return fmt.Sprintf("当前简历与%s的关键要求差距较明显，建议先补充相关经历再投递。", jobDesc)
	case "insufficient_information":
		return "当前简历信息不足，无法形成可靠的投递建议，建议先补充简历内容。"
	default:
		return fmt.Sprintf("当前简历与%s有一定匹配基础，建议优化后投递。", jobDesc)
	}
}

// hasQuantifiableHint checks if the resume text contains quantifiable hints.
// Does NOT fabricate metrics.
func hasQuantifiableHint(text string) bool {
	quantHints := []string{"接口", "表", "模块", "功能", "api", "页面", "系统", "个", "项"}
	for _, hint := range quantHints {
		if strings.Contains(text, hint) {
			return true
		}
	}
	return false
}
