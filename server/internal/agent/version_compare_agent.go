package agent

import (
	"context"
	"fmt"
	"log"

	"offer-catcher/server/internal/agent/pipeline"
	"offer-catcher/server/internal/model"
	"offer-catcher/server/internal/prompt"
	"offer-catcher/server/internal/schema"
)

// Ensure VersionCompareLLMAgent implements VersionComparer.
var _ VersionComparer = (*VersionCompareLLMAgent)(nil)

const versionCompareBusinessContext = `你是简历版本对比助手。
必须基于输入的两份简历内容进行对比。
不要编造简历内容。
输出必须是严格 JSON。
如果新旧版本中都没有某类信息，标记为 missing_items。`

// VersionCompareLLMAgent uses the LLM Pipeline to generate version comparisons.
type VersionCompareLLMAgent struct {
	runner PipelineRunner
}

// NewVersionCompareLLMAgent creates a new VersionCompareLLMAgent.
func NewVersionCompareLLMAgent(runner PipelineRunner) *VersionCompareLLMAgent {
	return &VersionCompareLLMAgent{runner: runner}
}

// Compare generates a version comparison via the LLM Pipeline.
func (a *VersionCompareLLMAgent) Compare(from *model.ResumeVersion, to *model.ResumeVersion) (*model.VersionCompareAnalysis, error) {
	promptValues, err := a.buildPromptValues(from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to build prompt values: %w", err)
	}

	runReq := pipeline.RunRequest{
		AgentName:        "VersionCompareAgent",
		PromptName:       prompt.PromptVersionCompare,
		SchemaName:       schema.SchemaVersionCompare,
		TargetOutputType: pipeline.TargetCompare,
		PromptValues:     promptValues,
	}

	runResult := a.runner.Run(context.Background(), runReq)

	log.Printf("[VersionCompareAgent] pipeline result: status=%s gen=%s repair=%v fallback=%v reason=%s err=%s",
		runResult.AgentStatus, runResult.GeneratedBy, runResult.RepairApplied,
		runResult.FallbackRequired, runResult.FallbackReason, runResult.ErrorCode)

	if runResult.FallbackRequired {
		return nil, &VersionCompareFallbackError{
			FallbackReason: runResult.FallbackReason,
			ErrorCode:      runResult.ErrorCode,
			ErrorMessage:   runResult.ErrorMessage,
		}
	}

	result, err := mapParsedToCompareResult(runResult.Parsed, from, to)
	if err != nil {
		return nil, &VersionCompareFallbackError{
			FallbackReason: "field_mapping_failed",
			ErrorCode:      "FIELD_MAPPING_FAILED",
			ErrorMessage:   fmt.Sprintf("failed to map LLM output: %s", err),
		}
	}

	result.GeneratedBy = runResult.GeneratedBy
	result.AgentStatus = string(runResult.AgentStatus)
	return result, nil
}

// VersionCompareFallbackError indicates the Pipeline returned fallback_required.
type VersionCompareFallbackError struct {
	FallbackReason string
	ErrorCode      string
	ErrorMessage   string
}

func (e *VersionCompareFallbackError) Error() string {
	return fmt.Sprintf("version_compare fallback: reason=%s code=%s msg=%s",
		e.FallbackReason, e.ErrorCode, e.ErrorMessage)
}

// buildPromptValues uses ResumeSnapshotV2 for both versions.
// K8-5-22-I: Replaces old resumeToMap with snapshot-based serialization.
func (a *VersionCompareLLMAgent) buildPromptValues(from *model.ResumeVersion, to *model.ResumeVersion) (map[string]string, error) {
	oldJSON, err := buildResumeSnapshotJSON(from)
	if err != nil {
		log.Printf("[VersionCompareAgent] WARNING: old version snapshot failure: %v", err)
		if oldJSON == "" { oldJSON = "{}" }
	}
	newJSON, err := buildResumeSnapshotJSON(to)
	if err != nil {
		log.Printf("[VersionCompareAgent] WARNING: new version snapshot failure: %v", err)
		if newJSON == "" { newJSON = "{}" }
	}
	return map[string]string{
		"OLD_RESUME_JSON":  oldJSON,
		"NEW_RESUME_JSON":  newJSON,
		"BUSINESS_CONTEXT": versionCompareBusinessContext,
	}, nil
}

// resumeToMap is removed; replaced by buildResumeSnapshotJSON using NormalizeResumeToSnapshotV2.

func mapParsedToCompareResult(parsed map[string]any, from, to *model.ResumeVersion) (*model.VersionCompareAnalysis, error) {
	if parsed == nil {
		return nil, fmt.Errorf("parsed output is nil")
	}

	result := &model.VersionCompareAnalysis{
		FromVersionNo: from.VersionNo,
		ToVersionNo:   to.VersionNo,
		Summary:       getString(parsed, "compare_summary", ""),
		OverallChange: getString(parsed, "overall_change", "neutral"),
		Recommendation: getString(parsed, "recommendation", ""),
		NextActions:   parseStringArray(parsed, "next_actions"),
		AuthenticityNotice: getString(parsed, "authenticity_notice",
			"版本对比仅用于呈现表达变化，请只补充真实经历。"),
		ImprovedItems: parseChangeItems(parsed, "improved_items"),
		WeakenedItems:  parseChangeItems(parsed, "weakened_items"),
		MissingItems:   parseStringArray(parsed, "missing_items"),
	}
	return result, nil
}

func parseChangeItems(parsed map[string]any, key string) []model.CompareChangeItem {
	items, ok := getSlice(parsed, key)
	if !ok {
		return []model.CompareChangeItem{}
	}
	var result []model.CompareChangeItem
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		result = append(result, model.CompareChangeItem{
			Field:  getString(m, "field", ""),
			Change: getString(m, "change", ""),
		})
	}
	return result
}

func parseStringArray(parsed map[string]any, key string) []string {
	items, ok := getSlice(parsed, key)
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
