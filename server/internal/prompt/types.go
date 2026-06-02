package prompt

// PromptName is a logical identifier for a Prompt template.
// Only predefined names are allowed; arbitrary strings are rejected.
type PromptName string

const (
	// PromptMatchAnalysis is the match analysis Agent prompt.
	PromptMatchAnalysis PromptName = "match_analysis"
	// PromptRecommendation is the job recommendation Agent prompt.
	PromptRecommendation PromptName = "recommendation"
	// PromptResumeOptimization is the resume optimization Agent prompt.
	PromptResumeOptimization PromptName = "resume_optimization"
	// PromptJSONRepair is the JSON repair Agent prompt.
	PromptJSONRepair PromptName = "json_repair"
	// PromptVersionCompare is the version compare Agent prompt.
	PromptVersionCompare PromptName = "version_compare"
)

// Template holds a loaded Prompt template.
type Template struct {
	// Name is the logical PromptName that was used to load this template.
	Name PromptName
	// Path is the resolved file path on disk.
	Path string
	// Content is the raw template text with placeholders still intact.
	Content string
}
