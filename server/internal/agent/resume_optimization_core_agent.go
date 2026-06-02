package agent

import (
	"context"
	"fmt"
	"log"
	"strings"

	"offer-catcher/server/internal/agent/pipeline"
	"offer-catcher/server/internal/prompt"
	"offer-catcher/server/internal/schema"
)

const resumeOptimizationCoreBusinessContext = `你给出了最多3条最重要的简历优化建议。
每条建议必须有JD证据和简历缺口。
不要编造经历。输出严格JSON。`

// CoreOptimizeResult is the output of the Core Agent.
type CoreOptimizeResult struct {
	OptimizationSummary string            `json:"optimization_summary"`
	ApplyReadiness      string            `json:"apply_readiness"`
	OptimizationStopSignal string         `json:"optimization_stop_signal,omitempty"`
	TopFixItems         []CoreFixItem     `json:"top_fix_items"`
	MissingKeywords     []string          `json:"missing_keywords"`
	FinalChecklist      []string          `json:"final_checklist"`
	AuthenticityNotice  string            `json:"authenticity_notice"`
	GeneratedBy         string            `json:"generated_by"`
	AgentStatus         string            `json:"agent_status,omitempty"`
}

// CoreFixItem is one optimization suggestion.
type CoreFixItem struct {
	Priority       string `json:"priority"`
	Issue          string `json:"issue"`
	JDEvidence     string `json:"jd_evidence"`
	ResumeGap      string `json:"resume_gap"`
	Action         string `json:"action"`
	RewriteExample string `json:"rewrite_example"`
	WhyItMatters   string `json:"why_it_matters"`
}

// CoreLLMAgent uses the Pipeline to generate core optimization results.
type CoreLLMAgent struct {
	runner PipelineRunner
}

// NewCoreLLMAgent creates a new CoreLLMAgent.
func NewCoreLLMAgent(runner PipelineRunner) *CoreLLMAgent {
	return &CoreLLMAgent{runner: runner}
}

// Optimize calls the LLM Pipeline and maps the result.
func (a *CoreLLMAgent) Optimize(ctx context.Context, input *OptimizeInput) (*CoreOptimizeResult, error) {
	promptValues := map[string]string{
		"RESUME_CONTEXT_JSON":    input.ResumeText,
		"TARGET_JOB_CONTEXT_JSON": fmt.Sprintf(`{"title":"%s","jd_text":"%s"}`, input.JobTitle, input.JDText),
	}

	runReq := pipeline.RunRequest{
		AgentName:        "ResumeOptimizationCoreAgent",
		PromptName:       prompt.PromptResumeOptimizationCore,
		SchemaName:       schema.SchemaResumeOptimizationCore,
		TargetOutputType: pipeline.TargetResumeOptimizationCore,
		PromptValues:     promptValues,
	}

	runResult := a.runner.Run(ctx, runReq)

	log.Printf("[CoreAgent] result: status=%s gen=%s repair=%v fallback=%v reason=%s",
		runResult.AgentStatus, runResult.GeneratedBy, runResult.RepairApplied,
		runResult.FallbackRequired, runResult.FallbackReason)

	if runResult.FallbackRequired {
		return nil, &CoreFallbackError{FallbackReason: runResult.FallbackReason, ErrorCode: runResult.ErrorCode}
	}

	result := mapCoreResult(runResult.Parsed)
	result.GeneratedBy = runResult.GeneratedBy
	result.AgentStatus = string(runResult.AgentStatus)
	return result, nil
}

// CoreFallbackError indicates fallback required.
type CoreFallbackError struct {
	FallbackReason string
	ErrorCode      string
}

func (e *CoreFallbackError) Error() string {
	return fmt.Sprintf("core optimization fallback: reason=%s", e.FallbackReason)
}

