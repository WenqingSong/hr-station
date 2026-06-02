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
