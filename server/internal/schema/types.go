package schema

// SchemaName is a logical identifier for a JSON Schema.
// Only predefined names are allowed; arbitrary strings are rejected.
type SchemaName string

const (
	// SchemaMatchAnalysis is the match analysis output Schema.
	SchemaMatchAnalysis SchemaName = "match_analysis"
	// SchemaRecommendation is the job recommendation output Schema.
	SchemaRecommendation SchemaName = "recommendation"
	// SchemaResumeOptimization is the resume optimization output Schema.
	SchemaResumeOptimization SchemaName = "resume_optimization"
	// SchemaVersionCompare is the version compare output Schema.
	SchemaVersionCompare SchemaName = "version_compare"
	// SchemaResumeOptimizationCore is the resume optimization core output Schema.
	SchemaResumeOptimizationCore SchemaName = "resume_optimization_core"
	// K13-Fix-C1: Deep recommendation 3-agent schemas.
	SchemaRecommendationDecision SchemaName = "recommendation_decision"
	SchemaRecommendationVerifier SchemaName = "recommendation_verifier"
	// K13-Fix-C2: Deep match 3-agent schemas.
	SchemaMatchDecision SchemaName = "match_decision"
	SchemaMatchVerifier SchemaName = "match_verifier"
	// K13-Fix-C3: Deep optimization 3-agent schemas.
	SchemaOptimizationDecision SchemaName = "optimization_decision"
	SchemaOptimizationVerifier SchemaName = "optimization_verifier"
	// K13-Fix-C4: Deep compare 3-agent schemas.
	SchemaCompareDecision SchemaName = "compare_decision"
	SchemaCompareVerifier SchemaName = "compare_verifier"
)

// Document holds a loaded JSON Schema file.
type Document struct {
	// Name is the logical SchemaName that was used to load this document.
	Name SchemaName
	// Path is the resolved file path on disk.
	Path string
	// Content is the raw JSON text of the Schema file.
	Content string
	// Raw is the parsed JSON object (top-level map).
	Raw map[string]any
}