func mapCoreResult(parsed map[string]any) *CoreOptimizeResult {
	r := &CoreOptimizeResult{
		OptimizationSummary: getString(parsed, "optimization_summary", ""),
		ApplyReadiness:      getString(parsed, "apply_readiness", "needs_revision"),
		OptimizationStopSignal: getString(parsed, "optimization_stop_signal", ""),
		MissingKeywords:     parseStringSlice(parsed, "missing_keywords"),
		FinalChecklist:      parseStringSlice(parsed, "final_checklist"),
		AuthenticityNotice:  getString(parsed, "authenticity_notice", "只强化真实经历，不编造内容。"),
		GeneratedBy:         getString(parsed, "generated_by", "llm_agent"),
	}
	items, _ := getSlice(parsed, "top_fix_items")
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		r.TopFixItems = append(r.TopFixItems, CoreFixItem{
			Priority:       getString(m, "priority", "P2"),
			Issue:          getString(m, "issue", ""),
			JDEvidence:     getString(m, "jd_evidence", ""),
			ResumeGap:      getString(m, "resume_gap", ""),
			Action:         getString(m, "action", ""),
			RewriteExample: getString(m, "rewrite_example", ""),
			WhyItMatters:   getString(m, "why_it_matters", ""),
		})
	}
	return r
}

func parseStringSlice(m map[string]any, key string) []string {
	items, ok := getSlice(m, key)
	if !ok {
		return []string{}
	}
	var result []string
	for _, item := range items {
		if s, ok := item.(string); ok {
			result = append(result, s)
		}
	}
	return result
}

// CoreMockAgent generates rule-based core optimization results.
type CoreMockAgent struct{}

// NewCoreMockAgent creates a new CoreMockAgent.
func NewCoreMockAgent() *CoreMockAgent { return &CoreMockAgent{} }

// Optimize generates mock core optimization based on real input.
func (a *CoreMockAgent) Optimize(input *OptimizeInput) *CoreOptimizeResult {
	resumeLow := strings.ToLower(input.ResumeText)
	jdLow := strings.ToLower(input.JDText)

	// Extract keywords from JD
	jdKeywords := extractKeywords(jdLow, 5)
	resumeKeywords := extractKeywords(resumeLow, 3)
	missing := findMissing(jdKeywords, resumeKeywords)

	items := []CoreFixItem{
		{
			Priority: "P0", Issue: "关键技能未在简历中充分体现",
			JDEvidence: fmt.Sprintf("岗位JD中提到了 %s 等技术要求", strings.Join(jdKeywords[:min(3, len(jdKeywords))], "、")),
			ResumeGap: "简历中对应的项目证据或技能描述不足",
			Action: "在项目经历中明确写出使用这些技术的具体场景和成果",
			WhyItMatters: "HR 初筛时会在 30 秒内扫描这些关键词",
		},
		{
			Priority: "P1", Issue: "项目职责描述不够具体",
			JDEvidence: fmt.Sprintf("岗位要求具备 %s 相关项目经验", jdKeywords[0]),
			ResumeGap: "简历中缺少职责、技术栈和量化结果的完整描述",
			Action: "补充你在项目中的角色、使用的技术栈和达成的效果",
			WhyItMatters: "清晰的职责描述能让面试官快速评估你的技术深度",
		},
	}

	if len(missing) > 0 {
		items[0].RewriteExample = fmt.Sprintf("如果你确实使用过 %s，可以写：基于 %s 开发了 XX 功能", missing[0], missing[0])
	}

	return &CoreOptimizeResult{
		OptimizationSummary: fmt.Sprintf("该简历针对「%s」岗位最需要强化的是核心技能的项目证据与清晰的技术栈描述。", input.JobTitle),
		ApplyReadiness:      "needs_revision",
		TopFixItems:         items,
		MissingKeywords:     missing,
		FinalChecklist:      []string{"确认所有修改基于真实经历", "修改后重新运行岗位匹配分析", "补充量化指标如接口数量、性能对比等"},
		AuthenticityNotice:  "不要编造经历。只强化你真实做过的事情的表达。",
		GeneratedBy:         "mock_agent",
		AgentStatus:         "mock_success",
	}
}

func extractKeywords(text string, limit int) []string {
	var words []string
	fields := strings.FieldsFunc(text, func(r rune) bool {
		return r == ',' || r == ' ' || r == '，' || r == '\n'
	})
	for _, w := range fields {
		w = strings.TrimSpace(w)
		if len(w) > 1 && len(words) < limit {
			words = append(words, w)
		}
	}
	return words
}

func findMissing(jd, resume []string) []string {
	set := make(map[string]bool)
	for _, k := range resume {
		set[k] = true
	}
	var missing []string
	for _, k := range jd {
		if !set[k] {
			missing = append(missing, k)
		}
	}
	return missing
}
