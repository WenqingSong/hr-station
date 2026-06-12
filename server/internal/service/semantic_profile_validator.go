package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"offer-catcher/server/internal/model"
)

// ─── Semantic Profile Validator ─────────────────────────────────────────

// ValidateSemanticProfileOutput performs deterministic quality checks on LLM output.
// Returns nil if the output passes all checks, or a multi-error message.
func ValidateSemanticProfileOutput(output *model.SemanticProfileLLMOutput, targetType string) error {
	if output == nil {
		return fmt.Errorf("validator: output is nil")
	}

	var errs []string

	// 1. summary_sentence must exist and have meaningful length
	if len(output.SummarySentence) < 20 {
		errs = append(errs, fmt.Sprintf("summary_sentence too short (%d chars, need >=20)", len(output.SummarySentence)))
	}

	// 2. tags count range
	if len(output.Tags) < 5 {
		errs = append(errs, fmt.Sprintf("tags count too low: %d (need >=5)", len(output.Tags)))
	}
	if len(output.Tags) > 24 {
		errs = append(errs, fmt.Sprintf("tags count too high: %d (need <=24)", len(output.Tags)))
	}

	// 3. per-tag validation
	seenCodes := map[string]bool{}
	for i, t := range output.Tags {
		prefix := fmt.Sprintf("tags[%d]", i)

		if t.TagCode == "" {
			errs = append(errs, fmt.Sprintf("%s: tag_code empty", prefix))
		} else if seenCodes[t.TagCode] {
			errs = append(errs, fmt.Sprintf("%s: duplicate tag_code %q", prefix, t.TagCode))
		}
		seenCodes[t.TagCode] = true

		if t.TagLabel == "" {
			errs = append(errs, fmt.Sprintf("%s: tag_label empty", prefix))
		}
		if model.IsBlacklistedRawCode(t.TagLabel) {
			errs = append(errs, fmt.Sprintf("%s: tag_label %q is blacklisted raw code", prefix, t.TagLabel))
		}

		if !isValidSemanticCategory(t.Category) {
			errs = append(errs, fmt.Sprintf("%s: invalid category %q", prefix, t.Category))
		}

		if t.Confidence < 0 || t.Confidence > 1 {
			errs = append(errs, fmt.Sprintf("%s: confidence out of range %.2f", prefix, t.Confidence))
		}

		if strings.TrimSpace(t.EvidenceText) == "" {
			errs = append(errs, fmt.Sprintf("%s: evidence_text empty for tag %q", prefix, t.TagCode))
		}
	}

	// 4. strengths
	if len(output.Strengths) < 1 {
		errs = append(errs, "strengths count too low (need >=1)")
	}
	for i, s := range output.Strengths {
		if model.IsBlacklistedRawCode(s) {
			errs = append(errs, fmt.Sprintf("strengths[%d] %q is blacklisted raw code", i, s))
		}
	}

	// 5. fit_directions
	if len(output.FitDirections) < 1 {
		errs = append(errs, "fit_directions count too low (need >=1)")
	}

	// 6. risks (flexible: accepts both structured array and raw JSON)
	if len(output.Risks) > 0 {
		var risks []map[string]interface{}
		if err := json.Unmarshal(output.Risks, &risks); err == nil {
			for _, r := range risks {
				ev := fmt.Sprint(r["evidence_text"])
				exp := fmt.Sprint(r["explanation"])
				if strings.TrimSpace(ev) == "" && strings.TrimSpace(exp) == "" {
					errs = append(errs, "risks: some entries have empty evidence and explanation")
					break
				}
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("validation failed for %s profile: %s", targetType, strings.Join(errs, "; "))
	}
	return nil
}

// ─── Helpers ─────────────────────────────────────────────────────────

var validCategories = map[string]bool{
	model.SemanticTagRoleDirection:         true,
	model.SemanticTagTechnicalSkill:        true,
	model.SemanticTagEngineeringCapability: true,
	model.SemanticTagDomainContext:         true,
	model.SemanticTagExperienceSignal:      true,
	model.SemanticTagRiskSignal:            true,
	model.SemanticTagPreferenceSignal:      true,
}

func isValidSemanticCategory(c string) bool { return validCategories[c] }

var validRiskSeverities = map[string]bool{
	"low": true, "medium": true, "high": true,
	"info": true, "warning": true, "critical": true,
	"": true, // empty defaults to info
}

func isValidRiskSeverity(s string) bool { return validRiskSeverities[s] }
