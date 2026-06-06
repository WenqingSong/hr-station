package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"offer-catcher/server/internal/model"
)

// ProfileAgent extracts candidate profiles via LLM with rule fallback.
// When LLM is disabled (runner==nil), always returns error to trigger rule fallback.
type ProfileAgent struct {
	runner PipelineRunner
}

// NewProfileAgent creates a ProfileAgent. runner may be nil if LLM is disabled.
func NewProfileAgent(runner PipelineRunner) *ProfileAgent {
	return &ProfileAgent{runner: runner}
}

// ProfileTagOutput is the LLM tag output format.
type ProfileTagOutput struct {
	Tag        string `json:"tag"`
	Group      string `json:"group"`
	Confidence int    `json:"confidence"`
	Evidence   string `json:"evidence,omitempty"`
}

// CandidateProfileOutput is the parsed LLM result for candidate profile extraction.
type CandidateProfileOutput struct {
	DominantProfile string             `json:"dominant_profile"`
	ProfileSummary  string             `json:"profile_summary"`
	PrimaryTags     []ProfileTagOutput `json:"primary_tags"`
	CandidateTags   []ProfileTagOutput `json:"candidate_tags"`
	RiskTags        []ProfileTagOutput `json:"risk_tags"`
}

// ── Prompt ───────────────────────────────────────────────────────────────

const candidateProfilePrompt = `You are a career profile extractor. Output ONLY pure JSON, no markdown.

Dominant profiles: algorithm_research, algorithm_engineering, backend_engineering, product_operation, business_solution, career_exploration

Tags (direction): algorithm, backend, product, business, research, data
Tags (capability): research_ability, experiment_ability, modeling_ability, engineering_ability, data_analysis_ability
Tags (domain): recommendation_system, search_algorithm, nlp, multimodal, computer_vision, data_mining, backend_service
Tags (risk): lack_industry_experience, lack_engineering_depth, lack_algorithm_project, lack_research_depth, over_general_skills

Rules: confidence 0-3, confidence>=2 needs evidence. Short skill tokens (go/redis/docker) must not override algorithm/research signals.

Output JSON: {"dominant_profile":"...","profile_summary":"...","primary_tags":[...],"candidate_tags":[...],"risk_tags":[...]}

Resume: %s`

// ── Public API ───────────────────────────────────────────────────────────

// ExtractCandidateProfile attempts LLM extraction.
// Returns error to trigger rule fallback (always when LLM disabled).
func (a *ProfileAgent) ExtractCandidateProfile(ctx context.Context, resume *model.ResumeVersion) (*CandidateProfileOutput, error) {
	if a.runner == nil {
		return nil, fmt.Errorf("LLM runner not available")
	}

	resumeText := buildProfileText(resume)
	promptText := fmt.Sprintf(candidateProfilePrompt, resumeText)

	log.Printf("[ProfileAgent] prompt built, len=%d resume_id=%d", len(promptText), resume.ID)

	// When LLM is available, call the pipeline here.
	// For now (LLM disabled), always fall back to rule extraction.
	_ = promptText
	return nil, fmt.Errorf("LLM profile extraction not yet enabled (pipeline integration pending)")
}

// ── Validation ───────────────────────────────────────────────────────────

var validDomProfiles = map[string]bool{
	"algorithm_research": true, "algorithm_engineering": true,
	"backend_engineering": true, "product_operation": true,
	"business_solution": true, "career_exploration": true,
}

var allowedTags = buildAllowedTags()

func buildAllowedTags() map[string]bool {
	m := map[string]bool{}
	for _, t := range []string{
		"algorithm", "backend", "product", "business", "research", "data",
		"ai_platform", "client", "solution",
		"research_ability", "experiment_ability", "modeling_ability",
		"engineering_ability", "data_analysis_ability", "system_design_ability",
		"business_understanding", "product_sense", "communication_ability",
		"recommendation_system", "search_algorithm", "nlp", "multimodal",
		"computer_vision", "data_mining", "machine_learning_platform",
		"high_performance_computing", "backend_service", "cloud_infrastructure",
		"client_development", "solution_consulting", "product_operation",
		"supply_chain_business",
		"lack_industry_experience", "lack_engineering_depth", "lack_algorithm_project",
		"lack_research_depth", "lack_business_context", "direction_unclear",
		"over_general_skills", "lack_backend_depth", "lack_cv_project",
	} {
		m[t] = true
	}
	return m
}

// ValidateAndCleanProfile validates and cleans LLM output tags.
func ValidateAndCleanProfile(p *CandidateProfileOutput) *CandidateProfileOutput {
	if !validDomProfiles[p.DominantProfile] {
		p.DominantProfile = "career_exploration"
	}
	p.PrimaryTags = cleanTags(p.PrimaryTags)
	p.CandidateTags = cleanTags(p.CandidateTags)
	p.RiskTags = cleanTags(p.RiskTags)
	return p
}

func cleanTags(tags []ProfileTagOutput) []ProfileTagOutput {
	var cleaned []ProfileTagOutput
	for _, t := range tags {
		if !allowedTags[t.Tag] {
			continue
		}
		if t.Confidence < 0 {
			t.Confidence = 0
		}
		if t.Confidence > 3 {
			t.Confidence = 3
		}
		if t.Confidence >= 2 && t.Evidence == "" {
			t.Confidence = 1
		}
		if t.Confidence <= 0 {
			continue
		}
		cleaned = append(cleaned, t)
	}
	return cleaned
}

// ParseProfileJSON parses JSON output into CandidateProfileOutput.
func ParseProfileJSON(rawJSON string) (*CandidateProfileOutput, error) {
	var result CandidateProfileOutput
	if err := json.Unmarshal([]byte(rawJSON), &result); err != nil {
		return nil, fmt.Errorf("parse profile JSON: %w", err)
	}
	return ValidateAndCleanProfile(&result), nil
}

// ConvertToProfileTags converts LLM tags to model.ProfileTag.
func ConvertToProfileTags(llmTags []ProfileTagOutput) []model.ProfileTag {
	result := make([]model.ProfileTag, len(llmTags))
	for i, t := range llmTags {
		result[i] = model.ProfileTag{Tag: t.Tag, Group: t.Group, Confidence: t.Confidence, Evidence: t.Evidence}
	}
	return result
}

// buildProfileText builds prompt input text from resume.
func buildProfileText(resume *model.ResumeVersion) string {
	return fmt.Sprintf("Name:%s Skills:%s Edu:%s Projects:%s Internships:%s Extra:%s",
		resume.Name, resume.ProfessionalSkills, educationsText(resume),
		projectsText(resume), internshipsText(resume), resume.AdditionalInfo)
}
